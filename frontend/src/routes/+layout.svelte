<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/themeStore';
	import { useSession } from '$lib/auth-client';
	import { streamingVoice, type VoiceState } from '$lib/services/livekitVoice';
	import VoiceOrbPanel from '$lib/components/desktop3d/VoiceOrbPanel.svelte';
	import LiveCaptions from '$lib/components/desktop3d/LiveCaptions.svelte';
	import { isOnboardingComplete } from '$lib/stores/onboardingStore';

	let { children } = $props();

	const session = useSession();

	// Derived: Show voice UI if user completed onboarding OR is on main app pages
	let showVoiceUI = $derived(
		$session.data && (
			$isOnboardingComplete ||
			$page.url.pathname === '/window' ||
			$page.url.pathname.startsWith('/(app)')
		)
	);

	// Debug: Track showVoiceUI changes
	$effect(() => {
		console.log('[Root Layout] showVoiceUI changed to:', showVoiceUI, {
			hasSession: !!$session.data,
			isOnboardingComplete: $isOnboardingComplete,
			pathname: $page.url.pathname
		});
	});

	// Voice state (only for authenticated users)
	let voiceState = $state<VoiceState>('disconnected');
	let isListening = $state(false);
	let isSpeaking = $state(false);
	let userMessage = $state('');
	let osaMessage = $state('');

	// Track if voice system has been initialized
	let voiceInitialized = false;

	// Initialize theme on mount
	onMount(() => {
		// Theme is already applied by the store on creation,
		// but we ensure it's set on the document
		if (browser) {
			const storedTheme = localStorage.getItem('theme');
			if (storedTheme === 'dark' || storedTheme === 'light' || storedTheme === 'system') {
				themeStore.setTheme(storedTheme);
			}
		}

		// Setup voice callbacks only once
		if (!voiceInitialized) {
			streamingVoice.setStateCallback((state: VoiceState) => {
				voiceState = state;
				isListening = state === 'connected' || state === 'transcribing' || state === 'speaking';
				isSpeaking = state === 'speaking';
			});

			streamingVoice.setUserCallback((text: string) => {
				userMessage = text;
				setTimeout(() => {
					if (userMessage === text) userMessage = '';
				}, 5000);
			});

			streamingVoice.setAgentCallback((text: string) => {
				osaMessage = text;
				setTimeout(() => {
					if (osaMessage === text) osaMessage = '';
				}, 8000);
			});

			streamingVoice.setErrorCallback((error: string) => {
				console.error('[Root Layout] Streaming voice error:', error);
				// Show error to user (optional)
			});

			voiceInitialized = true;
			console.log('[Root Layout] Streaming voice system initialized');
		}
	});

	// Cleanup - ONLY disconnect if user manually navigates away
	onDestroy(() => {
		console.log('[Root Layout] onDestroy called, voiceState:', voiceState);
		// Don't auto-disconnect - let user manually toggle off
		// Otherwise page navigations kill active conversations
	});

	// Toggle voice
	let isToggling = false;
	async function toggleVoice() {
		if (isToggling) {
			console.log('[Root Layout] toggleVoice called but already toggling, ignoring');
			return;
		}

		isToggling = true;
		console.log('[Root Layout] toggleVoice called, current state:', voiceState);

		try {
			if (voiceState === 'disconnected') {
				console.log('[Root Layout] Connecting voice...');
				await streamingVoice.connect();
			} else {
				console.log('[Root Layout] Disconnecting voice...');
				await streamingVoice.disconnect();
			}
		} finally {
			isToggling = false;
		}
	}
</script>

<svelte:head>
	<title>Business OS</title>
	<meta name="description" content="Your internal command center" />
</svelte:head>

<!-- Page content -->
{@render children()}

<!-- Voice Orb (for authenticated users on main app pages) -->
{#if showVoiceUI}
	<VoiceOrbPanel {isListening} {isSpeaking} onToggleListening={toggleVoice} />
	<LiveCaptions {userMessage} {osaMessage} {isListening} {isSpeaking} />
{/if}
