<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { get } from 'svelte/store';
	import { PillButton } from '$lib/components/osa';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { fade } from 'svelte/transition';
	import { generateInsights } from '$lib/utils/insightTemplates';
	import {
		CheckSquare,
		Users,
		Kanban,
		BookOpen,
		Calendar,
		BarChart3,
		Mail,
		Clock,
		FileText,
		Target,
		Briefcase,
		Wallet,
		MessageSquare,
		Settings,
		Layers,
		type Icon
	} from 'lucide-svelte';

	// Map app titles to Lucide icons
	const iconMap: Record<string, typeof Icon> = {
		'task': CheckSquare,
		'crm': Users,
		'client': Users,
		'project': Kanban,
		'journal': BookOpen,
		'calendar': Calendar,
		'analytics': BarChart3,
		'report': BarChart3,
		'email': Mail,
		'inbox': Mail,
		'time': Clock,
		'document': FileText,
		'docs': FileText,
		'goal': Target,
		'okr': Target,
		'work': Briefcase,
		'finance': Wallet,
		'invoice': Wallet,
		'chat': MessageSquare,
		'message': MessageSquare,
		'setting': Settings,
		'workflow': Layers
	};

	function getAppIcon(title: string): typeof Icon {
		const lowerTitle = title.toLowerCase();
		for (const [keyword, icon] of Object.entries(iconMap)) {
			if (lowerTitle.includes(keyword)) {
				return icon;
			}
		}
		return Layers; // Default icon
	}

	let phase = $state<'analyzing' | 'ready'>('analyzing');
	let currentInsightIndex = $state(0);
	let showInsight = $state(false);
	let isLoading = $state(true);

	const store = $derived($onboardingStore);
	const starterApps = $derived(store.userData.starterApps || []);
	const quickInfo = $derived(store.userData.quickInfo);
	const fallbackForm = $derived(store.userData.fallbackForm);

	// Generate personalized insights using template system
	const insights = $derived(
		generateInsights({
			role: quickInfo?.role || '',
			businessType: quickInfo?.businessType || '',
			mainFocus: fallbackForm?.mainFocus,
			workStyle: fallbackForm?.workStyle
		})
	);

	function delay(ms: number): Promise<void> {
		return new Promise(resolve => setTimeout(resolve, ms));
	}

	async function cycleInsights() {
		for (let i = 0; i < 3; i++) {
			currentInsightIndex = i;
			showInsight = true;
			await delay(1500);
			showInsight = false;
			await delay(300);
		}
	}

	// Generate personalized apps based on user's onboarding data
	function generatePersonalizedApps() {
		const role = quickInfo?.role || '';
		const businessType = quickInfo?.businessType || '';
		const mainFocus = fallbackForm?.mainFocus || '';

		// Base apps everyone gets
		const apps: Array<{id: string; title: string; description: string; reason: string}> = [
			{ id: '1', title: 'Task Manager', description: 'Organize and track your daily tasks', reason: 'Essential for productivity' }
		];

		// Add apps based on role
		if (role === 'founder' || role === 'consultant') {
			apps.push({ id: '2', title: 'Client CRM', description: 'Manage client relationships and deals', reason: 'Based on your role as ' + role });
		}

		// Add apps based on business type
		if (businessType === 'agency' || businessType === 'consulting') {
			apps.push({ id: '3', title: 'Project Tracker', description: 'Track project milestones and deliverables', reason: 'Perfect for ' + businessType + ' work' });
			apps.push({ id: '4', title: 'Time Tracker', description: 'Log billable hours and time spent', reason: 'Track client billing' });
		} else if (businessType === 'startup' || businessType === 'saas') {
			apps.push({ id: '3', title: 'Product Roadmap', description: 'Plan features and track development', reason: 'Build your product vision' });
			apps.push({ id: '4', title: 'Analytics Dashboard', description: 'Track key metrics and KPIs', reason: 'Data-driven decisions' });
		} else if (businessType === 'ecommerce') {
			apps.push({ id: '3', title: 'Inventory Tracker', description: 'Manage stock and orders', reason: 'Essential for e-commerce' });
			apps.push({ id: '4', title: 'Sales Dashboard', description: 'Track revenue and conversions', reason: 'Monitor your sales' });
		} else if (businessType === 'freelance') {
			apps.push({ id: '3', title: 'Invoice Manager', description: 'Create and track invoices', reason: 'Get paid on time' });
			apps.push({ id: '4', title: 'Time Tracker', description: 'Log hours for projects', reason: 'Track your work' });
		} else {
			apps.push({ id: '3', title: 'Project Tracker', description: 'Track project progress', reason: 'Stay organized' });
		}

		// Add apps based on main focus
		if (mainFocus.includes('Sales') || mainFocus.includes('BD')) {
			if (!apps.some(a => a.title.includes('CRM'))) {
				apps.push({ id: '5', title: 'Sales Pipeline', description: 'Track deals and opportunities', reason: 'Based on your focus on sales' });
			}
		} else if (mainFocus.includes('Marketing')) {
			apps.push({ id: '5', title: 'Campaign Tracker', description: 'Plan and track marketing campaigns', reason: 'For your marketing focus' });
		} else if (mainFocus.includes('Operations')) {
			apps.push({ id: '5', title: 'Process Manager', description: 'Document and optimize workflows', reason: 'Streamline operations' });
		}

		// Always add Daily Journal as last app if we have room
		if (apps.length < 4) {
			apps.push({ id: String(apps.length + 1), title: 'Daily Journal', description: 'Log activities and reflections', reason: 'Build better habits' });
		}

		// Limit to 4 apps
		return apps.slice(0, 4);
	}

	function loadApps() {
		if (starterApps.length > 0) {
			isLoading = false;
			return;
		}

		// Generate personalized apps based on onboarding data
		const personalizedApps = generatePersonalizedApps();
		onboardingStore.setStarterApps(personalizedApps);
		isLoading = false;
	}

	function handleContinue() {
		onboardingStore.nextStep();
		goto('/onboarding/ready');
	}

	function handleBack() {
		onboardingStore.prevStep();
		goto('/onboarding/connect');
	}

	onMount(async () => {
		// Handle OAuth callback - check URL params for successful OAuth
		if (browser) {
			const url = new URL(window.location.href);
			const source = url.searchParams.get('source');
			const integration = url.searchParams.get('integration');
			const status = url.searchParams.get('status');
			const error = url.searchParams.get('error');

			// Handle OAuth error
			if (error) {
				console.error('[Building] OAuth error:', error);
				// Could show a toast/notification here
			}

			// Handle successful Google OAuth from connect page
			if (source === 'google-oauth' || (integration === 'google' && status === 'connected')) {
				if (import.meta.env.DEV) console.log('[Building] Google OAuth successful, updating store');
				onboardingStore.setUserData({ gmailConnected: true });
				onboardingStore.setIntegrationsConnected(['google-oauth']);
			}

			// Handle other OAuth integrations (slack, notion, etc.)
			if (integration && status === 'connected') {
				if (import.meta.env.DEV) console.log(`[Building] ${integration} OAuth successful, updating store`);
				const storeValue = get(onboardingStore);
				const currentIntegrations = storeValue.userData.integrationsConnected || [];
				if (!currentIntegrations.includes(integration)) {
					onboardingStore.setIntegrationsConnected([...currentIntegrations, integration]);
				}
			}

			// Clean up URL params (remove sensitive OAuth data from URL)
			if (source || integration || status || error) {
				window.history.replaceState({}, '', '/onboarding/building');
			}
		}

		// Load apps immediately (sync)
		loadApps();

		// Cycle through insights animation
		await cycleInsights();

		await delay(500);
		phase = 'ready';
	});
</script>

<svelte:head>
	<title>Building Your Apps - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="building-screen">
		<div class="content">
			{#if phase === 'analyzing'}
				<div class="analyzing-phase" in:fade={{ duration: 300 }} out:fade={{ duration: 300 }}>
					<h1 class="title">Building your apps...</h1>

					{#if showInsight}
						<p class="insight" in:fade={{ duration: 300 }} out:fade={{ duration: 300 }}>
							{insights[currentInsightIndex]}
						</p>
					{/if}

					<div class="dots">
						{#each [0, 1, 2] as i}
							<span class="dot" class:active={i === currentInsightIndex}></span>
						{/each}
					</div>

					<div class="spinner"></div>
				</div>
			{:else}
				<div class="ready-phase" in:fade={{ duration: 400 }}>
					<h1 class="title">Your starter apps are ready</h1>

					{#if isLoading}
						<div class="spinner"></div>
					{:else}
						<div class="apps-grid">
							{#each starterApps as app}
								{@const AppIcon = getAppIcon(app.title)}
								<div class="app-card">
									<div class="app-icon">
										<AppIcon size={24} strokeWidth={1.5} />
									</div>
									<div class="app-info">
										<h3 class="app-title">{app.title}</h3>
										<p class="app-description">{app.description}</p>
									</div>
								</div>
							{/each}
						</div>

						<button class="btn-pill btn-pill-ghost browse-more">Browse more apps</button>

						<div class="cta">
							<PillButton variant="primary" size="lg" onclick={handleContinue}>
								Continue
							</PillButton>
							<button class="btn-pill btn-pill-ghost back-button" onclick={handleBack}>Back</button>
						</div>
					{/if}
				</div>
			{/if}
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

	.building-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.content {
		width: 100%;
		max-width: 700px;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}

	.title {
		font-size: 2.5rem;
		font-weight: 700;
		color: #1A1A1A;
		line-height: 1.2;
		letter-spacing: -0.02em;
		margin: 0 0 2rem 0;
	}

	.analyzing-phase {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2rem;
		min-height: 300px;
	}

	.insight {
		font-size: 1.25rem;
		color: #666666;
		margin: 0;
		min-height: 2rem;
	}

	.dots {
		display: flex;
		gap: 0.5rem;
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #D1D5DB;
		transition: all 0.3s ease;
	}

	.dot.active {
		background: #1A1A1A;
		transform: scale(1.25);
	}

	.spinner {
		width: 48px;
		height: 48px;
		border: 3px solid #E5E5E5;
		border-top-color: #1A1A1A;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.ready-phase {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2rem;
		animation: fadeIn 0.5s ease-out;
	}

	.apps-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		width: 100%;
	}

	.app-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.95);
		border-radius: 12px;
		text-align: left;
		border: 1px solid #E5E5E5;
	}

	.app-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		background: linear-gradient(135deg, #1A1A1A, #333333);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		color: #FFFFFF;
	}

	.app-info {
		flex: 1;
		min-width: 0;
	}

	.app-title {
		font-size: 1rem;
		font-weight: 600;
		color: #1A1A1A;
		margin: 0 0 0.25rem 0;
	}

	.app-description {
		font-size: 0.8rem;
		color: #666666;
		margin: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.browse-more {
		background: transparent;
		border: none;
		color: #666666;
		font-size: 0.875rem;
		cursor: pointer;
		font-family: inherit;
		text-decoration: underline;
	}

	.browse-more:hover {
		color: #1A1A1A;
	}

	.cta {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		margin-top: 1rem;
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

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	@keyframes fadeIn {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}

	@media (max-width: 768px) {
		.title { font-size: 2rem; }
		.apps-grid { grid-template-columns: 1fr; }
	}
</style>
