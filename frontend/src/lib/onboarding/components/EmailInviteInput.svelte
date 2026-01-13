<!-- 
  EmailInviteInput.svelte
  Email input for inviting team members
  With add/remove functionality
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';

  export let emails: string[] = [];
  export let placeholder: string = "colleague@company.com";

  const dispatch = createEventDispatcher();

  let inputValue = "";
  let error = "";

  function validateEmail(email: string): boolean {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
  }

  function handleAdd() {
    const email = inputValue.trim().toLowerCase();
    
    if (!email) return;

    if (!validateEmail(email)) {
      error = "Please enter a valid email address";
      return;
    }

    if (emails.includes(email)) {
      error = "This email has already been added";
      return;
    }

    error = "";
    emails = [...emails, email];
    inputValue = "";
    dispatch('change', { emails });
  }

  function handleRemove(email: string) {
    emails = emails.filter(e => e !== email);
    dispatch('change', { emails });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      handleAdd();
    }
  }

  function handleSendInvites() {
    if (emails.length > 0) {
      dispatch('send', { emails });
    }
  }

  function handleSkip() {
    dispatch('skip');
  }
</script>

<div class="email-invite-input">
  <p class="label">Invite your team:</p>

  <div class="input-row">
    <input
      type="email"
      bind:value={inputValue}
      {placeholder}
      on:keydown={handleKeydown}
      class="email-input"
      class:error={!!error}
    />
    <Button variant="outline" on:click={handleAdd}>
      + Add
    </Button>
  </div>

  {#if error}
    <p class="error-message">{error}</p>
  {/if}

  {#if emails.length > 0}
    <div class="added-emails">
      <span class="added-label">Added:</span>
      <ul class="email-list">
        {#each emails as email}
          <li class="email-item">
            <span class="email-text">{email}</span>
            <button class="remove-btn" on:click={() => handleRemove(email)}>
              ×
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <div class="actions">
    <Button 
      disabled={emails.length === 0}
      on:click={handleSendInvites}
    >
      Send Invites
    </Button>
    <Button 
      variant="ghost"
      on:click={handleSkip}
    >
      Skip for now
    </Button>
  </div>
</div>

<style>
  .email-invite-input {
    width: 100%;
    max-width: 28rem;
    padding: 1rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.75rem;
  }

  :global(.dark) .email-invite-input {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .label {
    margin: 0 0 0.75rem;
    font-size: 0.9375rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .label {
    color: var(--foreground, #f9fafb);
  }

  .input-row {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .email-input {
    flex: 1;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.375rem;
    background-color: transparent;
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
    outline: none;
    font-family: inherit;
  }

  .email-input::placeholder {
    color: var(--muted-foreground, #6b7280);
  }

  .email-input:focus {
    border-color: var(--ring, #000000);
  }

  .email-input.error {
    border-color: var(--error, #ef4444);
  }

  :global(.dark) .email-input {
    border-color: var(--border, #2a2a2a);
    color: var(--foreground, #f9fafb);
  }

  .error-message {
    margin: 0.25rem 0 0.5rem;
    font-size: 0.8125rem;
    color: var(--error, #ef4444);
  }

  .added-emails {
    margin-bottom: 1rem;
  }

  .added-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.8125rem;
    color: var(--muted-foreground, #6b7280);
  }

  .email-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .email-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0.5rem;
    background-color: var(--muted, #f9fafb);
    border-radius: 0.25rem;
  }

  :global(.dark) .email-item {
    background-color: var(--muted, #1a1a1a);
  }

  .email-text {
    font-size: 0.875rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .email-text {
    color: var(--foreground, #f9fafb);
  }

  .remove-btn {
    width: 1.5rem;
    height: 1.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 1.25rem;
    color: var(--muted-foreground, #6b7280);
    border-radius: 0.25rem;
    transition: all 150ms;
  }

  .remove-btn:hover {
    background-color: var(--destructive, #ef4444);
    color: white;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }
</style>
