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

	// Shell write function registry — keyed by paneId
	const shellRefs = new Map<string, (data: string) => void>();

	function handleShellReady(paneId: string, write: (data: string) => void | null) {
		if (write) {
			shellRefs.set(paneId, write);

			// Auto-launch agent if this pane was queued for it
			if (pendingAgentLaunch.has(paneId)) {
				pendingAgentLaunch.delete(paneId);
				// Find the tab that owns this pane to get its provider
				const ownerTab = tabs.find(t => t.rootPaneId === paneId);
				const command = ownerTab ? AGENT_COMMANDS[ownerTab.provider] : null;
				if (command) {
					// Small delay to let the shell fully initialize
					setTimeout(() => write(command), 500);
				}
			}
		} else {
			shellRefs.delete(paneId);
		}
	}

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

	// Track which panes need auto-launch after shell connects
	const pendingAgentLaunch = new Set<string>();

	function handleProviderChange(provider: TerminalProvider) {
		// Find an existing tab with this provider — switch to it instead of overwriting
		const existingTab = tabs.find(t => t.provider === provider);
		if (existingTab) {
			terminalStore.switchTab(existingTab.id);
			return;
		}
		// Create a new shell tab — all providers use real shell
		const tabId = terminalStore.createTab(provider);
		if (tabId && provider !== 'shell') {
			// Find the root pane of the new tab and queue agent launch
			const tab = tabs.find(t => t.id === tabId);
			if (tab) {
				pendingAgentLaunch.add(tab.rootPaneId);
			}
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

	// Agent launch commands — cd into businessos/ first so agents read CLAUDE.md
	const BOS_DIR = '~/Desktop/OptimalOS/businessos';
	const AGENT_COMMANDS: Record<string, string> = {
		claude: `cd ${BOS_DIR} && claude --dangerously-skip-permissions\n`,
		codex: `cd ${BOS_DIR} && codex --full-auto\n`,
		ollama: 'ollama run\n',
		osa: `cd ${BOS_DIR} && osa\n`,
	};

	function handleLaunchAgent(agent: string) {
		const command = AGENT_COMMANDS[agent];
		if (!command) return;

		// Use the focused pane, or fall back to the first leaf in the active tab
		const targetPaneId = focusedPaneId ?? getSplitTarget();
		if (!targetPaneId) return;

		const write = shellRefs.get(targetPaneId);
		if (write) {
			write(command);
		}
	}

	function handleStopAgent() {
		const targetPaneId = focusedPaneId ?? getSplitTarget();
		if (!targetPaneId) return;

		const write = shellRefs.get(targetPaneId);
		if (write) {
			// Send Ctrl+C (ASCII 0x03) to interrupt the running agent
			write('\x03');
		}
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
		onStopAgent={handleStopAgent}
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
						environmentMode={currentEnvMode}
						onSessionCreated={handleSessionCreated}
						onFocus={handlePaneFocus}
						onShellReady={handleShellReady}
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
