<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getUsageSummary,
		getUsageByProvider,
		getUsageByModel,
		getUsageByAgent,
		getUsageTrend
	} from '$lib/api/usage/usage';
	import { analyticsNotes } from '$lib/stores/analyticsNotes';
	import type {
		UsageSummary,
		ProviderUsage,
		ModelUsage,
		AgentUsage,
		UsageTrendPoint
	} from '$lib/api/usage/types';

	// ─── Fallback Mock Data (shown when API is unavailable) ───────────────────────

	// Seeded pseudo-random to keep SSR/CSR consistent
	function seededRand(seed: number): number {
		const x = Math.sin(seed + 1) * 10000;
		return x - Math.floor(x);
	}

	const MOCK_SUMMARY: UsageSummary = {
		total_requests: 4291,
		total_input_tokens: 5_234_100,
		total_output_tokens: 3_148_200,
		total_tokens: 8_382_300,
		total_cost: 24.87,
		period: 'month',
		start_date: '2026-03-01',
		end_date: '2026-03-24'
	};

	const MOCK_PROVIDERS: ProviderUsage[] = [
		{ provider: 'Anthropic', request_count: 2847, total_input_tokens: 3_102_000, total_output_tokens: 2_789_000, total_tokens: 5_891_000, total_cost: 18.42 },
		{ provider: 'OpenAI', request_count: 891, total_input_tokens: 1_041_000, total_output_tokens: 882_000, total_tokens: 1_923_000, total_cost: 5.67 },
		{ provider: 'Ollama', request_count: 553, total_input_tokens: 301_000, total_output_tokens: 267_300, total_tokens: 568_300, total_cost: 0 }
	];

	const MOCK_MODELS: ModelUsage[] = [
		{ model: 'claude-sonnet-4', provider: 'Anthropic', request_count: 1893, total_input_tokens: 1_890_000, total_output_tokens: 1_531_000, total_tokens: 3_421_000, total_cost: 10.26 },
		{ model: 'claude-haiku-4', provider: 'Anthropic', request_count: 672, total_input_tokens: 712_000, total_output_tokens: 522_000, total_tokens: 1_234_000, total_cost: 2.47 },
		{ model: 'gpt-4o', provider: 'OpenAI', request_count: 534, total_input_tokens: 601_000, total_output_tokens: 501_000, total_tokens: 1_102_000, total_cost: 3.31 },
		{ model: 'claude-opus-4', provider: 'Anthropic', request_count: 282, total_input_tokens: 700_000, total_output_tokens: 536_000, total_tokens: 1_236_000, total_cost: 5.69 },
		{ model: 'gpt-4o-mini', provider: 'OpenAI', request_count: 357, total_input_tokens: 440_000, total_output_tokens: 381_000, total_tokens: 821_000, total_cost: 2.36 },
		{ model: 'llama-3.3', provider: 'Ollama', request_count: 553, total_input_tokens: 301_000, total_output_tokens: 267_300, total_tokens: 568_300, total_cost: 0 }
	];

	const MOCK_AGENTS: AgentUsage[] = [
		{ agent_name: 'General Assistant', request_count: 891, total_input_tokens: 1_100_000, total_output_tokens: 1_034_000, total_tokens: 2_134_000, avg_duration_ms: 1240 },
		{ agent_name: 'Support Agent', request_count: 512, total_input_tokens: 430_000, total_output_tokens: 404_000, total_tokens: 834_000, avg_duration_ms: 890 },
		{ agent_name: 'Code Reviewer', request_count: 289, total_input_tokens: 810_000, total_output_tokens: 757_000, total_tokens: 1_567_000, avg_duration_ms: 2100 },
		{ agent_name: 'Sales Closer', request_count: 142, total_input_tokens: 218_000, total_output_tokens: 205_000, total_tokens: 423_000, avg_duration_ms: 1560 },
		{ agent_name: 'Deep Researcher', request_count: 78, total_input_tokens: 460_000, total_output_tokens: 432_000, total_tokens: 892_000, avg_duration_ms: 4200 }
	];

	const MOCK_TREND: UsageTrendPoint[] = Array.from({ length: 30 }, (_, i) => {
		const date = new Date(2026, 2, i + 1);
		const isWeekend = date.getDay() === 0 || date.getDay() === 6;
		const base = isWeekend ? 80 : 150;
		const variation = 0.6 + seededRand(i) * 0.8;
		const requests = Math.round(base * variation);
		const tokens = requests * (1800 + Math.round(seededRand(i + 100) * 600));
		return {
			date: date.toISOString().split('T')[0],
			ai_requests: requests,
			total_tokens: tokens,
			estimated_cost: Number((tokens * 0.000003).toFixed(2)),
			mcp_requests: Math.round(requests * 0.3),
			messages_sent: Math.round(requests * 1.4)
		};
	});

	// ─── Live Data State ──────────────────────────────────────────────────────────
	let isLoading = $state(true);
	let isUsingMockData = $state(false);

	let summary = $state<UsageSummary>(MOCK_SUMMARY);
	let providers = $state<ProviderUsage[]>(MOCK_PROVIDERS);
	let models = $state<ModelUsage[]>(MOCK_MODELS);
	let agents = $state<AgentUsage[]>(MOCK_AGENTS);
	let trendData = $state<UsageTrendPoint[]>(MOCK_TREND);

	// ─── State ────────────────────────────────────────────────────────────────────
	let activePeriod = $state<'7D' | '14D' | '30D' | 'All'>('30D');
	let hoverIndex = $state<number | null>(null);
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	// ─── Note Modal State ─────────────────────────────────────────────────────────
	let showNoteModal = $state(false);
	let noteModalDate = $state('');
	let noteModalMetric = $state<{ requests: number; tokens: number; cost: number; label: string } | null>(null);
	let noteContent = $state('');
	let existingNoteId = $state<string | null>(null);

	// ─── Load Real Data ───────────────────────────────────────────────────────────
	onMount(async () => {
		try {
			// TODO: map activePeriod selector to API period param when period switching lands
			const [summaryRes, providersRes, modelsRes, agentsRes, trendRes] = await Promise.all([
				getUsageSummary('month'),
				getUsageByProvider('month'),
				getUsageByModel('month'),
				getUsageByAgent('month'),
				getUsageTrend()
			]);

			// Only replace mock data if the API returned meaningful results
			const hasSummary = summaryRes && (summaryRes.total_requests > 0 || summaryRes.total_tokens > 0 || summaryRes.total_cost > 0);

			if (hasSummary) {
				summary = summaryRes;
			}
			if (providersRes && providersRes.length > 0) {
				providers = providersRes;
			}
			if (modelsRes && modelsRes.length > 0) {
				models = modelsRes;
			}
			if (agentsRes && agentsRes.length > 0) {
				agents = agentsRes;
			}
			if (trendRes && trendRes.length > 0) {
				trendData = trendRes;
			}

			// Show mock badge only if none of the endpoints returned real data
			const hasRealData = hasSummary ||
				(providersRes && providersRes.length > 0) ||
				(modelsRes && modelsRes.length > 0) ||
				(agentsRes && agentsRes.length > 0) ||
				(trendRes && trendRes.length > 0);

			isUsingMockData = !hasRealData;
		} catch {
			// API unavailable — keep mock data visible with indicator
			isUsingMockData = true;
		} finally {
			isLoading = false;
		}
	});

	// ─── Chart dimensions ─────────────────────────────────────────────────────────
	const SVG_WIDTH = 900;
	const SVG_HEIGHT = 200;
	const PAD = { top: 16, right: 24, bottom: 32, left: 48 };

	// ─── Utility functions ────────────────────────────────────────────────────────
	function formatNumber(num: number): string {
		if (num >= 1_000_000) return (num / 1_000_000).toFixed(1) + 'M';
		if (num >= 1_000) return (num / 1_000).toFixed(1) + 'K';
		return num.toString();
	}

	function formatCost(cost: number): string {
		return '$' + cost.toFixed(2);
	}

	function formatDuration(ms: number): string {
		if (ms < 1000) return `${Math.round(ms)}ms`;
		return `${(ms / 1000).toFixed(1)}s`;
	}

	function fmtDate(iso: string): string {
		const d = new Date(iso);
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	// ─── Sparkline ────────────────────────────────────────────────────────────────
	function sparklinePath(values: number[]): string {
		const min = Math.min(...values);
		const max = Math.max(...values);
		const range = max - min || 1;
		const w = 60;
		const h = 24;
		return values
			.map((v, i) => {
				const x = (i / (values.length - 1)) * w;
				const y = h - ((v - min) / range) * h;
				return `${x},${y}`;
			})
			.join(' ');
	}

	// Sparkline series derived from trend data (last 7 points)
	const sparkSpend   = $derived(trendData.slice(-7).map(d => d.estimated_cost * 100));
	const sparkReq     = $derived(trendData.slice(-7).map(d => d.ai_requests));
	const sparkTokens  = $derived(trendData.slice(-7).map(d => d.total_tokens));
	const sparkAvgCost = $derived(
		trendData.slice(-7).map(d => d.ai_requests > 0 ? (d.estimated_cost / d.ai_requests) * 10000 : 0)
	);

	// KPI derived values
	const avgCostPerReq = $derived(
		summary.total_requests > 0 ? summary.total_cost / summary.total_requests : 0
	);

	// ─── Area Chart ───────────────────────────────────────────────────────────────
	const chartW = SVG_WIDTH - PAD.left - PAD.right;
	const chartH = SVG_HEIGHT - PAD.top - PAD.bottom;

	const maxVal = $derived(Math.max(...trendData.map(d => d.ai_requests), 1));
	const niceMax = $derived(Math.ceil(maxVal / 50) * 50);

	const chartPoints = $derived(
		trendData.map((d, i) => ({
			x: PAD.left + (i / Math.max(trendData.length - 1, 1)) * chartW,
			y: PAD.top + chartH - (d.ai_requests / niceMax) * chartH,
			data: d
		}))
	);

	const linePath = $derived(
		chartPoints.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(' ')
	);

	const areaPath = $derived(
		linePath +
		` L${chartPoints[chartPoints.length - 1].x.toFixed(2)},${(PAD.top + chartH).toFixed(2)}` +
		` L${chartPoints[0].x.toFixed(2)},${(PAD.top + chartH).toFixed(2)} Z`
	);

	const gridLines = $derived(
		[0, 0.25, 0.5, 0.75, 1].map(frac => ({
			y: PAD.top + chartH - frac * chartH,
			label: Math.round(frac * niceMax).toString()
		}))
	);

	const xLabels = $derived(
		trendData
			.map((d, i) => ({ i, d }))
			.filter(({ i }) => i % 5 === 0)
			.map(({ i, d }) => ({
				x: PAD.left + (i / Math.max(trendData.length - 1, 1)) * chartW,
				label: fmtDate(d.date)
			}))
	);

	// ─── Chart hover ─────────────────────────────────────────────────────────────
	function handleChartMouseMove(e: MouseEvent) {
		const svg = e.currentTarget as SVGSVGElement;
		const rect = svg.getBoundingClientRect();
		const scaleX = SVG_WIDTH / rect.width;
		const mouseX = (e.clientX - rect.left) * scaleX;

		let nearest = 0;
		let minDist = Infinity;
		chartPoints.forEach((p, i) => {
			const dist = Math.abs(p.x - mouseX);
			if (dist < minDist) { minDist = dist; nearest = i; }
		});

		hoverIndex = nearest;
		tooltipX = e.clientX - rect.left;
		tooltipY = e.clientY - rect.top;
	}

	function handleChartMouseLeave() {
		hoverIndex = null;
	}

	// ─── Note Modal Handlers ──────────────────────────────────────────────────────
	function handleChartClick() {
		if (hoverIndex === null) return;
		const point = trendData[hoverIndex];
		noteModalDate = point.date;
		noteModalMetric = {
			requests: point.ai_requests,
			tokens: point.total_tokens,
			cost: point.estimated_cost,
			label: fmtDate(point.date)
		};
		// Check if a note already exists for this date
		const existing = analyticsNotes.getNotesForDate(point.date);
		if (existing.length > 0) {
			noteContent = existing[0].content;
			existingNoteId = existing[0].id;
		} else {
			noteContent = '';
			existingNoteId = null;
		}
		showNoteModal = true;
	}

	function saveNote() {
		if (!noteContent.trim() || !noteModalMetric) return;
		if (existingNoteId) {
			analyticsNotes.updateNote(existingNoteId, noteContent);
		} else {
			analyticsNotes.addNote(noteModalDate, noteContent, noteModalMetric);
		}
		showNoteModal = false;
		noteContent = '';
		existingNoteId = null;
	}

	function deleteNoteFromModal() {
		if (existingNoteId) {
			analyticsNotes.deleteNote(existingNoteId);
		}
		showNoteModal = false;
		noteContent = '';
		existingNoteId = null;
	}

	// ─── Note indicator derived set ───────────────────────────────────────────────
	const noteDates = $derived(new Set(
		analyticsNotes.getAllNotes().map(n => n.date)
	));

	// ─── Provider bar widths ─────────────────────────────────────────────────────
	const totalProviderTokens = $derived(providers.reduce((s, p) => s + p.total_tokens, 0));
	const maxAgentReq = $derived(Math.max(...agents.map(a => a.request_count), 1));
</script>

<div class="anl-page">

	<!-- ── Header ───────────────────────────────────────────────────────────── -->
	<div class="anl-header">
		<div class="anl-header-text">
			<h1 class="anl-title">Analytics</h1>
			<p class="anl-subtitle">
				AI usage, costs, and performance insights
				{#if isUsingMockData}
					<span class="anl-demo-badge" title="No live usage data available — showing sample data">Sample data</span>
				{/if}
				{#if isLoading}
					<span class="anl-loading-badge">Loading...</span>
				{/if}
			</p>
		</div>
		<div class="anl-period-selector" role="group" aria-label="Select time period">
			{#each (['7D', '14D', '30D', 'All'] as const) as period}
				<button
					class="anl-period-btn"
					class:anl-period-btn--active={activePeriod === period}
					onclick={() => activePeriod = period}
					aria-pressed={activePeriod === period}
				>
					{period}
				</button>
			{/each}
		</div>
	</div>

	<!-- ── KPI Ribbon ───────────────────────────────────────────────────────── -->
	<div class="anl-kpi-ribbon">

		<!-- Total Spend -->
		<div class="anl-kpi-card">
			<div class="anl-kpi-left">
				<span class="anl-kpi-label">Total Spend</span>
				<span class="anl-kpi-value">{formatCost(summary.total_cost)}</span>
			</div>
			<svg class="anl-sparkline" viewBox="0 0 60 24" aria-hidden="true">
				<polyline
					points={sparklinePath(sparkSpend)}
					fill="none"
					stroke="var(--bos-accent-blue)"
					stroke-width="1.5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</div>

		<!-- Requests -->
		<div class="anl-kpi-card">
			<div class="anl-kpi-left">
				<span class="anl-kpi-label">Requests</span>
				<span class="anl-kpi-value">{summary.total_requests.toLocaleString()}</span>
			</div>
			<svg class="anl-sparkline" viewBox="0 0 60 24" aria-hidden="true">
				<polyline
					points={sparklinePath(sparkReq)}
					fill="none"
					stroke="var(--bos-accent-blue)"
					stroke-width="1.5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</div>

		<!-- Tokens -->
		<div class="anl-kpi-card">
			<div class="anl-kpi-left">
				<span class="anl-kpi-label">Tokens</span>
				<span class="anl-kpi-value">{formatNumber(summary.total_tokens)}</span>
			</div>
			<svg class="anl-sparkline" viewBox="0 0 60 24" aria-hidden="true">
				<polyline
					points={sparklinePath(sparkTokens)}
					fill="none"
					stroke="var(--bos-accent-blue)"
					stroke-width="1.5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</div>

		<!-- Avg Cost/Req -->
		<div class="anl-kpi-card">
			<div class="anl-kpi-left">
				<span class="anl-kpi-label">Avg Cost / Req</span>
				<span class="anl-kpi-value">{formatCost(avgCostPerReq)}</span>
			</div>
			<svg class="anl-sparkline" viewBox="0 0 60 24" aria-hidden="true">
				<polyline
					points={sparklinePath(sparkAvgCost)}
					fill="none"
					stroke="var(--bos-accent-blue)"
					stroke-width="1.5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</div>

	</div>

	<!-- ── Trend Chart ──────────────────────────────────────────────────────── -->
	<div class="anl-card anl-chart-card">
		<div class="anl-card-header">
			<span class="anl-section-title">Request Volume — 30 Days</span>
		</div>

		<div class="anl-chart-wrap">
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<svg
				class="anl-chart-svg"
				viewBox="0 0 {SVG_WIDTH} {SVG_HEIGHT}"
				preserveAspectRatio="none"
				aria-label="30-day request volume trend chart"
				role="img"
				onmousemove={handleChartMouseMove}
				onmouseleave={handleChartMouseLeave}
				onclick={handleChartClick}
			>
				<defs>
					<linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="rgba(59,130,246,0.18)" />
						<stop offset="100%" stop-color="rgba(59,130,246,0)" />
					</linearGradient>
				</defs>

				<!-- Grid lines -->
				{#each gridLines as gl}
					<line
						x1={PAD.left}
						y1={gl.y}
						x2={SVG_WIDTH - PAD.right}
						y2={gl.y}
						stroke="var(--dbd2)"
						stroke-width="1"
					/>
					<text
						x={PAD.left - 6}
						y={gl.y + 4}
						text-anchor="end"
						class="anl-axis-label"
					>{gl.label}</text>
				{/each}

				<!-- X-axis labels -->
				{#each xLabels as xl}
					<text
						x={xl.x}
						y={SVG_HEIGHT - 6}
						text-anchor="middle"
						class="anl-axis-label"
					>{xl.label}</text>
				{/each}

				<!-- Area fill -->
				<path d={areaPath} fill="url(#areaGradient)" />

				<!-- Line -->
				<path
					d={linePath}
					fill="none"
					stroke="var(--bos-accent-blue)"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>

				<!-- Data dots (only on hover vicinity) -->
				{#each chartPoints as p, i}
					<!-- Note indicator dot (always visible when note exists for this date) -->
					{#if noteDates.has(p.data.date)}
						<circle
							class="anl-note-indicator"
							cx={p.x}
							cy={PAD.top + chartH + 8}
							r="3"
							fill="var(--bos-accent-blue)"
							opacity="0.7"
						/>
					{/if}
					{#if hoverIndex === i}
						<circle
							cx={p.x}
							cy={p.y}
							r="4"
							fill="var(--bos-accent-blue)"
							stroke="var(--dbg)"
							stroke-width="2"
						/>
						<!-- Crosshair -->
						<line
							x1={p.x}
							y1={PAD.top}
							x2={p.x}
							y2={PAD.top + chartH}
							stroke="var(--bos-accent-blue)"
							stroke-width="1"
							stroke-dasharray="3,3"
							opacity="0.5"
						/>
					{/if}
				{/each}
			</svg>

			<!-- Tooltip -->
			{#if hoverIndex !== null}
				{@const hd = trendData[hoverIndex]}
				<div
					class="anl-tooltip"
					style="left: {tooltipX + 12}px; top: {tooltipY - 8}px"
					role="tooltip"
				>
					<span class="anl-tooltip-date">{fmtDate(hd.date)}</span>
					<span class="anl-tooltip-row">
						<span class="anl-tooltip-lbl">Requests</span>
						<span class="anl-tooltip-val">{hd.ai_requests.toLocaleString()}</span>
					</span>
					<span class="anl-tooltip-row">
						<span class="anl-tooltip-lbl">Tokens</span>
						<span class="anl-tooltip-val">{formatNumber(hd.total_tokens)}</span>
					</span>
					<span class="anl-tooltip-row">
						<span class="anl-tooltip-lbl">Cost</span>
						<span class="anl-tooltip-val">{formatCost(hd.estimated_cost)}</span>
					</span>
					<span class="anl-tooltip-click-hint">
						{noteDates.has(hd.date) ? 'Click to edit note' : 'Click to add a note'}
					</span>
				</div>
			{/if}
		</div>
	</div>

	<!-- ── Two-column row ───────────────────────────────────────────────────── -->
	<div class="anl-two-col">

		<!-- Provider Breakdown -->
		<div class="anl-card">
			<div class="anl-card-header">
				<span class="anl-section-title">By Provider</span>
			</div>
			<div class="anl-provider-list">
				{#each providers as prov}
					{@const pct = totalProviderTokens > 0 ? (prov.total_tokens / totalProviderTokens) * 100 : 0}
					<div class="anl-provider-row">
						<div class="anl-provider-top">
							<span class="anl-provider-name">{prov.provider}</span>
							<span class="anl-provider-meta">
								<span class="anl-num">{formatNumber(prov.total_tokens)}</span>
								<span class="anl-provider-sep">·</span>
								<span class="anl-num">{prov.total_cost > 0 ? formatCost(prov.total_cost) : 'Free'}</span>
							</span>
						</div>
						<div class="anl-bar-track" role="progressbar" aria-valuenow={Math.round(pct)} aria-valuemin={0} aria-valuemax={100} aria-label="{prov.provider} token share">
							<div class="anl-bar-fill" style="width: {pct.toFixed(1)}%"></div>
						</div>
						<span class="anl-provider-pct">{pct.toFixed(0)}% of tokens</span>
					</div>
				{/each}
			</div>
		</div>

		<!-- Top Models -->
		<div class="anl-card">
			<div class="anl-card-header">
				<span class="anl-section-title">Top Models</span>
			</div>
			<div class="anl-model-list">
				{#each models as mdl, i}
					<div class="anl-model-row">
						<span class="anl-model-rank">{i + 1}</span>
						<div class="anl-model-info">
							<span class="anl-model-name">{mdl.model}</span>
							<span class="anl-provider-badge">{mdl.provider}</span>
						</div>
						<div class="anl-model-stats">
							<span class="anl-num">{mdl.request_count.toLocaleString()} req</span>
							<span class="anl-model-cost anl-num">{mdl.total_cost > 0 ? formatCost(mdl.total_cost) : 'Free'}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>

	</div>

	<!-- ── Agent Performance ─────────────────────────────────────────────────── -->
	<div class="anl-card">
		<div class="anl-card-header">
			<span class="anl-section-title">Agent Performance</span>
		</div>

		<div class="anl-agent-table" role="table" aria-label="Agent performance metrics">
			<div class="anl-agent-thead" role="row">
				<span role="columnheader">Agent</span>
				<span role="columnheader">Requests</span>
				<span role="columnheader">Tokens</span>
				<span role="columnheader">Avg Duration</span>
				<span role="columnheader" class="anl-col-volume">Volume</span>
			</div>
			{#each agents as agent}
				<div class="anl-agent-row" role="row">
					<span class="anl-agent-name" role="cell">{agent.agent_name}</span>
					<span class="anl-num" role="cell">{agent.request_count.toLocaleString()}</span>
					<span class="anl-num" role="cell">{formatNumber(agent.total_tokens)}</span>
					<span class="anl-num" role="cell">{formatDuration(agent.avg_duration_ms)}</span>
					<span class="anl-agent-bar-cell anl-col-volume" role="cell" aria-label="{agent.request_count} out of {maxAgentReq} requests">
						<div class="anl-agent-bar-track">
							<div
								class="anl-agent-bar-fill"
								style="width: {((agent.request_count / maxAgentReq) * 100).toFixed(1)}%"
							></div>
						</div>
					</span>
				</div>
			{/each}
		</div>
	</div>

</div>

<!-- ── Note Modal ─────────────────────────────────────────────────────────── -->
{#if showNoteModal && noteModalMetric}
<div class="anl-note-overlay" onclick={() => showNoteModal = false} role="dialog" aria-modal="true" aria-label="Add note for {noteModalMetric.label}">
	<div class="anl-note-modal" onclick={(e) => e.stopPropagation()}>
		<div class="anl-note-modal__header">
			<h3>Add Note — {noteModalMetric.label}</h3>
			<button onclick={() => showNoteModal = false} class="anl-note-modal__close" aria-label="Close">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M18 6L6 18M6 6l12 12"/>
				</svg>
			</button>
		</div>

		<!-- Metric context display -->
		<div class="anl-note-modal__context">
			<div class="anl-note-modal__metric">
				<span class="anl-note-modal__metric-label">Requests</span>
				<span class="anl-note-modal__metric-value">{noteModalMetric.requests.toLocaleString()}</span>
			</div>
			<div class="anl-note-modal__metric">
				<span class="anl-note-modal__metric-label">Tokens</span>
				<span class="anl-note-modal__metric-value">{formatNumber(noteModalMetric.tokens)}</span>
			</div>
			<div class="anl-note-modal__metric">
				<span class="anl-note-modal__metric-label">Cost</span>
				<span class="anl-note-modal__metric-value">{formatCost(noteModalMetric.cost)}</span>
			</div>
		</div>

		<textarea
			bind:value={noteContent}
			class="anl-note-modal__textarea"
			placeholder="Why did this metric change? What happened on this day?"
			rows="4"
		></textarea>

		<div class="anl-note-modal__actions">
			{#if existingNoteId}
				<button onclick={deleteNoteFromModal} class="anl-note-modal__delete">Delete Note</button>
			{/if}
			<div class="anl-note-modal__spacer"></div>
			<button onclick={() => showNoteModal = false} class="anl-note-modal__cancel">Cancel</button>
			<button onclick={saveNote} class="anl-note-modal__save" disabled={!noteContent.trim()}>
				{existingNoteId ? 'Update Note' : 'Save Note'}
			</button>
		</div>

		<p class="anl-note-modal__hint">This note will appear in your Daily Log for {noteModalMetric.label}</p>
	</div>
</div>
{/if}

<style>
	/* ── Layout ────────────────────────────────────────────────────────────── */
	.anl-page {
		padding: 24px;
		max-width: 1400px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 20px;
		background: var(--dbg2);
		min-height: 100%;
	}

	/* ── Header ────────────────────────────────────────────────────────────── */
	.anl-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}

	.anl-title {
		font-size: 22px;
		font-weight: 600;
		color: var(--dt);
		margin: 0 0 2px;
		letter-spacing: -0.02em;
	}

	.anl-subtitle {
		font-size: 13px;
		color: var(--dt3);
		margin: 0;
	}

	/* ── Period selector ───────────────────────────────────────────────────── */
	.anl-period-selector {
		display: flex;
		gap: 2px;
		background: var(--dbg3);
		border: 1px solid var(--dbd);
		border-radius: 10px;
		padding: 3px;
	}

	.anl-period-btn {
		font-size: 12px;
		font-weight: 500;
		color: var(--dt3);
		background: transparent;
		border: none;
		border-radius: 7px;
		padding: 5px 14px;
		cursor: pointer;
		transition: background 150ms ease-out, color 150ms ease-out;
		line-height: 1;
	}

	.anl-period-btn:hover {
		color: var(--dt);
		background: var(--dbg);
	}

	.anl-period-btn--active {
		background: var(--dbg);
		color: var(--dt);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	/* ── Card base ─────────────────────────────────────────────────────────── */
	.anl-card {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		overflow: hidden;
		transition: border-color 150ms ease-out;
	}

	.anl-card:hover {
		border-color: rgba(59, 130, 246, 0.2);
	}

	.anl-card-header {
		padding: 16px 20px 0;
	}

	.anl-section-title {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt2);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		display: block;
	}

	/* ── KPI Ribbon ────────────────────────────────────────────────────────── */
	.anl-kpi-ribbon {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
	}

	@media (max-width: 900px) {
		.anl-kpi-ribbon {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 560px) {
		.anl-kpi-ribbon {
			grid-template-columns: 1fr;
		}
	}

	.anl-kpi-card {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		padding: 18px 20px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		transition: border-color 150ms ease-out;
	}

	.anl-kpi-card:hover {
		border-color: rgba(59, 130, 246, 0.2);
	}

	.anl-kpi-left {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.anl-kpi-label {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt3);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		white-space: nowrap;
	}

	.anl-kpi-value {
		font-family: var(--bos-font-number-family);
		font-size: 28px;
		font-weight: 600;
		color: var(--dt);
		line-height: 1;
		letter-spacing: -0.02em;
	}

	.anl-delta {
		display: inline-block;
		font-size: 11px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 20px;
		font-family: var(--bos-font-number-family);
		width: fit-content;
	}

	.anl-delta--positive {
		color: var(--bos-status-success);
		background: var(--bos-status-success-bg);
	}

	.anl-delta--negative {
		color: var(--bos-status-error);
		background: var(--bos-status-error-bg);
	}

	.anl-sparkline {
		width: 60px;
		height: 24px;
		flex-shrink: 0;
	}

	/* ── Chart ─────────────────────────────────────────────────────────────── */
	.anl-chart-card {
		position: relative;
	}

	.anl-chart-wrap {
		padding: 12px 20px 20px;
		position: relative;
	}

	.anl-chart-svg {
		width: 100%;
		height: 200px;
		display: block;
		cursor: crosshair;
		overflow: visible;
	}

	.anl-axis-label {
		font-family: var(--bos-font-number-family);
		font-size: 10px;
		fill: var(--dt4);
	}

	/* ── Tooltip ───────────────────────────────────────────────────────────── */
	.anl-tooltip {
		position: absolute;
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 8px;
		padding: 10px 12px;
		pointer-events: none;
		display: flex;
		flex-direction: column;
		gap: 4px;
		box-shadow: var(--bos-shadow-2);
		z-index: 10;
		min-width: 130px;
	}

	.anl-tooltip-date {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt2);
		margin-bottom: 2px;
		display: block;
	}

	.anl-tooltip-row {
		display: flex;
		justify-content: space-between;
		gap: 16px;
	}

	.anl-tooltip-lbl {
		font-size: 12px;
		color: var(--dt3);
	}

	.anl-tooltip-val {
		font-family: var(--bos-font-number-family);
		font-size: 12px;
		font-weight: 600;
		color: var(--dt);
	}

	/* ── Two-column ────────────────────────────────────────────────────────── */
	.anl-two-col {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}

	@media (max-width: 780px) {
		.anl-two-col {
			grid-template-columns: 1fr;
		}
	}

	/* ── Provider breakdown ────────────────────────────────────────────────── */
	.anl-provider-list {
		padding: 14px 20px 20px;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}

	.anl-provider-row {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.anl-provider-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.anl-provider-name {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt);
	}

	.anl-provider-meta {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--dt3);
	}

	.anl-provider-sep {
		opacity: 0.5;
	}

	.anl-bar-track {
		height: 6px;
		background: var(--dbg3);
		border-radius: 4px;
		overflow: hidden;
	}

	.anl-bar-fill {
		height: 100%;
		background: var(--dt);
		border-radius: 4px;
		transition: width 400ms ease-out;
	}

	.anl-provider-pct {
		font-size: 11px;
		color: var(--dt4);
		font-family: var(--bos-font-number-family);
	}

	/* ── Top Models ────────────────────────────────────────────────────────── */
	.anl-model-list {
		padding: 10px 0 8px;
	}

	.anl-model-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 9px 20px;
		transition: background 120ms ease-out;
	}

	.anl-model-row:hover {
		background: var(--dbg2);
	}

	.anl-model-rank {
		font-family: var(--bos-font-number-family);
		font-size: 12px;
		color: var(--dt4);
		width: 16px;
		text-align: right;
		flex-shrink: 0;
	}

	.anl-model-info {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: 1;
		min-width: 0;
	}

	.anl-model-name {
		font-size: 13px;
		color: var(--dt);
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.anl-provider-badge {
		font-size: 10px;
		font-weight: 500;
		color: var(--dt3);
		background: var(--dbg3);
		border: 1px solid var(--dbd);
		border-radius: 20px;
		padding: 1px 7px;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.anl-model-stats {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 1px;
		flex-shrink: 0;
	}

	.anl-model-cost {
		font-size: 11px;
		color: var(--dt3);
	}

	/* ── Agent Table ───────────────────────────────────────────────────────── */
	.anl-agent-table {
		padding: 10px 0 8px;
	}

	.anl-agent-thead {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 2fr;
		padding: 0 20px 6px;
		gap: 12px;
	}

	.anl-agent-thead [role="columnheader"] {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt3);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.anl-agent-row {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 2fr;
		padding: 9px 20px;
		gap: 12px;
		align-items: center;
		transition: background 120ms ease-out;
	}

	.anl-agent-row:hover {
		background: var(--dbg2);
	}

	.anl-agent-name {
		font-size: 13px;
		color: var(--dt);
		font-weight: 500;
	}

	.anl-col-volume {
		/* reserved column for volume bar */
	}

	.anl-agent-bar-cell {
		display: flex;
		align-items: center;
	}

	.anl-agent-bar-track {
		flex: 1;
		height: 4px;
		background: var(--dbg3);
		border-radius: 4px;
		overflow: hidden;
	}

	.anl-agent-bar-fill {
		height: 100%;
		background: var(--bos-accent-blue);
		border-radius: 4px;
		opacity: 0.7;
		transition: width 400ms ease-out;
	}

	/* ── Shared number style ───────────────────────────────────────────────── */
	.anl-num {
		font-family: var(--bos-font-number-family);
		font-size: 13px;
		color: var(--dt);
	}

	/* ── Data source badges ─────────────────────────────────────────────────── */
	.anl-demo-badge {
		display: inline-block;
		font-size: 10px;
		font-weight: 600;
		color: var(--dt3);
		background: var(--dbg3);
		border: 1px solid var(--dbd);
		border-radius: 20px;
		padding: 1px 8px;
		margin-left: 8px;
		vertical-align: middle;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.anl-loading-badge {
		display: inline-block;
		font-size: 10px;
		font-weight: 500;
		color: var(--dt4);
		margin-left: 8px;
		vertical-align: middle;
		letter-spacing: 0.02em;
	}

	/* ── Tooltip click hint ─────────────────────────────────────────────────── */
	.anl-tooltip-click-hint {
		display: block;
		font-size: 10px;
		color: var(--dt4);
		margin-top: 4px;
		text-align: center;
		letter-spacing: 0.01em;
	}

	/* ── Note indicator on chart ────────────────────────────────────────────── */
	.anl-note-indicator {
		cursor: pointer;
	}

	/* ── Note Modal Overlay ─────────────────────────────────────────────────── */
	.anl-note-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.anl-note-modal {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		padding: 1.25rem;
		width: 100%;
		max-width: 28rem;
		box-shadow: 0 20px 40px rgba(0,0,0,0.3);
	}

	.anl-note-modal__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.anl-note-modal__header h3 {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.anl-note-modal__close {
		background: none;
		border: none;
		color: var(--dt3);
		cursor: pointer;
		padding: 4px;
		border-radius: 6px;
		transition: all 0.15s;
	}

	.anl-note-modal__close:hover {
		background: var(--dbg3);
		color: var(--dt);
	}

	.anl-note-modal__context {
		display: flex;
		gap: 0.75rem;
		margin-bottom: 1rem;
		padding: 0.65rem 0.85rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
	}

	.anl-note-modal__metric {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		flex: 1;
	}

	.anl-note-modal__metric-label {
		font-size: 0.68rem;
		color: var(--dt3);
		font-weight: 500;
	}

	.anl-note-modal__metric-value {
		font-size: 0.85rem;
		color: var(--dt);
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}

	.anl-note-modal__textarea {
		width: 100%;
		resize: none;
		padding: 0.65rem 0.75rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		color: var(--dt);
		font-family: inherit;
		font-size: 0.82rem;
		line-height: 1.5;
		outline: none;
		transition: border-color 0.15s;
		margin-bottom: 0.85rem;
		box-sizing: border-box;
	}

	.anl-note-modal__textarea:focus {
		border-color: var(--dt3);
	}

	.anl-note-modal__textarea::placeholder {
		color: var(--dt4);
	}

	.anl-note-modal__actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.anl-note-modal__spacer {
		flex: 1;
	}

	.anl-note-modal__cancel {
		padding: 0.4rem 0.85rem;
		background: none;
		border: 1px solid var(--dbd);
		border-radius: 7px;
		color: var(--dt2);
		font-size: 0.78rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.anl-note-modal__cancel:hover {
		background: var(--dbg2);
	}

	.anl-note-modal__save {
		padding: 0.4rem 0.85rem;
		background: var(--dt);
		border: none;
		border-radius: 7px;
		color: var(--dbg);
		font-size: 0.78rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s;
	}

	.anl-note-modal__save:hover {
		opacity: 0.9;
	}

	.anl-note-modal__save:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.anl-note-modal__delete {
		padding: 0.4rem 0.85rem;
		background: none;
		border: 1px solid color-mix(in srgb, var(--bos-status-error) 40%, transparent);
		border-radius: 7px;
		color: var(--bos-status-error);
		font-size: 0.78rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.anl-note-modal__delete:hover {
		background: color-mix(in srgb, var(--bos-status-error) 10%, transparent);
	}

	.anl-note-modal__hint {
		font-size: 0.68rem;
		color: var(--dt4);
		margin: 0.65rem 0 0;
		text-align: center;
	}
</style>
