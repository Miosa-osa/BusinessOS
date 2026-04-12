<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { agents, categoryLabels } from '$lib/stores/agents';
  import type { CustomAgent } from '$lib/api/ai/types';
  import AgentCard from '$lib/components/agents/AgentCard.svelte';
  import { Bot, LayoutGrid, List, Search, X, AlertCircle, Briefcase, Code2, Headphones, Sparkles } from 'lucide-svelte';

  let searchQuery = $state('');
  let selectedCategory = $state<string | null>(null);
  let selectedStatus = $state<'active' | 'inactive' | null>(null);
  let sortBy = $state<'name' | 'created' | 'usage'>('name');
  let showDeleteDialog = $state<string | null>(null);

  // Feature 1: view mode toggle
  let viewMode = $state<'grid' | 'list'>('grid');

  // Feature 2: bulk select
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

  const categories = ['general', 'coding', 'writing', 'analysis', 'research', 'support', 'sales', 'marketing'];
  const statusOptions = [
    { value: null, label: 'All' },
    { value: 'active' as const, label: 'Active' },
    { value: 'inactive' as const, label: 'Inactive' }
  ];

  // Feature 3: quickstart cards for empty state
  const quickstarts = [
    {
      name: 'Sales Agent',
      desc: 'Qualify leads, draft follow-ups',
      href: '/agents/presets?preset=sales',
      icon: 'briefcase'
    },
    {
      name: 'Code Reviewer',
      desc: 'Review PRs, find bugs',
      href: '/agents/presets?preset=coding',
      icon: 'code'
    },
    {
      name: 'Support Bot',
      desc: 'Answer questions, resolve tickets',
      href: '/agents/presets?preset=support',
      icon: 'headset'
    },
    {
      name: 'Custom Agent',
      desc: 'Build from scratch your way',
      href: '/agents/new',
      icon: 'sparkle'
    }
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

  // Auto-derive handle from display name (slug)
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

  // Bulk select helpers
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
        <!-- Bulk select toggle -->
        <button
          onclick={toggleBulkMode}
          class="btn-pill btn-pill-ghost btn-pill-sm"
          class:btn-pill-active={bulkMode}
        >
          {bulkMode ? 'Cancel' : 'Select'}
        </button>

        <!-- View mode toggle (Feature 1) -->
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
    <div class="ag-filters">
      <!-- Top row: search + sort -->
      <div class="ag-filters__row">
        <div class="ag-search">
          <Search class="ag-search__icon" size={14} strokeWidth={2} aria-hidden="true" />
          <input
            type="text"
            value={searchQuery}
            oninput={handleSearch}
            placeholder="Search agents..."
            class="ag-search__input"
            aria-label="Search agents"
          />
          {#if searchQuery}
            <button class="ag-search__clear" onclick={clearSearch} aria-label="Clear search" tabindex="-1">
              <X size={12} strokeWidth={2} aria-hidden="true" />
            </button>
          {/if}
        </div>
        <select value={sortBy} onchange={handleSortChange} class="ag-sort" aria-label="Sort agents">
          <option value="name">Name A–Z</option>
          <option value="created">Newest first</option>
          <option value="usage">Most used</option>
        </select>
      </div>

      <!-- Combined filter pills row -->
      <div class="ag-filters__pills-row">
        <!-- Category group -->
        <div class="ag-pill-group" role="group" aria-label="Filter by category">
          <button
            onclick={() => selectedCategory = null}
            class="ag-pill"
            class:ag-pill--active={selectedCategory === null}
          >All</button>
          {#each categories as cat}
            <button
              onclick={() => selectedCategory = selectedCategory === cat ? null : cat}
              class="ag-pill"
              class:ag-pill--active={selectedCategory === cat}
            >{categoryLabels[cat] || cat}</button>
          {/each}
        </div>

        <!-- Divider -->
        <span class="ag-pill-divider" aria-hidden="true"></span>

        <!-- Status group -->
        <div class="ag-pill-group" role="group" aria-label="Filter by status">
          {#each statusOptions as opt}
            <button
              onclick={() => selectedStatus = opt.value}
              class="ag-pill ag-pill--status"
              class:ag-pill--active={selectedStatus === opt.value}
            >{opt.label}</button>
          {/each}
        </div>

        <!-- Clear chip — only when filters active -->
        {#if hasActiveFilters}
          <button onclick={clearFilters} class="ag-pill ag-pill--clear" aria-label="Clear all filters">
            <X size={10} strokeWidth={2} aria-hidden="true" />
            Clear
          </button>
        {/if}
      </div>
    </div>

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

    <!-- Empty — no agents at all (Feature 3) -->
    {:else if filteredAgents.length === 0}
      <div class="ag-empty-full">
        <div class="ag-empty-full__icon" aria-hidden="true">
          <Bot size={28} strokeWidth={1.5} />
        </div>
        <h2 class="ag-empty-full__title">No agents yet</h2>
        <p class="ag-empty-full__subtitle">Create your first custom AI agent or start from a template</p>

        <div class="ag-quickstart">
          {#each quickstarts as qs}
            <button class="ag-qs" onclick={() => goto(qs.href)}>
              <div class="ag-qs__icon" aria-hidden="true">
                {#if qs.icon === 'briefcase'}
                  <Briefcase size={16} strokeWidth={2} />
                {:else if qs.icon === 'code'}
                  <Code2 size={16} strokeWidth={2} />
                {:else if qs.icon === 'headset'}
                  <Headphones size={16} strokeWidth={2} />
                {:else if qs.icon === 'sparkle'}
                  <Sparkles size={16} strokeWidth={2} />
                {/if}
              </div>
              <p class="ag-qs__name">{qs.name}</p>
              <p class="ag-qs__desc">{qs.desc}</p>
            </button>
          {/each}
        </div>

        <p class="ag-empty-full__or">or</p>
        <button onclick={() => goto('/agents/new')} class="btn-cta">Start from Scratch</button>
      </div>

    <!-- Agent grid / list (Feature 1 + Feature 2) -->
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
{#if showCreateModal}
  <div
    class="ag-backdrop"
    onclick={closeCreateModal}
    onkeydown={(e) => e.key === 'Escape' && closeCreateModal()}
    role="presentation"
    aria-hidden="true"
  >
    <div
      class="ag-create-modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="ag-create-title"
      tabindex="-1"
    >
      <!-- Modal header -->
      <div class="ag-cm__header">
        <div class="ag-cm__header-icon" aria-hidden="true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="8" width="18" height="12" rx="3"/>
            <path d="M9 8V6a3 3 0 0 1 6 0v2"/>
            <circle cx="9" cy="13" r="1.5" fill="currentColor" stroke="none"/>
            <circle cx="15" cy="13" r="1.5" fill="currentColor" stroke="none"/>
            <path d="M9 17h6"/>
          </svg>
        </div>
        <div>
          <h2 class="ag-cm__title" id="ag-create-title">New Agent</h2>
          <p class="ag-cm__subtitle">Configure your custom AI agent</p>
        </div>
        <button class="ag-cm__close" onclick={closeCreateModal} aria-label="Close">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true">
            <path d="M18 6 6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- Form -->
      <div class="ag-cm__body">
        <!-- Row 1: Name + Handle -->
        <div class="ag-cm__row">
          <div class="ag-cm__field ag-cm__field--grow">
            <label class="ag-cm__label" for="ag-create-name">Display Name <span class="ag-cm__req">*</span></label>
            <input
              id="ag-create-name"
              type="text"
              class="ag-cm__input"
              value={createName}
              oninput={handleNameInput}
              placeholder="e.g. Sales Assistant"
              autocomplete="off"
            />
          </div>
          <div class="ag-cm__field">
            <label class="ag-cm__label" for="ag-create-handle">Handle <span class="ag-cm__req">*</span></label>
            <div class="ag-cm__input-prefix">
              <span class="ag-cm__prefix">@</span>
              <input
                id="ag-create-handle"
                type="text"
                class="ag-cm__input ag-cm__input--prefixed"
                value={createHandle}
                oninput={handleHandleInput}
                placeholder="sales_assistant"
                autocomplete="off"
                spellcheck="false"
              />
            </div>
          </div>
        </div>

        <!-- Description -->
        <div class="ag-cm__field">
          <label class="ag-cm__label" for="ag-create-desc">Description</label>
          <input
            id="ag-create-desc"
            type="text"
            class="ag-cm__input"
            bind:value={createDesc}
            placeholder="What does this agent do?"
            autocomplete="off"
          />
        </div>

        <!-- Row 2: Category + Model -->
        <div class="ag-cm__row">
          <div class="ag-cm__field ag-cm__field--grow">
            <label class="ag-cm__label" for="ag-create-category">Category</label>
            <select id="ag-create-category" class="ag-cm__select" bind:value={createCategory}>
              {#each categories as cat}
                <option value={cat}>{categoryLabels[cat] || cat}</option>
              {/each}
            </select>
          </div>
          <div class="ag-cm__field ag-cm__field--grow">
            <label class="ag-cm__label" for="ag-create-model">Model</label>
            <select id="ag-create-model" class="ag-cm__select" bind:value={createModel}>
              <option value="claude-sonnet-4-6">Claude Sonnet 4.6</option>
              <option value="claude-opus-4-6">Claude Opus 4.6</option>
              <option value="claude-haiku-4-5-20251001">Claude Haiku 4.5</option>
            </select>
          </div>
        </div>

        <!-- System prompt -->
        <div class="ag-cm__field">
          <label class="ag-cm__label" for="ag-create-prompt">System Prompt</label>
          <textarea
            id="ag-create-prompt"
            class="ag-cm__textarea"
            bind:value={createPrompt}
            placeholder="You are a helpful assistant that..."
            rows="5"
          ></textarea>
          <p class="ag-cm__hint">Defines how your agent behaves. You can refine this later.</p>
        </div>

        <!-- Active toggle -->
        <div class="ag-cm__toggle-row">
          <div>
            <p class="ag-cm__toggle-label">Active</p>
            <p class="ag-cm__toggle-hint">Inactive agents won't appear in chat</p>
          </div>
          <button
            class="ag-cm__toggle"
            class:ag-cm__toggle--on={createActive}
            onclick={() => createActive = !createActive}
            role="switch"
            aria-checked={createActive}
            aria-label="Toggle active status"
          >
            <span class="ag-cm__toggle-thumb"></span>
          </button>
        </div>

        {#if createError}
          <p class="ag-cm__error" role="alert">{createError}</p>
        {/if}
      </div>

      <!-- Footer -->
      <div class="ag-cm__footer">
        <button onclick={closeCreateModal} class="btn-pill btn-pill-ghost btn-pill-sm" disabled={createSaving}>
          Cancel
        </button>
        <button onclick={submitCreateAgent} class="btn-cta" disabled={createSaving || !createName.trim() || !createHandle.trim()}>
          {createSaving ? 'Creating…' : 'Create Agent'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Floating bulk action bar (Feature 2) -->
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

<!-- Bulk delete confirmation modal (Feature 2) -->
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

  /* Bulk mode active pill state */
  .btn-pill-active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
  }

  /* View toggle (Feature 1) */
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

  /* Filter bar */
  .ag-filters {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    padding: 0.75rem 1rem;
    margin-bottom: 0.625rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .ag-filters__row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  /* Search */
  .ag-search {
    flex: 1;
    position: relative;
  }
  .ag-search__icon {
    position: absolute;
    left: 0.75rem;
    top: 50%;
    transform: translateY(-50%);
    color: var(--dt4, #bbb);
    pointer-events: none;
  }
  .ag-search__input {
    width: 100%;
    padding: 0.5625rem 2rem 0.5625rem 2.125rem;
    font-size: 0.875rem;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.5rem;
    background: var(--dbg, #fff);
    color: var(--dt, #111);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
    box-sizing: border-box;
  }
  .ag-search__input::placeholder { color: var(--dt4, #bbb); }
  .ag-search__input:focus {
    border-color: var(--dt2, #555);
    box-shadow: 0 0 0 3px rgba(0,0,0,0.06);
  }
  .ag-search__clear {
    position: absolute;
    right: 0.6rem;
    top: 50%;
    transform: translateY(-50%);
    width: 20px;
    height: 20px;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: var(--dt3, #888);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.1s, color 0.1s;
  }
  .ag-search__clear:hover {
    background: var(--dbg2, #f5f5f5);
    color: var(--dt, #111);
  }

  /* Sort */
  .ag-sort {
    padding: 0.5rem 2rem 0.5rem 0.75rem;
    font-size: 0.8125rem;
    font-family: inherit;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.5rem;
    background-color: var(--dbg, #fff);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888888' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.625rem center;
    color: var(--dt, #111);
    outline: none;
    cursor: pointer;
    white-space: nowrap;
    appearance: none;
    -webkit-appearance: none;
    transition: border-color 0.15s;
  }
  .ag-sort:focus { border-color: var(--dt2, #555); }

  /* Combined pills row */
  .ag-filters__pills-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.3rem;
  }
  .ag-pill-group {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }
  .ag-pill-divider {
    width: 1px;
    height: 16px;
    background: var(--dbd, #e0e0e0);
    flex-shrink: 0;
    margin: 0 0.1rem;
    align-self: center;
  }

  /* Pills */
  .ag-pill {
    padding: 0.25rem 0.6875rem;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: 9999px;
    border: 1px solid var(--dbd, #e0e0e0);
    cursor: pointer;
    transition: all 0.12s;
    background: var(--dbg, #fff);
    color: var(--dt2, #555);
    line-height: 1.5;
    white-space: nowrap;
  }
  .ag-pill:hover {
    background: var(--dbg2, #f5f5f5);
    border-color: var(--dt3, #999);
    color: var(--dt, #111);
  }
  .ag-pill--active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
    font-weight: 600;
    border-color: var(--dt, #111);
  }
  .ag-pill--active:hover {
    background: var(--dt2, #333);
    border-color: var(--dt2, #333);
  }
  .ag-pill--status {
    color: var(--dt3, #888);
  }
  .ag-pill--status.ag-pill--active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
  }
  /* Clear chip */
  .ag-pill--clear {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--dt3, #888);
    border-color: transparent;
    background: transparent;
    margin-left: 0.1rem;
  }
  .ag-pill--clear:hover {
    background: rgba(239, 68, 68, 0.07);
    border-color: rgba(239, 68, 68, 0.25);
    color: var(--bos-status-error, #ef4444);
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
  .ag-error__icon { color: var(--bos-status-error, #ef4444); margin-bottom: 0.25rem; }
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

  /* Full empty state with quickstarts (Feature 3) */
  .ag-empty-full {
    text-align: center;
    padding: 3.5rem 2rem;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }
  .ag-empty-full__icon {
    width: 3.5rem;
    height: 3.5rem;
    border-radius: 14px;
    background: var(--dbg2, #f5f5f5);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--dt3, #888);
    margin-bottom: 0.25rem;
  }
  .ag-empty-full__title {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
  }
  .ag-empty-full__subtitle {
    font-size: 0.875rem;
    color: var(--dt3, #888);
    margin: 0 0 0.5rem;
    max-width: 340px;
  }
  .ag-empty-full__or {
    font-size: 0.8125rem;
    color: var(--dt4, #bbb);
    margin: 0;
  }

  .ag-quickstart {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
    width: 100%;
    max-width: 480px;
  }
  .ag-qs {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 12px;
    padding: 1.25rem;
    cursor: pointer;
    text-align: left;
    transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }
  .ag-qs:hover {
    border-color: var(--bos-accent-blue, #3b82f6);
    box-shadow: 0 4px 16px rgba(59, 130, 246, 0.08);
    transform: translateY(-1px);
  }
  .ag-qs__icon {
    width: 2rem;
    height: 2rem;
    border-radius: 8px;
    background: var(--dbg2, #f5f5f5);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--dt2, #555);
    margin-bottom: 0.25rem;
  }
  .ag-qs__name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--dt, #111);
    margin: 0;
  }
  .ag-qs__desc {
    font-size: 0.75rem;
    color: var(--dt3, #888);
    margin: 0;
  }

  /* Grid view */
  .ag-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1rem;
  }

  /* Card wrapper for bulk select (Feature 2) */
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

  /* List view (Feature 1) */
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
  /* Override card radius/border in list mode */
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

  /* Floating bulk action bar (Feature 2) */
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

  /* ═══ Create Agent Modal ═══ */
  .ag-create-modal {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.16), 0 4px 16px rgba(0,0,0,0.06);
    width: 560px;
    max-width: calc(100vw - 2rem);
    max-height: calc(100vh - 4rem);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: ag-modal-in 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }
  @keyframes ag-modal-in {
    from { opacity: 0; transform: scale(0.96) translateY(4px); }
    to   { opacity: 1; transform: scale(1)    translateY(0); }
  }

  /* Modal header */
  .ag-cm__header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.25rem 1.375rem 1rem;
    border-bottom: 1px solid var(--dbd2, #f0f0f0);
  }
  .ag-cm__header-icon {
    width: 2rem;
    height: 2rem;
    border-radius: 8px;
    background: var(--dbg2, #f5f5f5);
    border: 1px solid var(--dbd, #e0e0e0);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--dt2, #555);
    flex-shrink: 0;
  }
  .ag-cm__title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
    letter-spacing: -0.01em;
  }
  .ag-cm__subtitle {
    font-size: 0.75rem;
    color: var(--dt3, #888);
    margin: 0.1rem 0 0;
  }
  .ag-cm__close {
    margin-left: auto;
    width: 28px;
    height: 28px;
    border-radius: 7px;
    border: none;
    background: transparent;
    color: var(--dt3, #888);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.12s, color 0.12s;
    flex-shrink: 0;
  }
  .ag-cm__close:hover { background: var(--dbg2, #f5f5f5); color: var(--dt, #111); }

  /* Modal body */
  .ag-cm__body {
    padding: 1.125rem 1.375rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    overflow-y: auto;
    flex: 1;
  }
  .ag-cm__row {
    display: flex;
    gap: 0.75rem;
  }
  .ag-cm__field {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .ag-cm__field--grow { flex: 1; min-width: 0; }
  .ag-cm__label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--dt2, #444);
    letter-spacing: 0.01em;
  }
  .ag-cm__req { color: var(--bos-status-error, #ef4444); }
  .ag-cm__input,
  .ag-cm__select,
  .ag-cm__textarea {
    padding: 0.5625rem 0.75rem;
    font-size: 0.875rem;
    font-family: inherit;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 8px;
    background: var(--dbg, #fff);
    color: var(--dt, #111);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
    width: 100%;
    box-sizing: border-box;
  }
  .ag-cm__input::placeholder,
  .ag-cm__textarea::placeholder { color: var(--dt4, #bbb); }
  .ag-cm__input:focus,
  .ag-cm__select:focus,
  .ag-cm__textarea:focus {
    border-color: var(--dt2, #555);
    box-shadow: 0 0 0 3px rgba(0,0,0,0.06);
  }
  .ag-cm__select {
    appearance: none;
    -webkit-appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888888' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.625rem center;
    padding-right: 2rem;
    cursor: pointer;
  }
  .ag-cm__textarea {
    resize: vertical;
    min-height: 100px;
    line-height: 1.55;
  }
  .ag-cm__hint {
    font-size: 0.71875rem;
    color: var(--dt4, #bbb);
    margin: 0;
  }

  /* Handle prefix */
  .ag-cm__input-prefix {
    position: relative;
    display: flex;
    align-items: center;
  }
  .ag-cm__prefix {
    position: absolute;
    left: 0.75rem;
    font-size: 0.875rem;
    color: var(--dt3, #888);
    pointer-events: none;
    user-select: none;
  }
  .ag-cm__input--prefixed { padding-left: 1.375rem; }

  /* Toggle row */
  .ag-cm__toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 0.875rem;
    background: var(--dbg2, #f9f9f9);
    border: 1px solid var(--dbd2, #f0f0f0);
    border-radius: 8px;
  }
  .ag-cm__toggle-label {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--dt, #111);
    margin: 0;
  }
  .ag-cm__toggle-hint {
    font-size: 0.71875rem;
    color: var(--dt3, #888);
    margin: 0.1rem 0 0;
  }
  .ag-cm__toggle {
    width: 36px;
    height: 20px;
    border-radius: 9999px;
    border: none;
    background: var(--dbd, #ddd);
    cursor: pointer;
    position: relative;
    transition: background 0.18s;
    flex-shrink: 0;
  }
  .ag-cm__toggle--on { background: var(--dt, #111); }
  .ag-cm__toggle-thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0,0,0,0.18);
    transition: transform 0.18s;
  }
  .ag-cm__toggle--on .ag-cm__toggle-thumb { transform: translateX(16px); }

  /* Error */
  .ag-cm__error {
    font-size: 0.8125rem;
    color: var(--bos-status-error, #ef4444);
    background: rgba(239, 68, 68, 0.06);
    border: 1px solid rgba(239, 68, 68, 0.18);
    border-radius: 7px;
    padding: 0.5rem 0.75rem;
    margin: 0;
  }

  /* Modal footer */
  .ag-cm__footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.875rem 1.375rem;
    border-top: 1px solid var(--dbd2, #f0f0f0);
    background: var(--dbg2, #fafafa);
  }
  .ag-cm__footer .btn-cta:disabled,
  .ag-cm__footer .btn-pill:disabled {
    opacity: 0.45;
    cursor: not-allowed;
    pointer-events: none;
  }
</style>
