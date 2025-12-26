<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { editor, wordCount, type EditorBlock, createEmptyBlock } from '$lib/stores/editor';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';
	import PageMenu from '$lib/components/kb/PageMenu.svelte';
	import { markdownToBlocks, blocksToMarkdown } from '$lib/utils/markdown-blocks';
	import { contexts } from '$lib/stores/contexts';
	import type { Context, ContextListItem } from '$lib/api/client';
	import { embeddings } from '$lib/api/embeddings';
	import { getBackgroundCSS } from '$lib/stores/desktopStore';

	interface Props {
		document: Context | ContextListItem | null;
		isNew?: boolean;
		parentId?: string;
		onSaved?: (doc: Context) => void;
		onTitleChange?: (title: string) => void;
		onPageClick?: (pageId: string) => void;
		allPages?: ContextListItem[];
	}

	let { document: contextDoc, isNew = false, parentId, onSaved, onTitleChange, onPageClick, allPages = [] }: Props = $props();

	// Build breadcrumb chain from current page to root
	const breadcrumbChain = $derived.by(() => {
		if (!contextDoc || !allPages.length) return [];

		const chain: ContextListItem[] = [];
		let currentId = contextDoc.parent_id;

		// Walk up the parent chain
		while (currentId) {
			const parent = allPages.find(p => p.id === currentId);
			if (parent) {
				chain.unshift(parent);
				currentId = parent.parent_id;
			} else {
				break;
			}
		}

		return chain;
	});

	// Default to empty string for new pages, shows "New page" placeholder
	let title = $state(contextDoc?.name || '');
	let saveTimeout: ReturnType<typeof setTimeout> | null = null;
	let isSaving = $state(false);
	let lastSaved = $state<Date | null>(null);
	let hasUnsavedChanges = $state(false);
	let titleInput: HTMLInputElement | null = $state(null);
	let fullDocument = $state<Context | null>(null);
	let showHoverControls = $state(false);
	let icon = $state(contextDoc?.icon || null);
	let coverImage = $state(contextDoc?.cover_image || null);
	let showCoverPicker = $state(false);
	let showIconPicker = $state(false);
	let showCoverHoverControls = $state(false);
	let showPageMenuButton = $state(false);

	// Page settings (loaded from document properties)
	let pageSettings = $derived({
		fullWidth: Boolean(fullDocument?.properties?.fullWidth),
		smallText: Boolean(fullDocument?.properties?.smallText),
		locked: Boolean(fullDocument?.properties?.locked)
	});

	// Solid color covers
	const solidColors = [
		{ value: '#f3f4f6', label: 'Light Gray' },
		{ value: '#e5e7eb', label: 'Gray' },
		{ value: '#d1d5db', label: 'Dark Gray' },
		{ value: '#fef3c7', label: 'Warm Yellow' },
		{ value: '#fde68a', label: 'Yellow' },
		{ value: '#fed7aa', label: 'Orange' },
		{ value: '#fecaca', label: 'Light Red' },
		{ value: '#fbcfe8', label: 'Pink' },
		{ value: '#e9d5ff', label: 'Light Purple' },
		{ value: '#ddd6fe', label: 'Purple' },
		{ value: '#c7d2fe', label: 'Indigo' },
		{ value: '#bfdbfe', label: 'Light Blue' },
		{ value: '#a5f3fc', label: 'Cyan' },
		{ value: '#99f6e4', label: 'Teal' },
		{ value: '#bbf7d0', label: 'Light Green' },
		{ value: '#a7f3d0', label: 'Green' },
	];

	// Gradient covers (CSS gradients)
	const gradientCovers = [
		{ value: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', label: 'Violet' },
		{ value: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)', label: 'Pink Red' },
		{ value: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', label: 'Blue Cyan' },
		{ value: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)', label: 'Green Teal' },
		{ value: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)', label: 'Pink Yellow' },
		{ value: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)', label: 'Purple Pink' },
		{ value: 'linear-gradient(135deg, #ff9a9e 0%, #fad0c4 100%)', label: 'Coral' },
		{ value: 'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)', label: 'Peach' },
		{ value: 'linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%)', label: 'Mint Blue' },
		{ value: 'linear-gradient(135deg, #cfd9df 0%, #e2ebf0 100%)', label: 'Silver' },
		{ value: 'linear-gradient(135deg, #0c0c0c 0%, #3a3a3a 100%)', label: 'Dark' },
		{ value: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)', label: 'Midnight' },
	];

	// Preset cover images (Unsplash)
	const coverPresets = [
		// Nature
		{ url: 'https://images.unsplash.com/photo-1519681393784-d120267933ba?w=1200&h=400&fit=crop', label: 'Mountain' },
		{ url: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=1200&h=400&fit=crop', label: 'Beach' },
		{ url: 'https://images.unsplash.com/photo-1476820865390-c52aeebb9891?w=1200&h=400&fit=crop', label: 'Forest' },
		{ url: 'https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?w=1200&h=400&fit=crop', label: 'Fog Mountains' },
		{ url: 'https://images.unsplash.com/photo-1500534623283-312aade485b7?w=1200&h=400&fit=crop', label: 'Aurora' },
		{ url: 'https://images.unsplash.com/photo-1475924156734-496f6cac6ec1?w=1200&h=400&fit=crop', label: 'Desert' },
		// Abstract
		{ url: 'https://images.unsplash.com/photo-1557682250-33bd709cbe85?w=1200&h=400&fit=crop', label: 'Gradient Purple' },
		{ url: 'https://images.unsplash.com/photo-1557683316-973673baf926?w=1200&h=400&fit=crop', label: 'Gradient Blue' },
		{ url: 'https://images.unsplash.com/photo-1557682224-5b8590cd9ec5?w=1200&h=400&fit=crop', label: 'Gradient Pink' },
		{ url: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1200&h=400&fit=crop', label: 'Waves' },
		{ url: 'https://images.unsplash.com/photo-1550684376-efcbd6e3f031?w=1200&h=400&fit=crop', label: 'Abstract Dark' },
		// Work
		{ url: 'https://images.unsplash.com/photo-1497366216548-37526070297c?w=1200&h=400&fit=crop', label: 'Office' },
		{ url: 'https://images.unsplash.com/photo-1497215728101-856f4ea42174?w=1200&h=400&fit=crop', label: 'Workspace' },
		{ url: 'https://images.unsplash.com/photo-1542744173-8e7e53415bb0?w=1200&h=400&fit=crop', label: 'Meeting' },
		// City
		{ url: 'https://images.unsplash.com/photo-1480714378408-67cf0d13bc1b?w=1200&h=400&fit=crop', label: 'City' },
		{ url: 'https://images.unsplash.com/photo-1514565131-fce0801e5785?w=1200&h=400&fit=crop', label: 'Night City' },
	];

	// Currently selected cover tab
	let coverTab = $state<'colors' | 'gradients' | 'images'>('colors');

	// Preset SVG icons (paths for SVG icons)
	const iconPresets = [
		{ id: 'document', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', label: 'Document' },
		{ id: 'folder', path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', label: 'Folder' },
		{ id: 'clipboard', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2', label: 'Clipboard' },
		{ id: 'chart', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z', label: 'Chart' },
		{ id: 'user', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z', label: 'User' },
		{ id: 'users', path: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z', label: 'Team' },
		{ id: 'building', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', label: 'Building' },
		{ id: 'briefcase', path: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', label: 'Briefcase' },
		{ id: 'lightbulb', path: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z', label: 'Idea' },
		{ id: 'star', path: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z', label: 'Star' },
		{ id: 'target', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z', label: 'Target' },
		{ id: 'check', path: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z', label: 'Check' },
		{ id: 'bookmark', path: 'M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z', label: 'Bookmark' },
		{ id: 'link', path: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1', label: 'Link' },
		{ id: 'chat', path: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z', label: 'Chat' },
		{ id: 'mail', path: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', label: 'Mail' },
		{ id: 'calendar', path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z', label: 'Calendar' },
		{ id: 'clock', path: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z', label: 'Clock' },
		{ id: 'cog', path: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', label: 'Settings' },
		{ id: 'rocket', path: 'M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z', label: 'Rocket' },
		{ id: 'heart', path: 'M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z', label: 'Heart' },
		{ id: 'globe', path: 'M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z', label: 'Globe' },
		{ id: 'database', path: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4', label: 'Database' },
		{ id: 'code', path: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4', label: 'Code' },
	];

	// Helper to get icon path from icon ID
	function getIconPath(iconId: string | null): string | null {
		if (!iconId) return null;
		const found = iconPresets.find(i => i.id === iconId);
		return found?.path || null;
	}

	// Load full document if we only have ContextListItem
	async function loadFullDocument() {
		if (contextDoc && 'blocks' in contextDoc) {
			fullDocument = contextDoc as Context;
		} else if (contextDoc?.id) {
			try {
				fullDocument = await contexts.loadContext(contextDoc.id);
			} catch (error) {
				console.error('Failed to load document:', error);
			}
		}
	}

	// Initialize editor with document content
	onMount(async () => {
		await loadFullDocument();

		if (fullDocument?.blocks && fullDocument.blocks.length > 0) {
			editor.initialize(fullDocument.blocks as EditorBlock[]);
		} else if (fullDocument?.content) {
			const blocks = markdownToBlocks(fullDocument.content);
			editor.initialize(blocks);
		} else {
			editor.initialize([createEmptyBlock()]);
		}

		// Focus title for new documents
		if (isNew && titleInput) {
			titleInput.select();
		}
	});

	// Track the current document ID to detect actual document changes
	let currentDocId = $state<string | null>(null);

	// Reload when document ID changes (not just any prop change)
	$effect(() => {
		const newId = contextDoc?.id || null;
		if (newId && newId !== currentDocId) {
			console.log('[InlineDocumentEditor] Document changed from', currentDocId, 'to:', newId, contextDoc?.name);
			currentDocId = newId;
			// Reset state for new document
			title = contextDoc?.name || '';
			icon = contextDoc?.icon || null;
			coverImage = contextDoc?.cover_image || null;
			hasUnsavedChanges = false;
			// Clear editor and reload
			editor.initialize([]);
			loadFullDocument();
		}
	});

	// Auto-save when editor becomes dirty
	$effect(() => {
		if ($editor.isDirty) {
			hasUnsavedChanges = true;
			// Debounce save
			if (saveTimeout) clearTimeout(saveTimeout);
			saveTimeout = setTimeout(() => {
				saveDocument();
			}, 1500);
		}
	});

	// Cleanup on destroy
	onDestroy(() => {
		if (saveTimeout) clearTimeout(saveTimeout);
		// Save before closing if there are unsaved changes
		if (hasUnsavedChanges && fullDocument) {
			saveDocumentSync();
		}
		editor.reset();
	});

	// Helper to extract text content from a block
	function extractBlockText(block: EditorBlock): string {
		if (typeof block.content === 'string') return block.content;
		// Handle legacy array format if it exists at runtime
		const content = block.content as unknown;
		if (Array.isArray(content)) {
			return (content as Array<{ text?: string }>).map((c) => c.text || '').join(' ');
		}
		return '';
	}

	// Index document for semantic search (non-blocking)
	async function indexDocumentForSearch(docId: string, blocks: EditorBlock[]) {
		try {
			const indexBlocks = blocks
				.map(b => ({
					id: b.id,
					type: b.type,
					content: extractBlockText(b)
				}))
				.filter(b => b.content.trim().length > 0);

			if (indexBlocks.length > 0) {
				await embeddings.indexDocument(docId, indexBlocks);
				console.log('[InlineDocumentEditor] Document indexed for search');
			}
		} catch (err) {
			// Non-blocking - save succeeded, indexing is optional
			console.warn('[InlineDocumentEditor] Failed to index document:', err);
		}
	}

	async function saveDocument() {
		if (isSaving) return;
		isSaving = true;

		try {
			const markdown = blocksToMarkdown($editor.blocks);
			const blocks = $editor.blocks;

			if (isNew && !fullDocument) {
				// Create new document
				const newDoc = await contexts.createContext({
					name: title,
					type: 'document',
					parent_id: parentId,
					content: markdown,
					blocks: blocks
				});
				fullDocument = newDoc;
				if (onSaved) onSaved(newDoc);

				// Index for semantic search (non-blocking)
				indexDocumentForSearch(newDoc.id, blocks);
			} else if (fullDocument) {
				// Update title/metadata separately from blocks
				if (title !== fullDocument.name) {
					await contexts.updateContext(fullDocument.id, { name: title });
					onTitleChange?.(title);
				}
				// Use the proper blocks endpoint for saving block content
				await contexts.updateBlocks(fullDocument.id, blocks, $wordCount);
				fullDocument = { ...fullDocument, name: title, content: markdown, blocks };
				if (onSaved) onSaved(fullDocument);

				// Index for semantic search (non-blocking)
				indexDocumentForSearch(fullDocument.id, blocks);
			}

			editor.markSaved();
			hasUnsavedChanges = false;
			lastSaved = new Date();
		} catch (error) {
			console.error('Failed to save document:', error);
		} finally {
			isSaving = false;
		}
	}

	function saveDocumentSync() {
		const blocks = $editor.blocks;

		if (fullDocument) {
			if (title !== fullDocument.name) {
				contexts.updateContext(fullDocument.id, { name: title }).catch(console.error);
			}
			contexts.updateBlocks(fullDocument.id, blocks, $wordCount).catch(console.error);
		}
	}

	function handleTitleChange(e: Event) {
		title = (e.target as HTMLInputElement).value;
		hasUnsavedChanges = true;
		if (saveTimeout) clearTimeout(saveTimeout);
		saveTimeout = setTimeout(() => {
			saveDocument();
		}, 1500);
	}

	function handleTitleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			// Focus first block
			const firstBlock = document.querySelector('[data-block-id]') as HTMLElement;
			if (firstBlock) {
				firstBlock.focus();
			}
		}
	}

	function addNewBlockAtEnd() {
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			editor.addBlockAfter(lastBlock.id, 'paragraph');
		}
	}

	function formatTime(date: Date) {
		return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
	}

	function getCoverStyle(cover: string | null): string {
		if (!cover) return '';
		if (cover.startsWith('preset:')) {
			const bgId = cover.replace('preset:', '');
			const cssObj = getBackgroundCSS(bgId);
			let style = `background: ${cssObj.background};`;
			if (cssObj.backgroundSize) {
				style += ` background-size: ${cssObj.backgroundSize};`;
			}
			return style;
		}
		return `background-image: url(${cover}); background-size: cover; background-position: center;`;
	}

	// Save icon change to database
	async function updateIcon(newIcon: string | null) {
		icon = newIcon;
		showIconPicker = false;
		if (fullDocument) {
			try {
				await contexts.updateContext(fullDocument.id, { icon: newIcon ?? undefined });
				fullDocument = { ...fullDocument, icon: newIcon };
				// Refresh contexts list so sidebar updates immediately
				await contexts.loadContexts();
			} catch (error) {
				console.error('Failed to update icon:', error);
			}
		}
	}

	// Save cover image change to database
	async function updateCoverImage(newCover: string | null) {
		coverImage = newCover;
		showCoverPicker = false;
		if (fullDocument) {
			try {
				await contexts.updateContext(fullDocument.id, { cover_image: newCover ?? undefined });
				fullDocument = { ...fullDocument, cover_image: newCover };
				// Refresh contexts list so sidebar updates immediately
				await contexts.loadContexts();
			} catch (error) {
				console.error('Failed to update cover image:', error);
			}
		}
	}
</script>

<!-- Full-width Inline Editor -->
<div
	class="inline-document-editor h-full flex flex-col bg-white dark:bg-[#1c1c1e] relative"
	onmouseenter={() => showPageMenuButton = true}
	onmouseleave={() => showPageMenuButton = false}
>
	<!-- Breadcrumb Navigation Bar (like Notion) -->
	{#if breadcrumbChain.length > 0}
		<div class="flex-shrink-0 px-4 py-2 border-b border-gray-100 dark:border-gray-800 flex items-center gap-1 text-sm">
			{#each breadcrumbChain as crumb, idx}
				<button
					onclick={() => onPageClick?.(crumb.id)}
					class="flex items-center gap-1.5 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors max-w-[180px] group"
					title={crumb.name}
				>
					{#if crumb.icon && getIconPath(crumb.icon)}
						<svg class="w-4 h-4 flex-shrink-0 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(crumb.icon)} />
						</svg>
					{:else}
						<svg class="w-4 h-4 flex-shrink-0 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					{/if}
					<span class="truncate">{crumb.name || 'New page'}</span>
				</button>
				<svg class="w-4 h-4 text-gray-300 dark:text-gray-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			{/each}
			<!-- Current page (non-clickable) -->
			<span class="flex items-center gap-1.5 px-2 py-1 text-gray-900 dark:text-white font-medium max-w-[180px]">
				{#if contextDoc?.icon && getIconPath(contextDoc.icon)}
					<svg class="w-4 h-4 flex-shrink-0 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(contextDoc.icon)} />
					</svg>
				{:else}
					<svg class="w-4 h-4 flex-shrink-0 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
				{/if}
				<span class="truncate">{title || 'New page'}</span>
			</span>
		</div>
	{/if}

	<!-- Floating Page Header with Menu -->
	<div class="absolute top-3 right-4 z-40 flex items-center gap-2 transition-opacity {showPageMenuButton ? 'opacity-100' : 'opacity-0'}" style="top: {breadcrumbChain.length > 0 ? '48px' : '12px'};">
		<!-- Page settings menu (three dots) -->
		<PageMenu
			document={fullDocument}
			onSettingsChange={(newSettings) => {
				if (fullDocument) {
					fullDocument = {
						...fullDocument,
						properties: { ...fullDocument.properties, ...newSettings }
					};
				}
			}}
			on:delete={() => {
				// Navigate back or to parent
				window.history.back();
			}}
		/>
	</div>
	<!-- Cover Image (if set) -->
	{#if coverImage}
		<div
			class="relative w-full h-64 bg-gradient-to-br from-blue-100 to-purple-100 dark:from-blue-900/30 dark:to-purple-900/30 group/cover"
			onmouseenter={() => showCoverHoverControls = true}
			onmouseleave={() => showCoverHoverControls = false}
		>
			{#if coverImage.startsWith('preset:')}
				<div class="w-full h-full" style={getCoverStyle(coverImage)}></div>
			{:else if coverImage.startsWith('#')}
				<!-- Solid color cover -->
				<div class="w-full h-full" style="background-color: {coverImage};"></div>
			{:else if coverImage.startsWith('linear-gradient')}
				<!-- Gradient cover -->
				<div class="w-full h-full" style="background: {coverImage};"></div>
			{:else}
				<img src={coverImage} alt="Cover" class="w-full h-full object-cover" />
			{/if}
			<!-- Cover hover controls -->
			<div class="absolute inset-0 flex items-end justify-end p-4 gap-2 transition-opacity {showCoverHoverControls ? 'opacity-100' : 'opacity-0'}">
				<button
					onclick={() => showCoverPicker = !showCoverPicker}
					class="px-3 py-1.5 text-sm bg-white/90 dark:bg-gray-800/90 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-white dark:hover:bg-gray-800 transition-colors shadow-sm"
				>
					Change cover
				</button>
				<button
					onclick={() => updateCoverImage(null)}
					class="px-3 py-1.5 text-sm bg-white/90 dark:bg-gray-800/90 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-white dark:hover:bg-gray-800 transition-colors shadow-sm"
				>
					Remove
				</button>
			</div>
			<!-- Cover picker dropdown with tabs -->
			{#if showCoverPicker}
				<div class="absolute bottom-full right-4 mb-2 w-[420px] bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 p-4 z-50">
					<!-- Tabs -->
					<div class="flex gap-1 mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
						<button
							onclick={() => coverTab = 'colors'}
							class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'colors' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
						>
							Colors
						</button>
						<button
							onclick={() => coverTab = 'gradients'}
							class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'gradients' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
						>
							Gradients
						</button>
						<button
							onclick={() => coverTab = 'images'}
							class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'images' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
						>
							Images
						</button>
					</div>

					{#if coverTab === 'colors'}
						<!-- Solid Colors -->
						<div class="grid grid-cols-8 gap-2">
							{#each solidColors as color}
								<button
									onclick={() => updateCoverImage(color.value)}
									class="w-10 h-10 rounded-lg hover:ring-2 hover:ring-blue-500 hover:ring-offset-2 dark:hover:ring-offset-gray-800 transition-all"
									style="background-color: {color.value};"
									title={color.label}
								></button>
							{/each}
						</div>
					{:else if coverTab === 'gradients'}
						<!-- Gradient Covers -->
						<div class="grid grid-cols-4 gap-2">
							{#each gradientCovers as gradient}
								<button
									onclick={() => updateCoverImage(gradient.value)}
									class="aspect-video rounded-lg hover:ring-2 hover:ring-blue-500 transition-all"
									style="background: {gradient.value};"
									title={gradient.label}
								></button>
							{/each}
						</div>
					{:else}
						<!-- Upload option -->
						<div class="mb-4">
							<label class="flex items-center justify-center gap-2 w-full px-4 py-3 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-blue-500 dark:hover:border-blue-400 transition-colors cursor-pointer group">
								<svg class="w-5 h-5 text-gray-400 group-hover:text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								<span class="text-sm text-gray-500 dark:text-gray-400 group-hover:text-blue-500">Upload image</span>
								<input type="file" accept="image/*" class="hidden" onchange={(e) => {
									const file = (e.target as HTMLInputElement).files?.[0];
									if (file) {
										const reader = new FileReader();
										reader.onload = (ev) => {
											updateCoverImage(ev.target?.result as string);
										};
										reader.readAsDataURL(file);
									}
								}} />
							</label>
						</div>
						<p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Gallery</p>
						<div class="grid grid-cols-4 gap-2 max-h-48 overflow-y-auto">
							{#each coverPresets as preset}
								<button
									onclick={() => updateCoverImage(preset.url)}
									class="aspect-video rounded-lg overflow-hidden hover:ring-2 hover:ring-blue-500 transition-all"
									title={preset.label}
								>
									<img src={preset.url} alt={preset.label} class="w-full h-full object-cover" />
								</button>
							{/each}
						</div>
						<!-- Link option -->
						<div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
							<button
								onclick={() => {
									const url = prompt('Enter image URL:');
									if (url) {
										updateCoverImage(url);
									}
								}}
								class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-blue-500 transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
								</svg>
								<span>Add from URL</span>
							</button>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Editor Content -->
	<div class="flex-1 overflow-y-auto">
		<div class="{pageSettings.fullWidth ? 'max-w-5xl' : 'max-w-3xl'} mx-auto px-8 {coverImage ? 'pt-8' : 'pt-24'} pb-12 {pageSettings.smallText ? 'small-text-mode' : ''}">
			<!-- Hover controls for icon/cover -->
			<div
				class="relative mb-4"
				onmouseenter={() => showHoverControls = true}
				onmouseleave={() => { showHoverControls = false; if (!showIconPicker && !showCoverPicker) { showIconPicker = false; showCoverPicker = false; } }}
			>
				<!-- Icon/Cover buttons (show on hover or if no icon) -->
				<div class="flex items-center gap-2 h-8 {showHoverControls || (!icon && !coverImage) ? 'opacity-100' : 'opacity-0'} transition-opacity">
					{#if icon}
						<div class="relative">
							<button
								onclick={() => showIconPicker = !showIconPicker}
								class="flex items-center gap-1.5 px-2 py-1 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
							>
								{#if getIconPath(icon)}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getIconPath(icon)} />
									</svg>
								{/if}
								<span class="text-xs">Change icon</span>
							</button>
							<button
								onclick={() => updateIcon(null)}
								class="ml-1 px-2 py-1 text-xs text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
							>
								Remove
							</button>
						</div>
					{:else}
						<div class="relative">
							<button
								onclick={() => showIconPicker = !showIconPicker}
								class="flex items-center gap-1.5 px-2 py-1 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								<span>Add icon</span>
							</button>
							<!-- Icon picker dropdown -->
							{#if showIconPicker}
								<div class="absolute top-full left-0 mt-1 w-80 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 p-3 z-50">
									<p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Choose an icon</p>
									<div class="grid grid-cols-6 gap-1">
										{#each iconPresets as iconPreset}
											<button
												onclick={() => updateIcon(iconPreset.id)}
												class="w-10 h-10 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors group"
												title={iconPreset.label}
											>
												<svg class="w-5 h-5 text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={iconPreset.path} />
												</svg>
											</button>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/if}

					{#if !coverImage}
						<div class="relative">
							<button
								onclick={() => showCoverPicker = !showCoverPicker}
								class="flex items-center gap-1.5 px-2 py-1 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								<span>Add cover</span>
							</button>
							<!-- Cover picker dropdown with tabs (when no cover yet) -->
							{#if showCoverPicker}
								<div class="absolute top-full left-0 mt-1 w-[420px] bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 p-4 z-50">
									<!-- Tabs -->
									<div class="flex gap-1 mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
										<button
											onclick={() => coverTab = 'colors'}
											class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'colors' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
										>
											Colors
										</button>
										<button
											onclick={() => coverTab = 'gradients'}
											class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'gradients' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
										>
											Gradients
										</button>
										<button
											onclick={() => coverTab = 'images'}
											class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {coverTab === 'images' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
										>
											Images
										</button>
									</div>

									{#if coverTab === 'colors'}
										<!-- Solid Colors -->
										<div class="grid grid-cols-8 gap-2">
											{#each solidColors as color}
												<button
													onclick={() => updateCoverImage(color.value)}
													class="w-10 h-10 rounded-lg hover:ring-2 hover:ring-blue-500 hover:ring-offset-2 dark:hover:ring-offset-gray-800 transition-all"
													style="background-color: {color.value};"
													title={color.label}
												></button>
											{/each}
										</div>
									{:else if coverTab === 'gradients'}
										<!-- Gradient Covers -->
										<div class="grid grid-cols-4 gap-2">
											{#each gradientCovers as gradient}
												<button
													onclick={() => updateCoverImage(gradient.value)}
													class="aspect-video rounded-lg hover:ring-2 hover:ring-blue-500 transition-all"
													style="background: {gradient.value};"
													title={gradient.label}
												></button>
											{/each}
										</div>
									{:else}
										<!-- Upload option -->
										<div class="mb-4">
											<label class="flex items-center justify-center gap-2 w-full px-4 py-3 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-blue-500 dark:hover:border-blue-400 transition-colors cursor-pointer group">
												<svg class="w-5 h-5 text-gray-400 group-hover:text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
												</svg>
												<span class="text-sm text-gray-500 dark:text-gray-400 group-hover:text-blue-500">Upload image</span>
												<input type="file" accept="image/*" class="hidden" onchange={(e) => {
													const file = (e.target as HTMLInputElement).files?.[0];
													if (file) {
														const reader = new FileReader();
														reader.onload = (ev) => {
															updateCoverImage(ev.target?.result as string);
														};
														reader.readAsDataURL(file);
													}
												}} />
											</label>
										</div>
										<p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Gallery</p>
										<div class="grid grid-cols-4 gap-2 max-h-48 overflow-y-auto">
											{#each coverPresets as preset}
												<button
													onclick={() => updateCoverImage(preset.url)}
													class="aspect-video rounded-lg overflow-hidden hover:ring-2 hover:ring-blue-500 transition-all"
													title={preset.label}
												>
													<img src={preset.url} alt={preset.label} class="w-full h-full object-cover" />
												</button>
											{/each}
										</div>
										<!-- Link option -->
										<div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
											<button
												onclick={() => {
													const url = prompt('Enter image URL:');
													if (url) {
														updateCoverImage(url);
													}
												}}
												class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-blue-500 transition-colors"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
												</svg>
												<span>Add from URL</span>
											</button>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Icon display (if set) -->
			{#if icon && getIconPath(icon)}
				<button
					onclick={() => showIconPicker = !showIconPicker}
					class="mb-4 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-xl p-3 -ml-3 transition-colors cursor-pointer relative"
				>
					<svg class="w-16 h-16 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(icon)} />
					</svg>
					<!-- Icon picker dropdown when clicking on icon -->
					{#if showIconPicker}
						<div class="absolute top-full left-0 mt-1 w-80 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 p-3 z-50" onclick={(e) => e.stopPropagation()}>
							<p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Choose an icon</p>
							<div class="grid grid-cols-6 gap-1">
								{#each iconPresets as iconPreset}
									<button
										onclick={() => { icon = iconPreset.id; showIconPicker = false; }}
										class="w-10 h-10 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors group"
										title={iconPreset.label}
									>
										<svg class="w-5 h-5 text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={iconPreset.path} />
										</svg>
									</button>
								{/each}
							</div>
							<button
								onclick={() => updateIcon(null)}
								class="w-full mt-2 px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
							>
								Remove icon
							</button>
						</div>
					{/if}
				</button>
			{/if}

			<!-- Large Title Input (Notion-style) -->
			<input
				bind:this={titleInput}
				type="text"
				value={title}
				oninput={handleTitleChange}
				onkeydown={handleTitleKeydown}
				readonly={pageSettings.locked}
				class="w-full text-4xl font-bold text-gray-900 dark:text-gray-100 bg-transparent border-0 focus:ring-0 focus:outline-none p-0 mb-8 placeholder:text-gray-400 dark:placeholder:text-gray-500 {pageSettings.locked ? 'cursor-default' : ''}"
				placeholder="New page"
			/>

			<!-- Locked page indicator -->
			{#if pageSettings.locked}
				<div class="mb-4 flex items-center gap-2 px-3 py-2 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800/50 rounded-lg">
					<svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					<span class="text-sm text-amber-700 dark:text-amber-300">This page is locked and cannot be edited</span>
				</div>
			{/if}

			<!-- Blocks -->
			<div class="blocks-container {pageSettings.locked ? 'locked-indicator pointer-events-none' : ''}" role="textbox" tabindex="-1">
				{#each $editor.blocks as block, index (block.id)}
					<BlockComponent {block} {index} readonly={pageSettings.locked} parentContextId={fullDocument?.id} {onPageClick} />
				{/each}
			</div>

			<!-- Click area to add new blocks (hidden when locked) -->
			{#if !pageSettings.locked}
				<button
					onclick={addNewBlockAtEnd}
					class="w-full min-h-32 mt-4 text-left cursor-text group"
				>
					<span class="text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
						Click to add a block, or press / for commands
					</span>
				</button>
			{:else}
				<div class="h-32"></div>
			{/if}
		</div>
	</div>

	<!-- Status Bar -->
	<div class="flex-shrink-0 px-8 py-2 border-t border-gray-200 dark:border-gray-700/50 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
		<div class="flex items-center gap-4">
			<span>{$wordCount} words</span>
			<span>{$editor.blocks.length} blocks</span>
		</div>
		<div class="flex items-center gap-3">
			{#if isSaving}
				<span class="flex items-center gap-1.5">
					<svg class="w-3 h-3 animate-spin" viewBox="0 0 24 24" fill="none">
						<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-opacity="0.3"></circle>
						<path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>
					</svg>
					Saving...
				</span>
			{:else if lastSaved}
				<span>Saved at {formatTime(lastSaved)}</span>
			{:else if hasUnsavedChanges}
				<span class="text-amber-500">Unsaved changes</span>
			{:else}
				<span>No changes</span>
			{/if}
		</div>
	</div>

	<!-- Slash Command Menu -->
	{#if $editor.showSlashMenu && $editor.slashMenuPosition}
		<div class="block-menu-wrapper">
			<BlockMenu />
		</div>
	{/if}
</div>

<style>
	/* Ensure proper focus styles for contenteditable elements */
	.blocks-container :global([contenteditable]:focus) {
		outline: none;
	}

	/* Light mode block styling (default) */
	.blocks-container :global(.block-wrapper) {
		color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]) {
		color: #1f2937;
		caret-color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]:empty::before) {
		color: #9ca3af;
	}

	/* Headings */
	.blocks-container :global(.block-wrapper h1),
	.blocks-container :global(.block-wrapper h2),
	.blocks-container :global(.block-wrapper h3) {
		color: #111827;
	}

	/* Quote */
	.blocks-container :global(.block-wrapper blockquote) {
		border-left-color: #d1d5db;
		color: #6b7280;
	}

	/* Code */
	.blocks-container :global(.block-wrapper pre) {
		background-color: #f3f4f6;
		border-color: #e5e7eb;
	}

	.blocks-container :global(.block-wrapper code) {
		color: #374151;
	}

	/* Divider */
	.blocks-container :global(.block-wrapper hr) {
		border-color: #e5e7eb;
	}

	/* Dark mode block styling */
	:global(.dark) .blocks-container :global(.block-wrapper) {
		color: #f5f5f7;
	}

	:global(.dark) .blocks-container :global(.block-wrapper [contenteditable]) {
		color: #f5f5f7;
		caret-color: #f5f5f7;
	}

	:global(.dark) .blocks-container :global(.block-wrapper [contenteditable]:empty::before) {
		color: #6b7280;
	}

	:global(.dark) .blocks-container :global(.block-wrapper h1),
	:global(.dark) .blocks-container :global(.block-wrapper h2),
	:global(.dark) .blocks-container :global(.block-wrapper h3) {
		color: #ffffff;
	}

	:global(.dark) .blocks-container :global(.block-wrapper blockquote) {
		border-left-color: #4b5563;
		color: #9ca3af;
	}

	:global(.dark) .blocks-container :global(.block-wrapper pre) {
		background-color: #0d0d0d;
		border-color: #374151;
	}

	:global(.dark) .blocks-container :global(.block-wrapper code) {
		color: #e5e7eb;
	}

	:global(.dark) .blocks-container :global(.block-wrapper hr) {
		border-color: #374151;
	}

	/* Small text mode */
	.small-text-mode :global(.block-wrapper) {
		font-size: 14px;
	}

	.small-text-mode :global(.block-wrapper h1) {
		font-size: 1.75rem;
	}

	.small-text-mode :global(.block-wrapper h2) {
		font-size: 1.375rem;
	}

	.small-text-mode :global(.block-wrapper h3) {
		font-size: 1.125rem;
	}

	.small-text-mode input[type="text"] {
		font-size: 2.25rem;
	}

	/* Locked page overlay indicator */
	.locked-indicator {
		background: repeating-linear-gradient(
			45deg,
			transparent,
			transparent 10px,
			rgba(0, 0, 0, 0.02) 10px,
			rgba(0, 0, 0, 0.02) 20px
		);
	}

	:global(.dark) .locked-indicator {
		background: repeating-linear-gradient(
			45deg,
			transparent,
			transparent 10px,
			rgba(255, 255, 255, 0.01) 10px,
			rgba(255, 255, 255, 0.01) 20px
		);
	}
</style>
