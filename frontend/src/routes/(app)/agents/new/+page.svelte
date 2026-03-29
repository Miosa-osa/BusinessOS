<script lang="ts">
  import { goto } from '$app/navigation';
  import { agents } from '$lib/stores/agents';
  import AgentBuilder from '$lib/components/agents/AgentBuilder.svelte';
  import type { CustomAgent } from '$lib/api/ai/types';

  let error = $state<string | null>(null);

  async function handleSave(data: Partial<CustomAgent>) {
    error = null;
    try {
      const newAgent = await agents.createAgent({
        ...data,
        is_active: data.is_active ?? true,
      });
      await goto(`/agents/${newAgent.id}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create agent. Please try again.';
    }
  }

  function handleCancel() {
    goto('/agents');
  }
</script>

<svelte:head>
  <title>New Agent — BusinessOS</title>
</svelte:head>

<div class="np-page">
  <!-- Top bar -->
  <header class="np-header">
    <button class="np-back" onclick={() => goto('/agents')}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M15 18l-6-6 6-6"/>
      </svg>
      Agents
    </button>
    <span class="np-header__title">New Agent</span>
    <span></span>
  </header>

  <!-- Error banner -->
  {#if error}
    <div class="np-error" role="alert">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/>
      </svg>
      {error}
      <button class="np-error__close" onclick={() => error = null} aria-label="Dismiss">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
      </button>
    </div>
  {/if}

  <!-- Builder fills remaining space -->
  <div class="np-builder">
    <AgentBuilder onSave={handleSave} onCancel={handleCancel} />
  </div>
</div>

<style>
  .np-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--dbg, #fff);
    overflow: hidden;
  }

  /* Top bar */
  .np-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.5rem;
    border-bottom: 1px solid var(--dbd, #e5e5e5);
    flex-shrink: 0;
    background: var(--dbg, #fff);
  }
  .np-back {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--dt2, #555);
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.25rem 0;
    transition: color 0.12s;
  }
  .np-back:hover { color: var(--dt, #111); }
  .np-header__title {
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--dt, #111);
  }

  /* Error banner */
  .np-error {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1.5rem;
    font-size: 0.8125rem;
    color: var(--bos-status-error, #dc2626);
    background: rgba(220, 38, 38, 0.06);
    border-bottom: 1px solid rgba(220, 38, 38, 0.15);
    flex-shrink: 0;
  }
  .np-error__close {
    margin-left: auto;
    width: 20px;
    height: 20px;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: currentColor;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.6;
    transition: opacity 0.1s;
  }
  .np-error__close:hover { opacity: 1; }

  /* Builder takes all remaining height */
  .np-builder {
    flex: 1;
    overflow: hidden;
    min-height: 0;
  }
</style>
