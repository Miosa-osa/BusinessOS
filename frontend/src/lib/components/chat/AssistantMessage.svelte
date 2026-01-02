<script lang="ts">
	import { fly, fade } from 'svelte/transition';

	import BlockRenderer from './BlockRenderer.svelte';
	import type { Block } from '$lib/api/conversations/types';

	interface Props {
		content: string;
		blocks?: Block[];
		timestamp?: string;
		isStreaming?: boolean;
		model?: string;
		onCopy?: () => void;
		onRegenerate?: () => void;
		onFeedback?: (type: 'good' | 'bad') => void;
	}

	let { content, blocks, timestamp, isStreaming = false, model, onCopy, onRegenerate, onFeedback }: Props = $props();

	let copied = $state(false);

	function formatTime(dateStr: string) {
		return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function handleCopy() {
		navigator.clipboard.writeText(content);
		copied = true;
		setTimeout(() => copied = false, 2000);
		onCopy?.();
	}

	// Basic markdown rendering (inline)
	function renderContent(text: string) {
		return text
			// Bold
			.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
			// Italic
			.replace(/\*(.*?)\*/g, '<em>$1</em>')
			// Inline code
			.replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1.5 py-0.5 rounded text-sm font-mono">$1</code>')
			// Headers
			.replace(/^### (.*$)/gm, '<h3 class="text-base font-semibold mt-4 mb-2">$1</h3>')
			.replace(/^## (.*$)/gm, '<h2 class="text-lg font-semibold mt-4 mb-2">$1</h2>')
			.replace(/^# (.*$)/gm, '<h1 class="text-xl font-bold mt-4 mb-2">$1</h1>')
			// Lists
			.replace(/^- (.*$)/gm, '<li class="ml-4 list-disc">$1</li>')
			.replace(/^\d+\. (.*$)/gm, '<li class="ml-4 list-decimal">$1</li>');
	}
</script>

<div class="flex justify-start" in:fly={{ y: 20, duration: 300 }}>
	<div class="max-w-[85%]">
		<div class="flex items-center gap-2 mb-2">
			<div class="w-7 h-7 rounded-lg bg-gray-100 flex items-center justify-center">
				<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
			</div>
			<span class="text-sm font-medium text-gray-700">Assistant</span>
			{#if model}
				<span class="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded">{model}</span>
			{/if}
		</div>

		<div class="bg-gray-50 border border-gray-100 rounded-2xl rounded-tl-md px-5 py-4">
			<div class="text-[15px] leading-relaxed text-gray-800 prose prose-sm max-w-none">
				{#if blocks && blocks.length > 0}
					<BlockRenderer {blocks} {isStreaming} />
				{:else}
					{@html renderContent(content)}
					{#if isStreaming}
						<span class="inline-block w-1.5 h-5 bg-gray-400 animate-pulse ml-0.5 rounded-sm align-middle"></span>
					{/if}
				{/if}
			</div>
		</div>

		{#if !isStreaming}
			<div class="flex items-center gap-3 mt-2 px-1">
				{#if timestamp}
					<span class="text-xs text-gray-400">{formatTime(timestamp)}</span>
				{/if}

				<div class="flex items-center gap-1">
					<button
						onclick={handleCopy}
						class="flex items-center gap-1 px-2 py-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors"
					>
						{#if copied}
							<svg class="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
							<span class="text-green-600">Copied</span>
						{:else}
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							<span>Copy</span>
						{/if}
					</button>

					{#if onRegenerate}
						<button
							onclick={onRegenerate}
							class="flex items-center gap-1 px-2 py-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors"
						>
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							<span>Regenerate</span>
						</button>
					{/if}

					{#if onFeedback}
						<div class="flex items-center gap-0.5 ml-2 border-l border-gray-200 pl-2">
							<button
								onclick={() => onFeedback?.('good')}
								class="p-1.5 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
								title="Good response"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
								</svg>
							</button>
							<button
								onclick={() => onFeedback?.('bad')}
								class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
								title="Bad response"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018a2 2 0 01.485.06l3.76.94m-7 10v5a2 2 0 002 2h.096c.5 0 .905-.405.905-.904 0-.715.211-1.413.608-2.008L17 13V4m-7 10h2m5-10h2a2 2 0 012 2v6a2 2 0 01-2 2h-2.5" />
								</svg>
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>
