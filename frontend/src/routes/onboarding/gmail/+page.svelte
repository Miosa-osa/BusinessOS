<!--
	Onboarding Screen 4: Connect Gmail (Optional)
	Simplified Gmail connection matching Wabi design
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { browser } from '$app/environment';

	let isLoading = $state(false);
	let errorMessage = $state<string | null>(null);

	onMount(() => {
		if (!browser) return;

		// Check for OAuth callback success/error in URL params
		const urlParams = new URLSearchParams(window.location.search);
		const success = urlParams.get('success');
		const error = urlParams.get('error');

		if (success === 'true') {
			// Gmail connected successfully
			onboardingStore.setUserData({ gmailConnected: true });
			onboardingStore.nextStep();
			goto('/onboarding/username');
		} else if (error) {
			// OAuth failed
			errorMessage = decodeURIComponent(error);
			isLoading = false;
		}
	});

	function handleConnect() {
		if (isLoading) return;

		isLoading = true;
		errorMessage = null;

		// Store return URL in sessionStorage for after OAuth completes
		if (browser) {
			sessionStorage.setItem('oauth_return', '/onboarding/gmail?success=true');
		}

		// Redirect to backend Google OAuth endpoint
		// Backend will handle the OAuth flow and redirect back
		const backendURL = import.meta.env.VITE_API_URL || 'http://localhost:8001';
		const redirectURL = `/onboarding/gmail`;

		// Redirect to Google OAuth
		window.location.href = `${backendURL}/api/auth/google?redirect=${encodeURIComponent(redirectURL)}`;
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

<GradientBackground>
	<div class="gmail-screen">
		<div class="content">
			<!-- Main Message -->
			<h1 class="title">
				Connect your<br />Gmail for<br />personalized apps.
			</h1>

			<p class="subtitle">
				OSA analyzes your email to build apps tailored to your needs.
			</p>

			<!-- Error Message -->
			{#if errorMessage}
				<div class="error-message">
					<p>{errorMessage}</p>
				</div>
			{/if}

			<!-- CTA Buttons -->
			<div class="cta">
				<PillButton
					variant="primary"
					size="lg"
					onclick={handleConnect}
					disabled={isLoading}
				>
					{#if isLoading}
						Connecting...
					{:else}
						Connect Gmail
					{/if}
				</PillButton>

				<div class="secondary-actions">
					<button class="btn-pill btn-pill-ghost skip-button" onclick={handleSkip} disabled={isLoading}>
						Skip for now
					</button>
					<button class="btn-pill btn-pill-ghost back-button" onclick={handleBack} disabled={isLoading}>
						Back
					</button>
				</div>
			</div>
		</div>
	</div>
</GradientBackground>

<style>
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

	.error-message {
		background: #fee;
		border: 1px solid #fcc;
		border-radius: 8px;
		padding: 1rem;
		max-width: 500px;
		animation: fadeIn 0.3s ease-out;
	}

	.error-message p {
		color: #c33;
		font-size: 0.875rem;
		margin: 0;
		text-align: center;
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

	.skip-button:disabled,
	.back-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
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
