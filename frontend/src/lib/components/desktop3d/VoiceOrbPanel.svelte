<!--
	Voice Cloud Panel - Clean Draggable Cloud

	- Simple cloud icon for voice activation
	- Subtle glow: blue (listening) → purple (speaking)
	- Draggable anywhere on screen
	- Position saved to localStorage
-->

<script lang="ts">
	import { scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { browser } from '$app/environment';

	interface Props {
		isListening: boolean;
		isSpeaking: boolean;
		onToggleListening: () => void;
	}

	let { isListening, isSpeaking, onToggleListening }: Props = $props();

	// Drag state
	let isDragging = $state(false);
	let dragOffset = $state({ x: 0, y: 0 });
	let hasMoved = $state(false); // Track if actually dragged vs just clicked
	let startPos = $state({ x: 0, y: 0 });

	// Position state - load from localStorage or use defaults
	let position = $state({ x: 0, y: 0 });
	let useCustomPosition = $state(false);

	// Load saved position on mount
	$effect(() => {
		if (browser) {
			const saved = localStorage.getItem('voiceCloudPosition');
			if (saved) {
				try {
					const parsed = JSON.parse(saved);
					position = parsed;
					useCustomPosition = true;
				} catch (e) {
					console.warn('[VoiceCloud] Failed to parse saved position');
				}
			}
		}
	});

	// Save position when it changes
	function savePosition() {
		if (browser && useCustomPosition) {
			localStorage.setItem('voiceCloudPosition', JSON.stringify(position));
		}
	}

	// Drag handlers - drag the whole cloud, but detect click vs drag
	function handleDragStart(e: MouseEvent) {
		// Don't start drag if clicking reset button
		if ((e.target as HTMLElement).closest('.reset-position')) {
			return;
		}
		isDragging = true;
		hasMoved = false;
		startPos = { x: e.clientX, y: e.clientY };
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		dragOffset = {
			x: e.clientX - rect.left,
			y: e.clientY - rect.top
		};
		e.preventDefault();
	}

	function handleDragMove(e: MouseEvent) {
		if (!isDragging) return;

		// Check if mouse moved more than 5px (threshold for drag vs click)
		const dx = Math.abs(e.clientX - startPos.x);
		const dy = Math.abs(e.clientY - startPos.y);
		if (dx > 5 || dy > 5) {
			hasMoved = true;
		}

		if (!hasMoved) return; // Don't move until threshold exceeded

		const newX = e.clientX - dragOffset.x;
		const newY = e.clientY - dragOffset.y;

		// Constrain to viewport (accounting for cloud size 180x144)
		const maxX = window.innerWidth - 190;
		const maxY = window.innerHeight - 160;

		position = {
			x: Math.max(0, Math.min(newX, maxX)),
			y: Math.max(0, Math.min(newY, maxY))
		};
		useCustomPosition = true;
	}

	function handleDragEnd() {
		if (isDragging) {
			isDragging = false;
			if (hasMoved) {
				// Was a drag - save position
				savePosition();
			} else {
				// Was a click - toggle listening
				onToggleListening();
			}
			hasMoved = false;
		}
	}

	// Reset position to default
	function resetPosition() {
		useCustomPosition = false;
		position = { x: 0, y: 0 };
		if (browser) {
			localStorage.removeItem('voiceCloudPosition');
		}
	}
</script>

<svelte:window onmousemove={handleDragMove} onmouseup={handleDragEnd} />

<div
	class="voice-cloud"
	class:active={isListening || isSpeaking}
	class:dragging={isDragging}
	class:custom-position={useCustomPosition}
	style={useCustomPosition ? `left: ${position.x}px; top: ${position.y}px;` : ''}
	onmousedown={handleDragStart}
>
	<!-- Reset Position Button (only show when custom position) -->
	{#if useCustomPosition}
		<button class="reset-position" onclick={resetPosition} title="Reset to default position">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
				<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
				<path d="M3 3v5h5" />
			</svg>
		</button>
	{/if}

	<!-- Cloud Button -->
	<button
		class="cloud-button"
		class:listening={isListening}
		class:speaking={isSpeaking}
		title={isListening ? 'Stop listening' : 'Tap to speak'}
		transition:scale={{ duration: 400, easing: cubicOut }}
	>
		<!-- Cloud Icon -->
		<div class="cloud-icon">
			<img src="/Cloudpngosa.png" alt="OSA Cloud" class="cloud-image" />
		</div>

	</button>
</div>

<style>
	/* Main container */
	.voice-cloud {
		position: fixed;
		bottom: 130px;
		right: 30px;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		align-items: center;
		cursor: grab;
	}

	/* When custom position is set, use absolute positioning from top-left */
	.voice-cloud.custom-position {
		bottom: auto;
		right: auto;
	}

	/* Dragging state */
	.voice-cloud.dragging {
		cursor: grabbing;
		user-select: none;
	}

	.voice-cloud.dragging .cloud-button {
		pointer-events: none;
	}

	/* Reset Position Button */
	.reset-position {
		position: absolute;
		top: -5px;
		right: -5px;
		width: 22px;
		height: 22px;
		border: none;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.6);
		color: #a1a1aa;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: all 0.2s ease;
		z-index: 10;
	}

	.voice-cloud:hover .reset-position {
		opacity: 0.8;
	}

	.reset-position:hover {
		opacity: 1 !important;
		color: white;
		background: rgba(239, 68, 68, 0.8);
		transform: scale(1.1);
	}

	/* Cloud Button - BIGGER SIZE */
	.cloud-button {
		position: relative;
		width: 180px;
		height: 144px;
		border: none;
		background: transparent;
		cursor: grab;
		transition: transform 0.2s ease;
		outline: none;
	}

	.voice-cloud.dragging .cloud-button {
		cursor: grabbing;
	}

	.cloud-button:hover {
		transform: scale(1.05);
	}

	.cloud-button:active {
		transform: scale(0.95);
	}

	/* Cloud Icon */
	.cloud-icon {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cloud-image {
		width: 100%;
		height: 100%;
		object-fit: contain;
		filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.15));
		transition: all 0.3s ease;
		opacity: 0.9;
	}

	/* Listening State - Soft blue glow + very slow breathing */
	.cloud-button.listening .cloud-image {
		filter: drop-shadow(0 0 12px rgba(59, 130, 246, 0.5))
		        drop-shadow(0 0 25px rgba(59, 130, 246, 0.3));
		animation: listening-breathe 3s cubic-bezier(0.4, 0, 0.6, 1) infinite;
		opacity: 1;
	}

	/* Speaking State - Warm purple glow + organic floating */
	.cloud-button.speaking .cloud-image {
		filter: drop-shadow(0 0 20px rgba(168, 85, 247, 0.6))
		        drop-shadow(0 0 40px rgba(168, 85, 247, 0.35))
		        brightness(1.05);
		animation: speaking-organic 4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
		opacity: 1;
	}

	/* Listening - very slow, gentle breathing like meditation */
	@keyframes listening-breathe {
		0%, 100% {
			transform: scale(1);
			filter: drop-shadow(0 0 12px rgba(59, 130, 246, 0.5))
			        drop-shadow(0 0 25px rgba(59, 130, 246, 0.3));
		}
		50% {
			transform: scale(1.04);
			filter: drop-shadow(0 0 18px rgba(59, 130, 246, 0.6))
			        drop-shadow(0 0 35px rgba(59, 130, 246, 0.4));
		}
	}

	/* Speaking - organic, flowing movement like a cloud drifting */
	@keyframes speaking-organic {
		0% {
			transform: scale(1) translateY(0) rotate(0deg);
		}
		25% {
			transform: scale(1.03) translateY(-2px) rotate(0.5deg);
		}
		50% {
			transform: scale(1.02) translateY(-3px) rotate(-0.5deg);
		}
		75% {
			transform: scale(1.04) translateY(-1px) rotate(0.3deg);
		}
		100% {
			transform: scale(1) translateY(0) rotate(0deg);
		}
	}

	/* Responsive */
	@media (max-width: 768px) {
		.voice-cloud {
			bottom: 100px;
			right: 20px;
		}

		.cloud-button {
			width: 120px;
			height: 96px;
		}
	}
</style>
