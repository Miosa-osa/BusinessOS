<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { contexts } from '$lib/stores/contexts';
	import { kbPreferences, type KBSection, type KBViewMode } from '$lib/stores/kb-preferences';
	import { learning } from '$lib/stores/learning';
	import { api, type Block, type Conversation, type ArtifactListItem, type CalendarEvent } from '$lib/api/client';
	import { Dialog, Popover } from 'bits-ui';
	import type { ContextType, Context, ContextListItem } from '$lib/api/client';
	import Tooltip from '$lib/components/ui/Tooltip.svelte';
	import DocumentPeek from '$lib/components/contexts/DocumentPeek.svelte';
	import { KBSidebar, NewPageWelcome, KnowledgeGraphView, ContextProfileView, InlineDocumentEditor, CommandPalette, HomeView } from '$lib/components/kb';
	import { KnowledgeGraph, KnowledgeChatPanel, KnowledgeDocumentPanel } from '$lib/components/knowledge';
	import type { Memory } from '$lib/api/memory/types';
	import type { Learning } from '$lib/api/learning/types';

	// Dual-panel layout state for Knowledge Graph view
	interface KnowledgeChatMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		timestamp: Date;
	}
	let chatMessages = $state<KnowledgeChatMessage[]>([]);
	let chatStreaming = $state(false);
	let selectedGraphMemory = $state<Memory | null>(null);
	let leftPanelWidthKB = $state(340);
	let rightPanelWidthKB = $state(360);
	let isResizingLeftKB = $state(false);
	let isResizingRightKB = $state(false);

	// Inline document state (for main content area editing, NOT side peek)
	let inlineDocument = $state<Context | ContextListItem | null>(null);

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	// Panel state
	let leftPanelWidth = $state(320);
	let isResizing = $state(false);
	let leftPanelCollapsed = $state(false);
	let selectedProfileId = $state<string | null>(null);
	let selectedProfile = $state<Context | null>(null);
	let loadingProfile = $state(false);

	// Side peek state
	let showDocumentPeek = $state(false);
	let peekDocument = $state<Context | null>(null);
	let peekIsNew = $state(false);
	let peekParentId = $state<string | undefined>(undefined);

	// Center peek modal state
	let showCenterPeek = $state(false);
	let centerPeekDocument = $state<Context | null>(null);
	let centerPeekIsNew = $state(false);
	let centerPeekParentId = $state<string | undefined>(undefined);

	// View mode: 'all' | 'profile' | 'loose'
	let viewMode = $state<'all' | 'profile' | 'loose'>('all');

	// Document view style: 'table' | 'grid'
	let docViewStyle = $state<'table' | 'grid'>('table');

	// Document search
	let docSearch = $state('');

	// Side peek panel width (resizable)
	let peekPanelWidth = $state(560);
	let isResizingPeek = $state(false);

	// New KB layout state (from preferences store)
	let useNewLayout = $state(true); // Toggle between old and new layout
	let kbActiveSection = $state<KBSection>($kbPreferences.activeSection);
	let kbViewMode = $state<KBViewMode>($kbPreferences.viewMode);
	let kbExpandedSections = $derived(new Set($kbPreferences.expandedSections));
	// Pass arrays directly for better reactivity - Svelte 5 tracks array changes more reliably than Set.has()
	let kbExpandedPagesArray = $derived($kbPreferences.expandedPages);
	let kbFavoriteIdsArray = $derived($kbPreferences.favorites);
	let kbSidebarWidth = $state($kbPreferences.sidebarWidth);
	let kbSidebarCollapsed = $state($kbPreferences.sidebarCollapsed);
	let showCommandPalette = $state(false);
	let showHome = $state(true); // Show home dashboard by default

	// Graph content type - 'contexts' shows the page graph, 'memories' shows the 3D memory bubble graph
	let graphContentType = $state<'contexts' | 'memories'>('contexts');

	// Convert Learning to Memory format for the 3D graph component
	function learningToMemory(learning: Learning): Memory {
		return {
			id: learning.id,
			user_id: learning.user_id,
			title: learning.learning_summary || 'Untitled Memory',
			summary: learning.learning_content.slice(0, 200),
			content: learning.learning_content,
			memory_type: learning.learning_type as Memory['memory_type'] || 'learning',
			importance_score: learning.confidence_score || 0.5,
			is_pinned: false,
			is_active: true,
			tags: learning.category ? [learning.category] : [],
			metadata: {},
			source_type: learning.source_type,
			source_id: learning.source_id || null,
			project_id: null,
			node_id: null,
			expires_at: null,
			access_count: learning.times_applied || 0,
			last_accessed_at: learning.last_applied_at || null,
			created_at: learning.created_at,
			updated_at: learning.updated_at
		};
	}

	// Convert learnings to Memory format for the graph
	let memoriesForGraph = $derived($learning.learnings.map(learningToMemory));

	// Keep a reference to learnings for handler functions
	let currentLearnings = $derived($learning.learnings);

	// Selected profile for KB view (for profile database view instead of document editor)
	let kbSelectedProfile = $state<ContextListItem | null>(null);
	let kbSelectedMemory = $state<any | null>(null);

	// Check if a type is a profile type (shows database view)
	const PROFILE_TYPES = ['business', 'person', 'project', 'custom'];
	function isProfileType(type: string | null): boolean {
		return PROFILE_TYPES.includes(type || '');
	}

	// Filter out archived contexts for sidebar display
	let activeContexts = $derived($contexts.contexts.filter(c => !c.is_archived));

	// Archived contexts for trash view
	let archivedContexts = $derived($contexts.contexts.filter(c => c.is_archived));

	// Get children of selected profile
	let profileChildren = $derived(() => {
		const profile = kbSelectedProfile;
		return profile
			? activeContexts.filter(c => c.parent_id === profile.id)
			: [];
	});

	// Get favorite pages
	let favoritePages = $derived(
		activeContexts.filter(c => kbFavoriteIdsArray.includes(c.id))
	);

	// Get recent pages
	let recentPages = $derived(
		$kbPreferences.recentPages
			.map(id => activeContexts.find(c => c.id === id))
			.filter((c): c is ContextListItem => c !== undefined)
	);

	// Dialog states
	let showNewContext = $state(false);
	let showAssignModal = $state(false);
	let documentToAssign = $state<ContextListItem | null>(null);
	let showDeleteConfirm = $state(false);
	let itemToDelete = $state<ContextListItem | null>(null);

	let newContext = $state({
		name: '',
		type: 'person' as ContextType,
		content: '',
		system_prompt_template: '',
		icon: '',
		client_id: ''
	});
	let showAdvancedProfile = $state(false);

	// SVG Document Icons
	interface DocIcon {
		id: string;
		label: string;
		path: string;
	}

	const documentIcons: DocIcon[] = [
		// Documents
		{ id: 'document', label: 'Document', path: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z' },
		{ id: 'document-text', label: 'Text Document', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ id: 'clipboard', label: 'Clipboard', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2' },
		{ id: 'clipboard-list', label: 'Checklist', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
		{ id: 'document-duplicate', label: 'Documents', path: 'M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2' },
		// Folders
		{ id: 'folder', label: 'Folder', path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ id: 'folder-open', label: 'Open Folder', path: 'M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z' },
		{ id: 'archive', label: 'Archive', path: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		// Writing & Notes
		{ id: 'pencil', label: 'Pencil', path: 'M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z' },
		{ id: 'pencil-alt', label: 'Edit', path: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' },
		{ id: 'annotation', label: 'Annotation', path: 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' },
		{ id: 'book-open', label: 'Book', path: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
		// Data & Charts
		{ id: 'chart-bar', label: 'Bar Chart', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
		{ id: 'chart-pie', label: 'Pie Chart', path: 'M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z' },
		{ id: 'trending-up', label: 'Trending Up', path: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6' },
		{ id: 'trending-down', label: 'Trending Down', path: 'M13 17h8m0 0V9m0 8l-8-8-4 4-6-6' },
		{ id: 'table', label: 'Table', path: 'M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
		// Goals & Stars
		{ id: 'flag', label: 'Flag', path: 'M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9' },
		{ id: 'star', label: 'Star', path: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z' },
		{ id: 'sparkles', label: 'Sparkles', path: 'M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z' },
		{ id: 'fire', label: 'Fire', path: 'M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z' },
		// Ideas & Creativity
		{ id: 'lightbulb', label: 'Idea', path: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
		{ id: 'puzzle', label: 'Puzzle', path: 'M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z' },
		{ id: 'color-swatch', label: 'Design', path: 'M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01' },
		// Technology
		{ id: 'code', label: 'Code', path: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
		{ id: 'terminal', label: 'Terminal', path: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'database', label: 'Database', path: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4' },
		{ id: 'server', label: 'Server', path: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01' },
		{ id: 'chip', label: 'Chip', path: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z' },
		// Media
		{ id: 'photograph', label: 'Image', path: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'film', label: 'Video', path: 'M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z' },
		{ id: 'music-note', label: 'Music', path: 'M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3' },
		{ id: 'microphone', label: 'Audio', path: 'M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z' },
		// Business
		{ id: 'briefcase', label: 'Briefcase', path: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'currency-dollar', label: 'Finance', path: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'receipt-tax', label: 'Receipt', path: 'M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z' },
		{ id: 'presentation-chart-line', label: 'Presentation', path: 'M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z' },
		// Security
		{ id: 'lock-closed', label: 'Lock', path: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' },
		{ id: 'key', label: 'Key', path: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' },
		{ id: 'shield-check', label: 'Shield', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
		// People
		{ id: 'user', label: 'User', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
		{ id: 'users', label: 'Team', path: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
		{ id: 'user-group', label: 'Group', path: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
		// Objects
		{ id: 'cube', label: 'Cube', path: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4' },
		{ id: 'globe', label: 'Globe', path: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
		{ id: 'heart', label: 'Heart', path: 'M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z' },
		{ id: 'calendar', label: 'Calendar', path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'clock', label: 'Clock', path: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'location-marker', label: 'Location', path: 'M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z M15 11a3 3 0 11-6 0 3 3 0 016 0z' },
		{ id: 'home', label: 'Home', path: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
		{ id: 'office-building', label: 'Office', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
		// Actions
		{ id: 'check-circle', label: 'Complete', path: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'exclamation-circle', label: 'Important', path: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'question-mark-circle', label: 'Question', path: 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'lightning-bolt', label: 'Quick', path: 'M13 10V3L4 14h7v7l9-11h-7z' },
		{ id: 'rocket', label: 'Launch', path: 'M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z' },
	];

	// Helper to get icon SVG path by ID
	function getDocIconPath(iconId: string | null): string {
		if (!iconId) return documentIcons[0].path;
		const icon = documentIcons.find(i => i.id === iconId);
		return icon?.path || documentIcons[0].path;
	}

	// Common emoji icons for profiles (keeping emojis for profiles for now)
	const profileIcons = ['👤', '👥', '🏢', '🏠', '🏭', '💼', '📊', '📁', '🎯', '⭐', '💡', '🔧', '📝', '📚', '🎨', '🎓', '🏆', '💰', '🌟', '🚀'];

	// Icon picker state
	let showDocIconPicker = $state<string | null>(null);
	let typeFilter = $state('');
	let searchQuery = $state('');
	let expandedProfiles = $state<Set<string>>(new Set());

	// Collapsible sections
	let sectionsCollapsed = $state<Record<string, boolean>>({
		people: false,
		businesses: false,
		projects: false,
		other: false,
		documents: false
	});

	// Load clients for linking
	let clients = $state<{id: string, name: string}[]>([]);

	// Data Hub state
	type DataHubTab = 'documents' | 'conversations' | 'artifacts' | 'events';
	let activeDataHubTab = $state<DataHubTab>('documents');
	let linkedConversations = $state<Conversation[]>([]);
	let linkedArtifacts = $state<ArtifactListItem[]>([]);
	let linkedEvents = $state<CalendarEvent[]>([]);
	let loadingLinkedData = $state(false);

	onMount(async () => {
		contexts.loadContexts();
		learning.loadLearnings();
		try {
			clients = await api.getClients();
		} catch (e) {
			console.error('Failed to load clients:', e);
		}
	});


	// Separate profiles from documents
	let profiles = $derived($contexts.contexts.filter(c =>
		c.type !== 'document' && !c.parent_id
	));

	// Group profiles by type
	let personProfiles = $derived(profiles.filter(p => p.type === 'person'));
	let businessProfiles = $derived(profiles.filter(p => p.type === 'business'));
	let projectProfiles = $derived(profiles.filter(p => p.type === 'project'));
	let otherProfiles = $derived(profiles.filter(p => !['person', 'business', 'project'].includes(p.type)));

	let standaloneDocuments = $derived($contexts.contexts.filter(c =>
		c.type === 'document' && !c.parent_id
	));

	// All documents (for 'all' view)
	let allDocuments = $derived($contexts.contexts.filter(c => c.type === 'document'));

	// Documents to display in right panel based on viewMode and search
	let displayedDocuments = $derived.by(() => {
		let docs: ContextListItem[];
		if (viewMode === 'loose') {
			docs = standaloneDocuments;
		} else if (viewMode === 'profile' && selectedProfileId) {
			docs = getChildDocuments(selectedProfileId);
		} else {
			// 'all' mode
			docs = allDocuments;
		}

		// Apply search filter
		if (docSearch.trim()) {
			const search = docSearch.toLowerCase().trim();
			docs = docs.filter(d => d.name.toLowerCase().includes(search));
		}

		// Sort by updated_at (newest first)
		return docs.sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime());
	});

	// Title for right panel header
	let rightPanelTitle = $derived.by(() => {
		if (viewMode === 'loose') {
			return 'Loose Documents';
		} else if (viewMode === 'profile' && selectedProfile) {
			return selectedProfile.name;
		} else {
			return 'All Documents';
		}
	});

	// Get child documents for a profile
	function getChildDocuments(profileId: string): ContextListItem[] {
		return $contexts.contexts.filter(c => c.parent_id === profileId);
	}

	// Get parent profile name for a document
	function getParentName(parentId: string | null | undefined): string {
		if (!parentId) return '';
		const parent = profiles.find(p => p.id === parentId);
		return parent?.name || '';
	}

	// Handle peek panel resize
	function startPeekResize(e: MouseEvent) {
		isResizingPeek = true;
		e.preventDefault();
	}

	function handlePeekResize(e: MouseEvent) {
		if (!isResizingPeek) return;
		const newWidth = window.innerWidth - e.clientX;
		peekPanelWidth = Math.max(400, Math.min(900, newWidth));
	}

	function stopPeekResize() {
		isResizingPeek = false;
	}

	// Load linked data for a context (conversations, artifacts, events)
	async function loadLinkedData(contextId: string) {
		loadingLinkedData = true;
		try {
			const [convs, arts, events] = await Promise.all([
				api.getConversationsByContext(contextId).catch(() => []),
				api.getArtifacts({ contextId }).catch(() => []),
				api.getCalendarEvents({ contextId }).catch(() => [])
			]);
			linkedConversations = convs;
			linkedArtifacts = arts;
			linkedEvents = events;
		} catch (error) {
			console.error('Failed to load linked data:', error);
		} finally {
			loadingLinkedData = false;
		}
	}

	// Load full profile details
	async function selectProfile(profileId: string) {
		if (selectedProfileId === profileId) {
			// Toggle off if clicking same profile
			selectedProfileId = null;
			selectedProfile = null;
			viewMode = 'all';
			linkedConversations = [];
			linkedArtifacts = [];
			linkedEvents = [];
			return;
		}

		selectedProfileId = profileId;
		viewMode = 'profile';
		loadingProfile = true;
		activeDataHubTab = 'documents'; // Reset to documents tab
		try {
			selectedProfile = await contexts.loadContext(profileId);
			// Load linked data in background
			loadLinkedData(profileId);
		} catch (error) {
			console.error('Failed to load profile:', error);
		} finally {
			loadingProfile = false;
		}
	}

	function closeProfile() {
		selectedProfileId = null;
		selectedProfile = null;
		viewMode = 'all';
	}

	// Open document in main content area
	async function openDocument(docId: string) {
		try {
			const doc = await contexts.loadContext(docId);
			// DON'T clear inline document - side peek should overlay on top
			// Keep whatever is currently showing in main content
			peekDocument = doc;
			peekIsNew = false;
			peekParentId = doc.parent_id || undefined;
			showDocumentPeek = true;
		} catch (error) {
			console.error('Failed to load document:', error);
		}
	}

	// Close document peek
	function closeDocumentPeek() {
		showDocumentPeek = false;
		peekDocument = null;
		peekIsNew = false;
		peekParentId = undefined;
		// Refresh the contexts list
		contexts.loadContexts();
	}

	// Handle document saved from peek
	function handleDocumentSaved(doc: Context) {
		if (peekIsNew) {
			peekDocument = doc;
			peekIsNew = false;
		}
		// Refresh contexts list
		contexts.loadContexts();
		// Refresh selected profile if applicable
		if (selectedProfileId) {
			contexts.loadContext(selectedProfileId).then(p => {
				selectedProfile = p;
			});
		}
	}

	// Open document in center peek modal
	async function openCenterPeek(docId: string) {
		try {
			const doc = await contexts.loadContext(docId);
			centerPeekDocument = doc;
			centerPeekIsNew = false;
			centerPeekParentId = doc.parent_id || undefined;
			showCenterPeek = true;
		} catch (error) {
			console.error('Failed to load document for center peek:', error);
		}
	}

	// Close center peek modal
	function closeCenterPeek() {
		showCenterPeek = false;
		centerPeekDocument = null;
		centerPeekIsNew = false;
		centerPeekParentId = undefined;
		// Refresh the contexts list
		contexts.loadContexts();
	}

	// Handle document saved from center peek
	function handleCenterPeekSaved(doc: Context) {
		if (centerPeekIsNew) {
			centerPeekDocument = doc;
			centerPeekIsNew = false;
		}
		// Refresh contexts list
		contexts.loadContexts();
	}

	function toggleSection(section: string) {
		sectionsCollapsed[section] = !sectionsCollapsed[section];
	}

	// Resizer handling
	function startResize(e: MouseEvent) {
		isResizing = true;
		e.preventDefault();
	}

	function handleMouseMove(e: MouseEvent) {
		if (isResizing) {
			const newWidth = e.clientX;
			if (newWidth >= 200 && newWidth <= 500) {
				leftPanelWidth = newWidth;
			}
		}
	}

	function stopResize() {
		isResizing = false;
	}

	async function handleCreateContext(e: Event) {
		e.preventDefault();
		try {
			const ctx = await contexts.createContext({
				name: newContext.name,
				type: newContext.type,
				content: newContext.content || undefined,
				system_prompt_template: newContext.system_prompt_template || undefined,
				icon: newContext.icon || undefined,
				client_id: newContext.client_id || undefined
			});
			showNewContext = false;
			showAdvancedProfile = false;
			newContext = { name: '', type: 'person', content: '', system_prompt_template: '', icon: '', client_id: '' };
			// Select the new profile
			await selectProfile(ctx.id);
		} catch (error) {
			console.error('Failed to create context:', error);
		}
	}

	function createNewDocument(parentId?: string) {
		// Open side peek with new document state
		peekDocument = null;
		peekIsNew = true;
		peekParentId = parentId;
		showDocumentPeek = true;
	}

	function getTypeIcon(type: string) {
		switch (type) {
			case 'person': return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business': return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			case 'project': return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'document': return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			default: return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}

	function getTypeColor(type: string) {
		switch (type) {
			case 'person': return 'text-purple-600 bg-purple-50 border-purple-200';
			case 'business': return 'text-blue-600 bg-blue-50 border-blue-200';
			case 'project': return 'text-green-600 bg-green-50 border-green-200';
			case 'document': return 'text-amber-600 bg-amber-50 border-amber-200';
			default: return 'text-gray-600 bg-gray-50 border-gray-200';
		}
	}

	function getTypeLabel(type: string) {
		switch (type) {
			case 'person': return 'Person';
			case 'business': return 'Business';
			case 'project': return 'Project';
			case 'custom': return 'Custom';
			default: return type;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
	}

	function formatWordCount(count: number) {
		if (count === 0) return '';
		if (count === 1) return '1 word';
		return `${count.toLocaleString()} words`;
	}

	function openAssignModal(doc: ContextListItem) {
		documentToAssign = doc;
		showAssignModal = true;
	}

	async function assignDocumentToProfile(profileId: string) {
		if (!documentToAssign) return;
		try {
			await contexts.updateContext(documentToAssign.id, { parent_id: profileId });
			showAssignModal = false;
			documentToAssign = null;
			await contexts.loadContexts();
			// Select the profile to show the newly assigned doc
			await selectProfile(profileId);
		} catch (error) {
			console.error('Failed to assign document:', error);
		}
	}

	async function unlinkDocument(doc: ContextListItem) {
		try {
			await contexts.updateContext(doc.id, { parent_id: null });
			await contexts.loadContexts();
			// Refresh selected profile
			if (selectedProfileId) {
				selectedProfile = await contexts.loadContext(selectedProfileId);
			}
		} catch (error) {
			console.error('Failed to unlink document:', error);
		}
	}

	async function confirmDelete() {
		if (!itemToDelete) return;
		try {
			// Use archive instead of hard delete (soft delete / move to trash)
			await contexts.archiveContext(itemToDelete.id);
			showDeleteConfirm = false;
			// If we deleted the selected profile, close it
			if (itemToDelete.id === selectedProfileId) {
				closeProfile();
			}
			// Also close inline document if it's the one being deleted
			if (itemToDelete.id === inlineDocument?.id) {
				inlineDocument = null;
			}
			// Close side peek if it's showing the deleted item
			if (itemToDelete.id === peekDocument?.id) {
				showDocumentPeek = false;
				peekDocument = null;
			}
			// Close center peek if showing deleted item
			if (itemToDelete.id === centerPeekDocument?.id) {
				showCenterPeek = false;
				centerPeekDocument = null;
			}
			itemToDelete = null;
			// Refresh contexts to update sidebar
			await contexts.loadContexts();
		} catch (error) {
			console.error('Failed to move to trash:', error);
		}
	}

	async function updateProfileContent(content: string) {
		if (!selectedProfile) return;
		try {
			await contexts.updateContext(selectedProfile.id, { content });
			selectedProfile = { ...selectedProfile, content };
		} catch (error) {
			console.error('Failed to update profile:', error);
		}
	}

	async function updateDocumentIcon(docId: string, icon: string) {
		try {
			await contexts.updateContext(docId, { icon });
			// Refresh lists
			await contexts.loadContexts();
			// Update selected profile if viewing one
			if (selectedProfileId) {
				selectedProfile = await contexts.loadContext(selectedProfileId);
			}
			showDocIconPicker = null;
		} catch (error) {
			console.error('Failed to update icon:', error);
		}
	}

	// ============================================
	// NEW KB SIDEBAR HANDLERS
	// ============================================

	function handleKBSectionToggle(sectionId: string) {
		kbPreferences.toggleSection(sectionId);
	}

	function handleKBPageSelect(page: ContextListItem) {
		// Add to recent
		kbPreferences.addToRecent(page.id);
		// Hide home view when selecting a page
		showHome = false;

		// Check if this is a profile type (business, person, project, custom)
		if (isProfileType(page.type)) {
			// Show profile database view
			kbSelectedProfile = page;
			kbSelectedMemory = null;
			inlineDocument = null;
			showDocumentPeek = false;
			peekDocument = null;
		} else if (page.type === ('memory' as any)) {
			// Show memory view (use currentLearnings derived value)
			kbSelectedMemory = currentLearnings.find(m => m.id === page.id);
			kbSelectedProfile = null;
			inlineDocument = null;
			showDocumentPeek = false;
			peekDocument = null;
		} else {
			// Open document INLINE (in main content area, not side peek)
			kbSelectedProfile = null;
			kbSelectedMemory = null;
			inlineDocument = page;
			showDocumentPeek = false;
			peekDocument = null;
		}
	}

	// Handler for memory selection from the 3D graph
	function handleMemoryGraphSelect(memory: Memory) {
		// Set the selected memory for the document panel
		selectedGraphMemory = memory;

		// Also update the KB selection state for other views
		const learning = currentLearnings.find(l => l.id === memory.id);
		if (learning) {
			handleKBPageSelect({
				id: learning.id,
				name: learning.learning_summary || learning.learning_content,
				type: 'memory' as any,
				updated_at: learning.updated_at
			} as any);
		}
	}

	// Knowledge chat handlers
	function handleKnowledgeChatSend(message: string) {
		// Add user message
		const userMessage: KnowledgeChatMessage = {
			id: `user-${Date.now()}`,
			role: 'user',
			content: message,
			timestamp: new Date()
		};
		chatMessages = [...chatMessages, userMessage];
		chatStreaming = true;

		// Simulate AI response (TODO: Replace with actual API call)
		setTimeout(() => {
			const assistantMessage: KnowledgeChatMessage = {
				id: `assistant-${Date.now()}`,
				role: 'assistant',
				content: `I've analyzed your knowledge base regarding: "${message}"\n\nBased on your **${memoriesForGraph.length} memories** and documents, I can help you find relevant information. Try clicking on nodes in the graph to explore specific memories.`,
				timestamp: new Date()
			};
			chatMessages = [...chatMessages, assistantMessage];
			chatStreaming = false;
		}, 1500);
	}

	function handleKnowledgeChatStop() {
		chatStreaming = false;
	}

	function handleCloseDocumentPanel() {
		selectedGraphMemory = null;
	}

	function handleEditMemory() {
		if (selectedGraphMemory) {
			// Navigate to memory edit (handled by KB selection)
			kbSelectedMemory = currentLearnings.find(l => l.id === selectedGraphMemory?.id) || null;
			selectedGraphMemory = null;
		}
	}

	// Resize handlers for dual-panel layout
	function handleMouseMoveLeftKB(e: MouseEvent) {
		if (!isResizingLeftKB) return;
		const newWidth = Math.max(280, Math.min(500, e.clientX - (kbSidebarCollapsed ? 0 : kbSidebarWidth)));
		leftPanelWidthKB = newWidth;
	}

	function handleMouseMoveRightKB(e: MouseEvent) {
		if (!isResizingRightKB) return;
		const newWidth = Math.max(280, Math.min(500, window.innerWidth - e.clientX));
		rightPanelWidthKB = newWidth;
	}

	function handleMouseUpKB() {
		isResizingLeftKB = false;
		isResizingRightKB = false;
	}

	function handleKBPageExpand(pageId: string) {
		kbPreferences.togglePageExpanded(pageId);
	}

	async function handleKBPageAddChild(page: ContextListItem) {
		// Create new document as child of this page - open INLINE, not in side peek
		kbSelectedProfile = null;
		showDocumentPeek = false;
		peekDocument = null;
		showHome = false;

		try {
			const newDoc = await contexts.createContext({
				name: 'New page',
				type: 'document',
				parent_id: page.id,
				blocks: []
			});

			inlineDocument = newDoc;
			// Auto-expand parent so user can see the new child
			kbPreferences.expandPage(page.id);
			await contexts.loadContexts();
		} catch (error) {
			console.error('[KB] Failed to create child page:', error);
		}
	}

	function handleKBPageOpenPeek(page: ContextListItem) {
		openDocument(page.id);
	}

	function handleKBPageOpenCenterPeek(page: ContextListItem) {
		openCenterPeek(page.id);
	}

	async function handleKBPageDuplicate(page: ContextListItem) {
		try {
			await contexts.duplicateContext(page.id);
			await contexts.loadContexts();
		} catch (error) {
			console.error('Failed to duplicate:', error);
		}
	}

	async function handleKBPageRename(page: ContextListItem, newName: string) {
		try {
			await contexts.updateContext(page.id, { name: newName });
			await contexts.loadContexts();
		} catch (error) {
			console.error('Failed to rename:', error);
		}
	}

	function handleKBPageMove(page: ContextListItem) {
		// Open assign modal for this page
		documentToAssign = page;
		showAssignModal = true;
	}

	async function handleKBPageDelete(page: ContextListItem) {
		itemToDelete = page;
		showDeleteConfirm = true;
	}

	function handleKBPageToggleFavorite(page: ContextListItem) {
		kbPreferences.toggleFavorite(page.id);
	}

	function handleKBPageCopyLink(page: ContextListItem) {
		const url = `${window.location.origin}/knowledge/${page.id}`;
		navigator.clipboard.writeText(url);
	}

	async function handleKBAddPage(parentId?: string) {
		// Clear profile view and home view
		kbSelectedProfile = null;
		showDocumentPeek = false;
		showHome = false;

		// Create a new blank page immediately
		try {
			const createData: any = {
				name: 'New page',
				type: 'document'
			};
			if (parentId) {
				createData.parent_id = parentId;
			}
			const newDoc = await contexts.createContext(createData);

			// Show the new document inline for editing
			inlineDocument = newDoc;

			// Refresh sidebar
			await contexts.loadContexts();

			// Switch to list view if in graph view
			if (kbViewMode === 'graph') {
				handleKBViewModeChange('list');
			}
		} catch (error) {
			console.error('Failed to create new page:', error);
			alert('Failed to create page: ' + (error instanceof Error ? error.message : 'Unknown error'));
		}
	}

	async function handleKBAddProfile(type: 'business' | 'person' | 'project' | 'custom') {
		const icon = type === 'business' ? '🏢' : type === 'person' ? '👤' : type === 'project' ? '📁' : '✨';
		const defaultName = type === 'business' ? 'New Business' : type === 'person' ? 'New Person' : type === 'project' ? 'New Project' : 'New Profile';

		try {
			// Create the profile directly
			const newProfile = await contexts.createContext({
				name: defaultName,
				type: type,
				icon: icon,
				content: '',
				system_prompt_template: '',
				client_id: undefined
			});

			// Clear other views and show the new profile
			inlineDocument = null;
			showDocumentPeek = false;
			peekDocument = null;
			showHome = false;

			// Open the new profile
			kbSelectedProfile = newProfile;
		} catch (error) {
			console.error('Failed to create profile:', error);
		}
	}

	function handleKBSearch(query: string) {
		searchQuery = query;
		contexts.loadContexts({ search: query || undefined });
	}

	function handleKBSidebarWidthChange(width: number) {
		kbSidebarWidth = width;
		kbPreferences.setSidebarWidth(width);
	}

	function handleKBToggleSidebarCollapse() {
		kbSidebarCollapsed = !kbSidebarCollapsed;
		kbPreferences.toggleSidebarCollapsed();
	}

	function handleKBOpenCommandPalette() {
		showCommandPalette = true;
	}

	function handleKBGoHome() {
		// Go to home dashboard view - clear all selections
		inlineDocument = null;
		kbSelectedProfile = null;
		peekDocument = null;
		showDocumentPeek = false;
		showHome = true; // Show the home dashboard
		// Switch back to list view if in graph view
		if (kbViewMode === 'graph') {
			handleKBViewModeChange('list');
		}
	}

	function handleKBSectionChange(section: KBSection) {
		kbActiveSection = section;
		kbPreferences.setActiveSection(section);
	}

	function handleKBViewModeChange(mode: KBViewMode) {
		kbViewMode = mode;
		kbPreferences.setViewMode(mode);
	}

	// Get child pages for sidebar
	function getKBChildPages(parentId: string): ContextListItem[] {
		return $contexts.contexts.filter(c => c.parent_id === parentId);
	}

	// Search input reference for Cmd+K focus
	let kbSearchInputRef: HTMLInputElement | null = $state(null);

	// Keyboard shortcuts
	function handleGlobalKeydown(event: KeyboardEvent) {
		// Only handle in new layout
		if (!useNewLayout) return;

		const isMeta = event.metaKey || event.ctrlKey;

		// Cmd+N: Create new page
		if (isMeta && event.key === 'n') {
			event.preventDefault();
			handleKBAddPage();
		}

		// Cmd+K: Open command palette
		if (isMeta && event.key === 'k') {
			event.preventDefault();
			showCommandPalette = true;
		}

		// Escape: Close peek or deselect
		if (event.key === 'Escape') {
			if (showDocumentPeek) {
				closeDocumentPeek();
			}
		}

		// Cmd+\: Toggle sidebar
		if (isMeta && event.key === '\\') {
			event.preventDefault();
			handleKBToggleSidebarCollapse();
		}
	}

	// Handler for sidebar to set search ref
	function handleKBSearchInputRef(input: HTMLInputElement) {
		kbSearchInputRef = input;
	}
</script>

<svelte:window on:mousemove={handleMouseMove} on:mouseup={stopResize} on:keydown={handleGlobalKeydown} />

{#if useNewLayout}
<!-- NEW NOTION-STYLE LAYOUT -->
<div class="h-full flex flex-col bg-white dark:bg-[#1c1c1e]">
	<div class="flex-1 flex min-h-0">
		<!-- Notion-style Sidebar -->
		<KBSidebar
			pages={activeContexts}
			favorites={favoritePages}
			{recentPages}
			memories={$learning.learnings}
			selectedPageId={inlineDocument?.id || kbSelectedProfile?.id || kbSelectedMemory?.id || null}
			expandedSections={kbExpandedSections}
			expandedPages={kbExpandedPagesArray}
			favoriteIds={kbFavoriteIdsArray}
			searchQuery={searchQuery}
			width={kbSidebarWidth}
			isCollapsed={kbSidebarCollapsed}
			onSectionToggle={handleKBSectionToggle}
			onPageSelect={handleKBPageSelect}
			onPageExpand={handleKBPageExpand}
			onPageAddChild={handleKBPageAddChild}
			onPageOpenPeek={handleKBPageOpenPeek}
			onPageOpenCenterPeek={handleKBPageOpenCenterPeek}
			onPageDuplicate={handleKBPageDuplicate}
			onPageRename={handleKBPageRename}
			onPageMove={handleKBPageMove}
			onPageDelete={handleKBPageDelete}
			onPageToggleFavorite={handleKBPageToggleFavorite}
			onPageCopyLink={handleKBPageCopyLink}
			onAddPage={handleKBAddPage}
			onAddProfile={handleKBAddProfile}
			onSearch={handleKBSearch}
			onOpenCommandPalette={handleKBOpenCommandPalette}
			onGoHome={handleKBGoHome}
			onOpenGraph={() => handleKBViewModeChange('graph')}
			isGraphView={kbViewMode === 'graph'}
			onWidthChange={handleKBSidebarWidthChange}
			onToggleCollapse={handleKBToggleSidebarCollapse}
			onSearchInputRef={handleKBSearchInputRef}
		/>

		<!-- Main Content Area -->
		<div class="flex-1 flex flex-col min-w-0">
			{#if kbViewMode === 'graph'}
				<!-- 3D Knowledge Graph View with Dual Panel Layout -->
				<div
					class="flex-1 flex overflow-hidden"
					onmousemove={(e) => { handleMouseMoveLeftKB(e); handleMouseMoveRightKB(e); }}
					onmouseup={handleMouseUpKB}
					onmouseleave={handleMouseUpKB}
				>
					<!-- Left Panel: Chat (only for memories view) -->
					{#if graphContentType === 'memories'}
						<div class="flex-shrink-0 h-full" style="width: {leftPanelWidthKB}px">
							<KnowledgeChatPanel
								messages={chatMessages}
								streaming={chatStreaming}
								onSend={handleKnowledgeChatSend}
								onStop={handleKnowledgeChatStop}
							/>
						</div>
						<!-- Left Resize Handle -->
						<div
							class="w-1 h-full bg-transparent hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors"
							class:bg-blue-400={isResizingLeftKB}
							onmousedown={() => isResizingLeftKB = true}
						></div>
					{/if}
					<!-- Center: Graph View -->
					<div class="flex-1 flex flex-col min-w-0 overflow-hidden">
						<!-- Graph Type Toggle -->
					<div class="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-[#1c1c1e]">
						<span class="text-sm text-gray-500 dark:text-gray-400 mr-2">View:</span>
						<button
							class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {graphContentType === 'contexts' ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
							onclick={() => graphContentType = 'contexts'}
						>
							<span class="flex items-center gap-1.5">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								Pages ({$contexts.contexts.length})
							</span>
						</button>
						<button
							class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {graphContentType === 'memories' ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
							onclick={() => graphContentType = 'memories'}
						>
							<span class="flex items-center gap-1.5">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
								</svg>
								Memories ({memoriesForGraph.length})
							</span>
						</button>
					</div>

					<!-- Graph Content -->
					<div class="flex-1 overflow-hidden">
						{#if graphContentType === 'contexts'}
							<div class="w-full h-full p-4">
								<KnowledgeGraphView
									contexts={$contexts.contexts}
									selectedId={peekDocument?.id || kbSelectedProfile?.id}
									onSelect={(context) => handleKBPageSelect(context)}
									onNavigate={(context) => handleKBPageOpenPeek(context)}
								/>
							</div>
						{:else}
							<KnowledgeGraph
								memories={memoriesForGraph}
								selectedId={selectedGraphMemory?.id || kbSelectedMemory?.id}
								onSelect={handleMemoryGraphSelect}
							/>
						{/if}
					</div>
					</div>
					<!-- Right Panel: Document Viewer (only for memories view) -->
					{#if graphContentType === 'memories'}
						<!-- Right Resize Handle -->
						<div
							class="w-1 h-full bg-transparent hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors"
							class:bg-blue-400={isResizingRightKB}
							onmousedown={() => isResizingRightKB = true}
						></div>
						<div class="flex-shrink-0 h-full" style="width: {rightPanelWidthKB}px">
							<KnowledgeDocumentPanel
								selectedMemory={selectedGraphMemory}
								onClose={handleCloseDocumentPanel}
								onEdit={handleEditMemory}
							/>
						</div>
					{/if}
				</div>
			{:else if showHome}
				<!-- Home Dashboard View -->
				<HomeView
					pages={activeContexts}
					recentPages={recentPages}
					memories={$learning.learnings}
					onSelectPage={(page) => handleKBPageSelect(page)}
					onSelectMemory={(memory) => handleKBPageSelect({
						id: memory.id,
						name: memory.learning_summary || memory.learning_content,
						type: 'memory' as any,
						updated_at: memory.updated_at
					} as any)}
					onCreatePage={() => handleKBAddPage()}
				/>
			{:else if kbSelectedProfile}
				<!-- Profile Database View (for business, person, project types) -->
				<div class="flex-1 overflow-hidden">
					<ContextProfileView
						profile={kbSelectedProfile}
						children={profileChildren()}
						allPages={$contexts.contexts}
						onAddPage={() => kbSelectedProfile && handleKBAddPage(kbSelectedProfile.id)}
						onSelectPage={(page) => handleKBPageSelect(page)}
						onPageAction={(action, page) => {
							if (action === 'menu') {
								// TODO: Show context menu for the page
							}
						}}
						onUpdateProfile={async (updates) => {
							if (!kbSelectedProfile) return;
							try {
								await contexts.updateContext(kbSelectedProfile.id, updates);
								// Update local state
								kbSelectedProfile = { ...kbSelectedProfile, ...updates };
								// Refresh sidebar
								await contexts.loadContexts();
							} catch (error) {
								console.error('[KB] Failed to update profile:', error);
							}
						}}
					/>
				</div>
			{:else if kbSelectedMemory}
				<!-- Memory View -->
				<div class="flex-1 flex flex-col h-full bg-white dark:bg-[#1c1c1e] overflow-hidden">
					<div class="px-8 py-10 max-w-4xl mx-auto w-full space-y-8 h-full overflow-y-auto">
						<div class="flex items-center gap-4">
							<div class="w-12 h-12 rounded-2xl bg-pink-100 dark:bg-pink-900/30 flex items-center justify-center">
								<span class="text-2xl">🧠</span>
							</div>
							<div>
								<h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100">
									{kbSelectedMemory.learning_summary || 'Extracted Memory'}
								</h2>
								<p class="text-sm text-gray-500 dark:text-gray-400">
									Learned on {new Date(kbSelectedMemory.created_at).toLocaleDateString()}
								</p>
							</div>
						</div>

						<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-2xl p-6 border border-gray-200 dark:border-gray-700/50">
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-4">Content</h3>
							<p class="text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap">
								{kbSelectedMemory.learning_content}
							</p>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div class="p-4 rounded-xl bg-gray-50 dark:bg-[#2c2c2e] border border-gray-200 dark:border-gray-700/50">
								<span class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase">Confidence</span>
								<div class="mt-2 flex items-center gap-2">
									<div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
										<div class="h-full bg-pink-500" style="width: {(kbSelectedMemory.confidence_score || 0) * 100}%"></div>
									</div>
									<span class="text-sm font-semibold text-gray-700 dark:text-gray-300">{((kbSelectedMemory.confidence_score || 0) * 100).toFixed(0)}%</span>
								</div>
							</div>
							<div class="p-4 rounded-xl bg-gray-50 dark:bg-[#2c2c2e] border border-gray-200 dark:border-gray-700/50">
								<span class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase">Status</span>
								<div class="mt-1 flex items-center gap-2">
									<span class="px-2 py-0.5 rounded-full text-xs font-medium {kbSelectedMemory.was_validated ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'}">
										{kbSelectedMemory.was_validated ? 'Validated' : 'Pending Validation'}
									</span>
									{#if kbSelectedMemory.is_active}
										<span class="px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
											Active In Context
										</span>
									{/if}
								</div>
							</div>
						</div>

						{#if kbSelectedMemory.tags && kbSelectedMemory.tags.length > 0}
							<div class="space-y-2">
								<span class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase">Tags</span>
								<div class="flex flex-wrap gap-2">
									{#each kbSelectedMemory.tags as tag}
										<span class="px-2.5 py-1 bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 rounded-lg text-xs border border-gray-200 dark:border-gray-700">
											#{tag}
										</span>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				</div>
			{:else if inlineDocument}
				<!-- Inline Document Editor (full width in main content area) -->
				<div class="flex-1 overflow-hidden">
					<InlineDocumentEditor
						document={inlineDocument}
						allPages={$contexts.contexts}
						onSaved={(doc) => {
							// Refresh the context list
							contexts.loadContexts();
						}}
						onTitleChange={(title) => {
							// Update the context list item name
							if (inlineDocument) {
								inlineDocument = { ...inlineDocument, name: title };
							}
						}}
						onPageClick={(pageId) => {
							// Navigate to the page within KB module (not via URL)
							const page = $contexts.contexts.find(c => c.id === pageId);
							if (page) {
								handleKBPageSelect(page);
							}
						}}
					/>
				</div>
			{:else}
				<!-- Fallback: New Page Welcome / Empty State -->
				<div class="flex-1 overflow-auto">
					<NewPageWelcome
						onCreateBlank={() => handleKBAddPage()}
						onAskAI={() => goto('/chat')}
						onSelectTemplate={(templateId) => {
							// TODO: Handle template selection
							handleKBAddPage();
						}}
					/>
				</div>
			{/if}

			<!-- Side Peek Panel (only when explicitly opened from context menu) -->
			{#if showDocumentPeek && peekDocument}
				<DocumentPeek
					document={peekDocument}
					isNew={peekIsNew}
					parentId={peekParentId}
					onClose={closeDocumentPeek}
					onSaved={handleDocumentSaved}
					{embedSuffix}
					width={peekPanelWidth}
					onResize={(w) => peekPanelWidth = w}
				/>
			{:else if showDocumentPeek && peekIsNew}
				<DocumentPeek
					document={null}
					isNew={true}
					parentId={peekParentId}
					onClose={closeDocumentPeek}
					onSaved={handleDocumentSaved}
					{embedSuffix}
					width={peekPanelWidth}
					onResize={(w) => peekPanelWidth = w}
				/>
			{/if}
		</div>
	</div>

	<!-- Command Palette (Cmd+K) -->
	<CommandPalette
		pages={activeContexts}
		isOpen={showCommandPalette}
		onClose={() => showCommandPalette = false}
		onSelect={(page) => handleKBPageSelect(page)}
	/>

	<!-- Center Peek Modal -->
	<Dialog.Root bind:open={showCenterPeek}>
		<Dialog.Portal>
			<Dialog.Overlay class="fixed inset-0 z-[100] bg-black/50" />
			<Dialog.Content class="fixed left-1/2 top-1/2 z-[101] -translate-x-1/2 -translate-y-1/2 w-[90vw] max-w-[900px] h-[85vh] bg-white dark:bg-[#1c1c1e] rounded-2xl shadow-2xl overflow-hidden flex flex-col">
				{#if centerPeekDocument}
					<!-- Header -->
					<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
						<div class="flex items-center gap-3">
							{#if centerPeekDocument.icon}
								<span class="text-2xl">{centerPeekDocument.icon}</span>
							{:else}
								<div class="w-8 h-8 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
									<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</div>
							{/if}
							<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
								{centerPeekDocument.name || 'New page'}
							</h2>
						</div>
						<div class="flex items-center gap-2">
							<button
								onclick={() => { goto(`/knowledge/${centerPeekDocument?.id}${embedSuffix}`); closeCenterPeek(); }}
								class="px-3 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors flex items-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
								Open full page
							</button>
							<Dialog.Close
								class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
								aria-label="Close"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</Dialog.Close>
						</div>
					</div>
					<!-- Content using InlineDocumentEditor -->
					<div class="flex-1 overflow-hidden">
						<InlineDocumentEditor
							document={centerPeekDocument}
							allPages={$contexts.contexts}
							onSaved={(doc) => {
								centerPeekDocument = doc;
								contexts.loadContexts();
							}}
							onPageClick={(pageId) => {
								// Navigate to the clicked page in center peek
								const page = $contexts.contexts.find(c => c.id === pageId);
								if (page) {
									centerPeekDocument = page as unknown as Context;
								}
							}}
						/>
					</div>
				{/if}
			</Dialog.Content>
		</Dialog.Portal>
	</Dialog.Root>
</div>

{:else}
<!-- ORIGINAL LAYOUT -->
<div class="h-full flex bg-white dark:bg-[#1c1c1e]">
	<!-- Left Panel: Profile List -->
	{#if !leftPanelCollapsed}
		<div
			class="bg-gray-50 dark:bg-[#2c2c2e] border-r border-gray-200 dark:border-gray-700/50 flex flex-col h-full flex-shrink-0"
			style="width: {leftPanelWidth}px"
		>
			<!-- Header -->
			<div class="p-4 border-b border-gray-200 dark:border-gray-700/50">
				<div class="flex items-center justify-between mb-3">
					<h1 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Knowledge Base</h1>
					<button
						onclick={() => showNewContext = true}
						class="p-1.5 rounded-lg bg-blue-600 text-white hover:bg-blue-500 transition-colors"
						title="Create new profile"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</button>
				</div>

				<!-- Search -->
				<div class="relative">
					<svg class="w-4 h-4 text-gray-400 dark:text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search..."
						class="w-full text-sm pl-9 pr-3 py-2 bg-white dark:bg-[#1c1c1e] border border-gray-200 dark:border-gray-700/50 rounded-lg text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500/50"
						oninput={() => contexts.loadContexts({ search: searchQuery || undefined })}
					/>
				</div>
			</div>

			<!-- Filter Tabs -->
			<div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700/50 flex gap-1 overflow-x-auto">
				{#each [
					{ value: '', label: 'All' },
					{ value: 'person', label: 'People' },
					{ value: 'business', label: 'Business' },
					{ value: 'project', label: 'Projects' },
				] as filter}
					<button
						onclick={() => { typeFilter = filter.value; contexts.loadContexts({ type: filter.value || undefined }); }}
						class="px-3 py-1 text-xs font-medium rounded-full whitespace-nowrap transition-colors {typeFilter === filter.value ? 'bg-blue-600 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-200'}"
					>
						{filter.label}
					</button>
				{/each}
			</div>

			<!-- Profile List -->
			<div class="flex-1 overflow-y-auto">
				{#if $contexts.loading}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full"></div>
					</div>
				{:else if profiles.length === 0 && standaloneDocuments.length === 0}
					<div class="p-6 text-center">
						<div class="w-12 h-12 rounded-xl bg-gray-100 dark:bg-gray-700/50 flex items-center justify-center mx-auto mb-3">
							<svg class="w-6 h-6 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
						</div>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-3">No profiles yet</p>
						<button onclick={() => showNewContext = true} class="text-sm text-blue-600 dark:text-blue-400 font-medium hover:underline">
							Create your first profile
						</button>
					</div>
				{:else}
					<!-- Profiles by Type -->
					{#each [
						{ key: 'people', type: 'person', profiles: personProfiles, label: 'People', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
						{ key: 'businesses', type: 'business', profiles: businessProfiles, label: 'Businesses', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
						{ key: 'projects', type: 'project', profiles: projectProfiles, label: 'Projects', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
						{ key: 'other', type: 'custom', profiles: otherProfiles, label: 'Other', icon: 'M4 6h16M4 10h16M4 14h16M4 18h16' }
					].filter(g => g.profiles.length > 0) as group}
						<div class="py-1">
							<button
								onclick={() => toggleSection(group.key)}
								class="w-full px-4 py-1.5 flex items-center gap-2 text-xs font-medium text-gray-500 uppercase tracking-wider hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
							>
								<svg class="w-3 h-3 transition-transform {sectionsCollapsed[group.key] ? '-rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={group.icon} />
								</svg>
								{group.label}
								<span class="text-gray-400 dark:text-gray-600">({group.profiles.length})</span>
							</button>
							{#if !sectionsCollapsed[group.key]}
								{#each group.profiles as profile}
									{@const childCount = getChildDocuments(profile.id).length}
									<button
										onclick={() => selectProfile(profile.id)}
										class="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors text-left {selectedProfileId === profile.id ? 'bg-gray-100 dark:bg-gray-700' : ''}"
									>
										{#if profile.icon}
											<span class="text-xl">{profile.icon}</span>
										{:else}
											<div class="w-8 h-8 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center flex-shrink-0">
												<svg class="w-4 h-4 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(profile.type)} />
												</svg>
											</div>
										{/if}
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{profile.name}</p>
											<p class="text-xs text-gray-500">{childCount} doc{childCount !== 1 ? 's' : ''}</p>
										</div>
										{#if selectedProfileId === profile.id}
											<div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
										{/if}
									</button>
								{/each}
							{/if}
						</div>
					{/each}

					<!-- Loose Documents Entry -->
					{#if standaloneDocuments.length > 0}
						<div class="py-1 border-t border-gray-200 dark:border-gray-700/50">
							<button
								onclick={() => { viewMode = 'loose'; selectedProfileId = null; selectedProfile = null; }}
								class="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors text-left {viewMode === 'loose' ? 'bg-gray-100 dark:bg-gray-700' : ''}"
							>
								<div class="w-8 h-8 rounded-lg bg-amber-100 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-700/50 flex items-center justify-center flex-shrink-0">
									<svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900 dark:text-gray-100">Loose Documents</p>
									<p class="text-xs text-gray-500">{standaloneDocuments.length} unassigned</p>
								</div>
								{#if viewMode === 'loose'}
									<div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
								{/if}
							</button>
						</div>
					{/if}
				{/if}
			</div>

			<!-- Quick Actions -->
			<div class="p-3 border-t border-gray-200 dark:border-gray-700/50">
				<button
					onclick={() => createNewDocument()}
					class="w-full px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700/50 hover:text-gray-900 dark:hover:text-gray-200 rounded-lg transition-colors flex items-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					New Document
				</button>
			</div>
		</div>

		<!-- Resize Handle -->
		<div
			class="w-1 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-500 cursor-col-resize transition-colors flex-shrink-0 {isResizing ? 'bg-gray-300 dark:bg-gray-500' : ''}"
			onmousedown={startResize}
			role="separator"
			aria-orientation="vertical"
		></div>
	{/if}

	<!-- Right Panel: Profile Detail or Empty State -->
	<div class="flex-1 flex flex-col h-full overflow-hidden">
		{#if leftPanelCollapsed}
			<!-- Collapsed sidebar toggle -->
			<button
				onclick={() => leftPanelCollapsed = false}
				class="absolute top-4 left-4 z-10 p-2 bg-white dark:bg-[#2c2c2e] rounded-lg shadow-md hover:shadow-lg transition-shadow border border-gray-200 dark:border-gray-700/50"
				title="Show sidebar"
			>
				<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
		{/if}

		{#if loadingProfile}
			<div class="flex-1 flex items-center justify-center bg-white dark:bg-[#1c1c1e]">
				<div class="animate-spin h-8 w-8 border-2 border-blue-500 border-t-transparent rounded-full"></div>
			</div>
		{:else if selectedProfile}
			<!-- Profile Detail View -->
			<div class="flex-1 overflow-y-auto bg-white dark:bg-[#1c1c1e]">
				<!-- Profile Header -->
				<div class="sticky top-0 bg-gray-50 dark:bg-[#2c2c2e] border-b border-gray-200 dark:border-gray-700/50 z-10">
					<div class="px-6 py-4 flex items-center gap-4">
						<button
							onclick={() => leftPanelCollapsed = !leftPanelCollapsed}
							class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
							title={leftPanelCollapsed ? 'Show sidebar' : 'Hide sidebar'}
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								{#if leftPanelCollapsed}
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
								{:else}
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
								{/if}
							</svg>
						</button>

						{#if selectedProfile.icon}
							<span class="text-3xl">{selectedProfile.icon}</span>
						{:else}
							<div class="w-12 h-12 rounded-xl bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
								<svg class="w-6 h-6 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(selectedProfile.type)} />
								</svg>
							</div>
						{/if}

						<div class="flex-1">
							<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{selectedProfile.name}</h2>
							<p class="text-sm text-gray-500 dark:text-gray-400">{getTypeLabel(selectedProfile.type)} Profile</p>
						</div>

						<div class="flex items-center gap-2">
							<button
								onclick={() => createNewDocument(selectedProfile?.id)}
								class="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors flex items-center gap-1.5"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Add Document
							</button>
							<button
								onclick={() => { itemToDelete = selectedProfile; showDeleteConfirm = true; }}
								class="p-2 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 text-gray-500 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
								title="Delete profile"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
							<button
								onclick={closeProfile}
								class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
								title="Close"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					</div>
				</div>

				<div class="p-6 space-y-6">
					<!-- Context Information -->
					<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 overflow-hidden">
						<div class="px-4 py-3 bg-gray-100 dark:bg-[#3c3c3e] border-b border-gray-200 dark:border-gray-700/50 flex items-center justify-between">
							<h3 class="text-sm font-medium text-gray-900 dark:text-gray-100">Context Information</h3>
							<span class="text-xs text-gray-500">This information is used by AI</span>
						</div>
						<div class="p-4">
							<textarea
								value={selectedProfile.content || ''}
								onchange={(e) => updateProfileContent((e.target as HTMLTextAreaElement).value)}
								placeholder="Add context information about this profile... (e.g., background, preferences, history, notes)"
								class="w-full min-h-[120px] text-sm text-gray-700 dark:text-gray-200 bg-transparent resize-none border-0 focus:ring-0 p-0 placeholder:text-gray-400 dark:placeholder:text-gray-500"
							></textarea>
						</div>
					</div>

					<!-- Data Hub Tabs -->
					<div>
						<div class="flex items-center justify-between mb-4">
							<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Data Hub</h3>
							{#if loadingLinkedData}
								<div class="animate-spin h-4 w-4 border-2 border-gray-500 border-t-blue-500 rounded-full"></div>
							{/if}
						</div>

						<!-- Tab Navigation -->
						<div class="flex gap-1 mb-4 bg-gray-100 dark:bg-[#3c3c3e] p-1 rounded-lg">
							{#each [
								{ id: 'documents', label: 'Documents', count: getChildDocuments(selectedProfile.id).length, icon: '📄' },
								{ id: 'conversations', label: 'Chats', count: linkedConversations.length, icon: '💬' },
								{ id: 'artifacts', label: 'Artifacts', count: linkedArtifacts.length, icon: '✨' },
								{ id: 'events', label: 'Events', count: linkedEvents.length, icon: '📅' }
							] as tab}
								<button
									onclick={() => activeDataHubTab = tab.id as DataHubTab}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium rounded-md transition-all {activeDataHubTab === tab.id ? 'bg-white dark:bg-[#2c2c2e] text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
								>
									<span>{tab.icon}</span>
									<span>{tab.label}</span>
									{#if tab.count > 0}
										<span class="px-1.5 py-0.5 rounded-full text-[10px] {activeDataHubTab === tab.id ? 'bg-blue-600 text-white' : 'bg-gray-300 dark:bg-gray-600 text-gray-600 dark:text-gray-300'}">{tab.count}</span>
									{/if}
								</button>
							{/each}
						</div>

						<!-- Documents Tab -->
						{#if activeDataHubTab === 'documents'}
							<div class="flex items-center justify-between mb-3">
								<span class="text-xs text-gray-500">{getChildDocuments(selectedProfile.id).length} documents</span>
								<button
									onclick={() => createNewDocument(selectedProfile?.id)}
									class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 flex items-center gap-1"
								>
									<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
									Add
								</button>
							</div>

						{#if getChildDocuments(selectedProfile.id).length === 0}
							<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 border-dashed p-8 text-center">
								<div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-700/50 flex items-center justify-center mx-auto mb-3">
									<svg class="w-6 h-6 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</div>
								<p class="text-sm text-gray-500 dark:text-gray-400 mb-3">No documents in this profile yet</p>
								<button
									onclick={() => createNewDocument(selectedProfile?.id)}
									class="text-sm text-blue-600 dark:text-blue-400 font-medium hover:underline"
								>
									Create your first document
								</button>
							</div>
						{:else}
							<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
								{#each getChildDocuments(selectedProfile.id) as doc}
									<div class="bg-white dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-lg transition-all group relative">
										<!-- Icon picker (outside the main button to avoid nesting) -->
										<div class="absolute top-4 left-4 z-10">
											<button
												onclick={(e) => { e.stopPropagation(); showDocIconPicker = showDocIconPicker === doc.id ? null : doc.id; }}
												class="hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg p-1.5 transition-colors"
											>
												<svg class="w-6 h-6 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getDocIconPath(doc.icon)} />
												</svg>
											</button>
											{#if showDocIconPicker === doc.id}
												<div class="absolute top-full left-0 mt-1 bg-white dark:bg-[#3c3c3e] rounded-xl shadow-xl border border-gray-200 dark:border-gray-600 p-3 w-72" role="menu">
													<p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">Choose icon</p>
													<div class="grid grid-cols-6 gap-1.5 max-h-64 overflow-y-auto">
														{#each documentIcons as docIcon}
															<button
																onclick={() => updateDocumentIcon(doc.id, docIcon.id)}
																class="w-9 h-9 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-600 flex items-center justify-center transition-colors {doc.icon === docIcon.id ? 'bg-blue-100 dark:bg-blue-900/50 ring-2 ring-blue-500' : ''}"
																title={docIcon.label}
															>
																<svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={docIcon.path} />
																</svg>
															</button>
														{/each}
													</div>
												</div>
											{/if}
										</div>
										<!-- Document card content -->
										<button onclick={() => openDocument(doc.id)} class="block p-4 pl-14 pb-10 text-left w-full">
											<div class="flex-1 min-w-0">
												<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{doc.name}</h4>
												<p class="text-xs text-gray-500 mt-0.5">
													{formatWordCount(doc.word_count)}
													{#if doc.word_count > 0} · {/if}
													{formatDate(doc.updated_at)}
												</p>
											</div>
										</button>
										<div class="absolute bottom-2 right-2 flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity bg-white/80 dark:bg-[#2c2c2e]/80 backdrop-blur-sm rounded-lg p-0.5">
											<Tooltip text="Unlink" position="top">
												<button
													onclick={() => unlinkDocument(doc)}
													class="p-1.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
													</svg>
												</button>
											</Tooltip>
											<Tooltip text="Delete" position="top">
												<button
													onclick={() => { itemToDelete = doc; showDeleteConfirm = true; }}
													class="p-1.5 rounded-md hover:bg-red-100 dark:hover:bg-red-900/50 text-gray-500 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
													</svg>
												</button>
											</Tooltip>
										</div>
									</div>
								{/each}
							</div>
						{/if}
						{/if}

						<!-- Conversations Tab -->
						{#if activeDataHubTab === 'conversations'}
							{#if linkedConversations.length === 0}
								<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 border-dashed p-8 text-center">
									<div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-700/50 flex items-center justify-center mx-auto mb-3">
										<span class="text-2xl">💬</span>
									</div>
									<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No conversations linked</p>
									<p class="text-xs text-gray-500">Start a chat with this context selected to link it here</p>
								</div>
							{:else}
								<div class="space-y-2">
									{#each linkedConversations as conv}
										<a
											href="/chat?conversation={conv.id}"
											class="block bg-white dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-lg transition-all group"
										>
											<div class="flex items-start gap-3">
												<div class="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
													<span class="text-lg">💬</span>
												</div>
												<div class="flex-1 min-w-0">
													<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{conv.title || 'Untitled Chat'}</h4>
													<p class="text-xs text-gray-500 mt-0.5">
														{conv.message_count || 0} messages · {formatDate(conv.updated_at)}
													</p>
												</div>
												<svg class="w-4 h-4 text-gray-400 dark:text-gray-600 group-hover:text-gray-600 dark:group-hover:text-gray-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
												</svg>
											</div>
										</a>
									{/each}
								</div>
							{/if}
						{/if}

						<!-- Artifacts Tab -->
						{#if activeDataHubTab === 'artifacts'}
							{#if linkedArtifacts.length === 0}
								<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 border-dashed p-8 text-center">
									<div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-700/50 flex items-center justify-center mx-auto mb-3">
										<span class="text-2xl">✨</span>
									</div>
									<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No artifacts linked</p>
									<p class="text-xs text-gray-500">Generated content from chats will appear here</p>
								</div>
							{:else}
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
									{#each linkedArtifacts as artifact}
										<div class="bg-white dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-lg transition-all group">
											<div class="flex items-start gap-3">
												<div class="w-10 h-10 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
													{#if artifact.type === 'code'}
														<span class="text-lg">💻</span>
													{:else if artifact.type === 'document'}
														<span class="text-lg">📄</span>
													{:else}
														<span class="text-lg">✨</span>
													{/if}
												</div>
												<div class="flex-1 min-w-0">
													<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{artifact.title}</h4>
													<p class="text-xs text-gray-500 mt-0.5 capitalize">{artifact.type} · {formatDate(artifact.created_at)}</p>
												</div>
											</div>
											{#if artifact.summary}
												<p class="text-xs text-gray-500 dark:text-gray-400 mt-2 line-clamp-2">{artifact.summary}</p>
											{/if}
										</div>
									{/each}
								</div>
							{/if}
						{/if}

						<!-- Events Tab -->
						{#if activeDataHubTab === 'events'}
							{#if linkedEvents.length === 0}
								<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 border-dashed p-8 text-center">
									<div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-700/50 flex items-center justify-center mx-auto mb-3">
										<span class="text-2xl">📅</span>
									</div>
									<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No events linked</p>
									<p class="text-xs text-gray-500">Calendar events associated with this profile will appear here</p>
								</div>
							{:else}
								<div class="space-y-2">
									{#each linkedEvents as event}
										<div class="bg-white dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-lg transition-all">
											<div class="flex items-start gap-3">
												<div class="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center flex-shrink-0">
													<span class="text-lg">📅</span>
												</div>
												<div class="flex-1 min-w-0">
													<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{event.title || 'Untitled Event'}</h4>
													<p class="text-xs text-gray-500 mt-0.5">
														{new Date(event.start_time).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })}
													</p>
													{#if event.location}
														<p class="text-xs text-gray-500 mt-0.5 flex items-center gap-1">
															<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
															</svg>
															{event.location}
														</p>
													{/if}
												</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						{/if}
					</div>

					<!-- System Prompt (Advanced) -->
					{#if selectedProfile.system_prompt_template}
						<div class="bg-gray-50 dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 overflow-hidden">
							<div class="px-4 py-3 bg-gray-100 dark:bg-[#3c3c3e] border-b border-gray-200 dark:border-gray-700/50">
								<h3 class="text-sm font-medium text-gray-900 dark:text-gray-100">System Prompt (Advanced)</h3>
							</div>
							<div class="p-4">
								<pre class="text-xs text-gray-600 dark:text-gray-400 whitespace-pre-wrap font-mono">{selectedProfile.system_prompt_template}</pre>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<!-- Document List View (for 'all' or 'loose' modes) -->
			<div class="flex-1 overflow-y-auto bg-white dark:bg-[#1c1c1e]">
				<!-- Header with Search -->
				<div class="sticky top-0 bg-white dark:bg-[#1c1c1e] border-b border-gray-200 dark:border-gray-700/50 z-10">
					<div class="px-6 py-3 flex items-center gap-4">
						<button
							onclick={() => leftPanelCollapsed = !leftPanelCollapsed}
							class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
							title={leftPanelCollapsed ? 'Show sidebar' : 'Hide sidebar'}
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								{#if leftPanelCollapsed}
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
								{:else}
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
								{/if}
							</svg>
						</button>

						<!-- Search Input -->
						<div class="flex-1 relative">
							<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
							</svg>
							<input
								type="text"
								bind:value={docSearch}
								placeholder="Search documents..."
								class="w-full pl-10 pr-4 py-2 bg-gray-100 dark:bg-[#2c2c2e] border-0 rounded-lg text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
							{#if docSearch}
								<button
									onclick={() => docSearch = ''}
									class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</button>
							{/if}
						</div>

						<!-- View Toggle -->
						<div class="flex items-center bg-gray-100 dark:bg-[#2c2c2e] rounded-lg p-1">
							<button
								onclick={() => docViewStyle = 'table'}
								class="p-1.5 rounded-md transition-colors {docViewStyle === 'table' ? 'bg-white dark:bg-gray-700 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}"
								title="Table view"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
								</svg>
							</button>
							<button
								onclick={() => docViewStyle = 'grid'}
								class="p-1.5 rounded-md transition-colors {docViewStyle === 'grid' ? 'bg-white dark:bg-gray-700 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}"
								title="Grid view"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
								</svg>
							</button>
						</div>

						<!-- New Document Button -->
						<button
							onclick={() => createNewDocument(viewMode === 'loose' ? undefined : selectedProfileId || undefined)}
							class="px-3 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors flex items-center gap-1.5"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							New
						</button>
					</div>

					<!-- Table Header (only in table view) -->
					{#if docViewStyle === 'table'}
						<div class="px-6 py-2 border-t border-gray-100 dark:border-gray-800 flex items-center gap-4 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
							<div class="w-8"></div>
							<div class="flex-1">Name</div>
							<div class="w-40 hidden md:block">Location</div>
							<div class="w-32 hidden sm:block">Updated</div>
							<div class="w-20 text-right">Words</div>
						</div>
					{/if}
				</div>

				<!-- Document Content -->
				<div class="{docViewStyle === 'table' ? '' : 'p-6'}">
					{#if displayedDocuments.length === 0}
						<div class="text-center py-12">
							<div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-purple-100 dark:from-purple-900/50 to-blue-100 dark:to-blue-900/50 flex items-center justify-center mx-auto mb-4">
								<svg class="w-8 h-8 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
							</div>
							<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
								{#if docSearch}
									No documents match "{docSearch}"
								{:else}
									No documents yet
								{/if}
							</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
								{#if docSearch}
									Try a different search term
								{:else if viewMode === 'loose'}
									All documents are assigned to profiles
								{:else}
									Create your first document to get started
								{/if}
							</p>
							{#if !docSearch}
								<button
									onclick={() => createNewDocument()}
									class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors text-sm font-medium"
								>
									Create Document
								</button>
							{/if}
						</div>
					{:else if docViewStyle === 'table'}
						<!-- Table View -->
						<div class="divide-y divide-gray-100 dark:divide-gray-800">
							{#each displayedDocuments as doc}
								<button
									onclick={() => openDocument(doc.id)}
									class="w-full px-6 py-3 flex items-center gap-4 hover:bg-gray-50 dark:hover:bg-[#2c2c2e] transition-colors text-left group"
								>
									<!-- Icon -->
									<div class="w-8 flex-shrink-0">
										<svg class="w-5 h-5 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getDocIconPath(doc.icon)} />
										</svg>
									</div>
									<!-- Name -->
									<div class="flex-1 min-w-0">
										<span class="text-sm text-gray-900 dark:text-gray-100 group-hover:text-blue-600 dark:group-hover:text-blue-400 truncate block">{doc.name}</span>
									</div>
									<!-- Location -->
									<div class="w-40 hidden md:block">
										{#if doc.parent_id}
											<span class="text-xs text-gray-500 dark:text-gray-400 truncate block">{getParentName(doc.parent_id)}</span>
										{:else}
											<span class="text-xs text-gray-400 dark:text-gray-500">—</span>
										{/if}
									</div>
									<!-- Updated -->
									<div class="w-32 hidden sm:block">
										<span class="text-xs text-gray-500 dark:text-gray-400">{formatDate(doc.updated_at)}</span>
									</div>
									<!-- Words -->
									<div class="w-20 text-right">
										<span class="text-xs text-gray-500 dark:text-gray-400">{formatWordCount(doc.word_count)}</span>
									</div>
								</button>
							{/each}
						</div>
					{:else}
						<!-- Grid View -->
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
							{#each displayedDocuments as doc}
								<div class="bg-white dark:bg-[#2c2c2e] rounded-xl border border-gray-200 dark:border-gray-700/50 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-lg transition-all group relative">
									<!-- Icon -->
									<div class="absolute top-4 left-4 z-10">
										<svg class="w-6 h-6 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getDocIconPath(doc.icon)} />
										</svg>
									</div>
									<!-- Document card content -->
									<button onclick={() => openDocument(doc.id)} class="block p-4 pl-14 pb-12 text-left w-full">
										<div class="flex-1 min-w-0">
											<h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{doc.name}</h4>
											<p class="text-xs text-gray-500 mt-1">
												{formatWordCount(doc.word_count)}
												{#if doc.word_count > 0} · {/if}
												{formatDate(doc.updated_at)}
											</p>
											{#if doc.parent_id && viewMode === 'all'}
												{@const parentProfile = profiles.find(p => p.id === doc.parent_id)}
												{#if parentProfile}
													<p class="text-xs text-gray-400 mt-1">{parentProfile.name}</p>
												{/if}
											{/if}
										</div>
									</button>
									<div class="absolute bottom-2 right-2 flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity bg-white/80 dark:bg-[#2c2c2e]/80 backdrop-blur-sm rounded-lg p-0.5">
										<Tooltip text="Delete" position="top">
											<button
												onclick={() => { itemToDelete = doc; showDeleteConfirm = true; }}
												class="p-1.5 rounded-md hover:bg-red-100 dark:hover:bg-red-900/50 text-gray-500 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
												</svg>
											</button>
										</Tooltip>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>

</div>

<!-- New Profile Dialog -->
<Dialog.Root bind:open={showNewContext}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg z-50 max-h-[90vh] overflow-y-auto">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Create New Profile</Dialog.Title>

			<form onsubmit={handleCreateContext} class="space-y-4">
				<!-- Icon & Name Row -->
				<div class="flex items-start gap-3">
					<!-- Icon Picker -->
					<div class="flex-shrink-0">
						<label class="block text-xs font-medium text-gray-500 mb-1">Icon</label>
						<div class="relative group">
							<button
								type="button"
								class="w-14 h-14 rounded-xl border-2 border-gray-200 hover:border-gray-300 flex items-center justify-center text-2xl bg-gray-50 transition-colors"
							>
								{newContext.icon || '📁'}
							</button>
							<div class="absolute top-full left-0 mt-1 bg-white rounded-lg shadow-xl border border-gray-200 p-2 grid grid-cols-5 gap-1 z-10 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
								{#each profileIcons as icon}
									<button
										type="button"
										onclick={() => newContext.icon = icon}
										class="w-8 h-8 rounded hover:bg-gray-100 text-lg flex items-center justify-center {newContext.icon === icon ? 'bg-gray-200' : ''}"
									>
										{icon}
									</button>
								{/each}
							</div>
						</div>
					</div>

					<!-- Name Input -->
					<div class="flex-1">
						<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
						<input
							id="name"
							type="text"
							bind:value={newContext.name}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900/10 focus:border-gray-300"
							placeholder="e.g. John Smith, Acme Corp"
							required
						/>
					</div>
				</div>

				<!-- Type Selection -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Profile Type</label>
					<div class="grid grid-cols-4 gap-2">
						{#each [
							{ value: 'person', label: 'Person', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z', emoji: '👤', desc: 'Individual contact' },
							{ value: 'business', label: 'Business', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', emoji: '🏢', desc: 'Company or org' },
							{ value: 'project', label: 'Project', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', emoji: '📁', desc: 'Work or initiative' },
							{ value: 'custom', label: 'Custom', icon: 'M4 6h16M4 10h16M4 14h16M4 18h16', emoji: '✨', desc: 'Other type' }
						] as typeOption}
							<button
								type="button"
								onclick={() => { newContext.type = typeOption.value as ContextType; if (!newContext.icon) newContext.icon = typeOption.emoji; }}
								class="p-3 rounded-lg border-2 transition-all flex flex-col items-center gap-1 {newContext.type === typeOption.value ? 'border-gray-900 bg-gray-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<span class="text-xl">{typeOption.emoji}</span>
								<span class="text-xs font-medium {newContext.type === typeOption.value ? 'text-gray-900' : 'text-gray-500'}">{typeOption.label}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Context/Description -->
				<div>
					<label for="content" class="block text-sm font-medium text-gray-700 mb-1">
						Context Information
						<span class="text-gray-400 font-normal">(used by AI)</span>
					</label>
					<textarea
						id="content"
						bind:value={newContext.content}
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900/10 focus:border-gray-300 resize-none"
						rows="3"
						placeholder="Background info, preferences, history, important notes..."
					></textarea>
					<p class="text-xs text-gray-400 mt-1">This information will be included when chatting with AI about this profile.</p>
				</div>

				<!-- Advanced Options Toggle -->
				<button
					type="button"
					onclick={() => showAdvancedProfile = !showAdvancedProfile}
					class="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700"
				>
					<svg class="w-4 h-4 transition-transform {showAdvancedProfile ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
					Advanced Options
				</button>

				<!-- Advanced Options -->
				{#if showAdvancedProfile}
					<div class="space-y-4 pt-2 border-t border-gray-100">
						<!-- Link to Client (for person/business types) -->
						{#if (newContext.type === 'person' || newContext.type === 'business') && clients.length > 0}
							<div>
								<label for="client" class="block text-sm font-medium text-gray-700 mb-1">Link to Client</label>
								<select
									id="client"
									bind:value={newContext.client_id}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900/10 focus:border-gray-300"
								>
									<option value="">No client linked</option>
									{#each clients as client}
										<option value={client.id}>{client.name}</option>
									{/each}
								</select>
							</div>
						{/if}

						<!-- System Prompt Template -->
						<div>
							<label for="system_prompt" class="block text-sm font-medium text-gray-700 mb-1">
								Custom System Prompt
								<span class="text-gray-400 font-normal">(advanced)</span>
							</label>
							<textarea
								id="system_prompt"
								bind:value={newContext.system_prompt_template}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900/10 focus:border-gray-300 resize-none font-mono text-xs"
								rows="4"
								placeholder="Custom instructions for AI when using this profile..."
							></textarea>
							<p class="text-xs text-gray-400 mt-1">Override default AI behavior for this specific profile.</p>
						</div>
					</div>
				{/if}

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => { showNewContext = false; showAdvancedProfile = false; }} class="btn-pill btn-pill-ghost flex-1">
						Cancel
					</button>
					<button type="submit" class="btn-pill btn-pill-primary flex-1">
						Create Profile
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Assign to Profile Dialog -->
<Dialog.Root bind:open={showAssignModal}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-2">Assign to Profile</Dialog.Title>
			<Dialog.Description class="text-sm text-gray-500 mb-4">
				{#if documentToAssign}
					Choose a profile for "{documentToAssign.name}"
				{/if}
			</Dialog.Description>

			<div class="space-y-1 max-h-64 overflow-y-auto">
				{#each profiles as profile}
					<button
						onclick={() => assignDocumentToProfile(profile.id)}
						class="btn-pill btn-pill-ghost btn-pill-block justify-start"
					>
						{#if profile.icon}
							<span class="text-xl">{profile.icon}</span>
						{:else}
							<div class="w-8 h-8 rounded-lg {getTypeColor(profile.type)} flex items-center justify-center">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(profile.type)} />
								</svg>
							</div>
						{/if}
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900">{profile.name}</p>
							<p class="text-xs text-gray-400">{getTypeLabel(profile.type)}</p>
						</div>
					</button>
				{/each}
			</div>

			<div class="mt-4 pt-4 border-t border-gray-100">
				<button type="button" onclick={() => { showAssignModal = false; documentToAssign = null; }} class="btn-pill btn-pill-ghost btn-pill-block">
					Cancel
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Delete Confirmation Dialog -->
<Dialog.Root bind:open={showDeleteConfirm}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm z-50">
			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center flex-shrink-0">
					<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
					</svg>
				</div>
				<div>
					<Dialog.Title class="text-lg font-semibold text-gray-900">Move to Trash?</Dialog.Title>
					<Dialog.Description class="text-sm text-gray-500">
						This item will be moved to trash and can be restored later.
					</Dialog.Description>
				</div>
			</div>

			{#if itemToDelete}
				<div class="bg-gray-50 rounded-lg p-3 mb-4">
					<div class="flex items-center gap-2">
						<span class="text-lg">{itemToDelete.icon || (itemToDelete.type === 'document' ? '📄' : '👤')}</span>
						<span class="font-medium text-gray-900">{itemToDelete.name || 'New page'}</span>
					</div>
				</div>
			{/if}

			<div class="flex gap-3">
				<button
					type="button"
					onclick={() => { showDeleteConfirm = false; itemToDelete = null; }}
					class="flex-1 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					type="button"
					onclick={confirmDelete}
					class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
				>
					Move to Trash
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Document Side Peek (Original Layout) -->
{#if showDocumentPeek}
	<DocumentPeek
		document={peekDocument}
		isNew={peekIsNew}
		parentId={peekParentId}
		onClose={closeDocumentPeek}
		onSaved={handleDocumentSaved}
		{embedSuffix}
		width={peekPanelWidth}
		onResize={(w) => peekPanelWidth = w}
	/>
{/if}
{/if}
<!-- End of useNewLayout conditional -->
