<script lang="ts">
	import { fly } from 'svelte/transition';

	interface Props {
		onAction?: (action: string) => void;
	}

	let { onAction }: Props = $props();

	const actions = [
		{
			id: 'new-chat',
			label: 'New Chat',
			shortcut: '⌘N'
		},
		{
			id: 'new-project',
			label: 'New Project',
			shortcut: '⌘P'
		},
		{
			id: 'new-task',
			label: 'Add Task',
			shortcut: '⌘T'
		},
		{
			id: 'daily-log',
			label: 'Daily Log',
			shortcut: '⌘L'
		}
	];

	let hoveredId = $state<string | null>(null);
</script>

{#snippet chatIcon()}
	<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
	</svg>
{/snippet}

{#snippet projectIcon()}
	<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
	</svg>
{/snippet}

{#snippet taskIcon()}
	<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
	</svg>
{/snippet}

{#snippet logIcon()}
	<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
	</svg>
{/snippet}

<div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center gap-2 mb-4">
		<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-sm">
			<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
			</svg>
		</div>
		<h2 class="text-base font-semibold text-gray-900">Quick Actions</h2>
	</div>

	<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
		{#each actions as action, index (action.id)}
			<button
				onclick={() => onAction?.(action.id)}
				onmouseenter={() => (hoveredId = action.id)}
				onmouseleave={() => (hoveredId = null)}
				class="group flex flex-col items-center justify-center gap-2 p-4 rounded-xl bg-gradient-to-br from-gray-50 to-gray-100/50 hover:from-gray-100 hover:to-gray-50 border border-gray-100 hover:border-gray-200 transition-all duration-200 {hoveredId ===
				action.id
					? 'scale-[1.02] shadow-md border-gray-300'
					: 'shadow-sm'}"
				in:fly={{ y: 10, duration: 300, delay: index * 50 }}
			>
				<span class="text-gray-600 group-hover:text-gray-900 group-hover:scale-110 transition-all duration-200">
					{#if action.id === 'new-chat'}
						{@render chatIcon()}
					{:else if action.id === 'new-project'}
						{@render projectIcon()}
					{:else if action.id === 'new-task'}
						{@render taskIcon()}
					{:else if action.id === 'daily-log'}
						{@render logIcon()}
					{/if}
				</span>
				<span class="text-sm font-medium text-gray-700">{action.label}</span>
				<span class="text-xs text-gray-400 hidden sm:block">{action.shortcut}</span>
			</button>
		{/each}
	</div>
</div>
