<!--
	Onboarding Screen 1: Welcome to OSA Build
	Full-screen layout matching Wabi iOS design
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	function handleGetStarted() {
		onboardingStore.nextStep();
		goto('/onboarding/quick-info');
	}
</script>

<svelte:head>
	<title>Welcome to OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="welcome-screen">
		<div class="content">
			<!-- Cloud Icon -->
			<div class="cloud-wrapper">
				<img src="/Cloudpngosa.png" alt="OSA Cloud" class="cloud-image" />
			</div>

			<!-- Main Message -->
			<h1 class="title">
				A new era of<br />business software<br />is here.
			</h1>

			<!-- OSA Introduction -->
			<p class="osa-intro">
				Meet OSA, your AI assistant that builds personalized apps based on how you work.
			</p>

			<!-- CTA Button -->
			<div class="cta">
				<PillButton variant="primary" size="lg" onclick={handleGetStarted}>
					Get Started
				</PillButton>
				<p class="helper-text">Takes less than 2 minutes</p>
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

	.welcome-screen {
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
		gap: 4rem;
		text-align: center;
	}

	.cloud-wrapper {
		animation: fadeIn 0.8s ease-out;
	}

	.cloud-image {
		width: 220px;
		height: auto;
		display: block;
		animation: float 6s ease-in-out infinite;
		filter: drop-shadow(0 12px 32px rgba(0, 0, 0, 0.2));
	}

	@keyframes float {
		0%, 100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-10px);
		}
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

	.osa-intro {
		font-size: 1.125rem;
		color: #666666;
		line-height: 1.6;
		max-width: 400px;
		margin: -2rem 0 0;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		animation: fadeIn 0.8s ease-out 0.4s both;
	}

	.helper-text {
		font-size: 0.875rem;
		color: #666666;
		margin: 0;
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
			gap: 3rem;
		}
	}
</style>
