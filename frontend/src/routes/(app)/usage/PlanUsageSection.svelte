<script lang="ts">
	import { onMount } from 'svelte';
	import { usageStore } from '$lib/stores/usageStore.svelte';

	onMount(() => {
		usageStore.fetchUsage();
	});

	// ── Formatting helpers ─────────────────────────────────────────────────────

	function formatBytes(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
		return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
	}

	function formatLimit(value: number, format: 'bytes' | 'number' = 'number'): string {
		if (value === -1) return 'Unlimited';
		return format === 'bytes' ? formatBytes(value) : value.toLocaleString();
	}

	function planColor(plan: string): string {
		switch (plan.toLowerCase()) {
			case 'pro':
				return 'plan-badge--pro';
			case 'enterprise':
				return 'plan-badge--enterprise';
			default:
				return 'plan-badge--free';
		}
	}

	/** Returns Tailwind-compatible bar-color class based on percentage */
	function barColor(pct: number): string {
		if (pct >= 95) return 'bar--red';
		if (pct >= 80) return 'bar--yellow';
		return 'bar--green';
	}

	/** Returns the status label shown inside/below the progress bar */
	function statusLabel(pct: number, isUnlimited: boolean): string {
		if (isUnlimited) return 'Unlimited';
		if (pct >= 100) return 'Limit reached';
		if (pct >= 95) return 'Critical';
		if (pct >= 80) return 'High usage';
		return 'Good';
	}

	// ── Derived values ─────────────────────────────────────────────────────────

	let planUsage = $derived(usageStore.planUsage);
	let planLimits = $derived(usageStore.planLimits);

	let aiPct = $derived(usageStore.getUsagePercentage('ai_calls'));
	let storagePct = $derived(usageStore.getUsagePercentage('storage'));
	let modulesPct = $derived(usageStore.getUsagePercentage('modules'));
	let teamPct = $derived(usageStore.getUsagePercentage('team_members'));

	let showUpgradeBanner = $derived(usageStore.isAnyMetricAbove(80));

	// Plan comparison table rows
	const planComparisonRows: Array<{
		label: string;
		free: string;
		pro: string;
		enterprise: string;
	}> = [
		{ label: 'AI calls / day', free: '100', pro: '5,000', enterprise: 'Unlimited' },
		{ label: 'AI model tier', free: 'Standard', pro: 'Advanced', enterprise: 'Elite' },
		{ label: 'Storage', free: '1 GB', pro: '10 GB', enterprise: '100 GB' },
		{ label: 'Modules', free: '3', pro: '25', enterprise: 'Unlimited' },
		{ label: 'Team members', free: '1', pro: '5', enterprise: 'Unlimited' },
		{ label: 'OSA modes', free: 'Basic', pro: 'All 5', enterprise: 'All 5 + Custom' },
		{ label: 'Compute CPU hrs / mo', free: '10', pro: '100', enterprise: 'Unlimited' },
	];
</script>

{#if usageStore.isLoading}
	<div class="plan-section">
		<div class="section-skeleton">
			<div class="skeleton-header"></div>
			<div class="skeleton-grid">
				{#each [1, 2, 3, 4] as _}
					<div class="skeleton-card"></div>
				{/each}
			</div>
		</div>
	</div>
{:else if planUsage}
	<!-- ── Plan header ───────────────────────────────────────────────────────── -->
	<div class="plan-section">
		<div class="plan-header">
			<div class="plan-header-left">
				<h2 class="plan-title">Plan &amp; Usage</h2>
				<p class="plan-subtitle">Your current quota usage resets daily for AI calls.</p>
			</div>
			<div class="plan-header-right">
				<span class="plan-badge {planColor(planUsage.plan)}">{planUsage.plan}</span>
				{#if planUsage.plan.toLowerCase() !== 'enterprise'}
					<a href="/settings/billing" class="upgrade-btn" aria-label="Upgrade your plan">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M4.5 10.5L12 3m0 0l7.5 7.5M12 3v18" />
						</svg>
						Upgrade Plan
					</a>
				{/if}
			</div>
		</div>

		<!-- ── Upgrade banner (conditional) ────────────────────────────────────── -->
		{#if showUpgradeBanner && planUsage.plan.toLowerCase() !== 'enterprise'}
			<div class="upgrade-banner" role="alert">
				<div class="upgrade-banner-icon" aria-hidden="true">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
					</svg>
				</div>
				<div class="upgrade-banner-body">
					<p class="upgrade-banner-title">You're approaching your plan limits</p>
					<p class="upgrade-banner-desc">
						Upgrade to Pro for 50x more AI calls, 10 GB storage, up to 25 modules, and all 5 OSA modes.
					</p>
				</div>
				<a href="/settings/billing" class="upgrade-banner-cta" aria-label="Upgrade to Pro now">
					Upgrade Now
				</a>
			</div>
		{/if}

		<!-- ── Usage cards grid ─────────────────────────────────────────────────── -->
		<div class="usage-grid">
			<!-- AI Calls card -->
			<div class="usage-card" class:usage-card--warning={aiPct >= 80 && aiPct < 95} class:usage-card--critical={aiPct >= 95}>
				<div class="usage-card-header">
					<div class="usage-card-icon usage-card-icon--ai" aria-hidden="true">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
						</svg>
					</div>
					<div>
						<h3 class="usage-card-title">AI Calls Today</h3>
						<p class="usage-card-detail">
							{planUsage.ai_calls_limit === -1
								? `${planUsage.ai_calls_today.toLocaleString()} used`
								: `${planUsage.ai_calls_today.toLocaleString()} / ${planUsage.ai_calls_limit.toLocaleString()}`}
						</p>
					</div>
					{#if planUsage.ai_calls_limit !== -1}
						<span class="usage-pct" class:usage-pct--warn={aiPct >= 80} class:usage-pct--crit={aiPct >= 95}>
							{aiPct}%
						</span>
					{/if}
				</div>
				{#if planUsage.ai_calls_limit !== -1}
					<div class="progress-track" role="progressbar" aria-valuenow={aiPct} aria-valuemin={0} aria-valuemax={100} aria-label="AI calls usage">
						<div class="progress-bar {barColor(aiPct)}" style="width: {aiPct}%"></div>
					</div>
					<p class="usage-status">{statusLabel(aiPct, false)}</p>
				{:else}
					<p class="usage-status usage-status--unlimited">Unlimited</p>
				{/if}
				{#if usageStore.isOverLimit('ai_calls')}
					<a href="/settings/billing" class="card-upgrade-cta" aria-label="Upgrade to get more AI calls">Upgrade for more AI calls</a>
				{/if}
			</div>

			<!-- Storage card -->
			<div class="usage-card" class:usage-card--warning={storagePct >= 80 && storagePct < 95} class:usage-card--critical={storagePct >= 95}>
				<div class="usage-card-header">
					<div class="usage-card-icon usage-card-icon--storage" aria-hidden="true">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 5.625c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
						</svg>
					</div>
					<div>
						<h3 class="usage-card-title">Storage</h3>
						<p class="usage-card-detail">
							{formatBytes(planUsage.storage_used_bytes)} / {formatLimit(planUsage.storage_limit_bytes, 'bytes')}
						</p>
					</div>
					{#if planUsage.storage_limit_bytes !== -1}
						<span class="usage-pct" class:usage-pct--warn={storagePct >= 80} class:usage-pct--crit={storagePct >= 95}>
							{storagePct}%
						</span>
					{/if}
				</div>
				{#if planUsage.storage_limit_bytes !== -1}
					<div class="progress-track" role="progressbar" aria-valuenow={storagePct} aria-valuemin={0} aria-valuemax={100} aria-label="Storage usage">
						<div class="progress-bar {barColor(storagePct)}" style="width: {storagePct}%"></div>
					</div>
					<p class="usage-status">{statusLabel(storagePct, false)}</p>
				{:else}
					<p class="usage-status usage-status--unlimited">Unlimited</p>
				{/if}
				{#if usageStore.isOverLimit('storage')}
					<a href="/settings/billing" class="card-upgrade-cta" aria-label="Upgrade to get more storage">Upgrade for more storage</a>
				{/if}
			</div>

			<!-- Modules card -->
			<div class="usage-card" class:usage-card--warning={modulesPct >= 80 && modulesPct < 95} class:usage-card--critical={modulesPct >= 95}>
				<div class="usage-card-header">
					<div class="usage-card-icon usage-card-icon--modules" aria-hidden="true">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
						</svg>
					</div>
					<div>
						<h3 class="usage-card-title">Modules</h3>
						<p class="usage-card-detail">
							{planUsage.modules_count} / {formatLimit(planUsage.modules_limit)} active
						</p>
					</div>
					{#if planUsage.modules_limit !== -1}
						<span class="usage-pct" class:usage-pct--warn={modulesPct >= 80} class:usage-pct--crit={modulesPct >= 95}>
							{modulesPct}%
						</span>
					{/if}
				</div>
				{#if planUsage.modules_limit !== -1}
					<div class="progress-track" role="progressbar" aria-valuenow={modulesPct} aria-valuemin={0} aria-valuemax={100} aria-label="Modules usage">
						<div class="progress-bar {barColor(modulesPct)}" style="width: {modulesPct}%"></div>
					</div>
					<p class="usage-status">{statusLabel(modulesPct, false)}</p>
				{:else}
					<p class="usage-status usage-status--unlimited">Unlimited</p>
				{/if}
				{#if usageStore.isOverLimit('modules')}
					<a href="/settings/billing" class="card-upgrade-cta" aria-label="Upgrade to add more modules">Upgrade for more modules</a>
				{/if}
			</div>

			<!-- Team members card -->
			<div class="usage-card" class:usage-card--warning={teamPct >= 80 && teamPct < 95} class:usage-card--critical={teamPct >= 95}>
				<div class="usage-card-header">
					<div class="usage-card-icon usage-card-icon--team" aria-hidden="true">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
						</svg>
					</div>
					<div>
						<h3 class="usage-card-title">Team Members</h3>
						<p class="usage-card-detail">
							{planUsage.team_members} / {formatLimit(planUsage.team_members_limit)} members
						</p>
					</div>
					{#if planUsage.team_members_limit !== -1}
						<span class="usage-pct" class:usage-pct--warn={teamPct >= 80} class:usage-pct--crit={teamPct >= 95}>
							{teamPct}%
						</span>
					{/if}
				</div>
				{#if planUsage.team_members_limit !== -1}
					<div class="progress-track" role="progressbar" aria-valuenow={teamPct} aria-valuemin={0} aria-valuemax={100} aria-label="Team members usage">
						<div class="progress-bar {barColor(teamPct)}" style="width: {teamPct}%"></div>
					</div>
					<p class="usage-status">{statusLabel(teamPct, false)}</p>
				{:else}
					<p class="usage-status usage-status--unlimited">Unlimited</p>
				{/if}
				{#if usageStore.isOverLimit('team_members')}
					<a href="/settings/billing" class="card-upgrade-cta" aria-label="Upgrade to add more team members">Upgrade for more seats</a>
				{/if}
			</div>
		</div>

		<!-- ── Plan comparison table ────────────────────────────────────────────── -->
		<div class="comparison-wrapper">
			<h3 class="comparison-title">Plan Comparison</h3>
			<div class="comparison-scroll">
				<table class="comparison-table" aria-label="Plan comparison">
					<thead>
						<tr>
							<th class="comparison-col-label" scope="col"></th>
							{#each ['Free', 'Pro', 'Enterprise'] as tier}
								<th
									scope="col"
									class="comparison-col-tier"
									class:comparison-col-tier--current={planUsage.plan.toLowerCase() === tier.toLowerCase()}
								>
									<div class="tier-header">
										<span class="tier-name">{tier}</span>
										{#if planUsage.plan.toLowerCase() === tier.toLowerCase()}
											<span class="current-badge">Current</span>
										{/if}
										{#if tier !== planUsage.plan && tier !== 'Free'}
											<a
												href="/settings/billing"
												class="tier-upgrade-btn"
												aria-label="Upgrade to {tier}"
											>
												Upgrade
											</a>
										{/if}
									</div>
								</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each planComparisonRows as row}
							<tr class="comparison-row">
								<td class="comparison-label">{row.label}</td>
								<td class="comparison-value" class:comparison-value--current={planUsage.plan.toLowerCase() === 'free'}>{row.free}</td>
								<td class="comparison-value" class:comparison-value--current={planUsage.plan.toLowerCase() === 'pro'}>{row.pro}</td>
								<td class="comparison-value" class:comparison-value--current={planUsage.plan.toLowerCase() === 'enterprise'}>{row.enterprise}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ── Section shell ─────────────────────────────────────────────────────── */
	.plan-section {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	/* ── Plan header ───────────────────────────────────────────────────────── */
	.plan-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		gap: 12px;
	}

	.plan-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		margin: 0;
	}

	:global(.dark) .plan-title {
		color: #f9fafb;
	}

	.plan-subtitle {
		font-size: 0.8125rem;
		color: var(--color-text-secondary, #6b7280);
		margin: 4px 0 0;
	}

	:global(.dark) .plan-subtitle {
		color: #9ca3af;
	}

	.plan-header-right {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* ── Plan badge ────────────────────────────────────────────────────────── */
	.plan-badge {
		display: inline-flex;
		align-items: center;
		padding: 5px 12px;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.6px;
	}

	.plan-badge--free {
		background: #f3f4f6;
		color: #374151;
	}

	:global(.dark) .plan-badge--free {
		background: #27272a;
		color: #d1d5db;
	}

	.plan-badge--pro {
		background: #e0e7ff;
		color: #3730a3;
	}

	:global(.dark) .plan-badge--pro {
		background: rgba(99, 102, 241, 0.2);
		color: #a5b4fc;
	}

	.plan-badge--enterprise {
		background: linear-gradient(135deg, #fef3c7, #fde68a);
		color: #92400e;
	}

	:global(.dark) .plan-badge--enterprise {
		background: rgba(245, 158, 11, 0.2);
		color: #fbbf24;
	}

	/* ── Upgrade button (header) ───────────────────────────────────────────── */
	.upgrade-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border-radius: 8px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		transition: opacity 0.15s, transform 0.15s;
	}

	.upgrade-btn:hover {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	.upgrade-btn svg {
		width: 14px;
		height: 14px;
	}

	/* ── Upgrade banner ────────────────────────────────────────────────────── */
	.upgrade-banner {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 20px;
		background: #fef9ec;
		border: 1px solid #fde68a;
		border-radius: 12px;
		flex-wrap: wrap;
	}

	:global(.dark) .upgrade-banner {
		background: rgba(245, 158, 11, 0.08);
		border-color: rgba(245, 158, 11, 0.3);
	}

	.upgrade-banner-icon {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: #fde68a;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		color: #92400e;
	}

	:global(.dark) .upgrade-banner-icon {
		background: rgba(245, 158, 11, 0.2);
		color: #fbbf24;
	}

	.upgrade-banner-icon svg {
		width: 18px;
		height: 18px;
	}

	.upgrade-banner-body {
		flex: 1;
		min-width: 0;
	}

	.upgrade-banner-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: #92400e;
		margin: 0 0 2px;
	}

	:global(.dark) .upgrade-banner-title {
		color: #fbbf24;
	}

	.upgrade-banner-desc {
		font-size: 0.8125rem;
		color: #b45309;
		margin: 0;
	}

	:global(.dark) .upgrade-banner-desc {
		color: #d97706;
	}

	.upgrade-banner-cta {
		display: inline-flex;
		padding: 8px 18px;
		border-radius: 8px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		font-size: 0.875rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
		transition: opacity 0.15s;
	}

	.upgrade-banner-cta:hover {
		opacity: 0.9;
	}

	/* ── Usage cards grid ──────────────────────────────────────────────────── */
	.usage-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	@media (max-width: 1100px) {
		.usage-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 560px) {
		.usage-grid {
			grid-template-columns: 1fr;
		}
	}

	.usage-card {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		transition: border-color 0.15s;
	}

	:global(.dark) .usage-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.usage-card--warning {
		border-color: #fde68a;
	}

	:global(.dark) .usage-card--warning {
		border-color: rgba(245, 158, 11, 0.4);
	}

	.usage-card--critical {
		border-color: #fca5a5;
	}

	:global(.dark) .usage-card--critical {
		border-color: rgba(239, 68, 68, 0.4);
	}

	.usage-card-header {
		display: flex;
		align-items: flex-start;
		gap: 12px;
	}

	.usage-card-icon {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.usage-card-icon svg {
		width: 20px;
		height: 20px;
	}

	.usage-card-icon--ai {
		background: #e0e7ff;
		color: #4338ca;
	}

	:global(.dark) .usage-card-icon--ai {
		background: rgba(99, 102, 241, 0.2);
		color: #a5b4fc;
	}

	.usage-card-icon--storage {
		background: #dcfce7;
		color: #15803d;
	}

	:global(.dark) .usage-card-icon--storage {
		background: rgba(22, 163, 74, 0.2);
		color: #86efac;
	}

	.usage-card-icon--modules {
		background: #fef3c7;
		color: #b45309;
	}

	:global(.dark) .usage-card-icon--modules {
		background: rgba(245, 158, 11, 0.2);
		color: #fcd34d;
	}

	.usage-card-icon--team {
		background: #fce7f3;
		color: #be185d;
	}

	:global(.dark) .usage-card-icon--team {
		background: rgba(236, 72, 153, 0.2);
		color: #f9a8d4;
	}

	.usage-card-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--color-text, #111827);
		margin: 0 0 2px;
	}

	:global(.dark) .usage-card-title {
		color: #f9fafb;
	}

	.usage-card-detail {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text-secondary, #6b7280);
		margin: 0;
	}

	:global(.dark) .usage-card-detail {
		color: #9ca3af;
	}

	/* percentage pill pinned to the right of the header */
	.usage-pct {
		margin-left: auto;
		font-size: 0.75rem;
		font-weight: 700;
		padding: 3px 8px;
		border-radius: 9999px;
		background: #dcfce7;
		color: #15803d;
		white-space: nowrap;
	}

	:global(.dark) .usage-pct {
		background: rgba(22, 163, 74, 0.2);
		color: #86efac;
	}

	.usage-pct--warn {
		background: #fef3c7;
		color: #b45309;
	}

	:global(.dark) .usage-pct--warn {
		background: rgba(245, 158, 11, 0.2);
		color: #fcd34d;
	}

	.usage-pct--crit {
		background: #fee2e2;
		color: #b91c1c;
	}

	:global(.dark) .usage-pct--crit {
		background: rgba(239, 68, 68, 0.2);
		color: #fca5a5;
	}

	/* ── Progress bar ──────────────────────────────────────────────────────── */
	.progress-track {
		width: 100%;
		height: 6px;
		background: #f3f4f6;
		border-radius: 9999px;
		overflow: hidden;
	}

	:global(.dark) .progress-track {
		background: #27272a;
	}

	.progress-bar {
		height: 100%;
		border-radius: 9999px;
		transition: width 0.5s ease;
	}

	.bar--green {
		background: #22c55e;
	}

	.bar--yellow {
		background: #eab308;
	}

	.bar--red {
		background: #ef4444;
	}

	.usage-status {
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
		margin: 0;
	}

	:global(.dark) .usage-status {
		color: #6b7280;
	}

	.usage-status--unlimited {
		color: #22c55e;
		font-weight: 500;
	}

	:global(.dark) .usage-status--unlimited {
		color: #4ade80;
	}

	/* ── Per-card upgrade CTA ──────────────────────────────────────────────── */
	.card-upgrade-cta {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 8px 14px;
		border-radius: 8px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		font-size: 0.8125rem;
		font-weight: 600;
		text-decoration: none;
		text-align: center;
		transition: opacity 0.15s;
	}

	.card-upgrade-cta:hover {
		opacity: 0.9;
	}

	/* ── Plan comparison table ─────────────────────────────────────────────── */
	.comparison-wrapper {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 24px;
	}

	:global(.dark) .comparison-wrapper {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.comparison-title {
		font-size: 1rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		margin: 0 0 20px;
	}

	:global(.dark) .comparison-title {
		color: #f9fafb;
	}

	.comparison-scroll {
		overflow-x: auto;
		-webkit-overflow-scrolling: touch;
	}

	.comparison-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.comparison-col-label {
		text-align: left;
		padding: 8px 12px 8px 0;
		width: 40%;
	}

	.comparison-col-tier {
		text-align: center;
		padding: 8px 12px;
		width: 20%;
	}

	.comparison-col-tier--current {
		background: linear-gradient(180deg, rgba(99, 102, 241, 0.06) 0%, transparent 100%);
		border-radius: 10px 10px 0 0;
	}

	:global(.dark) .comparison-col-tier--current {
		background: linear-gradient(180deg, rgba(99, 102, 241, 0.12) 0%, transparent 100%);
	}

	.tier-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding-bottom: 8px;
	}

	.tier-name {
		font-weight: 700;
		color: var(--color-text, #111827);
	}

	:global(.dark) .tier-name {
		color: #f9fafb;
	}

	.current-badge {
		font-size: 0.65rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.4px;
		padding: 2px 8px;
		border-radius: 9999px;
		background: #e0e7ff;
		color: #3730a3;
	}

	:global(.dark) .current-badge {
		background: rgba(99, 102, 241, 0.2);
		color: #a5b4fc;
	}

	.tier-upgrade-btn {
		font-size: 0.75rem;
		font-weight: 600;
		padding: 4px 12px;
		border-radius: 6px;
		background: #f3f4f6;
		color: #4338ca;
		text-decoration: none;
		transition: background 0.15s;
	}

	:global(.dark) .tier-upgrade-btn {
		background: #1f1f1f;
		color: #818cf8;
	}

	.tier-upgrade-btn:hover {
		background: #e0e7ff;
	}

	:global(.dark) .tier-upgrade-btn:hover {
		background: rgba(99, 102, 241, 0.2);
	}

	.comparison-row {
		border-top: 1px solid var(--color-border, #e5e7eb);
	}

	:global(.dark) .comparison-row {
		border-color: rgba(255, 255, 255, 0.06);
	}

	.comparison-label {
		padding: 12px 12px 12px 0;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .comparison-label {
		color: #9ca3af;
	}

	.comparison-value {
		padding: 12px;
		text-align: center;
		color: var(--color-text, #374151);
		font-variant-numeric: tabular-nums;
	}

	:global(.dark) .comparison-value {
		color: #d1d5db;
	}

	.comparison-value--current {
		font-weight: 600;
		color: #4338ca;
	}

	:global(.dark) .comparison-value--current {
		color: #a5b4fc;
	}

	/* ── Skeleton loading ──────────────────────────────────────────────────── */
	.section-skeleton {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.skeleton-header {
		height: 32px;
		width: 240px;
		border-radius: 8px;
		background: var(--color-bg-secondary, #f3f4f6);
		animation: shimmer 1.4s ease-in-out infinite;
	}

	:global(.dark) .skeleton-header {
		background: #1f1f1f;
	}

	.skeleton-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	@media (max-width: 1100px) {
		.skeleton-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	.skeleton-card {
		height: 120px;
		border-radius: 16px;
		background: var(--color-bg-secondary, #f3f4f6);
		animation: shimmer 1.4s ease-in-out infinite;
	}

	:global(.dark) .skeleton-card {
		background: #1f1f1f;
	}

	@keyframes shimmer {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}
</style>
