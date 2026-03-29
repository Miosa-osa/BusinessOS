<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { useSession } from '$lib/auth-client';
	import { Separator } from 'bits-ui';
	import { browser } from '$app/environment';
	import { isElectron as checkElectron, isMacOS } from '$lib/utils/platform';
	import { desktopSettings } from '$lib/stores/desktopStore';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { getBackendUrl, initCSRF } from '$lib/api/base';
	import { WorkspaceSwitcher } from '$lib/components/workspace';
	import { loadSavedWorkspace } from '$lib/stores/workspaces';
	import { notificationStore } from '$lib/stores/notifications';
	import { initializePush } from '$lib/services/pushService';
	import { getInstallations, getModule } from '$lib/api/modules';
	import { desktop3dStore, BUILTIN_MODULES, DYNAMIC_MODULE_COLORS } from '$lib/stores/desktop3dStore';
	import { windowStore } from '$lib/stores/windowStore';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { ChevronsLeft, Monitor, ChevronRight, ChevronDown } from 'lucide-svelte';

	// Projects state for dropdown
	let projects = $state<Array<{id: string, name: string, status: string}>>([]);
	let showProjectsDropdown = $state(false);

	// Load projects for sidebar dropdown
	async function loadProjects() {
		try {
			const data = await api.getProjects('active');
			projects = data.slice(0, 5); // Show top 5 active projects
		} catch (e) {
			console.error('Failed to load projects:', e);
		}
	}

	// Sync installed dynamic modules to 3D desktop + dock on startup
	async function syncDynamicModules() {
		try {
			const result = await getInstallations();
			const installations = result.installations ?? [];
			if (!installations.length) return;

			const builtinSet = new Set<string>(BUILTIN_MODULES);
			const dynamicInstalls = installations.filter(
				(inst) => inst.is_active && !builtinSet.has(inst.module_id)
			);
			if (!dynamicInstalls.length) return;

			// Fetch all module details in parallel — avoids N+1 sequential requests
			const results = await Promise.allSettled(
				dynamicInstalls.map((inst) => getModule(inst.module_id))
			);

			for (let idx = 0; idx < results.length; idx++) {
				const result = results[idx];
				const inst = dynamicInstalls[idx];

				if (result.status === 'fulfilled') {
					const mod = result.value;
					desktop3dStore.addModule({
						id: mod.id,
						title: mod.name,
						icon: mod.icon ?? 'box',
						color: DYNAMIC_MODULE_COLORS[mod.category] ?? DYNAMIC_MODULE_COLORS.general,
						category: mod.category,
					});
					windowStore.addToDock(mod.id);
				} else {
					// Module details unavailable — add with fallback
					desktop3dStore.addModule({
						id: inst.module_id,
						title: inst.module_id,
						icon: 'box',
						color: DYNAMIC_MODULE_COLORS.general,
					});
				}
			}
		} catch {
			// Module API not available yet — skip silently
		}
	}

	onMount(async () => {
		// Initialize CSRF token first (required before any state-changing requests)
		await initCSRF();

		// Skip rest of initialization in embed mode - iframes don't need workspace/notification systems
		if (isEmbedMode) return;

		// Fire all independent init tasks in parallel — no waterfall
		await Promise.all([
			loadSavedWorkspace(),
			loadProjects(),
		]);

		// These don't return promises — fire immediately
		notificationStore.initialize();
		initializePush();

		// Sync dynamic modules after workspace is loaded
		syncDynamicModules();
	});

	const APP_VERSION = '0.0.1';

	let { children } = $props();

	const session = useSession();

	// Check if we're in embed mode (used by desktop windows)
	const isEmbedMode = $derived($page.url.searchParams.get('embed') === 'true');

	// No loading screen for app routes - instant load
	let bootComplete = $state(true);

	// Check if running in Electron (for native window styling)
	const inElectron = $derived(browser && checkElectron());
	const onMac = $derived(browser && isMacOS());
	const needsTrafficLightSpace = $derived(inElectron && onMac);

	// Sidebar collapsed state (persisted to localStorage)
	let isCollapsed = $state(false);

	$effect(() => {
		// Load collapsed state from localStorage
		const stored = localStorage.getItem('sidebar-collapsed');
		if (stored !== null) {
			isCollapsed = stored === 'true';
		}
	});

	function toggleSidebar() {
		isCollapsed = !isCollapsed;
		localStorage.setItem('sidebar-collapsed', String(isCollapsed));
	}

	// Auth is handled server-side in +layout.server.ts — no duplicate client-side check needed.
	// The server redirects to /login if the session cookie is invalid, preventing double auth round-trips.

	interface NavItem {
		href: string;
		label: string;
		icon: string;
	}

	interface NavGroup {
		label: string;
		items: NavItem[];
	}

	let collapsedSections = $state<Set<string>>(new Set());

	function toggleSection(label: string) {
		if (collapsedSections.has(label)) {
			collapsedSections.delete(label);
		} else {
			collapsedSections.add(label);
		}
		collapsedSections = new Set(collapsedSections);
	}

	const navGroups: NavGroup[] = [
		{
			label: 'Core',
			items: [
				{
					href: '/dashboard',
					label: 'Dashboard',
					icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
				},
				{
					href: '/modules',
					label: 'Modules',
					icon: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z'
				},
				{
					href: '/chat',
					label: 'Chat',
					icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z'
				},
			]
		},
		{
			label: 'Work',
			items: [
				{
					href: '/tasks',
					label: 'Tasks',
					icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4'
				},
				{
					href: '/projects',
					label: 'Projects',
					icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
				},
				{
					href: '/daily',
					label: 'Daily Log',
					icon: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z'
				},
			]
		},
		{
			label: 'People',
			items: [
				{
					href: '/communication',
					label: 'Communication',
					icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z'
				},
				{
					href: '/team',
					label: 'Team',
					icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z'
				},
				{
					href: '/clients',
					label: 'Clients',
					icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4'
				},
				{
					href: '/crm',
					label: 'CRM',
					icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
				},
			]
		},
		{
			label: 'Tools',
			items: [
				{
					href: '/tables',
					label: 'Tables',
					icon: 'M3 10h18M3 14h18M9 3v18M15 3v18M3 6a2 2 0 012-2h14a2 2 0 012 2v12a2 2 0 01-2 2H5a2 2 0 01-2-2V6z'
				},
				{
					href: '/pages',
					label: 'Pages',
					icon: 'M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z'
				},
				{
					href: '/agents',
					label: 'Agents',
					icon: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z'
				},
				{
					href: '/nodes',
					label: 'Nodes',
					icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 6v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z'
				},
				{
					href: '/code',
					label: 'Code',
					icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4'
				},
			]
		},
		{
			label: 'System',
			items: [
				{
					href: '/usage',
					label: 'Analytics',
					icon: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z'
				},
				{
					href: '/integrations',
					label: 'Integrations',
					icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1'
				},
				{
					href: '/settings',
					label: 'Settings',
					icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z'
				},
			]
		},
	];

</script>

{#if $session.data}
	{#if isEmbedMode}
		<!-- Embed mode: no sidebar, just content -->
		<div class="h-screen w-screen overflow-hidden sb-main-bg">
			{@render children()}
		</div>
	{:else}
	<div class="h-screen flex overflow-hidden p-2 gap-2 sb-canvas">
		<!-- Sidebar -->
		<aside
			class="sb-sidebar flex flex-col flex-shrink-0 transition-all duration-300 ease-in-out rounded-[14px] overflow-hidden {isCollapsed ? (needsTrafficLightSpace ? 'w-20' : 'w-16') : 'w-64'}"
		>
			<!-- Draggable titlebar region for Electron (traffic light area) -->
			{#if needsTrafficLightSpace}
				<div
					class="h-12 flex-shrink-0 drag-region"
					style="-webkit-app-region: drag;"
				>
					<!-- Traffic light spacer - this area is for the macOS window controls -->
				</div>
			{:else}
				<div class="h-4 flex-shrink-0"></div>
			{/if}

			<!-- Header with toggle button -->
			<div class="px-4 pb-2 flex items-center {isCollapsed ? 'justify-center' : 'justify-between'}">
				{#if !isCollapsed}
					<h1 class="text-lg font-semibold sb-title">Business OS</h1>
				{/if}
				<button
					onclick={toggleSidebar}
					class="btn-pill btn-pill-icon btn-pill-ghost btn-pill-sm no-drag flex-shrink-0"
					style="-webkit-app-region: no-drag;"
					title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
				>
<ChevronsLeft
						size={20}
						strokeWidth={2}
						class="transition-transform duration-300 {isCollapsed ? 'rotate-180' : ''}"
					/>
				</button>
			</div>

			<!-- Workspace Switcher -->
			{#if !isCollapsed}
				<div class="px-2 pb-2">
					<WorkspaceSwitcher />
				</div>
			{:else}
				<div class="flex justify-center pb-2">
					<button
						class="sb-workspace-badge"
						title={$currentWorkspace?.name || 'Workspace'}
						onclick={toggleSidebar}
						aria-label="Expand sidebar to switch workspace"
					>
						{$currentWorkspace?.name?.charAt(0).toUpperCase() || 'W'}
					</button>
				</div>
			{/if}

			<!-- Window Desktop Button -->
			<div class="px-2 pb-2">
				<a
					href="/window"
					class="sb-desktop-btn flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200
						{isCollapsed ? 'justify-center' : ''}"
					title={isCollapsed ? 'Window Desktop' : ''}
				>
<Monitor size={20} strokeWidth={2} class="flex-shrink-0" />
					{#if !isCollapsed}
						<span class="font-medium">Window</span>
						<span class="ml-auto text-xs sb-desktop-badge">Desktop</span>
					{/if}
				</a>
			</div>

			<Separator.Root class="sb-sep h-px" />

			<!-- Navigation -->
			<nav class="flex-1 p-2 overflow-y-auto sb-nav-scroll">
				{#each navGroups as group, gi}
					<!-- Section header -->
					{#if !isCollapsed}
						<button
							class="sb-section-header"
							onclick={() => toggleSection(group.label)}
							aria-label="Toggle {group.label} section"
						>
							<span class="sb-section-label">{group.label}</span>
							<ChevronRight size={16} strokeWidth={2} class="sb-section-chevron {collapsedSections.has(group.label) ? '' : 'sb-section-chevron--open'}" />
						</button>
					{:else if gi > 0}
						<div class="sb-section-dot"></div>
					{/if}

					{#if !collapsedSections.has(group.label)}
						<div class="sb-section-items {collapsedSections.has(group.label) ? 'sb-section-items--collapsed' : ''}">
							{#each group.items as item}
								{#if item.label === 'Projects'}
									<!-- Projects with dropdown -->
									<div class="relative group">
										<div class="flex items-center">
											<a
												href={item.href}
												class="sb-nav-item flex-1 flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-all duration-200
													{$page.url.pathname.startsWith(item.href) ? 'sb-nav-item--active' : ''}
													{isCollapsed ? 'justify-center' : ''}"
												title={isCollapsed ? item.label : ''}
											>
												<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={item.icon} />
												</svg>
												{#if !isCollapsed}
													<span>{item.label}</span>
												{/if}
											</a>
											{#if !isCollapsed && projects.length > 0}
												<button
													onclick={() => showProjectsDropdown = !showProjectsDropdown}
													class="sb-chevron-btn p-1.5 rounded-lg transition-colors mr-1 opacity-0 group-hover:opacity-100 {showProjectsDropdown ? 'opacity-100' : ''} transition-opacity duration-200"
													title="Show recent projects"
												>
													<ChevronDown size={16} strokeWidth={2} class="sb-chevron-icon transition-transform {showProjectsDropdown ? 'rotate-180' : ''}" />
												</button>
											{/if}
										</div>
										{#if showProjectsDropdown && !isCollapsed && projects.length > 0}
											<div class="ml-6 mt-1 space-y-0.5">
												{#each projects as project}
													<a
														href="/projects/{project.id}"
														class="sb-sub-item flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs transition-colors
															{$page.url.pathname === `/projects/${project.id}` ? 'sb-sub-item--active' : ''}"
													>
														<span class="w-1.5 h-1.5 rounded-full {project.status === 'active' ? 'bg-green-500' : 'sb-dot-inactive'}"></span>
														<span class="truncate">{project.name}</span>
													</a>
												{/each}
											</div>
										{/if}
									</div>
								{:else}
									<a
										href={item.href}
										class="sb-nav-item flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-all duration-200
											{$page.url.pathname.startsWith(item.href) ? 'sb-nav-item--active' : ''}
											{isCollapsed ? 'justify-center' : ''}"
										title={isCollapsed ? item.label : ''}
									>
										<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={item.icon} />
										</svg>
										{#if !isCollapsed}
											<span>{item.label}</span>
										{/if}
									</a>
								{/if}
							{/each}
						</div>
					{/if}
				{/each}
			</nav>

			<Separator.Root class="sb-sep h-px" />

			<!-- User Section - Links to Profile -->
			<div class="p-3">
				<a
					href="/profile"
					class="sb-user-link flex items-center gap-3 p-2 rounded-xl transition-colors {$page.url.pathname === '/profile' ? 'sb-user-link--active' : ''}"
					title={isCollapsed ? 'Profile' : ''}
				>
					{#if $session.data.user?.image}
						<img
							src={$session.data.user.image.startsWith('/') ? `${getBackendUrl()}${$session.data.user.image}` : $session.data.user.image}
							alt={$session.data.user?.name || 'Profile'}
							class="w-9 h-9 rounded-full object-cover flex-shrink-0 sb-avatar-border"
						/>
					{:else}
						<div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">
							{$session.data.user?.name?.charAt(0).toUpperCase() || 'U'}
						</div>
					{/if}
					{#if !isCollapsed}
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium sb-user-name truncate">{$session.data.user?.name}</p>
							<p class="text-xs sb-user-email truncate">{$session.data.user?.email}</p>
						</div>
						<ChevronRight size={16} strokeWidth={2} class="sb-user-chevron" />
					{/if}
				</a>
			</div>
		</aside>

		<!-- Main Content -->
		<main class="flex-1 flex flex-col min-w-0 overflow-hidden rounded-[14px] sb-panel">
			{#if needsTrafficLightSpace}
				<!-- Draggable titlebar region for main content area (Electron macOS only) -->
				<div
					class="h-12 flex-shrink-0 drag-region sb-main-border-b"
					style="-webkit-app-region: drag;"
				>
					<!-- This provides a drag area across the top of the main content -->
				</div>
				<div class="flex-1 overflow-hidden -mt-12 pt-12">
					{@render children()}
				</div>
			{:else}
				<div class="flex-1 overflow-hidden">
					{@render children()}
				</div>
			{/if}
		</main>

	</div>
	{/if}
{/if}

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  SIDEBAR (sb-) — Foundation Design Tokens                    */
	/* ══════════════════════════════════════════════════════════════ */

	/* Canvas — the background layer behind floating panels */
	.sb-canvas {
		background: #e6e6e6;
	}
	:global(.dark) .sb-canvas {
		background: #000000;
	}

	/* Panel — shared bg for sidebar + main content */
	.sb-panel {
		background: var(--dbg, #fff);
	}

	/* Legacy: embed mode still uses flat background */
	.sb-main-bg {
		background: var(--dbg, #fff);
	}
	.sb-main-border-b {
		border-bottom: 1px solid transparent;
	}

	/* Sidebar container */
	.sb-sidebar {
		background: var(--dbg, #fff);
	}

	/* Title */
	.sb-title {
		color: var(--dt, #111);
	}

	/* Separator */
	.sb-sep {
		background: var(--dbd2, #f0f0f0);
	}

	/* Desktop button */
	.sb-desktop-btn {
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd2, #f0f0f0);
		color: var(--dt2, #555);
	}
	.sb-desktop-btn:hover {
		background: var(--dbg3, #eee);
		color: var(--dt, #111);
	}
	.sb-desktop-badge {
		color: var(--dt3, #888);
	}

	/* ── Nav Items ──────────────────────────────────────────────── */
	.sb-nav-item {
		color: var(--dt2, #555);
		position: relative;
		transition: background 280ms ease, color 200ms ease;
	}
	.sb-nav-item::after {
		content: '';
		position: absolute;
		right: 0;
		top: 50%;
		transform: translateY(-50%);
		width: 3px;
		height: 0%;
		border-radius: 3px 0 0 3px;
		background: var(--bos-nav-active);
		opacity: 0;
		filter: blur(0px);
		transition: height 300ms cubic-bezier(0.4, 0, 0.2, 1),
		            opacity 250ms ease,
		            box-shadow 350ms ease;
	}
	.sb-nav-item:hover {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}
	.sb-nav-item--active {
		background: linear-gradient(90deg, transparent 0%, transparent 60%, var(--bos-nav-active-bg) 100%);
		color: var(--dt, #fff);
	}
	.sb-nav-item--active::after {
		height: 55%;
		opacity: 1;
		box-shadow: 0 0 8px 2px var(--bos-nav-active-glow),
		            0 0 22px 6px var(--bos-nav-active-bg);
		animation: nav-glow-pulse 3s ease-in-out infinite;
	}
	.sb-nav-item--active:hover {
		background: linear-gradient(90deg, transparent 0%, transparent 55%, var(--bos-nav-active-bg) 100%);
		color: var(--dt, #fff);
	}
	@keyframes nav-glow-pulse {
		0%, 100% { box-shadow: 0 0 8px 2px var(--bos-nav-active-glow), 0 0 22px 6px var(--bos-nav-active-bg); }
		50% { box-shadow: 0 0 12px 3px var(--bos-nav-active-glow), 0 0 28px 8px var(--bos-nav-active-bg); }
	}

	/* ── Section Headers ────────────────────────────────────────── */
	.sb-section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 6px 12px 4px;
		margin-top: 8px;
		border: none;
		background: none;
		cursor: pointer;
	}
	.sb-section-header:first-child {
		margin-top: 0;
	}
	.sb-section-label {
		font-size: 0.65rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--dt4, #bbb);
	}
	.sb-section-chevron {
		width: 12px;
		height: 12px;
		color: var(--dt4, #bbb);
		transition: transform 200ms ease;
	}
	.sb-section-chevron--open {
		transform: rotate(90deg);
	}
	.sb-section-dot {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: var(--dbd, #e0e0e0);
		margin: 8px auto;
	}
	.sb-section-items {
		display: flex;
		flex-direction: column;
		gap: 1px;
	}
	.sb-nav-scroll {
		scrollbar-width: none;
	}
	.sb-nav-scroll::-webkit-scrollbar {
		display: none;
	}

	/* ── Workspace Badge (collapsed) ────────────────────────────── */
	.sb-workspace-badge {
		width: 34px;
		height: 34px;
		border-radius: 8px;
		background: linear-gradient(135deg, var(--bos-nav-active) 0%, var(--bos-category-productivity) 100%);
		border: none;
		color: var(--bos-surface-on-color);
		font-size: 0.8rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: transform 150ms ease, box-shadow 200ms ease;
	}
	.sb-workspace-badge:hover {
		transform: scale(1.06);
		box-shadow: 0 0 0 2px var(--dbg, #fff), 0 0 0 4px var(--bos-nav-active-glow);
	}

	/* Sub-items (projects dropdown) */
	.sb-sub-item {
		color: var(--dt3, #888);
	}
	.sb-sub-item:hover {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}
	.sb-sub-item--active {
		background: var(--bos-nav-active-bg);
		color: var(--bos-nav-active);
	}
	.sb-dot-inactive {
		background: var(--dt4, #bbb);
	}

	/* Chevron toggle for projects dropdown */
	.sb-chevron-btn:hover {
		background: var(--dbg3, #eee);
	}
	.sb-chevron-icon {
		color: var(--dt3, #888);
	}

	/* ── User Section ───────────────────────────────────────────── */
	.sb-user-link:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.sb-user-link--active {
		background: var(--dbg2, #f5f5f5);
	}
	.sb-avatar-border {
		border: 2px solid var(--dbd, #e0e0e0);
	}
	.sb-user-name {
		color: var(--dt, #111);
	}
	.sb-user-email {
		color: var(--dt3, #888);
	}
	.sb-user-chevron {
		color: var(--dt4, #bbb);
	}
</style>
