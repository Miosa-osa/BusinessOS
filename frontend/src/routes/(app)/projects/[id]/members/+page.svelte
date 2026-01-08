<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { ProjectMembersPanel } from '$lib/components/projects';
	import { checkProjectAccess } from '$lib/api/projects/members';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { ArrowLeft, Loader2, AlertCircle } from 'lucide-svelte';
	import type { ProjectAccessInfo } from '$lib/api/projects/types';

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
			// Get current user ID from auth/session (you'll need to implement this)
			// For now, using a placeholder - replace with actual auth context
			const userId = 'current-user-id'; // TODO: Get from auth store
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

<div class="min-h-screen bg-gray-50">
	<!-- Header -->
	<div class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
			<div class="flex items-center gap-4">
				<button
					onclick={handleBack}
					class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
					aria-label="Back to project"
				>
					<ArrowLeft class="w-5 h-5 text-gray-600" />
				</button>
				<div>
					<h1 class="text-2xl font-bold text-gray-900">Project Members</h1>
					<p class="text-sm text-gray-500 mt-1">
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
				<Loader2 class="w-10 h-10 text-blue-600 animate-spin mb-4" />
				<p class="text-sm text-gray-500">Loading access information...</p>
			</div>
		{:else if error}
			<div class="flex items-start gap-3 p-6 bg-red-50 border border-red-200 rounded-lg">
				<AlertCircle class="w-6 h-6 text-red-600 flex-shrink-0 mt-0.5" />
				<div>
					<h3 class="text-lg font-semibold text-red-900 mb-1">Access Denied</h3>
					<p class="text-sm text-red-800">{error}</p>
					<button
						onclick={handleBack}
						class="mt-4 px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
					>
						Go Back
					</button>
				</div>
			</div>
		{:else if accessInfo && accessInfo.has_access}
			<!-- Access Info Banner -->
			<div class="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
				<div class="flex items-start gap-3">
					<div
						class="flex-shrink-0 px-3 py-1 bg-blue-100 text-blue-700 rounded-lg text-sm font-medium"
					>
						{accessInfo.role?.toUpperCase() || 'VIEWER'}
					</div>
					<div class="flex-1">
						<h3 class="text-sm font-semibold text-blue-900 mb-1">Your Role</h3>
						<p class="text-xs text-blue-800">
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
			<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
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
