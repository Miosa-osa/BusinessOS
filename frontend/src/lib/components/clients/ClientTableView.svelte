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

<div class="cr-table-container">
	<div class="cr-table-wrap">
		<table class="cr-table">
			<thead>
				<tr>
					<th class="cr-table__th"><span class="cr-table__th-label">Client</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Type</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Status</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Contact</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Last Contact</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Value</span></th>
					<th class="cr-table__th"><span class="cr-table__th-label">Deals</span></th>
				</tr>
			</thead>
			<tbody>
				{#each clients as client}
					<tr
						class="cr-table__row"
						onclick={() => onClientClick(client.id)}
					>
						<td class="cr-table__td cr-table__td--name">
							<div class="cr-logo cr-logo--sm" style="background: rgba(99,102,241,0.1); border-color: rgba(99,102,241,0.15);">
								<span class="cr-logo__initials" style="color: #6366f1">{getInitials(client.name)}</span>
							</div>
							<div>
								<div class="cr-table__name">{client.name}</div>
								{#if client.tags && client.tags.length > 0}
									<div class="cr-table__tags">
										{#each client.tags.slice(0, 2) as tag}
											<span class="cr-table__tag">{tag}</span>
										{/each}
										{#if client.tags.length > 2}
											<span class="cr-table__tag-more">+{client.tags.length - 2}</span>
										{/if}
									</div>
								{/if}
							</div>
						</td>
						<td class="cr-table__td">
							<span class="cr-type-pill cr-type-pill--{client.type}">
								{client.type === 'company' ? 'Company' : 'Individual'}
							</span>
						</td>
						<td class="cr-table__td">
							<select
								value={client.status}
								onclick={(e) => e.stopPropagation()}
								onchange={(e) => {
									e.stopPropagation();
									onStatusChange(client.id, (e.target as HTMLSelectElement).value as ClientStatus);
								}}
								class="cr-table__status-select cr-status-select--{client.status}"
								aria-label="Change status for {client.name}"
							>
								<option value="lead">Lead</option>
								<option value="prospect">Prospect</option>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
								<option value="churned">Churned</option>
							</select>
						</td>
						<td class="cr-table__td">
							<div class="cr-table__contact">{client.email || '-'}</div>
							<div class="cr-table__contact-sub">{client.phone || ''}</div>
						</td>
						<td class="cr-table__td cr-table__td--muted">
							{formatDate(client.last_contacted_at)}
						</td>
						<td class="cr-table__td">
							<div class="cr-table__value">{formatCurrency(client.lifetime_value)}</div>
							{#if client.active_deals_value > 0}
								<div class="cr-table__pipeline-value">{formatCurrency(client.active_deals_value)} in pipeline</div>
							{/if}
						</td>
						<td class="cr-table__td cr-table__td--muted">
							<span>{client.deals_count} deals</span>
							<span class="cr-table__separator">·</span>
							<span>{client.contacts_count} contacts</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<style>
	/* ─── Table (Foundation CRM pattern) ───────────────────────── */
	.cr-table-container {
		flex: 1;
		overflow: auto;
	}
	.cr-table-wrap {
		border-radius: 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.cr-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}
	.cr-table__th {
		padding: 0;
		text-align: left;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
		background: var(--dbg2, #f5f5f5);
		position: sticky;
		top: 0;
		z-index: 1;
	}
	.cr-table__th-label {
		display: block;
		padding: 9px 16px;
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--dt3, #888);
	}
	.cr-table__row {
		cursor: pointer;
		transition: background 0.1s;
	}
	.cr-table__row:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.cr-table__row:not(:last-child) .cr-table__td {
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.cr-table__td {
		padding: 10px 16px;
		color: var(--dt, #111);
		vertical-align: middle;
		white-space: nowrap;
	}
	.cr-table__td--name {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.cr-table__td--muted {
		color: var(--dt3, #888);
		font-size: 12px;
	}
	.cr-table__name {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.cr-table__tags {
		display: flex;
		gap: 3px;
		margin-top: 3px;
	}
	.cr-table__tag {
		padding: 1px 6px;
		font-size: 10px;
		border-radius: 9999px;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt3, #888);
		border: 1px solid var(--dbd, #e0e0e0);
	}
	.cr-table__tag-more {
		font-size: 10px;
		color: var(--dt4, #bbb);
	}
	.cr-table__contact {
		font-size: 13px;
		color: var(--dt, #111);
	}
	.cr-table__contact-sub {
		font-size: 11px;
		color: var(--dt3, #888);
		margin-top: 1px;
	}
	.cr-table__value {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.cr-table__pipeline-value {
		font-size: 11px;
		color: #16a34a;
		margin-top: 1px;
	}
	:global(.dark) .cr-table__pipeline-value { color: #4ade80; }
	.cr-table__separator {
		margin: 0 4px;
		color: var(--dbd, #e0e0e0);
	}

	/* Status Select (styled native) */
	.cr-table__status-select {
		padding: 3px 8px;
		font-size: 11px;
		font-weight: 600;
		border-radius: 9999px;
		border: none;
		cursor: pointer;
		outline: none;
		appearance: none;
		-webkit-appearance: none;
	}
	.cr-status-select--active { background: rgba(34, 197, 94, 0.12); color: #16a34a; }
	.cr-status-select--lead { background: rgba(107, 114, 128, 0.12); color: #6b7280; }
	.cr-status-select--prospect { background: rgba(59, 130, 246, 0.12); color: #2563eb; }
	.cr-status-select--inactive { background: rgba(156, 163, 175, 0.12); color: #6b7280; }
	.cr-status-select--churned { background: rgba(239, 68, 68, 0.12); color: #ef4444; }
	:global(.dark) .cr-status-select--active { background: rgba(34, 197, 94, 0.15); color: #4ade80; }
	:global(.dark) .cr-status-select--lead { background: rgba(107, 114, 128, 0.15); color: #9ca3af; }
	:global(.dark) .cr-status-select--prospect { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
	:global(.dark) .cr-status-select--inactive { background: rgba(156, 163, 175, 0.15); color: #9ca3af; }
	:global(.dark) .cr-status-select--churned { background: rgba(239, 68, 68, 0.15); color: #f87171; }

	/* Logo small variant */
	.cr-logo {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		border: 1.5px solid transparent;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.cr-logo--sm { width: 34px; height: 34px; border-radius: 8px; }
	.cr-logo__initials {
		font-size: 12px;
		font-weight: 800;
		letter-spacing: -0.01em;
		line-height: 1;
	}

	/* Type pills */
	.cr-type-pill {
		display: inline-flex;
		align-items: center;
		height: 20px;
		padding: 0 8px;
		border-radius: 9999px;
		font-size: 11px;
		font-weight: 600;
	}
	.cr-type-pill--company { background: rgba(99, 102, 241, 0.1); color: #6366f1; }
	.cr-type-pill--individual { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
	:global(.dark) .cr-type-pill--company { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
	:global(.dark) .cr-type-pill--individual { background: rgba(139, 92, 246, 0.15); color: #a78bfa; }
</style>
