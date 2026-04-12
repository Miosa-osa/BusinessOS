<script lang="ts">
	/**
	 * Document Detail Page
	 * Direct link to a specific document — loads editor with sidebar.
	 */
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		activeDocumentStore,
		openAndFetchDocument,
		createDocument,
		KBSidebar,
		QuickSearch,
		DocumentEditor
	} from '$lib/modules/knowledge-base';

	// Get document ID from URL
	let documentId = $derived($page.params.id);

	// State
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let showQuickSearch = $state(false);

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
		goto(`/pages/${id}`);
	}

	async function handleNewDocument() {
		try {
			const doc = await createDocument({
				title: '',
				type: 'document'
			});
			goto(`/pages/${doc.id}`);
		} catch (e) {
			console.error('Failed to create document:', e);
		}
	}

	function handleCloseDocument() {
		goto('/pages');
	}

	function handleOpenSearch() {
		showQuickSearch = true;
	}
</script>

<svelte:head>
	<title>Document | Pages</title>
</svelte:head>

<div class="kb-doc">
	<KBSidebar
		onNewDocument={handleNewDocument}
		onOpenDocument={handleOpenDocument}
		onOpenSearch={handleOpenSearch}
	/>

	<main class="kb-doc__main">
		{#if isLoading}
			<div class="kb-doc__center">
				<div class="kb-doc__spinner"></div>
				<p class="kb-doc__center-text">Loading document...</p>
			</div>
		{:else if error}
			<div class="kb-doc__center">
				<p class="kb-doc__center-text">{error}</p>
				<button
					class="kb-doc__btn"
					aria-label="Back to pages"
					onclick={() => goto('/pages')}
				>Back to Pages</button>
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
	.kb-doc {
		display: flex;
		height: 100vh;
		width: 100%;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
	}

	.kb-doc__main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.kb-doc__center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	.kb-doc__center-text {
		font-size: 14px;
		margin: 0;
	}

	.kb-doc__spinner {
		width: 28px;
		height: 28px;
		border: 3px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
		border-radius: 50%;
		animation: kb-doc-spin 0.8s linear infinite;
	}

	@keyframes kb-doc-spin {
		to { transform: rotate(360deg); }
	}

	.kb-doc__btn {
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
		transition: background 0.15s;
	}

	.kb-doc__btn:hover {
		background: var(--bos-v2-layer-background-tertiary, #eeeef0);
	}

	/* Dark mode */
	:global(.dark) .kb-doc {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}

	:global(.dark) .kb-doc__spinner {
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
	}

	:global(.dark) .kb-doc__btn {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-text-primary, #e6e6e6);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .kb-doc__btn:hover {
		background: var(--bos-v2-layer-background-tertiary, #3a3a3a);
	}
</style>
