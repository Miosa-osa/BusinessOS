<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { getSession } from '$lib/auth-client';
	import { onboardingAnalysis } from '$lib/stores/onboardingAnalysis';
	import { currentWorkspaceId } from '$lib/stores/workspaces';
	import { get } from 'svelte/store';

	onMount(async () => {
		if (!browser) return;

		// The OAuth flow completed and set a session cookie
		// Check if this is a new user or if onboarding is incomplete

		// Check if OAuth was initiated during onboarding flow
		const oauthContext = localStorage.getItem('oauth_context');
		const oauthNextRoute = localStorage.getItem('oauth_next_route');

		// Clear OAuth context
		if (oauthContext) {
			localStorage.removeItem('oauth_context');
			localStorage.removeItem('oauth_next_route');
		}

		// First check for new_user cookie (set by backend during Google OAuth signup)
		const newUserCookie = document.cookie.split('; ').find(row => row.startsWith('new_user='));
		const isNewUser = newUserCookie?.split('=')[1] === 'true';

		// Clear the cookie
		if (newUserCookie) {
			document.cookie = 'new_user=; path=/; max-age=0';
		}

		// If OAuth was initiated during onboarding signin, continue to next onboarding step
		if (oauthContext === 'onboarding-signin' && oauthNextRoute) {
			// 🔥 NEW: Trigger AI analysis after Gmail OAuth completes
			try {
				// Get user session
				const { data } = await getSession();
				const userId = data?.user?.id;

				// Get workspace (created by backend during OAuth)
				const workspaceId = get(currentWorkspaceId);

				if (userId && workspaceId) {
					console.log('🚀 Starting AI analysis:', { userId, workspaceId });

					// Start async analysis (streams to onboardingAnalysis store)
					// This will run in background while user sees analyzing screens
					onboardingAnalysis.start(userId, workspaceId, 50);
				} else {
					console.warn('Missing context for analysis:', { userId, workspaceId });
					// Continue anyway - analyzing screen will use fallback insights
				}
			} catch (err) {
				console.error('Failed to start analysis:', err);
				// Non-blocking - continue to next screen
			}

			// Route to next onboarding screen (usually /onboarding/analyzing)
			goto(oauthNextRoute);
			return;
		}

		// If new user, go to onboarding start
		if (isNewUser) {
			goto('/onboarding');
			return;
		}

		// For existing users, check onboarding status from session
		const { data } = await getSession();
		if (data?.user?.onboardingCompleted === false) {
			// Existing user who hasn't completed onboarding - resume from start
			goto('/onboarding');
			return;
		}

		// User has completed onboarding, go to main app
		goto('/window');
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-white">
	<div class="text-center">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto mb-4"></div>
		<p class="text-gray-500 font-mono text-sm">Completing sign in...</p>
	</div>
</div>
