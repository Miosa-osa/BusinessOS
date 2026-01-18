<!--
	Live Captions Overlay

	Displays real-time voice transcriptions and command recognition feedback
	for the 3D Desktop active listening system.

	Features:
	- Real-time caption display
	- Command recognition feedback
	- Auto-fade after silence
	- Minimal, unobtrusive design
-->

<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	interface Props {
		/** User's spoken message */
		userMessage?: string;

		/** OSA's response message */
		osaMessage?: string;

		/** Whether actively listening */
		isListening: boolean;

		/** Whether OSA is speaking */
		isSpeaking?: boolean;
	}

	let { userMessage = '', osaMessage = '', command = null, isListening = false, isSpeaking = false }: Props = $props();

	// Simplified - messages are controlled by parent component timing

	// Get command feedback message
	function getCommandFeedback(cmd: VoiceCommand): string {
		if (!cmd) return '';

		switch (cmd.type) {
			case 'enter_edit_mode':
				return '✓ Entering edit mode';
			case 'exit_edit_mode':
				return '✓ Exiting edit mode';
			case 'save_layout':
				return `✓ Saving layout: ${cmd.name}`;
			case 'load_layout':
				return `✓ Loading layout: ${cmd.name}`;
			case 'focus_module':
				return `✓ Opening ${cmd.module}`;
			case 'close_module':
				return `✓ Closing ${cmd.module}`;
			case 'unfocus':
				return '✓ Unfocusing window';
			case 'switch_view':
				return `✓ Switching to ${cmd.view} view`;
			case 'toggle_auto_rotate':
				return '✓ Toggling auto-rotate';
			case 'zoom_in':
				return '✓ Zooming camera in';
			case 'zoom_out':
				return '✓ Zooming camera out';
			case 'reset_zoom':
				return '✓ Resetting zoom';
			case 'expand_orb':
				return '✓ Expanding orb';
			case 'contract_orb':
				return '✓ Contracting orb';
			case 'resize_window':
				return `✓ Making window ${cmd.direction}`;
			case 'next_window':
				return '✓ Next window';
			case 'previous_window':
				return '✓ Previous window';
			case 'help':
				return '✓ Showing help';
			case 'unknown':
				return `✗ Unknown command: "${cmd.text}"`;
			default:
				return '';
		}
	}

</script>

<div class="live-captions-container">
	<!-- User message - inline label -->
	{#if userMessage}
		<div class="user-message" transition:fly={{ y: 20, duration: 300 }}>
			<span class="message-label">You:</span> {userMessage}
		</div>
	{/if}

	<!-- OSA message - inline label -->
	{#if osaMessage}
		<div class="osa-message" transition:fly={{ y: 20, duration: 300 }}>
			<span class="message-label">OSA:</span> {osaMessage}
		</div>
	{/if}

	<!-- Command feedback -->
	{#if command}
		<div class="command-feedback" transition:fly={{ y: 20, duration: 300 }}>
			{getCommandFeedback(command)}
		</div>
	{/if}
</div>

<style>
	.live-captions-container {
		position: fixed;
		bottom: 120px; /* Above the dock */
		left: 50%;
		transform: translateX(-50%);
		z-index: 500;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		pointer-events: none; /* Don't interfere with clicks */
	}

	/* ===== USER MESSAGE ===== */
	.user-message {
		max-width: 900px; /* Wider for better readability */
		max-height: calc(100vh - 200px); /* INCREASED: Use most of viewport height (was 400px) */
		overflow-y: auto; /* Enable scrolling for very long messages */
		padding: 16px 28px; /* More padding for pill shape */
		background: rgba(255, 255, 255, 0.15); /* Glassy transparent */
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 50px; /* Pill-shaped */
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
	}

	/* ===== OSA MESSAGE ===== */
	.osa-message {
		max-width: 900px; /* Wider for better readability */
		max-height: calc(100vh - 200px); /* INCREASED: Use most of viewport height (was 400px) */
		overflow-y: auto; /* Enable scrolling for very long messages */
		padding: 16px 28px; /* More padding for pill shape */
		background: rgba(255, 255, 255, 0.15); /* Glassy transparent */
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 50px; /* Pill-shaped */
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
	}

	/* Inline label styles */
	.message-label {
		color: rgba(0, 0, 0, 0.6); /* Dark semi-transparent for glass */
		font-size: 15px;
		font-weight: 700;
		margin-right: 6px;
	}

	/* Message container text */
	.user-message,
	.osa-message {
		color: rgba(0, 0, 0, 0.9); /* Black text on glass */
		font-size: 15px;
		line-height: 1.5;
		font-weight: 500;
		word-wrap: break-word;
		white-space: pre-wrap;
		overflow-wrap: break-word;
	}

	/* ===== COMMAND FEEDBACK ===== */
	.command-feedback {
		padding: 10px 18px;
		background: rgba(34, 197, 94, 0.9); /* Green for success */
		backdrop-filter: blur(12px);
		border-radius: 20px;
		color: white;
		font-size: 14px;
		font-weight: 500;
		box-shadow: 0 4px 12px rgba(34, 197, 94, 0.3);
		animation: commandPulse 0.5s ease-out;
	}

	/* Unknown command gets red background */
	.command-feedback:has(:global(.unknown)) {
		background: rgba(239, 68, 68, 0.9); /* Red */
		box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
	}

	@keyframes commandPulse {
		0% {
			transform: scale(0.95);
			opacity: 0;
		}
		50% {
			transform: scale(1.05);
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}

	/* ===== DARK MODE STYLES ===== */
	:global(.dark) .user-message,
	:global(.dark) .osa-message {
		background: rgba(30, 30, 30, 0.3);
		color: rgba(255, 255, 255, 0.9);
	}

	:global(.dark) .message-label {
		color: rgba(255, 255, 255, 0.6);
	}
</style>
