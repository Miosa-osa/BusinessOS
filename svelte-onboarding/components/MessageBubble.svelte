<!-- 
  MessageBubble.svelte
  Chat message bubble for agent and user messages
  For the conversational onboarding flow
-->
<script lang="ts">
  import PurpleOrb from './PurpleOrb.svelte';

  export let role: 'agent' | 'user' = 'agent';
  export let content: string = "";
  export let timestamp: string = "";
  export let showAvatar: boolean = true;
</script>

<div class="message-bubble {role}">
  {#if role === 'agent' && showAvatar}
    <div class="avatar">
      <PurpleOrb size={32} />
    </div>
  {/if}
  
  <div class="bubble-content">
    <div class="bubble">
      <slot>
        {content}
      </slot>
    </div>
    {#if timestamp}
      <span class="timestamp">{timestamp}</span>
    {/if}
  </div>
</div>

<style>
  .message-bubble {
    display: flex;
    gap: 0.75rem;
    max-width: 85%;
  }

  .message-bubble.agent {
    align-self: flex-start;
  }

  .message-bubble.user {
    align-self: flex-end;
    flex-direction: row-reverse;
  }

  .avatar {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
  }

  .bubble-content {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .message-bubble.user .bubble-content {
    align-items: flex-end;
  }

  .bubble {
    padding: 0.75rem 1rem;
    border-radius: 1rem;
    font-size: 0.9375rem;
    line-height: 1.5;
    word-wrap: break-word;
  }

  .message-bubble.agent .bubble {
    background-color: var(--muted, #f9fafb);
    color: var(--foreground, #1f2937);
    border-bottom-left-radius: 0.25rem;
  }

  :global(.dark) .message-bubble.agent .bubble {
    background-color: var(--muted, #1a1a1a);
    color: var(--foreground, #f9fafb);
  }

  .message-bubble.user .bubble {
    background-color: var(--primary, #000000);
    color: var(--primary-foreground, #ffffff);
    border-bottom-right-radius: 0.25rem;
  }

  :global(.dark) .message-bubble.user .bubble {
    background-color: var(--primary, #ffffff);
    color: var(--primary-foreground, #000000);
  }

  .timestamp {
    font-size: 0.75rem;
    color: var(--muted-foreground, #6b7280);
    padding: 0 0.5rem;
  }
</style>
