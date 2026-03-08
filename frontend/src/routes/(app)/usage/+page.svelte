<script lang="ts">
	import { api, type UsageSummary, type ProviderUsage, type ModelUsage, type UsageTrendPoint, type AgentUsage, type MCPToolUsage } from '$lib/api';
	import { onMount } from 'svelte';
	import UsageFilters from './UsageFilters.svelte';
	import UsageDashboard from './UsageDashboard.svelte';
	import UsageProviderModels from './UsageProviderModels.svelte';
	import UsageTable from './UsageTable.svelte';
	import PlanUsageSection from './PlanUsageSection.svelte';

	// Usage analytics state
	let usageSummary = $state<UsageSummary | null>(null);
	let usageByProvider = $state<ProviderUsage[]>([]);
	let usageByModel = $state<ModelUsage[]>([]);
	let usageByAgent = $state<AgentUsage[]>([]);
	let usageTrend = $state<UsageTrendPoint[]>([]);
	let mcpUsage = $state<MCPToolUsage[]>([]);
	let usagePeriod = $state<'today' | 'week' | 'month' | 'all'>('month');
	let isLoading = $state(true);

	onMount(async () => {
		await loadUsageData();
	});

	async function loadUsageData() {
		isLoading = true;
		try {
			const [summary, providers, models, agents, trend, mcp] = await Promise.all([
				api.getUsageSummary(usagePeriod),
				api.getUsageByProvider(usagePeriod === 'all' ? 'year' : usagePeriod),
				api.getUsageByModel(usagePeriod === 'all' ? 'year' : usagePeriod),
				api.getUsageByAgent(usagePeriod === 'all' ? 'year' : usagePeriod).catch(() => []),
				api.getUsageTrend(),
				api.getMCPUsage(usagePeriod === 'all' ? 'year' : usagePeriod).catch(() => [])
			]);
			usageSummary = summary;
			usageByProvider = providers;
			usageByModel = models;
			usageByAgent = agents;
			usageTrend = trend;
			mcpUsage = mcp;
		} catch (error) {
			console.error('Error loading usage data:', error);
		} finally {
			isLoading = false;
		}
	}

	function formatNumber(num: number): string {
		if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
		if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
		return num.toString();
	}

	function formatCost(cost: number): string {
		return '$' + cost.toFixed(4);
	}

	function formatDuration(ms: number): string {
		if (ms < 1000) return `${Math.round(ms)}ms`;
		return `${(ms / 1000).toFixed(1)}s`;
	}

	function changePeriod(period: typeof usagePeriod) {
		usagePeriod = period;
		loadUsageData();
	}
</script>

<div class="usage-page">
	<!-- Plan quota section — always visible above analytics -->
	<PlanUsageSection />

	<UsageFilters {usagePeriod} {isLoading} onPeriodChange={changePeriod} />

	{#if !isLoading}
		{#if !usageSummary || usageSummary.total_requests === 0}
			<!-- Empty State -->
			<div class="empty-state">
				<div class="empty-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
					</svg>
				</div>
				<h3>No Usage Data Yet</h3>
				<p>Start chatting with the AI to see your usage analytics here.</p>
				<a href="/chat" class="start-chatting-btn">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-5 h-5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
					Start Chatting
				</a>
			</div>
		{:else}
			<UsageDashboard {usageSummary} {usageByProvider} {formatNumber} {formatCost} />
			<UsageProviderModels {usageByProvider} {usageByModel} {formatNumber} {formatCost} />
			<UsageTable
				{usageByAgent}
				{mcpUsage}
				{usageTrend}
				{formatNumber}
				{formatDuration}
			/>
		{/if}
	{/if}
</div>

<style>
	.usage-page {
		padding: 24px;
		max-width: 1400px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 120px 20px;
		text-align: center;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		border-radius: 20px;
		background: var(--color-bg-secondary, #f3f4f6);
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 20px;
	}

	:global(.dark) .empty-icon {
		background: #1f1f1f;
	}

	.empty-icon svg {
		width: 40px;
		height: 40px;
		color: var(--color-text-muted, #9ca3af);
	}

	:global(.dark) .empty-icon svg {
		color: #6b7280;
	}

	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--color-text, #111827);
		margin: 0 0 8px;
	}

	:global(.dark) .empty-state h3 {
		color: #f9fafb;
	}

	.empty-state p {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
		margin: 0 0 24px;
	}

	:global(.dark) .empty-state p {
		color: #9ca3af;
	}

	.start-chatting-btn {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 24px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		border-radius: 10px;
		font-weight: 500;
		text-decoration: none;
		transition: transform 0.15s, box-shadow 0.15s;
	}

	.start-chatting-btn:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
	}

	.start-chatting-btn svg {
		width: 20px;
		height: 20px;
	}
</style>
