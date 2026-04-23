<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { agents, categoryLabels } from '$lib/stores/agents';
  import type { CustomAgent } from '$lib/api/ai/types';
  import AgentCard from '$lib/components/agents/AgentCard.svelte';
  import CreateAgentModal from '$lib/components/agents/CreateAgentModal.svelte';
  import AgentFilterBar from '$lib/components/agents/AgentFilterBar.svelte';
  import AgentEmptyState from '$lib/components/agents/AgentEmptyState.svelte';
  import { Bot, LayoutGrid, List, AlertCircle } from 'lucide-svelte';

  let searchQuery = $state('');
  let selectedCategory = $state<string | null>(null);
  let selectedStatus = $state<'active' | 'inactive' | null>(null);
  let sortBy = $state<'name' | 'created' | 'usage'>('name');
  let showDeleteDialog = $state<string | null>(null);

  let viewMode = $state<'grid' | 'list'>('grid');

  let selectedIds = $state<Set<string>>(new Set());
  let bulkMode = $state(false);
  let showBulkDeleteConfirm = $state(false);

  // Create agent modal state
  let showCreateModal = $state(false);
  let createName = $state('');
  let createHandle = $state('');
  let createDesc = $state('');
  let createCategory = $state('general');
  let createModel = $state('claude-sonnet-4-6');
  let createPrompt = $state('');
  let createActive = $state(true);
  let createSaving = $state(false);
  let createError = $state<string | null>(null);
  let handleEdited = $state(false);

  const categories = ['general', 'coding', 'writing', 'analysis', 'research', 'support', 'sales', 'marketing'] as const;

  const quickstarts = [
    {
      name: 'Sales Agent',
      desc: 'Qualify leads, draft follow-ups',
      href: '/agents/presets?preset=sales',
      icon: 'briefcase' as const,
    },
    {
      name: 'Code Reviewer',
      desc: 'Review PRs, find bugs',
      href: '/agents/presets?preset=coding',
      icon: 'code' as const,
    },
    {
      name: 'Support Bot',
      desc: 'Answer questions, resolve tickets',
      href: '/agents/presets?preset=support',
      icon: 'headset' as const,
    },
    {
      name: 'Custom Agent',
      desc: 'Build from scratch your way',
      href: '/agents/new',
      icon: 'sparkle' as const,
    },
  ] as const;

  let filteredAgents = $derived.by(() => {
    let filtered = $agents.agents ?? [];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (agent) =>
          (agent.name ?? '').toLowerCase().includes(query) ||
          (agent.display_name ?? '').toLowerCase().includes(query) ||
          (agent.description ?? '').toLowerCase().includes(query)
      );
    }

    if (selectedCategory) {
      filtered = filtered.filter((agent) => agent.category === selectedCategory);
    }

    if (selectedStatus === 'active') {
      filtered = filtered.filter((agent) => agent.is_active);
    } else if (selectedStatus === 'inactive') {
      filtered = filtered.filter((agent) => !agent.is_active);
    }

    return [...filtered].sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return (a.display_name ?? '').localeCompare(b.display_name ?? '');
        case 'created':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        case 'usage':
          return (b.times_used || 0) - (a.times_used || 0);
        default:
          return 0;
      }
    });
  });

  onMount(async () => {
    try {
      await agents.loadAgents();
    } catch {
      // loadAgents handles its own error state
    }
  });

  function handleSearch(e: Event) {
    searchQuery = (e.target as HTMLInputElement).value;
  }

  function clearSearch() {
    searchQuery = '';
  }

  function deriveHandle(name: string): string {
    return name.toLowerCase().replace(/\s+/g, '_').replace(/[^a-z0-9_]/g, '').slice(0, 40);
  }

  function handleNameInput(e: Event) {
    createName = (e.target as HTMLInputElement).value;
    if (!handleEdited) createHandle = deriveHandle(createName);
  }

  function handleHandleInput(e: Event) {
    handleEdited = true;
    createHandle = (e.target as HTMLInputElement).value.toLowerCase().replace(/[^a-z0-9_]/g, '');
  }

  function openCreateModal() {
    createName = ''; createHandle = ''; createDesc = '';
    createCategory = 'general'; createModel = 'claude-sonnet-4-6';
    createPrompt = ''; createActive = true;
    createError = null; createSaving = false; handleEdited = false;
    showCreateModal = true;
  }

  function closeCreateModal() {
    showCreateModal = false;
  }

  async function submitCreateAgent() {
    if (!createName.trim()) { createError = 'Display name is required.'; return; }
    if (!createHandle.trim()) { createError = 'Handle is required.'; return; }
    createSaving = true;
    createError = null;
    try {
      const agent = await agents.createAgent({
        display_name: createName.trim(),
        name: createHandle.trim(),
        description: createDesc.trim() || undefined,
        category: createCategory,
        model_preference: createModel,
        system_prompt: createPrompt.trim() || '',
        is_active: createActive,
      });
      showCreateModal = false;
      goto(`/agents/${agent.id}`);
    } catch (err: unknown) {
      createError = err instanceof Error ? err.message : 'Failed to create agent.';
    } finally {
      createSaving = false;
    }
  }

  function handleChatAgent(agent: CustomAgent) {
    goto(`/chat?agent=${agent.id}`);
  }

  function handleSortChange(e: Event) {
    sortBy = (e.target as HTMLSelectElement).value as 'name' | 'created' | 'usage';
  }

  function handleSelectAgent(agent: CustomAgent) {
    if (bulkMode) {
      toggleSelection(agent.id);
      return;
    }
    goto(`/agents/${agent.id}`);
  }

  function handleEditAgent(agent: CustomAgent) {
    goto(`/agents/${agent.id}/edit`);
  }

  async function handleDeleteAgent(agent: CustomAgent) {
    showDeleteDialog = agent.id;
  }

  async function confirmDelete() {
    if (!showDeleteDialog) return;
    try {
      await agents.deleteAgent(showDeleteDialog);
      showDeleteDialog = null;
    } catch (err) {
      console.error('Failed to delete agent:', err);
    }
  }

  function clearFilters() {
    searchQuery = '';
    selectedCategory = null;
    selectedStatus = null;
    sortBy = 'name';
  }

  const hasActiveFilters = $derived(
    !!searchQuery || !!selectedCategory || selectedStatus !== null || sortBy !== 'name'
  );

  function toggleSelection(id: string) {
    const next = new Set(selectedIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedIds = next;
  }

  function toggleBulkMode() {
    if (bulkMode) {
      bulkMode = false;
      selectedIds = new Set();
    } else {
      bulkMode = true;
    }
  }

  async function bulkActivate() {
    for (const id of selectedIds) {
      await agents.updateAgent(id, { is_active: true });
    }
    selectedIds = new Set();
    await agents.loadAgents().catch(() => {});
  }

  async function bulkDeactivate() {
    for (const id of selectedIds) {
      await agents.updateAgent(id, { is_active: false });
    }
    selectedIds = new Set();
    await agents.loadAgents().catch(() => {});
  }

  async function confirmBulkDelete() {
    for (const id of selectedIds) {
      await agents.deleteAgent(id).catch(() => {});
    }
    selectedIds = new Set();
    showBulkDeleteConfirm = false;
    bulkMode = false;
  }
</script>

<svelte:head>
  <title>Agents — BusinessOS</title>
</svelte:head>

<div class="ag-page">
  <div class="ag-container">

    <!-- Page Header -->
    <header class="ag-header">
      <div class="ag-header__left">
        <div class="ag-header__icon" aria-hidden="true">
          <Bot size={16} strokeWidth={2} />
        </div>
        <div>
          <h1 class="ag-header__title">Agents</h1>
          <p class="ag-header__subtitle">Build and manage your custom AI agents</p>
        </div>
      </div>
      <div class="ag-header__actions">
        <button
          onclick={toggleBulkMode}
          class="btn-pill btn-pill-ghost btn-pill-sm"
          class:btn-pill-active={bulkMode}
        >
          {bulkMode ? 'Cancel' : 'Select'}
        </button>

        <div class="ag-view-toggle" role="group" aria-label="Switch view mode">
          <button
            onclick={() => viewMode = 'grid'}
            class="ag-view-btn"
            class:ag-view-btn--active={viewMode === 'grid'}
            aria-label="Grid view"
            aria-pressed={viewMode === 'grid'}
          >
            <LayoutGrid size={14} strokeWidth={2} aria-hidden="true" />
          </button>
          <button
            onclick={() => viewMode = 'list'}
            class="ag-view-btn"
            class:ag-view-btn--active={viewMode === 'list'}
            aria-label="List view"
            aria-pressed={viewMode === 'list'}
          >
            <List size={14} strokeWidth={2} aria-hidden="true" />
          </button>
        </div>

        <button onclick={() => goto('/agents/presets')} class="btn-pill btn-pill-secondary btn-pill-sm">
          Browse Presets
        </button>
        <button onclick={() => goto('/agents/new')} class="btn-cta">
          + New Agent
        </button>
      </div>
    </header>

    <!-- Filter Bar -->
    <AgentFilterBar
      {searchQuery}
      {selectedCategory}
      {selectedStatus}
      {sortBy}
      {categories}
      {hasActiveFilters}
      onSearchInput={handleSearch}
      onClearSearch={clearSearch}
      onSortChange={handleSortChange}
      onCategoryChange={(cat) => { selectedCategory = cat; }}
      onStatusChange={(status) => { selectedStatus = status; }}
      onClearFilters={clearFilters}
    />

    <!-- Stats ribbon -->
    {#if !$agents.loading && ($agents.agents ?? []).length > 0}
      {@const all = $agents.agents ?? []}
      <div class="ag-stats">
        <span class="ag-stats__item"><strong>{all.length}</strong> agents</span>
        <span class="ag-stats__dot" aria-hidden="true"></span>
        <span class="ag-stats__item"><strong>{all.filter(a => a.is_active).length}</strong> active</span>
        <span class="ag-stats__dot" aria-hidden="true"></span>
        <span class="ag-stats__item"><strong>{all.filter(a => !a.is_active).length}</strong> inactive</span>
        {#if all.some(a => a.is_featured)}
          <span class="ag-stats__dot" aria-hidden="true"></span>
          <span class="ag-stats__item"><strong>{all.filter(a => a.is_featured).length}</strong> featured</span>
        {/if}
      </div>
    {/if}

    <!-- Demo banner -->
    {#if $agents.error === 'demo'}
      <div class="ag-demo" role="status">
        <span class="ag-demo__dot" aria-hidden="true"></span>
        Demo mode — showing sample agents. Connect your backend to see real data.
      </div>
    {/if}

    <!-- Loading skeleton -->
    {#if $agents.loading}
      <div class="ag-grid" aria-busy="true" aria-label="Loading agents">
        {#each Array(6) as _}
          <div class="ag-skel" aria-hidden="true">
            <div class="ag-skel__top">
              <div class="ag-skel__avatar"></div>
              <div class="ag-skel__lines">
                <div class="ag-skel__line" style="width:58%"></div>
                <div class="ag-skel__line" style="width:32%;height:8px;margin-top:2px"></div>
              </div>
            </div>
            <div class="ag-skel__line" style="width:100%"></div>
            <div class="ag-skel__line" style="width:75%"></div>
            <div class="ag-skel__footer">
              <div class="ag-skel__line" style="width:48px;height:18px;border-radius:4px"></div>
              <div class="ag-skel__line" style="width:36px;height:10px"></div>
            </div>
          </div>
        {/each}
      </div>

    <!-- Error state -->
    {:else if $agents.error && $agents.error !== 'demo' && $agents.agents.length === 0}
      <div class="ag-error" role="alert">
        <AlertCircle class="ag-error__icon" size={32} strokeWidth={1.5} aria-hidden="true" />
        <p class="ag-error__title">Could not load agents</p>
        <p class="ag-error__msg">{$agents.error}</p>
        <button onclick={() => agents.loadAgents().catch(() => {})} class="btn-pill btn-pill-secondary btn-pill-sm">
          Retry
        </button>
      </div>

    <!-- Empty with active filters -->
    {:else if filteredAgents.length === 0 && hasActiveFilters}
      <div class="ag-empty">
        <p class="ag-empty__title">No agents match your filters</p>
        <button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-sm">Clear filters</button>
      </div>

    <!-- Empty — no agents at all -->
    {:else if filteredAgents.length === 0}
      <AgentEmptyState
        {quickstarts}
        onNavigate={(href) => goto(href)}
        onNewAgent={() => goto('/agents/new')}
      />

    <!-- Agent grid / list -->
    {:else}
      {#if viewMode === 'grid'}
        <div class="ag-grid">
          {#each filteredAgents as agent (agent.id)}
            <div
              class="ag-card-wrap"
              class:ag-card-wrap--selected={selectedIds.has(agent.id)}
            >
              {#if bulkMode}
                <label class="ag-checkbox" aria-label="Select {agent.display_name}">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(agent.id)}
                    onchange={() => toggleSelection(agent.id)}
                  />
                </label>
              {/if}
              <AgentCard
                {agent}
                onSelect={handleSelectAgent}
                onEdit={bulkMode ? undefined : handleEditAgent}
                onDelete={bulkMode ? undefined : handleDeleteAgent}
                onChat={bulkMode ? undefined : handleChatAgent}
                variant="default"
              />
            </div>
          {/each}
        </div>
      {:else}
        <div class="ag-list">
          {#each filteredAgents as agent (agent.id)}
            <div
              class="ag-list__item"
              class:ag-list__item--selected={selectedIds.has(agent.id)}
            >
              {#if bulkMode}
                <label class="ag-list__checkbox" aria-label="Select {agent.display_name}">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(agent.id)}
                    onchange={() => toggleSelection(agent.id)}
                  />
                </label>
              {/if}
              <div class="ag-list__card">
                <AgentCard
                  {agent}
                  onSelect={handleSelectAgent}
                  onEdit={bulkMode ? undefined : handleEditAgent}
                  onDelete={bulkMode ? undefined : handleDeleteAgent}
                  onChat={bulkMode ? undefined : handleChatAgent}
                  variant="compact"
                />
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <p class="ag-count">
        Showing {filteredAgents.length} {filteredAgents.length === 1 ? 'agent' : 'agents'}
        {#if ($agents.agents?.length ?? 0) !== filteredAgents.length}
          of {$agents.agents?.length ?? 0} total
        {/if}
      </p>
    {/if}

  </div>
</div>

<!-- Create Agent Modal -->
<CreateAgentModal
  open={showCreateModal}
  saving={createSaving}
  name={createName}
  handle={createHandle}
  bind:desc={createDesc}
  bind:category={createCategory}
  bind:model={createModel}
  bind:prompt={createPrompt}
  bind:active={createActive}
  error={createError}
  {categories}
  onclose={closeCreateModal}
  onsubmit={submitCreateAgent}
  onNameInput={handleNameInput}
  onHandleInput={handleHandleInput}
/>

<!-- Floating bulk action bar -->
{#if bulkMode && selectedIds.size > 0}
  <div class="ag-bulk-bar" role="toolbar" aria-label="Bulk actions">
    <span class="ag-bulk-bar__count">{selectedIds.size} selected</span>
    <span class="ag-bulk-bar__sep" aria-hidden="true"></span>
    <button onclick={bulkActivate} class="ag-bulk-btn">Activate</button>
    <button onclick={bulkDeactivate} class="ag-bulk-btn">Deactivate</button>
    <button onclick={() => showBulkDeleteConfirm = true} class="ag-bulk-btn ag-bulk-btn--danger">Delete</button>
  </div>
{/if}

<!-- Single delete confirmation modal -->
{#if showDeleteDialog}
  <div
    class="ag-backdrop"
    onclick={() => showDeleteDialog = null}
    onkeydown={(e) => e.key === 'Escape' && (showDeleteDialog = null)}
    role="presentation"
    aria-hidden="true"
  >
    <div
      class="ag-modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="ag-delete-title"
      tabindex="-1"
    >
      <div class="ag-modal__icon" aria-hidden="true">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6m4-6v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
        </svg>
      </div>
      <h3 class="ag-modal__title" id="ag-delete-title">Delete Agent</h3>
      <p class="ag-modal__body">This agent will be permanently removed. This action cannot be undone.</p>
      <div class="ag-modal__actions">
        <button onclick={() => showDeleteDialog = null} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
        <button onclick={confirmDelete} class="btn-pill btn-pill-danger btn-pill-sm">Delete</button>
      </div>
    </div>
  </div>
{/if}

<!-- Bulk delete confirmation modal -->
{#if showBulkDeleteConfirm}
  <div
    class="ag-backdrop"
    onclick={() => showBulkDeleteConfirm = false}
    onkeydown={(e) => e.key === 'Escape' && (showBulkDeleteConfirm = false)}
    role="presentation"
    aria-hidden="true"
  >
    <div
      class="ag-modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="ag-bulk-delete-title"
      tabindex="-1"
    >
      <div class="ag-modal__icon" aria-hidden="true">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6m4-6v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
        </svg>
      </div>
      <h3 class="ag-modal__title" id="ag-bulk-delete-title">Delete {selectedIds.size} {selectedIds.size === 1 ? 'Agent' : 'Agents'}</h3>
      <p class="ag-modal__body">
        {selectedIds.size === 1 ? 'This agent' : `These ${selectedIds.size} agents`} will be permanently removed. This action cannot be undone.
      </p>
      <div class="ag-modal__actions">
        <button onclick={() => showBulkDeleteConfirm = false} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
        <button onclick={confirmBulkDelete} class="btn-pill btn-pill-danger btn-pill-sm">Delete All</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ═══ Agents List Page — BOS Tokens ═══ */

  .ag-page {
    height: 100%;
    background: var(--dbg2, #f5f5f5);
    overflow-y: auto;
  }

  .ag-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem 1.5rem 3rem;
    width: 100%;
  }

  /* Header */
  .ag-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding-bottom: 1.25rem;
    margin-bottom: 1.25rem;
    border-bottom: 1px solid var(--dbd, #e5e5e5);
  }
  .ag-header__left {
    display: flex;
    align-items: center;
    gap: 0.875rem;
  }
  .ag-header__icon {
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 10px;
    background: var(--dt, #111);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: var(--dbg, #fff);
  }
  .ag-header__title {
    font-size: 1.375rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
    letter-spacing: -0.02em;
  }
  .ag-header__subtitle {
    font-size: 0.8125rem;
    color: var(--dt3, #888);
    margin: 0.1rem 0 0;
  }
  .ag-header__actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
    align-items: center;
  }

  .btn-pill-active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
  }

  /* View toggle */
  .ag-view-toggle {
    display: flex;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 7px;
    overflow: hidden;
  }
  .ag-view-btn {
    width: 32px;
    height: 32px;
    border: none;
    background: var(--dbg, #fff);
    color: var(--dt3, #888);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.12s, color 0.12s;
  }
  .ag-view-btn:hover {
    background: var(--dbg2, #f5f5f5);
    color: var(--dt, #111);
  }
  .ag-view-btn--active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
  }
  .ag-view-btn--active:hover {
    background: var(--dt2, #333);
  }

  /* Stats ribbon */
  .ag-stats {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
    color: var(--dt3, #888);
    margin-bottom: 0.5rem;
  }
  .ag-stats__item strong {
    font-weight: 700;
    color: var(--dt, #111);
    font-family: var(--bos-font-number-family, monospace);
  }
  .ag-stats__dot {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--dbd, #ccc);
    flex-shrink: 0;
  }

  /* Demo banner */
  .ag-demo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.875rem;
    font-size: 0.8125rem;
    color: var(--bos-status-info-text, #2563eb);
    background: var(--bos-status-info-bg, rgba(59,130,246,0.07));
    border: 1px solid rgba(59, 130, 246, 0.18);
    border-radius: 0.5rem;
    margin-bottom: 1rem;
  }
  .ag-demo__dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--bos-accent-blue, #3b82f6);
    flex-shrink: 0;
    animation: ag-pulse 2s ease-in-out infinite;
  }

  /* Skeleton */
  .ag-skel {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 12px;
    padding: 1.125rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    animation: ag-pulse 1.5s ease-in-out infinite;
  }
  .ag-skel__top {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .ag-skel__avatar {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 10px;
    background: var(--dbg3, #eee);
    flex-shrink: 0;
  }
  .ag-skel__lines {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .ag-skel__line {
    height: 10px;
    background: var(--dbg3, #eee);
    border-radius: 3px;
  }
  .ag-skel__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 0.625rem;
    border-top: 1px solid var(--dbd2, #f0f0f0);
  }

  /* Error */
  .ag-error {
    text-align: center;
    padding: 3.5rem 2rem;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }
  :global(.ag-error__icon) { color: var(--bos-status-error, #ef4444); margin-bottom: 0.25rem; }
  .ag-error__title { font-size: 0.9375rem; font-weight: 600; color: var(--dt, #111); margin: 0; }
  .ag-error__msg { font-size: 0.8125rem; color: var(--dt3, #888); margin: 0 0 0.5rem; }

  /* Empty (filter mismatch) */
  .ag-empty {
    text-align: center;
    padding: 4rem 2rem;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }
  .ag-empty__title { font-size: 1rem; font-weight: 600; color: var(--dt, #111); margin: 0; }

  /* Grid view */
  .ag-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1rem;
  }

  /* Card wrapper for bulk select */
  .ag-card-wrap {
    position: relative;
    border-radius: 12px;
  }
  .ag-card-wrap--selected {
    box-shadow: 0 0 0 2px var(--bos-accent-blue, #3b82f6);
    border-radius: 12px;
  }
  .ag-checkbox {
    position: absolute;
    top: 8px;
    left: 8px;
    z-index: 2;
    width: 18px;
    height: 18px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .ag-checkbox input {
    width: 18px;
    height: 18px;
    cursor: pointer;
    accent-color: var(--bos-accent-blue, #3b82f6);
  }

  /* List view */
  .ag-list {
    display: flex;
    flex-direction: column;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 12px;
    overflow: hidden;
  }
  .ag-list__item {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--dbd2, #f0f0f0);
    position: relative;
  }
  .ag-list__item:last-child {
    border-bottom: none;
  }
  .ag-list__item--selected {
    background: rgba(59, 130, 246, 0.04);
  }
  .ag-list__checkbox {
    flex-shrink: 0;
    padding: 0 0 0 0.875rem;
    display: flex;
    align-items: center;
    cursor: pointer;
  }
  .ag-list__checkbox input {
    width: 16px;
    height: 16px;
    cursor: pointer;
    accent-color: var(--bos-accent-blue, #3b82f6);
  }
  .ag-list__card {
    flex: 1;
    min-width: 0;
  }
  .ag-list__item :global(.ac) {
    border: none;
    border-radius: 0;
  }
  .ag-list__item :global(.ac--clickable:hover) {
    transform: none;
    box-shadow: none;
    border-color: transparent;
    background: var(--dbg2, #f5f5f5);
  }

  .ag-count {
    text-align: center;
    font-size: 0.8125rem;
    color: var(--dt4, #bbb);
    margin: 1.5rem 0 0;
  }

  /* Floating bulk action bar */
  .ag-bulk-bar {
    position: fixed;
    bottom: 2rem;
    left: 50%;
    transform: translateX(-50%);
    background: var(--dt, #111);
    color: var(--dbg, #fff);
    border-radius: 9999px;
    padding: 0.625rem 1.25rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
    z-index: 40;
    white-space: nowrap;
  }
  .ag-bulk-bar__count {
    font-size: 0.8125rem;
    font-weight: 600;
  }
  .ag-bulk-bar__sep {
    width: 1px;
    height: 16px;
    background: rgba(255, 255, 255, 0.2);
    flex-shrink: 0;
  }
  .ag-bulk-btn {
    font-size: 0.8125rem;
    font-weight: 500;
    background: none;
    border: none;
    color: var(--dbg, #fff);
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: 5px;
    transition: background 0.12s;
  }
  .ag-bulk-btn:hover {
    background: rgba(255, 255, 255, 0.12);
  }
  .ag-bulk-btn--danger {
    color: #fca5a5;
  }
  .ag-bulk-btn--danger:hover {
    background: rgba(239, 68, 68, 0.18);
    color: #fecaca;
  }

  /* Backdrop + modal */
  .ag-backdrop {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(2px);
  }
  .ag-modal {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 14px;
    padding: 1.75rem;
    max-width: 400px;
    width: calc(100% - 2rem);
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.18);
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.625rem;
  }
  .ag-modal__icon {
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 10px;
    background: var(--bos-status-error-bg, rgba(239,68,68,0.08));
    color: var(--bos-status-error, #ef4444);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.25rem;
  }
  .ag-modal__title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
  }
  .ag-modal__body {
    font-size: 0.875rem;
    color: var(--dt3, #888);
    margin: 0 0 0.625rem;
    line-height: 1.5;
  }
  .ag-modal__actions {
    display: flex;
    gap: 0.5rem;
    justify-content: center;
  }

  @keyframes ag-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.45; }
  }
</style>
