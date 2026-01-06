<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { editor, type EditorBlock, type BlockType, type EditorState } from '$lib/stores/editor';
	import { contexts } from '$lib/stores/contexts';
	import Block from './Block.svelte';

	interface Props {
		block: EditorBlock;
		index: number;
		readonly?: boolean;
		parentContextId?: string;
		onPageClick?: (pageId: string) => void;
	}

	let { block, index, readonly = false, parentContextId, onPageClick }: Props = $props();

	let blockElement: HTMLElement | null = $state(null);
	let showLanguagePicker = $state(false);
	let languageSearchQuery = $state('');
	let languagePickerRef: HTMLDivElement | null = $state(null);

	// Drag and drop state
	let isDragging = $state(false);
	let isDragOver = $state(false);
	let dragOverPosition = $state<'above' | 'below' | null>(null);

	// Drag handlers
	function handleDragStart(e: DragEvent) {
		isDragging = true;
		e.dataTransfer?.setData('text/plain', block.id);
		e.dataTransfer!.effectAllowed = 'move';
		// Add dragging class after a frame for visual feedback
		requestAnimationFrame(() => {
			const wrapper = (e.target as HTMLElement).closest('.block-wrapper');
			wrapper?.classList.add('dragging');
		});
	}

	function handleDragEnd(e: DragEvent) {
		isDragging = false;
		const wrapper = (e.target as HTMLElement).closest('.block-wrapper');
		wrapper?.classList.remove('dragging');
		// Clear all drag-over states
		document.querySelectorAll('.block-wrapper').forEach(el => {
			el.classList.remove('drag-over-above', 'drag-over-below');
		});
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		e.dataTransfer!.dropEffect = 'move';

		// Determine if dropping above or below
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const midY = rect.top + rect.height / 2;
		const position = e.clientY < midY ? 'above' : 'below';

		// Update visual indicator
		const wrapper = e.currentTarget as HTMLElement;
		wrapper.classList.remove('drag-over-above', 'drag-over-below');
		wrapper.classList.add(`drag-over-${position}`);

		isDragOver = true;
		dragOverPosition = position;
	}

	function handleDragLeave(e: DragEvent) {
		const wrapper = e.currentTarget as HTMLElement;
		wrapper.classList.remove('drag-over-above', 'drag-over-below');
		isDragOver = false;
		dragOverPosition = null;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		const sourceBlockId = e.dataTransfer?.getData('text/plain');
		if (!sourceBlockId || sourceBlockId === block.id) return;

		const wrapper = e.currentTarget as HTMLElement;
		wrapper.classList.remove('drag-over-above', 'drag-over-below');

		// Get source and target indices
		editor.update((s: EditorState) => {
			const sourceIdx = s.blocks.findIndex((b: EditorBlock) => b.id === sourceBlockId);
			const targetIdx = s.blocks.findIndex((b: EditorBlock) => b.id === block.id);
			if (sourceIdx === -1 || targetIdx === -1) return s;

			const newBlocks = [...s.blocks];
			const [movedBlock] = newBlocks.splice(sourceIdx, 1);

			// Calculate final insert position
			let insertIdx = targetIdx;
			if (dragOverPosition === 'below') {
				insertIdx = sourceIdx < targetIdx ? targetIdx : targetIdx + 1;
			} else {
				insertIdx = sourceIdx < targetIdx ? targetIdx - 1 : targetIdx;
			}

			newBlocks.splice(insertIdx, 0, movedBlock);

			return { ...s, blocks: newBlocks, isDirty: true };
		});

		isDragOver = false;
		dragOverPosition = null;
	}

	// Programming languages for code blocks
	const LANGUAGES = [
		{ id: 'plain', label: 'Plain Text' },
		{ id: 'javascript', label: 'JavaScript' },
		{ id: 'typescript', label: 'TypeScript' },
		{ id: 'python', label: 'Python' },
		{ id: 'go', label: 'Go' },
		{ id: 'rust', label: 'Rust' },
		{ id: 'java', label: 'Java' },
		{ id: 'c', label: 'C' },
		{ id: 'cpp', label: 'C++' },
		{ id: 'csharp', label: 'C#' },
		{ id: 'ruby', label: 'Ruby' },
		{ id: 'php', label: 'PHP' },
		{ id: 'swift', label: 'Swift' },
		{ id: 'kotlin', label: 'Kotlin' },
		{ id: 'sql', label: 'SQL' },
		{ id: 'html', label: 'HTML' },
		{ id: 'css', label: 'CSS' },
		{ id: 'json', label: 'JSON' },
		{ id: 'yaml', label: 'YAML' },
		{ id: 'markdown', label: 'Markdown' },
		{ id: 'bash', label: 'Bash' },
		{ id: 'shell', label: 'Shell' },
		{ id: 'dockerfile', label: 'Dockerfile' },
		{ id: 'graphql', label: 'GraphQL' },
		{ id: 'svelte', label: 'Svelte' },
		{ id: 'vue', label: 'Vue' },
		{ id: 'jsx', label: 'JSX' },
		{ id: 'tsx', label: 'TSX' },
		{ id: 'xml', label: 'XML' },
		{ id: 'toml', label: 'TOML' },
	];

	// Filtered languages based on search
	let filteredLanguages = $derived(
		languageSearchQuery
			? LANGUAGES.filter(lang =>
				lang.label.toLowerCase().includes(languageSearchQuery.toLowerCase()) ||
				lang.id.toLowerCase().includes(languageSearchQuery.toLowerCase())
			)
			: LANGUAGES
	);

	// Track if we're initializing to prevent immediate updates
	let initialized = false;

	onMount(() => {
		// Set initial content
		if (blockElement && block.content) {
			blockElement.innerText = block.content;
		}
		initialized = true;
	});

	// Update element content when block changes (only if not focused)
	$effect(() => {
		if (blockElement && initialized && document.activeElement !== blockElement) {
			blockElement.innerText = block.content || '';
		}
	});

	// Focus this block when it becomes the focused block in the store
	$effect(() => {
		if ($editor.focusedBlockId === block.id && blockElement && document.activeElement !== blockElement) {
			blockElement.focus();
			// Move cursor to end
			const range = document.createRange();
			const sel = window.getSelection();
			if (blockElement.childNodes.length > 0) {
				range.setStartAfter(blockElement.lastChild!);
			} else {
				range.setStart(blockElement, 0);
			}
			range.collapse(true);
			sel?.removeAllRanges();
			sel?.addRange(range);
		}
	});

	// Handle pending block type selection from BlockMenu
	$effect(() => {
		const pending = $editor.pendingBlockTypeSelection;
		if (pending && pending.blockId === block.id) {
			// Clear first to prevent re-triggering
			editor.clearPendingBlockTypeSelection();
			// Then handle the selection (async, but we've already cleared the pending state)
			selectBlockType(pending.type, pending.properties);
		}
	});

	function handleFocus() {
		editor.setFocusedBlock(block.id);
	}

	function handleBlur(e: FocusEvent) {
		// Don't blur if clicking on slash menu
		const relatedTarget = e.relatedTarget as HTMLElement;
		if (relatedTarget?.closest('[data-slash-menu]')) {
			return;
		}

		// Don't hide if menu is still showing (might be a click in progress)
		// The menu will hide itself after selection
		if ($editor.showSlashMenu) {
			// Still save content but don't hide menu
			if (blockElement) {
				const newContent = blockElement.innerText || '';
				if (newContent !== block.content) {
					editor.updateBlock(block.id, newContent);
				}
			}
			return;
		}

		// Save content on blur
		if (blockElement) {
			const newContent = blockElement.innerText || '';
			if (newContent !== block.content) {
				editor.updateBlock(block.id, newContent);
			}
		}
	}

	function handleInput(e: Event) {
		if (!blockElement) return;
		const content = blockElement.innerText || '';

		// Check for slash command trigger at the start of empty block
		if (content === '/' || content.startsWith('/')) {
			// Make sure this block is tracked as focused before showing menu
			editor.setFocusedBlock(block.id);
			const rect = blockElement.getBoundingClientRect();
			editor.showSlashMenu({ x: rect.left, y: rect.bottom + 4 });
			editor.setSlashMenuQuery(content.slice(1));
		} else {
			editor.hideSlashMenu();
		}

		// Update store (for dirty tracking)
		editor.updateBlock(block.id, content);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!blockElement) return;

		// Slash menu navigation is handled by BlockMenu component via svelte:window
		// Just handle Escape to close menu from block level
		if ($editor.showSlashMenu && e.key === 'Escape') {
			e.preventDefault();
			editor.hideSlashMenu();
			return;
		}

		// Regular key handling
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();

			const currentContent = blockElement.innerText || '';

			// Save current content first
			editor.updateBlock(block.id, currentContent);

			// If content is empty and it's a list type, convert to paragraph
			if (currentContent === '' && ['bulletList', 'numberedList', 'todo'].includes(block.type)) {
				editor.changeBlockType(block.id, 'paragraph');
				return;
			}

			// Create new block with same type for lists
			const newType: BlockType = ['bulletList', 'numberedList', 'todo'].includes(block.type)
				? block.type
				: 'paragraph';
			const newBlockId = editor.addBlockAfter(block.id, newType);

			// Focus the new block after DOM updates
			tick().then(() => {
				const newBlockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				newBlockEl?.focus();
			});
		}

		if (e.key === 'Backspace') {
			const currentContent = blockElement.innerText || '';
			const selection = window.getSelection();

			// Only delete block if content is empty and cursor is at start
			if (currentContent === '' && index > 0) {
				e.preventDefault();
				editor.deleteBlock(block.id);
				// Focus previous block
				tick().then(() => {
					const prevBlock = $editor.blocks[Math.max(0, index - 1)];
					if (prevBlock) {
						const prevEl = document.querySelector(`[data-block-id="${prevBlock.id}"]`) as HTMLElement;
						prevEl?.focus();
						// Move cursor to end
						const range = document.createRange();
						const sel = window.getSelection();
						if (prevEl.childNodes.length > 0) {
							range.setStartAfter(prevEl.lastChild!);
						} else {
							range.setStart(prevEl, 0);
						}
						range.collapse(true);
						sel?.removeAllRanges();
						sel?.addRange(range);
					}
				});
			}
		}

		if (e.key === 'ArrowUp' && index > 0) {
			const selection = window.getSelection();
			if (selection && selection.anchorOffset === 0) {
				e.preventDefault();
				// Save current content
				editor.updateBlock(block.id, blockElement.innerText || '');
				editor.focusPreviousBlock();
				tick().then(() => {
					const prevBlock = $editor.blocks[index - 1];
					if (prevBlock) {
						const prevEl = document.querySelector(`[data-block-id="${prevBlock.id}"]`) as HTMLElement;
						prevEl?.focus();
					}
				});
			}
		}

		if (e.key === 'ArrowDown' && index < $editor.blocks.length - 1) {
			const selection = window.getSelection();
			const contentLength = (blockElement.innerText || '').length;
			if (selection && selection.anchorOffset === contentLength) {
				e.preventDefault();
				// Save current content
				editor.updateBlock(block.id, blockElement.innerText || '');
				editor.focusNextBlock();
				tick().then(() => {
					const nextBlock = $editor.blocks[index + 1];
					if (nextBlock) {
						const nextEl = document.querySelector(`[data-block-id="${nextBlock.id}"]`) as HTMLElement;
						nextEl?.focus();
					}
				});
			}
		}
	}


	async function selectBlockType(type: BlockType, properties?: Record<string, unknown>) {
		try {
			// Handle page type specially - create a new sub-page
			if (type === 'page') {
				await createSubPage(properties?.icon as string | undefined);
				return;
			}

			editor.changeBlockType(block.id, type);
			editor.updateBlock(block.id, ''); // Clear the slash command

			// Refocus and clear the block after DOM update
			await tick();
			if (blockElement) {
				blockElement.innerText = '';
				blockElement.focus();
			}
		} catch (e) {
			console.error('Error in selectBlockType:', e);
		}
	}

	async function createSubPage(icon?: string) {
		console.log('[Block] createSubPage called, parentContextId:', parentContextId, 'icon:', icon);

		// FIRST: Clear the slash command text immediately
		editor.updateBlock(block.id, '');
		editor.hideSlashMenu();
		await tick();
		if (blockElement) {
			blockElement.innerText = '';
		}

		if (!parentContextId) {
			console.log('[Block] No parentContextId - creating local page block only');
			// Still change the block type even without parent context
			editor.changeBlockType(block.id, 'page');
			editor.updateBlock(block.id, 'New page', { icon: icon || 'document' });
			await tick();
			if (blockElement) {
				blockElement.innerText = 'New page';
			}
			return;
		}

		try {
			// Create new context as a child of current document with selected icon
			const newContext = await contexts.createContext({
				name: 'New page',
				type: 'document',
				parent_id: parentContextId,
				blocks: [],
				icon: icon || 'document'
			});

			// Update current block to be a page reference with icon
			editor.updateBlock(block.id, 'New page', { pageId: newContext.id, icon: icon || 'document' });
			editor.changeBlockType(block.id, 'page');

			// Refresh contexts list so sidebar shows the new nested page
			await contexts.loadContexts();

			// Auto-expand the parent page in sidebar so the new child is visible
			// Import and use kbPreferences to expand the parent
			const { kbPreferences } = await import('$lib/stores/kb-preferences');
			kbPreferences.expandPage(parentContextId);

			// Clear the DOM element
			await tick();
			if (blockElement) {
				blockElement.innerText = 'New page';
			}

			console.log('[Block] Sub-page created:', newContext.id, 'under parent:', parentContextId, 'with icon:', icon);
		} catch (e) {
			console.error('Failed to create sub-page:', e);
		}
	}

	function handlePageClick(pageId: string) {
		if (onPageClick) {
			onPageClick(pageId);
		} else {
			goto(`/knowledge/${pageId}`);
		}
	}

	function handleTodoToggle(e: Event) {
		e.stopPropagation();
		if (block.type === 'todo') {
			editor.toggleTodo(block.id);
		}
	}

	// Get placeholder text based on block type
	function getPlaceholder(): string {
		if (block.type === 'heading1') return 'Heading 1';
		if (block.type === 'heading2') return 'Heading 2';
		if (block.type === 'heading3') return 'Heading 3';
		if (block.type === 'quote') return 'Quote';
		if (block.type === 'code') return 'Code';
		if (block.type === 'todo') return 'To-do';
		if (block.type === 'bulletList') return 'List item';
		if (block.type === 'numberedList') return 'List item';
		if (block.type === 'toggle') return 'Toggle header';
		if (block.type === 'page') return 'Page title';
		return "Type '/' for commands...";
	}

	// Check if block is empty for placeholder display
	let isEmpty = $derived(!block.content || block.content === '');

	// Language picker functions
	function selectLanguage(langId: string) {
		editor.updateBlock(block.id, block.content, { ...block.properties, language: langId });
		showLanguagePicker = false;
		languageSearchQuery = '';
	}

	function getLanguageLabel(langId: string | undefined): string {
		if (!langId) return 'Plain Text';
		const found = LANGUAGES.find(l => l.id === langId);
		return found?.label || langId;
	}

	function handleLanguagePickerKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			showLanguagePicker = false;
			languageSearchQuery = '';
		}
	}

	// Close language picker when clicking outside
	function handleLanguagePickerClickOutside(e: MouseEvent) {
		if (languagePickerRef && !languagePickerRef.contains(e.target as Node)) {
			showLanguagePicker = false;
			languageSearchQuery = '';
		}
	}

	$effect(() => {
		if (showLanguagePicker) {
			globalThis.document.addEventListener('click', handleLanguagePickerClickOutside);
			return () => globalThis.document.removeEventListener('click', handleLanguagePickerClickOutside);
		}
	});

	// Handle keydown on divider blocks (which aren't contenteditable)
	function handleDividerKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			const newBlockId = editor.addBlockAfter(block.id, 'paragraph');
			tick().then(() => {
				const newBlockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				newBlockEl?.focus();
			});
		}

		if (e.key === 'Backspace') {
			e.preventDefault();
			if (index > 0) {
				editor.deleteBlock(block.id);
				tick().then(() => {
					const prevBlock = $editor.blocks[Math.max(0, index - 1)];
					if (prevBlock) {
						const prevEl = document.querySelector(`[data-block-id="${prevBlock.id}"]`) as HTMLElement;
						prevEl?.focus();
					}
				});
			}
		}

		if (e.key === 'ArrowUp' && index > 0) {
			e.preventDefault();
			editor.focusPreviousBlock();
			tick().then(() => {
				const prevBlock = $editor.blocks[index - 1];
				if (prevBlock) {
					const prevEl = document.querySelector(`[data-block-id="${prevBlock.id}"]`) as HTMLElement;
					prevEl?.focus();
				}
			});
		}

		if (e.key === 'ArrowDown' && index < $editor.blocks.length - 1) {
			e.preventDefault();
			editor.focusNextBlock();
			tick().then(() => {
				const nextBlock = $editor.blocks[index + 1];
				if (nextBlock) {
					const nextEl = document.querySelector(`[data-block-id="${nextBlock.id}"]`) as HTMLElement;
					nextEl?.focus();
				}
			});
		}
	}
</script>

<div
	class="block-wrapper group relative py-0.5"
	class:dragging={isDragging}
	data-block-index={index}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	ondrop={handleDrop}
>
	<!-- Block handle (drag/menu) -->
	{#if !readonly}
		<div class="absolute -left-8 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5">
			<div
				draggable="true"
				ondragstart={handleDragStart}
				ondragend={handleDragEnd}
				class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-grab active:cursor-grabbing"
				title="Drag to move"
				role="button"
				tabindex="-1"
			>
				<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
					<circle cx="9" cy="6" r="1.5"/>
					<circle cx="15" cy="6" r="1.5"/>
					<circle cx="9" cy="12" r="1.5"/>
					<circle cx="15" cy="12" r="1.5"/>
					<circle cx="9" cy="18" r="1.5"/>
					<circle cx="15" cy="18" r="1.5"/>
				</svg>
			</div>
		</div>
	{/if}

	<!-- Block content based on type -->
	{#if block.type === 'divider'}
		<div
			tabindex={readonly ? -1 : 0}
			data-block-id={block.id}
			onfocus={handleFocus}
			onkeydown={handleDividerKeydown}
			class="py-2 outline-none group cursor-pointer"
		>
			<hr class="border-gray-300 dark:border-gray-600 group-focus:border-blue-400 transition-colors" />
		</div>
	{:else if block.type === 'page'}
		<!-- Page block - Notion-style inline link to sub-page (flat, minimal styling) -->
		<button
			onclick={() => block.properties?.pageId && handlePageClick(block.properties.pageId as string)}
			class="page-link group inline-flex items-center gap-1.5 py-0.5 text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors text-left"
		>
			<!-- Document icon -->
			<svg class="w-[18px] h-[18px] text-gray-400 dark:text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
			</svg>
			<span class="underline decoration-gray-300 dark:decoration-gray-600 underline-offset-2">{block.content || 'New page'}</span>
		</button>
	{:else if block.type === 'heading1'}
		<h1
			bind:this={blockElement}
			contenteditable={!readonly}
			data-block-id={block.id}
			data-placeholder={getPlaceholder()}
			onfocus={handleFocus}
			onblur={handleBlur}
			oninput={handleInput}
			onkeydown={handleKeydown}
			class="text-3xl font-bold text-gray-900 dark:text-gray-100 outline-none min-h-[1.2em] block-editable"
			class:is-empty={isEmpty}
		></h1>
	{:else if block.type === 'heading2'}
		<h2
			bind:this={blockElement}
			contenteditable={!readonly}
			data-block-id={block.id}
			data-placeholder={getPlaceholder()}
			onfocus={handleFocus}
			onblur={handleBlur}
			oninput={handleInput}
			onkeydown={handleKeydown}
			class="text-2xl font-semibold text-gray-900 dark:text-gray-100 outline-none min-h-[1.2em] block-editable"
			class:is-empty={isEmpty}
		></h2>
	{:else if block.type === 'heading3'}
		<h3
			bind:this={blockElement}
			contenteditable={!readonly}
			data-block-id={block.id}
			data-placeholder={getPlaceholder()}
			onfocus={handleFocus}
			onblur={handleBlur}
			oninput={handleInput}
			onkeydown={handleKeydown}
			class="text-xl font-semibold text-gray-800 dark:text-gray-200 outline-none min-h-[1.2em] block-editable"
			class:is-empty={isEmpty}
		></h3>
	{:else if block.type === 'bulletList'}
		<div class="flex items-start gap-2">
			<span class="mt-2 w-1.5 h-1.5 rounded-full bg-gray-400 dark:bg-gray-500 flex-shrink-0"></span>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 text-gray-700 dark:text-gray-300 outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'numberedList'}
		<div class="flex items-start gap-2">
			<span class="w-5 h-5 rounded-full bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400 text-xs flex items-center justify-center flex-shrink-0">
				{index + 1}
			</span>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 text-gray-700 dark:text-gray-300 outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'todo'}
		<div class="flex items-start gap-2">
			<button
				onclick={handleTodoToggle}
				class="w-4 h-4 mt-1 rounded border flex items-center justify-center transition-colors flex-shrink-0
					{block.properties?.checked
						? 'bg-blue-500 border-blue-500'
						: 'border-gray-500 hover:border-blue-400'}"
				tabindex="-1"
			>
				{#if block.properties?.checked}
					<svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				{/if}
			</button>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 outline-none min-h-[1.5em] block-editable
					{block.properties?.checked ? 'line-through text-gray-500' : 'text-gray-700 dark:text-gray-300'}"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'quote'}
		<blockquote class="border-l-4 border-gray-300 dark:border-gray-600 pl-4 py-1">
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="text-gray-600 dark:text-gray-400 italic outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</blockquote>
	{:else if block.type === 'code'}
		<div class="code-block rounded-md overflow-hidden border border-gray-200 dark:border-transparent">
			<!-- Language selector bar -->
			<div class="flex items-center justify-between px-3 py-1.5 bg-gray-100 dark:bg-[#2f2f2f] border-b border-gray-200 dark:border-[#3d3d3d]">
				<!-- Language picker dropdown -->
				<div class="relative" bind:this={languagePickerRef}>
					<button
						onclick={(e) => { e.stopPropagation(); showLanguagePicker = !showLanguagePicker; }}
						class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400 font-mono hover:text-gray-800 dark:hover:text-gray-200 transition-colors"
						tabindex="-1"
					>
						<span>{getLanguageLabel(block.properties?.language as string | undefined)}</span>
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					{#if showLanguagePicker}
						<div class="absolute left-0 top-full mt-1 w-48 max-h-64 bg-white dark:bg-[#252525] rounded-lg shadow-xl border border-gray-200 dark:border-[#3d3d3d] overflow-hidden z-50">
							<!-- Search input -->
							<div class="p-2 border-b border-gray-200 dark:border-[#3d3d3d]">
								<input
									type="text"
									bind:value={languageSearchQuery}
									onkeydown={handleLanguagePickerKeydown}
									placeholder="Search languages..."
									class="w-full px-2 py-1.5 text-xs bg-gray-50 dark:bg-[#1e1e1e] border border-gray-200 dark:border-[#3d3d3d] rounded text-gray-700 dark:text-gray-200 placeholder:text-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
								/>
							</div>
							<!-- Language list -->
							<div class="overflow-y-auto max-h-48">
								{#each filteredLanguages as lang}
									<button
										onclick={() => selectLanguage(lang.id)}
										class="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-[#2f2f2f] transition-colors flex items-center justify-between"
									>
										<span>{lang.label}</span>
										{#if block.properties?.language === lang.id}
											<svg class="w-3 h-3 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
										{/if}
									</button>
								{/each}
								{#if filteredLanguages.length === 0}
									<div class="px-3 py-4 text-xs text-gray-500 text-center">
										No languages found
									</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
				<button
					onclick={() => {
						if (block.content) {
							navigator.clipboard.writeText(block.content);
						}
					}}
					class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
					tabindex="-1"
				>
					Copy
				</button>
			</div>
			<pre
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder="// code..."
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="bg-gray-50 dark:bg-[#1e1e1e] text-gray-800 dark:text-[#d4d4d4] font-mono text-sm p-4 outline-none min-h-[2.5em] whitespace-pre-wrap block-editable"
				class:is-empty={isEmpty}
			></pre>
		</div>
	{:else if block.type === 'callout'}
		<div class="callout-block flex items-start gap-3 p-4 rounded-md bg-amber-50 dark:bg-[#2f2f2f] border border-amber-200 dark:border-transparent">
			<!-- Icon (clickable to change later) -->
			<button
				class="flex-shrink-0 w-6 h-6 flex items-center justify-center text-lg hover:bg-amber-100 dark:hover:bg-[#3d3d3d] rounded transition-colors"
				tabindex="-1"
				title="Click to change icon"
			>
				{block.properties?.calloutIcon || '💡'}
			</button>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder="Type something..."
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 text-gray-800 dark:text-gray-200 outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'toggle'}
		<div class="flex items-start gap-1">
			<button
				onclick={() => editor.toggleToggleBlock(block.id)}
				class="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-transform flex-shrink-0"
				class:rotate-90={block.properties?.expanded}
				tabindex="-1"
				aria-label="Toggle expand"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
				</svg>
			</button>
			<div class="flex-1">
				<div
					bind:this={blockElement}
					contenteditable={!readonly}
					data-block-id={block.id}
					data-placeholder="Toggle header..."
					onfocus={handleFocus}
					onblur={handleBlur}
					oninput={handleInput}
					onkeydown={handleKeydown}
					class="text-gray-900 dark:text-gray-100 font-medium outline-none min-h-[1.5em] block-editable"
					class:is-empty={isEmpty}
				></div>
				{#if block.properties?.expanded}
					<div class="pl-6 mt-1 space-y-0.5">
						{#if block.children?.length}
							{#each block.children as childBlock, childIdx}
								<Block block={childBlock} index={childIdx} {readonly} {parentContextId} {onPageClick} />
							{/each}
						{/if}
						<!-- Empty block placeholder to add content -->
						{#if !readonly}
							<div
								contenteditable="true"
								data-placeholder="Type inside toggle..."
								class="text-gray-700 dark:text-gray-300 outline-none min-h-[1.5em] empty:before:content-[attr(data-placeholder)] empty:before:text-gray-400 dark:empty:before:text-gray-500"
								onkeydown={(e) => {
									if (e.key === 'Enter' && !e.shiftKey) {
										e.preventDefault();
										// TODO: Add proper child block creation
										const target = e.currentTarget as HTMLElement;
										const content = target.innerText || '';
										if (content.trim()) {
											// For now, just add a paragraph child
											editor.update((s: EditorState) => ({
												...s,
												blocks: s.blocks.map((b: EditorBlock) =>
													b.id === block.id
														? {
															...b,
															children: [...(b.children || []), {
																id: Math.random().toString(36).substring(2, 11),
																type: 'paragraph' as BlockType,
																content: content.trim(),
																properties: {}
															}]
														}
														: b
												),
												isDirty: true
											}));
											target.innerText = '';
										}
									}
								}}
							></div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{:else if block.type === 'tableOfContents'}
		<!-- Table of Contents - auto-generated from headings -->
		<div class="toc-block p-4 rounded-lg bg-gray-50 dark:bg-[#1e1e1e] border border-gray-200 dark:border-[#3d3d3d]">
			<div class="flex items-center gap-2 mb-3 text-gray-600 dark:text-gray-400">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
				</svg>
				<span class="text-sm font-medium">Table of Contents</span>
			</div>
			<nav class="space-y-1">
				{#each $editor.blocks.filter(b => ['heading1', 'heading2', 'heading3'].includes(b.type)) as heading}
					<a
						href="#{heading.id}"
						onclick={(e) => {
							e.preventDefault();
							const el = document.querySelector(`[data-block-id="${heading.id}"]`);
							el?.scrollIntoView({ behavior: 'smooth', block: 'start' });
						}}
						class="block text-sm transition-colors {
							heading.type === 'heading1' ? 'text-gray-800 dark:text-gray-200 font-medium' :
							heading.type === 'heading2' ? 'pl-4 text-gray-700 dark:text-gray-300' :
							'pl-8 text-gray-600 dark:text-gray-400'
						} hover:text-blue-600 dark:hover:text-blue-400"
					>
						{heading.content || 'Untitled'}
					</a>
				{/each}
				{#if $editor.blocks.filter(b => ['heading1', 'heading2', 'heading3'].includes(b.type)).length === 0}
					<p class="text-sm text-gray-400 dark:text-gray-500 italic">No headings found. Add headings to see them here.</p>
				{/if}
			</nav>
		</div>
	{:else if block.type === 'columns'}
		<!-- Columns layout - 2 column default -->
		<div class="columns-block grid grid-cols-2 gap-4 p-2 rounded-lg border border-dashed border-gray-300 dark:border-gray-600">
			<div class="min-h-[100px] rounded bg-gray-50 dark:bg-[#1e1e1e] p-3 flex items-center justify-center">
				<span class="text-xs text-gray-400 dark:text-gray-500">Column 1 - Click to add blocks</span>
			</div>
			<div class="min-h-[100px] rounded bg-gray-50 dark:bg-[#1e1e1e] p-3 flex items-center justify-center">
				<span class="text-xs text-gray-400 dark:text-gray-500">Column 2 - Click to add blocks</span>
			</div>
		</div>
	{:else if block.type === 'bookmark'}
		<!-- Bookmark - link preview -->
		<div class="bookmark-block rounded-lg border border-gray-200 dark:border-[#3d3d3d] overflow-hidden hover:border-gray-300 dark:hover:border-gray-500 transition-colors">
			{#if block.properties?.url}
				<a
					href={block.properties.url as string}
					target="_blank"
					rel="noopener noreferrer"
					class="flex items-stretch"
				>
					<div class="flex-1 p-4">
						<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-1 line-clamp-1">
							{block.properties.title || block.properties.url}
						</h4>
						{#if block.properties.description}
							<p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mb-2">
								{block.properties.description}
							</p>
						{/if}
						<div class="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
							</svg>
							<span class="truncate">{new URL(block.properties.url as string).hostname}</span>
						</div>
					</div>
					{#if block.properties.image}
						<div class="w-32 bg-gray-100 dark:bg-[#2f2f2f]">
							<img src={block.properties.image as string} alt="" class="w-full h-full object-cover" />
						</div>
					{/if}
				</a>
			{:else}
				<!-- Empty bookmark - show input -->
				<div class="p-4">
					<div class="flex items-center gap-2 text-gray-500 dark:text-gray-400">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
						</svg>
						<input
							type="url"
							placeholder="Paste a link..."
							class="flex-1 bg-transparent border-none text-sm text-gray-700 dark:text-gray-200 placeholder:text-gray-400 focus:outline-none"
							onkeydown={(e) => {
								if (e.key === 'Enter') {
									const input = e.target as HTMLInputElement;
									if (input.value) {
										editor.updateBlock(block.id, block.content, { ...block.properties, url: input.value, title: input.value });
									}
								}
							}}
						/>
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Default paragraph -->
		<p
			bind:this={blockElement}
			contenteditable={!readonly}
			data-block-id={block.id}
			data-placeholder={getPlaceholder()}
			onfocus={handleFocus}
			onblur={handleBlur}
			oninput={handleInput}
			onkeydown={handleKeydown}
			class="text-gray-800 dark:text-gray-100 outline-none min-h-[1.5em] block-editable"
			class:is-empty={isEmpty}
		></p>
	{/if}

</div>

<style>
	.block-editable {
		position: relative;
	}

	.block-editable:focus {
		outline: none;
	}

	/* Only show placeholder on focus for paragraphs, or when empty for headings */
	.block-editable.is-empty:focus:before,
	h1.block-editable.is-empty:before,
	h2.block-editable.is-empty:before,
	h3.block-editable.is-empty:before {
		content: attr(data-placeholder);
		color: #6b7280;
		pointer-events: none;
		position: absolute;
		font-style: normal;
		font-weight: normal;
	}

	/* Dark mode text selection */
	.block-editable::selection {
		background-color: rgba(59, 130, 246, 0.4);
		color: #ffffff;
	}

	/* For webkit browsers (Chrome, Safari) */
	.block-editable::-webkit-selection {
		background-color: rgba(59, 130, 246, 0.4);
		color: #ffffff;
	}

	/* Ensure no weird background on focus */
	.block-editable:focus {
		background-color: transparent;
	}

	/* Drag and drop visual indicators */
	.block-wrapper {
		position: relative;
	}

	.block-wrapper.dragging {
		opacity: 0.4;
	}

	.block-wrapper.drag-over-above::before {
		content: '';
		position: absolute;
		top: -2px;
		left: 0;
		right: 0;
		height: 3px;
		background: linear-gradient(90deg, #3b82f6, #60a5fa);
		border-radius: 2px;
		z-index: 10;
		box-shadow: 0 0 6px rgba(59, 130, 246, 0.5);
	}

	.block-wrapper.drag-over-below::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 0;
		right: 0;
		height: 3px;
		background: linear-gradient(90deg, #3b82f6, #60a5fa);
		border-radius: 2px;
		z-index: 10;
		box-shadow: 0 0 6px rgba(59, 130, 246, 0.5);
	}

	/* Add small dot indicators at the edges for extra visibility */
	.block-wrapper.drag-over-above::before,
	.block-wrapper.drag-over-below::after {
		animation: pulse-indicator 0.8s ease-in-out infinite alternate;
	}

	@keyframes pulse-indicator {
		from {
			opacity: 0.7;
		}
		to {
			opacity: 1;
		}
	}
</style>
