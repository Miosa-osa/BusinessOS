<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { useSession } from '$lib/auth-client';
	import { browser } from '$app/environment';
	import { isElectron as checkElectron, isMacOS } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { getBackendUrl, initCSRF } from '$lib/api/base';
	import { WorkspaceSwitcher } from '$lib/components/workspace';
	import { loadSavedWorkspace } from '$lib/stores/workspaces';
	import { notificationStore } from '$lib/stores/notifications';
	import { initializePush } from '$lib/services/pushService';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { ChevronsLeft, ChevronRight } from 'lucide-svelte';

	onMount(async () => {
		// First-run detection: only redirect to welcome in Electron (desktop app)
		// Web users who signed in are already in cloud mode — skip welcome
		const isElectron = browser && 'electron' in window;
		if (isElectron && !localStorage.getItem('businessos_setup_complete')) {
			goto('/welcome');
			return;
		}
		// Auto-mark web users as cloud setup complete
		if (browser && !localStorage.getItem('businessos_setup_complete')) {
			localStorage.setItem('businessos_setup_complete', 'cloud');
			localStorage.setItem('businessos_mode', 'cloud');
		}

		// Initialize CSRF token first (required before any state-changing requests)
		await initCSRF();

		// Skip rest of initialization in embed mode - iframes don't need workspace/notification systems
		if (isEmbedMode) return;

		await loadSavedWorkspace();

		// These don't return promises — fire immediately
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
			label: 'Workspace',
			items: [
				{
					href: '/pages',
					label: 'Knowledge',
					icon: 'M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z'
				},
			]
		},
		{
			label: 'System',
			items: [
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

			<div class="sb-sep h-px"></div>

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
							<ChevronRight
								size={16}
								strokeWidth={2}
								class="h-3 w-3 text-neutral-400 transition-transform duration-200 {collapsedSections.has(group.label) ? '' : 'rotate-90'}"
							/>
						</button>
					{:else if gi > 0}
						<div class="sb-section-dot"></div>
					{/if}

					{#if !collapsedSections.has(group.label)}
						<div class="sb-section-items {collapsedSections.has(group.label) ? 'sb-section-items--collapsed' : ''}">
							{#each group.items as item}
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
							{/each}
						</div>
					{/if}
				{/each}
			</nav>

			<div class="sb-sep h-px"></div>

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

	{:else if !bootComplete}
	<!-- Loading state only during initial auth check -->
	<div class="h-screen flex items-center justify-center sb-main-bg">
		<div class="text-center">
			<div class="w-8 h-8 border-2 border-gray-300 border-t-blue-500 rounded-full animate-spin mx-auto mb-4"></div>
			<p class="sb-muted">Loading...</p>
		</div>
	</div>
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
</style>
