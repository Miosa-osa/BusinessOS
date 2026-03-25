<script lang="ts">
	import { onMount } from 'svelte';

	interface AuditEvent {
		event_id: string;
		sequence_number: number;
		event_type: string;
		event_category: string;
		timestamp: string;
		severity: string;
		user_id: string | null;
		resource_type: string | null;
		resource_id: string | null;
		payload: Record<string, any>;
		pii_detected: boolean;
		legal_hold: boolean;
		retention_expires_at: string | null;
	}

	let events: AuditEvent[] = [];
	let filteredEvents: AuditEvent[] = [];
	let selectedEvent: AuditEvent | null = null;
	let showDetailModal = false;
	let isLoading = false;
	let error: string | null = null;

	let filterEventType = '';
	let filterSeverity = '';
	let filterFromDate = '';
	let filterToDate = '';

	let currentPage = 1;
	const pageSize = 25;
	let totalPages = 1;

	async function loadAuditLogs() {
		isLoading = true;
		error = null;

		try {
			const params = new URLSearchParams();
			if (filterEventType) params.append('event_type', filterEventType);
			if (filterFromDate) params.append('from_date', filterFromDate);
			if (filterToDate) params.append('to_date', filterToDate);
			params.append('limit', '1000');

			const response = await fetch(`/api/audit/logs?${params.toString()}`);
			if (!response.ok) throw new Error('Failed to load audit logs');

			const data = await response.json();
			events = data.events || [];
			filterEvents();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			isLoading = false;
		}
	}

	function filterEvents() {
		let filtered = [...events];

		if (filterSeverity) {
			filtered = filtered.filter(e => e.severity === filterSeverity);
		}

		totalPages = Math.ceil(filtered.length / pageSize);
		const start = (currentPage - 1) * pageSize;
		const end = start + pageSize;
		filteredEvents = filtered.slice(start, end);
	}

	function showEventDetail(event: AuditEvent) {
		selectedEvent = event;
		showDetailModal = true;
	}

	function closeDetailModal() {
		selectedEvent = null;
		showDetailModal = false;
	}

	async function exportAsCSV() {
		try {
			const params = new URLSearchParams();
			if (filterEventType) params.append('event_type', filterEventType);
			if (filterFromDate) params.append('from_date', filterFromDate);
			if (filterToDate) params.append('to_date', filterToDate);
			params.append('format', 'csv');

			window.location.href = `/api/audit/export?${params.toString()}`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Export failed';
		}
	}

	async function verifyChainIntegrity() {
		isLoading = true;
		error = null;

		try {
			const response = await fetch('/api/audit/verify', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					from_sequence: 1,
					to_sequence: 999999
				})
			});

			if (!response.ok) throw new Error('Verification failed');

			const result = await response.json();
			if (result.is_valid) {
				alert(`✓ Chain integrity verified (${result.verified_entries} entries)`);
			} else {
				alert(`✗ Chain integrity FAILED!\nIssues:\n${result.issues.join('\n')}`);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Verification failed';
		} finally {
			isLoading = false;
		}
	}

	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'critical':
				return 'text-red-700 bg-red-50';
			case 'warning':
				return 'text-yellow-700 bg-yellow-50';
			case 'info':
				return 'text-blue-700 bg-blue-50';
			default:
				return 'text-gray-700 bg-gray-50';
		}
	}

	function getCategoryColor(category: string): string {
		switch (category) {
			case 'Security':
				return 'bg-red-100 text-red-800';
			case 'ProcessMining':
				return 'bg-blue-100 text-blue-800';
			case 'Compliance':
				return 'bg-green-100 text-green-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	onMount(() => {
		loadAuditLogs();
	});
</script>

<div class="container mx-auto px-4 py-8">
	<h1 class="text-3xl font-bold mb-6">Audit Log Viewer</h1>

	{#if error}
		<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
			{error}
		</div>
	{/if}

	<div class="bg-white p-6 rounded-lg shadow mb-6">
		<h2 class="text-xl font-semibold mb-4">Filters</h2>
		<div class="grid grid-cols-1 md:grid-cols-5 gap-4">
			<div>
				<label class="block text-sm font-medium mb-2">Event Type</label>
				<input
					type="text"
					bind:value={filterEventType}
					placeholder="e.g., model_discovered"
					class="w-full px-3 py-2 border border-gray-300 rounded-md"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium mb-2">Severity</label>
				<select bind:value={filterSeverity} class="w-full px-3 py-2 border border-gray-300 rounded-md">
					<option value="">All</option>
					<option value="info">Info</option>
					<option value="warning">Warning</option>
					<option value="critical">Critical</option>
				</select>
			</div>
			<div>
				<label class="block text-sm font-medium mb-2">From Date</label>
				<input
					type="date"
					bind:value={filterFromDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-md"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium mb-2">To Date</label>
				<input
					type="date"
					bind:value={filterToDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-md"
				/>
			</div>
			<div class="flex items-end">
				<button
					on:click={loadAuditLogs}
					disabled={isLoading}
					class="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
				>
					{isLoading ? 'Loading...' : 'Search'}
				</button>
			</div>
		</div>
	</div>

	<div class="flex gap-2 mb-6">
		<button
			on:click={exportAsCSV}
			class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
		>
			Export CSV
		</button>
		<button
			on:click={verifyChainIntegrity}
			disabled={isLoading}
			class="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 disabled:opacity-50"
		>
			Verify Chain
		</button>
	</div>

	<div class="bg-white rounded-lg shadow overflow-hidden">
		<table class="w-full">
			<thead class="bg-gray-100 border-b">
				<tr>
					<th class="px-6 py-3 text-left text-sm font-semibold">Timestamp</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">Event Type</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">Category</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">Severity</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">PII</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">Hold</th>
					<th class="px-6 py-3 text-left text-sm font-semibold">Actions</th>
				</tr>
			</thead>
			<tbody>
				{#if filteredEvents.length === 0}
					<tr>
						<td colspan="7" class="px-6 py-4 text-center text-gray-500">
							{isLoading ? 'Loading...' : 'No events found'}
						</td>
					</tr>
				{:else}
					{#each filteredEvents as event (event.event_id)}
						<tr class="border-b hover:bg-gray-50">
							<td class="px-6 py-3 text-sm">{new Date(event.timestamp).toLocaleString()}</td>
							<td class="px-6 py-3 text-sm font-medium">{event.event_type}</td>
							<td class="px-6 py-3 text-sm">
								<span class="px-2 py-1 rounded text-xs font-medium {getCategoryColor(event.event_category)}">
									{event.event_category}
								</span>
							</td>
							<td class="px-6 py-3 text-sm">
								<span class="px-2 py-1 rounded {getSeverityColor(event.severity)}">
									{event.severity}
								</span>
							</td>
							<td class="px-6 py-3 text-sm">
								{event.pii_detected ? '📋 PII' : '-'}
							</td>
							<td class="px-6 py-3 text-sm">
								{event.legal_hold ? '🔒 HOLD' : '-'}
							</td>
							<td class="px-6 py-3 text-sm">
								<button
									on:click={() => showEventDetail(event)}
									class="px-3 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
								>
									View
								</button>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<div class="mt-6 flex justify-center gap-2">
			<button
				on:click={() => (currentPage = Math.max(1, currentPage - 1))}
				disabled={currentPage === 1}
				class="px-3 py-2 border rounded disabled:opacity-50"
			>
				Previous
			</button>
			<span class="px-4 py-2">Page {currentPage} of {totalPages}</span>
			<button
				on:click={() => (currentPage = Math.min(totalPages, currentPage + 1))}
				disabled={currentPage === totalPages}
				class="px-3 py-2 border rounded disabled:opacity-50"
			>
				Next
			</button>
		</div>
	{/if}

	{#if showDetailModal && selectedEvent}
		<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
			<div class="bg-white rounded-lg shadow-lg max-w-2xl w-full mx-4">
				<div class="p-6">
					<div class="flex justify-between items-center mb-4">
						<h2 class="text-2xl font-bold">Event Details</h2>
						<button on:click={closeDetailModal} class="text-2xl">&times;</button>
					</div>

					<div class="grid grid-cols-2 gap-4 mb-6">
						<div>
							<p class="text-sm text-gray-600">Sequence</p>
							<p class="font-mono">{selectedEvent.sequence_number}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Type</p>
							<p class="font-mono">{selectedEvent.event_type}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Category</p>
							<p class="font-mono">{selectedEvent.event_category}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Severity</p>
							<p class="font-mono">{selectedEvent.severity}</p>
						</div>
					</div>

					<div class="mb-6">
						<p class="text-sm text-gray-600 mb-2">Payload</p>
						<pre class="bg-gray-100 p-4 rounded text-xs overflow-x-auto">
{JSON.stringify(selectedEvent.payload, null, 2)}
</pre>
					</div>

					<button
						on:click={closeDetailModal}
						class="px-4 py-2 bg-gray-600 text-white rounded-md"
					>
						Close
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	:global(body) {
		background-color: #f5f5f5;
	}
</style>
