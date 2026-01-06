<script lang="ts">
	import { page } from '$app/stores';

	let { children } = $props();

	const tabs = [
		{ href: '/communication/calendar', label: 'Calendar', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ href: '/communication/email', label: 'Email', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ href: '/communication/channels', label: 'Channels', icon: 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' },
	];

	const isActiveTab = (tabHref: string) => {
		return $page.url.pathname.startsWith(tabHref);
	};
</script>

<div class="h-full flex flex-col">
	<!-- Header with Tabs -->
	<div class="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
		<div class="px-6 pt-4">
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">Communication Hub</h1>
			<nav class="flex gap-1">
				{#each tabs as tab}
					<a
						href={tab.href}
						class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium rounded-t-lg border-b-2 transition-colors
							{isActiveTab(tab.href)
								? 'border-blue-600 text-blue-600 dark:border-blue-400 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
								: 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={tab.icon} />
						</svg>
						{tab.label}
					</a>
				{/each}
			</nav>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-hidden">
		{@render children()}
	</div>
</div>
