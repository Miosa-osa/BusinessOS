<!-- 
  Button.svelte
  Reusable button component with multiple variants
  Converted from Next.js button.tsx with shadcn/ui styling
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let variant: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link' = 'default';
  export let size: 'default' | 'sm' | 'lg' | 'icon' | 'icon-sm' | 'icon-lg' = 'default';
  export let disabled: boolean = false;
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let className: string = "";

  const dispatch = createEventDispatcher();

  function handleClick(event: MouseEvent) {
    if (!disabled) {
      dispatch('click', event);
    }
  }

  // Compute variant class
  $: variantClass = {
    'default': 'btn-primary',
    'destructive': 'btn-destructive',
    'outline': 'btn-outline',
    'secondary': 'btn-secondary',
    'ghost': 'btn-ghost',
    'link': 'btn-link'
  }[variant];

  // Compute size class
  $: sizeClass = {
    'default': 'btn-md',
    'sm': 'btn-sm',
    'lg': 'btn-lg',
    'icon': 'btn-icon',
    'icon-sm': 'btn-icon-sm',
    'icon-lg': 'btn-icon-lg'
  }[size];
</script>

<button
  {type}
  {disabled}
  class="btn {variantClass} {sizeClass} {className}"
  on:click={handleClick}
>
  <slot />
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    white-space: nowrap;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 150ms cubic-bezier(0, 0, 0.2, 1);
    outline: none;
    cursor: pointer;
    border: none;
    font-family: inherit;
  }

  .btn:disabled {
    pointer-events: none;
    opacity: 0.5;
  }

  .btn:focus-visible {
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .btn:focus-visible {
    box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.1);
  }

  /* Variants */
  .btn-primary {
    background-color: var(--primary, #000000);
    color: var(--primary-foreground, #ffffff);
  }

  .btn-primary:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn-destructive {
    background-color: var(--destructive, #ef4444);
    color: #ffffff;
  }

  .btn-destructive:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn-outline {
    border: 1px solid var(--border, #e5e7eb);
    background-color: var(--background, #ffffff);
    color: var(--foreground, #1f2937);
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  }

  .btn-outline:hover:not(:disabled) {
    background-color: var(--accent, #f3f4f6);
    color: var(--accent-foreground, #1f2937);
  }

  :global(.dark) .btn-outline {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: var(--input, #2a2a2a);
  }

  :global(.dark) .btn-outline:hover:not(:disabled) {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .btn-secondary {
    background-color: var(--secondary, #f9fafb);
    color: var(--secondary-foreground, #1f2937);
  }

  .btn-secondary:hover:not(:disabled) {
    opacity: 0.8;
  }

  .btn-ghost {
    background-color: transparent;
    color: var(--foreground, #1f2937);
  }

  .btn-ghost:hover:not(:disabled) {
    background-color: var(--accent, #f3f4f6);
    color: var(--accent-foreground, #1f2937);
  }

  :global(.dark) .btn-ghost:hover:not(:disabled) {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .btn-link {
    background-color: transparent;
    color: var(--primary, #000000);
    text-decoration-line: underline;
    text-underline-offset: 4px;
  }

  .btn-link:hover:not(:disabled) {
    text-decoration-line: underline;
  }

  /* Sizes */
  .btn-sm {
    height: 2rem;
    padding: 0 0.75rem;
    border-radius: 0.375rem;
    gap: 0.375rem;
    font-size: 0.75rem;
  }

  .btn-md {
    height: 2.25rem;
    padding: 0 1rem;
  }

  .btn-lg {
    height: 2.5rem;
    padding: 0 1.5rem;
    border-radius: 0.375rem;
  }

  .btn-icon {
    width: 2.25rem;
    height: 2.25rem;
    padding: 0;
  }

  .btn-icon-sm {
    width: 2rem;
    height: 2rem;
    padding: 0;
  }

  .btn-icon-lg {
    width: 2.5rem;
    height: 2.5rem;
    padding: 0;
  }
</style>
