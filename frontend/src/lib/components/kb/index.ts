// Knowledge Base Components
export { default as KBSidebar } from './KBSidebar.svelte';
export { default as SidebarSection } from './SidebarSection.svelte';
export { default as SidebarPageItem } from './SidebarPageItem.svelte';
export { default as NewPageWelcome } from './NewPageWelcome.svelte';
export { default as KnowledgeGraphView } from './KnowledgeGraphView.svelte';
export { default as ContextProfileView } from './ContextProfileView.svelte';
export { default as InlineDocumentEditor } from './InlineDocumentEditor.svelte';
export { default as CommandPalette } from './CommandPalette.svelte';
export { default as HomeView } from './HomeView.svelte';

// Types
export type KBSection = 'home' | 'recent' | 'favorites' | 'all';
export type KBViewMode = 'list' | 'graph';
