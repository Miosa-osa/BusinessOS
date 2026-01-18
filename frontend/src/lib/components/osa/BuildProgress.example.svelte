<script lang="ts">
	/**
	 * Example usage of BuildProgress component
	 * This file demonstrates how to integrate the BuildProgress component
	 * into your application for real-time build tracking.
	 */

	import BuildProgress from './BuildProgress.svelte';
	import type { BuildResult } from './BuildProgress.svelte';
	import { ExternalLink, Play, RotateCcw } from 'lucide-svelte';

	// State
	let buildId = $state('build-example-12345');
	let showProgress = $state(true);
	let lastResult = $state<BuildResult | null>(null);
	let lastError = $state<string | null>(null);

	// Handle build completion
	function handleBuildComplete(result: BuildResult) {
		console.log('Build completed successfully!', result);
		lastResult = result;
		lastError = null;

		// Optional: Hide progress after 5 seconds
		setTimeout(() => {
			showProgress = false;
		}, 5000);
	}

	// Handle build error
	function handleBuildError(error: Error) {
		console.error('Build failed:', error.message);
		lastError = error.message;
		lastResult = null;
	}

	// Simulate starting a new build
	async function startNewBuild() {
		try {
			// Reset state
			lastResult = null;
			lastError = null;
			showProgress = true;

			// Call your API to start a build
			const response = await fetch('/api/osa/build', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					appName: 'My Test App',
					description: 'Test application'
				})
			});

			const data = await response.json();
			buildId = data.buildId;
		} catch (error) {
			console.error('Failed to start build:', error);
			lastError = 'Failed to start build';
		}
	}

	// Reset and try again
	function resetDemo() {
		buildId = `build-demo-${Date.now()}`;
		lastResult = null;
		lastError = null;
		showProgress = true;
	}
</script>

<div class="min-h-screen bg-gray-950 text-white">
	<div class="max-w-4xl mx-auto px-6 py-12">
		<!-- Header -->
		<div class="text-center mb-12">
			<h1 class="text-3xl font-bold mb-3">BuildProgress Component</h1>
			<p class="text-gray-400 text-lg">Real-time build progress tracking with SSE streaming</p>
		</div>

		<!-- Controls -->
		<div class="flex items-center justify-center gap-4 mb-8">
			<button
				class="btn-pill btn-pill-primary flex items-center gap-2"
				onclick={startNewBuild}
				disabled={showProgress}
			>
				<Play class="w-4 h-4" />
				Start New Build
			</button>

			<button
				class="btn-pill btn-pill-secondary flex items-center gap-2"
				onclick={resetDemo}
			>
				<RotateCcw class="w-4 h-4" />
				Reset Demo
			</button>
		</div>

		<!-- Result Badges -->
		{#if lastResult}
			<div
				class="flex items-center justify-center gap-3 mb-8 p-4 bg-green-500/10 border border-green-500/30 rounded-lg"
			>
				<span class="text-green-400 font-semibold">Build Completed Successfully!</span>
				{#if lastResult.deploymentUrl}
					<a
						href={lastResult.deploymentUrl}
						target="_blank"
						rel="noopener noreferrer"
						class="flex items-center gap-1.5 text-green-400 hover:text-green-300 underline"
					>
						<ExternalLink class="w-4 h-4" />
						Open Deployed App
					</a>
				{/if}
				{#if lastResult.duration}
					<span class="text-gray-400 text-sm">({lastResult.duration}s)</span>
				{/if}
			</div>
		{/if}

		{#if lastError && !showProgress}
			<div
				class="flex items-center justify-center gap-3 mb-8 p-4 bg-red-500/10 border border-red-500/30 rounded-lg"
			>
				<span class="text-red-400 font-semibold">Build Failed:</span>
				<span class="text-gray-300">{lastError}</span>
			</div>
		{/if}

		<!-- Build Progress Component -->
		{#if showProgress}
			<div class="mb-12">
				<BuildProgress {buildId} onComplete={handleBuildComplete} onError={handleBuildError} />
			</div>
		{:else}
			<div
				class="flex flex-col items-center justify-center py-16 px-8 bg-gray-900 border-2 border-dashed border-gray-700 rounded-xl text-center mb-12"
			>
				<svg
					class="w-16 h-16 text-gray-600 mb-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23-.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5"
					/>
				</svg>
				<h3 class="text-xl font-semibold text-gray-200 mb-2">No Active Build</h3>
				<p class="text-gray-500">Click "Start New Build" or "Reset Demo" to begin</p>
			</div>
		{/if}

		<!-- Documentation -->
		<div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
			<h2 class="text-xl font-semibold mb-4">How It Works</h2>

			<ol class="list-decimal list-inside space-y-2 text-gray-400 mb-6">
				<li>
					Start a build by calling your backend API (e.g., <code
						class="bg-gray-800 px-1.5 py-0.5 rounded text-pink-400">/api/osa/build</code
					>)
				</li>
				<li>
					Pass the returned <code class="bg-gray-800 px-1.5 py-0.5 rounded text-pink-400"
						>buildId</code
					> to the BuildProgress component
				</li>
				<li>
					The component connects to <code class="bg-gray-800 px-1.5 py-0.5 rounded text-pink-400"
						>/api/osa/builds/{'{buildId}'}/stream</code
					>
				</li>
				<li>Real-time updates are displayed as Server-Sent Events arrive</li>
				<li>
					The <code class="bg-gray-800 px-1.5 py-0.5 rounded text-pink-400">onComplete</code> callback
					fires with the full result when done
				</li>
				<li>
					The <code class="bg-gray-800 px-1.5 py-0.5 rounded text-pink-400">onError</code> callback fires
					if the build fails
				</li>
			</ol>

			<h3 class="text-lg font-semibold mb-3">Expected SSE Event Format</h3>
			<pre class="bg-gray-950 rounded-lg p-4 overflow-x-auto text-sm"><code class="text-gray-300"
					>{JSON.stringify(
						{
							progress: 50,
							phase: 'Building',
							log: 'Generating components...',
							status: 'in_progress',
							estimatedTimeRemaining: 45
						},
						null,
						2
					)}</code
				></pre>

			<h3 class="text-lg font-semibold mt-6 mb-3">Component Props</h3>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="text-left text-gray-400 border-b border-gray-800">
							<th class="pb-2">Prop</th>
							<th class="pb-2">Type</th>
							<th class="pb-2">Required</th>
							<th class="pb-2">Description</th>
						</tr>
					</thead>
					<tbody class="text-gray-300">
						<tr class="border-b border-gray-800/50">
							<td class="py-2 font-mono text-pink-400">buildId</td>
							<td class="py-2">string</td>
							<td class="py-2">Yes</td>
							<td class="py-2">The unique build ID to track</td>
						</tr>
						<tr class="border-b border-gray-800/50">
							<td class="py-2 font-mono text-pink-400">onComplete</td>
							<td class="py-2">(result: BuildResult) =&gt; void</td>
							<td class="py-2">No</td>
							<td class="py-2">Callback when build completes successfully</td>
						</tr>
						<tr>
							<td class="py-2 font-mono text-pink-400">onError</td>
							<td class="py-2">(error: Error) =&gt; void</td>
							<td class="py-2">No</td>
							<td class="py-2">Callback when build fails</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>
