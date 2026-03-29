<script lang="ts">
	/**
	 * Pages - BusinessOS Document System
	 * Document listing with sidebar, editor, graph views.
	 * Composes knowledge-base module components with Foundation kb- patterns.
	 */
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		// Stores
		documentsStore,
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
		DocumentEditor,
		GraphView
	} from '$lib/modules/knowledge-base';
	import type { DocumentMeta } from '$lib/modules/knowledge-base';

	// State
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let showQuickSearch = $state(false);
	let showNewDocForm = $state(false);
	let newDocTitle = $state('');

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

	// Initialize
	onMount(async () => {
		try {
			await fetchDocuments();

			if (!$activeDocumentStore.id) {
				const lastDocId = localStorage.getItem(LAST_DOC_KEY);
				if (lastDocId) {
					const docExists = $documentMetas.some(d => d.id === lastDocId);
					if (docExists) {
						await openAndFetchDocument(lastDocId);
					} else {
						const mostRecent = $documentMetas[0];
						if (mostRecent) {
							await openAndFetchDocument(mostRecent.id);
						}
					}
				} else if ($documentMetas.length > 0) {
					await openAndFetchDocument($documentMetas[0].id);
				}
			}

			isLoading = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load documents';
			isLoading = false;
		}
	});

	// Handlers
	async function handleNewDocument() {
		try {
			const doc = await createDocument({
				title: '',
				type: 'document'
			});
			await openAndFetchDocument(doc.id);
		} catch (e) {
			console.error('Failed to create document:', e);
			error = 'Failed to create document';
		}
	}

	async function handleOpenDocument(id: string) {
		try {
			await openAndFetchDocument(id);
		} catch (e) {
			console.error('Failed to open document:', e);
			error = 'Failed to open document';
		}
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
		{:else if currentView === 'graph'}
			<GraphView
				documents={$documentMetas}
				selectedId={null}
				onSelect={(doc) => handleOpenDocument(doc.id)}
				onNavigate={(doc) => handleOpenDocument(doc.id)}
			/>
		{:else if currentDocumentId}
			<DocumentEditor
				documentId={currentDocumentId}
				onClose={handleCloseDocument}
			/>
		{:else}
			<!-- Document listing / empty state -->
			<div class="kb-page__listing">
				<!-- Header -->
				<div class="kb-page__header">
					<h1 class="kb-page__title">Pages</h1>
					<div class="kb-page__actions">
						<button
							class="kb-page__btn kb-page__btn--search"
							aria-label="Search documents"
							onclick={handleOpenSearch}
						>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
							Search
						</button>
						<button
							class="kb-page__btn kb-page__btn--primary"
							aria-label="Create new page"
							onclick={handleNewDocument}
						>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="M5 12h14"/></svg>
							New Page
						</button>
					</div>
				</div>

				{#if documents.length === 0}
					<!-- Empty state -->
					<div class="kb-page__empty">
						<div class="kb-page__empty-icon">
							<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
								<polyline points="14,2 14,8 20,8"/>
								<line x1="16" y1="13" x2="8" y2="13"/>
								<line x1="16" y1="17" x2="8" y2="17"/>
							</svg>
						</div>
						<h2 class="kb-page__empty-title">No pages yet</h2>
						<p class="kb-page__empty-desc">Create your first page to start building your knowledge base.</p>
						<button
							class="kb-page__btn kb-page__btn--primary"
							aria-label="Create first page"
							onclick={handleNewDocument}
						>Create New Page</button>
					</div>
				{:else}
					<!-- Document grid -->
					<div class="kb-page__grid">
						{#each documents as doc (doc.id)}
							<button
								class="kb-page__card"
								class:kb-page__card--active={currentDocumentId === doc.id}
								aria-label="Open {doc.title || 'Untitled'}"
								onclick={() => handleOpenDocument(doc.id)}
							>
								<div class="kb-page__card-icon">
									{#if getDocIcon(doc)}
										<span class="kb-page__card-emoji">{getDocIcon(doc)}</span>
									{:else}
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
											<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
											<polyline points="14,2 14,8 20,8"/>
										</svg>
									{/if}
								</div>
								<div class="kb-page__card-body">
									<span class="kb-page__card-title">{doc.title || 'Untitled'}</span>
									<span class="kb-page__card-meta">
										{formatDate(doc.updated_at)}
										{#if doc.is_favorite}
											<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
										{/if}
									</span>
								</div>
								{#if doc.type}
									<span class="kb-page__card-badge">{doc.type}</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>
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
