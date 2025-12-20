<script lang="ts">
	import type { ClientListResponse, ClientStatus } from '$lib/api';
	import { statusColors, statusLabels } from '$lib/stores/clients';

	interface Props {
		clients: ClientListResponse[];
		onClientClick: (id: string) => void;
		onStatusChange: (id: string, status: ClientStatus) => void;
	}

	let { clients, onClientClick, onStatusChange }: Props = $props();

	const columns: { id: ClientStatus; label: string; color: string }[] = [
		{ id: 'lead', label: 'Leads', color: 'border-purple-300' },
		{ id: 'prospect', label: 'Prospects', color: 'border-blue-300' },
		{ id: 'active', label: 'Active', color: 'border-emerald-300' },
		{ id: 'inactive', label: 'Inactive', color: 'border-gray-300' },
		{ id: 'churned', label: 'Churned', color: 'border-red-300' }
	];

	function getClientsByStatus(status: ClientStatus): ClientListResponse[] {
		return clients.filter((c) => c.status === status);
	}

	function formatCurrency(value: number | null): string {
		if (value === null) return '-';
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0
		}).format(value);
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function getTotalValue(clients: ClientListResponse[]): number {
		return clients.reduce((sum, c) => sum + (c.lifetime_value || 0), 0);
	}

	// Drag and drop state
	let draggedClient = $state<ClientListResponse | null>(null);
	let dragOverColumn = $state<ClientStatus | null>(null);

	function handleDragStart(e: DragEvent, client: ClientListResponse) {
		draggedClient = client;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', client.id);
		}
	}

	function handleDragEnd() {
		draggedClient = null;
		dragOverColumn = null;
	}

	function handleDragOver(e: DragEvent, status: ClientStatus) {
		e.preventDefault();
		dragOverColumn = status;
	}

	function handleDragLeave() {
		dragOverColumn = null;
	}

	function handleDrop(e: DragEvent, status: ClientStatus) {
		e.preventDefault();
		if (draggedClient && draggedClient.status !== status) {
			onStatusChange(draggedClient.id, status);
		}
		draggedClient = null;
		dragOverColumn = null;
	}
</script>

<div class="flex-1 overflow-auto p-6">
	<div class="flex gap-4 h-full min-w-max">
		{#each columns as column}
			{@const columnClients = getClientsByStatus(column.id)}
			<div
				class="w-72 flex-shrink-0 flex flex-col bg-gray-50 rounded-xl border-t-4 {column.color} {dragOverColumn ===
				column.id
					? 'ring-2 ring-gray-400'
					: ''}"
				ondragover={(e) => handleDragOver(e, column.id)}
				ondragleave={handleDragLeave}
				ondrop={(e) => handleDrop(e, column.id)}
				role="region"
				aria-label="{column.label} column"
			>
				<!-- Column Header -->
				<div class="p-4 border-b border-gray-200">
					<div class="flex items-center justify-between">
						<h3 class="font-medium text-gray-900">{column.label}</h3>
						<span class="px-2 py-0.5 text-xs font-medium bg-gray-200 text-gray-700 rounded-full">
							{columnClients.length}
						</span>
					</div>
					<div class="text-sm text-gray-500 mt-1">
						{formatCurrency(getTotalValue(columnClients))} total
					</div>
				</div>

				<!-- Cards -->
				<div class="flex-1 overflow-auto p-3 space-y-3">
					{#each columnClients as client}
						<div
							class="bg-white rounded-lg border border-gray-200 p-3 cursor-pointer hover:shadow-md transition-shadow {draggedClient?.id ===
							client.id
								? 'opacity-50'
								: ''}"
							draggable="true"
							ondragstart={(e) => handleDragStart(e, client)}
							ondragend={handleDragEnd}
							onclick={() => onClientClick(client.id)}
							role="button"
							tabindex="0"
							onkeypress={(e) => e.key === 'Enter' && onClientClick(client.id)}
						>
							<!-- Client Header -->
							<div class="flex items-start gap-2">
								<div
									class="w-8 h-8 rounded-md bg-gray-100 flex items-center justify-center text-xs font-medium text-gray-600 flex-shrink-0"
								>
									{getInitials(client.name)}
								</div>
								<div class="min-w-0 flex-1">
									<h4 class="font-medium text-sm text-gray-900 truncate">{client.name}</h4>
									<span
										class="inline-flex items-center px-1 py-0.5 text-xs rounded {client.type ===
										'company'
											? 'bg-blue-50 text-blue-700'
											: 'bg-violet-50 text-violet-700'}"
									>
										{client.type === 'company' ? 'Company' : 'Individual'}
									</span>
								</div>
							</div>

							<!-- Value -->
							{#if client.lifetime_value}
								<div class="mt-2 text-sm font-medium text-gray-700">
									{formatCurrency(client.lifetime_value)}
								</div>
							{/if}

							<!-- Meta -->
							<div class="mt-2 flex items-center justify-between text-xs text-gray-500">
								<span>{client.deals_count} deals</span>
								<span>{client.contacts_count} contacts</span>
							</div>

							<!-- Tags -->
							{#if client.tags && client.tags.length > 0}
								<div class="mt-2 flex gap-1 flex-wrap">
									{#each client.tags.slice(0, 2) as tag}
										<span class="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
											{tag}
										</span>
									{/each}
								</div>
							{/if}
						</div>
					{/each}

					{#if columnClients.length === 0}
						<div class="text-center py-8 text-sm text-gray-400">No clients</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>
</div>
