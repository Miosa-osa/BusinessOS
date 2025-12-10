<script lang="ts">
	import { scale } from 'svelte/transition';
	import StatusBadge from './StatusBadge.svelte';

	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface Props {
		id: string;
		name: string;
		role: string;
		avatar?: string;
		status: Status;
		depth?: number;
		onClick?: () => void;
	}

	let { id, name, role, avatar, status, depth = 0, onClick }: Props = $props();

	function getInitials(name: string) {
		return name
			.split(' ')
			.map(n => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}
</script>

<div
	class="flex flex-col items-center"
	in:scale={{ duration: 300, delay: depth * 100 }}
>
	<button
		onclick={onClick}
		class="bg-white border border-gray-200 rounded-xl p-4 hover:shadow-md hover:border-gray-300 transition-all duration-200 min-w-[160px] text-center cursor-pointer"
	>
		<!-- Avatar -->
		{#if avatar}
			<img src={avatar} alt={name} class="w-12 h-12 rounded-full mx-auto mb-2 object-cover" />
		{:else}
			<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-2">
				<span class="text-sm font-semibold text-gray-600">{getInitials(name)}</span>
			</div>
		{/if}

		<h4 class="font-medium text-sm text-gray-900">{name}</h4>
		<p class="text-xs text-gray-500 mb-2">{role}</p>
		<div class="flex justify-center">
			<StatusBadge {status} size="sm" />
		</div>
	</button>
</div>
