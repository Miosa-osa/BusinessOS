<script lang="ts">
	import type { ProviderUsage, ModelUsage } from '$lib/api';

	interface Props {
		usageByProvider: ProviderUsage[];
		usageByModel: ModelUsage[];
		formatNumber: (num: number) => string;
		formatCost: (cost: number) => string;
	}

	let { usageByProvider, usageByModel, formatNumber, formatCost }: Props = $props();
</script>

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

<style>
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
</style>
