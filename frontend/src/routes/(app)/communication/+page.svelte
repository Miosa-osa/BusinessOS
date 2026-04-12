<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	// Redirect to default tab (calendar), preserving query params (like ?embed=true)
	onMount(() => {
		const searchParams = $page.url.searchParams.toString();
		const targetUrl = searchParams
			? `/communication/calendar?${searchParams}`
			: '/communication/calendar';
		goto(targetUrl, { replaceState: true });
	});
</script>

<div class="h-full flex items-center justify-center">
	<div class="text-gray-400 dark:text-gray-500">Loading...</div>
</div>
