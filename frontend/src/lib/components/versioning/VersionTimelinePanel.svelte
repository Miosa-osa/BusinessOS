<!--
  VersionTimelinePanel.svelte

  Full version history in a slide-over panel.

  Features:
  - Vertical timeline with connecting lines
  - Version details (label, trigger, time)
  - Preview and Restore actions
  - Current version indicator
  - Compare versions button
-->
<script lang="ts">
  import {
    X,
    Eye,
    RotateCcw,
    Sparkles,
    Pencil,
    Camera,
    Save,
    Star,
    GitCompare
  } from 'lucide-svelte';
  import type { Version, VersionTrigger } from '$lib/types/versions';
  import { formatRelativeTime, getTriggerLabel } from '$lib/types/versions';

  interface Props {
    appId: string;
    versions: Version[];
    isOpen: boolean;
    isLoading?: boolean;
    onClose: () => void;
    onPreview: (version: Version) => void;
    onRestore: (version: Version) => void;
    onCompare?: () => void;
  }

  let {
    appId,
    versions,
    isOpen,
    isLoading = false,
    onClose,
    onPreview,
    onRestore,
    onCompare
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
    if ((e.target as HTMLElement).classList.contains('panel-backdrop')) {
      onClose();
    }
  }
</script>

{#if isOpen}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="panel-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="panel-title"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
  >
    <aside class="timeline-panel">
      <!-- Header -->
      <header class="panel-header">
        <h2 id="panel-title" class="panel-title">Version History</h2>
        <button type="button" class="btn-pill btn-pill-ghost close-btn" onclick={onClose} aria-label="Close panel">
          <X size={20} strokeWidth={2} />
        </button>
      </header>

      <!-- Content -->
      <div class="panel-content">
        {#if isLoading}
          <div class="loading-state">
            <div class="spinner"></div>
            <span>Loading history...</span>
          </div>
        {:else if versions.length === 0}
          <div class="empty-state">
            <div class="empty-icon">
              <Save size={32} strokeWidth={1.5} />
            </div>
            <h3>No versions yet</h3>
            <p>Versions are created when you or AI make changes to this app.</p>
          </div>
        {:else}
          <div class="timeline">
            {#each versions as version, index (version.id)}
              {@const TriggerIcon = getTriggerIcon(version.trigger)}
              {@const isLast = index === versions.length - 1}

              <div class="timeline-item" class:current={version.isCurrent}>
                <!-- Timeline dot and line -->
                <div class="timeline-track">
                  <div class="timeline-dot" class:current={version.isCurrent}>
                    {#if version.isCurrent}
                      <div class="dot-inner"></div>
                    {/if}
                  </div>
                  {#if !isLast}
                    <div class="timeline-line"></div>
                  {/if}
                </div>

                <!-- Version content -->
                <div class="timeline-content">
                  <div class="version-header">
                    <span class="version-number">v{version.versionNumber}</span>
                    {#if version.isCurrent}
                      <span class="current-badge">current</span>
                    {/if}
                    <span class="version-time">{formatRelativeTime(version.createdAt)}</span>
                  </div>

                  {#if version.label}
                    <p class="version-label">
                      {#if version.trigger === 'manual_snapshot'}
                        <Star size={12} strokeWidth={2} class="label-icon" />
                      {/if}
                      {version.label}
                    </p>
                  {/if}

                  <div class="version-meta">
                    <TriggerIcon size={12} strokeWidth={2} />
                    <span>{getTriggerLabel(version.trigger)}</span>
                    {#if version.createdByName}
                      <span class="separator">•</span>
                      <span>{version.createdByName}</span>
                    {/if}
                  </div>

                  {#if version.prompt}
                    <p class="version-prompt">"{version.prompt}"</p>
                  {/if}

                  <!-- Actions -->
                  <div class="version-actions">
                    <button
                      type="button"
                      class="btn-pill btn-pill-ghost action-btn"
                      onclick={() => onPreview(version)}
                    >
                      <Eye size={14} strokeWidth={2} />
                      Preview
                    </button>
                    {#if !version.isCurrent}
                      <button
                        type="button"
                        class="btn-pill btn-pill-ghost action-btn restore"
                        onclick={() => onRestore(version)}
                      >
                        <RotateCcw size={14} strokeWidth={2} />
                        Restore
                      </button>
                    {/if}
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      {#if versions.length > 1 && onCompare}
        <footer class="panel-footer">
          <button type="button" class="btn-pill btn-pill-ghost compare-btn" onclick={onCompare}>
            <GitCompare size={16} strokeWidth={2} />
            Compare Versions
          </button>
        </footer>
      {/if}
    </aside>
  </div>
{/if}

<style>
  .panel-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.3);
    z-index: 100;
    animation: fadeIn 150ms ease-out;
  }

  :global(.dark) .panel-backdrop {
    background: rgba(0, 0, 0, 0.5);
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .timeline-panel {
    position: absolute;
    top: 0;
    right: 0;
    width: 400px;
    max-width: 100%;
    height: 100%;
    background: #ffffff;
    border-left: 1px solid #e5e7eb;
    display: flex;
    flex-direction: column;
    animation: slideIn 200ms ease-out;
  }

  :global(.dark) .timeline-panel {
    background: #111827;
    border-left-color: #374151;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
    }
    to {
      transform: translateX(0);
    }
  }

  /* Header */
  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  :global(.dark) .panel-header {
    border-bottom-color: #374151;
  }

  .panel-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  :global(.dark) .panel-title {
    color: #f3f4f6;
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

  /* Content */
  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  /* Loading and empty states */
  .loading-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    height: 200px;
    color: #9ca3af;
    text-align: center;
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 2px solid #e5e7eb;
    border-top-color: #6366f1;
    border-radius: 50%;
    animation: spin 600ms linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .empty-icon {
    width: 64px;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f3f4f6;
    border-radius: 12px;
    color: #9ca3af;
  }

  :global(.dark) .empty-icon {
    background: #374151;
  }

  .empty-state h3 {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    margin: 0;
  }

  :global(.dark) .empty-state h3 {
    color: #e5e7eb;
  }

  .empty-state p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
    max-width: 200px;
  }

  /* Timeline */
  .timeline {
    display: flex;
    flex-direction: column;
  }

  .timeline-item {
    display: flex;
    gap: 16px;
    padding-bottom: 24px;
  }

  .timeline-item:last-child {
    padding-bottom: 0;
  }

  .timeline-track {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 16px;
    flex-shrink: 0;
  }

  .timeline-dot {
    width: 12px;
    height: 12px;
    background: #e5e7eb;
    border: 2px solid #ffffff;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 0 2px #e5e7eb;
  }

  :global(.dark) .timeline-dot {
    background: #4b5563;
    border-color: #111827;
    box-shadow: 0 0 0 2px #374151;
  }

  .timeline-dot.current {
    background: #6366f1;
    box-shadow: 0 0 0 2px #c7d2fe;
  }

  :global(.dark) .timeline-dot.current {
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.3);
  }

  .dot-inner {
    width: 100%;
    height: 100%;
    background: #6366f1;
    border-radius: 50%;
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .timeline-line {
    width: 2px;
    flex: 1;
    background: #e5e7eb;
    margin-top: 4px;
  }

  :global(.dark) .timeline-line {
    background: #374151;
  }

  .timeline-content {
    flex: 1;
    min-width: 0;
  }

  .version-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .version-number {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
  }

  :global(.dark) .version-number {
    color: #f3f4f6;
  }

  .current-badge {
    font-size: 11px;
    font-weight: 500;
    color: #059669;
    background: #ecfdf5;
    padding: 2px 6px;
    border-radius: 4px;
  }

  :global(.dark) .current-badge {
    color: #34d399;
    background: rgba(16, 185, 129, 0.1);
  }

  .version-time {
    font-size: 12px;
    color: #9ca3af;
    margin-left: auto;
  }

  .version-label {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: #374151;
    margin: 0 0 6px 0;
  }

  :global(.dark) .version-label {
    color: #d1d5db;
  }

  .label-icon {
    color: #f59e0b;
  }

  .version-meta {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: #6b7280;
    margin-bottom: 6px;
  }

  :global(.dark) .version-meta {
    color: #9ca3af;
  }

  .separator {
    color: #d1d5db;
  }

  :global(.dark) .separator {
    color: #4b5563;
  }

  .version-prompt {
    font-size: 12px;
    font-style: italic;
    color: #6b7280;
    margin: 0 0 8px 0;
    padding-left: 8px;
    border-left: 2px solid #e5e7eb;
  }

  :global(.dark) .version-prompt {
    color: #9ca3af;
    border-left-color: #374151;
  }

  .version-actions {
    display: flex;
    gap: 8px;
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    font-size: 12px;
    font-weight: 500;
    color: #6b7280;
    background: transparent;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    cursor: pointer;
    font-family: inherit;
    transition: all 150ms ease;
  }

  :global(.dark) .action-btn {
    color: #9ca3af;
    border-color: #374151;
  }

  .action-btn:hover {
    color: #374151;
    background: #f3f4f6;
    border-color: #d1d5db;
  }

  :global(.dark) .action-btn:hover {
    color: #e5e7eb;
    background: #374151;
    border-color: #4b5563;
  }

  .action-btn.restore {
    color: #6366f1;
    border-color: #c7d2fe;
  }

  :global(.dark) .action-btn.restore {
    color: #a5b4fc;
    border-color: rgba(99, 102, 241, 0.3);
  }

  .action-btn.restore:hover {
    background: #eef2ff;
    border-color: #a5b4fc;
  }

  :global(.dark) .action-btn.restore:hover {
    background: rgba(99, 102, 241, 0.1);
  }

  /* Footer */
  .panel-footer {
    padding: 16px 20px;
    border-top: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  :global(.dark) .panel-footer {
    border-top-color: #374151;
  }

  .compare-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    padding: 10px 16px;
    font-size: 13px;
    font-weight: 500;
    color: #374151;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
    transition: all 150ms ease;
  }

  :global(.dark) .compare-btn {
    color: #d1d5db;
    background: #1f2937;
    border-color: #374151;
  }

  .compare-btn:hover {
    background: #f3f4f6;
    border-color: #d1d5db;
  }

  :global(.dark) .compare-btn:hover {
    background: #374151;
    border-color: #4b5563;
  }

  /* Responsive */
  @media (max-width: 480px) {
    .timeline-panel {
      width: 100%;
    }
  }
</style>
