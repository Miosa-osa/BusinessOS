<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { contexts } from '$lib/stores/contexts';
	import { editor, wordCount, type EditorBlock, type BlockType, blockTypes, createEmptyBlock } from '$lib/stores/editor';
	import type { Context, Block } from '$lib/api/client';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';
	import ChatInput from '$lib/components/chat/ChatInput.svelte';
	import AssistantMessage from '$lib/components/chat/AssistantMessage.svelte';
	import UserMessage from '$lib/components/chat/UserMessage.svelte';
	import TypingIndicator from '$lib/components/chat/TypingIndicator.svelte';
	import { desktopBackgrounds, getBackgroundCSS } from '$lib/stores/desktopStore';

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let context: Context | null = $state(null);
	let parentContext: Context | null = $state(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	// Profile selector state
	let availableProfiles: Context[] = $state([]);
	let showProfileSelector = $state(false);
	let loadingProfiles = $state(false);
	let titleInput: HTMLInputElement | null = $state(null);
	let title = $state('');
	let icon = $state<string | null>(null);
	let coverImage = $state<string | null>(null);
	let showCoverInput = $state(false);
	let showCoverPicker = $state(false);
	let coverInputValue = $state('');
	let coverTab = $state<'presets' | 'url'>('presets');

	// Cover preset backgrounds (subset of desktop backgrounds suitable for covers)
	const coverPresets = desktopBackgrounds.filter(bg =>
		bg.type === 'gradient' || (bg.type === 'solid' && !bg.id.includes('dark'))
	).slice(0, 20);
	let autoSaveTimer: ReturnType<typeof setTimeout>;
	let showShareMenu = $state(false);
	let shareUrl = $state('');

	// AI Panel state
	interface AIMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		timestamp: string;
	}
	let aiMessages = $state<AIMessage[]>([]);
	let aiInput = $state('');
	let isAIStreaming = $state(false);
	let aiMessagesContainer: HTMLDivElement;

	const contextId = $derived($page.params.id);

	// Icon options for picker
	const iconOptions = [
		{ icon: '📄', label: 'Document' },
		{ icon: '📝', label: 'Note' },
		{ icon: '📋', label: 'List' },
		{ icon: '📁', label: 'Folder' },
		{ icon: '💡', label: 'Idea' },
		{ icon: '⭐', label: 'Star' },
		{ icon: '🎯', label: 'Goal' },
		{ icon: '✅', label: 'Done' },
		{ icon: '📊', label: 'Chart' },
		{ icon: '💻', label: 'Code' },
		{ icon: '🔗', label: 'Link' },
		{ icon: '📅', label: 'Date' },
		{ icon: '👤', label: 'Person' },
		{ icon: '🏢', label: 'Business' },
		{ icon: '🚀', label: 'Project' },
		{ icon: '💬', label: 'Chat' },
	];

	let showIconPicker = $state(false);

	onMount(async () => {
		try {
			const ctx = await contexts.loadContext(contextId);
			context = ctx;
			title = ctx.name;
			icon = ctx.icon;
			coverImage = ctx.cover_image;
			editor.initialize(ctx.blocks);

			// Load parent context if exists
			if (ctx.parent_id) {
				try {
					parentContext = await contexts.loadContext(ctx.parent_id);
				} catch (e) {
					console.error('Failed to load parent context:', e);
				}
			}

			loading = false;
		} catch (e) {
			error = 'Failed to load document';
			loading = false;
		}
	});

	onDestroy(() => {
		if (autoSaveTimer) clearTimeout(autoSaveTimer);
		editor.reset();
	});

	// Auto-save with debounce
	$effect(() => {
		if ($editor.isDirty && context) {
			if (autoSaveTimer) clearTimeout(autoSaveTimer);
			autoSaveTimer = setTimeout(async () => {
				await saveDocument();
			}, 1500);
		}
	});

	async function saveDocument() {
		if (!context || $editor.isSaving) return;
		editor.setSaving(true);
		try {
			await contexts.updateBlocks(context.id, $editor.blocks as Block[], $wordCount);
			editor.markSaved();
		} catch (e) {
			console.error('Failed to save:', e);
			editor.setSaving(false);
		}
	}

	async function updateTitle() {
		if (!context || title === context.name) return;
		try {
			await contexts.updateContext(context.id, { name: title });
		} catch (e) {
			console.error('Failed to update title:', e);
		}
	}

	async function updateIcon(newIcon: string) {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { icon: newIcon || null });
			icon = newIcon || null;
		} catch (e) {
			console.error('Failed to update icon:', e);
		}
	}

	async function updateCoverImage() {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { cover_image: coverInputValue || null });
			coverImage = coverInputValue || null;
			showCoverInput = false;
			coverInputValue = '';
		} catch (e) {
			console.error('Failed to update cover:', e);
		}
	}

	async function removeCoverImage() {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { cover_image: null });
			coverImage = null;
			showCoverPicker = false;
		} catch (e) {
			console.error('Failed to remove cover:', e);
		}
	}

	async function selectCoverPreset(bgId: string) {
		if (!context) return;
		const bg = desktopBackgrounds.find(b => b.id === bgId);
		if (!bg) return;

		// For presets, we store the background CSS directly
		const coverValue = `preset:${bgId}`;
		try {
			await contexts.updateContext(context.id, { cover_image: coverValue });
			coverImage = coverValue;
			showCoverPicker = false;
		} catch (e) {
			console.error('Failed to set cover preset:', e);
		}
	}

	function getCoverStyle(cover: string | null): string {
		if (!cover) return '';
		if (cover.startsWith('preset:')) {
			const bgId = cover.replace('preset:', '');
			return getBackgroundCSS(bgId);
		}
		return `background-image: url(${cover}); background-size: cover; background-position: center;`;
	}

	async function toggleShare() {
		if (!context) return;
		try {
			if (context.is_public) {
				await contexts.disableSharing(context.id);
				showShareMenu = false;
			} else {
				const response = await contexts.enableSharing(context.id);
				shareUrl = response.share_url;
			}
		} catch (e) {
			console.error('Failed to toggle sharing:', e);
		}
	}

	async function copyShareLink() {
		if (shareUrl) {
			await navigator.clipboard.writeText(shareUrl);
		}
	}

	async function duplicateDoc() {
		if (!context) return;
		try {
			const newContext = await contexts.duplicateContext(context.id);
			goto(`/contexts/${newContext.id}${embedSuffix}`);
		} catch (e) {
			console.error('Failed to duplicate:', e);
		}
	}

	async function archiveDoc() {
		if (!context) return;
		try {
			await contexts.archiveContext(context.id);
			goto('/contexts' + embedSuffix);
		} catch (e) {
			console.error('Failed to archive:', e);
		}
	}

	async function deleteDoc() {
		if (!context) return;
		if (!confirm('Are you sure you want to delete this document? This cannot be undone.')) return;
		try {
			await contexts.deleteContext(context.id);
			goto('/contexts' + embedSuffix);
		} catch (e) {
			console.error('Failed to delete:', e);
		}
	}

	async function loadAvailableProfiles() {
		if (loadingProfiles) return;
		loadingProfiles = true;
		try {
			// Load all contexts and wait for them
			await contexts.loadContexts();
			// Get profiles from the store (non-document contexts) excluding current
			availableProfiles = $contexts.contexts.filter(
				(c) => c.type !== 'document' && c.id !== contextId
			) as Context[];
		} catch (e) {
			console.error('Failed to load profiles:', e);
		} finally {
			loadingProfiles = false;
		}
	}

	async function updateParentProfile(profileId: string | null) {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { parent_id: profileId });
			if (profileId) {
				parentContext = availableProfiles.find(p => p.id === profileId) || null;
			} else {
				parentContext = null;
			}
			showProfileSelector = false;
		} catch (e) {
			console.error('Failed to update parent profile:', e);
		}
	}

	function handleTitleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			updateTitle();
			// Focus first block
			const firstBlock = document.querySelector('[data-block-id]') as HTMLElement;
			firstBlock?.focus();
		}
	}

	function handleBlockKeydown(e: KeyboardEvent) {
		// Space on empty block opens AI panel
		if (e.key === ' ' && !e.ctrlKey && !e.metaKey) {
			const target = e.target as HTMLElement;
			const blockId = target.getAttribute('data-block-id');
			if (blockId) {
				const block = $editor.blocks.find((b) => b.id === blockId);
				if (block && block.content === '') {
					e.preventDefault();
					editor.showAIPanel();
				}
			}
		}
	}

	function getTypeIcon(type: BlockType) {
		return blockTypes.find(bt => bt.type === type)?.icon || 'T';
	}

	function addNewBlockAtEnd() {
		// Add a new empty block at the end
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			const newBlockId = editor.addBlockAfter(lastBlock.id);
			// Focus the new block after render
			setTimeout(() => {
				const blockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				blockEl?.focus();
			}, 10);
		}
	}

	// Close icon picker and profile selector when clicking outside
	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (showIconPicker && !target.closest('.relative.inline-block')) {
			showIconPicker = false;
		}
		if (showProfileSelector && !target.closest('.profile-selector-container')) {
			showProfileSelector = false;
		}
	}

	// AI Panel functions
	async function handleAISend(message: string) {
		if (!message.trim() || isAIStreaming) return;

		// Add user message
		const userMessage: AIMessage = {
			id: crypto.randomUUID(),
			role: 'user',
			content: message,
			timestamp: new Date().toISOString()
		};
		aiMessages = [...aiMessages, userMessage];

		// Scroll to bottom
		setTimeout(() => {
			if (aiMessagesContainer) {
				aiMessagesContainer.scrollTop = aiMessagesContainer.scrollHeight;
			}
		}, 10);

		isAIStreaming = true;

		try {
			// Get current document content for context
			const documentContent = $editor.blocks.map(b => b.content).join('\n');

			// Call AI endpoint
			const response = await fetch('/api/chat/ai/document', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					message,
					context: {
						documentTitle: title,
						documentContent,
						contextType: context?.type
					}
				})
			});

			if (!response.ok) {
				throw new Error('AI request failed');
			}

			const data = await response.json();

			// Add assistant message
			const assistantMessage: AIMessage = {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: data.response || 'I apologize, but I was unable to generate a response.',
				timestamp: new Date().toISOString()
			};
			aiMessages = [...aiMessages, assistantMessage];

		} catch (error) {
			console.error('AI error:', error);
			// Add error message
			const errorMessage: AIMessage = {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: 'Sorry, I encountered an error. Please try again.',
				timestamp: new Date().toISOString()
			};
			aiMessages = [...aiMessages, errorMessage];
		} finally {
			isAIStreaming = false;
			// Scroll to bottom
			setTimeout(() => {
				if (aiMessagesContainer) {
					aiMessagesContainer.scrollTop = aiMessagesContainer.scrollHeight;
				}
			}, 10);
		}
	}

	function handleAIStop() {
		isAIStreaming = false;
	}

	function insertAIContent(content: string) {
		// Insert AI-generated content as new blocks
		const lines = content.split('\n').filter(line => line.trim());
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];

		if (lastBlock) {
			let currentBlockId = lastBlock.id;
			for (const line of lines) {
				currentBlockId = editor.addBlockAfter(currentBlockId, 'paragraph');
				editor.updateBlock(currentBlockId, line);
			}
		}

		// Close AI panel after inserting
		editor.hideAIPanel();
	}

	function clearAIChat() {
		aiMessages = [];
	}
</script>

<svelte:head>
	<title>{title || 'Untitled'} - BusinessOS</title>
</svelte:head>

<svelte:window onclick={handleClickOutside} />

{#if loading}
	<div class="h-full flex items-center justify-center">
		<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
	</div>
{:else if error}
	<div class="h-full flex items-center justify-center">
		<div class="text-center">
			<p class="text-red-500 mb-4">{error}</p>
			<a href="/contexts{embedSuffix}" class="btn btn-secondary">Back to Contexts</a>
		</div>
	</div>
{:else if context}
	<div class="h-full flex flex-col bg-white">
		<!-- Top toolbar -->
		<div class="px-4 py-2 border-b border-gray-100 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<a href="/contexts{embedSuffix}" class="p-1.5 rounded hover:bg-gray-100 text-gray-500" title="Back to contexts">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</a>

				<!-- Profile Selector -->
				<div class="relative profile-selector-container">
					<button
						onclick={() => {
							loadAvailableProfiles();
							showProfileSelector = !showProfileSelector;
						}}
						class="flex items-center gap-1.5 px-2 py-1 rounded hover:bg-gray-100 text-sm transition-colors {parentContext ? 'text-gray-700 bg-gray-50' : 'text-gray-400'}"
						title={parentContext ? `Linked to ${parentContext.name}` : 'Add to a profile'}
					>
						{#if parentContext}
							<span class="text-base">{parentContext.icon || '📁'}</span>
							<span class="font-medium">{parentContext.name}</span>
							<span class="text-xs text-gray-400 ml-0.5">(linked)</span>
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							<span>Add to profile</span>
						{/if}
						<svg class="w-3 h-3 text-gray-400 ml-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					{#if showProfileSelector}
						<div class="absolute left-0 top-full mt-1 w-64 bg-white rounded-lg shadow-xl border border-gray-200 py-1 z-50 max-h-80 overflow-y-auto">
							<div class="px-3 py-2 border-b border-gray-100">
								<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Link to Profile</span>
							</div>

							{#if loadingProfiles}
								<div class="px-3 py-4 text-center">
									<div class="animate-spin h-5 w-5 border-2 border-gray-300 border-t-gray-600 rounded-full mx-auto"></div>
								</div>
							{:else if availableProfiles.length === 0}
								<div class="px-3 py-4 text-center text-sm text-gray-500">
									No profiles available
								</div>
							{:else}
								{#if parentContext}
									<button
										onclick={() => updateParentProfile(null)}
										class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
										Remove link
									</button>
									<hr class="my-1 border-gray-100" />
								{/if}

								{#each availableProfiles as profile}
									<button
										onclick={() => updateParentProfile(profile.id)}
										class="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2 {parentContext?.id === profile.id ? 'bg-blue-50 text-blue-700' : 'text-gray-700'}"
									>
										<span class="text-base">{profile.icon || '📁'}</span>
										<span class="truncate flex-1">{profile.name}</span>
										{#if parentContext?.id === profile.id}
											<svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
										{/if}
									</button>
								{/each}
							{/if}
						</div>
					{/if}
				</div>

				{#if parentContext}
					<span class="text-sm text-gray-400">/</span>
				{/if}
				<span class="text-sm text-gray-600">{title || 'Untitled'}</span>
			</div>

			<div class="flex items-center gap-2">
				<!-- Save status -->
				<div class="text-xs text-gray-400 mr-2">
					{#if $editor.isDirty}
						<span class="text-amber-500">Unsaved</span>
					{:else if $editor.isSaving}
						<span>Saving...</span>
					{:else if $editor.lastSavedAt}
						<span>Saved</span>
					{/if}
				</div>

				<!-- Share button -->
				<div class="relative">
					<button
						onclick={() => showShareMenu = !showShareMenu}
						class="btn btn-secondary text-sm flex items-center gap-1.5"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
						</svg>
						Share
					</button>

					{#if showShareMenu}
						<div class="absolute right-0 top-full mt-2 w-72 bg-white rounded-lg shadow-xl border border-gray-200 p-4 z-50">
							<div class="flex items-center justify-between mb-3">
								<span class="text-sm font-medium text-gray-900">Share to web</span>
								<button
									onclick={toggleShare}
									class="relative w-10 h-6 rounded-full transition-colors {context.is_public ? 'bg-blue-500' : 'bg-gray-200'}"
								>
									<span
										class="absolute top-1 w-4 h-4 rounded-full bg-white shadow transition-transform {context.is_public ? 'left-5' : 'left-1'}"
									></span>
								</button>
							</div>
							{#if context.is_public}
								<div class="space-y-2">
									<p class="text-xs text-gray-500">Anyone with the link can view</p>
									<div class="flex gap-2">
										<input
											type="text"
											value={shareUrl}
											readonly
											class="input input-square text-xs flex-1"
										/>
										<button onclick={copyShareLink} class="btn btn-primary text-xs">Copy</button>
									</div>
								</div>
							{:else}
								<p class="text-xs text-gray-500">Enable sharing to get a public link</p>
							{/if}
						</div>
					{/if}
				</div>

				<!-- More options -->
				<div class="relative group">
					<button class="p-2 rounded hover:bg-gray-100 text-gray-500">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
						</svg>
					</button>
					<div class="absolute right-0 top-full mt-1 w-48 bg-white rounded-lg shadow-xl border border-gray-200 py-1 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50">
						<button onclick={duplicateDoc} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							Duplicate
						</button>
						<button onclick={archiveDoc} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
							</svg>
							Archive
						</button>
						<hr class="my-1 border-gray-100" />
						<button onclick={deleteDoc} class="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Document content -->
		<div class="flex-1 overflow-y-auto">
			<div class="max-w-3xl mx-auto px-8 py-12">
				<!-- Cover Image -->
				{#if coverImage}
					<div class="relative -mx-8 -mt-12 mb-8 h-48 group">
						{#if coverImage.startsWith('preset:')}
							<div class="w-full h-full" style={getCoverStyle(coverImage)}></div>
						{:else}
							<img src={coverImage} alt="Cover" class="w-full h-full object-cover" />
						{/if}
						<div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center gap-2 opacity-0 group-hover:opacity-100">
							<button
								onclick={() => showCoverPicker = true}
								class="btn btn-secondary text-sm"
							>
								Change cover
							</button>
							<button onclick={removeCoverImage} class="btn btn-secondary text-sm">
								Remove
							</button>
						</div>
					</div>
				{/if}

				<!-- Add cover button -->
				{#if !coverImage}
					<div class="mb-4">
						<button
							onclick={() => showCoverPicker = true}
							class="text-sm text-gray-400 hover:text-gray-600 flex items-center gap-1"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							Add cover
						</button>
					</div>
				{/if}

				<!-- Icon and Title -->
				<div class="mb-6">
					<!-- Icon Picker - click to show -->
					<div class="relative inline-block mb-2">
						<button
							onclick={() => showIconPicker = !showIconPicker}
							class="w-12 h-12 flex items-center justify-center text-3xl hover:bg-gray-100 rounded-lg transition-colors"
						>
							{icon || '📄'}
						</button>
						{#if showIconPicker}
							<div class="absolute top-full left-0 mt-1 bg-white rounded-lg shadow-xl border border-gray-200 p-3 z-20 w-64">
								<div class="text-xs font-medium text-gray-500 mb-2">Choose icon</div>
								<div class="grid grid-cols-8 gap-1">
									{#each iconOptions as opt}
										<button
											onclick={() => { updateIcon(opt.icon); showIconPicker = false; }}
											class="w-7 h-7 flex items-center justify-center hover:bg-gray-100 rounded text-lg"
											title={opt.label}
										>
											{opt.icon}
										</button>
									{/each}
								</div>
								<hr class="my-2 border-gray-100" />
								<button
									onclick={() => { updateIcon(''); showIconPicker = false; }}
									class="w-full px-2 py-1.5 text-left text-sm text-gray-500 hover:bg-gray-100 rounded flex items-center gap-2"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
									Remove icon
								</button>
							</div>
						{/if}
					</div>

					<!-- Title -->
					<input
						bind:this={titleInput}
						bind:value={title}
						onblur={updateTitle}
						onkeydown={handleTitleKeydown}
						onfocus={(e) => { if (title === 'Untitled') { title = ''; (e.target as HTMLInputElement).select(); } }}
						placeholder="Untitled"
						class="w-full text-4xl font-bold border-none outline-none bg-transparent {title === 'Untitled' || !title ? 'text-gray-300' : 'text-gray-900'}"
					/>
				</div>


				<!-- Blocks -->
				<div
					class="blocks-container"
					onkeydown={handleBlockKeydown}
					role="textbox"
					tabindex="-1"
				>
					{#each $editor.blocks as block, index (block.id)}
						<BlockComponent {block} {index} readonly={false} parentContextId={contextId} />
					{/each}
				</div>

				<!-- Click area to add new blocks -->
				<button
					onclick={addNewBlockAtEnd}
					class="w-full min-h-32 mt-4 text-left cursor-text group"
				>
					<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
						Click to add a block, or press / for commands
					</span>
				</button>
			</div>
		</div>

		<!-- Status Bar -->
		<div class="px-4 py-2 border-t border-gray-100 flex items-center justify-between text-xs text-gray-400">
			<div class="flex items-center gap-4">
				<span>{$wordCount} words</span>
				<span>{$editor.blocks.length} blocks</span>
			</div>
			<div class="flex items-center gap-2">
				<button onclick={saveDocument} class="hover:text-gray-600" disabled={!$editor.isDirty}>
					Save now
				</button>
			</div>
		</div>

		<!-- Slash Command Menu (global) -->
		{#if $editor.showSlashMenu && $editor.slashMenuPosition}
			<BlockMenu />
		{/if}

		<!-- AI Panel -->
		{#if $editor.showAIPanel}
			<div class="fixed inset-y-0 right-0 w-[420px] bg-white border-l border-gray-200 shadow-xl z-50 flex flex-col">
				<!-- Header -->
				<div class="p-4 border-b border-gray-100 flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
							<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
						</div>
						<div>
							<h3 class="font-medium text-gray-900">AI Assistant</h3>
							<p class="text-xs text-gray-400">Help with writing & editing</p>
						</div>
					</div>
					<div class="flex items-center gap-1">
						{#if aiMessages.length > 0}
							<button
								onclick={clearAIChat}
								class="p-2 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
								title="Clear chat"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						{/if}
						<button
							onclick={() => editor.hideAIPanel()}
							class="p-2 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
							title="Close"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="px-4 py-3 border-b border-gray-100 flex flex-wrap gap-2">
					<button
						onclick={() => handleAISend('Help me write a summary of this document')}
						class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
					>
						✨ Summarize
					</button>
					<button
						onclick={() => handleAISend('Help me improve the writing in this document')}
						class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
					>
						📝 Improve writing
					</button>
					<button
						onclick={() => handleAISend('Check this document for grammar and spelling errors')}
						class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
					>
						🔍 Check grammar
					</button>
					<button
						onclick={() => handleAISend('Make this document shorter and more concise')}
						class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
					>
						📐 Make shorter
					</button>
					<button
						onclick={() => handleAISend('Expand on the ideas in this document with more detail')}
						class="px-3 py-1.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
					>
						📚 Expand
					</button>
				</div>

				<!-- Messages Area -->
				<div
					bind:this={aiMessagesContainer}
					class="flex-1 overflow-y-auto p-4 space-y-4"
				>
					{#if aiMessages.length === 0}
						<div class="text-center py-8">
							<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 flex items-center justify-center">
								<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
								</svg>
							</div>
							<h4 class="text-sm font-medium text-gray-700 mb-1">How can I help?</h4>
							<p class="text-xs text-gray-400 max-w-[240px] mx-auto">
								Ask me to write, edit, summarize, or improve your document content.
							</p>
						</div>
					{:else}
						{#each aiMessages as message (message.id)}
							{#if message.role === 'user'}
								<UserMessage
									content={message.content}
									timestamp={message.timestamp}
								/>
							{:else}
								<div>
									<AssistantMessage
										content={message.content}
										timestamp={message.timestamp}
										isStreaming={false}
										onCopy={() => navigator.clipboard.writeText(message.content)}
									/>
									<!-- Insert to document button -->
									<div class="ml-9 mt-1">
										<button
											onclick={() => insertAIContent(message.content)}
											class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors"
										>
											<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
											</svg>
											Insert into document
										</button>
									</div>
								</div>
							{/if}
						{/each}
						{#if isAIStreaming}
							<TypingIndicator />
						{/if}
					{/if}
				</div>

				<!-- Input Area -->
				<ChatInput
					bind:value={aiInput}
					placeholder="Ask AI to help with your document..."
					streaming={isAIStreaming}
					contextName={title || 'Document'}
					modelName="AI Assistant"
					onSend={handleAISend}
					onStop={handleAIStop}
				/>
			</div>
		{/if}

		<!-- Cover Picker Modal -->
		{#if showCoverPicker}
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div
				class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center"
				onclick={(e) => { if (e.target === e.currentTarget) showCoverPicker = false; }}
				onkeydown={(e) => { if (e.key === 'Escape') showCoverPicker = false; }}
				role="dialog"
				aria-modal="true"
				aria-label="Choose cover image"
				tabindex="-1"
			>
				<div class="bg-white rounded-xl shadow-2xl w-[480px] max-h-[80vh] flex flex-col">
					<!-- Header -->
					<div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
						<h3 class="font-semibold text-gray-900">Choose cover</h3>
						<button
							onclick={() => showCoverPicker = false}
							class="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600"
							aria-label="Close"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>

					<!-- Tabs -->
					<div class="px-5 pt-3 flex gap-1">
						<button
							onclick={() => coverTab = 'presets'}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {coverTab === 'presets' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
						>
							Gradients
						</button>
						<button
							onclick={() => coverTab = 'url'}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {coverTab === 'url' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
						>
							Image URL
						</button>
					</div>

					<!-- Content -->
					<div class="flex-1 overflow-y-auto p-5">
						{#if coverTab === 'presets'}
							<div class="grid grid-cols-4 gap-2">
								{#each coverPresets as bg}
									<button
										onclick={() => selectCoverPreset(bg.id)}
										class="aspect-video rounded-lg transition-all hover:scale-105 hover:shadow-lg ring-2 ring-transparent hover:ring-blue-500 {coverImage === `preset:${bg.id}` ? 'ring-blue-500' : ''}"
										style={getBackgroundCSS(bg.id)}
										title={bg.name}
									></button>
								{/each}
							</div>
						{:else}
							<div class="space-y-4">
								<p class="text-sm text-gray-500">Paste an image URL to use as your cover</p>
								<div class="flex gap-2">
									<input
										type="text"
										bind:value={coverInputValue}
										placeholder="https://example.com/image.jpg"
										class="input input-square text-sm flex-1"
										onkeydown={(e) => e.key === 'Enter' && updateCoverImage()}
									/>
									<button
										onclick={updateCoverImage}
										class="btn btn-primary text-sm"
										disabled={!coverInputValue.trim()}
									>
										Apply
									</button>
								</div>
								{#if coverInputValue}
									<div class="mt-4">
										<p class="text-xs text-gray-400 mb-2">Preview:</p>
										<div class="aspect-video rounded-lg overflow-hidden bg-gray-100">
											<img
												src={coverInputValue}
												alt="Preview"
												class="w-full h-full object-cover"
												onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
											/>
										</div>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Footer -->
					{#if coverImage}
						<div class="px-5 py-3 border-t border-gray-100">
							<button
								onclick={removeCoverImage}
								class="text-sm text-red-600 hover:text-red-700 flex items-center gap-1.5"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
								Remove cover
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
{/if}
