<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { createTerminalService, type TerminalService } from '$lib/services/terminal.service';
	import { Terminal } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import { SearchAddon } from '@xterm/addon-search';
	import '@xterm/xterm/css/xterm.css';

	let terminalContainer: HTMLDivElement;
	let xterm: Terminal | null = null;
	let fitAddon: FitAddon | null = null;
	let service: TerminalService | null = null;
	let isConnected = $state(false);
	let connectionError = $state<string | null>(null);

	function initTerminal() {
		if (!terminalContainer) return;

		// Create xterm instance
		xterm = new Terminal({
			fontFamily: '"SF Mono", "Monaco", "Inconsolata", "Fira Code", "Courier New", monospace',
			fontSize: 14,
			lineHeight: 1.2,
			cursorBlink: true,
			cursorStyle: 'block',
			convertEol: false, // CRITICAL: Don't convert \n to \r\n - PTY already sends correct sequences
			theme: {
				background: '#1a1a1a',
				foreground: '#00ff00',
				cursor: '#00ff00',
				cursorAccent: '#1a1a1a',
				black: '#000000',
				red: '#ff5555',
				green: '#00ff00',
				yellow: '#ffff55',
				blue: '#5555ff',
				magenta: '#ff55ff',
				cyan: '#00ccff',
				white: '#ffffff',
				brightBlack: '#555555',
				brightRed: '#ff5555',
				brightGreen: '#55ff55',
				brightYellow: '#ffff55',
				brightBlue: '#5555ff',
				brightMagenta: '#ff55ff',
				brightCyan: '#55ffff',
				brightWhite: '#ffffff',
				selectionBackground: '#00ff0033'
			},
			allowProposedApi: true
		});

		// Add addons
		fitAddon = new FitAddon();
		xterm.loadAddon(fitAddon);
		xterm.loadAddon(new WebLinksAddon());
		xterm.loadAddon(new SearchAddon());

		// Open terminal in container
		xterm.open(terminalContainer);

		// Auto-focus terminal for keyboard input (CRITICAL for arrow keys, Enter, etc.)
		setTimeout(() => {
			xterm?.focus();
		}, 0);

		// Fit to container
		setTimeout(() => {
			fitAddon?.fit();
		}, 100);

		// Handle user input
		xterm.onData((data) => {
			if (service?.isConnected()) {
				service.sendInput(data);
			}
		});

		// Handle resize
		xterm.onResize(({ cols, rows }) => {
			if (service?.isConnected()) {
				service.resize(cols, rows);
			}
		});

		// Create terminal service
		service = createTerminalService({
			onData: (data) => {
				xterm?.write(data);
			},
			onConnect: (sessionId, metadata) => {
				isConnected = true;
				connectionError = null;
				console.log('Terminal connected:', sessionId, metadata);
			},
			onDisconnect: () => {
				isConnected = false;
				xterm?.write('\r\n\x1b[31m[Disconnected from terminal]\x1b[0m\r\n');
			},
			onError: (error) => {
				connectionError = error;
				xterm?.write(`\r\n\x1b[31m[Error: ${error}]\x1b[0m\r\n`);
			}
		}, {
			cols: xterm.cols,
			rows: xterm.rows,
			shell: 'zsh'
		});

		// Connect to backend
		service.connect();
	}

	function handleResize() {
		if (fitAddon && xterm) {
			fitAddon.fit();
		}
	}

	function focusTerminal() {
		xterm?.focus();
	}

	onMount(() => {
		initTerminal();
		window.addEventListener('resize', handleResize);

		// Also observe container resize
		const resizeObserver = new ResizeObserver(() => {
			handleResize();
		});
		if (terminalContainer) {
			resizeObserver.observe(terminalContainer);
		}

		return () => {
			resizeObserver.disconnect();
		};
	});

	onDestroy(() => {
		window.removeEventListener('resize', handleResize);
		service?.disconnect();
		xterm?.dispose();
	});
</script>

<div class="terminal-wrapper">
	{#if connectionError}
		<div class="connection-status error">
			Connection Error: {connectionError}
		</div>
	{:else if !isConnected}
		<div class="connection-status connecting">
			Connecting to terminal...
		</div>
	{/if}
	<div
		class="terminal-container"
		bind:this={terminalContainer}
		onclick={focusTerminal}
		role="application"
		tabindex="-1"
	></div>
</div>

<style>
	.terminal-wrapper {
		width: 100%;
		height: 100%;
		background: #1a1a1a;
		position: relative;
		overflow: hidden;
	}

	.terminal-container {
		width: 100%;
		height: 100%;
		padding: 8px;
		box-sizing: border-box;
		outline: none; /* Remove focus outline */
		cursor: text; /* Show text cursor to indicate it's interactive */
	}

	.terminal-container:focus-within {
		/* Optional: Add subtle focus indicator */
		box-shadow: inset 0 0 0 1px rgba(0, 255, 0, 0.2);
	}

	.terminal-container :global(.xterm) {
		height: 100%;
	}

	.terminal-container :global(.xterm-viewport) {
		overflow-y: auto !important;
	}

	.terminal-container :global(.xterm-viewport::-webkit-scrollbar) {
		width: 8px;
	}

	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-track) {
		background: #0a0a0a;
	}

	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-thumb) {
		background: #333;
		border-radius: 4px;
	}

	.connection-status {
		position: absolute;
		top: 8px;
		right: 8px;
		padding: 4px 12px;
		border-radius: 4px;
		font-size: 12px;
		font-family: 'SF Mono', monospace;
		z-index: 10;
	}

	.connection-status.connecting {
		background: #333;
		color: #ffcc00;
	}

	.connection-status.error {
		background: #ff5555;
		color: white;
	}
</style>
