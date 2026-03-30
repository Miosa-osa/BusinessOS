<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { team } from '$lib/stores/team';
	import { goto } from '$app/navigation';
	import type { TeamMemberDetailResponse } from '$lib/api';
	import { mockTeamMembers } from '$lib/mock-data';
	import StatusBadge from '$lib/components/team/StatusBadge.svelte';
	import CapacityBar from '$lib/components/team/CapacityBar.svelte';
	import { ArrowLeft, Mail, Calendar, Briefcase, CheckSquare, Clock } from 'lucide-svelte';

	let member = $state<TeamMemberDetailResponse | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const memberId = $derived($page.params.memberId);

	onMount(async () => {
		await loadMember();
	});

	async function loadMember() {
		loading = true;
		error = null;
		try {
			const data = await team.loadMember(memberId ?? '');
			if (data) {
				member = data;
			} else {
				// Try mock data fallback
				const mockMember = getMockMember(memberId ?? '');
				if (mockMember) {
					member = mockMember;
				} else {
					error = 'Member not found';
				}
			}
		} catch {
			// Backend unavailable — try mock data
			const mockMember = getMockMember(memberId ?? '');
			if (mockMember) {
				member = mockMember;
			} else {
				error = 'Failed to load member details';
			}
		} finally {
			loading = false;
		}
	}

	function getMockMember(id: string): TeamMemberDetailResponse | null {
		const m = mockTeamMembers.find((m) => m.id === id);
		if (!m) return null;
		return {
			id: m.id,
			name: m.name,
			email: m.email,
			role: m.role,
			avatar_url: m.avatar_url,
			status: m.status,
			capacity: m.capacity,
			manager_id: m.manager_id,
			skills: [],
			hourly_rate: null,
			joined_at: m.joined_at,
			created_at: m.joined_at,
			updated_at: m.joined_at,
			active_projects: m.active_projects,
			open_tasks: m.open_tasks,
			activities: []
		};
	}

	function getInitials(name: string) {
		return name
			.split(' ')
			.map((n) => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function formatDate(dateStr: string | undefined) {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<div class="min-h-full" style="background: var(--dbg)">
	<!-- Header -->
	<div class="border-b" style="background: var(--dbg2); border-color: var(--dbd)">
		<div class="max-w-5xl mx-auto px-6 py-4">
			<button
				onclick={() => goto('/team')}
				class="flex items-center gap-2 text-sm transition-colors mb-4"
				style="color: var(--dt2)"
			>
				<ArrowLeft class="w-4 h-4" />
				Back to Team
			</button>

			{#if loading}
				<div class="animate-pulse">
					<div class="flex items-center gap-6">
						<div class="w-24 h-24 rounded-full" style="background: var(--dbg3)"></div>
						<div class="space-y-3">
							<div class="h-8 w-48 rounded" style="background: var(--dbg3)"></div>
							<div class="h-4 w-32 rounded" style="background: var(--dbg3)"></div>
						</div>
					</div>
				</div>
			{:else if error}
				<div class="text-center py-12">
					<p class="mb-4" style="color: var(--bos-status-error)">{error}</p>
					<button
						onclick={loadMember}
						class="btn-pill btn-pill-primary btn-pill-sm"
					>
						Try Again
					</button>
				</div>
			{:else if member}
				<div class="flex items-start gap-6">
					<!-- Avatar -->
					{#if member.avatar_url}
						<img
							src={member.avatar_url}
							alt={member.name}
							class="w-24 h-24 rounded-full object-cover border-4 shadow-lg"
							style="border-color: var(--dbg2)"
						/>
					{:else}
						<div class="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-2xl font-bold border-4 shadow-lg" style="border-color: var(--dbg2)">
							{getInitials(member.name)}
						</div>
					{/if}

					<!-- Info -->
					<div class="flex-1">
						<div class="flex items-center gap-3 mb-1">
							<h1 class="text-2xl font-bold" style="color: var(--dt)">{member.name}</h1>
							<StatusBadge status={member.status as 'available' | 'busy' | 'overloaded' | 'ooo'} />
						</div>
						<p class="text-lg" style="color: var(--dt2)">{member.role}</p>
						<div class="flex items-center gap-4 mt-3 text-sm" style="color: var(--dt3)">
							<span class="flex items-center gap-1.5">
								<Mail class="w-4 h-4" />
								{member.email}
							</span>
							<span class="flex items-center gap-1.5">
								<Calendar class="w-4 h-4" />
								Joined {formatDate(member.joined_at)}
							</span>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if member && !loading}
		<div class="max-w-5xl mx-auto px-6 py-8">
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Stats Cards -->
				<div class="lg:col-span-2 grid grid-cols-2 gap-4">
					<div class="rounded-xl p-5 border" style="background: var(--dbg2); border-color: var(--dbd)">
						<div class="flex items-center gap-3 mb-3">
							<div class="w-10 h-10 rounded-lg flex items-center justify-center" style="background: var(--bos-status-info-bg); color: var(--bos-status-info)">
								<Briefcase class="w-5 h-5" />
							</div>
							<span class="text-sm font-medium" style="color: var(--dt2)">Active Projects</span>
						</div>
						<p class="text-3xl font-bold" style="color: var(--dt)">{member.active_projects}</p>
					</div>

					<div class="rounded-xl p-5 border" style="background: var(--dbg2); border-color: var(--dbd)">
						<div class="flex items-center gap-3 mb-3">
							<div class="w-10 h-10 rounded-lg flex items-center justify-center" style="background: var(--bos-status-success-bg); color: var(--bos-status-success)">
								<CheckSquare class="w-5 h-5" />
							</div>
							<span class="text-sm font-medium" style="color: var(--dt2)">Open Tasks</span>
						</div>
						<p class="text-3xl font-bold" style="color: var(--dt)">{member.open_tasks}</p>
					</div>

					<!-- Capacity -->
					<div class="col-span-2 rounded-xl p-5 border" style="background: var(--dbg2); border-color: var(--dbd)">
						<div class="flex items-center justify-between mb-4">
							<span class="text-sm font-medium" style="color: var(--dt2)">Current Capacity</span>
							<span class="text-sm font-semibold" style="color: var(--dt)">{member.capacity}%</span>
						</div>
						<CapacityBar capacity={member.capacity} size="lg" />
					</div>
				</div>

				<!-- Skills -->
				<div class="rounded-xl p-5 border" style="background: var(--dbg2); border-color: var(--dbd)">
					<h3 class="text-sm font-semibold mb-4" style="color: var(--dt)">Skills</h3>
					{#if member.skills && member.skills.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each member.skills as skill}
								<span class="px-3 py-1.5 text-sm rounded-full" style="background: var(--dbg3); color: var(--dt2)">
									{skill}
								</span>
							{/each}
						</div>
					{:else}
						<p class="text-sm" style="color: var(--dt3)">No skills listed</p>
					{/if}
				</div>

				<!-- Recent Activity -->
				<div class="lg:col-span-3 rounded-xl p-5 border" style="background: var(--dbg2); border-color: var(--dbd)">
					<h3 class="text-sm font-semibold mb-4 flex items-center gap-2" style="color: var(--dt)">
						<Clock class="w-4 h-4" />
						Recent Activity
					</h3>
					{#if member.activities && member.activities.length > 0}
						<div class="space-y-3">
							{#each member.activities.slice(0, 5) as activity}
								<div class="flex items-start gap-3 p-3 rounded-lg" style="background: var(--dbg3)">
									<div class="w-2 h-2 rounded-full mt-2 flex-shrink-0" style="background: var(--bos-status-info)"></div>
									<div class="flex-1 min-w-0">
										<p class="text-sm" style="color: var(--dt)">{activity.description}</p>
										<p class="text-xs mt-1" style="color: var(--dt3)">
											{formatDate(activity.created_at)}
										</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm" style="color: var(--dt3)">No recent activity</p>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
