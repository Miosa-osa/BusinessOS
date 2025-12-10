<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { fly, fade } from 'svelte/transition';

	interface Props {
		id: string;
		title: string;
		preview?: string;
		timestamp: string;
		projectName?: string;
		pinned?: boolean;
		active?: boolean;
		messageCount?: number;
		onclick?: () => void;
		onRename?: () => void;
		onPin?: () => void;
		onLinkProject?: () => void;
		onExport?: () => void;
		onArchive?: () => void;
		onDelete?: () => void;
	}

	let {
		id,
		title,
		preview,
		timestamp,
		projectName,
		pinned = false,
		active = false,
		messageCount = 0,
		onclick,
		onRename,
		onPin,
		onLinkProject,
		onExport,
		onArchive,
		onDelete
	}: Props = $props();

	let menuOpen = $state(false);

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
</script>

<div
	class="group relative px-2 py-2.5 rounded-xl cursor-pointer transition-all duration-200
		{active ? 'bg-gray-900 text-white' : 'hover:bg-gray-100'}"
>
	<button
		onclick={onclick}
		class="w-full text-left"
	>
		<div class="flex items-start gap-2">
			{#if pinned}
				<svg class="w-3.5 h-3.5 flex-shrink-0 mt-1 {active ? 'text-gray-400' : 'text-gray-400'}" fill="currentColor" viewBox="0 0 20 20">
					<path d="M9.828.722a.5.5 0 01.354 0l3.5 1.5a.5.5 0 01.24.673L12.5 5.5l1.5 1.5v3.5l-3.5 3.5-3.5-3.5V7L5.5 5.5l-.922-2.605a.5.5 0 01.24-.673l3.5-1.5z"/>
				</svg>
			{/if}
			<div class="flex-1 min-w-0">
				<p class="font-medium text-sm truncate {active ? 'text-white' : 'text-gray-900'}">
					{title}
				</p>
				{#if preview}
					<p class="text-xs truncate mt-0.5 {active ? 'text-gray-300' : 'text-gray-500'}">
						{preview}
					</p>
				{/if}
				<div class="flex items-center gap-2 mt-1">
					<span class="text-xs {active ? 'text-gray-400' : 'text-gray-400'}">
						{formatRelativeTime(timestamp)}
					</span>
					{#if projectName}
						<span class="text-xs px-1.5 py-0.5 rounded {active ? 'bg-gray-700 text-gray-300' : 'bg-gray-100 text-gray-500'}">
							{projectName}
						</span>
					{/if}
				</div>
			</div>
		</div>
	</button>

	<!-- Menu Button -->
	<div class="absolute right-2 top-2 {menuOpen || active ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'} transition-opacity">
		<DropdownMenu.Root bind:open={menuOpen}>
			<DropdownMenu.Trigger
				class="p-1.5 rounded-lg transition-colors {active ? 'hover:bg-gray-700 text-gray-400' : 'hover:bg-gray-200 text-gray-400'}"
				onclick={(e: MouseEvent) => e.stopPropagation()}
			>
				<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 min-w-[180px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
					sideOffset={4}
					transition={fly}
					transitionConfig={{ y: -10, duration: 150 }}
				>
					{#if onRename}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onRename}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Rename
						</DropdownMenu.Item>
					{/if}
					{#if onPin}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onPin}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
							</svg>
							{pinned ? 'Unpin' : 'Pin'}
						</DropdownMenu.Item>
					{/if}
					{#if onLinkProject}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onLinkProject}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							Link to project
						</DropdownMenu.Item>
					{/if}
					{#if onExport}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onExport}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
							</svg>
							Export
						</DropdownMenu.Item>
					{/if}
					{#if onArchive}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onArchive}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
							</svg>
							Archive
						</DropdownMenu.Item>
					{/if}
					{#if onDelete}
						<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg cursor-pointer transition-colors"
							onclick={onDelete}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>
	</div>
</div>
