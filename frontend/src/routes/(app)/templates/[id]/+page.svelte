<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Sparkles, Star, Check, Loader2, FileCode, Package } from 'lucide-svelte';
	import { templateStore, categoryLabels, categoryColors } from '$lib/stores/templateStore';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import AgentProgressPanel from '$lib/components/osa/AgentProgressPanel.svelte';
	import type { AppTemplate, BuiltInTemplateInfo, GenerationResult } from '$lib/api/templates';

	const templateId = $derived($page.params.id);

	let template = $state<AppTemplate | null>(null);
	let builtInInfo = $state<BuiltInTemplateInfo | null>(null);
	let loading = $state(true);
	let generating = $state(false);
	let error = $state<string | null>(null);
	let generationResult = $state<GenerationResult | null>(null);
	let queueItemId = $state<string | null>(null);

	// Config form state
	let showConfigForm = $state(false);
	let appName = $state('');
	let workspaceId = $state('');
	let configValues = $state<Record<string, string>>({});

	// Subscribe to store
	$effect(() => {
		const unsubscribe = templateStore.subscribe((state) => {
			template = state.currentTemplate;
			loading = state.loading;
			error = state.error;
			generating = state.generating;
			generationResult = state.generationResult;
			queueItemId = state.queueItemId;

			// Find matching built-in template info
			if (template && state.builtInTemplates.length > 0) {
				builtInInfo = state.builtInTemplates.find(
					(bt) => bt.name === template?.name || bt.id === template?.id
				) || null;

				// Initialize config values from defaults
				if (builtInInfo && Object.keys(configValues).length === 0) {
					const defaults: Record<string, string> = {};
					for (const [key, field] of Object.entries(builtInInfo.config_schema)) {
						defaults[key] = field.default;
					}
					configValues = defaults;
					appName = defaults['app_name'] || template?.name || 'My App';
				}
			}
		});
		return unsubscribe;
	});

	// Load template on mount
	onMount(() => {
		if (templateId) {
			templateStore.loadTemplate(templateId);
			templateStore.loadBuiltInTemplates();
		}

		return () => {
			templateStore.clearGenerationResult();
		};
	});

	function handleStartGeneration() {
		showConfigForm = true;
	}

	async function handleGenerate() {
		if (!template || !workspaceId || !appName) return;

		const config: Record<string, string | number | boolean> = {};
		for (const [key, value] of Object.entries(configValues)) {
			if (key !== 'app_name') {
				config[key] = value;
			}
		}

		await templateStore.generateApp(templateId ?? '', workspaceId, appName, config);
	}

	function formatPopularity(score: number): string {
		if (score >= 1000) return `${(score / 1000).toFixed(1)}k users`;
		return `${score} users`;
	}

	function handleGenerationComplete() {
		templateStore.clearGenerationResult();
		// Navigate to the generated apps list on completion
		goto('/generated-apps');
	}

	function handleGenerationError(errorMsg: string) {
		templateStore.clearGenerationResult();
	}

	function handleViewApp() {
		if (generationResult) {
			goto(`/generated-apps/${generationResult.app_id}`);
		}
	}
</script>

<div class="flex flex-col h-full" style="background: var(--dbg)">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<LoadingSpinner size="lg" message="Loading template..." fullscreen />
		</div>
	{:else if error && !template}
		<div class="flex-1 flex items-center justify-center p-6">
			<div class="text-center">
				<p class="mb-4" style="color: var(--bos-status-error)">{error}</p>
				<button
					onclick={() => goto('/templates')}
					class="px-4 py-2 rounded-lg"
					style="background: var(--dt); color: var(--dbg)"
				>
					Back to Templates
				</button>
			</div>
		</div>
	{:else if template}
		<!-- Header -->
		<div class="px-6 py-4 border-b" style="border-color: var(--dbd)">
			<button
				onclick={() => goto('/templates')}
				class="flex items-center gap-2 text-sm mb-4"
				style="color: var(--dt2)"
			>
				<ArrowLeft class="w-4 h-4" />
				<span>Back to Templates</span>
			</button>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-auto">
			<div class="max-w-4xl mx-auto p-6">
				<!-- Header Section -->
				<div class="mb-8">
					<div class="flex items-start gap-6">
						<!-- Icon/Image -->
						{#if template.preview_image_url}
							<div class="w-32 h-32 rounded-xl overflow-hidden flex-shrink-0" style="background: var(--dbg3)">
								<img
									src={template.preview_image_url}
									alt={template.name}
									class="w-full h-full object-cover"
								/>
							</div>
						{:else if template.icon_url}
							<div
								class="w-32 h-32 rounded-xl flex items-center justify-center flex-shrink-0"
								style="background: var(--dbg3)"
							>
								<img src={template.icon_url} alt={template.name} class="w-20 h-20" />
							</div>
						{:else}
							<div
								class="w-32 h-32 rounded-xl flex items-center justify-center text-5xl flex-shrink-0"
								style="background: var(--dbg3); color: var(--dt)"
							>
								{template.name.charAt(0).toUpperCase()}
							</div>
						{/if}

						<!-- Info -->
						<div class="flex-1">
							<div class="flex items-start justify-between mb-3">
								<div>
									<h1 class="text-3xl font-bold mb-2" style="color: var(--dt)">{template.name}</h1>
									<div class="flex items-center gap-3">
										<span
											class="inline-flex items-center px-3 py-1 text-sm font-medium rounded-lg border {categoryColors[
												template.category
											]}"
										>
											{categoryLabels[template.category]}
										</span>
										{#if builtInInfo}
											<span class="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-lg border" style="background: var(--dbg3); color: var(--dt2); border-color: var(--dbd)">
												<FileCode class="w-3 h-3 mr-1" />
												{builtInInfo.stack_type}
											</span>
											<span class="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-lg border" style="background: var(--dbg3); color: var(--dt2); border-color: var(--dbd)">
												<Package class="w-3 h-3 mr-1" />
												{builtInInfo.file_count} files
											</span>
										{/if}
										{#if template.is_premium}
											<div
												class="flex items-center gap-1 px-2 py-1 bg-gradient-to-r from-amber-400 to-orange-500 text-white text-xs font-semibold rounded-full"
											>
												<Sparkles class="w-3 h-3" />
												<span>Premium</span>
											</div>
										{/if}
									</div>
								</div>
							</div>

							<p class="mb-4" style="color: var(--dt2)">{template.description}</p>

							<!-- Stats -->
							<div class="flex items-center gap-6 text-sm" style="color: var(--dt3)">
								<div class="flex items-center gap-1.5">
									<Star class="w-4 h-4" />
									<span>{formatPopularity(template.popularity_score)}</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- SSE Progress Panel (async queue-based generation) -->
				{#if queueItemId && generating}
					<div class="mb-8">
						<AgentProgressPanel
							queueItemId={queueItemId}
							onComplete={handleGenerationComplete}
							onError={handleGenerationError}
						/>
					</div>
				{/if}

				<!-- Generation Success -->
				{#if generationResult}
					<div class="mb-8 p-6 rounded-xl border" style="background: var(--bos-status-success-bg); border-color: color-mix(in srgb, var(--bos-status-success) 25%, transparent)">
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" style="background: color-mix(in srgb, var(--bos-status-success) 15%, transparent); color: var(--bos-status-success)">
								<Check class="w-5 h-5" />
							</div>
							<div class="flex-1">
								<h3 class="text-lg font-semibold mb-1" style="color: var(--bos-status-success)">App Generated Successfully!</h3>
								<p class="text-sm mb-3" style="color: var(--bos-status-success)">
									Your app "{generationResult.app_name}" has been created with {generationResult.total_files} files from the {generationResult.template_name} template.
								</p>
								<div class="flex items-center gap-3">
									<button
										onclick={handleViewApp}
										class="btn-pill btn-pill-success btn-pill-sm"
									>
										View Generated App
									</button>
									<span class="text-xs" style="color: var(--bos-status-success)">Version {generationResult.version_number}</span>
								</div>

								<!-- File List -->
								<div class="mt-4 pt-4" style="border-top: 1px solid color-mix(in srgb, var(--bos-status-success) 25%, transparent)">
									<h4 class="text-sm font-medium mb-2" style="color: var(--bos-status-success)">Generated Files:</h4>
									<div class="space-y-1">
										{#each generationResult.files as file}
											<div class="flex items-center justify-between text-xs px-3 py-1.5 rounded" style="background: color-mix(in srgb, var(--bos-status-success) 10%, transparent); color: var(--bos-status-success)">
												<span class="font-mono">{file.path}</span>
												<span>{(file.size / 1024).toFixed(1)} KB</span>
											</div>
										{/each}
									</div>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Configuration Form -->
				{#if showConfigForm && !generationResult && !queueItemId}
					<div class="mb-8 p-6 rounded-xl border" style="background: var(--dbg2); border-color: var(--dbd)">
						<h2 class="text-xl font-semibold mb-4" style="color: var(--dt)">Configure Your App</h2>

						{#if error}
							<div class="mb-4 p-3 rounded-lg text-sm border" style="background: var(--bos-status-error-bg); border-color: color-mix(in srgb, var(--bos-status-error) 25%, transparent); color: var(--bos-status-error)">
								{error}
							</div>
						{/if}

						<div class="space-y-4">
							<!-- Workspace ID -->
							<div>
								<label for="workspace-id" class="block text-sm font-medium mb-1" style="color: var(--dt2)">
									Workspace ID <span style="color: var(--bos-status-error)">*</span>
								</label>
								<input
									id="workspace-id"
									type="text"
									bind:value={workspaceId}
									placeholder="Enter your workspace ID"
									class="w-full px-4 py-2 rounded-lg"
									style="border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt)"
								/>
							</div>

							<!-- App Name -->
							<div>
								<label for="app-name" class="block text-sm font-medium mb-1" style="color: var(--dt2)">
									App Name <span style="color: var(--bos-status-error)">*</span>
								</label>
								<input
									id="app-name"
									type="text"
									bind:value={appName}
									placeholder="My Awesome App"
									class="w-full px-4 py-2 rounded-lg"
									style="border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt)"
								/>
							</div>

							<!-- Dynamic Config Fields from Schema -->
							{#if builtInInfo}
								{#each Object.entries(builtInInfo.config_schema) as [key, field]}
									{#if key !== 'app_name'}
										<div>
											<label for="config-{key}" class="block text-sm font-medium mb-1" style="color: var(--dt2)">
												{field.label}
												{#if field.required}
													<span style="color: var(--bos-status-error)">*</span>
												{/if}
											</label>
											{#if field.description}
												<p class="text-xs mb-1" style="color: var(--dt3)">{field.description}</p>
											{/if}

											{#if field.type === 'select' && field.options}
												<select
													id="config-{key}"
													bind:value={configValues[key]}
													class="w-full px-4 py-2 rounded-lg"
													style="border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt)"
												>
													{#each field.options as option}
														<option value={option}>{option}</option>
													{/each}
												</select>
											{:else if field.type === 'boolean'}
												<label class="flex items-center gap-2 cursor-pointer">
													<input
														type="checkbox"
														checked={configValues[key] === 'true'}
														onchange={(e) => {
															configValues[key] = e.currentTarget.checked ? 'true' : 'false';
														}}
														class="w-4 h-4 rounded"
														style="border-color: var(--dbd); color: var(--dt)"
													/>
													<span class="text-sm" style="color: var(--dt2)">Enabled</span>
												</label>
											{:else}
												<input
													id="config-{key}"
													type="text"
													bind:value={configValues[key]}
													placeholder={field.default}
													class="w-full px-4 py-2 rounded-lg"
													style="border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt)"
												/>
											{/if}
										</div>
									{/if}
								{/each}
							{/if}
						</div>

						<!-- Generate Button -->
						<div class="mt-6 flex items-center gap-3">
							<button
								onclick={handleGenerate}
								disabled={generating || !workspaceId || !appName}
								class="btn-pill btn-pill-primary flex items-center gap-2"
							>
								{#if generating}
									<Loader2 class="w-4 h-4 animate-spin" />
									<span>Generating...</span>
								{:else}
									<Sparkles class="w-4 h-4" />
									<span>Generate App</span>
								{/if}
							</button>
							<button
								onclick={() => showConfigForm = false}
								class="px-4 py-2 text-sm"
								style="color: var(--dt2)"
							>
								Cancel
							</button>
						</div>
					</div>
				{:else if !generationResult && !queueItemId}
					<!-- Action Button -->
					<div class="mb-8">
						<button
							onclick={handleStartGeneration}
							class="btn-pill btn-pill-primary w-full sm:w-auto"
						>
							Use This Template
						</button>
					</div>
				{/if}

				<!-- Features Section -->
				{#if template.features && template.features.length > 0}
					<div class="mb-8">
						<h2 class="text-xl font-semibold mb-4" style="color: var(--dt)">Features Included</h2>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							{#each template.features as feature}
								<div class="flex items-start gap-3 p-3 rounded-lg" style="background: var(--dbg2)">
									<Check class="w-5 h-5 flex-shrink-0 mt-0.5" style="color: var(--bos-status-success)" />
									<span class="text-sm" style="color: var(--dt2)">{feature}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Target Audience -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
					{#if template.business_types && template.business_types.length > 0}
						<div>
							<h2 class="text-lg font-semibold mb-3" style="color: var(--dt)">Best For</h2>
							<div class="flex flex-wrap gap-2">
								{#each template.business_types as type}
									<span class="px-3 py-1.5 text-sm rounded-lg" style="background: var(--bos-status-info-bg); color: var(--bos-status-info)">
										{type
											.split('_')
											.map((w) => w.charAt(0).toUpperCase() + w.slice(1))
											.join(' ')}
									</span>
								{/each}
							</div>
						</div>
					{/if}

					{#if template.team_sizes && template.team_sizes.length > 0}
						<div>
							<h2 class="text-lg font-semibold mb-3" style="color: var(--dt)">Team Size</h2>
							<div class="flex flex-wrap gap-2">
								{#each template.team_sizes as size}
									<span class="px-3 py-1.5 text-sm rounded-lg" style="background: var(--dbg3); color: var(--dt2)">
										{size === 'solo'
											? 'Solo'
											: size === 'small'
												? 'Small (2-10)'
												: size === 'medium'
													? 'Medium (11-50)'
													: 'Large (51+)'}
									</span>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<!-- Additional Info -->
				<div class="pt-6" style="border-top: 1px solid var(--dbd)">
					<div class="text-xs" style="color: var(--dt3)">
						<p>
							Created: {new Date(template.created_at).toLocaleDateString()}
						</p>
						{#if template.updated_at !== template.created_at}
							<p class="mt-1">
								Last updated: {new Date(template.updated_at).toLocaleDateString()}
							</p>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
