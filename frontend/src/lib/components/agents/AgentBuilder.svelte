<script lang="ts">
  import type { CustomAgent } from '$lib/api/ai/types';
  import SystemPromptEditor from './SystemPromptEditor.svelte';
  import {
    validateAgentForm,
    getCharacterCountStatus,
    VALIDATION_LIMITS,
    ALLOWED_CATEGORIES,
    type ValidationError
  } from '$lib/utils/agentValidation';

  interface Props {
    agent?: CustomAgent;
    onSave: (agent: Partial<CustomAgent>) => void;
    onCancel: () => void;
  }

  let { agent, onSave, onCancel }: Props = $props();

  // ============ IDENTITY ============
  let displayName = $state(agent?.display_name || '');
  let name = $state(agent?.name || '');
  let description = $state(agent?.description || '');
  let avatar = $state(agent?.avatar || '');
  let category = $state(agent?.category || '');

  // ============ BEHAVIOR ============
  let welcomeMessage = $state(agent?.welcome_message || '');
  let suggestedPrompts = $state<string[]>(agent?.suggested_prompts || []);
  let newPrompt = $state('');

  // ============ CONFIGURATION ============
  let modelPreference = $state(agent?.model_preference || '');
  let temperature = $state(agent?.temperature ?? 0.7);
  let maxTokens = $state(agent?.max_tokens || 4000);

  // ============ TOOLS & CAPABILITIES ============
  let toolsEnabled = $state<string[]>(agent?.tools_enabled || []);
  let capabilities = $state<string[]>(agent?.capabilities || []);
  let newCapability = $state('');

  // ============ CONTEXT SOURCES ============
  let contextSources = $state<string[]>(agent?.context_sources || []);
  let newContextSource = $state('');

  // ============ ADVANCED FEATURES ============
  let thinkingEnabled = $state(agent?.thinking_enabled ?? true);
  let streamingEnabled = $state(agent?.streaming_enabled ?? true);
  let applyPersonalization = $state(agent?.apply_personalization ?? false);

  // ============ ACCESS CONTROL ============
  let isPublic = $state(agent?.is_public ?? false);
  let isFeatured = $state(agent?.is_featured ?? false);
  let isActive = $state(agent?.is_active ?? true);

  // ============ SYSTEM PROMPT ============
  let systemPrompt = $state(agent?.system_prompt || '');

  // ============ VALIDATION & UI ============
  let errors = $state<Record<string, string>>({});
  let validationErrors = $state<ValidationError[]>([]);
  let isSubmitting = $state(false);
  let showAdvanced = $state(false);
  let showValidationSummary = $state(false);

  // ============ CONSTANTS ============
  const availableModels = [
    { value: '', label: 'Use default model' },
    { value: 'claude-sonnet-4.5', label: 'Claude Sonnet 4.5 (Recommended)' },
    { value: 'claude-opus-4.5', label: 'Claude Opus 4.5 (Most Capable)' },
    { value: 'claude-haiku-4', label: 'Claude Haiku 4 (Fast)' },
    { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
    { value: 'gpt-4', label: 'GPT-4' }
  ];

  const availableTools = [
    { id: 'web-search', name: 'Web Search', description: 'Search the internet' },
    { id: 'calculator', name: 'Calculator', description: 'Mathematical calculations' },
    { id: 'code-execution', name: 'Code Execution', description: 'Run code in sandbox' },
    { id: 'file-access', name: 'File Access', description: 'Read/write files' },
    { id: 'database-query', name: 'Database Query', description: 'Query databases' },
    { id: 'api-calls', name: 'API Calls', description: 'Make HTTP requests' }
  ];

  const commonSources = [
    'workspace-documents',
    'project-files',
    'conversation-history',
    'user-profile',
    'team-knowledge-base'
  ];

  // ============ VALIDATION ============
  function validateForm(): boolean {
    // Build agent data for validation
    const agentData: Partial<CustomAgent> = {
      display_name: displayName,
      name,
      description: description || undefined,
      avatar: avatar || undefined,
      system_prompt: systemPrompt,
      category: category || undefined,
      welcome_message: welcomeMessage || undefined,
      suggested_prompts: suggestedPrompts.length > 0 ? suggestedPrompts : undefined,
      temperature,
      max_tokens: maxTokens
    };

    // Use the validation utility
    const result = validateAgentForm(agentData);
    validationErrors = result.errors;

    // Convert to legacy errors format for existing UI
    const newErrors: Record<string, string> = {};
    result.errors.forEach(error => {
      newErrors[error.field] = error.message;
    });

    errors = newErrors;
    showValidationSummary = !result.valid;
    return result.valid;
  }

  // Real-time validation for specific fields
  function validateFieldRealtime(field: keyof CustomAgent, value: any) {
    const partial: Partial<CustomAgent> = { [field]: value };
    const result = validateAgentForm(partial);
    const error = result.errors.find(e => e.field === field);

    if (error) {
      errors[field] = error.message;
    } else {
      delete errors[field];
      errors = { ...errors }; // Trigger reactivity
    }
  }

  // ============ EVENT HANDLERS ============
  function handleSubmit(e: Event) {
    e.preventDefault();

    // Validate
    if (!validateForm()) {
      // Scroll to first error
      const firstError = Object.keys(errors)[0];
      const element = document.querySelector(`[name="${firstError}"]`);
      element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }

    isSubmitting = true;

    // Build complete agent data
    const agentData: Partial<CustomAgent> = {
      display_name: displayName,
      name,
      description: description || undefined,
      avatar: avatar || undefined,
      system_prompt: systemPrompt,
      category: category || undefined,

      // Behavior
      welcome_message: welcomeMessage || undefined,
      suggested_prompts: suggestedPrompts.length > 0 ? suggestedPrompts : undefined,

      // Configuration
      model_preference: modelPreference || undefined,
      temperature,
      max_tokens: maxTokens,

      // Tools & Capabilities
      tools_enabled: toolsEnabled.length > 0 ? toolsEnabled : undefined,
      capabilities: capabilities.length > 0 ? capabilities : undefined,
      context_sources: contextSources.length > 0 ? contextSources : undefined,

      // Advanced
      thinking_enabled: thinkingEnabled,
      streaming_enabled: streamingEnabled,
      apply_personalization: applyPersonalization,

      // Access
      is_public: isPublic,
      is_featured: isFeatured,
      is_active: isActive
    };

    onSave(agentData);
  }

  function addSuggestedPrompt() {
    if (newPrompt.trim()) {
      suggestedPrompts = [...suggestedPrompts, newPrompt.trim()];
      newPrompt = '';
    }
  }

  function removeSuggestedPrompt(index: number) {
    suggestedPrompts = suggestedPrompts.filter((_, i) => i !== index);
  }

  function addCapability() {
    if (newCapability.trim() && !capabilities.includes(newCapability.trim())) {
      capabilities = [...capabilities, newCapability.trim()];
      newCapability = '';
    }
  }

  function removeCapability(index: number) {
    capabilities = capabilities.filter((_, i) => i !== index);
  }

  function toggleTool(toolId: string) {
    if (toolsEnabled.includes(toolId)) {
      toolsEnabled = toolsEnabled.filter(t => t !== toolId);
    } else {
      toolsEnabled = [...toolsEnabled, toolId];
    }
  }

  function toggleContextSource(source: string) {
    if (contextSources.includes(source)) {
      contextSources = contextSources.filter(s => s !== source);
    } else {
      contextSources = [...contextSources, source];
    }
  }

  function addCustomContextSource() {
    if (newContextSource.trim() && !contextSources.includes(newContextSource.trim())) {
      contextSources = [...contextSources, newContextSource.trim()];
      newContextSource = '';
    }
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-4xl mx-auto">
  <h2 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">
    {agent ? 'Edit Agent' : 'Create New Agent'}
  </h2>

  <!-- Error Summary -->
  {#if showValidationSummary && validationErrors.length > 0}
    <div role="alert" aria-live="polite" class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
      <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
        <div class="flex-1">
          <p class="text-sm font-medium text-red-800 dark:text-red-200 mb-2">
            Please fix {validationErrors.length} error(s) before saving:
          </p>
          <ul class="text-sm text-red-700 dark:text-red-300 space-y-1 list-disc list-inside">
            {#each validationErrors as error}
              <li>
                <span class="font-medium">{error.field.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</span>
                {error.message}
              </li>
            {/each}
          </ul>
        </div>
        <button
          type="button"
          onclick={() => showValidationSummary = false}
          class="text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
          aria-label="Dismiss"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-6">
    <!-- SECTION 1: IDENTITY -->
    <div>
      <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Identity</h3>

      <div class="space-y-4">
        <!-- Display Name -->
        <div>
          <label for="agent-display-name" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">
            Display Name <span class="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="agent-display-name"
            name="displayName"
            autocomplete="off"
            bind:value={displayName}
            oninput={() => validateFieldRealtime('display_name', displayName)}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.display_name ? 'border-red-500 dark:border-red-400' : ''}"
            placeholder="My Custom Agent"
            required
          />
          {#if errors.display_name}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.display_name}</p>
          {/if}
          {#if true}
            {@const displayNameStatus = getCharacterCountStatus(displayName.length, VALIDATION_LIMITS.DISPLAY_NAME_MAX)}
            <p class="text-xs {displayNameStatus.statusClass} mt-1">
              {displayNameStatus.current} / {displayNameStatus.max} characters
            </p>
          {/if}
        </div>

        <!-- Name (ID) -->
        <div>
          <label for="agent-name" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">
            Name (ID) <span class="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="agent-name"
            name="name"
            autocomplete="off"
            bind:value={name}
            oninput={() => validateFieldRealtime('name', name)}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 font-mono text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.name ? 'border-red-500 dark:border-red-400' : ''}"
            placeholder="my-custom-agent"
            pattern="[a-z0-9\-]+"
            required
          />
          {#if errors.name}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.name}</p>
          {:else}
            {@const nameStatus = getCharacterCountStatus(name.length, VALIDATION_LIMITS.NAME_MAX)}
            <p class="text-xs {nameStatus.statusClass} mt-1">
              {nameStatus.current} / {nameStatus.max} characters - Lowercase letters, numbers, and hyphens only
            </p>
          {/if}
        </div>

        <!-- Description -->
        <div>
          <label for="agent-description" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Description</label>
          <textarea
            id="agent-description"
            name="description"
            autocomplete="off"
            bind:value={description}
            oninput={() => validateFieldRealtime('description', description)}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.description ? 'border-red-500 dark:border-red-400' : ''}"
            rows="2"
            placeholder="Brief description of the agent's purpose"
          ></textarea>
          {#if errors.description}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.description}</p>
          {/if}
          {#if true}
            {@const descStatus = getCharacterCountStatus(description.length, VALIDATION_LIMITS.DESCRIPTION_MAX)}
            <p class="text-xs {descStatus.statusClass} mt-1">
              {descStatus.current} / {descStatus.max} characters
            </p>
          {/if}
        </div>

        <!-- Avatar URL -->
        <div>
          <label for="agent-avatar" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Avatar URL (optional)</label>
          <div class="flex gap-3">
            <input
              type="url"
              id="agent-avatar"
              name="avatar"
              autocomplete="off"
              bind:value={avatar}
              oninput={() => validateFieldRealtime('avatar', avatar)}
              class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.avatar ? 'border-red-500 dark:border-red-400' : ''}"
              placeholder="https://example.com/avatar.png"
            />
            {#if avatar && !errors.avatar}
              <img
                src={avatar}
                alt="Avatar preview"
                class="w-12 h-12 rounded-full object-cover border-2 border-gray-200 dark:border-gray-600"
                onerror={() => {
                  errors.avatar = 'Failed to load image from URL';
                  errors = { ...errors };
                }}
              />
            {/if}
          </div>
          {#if errors.avatar}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.avatar}</p>
          {:else}
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Provide a publicly accessible image URL</p>
          {/if}
        </div>

        <!-- Category -->
        <div>
          <label for="agent-category" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Category</label>
          <select
            id="agent-category"
            name="category"
            autocomplete="off"
            bind:value={category}
            oninput={() => validateFieldRealtime('category', category)}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.category ? 'border-red-500 dark:border-red-400' : ''}"
          >
            <option value="">Select category</option>
            {#each ALLOWED_CATEGORIES as cat}
              <option value={cat}>{cat.charAt(0).toUpperCase() + cat.slice(1)}</option>
            {/each}
          </select>
          {#if errors.category}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.category}</p>
          {/if}
        </div>
      </div>
    </div>

    <!-- SECTION 2: BEHAVIOR -->
    <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Behavior & Interaction</h3>

      <div class="space-y-4">
        <!-- Welcome Message -->
        <div>
          <label for="agent-welcome-message" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">
            Welcome Message
            <span class="text-xs text-gray-500 dark:text-gray-400 font-normal ml-1">(First message shown to users)</span>
          </label>
          <textarea
            id="agent-welcome-message"
            name="welcomeMessage"
            autocomplete="off"
            bind:value={welcomeMessage}
            oninput={() => validateFieldRealtime('welcome_message', welcomeMessage)}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.welcome_message ? 'border-red-500 dark:border-red-400' : ''}"
            rows="3"
            placeholder="Hello! I'm here to help you with..."
          ></textarea>
          {#if errors.welcome_message}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.welcome_message}</p>
          {/if}
          {#if true}
            {@const charStatus = getCharacterCountStatus(welcomeMessage.length, VALIDATION_LIMITS.WELCOME_MESSAGE_MAX)}
            <div class="flex justify-between items-center mt-1">
            <p class="text-xs {charStatus.statusClass}">
              {charStatus.current} / {charStatus.max} characters
              {#if charStatus.isNearLimit && !charStatus.isOverLimit}
                <span class="text-orange-600 dark:text-orange-400">({charStatus.remaining} remaining)</span>
              {/if}
              {#if charStatus.isOverLimit}
                <span class="text-red-600 dark:text-red-400 font-semibold">({Math.abs(charStatus.remaining)} over limit!)</span>
              {/if}
            </p>
          </div>
          {/if}
        </div>

        <!-- Suggested Prompts -->
        <div>
          <div class="flex items-center justify-between mb-1">
            <label for="agent-new-prompt" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Suggested Prompts
              <span class="text-xs text-gray-500 dark:text-gray-400 font-normal ml-1">(Quick start options)</span>
            </label>
            {#if true}
              {@const promptCountStatus = getCharacterCountStatus(suggestedPrompts.length, VALIDATION_LIMITS.SUGGESTED_PROMPTS_MAX)}
              <span class="text-xs {promptCountStatus.statusClass}">
                {promptCountStatus.current} / {promptCountStatus.max} prompts
              </span>
            {/if}
          </div>

          {#if errors.suggested_prompts}
            <p class="text-xs text-red-600 dark:text-red-400 mb-2">{errors.suggested_prompts}</p>
          {/if}

          <!-- Existing prompts -->
          {#if suggestedPrompts.length > 0}
            <div class="space-y-2 mb-2">
              {#each suggestedPrompts as prompt, index}
                {@const promptCharStatus = getCharacterCountStatus(prompt.length, VALIDATION_LIMITS.SUGGESTED_PROMPT_MAX)}
                <div class="bg-gray-50 dark:bg-gray-700 px-3 py-2 rounded border {errors[`suggested_prompt_${index}`] ? 'border-red-500 dark:border-red-400' : 'border-gray-200 dark:border-gray-600'}">
                  <div class="flex gap-2 items-start">
                    <span class="flex-1 text-sm text-gray-900 dark:text-gray-100 break-words">{prompt}</span>
                    <button
                      type="button"
                      onclick={() => removeSuggestedPrompt(index)}
                      class="text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300 text-sm flex-shrink-0"
                      title="Remove"
                    >
                      ✕
                    </button>
                  </div>
                  <div class="flex justify-between items-center mt-1">
                    <span class="text-xs {promptCharStatus.statusClass}">
                      {promptCharStatus.current} / {promptCharStatus.max} characters
                    </span>
                  </div>
                  {#if errors[`suggested_prompt_${index}`]}
                    <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors[`suggested_prompt_${index}`]}</p>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}

          <!-- Add new prompt -->
          <div class="space-y-2">
            <div class="flex gap-2">
              <div class="flex-1">
                <input
                  type="text"
                  id="agent-new-prompt"
                  name="newPrompt"
                  autocomplete="off"
                  bind:value={newPrompt}
                  onkeydown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      addSuggestedPrompt();
                    }
                  }}
                  disabled={suggestedPrompts.length >= VALIDATION_LIMITS.SUGGESTED_PROMPTS_MAX}
                  class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  placeholder={suggestedPrompts.length >= VALIDATION_LIMITS.SUGGESTED_PROMPTS_MAX ? "Maximum prompts reached" : "Type a prompt and press Enter"}
                />
                {#if newPrompt.length > 0}
                  {@const newPromptCharStatus = getCharacterCountStatus(newPrompt.length, VALIDATION_LIMITS.SUGGESTED_PROMPT_MAX)}
                  <p class="text-xs {newPromptCharStatus.statusClass} mt-1">
                    {newPromptCharStatus.current} / {newPromptCharStatus.max} characters
                  </p>
                {/if}
              </div>
              <button
                type="button"
                onclick={addSuggestedPrompt}
                disabled={suggestedPrompts.length >= VALIDATION_LIMITS.SUGGESTED_PROMPTS_MAX || !newPrompt.trim()}
                class="btn-pill btn-pill-soft btn-pill-sm"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- SECTION 3: CONFIGURATION -->
    <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Model Configuration</h3>

      <div class="space-y-4">
        <!-- Model Preference -->
        <div>
          <label for="agent-model-preference" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Model Preference</label>
          <select
            id="agent-model-preference"
            name="modelPreference"
            autocomplete="off"
            bind:value={modelPreference}
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
          >
            {#each availableModels as model}
              <option value={model.value}>{model.label}</option>
            {/each}
          </select>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Leave default to use system-wide model setting</p>
        </div>

        <!-- Temperature Slider -->
        <div>
          <div class="flex justify-between items-center mb-1">
            <label for="agent-temperature" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Temperature</label>
            <span class="text-sm font-mono bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded text-gray-900 dark:text-white">
              {temperature.toFixed(2)}
            </span>
          </div>
          <input
            type="range"
            id="agent-temperature"
            name="temperature"
            bind:value={temperature}
            oninput={() => validateFieldRealtime('temperature', temperature)}
            min="0"
            max="2"
            step="0.01"
            class="w-full"
          />
          <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
            <span>Precise (0.0)</span>
            <span>Balanced (1.0)</span>
            <span>Creative (2.0)</span>
          </div>
          {#if errors.temperature}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.temperature}</p>
          {:else}
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
              Higher values make output more random and creative. Range: {VALIDATION_LIMITS.TEMPERATURE_MIN} - {VALIDATION_LIMITS.TEMPERATURE_MAX}
            </p>
          {/if}
        </div>

        <!-- Max Tokens -->
        <div>
          <label for="agent-max-tokens" class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Max Tokens</label>
          <input
            type="number"
            id="agent-max-tokens"
            name="maxTokens"
            autocomplete="off"
            bind:value={maxTokens}
            oninput={() => validateFieldRealtime('max_tokens', maxTokens)}
            min={VALIDATION_LIMITS.MAX_TOKENS_MIN}
            max={VALIDATION_LIMITS.MAX_TOKENS_MAX}
            step="100"
            class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 {errors.max_tokens ? 'border-red-500 dark:border-red-400' : ''}"
          />
          {#if errors.max_tokens}
            <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.max_tokens}</p>
          {:else}
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Maximum response length ({VALIDATION_LIMITS.MAX_TOKENS_MIN}-{VALIDATION_LIMITS.MAX_TOKENS_MAX}). ~4 characters = 1 token
            </p>
          {/if}
        </div>
      </div>
    </div>

    <!-- SECTION 4: SYSTEM PROMPT -->
    <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          System Prompt <span class="text-red-500">*</span>
        </h3>
        {#if true}
          {@const systemPromptStatus = getCharacterCountStatus(systemPrompt.length, VALIDATION_LIMITS.SYSTEM_PROMPT_MAX)}
          <span class="text-sm {systemPromptStatus.statusClass}">
            {systemPromptStatus.current} / {systemPromptStatus.max} characters
          </span>
        {/if}
      </div>
      <SystemPromptEditor
        value={systemPrompt}
        onChange={(newValue) => {
          systemPrompt = newValue;
          validateFieldRealtime('system_prompt', newValue);
        }}
        placeholder="Define how this agent should behave, respond, and interact with users..."
        rows={12}
        maxLength={VALIDATION_LIMITS.SYSTEM_PROMPT_MAX}
      />
      {#if errors.system_prompt}
        <p class="text-xs text-red-600 dark:text-red-400 mt-1">{errors.system_prompt}</p>
      {/if}
    </div>

    <!-- Advanced Sections Toggle -->
    <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
      <button
        type="button"
        onclick={() => showAdvanced = !showAdvanced}
        class="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
      >
        <svg
          class="w-4 h-4 transition-transform duration-200"
          class:rotate-90={showAdvanced}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        {showAdvanced ? 'Hide' : 'Show'} Advanced Configuration
      </button>
    </div>

    {#if showAdvanced}
      <!-- SECTION 5: TOOLS & CAPABILITIES -->
      <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Tools & Capabilities</h3>

        <div class="space-y-4">
          <!-- Tools Enabled -->
          <div>
            <div class="block text-sm font-medium mb-2 text-gray-700 dark:text-gray-300">Enabled Tools</div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mb-3">
              {#each availableTools as tool}
                <label
                  for="agent-tool-{tool.id}"
                  class="flex items-start gap-2 p-3 border rounded cursor-pointer transition-colors {toolsEnabled.includes(tool.id) ? 'bg-blue-50 border-blue-300 dark:bg-blue-900/20 dark:border-blue-700' : 'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'}"
                >
                  <input
                    type="checkbox"
                    id="agent-tool-{tool.id}"
                    name="tool-{tool.id}"
                    checked={toolsEnabled.includes(tool.id)}
                    onchange={() => toggleTool(tool.id)}
                    class="mt-0.5"
                  />
                  <div class="flex-1 min-w-0">
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{tool.name}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{tool.description}</div>
                  </div>
                </label>
              {/each}
            </div>

            <p class="text-xs text-gray-500 dark:text-gray-400">{toolsEnabled.length} tool(s) enabled</p>
          </div>

          <!-- Capabilities -->
          <div>
            <div class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">
              Capabilities
              <span class="text-xs text-gray-500 dark:text-gray-400 font-normal ml-1">(What this agent can do)</span>
            </div>

            {#if capabilities.length > 0}
              <div class="flex flex-wrap gap-2 mb-2">
                {#each capabilities as capability, index}
                  <span class="inline-flex items-center gap-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 px-2 py-1 rounded text-sm">
                    {capability}
                    <button
                      type="button"
                      onclick={() => removeCapability(index)}
                      class="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-200"
                    >
                      ✕
                    </button>
                  </span>
                {/each}
              </div>
            {/if}

            <div class="flex gap-2">
              <input
                type="text"
                id="agent-new-capability"
                name="newCapability"
                autocomplete="off"
                bind:value={newCapability}
                onkeydown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    addCapability();
                  }
                }}
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
                placeholder="e.g., 'Code review', 'Data analysis'"
              />
              <button
                type="button"
                onclick={addCapability}
                class="btn-pill btn-pill-soft btn-pill-sm"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- SECTION 6: CONTEXT SOURCES -->
      <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Context Sources</h3>

        <div>
          <div class="block text-sm font-medium mb-2 text-gray-700 dark:text-gray-300">
            Where should this agent pull knowledge from?
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mb-3">
            {#each commonSources as source}
              <button
                type="button"
                onclick={() => toggleContextSource(source)}
                class="px-3 py-2 border rounded text-sm text-left transition-colors {contextSources.includes(source) ? 'bg-green-50 border-green-300 dark:bg-green-900/20 dark:border-green-700' : 'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'}"
              >
                {#if contextSources.includes(source)}
                  <span class="text-green-600 dark:text-green-400 mr-1">✓</span>
                {/if}
                <span class="text-gray-900 dark:text-white">
                  {source.split('-').join(' ').replace(/\b\w/g, l => l.toUpperCase())}
                </span>
              </button>
            {/each}
          </div>

          <div class="flex gap-2">
            <input
              type="text"
              id="agent-new-context-source"
              name="newContextSource"
              autocomplete="off"
              bind:value={newContextSource}
              onkeydown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addCustomContextSource();
                }
              }}
              class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
              placeholder="Add custom source"
            />
            <button
              type="button"
              onclick={addCustomContextSource}
              class="btn-pill btn-pill-soft btn-pill-sm"
            >
              Add
            </button>
          </div>

          {#if contextSources.length > 0}
            <div class="mt-2 p-2 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600">
              <p class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Active Sources:</p>
              <div class="flex flex-wrap gap-1">
                {#each contextSources as source}
                  <span class="text-xs bg-white dark:bg-gray-600 px-2 py-0.5 rounded border border-gray-300 dark:border-gray-500 text-gray-900 dark:text-white">
                    {source}
                  </span>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- SECTION 7: ADVANCED FEATURES -->
      <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Advanced Features</h3>

        <div class="space-y-3">
          <label for="agent-thinking-enabled" class="flex items-start gap-3 p-3 border border-gray-300 dark:border-gray-600 rounded cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700">
            <input
              type="checkbox"
              id="agent-thinking-enabled"
              name="thinkingEnabled"
              bind:checked={thinkingEnabled}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Enable Chain-of-Thought</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Show the agent's reasoning process before final answer
              </div>
            </div>
          </label>

          <label for="agent-streaming-enabled" class="flex items-start gap-3 p-3 border border-gray-300 dark:border-gray-600 rounded cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700">
            <input
              type="checkbox"
              id="agent-streaming-enabled"
              name="streamingEnabled"
              bind:checked={streamingEnabled}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Enable Streaming Responses</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Stream responses word-by-word for better UX
              </div>
            </div>
          </label>

          <label for="agent-apply-personalization" class="flex items-start gap-3 p-3 border border-gray-300 dark:border-gray-600 rounded cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700">
            <input
              type="checkbox"
              id="agent-apply-personalization"
              name="applyPersonalization"
              bind:checked={applyPersonalization}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Apply Personalizations</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Use learned user preferences and patterns from the learning system
              </div>
            </div>
          </label>
        </div>
      </div>

      <!-- SECTION 8: ACCESS CONTROL -->
      <div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Access & Visibility</h3>

        <div class="space-y-3">
          <label
            for="agent-is-active"
            class="flex items-start gap-3 p-3 border rounded cursor-pointer transition-colors {isActive ? 'bg-green-50 border-green-300 dark:bg-green-900/20 dark:border-green-700' : 'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'}"
          >
            <input
              type="checkbox"
              id="agent-is-active"
              name="isActive"
              bind:checked={isActive}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Active</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Enable this agent for use
              </div>
            </div>
          </label>

          <label for="agent-is-public" class="flex items-start gap-3 p-3 border border-gray-300 dark:border-gray-600 rounded cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700">
            <input
              type="checkbox"
              id="agent-is-public"
              name="isPublic"
              bind:checked={isPublic}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Public</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Make available to other users in workspace
              </div>
            </div>
          </label>

          <label
            for="agent-is-featured"
            class="flex items-start gap-3 p-3 border border-gray-300 dark:border-gray-600 rounded {!isPublic ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700'}"
          >
            <input
              type="checkbox"
              id="agent-is-featured"
              name="isFeatured"
              bind:checked={isFeatured}
              disabled={!isPublic}
              class="mt-1"
            />
            <div class="flex-1">
              <div class="font-medium text-sm text-gray-900 dark:text-white">Featured</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Show in featured gallery {#if !isPublic}(Requires public){/if}
              </div>
            </div>
          </label>
        </div>
      </div>
    {/if}
  </form>

  <!-- Action Buttons - Always Visible at Bottom -->
  <div class="sticky bottom-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 p-4 -mx-6 -mb-6 rounded-b-lg">
    <div class="flex gap-3 justify-end">
      <button
        type="button"
        onclick={onCancel}
        class="btn-pill btn-pill-ghost"
      >
        Cancel
      </button>
      <button
        type="button"
        onclick={handleSubmit}
        disabled={isSubmitting}
        class="btn-pill btn-pill-primary"
      >
        {#if isSubmitting}
          Saving...
        {:else}
          {agent ? 'Update' : 'Create'} Agent
        {/if}
      </button>
    </div>
  </div>
</div>
