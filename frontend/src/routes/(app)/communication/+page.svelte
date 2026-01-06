<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// Redirect to default tab (calendar)
	onMount(() => {
		goto('/communication/calendar', { replaceState: true });
	});
</script>

<div class="h-full flex items-center justify-center">
	<div class="text-gray-400 dark:text-gray-500">Loading...</div>
</div>
