<!--
	GradientBackground.svelte
	iOS-inspired gradient background for onboarding screens

	Usage:
	<GradientBackground variant="welcome">
		<YourContent />
	</GradientBackground>
-->
<script lang="ts">
	type GradientVariant =
		| 'welcome'
		| 'signin'
		| 'personalization'
		| 'apps-showcase'
		| 'ready'
		| 'builder';

	interface Props {
		variant: GradientVariant;
		fullScreen?: boolean;
		children?: import('svelte').Snippet;
	}

	let {
		variant = 'welcome',
		fullScreen = true,
		children
	}: Props = $props();

	const gradientClass = `gradient-${variant}`;
</script>

<div
	class="{gradientClass} {fullScreen ? 'min-h-screen' : ''} transition-all duration-500"
	role="presentation"
>
	{#if children}
		{@render children()}
	{/if}
</div>

<style>
	/* Ensure smooth gradient transitions */
	div {
		background-attachment: fixed;
		background-size: cover;
	}
</style>

