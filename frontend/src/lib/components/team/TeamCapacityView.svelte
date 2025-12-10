<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import CapacityBar from './CapacityBar.svelte';
	import StatusBadge from './StatusBadge.svelte';

	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface TeamMember {
		id: string;
		name: string;
		role: string;
		avatar?: string;
		status: Status;
		capacity: number;
		projects: string[];
	}

	interface Props {
		members: TeamMember[];
		onMemberClick?: (memberId: string) => void;
	}

	let { members, onMemberClick }: Props = $props();

	const summary = $derived(() => {
		const overloaded = members.filter(m => m.capacity >= 90).length;
		const atCapacity = members.filter(m => m.capacity >= 70 && m.capacity < 90).length;
		const available = members.filter(m => m.capacity < 70).length;
		return { overloaded, atCapacity, available };
	});

	// Get current week range
	const weekRange = $derived(() => {
		const now = new Date();
		const startOfWeek = new Date(now);
		startOfWeek.setDate(now.getDate() - now.getDay());
		const endOfWeek = new Date(startOfWeek);
		endOfWeek.setDate(startOfWeek.getDate() + 6);

		const formatDate = (d: Date) => d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
		return `Week of ${formatDate(startOfWeek)}-${formatDate(endOfWeek).split(' ')[1]}`;
	});

	function getInitials(name: string) {
		return name
			.split(' ')
			.map(n => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}
</script>

<div class="flex-1 overflow-y-auto p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">Team Capacity Overview</h2>
			<p class="text-sm text-gray-500">{weekRange()}</p>
		</div>
	</div>

	{#if members.length === 0}
		<div class="flex flex-col items-center justify-center py-16" in:fade={{ duration: 200 }}>
			<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
				<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
			</div>
			<h3 class="text-lg font-medium text-gray-900 mb-1">No capacity data</h3>
			<p class="text-gray-500">Add team members to see capacity overview</p>
		</div>
	{:else}
		<!-- Capacity List -->
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
			{#each members as member, i (member.id)}
				<button
					onclick={() => onMemberClick?.(member.id)}
					class="w-full flex items-center gap-4 px-4 py-4 hover:bg-gray-50 transition-colors text-left
						{i < members.length - 1 ? 'border-b border-gray-100' : ''}"
					in:fly={{ x: -100, duration: 400, delay: i * 50 }}
				>
					<!-- Avatar & Info -->
					<div class="flex items-center gap-3 min-w-[200px]">
						{#if member.avatar}
							<img src={member.avatar} alt={member.name} class="w-10 h-10 rounded-full object-cover" />
						{:else}
							<div class="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
								<span class="text-sm font-medium text-gray-600">{getInitials(member.name)}</span>
							</div>
						{/if}
						<div>
							<p class="font-medium text-gray-900">{member.name}</p>
							<p class="text-sm text-gray-500">{member.role}</p>
						</div>
					</div>

					<!-- Capacity Bar -->
					<div class="flex-1">
						<CapacityBar capacity={member.capacity} size="md" />
					</div>

					<!-- Status -->
					<div class="flex items-center gap-3 min-w-[120px] justify-end">
						<StatusBadge status={member.status} size="sm" />
					</div>
				</button>

				<!-- Projects (collapsed row) -->
				{#if member.projects.length > 0}
					<div class="px-4 pb-3 -mt-1 ml-[52px] {i < members.length - 1 ? 'border-b border-gray-100' : ''}">
						<p class="text-xs text-gray-400 truncate">
							{member.projects.join(', ')}
						</p>
					</div>
				{/if}
			{/each}
		</div>

		<!-- Summary -->
		<div class="mt-6 p-4 bg-gray-50 rounded-xl">
			<p class="text-sm text-gray-600">
				<span class="font-medium">Summary:</span>
				{#if summary().overloaded > 0}
					<span class="text-red-600">{summary().overloaded} overloaded</span>,
				{/if}
				{#if summary().atCapacity > 0}
					<span class="text-yellow-600">{summary().atCapacity} at capacity</span>,
				{/if}
				{#if summary().available > 0}
					<span class="text-green-600">{summary().available} available</span>
				{/if}
			</p>
		</div>
	{/if}
</div>
