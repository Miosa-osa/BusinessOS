<script lang="ts">
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import LandingPage from './landing/+page.svelte';

	onMount(() => {
		// If already authenticated, go to dashboard
		if (document.cookie.includes('better-auth.session_token')) {
			goto('/dashboard');
		}
	});
</script>

<LandingPage />
