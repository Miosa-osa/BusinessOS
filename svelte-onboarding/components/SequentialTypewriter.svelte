<!-- 
  SequentialTypewriter.svelte
  Types multiple lines sequentially with pauses between
  Converted from Next.js sequential-typewriter.tsx
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import TypewriterText from './TypewriterText.svelte';

  export let lines: string[] = [];
  export let speed: number = 50;
  export let pauseBetween: number = 500;
  export let className: string = "";

  const dispatch = createEventDispatcher();

  let currentLineIndex = 0;
  let completedLines: string[] = [];

  function handleLineComplete() {
    completedLines = [...completedLines, lines[currentLineIndex]];

    setTimeout(() => {
      if (currentLineIndex < lines.length - 1) {
        currentLineIndex = currentLineIndex + 1;
      } else {
        dispatch('complete');
      }
    }, pauseBetween);
  }

  // Reset when lines change
  $: if (lines) {
    currentLineIndex = 0;
    completedLines = [];
  }
</script>

<div class="sequential-typewriter {className}">
  {#each completedLines as line, index (index)}
    <p class="line">{line}</p>
  {/each}
  
  {#if currentLineIndex < lines.length}
    <p class="line current-line">
      <TypewriterText 
        text={lines[currentLineIndex]} 
        {speed} 
        on:complete={handleLineComplete}
        showCursor={true}
      />
    </p>
  {/if}
</div>

<style>
  .sequential-typewriter {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .line {
    font-size: 1.125rem;
    color: var(--foreground, #1f2937);
    line-height: 1.625;
    margin: 0;
  }

  @media (min-width: 768px) {
    .line {
      font-size: 1.25rem;
    }
  }

  :global(.dark) .line {
    color: var(--foreground, #f9fafb);
  }
</style>
