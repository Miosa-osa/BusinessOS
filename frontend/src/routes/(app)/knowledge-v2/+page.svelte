<script lang="ts">
	/**
	 * Pages - BusinessOS Document System
	 * Clean, modular Notion-like document system with BusinessOS styling
	 *
	 * This is a thin orchestrator that composes the knowledge-base module components.
	 * All business logic lives in the module stores and services.
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
		// Components
		KBSidebar,
		QuickSearch,
		DocumentEditor,
		GraphView
	} from '$lib/modules/knowledge-base';
	import { KnowledgeGraph, KnowledgeChatPanel, KnowledgeDocumentPanel } from '$lib/components/knowledge';
	import type { Memory } from '$lib/api/memory/types';
	import type { DocumentMeta } from '$lib/modules/knowledge-base';

	// Convert DocumentMeta to Memory format for the KnowledgeGraph
	function documentsToMemories(docs: DocumentMeta[]): Memory[] {
		return docs.map((doc) => ({
			id: doc.id,
			user_id: '',
			title: doc.title || 'Untitled',
			summary: doc.title || 'Untitled',
			content: '',
			memory_type: 'context' as const,
			importance_score: doc.is_favorite ? 0.8 : 0.5,
			is_pinned: doc.is_favorite,
			is_active: !doc.is_archived,
			tags: [],
			metadata: {},
			source_type: 'document',
			source_id: doc.id,
			project_id: null,
			node_id: doc.parent_id,
			expires_at: null,
			access_count: 0,
			last_accessed_at: null,
			created_at: doc.updated_at,
			updated_at: doc.updated_at,
			icon: typeof doc.icon === 'string' ? doc.icon : doc.icon?.value || null,
			color: null,
			cover_image: null
		}));
	}

	// Memories derived from documents for the graph
	let graphMemories = $derived(documentsToMemories($documentMetas));
	let selectedGraphId = $state<string | null>(null);

	// Selected memory for the document panel
	let selectedMemory = $derived(
		selectedGraphId ? graphMemories.find(m => m.id === selectedGraphId) || null : null
	);

	// Chat panel state
	let showChatPanel = $state(false);
	let chatMessages = $state<Array<{ id: string; role: 'user' | 'assistant'; content: string; timestamp: Date }>>([]);

	// Check embed mode
	const isEmbed = $derived($page.url.searchParams.get('embed') === 'true');

	// Quick search state
	let showQuickSearch = $state(false);

	// Loading state
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Current document ID from URL or selection
	let currentDocumentId = $derived($activeDocumentStore.id);

	// Current sidebar view
	let currentView = $derived($sidebarStore.view);

	// LocalStorage key for last opened document
	const LAST_DOC_KEY = 'bos-pages-last-document';

	// Save last opened document to localStorage
	$effect(() => {
		if (currentDocumentId) {
			localStorage.setItem(LAST_DOC_KEY, currentDocumentId);
		}
	});

	// Initialize
	onMount(async () => {
		try {
			await fetchDocuments();

			// Restore last opened document if none is currently active
			if (!$activeDocumentStore.id) {
				const lastDocId = localStorage.getItem(LAST_DOC_KEY);
				if (lastDocId) {
					// Check if the document still exists in the loaded documents
					const docExists = $documentMetas.some(d => d.id === lastDocId);
					if (docExists) {
						await openAndFetchDocument(lastDocId);
					} else {
						// Last document doesn't exist anymore, open most recent
						const mostRecent = $documentMetas[0]; // Sorted by updated_at
						if (mostRecent) {
							await openAndFetchDocument(mostRecent.id);
						}
					}
				} else if ($documentMetas.length > 0) {
					// No last document, open most recent
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
		}
	}

	async function handleOpenDocument(id: string) {
		try {
			await openAndFetchDocument(id);
		} catch (e) {
			console.error('Failed to open document:', e);
		}
	}

	function handleCloseDocument() {
		activeDocumentStore.setActiveDocument(null);
	}

	function handleOpenSearch() {
		showQuickSearch = true;
	}
</script>

<svelte:head>
	<title>Pages | BusinessOS</title>
</svelte:head>

<div class="knowledge-page" class:knowledge-page--embed={isEmbed}>
	<!-- Sidebar -->
	<KBSidebar
		onNewDocument={handleNewDocument}
		onOpenDocument={handleOpenDocument}
		onOpenSearch={handleOpenSearch}
	/>

	<!-- Main Content -->
	<main class="knowledge-page__main">
		{#if isLoading}
			<div class="knowledge-page__loading">
				<div class="knowledge-page__spinner"></div>
				<p>Loading documents...</p>
			</div>
		{:else if error}
			<div class="knowledge-page__error">
				<p>{error}</p>
				<button class="knowledge-page__btn" onclick={() => window.location.reload()}>Retry</button>
			</div>
		{:else if currentView === 'graph'}
			<!-- Force-directed Graph View for Pages -->
			<GraphView
				documents={$documentMetas}
				selectedId={selectedGraphId}
				onSelect={(doc) => {
					selectedGraphId = doc.id;
				}}
				onNavigate={(doc) => {
					handleOpenDocument(doc.id);
				}}
			/>
		{:else if currentView === 'knowledge-graph'}
			<!-- 3D Sphere Knowledge Graph View with Panels -->
			<div class="knowledge-graph-layout">
				<!-- Chat Panel (Left) -->
				{#if showChatPanel}
					<div class="knowledge-panel knowledge-panel--chat">
						<KnowledgeChatPanel
							messages={chatMessages}
							streaming={false}
							onSend={(message) => {
								chatMessages = [...chatMessages, {
									id: crypto.randomUUID(),
									role: 'user',
									content: message,
									timestamp: new Date()
								}];
							}}
							onClose={() => showChatPanel = false}
						/>
					</div>
				{/if}

				<!-- 3D Graph -->
				<div class="knowledge-graph-container">
					<KnowledgeGraph
						memories={graphMemories}
						selectedId={selectedGraphId}
						onSelect={(memory) => {
							selectedGraphId = memory.id;
						}}
						onDeselect={() => {
							selectedGraphId = null;
						}}
					/>

					<!-- Toggle Chat Button -->
					<button
						class="knowledge-toggle-chat"
						onclick={() => showChatPanel = !showChatPanel}
						title={showChatPanel ? 'Hide Chat' : 'Show Chat'}
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</button>
				</div>

				<!-- Document Panel (Right) -->
				{#if selectedMemory}
					<div class="knowledge-panel knowledge-panel--document">
						<KnowledgeDocumentPanel
							selectedMemory={selectedMemory}
							onClose={() => selectedGraphId = null}
							onEdit={() => handleOpenDocument(selectedMemory!.id)}
						/>
					</div>
				{/if}
			</div>
		{:else if currentView.startsWith('profiles')}
			<!-- Context Profiles View -->
			<div class="knowledge-page__profiles">
				<div class="knowledge-page__profiles-content">
					<div class="knowledge-page__profiles-icon">
						<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
							<circle cx="9" cy="7" r="4"/>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"/>
						</svg>
					</div>
					<h2 class="knowledge-page__profiles-title">
						{#if currentView === 'profiles'}
							All Profiles
						{:else if currentView === 'profiles-person'}
							People
						{:else if currentView === 'profiles-business'}
							Businesses
						{:else if currentView === 'profiles-project'}
							Projects
						{/if}
					</h2>
					<p class="knowledge-page__profiles-description">
						Context profiles help you organize information about people, businesses, and projects.
						Link documents to profiles to build a knowledge graph.
					</p>
					<p class="knowledge-page__profiles-hint">
						Coming soon - Context profiles will be integrated with your documents.
					</p>
				</div>
			</div>
		{:else if currentDocumentId}
			<DocumentEditor
				documentId={currentDocumentId}
				onClose={handleCloseDocument}
			/>
		{:else}
			<!-- Empty state / Home view -->
			<div class="knowledge-page__empty">
				<div class="knowledge-page__empty-content">
					<div class="knowledge-page__empty-icon">
						<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
							<polyline points="14,2 14,8 20,8"/>
							<line x1="16" y1="13" x2="8" y2="13"/>
							<line x1="16" y1="17" x2="8" y2="17"/>
							<polyline points="10,9 9,9 8,9"/>
						</svg>
					</div>
					<h2 class="knowledge-page__empty-title">Welcome to Pages</h2>
					<p class="knowledge-page__empty-description">
						Create your first page or select one from the sidebar to get started.
					</p>
					<button class="knowledge-page__btn knowledge-page__btn--primary" onclick={handleNewDocument}>
						Create New Page
					</button>
				</div>
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
	/* BusinessOS-style Knowledge Page */
	.knowledge-page {
		display: flex;
		height: 100vh;
		width: 100%;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
		font-family: var(--bos-font-family);
	}

	.knowledge-page--embed {
		border-radius: 8px;
		overflow: hidden;
	}

	.knowledge-page__main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
	}

	.knowledge-page__loading,
	.knowledge-page__error {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	.knowledge-page__spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.knowledge-page__btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		height: 32px;
		padding: 0 16px;
		font-size: var(--bos-font-sm, 14px);
		font-weight: 500;
		border-radius: 8px;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
		background-color: var(--bos-v2-button-secondary, #f4f4f5);
		color: var(--bos-v2-text-primary, #121212);
	}

	.knowledge-page__btn:hover {
		background-color: var(--bos-v2-layer-background-tertiary, #eeeef0);
	}

	.knowledge-page__btn--primary {
		background-color: var(--bos-v2-button-primary, #1e96eb);
		color: var(--bos-v2-button-pureWhiteText, #ffffff);
	}

	.knowledge-page__btn--primary:hover {
		opacity: 0.9;
	}

	.knowledge-page__empty {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
	}

	.knowledge-page__empty-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		max-width: 400px;
		padding: 32px;
	}

	.knowledge-page__empty-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 96px;
		height: 96px;
		margin-bottom: 24px;
		background-color: var(--bos-v2-layer-background-secondary, #f4f4f5);
		border-radius: 50%;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
	}

	.knowledge-page__empty-title {
		font-size: 24px;
		font-weight: 600;
		margin-bottom: 8px;
		color: var(--bos-v2-text-primary, #121212);
	}

	.knowledge-page__empty-description {
		font-size: var(--bos-font-sm, 14px);
		color: var(--bos-v2-text-secondary, #8e8d91);
		margin-bottom: 24px;
		line-height: 1.5;
	}

	/* Dark mode */
	:global(.dark) .knowledge-page {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}

	:global(.dark) .knowledge-page__main {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
	}

	:global(.dark) .knowledge-page__spinner {
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
		border-top-color: var(--bos-brand-color, #1e96eb);
	}

	:global(.dark) .knowledge-page__btn {
		background-color: var(--bos-v2-button-secondary, #3a3a3a);
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .knowledge-page__btn:hover {
		background-color: var(--bos-v2-layer-background-tertiary, #3a3a3a);
	}

	:global(.dark) .knowledge-page__empty-icon {
		background-color: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-icon-secondary, #707076);
	}

	:global(.dark) .knowledge-page__empty-title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .knowledge-page__empty-description {
		color: var(--bos-v2-text-secondary, #8e8d91);
	}

	/* Knowledge Graph Layout */
	.knowledge-graph-layout {
		display: flex;
		height: 100%;
		width: 100%;
		position: relative;
	}

	.knowledge-graph-container {
		flex: 1;
		position: relative;
		min-width: 0;
	}

	.knowledge-panel {
		flex-shrink: 0;
		height: 100%;
		background: white;
		border-left: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
	}

	.knowledge-panel--chat {
		width: 380px;
		border-left: none;
		border-right: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
	}

	.knowledge-panel--document {
		width: 450px;
	}

	.knowledge-toggle-chat {
		position: absolute;
		top: 16px;
		left: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		background: var(--bos-v2-layer-background-primary, #ffffff);
		border: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		border-radius: 10px;
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		transition: all 0.15s;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
		z-index: 10;
	}

	.knowledge-toggle-chat:hover {
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
		color: var(--bos-v2-icon-activated, #1e96eb);
	}

	/* Dark mode for knowledge graph layout */
	:global(.dark) .knowledge-panel {
		background: var(--bos-v2-layer-background-primary, #1e1e1e);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .knowledge-toggle-chat {
		background: var(--bos-v2-layer-background-primary, #1e1e1e);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .knowledge-toggle-chat:hover {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
	}

	/* Profiles View */
	.knowledge-page__profiles {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
	}

	.knowledge-page__profiles-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		max-width: 450px;
		padding: 32px;
	}

	.knowledge-page__profiles-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 96px;
		height: 96px;
		margin-bottom: 24px;
		background-color: var(--bos-v2-layer-background-secondary, #f4f4f5);
		border-radius: 50%;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
	}

	.knowledge-page__profiles-title {
		font-size: 24px;
		font-weight: 600;
		margin-bottom: 8px;
		color: var(--bos-v2-text-primary, #121212);
	}

	.knowledge-page__profiles-description {
		font-size: var(--bos-font-sm, 14px);
		color: var(--bos-v2-text-secondary, #8e8d91);
		margin-bottom: 16px;
		line-height: 1.5;
	}

	.knowledge-page__profiles-hint {
		font-size: 13px;
		color: var(--bos-v2-text-tertiary, #a9a9ad);
		padding: 12px 16px;
		background-color: var(--bos-v2-layer-background-secondary, #f4f4f5);
		border-radius: 8px;
	}

	:global(.dark) .knowledge-page__profiles-icon {
		background-color: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-icon-secondary, #707076);
	}

	:global(.dark) .knowledge-page__profiles-title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .knowledge-page__profiles-hint {
		background-color: var(--bos-v2-layer-background-secondary, #2c2c2c);
	}
</style>
