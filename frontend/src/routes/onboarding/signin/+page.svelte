<!--
	Onboarding Screen 3: Sign In
	Simplified authentication step matching Wabi design
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	let isLoading = $state(false);
	let error = $state<string | null>(null);

	function handleGoogleSignIn() {
		isLoading = true;
		error = null;

		try {
			// Call backend OAuth directly with FULL frontend URL redirect
			// This requests full Gmail scopes for AI analysis
			const backendUrl = 'http://localhost:8001';
			const frontendUrl = window.location.origin; // http://localhost:5174
			const redirectAfter = `${frontendUrl}/onboarding/username`;

			console.log('[Signin] Redirecting to Google OAuth with redirect:', redirectAfter);
			window.location.href = `${backendUrl}/api/auth/google?redirect=${encodeURIComponent(redirectAfter)}`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to sign in';
			console.error('Sign in error:', err);
			isLoading = false;
		}
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/meet-osa');
	}
</script>

<svelte:head>
	<title>Sign In - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="signin-screen">
		<!-- Decorative Bubbles -->
		<div class="bubble bubble-1"></div>
		<div class="bubble bubble-2"></div>
		<div class="bubble bubble-3"></div>
		<div class="bubble bubble-4"></div>

		<div class="content">
			<!-- Gmail Icon -->
			<div class="icon-wrapper">
				<svg class="gmail-icon" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
					<path fill="#4285f4" d="M45,16.2l-5,2.75l-5,4.75L35,40h7c1.657,0,3-1.343,3-3V16.2z"/>
					<path fill="#34a853" d="M3,16.2l3.614,1.71L13,23.7V40H6c-1.657,0-3-1.343-3-3V16.2z"/>
					<polygon fill="#fbbc04" points="35,11.2 24,19.45 13,11.2 12,17 13,23.7 24,31.95 35,23.7 36,17"/>
					<path fill="#ea4335" d="M3,12.298V16.2l10,7.5V11.2L9.876,8.859C9.132,8.301,8.228,8,7.298,8h0C4.924,8,3,9.924,3,12.298z"/>
					<path fill="#c5221f" d="M45,12.298V16.2l-10,7.5V11.2l3.124-2.341C38.868,8.301,39.772,8,40.702,8h0 C43.076,8,45,9.924,45,12.298z"/>
				</svg>
			</div>

			<!-- Main Message -->
			<div class="text-content">
				<h1 class="title">
					Get your first AI apps
				</h1>
				<p class="subtitle">
					OSA analyzes your Gmail to understand your work and builds personalized apps just for you.
				</p>
			</div>

			<!-- Sign In Button -->
			<div class="cta">
				{#if error}
					<p class="error-text">{error}</p>
				{/if}

				<PillButton
					variant="primary"
					size="lg"
					onclick={handleGoogleSignIn}
					loading={isLoading}
				>
					{isLoading ? 'Connecting...' : 'Connect with Google'}
				</PillButton>

				<p class="privacy-text">
					Your data stays private and is only used to personalize your apps.
				</p>

				<button class="back-button" onclick={handleBack}>
					Back
				</button>
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

	.signin-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		position: relative;
		overflow: hidden;
	}

	/* Decorative Bubbles with Color Morphing */
	.bubble {
		position: absolute;
		border-radius: 50%;
		opacity: 0.6;
		filter: blur(40px);
		pointer-events: none;
		animation: float 12s ease-in-out infinite, colorMorph 15s ease-in-out infinite;
		transition: all 1.5s ease-in-out;
	}

	.bubble-1 {
		width: 300px;
		height: 300px;
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.4), rgba(168, 85, 247, 0.4));
		top: 10%;
		left: 15%;
		animation-delay: 0s, 0s;
	}

	.bubble-2 {
		width: 250px;
		height: 250px;
		background: linear-gradient(135deg, rgba(236, 72, 153, 0.3), rgba(239, 68, 68, 0.3));
		top: 15%;
		right: 10%;
		animation-delay: 3s, 5s;
	}

	.bubble-3 {
		width: 200px;
		height: 200px;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.3), rgba(147, 51, 234, 0.3));
		bottom: 20%;
		left: 10%;
		animation-delay: 6s, 10s;
	}

	.bubble-4 {
		width: 280px;
		height: 280px;
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.3), rgba(236, 72, 153, 0.3));
		bottom: 15%;
		right: 15%;
		animation-delay: 9s, 2s;
	}

	@keyframes float {
		0%, 100% {
			transform: translate(0, 0) scale(1) rotate(0deg);
		}
		20% {
			transform: translate(25px, -25px) scale(1.08) rotate(5deg);
		}
		40% {
			transform: translate(-20px, 20px) scale(0.92) rotate(-3deg);
		}
		60% {
			transform: translate(30px, 15px) scale(1.05) rotate(4deg);
		}
		80% {
			transform: translate(-15px, -20px) scale(0.95) rotate(-2deg);
		}
	}

	@keyframes colorMorph {
		0% {
			filter: blur(40px) hue-rotate(0deg) brightness(1);
		}
		25% {
			filter: blur(45px) hue-rotate(30deg) brightness(1.1);
		}
		50% {
			filter: blur(50px) hue-rotate(60deg) brightness(0.95);
		}
		75% {
			filter: blur(42px) hue-rotate(40deg) brightness(1.05);
		}
		100% {
			filter: blur(40px) hue-rotate(0deg) brightness(1);
		}
	}

	.content {
		width: 100%;
		max-width: 600px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2.5rem;
		text-align: center;
		position: relative;
		z-index: 10;
	}

	.icon-wrapper {
		animation: fadeIn 0.8s ease-out 0.1s both;
	}

	.gmail-icon {
		width: 120px;
		height: 120px;
		filter: drop-shadow(0 8px 24px rgba(0, 0, 0, 0.1));
	}

	.text-content {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		animation: fadeIn 0.8s ease-out 0.2s both;
	}

	.title {
		font-size: 2.75rem;
		font-weight: 700;
		color: #1A1A1A;
		line-height: 1.2;
		letter-spacing: -0.02em;
		margin: 0;
	}

	.subtitle {
		font-size: 1.125rem;
		line-height: 1.6;
		color: #666666;
		margin: 0;
		max-width: 480px;
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		animation: fadeIn 0.8s ease-out 0.3s both;
		width: 100%;
	}

	.error-text {
		font-size: 0.875rem;
		color: #DC2626;
		margin: 0;
	}

	.privacy-text {
		font-size: 0.875rem;
		color: #999999;
		margin: 0;
		max-width: 340px;
		line-height: 1.5;
	}

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

		.subtitle {
			font-size: 1rem;
		}

		.gmail-icon {
			width: 100px;
			height: 100px;
		}

		.content {
			gap: 2rem;
		}

		.bubble {
			opacity: 0.4;
		}

		.bubble-1,
		.bubble-2,
		.bubble-3,
		.bubble-4 {
			width: 200px;
			height: 200px;
		}
	}
</style>
