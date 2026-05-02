<script lang="ts">
	/**
	 * Pages - BusinessOS Document System
	 * Document listing with sidebar, editor, graph views.
	 * Composes knowledge-base module components with Foundation kb- patterns.
	 */
	import { onMount } from 'svelte';
	import {
		// Stores
		activeDocumentStore,
		sidebarStore,
		documentMetas,
		// Services
		fetchDocuments,
		openAndFetchDocument,
		createDocument,
		deleteDocument,
		// Components
		KBSidebar,
		QuickSearch,
		DocumentEditor
	} from '$lib/modules/knowledge-base';
	import type { DocumentMeta, SidebarView } from '$lib/modules/knowledge-base';
	import NodeDrillDown from '$lib/components/knowledge/NodeDrillDown.svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';
	import { browser } from '$app/environment';

	// State
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let showQuickSearch = $state(false);
	let folderView = $state<{ id: string; title: string; children: DocumentMeta[] } | null>(null);

	// Derived
	let currentDocumentId = $derived($activeDocumentStore.id);
	let currentView = $derived($sidebarStore.view);
	let documents = $derived($documentMetas);

	// LocalStorage for last opened doc
	const LAST_DOC_KEY = 'bos-pages-last-document';

	$effect(() => {
		if (currentDocumentId) {
			localStorage.setItem(LAST_DOC_KEY, currentDocumentId);
		}
	});

	// Recent documents tracking
	const RECENT_KEY = 'bos-recent-docs';
	const MAX_RECENT = 30;

	function addToRecent(id: string, title: string) {
		const stored = JSON.parse(localStorage.getItem(RECENT_KEY) || '[]');
		const filtered = stored.filter((r: { id: string }) => r.id !== id);
		filtered.unshift({ id, title, openedAt: Date.now() });
		localStorage.setItem(RECENT_KEY, JSON.stringify(filtered.slice(0, MAX_RECENT)));
	}

	function getRecent(): Array<{ id: string; title: string; openedAt: number }> {
		return JSON.parse(localStorage.getItem(RECENT_KEY) || '[]');
	}

	function formatTimeAgo(ts: number): string {
		const diff = Date.now() - ts;
		const mins = Math.floor(diff / 60000);
		if (mins < 1) return 'just now';
		if (mins < 60) return `${mins}m ago`;
		const hours = Math.floor(mins / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		return `${days}d ago`;
	}

	// OptimalOS node data for the hierarchy view
	interface NodeInfo { slug: string; name: string; type: string; signal_count: number; has_signals: boolean; }
	let optNodes = $state<NodeInfo[]>([]);

	// Initialize — load documents AND node hierarchy
	onMount(async () => {
		try {
			await fetchDocuments();
		} catch (e) {
			console.warn('[Pages] fetchDocuments failed:', e);
		}

		// Load node list for hierarchy view
		try {
			const headers: Record<string, string> = {};
			const csrf = getCSRFToken();
			if (csrf) headers['X-CSRF-Token'] = csrf;
			const res = await fetch(`${getApiBaseUrl()}/optimal/nodes`, {
				headers, credentials: 'include', signal: AbortSignal.timeout(5000)
			});
			if (res.ok) {
				const data = await res.json();
				optNodes = data.nodes ?? [];
			}
		} catch { /* degrade gracefully */ }

		// Don't auto-open any document — let user choose from hierarchy
		isLoading = false;
	});

	// Handlers
	async function handleNewDocument() {
		try {
			const doc = await createDocument({
				title: '',
				type: 'document'
			});
			folderView = null;
			await openAndFetchDocument(doc.id);
		} catch (e) {
			console.error('Failed to create document:', e);
			error = 'Failed to create document';
		}
	}

	async function handleOpenDocument(id: string) {
		error = null;

		// Check if this is a folder — find it in the document list
		const doc = documents.find(d => d.id === id);
		const isFolder = doc?.type === 'folder' || (!id.endsWith('.md') && id.includes('/'));

		if (isFolder) {
			// Show folder contents instead of opening as document
			const children = documents.filter(d => d.parent_id === id);
			const title = doc?.title || id.split('/').pop() || id;
			folderView = { id, title, children };
			activeDocumentStore.setActiveDocument(null);
			return;
		}

		folderView = null;
		try {
			await openAndFetchDocument(id);
			const title = documents.find(d => d.id === id)?.title || id.split('/').pop() || id;
			addToRecent(id, title);
		} catch (e) {
			console.error('Failed to open document:', e);
			error = 'Failed to open document';
		}
	}

	function handleViewChange(_view: SidebarView) {
		folderView = null;
		activeDocumentStore.setActiveDocument(null);
		error = null;
	}

	function handleCloseDocument() {
		activeDocumentStore.setActiveDocument(null);
	}

	function handleOpenSearch() {
		showQuickSearch = true;
	}

	async function handleDeleteDocument(id: string) {
		try {
			await deleteDocument(id);
			if (currentDocumentId === id) {
				activeDocumentStore.setActiveDocument(null);
			}
		} catch (e) {
			console.error('Failed to delete document:', e);
			error = 'Failed to delete document';
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));
		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function getDocIcon(doc: DocumentMeta): string {
		if (doc.icon && typeof doc.icon === 'string') return doc.icon;
		if (doc.icon && typeof doc.icon === 'object' && 'value' in doc.icon) return doc.icon.value;
		return '';
	}
</script>

<svelte:head>
	<title>Pages | BusinessOS</title>
</svelte:head>

<div class="kb-page">
	<!-- Sidebar -->
	<KBSidebar
		onNewDocument={handleNewDocument}
		onOpenDocument={handleOpenDocument}
		onOpenSearch={handleOpenSearch}
		onViewChange={handleViewChange}
	/>

	<!-- Main Content -->
	<main class="kb-page__main">
		{#if isLoading}
			<div class="kb-page__center">
				<div class="kb-page__spinner"></div>
				<p class="kb-page__center-text">Loading documents...</p>
			</div>
		{:else if error}
			<div class="kb-page__center">
				<p class="kb-page__center-text">{error}</p>
				<button
					class="kb-page__btn"
					aria-label="Retry loading documents"
					onclick={() => { error = null; isLoading = true; fetchDocuments().finally(() => { isLoading = false; }) }}
				>Retry</button>
			</div>
		{:else if currentDocumentId}
			<DocumentEditor
				documentId={currentDocumentId}
				onClose={handleCloseDocument}
			/>
		{:else if folderView}
			<!-- Folder contents view -->
			<div class="kb-page__listing">
				<div class="kb-page__header">
					<h1 class="kb-page__title">{folderView.title}</h1>
					<span style="color: var(--dt3, rgba(255,255,255,0.3)); font-size: 13px;">{folderView.children.length} items</span>
				</div>
				{#if folderView.children.length === 0}
					<div class="kb-page__empty">
						<p class="kb-page__empty-desc">This folder is empty</p>
					</div>
				{:else}
					<div class="kb-page__grid">
						{#each folderView.children as child (child.id)}
							<button
								class="kb-page__card"
								onclick={() => handleOpenDocument(child.id)}
							>
								<div class="kb-page__card-icon">
									{#if child.type === 'folder'}
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/></svg>
									{:else}
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14,2 14,8 20,8"/></svg>
									{/if}
								</div>
								<div class="kb-page__card-body">
									<span class="kb-page__card-title">{child.title || 'Untitled'}</span>
									{#if child.type === 'folder'}
										<span class="kb-page__card-meta">{child.children_count} items</span>
									{/if}
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{:else if currentView === 'graph' || currentView === 'knowledge-graph'}
			<div class="kg-engine-frame">
				{#if browser}
					<iframe
						src="http://localhost:4201/graph"
						title="OptimalEngine Knowledge Graph"
						class="kg-engine-iframe"
						allow="clipboard-read; clipboard-write"
					></iframe>
				{/if}
			</div>
		{:else if currentView === 'recent'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">Recent</h1></div>
				<div class="kb-page__hierarchy">
					{#each getRecent() as item (item.id)}
						<button class="kb-page__child-item kb-page__child-item--recent" onclick={() => handleOpenDocument(item.id)}>
							<span class="kb-page__cat-icon">📄</span>
							<span class="kb-page__child-title">{item.title}</span>
							<span class="kb-page__child-node">{item.id.split('/')[0]}</span>
							<span class="kb-page__child-time">{formatTimeAgo(item.openedAt)}</span>
						</button>
					{:else}
						<div class="kb-page__empty"><p class="kb-page__empty-desc">No recently opened documents</p></div>
					{/each}
				</div>
			</div>
		{:else if currentView === 'favorites'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">Favorites</h1></div>
				<div class="kb-page__empty"><p class="kb-page__empty-desc">Pin documents to see them here</p></div>
			</div>
		{:else if currentView === 'trash'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">Trash</h1></div>
				<div class="kb-page__empty"><p class="kb-page__empty-desc">No deleted documents</p></div>
			</div>
		{:else if currentView === 'profiles' || currentView === 'profiles-person'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">People</h1></div>
				{#each optNodes.filter(n => n.slug === '10-team' || n.slug === '01-roberto') as node}
					<div class="kb-page__node-section">
						<button class="kb-page__node-header" onclick={() => handleOpenDocument(node.slug + '/context.md')}>
							<div class="kb-page__node-info">
								<span class="kb-page__node-slug">{node.slug.split('-')[0]}</span>
								<span class="kb-page__node-name">{node.name}</span>
							</div>
						</button>
						<div class="kb-page__node-children">
							{#each documents.filter(d => d.parent_id === node.slug && (d.type === 'folder' || d.id.endsWith('.md'))) as child}
								<button class="kb-page__child-item" onclick={() => handleOpenDocument(child.id)}>
									<span class="kb-page__cat-icon">{child.type === 'folder' ? '👥' : '📄'}</span>
									<span>{child.title}</span>
								</button>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{:else if currentView === 'profiles-business'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">Businesses</h1></div>
				{#each optNodes.filter(n => ['02-miosa', '03-lunivate'].includes(n.slug)) as node}
					<div class="kb-page__node-section">
						<button class="kb-page__node-header" onclick={() => handleOpenDocument(node.slug + '/context.md')}>
							<div class="kb-page__node-info">
								<span class="kb-page__node-slug">{node.slug.split('-')[0]}</span>
								<span class="kb-page__node-name">{node.name}</span>
							</div>
							<span class="kb-page__type-badge">{node.type}</span>
						</button>
						<div class="kb-page__node-children">
							{#each documents.filter(d => d.parent_id === node.slug) as child}
								<button class="kb-page__child-item" onclick={() => handleOpenDocument(child.id)}>
									<span class="kb-page__cat-icon">{child.type === 'folder' ? '🏢' : '📄'}</span>
									<span>{child.title}</span>
									{#if child.children_count > 0}
										<span class="kb-page__child-count">{child.children_count}</span>
									{/if}
								</button>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{:else if currentView === 'profiles-project'}
			<div class="kb-page__listing">
				<div class="kb-page__header"><h1 class="kb-page__title">Projects</h1></div>
				<!-- Group projects by parent node -->
				{#each optNodes as node}
					{@const projectItems = documents.filter(d =>
						d.id.startsWith(node.slug + '/') && (
							d.id.toLowerCase().includes('project') ||
							d.id.toLowerCase().includes('autonomo') ||
							d.id.toLowerCase().includes('launch') ||
							d.id.toLowerCase().includes('pipeline')
						)
					)}
					{#if projectItems.length > 0}
						<div class="kb-page__node-section">
							<div class="kb-page__node-header" style="cursor: default;">
								<div class="kb-page__node-info">
									<span class="kb-page__node-slug">{node.slug.split('-')[0]}</span>
									<span class="kb-page__node-name">{node.name}</span>
								</div>
								<span class="kb-page__child-count">{projectItems.length}</span>
							</div>
							<div class="kb-page__node-children">
								{#each projectItems as proj}
									<button class="kb-page__child-item" class:kb-page__child-item--folder={proj.type === 'folder'} onclick={() => handleOpenDocument(proj.id)}>
										<span class="kb-page__cat-icon">{proj.type === 'folder' ? '🏗️' : '📄'}</span>
										<span>{proj.title}</span>
										{#if proj.children_count > 0}
											<span class="kb-page__child-count">{proj.children_count}</span>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{:else}
			<!-- Node Drill-Down View — breadcrumb file explorer -->
			<NodeDrillDown
				nodes={optNodes}
				onOpenFile={(filePath) => handleOpenDocument(filePath)}
			/>
		{/if}
	</main>

	<!-- Quick Search Modal -->
	<QuickSearch
		bind:open={showQuickSearch}
		onSelectDocument={handleOpenDocument}
	/>
</div>

<style>
	/* Foundation kb- page patterns with BOS v2 tokens */
	.kb-page {
		display: flex;
		height: 100vh;
		width: 100%;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	.kb-page__main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
		overflow-y: auto;
	}

	/* Center states (loading, error) */
	.kb-page__center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	.kb-page__center-text {
		font-size: 14px;
		margin: 0;
	}

	.kb-page__spinner {
		width: 28px;
		height: 28px;
		border: 3px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
		border-radius: 50%;
		animation: kb-spin 0.8s linear infinite;
	}

	@keyframes kb-spin {
		to { transform: rotate(360deg); }
	}

	/* Listing layout */
	.kb-page__listing {
		flex: 1;
		display: flex;
		flex-direction: column;
		max-width: 900px;
		width: 100%;
		margin: 0 auto;
		padding: 40px 32px;
	}

	.kb-page__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 32px;
	}

	.kb-page__title {
		font-size: 24px;
		font-weight: 600;
		color: var(--bos-v2-text-primary, #121212);
		margin: 0;
	}

	.kb-page__actions {
		display: flex;
		gap: 8px;
	}

	/* Buttons */
	.kb-page__btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 32px;
		padding: 0 14px;
		font-size: 13px;
		font-weight: 500;
		border-radius: 8px;
		border: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
		color: var(--bos-v2-text-primary, #121212);
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s;
	}

	.kb-page__btn:hover {
		background: var(--bos-v2-layer-background-tertiary, #eeeef0);
	}

	.kb-page__btn--primary {
		background: var(--bos-v2-button-primary, #1e96eb);
		color: var(--bos-v2-button-pureWhiteText, #ffffff);
		border-color: transparent;
	}

	.kb-page__btn--primary:hover {
		opacity: 0.9;
	}

	.kb-page__btn--search {
		background: transparent;
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	.kb-page__btn--search:hover {
		color: var(--bos-v2-text-primary, #121212);
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
	}

	/* Empty state */
	.kb-page__empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		flex: 1;
		padding: 64px 32px;
	}

	.kb-page__empty-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 80px;
		height: 80px;
		margin-bottom: 20px;
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
		border-radius: 50%;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
	}

	.kb-page__empty-title {
		font-size: 20px;
		font-weight: 600;
		color: var(--bos-v2-text-primary, #121212);
		margin: 0 0 8px;
	}

	.kb-page__empty-desc {
		font-size: 14px;
		color: var(--bos-v2-text-secondary, #8e8d91);
		margin: 0 0 24px;
		max-width: 320px;
		line-height: 1.5;
	}

	/* Document grid */
	.kb-page__grid {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.kb-page__stats {
		font-size: 12px;
		color: var(--dt3, #999);
	}

	.kb-page__hierarchy {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 0 16px 16px;
	}

	.kb-page__node-section {
		border-radius: 8px;
		overflow: hidden;
	}

	.kb-page__node-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 10px 12px;
		border: none;
		background: var(--dbg2, rgba(0,0,0,0.03));
		border-left: 3px solid var(--dbd, rgba(0,0,0,0.1));
		cursor: pointer;
		transition: all 0.15s;
		color: inherit;
	}

	.kb-page__node-header:hover {
		background: var(--dbg3, rgba(0,0,0,0.06));
		border-left-color: rgba(59,130,246,0.5);
	}

	.kb-page__node-info {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.kb-page__node-slug {
		font-size: 11px;
		font-family: ui-monospace, monospace;
		color: var(--dt4, #aaa);
		min-width: 20px;
	}

	.kb-page__node-name {
		font-size: 13px;
		font-weight: 500;
		color: var(--dt, inherit);
	}

	.kb-page__node-meta {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.kb-page__signal-badge {
		font-size: 10px;
		color: rgba(34,197,94,0.8);
		background: rgba(34,197,94,0.1);
		padding: 2px 6px;
		border-radius: 4px;
	}

	.kb-page__type-badge {
		font-size: 10px;
		color: var(--dt3, #999);
		background: var(--dbg2, rgba(0,0,0,0.05));
		padding: 2px 6px;
		border-radius: 4px;
	}

	.kb-page__node-children {
		display: flex;
		flex-direction: column;
		padding-left: 32px;
		gap: 1px;
	}

	.kb-page__child-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 10px;
		border: none;
		background: transparent;
		color: var(--dt2, #666);
		font-size: 12px;
		cursor: pointer;
		border-radius: 4px;
		transition: all 0.1s;
		text-align: left;
	}

	.kb-page__child-item:hover {
		background: var(--dbg2, rgba(0,0,0,0.04));
		color: var(--dt, inherit);
	}

	.kb-page__child-item--folder {
		color: var(--dt2, #555);
	}

	.kb-page__child-count {
		margin-left: auto;
		font-size: 10px;
		color: var(--dt4, #bbb);
	}

	.kb-page__child-item--special {
		color: var(--dt, inherit);
		font-weight: 500;
	}

	.kb-page__child-desc {
		margin-left: auto;
		font-size: 10px;
		color: var(--dt4, #bbb);
		font-weight: 400;
	}

	.kb-page__child-title {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.kb-page__child-node {
		font-size: 10px;
		color: var(--dt4, #bbb);
		font-weight: 400;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.kb-page__child-time {
		font-size: 10px;
		color: var(--dt3, rgba(0,0,0,0.4));
		font-weight: 400;
		white-space: nowrap;
		flex-shrink: 0;
	}

	/* Knowledge Graph floating panel */
	.kg-split {
		overflow: hidden;
	}

	.kg-engine-frame {
		width: 100%;
		height: 100%;
		overflow: hidden;
		background: #0d0d0d;
	}

	.kg-engine-iframe {
		width: 100%;
		height: 100%;
		border: none;
		background: #0d0d0d;
	}

	.kg-panel {
		position: absolute;
		right: 12px;
		top: 12px;
		bottom: 12px;
		width: 380px;
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.12));
		border-radius: 12px;
		box-shadow: 0 8px 32px rgba(0,0,0,0.15);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		z-index: 10;
		transition: width 0.2s ease;
	}

	.kg-panel--wide {
		width: 55%;
	}

	.kg-panel--full {
		left: 12px;
		width: auto;
	}

	.kg-panel__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		border-bottom: 1px solid var(--dbd, rgba(0,0,0,0.08));
		flex-shrink: 0;
	}

	.kg-panel__title {
		font-size: 12px;
		font-weight: 500;
		color: var(--dt2, #555);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.kg-panel__controls {
		display: flex;
		gap: 4px;
	}

	.kg-panel__btn {
		width: 28px;
		height: 28px;
		border: none;
		background: transparent;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--dt3, #999);
		transition: all 0.1s;
	}

	.kg-panel__btn:hover {
		background: var(--dbg2, rgba(0,0,0,0.05));
		color: var(--dt, inherit);
	}

	.kg-panel__body {
		flex: 1;
		overflow-y: auto;
	}

	.kb-page__cat-icon {
		font-size: 13px;
		flex-shrink: 0;
		width: 18px;
		text-align: center;
	}

	.kb-page__card {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		border-radius: 8px;
		border: none;
		background: transparent;
		cursor: pointer;
		transition: background 0.12s;
		width: 100%;
		text-align: left;
	}

	.kb-page__card:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.kb-page__card--active {
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
	}

	.kb-page__card-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		flex-shrink: 0;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
	}

	.kb-page__card-emoji {
		font-size: 18px;
		line-height: 1;
	}

	.kb-page__card-body {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.kb-page__card-title {
		font-size: 14px;
		font-weight: 500;
		color: var(--bos-v2-text-primary, #121212);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.kb-page__card-meta {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--bos-v2-text-tertiary, #bfbfc3);
	}

	.kb-page__card-badge {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
		color: var(--bos-v2-text-secondary, #8e8d91);
		text-transform: capitalize;
	}

	/* Dark mode */
	:global(.dark) .kb-page {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}

	:global(.dark) .kb-page__main {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}

	:global(.dark) .kb-page__title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .kb-page__spinner {
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
	}

	:global(.dark) .kb-page__btn {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-text-primary, #e6e6e6);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .kb-page__btn:hover {
		background: var(--bos-v2-layer-background-tertiary, #3a3a3a);
	}

	:global(.dark) .kb-page__btn--primary {
		background: var(--bos-v2-button-primary, #1e96eb);
		color: var(--bos-v2-button-pureWhiteText, #ffffff);
		border-color: transparent;
	}

	:global(.dark) .kb-page__btn--search {
		background: transparent;
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	:global(.dark) .kb-page__empty-icon {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-icon-secondary, #707076);
	}

	:global(.dark) .kb-page__empty-title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .kb-page__card:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.04));
	}

	:global(.dark) .kb-page__card--active {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
	}

	:global(.dark) .kb-page__card-title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .kb-page__card-badge {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	/* Hide scrollbar */
	.kb-page__main::-webkit-scrollbar {
		display: none;
	}

	.kb-page__main {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
