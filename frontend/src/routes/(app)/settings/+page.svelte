<script lang="ts">
	import { api, type UserSettings, type SystemInfo } from '$lib/api/client';
	import { useSession, signOut } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	const session = useSession();

	let settings = $state<UserSettings | null>(null);
	let systemInfo = $state<SystemInfo | null>(null);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let saveMessage = $state('');
	let activeTab = $state<'general' | 'ai' | 'notifications' | 'account'>('general');

	// Form state
	let selectedModel = $state('');
	let theme = $state('light');
	let emailNotifications = $state(true);
	let dailySummary = $state(false);
	let shareAnalytics = $state(true);

	onMount(async () => {
		await loadSettings();
		await loadSystemInfo();
		isLoading = false;
	});

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

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100">
		<h1 class="text-xl font-semibold text-gray-900">Settings</h1>
		<p class="text-sm text-gray-500 mt-0.5">Manage your account and preferences</p>
	</div>

	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto">
			<div class="max-w-4xl mx-auto p-6">
				<!-- Tab Navigation -->
				<div class="flex gap-1 mb-6 border-b border-gray-200">
					<button
						onclick={() => (activeTab = 'general')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'general'
							? 'text-gray-900 border-b-2 border-gray-900'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						General
					</button>
					<button
						onclick={() => (activeTab = 'ai')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'ai'
							? 'text-gray-900 border-b-2 border-gray-900'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						AI Settings
					</button>
					<button
						onclick={() => (activeTab = 'notifications')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'notifications'
							? 'text-gray-900 border-b-2 border-gray-900'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						Notifications
					</button>
					<button
						onclick={() => (activeTab = 'account')}
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'account'
							? 'text-gray-900 border-b-2 border-gray-900'
							: 'text-gray-500 hover:text-gray-700'}"
					>
						Account
					</button>
				</div>

				<!-- General Tab -->
				{#if activeTab === 'general'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">Appearance</h2>
							<div class="space-y-4">
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-2">Theme</label>
									<div class="flex gap-3">
										<button
											onclick={() => (theme = 'light')}
											class="flex-1 p-4 rounded-lg border-2 transition-colors {theme === 'light'
												? 'border-gray-900 bg-gray-50'
												: 'border-gray-200 hover:border-gray-300'}"
										>
											<div class="flex items-center gap-3">
												<div class="w-10 h-10 rounded-lg bg-white border border-gray-200 flex items-center justify-center">
													<svg class="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
														<path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z" />
													</svg>
												</div>
												<div class="text-left">
													<p class="font-medium text-gray-900">Light</p>
													<p class="text-xs text-gray-500">Default light theme</p>
												</div>
											</div>
										</button>
										<button
											onclick={() => (theme = 'dark')}
											class="flex-1 p-4 rounded-lg border-2 transition-colors {theme === 'dark'
												? 'border-gray-900 bg-gray-50'
												: 'border-gray-200 hover:border-gray-300'}"
										>
											<div class="flex items-center gap-3">
												<div class="w-10 h-10 rounded-lg bg-gray-900 flex items-center justify-center">
													<svg class="w-5 h-5 text-gray-100" fill="currentColor" viewBox="0 0 24 24">
														<path fill-rule="evenodd" d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z" clip-rule="evenodd" />
													</svg>
												</div>
												<div class="text-left">
													<p class="font-medium text-gray-900">Dark</p>
													<p class="text-xs text-gray-500">Easy on the eyes</p>
												</div>
											</div>
										</button>
									</div>
								</div>
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">Privacy</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900">Share anonymous analytics</p>
									<p class="text-sm text-gray-500">Help us improve by sharing usage data</p>
								</div>
								<button
									onclick={() => (shareAnalytics = !shareAnalytics)}
									class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {shareAnalytics
										? 'bg-gray-900'
										: 'bg-gray-200'}"
								>
									<span
										class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {shareAnalytics
											? 'translate-x-6'
											: 'translate-x-1'}"
									></span>
								</button>
							</div>
						</div>
					</div>
				{/if}

				<!-- AI Settings Tab -->
				{#if activeTab === 'ai'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">AI Model</h2>
							<div class="space-y-4">
								<div>
									<label for="model" class="block text-sm font-medium text-gray-700 mb-2">
										Default Model
									</label>
									<select
										id="model"
										bind:value={selectedModel}
										class="input"
									>
										{#if systemInfo}
											{#each systemInfo.available_models as model}
												<option value={model.name}>
													{model.display_name} ({model.provider})
												</option>
											{/each}
										{:else}
											<option value="">Loading models...</option>
										{/if}
									</select>
									<p class="text-xs text-gray-500 mt-1">
										This model will be used for new conversations
									</p>
								</div>

								{#if systemInfo}
									<div class="pt-4 border-t border-gray-200">
										<h3 class="text-sm font-medium text-gray-700 mb-3">Available Models</h3>
										<div class="space-y-2">
											{#each systemInfo.available_models as model}
												<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50">
													<div>
														<p class="font-medium text-gray-900">{model.display_name}</p>
														<p class="text-xs text-gray-500">{model.description}</p>
													</div>
													<span class="text-xs px-2 py-1 rounded-full bg-gray-200 text-gray-600">
														{model.provider}
													</span>
												</div>
											{/each}
										</div>
									</div>
								{/if}
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">System Status</h2>
							<div class="flex items-center gap-3">
								<div class="w-3 h-3 rounded-full {systemInfo?.ollama_mode === 'local' ? 'bg-green-500' : 'bg-blue-500'}"></div>
								<div>
									<p class="font-medium text-gray-900">
										{systemInfo?.ollama_mode === 'local' ? 'Local Mode' : 'Cloud Mode'}
									</p>
									<p class="text-sm text-gray-500">
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
							<h2 class="text-lg font-medium text-gray-900 mb-4">Email Notifications</h2>
							<div class="space-y-4">
								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium text-gray-900">Email notifications</p>
										<p class="text-sm text-gray-500">Receive important updates via email</p>
									</div>
									<button
										onclick={() => (emailNotifications = !emailNotifications)}
										class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {emailNotifications
											? 'bg-gray-900'
											: 'bg-gray-200'}"
									>
										<span
											class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {emailNotifications
												? 'translate-x-6'
												: 'translate-x-1'}"
										></span>
									</button>
								</div>

								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium text-gray-900">Daily summary</p>
										<p class="text-sm text-gray-500">Get a daily recap of your activity</p>
									</div>
									<button
										onclick={() => (dailySummary = !dailySummary)}
										class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {dailySummary
											? 'bg-gray-900'
											: 'bg-gray-200'}"
									>
										<span
											class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {dailySummary
												? 'translate-x-6'
												: 'translate-x-1'}"
										></span>
									</button>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Account Tab -->
				{#if activeTab === 'account'}
					<div class="space-y-6">
						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">Account Information</h2>
							<div class="space-y-4">
								<div>
									<label class="block text-sm font-medium text-gray-500 mb-1">Name</label>
									<p class="text-gray-900">{$session.data?.user?.name || 'Not set'}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-500 mb-1">Email</label>
									<p class="text-gray-900">{$session.data?.user?.email || 'Not set'}</p>
								</div>
							</div>
						</div>

						<div class="card">
							<h2 class="text-lg font-medium text-gray-900 mb-4">Sessions</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900">Current session</p>
									<p class="text-sm text-gray-500">You're signed in on this device</p>
								</div>
								<button
									onclick={handleLogout}
									class="btn btn-secondary text-sm"
								>
									Sign Out
								</button>
							</div>
						</div>

						<div class="card border-red-200">
							<h2 class="text-lg font-medium text-red-600 mb-4">Danger Zone</h2>
							<div class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900">Delete account</p>
									<p class="text-sm text-gray-500">Permanently delete your account and all data</p>
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

				<!-- Save Button -->
				{#if activeTab !== 'account'}
					<div class="mt-6 flex items-center justify-between">
						<p class="text-sm text-gray-500">
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
