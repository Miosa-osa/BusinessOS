<script lang="ts">
	import { goto } from '$app/navigation';
	import { fly, scale } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import { ArrowLeft, ArrowRight, Save, Eye, Check, Sparkles } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleEditor from '$lib/components/modules/ModuleEditor.svelte';
	import ActionBuilder from '$lib/components/modules/ActionBuilder.svelte';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import type { ModuleCategory, ModuleAction, ModuleManifest } from '$lib/types/modules';

	const store = customModulesStore;

	let currentStep = $state(1);
	let isCreating = $state(false);
	let shakeStep = $state(false);
	let hasVisited = $state<Record<number, boolean>>({ 1: true });

	// Step 1: Basic Info
	let name = $state('');
	let description = $state('');
	let category = $state<ModuleCategory>('custom');

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
		actions,
		config_schema: configSchema,
		dependencies: [],
		permissions: []
	});

	const stepLabels = ['Basic Info', 'Actions', 'Configuration'];
	const stepDescs = [
		'Name, describe, and categorize your module',
		'Define the actions your module provides',
		'Configure schema, visibility, and review'
	];

	const progressPercent = $derived(((currentStep - 1) / 2) * 100);
	const step1Valid = $derived(name.trim().length > 0 && description.trim().length > 0);
	const canProceed = $derived(currentStep === 1 ? step1Valid : true);

	function handleNext() {
		if (currentStep === 1 && (!name.trim() || !description.trim())) {
			shakeStep = true;
			setTimeout(() => shakeStep = false, 500);
			return;
		}
		if (currentStep < 3) {
			currentStep++;
			hasVisited[currentStep] = true;
		}
	}

	function handlePrevious() {
		if (currentStep > 1) {
			currentStep--;
		}
	}

	function goToStep(step: number) {
		if (hasVisited[step] || step < currentStep) {
			currentStep = step;
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
		if (!name.trim() || !description.trim()) return;

		isCreating = true;
		const module = await store.createModule({
			name,
			description,
			category,
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

<div class="mc-page">
	<!-- Header -->
	<div class="mc-header">
		<button onclick={() => goto('/modules')} class="mc-back" aria-label="Back to Modules">
			<ArrowLeft class="w-4 h-4" />
			<span>Back to Modules</span>
		</button>

		<div class="mc-header__title-row">
			<div>
				<h1 class="mc-title">Create Custom Module</h1>
				<p class="mc-subtitle">{stepDescs[currentStep - 1]}</p>
			</div>
			<div class="mc-step-badge">Step {currentStep} of 3</div>
		</div>

		<!-- Step Indicator with animated progress -->
		<div class="mc-steps">
			<div class="mc-steps__track">
				<div class="mc-steps__line-bg"></div>
				<div class="mc-steps__line-fill" style="width: {progressPercent}%;"></div>
				{#each [1, 2, 3] as step}
					<button
						class="mc-step__dot"
						class:mc-step__dot--active={currentStep === step}
						class:mc-step__dot--done={currentStep > step}
						class:mc-step__dot--clickable={hasVisited[step] || step < currentStep}
						onclick={() => goToStep(step)}
						aria-label="Go to step {step}: {stepLabels[step - 1]}"
					>
						{#if currentStep > step}
							<div in:scale={{ duration: 350, easing: elasticOut, start: 0.3 }}>
								<Check class="w-3.5 h-3.5" />
							</div>
						{:else}
							{step}
						{/if}
					</button>
				{/each}
			</div>
			<div class="mc-steps__labels">
				{#each [1, 2, 3] as step}
					<span
						class="mc-step__label"
						class:mc-step__label--active={currentStep === step}
						class:mc-step__label--done={currentStep > step}
					>
						{stepLabels[step - 1]}
					</span>
				{/each}
			</div>
		</div>
	</div>

	<!-- Carousel Content -->
	<div class="mc-content" class:mc-content--shake={shakeStep}>
		<div
			class="mc-carousel"
			style="transform: translateX(-{(currentStep - 1) * 100}%);"
		>
			<!-- Step 1: Basic Info -->
			<div class="mc-carousel__slide">
				<div class="mc-slide-inner">
					<div class="mc-section">
						<ModuleEditor
							{name}
							{description}
							{category}
							onNameChange={(v) => name = v}
							onDescriptionChange={(v) => description = v}
							onCategoryChange={(v) => category = v}
						/>
					</div>
				</div>
			</div>

			<!-- Step 2: Actions -->
			<div class="mc-carousel__slide">
				<div class="mc-slide-inner">
					{#if hasVisited[2]}
						<div class="mc-section" in:fly={{ y: 24, duration: 400, delay: 80, easing: cubicOut }}>
							<div class="mc-section__header">
								<h2 class="mc-section__title">Module Actions</h2>
								<p class="mc-section__desc">Define the actions your module provides. Actions are the building blocks of what your module can do.</p>
							</div>
							<ActionBuilder
								{actions}
								onActionsChange={(a) => actions = a}
							/>
						</div>
					{/if}
				</div>
			</div>

			<!-- Step 3: Configuration -->
			<div class="mc-carousel__slide">
				<div class="mc-slide-inner">
					{#if hasVisited[3]}
						<div class="mc-section">
							<div class="mc-section__header" in:fly={{ y: 20, duration: 350, easing: cubicOut }}>
								<h2 class="mc-section__title">Configuration</h2>
								<p class="mc-section__desc">Set up module configuration and visibility settings</p>
							</div>

							<div class="mc-field" in:fly={{ y: 20, duration: 350, delay: 60, easing: cubicOut }}>
								<label class="mc-label">Configuration Schema (JSON)</label>
								<textarea
									value={configSchemaText}
									oninput={(e) => handleConfigSchemaChange(e.currentTarget.value)}
									placeholder={'{"setting1": "string", "setting2": "number"}'}
									rows="6"
									class="mc-textarea"
								></textarea>
							</div>

							<div class="mc-field" in:fly={{ y: 20, duration: 350, delay: 120, easing: cubicOut }}>
								<label class="mc-label">Visibility</label>
								<div class="mc-radio-group">
									{#each [
										{ value: 'private', title: 'Private', desc: 'Only you can see and use this module' },
										{ value: 'workspace', title: 'Workspace', desc: 'Available to your entire workspace' },
										{ value: 'public', title: 'Public', desc: 'Available to everyone' }
									] as opt, i}
										<label
											class="mc-radio"
											class:mc-radio--active={visibility === opt.value}
											in:fly={{ y: 16, duration: 300, delay: 180 + i * 50, easing: cubicOut }}
										>
											<input
												type="radio"
												name="visibility"
												value={opt.value}
												checked={visibility === opt.value}
												onchange={() => visibility = opt.value as 'private' | 'workspace' | 'public'}
												class="mc-radio__input"
											/>
											<div>
												<p class="mc-radio__title">{opt.title}</p>
												<p class="mc-radio__desc">{opt.desc}</p>
											</div>
										</label>
									{/each}
								</div>
							</div>

							<div class="mc-field" in:fly={{ y: 20, duration: 350, delay: 350, easing: cubicOut }}>
								<div class="mc-label-row">
									<Eye class="w-4 h-4" style="color: var(--dt3, #888);" />
									<label class="mc-label">Manifest Preview</label>
								</div>
								<ManifestViewer {manifest} />
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>

	<!-- Footer Actions -->
	<div class="mc-footer">
		<div class="mc-footer__inner">
			<div>
				{#if currentStep > 1}
					<button
						onclick={handlePrevious}
						class="btn-pill btn-pill-ghost"
						aria-label="Previous step"
					>
						<ArrowLeft class="w-4 h-4" />
						<span>Previous</span>
					</button>
				{/if}
			</div>
			<div class="mc-footer__actions">
				{#if currentStep === 3}
					<button
						onclick={() => handleCreate(true)}
						disabled={isCreating}
						class="btn-pill btn-pill-ghost"
						aria-label="Save as draft"
					>
						<Save class="w-4 h-4" />
						<span>Save Draft</span>
					</button>
					<button
						onclick={() => handleCreate(false)}
						disabled={isCreating}
						class="btn-pill btn-pill-primary mc-create-btn"
						aria-label="Create module"
					>
						{#if isCreating}
							<span class="mc-spinner"></span>
							<span>Creating...</span>
						{:else}
							<Sparkles class="w-4 h-4" />
							<span>Create Module</span>
						{/if}
					</button>
				{:else}
					<button
						onclick={handleNext}
						class="btn-pill btn-pill-primary"
						class:mc-btn--disabled={!canProceed}
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
	/*  MODULE CREATE v3 (mc-) — Carousel + Micro-interactions      */
	/* ══════════════════════════════════════════════════════════════ */
	.mc-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg, #fff);
	}

	/* ── Header ─────────────────────────────────────────────────── */
	.mc-header {
		flex-shrink: 0;
		padding: 20px 32px 16px;
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
	}
	.mc-back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: var(--dt3, #888);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		margin-bottom: 12px;
		transition: color 0.15s;
	}
	.mc-back:hover {
		color: var(--dt, #111);
	}
	.mc-header__title-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
	}
	.mc-title {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111);
		margin-bottom: 2px;
	}
	.mc-subtitle {
		font-size: 13px;
		color: var(--dt3, #888);
		transition: opacity 0.2s;
	}
	.mc-step-badge {
		flex-shrink: 0;
		padding: 4px 12px;
		border-radius: 999px;
		background: var(--dbg3, #eee);
		color: var(--dt2, #555);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
	}

	/* ── Step Indicator ─────────────────────────────────────────── */
	.mc-steps {
		margin-top: 20px;
	}
	.mc-steps__track {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 32px;
	}
	.mc-steps__line-bg {
		position: absolute;
		top: 50%;
		left: 16px;
		right: 16px;
		height: 2px;
		background: var(--dbd, #e0e0e0);
		transform: translateY(-50%);
		border-radius: 1px;
	}
	.mc-steps__line-fill {
		position: absolute;
		top: 50%;
		left: 16px;
		height: 2px;
		background: var(--bos-status-success);
		transform: translateY(-50%);
		border-radius: 1px;
		transition: width 0.5s cubic-bezier(0.16, 1, 0.3, 1);
		max-width: calc(100% - 32px);
	}
	.mc-step__dot {
		position: relative;
		z-index: 1;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
		background: var(--dbg, #fff);
		border: 2px solid var(--dbd, #e0e0e0);
		color: var(--dt3, #888);
		cursor: default;
		transition: all 0.35s cubic-bezier(0.16, 1, 0.3, 1);
	}
	.mc-step__dot--active {
		background: var(--dt, #111);
		border-color: var(--dt, #111);
		color: var(--bos-surface-on-color);
		transform: scale(1.1);
		box-shadow: 0 0 0 4px var(--bos-v2-layer-insideBorder-border);
	}
	.mc-step__dot--done {
		background: var(--bos-status-success);
		border-color: var(--bos-status-success);
		color: var(--bos-surface-on-color);
	}
	.mc-step__dot--clickable {
		cursor: pointer;
	}
	.mc-step__dot--clickable:hover:not(.mc-step__dot--active) {
		border-color: var(--dt3, #888);
		transform: scale(1.05);
	}
	.mc-steps__labels {
		display: flex;
		justify-content: space-between;
		margin-top: 6px;
		padding: 0 2px;
	}
	.mc-step__label {
		font-size: 11px;
		font-weight: 500;
		color: var(--dt4, #bbb);
		text-align: center;
		flex: 1;
		transition: color 0.2s, font-weight 0.2s;
	}
	.mc-step__label--active {
		color: var(--dt, #111);
		font-weight: 600;
	}
	.mc-step__label--done {
		color: var(--bos-status-success);
	}

	/* ── Carousel ───────────────────────────────────────────────── */
	.mc-content {
		flex: 1;
		overflow: hidden;
		position: relative;
	}
	.mc-content--shake {
		animation: mc-shake 0.4s ease-out;
	}
	@keyframes mc-shake {
		0%, 100% { transform: translateX(0); }
		20% { transform: translateX(-8px); }
		40% { transform: translateX(6px); }
		60% { transform: translateX(-4px); }
		80% { transform: translateX(2px); }
	}
	.mc-carousel {
		display: flex;
		height: 100%;
		transition: transform 0.5s cubic-bezier(0.16, 1, 0.3, 1);
		will-change: transform;
	}
	.mc-carousel__slide {
		min-width: 100%;
		width: 100%;
		height: 100%;
		overflow-y: auto;
		flex-shrink: 0;
	}
	.mc-slide-inner {
		max-width: 800px;
		margin: 0 auto;
		padding: 28px 32px 48px;
	}

	/* ── Section ────────────────────────────────────────────────── */
	.mc-section {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.mc-section__header {
		margin-bottom: 4px;
	}
	.mc-section__title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111);
		margin-bottom: 4px;
	}
	.mc-section__desc {
		font-size: 13px;
		color: var(--dt3, #888);
		line-height: 1.5;
	}

	/* ── Field ──────────────────────────────────────────────────── */
	.mc-field {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.mc-label {
		font-size: 13px;
		font-weight: 500;
		color: var(--dt2, #555);
	}
	.mc-label-row {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.mc-textarea {
		width: 100%;
		padding: 10px 14px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		font-family: monospace;
		font-size: 13px;
		resize: vertical;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.mc-textarea::placeholder {
		color: var(--dt4, #bbb);
	}
	.mc-textarea:focus {
		outline: none;
		border-color: var(--dt3, #888);
		box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.04);
	}

	/* ── Radio Group ────────────────────────────────────────────── */
	.mc-radio-group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.mc-radio {
		display: flex;
		align-items: center;
		padding: 12px 14px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		cursor: pointer;
		transition: border-color 0.2s, background 0.2s, transform 0.15s;
	}
	.mc-radio:hover {
		border-color: var(--dt3, #888);
	}
	.mc-radio:active {
		transform: scale(0.99);
	}
	.mc-radio--active {
		border-color: var(--dt, #111);
		background: var(--dbg2, #fafafa);
	}
	.mc-radio__input {
		margin-right: 12px;
		accent-color: var(--dt, #111);
	}
	.mc-radio__title {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.mc-radio__desc {
		font-size: 12px;
		color: var(--dt3, #888);
		margin-top: 1px;
	}

	/* ── Footer ─────────────────────────────────────────────────── */
	.mc-footer {
		flex-shrink: 0;
		padding: 12px 32px;
		border-top: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
	}
	.mc-footer__inner {
		max-width: 800px;
		margin: 0 auto;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.mc-footer__actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* ── Button States ──────────────────────────────────────────── */
	.mc-btn--disabled {
		opacity: 0.5;
		pointer-events: none;
	}
	.mc-create-btn {
		position: relative;
		overflow: hidden;
	}
	.mc-create-btn::after {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, transparent, var(--bos-white-10), transparent);
		transform: translateX(-100%);
		animation: mc-shimmer 2.5s ease-in-out infinite;
		pointer-events: none;
	}
	@keyframes mc-shimmer {
		0% { transform: translateX(-100%); }
		50%, 100% { transform: translateX(100%); }
	}
	.mc-spinner {
		width: 14px;
		height: 14px;
		border: 2px solid var(--bos-white-30);
		border-top-color: var(--bos-surface-on-color);
		border-radius: 50%;
		animation: mc-spin 0.6s linear infinite;
	}
	@keyframes mc-spin {
		to { transform: rotate(360deg); }
	}
</style>
