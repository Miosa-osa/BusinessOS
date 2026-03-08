<!--
  VersionPreviewModal.svelte

  Modal showing a preview of the app at a specific version.

  Features:
  - Shows version details (number, label, trigger, time)
  - Displays config snapshot in a readable format
  - Restore button to apply this version
  - Close/dismiss actions
-->
<script lang="ts">
  import { X, RotateCcw, Clock, Sparkles, Pencil, Camera, Save, Eye } from 'lucide-svelte';
  import type { Version, VersionTrigger } from '$lib/types/versions';
  import { formatRelativeTime, getTriggerLabel } from '$lib/types/versions';

  interface Props {
    version: Version;
    isOpen: boolean;
    onClose: () => void;
    onRestore: (version: Version) => void;
  }

  let {
    version,
    isOpen,
    onClose,
    onRestore
  }: Props = $props();

  function getTriggerIcon(trigger: VersionTrigger) {
    switch (trigger) {
      case 'ai_generation':
        return Sparkles;
      case 'user_edit':
        return Pencil;
      case 'manual_snapshot':
        return Camera;
      case 'auto_snapshot':
        return Save;
      case 'restore':
        return RotateCcw;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('modal-backdrop')) {
      onClose();
    }
  }

  // Format config for display
  function formatConfig(config: Record<string, unknown>): string {
    return JSON.stringify(config, null, 2);
  }

  const TriggerIcon = $derived(getTriggerIcon(version.trigger));
</script>

{#if isOpen}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="modal-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="modal-title"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
  >
    <div class="modal">
      <!-- Header -->
      <header class="modal-header">
        <div class="header-left">
          <div class="preview-badge">
            <Eye size={14} strokeWidth={2} />
            Preview
          </div>
          <h2 id="modal-title" class="modal-title">Version {version.versionNumber}</h2>
          {#if version.label}
            <span class="version-label">{version.label}</span>
          {/if}
        </div>
        <button
          type="button"
          class="btn-pill btn-pill-ghost close-btn"
          onclick={onClose}
          aria-label="Close preview"
        >
          <X size={20} strokeWidth={2} />
        </button>
      </header>

      <!-- Version Meta -->
      <div class="version-meta">
        <div class="meta-item">
          <TriggerIcon size={14} strokeWidth={2} />
          <span>{getTriggerLabel(version.trigger)}</span>
        </div>
        <div class="meta-divider"></div>
        <div class="meta-item">
          <Clock size={14} strokeWidth={2} />
          <span>{formatRelativeTime(version.createdAt)}</span>
        </div>
        {#if version.createdByName}
          <div class="meta-divider"></div>
          <div class="meta-item">
            <span>by {version.createdByName}</span>
          </div>
        {/if}
      </div>

      {#if version.prompt}
        <div class="prompt-section">
          <span class="prompt-label">Prompt</span>
          <p class="prompt-text">"{version.prompt}"</p>
        </div>
      {/if}

      <!-- Content -->
      <div class="modal-content">
        <div class="config-section">
          <div class="config-header">
            <h3 class="config-title">Configuration Snapshot</h3>
            <span class="config-hint">Read-only preview of saved state</span>
          </div>
          <pre class="config-code">{formatConfig(version.configSnapshot)}</pre>
        </div>
      </div>

      <!-- Footer -->
      <footer class="modal-footer">
        <button
          type="button"
          class="btn-pill btn-pill-ghost btn btn-secondary"
          onclick={onClose}
        >
          Close
        </button>
        {#if !version.isCurrent}
          <button
            type="button"
            class="btn-pill btn-pill-ghost btn btn-primary"
            onclick={() => onRestore(version)}
          >
            <RotateCcw size={14} strokeWidth={2} />
            Restore this version
          </button>
        {:else}
          <span class="current-indicator">This is the current version</span>
        {/if}
      </footer>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
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

  :global(.dark) .modal-backdrop {
    background: rgba(0, 0, 0, 0.6);
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    width: 100%;
    max-width: 600px;
    max-height: 80vh;
    background: #ffffff;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    display: flex;
    flex-direction: column;
    animation: slideIn 200ms ease-out;
  }

  :global(.dark) .modal {
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
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  :global(.dark) .modal-header {
    border-bottom-color: #374151;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .preview-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6366f1;
    background: #eef2ff;
    border-radius: 4px;
  }

  :global(.dark) .preview-badge {
    background: rgba(99, 102, 241, 0.15);
    color: #a5b4fc;
  }

  .modal-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  :global(.dark) .modal-title {
    color: #f3f4f6;
  }

  .version-label {
    font-size: 13px;
    color: #6b7280;
  }

  :global(.dark) .version-label {
    color: #9ca3af;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    color: #6b7280;
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: all 150ms ease;
  }

  .close-btn:hover {
    color: #111827;
    background: #f3f4f6;
  }

  :global(.dark) .close-btn:hover {
    color: #f3f4f6;
    background: #374151;
  }

  /* Version Meta */
  .version-meta {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    background: #f9fafb;
    border-bottom: 1px solid #e5e7eb;
  }

  :global(.dark) .version-meta {
    background: #111827;
    border-bottom-color: #374151;
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;
  }

  :global(.dark) .meta-item {
    color: #9ca3af;
  }

  .meta-divider {
    width: 1px;
    height: 16px;
    background: #e5e7eb;
  }

  :global(.dark) .meta-divider {
    background: #374151;
  }

  /* Prompt Section */
  .prompt-section {
    padding: 12px 20px;
    border-bottom: 1px solid #e5e7eb;
  }

  :global(.dark) .prompt-section {
    border-bottom-color: #374151;
  }

  .prompt-label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #9ca3af;
    display: block;
    margin-bottom: 4px;
  }

  .prompt-text {
    font-size: 13px;
    font-style: italic;
    color: #374151;
    margin: 0;
  }

  :global(.dark) .prompt-text {
    color: #d1d5db;
  }

  /* Content */
  .modal-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  .config-section {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .config-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .config-title {
    font-size: 13px;
    font-weight: 600;
    color: #374151;
    margin: 0;
  }

  :global(.dark) .config-title {
    color: #d1d5db;
  }

  .config-hint {
    font-size: 12px;
    color: #9ca3af;
  }

  .config-code {
    padding: 16px;
    font-size: 12px;
    font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', monospace;
    color: #374151;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    overflow-x: auto;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }

  :global(.dark) .config-code {
    color: #d1d5db;
    background: #111827;
    border-color: #374151;
  }

  /* Footer */
  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 20px;
    border-top: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  :global(.dark) .modal-footer {
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

  .btn-secondary:hover {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  :global(.dark) .btn-secondary:hover {
    background: #4b5563;
  }

  .btn-primary {
    color: #ffffff;
    background: #6366f1;
    border: 1px solid #6366f1;
  }

  .btn-primary:hover {
    background: #4f46e5;
    border-color: #4f46e5;
  }

  .current-indicator {
    font-size: 13px;
    color: #059669;
    font-weight: 500;
  }

  :global(.dark) .current-indicator {
    color: #34d399;
  }

  /* Responsive */
  @media (max-width: 640px) {
    .modal {
      max-height: 90vh;
    }

    .header-left {
      flex-wrap: wrap;
      gap: 8px;
    }

    .version-meta {
      flex-wrap: wrap;
    }
  }
</style>
