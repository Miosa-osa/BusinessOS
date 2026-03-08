<script lang="ts">
	import type { CustomModule } from '$lib/types/modules';
	import { categoryLabels } from '$lib/types/modules';
	import { Download, Star, Package } from 'lucide-svelte';

	interface Props {
		module: CustomModule;
		onClick?: () => void;
	}

	let { module, onClick }: Props = $props();

	const categoryHexColors: Record<string, string> = {
		productivity: '#3b82f6',
		communication: '#a855f7',
		finance: '#10b981',
		analytics: '#f97316',
		automation: '#ec4899',
		integration: '#6366f1',
		utilities: '#6b7280',
		custom: '#06b6d4',
	};

	function fmtNum(n: number): string {
		if (n >= 1000) return (n / 1000).toFixed(n >= 10000 ? 0 : 1) + 'K';
		return String(n);
	}
</script>

<button
	onclick={onClick}
	class="am-module-card"
	aria-label="View {module.name}"
>
	<!-- Header with Icon and Meta -->
	<div class="am-module-card__header">
		{#if module.icon}
			<div
				class="am-module-card__icon"
				style="background: {categoryHexColors[module.category] || '#6366f1'}"
			>
				{module.icon}
			</div>
		{:else}
			<div class="am-module-card__icon am-module-card__icon--fallback">
				<Package class="w-5 h-5" />
			</div>
		{/if}
		<div class="am-module-card__meta">
			<div class="am-module-card__name">{module.name}</div>
			{#if module.creator_name}
				<div class="am-module-card__author">by {module.creator_name}</div>
			{/if}
		</div>
		<span
			class="am-visibility-badge am-visibility-badge--{module.visibility}"
		>{module.visibility}</span>
	</div>

	<!-- Description -->
	<p class="am-module-card__desc">{module.description}</p>

	<!-- Footer: Category badge + stats -->
	<div class="am-module-card__footer">
		<span
			class="am-cat-badge"
			style="background: {categoryHexColors[module.category] || '#6366f1'}16; color: {categoryHexColors[module.category] || '#6366f1'}"
		>{categoryLabels[module.category]}</span>
		<span class="am-stat" title="Installs">
			<Download class="w-3 h-3" />
			{fmtNum(module.install_count)}
		</span>
		<span class="am-stat" title="Stars">
			<Star class="w-3 h-3" />
			{fmtNum(module.star_count)}
		</span>
		<span class="am-card-version">v{module.version}</span>
	</div>
</button>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  MODULE CARD (am-) — Foundation AppMarketplace Pattern       */
	/* ══════════════════════════════════════════════════════════════ */
	.am-module-card {
		background: rgba(255, 255, 255, 0.04);
		backdrop-filter: blur(20px);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 16px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		cursor: pointer;
		text-align: left;
		width: 100%;
		transition: transform .15s, box-shadow .15s, border-color .15s;
	}
	.am-module-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 28px rgba(0, 0, 0, 0.1);
		border-color: var(--dbd2, #f0f0f0);
	}

	/* Header */
	.am-module-card__header {
		display: flex;
		align-items: flex-start;
		gap: 10px;
	}
	.am-module-card__icon {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		flex-shrink: 0;
	}
	.am-module-card__icon--fallback {
		background: var(--dbg3, #eee);
		color: var(--dt3, #888);
	}
	.am-module-card__meta {
		flex: 1;
		min-width: 0;
	}
	.am-module-card__name {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.am-module-card:hover .am-module-card__name {
		color: var(--accent-blue, #3b82f6);
	}
	.am-module-card__author {
		font-size: 11px;
		color: var(--dt3, #888);
	}

	/* Description */
	.am-module-card__desc {
		font-size: 12px;
		color: var(--dt2, #555);
		line-height: 1.5;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		flex: 1;
	}

	/* Footer */
	.am-module-card__footer {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: auto;
	}

	/* Shared atoms */
	.am-cat-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 600;
	}
	.am-visibility-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		flex-shrink: 0;
		margin-left: auto;
	}
	.am-visibility-badge--public {
		background: #10b98122;
		color: #10b981;
	}
	.am-visibility-badge--workspace {
		background: #6366f122;
		color: #6366f1;
	}
	.am-visibility-badge--private {
		background: #6b728022;
		color: #6b7280;
	}
	.am-stat {
		display: inline-flex;
		align-items: center;
		gap: 3px;
		font-size: 11px;
		color: var(--dt3, #888);
	}
	.am-card-version {
		font-size: 11px;
		color: var(--dt4, #bbb);
		margin-left: auto;
	}
</style>
