<!-- 
  ToolPicker.svelte
  Multi-select tool/integration picker grid
  For selecting which tools the business uses
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';
  import CheckIcon from './icons/CheckIcon.svelte';

  interface Tool {
    id: string;
    name: string;
    icon?: string;
  }

  export let tools: Tool[] = [
    { id: 'google', name: 'Google Workspace' },
    { id: 'microsoft', name: 'Microsoft 365' },
    { id: 'slack', name: 'Slack' },
    { id: 'notion', name: 'Notion' },
    { id: 'linear', name: 'Linear' },
    { id: 'hubspot', name: 'HubSpot' },
    { id: 'airtable', name: 'Airtable' },
    { id: 'clickup', name: 'ClickUp' },
    { id: 'fathom', name: 'Fathom' },
  ];

  export let selectedTools: string[] = [];
  export let otherTools: string = "";
  export let showOtherInput: boolean = true;

  const dispatch = createEventDispatcher();

  function toggleTool(toolId: string) {
    if (selectedTools.includes(toolId)) {
      selectedTools = selectedTools.filter(id => id !== toolId);
    } else {
      selectedTools = [...selectedTools, toolId];
    }
    dispatch('change', { selectedTools, otherTools });
  }

  function handleOtherChange(event: Event) {
    otherTools = (event.target as HTMLInputElement).value;
    dispatch('change', { selectedTools, otherTools });
  }

  function handleContinue() {
    dispatch('continue', { selectedTools, otherTools });
  }
</script>

<div class="tool-picker">
  <p class="label">Which tools do you use? (select all that apply)</p>
  
  <div class="tools-grid">
    {#each tools as tool}
      <button
        class="tool-item"
        class:selected={selectedTools.includes(tool.id)}
        on:click={() => toggleTool(tool.id)}
      >
        {#if tool.icon}
          <img src={tool.icon} alt={tool.name} class="tool-icon" />
        {:else}
          <div class="tool-icon-placeholder">{tool.name.charAt(0)}</div>
        {/if}
        <span class="tool-name">{tool.name}</span>
        {#if selectedTools.includes(tool.id)}
          <div class="check-mark">
            <CheckIcon size={14} />
          </div>
        {/if}
      </button>
    {/each}
  </div>

  {#if showOtherInput}
    <div class="other-tools">
      <label for="other-tools" class="other-label">Other tools:</label>
      <input
        id="other-tools"
        type="text"
        value={otherTools}
        on:input={handleOtherChange}
        placeholder="e.g., Trello, Asana, Monday..."
        class="other-input"
      />
    </div>
  {/if}

  <Button 
    className="continue-btn"
    disabled={selectedTools.length === 0 && !otherTools.trim()}
    on:click={handleContinue}
  >
    Continue with {selectedTools.length} selected
  </Button>
</div>

<style>
  .tool-picker {
    width: 100%;
    max-width: 32rem;
    margin: 0 auto;
    padding: 1rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.75rem;
  }

  :global(.dark) .tool-picker {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .label {
    margin: 0 0 1rem;
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .label {
    color: var(--foreground, #f9fafb);
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  @media (min-width: 480px) {
    .tools-grid {
      grid-template-columns: repeat(3, 1fr);
    }
  }

  .tool-item {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.5rem;
    background-color: transparent;
    cursor: pointer;
    transition: all 150ms ease;
    font-family: inherit;
  }

  :global(.dark) .tool-item {
    border-color: var(--border, #2a2a2a);
  }

  .tool-item:hover {
    border-color: var(--primary, #000000);
  }

  :global(.dark) .tool-item:hover {
    border-color: var(--primary, #ffffff);
  }

  .tool-item.selected {
    border-color: var(--primary, #000000);
    background-color: rgba(0, 0, 0, 0.03);
  }

  :global(.dark) .tool-item.selected {
    border-color: var(--primary, #ffffff);
    background-color: rgba(255, 255, 255, 0.05);
  }

  .tool-icon {
    width: 2rem;
    height: 2rem;
    object-fit: contain;
  }

  .tool-icon-placeholder {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--muted, #f9fafb);
    border-radius: 0.25rem;
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .tool-icon-placeholder {
    background-color: var(--muted, #1a1a1a);
    color: var(--foreground, #f9fafb);
  }

  .tool-name {
    font-size: 0.75rem;
    color: var(--foreground, #1f2937);
    text-align: center;
  }

  :global(.dark) .tool-name {
    color: var(--foreground, #f9fafb);
  }

  .check-mark {
    position: absolute;
    top: 0.375rem;
    right: 0.375rem;
    width: 1.25rem;
    height: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--primary, #000000);
    border-radius: 9999px;
    color: var(--primary-foreground, #ffffff);
  }

  :global(.dark) .check-mark {
    background-color: var(--primary, #ffffff);
    color: var(--primary-foreground, #000000);
  }

  .other-tools {
    margin-bottom: 1rem;
  }

  .other-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.875rem;
    color: var(--muted-foreground, #6b7280);
  }

  .other-input {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.375rem;
    background-color: transparent;
    font-size: 0.875rem;
    color: var(--foreground, #1f2937);
    outline: none;
    font-family: inherit;
  }

  .other-input::placeholder {
    color: var(--muted-foreground, #6b7280);
  }

  .other-input:focus {
    border-color: var(--ring, #000000);
  }

  :global(.dark) .other-input {
    border-color: var(--border, #2a2a2a);
    color: var(--foreground, #f9fafb);
  }

  :global(.continue-btn) {
    width: 100% !important;
  }
</style>
