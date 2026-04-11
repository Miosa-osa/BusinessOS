<script lang="ts">
	import { onMount } from 'svelte';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';
	import type { FileEntry, OptimalNode } from '$lib/stores/optimal';

	// ─── Types ───────────────────────────────────────────────────────────────────

	interface ProjectFolder {
		name: string;
		path: string;
		files: FileEntry[];
	}

	interface NodeProjects {
		node: OptimalNode;
		projects: ProjectFolder[];
	}

	// ─── State ────────────────────────────────────────────────────────────────────

	let loading = $state(true);
	let error = $state<string | null>(null);
	let nodeProjects = $state<NodeProjects[]>([]);

	// Side panel
	let panelOpen = $state(false);
	let panelLoading = $state(false);
	let panelContent = $state('');
	let panelTitle = $state('');
	let panelError = $state<string | null>(null);

	// ─── Node color map ──────────────────────────────────────────────────────────

	const NODE_COLORS: Record<string, string> = {
		'00-command':               '#6366f1',
		'01-roberto':               '#8b5cf6',
		'02-miosa':                 '#3b82f6',
		'03-lunivate':              '#06b6d4',
		'04-ai-masters':            '#f59e0b',
		'05-os-architect':          '#10b981',
		'06-agency-accelerants':    '#ef4444',
		'07-accelerants-community': '#f97316',
		'08-content-creators':      '#ec4899',
		'09-new-stuff':             '#64748b',
		'10-team':                  '#14b8a6',
		'11-money-revenue':         '#22c55e',
		'12-os-accelerator':        '#a855f7',
	};

	function nodeColor(slug: string): string {
		return NODE_COLORS[slug] ?? '#6366f1';
	}

	// ─── Helpers ─────────────────────────────────────────────────────────────────

	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) headers['X-CSRF-Token'] = csrf;
		return headers;
	}

	/** Walk the tree recursively, collect leaf files whose path contains "/projects/" */
	function collectProjectFiles(entries: FileEntry[], acc: FileEntry[] = []): FileEntry[] {
		for (const entry of entries) {
			if (!entry.is_dir && entry.path.includes('/projects/')) {
				acc.push(entry);
			}
			if (entry.is_dir && entry.children?.length) {
				collectProjectFiles(entry.children, acc);
			}
		}
		return acc;
	}

	/**
	 * From a flat list of project files, group into ProjectFolder objects
	 * keyed by the immediate parent directory (the project folder name).
	 */
	function groupIntoFolders(files: FileEntry[]): ProjectFolder[] {
		const map = new Map<string, ProjectFolder>();

		for (const file of files) {
			// e.g. "02-miosa/canopy/projects/canopy-launch/specs/REDESIGN-SPEC.md"
			// find the segment right after "projects/"
			const idx = file.path.indexOf('/projects/');
			if (idx === -1) continue;

			const afterProjects = file.path.slice(idx + '/projects/'.length); // "canopy-launch/specs/REDESIGN-SPEC.md"
			const projectName = afterProjects.split('/')[0];                   // "canopy-launch"

			// Reconstruct a canonical path up to the project folder
			const folderPath = file.path.slice(0, idx + '/projects/'.length) + projectName;

			if (!map.has(folderPath)) {
				map.set(folderPath, { name: projectName, path: folderPath, files: [] });
			}
			map.get(folderPath)!.files.push(file);
		}

		return Array.from(map.values()).sort((a, b) => a.name.localeCompare(b.name));
	}

	// ─── Fetch ────────────────────────────────────────────────────────────────────

	async function fetchFileTree(slug: string): Promise<FileEntry[]> {
		const res = await fetch(
			`${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/files`,
			{ method: 'GET', headers: buildHeaders(), credentials: 'include', signal: AbortSignal.timeout(10000) }
		);
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		const data: { files?: FileEntry[] } = await res.json();
		return Array.isArray(data.files) ? data.files : [];
	}

	async function loadAllProjects(): Promise<void> {
		loading = true;
		error = null;

		try {
			// 1. Fetch all nodes
			const res = await fetch(
				`${getApiBaseUrl()}/optimal/nodes`,
				{ method: 'GET', headers: buildHeaders(), credentials: 'include', signal: AbortSignal.timeout(8000) }
			);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data: { nodes?: OptimalNode[] } = await res.json();
			const nodes: OptimalNode[] = Array.isArray(data.nodes) ? data.nodes : [];

			// 2. Fetch file trees for all nodes in parallel
			const results = await Promise.allSettled(
				nodes.map(async (node) => {
					const tree = await fetchFileTree(node.slug);
					const files = collectProjectFiles(tree);
					if (files.length === 0) return null;
					const projects = groupIntoFolders(files);
					return { node, projects } satisfies NodeProjects;
				})
			);

			// 3. Collect successful results with at least one project
			nodeProjects = results
				.filter((r): r is PromiseFulfilledResult<NodeProjects | null> => r.status === 'fulfilled' && r.value !== null)
				.map((r) => r.value as NodeProjects)
				.sort((a, b) => a.node.slug.localeCompare(b.node.slug));

		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load projects';
		} finally {
			loading = false;
		}
	}

	async function openFile(slug: string, filePath: string): Promise<void> {
		panelOpen = true;
		panelLoading = true;
		panelContent = '';
		panelError = null;
		panelTitle = filePath.split('/').pop() ?? filePath;

		try {
			let cleanPath = filePath;
			if (cleanPath.startsWith(slug + '/')) cleanPath = cleanPath.slice(slug.length + 1);
			const encoded = cleanPath.split('/').map(encodeURIComponent).join('/');

			const res = await fetch(
				`${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/file/${encoded}`,
				{ method: 'GET', headers: buildHeaders(), credentials: 'include', signal: AbortSignal.timeout(8000) }
			);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data: { content?: string } = await res.json();
			panelContent = data.content ?? '';
		} catch (err) {
			panelError = err instanceof Error ? err.message : 'Failed to load file';
		} finally {
			panelLoading = false;
		}
	}

	function closePanel() {
		panelOpen = false;
		panelContent = '';
		panelTitle = '';
		panelError = null;
	}

	// ─── Lifecycle ────────────────────────────────────────────────────────────────

	onMount(() => { loadAllProjects(); });

	// ─── Derived counts ───────────────────────────────────────────────────────────

	let totalProjects = $derived(nodeProjects.reduce((n, np) => n + np.projects.length, 0));
	let totalFiles = $derived(
		nodeProjects.reduce((n, np) => n + np.projects.reduce((m, p) => m + p.files.length, 0), 0)
	);
</script>

<!-- ─── Layout ─────────────────────────────────────────────────────────────── -->
<div class="pj-root" class:pj-root--panel={panelOpen}>

	<!-- Main column -->
	<div class="pj-main">
		<!-- Header -->
		<header class="pj-header">
			<div class="pj-header__left">
				<h1 class="pj-header__title">Projects</h1>
				{#if !loading}
					<span class="pj-header__meta">
						{totalProjects} projects &middot; {totalFiles} files &middot; {nodeProjects.length} nodes
					</span>
				{/if}
			</div>
			<button class="pj-btn pj-btn--ghost" onclick={loadAllProjects} disabled={loading} aria-label="Refresh projects">
				<svg class="pj-icon" class:pj-spin={loading} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
				</svg>
			</button>
		</header>

		<!-- Error banner -->
		{#if error}
			<div class="pj-error" role="alert">
				<svg class="pj-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
				</svg>
				{error}
				<button class="pj-error__retry" onclick={loadAllProjects}>Retry</button>
			</div>
		{/if}

		<!-- Loading skeleton -->
		{#if loading}
			<div class="pj-skeleton-grid">
				{#each Array(6) as _}
					<div class="pj-skeleton-card">
						<div class="pj-skeleton-bar pj-skeleton-bar--header"></div>
						<div class="pj-skeleton-bar pj-skeleton-bar--title"></div>
						<div class="pj-skeleton-bar pj-skeleton-bar--line"></div>
						<div class="pj-skeleton-bar pj-skeleton-bar--line pj-skeleton-bar--short"></div>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Content -->
		{#if !loading && nodeProjects.length > 0}
			{#each nodeProjects as { node, projects } (node.slug)}
				<section class="pj-section">
					<!-- Node header -->
					<div class="pj-section__header" style="--node-color: {nodeColor(node.slug)}">
						<div class="pj-section__dot"></div>
						<h2 class="pj-section__name">{node.name}</h2>
						<span class="pj-section__count">{projects.length} project{projects.length !== 1 ? 's' : ''}</span>
					</div>

					<!-- Project cards -->
					<div class="pj-card-grid">
						{#each projects as project (project.path)}
							<div class="pj-card" style="--node-color: {nodeColor(node.slug)}">
								<!-- Card header -->
								<div class="pj-card__header">
									<svg class="pj-card__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
									</svg>
									<span class="pj-card__name">{project.name}</span>
									<span class="pj-card__file-count">{project.files.length} file{project.files.length !== 1 ? 's' : ''}</span>
								</div>

								<!-- File list -->
								<ul class="pj-file-list" role="list">
									{#each project.files as file (file.path)}
										<li class="pj-file-list__item">
											<button
												class="pj-file-btn"
												onclick={() => openFile(node.slug, file.path)}
												title={file.path}
											>
												<svg class="pj-file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
													<path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
												</svg>
												<span class="pj-file-btn__name">{file.name}</span>
											</button>
										</li>
									{/each}
								</ul>
							</div>
						{/each}
					</div>
				</section>
			{/each}
		{/if}

		{#if !loading && nodeProjects.length === 0 && !error}
			<div class="pj-empty">
				<svg class="pj-empty__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
				</svg>
				<p>No projects found across nodes.</p>
			</div>
		{/if}
	</div>

	<!-- Side panel -->
	{#if panelOpen}
		<aside class="pj-panel" aria-label="File content">
			<div class="pj-panel__header">
				<span class="pj-panel__title">{panelTitle}</span>
				<button class="pj-btn pj-btn--ghost" onclick={closePanel} aria-label="Close panel">
					<svg class="pj-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>

			<div class="pj-panel__body">
				{#if panelLoading}
					<div class="pj-panel__loading">
						<svg class="pj-icon pj-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
						</svg>
						Loading…
					</div>
				{:else if panelError}
					<div class="pj-error" role="alert">{panelError}</div>
				{:else}
					<pre class="pj-panel__pre"><code>{panelContent}</code></pre>
				{/if}
			</div>
		</aside>
	{/if}

</div>

<style>
/* ─── Root layout ────────────────────────────────────────────────────────── */
.pj-root {
	display: flex;
	height: 100%;
	min-height: 0;
	background: var(--color-bg, #0f1117);
	color: var(--color-text, #e2e8f0);
}

.pj-main {
	flex: 1;
	min-width: 0;
	overflow-y: auto;
	padding: 1.5rem 2rem;
}

/* ─── Header ─────────────────────────────────────────────────────────────── */
.pj-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 1.75rem;
}

.pj-header__left {
	display: flex;
	align-items: baseline;
	gap: 1rem;
}

.pj-header__title {
	font-size: 1.375rem;
	font-weight: 600;
	color: var(--color-text, #e2e8f0);
	margin: 0;
}

.pj-header__meta {
	font-size: 0.8125rem;
	color: var(--color-text-muted, #64748b);
}

/* ─── Buttons ────────────────────────────────────────────────────────────── */
.pj-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.375rem;
	padding: 0.375rem 0.625rem;
	border-radius: 0.375rem;
	border: none;
	cursor: pointer;
	font-size: 0.875rem;
	transition: background 0.15s;
}

.pj-btn--ghost {
	background: transparent;
	color: var(--color-text-secondary, #94a3b8);
}

.pj-btn--ghost:hover:not(:disabled) {
	background: var(--color-bg-secondary, #1e2432);
	color: var(--color-text, #e2e8f0);
}

.pj-btn:disabled {
	opacity: 0.5;
	cursor: not-allowed;
}

/* ─── Icons ──────────────────────────────────────────────────────────────── */
.pj-icon {
	width: 1rem;
	height: 1rem;
	flex-shrink: 0;
}

@keyframes pj-spin {
	to { transform: rotate(360deg); }
}

.pj-spin {
	animation: pj-spin 0.8s linear infinite;
}

/* ─── Error ──────────────────────────────────────────────────────────────── */
.pj-error {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.75rem 1rem;
	border-radius: 0.5rem;
	background: rgba(239,68,68,0.1);
	border: 1px solid rgba(239,68,68,0.25);
	color: #fca5a5;
	font-size: 0.875rem;
	margin-bottom: 1.25rem;
}

.pj-error__retry {
	margin-left: auto;
	padding: 0.25rem 0.625rem;
	border-radius: 0.25rem;
	border: 1px solid rgba(239,68,68,0.4);
	background: transparent;
	color: #fca5a5;
	font-size: 0.8125rem;
	cursor: pointer;
}

.pj-error__retry:hover {
	background: rgba(239,68,68,0.15);
}

/* ─── Skeleton ───────────────────────────────────────────────────────────── */
.pj-skeleton-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	gap: 1rem;
}

.pj-skeleton-card {
	border-radius: 0.625rem;
	background: var(--color-bg-secondary, #1e2432);
	padding: 1rem;
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

.pj-skeleton-bar {
	border-radius: 0.25rem;
	background: var(--color-bg-tertiary, #252d3d);
	animation: pj-pulse 1.4s ease-in-out infinite;
}

.pj-skeleton-bar--header { height: 6px; width: 100%; }
.pj-skeleton-bar--title  { height: 14px; width: 60%; }
.pj-skeleton-bar--line   { height: 12px; width: 100%; }
.pj-skeleton-bar--short  { width: 75%; }

@keyframes pj-pulse {
	0%, 100% { opacity: 1; }
	50%       { opacity: 0.4; }
}

/* ─── Section ────────────────────────────────────────────────────────────── */
.pj-section {
	margin-bottom: 2.25rem;
}

.pj-section__header {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	margin-bottom: 0.875rem;
}

.pj-section__dot {
	width: 10px;
	height: 10px;
	border-radius: 50%;
	background: var(--node-color, #6366f1);
	flex-shrink: 0;
}

.pj-section__name {
	font-size: 0.9375rem;
	font-weight: 600;
	color: var(--color-text, #e2e8f0);
	margin: 0;
}

.pj-section__count {
	font-size: 0.75rem;
	color: var(--color-text-muted, #64748b);
	margin-left: auto;
}

/* ─── Card grid ──────────────────────────────────────────────────────────── */
.pj-card-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	gap: 0.875rem;
}

.pj-card {
	border-radius: 0.625rem;
	border: 1px solid var(--color-border, rgba(255,255,255,0.07));
	background: var(--color-bg-secondary, #1e2432);
	overflow: hidden;
	transition: border-color 0.15s;
}

.pj-card:hover {
	border-color: var(--node-color, #6366f1);
}

/* ─── Card header ────────────────────────────────────────────────────────── */
.pj-card__header {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.75rem 0.875rem;
	background: color-mix(in srgb, var(--node-color, #6366f1) 12%, transparent);
	border-bottom: 1px solid color-mix(in srgb, var(--node-color, #6366f1) 20%, transparent);
}

.pj-card__icon {
	width: 1rem;
	height: 1rem;
	flex-shrink: 0;
	color: var(--node-color, #6366f1);
}

.pj-card__name {
	font-size: 0.8125rem;
	font-weight: 600;
	color: var(--color-text, #e2e8f0);
	flex: 1;
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.pj-card__file-count {
	font-size: 0.6875rem;
	color: var(--color-text-muted, #64748b);
	white-space: nowrap;
}

/* ─── File list ──────────────────────────────────────────────────────────── */
.pj-file-list {
	list-style: none;
	margin: 0;
	padding: 0.375rem 0;
}

.pj-file-list__item {
	display: flex;
}

.pj-file-btn {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	width: 100%;
	padding: 0.375rem 0.875rem;
	background: transparent;
	border: none;
	cursor: pointer;
	text-align: left;
	transition: background 0.12s;
}

.pj-file-btn:hover {
	background: color-mix(in srgb, var(--node-color, #6366f1) 8%, transparent);
}

.pj-file-icon {
	width: 0.875rem;
	height: 0.875rem;
	flex-shrink: 0;
	color: var(--color-text-muted, #64748b);
}

.pj-file-btn__name {
	font-size: 0.8125rem;
	color: var(--color-text-secondary, #94a3b8);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	transition: color 0.12s;
}

.pj-file-btn:hover .pj-file-btn__name {
	color: var(--color-text, #e2e8f0);
}

/* ─── Empty state ────────────────────────────────────────────────────────── */
.pj-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.75rem;
	padding: 4rem 2rem;
	color: var(--color-text-muted, #64748b);
	text-align: center;
}

.pj-empty__icon {
	width: 2.5rem;
	height: 2.5rem;
}

/* ─── Side panel ─────────────────────────────────────────────────────────── */
.pj-root--panel .pj-main {
	border-right: 1px solid var(--color-border, rgba(255,255,255,0.07));
}

.pj-panel {
	width: 480px;
	flex-shrink: 0;
	display: flex;
	flex-direction: column;
	background: var(--color-bg-secondary, #1e2432);
	overflow: hidden;
}

.pj-panel__header {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.875rem 1rem;
	border-bottom: 1px solid var(--color-border, rgba(255,255,255,0.07));
	flex-shrink: 0;
}

.pj-panel__title {
	font-size: 0.875rem;
	font-weight: 600;
	color: var(--color-text, #e2e8f0);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	flex: 1;
	min-width: 0;
}

.pj-panel__body {
	flex: 1;
	overflow-y: auto;
	padding: 1rem;
}

.pj-panel__loading {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	color: var(--color-text-muted, #64748b);
	font-size: 0.875rem;
	padding: 2rem 0;
	justify-content: center;
}

.pj-panel__pre {
	margin: 0;
	font-family: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;
	font-size: 0.8125rem;
	line-height: 1.6;
	color: var(--color-text-secondary, #94a3b8);
	white-space: pre-wrap;
	word-break: break-word;
}
</style>
