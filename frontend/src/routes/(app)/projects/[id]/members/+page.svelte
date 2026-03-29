<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { ProjectMembersPanel } from '$lib/components/projects';
	import { checkProjectAccess } from '$lib/api/projects/members';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { useSession } from '$lib/auth-client';
	import { ArrowLeft, Loader2, AlertCircle } from 'lucide-svelte';
	import type { ProjectAccessInfo } from '$lib/api/projects/types';

	const session = useSession();
	const projectId = $derived($page.params.id);

	let accessInfo = $state<ProjectAccessInfo | null>(null);
	let loading = $state(true);
	let error = $state('');
	let currentUserId = $state('');

	onMount(async () => {
		await loadAccessInfo();
	});

	async function loadAccessInfo() {
		loading = true;
		error = '';

		try {
			const userId = $session.data?.user?.id ?? '';
			currentUserId = userId;

			// Check access - ensure projectId is defined
			if (!projectId) {
				error = 'Project ID is missing';
				return;
			}
			accessInfo = await checkProjectAccess(projectId, userId);

			if (!accessInfo.has_access) {
				error = 'You do not have access to this project';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load project access';
			console.error('Failed to load project access:', err);
		} finally {
			loading = false;
		}
	}

	function handleBack() {
		goto(`/projects/${projectId}`);
	}
</script>

<div class="min-h-screen prm-mem-page">
	<!-- Header -->
	<div class="prm-mem-header">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
			<div class="flex items-center gap-4">
				<button
					onclick={handleBack}
					class="btn-pill btn-pill-ghost btn-pill-icon"
					aria-label="Back to project"
				>
					<ArrowLeft class="w-5 h-5 prm-mem-icon" />
				</button>
				<div>
					<h1 class="text-2xl font-bold prm-mem-title">Project Members</h1>
					<p class="text-sm prm-mem-muted mt-1">
						Manage team access and permissions for this project
					</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		{#if loading}
			<div class="flex flex-col items-center justify-center py-20">
				<Loader2 class="w-10 h-10 prm-mem-accent animate-spin mb-4" />
				<p class="text-sm prm-mem-muted">Loading access information...</p>
			</div>
		{:else if error}
			<div class="flex items-start gap-3 p-6 prm-mem-error-box rounded-lg">
				<AlertCircle class="w-6 h-6 prm-mem-error-icon flex-shrink-0 mt-0.5" />
				<div>
					<h3 class="text-lg font-semibold prm-mem-error-title mb-1">Access Denied</h3>
					<p class="text-sm prm-mem-error-text">{error}</p>
					<button
						onclick={handleBack}
						class="btn-pill btn-pill-danger btn-pill-sm mt-4"
					>
						Go Back
					</button>
				</div>
			</div>
		{:else if accessInfo && accessInfo.has_access}
			<!-- Access Info Banner -->
			<div class="mb-6 p-4 prm-mem-info-box rounded-lg">
				<div class="flex items-start gap-3">
					<div
						class="flex-shrink-0 px-3 py-1 prm-mem-info-badge rounded-lg text-sm font-medium"
					>
						{accessInfo.role?.toUpperCase() || 'VIEWER'}
					</div>
					<div class="flex-1">
						<h3 class="text-sm font-semibold prm-mem-info-title mb-1">Your Role</h3>
						<p class="text-xs prm-mem-info-text">
							{#if accessInfo.can_delete}
								Full control - You can edit, delete, and manage members
							{:else if accessInfo.can_invite}
								You can edit the project and invite new members
							{:else if accessInfo.can_edit}
								You can edit the project content
							{:else}
								You have read-only access to this project
							{/if}
						</p>
					</div>
				</div>
			</div>

			<!-- Members Panel -->
			<div class="prm-mem-panel p-6">
				<ProjectMembersPanel
					projectId={projectId ?? ''}
					workspaceId={$currentWorkspace?.id ?? ''}
					{currentUserId}
					userRole={accessInfo.role || 'viewer'}
					canInvite={accessInfo.can_invite}
				/>
			</div>
		{/if}
	</div>
</div>

<style>
	.prm-mem-page { background: var(--dbg2, #f9fafb); }
	.prm-mem-header { background: var(--dbg, #fff); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-mem-title { color: var(--dt, #111); }
	.prm-mem-muted { color: var(--dt3, #6b7280); }
	.prm-mem-icon { color: var(--dt3, #6b7280); }
	.prm-mem-panel { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); box-shadow: var(--bos-shadow-1); }
	.prm-mem-accent { color: var(--bos-status-info); }
	.prm-mem-error-box { background: var(--bos-status-error-bg); border: 1px solid color-mix(in srgb, var(--bos-status-error) 25%, var(--dbd, #e5e7eb)); }
	.prm-mem-error-icon { color: var(--bos-status-error); }
	.prm-mem-error-title { color: var(--dt, #111); }
	.prm-mem-error-text { color: var(--bos-status-error-text); }
	.prm-mem-info-box { background: var(--bos-status-info-bg); border: 1px solid color-mix(in srgb, var(--bos-status-info) 25%, var(--dbd, #e5e7eb)); }
	.prm-mem-info-badge { background: color-mix(in srgb, var(--bos-status-info) 15%, var(--dbg, #fff)); color: var(--bos-status-info); }
	.prm-mem-info-title { color: var(--dt, #111); }
	.prm-mem-info-text { color: var(--dt2, #4b5563); }
</style>
