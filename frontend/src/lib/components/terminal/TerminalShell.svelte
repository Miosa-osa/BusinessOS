<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { createTerminalService, getCloudTerminalSession, type TerminalService } from '$lib/services/terminal.service';
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
		environmentMode?: string;
		onSessionCreated?: (paneId: string, sessionId: string) => void;
		onFocus?: (paneId: string) => void;
	}

	let { paneId, config, visible = true, environmentMode = 'local', onSessionCreated, onFocus }: Props = $props();

	// Cloud computer URL — resolved when switching to production mode
	let cloudUrl = $state<string | null>(null);

	let terminalContainer = $state<HTMLDivElement | undefined>(undefined);
	let xterm: Terminal | null = null;
	let fitAddon: FitAddon | null = null;
	let service: TerminalService | null = null;
	let resizeObserver: ResizeObserver | null = null;
	let initTimers: ReturnType<typeof setTimeout>[] = [];
	let initialized = false;
	let isConnected = $state(false);
	let connectionError = $state<string | null>(null);

	async function initTerminal() {
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

		// Input/resize handlers are registered AFTER the connection is established
		// to avoid duplicate handlers between cloud and local modes.

		// Cloud mode: get a pre-authenticated WebSocket URL from the backend
		let cloudWsUrl: string | undefined = undefined;
		if (environmentMode === 'production') {
			xterm?.write('\x1b[36m[Connecting to cloud computer...]\x1b[0m\r\n');
			try {
				const session = await getCloudTerminalSession();
				if (session?.mode === 'cloud' && session.ws_url) {
					cloudWsUrl = session.ws_url;
					cloudUrl = session.ws_url;
					xterm?.write(`\x1b[36m[Session: ${session.session_id?.slice(0, 8)}... on ${session.slug}]\x1b[0m\r\n`);
				} else {
					const msg = session?.message || 'No active cloud computer';
					xterm?.write(`\x1b[33m[${msg} — falling back to local]\x1b[0m\r\n`);
				}
			} catch {
				xterm?.write('\x1b[33m[Could not reach cloud — falling back to local]\x1b[0m\r\n');
			}
		}

		if (cloudWsUrl) {
			// Direct WebSocket to cloud VM — bypass the normal terminal service
			try {
				const ws = new WebSocket(cloudWsUrl);
				ws.binaryType = 'arraybuffer';
				ws.onopen = () => {
					isConnected = true;
					connectionError = null;
					xterm?.write('\x1b[32m[Connected to cloud computer]\x1b[0m\r\n');
					onSessionCreated?.(paneId, 'cloud');
				};
				ws.onmessage = (event) => {
					if (event.data instanceof ArrayBuffer) {
						xterm?.write(new Uint8Array(event.data));
					} else {
						xterm?.write(event.data);
					}
				};
				ws.onclose = () => {
					isConnected = false;
					xterm?.write('\r\n\x1b[31m[Cloud terminal disconnected]\x1b[0m\r\n');
				};
				ws.onerror = () => {
					connectionError = 'Cloud terminal connection failed';
					xterm?.write('\r\n\x1b[31m[Cloud terminal error]\x1b[0m\r\n');
				};
				// Store reference for cleanup (input handler registered below, after branching)
				service = {
					isConnected: () => ws.readyState === WebSocket.OPEN,
					sendInput: (data: string) => { if (ws.readyState === WebSocket.OPEN) ws.send(data); },
					resize: (cols: number, rows: number) => {
						if (ws.readyState === WebSocket.OPEN) {
							ws.send(JSON.stringify({ type: 'resize', cols, rows }));
						}
					},
					disconnect: () => ws.close(),
					connect: () => {},
				} as unknown as TerminalService;
			} catch (err) {
				xterm?.write(`\r\n\x1b[31m[Failed to connect: ${err}]\x1b[0m\r\n`);
			}
		} else {
			// Local mode: use the standard terminal service
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
				shell: config.cursorStyle ? 'zsh' : 'zsh',
			});

			service = currentService;
			service.connect();
		}

		// Register unified input/resize handlers (works for both cloud and local)
		xterm.onData((data) => {
			if (service?.isConnected()) {
				service.sendInput(data);
			}
		});

		xterm.onResize(({ cols, rows }) => {
			if (service?.isConnected()) {
				service.resize(cols, rows);
			}
		});
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

	export function writeData(data: string) {
		if (service?.isConnected()) {
			service.sendInput(data);
		}
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

	// Mode switching is handled by {#key environmentMode} in TerminalPane —
	// it destroys and recreates this component on mode change, so no $effect needed.

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
