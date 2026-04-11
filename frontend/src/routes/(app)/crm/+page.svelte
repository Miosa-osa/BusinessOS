<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	// ── Types ──────────────────────────────────────────────────────────────────
	type DealTier  = 'tier1' | 'tier2' | 'tier3';
	type DealStage = 'active' | 'selling' | 'pre-launch' | 'strategy' | 'stalled' | 'closed';

	interface Deal {
		name: string;
		client: string;
		entity: string;
		value: string;
		valueRaw: number;
		stage: DealStage;
		tier: DealTier;
		owner: string;
		nextAction: string;
		deadline: string;
		blockers: string;
		filePath: string;
	}

	interface Deliverable {
		client: string;
		item: string;
		status: 'overdue' | 'in-progress' | 'blocked' | 'pending';
		owner: string;
		filePath: string;
	}

	// ── State ─────────────────────────────────────────────────────────────────
	let deals        = $state<Deal[]>([]);
	let deliverables = $state<Deliverable[]>([]);
	let loading      = $state(true);
	let error        = $state<string | null>(null);
	let activeTab    = $state<'pipeline' | 'deliverables'>('pipeline');

	// ── Config ────────────────────────────────────────────────────────────────
	const stageConfig: Record<DealStage, { label: string; cls: string }> = {
		active:       { label: 'Active',      cls: 'crm-badge--active'    },
		selling:      { label: 'Selling',     cls: 'crm-badge--selling'   },
		'pre-launch': { label: 'Pre-Launch',  cls: 'crm-badge--pre'       },
		strategy:     { label: 'Strategy',    cls: 'crm-badge--strategy'  },
		stalled:      { label: 'Stalled',     cls: 'crm-badge--stalled'   },
		closed:       { label: 'Closed',      cls: 'crm-badge--closed'    },
	};

	const delivConfig: Record<Deliverable['status'], { label: string; cls: string }> = {
		overdue:      { label: 'Overdue',     cls: 'crm-badge--overdue'   },
		'in-progress':{ label: 'In Progress', cls: 'crm-badge--selling'   },
		blocked:      { label: 'Blocked',     cls: 'crm-badge--stalled'   },
		pending:      { label: 'Pending',     cls: 'crm-badge--pre'       },
	};

	const TIERS: { key: DealTier; label: string }[] = [
		{ key: 'tier1', label: 'Tier 1 — Active Contracts' },
		{ key: 'tier2', label: 'Tier 2 — Pipeline (Selling Now)' },
		{ key: 'tier3', label: 'Tier 3 — Enterprise Track' },
	];

	// ── Helpers ───────────────────────────────────────────────────────────────
	function buildHeaders(): Record<string, string> {
		const h: Record<string, string> = {};
		const csrf = getCSRFToken();
		if (csrf) h['X-CSRF-Token'] = csrf;
		return h;
	}

	async function fetchFile(slug: string, filePath: string): Promise<string> {
		const res = await fetch(
			`${getApiBaseUrl()}/optimal/nodes/${slug}/file/${filePath}`,
			{ credentials: 'include', headers: buildHeaders(), signal: AbortSignal.timeout(8000) }
		);
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		const data = await res.json() as { content: string };
		return data.content ?? '';
	}

	function field(section: string, key: string): string {
		const re = new RegExp(`\\|\\s*\\*{0,2}${key}\\*{0,2}\\s*\\|\\s*(.+?)\\s*\\|`, 'i');
		return (section.match(re)?.[1] ?? '').replace(/\*\*/g, '').trim();
	}

	function parseDealsFromMd(md: string): Deal[] {
		const parsed: Deal[] = [];
		let tier: DealTier = 'tier1';
		const sections = md.split(/(?=^### )/m);
		for (const section of sections) {
			if (/^## Tier 1/m.test(section)) tier = 'tier1';
			if (/^## Tier 2/m.test(section)) tier = 'tier2';
			if (/^## Tier 3/m.test(section)) tier = 'tier3';
			const nameMatch = section.match(/^### (.+)/m);
			if (!nameMatch) continue;
			const rawStage = field(section, 'Current stage').toLowerCase();
			const stageMap: [string, DealStage][] = [
				['platform build', 'active'], ['active', 'active'], ['closed', 'closed'],
				['selling', 'selling'], ['pre-launch', 'pre-launch'],
				['strategy', 'strategy'], ['stalled', 'stalled'],
			];
			const stage    = stageMap.find(([k]) => rawStage.includes(k))?.[1] ?? 'selling';
			const value    = field(section, 'Contract value');
			const valueRaw = parseFloat(value.replace(/[^0-9.]/g, '')) || 0;
			const blockers = field(section, 'Blockers');
			parsed.push({
				name:       nameMatch[1].trim(),
				client:     field(section, 'Client'),
				entity:     field(section, 'Entity'),
				value,
				valueRaw,
				stage,
				tier,
				owner:      field(section, 'Owner'),
				nextAction: field(section, 'Next action'),
				deadline:   field(section, 'Deadline'),
				blockers,
				filePath:   'nodes/02-miosa/agency/projects/active-deals/phase-1-active-deals.md',
			});
		}
		return parsed.filter(d => d.name && d.client);
	}

	function parseDeliverablesFromMd(md: string): Deliverable[] {
		const result: Deliverable[] = [];
		if (md.includes('Mosaic')) {
			result.push(
				{ client: 'Mosaic Effect', item: 'Headlines / copy rework (10-15 posts) — "breaking news" style', status: 'overdue',     owner: 'Nejd + Ikram', filePath: 'nodes/03-lunivate/mosaic-situation/output/internal/08-Deliverables-Status.md' },
				{ client: 'Mosaic Effect', item: 'Textless versions of post images',                                status: 'blocked',     owner: 'Nejd',         filePath: 'nodes/03-lunivate/mosaic-situation/deliverables/status.md' },
				{ client: 'Mosaic Effect', item: 'Guide consistency fixes (Ikram + Sukhpreet)',                     status: 'in-progress', owner: 'Ikram',        filePath: 'nodes/03-lunivate/mosaic-situation/context.md' },
			);
		}
		if (md.includes('ClinicIQ') || md.includes('HBAI')) {
			result.push({
				client:   'ClinicIQ',
				item:     'In-person platform build (April 8-10, Austin) — context profile builder, module agents, campaign deploy',
				status:   'pending',
				owner:    'Roberto',
				filePath: 'nodes/02-miosa/agency/projects/active-deals/phase-1-active-deals.md',
			});
		}
		return result;
	}

	async function loadCRM() {
		loading = true;
		error   = null;
		try {
			const [dealsRes, agencyRes] = await Promise.allSettled([
				fetchFile('02-miosa', 'agency/projects/active-deals/phase-1-active-deals.md'),
				fetchFile('02-miosa', 'agency/context.md'),
			]);
			const dealsMd  = dealsRes.status  === 'fulfilled' ? dealsRes.value  : '';
			const agencyMd = agencyRes.status === 'fulfilled' ? agencyRes.value : '';
			deals        = parseDealsFromMd(dealsMd);
			deliverables = parseDeliverablesFromMd(dealsMd + '\n' + agencyMd);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load CRM data';
		} finally {
			loading = false;
		}
	}

	function openFile(filePath: string) {
		goto(`/pages?file=${encodeURIComponent(filePath)}`);
	}

	// ── Derived ────────────────────────────────────────────────────────────────
	const dealsByTier    = $derived(TIERS.map(t => ({ ...t, items: deals.filter(d => d.tier === t.key) })));
	const totalPipeline  = $derived(deals.reduce((s, d) => s + d.valueRaw, 0));
	const overdueCount   = $derived(deliverables.filter(d => d.status === 'overdue').length);
	const activeContracts = $derived(deals.filter(d => d.tier === 'tier1').length);

	onMount(loadCRM);
</script>

{#snippet dealCard(deal: Deal)}
	<div class="crm-deal">
		<div class="crm-deal__top">
			<span class="crm-deal__name">{deal.name}</span>
			<span class="crm-badge {stageConfig[deal.stage].cls}">{stageConfig[deal.stage].label}</span>
		</div>
		<div class="crm-deal__client">{deal.client}</div>
		<div class="crm-deal__grid">
			<div class="crm-field">
				<span class="crm-lbl">Value</span>
				<span class="crm-val crm-val--accent">{deal.value}</span>
			</div>
			<div class="crm-field">
				<span class="crm-lbl">Entity</span>
				<span class="crm-val">{deal.entity || '—'}</span>
			</div>
			<div class="crm-field">
				<span class="crm-lbl">Owner</span>
				<span class="crm-val">{deal.owner || '—'}</span>
			</div>
			<div class="crm-field">
				<span class="crm-lbl">Deadline</span>
				<span class="crm-val">{deal.deadline || '—'}</span>
			</div>
		</div>
		{#if deal.nextAction}
			<div class="crm-next">
				<span class="crm-lbl">Next →</span>
				<span class="crm-next__txt">{deal.nextAction}</span>
			</div>
		{/if}
		{#if deal.blockers && !deal.blockers.match(/^none/i)}
			<div class="crm-block">
				<span class="crm-lbl crm-lbl--warn">Blockers:</span>
				<span class="crm-block__txt">{deal.blockers}</span>
			</div>
		{/if}
		<button onclick={() => openFile(deal.filePath)} class="crm-open" aria-label="Open {deal.name}">
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
			</svg>
			Open file
		</button>
	</div>
{/snippet}

<div class="crm-page">
	<!-- Header -->
	<div class="crm-header">
		<div>
			<h1 class="crm-header__title">CRM</h1>
			<p class="crm-header__sub">Live from 02-miosa/agency — active deals &amp; deliverables</p>
		</div>
		<button onclick={loadCRM} class="crm-refresh" aria-label="Refresh CRM">
			<svg class:crm-spin={loading} style="width:15px;height:15px" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
		</button>
	</div>

	<!-- Stats -->
	{#if !loading && deals.length > 0}
		<div class="crm-stats">
			<div class="crm-stat"><span class="crm-stat__val">{deals.length}</span><span class="crm-stat__lbl">Deals</span></div>
			<div class="crm-stat"><span class="crm-stat__val">${totalPipeline.toLocaleString()}</span><span class="crm-stat__lbl">Pipeline</span></div>
			<div class="crm-stat"><span class="crm-stat__val">{activeContracts}</span><span class="crm-stat__lbl">Active Contracts</span></div>
			{#if overdueCount > 0}
				<div class="crm-stat"><span class="crm-stat__val crm-stat__val--warn">{overdueCount}</span><span class="crm-stat__lbl">Overdue</span></div>
			{/if}
		</div>
	{/if}

	<!-- Error -->
	{#if error}
		<div class="crm-error" role="alert">
			<span>{error}</span>
			<button onclick={loadCRM} class="crm-error__retry">Retry</button>
		</div>
	{/if}

	<!-- Tabs -->
	{#if !loading}
		<div class="crm-tabs">
			<button class="crm-tab" class:crm-tab--active={activeTab === 'pipeline'} onclick={() => activeTab = 'pipeline'}>
				Pipeline
			</button>
			<button class="crm-tab" class:crm-tab--active={activeTab === 'deliverables'} onclick={() => activeTab = 'deliverables'}>
				Deliverables
				{#if overdueCount > 0}<span class="crm-tab__badge">{overdueCount}</span>{/if}
			</button>
		</div>
	{/if}

	<!-- Loading -->
	{#if loading}
		<div class="crm-center">
			<svg class="crm-spin" style="width:24px;height:24px" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
			<p class="crm-center__txt">Loading from OptimalOS...</p>
		</div>

	{:else if activeTab === 'pipeline'}
		<div class="crm-pipeline">
			{#each dealsByTier as tier (tier.key)}
				{#if tier.items.length > 0}
					<div class="crm-tier-group">
						<div class="crm-tier-group__hd">
							<span class="crm-tier-group__lbl">{tier.label}</span>
							<span class="crm-tier-group__ct">{tier.items.length}</span>
						</div>
						{#each tier.items as deal (deal.name)}
							{@render dealCard(deal)}
						{/each}
					</div>
				{/if}
			{/each}
			{#if deals.length === 0}
				<div class="crm-center"><p class="crm-center__txt">No deals found in active-deals tracker</p></div>
			{/if}
		</div>

	{:else}
		<div class="crm-deliverables">
			{#if deliverables.length === 0}
				<div class="crm-center"><p class="crm-center__txt">No deliverables found</p></div>
			{:else}
				{#each deliverables as item (item.client + item.item)}
					<div class="crm-deliv">
						<div class="crm-deliv__top">
							<span class="crm-deliv__client">{item.client}</span>
							<span class="crm-badge {delivConfig[item.status].cls}">{delivConfig[item.status].label}</span>
						</div>
						<p class="crm-deliv__item">{item.item}</p>
						<div class="crm-deliv__meta">
							<span class="crm-lbl">Owner: {item.owner}</span>
							<button onclick={() => openFile(item.filePath)} class="crm-open" aria-label="Open {item.client} deliverable">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
										d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
								Open file
							</button>
						</div>
					</div>
				{/each}
			{/if}
		</div>
	{/if}

	<!-- Source footer -->
	{#if !loading}
		<div class="crm-sources">
			<span class="crm-sources__lbl">Source:</span>
			<button onclick={() => openFile('nodes/02-miosa/agency/projects/active-deals/phase-1-active-deals.md')} class="crm-sources__link">
				02-miosa/agency/projects/active-deals/phase-1-active-deals.md
			</button>
		</div>
	{/if}
</div>

<style>
/* ─── CRM page — crm- prefix ─────────────────────────────────────────────── */
.crm-page {
	display: flex; flex-direction: column;
	height: 100%; background: var(--dbg, #0a0a0a); overflow-y: auto;
}
.crm-header {
	display: flex; align-items: center; justify-content: space-between;
	padding: 20px 24px 16px;
	border-bottom: 1px solid var(--dbd, rgba(255,255,255,0.06)); flex-shrink: 0;
}
.crm-header__title { font-size: 20px; font-weight: 700; color: var(--dt,#fff); letter-spacing: -0.01em; }
.crm-header__sub   { font-size: 12px; color: var(--dt3,rgba(255,255,255,0.35)); margin-top: 3px; }
.crm-refresh {
	display: flex; align-items: center; justify-content: center;
	width: 32px; height: 32px; border-radius: 8px;
	background: rgba(255,255,255,0.05); border: 1px solid var(--dbd,rgba(255,255,255,0.08));
	cursor: pointer; color: var(--dt3,rgba(255,255,255,0.4)); transition: background 0.15s, color 0.15s;
}
.crm-refresh:hover { background: rgba(255,255,255,0.08); color: var(--dt,#fff); }

.crm-stats { display: flex; border-bottom: 1px solid var(--dbd,rgba(255,255,255,0.06)); flex-shrink: 0; }
.crm-stat {
	flex: 1; display: flex; flex-direction: column; align-items: center;
	padding: 12px 16px; border-right: 1px solid var(--dbd,rgba(255,255,255,0.06));
}
.crm-stat:last-child { border-right: none; }
.crm-stat__val { font-size: 22px; font-weight: 700; color: var(--dt,#fff); letter-spacing: -0.02em; }
.crm-stat__val--warn { color: #f87171; }
.crm-stat__lbl { font-size: 11px; color: var(--dt3,rgba(255,255,255,0.35)); text-transform: uppercase; letter-spacing: 0.05em; margin-top: 2px; }

.crm-tabs {
	display: flex; gap: 4px; padding: 12px 24px;
	border-bottom: 1px solid var(--dbd,rgba(255,255,255,0.06)); flex-shrink: 0;
}
.crm-tab {
	display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 7px;
	font-size: 13px; font-weight: 500; color: var(--dt3,rgba(255,255,255,0.45));
	background: none; border: 1px solid transparent; cursor: pointer;
	transition: background 0.15s, color 0.15s, border-color 0.15s;
}
.crm-tab:hover { color: var(--dt2,rgba(255,255,255,0.7)); background: rgba(255,255,255,0.04); }
.crm-tab--active { background: rgba(255,255,255,0.06); border-color: rgba(255,255,255,0.1); color: var(--dt,#fff); }
.crm-tab__badge { font-size: 10px; background: rgba(239,68,68,0.2); color: #f87171; border-radius: 4px; padding: 1px 5px; font-weight: 700; }

.crm-error {
	margin: 12px 24px; padding: 10px 14px; border-radius: 8px;
	background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
	display: flex; align-items: center; justify-content: space-between;
	font-size: 13px; color: #f87171;
}
.crm-error__retry { font-size: 12px; color: #f87171; text-decoration: underline; background: none; border: none; cursor: pointer; }

.crm-center { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; padding: 40px; }
.crm-center__txt { font-size: 13px; color: var(--dt3,rgba(255,255,255,0.35)); }

.crm-pipeline { display: flex; flex-direction: column; gap: 24px; padding: 20px 24px; overflow-y: auto; }

.crm-tier-group { display: flex; flex-direction: column; gap: 10px; }
.crm-tier-group__hd { display: flex; align-items: center; gap: 8px; padding-bottom: 8px; border-bottom: 1px solid rgba(255,255,255,0.05); }
.crm-tier-group__lbl { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--dt3,rgba(255,255,255,0.35)); }
.crm-tier-group__ct { font-size: 11px; background: rgba(255,255,255,0.07); color: var(--dt3,rgba(255,255,255,0.4)); border-radius: 4px; padding: 1px 6px; }

.crm-deal {
	background: rgba(255,255,255,0.03); border: 1px solid var(--dbd,rgba(255,255,255,0.06));
	border-radius: 10px; padding: 14px 16px; display: flex; flex-direction: column; gap: 10px;
	transition: border-color 0.15s, background 0.15s;
}
.crm-deal:hover { border-color: rgba(255,255,255,0.09); background: rgba(255,255,255,0.04); }
.crm-deal__top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.crm-deal__name { font-size: 14px; font-weight: 600; color: var(--dt,#fff); }
.crm-deal__client { font-size: 12px; color: var(--dt3,rgba(255,255,255,0.4)); margin-top: -6px; }
.crm-deal__grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); gap: 8px; }

.crm-field { display: flex; flex-direction: column; gap: 2px; }
.crm-lbl { font-size: 10px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--dt4,rgba(255,255,255,0.28)); }
.crm-lbl--warn { color: rgba(248,113,133,0.7); }
.crm-val { font-size: 12px; color: var(--dt2,rgba(255,255,255,0.6)); }
.crm-val--accent { color: #4ade80; font-weight: 600; }

.crm-next { display: flex; align-items: baseline; gap: 6px; background: rgba(255,255,255,0.02); border-radius: 6px; padding: 7px 10px; }
.crm-next__txt { font-size: 12px; color: var(--dt2,rgba(255,255,255,0.6)); line-height: 1.4; }
.crm-block { display: flex; align-items: baseline; gap: 6px; background: rgba(239,68,68,0.04); border-radius: 6px; padding: 6px 10px; }
.crm-block__txt { font-size: 12px; color: rgba(248,113,133,0.7); line-height: 1.4; }

.crm-open {
	display: flex; align-items: center; gap: 5px;
	font-size: 11px; color: rgba(99,102,241,0.7); background: rgba(99,102,241,0.07);
	border: 1px solid rgba(99,102,241,0.12); border-radius: 6px; padding: 4px 10px;
	cursor: pointer; width: fit-content; transition: background 0.15s, color 0.15s;
}
.crm-open:hover { background: rgba(99,102,241,0.14); color: #818cf8; }

.crm-deliverables { display: flex; flex-direction: column; gap: 10px; padding: 20px 24px; }
.crm-deliv {
	background: rgba(255,255,255,0.03); border: 1px solid var(--dbd,rgba(255,255,255,0.06));
	border-radius: 10px; padding: 14px 16px; display: flex; flex-direction: column; gap: 8px;
}
.crm-deliv__top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.crm-deliv__client { font-size: 13px; font-weight: 600; color: var(--dt,#fff); }
.crm-deliv__item { font-size: 12px; color: var(--dt2,rgba(255,255,255,0.6)); line-height: 1.5; }
.crm-deliv__meta { display: flex; align-items: center; gap: 12px; }

.crm-badge { font-size: 10px; font-weight: 600; padding: 2px 7px; border-radius: 4px; text-transform: uppercase; letter-spacing: 0.04em; white-space: nowrap; }
.crm-badge--active   { background: rgba(74,222,128,0.12);  color: #4ade80; border: 1px solid rgba(74,222,128,0.2);  }
.crm-badge--selling  { background: rgba(99,102,241,0.12);  color: #818cf8; border: 1px solid rgba(99,102,241,0.2);  }
.crm-badge--pre      { background: rgba(251,191,36,0.12);  color: #fbbf24; border: 1px solid rgba(251,191,36,0.2);  }
.crm-badge--strategy { background: rgba(139,92,246,0.12);  color: #a78bfa; border: 1px solid rgba(139,92,246,0.2);  }
.crm-badge--stalled  { background: rgba(156,163,175,0.12); color: #9ca3af; border: 1px solid rgba(156,163,175,0.2); }
.crm-badge--closed   { background: rgba(74,222,128,0.07);  color: #6ee7b7; border: 1px solid rgba(74,222,128,0.12); }
.crm-badge--overdue  { background: rgba(239,68,68,0.12);   color: #f87171; border: 1px solid rgba(239,68,68,0.2);   }

.crm-sources {
	display: flex; align-items: center; flex-wrap: wrap; gap: 8px;
	padding: 12px 24px 20px; border-top: 1px solid var(--dbd,rgba(255,255,255,0.05));
	margin-top: auto; flex-shrink: 0;
}
.crm-sources__lbl  { font-size: 11px; color: var(--dt4,rgba(255,255,255,0.25)); }
.crm-sources__link { font-size: 11px; color: rgba(99,102,241,0.6); background: none; border: none; cursor: pointer; padding: 0; text-decoration: underline; text-underline-offset: 2px; }
.crm-sources__link:hover { color: #818cf8; }

@keyframes crm-spin-anim { to { transform: rotate(360deg); } }
:global(.crm-spin) { animation: crm-spin-anim 0.8s linear infinite; }
</style>
