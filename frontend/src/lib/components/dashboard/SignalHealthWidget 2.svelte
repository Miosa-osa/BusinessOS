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

	// Derived from legacy Svelte stores via $ subscription
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

	const statusDotClass = $derived(() => {
		if (!health) return 'bg-gray-300';
		if (health.status === 'healthy') return 'bg-green-500';
		if (health.status === 'degraded') return 'bg-amber-500';
		return 'bg-gray-400';
	});

	const statusTextClass = $derived(() => {
		if (!health) return 'text-gray-400';
		if (health.status === 'healthy') return 'text-green-600';
		if (health.status === 'degraded') return 'text-amber-600';
		return 'text-gray-400';
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

	// Tailwind color classes for genre bars
	const genreColors: Record<string, string> = {
		DIRECT: 'bg-orange-400',
		INFORM: 'bg-blue-400',
		COMMIT: 'bg-emerald-400',
		DECIDE: 'bg-rose-400',
		EXPRESS: 'bg-pink-400'
	};

	// Tailwind color classes for mode segments
	const modeColors: Record<string, string> = {
		BUILD: 'bg-indigo-500',
		ASSIST: 'bg-green-500',
		ANALYZE: 'bg-violet-500',
		EXECUTE: 'bg-amber-500',
		MAINTAIN: 'bg-slate-400'
	};
</script>

<div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm hover:shadow-md transition-shadow duration-300">
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-violet-600 to-indigo-700 flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900">Signal Health</h2>
		</div>

		{#if loading}
			<span class="text-xs text-gray-400">Loading...</span>
		{:else if health}
			<span class="inline-flex items-center gap-1.5 text-xs font-medium {statusTextClass()}">
				<span class="h-2 w-2 rounded-full {statusDotClass()}"></span>
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
					<div class="h-12 bg-gray-100 rounded-lg"></div>
				{/each}
			</div>
			<div class="h-4 bg-gray-100 rounded w-3/4"></div>
			<div class="h-3 bg-gray-100 rounded"></div>
			<div class="h-3 bg-gray-100 rounded w-5/6"></div>
		</div>
	{:else}
		<!-- System metrics summary -->
		{#if health}
			<div class="grid grid-cols-3 gap-3 mb-4">
				<div class="text-center bg-gray-50 rounded-lg py-2 px-1">
					<div class="text-lg font-bold text-gray-900">{passingMetrics()}/6</div>
					<div class="text-[10px] text-gray-400 uppercase tracking-wider mt-0.5">Metrics</div>
				</div>
				<div class="text-center bg-gray-50 rounded-lg py-2 px-1">
					<div class="text-lg font-bold {health.feedback_loop.homeostatic_loop ? 'text-green-600' : 'text-gray-400'}">
						{health.feedback_loop.homeostatic_loop ? 'ON' : 'OFF'}
					</div>
					<div class="text-[10px] text-gray-400 uppercase tracking-wider mt-0.5">Feedback</div>
				</div>
				<div class="text-center bg-gray-50 rounded-lg py-2 px-1">
					<div class="text-lg font-bold text-gray-900">{totalEvents}</div>
					<div class="text-[10px] text-gray-400 uppercase tracking-wider mt-0.5">Signals</div>
				</div>
			</div>
		{/if}

		<!-- Genre distribution bars -->
		{#if genres.length > 0}
			<div class="mb-4">
				<div class="text-[10px] text-gray-400 uppercase tracking-wider mb-2">Genre Distribution</div>
				<div class="space-y-1.5">
					{#each genres as g}
						<div class="flex items-center gap-2">
							<span class="text-[10px] text-gray-500 w-16 text-right shrink-0">
								{GENRE_LABELS[g.genre] ?? g.genre}
							</span>
							<div class="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
								<div
									class="h-full rounded-full transition-all duration-500 {genreColors[g.genre] ?? 'bg-gray-300'}"
									style="width: {g.percentage}%"
									role="meter"
									aria-label="{GENRE_LABELS[g.genre] ?? g.genre}: {g.percentage}%"
									aria-valuenow={g.percentage}
									aria-valuemin={0}
									aria-valuemax={100}
								></div>
							</div>
							<span class="text-[10px] text-gray-400 w-7 text-right shrink-0">{g.percentage}%</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Mode distribution stacked bar -->
		{#if modes.length > 0}
			<div>
				<div class="text-[10px] text-gray-400 uppercase tracking-wider mb-2">Mode Distribution</div>
				<div class="flex gap-px h-3 rounded-full overflow-hidden bg-gray-100" role="group" aria-label="Mode distribution">
					{#each modes as m}
						<div
							class="h-full transition-all duration-500 {modeColors[m.mode] ?? 'bg-gray-300'}"
							style="width: {m.percentage}%"
							title="{MODE_LABELS[m.mode] ?? m.mode}: {m.percentage}%"
						></div>
					{/each}
				</div>
				<div class="flex flex-wrap gap-x-3 gap-y-1 mt-2">
					{#each modes as m}
						<div class="flex items-center gap-1">
							<span class="h-1.5 w-1.5 rounded-full {modeColors[m.mode] ?? 'bg-gray-300'} shrink-0"></span>
							<span class="text-[9px] text-gray-400">{MODE_LABELS[m.mode] ?? m.mode}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Empty state -->
		{#if genres.length === 0 && modes.length === 0}
			<p class="text-xs text-gray-400 text-center py-6">
				No signal data yet. Start a conversation to see classification.
			</p>
		{/if}
	{/if}
</div>
