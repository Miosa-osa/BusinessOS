<!-- 
  FloatingChatScreen.svelte
  Main layout container for the onboarding chat experience
  Converted from Next.js floating-chat-screen.tsx
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import PurpleOrb from './PurpleOrb.svelte';
  import ThemeToggle from './ThemeToggle.svelte';
  import Button from './Button.svelte';
  import ArrowLeftIcon from './icons/ArrowLeftIcon.svelte';

  export let showBack: boolean = true;

  const dispatch = createEventDispatcher();

  function handleBack() {
    dispatch('back');
  }
</script>

<div class="floating-chat-screen">
  {#if showBack}
    <Button 
      variant="ghost" 
      size="icon" 
      className="back-button rounded-full"
      on:click={handleBack}
    >
      <ArrowLeftIcon size={20} />
    </Button>
  {/if}

  <div class="theme-toggle-wrapper">
    <ThemeToggle />
  </div>

  <div class="content-container">
    <PurpleOrb size={80} className="orb" />
    <slot />
  </div>
</div>

<style>
  .floating-chat-screen {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--background, #ffffff);
    padding: 1rem;
    position: relative;
  }

  :global(.dark) .floating-chat-screen {
    background-color: var(--background, #0a0a0a);
  }

  :global(.back-button) {
    position: absolute !important;
    top: 1rem;
    left: 1rem;
    width: 2.75rem !important;
    height: 2.75rem !important;
    border-radius: 9999px !important;
  }

  .theme-toggle-wrapper {
    position: absolute;
    top: 1rem;
    right: 1rem;
  }

  .content-container {
    width: 100%;
    max-width: 42rem;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  :global(.orb) {
    margin-bottom: 3rem;
  }
</style>
