<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { agents, categoryLabels } from '$lib/stores/agents';
  import type { CustomAgent } from '$lib/api/ai/types';
  import AgentCard from '$lib/components/agents/AgentCard.svelte';

  let searchQuery = $state('');
  let selectedCategory = $state<string | null>(null);
  let selectedStatus = $state<'active' | 'inactive' | null>(null);
  let sortBy = $state<'name' | 'created' | 'usage'>('name');
  let showDeleteDialog = $state<string | null>(null);

  const categories = ['general', 'specialist', 'system', 'custom'];
  const statusOptions = [
    { value: null, label: 'All' },
    { value: 'active' as const, label: 'Active' },
    { value: 'inactive' as const, label: 'Inactive' }
  ];

  let filteredAgents = $derived.by(() => {
    let filtered = $agents.agents;

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (agent) =>
          agent.name.toLowerCase().includes(query) ||
          agent.display_name.toLowerCase().includes(query) ||
          agent.description?.toLowerCase().includes(query)
      );
    }

    if (selectedCategory) {
      filtered = filtered.filter((agent) => agent.category === selectedCategory);
    }

    if (selectedStatus === 'active') {
      filtered = filtered.filter((agent) => agent.is_active);
    } else if (selectedStatus === 'inactive') {
      filtered = filtered.filter((agent) => !agent.is_active);
    }

    const sorted = [...filtered].sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.display_name.localeCompare(b.display_name);
        case 'created':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        case 'usage':
          return (b.times_used || 0) - (a.times_used || 0);
        default:
          return 0;
      }
    });

    return sorted;
  });

  onMount(async () => {
    await agents.loadAgents();
  });

  function handleSearch(e: Event) {
    const target = e.target as HTMLInputElement;
    searchQuery = target.value;
  }

  function handleCategoryFilter(category: string | null) {
    selectedCategory = category;
  }

  function handleStatusFilter(status: 'active' | 'inactive' | null) {
    selectedStatus = status;
  }

  function handleSortChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    sortBy = target.value as 'name' | 'created' | 'usage';
  }

  function handleSelectAgent(agent: CustomAgent) {
    goto(`/agents/${agent.id}`);
  }

  function handleEditAgent(agent: CustomAgent) {
    goto(`/agents/${agent.id}/edit`);
  }

  async function handleDeleteAgent(agent: CustomAgent) {
    showDeleteDialog = agent.id;
  }

  async function confirmDelete() {
    if (!showDeleteDialog) return;

    try {
      await agents.deleteAgent(showDeleteDialog);
      showDeleteDialog = null;
    } catch (error) {
      console.error('Failed to delete agent:', error);
    }
  }

  function cancelDelete() {
    showDeleteDialog = null;
  }

  function handleCreateAgent() {
    goto('/agents/new');
  }

  function handleBrowsePresets() {
    goto('/agents/presets');
  }

  function clearFilters() {
    searchQuery = '';
    selectedCategory = null;
    selectedStatus = null;
    sortBy = 'name';
  }
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-start justify-between mb-4">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">Custom Agents</h1>
          <p class="text-gray-600 dark:text-gray-400">
            Manage and create your custom AI agents
          </p>
        </div>
        <div class="flex gap-3">
          <button
            onclick={handleBrowsePresets}
            class="btn-pill btn-pill-secondary btn-pill-sm"
          >
            Browse Presets
          </button>
          <button
            onclick={handleCreateAgent}
            class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            Create Agent
          </button>
        </div>
      </div>

      <!-- Filters & Search -->
      <div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <div class="flex flex-col lg:flex-row gap-4">
          <!-- Search -->
          <div class="flex-1">
            <div class="relative">
              <svg
                class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <input
                type="text"
                value={searchQuery}
                oninput={handleSearch}
                placeholder="Search agents by name or description..."
                class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          <!-- Sort -->
          <div class="w-full lg:w-48">
            <select
              value={sortBy}
              onchange={handleSortChange}
              class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="name">Sort by Name</option>
              <option value="created">Sort by Created Date</option>
              <option value="usage">Sort by Usage</option>
            </select>
          </div>
        </div>

        <!-- Category Filters -->
        <div class="flex flex-wrap gap-2 mt-4">
          <button
            onclick={() => handleCategoryFilter(null)}
            class="btn-pill btn-pill-sm {selectedCategory === null ? 'btn-pill-primary' : 'btn-pill-ghost'}"
          >
            All
          </button>
          {#each categories as category}
            <button
              onclick={() => handleCategoryFilter(category)}
              class="btn-pill btn-pill-sm {selectedCategory === category ? 'btn-pill-primary' : 'btn-pill-ghost'}"
            >
              {categoryLabels[category] || category}
            </button>
          {/each}
        </div>

        <!-- Status Filters -->
        <div class="flex flex-wrap gap-2 mt-3">
          {#each statusOptions as option}
            <button
              onclick={() => handleStatusFilter(option.value)}
              class="btn-pill btn-pill-sm {selectedStatus === option.value ? 'btn-pill-success' : 'btn-pill-ghost'}"
            >
              {option.label}
            </button>
          {/each}
        </div>

        <!-- Clear Filters -->
        {#if searchQuery || selectedCategory || selectedStatus !== null || sortBy !== 'name'}
          <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
            <button
              onclick={clearFilters}
              class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium"
            >
              Clear all filters
            </button>
          </div>
        {/if}
      </div>
    </div>

    <!-- Loading State -->
    {#if $agents.loading}
      <div class="flex items-center justify-center py-20">
        <div class="text-center">
          <div class="inline-block animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-700 border-t-blue-600 mb-4"></div>
          <p class="text-gray-600 dark:text-gray-400">Loading agents...</p>
        </div>
      </div>

    <!-- Error State -->
    {:else if $agents.error}
      <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 text-center">
        <svg class="w-12 h-12 text-red-600 dark:text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">Error Loading Agents</h3>
        <p class="text-red-700 dark:text-red-300 mb-4">{$agents.error}</p>
        <button
          onclick={() => agents.loadAgents()}
          class="btn-pill btn-pill-danger btn-pill-sm"
        >
          Retry
        </button>
      </div>

    <!-- Empty State -->
    {:else if filteredAgents.length === 0}
      <div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-12 text-center">
        <svg class="w-16 h-16 text-gray-400 dark:text-gray-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
        </svg>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          {searchQuery || selectedCategory || selectedStatus !== null ? 'No agents found' : 'No agents yet'}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {searchQuery || selectedCategory || selectedStatus !== null
            ? 'Try adjusting your filters or search query'
            : 'Get started by creating your first custom agent'}
        </p>
        {#if searchQuery || selectedCategory || selectedStatus !== null}
          <button
            onclick={clearFilters}
            class="btn-pill btn-pill-link btn-pill-sm"
          >
            Clear filters
          </button>
        {:else}
          <div class="flex gap-3 justify-center">
            <button
              onclick={handleCreateAgent}
              class="btn-pill btn-pill-primary"
            >
              Create Agent
            </button>
            <button
              onclick={handleBrowsePresets}
              class="btn-pill btn-pill-secondary"
            >
              Browse Presets
            </button>
          </div>
        {/if}
      </div>

    <!-- Agent Grid -->
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredAgents as agent (agent.id)}
          <AgentCard
            {agent}
            onSelect={handleSelectAgent}
            onEdit={handleEditAgent}
            onDelete={handleDeleteAgent}
          />
        {/each}
      </div>

      <!-- Results Count -->
      <div class="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
        Showing {filteredAgents.length} {filteredAgents.length === 1 ? 'agent' : 'agents'}
        {#if $agents.agents.length !== filteredAgents.length}
          of {$agents.agents.length} total
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Delete Confirmation Dialog -->
{#if showDeleteDialog}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6">
      <div class="flex items-center gap-3 mb-4">
        <div class="flex-shrink-0 w-10 h-10 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
          <svg class="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Delete Agent</h3>
        </div>
      </div>
      <p class="text-gray-600 dark:text-gray-400 mb-6">
        Are you sure you want to delete this agent? This action cannot be undone.
      </p>
      <div class="flex gap-3 justify-end">
        <button
          onclick={cancelDelete}
          class="btn-pill btn-pill-secondary btn-pill-sm"
        >
          Cancel
        </button>
        <button
          onclick={confirmDelete}
          class="btn-pill btn-pill-danger btn-pill-sm"
        >
          Delete
        </button>
      </div>
    </div>
  </div>
{/if}
