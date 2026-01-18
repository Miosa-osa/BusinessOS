<!--
	GlassCard.svelte
	Reusable glassmorphism card component

	Usage:
	<GlassCard padding="lg" hoverable>
		<h2>Card Content</h2>
	</GlassCard>
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	type PaddingSize = 'none' | 'sm' | 'md' | 'lg' | 'xl';

	interface Props {
		padding?: PaddingSize;
		hoverable?: boolean;
		onclick?: (e: MouseEvent) => void;
		class?: string;
		children?: Snippet;
	}

	let {
		padding = 'md',
		hoverable = false,
		onclick,
		class: className = '',
		children
	}: Props = $props();

	const paddingClasses = {
		none: '',
		sm: 'p-3',
		md: 'p-6',
		lg: 'p-8',
		xl: 'p-12'
	};

	const classes = `glass-card ${paddingClasses[padding]} ${className}`.trim();
	const isClickable = onclick !== undefined;
</script>

<div
	class={classes}
	class:cursor-pointer={isClickable || hoverable}
	role={isClickable ? 'button' : undefined}
	tabindex={isClickable ? 0 : undefined}
	{onclick}
	onkeypress={(e) => e.key === 'Enter' && isClickable && onclick?.(e as any)}
>
	{#if children}
		{@render children()}
	{/if}
</div>
