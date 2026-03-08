<script lang="ts">
	import { TerminalApp } from '$lib/components/terminal';
	import DesktopSettingsContent from '$lib/components/desktop/DesktopSettingsContent.svelte';
	import FolderWindow from '$lib/components/desktop/FolderWindow.svelte';
	import FileBrowser from '$lib/components/desktop/FileBrowser.svelte';
	import type { DeployedApp } from '$lib/stores/deployedAppsStore';

	// Module-to-iframe-URL mapping
	const MODULE_URLS: Record<string, string> = {
		platform: '/dashboard',
		dashboard: '/dashboard?embed=true',
		chat: '/chat?embed=true',
		tasks: '/tasks?embed=true',
		projects: '/projects?embed=true',
		team: '/team?embed=true',
		contexts: '/knowledge?embed=true',
		nodes: '/nodes?embed=true',
		daily: '/daily?embed=true',
		settings: '/settings?embed=true',
		clients: '/clients?embed=true',
		tables: '/tables?embed=true',
		communication: '/communication/calendar?embed=true',
		calendar: '/communication/calendar?embed=true',
		pages: '/pages?embed=true',
		knowledge: '/pages?embed=true',
		'ai-settings': '/settings/ai?embed=true',
		integrations: '/integrations?embed=true',
		help: '/help?embed=true'
	};

	// Iframe titles for accessibility
	const MODULE_TITLES: Record<string, string> = {
		platform: 'Business OS',
		dashboard: 'Dashboard',
		chat: 'Chat',
		tasks: 'Tasks',
		projects: 'Projects',
		team: 'Team',
		contexts: 'Pages',
		nodes: 'Nodes',
		daily: 'Daily Log',
		settings: 'Settings',
		clients: 'Clients',
		tables: 'Tables',
		communication: 'Communication',
		calendar: 'Calendar',
		pages: 'Pages',
		knowledge: 'Pages',
		'ai-settings': 'AI Settings',
		integrations: 'Integrations',
		help: 'Help'
	};

	interface Props {
		module: string;
		windowTitle: string;
		deployedApps: DeployedApp[];
	}

	let { module, windowTitle, deployedApps }: Props = $props();

	const isTerminal = $derived(module === 'terminal');
	const isDesktopSettings = $derived(module === 'desktop-settings');
	const isFolder = $derived(module.startsWith('folder-'));
	const isFileBrowser = $derived(module === 'files' || module === 'finder');
	const isOsaApp = $derived(module.startsWith('osa-app-'));
	const isIframe = $derived(!isTerminal && !isDesktopSettings && !isFolder && !isFileBrowser && !isOsaApp);

	const folderId = $derived(isFolder ? module.replace('folder-', '') : '');
	const osaAppId = $derived(isOsaApp ? module.replace('osa-app-', '') : '');
	const deployedApp = $derived(isOsaApp ? deployedApps.find(app => app.id === osaAppId) : undefined);

	const iframeUrl = $derived(MODULE_URLS[module] ?? null);
	const iframeTitle = $derived(MODULE_TITLES[module] ?? windowTitle);
</script>

<div class="window-module-content">
	{#if isTerminal}
		<TerminalApp />
	{:else if isDesktopSettings}
		<DesktopSettingsContent />
	{:else if isFolder}
		<FolderWindow {folderId} />
	{:else if isFileBrowser}
		<FileBrowser />
	{:else if isOsaApp}
		{#if deployedApp && deployedApp.status === 'running'}
			<iframe
				src={deployedApp.url}
				title={deployedApp.name}
				class="module-iframe osa-app-iframe"
				sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
			></iframe>
		{:else}
			<div class="module-placeholder">
				<span class="placeholder-icon">
					<svg class="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</span>
				<span class="placeholder-text">App {deployedApp?.status === 'crashed' ? 'Crashed' : 'Not Running'}</span>
			</div>
		{/if}
	{:else if isIframe && iframeUrl}
		<iframe src={iframeUrl} title={iframeTitle} class="module-iframe"></iframe>
	{:else}
		<div class="module-placeholder">
			<span class="placeholder-icon">
				<svg class="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
			</span>
			<span class="placeholder-text">{windowTitle}</span>
		</div>
	{/if}
</div>

<style>
	.window-module-content {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.module-iframe {
		width: 100%;
		height: 100%;
		border: none;
		background: white;
	}

	.module-placeholder {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 16px;
		color: #999;
		background: #fafafa;
	}

	.placeholder-icon {
		color: #ccc;
	}

	.placeholder-text {
		font-size: 14px;
		font-weight: 500;
	}
</style>
