<script lang="ts">
	import { useSession, signOut } from '$lib/auth-client';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { invalidateAll } from '$app/navigation';

	const session = useSession();

	// Profile state
	let isEditing = $state(false);
	let isSaving = $state(false);
	let saveMessage = $state('');
	let isUploadingPhoto = $state(false);
	let photoError = $state('');
	let profilePhoto = $state<string | null>(null);

	// Form fields
	let name = $state('');
	let email = $state('');
	let timezone = $state('America/New_York');
	let bio = $state('');

	// File input ref
	let fileInput: HTMLInputElement | undefined = $state(undefined);

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
			profilePhoto = $session.data.user.image || null;
		}
	});

	async function handlePhotoUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		// Validate file type
		const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
		if (!allowedTypes.includes(file.type)) {
			photoError = 'Please select a valid image file (JPEG, PNG, GIF, or WebP)';
			return;
		}

		// Validate file size (5MB max)
		if (file.size > 5 * 1024 * 1024) {
			photoError = 'Image must be less than 5MB';
			return;
		}

		isUploadingPhoto = true;
		photoError = '';

		try {
			const result = await api.uploadProfilePhoto(file);
			profilePhoto = result.url;
			// Refresh session to get updated user data
			await invalidateAll();
			saveMessage = 'Photo uploaded!';
			setTimeout(() => saveMessage = '', 2000);
		} catch (error) {
			console.error('Failed to upload photo:', error);
			photoError = error instanceof Error ? error.message : 'Failed to upload photo';
		} finally {
			isUploadingPhoto = false;
			// Clear the input
			if (fileInput) fileInput.value = '';
		}
	}

	async function handleDeletePhoto() {
		if (!profilePhoto) return;

		isUploadingPhoto = true;
		photoError = '';

		try {
			await api.deleteProfilePhoto();
			profilePhoto = null;
			await invalidateAll();
			saveMessage = 'Photo removed!';
			setTimeout(() => saveMessage = '', 2000);
		} catch (error) {
			console.error('Failed to delete photo:', error);
			photoError = error instanceof Error ? error.message : 'Failed to delete photo';
		} finally {
			isUploadingPhoto = false;
		}
	}

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
				memberSince: new Date().toISOString() // TODO: Get from API when available
			};
		} catch (error) {
			console.error('Failed to load profile stats:', error);
		}
	}

	async function handleSave() {
		isSaving = true;
		saveMessage = '';

		try {
			await api.updateProfile({ name });
			await invalidateAll();
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

<div class="h-full flex flex-col bg-white dark:bg-gray-900">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-gray-900 dark:text-white">Profile</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage your account information</p>
		</div>
		{#if !isEditing}
			<button
				onclick={() => isEditing = true}
				class="btn-pill btn-pill-secondary btn-pill-sm"
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
						<div class="relative">
							{#if profilePhoto}
								<img
									src={profilePhoto.startsWith('/') ? `http://localhost:8001${profilePhoto}` : profilePhoto}
									alt={name || 'Profile'}
									class="w-24 h-24 rounded-full object-cover border-4 border-gray-200 dark:border-gray-600"
								/>
							{:else}
								<div class="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-white flex items-center justify-center text-3xl font-semibold">
									{name?.charAt(0).toUpperCase() || 'U'}
								</div>
							{/if}
							{#if isUploadingPhoto}
								<div class="absolute inset-0 bg-black/50 rounded-full flex items-center justify-center">
									<svg class="w-6 h-6 text-white animate-spin" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
								</div>
							{/if}
						</div>
						{#if isEditing}
							<input
								bind:this={fileInput}
								type="file"
								accept="image/jpeg,image/png,image/gif,image/webp"
								class="hidden"
								onchange={handlePhotoUpload}
							/>
							<div class="mt-2 flex flex-col gap-1">
								<button
									onclick={() => fileInput?.click()}
									disabled={isUploadingPhoto}
									class="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 disabled:opacity-50"
								>
									{profilePhoto ? 'Change photo' : 'Upload photo'}
								</button>
								{#if profilePhoto}
									<button
										onclick={handleDeletePhoto}
										disabled={isUploadingPhoto}
										class="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 disabled:opacity-50"
									>
										Remove photo
									</button>
								{/if}
							</div>
							{#if photoError}
								<p class="mt-1 text-xs text-red-500">{photoError}</p>
							{/if}
						{/if}
					</div>

					<!-- Info -->
					<div class="flex-1">
						{#if isEditing}
							<div class="space-y-4">
								<div>
									<label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
									<input
										id="name"
										type="text"
										bind:value={name}
										class="input"
										placeholder="Your name"
									/>
								</div>
								<div>
									<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
									<input
										id="email"
										type="email"
										bind:value={email}
										class="input"
										placeholder="your@email.com"
										disabled
									/>
									<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">Email cannot be changed</p>
								</div>
								<div>
									<label for="bio" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bio</label>
									<textarea
										id="bio"
										bind:value={bio}
										class="input resize-none"
										rows="3"
										placeholder="Tell us about yourself..."
									></textarea>
								</div>
								<div>
									<label for="timezone" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timezone</label>
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
										class="btn-pill btn-pill-secondary btn-pill-sm"
									>
										Cancel
									</button>
									<button
										onclick={handleSave}
										disabled={isSaving}
										class="btn-pill btn-pill-primary btn-pill-sm"
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
								<h2 class="text-2xl font-semibold text-gray-900 dark:text-white">{name || 'No name set'}</h2>
								<p class="text-gray-500 dark:text-gray-400 mt-1">{email}</p>
								{#if bio}
									<p class="text-gray-600 dark:text-gray-300 mt-3">{bio}</p>
								{/if}
								<div class="flex items-center gap-4 mt-4 text-sm text-gray-500 dark:text-gray-400">
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
					<p class="text-3xl font-semibold text-gray-900 dark:text-white">{stats.totalProjects}</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Projects</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900 dark:text-white">{stats.activeTasks}</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Active Tasks</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900 dark:text-white">{stats.completedTasks}</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Completed</p>
				</div>
				<div class="card text-center">
					<p class="text-3xl font-semibold text-gray-900 dark:text-white">{stats.conversationCount}</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Conversations</p>
				</div>
			</div>

			<!-- Activity Section -->
			<div class="card">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Recent Activity</h3>
				<div class="space-y-4">
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
							<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900 dark:text-white">Started a new conversation</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">Just now</p>
						</div>
					</div>
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
							<svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900 dark:text-white">Completed a task</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">2 hours ago</p>
						</div>
					</div>
					<div class="flex items-center gap-3 text-sm">
						<div class="w-8 h-8 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center">
							<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
						</div>
						<div class="flex-1">
							<p class="text-gray-900 dark:text-white">Created a new project</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">Yesterday</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Quick Links -->
			<div class="card">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Quick Links</h3>
				<div class="grid grid-cols-2 gap-3">
					<a href="/settings" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
						<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						<span class="text-sm text-gray-700 dark:text-gray-300">Account Settings</span>
					</a>
					<a href="/settings" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
						<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<span class="text-sm text-gray-700 dark:text-gray-300">Notifications</span>
					</a>
					<a href="/daily" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
						<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						<span class="text-sm text-gray-700 dark:text-gray-300">Daily Log</span>
					</a>
					<a href="/chat" class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
						<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
						<span class="text-sm text-gray-700 dark:text-gray-300">Chat History</span>
					</a>
				</div>
			</div>
		</div>
	</div>
</div>
