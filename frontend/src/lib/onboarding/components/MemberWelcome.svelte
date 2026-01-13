<!-- 
  MemberWelcome.svelte
  Simplified welcome screen for team members (not full onboarding)
  Members inherit workspace settings from admin
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import PurpleOrb from './PurpleOrb.svelte';
  import Button from './Button.svelte';
  import Input from './Input.svelte';
  import CheckIcon from './icons/CheckIcon.svelte';

  interface WorkspaceInfo {
    name: string;
    inviterEmail: string;
    integrations: string[];
    memberCount: number;
  }

  export let workspace: WorkspaceInfo = {
    name: 'Your Workspace',
    inviterEmail: 'admin@company.com',
    integrations: [],
    memberCount: 1,
  };

  const dispatch = createEventDispatcher();

  let memberName = "";
  let memberRole = "";

  function handleGetStarted() {
    dispatch('getStarted', { name: memberName, role: memberRole });
  }
</script>

<div class="member-welcome">
  <div class="orb-wrapper">
    <PurpleOrb size={64} />
  </div>

  <div class="content">
    <h2 class="title">👋 Welcome to {workspace.name}!</h2>
    <p class="invited-by">You've been invited by {workspace.inviterEmail}</p>

    <div class="workspace-info">
      <h4 class="info-title">Your workspace has:</h4>
      {#if workspace.integrations.length > 0}
        {#each workspace.integrations as integration}
          <div class="info-item">
            <CheckIcon size={16} className="check-icon" />
            <span>{integration} connected</span>
          </div>
        {/each}
      {/if}
      <div class="info-item">
        <span class="member-icon">👥</span>
        <span>{workspace.memberCount} team member{workspace.memberCount !== 1 ? 's' : ''}</span>
      </div>
    </div>

    <div class="divider"></div>

    <div class="form-section">
      <p class="form-label">Let's personalize your experience:</p>
      
      <div class="form-field">
        <label for="member-name" class="field-label">Your name:</label>
        <Input
          id="member-name"
          bind:value={memberName}
          placeholder="Enter your name"
        />
      </div>

      <div class="form-field">
        <label for="member-role" class="field-label">Your role: <span class="optional">(optional)</span></label>
        <Input
          id="member-role"
          bind:value={memberRole}
          placeholder="e.g., Marketing Manager"
        />
      </div>
    </div>

    <Button 
      className="get-started-btn"
      disabled={!memberName.trim()}
      on:click={handleGetStarted}
    >
      Get Started →
    </Button>
  </div>
</div>

<style>
  .member-welcome {
    text-align: center;
    max-width: 28rem;
    margin: 0 auto;
    padding: 2rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 1rem;
  }

  :global(.dark) .member-welcome {
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
    gap: 0.75rem;
  }

  .title {
    margin: 0;
    font-size: 1.375rem;
    font-weight: 600;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .title {
    color: var(--foreground, #f9fafb);
  }

  .invited-by {
    margin: 0;
    font-size: 0.875rem;
    color: var(--muted-foreground, #6b7280);
  }

  .workspace-info {
    text-align: left;
    margin-top: 0.5rem;
  }

  .info-title {
    margin: 0 0 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .info-title {
    color: var(--foreground, #f9fafb);
  }

  .info-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--foreground, #1f2937);
    margin-bottom: 0.25rem;
  }

  :global(.dark) .info-item {
    color: var(--foreground, #f9fafb);
  }

  :global(.check-icon) {
    color: var(--success, #10b981) !important;
  }

  .member-icon {
    font-size: 1rem;
  }

  .divider {
    height: 1px;
    background-color: var(--border, #e5e7eb);
    margin: 0.75rem 0;
  }

  :global(.dark) .divider {
    background-color: var(--border, #2a2a2a);
  }

  .form-section {
    text-align: left;
  }

  .form-label {
    margin: 0 0 0.75rem;
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .form-label {
    color: var(--foreground, #f9fafb);
  }

  .form-field {
    margin-bottom: 0.75rem;
  }

  .field-label {
    display: block;
    margin-bottom: 0.25rem;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .field-label {
    color: var(--foreground, #f9fafb);
  }

  .optional {
    font-weight: 400;
    color: var(--muted-foreground, #6b7280);
  }

  :global(.get-started-btn) {
    margin-top: 0.5rem !important;
    width: 100% !important;
    height: 2.75rem !important;
  }
</style>
