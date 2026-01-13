<!-- 
  CompletionScreen.svelte
  Final screen showing onboarding summary and completion
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import PurpleOrb from './PurpleOrb.svelte';
  import Button from './Button.svelte';
  import CheckIcon from './icons/CheckIcon.svelte';

  interface WorkspaceData {
    name: string;
    industry: string;
    teamSize: string;
  }

  interface Integration {
    name: string;
    connected: boolean;
  }

  interface Invite {
    email: string;
  }

  export let workspace: WorkspaceData = {
    name: 'Your Workspace',
    industry: 'Not specified',
    teamSize: 'Not specified',
  };

  export let integrations: Integration[] = [];
  export let invites: Invite[] = [];

  const dispatch = createEventDispatcher();

  function handleGotoDashboard() {
    dispatch('gotoDashboard');
  }

  $: connectedIntegrations = integrations.filter(i => i.connected);
</script>

<div class="completion-screen">
  <div class="orb-wrapper">
    <PurpleOrb size={64} />
  </div>

  <div class="content">
    <h2 class="title">🎉 Your workspace is ready!</h2>
    <p class="subtitle">Here's what we set up:</p>

    <div class="summary">
      <div class="summary-item">
        <CheckIcon size={18} className="check-icon" />
        <span><strong>Workspace:</strong> {workspace.name}</span>
      </div>
      <div class="summary-item">
        <CheckIcon size={18} className="check-icon" />
        <span><strong>Industry:</strong> {workspace.industry}</span>
      </div>
      <div class="summary-item">
        <CheckIcon size={18} className="check-icon" />
        <span><strong>Team size:</strong> {workspace.teamSize}</span>
      </div>

      {#if connectedIntegrations.length > 0}
        <div class="section">
          <h4 class="section-title">Connected integrations:</h4>
          <ul class="integration-list">
            {#each connectedIntegrations as integration}
              <li class="integration-item">
                <span>• {integration.name}</span>
                <CheckIcon size={14} className="check-icon-small" />
              </li>
            {/each}
          </ul>
        </div>
      {/if}

      {#if invites.length > 0}
        <div class="section">
          <h4 class="section-title">Team invites sent:</h4>
          <ul class="invite-list">
            {#each invites as invite}
              <li class="invite-item">• {invite.email}</li>
            {/each}
          </ul>
        </div>
      {/if}
    </div>

    <Button 
      className="dashboard-btn"
      on:click={handleGotoDashboard}
    >
      Go to Dashboard →
    </Button>
  </div>
</div>

<style>
  .completion-screen {
    text-align: center;
    max-width: 28rem;
    margin: 0 auto;
    padding: 2rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 1rem;
  }

  :global(.dark) .completion-screen {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .orb-wrapper {
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: center;
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .title {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .title {
    color: var(--foreground, #f9fafb);
  }

  .subtitle {
    margin: 0;
    font-size: 0.9375rem;
    color: var(--muted-foreground, #6b7280);
  }

  .summary {
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
  }

  .summary-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .summary-item {
    color: var(--foreground, #f9fafb);
  }

  :global(.check-icon) {
    color: var(--success, #10b981) !important;
    flex-shrink: 0;
  }

  .section {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border, #e5e7eb);
  }

  :global(.dark) .section {
    border-color: var(--border, #2a2a2a);
  }

  .section-title {
    margin: 0 0 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--muted-foreground, #6b7280);
  }

  .integration-list,
  .invite-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .integration-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .integration-item {
    color: var(--foreground, #f9fafb);
  }

  :global(.check-icon-small) {
    color: var(--success, #10b981) !important;
  }

  .invite-item {
    font-size: 0.875rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .invite-item {
    color: var(--foreground, #f9fafb);
  }

  :global(.dashboard-btn) {
    margin-top: 1rem !important;
    width: 100% !important;
    height: 3rem !important;
    font-size: 1rem !important;
  }
</style>
