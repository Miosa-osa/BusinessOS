<script lang="ts">
	import type { Block, BlockType, RichText } from '../../entities/types';
	import { createBlock } from '../../services/documents.service';
	import SlashMenu from './SlashMenu.svelte';
	import FormatToolbar from './FormatToolbar.svelte';
	import BlockControls from './BlockControls.svelte';
	import BlockContentRenderer from './BlockContentRenderer.svelte';
	import { parseHTMLToRichText } from './blockUtils';

	interface Props {
		block: Block;
		index: number;
		readOnly?: boolean;
		shouldFocus?: boolean;
		isFirstBlock?: boolean;
		isOnlyEmptyBlock?: boolean;
		onBlockChange?: (block: Block) => void;
		onBlockDelete?: () => void;
		onBlockAdd?: (block: Block) => void;
		onFocused?: () => void;
		onBlockMove?: (fromIndex: number, toIndex: number) => void;
		totalBlocks?: number;
	}

	let {
		block,
		index,
		readOnly = false,
		shouldFocus = false,
		isFirstBlock = false,
		isOnlyEmptyBlock = false,
		onBlockChange,
		onBlockDelete,
		onBlockAdd,
		onFocused,
		onBlockMove,
		totalBlocks = 0
	}: Props = $props();

	// Determine placeholder text
	const getPlaceholder = (type: string): string => {
		if (isOnlyEmptyBlock && type === 'paragraph') {
			return "Press '/' for commands...";
		}
		switch (type) {
			case 'heading_1': return 'Heading 1';
			case 'heading_2': return 'Heading 2';
			case 'heading_3': return 'Heading 3';
			case 'bulleted_list': return 'List item';
			case 'numbered_list': return 'List item';
			case 'to_do': return 'To-do';
			case 'quote': return 'Quote';
			case 'code': return 'Code';
			case 'callout': return 'Callout';
			default: return '';
		}
	};

	let isHovered = $state(false);

	// Slash menu state
	let showSlashMenu = $state(false);
	let slashMenuPosition = $state({ x: 0, y: 0 });
	let slashFilter = $state('');
	let slashMenuRef: { handleKeydown: (e: KeyboardEvent) => void } | undefined;
	let blockElementRef: HTMLElement | null = $state(null);

	// Format toolbar state
	let showFormatToolbar = $state(false);
	let formatToolbarPosition = $state({ x: 0, y: 0 });
	let currentSelection: { start: number; end: number; text: string } | null = $state(null);

	// Divider style picker state
	let showDividerPicker = $state(false);

	// Drag and drop state
	let isDragging = $state(false);
	let isDragOver = $state(false);
	let dragOverPosition: 'before' | 'after' | null = $state(null);

	// Track block ID to detect external changes (document switch)
	let lastBlockId = $state('');
	let lastContent = $state('');

	// Initialize tracking state from block prop
	$effect(() => {
		if (lastBlockId === '') {
			lastBlockId = block.id;
			lastContent = getPlainText();
		}
	});

	function getPlainText(): string {
		return block.content.map((rt) => rt.plain_text).join('');
	}

	// Action for contenteditable elements - handles initial content and external updates
	function contenteditable(node: HTMLElement) {
		blockElementRef = node;
		node.textContent = getPlainText();

		return {
			update() {
				const currentContent = getPlainText();
				if (block.id !== lastBlockId || (currentContent !== lastContent && document.activeElement !== node)) {
					node.textContent = currentContent;
					lastBlockId = block.id;
					lastContent = currentContent;
				}
			},
			destroy() {
				if (blockElementRef === node) {
					blockElementRef = null;
				}
			}
		};
	}

	$effect(() => {
		const content = getPlainText();
		if (block.id !== lastBlockId) {
			lastBlockId = block.id;
			lastContent = content;
		}
	});

	$effect(() => {
		if (shouldFocus && blockElementRef) {
			blockElementRef.focus();
			const selection = window.getSelection();
			if (selection) {
				const range = document.createRange();
				range.selectNodeContents(blockElementRef);
				range.collapse(false);
				selection.removeAllRanges();
				selection.addRange(range);
			}
			onFocused?.();
		}
	});

	function handleContentInput(e: Event) {
		const target = e.target as HTMLElement;
		const newContent = target.textContent || '';

		lastContent = newContent;

		const slashMatch = newContent.match(/^\/(\S*)$/);
		if (slashMatch) {
			const selection = window.getSelection();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				const rect = range.getBoundingClientRect();
				slashMenuPosition = { x: rect.left, y: rect.bottom + 4 };
			} else if (target) {
				const rect = target.getBoundingClientRect();
				slashMenuPosition = { x: rect.left, y: rect.bottom + 4 };
			}
			slashFilter = slashMatch[1] || '';
			showSlashMenu = true;
		} else if (showSlashMenu) {
			const lastSlashIndex = newContent.lastIndexOf('/');
			if (lastSlashIndex >= 0) {
				slashFilter = newContent.slice(lastSlashIndex + 1);
			} else {
				showSlashMenu = false;
				slashFilter = '';
			}
		}

		const richText: RichText[] = [
			{
				type: 'text',
				text: { content: newContent, link: null },
				annotations: {
					bold: false,
					italic: false,
					strikethrough: false,
					underline: false,
					code: false,
					color: 'default'
				},
				plain_text: newContent,
				href: null
			}
		];

		onBlockChange?.({ ...block, content: richText });
	}

	function handleKeydown(e: KeyboardEvent) {
		if (showSlashMenu) {
			if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Escape') {
				e.preventDefault();
				slashMenuRef?.handleKeydown(e);
				return;
			}
			if (e.key === 'Enter') {
				e.preventDefault();
				slashMenuRef?.handleKeydown(e);
				return;
			}
		}

		const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
		const modifier = isMac ? e.metaKey : e.ctrlKey;

		if (modifier && !e.shiftKey) {
			switch (e.key.toLowerCase()) {
				case 'b':
					e.preventDefault();
					document.execCommand('bold', false);
					updateBlockFromDOM();
					return;
				case 'i':
					e.preventDefault();
					document.execCommand('italic', false);
					updateBlockFromDOM();
					return;
				case 'u':
					e.preventDefault();
					document.execCommand('underline', false);
					updateBlockFromDOM();
					return;
				case 'e':
					e.preventDefault();
					const selectionE = window.getSelection();
					if (selectionE && selectionE.rangeCount > 0 && !selectionE.isCollapsed) {
						const range = selectionE.getRangeAt(0);
						const codeElement = document.createElement('code');
						codeElement.className = 'inline-code';
						range.surroundContents(codeElement);
						updateBlockFromDOM();
					}
					return;
				case 'k':
					e.preventDefault();
					const url = prompt('Enter URL:');
					if (url) {
						document.execCommand('createLink', false, url);
						updateBlockFromDOM();
					}
					return;
			}
		}

		if (modifier && e.shiftKey && e.key.toLowerCase() === 's') {
			e.preventDefault();
			document.execCommand('strikeThrough', false);
			updateBlockFromDOM();
			return;
		}

		if (e.key === 'Escape' && showFormatToolbar) {
			showFormatToolbar = false;
			currentSelection = null;
			return;
		}

		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			onBlockAdd?.(createBlock('paragraph'));
		}
		if (e.key === 'Backspace' && getPlainText() === '') {
			e.preventDefault();
			onBlockDelete?.();
		}
	}

	function handleSlashMenuSelect(newBlock: Block) {
		const target = blockElementRef;
		if (target) {
			target.textContent = '';
		}
		const emptyRichText: RichText[] = [];
		onBlockChange?.({ ...block, content: emptyRichText });
		onBlockAdd?.(newBlock);
		showSlashMenu = false;
		slashFilter = '';
	}

	function handleSlashMenuClose() {
		showSlashMenu = false;
		slashFilter = '';
	}

	function showDividerStylePicker() {
		showDividerPicker = !showDividerPicker;
	}

	function closeDividerPicker() {
		showDividerPicker = false;
	}

	$effect(() => {
		if (showDividerPicker) {
			const handleClickOutside = (e: MouseEvent) => {
				const target = e.target as HTMLElement;
				if (!target.closest('.divider-wrapper')) {
					closeDividerPicker();
				}
			};
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});

	function selectDividerStyle(style: string) {
		const updatedBlock: Block = {
			...block,
			properties: {
				...block.properties,
				divider_style: style as 'solid' | 'dashed' | 'dotted' | 'thick' | 'double' | 'gradient'
			}
		};
		onBlockChange?.(updatedBlock);
		showDividerPicker = false;
	}

	function handleSelectionChange() {
		if (readOnly) return;

		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || selection.rangeCount === 0) {
			showFormatToolbar = false;
			currentSelection = null;
			return;
		}

		const range = selection.getRangeAt(0);
		if (!blockElementRef || !blockElementRef.contains(range.commonAncestorContainer)) {
			showFormatToolbar = false;
			currentSelection = null;
			return;
		}

		const selectedText = selection.toString();
		if (selectedText.length === 0) {
			showFormatToolbar = false;
			currentSelection = null;
			return;
		}

		const rect = range.getBoundingClientRect();
		formatToolbarPosition = {
			x: rect.left + rect.width / 2,
			y: rect.top
		};

		const fullText = blockElementRef.textContent || '';
		const startOffset = fullText.indexOf(selectedText);

		currentSelection = {
			start: startOffset >= 0 ? startOffset : 0,
			end: startOffset >= 0 ? startOffset + selectedText.length : selectedText.length,
			text: selectedText
		};

		showFormatToolbar = true;
	}

	function handleFormat(format: 'bold' | 'italic' | 'underline' | 'strikethrough' | 'code' | 'link') {
		if (!currentSelection || !blockElementRef) return;

		switch (format) {
			case 'bold':
				document.execCommand('bold', false);
				break;
			case 'italic':
				document.execCommand('italic', false);
				break;
			case 'underline':
				document.execCommand('underline', false);
				break;
			case 'strikethrough':
				document.execCommand('strikeThrough', false);
				break;
			case 'code':
				const selectionC = window.getSelection();
				if (selectionC && selectionC.rangeCount > 0) {
					const range = selectionC.getRangeAt(0);
					const codeElement = document.createElement('code');
					codeElement.className = 'inline-code';
					range.surroundContents(codeElement);
				}
				break;
			case 'link':
				const url = prompt('Enter URL:');
				if (url) {
					document.execCommand('createLink', false, url);
				}
				break;
		}

		updateBlockFromDOM();
		showFormatToolbar = false;
		currentSelection = null;
	}

	function updateBlockFromDOM() {
		if (!blockElementRef) return;

		const html = blockElementRef.innerHTML;
		const text = blockElementRef.textContent || '';
		const richText = parseHTMLToRichText(html, text);

		lastContent = text;
		onBlockChange?.({ ...block, content: richText });
	}

	function handleMouseUp() {
		setTimeout(handleSelectionChange, 10);
	}

	function handleDragStart(e: DragEvent) {
		if (readOnly) return;

		isDragging = true;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', index.toString());
			e.dataTransfer.setData('application/x-block-id', block.id);
		}
	}

	function handleDragEnd() {
		isDragging = false;
		isDragOver = false;
		dragOverPosition = null;
	}

	function handleDragOver(e: DragEvent) {
		if (readOnly) return;

		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}

		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const midpoint = rect.top + rect.height / 2;
		const position = e.clientY < midpoint ? 'before' : 'after';

		isDragOver = true;
		dragOverPosition = position;
	}

	function handleDragLeave() {
		isDragOver = false;
		dragOverPosition = null;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragOver = false;
		dragOverPosition = null;

		if (readOnly || !e.dataTransfer) return;

		const fromIndexStr = e.dataTransfer.getData('text/plain');
		const fromIndex = parseInt(fromIndexStr, 10);

		if (isNaN(fromIndex)) return;

		let toIndex = index;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const midpoint = rect.top + rect.height / 2;

		if (e.clientY > midpoint) {
			toIndex = index + 1;
		}

		if (fromIndex === toIndex || fromIndex + 1 === toIndex) return;

		if (fromIndex < toIndex) {
			toIndex -= 1;
		}

		onBlockMove?.(fromIndex, toIndex);
	}

	function handleCheckboxChange(e: Event) {
		const target = e.target as HTMLInputElement;
		onBlockChange?.({
			...block,
			properties: { ...block.properties, checked: target.checked }
		});
	}

	function handleToggleExpand() {
		onBlockChange?.({
			...block,
			properties: { ...block.properties, expanded: !block.properties.expanded }
		});
	}

	function addBlockOfType(type: BlockType) {
		onBlockAdd?.(createBlock(type));
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="block-wrapper"
	class:block-wrapper--hovered={isHovered}
	class:block-wrapper--dragging={isDragging}
	class:block-wrapper--drag-over={isDragOver}
	class:block-wrapper--drag-before={dragOverPosition === 'before'}
	class:block-wrapper--drag-after={dragOverPosition === 'after'}
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	ondrop={handleDrop}
>
	{#if !readOnly}
		<BlockControls
			onAddBlock={addBlockOfType}
			onDragStart={handleDragStart}
			onDragEnd={handleDragEnd}
		/>
	{/if}

	<BlockContentRenderer
		{block}
		{index}
		{readOnly}
		contenteditableAction={contenteditable}
		onContentInput={handleContentInput}
		onKeydown={handleKeydown}
		onMouseUp={handleMouseUp}
		onCheckboxChange={handleCheckboxChange}
		onToggleExpand={handleToggleExpand}
		onDividerClick={showDividerStylePicker}
		onSelectDividerStyle={selectDividerStyle}
		onCloseDividerPicker={closeDividerPicker}
		onBlockChange={(updated) => onBlockChange?.(updated)}
		{showDividerPicker}
		{getPlaceholder}
	/>

	<!-- Slash command menu -->
	<SlashMenu
		visible={showSlashMenu}
		position={slashMenuPosition}
		filter={slashFilter}
		onSelect={handleSlashMenuSelect}
		onClose={handleSlashMenuClose}
		bind:this={slashMenuRef}
	/>

	<!-- Format toolbar for text selection -->
	<FormatToolbar
		visible={showFormatToolbar}
		position={formatToolbarPosition}
		onFormat={handleFormat}
	/>
</div>

<style>
	.block-wrapper {
		position: relative;
		display: flex;
		align-items: flex-start;
		gap: 0.25rem;
		padding: 2px 0;
		padding-left: 52px;
		margin-left: -52px;
	}

	.block-wrapper--dragging {
		opacity: 0.5;
	}

	.block-wrapper--drag-over {
		position: relative;
	}

	.block-wrapper--drag-before::before {
		content: '';
		position: absolute;
		top: 0;
		left: 52px;
		right: 0;
		height: 2px;
		background-color: #1e96eb;
		border-radius: 1px;
	}

	.block-wrapper--drag-after::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 52px;
		right: 0;
		height: 2px;
		background-color: #1e96eb;
		border-radius: 1px;
	}
</style>
