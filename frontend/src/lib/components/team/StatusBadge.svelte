<script lang="ts">
	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface Props {
		status: Status;
		showLabel?: boolean;
		size?: 'sm' | 'md';
	}

	let { status, showLabel = true, size = 'sm' }: Props = $props();

	const statusConfig: Record<Status, { color: string; bg: string; label: string }> = {
		available: { color: 'bg-green-500', bg: 'bg-green-50 text-green-700', label: 'Available' },
		busy: { color: 'bg-yellow-500', bg: 'bg-yellow-50 text-yellow-700', label: 'Busy' },
		overloaded: { color: 'bg-red-500', bg: 'bg-red-50 text-red-700', label: 'Overloaded' },
		ooo: { color: 'bg-gray-500', bg: 'bg-gray-100 text-gray-600', label: 'Out of Office' }
	};

	const config = $derived(statusConfig[status]);
	const sizeClasses = size === 'sm' ? 'text-xs px-2 py-0.5' : 'text-sm px-2.5 py-1';
</script>

{#if showLabel}
	<span class="inline-flex items-center gap-1.5 rounded-full font-medium {config.bg} {sizeClasses}">
		<span class="w-1.5 h-1.5 rounded-full {config.color}"></span>
		{config.label}
	</span>
{:else}
	<span class="w-2.5 h-2.5 rounded-full {config.color}" title={config.label}></span>
{/if}
