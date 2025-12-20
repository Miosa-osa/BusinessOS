<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { fly, fade } from 'svelte/transition';

	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		streaming?: boolean;
		contextName?: string;
		modelName?: string;
		onSend?: (message: string) => void;
		onStop?: () => void;
		onAttach?: () => void;
	}

	let {
		value = $bindable(''),
		placeholder = 'Type your message...',
		disabled = false,
		streaming = false,
		contextName = 'Default',
		modelName = 'Local LLM',
		onSend,
		onStop,
		onAttach
	}: Props = $props();

	let textareaRef: HTMLTextAreaElement | undefined = $state(undefined);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSend();
		}
	}

	function handleSend() {
		if (!value.trim() || disabled || streaming) return;
		onSend?.(value);
		value = '';
		// Reset textarea height
		if (textareaRef) {
			textareaRef.style.height = 'auto';
		}
	}

	function handleInput() {
		// Auto-resize textarea
		if (textareaRef) {
			textareaRef.style.height = 'auto';
			textareaRef.style.height = Math.min(textareaRef.scrollHeight, 160) + 'px';
		}
	}
</script>

<div class="border-t border-gray-100 bg-white">
	<div class="max-w-4xl mx-auto p-4">
		<!-- Input Container -->
		<div class="flex items-end gap-3">
			<!-- Attachment Button -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex-shrink-0 w-10 h-10 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-xl transition-colors"
					disabled={streaming}
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						class="z-50 min-w-[180px] bg-white border border-gray-200 rounded-xl shadow-lg p-1"
						sideOffset={8}
						transition={fly}
						transitionConfig={{ y: -10, duration: 150 }}
					>
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={onAttach}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
							</svg>
							Upload file
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							Upload image
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
							</svg>
							Paste from clipboard
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>

			<!-- Textarea -->
			<div class="flex-1 relative">
				<textarea
					bind:this={textareaRef}
					bind:value
					oninput={handleInput}
					onkeydown={handleKeydown}
					{placeholder}
					rows={1}
					disabled={disabled || streaming}
					class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-[15px] placeholder-gray-400 resize-none focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
					style="min-height: 48px; max-height: 160px;"
				></textarea>
			</div>

			<!-- Voice Button (optional) -->
			<!-- <button
				class="flex-shrink-0 w-10 h-10 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-xl transition-colors"
				disabled={streaming}
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
				</svg>
			</button> -->

			<!-- Send / Stop Button -->
			{#if streaming}
				<button
					onclick={onStop}
					class="flex-shrink-0 w-12 h-12 flex items-center justify-center bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors"
				>
					<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
						<rect x="6" y="6" width="12" height="12" rx="2" />
					</svg>
				</button>
			{:else}
				<button
					onclick={handleSend}
					disabled={!value.trim() || disabled}
					class="flex-shrink-0 w-12 h-12 flex items-center justify-center bg-gray-900 text-white rounded-xl hover:bg-gray-800 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
					</svg>
				</button>
			{/if}
		</div>

		<!-- Status Bar -->
		<div class="flex items-center justify-between mt-2 px-1 text-xs text-gray-400">
			<div class="flex items-center gap-3">
				<span class="flex items-center gap-1">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
					</svg>
					{contextName}
				</span>
				<span class="text-gray-300">|</span>
				<span class="flex items-center gap-1">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
					{modelName}
				</span>
			</div>
			<span class="hidden sm:block">
				<kbd class="px-1.5 py-0.5 bg-gray-100 rounded text-gray-500">Enter</kbd> to send
			</span>
		</div>
	</div>
</div>
