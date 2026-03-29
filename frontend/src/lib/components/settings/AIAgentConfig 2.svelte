<script lang="ts">
  import AgentTestSandbox from '$lib/components/settings/AgentTestSandbox.svelte';
  import type { AgentInfo, CustomAgent } from '$lib/stores/aiSettings';
  import { getCategoryIconPath, getCategoryLabel } from '$lib/stores/aiSettings';
  import { apiClient } from '$lib/api';

  interface Props {
    agents: AgentInfo[];
    loadingAgents: boolean;
    customAgents: CustomAgent[];
    loadingCustomAgents: boolean;
    onAgentUpdated: () => void;
  }

  let {
    agents,
    loadingAgents,
    customAgents,
    loadingCustomAgents,
    onAgentUpdated
  }: Props = $props();

  // Custom agents UI state
  let showNewCustomAgent = $state(false);
  let editingCustomAgent = $state<CustomAgent | null>(null);
  let savingCustomAgent = $state(false);
  let testingAgentId = $state<string | null>(null);
  let newCustomAgent = $state({
    name: '',
    display_name: '',
    description: '',
    system_prompt: '',
    model_preference: '',
    temperature: 0.7,
    category: 'custom'
  });

  // Built-in agents UI state
  let expandedAgent = $state<string | null>(null);
  let editingAgent = $state<string | null>(null);
  let editedPrompt = $state<string>('');

  function toggleAgentExpand(agentId: string) {
    expandedAgent = expandedAgent === agentId ? null : agentId;
  }

  function startEditingAgent(agentId: string, currentPrompt: string) {
    editingAgent = agentId;
    editedPrompt = currentPrompt;
  }

  function cancelEditing() {
    editingAgent = null;
    editedPrompt = '';
  }

  async function saveAgentPrompt(agentId: string) {
    try {
      const res = await apiClient.put(`/ai/agents/${agentId}/prompt`, { prompt: editedPrompt });
      if (res.ok) {
        editingAgent = null;
        editedPrompt = '';
        onAgentUpdated();
      }
    } catch {
      // silently fail
    }
  }

  async function createCustomAgent() {
    if (!newCustomAgent.name.trim() || !newCustomAgent.system_prompt.trim()) return;
    savingCustomAgent = true;
    try {
      const res = await apiClient.post('/ai/custom-agents', newCustomAgent);
      if (res.ok) {
        showNewCustomAgent = false;
        newCustomAgent = {
          name: '',
          display_name: '',
          description: '',
          system_prompt: '',
          model_preference: '',
          temperature: 0.7,
          category: 'custom'
        };
        onAgentUpdated();
      }
    } catch {
      // silently fail
    } finally {
      savingCustomAgent = false;
    }
  }

  async function deleteCustomAgent(agentId: string) {
    if (!confirm('Delete this custom agent?')) return;
    try {
      const res = await apiClient.delete(`/ai/custom-agents/${agentId}`);
      if (res.ok) {
        onAgentUpdated();
      }
    } catch {
      // silently fail
    }
  }

  async function saveEditedCustomAgent() {
    if (!editingCustomAgent) return;
    savingCustomAgent = true;
    try {
      const res = await apiClient.put(`/ai/custom-agents/${editingCustomAgent.id}`, editingCustomAgent);
      if (res.ok) {
        editingCustomAgent = null;
        onAgentUpdated();
      }
    } catch {
      // silently fail
    } finally {
      savingCustomAgent = false;
    }
  }
</script>

<!-- Custom Agents Section -->
<section class="section">
  <div class="section-header">
    <h2>Custom Agents</h2>
    <button class="add-btn" onclick={() => showNewCustomAgent = true}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
      New Agent
    </button>
  </div>
  <p class="section-subtitle">Create custom agents you can mention with @name in chat</p>

  {#if showNewCustomAgent}
    <div class="custom-agent-form">
      <div class="form-header">
        <h3>Create Custom Agent</h3>
        <button class="close-btn" onclick={() => showNewCustomAgent = false}>x</button>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label>Agent Name (for @mentions)</label>
          <div class="input-prefix">
            <span>@</span>
            <input type="text" bind:value={newCustomAgent.name} placeholder="coder" />
          </div>
          <span class="form-hint">Lowercase, no spaces (use hyphens)</span>
        </div>
        <div class="form-group">
          <label>Display Name</label>
          <input type="text" bind:value={newCustomAgent.display_name} placeholder="My Coding Assistant" />
        </div>
        <div class="form-group full-width">
          <label>Description</label>
          <input type="text" bind:value={newCustomAgent.description} placeholder="A helpful coding assistant..." />
        </div>
        <div class="form-group full-width">
          <label>System Prompt</label>
          <textarea bind:value={newCustomAgent.system_prompt} rows="8" placeholder="You are an expert programmer..."></textarea>
          <span class="form-hint">This prompt defines the agent's personality and behavior</span>
        </div>
        <div class="form-group">
          <label>Model (optional)</label>
          <select bind:value={newCustomAgent.model_preference}>
            <option value="">Use default</option>
            <option value="llama-3.3-70b-versatile">Llama 3.3 70B</option>
            <option value="llama-3.1-8b-instant">Llama 3.1 8B</option>
            <option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
          </select>
        </div>
        <div class="form-group">
          <label>Temperature</label>
          <input type="number" bind:value={newCustomAgent.temperature} min="0" max="2" step="0.1" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn secondary" onclick={() => showNewCustomAgent = false}>Cancel</button>
        <button class="btn primary" onclick={createCustomAgent} disabled={savingCustomAgent}>
          {savingCustomAgent ? 'Creating...' : 'Create Agent'}
        </button>
      </div>
    </div>
  {/if}

  {#if loadingCustomAgents}
    <div class="loading">
      <div class="spinner"></div>
      <span>Loading custom agents...</span>
    </div>
  {:else if customAgents.length === 0}
    <div class="empty-state small">
      <p>No custom agents yet. Create one to mention with @name in chat!</p>
    </div>
  {:else}
    <div class="custom-agents-list">
      {#each customAgents as agent}
        <div class="custom-agent-card">
          <div class="agent-header">
            <div class="agent-info">
              <span class="agent-mention">@{agent.name}</span>
              <span class="agent-display-name">{agent.display_name}</span>
            </div>
            <div class="agent-actions">
              <span class="usage-badge">{agent.times_used} uses</span>
              <button class="icon-btn" onclick={() => testingAgentId = testingAgentId === agent.id ? null : agent.id} title="Test Agent">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"/></svg>
              </button>
              <button class="icon-btn" onclick={() => editingCustomAgent = {...agent}} title="Edit">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button class="icon-btn danger" onclick={() => deleteCustomAgent(agent.id)} title="Delete">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
          {#if agent.description}
            <p class="agent-desc">{agent.description}</p>
          {/if}
          <div class="agent-prompt-preview">
            <span class="label">Prompt:</span>
            <span class="preview">{agent.system_prompt.slice(0, 100)}...</span>
          </div>

          {#if testingAgentId === agent.id}
            <div class="agent-test-section" style="margin-top: 1rem; padding-top: 1rem; border-top: 1px solid #e5e7eb;">
              <h4 style="font-size: 0.875rem; font-weight: 600; color: #374151; margin-bottom: 0.75rem;">Test Agent</h4>
              <AgentTestSandbox {agent} />
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</section>

<!-- Edit Custom Agent Modal -->
{#if editingCustomAgent}
  <div class="modal-overlay" onclick={() => editingCustomAgent = null}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="form-header">
        <h3>Edit Agent</h3>
        <button class="close-btn" onclick={() => editingCustomAgent = null}>×</button>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label>Agent Name</label>
          <div class="input-prefix">
            <span>@</span>
            <input type="text" bind:value={editingCustomAgent.name} placeholder="coder" />
          </div>
        </div>
        <div class="form-group">
          <label>Display Name</label>
          <input type="text" bind:value={editingCustomAgent.display_name} placeholder="My Coding Assistant" />
        </div>
        <div class="form-group full-width">
          <label>Description</label>
          <input type="text" bind:value={editingCustomAgent.description} placeholder="A helpful coding assistant..." />
        </div>
        <div class="form-group full-width">
          <label>System Prompt</label>
          <textarea bind:value={editingCustomAgent.system_prompt} rows="8" placeholder="You are an expert programmer..."></textarea>
        </div>
      </div>
      <div class="form-actions">
        <button class="btn secondary" onclick={() => editingCustomAgent = null}>Cancel</button>
        <button class="btn primary" onclick={saveEditedCustomAgent} disabled={savingCustomAgent}>
          {savingCustomAgent ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Built-in Agents Section -->
<section class="section">
  <div class="section-header">
    <h2>Built-in Agents</h2>
    <span class="subtitle">View and customize agent prompts</span>
  </div>

  {#if loadingAgents}
    <div class="loading">
      <div class="spinner"></div>
      <span>Loading agents...</span>
    </div>
  {:else if agents.length === 0}
    <div class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><circle cx="12" cy="5" r="2"/><path d="M12 7v4"/><path d="M7 22v-3"/><path d="M17 22v-3"/></svg>
      </div>
      <h3>No Agents Found</h3>
      <p>Agent configuration is not available</p>
    </div>
  {:else}
    <div class="agents-list">
      {#each agents as agent}
        <div class="agent-card" class:expanded={expandedAgent === agent.id}>
          <button class="agent-header-btn" onclick={() => toggleAgentExpand(agent.id)}>
            <div class="agent-info">
              <span class="agent-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={getCategoryIconPath(agent.category)}/></svg>
              </span>
              <div class="agent-text">
                <span class="agent-name">{agent.name}</span>
                <span class="agent-desc">{agent.description}</span>
              </div>
            </div>
            <div class="agent-meta">
              <span class="agent-category">{getCategoryLabel(agent.category)}</span>
              <svg class="agent-chevron" class:rotated={expandedAgent === agent.id} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="6 9 12 15 18 9"/>
              </svg>
            </div>
          </button>

          {#if expandedAgent === agent.id}
            <div class="agent-content">
              <div class="prompt-header">
                <h4>System Prompt</h4>
                {#if editingAgent !== agent.id}
                  <button class="prompt-edit-btn" onclick={() => startEditingAgent(agent.id, agent.prompt)}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    Edit
                  </button>
                {/if}
              </div>

              {#if editingAgent === agent.id}
                <div class="prompt-editor">
                  <textarea
                    bind:value={editedPrompt}
                    rows="12"
                    placeholder="Enter system prompt..."
                  ></textarea>
                  <div class="editor-actions">
                    <button class="editor-btn cancel" onclick={cancelEditing}>Cancel</button>
                    <button class="editor-btn save" onclick={() => saveAgentPrompt(agent.id)}>Save Changes</button>
                  </div>
                </div>
              {:else}
                <div class="prompt-display">
                  <pre>{agent.prompt}</pre>
                </div>
              {/if}

              <div class="agent-stats">
                <div class="stat-item">
                  <span class="stat-label">ID</span>
                  <span class="stat-value-small">{agent.id}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Characters</span>
                  <span class="stat-value-small">{agent.prompt.length.toLocaleString()}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">~Tokens</span>
                  <span class="stat-value-small">{Math.ceil(agent.prompt.length / 4).toLocaleString()}</span>
                </div>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</section>

<style>
  .section {
    margin-bottom: 32px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-header h2 {
    font-size: 16px;
    font-weight: 600;
    margin: 0;
  }

  .section-subtitle {
    color: var(--color-text-muted);
    font-size: 14px;
    margin: -8px 0 16px 0;
  }

  .subtitle {
    font-size: 13px;
    color: var(--color-text-muted);
  }

  /* Add Button */
  .add-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .add-btn:hover { opacity: 0.9; }
  .add-btn svg { width: 16px; height: 16px; }

  /* Custom Agents */
  .custom-agent-form {
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 20px;
  }

  .custom-agents-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .custom-agent-card {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    padding: 16px;
  }

  .custom-agent-card:hover {
    border-color: var(--color-primary);
  }

  .custom-agent-card .agent-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .custom-agent-card .agent-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .agent-mention {
    font-family: monospace;
    font-weight: 600;
    color: var(--color-primary);
    background: var(--color-primary-bg);
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 14px;
  }

  .agent-display-name {
    font-weight: 500;
    color: var(--color-text);
  }

  .custom-agent-card .agent-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .usage-badge {
    font-size: 12px;
    color: var(--color-text-muted);
    background: var(--color-bg-elevated);
    padding: 4px 8px;
    border-radius: 4px;
  }

  .custom-agent-card .agent-desc {
    font-size: 14px;
    color: var(--color-text-muted);
    margin-bottom: 8px;
  }

  .agent-prompt-preview {
    font-size: 13px;
    display: flex;
    gap: 8px;
  }

  .agent-prompt-preview .label {
    color: var(--color-text-muted);
    flex-shrink: 0;
  }

  .agent-prompt-preview .preview {
    color: var(--color-text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty-state.small {
    padding: 24px;
    text-align: center;
    color: var(--color-text-muted);
    background: var(--color-bg-elevated);
    border-radius: 8px;
  }

  /* Icon buttons */
  .icon-btn {
    width: 28px;
    height: 28px;
    border: none;
    background: var(--color-bg-tertiary);
    color: var(--color-text-secondary);
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .icon-btn svg {
    width: 14px;
    height: 14px;
  }

  .icon-btn:hover {
    background: var(--color-bg);
    color: var(--color-text);
  }

  .icon-btn.danger:hover {
    color: var(--color-error);
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
  }

  /* Built-in Agents */
  .agents-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .agent-card {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    overflow: hidden;
    transition: all 0.2s;
  }

  .agent-card:hover {
    border-color: var(--color-border-hover);
  }

  .agent-card.expanded {
    border-color: var(--color-primary);
  }

  .agent-header-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
  }

  .agent-info {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .agent-icon {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-bg-secondary);
    border-radius: 12px;
    flex-shrink: 0;
  }

  .agent-icon svg {
    width: 24px;
    height: 24px;
    color: var(--color-text-muted);
  }

  :global(.dark) .agent-icon {
    background: rgba(255, 255, 255, 0.05);
  }

  :global(:not(.dark)) .agent-icon {
    background: rgba(0, 0, 0, 0.03);
  }

  .agent-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .agent-name {
    font-weight: 600;
    font-size: 15px;
    color: var(--color-text);
  }

  .agent-desc {
    font-size: 13px;
    color: var(--color-text-muted);
  }

  .agent-meta {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .agent-category {
    font-size: 11px;
    padding: 4px 10px;
    background: var(--color-bg-tertiary);
    border-radius: 6px;
    color: var(--color-text-secondary);
    text-transform: uppercase;
    font-weight: 500;
  }

  .agent-chevron {
    width: 20px;
    height: 20px;
    color: var(--color-text-muted);
    transition: transform 0.2s;
  }

  .agent-chevron.rotated {
    transform: rotate(180deg);
  }

  .agent-content {
    padding: 0 20px 20px;
    border-top: 1px solid var(--color-border);
    background: var(--color-bg-secondary);
  }

  .prompt-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0 12px;
  }

  .prompt-header h4 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }

  .prompt-edit-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    font-size: 12px;
    color: var(--color-text-secondary);
    cursor: pointer;
    transition: all 0.2s;
  }

  .prompt-edit-btn:hover {
    background: var(--color-bg-tertiary);
    color: var(--color-text);
  }

  .prompt-edit-btn svg {
    width: 14px;
    height: 14px;
  }

  .prompt-display {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    max-height: 400px;
    overflow-y: auto;
  }

  .prompt-display pre {
    margin: 0;
    padding: 16px;
    font-size: 12px;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--color-text-secondary);
  }

  .prompt-editor {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .prompt-editor textarea {
    width: 100%;
    padding: 16px;
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 10px;
    font-size: 12px;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    line-height: 1.6;
    color: var(--color-text);
    resize: vertical;
  }

  .prompt-editor textarea:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .editor-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .editor-btn {
    padding: 8px 16px;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .editor-btn.cancel {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    color: var(--color-text-secondary);
  }

  .editor-btn.cancel:hover {
    background: var(--color-bg-tertiary);
  }

  .editor-btn.save {
    background: var(--color-primary);
    border: none;
    color: var(--color-bg);
  }

  .editor-btn.save:hover {
    opacity: 0.9;
  }

  .agent-stats {
    display: flex;
    gap: 24px;
    padding-top: 16px;
    margin-top: 16px;
    border-top: 1px solid var(--color-border);
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .stat-item .stat-label {
    font-size: 10px;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .stat-value-small {
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-secondary);
    font-family: monospace;
  }

  /* Form styles */
  .form-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .form-header h3 {
    font-size: 16px;
    font-weight: 600;
    margin: 0;
  }

  .close-btn {
    width: 28px;
    height: 28px;
    border: none;
    background: var(--color-bg-tertiary);
    color: var(--color-text-secondary);
    border-radius: 6px;
    font-size: 18px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: var(--color-bg);
    color: var(--color-text);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-group.full-width {
    grid-column: span 2;
  }

  .form-group label {
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-secondary);
  }

  .form-group input,
  .form-group textarea,
  .form-group select {
    padding: 10px 12px;
    border: 1px solid var(--color-border);
    border-radius: 8px;
    background: var(--color-bg);
    color: var(--color-text);
    font-size: 14px;
    transition: border-color 0.2s;
  }

  .form-group input:focus,
  .form-group textarea:focus,
  .form-group select:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .form-group textarea {
    resize: vertical;
    font-family: monospace;
    font-size: 13px;
  }

  .input-prefix {
    display: flex;
    align-items: center;
    border: 1px solid var(--color-border);
    border-radius: 8px;
    background: var(--color-bg);
    overflow: hidden;
  }

  .input-prefix span {
    padding: 10px 4px 10px 12px;
    color: var(--color-text-secondary);
    font-family: monospace;
  }

  .input-prefix input {
    border: none;
    padding-left: 0;
    flex: 1;
  }

  .form-hint {
    font-size: 11px;
    color: var(--color-text-tertiary);
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid var(--color-border);
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn.primary {
    background: var(--color-primary);
    color: white;
  }

  .btn.primary:hover { opacity: 0.9; }
  .btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn.secondary {
    background: var(--color-bg-tertiary);
    color: var(--color-text);
  }

  .btn.secondary:hover {
    background: var(--color-bg);
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    padding: 20px;
    max-width: 600px;
    width: 90%;
    max-height: 85vh;
    overflow-y: auto;
  }

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

  @keyframes spin {
    to { transform: rotate(360deg); }
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

  .empty-icon svg {
    width: 32px;
    height: 32px;
  }

  .empty-state h3 {
    margin: 0;
    font-size: 16px;
  }

  .empty-state p {
    margin: 0;
    font-size: 14px;
  }

  /* Dark mode */
  :global(.dark) .agent-card {
    background: #2c2c2e;
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .agent-card:hover {
    border-color: rgba(255, 255, 255, 0.2);
  }

  :global(.dark) .agent-card.expanded {
    border-color: #0A84FF;
    background: rgba(10, 132, 255, 0.1);
  }

  :global(.dark) .agent-content {
    background: #1c1c1e;
  }

  :global(.dark) .prompt-display {
    background: #2c2c2e;
  }

  :global(.dark) .prompt-editor textarea {
    background: #2c2c2e;
  }

  :global(.dark) .add-btn,
  :global(.dark) .editor-btn.save,
  :global(.dark) .btn.primary {
    background: #0A84FF;
    color: white;
  }

  :global(.dark) .agent-icon {
    background: #3a3a3c;
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .icon-btn {
    background: #3a3a3c;
  }

  :global(.dark) .close-btn {
    background: #3a3a3c;
  }

  :global(.dark) .btn.secondary {
    background: #3a3a3c;
  }

  :global(.dark) .form-group input,
  :global(.dark) .form-group textarea,
  :global(.dark) .form-group select {
    background: #1c1c1e;
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .input-prefix {
    background: #1c1c1e;
    border-color: rgba(255, 255, 255, 0.1);
  }

  :global(.dark) .modal {
    background: #2c2c2e;
    border-color: rgba(255, 255, 255, 0.1);
  }
</style>
