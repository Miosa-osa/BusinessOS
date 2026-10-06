<!--
	OsaOrb.svelte
	Shared OSA video orb - used on both regular desktop and 3D desktop.

	- Always playing, never static — looks like a continuous animation
	- Idle: full video loops naturally (no fade, no crossfade)
	- Active (listening/speaking): steady video loop with blue glow
	- Draggable anywhere, position saved to localStorage
	- Live captions: shows user speech (blue) and OSA responses (purple)
-->

<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { browser } from '$app/environment';
	import { onMount, onDestroy } from 'svelte';
	import { OsaOrbVoice, type OrbVoicePhase } from '$lib/services/osaOrbVoice';
	

	interface CaptionMessage {
		id: number;
		sender: 'user' | 'osa';
		text: string;
	}

	interface Props {
		isListening?: boolean;
		isSpeaking?: boolean;
		onToggleListening?: (() => void) | null;
		/** Current transcript text (external mode) */
		transcript?: string;
		/** OSA response text (external mode) */
		osaMessage?: string;
	}

	let {
		isListening: externalListening = false,
		isSpeaking: externalSpeaking = false,
		onToggleListening = null,
		transcript: externalTranscript = '',
		osaMessage: externalOsaMessage = ''
	}: Props = $props();

	let phase = $state<OrbVoicePhase>('idle');
	let voiceError = $state('');
	let isListening = $derived(onToggleListening ? externalListening : phase === 'listening');
	let isSpeaking = $derived(onToggleListening ? externalSpeaking : phase === 'speaking');
	let liveTranscript = $state('');
	let currentTranscript = $derived(onToggleListening ? externalTranscript : liveTranscript);
	const captionTimers = new Set<ReturnType<typeof setTimeout>>();
	const voice = new OsaOrbVoice({
		phase: value => { phase = value; if (value !== 'error') voiceError = ''; },
		caption: (sender, text) => addCaption(sender, text),
        transcript: text => { liveTranscript = text; },
		error: message => { voiceError = message; }
	});

	// Conversation captions
	let captions = $state<CaptionMessage[]>([]);
	let captionIdCounter = 0;

	function addCaption(sender: 'user' | 'osa', text: string) {
		if (!text.trim()) return;
		const id = ++captionIdCounter;
		captions = [...captions.slice(-4), { id, sender, text: text.trim() }];
		// Keep the voice transcript readable after the spoken answer.
		const timer = setTimeout(() => {
			captions = captions.filter(c => c.id !== id);
			captionTimers.delete(timer);
		}, 60000);
		captionTimers.add(timer);
	}

	// Track external OSA messages
	$effect(() => {
		if (externalOsaMessage) {
			addCaption('osa', externalOsaMessage);
		}
	});

	function handleToggle() {
		if (onToggleListening) onToggleListening();
		else void voice.toggle();
	}

	let video: HTMLVideoElement | null = $state(null);

	// If the orb video fails to load (e.g. asset missing or protocol can't
	// serve it), degrade gracefully to a static decorative fallback instead
	// of surfacing a broken <video> element. The orb is purely decorative.
	let videoFailed = $state(false);

	function handleVideoError() {
		videoFailed = true;
	}

	// Drag state
	let isDragging = $state(false);
	let dragOffset = $state({ x: 0, y: 0 });
	let hasMoved = $state(false);
	let startPos = $state({ x: 0, y: 0 });
	let position = $state({ x: 0, y: 0 });
	let useCustomPosition = $state(false);

	function clampPosition(value: {x:number;y:number}) {
		return {
			x: Math.max(8, Math.min(value.x, Math.max(8, window.innerWidth - 96))),
			y: Math.max(8, Math.min(value.y, Math.max(8, window.innerHeight - 120)))
		};
	}
	function keepInBounds() { if (useCustomPosition) position = clampPosition(position); }
	onMount(() => {
		try {
			const saved = JSON.parse(localStorage.getItem('osaOrbPosition') || 'null');
			if (saved && Number.isFinite(saved.x) && Number.isFinite(saved.y)) {
				position = clampPosition(saved);
				useCustomPosition = true;
			}
		} catch { /* Invalid saved positions fall back to the default corner. */ }
	});
	onDestroy(() => { voice.destroy(); captionTimers.forEach(clearTimeout); });
	function handleCanPlay() { video?.play().catch(() => {}); }
	let suppressClick = false;
	let dragPointer: number | null = null;

	// Drag handlers
	function savePosition() {
		if (browser && useCustomPosition) {
			localStorage.setItem('osaOrbPosition', JSON.stringify(position));
		}
	}

	function handleDragStart(e: PointerEvent) {
		if (e.button !== 0 || !e.isPrimary) return;
        e.preventDefault();
		const target = e.currentTarget as HTMLButtonElement;
		const rect = target.parentElement!.getBoundingClientRect();
		isDragging = true;
		hasMoved = false;
		dragPointer = e.pointerId;
		startPos = {x:e.clientX,y:e.clientY};
		dragOffset = {x:e.clientX-rect.left,y:e.clientY-rect.top};
		target.setPointerCapture(e.pointerId);
	}
	function handleDragMove(e: PointerEvent) {
		if (!isDragging || dragPointer !== e.pointerId) return;
		if (Math.hypot(e.clientX-startPos.x,e.clientY-startPos.y) > 5) hasMoved = true;
		if (!hasMoved) return;
		position = clampPosition({x:e.clientX-dragOffset.x,y:e.clientY-dragOffset.y});
		useCustomPosition = true;
	}
	function handleDragEnd(e: PointerEvent) {
		if (dragPointer !== e.pointerId) return;
		suppressClick = hasMoved;
		if (hasMoved) savePosition();
		isDragging = false;
		dragPointer = null;
		const target = e.currentTarget as HTMLButtonElement;
		if (target.hasPointerCapture(e.pointerId)) target.releasePointerCapture(e.pointerId);
	}
	function cancelDrag() { isDragging = false; dragPointer = null; suppressClick = true; }
	function handleClick(e: MouseEvent) {
		if (suppressClick && e.detail !== 0) { suppressClick = false; return; }
		suppressClick = false;
		handleToggle();
	}
	function resetPosition() {
		useCustomPosition = false;
		position = { x: 0, y: 0 };
		if (browser) localStorage.removeItem('osaOrbPosition');
	}
</script>

<svelte:window onresize={keepInBounds} onblur={cancelDrag} />

<div
	class="osa-orb"
	class:dragging={isDragging}
	class:custom-position={useCustomPosition}
	class:captions-below={useCustomPosition && position.y < 220}
	class:captions-left={useCustomPosition && position.x < 300}
	style={useCustomPosition ? `left: ${position.x}px; top: ${position.y}px;` : ''}
>
	{#if useCustomPosition}
		<button class="reset-position" onclick={resetPosition} title="Reset position">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
				<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
				<path d="M3 3v5h5" />
			</svg>
		</button>
	{/if}

	<button
		class="orb-button"
		class:listening={isListening}
		class:speaking={isSpeaking}
		title={isListening ? 'Stop recording and send to OSA' : phase === 'idle' || phase === 'error' ? 'Tap to speak' : 'Cancel voice request'}
		aria-label={isListening ? 'Stop recording and send to OSA' : phase === 'idle' || phase === 'error' ? 'Talk to OSA' : 'Cancel voice request'}
		aria-pressed={isListening}
		onpointerdown={handleDragStart}
		onpointermove={handleDragMove}
		onpointerup={handleDragEnd}
		onpointercancel={cancelDrag}
		onlostpointercapture={() => { if (isDragging) cancelDrag(); }}
		onclick={handleClick}
	>
		<div class="orb-video-wrap">
			{#if videoFailed}
				<!-- Static decorative fallback when the video can't load -->
				<div class="orb-fallback" aria-hidden="true"></div>
			{:else}
				<!-- svelte-ignore a11y_media_has_caption -->
				<video
					bind:this={video}
					src="/OSAFinalNOBG.mp4"
					class="orb-video"
					preload="auto"
					autoplay
					muted
					playsinline
					loop
					oncanplay={handleCanPlay}
					onerror={handleVideoError}
				></video>
			{/if}
		</div>
	</button>

	<!-- Live captions — conversation bubbles near orb -->
	{#if captions.length > 0 || currentTranscript || voiceError || isListening || phase === 'transcribing' || phase === 'thinking'}
		<div class="captions-panel" aria-live="polite">
            {#if isListening && !currentTranscript}<div class="caption-bubble caption-user"><span class="caption-label">Voice conversation</span><span class="caption-text">Listening. Tap the orb when you’re done.</span></div>{/if}
            {#if phase === 'transcribing'}<div class="caption-bubble caption-user"><span class="caption-text">Finishing transcript…</span></div>{/if}
			{#if voiceError}<div class="caption-bubble caption-error" role="alert">{voiceError}</div>{/if}
			{#each captions as msg (msg.id)}
				<div
					class="caption-bubble"
					class:caption-user={msg.sender === 'user'}
					class:caption-osa={msg.sender === 'osa'}
					transition:fly={{ y: 16, duration: 250 }}
				>
					<span class="caption-label">{msg.sender === 'user' ? 'You' : 'OSA'}</span>
					<span class="caption-text">{msg.text}</span>
				</div>
			{/each}
			{#if currentTranscript}
				<div class="caption-bubble caption-user caption-interim" transition:fade={{ duration: 150 }}>
					<span class="caption-label">You</span>
					<span class="caption-text">{currentTranscript}<span class="typing-dot">...</span></span>
				</div>
			{/if}
		</div>
	{/if}

	<div class="status-label" aria-live="polite">
		{#if isListening}
			<span class="status-listening">Listening...</span>
		{:else if isSpeaking}
			<span class="status-speaking">Speaking</span>
		{:else if phase === 'starting'}<span>Microphone...</span>
		{:else if phase === 'transcribing'}<span>Transcribing...</span>
		{:else if phase === 'thinking'}<span>OSA thinking...</span>
		{:else if phase === 'error'}<span>Try again</span>
		{:else}
			<span class="status-idle">OSA</span>
		{/if}
	</div>
</div>

<style>
	.osa-orb {
		position: fixed;
		width: 88px;
		height: 112px;
		bottom: 100px;
		right: 30px;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		cursor: grab;
		pointer-events: auto;
	}

	.osa-orb.custom-position {
		bottom: auto;
		right: auto;
	}

	.osa-orb.dragging {
		cursor: grabbing;
		user-select: none;
	}

	/* Reset button */
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

	.osa-orb:hover .reset-position {
		opacity: 0.8;
	}

	.reset-position:hover {
		opacity: 1 !important;
		color: white;
		background: rgba(239, 68, 68, 0.8);
		transform: scale(1.1);
	}

	/* Orb button — explicit resets to prevent global style leaks */
	.orb-button {
		position: relative;
		width: 72px;
		height: 72px;
		padding: 0;
		border: none;
		background: transparent;
		backdrop-filter: none;
		-webkit-backdrop-filter: none;
		border-radius: 0;
		cursor: grab;
		transition: transform 0.2s ease;
		touch-action: none;
		outline-offset: 4px;
	}

	.osa-orb.dragging .orb-button {
		cursor: grabbing;

	}

	.orb-button:hover {
		transform: scale(1.06);
	}

	.orb-button:active {
		transform: scale(0.94);
	}

	/* Video wrapper — circle clip hides the black MP4 background */
	.orb-video-wrap {
        pointer-events: none;
        user-select: none;
		position: relative;
		width: 100%;
		height: 100%;
		border-radius: 50%;
		overflow: hidden;
	}

	.orb-video {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transform: scale(1.63);
		filter: drop-shadow(0 8px 20px rgba(0, 0, 0, 0.45))
		        drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3))
		        drop-shadow(0 16px 40px rgba(0, 0, 0, 0.2));
	}

	/* Static decorative fallback shown if the video fails to load */
	.orb-fallback {
		width: 100%;
		height: 100%;
		border-radius: 50%;
		background:
			radial-gradient(circle at 35% 30%, rgba(96, 165, 250, 0.95), rgba(59, 130, 246, 0.85) 45%, rgba(37, 99, 235, 0.9) 100%);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.45),
		            0 0 18px rgba(59, 130, 246, 0.45);
	}

	/* === LISTENING STATE: blue glow + breathing on wrap (not video, to preserve scale crop) === */
	.orb-button.listening .orb-video {
		filter: drop-shadow(0 0 14px rgba(59, 130, 246, 0.55))
		        drop-shadow(0 0 28px rgba(59, 130, 246, 0.3))
		        drop-shadow(0 10px 24px rgba(0, 0, 0, 0.35));
	}

	.orb-button.listening .orb-video-wrap {
		animation: listen-breathe 3s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	/* === SPEAKING STATE: brighter blue glow + organic float on wrap === */
	.orb-button.speaking .orb-video {
		filter: drop-shadow(0 0 18px rgba(59, 130, 246, 0.65))
		        drop-shadow(0 0 36px rgba(59, 130, 246, 0.35))
		        drop-shadow(0 12px 28px rgba(0, 0, 0, 0.3))
		        brightness(1.05);
	}

	.orb-button.speaking .orb-video-wrap {
		animation: listen-breathe 3s ease-in-out infinite;
	}

	@keyframes listen-breathe {
		0%, 100% { transform: scale(1); }
		50% { transform: scale(1.04); }
	}

	@keyframes speak-float {
		0%   { transform: scale(1) translateY(0); }
		25%  { transform: scale(1.03) translateY(-2px); }
		50%  { transform: scale(1.02) translateY(-3px); }
		75%  { transform: scale(1.04) translateY(-1px); }
		100% { transform: scale(1) translateY(0); }
	}

	/* Status label — frosted glass pill with readable dark text */
	.status-label {
		padding: 4px 14px;
		background: rgba(255, 255, 255, 0.55);
		backdrop-filter: blur(20px) saturate(1.4);
		-webkit-backdrop-filter: blur(20px) saturate(1.4);
		border-radius: 20px;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.02em;
		box-shadow: 0 1px 8px rgba(0, 0, 0, 0.1), inset 0 0.5px 0 rgba(255, 255, 255, 0.5);
		white-space: nowrap;
		border: 1px solid rgba(255, 255, 255, 0.4);
	}

	.status-listening {
		color: #1d4ed8;
		animation: text-pulse 2s ease-in-out infinite;
	}

	.status-speaking {
		color: #1d4ed8;
		animation: text-pulse 1.5s ease-in-out infinite;
	}

	.status-idle {
		color: #0f172a;
	}

	@keyframes text-pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.7; }
	}

	/* Dark mode — flip to light text on dark glass */
	:global(.dark) .status-label {
		background: rgba(255, 255, 255, 0.08);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3), inset 0 0.5px 0 rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .status-idle {
		color: rgba(255, 255, 255, 0.7);
	}

	:global(.dark) .status-listening,
	:global(.dark) .status-speaking {
		color: #60a5fa;
	}

	/* ===== CAPTIONS PANEL ===== */
	.captions-panel {
		position: absolute;
		bottom: calc(100% + 12px);
		right: 0;
		width: min(280px, calc(100vw - 24px));
		max-height: min(300px, 45vh);
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 6px;
		max-width: 280px;
		pointer-events: none;
	}

	.osa-orb.captions-below .captions-panel { bottom: auto; top: calc(100% + 12px); }
	.osa-orb.captions-left .captions-panel { right: auto; left: 0; }
	.caption-error { background: #7f1d1d; color: white; font-size: 13px; }
	.orb-button:focus-visible { outline: 2px solid #60a5fa; }
	@media (prefers-reduced-motion: reduce) { .orb-video-wrap { animation: none !important; } }
	.caption-bubble {
		padding: 6px 12px;
		border-radius: 12px;
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		max-width: 100%;
		word-wrap: break-word;
		overflow-wrap: break-word;
	}

	.caption-user {
		background: rgba(59, 130, 246, 0.88);
		box-shadow: 0 2px 10px rgba(59, 130, 246, 0.3);
	}

	.caption-osa {
		background: rgba(168, 85, 247, 0.88);
		box-shadow: 0 2px 10px rgba(168, 85, 247, 0.3);
	}

	.caption-interim {
		opacity: 0.75;
	}

	.caption-label {
		display: block;
		color: rgba(255, 255, 255, 0.7);
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin-bottom: 2px;
	}

	.caption-text {
		color: white;
		font-size: 13px;
		line-height: 1.35;
		font-weight: 500;
	}

	.typing-dot {
		animation: blink-dots 1s steps(3, end) infinite;
	}

	@keyframes blink-dots {
		0% { opacity: 0.3; }
		50% { opacity: 1; }
		100% { opacity: 0.3; }
	}

	:global(.dark) .caption-user {
		background: rgba(59, 130, 246, 0.8);
	}

	:global(.dark) .caption-osa {
		background: rgba(168, 85, 247, 0.8);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.osa-orb {
            width: 58px;
            height: 58px;
			bottom: max(13px, env(safe-area-inset-bottom));
			right: 12px;
			gap: 0;
		}

		.orb-button {
			width: 42px;
			height: 42px;
		}

		.status-label,
		.reset-position { display: none; }
	}

	@media (max-height: 500px) and (pointer: coarse) {
		.osa-orb { display: none; }
	}
</style>
