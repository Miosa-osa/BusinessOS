<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { AlertTriangle, X } from 'lucide-svelte';

  interface Props {
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
  }

  let {
    title,
    message,
    confirmText = 'Delete',
    cancelText = 'Cancel',
  }: Props = $props();

  const dispatch = createEventDispatcher();

  let confirmInput = $state('');
  const requiredText = 'DELETE';

  function handleConfirm() {
    if (confirmInput === requiredText) {
      dispatch('confirm');
    }
  }

  function handleCancel() {
    dispatch('cancel');
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      handleCancel();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-overlay" onclick={handleCancel}>
  <div class="modal-content" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div class="modal-icon">
        <AlertTriangle class="w-6 h-6" />
      </div>
      <button class="btn-pill btn-pill-icon btn-pill-ghost btn-pill-sm" onclick={handleCancel} type="button">
        <X class="w-5 h-5" />
      </button>
    </div>

    <div class="modal-body">
      <h2>{title}</h2>
      <p>{message}</p>

      <div class="confirm-input-group">
        <label for="confirm-input">
          Type <strong>{requiredText}</strong> to confirm
        </label>
        <input
          id="confirm-input"
          type="text"
          bind:value={confirmInput}
          placeholder={requiredText}
          autocomplete="off"
        />
      </div>
    </div>

    <div class="modal-footer">
      <button class="btn-pill btn-pill-secondary flex-1" onclick={handleCancel} type="button">
        {cancelText}
      </button>
      <button
        class="btn-pill btn-pill-danger flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
        onclick={handleConfirm}
        disabled={confirmInput !== requiredText}
        type="button"
      >
        {confirmText}
      </button>
    </div>
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
    padding: 1.5rem 1.5rem 0 1.5rem;
  }

  .modal-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 3rem;
    height: 3rem;
    background: #fef2f2;
    color: #dc2626;
    border-radius: 50%;
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

  .modal-body {
    padding: 1.5rem;
  }

  .modal-body h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.5rem 0;
  }

  .modal-body p {
    color: #6b7280;
    font-size: 0.875rem;
    line-height: 1.5;
    margin: 0 0 1.5rem 0;
  }

  .confirm-input-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .confirm-input-group label {
    font-size: 0.875rem;
    color: #374151;
  }

  .confirm-input-group strong {
    font-weight: 600;
    color: #dc2626;
  }

  .confirm-input-group input {
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-family: monospace;
    transition: all 0.15s;
  }

  .confirm-input-group input:focus {
    outline: none;
    border-color: #dc2626;
    box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.1);
  }

  .modal-footer {
    display: flex;
    gap: 0.75rem;
    padding: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  .cancel-button,
  .confirm-button {
    flex: 1;
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

  .cancel-button:hover {
    background: #f9fafb;
  }

  .confirm-button {
    background: #dc2626;
    color: white;
  }

  .confirm-button:hover:not(:disabled) {
    background: #b91c1c;
  }

  .confirm-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(.dark) .modal-content {
    background: #1f2937;
  }

  :global(.dark) .modal-body h2 {
    color: #f9fafb;
  }

  :global(.dark) .modal-body p {
    color: #9ca3af;
  }

  :global(.dark) .confirm-input-group input {
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

  :global(.dark) .cancel-button:hover {
    background: #0f172a;
  }
</style>
