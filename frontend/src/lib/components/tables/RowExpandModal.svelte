<script lang="ts">
	/**
	 * RowExpandModal - Slide-over panel for viewing/editing row details
	 * NocoDB-style expand record view with all fields in form layout
	 */
	import {
		X,
		ChevronLeft,
		ChevronRight,
		Copy,
		Trash2,
		ExternalLink,
		MoreHorizontal,
		Type,
		Hash,
		Calendar,
		CheckSquare,
		CircleDot,
		Link,
		Mail,
		Paperclip,
		User,
		DollarSign,
		Percent,
		Star,
		Timer,
		Phone,
		Search,
		Calculator,
		Sigma,
		Link2,
		QrCode,
		Barcode,
		MousePointer,
		Braces,
		Clock,
		AlignLeft
	} from 'lucide-svelte';
	import type { Column, Row, ColumnType } from '$lib/api/tables/types';
	import CellRenderer from './cells/CellRenderer.svelte';
	import type { ComponentType, SvelteComponent } from 'svelte';

	type IconComponent = ComponentType<SvelteComponent>;

	interface Props {
		isOpen: boolean;
		row: Row | null;
		columns: Column[];
		rowIndex: number;
		totalRows: number;
		onClose: () => void;
		onCellChange: (columnId: string, value: unknown) => void;
		onDelete: () => void;
		onDuplicate: () => void;
		onNavigate: (direction: 'prev' | 'next') => void;
	}

	let {
		isOpen,
		row,
		columns,
		rowIndex,
		totalRows,
		onClose,
		onCellChange,
		onDelete,
		onDuplicate,
		onNavigate
	}: Props = $props();

	let showMenu = $state(false);
	let editingField = $state<string | null>(null);

	// Get icon for column type
	function getColumnIcon(type: ColumnType): IconComponent {
		const icons: Record<ColumnType, IconComponent> = {
			text: Type as unknown as IconComponent,
			long_text: AlignLeft as unknown as IconComponent,
			number: Hash as unknown as IconComponent,
			single_select: CircleDot as unknown as IconComponent,
			multi_select: CheckSquare as unknown as IconComponent,
			date: Calendar as unknown as IconComponent,
			datetime: Clock as unknown as IconComponent,
			checkbox: CheckSquare as unknown as IconComponent,
			url: Link as unknown as IconComponent,
			email: Mail as unknown as IconComponent,
			attachment: Paperclip as unknown as IconComponent,
			user: User as unknown as IconComponent,
			currency: DollarSign as unknown as IconComponent,
			percent: Percent as unknown as IconComponent,
			rating: Star as unknown as IconComponent,
			duration: Timer as unknown as IconComponent,
			phone: Phone as unknown as IconComponent,
			lookup: Search as unknown as IconComponent,
			rollup: Calculator as unknown as IconComponent,
			formula: Sigma as unknown as IconComponent,
			link_to_record: Link2 as unknown as IconComponent,
			qr_code: QrCode as unknown as IconComponent,
			barcode: Barcode as unknown as IconComponent,
			button: MousePointer as unknown as IconComponent,
			json: Braces as unknown as IconComponent
		};
		return icons[type] || (Type as unknown as IconComponent);
	}

	// Get primary field value for title
	const primaryColumn = $derived(columns.find((c) => c.is_primary) || columns[0]);
	const title = $derived.by(() => {
		if (!row || !primaryColumn) return 'Untitled';
		const value = row.data[primaryColumn.id];
		return value ? String(value) : 'Untitled';
	});

	// Visible columns (excluding hidden)
	const visibleColumns = $derived(columns.filter((c) => !c.is_hidden));

	// Handle keyboard navigation
	function handleKeydown(e: KeyboardEvent) {
		if (!isOpen) return;

		if (e.key === 'Escape') {
			if (editingField) {
				editingField = null;
			} else {
				onClose();
			}
		} else if (e.key === 'ArrowLeft' && e.altKey) {
			e.preventDefault();
			onNavigate('prev');
		} else if (e.key === 'ArrowRight' && e.altKey) {
			e.preventDefault();
			onNavigate('next');
		}
	}

	function handleFieldClick(columnId: string) {
		editingField = columnId;
	}

	function handleFieldBlur() {
		editingField = null;
	}

	function handleValueChange(columnId: string, value: unknown) {
		onCellChange(columnId, value);
	}

	// Close menu when clicking outside
	function handleWindowClick() {
		if (showMenu) showMenu = false;
	}
</script>

<svelte:window on:keydown={handleKeydown} on:click={handleWindowClick} />

{#if isOpen && row}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-40 bg-black/30 transition-opacity"
		onclick={onClose}
		role="button"
		tabindex="-1"
		aria-label="Close modal"
	></div>

	<!-- Slide-over Panel -->
	<div
		class="fixed inset-y-0 right-0 z-50 flex w-full max-w-2xl flex-col bg-white shadow-2xl"
		role="dialog"
		aria-modal="true"
		aria-labelledby="row-expand-title"
	>
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
			<div class="flex items-center gap-3">
				<!-- Navigation -->
				<div class="flex items-center gap-1">
					<button
						type="button"
						class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 disabled:opacity-50"
						onclick={() => onNavigate('prev')}
						disabled={rowIndex <= 0}
						title="Previous row (Alt+←)"
					>
						<ChevronLeft class="h-5 w-5" />
					</button>
					<span class="text-sm text-gray-500">
						{rowIndex + 1} / {totalRows}
					</span>
					<button
						type="button"
						class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 disabled:opacity-50"
						onclick={() => onNavigate('next')}
						disabled={rowIndex >= totalRows - 1}
						title="Next row (Alt+→)"
					>
						<ChevronRight class="h-5 w-5" />
					</button>
				</div>

				<div class="h-6 w-px bg-gray-200"></div>

				<!-- Title -->
				<h2 id="row-expand-title" class="text-lg font-semibold text-gray-900 truncate max-w-md">
					{title}
				</h2>
			</div>

			<div class="flex items-center gap-2">
				<!-- Actions Menu -->
				<div class="relative">
					<button
						type="button"
						class="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
						onclick={(e) => {
							e.stopPropagation();
							showMenu = !showMenu;
						}}
					>
						<MoreHorizontal class="h-5 w-5" />
					</button>

					{#if showMenu}
						<div class="absolute right-0 top-full z-10 mt-1 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg">
							<button
								type="button"
								class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
								onclick={() => {
									onDuplicate();
									showMenu = false;
								}}
							>
								<Copy class="h-4 w-4" />
								Duplicate row
							</button>
							<button
								type="button"
								class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
								onclick={() => {
									navigator.clipboard.writeText(window.location.href + '?row=' + row.id);
									showMenu = false;
								}}
							>
								<ExternalLink class="h-4 w-4" />
								Copy link
							</button>
							<div class="my-1 border-t border-gray-200"></div>
							<button
								type="button"
								class="flex w-full items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50"
								onclick={() => {
									if (confirm('Delete this row?')) {
										onDelete();
									}
									showMenu = false;
								}}
							>
								<Trash2 class="h-4 w-4" />
								Delete row
							</button>
						</div>
					{/if}
				</div>

				<!-- Close button -->
				<button
					type="button"
					class="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
					onclick={onClose}
					title="Close (Esc)"
				>
					<X class="h-5 w-5" />
				</button>
			</div>
		</div>

		<!-- Content - Scrollable Fields -->
		<div class="flex-1 overflow-y-auto">
			<div class="divide-y divide-gray-100">
				{#each visibleColumns as column (column.id)}
					{@const isEditing = editingField === column.id}
					{@const value = row.data[column.id]}
					{@const ColumnIcon = getColumnIcon(column.type)}

					<div
						class="group px-6 py-4 hover:bg-gray-50 transition-colors"
						onclick={() => handleFieldClick(column.id)}
						onkeydown={(e) => e.key === 'Enter' && handleFieldClick(column.id)}
						role="button"
						tabindex="0"
					>
						<!-- Field Label -->
						<div class="mb-2 flex items-center gap-2">
							<svelte:component this={ColumnIcon} class="h-4 w-4 text-gray-400" />
							<span class="text-sm font-medium text-gray-600">{column.name}</span>
							{#if column.is_required}
								<span class="text-red-500">*</span>
							{/if}
							{#if column.is_primary}
								<span class="rounded bg-blue-100 px-1.5 py-0.5 text-xs font-medium text-blue-600">
									Primary
								</span>
							{/if}
						</div>

						<!-- Field Value -->
						<div
							class="min-h-[40px] rounded-lg border transition-colors {isEditing
								? 'border-blue-500 bg-white ring-2 ring-blue-500/20'
								: 'border-transparent bg-transparent group-hover:border-gray-200 group-hover:bg-white'}"
						>
							<div class="px-3 py-2">
								<CellRenderer
									type={column.type}
									{value}
									options={column.options}
									editing={isEditing}
									expanded={true}
									onChange={(newValue) => handleValueChange(column.id, newValue)}
									onBlur={handleFieldBlur}
								/>
							</div>
						</div>

						<!-- Empty state hint -->
						{#if !value && value !== false && value !== 0 && !isEditing}
							<p class="mt-1 text-xs text-gray-400">Click to add {column.name.toLowerCase()}</p>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<!-- Footer -->
		<div class="border-t border-gray-200 bg-gray-50 px-6 py-3">
			<div class="flex items-center justify-between text-xs text-gray-500">
				<div>
					Created: {new Date(row.created_at).toLocaleString()}
				</div>
				<div>
					Updated: {new Date(row.updated_at).toLocaleString()}
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Slide-in animation */
	@keyframes slideIn {
		from {
			transform: translateX(100%);
		}
		to {
			transform: translateX(0);
		}
	}

	div[role='dialog'] {
		animation: slideIn 0.2s ease-out;
	}
</style>
