<!--
	Voice Control Panel - Cloud Style

	Beautiful cloud-like voice interface with:
	- Fluffy cloud design with tap-to-speak
	- Animated when listening/speaking
	- Interactive particles and glow effects
	- Interrupt capability when user speaks over OSA
-->

<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		isListening: boolean;
		isSpeaking: boolean;
		onToggleListening: () => void;
	}

	let { isListening, isSpeaking, onToggleListening }: Props = $props();

	// Audio visualization state
	let audioLevels = $state<number[]>([0, 0, 0, 0, 0]);
	let particles = $state<{ x: number; y: number; delay: number; scale: number }[]>([]);

	// Generate particles for visual effect
	$effect(() => {
		if (isListening || isSpeaking) {
			// Create floating particles
			const newParticles = Array.from({ length: 8 }, (_, i) => ({
				x: Math.random() * 100,
				y: Math.random() * 100,
				delay: Math.random() * 2,
				scale: 0.5 + Math.random() * 0.5
			}));
			particles = newParticles;

			// Animate audio levels
			const interval = setInterval(() => {
				audioLevels = Array.from({ length: 5 }, () => {
					return isSpeaking
						? 30 + Math.random() * 70  // More energetic for speaking
						: 20 + Math.random() * 50; // Calmer for listening
				});
			}, 150);

			return () => {
				clearInterval(interval);
				particles = [];
				audioLevels = [0, 0, 0, 0, 0];
			};
		} else {
			particles = [];
			audioLevels = [0, 0, 0, 0, 0];
		}
	});
</script>

<div class="voice-cloud" class:active={isListening || isSpeaking}>
	<!-- Cloud Button -->
	<button
		class="cloud-button"
		class:listening={isListening}
		class:speaking={isSpeaking}
		onclick={onToggleListening}
		title={isListening ? 'Stop listening' : 'Tap to speak'}
		transition:scale={{ duration: 400, easing: cubicOut }}
	>
		<!-- Cloud Shape (multiple layers for depth) -->
		<div class="cloud-layer cloud-back"></div>
		<div class="cloud-layer cloud-middle"></div>
		<div class="cloud-layer cloud-front"></div>

		<!-- Floating Particles -->
		{#if isListening || isSpeaking}
			{#each particles as particle (particle.x + particle.y)}
				<div
					class="particle"
					style="left: {particle.x}%; top: {particle.y}%; animation-delay: {particle.delay}s; transform: scale({particle.scale});"
					transition:fade={{ duration: 300 }}
				></div>
			{/each}
		{/if}

		<!-- Center Icon -->
		<div class="cloud-icon">
			{#if isListening}
				<!-- Microphone (listening) -->
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
					/>
				</svg>
			{:else if isSpeaking}
				<!-- Sound waves (OSA speaking) -->
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<circle cx="12" cy="12" r="2" fill="currentColor" />
					<path
						stroke-linecap="round"
						d="M8 12a4 4 0 018 0M5 12a7 7 0 0114 0M2 12a10 10 0 0120 0"
					/>
				</svg>
			{:else}
				<!-- Cloud with sparkle (inactive) -->
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064"
					/>
				</svg>
			{/if}
		</div>

		<!-- Audio Level Bars (show when active) -->
		{#if isListening || isSpeaking}
			<div class="audio-bars" transition:fade={{ duration: 300 }}>
				{#each audioLevels as level, i}
					<div
						class="audio-bar"
						style="height: {level}%; animation-delay: {i * 0.1}s;"
					></div>
				{/each}
			</div>
		{/if}

		<!-- Glow effect when active -->
		{#if isListening || isSpeaking}
			<div class="cloud-glow" transition:fade={{ duration: 500 }}></div>
		{/if}
	</button>

	<!-- Status Label -->
	<div class="status-label" transition:fade={{ duration: 200 }}>
		{#if isListening}
			<span class="status-listening">Listening...</span>
		{:else if isSpeaking}
			<span class="status-speaking">OSA speaking...</span>
		{:else}
			<span class="status-idle">Tap to speak</span>
		{/if}
	</div>
</div>

<style>
	/* Main container */
	.voice-cloud {
		position: fixed;
		bottom: 130px;
		right: 30px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		z-index: 600;
		pointer-events: auto;
	}

	/* Cloud Button */
	.cloud-button {
		position: relative;
		width: 100px;
		height: 100px;
		border: none;
		background: transparent;
		cursor: pointer;
		transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		display: flex;
		align-items: center;
		justify-content: center;
		filter: drop-shadow(0 4px 20px rgba(0, 0, 0, 0.15));
	}

	.cloud-button:hover {
		transform: scale(1.08) translateY(-2px);
	}

	.cloud-button:active {
		transform: scale(0.95);
	}

	/* Cloud Layers (fluffy cloud effect) */
	.cloud-layer {
		position: absolute;
		background: linear-gradient(135deg, #ffffff 0%, #f0f4f8 100%);
		border-radius: 50%;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.cloud-back {
		width: 60px;
		height: 60px;
		top: 25%;
		left: 10%;
		opacity: 0.6;
	}

	.cloud-middle {
		width: 80px;
		height: 70px;
		top: 15%;
		left: 15%;
		opacity: 0.8;
	}

	.cloud-front {
		width: 70px;
		height: 65px;
		top: 20%;
		right: 15%;
	}

	/* Active states - cloud changes color */
	.cloud-button.listening .cloud-layer {
		background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
		box-shadow: 0 0 30px rgba(59, 130, 246, 0.4);
		animation: cloud-pulse-listening 2s ease-in-out infinite;
	}

	.cloud-button.speaking .cloud-layer {
		background: linear-gradient(135deg, #c084fc 0%, #a855f7 100%);
		box-shadow: 0 0 30px rgba(168, 85, 247, 0.4);
		animation: cloud-pulse-speaking 1.5s ease-in-out infinite;
	}

	@keyframes cloud-pulse-listening {
		0%,
		100% {
			transform: scale(1);
			opacity: 1;
		}
		50% {
			transform: scale(1.05);
			opacity: 0.9;
		}
	}

	@keyframes cloud-pulse-speaking {
		0%,
		100% {
			transform: scale(1) rotate(0deg);
		}
		25% {
			transform: scale(1.03) rotate(-1deg);
		}
		75% {
			transform: scale(1.03) rotate(1deg);
		}
	}

	/* Cloud Glow (background radial glow) */
	.cloud-glow {
		position: absolute;
		width: 140%;
		height: 140%;
		border-radius: 50%;
		z-index: -1;
		pointer-events: none;
	}

	.cloud-button.listening .cloud-glow {
		background: radial-gradient(circle, rgba(59, 130, 246, 0.3) 0%, transparent 70%);
		animation: glow-pulse 2s ease-in-out infinite;
	}

	.cloud-button.speaking .cloud-glow {
		background: radial-gradient(circle, rgba(168, 85, 247, 0.3) 0%, transparent 70%);
		animation: glow-pulse 1.5s ease-in-out infinite;
	}

	@keyframes glow-pulse {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.6;
		}
		50% {
			transform: scale(1.2);
			opacity: 1;
		}
	}

	/* Center Icon */
	.cloud-icon {
		position: relative;
		z-index: 10;
		width: 40px;
		height: 40px;
		color: #64748b;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cloud-button.listening .cloud-icon,
	.cloud-button.speaking .cloud-icon {
		color: white;
	}

	.cloud-icon svg {
		width: 100%;
		height: 100%;
		filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
	}

	/* Floating Particles */
	.particle {
		position: absolute;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: white;
		opacity: 0.6;
		pointer-events: none;
		animation: particle-float 3s ease-in-out infinite;
		box-shadow: 0 0 8px rgba(255, 255, 255, 0.6);
	}

	@keyframes particle-float {
		0%,
		100% {
			transform: translateY(0) scale(1);
			opacity: 0.3;
		}
		50% {
			transform: translateY(-20px) scale(1.2);
			opacity: 0.8;
		}
	}

	/* Audio Bars (sound visualization) */
	.audio-bars {
		position: absolute;
		bottom: -18px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		gap: 4px;
		align-items: flex-end;
		height: 30px;
	}

	.audio-bar {
		width: 4px;
		background: linear-gradient(to top, rgba(255, 255, 255, 0.4), rgba(255, 255, 255, 0.9));
		border-radius: 2px;
		transition: height 0.15s ease-out;
		animation: bar-bounce 0.8s ease-in-out infinite;
	}

	.cloud-button.listening .audio-bar {
		background: linear-gradient(to top, rgba(59, 130, 246, 0.5), rgba(59, 130, 246, 1));
	}

	.cloud-button.speaking .audio-bar {
		background: linear-gradient(to top, rgba(168, 85, 247, 0.5), rgba(168, 85, 247, 1));
	}

	@keyframes bar-bounce {
		0%,
		100% {
			transform: scaleY(0.8);
		}
		50% {
			transform: scaleY(1.2);
		}
	}

	/* Status Label */
	.status-label {
		padding: 8px 16px;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(12px);
		border-radius: 20px;
		font-size: 13px;
		font-weight: 500;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		white-space: nowrap;
		border: 1px solid rgba(0, 0, 0, 0.05);
	}

	.status-listening {
		color: #3b82f6;
		animation: text-pulse 2s ease-in-out infinite;
	}

	.status-speaking {
		color: #a855f7;
		animation: text-pulse 1.5s ease-in-out infinite;
	}

	.status-idle {
		color: #64748b;
	}

	@keyframes text-pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.7;
		}
	}

	/* Dark mode support */
	:global(.dark) .cloud-layer {
		background: linear-gradient(135deg, #2c2c2e 0%, #1c1c1e 100%);
	}

	:global(.dark) .cloud-icon {
		color: #aaaaaa;
	}

	:global(.dark) .status-label {
		background: rgba(30, 30, 30, 0.95);
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .status-idle {
		color: #aaaaaa;
	}

	/* Mobile responsive */
	@media (max-width: 640px) {
		.voice-cloud {
			bottom: 100px;
			right: 20px;
		}

		.cloud-button {
			width: 80px;
			height: 80px;
		}

		.cloud-icon {
			width: 32px;
			height: 32px;
		}
	}
</style>
