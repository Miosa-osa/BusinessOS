<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { agents, categoryColors, categoryLabels } from '$lib/stores/agents';
  import { goto } from '$app/navigation';
  import type { CustomAgent } from '$lib/api/ai/types';
  import AgentSandbox from '$lib/components/agents/AgentSandbox.svelte';

  let agent: CustomAgent | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let agentId = $derived($page.params.id);
  let activeTab = $state<'overview' | 'configuration' | 'testing' | 'history'>('overview');
  let showDeleteConfirm = $state(false);
  let isTogglingActive = $state(false);

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

  function getCategoryLabel(category?: string) {
    if (!category) return categoryLabels['uncategorized'];
    return categoryLabels[category] || category;
  }

  function formatDate(dateString: string) {
    return new Date(dateString).toLocaleString();
  }

  async function loadAgent() {
    if (!agentId) return;

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

  async function handleToggleActive() {
    if (!agent || isTogglingActive || !agentId) return;

    isTogglingActive = true;
    try {
      await agents.updateAgent(agentId as string, { is_active: !agent.is_active });
      // Reload agent to get updated state
      await loadAgent();
    } catch (err) {
      console.error('Failed to toggle agent status:', err);
      alert('Failed to toggle agent status');
    } finally {
      isTogglingActive = false;
    }
  }

  async function handleDelete() {
    if (!agentId) return;

    if (!showDeleteConfirm) {
      showDeleteConfirm = true;
      return;
    }

    try {
      await agents.deleteAgent(agentId as string);
      goto('/agents');
    } catch (err) {
      console.error('Failed to delete agent:', err);
      alert('Failed to delete agent');
      showDeleteConfirm = false;
    }
  }

  function handleEdit() {
    goto(`/agents/${agentId}/edit`);
  }

  async function handleClone() {
    if (!agent) return;

    try {
      const clonedAgent = await agents.createAgent({
        name: `${agent.name}-copy`,
        display_name: `${agent.display_name} (Copy)`,
        description: agent.description,
        avatar: agent.avatar,
        system_prompt: agent.system_prompt,
        model_preference: agent.model_preference,
        temperature: agent.temperature,
        max_tokens: agent.max_tokens,
        capabilities: agent.capabilities,
        tools_enabled: agent.tools_enabled,
        context_sources: agent.context_sources,
        thinking_enabled: agent.thinking_enabled,
        streaming_enabled: agent.streaming_enabled,
        category: agent.category,
        is_active: false
      });

      if (clonedAgent) {
        goto(`/agents/${clonedAgent.id}`);
      }
    } catch (err) {
      console.error('Failed to clone agent:', err);
      alert('Failed to clone agent');
    }
  }

  function handleCopySystemPrompt() {
    if (!agent) return;
    navigator.clipboard.writeText(agent.system_prompt);
    alert('System prompt copied to clipboard');
  }

  onMount(() => {
    loadAgent();
  });
</script>

<div class="min-h-screen bg-gray-50">
  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="text-center">
        <div class="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p class="text-gray-500">Loading agent...</p>
      </div>
    </div>
  {:else if error || !agent}
    <div class="max-w-4xl mx-auto p-6">
      <div class="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <svg class="w-12 h-12 text-red-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h2 class="text-xl font-semibold text-red-900 mb-2">Agent Not Found</h2>
        <p class="text-red-700 mb-4">{error || 'The requested agent could not be found.'}</p>
        <button
          onclick={() => goto('/agents')}
          class="btn-pill btn-pill-danger btn-pill-sm"
        >
          Back to Agents
        </button>
      </div>
    </div>
  {:else}
    <div class="max-w-7xl mx-auto p-6">
      <!-- Breadcrumb -->
      <div class="mb-6">
        <nav class="flex items-center gap-2 text-sm text-gray-600">
          <a href="/" class="hover:text-gray-900">Home</a>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          <a href="/agents" class="hover:text-gray-900">Agents</a>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          <span class="text-gray-900 font-medium">{agent.display_name}</span>
        </nav>
      </div>

      <!-- Header -->
      <div class="bg-white rounded-xl border border-gray-200 p-6 mb-6">
        <div class="flex items-start justify-between gap-6">
          <!-- Left: Avatar + Info -->
          <div class="flex items-start gap-4 flex-1">
            <!-- Avatar -->
            <div class="flex-shrink-0">
              {#if agent.avatar}
                <img
                  src={agent.avatar}
                  alt={agent.display_name}
                  class="w-20 h-20 rounded-full object-cover ring-4 ring-gray-100"
                />
              {:else}
                <div class="w-20 h-20 rounded-full bg-gradient-to-br from-blue-100 to-purple-100 flex items-center justify-center ring-4 ring-gray-100">
                  <span class="text-2xl font-semibold text-gray-700">{getInitials(agent.display_name)}</span>
                </div>
              {/if}
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-3 mb-2">
                <h1 class="text-3xl font-bold text-gray-900">{agent.display_name}</h1>

                <!-- Category Badge -->
                {#if agent.category}
                  <span class="text-sm px-3 py-1 rounded-full font-medium {getCategoryColor(agent.category)}">
                    {getCategoryLabel(agent.category)}
                  </span>
                {/if}
              </div>

              <p class="text-gray-500 font-mono text-sm mb-3">@{agent.name}</p>

              <!-- Active/Inactive Toggle -->
              <div class="flex items-center gap-3">
                <button
                  onclick={handleToggleActive}
                  disabled={isTogglingActive}
                  class="flex items-center gap-2 px-4 py-2 rounded-lg transition-colors {agent.is_active
                    ? 'bg-green-50 text-green-700 hover:bg-green-100'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'} disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <div class="w-2.5 h-2.5 rounded-full {agent.is_active ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}"></div>
                  <span class="text-sm font-medium">{agent.is_active ? 'Active' : 'Inactive'}</span>
                </button>

                {#if agent.times_used !== undefined && agent.times_used > 0}
                  <div class="flex items-center gap-2 text-sm text-gray-600 bg-gray-100 px-3 py-2 rounded-lg">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                    <span>{agent.times_used} uses</span>
                  </div>
                {/if}
              </div>
            </div>
          </div>

          <!-- Right: Action Buttons -->
          <div class="flex gap-2">
            <button
              onclick={handleEdit}
              class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit
            </button>

            <button
              onclick={handleClone}
              class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              Clone
            </button>

            <button
              onclick={() => (activeTab = 'testing')}
              class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Test
            </button>

            {#if !showDeleteConfirm}
              <button
                onclick={handleDelete}
                class="btn-pill btn-pill-danger btn-pill-sm flex items-center gap-2"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                Delete
              </button>
            {:else}
              <div class="flex items-center gap-2 bg-red-50 px-3 py-2 rounded-lg border border-red-200">
                <span class="text-sm text-red-700 font-medium">Confirm?</span>
                <button
                  onclick={handleDelete}
                  class="btn-pill btn-pill-danger btn-pill-xs"
                >
                  Yes
                </button>
                <button
                  onclick={() => (showDeleteConfirm = false)}
                  class="btn-pill btn-pill-secondary btn-pill-xs"
                >
                  No
                </button>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <!-- Tab Navigation -->
        <div class="border-b border-gray-200">
          <nav class="flex gap-1 px-6">
            <button
              onclick={() => (activeTab = 'overview')}
              class="px-4 py-3 text-sm font-medium border-b-2 transition-colors {activeTab === 'overview'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'}"
            >
              Overview
            </button>
            <button
              onclick={() => (activeTab = 'configuration')}
              class="px-4 py-3 text-sm font-medium border-b-2 transition-colors {activeTab === 'configuration'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'}"
            >
              Configuration
            </button>
            <button
              onclick={() => (activeTab = 'testing')}
              class="px-4 py-3 text-sm font-medium border-b-2 transition-colors {activeTab === 'testing'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'}"
            >
              Testing
            </button>
            <button
              onclick={() => (activeTab = 'history')}
              class="px-4 py-3 text-sm font-medium border-b-2 transition-colors {activeTab === 'history'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'}"
            >
              History & Stats
            </button>
          </nav>
        </div>

        <!-- Tab Content -->
        <div class="p-6">
          {#if activeTab === 'overview'}
            <div class="space-y-6">
              <!-- Description -->
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Description</h3>
                <p class="text-gray-700">
                  {agent.description || 'No description provided'}
                </p>
              </div>

              <!-- System Prompt -->
              <div>
                <div class="flex items-center justify-between mb-2">
                  <h3 class="text-lg font-semibold text-gray-900">System Prompt</h3>
                  <button
                    onclick={handleCopySystemPrompt}
                    class="btn-pill btn-pill-ghost btn-pill-xs flex items-center gap-2"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Copy
                  </button>
                </div>
                <div class="bg-gray-50 border border-gray-200 rounded-lg p-4 font-mono text-sm text-gray-800 whitespace-pre-wrap max-h-96 overflow-y-auto">
                  {agent.system_prompt}
                </div>
              </div>

              <!-- Capabilities -->
              {#if agent.capabilities && agent.capabilities.length > 0}
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 mb-3">Capabilities</h3>
                  <div class="flex flex-wrap gap-2">
                    {#each agent.capabilities as capability}
                      <span class="px-3 py-1.5 bg-purple-50 text-purple-700 rounded-lg text-sm font-medium">
                        {capability}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}

              <!-- Tools Enabled -->
              {#if agent.tools_enabled && agent.tools_enabled.length > 0}
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 mb-3">Tools Enabled</h3>
                  <div class="flex flex-wrap gap-2">
                    {#each agent.tools_enabled as tool}
                      <span class="px-3 py-1.5 bg-blue-50 text-blue-700 rounded-lg text-sm font-medium flex items-center gap-2">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        {tool}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}

              <!-- Context Sources -->
              {#if agent.context_sources && agent.context_sources.length > 0}
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 mb-3">Context Sources</h3>
                  <div class="flex flex-wrap gap-2">
                    {#each agent.context_sources as source}
                      <span class="px-3 py-1.5 bg-amber-50 text-amber-700 rounded-lg text-sm font-medium">
                        {source}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>

          {:else if activeTab === 'configuration'}
            <div class="space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Model Preference -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Model Preference</h4>
                  <p class="text-lg font-semibold text-gray-900">
                    {agent.model_preference || 'Default'}
                  </p>
                </div>

                <!-- Temperature -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Temperature</h4>
                  <p class="text-lg font-semibold text-gray-900">
                    {agent.temperature ?? 0.7}
                  </p>
                </div>

                <!-- Max Tokens -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Max Tokens</h4>
                  <p class="text-lg font-semibold text-gray-900">
                    {agent.max_tokens || 'Default'}
                  </p>
                </div>

                <!-- Thinking Enabled -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Thinking Mode</h4>
                  <div class="flex items-center gap-2">
                    <div class="w-2 h-2 rounded-full {agent.thinking_enabled ? 'bg-green-500' : 'bg-gray-400'}"></div>
                    <p class="text-lg font-semibold text-gray-900">
                      {agent.thinking_enabled ? 'Enabled' : 'Disabled'}
                    </p>
                  </div>
                </div>

                <!-- Streaming Enabled -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Streaming</h4>
                  <div class="flex items-center gap-2">
                    <div class="w-2 h-2 rounded-full {agent.streaming_enabled ? 'bg-green-500' : 'bg-gray-400'}"></div>
                    <p class="text-lg font-semibold text-gray-900">
                      {agent.streaming_enabled ? 'Enabled' : 'Disabled'}
                    </p>
                  </div>
                </div>

                <!-- Created At -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Created</h4>
                  <p class="text-sm text-gray-900">
                    {formatDate(agent.created_at)}
                  </p>
                </div>

                <!-- Updated At -->
                <div class="bg-gray-50 rounded-lg p-4">
                  <h4 class="text-sm font-medium text-gray-600 mb-1">Last Updated</h4>
                  <p class="text-sm text-gray-900">
                    {formatDate(agent.updated_at)}
                  </p>
                </div>
              </div>
            </div>

          {:else if activeTab === 'testing'}
            <div>
              <div class="mb-4">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Test Agent</h3>
                <p class="text-sm text-gray-600">
                  Send test messages to see how your agent responds in real-time.
                </p>
              </div>
              <AgentSandbox agentId={agent.id} />
            </div>

          {:else if activeTab === 'history'}
            <div class="space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <!-- Usage Count -->
                <div class="bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg p-6">
                  <div class="flex items-center justify-between mb-2">
                    <h4 class="text-sm font-medium text-blue-900">Total Uses</h4>
                    <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                  </div>
                  <p class="text-3xl font-bold text-blue-900">
                    {agent.times_used || 0}
                  </p>
                </div>

                <!-- Last Used -->
                <div class="bg-gradient-to-br from-purple-50 to-purple-100 rounded-lg p-6">
                  <div class="flex items-center justify-between mb-2">
                    <h4 class="text-sm font-medium text-purple-900">Last Used</h4>
                    <svg class="w-8 h-8 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <p class="text-sm text-purple-900">
                    {agent.updated_at ? new Date(agent.updated_at).toLocaleDateString() : 'Never'}
                  </p>
                </div>

                <!-- Status -->
                <div class="bg-gradient-to-br from-green-50 to-green-100 rounded-lg p-6">
                  <div class="flex items-center justify-between mb-2">
                    <h4 class="text-sm font-medium text-green-900">Status</h4>
                    <svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <p class="text-lg font-semibold text-green-900">
                    {agent.is_active ? 'Active' : 'Inactive'}
                  </p>
                </div>
              </div>

              <!-- Placeholder for future stats -->
              <div class="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
                <svg class="w-12 h-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <h4 class="text-lg font-semibold text-gray-700 mb-2">Detailed Analytics Coming Soon</h4>
                <p class="text-sm text-gray-600">
                  We're working on providing detailed usage analytics, token consumption, cost estimates, and performance metrics.
                </p>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
