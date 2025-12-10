<script lang="ts">
	import { fly, fade } from 'svelte/transition';

	interface Props {
		content: string;
		timestamp: string;
		onEdit?: () => void;
		onCopy?: () => void;
		onDelete?: () => void;
	}

	let { content, timestamp, onEdit, onCopy, onDelete }: Props = $props();

	let showActions = $state(false);

	function formatTime(dateStr: string) {
		return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function handleCopy() {
		navigator.clipboard.writeText(content);
		onCopy?.();
	}
</script>

<div
	class="flex justify-end group"
	onmouseenter={() => showActions = true}
	onmouseleave={() => showActions = false}
	in:fly={{ y: 20, duration: 300 }}
>
	<div class="max-w-[75%] relative">
		{#if showActions && (onEdit || onCopy || onDelete)}
			<div
				class="absolute -top-8 right-0 flex items-center gap-1 bg-white border border-gray-200 rounded-lg shadow-sm px-1 py-0.5"
				in:fade={{ duration: 150 }}
			>
				{#if onEdit}
					<button
						onclick={onEdit}
						class="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors"
						title="Edit"
					>
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
						</svg>
					</button>
				{/if}
				<button
					onclick={handleCopy}
					class="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors"
					title="Copy"
				>
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
				</button>
				{#if onDelete}
					<button
						onclick={onDelete}
						class="p-1.5 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
						title="Delete"
					>
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
						</svg>
					</button>
				{/if}
			</div>
		{/if}

		<div class="bg-gray-900 text-white px-4 py-3 rounded-2xl rounded-tr-md">
			<p class="text-[15px] leading-relaxed whitespace-pre-wrap">{content}</p>
		</div>
		<div class="flex items-center justify-end gap-1.5 mt-1 px-1">
			<span class="text-xs text-gray-400">{formatTime(timestamp)}</span>
			<svg class="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
		</div>
	</div>
</div>
