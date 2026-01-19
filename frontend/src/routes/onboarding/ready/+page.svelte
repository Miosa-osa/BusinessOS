<!--
	Onboarding Screen 13: Your OS is Ready
	Simplified final celebration screen matching Wabi design
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { get } from 'svelte/store';
	import { cloudServerUrl } from '$lib/auth-client';

	async function handleEnterOS() {
		try {
			// Call backend to mark onboarding complete (use relative URL)
			await fetch('http://localhost:8001/api/users/me/complete-onboarding', {
				method: 'POST',
				credentials: 'include',
			});
			console.log('Onboarding marked complete in backend');
		} catch (err) {
			console.error('Failed to mark onboarding complete:', err);
			// Don't block - continue anyway
		}

		await onboardingStore.complete();
		goto('/window');
	}
</script>

<svelte:head>
	<title>Your OS is Ready - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="ready-screen">
		<div class="content">
			<!-- Main Message -->
			<h1 class="title">
				Your OS<br />is ready.
			</h1>

			<p class="subtitle">
				4 personalized apps are waiting for you.
			</p>

			<!-- CTA Button -->
			<div class="cta">
				<PillButton variant="primary" size="lg" onclick={handleEnterOS}>
					Enter Your OS
				</PillButton>
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

	.ready-screen {
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
		gap: 2rem;
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
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.cta {
		animation: fadeIn 0.8s ease-out 0.4s both;
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
			gap: 2rem;
		}
	}
</style>
