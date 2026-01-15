<!-- 
  ActionButtons.svelte
  Quick action buttons that appear below agent messages
  For options like industry selection, tool selection, etc.
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';

  export let buttons: Array<{ label: string; action: string; variant?: 'default' | 'outline' | 'ghost' }> = [];
  export let layout: 'inline' | 'grid' | 'stack' = 'inline';
  export let className: string = "";

  const dispatch = createEventDispatcher();

  function handleClick(action: string) {
    dispatch('action', { action });
  }
</script>

<div class="action-buttons {layout} {className}">
  {#each buttons as button}
    <Button
      variant={button.variant || 'outline'}
      className="action-btn"
      on:click={() => handleClick(button.action)}
    >
      {button.label}
    </Button>
  {/each}
</div>

<style>
  .action-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .action-buttons.inline {
    flex-direction: row;
  }

  .action-buttons.stack {
    flex-direction: column;
    align-items: stretch;
  }

  .action-buttons.grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 0.5rem;
  }

  :global(.action-btn) {
    padding: 0.5rem 1rem !important;
    height: auto !important;
    min-height: 2.5rem !important;
  }
</style>
