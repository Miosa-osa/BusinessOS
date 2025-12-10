<script lang="ts">
	import { useSession } from '$lib/auth-client';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';

	const session = useSession();

	// Profile state
	let isEditing = $state(false);
	let isSaving = $state(false);
	let saveMessage = $state('');

	// Form fields
	let name = $state('');
	let email = $state('');
	let timezone = $state('America/New_York');
	let bio = $state('');

	// Stats (to be loaded from API)
	let stats = $state({
		totalProjects: 0,
		activeTasks: 0,
		completedTasks: 0,
		conversationCount: 0,
		memberSince: ''
	});

	$effect(() => {
		if ($session.data?.user) {
			name = $session.data.user.name || '';
			email = $session.data.user.email || '';
		}
	});

	onMount(async () => {
		await loadProfileStats();
	});

	async function loadProfileStats() {
		try {
			// Load stats from various endpoints
			const [projects, dashboard] = await Promise.all([
				api.getProjects().catch(() => []),
				api.getDashboardSummary().catch(() => null)
			]);

			stats = {
				totalProjects: projects.length,
				activeTasks: dashboard?.tasks.filter((t: any) => !t.completed).length || 0,
				completedTasks: dashboard?.tasks.filter((t: any) => t.completed).length || 0,
				conversationCount: 0, // Would need conversations endpoint
				memberSince: $session.data?.user?.createdAt || new Date().toISOString()
			};
		} catch (error) {
			console.error('Failed to load profile stats:', error);
		}
	}

	async function handleSave() {
		isSaving = true;
		saveMessage = '';

		try {
			// TODO: Implement profile update API
			await new Promise(resolve => setTimeout(resolve, 500)); // Simulated delay
			saveMessage = 'Profile updated!';
			isEditing = false;
			setTimeout(() => saveMessage = '', 2000);
		} catch (error) {
			console.error('Failed to save profile:', error);
			saveMessage = 'Error saving profile';
		} finally {
			isSaving = false;
		}
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-gray-900">Profile</h1>
			<p class="text-sm text-gray-500 mt-0.5">Manage your account information</p>
		</div>
		{#if !isEditing}
			<button
				onclick={() => isEditing = true}
				class="btn btn-secondary"
			>
				Edit Profile
			</button>
		{/if}
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto">
		<div class="max-w-4xl mx-auto p-6 space-y-6">
			<!-- Profile Card -->
			<div class="card">
				<div class="flex items-start gap-6">
					<!-- Avatar -->
					<div class="flex-shrink-0">
						<div class="w-24 h-24 rounded-full bg-gray-900 text-white flex items-center justify-center text-3xl font-semibold">
							{name?.charAt(0).toUpperCase() || 'U'}
						</div>
						{#if isEditing}
							<button class="mt-2 text-sm text-blue-600 hover:text-blue-700 w-full text-center">
								Change photo
							</button>
						{/if}
					</div>

					<!-- Info -->
					<div class="flex-1">
						{#if isEditing}
							<div class="space-y-4">
								<div>
									<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
									<input
										id="name"
										type="text"
										bind:value={name}
										class="input"
										placeholder="Your name"
									/>
								</div>
								<div>
									<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
									<input
										id="email"
										type="email"
										bind:value={email}
										class="input"
										placeholder="your@email.com"
										disabled
									/>
									<p class="text-xs text-gray-400 mt-1">Email cannot be changed</p>
								</div>
								<div>
									<label for="bio" class="block text-sm font-medium text-gray-700 mb-1">Bio</label>
									<textarea
										id="bio"
										bind:value={bio}
										class="input resize-none"
										rows="3"
										placeholder="Tell us about yourself..."
									></textarea>
								</div>
								<div>
									<label for="timezone" class="block text-sm font-medium text-gray-700 mb-1">Timezone</label>
									<select id="timezone" bind:value={timezone} class="input">
										<option value="America/New_York">Eastern Time (ET)</option>
										<option value="America/Chicago">Central Time (CT)</option>
										<option value="America/Denver">Mountain Time (MT)</option>
										<option value="America/Los_Angeles">Pacific Time (PT)</option>
										<option value="Europe/London">London (GMT)</option>
										<option value="Europe/Paris">Paris (CET)</option>
										<option value="Asia/Tokyo">Tokyo (JST)</option>
									</select>
								</div>

								<div class="flex gap-3 pt-2">
									<button
										onclick={() => isEditing = false}
										class="btn btn-secondary"
									>
										Cancel
									</button>
									<button
										onclick={handleSave}
										disabled={isSaving}
										class="btn btn-primary"
									>
										{#if isSaving}
											Saving...
										{:else}
											Save Changes
										{/if}
									</button>
									{#if saveMessage}
										<span class="text-sm text-green-600 self-center">{saveMessage}</span>
									{/if}
								</div>
							</div>
						{:else}
							<div>
								<h2 class="text-2xl font-semibold text-gray-900">{name || 'No name set'}</h2>
								<p class="text-gray-500 mt-1">{email}</p>
								{#if bio}
									<p class="text-gray-600 mt-3">{bio}</p>
								{/if}
								<div class="flex items-center gap-4 mt-4 text-sm text-gray-500">
									<span class="flex items-center gap-1.5">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
										Member since {formatDate(stats.memberSince)}
									</span>
								</div>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Stats Grid -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900">{stats.totalProjects}</p>
					<p class="text-sm text-gray-500 mt-1">Projects</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900">{stats.activeTasks}</p>
					<p class="text-sm text-gray-500 mt-1">Active Tasks</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900">{stats.completedTasks}</p>
					<p class="text-sm text-gray-500 mt-1">Completed</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900">{stats.conversationCount}</p>
					<p class="text-sm text-gray-500 mt-1">Conversations</p>
				</div>
			</div>

			<!-- Activity Section -->
			<div class="card">
				<h3 class="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
				<div class="space-y-4">
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
							<svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900">Started a new conversation</p>
							<p class="text-xs text-gray-500">Just now</p>
						</div>
					</div>
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-green-100 flex items-center justify-center">
							<svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900">Completed a task</p>
							<p class="text-xs text-gray-500">2 hours ago</p>
						</div>
					</div>
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-purple-100 flex items-center justify-center">
							<svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900">Created a new project</p>
							<p class="text-xs text-gray-500">Yesterday</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Quick Links -->
			<div class="card">
				<h3 class="text-lg font-medium text-gray-900 mb-4">Quick Links</h3>
				<div class="grid grid-cols-2 gap-3">
					<a href="/settings" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						<span class="text-sm text-gray-700">Account Settings</span>
					</a>
					<a href="/settings" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<span class="text-sm text-gray-700">Notifications</span>
					</a>
					<a href="/daily" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						<span class="text-sm text-gray-700">Daily Log</span>
					</a>
					<a href="/chat" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
						<span class="text-sm text-gray-700">Chat History</span>
					</a>
				</div>
			</div>
		</div>
	</div>
</div>
