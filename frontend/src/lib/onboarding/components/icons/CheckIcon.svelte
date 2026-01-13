<!-- 
  CheckIcon.svelte
  Checkmark icon
-->
<script lang="ts">
  export let size: number = 24;
  export let className: string = "";
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  class={className}
>
  <path d="M20 6 9 17l-5-5" />
</svg>
