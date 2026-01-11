<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { createWorkspaceInvite, type WorkspaceRole } from '$lib/api/workspaces';
  import { X, Mail, Send, Loader2, AlertCircle, CheckCircle } from 'lucide-svelte';

  interface Props {
    workspaceId: string;
    roles: WorkspaceRole[];
  }

  let { workspaceId, roles }: Props = $props();

  const dispatch = createEventDispatcher();

  let email = $state('');
  let selectedRole = $state('member');
  let isSending = $state(false);
  let error = $state<string | null>(null);
  let success = $state(false);

  // Filter out system roles that shouldn't be assigned via invite
  const assignableRoles = $derived(roles.filter((r) => r.name !== 'owner'));

  async function handleSend() {
    if (!email.trim() || !selectedRole) return;

    try {
      isSending = true;
      error = null;

      // In dev mode with mock workspace, simulate success
      if (import.meta.env.DEV && workspaceId.startsWith('mock-')) {
        console.log('[InviteMemberModal] Mock mode - simulating invite to:', email.trim());
        await new Promise(resolve => setTimeout(resolve, 800)); // Fake delay
        success = true;
        setTimeout(() => dispatch('success'), 1500);
        return;
      }

      await createWorkspaceInvite(workspaceId, {
        email: email.trim(),
        role: selectedRole,
      });

      dispatch('success');
    } catch (err) {
      console.error('Failed to send invitation:', err);
      error = err instanceof Error ? err.message : 'Failed to send invitation';
    } finally {
      isSending = false;
    }
  }

  function handleCancel() {
    dispatch('cancel');
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      handleCancel();
    } else if (e.key === 'Enter' && email.trim() && selectedRole && !isSending) {
      handleSend();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-overlay" onclick={handleCancel}>
  <div class="modal-content" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div>
        <h2>Invite Member</h2>
        <p>Send an invitation to join this workspace</p>
      </div>
      <button class="close-button" onclick={handleCancel} type="button">
        <X class="w-5 h-5" />
      </button>
    </div>

    {#if success}
      <div class="success-message">
        <CheckCircle class="w-5 h-5" />
        <span>Invitation sent to <strong>{email}</strong>!</span>
      </div>
    {:else}
      {#if error}
        <div class="error-message">
          <AlertCircle class="w-4 h-4" />
          <span>{error}</span>
        </div>
      {/if}

      <div class="modal-body">
        <div class="form-group">
          <label for="email">
            <Mail class="w-4 h-4" />
            Email Address
          </label>
          <input
            id="email"
            type="email"
            bind:value={email}
            placeholder="member@example.com"
            disabled={isSending}
            autofocus
          />
        </div>

        <div class="form-group">
          <label for="role">Role</label>
          <select id="role" bind:value={selectedRole} disabled={isSending}>
            {#each assignableRoles as role (role.id)}
              <option value={role.name}>
                {role.display_name}
                {#if role.description}
                  - {role.description}
                {/if}
              </option>
            {/each}
          </select>
          <p class="field-hint">The role determines what permissions the member will have</p>
        </div>
      </div>

      <div class="modal-footer">
        <button class="cancel-button" onclick={handleCancel} type="button" disabled={isSending}>
          Cancel
        </button>
        <button
          class="send-button"
          onclick={handleSend}
          disabled={!email.trim() || !selectedRole || isSending}
          type="button"
        >
          {#if isSending}
          <Loader2 class="w-4 h-4 animate-spin" />
        {:else}
          <Send class="w-4 h-4" />
        {/if}
        Send Invitation
      </button>
    </div>
    {/if}
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-content {
    width: 100%;
    max-width: 500px;
    background: white;
    border-radius: 0.75rem;
    box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.25rem 0;
  }

  .modal-header p {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

  .close-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    background: transparent;
    border: none;
    color: #9ca3af;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: all 0.15s;
  }

  .close-button:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: #fef2f2;
    color: #dc2626;
    font-size: 0.875rem;
    border-bottom: 1px solid #fee2e2;
  }

  .success-message {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 3rem 1.5rem;
    background: #f0fdf4;
    color: #16a34a;
    font-size: 1rem;
    text-align: center;
  }

  .success-message strong {
    color: #15803d;
  }

  .modal-body {
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-group label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .form-group input,
  .form-group select {
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    transition: all 0.15s;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-group input:disabled,
  .form-group select:disabled {
    background: #f9fafb;
    color: #9ca3af;
    cursor: not-allowed;
  }

  .field-hint {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0;
  }

  .modal-footer {
    display: flex;
    gap: 0.75rem;
    padding: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  .cancel-button,
  .send-button {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .cancel-button {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .cancel-button:hover:not(:disabled) {
    background: #f9fafb;
  }

  .send-button {
    background: #3b82f6;
    color: white;
  }

  .send-button:hover:not(:disabled) {
    background: #2563eb;
  }

  .send-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(.dark) .modal-content {
    background: #1f2937;
  }

  :global(.dark) .modal-header {
    border-bottom-color: #374151;
  }

  :global(.dark) .modal-header h2 {
    color: #f9fafb;
  }

  :global(.dark) .modal-header p {
    color: #9ca3af;
  }

  :global(.dark) .form-group label {
    color: #d1d5db;
  }

  :global(.dark) .form-group input,
  :global(.dark) .form-group select {
    background: #111827;
    border-color: #374151;
    color: #f9fafb;
  }

  :global(.dark) .modal-footer {
    border-top-color: #374151;
  }

  :global(.dark) .cancel-button {
    background: #111827;
    border-color: #374151;
    color: #d1d5db;
  }

  :global(.dark) .cancel-button:hover:not(:disabled) {
    background: #0f172a;
  }
</style>
