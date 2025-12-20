<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { fly, fade } from 'svelte/transition';

	type Priority = 'critical' | 'high' | 'medium' | 'low';

	interface Props {
		projectId?: string;
		projectName?: string;
		status?: string;
		onAdd?: (task: { title: string; priority: Priority; assigneeId?: string; dueDate?: string }) => void;
		onCancel?: () => void;
	}

	let { projectId, projectName, status, onAdd, onCancel }: Props = $props();

	let title = $state('');
	let priority: Priority = $state('medium');
	let assigneeId = $state<string | undefined>(undefined);
	let dueDate = $state<string | undefined>(undefined);
	let isExpanded = $state(false);
	let inputRef: HTMLInputElement | undefined = $state(undefined);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && title.trim()) {
			e.preventDefault();
			handleSubmit();
		} else if (e.key === 'Escape') {
			handleCancel();
		}
	}

	function handleSubmit() {
		if (!title.trim()) return;
		onAdd?.({ title, priority, assigneeId, dueDate });
		title = '';
		priority = 'medium';
		assigneeId = undefined;
		dueDate = undefined;
		isExpanded = false;
	}

	function handleCancel() {
		title = '';
		isExpanded = false;
		onCancel?.();
	}

	function handleFocus() {
		isExpanded = true;
	}

	const priorityOptions: { value: Priority; label: string; color: string }[] = [
		{ value: 'critical', label: 'Critical', color: 'bg-red-500' },
		{ value: 'high', label: 'High', color: 'bg-orange-500' },
		{ value: 'medium', label: 'Medium', color: 'bg-yellow-500' },
		{ value: 'low', label: 'Low', color: 'bg-gray-400' }
	];
</script>

<div class="px-4 py-2" in:fade={{ duration: 150 }}>
	<div class="border border-gray-200 rounded-xl bg-white overflow-hidden focus-within:border-gray-300 focus-within:shadow-sm transition-all">
		<input
			bind:this={inputRef}
			bind:value={title}
			onkeydown={handleKeydown}
			onfocus={handleFocus}
			type="text"
			placeholder="+ Add a task..."
			class="w-full px-4 py-3 text-sm text-gray-900 placeholder-gray-400 focus:outline-none bg-transparent"
		/>

		{#if isExpanded}
			<div class="flex items-center justify-between px-3 py-2 border-t border-gray-100 bg-gray-50" in:fly={{ y: -10, duration: 150 }}>
				<div class="flex items-center gap-2">
					<!-- Project (if not already set) -->
					{#if !projectId}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="flex items-center gap-1 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200 rounded transition-colors">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
								{projectName || 'Project'}
							</DropdownMenu.Trigger>
							<DropdownMenu.Portal>
								<DropdownMenu.Content
									class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
									sideOffset={4}
									transition={fly}
									transitionConfig={{ y: -10, duration: 150 }}
								>
									<DropdownMenu.Item class="px-3 py-2 text-sm text-gray-500 hover:bg-gray-100 rounded-lg cursor-pointer">
										No projects yet
									</DropdownMenu.Item>
								</DropdownMenu.Content>
							</DropdownMenu.Portal>
						</DropdownMenu.Root>
					{/if}

					<!-- Priority -->
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="flex items-center gap-1 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200 rounded transition-colors">
							<span class="w-2 h-2 rounded-full {priorityOptions.find(p => p.value === priority)?.color}"></span>
							{priorityOptions.find(p => p.value === priority)?.label}
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="z-50 min-w-[140px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
								sideOffset={4}
								transition={fly}
								transitionConfig={{ y: -10, duration: 150 }}
							>
								{#each priorityOptions as option}
									<DropdownMenu.Item
										class="flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer"
										onclick={() => priority = option.value}
									>
										<span class="w-2 h-2 rounded-full {option.color}"></span>
										{option.label}
									</DropdownMenu.Item>
								{/each}
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>

					<!-- Assignee -->
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="flex items-center gap-1 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200 rounded transition-colors">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							Assign
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
								sideOffset={4}
								transition={fly}
								transitionConfig={{ y: -10, duration: 150 }}
							>
								<DropdownMenu.Item class="px-3 py-2 text-sm text-gray-500 hover:bg-gray-100 rounded-lg cursor-pointer">
									Unassigned
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>

					<!-- Due Date -->
					<button class="flex items-center gap-1 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200 rounded transition-colors">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						Due
					</button>
				</div>

				<div class="flex items-center gap-2">
					<button
						onclick={handleCancel}
						class="px-3 py-1.5 text-xs text-gray-600 hover:bg-gray-200 rounded-lg transition-colors"
					>
						Cancel
					</button>
					<button
						onclick={handleSubmit}
						disabled={!title.trim()}
						class="px-3 py-1.5 text-xs font-medium text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Add
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
