<!--
	Onboarding Layout
	Minimal wrapper - only handles progress dots and resume prompt
	Each screen handles its own full-screen layout
-->
<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { ProgressDots, ResumePrompt } from '$lib/components/osa';
	import { getSession, checkOnboardingStatus } from '$lib/auth-client';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();

	const store = $derived($onboardingStore);
	let showResumePrompt = $state(false);
	let isAuthChecking = $state(true);
	let isAuthenticated = $state(false);

	onMount(async () => {
		if (!browser) return;

		// Check if user is authenticated
		const sessionResult = await getSession();

		if (!sessionResult.data?.user) {
			// Not authenticated, redirect to login
			goto('/login');
			return;
		}

		isAuthenticated = true;

		// Check if user actually needs onboarding
		const onboardingStatus = await checkOnboardingStatus();

		if (!onboardingStatus.needsOnboarding) {
			// User has completed onboarding, redirect to main app
			goto('/window');
			return;
		}

		isAuthChecking = false;

		// Only show resume prompt on the ROOT /onboarding page
		// Don't show it when user is on a sub-page (like /onboarding/building after OAuth)
		const currentPath = $page.url.pathname;
		const isRootOnboardingPage = currentPath === '/onboarding' || currentPath === '/onboarding/';

		if (isRootOnboardingPage && store.currentStep > 0 && !store.completed) {
			showResumePrompt = true;
		}
	});
</script>

{#if isAuthChecking}
	<!-- Loading state while checking auth -->
	<div class="min-h-screen flex items-center justify-center bg-white">
		<div class="text-center">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto mb-4"></div>
			<p class="text-gray-500 font-mono text-sm">Loading...</p>
		</div>
	</div>
{:else if isAuthenticated}
	<!-- Resume prompt for returning users -->
	<ResumePrompt bind:show={showResumePrompt} />

	<!-- Progress indicator - fixed at top -->
	<div class="progress-container">
		<ProgressDots total={store.totalSteps} current={store.currentStep} />
	</div>

	<!-- Full-screen content - no constraints -->
	{@render children()}
{/if}

<style>
	.progress-container {
		position: fixed;
		top: 2rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 100;
	}
</style>
