<script lang="ts">
	import { api, type UserSettings, type SystemInfo, type GoogleConnectionStatus, type UsageSummary, type ProviderUsage, type ModelUsage, type UsageTrendPoint } from '$lib/api';
	import { useSession, signOut } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/themeStore';
	import { learning } from '$lib/stores/learning';
	import type { PersonalizationProfile, DetectedPattern } from '$lib/api/learning';
	import UserFactsPanel from '$lib/components/settings/UserFactsPanel.svelte';

	const session = useSession();

	let settings = $state<UserSettings | null>(null);
	let systemInfo = $state<SystemInfo | null>(null);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let saveMessage = $state('');
	let activeTab = $state<'general' | 'ai' | 'notifications' | 'integrations' | 'account' | 'desktop' | 'usage' | 'voice' | 'personalization'>('general');

	// Personalization state
	let personalizationProfile = $state<PersonalizationProfile | null>(null);
	let detectedPatterns = $state<DetectedPattern[]>([]);
	let isLoadingPersonalization = $state(false);
	let isSavingPersonalization = $state(false);

	// Usage analytics state
	let usageSummary = $state<UsageSummary | null>(null);
	let usageByProvider = $state<ProviderUsage[]>([]);
	let usageByModel = $state<ModelUsage[]>([]);
	let usageTrend = $state<UsageTrendPoint[]>([]);
	let usagePeriod = $state<'today' | 'week' | 'month' | 'all'>('month');
	let isLoadingUsage = $state(false);

	// Desktop/Electron state
	let isDesktop = $state(false);
	let accessibilityGranted = $state(false);
	let shortcuts = $state<{ quickChat: string; spotlight: string; voiceInput: string }>({
		quickChat: 'CommandOrControl+Shift+Space',
		spotlight: 'CommandOrControl+Space',
		voiceInput: 'CommandOrControl+D',
	});
	let isCheckingPermissions = $state(false);
	let editingShortcut = $state<string | null>(null);

	// Google OAuth state
	let googleStatus = $state<GoogleConnectionStatus | null>(null);
	let isConnectingGoogle = $state(false);
	let isDisconnectingGoogle = $state(false);
	let googleMessage = $state('');

	// Form state
	let selectedModel = $state('');
	let theme = $state('light');
	let emailNotifications = $state(true);
	let dailySummary = $state(false);
	let shareAnalytics = $state(true);

	onMount(async () => {
		await loadSettings();
		await loadSystemInfo();
		await loadGoogleStatus();

		// Check if running in Electron
		if (typeof window !== 'undefined' && (window as any).electron) {
			isDesktop = true;
			await loadDesktopSettings();
		}

		isLoading = false;

		// Check for OAuth callback messages
		const url = new URL(window.location.href);
		if (url.searchParams.get('google_connected') === 'true') {
			activeTab = 'integrations';
			googleMessage = 'Google Calendar connected successfully!';
			setTimeout(() => (googleMessage = ''), 3000);
			// Clean up URL
			url.searchParams.delete('google_connected');
			window.history.replaceState({}, '', url.toString());
		}
		if (url.searchParams.get('google_error')) {
			activeTab = 'integrations';
			googleMessage = `Error: ${url.searchParams.get('google_error')}`;
			setTimeout(() => (googleMessage = ''), 5000);
			// Clean up URL
			url.searchParams.delete('google_error');
			window.history.replaceState({}, '', url.toString());
		}
	});

	async function loadDesktopSettings() {
		try {
			const electron = (window as any).electron;
			if (electron?.shortcuts) {
				const result = await electron.shortcuts.checkAccessibility();
				accessibilityGranted = result?.granted ?? false;

				const shortcutsResult = await electron.shortcuts.get();
				if (shortcutsResult) {
					shortcuts = shortcutsResult;
				}
			}
		} catch (error) {
			console.error('Error loading desktop settings:', error);
		}
	}

	async function checkAccessibility() {
		isCheckingPermissions = true;
		try {
			const electron = (window as any).electron;
			if (electron?.shortcuts) {
				const result = await electron.shortcuts.checkAccessibility();
				accessibilityGranted = result?.granted ?? false;
			}
		} catch (error) {
			console.error('Error checking accessibility:', error);
		}
		isCheckingPermissions = false;
	}

	async function requestAccessibility() {
		try {
			const electron = (window as any).electron;
			if (electron?.shortcuts) {
				await electron.shortcuts.requestAccessibility();
				// Give a moment for the dialog to open, then recheck
				setTimeout(checkAccessibility, 1000);
			}
		} catch (error) {
			console.error('Error requesting accessibility:', error);
		}
	}

	async function openSystemPreferences(pane: string) {
		try {
			const electron = (window as any).electron;
			if (electron?.shell) {
				// Open specific System Preferences pane on macOS
				const urls: Record<string, string> = {
					accessibility: 'x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility',
					screenRecording: 'x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture',
					microphone: 'x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone',
				};
				await electron.shell.openExternal(urls[pane] || 'x-apple.systempreferences:');
			}
		} catch (error) {
			console.error('Error opening system preferences:', error);
		}
	}

	async function resetShortcuts() {
		try {
			const electron = (window as any).electron;
			if (electron?.shortcuts) {
				const result = await electron.shortcuts.reset();
				if (result?.shortcuts) {
					shortcuts = result.shortcuts;
				}
			}
		} catch (error) {
			console.error('Error resetting shortcuts:', error);
		}
	}

	function formatShortcut(shortcut: string): string {
		return shortcut
			.replace('CommandOrControl', '⌘')
			.replace('Command', '⌘')
			.replace('Control', '⌃')
			.replace('Shift', '⇧')
			.replace('Alt', '⌥')
			.replace('Option', '⌥')
			.replace(/\+/g, ' ');
	}

	async function loadSettings() {
		try {
			settings = await api.getSettings();
			if (settings) {
				selectedModel = settings.default_model || '';
				theme = settings.theme;
				emailNotifications = settings.email_notifications;
				dailySummary = settings.daily_summary;
				shareAnalytics = settings.share_analytics;
			}
		} catch (error) {
			console.error('Error loading settings:', error);
		}
	}

	async function loadSystemInfo() {
		try {
			systemInfo = await api.getSystemInfo();
			if (!selectedModel && systemInfo) {
				selectedModel = systemInfo.default_model;
			}
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

	async function loadUsageData() {
		isLoadingUsage = true;
		try {
			const [summary, providers, models, trend] = await Promise.all([
				api.getUsageSummary(usagePeriod),
				api.getUsageByProvider(usagePeriod === 'all' ? 'year' : usagePeriod),
				api.getUsageByModel(usagePeriod === 'all' ? 'year' : usagePeriod),
				api.getUsageTrend()
			]);
			usageSummary = summary;
			usageByProvider = providers;
			usageByModel = models;
			usageTrend = trend;
		} catch (error) {
			console.error('Error loading usage data:', error);
		} finally {
			isLoadingUsage = false;
		}
	}

	async function loadPersonalizationData() {
		isLoadingPersonalization = true;
		try {
			const [profile, patterns] = await Promise.all([
				learning.loadProfile(),
				learning.detectPatterns()
			]);
			personalizationProfile = profile;
			detectedPatterns = patterns;
		} catch (error) {
			console.error('Error loading personalization data:', error);
		} finally {
			isLoadingPersonalization = false;
		}
	}

	async function savePersonalizationProfile() {
		if (!personalizationProfile) return;
		isSavingPersonalization = true;
		try {
			await learning.updateProfile(personalizationProfile);
			saveMessage = 'Personalization settings saved!';
			setTimeout(() => (saveMessage = ''), 2000);
		} catch (error) {
			console.error('Error saving personalization:', error);
			saveMessage = 'Error saving personalization settings';
		} finally {
			isSavingPersonalization = false;
		}
	}

	function formatNumber(num: number): string {
		if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
		if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
		return num.toString();
	}

	function formatCost(cost: number): string {
		return '$' + cost.toFixed(4);
	}

	async function connectGoogle() {
		isConnectingGoogle = true;
		googleMessage = '';
		try {
			const response = await api.initiateGoogleAuth();
			// Redirect to Google OAuth
			window.location.href = response.auth_url;
		} catch (error) {
			console.error('Error initiating Google auth:', error);
			googleMessage = 'Failed to initiate Google authentication';
			isConnectingGoogle = false;
		}
	}

	async function disconnectGoogle() {
		if (!confirm('Are you sure you want to disconnect your Google Calendar? Your synced events will remain in BusinessOS.')) {
			return;
		}
		isDisconnectingGoogle = true;
		googleMessage = '';
		try {
			await api.disconnectGoogle();
			googleStatus = { connected: false };
			googleMessage = 'Google Calendar disconnected';
			setTimeout(() => (googleMessage = ''), 3000);
		} catch (error) {
			console.error('Error disconnecting Google:', error);
			googleMessage = 'Failed to disconnect Google account';
		} finally {
			isDisconnectingGoogle = false;
		}
	}

	async function handleSave() {
		isSaving = true;
		saveMessage = '';

		try {
			await api.updateSettings({
				default_model: selectedModel || null,
				theme,
				email_notifications: emailNotifications,
				daily_summary: dailySummary,
				share_analytics: shareAnalytics,
			});
			// Apply theme immediately
			themeStore.setTheme(theme as 'light' | 'dark');
			saveMessage = 'Settings saved!';
			setTimeout(() => (saveMessage = ''), 2000);
		} catch (error) {
			console.error('Error saving settings:', error);
			saveMessage = 'Error saving settings';
		} finally {
			isSaving = false;
		}
	}

	async function handleDeleteAccount() {
		if (confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
			alert('Account deletion is not implemented yet. Please contact support.');
		}
	}

	async function handleLogout() {
		await signOut();
		goto('/login');
	}
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-900">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 dark:border-gray-800">
		<h1 class="text-xl font-semibold text-gray-900 dark:text-white">Settings</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage your account and preferences</p>
	</div>

	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 dark:border-white border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto">
			<div class="max-w-4xl mx-auto p-6">
				<!-- Tab Navigation -->
				<div class="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-700">
					<button
						onclick={() => (activeTab = 'general')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'general'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						General
					</button>
					<button
						onclick={() => (activeTab = 'ai')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'ai'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						AI Settings
					</button>
					<button
						onclick={() => (activeTab = 'notifications')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'notifications'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						Notifications
					</button>
					<button
						onclick={() => (activeTab = 'integrations')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'integrations'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						Integrations
					</button>
					<button
						onclick={() => (activeTab = 'account')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'account'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						Account
					</button>
					<button
						onclick={() => { activeTab = 'usage'; loadUsageData(); }}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'usage'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						Usage
					</button>
					<button
						onclick={() => (activeTab = 'voice')}
						class="px-4 py-2 text-sm font-medium transition-colors flex items-center gap-1.5 {activeTab === 'voice'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
							<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
							<line x1="12" y1="19" x2="12" y2="23"/>
							<line x1="8" y1="23" x2="16" y2="23"/>
						</svg>
						Voice Notes
					</button>
					<button
						onclick={() => { activeTab = 'personalization'; loadPersonalizationData(); }}
						class="px-4 py-2 text-sm font-medium transition-colors flex items-center gap-1.5 {activeTab === 'personalization'
							? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
							: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
					>
						<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
							<circle cx="12" cy="7" r="4"/>
						</svg>
						Personalization
					</button>
					{#if isDesktop}
						<button
							onclick={() => (activeTab = 'desktop')}
							class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'desktop'
								? 'text-gray-900 dark:text-white border-b-2 border-gray-900 dark:border-white'
								: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
						>
							Desktop
						</button>
					{/if}
				</div>

				<!-- General Tab -->
				{#if activeTab === 'general'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Appearance</h2>
							<div class="space-y-4">
								<div>
									<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Theme</label>
									<div class="flex gap-3">
										<button
											onclick={() => (theme = 'light')}
											class="flex-1 p-4 rounded-lg border-2 transition-colors {theme === 'light'
												? 'border-gray-900 dark:border-white bg-gray-50 dark:bg-gray-700'
												: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'}"
										>
											<div class="flex items-center gap-3">
												<div class="w-10 h-10 rounded-lg bg-white border border-gray-200 flex items-center justify-center">
													<svg class="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
														<path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z" />
													</svg>
												</div>
												<div class="text-left">
													<p class="font-medium text-gray-900 dark:text-white">Light</p>
													<p class="text-xs text-gray-500 dark:text-gray-400">Default light theme</p>
												</div>
											</div>
										</button>
										<button
											onclick={() => (theme = 'dark')}
											class="flex-1 p-4 rounded-lg border-2 transition-colors {theme === 'dark'
												? 'border-gray-900 dark:border-white bg-gray-50 dark:bg-gray-700'
												: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'}"
										>
											<div class="flex items-center gap-3">
												<div class="w-10 h-10 rounded-lg bg-gray-900 flex items-center justify-center">
													<svg class="w-5 h-5 text-gray-100" fill="currentColor" viewBox="0 0 24 24">
														<path fill-rule="evenodd" d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z" clip-rule="evenodd" />
													</svg>
												</div>
												<div class="text-left">
													<p class="font-medium text-gray-900 dark:text-white">Dark</p>
													<p class="text-xs text-gray-500 dark:text-gray-400">Easy on the eyes</p>
												</div>
											</div>
										</button>
									</div>
								</div>
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Privacy</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900 dark:text-white">Share anonymous analytics</p>
									<p class="text-sm text-gray-500 dark:text-gray-400">Help us improve by sharing usage data</p>
								</div>
								<button
									onclick={() => (shareAnalytics = !shareAnalytics)}
									class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {shareAnalytics
										? 'bg-gray-900 dark:bg-white'
										: 'bg-gray-200 dark:bg-gray-600'}"
								>
									<span
										class="inline-block h-4 w-4 transform rounded-full transition-transform {shareAnalytics
											? 'translate-x-6 bg-white dark:bg-gray-900'
											: 'translate-x-1 bg-white dark:bg-gray-300'}"
									></span>
								</button>
							</div>
						</div>
					</div>
				{/if}

				<!-- AI Settings Tab - Redirect to full page -->
				{#if activeTab === 'ai'}
					<div class="space-y-6">
				<!-- User Facts Management -->
				<div class="card" style="padding: 0; overflow: hidden; max-height: 600px;">
					<UserFactsPanel />
				</div>

				<!-- Advanced AI Settings Link -->
				<div class="card text-center py-6">
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
						Need AI model configuration, providers, or agent settings?
					</p>
					<a href="/settings/ai" class="inline-flex items-center gap-2 btn btn-secondary">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						Advanced AI Settings
					</a>
				</div>

						<!-- Quick status -->
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Current Status</h2>
							<div class="flex items-center gap-3">
								<div class="w-3 h-3 rounded-full {systemInfo?.ollama_mode === 'local' ? 'bg-green-500' : 'bg-blue-500'}"></div>
								<div>
									<p class="font-medium text-gray-900 dark:text-white">
										{#if systemInfo?.ollama_mode === 'local'}
											Local Mode (Ollama)
										{:else if systemInfo?.active_provider === 'groq'}
											Groq Cloud
										{:else if systemInfo?.active_provider === 'anthropic'}
											Claude (Anthropic)
										{:else}
											Cloud Mode
										{/if}
									</p>
									<p class="text-sm text-gray-500 dark:text-gray-400">
										{systemInfo?.ollama_mode === 'local'
											? 'Running AI models locally on your machine'
											: 'Using cloud-hosted AI models'}
									</p>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Notifications Tab -->
				{#if activeTab === 'notifications'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Email Notifications</h2>
							<div class="space-y-4">
								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium text-gray-900 dark:text-white">Email notifications</p>
										<p class="text-sm text-gray-500 dark:text-gray-400">Receive important updates via email</p>
									</div>
									<button
										onclick={() => (emailNotifications = !emailNotifications)}
										class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {emailNotifications
											? 'bg-gray-900 dark:bg-white'
											: 'bg-gray-200 dark:bg-gray-600'}"
									>
										<span
											class="inline-block h-4 w-4 transform rounded-full transition-transform {emailNotifications
												? 'translate-x-6 bg-white dark:bg-gray-900'
												: 'translate-x-1 bg-white dark:bg-gray-300'}"
										></span>
									</button>
								</div>

								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium text-gray-900 dark:text-white">Daily summary</p>
										<p class="text-sm text-gray-500 dark:text-gray-400">Get a daily recap of your activity</p>
									</div>
									<button
										onclick={() => (dailySummary = !dailySummary)}
										class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {dailySummary
											? 'bg-gray-900 dark:bg-white'
											: 'bg-gray-200 dark:bg-gray-600'}"
									>
										<span
											class="inline-block h-4 w-4 transform rounded-full transition-transform {dailySummary
												? 'translate-x-6 bg-white dark:bg-gray-900'
												: 'translate-x-1 bg-white dark:bg-gray-300'}"
										></span>
									</button>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Integrations Tab -->
				{#if activeTab === 'integrations'}
					<div class="space-y-6">
						{#if googleMessage}
							<div class="p-4 rounded-lg {googleMessage.includes('Error') || googleMessage.includes('Failed') ? 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400' : 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400'}">
								{googleMessage}
							</div>
						{/if}

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Google Calendar</h2>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
								Connect your Google Calendar to sync events, see your schedule, and let the AI help plan your tasks around your existing commitments.
							</p>

							{#if googleStatus?.connected}
								<div class="flex items-center justify-between p-4 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800">
									<div class="flex items-center gap-4">
										<div class="w-12 h-12 rounded-full bg-white dark:bg-gray-800 flex items-center justify-center shadow-sm">
											<svg class="w-6 h-6" viewBox="0 0 24 24">
												<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
												<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
												<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
												<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
											</svg>
										</div>
										<div>
											<p class="font-medium text-green-800 dark:text-green-400">Connected</p>
											{#if googleStatus.email}
												<p class="text-sm text-green-600 dark:text-green-500">{googleStatus.email}</p>
											{/if}
											{#if googleStatus.connected_at}
												<p class="text-xs text-green-500 dark:text-green-600">
													Connected {new Date(googleStatus.connected_at).toLocaleDateString()}
												</p>
											{/if}
										</div>
									</div>
									<button
										onclick={disconnectGoogle}
										disabled={isDisconnectingGoogle}
										class="btn btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-700 dark:text-white dark:border-gray-600"
									>
										{#if isDisconnectingGoogle}
											<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											Disconnecting...
										{:else}
											Disconnect
										{/if}
									</button>
								</div>
							{:else}
								<div class="flex items-center justify-between p-4 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
									<div class="flex items-center gap-4">
										<div class="w-12 h-12 rounded-full bg-white dark:bg-gray-700 flex items-center justify-center shadow-sm">
											<svg class="w-6 h-6 text-gray-400" viewBox="0 0 24 24">
												<path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
												<path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
												<path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
												<path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-700 dark:text-gray-200">Not connected</p>
											<p class="text-sm text-gray-500 dark:text-gray-400">Connect to sync your calendar</p>
										</div>
									</div>
									<button
										onclick={connectGoogle}
										disabled={isConnectingGoogle}
										class="btn btn-primary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
									>
										{#if isConnectingGoogle}
											<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											Connecting...
										{:else}
											Connect Google Calendar
										{/if}
									</button>
								</div>
							{/if}

							<div class="mt-4 text-xs text-gray-500 dark:text-gray-400">
								<p class="font-medium mb-1">Permissions requested:</p>
								<ul class="list-disc list-inside space-y-0.5">
									<li>Read your calendar events</li>
									<li>Create and update events (for two-way sync)</li>
									<li>View your email address</li>
								</ul>
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">More Integrations</h2>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
								Additional integrations coming soon. Let us know what you'd like to see!
							</p>
							<div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
								<div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 text-center opacity-50">
									<div class="w-8 h-8 mx-auto mb-2 rounded bg-gray-100 dark:bg-gray-700"></div>
									<p class="text-sm text-gray-500 dark:text-gray-400">Slack</p>
									<span class="text-xs text-gray-400 dark:text-gray-500">Coming soon</span>
								</div>
								<div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 text-center opacity-50">
									<div class="w-8 h-8 mx-auto mb-2 rounded bg-gray-100 dark:bg-gray-700"></div>
									<p class="text-sm text-gray-500 dark:text-gray-400">Notion</p>
									<span class="text-xs text-gray-400 dark:text-gray-500">Coming soon</span>
								</div>
								<div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 text-center opacity-50">
									<div class="w-8 h-8 mx-auto mb-2 rounded bg-gray-100 dark:bg-gray-700"></div>
									<p class="text-sm text-gray-500 dark:text-gray-400">Linear</p>
									<span class="text-xs text-gray-400 dark:text-gray-500">Coming soon</span>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Account Tab -->
				{#if activeTab === 'account'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Account Information</h2>
							<div class="space-y-4">
								<div>
									<label class="block text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Name</label>
									<p class="text-gray-900 dark:text-white">{$session.data?.user?.name || 'Not set'}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Email</label>
									<p class="text-gray-900 dark:text-white">{$session.data?.user?.email || 'Not set'}</p>
								</div>
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Sessions</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900 dark:text-white">Current session</p>
									<p class="text-sm text-gray-500 dark:text-gray-400">You're signed in on this device</p>
								</div>
								<button
									onclick={handleLogout}
									class="btn btn-secondary text-sm dark:bg-gray-700 dark:text-white dark:border-gray-600"
								>
									Sign Out
								</button>
							</div>
						</div>

						<div class="card border-red-200 dark:border-red-900">
							<h2 class="text-lg font-medium text-red-600 dark:text-red-400 mb-4">Danger Zone</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900 dark:text-white">Delete account</p>
									<p class="text-sm text-gray-500 dark:text-gray-400">Permanently delete your account and all data</p>
								</div>
								<button
									onclick={handleDeleteAccount}
									class="btn text-sm bg-red-600 text-white hover:bg-red-700"
								>
									Delete Account
								</button>
							</div>
						</div>
					</div>
				{/if}

				<!-- Usage Tab -->
				{#if activeTab === 'usage'}
					<div class="usage-dashboard">
						<!-- Header with Period Selector -->
						<div class="usage-header">
							<div>
								<h2 class="usage-title">Usage Analytics</h2>
								<p class="usage-subtitle">Track your AI usage and costs</p>
							</div>
							<div class="period-selector">
								{#each ['today', 'week', 'month', 'all'] as period}
									<button
										onclick={() => { usagePeriod = period as typeof usagePeriod; loadUsageData(); }}
										class="period-btn"
										class:active={usagePeriod === period}
									>
										{period === 'all' ? 'All Time' : period.charAt(0).toUpperCase() + period.slice(1)}
									</button>
								{/each}
							</div>
						</div>

						{#if isLoadingUsage}
							<div class="usage-loading">
								<div class="usage-spinner"></div>
								<p>Loading analytics...</p>
							</div>
						{:else if (!usageSummary || usageSummary.total_requests === 0)}
							<!-- Empty State -->
							<div class="usage-empty">
								<div class="usage-empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
									</svg>
								</div>
								<h3>No Usage Data Yet</h3>
								<p>Start chatting with the AI to see your usage analytics here.</p>
							</div>
						{:else}
							<!-- Stats Grid -->
							<div class="stats-grid">
								<div class="stat-card requests">
									<div class="stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
										</svg>
									</div>
									<div class="stat-content">
										<span class="stat-value">{formatNumber(usageSummary?.total_requests || 0)}</span>
										<span class="stat-label">Requests</span>
									</div>
								</div>

								<div class="stat-card tokens">
									<div class="stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
										</svg>
									</div>
									<div class="stat-content">
										<span class="stat-value">{formatNumber(usageSummary?.total_tokens || 0)}</span>
										<span class="stat-label">Total Tokens</span>
									</div>
								</div>

								<div class="stat-card input">
									<div class="stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
										</svg>
									</div>
									<div class="stat-content">
										<span class="stat-value">{formatNumber(usageSummary?.total_input_tokens || 0)}</span>
										<span class="stat-label">Input Tokens</span>
									</div>
								</div>

								<div class="stat-card output">
									<div class="stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
										</svg>
									</div>
									<div class="stat-content">
										<span class="stat-value">{formatNumber(usageSummary?.total_output_tokens || 0)}</span>
										<span class="stat-label">Output Tokens</span>
									</div>
								</div>
							</div>

							<!-- Cost Analysis -->
							<div class="cost-analysis">
								<div class="cost-card spent">
									<div class="cost-header">
										<span class="cost-label">Estimated Cloud Cost</span>
										<span class="cost-badge">API Usage</span>
									</div>
									<span class="cost-value">{formatCost(usageSummary?.total_cost || 0)}</span>
									<p class="cost-note">Based on current provider pricing</p>
								</div>

								<div class="cost-card saved">
									<div class="cost-header">
										<span class="cost-label">Local Processing Savings</span>
										<span class="cost-badge saved">Saved</span>
									</div>
									<span class="cost-value">
										{formatCost((usageByProvider.find(p => p.provider === 'ollama')?.total_tokens || 0) * 0.00002)}
									</span>
									<p class="cost-note">Running {formatNumber(usageByProvider.find(p => p.provider === 'ollama')?.total_tokens || 0)} tokens locally</p>
								</div>
							</div>

							<!-- Provider Breakdown -->
							{#if usageByProvider.length > 0}
								<div class="usage-section">
									<h3 class="section-title">Usage by Provider</h3>
									<div class="provider-list">
										{#each usageByProvider as provider}
											{@const maxTokens = Math.max(...usageByProvider.map(p => p.total_tokens))}
											{@const percentage = maxTokens > 0 ? (provider.total_tokens / maxTokens) * 100 : 0}
											<div class="provider-item">
												<div class="provider-info">
													<div class="provider-icon" class:local={provider.provider === 'ollama'} class:anthropic={provider.provider === 'anthropic'} class:groq={provider.provider === 'groq'}>
														{#if provider.provider === 'ollama'}
															<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<rect x="2" y="3" width="20" height="14" rx="2"/>
																<path d="M8 21h8M12 17v4"/>
															</svg>
														{:else}
															<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
															</svg>
														{/if}
													</div>
													<div class="provider-details">
														<span class="provider-name">{provider.provider}</span>
														<span class="provider-type">{provider.provider === 'ollama' ? 'Local' : 'Cloud'}</span>
													</div>
												</div>
												<div class="provider-stats">
													<div class="provider-bar-container">
														<div class="provider-bar" class:local={provider.provider === 'ollama'} style="width: {percentage}%"></div>
													</div>
													<div class="provider-numbers">
														<span class="provider-tokens">{formatNumber(provider.total_tokens)} tokens</span>
														<span class="provider-cost">{provider.provider === 'ollama' ? 'Free' : formatCost(provider.total_cost)}</span>
													</div>
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Model Usage -->
							{#if usageByModel.length > 0}
								<div class="usage-section">
									<h3 class="section-title">Model Usage</h3>
									<div class="model-grid">
										{#each usageByModel.slice(0, 6) as model}
											<div class="model-card">
												<div class="model-header">
													<span class="model-name">{model.model.split(':')[0]}</span>
													<span class="model-provider" class:local={model.provider === 'ollama'}>{model.provider}</span>
												</div>
												<div class="model-stats">
													<div class="model-stat">
														<span class="model-stat-value">{formatNumber(model.request_count)}</span>
														<span class="model-stat-label">requests</span>
													</div>
													<div class="model-stat">
														<span class="model-stat-value">{formatNumber(model.total_tokens)}</span>
														<span class="model-stat-label">tokens</span>
													</div>
												</div>
												<div class="model-cost">
													{model.provider === 'ollama' ? 'Free (Local)' : formatCost(model.total_cost)}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						{/if}
					</div>

					<style>
						.usage-dashboard {
							display: flex;
							flex-direction: column;
							gap: 24px;
						}

						.usage-header {
							display: flex;
							justify-content: space-between;
							align-items: flex-start;
							flex-wrap: wrap;
							gap: 16px;
						}

						.usage-title {
							font-size: 1.5rem;
							font-weight: 600;
							color: var(--color-text);
							margin: 0;
						}

						.usage-subtitle {
							font-size: 0.875rem;
							color: var(--color-text-secondary);
							margin-top: 4px;
						}

						.period-selector {
							display: flex;
							background: var(--color-bg-secondary);
							border-radius: 10px;
							padding: 4px;
							gap: 4px;
						}

						.period-btn {
							padding: 8px 16px;
							border: none;
							background: transparent;
							border-radius: 8px;
							font-size: 0.875rem;
							font-weight: 500;
							color: var(--color-text-secondary);
							cursor: pointer;
							transition: all 0.15s;
						}

						.period-btn:hover {
							color: var(--color-text);
						}

						.period-btn.active {
							background: var(--color-bg);
							color: var(--color-text);
							box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
						}

						:global(.dark) .period-btn.active {
							background: #1f1f1f;
							box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
						}

						.usage-loading {
							display: flex;
							flex-direction: column;
							align-items: center;
							justify-content: center;
							padding: 80px 20px;
							gap: 16px;
							color: var(--color-text-secondary);
						}

						.usage-spinner {
							width: 32px;
							height: 32px;
							border: 2px solid var(--color-border);
							border-top-color: var(--color-text);
							border-radius: 50%;
							animation: spin 0.8s linear infinite;
						}

						@keyframes spin {
							to { transform: rotate(360deg); }
						}

						.usage-empty {
							display: flex;
							flex-direction: column;
							align-items: center;
							justify-content: center;
							padding: 80px 20px;
							text-align: center;
						}

						.usage-empty-icon {
							width: 64px;
							height: 64px;
							border-radius: 16px;
							background: var(--color-bg-secondary);
							display: flex;
							align-items: center;
							justify-content: center;
							margin-bottom: 16px;
						}

						.usage-empty-icon svg {
							width: 32px;
							height: 32px;
							color: var(--color-text-muted);
						}

						.usage-empty h3 {
							font-size: 1.125rem;
							font-weight: 600;
							color: var(--color-text);
							margin: 0 0 8px;
						}

						.usage-empty p {
							font-size: 0.875rem;
							color: var(--color-text-secondary);
							margin: 0;
						}

						.stats-grid {
							display: grid;
							grid-template-columns: repeat(4, 1fr);
							gap: 16px;
						}

						@media (max-width: 768px) {
							.stats-grid { grid-template-columns: repeat(2, 1fr); }
						}

						.stat-card {
							background: var(--color-bg);
							border: 1px solid var(--color-border);
							border-radius: 16px;
							padding: 20px;
							display: flex;
							align-items: center;
							gap: 16px;
						}

						:global(.dark) .stat-card {
							background: #0a0a0a;
							border-color: rgba(255, 255, 255, 0.08);
						}

						.stat-icon {
							width: 48px;
							height: 48px;
							border-radius: 12px;
							display: flex;
							align-items: center;
							justify-content: center;
							flex-shrink: 0;
						}

						.stat-icon svg {
							width: 24px;
							height: 24px;
						}

						.stat-card.requests .stat-icon { background: #dbeafe; color: #2563eb; }
						.stat-card.tokens .stat-icon { background: #f3e8ff; color: #9333ea; }
						.stat-card.input .stat-icon { background: #dcfce7; color: #16a34a; }
						.stat-card.output .stat-icon { background: #fef3c7; color: #d97706; }

						:global(.dark) .stat-card.requests .stat-icon { background: rgba(37, 99, 235, 0.2); }
						:global(.dark) .stat-card.tokens .stat-icon { background: rgba(147, 51, 234, 0.2); }
						:global(.dark) .stat-card.input .stat-icon { background: rgba(22, 163, 74, 0.2); }
						:global(.dark) .stat-card.output .stat-icon { background: rgba(217, 119, 6, 0.2); }

						.stat-content {
							display: flex;
							flex-direction: column;
						}

						.stat-value {
							font-size: 1.75rem;
							font-weight: 700;
							color: var(--color-text);
							line-height: 1;
						}

						.stat-label {
							font-size: 0.75rem;
							color: var(--color-text-muted);
							margin-top: 4px;
							text-transform: uppercase;
							letter-spacing: 0.5px;
						}

						.cost-analysis {
							display: grid;
							grid-template-columns: 1fr 1fr;
							gap: 16px;
						}

						@media (max-width: 640px) {
							.cost-analysis { grid-template-columns: 1fr; }
						}

						.cost-card {
							background: var(--color-bg);
							border: 1px solid var(--color-border);
							border-radius: 16px;
							padding: 24px;
						}

						:global(.dark) .cost-card {
							background: #0a0a0a;
							border-color: rgba(255, 255, 255, 0.08);
						}

						.cost-card.spent {
							border-left: 4px solid #6366f1;
						}

						.cost-card.saved {
							border-left: 4px solid #10b981;
						}

						.cost-header {
							display: flex;
							justify-content: space-between;
							align-items: center;
							margin-bottom: 12px;
						}

						.cost-label {
							font-size: 0.875rem;
							color: var(--color-text-secondary);
						}

						.cost-badge {
							font-size: 0.625rem;
							font-weight: 600;
							text-transform: uppercase;
							letter-spacing: 0.5px;
							padding: 4px 8px;
							border-radius: 4px;
							background: #e0e7ff;
							color: #4338ca;
						}

						.cost-badge.saved {
							background: #d1fae5;
							color: #047857;
						}

						:global(.dark) .cost-badge {
							background: rgba(99, 102, 241, 0.2);
							color: #a5b4fc;
						}

						:global(.dark) .cost-badge.saved {
							background: rgba(16, 185, 129, 0.2);
							color: #6ee7b7;
						}

						.cost-value {
							font-size: 2.5rem;
							font-weight: 700;
							color: var(--color-text);
							display: block;
						}

						.cost-card.saved .cost-value {
							color: #10b981;
						}

						.cost-note {
							font-size: 0.75rem;
							color: var(--color-text-muted);
							margin-top: 8px;
						}

						.usage-section {
							background: var(--color-bg);
							border: 1px solid var(--color-border);
							border-radius: 16px;
							padding: 24px;
						}

						:global(.dark) .usage-section {
							background: #0a0a0a;
							border-color: rgba(255, 255, 255, 0.08);
						}

						.section-title {
							font-size: 1rem;
							font-weight: 600;
							color: var(--color-text);
							margin: 0 0 20px;
						}

						.provider-list {
							display: flex;
							flex-direction: column;
							gap: 16px;
						}

						.provider-item {
							display: flex;
							align-items: center;
							justify-content: space-between;
							gap: 24px;
						}

						.provider-info {
							display: flex;
							align-items: center;
							gap: 12px;
							min-width: 140px;
						}

						.provider-icon {
							width: 40px;
							height: 40px;
							border-radius: 10px;
							display: flex;
							align-items: center;
							justify-content: center;
							background: #e0e7ff;
							color: #4338ca;
						}

						.provider-icon.local {
							background: #d1fae5;
							color: #047857;
						}

						.provider-icon.anthropic {
							background: #fed7aa;
							color: #c2410c;
						}

						.provider-icon.groq {
							background: #dbeafe;
							color: #1d4ed8;
						}

						:global(.dark) .provider-icon {
							background: rgba(99, 102, 241, 0.2);
						}

						:global(.dark) .provider-icon.local {
							background: rgba(16, 185, 129, 0.2);
						}

						:global(.dark) .provider-icon.anthropic {
							background: rgba(194, 65, 12, 0.2);
						}

						.provider-icon svg {
							width: 20px;
							height: 20px;
						}

						.provider-details {
							display: flex;
							flex-direction: column;
						}

						.provider-name {
							font-weight: 600;
							color: var(--color-text);
							text-transform: capitalize;
						}

						.provider-type {
							font-size: 0.75rem;
							color: var(--color-text-muted);
						}

						.provider-stats {
							flex: 1;
							display: flex;
							flex-direction: column;
							gap: 8px;
						}

						.provider-bar-container {
							height: 8px;
							background: var(--color-bg-secondary);
							border-radius: 4px;
							overflow: hidden;
						}

						:global(.dark) .provider-bar-container {
							background: #1f1f1f;
						}

						.provider-bar {
							height: 100%;
							background: linear-gradient(90deg, #6366f1, #8b5cf6);
							border-radius: 4px;
							transition: width 0.3s ease;
						}

						.provider-bar.local {
							background: linear-gradient(90deg, #10b981, #34d399);
						}

						.provider-numbers {
							display: flex;
							justify-content: space-between;
						}

						.provider-tokens {
							font-size: 0.875rem;
							font-weight: 500;
							color: var(--color-text);
						}

						.provider-cost {
							font-size: 0.875rem;
							color: var(--color-text-secondary);
						}

						.model-grid {
							display: grid;
							grid-template-columns: repeat(3, 1fr);
							gap: 12px;
						}

						@media (max-width: 768px) {
							.model-grid { grid-template-columns: repeat(2, 1fr); }
						}

						@media (max-width: 480px) {
							.model-grid { grid-template-columns: 1fr; }
						}

						.model-card {
							background: var(--color-bg-secondary);
							border-radius: 12px;
							padding: 16px;
						}

						:global(.dark) .model-card {
							background: #141414;
						}

						.model-header {
							display: flex;
							justify-content: space-between;
							align-items: flex-start;
							margin-bottom: 12px;
						}

						.model-name {
							font-weight: 600;
							color: var(--color-text);
							font-size: 0.875rem;
						}

						.model-provider {
							font-size: 0.625rem;
							font-weight: 500;
							text-transform: uppercase;
							padding: 2px 6px;
							border-radius: 4px;
							background: #e0e7ff;
							color: #4338ca;
						}

						.model-provider.local {
							background: #d1fae5;
							color: #047857;
						}

						:global(.dark) .model-provider {
							background: rgba(99, 102, 241, 0.2);
							color: #a5b4fc;
						}

						:global(.dark) .model-provider.local {
							background: rgba(16, 185, 129, 0.2);
							color: #6ee7b7;
						}

						.model-stats {
							display: flex;
							gap: 16px;
							margin-bottom: 8px;
						}

						.model-stat {
							display: flex;
							flex-direction: column;
						}

						.model-stat-value {
							font-size: 1.125rem;
							font-weight: 700;
							color: var(--color-text);
						}

						.model-stat-label {
							font-size: 0.625rem;
							color: var(--color-text-muted);
							text-transform: uppercase;
						}

						.model-cost {
							font-size: 0.75rem;
							color: var(--color-text-secondary);
						}
					</style>
				{/if}

				<!-- Voice Notes Tab -->
				{#if activeTab === 'voice'}
					<div class="space-y-6">
						<div class="card text-center py-8">
							<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
								<svg class="w-8 h-8 text-green-600 dark:text-green-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
									<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
									<line x1="12" y1="19" x2="12" y2="23"/>
									<line x1="8" y1="23" x2="16" y2="23"/>
								</svg>
							</div>
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Voice Notes</h2>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto">
								View your voice transcription history, track speaking stats, and browse notes by date or project.
							</p>
							<a
								href="/voice-notes"
								class="inline-flex items-center gap-2 btn btn-primary"
							>
								<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
									<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
								</svg>
								View Voice Notes
							</a>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">About Voice Notes</h2>
							<div class="space-y-4 text-sm text-gray-600 dark:text-gray-400">
								<div class="flex items-start gap-3">
									<div class="w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="10"/>
											<polyline points="12 6 12 12 16 14"/>
										</svg>
									</div>
									<div>
										<p class="font-medium text-gray-900 dark:text-white">Track Your Speaking</p>
										<p>See duration, word count, and words per minute for each recording.</p>
									</div>
								</div>
								<div class="flex items-start gap-3">
									<div class="w-8 h-8 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
											<line x1="16" y1="2" x2="16" y2="6"/>
											<line x1="8" y1="2" x2="8" y2="6"/>
											<line x1="3" y1="10" x2="21" y2="10"/>
										</svg>
									</div>
									<div>
										<p class="font-medium text-gray-900 dark:text-white">Browse by Date</p>
										<p>Notes organized by today, yesterday, and older dates for easy review.</p>
									</div>
								</div>
								<div class="flex items-start gap-3">
									<div class="w-8 h-8 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4 text-orange-600 dark:text-orange-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
										</svg>
									</div>
									<div>
										<p class="font-medium text-gray-900 dark:text-white">Filter by Project</p>
										<p>View notes associated with specific projects and contexts.</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Personalization Tab -->
				{#if activeTab === 'personalization'}
					<div class="space-y-6">
						{#if isLoadingPersonalization}
							<div class="flex items-center justify-center py-12">
								<div class="animate-spin h-8 w-8 border-2 border-gray-900 dark:border-white border-t-transparent rounded-full"></div>
							</div>
						{:else}
							<!-- AI Response Preferences -->
							<div class="card">
								<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Response Preferences</h2>
								<p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
									Customize how the AI responds to you. These preferences help personalize your experience.
								</p>

								<div class="space-y-6">
									<!-- Tone -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Preferred Tone</label>
										<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
											{#each ['formal', 'professional', 'casual', 'friendly'] as tone}
												<button
													onclick={() => personalizationProfile && (personalizationProfile.preferred_tone = tone as any)}
													class="p-3 rounded-lg border-2 text-sm font-medium transition-colors {personalizationProfile?.preferred_tone === tone
														? 'border-gray-900 dark:border-white bg-gray-50 dark:bg-gray-700'
														: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'}"
												>
													{tone.charAt(0).toUpperCase() + tone.slice(1)}
												</button>
											{/each}
										</div>
									</div>

									<!-- Verbosity -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Response Length</label>
										<div class="grid grid-cols-3 gap-3">
											{#each ['concise', 'balanced', 'detailed'] as verbosity}
												<button
													onclick={() => personalizationProfile && (personalizationProfile.preferred_verbosity = verbosity as any)}
													class="p-3 rounded-lg border-2 text-sm font-medium transition-colors {personalizationProfile?.preferred_verbosity === verbosity
														? 'border-gray-900 dark:border-white bg-gray-50 dark:bg-gray-700'
														: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'}"
												>
													{verbosity.charAt(0).toUpperCase() + verbosity.slice(1)}
												</button>
											{/each}
										</div>
									</div>

									<!-- Format -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Response Format</label>
										<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
											{#each ['prose', 'bullets', 'structured', 'mixed'] as format}
												<button
													onclick={() => personalizationProfile && (personalizationProfile.preferred_format = format as any)}
													class="p-3 rounded-lg border-2 text-sm font-medium transition-colors {personalizationProfile?.preferred_format === format
														? 'border-gray-900 dark:border-white bg-gray-50 dark:bg-gray-700'
														: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'}"
												>
													{format.charAt(0).toUpperCase() + format.slice(1)}
												</button>
											{/each}
										</div>
									</div>

									<!-- Toggles -->
									<div class="space-y-4">
										<div class="flex items-center justify-between">
											<div>
												<p class="font-medium text-gray-900 dark:text-white">Include Examples</p>
												<p class="text-sm text-gray-500 dark:text-gray-400">AI will include relevant examples in responses</p>
											</div>
											<button
												onclick={() => personalizationProfile && (personalizationProfile.prefers_examples = !personalizationProfile.prefers_examples)}
												class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {personalizationProfile?.prefers_examples
													? 'bg-gray-900 dark:bg-white'
													: 'bg-gray-200 dark:bg-gray-600'}"
											>
												<span
													class="inline-block h-4 w-4 transform rounded-full transition-transform {personalizationProfile?.prefers_examples
														? 'translate-x-6 bg-white dark:bg-gray-900'
														: 'translate-x-1 bg-white dark:bg-gray-300'}"
												></span>
											</button>
										</div>

										<div class="flex items-center justify-between">
											<div>
												<p class="font-medium text-gray-900 dark:text-white">Include Code Samples</p>
												<p class="text-sm text-gray-500 dark:text-gray-400">AI will include code snippets when relevant</p>
											</div>
											<button
												onclick={() => personalizationProfile && (personalizationProfile.prefers_code_samples = !personalizationProfile.prefers_code_samples)}
												class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {personalizationProfile?.prefers_code_samples
													? 'bg-gray-900 dark:bg-white'
													: 'bg-gray-200 dark:bg-gray-600'}"
											>
												<span
													class="inline-block h-4 w-4 transform rounded-full transition-transform {personalizationProfile?.prefers_code_samples
														? 'translate-x-6 bg-white dark:bg-gray-900'
														: 'translate-x-1 bg-white dark:bg-gray-300'}"
												></span>
											</button>
										</div>
									</div>
								</div>
							</div>

							<!-- Detected Patterns -->
							{#if detectedPatterns.length > 0}
								<div class="card">
									<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Detected Patterns</h2>
									<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
										The AI has learned these patterns from your interactions.
									</p>
									<div class="space-y-3">
										{#each detectedPatterns as pattern}
											<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800">
												<div>
													<p class="font-medium text-gray-900 dark:text-white text-sm">{pattern.pattern_key}</p>
													<p class="text-xs text-gray-500 dark:text-gray-400">{pattern.pattern_value}</p>
												</div>
												<div class="flex items-center gap-2">
													<span class="text-xs text-gray-400 dark:text-gray-500">{pattern.observation_count} observations</span>
													<span class="px-2 py-1 text-xs font-medium rounded-full {pattern.confidence_score >= 0.8 ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : pattern.confidence_score >= 0.5 ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
														{Math.round(pattern.confidence_score * 100)}%
													</span>
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Expertise & Learning Areas -->
							<div class="card">
								<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Knowledge Areas</h2>
								<div class="space-y-4">
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Your Expertise</label>
										<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">Areas where you have strong knowledge</p>
										<div class="flex flex-wrap gap-2">
											{#each personalizationProfile?.expertise_areas || [] as area}
												<span class="px-3 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400 rounded-full text-sm">
													{area}
												</span>
											{/each}
											{#if !personalizationProfile?.expertise_areas?.length}
												<span class="text-sm text-gray-400 dark:text-gray-500 italic">No expertise areas detected yet</span>
											{/if}
										</div>
									</div>

									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Learning Interests</label>
										<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">Topics you're actively learning about</p>
										<div class="flex flex-wrap gap-2">
											{#each personalizationProfile?.learning_areas || [] as area}
												<span class="px-3 py-1 bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400 rounded-full text-sm">
													{area}
												</span>
											{/each}
											{#if !personalizationProfile?.learning_areas?.length}
												<span class="text-sm text-gray-400 dark:text-gray-500 italic">No learning areas detected yet</span>
											{/if}
										</div>
									</div>
								</div>
							</div>

							<!-- Save Button -->
							<div class="flex justify-end">
								<button
									onclick={savePersonalizationProfile}
									disabled={isSavingPersonalization || !personalizationProfile}
									class="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
								>
									{#if isSavingPersonalization}
										<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										Saving...
									{:else}
										Save Preferences
									{/if}
								</button>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Desktop Tab (Electron only) -->
				{#if activeTab === 'desktop' && isDesktop}
					<div class="space-y-6">
						<!-- System Permissions -->
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">System Permissions</h2>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
								BusinessOS requires certain system permissions for features like global shortcuts, screenshot capture, and voice input.
							</p>

							<div class="space-y-4">
								<!-- Accessibility -->
								<div class="flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700">
									<div class="flex items-center gap-4">
										<div class="w-10 h-10 rounded-lg {accessibilityGranted ? 'bg-green-100 dark:bg-green-900/30' : 'bg-yellow-100 dark:bg-yellow-900/30'} flex items-center justify-center">
											<svg class="w-5 h-5 {accessibilityGranted ? 'text-green-600 dark:text-green-400' : 'text-yellow-600 dark:text-yellow-400'}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
												<path stroke-linecap="round" stroke-linejoin="round" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900 dark:text-white">Accessibility</p>
											<p class="text-sm text-gray-500 dark:text-gray-400">
												{accessibilityGranted ? 'Global shortcuts enabled' : 'Required for global keyboard shortcuts'}
											</p>
										</div>
									</div>
									<div class="flex items-center gap-2">
										{#if accessibilityGranted}
											<span class="flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400">
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
													<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
												</svg>
												Enabled
											</span>
										{:else}
											<button
												onclick={requestAccessibility}
												class="btn btn-primary text-sm"
											>
												Enable
											</button>
										{/if}
										<button
											onclick={() => openSystemPreferences('accessibility')}
											class="btn btn-secondary text-sm dark:bg-gray-700 dark:text-white dark:border-gray-600"
										>
											Open Settings
										</button>
									</div>
								</div>

								<!-- Screen Recording -->
								<div class="flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700">
									<div class="flex items-center gap-4">
										<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
											<svg class="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
												<path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900 dark:text-white">Screen Recording</p>
											<p class="text-sm text-gray-500 dark:text-gray-400">Required for screenshot capture</p>
										</div>
									</div>
									<button
										onclick={() => openSystemPreferences('screenRecording')}
										class="btn btn-secondary text-sm dark:bg-gray-700 dark:text-white dark:border-gray-600"
									>
										Open Settings
									</button>
								</div>

								<!-- Microphone -->
								<div class="flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700">
									<div class="flex items-center gap-4">
										<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
											<svg class="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
												<path stroke-linecap="round" stroke-linejoin="round" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900 dark:text-white">Microphone</p>
											<p class="text-sm text-gray-500 dark:text-gray-400">Required for voice input and meeting recording</p>
										</div>
									</div>
									<button
										onclick={() => openSystemPreferences('microphone')}
										class="btn btn-secondary text-sm dark:bg-gray-700 dark:text-white dark:border-gray-600"
									>
										Open Settings
									</button>
								</div>
							</div>
						</div>

						<!-- Keyboard Shortcuts -->
						<div class="card">
							<div class="flex items-center justify-between mb-4">
								<h2 class="text-lg font-medium text-gray-900 dark:text-white">Keyboard Shortcuts</h2>
								<button
									onclick={resetShortcuts}
									class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
								>
									Reset to defaults
								</button>
							</div>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
								Global shortcuts work system-wide, even when the app is in the background.
							</p>

							<div class="space-y-3">
								<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800">
									<div>
										<p class="font-medium text-gray-900 dark:text-white text-sm">Quick Chat</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">Open chat popup from anywhere</p>
									</div>
									<div class="flex items-center gap-2 font-mono text-sm bg-white dark:bg-gray-700 px-3 py-1.5 rounded border border-gray-200 dark:border-gray-600">
										{formatShortcut(shortcuts.quickChat)}
									</div>
								</div>

								<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800">
									<div>
										<p class="font-medium text-gray-900 dark:text-white text-sm">Voice Input</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">Start voice dictation</p>
									</div>
									<div class="flex items-center gap-2 font-mono text-sm bg-white dark:bg-gray-700 px-3 py-1.5 rounded border border-gray-200 dark:border-gray-600">
										{formatShortcut(shortcuts.voiceInput)}
									</div>
								</div>
							</div>

							{#if !accessibilityGranted}
								<div class="mt-4 p-3 rounded-lg bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800">
									<p class="text-sm text-yellow-800 dark:text-yellow-400">
										Enable Accessibility permission above to use global shortcuts.
									</p>
								</div>
							{/if}
						</div>
					</div>
				{/if}

				<!-- Save Button -->
				{#if activeTab !== 'account' && activeTab !== 'integrations' && activeTab !== 'desktop' && activeTab !== 'voice'}
					<div class="mt-6 flex items-center justify-between">
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{saveMessage || 'Changes are saved automatically when you click Save'}
						</p>
						<button
							onclick={handleSave}
							disabled={isSaving}
							class="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{#if isSaving}
								<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Saving...
							{:else}
								Save Changes
							{/if}
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
