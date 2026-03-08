<script lang="ts">
  import { onMount } from 'svelte';
  import { apiClient } from '$lib/api';
  import AIModelSettings from '$lib/components/settings/AIModelSettings.svelte';
  import AIProviderConfig from '$lib/components/settings/AIProviderConfig.svelte';
  import AIAgentConfig from '$lib/components/settings/AIAgentConfig.svelte';
  import AIPromptTemplates from '$lib/components/settings/AIPromptTemplates.svelte';
  import AIUsageStats from '$lib/components/settings/AIUsageStats.svelte';
  import type {
    LLMModel,
    LLMProvider,
    SystemInfo,
    OutputStyle,
    OutputPreference,
    AgentInfo,
    CustomAgent,
    CommandInfo,
    UsageStats,
  } from '$lib/stores/aiSettings';
  import {
    getDateRange,
    generateRecentData,
    generateDailyTrend,
    calculateLocalStorageUsage,
    calculateLocalPowerCost,
  } from '$lib/stores/aiSettings';

  // --- Shared state ---
  let providers = $state<LLMProvider[]>([]);
  let models = $state<LLMModel[]>([]);
  let activeProvider = $state('ollama_local');
  let defaultModel = $state('');
  let isLoading = $state(true);
  let isSaving = $state(false);
  let error = $state('');
  let saveStatus = $state('');
  let systemInfo = $state<SystemInfo | null>(null);

  // Output style preferences
  let outputStyles = $state<OutputStyle[]>([]);
  let outputPreference = $state<OutputPreference | null>(null);
  let selectedDefaultStyleId = $state<string>('');
  let loadingOutputStyles = $state(false);
  let savingOutputPreference = $state(false);

  // Model settings
  let modelSettings = $state({
    temperature: 0.7,
    maxTokens: 8192,
    contextWindow: 32768,
    topP: 0.95,
    streamResponses: true,
    showUsageInChat: true,
  });

  // Agents
  let agents = $state<AgentInfo[]>([]);
  let loadingAgents = $state(false);
  let customAgents = $state<CustomAgent[]>([]);
  let loadingCustomAgents = $state(false);

  // Commands
  let commands = $state<CommandInfo[]>([]);
  let loadingCommands = $state(false);

  // Usage stats
  let usageStats = $state<UsageStats | null>(null);
  let loadingUsage = $state(false);
  let usagePeriod = $state<'today' | 'week' | 'month' | 'all'>('month');

  // Tab
  let activeTab = $state<'models' | 'providers' | 'settings' | 'agents' | 'commands' | 'stats'>('models');

  // --- Derived ---
  let localModelsCount = $derived(models.filter(m => {
    const isLocal = m.provider === 'ollama' || m.provider === 'ollama_local';
    const nameOrId = (m.id || '') + (m.name || '');
    const isCloudRef = nameOrId.toLowerCase().includes('cloud') &&
      (m.size === '< 1 KB' || m.size === '0 B' || !m.size);
    return isLocal && !isCloudRef;
  }).length);

  // --- Load functions ---
  async function loadConfig() {
    isLoading = true;
    error = '';
    try {
      const providersRes = await apiClient.get('/ai/providers');
      if (providersRes.ok) {
        const data = await providersRes.json();
        providers = data.providers || [];
        activeProvider = data.active_provider || 'ollama_local';
        defaultModel = data.default_model || '';
      }

      const modelsRes = await apiClient.get('/ai/models');
      if (modelsRes.ok) {
        const data = await modelsRes.json();
        models = data.models || [];
      }

      const settingsRes = await apiClient.get('/settings');
      if (settingsRes.ok) {
        const data = await settingsRes.json();
        if (data.model_settings) {
          modelSettings = {
            temperature: data.model_settings.temperature ?? 0.7,
            maxTokens: data.model_settings.maxTokens ?? 8192,
            contextWindow: data.model_settings.contextWindow ?? 32768,
            topP: data.model_settings.topP ?? 0.95,
            streamResponses: data.model_settings.streamResponses ?? true,
            showUsageInChat: data.model_settings.showUsageInChat ?? true,
          };
        }
        if (!defaultModel && data.default_model) {
          defaultModel = data.default_model;
        }
      }
    } catch (err) {
      error = 'Failed to load AI configuration. Make sure the backend is running.';
    } finally {
      isLoading = false;
    }
  }

  async function loadSystemInfo() {
    try {
      const res = await apiClient.get('/ai/system');
      if (res.ok) {
        systemInfo = await res.json();
      }
    } catch {
      // non-critical — system info just enhances the UI
    }
  }

  async function loadAgents() {
    loadingAgents = true;
    try {
      const res = await apiClient.get('/ai/agents');
      if (res.ok) {
        const data = await res.json();
        agents = data.agents || [];
      }
    } catch {
      // silently fail
    } finally {
      loadingAgents = false;
    }
  }

  async function loadCustomAgents() {
    loadingCustomAgents = true;
    try {
      const res = await apiClient.get('/ai/custom-agents');
      if (res.ok) {
        const data = await res.json();
        customAgents = data.agents || [];
      }
    } catch {
      // silently fail
    } finally {
      loadingCustomAgents = false;
    }
  }

  async function loadOutputStyles() {
    loadingOutputStyles = true;
    try {
      const res = await apiClient.get('/ai/output-styles');
      if (res.ok) {
        const data = await res.json();
        outputStyles = data.styles || [];
      }
    } catch {
      // silently fail
    } finally {
      loadingOutputStyles = false;
    }
  }

  async function loadOutputPreference() {
    try {
      const res = await apiClient.get('/ai/output-preferences');
      if (res.ok) {
        const data = await res.json();
        outputPreference = data.preference || null;
        selectedDefaultStyleId = outputPreference?.default_style_id || '';
      }
    } catch {
      // silently fail
    }
  }

  async function loadCommands() {
    loadingCommands = true;
    try {
      const res = await apiClient.get('/ai/commands');
      if (res.ok) {
        const data = await res.json();
        commands = data.commands || [];
      }
    } catch {
      // silently fail
    } finally {
      loadingCommands = false;
    }
  }

  async function loadUsageStats() {
    loadingUsage = true;
    try {
      const res = await apiClient.get(`/usage/summary?period=${usagePeriod}`);
      if (res.ok) {
        const data = await res.json();
        usageStats = {
          total_requests: data.total_requests || 0,
          total_tokens: data.total_tokens || 0,
          total_cost: data.total_cost || 0,
          input_tokens: data.input_tokens || 0,
          output_tokens: data.output_tokens || 0,
          by_provider: data.by_provider || {},
          by_model: data.by_model || {},
          by_agent: data.by_agent || {},
          recent: data.recent || [],
          daily_trend: data.daily_trend || [],
          session_count: data.session_count || 0,
          avg_session_duration_min: data.avg_session_duration_min || 0,
          avg_requests_per_session: data.avg_requests_per_session || 0,
          local_model_storage_gb: data.local_model_storage_gb || calculateLocalStorageUsage(models),
          avg_response_time_ms: data.avg_response_time_ms || 0,
          local_power_cost_estimate: data.local_power_cost_estimate || 0,
          cloud_api_cost: data.cloud_api_cost || data.total_cost || 0,
          period_start: data.period_start || getDateRange(usagePeriod).start,
          period_end: data.period_end || getDateRange(usagePeriod).end,
        };
      }
    } catch {
      // API unavailable — show empty usage stats
      usageStats = {
        total_requests: 0,
        total_tokens: 0,
        total_cost: 0,
        input_tokens: 0,
        output_tokens: 0,
        by_provider: {},
        by_model: {},
        by_agent: {},
        recent: [],
        daily_trend: [],
        session_count: 0,
        avg_session_duration_min: 0,
        avg_requests_per_session: 0,
        local_model_storage_gb: calculateLocalStorageUsage(models),
        avg_response_time_ms: 0,
        local_power_cost_estimate: 0,
        cloud_api_cost: 0,
        period_start: getDateRange(usagePeriod).start,
        period_end: getDateRange(usagePeriod).end,
      };
    } finally {
      loadingUsage = false;
    }
  }

  // --- Action callbacks ---
  async function saveSettings() {
    isSaving = true;
    saveStatus = '';
    try {
      await apiClient.put('/settings', {
        ai_provider: activeProvider,
        default_model: defaultModel,
        model_settings: modelSettings,
      });
      saveStatus = 'Settings saved!';
      setTimeout(() => (saveStatus = ''), 2000);
    } catch {
      saveStatus = 'Failed to save settings';
    } finally {
      isSaving = false;
    }
  }

  async function saveOutputPreference() {
    savingOutputPreference = true;
    error = '';
    saveStatus = '';
    try {
      const body = {
        default_style_id: selectedDefaultStyleId || null,
        style_overrides: outputPreference?.style_overrides || {},
        custom_instructions: outputPreference?.custom_instructions || null,
      };
      const res = await apiClient.put('/ai/output-preferences', body);
      if (res.ok) {
        const data = await res.json();
        outputPreference = data.preference || null;
        selectedDefaultStyleId = outputPreference?.default_style_id || '';
        saveStatus = 'Output style saved!';
        setTimeout(() => (saveStatus = ''), 2000);
      } else {
        const data = await res.json().catch(() => ({}));
        error = (data as { error?: string }).error || 'Failed to save output style';
      }
    } catch {
      error = 'Failed to save output style';
    } finally {
      savingOutputPreference = false;
    }
  }

  async function selectProvider(providerId: string) {
    const provider = providers.find(p => p.id === providerId);
    if (!provider?.configured && provider?.type === 'cloud') {
      error = `${provider.name} requires an API key. Configure it in the Providers tab.`;
      setTimeout(() => (error = ''), 4000);
      return;
    }
    activeProvider = providerId;
    try {
      const res = await apiClient.put('/ai/provider', { provider: providerId });
      if (res.ok) {
        saveStatus = 'Provider updated!';
        setTimeout(() => (saveStatus = ''), 3000);
      }
    } catch {
      // silently fail — optimistic update already applied
    }
  }

  function handlePeriodChange(period: 'today' | 'week' | 'month' | 'all') {
    usagePeriod = period;
    loadUsageStats();
  }

  onMount(() => {
    loadConfig();
    loadSystemInfo();
    loadAgents();
    loadCustomAgents();
    loadOutputStyles();
    loadOutputPreference();
  });
</script>

<div class="page">
  <!-- Status Toast -->
  {#if saveStatus}
    <div class="save-toast">{saveStatus}</div>
  {/if}

  {#if isLoading}
    <div class="loading">
      <div class="spinner"></div>
      <span>Loading configuration...</span>
    </div>
  {:else}
    {#if error}
      <div class="error-alert">
        <span>{error}</span>
        <button onclick={() => (error = '')} aria-label="Dismiss error">x</button>
      </div>
    {/if}

    <!-- Tab Navigation -->
    <div class="tabs">
      <button
        class="tab"
        class:active={activeTab === 'models'}
        onclick={() => (activeTab = 'models')}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
        </svg>
        Models
      </button>
      <button
        class="tab"
        class:active={activeTab === 'providers'}
        onclick={() => (activeTab = 'providers')}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
        </svg>
        Providers
      </button>
      <button
        class="tab"
        class:active={activeTab === 'settings'}
        onclick={() => (activeTab = 'settings')}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
        Settings
      </button>
      <button
        class="tab"
        class:active={activeTab === 'agents'}
        onclick={() => (activeTab = 'agents')}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 0 1 7 7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-1H2a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a7 7 0 0 1 7-7h1V5.73c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2M7.5 13A1.5 1.5 0 0 0 6 14.5 1.5 1.5 0 0 0 7.5 16 1.5 1.5 0 0 0 9 14.5 1.5 1.5 0 0 0 7.5 13m9 0a1.5 1.5 0 0 0-1.5 1.5 1.5 1.5 0 0 0 1.5 1.5 1.5 1.5 0 0 0 1.5-1.5 1.5 1.5 0 0 0-1.5-1.5"/>
        </svg>
        Agents
      </button>
      <button
        class="tab"
        class:active={activeTab === 'commands'}
        onclick={() => { activeTab = 'commands'; loadCommands(); }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M7 15l5 5 5-5M7 9l5-5 5 5"/>
        </svg>
        Commands
      </button>
      <button
        class="tab"
        class:active={activeTab === 'stats'}
        onclick={() => { activeTab = 'stats'; loadUsageStats(); }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 20V10M12 20V4M6 20v-6"/>
        </svg>
        Stats
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      {#if activeTab === 'models' || activeTab === 'settings'}
        <AIModelSettings
          {models}
          {providers}
          {activeProvider}
          {defaultModel}
          {systemInfo}
          {outputStyles}
          {outputPreference}
          {selectedDefaultStyleId}
          {loadingOutputStyles}
          {savingOutputPreference}
          {modelSettings}
          {isSaving}
          activeTab={activeTab}
          onSaveSettings={saveSettings}
          onSaveOutputPreference={saveOutputPreference}
          onSelectDefaultStyleId={(id) => (selectedDefaultStyleId = id)}
          onUpdateModelSettings={(s) => (modelSettings = s)}
          onSetDefaultModel={(id) => (defaultModel = id)}
        />
      {:else if activeTab === 'providers'}
        <AIProviderConfig
          {providers}
          {activeProvider}
          onSelectProvider={selectProvider}
          onProviderSaved={loadConfig}
        />
      {:else if activeTab === 'agents'}
        <AIAgentConfig
          {agents}
          {loadingAgents}
          {customAgents}
          {loadingCustomAgents}
          onAgentUpdated={loadCustomAgents}
        />
      {:else if activeTab === 'commands'}
        <AIPromptTemplates
          {commands}
          {loadingCommands}
          onCommandsChanged={loadCommands}
        />
      {:else if activeTab === 'stats'}
        <AIUsageStats
          {usageStats}
          {loadingUsage}
          {systemInfo}
          {providers}
          {localModelsCount}
          onPeriodChange={handlePeriodChange}
          onRefresh={loadUsageStats}
        />
      {/if}
    </div>
  {/if}
</div>

<style>
  .page {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--color-bg);
    color: var(--color-text);
    overflow: hidden;
  }

  /* Loading */
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 16px;
    color: var(--color-text-muted);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin { to { transform: rotate(360deg); } }

  /* Error banner */
  .error-alert {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    background: rgba(220, 38, 38, 0.1);
    border-bottom: 1px solid rgba(220, 38, 38, 0.2);
    color: var(--color-error);
    flex-shrink: 0;
  }

  .error-alert button {
    background: none;
    border: none;
    color: inherit;
    font-size: 20px;
    cursor: pointer;
    opacity: 0.6;
    line-height: 1;
  }

  .error-alert button:hover { opacity: 1; }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 4px;
    padding: 12px 24px;
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
    flex-shrink: 0;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 18px;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: var(--color-text-secondary);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tab:hover {
    background: var(--color-bg);
    color: var(--color-text);
  }

  .tab.active {
    background: var(--color-bg);
    color: var(--color-text);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .tab svg {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }

  /* Tab Content */
  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  /* Save toast */
  .save-toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    padding: 12px 20px;
    background: var(--color-success);
    color: white;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 500;
    z-index: 100;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    animation: slideUp 0.2s ease-out;
  }

  @keyframes slideUp {
    from { transform: translateY(8px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }

  /* Dark mode overrides */
  :global(.dark) .tab.active {
    background: #3a3a3c;
  }
</style>
