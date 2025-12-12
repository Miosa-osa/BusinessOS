<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { contexts } from '$lib/stores/contexts';
	import { api, type Block } from '$lib/api/client';
	import { Dialog, Popover } from 'bits-ui';
	import type { ContextType, Context, ContextListItem } from '$lib/api/client';
	import { editor, wordCount, type EditorBlock } from '$lib/stores/editor';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';
	import Tooltip from '$lib/components/ui/Tooltip.svelte';

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	// Panel state
	let leftPanelWidth = $state(320);
	let isResizing = $state(false);
	let leftPanelCollapsed = $state(false);
	let selectedProfileId = $state<string | null>(null);
	let selectedProfile = $state<Context | null>(null);
	let loadingProfile = $state(false);

	// Document Editor Panel State
	type DocumentPanelMode = 'hidden' | 'side' | 'center' | 'full';
	let documentPanelMode = $state<DocumentPanelMode>('hidden');
	let selectedDocumentId = $state<string | null>(null);
	let selectedDocument = $state<Context | null>(null);
	let loadingDocument = $state(false);
	let documentPanelWidth = $state(550);
	let isResizingDocument = $state(false);
	let documentTitle = $state('');
	let autoSaveTimer: ReturnType<typeof setTimeout>;

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

	// Common emoji icons for profiles
	const profileIcons = ['👤', '👥', '🏢', '🏠', '🏭', '💼', '📊', '📁', '🎯', '⭐', '💡', '🔧', '📝', '📚', '🎨', '🎓', '🏆', '💰', '🌟', '🚀'];

	// Document icons
	const documentIcons = ['📄', '📝', '📋', '📑', '📃', '📜', '📰', '🗒️', '🗂️', '📂', '📁', '🗃️', '💼', '📊', '📈', '📉', '🎯', '💡', '⭐', '🌟', '✨', '🔥', '💎', '🎨', '🎬', '📸', '🎵', '🎮', '🔬', '🧪', '💊', '🏥', '⚖️', '🔒', '🔑', '💳', '🏦', '📱', '💻', '🖥️'];

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

	onMount(async () => {
		contexts.loadContexts();
		try {
			clients = await api.getClients();
		} catch (e) {
			console.error('Failed to load clients:', e);
		}
	});

	onDestroy(() => {
		if (autoSaveTimer) clearTimeout(autoSaveTimer);
		editor.reset();
	});

	// Auto-save document with debounce
	$effect(() => {
		if ($editor.isDirty && selectedDocument && documentPanelMode !== 'hidden') {
			if (autoSaveTimer) clearTimeout(autoSaveTimer);
			autoSaveTimer = setTimeout(async () => {
				await saveDocument();
			}, 1500);
		}
	});

	async function saveDocument() {
		if (!selectedDocument || $editor.isSaving) return;
		editor.setSaving(true);
		try {
			await contexts.updateBlocks(selectedDocument.id, $editor.blocks, $wordCount);
			editor.markSaved();
		} catch (e) {
			console.error('Failed to save:', e);
			editor.setSaving(false);
		}
	}

	async function updateDocumentTitle() {
		if (!selectedDocument || documentTitle === selectedDocument.name) return;
		try {
			await contexts.updateContext(selectedDocument.id, { name: documentTitle });
			selectedDocument = { ...selectedDocument, name: documentTitle };
			// Refresh the list
			await contexts.loadContexts();
		} catch (e) {
			console.error('Failed to update title:', e);
		}
	}

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

	// Get child documents for a profile
	function getChildDocuments(profileId: string): ContextListItem[] {
		return $contexts.contexts.filter(c => c.parent_id === profileId);
	}

	// Load full profile details
	async function selectProfile(profileId: string) {
		if (selectedProfileId === profileId) {
			// Toggle off if clicking same profile
			selectedProfileId = null;
			selectedProfile = null;
			return;
		}

		selectedProfileId = profileId;
		loadingProfile = true;
		try {
			selectedProfile = await contexts.loadContext(profileId);
		} catch (error) {
			console.error('Failed to load profile:', error);
		} finally {
			loadingProfile = false;
		}
	}

	function closeProfile() {
		selectedProfileId = null;
		selectedProfile = null;
	}

	// Document editor panel
	async function openDocument(docId: string, mode: DocumentPanelMode = 'side') {
		if (selectedDocumentId === docId && documentPanelMode !== 'hidden') {
			// Already open, maybe switch mode
			documentPanelMode = mode;
			return;
		}

		loadingDocument = true;
		selectedDocumentId = docId;
		documentPanelMode = mode;

		try {
			const doc = await contexts.loadContext(docId);
			selectedDocument = doc;
			documentTitle = doc.name;
			editor.initialize(doc.blocks);
		} catch (error) {
			console.error('Failed to load document:', error);
			closeDocument();
		} finally {
			loadingDocument = false;
		}
	}

	function closeDocument() {
		documentPanelMode = 'hidden';
		selectedDocumentId = null;
		selectedDocument = null;
		editor.reset();
	}

	// Document resizer
	function startDocumentResize(e: MouseEvent) {
		isResizingDocument = true;
		e.preventDefault();
	}

	function handleDocumentResize(e: MouseEvent) {
		if (!isResizingDocument) return;
		const newWidth = window.innerWidth - e.clientX;
		documentPanelWidth = Math.min(Math.max(newWidth, 400), 900);
	}

	function stopDocumentResize() {
		isResizingDocument = false;
	}

	function addNewBlockAtEnd() {
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			const newBlockId = editor.addBlockAfter(lastBlock.id);
			setTimeout(() => {
				const blockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				blockEl?.focus();
			}, 10);
		}
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
		if (isResizingDocument) {
			handleDocumentResize(e);
		}
	}

	function stopResize() {
		isResizing = false;
		isResizingDocument = false;
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

	async function createNewDocument(parentId?: string) {
		try {
			const ctx = await contexts.createContext({
				name: 'Untitled',
				type: 'document',
				parent_id: parentId,
				blocks: []
			});
			// Open in side panel instead of navigating
			await openDocument(ctx.id);
			await contexts.loadContexts();
		} catch (error) {
			console.error('Failed to create document:', error);
		}
	}

	function openDocumentFullScreen(docId: string) {
		goto(`/contexts/${docId}${embedSuffix}`);
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
			await contexts.deleteContext(itemToDelete.id);
			showDeleteConfirm = false;
			// If we deleted the selected profile, close it
			if (itemToDelete.id === selectedProfileId) {
				closeProfile();
			}
			itemToDelete = null;
		} catch (error) {
			console.error('Failed to delete:', error);
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
			// Update selected document if open
			if (selectedDocument && selectedDocument.id === docId) {
				selectedDocument = { ...selectedDocument, icon };
			}
			showDocIconPicker = null;
		} catch (error) {
			console.error('Failed to update icon:', error);
		}
	}
</script>

<svelte:window on:mousemove={handleMouseMove} on:mouseup={stopResize} />

<div class="h-full flex bg-gray-50">
	<!-- Left Panel: Profile List -->
	{#if !leftPanelCollapsed}
		<div
			class="bg-white border-r border-gray-200 flex flex-col h-full flex-shrink-0"
			style="width: {leftPanelWidth}px"
		>
			<!-- Header -->
			<div class="p-4 border-b border-gray-100">
				<div class="flex items-center justify-between mb-3">
					<h1 class="text-lg font-semibold text-gray-900">Knowledge Base</h1>
					<button
						onclick={() => showNewContext = true}
						class="p-1.5 rounded-lg bg-gray-900 text-white hover:bg-gray-800 transition-colors"
						title="Create new profile"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</button>
				</div>

				<!-- Search -->
				<div class="relative">
					<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search..."
						class="w-full text-sm pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900/10 focus:border-gray-300"
						oninput={() => contexts.loadContexts({ search: searchQuery || undefined })}
					/>
				</div>
			</div>

			<!-- Filter Tabs -->
			<div class="px-4 py-2 border-b border-gray-100 flex gap-1 overflow-x-auto">
				{#each [
					{ value: '', label: 'All' },
					{ value: 'person', label: 'People' },
					{ value: 'business', label: 'Business' },
					{ value: 'project', label: 'Projects' },
				] as filter}
					<button
						onclick={() => { typeFilter = filter.value; contexts.loadContexts({ type: filter.value || undefined }); }}
						class="px-3 py-1 text-xs font-medium rounded-full whitespace-nowrap transition-colors {typeFilter === filter.value ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						{filter.label}
					</button>
				{/each}
			</div>

			<!-- Profile List -->
			<div class="flex-1 overflow-y-auto">
				{#if $contexts.loading}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
					</div>
				{:else if profiles.length === 0 && standaloneDocuments.length === 0}
					<div class="p-6 text-center">
						<div class="w-12 h-12 rounded-xl bg-gray-100 flex items-center justify-center mx-auto mb-3">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
						</div>
						<p class="text-sm text-gray-500 mb-3">No profiles yet</p>
						<button onclick={() => showNewContext = true} class="text-sm text-gray-900 font-medium hover:underline">
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
								class="w-full px-4 py-1.5 flex items-center gap-2 text-xs font-medium text-gray-400 uppercase tracking-wider hover:text-gray-600 transition-colors"
							>
								<svg class="w-3 h-3 transition-transform {sectionsCollapsed[group.key] ? '-rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={group.icon} />
								</svg>
								{group.label}
								<span class="text-gray-300">({group.profiles.length})</span>
							</button>
							{#if !sectionsCollapsed[group.key]}
								{#each group.profiles as profile}
									{@const childCount = getChildDocuments(profile.id).length}
									<button
										onclick={() => selectProfile(profile.id)}
										class="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-gray-50 transition-colors text-left {selectedProfileId === profile.id ? 'bg-gray-100' : ''}"
									>
										{#if profile.icon}
											<span class="text-xl">{profile.icon}</span>
										{:else}
											<div class="w-8 h-8 rounded-lg {getTypeColor(profile.type)} flex items-center justify-center flex-shrink-0">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(profile.type)} />
												</svg>
											</div>
										{/if}
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-900 truncate">{profile.name}</p>
											<p class="text-xs text-gray-400">{childCount} doc{childCount !== 1 ? 's' : ''}</p>
										</div>
										{#if selectedProfileId === profile.id}
											<div class="w-1.5 h-1.5 rounded-full bg-gray-900"></div>
										{/if}
									</button>
								{/each}
							{/if}
						</div>
					{/each}

					<!-- Standalone Documents -->
					{#if standaloneDocuments.length > 0}
						<div class="py-1 border-t border-gray-100">
							<button
								onclick={() => toggleSection('documents')}
								class="w-full px-4 py-1.5 flex items-center gap-2 text-xs font-medium text-gray-400 uppercase tracking-wider hover:text-gray-600 transition-colors"
							>
								<svg class="w-3 h-3 transition-transform {sectionsCollapsed['documents'] ? '-rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
								</svg>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								Loose Documents
								<span class="text-gray-300">({standaloneDocuments.length})</span>
							</button>
							{#if !sectionsCollapsed['documents']}
								{#each standaloneDocuments as doc}
									<button
										onclick={() => openDocument(doc.id)}
										class="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-gray-50 transition-colors text-left {selectedDocumentId === doc.id ? 'bg-blue-50' : ''}"
									>
										<span class="text-xl">{doc.icon || '📄'}</span>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-900 truncate">{doc.name}</p>
											<p class="text-xs text-gray-400">{formatDate(doc.updated_at)}</p>
										</div>
										<div class="flex items-center gap-1">
											{#if profiles.length > 0}
												<button
													onclick={(e) => { e.stopPropagation(); openAssignModal(doc); }}
													class="p-1 rounded hover:bg-gray-200 text-gray-400 transition-colors"
													title="Assign to profile"
												>
													<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
													</svg>
												</button>
											{/if}
											{#if selectedDocumentId === doc.id}
												<div class="w-1.5 h-1.5 rounded-full bg-blue-600"></div>
											{/if}
										</div>
									</button>
								{/each}
							{/if}
						</div>
					{/if}
				{/if}
			</div>

			<!-- Quick Actions -->
			<div class="p-3 border-t border-gray-100">
				<button
					onclick={() => createNewDocument()}
					class="w-full px-3 py-2 text-sm text-gray-600 hover:bg-gray-50 rounded-lg transition-colors flex items-center gap-2"
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
			class="w-1 bg-gray-200 hover:bg-gray-400 cursor-col-resize transition-colors flex-shrink-0 {isResizing ? 'bg-gray-400' : ''}"
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
				class="absolute top-4 left-4 z-10 p-2 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
				title="Show sidebar"
			>
				<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
		{/if}

		{#if loadingProfile}
			<div class="flex-1 flex items-center justify-center">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else if selectedProfile}
			<!-- Profile Detail View -->
			<div class="flex-1 overflow-y-auto">
				<!-- Profile Header -->
				<div class="sticky top-0 bg-white border-b border-gray-200 z-10">
					<div class="px-6 py-4 flex items-center gap-4">
						<button
							onclick={() => leftPanelCollapsed = !leftPanelCollapsed}
							class="p-2 rounded-lg hover:bg-gray-100 text-gray-500 transition-colors"
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
							<div class="w-12 h-12 rounded-xl {getTypeColor(selectedProfile.type)} flex items-center justify-center">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(selectedProfile.type)} />
								</svg>
							</div>
						{/if}

						<div class="flex-1">
							<h2 class="text-xl font-semibold text-gray-900">{selectedProfile.name}</h2>
							<p class="text-sm text-gray-500">{getTypeLabel(selectedProfile.type)} Profile</p>
						</div>

						<div class="flex items-center gap-2">
							<button
								onclick={() => createNewDocument(selectedProfile?.id)}
								class="px-3 py-1.5 text-sm bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors flex items-center gap-1.5"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Add Document
							</button>
							<button
								onclick={() => { itemToDelete = selectedProfile; showDeleteConfirm = true; }}
								class="p-2 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-600 transition-colors"
								title="Delete profile"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
							<button
								onclick={closeProfile}
								class="p-2 rounded-lg hover:bg-gray-100 text-gray-500 transition-colors"
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
					<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
						<div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
							<h3 class="text-sm font-medium text-gray-900">Context Information</h3>
							<span class="text-xs text-gray-400">This information is used by AI</span>
						</div>
						<div class="p-4">
							<textarea
								value={selectedProfile.content || ''}
								onchange={(e) => updateProfileContent((e.target as HTMLTextAreaElement).value)}
								placeholder="Add context information about this profile... (e.g., background, preferences, history, notes)"
								class="w-full min-h-[120px] text-sm text-gray-700 resize-none border-0 focus:ring-0 p-0 placeholder:text-gray-400"
							></textarea>
						</div>
					</div>

					<!-- Documents Section -->
					<div>
						<div class="flex items-center justify-between mb-3">
							<h3 class="text-sm font-medium text-gray-900">Documents</h3>
							<span class="text-xs text-gray-400">{getChildDocuments(selectedProfile.id).length} documents</span>
						</div>

						{#if getChildDocuments(selectedProfile.id).length === 0}
							<div class="bg-white rounded-xl border border-gray-200 border-dashed p-8 text-center">
								<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
									<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</div>
								<p class="text-sm text-gray-500 mb-3">No documents in this profile yet</p>
								<button
									onclick={() => createNewDocument(selectedProfile?.id)}
									class="text-sm text-gray-900 font-medium hover:underline"
								>
									Create your first document
								</button>
							</div>
						{:else}
							<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
								{#each getChildDocuments(selectedProfile.id) as doc}
									<div class="bg-white rounded-xl border border-gray-200 hover:shadow-md transition-all group relative {selectedDocumentId === doc.id ? 'ring-2 ring-blue-500' : ''}">
										<!-- Icon picker (outside the main button to avoid nesting) -->
										<div class="absolute top-4 left-4 z-10">
											<button
												onclick={(e) => { e.stopPropagation(); showDocIconPicker = showDocIconPicker === doc.id ? null : doc.id; }}
												class="text-2xl hover:bg-gray-100 rounded-lg p-1 transition-colors"
											>
												{doc.icon || '📄'}
											</button>
											{#if showDocIconPicker === doc.id}
												<div class="absolute top-full left-0 mt-1 bg-white rounded-xl shadow-xl border border-gray-200 p-3 w-64" role="menu">
													<p class="text-xs font-medium text-gray-500 mb-2">Choose icon</p>
													<div class="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
														{#each documentIcons as icon}
															<button
																onclick={() => updateDocumentIcon(doc.id, icon)}
																class="w-7 h-7 rounded hover:bg-gray-100 text-lg flex items-center justify-center transition-colors {doc.icon === icon ? 'bg-gray-200' : ''}"
															>
																{icon}
															</button>
														{/each}
													</div>
												</div>
											{/if}
										</div>
										<!-- Document card content -->
										<button onclick={() => openDocument(doc.id)} class="block p-4 pl-14 pb-10 text-left w-full">
											<div class="flex-1 min-w-0">
												<h4 class="text-sm font-medium text-gray-900 truncate">{doc.name}</h4>
												<p class="text-xs text-gray-400 mt-0.5">
													{formatWordCount(doc.word_count)}
													{#if doc.word_count > 0} · {/if}
													{formatDate(doc.updated_at)}
												</p>
											</div>
										</button>
										<div class="absolute bottom-2 right-2 flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity bg-white/80 backdrop-blur-sm rounded-lg p-0.5">
											<Tooltip text="Full screen" position="top">
												<button
													onclick={() => openDocumentFullScreen(doc.id)}
													class="p-1.5 rounded-md hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
													</svg>
												</button>
											</Tooltip>
											<Tooltip text="Unlink" position="top">
												<button
													onclick={() => unlinkDocument(doc)}
													class="p-1.5 rounded-md hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
													</svg>
												</button>
											</Tooltip>
											<Tooltip text="Delete" position="top">
												<button
													onclick={() => { itemToDelete = doc; showDeleteConfirm = true; }}
													class="p-1.5 rounded-md hover:bg-red-50 text-gray-400 hover:text-red-600 transition-colors"
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

					<!-- System Prompt (Advanced) -->
					{#if selectedProfile.system_prompt_template}
						<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
							<div class="px-4 py-3 bg-gray-50 border-b border-gray-200">
								<h3 class="text-sm font-medium text-gray-900">System Prompt (Advanced)</h3>
							</div>
							<div class="p-4">
								<pre class="text-xs text-gray-600 whitespace-pre-wrap font-mono">{selectedProfile.system_prompt_template}</pre>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<!-- Empty State -->
			<div class="flex-1 flex items-center justify-center p-6">
				<div class="text-center max-w-md">
					<div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-purple-100 to-blue-100 flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
						</svg>
					</div>
					<h2 class="text-xl font-semibold text-gray-900 mb-2">Knowledge Base</h2>
					<p class="text-sm text-gray-500 mb-6">
						Create profiles for people, businesses, or projects. Attach documents to organize information that AI can use when chatting.
					</p>
					<div class="flex flex-col sm:flex-row gap-3 justify-center">
						<button
							onclick={() => showNewContext = true}
							class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors text-sm font-medium"
						>
							Create Profile
						</button>
						<button
							onclick={() => createNewDocument()}
							class="px-4 py-2 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors text-sm font-medium"
						>
							New Document
						</button>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Document Editor Panel - Side Mode -->
	{#if documentPanelMode === 'side' && selectedDocument}
		<!-- Resize Handle -->
		<div
			onmousedown={startDocumentResize}
			class="w-1 bg-gray-200 hover:bg-purple-500 cursor-col-resize transition-colors flex-shrink-0 {isResizingDocument ? 'bg-purple-500' : ''}"
			role="separator"
			aria-orientation="vertical"
		></div>

		<div
			class="bg-white border-l border-gray-200 flex flex-col h-full flex-shrink-0 overflow-hidden"
			style="width: {documentPanelWidth}px"
		>
			<!-- Panel Header -->
			<div class="px-4 py-3 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
				<div class="flex items-center gap-3 min-w-0 flex-1">
					<span class="text-xl flex-shrink-0">{selectedDocument.icon || '📄'}</span>
					<input
						type="text"
						bind:value={documentTitle}
						onblur={updateDocumentTitle}
						onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
						class="flex-1 min-w-0 font-medium text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
					/>
				</div>
				<div class="flex items-center gap-1">
					<!-- Save status -->
					<div class="text-xs text-gray-400 mr-2">
						{#if $editor.isDirty}
							<span class="text-amber-500">Unsaved</span>
						{:else if $editor.isSaving}
							<span>Saving...</span>
						{:else if $editor.lastSavedAt}
							<span class="text-green-600">Saved</span>
						{/if}
					</div>

					<!-- Mode switcher -->
					<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
						<button
							onclick={() => documentPanelMode = 'side'}
							class="p-1.5 transition-colors bg-gray-900 text-white"
							title="Side panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'center'}
							class="p-1.5 transition-colors text-gray-500 hover:bg-gray-100"
							title="Center panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'full'}
							class="p-1.5 transition-colors text-gray-500 hover:bg-gray-100"
							title="Full screen"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
							</svg>
						</button>
					</div>

					<!-- Open in full page -->
					<a
						href="/contexts/{selectedDocument.id}{embedSuffix}"
						class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors ml-1"
						title="Open in full page"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>

					<!-- Close button -->
					<button
						onclick={closeDocument}
						class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						title="Close"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Editor Content -->
			{#if loadingDocument}
				<div class="flex-1 flex items-center justify-center">
					<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
				</div>
			{:else}
				<div class="flex-1 overflow-y-auto">
					<div class="max-w-none mx-auto px-6 py-8">
						<!-- Blocks -->
						<div class="blocks-container" role="textbox" tabindex="-1">
							{#each $editor.blocks as block, index (block.id)}
								<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
							{/each}
						</div>

						<!-- Click area to add new blocks -->
						<button
							onclick={addNewBlockAtEnd}
							class="w-full min-h-24 mt-4 text-left cursor-text group"
						>
							<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
								Click to add a block, or press / for commands
							</span>
						</button>
					</div>
				</div>

				<!-- Status Bar -->
				<div class="px-4 py-2 border-t border-gray-100 flex items-center justify-between text-xs text-gray-400 bg-gray-50/50">
					<div class="flex items-center gap-4">
						<span>{$wordCount} words</span>
						<span>{$editor.blocks.length} blocks</span>
					</div>
					<button onclick={saveDocument} class="hover:text-gray-600" disabled={!$editor.isDirty}>
						Save now
					</button>
				</div>
			{/if}

			<!-- Slash Command Menu -->
			{#if $editor.showSlashMenu && $editor.slashMenuPosition}
				<BlockMenu />
			{/if}
		</div>
	{/if}
</div>

<!-- Document Editor Panel - Center Mode -->
{#if documentPanelMode === 'center' && selectedDocument}
	<div class="fixed inset-0 bg-black/30 z-40 flex items-center justify-center p-8" onclick={(e) => { if (e.target === e.currentTarget) closeDocument(); }}>
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-4xl h-full max-h-[90vh] flex flex-col overflow-hidden">
			<!-- Panel Header -->
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<div class="flex items-center gap-3 min-w-0 flex-1">
					<span class="text-2xl flex-shrink-0">{selectedDocument.icon || '📄'}</span>
					<input
						type="text"
						bind:value={documentTitle}
						onblur={updateDocumentTitle}
						onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
						class="flex-1 min-w-0 text-lg font-semibold text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
					/>
				</div>
				<div class="flex items-center gap-2">
					<!-- Save status -->
					<div class="text-sm text-gray-400 mr-2">
						{#if $editor.isDirty}
							<span class="text-amber-500">Unsaved</span>
						{:else if $editor.isSaving}
							<span>Saving...</span>
						{:else if $editor.lastSavedAt}
							<span class="text-green-600">Saved</span>
						{/if}
					</div>

					<!-- Mode switcher -->
					<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
						<button
							onclick={() => documentPanelMode = 'side'}
							class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
							title="Side panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'center'}
							class="p-2 transition-colors bg-gray-900 text-white"
							title="Center panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'full'}
							class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
							title="Full screen"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
							</svg>
						</button>
					</div>

					<a
						href="/contexts/{selectedDocument.id}{embedSuffix}"
						class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						title="Open in full page"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>

					<button
						onclick={closeDocument}
						class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						title="Close"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Editor Content -->
			{#if loadingDocument}
				<div class="flex-1 flex items-center justify-center">
					<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
				</div>
			{:else}
				<div class="flex-1 overflow-y-auto">
					<div class="max-w-3xl mx-auto px-8 py-12">
						<!-- Blocks -->
						<div class="blocks-container" role="textbox" tabindex="-1">
							{#each $editor.blocks as block, index (block.id)}
								<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
							{/each}
						</div>

						<!-- Click area to add new blocks -->
						<button
							onclick={addNewBlockAtEnd}
							class="w-full min-h-24 mt-4 text-left cursor-text group"
						>
							<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
								Click to add a block, or press / for commands
							</span>
						</button>
					</div>
				</div>

				<!-- Status Bar -->
				<div class="px-6 py-3 border-t border-gray-100 flex items-center justify-between text-sm text-gray-400">
					<div class="flex items-center gap-4">
						<span>{$wordCount} words</span>
						<span>{$editor.blocks.length} blocks</span>
					</div>
					<button onclick={saveDocument} class="hover:text-gray-600" disabled={!$editor.isDirty}>
						Save now
					</button>
				</div>
			{/if}

			<!-- Slash Command Menu -->
			{#if $editor.showSlashMenu && $editor.slashMenuPosition}
				<BlockMenu />
			{/if}
		</div>
	</div>
{/if}

<!-- Document Editor Panel - Full Screen Mode -->
{#if documentPanelMode === 'full' && selectedDocument}
	<div class="fixed inset-0 bg-white z-50 flex flex-col">
		<!-- Panel Header -->
		<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between bg-white">
			<div class="flex items-center gap-3">
				<button
					onclick={closeDocument}
					class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					title="Back"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<span class="text-gray-300">|</span>
				<span class="text-2xl">{selectedDocument.icon || '📄'}</span>
				<input
					type="text"
					bind:value={documentTitle}
					onblur={updateDocumentTitle}
					onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
					class="text-xl font-semibold text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
				/>
			</div>
			<div class="flex items-center gap-2">
				<!-- Save status -->
				<div class="text-sm text-gray-400 mr-4">
					{#if $editor.isDirty}
						<span class="text-amber-500">Unsaved changes</span>
					{:else if $editor.isSaving}
						<span>Saving...</span>
					{:else if $editor.lastSavedAt}
						<span class="text-green-600">All changes saved</span>
					{/if}
				</div>

				<!-- Mode switcher -->
				<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
					<button
						onclick={() => documentPanelMode = 'side'}
						class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
						title="Side panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'center'}
						class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
						title="Center panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'full'}
						class="p-2 transition-colors bg-gray-900 text-white"
						title="Full screen"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
						</svg>
					</button>
				</div>

				<a
					href="/contexts/{selectedDocument.id}{embedSuffix}"
					class="btn btn-secondary text-sm ml-2"
					title="Open in dedicated page"
				>
					Open Full Editor
				</a>

				<button
					onclick={closeDocument}
					class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors ml-2"
					title="Exit full screen"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Editor Content -->
		{#if loadingDocument}
			<div class="flex-1 flex items-center justify-center">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else}
			<div class="flex-1 overflow-y-auto bg-gray-50/50">
				<div class="max-w-3xl mx-auto px-8 py-12 bg-white min-h-full shadow-sm">
					<!-- Blocks -->
					<div class="blocks-container" role="textbox" tabindex="-1">
						{#each $editor.blocks as block, index (block.id)}
							<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
						{/each}
					</div>

					<!-- Click area to add new blocks -->
					<button
						onclick={addNewBlockAtEnd}
						class="w-full min-h-32 mt-4 text-left cursor-text group"
					>
						<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
							Click to add a block, or press / for commands
						</span>
					</button>
				</div>
			</div>

			<!-- Status Bar -->
			<div class="px-6 py-3 border-t border-gray-200 flex items-center justify-between text-sm text-gray-500 bg-white">
				<div class="flex items-center gap-6">
					<span>{$wordCount} words</span>
					<span>{$editor.blocks.length} blocks</span>
				</div>
				<div class="flex items-center gap-4">
					<button onclick={saveDocument} class="text-purple-600 hover:text-purple-700 font-medium" disabled={!$editor.isDirty}>
						Save now
					</button>
				</div>
			</div>
		{/if}

		<!-- Slash Command Menu -->
		{#if $editor.showSlashMenu && $editor.slashMenuPosition}
			<BlockMenu />
		{/if}
	</div>
{/if}

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
					<button type="button" onclick={() => { showNewContext = false; showAdvancedProfile = false; }} class="flex-1 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
						Cancel
					</button>
					<button type="submit" class="flex-1 px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors">
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
						class="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors text-left"
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
				<button type="button" onclick={() => { showAssignModal = false; documentToAssign = null; }} class="w-full px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
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
					<Dialog.Title class="text-lg font-semibold text-gray-900">Delete {itemToDelete?.type === 'document' ? 'Document' : 'Profile'}?</Dialog.Title>
					<Dialog.Description class="text-sm text-gray-500">
						This cannot be undone.
					</Dialog.Description>
				</div>
			</div>

			{#if itemToDelete}
				<div class="bg-gray-50 rounded-lg p-3 mb-4">
					<div class="flex items-center gap-2">
						<span class="text-lg">{itemToDelete.icon || (itemToDelete.type === 'document' ? '📄' : '👤')}</span>
						<span class="font-medium text-gray-900">{itemToDelete.name}</span>
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
					Delete
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
