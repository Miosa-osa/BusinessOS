<!-- 
  TypingIndicator.svelte
  Animated typing dots indicator showing AI is thinking
  Converted from Next.js typing-indicator.tsx
-->
<script lang="ts">
  import PurpleOrb from './PurpleOrb.svelte';

  export let message: string = "MIOSA is thinking...";
</script>

<div class="typing-indicator">
  <div class="avatar">
    <PurpleOrb size={32} className="animate-pulse" />
  </div>

  <div class="bubble">
    <div class="dots">
      <span class="dot" style="animation-delay: 0ms;"></span>
      <span class="dot" style="animation-delay: 150ms;"></span>
      <span class="dot" style="animation-delay: 300ms;"></span>
    </div>
    <span class="message">{message}</span>
  </div>
</div>

<style>
  .typing-indicator {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .avatar {
    width: 2rem;
    height: 2rem;
    flex-shrink: 0;
  }

  .bubble {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background-color: var(--muted, #f9fafb);
    border-radius: 1rem;
  }

  :global(.dark) .bubble {
    background-color: var(--muted, #1a1a1a);
  }

  .dots {
    display: flex;
    gap: 0.25rem;
  }

  .dot {
    width: 0.5rem;
    height: 0.5rem;
    background-color: var(--primary, #000000);
    border-radius: 9999px;
    animation: bounce 600ms ease-in-out infinite;
  }

  :global(.dark) .dot {
    background-color: var(--primary, #ffffff);
  }

  .message {
    font-size: 0.875rem;
    color: var(--muted-foreground, #6b7280);
  }

  @keyframes bounce {
    0%, 100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(-4px);
    }
  }
</style>
