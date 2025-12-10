<script lang="ts">
	import { fly } from 'svelte/transition';
	import StatusBadge from './StatusBadge.svelte';
	import CapacityBar from './CapacityBar.svelte';

	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface Props {
		id: string;
		name: string;
		role: string;
		email?: string;
		avatar?: string;
		status: Status;
		activeProjects: number;
		openTasks: number;
		capacity: number;
		onClick?: () => void;
	}

	let {
		id,
		name,
		role,
		email,
		avatar,
		status,
		activeProjects,
		openTasks,
		capacity,
		onClick
	}: Props = $props();

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
	class="group bg-white border border-gray-200 rounded-xl p-5 hover:shadow-md hover:border-gray-300 transition-all duration-200 cursor-pointer"
	onclick={onClick}
	role="button"
	tabindex="0"
	onkeydown={(e) => e.key === 'Enter' && onClick?.()}
>
	<!-- Avatar -->
	<div class="flex flex-col items-center text-center mb-4">
		{#if avatar}
			<img src={avatar} alt={name} class="w-16 h-16 rounded-full mb-3 object-cover" />
		{:else}
			<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-3">
				<span class="text-xl font-semibold text-gray-600">{getInitials(name)}</span>
			</div>
		{/if}

		<h3 class="font-semibold text-gray-900">{name}</h3>
		<p class="text-sm text-gray-500">{role}</p>
	</div>

	<!-- Divider -->
	<div class="border-t border-gray-100 my-4"></div>

	<!-- Status -->
	<div class="flex justify-center mb-4">
		<StatusBadge {status} />
	</div>

	<!-- Stats -->
	<div class="space-y-2 text-sm">
		<div class="flex items-center justify-between text-gray-600">
			<span class="flex items-center gap-1.5">
				<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
				Active projects
			</span>
			<span class="font-medium text-gray-900">{activeProjects}</span>
		</div>
		<div class="flex items-center justify-between text-gray-600">
			<span class="flex items-center gap-1.5">
				<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
				</svg>
				Open tasks
			</span>
			<span class="font-medium text-gray-900">{openTasks}</span>
		</div>
	</div>

	<!-- Capacity -->
	<div class="mt-4">
		<div class="flex items-center justify-between text-xs text-gray-500 mb-1.5">
			<span>Capacity</span>
		</div>
		<CapacityBar {capacity} size="sm" />
	</div>

	<!-- View Profile Button -->
	<button
		class="w-full mt-4 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors group-hover:bg-gray-100"
	>
		View Profile
	</button>
</div>
