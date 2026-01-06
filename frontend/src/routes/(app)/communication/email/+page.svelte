<script lang="ts">
	import { onMount } from 'svelte';
	import {
		checkGmailAccess,
		requestGmailAccess,
		getEmails,
		getEmail,
		markAsRead,
		syncEmails,
		sendEmail,
		getGmailStats,
		type Email,
		type EmailFolder,
		type GmailAccessStatus,
		type GmailStats
	} from '$lib/api/gmail';

	// State
	let accessStatus = $state<GmailAccessStatus | null>(null);
	let emails = $state<Email[]>([]);
	let selectedEmail = $state<Email | null>(null);
	let stats = $state<GmailStats | null>(null);
	let isLoading = $state(true);
	let isSyncing = $state(false);
	let isSending = $state(false);
	let error = $state<string | null>(null);
	let currentFolder = $state<EmailFolder>('inbox');
	let showComposeModal = $state(false);
	let searchQuery = $state('');

	// Compose form state
	let composeTo = $state('');
	let composeCc = $state('');
	let composeSubject = $state('');
	let composeBody = $state('');
	let composeError = $state<string | null>(null);
	let replyTo = $state<Email | null>(null);

	// Folders
	const folders: { id: EmailFolder; label: string; icon: string }[] = [
		{ id: 'inbox', label: 'Inbox', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'sent', label: 'Sent', icon: 'M12 19l9 2-9-18-9 18 9-2zm0 0v-8' },
		{ id: 'drafts', label: 'Drafts', icon: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' },
		{ id: 'starred', label: 'Starred', icon: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z' },
		{ id: 'archive', label: 'Archive', icon: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		{ id: 'trash', label: 'Trash', icon: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' },
	];

	// Helper functions
	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const isToday = date.toDateString() === now.toDateString();
		const isThisYear = date.getFullYear() === now.getFullYear();

		if (isToday) {
			return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
		} else if (isThisYear) {
			return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
		} else {
			return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
		}
	}

	function truncate(str: string, length: number): string {
		if (str.length <= length) return str;
		return str.slice(0, length) + '...';
	}

	// API functions
	async function loadAccessStatus() {
		try {
			accessStatus = await checkGmailAccess();
		} catch (e) {
			console.error('Failed to check Gmail access:', e);
		}
	}

	async function loadEmails() {
		if (!accessStatus?.has_access) return;

		isLoading = true;
		error = null;
		try {
			emails = await getEmails({ folder: currentFolder, limit: 50 });
		} catch (e: any) {
			if (e.message === 'REQUIRES_UPGRADE') {
				error = 'Gmail access requires additional permissions';
			} else {
				error = 'Failed to load emails';
			}
			console.error(e);
		} finally {
			isLoading = false;
		}
	}

	async function loadStats() {
		if (!accessStatus?.has_access) return;

		try {
			stats = await getGmailStats();
		} catch (e) {
			console.error('Failed to load Gmail stats:', e);
		}
	}

	async function handleSync() {
		isSyncing = true;
		try {
			await syncEmails(100);
			await loadEmails();
			await loadStats();
		} catch (e: any) {
			if (e.message === 'REQUIRES_UPGRADE') {
				error = 'Gmail sync requires additional permissions';
			} else {
				error = 'Failed to sync emails';
			}
		} finally {
			isSyncing = false;
		}
	}

	async function handleSelectEmail(email: Email) {
		selectedEmail = email;

		if (!email.is_read) {
			try {
				await markAsRead(email.id);
				// Update local state
				const idx = emails.findIndex(e => e.id === email.id);
				if (idx !== -1) {
					emails[idx] = { ...emails[idx], is_read: true };
				}
			} catch (e) {
				console.error('Failed to mark email as read:', e);
			}
		}
	}

	async function handleRequestAccess() {
		try {
			const result = await requestGmailAccess();
			if (result.auth_url) {
				window.location.href = result.auth_url;
			}
		} catch (e) {
			error = 'Failed to request Gmail access';
			console.error(e);
		}
	}

	async function handleSendEmail() {
		if (!composeTo.trim()) {
			composeError = 'Please enter a recipient';
			return;
		}

		isSending = true;
		composeError = null;

		try {
			await sendEmail({
				to: composeTo.split(',').map(e => e.trim()).filter(Boolean),
				cc: composeCc ? composeCc.split(',').map(e => e.trim()).filter(Boolean) : undefined,
				subject: composeSubject,
				body: composeBody,
				is_html: false,
				reply_to: replyTo?.external_id
			});

			// Reset form and close modal
			resetComposeForm();
			showComposeModal = false;

			// Refresh sent folder if viewing it
			if (currentFolder === 'sent') {
				await loadEmails();
			}
		} catch (e: any) {
			composeError = e.message || 'Failed to send email';
		} finally {
			isSending = false;
		}
	}

	function resetComposeForm() {
		composeTo = '';
		composeCc = '';
		composeSubject = '';
		composeBody = '';
		composeError = null;
		replyTo = null;
	}

	function openReply(email: Email) {
		replyTo = email;
		composeTo = email.from_email;
		composeSubject = email.subject?.startsWith('Re:') ? email.subject : `Re: ${email.subject || ''}`;
		composeBody = `\n\n---\nOn ${new Date(email.date).toLocaleString()}, ${email.from_name || email.from_email} wrote:\n${email.body_text || email.snippet || ''}`;
		showComposeModal = true;
	}

	function openForward(email: Email) {
		composeTo = '';
		composeSubject = email.subject?.startsWith('Fwd:') ? email.subject : `Fwd: ${email.subject || ''}`;
		composeBody = `\n\n---\nForwarded message:\nFrom: ${email.from_name || email.from_email} <${email.from_email}>\nDate: ${new Date(email.date).toLocaleString()}\nSubject: ${email.subject || ''}\n\n${email.body_text || email.snippet || ''}`;
		showComposeModal = true;
	}

	// Filter emails by search query
	const filteredEmails = $derived(
		searchQuery.trim()
			? emails.filter(e =>
				e.subject?.toLowerCase().includes(searchQuery.toLowerCase()) ||
				e.from_email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
				e.from_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
				e.snippet?.toLowerCase().includes(searchQuery.toLowerCase())
			)
			: emails
	);

	// Folder change handler
	$effect(() => {
		if (accessStatus?.has_access) {
			loadEmails();
		}
	});

	onMount(async () => {
		await loadAccessStatus();
		if (accessStatus?.has_access) {
			await Promise.all([loadEmails(), loadStats()]);
		} else {
			isLoading = false;
		}
	});
</script>

<div class="h-full flex">
	{#if !accessStatus?.has_access}
		<!-- Gmail Not Connected -->
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center max-w-md">
				<div class="w-16 h-16 mx-auto mb-6 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
					<svg class="w-8 h-8 text-red-600 dark:text-red-400" viewBox="0 0 24 24" fill="currentColor">
						<path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
					</svg>
				</div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Connect Gmail</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					{accessStatus?.message || 'Connect your Gmail account to view and manage your emails directly from BusinessOS.'}
				</p>
				{#if accessStatus?.requires_upgrade}
					<p class="text-sm text-amber-600 dark:text-amber-400 mb-4">
						Your current Google connection needs additional permissions for Gmail access.
					</p>
				{/if}
				<button
					onclick={handleRequestAccess}
					class="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
				>
					<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
						<path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
					</svg>
					{accessStatus?.requires_upgrade ? 'Upgrade Permissions' : 'Connect Gmail'}
				</button>
			</div>
		</div>
	{:else}
		<!-- Folders Sidebar -->
		<div class="w-56 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col">
			<!-- Compose Button -->
			<div class="p-4">
				<button
					onclick={() => showComposeModal = true}
					class="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Compose
				</button>
			</div>

			<!-- Folder List -->
			<nav class="flex-1 px-2 space-y-1">
				{#each folders as folder}
					<button
						onclick={() => { currentFolder = folder.id; selectedEmail = null; }}
						class="w-full flex items-center gap-3 px-3 py-2 text-sm rounded-lg transition-colors
							{currentFolder === folder.id
								? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
								: 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={folder.icon} />
						</svg>
						<span class="flex-1 text-left">{folder.label}</span>
						{#if folder.id === 'inbox' && stats?.unread_count}
							<span class="px-2 py-0.5 text-xs font-medium bg-blue-600 text-white rounded-full">
								{stats.unread_count}
							</span>
						{/if}
					</button>
				{/each}
			</nav>

			<!-- Storage Info -->
			<div class="p-4 border-t border-gray-200 dark:border-gray-700">
				<p class="text-xs text-gray-500 dark:text-gray-400">
					{stats?.total_emails || 0} emails synced
				</p>
			</div>
		</div>

		<!-- Email List -->
		<div class="w-80 border-r border-gray-200 dark:border-gray-700 flex flex-col">
			<!-- List Header -->
			<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
				<div class="flex items-center justify-between mb-2">
					<h3 class="font-medium text-gray-900 dark:text-white capitalize">{currentFolder}</h3>
					<button
						onclick={handleSync}
						disabled={isSyncing}
						class="p-1.5 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors disabled:opacity-50"
						title="Sync emails"
					>
						<svg class="w-5 h-5 {isSyncing ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
					</button>
				</div>
				<!-- Search -->
				<div class="relative">
					<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search emails..."
						class="w-full pl-9 pr-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
			</div>

			<!-- Email List -->
			<div class="flex-1 overflow-y-auto">
				{#if isLoading}
					<div class="flex items-center justify-center py-12">
						<svg class="w-6 h-6 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
					</div>
				{:else if error}
					<div class="p-4 text-center">
						<p class="text-red-500 dark:text-red-400 text-sm">{error}</p>
						<button onclick={loadEmails} class="mt-2 text-blue-600 dark:text-blue-400 text-sm hover:underline">
							Try again
						</button>
					</div>
				{:else if filteredEmails.length === 0}
					<div class="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
						<svg class="w-10 h-10 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
						</svg>
						<p class="text-sm">{searchQuery ? 'No matching emails' : `No emails in ${currentFolder}`}</p>
					</div>
				{:else}
					{#each filteredEmails as email}
						<button
							onclick={() => handleSelectEmail(email)}
							class="w-full text-left px-4 py-3 border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors
								{selectedEmail?.id === email.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''}
								{!email.is_read ? 'bg-white dark:bg-gray-900' : ''}"
						>
							<div class="flex items-start gap-3">
								{#if !email.is_read}
									<span class="w-2 h-2 mt-2 rounded-full bg-blue-600 flex-shrink-0"></span>
								{:else}
									<span class="w-2 h-2 mt-2 rounded-full flex-shrink-0"></span>
								{/if}
								<div class="flex-1 min-w-0">
									<div class="flex items-center justify-between gap-2">
										<p class="text-sm font-medium text-gray-900 dark:text-white truncate {!email.is_read ? 'font-semibold' : ''}">
											{email.from_name || email.from_email}
										</p>
										<span class="text-xs text-gray-500 dark:text-gray-400 flex-shrink-0">
											{formatDate(email.date)}
										</span>
									</div>
									<p class="text-sm text-gray-900 dark:text-white truncate {!email.is_read ? 'font-medium' : ''}">
										{email.subject || '(no subject)'}
									</p>
									<p class="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
										{email.snippet || ''}
									</p>
								</div>
							</div>
						</button>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Email Preview -->
		<div class="flex-1 flex flex-col bg-white dark:bg-gray-900">
			{#if selectedEmail}
				<!-- Email Header -->
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h2 class="text-xl font-semibold text-gray-900 dark:text-white">
						{selectedEmail.subject || '(no subject)'}
					</h2>
					<div class="mt-3 flex items-start gap-3">
						<div class="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-600 dark:text-gray-300 font-medium">
							{(selectedEmail.from_name || selectedEmail.from_email).charAt(0).toUpperCase()}
						</div>
						<div class="flex-1">
							<p class="font-medium text-gray-900 dark:text-white">
								{selectedEmail.from_name || selectedEmail.from_email}
							</p>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								{selectedEmail.from_email}
							</p>
							<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
								{new Date(selectedEmail.date).toLocaleString()}
							</p>
						</div>
					</div>
				</div>

				<!-- Email Body -->
				<div class="flex-1 overflow-y-auto p-6">
					{#if selectedEmail.body_html}
						<div class="prose dark:prose-invert max-w-none">
							{@html selectedEmail.body_html}
						</div>
					{:else if selectedEmail.body_text}
						<pre class="whitespace-pre-wrap text-gray-700 dark:text-gray-300 font-sans">{selectedEmail.body_text}</pre>
					{:else}
						<p class="text-gray-500 dark:text-gray-400 italic">No content</p>
					{/if}
				</div>

				<!-- Email Actions -->
				<div class="px-6 py-3 border-t border-gray-200 dark:border-gray-700 flex items-center gap-2">
					<button
						onclick={() => openReply(selectedEmail)}
						class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
						</svg>
						Reply
					</button>
					<button
						onclick={() => openForward(selectedEmail)}
						class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
						</svg>
						Forward
					</button>
					<button class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8" />
						</svg>
						Archive
					</button>
				</div>
			{:else}
				<!-- No Email Selected -->
				<div class="flex-1 flex items-center justify-center text-gray-500 dark:text-gray-400">
					<div class="text-center">
						<svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
						</svg>
						<p>Select an email to read</p>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Compose Modal -->
{#if showComposeModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => { resetComposeForm(); showComposeModal = false; }}>
		<div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-2xl mx-4" onclick={(e) => e.stopPropagation()}>
			<div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
				<h3 class="font-semibold text-gray-900 dark:text-white">
					{replyTo ? 'Reply' : 'New Message'}
				</h3>
				<button
					onclick={() => { resetComposeForm(); showComposeModal = false; }}
					class="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			{#if composeError}
				<div class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
					<p class="text-sm text-red-600 dark:text-red-400">{composeError}</p>
				</div>
			{/if}

			<div class="p-4 space-y-3">
				<div class="flex items-center gap-2">
					<label class="w-12 text-sm text-gray-500 dark:text-gray-400">To:</label>
					<input
						type="text"
						bind:value={composeTo}
						placeholder="recipient@example.com"
						class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
				<div class="flex items-center gap-2">
					<label class="w-12 text-sm text-gray-500 dark:text-gray-400">Cc:</label>
					<input
						type="text"
						bind:value={composeCc}
						placeholder="cc@example.com (optional)"
						class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
				<div class="flex items-center gap-2">
					<label class="w-12 text-sm text-gray-500 dark:text-gray-400">Subject:</label>
					<input
						type="text"
						bind:value={composeSubject}
						placeholder="Email subject"
						class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
				<div>
					<textarea
						bind:value={composeBody}
						rows="12"
						placeholder="Write your message..."
						class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
					></textarea>
				</div>
			</div>

			<div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700">
				<div class="flex items-center gap-2">
					<button
						class="p-2 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
						title="Attach file"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
						</svg>
					</button>
				</div>
				<div class="flex items-center gap-3">
					<button
						onclick={() => { resetComposeForm(); showComposeModal = false; }}
						class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
					>
						Discard
					</button>
					<button
						onclick={handleSendEmail}
						disabled={isSending || !composeTo.trim()}
						class="flex items-center gap-2 px-4 py-2 text-sm font-medium bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
					>
						{#if isSending}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
							</svg>
							Sending...
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
							</svg>
							Send
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
