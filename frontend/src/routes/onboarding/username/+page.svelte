<!--
	Onboarding Screen 5: Claim Username
	Simplified username selection matching Wabi design
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { checkUsernameAvailability, setUsername } from '$lib/api/users';
	import { Check, X } from 'lucide-svelte';

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
		const validationError = validateUsername(username);
		if (validationError) {
			error = validationError;
			isAvailable = false;
			return;
		}

		isChecking = true;
		error = '';

		try {
			console.log('[Username] Checking availability for:', username);
			const response = await checkUsernameAvailability(username);
			console.log('[Username] Response:', response);
			isAvailable = response.available;

			if (!response.available) {
				error = response.reason || 'Username is already taken';
			}
		} catch (err) {
			console.error('[Username] Error checking availability:', err);
			console.error('[Username] Error details:', {
				message: err instanceof Error ? err.message : String(err),
				username: username
			});
			error = 'Unable to check availability. Please try again.';
			isAvailable = null;
		} finally {
			isChecking = false;
		}
	}

	// Watch for username changes and debounce availability check
	$effect(() => {
		isAvailable = null;
		error = '';

		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		if (!username) {
			return;
		}

		const validationError = validateUsername(username);
		if (validationError) {
			error = validationError;
			return;
		}

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
			const response = await setUsername(username);

			if (response.success) {
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

<div class="onboarding-background">
	<div class="username-screen">
		<div class="content">
			<!-- Main Message -->
			<h1 class="title">
				Claim your<br />username.
			</h1>

			<!-- Username Input -->
			<div class="input-section">
				<div class="input-wrapper">
					<input
						id="username"
						type="text"
						bind:value={username}
						placeholder="yourname"
						class="input-field"
						class:input-error={error && username}
						class:input-success={isAvailable === true}
					/>
					{#if isAvailable === true}
						<div class="input-icon input-icon-success">
							<Check size={20} />
						</div>
					{:else if error && username}
						<div class="input-icon input-icon-error">
							<X size={20} />
						</div>
					{/if}
				</div>

				{#if error}
					<p class="helper-text error-text">{error}</p>
				{:else if isAvailable === true}
					<p class="helper-text success-text">Available!</p>
				{:else}
					<p class="helper-text">At least 3 characters, letters, numbers, and underscores only</p>
				{/if}
			</div>

			<!-- CTA Buttons -->
			<div class="cta">
				<PillButton
					variant="primary"
					size="lg"
					onclick={handleContinue}
					disabled={!isAvailable || isChecking}
				>
					{#if isChecking}
						Processing...
					{:else}
						Continue
					{/if}
				</PillButton>

				<button class="back-button" onclick={handleBack}>
					Back
				</button>
			</div>
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

	.username-screen {
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

	.input-section {
		width: 100%;
		max-width: 400px;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.input-wrapper {
		position: relative;
		width: 100%;
	}

	.input-field {
		width: 100%;
		padding: 1rem 3rem 1rem 1rem;
		font-size: 1.125rem;
		color: #1A1A1A;
		background-color: white;
		border: 2px solid #D1D5DB;
		border-radius: 0.75rem;
		font-family: inherit;
		transition: all 0.2s ease;
		/* DEEPER inset shadow to create more pronounced recessed/embedded effect */
		box-shadow:
			inset 0 3px 8px 0 rgba(0, 0, 0, 0.12),
			inset 0 1px 3px 0 rgba(0, 0, 0, 0.08),
			0 1px 2px 0 rgba(0, 0, 0, 0.05);
	}

	.input-field:focus {
		outline: none;
		border-color: #1A1A1A;
		/* Enhanced deeper shadow on focus with ring effect */
		box-shadow:
			inset 0 3px 8px 0 rgba(0, 0, 0, 0.15),
			inset 0 1px 3px 0 rgba(0, 0, 0, 0.1),
			0 0 0 3px rgba(26, 26, 26, 0.1);
	}

	.input-field::placeholder {
		color: #9CA3AF;
	}

	.input-error {
		border-color: #DC2626;
	}

	.input-success {
		border-color: #10B981;
	}

	.input-icon {
		position: absolute;
		right: 1rem;
		top: 50%;
		transform: translateY(-50%);
		pointer-events: none;
	}

	.input-icon-success {
		color: #10B981;
	}

	.input-icon-error {
		color: #DC2626;
	}

	.helper-text {
		font-size: 0.875rem;
		color: #666666;
		margin: 0;
		text-align: left;
	}

	.error-text {
		color: #DC2626;
	}

	.success-text {
		color: #10B981;
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		animation: fadeIn 0.8s ease-out 0.4s both;
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
		.title {
			font-size: 2rem;
		}

		.content {
			gap: 2.5rem;
		}

		.input-field {
			font-size: 1rem;
		}
	}
</style>
