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

		await templateStore.generateApp(templateId, workspaceId, appName, config);
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

<div class="flex flex-col h-full bg-white">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<LoadingSpinner size="lg" message="Loading template..." fullscreen />
		</div>
	{:else if error && !template}
		<div class="flex-1 flex items-center justify-center p-6">
			<div class="text-center">
				<p class="text-red-600 mb-4">{error}</p>
				<button
					onclick={() => goto('/templates')}
					class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800"
				>
					Back to Templates
				</button>
			</div>
		</div>
	{:else if template}
		<!-- Header -->
		<div class="px-6 py-4 border-b border-gray-200">
			<button
				onclick={() => goto('/templates')}
				class="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 mb-4"
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
							<div class="w-32 h-32 rounded-xl overflow-hidden bg-gray-100 flex-shrink-0">
								<img
									src={template.preview_image_url}
									alt={template.name}
									class="w-full h-full object-cover"
								/>
							</div>
						{:else if template.icon_url}
							<div
								class="w-32 h-32 rounded-xl bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center flex-shrink-0"
							>
								<img src={template.icon_url} alt={template.name} class="w-20 h-20" />
							</div>
						{:else}
							<div
								class="w-32 h-32 rounded-xl bg-gradient-to-br from-blue-100 to-purple-100 flex items-center justify-center text-5xl flex-shrink-0"
							>
								{template.name.charAt(0).toUpperCase()}
							</div>
						{/if}

						<!-- Info -->
						<div class="flex-1">
							<div class="flex items-start justify-between mb-3">
								<div>
									<h1 class="text-3xl font-bold text-gray-900 mb-2">{template.name}</h1>
									<div class="flex items-center gap-3">
										<span
											class="inline-flex items-center px-3 py-1 text-sm font-medium rounded-lg border {categoryColors[
												template.category
											]}"
										>
											{categoryLabels[template.category]}
										</span>
										{#if builtInInfo}
											<span class="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-700 border border-gray-200">
												<FileCode class="w-3 h-3 mr-1" />
												{builtInInfo.stack_type}
											</span>
											<span class="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-700 border border-gray-200">
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

							<p class="text-gray-600 mb-4">{template.description}</p>

							<!-- Stats -->
							<div class="flex items-center gap-6 text-sm text-gray-500">
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
					<div class="mb-8 p-6 bg-green-50 border border-green-200 rounded-xl">
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center flex-shrink-0">
								<Check class="w-5 h-5 text-green-600" />
							</div>
							<div class="flex-1">
								<h3 class="text-lg font-semibold text-green-900 mb-1">App Generated Successfully!</h3>
								<p class="text-sm text-green-700 mb-3">
									Your app "{generationResult.app_name}" has been created with {generationResult.total_files} files from the {generationResult.template_name} template.
								</p>
								<div class="flex items-center gap-3">
									<button
										onclick={handleViewApp}
										class="btn-pill btn-pill-success btn-pill-sm"
									>
										View Generated App
									</button>
									<span class="text-xs text-green-600">Version {generationResult.version_number}</span>
								</div>

								<!-- File List -->
								<div class="mt-4 border-t border-green-200 pt-4">
									<h4 class="text-sm font-medium text-green-800 mb-2">Generated Files:</h4>
									<div class="space-y-1">
										{#each generationResult.files as file}
											<div class="flex items-center justify-between text-xs text-green-700 bg-green-100 px-3 py-1.5 rounded">
												<span class="font-mono">{file.path}</span>
												<span class="text-green-500">{(file.size / 1024).toFixed(1)} KB</span>
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
					<div class="mb-8 p-6 bg-gray-50 border border-gray-200 rounded-xl">
						<h2 class="text-xl font-semibold text-gray-900 mb-4">Configure Your App</h2>

						{#if error}
							<div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
								{error}
							</div>
						{/if}

						<div class="space-y-4">
							<!-- Workspace ID -->
							<div>
								<label for="workspace-id" class="block text-sm font-medium text-gray-700 mb-1">
									Workspace ID <span class="text-red-500">*</span>
								</label>
								<input
									id="workspace-id"
									type="text"
									bind:value={workspaceId}
									placeholder="Enter your workspace ID"
									class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-gray-900 focus:border-transparent"
								/>
							</div>

							<!-- App Name -->
							<div>
								<label for="app-name" class="block text-sm font-medium text-gray-700 mb-1">
									App Name <span class="text-red-500">*</span>
								</label>
								<input
									id="app-name"
									type="text"
									bind:value={appName}
									placeholder="My Awesome App"
									class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-gray-900 focus:border-transparent"
								/>
							</div>

							<!-- Dynamic Config Fields from Schema -->
							{#if builtInInfo}
								{#each Object.entries(builtInInfo.config_schema) as [key, field]}
									{#if key !== 'app_name'}
										<div>
											<label for="config-{key}" class="block text-sm font-medium text-gray-700 mb-1">
												{field.label}
												{#if field.required}
													<span class="text-red-500">*</span>
												{/if}
											</label>
											{#if field.description}
												<p class="text-xs text-gray-500 mb-1">{field.description}</p>
											{/if}

											{#if field.type === 'select' && field.options}
												<select
													id="config-{key}"
													bind:value={configValues[key]}
													class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-gray-900 focus:border-transparent"
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
														class="w-4 h-4 text-gray-900 border-gray-300 rounded focus:ring-gray-900"
													/>
													<span class="text-sm text-gray-700">Enabled</span>
												</label>
											{:else}
												<input
													id="config-{key}"
													type="text"
													bind:value={configValues[key]}
													placeholder={field.default}
													class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-gray-900 focus:border-transparent"
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
								class="px-4 py-2 text-sm text-gray-600 hover:text-gray-900"
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
						<h2 class="text-xl font-semibold text-gray-900 mb-4">Features Included</h2>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							{#each template.features as feature}
								<div class="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
									<Check class="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
									<span class="text-sm text-gray-700">{feature}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Target Audience -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
					{#if template.business_types && template.business_types.length > 0}
						<div>
							<h2 class="text-lg font-semibold text-gray-900 mb-3">Best For</h2>
							<div class="flex flex-wrap gap-2">
								{#each template.business_types as type}
									<span class="px-3 py-1.5 bg-blue-50 text-blue-700 text-sm rounded-lg">
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
							<h2 class="text-lg font-semibold text-gray-900 mb-3">Team Size</h2>
							<div class="flex flex-wrap gap-2">
								{#each template.team_sizes as size}
									<span class="px-3 py-1.5 bg-purple-50 text-purple-700 text-sm rounded-lg">
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
				<div class="border-t border-gray-200 pt-6">
					<div class="text-xs text-gray-500">
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
