<!-- 
  Input.svelte
  Reusable input component
  Converted from Next.js input.tsx with shadcn/ui styling
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let value: string = "";
  export let placeholder: string = "";
  export let type: string = "text";
  export let disabled: boolean = false;
  export let className: string = "";
  export let id: string = "";
  export let name: string = "";

  const dispatch = createEventDispatcher();

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('input', { value });
  }

  function handleKeydown(event: KeyboardEvent) {
    dispatch('keydown', event);
  }

  function handleFocus(event: FocusEvent) {
    dispatch('focus', event);
  }

  function handleBlur(event: FocusEvent) {
    dispatch('blur', event);
  }
</script>

<input
  {type}
  {id}
  {name}
  {placeholder}
  {disabled}
  {value}
  class="input {className}"
  on:input={handleInput}
  on:keydown={handleKeydown}
  on:focus={handleFocus}
  on:blur={handleBlur}
/>

<style>
  .input {
    width: 100%;
    min-width: 0;
    height: 2.25rem;
    padding: 0.25rem 0.75rem;
    border: 1px solid var(--input, #e5e7eb);
    border-radius: 0.375rem;
    background-color: transparent;
    font-size: 1rem;
    color: var(--foreground, #1f2937);
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    transition: color 150ms, box-shadow 150ms;
    outline: none;
    font-family: inherit;
  }

  .input::placeholder {
    color: var(--muted-foreground, #6b7280);
  }

  .input:focus {
    border-color: var(--ring, #000000);
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .input {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: var(--input, #2a2a2a);
    color: var(--foreground, #f9fafb);
  }

  :global(.dark) .input:focus {
    box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.1);
  }

  .input:disabled {
    pointer-events: none;
    cursor: not-allowed;
    opacity: 0.5;
  }

  @media (min-width: 768px) {
    .input {
      font-size: 0.875rem;
    }
  }
</style>
