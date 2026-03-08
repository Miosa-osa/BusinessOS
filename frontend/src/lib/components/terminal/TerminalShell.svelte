<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { createTerminalService, type TerminalService } from '$lib/services/terminal.service';
	import { Terminal } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import { SearchAddon } from '@xterm/addon-search';
	import '@xterm/xterm/css/xterm.css';
	import { getThemeColors } from './themes';
	import type { TerminalConfig } from '$lib/stores/terminal/terminalTypes';

	interface Props {
		paneId: string;
		config: TerminalConfig;
		visible?: boolean;
		onSessionCreated?: (paneId: string, sessionId: string) => void;
		onFocus?: (paneId: string) => void;
	}

	let { paneId, config, visible = true, onSessionCreated, onFocus }: Props = $props();

	let terminalContainer = $state<HTMLDivElement | undefined>(undefined);
	let xterm: Terminal | null = null;
	let fitAddon: FitAddon | null = null;
	let service: TerminalService | null = null;
	let resizeObserver: ResizeObserver | null = null;
	let initTimers: ReturnType<typeof setTimeout>[] = [];
	let initialized = false;
	let isConnected = $state(false);
	let connectionError = $state<string | null>(null);

	function initTerminal() {
		if (!terminalContainer || initialized) return;
		initialized = true;

		const themeColors = getThemeColors(config.theme);

		xterm = new Terminal({
			fontFamily: config.fontFamily,
			fontSize: config.fontSize,
			lineHeight: 1.2,
			cursorBlink: config.cursorBlink,
			cursorStyle: config.cursorStyle,
			convertEol: false,
			theme: themeColors,
			allowProposedApi: true
		});

		fitAddon = new FitAddon();
		xterm.loadAddon(fitAddon);
		xterm.loadAddon(new WebLinksAddon());
		xterm.loadAddon(new SearchAddon());

		xterm.open(terminalContainer);

		const t1 = setTimeout(() => xterm?.focus(), 0);
		const t2 = setTimeout(() => fitAddon?.fit(), 100);
		initTimers.push(t1, t2);

		// User input → backend
		xterm.onData((data) => {
			if (service?.isConnected()) {
				service.sendInput(data);
			}
		});

		// Resize → backend
		xterm.onResize(({ cols, rows }) => {
			if (service?.isConnected()) {
				service.resize(cols, rows);
			}
		});

		// Capture service instance for closure safety
		const currentService = createTerminalService({
			onData: (data) => {
				xterm?.write(data);
			},
			onConnect: (sessionId) => {
				isConnected = true;
				connectionError = null;
				onSessionCreated?.(paneId, sessionId);

				const svc = currentService;
				const t3 = setTimeout(() => {
					if (xterm && fitAddon && svc.isConnected()) {
						const dims = fitAddon.proposeDimensions();
						if (dims) svc.resize(dims.cols, dims.rows);
					}
				}, 150);
				initTimers.push(t3);
			},
			onDisconnect: () => {
				isConnected = false;
				xterm?.write('\r\n\x1b[31m[Disconnected]\x1b[0m\r\n');
			},
			onError: (error) => {
				connectionError = error;
				xterm?.write(`\r\n\x1b[31m[Error: ${error}]\x1b[0m\r\n`);
			}
		}, {
			cols: xterm.cols,
			rows: xterm.rows,
			shell: config.cursorStyle ? 'zsh' : 'zsh' // TODO: make configurable
		});

		service = currentService;
		service.connect();
	}

	function handleResize() {
		if (fitAddon && xterm && visible) {
			fitAddon.fit();
		}
	}

	export function focus() {
		xterm?.focus();
		const textarea = terminalContainer?.querySelector('.xterm-helper-textarea') as HTMLTextAreaElement;
		if (textarea) textarea.focus();
	}

	// React to config changes (only theme/font, not cursor blink)
	$effect(() => {
		if (xterm) {
			const themeColors = getThemeColors(config.theme);
			xterm.options.theme = themeColors;
			xterm.options.fontFamily = config.fontFamily;
			xterm.options.fontSize = config.fontSize;
			xterm.options.cursorBlink = config.cursorBlink;
			xterm.options.cursorStyle = config.cursorStyle;
		}
	});

	// React to visibility changes
	$effect(() => {
		if (visible && fitAddon && xterm) {
			setTimeout(() => fitAddon?.fit(), 50);
		}
	});

	onMount(() => {
		initTerminal();
		window.addEventListener('resize', handleResize);

		resizeObserver = new ResizeObserver(() => handleResize());
		if (terminalContainer) {
			resizeObserver.observe(terminalContainer);
		}
	});

	onDestroy(() => {
		// Clear all pending timers
		initTimers.forEach(clearTimeout);
		initTimers = [];

		// Remove global listener
		window.removeEventListener('resize', handleResize);

		// Disconnect resize observer
		resizeObserver?.disconnect();
		resizeObserver = null;

		// Disconnect WebSocket service
		if (service) {
			service.disconnect();
			service = null;
		}

		// Dispose xterm
		if (xterm) {
			xterm.dispose();
			xterm = null;
		}

		fitAddon = null;
		initialized = false;
	});
</script>

<div
	class="terminal-shell"
	class:hidden={!visible}
	onclick={() => { focus(); onFocus?.(paneId); }}
	onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') focus(); }}
	role="application"
	tabindex="-1"
	aria-label="Terminal shell"
>
	{#if connectionError}
		<div class="connection-status error">
			{connectionError}
		</div>
	{:else if !isConnected}
		<div class="connection-status connecting">
			Connecting...
		</div>
	{/if}
	<div
		class="xterm-container"
		bind:this={terminalContainer}
		tabindex="-1"
		aria-label="Terminal content"
	></div>
</div>

<style>
	.terminal-shell {
		width: 100%;
		height: 100%;
		position: relative;
		overflow: hidden;
	}

	.terminal-shell.hidden {
		display: none;
	}

	.xterm-container {
		width: 100%;
		height: 100%;
		padding: 4px 8px;
		box-sizing: border-box;
		outline: none;
		cursor: text;
	}

	.xterm-container :global(.xterm) {
		height: 100%;
	}

	.xterm-container :global(.xterm-viewport) {
		overflow-y: auto !important;
	}

	.xterm-container :global(.xterm-viewport::-webkit-scrollbar) {
		width: 6px;
	}

	.xterm-container :global(.xterm-viewport::-webkit-scrollbar-track) {
		background: transparent;
	}

	.xterm-container :global(.xterm-viewport::-webkit-scrollbar-thumb) {
		background: #333;
		border-radius: 3px;
	}

	.connection-status {
		position: absolute;
		top: 4px;
		right: 8px;
		padding: 2px 8px;
		border-radius: 3px;
		font-size: 11px;
		font-family: 'SF Mono', monospace;
		z-index: 10;
	}

	.connection-status.connecting {
		background: #333;
		color: #ffcc00;
	}

	.connection-status.error {
		background: #ff555533;
		color: #ff5555;
	}
</style>
