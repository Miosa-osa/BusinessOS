# Block Editor Implementation Guide for BusinessOS

## Phase 1: Foundation (Week 1)

### Step 1: Create Type Definitions

```typescript
// src/lib/types/blocks.ts

export type BlockType =
  | 'paragraph'
  | 'heading'
  | 'list'
  | 'code'
  | 'image'
  | 'divider'
  | 'database';

export interface Block {
  id: string;
  type: BlockType;
  parent: string | null;
  children: string[];
  props: BlockProps;
  content: string;
  createdAt: Date;
  updatedAt: Date;
}

export type BlockProps =
  | ParagraphProps
  | HeadingProps
  | ListProps
  | CodeProps
  | ImageProps
  | DividerProps
  | DatabaseProps;

export interface ParagraphProps {
  textAlign?: 'left' | 'center' | 'right';
  color?: string;
}

export interface HeadingProps {
  level: 1 | 2 | 3 | 4 | 5 | 6;
  color?: string;
  collapsed?: boolean;
}

export interface ListProps {
  type: 'bulleted' | 'numbered' | 'todo';
  checked?: boolean;
  collapsed?: boolean;
}

export interface CodeProps {
  language: string;
  wrap: boolean;
  showLineNumbers: boolean;
}

export interface ImageProps {
  url: string;
  caption?: string;
  width?: number;
  height?: number;
}

export interface DividerProps {
  style?: 'solid' | 'dashed' | 'dotted';
}

export interface DatabaseProps {
  // To be implemented in Phase 5
}
```

### Step 2: Create Editor Store

```typescript
// src/lib/stores/editor.ts

import { writable, derived, get } from 'svelte/store';
import type { Block, BlockType, BlockProps } from '$lib/types/blocks';
import { nanoid } from 'nanoid';

interface EditorState {
  blocks: Map<string, Block>;
  rootBlocks: string[]; // Top-level blocks
  focusedBlockId: string | null;
}

function createEditorStore() {
  const { subscribe, update, set } = writable<EditorState>({
    blocks: new Map(),
    rootBlocks: [],
    focusedBlockId: null,
  });

  return {
    subscribe,

    // Initialize with default content
    init() {
      const firstBlock = createBlock('paragraph', null);
      set({
        blocks: new Map([[firstBlock.id, firstBlock]]),
        rootBlocks: [firstBlock.id],
        focusedBlockId: firstBlock.id,
      });
    },

    // Block CRUD
    addBlock(type: BlockType, parentId: string | null, position?: number) {
      const newBlock = createBlock(type, parentId);

      update(state => {
        state.blocks.set(newBlock.id, newBlock);

        if (parentId) {
          // Add as child
          const parent = state.blocks.get(parentId);
          if (parent) {
            if (position !== undefined) {
              parent.children.splice(position, 0, newBlock.id);
            } else {
              parent.children.push(newBlock.id);
            }
            state.blocks.set(parentId, parent);
          }
        } else {
          // Add as root block
          if (position !== undefined) {
            state.rootBlocks.splice(position, 0, newBlock.id);
          } else {
            state.rootBlocks.push(newBlock.id);
          }
        }

        state.focusedBlockId = newBlock.id;
        return state;
      });

      return newBlock.id;
    },

    updateBlock(id: string, updates: Partial<Omit<Block, 'id'>>) {
      update(state => {
        const block = state.blocks.get(id);
        if (block) {
          state.blocks.set(id, {
            ...block,
            ...updates,
            updatedAt: new Date(),
          });
        }
        return state;
      });
    },

    deleteBlock(id: string) {
      update(state => {
        const block = state.blocks.get(id);
        if (!block) return state;

        // Remove from parent's children or root blocks
        if (block.parent) {
          const parent = state.blocks.get(block.parent);
          if (parent) {
            parent.children = parent.children.filter(childId => childId !== id);
            state.blocks.set(block.parent, parent);
          }
        } else {
          state.rootBlocks = state.rootBlocks.filter(blockId => blockId !== id);
        }

        // Recursively delete children
        block.children.forEach(childId => {
          deleteBlockRecursive(state, childId);
        });

        // Delete the block itself
        state.blocks.delete(id);

        // Update focus
        if (state.focusedBlockId === id) {
          state.focusedBlockId = null;
        }

        return state;
      });
    },

    moveBlock(blockId: string, targetId: string, position: 'before' | 'after' | 'inside') {
      update(state => {
        const block = state.blocks.get(blockId);
        const target = state.blocks.get(targetId);

        if (!block || !target) return state;

        // Remove from current position
        if (block.parent) {
          const parent = state.blocks.get(block.parent);
          if (parent) {
            parent.children = parent.children.filter(id => id !== blockId);
            state.blocks.set(block.parent, parent);
          }
        } else {
          state.rootBlocks = state.rootBlocks.filter(id => id !== blockId);
        }

        // Add to new position
        if (position === 'inside') {
          block.parent = targetId;
          target.children.push(blockId);
          state.blocks.set(targetId, target);
        } else {
          block.parent = target.parent;

          if (target.parent) {
            const parent = state.blocks.get(target.parent);
            if (parent) {
              const targetIndex = parent.children.indexOf(targetId);
              const insertIndex = position === 'before' ? targetIndex : targetIndex + 1;
              parent.children.splice(insertIndex, 0, blockId);
              state.blocks.set(target.parent, parent);
            }
          } else {
            const targetIndex = state.rootBlocks.indexOf(targetId);
            const insertIndex = position === 'before' ? targetIndex : targetIndex + 1;
            state.rootBlocks.splice(insertIndex, 0, blockId);
          }
        }

        state.blocks.set(blockId, block);
        return state;
      });
    },

    // Convert block type
    convertBlock(id: string, newType: BlockType) {
      update(state => {
        const block = state.blocks.get(id);
        if (!block) return state;

        const newProps = getDefaultPropsForType(newType);

        state.blocks.set(id, {
          ...block,
          type: newType,
          props: newProps,
          updatedAt: new Date(),
        });

        return state;
      });
    },

    // Focus management
    setFocus(blockId: string | null) {
      update(state => {
        state.focusedBlockId = blockId;
        return state;
      });
    },

    // Get block by ID
    getBlock(id: string) {
      return get({ subscribe }).blocks.get(id);
    },
  };
}

// Helper functions
function createBlock(type: BlockType, parent: string | null): Block {
  return {
    id: nanoid(),
    type,
    parent,
    children: [],
    props: getDefaultPropsForType(type),
    content: '',
    createdAt: new Date(),
    updatedAt: new Date(),
  };
}

function getDefaultPropsForType(type: BlockType): BlockProps {
  switch (type) {
    case 'paragraph':
      return {} as ParagraphProps;
    case 'heading':
      return { level: 1 } as HeadingProps;
    case 'list':
      return { type: 'bulleted' } as ListProps;
    case 'code':
      return { language: 'javascript', wrap: false, showLineNumbers: true } as CodeProps;
    case 'image':
      return { url: '' } as ImageProps;
    case 'divider':
      return {} as DividerProps;
    case 'database':
      return {} as DatabaseProps;
    default:
      return {} as ParagraphProps;
  }
}

function deleteBlockRecursive(state: EditorState, id: string) {
  const block = state.blocks.get(id);
  if (block) {
    block.children.forEach(childId => deleteBlockRecursive(state, childId));
    state.blocks.delete(id);
  }
}

export const editor = createEditorStore();

// Derived stores
export const focusedBlock = derived(
  editor,
  $editor => $editor.focusedBlockId ? $editor.blocks.get($editor.focusedBlockId) : null
);

export const blockTree = derived(
  editor,
  $editor => buildTree($editor.blocks, $editor.rootBlocks)
);

function buildTree(blocks: Map<string, Block>, rootIds: string[]): Block[] {
  return rootIds
    .map(id => blocks.get(id))
    .filter((block): block is Block => block !== undefined);
}
```

### Step 3: Create Block Registry

```typescript
// src/lib/blocks/registry.ts

import type { ComponentType, SvelteComponent } from 'svelte';
import type { BlockType } from '$lib/types/blocks';

// Import block components (we'll create these next)
import ParagraphBlock from './ParagraphBlock.svelte';
import HeadingBlock from './HeadingBlock.svelte';
import ListBlock from './ListBlock.svelte';
import CodeBlock from './CodeBlock.svelte';
import ImageBlock from './ImageBlock.svelte';
import DividerBlock from './DividerBlock.svelte';

type BlockComponent = ComponentType<SvelteComponent>;

const registry: Record<BlockType, BlockComponent> = {
  paragraph: ParagraphBlock,
  heading: HeadingBlock,
  list: ListBlock,
  code: CodeBlock,
  image: ImageBlock,
  divider: DividerBlock,
  database: ParagraphBlock, // Placeholder
};

export function getBlockComponent(type: BlockType): BlockComponent {
  return registry[type] ?? ParagraphBlock;
}

export function registerBlockComponent(type: BlockType, component: BlockComponent) {
  registry[type] = component;
}
```

### Step 4: Create Base Block Components

```svelte
<!-- src/lib/blocks/ParagraphBlock.svelte -->
<script lang="ts">
  import type { Block, ParagraphProps } from '$lib/types/blocks';
  import { editor } from '$lib/stores/editor';

  export let block: Block;

  $: props = block.props as ParagraphProps;

  function handleInput(e: Event) {
    const target = e.target as HTMLElement;
    editor.updateBlock(block.id, { content: target.textContent ?? '' });
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      // Create new paragraph block after this one
      const parent = block.parent;
      const siblings = parent
        ? $editor.blocks.get(parent)?.children ?? []
        : $editor.rootBlocks;
      const currentIndex = siblings.indexOf(block.id);

      editor.addBlock('paragraph', parent, currentIndex + 1);
    } else if (e.key === 'Backspace' && block.content === '') {
      e.preventDefault();
      editor.deleteBlock(block.id);
    }
  }

  function handleFocus() {
    editor.setFocus(block.id);
  }
</script>

<div
  class="block-paragraph"
  contenteditable="true"
  bind:textContent={block.content}
  on:input={handleInput}
  on:keydown={handleKeydown}
  on:focus={handleFocus}
  style:text-align={props.textAlign}
  style:color={props.color}
  data-block-id={block.id}
/>

<style>
  .block-paragraph {
    min-height: 1.5em;
    padding: 3px 2px;
    outline: none;
    cursor: text;
  }

  .block-paragraph:empty::before {
    content: 'Type / for commands...';
    color: #999;
  }
</style>
```

```svelte
<!-- src/lib/blocks/HeadingBlock.svelte -->
<script lang="ts">
  import type { Block, HeadingProps } from '$lib/types/blocks';
  import { editor } from '$lib/stores/editor';

  export let block: Block;

  $: props = block.props as HeadingProps;
  $: tag = `h${props.level}` as 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

  function handleInput(e: Event) {
    const target = e.target as HTMLElement;
    editor.updateBlock(block.id, { content: target.textContent ?? '' });
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      // Create new paragraph after heading
      const parent = block.parent;
      const siblings = parent
        ? $editor.blocks.get(parent)?.children ?? []
        : $editor.rootBlocks;
      const currentIndex = siblings.indexOf(block.id);

      editor.addBlock('paragraph', parent, currentIndex + 1);
    } else if (e.key === 'Backspace' && block.content === '') {
      e.preventDefault();
      editor.deleteBlock(block.id);
    }
  }

  function handleFocus() {
    editor.setFocus(block.id);
  }
</script>

<svelte:element
  this={tag}
  class="block-heading"
  contenteditable="true"
  bind:textContent={block.content}
  on:input={handleInput}
  on:keydown={handleKeydown}
  on:focus={handleFocus}
  style:color={props.color}
  data-block-id={block.id}
/>

<style>
  .block-heading {
    outline: none;
    cursor: text;
    font-weight: 600;
  }

  .block-heading:empty::before {
    content: 'Heading';
    color: #999;
    font-weight: normal;
  }

  :global(h1.block-heading) {
    font-size: 2em;
    margin: 0.67em 0;
  }

  :global(h2.block-heading) {
    font-size: 1.5em;
    margin: 0.75em 0;
  }

  :global(h3.block-heading) {
    font-size: 1.17em;
    margin: 0.83em 0;
  }
</style>
```

```svelte
<!-- src/lib/blocks/DividerBlock.svelte -->
<script lang="ts">
  import type { Block, DividerProps } from '$lib/types/blocks';

  export let block: Block;

  $: props = block.props as DividerProps;
</script>

<hr
  class="block-divider"
  class:dashed={props.style === 'dashed'}
  class:dotted={props.style === 'dotted'}
  data-block-id={block.id}
/>

<style>
  .block-divider {
    border: none;
    border-top: 2px solid #e0e0e0;
    margin: 1.5em 0;
  }

  .block-divider.dashed {
    border-top-style: dashed;
  }

  .block-divider.dotted {
    border-top-style: dotted;
  }
</style>
```

### Step 5: Create Block Wrapper

```svelte
<!-- src/lib/components/BlockWrapper.svelte -->
<script lang="ts">
  import { getBlockComponent } from '$lib/blocks/registry';
  import type { Block } from '$lib/types/blocks';
  import { editor, focusedBlock } from '$lib/stores/editor';
  import BlockActions from './BlockActions.svelte';

  export let block: Block;

  $: component = getBlockComponent(block.type);
  $: isFocused = $focusedBlock?.id === block.id;

  let isHovered = false;
</script>

<div
  class="block-wrapper"
  class:focused={isFocused}
  class:hovered={isHovered}
  on:mouseenter={() => (isHovered = true)}
  on:mouseleave={() => (isHovered = false)}
>
  {#if isHovered || isFocused}
    <BlockActions {block} />
  {/if}

  <div class="block-content">
    <svelte:component this={component} {block} />
  </div>

  {#if block.children.length > 0}
    <div class="block-children">
      {#each block.children as childId}
        {@const child = $editor.blocks.get(childId)}
        {#if child}
          <svelte:self block={child} />
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .block-wrapper {
    position: relative;
    margin: 2px 0;
    padding-left: 32px; /* Space for actions */
  }

  .block-wrapper.focused {
    background: rgba(0, 123, 255, 0.05);
  }

  .block-content {
    position: relative;
  }

  .block-children {
    margin-left: 24px;
    border-left: 2px solid #e0e0e0;
    padding-left: 8px;
  }
</style>
```

### Step 6: Create Block Actions

```svelte
<!-- src/lib/components/BlockActions.svelte -->
<script lang="ts">
  import type { Block } from '$lib/types/blocks';
  import { editor } from '$lib/stores/editor';
  import { Icon } from 'lucide-svelte';
  import { GripVertical, MoreHorizontal } from 'lucide-svelte';

  export let block: Block;

  let menuOpen = false;

  function handleDragStart(e: DragEvent) {
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('blockId', block.id);
    }
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    const draggedBlockId = e.dataTransfer?.getData('blockId');

    if (draggedBlockId && draggedBlockId !== block.id) {
      // Determine drop position based on mouse Y position
      const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
      const y = e.clientY - rect.top;
      const position = y < rect.height / 2 ? 'before' : 'after';

      editor.moveBlock(draggedBlockId, block.id, position);
    }
  }
</script>

<div class="block-actions">
  <button
    class="action-btn drag-handle"
    draggable="true"
    on:dragstart={handleDragStart}
    on:dragover={handleDragOver}
    on:drop={handleDrop}
    title="Drag to move"
  >
    <GripVertical size={16} />
  </button>

  <button
    class="action-btn more-btn"
    on:click={() => (menuOpen = !menuOpen)}
    title="More actions"
  >
    <MoreHorizontal size={16} />
  </button>

  {#if menuOpen}
    <!-- Context menu will go here -->
    <div class="action-menu">
      <button on:click={() => editor.deleteBlock(block.id)}>
        Delete
      </button>
      <button on:click={() => editor.addBlock('paragraph', block.parent)}>
        Add below
      </button>
    </div>
  {/if}
</div>

<style>
  .block-actions {
    position: absolute;
    left: 0;
    top: 0;
    display: flex;
    gap: 2px;
    opacity: 0.6;
    transition: opacity 0.15s;
  }

  .block-actions:hover {
    opacity: 1;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 4px;
    color: #666;
    transition: background 0.15s;
  }

  .action-btn:hover {
    background: rgba(0, 0, 0, 0.05);
  }

  .drag-handle {
    cursor: grab;
  }

  .drag-handle:active {
    cursor: grabbing;
  }

  .action-menu {
    position: absolute;
    top: 100%;
    left: 0;
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    padding: 4px;
    z-index: 10;
  }

  .action-menu button {
    display: block;
    width: 100%;
    padding: 6px 12px;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    border-radius: 2px;
  }

  .action-menu button:hover {
    background: rgba(0, 0, 0, 0.05);
  }
</style>
```

### Step 7: Create Main Editor Component

```svelte
<!-- src/lib/components/Editor.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { editor } from '$lib/stores/editor';
  import BlockWrapper from './BlockWrapper.svelte';

  onMount(() => {
    editor.init();
  });
</script>

<div class="editor">
  <div class="editor-content">
    {#each $editor.rootBlocks as blockId}
      {@const block = $editor.blocks.get(blockId)}
      {#if block}
        <BlockWrapper {block} />
      {/if}
    {/each}
  </div>
</div>

<style>
  .editor {
    max-width: 800px;
    margin: 0 auto;
    padding: 40px 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 16px;
    line-height: 1.6;
  }

  .editor-content {
    min-height: 400px;
  }
</style>
```

---

## Phase 2: Slash Commands (Week 2)

### Create Slash Menu

```svelte
<!-- src/lib/components/SlashMenu.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';
  import type { BlockType } from '$lib/types/blocks';
  import { Icon } from 'lucide-svelte';
  import {
    Type,
    Heading1,
    Heading2,
    Heading3,
    List,
    CheckSquare,
    Code,
    Image,
    Minus,
  } from 'lucide-svelte';

  export let searchQuery = '';
  export let position: { x: number; y: number };

  const dispatch = createEventDispatcher();

  interface Command {
    id: string;
    label: string;
    aliases: string[];
    icon: any;
    type: BlockType;
    props?: any;
  }

  const commands: Command[] = [
    {
      id: 'paragraph',
      label: 'Paragraph',
      aliases: ['text', 'p'],
      icon: Type,
      type: 'paragraph',
    },
    {
      id: 'h1',
      label: 'Heading 1',
      aliases: ['h1', 'title'],
      icon: Heading1,
      type: 'heading',
      props: { level: 1 },
    },
    {
      id: 'h2',
      label: 'Heading 2',
      aliases: ['h2', 'subtitle'],
      icon: Heading2,
      type: 'heading',
      props: { level: 2 },
    },
    {
      id: 'h3',
      label: 'Heading 3',
      aliases: ['h3'],
      icon: Heading3,
      type: 'heading',
      props: { level: 3 },
    },
    {
      id: 'bullet',
      label: 'Bulleted List',
      aliases: ['ul', 'bullet', 'list'],
      icon: List,
      type: 'list',
      props: { type: 'bulleted' },
    },
    {
      id: 'todo',
      label: 'To-do List',
      aliases: ['todo', 'checkbox', 'check'],
      icon: CheckSquare,
      type: 'list',
      props: { type: 'todo' },
    },
    {
      id: 'code',
      label: 'Code Block',
      aliases: ['code', 'snippet'],
      icon: Code,
      type: 'code',
    },
    {
      id: 'image',
      label: 'Image',
      aliases: ['img', 'picture', 'photo'],
      icon: Image,
      type: 'image',
    },
    {
      id: 'divider',
      label: 'Divider',
      aliases: ['hr', 'line', 'separator'],
      icon: Minus,
      type: 'divider',
    },
  ];

  $: filteredCommands = commands.filter(cmd => {
    const query = searchQuery.toLowerCase();
    return (
      cmd.label.toLowerCase().includes(query) ||
      cmd.aliases.some(alias => alias.includes(query))
    );
  });

  let selectedIndex = 0;

  $: if (selectedIndex >= filteredCommands.length) {
    selectedIndex = Math.max(0, filteredCommands.length - 1);
  }

  function selectCommand(cmd: Command) {
    dispatch('select', cmd);
  }

  function handleKeydown(e: KeyboardEvent) {
    switch (e.key) {
      case 'ArrowDown':
        selectedIndex = Math.min(selectedIndex + 1, filteredCommands.length - 1);
        e.preventDefault();
        break;
      case 'ArrowUp':
        selectedIndex = Math.max(selectedIndex - 1, 0);
        e.preventDefault();
        break;
      case 'Enter':
        if (filteredCommands[selectedIndex]) {
          selectCommand(filteredCommands[selectedIndex]);
        }
        e.preventDefault();
        break;
      case 'Escape':
        dispatch('close');
        e.preventDefault();
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div
  class="slash-menu"
  style="left: {position.x}px; top: {position.y}px"
  transition:fade={{ duration: 150 }}
>
  {#if filteredCommands.length === 0}
    <div class="no-results">No commands found</div>
  {:else}
    {#each filteredCommands as cmd, i}
      <button
        class="command-item"
        class:selected={i === selectedIndex}
        on:click={() => selectCommand(cmd)}
        on:mouseenter={() => (selectedIndex = i)}
      >
        <svelte:component this={cmd.icon} size={16} />
        <span>{cmd.label}</span>
      </button>
    {/each}
  {/if}
</div>

<style>
  .slash-menu {
    position: fixed;
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    padding: 4px;
    min-width: 200px;
    max-height: 300px;
    overflow-y: auto;
    z-index: 1000;
  }

  .command-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 12px;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    border-radius: 4px;
    transition: background 0.1s;
  }

  .command-item:hover,
  .command-item.selected {
    background: rgba(0, 123, 255, 0.1);
  }

  .no-results {
    padding: 12px;
    text-align: center;
    color: #999;
  }
</style>
```

### Integrate Slash Menu into ParagraphBlock

Update `ParagraphBlock.svelte`:

```svelte
<script lang="ts">
  import type { Block, ParagraphProps } from '$lib/types/blocks';
  import { editor } from '$lib/stores/editor';
  import SlashMenu from '$lib/components/SlashMenu.svelte';

  export let block: Block;

  $: props = block.props as ParagraphProps;

  let showSlashMenu = false;
  let slashMenuPosition = { x: 0, y: 0 };
  let slashQuery = '';

  function handleInput(e: Event) {
    const target = e.target as HTMLElement;
    const content = target.textContent ?? '';

    editor.updateBlock(block.id, { content });

    // Check for slash command
    if (content.startsWith('/')) {
      slashQuery = content.slice(1);
      showSlashMenu = true;

      // Calculate position
      const rect = target.getBoundingClientRect();
      slashMenuPosition = {
        x: rect.left,
        y: rect.bottom + 4,
      };
    } else {
      showSlashMenu = false;
    }
  }

  function handleSlashCommand(e: CustomEvent) {
    const { type, props: cmdProps } = e.detail;

    // Convert this block to the selected type
    editor.convertBlock(block.id, type);

    if (cmdProps) {
      editor.updateBlock(block.id, { props: cmdProps });
    }

    // Clear the slash command
    editor.updateBlock(block.id, { content: '' });

    showSlashMenu = false;
  }

  // ... rest of the component
</script>

<div
  class="block-paragraph"
  contenteditable="true"
  bind:textContent={block.content}
  on:input={handleInput}
  on:keydown={handleKeydown}
  on:focus={handleFocus}
  style:text-align={props.textAlign}
  style:color={props.color}
  data-block-id={block.id}
/>

{#if showSlashMenu}
  <SlashMenu
    {slashQuery}
    position={slashMenuPosition}
    on:select={handleSlashCommand}
    on:close={() => (showSlashMenu = false)}
  />
{/if}
```

---

## Phase 3: Keyboard Shortcuts (Week 3)

### Create Shortcuts Action

```typescript
// src/lib/actions/shortcuts.ts

import type { Action } from 'svelte/action';

export interface ShortcutConfig {
  [key: string]: (e: KeyboardEvent) => void;
}

export const shortcuts: Action<HTMLElement, ShortcutConfig> = (node, config) => {
  function handleKeydown(e: KeyboardEvent) {
    const parts: string[] = [];

    if (e.ctrlKey || e.metaKey) parts.push(e.ctrlKey ? 'Ctrl' : 'Cmd');
    if (e.altKey) parts.push('Alt');
    if (e.shiftKey) parts.push('Shift');

    // Handle special keys
    const key = e.key === ' ' ? 'Space' : e.key;
    parts.push(key);

    const combo = parts.join('+');
    const handler = config[combo];

    if (handler) {
      handler(e);
      e.preventDefault();
    }
  }

  node.addEventListener('keydown', handleKeydown);

  return {
    update(newConfig: ShortcutConfig) {
      config = newConfig;
    },
    destroy() {
      node.removeEventListener('keydown', handleKeydown);
    },
  };
};
```

### Use Shortcuts in Editor

```svelte
<!-- Update Editor.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { editor, focusedBlock } from '$lib/stores/editor';
  import { shortcuts } from '$lib/actions/shortcuts';
  import BlockWrapper from './BlockWrapper.svelte';

  onMount(() => {
    editor.init();
  });

  const editorShortcuts = {
    'Cmd+z': () => console.log('Undo'), // Implement undo
    'Cmd+Shift+z': () => console.log('Redo'), // Implement redo
    'Cmd+b': (e: KeyboardEvent) => {
      // Toggle bold
      document.execCommand('bold');
    },
    'Cmd+i': (e: KeyboardEvent) => {
      // Toggle italic
      document.execCommand('italic');
    },
    'Cmd+u': (e: KeyboardEvent) => {
      // Toggle underline
      document.execCommand('underline');
    },
  };
</script>

<div class="editor" use:shortcuts={editorShortcuts}>
  <!-- ... rest of component -->
</div>
```

---

## Next Steps

1. **Implement List Blocks** with nested items
2. **Add Code Block** with syntax highlighting (using Shiki or Prism)
3. **Create Image Block** with upload/URL input
4. **Build Formatting Toolbar** (text selection popup)
5. **Add Undo/Redo** (history management in store)
6. **Implement Sidebar** with page tree
7. **Add Search/Command Palette** (Cmd+K)

---

## Testing Plan

```typescript
// tests/editor.test.ts

import { describe, it, expect } from 'vitest';
import { editor } from '$lib/stores/editor';

describe('Editor Store', () => {
  it('should initialize with one paragraph block', () => {
    editor.init();
    const state = get(editor);

    expect(state.blocks.size).toBe(1);
    expect(state.rootBlocks.length).toBe(1);
    expect(state.blocks.get(state.rootBlocks[0])?.type).toBe('paragraph');
  });

  it('should add a new block', () => {
    editor.init();
    const newBlockId = editor.addBlock('heading', null);
    const state = get(editor);

    expect(state.blocks.has(newBlockId)).toBe(true);
    expect(state.rootBlocks).toContain(newBlockId);
  });

  it('should delete a block', () => {
    editor.init();
    const state = get(editor);
    const firstBlockId = state.rootBlocks[0];

    editor.deleteBlock(firstBlockId);
    const newState = get(editor);

    expect(newState.blocks.has(firstBlockId)).toBe(false);
  });
});
```

---

This gives you a complete foundation to build on. Start with Phase 1, get it working, then move to Phase 2.
