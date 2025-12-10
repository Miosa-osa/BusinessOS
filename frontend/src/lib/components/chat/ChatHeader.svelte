<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { fly, fade } from 'svelte/transition';

	interface Props {
		title?: string;
		contextOptions?: Array<{ id: string; name: string }>;
		modelOptions?: Array<{ id: string; name: string; type: 'local' | 'cloud' }>;
		selectedContext?: string;
		selectedModel?: string;
		showBackButton?: boolean;
		onBack?: () => void;
		onTitleChange?: (title: string) => void;
		onContextChange?: (contextId: string) => void;
		onModelChange?: (modelId: string) => void;
		onLinkProject?: () => void;
		onExport?: () => void;
		onClearMessages?: () => void;
		onArchive?: () => void;
		onDelete?: () => void;
	}

	let {
		title = 'New Conversation',
		contextOptions = [
			{ id: 'default', name: 'Default' },
			{ id: 'daily', name: 'Daily Planning' },
			{ id: 'project', name: 'Project Analysis' },
			{ id: 'strategic', name: 'Strategic Thinking' },
			{ id: 'code', name: 'Code Review' }
		],
		modelOptions = [
			{ id: 'qwen3-coder:480b', name: 'Qwen3 Coder 480B', type: 'local' as const },
			{ id: 'qwen3-coder:30b', name: 'Qwen3 Coder 30B', type: 'local' as const },
			{ id: 'deepseek-r1:70b', name: 'DeepSeek R1 70B', type: 'local' as const },
			{ id: 'llama3.3:70b', name: 'Llama 3.3 70B', type: 'local' as const },
			{ id: 'claude', name: 'Claude 3.5 Sonnet', type: 'cloud' as const },
			{ id: 'gpt4o', name: 'GPT-4o', type: 'cloud' as const }
		],
		selectedContext = 'default',
		selectedModel = 'qwen3-coder:480b',
		showBackButton = false,
		onBack,
		onTitleChange,
		onContextChange,
		onModelChange,
		onLinkProject,
		onExport,
		onClearMessages,
		onArchive,
		onDelete
	}: Props = $props();

	let isEditingTitle = $state(false);
	let editedTitle = $state(title);

	function handleTitleEdit() {
		isEditingTitle = true;
		editedTitle = title;
	}

	function handleTitleSave() {
		if (editedTitle.trim()) {
			onTitleChange?.(editedTitle.trim());
		}
		isEditingTitle = false;
	}

	function handleTitleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			handleTitleSave();
		} else if (e.key === 'Escape') {
			isEditingTitle = false;
		}
	}

	const selectedContextName = $derived(contextOptions.find(c => c.id === selectedContext)?.name || 'Default');
	const selectedModelName = $derived(modelOptions.find(m => m.id === selectedModel)?.name || 'Local LLM');

	const localModels = $derived(modelOptions.filter(m => m.type === 'local'));
	const cloudModels = $derived(modelOptions.filter(m => m.type === 'cloud'));
</script>

<div class="h-14 border-b border-gray-100 bg-white flex items-center justify-between px-4">
	<div class="flex items-center gap-3 flex-1 min-w-0">
		{#if showBackButton}
			<button
				onclick={onBack}
				class="p-2 -ml-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors lg:hidden"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
		{/if}

		<!-- Title -->
		{#if isEditingTitle}
			<input
				type="text"
				bind:value={editedTitle}
				onblur={handleTitleSave}
				onkeydown={handleTitleKeydown}
				class="flex-1 min-w-0 px-2 py-1 text-base font-medium text-gray-900 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
				autofocus
			/>
		{:else}
			<button
				onclick={handleTitleEdit}
				class="group flex items-center gap-2 min-w-0"
			>
				<span class="text-base font-medium text-gray-900 truncate">{title}</span>
				<svg class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
				</svg>
			</button>
		{/if}
	</div>

	<div class="flex items-center gap-2">
		<!-- Context Selector -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
				</svg>
				<span class="hidden sm:inline">{selectedContextName}</span>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 min-w-[200px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
					sideOffset={4}
					transition={fly}
					transitionConfig={{ y: -10, duration: 150 }}
				>
					<div class="px-3 py-2 text-xs font-medium text-gray-400 uppercase tracking-wider">Context</div>
					{#each contextOptions as context}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm rounded-lg cursor-pointer transition-colors
								{context.id === selectedContext ? 'bg-gray-100 text-gray-900' : 'text-gray-700 hover:bg-gray-100'}"
							onclick={() => onContextChange?.(context.id)}
						>
							<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center
								{context.id === selectedContext ? 'border-gray-900' : 'border-gray-300'}">
								{#if context.id === selectedContext}
									<div class="w-2 h-2 bg-gray-900 rounded-full"></div>
								{/if}
							</div>
							{context.name}
						</DropdownMenu.Item>
					{/each}
					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
						</svg>
						Load from Contexts...
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Model Selector -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
				<span class="hidden sm:inline">{selectedModelName}</span>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 min-w-[200px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
					sideOffset={4}
					transition={fly}
					transitionConfig={{ y: -10, duration: 150 }}
				>
					<div class="px-3 py-2 text-xs font-medium text-gray-400 uppercase tracking-wider">Local Models</div>
					{#each localModels as model}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm rounded-lg cursor-pointer transition-colors
								{model.id === selectedModel ? 'bg-gray-100 text-gray-900' : 'text-gray-700 hover:bg-gray-100'}"
							onclick={() => onModelChange?.(model.id)}
						>
							<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center
								{model.id === selectedModel ? 'border-gray-900' : 'border-gray-300'}">
								{#if model.id === selectedModel}
									<div class="w-2 h-2 bg-gray-900 rounded-full"></div>
								{/if}
							</div>
							{model.name}
						</DropdownMenu.Item>
					{/each}
					{#if cloudModels.length > 0}
						<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />
						<div class="px-3 py-2 text-xs font-medium text-gray-400 uppercase tracking-wider">Cloud (Optional)</div>
						{#each cloudModels as model}
							<DropdownMenu.Item
								class="flex items-center gap-3 px-3 py-2 text-sm rounded-lg cursor-pointer transition-colors
									{model.id === selectedModel ? 'bg-gray-100 text-gray-900' : 'text-gray-700 hover:bg-gray-100'}"
								onclick={() => onModelChange?.(model.id)}
							>
								<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center
									{model.id === selectedModel ? 'border-gray-900' : 'border-gray-300'}">
									{#if model.id === selectedModel}
										<div class="w-2 h-2 bg-gray-900 rounded-full"></div>
									{/if}
								</div>
								{model.name}
							</DropdownMenu.Item>
						{/each}
					{/if}
					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						Manage Models...
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Menu -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 min-w-[180px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
					sideOffset={4}
					align="end"
					transition={fly}
					transitionConfig={{ y: -10, duration: 150 }}
				>
					{#if onLinkProject}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onLinkProject}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							Link to project
						</DropdownMenu.Item>
					{/if}
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
						</svg>
						Link to context
					</DropdownMenu.Item>
					{#if onExport}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onExport}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
							</svg>
							Export conversation
						</DropdownMenu.Item>
					{/if}
					{#if onClearMessages}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onClearMessages}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Clear messages
						</DropdownMenu.Item>
					{/if}
					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />
					{#if onArchive}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onArchive}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
							</svg>
							Archive
						</DropdownMenu.Item>
					{/if}
					{#if onDelete}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg cursor-pointer transition-colors"
							onclick={onDelete}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete conversation
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>
	</div>
</div>
