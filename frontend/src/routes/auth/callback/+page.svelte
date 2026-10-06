<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { checkOnboardingStatus, getSession, refreshSession } from '$lib/auth-client';

	let errorMessage = $state('');

	const redirectTo = $derived.by(() => {
		const value = $page.url.searchParams.get('redirect');
		return value && value.startsWith('/') && !value.startsWith('//') ? value : null;
	});

	onMount(async () => {
		if (!browser) return;
		try {
			await refreshSession();
			const session = await getSession();
			if (!session.data?.user) {
				errorMessage = 'BusinessOS could not verify your new session. Please sign in again.';
				return;
			}
			const onboardingStatus = await checkOnboardingStatus();
			if (redirectTo) {
				await goto(redirectTo);
			} else if (onboardingStatus.needsOnboarding) {
				await goto('/onboarding');
			} else {
				await goto('/window');
			}
		} catch {
			errorMessage = 'BusinessOS could not complete sign in. Please try again.';
		}
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-white">
	<div class="text-center">
		{#if errorMessage}
			<h1 class="text-xl font-semibold text-gray-900 mb-3">Sign in was not completed</h1>
			<p class="text-gray-500 text-sm mb-6">{errorMessage}</p>
			<a class="inline-flex min-h-11 items-center justify-center rounded border border-gray-900 bg-gray-900 px-5 text-sm font-medium text-white" href={redirectTo ? `/login?redirect=${encodeURIComponent(redirectTo)}` : '/login'}>Return to sign in</a>
		{:else}
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto mb-4"></div>
			<p class="text-gray-500 font-mono text-sm">Completing sign in...</p>
		{/if}
	</div>
</div>
