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
</script>

<div class="h-full flex flex-col st-page-bg">
	<!-- Header -->
	<div class="px-6 py-3 st-page-header">
		<h1 class="st-title">Settings</h1>
		<p class="st-muted mt-0">Manage your account and preferences</p>
	</div>

	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 st-spinner rounded-full"></div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto">
			<div class="max-w-3xl mx-auto px-6 py-4">
				<!-- Tab Navigation -->
				<div class="mb-4">
					<div class="st-tab-row">
						{#each tabs as tab}
							{#if !tab.desktopOnly || isDesktop}
								<button
									onclick={() => (activeTab = tab.id)}
									class="st-tab"
									class:st-tab--active={activeTab === tab.id}
								>
									{tab.label}
								</button>
							{/if}
						{/each}
						{#each externalTabs as extTab}
							<button
								onclick={() => goto(extTab.href)}
								class="st-tab"
							>
								{extTab.label}
							</button>
						{/each}
					</div>
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
  :global(.st-page-bg) { background: var(--dbg); }
  :global(.st-page-header) { border-bottom: 1px solid var(--dbd); }
  :global(.st-title) {
    color: var(--dt);
    font-size: 1.25rem;
    font-weight: var(--font-semibold);
  }
  :global(.st-muted) {
    color: var(--dt3);
    font-size: var(--text-sm);
  }
  :global(.st-spinner) {
    border-color: var(--dbd);
    border-top-color: var(--dt);
  }

  .st-tab-row {
    display: flex;
    gap: 4px;
    border-bottom: 1px solid var(--dbd);
    overflow-x: auto;
    scrollbar-width: none;
  }
  .st-tab-row::-webkit-scrollbar { display: none; }

  .st-tab {
    padding: 8px 14px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--dt3);
    border-bottom: 2px solid transparent;
    transition: all var(--bos-transition-fast);
    cursor: pointer;
    white-space: nowrap;
    background: none;
    border-top: none;
    border-left: none;
    border-right: none;
    flex-shrink: 0;
    margin-bottom: -1px;
  }
  .st-tab:hover {
    color: var(--dt2);
    background: var(--bos-hover-color);
    border-radius: 6px 6px 0 0;
  }
  .st-tab--active {
    color: var(--dt);
    font-weight: 600;
    border-bottom-color: var(--dt);
  }
</style>
