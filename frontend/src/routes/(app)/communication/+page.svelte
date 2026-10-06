<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Archive,
		CalendarDays,
		Check,
		ChevronRight,
		CircleAlert,
		Clock3,
		Inbox,
		Link2,
		Mail,
		MessageCircle,
		MessageSquare,
		RefreshCw,
		Search,
		UserRound
	} from 'lucide-svelte';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import {
		getCommunicationOverview,
		removeCommunicationItem,
		saveCommunicationItem,
		type CommunicationItem,
		type CommunicationOverview,
		type CommunicationProvider,
		type CommunicationState
	} from '$lib/api/communications';

	type View = 'triage' | 'meetings' | 'unrouted';

	let overview = $state<CommunicationOverview | null>(null);
	let activeView = $state<View>('triage');
	let selectedKey = $state<string | null>(null);
	let searchQuery = $state('');
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let isSaving = $state(false);
	let error = $state<string | null>(null);
	let loadedWorkspaceId = $state<string | null>(null);

	const stateLabels: Record<CommunicationState, string> = {
		triage: 'Needs attention',
		waiting: 'Waiting',
		delegated: 'Delegated',
		scheduled: 'Scheduled',
		resolved: 'Resolved',
		archived: 'Archived'
	};

	const providerLabels: Record<CommunicationProvider, string> = {
		gmail: 'Email',
		slack: 'Slack',
		calendar: 'Calendar',
		whatsapp: 'WhatsApp'
	};

	const itemKey = (item: CommunicationItem) =>
		`${item.provider}:${item.source_kind}:${item.source_external_id}`;

	const workspaceItems = $derived(overview?.items ?? []);
	const unroutedItems = $derived(overview?.unrouted_items ?? []);
	const visibleItems = $derived.by(() => {
		const source = activeView === 'unrouted' ? unroutedItems : workspaceItems;
		const query = searchQuery.trim().toLowerCase();
		return source.filter((item) => {
			if (activeView === 'meetings' && item.source_kind !== 'meeting') return false;
			if (activeView === 'triage' && item.state && ['resolved', 'archived'].includes(item.state)) return false;
			if (!query) return true;
			return [item.title, item.preview, item.participant, item.participant_email]
				.filter(Boolean)
				.join(' ')
				.toLowerCase()
				.includes(query);
		});
	});
	const selectedItem = $derived(
		visibleItems.find((item) => itemKey(item) === selectedKey) ?? visibleItems[0] ?? null
	);

	$effect(() => {
		const workspaceId = $currentWorkspace?.id;
		if (!workspaceId || workspaceId === loadedWorkspaceId) return;
		loadedWorkspaceId = workspaceId;
		void loadOverview(workspaceId);
	});

	onMount(() => {
		if ($currentWorkspace?.id) void loadOverview($currentWorkspace.id);
	});

	async function loadOverview(workspaceId = $currentWorkspace?.id, refresh = false) {
		if (!workspaceId) return;
		if (refresh) isRefreshing = true;
		else isLoading = true;
		error = null;
		try {
			overview = await getCommunicationOverview(workspaceId);
			if (!selectedKey && overview.items[0]) selectedKey = itemKey(overview.items[0]);
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load communications.';
			overview = null;
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	async function routeToWorkspace(item: CommunicationItem, state: CommunicationState = 'triage') {
		const workspaceId = $currentWorkspace?.id;
		if (!workspaceId) return;
		isSaving = true;
		try {
			await saveCommunicationItem(workspaceId, {
				provider: item.provider,
				source_kind: item.source_kind,
				source_external_id: item.source_external_id,
				state
			});
			activeView = 'triage';
			selectedKey = itemKey(item);
			await loadOverview(workspaceId, true);
		} finally {
			isSaving = false;
		}
	}

	async function updateState(item: CommunicationItem, state: CommunicationState) {
		await routeToWorkspace(item, state);
	}

	async function unroute(item: CommunicationItem) {
		const workspaceId = $currentWorkspace?.id;
		if (!workspaceId || !item.id) return;
		isSaving = true;
		try {
			await removeCommunicationItem(workspaceId, item.id);
			selectedKey = null;
			await loadOverview(workspaceId, true);
		} finally {
			isSaving = false;
		}
	}

	function selectView(view: View) {
		activeView = view;
		selectedKey = null;
	}

	function formatTime(value: string) {
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '';
		const now = new Date();
		if (date.toDateString() === now.toDateString()) {
			return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
		}
		return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
	}
</script>

<div class="communications-console">
	<header class="console-header">
		<div>
			<div class="eyebrow">Workspace communications</div>
			<h1>Communications</h1>
			<p>{$currentWorkspace?.name ?? 'Select a workspace'}</p>
		</div>
		<button class="icon-button" aria-label="Refresh communications" title="Refresh communications" onclick={() => loadOverview(undefined, true)} disabled={isRefreshing}>
			<RefreshCw size={16} class={isRefreshing ? 'spin' : ''} />
		</button>
	</header>

	{#if error}
		<div class="console-error" role="alert">
			<CircleAlert size={16} />
			<span>{error}</span>
			<button onclick={() => loadOverview()}>Try again</button>
		</div>
	{/if}

	<div class="console-grid">
		<aside class="views-panel">
			<div class="view-list" aria-label="Communication views">
				<button class:active={activeView === 'triage'} onclick={() => selectView('triage')}>
					<Inbox size={16} />
					<span>Needs attention</span>
					<strong>{workspaceItems.filter((item) => !['resolved', 'archived'].includes(item.state ?? 'triage')).length}</strong>
				</button>
				<button class:active={activeView === 'meetings'} onclick={() => selectView('meetings')}>
					<CalendarDays size={16} />
					<span>Meetings</span>
				</button>
				<button class:active={activeView === 'unrouted'} onclick={() => selectView('unrouted')}>
					<Archive size={16} />
					<span>Unrouted</span>
					<strong>{unroutedItems.length}</strong>
				</button>
			</div>

			<div class="connector-status">
				<div class="section-label">Sources</div>
				{#each Object.entries(overview?.connection_state ?? { gmail: false, slack: false, calendar: false, whatsapp: false }) as [provider, connected]}
					<div class="connector-row">
						<span class:connected>{connected}</span>
						{providerLabels[provider as CommunicationProvider]}
					</div>
				{/each}
			</div>
		</aside>

		<section class="list-panel">
			<div class="list-toolbar">
				<label class="search-field">
					<Search size={15} />
					<input bind:value={searchQuery} placeholder="Search this view" aria-label="Search communications" />
				</label>
			</div>

			{#if isLoading}
				<div class="center-state">Loading communications...</div>
			{:else if visibleItems.length === 0}
				<div class="center-state">
					{#if activeView === 'unrouted'}
						No unrouted account activity is available.
					{:else}
						No communications are routed to this workspace yet.
					{/if}
				</div>
			{:else}
				<div class="item-list">
					{#each visibleItems as item (itemKey(item))}
						<button class="item-row" class:selected={itemKey(item) === itemKey(selectedItem ?? item)} onclick={() => (selectedKey = itemKey(item))}>
							<span class="provider-mark" data-provider={item.provider}>
								{#if item.provider === 'gmail'}<Mail size={14} />
								{:else if item.provider === 'calendar'}<CalendarDays size={14} />
								{:else if item.provider === 'whatsapp'}<MessageCircle size={14} />
								{:else}<MessageSquare size={14} />{/if}
							</span>
							<span class="item-copy">
								<span class="item-line"><strong>{item.participant || item.title}</strong><time>{formatTime(item.occurred_at)}</time></span>
								<span class="item-title">{item.title}</span>
								{#if item.preview}<span class="item-preview">{item.preview}</span>{/if}
							</span>
						</button>
					{/each}
				</div>
			{/if}
		</section>

		<section class="detail-panel">
			{#if selectedItem}
				<div class="detail-header">
					<div class="detail-source">
						<span class="provider-mark" data-provider={selectedItem.provider}>
							{#if selectedItem.provider === 'gmail'}<Mail size={14} />
							{:else if selectedItem.provider === 'calendar'}<CalendarDays size={14} />
							{:else if selectedItem.provider === 'whatsapp'}<MessageCircle size={14} />
							{:else}<MessageSquare size={14} />{/if}
						</span>
						<span>{providerLabels[selectedItem.provider]}</span>
					</div>
					{#if activeView !== 'unrouted'}
						<button class="text-button" onclick={() => unroute(selectedItem)} disabled={isSaving}>Remove route</button>
					{/if}
				</div>

				<div class="detail-body">
					<h2>{selectedItem.title}</h2>
					{#if selectedItem.participant || selectedItem.participant_email}
						<div class="person-line"><UserRound size={15} /> {selectedItem.participant || selectedItem.participant_email}</div>
					{/if}
					<p class="message-preview">{selectedItem.preview || 'No preview was provided by this source.'}</p>
				</div>

				<div class="detail-controls">
					{#if activeView === 'unrouted'}
						<button class="primary-command" onclick={() => routeToWorkspace(selectedItem)} disabled={isSaving}>
							<Link2 size={15} /> Route to this workspace
						</button>
					{:else}
						<label class="status-control">
							<span>Status</span>
							<select value={selectedItem.state ?? 'triage'} onchange={(event) => updateState(selectedItem, event.currentTarget.value as CommunicationState)} disabled={isSaving}>
								{#each Object.entries(stateLabels) as [value, label]}
									<option {value}>{label}</option>
								{/each}
							</select>
						</label>
					{/if}
				</div>

				<div class="context-panel">
					<div class="section-label">Workspace context</div>
					<div class="context-row"><Link2 size={14} /> {$currentWorkspace?.name ?? 'No workspace selected'}</div>
					{#if selectedItem.waiting_on}<div class="context-row"><Clock3 size={14} /> Waiting on {selectedItem.waiting_on}</div>{/if}
					{#if selectedItem.assigned_to}<div class="context-row"><UserRound size={14} /> Assigned to {selectedItem.assigned_to}</div>{/if}
				</div>
			{:else}
				<div class="detail-empty">
					<MessageSquare size={28} />
					<p>Select a conversation or meeting.</p>
				</div>
			{/if}
		</section>
	</div>
</div>

<style>
	.communications-console { height: 100%; min-height: 660px; display: flex; flex-direction: column; background: var(--dbg); color: var(--dt); }
	.console-header { min-height: 76px; display: flex; align-items: center; justify-content: space-between; padding: 16px 20px 12px; border-bottom: 1px solid var(--dbd); }
	.console-header h1 { margin: 1px 0 0; font-size: 20px; line-height: 1.2; font-weight: 650; letter-spacing: 0; }
	.console-header p, .eyebrow { margin: 0; color: var(--dt3); font-size: 12px; }
	.eyebrow, .section-label { text-transform: uppercase; font-size: 10px; font-weight: 700; letter-spacing: .07em; color: var(--dt3); }
	.icon-button { width: 32px; height: 32px; display: inline-grid; place-items: center; border: 1px solid var(--dbd); border-radius: 6px; background: var(--dbg); color: var(--dt2); cursor: pointer; }
	.icon-button:hover:not(:disabled) { background: var(--dbg3); color: var(--dt); }
	.icon-button:disabled { opacity: .55; cursor: progress; }
	.console-error { display: flex; align-items: center; gap: 8px; padding: 9px 20px; color: #b45309; background: #fffbeb; border-bottom: 1px solid #fde68a; font-size: 13px; }
	.console-error button { margin-left: auto; border: 0; background: transparent; color: inherit; text-decoration: underline; cursor: pointer; }
	.console-grid { min-height: 0; flex: 1; display: grid; grid-template-columns: 210px minmax(270px, 350px) minmax(390px, 1fr); }
	.views-panel, .list-panel { min-height: 0; border-right: 1px solid var(--dbd); background: var(--dbg); }
	.views-panel { display: flex; flex-direction: column; padding: 12px 8px; }
	.view-list { display: grid; gap: 2px; }
	.view-list button { min-width: 0; height: 34px; display: grid; grid-template-columns: 18px 1fr auto; gap: 8px; align-items: center; padding: 0 9px; border: 0; border-radius: 5px; background: transparent; color: var(--dt2); text-align: left; font-size: 13px; cursor: pointer; }
	.view-list button:hover { background: var(--dbg3); color: var(--dt); }
	.view-list button.active { background: var(--bos-nav-active-bg); color: var(--bos-nav-active); font-weight: 600; }
	.view-list strong { font-size: 11px; font-weight: 600; color: var(--dt3); }
	.connector-status { margin-top: auto; padding: 16px 9px 3px; border-top: 1px solid var(--dbd); }
	.connector-status .section-label { margin-bottom: 9px; }
	.connector-row { display: flex; align-items: center; gap: 8px; min-height: 26px; font-size: 12px; color: var(--dt2); }
	.connector-row span { width: 7px; height: 7px; border-radius: 50%; background: var(--dbd); }
	.connector-row span.connected { background: #22c55e; }
	.list-panel { display: flex; flex-direction: column; }
	.list-toolbar { padding: 12px; border-bottom: 1px solid var(--dbd); }
	.search-field { height: 32px; display: flex; align-items: center; gap: 8px; padding: 0 9px; border: 1px solid var(--dbd); border-radius: 6px; color: var(--dt3); background: var(--dbg2); }
	.search-field:focus-within { border-color: var(--bos-nav-active); box-shadow: 0 0 0 2px color-mix(in srgb, var(--bos-nav-active) 15%, transparent); }
	.search-field input { min-width: 0; width: 100%; border: 0; outline: 0; color: var(--dt); background: transparent; font: inherit; font-size: 13px; }
	.item-list { min-height: 0; overflow: auto; }
	.item-row { width: 100%; display: grid; grid-template-columns: 28px minmax(0, 1fr); gap: 8px; padding: 12px; border: 0; border-bottom: 1px solid var(--dbd); background: transparent; color: inherit; text-align: left; cursor: pointer; }
	.item-row:hover, .item-row.selected { background: var(--bos-nav-active-bg); }
	.provider-mark { width: 25px; height: 25px; display: inline-grid; place-items: center; border: 1px solid var(--dbd); border-radius: 5px; color: var(--dt3); background: var(--dbg2); }
	.provider-mark[data-provider='gmail'] { color: #d95743; } .provider-mark[data-provider='slack'] { color: #7c3aed; } .provider-mark[data-provider='calendar'] { color: #2563eb; } .provider-mark[data-provider='whatsapp'] { color: #16a34a; }
	.item-copy { min-width: 0; display: grid; gap: 3px; }
	.item-line { min-width: 0; display: flex; justify-content: space-between; gap: 8px; font-size: 12px; }
	.item-line strong, .item-title, .item-preview { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.item-line strong { color: var(--dt); font-weight: 600; } .item-line time { flex: 0 0 auto; color: var(--dt3); font-size: 11px; }
	.item-title { color: var(--dt2); font-size: 13px; font-weight: 500; } .item-preview { color: var(--dt3); font-size: 12px; }
	.center-state, .detail-empty { display: grid; place-content: center; gap: 8px; min-height: 220px; padding: 24px; text-align: center; color: var(--dt3); font-size: 13px; }
	.detail-panel { min-width: 0; display: flex; flex-direction: column; background: var(--dbg); }
	.detail-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 14px 18px; border-bottom: 1px solid var(--dbd); }
	.detail-source { display: flex; align-items: center; gap: 8px; color: var(--dt2); font-size: 12px; }
	.text-button { border: 0; background: transparent; color: var(--dt3); font-size: 12px; cursor: pointer; } .text-button:hover { color: #b91c1c; }
	.detail-body { padding: 28px max(28px, 6vw); border-bottom: 1px solid var(--dbd); }
	.detail-body h2 { max-width: 760px; margin: 0 0 14px; font-size: 22px; line-height: 1.28; font-weight: 650; letter-spacing: 0; }
	.person-line, .context-row { display: flex; align-items: center; gap: 7px; color: var(--dt2); font-size: 13px; }
	.message-preview { max-width: 760px; margin: 22px 0 0; color: var(--dt2); font-size: 14px; line-height: 1.6; white-space: pre-wrap; }
	.detail-controls { display: flex; align-items: center; gap: 10px; padding: 14px max(28px, 6vw); border-bottom: 1px solid var(--dbd); }
	.primary-command { display: inline-flex; align-items: center; gap: 7px; height: 34px; padding: 0 11px; border: 0; border-radius: 6px; background: var(--dt); color: var(--dbg); font-size: 13px; font-weight: 600; cursor: pointer; }
	.primary-command:disabled { opacity: .6; cursor: progress; }
	.status-control { display: grid; gap: 5px; color: var(--dt3); font-size: 11px; font-weight: 600; }
	.status-control select { width: 180px; height: 32px; border: 1px solid var(--dbd); border-radius: 6px; padding: 0 8px; background: var(--dbg2); color: var(--dt); font: inherit; font-size: 13px; }
	.context-panel { padding: 20px max(28px, 6vw); display: grid; gap: 11px; } .context-panel .section-label { margin-bottom: 2px; }
	.detail-empty { flex: 1; }
	@media (max-width: 1000px) { .console-grid { grid-template-columns: 185px minmax(250px, 330px) minmax(320px, 1fr); } .detail-body, .detail-controls, .context-panel { padding-left: 24px; padding-right: 24px; } }
	@media (max-width: 760px) { .communications-console { min-height: 0; } .console-grid { grid-template-columns: 170px minmax(0, 1fr); } .detail-panel { display: none; } .console-header { padding: 14px; } }
</style>
