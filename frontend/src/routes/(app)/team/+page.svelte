<script lang="ts">
	import { onMount } from 'svelte';
	import { team } from '$lib/stores/team';
	import { currentWorkspace, currentWorkspaceRoles } from '$lib/stores/workspaces';
	import type {
		TeamMemberListResponse,
		TeamMemberDetailResponse,
		TeamMemberStatus
	} from '$lib/api';
	import {
		TeamDirectoryView,
		TeamOrgChartView,
		TeamCapacityView,
		TeamViewSwitcher,
		MemberProfileSlideOver,
		AddMemberModal
	} from '$lib/components/team';
	import InviteMemberModal from '$lib/components/workspace/InviteMemberModal.svelte';
	import PendingInvitations from '$lib/components/team/PendingInvitations.svelte';

	type ViewMode = 'directory' | 'orgchart' | 'capacity';

	// State
	let viewMode = $state<ViewMode>('directory');
	let searchQuery = $state('');
	let showAddModal = $state(false);
	let showInviteModal = $state(false);
	let showProfileSlideOver = $state(false);
	let selectedMember = $state<TeamMemberDetailResponse | null>(null);
	let loadingMember = $state(false);

	// Subscribe to team store
	let members = $state<TeamMemberListResponse[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		const unsubscribe = team.subscribe((state) => {
			members = state.members;
			loading = state.loading;
			error = state.error;
		});
		return unsubscribe;
	});

	// Load members on mount
	onMount(() => {
		team.loadMembers();
	});

	// Open invite modal
	function openInviteModal() {
		if (!$currentWorkspace?.id) {
			alert('Please select a workspace first');
			return;
		}
		showInviteModal = true;
	}

	// Filtered members based on search
	const filteredMembers = $derived(() => {
		if (!searchQuery) return members;
		const query = searchQuery.toLowerCase();
		return members.filter(
			(m) =>
				m.name.toLowerCase().includes(query) ||
				m.role.toLowerCase().includes(query) ||
				m.email.toLowerCase().includes(query)
		);
	});

	// Transform for directory view
	const directoryMembers = $derived(() => {
		return filteredMembers().map((m) => ({
			id: m.id,
			name: m.name,
			role: m.role,
			email: m.email,
			avatar: m.avatar_url || undefined,
			status: m.status as 'available' | 'busy' | 'overloaded' | 'ooo',
			activeProjects: m.active_projects,
			openTasks: m.open_tasks,
			capacity: m.capacity
		}));
	});

	// Capacity view data format
	const capacityMembers = $derived(() => {
		return filteredMembers().map((m) => ({
			id: m.id,
			name: m.name,
			role: m.role,
			avatar: m.avatar_url || undefined,
			status: m.status as 'available' | 'busy' | 'overloaded' | 'ooo',
			capacity: m.capacity,
			projects: [] // Projects not yet loaded in list view
		}));
	});

	// Org chart data format
	const orgMembers = $derived(() => {
		return filteredMembers().map((m) => ({
			id: m.id,
			name: m.name,
			role: m.role,
			avatar: m.avatar_url || undefined,
			status: m.status as 'available' | 'busy' | 'overloaded' | 'ooo',
			managerId: m.manager_id
		}));
	});

	// Manager options for Add modal
	const managerOptions = $derived(() => {
		return members.map((m) => ({ id: m.id, name: m.name }));
	});

	async function handleMemberClick(memberId: string) {
		loadingMember = true;
		const member = await team.loadMember(memberId);
		if (member) {
			selectedMember = member;
			showProfileSlideOver = true;
		}
		loadingMember = false;
	}

	async function handleAddMember(data: {
		name: string;
		email: string;
		role: string;
		managerId?: string;
		skills: string[];
		hourlyRate?: number;
	}) {
		try {
			await team.createMember({
				name: data.name,
				email: data.email,
				role: data.role,
				manager_id: data.managerId,
				skills: data.skills,
				hourly_rate: data.hourlyRate
			});
			showAddModal = false;
		} catch (err) {
			console.error('Failed to add member:', err);
		}
	}

	function handleCloseProfile() {
		showProfileSlideOver = false;
		selectedMember = null;
		team.clearCurrent();
	}

	// Transform selectedMember to match MemberProfileSlideOver expected format
	const profileMember = $derived(() => {
		if (!selectedMember) return null;
		return {
			id: selectedMember.id,
			name: selectedMember.name,
			role: selectedMember.role,
			email: selectedMember.email,
			avatar: selectedMember.avatar_url || undefined,
			status: selectedMember.status as 'available' | 'busy' | 'overloaded' | 'ooo',
			activeProjects: selectedMember.active_projects,
			openTasks: selectedMember.open_tasks,
			capacity: selectedMember.capacity,
			managerId: selectedMember.manager_id,
			joinedAt: selectedMember.joined_at,
			skills: selectedMember.skills || [],
			projects: [], // TODO: Add projects relation
			activity: selectedMember.activities.map((a) => ({
				id: a.id,
				description: a.description,
				createdAt: a.created_at
			}))
		};
	});
</script>

<div class="flex flex-col h-full bg-white">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Team</h1>
			<p class="text-sm text-gray-500 mt-0.5">Manage your team and see who's working on what</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={openInviteModal}
				class="btn-pill btn-pill-secondary btn-pill-sm"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
				Invite
			</button>
			<button
				onclick={() => (showAddModal = true)}
				class="btn-pill btn-pill-primary btn-pill-sm"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add Member
			</button>
		</div>
	</div>

	<!-- View Switcher -->
	<TeamViewSwitcher
		bind:view={viewMode}
		bind:searchQuery
		onViewChange={(v) => (viewMode = v)}
		onSearchChange={(q) => (searchQuery = q)}
	/>

	<!-- Pending Invitations -->
	<PendingInvitations />

	<!-- Error State -->
	{#if error}
		<div class="mx-6 mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
			<p class="text-sm text-red-700">{error}</p>
			<button
				onclick={() => team.loadMembers()}
				class="btn-pill btn-pill-ghost btn-pill-xs mt-2"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && members.length === 0}
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
				<svg
					class="w-8 h-8 animate-spin"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					/>
				</svg>
				<p class="text-sm">Loading team members...</p>
			</div>
		</div>
	{:else if members.length === 0 && !loading}
		<!-- Empty State -->
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
				<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
					/>
				</svg>
				<p class="text-lg font-medium text-gray-900">No team members yet</p>
				<p class="text-sm">Add your first team member to get started</p>
				<button
					onclick={() => (showAddModal = true)}
					class="btn-pill btn-pill-primary btn-pill-sm mt-2"
				>
					Add Member
				</button>
			</div>
		</div>
	{:else}
		<!-- Content -->
		{#if viewMode === 'directory'}
			<TeamDirectoryView members={directoryMembers()} onMemberClick={handleMemberClick} />
		{:else if viewMode === 'orgchart'}
			<TeamOrgChartView members={orgMembers()} onMemberClick={handleMemberClick} />
		{:else if viewMode === 'capacity'}
			<TeamCapacityView members={capacityMembers()} onMemberClick={handleMemberClick} />
		{/if}
	{/if}
</div>

<!-- Add Member Modal -->
<AddMemberModal bind:open={showAddModal} managers={managerOptions()} onCreate={handleAddMember} />

<!-- Invite Member Modal -->
{#if showInviteModal && $currentWorkspace?.id}
	<InviteMemberModal
		workspaceId={$currentWorkspace.id}
		roles={$currentWorkspaceRoles}
		on:success={() => {
			showInviteModal = false;
		}}
		on:cancel={() => {
			showInviteModal = false;
		}}
	/>
{/if}

<!-- Member Profile Slide-over -->
<MemberProfileSlideOver
	bind:open={showProfileSlideOver}
	member={profileMember()}
	onClose={handleCloseProfile}
/>
