<!-- 
  IntegrationCard.svelte
  Card component for displaying integration options (Google, Slack, etc.)
  Used in the BusinessOS conversational onboarding
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';
  import CheckIcon from './icons/CheckIcon.svelte';

  export let provider: string = "";
  export let name: string = "";
  export let description: string = "";
  export let icon: string = "";
  export let connected: boolean = false;
  export let connecting: boolean = false;

  const dispatch = createEventDispatcher();

  function handleConnect() {
    dispatch('connect', { provider });
  }
</script>

<div class="integration-card" class:connected>
  <div class="card-content">
    <div class="icon-wrapper">
      {#if icon}
        <img src={icon} alt={name} class="provider-icon" />
      {:else}
        <div class="icon-placeholder">{name.charAt(0)}</div>
      {/if}
    </div>
    <div class="info">
      <h4 class="name">{name}</h4>
      <p class="description">{description}</p>
    </div>
  </div>
  
  <div class="action">
    {#if connected}
      <div class="connected-badge">
        <CheckIcon size={16} />
        <span>Connected</span>
      </div>
    {:else if connecting}
      <Button variant="outline" disabled className="connect-btn">
        <span class="spinner"></span>
        Connecting...
      </Button>
    {:else}
      <Button variant="outline" className="connect-btn" on:click={handleConnect}>
        Connect
      </Button>
    {/if}
  </div>
</div>

<style>
  .integration-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.75rem;
    background-color: var(--card, #ffffff);
    transition: all 150ms ease;
  }

  :global(.dark) .integration-card {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .integration-card:hover:not(.connected) {
    border-color: var(--primary, #000000);
  }

  :global(.dark) .integration-card:hover:not(.connected) {
    border-color: var(--primary, #ffffff);
  }

  .integration-card.connected {
    border-color: var(--success, #10b981);
    background-color: rgba(16, 185, 129, 0.05);
  }

  .card-content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .icon-wrapper {
    width: 2.5rem;
    height: 2.5rem;
    flex-shrink: 0;
  }

  .provider-icon {
    width: 100%;
    height: 100%;
    object-fit: contain;
    border-radius: 0.375rem;
  }

  .icon-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--muted, #f9fafb);
    border-radius: 0.375rem;
    font-weight: 600;
    font-size: 1.125rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .icon-placeholder {
    background-color: var(--muted, #1a1a1a);
    color: var(--foreground, #f9fafb);
  }

  .info {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .name {
    margin: 0;
    font-size: 0.9375rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .name {
    color: var(--foreground, #f9fafb);
  }

  .description {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--muted-foreground, #6b7280);
  }

  .action {
    flex-shrink: 0;
  }

  :global(.connect-btn) {
    min-width: 100px !important;
  }

  .connected-badge {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    background-color: rgba(16, 185, 129, 0.1);
    border-radius: 9999px;
    color: var(--success, #10b981);
    font-size: 0.875rem;
    font-weight: 500;
  }

  .spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid var(--border, #e5e7eb);
    border-top-color: var(--primary, #000000);
    border-radius: 9999px;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
