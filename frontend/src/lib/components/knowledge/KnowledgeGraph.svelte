<script lang="ts">
	import { Canvas } from '@threlte/core';
	import KnowledgeScene from './KnowledgeScene.svelte';
	import type { Memory } from '$lib/api/memory/types';

	interface Props {
		memories: Memory[];
		onSelect?: (memory: Memory) => void;
		onDeselect?: () => void;
		selectedId?: string | null;
		searchQuery?: string;
		highlightedIds?: string[];
		nodes?: Map<string, { name: string; type: string }>; // Node lookup for display
		zoomLevel?: number; // 0-100, triggers zoom when changed
		onZoomChange?: (level: number) => void;
	}

	let {
		memories = [],
		onSelect,
		onDeselect,
		selectedId = null,
		searchQuery = '',
		highlightedIds = [],
		nodes: nodeMap = new Map(),
		zoomLevel = 50,
		onZoomChange
	}: Props = $props();

	// React to zoom level changes - start zoomed out to see full sphere
	$effect(() => {
		if (zoomLevel !== undefined) {
			// Zoom 50 = 120 distance (default zoomed out), 0 = 180, 100 = 60 (close)
			const distance = 180 - (zoomLevel * 1.2);
			cameraPosition = [0, distance * 0.3, distance];
		}
	});

	// Layout algorithm - Node-aware clustering
	interface LayoutNode {
		memory: Memory;
		position: [number, number, number];
		scale: number;
		nodeId: string | null;
	}

	// Connection between memories (for rendering lines)
	interface MemoryConnection {
		from: [number, number, number];
		to: [number, number, number];
		color: string;
	}

	// Seeded random for consistent layout
	function seededRandom(seed: number): () => number {
		return function() {
			seed = (seed * 9301 + 49297) % 233280;
			return seed / 233280;
		};
	}

	// Node Viewer-style layout: SPHERICAL distribution - bubbles on surface of invisible sphere
	function layoutNodes(mems: Memory[]): LayoutNode[] {
		const layoutNodes: LayoutNode[] = [];
		const count = mems.length;

		if (count === 0) return layoutNodes;

		// Use memory id hash as seed for consistent positions
		const hashCode = (str: string) => {
			let hash = 0;
			for (let i = 0; i < str.length; i++) {
				hash = ((hash << 5) - hash) + str.charCodeAt(i);
				hash |= 0;
			}
			return Math.abs(hash);
		};

		// Spherical configuration - bubbles on surface of sphere, nothing in center
		// Larger radius for better spacing - Node Viewer style has well-spaced bubbles
		const sphereRadius = Math.min(80, 40 + count * 1.5); // Much larger for better spacing
		const centerY = 0; // Sphere centered at origin

		mems.forEach((memory, i) => {
			const seed = hashCode(memory.id || String(i));
			const rand = seededRandom(seed);

			// Fibonacci sphere distribution for even spacing on sphere surface
			const goldenAngle = Math.PI * (3 - Math.sqrt(5));
			const theta = i * goldenAngle; // Horizontal angle

			// Vertical position: distribute from -1 to 1, converted to polar angle
			const y_ratio = 1 - (i / Math.max(count - 1, 1)) * 2; // -1 to 1
			const phi = Math.acos(y_ratio); // Convert to polar angle (0 to PI)

			// Add slight randomness to make it feel organic
			const randTheta = theta + (rand() - 0.5) * 0.3;
			const randPhi = phi + (rand() - 0.5) * 0.2;
			const randRadius = sphereRadius * (0.9 + rand() * 0.2); // 90-110% of radius

			// Convert spherical to cartesian coordinates - TRUE SPHERE (no compression)
			const x = randRadius * Math.sin(randPhi) * Math.cos(randTheta);
			const y = randRadius * Math.cos(randPhi) + centerY;
			const z = randRadius * Math.sin(randPhi) * Math.sin(randTheta); // No compression

			// Scale: 1.0 to 1.4 based on importance (smaller for cleaner look)
			const importanceScore = memory.importance_score || 0.5;
			const scale = 1.0 + importanceScore * 0.4;

			layoutNodes.push({
				memory,
				position: [x, y, z],
				scale,
				nodeId: memory.node_id
			});
		});

		return layoutNodes;
	}

	let layoutResult = $derived(layoutNodes(memories));

	// Generate connections for SELECTED bubble only (Node Viewer style)
	let connections = $derived.by(() => {
		const conns: MemoryConnection[] = [];

		// Only show connections when a bubble is selected
		if (!selectedId) return conns;

		// Find the selected node
		const selectedNode = layoutResult.find(n => n.memory.id === selectedId);
		if (!selectedNode) return conns;

		// Find related bubbles (same nodeId or linked memories)
		const relatedNodes = layoutResult.filter(n => {
			if (n.memory.id === selectedId) return false;
			// Same node_id means they're related
			if (selectedNode.nodeId && n.nodeId === selectedNode.nodeId) return true;
			// Could also check for linked_memory_ids in the future
			return false;
		});

		// Create thin connection lines to related bubbles
		relatedNodes.forEach(related => {
			conns.push({
				from: selectedNode.position,
				to: related.position,
				color: '#888888' // Subtle gray line
			});
		});

		return conns;
	});

	// Camera state - angled view looking at sphere from above-right
	let cameraPosition = $state<[number, number, number]>([70, 60, 110]);
	// Camera ALWAYS looks at center [0,0,0] - never changes
	const targetLookAt: [number, number, number] = [0, 0, 0];
	// Auto-rotate like a globe by default
	let autoRotate = $state(true);

	// Warm earth-tone color palette for bubbles - Node Viewer style
	const warmPalette = [
		'#8B7355',   // Warm brown
		'#C4A77D',   // Tan/beige
		'#7A9E7E',   // Sage green
		'#B8860B',   // Dark goldenrod
		'#9FA8B3',   // Cool gray-blue
		'#A0826D',   // Dusty rose/brown
		'#8B6914',   // Bronze/olive
		'#9E8B7D',   // Taupe
		'#8BA07A',   // Moss green
		'#A89078',   // Warm gray
		'#7D8B9E',   // Steel blue
		'#9B8B6E',   // Khaki
	];

	// Get color based on memory ID for variety - consistent per bubble
	function getTypeColor(type: string, memoryId?: string): string {
		// If we have a memory ID, use it to pick a consistent color
		if (memoryId) {
			let hash = 0;
			for (let i = 0; i < memoryId.length; i++) {
				hash = ((hash << 5) - hash) + memoryId.charCodeAt(i);
				hash |= 0;
			}
			const index = Math.abs(hash) % warmPalette.length;
			return warmPalette[index];
		}

		// Fallback to type-based colors
		const typeColors: Record<string, string> = {
			'fact': '#8B7355',
			'preference': '#C4A77D',
			'decision': '#7A9E7E',
			'event': '#B8860B',
			'learning': '#9FA8B3',
			'context': '#A0826D',
			'relationship': '#8B6914'
		};
		return typeColors[type] || '#9E9E9E';
	}

	// Handle bubble click
	function handleBubbleClick(memory: Memory) {
		onSelect?.(memory);
	}

	// Handle background click to deselect
	function handleBackgroundClick() {
		onDeselect?.();
	}

	// Check if memory is highlighted (from search)
	function isHighlighted(memory: Memory): boolean {
		if (highlightedIds.length > 0) {
			return highlightedIds.includes(memory.id);
		}
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			return (
				memory.title?.toLowerCase().includes(query) ||
				memory.summary?.toLowerCase().includes(query) ||
				memory.content?.toLowerCase().includes(query)
			);
		}
		return true; // No filter, all visible
	}

	// Zoom controls
	function zoomIn() {
		cameraPosition = [
			cameraPosition[0] * 0.8,
			cameraPosition[1] * 0.8,
			cameraPosition[2] * 0.8
		];
	}

	function zoomOut() {
		cameraPosition = [
			cameraPosition[0] * 1.2,
			cameraPosition[1] * 1.2,
			cameraPosition[2] * 1.2
		];
	}

	function resetView() {
		cameraPosition = [0, 30, 80];
	}

	function toggleAutoRotate() {
		autoRotate = !autoRotate;
	}
</script>

<div class="knowledge-graph">
	<Canvas>
		<KnowledgeScene
			{layoutResult}
			{selectedId}
			{searchQuery}
			{cameraPosition}
			{targetLookAt}
			{autoRotate}
			nodeMap={nodeMap}
			{getTypeColor}
			{isHighlighted}
			onBubbleClick={handleBubbleClick}
			onBackgroundClick={handleBackgroundClick}
		/>
	</Canvas>

	<!-- Stats overlay -->
	<div class="graph-stats">
		<span class="stat">{memories.length} bubbles</span>
		{#if searchQuery}
			<span class="stat-filter">
				{layoutResult.filter(n => isHighlighted(n.memory)).length} matching
			</span>
		{/if}
	</div>
</div>

<style>
	.knowledge-graph {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 500px;
		/* Node Viewer style: white top, gray bottom - floating room effect */
		background: linear-gradient(180deg,
			#ffffff 0%,
			#fafafa 30%,
			#e8e8e8 70%,
			#c8c8c8 100%
		);
		border-radius: 0;
		overflow: hidden;
		/* Prevent canvas from blocking UI elements with higher z-index */
		z-index: 0;
	}

	.knowledge-graph :global(canvas) {
		/* Ensure canvas stays in its layer */
		z-index: 0;
	}

	.graph-stats {
		position: absolute;
		bottom: 16px;
		left: 16px;
		display: flex;
		gap: 12px;
		padding: 8px 14px;
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(12px);
		border-radius: 10px;
		font-size: 12px;
		color: #666666;
		border: 1px solid rgba(0, 0, 0, 0.08);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
	}

	.stat-filter {
		color: #888888;
	}
</style>
