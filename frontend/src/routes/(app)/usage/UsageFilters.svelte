<script lang="ts">
	interface Props {
		usagePeriod: 'today' | 'week' | 'month' | 'all';
		isLoading: boolean;
		onPeriodChange: (period: 'today' | 'week' | 'month' | 'all') => void;
	}

	let { usagePeriod, isLoading, onPeriodChange }: Props = $props();
</script>

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
	<div class="period-selector">
		{#each ['today', 'week', 'month', 'all'] as period}
			<button
				onclick={() => onPeriodChange(period as typeof usagePeriod)}
				class="period-btn"
				class:active={usagePeriod === period}
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
{/if}

<style>
	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		gap: 20px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--color-border);
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
		background: linear-gradient(135deg, var(--bos-category-productivity), var(--bos-category-ai));
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

	/* Period Selector */
	.period-selector {
		display: flex;
		background: var(--color-bg-secondary);
		border-radius: 12px;
		padding: 4px;
		gap: 4px;
	}

	:global(.dark) .period-selector {
		background: var(--dbg2);
	}

	.period-btn {
		padding: 10px 20px;
		border: none;
		background: transparent;
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text-secondary, #6b7280);
		cursor: pointer;
		transition: all 0.15s;
	}

	:global(.dark) .period-btn {
		color: #9ca3af;
	}

	.period-btn:hover {
		color: var(--color-text, #111827);
	}

	:global(.dark) .period-btn:hover {
		color: #f9fafb;
	}

	.period-btn.active {
		background: var(--color-bg, white);
		color: var(--color-text, #111827);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	:global(.dark) .period-btn.active {
		background: var(--dbg, #2c2c2e);
		color: #f9fafb;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
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
		border: 3px solid var(--color-border);
		border-top-color: var(--bos-category-productivity);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	:global(.dark) .loading-spinner {
		border-color: rgba(255, 255, 255, 0.1);
		border-top-color: var(--bos-category-productivity);
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
