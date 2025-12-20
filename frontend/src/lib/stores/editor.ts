import { writable, derived } from 'svelte/store';
import type { Block } from '$lib/api';

export type BlockType =
	| 'paragraph'
	| 'heading1'
	| 'heading2'
	| 'heading3'
	| 'bulletList'
	| 'numberedList'
	| 'todo'
	| 'quote'
	| 'code'
	| 'divider'
	| 'callout'
	| 'image'
	| 'table'
	| 'embed'
	| 'artifact'
	| 'page';

export interface EditorBlock extends Block {
	id: string;
	type: BlockType;
	content: string;
	properties?: {
		checked?: boolean;
		language?: string;
		artifactId?: string;
		url?: string;
		caption?: string;
		calloutType?: 'info' | 'warning' | 'success' | 'error';
		[key: string]: unknown;
	};
	children?: EditorBlock[];
}

interface EditorState {
	blocks: EditorBlock[];
	focusedBlockId: string | null;
	focusedBlockIndex: number;
	selectionStart: number;
	selectionEnd: number;
	isDirty: boolean;
	isSaving: boolean;
	lastSavedAt: Date | null;
	showSlashMenu: boolean;
	slashMenuPosition: { x: number; y: number } | null;
	slashMenuQuery: string;
	showAIPanel: boolean;
}

function generateBlockId(): string {
	return Math.random().toString(36).substring(2, 11);
}

export function createEmptyBlock(type: BlockType = 'paragraph'): EditorBlock {
	return {
		id: generateBlockId(),
		type,
		content: '',
		properties: type === 'todo' ? { checked: false } : undefined
	};
}

function createEditorStore() {
	const { subscribe, update, set } = writable<EditorState>({
		blocks: [createEmptyBlock()],
		focusedBlockId: null,
		focusedBlockIndex: 0,
		selectionStart: 0,
		selectionEnd: 0,
		isDirty: false,
		isSaving: false,
		lastSavedAt: null,
		showSlashMenu: false,
		slashMenuPosition: null,
		slashMenuQuery: '',
		showAIPanel: false
	});

	return {
		subscribe,

		initialize(blocks: Block[] | null) {
			const editorBlocks: EditorBlock[] =
				blocks && blocks.length > 0
					? blocks.map((b) => ({
							id: b.id || generateBlockId(),
							type: (b.type as BlockType) || 'paragraph',
							content: b.content || '',
							properties: b.properties as EditorBlock['properties'],
							children: b.children as EditorBlock[]
						}))
					: [createEmptyBlock()];

			update((s) => ({
				...s,
				blocks: editorBlocks,
				focusedBlockId: editorBlocks[0]?.id || null,
				focusedBlockIndex: 0,
				isDirty: false
			}));
		},

		setBlocks(blocks: EditorBlock[]) {
			update((s) => ({ ...s, blocks, isDirty: true }));
		},

		updateBlock(id: string, content: string, properties?: EditorBlock['properties']) {
			update((s) => ({
				...s,
				blocks: s.blocks.map((b) =>
					b.id === id ? { ...b, content, properties: properties ?? b.properties } : b
				),
				isDirty: true
			}));
		},

		addBlockAfter(afterId: string, type: BlockType = 'paragraph'): string {
			const newBlock = createEmptyBlock(type);
			update((s) => {
				const index = s.blocks.findIndex((b) => b.id === afterId);
				const newBlocks = [...s.blocks];
				newBlocks.splice(index + 1, 0, newBlock);
				return {
					...s,
					blocks: newBlocks,
					focusedBlockId: newBlock.id,
					focusedBlockIndex: index + 1,
					isDirty: true
				};
			});
			return newBlock.id;
		},

		addBlockBefore(beforeId: string, type: BlockType = 'paragraph'): string {
			const newBlock = createEmptyBlock(type);
			update((s) => {
				const index = s.blocks.findIndex((b) => b.id === beforeId);
				const newBlocks = [...s.blocks];
				newBlocks.splice(index, 0, newBlock);
				return {
					...s,
					blocks: newBlocks,
					focusedBlockId: newBlock.id,
					focusedBlockIndex: index,
					isDirty: true
				};
			});
			return newBlock.id;
		},

		deleteBlock(id: string) {
			update((s) => {
				if (s.blocks.length <= 1) {
					// Don't delete the last block, just clear it
					return {
						...s,
						blocks: [createEmptyBlock()],
						focusedBlockIndex: 0,
						isDirty: true
					};
				}
				const index = s.blocks.findIndex((b) => b.id === id);
				const newBlocks = s.blocks.filter((b) => b.id !== id);
				const newFocusIndex = Math.min(index, newBlocks.length - 1);
				return {
					...s,
					blocks: newBlocks,
					focusedBlockId: newBlocks[newFocusIndex]?.id || null,
					focusedBlockIndex: newFocusIndex,
					isDirty: true
				};
			});
		},

		changeBlockType(id: string, newType: BlockType) {
			update((s) => ({
				...s,
				blocks: s.blocks.map((b) =>
					b.id === id
						? {
								...b,
								type: newType,
								properties:
									newType === 'todo' ? { ...b.properties, checked: false } : b.properties
							}
						: b
				),
				isDirty: true,
				showSlashMenu: false,
				slashMenuQuery: ''
			}));
		},

		moveBlockUp(id: string) {
			update((s) => {
				const index = s.blocks.findIndex((b) => b.id === id);
				if (index <= 0) return s;
				const newBlocks = [...s.blocks];
				[newBlocks[index - 1], newBlocks[index]] = [newBlocks[index], newBlocks[index - 1]];
				return { ...s, blocks: newBlocks, focusedBlockIndex: index - 1, isDirty: true };
			});
		},

		moveBlockDown(id: string) {
			update((s) => {
				const index = s.blocks.findIndex((b) => b.id === id);
				if (index >= s.blocks.length - 1) return s;
				const newBlocks = [...s.blocks];
				[newBlocks[index], newBlocks[index + 1]] = [newBlocks[index + 1], newBlocks[index]];
				return { ...s, blocks: newBlocks, focusedBlockIndex: index + 1, isDirty: true };
			});
		},

		setFocusedBlock(id: string | null) {
			update((s) => {
				const index = id ? s.blocks.findIndex((b) => b.id === id) : 0;
				return { ...s, focusedBlockId: id, focusedBlockIndex: index >= 0 ? index : 0 };
			});
		},

		focusNextBlock() {
			update((s) => {
				const newIndex = Math.min(s.focusedBlockIndex + 1, s.blocks.length - 1);
				return {
					...s,
					focusedBlockIndex: newIndex,
					focusedBlockId: s.blocks[newIndex]?.id || null
				};
			});
		},

		focusPreviousBlock() {
			update((s) => {
				const newIndex = Math.max(s.focusedBlockIndex - 1, 0);
				return {
					...s,
					focusedBlockIndex: newIndex,
					focusedBlockId: s.blocks[newIndex]?.id || null
				};
			});
		},

		showSlashMenu(position: { x: number; y: number }) {
			update((s) => ({
				...s,
				showSlashMenu: true,
				slashMenuPosition: position,
				slashMenuQuery: ''
			}));
		},

		hideSlashMenu() {
			update((s) => ({
				...s,
				showSlashMenu: false,
				slashMenuPosition: null,
				slashMenuQuery: ''
			}));
		},

		setSlashMenuQuery(query: string) {
			update((s) => ({ ...s, slashMenuQuery: query }));
		},

		toggleAIPanel() {
			update((s) => ({ ...s, showAIPanel: !s.showAIPanel }));
		},

		showAIPanel() {
			update((s) => ({ ...s, showAIPanel: true }));
		},

		hideAIPanel() {
			update((s) => ({ ...s, showAIPanel: false }));
		},

		setSaving(isSaving: boolean) {
			update((s) => ({ ...s, isSaving }));
		},

		markSaved() {
			update((s) => ({ ...s, isDirty: false, isSaving: false, lastSavedAt: new Date() }));
		},

		toggleTodo(id: string) {
			update((s) => ({
				...s,
				blocks: s.blocks.map((b) =>
					b.id === id && b.type === 'todo'
						? { ...b, properties: { ...b.properties, checked: !b.properties?.checked } }
						: b
				),
				isDirty: true
			}));
		},

		getBlocks(): EditorBlock[] {
			let currentBlocks: EditorBlock[] = [];
			subscribe((s) => {
				currentBlocks = s.blocks;
			})();
			return currentBlocks;
		},

		reset() {
			set({
				blocks: [createEmptyBlock()],
				focusedBlockId: null,
				focusedBlockIndex: 0,
				selectionStart: 0,
				selectionEnd: 0,
				isDirty: false,
				isSaving: false,
				lastSavedAt: null,
				showSlashMenu: false,
				slashMenuPosition: null,
				slashMenuQuery: '',
				showAIPanel: false
			});
		}
	};
}

export const editor = createEditorStore();

// Derived store for word count
export const wordCount = derived(editor, ($editor) => {
	return $editor.blocks.reduce((count, block) => {
		if (block.content) {
			return count + block.content.trim().split(/\s+/).filter(Boolean).length;
		}
		return count;
	}, 0);
});

// Block type definitions for slash menu
export const blockTypes: {
	type: BlockType;
	label: string;
	description: string;
	icon: string;
	shortcut?: string;
}[] = [
	{ type: 'paragraph', label: 'Text', description: 'Plain text', icon: 'T', shortcut: '/text' },
	{
		type: 'heading1',
		label: 'Heading 1',
		description: 'Large heading',
		icon: 'H1',
		shortcut: '/h1'
	},
	{
		type: 'heading2',
		label: 'Heading 2',
		description: 'Medium heading',
		icon: 'H2',
		shortcut: '/h2'
	},
	{
		type: 'heading3',
		label: 'Heading 3',
		description: 'Small heading',
		icon: 'H3',
		shortcut: '/h3'
	},
	{
		type: 'bulletList',
		label: 'Bullet List',
		description: 'Unordered list',
		icon: '•',
		shortcut: '/bullet'
	},
	{
		type: 'numberedList',
		label: 'Numbered List',
		description: 'Ordered list',
		icon: '1.',
		shortcut: '/numbered'
	},
	{
		type: 'todo',
		label: 'To-do',
		description: 'Checkbox item',
		icon: '☐',
		shortcut: '/todo'
	},
	{
		type: 'quote',
		label: 'Quote',
		description: 'Block quote',
		icon: '"',
		shortcut: '/quote'
	},
	{
		type: 'code',
		label: 'Code',
		description: 'Code block',
		icon: '</>',
		shortcut: '/code'
	},
	{
		type: 'divider',
		label: 'Divider',
		description: 'Horizontal line',
		icon: '—',
		shortcut: '/divider'
	},
	{
		type: 'callout',
		label: 'Callout',
		description: 'Highlighted box',
		icon: '!',
		shortcut: '/callout'
	},
	{
		type: 'page',
		label: 'Page',
		description: 'Create a nested sub-page',
		icon: '📄',
		shortcut: '/page'
	}
];
