<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { agents } from '$lib/stores/agents';
  import { goto } from '$app/navigation';
  import AgentBuilder from '$lib/components/agents/AgentBuilder.svelte';
  import type { CustomAgent } from '$lib/api/ai/types';

  let agent = $state<CustomAgent | null>(null);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state<string | null>(null);
  let agentId = $derived($page.params.id || '');

  async function loadAgent() {
    if (!agentId) {
      error = 'Invalid agent ID';
      loading = false;
      return;
    }

    loading = true;
    error = null;
    try {
      const result = await agents.loadAgent(agentId as string);
      if (!result) {
        error = 'Agent not found';
        agent = null;
      } else {
        agent = result;
      }
    } catch (err) {
      console.error('Failed to load agent:', err);
      error = err instanceof Error ? err.message : 'Failed to load agent';
      agent = null;
    } finally {
      loading = false;
    }
  }

  async function handleSave(data: Partial<CustomAgent>) {
    if (saving || !agent || !agentId) return;

    saving = true;
    error = null;
    try {
      await agents.updateAgent(agentId as string, data);
      // Navigate to agent detail page on success
      goto(`/agents/${agentId}`);
    } catch (err) {
      console.error('Failed to update agent:', err);
      error = err instanceof Error ? err.message : 'Failed to update agent';
      saving = false;
    }
  }

  function handleCancel() {
    goto(`/agents/${agentId}`);
  }

  onMount(() => {
    loadAgent();
  });
</script>

<div class="h-full overflow-y-auto bg-gray-50 dark:bg-gray-900">
  <div class="max-w-4xl mx-auto p-6">
    <!-- Breadcrumb -->
    <div class="mb-6">
      <nav class="flex items-center gap-2 text-sm text-gray-600">
        <a href="/" class="hover:text-gray-900">Home</a>
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <a href="/agents" class="hover:text-gray-900">Agents</a>
        {#if agent}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          <a href="/agents/{agentId}" class="hover:text-gray-900">{agent.display_name}</a>
        {/if}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-gray-900 font-medium">Edit</span>
      </nav>
    </div>

    <!-- Header -->
    <div class="mb-6">
      <button
        onclick={handleCancel}
        class="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4 transition-colors"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back to Agent
      </button>
      <h1 class="text-3xl font-bold text-gray-900">Edit Agent</h1>
      <p class="text-gray-600 mt-2">Update agent configuration and settings</p>
    </div>

    <!-- Error Banner -->
    {#if error}
      <div class="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
        <div class="flex items-start gap-3">
          <svg class="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div class="flex-1">
            <h3 class="text-sm font-semibold text-red-900 mb-1">Error</h3>
            <p class="text-sm text-red-700">{error}</p>
          </div>
          <button
            onclick={() => (error = null)}
            class="text-red-600 hover:text-red-800 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    {/if}

    <!-- Loading State -->
    {#if loading}
      <div class="flex items-center justify-center py-20">
        <div class="text-center">
          <div class="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p class="text-gray-500">Loading agent...</p>
        </div>
      </div>

    <!-- Agent Not Found -->
    {:else if !agent}
      <div class="bg-red-50 border border-red-200 rounded-lg p-8 text-center">
        <svg class="w-16 h-16 text-red-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h2 class="text-xl font-semibold text-red-900 mb-2">Agent Not Found</h2>
        <p class="text-red-700 mb-6">The agent you're trying to edit could not be found.</p>
        <div class="flex gap-3 justify-center">
          <button
            onclick={() => goto('/agents')}
            class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium"
          >
            Back to Agents
          </button>
          <button
            onclick={loadAgent}
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
          >
            Try Again
          </button>
        </div>
      </div>

    <!-- Edit Form -->
    {:else}
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        {#if saving}
          <div class="absolute inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-10 rounded-xl">
            <div class="text-center">
              <div class="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-3"></div>
              <p class="text-gray-700 font-medium">Saving changes...</p>
            </div>
          </div>
        {/if}

        <AgentBuilder
          agent={agent}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      </div>
    {/if}
  </div>
</div>
