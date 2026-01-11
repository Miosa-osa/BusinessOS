<script lang="ts">
  import { onMount } from 'svelte';
  import { thinking } from '$lib/stores/thinking';
  import type { ReasoningTemplate, ReasoningStep, StepType } from '$lib/api/thinking/types';
  import {
    Brain,
    Plus,
    Edit,
    Trash2,
    Star,
    Loader2,
    AlertCircle,
    CheckCircle,
    X
  } from 'lucide-svelte';

  // State
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let templates = $state<ReasoningTemplate[]>([]);
  let showCreateModal = $state(false);
  let showEditModal = $state(false);
  let editingTemplate = $state<ReasoningTemplate | null>(null);
  let deletingId = $state<string | null>(null);

  // Form state
  let formName = $state('');
  let formDescription = $state('');
  let formSteps = $state<ReasoningStep[]>([
    { order: 0, type: 'exploration', prompt: '' }
  ]);

  onMount(async () => {
    await loadTemplates();
  });

  async function loadTemplates() {
    isLoading = true;
    error = null;
    try {
      await thinking.loadTemplates();
      thinking.subscribe(state => {
        templates = state.templates;
      });
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load templates';
    } finally {
      isLoading = false;
    }
  }

  function openCreateModal() {
    formName = '';
    formDescription = '';
    formSteps = [{ order: 0, type: 'exploration', prompt: '' }];
    showCreateModal = true;
  }

  function openEditModal(template: ReasoningTemplate) {
    editingTemplate = template;
    formName = template.name;
    formDescription = template.description || '';
    formSteps = [...template.steps];
    showEditModal = true;
  }

  function closeModals() {
    showCreateModal = false;
    showEditModal = false;
    editingTemplate = null;
  }

  function addStep() {
    formSteps = [...formSteps, { order: formSteps.length, type: 'exploration' as StepType, prompt: '' }];
  }

  function removeStep(index: number) {
    formSteps = formSteps.filter((_, i) => i !== index);
  }

  async function handleCreate() {
    if (!formName.trim() || formSteps.length === 0) {
      error = 'Please provide a name and at least one step';
      return;
    }

    try {
      await thinking.createTemplate({
        name: formName,
        description: formDescription,
        steps: formSteps
      });
      closeModals();
      await loadTemplates();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create template';
    }
  }

  async function handleUpdate() {
    if (!editingTemplate || !formName.trim() || formSteps.length === 0) {
      error = 'Please provide a name and at least one step';
      return;
    }

    try {
      await thinking.updateTemplate(editingTemplate.id, {
        name: formName,
        description: formDescription,
        steps: formSteps
      });
      closeModals();
      await loadTemplates();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update template';
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this template?')) {
      return;
    }

    deletingId = id;
    try {
      await thinking.deleteTemplate(id);
      await loadTemplates();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete template';
    } finally {
      deletingId = null;
    }
  }

  async function handleSetDefault(id: string) {
    try {
      await thinking.setDefaultTemplate(id);
      await loadTemplates();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to set default template';
    }
  }

  function formatDate(dateString: string) {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<div class="templates-page">
  <!-- Header -->
  <div class="page-header">
    <div class="header-content">
      <div class="header-title">
        <Brain class="w-6 h-6" />
        <h1>Reasoning Templates</h1>
      </div>
      <button class="btn-primary" onclick={openCreateModal}>
        <Plus class="w-5 h-5" />
        <span>Create Template</span>
      </button>
    </div>
    {#if error}
      <div class="error-banner">
        <AlertCircle class="w-5 h-5" />
        <p>{error}</p>
        <button onclick={() => error = null}>
          <X class="w-4 h-4" />
        </button>
      </div>
    {/if}
  </div>

  <!-- Content -->
  <div class="page-content">
    {#if isLoading}
      <div class="loading-state">
        <Loader2 class="w-8 h-8 animate-spin" />
        <p>Loading templates...</p>
      </div>
    {:else if templates.length === 0}
      <div class="empty-state">
        <Brain class="w-16 h-16 opacity-20" />
        <h2>No Templates Yet</h2>
        <p>Create your first reasoning template to guide AI thinking process</p>
        <button class="btn-primary" onclick={openCreateModal}>
          <Plus class="w-5 h-5" />
          <span>Create Template</span>
        </button>
      </div>
    {:else}
      <div class="templates-grid">
        {#each templates as template}
          <div class="template-card">
            <div class="card-header">
              <div class="card-title">
                <h3>{template.name}</h3>
                {#if template.is_default}
                  <span class="default-badge">
                    <Star class="w-3 h-3 fill-current" />
                    Default
                  </span>
                {/if}
              </div>
              <div class="card-actions">
                <button
                  class="btn-icon"
                  onclick={() => openEditModal(template)}
                  title="Edit template"
                >
                  <Edit class="w-4 h-4" />
                </button>
                <button
                  class="btn-icon danger"
                  onclick={() => handleDelete(template.id)}
                  disabled={deletingId === template.id}
                  title="Delete template"
                >
                  {#if deletingId === template.id}
                    <Loader2 class="w-4 h-4 animate-spin" />
                  {:else}
                    <Trash2 class="w-4 h-4" />
                  {/if}
                </button>
              </div>
            </div>

            <p class="card-description">{template.description}</p>

            <div class="card-meta">
              <div class="meta-item">
                <span class="meta-label">Steps</span>
                <span class="meta-value">{template.steps.length}</span>
              </div>
              <div class="meta-item">
                <span class="meta-label">Used</span>
                <span class="meta-value">{template.times_used || 0} times</span>
              </div>
              <div class="meta-item">
                <span class="meta-label">Created</span>
                <span class="meta-value">{formatDate(template.created_at)}</span>
              </div>
            </div>

            {#if !template.is_default}
              <button
                class="btn-secondary full-width"
                onclick={() => handleSetDefault(template.id)}
              >
                <Star class="w-4 h-4" />
                <span>Set as Default</span>
              </button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Modal -->
{#if showCreateModal}
  <div class="modal-overlay" onclick={closeModals}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>Create Reasoning Template</h2>
        <button class="btn-icon" onclick={closeModals}>
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="modal-content">
        <div class="form-group">
          <label for="template-name">Name</label>
          <input
            id="template-name"
            type="text"
            bind:value={formName}
            placeholder="e.g., Analytical Thinking"
          />
        </div>

        <div class="form-group">
          <label for="template-description">Description</label>
          <textarea
            id="template-description"
            bind:value={formDescription}
            placeholder="Describe when to use this template..."
            rows="3"
          ></textarea>
        </div>

        <div class="form-group">
          <div class="steps-header">
            <label>Steps</label>
            <button class="btn-sm" onclick={addStep}>
              <Plus class="w-4 h-4" />
              <span>Add Step</span>
            </button>
          </div>

          <div class="steps-list">
            {#each formSteps as step, i}
              <div class="step-item">
                <div class="step-number">{i + 1}</div>
                <div class="step-fields">
                  <select
                    bind:value={step.type}
                    class="step-type-select"
                  >
                    <option value="exploration">Exploration</option>
                    <option value="analysis">Analysis</option>
                    <option value="conclusion">Conclusion</option>
                    <option value="reflection">Reflection</option>
                  </select>
                  <textarea
                    bind:value={step.prompt}
                    placeholder="Step prompt or instruction..."
                    rows="2"
                  ></textarea>
                </div>
                {#if formSteps.length > 1}
                  <button
                    class="btn-icon danger"
                    onclick={() => removeStep(i)}
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" onclick={closeModals}>
          Cancel
        </button>
        <button class="btn-primary" onclick={handleCreate}>
          <CheckCircle class="w-4 h-4" />
          <span>Create Template</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Edit Modal -->
{#if showEditModal}
  <div class="modal-overlay" onclick={closeModals}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>Edit Reasoning Template</h2>
        <button class="btn-icon" onclick={closeModals}>
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="modal-content">
        <div class="form-group">
          <label for="edit-template-name">Name</label>
          <input
            id="edit-template-name"
            type="text"
            bind:value={formName}
            placeholder="e.g., Analytical Thinking"
          />
        </div>

        <div class="form-group">
          <label for="edit-template-description">Description</label>
          <textarea
            id="edit-template-description"
            bind:value={formDescription}
            placeholder="Describe when to use this template..."
            rows="3"
          ></textarea>
        </div>

        <div class="form-group">
          <div class="steps-header">
            <label>Steps</label>
            <button class="btn-sm" onclick={addStep}>
              <Plus class="w-4 h-4" />
              <span>Add Step</span>
            </button>
          </div>

          <div class="steps-list">
            {#each formSteps as step, i}
              <div class="step-item">
                <div class="step-number">{i + 1}</div>
                <div class="step-fields">
                  <select
                    bind:value={step.type}
                    class="step-type-select"
                  >
                    <option value="exploration">Exploration</option>
                    <option value="analysis">Analysis</option>
                    <option value="conclusion">Conclusion</option>
                    <option value="reflection">Reflection</option>
                  </select>
                  <textarea
                    bind:value={step.prompt}
                    placeholder="Step prompt or instruction..."
                    rows="2"
                  ></textarea>
                </div>
                {#if formSteps.length > 1}
                  <button
                    class="btn-icon danger"
                    onclick={() => removeStep(i)}
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" onclick={closeModals}>
          Cancel
        </button>
        <button class="btn-primary" onclick={handleUpdate}>
          <CheckCircle class="w-4 h-4" />
          <span>Save Changes</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .templates-page {
    min-height: 100vh;
    background: #f9fafb;
  }

  .page-header {
    background: white;
    border-bottom: 1px solid #e5e7eb;
    padding: 1.5rem 2rem;
  }

  .header-content {
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .header-title h1 {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .error-banner {
    max-width: 1400px;
    margin: 1rem auto 0;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 0.5rem;
    color: #dc2626;
  }

  .error-banner button {
    margin-left: auto;
    padding: 0.25rem;
    background: transparent;
    border: none;
    cursor: pointer;
    color: #dc2626;
  }

  .page-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
  }

  .loading-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
    text-align: center;
  }

  .empty-state h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .empty-state p {
    color: #6b7280;
    margin: 0 0 1rem;
  }

  .templates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
  }

  .template-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    transition: all 0.15s;
  }

  .template-card:hover {
    border-color: #d1d5db;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .card-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .card-title h3 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .default-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    background: #fef3c7;
    color: #92400e;
    font-size: 0.75rem;
    font-weight: 600;
    border-radius: 9999px;
  }

  .card-actions {
    display: flex;
    gap: 0.25rem;
  }

  .card-description {
    color: #6b7280;
    font-size: 0.875rem;
    line-height: 1.5;
    margin: 0;
  }

  .card-meta {
    display: flex;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .meta-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #9ca3af;
    text-transform: uppercase;
  }

  .meta-value {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
  }

  /* Buttons */
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    background: #2563eb;
    color: white;
    border: none;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-primary:hover {
    background: #1d4ed8;
  }

  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-secondary:hover {
    background: #f9fafb;
    border-color: #9ca3af;
  }

  .btn-sm {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.8125rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-sm:hover {
    background: #f9fafb;
  }

  .btn-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    color: #6b7280;
    transition: all 0.15s;
  }

  .btn-icon:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .btn-icon.danger {
    color: #dc2626;
  }

  .btn-icon.danger:hover {
    background: #fef2f2;
  }

  .btn-icon:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .full-width {
    width: 100%;
    justify-content: center;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal {
    background: white;
    border-radius: 0.5rem;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .modal-content {
    flex: 1;
    overflow-y: auto;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.75rem;
    padding: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  /* Forms */
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-group label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .form-group input[type="text"],
  .form-group textarea {
    padding: 0.625rem 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    color: #111827;
    transition: all 0.15s;
  }

  .form-group input[type="text"]:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #2563eb;
    ring: 2px;
    ring-color: #bfdbfe;
  }

  .steps-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .steps-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .step-item {
    display: flex;
    gap: 0.75rem;
    padding: 1rem;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
  }

  .step-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    background: white;
    border: 2px solid #2563eb;
    color: #2563eb;
    font-weight: 600;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .step-fields {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .step-fields input,
  .step-fields textarea,
  .step-fields select {
    padding: 0.5rem 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    background: white;
  }

  .step-type-select {
    font-weight: 500;
    color: #374151;
    cursor: pointer;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
    color: #6b7280;
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    width: 1rem;
    height: 1rem;
    cursor: pointer;
  }

  /* Dark mode */
  :global(.dark) .templates-page {
    background: #111827;
  }

  :global(.dark) .page-header {
    background: #1f2937;
    border-bottom-color: #374151;
  }

  :global(.dark) .header-title h1 {
    color: #f9fafb;
  }

  :global(.dark) .template-card,
  :global(.dark) .modal {
    background: #1f2937;
    border-color: #374151;
  }

  :global(.dark) .card-title h3,
  :global(.dark) .meta-value,
  :global(.dark) .modal-header h2 {
    color: #f9fafb;
  }

  :global(.dark) .card-description {
    color: #9ca3af;
  }

  :global(.dark) .card-meta {
    border-top-color: #374151;
  }

  :global(.dark) .btn-secondary,
  :global(.dark) .btn-sm {
    background: #374151;
    color: #d1d5db;
    border-color: #4b5563;
  }

  :global(.dark) .btn-secondary:hover,
  :global(.dark) .btn-sm:hover {
    background: #4b5563;
  }

  :global(.dark) .btn-icon:hover {
    background: #374151;
    color: #f9fafb;
  }

  :global(.dark) .modal-header,
  :global(.dark) .modal-footer {
    border-color: #374151;
  }

  :global(.dark) .form-group input,
  :global(.dark) .form-group textarea,
  :global(.dark) .step-fields input,
  :global(.dark) .step-fields textarea,
  :global(.dark) .step-fields select {
    background: #111827;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .step-item {
    background: #111827;
    border-color: #374151;
  }

  :global(.dark) .step-number {
    background: #1f2937;
  }
</style>
