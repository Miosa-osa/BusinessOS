<script lang="ts">
	import { api } from '$lib/api';
	import { themeStore } from '$lib/stores/themeStore';

	interface Props {
		initialTheme?: string;
		initialShareAnalytics?: boolean;
	}

	let { initialTheme = 'light', initialShareAnalytics = true }: Props = $props();

	// Local mutable copies — initialized once from props (intentional, single-mount panel)
	let theme = $state('light');
	let shareAnalytics = $state(true);

	$effect.pre(() => {
		theme = initialTheme;
		shareAnalytics = initialShareAnalytics;
	});
	let isSaving = $state(false);
	let saveMessage = $state('');

	async function handleSave() {
		isSaving = true;
		saveMessage = '';
		try {
			const settings = await api.getSettings();
			await api.updateSettings({
				default_model: settings?.default_model ?? null,
				theme,
				email_notifications: settings?.email_notifications ?? true,
				daily_summary: settings?.daily_summary ?? false,
				share_analytics: shareAnalytics,
			});
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
</script>

<div class="space-y-6">
	<!-- Appearance -->
	<div class="card">
		<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Appearance</h2>
		<div class="space-y-4">
			<div>
				<span class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Theme</span>
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

	<!-- Privacy -->
	<div class="card">
		<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Privacy</h2>
		<div class="flex items-center justify-between">
			<div>
				<p class="font-medium text-gray-900 dark:text-white">Share anonymous analytics</p>
				<p class="text-sm text-gray-500 dark:text-gray-400">Help us improve by sharing usage data</p>
			</div>
			<button
				aria-label="Toggle anonymous analytics"
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

	<!-- Save -->
	<div class="mt-6 flex items-center justify-between">
		<p class="text-sm text-gray-500 dark:text-gray-400">
			{saveMessage || 'Changes are saved when you click Save'}
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
</div>
