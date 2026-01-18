<!--
	Onboarding Screen 4: Connect Gmail (Optional)
	Allow OSA to personalize based on email data
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton, GlassCard } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	function handleConnect() {
		// TODO: Implement Gmail OAuth
		console.log('Connect Gmail clicked');
		onboardingStore.setUserData({ gmailConnected: true });
		onboardingStore.nextStep();
		goto('/onboarding/username');
	}

	function handleSkip() {
		onboardingStore.setUserData({ gmailConnected: false });
		onboardingStore.nextStep();
		goto('/onboarding/username');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/signin');
	}
</script>

<svelte:head>
	<title>Connect Gmail - OSA Build</title>
</svelte:head>

<GradientBackground variant="apps-showcase" fullScreen>
	<div class="gmail-screen text-center space-y-12 animate-slide-up">
		<div class="space-y-6">
			<div class="gmail-icon mx-auto w-24 h-24">
				<div class="text-8xl">📬</div>
			</div>

			<div class="space-y-4">
				<h1 class="text-5xl font-bold text-gradient">
					Connect Your Gmail
				</h1>
				<p class="text-xl text-gray-700 dark:text-gray-300 max-w-2xl mx-auto">
					Let OSA analyze your email to build apps tailored to your needs
				</p>
			</div>
		</div>

		<!-- Benefits -->
		<div class="benefits max-w-2xl mx-auto">
			<GlassCard padding="lg">
				<div class="space-y-6 text-left">
					<h3 class="text-2xl font-semibold text-center">Why Connect?</h3>

					<div class="space-y-4">
						<div class="flex gap-4 items-start">
							<div class="text-2xl">✨</div>
							<div>
								<h4 class="font-semibold">Smarter Personalization</h4>
								<p class="text-gray-600 dark:text-gray-400 text-sm">
									OSA understands your workflow patterns and priorities
								</p>
							</div>
						</div>

						<div class="flex gap-4 items-start">
							<div class="text-2xl">🎯</div>
							<div>
								<h4 class="font-semibold">Better First Apps</h4>
								<p class="text-gray-600 dark:text-gray-400 text-sm">
									Get 4 apps designed specifically for your actual needs
								</p>
							</div>
						</div>

						<div class="flex gap-4 items-start">
							<div class="text-2xl">🔒</div>
							<div>
								<h4 class="font-semibold">Privacy First</h4>
								<p class="text-gray-600 dark:text-gray-400 text-sm">
									Your data is encrypted and never shared. Disconnect anytime.
								</p>
							</div>
						</div>
					</div>
				</div>
			</GlassCard>
		</div>

		<!-- CTA -->
		<div class="cta-section space-y-4">
			<div class="flex gap-4 justify-center">
				<PillButton variant="ghost" size="md" onclick={handleBack}>
					Back
				</PillButton>
				<PillButton variant="primary" size="lg" onclick={handleConnect}>
					Connect Gmail
				</PillButton>
			</div>
			<button
				class="text-gray-500 hover:text-gray-700 text-sm underline"
				onclick={handleSkip}
			>
				Skip for now
			</button>
		</div>
	</div>
</GradientBackground>

<style>
	button {
		background: transparent;
		border: none;
		cursor: pointer;
		font-family: inherit;
	}
</style>
