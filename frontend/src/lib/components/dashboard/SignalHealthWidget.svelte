<script lang="ts">
	import { signalStore, genreDistribution, modeDistribution } from '$lib/stores/signal';
	import {
		getSignalHealth,
		type SignalHealthResponse,
		GENRE_LABELS,
		MODE_LABELS
	} from '$lib/api/signal';
	import { onMount } from 'svelte';

	let health = $state<SignalHealthResponse | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const genres = $derived($genreDistribution);
	const modes = $derived($modeDistribution);
	const totalEvents = $derived($signalStore.history.length);

	onMount(async () => {
		try {
			health = await getSignalHealth();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load signal health';
		} finally {
			loading = false;
		}
	});

	const statusClass = $derived(() => {
		if (!health) return 'dw-sh-status--unknown';
		if (health.status === 'healthy') return 'dw-sh-status--healthy';
		if (health.status === 'degraded') return 'dw-sh-status--degraded';
		return 'dw-sh-status--unknown';
	});

	/** Count how many of the 6 core metrics are passing */
	const passingMetrics = $derived(() => {
		if (!health) return null;
		const m = health.metrics;
		return [
			m.action_completion,
			m.re_encoding,
			m.signal_bounce,
			m.genre_recognition,
			m.feedback_closure,
			m.time_to_decide
		].filter(Boolean).length;
	});

	// Genre hex colors for inline style (dark-mode safe)
	const genreHex: Record<string, string> = {
		DIRECT: '#fb923c',
		INFORM: '#60a5fa',
		COMMIT: '#34d399',
		DECIDE: '#fb7185',
		EXPRESS: '#f472b6'
	};

	// Mode hex colors for inline style (dark-mode safe)
	const modeHex: Record<string, string> = {
		BUILD: '#6366f1',
		ASSIST: '#22c55e',
		ANALYZE: '#8b5cf6',
		EXECUTE: '#f59e0b',
		MAINTAIN: '#94a3b8'
	};
</script>

<div class="dw-sh-widget">
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="dw-sh-icon w-8 h-8 rounded-lg flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
			</div>
			<h2 class="dw-sh-title text-base font-semibold">Signal Health</h2>
		</div>

		{#if loading}
			<span class="dw-sh-meta text-xs">Loading...</span>
		{:else if health}
			<span class="inline-flex items-center gap-1.5 text-xs font-medium {statusClass()}">
				<span class="dw-sh-dot h-2 w-2 rounded-full"></span>
				{health.status}
			</span>
		{/if}
	</div>

	{#if error}
		<p class="text-xs text-red-500 py-2">{error}</p>
	{:else if loading}
		<!-- Skeleton -->
		<div class="space-y-3 animate-pulse">
			<div class="grid grid-cols-3 gap-3">
				{#each [0, 1, 2] as _}
					<div class="dw-sh-skeleton h-12 rounded-lg"></div>
				{/each}
			</div>
			<div class="dw-sh-skeleton h-4 rounded w-3/4"></div>
			<div class="dw-sh-skeleton h-3 rounded"></div>
			<div class="dw-sh-skeleton h-3 rounded w-5/6"></div>
		</div>
	{:else}
		<!-- System metrics summary -->
		{#if health}
			<div class="grid grid-cols-3 gap-3 mb-4">
				<div class="dw-sh-card text-center rounded-lg py-2 px-1">
					<div class="dw-sh-value text-lg font-bold">{passingMetrics()}/6</div>
					<div class="dw-sh-meta text-[10px] uppercase tracking-wider mt-0.5">Metrics</div>
				</div>
				<div class="dw-sh-card text-center rounded-lg py-2 px-1">
					<div class="text-lg font-bold" class:dw-sh-on={health.feedback_loop.homeostatic_loop} class:dw-sh-meta={!health.feedback_loop.homeostatic_loop}>
						{health.feedback_loop.homeostatic_loop ? 'ON' : 'OFF'}
					</div>
					<div class="dw-sh-meta text-[10px] uppercase tracking-wider mt-0.5">Feedback</div>
				</div>
				<div class="dw-sh-card text-center rounded-lg py-2 px-1">
					<div class="dw-sh-value text-lg font-bold">{totalEvents}</div>
					<div class="dw-sh-meta text-[10px] uppercase tracking-wider mt-0.5">Signals</div>
				</div>
			</div>
		{/if}

		<!-- Genre distribution bars -->
		{#if genres.length > 0}
			<div class="mb-4">
				<div class="dw-sh-meta text-[10px] uppercase tracking-wider mb-2">Genre Distribution</div>
				<div class="space-y-1.5">
					{#each genres as g}
						<div class="flex items-center gap-2">
							<span class="dw-sh-label text-[10px] w-16 text-right shrink-0">
								{GENRE_LABELS[g.genre] ?? g.genre}
							</span>
							<div class="dw-sh-bar-track flex-1 h-2 rounded-full overflow-hidden">
								<div
									class="h-full rounded-full transition-all duration-500"
									style="width: {g.percentage}%; background-color: {genreHex[g.genre] ?? '#999'}"
									role="meter"
									aria-label="{GENRE_LABELS[g.genre] ?? g.genre}: {g.percentage}%"
									aria-valuenow={g.percentage}
									aria-valuemin={0}
									aria-valuemax={100}
								></div>
							</div>
							<span class="dw-sh-meta text-[10px] w-7 text-right shrink-0">{g.percentage}%</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Mode distribution stacked bar -->
		{#if modes.length > 0}
			<div>
				<div class="dw-sh-meta text-[10px] uppercase tracking-wider mb-2">Mode Distribution</div>
				<div class="dw-sh-bar-track flex gap-px h-3 rounded-full overflow-hidden" role="group" aria-label="Mode distribution">
					{#each modes as m}
						<div
							class="h-full transition-all duration-500"
							style="width: {m.percentage}%; background-color: {modeHex[m.mode] ?? '#999'}"
							title="{MODE_LABELS[m.mode] ?? m.mode}: {m.percentage}%"
						></div>
					{/each}
				</div>
				<div class="flex flex-wrap gap-x-3 gap-y-1 mt-2">
					{#each modes as m}
						<div class="flex items-center gap-1">
							<span class="h-1.5 w-1.5 rounded-full shrink-0" style="background-color: {modeHex[m.mode] ?? '#999'}"></span>
							<span class="dw-sh-meta text-[9px]">{MODE_LABELS[m.mode] ?? m.mode}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Empty state -->
		{#if genres.length === 0 && modes.length === 0}
			<p class="dw-sh-meta text-xs text-center py-6">
				No signal data yet. Start a conversation to see classification.
			</p>
		{/if}
	{/if}
</div>

<style>
	/* Signal Health Widget — Foundation tokens with BOS fallbacks */
	.dw-sh-widget {
		background: var(--dbg, var(--bos-card, #fff));
		border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
		border-radius: 0.75rem;
		padding: 1.25rem;
		box-shadow: var(--bos-shadow-1, 0 1px 2px rgba(0,0,0,.05));
		transition: box-shadow 0.3s;
	}
	.dw-sh-widget:hover {
		box-shadow: var(--bos-shadow-2, 0 4px 6px rgba(0,0,0,.07));
	}

	.dw-sh-icon {
		background: linear-gradient(135deg, #7c3aed, #4338ca);
	}

	.dw-sh-title {
		color: var(--dt, var(--bos-text-primary, #111));
	}

	.dw-sh-value {
		color: var(--dt, var(--bos-text-primary, #111));
	}

	.dw-sh-meta {
		color: var(--dt3, var(--bos-text-tertiary, #888));
	}

	.dw-sh-label {
		color: var(--dt2, var(--bos-text-secondary, #555));
	}

	.dw-sh-card {
		background: var(--dbg2, var(--bos-hover, #f5f5f5));
	}

	.dw-sh-skeleton {
		background: var(--dbg3, var(--bos-hover, #eee));
	}

	.dw-sh-bar-track {
		background: var(--dbg3, var(--bos-hover, #eee));
	}

	/* Status variants */
	.dw-sh-status--healthy {
		color: #16a34a;
	}
	.dw-sh-status--healthy .dw-sh-dot {
		background-color: #22c55e;
	}
	.dw-sh-status--degraded {
		color: #d97706;
	}
	.dw-sh-status--degraded .dw-sh-dot {
		background-color: #f59e0b;
	}
	.dw-sh-status--unknown {
		color: var(--dt3, #888);
	}
	.dw-sh-status--unknown .dw-sh-dot {
		background-color: var(--dt4, #bbb);
	}

	.dw-sh-on {
		color: #16a34a;
	}
</style>
