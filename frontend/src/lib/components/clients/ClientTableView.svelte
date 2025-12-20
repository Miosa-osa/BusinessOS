<script lang="ts">
	import type { ClientListResponse, ClientStatus } from '$lib/api';
	import { statusColors, statusLabels } from '$lib/stores/clients';

	interface Props {
		clients: ClientListResponse[];
		onClientClick: (id: string) => void;
		onStatusChange: (id: string, status: ClientStatus) => void;
	}

	let { clients, onClientClick, onStatusChange }: Props = $props();

	function formatCurrency(value: number | null): string {
		if (value === null) return '-';
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0
		}).format(value);
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}
</script>

<div class="flex-1 overflow-auto">
	<table class="w-full">
		<thead class="bg-gray-50 sticky top-0">
			<tr>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Client
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Type
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Status
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Contact
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Last Contact
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Value
				</th>
				<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
					Deals
				</th>
			</tr>
		</thead>
		<tbody class="bg-white divide-y divide-gray-200">
			{#each clients as client}
				<tr
					class="hover:bg-gray-50 cursor-pointer transition-colors"
					onclick={() => onClientClick(client.id)}
				>
					<td class="px-6 py-4 whitespace-nowrap">
						<div class="flex items-center gap-3">
							<div
								class="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center text-sm font-medium text-gray-600"
							>
								{getInitials(client.name)}
							</div>
							<div>
								<div class="text-sm font-medium text-gray-900">{client.name}</div>
								{#if client.tags && client.tags.length > 0}
									<div class="flex gap-1 mt-0.5">
										{#each client.tags.slice(0, 2) as tag}
											<span class="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
												{tag}
											</span>
										{/each}
										{#if client.tags.length > 2}
											<span class="text-xs text-gray-400">+{client.tags.length - 2}</span>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<span
							class="inline-flex items-center px-2 py-1 text-xs rounded-md {client.type ===
							'company'
								? 'bg-blue-50 text-blue-700'
								: 'bg-violet-50 text-violet-700'}"
						>
							{client.type === 'company' ? 'Company' : 'Individual'}
						</span>
					</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<select
							value={client.status}
							onclick={(e) => e.stopPropagation()}
							onchange={(e) => {
								e.stopPropagation();
								onStatusChange(client.id, (e.target as HTMLSelectElement).value as ClientStatus);
							}}
							class="px-2 py-1 text-xs rounded-md border-0 {statusColors[client.status]} cursor-pointer focus:ring-2 focus:ring-gray-900"
						>
							<option value="lead">Lead</option>
							<option value="prospect">Prospect</option>
							<option value="active">Active</option>
							<option value="inactive">Inactive</option>
							<option value="churned">Churned</option>
						</select>
					</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<div class="text-sm text-gray-900">{client.email || '-'}</div>
						<div class="text-xs text-gray-500">{client.phone || ''}</div>
					</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
						{formatDate(client.last_contacted_at)}
					</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<div class="text-sm font-medium text-gray-900">
							{formatCurrency(client.lifetime_value)}
						</div>
						{#if client.active_deals_value > 0}
							<div class="text-xs text-emerald-600">
								{formatCurrency(client.active_deals_value)} in pipeline
							</div>
						{/if}
					</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<div class="flex items-center gap-3 text-sm text-gray-500">
							<span title="Deals">{client.deals_count} deals</span>
							<span class="text-gray-300">|</span>
							<span title="Contacts">{client.contacts_count} contacts</span>
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
