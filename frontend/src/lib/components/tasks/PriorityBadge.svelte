<script lang="ts">
	type Priority = 'critical' | 'high' | 'medium' | 'low';

	interface Props {
		priority: Priority | null | undefined;
		size?: 'sm' | 'md';
		showLabel?: boolean;
	}

	let { priority, size = 'sm', showLabel = true }: Props = $props();

	const priorityConfig = {
		critical: {
			bg: 'bg-red-100',
			text: 'text-red-700',
			dot: 'bg-red-500',
			label: 'Critical'
		},
		high: {
			bg: 'bg-orange-100',
			text: 'text-orange-700',
			dot: 'bg-orange-500',
			label: 'High'
		},
		medium: {
			bg: 'bg-yellow-100',
			text: 'text-yellow-700',
			dot: 'bg-yellow-500',
			label: 'Medium'
		},
		low: {
			bg: 'bg-gray-100',
			text: 'text-gray-600',
			dot: 'bg-gray-400',
			label: 'Low'
		}
	};

	// Default to 'medium' if priority is null/undefined
	const normalizedPriority = $derived(priority && priorityConfig[priority] ? priority : 'medium');
	const config = $derived(priorityConfig[normalizedPriority]);
	const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm';
</script>

<span class="inline-flex items-center gap-1.5 rounded-full font-medium {config.bg} {config.text} {sizeClasses}">
	<span class="w-1.5 h-1.5 rounded-full {config.dot}"></span>
	{#if showLabel}
		{config.label}
	{/if}
</span>
