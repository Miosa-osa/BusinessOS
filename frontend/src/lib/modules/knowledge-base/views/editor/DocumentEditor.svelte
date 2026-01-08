<script lang="ts">
	import { tick } from 'svelte';
	import { ScrollArea } from '$lib/ui';
	import { activeDocument, activeDocumentStore } from '../../stores/documents';
	import { updateDocument, toggleFavorite, deleteDocument } from '../../services/documents.service';
	import type { Document, Block } from '../../entities/types';
	import EditorHeader from './EditorHeader.svelte';
	import BlockRenderer from './BlockRenderer.svelte';
	import EditorToolbar from './EditorToolbar.svelte';
	import { debounce } from '$lib/utils';

	interface Props {
		documentId?: string;
		readOnly?: boolean;
		onClose?: () => void;
	}

	let { documentId, readOnly = false, onClose }: Props = $props();

	let doc = $derived($activeDocument);
	let isLoading = $derived($activeDocumentStore.loading);
	let isSaving = $derived($activeDocumentStore.saving);
	let lastSaved = $derived($activeDocumentStore.lastSaved);
	let loadError = $derived($activeDocumentStore.error);

	// Local state for editing
	let localTitle = $state('');
	let localContent = $state<Block[]>([]);
	let hasChanges = $state(false);

	// Cover hover state - passed to EditorHeader to show controls
	let isHoveringCover = $state(false);

	// Cover reposition state
	let isRepositioning = $state(false);
	let coverPositionY = $state(50); // 0-100 percentage, default center
	let dragStartY = $state(0);
	let dragStartPositionY = $state(50);
	let coverElement: HTMLDivElement | null = $state(null);

	// Focus management - track which block should receive focus
	let focusBlockId = $state<string | null>(null);

	// Track which document we're editing to detect document switches
	let currentDocId = $state<string | null>(null);

	// Sync local state ONLY when switching to a different document
	// This prevents the save response from overwriting user's in-progress edits
	$effect(() => {
		if (doc && doc.id !== currentDocId) {
			// Document switched - load the new document's content
			currentDocId = doc.id;
			localTitle = doc.title;
			localContent = [...doc.content];
			hasChanges = false;
		}
	});

	// Reset when document is closed/unloaded
	$effect(() => {
		if (!doc && currentDocId) {
			currentDocId = null;
			localTitle = '';
			localContent = [];
			hasChanges = false;
		}
	});

	// Auto-save with debounce
	const debouncedSave = debounce(async () => {
		if (!doc || !hasChanges) return;

		try {
			await updateDocument(doc.id, {
				title: localTitle,
				content: localContent
			});
			hasChanges = false;
		} catch (error) {
			console.error('Failed to save document:', error);
		}
	}, 1000);

	function handleTitleChange(newTitle: string) {
		localTitle = newTitle;
		hasChanges = true;
		debouncedSave();
	}

	function handleBlockChange(index: number, updatedBlock: Block) {
		localContent = localContent.map((block, i) =>
			i === index ? { ...updatedBlock, updated_at: new Date().toISOString() } : block
		);
		hasChanges = true;
		debouncedSave();
	}

	function handleBlockDelete(index: number) {
		if (localContent.length <= 1) return; // Keep at least one block
		localContent = localContent.filter((_, i) => i !== index);
		hasChanges = true;
		debouncedSave();
	}

	async function handleBlockAdd(index: number, newBlock: Block) {
		localContent = [
			...localContent.slice(0, index + 1),
			newBlock,
			...localContent.slice(index + 1)
		];
		hasChanges = true;
		debouncedSave();

		// Focus the new block after it renders
		await tick();
		focusBlockId = newBlock.id;
	}

	function handleBlockFocused() {
		// Clear focus target after block has been focused
		focusBlockId = null;
	}

	function handleBlockMove(fromIndex: number, toIndex: number) {
		if (fromIndex === toIndex) return;

		const newContent = [...localContent];
		const [movedBlock] = newContent.splice(fromIndex, 1);
		newContent.splice(toIndex, 0, movedBlock);

		localContent = newContent;
		hasChanges = true;
		debouncedSave();
	}

	// Handle click on the editor content area (outside blocks)
	async function handleContentAreaClick(e: MouseEvent) {
		if (readOnly) return;

		// Check if click was on empty space (not on a block or interactive element)
		const target = e.target as HTMLElement;
		const clickedOnBlock = target.closest('.block, .block-wrapper, .editor-header, button');

		if (!clickedOnBlock) {
			// Click was on empty space - focus or create the last block
			if (localContent.length > 0) {
				// Focus the last block
				const lastBlock = localContent[localContent.length - 1];
				focusBlockId = lastBlock.id;
			} else {
				// No blocks, create one
				const newBlock: Block = {
					id: crypto.randomUUID(),
					type: 'paragraph',
					content: [],
					properties: {},
					children: [],
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};
				localContent = [newBlock];
				hasChanges = true;
				debouncedSave();
				await tick();
				focusBlockId = newBlock.id;
			}
		}
	}

	function handleIconChange(icon: string | null) {
		if (!doc) return;
		updateDocument(doc.id, { icon });
	}

	function handleCoverChange(cover: string | null) {
		if (!doc) return;
		updateDocument(doc.id, { cover });
	}

	// Cover reposition handlers
	function handleStartReposition() {
		if (isRepositioning) {
			// Clicking "Done" - exit reposition mode
			isRepositioning = false;
		} else {
			// Start repositioning
			isRepositioning = true;
		}
	}

	function handleCoverMouseDown(e: MouseEvent) {
		if (!isRepositioning) return;
		e.preventDefault();
		dragStartY = e.clientY;
		dragStartPositionY = coverPositionY;

		// Add global listeners
		window.addEventListener('mousemove', handleCoverMouseMove);
		window.addEventListener('mouseup', handleCoverMouseUp);
	}

	function handleCoverMouseMove(e: MouseEvent) {
		if (!isRepositioning || !coverElement) return;

		const coverHeight = coverElement.offsetHeight;
		const deltaY = e.clientY - dragStartY;
		// Convert pixel movement to percentage (inverted: drag down = lower position value)
		const deltaPercent = (deltaY / coverHeight) * 100;
		const newPositionY = Math.max(0, Math.min(100, dragStartPositionY - deltaPercent));

		coverPositionY = newPositionY;
	}

	function handleCoverMouseUp() {
		window.removeEventListener('mousemove', handleCoverMouseMove);
		window.removeEventListener('mouseup', handleCoverMouseUp);
		// TODO: Persist coverPositionY to document when backend supports it
	}

	async function handleToggleFavorite() {
		if (!doc) return;
		try {
			await toggleFavorite(doc.id);
		} catch (error) {
			console.error('Failed to toggle favorite:', error);
		}
	}

	function handleShare() {
		// TODO: Implement share modal
		console.log('Share document:', doc?.id);
	}

	async function handleDelete() {
		if (!doc) return;
		if (!confirm('Are you sure you want to delete this document?')) return;

		try {
			await deleteDocument(doc.id);
			onClose?.();
		} catch (error) {
			console.error('Failed to delete document:', error);
		}
	}
</script>

<div class="document-editor">
	{#if isLoading}
		<div class="document-editor__loading">
			<div class="document-editor__spinner"></div>
			<p>Loading document...</p>
		</div>
	{:else if loadError}
		<div class="document-editor__error">
			<p>Failed to load document</p>
			<p class="document-editor__error-detail">{loadError}</p>
		</div>
	{:else if doc}
		<EditorToolbar
			{isSaving}
			{lastSaved}
			{hasChanges}
			{readOnly}
			isFavorite={doc.is_favorite}
			{onClose}
			onToggleFavorite={handleToggleFavorite}
			onShare={handleShare}
			onDelete={handleDelete}
		/>

		<ScrollArea class="document-editor__content">
			<!-- Cover rendered outside container for full-width -->
			{#if doc.cover}
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="document-editor__cover"
					class:document-editor__cover--repositioning={isRepositioning}
					bind:this={coverElement}
					onmouseenter={() => isHoveringCover = true}
					onmouseleave={() => isHoveringCover = false}
					onmousedown={handleCoverMouseDown}
				>
					{#if doc.cover.startsWith('linear-gradient') || (doc.cover.startsWith('#') && doc.cover.length === 7)}
						<div class="document-editor__cover-bg" style="background: {doc.cover}"></div>
					{:else}
						<img
							src={doc.cover}
							alt="Cover"
							class="document-editor__cover-img"
							style="object-position: center {coverPositionY}%"
							draggable="false"
						/>
					{/if}
					{#if isRepositioning}
						<div class="document-editor__reposition-hint">
							<span>Drag to reposition</span>
						</div>
					{/if}
				</div>
			{/if}
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div class="document-editor__container" onclick={handleContentAreaClick}>
				<EditorHeader
					title={localTitle}
					icon={doc.icon}
					cover={doc.cover}
					{readOnly}
					{isHoveringCover}
					{isRepositioning}
					onTitleChange={handleTitleChange}
					onIconChange={handleIconChange}
					onCoverChange={handleCoverChange}
					onStartReposition={handleStartReposition}
				/>

				<div class="document-editor__blocks">
					{#each localContent as block, index (block.id)}
						{@const blockText = block.content.map(rt => rt.plain_text || '').join('')}
						{@const isOnlyEmptyBlock = localContent.length === 1 && blockText === ''}
						<BlockRenderer
							{block}
							{index}
							{readOnly}
							shouldFocus={focusBlockId === block.id}
							isFirstBlock={index === 0}
							{isOnlyEmptyBlock}
							totalBlocks={localContent.length}
							onBlockChange={(updated) => handleBlockChange(index, updated)}
							onBlockDelete={() => handleBlockDelete(index)}
							onBlockAdd={(newBlock) => handleBlockAdd(index, newBlock)}
							onFocused={handleBlockFocused}
							onBlockMove={handleBlockMove}
						/>
					{/each}
				</div>

					<!-- Empty space acts as click target for adding blocks at the end -->
				<div class="document-editor__footer"></div>
			</div>
		</ScrollArea>
	{:else}
		<div class="document-editor__empty">
			<p>Select a document to start editing</p>
		</div>
	{/if}
</div>

<style>
	.document-editor {
		display: flex;
		flex-direction: column;
		height: 100%;
		background-color: hsl(var(--background));
	}

	.document-editor__content {
		flex: 1;
		overflow: hidden;
	}

	/* Full-width cover outside container */
	.document-editor__cover {
		position: relative;
		width: 100%;
		height: 200px;
		overflow: hidden;
	}

	.document-editor__cover-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.document-editor__cover-bg {
		width: 100%;
		height: 100%;
	}

	/* Reposition mode */
	.document-editor__cover--repositioning {
		cursor: ns-resize;
		user-select: none;
	}

	.document-editor__cover--repositioning::after {
		content: '';
		position: absolute;
		inset: 0;
		background: hsl(var(--foreground) / 0.1);
		pointer-events: none;
	}

	.document-editor__reposition-hint {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		padding: 0.5rem 1rem;
		background: hsl(var(--background) / 0.9);
		border: 1px solid hsl(var(--border));
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: hsl(var(--foreground));
		pointer-events: none;
		box-shadow: 0 2px 8px hsl(var(--foreground) / 0.1);
	}

	.document-editor__container {
		max-width: 900px;
		margin: 0 auto;
		padding: 0 4rem 8rem;
		min-height: 100%;
		cursor: text;
	}

	.document-editor__blocks {
		display: flex;
		flex-direction: column;
		min-height: 200px;
	}

	.document-editor__footer {
		min-height: 300px;
		cursor: text;
	}

	.document-editor__empty {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: hsl(var(--muted-foreground));
	}

	.document-editor__loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 1rem;
		color: hsl(var(--muted-foreground));
	}

	.document-editor__spinner {
		width: 32px;
		height: 32px;
		border: 3px solid hsl(var(--border));
		border-top-color: hsl(var(--primary));
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.document-editor__error {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 0.5rem;
		color: hsl(var(--destructive));
	}

	.document-editor__error-detail {
		font-size: 0.875rem;
		color: hsl(var(--muted-foreground));
	}
</style>
