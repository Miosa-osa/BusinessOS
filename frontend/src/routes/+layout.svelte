<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/themeStore';
	import { useSession } from '$lib/auth-client';
	import { simpleVoice, type VoiceState } from '$lib/services/simpleVoice';
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

	// Voice state (only for authenticated users)
	let voiceState = $state<VoiceState>('disconnected');
	let isListening = $state(false);
	let isSpeaking = $state(false);
	let userMessage = $state('');
	let osaMessage = $state('');

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

		// Setup voice callbacks (for authenticated users)
		simpleVoice.setStateCallback((state: VoiceState) => {
			voiceState = state;
			isListening = state === 'connected' || state === 'speaking';
			isSpeaking = state === 'speaking';
		});

		simpleVoice.setUserCallback((text: string) => {
			userMessage = text;
			setTimeout(() => {
				if (userMessage === text) userMessage = '';
			}, 5000);
		});

		simpleVoice.setAgentCallback((text: string) => {
			osaMessage = text;
			setTimeout(() => {
				if (osaMessage === text) osaMessage = '';
			}, 8000);
		});

		console.log('[Root Layout] Voice system initialized - shows for authenticated users');
	});

	// Cleanup
	onDestroy(() => {
		if (voiceState !== 'disconnected') {
			simpleVoice.disconnect();
		}
	});

	// Toggle voice
	async function toggleVoice() {
		if (voiceState === 'disconnected') {
			await simpleVoice.connect();
		} else {
			await simpleVoice.disconnect();
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
