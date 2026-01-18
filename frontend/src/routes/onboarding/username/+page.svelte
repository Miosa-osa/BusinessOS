<!--
	Onboarding Screen 5: Claim Username
	User chooses their unique username
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { GradientBackground, PillButton, RoundedInput } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { checkUsernameAvailability, setUsername } from '$lib/api/users';

	let username = $state('');
	let isChecking = $state(false);
	let isAvailable = $state<boolean | null>(null);
	let error = $state('');
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	// Username validation regex: alphanumeric + underscore only
	const USERNAME_REGEX = /^[a-zA-Z0-9_]+$/;

	function validateUsername(value: string): string | null {
		if (!value) {
			return 'Username is required';
		}
		if (value.length < 3) {
			return 'Username must be at least 3 characters';
		}
		if (!USERNAME_REGEX.test(value)) {
			return 'Username can only contain letters, numbers, and underscores';
		}
		return null;
	}

	async function checkAvailability() {
		// Validate first
		const validationError = validateUsername(username);
		if (validationError) {
			error = validationError;
			isAvailable = false;
			return;
		}

		isChecking = true;
		error = '';

		try {
			const response = await checkUsernameAvailability(username);
			isAvailable = response.available;

			if (!response.available) {
				error = response.reason || 'Username is already taken';
			}
		} catch (err) {
			error = 'Unable to check availability. Please try again.';
			isAvailable = null;
			console.error('Error checking username availability:', err);
		} finally {
			isChecking = false;
		}
	}

	// Watch for username changes and debounce availability check
	$effect(() => {
		// Reset state when username changes
		isAvailable = null;
		error = '';

		// Clear any existing timer
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		// Don't check empty username
		if (!username) {
			return;
		}

		// Only auto-check if username passes basic validation
		const validationError = validateUsername(username);
		if (validationError) {
			error = validationError;
			return;
		}

		// Auto-check after 500ms of no typing
		debounceTimer = setTimeout(() => {
			checkAvailability();
		}, 500);
	});

	async function handleContinue() {
		if (!isAvailable) {
			error = 'Please choose an available username';
			return;
		}

		isChecking = true;
		error = '';

		try {
			// Call API to set username
			const response = await setUsername(username);

			if (response.success) {
				// Update local store
				onboardingStore.setUserData({ username });
				onboardingStore.nextStep();
				goto('/onboarding/analyzing');
			} else {
				error = 'Failed to set username. Please try again.';
				isChecking = false;
			}
		} catch (err) {
			error = 'Something went wrong. Please try again.';
			isChecking = false;
			console.error('Error setting username:', err);
		}
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
				disabled={!isAvailable || isChecking}
				loading={isChecking}
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
