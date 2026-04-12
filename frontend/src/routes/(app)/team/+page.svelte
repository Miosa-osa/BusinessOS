<script lang="ts">
	import { onMount } from 'svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ─── Types ───────────────────────────────────────────────────────────────────

	interface TeamMember {
		name: string;
		role: string;
		genre: string;
		channel: string;
		status: 'active' | 'exited' | 'new' | 'uncertain' | 'partner';
	}

	interface TeamBlock {
		label: string;
		members: TeamMember[];
	}

	interface OptimalMember {
		name: string;
		role: string;
		status: 'active' | 'inactive' | 'exited';
		channel?: string;
		nodes?: string[];
	}

	interface WeeklyPriority {
		text: string;
		done: boolean;
	}

	interface Blocker {
		what: string;
		who: string;
		status: string;
	}

	interface WaitingOn {
		person: string;
		task: string;
	}

	// ─── State ───────────────────────────────────────────────────────────────────

	let loading = $state(true);
	let error = $state<string | null>(null);

	// Parsed from context.md
	let teamBlocks = $state<TeamBlock[]>([]);

	// Parsed from signal.md
	let weekOf = $state('');
	let signalStatus = $state('');
	let priorities = $state<WeeklyPriority[]>([]);
	let blockers = $state<Blocker[]>([]);
	let waitingOn = $state<WaitingOn[]>([]);

	// OptimalOS /api/optimal/team section
	let optimalMembers = $state<OptimalMember[]>([]);
	let optimalLoading = $state(false);
	let optimalSource = $state<'endpoint' | 'markdown' | null>(null);

	// File tree
	let fileTree = $state<{ name: string; path: string; is_dir: boolean; children?: { name: string; path: string }[] }[]>([]);
	let expandedFolders = $state<Set<string>>(new Set());
	let selectedFile = $state<{ path: string; content: string } | null>(null);
	let loadingFile = $state(false);

	// ─── API helpers ─────────────────────────────────────────────────────────────

	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) headers['X-CSRF-Token'] = csrf;
		return headers;
	}

	async function fetchJson<T>(path: string): Promise<T> {
		const res = await fetch(`${getApiBaseUrl()}${path}`, {
			method: 'GET',
			headers: buildHeaders(),
			credentials: 'include',
			signal: AbortSignal.timeout(8000)
		});
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		return res.json();
	}

	// ─── Markdown parsers ────────────────────────────────────────────────────────

	/** Parse the "People Involved" table from context.md */
	function parsePeopleTable(md: string): TeamMember[] {
		const tableRegex = /### People Involved\n\n\|[^\n]+\|\n\|[-|: ]+\|\n((?:\|[^\n]+\|\n?)+)/;
		const match = md.match(tableRegex);
		if (!match) return [];
		const results: TeamMember[] = [];
		for (const line of match[1].trim().split('\n').filter(Boolean)) {
			const cols = line.split('|').map((c) => c.trim()).filter(Boolean);
			if (cols.length < 4) continue;
			const name = cols[0] ?? '';
			const role = cols[1] ?? '';
			const genre = cols[2] ?? '';
			const channel = cols[3] ?? '';
			let status: TeamMember['status'] = 'active';
			if (role.includes('EXITED')) status = 'exited';
			else if (role.includes('50/50 partner')) status = 'partner';
			else if (role.includes('uncertain')) status = 'uncertain';
			results.push({ name, role, genre, channel, status });
		}
		return results;
	}

	/** Parse signal.md weekly priorities */
	function parsePriorities(md: string): WeeklyPriority[] {
		const section = md.match(/## This Week's Priorities\n\n([\s\S]*?)(?:\n## |\n---)/);
		if (!section) return [];
		return section[1]
			.trim()
			.split('\n')
			.filter((l) => l.startsWith('- ['))
			.map((l) => ({
				done: l.startsWith('- [x]'),
				text: l.replace(/^- \[[ x]\] /, '').trim()
			}));
	}

	/** Parse signal.md blockers table */
	function parseBlockers(md: string): Blocker[] {
		const tableRegex = /## Blockers \/ Fires\n\n\|[^\n]+\|\n\|[-|: ]+\|\n((?:\|[^\n]+\|\n?)+)/;
		const match = md.match(tableRegex);
		if (!match) return [];
		const results: Blocker[] = [];
		for (const line of match[1].trim().split('\n').filter(Boolean)) {
			const cols = line.split('|').map((c) => c.trim()).filter(Boolean);
			if (cols.length >= 3) results.push({ what: cols[0], who: cols[1], status: cols[2] });
		}
		return results;
	}

	/** Parse signal.md waiting-on table */
	function parseWaitingOn(md: string): WaitingOn[] {
		const tableRegex = /## Waiting on Roberto\n\n\|[^\n]+\|\n\|[-|: ]+\|\n((?:\|[^\n]+\|\n?)+)/;
		const match = md.match(tableRegex);
		if (!match) return [];
		const results: WaitingOn[] = [];
		for (const line of match[1].trim().split('\n').filter(Boolean)) {
			const cols = line.split('|').map((c) => c.trim()).filter(Boolean);
			if (cols.length >= 2) results.push({ person: cols[0], task: cols[1] });
		}
		return results;
	}

	/** Parse week-of and status from signal.md */
	function parseSignalMeta(md: string): { weekOf: string; status: string } {
		const weekMatch = md.match(/\*\*Week of:\*\* ([^\n]+)/);
		const statusMatch = md.match(/\*\*Status:\*\* ([^\n]+)/);
		return {
			weekOf: weekMatch?.[1]?.trim() ?? '',
			status: statusMatch?.[1]?.trim() ?? ''
		};
	}

	/** Group people table into blocks by section heading */
	function buildTeamBlocks(people: TeamMember[]): TeamBlock[] {
		const core = people.filter(
			(p) => ['Roberto', 'Javaris', 'Pedro', 'Nejd', 'Tejas', 'Pedram', 'Abdul', 'Nick'].some(
				(n) => p.name.startsWith(n)
			)
		);
		const sales = people.filter(
			(p) => ['Ed Honour', 'Robert Potter', 'Adam', 'Len', 'Liam', 'Jared', 'Bennett'].some(
				(n) => p.name.startsWith(n)
			)
		);
		const extended = people.filter(
			(p) => !core.includes(p) && !sales.includes(p)
		);

		const blocks: TeamBlock[] = [];
		if (core.length) blocks.push({ label: 'Core Team', members: core });
		if (sales.length) blocks.push({ label: 'Sales & Community', members: sales });
		if (extended.length) blocks.push({ label: 'Extended Network', members: extended });
		return blocks;
	}

	// ─── OptimalOS /api/optimal/team ─────────────────────────────────────────────

	/** Parse "**Name** — Role" patterns from raw markdown (fallback when endpoint 404s) */
	function parseTeamFromMarkdown(content: string): OptimalMember[] {
		const people: OptimalMember[] = [];
		const lines = content.split('\n');
		for (const line of lines) {
			const match = line.match(/\*\*(.+?)\*\*\s*[—–-]\s*(.+)/);
			if (match) {
				const name = match[1].trim();
				const role = match[2].trim();
				let status: OptimalMember['status'] = 'active';
				if (role.toLowerCase().includes('exited')) status = 'exited';
				else if (role.toLowerCase().includes('inactive')) status = 'inactive';
				people.push({ name, role, status });
			}
		}
		return people;
	}

	async function fetchOptimalTeam() {
		optimalLoading = true;
		try {
			// Attempt the dedicated /optimal/team endpoint first
			const res = await fetch(`${getApiBaseUrl()}/optimal/team`, {
				method: 'GET',
				headers: buildHeaders(),
				credentials: 'include',
				signal: AbortSignal.timeout(6000)
			});

			if (res.ok) {
				const data = await res.json() as { members: OptimalMember[]; count: number };
				optimalMembers = data.members ?? [];
				optimalSource = 'endpoint';
				return;
			}

			// 404 or any error → fall back to parsing context.md via existing node endpoint
			if (res.status === 404 || res.status >= 400) {
				throw new Error(`HTTP ${res.status}`);
			}
		} catch {
			// Fallback: re-use already-loaded context_md if available,
			// otherwise fetch the node endpoint
			try {
				const nodeData = await fetchJson<{ context_md: string; signal_md: string }>('/optimal/nodes/10-team');
				optimalMembers = parseTeamFromMarkdown(nodeData.context_md);
				optimalSource = 'markdown';
			} catch {
				// Non-fatal — optimal section will just be empty
				optimalMembers = [];
				optimalSource = null;
			}
		} finally {
			optimalLoading = false;
		}
	}

	// ─── File tree helpers ────────────────────────────────────────────────────────

	function toggleFolder(path: string) {
		expandedFolders = new Set(
			expandedFolders.has(path)
				? [...expandedFolders].filter((p) => p !== path)
				: [...expandedFolders, path]
		);
	}

	async function openFile(filePath: string) {
		if (loadingFile) return;
		loadingFile = true;
		selectedFile = null;
		try {
			const slug = '10-team';
			let clean = filePath.startsWith(slug + '/') ? filePath.slice(slug.length + 1) : filePath;
			const encoded = clean.split('/').map(encodeURIComponent).join('/');
			const data = await fetchJson<{ content: string; path: string }>(`/optimal/nodes/${slug}/file/${encoded}`);
			selectedFile = { path: filePath, content: data.content };
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load file';
		} finally {
			loadingFile = false;
		}
	}

	// ─── Status + role helpers ────────────────────────────────────────────────────

	const STATUS_COLORS: Record<TeamMember['status'], string> = {
		active:    'tm-badge--active',
		partner:   'tm-badge--partner',
		new:       'tm-badge--new',
		uncertain: 'tm-badge--uncertain',
		exited:    'tm-badge--exited'
	};

	const STATUS_LABELS: Record<TeamMember['status'], string> = {
		active:    'Active',
		partner:   'Partner',
		new:       'New',
		uncertain: 'Uncertain',
		exited:    'Exited'
	};

	function initials(name: string): string {
		return name.split(' ').slice(0, 2).map((w) => w[0]).join('').toUpperCase();
	}

	// ─── Load ─────────────────────────────────────────────────────────────────────

	onMount(async () => {
		// Fire the OptimalOS team section fetch in parallel — non-blocking
		fetchOptimalTeam();

		try {
			const [nodeData, treeData] = await Promise.all([
				fetchJson<{ context_md: string; signal_md: string }>('/optimal/nodes/10-team'),
				fetchJson<{ files: typeof fileTree }>('/optimal/nodes/10-team/files')
			]);

			// Parse context.md
			const people = parsePeopleTable(nodeData.context_md);
			teamBlocks = buildTeamBlocks(people);

			// Parse signal.md
			const meta = parseSignalMeta(nodeData.signal_md);
			weekOf = meta.weekOf;
			signalStatus = meta.status;
			priorities = parsePriorities(nodeData.signal_md);
			blockers = parseBlockers(nodeData.signal_md);
			waitingOn = parseWaitingOn(nodeData.signal_md);

			// File tree — only subfolders (not root md files)
			fileTree = (treeData.files ?? []).filter((f) => f.is_dir);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load team data';
		} finally {
			loading = false;
		}
	});
</script>

<div class="tm-page">

	<!-- ── Header ───────────────────────────────────────────────────────────────── -->
	<div class="tm-header">
		<div>
			<h1 class="tm-header__title">Team</h1>
			<p class="tm-header__sub">Node 10-team · OptimalOS registry</p>
		</div>
		{#if weekOf}
			<div class="tm-header__meta">
				<span class="tm-meta-label">Week of</span>
				<span class="tm-meta-value">{weekOf}</span>
				<span class="tm-meta-status tm-meta-status--{signalStatus.toLowerCase()}">{signalStatus}</span>
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="tm-center">
			<div class="tm-spinner" aria-label="Loading"></div>
			<p class="tm-muted">Loading team data...</p>
		</div>

	{:else if error}
		<div class="tm-error" role="alert">
			<p>{error}</p>
		</div>

	{:else}

		<!-- ── Signal Banner ─────────────────────────────────────────────────────── -->
		{#if priorities.length || blockers.length || waitingOn.length}
			<div class="tm-signal-banner">

				{#if priorities.length}
					<div class="tm-signal-section">
						<h3 class="tm-signal-section__label">This Week</h3>
						<ul class="tm-priority-list">
							{#each priorities as p}
								<li class="tm-priority-item" class:tm-priority-item--done={p.done}>
									<span class="tm-priority-dot" class:tm-priority-dot--done={p.done}></span>
									{p.text}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if blockers.length}
					<div class="tm-signal-section">
						<h3 class="tm-signal-section__label">Blockers</h3>
						<ul class="tm-blocker-list">
							{#each blockers as b}
								<li class="tm-blocker-item">
									<span class="tm-blocker-what">{b.what}</span>
									<span class="tm-blocker-sep">·</span>
									<span class="tm-blocker-who">{b.who}</span>
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if waitingOn.length}
					<div class="tm-signal-section">
						<h3 class="tm-signal-section__label">Waiting on Roberto</h3>
						<ul class="tm-waiting-list">
							{#each waitingOn as w}
								<li class="tm-waiting-item">
									<span class="tm-waiting-person">{w.person}</span>
									<span class="tm-waiting-task">{w.task}</span>
								</li>
							{/each}
						</ul>
					</div>
				{/if}

			</div>
		{/if}

		<!-- ── Team Grid ─────────────────────────────────────────────────────────── -->
		<div class="tm-body">
			<div class="tm-grid-area">
				{#each teamBlocks as block}
					<section class="tm-section" aria-labelledby="block-{block.label}">
						<h2 class="tm-section__label" id="block-{block.label}">{block.label}</h2>
						<div class="tm-grid">
							{#each block.members as member}
								<div class="tm-card" class:tm-card--exited={member.status === 'exited'}>
									<div class="tm-card__avatar" aria-hidden="true">
										{initials(member.name)}
									</div>
									<div class="tm-card__info">
										<div class="tm-card__name">{member.name}</div>
										<div class="tm-card__role">{member.role}</div>
										{#if member.channel && member.status !== 'exited'}
											<div class="tm-card__channel">{member.channel}</div>
										{/if}
									</div>
									<div class="tm-card__badges">
										<span class="tm-badge {STATUS_COLORS[member.status]}">
											{STATUS_LABELS[member.status]}
										</span>
										{#if member.genre && member.status !== 'exited'}
											<span class="tm-badge tm-badge--genre">{member.genre.split(',')[0].trim()}</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</section>
				{/each}

				<!-- ── OptimalOS Roster ─────────────────────────────────────────────── -->
				{#if optimalLoading}
					<section class="tm-section tm-osa-section">
						<h2 class="tm-section__label tm-osa-heading">
							<svg viewBox="0 0 16 16" class="tm-osa-icon" aria-hidden="true"><circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.5"/><circle cx="8" cy="8" r="2" fill="currentColor"/></svg>
							OptimalOS Roster
						</h2>
						<div class="tm-osa-loading">
							<span class="tm-spinner tm-spinner--sm" aria-label="Loading OptimalOS roster"></span>
							<span class="tm-osa-loading-text">Fetching roster...</span>
						</div>
					</section>
				{:else if optimalMembers.length > 0}
					<section class="tm-section tm-osa-section" aria-labelledby="osa-roster-label">
						<h2 class="tm-section__label tm-osa-heading" id="osa-roster-label">
							<svg viewBox="0 0 16 16" class="tm-osa-icon" aria-hidden="true"><circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.5"/><circle cx="8" cy="8" r="2" fill="currentColor"/></svg>
							OptimalOS Roster
							{#if optimalSource === 'endpoint'}
								<span class="tm-osa-source-tag">live</span>
							{:else if optimalSource === 'markdown'}
								<span class="tm-osa-source-tag tm-osa-source-tag--md">markdown</span>
							{/if}
						</h2>
						<div class="tm-grid">
							{#each optimalMembers as member}
								<div class="tm-card tm-osa-card" class:tm-card--exited={member.status === 'exited'}>
									<div class="tm-card__avatar" aria-hidden="true">
										{initials(member.name)}
									</div>
									<div class="tm-card__info">
										<div class="tm-card__name">{member.name}</div>
										<div class="tm-card__role">{member.role}</div>
										{#if member.channel && member.status !== 'exited'}
											<div class="tm-card__channel">{member.channel}</div>
										{/if}
										{#if member.nodes && member.nodes.length > 0}
											<div class="tm-osa-nodes" aria-label="Associated nodes">
												{#each member.nodes as node}
													<span class="tm-osa-node-tag">{node}</span>
												{/each}
											</div>
										{/if}
									</div>
									<div class="tm-card__badges">
										{#if member.status === 'active'}
											<span class="tm-badge tm-badge--active">Active</span>
										{:else if member.status === 'exited'}
											<span class="tm-badge tm-badge--exited">Exited</span>
										{:else}
											<span class="tm-badge tm-badge--uncertain">Inactive</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</section>
				{/if}

			</div>

			<!-- ── File Sidebar ───────────────────────────────────────────────────── -->
			{#if fileTree.length}
				<aside class="tm-sidebar" aria-label="Team files">
					<h2 class="tm-sidebar__title">Files</h2>
					<nav class="tm-filetree">
						{#each fileTree as folder}
							<div class="tm-folder">
								<button
									class="tm-folder__toggle"
									onclick={() => toggleFolder(folder.path)}
									aria-expanded={expandedFolders.has(folder.path)}
								>
									<svg class="tm-folder__icon" class:tm-folder__icon--open={expandedFolders.has(folder.path)} viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
										<path d="M6 2l2 2h6a1 1 0 011 1v8a1 1 0 01-1 1H2a1 1 0 01-1-1V3a1 1 0 011-1h4z"/>
									</svg>
									{folder.name}
								</button>
								{#if expandedFolders.has(folder.path) && folder.children}
									<ul class="tm-folder__files">
										{#each folder.children as file}
											{#if !file.name.endsWith('.md') || true}
												<li>
													<button
														class="tm-file-btn"
														class:tm-file-btn--active={selectedFile?.path === file.path}
														onclick={() => openFile(file.path)}
													>
														<svg class="tm-file-icon" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
															<path d="M4 0h6l4 4v11a1 1 0 01-1 1H3a1 1 0 01-1-1V1a1 1 0 011-1zm6 1v3h3L10 1z"/>
														</svg>
														{file.name}
													</button>
												</li>
											{/if}
										{/each}
									</ul>
								{/if}
							</div>
						{/each}
					</nav>

					{#if loadingFile}
						<div class="tm-file-loading" aria-live="polite">Loading...</div>
					{/if}

					{#if selectedFile}
						<div class="tm-file-preview">
							<div class="tm-file-preview__header">
								<span class="tm-file-preview__path">{selectedFile.path.split('/').pop()}</span>
								<button
									class="tm-file-preview__close"
									onclick={() => (selectedFile = null)}
									aria-label="Close file preview"
								>
									<svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true" class="w-3 h-3">
										<path d="M3.293 3.293a1 1 0 011.414 0L8 6.586l3.293-3.293a1 1 0 111.414 1.414L9.414 8l3.293 3.293a1 1 0 01-1.414 1.414L8 9.414l-3.293 3.293a1 1 0 01-1.414-1.414L6.586 8 3.293 4.707a1 1 0 010-1.414z"/>
									</svg>
								</button>
							</div>
							<pre class="tm-file-preview__content">{selectedFile.content}</pre>
						</div>
					{/if}
				</aside>
			{/if}
		</div>

	{/if}
</div>

<style>
	/* ── Page shell ──────────────────────────────────────────────────────────── */
	.tm-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #0a0a0a;
		overflow: hidden;
	}

	/* ── Header ──────────────────────────────────────────────────────────────── */
	.tm-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.5rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
		flex-shrink: 0;
	}
	.tm-header__title {
		font-size: 1.375rem;
		font-weight: 600;
		color: var(--dt, #f1f1f1);
		margin: 0;
	}
	.tm-header__sub {
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.4));
		margin: 0.125rem 0 0;
		font-family: monospace;
	}
	.tm-header__meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.tm-meta-label {
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.4));
	}
	.tm-meta-value {
		font-size: 0.8125rem;
		color: var(--dt2, rgba(255,255,255,0.7));
		font-weight: 500;
	}
	.tm-meta-status {
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		padding: 0.1875rem 0.5rem;
		border-radius: 4px;
		background: rgba(34, 197, 94, 0.12);
		color: #4ade80;
		border: 1px solid rgba(34, 197, 94, 0.2);
	}

	/* ── Signal banner ───────────────────────────────────────────────────────── */
	.tm-signal-banner {
		display: flex;
		gap: 1.5rem;
		padding: 0.875rem 1.5rem;
		border-bottom: 1px solid rgba(255,255,255,0.06);
		background: rgba(255,255,255,0.015);
		flex-shrink: 0;
		overflow-x: auto;
	}
	.tm-signal-section {
		min-width: 180px;
	}
	.tm-signal-section__label {
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--dt3, rgba(255,255,255,0.4));
		margin: 0 0 0.5rem;
	}
	.tm-priority-list,
	.tm-blocker-list,
	.tm-waiting-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.3125rem;
	}
	.tm-priority-item {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.8125rem;
		color: var(--dt2, rgba(255,255,255,0.7));
	}
	.tm-priority-item--done {
		color: var(--dt3, rgba(255,255,255,0.4));
		text-decoration: line-through;
	}
	.tm-priority-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: rgba(251, 191, 36, 0.7);
		flex-shrink: 0;
	}
	.tm-priority-dot--done {
		background: rgba(74, 222, 128, 0.5);
	}
	.tm-blocker-item {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.8125rem;
		color: var(--dt2, rgba(255,255,255,0.7));
	}
	.tm-blocker-what { color: rgba(252, 165, 165, 0.9); }
	.tm-blocker-sep { color: var(--dt3, rgba(255,255,255,0.3)); }
	.tm-blocker-who { color: var(--dt3, rgba(255,255,255,0.5)); font-size: 0.75rem; }
	.tm-waiting-item {
		display: flex;
		flex-direction: column;
		gap: 0.0625rem;
	}
	.tm-waiting-person {
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt2, rgba(255,255,255,0.7));
	}
	.tm-waiting-task {
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.45));
	}

	/* ── Body layout ─────────────────────────────────────────────────────────── */
	.tm-body {
		display: flex;
		flex: 1;
		overflow: hidden;
	}
	.tm-grid-area {
		flex: 1;
		overflow-y: auto;
		padding: 1.25rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.75rem;
	}

	/* ── Section ─────────────────────────────────────────────────────────────── */
	.tm-section__label {
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--dt3, rgba(255,255,255,0.4));
		margin: 0 0 0.75rem;
	}

	/* ── Team card grid ──────────────────────────────────────────────────────── */
	.tm-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 0.625rem;
	}
	.tm-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		background: rgba(255,255,255,0.03);
		border: 1px solid rgba(255,255,255,0.06);
		border-radius: 8px;
		transition: background 0.15s, border-color 0.15s;
	}
	.tm-card:hover {
		background: rgba(255,255,255,0.05);
		border-color: rgba(255,255,255,0.1);
	}
	.tm-card--exited {
		opacity: 0.45;
	}
	.tm-card__avatar {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: rgba(255,255,255,0.08);
		border: 1px solid rgba(255,255,255,0.1);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--dt2, rgba(255,255,255,0.7));
		flex-shrink: 0;
		letter-spacing: 0.02em;
	}
	.tm-card__info {
		flex: 1;
		min-width: 0;
	}
	.tm-card__name {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--dt, #f1f1f1);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tm-card__role {
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.45));
		margin-top: 0.0625rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tm-card__channel {
		font-size: 0.6875rem;
		color: var(--dt3, rgba(255,255,255,0.35));
		margin-top: 0.125rem;
		font-family: monospace;
	}
	.tm-card__badges {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.25rem;
		flex-shrink: 0;
	}

	/* ── Badges ──────────────────────────────────────────────────────────────── */
	.tm-badge {
		font-size: 0.625rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		padding: 0.1875rem 0.4375rem;
		border-radius: 4px;
		white-space: nowrap;
	}
	.tm-badge--active {
		background: rgba(34, 197, 94, 0.1);
		color: #4ade80;
		border: 1px solid rgba(34, 197, 94, 0.2);
	}
	.tm-badge--partner {
		background: rgba(139, 92, 246, 0.1);
		color: #a78bfa;
		border: 1px solid rgba(139, 92, 246, 0.2);
	}
	.tm-badge--new {
		background: rgba(59, 130, 246, 0.1);
		color: #60a5fa;
		border: 1px solid rgba(59, 130, 246, 0.2);
	}
	.tm-badge--uncertain {
		background: rgba(251, 191, 36, 0.1);
		color: #fbbf24;
		border: 1px solid rgba(251, 191, 36, 0.2);
	}
	.tm-badge--exited {
		background: rgba(255,255,255,0.04);
		color: rgba(255,255,255,0.3);
		border: 1px solid rgba(255,255,255,0.08);
	}
	.tm-badge--genre {
		background: rgba(255,255,255,0.04);
		color: var(--dt3, rgba(255,255,255,0.4));
		border: 1px solid rgba(255,255,255,0.07);
		font-weight: 400;
		letter-spacing: 0;
		text-transform: none;
	}

	/* ── File sidebar ────────────────────────────────────────────────────────── */
	.tm-sidebar {
		width: 220px;
		border-left: 1px solid rgba(255,255,255,0.06);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		flex-shrink: 0;
	}
	.tm-sidebar__title {
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--dt3, rgba(255,255,255,0.4));
		padding: 0.875rem 1rem 0.5rem;
		margin: 0;
		border-bottom: 1px solid rgba(255,255,255,0.05);
	}
	.tm-filetree {
		padding: 0.5rem 0;
		overflow-y: auto;
		flex-shrink: 0;
	}
	.tm-folder {
		padding: 0 0.25rem;
	}
	.tm-folder__toggle {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.3125rem 0.75rem;
		font-size: 0.8125rem;
		color: var(--dt2, rgba(255,255,255,0.7));
		background: none;
		border: none;
		cursor: pointer;
		border-radius: 5px;
		text-align: left;
	}
	.tm-folder__toggle:hover {
		background: rgba(255,255,255,0.04);
	}
	.tm-folder__icon {
		width: 14px;
		height: 14px;
		color: rgba(251,191,36,0.6);
		flex-shrink: 0;
		transition: opacity 0.1s;
	}
	.tm-folder__icon--open {
		color: rgba(251,191,36,0.9);
	}
	.tm-folder__files {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.tm-file-btn {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.25rem 0.75rem 0.25rem 1.75rem;
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.45));
		background: none;
		border: none;
		cursor: pointer;
		border-radius: 4px;
		text-align: left;
	}
	.tm-file-btn:hover {
		background: rgba(255,255,255,0.04);
		color: var(--dt2, rgba(255,255,255,0.7));
	}
	.tm-file-btn--active {
		background: rgba(255,255,255,0.06);
		color: var(--dt, #f1f1f1);
	}
	.tm-file-icon {
		width: 11px;
		height: 11px;
		color: rgba(255,255,255,0.25);
		flex-shrink: 0;
	}
	.tm-file-loading {
		font-size: 0.75rem;
		color: var(--dt3, rgba(255,255,255,0.4));
		padding: 0.5rem 1rem;
	}
	.tm-file-preview {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		border-top: 1px solid rgba(255,255,255,0.06);
		margin-top: 0.5rem;
	}
	.tm-file-preview__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 1rem;
		border-bottom: 1px solid rgba(255,255,255,0.05);
	}
	.tm-file-preview__path {
		font-size: 0.75rem;
		color: var(--dt2, rgba(255,255,255,0.6));
		font-family: monospace;
	}
	.tm-file-preview__close {
		background: none;
		border: none;
		cursor: pointer;
		color: var(--dt3, rgba(255,255,255,0.4));
		padding: 0.125rem;
		border-radius: 3px;
		display: flex;
		align-items: center;
	}
	.tm-file-preview__close:hover { color: var(--dt, #f1f1f1); }
	.tm-file-preview__content {
		flex: 1;
		overflow: auto;
		padding: 0.75rem 1rem;
		font-size: 0.6875rem;
		font-family: monospace;
		color: var(--dt3, rgba(255,255,255,0.45));
		white-space: pre-wrap;
		word-break: break-word;
		margin: 0;
		line-height: 1.6;
	}

	/* ── Utility ─────────────────────────────────────────────────────────────── */
	.tm-center {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
	}
	.tm-muted {
		font-size: 0.875rem;
		color: var(--dt3, rgba(255,255,255,0.4));
		margin: 0;
	}
	.tm-spinner {
		width: 28px;
		height: 28px;
		border: 2px solid rgba(255,255,255,0.1);
		border-top-color: rgba(255,255,255,0.5);
		border-radius: 50%;
		animation: tm-spin 0.7s linear infinite;
	}
	@keyframes tm-spin {
		to { transform: rotate(360deg); }
	}
	.tm-error {
		margin: 1rem 1.5rem;
		padding: 0.875rem 1rem;
		background: rgba(239, 68, 68, 0.08);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: 8px;
		font-size: 0.875rem;
		color: #f87171;
	}
	.tm-error p { margin: 0; }

	/* ── OptimalOS Roster section ────────────────────────────────────────────── */
	.tm-osa-section {
		border-top: 1px solid rgba(255,255,255,0.06);
		padding-top: 1.25rem;
		margin-top: 0.5rem;
	}
	.tm-osa-heading {
		display: flex;
		align-items: center;
		gap: 0.375rem;
	}
	.tm-osa-icon {
		width: 12px;
		height: 12px;
		color: rgba(139, 92, 246, 0.7);
		flex-shrink: 0;
	}
	.tm-osa-source-tag {
		font-size: 0.5625rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 0.125rem 0.375rem;
		border-radius: 3px;
		background: rgba(34, 197, 94, 0.1);
		color: #4ade80;
		border: 1px solid rgba(34, 197, 94, 0.2);
		margin-left: 0.25rem;
	}
	.tm-osa-source-tag--md {
		background: rgba(251, 191, 36, 0.1);
		color: #fbbf24;
		border-color: rgba(251, 191, 36, 0.2);
	}
	.tm-osa-loading {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0;
	}
	.tm-osa-loading-text {
		font-size: 0.8125rem;
		color: var(--dt3, rgba(255,255,255,0.4));
	}
	.tm-spinner--sm {
		width: 16px;
		height: 16px;
		border-width: 1.5px;
	}
	.tm-osa-card {
		border-color: rgba(139, 92, 246, 0.1);
	}
	.tm-osa-card:hover {
		border-color: rgba(139, 92, 246, 0.2);
	}
	.tm-osa-nodes {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		margin-top: 0.3125rem;
	}
	.tm-osa-node-tag {
		font-size: 0.5625rem;
		font-weight: 500;
		letter-spacing: 0.03em;
		padding: 0.125rem 0.375rem;
		border-radius: 3px;
		background: rgba(139, 92, 246, 0.08);
		color: rgba(167, 139, 250, 0.8);
		border: 1px solid rgba(139, 92, 246, 0.15);
		white-space: nowrap;
	}
</style>
