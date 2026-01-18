<!--
	Onboarding Screen 2: Meet OSA
	Full-screen layout matching Wabi iOS design
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	let typedMessage = $state('');
	let showContinue = $state(false);
	let typingComplete = $state(false);

	const fullMessage = "Hi, I'm OSA. I'm an AI that learns about your work and builds personalized apps just for you.";

	onMount(() => {
		// Start typewriter after a short delay
		setTimeout(() => {
			typewriterEffect();
		}, 800);
	});

	function typewriterEffect() {
		let index = 0;
		const interval = setInterval(() => {
			if (index < fullMessage.length) {
				typedMessage += fullMessage[index];
				index++;
			} else {
				clearInterval(interval);
				typingComplete = true;
				// Show continue button after typing is done
				setTimeout(() => {
					showContinue = true;
				}, 500);
			}
		}, 30); // Speed of typing (30ms per character)
	}

	function handleContinue() {
		onboardingStore.nextStep();
		goto('/onboarding/signin');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding');
	}
</script>

<svelte:head>
	<title>Meet OSA - Your AI Assistant</title>
</svelte:head>

<div class="meet-osa-background">
	<div class="meet-osa-screen">
		<div class="content">
			<!-- OSA Icon -->
			<div class="icon-wrapper">
				<img src="/Cloudpngosa.png" alt="OSA" class="cloud-icon" />
			</div>

			<!-- Main Message -->
			<div class="title-row">
				<span class="title-text">Meet</span>
				<img src="/osa-logo.png" alt="OSA" class="osa-logo-inline" />
			</div>

			<!-- Typewriter Message -->
			<div class="message-container">
				<p class="typed-message">
					{typedMessage}{#if !typingComplete}<span class="cursor"></span>{/if}
				</p>
			</div>

			<!-- CTA Buttons -->
			{#if showContinue}
				<div class="cta">
					<PillButton variant="primary" size="lg" onclick={handleContinue}>
						Continue
					</PillButton>
					<button class="back-button" onclick={handleBack}>
						Back
					</button>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.meet-osa-background {
		min-height: 100vh;
		width: 100%;
		background-image: url('/logos/integrations/MIOSABRANDBackround.png');
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
	}

	.meet-osa-screen {
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
		gap: 3.5rem;
		text-align: center;
	}

	.icon-wrapper {
		animation: fadeIn 0.8s ease-out;
	}

	.cloud-icon {
		width: 240px;
		height: auto;
		filter: drop-shadow(0 12px 32px rgba(0, 0, 0, 0.25))
		        drop-shadow(0 6px 16px rgba(0, 0, 0, 0.15));
		animation: float 6s ease-in-out infinite;
	}

	@keyframes float {
		0%, 100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-10px);
		}
	}

	.title-row {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		animation: fadeIn 0.8s ease-out 0.2s both;
	}

	.title-text {
		font-size: 2.75rem;
		font-weight: 700;
		color: #000000;
		line-height: 1.2;
		letter-spacing: -0.02em;
	}

	.osa-logo-inline {
		height: 9rem;
		width: auto;
		filter: drop-shadow(0 8px 24px rgba(0, 0, 0, 0.3))
		        drop-shadow(0 4px 8px rgba(0, 0, 0, 0.2));
	}

	.message-container {
		width: 100%;
		max-width: 500px;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.typed-message {
		font-size: 1.25rem;
		line-height: 1.6;
		color: #1A1A1A;
		font-weight: 500;
		margin: 0;
		min-height: 3em;
		text-shadow: 0 1px 2px rgba(255, 255, 255, 0.5);
	}

	.cursor {
		display: inline-block;
		width: 3px;
		height: 1.2em;
		background-color: #000000;
		margin-left: 2px;
		animation: blink 1s step-end infinite;
		vertical-align: text-bottom;
	}

	@keyframes blink {
		0%, 50% {
			opacity: 1;
		}
		51%, 100% {
			opacity: 0;
		}
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		animation: fadeIn 0.6s ease-out both;
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
		.title-text {
			font-size: 2rem;
		}

		.osa-logo-inline {
			height: 6rem;
		}

		.content {
			gap: 2.5rem;
		}

		.cloud-icon {
			width: 180px;
		}

		.typed-message {
			font-size: 1.125rem;
		}
	}
</style>
