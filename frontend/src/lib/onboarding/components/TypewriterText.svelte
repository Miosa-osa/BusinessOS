<!-- 
  TypewriterText.svelte
  A single-line typewriter text effect component
  Converted from Next.js typewriter-text.tsx
-->
<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';

  export let text: string = "";
  export let speed: number = 50;
  export let showCursor: boolean = false;
  export let className: string = "";

  const dispatch = createEventDispatcher();

  let displayText = "";
  let isComplete = false;

  onMount(() => {
    let currentIndex = 0;
    displayText = "";
    isComplete = false;

    const interval = setInterval(() => {
      if (currentIndex < text.length) {
        displayText = text.slice(0, currentIndex + 1);
        currentIndex++;
      } else {
        isComplete = true;
        clearInterval(interval);
        dispatch('complete');
      }
    }, speed);

    return () => clearInterval(interval);
  });

  // Re-run animation when text changes
  $: if (text) {
    let currentIndex = 0;
    displayText = "";
    isComplete = false;
    
    const runAnimation = () => {
      const interval = setInterval(() => {
        if (currentIndex < text.length) {
          displayText = text.slice(0, currentIndex + 1);
          currentIndex++;
        } else {
          isComplete = true;
          clearInterval(interval);
          dispatch('complete');
        }
      }, speed);
    };
    
    // Small delay to reset
    setTimeout(runAnimation, 10);
  }
</script>

<span class={className}>
  {displayText}
  {#if showCursor && !isComplete}
    <span class="cursor"></span>
  {/if}
</span>

<style>
  .cursor {
    display: inline-block;
    width: 0.125rem;
    height: 1.2em;
    background-color: currentColor;
    margin-left: 0.125rem;
    vertical-align: text-bottom;
    animation: blink-cursor 1s step-end infinite;
  }

  @keyframes blink-cursor {
    0%, 49% {
      opacity: 1;
    }
    50%, 100% {
      opacity: 0;
    }
  }
</style>
