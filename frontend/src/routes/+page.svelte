<script lang="ts">
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';

	// Check if running in Electron
	const isElectron = browser && typeof window !== 'undefined' && 'electron' in window;

	onMount(() => {
		// If already authenticated (has session cookie), go straight to dashboard
		const hasSession = document.cookie.includes('better-auth.session_token');
		if (hasSession) {
			goto('/dashboard');
			return;
		}

		// Electron without session → show mode selector or landing
		// Web without session → show landing page
		goto('/landing');
	});
</script>

<!-- Brief loading state while redirect happens -->
<div style="min-height: 100vh; background: white;"></div>
