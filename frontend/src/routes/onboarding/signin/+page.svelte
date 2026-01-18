<!--
	Onboarding Screen 3: Sign In
	OAuth options for authentication
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton, GlassCard } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	function handleGoogleSignIn() {
		// TODO: Implement Google OAuth
		console.log('Google Sign In clicked');
		onboardingStore.nextStep();
		goto('/onboarding/gmail');
	}

	function handleAppleSignIn() {
		// TODO: Implement Apple OAuth
		console.log('Apple Sign In clicked');
		onboardingStore.nextStep();
		goto('/onboarding/gmail');
	}

	function handleEmailSignIn() {
		// TODO: Implement email/password
		console.log('Email Sign In clicked');
		onboardingStore.nextStep();
		goto('/onboarding/gmail');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/meet-osa');
	}
</script>

<svelte:head>
	<title>Sign In - OSA Build</title>
</svelte:head>

<GradientBackground variant="personalization" fullScreen>
	<div class="signin-screen text-center space-y-12 animate-slide-up">
		<div class="space-y-4">
			<h1 class="text-5xl font-bold text-gradient">
				Create Your Account
			</h1>
			<p class="text-xl text-gray-700 dark:text-gray-300 max-w-xl mx-auto">
				Choose how you'd like to sign in
			</p>
		</div>

		<!-- Sign-in options -->
		<div class="signin-options space-y-4 max-w-md mx-auto">
			<GlassCard padding="md" hoverable onclick={handleGoogleSignIn}>
				<button class="w-full flex items-center justify-center gap-3 p-4">
					<div class="text-2xl">🔵</div>
					<span class="text-lg font-semibold">Continue with Google</span>
				</button>
			</GlassCard>

			<GlassCard padding="md" hoverable onclick={handleAppleSignIn}>
				<button class="w-full flex items-center justify-center gap-3 p-4">
					<div class="text-2xl">🍎</div>
					<span class="text-lg font-semibold">Continue with Apple</span>
				</button>
			</GlassCard>

			<GlassCard padding="md" hoverable onclick={handleEmailSignIn}>
				<button class="w-full flex items-center justify-center gap-3 p-4">
					<div class="text-2xl">📧</div>
					<span class="text-lg font-semibold">Continue with Email</span>
				</button>
			</GlassCard>
		</div>

		<div class="text-sm text-gray-500 max-w-md mx-auto">
			By signing in, you agree to our Terms of Service and Privacy Policy
		</div>

		<div>
			<PillButton variant="ghost" size="md" onclick={handleBack}>
				Back
			</PillButton>
		</div>
	</div>
</GradientBackground>

<style>
	button {
		background: transparent;
		border: none;
		cursor: pointer;
		color: inherit;
		font-family: inherit;
	}
</style>
