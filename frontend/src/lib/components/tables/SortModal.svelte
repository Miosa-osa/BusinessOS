<script lang="ts">
	/**
	 * SortModal - Multi-column sorting configuration
	 * NocoDB-style sort builder with add/remove/reorder
	 */
	import { X, Plus, Trash2, GripVertical, ArrowUp, ArrowDown } from 'lucide-svelte';
	import type { Column, Sort } from '$lib/api/tables/types';

	interface Props {
		open: boolean;
		columns: Column[];
		sorts: Sort[];
		onClose: () => void;
		onSave: (sorts: Sort[]) => void;
	}

	let { open, columns, sorts, onClose, onSave }: Props = $props();

	// Local copy of sorts for editing
	let localSorts = $state<Sort[]>([]);

	// Initialize local sorts when modal opens
	$effect(() => {
		if (open) {
			localSorts = sorts.map((s) => ({ ...s }));
		}
	});

	// Get sortable columns (exclude computed types for now)
	const sortableColumns = $derived(
		columns.filter((c) => !['formula', 'rollup', 'lookup', 'button'].includes(c.type))
	);

	// Get available columns (not already used in sorts)
	const availableColumns = $derived(
		sortableColumns.filter((c) => !localSorts.some((s) => s.column_id === c.id))
	);

	function addSort() {
		if (availableColumns.length === 0) return;

		const newSort: Sort = {
			id: crypto.randomUUID(),
			column_id: availableColumns[0].id,
			direction: 'asc'
		};
		localSorts = [...localSorts, newSort];
	}

	function removeSort(sortId: string) {
		localSorts = localSorts.filter((s) => s.id !== sortId);
	}

	function updateSortColumn(sortId: string, columnId: string) {
		localSorts = localSorts.map((s) => (s.id === sortId ? { ...s, column_id: columnId } : s));
	}

	function toggleDirection(sortId: string) {
		localSorts = localSorts.map((s) =>
			s.id === sortId ? { ...s, direction: s.direction === 'asc' ? 'desc' : 'asc' } : s
		);
	}

	function moveSort(index: number, direction: 'up' | 'down') {
		if (
			(direction === 'up' && index === 0) ||
			(direction === 'down' && index === localSorts.length - 1)
		) {
			return;
		}

		const newSorts = [...localSorts];
		const targetIndex = direction === 'up' ? index - 1 : index + 1;
		[newSorts[index], newSorts[targetIndex]] = [newSorts[targetIndex], newSorts[index]];
		localSorts = newSorts;
	}

	function handleSave() {
		onSave(localSorts);
		onClose();
	}

	function handleClearAll() {
		localSorts = [];
	}

	function getColumnName(columnId: string): string {
		return columns.find((c) => c.id === columnId)?.name ?? 'Unknown';
	}

	// Handle escape key
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onClose();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-40 bg-black/50"
		onclick={onClose}
		role="button"
		tabindex="-1"
		aria-label="Close modal"
	></div>

	<!-- Modal -->
	<div
		class="fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 rounded-xl bg-white shadow-2xl"
		role="dialog"
		aria-modal="true"
		aria-labelledby="sort-modal-title"
	>
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
			<h2 id="sort-modal-title" class="text-lg font-semibold text-gray-900">Sort</h2>
			<button
				type="button"
				class="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
				onclick={onClose}
			>
				<X class="h-5 w-5" />
			</button>
		</div>

		<!-- Content -->
		<div class="max-h-[400px] overflow-y-auto p-6">
			{#if localSorts.length === 0}
				<!-- Empty State -->
				<div class="py-8 text-center">
					<div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100">
						<ArrowUp class="h-6 w-6 text-gray-400" />
					</div>
					<p class="mb-4 text-sm text-gray-500">No sorts applied</p>
					<button
						type="button"
						class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
						onclick={addSort}
					>
						<Plus class="h-4 w-4" />
						Add sort
					</button>
				</div>
			{:else}
				<!-- Sort List -->
				<div class="space-y-3">
					{#each localSorts as sort, index (sort.id)}
						<div class="flex items-center gap-3 rounded-lg border border-gray-200 bg-gray-50 p-3">
							<!-- Drag Handle (placeholder for future drag-and-drop) -->
							<div class="cursor-grab text-gray-400">
								<GripVertical class="h-4 w-4" />
							</div>

							<!-- Sort Order Label -->
							<span class="text-xs font-medium text-gray-500">
								{index === 0 ? 'Sort by' : 'Then by'}
							</span>

							<!-- Column Selector -->
							<select
								class="flex-1 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
								value={sort.column_id}
								onchange={(e) => updateSortColumn(sort.id, (e.target as HTMLSelectElement).value)}
							>
								<!-- Current column always shown -->
								<option value={sort.column_id}>{getColumnName(sort.column_id)}</option>
								<!-- Other available columns -->
								{#each availableColumns as col}
									<option value={col.id}>{col.name}</option>
								{/each}
							</select>

							<!-- Direction Toggle -->
							<button
								type="button"
								class="flex items-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm hover:bg-gray-50"
								onclick={() => toggleDirection(sort.id)}
							>
								{#if sort.direction === 'asc'}
									<ArrowUp class="h-4 w-4 text-blue-600" />
									<span>A → Z</span>
								{:else}
									<ArrowDown class="h-4 w-4 text-blue-600" />
									<span>Z → A</span>
								{/if}
							</button>

							<!-- Move Buttons -->
							<div class="flex items-center gap-1">
								<button
									type="button"
									class="rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-gray-600 disabled:opacity-50"
									onclick={() => moveSort(index, 'up')}
									disabled={index === 0}
									title="Move up"
								>
									<ArrowUp class="h-4 w-4" />
								</button>
								<button
									type="button"
									class="rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-gray-600 disabled:opacity-50"
									onclick={() => moveSort(index, 'down')}
									disabled={index === localSorts.length - 1}
									title="Move down"
								>
									<ArrowDown class="h-4 w-4" />
								</button>
							</div>

							<!-- Remove Button -->
							<button
								type="button"
								class="rounded-lg p-2 text-gray-400 hover:bg-red-50 hover:text-red-600"
								onclick={() => removeSort(sort.id)}
								title="Remove sort"
							>
								<Trash2 class="h-4 w-4" />
							</button>
						</div>
					{/each}
				</div>

				<!-- Add Another -->
				{#if availableColumns.length > 0}
					<button
						type="button"
						class="mt-4 flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-gray-300 py-2 text-sm text-gray-500 hover:border-gray-400 hover:bg-gray-50 hover:text-gray-600"
						onclick={addSort}
					>
						<Plus class="h-4 w-4" />
						Add another sort
					</button>
				{/if}
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex items-center justify-between border-t border-gray-200 px-6 py-4">
			<button
				type="button"
				class="text-sm text-gray-500 hover:text-gray-700"
				onclick={handleClearAll}
				disabled={localSorts.length === 0}
			>
				Clear all
			</button>
			<div class="flex items-center gap-3">
				<button
					type="button"
					class="rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
					onclick={onClose}
				>
					Cancel
				</button>
				<button
					type="button"
					class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					onclick={handleSave}
				>
					Apply sort
				</button>
			</div>
		</div>
	</div>
{/if}
