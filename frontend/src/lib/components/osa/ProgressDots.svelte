<!--
	ProgressDots.svelte
	iOS-style progress dots indicator for onboarding

	Usage:
	<ProgressDots total={4} current={1} />
-->
<script lang="ts">
	interface Props {
		total: number;
		current: number;
		class?: string;
	}

	let {
		total,
		current,
		class: className = ''
	}: Props = $props();

	const classes = `progress-dots ${className}`.trim();

	// Generate array of dots
	const dots = Array.from({ length: total }, (_, i) => i);
</script>

<div class={classes} role="progressbar" aria-valuenow={current} aria-valuemin={0} aria-valuemax={total}>
	{#each dots as index}
		<div
			class="progress-dot"
			class:active={index === current}
		></div>
	{/each}
</div>
