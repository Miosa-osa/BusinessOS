<script lang="ts">
	import { page } from '$app/stores';
	import Terminal from '$lib/components/desktop/Terminal.svelte';

	// Check if embedded (iframe mode)
	let isEmbed = $derived($page.url.searchParams.get('embed') === 'true');
</script>

<div class="terminal-page" class:embed={isEmbed}>
	<Terminal />
</div>

<style>
	.terminal-page {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		background: #1a1a1a;
	}

	.terminal-page.embed {
		/* Remove any padding/margin in embed mode */
		padding: 0;
		margin: 0;
	}

	.terminal-page :global(.terminal-wrapper) {
		flex: 1;
		width: 100%;
		height: 100%;
	}
</style>
