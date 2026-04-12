<script lang="ts">
	import type { UsageSummary, ProviderUsage } from '$lib/api';

	interface Props {
		usageSummary: UsageSummary;
		usageByProvider: ProviderUsage[];
		formatNumber: (num: number) => string;
		formatCost: (cost: number) => string;
	}

	let { usageSummary, usageByProvider, formatNumber, formatCost }: Props = $props();
</script>

<!-- Stats Grid -->
<div class="stats-grid">
	<div class="stat-card requests">
		<div class="stat-icon">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
			</svg>
		</div>
		<div class="stat-content">
			<span class="stat-value">{formatNumber(usageSummary.total_requests || 0)}</span>
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
			<span class="stat-value">{formatNumber(usageSummary.total_tokens || 0)}</span>
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
			<span class="stat-value">{formatNumber(usageSummary.total_input_tokens || 0)}</span>
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
			<span class="stat-value">{formatNumber(usageSummary.total_output_tokens || 0)}</span>
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
		<span class="cost-value">{formatCost(usageSummary.total_cost || 0)}</span>
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

<style>
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
		border: 1px solid var(--color-border);
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

	.stat-card.requests .stat-icon { background: var(--bos-status-info-bg); color: var(--bos-status-info); }
	.stat-card.tokens .stat-icon { background: color-mix(in srgb, var(--bos-category-ai) 15%, transparent); color: var(--bos-category-ai); }
	.stat-card.input .stat-icon { background: var(--bos-status-success-bg); color: var(--bos-status-success); }
	.stat-card.output .stat-icon { background: var(--bos-status-warning-bg); color: var(--bos-status-warning); }

	:global(.dark) .stat-card.requests .stat-icon { background: color-mix(in srgb, var(--bos-status-info) 20%, transparent); }
	:global(.dark) .stat-card.tokens .stat-icon { background: color-mix(in srgb, var(--bos-category-ai) 20%, transparent); }
	:global(.dark) .stat-card.input .stat-icon { background: color-mix(in srgb, var(--bos-status-success) 20%, transparent); }
	:global(.dark) .stat-card.output .stat-icon { background: color-mix(in srgb, var(--bos-status-warning) 20%, transparent); }

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
		border: 1px solid var(--color-border);
		border-radius: 16px;
		padding: 24px;
	}

	:global(.dark) .cost-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.cost-card.spent {
		border-left: 4px solid var(--bos-category-productivity);
	}

	.cost-card.saved {
		border-left: 4px solid var(--bos-status-success);
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
		background: color-mix(in srgb, var(--bos-category-productivity) 15%, transparent);
		color: var(--bos-category-productivity);
	}

	.cost-badge.saved {
		background: var(--bos-status-success-bg);
		color: var(--bos-status-success);
	}

	:global(.dark) .cost-badge {
		background: color-mix(in srgb, var(--bos-category-productivity) 20%, transparent);
		color: var(--bos-category-productivity);
	}

	:global(.dark) .cost-badge.saved {
		background: color-mix(in srgb, var(--bos-status-success) 20%, transparent);
		color: var(--bos-status-success);
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
		color: var(--bos-status-success);
	}

	.cost-note {
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
		margin-top: 8px;
	}

	:global(.dark) .cost-note {
		color: #6b7280;
	}
</style>
