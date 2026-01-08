<script lang="ts">
  import { onMount } from 'svelte';
  import {
    workspaces,
    currentWorkspace,
    currentUserRole,
    workspaceLoading,
    workspaceError,
    switchWorkspace,
    loadSavedWorkspace,
  } from '$lib/stores/workspaces';
  import { ChevronDown, Building2, Loader2, AlertCircle } from 'lucide-svelte';

  let isOpen = false;
  let dropdownRef: HTMLDivElement;

  onMount(() => {
    // Load workspaces on mount
    loadSavedWorkspace();

    // Close dropdown when clicking outside
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
        isOpen = false;
      }
    };

    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  });

  async function handleWorkspaceSelect(workspaceId: string) {
    if ($currentWorkspace?.id === workspaceId) {
      isOpen = false;
      return;
    }

    try {
      await switchWorkspace(workspaceId);
      isOpen = false;
    } catch (error) {
      console.error('Failed to switch workspace:', error);
    }
  }

  function toggleDropdown() {
    isOpen = !isOpen;
  }
</script>

<div class="workspace-switcher" bind:this={dropdownRef}>
  <!-- Trigger Button -->
  <button
    class="workspace-trigger"
    class:loading={$workspaceLoading.switching}
    on:click={toggleDropdown}
    disabled={$workspaceLoading.switching}
    aria-label="Switch workspace"
    aria-expanded={isOpen}
  >
    {#if $workspaceLoading.switching}
      <Loader2 class="w-4 h-4 animate-spin text-gray-400" />
    {:else}
      <Building2 class="w-4 h-4 text-gray-400" />
    {/if}

    <div class="workspace-info">
      {#if $currentWorkspace}
        <span class="workspace-name">{$currentWorkspace.name}</span>
        {#if $currentUserRole}
          <span class="workspace-role">{$currentUserRole}</span>
        {/if}
      {:else}
        <span class="workspace-name text-gray-400">Select Workspace</span>
      {/if}
    </div>

    <div class="transition-transform" class:rotate-180={isOpen}>
      <ChevronDown class="w-4 h-4 text-gray-400" />
    </div>
  </button>

  <!-- Dropdown Menu -->
  {#if isOpen}
    <div class="workspace-dropdown">
      {#if $workspaceError}
        <div class="error-message">
          <AlertCircle class="w-4 h-4" />
          <span>{$workspaceError}</span>
        </div>
      {/if}

      {#if $workspaceLoading.workspaces}
        <div class="loading-state">
          <Loader2 class="w-5 h-5 animate-spin" />
          <span>Loading workspaces...</span>
        </div>
      {:else if $workspaces.length === 0}
        <div class="empty-state">
          <Building2 class="w-8 h-8 text-gray-300 mb-2" />
          <p class="text-sm text-gray-500">No workspaces available</p>
        </div>
      {:else}
        <div class="workspace-list">
          {#each $workspaces as workspace (workspace.id)}
            <button
              class="workspace-item"
              class:active={$currentWorkspace?.id === workspace.id}
              on:click={() => handleWorkspaceSelect(workspace.id)}
            >
              <div class="workspace-item-icon">
                {#if workspace.logo_url}
                  <img src={workspace.logo_url} alt={workspace.name} class="w-8 h-8 rounded" />
                {:else}
                  <div class="workspace-avatar">
                    {workspace.name.charAt(0).toUpperCase()}
                  </div>
                {/if}
              </div>

              <div class="workspace-item-info">
                <span class="workspace-item-name">{workspace.name}</span>
                <span class="workspace-item-slug">{workspace.slug}</span>
              </div>

              {#if $currentWorkspace?.id === workspace.id}
                <div class="workspace-item-check">
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              {/if}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .workspace-switcher {
    position: relative;
    display: inline-block;
  }

  .workspace-trigger {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 1rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
    min-width: 200px;
  }

  .workspace-trigger:hover:not(:disabled) {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  .workspace-trigger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .workspace-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.125rem;
  }

  .workspace-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
    line-height: 1.25;
  }

  .workspace-role {
    font-size: 0.75rem;
    color: #6b7280;
    line-height: 1;
  }

  .workspace-dropdown {
    position: absolute;
    top: calc(100% + 0.5rem);
    left: 0;
    width: 320px;
    max-height: 400px;
    overflow-y: auto;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
    z-index: 50;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: #fef2f2;
    border-bottom: 1px solid #fee2e2;
    color: #dc2626;
    font-size: 0.875rem;
  }

  .loading-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    color: #6b7280;
  }

  .loading-state {
    gap: 0.75rem;
  }

  .workspace-list {
    padding: 0.5rem;
  }

  .workspace-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.75rem;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.15s;
    border: none;
    background: transparent;
    text-align: left;
  }

  .workspace-item:hover {
    background: #f3f4f6;
  }

  .workspace-item.active {
    background: #eff6ff;
  }

  .workspace-item-icon {
    flex-shrink: 0;
  }

  .workspace-avatar {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    font-weight: 600;
    font-size: 0.875rem;
    border-radius: 0.375rem;
  }

  .workspace-item-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  .workspace-item-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
    line-height: 1.25;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .workspace-item-slug {
    font-size: 0.75rem;
    color: #6b7280;
    line-height: 1;
  }

  .workspace-item-check {
    flex-shrink: 0;
    color: #2563eb;
  }

  :global(.dark) .workspace-trigger {
    background: #1f2937;
    border-color: #374151;
  }

  :global(.dark) .workspace-trigger:hover:not(:disabled) {
    background: #111827;
    border-color: #4b5563;
  }

  :global(.dark) .workspace-name {
    color: #f9fafb;
  }

  :global(.dark) .workspace-role {
    color: #9ca3af;
  }

  :global(.dark) .workspace-dropdown {
    background: #1f2937;
    border-color: #374151;
  }

  :global(.dark) .workspace-item:hover {
    background: #111827;
  }

  :global(.dark) .workspace-item.active {
    background: #1e3a8a;
  }

  :global(.dark) .workspace-item-name {
    color: #f9fafb;
  }

  :global(.dark) .workspace-item-slug {
    color: #9ca3af;
  }
</style>

