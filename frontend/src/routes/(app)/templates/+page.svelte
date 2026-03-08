<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { templateStore } from '$lib/stores/templateStore';
	import { TemplateGallery } from '$lib/components/templates';
	import type { AppTemplate, TemplateCategory, BusinessType, TeamSize } from '$lib/api/templates';

	// State from store
	let templates = $state<AppTemplate[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Subscribe to store
	$effect(() => {
		const unsubscribe = templateStore.subscribe((state) => {
			templates = state.templates;
			loading = state.loading;
			error = state.error;
		});
		return unsubscribe;
	});

	// Load templates on mount
	onMount(() => {
		templateStore.loadTemplates();
	});

	function handleFilterChange(filters: {
		category?: TemplateCategory | null;
		business_type?: BusinessType | null;
		team_size?: TeamSize | null;
		search?: string;
		sort?: 'popular' | 'newest' | 'name';
	}) {
		templateStore.setFilters(filters);
		templateStore.loadTemplates();
	}

	async function handleUseTemplate(templateId: string) {
		// Navigate to template detail page where generation form is shown
		goto(`/templates/${templateId}`);
	}

	function handlePreviewTemplate(templateId: string) {
		// Navigate to template detail page
		goto(`/templates/${templateId}`);
	}
</script>

<div class="flex flex-col h-full bg-white">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">App Templates</h1>
			<p class="text-sm text-gray-500 mt-0.5">
				Choose from pre-built templates to quickly set up your workspace
			</p>
		</div>
	</div>

	<!-- Gallery -->
	<TemplateGallery
		{templates}
		{loading}
		{error}
		onFilterChange={handleFilterChange}
		onUseTemplate={handleUseTemplate}
		onPreviewTemplate={handlePreviewTemplate}
	/>
</div>
