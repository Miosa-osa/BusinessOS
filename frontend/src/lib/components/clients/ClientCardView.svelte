<script lang="ts">
	import type { ClientListResponse } from '$lib/api';
	import { statusColors, statusLabels } from '$lib/stores/clients';

	interface Props {
		clients: ClientListResponse[];
		onClientClick: (id: string) => void;
	}

	let { clients, onClientClick }: Props = $props();

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

	function getTimeAgo(dateStr: string | null): string {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

		if (diffDays === 0) return 'Today';
		if (diffDays === 1) return 'Yesterday';
		if (diffDays < 7) return `${diffDays} days ago`;
		if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
		if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
		return `${Math.floor(diffDays / 365)} years ago`;
	}
</script>

<div class="flex-1 overflow-auto p-6">
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
		{#each clients as client}
			<button
				onclick={() => onClientClick(client.id)}
				class="text-left bg-white border border-gray-200 rounded-xl p-4 hover:shadow-md hover:border-gray-300 transition-all cursor-pointer focus:outline-none focus:ring-2 focus:ring-gray-900 focus:ring-offset-2"
			>
				<!-- Header -->
				<div class="flex items-start gap-3">
					<div
						class="w-12 h-12 rounded-lg bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-base font-semibold text-gray-600 flex-shrink-0"
					>
						{getInitials(client.name)}
					</div>
					<div class="min-w-0 flex-1">
						<h3 class="font-medium text-gray-900 truncate">{client.name}</h3>
						<div class="flex items-center gap-2 mt-1">
							<span
								class="inline-flex items-center px-1.5 py-0.5 text-xs rounded {client.type ===
								'company'
									? 'bg-blue-50 text-blue-700'
									: 'bg-violet-50 text-violet-700'}"
							>
								{client.type === 'company' ? 'Company' : 'Individual'}
							</span>
							<span class="inline-flex items-center px-1.5 py-0.5 text-xs rounded {statusColors[client.status]}">
								{statusLabels[client.status]}
							</span>
						</div>
					</div>
				</div>

				<!-- Contact Info -->
				<div class="mt-4 space-y-1.5">
					{#if client.email}
						<div class="flex items-center gap-2 text-sm text-gray-600">
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
								/>
							</svg>
							<span class="truncate">{client.email}</span>
						</div>
					{/if}
					{#if client.phone}
						<div class="flex items-center gap-2 text-sm text-gray-600">
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
								/>
							</svg>
							<span>{client.phone}</span>
						</div>
					{/if}
				</div>

				<!-- Stats -->
				<div class="mt-4 pt-4 border-t border-gray-100">
					<div class="flex items-center justify-between">
						<div>
							<div class="text-lg font-semibold text-gray-900">
								{formatCurrency(client.lifetime_value)}
							</div>
							<div class="text-xs text-gray-500">Lifetime Value</div>
						</div>
						<div class="text-right">
							<div class="text-sm font-medium text-gray-700">{client.deals_count} deals</div>
							<div class="text-xs text-gray-500">{client.contacts_count} contacts</div>
						</div>
					</div>
				</div>

				<!-- Tags & Last Contact -->
				<div class="mt-3 flex items-center justify-between">
					<div class="flex gap-1 flex-wrap">
						{#if client.tags && client.tags.length > 0}
							{#each client.tags.slice(0, 2) as tag}
								<span class="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">{tag}</span>
							{/each}
							{#if client.tags.length > 2}
								<span class="text-xs text-gray-400">+{client.tags.length - 2}</span>
							{/if}
						{/if}
					</div>
					<div class="text-xs text-gray-400">
						{getTimeAgo(client.last_contacted_at)}
					</div>
				</div>
			</button>
		{/each}
	</div>
</div>
