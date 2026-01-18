<script lang="ts">
	import { api, type UsageSummary, type ProviderUsage, type ModelUsage, type UsageTrendPoint, type AgentUsage, type MCPToolUsage } from '$lib/api';
	import { onMount } from 'svelte';

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
	<!-- Page Header -->
	<div class="page-header">
		<div class="header-content">
			<div class="header-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
				</svg>
			</div>
			<div>
				<h1 class="page-title">Usage Analytics</h1>
				<p class="page-subtitle">Track your AI usage, tokens, and estimated costs</p>
			</div>
		</div>
		<div class="btn-pill-group">
			{#each ['today', 'week', 'month', 'all'] as period}
				<button
					onclick={() => changePeriod(period as typeof usagePeriod)}
					class="btn-pill {usagePeriod === period ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				>
					{period === 'all' ? 'All Time' : period.charAt(0).toUpperCase() + period.slice(1)}
				</button>
			{/each}
		</div>
	</div>

	{#if isLoading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Loading analytics...</p>
		</div>
	{:else if !usageSummary || usageSummary.total_requests === 0}
		<!-- Empty State -->
		<div class="empty-state">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
				</svg>
			</div>
			<h3>No Usage Data Yet</h3>
			<p>Start chatting with the AI to see your usage analytics here.</p>
			<a href="/chat" class="btn-pill btn-pill-primary">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-5 h-5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
				</svg>
				Start Chatting
			</a>
		</div>
	{:else}
		<!-- Stats Grid -->
		<div class="stats-grid">
			<div class="stat-card requests">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
				</div>
				<div class="stat-content">
					<span class="stat-value">{formatNumber(usageSummary?.total_requests || 0)}</span>
					<span class="stat-label">Requests</span>
				</div>
			</div>

			<div class="stat-card tokens">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
					</svg>
				</div>
				<div class="stat-content">
					<span class="stat-value">{formatNumber(usageSummary?.total_tokens || 0)}</span>
					<span class="stat-label">Total Tokens</span>
				</div>
			</div>

			<div class="stat-card input">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
					</svg>
				</div>
				<div class="stat-content">
					<span class="stat-value">{formatNumber(usageSummary?.total_input_tokens || 0)}</span>
					<span class="stat-label">Input Tokens</span>
				</div>
			</div>

			<div class="stat-card output">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
					</svg>
				</div>
				<div class="stat-content">
					<span class="stat-value">{formatNumber(usageSummary?.total_output_tokens || 0)}</span>
					<span class="stat-label">Output Tokens</span>
				</div>
			</div>
		</div>

		<!-- Cost Analysis -->
		<div class="cost-analysis">
			<div class="cost-card spent">
				<div class="cost-header">
					<span class="cost-label">Estimated Cloud Cost</span>
					<span class="cost-badge">API Usage</span>
				</div>
				<span class="cost-value">{formatCost(usageSummary?.total_cost || 0)}</span>
				<p class="cost-note">Based on current provider pricing</p>
			</div>

			<div class="cost-card saved">
				<div class="cost-header">
					<span class="cost-label">Local Processing Savings</span>
					<span class="cost-badge saved">Saved</span>
				</div>
				<span class="cost-value">
					{formatCost((usageByProvider.find(p => p.provider === 'ollama')?.total_tokens || 0) * 0.00002)}
				</span>
				<p class="cost-note">Running {formatNumber(usageByProvider.find(p => p.provider === 'ollama')?.total_tokens || 0)} tokens locally</p>
			</div>
		</div>

		<!-- Two Column Layout for Provider and Models -->
		<div class="two-column-grid">
			<!-- Provider Breakdown -->
			{#if usageByProvider.length > 0}
				<div class="usage-section">
					<h3 class="section-title">Usage by Provider</h3>
					<div class="provider-list">
						{#each usageByProvider as provider}
							{@const maxTokens = Math.max(...usageByProvider.map(p => p.total_tokens))}
							{@const percentage = maxTokens > 0 ? (provider.total_tokens / maxTokens) * 100 : 0}
							<div class="provider-item">
								<div class="provider-info">
									<div class="provider-icon" class:local={provider.provider === 'ollama'} class:anthropic={provider.provider === 'anthropic'} class:groq={provider.provider === 'groq'} class:openai={provider.provider === 'openai'}>
										{#if provider.provider === 'ollama'}
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<rect x="2" y="3" width="20" height="14" rx="2"/>
												<path d="M8 21h8M12 17v4"/>
											</svg>
										{:else}
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
											</svg>
										{/if}
									</div>
									<div class="provider-details">
										<span class="provider-name">{provider.provider}</span>
										<span class="provider-type">{provider.provider === 'ollama' ? 'Local' : 'Cloud'}</span>
									</div>
								</div>
								<div class="provider-stats">
									<div class="provider-bar-container">
										<div class="provider-bar" class:local={provider.provider === 'ollama'} style="width: {percentage}%"></div>
									</div>
									<div class="provider-numbers">
										<span class="provider-tokens">{formatNumber(provider.total_tokens)} tokens</span>
										<span class="provider-cost">{provider.provider === 'ollama' ? 'Free' : formatCost(provider.total_cost)}</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Model Usage -->
			{#if usageByModel.length > 0}
				<div class="usage-section">
					<h3 class="section-title">Model Usage</h3>
					<div class="model-list">
						{#each usageByModel.slice(0, 6) as model}
							<div class="model-card">
								<div class="model-header">
									<span class="model-name">{model.model.split(':')[0]}</span>
									<span class="model-provider" class:local={model.provider === 'ollama'}>{model.provider}</span>
								</div>
								<div class="model-stats">
									<div class="model-stat">
										<span class="model-stat-value">{formatNumber(model.request_count)}</span>
										<span class="model-stat-label">requests</span>
									</div>
									<div class="model-stat">
										<span class="model-stat-value">{formatNumber(model.total_tokens)}</span>
										<span class="model-stat-label">tokens</span>
									</div>
								</div>
								<div class="model-cost">
									{model.provider === 'ollama' ? 'Free (Local)' : formatCost(model.total_cost)}
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<!-- Agent Usage -->
		{#if usageByAgent.length > 0}
			<div class="usage-section full-width">
				<h3 class="section-title">Agent Usage</h3>
				<div class="agent-grid">
					{#each usageByAgent as agent}
						<div class="agent-card">
							<div class="agent-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
							</div>
							<div class="agent-info">
								<span class="agent-name">{agent.agent_name}</span>
								<div class="agent-stats-row">
									<span>{formatNumber(agent.request_count)} calls</span>
									<span class="dot"></span>
									<span>{formatNumber(agent.total_tokens)} tokens</span>
									<span class="dot"></span>
									<span>{formatDuration(agent.avg_duration_ms)} avg</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- MCP Tool Usage -->
		{#if mcpUsage.length > 0}
			<div class="usage-section full-width">
				<h3 class="section-title">MCP Tool Usage</h3>
				<div class="mcp-grid">
					{#each mcpUsage as tool}
						{@const successRate = tool.request_count > 0 ? (tool.success_count / tool.request_count) * 100 : 0}
						<div class="mcp-card">
							<div class="mcp-header">
								<span class="mcp-name">{tool.tool_name}</span>
								{#if tool.server_name}
									<span class="mcp-server">{tool.server_name}</span>
								{/if}
							</div>
							<div class="mcp-stats">
								<div class="mcp-stat">
									<span class="mcp-stat-value">{tool.request_count}</span>
									<span class="mcp-stat-label">calls</span>
								</div>
								<div class="mcp-stat">
									<span class="mcp-stat-value success">{successRate.toFixed(0)}%</span>
									<span class="mcp-stat-label">success</span>
								</div>
								<div class="mcp-stat">
									<span class="mcp-stat-value">{formatDuration(tool.avg_duration_ms)}</span>
									<span class="mcp-stat-label">avg time</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Usage Trend Chart (Simple) -->
		{#if usageTrend.length > 0}
			{@const maxTokens = Math.max(...usageTrend.map(t => t.total_tokens), 1)}
			<div class="usage-section full-width">
				<h3 class="section-title">Usage Trend (Last 30 Days)</h3>
				<div class="trend-chart">
					{#each usageTrend.slice(-14) as point, i}
						{@const height = (point.total_tokens / maxTokens) * 100}
						<div class="trend-bar-container" title="{new Date(point.date).toLocaleDateString()}: {formatNumber(point.total_tokens)} tokens">
							<div class="trend-bar" style="height: {Math.max(height, 2)}%"></div>
							<span class="trend-label">{new Date(point.date).getDate()}</span>
						</div>
					{/each}
				</div>
				<div class="trend-legend">
					<span class="trend-legend-item">
						<span class="trend-legend-dot"></span>
						Daily Tokens
					</span>
				</div>
			</div>
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

	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		gap: 20px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--color-border, #e5e7eb);
	}

	:global(.dark) .page-header {
		border-color: rgba(255, 255, 255, 0.08);
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.header-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.header-icon svg {
		width: 24px;
		height: 24px;
		color: white;
	}

	.page-title {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		margin: 0;
	}

	:global(.dark) .page-title {
		color: #f9fafb;
	}

	.page-subtitle {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
		margin: 4px 0 0;
	}

	:global(.dark) .page-subtitle {
		color: #9ca3af;
	}


	/* Loading State */
	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 120px 20px;
		gap: 16px;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .loading-state {
		color: #9ca3af;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--color-border, #e5e7eb);
		border-top-color: #6366f1;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	:global(.dark) .loading-spinner {
		border-color: rgba(255, 255, 255, 0.1);
		border-top-color: #6366f1;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
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


	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	@media (max-width: 1024px) {
		.stats-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 480px) {
		.stats-grid { grid-template-columns: 1fr; }
	}

	.stat-card {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 24px;
		display: flex;
		align-items: center;
		gap: 16px;
	}

	:global(.dark) .stat-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.stat-icon {
		width: 56px;
		height: 56px;
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
	}

	.stat-card.requests .stat-icon { background: #dbeafe; color: #2563eb; }
	.stat-card.tokens .stat-icon { background: #f3e8ff; color: #9333ea; }
	.stat-card.input .stat-icon { background: #dcfce7; color: #16a34a; }
	.stat-card.output .stat-icon { background: #fef3c7; color: #d97706; }

	:global(.dark) .stat-card.requests .stat-icon { background: rgba(37, 99, 235, 0.2); }
	:global(.dark) .stat-card.tokens .stat-icon { background: rgba(147, 51, 234, 0.2); }
	:global(.dark) .stat-card.input .stat-icon { background: rgba(22, 163, 74, 0.2); }
	:global(.dark) .stat-card.output .stat-icon { background: rgba(217, 119, 6, 0.2); }

	.stat-content {
		display: flex;
		flex-direction: column;
	}

	.stat-value {
		font-size: 2rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		line-height: 1;
	}

	:global(.dark) .stat-value {
		color: #f9fafb;
	}

	.stat-label {
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
		margin-top: 6px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	:global(.dark) .stat-label {
		color: #6b7280;
	}

	/* Cost Analysis */
	.cost-analysis {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	@media (max-width: 640px) {
		.cost-analysis { grid-template-columns: 1fr; }
	}

	.cost-card {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 24px;
	}

	:global(.dark) .cost-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.cost-card.spent {
		border-left: 4px solid #6366f1;
	}

	.cost-card.saved {
		border-left: 4px solid #10b981;
	}

	.cost-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.cost-label {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .cost-label {
		color: #9ca3af;
	}

	.cost-badge {
		font-size: 0.625rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		padding: 4px 8px;
		border-radius: 4px;
		background: #e0e7ff;
		color: #4338ca;
	}

	.cost-badge.saved {
		background: #d1fae5;
		color: #047857;
	}

	:global(.dark) .cost-badge {
		background: rgba(99, 102, 241, 0.2);
		color: #a5b4fc;
	}

	:global(.dark) .cost-badge.saved {
		background: rgba(16, 185, 129, 0.2);
		color: #6ee7b7;
	}

	.cost-value {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		display: block;
	}

	:global(.dark) .cost-value {
		color: #f9fafb;
	}

	.cost-card.saved .cost-value {
		color: #10b981;
	}

	.cost-note {
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
		margin-top: 8px;
	}

	:global(.dark) .cost-note {
		color: #6b7280;
	}

	/* Two Column Grid */
	.two-column-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 24px;
	}

	@media (max-width: 768px) {
		.two-column-grid { grid-template-columns: 1fr; }
	}

	/* Usage Section */
	.usage-section {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 24px;
	}

	.usage-section.full-width {
		grid-column: 1 / -1;
	}

	:global(.dark) .usage-section {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.section-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text, #111827);
		margin: 0 0 20px;
	}

	:global(.dark) .section-title {
		color: #f9fafb;
	}

	/* Provider List */
	.provider-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.provider-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
	}

	.provider-info {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 140px;
	}

	.provider-icon {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #e0e7ff;
		color: #4338ca;
	}

	.provider-icon.local {
		background: #d1fae5;
		color: #047857;
	}

	.provider-icon.anthropic {
		background: #fed7aa;
		color: #c2410c;
	}

	.provider-icon.groq {
		background: #dbeafe;
		color: #1d4ed8;
	}

	.provider-icon.openai {
		background: #dcfce7;
		color: #16a34a;
	}

	:global(.dark) .provider-icon {
		background: rgba(99, 102, 241, 0.2);
	}

	:global(.dark) .provider-icon.local {
		background: rgba(16, 185, 129, 0.2);
	}

	:global(.dark) .provider-icon.anthropic {
		background: rgba(194, 65, 12, 0.2);
	}

	.provider-icon svg {
		width: 20px;
		height: 20px;
	}

	.provider-details {
		display: flex;
		flex-direction: column;
	}

	.provider-name {
		font-weight: 600;
		color: var(--color-text, #111827);
		text-transform: capitalize;
	}

	:global(.dark) .provider-name {
		color: #f9fafb;
	}

	.provider-type {
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
	}

	:global(.dark) .provider-type {
		color: #6b7280;
	}

	.provider-stats {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.provider-bar-container {
		height: 8px;
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 4px;
		overflow: hidden;
	}

	:global(.dark) .provider-bar-container {
		background: #1f1f1f;
	}

	.provider-bar {
		height: 100%;
		background: linear-gradient(90deg, #6366f1, #8b5cf6);
		border-radius: 4px;
		transition: width 0.3s ease;
	}

	.provider-bar.local {
		background: linear-gradient(90deg, #10b981, #34d399);
	}

	.provider-numbers {
		display: flex;
		justify-content: space-between;
	}

	.provider-tokens {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text, #111827);
	}

	:global(.dark) .provider-tokens {
		color: #f9fafb;
	}

	.provider-cost {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .provider-cost {
		color: #9ca3af;
	}

	/* Model List */
	.model-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.model-card {
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 12px;
		padding: 16px;
	}

	:global(.dark) .model-card {
		background: #141414;
	}

	.model-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 12px;
	}

	.model-name {
		font-weight: 600;
		color: var(--color-text, #111827);
		font-size: 0.875rem;
	}

	:global(.dark) .model-name {
		color: #f9fafb;
	}

	.model-provider {
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		padding: 2px 6px;
		border-radius: 4px;
		background: #e0e7ff;
		color: #4338ca;
	}

	.model-provider.local {
		background: #d1fae5;
		color: #047857;
	}

	:global(.dark) .model-provider {
		background: rgba(99, 102, 241, 0.2);
		color: #a5b4fc;
	}

	:global(.dark) .model-provider.local {
		background: rgba(16, 185, 129, 0.2);
		color: #6ee7b7;
	}

	.model-stats {
		display: flex;
		gap: 24px;
		margin-bottom: 8px;
	}

	.model-stat {
		display: flex;
		flex-direction: column;
	}

	.model-stat-value {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--color-text, #111827);
	}

	:global(.dark) .model-stat-value {
		color: #f9fafb;
	}

	.model-stat-label {
		font-size: 0.625rem;
		color: var(--color-text-muted, #9ca3af);
		text-transform: uppercase;
	}

	:global(.dark) .model-stat-label {
		color: #6b7280;
	}

	.model-cost {
		font-size: 0.75rem;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .model-cost {
		color: #9ca3af;
	}

	/* Agent Grid */
	.agent-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}

	.agent-card {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 12px;
	}

	:global(.dark) .agent-card {
		background: #141414;
	}

	.agent-icon {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		background: #dbeafe;
		color: #2563eb;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	:global(.dark) .agent-icon {
		background: rgba(37, 99, 235, 0.2);
	}

	.agent-icon svg {
		width: 20px;
		height: 20px;
	}

	.agent-info {
		flex: 1;
		min-width: 0;
	}

	.agent-name {
		font-weight: 600;
		color: var(--color-text, #111827);
		font-size: 0.875rem;
		text-transform: capitalize;
	}

	:global(.dark) .agent-name {
		color: #f9fafb;
	}

	.agent-stats-row {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.75rem;
		color: var(--color-text-secondary, #6b7280);
		margin-top: 4px;
	}

	:global(.dark) .agent-stats-row {
		color: #9ca3af;
	}

	.dot {
		width: 3px;
		height: 3px;
		border-radius: 50%;
		background: currentColor;
		opacity: 0.5;
	}

	/* MCP Grid */
	.mcp-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 12px;
	}

	.mcp-card {
		padding: 16px;
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 12px;
	}

	:global(.dark) .mcp-card {
		background: #141414;
	}

	.mcp-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 12px;
		gap: 8px;
	}

	.mcp-name {
		font-weight: 600;
		color: var(--color-text, #111827);
		font-size: 0.875rem;
	}

	:global(.dark) .mcp-name {
		color: #f9fafb;
	}

	.mcp-server {
		font-size: 0.625rem;
		color: var(--color-text-muted, #9ca3af);
		background: var(--color-border, #e5e7eb);
		padding: 2px 6px;
		border-radius: 4px;
		white-space: nowrap;
	}

	:global(.dark) .mcp-server {
		background: #2c2c2e;
		color: #6b7280;
	}

	.mcp-stats {
		display: flex;
		gap: 12px;
	}

	.mcp-stat {
		display: flex;
		flex-direction: column;
	}

	.mcp-stat-value {
		font-size: 1rem;
		font-weight: 700;
		color: var(--color-text, #111827);
	}

	:global(.dark) .mcp-stat-value {
		color: #f9fafb;
	}

	.mcp-stat-value.success {
		color: #10b981;
	}

	.mcp-stat-label {
		font-size: 0.625rem;
		color: var(--color-text-muted, #9ca3af);
		text-transform: uppercase;
	}

	:global(.dark) .mcp-stat-label {
		color: #6b7280;
	}

	/* Trend Chart */
	.trend-chart {
		display: flex;
		align-items: flex-end;
		height: 150px;
		gap: 8px;
		padding: 20px 0;
	}

	.trend-bar-container {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		height: 100%;
		cursor: pointer;
	}

	.trend-bar {
		width: 100%;
		max-width: 24px;
		background: linear-gradient(180deg, #6366f1, #8b5cf6);
		border-radius: 4px 4px 0 0;
		transition: height 0.3s ease;
		margin-top: auto;
	}

	.trend-bar-container:hover .trend-bar {
		background: linear-gradient(180deg, #4f46e5, #7c3aed);
	}

	.trend-label {
		font-size: 0.625rem;
		color: var(--color-text-muted, #9ca3af);
		margin-top: 8px;
	}

	:global(.dark) .trend-label {
		color: #6b7280;
	}

	.trend-legend {
		display: flex;
		justify-content: center;
		gap: 24px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border, #e5e7eb);
	}

	:global(.dark) .trend-legend {
		border-color: rgba(255, 255, 255, 0.08);
	}

	.trend-legend-item {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.75rem;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .trend-legend-item {
		color: #9ca3af;
	}

	.trend-legend-dot {
		width: 12px;
		height: 12px;
		border-radius: 3px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
	}
</style>
