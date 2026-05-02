<script lang="ts">
	/**
	 * OptimalGraphView — 2-panel layout wrapping SDK KnowledgeGraph + MemoryBrowser.
	 * Left panel: MemoryBrowser (searchable memory sidebar).
	 * Right panel: KnowledgeGraph (SVG force-directed graph from Elixir engine).
	 * Engine is sourced from $lib/optimal-engine; no external fetch calls needed.
	 */
	import { KnowledgeGraph, MemoryBrowser } from '@miosa/optimal-engine';
	import { getEngine } from '$lib/optimal-engine';

	interface NodeSelectPayload {
		id: string;
		name: string;
		type: string;
	}

	interface Props {
		/** Called when a node is selected in the graph so the parent can open details. */
		onNodeSelect?: (node: NodeSelectPayload) => void;
		height?: number;
	}

	let { onNodeSelect, height = 600 }: Props = $props();

	const engine = getEngine();
</script>

<div class="ogv-root">
	<!-- Left sidebar: searchable memory list -->
	<aside class="ogv-sidebar" aria-label="Memory browser">
		<div class="ogv-sidebar__header">
			<span class="ogv-sidebar__title">Memories</span>
		</div>
		<div class="ogv-sidebar__body">
			<MemoryBrowser {engine} limit={100} />
		</div>
	</aside>

	<!-- Right main: force-directed knowledge graph -->
	<section class="ogv-graph" aria-label="Knowledge graph">
		<KnowledgeGraph
			{engine}
			{height}
			onselect={(node) => onNodeSelect?.(node)}
		/>
	</section>
</div>

<style>
	.ogv-root {
		display: flex;
		height: 100%;
		width: 100%;
		overflow: hidden;
		background: var(--dbg, #0d1117);
	}

	/* Sidebar */
	.ogv-sidebar {
		width: 280px;
		min-width: 220px;
		max-width: 340px;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--dbd, rgba(255, 255, 255, 0.08));
		background: var(--dbg, #0d1117);
		overflow: hidden;
	}

	.ogv-sidebar__header {
		display: flex;
		align-items: center;
		padding: 12px 14px 10px;
		border-bottom: 1px solid var(--dbd, rgba(255, 255, 255, 0.06));
		flex-shrink: 0;
	}

	.ogv-sidebar__title {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--dt4, #555);
	}

	.ogv-sidebar__body {
		flex: 1;
		overflow-y: auto;
		padding: 8px 10px;
		scrollbar-width: thin;
		scrollbar-color: var(--dbd, rgba(255, 255, 255, 0.08)) transparent;
	}

	.ogv-sidebar__body::-webkit-scrollbar {
		width: 4px;
	}

	.ogv-sidebar__body::-webkit-scrollbar-thumb {
		background: var(--dbd, rgba(255, 255, 255, 0.08));
		border-radius: 2px;
	}

	/* Graph area */
	.ogv-graph {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
</style>
