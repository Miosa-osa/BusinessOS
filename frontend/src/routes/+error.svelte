<script>
	import { page } from '$app/stores';
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
	<div class="text-center p-8">
		<h1 class="text-6xl font-bold text-gray-300 dark:text-gray-600">{$page.status}</h1>
		<p class="mt-4 text-xl text-gray-600 dark:text-gray-400">{$page.error?.message || 'Something went wrong'}</p>
		<a href="/" class="mt-6 inline-block px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
			Go Home
		</a>
	</div>
</div>
