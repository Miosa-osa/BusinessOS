<script lang="ts">
	import type { ViewMode } from '$lib/stores/clients';
	import type { ClientStatus, ClientType } from '$lib/api';

	interface Props {
		view: ViewMode;
		searchQuery: string;
		statusFilter: ClientStatus | null;
		typeFilter: ClientType | null;
		onViewChange: (view: ViewMode) => void;
		onSearchChange: (query: string) => void;
		onStatusChange: (status: ClientStatus | null) => void;
		onTypeChange: (type: ClientType | null) => void;
	}

	let {
		view = $bindable(),
		searchQuery = $bindable(),
		statusFilter,
		typeFilter,
		onViewChange,
		onSearchChange,
		onStatusChange,
		onTypeChange
	}: Props = $props();

	const views: { id: ViewMode; label: string; icon: string }[] = [
		{ id: 'table', label: 'Table', icon: 'M3 10h18M3 14h18M3 18h18M3 6h18' },
		{
			id: 'cards',
			label: 'Cards',
			icon: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z'
		},
		{
			id: 'kanban',
			label: 'Pipeline',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2'
		}
	];

	const statuses: { id: ClientStatus | ''; label: string }[] = [
		{ id: '', label: 'All Statuses' },
		{ id: 'lead', label: 'Leads' },
		{ id: 'prospect', label: 'Prospects' },
		{ id: 'active', label: 'Active' },
		{ id: 'inactive', label: 'Inactive' },
		{ id: 'churned', label: 'Churned' }
	];

	const types: { id: ClientType | ''; label: string }[] = [
		{ id: '', label: 'All Types' },
		{ id: 'company', label: 'Companies' },
		{ id: 'individual', label: 'Individuals' }
	];
</script>

<div class="px-6 py-3 border-b border-gray-200 bg-gray-50/50">
	<div class="flex items-center justify-between gap-4">
		<!-- Left: Search & Filters -->
		<div class="flex items-center gap-3 flex-1">
			<!-- Search -->
			<div class="relative flex-1 max-w-xs">
				<svg
					class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
					/>
				</svg>
				<input
					type="text"
					placeholder="Search clients..."
					bind:value={searchQuery}
					oninput={() => onSearchChange(searchQuery)}
					class="w-full pl-9 pr-4 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
				/>
			</div>

			<!-- Status Filter -->
			<select
				value={statusFilter || ''}
				onchange={(e) => onStatusChange((e.target as HTMLSelectElement).value as ClientStatus || null)}
				class="btn-pill btn-pill-secondary btn-pill-sm"
			>
				{#each statuses as status}
					<option value={status.id}>{status.label}</option>
				{/each}
			</select>

			<!-- Type Filter -->
			<select
				value={typeFilter || ''}
				onchange={(e) => onTypeChange((e.target as HTMLSelectElement).value as ClientType || null)}
				class="btn-pill btn-pill-secondary btn-pill-sm"
			>
				{#each types as type}
					<option value={type.id}>{type.label}</option>
				{/each}
			</select>
		</div>

		<!-- Right: View Switcher -->
		<div class="btn-pill-group">
			{#each views as v}
				<button
					onclick={() => {
						view = v.id;
						onViewChange(v.id);
					}}
					class="btn-pill btn-pill-sm {view === v.id ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={v.icon} />
					</svg>
					{v.label}
				</button>
			{/each}
		</div>
	</div>
</div>
