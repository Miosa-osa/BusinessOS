<!--
	Onboarding Screens 9-12: Starter Apps Showcase
	Simplified carousel showing 4 personalized apps matching Wabi design
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	let store = $state($onboardingStore);
	let starterApps = $derived(store.userData.starterApps || []);

	let currentAppIndex = $state(0);
	let isLoading = $state(false);

	const currentApp = $derived(starterApps[currentAppIndex]);
	const isFirstApp = $derived(currentAppIndex === 0);
	const isLastApp = $derived(currentAppIndex === starterApps.length - 1);

	let hasViewedAll = $state(false);
	let viewedApps = $state<Set<number>>(new Set([0]));

	function goToNextApp() {
		if (currentAppIndex < starterApps.length - 1) {
			currentAppIndex++;
			viewedApps.add(currentAppIndex);

			if (viewedApps.size === starterApps.length) {
				hasViewedAll = true;
			}
		}
	}

	function goToPrevApp() {
		if (currentAppIndex > 0) {
			currentAppIndex--;
			viewedApps.add(currentAppIndex);
		}
	}

	function handleContinue() {
		onboardingStore.nextStep();
		goto('/onboarding/ready');
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowLeft') {
			goToPrevApp();
		} else if (e.key === 'ArrowRight') {
			goToNextApp();
		}
	}

	async function loadApps() {
		if (starterApps.length > 0) {
			return;
		}

		isLoading = true;

		try {
			if (starterApps.length === 0) {
				onboardingStore.setStarterApps([
					{
						id: '1',
						title: 'No-Code Book Finds',
						description: 'Discover books about no-code tools and platforms',
						reason: 'Because you follow Bubble, Framer, and Synthflow'
					},
					{
						id: '2',
						title: 'Motion Design Muse',
						description: 'Inspiration for motion graphics and animations',
						reason: 'Based on your design tool usage'
					},
					{
						id: '3',
						title: 'Feature Feedback Hub',
						description: 'Collect and prioritize user feedback',
						reason: 'For building with your community'
					},
					{
						id: '4',
						title: 'SF Founder Weekend',
						description: 'Connect with founders at weekend events',
						reason: 'Based on your location and interests'
					}
				]);
			}
		} catch (err) {
			console.error('Error loading starter apps:', err);
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadApps();

		const unsubscribe = onboardingStore.subscribe((value) => {
			store = value;
		});

		return unsubscribe;
	});
</script>

<svelte:head>
	<title>Your Starter Apps - OSA Build</title>
</svelte:head>

<svelte:window on:keydown={handleKeydown} />

<div class="onboarding-background">
	<div class="starter-apps-screen">
		<div class="content">
			<h1 class="title">
				Your starter apps.
			</h1>

			{#if isLoading}
				<div class="spinner-wrapper">
					<div class="spinner"></div>
				</div>
			{:else if starterApps.length === 0}
				<p class="subtitle">No apps available yet.</p>
			{:else}
				<!-- Carousel -->
				<div class="carousel">
					<!-- Navigation indicator -->
					<div class="nav-indicator">
						<button
							onclick={goToPrevApp}
							disabled={isFirstApp}
							class="nav-button"
							class:disabled={isFirstApp}
							aria-label="Previous app"
						>
							<ChevronLeft size={24} />
						</button>

						<span class="page-number">
							{currentAppIndex + 1} of {starterApps.length}
						</span>

						<button
							onclick={goToNextApp}
							disabled={isLastApp}
							class="nav-button"
							class:disabled={isLastApp}
							aria-label="Next app"
						>
							<ChevronRight size={24} />
						</button>
					</div>

					<!-- App card -->
					<div class="app-display">
						{#key currentAppIndex}
							<div
								class="app-card"
								in:fly={{ x: 300, duration: 500, opacity: 0 }}
								out:fly={{ x: -300, duration: 500, opacity: 0 }}
							>
								<div class="app-icon">
									{currentApp.title.charAt(0)}
								</div>

								<h2 class="app-title">
									{currentApp.title}
								</h2>

								<p class="app-description">
									{currentApp.description}
								</p>

								<p class="app-reason">
									{currentApp.reason}
								</p>
							</div>
						{/key}
					</div>

					<!-- Continue button -->
					{#if hasViewedAll}
						<div class="cta">
							<PillButton variant="primary" size="lg" onclick={handleContinue}>
								Continue
							</PillButton>
						</div>
					{:else}
						<p class="hint">View all apps to continue</p>
					{/if}
				</div>
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

	.starter-apps-screen {
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

	.subtitle {
		font-size: 1.125rem;
		color: #666666;
		margin: 0;
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

	.carousel {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.nav-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1.5rem;
	}

	.nav-button {
		background: white;
		border: 2px solid #E5E5E5;
		border-radius: 50%;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s ease;
		color: #1A1A1A;
	}

	.nav-button:hover:not(.disabled) {
		border-color: #1A1A1A;
		background: #F5F5F5;
	}

	.nav-button.disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.page-number {
		font-size: 1rem;
		font-weight: 500;
		color: #1A1A1A;
		min-width: 80px;
	}

	.app-display {
		position: relative;
		min-height: 450px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden; /* Prevent horizontal scrollbar during slide */
	}

	.app-card {
		position: absolute;
		width: 100%;
		max-width: 600px;
		padding: 2rem;
		background: white;
		border-radius: 1.5rem;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1.5rem;
		transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.5s ease;
	}

	.app-icon {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: linear-gradient(135deg, #E5E5E5, #D1D1D1);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 700;
		color: #666666;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.app-title {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1A1A1A;
		margin: 0;
	}

	.app-description {
		font-size: 1.125rem;
		color: #666666;
		margin: 0;
	}

	.app-reason {
		font-size: 0.9375rem;
		color: #999999;
		font-style: italic;
		margin: 0;
		padding-top: 0.75rem;
		border-top: 1px solid #E5E5E5;
	}

	.cta {
		animation: fadeIn 0.8s ease-out 0.4s both;
	}

	.hint {
		font-size: 0.875rem;
		color: #999999;
		margin: 0;
		font-style: italic;
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

		.app-card {
			padding: 1.5rem;
		}

		.app-icon {
			width: 64px;
			height: 64px;
			font-size: 1.5rem;
		}

		.app-title {
			font-size: 1.5rem;
		}

		.app-description {
			font-size: 1rem;
		}
	}
</style>
