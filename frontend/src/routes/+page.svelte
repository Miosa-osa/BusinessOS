<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { useSession } from '$lib/auth-client';

	const session = useSession();

	onMount(() => {
		// Check authentication and onboarding status
		if ($session.data) {
			// User is authenticated - check if they've completed onboarding
			if ($session.data.user?.onboardingCompleted === false) {
				// Send to onboarding even if authenticated
				goto('/onboarding');
			} else {
				// Completed onboarding, go to main app
				goto('/window');
			}
		} else {
			// Not authenticated, start onboarding flow
			goto('/onboarding');
		}
	});
</script>

<svelte:head>
	<title>BusinessOS - Getting Started</title>
</svelte:head>

<!-- Loading state while redirect happens -->
<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-white">
	<div class="text-center">
		<div class="w-16 h-16 mx-auto mb-4">
			<svg class="animate-spin w-full h-full text-gray-400" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
		</div>
		<p class="text-sm text-gray-500 font-mono">Loading...</p>
	</div>
</div>
