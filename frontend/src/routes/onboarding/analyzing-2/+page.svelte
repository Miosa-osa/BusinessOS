<!--
	Onboarding Screen 7: AI Analysis (Second Insight)
	Simplified analyzing screen matching Wabi design
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { analyzingInsights } from '$lib/stores/onboardingAnalysis';

	let insightMessage = '';

	$: insightMessage = $analyzingInsights.message2;

	onMount(() => {
		setTimeout(() => {
			onboardingStore.nextStep();
			goto('/onboarding/analyzing-3');
		}, 2000);
	});
</script>

<svelte:head>
	<title>Analyzing - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="analyzing-screen">
		<div class="content">
			<h1 class="title">
				{insightMessage}
			</h1>

			<!-- Loading Spinner -->
			<div class="spinner-wrapper">
				<div class="spinner"></div>
			</div>
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
