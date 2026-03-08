<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { themeStore } from '$lib/stores/themeStore';
	import CookieConsent from '$lib/components/CookieConsent.svelte';

	let { children } = $props();

	// Initialize theme on mount
	onMount(() => {
		// Theme is already applied by the store on creation,
		// but we ensure it's set on the document
		const storedTheme = localStorage.getItem('theme');
		if (storedTheme === 'dark' || storedTheme === 'light' || storedTheme === 'system') {
			themeStore.setTheme(storedTheme);
		}
	});
</script>

<svelte:head>
	<title>Business OS</title>
	<meta name="description" content="Your internal command center" />
</svelte:head>

{@render children()}
<CookieConsent />
