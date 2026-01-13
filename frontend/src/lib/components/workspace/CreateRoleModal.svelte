<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { createWorkspaceRole, type WorkspaceRole } from '$lib/api/workspaces';
  import { X, Shield, Save, Loader2, AlertCircle, Palette, ChevronDown, ChevronUp } from 'lucide-svelte';
  import PermissionsMatrixEditor from './PermissionsMatrixEditor.svelte';

  interface Props {
    workspaceId: string;
    existingRoles: WorkspaceRole[];
  }

  let { workspaceId, existingRoles }: Props = $props();

  const dispatch = createEventDispatcher();

  // Form state
  let name = $state('');
  let displayName = $state('');
  let description = $state('');
  let color = $state('#6366f1');
  let permissions = $state<Record<string, Record<string, boolean | string>>>({});
  
  // UI state
  let isSaving = $state(false);
  let error = $state<string | null>(null);
  let showAdvanced = $state(false);

  // Predefined colors
  const colorOptions = [
    { value: '#8b5cf6', label: 'Purple' },
    { value: '#3b82f6', label: 'Blue' },
    { value: '#10b981', label: 'Green' },
    { value: '#f59e0b', label: 'Amber' },
    { value: '#ef4444', label: 'Red' },
    { value: '#ec4899', label: 'Pink' },
    { value: '#06b6d4', label: 'Cyan' },
    { value: '#6366f1', label: 'Indigo' },
  ];

  // Default permission templates
  const roleTemplates = [
    {
      id: 'custom',
      label: 'Start from scratch',
      description: 'Create a role with no permissions',
      permissions: {},
    },
    {
      id: 'viewer',
      label: 'Viewer-like',
      description: 'Read-only access to most resources',
      permissions: {
        projects: { create: false, read: true, update: false, delete: false, manage_members: false },
        tasks: { create: false, read: true, update: false, delete: false, assign: false },
        contexts: { create: false, read: true, update: false, delete: false },
        clients: { create: false, read: true, update: false, delete: false },
        artifacts: { create: false, read: true, update: false, delete: false },
        members: { view: true, invite: false, manage: false },
        roles: { view: true, manage: false },
        workspace: { view: false, manage: false },
        agent: { use_all_agents: true, create_custom_agents: false, access_workspace_memory: true, modify_workspace_memory: false },
      },
    },
    {
      id: 'contributor',
      label: 'Contributor-like',
      description: 'Can create and edit content',
      permissions: {
        projects: { create: true, read: true, update: true, delete: false, manage_members: false },
        tasks: { create: true, read: true, update: true, delete: true, assign: false },
        contexts: { create: true, read: true, update: true, delete: false },
        clients: { create: false, read: true, update: false, delete: false },
        artifacts: { create: true, read: true, update: true, delete: true },
        members: { view: true, invite: false, manage: false },
        roles: { view: true, manage: false },
        workspace: { view: false, manage: false },
        agent: { use_all_agents: true, create_custom_agents: false, access_workspace_memory: true, modify_workspace_memory: false },
      },
    },
    {
      id: 'manager',
      label: 'Manager-like',
      description: 'Full content access plus team management',
      permissions: {
        projects: { create: true, read: true, update: true, delete: true, manage_members: true },
        tasks: { create: true, read: true, update: true, delete: true, assign: true },
        contexts: { create: true, read: true, update: true, delete: true },
        clients: { create: true, read: true, update: true, delete: false },
        artifacts: { create: true, read: true, update: true, delete: true },
        members: { view: true, invite: true, manage: false },
        roles: { view: true, manage: false },
        workspace: { view: true, manage: false },
        agent: { use_all_agents: true, create_custom_agents: true, access_workspace_memory: true, modify_workspace_memory: true },
      },
    },
  ];

  let selectedTemplate = $state('custom');

  // Auto-generate name from display name
  $effect(() => {
    if (displayName && !name) {
      name = displayName.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
    }
  });

  // Apply template when selected
  function applyTemplate(templateId: string) {
    selectedTemplate = templateId;
    const template = roleTemplates.find(t => t.id === templateId);
    if (template) {
      permissions = JSON.parse(JSON.stringify(template.permissions));
    }
  }

  // Validation
  const isValid = $derived(() => {
    if (!name.trim()) return false;
    if (!displayName.trim()) return false;
    if (!/^[a-z][a-z0-9_]*$/.test(name)) return false;
    if (existingRoles.some(r => r.name === name)) return false;
    return true;
  });

  const validationError = $derived(() => {
    if (!name.trim()) return 'Name is required';
    if (!/^[a-z][a-z0-9_]*$/.test(name)) return 'Name must start with a letter and contain only lowercase letters, numbers, and underscores';
    if (existingRoles.some(r => r.name === name)) return 'A role with this name already exists';
    if (!displayName.trim()) return 'Display name is required';
    return null;
  });

  async function handleSave() {
    if (!isValid()) return;

    try {
      isSaving = true;
      error = null;

      await createWorkspaceRole(workspaceId, {
        name: name.trim(),
        display_name: displayName.trim(),
        description: description.trim() || undefined,
        color,
        permissions,
      });

      dispatch('success');
    } catch (err) {
      console.error('Failed to create role:', err);
      error = err instanceof Error ? err.message : 'Failed to create role';
    } finally {
      isSaving = false;
    }
  }

  function handleCancel() {
    dispatch('cancel');
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      handleCancel();
    }
  }

  function handlePermissionsChange(newPermissions: Record<string, Record<string, boolean | string>>) {
    permissions = newPermissions;
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-overlay" onclick={handleCancel}>
  <div class="modal-content" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div>
        <h2>Create Custom Role</h2>
        <p>Define a new role with specific permissions</p>
      </div>
      <button class="close-button" onclick={handleCancel} type="button">
        <X class="w-5 h-5" />
      </button>
    </div>

    {#if error}
      <div class="error-message">
        <AlertCircle class="w-4 h-4" />
        <span>{error}</span>
      </div>
    {/if}

    <div class="modal-body">
      <!-- Template Selection -->
      <div class="form-section">
        <label class="section-label">Start with a template</label>
        <div class="template-grid">
          {#each roleTemplates as template}
            <button
              type="button"
              class="template-card"
              class:selected={selectedTemplate === template.id}
              onclick={() => applyTemplate(template.id)}
            >
              <span class="template-name">{template.label}</span>
              <span class="template-desc">{template.description}</span>
            </button>
          {/each}
        </div>
      </div>

      <!-- Basic Info -->
      <div class="form-row">
        <div class="form-group">
          <label for="displayName">
            <Shield class="w-4 h-4" />
            Display Name *
          </label>
          <input
            id="displayName"
            type="text"
            bind:value={displayName}
            placeholder="e.g., Project Lead"
            disabled={isSaving}
            autofocus
          />
        </div>

        <div class="form-group">
          <label for="name">
            Internal Name *
          </label>
          <input
            id="name"
            type="text"
            bind:value={name}
            placeholder="e.g., project_lead"
            disabled={isSaving}
            class:error={name && !/^[a-z][a-z0-9_]*$/.test(name)}
          />
          <p class="field-hint">Lowercase letters, numbers, and underscores only</p>
        </div>
      </div>

      <div class="form-group">
        <label for="description">Description</label>
        <textarea
          id="description"
          bind:value={description}
          placeholder="What can this role do?"
          rows="2"
          disabled={isSaving}
        ></textarea>
      </div>

      <!-- Color Selection -->
      <div class="form-group">
        <label>
          <Palette class="w-4 h-4" />
          Role Color
        </label>
        <div class="color-options">
          {#each colorOptions as colorOpt}
            <button
              type="button"
              class="color-swatch"
              class:selected={color === colorOpt.value}
              style="background-color: {colorOpt.value}"
              onclick={() => color = colorOpt.value}
              title={colorOpt.label}
            >
              {#if color === colorOpt.value}
                <span class="color-check">✓</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Permissions Matrix -->
      <div class="form-section">
        <button
          type="button"
          class="section-toggle"
          onclick={() => showAdvanced = !showAdvanced}
        >
          <span class="section-label">Permissions</span>
          {#if showAdvanced}
            <ChevronUp class="w-4 h-4" />
          {:else}
            <ChevronDown class="w-4 h-4" />
          {/if}
        </button>
        
        {#if showAdvanced}
          <div class="permissions-wrapper">
            <PermissionsMatrixEditor
              bind:permissions
              onchange={handlePermissionsChange}
              compact
            />
          </div>
        {:else}
          <p class="permissions-hint">
            Click to customize specific permissions for this role
          </p>
        {/if}
      </div>

      {#if validationError()}
        <div class="validation-hint">
          <AlertCircle class="w-4 h-4" />
          {validationError()}
        </div>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="cancel-button" onclick={handleCancel} type="button" disabled={isSaving}>
        Cancel
      </button>
      <button
        class="save-button"
        onclick={handleSave}
        disabled={!isValid() || isSaving}
        type="button"
      >
        {#if isSaving}
          <Loader2 class="w-4 h-4 animate-spin" />
        {:else}
          <Save class="w-4 h-4" />
        {/if}
        Create Role
      </button>
    </div>
  </div>
</div>

<style>
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

  .modal-content {
    width: 100%;
    max-width: 700px;
    max-height: 90vh;
    background: white;
    border-radius: 0.75rem;
    box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  .modal-header h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.25rem 0;
  }

  .modal-header p {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

  .close-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    background: transparent;
    border: none;
    color: #9ca3af;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: all 0.15s;
  }

  .close-button:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: #fef2f2;
    color: #dc2626;
    font-size: 0.875rem;
  }

  .modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .form-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .section-label {
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
  }

  .section-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.75rem;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.15s;
  }

  .section-toggle:hover {
    background: #f3f4f6;
  }

  .template-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .template-card {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem;
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 0.5rem;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s;
  }

  .template-card:hover {
    border-color: #d1d5db;
  }

  .template-card.selected {
    border-color: #3b82f6;
    background: #eff6ff;
  }

  .template-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
  }

  .template-desc {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-group label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .form-group input,
  .form-group textarea {
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    transition: all 0.15s;
  }

  .form-group input:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-group input.error {
    border-color: #dc2626;
  }

  .form-group input:disabled,
  .form-group textarea:disabled {
    background: #f9fafb;
    color: #9ca3af;
    cursor: not-allowed;
  }

  .field-hint {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0;
  }

  .color-options {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .color-swatch {
    width: 2rem;
    height: 2rem;
    border: 2px solid transparent;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.15s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .color-swatch:hover {
    transform: scale(1.1);
  }

  .color-swatch.selected {
    border-color: #111827;
    box-shadow: 0 0 0 2px white, 0 0 0 4px currentColor;
  }

  .color-check {
    color: white;
    font-weight: bold;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  }

  .permissions-wrapper {
    margin-top: 0.5rem;
  }

  .permissions-hint {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 0.375rem;
  }

  .validation-hint {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    background: #fef3c7;
    color: #92400e;
    font-size: 0.875rem;
    border-radius: 0.375rem;
  }

  .modal-footer {
    display: flex;
    gap: 0.75rem;
    padding: 1.5rem;
    border-top: 1px solid #e5e7eb;
    flex-shrink: 0;
  }

  .cancel-button,
  .save-button {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .cancel-button {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .cancel-button:hover:not(:disabled) {
    background: #f9fafb;
  }

  .save-button {
    background: #3b82f6;
    color: white;
  }

  .save-button:hover:not(:disabled) {
    background: #2563eb;
  }

  .save-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Dark mode */
  :global(.dark) .modal-content {
    background: #1f2937;
  }

  :global(.dark) .modal-header {
    border-bottom-color: #374151;
  }

  :global(.dark) .modal-header h2 {
    color: #f9fafb;
  }

  :global(.dark) .modal-header p {
    color: #9ca3af;
  }

  :global(.dark) .section-label {
    color: #f3f4f6;
  }

  :global(.dark) .section-toggle {
    background: #111827;
    border-color: #374151;
  }

  :global(.dark) .template-card {
    background: #1f2937;
    border-color: #374151;
  }

  :global(.dark) .template-card.selected {
    border-color: #3b82f6;
    background: #1e3a8a;
  }

  :global(.dark) .template-name {
    color: #f9fafb;
  }

  :global(.dark) .form-group label {
    color: #d1d5db;
  }

  :global(.dark) .form-group input,
  :global(.dark) .form-group textarea {
    background: #111827;
    border-color: #374151;
    color: #f9fafb;
  }

  :global(.dark) .permissions-hint {
    background: #111827;
    color: #9ca3af;
  }

  :global(.dark) .modal-footer {
    border-top-color: #374151;
  }

  :global(.dark) .cancel-button {
    background: #111827;
    border-color: #374151;
    color: #d1d5db;
  }
</style>
