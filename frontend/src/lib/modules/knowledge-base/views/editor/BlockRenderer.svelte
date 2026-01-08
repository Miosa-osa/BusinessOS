<script lang="ts">
	import type { Block, BlockType, RichText } from '../../entities/types';
	import { createBlock } from '../../services/documents.service';
	import { GripVertical, Plus, Trash2, ChevronRight, ChevronDown, ImageIcon, Link2, ExternalLink, Lightbulb, AlertCircle, Info, AlertTriangle, CheckCircle } from 'lucide-svelte';
	import { Menu, MenuItem, MenuSeparator, MenuLabel } from '$lib/ui';
	import SlashMenu from './SlashMenu.svelte';
	import FormatToolbar from './FormatToolbar.svelte';
	import TableBlock from './TableBlock.svelte';

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
	let showBlockMenu = $state(false);
	let showAddMenu = $state(false);

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
	const dividerStyles = [
		{ value: 'solid', label: 'Solid' },
		{ value: 'dashed', label: 'Dashed' },
		{ value: 'dotted', label: 'Dotted' },
		{ value: 'thick', label: 'Thick' },
		{ value: 'double', label: 'Double' },
		{ value: 'gradient', label: 'Gradient' }
	] as const;

	// Drag and drop state
	let isDragging = $state(false);
	let isDragOver = $state(false);
	let dragOverPosition: 'before' | 'after' | null = $state(null);

	// Track block ID to detect external changes (document switch)
	let lastBlockId = $state(block.id);
	let lastContent = $state(getPlainText());

	// Block type options for the add menu
	const blockTypes: { type: BlockType; label: string; icon: string }[] = [
		{ type: 'paragraph', label: 'Text', icon: 'T' },
		{ type: 'heading_1', label: 'Heading 1', icon: 'H1' },
		{ type: 'heading_2', label: 'Heading 2', icon: 'H2' },
		{ type: 'heading_3', label: 'Heading 3', icon: 'H3' },
		{ type: 'bulleted_list', label: 'Bulleted List', icon: '•' },
		{ type: 'numbered_list', label: 'Numbered List', icon: '1.' },
		{ type: 'to_do', label: 'To-do', icon: '[]' },
		{ type: 'toggle', label: 'Toggle', icon: '>' },
		{ type: 'quote', label: 'Quote', icon: '"' },
		{ type: 'divider', label: 'Divider', icon: '—' },
		{ type: 'code', label: 'Code', icon: '</>' },
		{ type: 'callout', label: 'Callout', icon: '!' },
		{ type: 'table', label: 'Table', icon: '#' }
	];

	function getPlainText(): string {
		return block.content.map((rt) => rt.plain_text).join('');
	}

	// Action for contenteditable elements - handles initial content and external updates
	function contenteditable(node: HTMLElement) {
		// Store reference for focus management
		blockElementRef = node;

		// Set initial content
		node.textContent = getPlainText();

		// Return object for Svelte action interface
		return {
			update() {
				// Only update DOM if block ID changed (document switch) or external content change
				const currentContent = getPlainText();
				if (block.id !== lastBlockId || (currentContent !== lastContent && document.activeElement !== node)) {
					node.textContent = currentContent;
					lastBlockId = block.id;
					lastContent = currentContent;
				}
			},
			destroy() {
				// Clean up reference when element is destroyed
				if (blockElementRef === node) {
					blockElementRef = null;
				}
			}
		};
	}

	// Update tracking when block changes externally
	$effect(() => {
		const content = getPlainText();
		if (block.id !== lastBlockId) {
			lastBlockId = block.id;
			lastContent = content;
		}
	});

	// Focus management - focus this block when shouldFocus is true
	$effect(() => {
		if (shouldFocus && blockElementRef) {
			blockElementRef.focus();
			// Move cursor to end of content
			const selection = window.getSelection();
			if (selection) {
				const range = document.createRange();
				range.selectNodeContents(blockElementRef);
				range.collapse(false); // false = collapse to end
				selection.removeAllRanges();
				selection.addRange(range);
			}
			onFocused?.();
		}
	});

	function handleContentInput(e: Event) {
		const target = e.target as HTMLElement;
		const newContent = target.textContent || '';

		// Track this as local input
		lastContent = newContent;

		// Check for slash command
		const slashMatch = newContent.match(/^\/(\S*)$/);
		if (slashMatch) {
			// Show slash menu
			const selection = window.getSelection();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				const rect = range.getBoundingClientRect();
				slashMenuPosition = {
					x: rect.left,
					y: rect.bottom + 4
				};
			} else if (target) {
				const rect = target.getBoundingClientRect();
				slashMenuPosition = {
					x: rect.left,
					y: rect.bottom + 4
				};
			}
			slashFilter = slashMatch[1] || '';
			showSlashMenu = true;
		} else if (showSlashMenu) {
			// Check if still typing after slash
			const lastSlashIndex = newContent.lastIndexOf('/');
			if (lastSlashIndex >= 0) {
				slashFilter = newContent.slice(lastSlashIndex + 1);
			} else {
				showSlashMenu = false;
				slashFilter = '';
			}
		}

		// Convert to rich text
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
		// Forward key events to slash menu if open
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

		// Formatting keyboard shortcuts
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
					// Toggle inline code
					const selection = window.getSelection();
					if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
						const range = selection.getRangeAt(0);
						const codeElement = document.createElement('code');
						codeElement.className = 'inline-code';
						range.surroundContents(codeElement);
						updateBlockFromDOM();
					}
					return;
				case 'k':
					e.preventDefault();
					// Add link
					const url = prompt('Enter URL:');
					if (url) {
						document.execCommand('createLink', false, url);
						updateBlockFromDOM();
					}
					return;
			}
		}

		// Shift+Cmd+S for strikethrough
		if (modifier && e.shiftKey && e.key.toLowerCase() === 's') {
			e.preventDefault();
			document.execCommand('strikeThrough', false);
			updateBlockFromDOM();
			return;
		}

		// Close format toolbar on Escape
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
		// Clear the slash command from current block
		const target = blockElementRef;
		if (target) {
			target.textContent = '';
		}

		// Clear block content and add the new block
		const emptyRichText: RichText[] = [];
		onBlockChange?.({ ...block, content: emptyRichText });

		// Add the new block
		onBlockAdd?.(newBlock);

		showSlashMenu = false;
		slashFilter = '';
	}

	function handleSlashMenuClose() {
		showSlashMenu = false;
		slashFilter = '';
	}

	// Divider style picker functions
	function showDividerStylePicker() {
		showDividerPicker = !showDividerPicker;
	}

	function closeDividerPicker() {
		showDividerPicker = false;
	}

	// Close picker when clicking outside
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

	// Handle text selection for format toolbar
	function handleSelectionChange() {
		if (readOnly) return;

		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || selection.rangeCount === 0) {
			showFormatToolbar = false;
			currentSelection = null;
			return;
		}

		// Check if selection is within this block
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

		// Get selection bounds for toolbar position
		const rect = range.getBoundingClientRect();
		formatToolbarPosition = {
			x: rect.left + rect.width / 2,
			y: rect.top
		};

		// Track selection offsets within the block
		const fullText = blockElementRef.textContent || '';
		const startOffset = fullText.indexOf(selectedText);

		currentSelection = {
			start: startOffset >= 0 ? startOffset : 0,
			end: startOffset >= 0 ? startOffset + selectedText.length : selectedText.length,
			text: selectedText
		};

		showFormatToolbar = true;
	}

	// Handle format toolbar actions
	function handleFormat(format: 'bold' | 'italic' | 'underline' | 'strikethrough' | 'code' | 'link') {
		if (!currentSelection || !blockElementRef) return;

		const { start, end, text } = currentSelection;
		const fullText = blockElementRef.textContent || '';

		// For now, use document.execCommand for basic formatting
		// This maintains browser's native formatting behavior
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
				// Wrap in code tag manually
				const selection = window.getSelection();
				if (selection && selection.rangeCount > 0) {
					const range = selection.getRangeAt(0);
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

		// Update block content to reflect formatting
		updateBlockFromDOM();

		// Hide toolbar after formatting
		showFormatToolbar = false;
		currentSelection = null;
	}

	// Update block content from DOM (to capture formatting)
	function updateBlockFromDOM() {
		if (!blockElementRef) return;

		const html = blockElementRef.innerHTML;
		const text = blockElementRef.textContent || '';

		// Parse HTML to RichText segments
		const richText = parseHTMLToRichText(html, text);

		lastContent = text;
		onBlockChange?.({ ...block, content: richText });
	}

	// Parse HTML content to RichText array
	function parseHTMLToRichText(html: string, plainText: string): RichText[] {
		// Simple parsing - for now just capture the plain text
		// TODO: Implement full HTML to RichText parsing for proper annotation support
		const segments: RichText[] = [];

		// If no HTML formatting, return simple text
		if (html === plainText || !html.includes('<')) {
			if (plainText) {
				segments.push({
					type: 'text',
					text: { content: plainText, link: null },
					annotations: {
						bold: false,
						italic: false,
						strikethrough: false,
						underline: false,
						code: false,
						color: 'default'
					},
					plain_text: plainText,
					href: null
				});
			}
			return segments;
		}

		// Parse formatted HTML - simplified approach
		// Create a temporary element to parse the HTML
		const temp = document.createElement('div');
		temp.innerHTML = html;

		function processNode(node: Node, annotations: RichText['annotations']): void {
			if (node.nodeType === Node.TEXT_NODE) {
				const text = node.textContent || '';
				if (text) {
					segments.push({
						type: 'text',
						text: { content: text, link: null },
						annotations: { ...annotations },
						plain_text: text,
						href: null
					});
				}
			} else if (node.nodeType === Node.ELEMENT_NODE) {
				const element = node as Element;
				const newAnnotations = { ...annotations };

				// Check for formatting tags
				const tagName = element.tagName.toLowerCase();
				if (tagName === 'b' || tagName === 'strong') newAnnotations.bold = true;
				if (tagName === 'i' || tagName === 'em') newAnnotations.italic = true;
				if (tagName === 'u') newAnnotations.underline = true;
				if (tagName === 's' || tagName === 'strike' || tagName === 'del') newAnnotations.strikethrough = true;
				if (tagName === 'code') newAnnotations.code = true;

				// Handle links
				if (tagName === 'a') {
					const href = element.getAttribute('href');
					if (href) {
						const text = element.textContent || '';
						segments.push({
							type: 'text',
							text: { content: text, link: href },
							annotations: newAnnotations,
							plain_text: text,
							href
						});
						return;
					}
				}

				// Process child nodes
				node.childNodes.forEach(child => processNode(child, newAnnotations));
			}
		}

		const defaultAnnotations: RichText['annotations'] = {
			bold: false,
			italic: false,
			strikethrough: false,
			underline: false,
			code: false,
			color: 'default'
		};

		temp.childNodes.forEach(child => processNode(child, defaultAnnotations));

		return segments.length > 0 ? segments : [{
			type: 'text',
			text: { content: plainText, link: null },
			annotations: defaultAnnotations,
			plain_text: plainText,
			href: null
		}];
	}

	// Add mouseup listener for selection detection
	function handleMouseUp() {
		// Small delay to let selection complete
		setTimeout(handleSelectionChange, 10);
	}

	// Drag and drop handlers
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

		// Determine if dropping before or after this block
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

		// Calculate target index based on drop position
		let toIndex = index;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const midpoint = rect.top + rect.height / 2;

		if (e.clientY > midpoint) {
			toIndex = index + 1;
		}

		// Don't move if same position
		if (fromIndex === toIndex || fromIndex + 1 === toIndex) return;

		// Adjust toIndex if moving from before to after
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

	function addBlockOfType(type: BlockType) {
		onBlockAdd?.(createBlock(type));
		showAddMenu = false;
	}

	// Map callout icon names to Lucide components
	const calloutIcons: Record<string, typeof Lightbulb> = {
		Lightbulb,
		AlertCircle,
		Info,
		AlertTriangle,
		CheckCircle
	};

	function getCalloutIcon(iconName: string | unknown): typeof Lightbulb {
		if (typeof iconName === 'string' && iconName in calloutIcons) {
			return calloutIcons[iconName];
		}
		return Lightbulb; // Default
	}
</script>

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
		<div class="block-wrapper__controls">
			<Menu bind:open={showAddMenu}>
				{#snippet trigger()}
					<button class="block-wrapper__btn" aria-label="Add block">
						<Plus class="h-3.5 w-3.5" />
					</button>
				{/snippet}
				<MenuLabel>Add block</MenuLabel>
				{#each blockTypes as bt}
					<MenuItem onSelect={() => addBlockOfType(bt.type)}>
						<span class="block-type-icon">{bt.icon}</span>
						{bt.label}
					</MenuItem>
				{/each}
			</Menu>

			<button
				class="block-wrapper__btn block-wrapper__drag"
				aria-label="Drag block"
				draggable="true"
				ondragstart={handleDragStart}
				ondragend={handleDragEnd}
			>
				<GripVertical class="h-3.5 w-3.5" />
			</button>
		</div>
	{/if}

	<div class="block-content" data-block-type={block.type}>
		{#if block.type === 'paragraph'}
			<p
				class="block block--paragraph"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder={getPlaceholder('paragraph')}
				use:contenteditable
			></p>
		{:else if block.type === 'heading_1'}
			<h1
				class="block block--h1"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder="Heading 1"
				use:contenteditable
			></h1>
		{:else if block.type === 'heading_2'}
			<h2
				class="block block--h2"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder="Heading 2"
				use:contenteditable
			></h2>
		{:else if block.type === 'heading_3'}
			<h3
				class="block block--h3"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder="Heading 3"
				use:contenteditable
			></h3>
		{:else if block.type === 'bulleted_list'}
			<div class="block block--list">
				<span class="block__bullet">•</span>
				<span
					class="block__list-content"
					contenteditable={!readOnly}
					oninput={handleContentInput}
					onkeydown={handleKeydown}
					onmouseup={handleMouseUp}
					data-placeholder="List item"
					use:contenteditable
				></span>
			</div>
		{:else if block.type === 'numbered_list'}
			<div class="block block--list">
				<span class="block__number">{index + 1}.</span>
				<span
					class="block__list-content"
					contenteditable={!readOnly}
					oninput={handleContentInput}
					onkeydown={handleKeydown}
					onmouseup={handleMouseUp}
					data-placeholder="List item"
					use:contenteditable
				></span>
			</div>
		{:else if block.type === 'to_do'}
			<div class="block block--todo">
				<input
					type="checkbox"
					class="block__checkbox"
					checked={block.properties.checked ?? false}
					onchange={handleCheckboxChange}
					disabled={readOnly}
				/>
				<span
					class="block__todo-content"
					class:block__todo-content--checked={block.properties.checked}
					contenteditable={!readOnly}
					oninput={handleContentInput}
					onkeydown={handleKeydown}
					onmouseup={handleMouseUp}
					data-placeholder="To-do"
					use:contenteditable
				></span>
			</div>
		{:else if block.type === 'quote'}
			<blockquote
				class="block block--quote"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder="Quote"
				use:contenteditable
			></blockquote>
		{:else if block.type === 'divider'}
			{@const dividerStyle = block.properties?.divider_style || 'solid'}
			<div class="divider-wrapper">
				<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
				<div
					class="block block--divider block--divider--{dividerStyle}"
					onclick={() => !readOnly && showDividerStylePicker()}
					role={readOnly ? undefined : 'button'}
					tabindex={readOnly ? undefined : 0}
				>
					{#if dividerStyle === 'gradient'}
						<div class="divider-gradient"></div>
					{:else}
						<hr />
					{/if}
				</div>
				{#if showDividerPicker && !readOnly}
					<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
					<div class="divider-picker" onclick={(e) => e.stopPropagation()}>
						<div class="divider-picker__label">Divider Style</div>
						<div class="divider-picker__options">
							{#each dividerStyles as style}
								<button
									class="divider-picker__option"
									class:divider-picker__option--active={dividerStyle === style.value}
									onclick={() => selectDividerStyle(style.value)}
								>
									<span class="divider-picker__preview divider-picker__preview--{style.value}"></span>
									<span>{style.label}</span>
								</button>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else if block.type === 'code'}
			<pre class="block block--code"><code
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				data-placeholder="Code"
				use:contenteditable
			></code></pre>
		{:else if block.type === 'callout'}
			{@const CalloutIcon = getCalloutIcon(block.properties.icon)}
			<div class="block block--callout">
				<span class="block__callout-icon">
					<CalloutIcon class="h-5 w-5" />
				</span>
				<span
					class="block__callout-content"
					contenteditable={!readOnly}
					oninput={handleContentInput}
					onkeydown={handleKeydown}
					onmouseup={handleMouseUp}
					data-placeholder="Callout"
					use:contenteditable
				></span>
			</div>
		{:else if block.type === 'toggle'}
			<div class="block block--toggle">
				<button
					class="block__toggle-btn"
					onclick={() => onBlockChange?.({
						...block,
						properties: { ...block.properties, expanded: !block.properties.expanded }
					})}
					disabled={readOnly}
				>
					{#if block.properties.expanded}
						<ChevronDown class="h-4 w-4" />
					{:else}
						<ChevronRight class="h-4 w-4" />
					{/if}
				</button>
				<span
					class="block__toggle-content"
					contenteditable={!readOnly}
					oninput={handleContentInput}
					onkeydown={handleKeydown}
					onmouseup={handleMouseUp}
					data-placeholder="Toggle"
					use:contenteditable
				></span>
			</div>
		{:else if block.type === 'image'}
			<div class="block block--image">
				{#if block.properties.url}
					<img
						src={block.properties.url as string}
						alt={typeof block.properties.caption === 'string' ? block.properties.caption : 'Image'}
						class="block__image-img"
					/>
					{#if block.properties.caption}
						<p class="block__image-caption">{typeof block.properties.caption === 'string' ? block.properties.caption : ''}</p>
					{/if}
				{:else}
					<div class="block__image-placeholder">
						<ImageIcon class="h-8 w-8" />
						<span>Click to add an image</span>
						<input
							type="file"
							accept="image/*"
							class="block__image-input"
							onchange={(e) => {
								const file = (e.target as HTMLInputElement).files?.[0];
								if (file) {
									const url = URL.createObjectURL(file);
									onBlockChange?.({
										...block,
										properties: { ...block.properties, url }
									});
								}
							}}
						/>
					</div>
				{/if}
			</div>
		{:else if block.type === 'bookmark'}
			<div class="block block--bookmark">
				{#if block.properties.url}
					<a
						href={block.properties.url as string}
						target="_blank"
						rel="noopener noreferrer"
						class="block__bookmark-link"
					>
						<div class="block__bookmark-content">
							<div class="block__bookmark-title">
								{block.properties.title || block.properties.url}
							</div>
							{#if block.properties.description}
								<div class="block__bookmark-description">
									{block.properties.description}
								</div>
							{/if}
							<div class="block__bookmark-url">
								<Link2 class="h-3 w-3" />
								{block.properties.url}
							</div>
						</div>
						<ExternalLink class="h-4 w-4 block__bookmark-icon" />
					</a>
				{:else}
					<div class="block__bookmark-empty">
						<Link2 class="h-5 w-5" />
						<input
							type="url"
							placeholder="Paste a link..."
							class="block__bookmark-input"
							onkeydown={(e) => {
								if (e.key === 'Enter') {
									const url = (e.target as HTMLInputElement).value;
									if (url) {
										onBlockChange?.({
											...block,
											properties: { ...block.properties, url, title: url }
										});
									}
								}
							}}
						/>
					</div>
				{/if}
			</div>
		{:else if block.type === 'table'}
			<TableBlock
				{block}
				{readOnly}
				onBlockChange={(updated) => onBlockChange?.(updated)}
			/>
		{:else}
			<p
				class="block block--paragraph"
				contenteditable={!readOnly}
				oninput={handleContentInput}
				onkeydown={handleKeydown}
				onmouseup={handleMouseUp}
				use:contenteditable
			></p>
		{/if}
	</div>

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

	.block-wrapper__controls {
		position: absolute;
		left: 0;
		top: 2px;
		display: flex;
		align-items: center;
		gap: 2px;
		opacity: 0;
		transition: opacity 0.15s;
	}

	.block-wrapper--hovered .block-wrapper__controls {
		opacity: 1;
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
		background-color: hsl(var(--primary));
		border-radius: 1px;
	}

	.block-wrapper--drag-after::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 52px;
		right: 0;
		height: 2px;
		background-color: hsl(var(--primary));
		border-radius: 1px;
	}

	.block-wrapper__btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		padding: 0;
		background: transparent;
		border: none;
		border-radius: 0.25rem;
		color: hsl(var(--muted-foreground));
		cursor: pointer;
		transition: background-color 0.1s;
	}

	.block-wrapper__btn:hover {
		background-color: hsl(var(--muted));
	}

	.block-wrapper__drag {
		cursor: grab;
	}

	.block-wrapper__drag:active {
		cursor: grabbing;
	}

	.block-content {
		flex: 1;
		min-width: 0;
	}

	.block {
		outline: none;
		word-break: break-word;
	}

	.block:empty::before {
		content: attr(data-placeholder);
		color: hsl(var(--muted-foreground) / 0.5);
	}

	.block--paragraph {
		font-size: 1rem;
		line-height: 1.6;
		margin: 0;
		padding: 0.125rem 0;
	}

	.block--h1 {
		font-size: 1.875rem;
		font-weight: 700;
		line-height: 1.3;
		margin: 1rem 0 0.5rem;
	}

	.block--h2 {
		font-size: 1.5rem;
		font-weight: 600;
		line-height: 1.35;
		margin: 0.875rem 0 0.375rem;
	}

	.block--h3 {
		font-size: 1.25rem;
		font-weight: 600;
		line-height: 1.4;
		margin: 0.75rem 0 0.25rem;
	}

	.block--list {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		padding: 0.125rem 0;
	}

	.block__bullet,
	.block__number {
		color: hsl(var(--muted-foreground));
		font-size: 1rem;
		line-height: 1.6;
		min-width: 1.25rem;
	}

	.block__list-content {
		flex: 1;
		outline: none;
	}

	.block--todo {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		padding: 0.125rem 0;
	}

	.block__checkbox {
		width: 16px;
		height: 16px;
		margin-top: 0.25rem;
		cursor: pointer;
	}

	.block__todo-content {
		flex: 1;
		outline: none;
	}

	.block__todo-content--checked {
		text-decoration: line-through;
		color: hsl(var(--muted-foreground));
	}

	.block--quote {
		border-left: 3px solid hsl(var(--border));
		padding-left: 1rem;
		margin: 0.25rem 0;
		font-style: italic;
		color: hsl(var(--muted-foreground));
	}

	/* Divider wrapper */
	.divider-wrapper {
		position: relative;
		margin: 1rem 0;
	}

	.block--divider {
		cursor: pointer;
		padding: 0.5rem 0;
		transition: opacity 0.15s;
	}

	.block--divider:hover {
		opacity: 0.7;
	}

	.block--divider hr {
		border: none;
		margin: 0;
	}

	/* Solid (default) */
	.block--divider--solid hr {
		height: 1px;
		background-color: #e5e5e5;
		background-color: var(--bos-border-color, #e3e2e4);
	}

	/* Dashed */
	.block--divider--dashed hr {
		height: 1px;
		background-image: repeating-linear-gradient(
			90deg,
			var(--bos-border-color, #e3e2e4) 0px,
			var(--bos-border-color, #e3e2e4) 8px,
			transparent 8px,
			transparent 16px
		);
	}

	/* Dotted */
	.block--divider--dotted hr {
		height: 2px;
		background-image: repeating-linear-gradient(
			90deg,
			var(--bos-border-color, #e3e2e4) 0px,
			var(--bos-border-color, #e3e2e4) 3px,
			transparent 3px,
			transparent 8px
		);
	}

	/* Thick */
	.block--divider--thick hr {
		height: 3px;
		background-color: var(--bos-border-color, #e3e2e4);
	}

	/* Double */
	.block--divider--double hr {
		height: 5px;
		background-image: linear-gradient(
			to bottom,
			var(--bos-border-color, #e3e2e4) 0px,
			var(--bos-border-color, #e3e2e4) 1px,
			transparent 1px,
			transparent 4px,
			var(--bos-border-color, #e3e2e4) 4px,
			var(--bos-border-color, #e3e2e4) 5px
		);
	}

	/* Dark mode divider colors */
	:global(.dark) .block--divider--solid hr,
	:global(.dark) .block--divider--thick hr {
		background-color: #3c3c42;
	}

	:global(.dark) .block--divider--dashed hr {
		background-image: repeating-linear-gradient(
			90deg,
			#3c3c42 0px,
			#3c3c42 8px,
			transparent 8px,
			transparent 16px
		);
	}

	:global(.dark) .block--divider--dotted hr {
		background-image: repeating-linear-gradient(
			90deg,
			#3c3c42 0px,
			#3c3c42 3px,
			transparent 3px,
			transparent 8px
		);
	}

	:global(.dark) .block--divider--double hr {
		background-image: linear-gradient(
			to bottom,
			#3c3c42 0px,
			#3c3c42 1px,
			transparent 1px,
			transparent 4px,
			#3c3c42 4px,
			#3c3c42 5px
		);
	}

	/* Gradient */
	.divider-gradient {
		height: 2px;
		background: linear-gradient(90deg,
			transparent 0%,
			hsl(var(--primary) / 0.5) 20%,
			hsl(var(--primary)) 50%,
			hsl(var(--primary) / 0.5) 80%,
			transparent 100%
		);
		border-radius: 1px;
	}

	/* Divider style picker */
	.divider-picker {
		position: absolute;
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		z-index: 50;
		margin-top: 0.5rem;
		padding: 0.5rem;
		background: hsl(var(--background));
		border: 1px solid hsl(var(--border));
		border-radius: 0.5rem;
		box-shadow: 0 4px 16px hsl(var(--foreground) / 0.12);
		min-width: 160px;
	}

	.divider-picker__label {
		font-size: 0.75rem;
		font-weight: 500;
		color: hsl(var(--muted-foreground));
		margin-bottom: 0.5rem;
		padding: 0 0.25rem;
	}

	.divider-picker__options {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.divider-picker__option {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0.5rem;
		background: transparent;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		color: hsl(var(--foreground));
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.divider-picker__option:hover {
		background: hsl(var(--muted));
	}

	.divider-picker__option--active {
		background: hsl(var(--primary) / 0.1);
		color: hsl(var(--primary));
	}

	.divider-picker__preview {
		width: 40px;
		height: 2px;
		flex-shrink: 0;
		border-radius: 1px;
	}

	.divider-picker__preview--solid {
		background-color: currentColor;
		opacity: 0.3;
	}

	.divider-picker__preview--dashed {
		background-image: repeating-linear-gradient(
			90deg,
			currentColor 0px,
			currentColor 4px,
			transparent 4px,
			transparent 8px
		);
		opacity: 0.3;
	}

	.divider-picker__preview--dotted {
		background-image: repeating-linear-gradient(
			90deg,
			currentColor 0px,
			currentColor 2px,
			transparent 2px,
			transparent 6px
		);
		opacity: 0.3;
	}

	.divider-picker__preview--thick {
		height: 4px;
		background-color: currentColor;
		opacity: 0.3;
	}

	.divider-picker__preview--double {
		height: 6px;
		background-image: linear-gradient(
			to bottom,
			currentColor 0px,
			currentColor 2px,
			transparent 2px,
			transparent 4px,
			currentColor 4px,
			currentColor 6px
		);
		opacity: 0.3;
	}

	.divider-picker__preview--gradient {
		background: linear-gradient(90deg,
			transparent 0%,
			hsl(var(--primary)) 50%,
			transparent 100%
		);
	}

	.block--code {
		background-color: hsl(var(--muted));
		border-radius: 0.375rem;
		padding: 1rem;
		margin: 0.25rem 0;
		font-family: ui-monospace, monospace;
		font-size: 0.875rem;
		overflow-x: auto;
	}

	.block--code code {
		display: block;
		outline: none;
	}

	.block--callout {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		background-color: hsl(var(--muted));
		border-radius: 0.375rem;
		padding: 1rem;
		margin: 0.25rem 0;
	}

	.block__callout-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		color: hsl(var(--foreground));
		flex-shrink: 0;
	}

	.block__callout-content {
		flex: 1;
		outline: none;
	}

	/* Toggle block */
	.block--toggle {
		display: flex;
		align-items: flex-start;
		gap: 0.25rem;
		padding: 0.125rem 0;
	}

	.block__toggle-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		padding: 0;
		background: transparent;
		border: none;
		border-radius: 0.25rem;
		color: hsl(var(--muted-foreground));
		cursor: pointer;
		transition: background-color 0.1s;
		flex-shrink: 0;
	}

	.block__toggle-btn:hover {
		background-color: hsl(var(--muted));
	}

	.block__toggle-content {
		flex: 1;
		outline: none;
		line-height: 1.6;
	}

	/* Image block */
	.block--image {
		margin: 0.5rem 0;
	}

	.block__image-img {
		max-width: 100%;
		border-radius: 0.375rem;
	}

	.block__image-caption {
		font-size: 0.875rem;
		color: hsl(var(--muted-foreground));
		text-align: center;
		margin-top: 0.5rem;
	}

	.block__image-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 2rem;
		background-color: hsl(var(--muted) / 0.5);
		border: 2px dashed hsl(var(--border));
		border-radius: 0.5rem;
		color: hsl(var(--muted-foreground));
		cursor: pointer;
		position: relative;
		transition: border-color 0.15s, background-color 0.15s;
	}

	.block__image-placeholder:hover {
		border-color: hsl(var(--muted-foreground));
		background-color: hsl(var(--muted));
	}

	.block__image-input {
		position: absolute;
		inset: 0;
		opacity: 0;
		cursor: pointer;
	}

	/* Bookmark block */
	.block--bookmark {
		margin: 0.5rem 0;
	}

	.block__bookmark-link {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background-color: hsl(var(--muted) / 0.5);
		border: 1px solid hsl(var(--border));
		border-radius: 0.5rem;
		text-decoration: none;
		color: inherit;
		transition: background-color 0.15s;
	}

	.block__bookmark-link:hover {
		background-color: hsl(var(--muted));
	}

	.block__bookmark-content {
		flex: 1;
		min-width: 0;
	}

	.block__bookmark-title {
		font-weight: 500;
		color: hsl(var(--foreground));
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.block__bookmark-description {
		font-size: 0.875rem;
		color: hsl(var(--muted-foreground));
		margin-top: 0.25rem;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.block__bookmark-url {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		color: hsl(var(--muted-foreground));
		margin-top: 0.5rem;
	}

	.block__bookmark-icon {
		color: hsl(var(--muted-foreground));
		flex-shrink: 0;
	}

	.block__bookmark-empty {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		background-color: hsl(var(--muted) / 0.5);
		border: 1px solid hsl(var(--border));
		border-radius: 0.5rem;
		color: hsl(var(--muted-foreground));
	}

	.block__bookmark-input {
		flex: 1;
		background: transparent;
		border: none;
		outline: none;
		font-size: 0.875rem;
		color: hsl(var(--foreground));
	}

	.block__bookmark-input::placeholder {
		color: hsl(var(--muted-foreground));
	}

	.block-type-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		margin-right: 0.5rem;
		font-size: 0.75rem;
		font-weight: 500;
		color: hsl(var(--muted-foreground));
	}

	/* Inline formatting styles */
	.block :global(.inline-code),
	.block :global(code:not([class])) {
		background-color: hsl(var(--muted));
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		font-family: ui-monospace, monospace;
		font-size: 0.875em;
		color: hsl(var(--foreground));
	}

	.block :global(a) {
		color: hsl(var(--primary));
		text-decoration: underline;
		text-underline-offset: 2px;
	}

	.block :global(a:hover) {
		text-decoration-thickness: 2px;
	}

	/* Bold, italic, underline, strikethrough are handled by browser's default styles */
	.block :global(b),
	.block :global(strong) {
		font-weight: 600;
	}

	.block :global(i),
	.block :global(em) {
		font-style: italic;
	}

	.block :global(u) {
		text-decoration: underline;
	}

	.block :global(s),
	.block :global(strike),
	.block :global(del) {
		text-decoration: line-through;
	}
</style>
