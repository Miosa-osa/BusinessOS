<script lang="ts">
	/**
	 * AddTableModal - Create new table dialog
	 */
	import { X, Table2, FileSpreadsheet, Database, Upload } from 'lucide-svelte';
	import type { CreateTableData, TableSource } from '$lib/api/tables/types';

	interface Props {
		open: boolean;
		onClose: () => void;
		onCreate: (data: CreateTableData) => void;
	}

	let { open, onClose, onCreate }: Props = $props();

	let name = $state('');
	let description = $state('');
	let source = $state<TableSource>('custom');
	let loading = $state(false);
	let error = $state('');

	const sourceOptions: { value: TableSource; label: string; icon: typeof Table2; description: string }[] = [
		{
			value: 'custom',
			label: 'Blank Table',
			icon: Table2,
			description: 'Start with an empty table'
		},
		{
			value: 'import',
			label: 'Import CSV/Excel',
			icon: Upload,
			description: 'Import data from a file'
		},
		{
			value: 'integration',
			label: 'From Integration',
			icon: Database,
			description: 'Sync from connected app'
		}
	];

	function resetForm() {
		name = '';
		description = '';
		source = 'custom';
		error = '';
	}

	function handleClose() {
		resetForm();
		onClose();
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!name.trim()) {
			error = 'Table name is required';
			return;
		}

		loading = true;
		error = '';

		try {
			await onCreate({
				name: name.trim(),
				description: description.trim() || undefined,
				source,
				// Default columns for a new table
				columns: [
					{
						name: 'Name',
						type: 'text',
						is_primary: true
					}
				]
			});
			handleClose();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create table';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			handleClose();
		}
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		onkeydown={handleKeydown}
	>
		<!-- Backdrop -->
		<button
			type="button"
			class="absolute inset-0 cursor-default"
			onclick={handleClose}
			aria-label="Close modal"
		></button>

		<!-- Modal -->
		<div
			class="relative w-full max-w-lg rounded-xl bg-white shadow-2xl"
		>
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
				<h2 id="modal-title" class="text-lg font-semibold text-gray-900">Create New Table</h2>
				<button
					type="button"
					onclick={handleClose}
					class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
				>
					<X class="h-5 w-5" />
				</button>
			</div>

			<!-- Form -->
			<form onsubmit={handleSubmit} class="p-6">
				{#if error}
					<div class="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-600">
						{error}
					</div>
				{/if}

				<!-- Source Selection -->
				<div class="mb-6">
					<label class="mb-2 block text-sm font-medium text-gray-700">Start from</label>
					<div class="grid grid-cols-3 gap-3">
						{#each sourceOptions as option}
							<button
								type="button"
								class="flex flex-col items-center rounded-lg border-2 p-4 text-center transition-colors {source ===
								option.value
									? 'border-blue-500 bg-blue-50'
									: 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'}"
								onclick={() => (source = option.value)}
							>
								<option.icon
									class="mb-2 h-6 w-6 {source === option.value ? 'text-blue-600' : 'text-gray-400'}"
								/>
								<span
									class="text-sm font-medium {source === option.value
										? 'text-blue-600'
										: 'text-gray-700'}"
								>
									{option.label}
								</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Name -->
				<div class="mb-4">
					<label for="table-name" class="mb-1.5 block text-sm font-medium text-gray-700">
						Table Name <span class="text-red-500">*</span>
					</label>
					<input
						id="table-name"
						type="text"
						bind:value={name}
						placeholder="e.g., Customer Database"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						required
					/>
				</div>

				<!-- Description -->
				<div class="mb-6">
					<label for="table-description" class="mb-1.5 block text-sm font-medium text-gray-700">
						Description
					</label>
					<textarea
						id="table-description"
						bind:value={description}
						placeholder="What is this table for?"
						rows="2"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					></textarea>
				</div>

				<!-- Actions -->
				<div class="flex justify-end gap-3">
					<button
						type="button"
						onclick={handleClose}
						class="btn-pill btn-pill-ghost btn-pill-sm"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={loading || !name.trim()}
						class="btn-pill btn-pill-primary btn-pill-sm"
					>
						{#if loading}
							Creating...
						{:else}
							Create Table
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
