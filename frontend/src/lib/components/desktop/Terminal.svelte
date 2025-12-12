<script lang="ts">
	import { onMount } from 'svelte';

	let terminalLines: { type: 'input' | 'output' | 'system' | 'link'; text: string; href?: string }[] = $state([
		{ type: 'system', text: '╔════════════════════════════════════════════════════════════════════════╗' },
		{ type: 'system', text: '║                                                                        ║' },
		{ type: 'system', text: '║    ██████╗ ███████╗     █████╗  ██████╗ ███████╗███╗   ██╗████████╗    ║' },
		{ type: 'system', text: '║   ██╔═══██╗██╔════╝    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝    ║' },
		{ type: 'system', text: '║   ██║   ██║███████╗    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║       ║' },
		{ type: 'system', text: '║   ██║   ██║╚════██║    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║       ║' },
		{ type: 'system', text: '║   ╚██████╔╝███████║    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║       ║' },
		{ type: 'system', text: '║    ╚═════╝ ╚══════╝    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝       ║' },
		{ type: 'system', text: '║                                                                        ║' },
		{ type: 'system', text: '║                 Business OS AI Agent Terminal v1.0                     ║' },
		{ type: 'system', text: '║                                                                        ║' },
		{ type: 'system', text: '╚════════════════════════════════════════════════════════════════════════╝' },
		{ type: 'system', text: '' },
		{ type: 'output', text: 'Welcome to OS Agent Terminal - Your AI Operations Assistant' },
		{ type: 'output', text: '' },
		{ type: 'system', text: '┌───────────────────────────────────────────────────────────────┐' },
		{ type: 'system', text: '│  COMING SOON                                                  │' },
		{ type: 'system', text: '│                                                               │' },
		{ type: 'system', text: '│  OS Agent is under development.                               │' },
		{ type: 'system', text: '│                                                               │' },
		{ type: 'system', text: '│  Features in progress:                                        │' },
		{ type: 'system', text: '│  - Natural language commands                                  │' },
		{ type: 'system', text: '│  - Task automation                                            │' },
		{ type: 'system', text: '│  - Project insights & analytics                               │' },
		{ type: 'system', text: '│  - Smart scheduling assistance                                │' },
		{ type: 'system', text: '│  - Integration with all Business OS modules                   │' },
		{ type: 'system', text: '│                                                               │' },
		{ type: 'system', text: '│  Type "help" for available commands                           │' },
		{ type: 'system', text: '└───────────────────────────────────────────────────────────────┘' },
		{ type: 'output', text: '' },
		{ type: 'link', text: '  Join the waitlist at osa.dev', href: 'https://osa.dev' },
		{ type: 'output', text: '' },
	]);

	let currentInput = $state('');
	let inputElement: HTMLInputElement;
	let terminalContent: HTMLDivElement;

	const availableCommands: Record<string, { description: string; response: string[] }> = {
		help: {
			description: 'Show available commands',
			response: [
				'',
				'Available commands:',
				'  help      - Show this help message',
				'  clear     - Clear the terminal',
				'  about     - About OS Agent',
				'  status    - Check OS Agent status',
				'  waitlist  - Join the OS Agent waitlist',
				'  version   - Show version info',
				'  whoami    - Display current user',
				'  date      - Show current date/time',
				'  echo      - Echo a message',
				''
			]
		},
		waitlist: {
			description: 'Join the OS Agent waitlist',
			response: [
				'',
				'Join the OS Agent waitlist to get early access!',
				'',
				'Visit: https://osa.dev',
				'',
				'Opening waitlist page...',
				''
			]
		},
		about: {
			description: 'About OS Agent',
			response: [
				'',
				'OS Agent - Your AI Operations Assistant',
				'━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━',
				'',
				'OS Agent is your AI-powered assistant for Business OS.',
				'It helps you create applications, integrate with data,',
				'customize your system, and automate your workflow',
				'using natural language.',
				'',
				'Status: Coming Soon',
				''
			]
		},
		status: {
			description: 'Check OS Agent status',
			response: [
				'',
				'┌─ OS Agent Status ────────────────┐',
				'│ Core Engine:    ⏳ In Development │',
				'│ NLP Module:     ⏳ In Development │',
				'│ Task Engine:    ⏳ In Development │',
				'│ Integrations:   ⏳ In Development │',
				'└──────────────────────────────────┘',
				''
			]
		},
		version: {
			description: 'Show version info',
			response: [
				'',
				'OSA Terminal v1.0.0-alpha',
				'Business OS Desktop Environment',
				'Built with Svelte 5 + TypeScript',
				''
			]
		},
		clear: {
			description: 'Clear the terminal',
			response: []
		}
	};

	function handleCommand(command: string) {
		const trimmed = command.trim().toLowerCase();
		const parts = trimmed.split(' ');
		const cmd = parts[0];
		const args = parts.slice(1).join(' ');

		// Add input line
		terminalLines = [...terminalLines, { type: 'input', text: `$ ${command}` }];

		if (cmd === '') {
			// Empty command
			return;
		}

		if (cmd === 'clear') {
			terminalLines = [{ type: 'output', text: 'Terminal cleared.' }, { type: 'output', text: '' }];
			return;
		}

		if (cmd === 'whoami') {
			terminalLines = [...terminalLines, { type: 'output', text: '' }, { type: 'output', text: 'business-os-user' }, { type: 'output', text: '' }];
			return;
		}

		if (cmd === 'date') {
			const now = new Date();
			terminalLines = [...terminalLines, { type: 'output', text: '' }, { type: 'output', text: now.toString() }, { type: 'output', text: '' }];
			return;
		}

		if (cmd === 'echo') {
			terminalLines = [...terminalLines, { type: 'output', text: '' }, { type: 'output', text: args || '' }, { type: 'output', text: '' }];
			return;
		}

		if (cmd === 'waitlist') {
			const responses = availableCommands[cmd].response;
			terminalLines = [...terminalLines, ...responses.map(text => ({ type: 'output' as const, text }))];
			// Open waitlist page
			window.open('https://osa.dev', '_blank');
			return;
		}

		if (availableCommands[cmd]) {
			const responses = availableCommands[cmd].response;
			terminalLines = [...terminalLines, ...responses.map(text => ({ type: 'output' as const, text }))];
			return;
		}

		// Unknown command - simulate AI response coming soon
		terminalLines = [
			...terminalLines,
			{ type: 'output', text: '' },
			{ type: 'system', text: `OS Agent: I received "${command}"` },
			{ type: 'system', text: '   AI processing is coming soon!' },
			{ type: 'system', text: '   Type "help" for available commands.' },
			{ type: 'output', text: '' }
		];
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleCommand(currentInput);
			currentInput = '';
		}
	}

	function focusInput() {
		inputElement?.focus();
	}

	$effect(() => {
		// Auto-scroll to bottom when new lines are added
		// Access terminalLines.length to create a dependency
		const _len = terminalLines.length;
		if (terminalContent) {
			// Use requestAnimationFrame to ensure DOM has updated
			requestAnimationFrame(() => {
				terminalContent.scrollTop = terminalContent.scrollHeight;
			});
		}
	});

	onMount(() => {
		focusInput();
		// Initial scroll to bottom
		if (terminalContent) {
			terminalContent.scrollTop = terminalContent.scrollHeight;
		}
	});
</script>

<div class="terminal" onclick={focusInput} role="application" aria-label="OSA Terminal">
	<div class="terminal-content" bind:this={terminalContent}>
		{#each terminalLines as line}
			<div class="terminal-line" class:input={line.type === 'input'} class:system={line.type === 'system'} class:link={line.type === 'link'}>
				{#if line.type === 'link' && line.href}
					<a href={line.href} target="_blank" rel="noopener noreferrer" class="terminal-link">{line.text}</a>
				{:else if line.text}
					<pre>{line.text}</pre>
				{:else}
					<br/>
				{/if}
			</div>
		{/each}

		<div class="terminal-input-line">
			<span class="prompt">$</span>
			<input
				bind:this={inputElement}
				bind:value={currentInput}
				onkeydown={handleKeyDown}
				type="text"
				class="terminal-input"
				spellcheck="false"
				autocomplete="off"
				aria-label="Terminal input"
			/>
		</div>
	</div>
</div>

<style>
	.terminal {
		width: 100%;
		height: 100%;
		background: #1a1a1a;
		font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', 'Courier New', monospace;
		font-size: 13px;
		color: #00ff00;
		overflow: hidden;
		cursor: text;
	}

	.terminal-content {
		height: 100%;
		overflow-y: auto;
		padding: 16px;
	}

	.terminal-content::-webkit-scrollbar {
		width: 8px;
	}

	.terminal-content::-webkit-scrollbar-track {
		background: #0a0a0a;
	}

	.terminal-content::-webkit-scrollbar-thumb {
		background: #333;
		border-radius: 4px;
	}

	.terminal-line {
		line-height: 1.4;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.terminal-line pre {
		margin: 0;
		font-family: inherit;
		font-size: inherit;
	}

	.terminal-line.input {
		color: #00ff00;
		background: transparent;
	}

	.terminal-line.input pre {
		background: transparent;
		color: #00ff00;
	}

	.terminal-line.system {
		color: #00ccff;
	}

	.terminal-line.link {
		color: #ffcc00;
	}

	.terminal-link {
		color: #ffcc00;
		text-decoration: underline;
		cursor: pointer;
	}

	.terminal-link:hover {
		color: #ffdd44;
		text-decoration: underline;
	}

	.terminal-input-line {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-top: 4px;
		background: transparent;
	}

	.prompt {
		color: #00ff00;
		font-weight: bold;
		user-select: none;
	}

	.terminal-input {
		flex: 1;
		background: transparent !important;
		background-color: transparent !important;
		border: none !important;
		outline: none !important;
		box-shadow: none !important;
		color: #00ff00;
		font-family: inherit;
		font-size: inherit;
		caret-color: #00ff00;
		padding: 0;
		margin: 0;
		-webkit-appearance: none;
		-moz-appearance: none;
		appearance: none;
		border-radius: 0;
	}

	.terminal-input:focus {
		background: transparent !important;
		border: none !important;
		outline: none !important;
		box-shadow: none !important;
	}

	.terminal-input::placeholder {
		color: #555;
	}

</style>
