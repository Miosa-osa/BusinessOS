<script lang="ts">
	import { onMount } from 'svelte';
	import { contexts } from '$lib/stores/contexts';
	import { Dialog } from 'bits-ui';

	let showNewContext = $state(false);
	let newContext = $state({
		name: '',
		type: 'custom' as const,
		content: '',
		system_prompt_template: ''
	});
	let typeFilter = $state('');

	onMount(() => {
		contexts.loadContexts();
	});

	async function handleCreateContext(e: Event) {
		e.preventDefault();
		try {
			await contexts.createContext(newContext);
			showNewContext = false;
			newContext = { name: '', type: 'custom', content: '', system_prompt_template: '' };
		} catch (error) {
			console.error('Failed to create context:', error);
		}
	}

	function getTypeIcon(type: string) {
		switch (type) {
			case 'person': return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business': return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			case 'project': return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			default: return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}

	function getTypeColor(type: string) {
		switch (type) {
			case 'person': return 'bg-purple-50 text-purple-600';
			case 'business': return 'bg-blue-50 text-blue-600';
			case 'project': return 'bg-green-50 text-green-600';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-gray-900">Contexts</h1>
			<p class="text-sm text-gray-500 mt-0.5">Store information about people, businesses, and projects</p>
		</div>
		<button onclick={() => showNewContext = true} class="btn btn-primary">
			+ New Context
		</button>
	</div>

	<!-- Filters -->
	<div class="px-6 py-3 border-b border-gray-100 flex gap-2">
		<button
			onclick={() => { typeFilter = ''; contexts.loadContexts(); }}
			class="btn {typeFilter === '' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			All
		</button>
		<button
			onclick={() => { typeFilter = 'person'; contexts.loadContexts('person'); }}
			class="btn {typeFilter === 'person' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			People
		</button>
		<button
			onclick={() => { typeFilter = 'business'; contexts.loadContexts('business'); }}
			class="btn {typeFilter === 'business' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			Businesses
		</button>
		<button
			onclick={() => { typeFilter = 'project'; contexts.loadContexts('project'); }}
			class="btn {typeFilter === 'project' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			Projects
		</button>
	</div>

	<!-- Contexts List -->
	<div class="flex-1 overflow-y-auto p-6">
		{#if $contexts.loading}
			<div class="flex items-center justify-center h-48">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else if $contexts.contexts.length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
					<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
					</svg>
				</div>
				<p class="text-gray-500">No contexts yet</p>
				<button onclick={() => showNewContext = true} class="btn btn-secondary text-sm mt-3">
					Create your first context
				</button>
			</div>
		{:else}
			<div class="space-y-3 max-w-3xl">
				{#each $contexts.contexts as context}
					<a href="/contexts/{context.id}" class="card flex items-start gap-4 hover:shadow-md transition-shadow duration-200 cursor-pointer">
						<div class="w-10 h-10 rounded-full {getTypeColor(context.type)} flex items-center justify-center shrink-0">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(context.type)} />
							</svg>
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<h3 class="font-medium text-gray-900">{context.name}</h3>
								<span class="text-xs text-gray-400 capitalize">{context.type}</span>
							</div>
							{#if context.content}
								<p class="text-sm text-gray-500 line-clamp-2">{context.content}</p>
							{/if}
							<p class="text-xs text-gray-400 mt-2">Updated {formatDate(context.updated_at)}</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- New Context Dialog -->
<Dialog.Root bind:open={showNewContext}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">New Context</Dialog.Title>

			<form onsubmit={handleCreateContext} class="space-y-4">
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
						<input
							id="name"
							type="text"
							bind:value={newContext.name}
							class="input input-square"
							placeholder="e.g. John Smith"
							required
						/>
					</div>
					<div>
						<label for="type" class="block text-sm font-medium text-gray-700 mb-1">Type</label>
						<select id="type" bind:value={newContext.type} class="input input-square">
							<option value="person">Person</option>
							<option value="business">Business</option>
							<option value="project">Project</option>
							<option value="custom">Custom</option>
						</select>
					</div>
				</div>

				<div>
					<label for="content" class="block text-sm font-medium text-gray-700 mb-1">Context Information</label>
					<textarea
						id="content"
						bind:value={newContext.content}
						class="input input-square resize-none"
						rows="4"
						placeholder="Store important information, background, preferences, history..."
					></textarea>
				</div>

				<div>
					<label for="prompt" class="block text-sm font-medium text-gray-700 mb-1">System Prompt (optional)</label>
					<textarea
						id="prompt"
						bind:value={newContext.system_prompt_template}
						class="input input-square resize-none"
						rows="3"
						placeholder="Custom instructions for AI when using this context..."
					></textarea>
					<p class="text-xs text-gray-400 mt-1">This will be used as the system prompt when chatting with this context</p>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => showNewContext = false} class="btn btn-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn btn-primary flex-1">
						Create Context
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
