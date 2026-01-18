<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import StatusBadge from './StatusBadge.svelte';
	import CapacityBar from './CapacityBar.svelte';

	type Status = 'available' | 'busy' | 'overloaded' | 'ooo';

	interface Project {
		id: string;
		name: string;
		taskCount: number;
		overdueCount: number;
	}

	interface Activity {
		id: string;
		description: string;
		createdAt: string;
	}

	interface TeamMember {
		id: string;
		name: string;
		role: string;
		email: string;
		avatar?: string;
		status: Status;
		capacity: number;
		joinedAt: string;
		skills: string[];
		projects: Project[];
		activity: Activity[];
	}

	interface Props {
		open?: boolean;
		member?: TeamMember | null;
		onClose?: () => void;
		onEdit?: () => void;
		onAssignTask?: () => void;
	}

	let {
		open = $bindable(false),
		member = null,
		onClose,
		onEdit,
		onAssignTask
	}: Props = $props();

	function handleClose() {
		open = false;
		onClose?.();
	}

	function getInitials(name: string) {
		return name
			.split(' ')
			.map(n => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function formatRelativeTime(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);

		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString();
	}

	function formatJoinDate(dateStr: string) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
	}
</script>

{#if open && member}
	<!-- Overlay -->
	<div
		class="fixed inset-0 bg-black/30 z-40"
		transition:fade={{ duration: 200 }}
		onclick={handleClose}
		role="button"
		tabindex="-1"
		onkeydown={(e) => e.key === 'Escape' && handleClose()}
	></div>

	<!-- Slide-over Panel -->
	<div
		class="fixed right-0 top-0 bottom-0 w-full max-w-md bg-white shadow-xl z-50 flex flex-col"
		transition:fly={{ x: 400, duration: 300 }}
	>
		<!-- Header -->
		<div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
			<h2 class="text-lg font-semibold text-gray-900">Team Member</h2>
			<button
				onclick={handleClose}
				class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 transition-colors"
			>
				<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-y-auto">
			<div class="px-6 py-6">
				<!-- Profile Header -->
				<div class="flex flex-col items-center text-center mb-6">
					{#if member.avatar}
						<img src={member.avatar} alt={member.name} class="w-20 h-20 rounded-full mb-3 object-cover" />
					{:else}
						<div class="w-20 h-20 rounded-full bg-gray-100 flex items-center justify-center mb-3">
							<span class="text-2xl font-semibold text-gray-600">{getInitials(member.name)}</span>
						</div>
					{/if}
					<h3 class="text-xl font-semibold text-gray-900">{member.name}</h3>
					<p class="text-gray-500">{member.role}</p>
					<p class="text-sm text-gray-400">{member.email}</p>
				</div>

				<!-- Status & Since -->
				<div class="grid grid-cols-2 gap-3 mb-6">
					<div class="bg-gray-50 rounded-xl p-3 text-center">
						<p class="text-xs text-gray-500 uppercase mb-1">Status</p>
						<div class="flex justify-center">
							<StatusBadge status={member.status} />
						</div>
					</div>
					<div class="bg-gray-50 rounded-xl p-3 text-center">
						<p class="text-xs text-gray-500 uppercase mb-1">Since</p>
						<p class="font-medium text-gray-900">{formatJoinDate(member.joinedAt)}</p>
					</div>
				</div>

				<hr class="border-gray-200 mb-6" />

				<!-- Current Workload -->
				<div class="mb-6">
					<h4 class="text-sm font-medium text-gray-700 mb-3">Current Workload</h4>
					<CapacityBar capacity={member.capacity} size="lg" />
				</div>

				<!-- Active Projects -->
				<div class="mb-6">
					<h4 class="text-sm font-medium text-gray-700 mb-3">Active Projects ({member.projects.length})</h4>
					{#if member.projects.length > 0}
						<div class="space-y-2">
							{#each member.projects as project}
								<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
									<div class="flex items-center gap-2">
										<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
										<span class="text-sm font-medium text-gray-900">{project.name}</span>
									</div>
									<div class="text-xs text-gray-500">
										{project.taskCount} tasks
										{#if project.overdueCount > 0}
											<span class="text-red-500">• {project.overdueCount} overdue</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-400 bg-gray-50 rounded-lg p-3">No active projects</p>
					{/if}
				</div>

				<hr class="border-gray-200 mb-6" />

				<!-- Skills -->
				<div class="mb-6">
					<h4 class="text-sm font-medium text-gray-700 mb-3">Skills</h4>
					{#if member.skills.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each member.skills as skill}
								<span class="px-2.5 py-1 bg-gray-100 text-gray-700 text-sm rounded-lg">{skill}</span>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-400">No skills listed</p>
					{/if}
				</div>

				<hr class="border-gray-200 mb-6" />

				<!-- Recent Activity -->
				<div>
					<h4 class="text-sm font-medium text-gray-700 mb-3">Recent Activity</h4>
					{#if member.activity.length > 0}
						<div class="space-y-3">
							{#each member.activity.slice(0, 5) as item}
								<div class="flex items-start gap-2 text-sm">
									<div class="w-1.5 h-1.5 rounded-full bg-gray-400 mt-2 flex-shrink-0"></div>
									<div>
										<p class="text-gray-700">{item.description}</p>
										<p class="text-xs text-gray-400">{formatRelativeTime(item.createdAt)}</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-400">No recent activity</p>
					{/if}
				</div>
			</div>
		</div>

		<!-- Footer Actions -->
		<div class="flex items-center gap-3 px-6 py-4 border-t border-gray-100">
			<button
				onclick={onEdit}
				class="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
			>
				Edit Member
			</button>
			<button
				onclick={onAssignTask}
				class="flex-1 btn-pill btn-pill-primary"
			>
				Assign Task
			</button>
		</div>
	</div>
{/if}
