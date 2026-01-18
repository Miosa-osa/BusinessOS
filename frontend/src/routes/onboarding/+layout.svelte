<!--
	Onboarding Layout
	Minimal wrapper - only handles progress dots
	Each screen handles its own full-screen layout
-->
<script lang="ts">
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { ProgressDots } from '$lib/components/osa';

	const store = $derived(onboardingStore);
</script>

<!-- Progress indicator - fixed at top -->
<div class="progress-container">
	<ProgressDots total={$store.totalSteps} current={$store.currentStep} />
</div>

<!-- Full-screen content - no constraints -->
<slot />

<style>
	.progress-container {
		position: fixed;
		top: 2rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 100;
	}
</style>
