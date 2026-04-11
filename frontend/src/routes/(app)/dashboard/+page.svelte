<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fade, fly } from 'svelte/transition';
	import { optimalStore, type OptimalNode, type DashboardData } from '$lib/stores/optimal';

	// ── State ────────────────────────────────────────────────────────────────────

	let storeState = $state({
		nodes: [] as OptimalNode[],
		todayRhythm: null as { date: string; content: string } | null,
		dashboard: null as DashboardData | null,
		loading: false,
		error: null as string | null,
	});

	$effect(() => {
		const unsub = optimalStore.subscribe((s) => {
			storeState = { nodes: s.nodes, todayRhythm: s.todayRhythm, dashboard: s.dashboard, loading: s.loading, error: s.error };
		});
		return unsub;
	});

	// ── Derived ──────────────────────────────────────────────────────────────────

	let totalSignals = $derived(storeState.nodes.reduce((sum, n) => sum + n.signal_count, 0));
	let activeNodes = $derived(storeState.nodes.filter((n) => n.has_signals).length);

	let rhythmMode = $derived(parseMode(storeState.todayRhythm?.content ?? ''));
	let rhythmPriorities = $derived(parsePriorities(storeState.todayRhythm?.content ?? ''));
	let rhythmDate = $derived(storeState.todayRhythm?.date ?? '');

	// ── Helpers ──────────────────────────────────────────────────────────────────

	function parseMode(content: string): string {
		const match = content.match(/mode[:\s]+([A-Z]+)/i) ?? content.match(/#+\s*(BUILD|OPERATE|LEARN|SYNTHESIZE|EXTRACT)/i);
		return match?.[1]?.toUpperCase() ?? '';
	}

	function parsePriorities(content: string): string[] {
		const lines = content.split('\n');
		const results: string[] = [];

		// Look for numbered list items or bullet points near "priority" or "non-negotiable"
		let inPrioritySection = false;
		for (const line of lines) {
			const lower = line.toLowerCase();
			if (lower.includes('non-negotiable') || lower.includes('priority') || lower.includes('top 3')) {
				inPrioritySection = true;
				continue;
			}
			if (inPrioritySection) {
				const bullet = line.match(/^\s*[-*\d.]+\s+(.+)/);
				if (bullet) {
					results.push(bullet[1].trim());
					if (results.length >= 3) break;
				} else if (line.trim() && line.startsWith('#')) {
					break; // hit next section
				}
			}
		}

		// Fallback: first 3 non-empty, non-heading lines
		if (results.length === 0) {
			for (const line of lines) {
				const t = line.trim();
				if (t && !t.startsWith('#') && !t.startsWith('---') && t.length > 10) {
					results.push(t.replace(/^[-*\d.]+\s*/, ''));
					if (results.length >= 3) break;
				}
			}
		}

		return results.slice(0, 3);
	}

	function extractSignalPreview(signal_md: string): string {
		if (!signal_md) return '';
		const statusLine = signal_md.split('\n').find((l) => /status/i.test(l));
		if (statusLine) return statusLine.replace(/^#+\s*/, '').trim();
		// first meaningful line
		const first = signal_md.split('\n').find((l) => l.trim() && !l.startsWith('#') && !l.startsWith('---'));
		return first?.trim().slice(0, 120) ?? '';
	}

	function nodeTypeColor(type: string): string {
		const map: Record<string, string> = {
			person: 'od-badge--person',
			entity: 'od-badge--entity',
			'operation:program': 'od-badge--operation',
			'operation:research': 'od-badge--research',
			'operation:network': 'od-badge--network',
			'unit:community': 'od-badge--community',
			'domain:media': 'od-badge--media',
			inbox: 'od-badge--inbox',
			registry: 'od-badge--registry',
			'cross-cutting': 'od-badge--cross',
			'layer:command': 'od-badge--command',
			'layer:operating': 'od-badge--operating',
		};
		return map[type] ?? 'od-badge--default';
	}

	function shortType(type: string): string {
		return type.split(':').pop() ?? type;
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}

	const MODE_COLORS: Record<string, string> = {
		BUILD: '#3b82f6',
		OPERATE: '#f59e0b',
		LEARN: '#8b5cf6',
		SYNTHESIZE: '#10b981',
		EXTRACT: '#f97316',
	};

	// ── Quick Access ─────────────────────────────────────────────────────────────

	const quickAccessModules = [
		{ label: 'Chat', href: '/chat', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z', color: '#3b82f6' },
		{ label: 'Tasks', href: '/tasks', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4', color: '#22c55e' },
		{ label: 'Projects', href: '/projects', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', color: '#a855f7' },
		{ label: 'Team', href: '/team', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z', color: '#f59e0b' },
		{ label: 'Daily Log', href: '/daily', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z', color: '#ef4444' },
		{ label: 'Clients', href: '/clients', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', color: '#06b6d4' },
		{ label: 'Pages', href: '/pages', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', color: '#8b5cf6' },
		{ label: 'Terminal', href: '/terminal', icon: 'M4 17l6-6-6-6m8 14h8', color: '#10b981' },
	];

	// ── Lifecycle ────────────────────────────────────────────────────────────────

	onMount(() => {
		Promise.all([
			optimalStore.loadNodes(),
			optimalStore.loadTodayRhythm(),
			optimalStore.loadDashboard(),
		]);
	});
</script>

<div class="od-page">

	{#if storeState.loading && storeState.nodes.length === 0}
		<div class="od-center" in:fade>
			<div class="od-spinner" aria-label="Loading"></div>
			<p class="od-muted">Loading OptimalOS...</p>
		</div>

	{:else if storeState.error}
		<div class="od-center" in:fade>
			<p class="od-error-text">{storeState.error}</p>
			<button class="od-retry-btn" onclick={() => optimalStore.loadNodes()}>Retry</button>
		</div>

	{:else}
		<div class="od-content" in:fade={{ duration: 250 }}>

			<!-- ── Stats bar ─────────────────────────────────────────────────── -->
			<div class="od-stats-bar" in:fly={{ y: -8, duration: 250 }}>
				<div class="od-stat">
					<span class="od-stat-value">{storeState.nodes.length}</span>
					<span class="od-stat-label">nodes</span>
				</div>
				<div class="od-stat-sep"></div>
				<div class="od-stat">
					<span class="od-stat-value">{totalSignals}</span>
					<span class="od-stat-label">signals</span>
				</div>
				<div class="od-stat-sep"></div>
				<div class="od-stat">
					<span class="od-stat-value od-stat-value--active">{activeNodes}</span>
					<span class="od-stat-label">active</span>
				</div>
				<div class="od-stat-sep"></div>
				<div class="od-stat">
					<span class="od-stat-value od-stat-value--dim">{storeState.nodes.length - activeNodes}</span>
					<span class="od-stat-label">quiet</span>
				</div>
			</div>

			<!-- ── Command Center ───────────────────────────────────────────── -->
			{#if storeState.dashboard}
				<section aria-label="Command Center" in:fly={{ y: 12, duration: 300, delay: 40 }}>
					<div class="od-section-header">
						<span class="od-section-label">Command Center</span>
					</div>
					<div class="od-cc-grid">
						<!-- Total Nodes -->
						<div class="od-cc-card">
							<div class="od-cc-icon od-cc-icon--nodes" aria-hidden="true">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" width="16" height="16">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
								</svg>
							</div>
							<div class="od-cc-body">
								<span class="od-cc-value">{storeState.dashboard.nodes?.length ?? storeState.nodes.length}</span>
								<span class="od-cc-label">Total Nodes</span>
							</div>
						</div>

						<!-- Total Signals -->
						<div class="od-cc-card">
							<div class="od-cc-icon od-cc-icon--signals" aria-hidden="true">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" width="16" height="16">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M13 10V3L4 14h7v7l9-11h-7z" />
								</svg>
							</div>
							<div class="od-cc-body">
								<span class="od-cc-value">{storeState.dashboard.stats?.signal_count ?? totalSignals}</span>
								<span class="od-cc-label">Total Signals</span>
							</div>
						</div>

						<!-- Revenue / MRR -->
						<div class="od-cc-card">
							<div class="od-cc-icon od-cc-icon--revenue" aria-hidden="true">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" width="16" height="16">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
							</div>
							<div class="od-cc-body">
								<span class="od-cc-value">${(storeState.dashboard.revenue?.total_mrr ?? 0).toLocaleString()}</span>
								<span class="od-cc-label">MRR</span>
							</div>
						</div>

						<!-- Entity Count -->
						<div class="od-cc-card">
							<div class="od-cc-icon od-cc-icon--entities" aria-hidden="true">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" width="16" height="16">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
							</div>
							<div class="od-cc-body">
								<span class="od-cc-value">{storeState.dashboard.stats?.entity_count ?? 0}</span>
								<span class="od-cc-label">Entities</span>
							</div>
						</div>
					</div>

					<!-- Non-negotiables from command center -->
					{#if storeState.dashboard.command_center?.non_negotiables?.length > 0}
						<div class="od-cc-nonneg">
							<span class="od-cc-nonneg-label">Non-negotiables</span>
							<ul class="od-cc-nonneg-list">
								{#each storeState.dashboard.command_center.non_negotiables as item, i}
									<li class="od-cc-nonneg-item">
										<span class="od-cc-nonneg-num">{i + 1}</span>
										<span class="od-cc-nonneg-text">{item}</span>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				</section>
			{/if}

			<!-- ── Today's Rhythm ────────────────────────────────────────────── -->
			<section class="od-rhythm" aria-label="Today's Rhythm" in:fly={{ y: 12, duration: 300, delay: 60 }}>
				<div class="od-rhythm-header">
					<div class="od-rhythm-title-row">
						<span class="od-section-label">Today's Rhythm</span>
						{#if rhythmDate}
							<span class="od-rhythm-date">{formatDate(rhythmDate)}</span>
						{/if}
					</div>
					{#if rhythmMode}
						<div
							class="od-mode-badge"
							style="--mode-color: {MODE_COLORS[rhythmMode] ?? '#6b7280'}"
							aria-label="Cognitive mode: {rhythmMode}"
						>
							<span class="od-mode-dot"></span>
							{rhythmMode}
						</div>
					{/if}
				</div>

				{#if storeState.todayRhythm}
					<div class="od-rhythm-body">
						{#if rhythmPriorities.length > 0}
							<ul class="od-priorities" aria-label="Top priorities">
								{#each rhythmPriorities as p, i}
									<li class="od-priority-item">
										<span class="od-priority-num">{i + 1}</span>
										<span class="od-priority-text">{p}</span>
									</li>
								{/each}
							</ul>
						{:else}
							<p class="od-muted">No priorities parsed from today's plan.</p>
						{/if}
					</div>
				{:else}
					<div class="od-rhythm-empty">
						<p class="od-muted">No daily plan yet.</p>
						<p class="od-muted-sm">Run <code>rhythm/daily/boot.md</code> to generate today's file.</p>
					</div>
				{/if}
			</section>

			<!-- ── Quick Access ──────────────────────────────────────────────── -->
			<section aria-label="Quick access modules" in:fly={{ y: 12, duration: 300, delay: 120 }}>
				<div class="od-section-header">
					<span class="od-section-label">Quick Access</span>
				</div>
				<div class="od-qa-grid">
					{#each quickAccessModules as mod}
						<button
							class="od-qa-card"
							style="--qa-color: {mod.color}"
							onclick={() => goto(mod.href)}
							aria-label="Go to {mod.label}"
						>
							<div class="od-qa-icon">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true" width="18" height="18">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d={mod.icon} />
								</svg>
							</div>
							<span class="od-qa-label">{mod.label}</span>
						</button>
					{/each}
				</div>
			</section>

			<!-- ── Node Grid ─────────────────────────────────────────────────── -->
			<section aria-label="Node grid">
				<div class="od-section-header">
					<span class="od-section-label">Node Network</span>
				</div>

				<div class="od-grid">
					{#each storeState.nodes as node (node.slug)}
						<button
							class="od-card"
							onclick={() => goto(`/nodes/${node.slug}`)}
							aria-label="Open node {node.name}"
							in:fly={{ y: 16, duration: 250 }}
						>
							<!-- Card header -->
							<div class="od-card-header">
								<h3 class="od-card-name">{node.name}</h3>
								<span class="od-badge {nodeTypeColor(node.type)}" aria-label="Type: {node.type}">
									{shortType(node.type)}
								</span>
							</div>

							<!-- Slug -->
							<p class="od-card-slug">/{node.slug}</p>

							<!-- Signal preview -->
							{#if node.signal_md}
								<p class="od-card-preview">{extractSignalPreview(node.signal_md)}</p>
							{/if}

							<!-- Footer -->
							<div class="od-card-footer">
								<div class="od-signal-pill" class:od-signal-pill--active={node.has_signals}>
									<span class="od-signal-dot" class:od-signal-dot--active={node.has_signals}></span>
									<span>{node.signal_count} signal{node.signal_count !== 1 ? 's' : ''}</span>
								</div>
								<svg class="od-card-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5l7 7-7 7" />
								</svg>
							</div>
						</button>
					{/each}
				</div>
			</section>

		</div>
	{/if}

</div>

<style>
	/* ── Page shell ─────────────────────────────────────────────────────────── */
	.od-page {
		min-height: 100vh;
		background: #f8f9fa;
		color: #1c1c1e;
		padding: 2rem 2rem 4rem;
		font-family: inherit;
	}

	:global(.dark) .od-page {
		background: #0a0a0a;
		color: #e5e7eb;
	}

	.od-content {
		max-width: 1200px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	/* ── Loading / error ────────────────────────────────────────────────────── */
	.od-center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		gap: 1rem;
	}

	.od-spinner {
		width: 32px;
		height: 32px;
		border: 2px solid rgba(0, 0, 0, 0.08);
		border-top-color: rgba(0, 0, 0, 0.4);
		border-radius: 50%;
		animation: od-spin 0.8s linear infinite;
	}

	:global(.dark) .od-spinner {
		border-color: rgba(255, 255, 255, 0.08);
		border-top-color: rgba(255, 255, 255, 0.4);
	}

	@keyframes od-spin {
		to { transform: rotate(360deg); }
	}

	.od-muted { color: #6b7280; font-size: 0.875rem; }
	.od-muted-sm { color: #9ca3af; font-size: 0.75rem; margin-top: 0.25rem; }

	:global(.dark) .od-muted { color: rgba(255, 255, 255, 0.35); }
	:global(.dark) .od-muted-sm { color: rgba(255, 255, 255, 0.25); }

	.od-error-text { color: #dc2626; font-size: 0.875rem; }
	:global(.dark) .od-error-text { color: #f87171; }

	.od-retry-btn {
		padding: 0.5rem 1.25rem;
		border-radius: 0.5rem;
		background: rgba(0, 0, 0, 0.04);
		border: 1px solid rgba(0, 0, 0, 0.1);
		color: #1c1c1e;
		font-size: 0.875rem;
		cursor: pointer;
		transition: background 0.15s;
	}
	.od-retry-btn:hover { background: rgba(0, 0, 0, 0.08); }

	:global(.dark) .od-retry-btn {
		background: rgba(255, 255, 255, 0.06);
		border-color: rgba(255, 255, 255, 0.1);
		color: #e5e7eb;
	}
	:global(.dark) .od-retry-btn:hover { background: rgba(255, 255, 255, 0.1); }

	/* ── Stats bar ──────────────────────────────────────────────────────────── */
	.od-stats-bar {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 1rem 1.25rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-radius: 0.875rem;
	}

	:global(.dark) .od-stats-bar {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
	}

	.od-stat {
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
	}

	.od-stat-value {
		font-size: 1.375rem;
		font-weight: 600;
		color: #111827;
		line-height: 1;
	}

	:global(.dark) .od-stat-value { color: #f3f4f6; }

	.od-stat-value--active { color: #059669; }
	:global(.dark) .od-stat-value--active { color: #34d399; }

	.od-stat-value--dim { color: #d1d5db; }
	:global(.dark) .od-stat-value--dim { color: rgba(255, 255, 255, 0.3); }

	.od-stat-label {
		font-size: 0.75rem;
		color: #9ca3af;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	:global(.dark) .od-stat-label { color: rgba(255, 255, 255, 0.35); }

	.od-stat-sep {
		width: 1px;
		height: 1.5rem;
		background: rgba(0, 0, 0, 0.07);
	}

	:global(.dark) .od-stat-sep { background: rgba(255, 255, 255, 0.07); }

	/* ── Command Center ─────────────────────────────────────────────────────── */
	.od-cc-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0.75rem;
		margin-bottom: 0.875rem;
	}

	@media (max-width: 1024px) {
		.od-cc-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 640px) {
		.od-cc-grid { grid-template-columns: repeat(2, 1fr); }
	}

	.od-cc-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.875rem 1rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-radius: 0.875rem;
	}

	:global(.dark) .od-cc-card {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
	}

	.od-cc-icon {
		flex-shrink: 0;
		width: 32px;
		height: 32px;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.od-cc-icon--nodes    { background: rgba(59, 130, 246, 0.1);  color: #3b82f6; }
	.od-cc-icon--signals  { background: rgba(139, 92, 246, 0.1);  color: #8b5cf6; }
	.od-cc-icon--revenue  { background: rgba(16, 185, 129, 0.1);  color: #059669; }
	.od-cc-icon--entities { background: rgba(245, 158, 11, 0.1);  color: #d97706; }

	:global(.dark) .od-cc-icon--nodes    { background: rgba(59, 130, 246, 0.15);  color: #93c5fd; }
	:global(.dark) .od-cc-icon--signals  { background: rgba(139, 92, 246, 0.15);  color: #c4b5fd; }
	:global(.dark) .od-cc-icon--revenue  { background: rgba(16, 185, 129, 0.15);  color: #34d399; }
	:global(.dark) .od-cc-icon--entities { background: rgba(245, 158, 11, 0.15);  color: #fbbf24; }

	.od-cc-body {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.od-cc-value {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		line-height: 1;
	}

	:global(.dark) .od-cc-value { color: #f3f4f6; }

	.od-cc-label {
		font-size: 0.68rem;
		font-weight: 500;
		color: #9ca3af;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	:global(.dark) .od-cc-label { color: rgba(255, 255, 255, 0.35); }

	/* Non-negotiables strip */
	.od-cc-nonneg {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 0.875rem 1rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-radius: 0.875rem;
	}

	:global(.dark) .od-cc-nonneg {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
	}

	.od-cc-nonneg-label {
		flex-shrink: 0;
		font-size: 0.68rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #dc2626;
		padding-top: 0.1rem;
	}

	:global(.dark) .od-cc-nonneg-label { color: #f87171; }

	.od-cc-nonneg-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.od-cc-nonneg-item {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.2rem 0.6rem;
		background: rgba(220, 38, 38, 0.06);
		border: 1px solid rgba(220, 38, 38, 0.15);
		border-radius: 999px;
	}

	:global(.dark) .od-cc-nonneg-item {
		background: rgba(248, 113, 113, 0.08);
		border-color: rgba(248, 113, 113, 0.18);
	}

	.od-cc-nonneg-num {
		font-size: 0.62rem;
		font-weight: 700;
		color: #dc2626;
	}

	:global(.dark) .od-cc-nonneg-num { color: #f87171; }

	.od-cc-nonneg-text {
		font-size: 0.75rem;
		color: #374151;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 20ch;
	}

	:global(.dark) .od-cc-nonneg-text { color: #d1d5db; }

	/* ── Section labels ─────────────────────────────────────────────────────── */
	.od-section-label {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #9ca3af;
	}

	:global(.dark) .od-section-label { color: rgba(255, 255, 255, 0.3); }

	.od-section-header {
		margin-bottom: 1rem;
	}

	/* ── Rhythm section ─────────────────────────────────────────────────────── */
	.od-rhythm {
		padding: 1.25rem 1.5rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-radius: 0.875rem;
	}

	:global(.dark) .od-rhythm {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
	}

	.od-rhythm-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.od-rhythm-title-row {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.od-rhythm-date {
		font-size: 0.8rem;
		color: #9ca3af;
	}

	:global(.dark) .od-rhythm-date { color: rgba(255, 255, 255, 0.4); }

	.od-mode-badge {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.25rem 0.75rem;
		border-radius: 999px;
		background: color-mix(in srgb, var(--mode-color) 15%, transparent);
		border: 1px solid color-mix(in srgb, var(--mode-color) 35%, transparent);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		color: var(--mode-color);
	}

	.od-mode-dot {
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: var(--mode-color);
		box-shadow: 0 0 6px var(--mode-color);
	}

	.od-rhythm-body { margin-top: 0.25rem; }

	.od-priorities {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.od-priority-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}

	.od-priority-num {
		flex-shrink: 0;
		width: 1.25rem;
		height: 1.25rem;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.05);
		border: 1px solid rgba(0, 0, 0, 0.08);
		font-size: 0.65rem;
		font-weight: 700;
		color: #6b7280;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-top: 1px;
	}

	:global(.dark) .od-priority-num {
		background: rgba(255, 255, 255, 0.07);
		border-color: rgba(255, 255, 255, 0.1);
		color: rgba(255, 255, 255, 0.5);
	}

	.od-priority-text {
		font-size: 0.875rem;
		color: #374151;
		line-height: 1.4;
	}

	:global(.dark) .od-priority-text { color: #d1d5db; }

	.od-rhythm-empty {
		padding: 0.5rem 0;
	}

	/* ── Quick Access grid ──────────────────────────────────────────────────── */
	.od-qa-grid {
		display: grid;
		grid-template-columns: repeat(8, 1fr);
		gap: 0.625rem;
	}

	@media (max-width: 1024px) {
		.od-qa-grid { grid-template-columns: repeat(4, 1fr); }
	}

	@media (max-width: 640px) {
		.od-qa-grid { grid-template-columns: repeat(4, 1fr); }
	}

	.od-qa-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 0.875rem 0.5rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-left: 3px solid var(--qa-color);
		border-radius: 0.75rem;
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s, transform 0.1s;
		width: 100%;
	}

	.od-qa-card:hover {
		background: rgba(0, 0, 0, 0.02);
		border-color: rgba(0, 0, 0, 0.1);
		border-left-color: var(--qa-color);
		transform: translateY(-1px);
	}

	.od-qa-card:focus-visible {
		outline: 2px solid var(--qa-color);
		outline-offset: 2px;
	}

	:global(.dark) .od-qa-card {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
		border-left-color: var(--qa-color);
	}

	:global(.dark) .od-qa-card:hover {
		background: rgba(255, 255, 255, 0.055);
		border-color: rgba(255, 255, 255, 0.12);
		border-left-color: var(--qa-color);
	}

	.od-qa-icon {
		color: var(--qa-color);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0.85;
		transition: opacity 0.15s;
	}

	.od-qa-card:hover .od-qa-icon { opacity: 1; }

	.od-qa-label {
		font-size: 0.7rem;
		font-weight: 600;
		color: #374151;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	:global(.dark) .od-qa-label { color: #d1d5db; }

	/* ── Node grid ──────────────────────────────────────────────────────────── */
	.od-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.875rem;
	}

	@media (max-width: 1024px) {
		.od-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 640px) {
		.od-grid { grid-template-columns: 1fr; }
		.od-page { padding: 1rem 1rem 3rem; }
	}

	/* ── Node card ──────────────────────────────────────────────────────────── */
	.od-card {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 1rem 1.125rem;
		background: #ffffff;
		border: 1px solid rgba(0, 0, 0, 0.06);
		border-radius: 0.875rem;
		cursor: pointer;
		text-align: left;
		width: 100%;
		color: inherit;
		transition: background 0.15s, border-color 0.15s, transform 0.1s;
	}

	.od-card:hover {
		background: rgba(0, 0, 0, 0.02);
		border-color: rgba(0, 0, 0, 0.12);
		transform: translateY(-1px);
	}

	.od-card:focus-visible {
		outline: 2px solid rgba(0, 0, 0, 0.2);
		outline-offset: 2px;
	}

	:global(.dark) .od-card {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.06);
	}

	:global(.dark) .od-card:hover {
		background: rgba(255, 255, 255, 0.055);
		border-color: rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .od-card:focus-visible {
		outline-color: rgba(255, 255, 255, 0.3);
	}

	.od-card-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.od-card-name {
		font-size: 0.875rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
		line-height: 1.3;
	}

	:global(.dark) .od-card-name { color: #f3f4f6; }

	.od-card-slug {
		font-size: 0.7rem;
		color: #d1d5db;
		margin: 0;
		font-family: ui-monospace, monospace;
	}

	:global(.dark) .od-card-slug { color: rgba(255, 255, 255, 0.25); }

	.od-card-preview {
		font-size: 0.75rem;
		color: #6b7280;
		margin: 0;
		line-height: 1.5;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		flex: 1;
	}

	:global(.dark) .od-card-preview { color: rgba(255, 255, 255, 0.4); }

	.od-card-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: auto;
		padding-top: 0.25rem;
	}

	.od-card-arrow {
		width: 14px;
		height: 14px;
		color: #d1d5db;
		flex-shrink: 0;
		transition: color 0.15s, transform 0.15s;
	}

	.od-card:hover .od-card-arrow {
		color: #6b7280;
		transform: translateX(2px);
	}

	:global(.dark) .od-card-arrow { color: rgba(255, 255, 255, 0.2); }
	:global(.dark) .od-card:hover .od-card-arrow { color: rgba(255, 255, 255, 0.5); }

	/* ── Signal pill ────────────────────────────────────────────────────────── */
	.od-signal-pill {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.68rem;
		color: #9ca3af;
		letter-spacing: 0.03em;
	}

	:global(.dark) .od-signal-pill { color: rgba(255, 255, 255, 0.3); }

	.od-signal-pill--active { color: #059669; }
	:global(.dark) .od-signal-pill--active { color: rgba(52, 211, 153, 0.7); }

	.od-signal-dot {
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: #d1d5db;
		flex-shrink: 0;
	}

	:global(.dark) .od-signal-dot { background: rgba(255, 255, 255, 0.15); }

	.od-signal-dot--active {
		background: #10b981;
		box-shadow: 0 0 5px rgba(16, 185, 129, 0.5);
	}

	:global(.dark) .od-signal-dot--active {
		background: #34d399;
		box-shadow: 0 0 5px rgba(52, 211, 153, 0.5);
	}

	/* ── Type badges ────────────────────────────────────────────────────────── */
	.od-badge {
		flex-shrink: 0;
		padding: 0.15rem 0.5rem;
		border-radius: 999px;
		font-size: 0.62rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		white-space: nowrap;
	}

	.od-badge--person    { background: rgba(99, 102, 241, 0.12); color: #6366f1; border: 1px solid rgba(99, 102, 241, 0.2); }
	.od-badge--entity    { background: rgba(14, 165, 233, 0.1);  color: #0ea5e9; border: 1px solid rgba(14, 165, 233, 0.18); }
	.od-badge--operation { background: rgba(59, 130, 246, 0.1);  color: #3b82f6; border: 1px solid rgba(59, 130, 246, 0.18); }
	.od-badge--research  { background: rgba(139, 92, 246, 0.1);  color: #8b5cf6; border: 1px solid rgba(139, 92, 246, 0.18); }
	.od-badge--network   { background: rgba(20, 184, 166, 0.1);  color: #14b8a6; border: 1px solid rgba(20, 184, 166, 0.18); }
	.od-badge--community { background: rgba(245, 158, 11, 0.1);  color: #d97706; border: 1px solid rgba(245, 158, 11, 0.18); }
	.od-badge--media     { background: rgba(236, 72, 153, 0.1);  color: #db2777; border: 1px solid rgba(236, 72, 153, 0.18); }
	.od-badge--inbox     { background: rgba(107, 114, 128, 0.1); color: #6b7280; border: 1px solid rgba(107, 114, 128, 0.18); }
	.od-badge--registry  { background: rgba(16, 185, 129, 0.1);  color: #059669; border: 1px solid rgba(16, 185, 129, 0.18); }
	.od-badge--cross     { background: rgba(249, 115, 22, 0.1);  color: #ea580c; border: 1px solid rgba(249, 115, 22, 0.18); }
	.od-badge--command   { background: rgba(239, 68, 68, 0.1);   color: #dc2626; border: 1px solid rgba(239, 68, 68, 0.18); }
	.od-badge--operating { background: rgba(52, 211, 153, 0.1);  color: #059669; border: 1px solid rgba(52, 211, 153, 0.18); }
	.od-badge--default   { background: rgba(0, 0, 0, 0.04);      color: #6b7280; border: 1px solid rgba(0, 0, 0, 0.08); }

	:global(.dark) .od-badge--person    { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; border-color: rgba(99, 102, 241, 0.25); }
	:global(.dark) .od-badge--entity    { background: rgba(14, 165, 233, 0.12); color: #7dd3fc; border-color: rgba(14, 165, 233, 0.22); }
	:global(.dark) .od-badge--operation { background: rgba(59, 130, 246, 0.12); color: #93c5fd; border-color: rgba(59, 130, 246, 0.22); }
	:global(.dark) .od-badge--research  { background: rgba(139, 92, 246, 0.12); color: #c4b5fd; border-color: rgba(139, 92, 246, 0.22); }
	:global(.dark) .od-badge--network   { background: rgba(20, 184, 166, 0.12); color: #5eead4; border-color: rgba(20, 184, 166, 0.22); }
	:global(.dark) .od-badge--community { background: rgba(245, 158, 11, 0.12); color: #fcd34d; border-color: rgba(245, 158, 11, 0.22); }
	:global(.dark) .od-badge--media     { background: rgba(236, 72, 153, 0.12); color: #f9a8d4; border-color: rgba(236, 72, 153, 0.22); }
	:global(.dark) .od-badge--inbox     { background: rgba(107, 114, 128, 0.12); color: #9ca3af; border-color: rgba(107, 114, 128, 0.22); }
	:global(.dark) .od-badge--registry  { background: rgba(16, 185, 129, 0.12); color: #6ee7b7; border-color: rgba(16, 185, 129, 0.22); }
	:global(.dark) .od-badge--cross     { background: rgba(249, 115, 22, 0.12); color: #fdba74; border-color: rgba(249, 115, 22, 0.22); }
	:global(.dark) .od-badge--command   { background: rgba(239, 68, 68, 0.12);  color: #fca5a5; border-color: rgba(239, 68, 68, 0.22); }
	:global(.dark) .od-badge--operating { background: rgba(52, 211, 153, 0.12); color: #6ee7b7; border-color: rgba(52, 211, 153, 0.22); }
	:global(.dark) .od-badge--default   { background: rgba(255, 255, 255, 0.06); color: rgba(255, 255, 255, 0.4); border-color: rgba(255, 255, 255, 0.1); }
</style>
