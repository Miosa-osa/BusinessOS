<script lang="ts">
	let accessibilityGranted = $state(false);
	let shortcuts = $state<{ quickChat: string; spotlight: string; voiceInput: string }>({
		quickChat: 'CommandOrControl+Shift+Space',
		spotlight: 'CommandOrControl+Space',
		voiceInput: 'CommandOrControl+D',
	});
	let isCheckingPermissions = $state(false);

	$effect(() => {
		loadDesktopSettings();
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
</script>

<div class="space-y-6">
	<!-- System Permissions -->
	<div class="card">
		<h2 class="text-lg font-medium st-title mb-4">System Permissions</h2>
		<p class="text-sm st-muted mb-6">
			BusinessOS requires certain system permissions for features like global shortcuts, screenshot capture, and voice input.
		</p>

		<div class="space-y-4">
			<!-- Accessibility -->
			<div class="flex items-center justify-between p-4 rounded-lg st-perm-card">
				<div class="flex items-center gap-4">
					<div class="w-10 h-10 rounded-lg {accessibilityGranted ? 'bg-green-100 dark:bg-green-900/30' : 'bg-yellow-100 dark:bg-yellow-900/30'} flex items-center justify-center">
						<svg class="w-5 h-5 {accessibilityGranted ? 'text-green-600 dark:text-green-400' : 'text-yellow-600 dark:text-yellow-400'}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
						</svg>
					</div>
					<div>
						<p class="font-medium st-title">Accessibility</p>
						<p class="text-sm st-muted">
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
						<button onclick={requestAccessibility} class="btn-pill btn-pill-ghost btn-pill-sm btn btn-primary">Enable</button>
					{/if}
					<button
						onclick={() => openSystemPreferences('accessibility')}
						class="btn btn-secondary text-sm st-btn-secondary"
					>
						Open Settings
					</button>
				</div>
			</div>

			<!-- Screen Recording -->
			<div class="flex items-center justify-between p-4 rounded-lg st-perm-card">
				<div class="flex items-center gap-4">
					<div class="w-10 h-10 rounded-lg st-icon-bg flex items-center justify-center">
						<svg class="w-5 h-5 st-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<div>
						<p class="font-medium st-title">Screen Recording</p>
						<p class="text-sm st-muted">Required for screenshot capture</p>
					</div>
				</div>
				<button
					onclick={() => openSystemPreferences('screenRecording')}
					class="btn btn-secondary text-sm st-btn-secondary"
				>
					Open Settings
				</button>
			</div>

			<!-- Microphone -->
			<div class="flex items-center justify-between p-4 rounded-lg st-perm-card">
				<div class="flex items-center gap-4">
					<div class="w-10 h-10 rounded-lg st-icon-bg flex items-center justify-center">
						<svg class="w-5 h-5 st-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
						</svg>
					</div>
					<div>
						<p class="font-medium st-title">Microphone</p>
						<p class="text-sm st-muted">Required for voice input and meeting recording</p>
					</div>
				</div>
				<button
					onclick={() => openSystemPreferences('microphone')}
					class="btn btn-secondary text-sm st-btn-secondary"
				>
					Open Settings
				</button>
			</div>
		</div>
	</div>

	<!-- Keyboard Shortcuts -->
	<div class="card">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-medium st-title">Keyboard Shortcuts</h2>
			<button
				onclick={resetShortcuts}
				class="btn-pill btn-pill-ghost btn-pill-sm"
			>
				Reset to defaults
			</button>
		</div>
		<p class="text-sm st-muted mb-6">
			Global shortcuts work system-wide, even when the app is in the background.
		</p>

		<div class="space-y-3">
			<div class="flex items-center justify-between p-3 rounded-lg st-shortcut-row">
				<div>
					<p class="font-medium st-title text-sm">Quick Chat</p>
					<p class="text-xs st-muted">Open chat popup from anywhere</p>
				</div>
				<div class="flex items-center gap-2 font-mono text-sm st-kbd px-3 py-1.5 rounded">
					{formatShortcut(shortcuts.quickChat)}
				</div>
			</div>

			<div class="flex items-center justify-between p-3 rounded-lg st-shortcut-row">
				<div>
					<p class="font-medium st-title text-sm">Voice Input</p>
					<p class="text-xs st-muted">Start voice dictation</p>
				</div>
				<div class="flex items-center gap-2 font-mono text-sm st-kbd px-3 py-1.5 rounded">
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

<style>
	.st-title { color: var(--dt, var(--bos-text-primary, #111)); }
	.st-muted { color: var(--dt3, var(--bos-text-tertiary, #888)); }
	.st-icon  { color: var(--dt3, var(--bos-text-tertiary, #888)); }
	.st-perm-card { border: 1px solid var(--dbd, var(--bos-border, #e0e0e0)); }
	.st-icon-bg { background: var(--dbg3, #eee); }
	.st-btn-secondary {
		background: var(--dbg2, var(--bos-bg-secondary, #f5f5f5));
		color: var(--dt, var(--bos-text-primary, #111));
		border-color: var(--dbd, var(--bos-border, #e0e0e0));
	}
	.st-shortcut-row { background: var(--dbg2, var(--bos-bg-secondary, #f5f5f5)); }
	.st-kbd {
		background: var(--dbg, var(--bos-card, #fff));
		border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
		color: var(--dt, var(--bos-text-primary, #111));
	}
</style>
