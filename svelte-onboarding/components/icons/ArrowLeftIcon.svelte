<!-- 
  ArrowLeftIcon.svelte
  Back arrow icon
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
  <path d="m12 19-7-7 7-7" />
  <path d="M19 12H5" />
</svg>
