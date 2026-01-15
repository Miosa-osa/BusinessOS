<script lang="ts">
	import { userAppsStore, type CreateUserAppParams } from '$lib/stores/userAppsStore';
	import { desktop3dStore } from '$lib/stores/desktop3dStore';
	import { X, Plus, Search, Globe, Sparkles, ExternalLink, Trash2, AlertTriangle } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	const dispatch = createEventDispatcher();

	let { workspaceId, onClose } = $props<{
		workspaceId: string;
		onClose: () => void;
	}>();

	// Form state
	let name = $state('');
	let url = $state('');
	let category = $state('productivity');
	let description = $state('');
	let isSubmitting = $state(false);
	let searchQuery = $state('');
	let activeTab = $state<'browse' | 'custom'>('browse'); // Web apps only - no native

	// Helper to detect if app URL is a native bundle ID (legacy)
	function isNativeBundleId(url: string): boolean {
		return url.startsWith('com.') || url.startsWith('org.') || url.startsWith('io.');
	}

	// Check if any apps are native (for warning display)
	const hasNativeApps = $derived($userAppsStore.apps.some(app => isNativeBundleId(app.url)));

	// Popular web apps with logos (iframe-embedded)
	const popularApps = [
		// AI Tools
		{
			name: 'Claude',
			url: 'https://claude.ai',
			color: '#D97757',
			category: 'ai',
			description: 'AI assistant by Anthropic',
			logo: 'https://www.google.com/s2/favicons?domain=claude.ai&sz=128'
		},
		{
			name: 'ChatGPT',
			url: 'https://chat.openai.com',
			color: '#10A37F',
			category: 'ai',
			description: 'AI assistant by OpenAI',
			logo: 'https://www.google.com/s2/favicons?domain=chat.openai.com&sz=128'
		},
		{
			name: 'Perplexity',
			url: 'https://www.perplexity.ai',
			color: '#20808D',
			category: 'ai',
			description: 'AI search engine',
			logo: 'https://www.google.com/s2/favicons?domain=perplexity.ai&sz=128'
		},
		{
			name: 'Gemini',
			url: 'https://gemini.google.com',
			color: '#8E75B2',
			category: 'ai',
			description: 'AI assistant by Google',
			logo: 'https://www.google.com/s2/favicons?domain=gemini.google.com&sz=128'
		},
		// Productivity
		{
			name: 'Notion',
			url: 'https://notion.so',
			color: '#000000',
			category: 'productivity',
			description: 'All-in-one workspace',
			logo: 'https://www.google.com/s2/favicons?domain=notion.so&sz=128'
		},
		{
			name: 'Google Docs',
			url: 'https://docs.google.com',
			color: '#4285F4',
			category: 'productivity',
			description: 'Document editing',
			logo: 'https://www.google.com/s2/favicons?domain=docs.google.com&sz=128'
		},
		{
			name: 'Google Sheets',
			url: 'https://sheets.google.com',
			color: '#0F9D58',
			category: 'productivity',
			description: 'Spreadsheets',
			logo: 'https://www.google.com/s2/favicons?domain=sheets.google.com&sz=128'
		},
		{
			name: 'Airtable',
			url: 'https://airtable.com',
			color: '#18BFFF',
			category: 'productivity',
			description: 'Database & spreadsheet',
			logo: 'https://www.google.com/s2/favicons?domain=airtable.com&sz=128'
		},
		// Project Management
		{
			name: 'Linear',
			url: 'https://linear.app',
			color: '#5E6AD2',
			category: 'project-management',
			description: 'Issue tracking',
			logo: 'https://www.google.com/s2/favicons?domain=linear.app&sz=128'
		},
		{
			name: 'Asana',
			url: 'https://app.asana.com',
			color: '#F06A6A',
			category: 'project-management',
			description: 'Work management',
			logo: 'https://www.google.com/s2/favicons?domain=asana.com&sz=128'
		},
		{
			name: 'ClickUp',
			url: 'https://app.clickup.com',
			color: '#7B68EE',
			category: 'project-management',
			description: 'Project management',
			logo: 'https://www.google.com/s2/favicons?domain=clickup.com&sz=128'
		},
		{
			name: 'Trello',
			url: 'https://trello.com',
			color: '#0079BF',
			category: 'project-management',
			description: 'Kanban boards',
			logo: 'https://www.google.com/s2/favicons?domain=trello.com&sz=128'
		},
		{
			name: 'Monday',
			url: 'https://monday.com',
			color: '#FF3D57',
			category: 'project-management',
			description: 'Work OS',
			logo: 'https://www.google.com/s2/favicons?domain=monday.com&sz=128'
		},
		// Communication
		{
			name: 'Slack',
			url: 'https://app.slack.com',
			color: '#4A154B',
			category: 'communication',
			description: 'Team messaging',
			logo: 'https://www.google.com/s2/favicons?domain=slack.com&sz=128'
		},
		{
			name: 'Discord',
			url: 'https://discord.com/app',
			color: '#5865F2',
			category: 'communication',
			description: 'Voice & text chat',
			logo: 'https://www.google.com/s2/favicons?domain=discord.com&sz=128'
		},
		{
			name: 'Loom',
			url: 'https://www.loom.com',
			color: '#625DF5',
			category: 'communication',
			description: 'Video messages',
			logo: 'https://www.google.com/s2/favicons?domain=loom.com&sz=128'
		},
		// Design
		{
			name: 'Figma',
			url: 'https://www.figma.com',
			color: '#F24E1E',
			category: 'design',
			description: 'Design tool',
			logo: 'https://www.google.com/s2/favicons?domain=figma.com&sz=128'
		},
		{
			name: 'Canva',
			url: 'https://www.canva.com',
			color: '#00C4CC',
			category: 'design',
			description: 'Graphic design',
			logo: 'https://www.google.com/s2/favicons?domain=canva.com&sz=128'
		},
		{
			name: 'Miro',
			url: 'https://miro.com',
			color: '#FFD02F',
			category: 'collaboration',
			description: 'Whiteboard',
			logo: 'https://www.google.com/s2/favicons?domain=miro.com&sz=128'
		},
		// Development
		{
			name: 'GitHub',
			url: 'https://github.com',
			color: '#24292E',
			category: 'development',
			description: 'Code hosting',
			logo: 'https://www.google.com/s2/favicons?domain=github.com&sz=128'
		},
		{
			name: 'GitLab',
			url: 'https://gitlab.com',
			color: '#FC6D26',
			category: 'development',
			description: 'DevOps platform',
			logo: 'https://www.google.com/s2/favicons?domain=gitlab.com&sz=128'
		},
		{
			name: 'Vercel',
			url: 'https://vercel.com/dashboard',
			color: '#000000',
			category: 'development',
			description: 'Deploy & host',
			logo: 'https://www.google.com/s2/favicons?domain=vercel.com&sz=128'
		},
		// Storage
		{
			name: 'Google Drive',
			url: 'https://drive.google.com',
			color: '#4285F4',
			category: 'storage',
			description: 'Cloud storage',
			logo: 'https://www.google.com/s2/favicons?domain=drive.google.com&sz=128'
		},
		{
			name: 'Dropbox',
			url: 'https://www.dropbox.com',
			color: '#0061FF',
			category: 'storage',
			description: 'File storage',
			logo: 'https://www.google.com/s2/favicons?domain=dropbox.com&sz=128'
		},
		// Email & Calendar
		{
			name: 'Gmail',
			url: 'https://mail.google.com',
			color: '#EA4335',
			category: 'communication',
			description: 'Email by Google',
			logo: 'https://www.google.com/s2/favicons?domain=mail.google.com&sz=128'
		},
		{
			name: 'Google Calendar',
			url: 'https://calendar.google.com',
			color: '#4285F4',
			category: 'productivity',
			description: 'Calendar & scheduling',
			logo: 'https://www.google.com/s2/favicons?domain=calendar.google.com&sz=128'
		},
		{
			name: 'Outlook',
			url: 'https://outlook.live.com',
			color: '#0078D4',
			category: 'communication',
			description: 'Email by Microsoft',
			logo: 'https://www.google.com/s2/favicons?domain=outlook.live.com&sz=128'
		},
		// Video & Meeting
		{
			name: 'Zoom',
			url: 'https://zoom.us/meeting',
			color: '#2D8CFF',
			category: 'communication',
			description: 'Video meetings',
			logo: 'https://www.google.com/s2/favicons?domain=zoom.us&sz=128'
		},
		{
			name: 'Google Meet',
			url: 'https://meet.google.com',
			color: '#00897B',
			category: 'communication',
			description: 'Video meetings',
			logo: 'https://www.google.com/s2/favicons?domain=meet.google.com&sz=128'
		},
		{
			name: 'Microsoft Teams',
			url: 'https://teams.microsoft.com',
			color: '#6264A7',
			category: 'communication',
			description: 'Team collaboration',
			logo: 'https://www.google.com/s2/favicons?domain=teams.microsoft.com&sz=128'
		},
		// Social & Media
		{
			name: 'YouTube',
			url: 'https://www.youtube.com',
			color: '#FF0000',
			category: 'media',
			description: 'Video streaming',
			logo: 'https://www.google.com/s2/favicons?domain=youtube.com&sz=128'
		},
		{
			name: 'X (Twitter)',
			url: 'https://x.com',
			color: '#000000',
			category: 'social',
			description: 'Social network',
			logo: 'https://www.google.com/s2/favicons?domain=x.com&sz=128'
		},
		{
			name: 'LinkedIn',
			url: 'https://www.linkedin.com',
			color: '#0A66C2',
			category: 'social',
			description: 'Professional network',
			logo: 'https://www.google.com/s2/favicons?domain=linkedin.com&sz=128'
		},
		{
			name: 'WhatsApp',
			url: 'https://web.whatsapp.com',
			color: '#25D366',
			category: 'communication',
			description: 'Messaging',
			logo: 'https://www.google.com/s2/favicons?domain=web.whatsapp.com&sz=128'
		},
		// Music & Entertainment
		{
			name: 'Spotify',
			url: 'https://open.spotify.com',
			color: '#1DB954',
			category: 'media',
			description: 'Music streaming',
			logo: 'https://www.google.com/s2/favicons?domain=open.spotify.com&sz=128'
		},
		// CRM & Business
		{
			name: 'HubSpot',
			url: 'https://app.hubspot.com',
			color: '#FF7A59',
			category: 'business',
			description: 'CRM & marketing',
			logo: 'https://www.google.com/s2/favicons?domain=hubspot.com&sz=128'
		},
		{
			name: 'Salesforce',
			url: 'https://login.salesforce.com',
			color: '#00A1E0',
			category: 'business',
			description: 'CRM platform',
			logo: 'https://www.google.com/s2/favicons?domain=salesforce.com&sz=128'
		},
		// More Project Management
		{
			name: 'Jira',
			url: 'https://www.atlassian.com/software/jira',
			color: '#0052CC',
			category: 'project-management',
			description: 'Issue tracking',
			logo: 'https://www.google.com/s2/favicons?domain=jira.atlassian.com&sz=128'
		},
		{
			name: 'Todoist',
			url: 'https://todoist.com',
			color: '#E44332',
			category: 'productivity',
			description: 'Task management',
			logo: 'https://www.google.com/s2/favicons?domain=todoist.com&sz=128'
		},
		// Notes & Writing
		{
			name: 'Obsidian',
			url: 'https://obsidian.md',
			color: '#7C3AED',
			category: 'productivity',
			description: 'Note-taking',
			logo: 'https://www.google.com/s2/favicons?domain=obsidian.md&sz=128'
		},
		{
			name: 'Evernote',
			url: 'https://www.evernote.com',
			color: '#00A82D',
			category: 'productivity',
			description: 'Note-taking',
			logo: 'https://www.google.com/s2/favicons?domain=evernote.com&sz=128'
		},
		// Development Tools
		{
			name: 'Bitbucket',
			url: 'https://bitbucket.org',
			color: '#0052CC',
			category: 'development',
			description: 'Code hosting',
			logo: 'https://www.google.com/s2/favicons?domain=bitbucket.org&sz=128'
		},
		{
			name: 'Netlify',
			url: 'https://app.netlify.com',
			color: '#00C7B7',
			category: 'development',
			description: 'Deploy & host',
			logo: 'https://www.google.com/s2/favicons?domain=netlify.com&sz=128'
		},
		{
			name: 'AWS Console',
			url: 'https://console.aws.amazon.com',
			color: '#FF9900',
			category: 'development',
			description: 'Cloud services',
			logo: 'https://www.google.com/s2/favicons?domain=aws.amazon.com&sz=128'
		}
	];

	const filteredApps = $derived(
		searchQuery
			? popularApps.filter(
					(app) =>
						app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
						app.description.toLowerCase().includes(searchQuery.toLowerCase())
				)
			: popularApps
	);

	// Fetch user apps on mount
	$effect(() => {
		userAppsStore.fetch(workspaceId);
	});

	async function quickAdd(app: (typeof popularApps)[0]) {
		isSubmitting = true;
		try {
			const params: CreateUserAppParams = {
				workspace_id: workspaceId,
				name: app.name,
				url: app.url,
				color: app.color,
				logo_url: app.logo, // Pass the logo URL from popular apps
				category: app.category,
				description: app.description,
				app_type: 'web'
			};

			const newApp = await userAppsStore.create(params);

			// Add to 3D desktop sphere if available
			if (newApp) {
				desktop3dStore.addUserApp(newApp);
			}

			dispatch('appCreated');
		} catch (error) {
			console.error('Failed to add app:', error);
			alert('Failed to add app. Please try again.');
		} finally {
			isSubmitting = false;
		}
	}

	async function handleCustomSubmit(e: Event) {
		e.preventDefault();

		if (!name || !url) {
			return;
		}

		isSubmitting = true;

		try {
			const params: CreateUserAppParams = {
				workspace_id: workspaceId,
				name,
				url,
				color: '#6366F1',
				category,
				description: description || undefined,
				app_type: 'web'
			};

			const newApp = await userAppsStore.create(params);

			// Add to 3D desktop sphere if available
			if (newApp) {
				desktop3dStore.addUserApp(newApp);
			}

			// Reset form
			name = '';
			url = '';
			category = 'productivity';
			description = '';

			dispatch('appCreated');
			onClose();
		} catch (error) {
			console.error('Failed to create app:', error);
			alert('Failed to create app. Please try again.');
		} finally {
			isSubmitting = false;
		}
	}

	async function deleteApp(appId: string) {
		if (confirm('Remove this app?')) {
			try {
				await userAppsStore.delete(appId, workspaceId);
				// Also remove from 3D desktop sphere
				desktop3dStore.removeUserApp(appId);
			} catch (error) {
				console.error('Failed to delete app:', error);
				alert('Failed to delete app. Please try again.');
			}
		}
	}
</script>

<div class="modal-overlay" role="dialog" aria-modal="true" transition:fade={{ duration: 200 }}>
	<div class="modal-container" transition:fly={{ y: 20, duration: 300, easing: cubicOut }}>
		<!-- Header -->
		<div class="modal-header">
			<div class="header-content">
				<div class="header-icon">
					<Sparkles size={24} />
				</div>
				<div class="header-text">
					<h2>Add Applications</h2>
					<p>Connect your favorite tools to BusinessOS</p>
				</div>
			</div>
			<button class="close-btn" onclick={onClose} aria-label="Close">
				<X size={20} />
			</button>
		</div>

		<!-- Tabs -->
		<div class="tabs">
			<button
				class="tab"
				class:active={activeTab === 'browse'}
				onclick={() => (activeTab = 'browse')}
			>
				<Globe size={16} />
				Popular Apps
			</button>
			<button
				class="tab"
				class:active={activeTab === 'custom'}
				onclick={() => (activeTab = 'custom')}
			>
				<Plus size={16} />
				Custom URL
			</button>
		</div>

		<!-- Content -->
		<div class="modal-body">
			{#if activeTab === 'browse'}
				<!-- Search -->
				<div class="search-container">
					<Search size={18} class="search-icon" />
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search apps..."
						class="search-input"
					/>
				</div>

				<!-- Your Apps Section -->
				{#if $userAppsStore.apps.length > 0}
					<div class="section">
						<h3 class="section-title">Your Apps ({$userAppsStore.apps.length})</h3>

						<!-- Warning for native apps -->
						{#if hasNativeApps}
							<div class="native-warning">
								<AlertTriangle size={16} />
								<span>Some apps are native desktop apps (not web apps). Delete them and re-add from Popular Apps to use in 3D Desktop.</span>
							</div>
						{/if}

						<div class="your-apps-grid">
							{#each $userAppsStore.apps as app (app.id)}
								<div
									class="your-app-card"
									class:native-app={isNativeBundleId(app.url)}
									transition:scale={{ duration: 200 }}
								>
									{#if isNativeBundleId(app.url)}
										<div class="native-badge" title="Native app - delete and re-add as web app">
											<AlertTriangle size={12} />
										</div>
									{/if}
									<div class="app-logo-container">
										{#if app.logo_url}
											<img src={app.logo_url} alt={app.name} class="app-logo-img" />
										{:else}
											<div class="app-logo-placeholder" style="background-color: {app.color};">
												<ExternalLink size={20} class="logo-icon" />
											</div>
										{/if}
									</div>
									<div class="app-info">
										<span class="app-name">{app.name}</span>
										<span class="app-url" class:native-url={isNativeBundleId(app.url)}>
											{#if isNativeBundleId(app.url)}
												Native App
											{:else}
												{(() => { try { return new URL(app.url).hostname; } catch { return app.url; } })()}
											{/if}
										</span>
									</div>
									<button
										class="delete-btn"
										onclick={() => deleteApp(app.id)}
										aria-label="Delete {app.name}"
										title={isNativeBundleId(app.url) ? 'Delete this native app' : 'Delete app'}
									>
										<Trash2 size={14} />
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Popular Apps Grid -->
				<div class="section">
					<h3 class="section-title">Popular Apps</h3>
					<div class="apps-grid">
						{#each filteredApps as app (app.name)}
							<button
								class="app-card"
								onclick={() => quickAdd(app)}
								disabled={isSubmitting}
								transition:scale={{ duration: 200, delay: 50 }}
							>
								<div class="app-card-content">
									<div class="app-logo-container">
										<img src={app.logo} alt={app.name} class="app-logo-img" />
									</div>
									<div class="app-details">
										<span class="app-title">{app.name}</span>
										<span class="app-desc">{app.description}</span>
									</div>
								</div>
								<div class="app-add-icon">
									<Plus size={16} />
								</div>
							</button>
						{/each}
					</div>
				</div>
			{:else}
				<!-- Custom App Form -->
				<form onsubmit={handleCustomSubmit} class="custom-form">
					<div class="form-group">
						<label for="app-name" class="form-label">
							App Name
							<span class="required">*</span>
						</label>
						<input
							id="app-name"
							type="text"
							bind:value={name}
							placeholder="e.g., My Custom Tool"
							required
							class="form-input"
						/>
					</div>

					<div class="form-group">
						<label for="app-url" class="form-label">
							URL
							<span class="required">*</span>
						</label>
						<div class="url-input-wrapper">
							<Globe size={18} class="url-icon" />
							<input
								id="app-url"
								type="url"
								bind:value={url}
								placeholder="https://example.com"
								required
								class="form-input url-input"
							/>
						</div>
						<span class="input-hint">The logo will be automatically fetched from this URL</span>
					</div>

					<div class="form-group">
						<label for="app-category" class="form-label">Category</label>
						<select id="app-category" bind:value={category} class="form-select">
							<option value="productivity">Productivity</option>
							<option value="communication">Communication</option>
							<option value="project-management">Project Management</option>
							<option value="design">Design</option>
							<option value="development">Development</option>
							<option value="ai">AI & Automation</option>
							<option value="collaboration">Collaboration</option>
							<option value="other">Other</option>
						</select>
					</div>

					<div class="form-group">
						<label for="app-description" class="form-label">Description (optional)</label>
						<textarea
							id="app-description"
							bind:value={description}
							placeholder="Brief description of this app..."
							rows="3"
							class="form-textarea"
						/>
					</div>

					<button type="submit" class="submit-btn" disabled={!name || !url || isSubmitting}>
						{#if isSubmitting}
							<span class="spinner"></span>
							Adding...
						{:else}
							<Plus size={18} />
							Add Custom App
						{/if}
					</button>
				</form>
			{/if}
		</div>
	</div>
</div>

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10000;
		padding: 1.5rem;
	}

	.modal-container {
		background: white;
		border-radius: 1rem;
		box-shadow:
			0 25px 50px -12px rgba(0, 0, 0, 0.25),
			0 0 0 1px rgba(0, 0, 0, 0.05);
		max-width: 900px;
		width: 100%;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* Header */
	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 2rem 2rem 1.5rem;
		border-bottom: 1px solid #f3f4f6;
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.header-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		border-radius: 0.75rem;
		color: white;
		box-shadow: 0 4px 6px -1px rgba(102, 126, 234, 0.3);
	}

	.header-text h2 {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.25rem 0;
	}

	.header-text p {
		font-size: 0.875rem;
		color: #6b7280;
		margin: 0;
	}

	.close-btn {
		background: #f3f4f6;
		border: none;
		border-radius: 0.5rem;
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: #6b7280;
		transition: all 0.2s;
	}

	.close-btn:hover {
		background: #e5e7eb;
		color: #111827;
	}

	/* Tabs */
	.tabs {
		display: flex;
		gap: 0.5rem;
		padding: 0 2rem;
		border-bottom: 1px solid #f3f4f6;
	}

	.tab {
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}

	.tab:hover {
		color: #111827;
	}

	.tab.active {
		color: #6366f1;
		border-bottom-color: #6366f1;
	}

	/* Body */
	.modal-body {
		flex: 1;
		overflow-y: auto;
		padding: 2rem;
	}

	/* Search */
	.search-container {
		position: relative;
		margin-bottom: 2rem;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem 1rem 0.75rem 2.75rem;
		border: 2px solid #e5e7eb;
		border-radius: 0.75rem;
		font-size: 0.875rem;
		transition: all 0.2s;
		background: white;
	}

	.search-input:focus {
		outline: none;
		border-color: #6366f1;
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	:global(.search-icon) {
		position: absolute;
		left: 1rem;
		top: 50%;
		transform: translateY(-50%);
		color: #9ca3af;
		pointer-events: none;
	}

	/* Section */
	.section {
		margin-bottom: 2.5rem;
	}

	.section:last-child {
		margin-bottom: 0;
	}

	.section-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0 0 1rem 0;
	}

	/* Your Apps Grid */
	.your-apps-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 0.75rem;
		margin-bottom: 2rem;
	}

	.your-app-card {
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		transition: all 0.2s;
	}

	.your-app-card:hover {
		background: #f3f4f6;
		border-color: #d1d5db;
	}

	.your-app-card:hover .delete-btn {
		opacity: 1;
	}

	/* Native app warning styles */
	.native-warning {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: #fef3c7;
		border: 1px solid #fcd34d;
		border-radius: 0.5rem;
		margin-bottom: 1rem;
		font-size: 0.813rem;
		color: #92400e;
	}

	.native-warning :global(svg) {
		flex-shrink: 0;
		color: #d97706;
	}

	.your-app-card.native-app {
		background: #fef3c7;
		border-color: #fcd34d;
	}

	.your-app-card.native-app:hover {
		background: #fde68a;
		border-color: #fbbf24;
	}

	.native-badge {
		position: absolute;
		top: -6px;
		right: -6px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		background: #d97706;
		border-radius: 50%;
		color: white;
	}

	.native-url {
		color: #d97706 !important;
		font-weight: 500;
	}

	.delete-btn {
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.375rem;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: #ef4444;
		opacity: 0;
		transition: all 0.2s;
	}

	.delete-btn:hover {
		background: #fee2e2;
		border-color: #fecaca;
	}

	/* Apps Grid */
	.apps-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 1rem;
	}

	.app-card {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem;
		background: white;
		border: 2px solid #e5e7eb;
		border-radius: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
	}

	.app-card:hover:not(:disabled) {
		border-color: #6366f1;
		box-shadow: 0 4px 6px -1px rgba(99, 102, 241, 0.1);
		transform: translateY(-2px);
	}

	.app-card:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.app-card-content {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
	}

	.app-logo-container {
		width: 48px;
		height: 48px;
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.5rem;
		overflow: hidden;
		background: #f9fafb;
	}

	.app-logo-img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.app-logo-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
	}

	:global(.logo-icon) {
		color: white;
	}

	.app-details {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.app-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: #111827;
	}

	.app-desc {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.app-add-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		background: #f3f4f6;
		border-radius: 0.5rem;
		color: #6b7280;
		transition: all 0.2s;
	}

	.app-card:hover:not(:disabled) .app-add-icon {
		background: #6366f1;
		color: white;
	}

	.app-info {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
		flex: 1;
		min-width: 0;
	}

	.app-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: #111827;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.app-url {
		font-size: 0.75rem;
		color: #6b7280;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	/* Custom Form */
	.custom-form {
		max-width: 500px;
		margin: 0 auto;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.required {
		color: #ef4444;
	}

	.form-input,
	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 2px solid #e5e7eb;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		transition: all 0.2s;
		background: white;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #6366f1;
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.url-input-wrapper {
		position: relative;
	}

	.url-input {
		padding-left: 2.75rem;
	}

	:global(.url-icon) {
		position: absolute;
		left: 1rem;
		top: 50%;
		transform: translateY(-50%);
		color: #9ca3af;
		pointer-events: none;
	}

	.input-hint {
		display: block;
		font-size: 0.75rem;
		color: #6b7280;
		margin-top: 0.5rem;
	}

	.form-textarea {
		resize: vertical;
		min-height: 80px;
	}

	.submit-btn {
		width: 100%;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		border: none;
		padding: 0.875rem 1.5rem;
		border-radius: 0.75rem;
		font-weight: 600;
		font-size: 0.875rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		transition: all 0.2s;
		box-shadow: 0 4px 6px -1px rgba(102, 126, 234, 0.3);
	}

	.submit-btn:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 6px 8px -1px rgba(102, 126, 234, 0.4);
	}

	.submit-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	/* Loading and Empty States */
	.loading-state,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		text-align: center;
		color: #64748b;
	}

	.loading-state .spinner {
		width: 32px;
		height: 32px;
		border: 3px solid rgba(99, 102, 241, 0.2);
		border-top-color: #6366f1;
		margin-bottom: 1rem;
	}

	.loading-state p {
		font-size: 0.875rem;
		margin: 0;
	}

	.empty-state p {
		margin: 0.5rem 0;
	}

	.empty-state .hint {
		font-size: 0.8125rem;
		color: #94a3b8;
	}

	/* Running Badge */
	.running-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.125rem 0.5rem;
		background: linear-gradient(135deg, #10b981 0%, #059669 100%);
		color: white;
		font-size: 0.6875rem;
		font-weight: 600;
		border-radius: 9999px;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.running-badge::before {
		content: '';
		width: 6px;
		height: 6px;
		background: white;
		border-radius: 50%;
		animation: pulse 2s ease-in-out infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	/* Native App Card */
	.native-app-card {
		border: 2px solid #e2e8f0;
	}

	.native-app-card:hover {
		border-color: #818cf8;
		transform: translateY(-2px);
	}

	.native-logo {
		width: 64px;
		height: 64px;
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.native-logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.app-logo {
		width: 64px;
		height: 64px;
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.app-logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
</style>
