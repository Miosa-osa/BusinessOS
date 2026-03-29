<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { agents, categoryLabels } from '$lib/stores/agents';
  import type { CustomAgent } from '$lib/api/ai/types';
  import AgentSandbox from '$lib/components/agents/AgentSandbox.svelte';

  let agent = $state<CustomAgent | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let agentId = $derived($page.params.id);
  let activeTab = $state<'overview' | 'configuration' | 'testing' | 'history'>('overview');
  let showDeleteConfirm = $state(false);
  let isTogglingActive = $state(false);
  let copyFeedback = $state(false);

  // Inline edit state
  let editingName = $state(false);
  let editingDesc = $state(false);
  let editNameValue = $state('');
  let editDescValue = $state('');
  let savingName = $state(false);
  let savingDesc = $state(false);

  const tabs: { id: typeof activeTab; label: string }[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'configuration', label: 'Configuration' },
    { id: 'testing', label: 'Testing' },
    { id: 'history', label: 'History' }
  ];

  function getInitials(name: string | undefined): string {
    if (!name) return '??';
    const words = name.trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return '??';
    if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
    return (words[0][0] + words[1][0]).toUpperCase();
  }

  function formatDate(s: string): string {
    return new Date(s).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function formatUsage(n: number): string {
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
    return String(n);
  }

  async function loadAgent() {
    if (!agentId) return;
    loading = true;
    error = null;
    try {
      const result = await agents.loadAgent(agentId as string);
      if (!result) {
        error = 'Agent not found';
        agent = null;
      } else {
        agent = result;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load agent';
      agent = null;
    } finally {
      loading = false;
    }
  }

  async function handleToggleActive() {
    if (!agent || isTogglingActive || !agentId) return;
    isTogglingActive = true;
    try {
      await agents.updateAgent(agentId as string, { is_active: !agent.is_active });
      await loadAgent();
    } catch (err) {
      console.error('Failed to toggle agent status:', err);
    } finally {
      isTogglingActive = false;
    }
  }

  async function handleClone() {
    if (!agent) return;
    try {
      const cloned = await agents.createAgent({
        name: `${agent.name}-copy`,
        display_name: `${agent.display_name} (Copy)`,
        description: agent.description,
        avatar: agent.avatar,
        system_prompt: agent.system_prompt,
        model_preference: agent.model_preference,
        temperature: agent.temperature,
        max_tokens: agent.max_tokens,
        capabilities: agent.capabilities,
        tools_enabled: agent.tools_enabled,
        context_sources: agent.context_sources,
        thinking_enabled: agent.thinking_enabled,
        streaming_enabled: agent.streaming_enabled,
        category: agent.category,
        is_active: false
      });
      if (cloned) goto(`/agents/${cloned.id}`);
    } catch (err) {
      console.error('Failed to clone agent:', err);
    }
  }

  async function handleDelete() {
    if (!agentId) return;
    if (!showDeleteConfirm) { showDeleteConfirm = true; return; }
    try {
      await agents.deleteAgent(agentId as string);
      goto('/agents');
    } catch (err) {
      console.error('Failed to delete agent:', err);
      showDeleteConfirm = false;
    }
  }

  async function handleCopySystemPrompt() {
    if (!agent) return;
    await navigator.clipboard.writeText(agent.system_prompt);
    copyFeedback = true;
    setTimeout(() => { copyFeedback = false; }, 1800);
  }

  async function saveNameEdit() {
    if (!agent || !agentId || !editNameValue.trim()) { editingName = false; return; }
    savingName = true;
    try {
      await agents.updateAgent(agentId as string, { display_name: editNameValue.trim() });
      await loadAgent();
      editingName = false;
    } catch (err) { console.error('Failed to update name:', err); }
    finally { savingName = false; }
  }

  async function saveDescEdit() {
    if (!agent || !agentId) { editingDesc = false; return; }
    savingDesc = true;
    try {
      await agents.updateAgent(agentId as string, { description: editDescValue });
      await loadAgent();
      editingDesc = false;
    } catch (err) { console.error('Failed to update description:', err); }
    finally { savingDesc = false; }
  }

  onMount(() => { loadAgent(); });
</script>

<svelte:head>
  <title>{agent?.display_name ?? 'Agent'} — BusinessOS</title>
</svelte:head>

<div class="agd-page">

  {#if loading}
    <div class="agd-loading" aria-busy="true" aria-label="Loading agent">
      <div class="agd-spinner" aria-hidden="true"></div>
      <p class="agd-loading__text">Loading agent...</p>
    </div>

  {:else if error || !agent}
    <div class="agd-container">
      <a href="/agents" class="agd-back">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
        Agents
      </a>
      <div class="agd-error-state" role="alert">
        <div class="agd-error-state__icon" aria-hidden="true">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
        </div>
        <p class="agd-error-state__title">Agent not found</p>
        <p class="agd-error-state__msg">{error ?? 'The requested agent could not be loaded.'}</p>
        <button onclick={() => goto('/agents')} class="btn-pill btn-pill-secondary btn-pill-sm">Back to Agents</button>
      </div>
    </div>

  {:else}
    <div class="agd-container">

      <!-- Back link -->
      <a href="/agents" class="agd-back">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
        Agents
      </a>

      <!-- Two-column layout -->
      <div class="agd-layout">

        <!-- Sidebar -->
        <aside class="agd-sidebar">
          <!-- Avatar -->
          <div class="agd-sb__avatar" aria-hidden="true">
            <span class="agd-sb__initials">{getInitials(agent.display_name)}</span>
            {#if agent.is_featured}
              <span class="agd-sb__featured-dot" title="Featured"></span>
            {/if}
          </div>

          <!-- Name + handle -->
          <div class="agd-sb__identity">
            {#if editingName}
              <div class="agd-inline-edit">
                <input
                  class="agd-inline-input"
                  bind:value={editNameValue}
                  onblur={saveNameEdit}
                  onkeydown={(e) => { if (e.key === 'Enter') saveNameEdit(); if (e.key === 'Escape') editingName = false; }}
                  disabled={savingName}
                  aria-label="Edit agent name"
                />
                {#if savingName}<span class="agd-inline-saving" aria-label="Saving..."></span>{/if}
              </div>
            {:else}
              <button class="agd-editable-field agd-sb__name-btn" onclick={() => { editNameValue = agent.display_name ?? ''; editingName = true; }} title="Click to edit">
                <span class="agd-sb__name">{agent.display_name}</span>
                <svg class="agd-edit-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
            {/if}
            <span class="agd-sb__handle">@{agent.name}</span>
          </div>

          <!-- Status + category -->
          <div class="agd-sb__badges">
            <span class="agd-sb__status" class:agd-sb__status--active={agent.is_active}>
              <span class="agd-sb__status-dot" aria-hidden="true"></span>
              {agent.is_active ? 'Active' : 'Inactive'}
            </span>
            {#if agent.category}
              <span class="agd-sb__category">{categoryLabels[agent.category] ?? agent.category}</span>
            {/if}
          </div>

          <!-- Stats — micro-card grid -->
          <div class="agd-sb__stats">
            <!-- Model row stays as plain key-value (it's wide text) -->
            <div class="agd-sb__stat">
              <span class="agd-sb__stat-label">Model</span>
              <span class="agd-sb__stat-value agd-sb__stat-value--mono">
                {agent.model_preference ? agent.model_preference.replace('claude-', '').replace('-4', ' 4') : 'Default'}
              </span>
            </div>

            <!-- Micro-card grid for numeric/date stats -->
            <div class="agd-stats-grid">
              <div class="agd-stat">
                <div class="agd-stat__icon" aria-hidden="true">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2 3 14h9l-1 8 10-12h-9l1-8z"/></svg>
                </div>
                <p class="agd-stat__value">{formatUsage(agent.times_used ?? 0)}</p>
                <p class="agd-stat__label">Uses</p>
              </div>

              <div class="agd-stat">
                <div class="agd-stat__icon" aria-hidden="true">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
                </div>
                <p class="agd-stat__value">{formatDate(agent.created_at)}</p>
                <p class="agd-stat__label">Created</p>
              </div>

              <div class="agd-stat">
                <div class="agd-stat__icon" aria-hidden="true">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z"/></svg>
                </div>
                <p class="agd-stat__value">{agent.temperature ?? 0.7}</p>
                <p class="agd-stat__label">Temp</p>
              </div>

              <div class="agd-stat">
                <div class="agd-stat__icon" aria-hidden="true">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="4" y1="9" x2="20" y2="9"/><line x1="4" y1="15" x2="20" y2="15"/><line x1="10" y1="3" x2="8" y2="21"/><line x1="16" y1="3" x2="14" y2="21"/></svg>
                </div>
                <p class="agd-stat__value">{agent.max_tokens ? (agent.max_tokens / 1000).toFixed(0) + 'k' : '—'}</p>
                <p class="agd-stat__label">Max tokens</p>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="agd-sb__actions">
            <button onclick={() => goto(`/agents/${agentId}/edit`)} class="agd-sb__action-btn">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4z"/>
              </svg>
              Edit
            </button>
            <button onclick={handleClone} class="agd-sb__action-btn">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
              Clone
            </button>
            <button
              onclick={handleToggleActive}
              disabled={isTogglingActive}
              class="agd-sb__action-btn"
              aria-busy={isTogglingActive}
            >
              {#if isTogglingActive}
                <span class="agd-mini-spinner" aria-hidden="true"></span>
              {:else}
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/>
                </svg>
              {/if}
              {agent.is_active ? 'Deactivate' : 'Activate'}
            </button>
            <button onclick={() => activeTab = 'testing'} class="btn-cta agd-sb__test-btn">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
              Test Agent
            </button>
            <button onclick={handleDelete} class="agd-sb__action-btn agd-sb__action-btn--danger">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <polyline points="3 6 5 6 21 6"/>
                <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                <path d="M10 11v6m4-6v6"/>
                <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
              </svg>
              Delete
            </button>
          </div>
        </aside>

        <!-- Main content -->
        <div class="agd-main">

          <!-- Tab nav -->
          <div class="agd-tabs" role="tablist">
            {#each tabs as tab}
              <button
                class="agd-tab"
                class:agd-tab--active={activeTab === tab.id}
                onclick={() => activeTab = tab.id}
                role="tab"
                aria-selected={activeTab === tab.id}
                aria-controls={`agd-panel-${tab.id}`}
                id={`agd-tab-${tab.id}`}
              >{tab.label}</button>
            {/each}
          </div>

          <!-- Tab panels -->
          <div class="agd-panel" id={`agd-panel-${activeTab}`} role="tabpanel" aria-labelledby={`agd-tab-${activeTab}`}>

            <!-- Overview -->
            {#if activeTab === 'overview'}
              <div class="agd-overview">

                <!-- Description card -->
                <div class="agd-card">
                  <h2 class="agd-card__title">Description</h2>
                  {#if editingDesc}
                    <div class="agd-inline-edit agd-inline-edit--block">
                      <textarea
                        class="agd-inline-textarea"
                        bind:value={editDescValue}
                        onblur={saveDescEdit}
                        onkeydown={(e) => { if (e.key === 'Escape') editingDesc = false; }}
                        disabled={savingDesc}
                        rows="4"
                        aria-label="Edit agent description"
                      ></textarea>
                      <div class="agd-inline-edit__actions">
                        <button class="btn-pill btn-pill-ghost btn-pill-sm" onclick={() => editingDesc = false} disabled={savingDesc}>Cancel</button>
                        <button class="btn-cta" onclick={saveDescEdit} disabled={savingDesc} style="font-size:0.8rem;padding:0.35rem 0.875rem">
                          {savingDesc ? 'Saving…' : 'Save'}
                        </button>
                      </div>
                    </div>
                  {:else}
                    <div class="agd-editable-field agd-editable-field--block" onclick={() => { editDescValue = agent.description ?? ''; editingDesc = true; }} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && (() => { editDescValue = agent.description ?? ''; editingDesc = true; })()} title="Click to edit">
                      <p class="agd-overview__desc">{agent.description || 'No description. Click to add one.'}</p>
                      <svg class="agd-edit-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    </div>
                  {/if}
                </div>

                <!-- System prompt card -->
                <div class="agd-card">
                  <div class="agd-card__header">
                    <h2 class="agd-card__title">System Prompt</h2>
                    <button
                      onclick={handleCopySystemPrompt}
                      class="btn-pill btn-pill-ghost btn-pill-sm agd-copy-btn"
                      aria-label="Copy system prompt"
                    >
                      {#if copyFeedback}
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true"><polyline points="20 6 9 17 4 12"/></svg>
                        Copied
                      {:else}
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                          <rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                        </svg>
                        Copy
                      {/if}
                    </button>
                  </div>
                  <pre class="agd-prompt">{agent.system_prompt || 'No system prompt defined.'}</pre>
                </div>

                <!-- Chips section: capabilities -->
                {#if agent.capabilities && agent.capabilities.length > 0}
                  <div class="agd-card">
                    <h2 class="agd-card__title">Capabilities</h2>
                    <div class="agd-chips">
                      {#each agent.capabilities as cap}
                        <span class="agd-chip agd-chip--cap">{cap}</span>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Tools -->
                {#if agent.tools_enabled && agent.tools_enabled.length > 0}
                  <div class="agd-card">
                    <h2 class="agd-card__title">Tools Enabled</h2>
                    <div class="agd-chips">
                      {#each agent.tools_enabled as tool}
                        <span class="agd-chip agd-chip--tool">
                          <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                          </svg>
                          {tool.replace(/_/g, ' ')}
                        </span>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Context sources -->
                {#if agent.context_sources && agent.context_sources.length > 0}
                  <div class="agd-card">
                    <h2 class="agd-card__title">Context Sources</h2>
                    <div class="agd-chips">
                      {#each agent.context_sources as src}
                        <span class="agd-chip agd-chip--ctx">{src}</span>
                      {/each}
                    </div>
                  </div>
                {/if}

              </div>

            <!-- Configuration -->
            {:else if activeTab === 'configuration'}
              <div class="agd-config">
                <div class="agd-kv-grid">
                  {#each [
                    { label: 'Model Preference', value: agent.model_preference || 'Default', mono: true },
                    { label: 'Temperature', value: String(agent.temperature ?? 0.7) },
                    { label: 'Max Tokens', value: agent.max_tokens ? String(agent.max_tokens) : 'Default' },
                    { label: 'Streaming', value: agent.streaming_enabled ? 'Enabled' : 'Disabled', dot: agent.streaming_enabled },
                    { label: 'Thinking Mode', value: agent.thinking_enabled ? 'Enabled' : 'Disabled', dot: agent.thinking_enabled },
                    { label: 'Created', value: formatDate(agent.created_at) },
                    { label: 'Last Updated', value: formatDate(agent.updated_at) }
                  ] as row}
                    <div class="agd-kv">
                      <span class="agd-kv__label">{row.label}</span>
                      <span class="agd-kv__value" class:agd-kv__value--mono={row.mono}>
                        {#if row.dot !== undefined}
                          <span class="agd-kv__dot" class:agd-kv__dot--on={row.dot} aria-hidden="true"></span>
                        {/if}
                        {row.value}
                      </span>
                    </div>
                  {/each}
                </div>
              </div>

            <!-- Testing -->
            {:else if activeTab === 'testing'}
              <div class="agd-testing">
                <div class="agd-testing__intro">
                  <h2 class="agd-testing__title">Test Agent</h2>
                  <p class="agd-testing__subtitle">Send messages and see how {agent.display_name} responds in real time.</p>
                </div>
                <AgentSandbox agentId={agent.id} systemPrompt={agent.system_prompt} />
              </div>

            <!-- History -->
            {:else if activeTab === 'history'}
              <div class="agd-history">
                <div class="agd-history__empty">
                  <svg class="agd-history__icon" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                    <path d="M3 3h7v7H3z"/><path d="M14 3h7v7h-7z"/><path d="M14 14h7v7h-7z"/><path d="M3 14h7v7H3z"/>
                  </svg>
                  <p class="agd-history__title">No history yet</p>
                  <p class="agd-history__msg">Conversation history will appear here once the agent has been used.</p>
                  {#if agent.times_used && agent.times_used > 0}
                    <p class="agd-history__count">
                      This agent has been used <strong>{formatUsage(agent.times_used)}</strong> {agent.times_used === 1 ? 'time' : 'times'}.
                    </p>
                  {/if}
                </div>
              </div>
            {/if}

          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Delete confirmation modal -->
{#if showDeleteConfirm}
  <div
    class="agd-backdrop"
    onclick={() => showDeleteConfirm = false}
    onkeydown={(e) => e.key === 'Escape' && (showDeleteConfirm = false)}
    role="presentation"
    aria-hidden="true"
  >
    <div
      class="agd-modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="agd-del-title"
      tabindex="-1"
    >
      <div class="agd-modal__icon" aria-hidden="true">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
          <path d="M10 11v6m4-6v6"/>
          <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
        </svg>
      </div>
      <h3 class="agd-modal__title" id="agd-del-title">Delete Agent</h3>
      <p class="agd-modal__body">
        <strong>{agent?.display_name}</strong> will be permanently deleted. This cannot be undone.
      </p>
      <div class="agd-modal__actions">
        <button onclick={() => showDeleteConfirm = false} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
        <button onclick={handleDelete} class="btn-pill btn-pill-danger btn-pill-sm">Delete permanently</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ═══ Agent Detail Page — BOS Tokens ═══ */

  .agd-page {
    min-height: 100%;
    background: var(--dbg2, #f5f5f5);
  }

  .agd-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.75rem 1.5rem 3rem;
  }

  /* Back link */
  .agd-back {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.8125rem;
    color: var(--dt3, #888);
    text-decoration: none;
    margin-bottom: 1.5rem;
    transition: color 0.12s;
  }
  .agd-back:hover { color: var(--dt, #111); }

  /* Loading */
  .agd-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6rem 2rem;
    gap: 0.875rem;
  }
  .agd-spinner {
    width: 2rem;
    height: 2rem;
    border: 2px solid var(--dbd, #e0e0e0);
    border-top-color: var(--dt, #111);
    border-radius: 50%;
    animation: agd-spin 0.7s linear infinite;
  }
  .agd-loading__text { font-size: 0.875rem; color: var(--dt3, #888); }

  .agd-mini-spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 1.5px solid var(--dbd, #e0e0e0);
    border-top-color: var(--dt, #111);
    border-radius: 50%;
    animation: agd-spin 0.6s linear infinite;
    flex-shrink: 0;
  }

  /* Error state */
  .agd-error-state {
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
  .agd-error-state__icon {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    background: rgba(239, 68, 68, 0.08);
    color: var(--bos-status-error, #ef4444);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.25rem;
  }
  .agd-error-state__title { font-size: 1rem; font-weight: 700; color: var(--dt, #111); margin: 0; }
  .agd-error-state__msg { font-size: 0.875rem; color: var(--dt3, #888); margin: 0 0 0.5rem; }

  /* Two-column layout */
  .agd-layout {
    display: flex;
    gap: 1.5rem;
    align-items: flex-start;
  }

  /* Sidebar */
  .agd-sidebar {
    width: 260px;
    flex-shrink: 0;
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 12px;
    padding: 1.5rem 1.25rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    position: sticky;
    top: 1.5rem;
  }

  .agd-sb__avatar {
    width: 56px;
    height: 56px;
    border-radius: 14px;
    background: var(--dbg3, #eee);
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    flex-shrink: 0;
  }
  .agd-sb__initials {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--dt2, #555);
    letter-spacing: 0.02em;
    user-select: none;
  }
  .agd-sb__featured-dot {
    position: absolute;
    top: -3px;
    right: -3px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--bos-accent-blue, #3b82f6);
    border: 2px solid var(--dbg, #fff);
  }

  .agd-sb__identity { text-align: center; width: 100%; }
  .agd-sb__name {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0 0 0.125rem;
    word-break: break-word;
  }
  /* Name as editable button */
  .agd-sb__name-btn {
    justify-content: center;
    width: 100%;
  }
  .agd-sb__name-btn .agd-sb__name { margin: 0; }
  .agd-sb__handle {
    font-size: 0.75rem;
    color: var(--dt3, #888);
    font-family: var(--bos-font-code-family, monospace);
  }

  .agd-sb__badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    justify-content: center;
  }
  .agd-sb__status {
    display: inline-flex;
    align-items: center;
    gap: 0.3125rem;
    font-size: 0.6875rem;
    font-weight: 600;
    padding: 0.2rem 0.5625rem;
    border-radius: 9999px;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt3, #888);
  }
  .agd-sb__status--active { color: var(--bos-status-success, #22c55e); }
  .agd-sb__status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    flex-shrink: 0;
  }
  .agd-sb__status--active .agd-sb__status-dot { animation: agd-pulse 2s ease-in-out infinite; }

  .agd-sb__category {
    font-size: 0.6875rem;
    font-weight: 600;
    padding: 0.2rem 0.5625rem;
    border-radius: 9999px;
    background: var(--dbg2, #f5f5f5);
    color: var(--dt2, #555);
    text-transform: capitalize;
  }

  /* Sidebar stats wrapper */
  .agd-sb__stats {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.875rem 0;
    border-top: 1px solid var(--dbd2, #f0f0f0);
    border-bottom: 1px solid var(--dbd2, #f0f0f0);
  }
  .agd-sb__stat {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 0.5rem;
  }
  .agd-sb__stat-label { font-size: 0.75rem; color: var(--dt3, #888); flex-shrink: 0; }
  .agd-sb__stat-value {
    font-size: 0.75rem;
    color: var(--dt, #111);
    font-weight: 500;
    text-align: right;
    word-break: break-all;
  }
  .agd-sb__stat-value--mono { font-family: var(--bos-font-code-family, monospace); }

  /* Stat micro-card grid */
  .agd-stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
    margin-top: 0.25rem;
  }
  .agd-stat {
    background: var(--dbg2, #f5f5f5);
    border: 1px solid var(--dbd2, #ebebeb);
    border-radius: 8px;
    padding: 0.625rem 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }
  .agd-stat__icon {
    color: var(--dt3, #888);
    margin-bottom: 0.125rem;
  }
  .agd-stat__value {
    font-size: 0.9375rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0;
    font-family: var(--bos-font-number-family, monospace);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .agd-stat__label {
    font-size: 0.625rem;
    font-weight: 600;
    color: var(--dt3, #888);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0;
  }

  /* Sidebar actions */
  .agd-sb__actions {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }
  .agd-sb__action-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: 0.4375rem 0.75rem;
    font-size: 0.8125rem;
    color: var(--dt, #111);
    background: transparent;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 7px;
    cursor: pointer;
    transition: background 0.12s, border-color 0.12s;
  }
  .agd-sb__action-btn:hover {
    background: var(--dbg2, #f5f5f5);
    border-color: var(--dt, #111);
  }
  .agd-sb__action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .agd-sb__action-btn--danger { color: var(--bos-status-error, #ef4444); }
  .agd-sb__action-btn--danger:hover {
    background: rgba(239, 68, 68, 0.06);
    border-color: var(--bos-status-error, #ef4444);
  }
  .agd-sb__test-btn {
    width: 100%;
    justify-content: center;
    margin-top: 0.125rem;
  }

  /* Main content */
  .agd-main { flex: 1; min-width: 0; }

  /* Tabs */
  .agd-tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--dbd, #e0e0e0);
    margin-bottom: 1.5rem;
    background: var(--dbg, #fff);
    border-radius: 12px 12px 0 0;
    padding: 0 0.25rem;
  }
  .agd-tab {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--dt2, #555);
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: color 0.12s, border-color 0.12s;
    white-space: nowrap;
    margin-bottom: -1px;
  }
  .agd-tab:hover { color: var(--dt, #111); }
  .agd-tab--active {
    color: var(--dt, #111);
    font-weight: 600;
    border-bottom-color: var(--dt, #111);
  }

  /* Panel */
  .agd-panel {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 0 0 12px 12px;
    padding: 1.5rem;
  }

  /* Overview */
  .agd-overview { display: flex; flex-direction: column; gap: 1.25rem; }
  .agd-overview__desc {
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--dt2, #555);
    margin: 0;
  }
  .agd-card {
    background: var(--dbg2, #f5f5f5);
    border: 1px solid var(--dbd2, #ebebeb);
    border-radius: 8px;
    padding: 1rem 1.125rem;
  }
  .agd-card__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.625rem;
  }
  .agd-card__title {
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0 0 0.625rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
  .agd-card__header .agd-card__title { margin: 0; }
  .agd-card__text {
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--dt2, #555);
    margin: 0;
  }

  .agd-copy-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.3125rem;
  }

  .agd-prompt {
    font-size: 0.8125rem;
    font-family: var(--bos-font-code-family, monospace);
    line-height: 1.65;
    color: var(--dt, #111);
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 320px;
    overflow-y: auto;
    scrollbar-width: thin;
  }

  /* Chips */
  .agd-chips { display: flex; flex-wrap: wrap; gap: 0.375rem; }
  .agd-chip {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.25rem 0.625rem;
    border-radius: 5px;
    display: inline-flex;
    align-items: center;
    gap: 0.3125rem;
  }
  .agd-chip--cap {
    background: var(--dbg, #fff);
    color: var(--dt2, #555);
    border: 1px solid var(--dbd, #e0e0e0);
  }
  .agd-chip--tool {
    background: rgba(59, 130, 246, 0.08);
    color: var(--bos-accent-blue, #2563eb);
    border: 1px solid rgba(59, 130, 246, 0.15);
  }
  .agd-chip--ctx {
    background: rgba(234, 179, 8, 0.08);
    color: #a16207;
    border: 1px solid rgba(234, 179, 8, 0.2);
  }

  /* Configuration */
  .agd-config { padding: 0.25rem 0; }
  .agd-kv-grid {
    display: flex;
    flex-direction: column;
    gap: 0;
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 8px;
    overflow: hidden;
  }
  .agd-kv {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--dbd2, #f0f0f0);
    gap: 1rem;
  }
  .agd-kv:last-child { border-bottom: none; }
  .agd-kv:nth-child(even) { background: var(--dbg2, #f9f9f9); }
  .agd-kv__label { font-size: 0.8125rem; color: var(--dt3, #888); flex-shrink: 0; }
  .agd-kv__value {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--dt, #111);
    text-align: right;
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
  }
  .agd-kv__value--mono { font-family: var(--bos-font-code-family, monospace); }
  .agd-kv__dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--dt4, #ccc);
    flex-shrink: 0;
  }
  .agd-kv__dot--on { background: var(--bos-status-success, #22c55e); }

  /* Testing */
  .agd-testing__intro { margin-bottom: 1.25rem; }
  .agd-testing__title {
    font-size: 0.9375rem;
    font-weight: 700;
    color: var(--dt, #111);
    margin: 0 0 0.25rem;
  }
  .agd-testing__subtitle { font-size: 0.8125rem; color: var(--dt3, #888); margin: 0; }

  /* History */
  .agd-history__empty {
    text-align: center;
    padding: 3rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }
  .agd-history__icon {
    color: var(--dt4, #ccc);
    margin-bottom: 0.25rem;
  }
  .agd-history__title { font-size: 0.9375rem; font-weight: 600; color: var(--dt2, #555); margin: 0; }
  .agd-history__msg { font-size: 0.8125rem; color: var(--dt3, #888); margin: 0; max-width: 340px; }
  .agd-history__count { font-size: 0.8125rem; color: var(--dt3, #888); margin: 0; }
  .agd-history__count strong { color: var(--dt, #111); }

  /* Backdrop + modal */
  .agd-backdrop {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(2px);
  }
  .agd-modal {
    background: var(--dbg, #fff);
    border: 1px solid var(--dbd, #e0e0e0);
    border-radius: 14px;
    padding: 1.75rem;
    max-width: 420px;
    width: calc(100% - 2rem);
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.18);
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.625rem;
  }
  .agd-modal__icon {
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 10px;
    background: rgba(239, 68, 68, 0.08);
    color: var(--bos-status-error, #ef4444);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.25rem;
  }
  .agd-modal__title { font-size: 1rem; font-weight: 700; color: var(--dt, #111); margin: 0; }
  .agd-modal__body { font-size: 0.875rem; color: var(--dt3, #888); margin: 0 0 0.5rem; line-height: 1.5; }
  .agd-modal__body strong { color: var(--dt, #111); }
  .agd-modal__actions { display: flex; gap: 0.5rem; justify-content: center; }

  /* ═══ Inline editing ═══ */
  .agd-editable-field {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.125rem 0;
    border-radius: 4px;
    transition: background 0.12s;
    text-align: left;
  }
  .agd-editable-field:hover { background: var(--dbg2, #f5f5f5); }
  .agd-editable-field--block {
    display: flex;
    width: 100%;
    padding: 0.375rem;
    border-radius: 6px;
  }
  .agd-edit-icon {
    color: var(--dt4, #bbb);
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.12s;
  }
  .agd-editable-field:hover .agd-edit-icon { opacity: 1; }

  .agd-inline-edit { display: flex; align-items: center; gap: 0.5rem; }
  .agd-inline-edit--block { flex-direction: column; align-items: flex-start; gap: 0.5rem; }
  .agd-inline-input {
    font-size: 1rem;
    font-weight: 700;
    color: var(--dt, #111);
    border: 1px solid var(--bos-accent-blue, #3b82f6);
    border-radius: 6px;
    padding: 0.25rem 0.5rem;
    background: var(--dbg, #fff);
    outline: none;
    flex: 1;
    width: 100%;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
  }
  .agd-inline-textarea {
    width: 100%;
    font-size: 0.875rem;
    color: var(--dt, #111);
    border: 1px solid var(--bos-accent-blue, #3b82f6);
    border-radius: 6px;
    padding: 0.5rem 0.625rem;
    background: var(--dbg, #fff);
    outline: none;
    resize: vertical;
    line-height: 1.55;
    box-sizing: border-box;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
  }
  .agd-inline-edit__actions { display: flex; gap: 0.375rem; }
  .agd-inline-saving {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 1.5px solid var(--dbd, #e0e0e0);
    border-top-color: var(--dt, #111);
    border-radius: 50%;
    animation: agd-spin 0.6s linear infinite;
    flex-shrink: 0;
  }

  @keyframes agd-spin { to { transform: rotate(360deg); } }
  @keyframes agd-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }

  /* Responsive */
  @media (max-width: 768px) {
    .agd-layout { flex-direction: column; }
    .agd-sidebar { width: 100%; position: static; }
    .agd-tabs { overflow-x: auto; scrollbar-width: none; }
    .agd-tabs::-webkit-scrollbar { display: none; }
  }
</style>
