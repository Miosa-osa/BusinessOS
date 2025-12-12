<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { editor, type EditorBlock, type BlockType, blockTypes } from '$lib/stores/editor';
	import { contexts } from '$lib/stores/contexts';

	interface Props {
		block: EditorBlock;
		index: number;
		readonly?: boolean;
		parentContextId?: string;
	}

	let { block, index, readonly = false, parentContextId }: Props = $props();

	let blockElement: HTMLElement | null = $state(null);
	let showSlashMenu = $state(false);
	let slashFilter = $state('');
	let slashSelectedIndex = $state(0);

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

	function handleFocus() {
		editor.setFocusedBlock(block.id);
	}

	function handleBlur(e: FocusEvent) {
		// Don't blur if clicking on slash menu
		const relatedTarget = e.relatedTarget as HTMLElement;
		if (relatedTarget?.closest('[data-slash-menu]')) {
			return;
		}
		showSlashMenu = false;
		slashFilter = '';

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
		if (content === '/') {
			showSlashMenu = true;
			slashFilter = '';
			slashSelectedIndex = 0;
		} else if (content.startsWith('/') && content.length > 1) {
			showSlashMenu = true;
			slashFilter = content.slice(1);
			slashSelectedIndex = 0;
		} else {
			showSlashMenu = false;
			slashFilter = '';
		}

		// Update store (for dirty tracking)
		editor.updateBlock(block.id, content);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!blockElement) return;

		// Handle slash menu navigation
		if (showSlashMenu) {
			const filteredTypes = getFilteredBlockTypes();
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				slashSelectedIndex = Math.min(slashSelectedIndex + 1, filteredTypes.length - 1);
				return;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				slashSelectedIndex = Math.max(slashSelectedIndex - 1, 0);
				return;
			}
			if (e.key === 'Enter' || e.key === 'Tab') {
				e.preventDefault();
				if (filteredTypes[slashSelectedIndex]) {
					selectBlockType(filteredTypes[slashSelectedIndex].type);
				}
				return;
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				showSlashMenu = false;
				slashFilter = '';
				return;
			}
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

	function getFilteredBlockTypes() {
		if (!slashFilter) return blockTypes;
		return blockTypes.filter(bt =>
			bt.label.toLowerCase().includes(slashFilter.toLowerCase()) ||
			bt.type.toLowerCase().includes(slashFilter.toLowerCase())
		);
	}

	async function selectBlockType(type: BlockType) {
		// Handle page type specially - create a new sub-page
		if (type === 'page') {
			await createSubPage();
			return;
		}

		editor.changeBlockType(block.id, type);
		editor.updateBlock(block.id, ''); // Clear the slash command
		showSlashMenu = false;
		slashFilter = '';

		// Refocus and clear the block after DOM update
		await tick();
		if (blockElement) {
			blockElement.innerText = '';
			blockElement.focus();
		}
	}

	async function createSubPage() {
		if (!parentContextId) return;

		try {
			// Create new context as a child of current document
			const newContext = await contexts.createContext({
				name: 'Untitled',
				type: 'document',
				parent_id: parentContextId,
				blocks: []
			});

			// Update current block to be a page reference
			editor.updateBlock(block.id, 'Untitled', { pageId: newContext.id });
			editor.changeBlockType(block.id, 'page');
			showSlashMenu = false;
			slashFilter = '';

			// Clear the DOM element
			await tick();
			if (blockElement) {
				blockElement.innerText = 'Untitled';
			}
		} catch (e) {
			console.error('Failed to create sub-page:', e);
		}
	}

	function handlePageClick(pageId: string) {
		goto(`/contexts/${pageId}`);
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
		if (block.type === 'page') return 'Page title';
		return "Type '/' for commands...";
	}

	// Check if block is empty for placeholder display
	let isEmpty = $derived(!block.content || block.content === '');

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

<div class="block-wrapper group relative py-0.5" data-block-index={index}>
	<!-- Block handle (drag/menu) -->
	{#if !readonly}
		<div class="absolute -left-6 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity flex items-center">
			<button
				class="p-0.5 rounded hover:bg-gray-100 text-gray-400 cursor-grab"
				title="Drag to move"
				tabindex="-1"
			>
				<svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
					<circle cx="9" cy="6" r="1.5"/>
					<circle cx="15" cy="6" r="1.5"/>
					<circle cx="9" cy="12" r="1.5"/>
					<circle cx="15" cy="12" r="1.5"/>
					<circle cx="9" cy="18" r="1.5"/>
					<circle cx="15" cy="18" r="1.5"/>
				</svg>
			</button>
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
			<hr class="border-gray-200 group-focus:border-blue-400 transition-colors" />
		</div>
	{:else if block.type === 'page'}
		<!-- Page block - compact link to sub-page -->
		<a
			href={block.properties?.pageId ? `/contexts/${block.properties.pageId}` : '#'}
			class="inline-flex items-center gap-1.5 px-2 py-1 rounded-md bg-gray-100 hover:bg-gray-200 transition-colors text-sm"
		>
			<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
			</svg>
			<span class="text-gray-700">{block.content || 'Untitled'}</span>
		</a>
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
			class="text-3xl font-bold text-gray-900 outline-none min-h-[1.2em] block-editable"
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
			class="text-2xl font-semibold text-gray-800 outline-none min-h-[1.2em] block-editable"
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
			class="text-xl font-semibold text-gray-700 outline-none min-h-[1.2em] block-editable"
			class:is-empty={isEmpty}
		></h3>
	{:else if block.type === 'bulletList'}
		<div class="flex items-start gap-2">
			<span class="mt-2 w-1.5 h-1.5 rounded-full bg-gray-400 flex-shrink-0"></span>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 text-gray-600 outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'numberedList'}
		<div class="flex items-start gap-2">
			<span class="w-5 h-5 rounded-full bg-blue-50 text-blue-600 text-xs flex items-center justify-center flex-shrink-0">
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
				class="flex-1 text-gray-600 outline-none min-h-[1.5em] block-editable"
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
						: 'border-gray-300 hover:border-blue-400'}"
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
					{block.properties?.checked ? 'line-through text-gray-400' : 'text-gray-600'}"
				class:is-empty={isEmpty}
			></div>
		</div>
	{:else if block.type === 'quote'}
		<blockquote class="border-l-4 border-gray-300 pl-4 py-1">
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="text-gray-600 italic outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
		</blockquote>
	{:else if block.type === 'code'}
		<div class="bg-gray-100 rounded-lg p-3 font-mono text-sm">
			<pre
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder={getPlaceholder()}
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="text-gray-800 outline-none min-h-[1.5em] whitespace-pre-wrap block-editable"
				class:is-empty={isEmpty}
			></pre>
		</div>
	{:else if block.type === 'callout'}
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-3 flex items-start gap-2">
			<span class="text-blue-500">💡</span>
			<div
				bind:this={blockElement}
				contenteditable={!readonly}
				data-block-id={block.id}
				data-placeholder="Callout text..."
				onfocus={handleFocus}
				onblur={handleBlur}
				oninput={handleInput}
				onkeydown={handleKeydown}
				class="flex-1 text-blue-800 outline-none min-h-[1.5em] block-editable"
				class:is-empty={isEmpty}
			></div>
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
			class="text-gray-600 outline-none min-h-[1.5em] block-editable"
			class:is-empty={isEmpty}
		></p>
	{/if}

	<!-- Slash Command Menu -->
	{#if showSlashMenu && !readonly}
		<div
			data-slash-menu
			class="absolute left-0 top-full mt-1 bg-white rounded-lg shadow-xl border border-gray-200 z-50 overflow-hidden w-64"
		>
			<div class="py-1 max-h-64 overflow-auto">
				{#each getFilteredBlockTypes() as blockType, idx}
					<button
						onclick={() => selectBlockType(blockType.type)}
						class="w-full px-3 py-2 flex items-center gap-3 text-left transition-colors
							{idx === slashSelectedIndex ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-700'}"
					>
						<span class="w-8 h-8 rounded bg-gray-100 flex items-center justify-center text-sm">
							{blockType.icon}
						</span>
						<div class="flex-1 min-w-0">
							<div class="text-sm font-medium">{blockType.label}</div>
							<div class="text-xs text-gray-400">{blockType.description}</div>
						</div>
						{#if idx === slashSelectedIndex}
							<span class="text-xs text-gray-400">Enter</span>
						{/if}
					</button>
				{/each}
				{#if getFilteredBlockTypes().length === 0}
					<div class="px-3 py-4 text-sm text-gray-400 text-center">
						No matching blocks
					</div>
				{/if}
			</div>
		</div>
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
		color: #d1d5db;
		pointer-events: none;
		position: absolute;
		font-style: normal;
		font-weight: normal;
	}
</style>
