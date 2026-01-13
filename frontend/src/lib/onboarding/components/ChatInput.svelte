<!-- 
  ChatInput.svelte
  Chat-style input with send button and optional mic
  Converted from Next.js chat-input.tsx
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';
  import SendIcon from './icons/SendIcon.svelte';
  import MicIcon from './icons/MicIcon.svelte';

  export let value: string = "";
  export let placeholder: string = "Type here...";
  export let showMic: boolean = false;
  export let disabled: boolean = false;
  export let className: string = "";

  const dispatch = createEventDispatcher();

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey && value.trim()) {
      event.preventDefault();
      dispatch('submit');
    }
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('change', { value });
  }

  function handleSubmit() {
    if (value.trim()) {
      dispatch('submit');
    }
  }

  function handleMicClick() {
    dispatch('mic');
  }
</script>

<div class="chat-input-wrapper {className}">
  <input
    type="text"
    {value}
    {placeholder}
    {disabled}
    class="chat-input"
    on:input={handleInput}
    on:keydown={handleKeyPress}
  />
  <div class="actions">
    {#if showMic}
      <Button 
        variant="ghost" 
        size="icon" 
        className="action-btn rounded-full"
        on:click={handleMicClick}
      >
        <MicIcon size={20} />
      </Button>
    {/if}
    <Button 
      size="icon" 
      className="action-btn send-btn rounded-full"
      disabled={disabled || !value.trim()}
      on:click={handleSubmit}
    >
      <SendIcon size={20} />
    </Button>
  </div>
</div>

<style>
  .chat-input-wrapper {
    position: relative;
    max-width: 42rem;
    width: 100%;
    margin: 0 auto;
  }

  .chat-input {
    width: 100%;
    height: 3.5rem;
    padding: 0 6rem 0 1.5rem;
    border: 2px solid var(--border, #e5e7eb);
    border-radius: 9999px;
    background-color: var(--background, #ffffff);
    font-size: 1rem;
    color: var(--foreground, #1f2937);
    outline: none;
    transition: border-color 150ms;
    font-family: inherit;
  }

  .chat-input::placeholder {
    color: var(--muted-foreground, #6b7280);
  }

  .chat-input:focus {
    border-color: var(--primary, #000000);
  }

  :global(.dark) .chat-input {
    background-color: var(--background, #0a0a0a);
    border-color: var(--border, #2a2a2a);
    color: var(--foreground, #f9fafb);
  }

  :global(.dark) .chat-input:focus {
    border-color: var(--primary, #ffffff);
  }

  .actions {
    position: absolute;
    right: 0.75rem;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    gap: 0.5rem;
  }

  :global(.action-btn) {
    width: 2.75rem !important;
    height: 2.75rem !important;
    border-radius: 9999px !important;
  }

  :global(.send-btn) {
    background-color: var(--primary, #000000) !important;
    color: var(--primary-foreground, #ffffff) !important;
  }

  :global(.dark .send-btn) {
    background-color: var(--primary, #ffffff) !important;
    color: var(--primary-foreground, #000000) !important;
  }
</style>
