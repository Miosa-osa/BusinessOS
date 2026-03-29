<script lang="ts">
	import { fly } from 'svelte/transition';
	import ProgressPanel, { type DelegatedTask } from '$lib/components/chat/panels/ProgressPanel.svelte';
	import ContextPanel, { type ActiveResource } from '$lib/components/chat/panels/ContextPanel.svelte';
	import { getArtifactIcon, getArtifactColor } from '../chatActions';

	interface ArtifactItem {
		id: string;
		title: string;
		type: string;
		content: string;
		fromMessage?: boolean;
	}

	interface ContextItem {
		id: string;
		name: string;
	}

	type PanelTab = 'progress' | 'context' | 'artifacts';

	interface Props {
		rightPanelTab: PanelTab;
		rightPanelWidth: number;
		delegatedTasks: DelegatedTask[];
		activeResources: ActiveResource[];
		allArtifacts: ArtifactItem[];
		availableContexts: ContextItem[];
		selectedContextIds: string[];
		onTabChange: (tab: PanelTab) => void;
		onClose: () => void;
		onStartResize: (e: MouseEvent) => void;
		onContextToggle: (id: string) => void;
		onMemoriesSelected: (ids: string[]) => void;
		onSelectArtifact: (id: string) => void;
		onViewMessageArtifact: (artifact: { title: string; type: string; content: string }) => void;
	}

	let {
		rightPanelTab,
		rightPanelWidth,
		delegatedTasks,
		activeResources,
		allArtifacts,
		availableContexts,
		selectedContextIds,
		onTabChange,
		onClose,
		onStartResize,
		onContextToggle,
		onMemoriesSelected,
		onSelectArtifact,
		onViewMessageArtifact,
	}: Props = $props();

	let artifactsCount = $derived(allArtifacts.filter(a => !a.fromMessage).length);
</script>

<!-- Resize Handle -->
<div
	class="w-1 h-full bg-gray-200 hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors relative group"
	onmousedown={onStartResize}
	role="separator"
	aria-orientation="vertical"
>
	<div class="absolute inset-y-0 -left-1 -right-1 group-hover:bg-blue-400/20"></div>
	<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex flex-col gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
		<div class="w-1 h-1 rounded-full bg-gray-400"></div>
		<div class="w-1 h-1 rounded-full bg-gray-400"></div>
		<div class="w-1 h-1 rounded-full bg-gray-400"></div>
	</div>
</div>

<div
	class="h-full flex flex-col bg-white dark:bg-gray-900 flex-shrink-0"
	style="width: {rightPanelWidth}px"
	transition:fly={{ x: 320, duration: 200 }}
>
	<!-- Panel Tabs -->
	<div class="flex border-b border-gray-200 dark:border-gray-700">
		<button
			onclick={() => onTabChange('progress')}
			class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'progress' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
		>
			Progress
		</button>
		<button
			onclick={() => onTabChange('context')}
			class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'context' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
		>
			Context
		</button>
		<button
			onclick={() => onTabChange('artifacts')}
			class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'artifacts' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
		>
			Artifacts
			{#if artifactsCount > 0}
				<span class="ml-1 px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-gray-200">{artifactsCount}</span>
			{/if}
		</button>
		<button
			onclick={onClose}
			class="btn-pill btn-pill-ghost btn-pill-icon"
			aria-label="Close panel"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
			</svg>
		</button>
	</div>

	<!-- Panel Content -->
	<div class="flex-1 overflow-hidden">
		{#if rightPanelTab === 'progress'}
			<ProgressPanel tasks={delegatedTasks} />
		{:else if rightPanelTab === 'context'}
			<ContextPanel
				resources={activeResources}
				availableContexts={availableContexts}
				{selectedContextIds}
				onContextToggle={(id) => onContextToggle(id)}
				{onMemoriesSelected}
			/>
		{:else if rightPanelTab === 'artifacts'}
			<!-- Artifacts List in Panel -->
			<div class="flex flex-col h-full bg-gray-50 dark:bg-[#1c1c1e]">
				<div class="p-4 border-b border-white/10">
					<div class="flex items-center justify-between">
						<h3 class="text-sm font-semibold text-gray-100">Artifacts</h3>
						{#if allArtifacts.length > 0}
							<span class="text-xs text-gray-500">{allArtifacts.length} items</span>
						{/if}
					</div>
				</div>
				<div class="flex-1 overflow-y-auto p-2">
					{#if allArtifacts.length === 0}
						<div class="flex flex-col items-center justify-center py-12 px-4 text-center">
							<svg class="w-10 h-10 text-gray-500 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
							<p class="text-sm text-gray-400">No artifacts yet</p>
							<p class="text-xs text-gray-500 mt-1">Artifacts created by AI will appear here</p>
						</div>
					{:else}
						<div class="space-y-1">
							{#each allArtifacts as artifact (artifact.id)}
								<button
									onclick={() => {
										if (artifact.fromMessage) {
											onViewMessageArtifact({
												title: artifact.title,
												type: artifact.type,
												content: artifact.content || '',
											});
										} else {
											onSelectArtifact(artifact.id);
										}
									}}
									class="w-full p-3 rounded-lg hover:bg-white/5 transition-colors text-left group"
								>
									<div class="flex items-start gap-3">
										<div class="w-8 h-8 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getArtifactIcon(artifact.type)} />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-100 truncate">{artifact.title}</p>
											<p class="text-xs text-gray-500 capitalize">{artifact.type}</p>
										</div>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>
