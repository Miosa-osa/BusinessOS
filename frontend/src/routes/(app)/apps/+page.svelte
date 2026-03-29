<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { currentWorkspaceId } from '$lib/stores/workspaces';
	import { generatedAppsStore } from '$lib/stores/generatedAppsStore';
	import type { App, AppStatus } from '$lib/types/apps';
	import { APP_TEMPLATES } from '$lib/types/apps';
	import { AppCard, AppEmptyState, AppFilterDropdown, AppCommandPalette, AppContextMenu } from '$lib/components/apps';
	import { CreateAppModal as OSACreateAppModal } from '$lib/components/osa';
	import { Pagination } from '$lib/components/ui';
	import { LayoutGrid, List, Users, CheckSquare, Wallet, Kanban, Search, Command, AlertCircle, X, BookOpen, BarChart3, Clock, FileText } from 'lucide-svelte';

	// Fetch generated modules from the API (paginated)
	onMount(() => {
		generatedAppsStore.fetchApps();
	});

	// Reactive snapshot of the generated apps store
	const generatedState = $derived($generatedAppsStore);
	const apiPagination = $derived(generatedState.pagination);

	function handleApiPageChange(newPage: number) {
		generatedAppsStore.goToPage(newPage);
	}

	function handleApiPageSizeChange(newSize: number) {
		generatedAppsStore.setPageSize(newSize);
	}

	// Command palette state
	let isCommandPaletteOpen = $state(false);

	// Create app modal state (pure AI generation)
	let isCreateModalOpen = $state(false);

	// Context menu state
	let contextMenu = $state<{ app: App; x: number; y: number } | null>(null);

	// Error toast state
	let errorToast = $state<{ message: string; details?: string } | null>(null);
	let errorTimeout: ReturnType<typeof setTimeout> | null = null;

	function showError(message: string, details?: string) {
		if (errorTimeout) clearTimeout(errorTimeout);
		errorToast = { message, details };
		errorTimeout = setTimeout(() => {
			errorToast = null;
		}, 5000);
	}

	function dismissError() {
		if (errorTimeout) clearTimeout(errorTimeout);
		errorToast = null;
	}

	// View mode
	let viewMode = $state<'grid' | 'list'>('grid');

	// Local pin state (toggleable, not persisted to backend yet)
	let pinnedIds = $state<Set<string>>(new Set());

	// Get starter apps from onboarding store and convert to App type
	const starterApps = $derived($onboardingStore.userData.starterApps || []);

	// Extended App type with lastOpened
	interface AppWithActivity extends App {
		lastOpened?: string;
	}

	// Convert starter apps to App format
	// Note: Demo/mock data removed - use empty state component when no apps exist
	const allApps = $derived<AppWithActivity[]>(
		starterApps.map((app, index) => ({
			id: app.id,
			name: app.title,
			description: app.description,
			icon: getIconFromTitle(app.title),
			status: 'active' as AppStatus,
			version: 1,
			versionCount: 1,
			isPinned: pinnedIds.has(app.id),
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString(),
			lastOpened: index === 0 ? new Date(Date.now() - 1000 * 60 * 30).toISOString() : undefined
		}))
	);

	// Recently opened apps (last 7 days, sorted by lastOpened)
	const recentlyOpened = $derived(
		allApps
			.filter((app) => app.lastOpened)
			.sort((a, b) => new Date(b.lastOpened!).getTime() - new Date(a.lastOpened!).getTime())
			.slice(0, 4)
	);

	// Format relative time
	function formatRelativeTime(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / (1000 * 60));
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays === 1) return 'Yesterday';
		if (diffDays < 7) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	}

	// Get icon name from app title
	function getIconFromTitle(title: string): string {
		const lower = title.toLowerCase();
		if (lower.includes('crm') || lower.includes('client')) return 'Users';
		if (lower.includes('project') || lower.includes('kanban') || lower.includes('tracker')) return 'Kanban';
		if (lower.includes('invoice') || lower.includes('finance')) return 'Wallet';
		if (lower.includes('task')) return 'CheckSquare';
		if (lower.includes('calendar')) return 'Calendar';
		if (lower.includes('analytics') || lower.includes('report') || lower.includes('dashboard')) return 'BarChart3';
		if (lower.includes('time')) return 'Clock';
		if (lower.includes('document') || lower.includes('doc')) return 'FileText';
		if (lower.includes('journal')) return 'BookOpen';
		return 'Layers';
	}

	// Get gradient class from title (for recently opened cards)
	function getGradientFromTitle(title: string): string {
		const lower = title.toLowerCase();
		if (lower.includes('crm') || lower.includes('client')) return 'from-violet-500 to-purple-600';
		if (lower.includes('project') || lower.includes('tracker')) return 'from-emerald-500 to-teal-600';
		if (lower.includes('invoice') || lower.includes('finance')) return 'from-green-500 to-emerald-600';
		if (lower.includes('task')) return 'from-blue-500 to-blue-600';
		if (lower.includes('calendar')) return 'from-pink-500 to-rose-500';
		if (lower.includes('analytics') || lower.includes('dashboard')) return 'from-orange-500 to-red-500';
		if (lower.includes('journal')) return 'from-amber-400 to-orange-500';
		return 'from-gray-500 to-gray-600';
	}

	// Get icon component from title
	function getIconComponent(title: string) {
		const lower = title.toLowerCase();
		if (lower.includes('crm') || lower.includes('client')) return Users;
		if (lower.includes('project') || lower.includes('tracker')) return Kanban;
		if (lower.includes('invoice') || lower.includes('finance')) return Wallet;
		if (lower.includes('task')) return CheckSquare;
		if (lower.includes('journal')) return BookOpen;
		if (lower.includes('analytics') || lower.includes('dashboard')) return BarChart3;
		return Kanban;
	}

	// Filter state
	let statusFilter = $state<AppStatus | 'all'>('all');

	// Filtered apps (status filter only, search is via command palette)
	const filteredApps = $derived(
		allApps.filter((app) => {
			return statusFilter === 'all' || app.status === statusFilter;
		})
	);

	// Separate pinned and unpinned
	const pinnedApps = $derived(filteredApps.filter((app) => app.isPinned));
	const unpinnedApps = $derived(filteredApps.filter((app) => !app.isPinned));

	// Handlers
	function handleOpenApp(app: App) {
		goto(`/apps/${app.id}`);
	}

	function handleCreateApp() {
		// Check if valid workspace exists
		if (!$currentWorkspaceId) {
			showError(
				'No workspace selected',
				'Please select a workspace before creating an app.'
			);
			return;
		}
		// Open pure AI generation modal (no template needed)
		isCreateModalOpen = true;
	}

	function handleSelectTemplate(templateId: string) {
		// Check if valid workspace exists
		if (!$currentWorkspaceId) {
			showError(
				'No workspace selected',
				'Please select a workspace before creating an app.'
			);
			return;
		}
		// For now, all creation goes through AI generation modal
		// Templates can be used as inspiration/context in the description
		isCreateModalOpen = true;
	}

	function handleCloseCreateModal() {
		isCreateModalOpen = false;
	}

	// Context menu handlers
	function handleContextMenu(app: App, x: number, y: number) {
		contextMenu = { app, x, y };
	}

	function handleCloseContextMenu() {
		contextMenu = null;
	}

	async function handlePinApp(app: App) {
		try {
			const updated = new Set(pinnedIds);
			if (updated.has(app.id)) {
				updated.delete(app.id);
			} else {
				updated.add(app.id);
			}
			pinnedIds = updated;
		} catch (err) {
			showError(
				`Failed to ${app.isPinned ? 'unpin' : 'pin'} app`,
				err instanceof Error ? err.message : 'Please try again later.'
			);
		}
	}

	function handleEditApp(app: App) {
		goto(`/apps/${app.id}/edit`);
	}

	async function handleDuplicateApp(app: App) {
		try {
			const existingApps = $onboardingStore.userData.starterApps || [];
			const original = existingApps.find(a => a.id === app.id);
			if (!original) return;

			const duplicate = {
				...original,
				id: `${original.id}-copy-${Date.now()}`,
				title: `${original.title} (Copy)`,
			};
			onboardingStore.setStarterApps([...existingApps, duplicate]);
		} catch (err) {
			showError(
				'Failed to duplicate app',
				err instanceof Error ? err.message : 'Please try again later.'
			);
		}
	}

	async function handleDeleteApp(app: App) {
		// Confirm before deleting
		if (!confirm(`Are you sure you want to delete "${app.name}"? This action cannot be undone.`)) {
			return;
		}

		try {
			await generatedAppsStore.deleteApp(app.id);
		} catch (err) {
			showError(
				'Failed to delete app',
				err instanceof Error ? err.message : 'Please try again later.'
			);
		}
	}
</script>

<!-- Keyboard shortcut for command palette -->
<svelte:window
	onkeydown={(e) => {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			isCommandPaletteOpen = !isCommandPaletteOpen;
		}
	}}
/>

<svelte:head>
	<title>Modules | Business OS</title>
</svelte:head>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900 overflow-hidden">
	<!-- Header -->
	<header class="flex-shrink-0 px-6 py-5 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
		<div class="flex items-center justify-between mb-4">
			<div>
				<div class="flex items-center gap-3">
					<h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Your Modules</h1>
					<span class="px-2.5 py-1 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 rounded-full">
						{allApps.length}
					</span>
				</div>
				<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Manage and launch your workspace modules</p>
			</div>
		</div>

		<!-- Controls Row -->
		<div class="flex items-center gap-3">
			<!-- Search Trigger (opens command palette) -->
			<button
				onclick={() => (isCommandPaletteOpen = true)}
				class="flex items-center gap-3 w-64 h-10 px-3 bg-gray-100 dark:bg-gray-800 border border-transparent
					rounded-xl text-sm text-gray-500 dark:text-gray-400 transition-all duration-150
					hover:bg-gray-200 dark:hover:bg-gray-700 hover:border-gray-300 dark:hover:border-gray-600"
			>
				<Search class="w-4 h-4 flex-shrink-0" strokeWidth={2} />
				<span class="flex-1 text-left">Search modules...</span>
				<kbd class="hidden sm:flex items-center gap-0.5 px-1.5 py-0.5 text-xs text-gray-400 bg-white dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600">
					<Command class="w-3 h-3" />K
				</kbd>
			</button>
			<AppFilterDropdown value={statusFilter} onChange={(v) => (statusFilter = v)} />

			<!-- View Toggle -->
			<div class="flex items-center bg-gray-100 dark:bg-gray-800 rounded-lg p-1" role="group" aria-label="View mode">
				<button
					onclick={() => (viewMode = 'grid')}
					class="p-1.5 rounded-md transition-all duration-150
						{viewMode === 'grid'
						? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
					aria-label="Grid view"
					aria-pressed={viewMode === 'grid'}
				>
					<LayoutGrid class="w-4 h-4" strokeWidth={2} aria-hidden="true" />
				</button>
				<button
					onclick={() => (viewMode = 'list')}
					class="p-1.5 rounded-md transition-all duration-150
						{viewMode === 'list'
						? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
					aria-label="List view"
					aria-pressed={viewMode === 'list'}
				>
					<List class="w-4 h-4" strokeWidth={2} aria-hidden="true" />
				</button>
			</div>

			<div class="flex-1"></div>

			<button
				onclick={handleCreateApp}
				class="btn-cta"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				<span>Create Module</span>
			</button>
		</div>
	</header>

	<!-- Content -->
	<main class="flex-1 overflow-y-auto">
		{#if allApps.length === 0}
			<!-- Empty State -->
			<AppEmptyState onCreateApp={handleCreateApp} onSelectTemplate={handleSelectTemplate} />
		{:else if filteredApps.length === 0}
			<!-- No Results -->
			<div class="flex flex-col items-center justify-center py-16 px-4">
				<div class="w-16 h-16 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
				</div>
				<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-1">No modules found</h2>
				<p class="text-gray-500 dark:text-gray-400 text-sm">
					Try adjusting your search or filter criteria.
				</p>
			</div>
		{:else}
			<div class="p-6 space-y-8">
				<!-- Recently Opened Section -->
				{#if recentlyOpened.length > 0 && statusFilter === 'all'}
					<section>
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
								Recently Opened
							</h2>
						</div>
						<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
							{#each recentlyOpened as app, i (app.id)}
								<button
									onclick={() => handleOpenApp(app)}
									aria-label="Open {app.name}, last opened {formatRelativeTime(app.lastOpened!)}"
									class="flex items-center gap-3 p-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700
										rounded-xl transition-all duration-150 hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-sm
										text-left group hover:scale-[1.02]"
									in:fade={{ duration: 200, delay: i * 50 }}
								>
									<div class="w-9 h-9 rounded-lg bg-gradient-to-br {getGradientFromTitle(app.name)} flex items-center justify-center text-white flex-shrink-0">
										<svelte:component this={getIconComponent(app.name)} class="w-4 h-4" strokeWidth={2} />
									</div>
									<div class="min-w-0 flex-1">
										<p class="text-sm font-medium text-gray-900 dark:text-white truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
											{app.name}
										</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">
											{formatRelativeTime(app.lastOpened!)}
										</p>
									</div>
								</button>
							{/each}
						</div>
					</section>
				{/if}

				<!-- Pinned Section -->
				{#if pinnedApps.length > 0}
					<section>
						<h2 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-4">
							Pinned
						</h2>
						<div class="{viewMode === 'grid' ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5' : 'flex flex-col gap-3'}">
							{#each pinnedApps as app (app.id)}
								<AppCard {app} onOpen={() => handleOpenApp(app)} onContextMenu={handleContextMenu} />
							{/each}
						</div>
					</section>
				{/if}

				<!-- All Apps Section -->
				<section>
					{#if pinnedApps.length > 0}
						<h2 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-4">
							All Modules
						</h2>
					{/if}
					<div class="{viewMode === 'grid' ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5' : 'flex flex-col gap-3'}">
						{#each unpinnedApps as app (app.id)}
							<AppCard {app} onOpen={() => handleOpenApp(app)} onContextMenu={handleContextMenu} />
						{/each}
					</div>

					<!-- Pagination for API-backed modules -->
					{#if apiPagination.total_items > apiPagination.page_size}
						<Pagination
							page={apiPagination.page}
							pageSize={apiPagination.page_size}
							total={apiPagination.total_items}
							pageSizeOptions={[10, 20, 50]}
							onPageChange={handleApiPageChange}
							onPageSizeChange={handleApiPageSizeChange}
						/>
					{/if}
				</section>

				<!-- Quick Templates Section -->
				{#if allApps.length < 6 && statusFilter === 'all'}
					<section class="pt-4 border-t border-gray-200 dark:border-gray-800">
						<h2 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-4">
							Create from Template
						</h2>
						<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
							{#each APP_TEMPLATES as template}
								<button
									onclick={() => handleSelectTemplate(template.id)}
									class="flex flex-col items-center p-4 border-2 border-dashed border-gray-200 dark:border-gray-700
										rounded-xl transition-all duration-150 hover:border-gray-400 dark:hover:border-gray-500 hover:bg-gray-50 dark:hover:bg-gray-800/50 group"
								>
									<div class="w-10 h-10 rounded-xl bg-gray-100 dark:bg-gray-700 flex items-center justify-center mb-2.5
										text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors">
										{#if template.icon === 'Users'}
											<Users class="w-5 h-5" strokeWidth={1.75} />
										{:else if template.icon === 'CheckSquare'}
											<CheckSquare class="w-5 h-5" strokeWidth={1.75} />
										{:else if template.icon === 'Receipt' || template.icon === 'Wallet'}
											<Wallet class="w-5 h-5" strokeWidth={1.75} />
										{:else}
											<Kanban class="w-5 h-5" strokeWidth={1.75} />
										{/if}
									</div>
									<span class="text-sm font-medium text-gray-900 dark:text-white mb-0.5">
										{template.name}
									</span>
									<span class="text-xs text-gray-500 dark:text-gray-400 text-center">
										{template.description}
									</span>
								</button>
							{/each}
						</div>
					</section>
				{/if}
			</div>
		{/if}
	</main>
</div>

<!-- Command Palette -->
<AppCommandPalette
	apps={allApps}
	recentApps={recentlyOpened}
	isOpen={isCommandPaletteOpen}
	onClose={() => (isCommandPaletteOpen = false)}
	onSelect={handleOpenApp}
	onCreateApp={handleCreateApp}
/>

<!-- Context Menu -->
{#if contextMenu}
	<AppContextMenu
		app={contextMenu.app}
		x={contextMenu.x}
		y={contextMenu.y}
		onClose={handleCloseContextMenu}
		onOpen={handleOpenApp}
		onEdit={handleEditApp}
		onPin={handlePinApp}
		onDuplicate={handleDuplicateApp}
		onDelete={handleDeleteApp}
	/>
{/if}

<!-- Create Module Modal -->
<OSACreateAppModal
	workspaceId={$currentWorkspaceId ?? ''}
	bind:open={isCreateModalOpen}
/>

<!-- Error Toast -->
{#if errorToast}
	<div
		class="fixed bottom-6 right-6 z-50 max-w-sm"
		transition:fade={{ duration: 200 }}
		role="alert"
		aria-live="assertive"
	>
		<div class="flex items-start gap-3 p-4 bg-red-50 dark:bg-red-900/90 border border-red-200 dark:border-red-800 rounded-xl shadow-lg">
			<AlertCircle class="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" strokeWidth={2} aria-hidden="true" />
			<div class="flex-1 min-w-0">
				<p class="text-sm font-medium text-red-800 dark:text-red-200">{errorToast.message}</p>
				{#if errorToast.details}
					<p class="text-sm text-red-600 dark:text-red-300 mt-0.5">{errorToast.details}</p>
				{/if}
			</div>
			<button
				onclick={dismissError}
				class="btn-pill btn-pill-ghost btn-pill-icon"
				aria-label="Dismiss error"
			>
				<X class="w-4 h-4" strokeWidth={2} aria-hidden="true" />
			</button>
		</div>
	</div>
{/if}
