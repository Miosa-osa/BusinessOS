<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { optimalStore } from '$lib/stores/optimal';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';
	import type { OptimalNode, OptimalSignalFile, FileEntry, SelectedFile } from '$lib/stores/optimal';

	// ── Store state ────────────────────────────────────────────────────────────
	let store = $derived($optimalStore);
	let node = $derived(store.selectedNode);
	let signals = $derived(store.selectedNodeSignals);
	let isLoading = $derived(store.loading);
	let error = $derived(store.error);

	// ── UI state ───────────────────────────────────────────────────────────────
	let activeTab = $state<'context' | 'signal' | 'signals' | 'files'>('context');
	let expandedSignals = $state<Set<string>>(new Set());

	// ── Files tab state ────────────────────────────────────────────────────────
	let fileTree = $state<FileEntry[]>([]);
	let filesLoading = $state(false);
	let filesError = $state<string | null>(null);
	let expandedDirs = $state<Set<string>>(new Set());
	let selectedFile = $state<SelectedFile | null>(null);
	let fileContentLoading = $state(false);

	// ── Frontmatter metadata ───────────────────────────────────────────────────
	interface FrontmatterMeta {
		type?: string;
		health?: string;
		owner?: string;
		cross_ref?: string[];
	}

	let meta = $derived<FrontmatterMeta>(node ? parseFrontmatter(node.context_md) : {});

	// ── Derived sections ───────────────────────────────────────────────────────
	let contextSections = $derived(node ? parseMarkdownSections(node.context_md) : []);
	let signalSections = $derived(node ? parseMarkdownSections(node.signal_md) : []);

	// ── Mount ──────────────────────────────────────────────────────────────────
	onMount(() => {
		const slug = $page.params.id;
		if (slug) {
			optimalStore.selectNode(slug);
		}
	});

	// ── Fetch helpers ──────────────────────────────────────────────────────────
	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) headers['X-CSRF-Token'] = csrf;
		return headers;
	}

	async function loadFileTree(slug: string) {
		filesLoading = true;
		filesError = null;
		fileTree = [];
		try {
			const res = await fetch(
				`${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/files`,
				{ method: 'GET', headers: buildHeaders(), credentials: 'include', signal: AbortSignal.timeout(8000) }
			);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data: { files?: FileEntry[] } = await res.json();
			fileTree = Array.isArray(data.files) ? data.files : [];
		} catch (err) {
			filesError = err instanceof Error ? err.message : 'Failed to load files';
		} finally {
			filesLoading = false;
		}
	}

	async function loadFileContent(slug: string, filePath: string) {
		fileContentLoading = true;
		selectedFile = null;
		try {
			const res = await fetch(
				`${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/file/${filePath}`,
				{ method: 'GET', headers: buildHeaders(), credentials: 'include', signal: AbortSignal.timeout(8000) }
			);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			selectedFile = await res.json();
		} catch (err) {
			filesError = err instanceof Error ? err.message : 'Failed to load file';
		} finally {
			fileContentLoading = false;
		}
	}

	// ── Tab switching ──────────────────────────────────────────────────────────
	function switchTab(tab: typeof activeTab) {
		activeTab = tab;
		if (tab === 'files' && node && fileTree.length === 0 && !filesLoading) {
			loadFileTree(node.slug);
		}
	}

	// ── Tree helpers ───────────────────────────────────────────────────────────
	function toggleDir(path: string) {
		const next = new Set(expandedDirs);
		if (next.has(path)) next.delete(path);
		else next.add(path);
		expandedDirs = next;
	}

	function handleFileClick(entry: FileEntry) {
		if (entry.is_dir) {
			toggleDir(entry.path);
		} else if (node) {
			loadFileContent(node.slug, entry.path);
		}
	}

	// ── Signal expand ──────────────────────────────────────────────────────────
	function toggleSignal(slug: string) {
		const next = new Set(expandedSignals);
		if (next.has(slug)) next.delete(slug);
		else next.add(slug);
		expandedSignals = next;
	}

	// ── Markdown parsing ───────────────────────────────────────────────────────
	interface MarkdownSection {
		level: number;
		title: string;
		content: string;
		isTable: boolean;
	}

	function parseFrontmatter(md: string): FrontmatterMeta {
		if (!md || !md.startsWith('---')) return {};
		const end = md.indexOf('\n---', 3);
		if (end <= 0) return {};
		const block = md.slice(3, end).trim();
		const result: FrontmatterMeta = {};
		for (const line of block.split('\n')) {
			const colon = line.indexOf(':');
			if (colon < 0) continue;
			const key = line.slice(0, colon).trim();
			const val = line.slice(colon + 1).trim();
			if (key === 'type') result.type = val;
			else if (key === 'health') result.health = val;
			else if (key === 'owner') result.owner = val;
			else if (key === 'cross_ref') {
				// Accept "- item" list or comma separated
				result.cross_ref = val
					? val.replace(/[\[\]]/g, '').split(',').map(s => s.trim()).filter(Boolean)
					: [];
			}
		}
		return result;
	}

	function parseMarkdownSections(md: string): MarkdownSection[] {
		if (!md) return [];

		// Strip frontmatter
		let content = md;
		if (content.startsWith('---')) {
			const end = content.indexOf('\n---', 3);
			if (end > 0) content = content.slice(end + 4).trim();
		}

		const lines = content.split('\n');
		const sections: MarkdownSection[] = [];
		let currentSection: MarkdownSection | null = null;
		let buffer: string[] = [];

		function flushBuffer() {
			if (currentSection) {
				currentSection.content = buffer.join('\n').trim();
				currentSection.isTable =
					currentSection.content.includes('|---') ||
					currentSection.content.includes('| ---');
				sections.push(currentSection);
			}
			buffer = [];
		}

		for (const line of lines) {
			const h1 = line.match(/^# (.+)/);
			const h2 = line.match(/^## (.+)/);
			const h3 = line.match(/^### (.+)/);
			if (h1 || h2 || h3) {
				flushBuffer();
				const match = (h1 || h2 || h3)!;
				const level = h1 ? 1 : h2 ? 2 : 3;
				currentSection = { level, title: match[1].trim(), content: '', isTable: false };
			} else {
				buffer.push(line);
			}
		}
		flushBuffer();
		return sections;
	}

	/** Parse and render a markdown table as HTML */
	function renderTable(content: string): string {
		const lines = content.split('\n').filter(l => l.trim().startsWith('|'));
		if (lines.length < 2) return renderContent(content);

		const parseRow = (line: string) =>
			line.split('|').slice(1, -1).map(c => c.trim());

		const headers = parseRow(lines[0]);
		const rows = lines.slice(2).map(parseRow);

		let html = '<table class="nd-table"><thead><tr>';
		for (const h of headers) {
			html += `<th class="nd-th">${escapeHtml(h)}</th>`;
		}
		html += '</tr></thead><tbody>';
		for (const row of rows) {
			html += '<tr class="nd-tr">';
			for (const cell of row) {
				html += `<td class="nd-td">${renderInline(cell)}</td>`;
			}
			html += '</tr>';
		}
		html += '</tbody></table>';
		return html;
	}

	function escapeHtml(s: string): string {
		return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	}

	function renderInline(text: string): string {
		return escapeHtml(text)
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/\*(.+?)\*/g, '<em>$1</em>')
			.replace(/`(.+?)`/g, '<code class="nd-code">$1</code>');
	}

	function renderContent(content: string): string {
		if (!content) return '';
		// Check if this block contains a table — if so, handle table lines separately
		if (content.includes('|---') || content.includes('| ---')) {
			const tableLines = content.split('\n');
			const tableStart = tableLines.findIndex(l => l.trim().startsWith('|'));
			if (tableStart >= 0) {
				const before = tableLines.slice(0, tableStart).join('\n');
				const tableBlock = tableLines.slice(tableStart).join('\n');
				return (before ? renderContent(before) + '\n' : '') + renderTable(tableBlock);
			}
		}
		return content
			.replace(/^- \[x\] (.+)$/gm, '<div class="nd-check nd-check--done"><span class="nd-check-icon">✓</span><span>$1</span></div>')
			.replace(/^- \[ \] (.+)$/gm, '<div class="nd-check"><span class="nd-check-icon nd-check-icon--empty">○</span><span>$1</span></div>')
			.replace(/^- (.+)$/gm, '<div class="nd-li"><span class="nd-bullet">•</span><span>$1</span></div>')
			.replace(/^\d+\. (.+)$/gm, '<div class="nd-oli">$1</div>')
			.replace(/\*\*(.+?)\*\*/g, '<strong class="nd-strong">$1</strong>')
			.replace(/\*(.+?)\*/g, '<em class="nd-em">$1</em>')
			.replace(/`(.+?)`/g, '<code class="nd-code">$1</code>')
			.replace(/^> (.+)$/gm, '<div class="nd-blockquote">$1</div>')
			.replace(/\n{2,}/g, '<div class="nd-spacer"></div>')
			.replace(/\n/g, '<br/>');
	}

	// ── Formatting helpers ─────────────────────────────────────────────────────
	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
		} catch { return dateStr; }
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes}B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
	}

	function getTypeColor(type: string): string {
		const map: Record<string, string> = {
			person: 'nd-badge--blue',
			entity: 'nd-badge--purple',
			'operation:program': 'nd-badge--green',
			'operation:research': 'nd-badge--yellow',
			'operation:network': 'nd-badge--orange',
			'domain:media': 'nd-badge--pink',
			'unit:community': 'nd-badge--cyan',
			'command-center': 'nd-badge--red',
			inbox: 'nd-badge--amber',
			'cross-cutting': 'nd-badge--teal',
			registry: 'nd-badge--indigo',
		};
		return map[type] ?? 'nd-badge--default';
	}

	function getHealthColor(health: string): string {
		const h = health?.toLowerCase();
		if (h === 'healthy' || h === 'green') return 'nd-health--green';
		if (h === 'warning' || h === 'yellow') return 'nd-health--yellow';
		if (h === 'critical' || h === 'red') return 'nd-health--red';
		return 'nd-health--neutral';
	}

	function getSectionTitleClass(level: number): string {
		if (level === 1) return 'nd-section-title--h1';
		if (level === 2) return 'nd-section-title--h2';
		return 'nd-section-title--h3';
	}
</script>

<div class="nd-root">
	{#if isLoading}
		<div class="nd-center">
			<div class="nd-spinner"></div>
		</div>

	{:else if error}
		<div class="nd-center nd-center--col">
			<p class="nd-error-text">{error}</p>
			<button onclick={() => goto('/nodes')} class="nd-back-btn">Back to Nodes</button>
		</div>

	{:else if node}
		<!-- Header -->
		<div class="nd-header">
			<button onclick={() => goto('/nodes')} class="nd-nav-back" aria-label="Back to nodes">
				<svg class="nd-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
				</svg>
			</button>
			<div class="nd-header-info">
				<div class="nd-header-row">
					<h1 class="nd-node-title">{node.name}</h1>
					<span class="nd-badge {getTypeColor(node.type)}">{node.type || 'node'}</span>
					{#if node.has_signals}
						<span class="nd-signal-count">{node.signal_count} signals</span>
					{/if}
				</div>
				<p class="nd-node-slug">{node.slug}</p>
			</div>
		</div>

		<!-- Metadata bar (frontmatter fields) -->
		{#if meta.health || meta.owner || (meta.cross_ref && meta.cross_ref.length > 0)}
			<div class="nd-meta-bar">
				{#if meta.health}
					<div class="nd-meta-item">
						<span class="nd-meta-label">Health</span>
						<span class="nd-health {getHealthColor(meta.health)}">{meta.health}</span>
					</div>
				{/if}
				{#if meta.owner}
					<div class="nd-meta-item">
						<span class="nd-meta-label">Owner</span>
						<span class="nd-meta-value">{meta.owner}</span>
					</div>
				{/if}
				{#if meta.cross_ref && meta.cross_ref.length > 0}
					<div class="nd-meta-item nd-meta-item--refs">
						<span class="nd-meta-label">Cross-refs</span>
						<div class="nd-ref-list">
							{#each meta.cross_ref as ref}
								<span class="nd-ref-tag">{ref}</span>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Tabs -->
		<div class="nd-tabs">
			<button
				class="nd-tab {activeTab === 'context' ? 'nd-tab--active nd-tab--blue' : ''}"
				onclick={() => switchTab('context')}
			>Context</button>
			<button
				class="nd-tab {activeTab === 'signal' ? 'nd-tab--active nd-tab--green' : ''}"
				onclick={() => switchTab('signal')}
			>Weekly Signal</button>
			<button
				class="nd-tab {activeTab === 'signals' ? 'nd-tab--active nd-tab--purple' : ''}"
				onclick={() => switchTab('signals')}
			>Signal Files ({node.signal_count})</button>
			<button
				class="nd-tab {activeTab === 'files' ? 'nd-tab--active nd-tab--amber' : ''}"
				onclick={() => switchTab('files')}
			>Files</button>
		</div>

		<!-- Tab content -->
		<div class="nd-content">
			<!-- Context tab -->
			{#if activeTab === 'context'}
				<div class="nd-sections">
					{#each contextSections as section}
						<div class="nd-section">
							<div class="nd-section-header">
								<span class="nd-section-title {getSectionTitleClass(section.level)}">
									{section.title}
								</span>
							</div>
							<div class="nd-section-body">
								{#if section.isTable}
									{@html renderTable(section.content)}
								{:else}
									<div class="nd-prose">
										{@html renderContent(section.content)}
									</div>
								{/if}
							</div>
						</div>
					{/each}
					{#if contextSections.length === 0}
						<p class="nd-empty">No context data</p>
					{/if}
				</div>

			<!-- Weekly Signal tab -->
			{:else if activeTab === 'signal'}
				<div class="nd-sections">
					{#each signalSections as section}
						<div class="nd-section">
							<div class="nd-section-header">
								<span class="nd-section-title nd-section-title--h2">{section.title}</span>
							</div>
							<div class="nd-section-body">
								{#if section.isTable}
									{@html renderTable(section.content)}
								{:else}
									<div class="nd-prose">
										{@html renderContent(section.content)}
									</div>
								{/if}
							</div>
						</div>
					{/each}
					{#if signalSections.length === 0}
						<p class="nd-empty">No weekly signal data</p>
					{/if}
				</div>

			<!-- Signal Files tab -->
			{:else if activeTab === 'signals'}
				<div class="nd-signal-list">
					{#if signals.length === 0}
						<p class="nd-empty">No signal files</p>
					{:else}
						{#each signals as sig}
							<button
								class="nd-signal-item"
								onclick={() => toggleSignal(sig.slug)}
							>
								<div class="nd-signal-row">
									<div class="nd-signal-left">
										<svg
											class="nd-chevron {expandedSignals.has(sig.slug) ? 'nd-chevron--open' : ''}"
											fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
										>
											<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7"/>
										</svg>
										<span class="nd-signal-name">{sig.name}</span>
									</div>
									<span class="nd-signal-date">{formatDate(sig.date)}</span>
								</div>
							</button>
							{#if expandedSignals.has(sig.slug)}
								<div class="nd-signal-body">
									<pre class="nd-signal-pre">{sig.content}</pre>
								</div>
							{/if}
						{/each}
					{/if}
				</div>

			<!-- Files tab -->
			{:else if activeTab === 'files'}
				<div class="nd-files-layout">
					<!-- File tree -->
					<div class="nd-file-tree">
						{#if filesLoading}
							<div class="nd-center"><div class="nd-spinner nd-spinner--sm"></div></div>
						{:else if filesError}
							<p class="nd-error-text nd-error-text--sm">{filesError}</p>
						{:else if fileTree.length === 0}
							<p class="nd-empty">No files found</p>
						{:else}
							{#each fileTree as entry}
								{@render fileNode(entry, 0)}
							{/each}
						{/if}
					</div>

					<!-- File content panel -->
					{#if selectedFile || fileContentLoading}
						<div class="nd-file-panel">
							{#if fileContentLoading}
								<div class="nd-center"><div class="nd-spinner nd-spinner--sm"></div></div>
							{:else if selectedFile}
								<div class="nd-file-panel-header">
									<span class="nd-file-path">{selectedFile.path}</span>
									<button
										class="nd-close-btn"
										onclick={() => { selectedFile = null; }}
										aria-label="Close file"
									>
										<svg class="nd-icon nd-icon--sm" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
										</svg>
									</button>
								</div>
								<pre class="nd-file-content">{selectedFile.content}</pre>
							{/if}
						</div>
					{/if}
				</div>
			{/if}
		</div>

	{:else}
		<div class="nd-center nd-center--col">
			<p class="nd-empty">Node not found</p>
			<button onclick={() => goto('/nodes')} class="nd-back-btn">Back to Nodes</button>
		</div>
	{/if}
</div>

{#snippet fileNode(entry: FileEntry, depth: number)}
	<div class="nd-file-row" style="padding-left: {depth * 16 + 8}px">
		<button
			class="nd-file-btn {!entry.is_dir && selectedFile?.path === entry.path ? 'nd-file-btn--active' : ''}"
			onclick={() => handleFileClick(entry)}
		>
			{#if entry.is_dir}
				<svg
					class="nd-file-icon nd-file-icon--dir nd-chevron {expandedDirs.has(entry.path) ? 'nd-chevron--open' : ''}"
					fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7"/>
				</svg>
				<svg class="nd-file-icon nd-file-icon--folder" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z"/>
				</svg>
			{:else}
				<svg class="nd-file-icon nd-file-icon--file" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"/>
				</svg>
			{/if}
			<span class="nd-file-name">{entry.name}</span>
			{#if !entry.is_dir && entry.size > 0}
				<span class="nd-file-size">{formatFileSize(entry.size)}</span>
			{/if}
		</button>
	</div>
	{#if entry.is_dir && expandedDirs.has(entry.path) && entry.children}
		{#each entry.children as child}
			{@render fileNode(child, depth + 1)}
		{/each}
	{/if}
{/snippet}

<style>
	/* ── Layout ──────────────────────────────────────────────────────────────── */
	.nd-root {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #0a0a0a;
		color: white;
	}
	.nd-center {
		display: flex;
		flex: 1;
		align-items: center;
		justify-content: center;
	}
	.nd-center--col { flex-direction: column; gap: 12px; }

	/* ── Spinner ─────────────────────────────────────────────────────────────── */
	.nd-spinner {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 2px solid rgba(255,255,255,0.12);
		border-top-color: rgba(255,255,255,0.6);
		animation: nd-spin 0.8s linear infinite;
	}
	.nd-spinner--sm { width: 18px; height: 18px; }
	@keyframes nd-spin { to { transform: rotate(360deg); } }

	/* ── Header ──────────────────────────────────────────────────────────────── */
	.nd-header {
		display: flex;
		align-items: center;
		gap: 16px;
		border-bottom: 1px solid rgba(255,255,255,0.08);
		padding: 16px 24px;
		flex-shrink: 0;
	}
	.nd-nav-back {
		color: rgba(255,255,255,0.3);
		transition: color 0.15s;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		border-radius: 6px;
	}
	.nd-nav-back:hover { color: rgba(255,255,255,0.7); }
	.nd-icon { width: 20px; height: 20px; }
	.nd-icon--sm { width: 16px; height: 16px; }
	.nd-header-info { flex: 1; min-width: 0; }
	.nd-header-row {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.nd-node-title { font-size: 17px; font-weight: 600; margin: 0; }
	.nd-node-slug {
		font-size: 11px;
		color: rgba(255,255,255,0.25);
		font-family: monospace;
		margin-top: 2px;
	}
	.nd-signal-count { font-size: 12px; color: rgba(255,255,255,0.3); }

	/* ── Badges ──────────────────────────────────────────────────────────────── */
	.nd-badge {
		border-radius: 9999px;
		padding: 2px 10px;
		font-size: 11px;
		font-weight: 500;
	}
	.nd-badge--blue    { color: #60a5fa; background: rgba(59,130,246,0.12); }
	.nd-badge--purple  { color: #c084fc; background: rgba(168,85,247,0.12); }
	.nd-badge--green   { color: #4ade80; background: rgba(34,197,94,0.12); }
	.nd-badge--yellow  { color: #facc15; background: rgba(234,179,8,0.12); }
	.nd-badge--orange  { color: #fb923c; background: rgba(249,115,22,0.12); }
	.nd-badge--pink    { color: #f472b6; background: rgba(236,72,153,0.12); }
	.nd-badge--cyan    { color: #22d3ee; background: rgba(6,182,212,0.12); }
	.nd-badge--red     { color: #f87171; background: rgba(239,68,68,0.12); }
	.nd-badge--amber   { color: #fbbf24; background: rgba(245,158,11,0.12); }
	.nd-badge--teal    { color: #2dd4bf; background: rgba(20,184,166,0.12); }
	.nd-badge--indigo  { color: #818cf8; background: rgba(99,102,241,0.12); }
	.nd-badge--default { color: rgba(255,255,255,0.4); background: rgba(255,255,255,0.06); }

	/* ── Meta bar ────────────────────────────────────────────────────────────── */
	.nd-meta-bar {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 10px 24px;
		border-bottom: 1px solid rgba(255,255,255,0.05);
		background: rgba(255,255,255,0.015);
		flex-wrap: wrap;
		flex-shrink: 0;
	}
	.nd-meta-item { display: flex; align-items: center; gap: 8px; }
	.nd-meta-item--refs { align-items: flex-start; }
	.nd-meta-label { font-size: 11px; color: rgba(255,255,255,0.3); text-transform: uppercase; letter-spacing: 0.05em; }
	.nd-meta-value { font-size: 13px; color: rgba(255,255,255,0.7); }
	.nd-health {
		font-size: 12px;
		font-weight: 500;
		border-radius: 4px;
		padding: 1px 8px;
	}
	.nd-health--green  { color: #4ade80; background: rgba(34,197,94,0.1); }
	.nd-health--yellow { color: #facc15; background: rgba(234,179,8,0.1); }
	.nd-health--red    { color: #f87171; background: rgba(239,68,68,0.1); }
	.nd-health--neutral { color: rgba(255,255,255,0.4); background: rgba(255,255,255,0.06); }
	.nd-ref-list { display: flex; gap: 6px; flex-wrap: wrap; }
	.nd-ref-tag {
		font-size: 11px;
		font-family: monospace;
		color: rgba(255,255,255,0.5);
		background: rgba(255,255,255,0.05);
		border: 1px solid rgba(255,255,255,0.08);
		border-radius: 4px;
		padding: 1px 6px;
	}

	/* ── Tabs ────────────────────────────────────────────────────────────────── */
	.nd-tabs {
		display: flex;
		gap: 4px;
		border-bottom: 1px solid rgba(255,255,255,0.08);
		padding: 0 24px;
		flex-shrink: 0;
	}
	.nd-tab {
		padding: 10px 16px;
		font-size: 13px;
		color: rgba(255,255,255,0.35);
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		cursor: pointer;
		transition: color 0.15s, border-color 0.15s;
		white-space: nowrap;
	}
	.nd-tab:hover { color: rgba(255,255,255,0.6); }
	.nd-tab--active { color: white; }
	.nd-tab--active.nd-tab--blue   { border-bottom-color: #60a5fa; }
	.nd-tab--active.nd-tab--green  { border-bottom-color: #4ade80; }
	.nd-tab--active.nd-tab--purple { border-bottom-color: #c084fc; }
	.nd-tab--active.nd-tab--amber  { border-bottom-color: #fbbf24; }

	/* ── Scrollable content ──────────────────────────────────────────────────── */
	.nd-content {
		flex: 1;
		overflow-y: auto;
		min-height: 0;
	}

	/* ── Sections (context + signal tabs) ───────────────────────────────────── */
	.nd-sections { padding: 24px; display: flex; flex-direction: column; gap: 16px; }
	.nd-section {
		border-radius: 10px;
		background: rgba(255,255,255,0.025);
		border: 1px solid rgba(255,255,255,0.06);
		overflow: hidden;
	}
	.nd-section-header {
		padding: 10px 16px;
		border-bottom: 1px solid rgba(255,255,255,0.06);
	}
	.nd-section-title { font-size: 13px; font-weight: 500; }
	.nd-section-title--h1 { color: rgba(255,255,255,0.95); }
	.nd-section-title--h2 { color: rgba(255,255,255,0.75); }
	.nd-section-title--h3 { color: rgba(255,255,255,0.55); }
	.nd-section-body { padding: 12px 16px; }

	/* ── Prose (rendered markdown) ───────────────────────────────────────────── */
	.nd-prose { font-size: 13px; color: rgba(255,255,255,0.65); line-height: 1.6; }
	:global(.nd-li) { display: flex; align-items: flex-start; gap: 8px; padding: 2px 0; }
	:global(.nd-bullet) { color: rgba(255,255,255,0.25); margin-top: 2px; flex-shrink: 0; }
	:global(.nd-oli) { color: rgba(255,255,255,0.75); padding: 2px 0 2px 16px; }
	:global(.nd-check) { display: flex; align-items: center; gap: 8px; padding: 2px 0; }
	:global(.nd-check--done span:last-child) { text-decoration: line-through; color: rgba(255,255,255,0.35); }
	:global(.nd-check-icon) { color: #4ade80; font-size: 12px; flex-shrink: 0; }
	:global(.nd-check-icon--empty) { color: rgba(255,255,255,0.2); }
	:global(.nd-blockquote) {
		border-left: 2px solid rgba(255,255,255,0.15);
		padding: 2px 12px;
		color: rgba(255,255,255,0.4);
		font-style: italic;
	}
	:global(.nd-strong) { color: white; font-weight: 500; }
	:global(.nd-em) { color: rgba(255,255,255,0.55); }
	:global(.nd-code) {
		color: #93c5fd;
		background: rgba(59,130,246,0.1);
		padding: 1px 5px;
		border-radius: 4px;
		font-size: 12px;
		font-family: monospace;
	}
	:global(.nd-spacer) { height: 8px; }

	/* ── Tables ──────────────────────────────────────────────────────────────── */
	:global(.nd-table) { width: 100%; font-size: 12px; border-collapse: collapse; }
	:global(.nd-th) {
		text-align: left;
		padding: 8px 12px;
		color: rgba(255,255,255,0.5);
		border-bottom: 1px solid rgba(255,255,255,0.08);
		font-weight: 500;
		white-space: nowrap;
	}
	:global(.nd-tr) { border-bottom: 1px solid rgba(255,255,255,0.04); }
	:global(.nd-tr:hover) { background: rgba(255,255,255,0.02); }
	:global(.nd-td) { padding: 7px 12px; color: rgba(255,255,255,0.75); }

	/* ── Signal files tab ────────────────────────────────────────────────────── */
	.nd-signal-list { padding: 16px 24px; display: flex; flex-direction: column; gap: 4px; }
	.nd-signal-item {
		width: 100%;
		text-align: left;
		border: 1px solid rgba(255,255,255,0.06);
		border-radius: 8px;
		background: rgba(255,255,255,0.015);
		cursor: pointer;
		overflow: hidden;
		transition: background 0.15s;
	}
	.nd-signal-item:hover { background: rgba(255,255,255,0.04); }
	.nd-signal-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 16px;
	}
	.nd-signal-left { display: flex; align-items: center; gap: 10px; }
	.nd-signal-name { font-size: 13px; color: rgba(255,255,255,0.8); }
	.nd-signal-date { font-size: 12px; color: rgba(255,255,255,0.3); }
	.nd-signal-body {
		margin-left: 16px;
		margin-bottom: 8px;
		border-radius: 8px;
		background: rgba(255,255,255,0.015);
		border: 1px solid rgba(255,255,255,0.04);
		padding: 16px;
	}
	.nd-signal-pre {
		font-size: 12px;
		color: rgba(255,255,255,0.55);
		white-space: pre-wrap;
		font-family: monospace;
		line-height: 1.6;
		overflow-x: auto;
		margin: 0;
	}

	/* ── Chevron ─────────────────────────────────────────────────────────────── */
	.nd-chevron { width: 14px; height: 14px; color: rgba(255,255,255,0.3); transition: transform 0.15s; flex-shrink: 0; }
	.nd-chevron--open { transform: rotate(90deg); }

	/* ── Files tab ───────────────────────────────────────────────────────────── */
	.nd-files-layout {
		display: flex;
		height: 100%;
		min-height: 0;
	}
	.nd-file-tree {
		width: 280px;
		flex-shrink: 0;
		border-right: 1px solid rgba(255,255,255,0.06);
		overflow-y: auto;
		padding: 8px 0;
	}
	.nd-file-row { display: flex; }
	.nd-file-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		padding: 5px 8px 5px 0;
		background: none;
		border: none;
		cursor: pointer;
		color: rgba(255,255,255,0.65);
		font-size: 12.5px;
		text-align: left;
		border-radius: 4px;
		transition: background 0.12s, color 0.12s;
		white-space: nowrap;
		overflow: hidden;
	}
	.nd-file-btn:hover { background: rgba(255,255,255,0.05); color: white; }
	.nd-file-btn--active { background: rgba(250,191,36,0.1); color: #fbbf24; }
	.nd-file-icon { width: 14px; height: 14px; flex-shrink: 0; }
	.nd-file-icon--dir  { color: rgba(255,255,255,0.25); }
	.nd-file-icon--folder { color: #fbbf24; }
	.nd-file-icon--file { color: rgba(255,255,255,0.35); }
	.nd-file-name { flex: 1; overflow: hidden; text-overflow: ellipsis; }
	.nd-file-size { font-size: 11px; color: rgba(255,255,255,0.2); flex-shrink: 0; margin-left: 4px; }

	.nd-file-panel {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		overflow: hidden;
	}
	.nd-file-panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 16px;
		border-bottom: 1px solid rgba(255,255,255,0.06);
		flex-shrink: 0;
	}
	.nd-file-path { font-size: 12px; font-family: monospace; color: rgba(255,255,255,0.5); }
	.nd-close-btn {
		background: none;
		border: none;
		cursor: pointer;
		color: rgba(255,255,255,0.3);
		padding: 2px;
		border-radius: 4px;
		transition: color 0.12s;
	}
	.nd-close-btn:hover { color: rgba(255,255,255,0.7); }
	.nd-file-content {
		flex: 1;
		overflow: auto;
		padding: 16px;
		font-size: 12px;
		font-family: monospace;
		color: rgba(255,255,255,0.65);
		white-space: pre;
		line-height: 1.6;
		margin: 0;
	}

	/* ── Misc ────────────────────────────────────────────────────────────────── */
	.nd-empty { color: rgba(255,255,255,0.25); font-size: 13px; text-align: center; padding: 32px 0; }
	.nd-error-text { color: #f87171; font-size: 13px; }
	.nd-error-text--sm { font-size: 12px; padding: 8px 16px; }
	.nd-back-btn {
		border-radius: 8px;
		background: rgba(255,255,255,0.08);
		border: none;
		padding: 8px 16px;
		font-size: 13px;
		color: rgba(255,255,255,0.7);
		cursor: pointer;
		transition: background 0.15s;
	}
	.nd-back-btn:hover { background: rgba(255,255,255,0.12); }
</style>
