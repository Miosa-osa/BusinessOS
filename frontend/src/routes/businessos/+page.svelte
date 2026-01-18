<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession, appMode, setAppMode, initiateGoogleOAuth, cloudServerUrl } from '$lib/auth-client';
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { browser } from '$app/environment';

	// Check if running in Electron
	const isElectron = browser && typeof window !== 'undefined' && 'electron' in window;

	// Mode selection state for Electron
	let showModeSelector = $state(false);
	let cloudUrl = $state('');
	let showEmailForm = $state(false);

	const session = useSession();

	$effect(() => {
		// In Electron, check if mode is set
		if (isElectron && $appMode === null) {
			showModeSelector = true;
			return;
		}

		// If local mode in Electron, go directly to dashboard
		if (isElectron && $appMode === 'local') {
			goto('/dashboard');
			return;
		}

		// Normal auth flow
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

	// Default cloud URL - use local backend in development
	const isDev = browser && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');
	const defaultCloudUrl = isDev ? 'http://localhost:8001' : 'https://api.businessos.app';

	function signInWithGoogle() {
		// Set cloud mode with default URL and initiate OAuth
		localStorage.setItem('businessos_mode', 'cloud');
		localStorage.setItem('businessos_cloud_url', defaultCloudUrl);
		initiateGoogleOAuth(defaultCloudUrl);
	}

	function showEmailSignIn() {
		// Set cloud mode first, then redirect to login page
		localStorage.setItem('businessos_mode', 'cloud');
		localStorage.setItem('businessos_cloud_url', defaultCloudUrl);
		window.location.href = '/login';
	}

	let scrolled = $state(false);
	let showContent = $state(false);

	// Typewriter state
	let typedTagline = $state('');
	let showTaglineCursor = $state(true);
	let terminalLine1 = $state('');
	let terminalLine2 = $state('');
	let terminalLine3 = $state('');
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

		// Start typewriter sequence
		setTimeout(async () => {
			// Type tagline
			await typeWriter(tagline, (val) => typedTagline = val, 30);
			showTaglineCursor = false;
		}, 800);

		// Terminal typing sequence
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
		{
			step: '01',
			title: 'Contexts',
			desc: 'Create knowledge bases for each client, project, or domain. Contexts store documents, notes, and data that inform your AI agents.',
			icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10'
		},
		{
			step: '02',
			title: 'Projects',
			desc: 'Organize work into projects with tasks, milestones, and team assignments. Link projects to contexts for intelligent assistance.',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01'
		},
		{
			step: '03',
			title: 'Agents',
			desc: 'AI agents understand your contexts and execute tasks. They can research, draft, analyze, and integrate with your tools via aMCP.',
			icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
		},
		{
			step: '04',
			title: 'Automate',
			desc: 'Connect your existing tools and let agents handle repetitive work. Build custom workflows that scale with your business.',
			icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15'
		},
	];
</script>

<!-- Mode selector for Electron first launch -->
{#if showModeSelector}
	<div class="min-h-screen flex bg-white">
		<!-- Left Panel - Terminal Style Branding (same as AuthLayout) -->
		<div class="hidden lg:flex lg:w-1/2 bg-black p-12 flex-col justify-between relative overflow-hidden">
			<!-- Grid Pattern Overlay -->
			<div class="absolute inset-0">
				<div
					class="w-full h-full opacity-[0.12]"
					style="
						background-image:
							linear-gradient(rgba(255,255,255,0.25) 1px, transparent 1px),
							linear-gradient(90deg, rgba(255,255,255,0.25) 1px, transparent 1px);
						background-size: 50px 50px;
					"
				></div>
				<div
					class="absolute inset-0 w-full h-full opacity-[0.06]"
					style="
						background-image:
							linear-gradient(rgba(255,255,255,0.2) 1px, transparent 1px),
							linear-gradient(90deg, rgba(255,255,255,0.2) 1px, transparent 1px);
						background-size: 10px 10px;
					"
				></div>
				<div class="absolute top-6 left-6 w-24 h-24 border-l-2 border-t-2 border-white/20"></div>
				<div class="absolute top-6 right-6 w-24 h-24 border-r-2 border-t-2 border-white/20"></div>
				<div class="absolute bottom-6 left-6 w-24 h-24 border-l-2 border-b-2 border-white/20"></div>
				<div class="absolute bottom-6 right-6 w-24 h-24 border-r-2 border-b-2 border-white/20"></div>
				<div class="absolute top-1/4 left-1/4 w-32 h-32 bg-green-500/5 rounded-full blur-3xl animate-pulse"></div>
				<div class="absolute bottom-1/3 right-1/4 w-40 h-40 bg-blue-500/5 rounded-full blur-3xl animate-pulse"></div>
			</div>

			<!-- Scan Line Effect -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				<div class="scan-line-moving w-full h-px bg-gradient-to-r from-transparent via-green-500/30 to-transparent"></div>
			</div>

			<!-- Logo & Tagline -->
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

			<!-- Terminal Preview -->
			<div class="relative z-10 flex-1 flex items-center justify-center py-12" in:fly={{ y: 30, duration: 600 }}>
				<div class="w-full max-w-sm">
					<div class="bg-gray-900/90 border border-gray-700 rounded-lg overflow-hidden shadow-2xl shadow-black/50 backdrop-blur-sm">
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
								<div class="flex gap-4 items-center">
									<span class="text-gray-600 w-4">01</span>
									<span class="flex-1">local</span>
									<span class="text-green-500/60 text-xs">[offline]</span>
								</div>
								<div class="flex gap-4 items-center">
									<span class="text-gray-600 w-4">02</span>
									<span class="flex-1">cloud</span>
									<span class="text-blue-500/60 text-xs">[sync]</span>
								</div>
							</div>
							<div class="text-gray-500 pt-2 flex items-center">
								<span>$ </span>
								<span class="inline-block w-2 h-4 bg-green-500 ml-0.5 animate-pulse"></span>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Version Info -->
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
			<!-- Grid for right panel -->
			<div class="absolute inset-0 pointer-events-none">
				<div
					class="w-full h-full opacity-[0.06]"
					style="
						background-image:
							linear-gradient(rgba(0,0,0,0.2) 1px, transparent 1px),
							linear-gradient(90deg, rgba(0,0,0,0.2) 1px, transparent 1px);
						background-size: 50px 50px;
					"
				></div>
				<div class="absolute top-4 left-4 w-16 h-16 border-l border-t border-gray-200"></div>
				<div class="absolute top-4 right-4 w-16 h-16 border-r border-t border-gray-200"></div>
				<div class="absolute bottom-4 left-4 w-16 h-16 border-l border-b border-gray-200"></div>
				<div class="absolute bottom-4 right-4 w-16 h-16 border-r border-b border-gray-200"></div>
			</div>

			<div class="w-full max-w-md relative z-10" in:fly={{ y: 20, duration: 400 }}>
				<!-- Mobile Logo -->
				<div class="lg:hidden flex items-baseline gap-0.5 mb-10 justify-center">
					<span class="text-black text-2xl font-extrabold tracking-[0.15em] font-mono">BUSINESS</span>
					<span class="text-black/30 text-xl font-light font-mono">OS</span>
				</div>

				<!-- Header -->
				<div class="mb-8">
					<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Select Mode</h1>
					<p class="text-gray-500 text-sm font-mono">Choose how you want to use BusinessOS</p>
				</div>

				<!-- Mode Options -->
				<div class="space-y-4">
					<!-- Local Mode -->
					<button
						onclick={selectLocalMode}
						class="w-full text-left p-5 border-2 border-gray-200 rounded-xl hover:border-black hover:shadow-md transition-all group"
					>
						<div class="flex items-start gap-4">
							<div class="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center group-hover:bg-black group-hover:text-white transition-colors">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
							</div>
							<div class="flex-1">
								<div class="flex items-center justify-between">
									<h3 class="font-semibold text-gray-900 font-mono">Local Mode</h3>
									<svg class="w-5 h-5 text-gray-400 group-hover:text-black group-hover:translate-x-1 transition-all" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
									</svg>
								</div>
								<p class="text-sm text-gray-500 mt-1 font-mono">Offline-first, data stored locally</p>
								<div class="flex flex-wrap gap-2 mt-3">
									<span class="inline-flex items-center gap-1 text-xs text-green-600 bg-green-50 px-2 py-1 rounded font-mono">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										No account needed
									</span>
									<span class="inline-flex items-center gap-1 text-xs text-green-600 bg-green-50 px-2 py-1 rounded font-mono">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										Works offline
									</span>
								</div>
							</div>
						</div>
					</button>

					<!-- Divider -->
					<div class="flex items-center gap-4">
						<div class="flex-1 h-px bg-gray-100"></div>
						<span class="text-xs text-gray-400 font-mono uppercase tracking-wider">or</span>
						<div class="flex-1 h-px bg-gray-100"></div>
					</div>

					<!-- Cloud Mode -->
					<div class="p-5 border-2 border-gray-200 rounded-xl">
						<div class="flex items-start gap-4 mb-4">
							<div class="w-12 h-12 bg-blue-50 rounded-lg flex items-center justify-center text-blue-600">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
								</svg>
							</div>
							<div class="flex-1">
								<h3 class="font-semibold text-gray-900 font-mono">Cloud Mode</h3>
								<p class="text-sm text-gray-500 mt-1 font-mono">Sync across devices, team collaboration</p>
								<div class="flex flex-wrap gap-2 mt-3">
									<span class="inline-flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded font-mono">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										Sync data
									</span>
									<span class="inline-flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded font-mono">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										Team features
									</span>
								</div>
							</div>
						</div>

						<div class="space-y-3">
							<!-- Google Sign In Button -->
							<button
								onclick={signInWithGoogle}
								class="btn-pill btn-pill-outline w-full h-12 flex items-center justify-center gap-3"
							>
								<svg class="w-5 h-5" viewBox="0 0 24 24">
									<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
									<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
									<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
									<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
								</svg>
								Sign in with Google
							</button>

							<!-- Email Sign In Button -->
							<button
								onclick={showEmailSignIn}
								class="btn-pill btn-pill-active font-mono w-full h-12 flex items-center justify-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
								Sign in with Email
							</button>

							<!-- Divider -->
							<div class="flex items-center gap-3">
								<div class="flex-1 h-px bg-gray-100"></div>
								<span class="text-xs text-gray-400 font-mono">or self-host</span>
								<div class="flex-1 h-px bg-gray-100"></div>
							</div>

							<!-- Self-hosted server option -->
							<div>
								<input
									id="cloudUrl"
									type="url"
									bind:value={cloudUrl}
									placeholder="https://your-server.com"
									class="w-full border border-gray-200 rounded-lg px-4 py-2.5 text-sm font-mono placeholder-gray-400 focus:border-black focus:outline-none focus:ring-1 focus:ring-black transition-colors"
								/>
							</div>
							<button
								onclick={selectCloudMode}
								disabled={!cloudUrl.trim()}
								class="btn-pill btn-pill-inactive font-mono w-full h-10 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								Connect to Self-Hosted Server
							</button>
						</div>
					</div>
				</div>

				<p class="mt-8 text-center text-sm text-gray-400 font-mono">
					You can change this later in Settings
				</p>
			</div>
		</div>
	</div>
{:else}
<div class="min-h-screen bg-white relative">
	<!-- Grid overlay background -->
	{#if showContent}
		<div
			class="fixed inset-0 pointer-events-none"
			in:fade={{ duration: 1000 }}
		>
			<!-- Primary grid - more visible -->
			<div
				class="w-full h-full opacity-[0.08]"
				style="
					background-image:
						linear-gradient(rgba(0,0,0,0.4) 1px, transparent 1px),
						linear-gradient(90deg, rgba(0,0,0,0.4) 1px, transparent 1px);
					background-size: 60px 60px;
				"
			></div>
			<!-- Secondary finer grid -->
			<div
				class="absolute inset-0 w-full h-full opacity-[0.04]"
				style="
					background-image:
						linear-gradient(rgba(0,0,0,0.3) 1px, transparent 1px),
						linear-gradient(90deg, rgba(0,0,0,0.3) 1px, transparent 1px);
					background-size: 20px 20px;
				"
			></div>
			<!-- Dot pattern overlay -->
			<div
				class="absolute inset-0 w-full h-full opacity-[0.03]"
				style="
					background-image: radial-gradient(circle, rgba(0,0,0,0.8) 1px, transparent 1px);
					background-size: 30px 30px;
				"
			></div>
			<!-- Corner accents - more visible -->
			<div class="absolute top-0 left-0 w-40 h-40 border-l-2 border-t-2 border-gray-300/50"></div>
			<div class="absolute top-0 right-0 w-40 h-40 border-r-2 border-t-2 border-gray-300/50"></div>
			<div class="absolute bottom-0 left-0 w-40 h-40 border-l-2 border-b-2 border-gray-300/50"></div>
			<div class="absolute bottom-0 right-0 w-40 h-40 border-r-2 border-b-2 border-gray-300/50"></div>
			<!-- Scan line effect -->
			<div class="absolute inset-0 overflow-hidden opacity-[0.03]">
				<div class="scan-line-moving w-full h-1 bg-gradient-to-r from-transparent via-black to-transparent"></div>
			</div>
			<!-- Floating particles -->
			<div class="absolute top-1/4 left-1/4 w-1 h-1 bg-gray-400/30 rounded-full animate-float-slow"></div>
			<div class="absolute top-1/3 right-1/3 w-1.5 h-1.5 bg-gray-400/20 rounded-full animate-float-medium"></div>
			<div class="absolute bottom-1/4 left-1/3 w-1 h-1 bg-gray-400/25 rounded-full animate-float-fast"></div>
			<div class="absolute top-2/3 right-1/4 w-0.5 h-0.5 bg-gray-400/30 rounded-full animate-float-slow"></div>
		</div>
	{/if}

	<!-- Header -->
	<header
		class="fixed top-0 left-0 right-0 z-50 transition-all duration-300
			{scrolled ? 'bg-white/95 backdrop-blur-md border-b border-gray-100' : 'bg-transparent'}"
	>
		<div class="max-w-5xl mx-auto px-6 h-14 flex items-center justify-between">
			<div class="flex items-baseline gap-0.5">
				<span class="text-black text-lg font-extrabold tracking-[0.15em] font-mono">BUSINESS</span>
				<span class="text-black/30 text-base font-light font-mono">OS</span>
			</div>
			<div class="flex items-center gap-6">
				<a href="/docs" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">
					Docs
				</a>
				<a href="#download" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">
					Download
				</a>
				<a href="https://github.com" target="_blank" rel="noopener" class="text-xs text-gray-500 hover:text-black transition-colors font-mono hidden sm:inline">
					GitHub
				</a>
				<a href="/register" class="btn-pill btn-pill-active font-mono">
					Get Started
				</a>
			</div>
		</div>
	</header>

	<!-- Hero -->
	{#if showContent}
		<section class="pt-28 pb-20 px-6">
			<div class="max-w-3xl mx-auto">
				<!-- Terminal-style status -->
				<div
					class="font-mono text-xs text-gray-400 mb-8 flex items-center gap-3"
					in:fly={{ y: 20, duration: 500 }}
				>
					<span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
					<span>SYSTEM ONLINE</span>
					<span class="text-gray-300">|</span>
					<span>v0.0.1</span>
				</div>

				<!-- Main headline -->
				<h1
					class="text-4xl md:text-6xl font-bold text-black leading-[1.1] mb-6 tracking-tight"
					in:fly={{ y: 30, duration: 600, delay: 100 }}
				>
					Your operating system for the{' '}
					<span class="text-gray-400 glitch-text" data-text="agentic era">agentic era</span>
				</h1>

				<p
					class="text-lg text-gray-500 max-w-xl mb-12 font-mono h-7"
					in:fly={{ y: 30, duration: 600, delay: 200 }}
				>
					{typedTagline}<span class="inline-block w-0.5 h-5 bg-gray-400 ml-0.5 align-middle {showTaglineCursor ? 'animate-pulse' : 'opacity-0'}"></span>
				</p>

				<!-- CTA -->
				<div
					class="flex flex-col sm:flex-row items-start gap-4"
					in:fly={{ y: 30, duration: 600, delay: 300 }}
				>
					<a href="/register" class="btn-pill btn-pill-active font-mono flex items-center gap-2 px-8 py-3">
						Initialize workspace
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
						</svg>
					</a>
					<a href="/login" class="btn-pill btn-pill-inactive font-mono px-4 py-3">
						Sign in
					</a>
				</div>
			</div>
		</section>

		<!-- Terminal Preview -->
		<section class="pb-24 px-6" in:fly={{ y: 40, duration: 800, delay: 400 }}>
			<div class="max-w-4xl mx-auto">
				<div class="bg-black rounded-xl overflow-hidden shadow-2xl">
					<!-- Terminal header -->
					<div class="flex items-center gap-2 px-4 py-3 border-b border-gray-800">
						<div class="flex gap-1.5">
							<div class="w-3 h-3 rounded-full bg-red-500"></div>
							<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
							<div class="w-3 h-3 rounded-full bg-green-500"></div>
						</div>
						<span class="ml-4 text-xs text-gray-500 font-mono">businessos ~ terminal</span>
					</div>
					<!-- Terminal content -->
					<div class="p-6 font-mono text-sm space-y-3">
						<div class="text-gray-500 flex items-center">
							<span>{terminalLine1}</span>
							{#if terminalLine1.length > 0 && terminalLine1.length < cmd1.length}
								<span class="inline-block w-2 h-4 bg-gray-500 ml-0.5 animate-pulse"></span>
							{/if}
						</div>
						{#if showTerminalOutput}
							<div class="text-green-400" in:fade={{ duration: 200 }}>
								<span class="inline-block animate-pulse mr-2">></span>
								All systems operational
							</div>
						{/if}
						{#if terminalLine2.length > 0}
							<div class="text-gray-500 flex items-center">
								<span>{terminalLine2}</span>
								{#if terminalLine2.length > 0 && terminalLine2.length < cmd2.length}
									<span class="inline-block w-2 h-4 bg-gray-500 ml-0.5 animate-pulse"></span>
								{/if}
							</div>
						{/if}
						{#if showModuleList}
							<div class="grid grid-cols-2 sm:grid-cols-3 gap-2 pl-2" in:fade={{ duration: 300 }}>
								{#each modules as mod, i}
									<div class="text-gray-400 flex gap-2" in:fly={{ y: 10, duration: 300, delay: i * 80 }}>
										<span class="text-gray-600 w-4">{String(i + 1).padStart(2, '0')}</span>
										<span class="typewriter-text" style="animation-delay: {i * 80}ms">{mod.name.toLowerCase().replace(' ', '-')}</span>
									</div>
								{/each}
							</div>
						{/if}
						<div class="text-gray-500 pt-2 flex items-center">
							<span>$ </span>
							<span class="inline-block w-2 h-4 bg-green-500 ml-0.5 animate-blink"></span>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Download Desktop App -->
		<section id="download" class="py-20 px-6 bg-gray-50">
			<div class="max-w-5xl mx-auto">
				<div class="text-center mb-10">
					<div class="font-mono text-xs text-gray-400 mb-3 tracking-wider">DESKTOP APP</div>
					<h3 class="text-2xl md:text-3xl font-bold text-black mb-3">Download BusinessOS Desktop</h3>
					<p class="text-gray-500 max-w-xl mx-auto text-sm">
						Native app with global shortcuts, voice input, screenshot capture, and meeting recording.
					</p>
				</div>

				<div class="grid md:grid-cols-3 gap-4 max-w-3xl mx-auto">
					<!-- macOS -->
					<div class="bg-white border-2 border-gray-200 rounded-xl p-5 hover:border-black hover:shadow-lg transition-all group">
						<div class="flex items-center gap-3 mb-4">
							<div class="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center group-hover:bg-black group-hover:text-white transition-colors">
								<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
									<path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
								</svg>
							</div>
							<div>
								<div class="font-bold text-black text-sm">macOS</div>
								<div class="text-[10px] text-gray-400 font-mono">Silicon & Intel</div>
							</div>
						</div>
						<div class="flex gap-2">
							<a
								href="/downloads/BusinessOS-1.0.1-arm64.dmg"
								download
								class="flex-1 bg-black text-white py-2.5 rounded-lg text-xs font-medium hover:bg-gray-800 transition-colors font-mono flex items-center justify-center gap-1.5"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
								</svg>
								Silicon
							</a>
							<a
								href="/downloads/BusinessOS-1.0.1-arm64.dmg"
								download
								class="flex-1 bg-gray-100 text-black py-2.5 rounded-lg text-xs font-medium hover:bg-gray-200 transition-colors font-mono flex items-center justify-center gap-1.5"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
								</svg>
								Intel
							</a>
						</div>
					</div>

					<!-- Windows -->
					<div class="bg-white border-2 border-gray-200 rounded-xl p-5 transition-all opacity-60">
						<div class="flex items-center gap-3 mb-4">
							<div class="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
								<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
									<path d="M3 5.548l7.262-1.006v7.021H3V5.548zm0 12.904l7.262 1.006v-6.895H3v5.889zm8.138 1.118L21 21v-7.437h-9.862v6.007zm0-14.036v6.029H21V3l-9.862 1.534z"/>
								</svg>
							</div>
							<div>
								<div class="font-bold text-black text-sm">Windows</div>
								<div class="text-[10px] text-gray-400 font-mono">Windows 10+</div>
							</div>
						</div>
						<button
							disabled
							class="w-full bg-gray-200 text-gray-400 py-2.5 rounded-lg text-xs font-medium cursor-not-allowed font-mono"
						>
							Coming Soon
						</button>
					</div>

					<!-- Linux -->
					<div class="bg-white border-2 border-gray-200 rounded-xl p-5 transition-all opacity-60">
						<div class="flex items-center gap-3 mb-4">
							<div class="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
								<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
									<path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.132 1.884 1.071.771-.06 1.592-.536 2.257-1.306.631-.765 1.683-1.084 2.378-1.503.348-.199.629-.469.649-.853.023-.4-.2-.811-.714-1.376v-.097l-.003-.003c-.17-.2-.25-.535-.338-.926-.085-.401-.182-.786-.492-1.046h-.003c-.059-.054-.123-.067-.188-.135a.357.357 0 00-.19-.064c.431-1.278.264-2.55-.173-3.694-.533-1.41-1.465-2.638-2.175-3.483-.796-1.005-1.576-1.957-1.56-3.368.026-2.152.236-6.133-3.544-6.139z"/>
								</svg>
							</div>
							<div>
								<div class="font-bold text-black text-sm">Linux</div>
								<div class="text-[10px] text-gray-400 font-mono">.deb / .rpm</div>
							</div>
						</div>
						<button
							disabled
							class="w-full bg-gray-200 text-gray-400 py-2.5 rounded-lg text-xs font-medium cursor-not-allowed font-mono"
						>
							Coming Soon
						</button>
					</div>
				</div>

				<!-- Features row -->
				<div class="mt-10 flex flex-wrap justify-center gap-6 text-center">
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
						</svg>
						<span class="text-xs text-gray-500 font-mono">Global Shortcuts</span>
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
						</svg>
						<span class="text-xs text-gray-500 font-mono">Voice Input</span>
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						<span class="text-xs text-gray-500 font-mono">Screenshot Capture</span>
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
						<span class="text-xs text-gray-500 font-mono">Meeting Recording</span>
					</div>
				</div>
			</div>
		</section>

		<!-- Capabilities -->
		<section class="py-20 px-6">
			<div class="max-w-5xl mx-auto">
				<div class="grid grid-cols-2 md:grid-cols-4 gap-8">
					{#each capabilities as cap, i}
						<div in:fly={{ y: 30, duration: 500, delay: 100 * i }}>
							<h3 class="font-mono text-sm font-bold text-black mb-2 tracking-wider">{cap.title.toUpperCase()}</h3>
							<p class="text-gray-500 text-sm">{cap.desc}</p>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- Modules Grid -->
		<section class="py-20 px-6">
			<div class="max-w-5xl mx-auto">
				<div class="flex items-center justify-between mb-8">
					<h2 class="font-mono text-xs text-gray-400 tracking-wider">AVAILABLE MODULES</h2>
					<a href="/docs" class="text-xs text-gray-500 hover:text-black transition-colors font-mono flex items-center gap-1">
						View all docs
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				</div>
				<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4">
					{#each modules as mod, i}
						<a
							href="/docs/{mod.slug}"
							class="border border-gray-200 rounded-lg p-4 hover:border-gray-400 hover:bg-gray-50 hover:shadow-sm transition-all group block"
							in:fly={{ y: 20, duration: 400, delay: 50 * i }}
						>
							<div class="font-medium text-gray-900 text-sm mb-1 group-hover:text-black">{mod.name}</div>
							<div class="text-xs text-gray-400 font-mono">{mod.desc}</div>
						</a>
					{/each}
				</div>
			</div>
		</section>

		<!-- How It Works -->
		<section class="py-24 px-6 bg-gray-50">
			<div class="max-w-5xl mx-auto">
				<div class="text-center mb-16">
					<h2 class="font-mono text-xs text-gray-400 mb-4 tracking-wider">HOW IT WORKS</h2>
					<h3 class="text-3xl md:text-4xl font-bold text-black mb-4">Context-driven intelligence</h3>
					<p class="text-gray-500 max-w-2xl mx-auto">
						Business OS uses a context-first approach. Your data, documents, and knowledge inform AI agents that understand your business.
					</p>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
					{#each howItWorks as item, i}
						<div
							class="relative bg-white border border-gray-200 rounded-xl p-6 hover:border-gray-300 hover:shadow-lg transition-all group"
							in:fly={{ y: 30, duration: 500, delay: 100 * i }}
						>
							<div class="flex items-center gap-3 mb-4">
								<div class="w-10 h-10 bg-black rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={item.icon} />
									</svg>
								</div>
								<span class="font-mono text-xs text-gray-400">{item.step}</span>
							</div>
							<h4 class="font-bold text-black mb-2">{item.title}</h4>
							<p class="text-sm text-gray-500 leading-relaxed">{item.desc}</p>
							{#if i < howItWorks.length - 1}
								<div class="hidden lg:block absolute -right-3 top-1/2 transform -translate-y-1/2 text-gray-300">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</div>
							{/if}
						</div>
					{/each}
				</div>

				<!-- Flow diagram -->
				<div class="mt-16 bg-white border border-gray-200 rounded-xl p-8 md:p-12">
					<div class="flex flex-col md:flex-row items-center justify-between gap-8">
						<div class="text-center md:text-left">
							<div class="font-mono text-xs text-gray-400 mb-2">THE FLOW</div>
							<h4 class="text-xl font-bold text-black mb-3">From data to action</h4>
							<p class="text-gray-500 text-sm max-w-md">
								Import your existing data, create contexts for each domain, link them to projects, and let AI agents handle the rest.
							</p>
						</div>
						<div class="flex items-center gap-3 font-mono text-sm">
							<div class="px-4 py-2 bg-gray-100 rounded-lg text-gray-600">Data</div>
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
							<div class="px-4 py-2 bg-gray-100 rounded-lg text-gray-600">Context</div>
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
							<div class="px-4 py-2 bg-gray-100 rounded-lg text-gray-600">Project</div>
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
							<div class="px-4 py-2 bg-black rounded-lg text-white">Agent</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- aMCP Integrations -->
		<section class="py-24 px-6">
			<div class="max-w-5xl mx-auto">
				<div class="grid md:grid-cols-2 gap-12 items-center">
					<div>
						<div class="font-mono text-xs text-gray-400 mb-4 tracking-wider">AMCP PROTOCOL</div>
						<h3 class="text-3xl font-bold text-black mb-4">Connect everything via aMCP</h3>
						<p class="text-gray-500 mb-6">
							Business OS uses <a href="https://amcp.ai" target="_blank" rel="noopener" class="text-black hover:underline font-medium">aMCP</a> (Advanced Model Context Protocol) to integrate with your existing tools. Connect CRMs, databases, APIs, and legacy systems - all through a unified interface.
						</p>
						<div class="space-y-4">
							<div class="flex items-start gap-3">
								<div class="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
									<svg class="w-3.5 h-3.5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								</div>
								<div>
									<div class="font-medium text-black text-sm">AI-Generated Integrations</div>
									<div class="text-gray-500 text-sm">OSA can build custom aMCP integrations for your legacy systems</div>
								</div>
							</div>
							<div class="flex items-start gap-3">
								<div class="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
									<svg class="w-3.5 h-3.5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								</div>
								<div>
									<div class="font-medium text-black text-sm">Enterprise Scale</div>
									<div class="text-gray-500 text-sm">Handle thousands of connections with proper rate limiting and auth</div>
								</div>
							</div>
							<div class="flex items-start gap-3">
								<div class="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
									<svg class="w-3.5 h-3.5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								</div>
								<div>
									<div class="font-medium text-black text-sm">Bi-directional Sync</div>
									<div class="text-gray-500 text-sm">Keep data synchronized across all your connected systems</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Integrations grid -->
					<div class="bg-gray-50 rounded-xl p-6 border border-gray-200">
						<div class="font-mono text-xs text-gray-400 mb-4">SUPPORTED INTEGRATIONS</div>
						<div class="grid grid-cols-3 gap-3">
							{#each integrations as int, i}
								<div
									class="bg-white border border-gray-200 rounded-lg p-3 hover:border-gray-400 hover:shadow-sm transition-all text-center"
									in:fly={{ y: 10, duration: 300, delay: 50 * i }}
								>
									<div class="font-medium text-gray-900 text-xs">{int.name}</div>
									<div class="text-[10px] text-gray-400 font-mono mt-1">{int.category}</div>
								</div>
							{/each}
						</div>
						<div class="mt-4 text-center">
							<a href="https://amcp.ai" target="_blank" rel="noopener" class="text-xs text-gray-400 font-mono hover:text-gray-600 transition-colors">+ Custom integrations via aMCP →</a>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Data Import -->
		<section class="py-24 px-6 bg-black text-white">
			<div class="max-w-5xl mx-auto">
				<div class="text-center mb-12">
					<div class="font-mono text-xs text-gray-500 mb-4 tracking-wider">DATA MIGRATION</div>
					<h3 class="text-3xl font-bold mb-4">Import your existing data</h3>
					<p class="text-gray-400 max-w-2xl mx-auto">
						Migrate from your current tools in minutes. We support direct imports from popular platforms and custom data formats.
					</p>
				</div>

				<div class="grid md:grid-cols-3 gap-6">
					<div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
						<div class="w-10 h-10 bg-blue-500/20 rounded-lg flex items-center justify-center mb-4">
							<svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
						</div>
						<h4 class="font-bold text-white mb-2">CRM Data</h4>
						<p class="text-gray-400 text-sm mb-4">Import contacts, deals, and activities from Salesforce, HubSpot, or GHL.</p>
						<div class="flex flex-wrap gap-2">
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">Contacts</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">Deals</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">Activities</span>
						</div>
					</div>

					<div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
						<div class="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center mb-4">
							<svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
							</svg>
						</div>
						<h4 class="font-bold text-white mb-2">Spreadsheets & Databases</h4>
						<p class="text-gray-400 text-sm mb-4">Migrate from Airtable, Google Sheets, Excel, or connect directly to databases.</p>
						<div class="flex flex-wrap gap-2">
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">CSV</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">JSON</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">SQL</span>
						</div>
					</div>

					<div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
						<div class="w-10 h-10 bg-purple-500/20 rounded-lg flex items-center justify-center mb-4">
							<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
						</div>
						<h4 class="font-bold text-white mb-2">Documents & Files</h4>
						<p class="text-gray-400 text-sm mb-4">Import documents from Notion, Google Docs, or upload files directly.</p>
						<div class="flex flex-wrap gap-2">
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">PDF</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">Markdown</span>
							<span class="px-2 py-1 bg-gray-800 rounded text-xs text-gray-400">Docs</span>
						</div>
					</div>
				</div>

				<div class="mt-12 text-center">
					<div class="inline-flex items-center gap-4 bg-gray-900 border border-gray-800 rounded-full px-6 py-3">
						<span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
						<span class="font-mono text-sm text-gray-400">AI-assisted mapping automatically matches your data schema</span>
					</div>
				</div>
			</div>
		</section>

		<!-- OSA & Agent Capabilities Section -->
		<section class="py-24 px-6 bg-gray-50">
			<div class="max-w-5xl mx-auto">
				<div class="grid md:grid-cols-2 gap-12 items-start">
					<!-- OSA Info -->
					<div class="bg-white border border-gray-200 rounded-xl p-8">
						<div class="font-mono text-xs text-gray-400 mb-4 tracking-wider">POWERED BY</div>
						<h3 class="text-2xl font-bold text-black mb-4">OSA - The OS Agent</h3>
						<p class="text-gray-500 mb-6">
							Business OS was built with OSA. The same agent that created this system can extend it for your needs - building custom integrations, automations, and features on demand.
						</p>
						<div class="space-y-3 mb-6">
							<div class="flex items-center gap-3 text-sm">
								<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								<span class="text-gray-600">Auto-generate aMCP integrations for legacy systems</span>
							</div>
							<div class="flex items-center gap-3 text-sm">
								<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								<span class="text-gray-600">Create custom modules and workflows</span>
							</div>
							<div class="flex items-center gap-3 text-sm">
								<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								<span class="text-gray-600">Extend the system without developers</span>
							</div>
						</div>
						<a
							href="https://osa.dev"
							target="_blank"
							rel="noopener"
							class="inline-flex items-center gap-2 text-sm font-medium text-black hover:text-gray-600 transition-colors font-mono"
						>
							Learn more at osa.dev
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
							</svg>
						</a>
					</div>

					<!-- Enterprise Features -->
					<div class="space-y-6">
						<div class="bg-white border border-gray-200 rounded-xl p-6">
							<div class="flex items-start gap-4">
								<div class="w-10 h-10 bg-black rounded-lg flex items-center justify-center flex-shrink-0">
									<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
									</svg>
								</div>
								<div>
									<h4 class="font-bold text-black mb-1">Enterprise Security</h4>
									<p class="text-gray-500 text-sm">SOC2 ready. Self-hosted on your infrastructure with full data sovereignty.</p>
								</div>
							</div>
						</div>

						<div class="bg-white border border-gray-200 rounded-xl p-6">
							<div class="flex items-start gap-4">
								<div class="w-10 h-10 bg-black rounded-lg flex items-center justify-center flex-shrink-0">
									<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
									</svg>
								</div>
								<div>
									<h4 class="font-bold text-black mb-1">Team Collaboration</h4>
									<p class="text-gray-500 text-sm">Role-based access, shared contexts, and team workspaces built for scale.</p>
								</div>
							</div>
						</div>

						<div class="bg-white border border-gray-200 rounded-xl p-6">
							<div class="flex items-start gap-4">
								<div class="w-10 h-10 bg-black rounded-lg flex items-center justify-center flex-shrink-0">
									<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
									</svg>
								</div>
								<div>
									<h4 class="font-bold text-black mb-1">Local or Cloud LLMs</h4>
									<p class="text-gray-500 text-sm">Run Ollama locally or connect to OpenAI, Anthropic, or your preferred provider.</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- CTA -->
		<section class="py-24 px-6 bg-black text-white">
			<div class="max-w-3xl mx-auto text-center">
				<div class="font-mono text-xs text-gray-500 mb-4 tracking-wider">GET STARTED</div>
				<h2 class="text-3xl md:text-4xl font-bold mb-4">Ready to take control?</h2>
				<p class="text-gray-400 mb-8 max-w-xl mx-auto">
					Deploy on your infrastructure, import your data, and let AI agents handle the rest. Your business operating system awaits.
				</p>
				<div class="flex flex-col sm:flex-row items-center justify-center gap-4">
					<a href="/register" class="btn-pill btn-pill-light font-mono inline-flex items-center gap-2 px-8 py-3">
						Initialize workspace
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
						</svg>
					</a>
					<a href="https://github.com" target="_blank" rel="noopener" class="btn-pill btn-pill-outline-dark font-mono inline-flex items-center gap-2 px-6 py-3">
						<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
						</svg>
						View on GitHub
					</a>
				</div>
				<div class="mt-12 flex flex-wrap justify-center gap-8 font-mono text-xs text-gray-500">
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Self-hosted
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						End-to-end encrypted
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Open source
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						MIT Licensed
					</div>
				</div>
			</div>
		</section>

		<!-- Footer -->
		<footer class="border-t border-gray-100 py-12 px-6">
			<div class="max-w-4xl mx-auto">
				<div class="flex flex-col md:flex-row items-center justify-between gap-6">
					<div class="flex items-baseline gap-0.5">
						<span class="text-black text-sm font-extrabold tracking-[0.1em] font-mono">BUSINESS</span>
						<span class="text-black/30 text-xs font-light font-mono">OS</span>
					</div>
					<div class="flex items-center gap-6 font-mono text-xs text-gray-400">
						<a href="/docs" class="hover:text-black transition-colors">Docs</a>
						<a href="https://github.com" target="_blank" rel="noopener" class="hover:text-black transition-colors">GitHub</a>
						<a href="https://amcp.ai" target="_blank" rel="noopener" class="hover:text-black transition-colors">aMCP</a>
						<a href="https://osa.dev" target="_blank" rel="noopener" class="hover:text-black transition-colors">OSA</a>
						<a href="/terms" class="hover:text-black transition-colors">Terms</a>
						<a href="/privacy" class="hover:text-black transition-colors">Privacy</a>
					</div>
				</div>
				<div class="mt-8 text-center md:text-left">
					<p class="font-mono text-xs text-gray-400">
						Open Source · MIT Licensed · v0.0.1
					</p>
				</div>
			</div>
		</footer>
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

	.typewriter-text {
		opacity: 0;
		animation: fadeIn 0.3s ease forwards;
	}

	@keyframes fadeIn {
		from { opacity: 0; transform: translateX(-5px); }
		to { opacity: 1; transform: translateX(0); }
	}

	/* Glitch effect for headline */
	.glitch-text {
		position: relative;
		display: inline-block;
	}

	.glitch-text::before,
	.glitch-text::after {
		content: attr(data-text);
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		opacity: 0;
	}

	.glitch-text::before {
		color: #00ffff;
		animation: glitch-1 4s infinite linear alternate-reverse;
	}

	.glitch-text::after {
		color: #ff00ff;
		animation: glitch-2 3s infinite linear alternate-reverse;
	}

	@keyframes glitch-1 {
		0%, 94% { opacity: 0; transform: translate(0); }
		95% { opacity: 0.3; transform: translate(-2px, 1px); }
		96% { opacity: 0; transform: translate(0); }
		97% { opacity: 0.2; transform: translate(2px, -1px); }
		98%, 100% { opacity: 0; transform: translate(0); }
	}

	@keyframes glitch-2 {
		0%, 92% { opacity: 0; transform: translate(0); }
		93% { opacity: 0.2; transform: translate(1px, -1px); }
		94% { opacity: 0; transform: translate(0); }
		95% { opacity: 0.3; transform: translate(-1px, 1px); }
		96%, 100% { opacity: 0; transform: translate(0); }
	}

	/* Scan line effect for sections */
	:global(.scan-line) {
		position: relative;
		overflow: hidden;
	}

	:global(.scan-line)::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 2px;
		background: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent);
		animation: scanMove 3s linear infinite;
	}

	@keyframes scanMove {
		0% { transform: translateY(-100%); }
		100% { transform: translateY(100vh); }
	}

	/* Scan line moving animation */
	.scan-line-moving {
		animation: scanLineMove 4s linear infinite;
	}

	@keyframes scanLineMove {
		0% { transform: translateY(-10px); }
		100% { transform: translateY(100vh); }
	}

	/* Floating particles */
	:global(.animate-float-slow) {
		animation: floatSlow 8s ease-in-out infinite;
	}

	:global(.animate-float-medium) {
		animation: floatMedium 6s ease-in-out infinite;
	}

	:global(.animate-float-fast) {
		animation: floatFast 4s ease-in-out infinite;
	}

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

	/* Pulse ring effect */
	:global(.pulse-ring) {
		animation: pulseRing 2s ease-out infinite;
	}

	@keyframes pulseRing {
		0% { transform: scale(1); opacity: 0.5; }
		100% { transform: scale(1.5); opacity: 0; }
	}
</style>
