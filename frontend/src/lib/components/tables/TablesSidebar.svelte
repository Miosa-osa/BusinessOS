<script lang="ts">
	/**
	 * TablesSidebar - NocoDB-style sidebar with bases and tables tree
	 * Features: Bases/folders, table tree, create new, favorites
	 */
	import {
		Plus,
		ChevronRight,
		ChevronDown,
		Folder,
		FolderOpen,
		Table2,
		Star,
		Upload,
		Database,
		Link,
		MoreHorizontal,
		Trash2,
		Edit3,
		Copy,
		Settings
	} from 'lucide-svelte';
	import type { TableListItem } from '$lib/api/tables/types';

	interface Base {
		id: string;
		name: string;
		icon: string;
		color: string;
		tables: TableListItem[];
		isExpanded: boolean;
	}

	interface Props {
		tables: TableListItem[];
		favorites: TableListItem[];
		selectedTableId?: string;
		onTableClick: (id: string) => void;
		onCreateTable: () => void;
		onCreateBase: () => void;
		onImport: () => void;
	}

	let {
		tables,
		favorites,
		selectedTableId,
		onTableClick,
		onCreateTable,
		onCreateBase,
		onImport
	}: Props = $props();

	// Group tables into bases (for now, we'll create virtual groups)
	let bases = $state<Base[]>([]);
	let expandedSections = $state({
		favorites: true,
		myData: true,
		shared: false,
		imports: true
	});

	// Context menu state
	let contextMenu = $state<{ x: number; y: number; tableId: string } | null>(null);

	// Organize tables by source
	const customTables = $derived(tables.filter((t) => t.source === 'custom'));
	const importedTables = $derived(tables.filter((t) => t.source === 'import'));
	const integrationTables = $derived(tables.filter((t) => t.source === 'integration'));

	function toggleSection(section: keyof typeof expandedSections) {
		expandedSections[section] = !expandedSections[section];
	}

	function handleContextMenu(e: MouseEvent, tableId: string) {
		e.preventDefault();
		contextMenu = { x: e.clientX, y: e.clientY, tableId };
	}

	function closeContextMenu() {
		contextMenu = null;
	}

	// Close context menu on click outside
	function handleWindowClick() {
		if (contextMenu) closeContextMenu();
	}
</script>

<svelte:window on:click={handleWindowClick} />

<div class="flex h-full w-64 flex-col border-r border-gray-200 bg-gray-50">
	<!-- Header -->
	<div class="flex items-center justify-between border-b border-gray-200 px-4 py-3">
		<div class="flex items-center gap-2">
			<Database class="h-5 w-5 text-blue-600" />
			<span class="font-semibold text-gray-900">Tables</span>
		</div>
		<div class="flex items-center gap-1">
			<button
				type="button"
				class="rounded-md p-1.5 text-gray-500 hover:bg-gray-200 hover:text-gray-700"
				onclick={onImport}
				title="Import data"
			>
				<Upload class="h-4 w-4" />
			</button>
			<button
				type="button"
				class="rounded-md p-1.5 text-gray-500 hover:bg-gray-200 hover:text-gray-700"
				onclick={onCreateTable}
				title="Create table"
			>
				<Plus class="h-4 w-4" />
			</button>
		</div>
	</div>

	<!-- Scrollable Content -->
	<div class="flex-1 overflow-y-auto">
		<!-- Quick Actions -->
		<div class="border-b border-gray-200 p-3">
			<button
				type="button"
				class="flex w-full items-center gap-2 rounded-lg border border-dashed border-gray-300 px-3 py-2 text-sm text-gray-600 hover:border-blue-400 hover:bg-blue-50 hover:text-blue-600"
				onclick={onCreateBase}
			>
				<Plus class="h-4 w-4" />
				<span>New Base</span>
			</button>
		</div>

		<!-- Favorites Section -->
		{#if favorites.length > 0}
			<div class="border-b border-gray-200">
				<button
					type="button"
					class="flex w-full items-center gap-2 px-4 py-2 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 hover:bg-gray-100"
					onclick={() => toggleSection('favorites')}
				>
					{#if expandedSections.favorites}
						<ChevronDown class="h-3 w-3" />
					{:else}
						<ChevronRight class="h-3 w-3" />
					{/if}
					<Star class="h-3 w-3" />
					Favorites
					<span class="ml-auto rounded-full bg-gray-200 px-1.5 py-0.5 text-xs text-gray-600">
						{favorites.length}
					</span>
				</button>
				{#if expandedSections.favorites}
					<div class="pb-2">
						{#each favorites as table}
							<button
								type="button"
								class="flex w-full items-center gap-2 px-4 py-1.5 text-left text-sm {selectedTableId ===
								table.id
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-100'}"
								onclick={() => onTableClick(table.id)}
								oncontextmenu={(e) => handleContextMenu(e, table.id)}
							>
								<Table2 class="h-4 w-4 shrink-0 text-gray-400" />
								<span class="truncate">{table.name}</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- My Data Section (Custom Tables) -->
		<div class="border-b border-gray-200">
			<button
				type="button"
				class="flex w-full items-center gap-2 px-4 py-2 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 hover:bg-gray-100"
				onclick={() => toggleSection('myData')}
			>
				{#if expandedSections.myData}
					<ChevronDown class="h-3 w-3" />
				{:else}
					<ChevronRight class="h-3 w-3" />
				{/if}
				<FolderOpen class="h-3 w-3" />
				My Tables
				<span class="ml-auto rounded-full bg-gray-200 px-1.5 py-0.5 text-xs text-gray-600">
					{customTables.length}
				</span>
			</button>
			{#if expandedSections.myData}
				<div class="pb-2">
					{#if customTables.length === 0}
						<div class="px-4 py-2 text-xs text-gray-400">No custom tables yet</div>
					{:else}
						{#each customTables as table}
							<button
								type="button"
								class="group flex w-full items-center gap-2 px-4 py-1.5 text-left text-sm {selectedTableId ===
								table.id
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-100'}"
								onclick={() => onTableClick(table.id)}
								oncontextmenu={(e) => handleContextMenu(e, table.id)}
							>
								<Table2 class="h-4 w-4 shrink-0 text-gray-400" />
								<span class="truncate flex-1">{table.name}</span>
								<span class="text-xs text-gray-400">{table.row_count}</span>
							</button>
						{/each}
					{/if}
				</div>
			{/if}
		</div>

		<!-- Imports Section -->
		<div class="border-b border-gray-200">
			<button
				type="button"
				class="flex w-full items-center gap-2 px-4 py-2 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 hover:bg-gray-100"
				onclick={() => toggleSection('imports')}
			>
				{#if expandedSections.imports}
					<ChevronDown class="h-3 w-3" />
				{:else}
					<ChevronRight class="h-3 w-3" />
				{/if}
				<Upload class="h-3 w-3" />
				Imported
				<span class="ml-auto rounded-full bg-gray-200 px-1.5 py-0.5 text-xs text-gray-600">
					{importedTables.length}
				</span>
			</button>
			{#if expandedSections.imports}
				<div class="pb-2">
					{#if importedTables.length === 0}
						<button
							type="button"
							class="flex w-full items-center gap-2 px-4 py-2 text-left text-xs text-gray-500 hover:bg-gray-100 hover:text-blue-600"
							onclick={onImport}
						>
							<Plus class="h-3 w-3" />
							Import CSV or Excel
						</button>
					{:else}
						{#each importedTables as table}
							<button
								type="button"
								class="group flex w-full items-center gap-2 px-4 py-1.5 text-left text-sm {selectedTableId ===
								table.id
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-100'}"
								onclick={() => onTableClick(table.id)}
							>
								<Upload class="h-4 w-4 shrink-0 text-orange-400" />
								<span class="truncate flex-1">{table.name}</span>
								<span class="text-xs text-gray-400">{table.row_count}</span>
							</button>
						{/each}
					{/if}
				</div>
			{/if}
		</div>

		<!-- Integrations Section -->
		<div>
			<button
				type="button"
				class="flex w-full items-center gap-2 px-4 py-2 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 hover:bg-gray-100"
				onclick={() => toggleSection('shared')}
			>
				{#if expandedSections.shared}
					<ChevronDown class="h-3 w-3" />
				{:else}
					<ChevronRight class="h-3 w-3" />
				{/if}
				<Link class="h-3 w-3" />
				Connected
				<span class="ml-auto rounded-full bg-gray-200 px-1.5 py-0.5 text-xs text-gray-600">
					{integrationTables.length}
				</span>
			</button>
			{#if expandedSections.shared}
				<div class="pb-2">
					{#if integrationTables.length === 0}
						<div class="px-4 py-2 text-xs text-gray-400">No connected sources</div>
					{:else}
						{#each integrationTables as table}
							<button
								type="button"
								class="group flex w-full items-center gap-2 px-4 py-1.5 text-left text-sm {selectedTableId ===
								table.id
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-100'}"
								onclick={() => onTableClick(table.id)}
							>
								<Database class="h-4 w-4 shrink-0 text-green-500" />
								<span class="truncate flex-1">{table.name}</span>
								<span class="text-xs text-gray-400">{table.row_count}</span>
							</button>
						{/each}
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Footer -->
	<div class="border-t border-gray-200 p-3">
		<button
			type="button"
			class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
		>
			<Settings class="h-4 w-4" />
			<span>Settings</span>
		</button>
	</div>
</div>

<!-- Context Menu -->
{#if contextMenu}
	<div
		class="fixed z-50 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
		style="left: {contextMenu.x}px; top: {contextMenu.y}px"
	>
		<button
			type="button"
			class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
		>
			<Edit3 class="h-4 w-4" />
			Rename
		</button>
		<button
			type="button"
			class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
		>
			<Copy class="h-4 w-4" />
			Duplicate
		</button>
		<button
			type="button"
			class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
		>
			<Star class="h-4 w-4" />
			Add to favorites
		</button>
		<div class="my-1 border-t border-gray-200"></div>
		<button
			type="button"
			class="flex w-full items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50"
		>
			<Trash2 class="h-4 w-4" />
			Delete
		</button>
	</div>
{/if}
