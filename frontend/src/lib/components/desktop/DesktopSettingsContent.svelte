<script lang="ts">
	import { desktopSettings, desktopBackgrounds, iconStyles, iconSizePresets, backgroundFitOptions, type IconStyle, type BackgroundFit } from '$lib/stores/desktopStore';
	import { windowStore, type DesktopConfig } from '$lib/stores/windowStore';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	// Local state
	let selectedTab = $state<'icons' | 'background' | 'shortcuts' | 'permissions' | 'data'>('icons');

	// Data tab state
	let importError = $state('');
	let importSuccess = $state(false);
	let configFileInput: HTMLInputElement | undefined = $state(undefined);
	let customImageUrl = $state('');
	let companyNameInput = $state($desktopSettings.companyName);
	let fileInput: HTMLInputElement | undefined = $state(undefined);

	// Shortcut settings - will be loaded from Electron
	let shortcuts = $state({
		spotlight: '⌘+Space',
		quickChat: '⌘+Shift+Space',
		voiceInput: '⌘+D',
	});

	// Shortcut recording state
	let recordingKey = $state<string | null>(null);
	let accessibilityGranted = $state(false);
	let isElectron = $state(false);
	let isCheckingPermissions = $state(false);

	// Check if we're in Electron and load shortcuts
	onMount(async () => {
		// Check for window.electron (exposed by preload script)
		const electron = (window as any).electron;
		if (browser && electron) {
			isElectron = true;
			console.log('Electron detected, loading settings...');

			// Load current shortcuts from Electron
			try {
				const savedShortcuts = await electron.shortcuts?.get();
				if (savedShortcuts) {
					shortcuts = {
						spotlight: formatAcceleratorForDisplay(savedShortcuts.spotlight),
						quickChat: formatAcceleratorForDisplay(savedShortcuts.quickChat),
						voiceInput: formatAcceleratorForDisplay(savedShortcuts.voiceInput),
					};
				}

				// Check accessibility permissions
				const accessResult = await electron.shortcuts?.checkAccessibility();
				accessibilityGranted = accessResult?.granted ?? false;
				console.log('Accessibility granted:', accessibilityGranted);
			} catch (e) {
				console.error('Failed to load shortcuts:', e);
			}
		} else {
			console.log('Not running in Electron (window.electron not found)');
		}
	});

	// Convert Electron accelerator to display format
	function formatAcceleratorForDisplay(accelerator: string): string {
		if (!accelerator) return '';
		return accelerator
			.replace('CommandOrControl', '⌘')
			.replace('Command', '⌘')
			.replace('Control', '⌃')
			.replace('Shift', '⇧')
			.replace('Alt', '⌥')
			.replace('Option', '⌥')
			.replace(/\+/g, '');
	}

	// Convert display format back to Electron accelerator
	function formatDisplayToAccelerator(display: string): string {
		if (!display) return '';
		let result = display
			.replace('⌘', 'CommandOrControl+')
			.replace('⌃', 'Control+')
			.replace('⇧', 'Shift+')
			.replace('⌥', 'Alt+');
		// Remove trailing +
		if (result.endsWith('+')) {
			result = result.slice(0, -1);
		}
		return result;
	}

	// Start recording a shortcut
	function startRecording(key: string) {
		recordingKey = key;
		// Add keyboard listener
		if (browser) {
			window.addEventListener('keydown', handleRecordKeyDown);
		}
	}

	// Stop recording
	function stopRecording() {
		recordingKey = null;
		if (browser) {
			window.removeEventListener('keydown', handleRecordKeyDown);
		}
	}

	// Handle key press during recording
	async function handleRecordKeyDown(event: KeyboardEvent) {
		event.preventDefault();
		event.stopPropagation();

		// Ignore modifier-only keys
		if (['Control', 'Shift', 'Alt', 'Meta', 'Command'].includes(event.key)) {
			return;
		}

		// Build the accelerator string
		const parts: string[] = [];

		if (event.metaKey || event.ctrlKey) parts.push('CommandOrControl');
		if (event.shiftKey) parts.push('Shift');
		if (event.altKey) parts.push('Alt');

		// Get the key
		let key = event.key.toUpperCase();
		if (key === ' ') key = 'Space';
		if (key === 'ESCAPE') key = 'Escape';
		if (key === 'BACKSPACE') key = 'Backspace';
		if (key === 'TAB') key = 'Tab';
		if (key === 'ENTER') key = 'Enter';
		if (key === '`') key = '`';

		parts.push(key);

		const accelerator = parts.join('+');
		const displayFormat = formatAcceleratorForDisplay(accelerator);

		// Update local state
		if (recordingKey) {
			(shortcuts as any)[recordingKey] = displayFormat;

			// Save to Electron
			if (isElectron && browser) {
				try {
					const electron = window as any;
					await electron.electron?.shortcuts?.set(recordingKey, accelerator);
				} catch (e) {
					console.error('Failed to save shortcut:', e);
				}
			}
		}

		stopRecording();
	}

	// Request accessibility permissions
	async function requestAccessibility() {
		if (isElectron && browser) {
			try {
				const electron = (window as any).electron;
				await electron?.shortcuts?.requestAccessibility();
				// Check again after a delay (user needs to grant in System Preferences)
				setTimeout(async () => {
					const result = await electron?.shortcuts?.checkAccessibility();
					accessibilityGranted = result?.granted ?? false;
				}, 1000);
			} catch (e) {
				console.error('Failed to request accessibility:', e);
			}
		}
	}

	// Open System Preferences to specific pane
	async function openSystemPreferences(pane: string) {
		if (isElectron && browser) {
			try {
				const electron = (window as any).electron;
				const urls: Record<string, string> = {
					accessibility: 'x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility',
					screenRecording: 'x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture',
					microphone: 'x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone',
				};
				await electron?.shell?.openExternal(urls[pane] || 'x-apple.systempreferences:');
			} catch (e) {
				console.error('Failed to open system preferences:', e);
			}
		}
	}

	// Recheck permissions
	async function recheckPermissions() {
		if (isElectron && browser) {
			isCheckingPermissions = true;
			try {
				const electron = (window as any).electron;
				const result = await electron?.shortcuts?.checkAccessibility();
				accessibilityGranted = result?.granted ?? false;
			} catch (e) {
				console.error('Failed to check permissions:', e);
			}
			isCheckingPermissions = false;
		}
	}

	// Reset shortcuts to defaults
	async function resetShortcuts() {
		if (isElectron && browser) {
			try {
				const electron = (window as any).electron;
				const result = await electron?.shortcuts?.reset();
				if (result?.shortcuts) {
					shortcuts = {
						spotlight: formatAcceleratorForDisplay(result.shortcuts.spotlight),
						quickChat: formatAcceleratorForDisplay(result.shortcuts.quickChat),
						voiceInput: formatAcceleratorForDisplay(result.shortcuts.voiceInput),
					};
				}
			} catch (e) {
				console.error('Failed to reset shortcuts:', e);
			}
		}
	}

	// Tooltip state
	let tooltipText = $state('');
	let tooltipVisible = $state(false);
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	function showTooltip(event: MouseEvent, text: string) {
		tooltipText = text;
		tooltipX = event.clientX;
		tooltipY = event.clientY - 30;
		tooltipVisible = true;
	}

	function hideTooltip() {
		tooltipVisible = false;
	}

	function moveTooltip(event: MouseEvent) {
		tooltipX = event.clientX;
		tooltipY = event.clientY - 30;
	}

	function handleIconSizeChange(event: Event) {
		const target = event.target as HTMLInputElement;
		desktopSettings.setIconSize(parseInt(target.value, 10));
	}

	function handleIconStyleChange(style: IconStyle) {
		desktopSettings.setIconStyle(style);
	}

	function handleBackgroundChange(backgroundId: string) {
		desktopSettings.setBackground(backgroundId);
	}

	function applyCustomImage() {
		if (customImageUrl.trim()) {
			desktopSettings.setCustomBackground(customImageUrl.trim());
		}
	}

	function handleFileUpload(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (file && file.type.startsWith('image/')) {
			const reader = new FileReader();
			reader.onload = (e) => {
				const dataUrl = e.target?.result as string;
				desktopSettings.setCustomBackground(dataUrl);
			};
			reader.readAsDataURL(file);
		}
	}

	function triggerFileUpload() {
		fileInput?.click();
	}

	// Get current preset label
	function getSizeLabel(size: number): string {
		const preset = iconSizePresets.find(p => p.value === size);
		if (preset) return preset.label;
		return `${size}px`;
	}

	// Carousel scroll refs
	let colorScrollContainer: HTMLDivElement | undefined = $state(undefined);
	let gradientScrollContainer: HTMLDivElement | undefined = $state(undefined);
	let patternScrollContainer: HTMLDivElement | undefined = $state(undefined);

	function scrollCarousel(container: HTMLDivElement | undefined, direction: 'left' | 'right') {
		if (!container) return;
		const scrollAmount = 200;
		container.scrollBy({
			left: direction === 'right' ? scrollAmount : -scrollAmount,
			behavior: 'smooth'
		});
	}

	// Data export/import functions
	function exportConfig() {
		const config = windowStore.exportConfig();
		const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `businessos-desktop-config-${new Date().toISOString().split('T')[0]}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	function exportSchema() {
		const schema = windowStore.getConfigSchema();
		const blob = new Blob([JSON.stringify(schema, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'businessos-desktop-config-schema.json';
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	function triggerImport() {
		configFileInput?.click();
	}

	function handleConfigImport(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		importError = '';
		importSuccess = false;

		const reader = new FileReader();
		reader.onload = (e) => {
			try {
				const content = e.target?.result as string;
				const config = JSON.parse(content) as DesktopConfig;
				const result = windowStore.importConfig(config);

				if (result.success) {
					importSuccess = true;
					setTimeout(() => (importSuccess = false), 3000);
				} else {
					importError = result.error || 'Import failed';
					setTimeout(() => (importError = ''), 5000);
				}
			} catch (err) {
				importError = 'Invalid JSON file';
				setTimeout(() => (importError = ''), 5000);
			}
		};
		reader.readAsText(file);

		// Reset file input
		target.value = '';
	}
</script>

<div class="desktop-settings">
	<!-- Tabs -->
	<div class="tabs">
		<button
			class="tab"
			class:active={selectedTab === 'icons'}
			onclick={() => selectedTab = 'icons'}
		>
			Icons
		</button>
		<button
			class="tab"
			class:active={selectedTab === 'background'}
			onclick={() => selectedTab = 'background'}
		>
			Background
		</button>
		<button
			class="tab"
			class:active={selectedTab === 'shortcuts'}
			onclick={() => selectedTab = 'shortcuts'}
		>
			Shortcuts
		</button>
		{#if isElectron}
			<button
				class="tab"
				class:active={selectedTab === 'permissions'}
				onclick={() => selectedTab = 'permissions'}
			>
				Permissions
			</button>
		{/if}
		<button
			class="tab"
			class:active={selectedTab === 'data'}
			onclick={() => selectedTab = 'data'}
		>
			Data
		</button>
	</div>

	<!-- Content -->
	<div class="content">
		{#if selectedTab === 'icons'}
			<!-- Company Branding -->
			<div class="section">
				<label class="section-title">Company Branding</label>
				<div class="branding-row">
					<div class="branding-preview">
						<span class="preview-name">{companyNameInput || 'BUSINESS'}</span>
						<span class="preview-os">OS</span>
					</div>
					<div class="branding-input-row">
						<input
							type="text"
							placeholder="Company Name"
							class="company-name-input"
							bind:value={companyNameInput}
							onkeydown={(e) => e.key === 'Enter' && desktopSettings.setCompanyName(companyNameInput)}
							maxlength="16"
						/>
						<button
							class="save-name-btn"
							onclick={() => desktopSettings.setCompanyName(companyNameInput)}
						>
							Save
						</button>
					</div>
					<p class="branding-hint">This name appears on the loading screen</p>
				</div>
			</div>

			<!-- Icon Size -->
			<div class="section">
				<div class="section-header">
					<label class="section-title">Icon Size</label>
					<span class="section-value">{getSizeLabel($desktopSettings.iconSize)}</span>
				</div>
				<div class="slider-row">
					<span class="slider-label">Small</span>
					<input
						type="range"
						min="32"
						max="128"
						step="8"
						value={$desktopSettings.iconSize}
						oninput={handleIconSizeChange}
						class="slider"
					/>
					<span class="slider-label">Large</span>
				</div>
				<!-- Size Preview -->
				<div class="size-preview">
					{#each [48, 64, 96] as previewSize}
						<div
							class="preview-icon"
							class:active={$desktopSettings.iconSize === previewSize}
						>
							<div
								class="preview-box"
								style="width: {previewSize * 0.6}px; height: {previewSize * 0.6}px;"
							>
								<svg
									style="width: {previewSize * 0.35}px; height: {previewSize * 0.35}px;"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
								>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<span class="preview-label">{previewSize}px</span>
						</div>
					{/each}
				</div>
			</div>

			<!-- Icon Style -->
			<div class="section">
				<label class="section-title">Icon Style</label>
				<div class="style-grid">
					{#each iconStyles as style}
						<button
							class="style-option"
							class:selected={$desktopSettings.iconStyle === style.id}
							onclick={() => handleIconStyleChange(style.id)}
						>
							<div class="style-name">{style.name}</div>
							<div class="style-desc">{style.description}</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Toggles -->
			<div class="section">
				<label class="section-title">Options</label>
				<div class="toggles">
					<div class="toggle-row">
						<div class="toggle-info">
							<div class="toggle-label">Show Icon Labels</div>
							<div class="toggle-desc">Display text labels under icons</div>
						</div>
						<button
							onclick={() => desktopSettings.toggleIconLabels()}
							class="toggle-switch"
							class:active={$desktopSettings.showIconLabels}
							role="switch"
							aria-checked={$desktopSettings.showIconLabels}
						>
							<span class="toggle-thumb"></span>
						</button>
					</div>

					<div class="toggle-row">
						<div class="toggle-info">
							<div class="toggle-label">Snap to Grid</div>
							<div class="toggle-desc">Align icons to grid when dragging</div>
						</div>
						<button
							onclick={() => desktopSettings.toggleGridSnap()}
							class="toggle-switch"
							class:active={$desktopSettings.gridSnap}
							role="switch"
							aria-checked={$desktopSettings.gridSnap}
						>
							<span class="toggle-thumb"></span>
						</button>
					</div>

					<div class="toggle-row">
						<div class="toggle-info">
							<div class="toggle-label">Noise Texture</div>
							<div class="toggle-desc">Add subtle noise overlay to background</div>
						</div>
						<button
							onclick={() => desktopSettings.toggleNoise()}
							class="toggle-switch"
							class:active={$desktopSettings.showNoise}
							role="switch"
							aria-checked={$desktopSettings.showNoise}
						>
							<span class="toggle-thumb"></span>
						</button>
					</div>
				</div>
			</div>
		{:else if selectedTab === 'background'}
			<!-- Background Selection -->
			<div class="section">
				<label class="section-title">Solid Colors</label>
				<div class="carousel-container">
					<button class="carousel-btn left" onclick={() => scrollCarousel(colorScrollContainer, 'left')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M15 18l-6-6 6-6"/>
						</svg>
					</button>
					<div class="carousel-scroll" bind:this={colorScrollContainer}>
						<div class="color-grid">
							{#each desktopBackgrounds.filter(b => b.type === 'solid') as bg}
								<button
									class="color-swatch"
									class:selected={$desktopSettings.backgroundId === bg.id}
									style="background: {bg.preview};"
									onclick={() => handleBackgroundChange(bg.id)}
									onmouseenter={(e) => showTooltip(e, bg.name)}
									onmousemove={moveTooltip}
									onmouseleave={hideTooltip}
								></button>
							{/each}
						</div>
					</div>
					<button class="carousel-btn right" onclick={() => scrollCarousel(colorScrollContainer, 'right')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 18l6-6-6-6"/>
						</svg>
					</button>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Gradients</label>
				<div class="carousel-container">
					<button class="carousel-btn left" onclick={() => scrollCarousel(gradientScrollContainer, 'left')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M15 18l-6-6 6-6"/>
						</svg>
					</button>
					<div class="carousel-scroll" bind:this={gradientScrollContainer}>
						<div class="gradient-grid">
							{#each desktopBackgrounds.filter(b => b.type === 'gradient') as bg}
								<button
									class="gradient-swatch"
									class:selected={$desktopSettings.backgroundId === bg.id}
									style="background: {bg.preview};"
									onclick={() => handleBackgroundChange(bg.id)}
									onmouseenter={(e) => showTooltip(e, bg.name)}
									onmousemove={moveTooltip}
									onmouseleave={hideTooltip}
								></button>
							{/each}
						</div>
					</div>
					<button class="carousel-btn right" onclick={() => scrollCarousel(gradientScrollContainer, 'right')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 18l6-6-6-6"/>
						</svg>
					</button>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Patterns</label>
				<div class="carousel-container">
					<button class="carousel-btn left" onclick={() => scrollCarousel(patternScrollContainer, 'left')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M15 18l-6-6 6-6"/>
						</svg>
					</button>
					<div class="carousel-scroll" bind:this={patternScrollContainer}>
						<div class="pattern-grid">
							{#each desktopBackgrounds.filter(b => b.type === 'pattern') as bg}
								<button
									class="pattern-swatch"
									class:selected={$desktopSettings.backgroundId === bg.id}
									style="background: {bg.preview}; background-size: 10px 10px;"
									onclick={() => handleBackgroundChange(bg.id)}
									onmouseenter={(e) => showTooltip(e, bg.name)}
									onmousemove={moveTooltip}
									onmouseleave={hideTooltip}
								></button>
							{/each}
						</div>
					</div>
					<button class="carousel-btn right" onclick={() => scrollCarousel(patternScrollContainer, 'right')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 18l6-6-6-6"/>
						</svg>
					</button>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Custom Image</label>
				<input
					type="file"
					accept="image/*"
					bind:this={fileInput}
					onchange={handleFileUpload}
					class="hidden-file-input"
				/>

				{#if $desktopSettings.backgroundId === 'custom' && $desktopSettings.customBackgroundUrl}
					<!-- Show current custom background with preview -->
					<div class="custom-preview-container">
						<div
							class="custom-preview-image"
							style="background-image: url({$desktopSettings.customBackgroundUrl});"
						></div>
						<div class="custom-preview-info">
							<span class="preview-label">
								{#if $desktopSettings.customBackgroundUrl.startsWith('data:')}
									Uploaded Image
								{:else}
									Custom URL
								{/if}
							</span>
							<span class="preview-status">Currently active</span>
						</div>
						<div class="custom-preview-actions">
							<button class="change-btn" onclick={triggerFileUpload}>
								Change
							</button>
							<button class="remove-btn" onclick={() => desktopSettings.setBackground('classic-gray')}>
								Remove
							</button>
						</div>
					</div>

					<!-- Fit options -->
					<div class="fit-options">
						<span class="fit-label">Image Fit:</span>
						<div class="fit-buttons">
							{#each backgroundFitOptions as fit}
								<button
									class="fit-btn"
									class:active={$desktopSettings.backgroundFit === fit.id}
									onclick={() => desktopSettings.setBackgroundFit(fit.id)}
									title={fit.description}
								>
									{fit.name}
								</button>
							{/each}
						</div>
					</div>
				{:else}
					<!-- Show upload options when no custom background -->
					<div class="custom-image-options">
						<button class="upload-btn" onclick={triggerFileUpload}>
							<svg class="upload-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
							</svg>
							Upload Image
						</button>
						<span class="or-divider">or</span>
						<div class="url-input-row">
							<input
								type="text"
								placeholder="Paste image URL..."
								class="custom-url-input"
								bind:value={customImageUrl}
								onkeydown={(e) => e.key === 'Enter' && applyCustomImage()}
							/>
							<button class="apply-btn" onclick={applyCustomImage}>
								Apply
							</button>
						</div>
					</div>
				{/if}
			</div>
		{:else if selectedTab === 'shortcuts'}
			<!-- Accessibility Permission Banner -->
			{#if isElectron && !accessibilityGranted}
				<div class="accessibility-banner">
					<div class="banner-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
						</svg>
					</div>
					<div class="banner-content">
						<div class="banner-title">Accessibility Permission Required</div>
						<div class="banner-desc">Global shortcuts need accessibility access to work from anywhere on your Mac.</div>
					</div>
					<button class="banner-btn" onclick={requestAccessibility}>
						Grant Access
					</button>
				</div>
			{/if}

			<!-- Keyboard Shortcuts -->
			<div class="section">
				<label class="section-title">Global Shortcuts</label>
				<p class="section-subtitle">Click on a shortcut to change it. Press your desired key combination.</p>
				<div class="shortcuts-list">
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Spotlight Search</div>
							<div class="shortcut-desc">Quick search and app launcher</div>
						</div>
						<button
							class="shortcut-key-btn"
							class:recording={recordingKey === 'spotlight'}
							onclick={() => recordingKey === 'spotlight' ? stopRecording() : startRecording('spotlight')}
						>
							{#if recordingKey === 'spotlight'}
								<span class="recording-text">Press keys...</span>
							{:else}
								{shortcuts.spotlight}
							{/if}
						</button>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Quick Chat Popup</div>
							<div class="shortcut-desc">Open AI chat from anywhere</div>
						</div>
						<button
							class="shortcut-key-btn"
							class:recording={recordingKey === 'quickChat'}
							onclick={() => recordingKey === 'quickChat' ? stopRecording() : startRecording('quickChat')}
						>
							{#if recordingKey === 'quickChat'}
								<span class="recording-text">Press keys...</span>
							{:else}
								{shortcuts.quickChat}
							{/if}
						</button>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Voice Input</div>
							<div class="shortcut-desc">Start/stop voice recording</div>
						</div>
						<button
							class="shortcut-key-btn"
							class:recording={recordingKey === 'voiceInput'}
							onclick={() => recordingKey === 'voiceInput' ? stopRecording() : startRecording('voiceInput')}
						>
							{#if recordingKey === 'voiceInput'}
								<span class="recording-text">Press keys...</span>
							{:else}
								{shortcuts.voiceInput}
							{/if}
						</button>
					</div>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Window Management</label>
				<p class="section-subtitle">System shortcuts (not customizable)</p>
				<div class="shortcuts-list">
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Close Window</div>
							<div class="shortcut-desc">Close the active window</div>
						</div>
						<div class="shortcut-key">⌘W</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Minimize Window</div>
							<div class="shortcut-desc">Minimize to dock</div>
						</div>
						<div class="shortcut-key">⌘M</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Maximize/Restore</div>
							<div class="shortcut-desc">Toggle window fullscreen</div>
						</div>
						<div class="shortcut-key">⌘⇧F</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Cycle Windows</div>
							<div class="shortcut-desc">Switch between open windows</div>
						</div>
						<div class="shortcut-key">⌘`</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Snap Left</div>
							<div class="shortcut-desc">Snap window to left half</div>
						</div>
						<div class="shortcut-key">⌃⌥←</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Snap Right</div>
							<div class="shortcut-desc">Snap window to right half</div>
						</div>
						<div class="shortcut-key">⌃⌥→</div>
					</div>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Quick Actions</label>
				<p class="section-subtitle">Fast access to common tasks</p>
				<div class="shortcuts-list">
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">New Task</div>
							<div class="shortcut-desc">Create a new task quickly</div>
						</div>
						<div class="shortcut-key">⌘⇧T</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">New Project</div>
							<div class="shortcut-desc">Start a new project</div>
						</div>
						<div class="shortcut-key">⌘⇧P</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">New Note</div>
							<div class="shortcut-desc">Create a quick note</div>
						</div>
						<div class="shortcut-key">⌘⇧N</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Toggle Terminal</div>
							<div class="shortcut-desc">Open/close terminal</div>
						</div>
						<div class="shortcut-key">⌘⇧`</div>
					</div>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Navigation</label>
				<p class="section-subtitle">Move around the workspace</p>
				<div class="shortcuts-list">
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Go to Dashboard</div>
							<div class="shortcut-desc">Open dashboard view</div>
						</div>
						<div class="shortcut-key">⌘1</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Go to Chat</div>
							<div class="shortcut-desc">Open AI chat</div>
						</div>
						<div class="shortcut-key">⌘2</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Go to Tasks</div>
							<div class="shortcut-desc">Open tasks view</div>
						</div>
						<div class="shortcut-key">⌘3</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Go to Calendar</div>
							<div class="shortcut-desc">Open calendar view</div>
						</div>
						<div class="shortcut-key">⌘4</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-info">
							<div class="shortcut-name">Go to Projects</div>
							<div class="shortcut-desc">Open projects view</div>
						</div>
						<div class="shortcut-key">⌘5</div>
					</div>
				</div>
			</div>

			{#if isElectron}
				<div class="section">
					<button class="reset-shortcuts-btn" onclick={resetShortcuts}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
							<path d="M3 3v5h5"/>
						</svg>
						Reset to Default Shortcuts
					</button>
				</div>
			{/if}

			<div class="section">
				<div class="shortcut-note">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M12 16v-4M12 8h.01"/>
					</svg>
					<span>
						{#if isElectron}
							Global shortcuts work system-wide when BusinessOS Desktop is running. Some shortcuts may conflict with macOS defaults (like ⌘+Space for Spotlight).
						{:else}
							Global shortcuts are only available in the desktop app. Download BusinessOS Desktop to use shortcuts anywhere on your Mac.
						{/if}
					</span>
				</div>
			</div>
		{:else if selectedTab === 'permissions'}
			<!-- System Permissions -->
			<div class="section">
				<label class="section-title">System Permissions</label>
				<p class="section-subtitle">BusinessOS needs these permissions for global shortcuts, screenshots, and voice input.</p>

				<div class="permissions-list">
					<!-- Accessibility -->
					<div class="permission-row">
						<div class="permission-icon" class:granted={accessibilityGranted}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"/>
							</svg>
						</div>
						<div class="permission-info">
							<div class="permission-name">Accessibility</div>
							<div class="permission-desc">Required for global keyboard shortcuts to work system-wide</div>
						</div>
						<div class="permission-status">
							{#if accessibilityGranted}
								<span class="status-badge granted">
									<svg viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
									</svg>
									Granted
								</span>
							{:else}
								<button class="grant-btn" onclick={requestAccessibility}>
									Grant Access
								</button>
							{/if}
							<button class="settings-btn" onclick={() => openSystemPreferences('accessibility')} title="Open System Settings">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							</button>
						</div>
					</div>

					<!-- Screen Recording -->
					<div class="permission-row">
						<div class="permission-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
							</svg>
						</div>
						<div class="permission-info">
							<div class="permission-name">Screen Recording</div>
							<div class="permission-desc">Required for capturing screenshots from the popup chat</div>
						</div>
						<div class="permission-status">
							<button class="settings-btn" onclick={() => openSystemPreferences('screenRecording')} title="Open System Settings">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							</button>
						</div>
					</div>

					<!-- Microphone -->
					<div class="permission-row">
						<div class="permission-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
							</svg>
						</div>
						<div class="permission-info">
							<div class="permission-name">Microphone</div>
							<div class="permission-desc">Required for voice input and meeting recording features</div>
						</div>
						<div class="permission-status">
							<button class="settings-btn" onclick={() => openSystemPreferences('microphone')} title="Open System Settings">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							</button>
						</div>
					</div>
				</div>

				<button class="recheck-btn" onclick={recheckPermissions} disabled={isCheckingPermissions}>
					{#if isCheckingPermissions}
						<svg class="spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
						</svg>
						Checking...
					{:else}
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
						</svg>
						Recheck Permissions
					{/if}
				</button>
			</div>

			<div class="section">
				<div class="shortcut-note">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M12 16v-4M12 8h.01"/>
					</svg>
					<span>
						After granting permissions in System Settings, click "Recheck Permissions" to update the status. You may need to restart BusinessOS for some permissions to take effect.
					</span>
				</div>
			</div>
		{:else if selectedTab === 'data'}
			<!-- Data Export/Import -->
			<input
				type="file"
				accept=".json"
				bind:this={configFileInput}
				onchange={handleConfigImport}
				class="hidden-file-input"
			/>

			{#if importSuccess}
				<div class="status-banner success">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
					Configuration imported successfully!
				</div>
			{/if}

			{#if importError}
				<div class="status-banner error">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M15 9l-6 6M9 9l6 6"/>
					</svg>
					{importError}
				</div>
			{/if}

			<div class="section">
				<label class="section-title">Export Configuration</label>
				<p class="section-desc">
					Download your desktop layout, dock items, and folder structure as a JSON file. Use this to backup or transfer your setup.
				</p>
				<div class="data-actions">
					<button class="data-btn primary" onclick={exportConfig}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
							<path d="M7 10l5 5 5-5"/>
							<path d="M12 15V3"/>
						</svg>
						Export Desktop Config
					</button>
					<button class="data-btn secondary" onclick={exportSchema}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
							<path d="M14 2v6h6M16 13H8M16 17H8M10 9H8"/>
						</svg>
						Download Schema
					</button>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Import Configuration</label>
				<p class="section-desc">
					Load a previously exported configuration file. This will replace your current desktop layout.
				</p>
				<div class="import-area" onclick={triggerImport} onkeydown={(e) => e.key === 'Enter' && triggerImport()} tabindex="0" role="button">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
						<path d="M17 8l-5-5-5 5"/>
						<path d="M12 3v12"/>
					</svg>
					<span class="import-text">Click to import configuration file</span>
					<span class="import-hint">Accepts .json files only</span>
				</div>
			</div>

			<div class="section">
				<label class="section-title">Configuration Schema</label>
				<p class="section-desc">
					The desktop configuration follows a JSON schema for validation. You can use the schema for programmatic config generation.
				</p>
				<div class="schema-preview">
					<pre><code>{JSON.stringify({
	version: "1.0.0",
	desktopIcons: [{ id: "...", module: "...", label: "...", x: 0, y: 0 }],
	dockPinnedItems: ["finder", "dashboard", "..."],
	folders: [{ id: "...", name: "...", color: "#...", iconIds: [] }]
}, null, 2)}</code></pre>
				</div>
			</div>

			<div class="section">
				<div class="shortcut-note">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M12 16v-4M12 8h.01"/>
					</svg>
					<span>
						Configuration files store icon positions, dock items, and folders. Window states and open apps are not included.
					</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<div class="footer">
		<button class="reset-btn" onclick={() => desktopSettings.reset()}>
			Reset to Defaults
		</button>
	</div>

	<!-- Custom Tooltip -->
	{#if tooltipVisible}
		<div
			class="custom-tooltip"
			style="left: {tooltipX}px; top: {tooltipY}px;"
		>
			{tooltipText}
		</div>
	{/if}
</div>

<style>
	.desktop-settings {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #fafafa;
	}

	.tabs {
		display: flex;
		border-bottom: 1px solid #e5e5e5;
		background: white;
		flex-shrink: 0;
	}

	.tab {
		flex: 1;
		padding: 12px 16px;
		font-size: 13px;
		font-weight: 500;
		color: #666;
		background: none;
		border: none;
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition: all 0.15s ease;
	}

	.tab:hover {
		color: #333;
	}

	.tab.active {
		color: #111;
		border-bottom-color: #111;
	}

	.content {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}

	.section {
		margin-bottom: 24px;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.section-title {
		font-size: 13px;
		font-weight: 600;
		color: #333;
		display: block;
		margin-bottom: 12px;
	}

	.section-header .section-title {
		margin-bottom: 0;
	}

	.section-value {
		font-size: 12px;
		color: #666;
	}

	.slider-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.slider-label {
		font-size: 11px;
		color: #999;
	}

	.slider {
		flex: 1;
		height: 4px;
		background: #e5e5e5;
		border-radius: 2px;
		appearance: none;
		cursor: pointer;
	}

	.slider::-webkit-slider-thumb {
		appearance: none;
		width: 16px;
		height: 16px;
		background: #333;
		border-radius: 50%;
		cursor: pointer;
	}

	.size-preview {
		display: flex;
		align-items: flex-end;
		justify-content: center;
		gap: 24px;
		padding: 20px;
		background: white;
		border-radius: 8px;
		margin-top: 16px;
	}

	.preview-icon {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		opacity: 0.4;
		transition: all 0.2s ease;
	}

	.preview-icon.active {
		opacity: 1;
		transform: scale(1.1);
	}

	.preview-box {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	}

	.preview-label {
		font-size: 10px;
		color: #666;
	}

	.style-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 6px;
	}

	.style-option {
		padding: 8px;
		border-radius: 6px;
		border: 2px solid #e5e5e5;
		background: white;
		cursor: pointer;
		text-align: left;
		transition: all 0.15s ease;
	}

	.style-option:hover {
		border-color: #ccc;
	}

	.style-option.selected {
		border-color: #333;
		background: #f5f5f5;
	}

	.style-name {
		font-size: 12px;
		font-weight: 600;
		color: #333;
	}

	.style-desc {
		font-size: 10px;
		color: #999;
		margin-top: 2px;
	}

	.toggles {
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: white;
		border-radius: 8px;
		padding: 8px;
	}

	.toggle-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
	}

	.toggle-info {
		flex: 1;
	}

	.toggle-label {
		font-size: 13px;
		font-weight: 500;
		color: #333;
	}

	.toggle-desc {
		font-size: 11px;
		color: #999;
		margin-top: 2px;
	}

	.toggle-switch {
		position: relative;
		width: 44px;
		height: 24px;
		background: #ddd;
		border-radius: 12px;
		border: none;
		cursor: pointer;
		transition: background 0.2s ease;
	}

	.toggle-switch.active {
		background: #333;
	}

	.toggle-thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		background: white;
		border-radius: 50%;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
		transition: transform 0.2s ease;
	}

	.toggle-switch.active .toggle-thumb {
		transform: translateX(20px);
	}

	/* Carousel styles */
	.carousel-container {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.carousel-btn {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		border: 1px solid #ddd;
		background: white;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: all 0.15s ease;
	}

	.carousel-btn:hover {
		background: #f5f5f5;
		border-color: #ccc;
	}

	.carousel-btn svg {
		width: 14px;
		height: 14px;
		color: #666;
	}

	.carousel-scroll {
		flex: 1;
		overflow-x: auto;
		overflow-y: hidden;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.carousel-scroll::-webkit-scrollbar {
		display: none;
	}

	.color-grid {
		display: flex;
		gap: 8px;
		padding: 4px 0;
	}

	.color-swatch {
		width: 36px;
		height: 36px;
		flex-shrink: 0;
		border-radius: 8px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.color-swatch:hover {
		transform: scale(1.1);
	}

	.color-swatch.selected {
		border-color: #333;
		box-shadow: 0 0 0 2px white, 0 0 0 4px #333;
	}

	.gradient-grid {
		display: flex;
		gap: 8px;
		padding: 4px 0;
	}

	.gradient-swatch {
		width: 72px;
		height: 48px;
		flex-shrink: 0;
		border-radius: 8px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.gradient-swatch:hover {
		transform: scale(1.02);
	}

	.gradient-swatch.selected {
		border-color: #333;
		box-shadow: 0 0 0 2px white, 0 0 0 4px #333;
	}

	.pattern-grid {
		display: flex;
		gap: 8px;
		padding: 4px 0;
	}

	.pattern-swatch {
		width: 64px;
		height: 48px;
		flex-shrink: 0;
		border-radius: 8px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.pattern-swatch:hover {
		transform: scale(1.02);
	}

	.pattern-swatch.selected {
		border-color: #333;
		box-shadow: 0 0 0 2px white, 0 0 0 4px #333;
	}

	.hidden-file-input {
		display: none;
	}

	.custom-image-options {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.upload-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 14px 20px;
		background: #f5f5f5;
		border: 2px dashed #ccc;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		color: #666;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.upload-btn:hover {
		background: #eee;
		border-color: #999;
		color: #333;
	}

	.upload-icon {
		width: 18px;
		height: 18px;
	}

	.or-divider {
		text-align: center;
		font-size: 11px;
		color: #999;
		text-transform: uppercase;
	}

	.url-input-row {
		display: flex;
		gap: 8px;
	}

	.custom-url-input {
		flex: 1;
		padding: 10px 12px;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-size: 13px;
		outline: none;
		transition: border-color 0.15s ease;
	}

	.custom-url-input:focus {
		border-color: #333;
	}

	.custom-url-input::placeholder {
		color: #999;
	}

	.apply-btn {
		padding: 10px 16px;
		background: #333;
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.apply-btn:hover {
		background: #555;
	}


	.custom-preview-container {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 8px;
	}

	.custom-preview-image {
		width: 64px;
		height: 64px;
		border-radius: 6px;
		background-size: cover;
		background-position: center;
		border: 1px solid #ddd;
		flex-shrink: 0;
	}

	.custom-preview-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.preview-label {
		font-size: 13px;
		font-weight: 600;
		color: #333;
	}

	.preview-status {
		font-size: 11px;
		color: #28a745;
	}

	.custom-preview-actions {
		display: flex;
		gap: 8px;
	}

	.change-btn {
		padding: 8px 14px;
		background: #f5f5f5;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-size: 12px;
		font-weight: 500;
		color: #333;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.change-btn:hover {
		background: #eee;
		border-color: #ccc;
	}

	.remove-btn {
		padding: 8px 14px;
		background: #fff;
		border: 1px solid #dc3545;
		border-radius: 6px;
		font-size: 12px;
		font-weight: 500;
		color: #dc3545;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.remove-btn:hover {
		background: #dc3545;
		color: white;
	}

	.fit-options {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.fit-label {
		font-size: 12px;
		font-weight: 500;
		color: #666;
	}

	.fit-buttons {
		display: flex;
		gap: 4px;
	}

	.fit-btn {
		padding: 6px 12px;
		background: #f5f5f5;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 11px;
		font-weight: 500;
		color: #666;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.fit-btn:hover {
		background: #eee;
		color: #333;
	}

	.fit-btn.active {
		background: #333;
		border-color: #333;
		color: white;
	}

	.footer {
		padding: 16px 20px;
		border-top: 1px solid #e5e5e5;
		background: white;
		flex-shrink: 0;
	}

	.reset-btn {
		font-size: 12px;
		color: #666;
		background: none;
		border: none;
		cursor: pointer;
		padding: 8px 12px;
		border-radius: 6px;
		transition: all 0.15s ease;
	}

	.reset-btn:hover {
		background: #f0f0f0;
		color: #333;
	}

	.custom-tooltip {
		position: fixed;
		background: rgba(0, 0, 0, 0.85);
		color: white;
		padding: 6px 10px;
		border-radius: 4px;
		font-size: 12px;
		font-weight: 500;
		pointer-events: none;
		z-index: 9999;
		transform: translateX(-50%);
		white-space: nowrap;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
	}

	/* Shortcuts styles */
	.shortcuts-list {
		display: flex;
		flex-direction: column;
		background: white;
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid #e5e5e5;
	}

	.shortcut-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid #f0f0f0;
	}

	.shortcut-row:last-child {
		border-bottom: none;
	}

	.shortcut-info {
		flex: 1;
	}

	.shortcut-name {
		font-size: 13px;
		font-weight: 500;
		color: #333;
	}

	.shortcut-desc {
		font-size: 11px;
		color: #999;
		margin-top: 2px;
	}

	.shortcut-key {
		font-family: ui-monospace, 'SF Mono', SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 12px;
		font-weight: 500;
		color: #555;
		background: #f5f5f5;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		padding: 6px 10px;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}

	.shortcut-note {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 12px 14px;
		background: #f8f9fa;
		border: 1px solid #e9ecef;
		border-radius: 8px;
		font-size: 12px;
		color: #666;
		line-height: 1.5;
	}

	.shortcut-note svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
		color: #6c757d;
		margin-top: 1px;
	}

	/* Accessibility banner */
	.accessibility-banner {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
		border: 1px solid #f59e0b;
		border-radius: 10px;
		margin-bottom: 16px;
	}

	.banner-icon {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f59e0b;
		border-radius: 8px;
		flex-shrink: 0;
	}

	.banner-icon svg {
		width: 18px;
		height: 18px;
		color: white;
	}

	.banner-content {
		flex: 1;
	}

	.banner-title {
		font-size: 13px;
		font-weight: 600;
		color: #92400e;
	}

	.banner-desc {
		font-size: 11px;
		color: #b45309;
		margin-top: 2px;
	}

	.banner-btn {
		padding: 8px 16px;
		background: #f59e0b;
		border: none;
		border-radius: 6px;
		color: white;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s;
		white-space: nowrap;
	}

	.banner-btn:hover {
		background: #d97706;
	}

	/* Section subtitle */
	.section-subtitle {
		font-size: 11px;
		color: #999;
		margin: -4px 0 8px 0;
	}

	/* Shortcut key button (clickable) */
	.shortcut-key-btn {
		font-family: ui-monospace, 'SF Mono', SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 12px;
		font-weight: 500;
		color: #555;
		background: #f5f5f5;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		padding: 6px 12px;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
		cursor: pointer;
		transition: all 0.15s;
		min-width: 80px;
		text-align: center;
	}

	.shortcut-key-btn:hover {
		background: #eee;
		border-color: #ccc;
	}

	.shortcut-key-btn.recording {
		background: #3b82f6;
		border-color: #2563eb;
		color: white;
		animation: pulse-recording 1.5s infinite;
	}

	@keyframes pulse-recording {
		0%, 100% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.4); }
		50% { box-shadow: 0 0 0 8px rgba(59, 130, 246, 0); }
	}

	.recording-text {
		font-family: inherit;
		font-style: italic;
		font-size: 11px;
	}

	/* Reset shortcuts button */
	.reset-shortcuts-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
		padding: 10px 16px;
		background: #f5f5f5;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		color: #666;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.reset-shortcuts-btn:hover {
		background: #eee;
		color: #333;
	}

	.reset-shortcuts-btn svg {
		width: 16px;
		height: 16px;
	}

	/* Branding section */
	.branding-row {
		background: white;
		border-radius: 8px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.branding-preview {
		display: flex;
		align-items: baseline;
		justify-content: center;
		gap: 2px;
		padding: 16px;
		background: #fafafa;
		border-radius: 8px;
		font-family: 'SF Mono', 'Monaco', 'Fira Code', monospace;
	}

	.preview-name {
		font-size: 24px;
		font-weight: 800;
		color: #111;
		letter-spacing: 3px;
		text-transform: uppercase;
	}

	.preview-os {
		font-size: 20px;
		font-weight: 400;
		color: #111;
		opacity: 0.4;
	}

	.branding-input-row {
		display: flex;
		gap: 8px;
	}

	.company-name-input {
		flex: 1;
		padding: 10px 12px;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 1px;
		outline: none;
		transition: border-color 0.15s ease;
	}

	.company-name-input:focus {
		border-color: #333;
	}

	.company-name-input::placeholder {
		color: #999;
		text-transform: none;
		letter-spacing: 0;
	}

	.save-name-btn {
		padding: 10px 18px;
		background: #333;
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.save-name-btn:hover {
		background: #555;
	}

	.branding-hint {
		font-size: 11px;
		color: #999;
		margin: 0;
		text-align: center;
	}

	/* Data tab styles */
	.section-desc {
		font-size: 12px;
		color: #666;
		margin: -8px 0 12px 0;
		line-height: 1.5;
	}

	.data-actions {
		display: flex;
		gap: 12px;
	}

	.data-btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 14px 20px;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.data-btn svg {
		width: 18px;
		height: 18px;
	}

	.data-btn.primary {
		background: #111;
		color: white;
		border: none;
	}

	.data-btn.primary:hover {
		background: #333;
	}

	.data-btn.secondary {
		background: white;
		color: #333;
		border: 1px solid #ddd;
	}

	.data-btn.secondary:hover {
		background: #f5f5f5;
		border-color: #ccc;
	}

	.import-area {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 32px 20px;
		background: #f8f9fa;
		border: 2px dashed #ddd;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.import-area:hover {
		background: #f0f0f0;
		border-color: #bbb;
	}

	.import-area:focus {
		outline: none;
		border-color: #333;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
	}

	.import-area svg {
		width: 32px;
		height: 32px;
		color: #888;
	}

	.import-text {
		font-size: 14px;
		font-weight: 500;
		color: #333;
	}

	.import-hint {
		font-size: 11px;
		color: #999;
	}

	.schema-preview {
		background: #1e1e1e;
		border-radius: 8px;
		padding: 16px;
		overflow-x: auto;
	}

	.schema-preview pre {
		margin: 0;
	}

	.schema-preview code {
		font-family: 'SF Mono', Monaco, 'Fira Code', monospace;
		font-size: 11px;
		line-height: 1.6;
		color: #9cdcfe;
	}

	.status-banner {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		margin-bottom: 16px;
	}

	.status-banner svg {
		width: 18px;
		height: 18px;
		flex-shrink: 0;
	}

	.status-banner.success {
		background: #d4edda;
		color: #155724;
		border: 1px solid #c3e6cb;
	}

	.status-banner.error {
		background: #f8d7da;
		color: #721c24;
		border: 1px solid #f5c6cb;
	}

	/* Permissions tab styles */
	.permissions-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 16px;
	}

	.permission-row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 16px;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 10px;
		transition: border-color 0.15s ease;
	}

	.permission-row:hover {
		border-color: #ccc;
	}

	.permission-icon {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f5f5f5;
		border-radius: 10px;
		flex-shrink: 0;
		transition: all 0.15s ease;
	}

	.permission-icon.granted {
		background: #d4edda;
	}

	.permission-icon svg {
		width: 22px;
		height: 22px;
		color: #666;
	}

	.permission-icon.granted svg {
		color: #28a745;
	}

	.permission-info {
		flex: 1;
		min-width: 0;
	}

	.permission-name {
		font-size: 14px;
		font-weight: 600;
		color: #333;
	}

	.permission-desc {
		font-size: 12px;
		color: #666;
		margin-top: 3px;
		line-height: 1.4;
	}

	.permission-status {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
	}

	.grant-btn {
		padding: 8px 16px;
		background: #3b82f6;
		border: none;
		border-radius: 6px;
		color: white;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.grant-btn:hover {
		background: #2563eb;
	}

	.settings-btn {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f5f5f5;
		border: 1px solid #ddd;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.settings-btn:hover {
		background: #eee;
		border-color: #ccc;
	}

	.settings-btn svg {
		width: 18px;
		height: 18px;
		color: #666;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 6px 12px;
		border-radius: 6px;
		font-size: 12px;
		font-weight: 600;
	}

	.status-badge.granted {
		background: #d4edda;
		color: #155724;
	}

	.status-badge svg {
		width: 14px;
		height: 14px;
	}

	.recheck-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
		padding: 12px 16px;
		margin-top: 16px;
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
		color: #555;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.recheck-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.recheck-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.recheck-btn svg {
		width: 16px;
		height: 16px;
	}

	.recheck-btn .spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
