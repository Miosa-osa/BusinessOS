<!-- 
  PurpleOrb.svelte
  A glowing animated purple orb icon for the BusinessOS onboarding
  Converted from Next.js silver-orb.tsx component
-->
<script lang="ts">
  export let size: number = 80;
  export let className: string = "";
</script>

<div 
  class="orb-container {className}" 
  style="width: {size}px; height: {size}px;"
>
  <div class="orb-body animate-pulse-orb">
    <div class="orb-shimmer animate-shimmer"></div>
  </div>
</div>

<style>
  .orb-container {
    position: relative;
  }

  .orb-body {
    position: absolute;
    inset: 0;
    border-radius: 9999px;
    background: linear-gradient(
      135deg,
      #e0e7ff 0%,
      #a5b4fc 25%,
      #818cf8 50%,
      #6366f1 75%,
      #4f46e5 100%
    );
    box-shadow: 
      0 0 40px rgba(99, 102, 241, 0.4),
      inset 0 0 20px rgba(255, 255, 255, 0.3);
  }

  .orb-shimmer {
    position: absolute;
    inset: 0;
    border-radius: 9999px;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.4),
      transparent
    );
    background-size: 1000px 100%;
  }

  /* Animation Keyframes */
  @keyframes pulse-orb {
    0%, 100% {
      transform: scale(1);
      filter: brightness(1);
    }
    50% {
      transform: scale(1.05);
      filter: brightness(1.2);
    }
  }

  @keyframes shimmer {
    0% {
      background-position: -1000px 0;
    }
    100% {
      background-position: 1000px 0;
    }
  }

  .animate-pulse-orb {
    animation: pulse-orb 2s ease-in-out infinite;
  }

  .animate-shimmer {
    animation: shimmer 3s linear infinite;
  }
</style>
