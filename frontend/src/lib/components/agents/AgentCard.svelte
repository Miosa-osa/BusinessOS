<script lang="ts">
  import type { CustomAgent } from '$lib/api/ai/types';
  import { categoryColors } from '$lib/stores/agents';

  interface Props {
    agent: CustomAgent;
    onSelect?: (agent: CustomAgent) => void;
    onEdit?: (agent: CustomAgent) => void;
    onDelete?: (agent: CustomAgent) => void;
    variant?: 'default' | 'compact';
  }

  let { agent, onSelect, onEdit, onDelete, variant = 'default' }: Props = $props();

  let showMenu = $state(false);
  let showDeleteConfirm = $state(false);

  function getInitials(name: string | undefined) {
    if (!name) return '??';
    return name
      .split(' ')
      .map((n) => n.charAt(0))
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }

  function getCategoryColor(category?: string) {
    if (!category) return categoryColors['uncategorized'];
    return categoryColors[category] || categoryColors['custom'];
  }

  function truncateText(text: string, maxLines: number = 2) {
    // This will be handled by CSS line-clamp
    return text;
  }

  function handleSelect() {
    if (onSelect) {
      onSelect(agent);
    }
  }

  function handleEdit() {
    showMenu = false;
    onEdit?.(agent);
  }

  function handleDelete() {
    if (!showDeleteConfirm) {
      showDeleteConfirm = true;
      return;
    }
    showMenu = false;
    showDeleteConfirm = false;
    onDelete?.(agent);
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleSelect();
    }
  }

  function handleSelectClick(e: MouseEvent) {
    e.stopPropagation();
    handleSelect();
  }

  function handleMenuClick(e: MouseEvent) {
    e.stopPropagation();
    showMenu = !showMenu;
  }

  function handleMenuClose(e: MouseEvent) {
    e.stopPropagation();
    showMenu = false;
    showDeleteConfirm = false;
  }

  function handleCancelDelete(e: MouseEvent) {
    e.stopPropagation();
    showDeleteConfirm = false;
    showMenu = false;
  }
</script>

<div
  class="group bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-5 hover:shadow-lg hover:border-gray-300 dark:hover:border-gray-600 transition-all duration-200 {onSelect
    ? 'cursor-pointer'
    : ''}"
  class:compact={variant === 'compact'}
  role="button"
  tabindex={onSelect ? 0 : -1}
  onclick={handleSelect}
  onkeydown={handleKeyDown}
>
  <!-- Header: Avatar + Title + Actions -->
  <div class="flex items-start gap-4">
    <!-- Avatar -->
    <div class="flex-shrink-0">
      {#if agent.avatar}
        <img
          src={agent.avatar}
          alt={agent.display_name}
          class="w-12 h-12 rounded-full object-cover ring-2 ring-gray-100 dark:ring-gray-700"
        />
      {:else}
        <div
          class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-100 to-purple-100 dark:from-blue-900 dark:to-purple-900 flex items-center justify-center ring-2 ring-gray-100 dark:ring-gray-700"
        >
          <span class="text-lg font-semibold text-gray-700 dark:text-gray-300"
            >{getInitials(agent.display_name)}</span
          >
        </div>
      {/if}
    </div>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <!-- Title and Status -->
      <div class="flex items-start justify-between gap-2">
        <div class="flex-1 min-w-0">
          <h3 class="font-semibold text-gray-900 dark:text-white truncate">
            {agent.display_name}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-mono truncate">
            @{agent.name}
          </p>
        </div>

        <!-- Active/Inactive Indicator -->
        <div class="flex-shrink-0">
          {#if agent.is_active}
            <div
              class="flex items-center gap-1.5 text-xs text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-900/30 px-2 py-1 rounded-full"
            >
              <div class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></div>
              Active
            </div>
          {:else}
            <div
              class="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded-full"
            >
              Inactive
            </div>
          {/if}
        </div>
      </div>

      <!-- Description -->
      <p
        class="text-sm text-gray-600 dark:text-gray-400 mt-2 line-clamp-2"
        title={agent.description || 'No description'}
      >
        {agent.description || 'No description provided'}
      </p>

      <!-- Badges -->
      <div class="flex flex-wrap gap-2 mt-3">
        <!-- Category Badge -->
        {#if agent.category}
          <span class="text-xs px-2.5 py-1 rounded-full font-medium {getCategoryColor(agent.category)}">
            {agent.category}
          </span>
        {/if}

        <!-- Model Badge -->
        {#if agent.model_preference}
          <span
            class="text-xs px-2.5 py-1 rounded-full bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 font-mono"
          >
            {agent.model_preference}
          </span>
        {/if}

        <!-- Usage Count Badge -->
        {#if agent.times_used !== undefined && agent.times_used > 0}
          <span
            class="text-xs px-2.5 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 flex items-center gap-1"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
              />
            </svg>
            {agent.times_used}
          </span>
        {/if}
      </div>
    </div>
  </div>

  <!-- Divider -->
  <div class="border-t border-gray-100 dark:border-gray-700 my-4"></div>

  <!-- Actions -->
  <div class="flex items-center justify-between gap-2">
    <!-- Select/View Button -->
    {#if onSelect}
      <button
        onclick={handleSelectClick}
        class="btn-pill btn-pill-ghost btn-pill-sm flex-1"
      >
        Select
      </button>
    {/if}

    <!-- Actions Dropdown -->
    {#if onEdit || onDelete}
      <div class="relative">
        <button
          onclick={handleMenuClick}
          class="btn-pill btn-pill-icon btn-pill-ghost btn-pill-sm"
          aria-label="More actions"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"
            />
          </svg>
        </button>

        {#if showMenu}
          <div
            role="menu"
            tabindex="-1"
            class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10"
            onclick={(e) => e.stopPropagation()}
            onkeydown={(e) => e.stopPropagation()}
          >
            {#if onEdit}
              <button
                onclick={handleEdit}
                class="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2 rounded-t-lg"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                Edit
              </button>
            {/if}

            {#if onDelete}
              {#if !showDeleteConfirm}
                <button
                  onclick={handleDelete}
                  class="w-full px-4 py-2 text-left text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 flex items-center gap-2 rounded-b-lg"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                  Delete
                </button>
              {:else}
                <div class="p-3 bg-red-50 dark:bg-red-900/30 rounded-b-lg">
                  <p class="text-xs text-red-600 dark:text-red-400 mb-2">
                    Are you sure? This cannot be undone.
                  </p>
                  <div class="flex gap-2">
                    <button
                      onclick={handleDelete}
                      class="btn-pill btn-pill-danger btn-pill-xs flex-1"
                    >
                      Delete
                    </button>
                    <button
                      onclick={handleCancelDelete}
                      class="btn-pill btn-pill-ghost btn-pill-xs flex-1"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              {/if}
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Click outside to close menu -->
{#if showMenu}
  <button
    class="fixed inset-0 z-0"
    onclick={handleMenuClose}
    aria-hidden="true"
  ></button>
{/if}

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .compact {
    padding: 1rem;
  }

  .compact h3 {
    font-size: 0.9rem;
  }

  .compact .line-clamp-2 {
    font-size: 0.8rem;
  }
</style>
