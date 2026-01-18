<!--
	Onboarding Screen 2: Meet OSA
	Introduces the AI agent and its capabilities
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton, GlassCard } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

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

<GradientBackground variant="signin" fullScreen>
	<div class="meet-osa-screen text-center space-y-12 animate-slide-up">
		<!-- OSA Avatar / Icon -->
		<div class="space-y-6">
			<div class="osa-avatar mx-auto">
				<!-- Animated OSA orb - glowing gradient sphere -->
				<div class="relative w-48 h-48 mx-auto">
					<div class="absolute inset-0 rounded-full bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 pulse-glow"></div>
					<div class="absolute inset-4 rounded-full bg-gradient-to-br from-violet-300 via-purple-400 to-indigo-500 opacity-60 animate-pulse"></div>
					<div class="absolute inset-8 rounded-full bg-gradient-to-br from-white via-purple-200 to-indigo-200 opacity-40"></div>
				</div>
			</div>

			<div class="space-y-4">
				<h1 class="text-5xl font-bold text-gradient">
					Meet OSA
				</h1>
				<p class="text-xl text-gray-700 dark:text-gray-300 max-w-2xl mx-auto">
					Your Operating System Agent - an AI that understands you and builds what you need
				</p>
			</div>
		</div>

		<!-- What OSA Can Do -->
		<div class="capabilities-section space-y-6">
			<h2 class="text-2xl font-semibold text-gray-800 dark:text-gray-200">
				What OSA Can Do
			</h2>

			<div class="capabilities-grid grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl mx-auto">
				<GlassCard padding="lg" hoverable>
					<div class="space-y-3 text-left">
						<div class="text-3xl">🧠</div>
						<h3 class="text-lg font-semibold">Understand Your Needs</h3>
						<p class="text-gray-600 dark:text-gray-400 text-sm">
							OSA learns from your Gmail, calendar, and usage patterns to know exactly what you need
						</p>
					</div>
				</GlassCard>

				<GlassCard padding="lg" hoverable>
					<div class="space-y-3 text-left">
						<div class="text-3xl">⚡</div>
						<h3 class="text-lg font-semibold">Build Apps Instantly</h3>
						<p class="text-gray-600 dark:text-gray-400 text-sm">
							From idea to working app in seconds. No code, no complexity, just conversation
						</p>
					</div>
				</GlassCard>

				<GlassCard padding="lg" hoverable>
					<div class="space-y-3 text-left">
						<div class="text-3xl">🎯</div>
						<h3 class="text-lg font-semibold">Personalized Experience</h3>
						<p class="text-gray-600 dark:text-gray-400 text-sm">
							Every app is tailored to your workflow, preferences, and goals
						</p>
					</div>
				</GlassCard>

				<GlassCard padding="lg" hoverable>
					<div class="space-y-3 text-left">
						<div class="text-3xl">🔄</div>
						<h3 class="text-lg font-semibold">Continuously Evolving</h3>
						<p class="text-gray-600 dark:text-gray-400 text-sm">
							OSA keeps learning and improving your OS based on how you use it
						</p>
					</div>
				</GlassCard>
			</div>
		</div>

		<!-- CTA Buttons -->
		<div class="cta-section flex gap-4 justify-center">
			<PillButton variant="ghost" size="md" onclick={handleBack}>
				Back
			</PillButton>
			<PillButton variant="primary" size="lg" onclick={handleContinue}>
				Continue
			</PillButton>
		</div>
	</div>
</GradientBackground>

<style>
	.meet-osa-screen {
		padding: 2rem 0;
		max-width: 1200px;
		margin: 0 auto;
	}

	.osa-avatar {
		animation: fade-in 0.8s ease-out;
	}
</style>
