<!--
	Onboarding Screen 13: Your OS is Ready
	Final celebration screen after viewing all starter apps
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { onMount } from 'svelte';

	let showSuccess = $state(false);
	let showContent = $state(false);
	let animateButton = $state(false);

	onMount(() => {
		// Stagger animations for a smooth entrance
		setTimeout(() => showSuccess = true, 200);
		setTimeout(() => showContent = true, 600);
		setTimeout(() => animateButton = true, 1000);
	});

	async function handleEnterOS() {
		// Mark onboarding as complete
		onboardingStore.complete();

		// TODO: API call to mark onboarding complete in backend
		// await fetch('/api/onboarding/complete', { method: 'POST' });

		// Navigate to main app
		goto('/window');
	}
</script>

<svelte:head>
	<title>Your OS is Ready - OSA Build</title>
</svelte:head>

<GradientBackground variant="ready" fullScreen>
	<div class="ready-screen flex flex-col items-center justify-center min-h-screen px-6 text-center">

		<!-- Success Animation -->
		<div class="success-animation mb-12" class:show={showSuccess}>
			<div class="checkmark-circle">
				<svg class="checkmark" viewBox="0 0 52 52">
					<circle class="checkmark-circle-bg" cx="26" cy="26" r="25" fill="none"/>
					<path class="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/>
				</svg>
			</div>
		</div>

		<!-- Content -->
		<div class="content space-y-6" class:show={showContent}>
			<h1 class="text-6xl font-bold text-gradient">
				Your OS is ready!
			</h1>

			<p class="text-xl text-gray-700 dark:text-gray-300 max-w-2xl mx-auto leading-relaxed">
				We've created 4 personalized apps for you.<br />
				You can build more anytime by asking OSA.
			</p>
		</div>

		<!-- CTA Button -->
		<div class="cta-section mt-16" class:show={animateButton}>
			<PillButton
				variant="primary"
				size="lg"
				onclick={handleEnterOS}
				class="px-12 py-4 text-lg font-semibold shadow-2xl hover:scale-105 transition-transform"
			>
				Enter Your OS
			</PillButton>
		</div>

		<!-- Subtle hint -->
		<p class="mt-8 text-sm text-gray-500 dark:text-gray-400 opacity-0 animate-fade-in-delay">
			Your personalized apps are waiting
		</p>
	</div>
</GradientBackground>

<style>
	/* Success Animation */
	.success-animation {
		opacity: 0;
		transform: scale(0.8);
		transition: all 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	.success-animation.show {
		opacity: 1;
		transform: scale(1);
	}

	.checkmark-circle {
		width: 120px;
		height: 120px;
		position: relative;
	}

	.checkmark {
		width: 120px;
		height: 120px;
		border-radius: 50%;
		display: block;
		stroke-width: 3;
		stroke: #10b981; /* Green for success */
		stroke-miterlimit: 10;
		animation: fill 0.4s ease-in-out 0.4s forwards, scale 0.3s ease-in-out 0.9s both;
	}

	.checkmark-circle-bg {
		stroke-dasharray: 166;
		stroke-dashoffset: 166;
		stroke-width: 3;
		stroke-miterlimit: 10;
		stroke: #10b981;
		fill: none;
		animation: stroke 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
	}

	.checkmark-check {
		transform-origin: 50% 50%;
		stroke-dasharray: 48;
		stroke-dashoffset: 48;
		stroke: #10b981;
		stroke-width: 3;
		animation: stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.8s forwards;
	}

	@keyframes stroke {
		100% {
			stroke-dashoffset: 0;
		}
	}

	@keyframes scale {
		0%, 100% {
			transform: none;
		}
		50% {
			transform: scale3d(1.1, 1.1, 1);
		}
	}

	@keyframes fill {
		100% {
			box-shadow: inset 0px 0px 0px 30px #10b981;
		}
	}

	/* Content animations */
	.content {
		opacity: 0;
		transform: translateY(20px);
		transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.content.show {
		opacity: 1;
		transform: translateY(0);
	}

	.cta-section {
		opacity: 0;
		transform: translateY(20px);
		transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.cta-section.show {
		opacity: 1;
		transform: translateY(0);
	}

	/* Text gradient */
	.text-gradient {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	/* Delayed fade in for hint */
	@keyframes fade-in-delay {
		0% {
			opacity: 0;
			transform: translateY(10px);
		}
		100% {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-fade-in-delay {
		animation: fade-in-delay 1s ease-out 1.5s forwards;
	}
</style>
