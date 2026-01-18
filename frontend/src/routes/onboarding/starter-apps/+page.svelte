<!--
	Onboarding Screens 9-12: Starter Apps Showcase
	Carousel showing 4 personalized apps (1 at a time)
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { GradientBackground, ProgressDots, PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	// Get onboarding state
	let store = $state($onboardingStore);
	let starterApps = $derived(store.userData.starterApps || []);

	// Carousel state
	let currentAppIndex = $state(0);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// Current app being displayed
	const currentApp = $derived(starterApps[currentAppIndex]);
	const isFirstApp = $derived(currentAppIndex === 0);
	const isLastApp = $derived(currentAppIndex === starterApps.length - 1);

	// Has user viewed all apps?
	let hasViewedAll = $state(false);

	// Track which apps have been viewed
	let viewedApps = $state<Set<number>>(new Set([0]));

	// Navigation functions
	function goToNextApp() {
		if (currentAppIndex < starterApps.length - 1) {
			currentAppIndex++;
			viewedApps.add(currentAppIndex);

			// Check if all apps have been viewed
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

	// Keyboard navigation
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowLeft') {
			goToPrevApp();
		} else if (e.key === 'ArrowRight') {
			goToNextApp();
		}
	}

	// Load apps if not already loaded
	async function loadApps() {
		if (starterApps.length > 0) {
			return; // Apps already loaded
		}

		isLoading = true;
		error = null;

		try {
			// Try to get apps from backend
			// For now, we'll use mock data since analyzing page already sets them
			// But in a real scenario, this would call the API:
			// const response = await osaOnboardingApi.generateStarterApps(workspaceId, analysis);

			// If no apps in store, set mock data
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
			error = err instanceof Error ? err.message : 'Failed to load apps';
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadApps();

		// Subscribe to store updates
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

<GradientBackground variant="apps-showcase" fullScreen>
	<div class="starter-apps-screen flex flex-col items-center justify-center min-h-screen px-6 py-8">
		<!-- Progress dots -->
		<div class="absolute top-8 left-1/2 -translate-x-1/2">
			<ProgressDots total={13} current={9 + currentAppIndex} />
		</div>

		{#if isLoading}
			<!-- Loading state -->
			<div class="flex flex-col items-center gap-6 animate-slide-up">
				<div class="w-16 h-16 border-4 border-purple-300 border-t-purple-600 rounded-full animate-spin"></div>
				<p class="text-xl text-gray-700 dark:text-gray-300">Loading your personalized apps...</p>
			</div>
		{:else if error}
			<!-- Error state -->
			<div class="flex flex-col items-center gap-6 animate-slide-up max-w-md">
				<div class="text-red-500 text-5xl">⚠️</div>
				<h2 class="text-2xl font-bold text-gray-800 dark:text-gray-200">Oops! Something went wrong</h2>
				<p class="text-gray-600 dark:text-gray-400 text-center">{error}</p>
				<PillButton variant="primary" onclick={loadApps}>
					Try Again
				</PillButton>
			</div>
		{:else if starterApps.length === 0}
			<!-- No apps state -->
			<div class="flex flex-col items-center gap-6 animate-slide-up">
				<p class="text-xl text-gray-700 dark:text-gray-300">No apps available yet.</p>
				<PillButton variant="primary" onclick={handleContinue}>
					Continue
				</PillButton>
			</div>
		{:else}
			<!-- Apps carousel -->
			<div class="w-full max-w-2xl space-y-8 animate-slide-up">
				<!-- Navigation indicator -->
				<div class="flex items-center justify-center gap-4 text-gray-600 dark:text-gray-400">
					<button
						onclick={goToPrevApp}
						disabled={isFirstApp}
						class="p-2 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
						aria-label="Previous app"
					>
						<ChevronLeft class="w-6 h-6" />
					</button>

					<span class="text-lg font-medium min-w-[80px] text-center">
						{currentAppIndex + 1} of {starterApps.length}
					</span>

					<button
						onclick={goToNextApp}
						disabled={isLastApp}
						class="p-2 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
						aria-label="Next app"
					>
						<ChevronRight class="w-6 h-6" />
					</button>
				</div>

				<!-- Title -->
				<h1 class="text-4xl md:text-5xl font-bold text-gradient text-center">
					Here are the apps we built for you
				</h1>

				<!-- App card display -->
				<div class="relative min-h-[400px] flex items-center justify-center">
					{#key currentAppIndex}
						<div
							class="w-full max-w-md"
							in:fly={{ x: 100, duration: 400, delay: 100 }}
							out:fly={{ x: -100, duration: 300 }}
						>
							<div class="space-y-6">
								<!-- App Card -->
								<div class="app-card-wrapper flex flex-col items-center gap-6 p-8 bg-white/40 dark:bg-gray-800/40 rounded-3xl backdrop-blur-sm border border-white/50 dark:border-gray-700/50 shadow-xl">
									<!-- Circular icon (80px) -->
									<div class="w-20 h-20 rounded-full bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 flex items-center justify-center shadow-lg">
										{#if currentApp.iconUrl}
											<img src={currentApp.iconUrl} alt={currentApp.title} class="w-20 h-20 rounded-full object-cover" />
										{:else}
											<span class="text-white text-3xl font-bold">
												{currentApp.title.charAt(0)}
											</span>
										{/if}
									</div>

									<!-- App title -->
									<h2 class="text-2xl md:text-3xl font-bold text-gray-800 dark:text-gray-100 text-center">
										{currentApp.title}
									</h2>

									<!-- App description -->
									<p class="text-lg text-gray-700 dark:text-gray-300 text-center">
										{currentApp.description}
									</p>

									<!-- Reasoning -->
									<div class="w-full pt-4 border-t border-gray-300 dark:border-gray-600">
										<p class="text-base text-gray-600 dark:text-gray-400 italic text-center">
											{currentApp.reason}
										</p>
									</div>
								</div>
							</div>
						</div>
					{/key}
				</div>

				<!-- Navigation buttons -->
				<div class="flex items-center justify-between gap-4 pt-4">
					<!-- Previous button -->
					{#if !isFirstApp}
						<button
							onclick={goToPrevApp}
							class="px-6 py-3 rounded-full bg-white/60 dark:bg-gray-800/60 hover:bg-white/80 dark:hover:bg-gray-800/80 text-gray-700 dark:text-gray-300 font-medium transition-all shadow-md border border-gray-300 dark:border-gray-600"
						>
							Previous
						</button>
					{:else}
						<div></div>
					{/if}

					<!-- Next/Continue button -->
					{#if isLastApp && hasViewedAll}
						<PillButton variant="primary" onclick={handleContinue} class="ml-auto">
							Continue to Your OS
						</PillButton>
					{:else if !isLastApp}
						<button
							onclick={goToNextApp}
							class="px-6 py-3 rounded-full bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white font-medium transition-all shadow-lg ml-auto"
						>
							Next
						</button>
					{/if}
				</div>

				<!-- Hint text -->
				{#if !hasViewedAll}
					<p class="text-sm text-gray-500 dark:text-gray-400 text-center italic">
						View all apps to continue
					</p>
				{/if}
			</div>
		{/if}
	</div>
</GradientBackground>

<style>
	.starter-apps-screen {
		animation: fade-in 0.6s ease-out;
	}

	.app-card-wrapper {
		animation: float 3s ease-in-out infinite;
	}

	@keyframes float {
		0%, 100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-5px);
		}
	}

	@keyframes fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	:global(.text-gradient) {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	:global(.animate-slide-up) {
		animation: slide-up 0.8s ease-out;
	}

	@keyframes slide-up {
		from {
			opacity: 0;
			transform: translateY(30px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
