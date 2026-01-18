<!--
	Onboarding Screen 6: AI Analysis (First Insight)
	Shows OSA analyzing user data with first insight message
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { GradientBackground, GlassCard } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { analyzeUser } from '$lib/api/osa-onboarding';

	let analyzing = true;
	let error: string | null = null;
	let insightMessage = '';

	onMount(async () => {
		try {
			// Get user data from store
			const state = $onboardingStore;
			const email = state.userData.email || '';
			const gmailConnected = state.userData.gmailConnected || false;

			// Call analyze API
			const response = await analyzeUser(email, gmailConnected);

			// Store analysis results
			onboardingStore.setAnalysis({
				message1: response.analysis.insights[0] || 'No-code builder energy ✨',
				message2: response.analysis.insights[1] || 'Design tools are your playground',
				message3: response.analysis.insights[2] || 'AI-curious, testing new platforms'
			});

			// Store interests and tools for later use
			onboardingStore.setUserData({
				interests: response.analysis.interests
			});

			// Show the first insight
			insightMessage = response.analysis.insights[0] || 'No-code builder energy ✨';
			analyzing = false;

			// Auto-advance to next screen after 2 seconds
			setTimeout(() => {
				onboardingStore.nextStep();
				goto('/onboarding/analyzing-2');
			}, 2000);
		} catch (err) {
			console.error('Error analyzing user:', err);
			error = err instanceof Error ? err.message : 'Failed to analyze user data';
			analyzing = false;

			// Use fallback insights on error
			insightMessage = 'No-code builder energy ✨';

			// Store fallback analysis
			onboardingStore.setAnalysis({
				message1: 'No-code builder energy ✨',
				message2: 'Design tools are your playground',
				message3: 'AI-curious, testing new platforms'
			});

			// Still advance after delay
			setTimeout(() => {
				onboardingStore.nextStep();
				goto('/onboarding/analyzing-2');
			}, 2000);
		}
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

		<!-- Status Text -->
		<div class="status-section space-y-6">
			{#if analyzing}
				<h1 class="text-4xl font-bold text-gray-800 dark:text-gray-200">
					OSA is analyzing your data...
				</h1>
				<p class="text-lg text-gray-600 dark:text-gray-400">
					Discovering patterns and preferences
				</p>
			{:else if error}
				<h1 class="text-4xl font-bold text-gray-800 dark:text-gray-200">
					Analysis Complete
				</h1>
				<p class="text-lg text-red-500">
					{error}
				</p>
			{:else}
				<!-- First Insight Message -->
				<GlassCard padding="xl" class="max-w-2xl mx-auto animate-fade-in">
					<div class="space-y-4">
						<div class="text-5xl">💡</div>
						<p class="text-2xl font-semibold text-gray-800 dark:text-gray-200">
							{insightMessage}
						</p>
					</div>
				</GlassCard>
			{/if}
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
