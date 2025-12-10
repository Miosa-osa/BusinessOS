<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { useSession } from '$lib/auth-client';
	import { Separator } from 'bits-ui';

	let { children } = $props();

	const session = useSession();

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
			href: '/contexts',
			label: 'Contexts',
			icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10'
		},
		{
			href: '/nodes',
			label: 'Nodes',
			icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z'
		},
		{
			href: '/daily',
			label: 'Daily Log',
			icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z'
		},
		{
			href: '/settings',
			label: 'Settings',
			icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z'
		},
	];

</script>

{#if $session.isPending}
	<div class="min-h-screen flex items-center justify-center">
		<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
	</div>
{:else if $session.data}
	<div class="h-screen flex overflow-hidden">
		<!-- Sidebar -->
		<aside
			class="sidebar h-full flex flex-col flex-shrink-0 transition-all duration-300 ease-in-out {isCollapsed ? 'w-16' : 'w-64'}"
		>
			<!-- Header -->
			<div class="p-4 flex items-center {isCollapsed ? 'justify-center' : 'justify-between'}">
				{#if !isCollapsed}
					<h1 class="text-lg font-semibold text-gray-900">Business OS</h1>
				{/if}
				<button
					onclick={toggleSidebar}
					class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 transition-colors text-gray-500"
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

			<Separator.Root class="h-px bg-gray-200" />

			<!-- Navigation -->
			<nav class="flex-1 p-2 space-y-1">
				{#each navItems as item}
					<a
						href={item.href}
						class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200
							{$page.url.pathname.startsWith(item.href) ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}
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
				{/each}
			</nav>

			<Separator.Root class="h-px bg-gray-200" />

			<!-- User Section - Links to Profile -->
			<div class="p-3">
				<a
					href="/profile"
					class="flex items-center gap-3 p-2 rounded-xl hover:bg-gray-100 transition-colors {$page.url.pathname === '/profile' ? 'bg-gray-100' : ''}"
					title={isCollapsed ? 'Profile' : ''}
				>
					<div class="w-9 h-9 rounded-full bg-gray-900 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">
						{$session.data.user?.name?.charAt(0).toUpperCase() || 'U'}
					</div>
					{#if !isCollapsed}
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 truncate">{$session.data.user?.name}</p>
							<p class="text-xs text-gray-500 truncate">{$session.data.user?.email}</p>
						</div>
						<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					{/if}
				</a>
			</div>
		</aside>

		<!-- Main Content -->
		<main class="flex-1 h-full flex flex-col min-w-0 overflow-hidden">
			{@render children()}
		</main>
	</div>
{/if}
