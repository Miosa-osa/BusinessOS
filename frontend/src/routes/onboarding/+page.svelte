<!--
	Onboarding Screen 1: Welcome to OSA Build
	First screen - introduces the platform
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	function handleGetStarted() {
		onboardingStore.nextStep();
		goto('/onboarding/meet-osa');
	}
</script>

<svelte:head>
	<title>Welcome to OSA Build</title>
</svelte:head>

<GradientBackground variant="welcome" fullScreen>
	<div class="welcome-screen text-center space-y-12 animate-slide-up">
		<!-- Logo / Hero -->
		<div class="space-y-6">
			<div class="hero-icon mx-auto">
				<!-- Placeholder for OSA logo/icon - using gradient circle for now -->
				<div class="w-32 h-32 rounded-full bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 mx-auto pulse-glow"></div>
			</div>

			<div class="space-y-4">
				<h1 class="text-6xl font-bold text-gradient">
					Welcome to OSA Build
				</h1>
				<p class="text-2xl text-gray-700 dark:text-gray-300 max-w-xl mx-auto">
					Build your personalized operating system in minutes, not months
				</p>
			</div>
		</div>

		<!-- Key features -->
		<div class="features-grid grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
			<div class="feature-card glass-card p-6 space-y-3">
				<div class="text-4xl">✨</div>
				<h3 class="text-xl font-semibold">AI-Powered</h3>
				<p class="text-gray-600 dark:text-gray-400">
					OSA builds apps tailored to your needs
				</p>
			</div>

			<div class="feature-card glass-card p-6 space-y-3">
				<div class="text-4xl">🚀</div>
				<h3 class="text-xl font-semibold">Day 1 Apps</h3>
				<p class="text-gray-600 dark:text-gray-400">
					Get 4 personalized apps immediately
				</p>
			</div>

			<div class="feature-card glass-card p-6 space-y-3">
				<div class="text-4xl">🎨</div>
				<h3 class="text-xl font-semibold">Fully Customizable</h3>
				<p class="text-gray-600 dark:text-gray-400">
					Build, remix, and share with ease
				</p>
			</div>
		</div>

		<!-- CTA -->
		<div class="cta-section space-y-4">
			<PillButton variant="primary" size="lg" onclick={handleGetStarted}>
				Get Started
			</PillButton>
			<p class="text-sm text-gray-500">Takes less than 2 minutes</p>
		</div>
	</div>
</GradientBackground>

<style>
	.welcome-screen {
		padding: 2rem 0;
	}

	.feature-card {
		transition: transform 0.2s ease;
	}

	.feature-card:hover {
		transform: translateY(-4px);
	}
</style>
