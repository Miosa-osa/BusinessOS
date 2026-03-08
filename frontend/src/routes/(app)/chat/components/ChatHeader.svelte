<script lang="ts">
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import ChatModelSelector from '$lib/components/chat/input/ChatModelSelector.svelte';
	import RoleContextBadge from '$lib/components/chat/shared/RoleContextBadge.svelte';
	import AgentSelector from '$lib/components/agents/AgentSelector.svelte';
	import type { ModelOption } from '$lib/utils/chatHelpers';
	import type { CustomAgent } from '$lib/api/ai/types';

	interface Project {
		id: string;
		name: string;
		description?: string;
	}

	interface NodeItem {
		id: string;
		name: string;
		purpose?: string | null;
	}

	interface Props {
		// Sidebar
		chatSidebarOpen: boolean;
		onToggleSidebar: () => void;

		// Model selector
		selectedModel: string;
		currentModelName: string;
		warmingUpModel: string | null;
		installedModels: ModelOption[];
		ollamaCloudModels: ModelOption[];
		configuredProviders: Set<string>;
		loadingModels: boolean;
		onSelectModel: (id: string) => void;

		// Agent selector
		customAgents: CustomAgent[];
		selectedAgentId: string | null;
		onSelectAgent: (agent: CustomAgent | null) => void;

		// Project selector
		selectedProject: Project | null;
		projectsList: Project[];
		loadingProjects: boolean;
		showProjectDropdown: boolean;
		projectDropdownIndex: number;
		inputHasValue: boolean;
		onToggleProjectDropdown: () => void;
		onSelectProject: (id: string) => void;
		onProjectDropdownKeydown: (e: KeyboardEvent) => void;
		onCreateNewProject: () => void;

		// Node indicator
		activeNode: NodeItem | null;
		showNodeDropdown: boolean;
		onToggleNodeDropdown: () => void;
		onDeactivateNode: () => void;

		// Right panel
		rightPanelOpen: boolean;
		artifactsCount: number;
		onTogglePanel: () => void;

		// Context dropdown close (click-outside side-effect)
		onCloseHeaderDropdowns?: () => void;
	}

	let {
		chatSidebarOpen,
		onToggleSidebar,
		selectedModel,
		currentModelName,
		warmingUpModel,
		installedModels,
		ollamaCloudModels,
		configuredProviders,
		loadingModels,
		onSelectModel,
		customAgents,
		selectedAgentId,
		onSelectAgent,
		selectedProject,
		projectsList,
		loadingProjects,
		showProjectDropdown,
		projectDropdownIndex,
		inputHasValue,
		onToggleProjectDropdown,
		onSelectProject,
		onProjectDropdownKeydown,
		onCreateNewProject,
		activeNode,
		showNodeDropdown,
		onToggleNodeDropdown,
		onDeactivateNode,
		rightPanelOpen,
		artifactsCount,
		onTogglePanel,
	}: Props = $props();
</script>

<div class="h-12 flex items-center justify-between px-4 flex-shrink-0 min-w-0">
	<!-- Left group: Hamburger + Model Selector -->
	<div class="flex items-center gap-1 flex-shrink-0">
		<button
			onclick={onToggleSidebar}
			class="btn-pill btn-pill-ghost btn-pill-icon flex-shrink-0"
			aria-label="Toggle sidebar"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				{#if chatSidebarOpen}
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
				{:else}
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				{/if}
			</svg>
		</button>

		<ChatModelSelector
			{selectedModel}
			{currentModelName}
			{warmingUpModel}
			{installedModels}
			{ollamaCloudModels}
			{configuredProviders}
			{loadingModels}
			onSelectModel={(id) => onSelectModel(id)}
		/>
	</div>

	<!-- Center: Role Context Badge -->
	<div class="flex items-center justify-center flex-1 min-w-0">
		<RoleContextBadge size="sm" showLabel={true} showTooltip={true} />
	</div>

	<!-- Right group: Project, Node, Panel -->
	<div class="flex items-center gap-2 min-w-0 flex-1 justify-end">
		<!-- Agent Selector -->
		<div class="relative flex-shrink-0" style="width: 220px;">
			<AgentSelector
				agents={customAgents}
				selectedId={selectedAgentId}
				onSelect={onSelectAgent}
				placeholder="Default Agent"
				includeDefault={true}
				onManage={() => goto('/agents')}
			/>
		</div>

		<!-- Project Selector (required for chat) -->
		<div class="relative flex-shrink-0">
			<button
				onclick={onToggleProjectDropdown}
				onkeydown={onProjectDropdownKeydown}
				class="btn-pill btn-pill-icon {selectedProject ? 'btn-pill-secondary' : 'btn-pill-warning'}"
				title={selectedProject ? selectedProject.name : 'Select Project (Required)'}
			>
				<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
			</button>

			{#if showProjectDropdown}
				<div
					class="absolute left-0 top-full mt-2 w-72 bg-white border border-gray-200 rounded-xl shadow-lg py-2 z-20 max-h-80 overflow-y-auto"
					transition:fly={{ y: -10, duration: 200 }}
					onkeydown={onProjectDropdownKeydown}
					tabindex="-1"
				>
					<div class="px-3 py-1.5">
						<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Select Project</span>
					</div>
					{#if loadingProjects}
						<div class="px-4 py-3 text-sm text-gray-500">Loading projects...</div>
					{:else if projectsList.length === 0}
						<div class="px-4 py-6 text-center">
							<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
							<p class="text-sm text-gray-500">No projects yet</p>
							<a href="/projects" class="text-sm text-blue-600 hover:underline">Create a project</a>
						</div>
					{:else}
						{#each projectsList as project, i (project.id)}
							{@const isSelected = selectedProject?.id === project.id}
							{@const isFocused = projectDropdownIndex === i}
							<button
								onclick={() => onSelectProject(project.id)}
								class="w-full px-4 py-2 text-left transition-colors flex items-center gap-3 {isSelected ? 'bg-purple-50' : ''} {isFocused ? 'bg-blue-50 ring-2 ring-blue-400 ring-inset' : 'hover:bg-gray-50'}"
							>
								<div class="w-8 h-8 rounded-lg {isSelected ? 'bg-purple-500 text-white' : isFocused ? 'bg-blue-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
									{#if isSelected}
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
									{:else}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium {isSelected ? 'text-purple-600' : isFocused ? 'text-blue-600' : 'text-gray-700'} truncate">{project.name}</div>
									{#if project.description}
										<div class="text-xs text-gray-500 truncate">{project.description}</div>
									{/if}
								</div>
							</button>
						{/each}
					{/if}
					<!-- Create New Project Option -->
					<div class="border-t border-gray-100 mt-1 pt-1">
						<button
							onclick={onCreateNewProject}
							class="w-full flex items-center gap-3 text-left btn-pill btn-pill-ghost btn-pill-sm {projectDropdownIndex === projectsList.length ? 'btn-pill-soft' : ''}"
						>
							<div class="w-8 h-8 rounded-lg {projectDropdownIndex === projectsList.length ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600'} flex items-center justify-center flex-shrink-0">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium {projectDropdownIndex === projectsList.length ? 'text-gray-900' : 'text-gray-700'}">Create new project</div>
								<div class="text-xs text-gray-500">Start a new project for this chat</div>
							</div>
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Active Node Indicator -->
		{#if activeNode}
			<div class="relative flex-shrink-0">
				<button
					onclick={onToggleNodeDropdown}
					class="btn-pill btn-pill-icon btn-pill-secondary"
					title={activeNode.name}
				>
					<svg class="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
						<path d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
				</button>

				{#if showNodeDropdown}
					<div
						class="absolute right-0 top-full mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg p-3 z-20"
						transition:fly={{ y: -10, duration: 200 }}
					>
						<div class="text-xs font-semibold text-gray-500 uppercase mb-2">Active Node</div>
						<div class="mb-3">
							<p class="text-sm font-medium text-gray-900">{activeNode.name}</p>
							{#if activeNode.purpose}
								<p class="text-xs text-gray-500 mt-1 line-clamp-2">{activeNode.purpose}</p>
							{/if}
						</div>
						<div class="flex gap-2">
							<a
								href="/nodes/{activeNode.id}"
								class="flex-1 text-center btn-pill btn-pill-ghost btn-pill-xs"
							>
								View
							</a>
							<button
								onclick={onDeactivateNode}
								class="flex-1 btn-pill btn-pill-danger btn-pill-xs"
							>
								Deactivate
							</button>
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<a
				href="/nodes"
				class="btn-pill btn-pill-secondary btn-pill-sm whitespace-nowrap flex-shrink-0"
			>
				<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				<span>MIOSA Platform</span>
			</a>
		{/if}

		<!-- Panel Toggle (combines Progress, Context, Artifacts) -->
		<button
			onclick={onTogglePanel}
			class="btn-pill btn-pill-icon {rightPanelOpen ? 'btn-pill-secondary' : 'btn-pill-ghost'}"
			title="Toggle Side Panel"
		>
			<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
			</svg>
			{#if artifactsCount > 0}
				<span class="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-blue-500 text-white text-[10px] font-medium flex items-center justify-center">{artifactsCount}</span>
			{/if}
		</button>
	</div>
</div>
