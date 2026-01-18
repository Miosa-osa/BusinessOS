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
	import { WorkspaceSwitcher } from '$lib/components/workspace';
	import { loadSavedWorkspace } from '$lib/stores/workspaces';
	import { notificationStore } from '$lib/stores/notifications';
	import { initializePush } from '$lib/services/pushService';

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

	onMount(() => {
		// Skip initialization in embed mode - iframes don't need workspace/notification systems
		if (isEmbedMode) return;

		// Initialize workspace store
		loadSavedWorkspace();
		// Load projects for sidebar
		loadProjects();

		// Initialize notifications (SSE + Push)
		notificationStore.initialize();
		initializePush();
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

	$effect(() => {
		// Skip auth check in embed mode (for 3D Desktop iframes)
		if (isEmbedMode) return;

		if (!$session.isPending && !$session.data) {
			goto('/login');
		}
	});

	const navItems = [
		{
			href: '/dashboard',
			label: 'Dashboard',
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			href: '/chat',
			label: 'Chat',
			icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z'
		},
		{
			href: '/tasks',
			label: 'Tasks',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4'
		},
		{
			href: '/communication',
			label: 'Communication',
			icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z'
		},
		{
			href: '/projects',
			label: 'Projects',
			icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
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
		{
			href: '/tables',
			label: 'Tables',
			icon: 'M3 10h18M3 14h18M9 3v18M15 3v18M3 6a2 2 0 012-2h14a2 2 0 012 2v12a2 2 0 01-2 2H5a2 2 0 01-2-2V6z'
		},
		{
			href: '/knowledge-v2',
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
			href: '/daily',
			label: 'Daily Log',
			icon: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z'
		},
		{
			href: '/usage',
			label: 'Usage',
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
	];

</script>

{#if $session.data}
	{#if isEmbedMode}
		<!-- Embed mode: no sidebar, just content -->
		<div class="h-screen w-screen overflow-hidden bg-white dark:bg-gray-900">
			{@render children()}
		</div>
	{:else}
	<div class="h-screen flex overflow-hidden bg-white dark:bg-gray-900">
		<!-- Sidebar -->
		<aside
			class="sidebar h-full flex flex-col flex-shrink-0 transition-all duration-300 ease-in-out {isCollapsed ? (needsTrafficLightSpace ? 'w-20' : 'w-16') : 'w-64'}"
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
					<h1 class="text-lg font-semibold text-gray-900 dark:text-white">Business OS</h1>
				{/if}
				<button
					onclick={toggleSidebar}
					class="btn-pill btn-pill-icon btn-pill-ghost btn-pill-sm no-drag"
					style="-webkit-app-region: no-drag;"
					title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
				>
					<svg
						class="w-5 h-5 transition-transform duration-300 {isCollapsed ? 'rotate-180' : ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					</svg>
				</button>
			</div>

			<!-- Workspace Switcher -->
			{#if !isCollapsed}
				<div class="px-2 pb-2">
					<WorkspaceSwitcher />
				</div>
			{/if}

			<!-- Window Desktop Button -->
			<div class="px-2 pb-2">
				<a
					href="/window"
					class="flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-all duration-200
						bg-gradient-to-r from-gray-100 to-gray-50 hover:from-gray-200 hover:to-gray-100
						dark:from-gray-700 dark:to-gray-800 dark:hover:from-gray-600 dark:hover:to-gray-700
						border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-200 hover:text-gray-900 dark:hover:text-white
						{isCollapsed ? 'justify-center' : ''}"
					title={isCollapsed ? 'Window Desktop' : ''}
				>
					<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
					{#if !isCollapsed}
						<span class="font-medium">Window</span>
						<span class="ml-auto text-xs text-gray-400 dark:text-gray-500">Desktop</span>
					{/if}
				</a>
			</div>

			<Separator.Root class="h-px bg-gray-200 dark:bg-gray-700" />

			<!-- Navigation -->
			<nav class="flex-1 p-2 space-y-1 overflow-y-auto">
				{#each navItems as item}
					{#if item.label === 'Projects'}
						<!-- Projects with dropdown -->
						<div class="relative group">
							<div class="flex items-center">
								<a
									href={item.href}
									class="flex-1 nav-pill flex items-center gap-3 px-3 py-2.5 rounded-full text-sm font-medium transition-all duration-200
										{$page.url.pathname.startsWith(item.href)
											? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/30'
											: 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-white/5 hover:shadow-md'}
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
										class="btn-pill btn-pill-icon btn-pill-xs btn-pill-ghost mr-1 opacity-0 group-hover:opacity-100 {showProjectsDropdown ? 'opacity-100' : ''} transition-opacity duration-200"
										title="Show recent projects"
									>
										<svg class="w-4 h-4 text-gray-500 dark:text-gray-400 transition-transform {showProjectsDropdown ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
										</svg>
									</button>
								{/if}
							</div>
							{#if showProjectsDropdown && !isCollapsed && projects.length > 0}
								<div class="ml-6 mt-1 space-y-0.5">
									{#each projects as project}
										<a
											href="/projects/{project.id}"
											class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs transition-colors
												{$page.url.pathname === `/projects/${project.id}` ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-300'}"
										>
											<span class="w-1.5 h-1.5 rounded-full {project.status === 'active' ? 'bg-green-500' : 'bg-gray-400'}"></span>
											<span class="truncate">{project.name}</span>
										</a>
									{/each}
								</div>
							{/if}
						</div>
					{:else}
						<a
							href={item.href}
							class="nav-pill flex items-center gap-3 px-3 py-2.5 rounded-full text-sm font-medium transition-all duration-200
								{$page.url.pathname.startsWith(item.href)
									? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/30'
									: 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-white/5 hover:shadow-md'}
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
			</nav>

			<Separator.Root class="h-px bg-gray-200 dark:bg-gray-700" />

			<!-- User Section - Links to Profile -->
			<div class="p-3">
				<a
					href="/profile"
					class="flex items-center gap-3 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors {$page.url.pathname === '/profile' ? 'bg-gray-100 dark:bg-gray-700' : ''}"
					title={isCollapsed ? 'Profile' : ''}
				>
					{#if $session.data.user?.image}
						<img
							src={$session.data.user.image.startsWith('/') ? `http://localhost:8001${$session.data.user.image}` : $session.data.user.image}
							alt={$session.data.user?.name || 'Profile'}
							class="w-9 h-9 rounded-full object-cover flex-shrink-0 border-2 border-gray-200 dark:border-gray-600"
						/>
					{:else}
						<div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">
							{$session.data.user?.name?.charAt(0).toUpperCase() || 'U'}
						</div>
					{/if}
					{#if !isCollapsed}
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{$session.data.user?.name}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400 truncate">{$session.data.user?.email}</p>
						</div>
						<svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					{/if}
				</a>
			</div>
		</aside>

		<!-- Main Content -->
		<main class="flex-1 h-full flex flex-col min-w-0 overflow-hidden bg-white dark:bg-gray-900">
			{#if needsTrafficLightSpace}
				<!-- Draggable titlebar region for main content area (Electron macOS only) -->
				<div
					class="h-12 flex-shrink-0 drag-region border-b border-gray-100 dark:border-gray-800"
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
