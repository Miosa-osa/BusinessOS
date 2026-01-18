<!--
	Onboarding Screen 5: Claim Username
	User chooses their unique username
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton, RoundedInput } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';

	let username = $state('');
	let isChecking = $state(false);
	let isAvailable = $state<boolean | null>(null);
	let error = $state('');

	async function checkAvailability() {
		if (!username || username.length < 3) {
			error = 'Username must be at least 3 characters';
			isAvailable = false;
			return;
		}

		isChecking = true;
		error = '';

		// TODO: Implement actual availability check API call
		await new Promise(resolve => setTimeout(resolve, 500));

		// Mock check
		const taken = ['admin', 'osa', 'test'];
		isAvailable = !taken.includes(username.toLowerCase());
		isChecking = false;

		if (!isAvailable) {
			error = 'Username is already taken';
		}
	}

	function handleUsernameChange(event: Event) {
		const target = event.target as HTMLInputElement;
		username = target.value;
		isAvailable = null;
		error = '';
	}

	function handleContinue() {
		if (!isAvailable) {
			error = 'Please choose an available username';
			return;
		}

		onboardingStore.setUserData({ username });
		onboardingStore.nextStep();
		goto('/onboarding/analyzing');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/gmail');
	}
</script>

<svelte:head>
	<title>Claim Username - OSA Build</title>
</svelte:head>

<GradientBackground variant="ready" fullScreen>
	<div class="username-screen text-center space-y-12 animate-slide-up">
		<div class="space-y-4">
			<h1 class="text-5xl font-bold text-gradient">
				Claim Your Username
			</h1>
			<p class="text-xl text-gray-700 dark:text-gray-300 max-w-xl mx-auto">
				This will be your unique identity in OSA Build
			</p>
		</div>

		<!-- Username input -->
		<div class="username-input max-w-md mx-auto space-y-6">
			<div class="space-y-2">
				<RoundedInput
					label="Username"
					type="text"
					bind:value={username}
					placeholder="bekorains"
					required
					error={error}
					helperText={isAvailable === true ? '✓ Available!' : ''}
					oninput={handleUsernameChange}
				/>

				<PillButton
					variant="secondary"
					size="sm"
					onclick={checkAvailability}
					loading={isChecking}
					disabled={!username || username.length < 3}
				>
					Check Availability
				</PillButton>
			</div>

			<!-- Username tips -->
			<div class="tips glass-card p-6 text-left">
				<h3 class="font-semibold mb-3">Username tips:</h3>
				<ul class="text-sm text-gray-600 dark:text-gray-400 space-y-2">
					<li>• At least 3 characters</li>
					<li>• Letters, numbers, and underscores only</li>
					<li>• Choose something memorable - this is how others will find you</li>
				</ul>
			</div>
		</div>

		<!-- CTA -->
		<div class="cta-section flex gap-4 justify-center">
			<PillButton variant="ghost" size="md" onclick={handleBack}>
				Back
			</PillButton>
			<PillButton
				variant="primary"
				size="lg"
				onclick={handleContinue}
				disabled={!isAvailable}
			>
				Continue
			</PillButton>
		</div>
	</div>
</GradientBackground>

<style>
	.username-input {
		animation: fade-in 0.5s ease-out;
	}
</style>
