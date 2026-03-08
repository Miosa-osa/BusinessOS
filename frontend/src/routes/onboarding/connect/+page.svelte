<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { PillButton, PillSelect, RoundedInput } from '$lib/components/osa';
	import { onboardingStore, FALLBACK_OPTIONS, type FallbackFormData } from '$lib/stores/onboardingStore';
	import { cloudServerUrl } from '$lib/auth-client';

	let connectedIntegrations = $state<string[]>([]);
	let showFallbackForm = $state(false);
	let isLoadingOAuth = $state(false);
	let loadingIntegration = $state<string | null>(null);
	let oauthError = $state<string | null>(null);
	let fallbackForm = $state<FallbackFormData>({
		toolsUsed: [],
		mainFocus: '',
		challenge: '',
		workStyle: '',
		whatWouldHelp: []
	});

	let errors = $state({
		toolsUsed: '',
		mainFocus: '',
		challenge: '',
		workStyle: '',
		whatWouldHelp: ''
	});

	// Get quick info for smart defaults
	const store = $derived($onboardingStore);
	const quickInfo = $derived(store.userData.quickInfo);

	// Smart defaults based on role
	$effect(() => {
		if (showFallbackForm && quickInfo?.role && !fallbackForm.mainFocus) {
			// Pre-fill main_focus based on role
			if (quickInfo.role === 'founder' || quickInfo.role === 'consultant') {
				fallbackForm.mainFocus = 'Client work & delivery';
			} else if (quickInfo.role === 'freelancer') {
				fallbackForm.mainFocus = 'Client work & delivery';
			}
		}
	});

	// Example prompts for challenge field
	const exampleChallenges = [
		{ label: 'Managing deadlines', text: 'Managing client expectations and deadlines across multiple projects' },
		{ label: 'Team alignment', text: 'Keeping my team aligned on priorities and project status' },
		{ label: 'Tool overload', text: 'Too many tools and not enough integration between them' },
		{ label: 'Time tracking', text: 'Tracking time accurately for client billing and reporting' }
	];

	const integrations = [
		{ id: 'google', name: 'Google', icon: '/logos/integrations/google.svg', description: 'Gmail, Calendar, Drive' },
		{ id: 'slack', name: 'Slack', icon: '/logos/integrations/slack.svg', description: 'Messages & channels' },
		{ id: 'notion', name: 'Notion', icon: '/logos/integrations/notion.svg', description: 'Docs & databases' },
		{ id: 'outlook', name: 'Outlook', icon: '/logos/integrations/outlook.svg', description: 'Email & calendar' },
		{ id: 'linear', name: 'Linear', icon: '/logos/integrations/linear.svg', description: 'Issues & projects' }
	];

	const toolsUsedOptions = FALLBACK_OPTIONS.toolsUsed.map(t => ({ value: t, label: t }));
	const mainFocusOptions = FALLBACK_OPTIONS.mainFocus.map(f => ({ value: f, label: f }));
	const workStyleOptions = FALLBACK_OPTIONS.workStyle.map(w => ({ value: w, label: w }));
	const whatWouldHelpOptions = FALLBACK_OPTIONS.whatWouldHelp.map(h => ({ value: h, label: h }));

	const hasConnectedIntegrations = $derived(connectedIntegrations.length > 0 || connectedIntegrations.includes('google-oauth'));
	const canContinue = $derived(hasConnectedIntegrations || isFormValid());

	function isFormValid(): boolean {
		if (!showFallbackForm) return false;
		return (
			fallbackForm.toolsUsed.length > 0 &&
			fallbackForm.mainFocus !== '' &&
			fallbackForm.challenge.trim() !== '' &&
			fallbackForm.workStyle !== '' &&
			fallbackForm.whatWouldHelp.length > 0 &&
			fallbackForm.whatWouldHelp.length <= 3
		);
	}

	function handleGoogleOAuth() {
		isLoadingOAuth = true;
		oauthError = null;

		try {
			// Call backend OAuth directly with FULL frontend URL redirect
			const backendUrl = get(cloudServerUrl);
			const frontendUrl = window.location.origin;
			// Include source param so the callback knows this came from onboarding
			const redirectAfter = `${frontendUrl}/onboarding/building?source=google-oauth`;

			// DO NOT update state here - the OAuth callback page will handle it
			// Updating state before redirect causes issues if user cancels or OAuth fails
			// The /onboarding/building page will check URL params and update state on success

			if (import.meta.env.DEV) console.log('[Connect] Redirecting to Google OAuth with redirect:', redirectAfter);
			window.location.href = `${backendUrl}/api/auth/google?redirect=${encodeURIComponent(redirectAfter)}`;
		} catch (err) {
			oauthError = err instanceof Error ? err.message : 'Failed to connect';
			console.error('OAuth error:', err);
			isLoadingOAuth = false;
		}
	}

	function handleIntegrationClick(integrationId: string) {
		// For OAuth integrations, initiate the flow
		const oauthIntegrations = ['slack', 'notion', 'outlook', 'linear'];

		if (oauthIntegrations.includes(integrationId)) {
			loadingIntegration = integrationId;
			oauthError = null;

			try {
				const backendUrl = get(cloudServerUrl);
				const frontendUrl = window.location.origin;
				const redirectAfter = `${frontendUrl}/onboarding/building?source=${integrationId}-oauth&integration=${integrationId}&status=connected`;

				// Map integration IDs to API endpoints
				const endpointMap: Record<string, string> = {
					'slack': 'slack',
					'notion': 'notion',
					'outlook': 'microsoft',
					'linear': 'linear'
				};

				const endpoint = endpointMap[integrationId];
				window.location.href = `${backendUrl}/api/auth/${endpoint}?redirect=${encodeURIComponent(redirectAfter)}`;
			} catch (err) {
				oauthError = `Failed to connect to ${integrationId}. Please try again.`;
				loadingIntegration = null;
			}
			return;
		}

		// For non-OAuth integrations, toggle selection
		if (connectedIntegrations.includes(integrationId)) {
			connectedIntegrations = connectedIntegrations.filter(id => id !== integrationId);
		} else {
			connectedIntegrations = [...connectedIntegrations, integrationId];
		}
	}

	function toggleFallbackForm() {
		showFallbackForm = !showFallbackForm;
	}

	function handleToolsUsedChange(value: string) {
		if (fallbackForm.toolsUsed.includes(value)) {
			fallbackForm.toolsUsed = fallbackForm.toolsUsed.filter(t => t !== value);
		} else {
			fallbackForm.toolsUsed = [...fallbackForm.toolsUsed, value];
		}
	}

	function handleWhatWouldHelpChange(value: string) {
		if (fallbackForm.whatWouldHelp.includes(value)) {
			fallbackForm.whatWouldHelp = fallbackForm.whatWouldHelp.filter(h => h !== value);
		} else if (fallbackForm.whatWouldHelp.length < 3) {
			fallbackForm.whatWouldHelp = [...fallbackForm.whatWouldHelp, value];
		}
	}

	function validate(): boolean {
		let isValid = true;
		errors = { toolsUsed: '', mainFocus: '', challenge: '', workStyle: '', whatWouldHelp: '' };

		if (hasConnectedIntegrations) return true;
		if (!showFallbackForm) return false;

		if (fallbackForm.toolsUsed.length === 0) {
			errors.toolsUsed = 'Please select at least one tool';
			isValid = false;
		}

		if (!fallbackForm.mainFocus) {
			errors.mainFocus = 'Please select your main focus';
			isValid = false;
		}

		if (!fallbackForm.challenge.trim()) {
			errors.challenge = 'Please describe your biggest challenge';
			isValid = false;
		}

		if (!fallbackForm.workStyle) {
			errors.workStyle = 'Please select your work style';
			isValid = false;
		}

		if (fallbackForm.whatWouldHelp.length === 0) {
			errors.whatWouldHelp = 'Please select at least one option';
			isValid = false;
		} else if (fallbackForm.whatWouldHelp.length > 3) {
			errors.whatWouldHelp = 'Please select up to 3 options';
			isValid = false;
		}

		return isValid;
	}

	function handleContinue() {
		if (!validate()) return;

		if (hasConnectedIntegrations) {
			onboardingStore.setIntegrationsConnected(connectedIntegrations);
		}

		if (showFallbackForm && isFormValid()) {
			onboardingStore.setFallbackForm(fallbackForm);
		}

		onboardingStore.nextStep();
		goto('/onboarding/building');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/username');
	}
</script>

<svelte:head>
	<title>Connect - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="connect-screen">
		<div class="content">
			<div class="header">
				<h1 class="title">Connect your tools</h1>
				<p class="subtitle">We'll analyze your work to build personalized apps</p>
			</div>

			<!-- Google OAuth Section -->
			<div class="oauth-section">
				<div class="oauth-card">
					<div class="gmail-icon-wrapper">
						<svg class="gmail-icon" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
							<path fill="#4285f4" d="M45,16.2l-5,2.75l-5,4.75L35,40h7c1.657,0,3-1.343,3-3V16.2z"/>
							<path fill="#34a853" d="M3,16.2l3.614,1.71L13,23.7V40H6c-1.657,0-3-1.343-3-3V16.2z"/>
							<polygon fill="#fbbc04" points="35,11.2 24,19.45 13,11.2 12,17 13,23.7 24,31.95 35,23.7 36,17"/>
							<path fill="#ea4335" d="M3,12.298V16.2l10,7.5V11.2L9.876,8.859C9.132,8.301,8.228,8,7.298,8h0C4.924,8,3,9.924,3,12.298z"/>
							<path fill="#c5221f" d="M45,12.298V16.2l-10,7.5V11.2l3.124-2.341C38.868,8.301,39.772,8,40.702,8h0 C43.076,8,45,9.924,45,12.298z"/>
						</svg>
					</div>
					<div class="oauth-content">
						<h3 class="oauth-title">Get personalized apps from your Gmail</h3>
						<p class="oauth-description">OSA analyzes your email to understand your work and builds custom apps for you</p>
					</div>
					{#if oauthError}
						<div class="error-banner" role="alert" aria-live="polite">
							<svg class="error-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							<div class="error-content">
								<p class="error-title">Connection failed</p>
								<p class="error-text">{oauthError}</p>
							</div>
							<button class="btn-pill btn-pill-ghost error-dismiss" onclick={() => oauthError = null} aria-label="Dismiss error">
								<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					{/if}
					<PillButton
						variant="primary"
						size="md"
						onclick={handleGoogleOAuth}
						disabled={isLoadingOAuth}
					>
						{isLoadingOAuth ? 'Connecting...' : 'Connect with Google'}
					</PillButton>
					<p class="privacy-note">Your data stays private and secure</p>
				</div>
			</div>

			<div class="divider"><span>OR CONNECT OTHER TOOLS</span></div>

			<div class="integrations-section">
				<div class="integration-grid">
					{#each integrations as integration}
						<button
							class="btn-pill btn-pill-ghost integration-card"
							class:connected={connectedIntegrations.includes(integration.id)}
							class:loading={loadingIntegration === integration.id}
							onclick={() => handleIntegrationClick(integration.id)}
							disabled={loadingIntegration !== null}
							aria-label="Connect {integration.name}: {integration.description}{connectedIntegrations.includes(integration.id) ? ' (Connected)' : ''}{loadingIntegration === integration.id ? ' (Connecting...)' : ''}"
							aria-pressed={connectedIntegrations.includes(integration.id)}
							aria-busy={loadingIntegration === integration.id}
						>
							<div class="integration-icon">
								{#if loadingIntegration === integration.id}
									<svg class="loading-spinner" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
										<circle class="spinner-track" cx="12" cy="12" r="10" stroke="#E5E5E5" stroke-width="3" />
										<path class="spinner-head" d="M12 2a10 10 0 0 1 10 10" stroke="#1A1A1A" stroke-width="3" stroke-linecap="round" />
									</svg>
								{:else}
									<img src={integration.icon} alt="" aria-hidden="true" />
								{/if}
							</div>
							<div class="integration-info">
								<span class="integration-name">{integration.name}</span>
								<span class="integration-desc">
									{#if loadingIntegration === integration.id}
										Connecting...
									{:else}
										{integration.description}
									{/if}
								</span>
							</div>
							{#if connectedIntegrations.includes(integration.id)}
								<div class="connected-badge" aria-hidden="true">Connected</div>
							{/if}
						</button>
					{/each}
				</div>

				{#if hasConnectedIntegrations}
					<p class="connected-message">
						{connectedIntegrations.length} integration{connectedIntegrations.length > 1 ? 's' : ''} connected
					</p>
				{/if}
			</div>

			<div class="divider"><span>OR</span></div>

			{#if !showFallbackForm}
				<button class="btn-pill btn-pill-ghost fallback-toggle" onclick={toggleFallbackForm}>
					I'll answer a few questions instead
				</button>
			{:else}
				<div class="fallback-form">
					<div class="form-header">
						<h2>Tell us about your work</h2>
						<button class="btn-pill btn-pill-ghost collapse-btn" onclick={toggleFallbackForm}>Hide form</button>
					</div>

					<div class="form-field">
						<label class="field-label">What tools do you currently use?</label>
						<p class="helper-text">Select all that apply - helps us recommend the right integrations</p>
						<div class="multi-select-grid">
							{#each toolsUsedOptions as option}
								<button
									class="multi-select-chip"
									class:selected={fallbackForm.toolsUsed.includes(option.value)}
									onclick={() => handleToolsUsedChange(option.value)}
								>
									{option.label}
								</button>
							{/each}
						</div>
						{#if errors.toolsUsed}<span class="error">{errors.toolsUsed}</span>{/if}
					</div>

					<PillSelect
						label="What's your main work focus?"
						bind:value={fallbackForm.mainFocus}
						options={mainFocusOptions}
						error={errors.mainFocus}
						columns={2}
						required
					/>

					<div class="form-field">
						<label class="field-label">What's your biggest challenge?</label>
						<p class="helper-text">Be specific - this helps us build the right tools for you</p>

						{#if !fallbackForm.challenge}
							<div class="example-prompts">
								<p class="example-label">Quick examples:</p>
								<div class="example-chips">
									{#each exampleChallenges as example}
										<button
											class="btn-pill btn-pill-ghost example-chip"
											onclick={() => fallbackForm.challenge = example.text}
										>
											{example.label}
										</button>
									{/each}
								</div>
							</div>
						{/if}

						<RoundedInput
							label=""
							bind:value={fallbackForm.challenge}
							placeholder="Or type your own..."
							error={errors.challenge}
							required
						/>
					</div>

					<PillSelect
						label="How do you prefer to work?"
						bind:value={fallbackForm.workStyle}
						options={workStyleOptions}
						error={errors.workStyle}
						columns={2}
						required
					/>

					<div class="form-field">
						<label class="field-label">What would help you most? <span class="hint">(pick up to 3)</span></label>
						<p class="helper-text">Select your top priorities - we'll customize your workspace</p>
						<div class="multi-select-grid">
							{#each whatWouldHelpOptions as option}
								<button
									class="multi-select-chip"
									class:selected={fallbackForm.whatWouldHelp.includes(option.value)}
									class:disabled={!fallbackForm.whatWouldHelp.includes(option.value) && fallbackForm.whatWouldHelp.length >= 3}
									onclick={() => handleWhatWouldHelpChange(option.value)}
									disabled={!fallbackForm.whatWouldHelp.includes(option.value) && fallbackForm.whatWouldHelp.length >= 3}
								>
									{option.label}
								</button>
							{/each}
						</div>
						{#if errors.whatWouldHelp}<span class="error">{errors.whatWouldHelp}</span>{/if}
					</div>
				</div>
			{/if}

			<div class="cta">
				<PillButton variant="primary" size="lg" onclick={handleContinue} disabled={!canContinue}>
					Continue
				</PillButton>
				<button class="btn-pill btn-pill-ghost back-button" onclick={handleBack}>Back</button>
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

	.connect-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.content {
		width: 100%;
		max-width: 640px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2rem;
		text-align: center;
	}

	.header {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		animation: fadeIn 0.8s ease-out 0.2s both;
	}

	.title {
		font-size: 2.5rem;
		font-weight: 700;
		color: #1A1A1A;
		line-height: 1.2;
		letter-spacing: -0.02em;
		margin: 0;
	}

	.subtitle {
		font-size: 1rem;
		color: #666666;
		margin: 0;
	}

	.integrations-section {
		width: 100%;
		animation: fadeIn 0.8s ease-out 0.3s both;
	}

	.integration-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 1rem;
	}

	.integration-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 1.25rem;
		background: rgba(255, 255, 255, 0.9);
		border: 2px solid #E5E5E5;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.2s ease;
		position: relative;
	}

	.integration-card:hover {
		border-color: #CCCCCC;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.integration-card.connected {
		border-color: #10B981;
		background: rgba(16, 185, 129, 0.05);
	}

	.integration-card.loading {
		border-color: #3B82F6;
		background: rgba(59, 130, 246, 0.05);
		cursor: wait;
	}

	.integration-card:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.loading-spinner {
		width: 32px;
		height: 32px;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.integration-icon {
		width: 48px;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.integration-icon img {
		max-width: 100%;
		max-height: 100%;
	}

	.integration-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.integration-name {
		font-weight: 600;
		color: #1A1A1A;
		font-size: 0.95rem;
	}

	.integration-desc {
		font-size: 0.75rem;
		color: #666666;
	}

	.connected-badge {
		position: absolute;
		top: 8px;
		right: 8px;
		background: #10B981;
		color: white;
		font-size: 0.65rem;
		font-weight: 600;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		text-transform: uppercase;
	}

	.connected-message {
		margin-top: 1rem;
		font-size: 0.875rem;
		color: #10B981;
		font-weight: 500;
	}

	.oauth-section {
		width: 100%;
		animation: fadeIn 0.8s ease-out 0.25s both;
	}

	.oauth-card {
		background: rgba(255, 255, 255, 0.95);
		border-radius: 16px;
		padding: 2rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		text-align: center;
		border: 2px solid #E5E5E5;
	}

	.gmail-icon-wrapper {
		margin-bottom: 0.5rem;
	}

	.gmail-icon {
		width: 64px;
		height: 64px;
		filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.1));
	}

	.oauth-content {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.oauth-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #1A1A1A;
		margin: 0;
	}

	.oauth-description {
		font-size: 0.875rem;
		color: #666666;
		line-height: 1.5;
		margin: 0;
	}

	.privacy-note {
		font-size: 0.75rem;
		color: #999999;
		margin: 0;
	}

	.error-banner {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		width: 100%;
		padding: 0.875rem 1rem;
		background: #FEF2F2;
		border: 1px solid #FECACA;
		border-radius: 10px;
		text-align: left;
	}

	.error-icon {
		width: 20px;
		height: 20px;
		color: #DC2626;
		flex-shrink: 0;
		margin-top: 1px;
	}

	.error-content {
		flex: 1;
		min-width: 0;
	}

	.error-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: #991B1B;
		margin: 0;
	}

	.error-text {
		font-size: 0.8125rem;
		color: #DC2626;
		margin: 0.25rem 0 0 0;
		line-height: 1.4;
	}

	.error-dismiss {
		padding: 0.25rem;
		background: transparent;
		border: none;
		color: #DC2626;
		cursor: pointer;
		border-radius: 4px;
		transition: background-color 0.15s;
	}

	.error-dismiss:hover {
		background: rgba(220, 38, 38, 0.1);
	}

	.error-dismiss svg {
		width: 16px;
		height: 16px;
	}

	.divider {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 1rem;
		color: #999999;
		font-size: 0.75rem;
		font-weight: 500;
		margin: 0.5rem 0;
	}

	.divider::before,
	.divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: #E5E5E5;
	}

	.fallback-toggle {
		background: transparent;
		border: 2px dashed #CCCCCC;
		border-radius: 8px;
		padding: 1rem 2rem;
		color: #666666;
		font-size: 0.95rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		font-family: inherit;
	}

	.fallback-toggle:hover {
		border-color: #999999;
		color: #1A1A1A;
	}

	.fallback-form {
		width: 100%;
		background: rgba(255, 255, 255, 0.95);
		border-radius: 16px;
		padding: 2rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		animation: slideIn 0.3s ease-out;
		text-align: left;
	}

	.form-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.form-header h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #1A1A1A;
		margin: 0;
	}

	.collapse-btn {
		background: transparent;
		border: none;
		color: #666666;
		font-size: 0.875rem;
		cursor: pointer;
		font-family: inherit;
	}

	.collapse-btn:hover {
		color: #1A1A1A;
	}

	.form-field {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.field-label {
		font-size: 0.95rem;
		font-weight: 500;
		color: #1A1A1A;
	}

	.hint {
		font-weight: 400;
		color: #999999;
		font-size: 0.875rem;
	}

	.multi-select-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.multi-select-chip {
		padding: 0.5rem 1rem;
		background: white;
		border: 1.5px solid #E5E5E5;
		border-radius: 20px;
		font-size: 0.875rem;
		color: #666666;
		cursor: pointer;
		transition: all 0.15s ease;
		font-family: inherit;
	}

	.multi-select-chip:hover:not(.disabled) {
		border-color: #CCCCCC;
	}

	.multi-select-chip.selected {
		background: #1A1A1A;
		border-color: #1A1A1A;
		color: white;
	}

	.multi-select-chip.disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.error {
		font-size: 0.8rem;
		color: #EF4444;
	}

	.helper-text {
		font-size: 0.8rem;
		color: #999999;
		margin: 0.25rem 0 0.75rem 0;
		line-height: 1.4;
	}

	.example-prompts {
		margin-bottom: 0.75rem;
		padding: 1rem;
		background: rgba(249, 250, 251, 0.8);
		border-radius: 8px;
		border: 1px dashed #E5E5E5;
	}

	.example-label {
		font-size: 0.8rem;
		color: #666666;
		margin: 0 0 0.5rem 0;
		font-weight: 500;
	}

	.example-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.example-chip {
		padding: 0.5rem 1rem;
		background: white;
		border: 1.5px solid #E5E5E5;
		border-radius: 20px;
		font-size: 0.8125rem;
		color: #666666;
		cursor: pointer;
		transition: all 0.2s ease;
		font-family: inherit;
	}

	.example-chip:hover {
		border-color: #1A1A1A;
		color: #1A1A1A;
		background: #F9FAFB;
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

	@keyframes slideIn {
		from { opacity: 0; transform: translateY(-10px); }
		to { opacity: 1; transform: translateY(0); }
	}

	@keyframes fadeIn {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}

	@media (max-width: 768px) {
		.title { font-size: 2rem; }
		.integration-grid { grid-template-columns: repeat(2, 1fr); }
		.fallback-form { padding: 1.5rem; }
	}
</style>
