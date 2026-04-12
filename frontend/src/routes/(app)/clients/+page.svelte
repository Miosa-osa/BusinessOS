<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ── Types ──────────────────────────────────────────────────────────────────
	interface Client {
		name: string;
		status: 'active' | 'pipeline' | 'stalled' | 'legal';
		value: string;
		valueRaw: number;
		nextAction: string;
		owner: string;
		source: string;       // node slug
		nodeLabel: string;    // human label
		filePath?: string;    // file path to open in /pages
	}

	interface FileNode {
		path: string;
		name: string;
		type: 'file' | 'directory';
		children?: FileNode[];
	}

	// ── State ─────────────────────────────────────────────────────────────────
	let clients    = $state<Client[]>([]);
	let loading    = $state(true);
	let error      = $state<string | null>(null);
	let expandedNodes = $state<Set<string>>(new Set());

	// ── Helpers ───────────────────────────────────────────────────────────────
	function buildHeaders(): Record<string, string> {
		const headers: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) headers['X-CSRF-Token'] = csrf;
		return headers;
	}

	async function fetchNodeContent(slug: string): Promise<{ context_md: string; signal_md: string }> {
		const res = await fetch(`${getApiBaseUrl()}/optimal/nodes/${slug}`, {
			credentials: 'include',
			headers: buildHeaders(),
			signal: AbortSignal.timeout(8000),
		});
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		return res.json();
	}

	// Parse clients table from signal_md — matches "| Client | Status | Revenue | Next Action |"
	function parseClientsTable(md: string): Client[] {
		const lines = md.split('\n');
		const tableStart = lines.findIndex(l => /^\|\s*Client/.test(l));
		if (tableStart === -1) return [];
		const rows: Client[] = [];
		for (let i = tableStart + 2; i < lines.length; i++) {
			const line = lines[i].trim();
			if (!line.startsWith('|')) break;
			const cols = line.split('|').map(c => c.trim()).filter(Boolean);
			if (cols.length < 4) continue;
			const [rawName, rawStatus, rawRevenue, rawNext] = cols;
			const statusMap: Record<string, Client['status']> = {
				'platform build': 'active', 'active': 'active',
				'pipeline': 'pipeline', 'phase 1 closed': 'active',
				'legal issues': 'legal', 'stalled': 'stalled',
			};
			const statusKey = rawStatus.toLowerCase().replace(/[^a-z0-9 ]/g, '').trim();
			const status = Object.entries(statusMap).find(([k]) => statusKey.includes(k))?.[1] ?? 'pipeline';
			const valueRaw = parseFloat(rawRevenue.replace(/[^0-9.]/g, '')) || 0;
			rows.push({
				name:       rawName.replace(/\*\*/g, '').split('(')[0].trim(),
				status,
				value:      rawRevenue.replace(/\*\*/g, ''),
				valueRaw,
				nextAction: rawNext.replace(/\*\*/g, ''),
				owner:      'Roberto',
				source:     '03-lunivate',
				nodeLabel:  'Lunivate',
			});
		}
		return rows;
	}

	// Parse pipeline info from 06-agency-accelerants context
	function parsePipelineClients(contextMd: string): Client[] {
		const clients: Client[] = [];
		// Closers / people as pipeline leads
		if (contextMd.includes('Bennett')) {
			clients.push({
				name: 'Bennett Pipeline (Local Installs)',
				status: 'pipeline',
				value: '$30K+ per install',
				valueRaw: 30000,
				nextAction: 'Bennett autonomous — track for cloud conversion',
				owner: 'Bennett',
				source: '06-agency-accelerants',
				nodeLabel: 'Agency Accelerants',
				filePath: 'nodes/06-agency-accelerants/context.md',
			});
		}
		if (contextMd.includes('Gary')) {
			clients.push({
				name: 'Gary / Applied AI Society',
				status: 'pipeline',
				value: 'TBD (affiliate)',
				valueRaw: 0,
				nextAction: 'Formalize cut structure (10-20%)',
				owner: 'Roberto',
				source: '06-agency-accelerants',
				nodeLabel: 'Agency Accelerants',
				filePath: 'nodes/06-agency-accelerants/context.md',
			});
		}
		return clients;
	}

	async function loadClients() {
		loading = true;
		error   = null;
		try {
			const [lunivateData, aaData] = await Promise.allSettled([
				fetchNodeContent('03-lunivate'),
				fetchNodeContent('06-agency-accelerants'),
			]);

			const result: Client[] = [];

			if (lunivateData.status === 'fulfilled') {
				const { signal_md } = lunivateData.value;
				const parsed = parseClientsTable(signal_md);
				// Enrich with file paths
				const fileMap: Record<string, string> = {
					'ClinicIQ': 'nodes/03-lunivate/context.md',
					'Mosaic Effect': 'nodes/03-lunivate/mosaic-situation/context.md',
					'Amanda': 'nodes/03-lunivate/context.md',
					'Faster Way': 'nodes/03-lunivate/context.md',
				};
				for (const c of parsed) {
					c.filePath = Object.entries(fileMap).find(([k]) => c.name.includes(k))?.[1]
						?? 'nodes/03-lunivate/context.md';
					result.push(c);
				}
			}

			if (aaData.status === 'fulfilled') {
				const pipelineClients = parsePipelineClients(aaData.value.context_md);
				result.push(...pipelineClients);
			}

			clients = result;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load clients';
		} finally {
			loading = false;
		}
	}

	function openFile(filePath: string | undefined) {
		if (!filePath) return;
		const encoded = encodeURIComponent(filePath);
		goto(`/pages?file=${encoded}`);
	}

	function toggleExpand(name: string) {
		const next = new Set(expandedNodes);
		if (next.has(name)) next.delete(name);
		else next.add(name);
		expandedNodes = next;
	}

	const statusConfig: Record<Client['status'], { label: string; cls: string }> = {
		active:   { label: 'Active',    cls: 'cl-badge--active'   },
		pipeline: { label: 'Pipeline',  cls: 'cl-badge--pipeline' },
		stalled:  { label: 'Stalled',   cls: 'cl-badge--stalled'  },
		legal:    { label: 'Legal',     cls: 'cl-badge--legal'    },
	};

	const totalACV = $derived(clients.reduce((sum, c) => sum + c.valueRaw, 0));
	const activeCount = $derived(clients.filter(c => c.status === 'active').length);

	onMount(loadClients);
</script>

<div class="cl-page">
	<!-- Header -->
	<div class="cl-header">
		<div>
			<h1 class="cl-header__title">Clients</h1>
			<p class="cl-header__sub">Live from OptimalOS nodes 03-lunivate &amp; 06-agency-accelerants</p>
		</div>
		<button onclick={loadClients} class="cl-refresh" aria-label="Refresh clients">
			<svg class="cl-refresh__icon" class:cl-spin={loading} fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
		</button>
	</div>

	<!-- Stats bar -->
	{#if !loading && clients.length > 0}
		<div class="cl-stats">
			<div class="cl-stat">
				<span class="cl-stat__val">{clients.length}</span>
				<span class="cl-stat__lbl">Total</span>
			</div>
			<div class="cl-stat">
				<span class="cl-stat__val">{activeCount}</span>
				<span class="cl-stat__lbl">Active</span>
			</div>
			<div class="cl-stat">
				<span class="cl-stat__val">${totalACV.toLocaleString()}</span>
				<span class="cl-stat__lbl">Pipeline ACV</span>
			</div>
		</div>
	{/if}

	<!-- Error -->
	{#if error}
		<div class="cl-error" role="alert">
			<span>{error}</span>
			<button onclick={loadClients} class="cl-error__retry">Retry</button>
		</div>
	{/if}

	<!-- Loading -->
	{#if loading}
		<div class="cl-center">
			<svg class="cl-spin w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
			<p class="cl-center__txt">Loading from OptimalOS...</p>
		</div>
	{:else if clients.length === 0 && !error}
		<div class="cl-center">
			<p class="cl-center__txt">No clients found in nodes</p>
		</div>
	{:else}
		<!-- Client grid -->
		<div class="cl-grid">
			{#each clients as client (client.name)}
				<div class="cl-card">
					<div class="cl-card__top">
						<div class="cl-card__name-row">
							<span class="cl-card__name">{client.name}</span>
							<span class="cl-badge {statusConfig[client.status].cls}">
								{statusConfig[client.status].label}
							</span>
						</div>
						<div class="cl-card__meta-row">
							<span class="cl-node-pill">{client.nodeLabel}</span>
							<span class="cl-card__owner">Owner: {client.owner}</span>
						</div>
					</div>

					<div class="cl-card__body">
						<div class="cl-card__field">
							<span class="cl-card__label">Value</span>
							<span class="cl-card__value cl-card__value--green">{client.value}</span>
						</div>
						<div class="cl-card__field">
							<span class="cl-card__label">Next Action</span>
							<span class="cl-card__value">{client.nextAction}</span>
						</div>
					</div>

					{#if client.filePath}
						<button
							onclick={() => openFile(client.filePath)}
							class="cl-card__open"
							aria-label="Open {client.name} context file"
						>
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
									d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
							</svg>
							Open in editor
						</button>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Source links -->
		<div class="cl-sources">
			<span class="cl-sources__lbl">Data sources:</span>
			<button onclick={() => openFile('nodes/03-lunivate/context.md')} class="cl-sources__link">
				03-lunivate/context.md
			</button>
			<button onclick={() => openFile('nodes/03-lunivate/signal.md')} class="cl-sources__link">
				03-lunivate/signal.md
			</button>
			<button onclick={() => openFile('nodes/06-agency-accelerants/context.md')} class="cl-sources__link">
				06-agency-accelerants/context.md
			</button>
		</div>
	{/if}
</div>

<style>
/* ─── Clients page — cl- prefix ─────────────────────────────────────────── */
.cl-page {
	display: flex;
	flex-direction: column;
	height: 100%;
	background: var(--dbg, #0a0a0a);
	overflow-y: auto;
}

.cl-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20px 24px 16px;
	border-bottom: 1px solid var(--dbd, rgba(255,255,255,0.06));
	flex-shrink: 0;
}
.cl-header__title {
	font-size: 20px;
	font-weight: 700;
	color: var(--dt, #fff);
	letter-spacing: -0.01em;
}
.cl-header__sub {
	font-size: 12px;
	color: var(--dt3, rgba(255,255,255,0.35));
	margin-top: 3px;
}
.cl-refresh {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 32px;
	height: 32px;
	border-radius: 8px;
	background: rgba(255,255,255,0.05);
	border: 1px solid var(--dbd, rgba(255,255,255,0.08));
	cursor: pointer;
	color: var(--dt3, rgba(255,255,255,0.4));
	transition: background 0.15s, color 0.15s;
}
.cl-refresh:hover { background: rgba(255,255,255,0.08); color: var(--dt, #fff); }
.cl-refresh__icon { width: 15px; height: 15px; }

/* Stats bar */
.cl-stats {
	display: flex;
	gap: 0;
	border-bottom: 1px solid var(--dbd, rgba(255,255,255,0.06));
	flex-shrink: 0;
}
.cl-stat {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 12px 16px;
	border-right: 1px solid var(--dbd, rgba(255,255,255,0.06));
}
.cl-stat:last-child { border-right: none; }
.cl-stat__val {
	font-size: 22px;
	font-weight: 700;
	color: var(--dt, #fff);
	letter-spacing: -0.02em;
}
.cl-stat__lbl {
	font-size: 11px;
	color: var(--dt3, rgba(255,255,255,0.35));
	text-transform: uppercase;
	letter-spacing: 0.05em;
	margin-top: 2px;
}

/* Error */
.cl-error {
	margin: 12px 24px;
	padding: 10px 14px;
	border-radius: 8px;
	background: rgba(239,68,68,0.08);
	border: 1px solid rgba(239,68,68,0.2);
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: 13px;
	color: #f87171;
}
.cl-error__retry {
	font-size: 12px;
	color: #f87171;
	text-decoration: underline;
	background: none;
	border: none;
	cursor: pointer;
}

/* Center state */
.cl-center {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 10px;
}
.cl-center__txt { font-size: 13px; color: var(--dt3, rgba(255,255,255,0.35)); }

/* Grid */
.cl-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	gap: 12px;
	padding: 20px 24px;
}

/* Cards */
.cl-card {
	background: rgba(255,255,255,0.03);
	border: 1px solid var(--dbd, rgba(255,255,255,0.06));
	border-radius: 12px;
	padding: 16px;
	display: flex;
	flex-direction: column;
	gap: 12px;
	transition: border-color 0.15s, background 0.15s;
}
.cl-card:hover {
	border-color: rgba(255,255,255,0.1);
	background: rgba(255,255,255,0.04);
}

.cl-card__top { display: flex; flex-direction: column; gap: 6px; }
.cl-card__name-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.cl-card__name {
	font-size: 14px;
	font-weight: 600;
	color: var(--dt, #fff);
	line-height: 1.3;
}
.cl-card__meta-row { display: flex; align-items: center; gap: 8px; }
.cl-card__owner { font-size: 11px; color: var(--dt4, rgba(255,255,255,0.25)); }

.cl-card__body { display: flex; flex-direction: column; gap: 8px; }
.cl-card__field { display: flex; flex-direction: column; gap: 2px; }
.cl-card__label { font-size: 10px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--dt4, rgba(255,255,255,0.3)); }
.cl-card__value { font-size: 12px; color: var(--dt2, rgba(255,255,255,0.65)); line-height: 1.4; }
.cl-card__value--green { color: #4ade80; font-weight: 600; }

.cl-card__open {
	display: flex;
	align-items: center;
	gap: 5px;
	font-size: 11px;
	color: rgba(99,102,241,0.8);
	background: rgba(99,102,241,0.08);
	border: 1px solid rgba(99,102,241,0.15);
	border-radius: 6px;
	padding: 5px 10px;
	cursor: pointer;
	width: fit-content;
	transition: background 0.15s, color 0.15s;
}
.cl-card__open:hover { background: rgba(99,102,241,0.15); color: #818cf8; }

/* Badge */
.cl-badge {
	font-size: 10px;
	font-weight: 600;
	padding: 2px 7px;
	border-radius: 4px;
	text-transform: uppercase;
	letter-spacing: 0.04em;
	white-space: nowrap;
}
.cl-badge--active   { background: rgba(74,222,128,0.12); color: #4ade80; border: 1px solid rgba(74,222,128,0.2); }
.cl-badge--pipeline { background: rgba(251,191,36,0.12); color: #fbbf24; border: 1px solid rgba(251,191,36,0.2); }
.cl-badge--stalled  { background: rgba(156,163,175,0.12); color: #9ca3af; border: 1px solid rgba(156,163,175,0.2); }
.cl-badge--legal    { background: rgba(239,68,68,0.12); color: #f87171; border: 1px solid rgba(239,68,68,0.2); }

/* Node pill */
.cl-node-pill {
	font-size: 10px;
	color: rgba(129,140,248,0.7);
	background: rgba(99,102,241,0.08);
	border: 1px solid rgba(99,102,241,0.12);
	border-radius: 4px;
	padding: 1px 6px;
}

/* Sources footer */
.cl-sources {
	display: flex;
	align-items: center;
	flex-wrap: wrap;
	gap: 8px;
	padding: 12px 24px 20px;
	border-top: 1px solid var(--dbd, rgba(255,255,255,0.05));
}
.cl-sources__lbl { font-size: 11px; color: var(--dt4, rgba(255,255,255,0.25)); }
.cl-sources__link {
	font-size: 11px;
	color: rgba(99,102,241,0.6);
	background: none;
	border: none;
	cursor: pointer;
	padding: 0;
	text-decoration: underline;
	text-underline-offset: 2px;
}
.cl-sources__link:hover { color: #818cf8; }

/* Spin animation */
@keyframes cl-spin-anim { to { transform: rotate(360deg); } }
:global(.cl-spin) { animation: cl-spin-anim 0.8s linear infinite; }
</style>
