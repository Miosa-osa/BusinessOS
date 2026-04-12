<script lang="ts">
	import { onMount } from 'svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ── Types ─────────────────────────────────────────────────────────────────
	interface Section {
		title: string;
		content: string;
	}

	interface WeeklyData {
		content: string;
		date: string;
	}

	// ── State ─────────────────────────────────────────────────────────────────
	let rhythmContent   = $state<string | null>(null);
	let rhythmDate      = $state<string>('');
	let weeklyContent   = $state<string | null>(null);
	let loadingRhythm   = $state(true);
	let loadingWeekly   = $state(false);
	let rhythmError     = $state<string | null>(null);
	let weeklyError     = $state<string | null>(null);
	let activeTab       = $state<'today' | 'weekly'>('today');
	let nodeCount       = $state<number | null>(null);

	// ── Derived ───────────────────────────────────────────────────────────────
	const now        = new Date();
	const weekday    = now.toLocaleDateString(undefined, { weekday: 'long' });
	const dayMonth   = now.toLocaleDateString(undefined, { month: 'long', day: 'numeric' });
	const year       = now.getFullYear();

	const parsedSections = $derived(parseMarkdownSections(rhythmContent ?? ''));
	const weeklyTop3    = $derived(extractTop3(weeklyContent ?? ''));
	const currentMode   = $derived(extractCurrentMode(rhythmContent ?? ''));

	// ── Mode badge colors ─────────────────────────────────────────────────────
	const MODE_COLORS: Record<string, string> = {
		BUILD:      'rd-mode--build',
		OPERATE:    'rd-mode--operate',
		LEARN:      'rd-mode--learn',
		SYNTHESIZE: 'rd-mode--synthesize',
		EXTRACT:    'rd-mode--extract',
	};

	function modeClass(mode: string): string {
		return MODE_COLORS[mode.toUpperCase()] ?? 'rd-mode--default';
	}

	// ── Helpers ───────────────────────────────────────────────────────────────
	function parseMarkdownSections(md: string): Section[] {
		if (!md.trim()) return [];
		// Strip YAML frontmatter
		const stripped = md.replace(/^---[\s\S]*?---\n/, '');
		const lines    = stripped.split('\n');
		const sections: Section[] = [];
		let current: Section | null = null;

		for (const line of lines) {
			const h2 = line.match(/^## (.+)/);
			if (h2) {
				if (current) sections.push(current);
				current = { title: h2[1].trim(), content: '' };
			} else if (current) {
				current.content += line + '\n';
			}
		}
		if (current) sections.push(current);
		return sections;
	}

	function extractTop3(md: string): string[] {
		if (!md.trim()) return [];
		const lines = md.split('\n');
		const items: string[] = [];
		let inSection = false;
		for (const line of lines) {
			if (/^## Top 3/i.test(line)) { inSection = true; continue; }
			if (inSection && /^##/.test(line)) break;
			if (inSection) {
				const m = line.match(/^\d+\.\s+(.+)/);
				if (m) items.push(m[1].replace(/\*\*/g, '').trim());
			}
		}
		return items.slice(0, 3);
	}

	function parseScheduleTable(content: string): Array<Record<string, string>> {
		const rows: Array<Record<string, string>> = [];
		const lines = content.split('\n').filter(l => l.trim().startsWith('|'));
		if (lines.length < 2) return rows;
		const headers = lines[0].split('|').map(h => h.trim()).filter(Boolean);
		for (let i = 2; i < lines.length; i++) {
			const cells = lines[i].split('|').map(c => c.trim()).filter(Boolean);
			if (cells.length < headers.length) continue;
			const row: Record<string, string> = {};
			headers.forEach((h, idx) => { row[h] = cells[idx] ?? ''; });
			rows.push(row);
		}
		return rows;
	}

	function parseCallsTable(content: string): Array<Record<string, string>> {
		return parseScheduleTable(content);
	}

	function renderInlineMarkdown(text: string): string {
		return text
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/`(.+?)`/g, '<code>$1</code>')
			.replace(/\[([^\]]+)\]\([^)]+\)/g, '<span class="rd-link">$1</span>');
	}

	function getMonday(): string {
		const d = new Date(now);
		const day = d.getDay();
		const diff = (day === 0 ? -6 : 1 - day);
		d.setDate(d.getDate() + diff);
		return d.toISOString().split('T')[0];
	}

	/**
	 * Extract the dominant cognitive mode from the rhythm file.
	 * Checks the Boot section first (Mode: BUILD), then falls back to
	 * scanning the first mode badge found anywhere in the content.
	 */
	function extractCurrentMode(md: string): string | null {
		if (!md.trim()) return null;
		// Boot section key-value: "Mode: BUILD"
		const bootMode = md.match(/^-\s*Mode:\s*(\w+)/im);
		if (bootMode) return bootMode[1].toUpperCase();
		// Fallback: first schedule mode column value
		const scheduleMode = md.match(/\|\s*(BUILD|OPERATE|LEARN|SYNTHESIZE|EXTRACT)\s*\|/i);
		if (scheduleMode) return scheduleMode[1].toUpperCase();
		return null;
	}

	// ── Data fetching ─────────────────────────────────────────────────────────
	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) headers['X-CSRF-Token'] = csrf;
		return headers;
	}

	async function loadRhythm() {
		loadingRhythm = true;
		rhythmError   = null;
		try {
			const res = await fetch(`${getApiBaseUrl()}/optimal/rhythm/today`, {
				method: 'GET',
				headers: buildHeaders(),
				credentials: 'include',
				signal: AbortSignal.timeout(6000),
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data: { date: string; content: string } = await res.json();
			rhythmContent = data.content;
			rhythmDate    = data.date;
		} catch (e) {
			rhythmError   = e instanceof Error ? e.message : 'Failed to load today\'s rhythm';
			rhythmContent = null;
		} finally {
			loadingRhythm = false;
		}
	}

	async function loadWeekly() {
		if (weeklyContent !== null) return; // already loaded
		loadingWeekly = true;
		weeklyError   = null;
		const monday  = getMonday();
		try {
			const res = await fetch(`${getApiBaseUrl()}/optimal/rhythm/weekly?date=${monday}`, {
				method: 'GET',
				headers: buildHeaders(),
				credentials: 'include',
				signal: AbortSignal.timeout(6000),
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data: WeeklyData = await res.json();
			weeklyContent = data.content;
		} catch (e) {
			weeklyError   = e instanceof Error ? e.message : 'Failed to load weekly plan';
			weeklyContent = null;
		} finally {
			loadingWeekly = false;
		}
	}

	async function loadNodeCount() {
		try {
			const res = await fetch(`${getApiBaseUrl()}/optimal/nodes`, {
				method: 'GET',
				headers: buildHeaders(),
				credentials: 'include',
				signal: AbortSignal.timeout(6000),
			});
			if (!res.ok) return;
			const data: { nodes?: unknown[] } = await res.json();
			if (Array.isArray(data.nodes)) nodeCount = data.nodes.length;
		} catch {
			// Node count is optional — silently ignore
		}
	}

	function handleTabChange(tab: 'today' | 'weekly') {
		activeTab = tab;
		if (tab === 'weekly') loadWeekly();
	}

	// ── Lifecycle ─────────────────────────────────────────────────────────────
	onMount(() => {
		loadRhythm();
		loadNodeCount();
	});
</script>

<div class="rd">
	<!-- ── Header ── -->
	<div class="rd-header">
		<div class="rd-header__date">
			<span class="rd-header__weekday">{weekday}</span>
			<h1 class="rd-header__day">{dayMonth}<span class="rd-header__year">{year}</span></h1>
		</div>
		<div class="rd-header__tabs">
			<button
				class="rd-tab"
				class:rd-tab--active={activeTab === 'today'}
				onclick={() => handleTabChange('today')}
			>Today</button>
			<button
				class="rd-tab"
				class:rd-tab--active={activeTab === 'weekly'}
				onclick={() => handleTabChange('weekly')}
			>This Week</button>
		</div>
	</div>

	<!-- ── Quick Stats ── -->
	<div class="rd-stats">
		<div class="rd-stat">
			<span class="rd-stat__label">Date</span>
			<span class="rd-stat__val">{now.toISOString().split('T')[0]}</span>
		</div>
		{#if nodeCount !== null}
			<div class="rd-stat">
				<span class="rd-stat__label">Nodes</span>
				<span class="rd-stat__val">{nodeCount} tracked</span>
			</div>
		{/if}
		{#if currentMode}
			<div class="rd-stat">
				<span class="rd-stat__label">Mode</span>
				<span class="rd-mode {modeClass(currentMode)} rd-stat__mode">{currentMode}</span>
			</div>
		{:else if !loadingRhythm && rhythmContent}
			<div class="rd-stat">
				<span class="rd-stat__label">Mode</span>
				<span class="rd-stat__val rd-stat__val--dim">Not set</span>
			</div>
		{/if}
	</div>

	<!-- ── Today tab ── -->
	{#if activeTab === 'today'}
		{#if loadingRhythm}
			<div class="rd-loading">
				<div class="rd-spinner"></div>
				<span>Loading rhythm...</span>
			</div>
		{:else if rhythmError}
			<div class="rd-empty">
				<div class="rd-empty__icon">
					<svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<p class="rd-empty__title">No plan for today</p>
				<p class="rd-empty__sub">
					Create <code>rhythm/daily/{now.toISOString().split('T')[0]}.md</code> to get started
				</p>
			</div>
		{:else if rhythmContent && parsedSections.length > 0}
			<div class="rd-content">
				{#each parsedSections as section}
					<!-- Boot section -->
					{#if section.title === 'Boot'}
						<div class="rd-card rd-card--boot">
							<h2 class="rd-section-title">Boot</h2>
							<div class="rd-boot-grid">
								{#each section.content.split('\n').filter(l => l.startsWith('-')) as line}
									{@const clean = line.replace(/^-\s*/, '')}
									<div class="rd-boot-item">
										<span class="rd-boot-label">{clean.split(':')[0]}</span>
										<span class="rd-boot-val">{clean.split(':').slice(1).join(':').trim() || '—'}</span>
									</div>
								{/each}
							</div>
						</div>

					<!-- Top 3 section -->
					{:else if section.title === 'Top 3 Today'}
						<div class="rd-card">
							<h2 class="rd-section-title">Top 3 Today</h2>
							<ol class="rd-top3">
								{#each section.content.split('\n').filter(l => /^\d+\./.test(l.trim())) as item, i}
									<li class="rd-top3-item">
										<span class="rd-top3-num">{i + 1}</span>
										<!-- eslint-disable-next-line svelte/no-at-html-tags -->
										<span>{@html renderInlineMarkdown(item.replace(/^\d+\.\s*/, ''))}</span>
									</li>
								{/each}
							</ol>
						</div>

					<!-- Schedule section -->
					{:else if section.title === 'Schedule'}
						{@const rows = parseScheduleTable(section.content)}
						{#if rows.length > 0}
							<div class="rd-card">
								<h2 class="rd-section-title">Schedule</h2>
								<div class="rd-schedule">
									{#each rows as row}
										{@const mode = (row['Mode'] ?? '').replace(/—/g, '').trim()}
										<div class="rd-schedule-row">
											<div class="rd-schedule-block">
												<span class="rd-schedule-name">{row['Block'] ?? ''}</span>
												<span class="rd-schedule-time">{row['Time'] ?? ''}</span>
											</div>
											{#if mode}
												<span class="rd-mode {modeClass(mode)}">{mode}</span>
											{:else}
												<span class="rd-mode rd-mode--break">—</span>
											{/if}
											<div class="rd-schedule-focus">
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												{@html renderInlineMarkdown(row['Focus'] ?? '')}
											</div>
											<div class="rd-schedule-done">
												{#if (row['Done?'] ?? '') === '[ ]'}
													<span class="rd-check rd-check--open"></span>
												{:else if (row['Done?'] ?? '') === '[x]' || (row['Done?'] ?? '') === '[X]'}
													<span class="rd-check rd-check--done">
														<svg width="10" height="10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
														</svg>
													</span>
												{:else}
													<span class="rd-check rd-check--na">—</span>
												{/if}
											</div>
										</div>
									{/each}
								</div>
							</div>
						{/if}

					<!-- Calls section -->
					{:else if section.title === 'Calls Today'}
						{@const rows = parseCallsTable(section.content)}
						{@const hasRealRows = rows.some(r => Object.values(r).some(v => v.trim()))}
						{#if hasRealRows}
							<div class="rd-card">
								<h2 class="rd-section-title">Calls Today</h2>
								<div class="rd-calls">
									{#each rows as row}
										{#if Object.values(row).some(v => v.trim())}
											<div class="rd-call-row">
												<span class="rd-call-time">{row['Time'] ?? ''}</span>
												<span class="rd-call-with">{row['With'] ?? ''}</span>
												<span class="rd-call-audience">{row['Audience Type'] ?? ''}</span>
											</div>
										{/if}
									{/each}
								</div>
							</div>
						{/if}

					<!-- Running Notes section -->
					{:else if section.title === 'Running Notes'}
						<div class="rd-card">
							<h2 class="rd-section-title">Running Notes</h2>
							<div class="rd-notes">
								{#each section.content.split('\n').filter(l => l.startsWith('-')) as line}
									{@const text = line.replace(/^-\s*/, '').trim()}
									{#if text}
										<div class="rd-note-item">
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html renderInlineMarkdown(text)}
										</div>
									{/if}
								{/each}
								{#if section.content.split('\n').filter(l => l.startsWith('-') && l.replace(/^-\s*/, '').trim()).length === 0}
									<p class="rd-notes-empty">No notes yet — capture anything during the day here.</p>
								{/if}
							</div>
						</div>

					<!-- End of Day section -->
					{:else if section.title === 'End of Day'}
						<div class="rd-card">
							<h2 class="rd-section-title">End of Day</h2>
							<div class="rd-eod">
								{#each section.content.split('\n').filter(l => l.trim().startsWith('-')) as line}
									{@const text = line.replace(/^-\s*/, '').trim()}
									{#if text}
										<div class="rd-eod-item">
											{#if text.startsWith('[ ]')}
												<span class="rd-check rd-check--open"></span>
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												<span>{@html renderInlineMarkdown(text.replace('[ ]', '').trim())}</span>
											{:else if text.startsWith('[x]') || text.startsWith('[X]')}
												<span class="rd-check rd-check--done">
													<svg width="10" height="10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
													</svg>
												</span>
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												<span>{@html renderInlineMarkdown(text.replace(/\[[xX]\]/, '').trim())}</span>
											{:else}
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												<span>{@html renderInlineMarkdown(text)}</span>
											{/if}
										</div>
									{/if}
								{/each}
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{:else}
			<div class="rd-empty">
				<div class="rd-empty__icon">
					<svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
				</div>
				<p class="rd-empty__title">No plan for today</p>
				<p class="rd-empty__sub">
					Create <code>rhythm/daily/{now.toISOString().split('T')[0]}.md</code> from the daily template to get started.
				</p>
			</div>
		{/if}
	{/if}

	<!-- ── Weekly tab ── -->
	{#if activeTab === 'weekly'}
		{#if loadingWeekly}
			<div class="rd-loading">
				<div class="rd-spinner"></div>
				<span>Loading weekly plan...</span>
			</div>
		{:else if weeklyError}
			<div class="rd-empty">
				<div class="rd-empty__icon">
					<svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<p class="rd-empty__title">No weekly plan found</p>
				<p class="rd-empty__sub">
					Create <code>rhythm/weekly/week-of-{getMonday()}.md</code> to see this week's plan.
				</p>
			</div>
		{:else if weeklyContent}
			{@const weeklySections = parseMarkdownSections(weeklyContent)}
			<div class="rd-content">
				{#if weeklyTop3.length > 0}
					<div class="rd-card rd-card--highlight">
						<h2 class="rd-section-title">Top 3 Non-Negotiables</h2>
						<ol class="rd-top3">
							{#each weeklyTop3 as item, i}
								<li class="rd-top3-item">
									<span class="rd-top3-num">{i + 1}</span>
									<!-- eslint-disable-next-line svelte/no-at-html-tags -->
									<span>{@html renderInlineMarkdown(item)}</span>
								</li>
							{/each}
						</ol>
					</div>
				{/if}

				{#each weeklySections.filter(s => s.title !== 'Top 3 Non-Negotiables') as section}
					<div class="rd-card">
						<h2 class="rd-section-title">{section.title}</h2>
						{#if section.content.includes('|')}
							{@const rows = parseScheduleTable(section.content)}
							{#if rows.length > 0}
								<div class="rd-table-wrap">
									<table class="rd-table">
										<thead>
											<tr>
												{#each Object.keys(rows[0]) as h}
													<th>{h}</th>
												{/each}
											</tr>
										</thead>
										<tbody>
											{#each rows as row}
												<tr>
													{#each Object.values(row) as cell}
														<!-- eslint-disable-next-line svelte/no-at-html-tags -->
														<td>{@html renderInlineMarkdown(cell)}</td>
													{/each}
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							{:else}
								<div class="rd-prose">{section.content.trim() || '—'}</div>
							{/if}
						{:else}
							<div class="rd-prose">
								{#each section.content.split('\n') as line}
									{#if line.trim().startsWith('-') || line.trim().startsWith('*')}
										<div class="rd-prose-bullet">
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html renderInlineMarkdown(line.replace(/^[\-\*]\s*/, ''))}
										</div>
									{:else if line.trim().startsWith('#')}
										<h3 class="rd-prose-subhead">{line.replace(/^#+\s*/, '')}</h3>
									{:else if line.trim()}
										<!-- eslint-disable-next-line svelte/no-at-html-tags -->
										<p class="rd-prose-p">{@html renderInlineMarkdown(line)}</p>
									{/if}
								{/each}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<style>
/* ── Root ── */
.rd {
	padding: 2rem 2.25rem;
	min-height: 100%;
	max-width: 860px;
}

/* ── Header ── */
.rd-header {
	display: flex;
	align-items: flex-end;
	justify-content: space-between;
	margin-bottom: 2rem;
	flex-wrap: wrap;
	gap: 1rem;
}
.rd-header__date {
	display: flex;
	flex-direction: column;
	gap: 0.15rem;
}
.rd-header__weekday {
	font-size: 0.75rem;
	font-weight: 500;
	letter-spacing: 0.08em;
	text-transform: uppercase;
	color: var(--dt3);
}
.rd-header__day {
	font-size: 1.75rem;
	font-weight: 700;
	line-height: 1.1;
	color: var(--dt);
	margin: 0;
}
.rd-header__year {
	font-size: 1rem;
	font-weight: 400;
	color: var(--dt3);
	margin-left: 0.4rem;
}

/* ── Tabs ── */
.rd-header__tabs {
	display: flex;
	gap: 0.25rem;
	background: var(--dbg2);
	border: 1px solid var(--dbd);
	border-radius: 8px;
	padding: 3px;
}
.rd-tab {
	padding: 0.35rem 1rem;
	border-radius: 6px;
	font-size: 0.8rem;
	font-weight: 500;
	color: var(--dt3);
	background: transparent;
	border: none;
	cursor: pointer;
	transition: background 0.15s, color 0.15s;
}
.rd-tab:hover { color: var(--dt2); }
.rd-tab--active {
	background: var(--dbg);
	color: var(--dt);
	box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}

/* ── Loading ── */
.rd-loading {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 3rem 0;
	color: var(--dt3);
	font-size: 0.875rem;
}
.rd-spinner {
	width: 18px;
	height: 18px;
	border: 2px solid var(--dbd);
	border-top-color: var(--dt2);
	border-radius: 50%;
	animation: rd-spin 0.7s linear infinite;
}
@keyframes rd-spin { to { transform: rotate(360deg); } }

/* ── Empty state ── */
.rd-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.75rem;
	padding: 5rem 2rem;
	text-align: center;
}
.rd-empty__icon {
	width: 52px;
	height: 52px;
	border-radius: 50%;
	background: var(--dbg2);
	border: 1px solid var(--dbd);
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--dt3);
}
.rd-empty__title {
	font-size: 1rem;
	font-weight: 600;
	color: var(--dt);
	margin: 0;
}
.rd-empty__sub {
	font-size: 0.825rem;
	color: var(--dt3);
	margin: 0;
}
.rd-empty__sub code {
	font-family: monospace;
	background: var(--dbg2);
	padding: 0.1rem 0.35rem;
	border-radius: 4px;
	border: 1px solid var(--dbd);
}

/* ── Content ── */
.rd-content {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

/* ── Card ── */
.rd-card {
	background: var(--dbg2);
	border: 1px solid var(--dbd);
	border-radius: 10px;
	padding: 1.25rem 1.5rem;
}
.rd-card--boot {
	border-left: 3px solid #6366f1;
}
.rd-card--highlight {
	border-left: 3px solid #f59e0b;
}
.rd-section-title {
	font-size: 0.7rem;
	font-weight: 600;
	letter-spacing: 0.1em;
	text-transform: uppercase;
	color: var(--dt3);
	margin: 0 0 0.9rem 0;
}

/* ── Boot grid ── */
.rd-boot-grid {
	display: flex;
	flex-wrap: wrap;
	gap: 0.75rem;
}
.rd-boot-item {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
	background: var(--dbg);
	border: 1px solid var(--dbd);
	border-radius: 8px;
	padding: 0.6rem 0.9rem;
	min-width: 130px;
}
.rd-boot-label {
	font-size: 0.7rem;
	text-transform: uppercase;
	letter-spacing: 0.06em;
	color: var(--dt4, var(--dt3));
}
.rd-boot-val {
	font-size: 0.875rem;
	font-weight: 600;
	color: var(--dt);
}

/* ── Top 3 ── */
.rd-top3 {
	list-style: none;
	padding: 0;
	margin: 0;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
.rd-top3-item {
	display: flex;
	align-items: baseline;
	gap: 0.75rem;
	font-size: 0.875rem;
	color: var(--dt2);
	line-height: 1.4;
}
.rd-top3-num {
	width: 22px;
	height: 22px;
	border-radius: 50%;
	background: var(--dbg);
	border: 1px solid var(--dbd);
	font-size: 0.7rem;
	font-weight: 700;
	color: var(--dt3);
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

/* ── Mode badges ── */
.rd-mode {
	display: inline-flex;
	align-items: center;
	padding: 0.2rem 0.55rem;
	border-radius: 99px;
	font-size: 0.65rem;
	font-weight: 700;
	letter-spacing: 0.05em;
	text-transform: uppercase;
	white-space: nowrap;
}
.rd-mode--build      { background: rgba(59,130,246,0.15); color: #60a5fa; }
.rd-mode--operate    { background: rgba(34,197,94,0.15);  color: #4ade80; }
.rd-mode--learn      { background: rgba(168,85,247,0.15); color: #c084fc; }
.rd-mode--synthesize { background: rgba(234,179,8,0.15);  color: #facc15; }
.rd-mode--extract    { background: rgba(249,115,22,0.15); color: #fb923c; }
.rd-mode--default    { background: var(--dbg); color: var(--dt3); }
.rd-mode--break      { color: var(--dt4, var(--dt3)); background: transparent; padding: 0; }

/* ── Schedule ── */
.rd-schedule {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}
.rd-schedule-row {
	display: grid;
	grid-template-columns: 180px 90px 1fr 28px;
	align-items: center;
	gap: 0.75rem;
	padding: 0.55rem 0.75rem;
	border-radius: 7px;
	background: var(--dbg);
	border: 1px solid var(--dbd);
}
.rd-schedule-block {
	display: flex;
	flex-direction: column;
	gap: 0.1rem;
}
.rd-schedule-name {
	font-size: 0.8rem;
	font-weight: 600;
	color: var(--dt);
}
.rd-schedule-time {
	font-size: 0.7rem;
	color: var(--dt3);
}
.rd-schedule-focus {
	font-size: 0.8rem;
	color: var(--dt2);
}

/* ── Check ── */
.rd-check {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 16px;
	height: 16px;
	border-radius: 4px;
	flex-shrink: 0;
}
.rd-check--open {
	border: 1.5px solid var(--dbd2, var(--dbd));
}
.rd-check--done {
	background: #22c55e;
	border: none;
	color: white;
}
.rd-check--na {
	color: var(--dt4, var(--dt3));
	font-size: 0.75rem;
	border: none;
}

/* ── Calls ── */
.rd-calls {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}
.rd-call-row {
	display: grid;
	grid-template-columns: 90px 1fr 1fr;
	gap: 0.75rem;
	padding: 0.5rem 0.75rem;
	background: var(--dbg);
	border: 1px solid var(--dbd);
	border-radius: 7px;
	font-size: 0.8rem;
}
.rd-call-time  { color: var(--dt3); }
.rd-call-with  { font-weight: 600; color: var(--dt); }
.rd-call-audience { color: var(--dt2); }

/* ── Notes ── */
.rd-notes {
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
}
.rd-note-item {
	font-size: 0.85rem;
	color: var(--dt2);
	padding: 0.4rem 0.6rem;
	background: var(--dbg);
	border-radius: 6px;
	border-left: 2px solid var(--dbd);
	line-height: 1.5;
}
.rd-notes-empty {
	font-size: 0.8rem;
	color: var(--dt3);
	font-style: italic;
	margin: 0;
}

/* ── End of Day ── */
.rd-eod {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
.rd-eod-item {
	display: flex;
	align-items: center;
	gap: 0.6rem;
	font-size: 0.85rem;
	color: var(--dt2);
}

/* ── Weekly prose ── */
.rd-prose {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}
.rd-prose-bullet {
	display: flex;
	align-items: baseline;
	gap: 0.5rem;
	font-size: 0.85rem;
	color: var(--dt2);
	padding-left: 0.75rem;
	position: relative;
}
.rd-prose-bullet::before {
	content: '–';
	position: absolute;
	left: 0;
	color: var(--dt3);
}
.rd-prose-p {
	font-size: 0.85rem;
	color: var(--dt2);
	margin: 0;
	line-height: 1.5;
}
.rd-prose-subhead {
	font-size: 0.8rem;
	font-weight: 600;
	color: var(--dt);
	margin: 0.5rem 0 0.25rem;
}

/* ── Table ── */
.rd-table-wrap {
	overflow-x: auto;
}
.rd-table {
	width: 100%;
	border-collapse: collapse;
	font-size: 0.8rem;
}
.rd-table th {
	text-align: left;
	padding: 0.4rem 0.75rem;
	background: var(--dbg);
	color: var(--dt3);
	font-weight: 600;
	font-size: 0.7rem;
	text-transform: uppercase;
	letter-spacing: 0.06em;
	border-bottom: 1px solid var(--dbd);
}
.rd-table td {
	padding: 0.45rem 0.75rem;
	color: var(--dt2);
	border-bottom: 1px solid var(--dbd);
	line-height: 1.4;
}
.rd-table tr:last-child td { border-bottom: none; }
.rd-table tr:hover td { background: var(--dbg); }

/* ── Quick Stats ── */
.rd-stats {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-wrap: wrap;
	margin-bottom: 1.5rem;
}
.rd-stat {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	background: var(--dbg2);
	border: 1px solid var(--dbd);
	border-radius: 7px;
	padding: 0.3rem 0.75rem;
}
.rd-stat__label {
	font-size: 0.68rem;
	font-weight: 600;
	letter-spacing: 0.07em;
	text-transform: uppercase;
	color: var(--dt3);
}
.rd-stat__val {
	font-size: 0.8rem;
	font-weight: 500;
	color: var(--dt);
}
.rd-stat__val--dim {
	color: var(--dt3);
	font-style: italic;
}
.rd-stat__mode {
	/* inherits .rd-mode sizing; slight override for compact fit */
	font-size: 0.62rem;
	padding: 0.15rem 0.45rem;
}

/* ── Inline code / link ── */
:global(.rd-card code) {
	font-family: monospace;
	font-size: 0.8em;
	background: var(--dbg);
	padding: 0.1em 0.35em;
	border-radius: 4px;
	border: 1px solid var(--dbd);
}
:global(.rd-link) {
	color: #60a5fa;
	text-decoration: underline;
	text-underline-offset: 2px;
}
</style>
