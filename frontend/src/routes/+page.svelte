<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';

	const session = useSession();

	$effect(() => {
		if (!$session.isPending && $session.data) {
			goto('/chat');
		}
	});

	let scrolled = $state(false);

	onMount(() => {
		const handleScroll = () => {
			scrolled = window.scrollY > 20;
		};
		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});

	let visibleSections = $state<Record<string, boolean>>({});

	function observeSection(node: HTMLElement, id: string) {
		const observer = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						visibleSections[id] = true;
					}
				});
			},
			{ threshold: 0.15 }
		);
		observer.observe(node);
		return {
			destroy() {
				observer.disconnect();
			}
		};
	}

	const techStack = [
		{ name: 'SvelteKit', category: 'Frontend' },
		{ name: 'FastAPI', category: 'Backend' },
		{ name: 'PostgreSQL', category: 'Database' },
		{ name: 'Ollama', category: 'Local AI' },
		{ name: 'Better Auth', category: 'Auth' },
		{ name: 'Tailwind', category: 'Styling' }
	];
</script>

<div class="min-h-screen bg-white">
	<!-- Header -->
	<header
		class="fixed top-0 left-0 right-0 z-50 transition-all duration-300
			{scrolled ? 'bg-white/95 backdrop-blur-md border-b border-gray-200' : 'bg-transparent'}"
		in:fade={{ duration: 300 }}
	>
		<div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<div class="w-8 h-8 bg-black rounded-lg flex items-center justify-center">
					<span class="text-white text-sm font-bold">B</span>
				</div>
				<span class="font-semibold text-gray-900">Business OS</span>
			</div>
			<nav class="hidden md:flex items-center gap-8">
				<a href="#features" class="text-sm text-gray-600 hover:text-black transition-colors">Features</a>
				<a href="#stack" class="text-sm text-gray-600 hover:text-black transition-colors">Stack</a>
				<a href="https://github.com" target="_blank" rel="noopener" class="text-sm text-gray-600 hover:text-black transition-colors">GitHub</a>
			</nav>
			<a href="/register" class="bg-black text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors">
				Get Started
			</a>
		</div>
	</header>

	<!-- Hero -->
	<section class="pt-32 pb-20 px-6">
		<div class="max-w-4xl mx-auto text-center">
			<div
				class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gray-100 text-sm text-gray-700 mb-8"
				in:fly={{ y: 20, duration: 500 }}
			>
				<span class="w-2 h-2 bg-green-500 rounded-full"></span>
				Open Source · Self-Hosted · AI-Native
			</div>

			<h1
				class="text-5xl md:text-7xl font-bold text-black leading-[1.1] mb-6 tracking-tight"
				in:fly={{ y: 30, duration: 600, delay: 100 }}
			>
				Business OS
			</h1>

			<p
				class="text-xl md:text-2xl text-gray-600 max-w-2xl mx-auto mb-4"
				in:fly={{ y: 30, duration: 600, delay: 200 }}
			>
				Your operating system for the agentic era.
			</p>
			<p
				class="text-lg text-gray-500 max-w-xl mx-auto mb-10"
				in:fly={{ y: 30, duration: 600, delay: 300 }}
			>
				AI-native. Self-hosted. Built for fast software.
			</p>

			<div
				class="flex flex-col sm:flex-row items-center justify-center gap-4"
				in:fly={{ y: 30, duration: 600, delay: 400 }}
			>
				<a href="/register" class="bg-black text-white px-8 py-3.5 rounded-xl text-base font-medium hover:bg-gray-800 transition-colors">
					Get Started
				</a>
				<a href="/login" class="bg-white text-black px-8 py-3.5 rounded-xl text-base font-medium border border-gray-300 hover:border-gray-400 transition-colors">
					Live Demo
				</a>
			</div>
		</div>
	</section>

	<!-- App Preview -->
	<section class="pb-24 px-6" in:fly={{ y: 40, duration: 800, delay: 500 }}>
		<div class="max-w-5xl mx-auto">
			<div class="relative bg-gray-950 rounded-2xl p-1.5 shadow-2xl ring-1 ring-white/10">
				<div class="flex items-center gap-2 px-4 py-3">
					<div class="flex gap-1.5">
						<div class="w-3 h-3 rounded-full bg-red-500"></div>
						<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
					</div>
				</div>
				<div class="bg-white rounded-xl overflow-hidden">
					<div class="flex min-h-[420px]">
						<div class="w-60 bg-gray-50 border-r border-gray-200 p-4 hidden sm:block">
							<div class="flex items-center gap-2 mb-8">
								<div class="w-8 h-8 bg-black rounded-lg"></div>
								<span class="font-medium text-gray-900 text-sm">Your Company</span>
							</div>
							<div class="space-y-1">
								{#each ['Chat', 'Projects', 'Contexts', 'Daily Log'] as item, i}
									<div class="flex items-center gap-3 px-3 py-2.5 {i === 0 ? 'bg-black text-white' : 'text-gray-600 hover:bg-gray-100'} rounded-lg text-sm transition-colors">
										<div class="w-4 h-4 rounded {i === 0 ? 'bg-gray-600' : 'bg-gray-300'}"></div>
										{item}
									</div>
								{/each}
							</div>
						</div>
						<div class="flex-1 p-6">
							<div class="mb-6">
								<div class="h-7 w-48 bg-gray-100 rounded-lg mb-2"></div>
								<div class="h-4 w-72 bg-gray-50 rounded"></div>
							</div>
							<div class="space-y-4">
								<div class="flex items-start gap-3">
									<div class="w-8 h-8 rounded-full bg-gray-200 flex-shrink-0"></div>
									<div class="bg-gray-100 rounded-2xl rounded-tl-md px-4 py-3 max-w-md">
										<div class="h-3 w-64 bg-gray-200 rounded mb-2"></div>
										<div class="h-3 w-48 bg-gray-200 rounded"></div>
									</div>
								</div>
								<div class="flex items-start gap-3 justify-end">
									<div class="bg-black rounded-2xl rounded-tr-md px-4 py-3 max-w-md">
										<div class="h-3 w-56 bg-gray-700 rounded mb-2"></div>
										<div class="h-3 w-72 bg-gray-700 rounded mb-2"></div>
										<div class="h-3 w-44 bg-gray-700 rounded"></div>
									</div>
									<div class="w-8 h-8 rounded-full bg-black flex-shrink-0"></div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Trust Strip -->
	<section class="py-8 px-6 border-y border-gray-100" in:fade={{ duration: 400, delay: 600 }}>
		<div class="max-w-5xl mx-auto">
			<div class="flex flex-wrap items-center justify-center gap-x-12 gap-y-4 text-sm">
				<span class="flex items-center gap-2 text-gray-600">
					<svg class="w-5 h-5 text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					Your data stays yours
				</span>
				<span class="flex items-center gap-2 text-gray-600">
					<svg class="w-5 h-5 text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
					AI agents built-in
				</span>
				<span class="flex items-center gap-2 text-gray-600">
					<svg class="w-5 h-5 text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
					Fast to customize
				</span>
				<span class="flex items-center gap-2 text-gray-600">
					<svg class="w-5 h-5 text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					Open source
				</span>
			</div>
		</div>
	</section>

	<!-- Agentic Era Section -->
	<section class="py-28 px-6" use:observeSection={'agentic'}>
		<div class="max-w-5xl mx-auto">
			{#if visibleSections['agentic']}
				<div class="text-center mb-16">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						Built for the Agentic Era
					</h2>
					<p class="text-xl text-gray-600 max-w-2xl mx-auto" in:fly={{ y: 30, duration: 500, delay: 100 }}>
						AI agents that work for you, connect your tools, and operate on your terms.
					</p>
				</div>

				<!-- Flow Diagram -->
				<div class="bg-gray-50 rounded-2xl p-8 md:p-12 mb-16" in:fly={{ y: 30, duration: 500, delay: 200 }}>
					<div class="flex flex-col md:flex-row items-center justify-center gap-4 md:gap-8">
						<div class="bg-white border border-gray-200 rounded-xl p-6 text-center min-w-[160px] shadow-sm">
							<svg class="w-8 h-8 mx-auto mb-3 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
							</svg>
							<p class="font-semibold text-gray-900">Your OS</p>
							<p class="text-xs text-gray-500 mt-1">Data + LLMs</p>
						</div>

						<svg class="w-8 h-8 text-gray-400 rotate-90 md:rotate-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
						</svg>

						<div class="bg-black text-white rounded-xl p-6 text-center min-w-[160px] shadow-lg">
							<svg class="w-8 h-8 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
							</svg>
							<p class="font-semibold">Agent</p>
							<p class="text-xs text-gray-400 mt-1">AMCP Ready</p>
						</div>

						<svg class="w-8 h-8 text-gray-400 rotate-90 md:rotate-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
						</svg>

						<div class="bg-white border border-gray-200 rounded-xl p-6 text-center min-w-[160px] shadow-sm">
							<svg class="w-8 h-8 mx-auto mb-3 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
							</svg>
							<p class="font-semibold text-gray-900">External</p>
							<p class="text-xs text-gray-500 mt-1">Tools & APIs</p>
						</div>
					</div>
					<p class="text-center text-sm text-gray-500 mt-8">
						You control the flow. Nothing leaves without your permission.
					</p>
				</div>

				<!-- Cards -->
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each [
						{
							title: 'AI Agents',
							description: 'Agents that execute tasks for you, connected to your tools and data',
							icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
						},
						{
							title: 'AMCP Ready',
							description: 'Connect to any tool securely via Advanced Model Context Protocol',
							icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1'
						},
						{
							title: 'Any LLM',
							description: 'Local or cloud - you choose. Ollama, OpenAI, Anthropic, whatever',
							icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
						}
					] as card, i}
						<div
							class="bg-white border border-gray-200 rounded-2xl p-6 hover:border-gray-300 hover:shadow-lg transition-all"
							in:fly={{ y: 30, duration: 500, delay: 300 + i * 100 }}
						>
							<div class="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center mb-4">
								<svg class="w-6 h-6 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={card.icon} />
								</svg>
							</div>
							<h3 class="font-semibold text-gray-900 mb-2">{card.title}</h3>
							<p class="text-gray-600 text-sm leading-relaxed">{card.description}</p>
						</div>
					{/each}
				</div>

				<p class="text-center text-sm text-gray-500 mt-10" in:fly={{ y: 20, duration: 400, delay: 700 }}>
					Works with <a href="https://amcp.ai" target="_blank" rel="noopener" class="text-black font-medium hover:underline">AMCP</a> for secure agent-to-world connections
				</p>
			{/if}
		</div>
	</section>

	<!-- Features Section -->
	<section id="features" class="py-28 px-6 bg-gray-50" use:observeSection={'features'}>
		<div class="max-w-5xl mx-auto">
			{#if visibleSections['features']}
				<div class="text-center mb-16">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						Everything you need
					</h2>
					<p class="text-gray-600 max-w-2xl mx-auto" in:fly={{ y: 30, duration: 500, delay: 100 }}>
						A complete business operating system. Projects, tasks, team, AI chat, and more.
					</p>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					{#each [
						{ icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', title: 'Projects', description: 'Track work across your business' },
						{ icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4', title: 'Tasks', description: 'Kanban, assign, due dates' },
						{ icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z', title: 'Team', description: 'Org chart, capacity, workload' },
						{ icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z', title: 'AI Chat', description: 'Chat with local AI, on-device' },
						{ icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', title: 'Artifacts', description: 'Generate code, docs, anything' },
						{ icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10', title: 'Contexts', description: 'Store your business knowledge' },
						{ icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z', title: 'Daily Log', description: 'Track your day and patterns' },
						{ icon: 'M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z', title: 'Dashboard', description: 'Your daily command center' },
						{ icon: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z', title: 'Search', description: 'Find anything instantly' }
					] as feature, i}
						<div
							class="bg-white border border-gray-200 rounded-xl p-5 hover:border-gray-300 hover:shadow-sm transition-all"
							in:fly={{ y: 20, duration: 400, delay: 150 + i * 40 }}
						>
							<div class="flex items-start gap-4">
								<div class="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center flex-shrink-0">
									<svg class="w-5 h-5 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={feature.icon} />
									</svg>
								</div>
								<div>
									<h3 class="font-semibold text-gray-900 mb-1">{feature.title}</h3>
									<p class="text-gray-600 text-sm">{feature.description}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<!-- Your Data Section -->
	<section class="py-28 px-6" use:observeSection={'data'}>
		<div class="max-w-5xl mx-auto">
			{#if visibleSections['data']}
				<div class="text-center mb-16">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						Your Data. Your Rules.
					</h2>
					<p class="text-xl text-gray-600 max-w-2xl mx-auto" in:fly={{ y: 30, duration: 500, delay: 100 }}>
						No Big Tech watching. No data harvesting. You control what stays and what goes.
					</p>
				</div>

				<!-- Diagram -->
				<div class="bg-black rounded-2xl p-8 md:p-12 mb-16 text-white" in:fly={{ y: 30, duration: 500, delay: 200 }}>
					<div class="max-w-2xl mx-auto">
						<div class="bg-white/10 backdrop-blur rounded-xl p-6 mb-6">
							<p class="text-xs font-medium text-gray-400 mb-4 text-center tracking-wider">YOUR SERVERS</p>
							<div class="flex flex-wrap justify-center gap-3">
								<div class="bg-white/10 rounded-lg px-4 py-2.5 text-sm flex items-center gap-2">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
									</svg>
									Data
								</div>
								<div class="bg-white/10 rounded-lg px-4 py-2.5 text-sm flex items-center gap-2">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
									</svg>
									LLMs
								</div>
								<div class="bg-white/10 rounded-lg px-4 py-2.5 text-sm flex items-center gap-2">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
									</svg>
									Agents
								</div>
							</div>
							<p class="text-center text-sm text-gray-400 mt-4 flex items-center justify-center gap-2">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
								</svg>
								Nothing leaves without permission
							</p>
						</div>

						<div class="flex justify-center">
							<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
							</svg>
						</div>

						<div class="text-center my-4">
							<span class="inline-flex items-center gap-2 bg-green-500/20 text-green-400 px-4 py-2 rounded-full text-sm font-medium">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								YOU CONTROL
							</span>
						</div>

						<div class="flex justify-center">
							<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
							</svg>
						</div>

						<div class="border border-dashed border-gray-600 rounded-xl p-4 mt-4">
							<p class="text-xs font-medium text-gray-500 mb-2 text-center tracking-wider">EXTERNAL (Optional)</p>
							<p class="text-center text-gray-500 text-sm">
								Cloud LLMs · APIs · Integrations
							</p>
						</div>
					</div>
				</div>

				<!-- Cards -->
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each [
						{ icon: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01', title: 'Self-hosted', description: 'Runs on your servers. Docker, VPS, bare metal - your choice' },
						{ icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z', title: 'You Decide', description: 'Choose what connects externally. Default is local-only' },
						{ icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z', title: 'Train Privately', description: 'Train your AI with your data. No Big Tech watching' }
					] as card, i}
						<div
							class="bg-white border border-gray-200 rounded-2xl p-6 hover:border-gray-300 hover:shadow-lg transition-all"
							in:fly={{ y: 30, duration: 500, delay: 300 + i * 100 }}
						>
							<div class="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center mb-4">
								<svg class="w-6 h-6 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={card.icon} />
								</svg>
							</div>
							<h3 class="font-semibold text-gray-900 mb-2">{card.title}</h3>
							<p class="text-gray-600 text-sm leading-relaxed">{card.description}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<!-- Fast Software Section -->
	<section class="py-28 px-6 bg-gray-50" use:observeSection={'fast'}>
		<div class="max-w-5xl mx-auto">
			{#if visibleSections['fast']}
				<div class="text-center mb-16">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						The Era of Fast Software
					</h2>
					<p class="text-xl text-gray-600 max-w-2xl mx-auto" in:fly={{ y: 30, duration: 500, delay: 100 }}>
						Build and customize faster than ever. No waiting for vendors.
					</p>
				</div>

				<!-- Comparison -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-16 max-w-3xl mx-auto">
					<div class="bg-white border border-gray-200 rounded-2xl p-8" in:fly={{ x: -30, duration: 500, delay: 200 }}>
						<p class="text-xs font-medium text-gray-400 mb-6 tracking-wider">TRADITIONAL</p>
						<div class="space-y-4 text-gray-500">
							<div class="flex items-center gap-3">
								<div class="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs">1</div>
								<span>Request feature</span>
							</div>
							<div class="flex items-center gap-3">
								<div class="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs">2</div>
								<span>Wait 6 months</span>
							</div>
							<div class="flex items-center gap-3">
								<div class="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs">3</div>
								<span>Maybe get it</span>
							</div>
						</div>
					</div>
					<div class="bg-black text-white rounded-2xl p-8" in:fly={{ x: 30, duration: 500, delay: 200 }}>
						<p class="text-xs font-medium text-gray-400 mb-6 tracking-wider">FAST SOFTWARE</p>
						<div class="space-y-4">
							<div class="flex items-center gap-3">
								<div class="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center text-xs">1</div>
								<span>Need a feature?</span>
							</div>
							<div class="flex items-center gap-3">
								<div class="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center text-xs">2</div>
								<span>Build it today</span>
							</div>
							<div class="flex items-center gap-3 text-green-400">
								<div class="w-6 h-6 rounded-full bg-green-500/20 flex items-center justify-center">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								</div>
								<span>Ship it tonight</span>
							</div>
						</div>
					</div>
				</div>

				<!-- Bullets -->
				<div class="max-w-xl mx-auto" in:fly={{ y: 20, duration: 400, delay: 400 }}>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						{#each [
							'Modify anything',
							'Extend with features',
							'Integrate your tools',
							'Ship in hours'
						] as bullet}
							<div class="flex items-center gap-3 text-gray-700">
								<svg class="w-5 h-5 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
								{bullet}
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</section>

	<!-- OSA Section -->
	<section class="py-28 px-6" use:observeSection={'osa'}>
		<div class="max-w-4xl mx-auto">
			{#if visibleSections['osa']}
				<div
					class="relative overflow-hidden bg-gradient-to-br from-gray-900 via-gray-800 to-black rounded-3xl p-10 md:p-14"
					in:fly={{ y: 30, duration: 500 }}
				>
					<!-- Background Pattern -->
					<div class="absolute inset-0 opacity-5">
						<div class="absolute inset-0" style="background-image: url(&quot;data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E&quot;);"></div>
					</div>

					<div class="relative flex flex-col md:flex-row items-start gap-8">
						<div class="flex-shrink-0">
							<div class="w-16 h-16 bg-white/10 backdrop-blur rounded-2xl flex items-center justify-center">
								<svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
							</div>
						</div>
						<div class="flex-1">
							<p class="text-sm font-medium text-gray-400 mb-2">Built with OSA</p>
							<h3 class="text-2xl md:text-3xl font-bold text-white mb-4">The OS Agent</h3>
							<p class="text-gray-300 mb-6 leading-relaxed">
								Business OS was built using OSA - the OS Agent that works at the kernel level. Design, customize, and build your perfect operating system with AI assistance.
							</p>
							<p class="text-gray-400 mb-8">
								Want to build your own OS or customize this one?
							</p>
							<a
								href="https://osa.dev"
								target="_blank"
								rel="noopener"
								class="inline-flex items-center gap-3 bg-white text-black px-6 py-3 rounded-xl font-medium hover:bg-gray-100 transition-colors"
							>
								Join the waitlist
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
								</svg>
							</a>
							<p class="text-sm text-gray-500 mt-4">osa.dev</p>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</section>

	<!-- Tech Stack -->
	<section id="stack" class="py-28 px-6 bg-gray-50" use:observeSection={'stack'}>
		<div class="max-w-4xl mx-auto">
			{#if visibleSections['stack']}
				<div class="text-center mb-12">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						Modern Stack
					</h2>
					<p class="text-gray-600" in:fly={{ y: 30, duration: 500, delay: 100 }}>
						Built with tools you know and love.
					</p>
				</div>

				<div class="grid grid-cols-2 md:grid-cols-3 gap-4" in:fly={{ y: 30, duration: 500, delay: 200 }}>
					{#each techStack as tech, i}
						<div class="bg-white border border-gray-200 rounded-xl p-5 text-center hover:border-gray-300 transition-all">
							<p class="font-semibold text-gray-900">{tech.name}</p>
							<p class="text-sm text-gray-500">{tech.category}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<!-- Use Cases -->
	<section class="py-28 px-6" use:observeSection={'usecases'}>
		<div class="max-w-4xl mx-auto">
			{#if visibleSections['usecases']}
				<div class="text-center mb-12">
					<h2 class="text-3xl md:text-4xl font-bold text-black mb-4" in:fly={{ y: 30, duration: 500 }}>
						Built for
					</h2>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each [
						{ icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', title: 'Agencies', description: 'Manage clients, projects, and delivery in one place' },
						{ icon: 'M13 10V3L4 14h7v7l9-11h-7z', title: 'Startups', description: 'Move fast with tools you control and customize' },
						{ icon: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', title: 'Consultants & Freelancers', description: 'Your personal business command center' },
						{ icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4', title: 'Developers', description: 'A foundation to build on, not fight against' }
					] as useCase, i}
						<div
							class="bg-white border border-gray-200 rounded-xl p-6 hover:border-gray-300 hover:shadow-sm transition-all"
							in:fly={{ y: 20, duration: 400, delay: 100 + i * 100 }}
						>
							<div class="flex items-start gap-4">
								<div class="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center flex-shrink-0">
									<svg class="w-6 h-6 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={useCase.icon} />
									</svg>
								</div>
								<div>
									<h3 class="font-semibold text-gray-900 mb-1">{useCase.title}</h3>
									<p class="text-gray-600 text-sm">{useCase.description}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<!-- CTA -->
	<section class="py-28 px-6 bg-gray-50" use:observeSection={'cta'}>
		<div class="max-w-3xl mx-auto">
			{#if visibleSections['cta']}
				<div class="bg-black rounded-3xl p-12 md:p-16 text-center text-white" in:fly={{ y: 30, duration: 500 }}>
					<h2 class="text-3xl md:text-4xl font-bold mb-4">
						Ready to own your business system?
					</h2>
					<p class="text-xl text-gray-400 mb-10">
						Self-hosted. AI-native. Yours to customize.
					</p>
					<div class="flex flex-col sm:flex-row items-center justify-center gap-4">
						<a href="/register" class="bg-white text-black px-8 py-3.5 rounded-xl text-base font-medium hover:bg-gray-100 transition-colors">
							Get Started
						</a>
						<a
							href="https://github.com"
							target="_blank"
							rel="noopener"
							class="border border-gray-600 text-white px-8 py-3.5 rounded-xl text-base font-medium hover:bg-white/5 transition-colors flex items-center gap-2"
						>
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
							</svg>
							GitHub
						</a>
					</div>
				</div>
			{/if}
		</div>
	</section>

	<!-- Footer -->
	<footer class="border-t border-gray-200 py-16 px-6">
		<div class="max-w-5xl mx-auto">
			<div class="grid grid-cols-2 md:grid-cols-4 gap-12 mb-12">
				<div class="col-span-2 md:col-span-1">
					<div class="flex items-center gap-2 mb-4">
						<div class="w-8 h-8 bg-black rounded-lg flex items-center justify-center">
							<span class="text-white text-sm font-bold">B</span>
						</div>
						<span class="font-semibold text-gray-900">Business OS</span>
					</div>
					<p class="text-sm text-gray-500">
						Your operating system for the agentic era.
					</p>
				</div>

				<div>
					<h4 class="font-medium text-gray-900 mb-4">Product</h4>
					<ul class="space-y-3 text-sm text-gray-600">
						<li><a href="#features" class="hover:text-black transition-colors">Features</a></li>
						<li><a href="/login" class="hover:text-black transition-colors">Demo</a></li>
						<li><a href="https://github.com" target="_blank" rel="noopener" class="hover:text-black transition-colors">GitHub</a></li>
					</ul>
				</div>

				<div>
					<h4 class="font-medium text-gray-900 mb-4">Resources</h4>
					<ul class="space-y-3 text-sm text-gray-600">
						<li><a href="https://github.com" target="_blank" rel="noopener" class="hover:text-black transition-colors">Documentation</a></li>
						<li><a href="https://github.com" target="_blank" rel="noopener" class="hover:text-black transition-colors">API Reference</a></li>
					</ul>
				</div>

				<div>
					<h4 class="font-medium text-gray-900 mb-4">Ecosystem</h4>
					<ul class="space-y-3 text-sm text-gray-600">
						<li><a href="https://osa.dev" target="_blank" rel="noopener" class="hover:text-black transition-colors">OSA</a></li>
						<li><a href="https://amcp.ai" target="_blank" rel="noopener" class="hover:text-black transition-colors">AMCP</a></li>
					</ul>
				</div>
			</div>

			<div class="border-t border-gray-200 pt-8 flex flex-col md:flex-row items-center justify-between gap-4">
				<p class="text-sm text-gray-500">
					© 2025 Business OS · Open Source · MIT Licensed
				</p>
				<div class="flex items-center gap-4">
					<a href="https://github.com" target="_blank" rel="noopener" class="text-gray-400 hover:text-black transition-colors" aria-label="GitHub">
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
						</svg>
					</a>
					<a href="https://twitter.com" target="_blank" rel="noopener" class="text-gray-400 hover:text-black transition-colors" aria-label="Twitter">
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path d="M8.29 20.251c7.547 0 11.675-6.253 11.675-11.675 0-.178 0-.355-.012-.53A8.348 8.348 0 0022 5.92a8.19 8.19 0 01-2.357.646 4.118 4.118 0 001.804-2.27 8.224 8.224 0 01-2.605.996 4.107 4.107 0 00-6.993 3.743 11.65 11.65 0 01-8.457-4.287 4.106 4.106 0 001.27 5.477A4.072 4.072 0 012.8 9.713v.052a4.105 4.105 0 003.292 4.022 4.095 4.095 0 01-1.853.07 4.108 4.108 0 003.834 2.85A8.233 8.233 0 012 18.407a11.616 11.616 0 006.29 1.84" />
						</svg>
					</a>
				</div>
			</div>
		</div>
	</footer>
</div>
