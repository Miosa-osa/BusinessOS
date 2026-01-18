<!--
	OSA Components Demo Page
	Test all iOS-inspired components
-->
<script lang="ts">
	import {
		GradientBackground,
		PillButton,
		RoundedInput,
		AppCard,
		ProgressDots,
		GlassCard
	} from '$lib/components/osa';

	let email = $state('');
	let testValue = $state('');
	let currentStep = $state(0);
	const totalSteps = 4;

	function handleNext() {
		currentStep = Math.min(currentStep + 1, totalSteps - 1);
	}

	function handlePrev() {
		currentStep = Math.max(currentStep - 1, 0);
	}
</script>

<svelte:head>
	<title>OSA Components Demo</title>
</svelte:head>

<GradientBackground variant="welcome">
	<div class="container mx-auto px-4 py-16">
		<div class="max-w-4xl mx-auto space-y-16">
			<!-- Header -->
			<div class="text-center space-y-4 animate-slide-down">
				<h1 class="text-5xl font-bold text-gradient">
					OSA Build Components
				</h1>
				<p class="text-xl text-gray-600 dark:text-gray-400">
					iOS-inspired design system
				</p>
			</div>

			<!-- Progress Dots -->
			<div class="text-center space-y-4">
				<h2 class="text-2xl font-semibold">Progress Dots</h2>
				<ProgressDots total={totalSteps} current={currentStep} />
				<div class="flex gap-2 justify-center">
					<PillButton variant="secondary" size="sm" onclick={handlePrev} disabled={currentStep === 0}>
						Previous
					</PillButton>
					<PillButton variant="secondary" size="sm" onclick={handleNext} disabled={currentStep === totalSteps - 1}>
						Next
					</PillButton>
				</div>
			</div>

			<!-- Buttons -->
			<div class="space-y-4">
				<h2 class="text-2xl font-semibold text-center">Pill Buttons</h2>
				<div class="flex flex-wrap gap-4 justify-center">
					<PillButton variant="primary">
						Primary Button
					</PillButton>
					<PillButton variant="secondary">
						Secondary Button
					</PillButton>
					<PillButton variant="ghost">
						Ghost Button
					</PillButton>
					<PillButton variant="primary" size="sm">
						Small Button
					</PillButton>
					<PillButton variant="primary" size="lg">
						Large Button
					</PillButton>
					<PillButton variant="primary" loading>
						Loading...
					</PillButton>
					<PillButton variant="primary" disabled>
						Disabled
					</PillButton>
				</div>
			</div>

			<!-- Inputs -->
			<div class="space-y-4 max-w-md mx-auto">
				<h2 class="text-2xl font-semibold text-center">Rounded Inputs</h2>
				<RoundedInput
					label="Email"
					type="email"
					bind:value={email}
					placeholder="you@company.com"
					helperText="We'll never share your email"
				/>
				<RoundedInput
					label="Username"
					type="text"
					bind:value={testValue}
					placeholder="bekorains"
					required
				/>
				<RoundedInput
					label="Error Example"
					type="text"
					value="invalid"
					error="This field has an error"
				/>
			</div>

			<!-- Glass Cards -->
			<div class="space-y-4">
				<h2 class="text-2xl font-semibold text-center">Glass Cards</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<GlassCard padding="lg" hoverable>
						<h3 class="text-xl font-semibold mb-2">Hoverable Card</h3>
						<p class="text-gray-600 dark:text-gray-400">
							This card has glassmorphism effect and hover animation.
						</p>
					</GlassCard>
					<GlassCard padding="md">
						<h3 class="text-xl font-semibold mb-2">Static Card</h3>
						<p class="text-gray-600 dark:text-gray-400">
							Beautiful glass effect without hover.
						</p>
					</GlassCard>
				</div>
			</div>

			<!-- App Cards -->
			<div class="space-y-4">
				<h2 class="text-2xl font-semibold text-center">App Cards</h2>
				<div class="grid grid-cols-2 md:grid-cols-4 gap-6">
					<AppCard
						title="Book Finds"
						description="Discover books"
						usagePercentage={98}
					/>
					<AppCard
						title="Motion Design"
						description="Creative tools"
						usagePercentage={76}
					/>
					<AppCard
						title="Feedback Hub"
						description="Community"
					/>
					<AppCard
						title="SF Founder"
						description="Networking"
						usagePercentage={42}
					/>
				</div>
			</div>

			<!-- Gradient Showcase -->
			<div class="space-y-4">
				<h2 class="text-2xl font-semibold text-center">All Gradients</h2>
				<div class="grid grid-cols-2 md:grid-cols-3 gap-4">
					{#each ['welcome', 'signin', 'personalization', 'apps-showcase', 'ready', 'builder'] as variant}
						<div class="gradient-{variant} p-8 rounded-xl text-center font-semibold">
							{variant}
						</div>
					{/each}
				</div>
			</div>

			<!-- Back to app -->
			<div class="text-center">
				<a href="/window">
					<PillButton variant="secondary">
						← Back to App
					</PillButton>
				</a>
			</div>
		</div>
	</div>
</GradientBackground>
