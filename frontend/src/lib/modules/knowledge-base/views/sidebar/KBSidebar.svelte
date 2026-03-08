<script lang="ts">
	/**
	 * Pages Sidebar - BusinessOS Style
	 * Modern document-centric sidebar with knowledge graph integration
	 */
	import { sidebarStore, activeDocumentStore, favoriteDocuments, documentTree } from '../../stores/documents';
	import { createDocument, deleteDocument, duplicateDocument, toggleFavorite, fetchProfiles, createProfile, type ProfileType, defaultProfileIcons } from '../../services/documents.service';
	import { ScrollArea, Separator, Tooltip, Modal } from '$lib/ui';
	import { Search, Plus, ChevronLeft, ChevronRight, Star, Clock, FileText, Trash2, Network, Globe, Users, Building2, FolderKanban, User, UserPlus, X } from 'lucide-svelte';
	import SettingsPanel from './SettingsPanel.svelte';
	import SidebarHeader from './SidebarHeader.svelte';
	import SidebarSection from './SidebarSection.svelte';
	import SidebarTreeItem from './SidebarTreeItem.svelte';
	import RecursiveTreeItem from './RecursiveTreeItem.svelte';
	import type { SidebarView, TreeNode, DocumentMeta } from '../../entities/types';

	interface Props {
		onNewDocument?: () => void;
		onOpenDocument?: (id: string) => void;
		onOpenSearch?: () => void;
	}

	let { onNewDocument, onOpenDocument, onOpenSearch }: Props = $props();

	// Sidebar state
	let sidebarWidth = $derived($sidebarStore.width);
	let isCollapsed = $derived($sidebarStore.collapsed);
	let currentView = $derived($sidebarStore.view);

	// Document data
	let favorites = $derived($favoriteDocuments);
	let tree = $derived($documentTree);
	let activeId = $derived($activeDocumentStore.id);

	// Profile data
	let profiles = $state<DocumentMeta[]>([]);
	let profilesLoading = $state(false);

	// Settings modal
	let showSettings = $state(false);

	function handleOpenSettings() {
		showSettings = true;
	}

	function handleCloseSettings() {
		showSettings = false;
	}

	// Check if current view is a profile view
	const isProfileView = $derived(
		currentView === 'profiles' ||
		currentView === 'profiles-person' ||
		currentView === 'profiles-business' ||
		currentView === 'profiles-project'
	);

	// Fetch profiles when switching to profile views
	$effect(() => {
		if (isProfileView) {
			loadProfiles();
		}
	});

	async function loadProfiles() {
		profilesLoading = true;
		try {
			let type: ProfileType | undefined;
			if (currentView === 'profiles-person') type = 'person';
			else if (currentView === 'profiles-business') type = 'business';
			else if (currentView === 'profiles-project') type = 'project';

			profiles = await fetchProfiles(type);
		} catch (error) {
			console.error('Failed to load profiles:', error);
		} finally {
			profilesLoading = false;
		}
	}

	// Document view options
	const documentViews: { id: SidebarView; label: string; icon: typeof FileText }[] = [
		{ id: 'all', label: 'All Pages', icon: FileText },
		{ id: 'favorites', label: 'Favorites', icon: Star },
		{ id: 'recent', label: 'Recent', icon: Clock },
		{ id: 'graph', label: 'Graph View', icon: Network },
		{ id: 'knowledge-graph', label: 'Knowledge Graph', icon: Globe },
		{ id: 'trash', label: 'Trash', icon: Trash2 }
	];

	// Context Profile view options
	const profileViews: { id: SidebarView; label: string; icon: typeof Users }[] = [
		{ id: 'profiles', label: 'All Profiles', icon: Users },
		{ id: 'profiles-person', label: 'People', icon: User },
		{ id: 'profiles-business', label: 'Businesses', icon: Building2 },
		{ id: 'profiles-project', label: 'Projects', icon: FolderKanban }
	];

	// Combined for collapsed view
	const viewOptions = [...documentViews, ...profileViews];

	function handleViewChange(view: SidebarView) {
		sidebarStore.setView(view);
	}

	function handleToggleCollapse() {
		sidebarStore.toggleCollapsed();
	}

	function handleTreeItemClick(node: TreeNode) {
		onOpenDocument?.(node.id);
	}

	async function handleAddChild(node: TreeNode) {
		try {
			const newDoc = await createDocument({
				title: '',
				type: 'document',
				parent_id: node.id
			});
			onOpenDocument?.(newDoc.id);
		} catch (error) {
			console.error('Failed to create subpage:', error);
		}
	}

	async function handleDeleteNode(node: TreeNode) {
		if (!confirm(`Delete "${node.document.title || 'Untitled'}"?`)) return;
		try {
			await deleteDocument(node.id);
		} catch (error) {
			console.error('Failed to delete document:', error);
		}
	}

	async function handleDuplicateNode(node: TreeNode) {
		try {
			const newDoc = await duplicateDocument(node.id);
			onOpenDocument?.(newDoc.id);
		} catch (error) {
			console.error('Failed to duplicate document:', error);
		}
	}

	async function handleToggleFavoriteNode(node: TreeNode) {
		try {
			await toggleFavorite(node.id);
		} catch (error) {
			console.error('Failed to toggle favorite:', error);
		}
	}

	// Handlers for favorites section
	function handleFavoriteSelect(doc: DocumentMeta) {
		onOpenDocument?.(doc.id);
	}

	async function handleFavoriteAddChild(doc: DocumentMeta) {
		try {
			const newDoc = await createDocument({
				title: '',
				type: 'document',
				parent_id: doc.id
			});
			onOpenDocument?.(newDoc.id);
		} catch (error) {
			console.error('Failed to create subpage:', error);
		}
	}

	async function handleFavoriteDelete(doc: DocumentMeta) {
		if (!confirm(`Delete "${doc.title || 'Untitled'}"?`)) return;
		try {
			await deleteDocument(doc.id);
		} catch (error) {
			console.error('Failed to delete document:', error);
		}
	}

	async function handleFavoriteDuplicate(doc: DocumentMeta) {
		try {
			const newDoc = await duplicateDocument(doc.id);
			onOpenDocument?.(newDoc.id);
		} catch (error) {
			console.error('Failed to duplicate document:', error);
		}
	}

	async function handleFavoriteToggle(doc: DocumentMeta) {
		try {
			await toggleFavorite(doc.id);
		} catch (error) {
			console.error('Failed to toggle favorite:', error);
		}
	}

	// Profile handlers
	async function handleNewProfile() {
		let type: ProfileType = 'person';
		let defaultName = 'New Person';
		let defaultIcon = defaultProfileIcons.person;

		if (currentView === 'profiles-business') {
			type = 'business';
			defaultName = 'New Business';
			defaultIcon = defaultProfileIcons.business;
		} else if (currentView === 'profiles-project') {
			type = 'project';
			defaultName = 'New Project';
			defaultIcon = defaultProfileIcons.project;
		}

		try {
			const newProfile = await createProfile({
				name: defaultName,
				type,
				icon: defaultIcon
			});
			onOpenDocument?.(newProfile.id);
			loadProfiles(); // Refresh the list
		} catch (error) {
			console.error('Failed to create profile:', error);
		}
	}

	function handleProfileSelect(doc: DocumentMeta) {
		onOpenDocument?.(doc.id);
	}

	async function handleProfileDelete(doc: DocumentMeta) {
		if (!confirm(`Delete "${doc.title || 'Untitled'}"?`)) return;
		try {
			await deleteDocument(doc.id);
			loadProfiles(); // Refresh the list
		} catch (error) {
			console.error('Failed to delete profile:', error);
		}
	}

	async function handleProfileToggleFavorite(doc: DocumentMeta) {
		try {
			await toggleFavorite(doc.id);
			loadProfiles(); // Refresh the list
		} catch (error) {
			console.error('Failed to toggle favorite:', error);
		}
	}

	// Get section title for profile views
	function getProfileSectionTitle(): string {
		switch (currentView) {
			case 'profiles-person': return 'People';
			case 'profiles-business': return 'Businesses';
			case 'profiles-project': return 'Projects';
			default: return 'All Profiles';
		}
	}

	// Resize handling
	let isResizing = $state(false);
	let startX = $state(0);
	let startWidth = $state(0);

	function handleResizeStart(e: MouseEvent) {
		isResizing = true;
		startX = e.clientX;
		startWidth = sidebarWidth;
		document.addEventListener('mousemove', handleResizeMove);
		document.addEventListener('mouseup', handleResizeEnd);
	}

	function handleResizeMove(e: MouseEvent) {
		if (!isResizing) return;
		const delta = e.clientX - startX;
		sidebarStore.setWidth(startWidth + delta);
	}

	function handleResizeEnd() {
		isResizing = false;
		document.removeEventListener('mousemove', handleResizeMove);
		document.removeEventListener('mouseup', handleResizeEnd);
	}
</script>

<aside
	class="bos-sidebar"
	class:bos-sidebar--collapsed={isCollapsed}
	style:width={isCollapsed ? '48px' : `${sidebarWidth}px`}
>
	{#if !isCollapsed}
		<SidebarHeader
			{onNewDocument}
			{onOpenSearch}
			onOpenSettings={handleOpenSettings}
		/>

		<nav class="bos-sidebar__nav">
			{#each viewOptions as option}
				<button
					class="bos-sidebar__nav-item"
					class:bos-sidebar__nav-item--active={currentView === option.id}
					onclick={() => handleViewChange(option.id)}
				>
					<span class="bos-sidebar__nav-icon">
						<svelte:component this={option.icon} />
					</span>
					<span class="bos-sidebar__nav-label">{option.label}</span>
				</button>
			{/each}
		</nav>

		<div class="bos-sidebar__divider"></div>

		<ScrollArea class="bos-sidebar__content">
			{#if isProfileView}
				<!-- Profile Views -->
				<div class="bos-sidebar__profile-header">
					<span class="bos-sidebar__profile-title">{getProfileSectionTitle()}</span>
					<Tooltip content="New {currentView === 'profiles-business' ? 'Business' : currentView === 'profiles-project' ? 'Project' : 'Person'}" side="top">
						<button class="bos-sidebar__profile-add" onclick={handleNewProfile}>
							<UserPlus />
						</button>
					</Tooltip>
				</div>

				{#if profilesLoading}
					<div class="bos-sidebar__loading">Loading...</div>
				{:else if profiles.length === 0}
					<div class="bos-sidebar__empty">
						<p>No {getProfileSectionTitle().toLowerCase()} yet</p>
						<button class="bos-sidebar__empty-btn" onclick={handleNewProfile}>
							<Plus class="h-4 w-4" />
							Create {currentView === 'profiles-business' ? 'Business' : currentView === 'profiles-project' ? 'Project' : 'Person'}
						</button>
					</div>
				{:else}
					<SidebarSection title="" collapsible={false} defaultExpanded>
						{#each profiles as doc}
							<SidebarTreeItem
								document={doc}
								depth={0}
								isActive={activeId === doc.id}
								onSelect={() => handleProfileSelect(doc)}
								onDelete={() => handleProfileDelete(doc)}
								onToggleFavorite={() => handleProfileToggleFavorite(doc)}
							/>
						{/each}
					</SidebarSection>
				{/if}
			{:else}
				<!-- Document Views -->
				{#if favorites.length > 0}
					<SidebarSection title="Favorites" collapsible defaultExpanded>
						{#each favorites as doc}
							<SidebarTreeItem
								document={doc}
								depth={0}
								isActive={activeId === doc.id}
								onSelect={() => handleFavoriteSelect(doc)}
								onAddChild={() => handleFavoriteAddChild(doc)}
								onDelete={() => handleFavoriteDelete(doc)}
								onDuplicate={() => handleFavoriteDuplicate(doc)}
								onToggleFavorite={() => handleFavoriteToggle(doc)}
							/>
						{/each}
					</SidebarSection>
				{/if}

				<SidebarSection title="Pages" collapsible defaultExpanded>
					{#each tree as node (node.id)}
						<RecursiveTreeItem
							{node}
							{activeId}
							onSelect={handleTreeItemClick}
							onAddChild={handleAddChild}
							onDelete={handleDeleteNode}
							onDuplicate={handleDuplicateNode}
							onToggleFavorite={handleToggleFavoriteNode}
						/>
					{/each}
				</SidebarSection>
			{/if}
		</ScrollArea>

		<!-- Resize handle -->
		<div
			class="bos-sidebar__resize"
			onmousedown={handleResizeStart}
			role="separator"
			aria-orientation="vertical"
			tabindex={-1}
		></div>
	{:else}
		<!-- Collapsed state -->
		<div class="bos-sidebar__collapsed">
			<Tooltip content="Search" side="right">
				<button class="bos-sidebar__action" onclick={onOpenSearch}>
					<Search />
				</button>
			</Tooltip>
			<Tooltip content="New Page" side="right">
				<button class="bos-sidebar__action" onclick={onNewDocument}>
					<Plus />
				</button>
			</Tooltip>
			<div class="bos-sidebar__divider"></div>
			{#each viewOptions as option}
				<Tooltip content={option.label} side="right">
					<button
						class="bos-sidebar__action"
						class:bos-sidebar__action--active={currentView === option.id}
						onclick={() => handleViewChange(option.id)}
					>
						<svelte:component this={option.icon} />
					</button>
				</Tooltip>
			{/each}
		</div>
	{/if}

	<!-- Toggle button -->
	<button
		class="bos-sidebar__toggle"
		onclick={handleToggleCollapse}
		aria-label={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
	>
		{#if isCollapsed}
			<ChevronRight />
		{:else}
			<ChevronLeft />
		{/if}
	</button>
</aside>

<!-- Settings Modal -->
<Modal
	bind:open={showSettings}
	title="Pages Settings"
	description="Configure your Pages experience"
	size="lg"
>
	<SettingsPanel onClose={handleCloseSettings} />
</Modal>

<style>
	/* BusinessOS Sidebar */
	.bos-sidebar {
		position: relative;
		display: flex;
		flex-direction: column;
		height: 100%;
		background-color: var(--bos-v2-layer-background-secondary, #f4f4f5);
		border-right: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		transition: width 0.2s ease;
	}

	.bos-sidebar--collapsed {
		width: 48px;
	}

	/* Navigation */
	.bos-sidebar__nav {
		display: flex;
		flex-direction: column;
		padding: 4px 8px;
		gap: 2px;
	}

	.bos-sidebar__nav-item {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		min-height: 30px;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: var(--bos-font-sm, 14px);
		color: var(--bos-v2-text-primary, #121212);
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
		transition: background-color 0.15s;
	}

	.bos-sidebar__nav-item:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-sidebar__nav-item--active {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
		font-weight: 500;
	}

	.bos-sidebar__nav-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		color: var(--bos-v2-icon-primary, #77757d);
	}

	.bos-sidebar__nav-icon :global(svg) {
		width: 16px;
		height: 16px;
	}

	.bos-sidebar__nav-label {
		flex: 1;
	}

	/* Divider */
	.bos-sidebar__divider {
		height: 1px;
		margin: 8px;
		background-color: var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
	}

	/* Content area */
	.bos-sidebar__content {
		flex: 1;
		overflow: hidden;
		padding: 0 8px 8px;
	}

	/* Profile header */
	.bos-sidebar__profile-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 4px;
		margin-bottom: 4px;
	}

	.bos-sidebar__profile-title {
		font-size: var(--bos-font-sm, 14px);
		font-weight: 600;
		color: var(--bos-v2-text-primary, #121212);
	}

	.bos-sidebar__profile-add {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 4px;
		background: transparent;
		border: none;
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.bos-sidebar__profile-add:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-sidebar__profile-add :global(svg) {
		width: 16px;
		height: 16px;
	}

	/* Loading and empty states */
	.bos-sidebar__loading {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem 1rem;
		color: var(--bos-v2-text-tertiary, #77757d);
		font-size: var(--bos-font-sm, 14px);
	}

	.bos-sidebar__empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 2rem 1rem;
		text-align: center;
	}

	.bos-sidebar__empty p {
		color: var(--bos-v2-text-tertiary, #77757d);
		font-size: var(--bos-font-sm, 14px);
		margin: 0;
	}

	.bos-sidebar__empty-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
		border: none;
		color: var(--bos-v2-text-primary, #121212);
		font-size: var(--bos-font-sm, 14px);
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.bos-sidebar__empty-btn:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.08));
	}

	/* Resize handle */
	.bos-sidebar__resize {
		position: absolute;
		top: 0;
		right: -2px;
		width: 4px;
		height: 100%;
		cursor: ew-resize;
		background: transparent;
		transition: background-color 0.15s;
	}

	.bos-sidebar__resize:hover,
	.bos-sidebar__resize:active {
		background-color: var(--bos-v2-layer-insideBorder-primaryBorder, #1e96eb);
	}

	/* Collapsed state */
	.bos-sidebar__collapsed {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 8px 4px;
		gap: 4px;
	}

	.bos-sidebar__action {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 4px;
		background: transparent;
		border: none;
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.bos-sidebar__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-sidebar__action--active {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
		color: var(--bos-v2-icon-activated, #1e96eb);
	}

	.bos-sidebar__action :global(svg) {
		width: 16px;
		height: 16px;
	}

	/* Toggle button */
	.bos-sidebar__toggle {
		position: absolute;
		top: 50%;
		right: -12px;
		transform: translateY(-50%);
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
		border: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		opacity: 0;
		transition: opacity 0.15s, background-color 0.15s;
		z-index: 10;
		box-shadow: var(--bos-shadow-1);
	}

	.bos-sidebar:hover .bos-sidebar__toggle {
		opacity: 1;
	}

	.bos-sidebar__toggle:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-sidebar__toggle :global(svg) {
		width: 14px;
		height: 14px;
	}

	/* Dark mode */
	:global(.dark) .bos-sidebar {
		background-color: var(--bos-v2-layer-background-secondary, #2c2c2c);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .bos-sidebar__nav-item {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .bos-sidebar__nav-item:hover,
	:global(.dark) .bos-sidebar__nav-item--active {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .bos-sidebar__nav-icon {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar__divider {
		background-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .bos-sidebar__action {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .bos-sidebar__toggle {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar__profile-title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .bos-sidebar__profile-add {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar__profile-add:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .bos-sidebar__empty p {
		color: var(--bos-v2-text-tertiary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar__empty-btn {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .bos-sidebar__empty-btn:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.12));
	}
</style>
