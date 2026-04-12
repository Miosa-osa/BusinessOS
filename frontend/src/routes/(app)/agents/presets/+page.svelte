<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { agents } from '$lib/stores/agents';
  import type { AgentPreset } from '$lib/api/ai/types';
  import PresetCard from '$lib/components/agents/PresetCard.svelte';

  let loading = $state(true);
  let error = $state<string | null>(null);
  let selectedCategory = $state('all');
  let searchQuery = $state('');
  let selectedPreset = $state<AgentPreset | null>(null);
  let showModal = $state(false);
  let customName = $state('');
  let creating = $state(false);
  let createError = $state<string | null>(null);

  const categories = [
    { id: 'all', label: 'All' },
    { id: 'business', label: 'Business' },
    { id: 'creative', label: 'Creative' },
    { id: 'technical', label: 'Technical' },
    { id: 'research', label: 'Research' },
    { id: 'support', label: 'Support' }
  ];

  let allPresets = $derived($agents.presets ?? []);

  let filteredPresets = $derived.by(() => {
    let result = allPresets;
    if (selectedCategory !== 'all') {
      result = result.filter((p) => p.category === selectedCategory);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          (p.display_name ?? '').toLowerCase().includes(q) ||
          p.description.toLowerCase().includes(q)
      );
    }
    return result;
  });

  let featuredPresets = $derived(filteredPresets.filter((p) => p.is_featured));
  let regularPresets = $derived(filteredPresets.filter((p) => !p.is_featured));
  let showFeatured = $derived(featuredPresets.length > 0 && selectedCategory === 'all' && !searchQuery.trim());

  function getCategoryCount(id: string): number {
    if (id === 'all') return allPresets.length;
    return allPresets.filter((p) => p.category === id).length;
  }

  async function loadPresets() {
    loading = true;
    error = null;
    try {
      await agents.loadPresets();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load presets';
    } finally {
      loading = false;
    }
  }

  function handleUsePreset(preset: AgentPreset) {
    selectedPreset = preset;
    customName = preset.display_name ?? preset.name;
    createError = null;
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    selectedPreset = null;
    customName = '';
    creating = false;
    createError = null;
  }

  async function handleCreateAgent() {
    if (!selectedPreset) return;
    creating = true;
    createError = null;
    try {
      const newAgent = await agents.createFromPreset(
        selectedPreset.id,
        customName.trim() || undefined
      );
      if (!newAgent?.id) throw new Error('Agent created but no ID returned');
      closeModal();
      await goto(`/agents/${newAgent.id}/edit`);
    } catch (err) {
      createError = err instanceof Error ? err.message : 'Failed to create agent';
    } finally {
      creating = false;
    }
  }

  onMount(() => { loadPresets(); });
</script>

<svelte:head>
  <title>Agent Presets — BusinessOS</title>
</svelte:head>

<div class="agp-page">
  <div class="agp-container">

    <!-- Header -->
    <header class="agp-header">
      <div class="agp-header__left">
        <a href="/agents" class="agp-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
          Agents
        </a>
        <h1 class="agp-header__title">Agent Presets</h1>
        <p class="agp-header__subtitle">Start from a pre-built template and customize to your needs</p>
      </div>
    </header>

    <!-- Search + filter bar -->
    <div class="agp-toolbar">
      <div class="agp-search">
        <svg class="agp-search__icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>
        </svg>
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search presets..."
          class="agp-search__input"
          aria-label="Search presets"
        />
        {#if searchQuery}
          <button
            onclick={() => searchQuery = ''}
            class="agp-search__clear"
            aria-label="Clear search"
          >
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><path d="M18 6 6 18M6 6l12 12"/></svg>
          </button>
        {/if}
      </div>

      <div class="agp-cats" role="group" aria-label="Filter by category">
        {#each categories as cat}
          <button
            class="agp-cat"
            class:agp-cat--active={selectedCategory === cat.id}
            onclick={() => selectedCategory = cat.id}
          >
            {cat.label}
            <span class="agp-cat__count">{getCategoryCount(cat.id)}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- Loading skeleton -->
    {#if loading}
      <div class="agp-grid" aria-busy="true" aria-label="Loading presets">
        {#each Array(6) as _}
          <div class="agp-skel" aria-hidden="true">
            <div class="agp-skel__top">
              <div class="agp-skel__icon"></div>
              <div class="agp-skel__lines">
                <div class="agp-skel__line" style="width:55%"></div>
                <div class="agp-skel__line" style="width:30%;height:8px;margin-top:2px"></div>
              </div>
            </div>
            <div class="agp-skel__line" style="width:100%"></div>
            <div class="agp-skel__line" style="width:80%"></div>
            <div class="agp-skel__line" style="width:65%"></div>
            <div class="agp-skel__footer">
              <div class="agp-skel__line" style="width:100%;height:32px;border-radius:6px"></div>
            </div>
          </div>
        {/each}
      </div>

    <!-- Error -->
    {:else if error && allPresets.length === 0}
      <div class="agp-error" role="alert">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/>
        </svg>
        <p class="agp-error__title">Failed to load presets</p>
        <p class="agp-error__msg">{error}</p>
        <button onclick={loadPresets} class="btn-pill btn-pill-secondary btn-pill-sm">Retry</button>
      </div>

    <!-- Empty -->
    {:else if filteredPresets.length === 0}
      <div class="agp-empty">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>
        </svg>
        <p class="agp-empty__title">No presets found</p>
        <p class="agp-empty__msg">
          {#if searchQuery}
            No presets match "{searchQuery}"
          {:else if selectedCategory !== 'all'}
            No presets in this category
          {:else}
            No presets available
          {/if}
        </p>
        {#if searchQuery || selectedCategory !== 'all'}
          <button
            onclick={() => { searchQuery = ''; selectedCategory = 'all'; }}
            class="btn-pill btn-pill-ghost btn-pill-sm"
          >Clear filters</button>
        {/if}
      </div>

    <!-- Preset content -->
    {:else}
      {#if showFeatured}
        <section class="agp-section">
          <div class="agp-section__header">
            <svg class="agp-section__star" width="15" height="15" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
            </svg>
            <h2 class="agp-section__title">Featured Templates</h2>
          </div>
          <div class="agp-grid">
            {#each featuredPresets as preset (preset.id)}
              <div class="agp-featured-wrap">
                <span class="agp-featured-badge" aria-label="Featured">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
                  </svg>
                  Featured
                </span>
                <PresetCard {preset} onUse={handleUsePreset} />
              </div>
            {/each}
          </div>
        </section>
      {/if}

      {#if regularPresets.length > 0}
        <section class="agp-section">
          {#if showFeatured}
            <h2 class="agp-section__title agp-section__title--plain">All Templates</h2>
          {/if}
          <div class="agp-grid">
            {#each regularPresets as preset (preset.id)}
              <PresetCard {preset} onUse={handleUsePreset} />
            {/each}
          </div>
        </section>
      {/if}

      <p class="agp-count">
        {filteredPresets.length} {filteredPresets.length === 1 ? 'preset' : 'presets'}
        {#if allPresets.length !== filteredPresets.length}
          of {allPresets.length} total
        {/if}
      </p>
    {/if}

  </div>
</div>

<!-- Preset side drawer -->
{#if showModal && selectedPreset}
  <!-- Backdrop (click to close) -->
  <div
    class="agp-overlay"
    onclick={closeModal}
    onkeydown={(e) => e.key === 'Escape' && closeModal()}
    role="presentation"
    aria-hidden="true"
  ></div>

  <!-- Drawer panel -->
  <div
    class="agp-drawer"
    role="dialog"
    aria-modal="true"
    aria-labelledby="agp-drawer-title"
    tabindex="-1"
  >
    <!-- Drawer header -->
    <div class="agp-drawer__header">
      <div class="agp-drawer__preset-top">
        <div class="agp-drawer__avatar" aria-hidden="true">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
        </div>
        <div>
          <h2 class="agp-drawer__title" id="agp-drawer-title">{selectedPreset.display_name ?? selectedPreset.name}</h2>
          {#if selectedPreset.category}
            <span class="agp-drawer__cat">{selectedPreset.category}</span>
          {/if}
        </div>
      </div>
      <button onclick={closeModal} class="agp-drawer__close" aria-label="Close" disabled={creating}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><path d="M18 6 6 18M6 6l12 12"/></svg>
      </button>
    </div>

    <!-- Drawer body (scrollable) -->
    <div class="agp-drawer__body">
      <p class="agp-drawer__desc">{selectedPreset.description}</p>

      <!-- Model recommendation -->
      {#if selectedPreset.model_preference}
        <div class="agp-drawer__section">
          <p class="agp-drawer__section-label">Recommended Model</p>
          <span class="agp-drawer__mono">{selectedPreset.model_preference}</span>
        </div>
      {/if}

      <!-- Capabilities -->
      {#if selectedPreset.capabilities && selectedPreset.capabilities.length > 0}
        <div class="agp-drawer__section">
          <p class="agp-drawer__section-label">Capabilities</p>
          <div class="agp-drawer__chips">
            {#each selectedPreset.capabilities as cap}
              <span class="agp-drawer__chip">{cap}</span>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Tags -->
      {#if selectedPreset.tags && selectedPreset.tags.length > 0}
        <div class="agp-drawer__section">
          <p class="agp-drawer__section-label">Tags</p>
          <div class="agp-drawer__chips">
            {#each selectedPreset.tags as tag}
              <span class="agp-drawer__chip agp-drawer__chip--tag">{tag}</span>
            {/each}
          </div>
        </div>
      {/if}

      <!-- System prompt (full — key advantage over modal) -->
      {#if selectedPreset.system_prompt}
        <div class="agp-drawer__section">
          <p class="agp-drawer__section-label">System Prompt</p>
          <pre class="agp-drawer__prompt">{selectedPreset.system_prompt}</pre>
        </div>
      {/if}

      <!-- Name input -->
      <div class="agp-drawer__section">
        <label for="agp-drawer-name" class="agp-drawer__section-label">Agent Name</label>
        <input
          id="agp-drawer-name"
          type="text"
          bind:value={customName}
          placeholder={selectedPreset.display_name ?? selectedPreset.name}
          disabled={creating}
          class="agp-drawer__input"
        />
        <p class="agp-drawer__hint">You can change this any time after creation</p>
      </div>

      {#if createError}
        <div class="agp-drawer__error" role="alert">{createError}</div>
      {/if}
    </div>

    <!-- Drawer footer (sticky) -->
    <div class="agp-drawer__footer">
      <button onclick={closeModal} disabled={creating} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
      <button
        onclick={handleCreateAgent}
        disabled={creating}
        class="btn-cta agp-drawer__create-btn"
        aria-busy={creating}
      >
        {#if creating}
          <span class="agp-spinner" aria-hidden="true"></span>
          Creating…
        {:else}
          Create Agent
        {/if}
      </button>
    </div>
  </div>
{/if}

<style>
  /* ═══ Agent Presets Page — BOS Tokens ═══ */

  .agp-page {
    min-height: 100%;
    background: var(--dbg2, #f5f5f5);
  }

  .agp-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.75rem 1.5rem 3rem;
  }

  /* Header */
  .agp-header { margin-bottom: 1.5rem; }
  .agp-back {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.8125rem;
    color: var(--dt3, #888);
    text-decoration: none;
    margin-bottom: 0.875rem;
    transition: color 0.12s;
  }
  .agp-back:hover { color: var(--dt, #111); }
  .agp-header__title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
    letter-spacing: -0.02em;
  }
  .agp-header__subtitle {
    font-size: 0.875rem;
    color: var(--dt3, #888);
    margin: 0.2rem 0 0;
  }

  /* Toolbar */
  .agp-toolbar {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    padding: 1rem 1.125rem;
    margin-bottom: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  /* Search */
  .agp-search {
    position: relative;
    max-width: 440px;
  }
  .agp-search__icon {
    position: absolute;
    left: 0.75rem;
    top: 50%;
    transform: translateY(-50%);
    color: var(--dt3, #888);
    pointer-events: none;
  }
  .agp-search__input {
    width: 100%;
    padding: 0.5rem 2.25rem 0.5rem 2.25rem;
    font-size: 0.875rem;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.5rem;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt, #111);
    outline: none;
    transition: border-color 0.15s, background 0.15s;
    box-sizing: border-box;
  }
  .agp-search__input::placeholder { color: var(--dt4, #bbb); }
  .agp-search__input:focus {
    border-color: var(--dt, #111);
    background: var(--dbg, #fff);
  }
  .agp-search__clear {
    position: absolute;
    right: 0.625rem;
    top: 50%;
    transform: translateY(-50%);
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--dt3, #888);
    border-radius: 4px;
    transition: color 0.12s, background 0.12s;
  }
  .agp-search__clear:hover { color: var(--dt, #111); background: var(--dbg3, #eee); }

  /* Category pills */
  .agp-cats { display: flex; flex-wrap: wrap; gap: 0.3125rem; }
  .agp-cat {
    display: inline-flex;
    align-items: center;
    gap: 0.3125rem;
    padding: 0.2rem 0.6875rem;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: 9999px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.12s;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt2, #555);
  }
  .agp-cat:hover { background: var(--dbg3, #eee); color: var(--dt, #111); }
  .agp-cat--active {
    background: var(--dt, #111);
    color: var(--dbg, #fff);
    font-weight: 600;
    border-color: var(--dt, #111);
  }
  .agp-cat--active:hover { background: var(--dt2, #333); border-color: var(--dt2, #333); }
  .agp-cat__count {
    font-size: 0.6875rem;
    opacity: 0.65;
    font-family: var(--bos-font-number-family, monospace);
  }

  /* Sections */
  .agp-section { margin-bottom: 2rem; }
  .agp-section__header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  .agp-section__star { color: #ca8a04; flex-shrink: 0; }
  .agp-section__title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
  }
  .agp-section__title--plain { margin-bottom: 1rem; }

  /* Featured wrapper */
  .agp-featured-wrap { position: relative; }
  .agp-featured-badge {
    position: absolute;
    top: -8px;
    right: -6px;
    z-index: 2;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.15rem 0.5rem;
    font-size: 0.625rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    background: #ca8a04;
    color: #fff;
    border-radius: 9999px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15);
  }

  /* Grid */
  .agp-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
  }

  /* Skeleton */
  .agp-skel {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 12px;
    padding: 1.125rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    animation: agp-pulse 1.5s ease-in-out infinite;
  }
  .agp-skel__top {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .agp-skel__icon {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 10px;
    background: var(--dbg3, #eee);
    flex-shrink: 0;
  }
  .agp-skel__lines { flex: 1; display: flex; flex-direction: column; gap: 6px; }
  .agp-skel__line { height: 10px; background: var(--dbg3, #eee); border-radius: 3px; }
  .agp-skel__footer { padding-top: 0.5rem; }

  /* Error */
  .agp-error {
    text-align: center;
    padding: 3.5rem 2rem;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    color: var(--bos-status-error, #ef4444);
  }
  .agp-error__title { font-size: 0.9375rem; font-weight: 600; color: var(--dt, #111); margin: 0; }
  .agp-error__msg { font-size: 0.8125rem; color: var(--dt3, #888); margin: 0 0 0.5rem; }

  /* Empty */
  .agp-empty {
    text-align: center;
    padding: 4rem 2rem;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    color: var(--dt4, #bbb);
  }
  .agp-empty__title { font-size: 1rem; font-weight: 600; color: var(--dt, #111); margin: 0; }
  .agp-empty__msg { font-size: 0.875rem; color: var(--dt3, #888); margin: 0 0 0.5rem; }

  .agp-count {
    text-align: center;
    font-size: 0.8125rem;
    color: var(--dt4, #bbb);
    margin-top: 0.5rem;
  }

  /* Side drawer */
  .agp-overlay {
    position: fixed;
    inset: 0;
    z-index: 40;
    background: rgba(0, 0, 0, 0.35);
    backdrop-filter: blur(1px);
  }
  .agp-drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 50;
    width: 100%;
    max-width: 480px;
    background: var(--dbg, #fff);
    border-left: 1px solid var(--dbd, #e0e0e0);
    box-shadow: -8px 0 32px rgba(0, 0, 0, 0.12);
    display: flex;
    flex-direction: column;
    animation: agp-slide-in 0.22s ease-out;
  }
  @keyframes agp-slide-in {
    from { transform: translateX(100%); }
    to   { transform: translateX(0); }
  }
  .agp-drawer__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.25rem 1.5rem;
    border-bottom: 1px solid var(--dbd, #e0e0e0);
    flex-shrink: 0;
  }
  .agp-drawer__preset-top {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-width: 0;
  }
  .agp-drawer__avatar {
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 10px;
    background: var(--dbg2, #f5f5f5);
    border: 1px solid var(--dbd, #e0e0e0);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--dt2, #555);
    flex-shrink: 0;
  }
  .agp-drawer__title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0 0 0.25rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .agp-drawer__cat {
    font-size: 0.6875rem;
    font-weight: 600;
    padding: 0.125rem 0.4375rem;
    border-radius: 4px;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt2, #555);
    text-transform: capitalize;
  }
  .agp-drawer__close {
    width: 30px; height: 30px;
    border-radius: 7px; border: none;
    background: none; color: var(--dt3, #888);
    cursor: pointer; display: flex;
    align-items: center; justify-content: center;
    flex-shrink: 0;
    transition: background 0.12s, color 0.12s;
  }
  .agp-drawer__close:hover { background: var(--dbg2, #f5f5f5); color: var(--dt, #111); }
  .agp-drawer__close:disabled { opacity: 0.4; cursor: not-allowed; }

  .agp-drawer__body {
    flex: 1;
    overflow-y: auto;
    padding: 1.25rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    scrollbar-width: thin;
  }
  .agp-drawer__desc {
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--dt2, #555);
    margin: 0;
  }
  .agp-drawer__section {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }
  .agp-drawer__section-label {
    font-size: 0.6875rem;
    font-weight: 700;
    color: var(--dt3, #888);
    text-transform: uppercase;
    letter-spacing: 0.07em;
    margin: 0;
  }
  .agp-drawer__mono {
    font-size: 0.8125rem;
    font-family: var(--bos-font-code-family, monospace);
    color: var(--dt, #111);
  }
  .agp-drawer__chips { display: flex; flex-wrap: wrap; gap: 0.3125rem; }
  .agp-drawer__chip {
    font-size: 0.75rem;
    padding: 0.2rem 0.5625rem;
    border-radius: 4px;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt2, #555);
    border: 1px solid var(--dbd, #e0e0e0);
  }
  .agp-drawer__chip--tag {
    background: transparent;
    color: var(--dt3, #888);
    border-color: var(--dbd2, #ebebeb);
  }
  .agp-drawer__prompt {
    font-size: 0.75rem;
    font-family: var(--bos-font-code-family, monospace);
    line-height: 1.7;
    color: var(--dt, #111);
    background: var(--dbg2, #f5f5f5);
    border: 1px solid var(--dbd2, #ebebeb);
    border-radius: 7px;
    padding: 0.875rem;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 280px;
    overflow-y: auto;
    scrollbar-width: thin;
  }
  .agp-drawer__input {
    width: 100%;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 7px;
    background: var(--dbg, #fff);
    color: var(--dt, #111);
    outline: none;
    transition: border-color 0.15s;
    box-sizing: border-box;
  }
  .agp-drawer__input::placeholder { color: var(--dt4, #bbb); }
  .agp-drawer__input:focus { border-color: var(--dt, #111); }
  .agp-drawer__input:disabled { opacity: 0.6; cursor: not-allowed; background: var(--dbg2, #f5f5f5); }
  .agp-drawer__hint { font-size: 0.75rem; color: var(--dt3, #888); margin: 0; }
  .agp-drawer__error {
    font-size: 0.8125rem;
    color: var(--bos-status-error, #ef4444);
    background: rgba(239, 68, 68, 0.07);
    border: 1px solid rgba(239, 68, 68, 0.15);
    border-radius: 6px;
    padding: 0.625rem 0.875rem;
  }
  .agp-drawer__footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--dbd, #e0e0e0);
    background: var(--dbg2, #f5f5f5);
    flex-shrink: 0;
  }
  .agp-drawer__create-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
  }
  .agp-drawer__create-btn:disabled { opacity: 0.65; cursor: not-allowed; }

  .agp-spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 1.5px solid rgba(255,255,255,0.4);
    border-top-color: #fff;
    border-radius: 50%;
    animation: agp-spin 0.6s linear infinite;
    flex-shrink: 0;
  }

  @keyframes agp-spin { to { transform: rotate(360deg); } }
  @keyframes agp-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.45; } }

  @media (max-width: 600px) {
    .agp-grid { grid-template-columns: 1fr; }
  }
  @media (max-width: 520px) {
    .agp-drawer { max-width: 100%; }
  }
</style>
