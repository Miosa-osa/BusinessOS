<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession, appMode, setAppMode, initiateGoogleOAuth, cloudServerUrl } from '$lib/auth-client';
	import { getBackendUrl } from '$lib/api/base';
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { browser } from '$app/environment';
	import LandingHero from './landing/LandingHero.svelte';
	import LandingFeatures from './landing/LandingFeatures.svelte';
	import LandingFooter from './landing/LandingFooter.svelte';

	// Check if running in Electron
	const isElectron = browser && typeof window !== 'undefined' && 'electron' in window;

	// Mode selection state for Electron
	let showModeSelector = $state(false);
	let cloudUrl = $state('');

	const session = useSession();

	$effect(() => {
		if (isElectron && $appMode === null) {
			showModeSelector = true;
			return;
		}
		if (isElectron && $appMode === 'local') {
			goto('/dashboard');
			return;
		}
		if (!$session.isPending && $session.data) {
			goto('/window');
		}
	});

	function selectLocalMode() {
		setAppMode('local');
	}

	function selectCloudMode() {
		if (cloudUrl.trim()) {
			setAppMode('cloud', cloudUrl.trim());
		}
	}

	const defaultCloudUrl = browser ? getBackendUrl() : 'https://api.businessos.app';

	function signInWithGoogle() {
		localStorage.setItem('businessos_mode', 'cloud');
		localStorage.setItem('businessos_cloud_url', defaultCloudUrl);
		initiateGoogleOAuth(defaultCloudUrl);
	}

	function showEmailSignIn() {
		localStorage.setItem('businessos_mode', 'cloud');
		localStorage.setItem('businessos_cloud_url', defaultCloudUrl);
		goto('/login');
	}

	let scrolled = $state(false);
	let showContent = $state(false);

	// Typewriter state
	let typedTagline = $state('');
	let showTaglineCursor = $state(true);
	let terminalLine1 = $state('');
	let terminalLine2 = $state('');
	let showTerminalOutput = $state(false);
	let showModuleList = $state(false);

	const tagline = 'Self-hosted. AI-native. Built for fast software.';
	const cmd1 = '$ businessos status';
	const cmd2 = '$ businessos modules --list';

	function typeWriter(text: string, setter: (val: string) => void, speed: number = 50): Promise<void> {
		return new Promise((resolve) => {
			let i = 0;
			const interval = setInterval(() => {
				if (i < text.length) {
					setter(text.slice(0, i + 1));
					i++;
				} else {
					clearInterval(interval);
					resolve();
				}
			}, speed);
		});
	}

	onMount(() => {
		setTimeout(() => (showContent = true), 100);

		setTimeout(async () => {
			await typeWriter(tagline, (val) => typedTagline = val, 30);
			showTaglineCursor = false;
		}, 800);

		setTimeout(async () => {
			await typeWriter(cmd1, (val) => terminalLine1 = val, 40);
			setTimeout(() => showTerminalOutput = true, 200);

			setTimeout(async () => {
				await typeWriter(cmd2, (val) => terminalLine2 = val, 40);
				setTimeout(() => showModuleList = true, 300);
			}, 800);
		}, 1500);

		const handleScroll = () => {
			scrolled = window.scrollY > 20;
		};
		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});

	const modules = [
		{ name: 'Desktop', desc: 'Native app with voice & shortcuts', slug: 'desktop' },
		{ name: 'Dashboard', desc: 'Command center', slug: 'dashboard' },
		{ name: 'Chat', desc: 'AI conversations', slug: 'chat' },
		{ name: 'Tasks', desc: 'Track work', slug: 'tasks' },
		{ name: 'Projects', desc: 'Organize teams', slug: 'projects' },
		{ name: 'Calendar', desc: 'Schedule', slug: 'calendar' },
		{ name: 'Clients', desc: 'Relationships', slug: 'clients' },
		{ name: 'Contexts', desc: 'Knowledge base', slug: 'contexts' },
		{ name: 'Nodes', desc: 'Connections', slug: 'nodes' },
		{ name: 'Daily Log', desc: 'Journal', slug: 'daily-log' },
	];

	const capabilities = [
		{ title: 'Self-Hosted', desc: 'Your servers. Your data. Full control.' },
		{ title: 'AI Native', desc: 'Built-in agents. Local or cloud LLMs.' },
		{ title: 'Open Source', desc: 'Modify anything. Extend everything.' },
		{ title: 'Enterprise Ready', desc: 'Built for scale. Team collaboration.' },
	];

	const integrations = [
		{ name: 'Salesforce', category: 'CRM' },
		{ name: 'HubSpot', category: 'CRM' },
		{ name: 'GoHighLevel', category: 'CRM' },
		{ name: 'Airtable', category: 'Data' },
		{ name: 'Notion', category: 'Docs' },
		{ name: 'Slack', category: 'Comms' },
		{ name: 'Google Workspace', category: 'Suite' },
		{ name: 'Microsoft 365', category: 'Suite' },
		{ name: 'PostgreSQL', category: 'DB' },
		{ name: 'MongoDB', category: 'DB' },
		{ name: 'REST APIs', category: 'Custom' },
		{ name: 'Legacy Systems', category: 'Custom' },
	];

	const howItWorks = [
		{ step: '01', title: 'Contexts', desc: 'Create knowledge bases for each client, project, or domain.', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' },
		{ step: '02', title: 'Projects', desc: 'Organize work into projects with tasks, milestones, and team assignments.', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' },
		{ step: '03', title: 'Agents', desc: 'AI agents understand your contexts and execute tasks.', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ step: '04', title: 'Automate', desc: 'Connect your existing tools and let agents handle repetitive work.', icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' },
	];
</script>

<!-- Mode selector for Electron first launch -->
{#if showModeSelector}
	<div class="min-h-screen flex bg-white">
		<!-- Left Panel - Terminal Style Branding -->
		<div class="hidden lg:flex lg:w-1/2 bg-black p-12 flex-col justify-between relative overflow-hidden">
			<div class="absolute inset-0">
				<div class="w-full h-full opacity-[0.12]" style="background-image: linear-gradient(rgba(255,255,255,0.25) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.25) 1px, transparent 1px); background-size: 50px 50px;"></div>
				<div class="absolute inset-0 w-full h-full opacity-[0.06]" style="background-image: linear-gradient(rgba(255,255,255,0.2) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.2) 1px, transparent 1px); background-size: 10px 10px;"></div>
				<div class="absolute top-6 left-6 w-24 h-24 border-l-2 border-t-2 border-white/20"></div>
				<div class="absolute top-6 right-6 w-24 h-24 border-r-2 border-t-2 border-white/20"></div>
				<div class="absolute bottom-6 left-6 w-24 h-24 border-l-2 border-b-2 border-white/20"></div>
				<div class="absolute bottom-6 right-6 w-24 h-24 border-r-2 border-b-2 border-white/20"></div>
			</div>
			<div class="relative z-10" in:fly={{ y: -20, duration: 500 }}>
				<div class="flex items-baseline gap-0.5 mb-6">
					<span class="text-white text-3xl font-extrabold tracking-[0.2em] font-mono">BUSINESS</span>
					<span class="text-white/40 text-2xl font-light font-mono">OS</span>
				</div>
				<div class="font-mono text-gray-500 text-sm space-y-1">
					<p>> Desktop application ready</p>
					<p>> Select operation mode_<span class="inline-block w-2 h-4 bg-green-500 ml-0.5 animate-pulse"></span></p>
				</div>
			</div>
			<div class="relative z-10 flex-1 flex items-center justify-center py-12" in:fly={{ y: 30, duration: 600 }}>
				<div class="w-full max-w-sm">
					<div class="bg-gray-900/90 border border-gray-700 rounded-lg overflow-hidden shadow-2xl shadow-black/50">
						<div class="flex items-center gap-2 px-4 py-2.5 bg-gray-900 border-b border-gray-800">
							<div class="flex gap-1.5">
								<div class="w-3 h-3 rounded-full bg-red-500"></div>
								<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
								<div class="w-3 h-3 rounded-full bg-green-500"></div>
							</div>
							<span class="ml-2 text-xs text-gray-500 font-mono">businessos ~ setup</span>
						</div>
						<div class="p-4 font-mono text-sm space-y-2">
							<div class="text-gray-500">$ businessos --mode</div>
							<div class="text-gray-400 pl-2 space-y-1">
								<div class="flex gap-4 items-center"><span class="text-gray-600 w-4">01</span><span class="flex-1">local</span><span class="text-green-500/60 text-xs">[offline]</span></div>
								<div class="flex gap-4 items-center"><span class="text-gray-600 w-4">02</span><span class="flex-1">cloud</span><span class="text-blue-500/60 text-xs">[sync]</span></div>
							</div>
							<div class="text-gray-500 pt-2 flex items-center"><span>$ </span><span class="inline-block w-2 h-4 bg-green-500 ml-0.5 animate-pulse"></span></div>
						</div>
					</div>
				</div>
			</div>
			<div class="relative z-10 font-mono text-xs text-gray-600">
				<div class="flex items-center gap-4">
					<span class="text-gray-500">v1.0.1</span>
					<span class="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse"></span>
					<span class="text-gray-500">Desktop Edition</span>
				</div>
			</div>
		</div>

		<!-- Right Panel - Mode Selection Form -->
		<div class="flex-1 flex items-center justify-center p-6 sm:p-12 bg-white relative overflow-hidden">
			<div class="absolute inset-0 pointer-events-none">
				<div class="w-full h-full opacity-[0.06]" style="background-image: linear-gradient(rgba(0,0,0,0.2) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.2) 1px, transparent 1px); background-size: 50px 50px;"></div>
				<div class="absolute top-4 left-4 w-16 h-16 border-l border-t border-gray-200"></div>
				<div class="absolute top-4 right-4 w-16 h-16 border-r border-t border-gray-200"></div>
				<div class="absolute bottom-4 left-4 w-16 h-16 border-l border-b border-gray-200"></div>
				<div class="absolute bottom-4 right-4 w-16 h-16 border-r border-b border-gray-200"></div>
			</div>
			<div class="w-full max-w-md relative z-10" in:fly={{ y: 20, duration: 400 }}>
				<div class="lg:hidden flex items-baseline gap-0.5 mb-10 justify-center">
					<span class="text-black text-2xl font-extrabold tracking-[0.15em] font-mono">BUSINESS</span>
					<span class="text-black/30 text-xl font-light font-mono">OS</span>
				</div>
				<div class="mb-8">
					<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Select Mode</h1>
					<p class="text-gray-500 text-sm font-mono">Choose how you want to use BusinessOS</p>
				</div>
				<div class="space-y-4">
					<button onclick={selectLocalMode} class="btn-pill btn-pill-outline w-full text-left p-5 border-2 group">
						<div class="flex items-start gap-4">
							<div class="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center group-hover:bg-black group-hover:text-white transition-colors">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
							</div>
							<div class="flex-1">
								<div class="flex items-center justify-between">
									<h3 class="font-semibold text-gray-900 font-mono">Local Mode</h3>
									<svg class="w-5 h-5 text-gray-400 group-hover:text-black group-hover:translate-x-1 transition-all" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" /></svg>
								</div>
								<p class="text-sm text-gray-500 mt-1 font-mono">Offline-first, data stored locally</p>
								<div class="flex flex-wrap gap-2 mt-3">
									<span class="inline-flex items-center gap-1 text-xs text-green-600 bg-green-50 px-2 py-1 rounded font-mono"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>No account needed</span>
									<span class="inline-flex items-center gap-1 text-xs text-green-600 bg-green-50 px-2 py-1 rounded font-mono"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>Works offline</span>
								</div>
							</div>
						</div>
					</button>

					<div class="flex items-center gap-4">
						<div class="flex-1 h-px bg-gray-100"></div>
						<span class="text-xs text-gray-400 font-mono uppercase tracking-wider">or</span>
						<div class="flex-1 h-px bg-gray-100"></div>
					</div>

					<div class="p-5 border-2 border-gray-200 rounded-xl">
						<div class="flex items-start gap-4 mb-4">
							<div class="w-12 h-12 bg-blue-50 rounded-lg flex items-center justify-center text-blue-600">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" /></svg>
							</div>
							<div class="flex-1">
								<h3 class="font-semibold text-gray-900 font-mono">Cloud Mode</h3>
								<p class="text-sm text-gray-500 mt-1 font-mono">Sync across devices, team collaboration</p>
								<div class="flex flex-wrap gap-2 mt-3">
									<span class="inline-flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded font-mono"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>Sync data</span>
									<span class="inline-flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded font-mono"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>Team features</span>
								</div>
							</div>
						</div>
						<div class="space-y-3">
							<button onclick={signInWithGoogle} class="btn-pill btn-pill-secondary btn-pill-sm w-full h-12 flex items-center justify-center gap-3">
								<svg class="w-5 h-5" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/></svg>
								Sign in with Google
							</button>
							<button onclick={showEmailSignIn} class="btn-pill btn-pill-primary btn-pill-sm w-full h-12 flex items-center justify-center gap-2 font-mono">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
								Sign in with Email
							</button>
							<div class="flex items-center gap-3"><div class="flex-1 h-px bg-gray-100"></div><span class="text-xs text-gray-400 font-mono">or self-host</span><div class="flex-1 h-px bg-gray-100"></div></div>
							<input id="cloudUrl" type="url" bind:value={cloudUrl} placeholder="https://your-server.com" class="w-full border border-gray-200 rounded-lg px-4 py-2.5 text-sm font-mono placeholder-gray-400 focus:border-black focus:outline-none focus:ring-1 focus:ring-black transition-colors" />
							<button onclick={selectCloudMode} disabled={!cloudUrl.trim()} class="btn-pill btn-pill-soft btn-pill-sm w-full h-10 disabled:bg-gray-50 disabled:text-gray-300 flex items-center justify-center gap-2 font-mono">
								Connect to Self-Hosted Server
							</button>
						</div>
					</div>
				</div>
				<p class="mt-8 text-center text-sm text-gray-400 font-mono">You can change this later in Settings</p>
			</div>
		</div>
	</div>
{:else}
<div class="min-h-screen bg-white relative">
	<!-- Grid overlay background -->
	{#if showContent}
		<div class="fixed inset-0 pointer-events-none" in:fade={{ duration: 1000 }}>
			<div class="w-full h-full opacity-[0.08]" style="background-image: linear-gradient(rgba(0,0,0,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.4) 1px, transparent 1px); background-size: 60px 60px;"></div>
			<div class="absolute inset-0 w-full h-full opacity-[0.04]" style="background-image: linear-gradient(rgba(0,0,0,0.3) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.3) 1px, transparent 1px); background-size: 20px 20px;"></div>
			<div class="absolute top-0 left-0 w-40 h-40 border-l-2 border-t-2 border-gray-300/50"></div>
			<div class="absolute top-0 right-0 w-40 h-40 border-r-2 border-t-2 border-gray-300/50"></div>
			<div class="absolute bottom-0 left-0 w-40 h-40 border-l-2 border-b-2 border-gray-300/50"></div>
			<div class="absolute bottom-0 right-0 w-40 h-40 border-r-2 border-b-2 border-gray-300/50"></div>
		</div>
	{/if}

	<!-- Header -->
	<header class="fixed top-0 left-0 right-0 z-50 transition-all duration-300 {scrolled ? 'bg-white/95 backdrop-blur-md border-b border-gray-100' : 'bg-transparent'}">
		<div class="max-w-5xl mx-auto px-6 h-14 flex items-center justify-between">
			<div class="flex items-baseline gap-0.5">
				<span class="text-black text-lg font-extrabold tracking-[0.15em] font-mono">BUSINESS</span>
				<span class="text-black/30 text-base font-light font-mono">OS</span>
			</div>
			<div class="flex items-center gap-6">
				<a href="/docs" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">Docs</a>
				<a href="#download" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">Download</a>
				<a href="https://github.com" target="_blank" rel="noopener" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">GitHub</a>
				<a href="/register" class="bg-black text-white px-4 py-2 rounded-lg text-xs font-medium hover:bg-gray-800 transition-colors font-mono">Get Started</a>
			</div>
		</div>
	</header>

	{#if showContent}
		<LandingHero
			{typedTagline}
			{showTaglineCursor}
			{terminalLine1}
			{terminalLine2}
			{showTerminalOutput}
			{showModuleList}
			{modules}
			{cmd1}
			{cmd2}
		/>
		<LandingFeatures {modules} {capabilities} {howItWorks} {integrations} />
		<LandingFooter />
	{/if}
</div>
{/if}

<style>
	@keyframes blink {
		0%, 50% { opacity: 1; }
		51%, 100% { opacity: 0; }
	}

	:global(.animate-blink) {
		animation: blink 1s step-end infinite;
	}

	/* Scan line moving animation */
	:global(.scan-line-moving) {
		animation: scanLineMove 4s linear infinite;
	}

	@keyframes scanLineMove {
		0% { transform: translateY(-10px); }
		100% { transform: translateY(100vh); }
	}

	/* Floating particles */
	:global(.animate-float-slow) { animation: floatSlow 8s ease-in-out infinite; }
	:global(.animate-float-medium) { animation: floatMedium 6s ease-in-out infinite; }
	:global(.animate-float-fast) { animation: floatFast 4s ease-in-out infinite; }

	@keyframes floatSlow {
		0%, 100% { transform: translate(0, 0); }
		25% { transform: translate(10px, -15px); }
		50% { transform: translate(-5px, -25px); }
		75% { transform: translate(-15px, -10px); }
	}

	@keyframes floatMedium {
		0%, 100% { transform: translate(0, 0); }
		33% { transform: translate(-12px, -20px); }
		66% { transform: translate(8px, -10px); }
	}

	@keyframes floatFast {
		0%, 100% { transform: translate(0, 0); }
		50% { transform: translate(15px, -20px); }
	}
</style>
