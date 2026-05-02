<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';

	const ENGINE_UI_URL = 'http://localhost:4201';

	let view = $derived($page.url.searchParams.get('view') || 'graph');
</script>

<svelte:head>
	<title>Engine | BusinessOS</title>
</svelte:head>

<div class="engine-page">
	{#if browser}
		<iframe
			src="{ENGINE_UI_URL}/{view}"
			title="Optimal Engine — {view}"
			class="engine-iframe"
			allow="clipboard-read; clipboard-write"
		></iframe>
	{/if}
</div>

<style>
	.engine-page {
		width: 100%;
		height: 100vh;
		overflow: hidden;
		background: #0d0d0d;
	}

	.engine-iframe {
		width: 100%;
		height: 100%;
		border: none;
		background: #0d0d0d;
	}
</style>
