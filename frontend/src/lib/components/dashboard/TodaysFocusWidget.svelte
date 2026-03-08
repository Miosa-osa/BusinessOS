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

<div class="dw-focus-card">
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="dw-focus-icon-wrap">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<h2 class="dw-focus-title">Today's Focus</h2>
		</div>
		{#if items.length > 0}
			<button
				onclick={() => onEdit?.()}
				class="btn-pill btn-pill-ghost btn-pill-xs"
			>
				Edit
			</button>
		{/if}
	</div>

	{#if items.length === 0 && !isAdding}
		<div class="text-center py-8">
			<div class="dw-focus-empty-icon">
				<svg class="w-7 h-7" style="color: var(--dt3)" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<p class="text-sm mb-3" style="color: var(--dt2)">No focus items for today</p>
			<button
				onclick={() => (isAdding = true)}
				class="btn-pill btn-pill-ghost btn-pill-sm"
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
				class="group flex items-start gap-3 p-3 rounded-lg dw-focus-item {item.completed
					? 'dw-focus-item--done'
					: ''}"
					animate:flip={{ duration: 200 }}
					in:fly={{ y: 10, duration: 300, delay: index * 50 }}
				>
					<button
						onclick={() => handleToggle(item.id)}
						class="flex-shrink-0 mt-0.5 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors dw-focus-check {item.completed
							? 'dw-focus-check--done'
							: ''}"
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
							<span class="text-sm font-medium" style="color: var(--dt4)">{index + 1}.</span>
							<p
								class="text-sm font-medium {item.completed
									? 'dw-focus-text--done'
									: ''}" style="color: {item.completed ? 'var(--dt4)' : 'var(--dt)'}"
							>
								{item.title || item.text}
							</p>
						</div>
						{#if item.description}
							<p class="text-xs mt-1 ml-5" style="color: var(--dt2)">{item.description}</p>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if isAdding}
			<div class="dw-focus-add-form" in:fly={{ y: 10, duration: 200 }}>
				<input
					type="text"
					bind:value={newTitle}
					onkeydown={handleKeydown}
					placeholder="What's your focus?"
					class="dw-focus-input"
					autofocus
				/>
				<input
					type="text"
					bind:value={newDescription}
					onkeydown={handleKeydown}
					placeholder="Add context (optional)"
					class="dw-focus-input dw-focus-input--sub"
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
				class="btn-pill btn-pill-ghost btn-pill-xs mt-3"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add focus item
			</button>
		{/if}
	{/if}
</div>

<style>
	/* ── Today's Focus Widget (dw-focus-*) — Foundation Tokens ── */
	.dw-focus-card {
		background: var(--dbg, #fff);
		border-radius: 0.75rem;
		border: 1px solid var(--dbd, #e0e0e0);
		padding: 1.25rem;
		box-shadow: var(--shadow-sm, 0 1px 2px rgba(0,0,0,0.05));
		transition: box-shadow 0.3s ease;
	}
	.dw-focus-card:hover {
		box-shadow: var(--shadow-md, 0 4px 6px rgba(0,0,0,0.07));
	}
	.dw-focus-icon-wrap {
		width: 2rem;
		height: 2rem;
		border-radius: 0.5rem;
		background: linear-gradient(135deg, var(--dt2, #555), var(--dt, #111));
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-xs, 0 1px 2px rgba(0,0,0,0.05));
	}
	.dw-focus-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.dw-focus-empty-icon {
		width: 3.5rem;
		height: 3.5rem;
		background: var(--dbg2, #f5f5f5);
		border-radius: 0.75rem;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 0.75rem;
		box-shadow: var(--shadow-xs, 0 1px 2px rgba(0,0,0,0.05));
	}
	.dw-focus-item {
		border: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
		transition: border-color 0.2s;
	}
	.dw-focus-item:hover {
		border-color: var(--dbd, #e0e0e0);
	}
	.dw-focus-item--done {
		background: var(--dbg2, #f5f5f5);
	}
	.dw-focus-check {
		border-color: var(--dbd, #e0e0e0);
	}
	.dw-focus-check:hover {
		border-color: var(--dt3, #888);
	}
	.dw-focus-check--done {
		background: var(--dt, #111);
		border-color: var(--dt, #111);
	}
	.dw-focus-text--done {
		text-decoration: line-through;
	}
	.dw-focus-add-form {
		margin-top: 0.75rem;
		padding: 0.75rem;
		border-radius: 0.5rem;
		border: 1px solid var(--dbd, #e0e0e0);
		background: var(--dbg2, #f5f5f5);
	}
	.dw-focus-input {
		width: 100%;
		font-size: 0.875rem;
		background: transparent;
		border: none;
		outline: none;
		color: var(--dt, #111);
	}
	.dw-focus-input::placeholder {
		color: var(--dt4, #bbb);
	}
	.dw-focus-input--sub {
		font-size: 0.75rem;
		color: var(--dt2, #555);
		margin-top: 0.25rem;
	}
</style>
