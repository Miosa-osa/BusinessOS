# AFFiNE Block Editor Architecture Analysis

## Executive Summary

This document analyzes AFFiNE's block-based editor architecture to extract patterns and approaches for implementing a similar system in Svelte for BusinessOS.

---

## 1. Block System Architecture

### 1.1 Core Block Structure

AFFiNE uses a hierarchical block system built on BlockSuite, their custom framework:

```typescript
// Conceptual Block Structure
interface Block {
  id: string;                    // Unique block identifier
  type: string;                  // Block type (paragraph, heading, list, etc.)
  parent: string | null;         // Parent block ID
  children: string[];            // Child block IDs
  props: Record<string, any>;    // Block-specific properties
  content: Y.XmlText | Y.Map;    // CRDT-based content (Yjs)
}
```

**Key Insights:**
- Blocks are **immutable data structures** with explicit IDs
- Parent-child relationships form a **tree structure**
- Content uses **Yjs CRDTs** for real-time collaboration
- Block state is separate from block rendering

### 1.2 Block Types

AFFiNE implements various block types, each with specific behaviors:

| Block Type | Purpose | Key Features |
|------------|---------|--------------|
| `paragraph` | Basic text content | Rich text, inline formatting |
| `heading` | Section headers | H1-H6 levels |
| `list` | Ordered/unordered lists | Nested items, checkboxes |
| `code` | Code snippets | Syntax highlighting, language selection |
| `database` | Table/board views | Rows, columns, filters, sorts |
| `image` | Media content | Upload, resize, captions |
| `divider` | Section separator | Visual break |
| `page` | Nested pages | Backlinks, references |

**Pattern to Adopt:**
```typescript
// Svelte Block Registry Pattern
export const blockRegistry = {
  paragraph: ParagraphBlock,
  heading: HeadingBlock,
  list: ListBlock,
  code: CodeBlock,
  database: DatabaseBlock,
  image: ImageBlock,
  divider: DividerBlock,
  page: PageBlock,
};

// Dynamic component loading
export function getBlockComponent(type: string) {
  return blockRegistry[type] ?? ParagraphBlock;
}
```

### 1.3 Block Properties Storage

AFFiNE stores block properties in a **flat key-value structure**:

```typescript
// Example: Heading Block Props
{
  type: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6',
  color?: string,
  backgroundColor?: string,
  collapsed?: boolean,
}

// Example: List Block Props
{
  type: 'bulleted' | 'numbered' | 'todo',
  checked?: boolean,  // For todo lists
  collapsed?: boolean,
}

// Example: Code Block Props
{
  language: string,
  wrap: boolean,
  caption?: string,
}
```

**Storage Pattern:**
- Properties are stored in a **Y.Map** (Yjs collaborative map)
- Updates trigger reactive re-renders
- Type-safe property schemas per block type

### 1.4 Block Interactions

#### Selection System

```typescript
// Selection State
interface BlockSelection {
  type: 'block' | 'text' | 'range';
  blockId?: string;
  startOffset?: number;
  endOffset?: number;
  blocks?: string[];  // Multi-block selection
}

// Selection Manager
class SelectionManager {
  private selection: Writable<BlockSelection | null>;

  selectBlock(blockId: string) { /* ... */ }
  selectText(blockId: string, start: number, end: number) { /* ... */ }
  selectRange(startBlockId: string, endBlockId: string) { /* ... */ }
  clearSelection() { /* ... */ }
}
```

#### Drag and Drop

```typescript
// Drag-Drop Pattern
interface DragState {
  draggingBlockId: string;
  dropTargetId: string | null;
  dropPosition: 'before' | 'after' | 'inside';
}

// Handlers
function handleDragStart(blockId: string) {
  dragState.set({ draggingBlockId: blockId, /* ... */ });
}

function handleDrop(targetId: string, position: string) {
  const block = getBlock(dragState.draggingBlockId);
  moveBlock(block, targetId, position);
}
```

#### Block Commands

AFFiNE uses a **command pattern** for block operations:

```typescript
// Command Registry
const commands = {
  'block:indent': (blockId: string) => { /* ... */ },
  'block:outdent': (blockId: string) => { /* ... */ },
  'block:delete': (blockId: string) => { /* ... */ },
  'block:duplicate': (blockId: string) => { /* ... */ },
  'block:convert': (blockId: string, newType: string) => { /* ... */ },
  'block:moveUp': (blockId: string) => { /* ... */ },
  'block:moveDown': (blockId: string) => { /* ... */ },
};

// Execute command
function executeCommand(commandId: string, ...args: any[]) {
  const command = commands[commandId];
  if (command) {
    command(...args);
  }
}
```

---

## 2. Editor Features

### 2.1 Cover/Header System

AFFiNE implements a page cover and header icon system:

```typescript
// Page Metadata
interface PageMeta {
  title: string;
  icon?: string;        // Emoji or custom icon
  cover?: {
    type: 'color' | 'gradient' | 'image';
    value: string;      // Color hex, gradient ID, or image URL
    position?: number;  // Vertical position (0-1)
  };
}

// Cover Component Structure
<script lang="ts">
  export let cover: PageMeta['cover'];
  export let editable = false;

  let coverPosition = cover?.position ?? 0.5;

  function handleCoverChange(newCover: PageMeta['cover']) {
    // Update cover
  }
</script>

<div class="page-cover" class:editable>
  {#if cover?.type === 'color'}
    <div class="cover-color" style="background: {cover.value}"></div>
  {:else if cover?.type === 'gradient'}
    <div class="cover-gradient" style="background: {getGradient(cover.value)}"></div>
  {:else if cover?.type === 'image'}
    <img
      src={cover.value}
      alt="Cover"
      style="object-position: 0 {coverPosition * 100}%"
    />
  {/if}

  {#if editable}
    <CoverToolbar on:change={handleCoverChange} />
  {/if}
</div>
```

**Pattern to Adopt:**
- Separate cover/icon from block content
- Store as page-level metadata
- Provide inline editing UI on hover
- Support multiple cover types (solid, gradient, image, unsplash)

### 2.2 Toolbar Implementation

AFFiNE uses multiple toolbar types:

#### Inline Formatting Toolbar (Text Selection)

```svelte
<!-- FormattingToolbar.svelte -->
<script lang="ts">
  import { Portal } from 'svelte-portal';
  import { fade } from 'svelte/transition';

  export let selection: TextSelection;
  export let position: { x: number; y: number };

  let formats = ['bold', 'italic', 'underline', 'strikethrough', 'code', 'link'];

  function applyFormat(format: string) {
    document.execCommand(format);
  }
</script>

<Portal target="body">
  <div
    class="formatting-toolbar"
    style="left: {position.x}px; top: {position.y}px"
    transition:fade={{ duration: 150 }}
  >
    {#each formats as format}
      <button
        class="format-btn"
        class:active={isFormatActive(format)}
        on:click={() => applyFormat(format)}
      >
        <Icon name={format} />
      </button>
    {/each}
  </div>
</Portal>
```

#### Block Action Menu (Block Hover)

```svelte
<!-- BlockActions.svelte -->
<script lang="ts">
  export let blockId: string;
  export let showDragHandle = true;

  let menuOpen = false;
  let menuPosition = { x: 0, y: 0 };

  const actions = [
    { id: 'delete', label: 'Delete', icon: 'trash', shortcut: 'Del' },
    { id: 'duplicate', label: 'Duplicate', icon: 'copy', shortcut: 'Cmd+D' },
    { id: 'convert', label: 'Turn into', icon: 'wand', shortcut: '/' },
    { id: 'moveUp', label: 'Move up', icon: 'arrow-up', shortcut: 'Cmd+Shift+↑' },
    { id: 'moveDown', label: 'Move down', icon: 'arrow-down', shortcut: 'Cmd+Shift+↓' },
  ];
</script>

<div class="block-actions">
  {#if showDragHandle}
    <button
      class="drag-handle"
      draggable="true"
      on:dragstart={() => handleDragStart(blockId)}
    >
      <Icon name="grip-vertical" />
    </button>
  {/if}

  <button
    class="more-actions"
    on:click={() => menuOpen = !menuOpen}
  >
    <Icon name="more-horizontal" />
  </button>

  {#if menuOpen}
    <ContextMenu
      items={actions}
      position={menuPosition}
      on:select={(e) => executeCommand(`block:${e.detail.id}`, blockId)}
      on:close={() => menuOpen = false}
    />
  {/if}
</div>
```

### 2.3 Slash Commands

AFFiNE's slash command system is one of its best features:

```typescript
// Slash Command Registry
interface SlashCommand {
  id: string;
  label: string;
  aliases: string[];
  category: 'basic' | 'media' | 'database' | 'advanced';
  icon: string;
  keywords: string[];
  action: (blockId: string) => void;
}

const slashCommands: SlashCommand[] = [
  {
    id: 'heading1',
    label: 'Heading 1',
    aliases: ['h1', 'title'],
    category: 'basic',
    icon: 'heading-1',
    keywords: ['heading', 'title', 'h1'],
    action: (blockId) => convertBlock(blockId, 'heading', { type: 'h1' }),
  },
  {
    id: 'todo',
    label: 'To-do list',
    aliases: ['checkbox', 'check', 'task'],
    category: 'basic',
    icon: 'check-square',
    keywords: ['todo', 'task', 'checkbox', 'list'],
    action: (blockId) => convertBlock(blockId, 'list', { type: 'todo' }),
  },
  // ... more commands
];

// Slash Menu Component
<script lang="ts">
  export let searchQuery = '';
  export let blockId: string;

  $: filteredCommands = slashCommands.filter(cmd => {
    const query = searchQuery.toLowerCase();
    return (
      cmd.label.toLowerCase().includes(query) ||
      cmd.aliases.some(alias => alias.includes(query)) ||
      cmd.keywords.some(keyword => keyword.includes(query))
    );
  });

  $: groupedCommands = groupBy(filteredCommands, 'category');

  let selectedIndex = 0;

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      selectedIndex = Math.min(selectedIndex + 1, filteredCommands.length - 1);
      e.preventDefault();
    } else if (e.key === 'ArrowUp') {
      selectedIndex = Math.max(selectedIndex - 1, 0);
      e.preventDefault();
    } else if (e.key === 'Enter') {
      executeCommand(filteredCommands[selectedIndex], blockId);
      e.preventDefault();
    }
  }
</script>

<div class="slash-menu" on:keydown={handleKeydown}>
  {#each Object.entries(groupedCommands) as [category, commands]}
    <div class="command-group">
      <div class="group-label">{category}</div>
      {#each commands as command, i}
        <button
          class="command-item"
          class:selected={i === selectedIndex}
          on:click={() => executeCommand(command, blockId)}
        >
          <Icon name={command.icon} />
          <span>{command.label}</span>
          {#if command.aliases.length}
            <span class="aliases">{command.aliases.join(', ')}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/each}
</div>
```

**Pattern to Adopt:**
- Fuzzy search across labels, aliases, and keywords
- Group commands by category
- Keyboard navigation (↑/↓ arrows, Enter)
- Show command preview on hover
- Support custom command plugins

### 2.4 Format Toolbar

The main toolbar at the top:

```svelte
<!-- EditorToolbar.svelte -->
<script lang="ts">
  import { editorState } from './stores';

  const toolGroups = [
    {
      label: 'Basic',
      tools: [
        { id: 'bold', icon: 'bold', shortcut: 'Cmd+B' },
        { id: 'italic', icon: 'italic', shortcut: 'Cmd+I' },
        { id: 'underline', icon: 'underline', shortcut: 'Cmd+U' },
        { id: 'strikethrough', icon: 'strikethrough', shortcut: 'Cmd+Shift+X' },
      ],
    },
    {
      label: 'Insert',
      tools: [
        { id: 'link', icon: 'link', shortcut: 'Cmd+K' },
        { id: 'image', icon: 'image' },
        { id: 'code', icon: 'code', shortcut: 'Cmd+E' },
      ],
    },
    {
      label: 'Blocks',
      tools: [
        { id: 'heading', icon: 'heading', hasDropdown: true },
        { id: 'list', icon: 'list', hasDropdown: true },
        { id: 'database', icon: 'table' },
      ],
    },
  ];
</script>

<div class="editor-toolbar">
  {#each toolGroups as group}
    <div class="tool-group">
      {#each group.tools as tool}
        <button
          class="tool-btn"
          class:active={$editorState.activeFormats.includes(tool.id)}
          on:click={() => executeTool(tool.id)}
          title="{tool.id} ({tool.shortcut})"
        >
          <Icon name={tool.icon} />
          {#if tool.hasDropdown}
            <Icon name="chevron-down" size={12} />
          {/if}
        </button>
      {/each}
    </div>
  {/each}

  <div class="tool-group ml-auto">
    <button class="tool-btn" title="Share">
      <Icon name="share" />
    </button>
    <button class="tool-btn" title="More">
      <Icon name="more-horizontal" />
    </button>
  </div>
</div>
```

### 2.5 Block Menu/Actions

Right-click context menu:

```svelte
<!-- BlockContextMenu.svelte -->
<script lang="ts">
  export let blockId: string;
  export let position: { x: number; y: number };
  export let onClose: () => void;

  const menuItems = [
    { id: 'cut', label: 'Cut', icon: 'scissors', shortcut: 'Cmd+X' },
    { id: 'copy', label: 'Copy', icon: 'copy', shortcut: 'Cmd+C' },
    { id: 'paste', label: 'Paste', icon: 'clipboard', shortcut: 'Cmd+V' },
    { divider: true },
    { id: 'duplicate', label: 'Duplicate', icon: 'copy', shortcut: 'Cmd+D' },
    { id: 'delete', label: 'Delete', icon: 'trash', shortcut: 'Del', danger: true },
    { divider: true },
    { id: 'convert', label: 'Turn into', icon: 'wand', hasSubmenu: true },
    { id: 'color', label: 'Color', icon: 'palette', hasSubmenu: true },
    { divider: true },
    { id: 'copyLink', label: 'Copy link to block', icon: 'link' },
  ];
</script>

<div
  class="context-menu"
  style="left: {position.x}px; top: {position.y}px"
  use:clickOutside={onClose}
>
  {#each menuItems as item}
    {#if item.divider}
      <div class="menu-divider"></div>
    {:else}
      <button
        class="menu-item"
        class:danger={item.danger}
        on:click={() => handleMenuAction(item.id)}
      >
        <Icon name={item.icon} />
        <span class="label">{item.label}</span>
        {#if item.shortcut}
          <span class="shortcut">{item.shortcut}</span>
        {/if}
        {#if item.hasSubmenu}
          <Icon name="chevron-right" size={12} />
        {/if}
      </button>
    {/if}
  {/each}
</div>
```

---

## 3. Sidebar Structure

### 3.1 Tree View Implementation

AFFiNE's sidebar uses a recursive tree component:

```svelte
<!-- SidebarTree.svelte -->
<script lang="ts">
  import { writable } from 'svelte/store';

  export let pages: Page[];

  interface TreeNode {
    id: string;
    title: string;
    icon?: string;
    children?: TreeNode[];
    collapsed?: boolean;
  }

  const expandedNodes = writable<Set<string>>(new Set());

  function toggleNode(nodeId: string) {
    expandedNodes.update(set => {
      if (set.has(nodeId)) {
        set.delete(nodeId);
      } else {
        set.add(nodeId);
      }
      return set;
    });
  }
</script>

<nav class="sidebar-tree">
  {#each pages as page}
    <TreeNode
      node={page}
      level={0}
      expanded={$expandedNodes.has(page.id)}
      on:toggle={() => toggleNode(page.id)}
    />
  {/each}
</nav>

<!-- TreeNode.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let node: TreeNode;
  export let level = 0;
  export let expanded = false;

  const dispatch = createEventDispatcher();

  let isHovered = false;
  let isDragging = false;
</script>

<div
  class="tree-node"
  style="padding-left: {level * 20}px"
  class:expanded
  class:hovered={isHovered}
  draggable="true"
  on:mouseenter={() => isHovered = true}
  on:mouseleave={() => isHovered = false}
  on:dragstart={() => isDragging = true}
  on:dragend={() => isDragging = false}
>
  <button
    class="node-toggle"
    class:has-children={node.children?.length}
    on:click={() => dispatch('toggle')}
  >
    {#if node.children?.length}
      <Icon name={expanded ? 'chevron-down' : 'chevron-right'} size={14} />
    {/if}
  </button>

  <a href="/page/{node.id}" class="node-link">
    {#if node.icon}
      <span class="node-icon">{node.icon}</span>
    {:else}
      <Icon name="file-text" size={16} />
    {/if}
    <span class="node-title">{node.title}</span>
  </a>

  {#if isHovered}
    <div class="node-actions">
      <button class="action-btn" title="Add page">
        <Icon name="plus" size={14} />
      </button>
      <button class="action-btn" title="More">
        <Icon name="more-horizontal" size={14} />
      </button>
    </div>
  {/if}
</div>

{#if expanded && node.children?.length}
  <div class="node-children">
    {#each node.children as child}
      <svelte:self
        node={child}
        level={level + 1}
        expanded={$expandedNodes.has(child.id)}
        on:toggle
      />
    {/each}
  </div>
{/if}
```

### 3.2 Navigation Patterns

```svelte
<!-- Sidebar.svelte -->
<script lang="ts">
  import { page } from '$app/stores';
  import { sidebar } from './stores';

  const sections = [
    {
      id: 'quick-access',
      label: 'Quick Access',
      items: [
        { id: 'all-pages', label: 'All Pages', icon: 'files', href: '/pages' },
        { id: 'journal', label: 'Journal', icon: 'calendar', href: '/journal' },
        { id: 'favorites', label: 'Favorites', icon: 'star', href: '/favorites' },
      ],
    },
    {
      id: 'workspaces',
      label: 'Workspaces',
      collapsible: true,
      items: [], // Dynamic workspace list
    },
    {
      id: 'pages',
      label: 'Pages',
      collapsible: true,
      component: SidebarTree,
    },
  ];
</script>

<aside class="sidebar" class:collapsed={$sidebar.collapsed}>
  <div class="sidebar-header">
    <button class="workspace-switcher">
      <Icon name="briefcase" />
      <span>My Workspace</span>
      <Icon name="chevron-down" size={12} />
    </button>
  </div>

  {#each sections as section}
    <div class="sidebar-section">
      <div class="section-header">
        <h3>{section.label}</h3>
        {#if section.collapsible}
          <button class="collapse-btn">
            <Icon name="chevron-down" size={14} />
          </button>
        {/if}
      </div>

      {#if section.component}
        <svelte:component this={section.component} />
      {:else}
        <ul class="section-items">
          {#each section.items as item}
            <li>
              <a
                href={item.href}
                class="item-link"
                class:active={$page.url.pathname === item.href}
              >
                <Icon name={item.icon} size={16} />
                <span>{item.label}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {/each}

  <div class="sidebar-footer">
    <button class="footer-btn">
      <Icon name="settings" />
      Settings
    </button>
  </div>
</aside>
```

### 3.3 Quick Actions

```svelte
<!-- QuickActions.svelte -->
<script lang="ts">
  let searchQuery = '';
  let commandPaletteOpen = false;

  function handleKeydown(e: KeyboardEvent) {
    // Cmd+K to open command palette
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      commandPaletteOpen = true;
      e.preventDefault();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="quick-actions">
  <button
    class="search-btn"
    on:click={() => commandPaletteOpen = true}
  >
    <Icon name="search" />
    <span>Search...</span>
    <kbd>Cmd+K</kbd>
  </button>

  <button class="new-page-btn" title="New Page">
    <Icon name="plus" />
  </button>
</div>

{#if commandPaletteOpen}
  <CommandPalette on:close={() => commandPaletteOpen = false} />
{/if}
```

---

## 4. Key Patterns to Extract

### 4.1 Component Composition Patterns

**1. Compound Components Pattern**

```svelte
<!-- Editor.svelte (Container) -->
<script lang="ts">
  import { setContext } from 'svelte';
  import { writable } from 'svelte/store';

  const editorContext = writable({
    selection: null,
    activeFormats: [],
    blocks: new Map(),
  });

  setContext('editor', editorContext);
</script>

<div class="editor">
  <EditorToolbar />
  <EditorContent />
  <EditorFooter />
</div>

<!-- EditorContent.svelte (Child) -->
<script lang="ts">
  import { getContext } from 'svelte';

  const editor = getContext('editor');
</script>

<div class="editor-content">
  {#each $editor.blocks as block}
    <BlockComponent {block} />
  {/each}
</div>
```

**2. Render Props Pattern (Slots)**

```svelte
<!-- BlockWrapper.svelte -->
<script lang="ts">
  export let block;
  let isHovered = false;
</script>

<div
  class="block-wrapper"
  class:hovered={isHovered}
  on:mouseenter={() => isHovered = true}
  on:mouseleave={() => isHovered = false}
>
  <!-- Pass state to slot -->
  <slot {isHovered} {block} />

  {#if isHovered}
    <slot name="actions" {block} />
  {/if}
</div>

<!-- Usage -->
<BlockWrapper {block} let:isHovered let:block>
  <BlockContent {block} />

  <svelte:fragment slot="actions" let:block>
    <BlockActions blockId={block.id} />
  </svelte:fragment>
</BlockWrapper>
```

**3. Higher-Order Component Pattern (Actions)**

```typescript
// withDragDrop.ts
export function dragDrop(node: HTMLElement, options: DragDropOptions) {
  let isDragging = false;

  function handleDragStart(e: DragEvent) {
    isDragging = true;
    e.dataTransfer?.setData('blockId', options.id);
    options.onDragStart?.(e);
  }

  function handleDragEnd(e: DragEvent) {
    isDragging = false;
    options.onDragEnd?.(e);
  }

  node.addEventListener('dragstart', handleDragStart);
  node.addEventListener('dragend', handleDragEnd);

  return {
    destroy() {
      node.removeEventListener('dragstart', handleDragStart);
      node.removeEventListener('dragend', handleDragEnd);
    },
  };
}

// Usage
<div use:dragDrop={{ id: block.id, onDragStart, onDragEnd }}>
  Block content
</div>
```

### 4.2 State Management Approach

**1. Centralized Editor Store**

```typescript
// stores/editor.ts
import { writable, derived } from 'svelte/store';

interface Block {
  id: string;
  type: string;
  parent: string | null;
  children: string[];
  props: Record<string, any>;
  content: string;
}

interface EditorState {
  blocks: Map<string, Block>;
  selection: Selection | null;
  activeFormats: string[];
  history: {
    past: EditorState[];
    future: EditorState[];
  };
}

function createEditorStore() {
  const { subscribe, update, set } = writable<EditorState>({
    blocks: new Map(),
    selection: null,
    activeFormats: [],
    history: { past: [], future: [] },
  });

  return {
    subscribe,

    // Block operations
    addBlock(block: Block) {
      update(state => {
        state.blocks.set(block.id, block);
        return state;
      });
    },

    updateBlock(id: string, updates: Partial<Block>) {
      update(state => {
        const block = state.blocks.get(id);
        if (block) {
          state.blocks.set(id, { ...block, ...updates });
        }
        return state;
      });
    },

    deleteBlock(id: string) {
      update(state => {
        state.blocks.delete(id);
        return state;
      });
    },

    // Selection
    setSelection(selection: Selection | null) {
      update(state => ({ ...state, selection }));
    },

    // Undo/Redo
    undo() {
      update(state => {
        const previous = state.history.past.pop();
        if (previous) {
          state.history.future.push({ ...state });
          return previous;
        }
        return state;
      });
    },

    redo() {
      update(state => {
        const next = state.history.future.pop();
        if (next) {
          state.history.past.push({ ...state });
          return next;
        }
        return state;
      });
    },
  };
}

export const editor = createEditorStore();

// Derived stores
export const selectedBlock = derived(
  editor,
  $editor => {
    if ($editor.selection?.type === 'block') {
      return $editor.blocks.get($editor.selection.blockId);
    }
    return null;
  }
);

export const blockTree = derived(
  editor,
  $editor => {
    // Build hierarchical tree from flat blocks map
    const roots: Block[] = [];
    // ... tree building logic
    return roots;
  }
);
```

**2. Local Component State**

```svelte
<script lang="ts">
  import { editor } from './stores/editor';

  export let blockId: string;

  // Global state
  $: block = $editor.blocks.get(blockId);

  // Local UI state (doesn't need to be in store)
  let isHovered = false;
  let isEditing = false;
  let menuOpen = false;

  // Local derived state
  $: hasChildren = block?.children.length > 0;
  $: isSelected = $editor.selection?.blockId === blockId;
</script>
```

### 4.3 Event Handling Patterns

**1. Keyboard Shortcuts System**

```typescript
// keyboard.ts
import type { Action } from 'svelte/action';

interface ShortcutConfig {
  [key: string]: (e: KeyboardEvent) => void;
}

export const shortcuts: Action<HTMLElement, ShortcutConfig> = (
  node,
  config
) => {
  function handleKeydown(e: KeyboardEvent) {
    const key = [
      e.ctrlKey && 'Ctrl',
      e.metaKey && 'Cmd',
      e.altKey && 'Alt',
      e.shiftKey && 'Shift',
      e.key,
    ]
      .filter(Boolean)
      .join('+');

    const handler = config[key];
    if (handler) {
      handler(e);
      e.preventDefault();
    }
  }

  node.addEventListener('keydown', handleKeydown);

  return {
    update(newConfig) {
      config = newConfig;
    },
    destroy() {
      node.removeEventListener('keydown', handleKeydown);
    },
  };
};

// Usage
<div use:shortcuts={{
  'Cmd+b': () => toggleFormat('bold'),
  'Cmd+i': () => toggleFormat('italic'),
  'Cmd+k': () => openLinkDialog(),
  'Enter': (e) => handleEnter(e),
  'Backspace': (e) => handleBackspace(e),
}}>
  Editor content
</div>
```

**2. Event Delegation Pattern**

```svelte
<script lang="ts">
  function handleEditorClick(e: MouseEvent) {
    const target = e.target as HTMLElement;

    // Handle block click
    const blockEl = target.closest('[data-block-id]');
    if (blockEl) {
      const blockId = blockEl.getAttribute('data-block-id');
      selectBlock(blockId);
      return;
    }

    // Handle link click
    const linkEl = target.closest('a[data-link]');
    if (linkEl) {
      handleLinkClick(linkEl);
      return;
    }
  }
</script>

<div
  class="editor"
  on:click={handleEditorClick}
  on:keydown={handleEditorKeydown}
>
  {#each blocks as block}
    <div data-block-id={block.id}>
      {block.content}
    </div>
  {/each}
</div>
```

**3. Custom Events**

```svelte
<!-- BlockComponent.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  function handleBlockUpdate(updates: Partial<Block>) {
    dispatch('update', { blockId: block.id, updates });
  }

  function handleBlockDelete() {
    dispatch('delete', { blockId: block.id });
  }
</script>

<!-- Parent -->
<BlockComponent
  on:update={(e) => editor.updateBlock(e.detail.blockId, e.detail.updates)}
  on:delete={(e) => editor.deleteBlock(e.detail.blockId)}
/>
```

### 4.4 Keyboard Shortcuts System

**Global Shortcuts Registry**

```typescript
// shortcuts.ts
export interface Shortcut {
  key: string;
  description: string;
  handler: () => void;
  contexts?: string[]; // Where this shortcut applies
}

export const globalShortcuts: Record<string, Shortcut> = {
  'Cmd+b': {
    key: 'Cmd+b',
    description: 'Bold',
    handler: () => toggleFormat('bold'),
    contexts: ['editor'],
  },
  'Cmd+i': {
    key: 'Cmd+i',
    description: 'Italic',
    handler: () => toggleFormat('italic'),
    contexts: ['editor'],
  },
  'Cmd+k': {
    key: 'Cmd+k',
    description: 'Insert link',
    handler: () => insertLink(),
    contexts: ['editor'],
  },
  'Cmd+/': {
    key: 'Cmd+/',
    description: 'Show command palette',
    handler: () => openCommandPalette(),
    contexts: ['global'],
  },
  'Cmd+z': {
    key: 'Cmd+z',
    description: 'Undo',
    handler: () => editor.undo(),
    contexts: ['editor'],
  },
  'Cmd+Shift+z': {
    key: 'Cmd+Shift+z',
    description: 'Redo',
    handler: () => editor.redo(),
    contexts: ['editor'],
  },
};

// Shortcut Help Dialog
<script lang="ts">
  import { globalShortcuts } from './shortcuts';

  $: shortcutGroups = Object.values(globalShortcuts).reduce((acc, shortcut) => {
    const context = shortcut.contexts?.[0] ?? 'global';
    if (!acc[context]) acc[context] = [];
    acc[context].push(shortcut);
    return acc;
  }, {} as Record<string, Shortcut[]>);
</script>

<Dialog title="Keyboard Shortcuts">
  {#each Object.entries(shortcutGroups) as [context, shortcuts]}
    <div class="shortcut-section">
      <h3>{context}</h3>
      {#each shortcuts as shortcut}
        <div class="shortcut-row">
          <span class="description">{shortcut.description}</span>
          <kbd>{shortcut.key}</kbd>
        </div>
      {/each}
    </div>
  {/each}
</Dialog>
```

---

## 5. Recommended Architecture for Svelte Implementation

### 5.1 Project Structure

```
src/
├── lib/
│   ├── components/
│   │   ├── editor/
│   │   │   ├── Editor.svelte              # Main editor container
│   │   │   ├── EditorToolbar.svelte       # Top toolbar
│   │   │   ├── EditorContent.svelte       # Content area
│   │   │   ├── BlockWrapper.svelte        # Wrapper for all blocks
│   │   │   └── blocks/
│   │   │       ├── ParagraphBlock.svelte
│   │   │       ├── HeadingBlock.svelte
│   │   │       ├── ListBlock.svelte
│   │   │       ├── CodeBlock.svelte
│   │   │       ├── DatabaseBlock.svelte
│   │   │       └── ...
│   │   ├── toolbar/
│   │   │   ├── FormattingToolbar.svelte   # Inline text toolbar
│   │   │   ├── BlockActions.svelte        # Block hover actions
│   │   │   └── SlashMenu.svelte           # / command menu
│   │   ├── sidebar/
│   │   │   ├── Sidebar.svelte
│   │   │   ├── SidebarTree.svelte
│   │   │   ├── TreeNode.svelte
│   │   │   └── QuickActions.svelte
│   │   └── ui/
│   │       ├── ContextMenu.svelte
│   │       ├── CommandPalette.svelte
│   │       └── ...
│   ├── stores/
│   │   ├── editor.ts                      # Main editor store
│   │   ├── selection.ts                   # Selection management
│   │   ├── sidebar.ts                     # Sidebar state
│   │   └── commands.ts                    # Command registry
│   ├── actions/
│   │   ├── shortcuts.ts                   # Keyboard shortcuts
│   │   ├── dragDrop.ts                    # Drag & drop
│   │   └── clickOutside.ts                # Click outside detector
│   ├── utils/
│   │   ├── blocks.ts                      # Block utilities
│   │   ├── selection.ts                   # Selection utilities
│   │   └── commands.ts                    # Command execution
│   └── types/
│       ├── blocks.ts                      # Block type definitions
│       ├── editor.ts                      # Editor types
│       └── commands.ts                    # Command types
└── routes/
    └── editor/
        └── [pageId]/
            └── +page.svelte
```

### 5.2 Core Architecture Decisions

| Aspect | Recommendation | Rationale |
|--------|---------------|-----------|
| State Management | Svelte stores + Context API | Native, performant, simple |
| Block Storage | Flat Map with parent/child refs | Easy lookups, flexible hierarchy |
| Real-time Sync | Yjs CRDTs (optional) | If collaboration needed |
| Content Editing | ContentEditable + execCommand | Standard, well-supported |
| Drag & Drop | HTML5 Drag & Drop API | Native browser support |
| Keyboard Shortcuts | Custom action + registry | Flexible, declarative |
| Command System | Registry pattern | Extensible, testable |
| Undo/Redo | History stack in store | Simple, effective |

### 5.3 Implementation Phases

**Phase 1: Foundation**
- [ ] Basic block structure (paragraph only)
- [ ] Editor store with CRUD operations
- [ ] Block wrapper with hover state
- [ ] Simple toolbar (bold, italic, underline)
- [ ] Keyboard shortcuts system

**Phase 2: Core Blocks**
- [ ] Heading blocks (H1-H6)
- [ ] List blocks (bulleted, numbered, todo)
- [ ] Code blocks with syntax highlighting
- [ ] Divider block
- [ ] Image block

**Phase 3: Advanced Features**
- [ ] Slash command menu
- [ ] Block actions menu (hover + right-click)
- [ ] Drag & drop reordering
- [ ] Formatting toolbar (selection)
- [ ] Undo/Redo

**Phase 4: Sidebar & Navigation**
- [ ] Sidebar with tree view
- [ ] Page hierarchy
- [ ] Quick actions
- [ ] Command palette

**Phase 5: Polish**
- [ ] Cover/header system
- [ ] Animations & transitions
- [ ] Keyboard navigation
- [ ] Accessibility (ARIA)
- [ ] Performance optimization

### 5.4 Code Examples for Key Features

#### Block Registry & Dynamic Components

```typescript
// lib/blocks/registry.ts
import ParagraphBlock from './ParagraphBlock.svelte';
import HeadingBlock from './HeadingBlock.svelte';
import ListBlock from './ListBlock.svelte';
import CodeBlock from './CodeBlock.svelte';

export const blockComponents = {
  paragraph: ParagraphBlock,
  heading: HeadingBlock,
  list: ListBlock,
  code: CodeBlock,
} as const;

export type BlockType = keyof typeof blockComponents;

export function getBlockComponent(type: BlockType) {
  return blockComponents[type] ?? ParagraphBlock;
}
```

```svelte
<!-- BlockRenderer.svelte -->
<script lang="ts">
  import { getBlockComponent } from './registry';
  import type { Block } from './types';

  export let block: Block;

  $: component = getBlockComponent(block.type);
</script>

<svelte:component this={component} {block} />
```

#### Selection Management

```typescript
// stores/selection.ts
import { writable, derived } from 'svelte/store';

export type SelectionType = 'none' | 'block' | 'text' | 'range';

export interface Selection {
  type: SelectionType;
  blockId?: string;
  anchorOffset?: number;
  focusOffset?: number;
  anchorBlockId?: string;
  focusBlockId?: string;
}

function createSelectionStore() {
  const { subscribe, set, update } = writable<Selection>({
    type: 'none',
  });

  return {
    subscribe,

    selectBlock(blockId: string) {
      set({ type: 'block', blockId });
    },

    selectText(blockId: string, anchorOffset: number, focusOffset: number) {
      set({
        type: 'text',
        blockId,
        anchorOffset,
        focusOffset,
      });
    },

    selectRange(anchorBlockId: string, focusBlockId: string) {
      set({
        type: 'range',
        anchorBlockId,
        focusBlockId,
      });
    },

    clear() {
      set({ type: 'none' });
    },
  };
}

export const selection = createSelectionStore();
```

#### Command Palette

```svelte
<!-- CommandPalette.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { commands, type Command } from '$lib/stores/commands';
  import { Portal } from 'svelte-portal';

  export let onClose: () => void;

  let searchQuery = '';
  let selectedIndex = 0;
  let inputEl: HTMLInputElement;

  $: filteredCommands = $commands.filter(cmd =>
    cmd.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
    cmd.keywords.some(k => k.includes(searchQuery.toLowerCase()))
  );

  $: if (selectedIndex >= filteredCommands.length) {
    selectedIndex = Math.max(0, filteredCommands.length - 1);
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
          executeCommand(filteredCommands[selectedIndex]);
          onClose();
        }
        e.preventDefault();
        break;
      case 'Escape':
        onClose();
        e.preventDefault();
        break;
    }
  }

  function executeCommand(cmd: Command) {
    cmd.action();
  }

  onMount(() => {
    inputEl?.focus();
  });
</script>

<Portal target="body">
  <div class="command-palette-backdrop" transition:fade={{ duration: 150 }} on:click={onClose}>
    <div
      class="command-palette"
      transition:fly={{ y: -20, duration: 200 }}
      on:click|stopPropagation
      on:keydown={handleKeydown}
    >
      <input
        bind:this={inputEl}
        bind:value={searchQuery}
        type="text"
        placeholder="Type a command or search..."
        class="command-input"
      />

      <div class="command-list">
        {#if filteredCommands.length === 0}
          <div class="no-results">No commands found</div>
        {:else}
          {#each filteredCommands as cmd, i}
            <button
              class="command-item"
              class:selected={i === selectedIndex}
              on:click={() => executeCommand(cmd)}
              on:mouseenter={() => selectedIndex = i}
            >
              <Icon name={cmd.icon} />
              <span class="command-label">{cmd.label}</span>
              {#if cmd.shortcut}
                <kbd class="command-shortcut">{cmd.shortcut}</kbd>
              {/if}
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</Portal>

<style>
  .command-palette-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 20vh;
    z-index: 1000;
  }

  .command-palette {
    background: white;
    border-radius: 8px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
    width: 90%;
    max-width: 600px;
    max-height: 60vh;
    display: flex;
    flex-direction: column;
  }

  .command-input {
    padding: 16px;
    border: none;
    border-bottom: 1px solid #e5e5e5;
    font-size: 16px;
    outline: none;
  }

  .command-list {
    overflow-y: auto;
    max-height: 400px;
  }

  .command-item {
    width: 100%;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 12px;
    border: none;
    background: none;
    cursor: pointer;
    transition: background 0.1s;
  }

  .command-item:hover,
  .command-item.selected {
    background: #f5f5f5;
  }

  .command-label {
    flex: 1;
    text-align: left;
  }

  .command-shortcut {
    padding: 4px 8px;
    background: #e5e5e5;
    border-radius: 4px;
    font-size: 12px;
  }
</style>
```

---

## 6. Performance Considerations

### 6.1 Virtualization for Large Documents

For documents with 100+ blocks, implement virtual scrolling:

```svelte
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let blocks: Block[];

  let containerEl: HTMLElement;
  let visibleBlocks: Block[] = [];
  let scrollTop = 0;

  const BLOCK_HEIGHT = 40; // Average block height
  const OVERSCAN = 5; // Render extra blocks for smooth scrolling

  $: {
    const startIndex = Math.max(0, Math.floor(scrollTop / BLOCK_HEIGHT) - OVERSCAN);
    const endIndex = Math.min(
      blocks.length,
      Math.ceil((scrollTop + containerEl?.clientHeight) / BLOCK_HEIGHT) + OVERSCAN
    );
    visibleBlocks = blocks.slice(startIndex, endIndex);
  }

  function handleScroll() {
    scrollTop = containerEl.scrollTop;
  }
</script>

<div
  bind:this={containerEl}
  class="virtual-list"
  on:scroll={handleScroll}
  style="height: {blocks.length * BLOCK_HEIGHT}px"
>
  {#each visibleBlocks as block}
    <BlockComponent {block} />
  {/each}
</div>
```

### 6.2 Debounced Updates

```typescript
// utils/debounce.ts
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;

  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };

    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// Usage in block component
const debouncedUpdate = debounce((content: string) => {
  editor.updateBlock(block.id, { content });
}, 300);
```

---

## 7. Summary: Key Patterns to Adopt

### Must Have
1. **Block Registry System** - Dynamic component loading based on block type
2. **Flat Block Storage** - Map with parent/child references for flexible hierarchy
3. **Compound Components** - Editor context shared via Svelte context API
4. **Slash Commands** - Searchable command menu with fuzzy matching
5. **Keyboard Shortcuts** - Global registry with context-aware handling
6. **Selection Management** - Dedicated store for block/text/range selection
7. **Drag & Drop** - Native HTML5 API with visual feedback
8. **Undo/Redo** - History stack in editor store

### Nice to Have
1. **Command Palette** - Cmd+K global search/actions
2. **Virtual Scrolling** - For large documents (100+ blocks)
3. **Debounced Updates** - Optimize performance during typing
4. **Block Actions Menu** - Hover + right-click context menu
5. **Formatting Toolbar** - Inline toolbar on text selection
6. **Tree View Sidebar** - Recursive component for page hierarchy
7. **Cover/Header System** - Page-level metadata (icon, cover image)

### Can Wait
1. **Real-time Collaboration** - Yjs CRDTs (complex, add later if needed)
2. **Plugin System** - Extensibility for custom blocks/commands
3. **Export/Import** - Markdown, PDF, etc.
4. **Database Blocks** - Complex table/kanban views
5. **AI Features** - Writing assistant, summarization

---

## 8. Next Steps

1. **Start with MVP**: Implement basic editor with paragraph blocks only
2. **Add Core Blocks**: Heading, list, code blocks
3. **Implement Slash Commands**: Most impactful UX feature
4. **Add Keyboard Shortcuts**: Power user efficiency
5. **Build Sidebar**: Navigation and page hierarchy
6. **Polish UI**: Animations, hover states, visual feedback
7. **Optimize Performance**: Virtual scrolling, debouncing
8. **Add Advanced Features**: Database blocks, collaboration

---

## Additional Resources

- **AFFiNE GitHub**: https://github.com/toeverything/AFFiNE
- **BlockSuite Docs**: https://blocksuite.io/
- **Yjs (CRDTs)**: https://yjs.dev/
- **ContentEditable**: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/contenteditable
- **Svelte Actions**: https://svelte.dev/docs#use_action

---

**Document Version**: 1.0
**Last Updated**: 2026-01-07
**Author**: Codebase Analyzer Agent
