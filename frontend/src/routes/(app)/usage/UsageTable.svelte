<script lang="ts">
	import type { AgentUsage, MCPToolUsage, UsageTrendPoint } from '$lib/api';

	interface Props {
		usageByAgent: AgentUsage[];
		mcpUsage: MCPToolUsage[];
		usageTrend: UsageTrendPoint[];
		formatNumber: (num: number) => string;
		formatDuration: (ms: number) => string;
	}

	let {
		usageByAgent,
		mcpUsage,
		usageTrend,
		formatNumber,
		formatDuration
	}: Props = $props();
</script>

<!-- Agent Usage -->
{#if usageByAgent.length > 0}
	<div class="usage-section">
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
	<div class="usage-section">
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
	<div class="usage-section">
		<h3 class="section-title">Usage Trend (Last 30 Days)</h3>
		<div class="trend-chart">
			{#each usageTrend.slice(-14) as point}
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

<style>
	/* Usage Section */
	.usage-section {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 24px;
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
