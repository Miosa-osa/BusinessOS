<script lang="ts">
	import { api, type UserSettings, type SystemInfo, type GoogleConnectionStatus } from '$lib/api';
	import { useSession } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { notificationStore } from '$lib/stores/notifications';

	import GeneralSettings from '$lib/components/settings/GeneralSettings.svelte';
	import AISettingsTab from '$lib/components/settings/AISettingsTab.svelte';
	import NotificationsSettings from '$lib/components/settings/NotificationsSettings.svelte';
	import IntegrationsSettings from '$lib/components/settings/IntegrationsSettings.svelte';
	import AccountSettings from '$lib/components/settings/AccountSettings.svelte';
	import UsageSettings from '$lib/components/settings/UsageSettings.svelte';
	import VoiceSettings from '$lib/components/settings/VoiceSettings.svelte';
	import PersonalizationSettings from '$lib/components/settings/PersonalizationSettings.svelte';
	import DesktopSettings from '$lib/components/settings/DesktopSettings.svelte';

	type TabId = 'general' | 'ai' | 'notifications' | 'integrations' | 'account' | 'usage' | 'voice' | 'personalization' | 'desktop';

	const session = useSession();

	let settings = $state<UserSettings | null>(null);
	let systemInfo = $state<SystemInfo | null>(null);
	let googleStatus = $state<GoogleConnectionStatus | null>(null);
	let isLoading = $state(true);
	let activeTab = $state<TabId>('general');
	let isDesktop = $state(false);
	let googleMessage = $state('');

	const tabs: Array<{ id: TabId; label: string; desktopOnly?: boolean }> = [
		{ id: 'general', label: 'General' },
		{ id: 'ai', label: 'AI' },
		{ id: 'notifications', label: 'Notifications' },
		{ id: 'integrations', label: 'Integrations' },
		{ id: 'account', label: 'Account' },
		{ id: 'usage', label: 'Usage' },
		{ id: 'voice', label: 'Voice' },
		{ id: 'personalization', label: 'Personalize' },
		{ id: 'desktop', label: 'Desktop', desktopOnly: true },
	];

	const externalTabs: Array<{ label: string; href: string }> = [
		{ label: 'Workspace', href: '/settings/workspace' },
		{ label: 'Modules', href: '/settings/modules' },
	];

	onMount(async () => {
		await Promise.all([loadSettings(), loadSystemInfo(), loadGoogleStatus()]);
		await notificationStore.fetchNotifications().catch(() => {});

		if (typeof window !== 'undefined' && (window as any).electron) {
			isDesktop = true;
		}

		isLoading = false;

		// Handle OAuth callback query params
		const url = new URL(window.location.href);
		if (url.searchParams.get('google_connected') === 'true') {
			activeTab = 'integrations';
			googleMessage = 'Google Calendar connected successfully!';
			setTimeout(() => (googleMessage = ''), 3000);
			url.searchParams.delete('google_connected');
			window.history.replaceState({}, '', url.toString());
		}
		if (url.searchParams.get('google_error')) {
			activeTab = 'integrations';
			googleMessage = `Error: ${url.searchParams.get('google_error')}`;
			setTimeout(() => (googleMessage = ''), 5000);
			url.searchParams.delete('google_error');
			window.history.replaceState({}, '', url.toString());
		}
	});

	async function loadSettings() {
		try {
			settings = await api.getSettings();
		} catch (error) {
			console.error('Error loading settings:', error);
		}
	}

	async function loadSystemInfo() {
		try {
			systemInfo = await api.getSystemInfo();
		} catch (error) {
			console.error('Error loading system info:', error);
		}
	}

	async function loadGoogleStatus() {
		try {
			googleStatus = await api.getGoogleConnectionStatus();
		} catch (error) {
			console.error('Error loading Google status:', error);
			googleStatus = { connected: false };
		}
	}

	function isActive(tabId: TabId): boolean {
		return activeTab === tabId;
	}

	function tabClass(tabId: TabId): string {
		return isActive(tabId)
			? 'st-tab-active border-b-2 -mb-px'
			: 'st-tab-inactive';
	}
</script>

<div class="h-full flex flex-col st-page-bg">
	<!-- Header -->
	<div class="px-6 py-4 st-page-header">
		<h1 class="text-xl font-semibold st-title">Settings</h1>
		<p class="text-sm st-muted mt-0.5">Manage your account and preferences</p>
	</div>

	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 st-spinner border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto">
			<div class="max-w-4xl mx-auto p-6">
				<!-- Tab Navigation -->
				<div class="relative mb-6">
					<div class="flex gap-1 overflow-x-auto scrollbar-hide pb-px st-tab-border">
						{#each tabs as tab}
							{#if !tab.desktopOnly || isDesktop}
								<button
									onclick={() => (activeTab = tab.id)}
									class="btn-pill btn-pill-ghost btn-pill-sm whitespace-nowrap flex-shrink-0 {tabClass(tab.id)}"
								>
									{tab.label}
								</button>
							{/if}
						{/each}
						{#each externalTabs as extTab}
							<button
								onclick={() => goto(extTab.href)}
								class="btn-pill btn-pill-ghost btn-pill-sm whitespace-nowrap flex-shrink-0"
							>
								{extTab.label}
							</button>
						{/each}
					</div>
					<!-- Scroll fade -->
					<div class="absolute right-0 top-0 bottom-px w-8 st-fade-gradient pointer-events-none"></div>
				</div>

				<!-- Tab Panels -->
				{#if activeTab === 'general'}
					<GeneralSettings
						initialTheme={settings?.theme ?? 'light'}
						initialShareAnalytics={settings?.share_analytics ?? true}
					/>
				{/if}

				{#if activeTab === 'ai'}
					<AISettingsTab {systemInfo} />
				{/if}

				{#if activeTab === 'notifications'}
					<NotificationsSettings
						initialEmailNotifications={settings?.email_notifications ?? true}
						initialDailySummary={settings?.daily_summary ?? false}
					/>
				{/if}

				{#if activeTab === 'integrations'}
					<IntegrationsSettings
						initialGoogleStatus={googleStatus}
						initialMessage={googleMessage}
					/>
				{/if}

				{#if activeTab === 'account'}
					<AccountSettings />
				{/if}

				{#if activeTab === 'usage'}
					<UsageSettings />
				{/if}

				{#if activeTab === 'voice'}
					<VoiceSettings />
				{/if}

				{#if activeTab === 'personalization'}
					<PersonalizationSettings />
				{/if}

				{#if activeTab === 'desktop' && isDesktop}
					<DesktopSettings />
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
  :global(.st-page-bg) { background: var(--dbg, var(--bos-card, #fff)); }
  :global(.st-page-header) { border-bottom: 1px solid var(--dbd2, #f0f0f0); }
  :global(.st-title) { color: var(--dt, var(--bos-text-primary, #111)); }
  :global(.st-muted) { color: var(--dt3, var(--bos-text-tertiary, #888)); }
  :global(.st-spinner) { border-color: var(--dt, var(--bos-text-primary, #111)); }
  :global(.st-tab-border) { border-bottom: 1px solid var(--dbd, var(--bos-border, #e0e0e0)); }
  :global(.st-tab-active) {
    color: var(--dt, var(--bos-text-primary, #111));
    border-color: var(--dt, var(--bos-text-primary, #111));
  }
  :global(.st-tab-inactive) {
    color: var(--dt3, var(--bos-text-tertiary, #888));
  }
  :global(.st-tab-inactive:hover) {
    color: var(--dt2, var(--bos-text-secondary, #555));
  }
  :global(.st-fade-gradient) {
    background: linear-gradient(to left, var(--dbg, var(--bos-card, #fff)), transparent);
  }
</style>
