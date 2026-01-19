<script lang="ts">
	interface Props {
		userMessage?: string;
		osaMessage?: string;
		isListening?: boolean;
		isSpeaking?: boolean;
	}

	let { userMessage = '', osaMessage = '', isListening = false, isSpeaking = false }: Props = $props();
</script>

{#if userMessage || osaMessage}
	<div class="live-captions">
		{#if userMessage}
			<div class="caption user-caption">
				<span class="caption-label">You:</span>
				<span class="caption-text">{userMessage}</span>
			</div>
		{/if}
		{#if osaMessage}
			<div class="caption osa-caption">
				<span class="caption-label">OSA:</span>
				<span class="caption-text">{osaMessage}</span>
			</div>
		{/if}
	</div>
{/if}

<style>
	.live-captions {
		position: fixed;
		bottom: 250px;
		left: 50%;
		transform: translateX(-50%);
		max-width: 600px;
		width: 90%;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		gap: 12px;
		pointer-events: none;
	}

	.caption {
		/* Glassmorphism pill-shaped design */
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		border: 1px solid rgba(255, 255, 255, 0.2);
		padding: 14px 24px;
		border-radius: 9999px; /* Pill shape */
		display: flex;
		gap: 12px;
		align-items: center;
		animation: slideIn 0.3s ease-out;
		box-shadow:
			0 8px 32px rgba(0, 0, 0, 0.2),
			inset 0 1px 0 rgba(255, 255, 255, 0.1);
	}

	.user-caption {
		background: rgba(96, 165, 250, 0.15);
		border: 1px solid rgba(96, 165, 250, 0.3);
	}

	.osa-caption {
		background: rgba(167, 139, 250, 0.15);
		border: 1px solid rgba(167, 139, 250, 0.3);
	}

	.caption-label {
		font-weight: 600;
		font-size: 13px;
		flex-shrink: 0;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.user-caption .caption-label {
		color: #60a5fa;
	}

	.osa-caption .caption-label {
		color: #a78bfa;
	}

	.caption-text {
		color: rgba(255, 255, 255, 0.95);
		font-size: 14px;
		line-height: 1.5;
		font-weight: 500;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
