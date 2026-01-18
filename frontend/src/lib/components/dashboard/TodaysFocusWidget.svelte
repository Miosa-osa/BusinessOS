<script lang="ts">
	import { fly, scale } from 'svelte/transition';
	import { flip } from 'svelte/animate';

	interface FocusItem {
		id: string;
		title?: string;
		text?: string;
		description?: string;
		completed: boolean;
	}

	interface Props {
		items?: FocusItem[];
		onToggle?: (id: string) => void;
		onAdd?: (title: string, description?: string) => void;
		onRemove?: (id: string) => void;
		onReorder?: (items: FocusItem[]) => void;
		onEdit?: () => void;
	}

	let { items = [], onToggle, onAdd, onRemove, onReorder, onEdit }: Props = $props();

	let isAdding = $state(false);
	let newTitle = $state('');
	let newDescription = $state('');

	function handleToggle(id: string) {
		onToggle?.(id);
	}

	function handleAdd() {
		if (newTitle.trim()) {
			onAdd?.(newTitle.trim(), newDescription.trim() || undefined);
			newTitle = '';
			newDescription = '';
			isAdding = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleAdd();
		}
		if (e.key === 'Escape') {
			isAdding = false;
			newTitle = '';
			newDescription = '';
		}
	}
</script>

<div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-gray-800 to-gray-900 flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900">Today's Focus</h2>
		</div>
		{#if items.length > 0}
			<button
				onclick={() => onEdit?.()}
				class="text-xs text-gray-500 hover:text-gray-700 transition-colors px-2 py-1 rounded-md hover:bg-gray-100"
			>
				Edit
			</button>
		{/if}
	</div>

	{#if items.length === 0 && !isAdding}
		<div class="text-center py-8">
			<div class="w-14 h-14 bg-gradient-to-br from-gray-100 to-gray-50 rounded-xl flex items-center justify-center mx-auto mb-3 shadow-sm">
				<svg class="w-7 h-7 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<p class="text-sm text-gray-500 mb-3">No focus items for today</p>
			<button
				onclick={() => (isAdding = true)}
				class="inline-flex items-center gap-1.5 text-sm text-gray-600 hover:text-gray-900 font-medium px-3 py-1.5 rounded-lg hover:bg-gray-50 transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add your first focus
			</button>
		</div>
	{:else}
		<div class="space-y-3">
			{#each items as item, index (item.id)}
				<div
					class="group flex items-start gap-3 p-3 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors {item.completed
						? 'bg-gray-50'
						: 'bg-white'}"
					animate:flip={{ duration: 200 }}
					in:fly={{ y: 10, duration: 300, delay: index * 50 }}
				>
					<button
						onclick={() => handleToggle(item.id)}
						class="flex-shrink-0 mt-0.5 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {item.completed
							? 'bg-gray-900 border-gray-900'
							: 'border-gray-300 hover:border-gray-400'}"
					>
						{#if item.completed}
							<svg
								class="w-3 h-3 text-white"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								in:scale={{ duration: 200, start: 0.5 }}
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="3"
									d="M5 13l4 4L19 7"
								/>
							</svg>
						{/if}
					</button>
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2">
							<span class="text-sm text-gray-400 font-medium">{index + 1}.</span>
							<p
								class="text-sm font-medium {item.completed
									? 'text-gray-400 line-through'
									: 'text-gray-900'}"
							>
								{item.title || item.text}
							</p>
						</div>
						{#if item.description}
							<p class="text-xs text-gray-500 mt-1 ml-5">{item.description}</p>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if isAdding}
			<div class="mt-3 p-3 rounded-lg border border-gray-200 bg-gray-50" in:fly={{ y: 10, duration: 200 }}>
				<input
					type="text"
					bind:value={newTitle}
					onkeydown={handleKeydown}
					placeholder="What's your focus?"
					class="w-full text-sm bg-transparent border-none outline-none placeholder-gray-400"
					autofocus
				/>
				<input
					type="text"
					bind:value={newDescription}
					onkeydown={handleKeydown}
					placeholder="Add context (optional)"
					class="w-full text-xs text-gray-500 bg-transparent border-none outline-none placeholder-gray-400 mt-1"
				/>
				<div class="flex items-center gap-2 mt-2">
					<button
						onclick={handleAdd}
						disabled={!newTitle.trim()}
						class="btn-pill btn-pill-primary btn-pill-xs"
					>
						Add
					</button>
					<button
						onclick={() => {
							isAdding = false;
							newTitle = '';
							newDescription = '';
						}}
						class="btn-pill btn-pill-ghost btn-pill-xs"
					>
						Cancel
					</button>
				</div>
			</div>
		{:else if items.length < 5}
			<button
				onclick={() => (isAdding = true)}
				class="mt-3 flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700 transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add focus item
			</button>
		{/if}
	{/if}
</div>
