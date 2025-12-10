<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	let { children } = $props();
</script>

<div class="min-h-screen flex">
	<!-- Left Panel - Branding -->
	<div class="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-12 flex-col justify-between relative overflow-hidden">
		<!-- Background Pattern -->
		<div class="absolute inset-0 opacity-10">
			<div class="absolute inset-0" style="background-image: url(&quot;data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.4'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E&quot;);"></div>
		</div>

		<!-- Floating Elements Animation -->
		<div class="absolute top-20 right-20 w-64 h-64 bg-white/5 rounded-full blur-3xl animate-pulse"></div>
		<div class="absolute bottom-40 left-10 w-48 h-48 bg-white/5 rounded-full blur-3xl animate-pulse" style="animation-delay: 1s;"></div>

		<!-- Logo & Tagline -->
		<div class="relative z-10" in:fly={{ y: -20, duration: 500 }}>
			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 bg-white rounded-xl flex items-center justify-center">
					<span class="text-gray-900 text-lg font-bold">B</span>
				</div>
				<span class="text-white text-xl font-semibold">Business OS</span>
			</div>
			<p class="text-gray-400 text-lg max-w-sm">
				Your business, one system.
			</p>
		</div>

		<!-- App Preview -->
		<div class="relative z-10 flex-1 flex items-center justify-center py-12" in:fly={{ y: 30, duration: 600, delay: 200 }}>
			<div class="w-full max-w-md bg-gray-800/50 backdrop-blur rounded-2xl border border-gray-700/50 p-4 shadow-2xl">
				<!-- Window Controls -->
				<div class="flex items-center gap-2 mb-4">
					<div class="w-3 h-3 rounded-full bg-red-500/80"></div>
					<div class="w-3 h-3 rounded-full bg-yellow-500/80"></div>
					<div class="w-3 h-3 rounded-full bg-green-500/80"></div>
				</div>
				<!-- Sidebar + Content Preview -->
				<div class="flex gap-3">
					<div class="w-1/3 space-y-2">
						<div class="h-8 bg-gray-700/50 rounded-lg"></div>
						<div class="h-6 bg-white/10 rounded-lg"></div>
						<div class="h-6 bg-gray-700/30 rounded-lg"></div>
						<div class="h-6 bg-gray-700/30 rounded-lg"></div>
						<div class="h-6 bg-gray-700/30 rounded-lg"></div>
					</div>
					<div class="flex-1 space-y-3">
						<div class="h-4 w-24 bg-gray-600/50 rounded"></div>
						<div class="h-20 bg-gray-700/30 rounded-xl"></div>
						<div class="h-16 bg-gray-700/30 rounded-xl"></div>
					</div>
				</div>
			</div>
		</div>

		<!-- Testimonial -->
		<div class="relative z-10" in:fly={{ y: 20, duration: 500, delay: 400 }}>
			<blockquote class="text-gray-300 text-sm italic mb-3">
				"Finally, a system I actually own. No more switching between tools."
			</blockquote>
			<div class="flex items-center gap-3">
				<div class="w-8 h-8 bg-gray-700 rounded-full"></div>
				<div>
					<p class="text-white text-sm font-medium">Founder</p>
					<p class="text-gray-500 text-xs">Your Company</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Right Panel - Form -->
	<div class="flex-1 flex items-center justify-center p-6 sm:p-12 bg-white">
		<div class="w-full max-w-md" in:fade={{ duration: 300 }}>
			<!-- Mobile Logo -->
			<div class="lg:hidden flex items-center gap-2 mb-8 justify-center">
				<div class="w-8 h-8 bg-gray-900 rounded-lg flex items-center justify-center">
					<span class="text-white text-sm font-bold">B</span>
				</div>
				<span class="text-gray-900 font-semibold">Business OS</span>
			</div>

			{@render children()}
		</div>
	</div>
</div>
