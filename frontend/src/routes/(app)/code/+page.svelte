<script lang="ts">
	import { Code2, PanelLeftClose, PanelLeftOpen } from 'lucide-svelte';
	import CodeSidebar from '$lib/components/code/CodeSidebar.svelte';
	import CodeViewer from '$lib/components/code/CodeViewer.svelte';
	import { getWorkflowFiles, getFileContent } from '$lib/api/osa/files';
	import type { OSAFile } from '$lib/components/osa/types';

	// ── Page data from +page.ts load ──────────────────────────────────────────
	let { data } = $props();
	let modules = $derived(data.modules);

	// ── Sidebar & module state ────────────────────────────────────────────────
	let sidebarVisible = $state(true);
	let selectedModuleId = $state<string | null>(null);
	let loadingModuleId = $state<string | null>(null);

	// Map of moduleId → files fetched from API (undefined = not yet fetched)
	let moduleFiles = $state<Record<string, OSAFile[]>>({});

	// ── File viewer state ─────────────────────────────────────────────────────
	let selectedFile = $state<OSAFile | null>(null);
	let fileContent = $state('');
	let fileLoading = $state(false);
	let fileError = $state<string | null>(null);

	// ── Derived breadcrumb info ───────────────────────────────────────────────
	let selectedModuleName = $derived(
		modules.find((m: { id: string }) => m.id === selectedModuleId)?.title ?? ''
	);

	// ── Handlers ─────────────────────────────────────────────────────────────

	async function handleModuleSelect(moduleId: string) {
		// Already fetched — skip network call
		if (moduleFiles[moduleId] !== undefined) {
			selectedModuleId = moduleId;
			return;
		}

		selectedModuleId = moduleId;
		loadingModuleId = moduleId;

		try {
			// Generated-app modules use workflow-based file API.
			// Built-in BOS modules don't have a backend source API yet — show placeholder.
			const files = await getWorkflowFiles(moduleId).catch(() => null);

			moduleFiles = {
				...moduleFiles,
				[moduleId]: files ?? [],
			};
		} catch {
			moduleFiles = { ...moduleFiles, [moduleId]: [] };
		} finally {
			loadingModuleId = null;
		}
	}

	async function handleFileSelect(file: OSAFile) {
		if (selectedFile?.id === file.id) return;

		selectedFile = file;
		fileContent = '';
		fileError = null;

		// If file already has inline content, use it
		if (file.content) {
			fileContent = file.content;
			return;
		}

		fileLoading = true;
		try {
			const result = await getFileContent(file.id);
			fileContent = result.content;
		} catch (err) {
			fileError =
				err instanceof Error
					? err.message
					: 'Source not available for this module file.';
			fileContent = '';
		} finally {
			fileLoading = false;
		}
	}

	function toggleSidebar() {
		sidebarVisible = !sidebarVisible;
	}
</script>

<svelte:head>
	<title>Code Browser | BusinessOS</title>
</svelte:head>

<div class="h-full flex flex-col bg-gray-950 overflow-hidden">
	<!-- Top bar -->
	<header class="flex items-center gap-3 px-4 h-11 border-b border-white/8 flex-shrink-0 bg-gray-950">
		<!-- Sidebar toggle -->
		<button
			onclick={toggleSidebar}
			class="btn-pill btn-pill-ghost flex items-center justify-center w-7 h-7"
			aria-label={sidebarVisible ? 'Collapse sidebar' : 'Expand sidebar'}
			title={sidebarVisible ? 'Collapse sidebar' : 'Expand sidebar'}
		>
			{#if sidebarVisible}
				<PanelLeftClose size={16} />
			{:else}
				<PanelLeftOpen size={16} />
			{/if}
		</button>

		<!-- Title / breadcrumb -->
		<div class="flex items-center gap-2 text-sm">
			<Code2 class="text-gray-500 flex-shrink-0" size={15} aria-hidden="true" />
			<span class="text-gray-500">Code</span>
			{#if selectedModuleName}
				<span class="text-gray-700" aria-hidden="true">›</span>
				<span class="text-gray-300 font-medium">{selectedModuleName}</span>
			{/if}
			{#if selectedFile}
				<span class="text-gray-700" aria-hidden="true">›</span>
				<span class="text-gray-400 font-mono text-xs">{selectedFile.name}</span>
			{/if}
		</div>
	</header>

	<!-- Body: sidebar + viewer -->
	<div class="flex flex-1 min-h-0 overflow-hidden">
		<!-- Sidebar — hidden on mobile (md:flex), toggleable via button -->
		{#if sidebarVisible}
			<div class="hidden md:flex flex-shrink-0">
				<CodeSidebar
					{modules}
					{moduleFiles}
					{selectedModuleId}
					{selectedFile}
					{loadingModuleId}
					onModuleSelect={handleModuleSelect}
					onFileSelect={handleFileSelect}
				/>
			</div>
		{/if}

		<!-- Editor panel -->
		<main class="flex-1 min-w-0 h-full overflow-hidden">
			<CodeViewer
				file={selectedFile}
				content={fileContent}
				loading={fileLoading}
				error={fileError}
				moduleName={selectedModuleName}
			/>
		</main>
	</div>
</div>
