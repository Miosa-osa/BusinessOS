<!--
	Onboarding Screen 4: Connect Gmail (Optional)
	Simplified Gmail connection matching Wabi design
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	function handleConnect() {
		// Initiate Google OAuth flow
		// After OAuth completes, backend will redirect to /onboarding/username
		const backendUrl = 'http://localhost:8001';
		const redirectAfter = '/onboarding/username';
		window.location.href = `${backendUrl}/api/auth/google?redirect=${encodeURIComponent(redirectAfter)}`;
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

<div class="onboarding-background">
	<div class="gmail-screen">
		<div class="content">
			<!-- Main Message -->
			<h1 class="title">
				Connect your<br />Gmail for<br />personalized apps.
			</h1>

			<p class="subtitle">
				OSA analyzes your email to build apps tailored to your needs.
			</p>

			<!-- CTA Buttons -->
			<div class="cta">
				<PillButton variant="primary" size="lg" onclick={handleConnect}>
					Connect Gmail
				</PillButton>

				<div class="secondary-actions">
					<button class="skip-button" onclick={handleSkip}>
						Skip for now
					</button>
					<button class="back-button" onclick={handleBack}>
						Back
					</button>
				</div>
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

	.gmail-screen {
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
		gap: 2rem;
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

	.subtitle {
		font-size: 1.125rem;
		color: #666666;
		margin: 0;
		max-width: 500px;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		animation: fadeIn 0.8s ease-out 0.4s both;
	}

	.secondary-actions {
		display: flex;
		gap: 2rem;
		align-items: center;
	}

	.skip-button,
	.back-button {
		background: transparent;
		border: none;
		color: #666666;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		padding: 0.5rem 1rem;
		font-family: inherit;
		transition: color 0.2s ease;
	}

	.skip-button {
		text-decoration: underline;
	}

	.skip-button:hover,
	.back-button:hover {
		color: #1A1A1A;
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
			gap: 2rem;
		}
	}
</style>
