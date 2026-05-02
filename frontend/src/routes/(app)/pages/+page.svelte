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
		// Components
		KBSidebar,
		QuickSearch,
		DocumentEditor,
		EmptyPageView,
		FolderContentsView,
		GraphFrame,
		ProfilesView,
		RecentDocumentsView,
		StatusView
	} from '$lib/modules/knowledge-base';
	import type { DocumentMeta, PagesNodeInfo, SidebarView } from '$lib/modules/knowledge-base';
	import {
		addRecentDocument,
		formatRecentTimeAgo,
		readRecentDocuments
	} from '$lib/modules/knowledge-base/utils/recent-documents';
	import NodeDrillDown from '$lib/components/knowledge/NodeDrillDown.svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

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

	function addToRecent(id: string, title: string) {
		addRecentDocument(localStorage, RECENT_KEY, { id, title });
	}

	function getRecent(): Array<{ id: string; title: string; openedAt: number }> {
		return readRecentDocuments(localStorage, RECENT_KEY);
	}

	function formatTimeAgo(ts: number): string {
		return formatRecentTimeAgo(ts);
	}

	// OptimalOS node data for the hierarchy view
	let optNodes = $state<PagesNodeInfo[]>([]);

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

	function retryLoadDocuments() {
		error = null;
		isLoading = true;
		fetchDocuments().finally(() => {
			isLoading = false;
		});
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
			<StatusView state="loading" />
		{:else if error}
			<StatusView state="error" message={error} onRetry={retryLoadDocuments} />
		{:else if currentDocumentId}
			<DocumentEditor
				documentId={currentDocumentId}
				onClose={handleCloseDocument}
			/>
		{:else if folderView}
			<FolderContentsView
				title={folderView.title}
				children={folderView.children}
				onOpenDocument={handleOpenDocument}
			/>
		{:else if currentView === 'graph' || currentView === 'knowledge-graph'}
			<GraphFrame />
		{:else if currentView === 'recent'}
			<RecentDocumentsView
				items={getRecent()}
				{formatTimeAgo}
				onOpenDocument={handleOpenDocument}
			/>
		{:else if currentView === 'favorites'}
			<EmptyPageView title="Favorites" message="Pin documents to see them here" />
		{:else if currentView === 'trash'}
			<EmptyPageView title="Trash" message="No deleted documents" />
		{:else if currentView === 'profiles' || currentView === 'profiles-person' || currentView === 'profiles-business' || currentView === 'profiles-project'}
			<ProfilesView
				view={currentView}
				nodes={optNodes}
				{documents}
				onOpenDocument={handleOpenDocument}
			/>
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
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	.kb-page__main::-webkit-scrollbar {
		display: none;
	}

	:global(.dark) .kb-page,
	:global(.dark) .kb-page__main {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}
</style>
