<script lang="ts">
	import { currentUserRoleContext, currentWorkspace } from '$lib/stores/workspaces';

	interface Props {
		showLabel?: boolean;
		size?: 'sm' | 'md';
		showTooltip?: boolean;
	}

	let { showLabel = true, size = 'sm', showTooltip = true }: Props = $props();

	// Role configuration with colors and labels
	const roleConfig: Record<string, { color: string; bg: string; icon: string }> = {
		owner: { color: 'bg-purple-500', bg: 'bg-purple-50 text-purple-700', icon: '👑' },
		admin: { color: 'bg-blue-500', bg: 'bg-blue-50 text-blue-700', icon: '⚡' },
		editor: { color: 'bg-green-500', bg: 'bg-green-50 text-green-700', icon: '✏️' },
		member: { color: 'bg-yellow-500', bg: 'bg-yellow-50 text-yellow-700', icon: '👤' },
		viewer: { color: 'bg-gray-500', bg: 'bg-gray-100 text-gray-600', icon: '👁️' },
		guest: { color: 'bg-gray-400', bg: 'bg-gray-50 text-gray-500', icon: '🔒' }
	};

	// Derive current role configuration
	const currentRole = $derived($currentUserRoleContext?.role_name || 'guest');
	const config = $derived(roleConfig[currentRole] || roleConfig.guest);
	const sizeClasses = size === 'sm' ? 'text-xs px-2 py-0.5' : 'text-sm px-2.5 py-1';

	// Build tooltip text
	const tooltipText = $derived(() => {
		if (!$currentUserRoleContext) return 'No role context';

		const permissions = Object.entries($currentUserRoleContext.permissions || {})
			.flatMap(([resource, perms]) =>
				Object.entries(perms)
					.filter(([_, value]) => value === true)
					.map(([perm]) => `${resource}.${perm}`)
			);

		return `Role: ${$currentUserRoleContext.role_display_name}\nLevel: ${$currentUserRoleContext.hierarchy_level}\nPermissions: ${permissions.length}`;
	});

	// Show/hide tooltip
	let showTooltipPopup = $state(false);
</script>

{#if $currentUserRoleContext}
	{#if showLabel}
		<div class="relative inline-flex">
			<button
				onmouseenter={() => showTooltip && (showTooltipPopup = true)}
				onmouseleave={() => showTooltipPopup = false}
				onclick={() => showTooltipPopup = !showTooltipPopup}
				class="btn-pill btn-pill-soft {sizeClasses} {config.bg} gap-1.5"
			>
				<span class="w-1.5 h-1.5 rounded-full {config.color}"></span>
				<span class="font-medium">{$currentUserRoleContext.role_display_name}</span>
				{#if $currentWorkspace}
					<span class="text-[10px] opacity-50 font-normal ml-0.5">in {$currentWorkspace.name}</span>
				{/if}
			</button>

			{#if showTooltip && showTooltipPopup}
				<div
					class="absolute top-full left-0 mt-2 w-64 bg-gray-900 text-white text-xs rounded-lg shadow-lg p-3 z-50 pointer-events-none"
				>
					<div class="font-semibold mb-2">{$currentUserRoleContext.role_display_name}</div>
					<div class="space-y-1 text-gray-300">
						<div>Hierarchy Level: {$currentUserRoleContext.hierarchy_level}</div>
						{#if $currentUserRoleContext.title}
							<div>Title: {$currentUserRoleContext.title}</div>
						{/if}
						{#if $currentUserRoleContext.department}
							<div>Department: {$currentUserRoleContext.department}</div>
						{/if}
					</div>

					<div class="mt-3 pt-3 border-t border-gray-700">
						<div class="font-semibold mb-2">Key Permissions</div>
						<div class="space-y-1 text-gray-300">
							{#each Object.entries($currentUserRoleContext.permissions || {}) as [resource, perms]}
								{#each Object.entries(perms) as [perm, value]}
									{#if value === true}
										<div class="flex items-center gap-1">
											<span class="w-1 h-1 rounded-full bg-green-400"></span>
											<span>{resource}.{perm}</span>
										</div>
									{/if}
								{/each}
							{/each}
						</div>
					</div>

					{#if $currentUserRoleContext.expertise_areas && $currentUserRoleContext.expertise_areas.length > 0}
						<div class="mt-3 pt-3 border-t border-gray-700">
							<div class="font-semibold mb-2">Expertise Areas</div>
							<div class="flex flex-wrap gap-1">
								{#each $currentUserRoleContext.expertise_areas as area}
									<span class="px-2 py-0.5 bg-gray-800 rounded text-[10px]">{area}</span>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{:else}
		<span
			class="w-2.5 h-2.5 rounded-full {config.color}"
			title={tooltipText()}
		></span>
	{/if}
{/if}
