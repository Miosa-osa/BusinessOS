<script lang="ts">
  import { apiClient } from '$lib/api';
  import OutputStyleSelector from './OutputStyleSelector.svelte';
  import UserFactsPanel from './UserFactsPanel.svelte';
  import type {
    LLMModel,
    LLMProvider,
    PullProgress,
    SystemInfo,
    OutputStyle,
    OutputPreference,
    AvailableModel,
    ModelCapability,
  } from '$lib/stores/aiSettings';
  import {
    availableModels,
    cloudModels,
    capabilityInfo,
    formatBytes,
  } from '$lib/stores/aiSettings';

  interface Props {
    models: LLMModel[];
    providers: LLMProvider[];
    activeProvider: string;
    defaultModel: string;
    systemInfo: SystemInfo | null;
    outputStyles: OutputStyle[];
    outputPreference: OutputPreference | null;
    selectedDefaultStyleId: string;
    loadingOutputStyles: boolean;
    savingOutputPreference: boolean;
    modelSettings: {
      temperature: number;
      maxTokens: number;
      contextWindow: number;
      topP: number;
      streamResponses: boolean;
      showUsageInChat: boolean;
    };
    isSaving: boolean;
    onSaveSettings: () => void;
    onSaveOutputPreference: () => void;
    onSelectDefaultStyleId: (id: string) => void;
    onUpdateModelSettings: (settings: {
      temperature: number;
      maxTokens: number;
      contextWindow: number;
      topP: number;
      streamResponses: boolean;
      showUsageInChat: boolean;
    }) => void;
    onSetDefaultModel: (modelId: string) => void;
    /** Which parent tab is active — controls which section renders */
    activeTab?: 'models' | 'settings';
  }

  let {
    models,
    providers,
    activeProvider,
    defaultModel,
    systemInfo,
    outputStyles,
    outputPreference,
    selectedDefaultStyleId,
    loadingOutputStyles,
    savingOutputPreference,
    modelSettings,
    isSaving,
    onSaveSettings,
    onSaveOutputPreference,
    onSelectDefaultStyleId,
    onUpdateModelSettings,
    onSetDefaultModel,
    activeTab = 'models',
  }: Props = $props();

  // Model browser state
  let modelSearchQuery = $state('');
  let selectedCapabilityFilters = $state<ModelCapability[]>([]);
  let selectedProviderFilter = $state<'all' | 'local' | 'cloud'>('all');
  let modelSortBy = $state<'recommended' | 'name' | 'size' | 'downloads'>('recommended');
  let showOnlyInstalled = $state(false);
  let showSourceDropdown = $state(false);
  let showFiltersDropdown = $state(false);
  let selectedVariants = $state<Record<string, string>>({});

  // Pull model state
  let pullModelName = $state('');
  let isPulling = $state(false);
  let pullProgress = $state<PullProgress | null>(null);
  let pullError = $state('');
  let pullStartTime = $state<number>(0);
  let pullSpeed = $state<string>('');

  function getLocalModels(): LLMModel[] {
    return (models || []).filter((m) => {
      const isLocalProvider = m.provider === 'ollama' || m.provider === 'ollama_local';
      const nameOrId = (m.id || '') + (m.name || '');
      const isCloudRef =
        nameOrId.toLowerCase().includes('cloud') &&
        (m.size === '< 1 KB' || m.size === '0 B' || !m.size);
      return isLocalProvider && !isCloudRef;
    });
  }

  function getSelectedVariant(model: AvailableModel): string {
    if (!model.variants || model.variants.length === 0) return model.id;
    return selectedVariants[model.id] || model.variants[0].id;
  }

  function getVariantInfo(model: AvailableModel) {
    if (!model.variants || model.variants.length === 0) return null;
    const selectedId = getSelectedVariant(model);
    return model.variants.find((v) => v.id === selectedId) || model.variants[0];
  }

  function getFilteredModels(): AvailableModel[] {
    let filtered = [...availableModels];
    const installedIds = new Set(models.map((m) => m.id.toLowerCase()));
    filtered = filtered.map((m) => ({
      ...m,
      isInstalled:
        installedIds.has(m.id.toLowerCase()) ||
        installedIds.has(m.id.split(':')[0].toLowerCase()),
    }));

    if (modelSearchQuery) {
      const query = modelSearchQuery.toLowerCase();
      filtered = filtered.filter(
        (m) =>
          m.name.toLowerCase().includes(query) ||
          m.description.toLowerCase().includes(query) ||
          m.id.toLowerCase().includes(query),
      );
    }

    if (selectedCapabilityFilters.length > 0) {
      filtered = filtered.filter((m) =>
        selectedCapabilityFilters.some((cap) => m.capabilities.includes(cap)),
      );
    }

    if (selectedProviderFilter !== 'all') {
      filtered = filtered.filter((m) => m.provider === selectedProviderFilter);
    }

    if (showOnlyInstalled) {
      filtered = filtered.filter((m) => m.isInstalled);
    }

    switch (modelSortBy) {
      case 'name':
        filtered.sort((a, b) => a.name.localeCompare(b.name));
        break;
      case 'size':
        filtered.sort((a, b) => {
          const sizeA = parseFloat(a.size) || 999;
          const sizeB = parseFloat(b.size) || 999;
          return sizeA - sizeB;
        });
        break;
      case 'downloads':
        filtered.sort((a, b) => {
          const parseDownloads = (d: string | undefined) => {
            if (!d) return 0;
            const num = parseFloat(d);
            if (d.includes('M')) return num * 1000000;
            if (d.includes('K')) return num * 1000;
            return num;
          };
          return parseDownloads(b.downloads) - parseDownloads(a.downloads);
        });
        break;
      default:
        filtered.sort((a, b) => {
          if (a.isInstalled && !b.isInstalled) return -1;
          if (!a.isInstalled && b.isInstalled) return 1;
          const parseDownloads = (d: string | undefined) => {
            if (!d) return 0;
            const num = parseFloat(d);
            if (d.includes('M')) return num * 1000000;
            if (d.includes('K')) return num * 1000;
            return num;
          };
          return parseDownloads(b.downloads) - parseDownloads(a.downloads);
        });
    }

    return filtered;
  }

  function getProgressPercent(): number {
    if (!pullProgress?.total || !pullProgress?.completed) return 0;
    return Math.round((pullProgress.completed / pullProgress.total) * 100);
  }

  function getTimeRemaining(): string {
    if (!pullProgress?.total || !pullProgress?.completed || !pullSpeed) return '';
    const remaining = pullProgress.total - pullProgress.completed;
    const speedMatch = pullSpeed.match(/[\d.]+/);
    if (!speedMatch) return '';
    let speedBytes = parseFloat(speedMatch[0]);
    if (pullSpeed.includes('KB')) speedBytes *= 1024;
    if (pullSpeed.includes('MB')) speedBytes *= 1024 * 1024;
    if (pullSpeed.includes('GB')) speedBytes *= 1024 * 1024 * 1024;
    if (speedBytes <= 0) return '';
    const seconds = remaining / speedBytes;
    if (seconds < 60) return `~${Math.ceil(seconds)}s`;
    if (seconds < 3600) return `~${Math.ceil(seconds / 60)}m`;
    return `~${(seconds / 3600).toFixed(1)}h`;
  }

  async function pullModel(modelName: string) {
    if (!modelName.trim() || isPulling) return;
    isPulling = true;
    pullError = '';
    pullProgress = { status: 'Connecting...' };
    pullStartTime = Date.now();
    const model = modelName.trim();

    try {
      const response = await fetch(`/api/ai/models/pull`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ model }),
        credentials: 'include',
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to pull model');
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error('No response body');

      const decoder = new TextDecoder();
      let buffer = '';
      let lastCompleted = 0;
      let lastTime = Date.now();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6));
              pullProgress = data;
              if (data.completed && data.completed > lastCompleted) {
                const now = Date.now();
                const timeDiff = (now - lastTime) / 1000;
                const bytesDiff = data.completed - lastCompleted;
                if (timeDiff > 0) {
                  const speed = bytesDiff / timeDiff;
                  pullSpeed = formatBytes(speed) + '/s';
                }
                lastCompleted = data.completed;
                lastTime = now;
              }
              if (data.status === 'complete' || data.status === 'success') {
                pullProgress = { status: 'Complete!' };
                pullModelName = '';
              }
            } catch {}
          }
        }
      }
    } catch (err) {
      pullError = err instanceof Error ? err.message : 'Failed to pull model';
      pullProgress = null;
    } finally {
      isPulling = false;
      pullSpeed = '';
    }
  }

  async function deleteModel(modelId: string) {
    if (!confirm(`Delete model ${modelId}?`)) return;
    try {
      const res = await apiClient.delete(`/ai/models/${encodeURIComponent(modelId)}`);
      if (!res.ok) {
        console.error('Failed to delete model');
      }
    } catch (err) {
      console.error('Failed to delete model:', err);
    }
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest('.filter-dropdown-wrapper')) {
      showSourceDropdown = false;
      showFiltersDropdown = false;
    }
  }
</script>

<svelte:document onclick={handleClickOutside} />

{#if activeTab === 'models'}
<!-- Models Tab -->
<section class="section model-browser-section">
  <!-- Compact Filter Bar -->
  <div class="browser-controls">
    <div class="compact-search">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
      <input type="text" bind:value={modelSearchQuery} placeholder="Search..." />
      {#if modelSearchQuery}
        <button class="clear-search" onclick={() => modelSearchQuery = ''} aria-label="Clear search">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        </button>
      {:else}
        <span class="search-shortcut">⌘K</span>
      {/if}
    </div>

    <div class="filter-dropdown-wrapper">
      <button
        class="filter-dropdown-btn"
        class:active={selectedProviderFilter !== 'all'}
        onclick={() => { showSourceDropdown = !showSourceDropdown; showFiltersDropdown = false; }}
      >
        {#if selectedProviderFilter === 'local'}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><circle cx="6" cy="6" r="1" fill="currentColor"/><circle cx="6" cy="18" r="1" fill="currentColor"/></svg>
          Local
        {:else if selectedProviderFilter === 'cloud'}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg>
          Cloud
        {:else}
          Source
        {/if}
        <span class="dropdown-count">{selectedProviderFilter === 'all' ? availableModels.length : selectedProviderFilter === 'local' ? availableModels.filter(m => m.provider === 'local').length : availableModels.filter(m => m.provider === 'cloud').length}</span>
        <svg class="dropdown-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>
      </button>
      {#if showSourceDropdown}
        <div class="filter-dropdown-menu" role="menu">
          <button class="dropdown-item" class:selected={selectedProviderFilter === 'all'} onclick={() => { selectedProviderFilter = 'all'; showSourceDropdown = false; }}>
            All <span class="item-count">{availableModels.length}</span>
          </button>
          <button class="dropdown-item" class:selected={selectedProviderFilter === 'local'} onclick={() => { selectedProviderFilter = 'local'; showSourceDropdown = false; }}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><circle cx="6" cy="6" r="1" fill="currentColor"/><circle cx="6" cy="18" r="1" fill="currentColor"/></svg>
            Local <span class="item-count">{availableModels.filter(m => m.provider === 'local').length}</span>
          </button>
          <button class="dropdown-item" class:selected={selectedProviderFilter === 'cloud'} onclick={() => { selectedProviderFilter = 'cloud'; showSourceDropdown = false; }}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg>
            Cloud <span class="item-count">{availableModels.filter(m => m.provider === 'cloud').length}</span>
          </button>
        </div>
      {/if}
    </div>

    <div class="filter-dropdown-wrapper">
      <button
        class="filter-dropdown-btn"
        class:active={selectedCapabilityFilters.length > 0}
        onclick={() => { showFiltersDropdown = !showFiltersDropdown; showSourceDropdown = false; }}
      >
        Filters
        {#if selectedCapabilityFilters.length > 0}
          <span class="dropdown-count">{selectedCapabilityFilters.length}</span>
        {/if}
        <svg class="dropdown-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>
      </button>
      {#if showFiltersDropdown}
        <div class="filter-dropdown-menu capabilities-menu" role="menu">
          {#each Object.entries(capabilityInfo) as [cap, info]}
            <label class="dropdown-checkbox-item">
              <input
                type="checkbox"
                checked={selectedCapabilityFilters.includes(cap as ModelCapability)}
                onchange={() => {
                  if (selectedCapabilityFilters.includes(cap as ModelCapability)) {
                    selectedCapabilityFilters = selectedCapabilityFilters.filter(c => c !== cap);
                  } else {
                    selectedCapabilityFilters = [...selectedCapabilityFilters, cap as ModelCapability];
                  }
                }}
              />
              <svg class="cap-icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={info.iconPath}/></svg>
              {info.label}
            </label>
          {/each}
          {#if selectedCapabilityFilters.length > 0}
            <button class="dropdown-clear" onclick={() => selectedCapabilityFilters = []}>Clear all</button>
          {/if}
        </div>
      {/if}
    </div>

    {#if selectedCapabilityFilters.length > 0}
      <div class="active-filter-chips">
        {#each selectedCapabilityFilters as cap}
          <button class="filter-chip" onclick={() => selectedCapabilityFilters = selectedCapabilityFilters.filter(c => c !== cap)}>
            {capabilityInfo[cap].label}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        {/each}
      </div>
    {/if}

    <label class="compact-toggle">
      <input type="checkbox" bind:checked={showOnlyInstalled} />
      <span class="toggle-slider"></span>
      <span class="toggle-label">Installed</span>
    </label>

    <select bind:value={modelSortBy} class="compact-sort-select">
      <option value="recommended">Recommended</option>
      <option value="name">Name</option>
      <option value="size">Size</option>
      <option value="downloads">Downloads</option>
    </select>
  </div>

  {#if systemInfo && systemInfo.recommended_models.length > 0 && activeProvider === 'ollama_local' && !modelSearchQuery && selectedCapabilityFilters.length === 0}
    <div class="recommended-banner">
      <div class="rec-banner-header">
        <h3>Recommended for You</h3>
        <span class="rec-badge-info">Based on {systemInfo.total_ram_gb}GB RAM</span>
      </div>
      <div class="rec-chips">
        {#each systemInfo.recommended_models.slice(0, 4) as model}
          {@const isInstalled = getLocalModels().some(m => m.id.startsWith(model.name.split(':')[0]))}
          <button
            class="rec-chip"
            class:installed={isInstalled}
            onclick={() => { if (!isInstalled) pullModel(model.name); }}
            disabled={isPulling}
          >
            <span class="chip-name">{model.name}</span>
            <span class="chip-meta">{model.speed} • {model.quality}</span>
            {#if isInstalled}
              <span class="chip-status installed">Installed</span>
            {:else}
              <span class="chip-status pull">Pull</span>
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <div class="browser-content">
    <div class="browser-header">
      <h3>
        {#if showOnlyInstalled}
          Installed Models
        {:else if selectedCapabilityFilters.length === 1}
          {capabilityInfo[selectedCapabilityFilters[0]].label} Models
        {:else if selectedCapabilityFilters.length > 1}
          Filtered Models
        {:else}
          All Models
        {/if}
      </h3>
      <span class="model-count">{getFilteredModels().length} models</span>
    </div>

    {#if getFilteredModels().length === 0}
      <div class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        </div>
        <h3>No models found</h3>
        <p>Try adjusting your search or filters</p>
      </div>
    {:else}
      <div class="model-browser-grid">
        {#each getFilteredModels() as model}
          {@const isDefault = defaultModel === model.id || defaultModel.startsWith(model.id.split(':')[0])}
          {@const variantInfo = getVariantInfo(model)}
          {@const displaySize = variantInfo ? variantInfo.size : model.size}
          {@const displayParams = variantInfo ? variantInfo.params : model.params}
          {@const pullId = getSelectedVariant(model)}
          <div class="browser-model-card" class:installed={model.isInstalled} class:default={isDefault}>
            <div class="bmc-header">
              <div class="bmc-title">
                <span class="bmc-name">{model.name}</span>
                {#if model.isInstalled}
                  <span class="bmc-installed-badge">Installed</span>
                {/if}
              </div>
              <div class="bmc-meta">
                <span class="bmc-size">{displaySize}</span>
                <span class="bmc-params">{displayParams}</span>
              </div>
            </div>
            <p class="bmc-description">{model.description}</p>
            <div class="bmc-capabilities">
              {#each model.capabilities as cap}
                {@const info = capabilityInfo[cap]}
                <span class="cap-badge {cap}" title={info.label}>
                  <svg class="cap-badge-icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={info.iconPath}/></svg>
                  <span class="cap-badge-label">{info.label}</span>
                </span>
              {/each}
            </div>
            {#if model.variants && model.variants.length > 1}
              <div class="bmc-variants">
                <span class="variants-label">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
                  Select size:
                </span>
                <div class="variant-buttons" class:many={model.variants.length > 4}>
                  {#each model.variants as variant}
                    <button
                      class="variant-btn"
                      class:selected={getSelectedVariant(model) === variant.id}
                      onclick={() => selectedVariants[model.id] = variant.id}
                      title="{variant.params} parameters • {variant.size} download"
                    >
                      <span class="variant-params">{variant.params}</span>
                      <span class="variant-size">{variant.size}</span>
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
            <div class="bmc-footer">
              {#if model.downloads}
                <span class="bmc-downloads">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
                  {model.downloads}
                </span>
              {/if}
              {#if model.isInstalled}
                <div class="bmc-actions">
                  <button
                    class="bmc-btn default-btn"
                    class:is-default={isDefault}
                    onclick={() => { onSetDefaultModel(model.id); onSaveSettings(); }}
                  >
                    {isDefault ? 'Default' : 'Set Default'}
                  </button>
                  <button class="bmc-btn delete-btn" onclick={() => deleteModel(model.id)} title="Delete model">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
                  </button>
                </div>
              {:else}
                <button class="bmc-btn pull-btn" onclick={() => pullModel(pullId)} disabled={isPulling}>
                  {#if isPulling && pullModelName === pullId}
                    <div class="btn-spinner-small"></div>
                    Pulling...
                  {:else}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
                    Pull Model
                  {/if}
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Pull Custom Model Form -->
  <div class="pull-card-compact">
    <div class="pull-compact-header">
      <h4>Pull Custom Model</h4>
      <a href="https://ollama.com/library" target="_blank" class="browse-link">
        Browse Ollama Library
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3"/></svg>
      </a>
    </div>
    <div class="pull-form-compact">
      <input
        type="text"
        bind:value={pullModelName}
        placeholder="Enter model name (e.g., llama3.2:3b, phi3:medium)"
        disabled={isPulling}
      />
      <button class="pull-btn-compact" onclick={() => pullModel(pullModelName)} disabled={isPulling || !pullModelName.trim()}>
        {#if isPulling}
          <div class="btn-spinner-small"></div>
        {:else}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
        {/if}
        Pull
      </button>
    </div>
    {#if isPulling && pullProgress}
      <div class="pull-progress-compact">
        <div class="progress-info">
          <span class="progress-status">{pullProgress.status}</span>
          <span class="progress-percent">{getProgressPercent()}%</span>
        </div>
        <div class="progress-bar-compact">
          <div class="progress-fill" style="width: {getProgressPercent()}%"></div>
        </div>
        <div class="progress-details">
          {#if pullProgress.total && pullProgress.completed}
            <span>{formatBytes(pullProgress.completed)} / {formatBytes(pullProgress.total)}</span>
          {/if}
          {#if pullSpeed}<span class="speed">{pullSpeed}</span>{/if}
          {#if getTimeRemaining()}<span class="time">{getTimeRemaining()}</span>{/if}
        </div>
      </div>
    {/if}
    {#if pullError}
      <div class="pull-error-compact">{pullError}</div>
    {/if}
  </div>
</section>
{/if}

{#if activeTab === 'settings'}
<!-- Settings Tab -->
<section class="section">
  <div class="section-header">
    <h2>Model Settings</h2>
    <span class="subtitle">Fine-tune AI behavior</span>
  </div>

  <div class="setting-card" style="margin-bottom: 16px;">
    <div class="setting-header">
      <label>Default Output Style</label>
    </div>
    <p class="setting-desc" style="margin-bottom: 1rem;">Choose how the AI formats its responses.</p>
    {#if loadingOutputStyles}
      <div class="loading" style="padding: 2rem; text-align: center;">
        <div class="spinner"></div>
        <span>Loading output styles...</span>
      </div>
    {:else}
      <OutputStyleSelector
        styles={outputStyles}
        selectedStyleId={selectedDefaultStyleId}
        onSelect={(styleId) => onSelectDefaultStyleId(styleId)}
        disabled={savingOutputPreference}
      />
      <button class="save-btn" style="margin-top: 1rem;" onclick={onSaveOutputPreference} disabled={savingOutputPreference}>
        {savingOutputPreference ? 'Saving...' : 'Save Default Style'}
      </button>
    {/if}
  </div>

  <div class="setting-card" style="margin-bottom: 16px; padding: 0; overflow: hidden; max-height: 600px;">
    <UserFactsPanel />
  </div>

  <div class="presets-row">
    <span class="presets-label">Quick Presets:</span>
    <button class="preset-btn" onclick={() => onUpdateModelSettings({ ...modelSettings, temperature: 0.3, maxTokens: 4096, topP: 0.8, contextWindow: 16384 })}>
      <svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
      Fast
    </button>
    <button class="preset-btn" onclick={() => onUpdateModelSettings({ ...modelSettings, temperature: 0.7, maxTokens: 8192, topP: 0.95, contextWindow: 32768 })}>
      <svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
      Balanced
    </button>
    <button class="preset-btn" onclick={() => onUpdateModelSettings({ ...modelSettings, temperature: 0.9, maxTokens: 16384, topP: 1.0, contextWindow: 65536 })}>
      <svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
      Quality
    </button>
    <button class="preset-btn" onclick={() => onUpdateModelSettings({ ...modelSettings, temperature: 1.0, maxTokens: 32768, topP: 1.0, contextWindow: 131072 })}>
      <svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 19h20L12 2z"/><path d="M12 9v4"/><circle cx="12" cy="16" r="1"/></svg>
      Maximum
    </button>
  </div>

  <div class="settings-grid">
    <div class="setting-card">
      <div class="setting-header">
        <label for="contextWindow">Context Window</label>
        <span class="setting-value">{(modelSettings.contextWindow / 1000).toFixed(0)}K tokens</span>
      </div>
      <input type="range" id="contextWindow" value={modelSettings.contextWindow} oninput={(e) => onUpdateModelSettings({ ...modelSettings, contextWindow: Number((e.target as HTMLInputElement).value) })} min="4096" max="131072" step="4096" />
      <p class="setting-desc">How much conversation history to include</p>
    </div>
    <div class="setting-card">
      <div class="setting-header">
        <label for="maxTokens">Max Output Tokens</label>
        <span class="setting-value">{(modelSettings.maxTokens / 1000).toFixed(1)}K</span>
      </div>
      <input type="range" id="maxTokens" value={modelSettings.maxTokens} oninput={(e) => onUpdateModelSettings({ ...modelSettings, maxTokens: Number((e.target as HTMLInputElement).value) })} min="512" max="32768" step="512" />
      <p class="setting-desc">Maximum length of AI responses</p>
    </div>
    <div class="setting-card">
      <div class="setting-header">
        <label for="temperature">Temperature</label>
        <span class="setting-value">{modelSettings.temperature}</span>
      </div>
      <input type="range" id="temperature" value={modelSettings.temperature} oninput={(e) => onUpdateModelSettings({ ...modelSettings, temperature: Number((e.target as HTMLInputElement).value) })} min="0" max="2" step="0.1" />
      <p class="setting-desc">Lower = more focused, Higher = more creative</p>
    </div>
    <div class="setting-card">
      <div class="setting-header">
        <label for="topP">Top P (Nucleus Sampling)</label>
        <span class="setting-value">{modelSettings.topP}</span>
      </div>
      <input type="range" id="topP" value={modelSettings.topP} oninput={(e) => onUpdateModelSettings({ ...modelSettings, topP: Number((e.target as HTMLInputElement).value) })} min="0.1" max="1" step="0.05" />
      <p class="setting-desc">Controls diversity of responses</p>
    </div>
    <div class="setting-card toggle-card">
      <div class="setting-header">
        <label for="streaming">Stream Responses</label>
        <button class="toggle" class:on={modelSettings.streamResponses} onclick={() => onUpdateModelSettings({ ...modelSettings, streamResponses: !modelSettings.streamResponses })}>
          <span class="toggle-knob"></span>
        </button>
      </div>
      <p class="setting-desc">Show responses as they're generated</p>
    </div>
    <div class="setting-card toggle-card">
      <div class="setting-header">
        <label for="showUsage">Show Usage Stats</label>
        <button class="toggle" class:on={modelSettings.showUsageInChat} onclick={() => onUpdateModelSettings({ ...modelSettings, showUsageInChat: !modelSettings.showUsageInChat })}>
          <span class="toggle-knob"></span>
        </button>
      </div>
      <p class="setting-desc">Display tokens/second and token count after each response</p>
    </div>
  </div>

  <div class="settings-actions">
    <button class="action-btn primary" onclick={onSaveSettings} disabled={isSaving}>
      {isSaving ? 'Saving...' : 'Save Settings'}
    </button>
    <button class="action-btn" onclick={() => onUpdateModelSettings({ temperature: 0.7, maxTokens: 8192, topP: 0.95, contextWindow: 32768, streamResponses: true, showUsageInChat: true })}>
      Reset to Defaults
    </button>
  </div>
</section>

<!-- Default Model Selection -->
<section class="section">
  <div class="section-header">
    <h2>Default Model</h2>
    <span class="subtitle">Select your preferred model</span>
  </div>
  <div class="default-model-grid">
    {#if activeProvider === 'ollama_local'}
      {#each getLocalModels() as model}
        <button
          class="default-model-btn"
          class:selected={defaultModel === model.id}
          onclick={() => { onSetDefaultModel(model.id); onSaveSettings(); }}
        >
          <span class="dm-name">{model.name}</span>
          {#if model.size}<span class="dm-size">{model.size}</span>{/if}
          {#if defaultModel === model.id}
            <svg class="dm-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
          {/if}
        </button>
      {/each}
    {:else}
      {#each cloudModels[activeProvider] || [] as model}
        <button
          class="default-model-btn"
          class:selected={defaultModel === model.id}
          onclick={() => { onSetDefaultModel(model.id); onSaveSettings(); }}
        >
          <span class="dm-name">{model.name}</span>
          <span class="dm-desc">{model.description}</span>
          {#if defaultModel === model.id}
            <svg class="dm-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
          {/if}
        </button>
      {/each}
    {/if}
  </div>
</section>
{/if}

<style>
  /* Loading */
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
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

  /* Sections */
  .section { margin-bottom: 32px; }
  .section-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .section-header h2 { font-size: 18px; font-weight: 600; margin: 0; }
  .subtitle { font-size: 13px; color: var(--color-text-muted); }

  /* Empty State */
  .empty-state { text-align: center; padding: 48px; background: var(--color-bg); border: 1px dashed var(--color-border); border-radius: 12px; }
  .empty-icon { width: 48px; height: 48px; margin-bottom: 12px; color: var(--color-text-muted); }
  .empty-icon svg { width: 100%; height: 100%; }
  .empty-state h3 { margin: 0 0 8px; font-size: 16px; }
  .empty-state p { margin: 0; color: var(--color-text-muted); font-size: 14px; }

  /* Model Browser */
  .model-browser-section { display: flex; flex-direction: column; gap: 20px; }

  .browser-controls {
    display: flex; flex-direction: row; align-items: center; gap: 10px;
    padding: 10px 14px; background: #ffffff; border: 1px solid var(--color-border);
    border-radius: 10px; position: sticky; top: 0; z-index: 100;
    box-shadow: 0 2px 12px rgba(0,0,0,0.06); min-height: 52px;
  }
  :global(.dark) .browser-controls { background: #1a1a1a; box-shadow: 0 2px 12px rgba(0,0,0,0.3); border-color: rgba(255,255,255,0.1); }

  .compact-search {
    display: flex; align-items: center; gap: 8px; padding: 6px 12px;
    background: var(--color-bg-secondary); border: 1px solid var(--color-border);
    border-radius: 8px; min-width: 160px; max-width: 220px; transition: all 0.2s ease;
  }
  .compact-search:focus-within { border-color: var(--color-primary); box-shadow: 0 0 0 2px rgba(59,130,246,0.12); min-width: 200px; }
  :global(.dark) .compact-search { background: rgba(255,255,255,0.04); }
  .compact-search svg { width: 15px; height: 15px; color: var(--color-text-muted); flex-shrink: 0; }
  .compact-search input { flex: 1; min-width: 0; background: none; border: none; font-size: 13px; color: var(--color-text); outline: none; }
  .compact-search input::placeholder { color: var(--color-text-muted); }

  .clear-search { display: flex; align-items: center; justify-content: center; width: 18px; height: 18px; background: var(--color-bg-tertiary); border: none; border-radius: 4px; cursor: pointer; color: var(--color-text-muted); padding: 0; }
  .clear-search:hover { background: var(--color-bg-secondary); color: var(--color-text); }
  .clear-search svg { width: 12px; height: 12px; }
  .search-shortcut { font-size: 10px; font-weight: 500; color: var(--color-text-muted); background: var(--color-bg-tertiary); padding: 2px 6px; border-radius: 4px; border: 1px solid var(--color-border); opacity: 0.6; }

  .filter-dropdown-wrapper { position: relative; }
  .filter-dropdown-btn { display: flex; align-items: center; gap: 6px; padding: 6px 10px; background: var(--color-bg-secondary); border: 1px solid var(--color-border); border-radius: 8px; font-size: 13px; font-weight: 500; color: var(--color-text-secondary); cursor: pointer; transition: all 0.15s ease; white-space: nowrap; }
  .filter-dropdown-btn:hover { border-color: var(--color-border-hover); color: var(--color-text); }
  .filter-dropdown-btn.active { background: var(--color-bg-tertiary); border-color: var(--color-border-hover); color: var(--color-text); }
  :global(.dark) .filter-dropdown-btn { background: rgba(255,255,255,0.04); }
  :global(.dark) .filter-dropdown-btn.active { background: rgba(255,255,255,0.08); }
  .filter-dropdown-btn svg:not(.dropdown-chevron) { width: 14px; height: 14px; }
  .dropdown-chevron { width: 12px; height: 12px; opacity: 0.5; margin-left: 2px; }
  .dropdown-count { font-size: 11px; font-weight: 600; padding: 1px 5px; background: rgba(0,0,0,0.06); border-radius: 8px; color: var(--color-text-muted); min-width: 18px; text-align: center; }
  :global(.dark) .dropdown-count { background: rgba(255,255,255,0.1); }

  .filter-dropdown-menu { position: absolute; top: calc(100% + 6px); left: 0; min-width: 160px; background: #ffffff; border: 1px solid var(--color-border); border-radius: 10px; box-shadow: 0 8px 24px rgba(0,0,0,0.12); z-index: 200; padding: 6px; animation: dropdownFadeIn 0.15s ease; }
  :global(.dark) .filter-dropdown-menu { background: #252525; border-color: rgba(255,255,255,0.1); box-shadow: 0 8px 24px rgba(0,0,0,0.4); }
  @keyframes dropdownFadeIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

  .dropdown-item { display: flex; align-items: center; gap: 8px; width: 100%; padding: 8px 10px; background: transparent; border: none; border-radius: 6px; font-size: 13px; color: var(--color-text); cursor: pointer; text-align: left; transition: background 0.1s ease; }
  .dropdown-item:hover { background: var(--color-bg-secondary); }
  .dropdown-item.selected { background: var(--color-bg-tertiary); font-weight: 500; }
  .dropdown-item svg { width: 14px; height: 14px; opacity: 0.7; }
  .item-count { margin-left: auto; font-size: 11px; color: var(--color-text-muted); font-weight: 500; }

  .capabilities-menu { min-width: 180px; }
  .dropdown-checkbox-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 6px; font-size: 13px; color: var(--color-text); cursor: pointer; transition: background 0.1s ease; }
  .dropdown-checkbox-item:hover { background: var(--color-bg-secondary); }
  .dropdown-checkbox-item input[type="checkbox"] { width: 14px; height: 14px; accent-color: #34c759; cursor: pointer; }
  .dropdown-clear { display: block; width: 100%; padding: 8px 10px; margin-top: 4px; background: transparent; border: none; border-top: 1px solid var(--color-border); font-size: 12px; color: var(--color-text-muted); cursor: pointer; text-align: center; }
  .dropdown-clear:hover { color: var(--color-text); }
  .cap-icon-svg { width: 14px; height: 14px; flex-shrink: 0; }

  .active-filter-chips { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
  .filter-chip { display: flex; align-items: center; gap: 4px; padding: 4px 8px; background: var(--color-bg-tertiary); border: 1px solid var(--color-border); border-radius: 6px; font-size: 11px; font-weight: 500; color: var(--color-text); cursor: pointer; transition: all 0.15s ease; }
  .filter-chip:hover { background: var(--color-bg-secondary); border-color: var(--color-border-hover); }
  .filter-chip svg { width: 10px; height: 10px; }

  .compact-toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 12px; color: var(--color-text-secondary); white-space: nowrap; }
  .compact-toggle input { position: absolute; opacity: 0; width: 0; height: 0; }
  .toggle-slider { position: relative; width: 36px; height: 20px; background: rgba(0,0,0,0.15); border-radius: 20px; transition: background 0.2s ease; flex-shrink: 0; }
  .toggle-slider::after { content: ''; position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; background: white; border-radius: 50%; box-shadow: 0 1px 3px rgba(0,0,0,0.2); transition: transform 0.2s ease; }
  .compact-toggle input:checked + .toggle-slider { background: #34c759; }
  .compact-toggle input:checked + .toggle-slider::after { transform: translateX(16px); }
  :global(.dark) .toggle-slider { background: rgba(255,255,255,0.15); }
  :global(.dark) .toggle-slider::after { background: #e5e5e5; }
  .toggle-label { font-weight: 500; }

  .compact-sort-select { padding: 6px 10px; padding-right: 28px; background: var(--color-bg-secondary); border: 1px solid var(--color-border); border-radius: 8px; font-size: 12px; font-weight: 500; color: var(--color-text); cursor: pointer; appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%239ca3af' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; }
  .compact-sort-select:hover { border-color: var(--color-border-hover); }
  :global(.dark) .compact-sort-select { background-color: rgba(255,255,255,0.04); }

  .recommended-banner { padding: 16px; background: linear-gradient(135deg, rgba(59,130,246,0.08), rgba(139,92,246,0.08)); border: 1px solid rgba(59,130,246,0.15); border-radius: 12px; }
  :global(.dark) .recommended-banner { background: linear-gradient(135deg, rgba(59,130,246,0.15), rgba(139,92,246,0.15)); border-color: rgba(59,130,246,0.25); }
  .rec-banner-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
  .rec-banner-header h3 { margin: 0; font-size: 15px; font-weight: 600; }
  .rec-badge-info { font-size: 11px; padding: 3px 10px; background: rgba(59,130,246,0.12); border-radius: 10px; color: #3b82f6; }
  :global(.dark) .rec-badge-info { background: rgba(59,130,246,0.2); color: #60a5fa; }
  .rec-chips { display: flex; flex-wrap: wrap; gap: 10px; }
  .rec-chip { display: flex; flex-direction: column; gap: 4px; padding: 12px 16px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 10px; cursor: pointer; transition: all 0.2s; min-width: 160px; }
  .rec-chip:hover { border-color: var(--color-primary); transform: translateY(-2px); }
  .rec-chip.installed { border-color: rgba(34,197,94,0.4); background: rgba(34,197,94,0.05); }
  .rec-chip:disabled { opacity: 0.7; cursor: not-allowed; }
  .chip-name { font-weight: 600; font-size: 13px; }
  .chip-meta { font-size: 11px; color: var(--color-text-muted); }
  .chip-status { font-size: 10px; padding: 2px 8px; border-radius: 4px; width: fit-content; }
  .chip-status.installed { background: rgba(34,197,94,0.1); color: var(--color-success); }
  .chip-status.pull { background: rgba(59,130,246,0.1); color: #3b82f6; }

  .browser-content { display: flex; flex-direction: column; gap: 16px; }
  .browser-header { display: flex; align-items: center; justify-content: space-between; }
  .browser-header h3 { margin: 0; font-size: 16px; font-weight: 600; }
  .model-count { font-size: 13px; color: var(--color-text-muted); }

  .model-browser-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 14px; }
  @keyframes fadeInUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
  .model-browser-grid .browser-model-card { animation: fadeInUp 0.3s ease backwards; }
  .model-browser-grid .browser-model-card:nth-child(1) { animation-delay: 0.02s; }
  .model-browser-grid .browser-model-card:nth-child(2) { animation-delay: 0.04s; }
  .model-browser-grid .browser-model-card:nth-child(3) { animation-delay: 0.06s; }
  .model-browser-grid .browser-model-card:nth-child(4) { animation-delay: 0.08s; }
  .model-browser-grid .browser-model-card:nth-child(n+5) { animation-delay: 0.1s; }

  .browser-model-card { display: flex; flex-direction: column; gap: 10px; padding: 16px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 14px; transition: all 0.25s ease; position: relative; overflow: hidden; }
  .browser-model-card::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px; background: transparent; transition: background 0.25s ease; }
  .browser-model-card:hover { border-color: var(--color-border-hover); box-shadow: 0 8px 24px rgba(0,0,0,0.08); transform: translateY(-2px); }
  :global(.dark) .browser-model-card:hover { box-shadow: 0 8px 24px rgba(0,0,0,0.25); }
  .browser-model-card.installed { border-color: rgba(34,197,94,0.35); background: linear-gradient(135deg, rgba(34,197,94,0.04), transparent); }
  .browser-model-card.installed::before { background: linear-gradient(90deg, #22c55e, #16a34a); }
  .browser-model-card.default { border-color: rgba(59,130,246,0.4); background: linear-gradient(135deg, rgba(59,130,246,0.05), transparent); }
  .browser-model-card.default::before { background: linear-gradient(90deg, #3b82f6, #2563eb); }
  :global(.dark) .browser-model-card { background: rgba(255,255,255,0.02); border-color: rgba(255,255,255,0.08); }
  :global(.dark) .browser-model-card.installed { background: linear-gradient(135deg, rgba(34,197,94,0.08), transparent); border-color: rgba(34,197,94,0.4); }
  :global(.dark) .browser-model-card.default { background: linear-gradient(135deg, rgba(59,130,246,0.08), transparent); border-color: rgba(59,130,246,0.4); }

  .bmc-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; }
  .bmc-title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
  .bmc-name { font-weight: 700; font-size: 15px; color: var(--color-text); letter-spacing: -0.01em; }
  .bmc-installed-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; background: rgba(34,197,94,0.15); color: #22c55e; border-radius: 6px; display: flex; align-items: center; gap: 4px; }
  .bmc-installed-badge::before { content: '✓'; font-size: 9px; }
  .bmc-meta { display: flex; gap: 6px; flex-shrink: 0; }
  .bmc-size { font-size: 11px; padding: 4px 8px; background: linear-gradient(135deg, rgba(59,130,246,0.1), rgba(59,130,246,0.05)); border-radius: 6px; color: #3b82f6; font-family: 'SF Mono', 'Menlo', monospace; font-weight: 600; }
  .bmc-params { font-size: 11px; padding: 4px 8px; background: linear-gradient(135deg, rgba(168,85,247,0.1), rgba(168,85,247,0.05)); border-radius: 6px; color: #a855f7; font-family: 'SF Mono', 'Menlo', monospace; font-weight: 600; }
  .bmc-description { font-size: 13px; color: var(--color-text-secondary); margin: 0; line-height: 1.4; }
  .bmc-capabilities { display: flex; flex-wrap: wrap; gap: 6px; }
  .cap-badge { display: flex; align-items: center; gap: 4px; padding: 4px 8px; border-radius: 6px; font-size: 11px; }
  .cap-badge.vision { background: rgba(139,92,246,0.12); color: #8b5cf6; }
  .cap-badge.tools { background: rgba(59,130,246,0.12); color: #3b82f6; }
  .cap-badge.coding { background: rgba(34,197,94,0.12); color: #22c55e; }
  .cap-badge.reasoning { background: rgba(249,115,22,0.12); color: #f97316; }
  .cap-badge.rag { background: rgba(6,182,212,0.12); color: #06b6d4; }
  .cap-badge.multilingual { background: rgba(236,72,153,0.12); color: #ec4899; }
  .cap-badge.fast { background: rgba(234,179,8,0.12); color: #eab308; }
  :global(.dark) .cap-badge.vision { background: rgba(139,92,246,0.2); color: #a78bfa; }
  :global(.dark) .cap-badge.tools { background: rgba(59,130,246,0.2); color: #60a5fa; }
  :global(.dark) .cap-badge.coding { background: rgba(34,197,94,0.2); color: #4ade80; }
  :global(.dark) .cap-badge.reasoning { background: rgba(249,115,22,0.2); color: #fb923c; }
  :global(.dark) .cap-badge.rag { background: rgba(6,182,212,0.2); color: #22d3ee; }
  :global(.dark) .cap-badge.multilingual { background: rgba(236,72,153,0.2); color: #f472b6; }
  :global(.dark) .cap-badge.fast { background: rgba(234,179,8,0.2); color: #facc15; }
  .cap-badge-icon-svg { width: 12px; height: 12px; flex-shrink: 0; }
  .cap-badge-label { font-weight: 500; }

  .bmc-variants { display: flex; align-items: flex-start; gap: 10px; padding-top: 10px; border-top: 1px solid var(--color-border); margin-top: 6px; }
  .variants-label { display: flex; align-items: center; gap: 5px; font-size: 11px; color: var(--color-text-muted); flex-shrink: 0; padding-top: 6px; }
  .variant-buttons { display: flex; flex-wrap: wrap; gap: 6px; }
  .variant-buttons.many { gap: 4px; }
  .variant-btn { display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 8px 12px; background: var(--color-bg-secondary); border: 1px solid var(--color-border); border-radius: 8px; cursor: pointer; transition: all 0.2s ease; min-width: 52px; }
  .variant-btn:hover { border-color: var(--color-border-hover); background: var(--color-bg-tertiary); transform: translateY(-1px); }
  .variant-btn.selected { border-color: #3b82f6; background: linear-gradient(135deg, rgba(59,130,246,0.15), rgba(59,130,246,0.05)); box-shadow: 0 0 0 2px rgba(59,130,246,0.2); }
  :global(.dark) .variant-btn { background: rgba(255,255,255,0.03); border-color: rgba(255,255,255,0.08); }
  :global(.dark) .variant-btn:hover { background: rgba(255,255,255,0.06); }
  :global(.dark) .variant-btn.selected { background: linear-gradient(135deg, rgba(59,130,246,0.25), rgba(59,130,246,0.1)); box-shadow: 0 0 0 2px rgba(59,130,246,0.3); }
  .variant-params { font-size: 12px; font-weight: 700; color: var(--color-text); font-family: 'SF Mono','Menlo',monospace; }
  .variant-btn.selected .variant-params { color: #3b82f6; }
  .variant-size { font-size: 10px; color: var(--color-text-muted); font-family: 'SF Mono','Menlo',monospace; }
  .variant-btn.selected .variant-size { color: #60a5fa; }

  .bmc-footer { display: flex; align-items: center; justify-content: space-between; margin-top: auto; padding-top: 10px; border-top: 1px solid var(--color-border); }
  .bmc-downloads { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--color-text-muted); font-weight: 500; }
  .bmc-actions { display: flex; gap: 8px; margin-left: auto; }
  .bmc-btn { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s ease; border: none; }
  .bmc-btn svg { width: 14px; height: 14px; }
  .bmc-btn.pull-btn { background: linear-gradient(135deg, #3b82f6, #2563eb); color: white; margin-left: auto; box-shadow: 0 2px 8px rgba(59,130,246,0.25); }
  .bmc-btn.pull-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(59,130,246,0.35); }
  .bmc-btn.pull-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; box-shadow: none; }
  .bmc-btn.default-btn { background: rgba(128,128,128,0.15); border: 1px solid var(--color-border); color: var(--color-text); }
  .bmc-btn.default-btn:hover { background: rgba(128,128,128,0.25); }
  .bmc-btn.default-btn.is-default { background: var(--color-primary); border-color: var(--color-primary); color: white; }
  :global(.dark) .bmc-btn.default-btn { background: rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.15); }
  :global(.dark) .bmc-btn.default-btn:hover { background: rgba(255,255,255,0.15); }
  .bmc-btn.delete-btn { background: transparent; color: var(--color-error); padding: 8px; border: 1px solid var(--color-border); }
  .bmc-btn.delete-btn:hover { background: rgba(220,38,38,0.1); border-color: var(--color-error); }

  .btn-spinner-small { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%; animation: spin 0.8s linear infinite; }

  .pull-card-compact { padding: 16px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 12px; }
  .pull-compact-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
  .pull-compact-header h4 { margin: 0; font-size: 14px; font-weight: 600; }
  .browse-link { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--color-primary); text-decoration: none; }
  .browse-link:hover { text-decoration: underline; }
  .pull-form-compact { display: flex; gap: 8px; }
  .pull-form-compact input { flex: 1; padding: 10px 14px; background: var(--color-bg-secondary); border: 1px solid var(--color-border); border-radius: 8px; font-size: 13px; color: var(--color-text); }
  .pull-form-compact input:focus { outline: none; border-color: var(--color-primary); }
  .pull-form-compact input::placeholder { color: var(--color-text-muted); }
  .pull-btn-compact { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: var(--color-primary); color: white; border: none; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
  .pull-btn-compact:hover { opacity: 0.9; }
  .pull-btn-compact:disabled { opacity: 0.6; cursor: not-allowed; }
  .pull-btn-compact svg { width: 16px; height: 16px; }
  .pull-progress-compact { margin-top: 12px; padding: 12px; background: var(--color-bg-secondary); border-radius: 8px; }
  .progress-info { display: flex; justify-content: space-between; margin-bottom: 8px; font-size: 12px; }
  .progress-percent { font-weight: 600; color: var(--color-primary); }
  .progress-bar-compact { height: 6px; background: var(--color-border); border-radius: 3px; overflow: hidden; }
  .progress-fill { height: 100%; background: var(--color-primary); transition: width 0.3s; }
  .progress-details { display: flex; justify-content: space-between; margin-top: 8px; font-size: 11px; color: var(--color-text-muted); }
  .pull-error-compact { margin-top: 12px; padding: 10px 14px; background: rgba(220,38,38,0.1); border-radius: 8px; color: var(--color-error); font-size: 13px; }

  /* Settings Tab */
  .presets-row { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; flex-wrap: wrap; }
  .presets-label { font-size: 13px; color: var(--color-text-tertiary); margin-right: 4px; }
  .preset-btn { display: flex; align-items: center; gap: 6px; padding: 8px 14px; font-size: 12px; font-weight: 500; border-radius: 8px; border: 1px solid var(--color-border); background: rgba(255,255,255,0.05); color: var(--color-text-secondary); cursor: pointer; transition: all 0.15s ease; }
  :global(.dark) .preset-btn { background: rgba(255,255,255,0.05); border-color: rgba(255,255,255,0.1); }
  :global(:not(.dark)) .preset-btn { background: rgba(0,0,0,0.02); }
  .preset-btn:hover { background: rgba(255,255,255,0.1); border-color: var(--color-primary); color: var(--color-primary); }
  :global(:not(.dark)) .preset-btn:hover { background: rgba(0,0,0,0.05); }
  .preset-icon { width: 16px; height: 16px; flex-shrink: 0; }

  .settings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
  .setting-card { padding: 18px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 12px; }
  .setting-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
  .setting-header label { font-weight: 500; font-size: 14px; }
  .setting-value { font-size: 14px; font-weight: 600; color: var(--color-primary); font-family: monospace; }
  .setting-card input[type="range"] { width: 100%; height: 6px; background: var(--color-border); border-radius: 3px; appearance: none; cursor: pointer; }
  .setting-card input[type="range"]::-webkit-slider-thumb { appearance: none; width: 18px; height: 18px; background: var(--color-primary); border-radius: 50%; cursor: pointer; }
  .setting-desc { margin: 10px 0 0; font-size: 12px; color: var(--color-text-muted); }
  .toggle-card .setting-header { margin-bottom: 0; }
  .toggle { width: 44px; height: 24px; background: var(--color-border); border: none; border-radius: 12px; cursor: pointer; position: relative; transition: all 0.2s; }
  .toggle.on { background: var(--color-success); }
  .toggle-knob { position: absolute; top: 2px; left: 2px; width: 20px; height: 20px; background: white; border-radius: 50%; transition: all 0.2s; }
  .toggle.on .toggle-knob { left: 22px; }
  .toggle-card .setting-desc { margin-top: 12px; }
  .settings-actions { display: flex; gap: 10px; margin-top: 20px; }
  .action-btn { padding: 12px 24px; background: var(--color-bg-tertiary); border: 1px solid var(--color-border); border-radius: 10px; font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.2s; color: var(--color-text); }
  .action-btn:hover { background: var(--color-bg-secondary); }
  .action-btn.primary { background: var(--color-primary); color: var(--color-bg); border-color: transparent; }
  .action-btn.primary:hover { opacity: 0.9; }
  .action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .save-btn { padding: 10px 16px; background: var(--color-primary); color: var(--color-bg); border: none; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; }
  .save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  :global(.dark) .setting-value { color: #0A84FF; }

  /* Default Model */
  .default-model-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; }
  .default-model-btn { position: relative; display: flex; flex-direction: column; gap: 4px; padding: 14px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 10px; text-align: left; cursor: pointer; transition: all 0.2s; }
  .default-model-btn:hover { border-color: var(--color-border-hover); }
  .default-model-btn.selected { border-color: var(--color-primary); background: rgba(0,0,0,0.02); }
  .dm-name { font-weight: 600; font-size: 14px; }
  .dm-size, .dm-desc { font-size: 12px; color: var(--color-text-muted); }
  .dm-check { position: absolute; top: 10px; right: 10px; width: 18px; height: 18px; color: var(--color-success); }
  :global(.dark) .default-model-btn.selected { border-color: #0A84FF; background: rgba(10,132,255,0.1); }
  :global(.dark) .setting-card { background: #2c2c2e; border-color: rgba(255,255,255,0.1); }
</style>
