<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	// State
	let slackConnected = $state(false);
	let slackChannels = $state<any[]>([]);
	let selectedChannel = $state<any | null>(null);
	let isLoading = $state(true);

	// Check Slack connection
	async function checkSlackConnection() {
		try {
			const status = await api.getSlackConnectionStatus();
			slackConnected = status.connected;
			if (slackConnected) {
				await loadChannels();
			}
		} catch (e) {
			console.error('Failed to check Slack connection:', e);
		} finally {
			isLoading = false;
		}
	}

	async function loadChannels() {
		try {
			const response = await api.getSlackChannels();
			slackChannels = response.channels ?? [];
		} catch (e) {
			console.error('Failed to load Slack channels:', e);
		}
	}

	async function handleConnectSlack() {
		try {
			const result = await api.initiateSlackAuth();
			if (result.auth_url) {
				window.location.href = result.auth_url;
			}
		} catch (e) {
			console.error('Failed to initiate Slack auth:', e);
		}
	}

	onMount(() => {
		checkSlackConnection();
	});

	// Mock data for demonstration
	const mockChannels = [
		{ id: '1', name: 'general', is_private: false, member_count: 42, last_message: 'Welcome to the team!', last_message_time: '10:30 AM' },
		{ id: '2', name: 'engineering', is_private: false, member_count: 15, last_message: 'PR merged!', last_message_time: '9:45 AM' },
		{ id: '3', name: 'design', is_private: false, member_count: 8, last_message: 'New mockups ready', last_message_time: 'Yesterday' },
		{ id: '4', name: 'product-updates', is_private: false, member_count: 35, last_message: 'v2.0 release notes', last_message_time: 'Yesterday' },
		{ id: '5', name: 'team-leads', is_private: true, member_count: 5, last_message: 'Meeting at 2pm', last_message_time: 'Monday' },
	];

	const mockMessages = [
		{ id: '1', sender: 'John Doe', avatar: '', content: 'Hey team, the new feature is live!', time: '10:30 AM', reactions: ['🎉', '👍'] },
		{ id: '2', sender: 'Jane Smith', avatar: '', content: 'Great work everyone!', time: '10:32 AM', reactions: [] },
		{ id: '3', sender: 'Bob Johnson', avatar: '', content: 'Let me know if you encounter any issues.', time: '10:35 AM', reactions: ['👀'] },
	];
</script>

<div class="h-full flex">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<svg class="w-8 h-8 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
		</div>
	{:else if !slackConnected}
		<!-- Slack Not Connected -->
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center max-w-md">
				<div class="w-16 h-16 mx-auto mb-6 rounded-xl bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center">
					<svg class="w-10 h-10 text-purple-600 dark:text-purple-400" viewBox="0 0 24 24" fill="currentColor">
						<path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z"/>
					</svg>
				</div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Connect Slack</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					Connect your Slack workspace to view channels and messages directly from BusinessOS.
				</p>
				<button
					onclick={handleConnectSlack}
					class="inline-flex items-center gap-2 px-6 py-3 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors"
				>
					<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
						<path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313z"/>
					</svg>
					Connect Slack
				</button>

				<!-- Preview with mock data -->
				<div class="mt-8 pt-8 border-t border-gray-200 dark:border-gray-700">
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">Preview of what you'll see:</p>
					<div class="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 text-left">
						<div class="space-y-2">
							{#each mockChannels.slice(0, 3) as channel}
								<div class="flex items-center gap-3 p-2 rounded-lg bg-white dark:bg-gray-700">
									<span class="text-gray-500 dark:text-gray-400">#</span>
									<span class="font-medium text-gray-900 dark:text-white">{channel.name}</span>
									<span class="text-xs text-gray-400 dark:text-gray-500">{channel.member_count} members</span>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
		</div>
	{:else}
		<!-- Channel List -->
		<div class="w-64 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col">
			<div class="p-4 border-b border-gray-200 dark:border-gray-700">
				<h3 class="font-semibold text-gray-900 dark:text-white">Channels</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
					{slackChannels.length || mockChannels.length} channels
				</p>
			</div>

			<div class="flex-1 overflow-y-auto p-2 space-y-1">
				{#each (slackChannels.length > 0 ? slackChannels : mockChannels) as channel}
					<button
						onclick={() => selectedChannel = channel}
						class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors
							{selectedChannel?.id === channel.id
								? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300'
								: 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}"
					>
						<span class="text-gray-400 dark:text-gray-500">
							{channel.is_private ? '🔒' : '#'}
						</span>
						<span class="flex-1 text-left truncate">{channel.name}</span>
					</button>
				{/each}
			</div>
		</div>

		<!-- Channel Content -->
		<div class="flex-1 flex flex-col">
			{#if selectedChannel}
				<!-- Channel Header -->
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
					<div>
						<h2 class="font-semibold text-gray-900 dark:text-white flex items-center gap-2">
							<span class="text-gray-400 dark:text-gray-500">
								{selectedChannel.is_private ? '🔒' : '#'}
							</span>
							{selectedChannel.name}
						</h2>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{selectedChannel.member_count} members
						</p>
					</div>
					<div class="flex items-center gap-2">
						<button class="p-2 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
							</svg>
						</button>
						<button class="p-2 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
							</svg>
						</button>
					</div>
				</div>

				<!-- Messages -->
				<div class="flex-1 overflow-y-auto p-6 space-y-4">
					{#each mockMessages as message}
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-medium">
								{message.sender.charAt(0)}
							</div>
							<div class="flex-1">
								<div class="flex items-baseline gap-2">
									<span class="font-medium text-gray-900 dark:text-white">{message.sender}</span>
									<span class="text-xs text-gray-500 dark:text-gray-400">{message.time}</span>
								</div>
								<p class="text-gray-700 dark:text-gray-300 mt-1">{message.content}</p>
								{#if message.reactions.length > 0}
									<div class="flex items-center gap-1 mt-2">
										{#each message.reactions as reaction}
											<span class="px-2 py-0.5 bg-gray-100 dark:bg-gray-700 rounded-full text-sm">
												{reaction}
											</span>
										{/each}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>

				<!-- Message Input -->
				<div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
					<div class="flex items-center gap-3">
						<input
							type="text"
							placeholder="Message #{selectedChannel.name}"
							class="flex-1 px-4 py-2 bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500"
						/>
						<button class="p-2 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
							</svg>
						</button>
						<button class="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors">
							Send
						</button>
					</div>
				</div>
			{:else}
				<!-- No Channel Selected -->
				<div class="flex-1 flex items-center justify-center text-gray-500 dark:text-gray-400">
					<div class="text-center">
						<svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
						</svg>
						<p>Select a channel to view messages</p>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
