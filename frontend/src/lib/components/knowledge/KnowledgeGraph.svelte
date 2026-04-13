<script lang="ts">
	import { onMount } from 'svelte';
	// force-graph ships incorrect .d.ts (declares class; runtime is a Kapsule factory).
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	import _FG from 'force-graph';
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const ForceGraph = _FG as unknown as () => (el: HTMLElement) => any;

	import { getApiBaseUrl } from '$lib/api/base';
	import type { Memory } from '$lib/api/memory/types';

	// ── Props ─────────────────────────────────────────────────────────────────
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
		selectedId = null,       // eslint-disable-line @typescript-eslint/no-unused-vars
		searchQuery = '',        // eslint-disable-line @typescript-eslint/no-unused-vars
		highlightedIds = [],     // eslint-disable-line @typescript-eslint/no-unused-vars
		nodes: nodeMap = new Map(), // eslint-disable-line @typescript-eslint/no-unused-vars
		zoomLevel = 50,          // eslint-disable-line @typescript-eslint/no-unused-vars
		onZoomChange
	}: Props = $props();

	// ── API types ─────────────────────────────────────────────────────────────
	interface ApiNode   { slug: string; name: string; type: string; signal_count: number; }
	interface ApiFile   { name: string; path: string; is_dir: boolean; size: number; children?: ApiFile[]; }
	interface ApiEntity { name: string; type: string; connections: number; }
	interface ApiEdge   { source: string; target: string; relation: string; weight?: number; }

	type GNodeType = 'core' | 'folder' | 'document' | 'entity';

	interface GNode {
		id: string; label: string; nodeType: GNodeType; val: number;
		connections?: number; entityType?: string;
		x?: number; y?: number; fx?: number | null; fy?: number | null;
		_maxConn?: number;
	}
	interface GLink { source: string | GNode; target: string | GNode; relation: string; }

	// ── Design constants ──────────────────────────────────────────────────────
	const C_CORE = '#1a1a1a', C_ENTITY = '#3b82f6', C_FOLDER = '#6b7280';
	const C_DOC  = '#9ca3af', C_HOV    = '#7c3aed', C_CON    = '#7c3aedbb';
	const L_DEF  = 'rgba(0,0,0,0.08)', L_HOV = 'rgba(124,58,237,0.5)', L_FD = 'rgba(0,0,0,0.02)';

	// ── Component state ───────────────────────────────────────────────────────
	let container = $state<HTMLDivElement | null>(null);
	let loading   = $state(true);
	let error     = $state<string | null>(null);
	let statsText = $state('');

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let graph: any = null;

	// ── Hover state (plain JS — NOT $state — runs at 60fps) ──────────────────
	let _hovId = '';
	const _connN = new Set<string>();
	const _connL = new Set<string>();
	const _adj   = new Map<string, Array<{ src: string; tgt: string }>>();

	function setHover(node: { id: string } | null) {
		_hovId = node?.id ?? '';
		_connN.clear(); _connL.clear();
		if (node) {
			for (const e of (_adj.get(node.id) ?? [])) {
				_connN.add(e.src); _connN.add(e.tgt);
				_connL.add(`${e.src}__${e.tgt}`);
			}
		}
	}

	function linkHighlighted(link: GLink): boolean {
		const s = typeof link.source === 'object' ? link.source.id : link.source;
		const t = typeof link.target === 'object' ? link.target.id : link.target;
		return _connL.has(`${s}__${t}`) || _connL.has(`${t}__${s}`);
	}

	function buildAdj(links: GLink[]) {
		_adj.clear();
		for (const lk of links) {
			const s = typeof lk.source === 'string' ? lk.source : lk.source.id;
			const t = typeof lk.target === 'string' ? lk.target : lk.target.id;
			if (!_adj.has(s)) _adj.set(s, []);
			if (!_adj.has(t)) _adj.set(t, []);
			_adj.get(s)!.push({ src: s, tgt: t });
			_adj.get(t)!.push({ src: s, tgt: t });
		}
	}

	// ── Node canvas renderer (force2d reference pattern) ─────────────────────
	function paintNode(node: GNode & { x: number; y: number }, ctx: CanvasRenderingContext2D, gs: number) {
		const { id, nodeType: nType = 'document', x, y, connections: conn = 0, label } = node;
		const maxC  = node._maxConn ?? 10;
		const r     = 2 + (conn / Math.max(maxC, 1)) * 4;  // range 2–6px
		const isHov = id === _hovId;
		const isCon = _connN.has(id);
		const faded = _hovId !== '' && !isHov && !isCon;

		const baseCol = nType === 'core' ? C_CORE : nType === 'entity' ? C_ENTITY
		              : nType === 'folder' ? C_FOLDER : C_DOC;

		// Glow halo (from force2d reference)
		if (isHov || isCon) {
			ctx.beginPath();
			ctx.arc(x, y, r + (isHov ? 5 : 3), 0, Math.PI * 2);
			ctx.fillStyle = isHov ? 'rgba(124,58,237,0.25)' : 'rgba(124,58,237,0.12)';
			ctx.fill();
		}

		// Main circle
		ctx.beginPath();
		ctx.arc(x, y, r, 0, Math.PI * 2);
		ctx.fillStyle = faded ? baseCol + '1a' : isHov ? C_HOV : isCon ? C_CON : baseCol;
		ctx.fill();

		// Specular highlight (from force2d reference)
		if (!faded) {
			ctx.beginPath();
			ctx.arc(x - r * 0.28, y - r * 0.28, r * 0.32, 0, Math.PI * 2);
			ctx.fillStyle = isHov ? 'rgba(255,255,255,0.55)' : 'rgba(255,255,255,0.22)';
			ctx.fill();
		}

		// Stroke ring for core + hovered
		if (isHov || nType === 'core') {
			ctx.beginPath();
			ctx.arc(x, y, r, 0, Math.PI * 2);
			ctx.strokeStyle = isHov ? C_HOV : baseCol + '55';
			ctx.lineWidth   = (isHov ? 1.5 : 1) / gs;
			ctx.stroke();
		}

		// Labels: hover, connected, core at gs>1.5, entity at gs>3, all at gs>5
		if (isHov || isCon || (nType === 'core' && gs > 1.5) || (nType === 'entity' && gs > 3) || gs > 5) {
			const fs   = nType === 'core' ? 12 / gs : isHov ? 11 / gs : 9 / gs;
			const bold = nType === 'core' || isHov;
			ctx.font         = `${bold ? '600 ' : ''}${fs}px -apple-system, system-ui, sans-serif`;
			ctx.textBaseline = 'middle';
			ctx.textAlign    = 'left';
			ctx.fillStyle    = isHov ? '#1a1a1a' : isCon ? '#374151'
			                 : nType === 'core' ? '#111827' : nType === 'entity' ? '#1d4ed8'
			                 : 'rgba(0,0,0,0.4)';
			ctx.fillText(label ?? id, x + r + 2 / gs, y);
		}
	}

	// ── Apply tuned forces (Reference 1 physics) ─────────────────────────────
	function tuneForces() {
		graph.d3Force('charge')?.strength(-120)?.distanceMax(400);
		graph.d3Force('link')?.distance(70)?.strength(0.5);
		graph.d3Force('center')?.strength(0.05);
		graph.d3Force('collide', null);
		graph.d3Force('radial',  null);
	}

	// ── Shared link callbacks ─────────────────────────────────────────────────
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	function applyLinkConfig(g: any) {
		return g
			.linkColor((l: GLink) => linkHighlighted(l) ? L_HOV : _hovId ? L_FD : L_DEF)
			.linkWidth((l: GLink) => linkHighlighted(l) ? 1.5 : _hovId ? 0.2 : 0.5)
			.linkCurvature(0.18)
			.linkDirectionalParticles((l: GLink) => _hovId && linkHighlighted(l) ? 4 : 0)
			.linkDirectionalParticleWidth(1.5)
			.linkDirectionalParticleColor(() => 'rgba(124,58,237,0.8)')
			.linkDirectionalParticleSpeed(0.004);
	}

	// ── Shared drag callbacks ─────────────────────────────────────────────────
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	function applyDragConfig(g: any) {
		return g
			.onNodeDrag((node: GNode & { x: number; y: number }) => {
				node.fx = node.x; node.fy = node.y;
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				(graph as any)?._simulation?.alphaTarget(0.3).restart();
				setHover(node);
				if (container) container.style.cursor = 'grabbing';
			})
			.onNodeDragEnd((node: GNode) => {
				node.fx = null; node.fy = null;
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				(graph as any)?._simulation?.alphaTarget(0);
				setHover(null);
				if (container) container.style.cursor = 'default';
			});
	}

	// ── File tree walker ──────────────────────────────────────────────────────
	function walkTree(files: ApiFile[], parentId: string, gNodes: GNode[], gLinks: GLink[]) {
		for (const f of files) {
			const type: GNodeType = f.is_dir ? 'folder' : 'document';
			gNodes.push({ id: f.path, label: f.name, nodeType: type, val: f.is_dir ? 1.5 : 1 });
			gLinks.push({ source: parentId, target: f.path, relation: 'contains' });
			if (f.is_dir && f.children?.length) walkTree(f.children, f.path, gNodes, gLinks);
		}
	}

	// ── Main fetch + render ───────────────────────────────────────────────────
	async function initGraph() {
		if (!container) return;
		loading = true; error = null;

		try {
			const base = getApiBaseUrl();
			const [nodesRes, graphRes] = await Promise.all([
				fetch(`${base}/optimal/nodes`),
				fetch(`${base}/optimal/graph`)
			]);
			if (!nodesRes.ok) throw new Error(`/optimal/nodes → HTTP ${nodesRes.status}`);
			if (!graphRes.ok) throw new Error(`/optimal/graph → HTTP ${graphRes.status}`);

			const nodesData  = await nodesRes.json();
			const coreNodes: ApiNode[] = Array.isArray(nodesData) ? nodesData : (nodesData.nodes ?? []);
			const graphRaw   = await graphRes.json();
			const graphData: { entities: ApiEntity[]; edges: ApiEdge[] } = {
				entities: graphRaw.entities ?? [],
				edges:    graphRaw.edges    ?? []
			};

			const fileTreeResults = await Promise.allSettled(
				coreNodes.map(n =>
					fetch(`${base}/optimal/nodes/${n.slug}/files`)
						.then(r => r.ok ? r.json() : Promise.resolve({ files: [] }))
						.then((d: { files: ApiFile[] }) => ({ slug: n.slug, files: d.files ?? [] }))
				)
			);

			const gNodes: GNode[] = [];
			const gLinks: GLink[] = [];
			const slugSet = new Set(coreNodes.map(n => n.slug));

			for (const n of coreNodes) {
				gNodes.push({ id: n.slug, label: n.name, nodeType: 'core', val: 2 });
			}

			for (const res of fileTreeResults) {
				if (res.status === 'fulfilled') walkTree(res.value.files, res.value.slug, gNodes, gLinks);
			}

			const entityIds = new Set<string>();
			let maxConnections = 1;
			for (const e of graphData.entities) {
				if (!entityIds.has(e.name)) {
					entityIds.add(e.name);
					if (e.connections > maxConnections) maxConnections = e.connections;
					gNodes.push({ id: e.name, label: e.name, nodeType: 'entity', val: 2,
						connections: e.connections, entityType: e.type });
				}
			}

			// Core–core sibling links
			const coreList = coreNodes.map(n => n.slug);
			for (let i = 0; i < coreList.length; i++)
				for (let j = i + 1; j < coreList.length; j++)
					gLinks.push({ source: coreList[i], target: coreList[j], relation: 'sibling' });

			// Entity–core links
			for (const e of graphData.edges) {
				if (entityIds.has(e.source) && slugSet.has(e.target))
					gLinks.push({ source: e.source, target: e.target, relation: 'belongs_to' });
				if (entityIds.has(e.target) && slugSet.has(e.source))
					gLinks.push({ source: e.target, target: e.source, relation: 'belongs_to' });
			}

			// Deduplicated cross-ref edges
			const seenLinks = new Set<string>();
			function addLink(s: string, t: string, rel: string) {
				const key = `${s}||${t}||${rel}`;
				if (!seenLinks.has(key)) { seenLinks.add(key); gLinks.push({ source: s, target: t, relation: rel }); }
			}
			for (const e of graphData.edges) {
				const sc = slugSet.has(e.source), tc = slugSet.has(e.target);
				const se = entityIds.has(e.source), te = entityIds.has(e.target);
				if (e.relation === 'cross_ref') {
					if (sc && tc) addLink(e.source, e.target, 'cross_ref');
					if (se && te) addLink(e.source, e.target, 'cross_ref');
					if (se && tc) addLink(e.source, e.target, 'mentioned_in');
					if (te && sc) addLink(e.target, e.source, 'mentioned_in');
				} else if (e.relation === 'mentioned_in' || e.relation === 'lives_in') {
					if (se && tc) addLink(e.source, e.target, 'mentioned_in');
					if (te && sc) addLink(e.target, e.source, 'mentioned_in');
				}
			}

			// Remove isolated nodes
			const connectedIds = new Set<string>();
			for (const lk of gLinks) {
				connectedIds.add(typeof lk.source === 'string' ? lk.source : lk.source.id);
				connectedIds.add(typeof lk.target === 'string' ? lk.target : lk.target.id);
			}
			const filteredNodes = gNodes.filter(n => connectedIds.has(n.id));
			for (const n of filteredNodes) {
				n._maxConn = maxConnections;
				const angle = Math.random() * Math.PI * 2;
				const rad = Math.random() * 50;
				n.x = Math.cos(angle) * rad;
				n.y = Math.sin(angle) * rad;
			}

			statsText = `${filteredNodes.length} nodes · ${gLinks.length} edges`;
			buildAdj(gLinks);

			destroyGraph();
			container.innerHTML = '';

			graph = ForceGraph()(container)
				.graphData({ nodes: filteredNodes, links: gLinks })
				.backgroundColor('#fafafa')
				.nodeVal('val')
				.nodeLabel('')
				.autoPauseRedraw(false)
				.minZoom(0.1).maxZoom(16)
				.nodeCanvasObject((n: GNode & { x: number; y: number }, c: CanvasRenderingContext2D, g: number) => paintNode(n, c, g))
				.nodeCanvasObjectMode(() => 'replace')
				.nodePointerAreaPaint((n: GNode & { x: number; y: number }, color: string, c: CanvasRenderingContext2D) => {
					const r = Math.max(8, 2 + (n.connections ?? 0) / Math.max(n._maxConn ?? 10, 1) * 4 + 6);
					c.beginPath(); c.arc(n.x, n.y, r, 0, Math.PI * 2);
					c.fillStyle = color; c.fill();
				})
				.warmupTicks(400)
				.cooldownTime(15000)
				.d3AlphaDecay(0.02)
				.d3AlphaMin(0.001)
				.d3VelocityDecay(0.4)
				.enableNodeDrag(true)
				.onZoom(({ k }: { k: number }) => {
					onZoomChange?.(Math.max(0, Math.min(100, Math.round(((k - 0.1) / 15.9) * 100))));
				})
				.onNodeHover((n: GNode | null) => {
					setHover(n);
					if (container) container.style.cursor = n ? 'grab' : 'default';
				})
				.onNodeClick((n: GNode & { x: number; y: number }) => {
					graph?.centerAt(n.x, n.y, 600);
					graph?.zoom(4, 600);
					setHover(n);
					const docId = n.nodeType === 'core' ? `${n.id}/context.md` : n.id;
					onSelect?.({
						id: docId, user_id: 'graph', title: n.label ?? n.id,
						summary: `${n.nodeType} — ${n.connections ?? 0} connections`,
						content: '', memory_type: 'context',
						importance_score: (n.connections ?? 0) / 200,
						is_pinned: false, is_active: true, tags: [],
						metadata: { entity_type: n.entityType ?? n.nodeType },
						source_type: null, source_id: null, project_id: null, node_id: null,
						expires_at: null, access_count: 0, last_accessed_at: null, created_at: '', updated_at: ''
					});
				})
				.onBackgroundClick(() => onDeselect?.());

			applyLinkConfig(graph);
			applyDragConfig(graph);

			const w = container.clientWidth, h = container.clientHeight;
			if (w > 0 && h > 0) graph.width(w).height(h);
			tuneForces();
			setTimeout(() => { graph?.zoomToFit(400, 30); }, 500);
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

		const maxConn = Math.max(...memories.map(m => Math.round((m.importance_score ?? 0.5) * 100)), 1);
		const gNodes: GNode[] = memories.map(m => ({
			id: m.id, label: m.title ?? m.id, nodeType: 'document' as GNodeType,
			val: Math.log((m.importance_score ?? 0.5) * 100 + 1),
			connections: Math.round((m.importance_score ?? 0.5) * 100),
			_maxConn: maxConn
		}));

		const gLinks: GLink[] = [];
		for (let i = 0; i < memories.length; i++)
			for (let j = i + 1; j < memories.length; j++)
				if (memories[i].node_id && memories[i].node_id === memories[j].node_id)
					gLinks.push({ source: memories[i].id, target: memories[j].id, relation: 'co_node' });

		buildAdj(gLinks);
		destroyGraph();
		container.innerHTML = '';

		graph = ForceGraph()(container)
			.graphData({ nodes: gNodes, links: gLinks })
			.backgroundColor('#fafafa')
			.nodeVal('val').nodeLabel('').autoPauseRedraw(false)
			.nodeCanvasObject((n: GNode & { x: number; y: number }, c: CanvasRenderingContext2D, g: number) => paintNode(n, c, g))
			.nodeCanvasObjectMode(() => 'replace')
			.d3AlphaDecay(0.02).d3VelocityDecay(0.4)
			.onNodeHover((n: GNode | null) => {
				setHover(n);
				if (container) container.style.cursor = n ? 'grab' : 'default';
			})
			.onNodeClick((n: GNode) => { const mem = memories.find(m => m.id === n.id); if (mem) onSelect?.(mem); })
			.onBackgroundClick(() => onDeselect?.());

		applyLinkConfig(graph);
		applyDragConfig(graph);
		tuneForces();
		graph.width(container.clientWidth).height(container.clientHeight);
		statsText = `${memories.length} memories · ${gLinks.length} edges`;
		loading   = false;
	}

	// ── Mount ─────────────────────────────────────────────────────────────────
	let ro: ResizeObserver | null = null;

	onMount(() => {
		if (memories.length === 0) initGraph(); else buildFromMemories();

		ro = new ResizeObserver(() => {
			if (graph && container) graph.width(container.clientWidth).height(container.clientHeight);
		});
		if (container) ro.observe(container);

		return () => { ro?.disconnect(); destroyGraph(); };
	});

	function destroyGraph() {
		if (!graph) return;
		try { graph._destructor?.(); } catch { /* ignore */ }
		graph = null;
	}
</script>

<div class="kg-root" bind:this={container}></div>

{#if statsText && !loading}
	<div class="kg-stats" aria-label="Graph statistics">
		<span class="kg-stat-item">{statsText}</span>
		<span class="kg-stat-sep">·</span>
		<span class="kg-stat-legend">
			<span class="kg-dot" style="background:{C_CORE}"></span>core
			<span class="kg-dot" style="background:{C_ENTITY}"></span>entity
			<span class="kg-dot" style="background:{C_FOLDER}"></span>folder
			<span class="kg-dot" style="background:{C_DOC}"></span>doc
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
		position: relative; width: 100%; height: 100%;
		min-height: 500px; background: #fafafa; overflow: hidden;
	}
	.kg-stats {
		position: absolute; bottom: 16px; left: 16px;
		display: flex; align-items: center; gap: 8px;
		padding: 6px 12px;
		background: rgba(255,255,255,0.9); backdrop-filter: blur(8px);
		border: 1px solid rgba(0,0,0,0.08); border-radius: 8px;
		font-size: 11px; color: rgba(0,0,0,0.5);
		pointer-events: none; z-index: 10;
	}
	.kg-stat-item  { white-space: nowrap; }
	.kg-stat-sep   { color: rgba(0,0,0,0.3); }
	.kg-stat-legend {
		display: flex; align-items: center; gap: 5px;
		font-size: 10px; color: rgba(0,0,0,0.4);
	}
	.kg-dot {
		display: inline-block; width: 7px; height: 7px;
		border-radius: 50%; margin-left: 4px;
	}
	.kg-overlay {
		position: absolute; inset: 0;
		display: flex; flex-direction: column; align-items: center; justify-content: center;
		gap: 12px; background: #fafafa; z-index: 20; pointer-events: none;
	}
	.kg-overlay-label { font-size: 13px; color: rgba(0,0,0,0.4); letter-spacing: 0.03em; }
	.kg-error-label   { font-size: 13px; color: rgba(180,30,30,0.7); letter-spacing: 0.03em; }
	.kg-spinner {
		width: 24px; height: 24px;
		border: 2px solid rgba(0,0,0,0.08); border-top-color: #7c3aed;
		border-radius: 50%; animation: kg-spin 0.8s linear infinite;
	}
	@keyframes kg-spin { to { transform: rotate(360deg); } }
	.kg-retry {
		pointer-events: all; padding: 6px 16px; font-size: 12px;
		border: 1px solid rgba(0,0,0,0.15); border-radius: 6px;
		background: white; color: #1a1a1a; cursor: pointer; transition: background 0.15s;
	}
	.kg-retry:hover {
		background: rgba(124,58,237,0.06);
		border-color: rgba(124,58,237,0.3);
	}
</style>
