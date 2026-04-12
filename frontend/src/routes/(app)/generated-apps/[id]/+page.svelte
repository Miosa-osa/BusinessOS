<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		Monitor,
		Maximize2,
		Minimize2,
		Code2,
		Eye,
		Search,
		ExternalLink,
		RefreshCw,
		AlertTriangle,
		History,
		Terminal
	} from 'lucide-svelte';
	import { generatedAppsStore } from '$lib/stores/generatedAppsStore';
	import BuildProgress from '$lib/components/osa/BuildProgress.svelte';
	import AppActions from '$lib/components/osa/AppActions.svelte';
	import { Package, Download as DownloadIcon } from 'lucide-svelte';
	import SandboxStatusBadge from '$lib/components/osa/SandboxStatusBadge.svelte';
	import OpenSandboxButton from '$lib/components/osa/OpenSandboxButton.svelte';
	import FileTree from '$lib/components/osa/FileTree.svelte';
	import MonacoEditor from '$lib/editor/MonacoEditor.svelte';
	import EditorToolbar from '$lib/editor/EditorToolbar.svelte';
	import EditorStatusBar from '$lib/editor/EditorStatusBar.svelte';
	import {
		VersionBadge,
		VersionDropdown,
		VersionTimelinePanel,
		SaveVersionModal,
		VersionPreviewModal,
		RestoreConfirmDialog,
		VersionDiffModal
	} from '$lib/components/versioning';
	import { extractDisplayNumber } from '$lib/api/versions/mappers';
	import type { VersionSummary } from '$lib/types/versions';

	// ── Stores ─────────────────────────────────────────────────────────────────
	import { appStore as a } from '$lib/stores/generated-apps/appStore.svelte';
	import { editorStore as ed } from '$lib/stores/generated-apps/editorStore.svelte';
	import { sandboxStore as sb } from '$lib/stores/generated-apps/sandboxStore.svelte';
	import { versionStore as vs } from '$lib/stores/generated-apps/versionStore.svelte';
	import { moduleStore as ms } from '$lib/stores/generated-apps/moduleStore.svelte';

	// ── Derived from route ─────────────────────────────────────────────────────
	let appId = $derived($page.params.id as string);

	// ── DOM ref (ephemeral — belongs in component, not store) ─────────────────
	let editorRef = $state<MonacoEditor | undefined>(undefined);

	// ── Sandbox app-refresh callback ───────────────────────────────────────────
	function onAppRefresh(refreshed: typeof a.app) {
		a.app = refreshed;
	}

	// ── Lifecycle ──────────────────────────────────────────────────────────────
	onMount(async () => {
		a.loading = true;
		try {
			a.app = await generatedAppsStore.getAppById(appId);
			if (!a.app) {
				a.error = 'App not found';
				return;
			}

			if (a.app.status === 'generating') {
				generatedAppsStore.subscribeToAppProgress(appId);
			}

			if (a.app.status === 'generated' || a.app.status === 'deployed') {
				await Promise.all([
					ed.loadFiles(appId),
					vs.loadVersions(a.app.workspace_id, appId),
				]);
			}

			sb.beginPollingIfDeploying(a.app, onAppRefresh);
		} catch (err) {
			a.error = err instanceof Error ? err.message : 'Failed to load app';
		} finally {
			a.loading = false;
		}
	});

	onDestroy(() => {
		if (a.app?.status === 'generating') {
			generatedAppsStore.unsubscribeFromAppProgress(appId);
		}
		sb.destroy();
	});

	// ── App-level handlers ─────────────────────────────────────────────────────
	function handleBack() {
		goto('/generated-apps');
	}

	async function handleDeploy() {
		if (!a.app) return;
		try {
			await generatedAppsStore.deployApp(a.app.id);
			a.app = await generatedAppsStore.getAppById(appId);
		} catch (err) {
			a.error = err instanceof Error ? err.message : 'Failed to deploy app';
		}
	}

	async function handleDelete() {
		if (!a.app) return;
		try {
			await generatedAppsStore.deleteApp(a.app.id);
			goto('/generated-apps');
		} catch (err) {
			a.error = err instanceof Error ? err.message : 'Failed to delete app';
		}
	}

	// ── Editor handlers ────────────────────────────────────────────────────────
	function handleEditorChange(_value: string) {
		const editor = editorRef?.getEditor?.();
		if (editor) {
			ed.updateCursorFromEditor(editor);
		}
	}

	// ── Version handlers ───────────────────────────────────────────────────────
	function handleVersionSelect(summary: VersionSummary) {
		vs.selectVersionSummary(summary.id);
	}

	async function handleVersionReloads() {
		if (!a.app) return;
		await Promise.all([
			vs.loadVersions(a.app.workspace_id, appId),
			ed.loadFiles(appId),
		]);
	}
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
</svelte:head>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center gap-4 mb-4">
				<button
					onclick={handleBack}
					class="btn-pill btn-pill-ghost btn-pill-icon"
					aria-label="Back to apps"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 19l-7-7 7-7"
						/>
					</svg>
				</button>
				<div class="flex-1">
					<div class="flex items-center gap-3">
						<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
							{a.app?.app_name || 'Loading...'}
						</h1>
						{#if vs.versions.length > 0}
							<VersionBadge
								currentVersion={vs.currentVersion}
								onclick={() => { vs.timelinePanelOpen = true; }}
							/>
						{/if}
					</div>
					{#if a.app}
						<p class="text-sm text-gray-600 dark:text-gray-400 mt-1">{a.app.description}</p>
					{/if}
				</div>
				{#if vs.versions.length > 0}
					<VersionDropdown
						currentVersion={vs.currentVersion}
						versions={vs.versionSummaries}
						onVersionSelect={handleVersionSelect}
						onViewAll={() => { vs.timelinePanelOpen = true; }}
						onSaveVersion={() => { vs.saveModalOpen = true; }}
					/>
				{/if}
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-auto">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			{#if a.error}
				<!-- Error State -->
				<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-6">
					<div class="flex items-start gap-3">
						<svg
							class="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<div>
							<h3 class="text-lg font-semibold text-red-900 dark:text-red-200">Error Loading App</h3>
							<p class="text-sm text-red-700 dark:text-red-300 mt-1">{a.error}</p>
							<button
								onclick={handleBack}
								class="btn-pill btn-pill-danger btn-pill-sm mt-3"
							>
								Back to Apps
							</button>
						</div>
					</div>
				</div>
			{:else if a.loading}
				<!-- Loading State -->
				<div class="flex items-center justify-center py-12">
					<div class="flex flex-col items-center gap-4">
						<svg
							class="w-12 h-12 text-blue-500 animate-spin"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						<p class="text-gray-600 dark:text-gray-400">Loading app details...</p>
					</div>
				</div>
			{:else if a.app}
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Main Content -->
					<div class="lg:col-span-2 space-y-6">
						<!-- Build Progress (if generating) -->
						{#if a.app.status === 'generating'}
							<BuildProgress
								buildId={a.app.id}
								onComplete={(result) => {
									if (import.meta.env.DEV) console.log('Build complete:', result);
									generatedAppsStore.fetchApps();
								}}
								onError={(err) => {
									console.error('Build error:', err);
									a.error = err.message;
								}}
							/>
						{/if}

						<!-- App Info Card -->
						<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
							<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">App Information</h2>
							<dl class="space-y-3">
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
									<dd class="mt-1 text-sm text-gray-900 dark:text-white capitalize">{a.app.status}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created</dt>
									<dd class="mt-1 text-sm text-gray-900 dark:text-white">
										{new Date(a.app.generated_at).toLocaleString()}
									</dd>
								</div>
								{#if a.app.deployed_at}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Deployed</dt>
										<dd class="mt-1 text-sm text-gray-900 dark:text-white">
											{new Date(a.app.deployed_at).toLocaleString()}
										</dd>
									</div>
								{/if}
								{#if a.app.custom_config?.category}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Category</dt>
										<dd class="mt-1 text-sm text-gray-900 dark:text-white">
											{a.app.custom_config.category}
										</dd>
									</div>
								{/if}
								{#if a.app.custom_config?.keywords && a.app.custom_config.keywords.length > 0}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Keywords</dt>
										<dd class="mt-1 flex flex-wrap gap-2">
											{#each a.app.custom_config.keywords as keyword}
												<span
													class="px-2 py-1 text-xs rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300"
												>
													{keyword}
												</span>
											{/each}
										</dd>
									</div>
								{/if}
							</dl>
						</div>

						<!-- Tab Switcher (Preview / Code / Terminal) -->
						{#if a.app.status === 'generated' || a.app.status === 'deployed'}
							<div class="flex items-center gap-1 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-1">
								<button
									class="flex items-center gap-2 btn-pill btn-pill-sm {ed.activeTab === 'preview' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
									onclick={() => { ed.activeTab = 'preview'; }}
								>
									<Eye class="w-4 h-4" />
									Preview
								</button>
								<button
									class="flex items-center gap-2 btn-pill btn-pill-sm {ed.activeTab === 'code' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
									onclick={() => { ed.activeTab = 'code'; }}
								>
									<Code2 class="w-4 h-4" />
									Code
									{#if ed.files.length > 0}
										<span class="text-xs bg-gray-200 dark:bg-gray-600 px-1.5 py-0.5 rounded-full">{ed.files.length}</span>
									{/if}
								</button>
								<button
									class="flex items-center gap-2 btn-pill btn-pill-sm {ed.activeTab === 'terminal' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
									onclick={() => { ed.activeTab = 'terminal'; }}
								>
									<Terminal class="w-4 h-4" />
									Terminal
								</button>
							</div>

							<!-- Sandbox Preview Tab -->
							{#if ed.activeTab === 'preview'}
								<div
									class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden"
									class:fixed={a.previewExpanded}
									class:inset-4={a.previewExpanded}
									class:z-50={a.previewExpanded}
								>
									<div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
										<div class="flex items-center gap-3">
											<Monitor class="w-5 h-5 text-gray-500" />
											<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Live Preview</h2>
											{#if a.app.sandbox}
												<SandboxStatusBadge status={a.app.sandbox.status} size="sm" />
											{/if}
										</div>
										<div class="flex items-center gap-2">
											{#if a.app.sandbox?.status === 'running' && a.app.sandbox.url}
												<button
													onclick={() => {
														if (a.app?.sandbox?.url) navigator.clipboard.writeText(a.app.sandbox.url);
													}}
													class="btn-pill btn-pill-soft btn-pill-xs font-mono flex items-center gap-1.5 max-w-[240px]"
													title="Copy sandbox URL"
												>
													<span class="health-dot running"></span>
													<span class="truncate">{a.app.sandbox.url}</span>
													<ExternalLink class="w-3 h-3 flex-shrink-0" />
												</button>
											{/if}
											<OpenSandboxButton
												sandbox={a.app.sandbox}
												appId={a.app.id}
												variant="secondary"
												onStart={() => sb.start(a.app!, onAppRefresh)}
												onStop={async () => { sb.showStopConfirm = true; }}
											/>
											{#if a.app.sandbox?.status === 'running'}
												<button
													onclick={() => a.togglePreviewExpanded()}
													class="btn-pill btn-pill-ghost btn-pill-icon"
													title={a.previewExpanded ? 'Exit fullscreen' : 'Fullscreen'}
												>
													{#if a.previewExpanded}
														<Minimize2 class="w-5 h-5" />
													{:else}
														<Maximize2 class="w-5 h-5" />
													{/if}
												</button>
											{/if}
										</div>
									</div>

									<!-- Deployment Progress Bar -->
									{#if (a.app.sandbox?.status === 'deploying' || a.app.sandbox?.status === 'pending') && sb.sandboxDeployProgress > 0}
										<div class="h-1 bg-gray-200 dark:bg-gray-700">
											<div
												class="h-full bg-blue-500 transition-all duration-500 ease-out"
												style="width: {sb.sandboxDeployProgress}%"
											></div>
										</div>
									{/if}

									<!-- Sandbox Error Banner -->
									{#if sb.sandboxError}
										<div class="flex items-center gap-3 px-4 py-3 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
											<AlertTriangle class="w-4 h-4 text-red-500 flex-shrink-0" />
											<span class="text-sm text-red-700 dark:text-red-300 flex-1">{sb.sandboxError}</span>
											<button
												onclick={() => sb.retry(a.app!, onAppRefresh)}
												class="flex items-center gap-1.5 btn-pill btn-pill-danger btn-pill-xs"
											>
												<RefreshCw class="w-3 h-3" />
												Retry
											</button>
										</div>
									{/if}

									<!-- Preview Content -->
									<div class="bg-gray-100 dark:bg-gray-900" class:h-96={!a.previewExpanded} class:flex-1={a.previewExpanded} style={a.previewExpanded ? 'height: calc(100% - 65px)' : ''}>
										{#if a.app.sandbox?.status === 'running' && a.app.sandbox.url}
											<iframe
												src={a.app.sandbox.url}
												title="{a.app.app_name} Preview"
												class="w-full h-full border-0"
												sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
											></iframe>
										{:else}
											<div class="flex flex-col items-center justify-center h-full text-gray-500 dark:text-gray-400">
												{#if a.app.sandbox?.status === 'pending' || a.app.sandbox?.status === 'deploying'}
													<div class="w-10 h-10 border-[3px] border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
													<p class="text-sm font-medium">Deploying sandbox...</p>
													{#if sb.sandboxDeployProgress > 0}
														<p class="text-xs text-gray-400 mt-1">{sb.sandboxDeployProgress}% complete</p>
													{/if}
												{:else if a.app.sandbox?.status === 'failed'}
													<AlertTriangle class="w-12 h-12 mb-3 text-red-400 opacity-70" />
													<p class="text-sm font-medium text-red-400">Sandbox failed to deploy</p>
													<button
														onclick={() => sb.retry(a.app!, onAppRefresh)}
														class="mt-4 flex items-center gap-2 btn-pill btn-pill-primary"
													>
														<RefreshCw class="w-4 h-4" />
														Retry Deployment
													</button>
												{:else}
													<Monitor class="w-12 h-12 mb-3 opacity-50" />
													<p class="text-sm font-medium">Start the sandbox to see a live preview</p>
													<button
														onclick={() => sb.start(a.app!, onAppRefresh)}
														class="mt-4 btn-pill btn-pill-primary"
													>
														Start Sandbox
													</button>
												{/if}
											</div>
										{/if}
									</div>
								</div>

								<!-- Stop Confirmation Dialog -->
								{#if sb.showStopConfirm}
									<div class="fixed inset-0 z-[200] flex items-center justify-center bg-black/40">
										<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 max-w-sm mx-4 shadow-xl">
											<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Stop Sandbox?</h3>
											<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
												This will stop the running sandbox. The live preview will become unavailable until you restart it.
											</p>
											<div class="flex justify-end gap-3">
												<button
													onclick={() => { sb.showStopConfirm = false; }}
													class="btn-pill btn-pill-secondary"
												>
													Cancel
												</button>
												<button
													onclick={() => sb.stop(a.app!, onAppRefresh)}
													class="btn-pill btn-pill-danger"
												>
													Stop Sandbox
												</button>
											</div>
										</div>
									</div>
								{/if}
							{/if}

							<!-- Terminal Tab -->
							{#if ed.activeTab === 'terminal'}
								<div class="bg-gray-900 rounded-xl border border-gray-700 overflow-hidden" style="height: 600px;">
									<div class="flex items-center gap-2 px-4 py-3 border-b border-gray-700 bg-gray-800/50">
										<div class="flex gap-1.5">
											<span class="w-3 h-3 rounded-full bg-red-500"></span>
											<span class="w-3 h-3 rounded-full bg-yellow-500"></span>
											<span class="w-3 h-3 rounded-full bg-green-500"></span>
										</div>
										<span class="text-sm text-gray-400 ml-2">Terminal</span>
									</div>
									<div class="flex flex-col items-center justify-center h-[calc(100%-48px)] text-gray-500">
										<Terminal class="w-12 h-12 mb-3 opacity-30" />
										<p class="text-sm font-medium">Terminal integration coming soon</p>
										<p class="text-xs text-gray-600 mt-1">
											{#if a.app?.sandbox?.status === 'running'}
												Sandbox is running — terminal access will connect to your sandbox
											{:else}
												Start a sandbox to enable terminal access
											{/if}
										</p>
									</div>
								</div>
							{/if}

							<!-- Code Workspace Tab -->
							{#if ed.activeTab === 'code'}
								<div class="bg-gray-900 rounded-xl border border-gray-700 overflow-hidden" style="height: 600px;">
									<div class="flex h-full">
										<!-- File Tree Sidebar -->
										<div class="w-64 flex-shrink-0 border-r border-gray-700 flex flex-col bg-gray-800/50">
											<div class="flex items-center justify-between px-3 py-2 border-b border-gray-700">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Explorer</span>
												<span class="text-xs text-gray-500 bg-gray-700/50 px-1.5 py-0.5 rounded-full">{ed.files.length}</span>
											</div>
											{#if ed.files.length > 6}
												<div class="px-2 py-2 border-b border-gray-700">
													<div class="relative">
														<Search class="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-500" />
														<input
															type="text"
															placeholder="Search files..."
															bind:value={ed.fileSearchQuery}
															class="w-full pl-7 pr-2 py-1.5 text-xs bg-gray-700/50 border border-gray-600 rounded text-gray-200 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
														/>
													</div>
												</div>
											{/if}
											<div class="flex-1 overflow-y-auto p-1">
												{#if ed.filteredTreeNodes.length > 0}
													<FileTree
														nodes={ed.filteredTreeNodes}
														selectedFile={ed.selectedFile}
														onFileSelect={(file) => ed.selectFile(file)}
													/>
												{:else if ed.files.length === 0}
													<div class="flex flex-col items-center justify-center h-full text-gray-500 text-sm">
														<Code2 class="w-8 h-8 mb-2 opacity-50" />
														<p>No files generated yet</p>
													</div>
												{:else}
													<div class="p-3 text-center text-xs text-gray-500">
														No files match "{ed.fileSearchQuery}"
													</div>
												{/if}
											</div>
										</div>

										<!-- Editor Panel -->
										<div class="flex-1 flex flex-col min-w-0">
											{#if ed.saveError}
												<div class="flex items-center gap-2 px-3 py-1.5 bg-red-900/30 border-b border-red-800 text-xs text-red-300">
													<AlertTriangle class="w-3.5 h-3.5 flex-shrink-0" />
													<span class="flex-1">{ed.saveError}</span>
													<button onclick={() => ed.clearSaveError()} class="text-red-400 hover:text-red-200">Dismiss</button>
												</div>
											{/if}
											{#if ed.selectedFile}
												<EditorToolbar
													filename={ed.selectedFile.path || ed.selectedFile.name}
													isEditing={ed.isEditing}
													isDirty={ed.isDirty}
													readonly={!ed.isEditing}
													cursorLine={ed.cursorLine}
													cursorColumn={ed.cursorColumn}
													onToggleEdit={() => ed.toggleEdit()}
													onSave={() => ed.saveFile(appId, ed.editorValue)}
													onCopy={() => ed.copyToClipboard()}
												/>

												<div
													class="flex-1 overflow-hidden transition-opacity duration-150"
													class:opacity-30={ed.isTransitioning}
													class:editor-editing={ed.isEditing}
												>
													{#if ed.fileLoading}
														<div class="flex items-center justify-center h-full text-gray-400">
															<div class="flex flex-col items-center gap-3">
																<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
																<span class="text-sm">Loading file...</span>
															</div>
														</div>
													{:else}
														<MonacoEditor
															bind:this={editorRef}
															bind:value={ed.editorValue}
															filename={ed.selectedFile.path || ed.selectedFile.name}
															readonly={!ed.isEditing}
															onSave={(value) => ed.saveFile(appId, value)}
															onChange={handleEditorChange}
														/>
													{/if}
												</div>

												<EditorStatusBar
													languageId={ed.languageId}
													isReadonly={!ed.isEditing}
													isEditing={ed.isEditing}
												/>
											{:else}
												<div class="flex flex-col items-center justify-center h-full text-gray-500">
													<Code2 class="w-12 h-12 mb-3 opacity-30" />
													<p class="text-sm font-medium">Select a file to view</p>
													<p class="text-xs text-gray-600 mt-1">Choose a file from the explorer</p>
												</div>
											{/if}
										</div>
									</div>
								</div>
							{/if}
						{/if}

						<!-- Error Message (if failed) -->
						{#if a.app.status === 'failed' && a.app.error_message}
							<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl p-6">
								<h2 class="text-lg font-semibold text-red-900 dark:text-red-200 mb-2 flex items-center gap-2">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
										/>
									</svg>
									Build Failed
								</h2>
								<p class="text-sm text-red-700 dark:text-red-300">{a.app.error_message}</p>
							</div>
						{/if}
					</div>

					<!-- Sidebar -->
					<div class="space-y-6">
						<!-- Actions Card -->
						<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
							<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Actions</h2>
							<AppActions app={a.app} onDeploy={handleDeploy} onDelete={handleDelete} />
						</div>

						<!-- Install Module Card -->
						{#if a.app.module_id}
							<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
								<h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
									<Package class="w-4 h-4" />
									Module
								</h2>
								{#if ms.moduleInstalled}
									<a
										href="/settings/modules"
										class="w-full inline-flex items-center justify-center gap-2 btn-pill btn-pill-success"
										aria-label="View installed module in settings"
									>
										<Package class="w-4 h-4" />
										Installed — View in Settings
									</a>
								{:else}
									<button
										onclick={() => ms.install(a.app!.module_id!)}
										disabled={ms.isInstallingModule}
										class="w-full flex items-center justify-center gap-2 btn-pill btn-pill-primary disabled:opacity-50"
										aria-label="Install this app as a module"
									>
										{#if ms.isInstallingModule}
											<RefreshCw class="w-4 h-4 animate-spin" />
											Installing...
										{:else}
											<DownloadIcon class="w-4 h-4" />
											Install Module
										{/if}
									</button>
								{/if}
								{#if ms.moduleInstallError}
									<p class="mt-2 text-xs text-red-500 dark:text-red-400">{ms.moduleInstallError}</p>
								{/if}
							</div>
						{/if}

						<!-- Version History Card -->
						{#if a.app.status === 'generated' || a.app.status === 'deployed'}
							<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
								<div class="flex items-center justify-between mb-4">
									<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Versions</h2>
									{#if vs.versions.length > 0}
										<span class="text-xs text-gray-500 bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded-full">
											v{vs.currentVersion}
										</span>
									{/if}
								</div>
								{#if vs.versions.length > 0}
									<div class="space-y-3">
										{#each vs.versions.slice(0, 3) as version (version.id)}
											<button
												onclick={() => vs.openPreview(version)}
												class="w-full text-left btn-pill btn-pill-ghost btn-pill-sm"
											>
												<div class="flex items-center justify-between">
													<span class="text-sm font-medium text-gray-900 dark:text-white">
														v{version.versionNumber}
														{#if version.isCurrent}
															<span class="text-xs text-green-600 dark:text-green-400 ml-1">current</span>
														{/if}
													</span>
												</div>
												{#if version.label}
													<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">{version.label}</p>
												{/if}
											</button>
										{/each}
										<button
											onclick={() => { vs.timelinePanelOpen = true; }}
											class="w-full flex items-center justify-center gap-2 btn-pill btn-pill-soft btn-pill-sm"
										>
											<History class="w-4 h-4" />
											View all versions
										</button>
									</div>
								{:else}
									<p class="text-sm text-gray-500 dark:text-gray-400">No version history yet</p>
									<button
										onclick={() => { vs.saveModalOpen = true; }}
										class="mt-3 w-full btn-pill btn-pill-secondary btn-pill-sm"
									>
										Save first version
									</button>
								{/if}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Version History Panels & Modals -->
{#if a.app}
	<VersionTimelinePanel
		appId={appId}
		versions={vs.versions}
		isOpen={vs.timelinePanelOpen}
		isLoading={vs.versionsLoading}
		onClose={() => { vs.timelinePanelOpen = false; }}
		onPreview={(version) => vs.openPreview(version)}
		onRestore={(version) => vs.openRestore(version)}
		onCompare={() => vs.openDiff()}
	/>

	<SaveVersionModal
		appId={appId}
		currentVersion={vs.currentVersion}
		isOpen={vs.saveModalOpen}
		isSaving={vs.isSavingVersion}
		onClose={() => { vs.saveModalOpen = false; }}
		onSave={(label) => vs.saveVersion(a.app!.workspace_id, appId, label, handleVersionReloads)}
	/>

	{#if vs.previewVersion}
		<VersionPreviewModal
			version={vs.previewVersion}
			isOpen={vs.previewModalOpen}
			onClose={() => { vs.previewModalOpen = false; }}
			onRestore={(version) => vs.openRestore(version)}
		/>
	{/if}

	{#if vs.restoreVersion}
		<RestoreConfirmDialog
			version={vs.restoreVersion}
			currentVersion={vs.currentVersion}
			isOpen={vs.restoreDialogOpen}
			isRestoring={vs.isRestoring}
			onClose={() => { vs.restoreDialogOpen = false; }}
			onConfirm={() => vs.confirmRestore(a.app!.workspace_id, appId, handleVersionReloads)}
		/>
	{/if}

	{#if vs.diffModalOpen && vs.diffFromVersion && vs.diffToVersion}
		<VersionDiffModal
			isOpen={vs.diffModalOpen}
			workspaceId={a.app.workspace_id}
			fromVersion={vs.diffFromVersion.version_number}
			toVersion={vs.diffToVersion.version_number}
			fromDisplayNum={extractDisplayNumber(vs.diffFromVersion.version_number)}
			toDisplayNum={extractDisplayNumber(vs.diffToVersion.version_number)}
			onClose={() => { vs.diffModalOpen = false; }}
		/>
	{/if}
{/if}

<style>
	.editor-editing {
		box-shadow:
			inset 0 0 0 1px rgba(99, 102, 241, 0.3),
			inset 0 0 20px rgba(99, 102, 241, 0.05);
	}

	.health-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.health-dot.running {
		background: #22c55e;
		box-shadow: 0 0 4px rgba(34, 197, 94, 0.5);
		animation: health-pulse 2s ease-in-out infinite;
	}

	@keyframes health-pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}
</style>
