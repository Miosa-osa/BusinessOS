<script lang="ts">
  import { onMount } from 'svelte';
  import { agents } from '$lib/stores/agents';
  import { goto } from '$app/navigation';
  import type { AgentPreset } from '$lib/api/ai/types';
  import PresetCard from '$lib/components/agents/PresetCard.svelte';
  import { Search, X, Sparkles, ChevronLeft } from 'lucide-svelte';

  let loading = $state(true);
  let error = $state<string | null>(null);
  let selectedCategory = $state('all');
  let searchQuery = $state('');
  let selectedPreset = $state<AgentPreset | null>(null);
  let showModal = $state(false);
  let customName = $state('');
  let creating = $state(false);

  const categories = [
    { id: 'all', label: 'All Presets', count: 0 },
    { id: 'business', label: 'Business', count: 0 },
    { id: 'creative', label: 'Creative', count: 0 },
    { id: 'technical', label: 'Technical', count: 0 },
    { id: 'research', label: 'Research', count: 0 },
    { id: 'support', label: 'Support', count: 0 }
  ];

  // Subscribe to agents store for reactive updates
  let storeState = $derived($agents);
  let allPresets = $derived(storeState.presets || []);

  // Update category counts
  $effect(() => {
    const categoryCounts = allPresets.reduce((acc, preset) => {
      const cat = preset.category || 'uncategorized';
      acc[cat] = (acc[cat] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    categories.forEach(cat => {
      if (cat.id === 'all') {
        cat.count = allPresets.length;
      } else {
        cat.count = categoryCounts[cat.id] || 0;
      }
    });
  });

  // Filter presets by category and search
  let filteredPresets = $derived(() => {
    let result = allPresets;

    // Filter by category
    if (selectedCategory !== 'all') {
      result = result.filter(p => p.category === selectedCategory);
    }

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(p =>
        p.name.toLowerCase().includes(query) ||
        p.display_name.toLowerCase().includes(query) ||
        p.description.toLowerCase().includes(query)
      );
    }

    return result;
  });

  // Separate featured and regular presets
  let featuredPresets = $derived(filteredPresets().filter((p: AgentPreset) => p.is_featured));
  let regularPresets = $derived(filteredPresets().filter((p: AgentPreset) => !p.is_featured));

  async function loadPresets() {
    loading = true;
    error = null;
    try {
      await agents.loadPresets();
    } catch (err) {
      console.error('Failed to load presets:', err);
      error = err instanceof Error ? err.message : 'Failed to load agent presets';
    } finally {
      loading = false;
    }
  }

  function handleUsePreset(preset: AgentPreset) {
    selectedPreset = preset;
    customName = preset.display_name;
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    selectedPreset = null;
    customName = '';
    creating = false;
  }

  async function handleCreateAgent() {
    if (!selectedPreset) return;

    creating = true;
    error = null;

    try {
      const newAgent = await agents.createFromPreset(
        selectedPreset.id,
        customName.trim() || undefined
      );

      closeModal();

      // Navigate to the edit page for the new agent
      await goto(`/agents/${newAgent.id}/edit`);
    } catch (err) {
      console.error('Failed to create agent from preset:', err);
      error = err instanceof Error ? err.message : 'Failed to create agent';
    } finally {
      creating = false;
    }
  }

  function getCategoryColor(category: string): string {
    const colors: Record<string, string> = {
      business: 'bg-blue-100 text-blue-700',
      creative: 'bg-purple-100 text-purple-700',
      technical: 'bg-green-100 text-green-700',
      research: 'bg-orange-100 text-orange-700',
      support: 'bg-pink-100 text-pink-700'
    };
    return colors[category] || 'bg-gray-100 text-gray-700';
  }

  onMount(() => {
    loadPresets();
  });
</script>

<div class="min-h-screen bg-gray-50">
  <div class="max-w-7xl mx-auto px-4 py-8">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-sm text-gray-600 mb-6">
      <a href="/" class="hover:text-gray-900">Home</a>
      <span>/</span>
      <a href="/agents" class="hover:text-gray-900">Agents</a>
      <span>/</span>
      <span class="text-gray-900">Presets</span>
    </nav>

    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-center gap-3 mb-2">
        <button
          onclick={() => goto('/agents')}
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          aria-label="Back to agents"
        >
          <ChevronLeft class="w-5 h-5" />
        </button>
        <h1 class="text-3xl font-bold text-gray-900">Agent Presets</h1>
      </div>
      <p class="text-gray-600 ml-14">Start with a pre-configured agent template</p>
    </div>

    <!-- Search Bar -->
    <div class="mb-6">
      <div class="relative max-w-md">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          placeholder="Search presets..."
          bind:value={searchQuery}
          class="w-full pl-10 pr-10 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        {#if searchQuery}
          <button
            onclick={() => searchQuery = ''}
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
            aria-label="Clear search"
          >
            <X class="w-5 h-5" />
          </button>
        {/if}
      </div>
    </div>

    <!-- Category Filter -->
    <div class="mb-8 flex flex-wrap gap-2">
      {#each categories as category}
        <button
          type="button"
          class="btn-pill btn-pill-sm {selectedCategory === category.id
            ? 'btn-pill-primary'
            : 'btn-pill-secondary'}"
          onclick={() => selectedCategory = category.id}
        >
          {category.label}
          <span class="ml-1.5 text-sm opacity-75">({category.count})</span>
        </button>
      {/each}
    </div>

    <!-- Loading State -->
    {#if loading}
      <div class="flex flex-col items-center justify-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
        <p class="text-gray-500">Loading presets...</p>
      </div>

    <!-- Error State -->
    {:else if error && allPresets.length === 0}
      <div class="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <p class="text-red-700 font-medium mb-2">Failed to load presets</p>
        <p class="text-red-600 text-sm mb-4">{error}</p>
        <button
          onclick={loadPresets}
          class="btn-pill btn-pill-danger btn-pill-sm"
        >
          Try Again
        </button>
      </div>

    <!-- Empty State -->
    {:else if filteredPresets.length === 0}
      <div class="bg-white border border-gray-200 rounded-lg p-12 text-center">
        <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <Search class="w-8 h-8 text-gray-400" />
        </div>
        <h3 class="text-lg font-semibold text-gray-900 mb-2">No presets found</h3>
        <p class="text-gray-600 mb-4">
          {#if searchQuery}
            No presets match "{searchQuery}"
          {:else if selectedCategory !== 'all'}
            No presets in this category
          {:else}
            No presets available
          {/if}
        </p>
        {#if searchQuery || selectedCategory !== 'all'}
          <button
            onclick={() => {
              searchQuery = '';
              selectedCategory = 'all';
            }}
            class="btn-pill btn-pill-link btn-pill-sm"
          >
            Clear filters
          </button>
        {/if}
      </div>

    <!-- Presets Grid -->
    {:else}
      <div class="space-y-8">
        <!-- Featured Section -->
        {#if featuredPresets.length > 0 && selectedCategory === 'all' && !searchQuery}
          <section>
            <div class="flex items-center gap-2 mb-4">
              <Sparkles class="w-5 h-5 text-yellow-500" />
              <h2 class="text-xl font-bold text-gray-900">Featured Templates</h2>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {#each featuredPresets as preset (preset.id)}
                <div class="relative">
                  <div class="absolute -top-2 -right-2 z-10">
                    <span class="inline-flex items-center gap-1 px-2 py-1 bg-yellow-500 text-white text-xs font-medium rounded-full shadow-md">
                      <Sparkles class="w-3 h-3" />
                      Featured
                    </span>
                  </div>
                  <PresetCard
                    preset={{
                      id: preset.id,
                      name: preset.display_name,
                      description: preset.description,
                      category: preset.category,
                      role: preset.name,
                      system_prompt: preset.system_prompt,
                      model: preset.model_preference,
                      tags: preset.capabilities
                    }}
                    onUse={() => handleUsePreset(preset)}
                  />
                </div>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Regular Presets -->
        {#if regularPresets.length > 0}
          <section>
            {#if featuredPresets.length > 0 && selectedCategory === 'all' && !searchQuery}
              <h2 class="text-xl font-bold text-gray-900 mb-4">All Templates</h2>
            {/if}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {#each regularPresets as preset (preset.id)}
                <PresetCard
                  preset={{
                    id: preset.id,
                    name: preset.display_name,
                    description: preset.description,
                    category: preset.category,
                    role: preset.name,
                    system_prompt: preset.system_prompt,
                    model: preset.model_preference,
                    tags: preset.capabilities
                  }}
                  onUse={() => handleUsePreset(preset)}
                />
              {/each}
            </div>
          </section>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Create from Preset Modal -->
{#if showModal && selectedPreset}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
    <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-6 py-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-gray-900">Create from Preset</h2>
          <button
            onclick={closeModal}
            class="text-gray-400 hover:text-gray-600 transition-colors"
            disabled={creating}
            aria-label="Close modal"
          >
            <X class="w-6 h-6" />
          </button>
        </div>
      </div>

      <div class="px-6 py-6 space-y-6">
        <!-- Preset Info -->
        <div>
          <div class="flex items-center gap-3 mb-3">
            <h3 class="text-lg font-semibold text-gray-900">{selectedPreset.display_name}</h3>
            <span class="text-xs px-2.5 py-1 rounded-full {getCategoryColor(selectedPreset.category)}">
              {selectedPreset.category}
            </span>
          </div>
          <p class="text-gray-600 text-sm">{selectedPreset.description}</p>
        </div>

        <!-- Capabilities -->
        {#if selectedPreset.capabilities && selectedPreset.capabilities.length > 0}
          <div>
            <h4 class="text-sm font-medium text-gray-700 mb-2">Capabilities</h4>
            <div class="flex flex-wrap gap-2">
              {#each selectedPreset.capabilities as capability}
                <span class="px-2.5 py-1 bg-gray-100 text-gray-700 text-xs rounded">
                  {capability}
                </span>
              {/each}
            </div>
          </div>
        {/if}

        <!-- System Prompt Preview -->
        <div>
          <h4 class="text-sm font-medium text-gray-700 mb-2">System Prompt</h4>
          <div class="bg-gray-50 rounded-lg p-4 max-h-48 overflow-y-auto">
            <pre class="text-xs text-gray-700 whitespace-pre-wrap font-mono">{selectedPreset.system_prompt}</pre>
          </div>
        </div>

        <!-- Custom Name Input -->
        <div>
          <label for="custom-name" class="block text-sm font-medium text-gray-700 mb-2">
            Agent Name (optional)
          </label>
          <input
            id="custom-name"
            type="text"
            bind:value={customName}
            placeholder={selectedPreset.display_name}
            disabled={creating}
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
          />
          <p class="mt-1 text-xs text-gray-500">
            Leave empty to use the default name
          </p>
        </div>

        <!-- Error Message -->
        {#if error}
          <div class="bg-red-50 border border-red-200 rounded-lg p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}
      </div>

      <!-- Modal Footer -->
      <div class="sticky bottom-0 bg-gray-50 border-t border-gray-200 px-6 py-4 flex items-center justify-end gap-3">
        <button
          onclick={closeModal}
          disabled={creating}
          class="btn-pill btn-pill-secondary"
        >
          Cancel
        </button>
        <button
          onclick={handleCreateAgent}
          disabled={creating}
          class="btn-pill btn-pill-primary {creating ? 'btn-pill-loading' : ''}"
        >
          {#if !creating}
            Create Agent
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
