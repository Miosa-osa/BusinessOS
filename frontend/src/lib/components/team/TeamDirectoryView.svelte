<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import MemberCard from './MemberCard.svelte';

	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface TeamMember {
		id: string;
		name: string;
		role: string;
		email?: string;
		avatar?: string;
		status: Status;
		activeProjects: number;
		openTasks: number;
		capacity: number;
	}

	interface Props {
		members: TeamMember[];
		onMemberClick?: (memberId: string) => void;
	}

	let { members, onMemberClick }: Props = $props();
</script>

<div class="flex-1 overflow-y-auto p-6">
	{#if members.length === 0}
		<div class="flex flex-col items-center justify-center py-16" in:fade={{ duration: 200 }}>
			<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
				<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
				</svg>
			</div>
			<h3 class="text-lg font-medium text-gray-900 mb-1">No team members yet</h3>
			<p class="text-gray-500">Add your first team member to get started</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each members as member, i (member.id)}
				<div in:fly={{ y: 20, duration: 400, delay: i * 50 }}>
					<MemberCard
						{...member}
						onClick={() => onMemberClick?.(member.id)}
					/>
				</div>
			{/each}
		</div>
	{/if}
</div>
