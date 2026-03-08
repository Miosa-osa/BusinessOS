<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { terminalStore } from '$lib/stores/terminal';
	import { chatModelStore } from '$lib/stores/chat/chatModelStore.svelte';
	import TerminalTabBar from './TerminalTabBar.svelte';
	import TerminalProviderBar from './TerminalProviderBar.svelte';
	import TerminalFocusBar from './TerminalFocusBar.svelte';
	import TerminalSplitContainer from './TerminalSplitContainer.svelte';
	import TerminalEnvironmentBar from './TerminalEnvironmentBar.svelte';
	import TerminalSandboxAnalysis from './TerminalSandboxAnalysis.svelte';
	import type { TerminalProvider, TerminalConfig, EnvironmentMode } from '$lib/stores/terminal/terminalTypes';
	import { getThemeColors } from './themes';

	// Current state from store
	const tabs = $derived(terminalStore.tabs);
	const activeTabId = $derived(terminalStore.activeTabId);
	const activeTab = $derived(terminalStore.activeTab);
	const config = $derived(terminalStore.config);
	const focusedPaneId = $derived(terminalStore.focusedPaneId);

	// Focus mode state
	let activeFocusMode = $state('general');

	// Sandbox analysis state
	let showSandboxAnalysis = $state(false);

	// Background color follows theme
	const bgColor = $derived(getThemeColors(config.theme).background);

	// Current provider for focus bar
	const currentProvider = $derived(activeTab?.provider ?? 'shell');

	// Environment mode derived values
	const currentEnvMode = $derived(activeTab ? terminalStore.getEnvironmentInfo(activeTab.id).mode : 'local');
	const currentEnvInfo = $derived(activeTab ? terminalStore.getEnvironmentInfo(activeTab.id) : { mode: 'local' as EnvironmentMode });

	function handleNewTab() {
		terminalStore.createTab(currentProvider);
	}

	function handleCloseTab(tabId: string) {
		terminalStore.closeTab(tabId);
	}

	function handleSwitchTab(tabId: string) {
		terminalStore.switchTab(tabId);
	}

	function handleProviderChange(provider: TerminalProvider) {
		if (activeTabId) {
			terminalStore.setTabProvider(activeTabId, provider);
		}
	}

	function handleConfigChange(partial: Partial<TerminalConfig>) {
		terminalStore.updateConfig(partial);
	}

	function handleSessionCreated(paneId: string, sessionId: string) {
		terminalStore.setPaneSessionId(paneId, sessionId);
	}

	function handlePaneFocus(paneId: string) {
		terminalStore.setFocusedPane(paneId);
	}

	function handleFocusModeChange(mode: string) {
		activeFocusMode = mode;
	}

	function handleEnvironmentChange(mode: EnvironmentMode) {
		if (activeTabId) {
			terminalStore.setEnvironmentMode(activeTabId, mode);
		}
	}

	function handleExitSandbox() {
		showSandboxAnalysis = true;
	}

	function handleSandboxMerge() {
		showSandboxAnalysis = false;
		if (activeTabId) {
			terminalStore.setEnvironmentMode(activeTabId, 'production');
		}
	}

	function handleSandboxStay() {
		showSandboxAnalysis = false;
	}

	function handleLaunchAgent(agent: string) {
		// TODO: Send agent command to active shell pane via WebSocket
		// For now this is a placeholder - needs integration with TerminalShell
		console.log('Launch agent:', agent);
	}

	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		const isMeta = e.metaKey || e.ctrlKey;

		if (isMeta && e.key === 't') {
			e.preventDefault();
			handleNewTab();
		} else if (isMeta && e.key === 'w') {
			e.preventDefault();
			if (activeTabId) handleCloseTab(activeTabId);
		} else if (isMeta && e.shiftKey && e.key === '[') {
			e.preventDefault();
			terminalStore.switchTabRelative(-1);
		} else if (isMeta && e.shiftKey && e.key === ']') {
			e.preventDefault();
			terminalStore.switchTabRelative(1);
		} else if (isMeta && !e.shiftKey && e.key === 'd') {
			e.preventDefault();
			const targetId = getSplitTarget();
			if (targetId) terminalStore.splitPane(targetId, 'horizontal');
		} else if (isMeta && e.shiftKey && (e.key === 'D' || e.key === 'd')) {
			if (e.shiftKey) {
				e.preventDefault();
				const targetId = getSplitTarget();
				if (targetId) terminalStore.splitPane(targetId, 'vertical');
			}
		} else if (isMeta && /^[1-5]$/.test(e.key)) {
			e.preventDefault();
			const providers: TerminalProvider[] = ['shell', 'osa', 'claude', 'codex', 'ollama'];
			const idx = parseInt(e.key) - 1;
			if (idx >= 0 && idx < providers.length) {
				handleProviderChange(providers[idx]);
			}
		}
	}

	function getSplitTarget(): string | null {
		if (!activeTab) return null;
		const rootPane = terminalStore.panes[activeTab.rootPaneId];
		if (!rootPane) return null;

		if (focusedPaneId) {
			const found = terminalStore.findPaneNode(rootPane, focusedPaneId);
			if (found && found.type === 'leaf') return focusedPaneId;
		}

		const firstLeaf = terminalStore.getFirstLeaf(rootPane);
		return firstLeaf?.id ?? null;
	}

	onMount(async () => {
		terminalStore.init();
		window.addEventListener('keydown', handleKeydown);

		// Initialize chatModelStore for AI providers
		await chatModelStore.loadUserSettings();
		await chatModelStore.loadModels();
	});

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="terminal-app" style="background: {bgColor}">
	<!-- Environment Mode Bar (top) -->
	<TerminalEnvironmentBar
		currentMode={currentEnvMode}
		environmentInfo={currentEnvInfo}
		onModeChange={handleEnvironmentChange}
		onExitSandbox={handleExitSandbox}
	/>

	<!-- Provider Bar -->
	<TerminalProviderBar
		activeProvider={currentProvider}
		{config}
		environmentMode={currentEnvMode}
		onProviderChange={handleProviderChange}
		onConfigChange={handleConfigChange}
		onLaunchAgent={handleLaunchAgent}
	/>

	<!-- Focus Mode Bar (only when AI provider active) -->
	<TerminalFocusBar
		provider={currentProvider}
		{activeFocusMode}
		onFocusModeChange={handleFocusModeChange}
	/>

	<!-- Tab Bar -->
	<TerminalTabBar
		{tabs}
		{activeTabId}
		onSwitchTab={handleSwitchTab}
		onCloseTab={handleCloseTab}
		onNewTab={handleNewTab}
	/>

	<!-- Pane Area — one pane tree per tab, only active shown -->
	<div class="pane-area">
		{#each tabs as tab (tab.id)}
			{@const paneTree = terminalStore.panes[tab.rootPaneId]}
			{#if paneTree}
				<div
					class="tab-pane"
					class:active={tab.id === activeTabId}
				>
					<TerminalSplitContainer
						node={paneTree}
						{config}
						{activeFocusMode}
						onSessionCreated={handleSessionCreated}
						onFocus={handlePaneFocus}
					/>
				</div>
			{/if}
		{/each}
	</div>

	<!-- Sandbox Analysis Modal -->
	<TerminalSandboxAnalysis
		sessionId={activeTab?.sessionId ?? ''}
		visible={showSandboxAnalysis}
		onMerge={handleSandboxMerge}
		onStay={handleSandboxStay}
	/>
</div>

<style>
	.terminal-app {
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		overflow: hidden;
	}

	.pane-area {
		flex: 1;
		position: relative;
		overflow: hidden;
	}

	.tab-pane {
		position: absolute;
		inset: 0;
		display: none;
	}

	.tab-pane.active {
		display: flex;
	}
</style>
