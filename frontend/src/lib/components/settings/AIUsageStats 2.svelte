<script lang="ts">
  import type { UsageStats, SystemInfo, LLMProvider } from '$lib/stores/aiSettings';
  import { formatTokens, formatDuration, getProviderIconPath, getProviderLabel, getCategoryIconPath } from '$lib/stores/aiSettings';

  interface Props {
    usageStats: UsageStats | null;
    loadingUsage: boolean;
    systemInfo: SystemInfo | null;
    providers: LLMProvider[];
    localModelsCount: number;
    onPeriodChange: (period: 'today' | 'week' | 'month' | 'all') => void;
    onRefresh: () => void;
  }

  let {
    usageStats,
    loadingUsage,
    systemInfo,
    providers,
    localModelsCount,
    onPeriodChange,
    onRefresh
  }: Props = $props();

  let usagePeriod = $state<'today' | 'week' | 'month' | 'all'>('month');

  function selectPeriod(period: 'today' | 'week' | 'month' | 'all') {
    usagePeriod = period;
    onPeriodChange(period);
  }
</script>

<div class="stats-page">
  <!-- Stats Header with Period Selector -->
  <div class="stats-header">
    <div class="stats-title-area">
      <span class="stats-eyebrow">Analytics</span>
      {#if usageStats}
        <button class="stats-date-picker">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
          <span>{usageStats.period_start} — {usageStats.period_end}</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><polyline points="6 9 12 15 18 9"/></svg>
        </button>
      {/if}
    </div>
    <div class="stats-controls">
      <div class="period-selector">
        <button class="period-btn" class:active={usagePeriod === 'today'} onclick={() => selectPeriod('today')}>Today</button>
        <button class="period-btn" class:active={usagePeriod === 'week'} onclick={() => selectPeriod('week')}>Week</button>
        <button class="period-btn" class:active={usagePeriod === 'month'} onclick={() => selectPeriod('month')}>Month</button>
        <button class="period-btn" class:active={usagePeriod === 'all'} onclick={() => selectPeriod('all')}>All Time</button>
      </div>
      <button class="refresh-btn" onclick={onRefresh} disabled={loadingUsage} title="Refresh stats" aria-label="Refresh stats">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class:spinning={loadingUsage}><path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
      </button>
    </div>
  </div>

  {#if loadingUsage}
    <div class="loading">
      <div class="spinner"></div>
      <span>Loading analytics...</span>
    </div>
  {:else if usageStats}
    <!-- System Overview Row -->
    <div class="stats-system-row">
      <div class="system-card">
        <div class="system-platform-info">
          <div class="platform-icon-large">
            {#if systemInfo?.platform === 'darwin'}
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/></svg>
            {:else}
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3,12V6.75L9,5.43V11.91L3,12M20,3V11.75L10,11.9V5.21L20,3M3,13L9,13.09V19.9L3,18.75V13M20,13.25V22L10,20.09V13.1L20,13.25Z"/></svg>
            {/if}
          </div>
          <div class="platform-text">
            <span class="platform-label">{systemInfo?.platform === 'darwin' ? 'macOS' : systemInfo?.platform || 'System'}</span>
            {#if systemInfo?.has_gpu}
              <span class="gpu-badge">{systemInfo.gpu_name || 'Apple Silicon / Metal'}</span>
            {/if}
          </div>
        </div>
        <div class="system-metrics">
          <div class="sys-metric">
            <span class="sys-metric-value">{localModelsCount}</span>
            <span class="sys-metric-label">Local Models</span>
          </div>
          <div class="sys-metric">
            <span class="sys-metric-value">{usageStats.local_model_storage_gb}GB</span>
            <span class="sys-metric-label">Storage Used</span>
          </div>
          <div class="sys-metric">
            <span class="sys-metric-value">{providers.filter(p => p.configured).length}</span>
            <span class="sys-metric-label">Providers</span>
          </div>
        </div>
      </div>
      {#if systemInfo}
        <div class="ram-gauge-card">
          <div class="ram-gauge-header">
            <span class="ram-title">Memory</span>
            <span class="ram-detail">{systemInfo.available_ram_gb.toFixed(0)}GB free / {systemInfo.total_ram_gb}GB</span>
          </div>
          <div class="ram-gauge-visual">
            <svg class="gauge-svg-large" viewBox="0 0 120 120">
              <circle class="gauge-bg-large" cx="60" cy="60" r="50" fill="none" />
              <circle
                class="gauge-fill-large"
                cx="60" cy="60" r="50"
                fill="none"
                stroke-dasharray="{(systemInfo.available_ram_gb / systemInfo.total_ram_gb) * 314} 314"
                transform="rotate(-90 60 60)"
              />
            </svg>
            <div class="gauge-center-text">
              <span class="gauge-percent">{Math.round((systemInfo.available_ram_gb / systemInfo.total_ram_gb) * 100)}%</span>
              <span class="gauge-subtitle">Free</span>
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Key Metrics Grid -->
    <div class="key-metrics-grid">
      <div class="metric-card requests">
        <div class="metric-sparkline">
          {#if usageStats.recent.length > 0}
            {@const recent = usageStats.recent}
            <svg viewBox="0 0 100 30" preserveAspectRatio="none">
              <polyline
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                points="{recent.map((d, i) => `${i * (100 / (recent.length - 1 || 1))},${30 - (d.requests / Math.max(...recent.map(r => r.requests), 1)) * 25}`).join(' ')}"
              />
            </svg>
          {/if}
        </div>
        <div class="metric-header">
          <div class="metric-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          </div>
          <span class="metric-trend up">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="12" height="12"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
          </span>
        </div>
        <div class="metric-content">
          <span class="metric-value animate-number">{usageStats.total_requests.toLocaleString()}</span>
          <span class="metric-label">Total Requests</span>
        </div>
        <div class="metric-sub">
          <span class="metric-badge">{usageStats.session_count} sessions</span>
        </div>
      </div>
      <div class="metric-card tokens">
        <div class="metric-sparkline">
          {#if usageStats.recent.length > 0}
            {@const recent = usageStats.recent}
            <svg viewBox="0 0 100 30" preserveAspectRatio="none">
              <polyline
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                points="{recent.map((d, i) => `${i * (100 / (recent.length - 1 || 1))},${30 - (d.tokens / Math.max(...recent.map(r => r.tokens), 1)) * 25}`).join(' ')}"
              />
            </svg>
          {/if}
        </div>
        <div class="metric-header">
          <div class="metric-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="2" width="16" height="20" rx="2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
          </div>
        </div>
        <div class="metric-content">
          <span class="metric-value">{formatTokens(usageStats.total_tokens)}</span>
          <span class="metric-label">Total Tokens</span>
        </div>
        <div class="metric-breakdown">
          <span class="metric-in"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="10" height="10"><polyline points="19 12 12 19 5 12"/><line x1="12" y1="19" x2="12" y2="5"/></svg> {formatTokens(usageStats.input_tokens)}</span>
          <span class="metric-out"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="10" height="10"><polyline points="5 12 12 5 19 12"/><line x1="12" y1="5" x2="12" y2="19"/></svg> {formatTokens(usageStats.output_tokens)}</span>
        </div>
      </div>
      <div class="metric-card speed">
        <div class="metric-header">
          <div class="metric-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
          </div>
          {#if usageStats.avg_response_time_ms < 500}
            <span class="metric-badge good">Fast</span>
          {:else if usageStats.avg_response_time_ms < 2000}
            <span class="metric-badge">Average</span>
          {:else}
            <span class="metric-badge slow">Slow</span>
          {/if}
        </div>
        <div class="metric-content">
          <span class="metric-value">{formatDuration(usageStats.avg_response_time_ms)}</span>
          <span class="metric-label">Avg Response</span>
        </div>
        <div class="metric-sub">
          <span class="metric-badge">{usageStats.avg_requests_per_session} req/session</span>
        </div>
      </div>
      <div class="metric-card cost">
        <div class="metric-header">
          <div class="metric-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
          </div>
        </div>
        <div class="metric-content">
          <span class="metric-value">${(usageStats.cloud_api_cost + usageStats.local_power_cost_estimate).toFixed(2)}</span>
          <span class="metric-label">Total Cost</span>
        </div>
        <div class="metric-breakdown cost-breakdown">
          <span class="metric-cloud" title="Cloud API Cost">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/></svg>
            ${usageStats.cloud_api_cost.toFixed(2)}
          </span>
          <span class="metric-local" title="Est. Power Cost">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
            ${usageStats.local_power_cost_estimate.toFixed(2)}
          </span>
        </div>
      </div>
    </div>

    <!-- Local vs Cloud Comparison -->
    <div class="comparison-section">
      <div class="section-header">
        <h3>Local vs Cloud Usage</h3>
        {#if usageStats.total_requests > 0}
          <span class="section-hint">Based on {usageStats.total_requests} requests</span>
        {/if}
      </div>
      {#if usageStats.total_requests === 0}
        <div class="comparison-empty">
          <div class="comparison-empty-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48">
              <path d="M21.21 15.89A10 10 0 1 1 8 2.83"/>
              <path d="M22 12A10 10 0 0 0 12 2v10z"/>
            </svg>
          </div>
          <p class="comparison-empty-text">Start using AI features to see your local vs cloud breakdown</p>
          <p class="comparison-empty-hint">Local inference is free, cloud APIs are billed per token</p>
        </div>
      {:else}
        {@const localReqs = usageStats.by_provider['ollama_local']?.requests || 0}
        {@const cloudReqs = Object.entries(usageStats.by_provider).filter(([k]) => k !== 'ollama_local').reduce((sum, [, v]) => sum + v.requests, 0)}
        {@const localPct = (localReqs / usageStats.total_requests * 100)}
        {@const cloudPct = (cloudReqs / usageStats.total_requests * 100)}
        <div class="comparison-split-bar">
          <div class="split-bar-container">
            <div class="split-bar-fill local" style="width: {localPct}%">
              {#if localPct > 15}<span class="split-bar-label">{localPct.toFixed(0)}%</span>{/if}
            </div>
            <div class="split-bar-fill cloud" style="width: {cloudPct}%">
              {#if cloudPct > 15}<span class="split-bar-label">{cloudPct.toFixed(0)}%</span>{/if}
            </div>
          </div>
        </div>
        <div class="comparison-grid">
          <div class="comparison-card local">
            <div class="comp-header">
              <div class="comp-icon-wrapper local">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                  <rect x="2" y="3" width="20" height="14" rx="2"/>
                  <path d="M8 21h8"/>
                  <path d="M12 17v4"/>
                </svg>
              </div>
              <div class="comp-title-group">
                <span class="comp-title">Local Inference</span>
                <span class="comp-subtitle">On-device processing</span>
              </div>
            </div>
            <div class="comp-stats">
              <div class="comp-stat">
                <span class="comp-value">{localReqs}</span>
                <span class="comp-label">Requests</span>
              </div>
              <div class="comp-stat">
                <span class="comp-value">{formatTokens(usageStats.by_provider['ollama_local']?.tokens || 0)}</span>
                <span class="comp-label">Tokens</span>
              </div>
              <div class="comp-stat">
                <span class="comp-value">${usageStats.local_power_cost_estimate.toFixed(2)}</span>
                <span class="comp-label" title="Estimated electricity cost based on average GPU power usage">Est. Power</span>
              </div>
            </div>
          </div>
          <div class="comparison-card cloud">
            <div class="comp-header">
              <div class="comp-icon-wrapper cloud">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                  <path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
                </svg>
              </div>
              <div class="comp-title-group">
                <span class="comp-title">Cloud API</span>
                <span class="comp-subtitle">Remote processing</span>
              </div>
            </div>
            <div class="comp-stats">
              <div class="comp-stat">
                <span class="comp-value">{cloudReqs}</span>
                <span class="comp-label">Requests</span>
              </div>
              <div class="comp-stat">
                <span class="comp-value">{formatTokens(Object.entries(usageStats.by_provider).filter(([k]) => k !== 'ollama_local').reduce((sum, [, v]) => sum + v.tokens, 0))}</span>
                <span class="comp-label">Tokens</span>
              </div>
              <div class="comp-stat">
                <span class="comp-value">${usageStats.cloud_api_cost.toFixed(2)}</span>
                <span class="comp-label">API Cost</span>
              </div>
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Detailed Breakdowns -->
    <div class="breakdowns-grid">
      <!-- Provider Breakdown -->
      <div class="breakdown-card">
        <div class="breakdown-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M2 20h.01M7 20v-4M12 20v-8M17 20V8M22 4v16"/></svg>
          <h4>By Provider</h4>
        </div>
        {#if Object.keys(usageStats.by_provider).length === 0}
          <div class="breakdown-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><circle cx="12" cy="12" r="10"/><path d="M8 12h8"/><path d="M12 8v8"/></svg>
            <span>No provider data yet</span>
          </div>
        {:else}
          <div class="breakdown-list">
            {#each Object.entries(usageStats.by_provider).sort((a, b) => b[1].requests - a[1].requests) as [provider, stats]}
              <div class="breakdown-row">
                <div class="breakdown-info">
                  <span class="breakdown-icon provider">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d={getProviderIconPath(provider)}/></svg>
                  </span>
                  <span class="breakdown-name">{getProviderLabel(provider)}</span>
                </div>
                <div class="breakdown-values">
                  <span class="breakdown-stat-value requests">{stats.requests}</span>
                  <span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
                  <span class="breakdown-stat-value cost">${stats.cost.toFixed(2)}</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-bar-fill provider" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Model Breakdown -->
      <div class="breakdown-card">
        <div class="breakdown-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
          <h4>By Model</h4>
        </div>
        {#if Object.keys(usageStats.by_model).length === 0}
          <div class="breakdown-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h6v6H9z"/></svg>
            <span>No model usage yet</span>
          </div>
        {:else}
          <div class="breakdown-list">
            {#each Object.entries(usageStats.by_model).sort((a, b) => b[1].requests - a[1].requests).slice(0, 5) as [model, stats]}
              <div class="breakdown-row">
                <div class="breakdown-info">
                  <span class="breakdown-name truncate">{model}</span>
                </div>
                <div class="breakdown-values">
                  <span class="breakdown-stat-value requests">{stats.requests}</span>
                  <span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
                  <span class="breakdown-stat-value latency">{formatDuration(stats.avg_latency_ms)}</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-bar-fill model" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Agent Breakdown -->
      <div class="breakdown-card">
        <div class="breakdown-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          <h4>By Agent</h4>
        </div>
        {#if Object.keys(usageStats.by_agent).length === 0}
          <div class="breakdown-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>
            <span>No agent usage yet</span>
          </div>
        {:else}
          <div class="breakdown-list">
            {#each Object.entries(usageStats.by_agent).sort((a, b) => b[1].requests - a[1].requests).slice(0, 5) as [agent, stats]}
              <div class="breakdown-row">
                <div class="breakdown-info">
                  <span class="breakdown-icon agent">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d={getCategoryIconPath(agent.toLowerCase())}/></svg>
                  </span>
                  <span class="breakdown-name">{agent}</span>
                </div>
                <div class="breakdown-values">
                  <span class="breakdown-stat-value requests">{stats.requests}</span>
                  <span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-bar-fill agent" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Activity Timeline -->
    <div class="activity-section">
      <div class="section-header">
        <h3>Activity Timeline</h3>
        {#if usageStats.recent.length > 0}
          <span class="section-hint">{usageStats.recent.length} days shown</span>
        {/if}
      </div>
      {#if usageStats.recent.length > 0}
        {@const maxRequests = Math.max(...usageStats.recent.map(d => d.requests), 1)}
        <div class="activity-chart">
          <div class="activity-area-chart">
            <svg viewBox="0 0 {usageStats.recent.length * 50} 120" preserveAspectRatio="none">
              <path
                d="M 0 120 {usageStats.recent.map((d, i) => `L ${i * 50 + 25} ${120 - (d.requests / maxRequests) * 100}`).join(' ')} L {(usageStats.recent.length - 1) * 50 + 25} 120 Z"
                fill="url(#areaGradient)"
                opacity="0.3"
              />
              <polyline
                fill="none"
                stroke="var(--color-primary)"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
                points="{usageStats.recent.map((d, i) => `${i * 50 + 25},${120 - (d.requests / maxRequests) * 100}`).join(' ')}"
              />
              {#each usageStats.recent as day, i}
                <circle
                  cx="{i * 50 + 25}"
                  cy="{120 - (day.requests / maxRequests) * 100}"
                  r="4"
                  fill="var(--color-bg)"
                  stroke="var(--color-primary)"
                  stroke-width="2"
                />
              {/each}
              <defs>
                <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="var(--color-primary)"/>
                  <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0"/>
                </linearGradient>
              </defs>
            </svg>
          </div>
          <div class="activity-labels">
            {#each usageStats.recent as day}
              <div class="activity-label-item">
                <span class="activity-day">{new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}</span>
                <span class="activity-count">{day.requests}</span>
              </div>
            {/each}
          </div>
        </div>
      {:else}
        <div class="activity-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="40" height="40">
            <path d="M3 3v18h18"/>
            <path d="M18 17V9"/>
            <path d="M13 17V5"/>
            <path d="M8 17v-3"/>
          </svg>
          <p>Activity data will appear as you use AI features</p>
        </div>
      {/if}
    </div>

    <!-- Session Stats -->
    <div class="session-stats">
      <div class="section-header">
        <h3>Session Statistics</h3>
      </div>
      <div class="session-grid">
        <div class="session-stat-card">
          <div class="session-stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          </div>
          <div class="session-stat-content">
            <span class="session-value">{usageStats.session_count}</span>
            <span class="session-label">Total Sessions</span>
          </div>
        </div>
        <div class="session-stat-card">
          <div class="session-stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
          <div class="session-stat-content">
            <span class="session-value">{usageStats.avg_session_duration_min}<span class="session-unit">min</span></span>
            <span class="session-label">Avg Duration</span>
          </div>
        </div>
        <div class="session-stat-card">
          <div class="session-stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          </div>
          <div class="session-stat-content">
            <span class="session-value">{usageStats.avg_requests_per_session}</span>
            <span class="session-label">Requests/Session</span>
          </div>
        </div>
        <div class="session-stat-card">
          <div class="session-stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><rect x="4" y="2" width="16" height="20" rx="2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
          </div>
          <div class="session-stat-content">
            <span class="session-value">{usageStats.total_requests > 0 ? Math.round(usageStats.total_tokens / usageStats.total_requests) : 0}</span>
            <span class="session-label">Tokens/Request</span>
          </div>
        </div>
      </div>
    </div>
  {:else}
    <div class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
      </div>
      <h3>No Usage Data</h3>
      <p>Usage statistics will appear here once you start using AI features</p>
    </div>
  {/if}
</div>

<style>
  .stats-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  /* Header */
  .stats-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 16px;
  }

  .stats-title-area {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .stats-eyebrow {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--color-text-muted);
  }

  .stats-date-picker {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .stats-date-picker:hover {
    background: var(--color-bg-tertiary);
    border-color: var(--color-text-muted);
  }

  :global(.dark) .stats-date-picker {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .stats-date-picker:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .stats-date-picker svg { color: var(--color-text-muted); }

  .stats-controls {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .period-selector {
    display: flex;
    gap: 4px;
    padding: 4px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 10px;
  }

  :global(.dark) .period-selector {
    background: rgba(255, 255, 255, 0.05);
  }

  .period-btn {
    padding: 8px 14px;
    background: transparent;
    border: none;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-secondary);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .period-btn:hover {
    color: var(--color-text);
    background: rgba(0, 0, 0, 0.05);
  }

  :global(.dark) .period-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .period-btn.active {
    background: #3b82f6 !important;
    color: white !important;
  }

  /* Refresh button */
  .refresh-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    background: var(--color-bg-tertiary);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    font-size: 13px;
    color: var(--color-text);
    cursor: pointer;
    transition: all 0.2s;
  }

  .refresh-btn:hover { background: var(--color-bg-secondary); }
  .refresh-btn:disabled { opacity: 0.6; cursor: not-allowed; }
  .refresh-btn svg { width: 16px; height: 16px; }

  .refresh-btn svg.spinning {
    animation: spin 1s linear infinite;
  }

  /* System Row */
  .stats-system-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 20px;
  }

  .system-card {
    padding: 20px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
  }

  .system-platform-info {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;
  }

  .platform-icon-large {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-bg);
    border-radius: 12px;
  }

  .platform-icon-large svg {
    width: 28px;
    height: 28px;
    color: var(--color-text);
  }

  .platform-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .platform-label {
    font-size: 18px;
    font-weight: 600;
  }

  .gpu-badge {
    font-size: 12px;
    padding: 3px 10px;
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.15), rgba(34, 197, 94, 0.05));
    color: #22c55e;
    border-radius: 20px;
    width: fit-content;
  }

  :global(.dark) .gpu-badge {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(34, 197, 94, 0.08));
  }

  .system-metrics {
    display: flex;
    gap: 32px;
  }

  .sys-metric {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .sys-metric-value {
    font-size: 24px;
    font-weight: 700;
    color: var(--color-text);
  }

  .sys-metric-label {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  /* RAM Gauge */
  .ram-gauge-card {
    padding: 20px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    min-width: 180px;
  }

  .ram-gauge-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 12px;
  }

  .ram-title { font-size: 14px; font-weight: 600; }
  .ram-detail { font-size: 12px; color: var(--color-text-muted); }

  .ram-gauge-visual {
    position: relative;
    width: 120px;
    height: 120px;
    margin: 0 auto;
  }

  .gauge-svg-large { width: 100%; height: 100%; }
  .gauge-bg-large { stroke: var(--color-border); stroke-width: 8; }
  .gauge-fill-large {
    stroke: #22c55e;
    stroke-width: 8;
    stroke-linecap: round;
    transition: stroke-dasharray 0.5s ease;
  }

  .gauge-center-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
  }

  .gauge-percent {
    display: block;
    font-size: 24px;
    font-weight: 700;
    color: #22c55e;
  }

  .gauge-subtitle { font-size: 11px; color: var(--color-text-muted); }

  /* Key Metrics Grid */
  .key-metrics-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
  }

  .metric-card {
    position: relative;
    padding: 20px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow: hidden;
    transition: all 0.2s ease;
  }

  .metric-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .metric-card:hover {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  .metric-card.requests {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.12), rgba(59, 130, 246, 0.02));
    border-color: rgba(59, 130, 246, 0.25);
  }
  .metric-card.requests .metric-icon svg { color: #3b82f6; }
  .metric-card.requests .metric-sparkline { color: rgba(59, 130, 246, 0.5); }

  .metric-card.tokens {
    background: linear-gradient(135deg, rgba(168, 85, 247, 0.12), rgba(168, 85, 247, 0.02));
    border-color: rgba(168, 85, 247, 0.25);
  }
  .metric-card.tokens .metric-icon svg { color: #a855f7; }
  .metric-card.tokens .metric-sparkline { color: rgba(168, 85, 247, 0.5); }

  .metric-card.speed {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.12), rgba(34, 197, 94, 0.02));
    border-color: rgba(34, 197, 94, 0.25);
  }
  .metric-card.speed .metric-icon svg { color: #22c55e; }

  .metric-card.cost {
    background: linear-gradient(135deg, rgba(249, 115, 22, 0.12), rgba(249, 115, 22, 0.02));
    border-color: rgba(249, 115, 22, 0.25);
  }
  .metric-card.cost .metric-icon svg { color: #f97316; }

  .metric-sparkline {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 80px;
    height: 30px;
    opacity: 0.6;
  }

  .metric-sparkline svg { width: 100%; height: 100%; }

  .metric-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .metric-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 10px;
  }

  :global(.dark) .metric-icon { background: rgba(255, 255, 255, 0.05); }
  .metric-icon svg { width: 20px; height: 20px; }

  .metric-trend {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    font-weight: 600;
    padding: 4px 8px;
    border-radius: 20px;
  }

  .metric-trend.up { color: #22c55e; background: rgba(34, 197, 94, 0.15); }
  .metric-trend.down { color: #ef4444; background: rgba(239, 68, 68, 0.15); }

  .metric-content {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .metric-value {
    font-size: 32px;
    font-weight: 700;
    font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
    letter-spacing: -0.02em;
  }

  .metric-label {
    font-size: 13px;
    color: var(--color-text-muted);
    font-weight: 500;
  }

  .metric-sub { margin-top: 8px; }

  .metric-badge {
    display: inline-block;
    font-size: 11px;
    font-weight: 500;
    padding: 4px 10px;
    background: rgba(0, 0, 0, 0.05);
    border-radius: 20px;
    color: var(--color-text-muted);
  }

  :global(.dark) .metric-badge { background: rgba(255, 255, 255, 0.08); }
  .metric-badge.good { color: #22c55e; background: rgba(34, 197, 94, 0.12); }
  .metric-badge.slow { color: #ef4444; background: rgba(239, 68, 68, 0.12); }

  .metric-breakdown {
    display: flex;
    gap: 12px;
    font-size: 12px;
    font-weight: 500;
    margin-top: 8px;
  }

  .metric-in { display: flex; align-items: center; gap: 4px; color: #22c55e; }
  .metric-out { display: flex; align-items: center; gap: 4px; color: #f97316; }

  .cost-breakdown { flex-direction: column; gap: 6px; }
  .metric-cloud, .metric-local { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; }
  .metric-cloud { color: #3b82f6; }
  .metric-local { color: #22c55e; }
  .metric-cloud svg, .metric-local svg { flex-shrink: 0; opacity: 0.7; }

  /* Comparison Section */
  .comparison-section {
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    padding: 24px;
  }

  .comparison-section h3 { margin: 0; font-size: 15px; font-weight: 600; }

  .comparison-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
    text-align: center;
  }

  .comparison-empty-icon { color: var(--color-text-muted); margin-bottom: 16px; opacity: 0.5; }
  .comparison-empty-text { font-size: 14px; color: var(--color-text-muted); margin: 0 0 8px 0; }
  .comparison-empty-hint { font-size: 12px; color: var(--color-text-muted); opacity: 0.7; margin: 0; }

  .comparison-split-bar { margin-bottom: 20px; }

  .split-bar-container {
    display: flex;
    height: 32px;
    border-radius: 8px;
    overflow: hidden;
    background: var(--color-bg);
  }

  .split-bar-fill {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 30px;
    transition: width 0.5s ease;
  }

  .split-bar-fill.local { background: linear-gradient(90deg, #22c55e, #16a34a); }
  .split-bar-fill.cloud { background: linear-gradient(90deg, #3b82f6, #2563eb); }
  .split-bar-label { font-size: 12px; font-weight: 600; color: white; text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2); }

  .comparison-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .comparison-card {
    padding: 20px;
    background: var(--color-bg);
    border-radius: 14px;
    border: 1px solid var(--color-border);
    transition: all 0.2s ease;
  }

  .comparison-card:hover { border-color: var(--color-text-muted); }

  .comparison-card.local {
    border-color: rgba(34, 197, 94, 0.3);
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.05), transparent);
  }

  .comparison-card.local:hover { border-color: rgba(34, 197, 94, 0.5); }

  .comparison-card.cloud {
    border-color: rgba(59, 130, 246, 0.3);
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), transparent);
  }

  .comparison-card.cloud:hover { border-color: rgba(59, 130, 246, 0.5); }

  :global(.dark) .comparison-card.local {
    border-color: rgba(34, 197, 94, 0.4);
    background: rgba(34, 197, 94, 0.05);
  }

  :global(.dark) .comparison-card.cloud {
    border-color: rgba(59, 130, 246, 0.4);
    background: rgba(59, 130, 246, 0.05);
  }

  .comp-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }

  .comp-icon-wrapper {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
  }

  .comp-icon-wrapper.local { background: rgba(34, 197, 94, 0.15); color: #22c55e; }
  .comp-icon-wrapper.cloud { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }

  .comp-title-group { display: flex; flex-direction: column; gap: 2px; }
  .comp-subtitle { font-size: 11px; color: var(--color-text-muted); }
  .comp-title { font-size: 15px; font-weight: 600; }
  .comp-stats { display: flex; gap: 24px; margin-bottom: 16px; }
  .comp-stat { display: flex; flex-direction: column; gap: 2px; }
  .comp-value { font-size: 20px; font-weight: 700; }
  .comp-label { font-size: 11px; color: var(--color-text-muted); }

  /* Breakdowns Grid */
  .breakdowns-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
  }

  .breakdown-card {
    padding: 20px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    min-height: 200px;
    display: flex;
    flex-direction: column;
  }

  .breakdown-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
  }

  .breakdown-header svg { color: var(--color-text-muted); }
  .breakdown-header h4 { margin: 0; font-size: 14px; font-weight: 600; }

  .breakdown-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--color-text-muted);
    opacity: 0.6;
  }

  .breakdown-empty span { font-size: 13px; }

  .breakdown-list { display: flex; flex-direction: column; gap: 12px; }

  .breakdown-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 10px;
    background: var(--color-bg);
    border-radius: 8px;
    transition: all 0.15s ease;
  }

  .breakdown-row:hover { background: var(--color-bg-tertiary); }

  :global(.dark) .breakdown-row { background: rgba(255, 255, 255, 0.03); }
  :global(.dark) .breakdown-row:hover { background: rgba(255, 255, 255, 0.06); }

  .breakdown-info { display: flex; align-items: center; gap: 8px; }

  .breakdown-icon {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    border-radius: 6px;
  }

  .breakdown-icon.provider { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .breakdown-icon.agent { background: rgba(168, 85, 247, 0.1); color: #a855f7; }
  .breakdown-icon svg { width: 14px; height: 14px; }

  .breakdown-name { font-size: 13px; font-weight: 500; }
  .breakdown-name.truncate { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 120px; }

  .breakdown-values { display: flex; gap: 12px; font-size: 11px; font-family: 'SF Mono', 'Menlo', monospace; }
  .breakdown-stat-value { font-weight: 500; }
  .breakdown-stat-value.requests { color: #3b82f6; }
  .breakdown-stat-value.tokens { color: #f97316; }
  .breakdown-stat-value.cost { color: #22c55e; }
  .breakdown-stat-value.latency { color: #a855f7; }

  .breakdown-bar { height: 4px; background: var(--color-border); border-radius: 2px; overflow: hidden; }
  .breakdown-bar-fill { height: 100%; border-radius: 2px; transition: width 0.5s ease; }
  .breakdown-bar-fill.provider { background: #3b82f6; }
  .breakdown-bar-fill.model { background: #f97316; }
  .breakdown-bar-fill.agent { background: #a855f7; }

  /* Activity Chart */
  .activity-section {
    padding: 24px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
  }

  .activity-section h3 { margin: 0; font-size: 15px; font-weight: 600; }

  .activity-chart { display: flex; flex-direction: column; gap: 12px; }

  .activity-area-chart { height: 120px; border-radius: 8px; overflow: hidden; }
  .activity-area-chart svg { width: 100%; height: 100%; }

  .activity-labels { display: flex; justify-content: space-between; }

  .activity-label-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .activity-day { font-size: 10px; color: var(--color-text-muted); text-transform: uppercase; }
  .activity-count { font-size: 11px; font-weight: 600; color: var(--color-primary); }

  .activity-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    gap: 12px;
    color: var(--color-text-muted);
  }

  .activity-empty p { font-size: 13px; margin: 0; }

  /* Session Stats */
  .session-stats {
    padding: 24px;
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 16px;
  }

  .session-stats h3 { margin: 0; font-size: 15px; font-weight: 600; }

  .session-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
  }

  .session-stat-card {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 16px;
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    transition: all 0.2s ease;
  }

  .session-stat-card:hover { border-color: var(--color-text-muted); transform: translateY(-1px); }

  :global(.dark) .session-stat-card {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.08);
  }

  :global(.dark) .session-stat-card:hover {
    background: rgba(255, 255, 255, 0.06);
    border-color: rgba(255, 255, 255, 0.15);
  }

  .session-stat-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(59, 130, 246, 0.02));
    border-radius: 10px;
    color: #3b82f6;
    flex-shrink: 0;
  }

  .session-stat-content { display: flex; flex-direction: column; gap: 2px; }

  .session-value {
    font-size: 22px;
    font-weight: 700;
    font-family: 'SF Mono', 'Menlo', monospace;
    color: var(--color-text);
    line-height: 1;
  }

  .session-unit { font-size: 14px; font-weight: 500; opacity: 0.7; }
  .session-label { font-size: 12px; color: var(--color-text-muted); font-weight: 500; }

  /* Section header (used in comparison and activity sections) */
  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .section-header h3 { margin: 0; font-size: 15px; font-weight: 600; }
  .section-hint { font-size: 12px; color: var(--color-text-muted); }

  /* Loading */
  .loading {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 24px;
    color: var(--color-text-muted);
  }

  .spinner {
    width: 20px;
    height: 20px;
    border: 2px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  /* Empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 48px;
    text-align: center;
    color: var(--color-text-muted);
  }

  .empty-icon {
    width: 64px;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-bg-tertiary);
    border-radius: 16px;
  }

  .empty-icon svg { width: 32px; height: 32px; }
  .empty-state h3 { margin: 0; font-size: 16px; }
  .empty-state p { margin: 0; font-size: 14px; }

  /* Responsive */
  @media (max-width: 1200px) {
    .key-metrics-grid { grid-template-columns: repeat(2, 1fr); }
    .breakdowns-grid { grid-template-columns: 1fr; }
    .session-grid { grid-template-columns: repeat(2, 1fr); }
  }

  @media (max-width: 768px) {
    .stats-system-row { grid-template-columns: 1fr; }
    .comparison-grid { grid-template-columns: 1fr; }
    .key-metrics-grid { grid-template-columns: 1fr; }
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
