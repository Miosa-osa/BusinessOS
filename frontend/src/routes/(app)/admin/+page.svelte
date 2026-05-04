<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ── Types ──────────────────────────────────────────────────────────────────

	type TabId = 'overview' | 'users' | 'computers' | 'workspaces' | 'settings' | 'audit';

	interface DashboardData {
		users: { total: number; active: number };
		workspaces: { total: number; active: number };
		computers: { total: number; running: number; hibernated: number };
		revenue: { mrr_cents: number; active_subscriptions: number };
	}

	interface AdminUser {
		id: string;
		email: string;
		display_name: string;
		platform_role: string;
		workspace_name: string;
		plan: string;
		created_at: string;
	}

	interface AdminComputer {
		id: string;
		workspace_name: string;
		owner_email: string;
		miosa_slug: string;
		status: 'running' | 'hibernated' | 'stopped' | 'error';
		plan: string;
		ram_mb: number;
		vcpus: number;
		created_at: string;
	}

	interface AdminWorkspace {
		id: string;
		name: string;
		owner_email: string;
		owner_name: string;
		plan: string;
		member_count: number;
		created_at: string;
	}

	interface AdminSettings {
		environment: string;
		deployment_mode: string;
		ai_provider: string;
		default_model: string;
		allowed_origins: string[];
		feature_flags: {
			osa_enabled: boolean;
			carrier_enabled: boolean;
			nats_enabled: boolean;
		};
	}

	interface AuditEntry {
		actor: string;
		action: string;
		target: string;
		time: string;
	}

	// ── State ──────────────────────────────────────────────────────────────────

	let activeTab = $state<TabId>('overview');

	// Overview
	let dashboard = $state<DashboardData | null>(null);
	let dashboardLoading = $state(true);

	// Users
	let users = $state<AdminUser[]>([]);
	let usersTotal = $state(0);
	let usersLoading = $state(true);

	// Computers
	let computers = $state<AdminComputer[]>([]);
	let computersTotal = $state(0);
	let computersLoading = $state(true);

	// Workspaces
	let workspaces = $state<AdminWorkspace[]>([]);
	let workspacesTotal = $state(0);
	let workspacesLoading = $state(true);
	let expandedWorkspaceId = $state<string | null>(null);

	// Settings
	let settings = $state<AdminSettings | null>(null);
	let settingsLoading = $state(true);
	let settingsUnavailable = $state(false);

	// Audit
	let auditEntries = $state<AuditEntry[]>([]);
	let auditLoading = $state(true);
	let auditUnavailable = $state(false);

	// Action menu (shared across tabs)
	let openMenuId = $state<string | null>(null);
	let actionBusy = $state(false);
	let actionError = $state<string | null>(null);
	let accessDenied = $state(false);

	let refreshInterval: ReturnType<typeof setInterval> | null = null;

	// ── Derived ────────────────────────────────────────────────────────────────

	const tabs: Array<{ id: TabId; label: string }> = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'users', label: 'Users' },
		{ id: 'computers', label: 'Computers' },
		{ id: 'workspaces', label: 'Workspaces' },
		{ id: 'settings', label: 'Settings' },
		{ id: 'audit', label: 'Audit Log' },
	];

	const PLAN_OPTIONS: Array<{ value: string; label: string }> = [
		{ value: 'free', label: 'Free' },
		{ value: 'pro', label: 'Pro' },
		{ value: 'growth', label: 'Growth' },
		{ value: 'business', label: 'Business' },
	];

	// ── API helpers ────────────────────────────────────────────────────────────

	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrfToken = getCSRFToken();
		if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
		return headers;
	}

	async function adminFetch<T>(path: string): Promise<T | null> {
		try {
			const res = await fetch(`${getApiBaseUrl()}${path}`, {
				credentials: 'include',
				headers: buildHeaders(),
				signal: AbortSignal.timeout(10000),
			});
			if (res.status === 403) {
				accessDenied = true;
				return null;
			}
			if (!res.ok) return null;
			return res.json() as Promise<T>;
		} catch {
			return null;
		}
	}

	async function adminPost(path: string, body: Record<string, unknown>): Promise<boolean> {
		try {
			const res = await fetch(`${getApiBaseUrl()}${path}`, {
				method: 'POST',
				credentials: 'include',
				headers: { ...buildHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
				signal: AbortSignal.timeout(10000),
			});
			return res.ok;
		} catch {
			return false;
		}
	}

	// ── Data loading ────────────────────────────────────────────────────────────

	async function loadDashboard() {
		dashboardLoading = true;
		dashboard = await adminFetch<DashboardData>('/admin/dashboard');
		dashboardLoading = false;
	}

	async function loadUsers() {
		usersLoading = true;
		const data = await adminFetch<{ users: AdminUser[]; total: number }>('/admin/users');
		users = data?.users ?? [];
		usersTotal = data?.total ?? 0;
		usersLoading = false;
	}

	async function loadComputers() {
		computersLoading = true;
		const data = await adminFetch<{ computers: AdminComputer[]; total: number }>('/admin/computers');
		computers = data?.computers ?? [];
		computersTotal = data?.total ?? 0;
		computersLoading = false;
	}

	async function loadWorkspaces() {
		workspacesLoading = true;
		const data = await adminFetch<{ workspaces: AdminWorkspace[]; total: number }>('/admin/workspaces');
		workspaces = data?.workspaces ?? [];
		workspacesTotal = data?.total ?? 0;
		workspacesLoading = false;
	}

	async function loadSettings() {
		settingsLoading = true;
		const data = await adminFetch<AdminSettings>('/admin/settings');
		if (data === null) {
			settingsUnavailable = true;
			settings = null;
		} else {
			settingsUnavailable = false;
			settings = data;
		}
		settingsLoading = false;
	}

	async function loadAudit() {
		auditLoading = true;
		const data = await adminFetch<{ entries: AuditEntry[] }>('/admin/audit-log');
		if (data === null) {
			auditUnavailable = true;
			auditEntries = [];
		} else {
			auditUnavailable = false;
			auditEntries = data.entries ?? [];
		}
		auditLoading = false;
	}

	// ── Actions ────────────────────────────────────────────────────────────────

	async function setUserRole(userId: string, role: 'admin' | 'user') {
		if (actionBusy) return;
		actionBusy = true;
		actionError = null;
		openMenuId = null;
		const ok = await adminPost(`/admin/users/${userId}/role`, { role });
		if (!ok) {
			actionError = 'Failed to update user role. Please try again.';
		} else {
			await loadUsers();
		}
		actionBusy = false;
	}

	async function setUserPlan(userId: string, plan: string) {
		if (actionBusy) return;
		actionBusy = true;
		actionError = null;
		openMenuId = null;
		const ok = await adminPost(`/admin/users/${userId}/plan`, { plan });
		if (!ok) {
			actionError = 'Failed to update user plan. Please try again.';
		} else {
			await loadUsers();
		}
		actionBusy = false;
	}

	async function computerAction(computerId: string, action: 'restart' | 'stop' | 'hibernate') {
		if (actionBusy) return;
		actionBusy = true;
		actionError = null;
		openMenuId = null;
		const ok = await adminPost(`/admin/computers/${computerId}/${action}`, {});
		if (!ok) {
			actionError = `Failed to ${action} computer. Please try again.`;
		} else {
			await loadComputers();
		}
		actionBusy = false;
	}

	// ── Formatters ─────────────────────────────────────────────────────────────

	function formatDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString('en-US', {
				month: 'short',
				day: 'numeric',
				year: 'numeric',
			});
		} catch {
			return iso;
		}
	}

	function formatMRR(cents: number): string {
		return `$${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
	}

	function formatRam(mb: number): string {
		if (mb >= 1024) return `${(mb / 1024).toFixed(0)} GB`;
		return `${mb} MB`;
	}

	function planLabel(plan: string): string {
		return plan.charAt(0).toUpperCase() + plan.slice(1);
	}

	// ── Lifecycle ──────────────────────────────────────────────────────────────

	onMount(async () => {
		await Promise.all([
			loadDashboard(),
			loadUsers(),
			loadComputers(),
			loadWorkspaces(),
			loadSettings(),
			loadAudit(),
		]);

		// Auto-refresh overview stats every 60s
		refreshInterval = setInterval(loadDashboard, 60_000);
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});

	function handleOutsideClick() {
		openMenuId = null;
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="h-full flex flex-col adm-page"
	onclick={handleOutsideClick}
	onkeydown={(e) => { if (e.key === 'Escape') openMenuId = null; }}
>
	<!-- Header -->
	<div class="px-6 py-3 adm-header">
		<h1 class="adm-title">Platform Admin</h1>
		<p class="adm-muted mt-0">System-wide management — superadmin only</p>
	</div>

	<div class="flex-1 overflow-y-auto">
		<div class="max-w-6xl mx-auto px-6 py-4">

			<!-- Access denied banner -->
			{#if accessDenied}
				<div class="adm-error-banner" role="alert">
					Access denied. You must be a superadmin to view this page.
				</div>
			{/if}

			<!-- Action error banner -->
			{#if actionError}
				<div class="adm-error-banner" role="alert">
					{actionError}
					<button onclick={() => (actionError = null)} class="adm-error-dismiss" aria-label="Dismiss error">Dismiss</button>
				</div>
			{/if}

			<!-- Tab Navigation -->
			<div class="mb-4">
				<div class="adm-tab-row">
					{#each tabs as tab}
						<button
							onclick={() => (activeTab = tab.id)}
							class="adm-tab"
							class:adm-tab--active={activeTab === tab.id}
						>
							{tab.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- ── Overview Tab ────────────────────────────────────────────── -->
			{#if activeTab === 'overview'}
				{#if dashboardLoading}
					<div class="adm-grid-4 mb-6">
						{#each Array(4) as _}
							<div class="adm-stat-card adm-skeleton-card">
								<div class="adm-skel adm-skel--sm mb-2"></div>
								<div class="adm-skel adm-skel--lg mb-1"></div>
								<div class="adm-skel adm-skel--sm w-2/3"></div>
							</div>
						{/each}
					</div>
				{:else if dashboard}
					<div class="adm-grid-4 mb-6">
						<!-- Users card -->
						<div class="adm-stat-card">
							<div class="adm-stat-icon adm-stat-icon--blue">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
							</div>
							<div class="adm-stat-value">{dashboard.users.total.toLocaleString()}</div>
							<div class="adm-stat-label">Total Users</div>
							<div class="adm-stat-sub">
								<span class="adm-dot adm-dot--green"></span>
								{dashboard.users.active.toLocaleString()} active
							</div>
						</div>

						<!-- Workspaces card -->
						<div class="adm-stat-card">
							<div class="adm-stat-icon adm-stat-icon--purple">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
								</svg>
							</div>
							<div class="adm-stat-value">{dashboard.workspaces.total.toLocaleString()}</div>
							<div class="adm-stat-label">Workspaces</div>
							<div class="adm-stat-sub">
								<span class="adm-dot adm-dot--green"></span>
								{dashboard.workspaces.active.toLocaleString()} active
							</div>
						</div>

						<!-- Computers card -->
						<div class="adm-stat-card">
							<div class="adm-stat-icon adm-stat-icon--teal">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
								</svg>
							</div>
							<div class="adm-stat-value">{dashboard.computers.total.toLocaleString()}</div>
							<div class="adm-stat-label">Computers</div>
							<div class="adm-stat-sub">
								<span class="adm-dot adm-dot--green"></span>
								{dashboard.computers.running} running
								<span class="adm-dot adm-dot--yellow ml-2"></span>
								{dashboard.computers.hibernated} hibernated
							</div>
						</div>

						<!-- Revenue card -->
						<div class="adm-stat-card">
							<div class="adm-stat-icon adm-stat-icon--green">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
							</div>
							<div class="adm-stat-value">{formatMRR(dashboard.revenue.mrr_cents)}</div>
							<div class="adm-stat-label">Monthly Recurring Revenue</div>
							<div class="adm-stat-sub">
								<span class="adm-dot adm-dot--green"></span>
								{dashboard.revenue.active_subscriptions} active subscriptions
							</div>
						</div>
					</div>
				{:else}
					<div class="adm-empty mb-6">
						<p class="adm-empty-text">Dashboard data unavailable — admin endpoints may not be deployed yet.</p>
					</div>
				{/if}
			{/if}

			<!-- ── Users Tab ───────────────────────────────────────────────── -->
			{#if activeTab === 'users'}
				<div class="adm-section-header mb-3">
					<span class="adm-section-title">All Users</span>
					{#if !usersLoading}
						<span class="adm-count-badge">{usersTotal.toLocaleString()} total</span>
					{/if}
				</div>

				{#if usersLoading}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									{#each ['Name', 'Email', 'Role', 'Workspace', 'Plan', 'Joined', 'Actions'] as col}
										<th class="adm-th">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each Array(5) as _}
									<tr class="adm-tr">
										{#each Array(7) as _}
											<td class="adm-td"><div class="adm-skel adm-skel--sm"></div></td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else if users.length === 0}
					<div class="adm-empty">
						<p class="adm-empty-text">No users found.</p>
					</div>
				{:else}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									<th class="adm-th">Name</th>
									<th class="adm-th">Email</th>
									<th class="adm-th">Role</th>
									<th class="adm-th">Workspace</th>
									<th class="adm-th">Plan</th>
									<th class="adm-th">Joined</th>
									<th class="adm-th adm-th--right">Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each users as user, i}
									<tr class="adm-tr" class:adm-tr--alt={i % 2 === 1}>
										<td class="adm-td adm-td--name">{user.display_name || '—'}</td>
										<td class="adm-td adm-td--muted">{user.email}</td>
										<td class="adm-td">
											<span class="adm-role-badge adm-role-badge--{user.platform_role}">
												{user.platform_role}
											</span>
										</td>
										<td class="adm-td adm-td--muted">{user.workspace_name || '—'}</td>
										<td class="adm-td">
											<span class="adm-plan-badge">{planLabel(user.plan)}</span>
										</td>
										<td class="adm-td adm-td--muted">{formatDate(user.created_at)}</td>
										<td class="adm-td adm-td--right">
											<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
								<div class="adm-action-wrap" role="none" onclick={(e) => e.stopPropagation()}>
												<button
													class="adm-action-btn"
													onclick={() => { openMenuId = openMenuId === user.id ? null : user.id; }}
													aria-label="User actions"
												>
													<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
														<circle cx="5" cy="12" r="2" />
														<circle cx="12" cy="12" r="2" />
														<circle cx="19" cy="12" r="2" />
													</svg>
												</button>
												{#if openMenuId === user.id}
													<div class="adm-dropdown">
														<button
															class="adm-dropdown-item"
															onclick={() => setUserRole(user.id, 'admin')}
															disabled={actionBusy}
														>
															Set Admin
														</button>
														<button
															class="adm-dropdown-item"
															onclick={() => setUserRole(user.id, 'user')}
															disabled={actionBusy}
														>
															Set User
														</button>
														<div class="adm-dropdown-sep"></div>
														<div class="adm-dropdown-label">Change Plan</div>
														{#each PLAN_OPTIONS as opt}
															<button
																class="adm-dropdown-item"
																class:adm-dropdown-item--active={user.plan === opt.value}
																onclick={() => setUserPlan(user.id, opt.value)}
																disabled={actionBusy}
															>
																{opt.label}
																{#if user.plan === opt.value}
																	<svg class="w-3 h-3 adm-dropdown-check" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
																	</svg>
																{/if}
															</button>
														{/each}
														<div class="adm-dropdown-sep"></div>
														<a
															href="/admin/workspace?id={user.id}"
															class="adm-dropdown-item"
														>
															View Workspace
														</a>
													</div>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}

			<!-- ── Computers Tab ───────────────────────────────────────────── -->
			{#if activeTab === 'computers'}
				<div class="adm-section-header mb-3">
					<span class="adm-section-title">All Computers</span>
					{#if !computersLoading}
						<span class="adm-count-badge">{computersTotal.toLocaleString()} total</span>
					{/if}
				</div>

				{#if computersLoading}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									{#each ['Workspace', 'Owner', 'Slug', 'Status', 'Plan', 'RAM', 'CPUs', 'Created', 'Actions'] as col}
										<th class="adm-th">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each Array(5) as _}
									<tr class="adm-tr">
										{#each Array(9) as _}
											<td class="adm-td"><div class="adm-skel adm-skel--sm"></div></td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else if computers.length === 0}
					<div class="adm-empty">
						<p class="adm-empty-text">No computers found.</p>
					</div>
				{:else}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									<th class="adm-th">Workspace</th>
									<th class="adm-th">Owner</th>
									<th class="adm-th">Slug</th>
									<th class="adm-th">Status</th>
									<th class="adm-th">Plan</th>
									<th class="adm-th">RAM</th>
									<th class="adm-th">CPUs</th>
									<th class="adm-th">Created</th>
									<th class="adm-th adm-th--right">Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each computers as comp, i}
									<tr class="adm-tr" class:adm-tr--alt={i % 2 === 1}>
										<td class="adm-td adm-td--name">{comp.workspace_name || '—'}</td>
										<td class="adm-td adm-td--muted">{comp.owner_email}</td>
										<td class="adm-td">
											<code class="adm-slug">{comp.miosa_slug || '—'}</code>
										</td>
										<td class="adm-td">
											<span class="adm-status-badge adm-status-badge--{comp.status}">
												<span class="adm-status-dot adm-status-dot--{comp.status}"></span>
												{comp.status}
											</span>
										</td>
										<td class="adm-td">
											<span class="adm-plan-badge">{planLabel(comp.plan)}</span>
										</td>
										<td class="adm-td adm-td--mono">{formatRam(comp.ram_mb)}</td>
										<td class="adm-td adm-td--mono">{comp.vcpus}</td>
										<td class="adm-td adm-td--muted">{formatDate(comp.created_at)}</td>
										<td class="adm-td adm-td--right">
											<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
											<div class="adm-action-wrap" role="none" onclick={(e) => e.stopPropagation()}>
												<button
													class="adm-action-btn"
													onclick={() => { openMenuId = openMenuId === comp.id ? null : comp.id; }}
													aria-label="Computer actions"
												>
													<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
														<circle cx="5" cy="12" r="2" />
														<circle cx="12" cy="12" r="2" />
														<circle cx="19" cy="12" r="2" />
													</svg>
												</button>
												{#if openMenuId === comp.id}
													<div class="adm-dropdown">
														<button
															class="adm-dropdown-item adm-dropdown-item--icon"
															onclick={() => computerAction(comp.id, 'restart')}
															disabled={actionBusy}
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75"
																	d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
															</svg>
															Restart
														</button>
														<button
															class="adm-dropdown-item adm-dropdown-item--icon"
															onclick={() => computerAction(comp.id, 'hibernate')}
															disabled={actionBusy}
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75"
																	d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
															</svg>
															Hibernate
														</button>
														<button
															class="adm-dropdown-item adm-dropdown-item--icon adm-dropdown-item--danger"
															onclick={() => computerAction(comp.id, 'stop')}
															disabled={actionBusy}
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75"
																	d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75"
																	d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
															</svg>
															Stop
														</button>
													</div>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}

			<!-- ── Workspaces Tab ──────────────────────────────────────────── -->
			{#if activeTab === 'workspaces'}
				<div class="adm-section-header mb-3">
					<span class="adm-section-title">All Workspaces</span>
					{#if !workspacesLoading}
						<span class="adm-count-badge">{workspacesTotal.toLocaleString()} total</span>
					{/if}
				</div>

				{#if workspacesLoading}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									{#each ['Workspace', 'Owner', 'Owner Name', 'Plan', 'Members', 'Created'] as col}
										<th class="adm-th">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each Array(5) as _}
									<tr class="adm-tr">
										{#each Array(6) as _}
											<td class="adm-td"><div class="adm-skel adm-skel--sm"></div></td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else if workspaces.length === 0}
					<div class="adm-empty">
						<p class="adm-empty-text">No workspaces found.</p>
					</div>
				{:else}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									<th class="adm-th">Workspace</th>
									<th class="adm-th">Owner Email</th>
									<th class="adm-th">Owner Name</th>
									<th class="adm-th">Plan</th>
									<th class="adm-th">Members</th>
									<th class="adm-th">Created</th>
								</tr>
							</thead>
							<tbody>
								{#each workspaces as ws, i}
									<!-- svelte-ignore a11y_click_events_have_key_events -->
									<tr
										class="adm-tr adm-tr--clickable"
										class:adm-tr--alt={i % 2 === 1}
										class:adm-tr--expanded={expandedWorkspaceId === ws.id}
										onclick={() => { expandedWorkspaceId = expandedWorkspaceId === ws.id ? null : ws.id; }}
										role="button"
										tabindex="0"
										onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { expandedWorkspaceId = expandedWorkspaceId === ws.id ? null : ws.id; } }}
									>
										<td class="adm-td adm-td--name">
											<div class="adm-ws-name-cell">
												<svg
													class="adm-ws-chevron"
													class:adm-ws-chevron--open={expandedWorkspaceId === ws.id}
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
													aria-hidden="true"
												>
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
												</svg>
												{ws.name}
											</div>
										</td>
										<td class="adm-td adm-td--muted">{ws.owner_email}</td>
										<td class="adm-td adm-td--muted">{ws.owner_name || '—'}</td>
										<td class="adm-td">
											<span class="adm-plan-badge">{planLabel(ws.plan)}</span>
										</td>
										<td class="adm-td adm-td--mono">{ws.member_count}</td>
										<td class="adm-td adm-td--muted">{formatDate(ws.created_at)}</td>
									</tr>
									{#if expandedWorkspaceId === ws.id}
										<tr class="adm-tr adm-tr--detail">
											<td colspan="6" class="adm-td adm-td--detail">
												<div class="adm-ws-detail">
													<div class="adm-ws-detail-grid">
														<div class="adm-ws-detail-item">
															<span class="adm-ws-detail-label">Workspace ID</span>
															<code class="adm-slug">{ws.id}</code>
														</div>
														<div class="adm-ws-detail-item">
															<span class="adm-ws-detail-label">Owner Email</span>
															<span class="adm-ws-detail-value">{ws.owner_email}</span>
														</div>
														<div class="adm-ws-detail-item">
															<span class="adm-ws-detail-label">Plan</span>
															<span class="adm-ws-detail-value">{planLabel(ws.plan)}</span>
														</div>
														<div class="adm-ws-detail-item">
															<span class="adm-ws-detail-label">Members</span>
															<span class="adm-ws-detail-value">{ws.member_count}</span>
														</div>
														<div class="adm-ws-detail-item">
															<span class="adm-ws-detail-label">Created</span>
															<span class="adm-ws-detail-value">{formatDate(ws.created_at)}</span>
														</div>
													</div>
												</div>
											</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}

			<!-- ── Settings Tab ────────────────────────────────────────────── -->
			{#if activeTab === 'settings'}
				<div class="adm-section-header mb-3">
					<span class="adm-section-title">Platform Settings</span>
					<span class="adm-muted-label">Read-only view</span>
				</div>

				{#if settingsLoading}
					<div class="adm-settings-grid">
						{#each Array(6) as _}
							<div class="adm-settings-card">
								<div class="adm-skel adm-skel--sm mb-3 w-1/3"></div>
								<div class="adm-skel adm-skel--sm mb-2"></div>
								<div class="adm-skel adm-skel--sm w-2/3"></div>
							</div>
						{/each}
					</div>
				{:else if settingsUnavailable || !settings}
					<div class="adm-empty adm-empty--notice">
						<svg class="w-5 h-5 adm-empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
								d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="adm-empty-text">Settings endpoint is not available yet.</p>
					</div>
				{:else}
					<div class="adm-settings-grid">
						<!-- Environment & Deployment -->
						<div class="adm-settings-card">
							<div class="adm-settings-card-title">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
								</svg>
								Environment
							</div>
							<div class="adm-settings-rows">
								<div class="adm-settings-row">
									<span class="adm-settings-key">Environment</span>
									<span class="adm-settings-val">
										<span class="adm-env-badge adm-env-badge--{settings.environment}">
											{settings.environment}
										</span>
									</span>
								</div>
								<div class="adm-settings-row">
									<span class="adm-settings-key">Deployment Mode</span>
									<code class="adm-slug">{settings.deployment_mode}</code>
								</div>
							</div>
						</div>

						<!-- AI Configuration -->
						<div class="adm-settings-card">
							<div class="adm-settings-card-title">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
								</svg>
								AI Configuration
							</div>
							<div class="adm-settings-rows">
								<div class="adm-settings-row">
									<span class="adm-settings-key">Provider</span>
									<span class="adm-settings-val">{settings.ai_provider || '—'}</span>
								</div>
								<div class="adm-settings-row">
									<span class="adm-settings-key">Default Model</span>
									<code class="adm-slug">{settings.default_model || '—'}</code>
								</div>
							</div>
						</div>

						<!-- Allowed Origins -->
						<div class="adm-settings-card adm-settings-card--wide">
							<div class="adm-settings-card-title">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
								</svg>
								Allowed Origins
							</div>
							{#if settings.allowed_origins && settings.allowed_origins.length > 0}
								<div class="adm-origins-list">
									{#each settings.allowed_origins as origin}
										<code class="adm-slug">{origin}</code>
									{/each}
								</div>
							{:else}
								<span class="adm-settings-val adm-settings-empty">No origins configured</span>
							{/if}
						</div>

						<!-- Feature Flags -->
						<div class="adm-settings-card adm-settings-card--wide">
							<div class="adm-settings-card-title">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
										d="M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9" />
								</svg>
								Feature Flags
							</div>
							<div class="adm-settings-rows">
								<div class="adm-settings-row">
									<span class="adm-settings-key">OSA</span>
									<span class="adm-flag-badge adm-flag-badge--{settings.feature_flags.osa_enabled ? 'on' : 'off'}">
										{settings.feature_flags.osa_enabled ? 'Enabled' : 'Disabled'}
									</span>
								</div>
								<div class="adm-settings-row">
									<span class="adm-settings-key">Carrier</span>
									<span class="adm-flag-badge adm-flag-badge--{settings.feature_flags.carrier_enabled ? 'on' : 'off'}">
										{settings.feature_flags.carrier_enabled ? 'Enabled' : 'Disabled'}
									</span>
								</div>
								<div class="adm-settings-row">
									<span class="adm-settings-key">NATS</span>
									<span class="adm-flag-badge adm-flag-badge--{settings.feature_flags.nats_enabled ? 'on' : 'off'}">
										{settings.feature_flags.nats_enabled ? 'Enabled' : 'Disabled'}
									</span>
								</div>
							</div>
						</div>
					</div>
				{/if}
			{/if}

			<!-- ── Audit Log Tab ───────────────────────────────────────────── -->
			{#if activeTab === 'audit'}
				<div class="adm-section-header mb-3">
					<span class="adm-section-title">Audit Log</span>
					<span class="adm-muted-label">Recent admin actions</span>
				</div>

				{#if auditLoading}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									{#each ['Actor', 'Action', 'Target', 'Time'] as col}
										<th class="adm-th">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each Array(5) as _}
									<tr class="adm-tr">
										{#each Array(4) as _}
											<td class="adm-td"><div class="adm-skel adm-skel--sm"></div></td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else if auditUnavailable}
					<div class="adm-empty adm-empty--notice">
						<svg class="w-5 h-5 adm-empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
								d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="adm-empty-text">Audit log endpoint is not available yet.</p>
					</div>
				{:else if auditEntries.length === 0}
					<div class="adm-empty">
						<p class="adm-empty-text">No audit entries found.</p>
					</div>
				{:else}
					<div class="adm-table-wrap">
						<table class="adm-table">
							<thead>
								<tr>
									<th class="adm-th">Actor</th>
									<th class="adm-th">Action</th>
									<th class="adm-th">Target</th>
									<th class="adm-th">Time</th>
								</tr>
							</thead>
							<tbody>
								{#each auditEntries as entry, i}
									<tr class="adm-tr" class:adm-tr--alt={i % 2 === 1}>
										<td class="adm-td adm-td--name">{entry.actor}</td>
										<td class="adm-td">
											<code class="adm-slug">{entry.action}</code>
										</td>
										<td class="adm-td adm-td--muted">{entry.target}</td>
										<td class="adm-td adm-td--muted">{formatDate(entry.time)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}

		</div>
	</div>
</div>

<style>
	/* ── Page shell ──────────────────────────────────────────────────────────── */
	.adm-page {
		background: var(--dbg);
	}

	/* ── Error banners ───────────────────────────────────────────────────────── */
	.adm-error-banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
		padding: 10px 14px;
		border-radius: 8px;
		border: 1px solid #fecaca;
		background: #fef2f2;
		color: #dc2626;
		font-size: 0.8125rem;
	}
	:global(.dark) .adm-error-banner {
		border-color: #991b1b;
		background: rgba(127, 29, 29, 0.25);
		color: #fca5a5;
	}
	.adm-error-dismiss {
		font-size: 0.75rem;
		text-decoration: underline;
		opacity: 0.75;
		flex-shrink: 0;
	}
	.adm-error-dismiss:hover { opacity: 1; }
	.adm-header {
		border-bottom: 1px solid var(--dbd);
	}
	.adm-title {
		color: var(--dt);
		font-size: 1.25rem;
		font-weight: 600;
	}
	.adm-muted {
		color: var(--dt3);
		font-size: 0.8125rem;
	}

	/* ── Tab bar ─────────────────────────────────────────────────────────────── */
	.adm-tab-row {
		display: flex;
		gap: 4px;
		border-bottom: 1px solid var(--dbd);
		overflow-x: auto;
		scrollbar-width: none;
	}
	.adm-tab-row::-webkit-scrollbar { display: none; }

	.adm-tab {
		padding: 8px 14px;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt3);
		border-bottom: 2px solid transparent;
		background: none;
		border-top: none;
		border-left: none;
		border-right: none;
		cursor: pointer;
		white-space: nowrap;
		flex-shrink: 0;
		margin-bottom: -1px;
		transition: color 150ms ease;
	}
	.adm-tab:hover {
		color: var(--dt2);
		border-radius: 6px 6px 0 0;
	}
	.adm-tab--active {
		color: var(--dt);
		font-weight: 600;
		border-bottom-color: var(--dt);
	}

	/* ── Stat cards ──────────────────────────────────────────────────────────── */
	.adm-grid-4 {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}
	@media (max-width: 900px) {
		.adm-grid-4 { grid-template-columns: repeat(2, 1fr); }
	}
	@media (max-width: 560px) {
		.adm-grid-4 { grid-template-columns: 1fr; }
	}

	.adm-stat-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		padding: 20px;
	}

	.adm-stat-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 8px;
		margin-bottom: 12px;
	}
	.adm-stat-icon--blue  { background: rgba(59,130,246,0.12); color: #3b82f6; }
	.adm-stat-icon--purple { background: rgba(139,92,246,0.12); color: #8b5cf6; }
	.adm-stat-icon--teal  { background: rgba(20,184,166,0.12); color: #14b8a6; }
	.adm-stat-icon--green { background: rgba(34,197,94,0.12); color: #22c55e; }
	:global(.dark) .adm-stat-icon--blue   { background: rgba(59,130,246,0.15); }
	:global(.dark) .adm-stat-icon--purple { background: rgba(139,92,246,0.15); }
	:global(.dark) .adm-stat-icon--teal   { background: rgba(20,184,166,0.15); }
	:global(.dark) .adm-stat-icon--green  { background: rgba(34,197,94,0.15); }

	.adm-stat-value {
		font-size: 1.875rem;
		font-weight: 700;
		color: var(--dt);
		line-height: 1;
		margin-bottom: 4px;
	}
	.adm-stat-label {
		font-size: 0.8125rem;
		color: var(--dt3);
		margin-bottom: 8px;
	}
	.adm-stat-sub {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.75rem;
		color: var(--dt3);
	}

	/* Status dots */
	.adm-dot {
		display: inline-block;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.adm-dot--green  { background: #22c55e; }
	.adm-dot--yellow { background: #f59e0b; }
	.adm-dot--gray   { background: #9ca3af; }
	.adm-dot--red    { background: #ef4444; }

	/* ── Section header ──────────────────────────────────────────────────────── */
	.adm-section-header {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.adm-section-title {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--dt);
	}
	.adm-count-badge {
		font-size: 0.75rem;
		color: var(--dt3);
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 20px;
		padding: 2px 8px;
	}
	.adm-muted-label {
		font-size: 0.8125rem;
		color: var(--dt3);
	}

	/* ── Tables ──────────────────────────────────────────────────────────────── */
	.adm-table-wrap {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		overflow: hidden;
		overflow-x: auto;
	}
	.adm-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.8125rem;
	}
	.adm-th {
		padding: 10px 14px;
		text-align: left;
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dt3);
		border-bottom: 1px solid var(--dbd);
		background: var(--dbg2);
		white-space: nowrap;
	}
	.adm-th--right { text-align: right; }
	.adm-tr {
		transition: background 120ms ease;
	}
	.adm-tr:hover { background: var(--dbg3); }
	.adm-tr--alt { background: var(--dbg); }
	.adm-tr--alt:hover { background: var(--dbg3); }
	.adm-tr--clickable { cursor: pointer; }
	.adm-tr--expanded { background: var(--dbg3); }
	.adm-tr--detail { background: var(--dbg); }
	.adm-tr--detail:hover { background: var(--dbg); }
	.adm-td {
		padding: 10px 14px;
		color: var(--dt2);
		border-bottom: 1px solid var(--dbd2);
		white-space: nowrap;
	}
	.adm-tr:last-child .adm-td { border-bottom: none; }
	.adm-td--name  { color: var(--dt); font-weight: 500; }
	.adm-td--muted { color: var(--dt3); }
	.adm-td--mono  { font-family: ui-monospace, monospace; font-size: 0.8125rem; }
	.adm-td--right { text-align: right; }
	.adm-td--detail {
		padding: 0;
		border-bottom: 1px solid var(--dbd);
	}

	/* ── Badges ──────────────────────────────────────────────────────────────── */
	.adm-role-badge {
		display: inline-flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 20px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.adm-role-badge--admin       { background: rgba(139,92,246,0.12); color: #8b5cf6; }
	.adm-role-badge--superadmin  { background: rgba(239,68,68,0.12);  color: #ef4444; }
	.adm-role-badge--user        { background: var(--dbg3); color: var(--dt3); }
	:global(.dark) .adm-role-badge--admin      { background: rgba(139,92,246,0.2); }
	:global(.dark) .adm-role-badge--superadmin { background: rgba(239,68,68,0.2); }

	.adm-plan-badge {
		display: inline-flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 500;
		padding: 2px 8px;
		border-radius: 20px;
		background: var(--dbg3);
		color: var(--dt2);
		border: 1px solid var(--dbd);
	}

	.adm-status-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 0.75rem;
		font-weight: 500;
		padding: 3px 8px;
		border-radius: 20px;
	}
	.adm-status-badge--running   { background: rgba(34,197,94,0.1);  color: #16a34a; }
	.adm-status-badge--hibernated { background: rgba(245,158,11,0.1); color: #d97706; }
	.adm-status-badge--stopped   { background: var(--dbg3); color: var(--dt3); }
	.adm-status-badge--error     { background: rgba(239,68,68,0.1);  color: #dc2626; }
	:global(.dark) .adm-status-badge--running   { background: rgba(34,197,94,0.15);  color: #4ade80; }
	:global(.dark) .adm-status-badge--hibernated { background: rgba(245,158,11,0.15); color: #fbbf24; }
	:global(.dark) .adm-status-badge--error     { background: rgba(239,68,68,0.15);  color: #f87171; }

	.adm-status-dot {
		display: inline-block;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.adm-status-dot--running    { background: #22c55e; }
	.adm-status-dot--hibernated { background: #f59e0b; }
	.adm-status-dot--stopped    { background: #9ca3af; }
	.adm-status-dot--error      { background: #ef4444; }

	/* ── Environment badge ───────────────────────────────────────────────────── */
	.adm-env-badge {
		display: inline-flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 20px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.adm-env-badge--production { background: rgba(239,68,68,0.12); color: #dc2626; }
	.adm-env-badge--staging    { background: rgba(245,158,11,0.12); color: #d97706; }
	.adm-env-badge--development,
	.adm-env-badge--local      { background: rgba(34,197,94,0.12); color: #16a34a; }
	:global(.dark) .adm-env-badge--production { background: rgba(239,68,68,0.2); color: #f87171; }
	:global(.dark) .adm-env-badge--staging    { background: rgba(245,158,11,0.2); color: #fbbf24; }
	:global(.dark) .adm-env-badge--development,
	:global(.dark) .adm-env-badge--local      { background: rgba(34,197,94,0.2); color: #4ade80; }

	/* ── Feature flag badge ──────────────────────────────────────────────────── */
	.adm-flag-badge {
		display: inline-flex;
		align-items: center;
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 20px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.adm-flag-badge--on  { background: rgba(34,197,94,0.12); color: #16a34a; }
	.adm-flag-badge--off { background: var(--dbg3); color: var(--dt3); }
	:global(.dark) .adm-flag-badge--on { background: rgba(34,197,94,0.2); color: #4ade80; }

	/* ── Slug code ───────────────────────────────────────────────────────────── */
	.adm-slug {
		font-family: ui-monospace, monospace;
		font-size: 0.75rem;
		background: var(--dbg3);
		color: var(--dt2);
		padding: 2px 6px;
		border-radius: 4px;
		border: 1px solid var(--dbd);
	}

	/* ── Action dropdown ─────────────────────────────────────────────────────── */
	.adm-action-wrap {
		position: relative;
		display: inline-flex;
		justify-content: flex-end;
	}
	.adm-action-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 6px;
		border: none;
		background: none;
		color: var(--dt3);
		cursor: pointer;
		transition: background 120ms ease, color 120ms ease;
	}
	.adm-action-btn:hover {
		background: var(--dbg3);
		color: var(--dt);
	}
	.adm-dropdown {
		position: absolute;
		top: calc(100% + 4px);
		right: 0;
		min-width: 160px;
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 8px;
		box-shadow: 0 4px 20px rgba(0,0,0,0.12);
		z-index: 100;
		overflow: hidden;
	}
	:global(.dark) .adm-dropdown {
		box-shadow: 0 4px 20px rgba(0,0,0,0.4);
	}
	.adm-dropdown-label {
		padding: 5px 12px 3px;
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dt3);
	}
	.adm-dropdown-item {
		display: flex;
		align-items: center;
		gap: 7px;
		width: 100%;
		text-align: left;
		padding: 8px 12px;
		font-size: 0.8125rem;
		color: var(--dt2);
		background: none;
		border: none;
		cursor: pointer;
		transition: background 100ms ease;
		text-decoration: none;
	}
	.adm-dropdown-item:hover {
		background: var(--dbg2);
		color: var(--dt);
	}
	.adm-dropdown-item:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.adm-dropdown-item--active {
		color: var(--dt);
		font-weight: 500;
	}
	.adm-dropdown-item--danger {
		color: #dc2626;
	}
	.adm-dropdown-item--danger:hover {
		background: rgba(239,68,68,0.08);
		color: #dc2626;
	}
	:global(.dark) .adm-dropdown-item--danger { color: #f87171; }
	:global(.dark) .adm-dropdown-item--danger:hover { background: rgba(239,68,68,0.15); color: #f87171; }
	.adm-dropdown-check {
		margin-left: auto;
		color: var(--dt3);
		flex-shrink: 0;
	}
	.adm-dropdown-sep {
		height: 1px;
		background: var(--dbd);
		margin: 2px 0;
	}

	/* ── Workspace expand ────────────────────────────────────────────────────── */
	.adm-ws-name-cell {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.adm-ws-chevron {
		width: 14px;
		height: 14px;
		color: var(--dt3);
		flex-shrink: 0;
		transition: transform 150ms ease;
	}
	.adm-ws-chevron--open {
		transform: rotate(90deg);
	}
	.adm-ws-detail {
		padding: 14px 18px 14px 34px;
		background: var(--dbg);
		border-top: 1px solid var(--dbd2);
	}
	.adm-ws-detail-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 20px;
	}
	.adm-ws-detail-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 140px;
	}
	.adm-ws-detail-label {
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--dt3);
	}
	.adm-ws-detail-value {
		font-size: 0.8125rem;
		color: var(--dt2);
	}

	/* ── Settings grid ───────────────────────────────────────────────────────── */
	.adm-settings-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 16px;
	}
	@media (max-width: 700px) {
		.adm-settings-grid { grid-template-columns: 1fr; }
	}
	.adm-settings-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		padding: 18px 20px;
	}
	.adm-settings-card--wide {
		grid-column: 1 / -1;
	}
	.adm-settings-card-title {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 14px;
	}
	.adm-settings-rows {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.adm-settings-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.adm-settings-key {
		font-size: 0.8125rem;
		color: var(--dt3);
		flex-shrink: 0;
	}
	.adm-settings-val {
		font-size: 0.8125rem;
		color: var(--dt2);
		text-align: right;
	}
	.adm-settings-empty {
		color: var(--dt3);
		font-style: italic;
	}
	.adm-origins-list {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	/* ── Empty states ────────────────────────────────────────────────────────── */
	.adm-empty {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 48px 24px;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 12px;
	}
	.adm-empty--notice {
		padding: 32px 24px;
		justify-content: flex-start;
	}
	.adm-empty-icon {
		color: var(--dt3);
		flex-shrink: 0;
	}
	.adm-empty-text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin: 0;
	}

	/* ── Skeletons ───────────────────────────────────────────────────────────── */
	.adm-skeleton-card {
		min-height: 120px;
	}
	.adm-skel {
		height: 12px;
		border-radius: 6px;
		background: var(--dbd);
		animation: adm-pulse 1.4s ease-in-out infinite;
	}
	.adm-skel--sm  { height: 10px; }
	.adm-skel--lg  { height: 28px; }
	@keyframes adm-pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
</style>
