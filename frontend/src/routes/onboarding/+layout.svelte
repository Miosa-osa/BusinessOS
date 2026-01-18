<!--
	Onboarding Layout
	Wraps all onboarding screens with consistent structure
-->
<script lang="ts">
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { ProgressDots } from '$lib/components/osa';

	const store = $derived(onboardingStore);
</script>

<div class="onboarding-layout min-h-screen flex flex-col">
	<!-- Progress indicator - fixed at top -->
	<div class="progress-container py-8">
		<ProgressDots total={$store.totalSteps} current={$store.currentStep} />
	</div>

	<!-- Main content area -->
	<div class="flex-1 flex items-center justify-center px-4">
		<div class="onboarding-content w-full max-w-2xl">
			<slot />
		</div>
	</div>
</div>

<style>
	.onboarding-layout {
		position: relative;
	}

	.progress-container {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		display: flex;
		justify-content: center;
		background: transparent;
		z-index: 100;
	}

	.onboarding-content {
		animation: fade-in 0.5s ease-out;
	}
</style>
