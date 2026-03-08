<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowLeft, ArrowRight, Save, Eye } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleEditor from '$lib/components/modules/ModuleEditor.svelte';
	import ActionBuilder from '$lib/components/modules/ActionBuilder.svelte';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import type { ModuleCategory, ModuleAction, ModuleManifest } from '$lib/types/modules';

	let store = $state(customModulesStore);

	let currentStep = $state(1);
	let isCreating = $state(false);

	// Step 1: Basic Info
	let name = $state('');
	let description = $state('');
	let category = $state<ModuleCategory>('custom');
	let icon = $state('📦');

	// Step 2: Actions
	let actions = $state<ModuleAction[]>([]);

	// Step 3: Configuration
	let configSchema = $state<Record<string, unknown>>({});
	let configSchemaText = $state('{}');
	let visibility = $state<'private' | 'workspace' | 'public'>('private');

	const manifest = $derived<ModuleManifest>({
		name,
		version: '1.0.0',
		description,
		author: 'Current User',
		category,
		icon,
		actions,
		config_schema: configSchema,
		dependencies: [],
		permissions: []
	});

	function handleNext() {
		if (currentStep === 1 && !name.trim()) {
			alert('Please enter a module name');
			return;
		}
		if (currentStep === 1 && !description.trim()) {
			alert('Please enter a description');
			return;
		}
		if (currentStep < 3) {
			currentStep++;
		}
	}

	function handlePrevious() {
		if (currentStep > 1) {
			currentStep--;
		}
	}

	function handleConfigSchemaChange(value: string) {
		configSchemaText = value;
		try {
			configSchema = JSON.parse(value);
		} catch {
			// Invalid JSON, keep old value
		}
	}

	async function handleCreate(isDraft: boolean) {
		if (!name.trim() || !description.trim()) {
			alert('Please fill in all required fields');
			return;
		}

		isCreating = true;
		const module = await store.createModule({
			name,
			description,
			category,
			icon,
			manifest,
			config_schema: configSchema,
			visibility
		});

		if (module) {
			goto(`/modules/${module.id}`);
		}
		isCreating = false;
	}
</script>

<div class="am-create-page">
	<!-- Header -->
	<div class="am-create-header">
		<!-- Back Button -->
		<button
			onclick={() => goto('/modules')}
			class="am-back-btn"
			aria-label="Back to Modules"
		>
			<ArrowLeft class="w-4 h-4" />
			<span>Back to Modules</span>
		</button>

		<!-- Title -->
		<h1 class="am-create-title">Create Custom Module</h1>
		<p class="am-create-subtitle">Build a custom module for your workspace</p>

		<!-- Step Indicator -->
		<div class="am-stepper">
			{#each [1, 2, 3] as step}
				<div class="am-stepper__item">
					<div class="am-stepper__circle {currentStep === step ? 'am-stepper__circle--active' : currentStep > step ? 'am-stepper__circle--done' : ''}">
						{step}
					</div>
					<span class="am-stepper__label {currentStep === step ? 'am-stepper__label--active' : ''}">
						{step === 1 ? 'Basic Info' : step === 2 ? 'Actions' : 'Configuration'}
					</span>
					{#if step < 3}
						<div class="am-stepper__line {currentStep > step ? 'am-stepper__line--done' : ''}"></div>
					{/if}
				</div>
			{/each}
		</div>
	</div>

	<!-- Content -->
	<div class="am-create-content">
		<div class="am-create-inner">
			{#if currentStep === 1}
				<!-- Step 1: Basic Info -->
				<ModuleEditor
					{name}
					{description}
					{category}
					{icon}
					onNameChange={(v) => name = v}
					onDescriptionChange={(v) => description = v}
					onCategoryChange={(v) => category = v}
					onIconChange={(v) => icon = v}
				/>
			{:else if currentStep === 2}
				<!-- Step 2: Actions -->
				<div class="am-create-section">
					<h2 class="am-section-title">Module Actions</h2>
					<p class="am-section-desc">Define the actions your module provides</p>
				</div>
				<ActionBuilder
					{actions}
					onActionsChange={(a) => actions = a}
				/>
			{:else if currentStep === 3}
				<!-- Step 3: Configuration -->
				<div class="am-create-section">
					<h2 class="am-section-title">Configuration</h2>
					<p class="am-section-desc">Set up module configuration and visibility</p>
				</div>

				<!-- Config Schema -->
				<div class="am-form-group">
					<label class="am-form-label">
						Configuration Schema (JSON)
					</label>
					<textarea
						value={configSchemaText}
						oninput={(e) => handleConfigSchemaChange(e.currentTarget.value)}
						placeholder={'{"setting1": "string", "setting2": "number"}'}
						rows="6"
						class="am-form-textarea"
						aria-label="Configuration schema JSON"
					></textarea>
				</div>

				<!-- Visibility -->
				<div class="am-form-group">
					<label class="am-form-label">Visibility</label>
					<div class="am-radio-group">
						{#each [
							{ value: 'private', label: 'Private', desc: 'Only you can see and use this module' },
							{ value: 'workspace', label: 'Workspace', desc: 'Available to your entire workspace' },
							{ value: 'public', label: 'Public', desc: 'Available to everyone' }
						] as opt}
							<label class="am-radio-card {visibility === opt.value ? 'am-radio-card--active' : ''}">
								<input
									type="radio"
									name="visibility"
									value={opt.value}
									checked={visibility === opt.value}
									onchange={() => visibility = opt.value as 'private' | 'workspace' | 'public'}
									class="am-radio-input"
									aria-label={opt.label}
								/>
								<div>
									<p class="am-radio-card__title">{opt.label}</p>
									<p class="am-radio-card__desc">{opt.desc}</p>
								</div>
							</label>
						{/each}
					</div>
				</div>

				<!-- Manifest Preview -->
				<div class="am-form-group">
					<div class="am-form-label-row">
						<Eye class="w-4 h-4" />
						<label class="am-form-label">Manifest Preview</label>
					</div>
					<ManifestViewer {manifest} />
				</div>
			{/if}
		</div>
	</div>

	<!-- Footer Actions -->
	<div class="am-create-footer">
		<div class="am-create-footer__inner">
			<div>
				{#if currentStep > 1}
					<button
						onclick={handlePrevious}
						class="am-icon-btn"
						aria-label="Previous step"
					>
						<ArrowLeft class="w-4 h-4" />
						<span>Previous</span>
					</button>
				{/if}
			</div>
			<div class="am-create-footer__actions">
				{#if currentStep === 3}
					<button
						onclick={() => handleCreate(true)}
						disabled={isCreating}
						class="am-icon-btn"
						aria-label="Save draft"
					>
						<Save class="w-4 h-4" />
						<span>Save Draft</span>
					</button>
					<button
						onclick={() => handleCreate(false)}
						disabled={isCreating}
						class="am-btn am-btn--install"
						aria-label="Create module"
					>
						<span>{isCreating ? 'Creating...' : 'Create Module'}</span>
					</button>
				{:else}
					<button
						onclick={handleNext}
						class="am-btn am-btn--install"
						aria-label="Next step"
					>
						<span>Next</span>
						<ArrowRight class="w-4 h-4" />
					</button>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  CREATE MODULE PAGE (am-create-) — Foundation Tokens         */
	/* ══════════════════════════════════════════════════════════════ */
	.am-create-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg, #fff);
	}
	.am-create-header {
		flex-shrink: 0;
		padding: 20px 32px 16px;
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
	}
	.am-create-title {
		font-size: 22px;
		font-weight: 700;
		color: var(--dt, #111);
		margin-bottom: 4px;
	}
	.am-create-subtitle {
		font-size: 13px;
		color: var(--dt2, #555);
	}

	/* Back button */
	.am-back-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: var(--dt3, #888);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		margin-bottom: 16px;
		transition: color .15s;
	}
	.am-back-btn:hover {
		color: var(--dt, #111);
	}

	/* Stepper */
	.am-stepper {
		display: flex;
		align-items: center;
		gap: 4px;
		margin-top: 20px;
	}
	.am-stepper__item {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.am-stepper__circle {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 13px;
		font-weight: 600;
		background: var(--dbg3, #eee);
		color: var(--dt3, #888);
		transition: all .2s;
	}
	.am-stepper__circle--active {
		background: var(--accent-blue, #3b82f6);
		color: #fff;
	}
	.am-stepper__circle--done {
		background: #10b981;
		color: #fff;
	}
	.am-stepper__label {
		font-size: 13px;
		font-weight: 500;
		color: var(--dt3, #888);
	}
	.am-stepper__label--active {
		color: var(--dt, #111);
	}
	.am-stepper__line {
		width: 40px;
		height: 2px;
		background: var(--dbg3, #eee);
		margin: 0 4px;
		transition: background .2s;
	}
	.am-stepper__line--done {
		background: #10b981;
	}

	/* Content */
	.am-create-content {
		flex: 1;
		overflow-y: auto;
	}
	.am-create-inner {
		max-width: 800px;
		margin: 0 auto;
		padding: 24px 32px;
	}
	.am-create-section {
		margin-bottom: 16px;
	}
	.am-section-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111);
		margin-bottom: 4px;
	}
	.am-section-desc {
		font-size: 13px;
		color: var(--dt3, #888);
	}

	/* Form groups */
	.am-form-group {
		margin-bottom: 20px;
	}
	.am-form-label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: var(--dt2, #555);
		margin-bottom: 8px;
	}
	.am-form-label-row {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--dt2, #555);
		margin-bottom: 8px;
	}
	.am-form-label-row .am-form-label {
		margin-bottom: 0;
	}
	.am-form-textarea {
		width: 100%;
		padding: 10px 14px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
		font-family: monospace;
		font-size: 13px;
		outline: none;
		transition: border-color .15s;
		resize: vertical;
	}
	.am-form-textarea:focus {
		border-color: var(--accent-blue, #3b82f6);
	}

	/* Radio cards */
	.am-radio-group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.am-radio-card {
		display: flex;
		align-items: center;
		padding: 14px 16px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 12px;
		cursor: pointer;
		transition: all .15s;
		background: var(--dbg, #fff);
	}
	.am-radio-card:hover {
		border-color: var(--dbd2, #f0f0f0);
	}
	.am-radio-card--active {
		border-color: var(--accent-blue, #3b82f6);
		background: rgba(59, 130, 246, 0.04);
	}
	.am-radio-input {
		margin-right: 12px;
		accent-color: var(--accent-blue, #3b82f6);
	}
	.am-radio-card__title {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.am-radio-card__desc {
		font-size: 12px;
		color: var(--dt3, #888);
	}

	/* Footer */
	.am-create-footer {
		flex-shrink: 0;
		border-top: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg2, #f5f5f5);
		padding: 12px 32px;
	}
	.am-create-footer__inner {
		max-width: 800px;
		margin: 0 auto;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.am-create-footer__actions {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* Buttons (shared with detail) */
	.am-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 8px 18px;
		border-radius: 999px;
		border: none;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all .18s;
		white-space: nowrap;
		gap: 6px;
	}
	.am-btn--install {
		background: var(--accent-blue, #3b82f6);
		color: #fff;
	}
	.am-btn--install:hover {
		filter: brightness(0.9);
		transform: translateY(-1px);
	}
	.am-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
		transform: none;
	}
	.am-icon-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		border-radius: 999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: transparent;
		color: var(--dt2, #555);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all .15s;
	}
	.am-icon-btn:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
	.am-icon-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
