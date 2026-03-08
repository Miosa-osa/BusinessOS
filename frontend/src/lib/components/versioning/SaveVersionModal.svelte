<!--
  SaveVersionModal.svelte

  Modal for creating a manual version snapshot.

  Features:
  - Optional label input
  - Shows what version number will be created
  - Quick save without label option
-->
<script lang="ts">
  import { Save, X } from 'lucide-svelte';

  interface Props {
    appId: string;
    currentVersion: number;
    isOpen: boolean;
    isSaving?: boolean;
    onClose: () => void;
    onSave: (label?: string) => void;
  }

  let {
    appId,
    currentVersion,
    isOpen,
    isSaving = false,
    onClose,
    onSave
  }: Props = $props();

  let label = $state('');
  let inputRef: HTMLInputElement;

  const newVersionNumber = $derived(currentVersion + 1);

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    onSave(label.trim() || undefined);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && !isSaving) {
      onClose();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('dialog-backdrop') && !isSaving) {
      onClose();
    }
  }

  // Reset and focus on open
  $effect(() => {
    if (isOpen) {
      label = '';
      // Focus input after animation
      setTimeout(() => inputRef?.focus(), 100);
    }
  });
</script>

{#if isOpen}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="dialog-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="dialog-title"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
  >
    <div class="dialog">
      <!-- Header -->
      <header class="dialog-header">
        <div class="dialog-icon">
          <Save size={20} strokeWidth={2} />
        </div>
        <div>
          <h2 id="dialog-title" class="dialog-title">Save Version</h2>
          <p class="dialog-subtitle">Create v{newVersionNumber}</p>
        </div>
        <button
          type="button"
          class="btn-pill btn-pill-ghost close-btn"
          onclick={onClose}
          disabled={isSaving}
          aria-label="Close dialog"
        >
          <X size={18} strokeWidth={2} />
        </button>
      </header>

      <!-- Content -->
      <form onsubmit={handleSubmit}>
        <div class="dialog-content">
          <p class="description">
            Save the current state as a new version you can restore later.
          </p>

          <div class="form-group">
            <label for="version-label" class="label">
              Label <span class="optional">(optional)</span>
            </label>
            <input
              bind:this={inputRef}
              bind:value={label}
              type="text"
              id="version-label"
              class="input"
              placeholder="e.g., Pre-demo stable, Before refactor"
              maxlength="100"
              disabled={isSaving}
            />
            <p class="hint">A short description to help you remember this version</p>
          </div>
        </div>

        <!-- Footer -->
        <footer class="dialog-footer">
          <button
            type="button"
            class="btn-pill btn-pill-ghost btn btn-secondary"
            onclick={onClose}
            disabled={isSaving}
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn-pill btn-pill-ghost btn btn-primary"
            disabled={isSaving}
          >
            {#if isSaving}
              <span class="spinner"></span>
              Saving...
            {:else}
              <Save size={14} strokeWidth={2} />
              Save Version
            {/if}
          </button>
        </footer>
      </form>
    </div>
  </div>
{/if}

<style>
  .dialog-backdrop {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.4);
    z-index: 200;
    padding: 20px;
    animation: fadeIn 150ms ease-out;
  }

  :global(.dark) .dialog-backdrop {
    background: rgba(0, 0, 0, 0.6);
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .dialog {
    width: 100%;
    max-width: 400px;
    background: #ffffff;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    animation: slideIn 200ms ease-out;
  }

  :global(.dark) .dialog {
    background: #1f2937;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(-10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  /* Header */
  .dialog-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 20px 20px 0;
  }

  .dialog-icon {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #eef2ff;
    color: #6366f1;
    border-radius: 10px;
    flex-shrink: 0;
  }

  :global(.dark) .dialog-icon {
    background: rgba(99, 102, 241, 0.15);
  }

  .dialog-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  :global(.dark) .dialog-title {
    color: #f3f4f6;
  }

  .dialog-subtitle {
    font-size: 13px;
    color: #6b7280;
    margin: 2px 0 0 0;
  }

  :global(.dark) .dialog-subtitle {
    color: #9ca3af;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    margin-left: auto;
    color: #9ca3af;
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: all 150ms ease;
  }

  .close-btn:hover:not(:disabled) {
    color: #6b7280;
    background: #f3f4f6;
  }

  :global(.dark) .close-btn:hover:not(:disabled) {
    color: #d1d5db;
    background: #374151;
  }

  .close-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Content */
  .dialog-content {
    padding: 20px;
  }

  .description {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 16px 0;
  }

  :global(.dark) .description {
    color: #9ca3af;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .label {
    font-size: 13px;
    font-weight: 500;
    color: #374151;
  }

  :global(.dark) .label {
    color: #d1d5db;
  }

  .optional {
    font-weight: 400;
    color: #9ca3af;
  }

  .input {
    width: 100%;
    padding: 10px 12px;
    font-size: 14px;
    color: #111827;
    background: #ffffff;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-family: inherit;
    transition: all 150ms ease;
  }

  :global(.dark) .input {
    color: #f3f4f6;
    background: #111827;
    border-color: #374151;
  }

  .input:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  }

  :global(.dark) .input:focus {
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
  }

  .input::placeholder {
    color: #9ca3af;
  }

  :global(.dark) .input::placeholder {
    color: #6b7280;
  }

  .input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .hint {
    font-size: 12px;
    color: #9ca3af;
    margin: 0;
  }

  /* Footer */
  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 20px;
    border-top: 1px solid #e5e7eb;
  }

  :global(.dark) .dialog-footer {
    border-top-color: #374151;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
    transition: all 150ms ease;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-secondary {
    color: #374151;
    background: #ffffff;
    border: 1px solid #e5e7eb;
  }

  :global(.dark) .btn-secondary {
    color: #d1d5db;
    background: #374151;
    border-color: #4b5563;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  :global(.dark) .btn-secondary:hover:not(:disabled) {
    background: #4b5563;
  }

  .btn-primary {
    color: #ffffff;
    background: #6366f1;
    border: 1px solid #6366f1;
  }

  .btn-primary:hover:not(:disabled) {
    background: #4f46e5;
    border-color: #4f46e5;
  }

  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #ffffff;
    border-radius: 50%;
    animation: spin 600ms linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
