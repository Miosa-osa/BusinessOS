<!--
  VersionBadge.svelte

  Always-visible indicator showing current version status.

  States:
  - Normal: "v5 • Saved"
  - Unsaved: "v5 • Unsaved" (with dot indicator)
  - Previewing: "Viewing v3"
-->
<script lang="ts">
  import { Check, Circle, Eye } from 'lucide-svelte';

  interface Props {
    currentVersion: number;
    hasUnsavedChanges?: boolean;
    isPreviewingOldVersion?: boolean;
    previewingVersion?: number;
    onclick?: () => void;
    class?: string;
  }

  let {
    currentVersion,
    hasUnsavedChanges = false,
    isPreviewingOldVersion = false,
    previewingVersion,
    onclick,
    class: className = ''
  }: Props = $props();

  const isClickable = $derived(!!onclick);
</script>

<button
  type="button"
  class="btn-pill btn-pill-ghost version-badge {className}"
  class:clickable={isClickable}
  class:unsaved={hasUnsavedChanges}
  class:previewing={isPreviewingOldVersion}
  onclick={onclick}
  disabled={!isClickable}
>
  {#if isPreviewingOldVersion && previewingVersion}
    <Eye size={14} strokeWidth={2} class="icon" />
    <span class="label">Viewing v{previewingVersion}</span>
  {:else}
    <span class="version">v{currentVersion}</span>
    <span class="separator">•</span>
    {#if hasUnsavedChanges}
      <Circle size={8} fill="currentColor" class="dot unsaved" />
      <span class="status">Unsaved</span>
    {:else}
      <Check size={14} strokeWidth={2} class="icon saved" />
      <span class="status">Saved</span>
    {/if}
  {/if}
</button>

<style>
  .version-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    font-size: 13px;
    font-weight: 500;
    color: #6b7280;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-family: inherit;
    cursor: default;
    transition: all 150ms ease;
  }

  :global(.dark) .version-badge {
    color: #9ca3af;
    background: #1f2937;
    border-color: #374151;
  }

  .version-badge.clickable {
    cursor: pointer;
  }

  .version-badge.clickable:hover {
    background: #f3f4f6;
    border-color: #d1d5db;
  }

  :global(.dark) .version-badge.clickable:hover {
    background: #374151;
    border-color: #4b5563;
  }

  .version-badge.unsaved {
    color: #d97706;
    background: #fffbeb;
    border-color: #fde68a;
  }

  :global(.dark) .version-badge.unsaved {
    color: #fbbf24;
    background: rgba(251, 191, 36, 0.1);
    border-color: rgba(251, 191, 36, 0.3);
  }

  .version-badge.previewing {
    color: #6366f1;
    background: #eef2ff;
    border-color: #c7d2fe;
  }

  :global(.dark) .version-badge.previewing {
    color: #a5b4fc;
    background: rgba(99, 102, 241, 0.1);
    border-color: rgba(99, 102, 241, 0.3);
  }

  .version {
    font-weight: 600;
    color: #374151;
  }

  :global(.dark) .version {
    color: #e5e7eb;
  }

  .separator {
    color: #d1d5db;
  }

  :global(.dark) .separator {
    color: #4b5563;
  }

  .icon {
    flex-shrink: 0;
  }

  .icon.saved {
    color: #10b981;
  }

  .dot.unsaved {
    color: #f59e0b;
  }

  .status {
    font-weight: 500;
  }

  .label {
    font-weight: 500;
  }
</style>
