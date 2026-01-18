<!--
	Onboarding Screen 7: AI Analysis (Second Insight)
	Shows second insight message from analysis
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { GradientBackground, GlassCard } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	let insightMessage = '';

	onMount(() => {
		// Get the second insight from store
		const state = $onboardingStore;
		insightMessage = state.analysis.message2 || 'Design tools are your playground';

		// Auto-advance to next screen after 2 seconds
		setTimeout(() => {
			onboardingStore.nextStep();
			goto('/onboarding/analyzing-3');
		}, 2000);
	});
</script>

<svelte:head>
	<title>Analyzing - OSA Build</title>
</svelte:head>

<GradientBackground variant="personalization" fullScreen>
	<div class="analyzing-screen text-center space-y-12 animate-slide-up">
		<!-- OSA Orb Animation -->
		<div class="orb-container">
			<div class="relative w-48 h-48 mx-auto">
				<!-- Main pulsing orb -->
				<div class="absolute inset-0 rounded-full bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 animate-pulse-glow"></div>
				<div class="absolute inset-4 rounded-full bg-gradient-to-br from-violet-300 via-purple-400 to-indigo-500 opacity-60 animate-pulse"></div>
				<div class="absolute inset-8 rounded-full bg-gradient-to-br from-white via-purple-200 to-indigo-200 opacity-40"></div>
			</div>
		</div>

		<!-- Second Insight Message -->
		<div class="status-section space-y-6">
			<GlassCard padding="xl" class="max-w-2xl mx-auto animate-fade-in">
				<div class="space-y-4">
					<div class="text-5xl">💡</div>
					<p class="text-2xl font-semibold text-gray-800 dark:text-gray-200">
						{insightMessage}
					</p>
				</div>
			</GlassCard>
		</div>

		<!-- Progress Indicator -->
		<div class="progress-dots flex justify-center gap-3">
			<div class="w-2 h-2 rounded-full bg-purple-500 animate-pulse"></div>
			<div class="w-2 h-2 rounded-full bg-purple-400 animate-pulse" style="animation-delay: 0.2s;"></div>
			<div class="w-2 h-2 rounded-full bg-purple-300 animate-pulse" style="animation-delay: 0.4s;"></div>
		</div>
	</div>
</GradientBackground>

<style>
	.analyzing-screen {
		padding: 4rem 2rem;
		max-width: 1200px;
		margin: 0 auto;
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		justify-content: center;
	}

	.orb-container {
		animation: fade-in 0.8s ease-out;
	}

	.status-section {
		min-height: 200px;
		display: flex;
		flex-direction: column;
		justify-content: center;
	}
</style>
