<script lang="ts">
	import { onMount } from 'svelte';
	import { thinking } from '$lib/stores/thinking';
	import { Brain, Loader2, AlertCircle, Sparkles, Save, RefreshCw, Settings as SettingsIcon } from 'lucide-svelte';
	import type { ThinkingSettings, ReasoningTemplate, UpdateSettingsData } from '$lib/api/thinking/types';

	// State
	let isLoading = $state(true);
	let isSaving = $state(false);
	let error = $state<string | null>(null);
	let saveStatus = $state('');

	// Local settings (for editing before save)
	let thinkingEnabled = $state(true);
	let showInUI = $state(false);
	let saveTraces = $state(true);
	let maxTokens = $state(4096);
	let selectedTemplateId = $state<string | undefined>(undefined);

	// Templates
	let templates = $state<ReasoningTemplate[]>([]);

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		try {
			isLoading = true;
			error = null;

			// Load settings and templates in parallel
			const [settingsData, _] = await Promise.all([
				thinking.loadSettings(),
				thinking.loadTemplates()
			]);

			// Get templates from store
			const unsubscribe = thinking.subscribe(state => {
				templates = state.templates;
			});
			unsubscribe();

			if (settingsData) {
				// Update local state
				thinkingEnabled = settingsData.enabled ?? true;
				showInUI = settingsData.show_in_ui ?? false;
				saveTraces = settingsData.save_traces ?? true;
				maxTokens = settingsData.max_tokens ?? 4096;
				selectedTemplateId = settingsData.default_template_id ?? undefined;
			}
		} catch (err) {
			console.error('Failed to load thinking settings:', err);
			error = err instanceof Error ? err.message : 'Failed to load settings';
		} finally {
			isLoading = false;
		}
	}

	async function saveSettings() {
		try {
			isSaving = true;
			error = null;
			saveStatus = '';

			const updatedSettings: UpdateSettingsData = {
				enabled: thinkingEnabled,
				show_in_ui: showInUI,
				save_traces: saveTraces,
				max_tokens: maxTokens,
				default_template_id: selectedTemplateId ?? null
			};

			await thinking.updateSettings(updatedSettings);

			saveStatus = 'Settings saved successfully!';
			setTimeout(() => saveStatus = '', 3000);
		} catch (err) {
			console.error('Failed to save thinking settings:', err);
			error = err instanceof Error ? err.message : 'Failed to save settings';
			saveStatus = '';
		} finally {
			isSaving = false;
		}
	}

	function resetToDefaults() {
		thinkingEnabled = true;
		showInUI = false;
		saveTraces = true;
		maxTokens = 4096;
		selectedTemplateId = undefined;
	}

	function getTokensLabel(tokens: number): string {
		if (tokens >= 1024) {
			return `${(tokens / 1024).toFixed(1)}K tokens`;
		}
		return `${tokens} tokens`;
	}
</script>

<div class="thinking-settings-page">
	<div class="settings-header">
		<div class="header-content">
			<div class="header-title">
				<Brain class="w-6 h-6" />
				<h1>Thinking Settings</h1>
			</div>
			<p class="header-subtitle">Configure Extended Thinking (Chain of Thought) display and behavior</p>
		</div>
	</div>

	{#if isLoading}
		<div class="loading-state">
			<Loader2 class="w-8 h-8 animate-spin" />
			<p>Loading thinking settings...</p>
		</div>
	{:else if error && !isSaving}
		<div class="error-state">
			<AlertCircle class="w-8 h-8" />
			<p>{error}</p>
			<button class="btn-pill btn-pill-ghost retry-btn" onclick={loadSettings}>
				<RefreshCw class="w-4 h-4" />
				Retry
			</button>
		</div>
	{:else}
		<div class="settings-container">
			<!-- How It Works Section -->
			<section class="info-section">
				<div class="info-header">
					<Sparkles class="w-5 h-5 text-purple-600" />
					<h2>How Extended Thinking Works</h2>
				</div>
				<div class="info-content">
					<p class="info-text">
						Extended Thinking (Chain of Thought) allows the AI to show its reasoning process before providing an answer.
						This helps with complex queries by breaking down the thinking into visible steps.
					</p>
					<div class="info-grid">
						<div class="info-card">
							<div class="info-card-header">
								<span class="info-number">1</span>
								<h3>Enable Thinking</h3>
							</div>
							<p>Toggle "Enable Extended Thinking" below to allow the AI to use Chain of Thought reasoning.</p>
						</div>
						<div class="info-card">
							<div class="info-card-header">
								<span class="info-number">2</span>
								<h3>Show in UI</h3>
							</div>
							<p>Enable "Show Thinking by Default" to display thinking traces in chat messages automatically.</p>
						</div>
						<div class="info-card">
							<div class="info-card-header">
								<span class="info-number">3</span>
								<h3>Save Traces</h3>
							</div>
							<p>Save thinking traces to database for later analysis and debugging of AI reasoning.</p>
						</div>
						<div class="info-card">
							<div class="info-card-header">
								<span class="info-number">4</span>
								<h3>Templates</h3>
							</div>
							<p>Create custom reasoning templates to guide the AI's thinking process for specific tasks.</p>
						</div>
					</div>
					<div class="integration-note">
						<strong>Integration:</strong> The thinking system works automatically with the chat. When enabled,
						you'll see a collapsible "Thinking" panel in AI responses showing the reasoning steps.
						The ThinkingPanel component displays: analyzing → planning → reasoning → concluding steps.
					</div>
				</div>
			</section>

			<!-- Thinking Display Settings -->
			<section class="section">
				<div class="section-header">
					<h2>Display Settings</h2>
					<span class="subtitle">Control how thinking traces are shown</span>
				</div>

				<div class="settings-grid">
					<div class="setting-card toggle-card">
						<div class="setting-header">
							<label for="thinkingEnabled">Enable Extended Thinking</label>
							<button
								class="btn-pill btn-pill-ghost toggle"
								class:on={thinkingEnabled}
								onclick={() => thinkingEnabled = !thinkingEnabled}
							>
								<span class="toggle-knob"></span>
							</button>
						</div>
						<p class="setting-desc">Allow the AI to use extended thinking (Chain of Thought) for complex queries</p>
					</div>

					<div class="setting-card toggle-card">
						<div class="setting-header">
							<label for="showInUI">Show Thinking by Default</label>
							<button
								class="btn-pill btn-pill-ghost toggle"
								class:on={showInUI}
								class:disabled={!thinkingEnabled}
								onclick={() => {
									if (thinkingEnabled) showInUI = !showInUI;
								}}
							>
								<span class="toggle-knob"></span>
							</button>
						</div>
						<p class="setting-desc">Display thinking process automatically in chat (can be toggled per message)</p>
					</div>

					<div class="setting-card toggle-card">
						<div class="setting-header">
							<label for="saveTraces">Save Thinking Traces</label>
							<button
								class="btn-pill btn-pill-ghost toggle"
								class:on={saveTraces}
								class:disabled={!thinkingEnabled}
								onclick={() => {
									if (thinkingEnabled) saveTraces = !saveTraces;
								}}
							>
								<span class="toggle-knob"></span>
							</button>
						</div>
						<p class="setting-desc">Store thinking traces for later review and analysis</p>
					</div>
				</div>
			</section>

			<!-- Performance Settings -->
			<section class="section">
				<div class="section-header">
					<h2>Performance Settings</h2>
					<span class="subtitle">Adjust thinking token limits</span>
				</div>

				<div class="setting-card">
					<div class="setting-header">
						<label for="maxTokens">Max Thinking Tokens</label>
						<span class="setting-value">{getTokensLabel(maxTokens)}</span>
					</div>
					<input
						type="range"
						id="maxTokens"
						bind:value={maxTokens}
						min="0"
						max="8192"
						step="256"
						disabled={!thinkingEnabled}
					/>
					<p class="setting-desc">
						Maximum tokens to use for thinking process. Higher values allow deeper reasoning but increase response time.
						{#if maxTokens === 0}
							<span class="warning-text">Warning: Setting to 0 may disable extended thinking features.</span>
						{/if}
					</p>
				</div>
			</section>

			<!-- Template Settings -->
			<section class="section">
				<div class="section-header">
					<h2>Reasoning Templates</h2>
					<span class="subtitle">Choose default thinking template</span>
				</div>

				<div class="setting-card">
					<div class="setting-header">
						<label for="defaultTemplate">Default Template</label>
					</div>
					<p class="setting-desc" style="margin-bottom: 1rem;">
						Select a reasoning template to guide the AI's thinking process
					</p>

					{#if templates.length === 0}
						<div class="empty-state">
							<Sparkles class="w-6 h-6" />
							<p>No reasoning templates found</p>
							<p class="empty-hint">Create templates to customize thinking behavior</p>
						</div>
					{:else}
						<select
							id="defaultTemplate"
							bind:value={selectedTemplateId}
							disabled={!thinkingEnabled}
							class="template-select"
						>
							<option value={undefined}>None (use default reasoning)</option>
							{#each templates as template}
								<option value={template.id}>
									{template.name}
									{#if template.is_default} (Current Default){/if}
								</option>
							{/each}
						</select>

						{#if selectedTemplateId}
							{@const selectedTemplate = templates.find(t => t.id === selectedTemplateId)}
							{#if selectedTemplate?.description}
								<p class="template-description">{selectedTemplate.description}</p>
							{/if}
						{/if}
					{/if}

					<a href="/settings/ai/thinking/templates" class="manage-link">
						<SettingsIcon class="w-4 h-4" />
						Manage Templates
					</a>
				</div>
			</section>

			<!-- Action Buttons -->
			<div class="settings-actions">
				<button
					class="btn-pill btn-pill-ghost action-btn primary"
					onclick={saveSettings}
					disabled={isSaving}
				>
					{#if isSaving}
						<Loader2 class="w-4 h-4 animate-spin" />
						Saving...
					{:else}
						<Save class="w-4 h-4" />
						Save Settings
					{/if}
				</button>

				<button class="btn-pill btn-pill-ghost action-btn" onclick={resetToDefaults}>
					<RefreshCw class="w-4 h-4" />
					Reset to Defaults
				</button>
			</div>

			{#if saveStatus}
				<div class="save-success">
					{saveStatus}
				</div>
			{/if}

			{#if error && isSaving}
				<div class="save-error">
					<AlertCircle class="w-4 h-4" />
					{error}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	* {
		box-sizing: border-box;
	}

	.thinking-settings-page {
		height: 100%;
		background: var(--bos-settings-card-bg);
		overflow-y: auto;
		overflow-x: hidden;
		display: flex;
		flex-direction: column;
	}

	.settings-header {
		background: var(--dbg);
		border-bottom: 1px solid var(--dbd);
		padding: 1.5rem 2rem;
		flex-shrink: 0;
	}

	.header-content {
		max-width: 1200px;
		margin: 0 auto;
	}

	.header-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.5rem;
	}

	.header-title h1 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.header-subtitle {
		color: var(--dt3);
		font-size: 0.875rem;
		margin: 0;
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		gap: 1rem;
		color: var(--dt3);
		flex: 1;
	}

	.error-state {
		color: var(--bos-status-error);
	}

	.retry-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: var(--bos-nav-active);
		color: var(--bos-surface-on-color);
		border: none;
		border-radius: 0.375rem;
		cursor: pointer;
		font-size: 0.875rem;
		transition: background 0.15s;
	}

	.retry-btn:hover {
		background: var(--bos-nav-active);
	}

	.settings-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
		padding-bottom: 6rem;
		box-sizing: border-box;
		width: 100%;
		flex: 1;
	}

	/* Info Section */
	.info-section {
		background: linear-gradient(135deg, var(--bos-category-productivity) 0%, var(--bos-category-ai) 100%);
		border-radius: 0.75rem;
		padding: 2rem;
		margin-bottom: 2rem;
		color: var(--bos-surface-on-color);
		box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
		width: 100%;
		max-width: 100%;
	}

	.info-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1.5rem;
	}

	.info-header h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
	}

	.info-content {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.info-text {
		font-size: 1rem;
		line-height: 1.6;
		margin: 0;
		opacity: 0.95;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 250px), 1fr));
		gap: 1rem;
		width: 100%;
	}

	.info-card {
		background: rgba(255, 255, 255, 0.15);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 0.5rem;
		padding: 1.25rem;
		transition: all 0.3s;
	}

	.info-card:hover {
		background: rgba(255, 255, 255, 0.2);
		transform: translateY(-2px);
	}

	.info-card-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}

	.info-number {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		background: rgba(255, 255, 255, 0.9);
		color: var(--bos-category-productivity);
		font-weight: 700;
		font-size: 1rem;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.info-card h3 {
		font-size: 1rem;
		font-weight: 600;
		margin: 0;
	}

	.info-card p {
		margin: 0;
		font-size: 0.875rem;
		line-height: 1.5;
		opacity: 0.9;
	}

	.integration-note {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 0.5rem;
		padding: 1rem;
		font-size: 0.875rem;
		line-height: 1.6;
	}

	.integration-note strong {
		font-weight: 600;
		color: #fbbf24;
	}

	.section {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
		width: 100%;
		max-width: 100%;
	}

	.section-header {
		margin-bottom: 1.5rem;
	}

	.section-header h2 {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0 0 0.25rem 0;
	}

	.subtitle {
		font-size: 0.875rem;
		color: var(--dt3);
	}

	.settings-grid {
		display: grid;
		gap: 1rem;
	}

	.setting-card {
		padding: 1rem;
		background: var(--bos-settings-card-bg);
		border: 1px solid var(--bos-settings-card-border);
		border-radius: var(--bos-settings-card-radius);
	}

	.setting-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.setting-header label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt);
	}

	.setting-value {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--bos-nav-active);
	}

	.setting-desc {
		font-size: 0.875rem;
		color: var(--dt3);
		margin: 0;
		line-height: 1.5;
	}

	.warning-text {
		color: var(--bos-status-warning);
		font-weight: 500;
	}

	/* Toggle Switch */
	.toggle {
		position: relative;
		width: 44px;
		height: 24px;
		background: var(--dbd);
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: background 0.2s;
	}

	.toggle:hover:not(.disabled) {
		background: var(--dt3);
	}

	.toggle.on {
		background: var(--bos-nav-active);
	}

	.toggle.on:hover:not(.disabled) {
		background: var(--bos-nav-active);
	}

	.toggle.disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.toggle-knob {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		background: var(--dbg);
		border-radius: 50%;
		transition: transform 0.2s;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
	}

	.toggle.on .toggle-knob {
		transform: translateX(20px);
	}

	/* Range Slider */
	input[type="range"] {
		width: 100%;
		height: 6px;
		background: var(--dbg3);
		border-radius: 3px;
		outline: none;
		margin-bottom: 0.75rem;
	}

	input[type="range"]::-webkit-slider-thumb {
		appearance: none;
		width: 18px;
		height: 18px;
		background: var(--dt);
		border-radius: 50%;
		cursor: pointer;
	}

	input[type="range"]::-moz-range-thumb {
		width: 18px;
		height: 18px;
		background: var(--dt);
		border: none;
		border-radius: 50%;
		cursor: pointer;
	}

	input[type="range"]:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	input[type="range"]:disabled::-webkit-slider-thumb {
		cursor: not-allowed;
	}

	input[type="range"]:disabled::-moz-range-thumb {
		cursor: not-allowed;
	}

	/* Template Select */
	.template-select {
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: var(--bos-settings-input-bg);
		border: 1px solid var(--bos-settings-card-border);
		border-radius: 0.375rem;
		font-size: 0.875rem;
		color: var(--dt);
		cursor: pointer;
		margin-bottom: 0.5rem;
	}

	.template-select:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		background: var(--bos-settings-input-bg);
	}

	.template-select:focus {
		outline: none;
		border-color: var(--bos-nav-active);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.template-description {
		font-size: 0.8125rem;
		color: var(--dt3);
		padding: 0.5rem;
		background: var(--bos-settings-input-bg);
		border: 1px solid var(--bos-settings-card-border);
		border-radius: 0.25rem;
		margin-bottom: 0.75rem;
	}

	.manage-link {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.875rem;
		color: var(--bos-nav-active);
		text-decoration: none;
		margin-top: 0.75rem;
	}

	.manage-link:hover {
		color: var(--bos-nav-active);
		text-decoration: underline;
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 2rem;
		gap: 0.5rem;
		color: var(--dt3);
	}

	.empty-hint {
		font-size: 0.875rem;
		color: var(--dt3);
	}

	/* Action Buttons */
	.settings-actions {
		display: flex;
		gap: 1rem;
		padding: 1.5rem 2rem;
		border-top: 1px solid var(--dbd);
		background: var(--dbg);
		position: sticky;
		bottom: 0;
		z-index: 10;
		box-shadow: 0 -4px 6px -1px rgba(0, 0, 0, 0.05);
		margin-top: 2rem;
		margin-left: -2rem;
		margin-right: -2rem;
		width: calc(100% + 4rem);
		flex-shrink: 0;
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 1.25rem;
		border: 1px solid var(--dbd);
		border-radius: 0.375rem;
		background: var(--dbg);
		color: var(--dt);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.action-btn:hover:not(:disabled) {
		background: var(--dbg2);
		border-color: var(--dt3);
	}

	.action-btn.primary {
		background: var(--bos-nav-active);
		color: var(--bos-surface-on-color);
		border-color: var(--bos-nav-active);
	}

	.action-btn.primary:hover:not(:disabled) {
		background: var(--bos-nav-active);
		border-color: var(--bos-nav-active);
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Status Messages */
	.save-success {
		margin-top: 1rem;
		padding: 0.75rem 1rem;
		background: var(--bos-status-success-bg, #d1fae5);
		color: var(--bos-status-success, #065f46);
		border-radius: 0.375rem;
		font-size: 0.875rem;
		text-align: center;
	}

	.save-error {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		margin-top: 1rem;
		padding: 0.75rem 1rem;
		background: var(--bos-status-error-bg);
		color: var(--bos-status-error);
		border-radius: 0.375rem;
		font-size: 0.875rem;
	}

	/* Dark Mode - all colors now use CSS variables, no overrides needed */
</style>
