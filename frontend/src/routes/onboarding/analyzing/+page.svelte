<!--
	Onboarding Screen 6: AI Analysis (First Insight)
	Streams AI-generated insights in real-time using Groq AI
	Displays the first of 3 personalized insights
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { onboardingAnalysis, analyzingInsights, analysisFailed } from '$lib/stores/onboardingAnalysis';
	import { getSession } from '$lib/auth-client';

	let analyzing = true;
	let error: string | null = null;
	let insightMessage = '';

	onMount(async () => {
		// Subscribe to streaming analysis store
		const unsubscribe = analyzingInsights.subscribe(($insights) => {
			insightMessage = $insights.message1;

			// Update legacy onboarding store for backward compatibility
			onboardingStore.setAnalysis({
				message1: $insights.message1,
				message2: $insights.message2,
				message3: $insights.message3
			});
		});

		// Subscribe to analysis state
		const unsubscribeAnalysis = onboardingAnalysis.subscribe(($analysis) => {
			// Update analyzing state based on stream status
			analyzing = $analysis.isStreaming || $analysis.isLoading;
			error = $analysis.error;

			// Store additional data
			if ($analysis.interests.length > 0) {
				onboardingStore.setUserData({
					interests: $analysis.interests
				});
			}

			// When analysis completes (or fails but has fallback data)
			if ($analysis.status === 'completed' || $analysis.status === 'failed') {
				analyzing = false;

				// Auto-advance to next screen after 2s
				setTimeout(() => {
					onboardingStore.nextStep();
					goto('/onboarding/analyzing-2');
				}, 2000);
			}
		});

		// Get user session and start polling for real data
		const session = await getSession();
		if (session.data && session.data.user && session.data.user.id) {
			const userId = session.data.user.id;
			console.log('[Analyzing] Starting analysis polling for user:', userId);

			// Start polling for analysis status by user_id
			onboardingAnalysis.pollByUserId(userId);
		} else {
			console.warn('[Analyzing] No user session found - using fallback insights');

			// Set fallback insights
			insightMessage = 'No-code builder energy';
			analyzing = false;

			onboardingStore.setAnalysis({
				message1: 'No-code builder energy',
				message2: 'Design tools are your playground',
				message3: 'AI-curious, testing new platforms'
			});

			// Still auto-advance
			setTimeout(() => {
				onboardingStore.nextStep();
				goto('/onboarding/analyzing-2');
			}, 2000);
		}

		return () => {
			unsubscribe();
			unsubscribeAnalysis();
		};
	});
</script>

<svelte:head>
	<title>Analyzing - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="analyzing-screen">
		<div class="content">
			{#if analyzing}
				<h1 class="title">
					Analyzing your workspace...
				</h1>

				<!-- Loading Spinner -->
				<div class="spinner-wrapper">
					<div class="spinner"></div>
				</div>

				<!-- Streaming indicator -->
				{#if $onboardingAnalysis.isStreaming}
					<p class="streaming-text">Reading your emails with AI...</p>
				{/if}
			{:else if error}
				<h1 class="title">
					Analysis complete
				</h1>
				<p class="error-text">{error}</p>
				<p class="fallback-text">Using default insights</p>
			{:else}
				<h1 class="title">
					{insightMessage}
				</h1>

				{#if $analyzingInsights.hasRealData}
					<p class="ai-badge">✨ AI-Generated</p>
				{/if}
			{/if}
		</div>
	</div>
</div>

<style>
	.onboarding-background {
		min-height: 100vh;
		width: 100%;
		background-image: url('/logos/integrations/MIOSABRANDBackround.png');
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
	}

	.analyzing-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.content {
		width: 100%;
		max-width: 600px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 3rem;
		text-align: center;
	}

	.title {
		font-size: 2.75rem;
		font-weight: 700;
		color: #1A1A1A;
		line-height: 1.2;
		letter-spacing: -0.02em;
		margin: 0;
		animation: fadeIn 0.8s ease-out 0.2s both;
	}

	.spinner-wrapper {
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.spinner {
		width: 64px;
		height: 64px;
		border: 3px solid #E5E5E5;
		border-top-color: #1A1A1A;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.streaming-text {
		font-size: 0.875rem;
		color: #666;
		margin: 0;
		animation: fadeIn 0.8s ease-out 0.5s both;
	}

	.error-text {
		font-size: 1rem;
		color: #DC2626;
		margin: 0;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.fallback-text {
		font-size: 0.875rem;
		color: #999;
		margin: 0;
	}

	.ai-badge {
		font-size: 0.75rem;
		color: #666;
		background: #F5F5F5;
		padding: 0.375rem 0.75rem;
		border-radius: 1rem;
		margin: 0;
		animation: fadeIn 0.8s ease-out 0.5s both;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 768px) {
		.title {
			font-size: 2rem;
		}

		.content {
			gap: 2.5rem;
		}
	}
</style>
