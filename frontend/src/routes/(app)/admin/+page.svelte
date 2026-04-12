<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ── Types ──────────────────────────────────────────────────────────────────

	type TabId = 'overview' | 'users' | 'computers' | 'audit';

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

	// Audit
	let auditEntries = $state<AuditEntry[]>([]);
	let auditLoading = $state(true);
	let auditUnavailable = $state(false);

	// Action menu
	let openMenuId = $state<string | null>(null);
	let actionBusy = $state(false);

	let refreshInterval: ReturnType<typeof setInterval> | null = null;

	// ── Derived ────────────────────────────────────────────────────────────────

	const tabs: Array<{ id: TabId; label: string }> = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'users', label: 'Users' },
		{ id: 'computers', label: 'Computers' },
		{ id: 'audit', label: 'Audit Log' },
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
		openMenuId = null;
		await adminPost(`/admin/users/${userId}/role`, { role });
		await loadUsers();
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
		await Promise.all([loadDashboard(), loadUsers(), loadComputers(), loadAudit()]);

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
									{#each ['Workspace', 'Owner', 'Slug', 'Status', 'Plan', 'RAM', 'CPUs', 'Created'] as col}
										<th class="adm-th">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each Array(5) as _}
									<tr class="adm-tr">
										{#each Array(8) as _}
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
									</tr>
								{/each}
							</tbody>
						</table>
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
		min-width: 150px;
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
	.adm-dropdown-item {
		display: block;
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
	.adm-dropdown-sep {
		height: 1px;
		background: var(--dbd);
		margin: 2px 0;
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
