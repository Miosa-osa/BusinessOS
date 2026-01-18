<script lang="ts">
  import { goto } from '$app/navigation';
  import { agents } from '$lib/stores/agents';
  import AgentBuilder from '$lib/components/agents/AgentBuilder.svelte';
  import type { CustomAgent } from '$lib/api/ai/types';

  // State management
  let saving = $state(false);
  let error = $state<string | null>(null);
  let showTemplates = $state(false);

  // Template options for quick start
  const templates = [
    {
      name: 'Assistant',
      description: 'General-purpose helpful assistant',
      data: {
        display_name: 'General Assistant',
        role: 'assistant',
        description: 'A helpful AI assistant for general tasks',
        system_prompt: 'You are a helpful, knowledgeable, and friendly AI assistant. Provide clear, accurate, and concise responses to user queries.',
        model: 'claude-sonnet-4.5'
      }
    },
    {
      name: 'Code Helper',
      description: 'Specialized in programming tasks',
      data: {
        display_name: 'Code Helper',
        role: 'developer',
        description: 'An AI assistant specialized in programming and software development',
        system_prompt: 'You are an expert programming assistant. Help users with code, debugging, architecture, and best practices. Always explain your reasoning and provide clean, well-documented code.',
        model: 'claude-sonnet-4.5'
      }
    },
    {
      name: 'Analyst',
      description: 'Data analysis and insights',
      data: {
        display_name: 'Data Analyst',
        role: 'analyst',
        description: 'An AI assistant specialized in data analysis and insights',
        system_prompt: 'You are a data analysis expert. Help users understand data, identify patterns, and provide actionable insights. Use clear visualizations and explanations.',
        model: 'claude-sonnet-4.5'
      }
    }
  ];

  // Handler for save action
  async function handleSave(data: Partial<CustomAgent>) {
    // Validate required fields
    if (!data.display_name?.trim()) {
      error = 'Display name is required';
      return;
    }

    if (!data.system_prompt?.trim()) {
      error = 'System prompt is required';
      return;
    }

    // Validate name if provided (alphanumeric + hyphens only)
    if (data.display_name) {
      const namePattern = /^[a-zA-Z0-9\s-]+$/;
      if (!namePattern.test(data.display_name)) {
        error = 'Name can only contain letters, numbers, spaces, and hyphens';
        return;
      }

      // Check character limits
      if (data.display_name.length > 100) {
        error = 'Display name must be 100 characters or less';
        return;
      }
    }

    if (data.system_prompt && data.system_prompt.length > 10000) {
      error = 'System prompt must be 10,000 characters or less';
      return;
    }

    if (data.description && data.description.length > 500) {
      error = 'Description must be 500 characters or less';
      return;
    }

    // Clear any previous errors
    error = null;
    saving = true;

    try {
      // Create the agent
      const newAgent = await agents.createAgent({
        ...data,
        name: data.display_name?.toLowerCase().replace(/\s+/g, '-') || 'custom-agent',
        is_active: true
      });

      // Redirect to agent detail page
      await goto(`/agents/${newAgent.id}`);
    } catch (err) {
      console.error('Failed to create agent:', err);
      error = err instanceof Error ? err.message : 'Failed to create agent. Please try again.';
      saving = false;
    }
  }

  // Handler for save & test action
  async function handleSaveAndTest(data: Partial<CustomAgent>) {
    // Validate and save first
    await handleSave(data);

    // If save was successful (no error), redirect to sandbox
    if (!error) {
      // Note: The goto in handleSave will redirect to the agent detail page
      // We could add a query param to open sandbox mode, or implement a separate flow
    }
  }

  // Handler for cancel action
  function handleCancel() {
    goto('/agents');
  }

  // Handler for template selection
  function applyTemplate(template: typeof templates[0]) {
    // This would need to be implemented with a way to pass data to AgentBuilder
    // For now, we'll keep it simple and let users manually copy template data
    showTemplates = false;
  }
</script>

<div class="h-full overflow-y-auto bg-gray-50 dark:bg-gray-900">
  <div class="max-w-5xl mx-auto px-4 py-8">
    <!-- Header with breadcrumb -->
    <div class="mb-8">
      <nav class="flex items-center space-x-2 text-sm text-gray-600 mb-4">
        <a href="/" class="hover:text-gray-900 transition-colors">Home</a>
        <span>/</span>
        <a href="/agents" class="hover:text-gray-900 transition-colors">Agents</a>
        <span>/</span>
        <span class="text-gray-900 font-medium">Create</span>
      </nav>

      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 mb-2">Create New Agent</h1>
          <p class="text-gray-600">
            Configure your custom AI agent with specific instructions and behavior
          </p>
        </div>
        <button
          type="button"
          onclick={() => goto('/agents')}
          class="btn-pill btn-pill-ghost btn-pill-sm flex items-center gap-2"
          aria-label="Back to agents"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          <span>Back</span>
        </button>
      </div>
    </div>

    <!-- Error Banner -->
    {#if error}
      <div class="mb-6 bg-red-50 border border-red-200 rounded-lg p-4 flex items-start space-x-3">
        <svg class="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
        <div class="flex-1">
          <h3 class="text-sm font-medium text-red-800">Error creating agent</h3>
          <p class="text-sm text-red-700 mt-1">{error}</p>
        </div>
        <button
          type="button"
          onclick={() => error = null}
          class="text-red-600 hover:text-red-800"
          aria-label="Dismiss error"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    {/if}

    <!-- Template Selector (Optional) -->
    {#if showTemplates}
      <div class="mb-6 bg-white rounded-lg border shadow-sm p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-900">Start with a template</h2>
          <button
            type="button"
            onclick={() => showTemplates = false}
            class="text-gray-500 hover:text-gray-700"
            aria-label="Close template selector"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          {#each templates as template}
            <button
              type="button"
              onclick={() => applyTemplate(template)}
              class="text-left p-4 border rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors"
            >
              <h3 class="font-medium text-gray-900 mb-1">{template.name}</h3>
              <p class="text-sm text-gray-600">{template.description}</p>
            </button>
          {/each}
        </div>
      </div>
    {:else}
      <div class="mb-4">
        <button
          type="button"
          onclick={() => showTemplates = true}
          class="btn-pill btn-pill-link btn-pill-sm"
        >
          Or start with a template →
        </button>
      </div>
    {/if}

    <!-- Agent Builder Component -->
    <AgentBuilder
      agent={undefined}
      onSave={handleSave}
      onCancel={handleCancel}
    />

    <!-- Loading Overlay -->
    {#if saving}
      <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div class="bg-white rounded-lg p-8 max-w-sm w-full mx-4 text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <h3 class="text-lg font-semibold text-gray-900 mb-2">Creating Agent</h3>
          <p class="text-sm text-gray-600">Please wait while we set up your custom agent...</p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Smooth transitions for interactive elements */
  button {
    transition: all 150ms ease-in-out;
  }
</style>
