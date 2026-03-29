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
	import { Pagination } from '$lib/components/ui';

	type ViewMode = 'directory' | 'orgchart' | 'capacity';

	/** Parse skills from backend — handles array, JSON string, base64-encoded JSON, or comma-separated */
	function parseSkills(raw: unknown): string[] {
		if (Array.isArray(raw)) return raw;
		if (typeof raw !== 'string' || !raw) return [];
		// Try JSON parse first (e.g. '["Go","Python"]')
		try {
			const parsed = JSON.parse(raw);
			if (Array.isArray(parsed)) return parsed;
		} catch { /* not JSON */ }
		// Try base64 decode then JSON parse
		try {
			const decoded = atob(raw);
			const parsed = JSON.parse(decoded);
			if (Array.isArray(parsed)) return parsed;
		} catch { /* not base64 JSON */ }
		// Fallback: comma-separated
		return raw.split(',').map(s => s.trim()).filter(Boolean);
	}

	// State
	let viewMode = $state<ViewMode>('directory');
	let searchQuery = $state('');
	let showAddModal = $state(false);
	let editingMember = $state<{
		id: string;
		name: string;
		email: string;
		role: string;
		managerId?: string;
		skills: string[];
		hourlyRate?: number;
	} | null>(null);
	// Pagination state
	let page = $state(1);
	let pageSize = $state(20);
	let total = $state(0);
	let showProfileSlideOver = $state(false);
	let selectedMember = $state<TeamMemberDetailResponse | null>(null);
	let loadingMember = $state(false);

	// Subscribe to team store using Svelte's auto-subscription ($store)
	let members = $derived($team.members);
	let loading = $derived($team.loading);
	let error = $derived($team.error);

	// Load members on mount
	onMount(() => {
		loadMembers();
	});

	// Load members with pagination
	async function loadMembers() {
		try {
			await team.loadMembers({ page, pageSize });
		} catch {
			// Backend unavailable — empty state will show
		}
		total = members.length > 0 ? (page - 1) * pageSize + members.length : 0;
	}

	// Handle page change
	function handlePageChange(newPage: number) {
		page = newPage;
		loadMembers();
		// Scroll to top of content
		window.scrollTo({ top: 0, behavior: 'smooth' });
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
			projects: [],
			activeProjects: m.active_projects
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
		try {
			const member = await team.loadMember(memberId);
			if (member) {
				selectedMember = member;
				showProfileSlideOver = true;
			}
		} catch (err) {
			console.error('Failed to load member profile:', err);
		} finally {
			loadingMember = false;
		}
	}

	async function handleAddMember(data: { email: string; role: string }) {
		try {
			await team.createMember({
				name: data.email.split('@')[0],
				email: data.email,
				role: data.role,
				skills: []
			});
			showAddModal = false;
			await loadMembers();
		} catch (err) {
			console.error('Failed to add member:', err);
		}
	}

	function handleCloseProfile() {
		showProfileSlideOver = false;
		selectedMember = null;
		team.clearCurrent();
	}

	function handleEditMember() {
		if (!selectedMember) return;
		editingMember = {
			id: selectedMember.id,
			name: selectedMember.name,
			email: selectedMember.email,
			role: selectedMember.role,
			managerId: selectedMember.manager_id || undefined,
			skills: parseSkills(selectedMember.skills),
			hourlyRate: selectedMember.hourly_rate || undefined
		};
		showAddModal = true;
	}

	async function handleUpdateMember(id: string, data: {
		name: string;
		email: string;
		role: string;
		managerId?: string;
		skills: string[];
		hourlyRate?: number;
	}) {
		try {
			await team.updateMember(id, {
				name: data.name,
				email: data.email,
				role: data.role,
				manager_id: data.managerId,
				skills: data.skills,
				hourly_rate: data.hourlyRate
			});
			showAddModal = false;
			editingMember = null;
			handleCloseProfile();
			await loadMembers();
		} catch (err) {
			console.error('Failed to update member:', err);
		}
	}

	async function handleDeleteMember() {
		if (!selectedMember) return;
		try {
			await team.deleteMember(selectedMember.id);
			handleCloseProfile();
			await loadMembers();
		} catch (err) {
			console.error('Failed to delete member:', err);
		}
	}

	// Transform selectedMember to match MemberProfileSlideOver expected format
	const profileMember = $derived(() => {
		if (!selectedMember) return null;
		const skills = parseSkills(selectedMember.skills);
		return {
			id: selectedMember.id,
			name: selectedMember.name,
			role: selectedMember.role ?? '',
			email: selectedMember.email ?? '',
			avatar: selectedMember.avatar_url || undefined,
			status: (selectedMember.status ?? 'available') as 'available' | 'busy' | 'overloaded' | 'ooo',
			activeProjects: selectedMember.active_projects ?? 0,
			openTasks: selectedMember.open_tasks ?? 0,
			capacity: selectedMember.capacity ?? 0,
			managerId: selectedMember.manager_id,
			joinedAt: selectedMember.joined_at,
			skills,
			projects: [],
			activity: (selectedMember.activities || []).map((a) => ({
				id: a.id,
				description: a.description,
				createdAt: a.created_at
			}))
		};
	});
</script>

<div class="td-page">
	<!-- Header -->
	<div class="td-page__header">
		<div>
			<h1 class="td-page__title">Team</h1>
			<p class="td-page__subtitle">Manage your team and see who's working on what</p>
		</div>
		<div class="td-page__actions">
			<button
				onclick={() => (showAddModal = true)}
				class="btn-cta"
				aria-label="Add team member"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
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

	<!-- Error State -->
	{#if error}
		<div class="td-page__error">
			<p class="td-page__error-text">{error}</p>
			<button
				onclick={() => loadMembers()}
				class="td-page__error-retry"
				aria-label="Retry loading"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && members.length === 0}
		<div class="td-page__center">
			<div class="td-page__center-content">
				<svg
					class="w-8 h-8 animate-spin"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					/>
				</svg>
				<p class="td-page__center-text">Loading team members...</p>
			</div>
		</div>
	{:else if members.length === 0 && !loading}
		<!-- Empty State -->
		<div class="td-page__center">
			<div class="td-page__center-content">
				<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
					/>
				</svg>
				<p class="td-page__empty-title">No team members yet</p>
				<p class="td-page__center-text">Add your first team member to get started</p>
				<button
					onclick={() => (showAddModal = true)}
					class="btn-cta"
					aria-label="Add your first member"
				>
					Add Member
				</button>
			</div>
		</div>
	{:else}
		<!-- Content Wrapper with Flex Layout -->
		<div class="flex-1 flex flex-col overflow-hidden">
			<!-- Content -->
			<div class="flex-1 overflow-auto">
				{#if viewMode === 'directory'}
					<TeamDirectoryView members={directoryMembers()} onMemberClick={handleMemberClick} />
				{:else if viewMode === 'orgchart'}
					<TeamOrgChartView members={orgMembers()} onMemberClick={handleMemberClick} />
				{:else if viewMode === 'capacity'}
					<TeamCapacityView members={capacityMembers()} onMemberClick={handleMemberClick} />
				{/if}
			</div>

			<!-- Pagination -->
			<Pagination
				{page}
				{pageSize}
				{total}
				onPageChange={handlePageChange}
			/>
		</div>
	{/if}
</div>

<!-- Add Member Modal (includes Invite tab) -->
<AddMemberModal
	bind:open={showAddModal}
	members={$team.members.map(m => ({ id: m.id, name: m.name, email: m.email, avatar: m.avatar, role: m.role }))}
	workspaceId={$currentWorkspace?.id}
	roles={$currentWorkspaceRoles}
	onCreate={handleAddMember}
	onClose={() => { editingMember = null; }}
/>

<!-- Member Profile Slide-over -->
<MemberProfileSlideOver
	bind:open={showProfileSlideOver}
	member={profileMember()}
	onClose={handleCloseProfile}
	onEdit={handleEditMember}
	onDelete={handleDeleteMember}
/>

<style>
	/* ── Team Page Layout ── */
	.td-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg);
	}

	/* ── Header ── */
	.td-page__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.5rem;
		border-bottom: 1px solid var(--dbd);
	}
	.td-page__title {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}
	.td-page__subtitle {
		font-size: 0.8125rem;
		color: var(--dt3);
		margin: 0.125rem 0 0;
	}
	.td-page__actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* ── Error State ── */
	.td-page__error {
		margin: 1rem 1.5rem 0;
		padding: 1rem;
		background: color-mix(in srgb, var(--bos-status-error) 10%, var(--dbg));
		border: 1px solid color-mix(in srgb, var(--bos-status-error) 30%, var(--dbd));
		border-radius: 8px;
	}
	.td-page__error-text {
		font-size: 0.875rem;
		color: var(--bos-status-error);
		margin: 0;
	}
	.td-page__error-retry {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		color: var(--bos-status-error);
		text-decoration: underline;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.td-page__error-retry:hover {
		opacity: 0.8;
	}

	/* ── Center / Loading / Empty ── */
	.td-page__center {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.td-page__center-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		color: var(--dt3);
	}
	.td-page__center-text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin: 0;
	}
	.td-page__empty-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}
</style>
