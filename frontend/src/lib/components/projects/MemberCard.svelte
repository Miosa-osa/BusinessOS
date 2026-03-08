<script lang="ts">
	import type { ProjectMember, ProjectRole } from '$lib/api/projects/types';
	import RoleSelector from './RoleSelector.svelte';
	import { Shield, Users, Eye, Edit3, Trash2, MoreVertical } from 'lucide-svelte';
	import { DropdownMenu } from 'bits-ui';

	interface Props {
		member: ProjectMember;
		canEdit?: boolean;
		canRemove?: boolean;
		currentUserId?: string;
		onRoleChange?: (memberId: string, newRole: ProjectRole) => void;
		onRemove?: (memberId: string) => void;
	}

	let {
		member,
		canEdit = false,
		canRemove = false,
		currentUserId = '',
		onRoleChange,
		onRemove
	}: Props = $props();

	function getInitials(name: string): string {
		if (!name) return '?';
		return name
			.split(' ')
			.map((n) => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function getRoleIcon(role: ProjectRole) {
		switch (role) {
			case 'lead':
				return Shield;
			case 'contributor':
				return Edit3;
			case 'reviewer':
				return Users;
			case 'viewer':
				return Eye;
		}
	}

	function getRoleColor(role: ProjectRole): { bg: string; text: string; border: string } {
		switch (role) {
			case 'lead':
				return { bg: 'rgba(147,51,234,0.12)', text: '#7c3aed', border: 'rgba(147,51,234,0.25)' };
			case 'contributor':
				return { bg: 'rgba(37,99,235,0.12)', text: '#2563eb', border: 'rgba(37,99,235,0.25)' };
			case 'reviewer':
				return { bg: 'rgba(22,163,74,0.12)', text: '#16a34a', border: 'rgba(22,163,74,0.25)' };
			case 'viewer':
				return { bg: 'var(--dbg2, #f5f5f5)', text: 'var(--dt2, #555)', border: 'var(--dbd, #e0e0e0)' };
		}
	}

	function getRoleLabel(role: ProjectRole): string {
		switch (role) {
			case 'lead':
				return 'Project Lead';
			case 'contributor':
				return 'Contributor';
			case 'reviewer':
				return 'Reviewer';
			case 'viewer':
				return 'Viewer';
		}
	}

	function getPermissionsList(member: ProjectMember): string[] {
		const permissions: string[] = [];
		if (member.can_edit) permissions.push('Edit');
		if (member.can_delete) permissions.push('Delete');
		if (member.can_invite) permissions.push('Invite');
		if (permissions.length === 0) permissions.push('Read-only');
		return permissions;
	}

	function handleRoleChange(newRole: ProjectRole) {
		onRoleChange?.(member.id, newRole);
	}

	function handleRemove() {
		if (confirm('Are you sure you want to remove this member from the project?')) {
			onRemove?.(member.id);
		}
	}

	const isCurrentUser = $derived(member.user_id === currentUserId);
	const RoleIcon = $derived(getRoleIcon(member.role));
	const permissions = $derived(getPermissionsList(member));
</script>

<div class="prm-member-card">
	<div class="flex items-start gap-4">
		<!-- Avatar -->
		<div class="flex-shrink-0">
			{#if member.user_avatar}
				<img
					src={member.user_avatar}
					alt={member.user_name || 'User'}
					class="w-12 h-12 rounded-full object-cover"
				/>
			{:else}
				<div
					class="w-12 h-12 rounded-full prm-member-card__avatar-gradient flex items-center justify-center"
				>
					<span class="text-white font-semibold text-sm">
						{getInitials(member.user_name || member.user_id)}
					</span>
				</div>
			{/if}
		</div>

		<!-- Member Info -->
		<div class="flex-1 min-w-0">
			<div class="flex items-start justify-between gap-2">
				<div class="flex-1 min-w-0">
					<div class="flex items-center gap-2">
						<h3 class="prm-member-card__name">
							{member.user_name || member.user_id}
						</h3>
						{#if isCurrentUser}
							<span class="px-2 py-0.5 text-xs font-medium prm-member-card__you-badge rounded-full">
								You
							</span>
						{/if}
					</div>
					{#if member.user_email}
						<p class="prm-member-card__email">{member.user_email}</p>
					{/if}
				</div>

				<!-- Actions Menu -->
				{#if (canEdit || canRemove) && !isCurrentUser}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="prm-member-card__action-btn"
							aria-label="Member actions"
						>
							<MoreVertical class="w-4 h-4" />
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="prm-member-card__dropdown z-50 min-w-[160px] rounded-lg shadow-lg p-1 animate-in fade-in-0 zoom-in-95"
								sideOffset={4}
							>
								{#if canRemove}
									<DropdownMenu.Item
										class="flex items-center gap-2 px-3 py-2 text-sm prm-member-card__danger-item rounded-md cursor-pointer"
										onclick={handleRemove}
									>
										<Trash2 class="w-4 h-4" />
										<span>Remove member</span>
									</DropdownMenu.Item>
								{/if}
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>
				{/if}
			</div>

			<!-- Role Badge and Selector -->
			<div class="mt-3">
				{#if canEdit && !isCurrentUser}
					<RoleSelector value={member.role} onChange={handleRoleChange} />
				{:else}
					{@const rc = getRoleColor(member.role)}
					<div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border" style="background:{rc.bg}; color:{rc.text}; border-color:{rc.border}">
						<svelte:component this={RoleIcon} class="w-3.5 h-3.5" />
						<span class="text-xs font-medium">{getRoleLabel(member.role)}</span>
					</div>
				{/if}
			</div>

			<!-- Permissions -->
			<div class="mt-3 flex flex-wrap gap-1.5">
				{#each permissions as permission}
					<span class="prm-member-card__perm">
						{permission}
					</span>
				{/each}
			</div>

			<!-- Member Since -->
			<div class="mt-2 text-xs prm-member-card__meta">
				Member since {new Date(member.assigned_at).toLocaleDateString()}
			</div>
		</div>
	</div>
</div>

<style>
	.prm-member-card {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.75rem;
		padding: 1rem;
		transition: box-shadow 0.2s;
	}
	.prm-member-card:hover {
		box-shadow: 0 4px 12px rgba(0,0,0,0.08);
	}
	.prm-member-card__name {
		font-weight: 600;
		color: var(--dt, #111);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.prm-member-card__email {
		font-size: 0.875rem;
		color: var(--dt2, #555);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.prm-member-card__action-btn {
		padding: 0.375rem;
		border-radius: 0.5rem;
		color: var(--dt2, #555);
		transition: background 0.15s;
	}
	.prm-member-card__action-btn:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.prm-member-card__dropdown {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
	}
	.prm-member-card__perm {
		padding: 0.125rem 0.5rem;
		font-size: 0.75rem;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt2, #555);
		border-radius: 0.25rem;
		border: 1px solid var(--dbd, #e0e0e0);
	}
	.prm-member-card__meta {
		color: var(--dt3, #888);
	}
	.prm-member-card__you-badge { background: color-mix(in srgb, #3b82f6 12%, var(--dbg)); color: #3b82f6; }
	.prm-member-card__danger-item { color: #ef4444; }
	.prm-member-card__danger-item:hover { background: color-mix(in srgb, #ef4444 10%, var(--dbg)); }
	.prm-member-card__avatar-gradient { background: linear-gradient(135deg, #3b82f6, #9333ea); }
</style>
