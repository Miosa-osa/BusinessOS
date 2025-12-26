<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { contexts } from '$lib/stores/contexts';
	import type { Context } from '$lib/api/client';

	interface Props {
		document: Context | null;
		onSettingsChange?: (settings: PageSettings) => void;
	}

	interface PageSettings {
		fullWidth: boolean;
		smallText: boolean;
		locked: boolean;
	}

	// Use renamed prop to avoid conflict with browser's document object
	let { document: pageDoc, onSettingsChange }: Props = $props();
	const dispatch = createEventDispatcher();

	let showMenu = $state(false);
	let menuRef: HTMLDivElement | null = $state(null);

	// Parse settings from document properties or use defaults
	let settings = $derived<PageSettings>({
		fullWidth: Boolean(pageDoc?.properties?.fullWidth) ?? false,
		smallText: Boolean(pageDoc?.properties?.smallText) ?? false,
		locked: Boolean(pageDoc?.properties?.locked) ?? false
	});

	function toggleMenu() {
		showMenu = !showMenu;
	}

	function closeMenu() {
		showMenu = false;
	}

	async function toggleSetting(key: keyof PageSettings) {
		if (!pageDoc) return;

		const newSettings = {
			...settings,
			[key]: !settings[key]
		};

		try {
			await contexts.updateContext(pageDoc.id, {
				properties: {
					...pageDoc.properties,
					[key]: newSettings[key]
				}
			});
			onSettingsChange?.(newSettings);
		} catch (err) {
			console.error('Failed to update page setting:', err);
		}
	}

	async function duplicatePage() {
		if (!pageDoc) return;
		closeMenu();

		try {
			const newDoc = await contexts.createContext({
				name: `${pageDoc.name} (copy)`,
				type: pageDoc.type,
				parent_id: pageDoc.parent_id ?? undefined,
				content: pageDoc.content ?? undefined,
				blocks: pageDoc.blocks ?? undefined,
				icon: pageDoc.icon ?? undefined,
				cover_image: pageDoc.cover_image ?? undefined
			});

			// Navigate to the new page
			dispatch('duplicate', { newDocument: newDoc });
		} catch (err) {
			console.error('Failed to duplicate page:', err);
		}
	}

	async function deletePage() {
		if (!pageDoc) return;

		const confirmed = confirm(`Are you sure you want to delete "${pageDoc.name || 'Untitled'}"? This action cannot be undone.`);
		if (!confirmed) return;

		closeMenu();

		try {
			await contexts.deleteContext(pageDoc.id);
			dispatch('delete', { documentId: pageDoc.id });
		} catch (err) {
			console.error('Failed to delete page:', err);
		}
	}

	function handleClickOutside(e: MouseEvent) {
		if (menuRef && !menuRef.contains(e.target as Node)) {
			closeMenu();
		}
	}

	$effect(() => {
		if (showMenu) {
			// Use globalThis.document to reference browser's document
			globalThis.document.addEventListener('click', handleClickOutside);
			return () => globalThis.document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="relative" bind:this={menuRef}>
	<!-- Three dots button -->
	<button
		onclick={toggleMenu}
		class="p-1.5 rounded-md text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
		aria-label="Page options"
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z" />
		</svg>
	</button>

	<!-- Dropdown menu -->
	{#if showMenu}
		<div
			class="absolute right-0 top-full mt-1 w-64 bg-white dark:bg-[#252525] rounded-lg shadow-xl border border-gray-200 dark:border-[#3d3d3d] py-1 z-50"
		>
			<!-- Page Settings Section -->
			<div class="px-3 py-1.5">
				<p class="text-[11px] font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">Page settings</p>
			</div>

			<!-- Full Width Toggle -->
			<button
				onclick={() => toggleSetting('fullWidth')}
				class="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<div class="flex items-center gap-3">
					<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
					<span class="text-sm text-gray-700 dark:text-gray-200">Full width</span>
				</div>
				<div class="w-8 h-5 rounded-full relative transition-colors {settings.fullWidth ? 'bg-blue-500' : 'bg-gray-200 dark:bg-gray-600'}">
					<div class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {settings.fullWidth ? 'translate-x-3' : ''}"></div>
				</div>
			</button>

			<!-- Small Text Toggle -->
			<button
				onclick={() => toggleSetting('smallText')}
				class="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<div class="flex items-center gap-3">
					<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7V4h16v3M9 20h6M12 4v16" />
					</svg>
					<span class="text-sm text-gray-700 dark:text-gray-200">Small text</span>
				</div>
				<div class="w-8 h-5 rounded-full relative transition-colors {settings.smallText ? 'bg-blue-500' : 'bg-gray-200 dark:bg-gray-600'}">
					<div class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {settings.smallText ? 'translate-x-3' : ''}"></div>
				</div>
			</button>

			<!-- Divider -->
			<div class="my-1 border-t border-gray-200 dark:border-[#3d3d3d]"></div>

			<!-- Lock Page Toggle -->
			<button
				onclick={() => toggleSetting('locked')}
				class="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<div class="flex items-center gap-3">
					{#if settings.locked}
						<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					{:else}
						<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
						</svg>
					{/if}
					<span class="text-sm text-gray-700 dark:text-gray-200">Lock page</span>
				</div>
				<div class="w-8 h-5 rounded-full relative transition-colors {settings.locked ? 'bg-blue-500' : 'bg-gray-200 dark:bg-gray-600'}">
					<div class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {settings.locked ? 'translate-x-3' : ''}"></div>
				</div>
			</button>

			<!-- Divider -->
			<div class="my-1 border-t border-gray-200 dark:border-[#3d3d3d]"></div>

			<!-- Actions Section -->
			<div class="px-3 py-1.5">
				<p class="text-[11px] font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">Actions</p>
			</div>

			<!-- Duplicate -->
			<button
				onclick={duplicatePage}
				class="w-full px-3 py-2 flex items-center gap-3 text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
				</svg>
				<span class="text-sm text-gray-700 dark:text-gray-200">Duplicate</span>
			</button>

			<!-- Copy link -->
			<button
				onclick={() => {
					if (pageDoc) {
						navigator.clipboard.writeText(window.location.href);
						closeMenu();
					}
				}}
				class="w-full px-3 py-2 flex items-center gap-3 text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
				</svg>
				<span class="text-sm text-gray-700 dark:text-gray-200">Copy link</span>
			</button>

			<!-- Export -->
			<button
				onclick={() => {
					// TODO: Implement export functionality
					closeMenu();
					alert('Export coming soon!');
				}}
				class="w-full px-3 py-2 flex items-center gap-3 text-left hover:bg-gray-50 dark:hover:bg-[#2f2f2f] transition-colors"
			>
				<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
				<span class="text-sm text-gray-700 dark:text-gray-200">Export</span>
			</button>

			<!-- Divider -->
			<div class="my-1 border-t border-gray-200 dark:border-[#3d3d3d]"></div>

			<!-- Delete (danger) -->
			<button
				onclick={deletePage}
				class="w-full px-3 py-2 flex items-center gap-3 text-left hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
			>
				<svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
				</svg>
				<span class="text-sm text-red-500">Delete</span>
			</button>

			<!-- Footer hint -->
			<div class="px-3 py-2 border-t border-gray-200 dark:border-[#3d3d3d] mt-1">
				<p class="text-[11px] text-gray-400 dark:text-gray-500">
					Last edited {pageDoc?.updated_at ? new Date(pageDoc.updated_at).toLocaleDateString() : 'just now'}
				</p>
			</div>
		</div>
	{/if}
</div>

<style>
	/* Smooth shadow for the dropdown */
	div[class*="shadow-xl"] {
		box-shadow:
			0 0 0 1px rgba(0, 0, 0, 0.05),
			0 4px 8px rgba(0, 0, 0, 0.1),
			0 12px 24px rgba(0, 0, 0, 0.1);
	}

	:global(.dark) div[class*="shadow-xl"] {
		box-shadow:
			0 0 0 1px rgba(0, 0, 0, 0.2),
			0 4px 8px rgba(0, 0, 0, 0.2),
			0 12px 24px rgba(0, 0, 0, 0.2);
	}
</style>
