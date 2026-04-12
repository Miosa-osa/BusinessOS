<script lang="ts">
	import { onMount } from 'svelte';
	// force-graph ships incorrect .d.ts (declares class; runtime is a Kapsule factory).
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	import _FG from 'force-graph';
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const ForceGraph = _FG as unknown as () => (el: HTMLElement) => any;

	import { getApiBaseUrl } from '$lib/api/base';
	import type { Memory } from '$lib/api/memory/types';

	// ── Props (backward-compatible) ───────────────────────────────────────────
	interface Props {
		memories?: Memory[];
		onSelect?: (memory: Memory) => void;
		onDeselect?: () => void;
		selectedId?: string | null;
		searchQuery?: string;
		highlightedIds?: string[];
		nodes?: Map<string, { name: string; type: string }>;
		zoomLevel?: number;
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

	// Backward-compat props declared for API compatibility; not used in this component.

	// ── API types ──────────────────────────────────────────────────────────────
	interface ApiNode    { slug: string; name: string; type: string; signal_count: number; }
	interface ApiFile    { name: string; path: string; is_dir: boolean; size: number; children?: ApiFile[]; }
	interface ApiEntity  { name: string; type: string; connections: number; }
	interface ApiEdge    { source: string; target: string; relation: string; weight?: number; }

	// ── Graph node types ───────────────────────────────────────────────────────
	type GNodeType = 'core' | 'folder' | 'document' | 'entity';

	interface GNode {
		id: string;
		label: string;
		nodeType: GNodeType;
		val: number;
		// for entity nodes
		connections?: number;
		entityType?: string;
	}

	interface GLink {
		source: string;
		target: string;
		relation: string;
	}

	// ── Colors ────────────────────────────────────────────────────────────────
	const NODE_HOVER   = '#7c3aed';
	const LINK_DEFAULT = 'rgba(0,0,0,0.04)';
	const LINK_HOVER   = 'rgba(124,58,237,0.5)';

	// Node type → color
	const TYPE_COLOR: Record<GNodeType, string> = {
		core:     '#1a1a1a',
		folder:   '#9ca3af',
		document: '#d1d5db',
		entity:   '#3b82f6'
	};

	const CHARGE = -5;

	// ── Component state ────────────────────────────────────────────────────────
	let container  = $state<HTMLDivElement | null>(null);
	let loading    = $state(true);
	let error      = $state<string | null>(null);
	let statsText  = $state('');

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let graph: any = null;

	function destroyGraph() {
		if (!graph) return;
		try { graph._destructor?.(); } catch { /* ignore */ }
		graph = null;
	}

	// ── File tree walker ───────────────────────────────────────────────────────
	function walkTree(
		files: ApiFile[],
		parentId: string,
		gNodes: GNode[],
		gLinks: GLink[]
	) {
		for (const f of files) {
			const id    = f.path;                          // full path as id
			const label = f.name;                          // last segment as label
			const type: GNodeType = f.is_dir ? 'folder' : 'document';

			gNodes.push({ id, label, nodeType: type, val: f.is_dir ? 1.5 : 1 });
			gLinks.push({ source: parentId, target: id, relation: 'contains' });

			if (f.is_dir && f.children?.length) {
				walkTree(f.children, id, gNodes, gLinks);
			}
		}
	}

	// ── Core node slugs set (used for entity→core edge routing) ───────────────
	function buildSlugSet(coreNodes: ApiNode[]): Set<string> {
		const s = new Set<string>();
		for (const n of coreNodes) s.add(n.slug);
		// also add numeric prefixes like "02-miosa"
		return s;
	}

	// ── Main fetch + render ────────────────────────────────────────────────────
	async function initGraph() {
		if (!container) return;
		loading = true;
		error   = null;

		try {
			const base = getApiBaseUrl();

			// 1. Fetch all 3 endpoints in parallel
			const [nodesRes, graphRes] = await Promise.all([
				fetch(`${base}/optimal/nodes`),
				fetch(`${base}/optimal/graph`)
			]);

			if (!nodesRes.ok) throw new Error(`/optimal/nodes → HTTP ${nodesRes.status}`);
			if (!graphRes.ok) throw new Error(`/optimal/graph → HTTP ${graphRes.status}`);

			const nodesData = await nodesRes.json();
			const coreNodes: ApiNode[] = Array.isArray(nodesData) ? nodesData : (nodesData.nodes ?? []);
			const graphRaw = await graphRes.json();
			const graphData: { entities: ApiEntity[]; edges: ApiEdge[] } = {
				entities: graphRaw.entities ?? [],
				edges: graphRaw.edges ?? []
			};

			// 2. Fetch file trees for all nodes in parallel
			const fileTreeResults = await Promise.allSettled(
				coreNodes.map(n =>
					fetch(`${base}/optimal/nodes/${n.slug}/files`)
						.then(r => r.ok ? r.json() : Promise.resolve({ files: [] }))
						.then((d: { files: ApiFile[] }) => ({ slug: n.slug, files: d.files ?? [] }))
				)
			);

			const gNodes: GNode[] = [];
			const gLinks: GLink[] = [];

			// 3. Add core nodes
			const slugSet = buildSlugSet(coreNodes);
			for (const n of coreNodes) {
				gNodes.push({
					id:       n.slug,
					label:    n.name,
					nodeType: 'core',
					val:      2
				});
			}

			// 4. Walk file trees → folder + document nodes + containment links
			for (const result of fileTreeResults) {
				if (result.status !== 'fulfilled') continue;
				const { slug, files } = result.value;
				walkTree(files, slug, gNodes, gLinks);
			}

			// 5. Add entity nodes
			const entityIds = new Set<string>();
			for (const e of graphData.entities) {
				if (!entityIds.has(e.name)) {
					entityIds.add(e.name);
					gNodes.push({
						id:          e.name,
						label:       e.name,
						nodeType:    'entity',
						val:         2,
						connections: e.connections,
						entityType:  e.type
					});
				}
			}

			// 5b. Connect ALL core nodes to each other with weak links
			// This keeps them clustered together instead of flying to corners
			const coreList = coreNodes.map(n => n.slug);
			for (let i = 0; i < coreList.length; i++) {
				for (let j = i + 1; j < coreList.length; j++) {
					gLinks.push({ source: coreList[i], target: coreList[j], relation: 'sibling' });
				}
			}

			// 5c. Connect entities to their most relevant core nodes
			// This chains people to companies
			for (const e of graphData.edges) {
				if (entityIds.has(e.source) && slugSet.has(e.target)) {
					gLinks.push({ source: e.source, target: e.target, relation: 'belongs_to' });
				}
				if (entityIds.has(e.target) && slugSet.has(e.source)) {
					gLinks.push({ source: e.target, target: e.source, relation: 'belongs_to' });
				}
			}

			// 6. Add edges from graph data
			//    - entity→core: when source is entity and target matches a slug
			//    - entity→entity: cross_ref between two entities
			//    - core→core: cross_ref between two cores
			const seenLinks = new Set<string>();

			function addLink(source: string, target: string, relation: string) {
				const key = `${source}||${target}||${relation}`;
				if (seenLinks.has(key)) return;
				seenLinks.add(key);
				gLinks.push({ source, target, relation });
			}

			for (const e of graphData.edges) {
				const srcIsCore   = slugSet.has(e.source);
				const tgtIsCore   = slugSet.has(e.target);
				const srcIsEntity = entityIds.has(e.source);
				const tgtIsEntity = entityIds.has(e.target);

				if (e.relation === 'cross_ref') {
					if (srcIsCore && tgtIsCore)     addLink(e.source, e.target, 'cross_ref');
					if (srcIsEntity && tgtIsEntity) addLink(e.source, e.target, 'cross_ref');
					if (srcIsEntity && tgtIsCore)   addLink(e.source, e.target, 'mentioned_in');
					if (tgtIsEntity && srcIsCore)   addLink(e.target, e.source, 'mentioned_in');
				} else if (e.relation === 'mentioned_in' || e.relation === 'lives_in') {
					// Only keep entity→core direction
					if (srcIsEntity && tgtIsCore)   addLink(e.source, e.target, 'mentioned_in');
					if (tgtIsEntity && srcIsCore)   addLink(e.target, e.source, 'mentioned_in');
				}
			}

			// 7. Remove isolated nodes (zero edges) — prevents outlier dots flying to edges
			const connectedIds = new Set<string>();
			for (const lk of gLinks) {
				connectedIds.add(lk.source);
				connectedIds.add(lk.target);
			}
			const filteredNodes = gNodes.filter(n => connectedIds.has(n.id));

			// 8. Set initial positions within a small circle so the simulation
			//    converges to a sphere instead of exploding outward
			for (const n of filteredNodes) {
				const angle = Math.random() * Math.PI * 2;
				const r = Math.random() * 50;
				(n as GNode & { x?: number; y?: number }).x = Math.cos(angle) * r;
				(n as GNode & { x?: number; y?: number }).y = Math.sin(angle) * r;
			}

			statsText = `${filteredNodes.length} nodes · ${gLinks.length} edges`;

			destroyGraph();
			container.innerHTML = '';

			// ── Hover state (bypasses Svelte diffing — runs at 60fps) ──────────
			let _hovId: string | null = null;
			const _connN = new Set<string>();
			const _connL = new Set<string>();

			// O(k) adjacency lookup (uses filteredNodes — isolated nodes already removed)
			const adj = new Map<string, Array<{ src: string; tgt: string }>>();
			for (const lk of gLinks) {
				if (!adj.has(lk.source)) adj.set(lk.source, []);
				if (!adj.has(lk.target)) adj.set(lk.target, []);
				adj.get(lk.source)!.push({ src: lk.source, tgt: lk.target });
				adj.get(lk.target)!.push({ src: lk.source, tgt: lk.target });
			}

			function setHover(node: { id: string } | null) {
				_hovId = node?.id ?? null;
				_connN.clear();
				_connL.clear();
				if (node) {
					for (const e of (adj.get(node.id) ?? [])) {
						_connN.add(e.src);
						_connN.add(e.tgt);
						_connL.add(`${e.src}__${e.tgt}`);
					}
				}
			}

			// ── Build graph ─────────────────────────────────────────────────────
			graph = ForceGraph()(container)
				.graphData({ nodes: filteredNodes, links: gLinks })
				.backgroundColor('#fafafa')
				.nodeVal('val')
				.nodeLabel('')
				.autoPauseRedraw(false)
				.minZoom(0.1)
				.maxZoom(16)
				// ── Custom node rendering ──────────────────────────────────────
				.nodeCanvasObject((node: GNode & { x: number; y: number }, ctx: CanvasRenderingContext2D, gs: number) => {
					const id      = node.id;
					const nType   = node.nodeType ?? 'document';
					const isHov   = id === _hovId;
					const isCon   = _connN.has(id);
					// ALL dots tiny and uniform — like Obsidian
					const r = nType === 'core' ? 2.5 : nType === 'entity' ? 2 : 1.5;
					const baseCol = TYPE_COLOR[nType] ?? '#4a4a4a';

					ctx.beginPath();
					ctx.arc(node.x, node.y, r, 0, 2 * Math.PI);
					ctx.fillStyle = isHov ? NODE_HOVER
					              : isCon ? NODE_HOVER + 'bb'
					              : _hovId ? baseCol + '33'
					              : baseCol;
					ctx.fill();

					if (isHov || nType === 'core') {
						ctx.strokeStyle = isHov ? NODE_HOVER : baseCol + '66';
						ctx.lineWidth   = (nType === 'core' ? 1.5 : 1) / gs;
						ctx.stroke();
					}

					// Labels: ONLY on hover/connected or when zoomed in very close
					const showLabel = isHov
					  || isCon
					  || (nType === 'core' && gs > 1.5)
					  || (nType === 'entity' && gs > 3)
					  || (gs > 5);

					if (showLabel) {
						const fs = nType === 'core' ? 12 / gs
						         : isHov             ? 11 / gs
						         :                     9 / gs;
						const bold = nType === 'core' || isHov;
						ctx.font         = `${bold ? '600 ' : ''}${fs}px -apple-system, system-ui, sans-serif`;
						ctx.fillStyle    = isHov ? '#1a1a1a'
						                 : isCon ? '#374151'
						                 : nType === 'core' ? '#111827'
						                 : nType === 'entity' ? '#1d4ed8'
						                 : 'rgba(0,0,0,0.35)';
						ctx.textBaseline = 'middle';
						ctx.fillText(node.label ?? id, node.x + r + 2 / gs, node.y);
					}
				})
				.nodePointerAreaPaint((node: GNode & { x: number; y: number }, color: string, ctx: CanvasRenderingContext2D) => {
					const r = Math.max(6, Math.sqrt(node.val ?? 1) * 2 + 4);
					ctx.beginPath();
					ctx.arc(node.x, node.y, r, 0, 2 * Math.PI);
					ctx.fillStyle = color;
					ctx.fill();
				})
				// ── Links ────────────────────────────────────────────────────────
				.linkColor((link: { source: GNode | string; target: GNode | string }) => {
					const s = typeof link.source === 'object' ? link.source.id : link.source;
					const t = typeof link.target === 'object' ? link.target.id : link.target;
					if (_connL.has(`${s}__${t}`) || _connL.has(`${t}__${s}`)) return LINK_HOVER;
					if (_hovId) return 'rgba(0,0,0,0.01)';
					return LINK_DEFAULT;
				})
				.linkWidth((link: { source: GNode | string; target: GNode | string }) => {
					const s = typeof link.source === 'object' ? link.source.id : link.source;
					const t = typeof link.target === 'object' ? link.target.id : link.target;
					return (_connL.has(`${s}__${t}`) || _connL.has(`${t}__${s}`)) ? 1.5 : 0.2;
				})
				// ── Physics — FROZEN after layout ────────────────────────────────
				// Run 500 ticks before first render, then STOP simulation completely.
				// No bouncing, no jittering. Completely static after initial layout.
				// Obsidian-matched physics: settles, reheats on drag
				.warmupTicks(600)
				.cooldownTime(20000)
				.d3AlphaDecay(0.0228)
				.d3AlphaMin(0.001)
				.d3VelocityDecay(0.4)
				.enableNodeDrag(true)
				// ── Events ───────────────────────────────────────────────────────
				.onZoom(({ k }: { k: number }) => {
					const pct = Math.round(((k - 0.1) / (16 - 0.1)) * 100);
					onZoomChange?.(Math.max(0, Math.min(100, pct)));
				})
				.onNodeHover((node: GNode | null) => {
					setHover(node);
					if (container) container.style.cursor = node ? 'grab' : 'default';
				})
				.onNodeDrag((node: any) => {
					setHover(node as GNode);
					// Obsidian drag: pin node to cursor, reheat simulation
					// Spring forces PULL connected nodes. Repulsion PUSHES others aside.
					node.fx = node.x;
					node.fy = node.y;
					const sim = (graph as any)?._simulation;
					if (sim) sim.alphaTarget(0.3).restart();
					if (container) container.style.cursor = 'grabbing';
				})
				.onNodeDragEnd((node: any) => {
					// Unpin node, let simulation cool down naturally
					node.fx = undefined;
					node.fy = undefined;
					const sim = (graph as any)?._simulation;
					if (sim) sim.alphaTarget(0);
					setHover(null);
					if (container) container.style.cursor = 'default';
				})
				.onNodeClick((node: GNode & { x: number; y: number }) => {
					// Zoom to the clicked node — Obsidian style
					graph?.centerAt(node.x, node.y, 600);
					graph?.zoom(4, 600);
					// Highlight its connections
					setHover(node);

					// For core nodes, open their context.md. For folders, open their path.
					// For documents (.md), open directly. For entities, open as search.
					const docId = node.nodeType === 'core' ? `${node.id}/context.md`
					            : node.nodeType === 'entity' ? node.id
					            : node.id;

					const syn: Memory = {
						id:               docId,
						user_id:          'graph',
						title:            node.label ?? node.id,
						summary:          `${node.nodeType} — ${node.connections ?? 0} connections`,
						content:          '',
						memory_type:      'context',
						importance_score: (node.connections ?? 0) / 200,
						is_pinned:        false,
						is_active:        true,
						tags:             [],
						metadata:         { entity_type: node.entityType ?? node.nodeType },
						source_type:      null,
						source_id:        null,
						project_id:       null,
						node_id:          null,
						expires_at:       null,
						access_count:     0,
						last_accessed_at: null,
						created_at:       '',
						updated_at:       ''
					};
					onSelect?.(syn);
				})
				.onBackgroundClick(() => onDeselect?.());

			// ── Obsidian-density forces — packed tight, collision prevents overlap ──
			graph.d3Force('charge')?.strength(CHARGE);
			graph.d3Force('link')?.distance(8).strength(0.6);
			graph.d3Force('center')?.strength(0.3);

			// Collision is the PRIMARY spacing mechanism — like Obsidian
			const d3 = await import('d3-force');
			graph.d3Force('collide', d3.forceCollide().radius(4).strength(1.0));

			// Size the canvas
			const w = container.clientWidth;
			const h = container.clientHeight;
			if (w > 0 && h > 0) graph.width(w).height(h);

			// Zoom to fit after warmup settles
			setTimeout(() => {
				graph?.zoomToFit(400, 30);
			}, 500);
			loading = false;
		} catch (err) {
			error   = err instanceof Error ? err.message : 'Failed to load graph';
			loading = false;
		}
	}

	// ── Legacy mode: build from memories prop ────────────────────────────────
	function buildFromMemories() {
		if (!container) return;
		loading = true;

		const gNodes = memories.map(m => ({
			id:        m.id,
			label:     m.title ?? m.id,
			nodeType:  'document' as GNodeType,
			val:       Math.log((m.importance_score ?? 0.5) * 100 + 1),
			connections: Math.round((m.importance_score ?? 0.5) * 100)
		}));

		const gLinks: GLink[] = [];
		for (let i = 0; i < memories.length; i++) {
			for (let j = i + 1; j < memories.length; j++) {
				if (memories[i].node_id && memories[i].node_id === memories[j].node_id) {
					gLinks.push({ source: memories[i].id, target: memories[j].id, relation: 'co_node' });
				}
			}
		}

		destroyGraph();
		container.innerHTML = '';

		const NODE_DEFAULT = '#4a4a4a';

		graph = ForceGraph()(container)
			.graphData({ nodes: gNodes, links: gLinks })
			.backgroundColor('#fafafa')
			.nodeVal('val')
			.nodeLabel('label')
			.nodeColor(() => NODE_DEFAULT)
			.linkColor(() => 'rgba(0,0,0,0.06)')
			.linkWidth(0.5)
			.cooldownTime(2000)
			.d3AlphaDecay(0.02)
			.d3VelocityDecay(0.3)
			.onNodeClick((node: GNode) => {
				const mem = memories.find(m => m.id === node.id);
				if (mem) onSelect?.(mem);
			})
			.onBackgroundClick(() => onDeselect?.());

		graph.width(container.clientWidth).height(container.clientHeight);

		statsText = `${memories.length} memories · ${gLinks.length} edges`;
		loading   = false;
	}

	// ── Mount ──────────────────────────────────────────────────────────────────
	let ro: ResizeObserver | null = null;

	onMount(() => {
		if (memories.length === 0) {
			initGraph();
		} else {
			buildFromMemories();
		}

		ro = new ResizeObserver(() => {
			if (graph && container) {
				graph.width(container.clientWidth).height(container.clientHeight);
			}
		});
		if (container) ro.observe(container);

		return () => {
			ro?.disconnect();
			destroyGraph();
		};
	});
</script>

<div class="kg-root" bind:this={container}>
	<!-- force-graph mounts its canvas directly into this element -->
</div>

{#if statsText && !loading}
	<div class="kg-stats" aria-label="Graph statistics">
		<span class="kg-stat-item">{statsText}</span>
		<span class="kg-stat-sep">·</span>
		<span class="kg-stat-legend">
			<span class="kg-dot" style="background:#1a1a1a"></span>core
			<span class="kg-dot" style="background:#3b82f6"></span>entity
			<span class="kg-dot" style="background:#9ca3af"></span>folder
			<span class="kg-dot" style="background:#d1d5db"></span>doc
		</span>
	</div>
{/if}

{#if loading && !error}
	<div class="kg-overlay">
		<div class="kg-spinner" aria-label="Loading graph"></div>
		<span class="kg-overlay-label">Building full hierarchy…</span>
	</div>
{/if}

{#if error}
	<div class="kg-overlay">
		<span class="kg-error-label">{error}</span>
		<button class="kg-retry" onclick={() => initGraph()}>Retry</button>
	</div>
{/if}

<style>
	.kg-root {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 500px;
		background: #fafafa;
		overflow: hidden;
	}

	.kg-stats {
		position: absolute;
		bottom: 16px;
		left: 16px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 12px;
		background: rgba(255, 255, 255, 0.9);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 8px;
		font-size: 11px;
		color: rgba(0, 0, 0, 0.5);
		pointer-events: none;
		z-index: 10;
	}

	.kg-stat-item  { white-space: nowrap; }
	.kg-stat-sep   { color: rgba(0, 0, 0, 0.3); }

	.kg-stat-legend {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 10px;
		color: rgba(0, 0, 0, 0.4);
	}

	.kg-dot {
		display: inline-block;
		width: 7px;
		height: 7px;
		border-radius: 50%;
		margin-left: 4px;
	}

	.kg-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		background: #fafafa;
		z-index: 20;
		pointer-events: none;
	}

	.kg-overlay-label {
		font-size: 13px;
		color: rgba(0, 0, 0, 0.4);
		letter-spacing: 0.03em;
	}

	.kg-error-label {
		font-size: 13px;
		color: rgba(180, 30, 30, 0.7);
		letter-spacing: 0.03em;
	}

	.kg-spinner {
		width: 24px;
		height: 24px;
		border: 2px solid rgba(0, 0, 0, 0.08);
		border-top-color: #7c3aed;
		border-radius: 50%;
		animation: kg-spin 0.8s linear infinite;
	}

	@keyframes kg-spin { to { transform: rotate(360deg); } }

	.kg-retry {
		pointer-events: all;
		padding: 6px 16px;
		font-size: 12px;
		border: 1px solid rgba(0, 0, 0, 0.15);
		border-radius: 6px;
		background: white;
		color: #1a1a1a;
		cursor: pointer;
		transition: background 0.15s;
	}

	.kg-retry:hover {
		background: rgba(124, 58, 237, 0.06);
		border-color: rgba(124, 58, 237, 0.3);
	}
</style>
