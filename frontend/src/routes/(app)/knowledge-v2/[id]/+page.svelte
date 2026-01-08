<script lang="ts">
	/**
	 * Document Detail Page
	 * Direct link to a specific document
	 */
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		activeDocumentStore,
		openAndFetchDocument,
		KBSidebar,
		QuickSearch,
		DocumentEditor
	} from '$lib/modules/knowledge-base';

	// Get document ID from URL params
	let documentId = $derived($page.params.id);

	// Quick search state
	let showQuickSearch = $state(false);

	// Loading state
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Load document on mount
	onMount(async () => {
		if (documentId) {
			try {
				await openAndFetchDocument(documentId);
				isLoading = false;
			} catch (e) {
				error = e instanceof Error ? e.message : 'Document not found';
				isLoading = false;
			}
		}
	});

	// Handlers
	function handleOpenDocument(id: string) {
		goto(`/knowledge-v2/${id}`);
	}

	function handleNewDocument() {
		goto('/knowledge-v2');
	}

	function handleCloseDocument() {
		goto('/knowledge-v2');
	}

	function handleOpenSearch() {
		showQuickSearch = true;
	}
</script>

<svelte:head>
	<title>Document | Knowledge Base</title>
</svelte:head>

<div class="document-page">
	<KBSidebar
		onNewDocument={handleNewDocument}
		onOpenDocument={handleOpenDocument}
		onOpenSearch={handleOpenSearch}
	/>

	<main class="document-page__main">
		{#if isLoading}
			<div class="document-page__loading">
				<div class="document-page__spinner"></div>
				<p>Loading document...</p>
			</div>
		{:else if error}
			<div class="document-page__error">
				<p>{error}</p>
				<button onclick={() => goto('/knowledge-v2')}>Back to Knowledge Base</button>
			</div>
		{:else}
			<DocumentEditor
				{documentId}
				onClose={handleCloseDocument}
			/>
		{/if}
	</main>

	<QuickSearch
		bind:open={showQuickSearch}
		onSelectDocument={handleOpenDocument}
	/>
</div>

<style>
	.document-page {
		display: flex;
		height: 100vh;
		width: 100%;
		background-color: hsl(var(--background));
	}

	.document-page__main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.document-page__loading,
	.document-page__error {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 1rem;
		color: hsl(var(--muted-foreground));
	}

	.document-page__spinner {
		width: 32px;
		height: 32px;
		border: 3px solid hsl(var(--muted-foreground) / 0.2);
		border-top-color: hsl(var(--primary));
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.document-page__error button {
		padding: 0.5rem 1rem;
		background-color: hsl(var(--secondary));
		color: hsl(var(--secondary-foreground));
		border: none;
		border-radius: 0.375rem;
		cursor: pointer;
	}
</style>
