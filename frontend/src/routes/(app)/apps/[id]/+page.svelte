<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { Loader2, CheckCircle, XCircle, ArrowLeft, Sparkles, Code, Database, Layout, TestTube } from 'lucide-svelte';

	const appId = $derived($page.params.id);

	// App state
	let appStatus = $state<'loading' | 'generating' | 'running' | 'error' | 'not_found'>('loading');
	let appData = $state<{
		name?: string;
		description?: string;
		status?: string;
		deployment_url?: string;
		error_message?: string;
	} | null>(null);

	// Generation progress
	let progress = $state(0);
	let currentPhase = $state('');
	let statusMessage = $state('');
	let eventSource: EventSource | null = null;

	// Generation phases for display
	const phases = [
		{ id: 'planning', label: 'Planning Architecture', icon: Layout },
		{ id: 'generation', label: 'Generating Code', icon: Code },
		{ id: 'database', label: 'Setting Up Database', icon: Database },
		{ id: 'testing', label: 'Running Tests', icon: TestTube },
		{ id: 'deployment', label: 'Deploying App', icon: Sparkles }
	];

	function getCurrentPhaseIndex(phase: string): number {
		const index = phases.findIndex(p => p.id === phase);
		return index >= 0 ? index : 0;
	}

	async function fetchAppStatus() {
		try {
			const response = await fetch(`/api/osa/module-instances/queue/${appId}/status`, {
				credentials: 'include'
			});

			if (!response.ok) {
				if (response.status === 404) {
					appStatus = 'not_found';
					return;
				}
				throw new Error('Failed to fetch app status');
			}

			const data = await response.json();
			appData = data;

			if (data.status === 'pending' || data.status === 'processing' || data.status === 'generating') {
				appStatus = 'generating';
				progress = (data.progress || 0) * 100;
				currentPhase = data.current_step || 'planning';
				statusMessage = data.message || 'Starting generation...';
				// Subscribe to SSE for real-time updates
				subscribeToProgress();
			} else if (data.status === 'completed' || data.status === 'running') {
				appStatus = 'running';
				progress = 100;
			} else if (data.status === 'failed' || data.status === 'error') {
				appStatus = 'error';
			}
		} catch (error) {
			console.error('Failed to fetch app status:', error);
			appStatus = 'error';
		}
	}

	function subscribeToProgress() {
		if (eventSource) {
			eventSource.close();
		}

		eventSource = new EventSource(`/api/osa/module-instances/generate/${appId}/stream`);

		eventSource.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);

				if (data.event_type === 'build_progress') {
					progress = data.progress_percent || progress;
					currentPhase = data.phase || currentPhase;
					statusMessage = data.status_message || statusMessage;
				} else if (data.event_type === 'build_completed') {
					appStatus = 'running';
					progress = 100;
					appData = { ...appData, ...data.data };
					eventSource?.close();
				} else if (data.event_type === 'build_error') {
					appStatus = 'error';
					appData = { ...appData, error_message: data.error || data.status_message };
					eventSource?.close();
				} else if (data.event_type === 'heartbeat') {
					// Keep-alive, ignore
				}
			} catch (e) {
				console.error('Failed to parse SSE event:', e);
			}
		};

		eventSource.onerror = () => {
			console.error('SSE connection error');
			// Try to reconnect after 5 seconds
			setTimeout(() => {
				if (appStatus === 'generating') {
					subscribeToProgress();
				}
			}, 5000);
		};
	}

	onMount(() => {
		fetchAppStatus();
	});

	onDestroy(() => {
		if (eventSource) {
			eventSource.close();
		}
	});
</script>

<svelte:head>
	<title>{appData?.name || 'App Details'} | Business OS</title>
</svelte:head>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<header class="flex items-center gap-4 px-6 py-4 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
		<button
			onclick={() => goto('/apps')}
			class="p-2 -ml-2 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
		>
			<ArrowLeft class="w-5 h-5" />
		</button>
		<div class="flex-1">
			<h1 class="text-lg font-semibold text-gray-900 dark:text-white">
				{appData?.name || 'App Details'}
			</h1>
			{#if appData?.description}
				<p class="text-sm text-gray-500 dark:text-gray-400 truncate">{appData.description}</p>
			{/if}
		</div>
	</header>

	<!-- Content -->
	<main class="flex-1 flex flex-col items-center justify-center p-8">
		{#if appStatus === 'loading'}
			<!-- Loading State -->
			<div class="flex flex-col items-center">
				<Loader2 class="w-10 h-10 text-gray-400 animate-spin mb-4" />
				<p class="text-gray-600 dark:text-gray-400">Loading app details...</p>
			</div>

		{:else if appStatus === 'generating'}
			<!-- Generation Progress -->
			<div class="max-w-lg w-full">
				<div class="text-center mb-8">
					<div class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
						<Sparkles class="w-10 h-10 text-white animate-pulse" />
					</div>
					<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
						Building Your App
					</h2>
					<p class="text-gray-600 dark:text-gray-400">
						{statusMessage || 'AI is generating your application...'}
					</p>
				</div>

				<!-- Progress Bar -->
				<div class="mb-8">
					<div class="flex justify-between text-sm mb-2">
						<span class="text-gray-600 dark:text-gray-400">Progress</span>
						<span class="font-medium text-gray-900 dark:text-white">{Math.round(progress)}%</span>
					</div>
					<div class="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
						<div
							class="h-full bg-gradient-to-r from-blue-500 to-purple-600 transition-all duration-500 ease-out"
							style="width: {progress}%"
						></div>
					</div>
				</div>

				<!-- Phase Indicators -->
				<div class="space-y-3">
					{#each phases as phase, i}
						{@const phaseIndex = getCurrentPhaseIndex(currentPhase)}
						{@const isActive = phase.id === currentPhase}
						{@const isComplete = i < phaseIndex}
						{@const isPending = i > phaseIndex}

						<div class="flex items-center gap-3 p-3 rounded-xl transition-colors
							{isActive ? 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800' :
							 isComplete ? 'bg-green-50 dark:bg-green-900/10' : 'bg-gray-50 dark:bg-gray-800/50'}">
							<div class="w-8 h-8 rounded-lg flex items-center justify-center
								{isActive ? 'bg-blue-500 text-white' :
								 isComplete ? 'bg-green-500 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-400'}">
								{#if isComplete}
									<CheckCircle class="w-4 h-4" />
								{:else if isActive}
									<Loader2 class="w-4 h-4 animate-spin" />
								{:else}
									<svelte:component this={phase.icon} class="w-4 h-4" />
								{/if}
							</div>
							<span class="text-sm font-medium
								{isActive ? 'text-blue-700 dark:text-blue-300' :
								 isComplete ? 'text-green-700 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}">
								{phase.label}
							</span>
						</div>
					{/each}
				</div>
			</div>

		{:else if appStatus === 'running'}
			<!-- App Running -->
			<div class="max-w-md text-center">
				<div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-green-400 to-emerald-500 flex items-center justify-center">
					<CheckCircle class="w-10 h-10 text-white" />
				</div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
					App Ready!
				</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					Your app has been generated and is ready to use.
				</p>

				{#if appData?.deployment_url}
					<a
						href={appData.deployment_url}
						target="_blank"
						rel="noopener noreferrer"
						class="inline-flex items-center gap-2 px-5 py-2.5 bg-gray-900 dark:bg-white text-white dark:text-gray-900
							rounded-xl font-medium text-sm transition-all duration-150
							hover:bg-gray-800 dark:hover:bg-gray-100"
					>
						Open App
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>
				{:else}
					<div class="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-800 rounded-lg text-sm text-gray-600 dark:text-gray-400">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Deployment URL will be available soon
					</div>
				{/if}
			</div>

		{:else if appStatus === 'error'}
			<!-- Error State -->
			<div class="max-w-md text-center">
				<div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-red-400 to-red-500 flex items-center justify-center">
					<XCircle class="w-10 h-10 text-white" />
				</div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
					Generation Failed
				</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-4">
					{appData?.error_message || 'Something went wrong while generating your app.'}
				</p>
				<button
					onclick={() => goto('/apps')}
					class="inline-flex items-center gap-2 px-5 py-2.5 bg-gray-900 dark:bg-white text-white dark:text-gray-900
						rounded-xl font-medium text-sm transition-all duration-150
						hover:bg-gray-800 dark:hover:bg-gray-100"
				>
					<ArrowLeft class="w-4 h-4" />
					Back to Apps
				</button>
			</div>

		{:else}
			<!-- Not Found -->
			<div class="max-w-md text-center">
				<div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
					<svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
					App Not Found
				</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					The app you're looking for doesn't exist or has been deleted.
				</p>
				<button
					onclick={() => goto('/apps')}
					class="inline-flex items-center gap-2 px-5 py-2.5 bg-gray-900 dark:bg-white text-white dark:text-gray-900
						rounded-xl font-medium text-sm transition-all duration-150
						hover:bg-gray-800 dark:hover:bg-gray-100"
				>
					<ArrowLeft class="w-4 h-4" />
					Back to Apps
				</button>
			</div>
		{/if}
	</main>
</div>
