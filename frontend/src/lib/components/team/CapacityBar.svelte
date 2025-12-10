<script lang="ts">
	import { fly } from 'svelte/transition';

	interface Props {
		capacity: number; // 0-100
		showPercentage?: boolean;
		size?: 'sm' | 'md' | 'lg';
		animated?: boolean;
	}

	let { capacity, showPercentage = true, size = 'md', animated = true }: Props = $props();

	const getColor = (value: number) => {
		if (value < 70) return 'bg-green-500';
		if (value < 90) return 'bg-yellow-500';
		return 'bg-red-500';
	};

	const heightClasses = {
		sm: 'h-1.5',
		md: 'h-2',
		lg: 'h-3'
	};

	const color = $derived(getColor(capacity));
	const clampedCapacity = $derived(Math.min(100, Math.max(0, capacity)));
</script>

<div class="flex items-center gap-3 w-full">
	<div class="flex-1 bg-gray-100 rounded-full overflow-hidden {heightClasses[size]}">
		{#if animated}
			<div
				class="{color} {heightClasses[size]} rounded-full transition-all duration-500 ease-out"
				style="width: {clampedCapacity}%"
				in:fly={{ x: -100, duration: 600 }}
			></div>
		{:else}
			<div
				class="{color} {heightClasses[size]} rounded-full"
				style="width: {clampedCapacity}%"
			></div>
		{/if}
	</div>
	{#if showPercentage}
		<span class="text-sm font-medium text-gray-700 min-w-[40px] text-right">
			{clampedCapacity}%
		</span>
	{/if}
</div>
