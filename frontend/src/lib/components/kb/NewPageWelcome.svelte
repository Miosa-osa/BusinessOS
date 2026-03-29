<script lang="ts">
	interface Template {
		id: string;
		name: string;
		description: string;
		icon: string;
		disabled?: boolean;
	}

	interface Props {
		onCreateBlank: () => void;
		onAskAI?: () => void;
		onSelectTemplate?: (templateId: string) => void;
	}

	let { onCreateBlank, onAskAI, onSelectTemplate }: Props = $props();

	// Handle Enter key to create blank page immediately
	function handleKeydown(e: KeyboardEvent) {
		// Only trigger if no input is focused
		const activeElement = document.activeElement;
		const isInputFocused = activeElement?.tagName === 'INPUT' ||
			activeElement?.tagName === 'TEXTAREA' ||
			activeElement?.getAttribute('contenteditable') === 'true';

		if (e.key === 'Enter' && !isInputFocused) {
			e.preventDefault();
			onCreateBlank();
		}
	}

	const templates: Template[] = [
		{
			id: 'blank',
			name: 'Blank page',
			description: 'Start from scratch',
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			id: 'ai-notes',
			name: 'AI Notes',
			description: 'Let AI help you organize',
			icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
		},
		{
			id: 'meeting-notes',
			name: 'Meeting Notes',
			description: 'Template for meeting notes',
			icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z'
		},
		{
			id: 'project-brief',
			name: 'Project Brief',
			description: 'Outline for new projects',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01'
		},
		{
			id: 'database',
			name: 'Database',
			description: 'Structured data view',
			icon: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4',
			disabled: true
		},
		{
			id: 'form',
			name: 'Form',
			description: 'Collect responses',
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
			disabled: true
		}
	];

	function handleTemplateClick(template: Template) {
		console.log('[NewPageWelcome] Template clicked:', template.id, template);
		if (template.disabled) {
			console.log('[NewPageWelcome] Template is disabled, returning');
			return;
		}
		if (template.id === 'blank') {
			console.log('[NewPageWelcome] Calling onCreateBlank, function exists:', !!onCreateBlank);
			if (onCreateBlank) {
				onCreateBlank();
				console.log('[NewPageWelcome] onCreateBlank called successfully');
			} else {
				console.error('[NewPageWelcome] onCreateBlank is not defined!');
			}
		} else if (template.id === 'ai-notes' && onAskAI) {
			onAskAI();
		} else {
			onSelectTemplate?.(template.id);
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="flex flex-col items-center justify-center min-h-[400px] py-12 px-6">
	<!-- Title -->
	<div class="text-center mb-8">
		<h1 class="text-3xl font-light text-gray-300 dark:text-gray-600 mb-2">
			New page
		</h1>
		<p class="text-sm text-gray-400 dark:text-gray-500">
			Choose how you want to start
		</p>
	</div>

	<!-- Get started with section -->
	<div class="w-full max-w-2xl">
		<p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">
			Get started with
		</p>

		<div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
			{#each templates as template (template.id)}
				<button
					onclick={() => handleTemplateClick(template)}
					disabled={template.disabled}
					class="group flex flex-col items-center gap-3 p-4 rounded-xl border border-gray-200 dark:border-gray-700 transition-all
						{template.disabled
							? 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800/50'
							: 'hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/20 cursor-pointer'}"
				>
					<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center
						{template.disabled ? '' : 'group-hover:bg-blue-100 dark:group-hover:bg-blue-900/50'}">
						<svg
							class="w-5 h-5 text-gray-500 dark:text-gray-400 {template.disabled ? '' : 'group-hover:text-blue-600 dark:group-hover:text-blue-400'}"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={template.icon} />
						</svg>
					</div>
					<div class="text-center">
						<div class="text-sm font-medium text-gray-700 dark:text-gray-200 {template.disabled ? '' : 'group-hover:text-blue-600 dark:group-hover:text-blue-400'}">
							{template.name}
						</div>
						<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
							{template.description}
						</div>
						{#if template.disabled}
							<span class="inline-block mt-1 px-2 py-0.5 text-[10px] bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 rounded-full">
								Coming soon
							</span>
						{/if}
					</div>
				</button>
			{/each}
		</div>
	</div>

	<!-- Quick action -->
	<div class="mt-8 flex items-center gap-3">
		<span class="text-xs text-gray-400 dark:text-gray-500">or press</span>
		<kbd class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded text-gray-600 dark:text-gray-300">
			Enter
		</kbd>
		<span class="text-xs text-gray-400 dark:text-gray-500">to start with a blank page</span>
	</div>
</div>
