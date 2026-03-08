<script lang="ts">
	import { onMount } from 'svelte';
	import DOMPurify from 'dompurify';
	import { Inbox, Send, FileEdit, Star, Archive, Trash2, Plus, Search, RefreshCw, Reply, Forward, Archive as ArchiveIcon, X, Paperclip, Mail, Loader2 } from 'lucide-svelte';
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

	// Folder icon components mapped
	const folderMeta: { id: EmailFolder; label: string }[] = [
		{ id: 'inbox', label: 'Inbox' },
		{ id: 'sent', label: 'Sent' },
		{ id: 'drafts', label: 'Drafts' },
		{ id: 'starred', label: 'Starred' },
		{ id: 'archive', label: 'Archive' },
		{ id: 'trash', label: 'Trash' },
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

<div class="ch-inbox-layout">
	{#if !accessStatus?.has_access}
		<!-- Gmail Not Connected -->
		<div class="ch-inbox-empty">
			<div class="ch-inbox-empty__icon">
				<Mail size={32} />
			</div>
			<h2 class="ch-inbox-empty__title">Connect Gmail</h2>
			<p class="ch-inbox-empty__desc">
				{accessStatus?.message || 'Connect your Gmail account to view and manage your emails directly from BusinessOS.'}
			</p>
			{#if accessStatus?.requires_upgrade}
				<p class="ch-inbox-empty__upgrade">
					Your current Google connection needs additional permissions for Gmail access.
				</p>
			{/if}
			<button onclick={handleRequestAccess} class="btn-pill btn-pill-primary" aria-label="Connect Gmail account">
				<Mail size={18} />
				{accessStatus?.requires_upgrade ? 'Upgrade Permissions' : 'Connect Gmail'}
			</button>
		</div>
	{:else}
		<!-- Folder Sidebar -->
		<aside class="ch-inbox-sidebar">
			<button
				onclick={() => { showComposeModal = true; }}
				class="btn-pill btn-pill-primary btn-pill-sm ch-inbox-compose-btn"
				aria-label="Compose new email"
			>
				<Plus size={16} />
				Compose
			</button>

			<ul class="ch-inbox-folders">
				{#each folderMeta as folder}
					<li class="ch-inbox-folder {currentFolder === folder.id ? 'ch-inbox-folder--active' : ''}">
						<button
							class="ch-inbox-folder__btn"
							aria-label="Go to {folder.label} folder"
							onclick={() => { currentFolder = folder.id; selectedEmail = null; }}
						>
							<span class="ch-inbox-folder__icon">
								{#if folder.id === 'inbox'}<Inbox size={15} />
								{:else if folder.id === 'sent'}<Send size={15} />
								{:else if folder.id === 'drafts'}<FileEdit size={15} />
								{:else if folder.id === 'starred'}<Star size={15} />
								{:else if folder.id === 'archive'}<Archive size={15} />
								{:else if folder.id === 'trash'}<Trash2 size={15} />
								{/if}
							</span>
							<span class="ch-inbox-folder__label">{folder.label}</span>
							{#if folder.id === 'inbox' && stats?.unread_count}
								<span class="ch-inbox-folder__count">{stats.unread_count}</span>
							{/if}
						</button>
					</li>
				{/each}
			</ul>

			<div class="ch-inbox-sidebar__footer">
				<p class="ch-inbox-sidebar__stat">{stats?.total_emails || 0} emails synced</p>
			</div>
		</aside>

		<!-- Email List Panel -->
		<div class="ch-inbox-main">
			<div class="ch-inbox-toolbar">
				<h3 class="ch-inbox-toolbar__title">{currentFolder}</h3>
				<div class="ch-inbox-toolbar__spacer"></div>
				<button
					onclick={handleSync}
					disabled={isSyncing}
					class="btn-compact btn-compact-ghost btn-compact-icon"
					aria-label="Sync emails"
				>
					<RefreshCw size={16} class={isSyncing ? 'ch-spin' : ''} />
				</button>
			</div>

			<div class="ch-inbox-search">
				<Search size={14} class="ch-inbox-search__icon" />
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search emails..."
					class="ch-inbox-search__input"
					aria-label="Search emails"
				/>
			</div>

			<ul class="ch-inbox-list">
				{#if isLoading}
					<li class="ch-inbox-loading">
						<Loader2 size={20} class="ch-spin" />
					</li>
				{:else if error}
					<li class="ch-inbox-error">
						<p class="ch-inbox-error__text">{error}</p>
						<button onclick={loadEmails} class="btn-compact btn-compact-ghost btn-compact-sm" aria-label="Retry loading emails">
							Try again
						</button>
					</li>
				{:else if filteredEmails.length === 0}
					<li class="ch-inbox-loading">
						<Mail size={28} strokeWidth={1.5} />
						<span class="ch-inbox-loading__text">{searchQuery ? 'No matching emails' : `No emails in ${currentFolder}`}</span>
					</li>
				{:else}
					{#each filteredEmails as email}
						<li class="ch-inbox-row {!email.is_read ? 'ch-inbox-row--unread' : ''} {selectedEmail?.id === email.id ? 'ch-inbox-row--selected' : ''}">
							<button
								class="ch-inbox-row__btn"
								aria-label="Email from {email.from_name || email.from_email}: {email.subject || 'no subject'}"
								onclick={() => handleSelectEmail(email)}
							>
							{#if !email.is_read}
								<div class="ch-inbox-unread-dot"></div>
							{:else}
								<div class="ch-inbox-unread-dot ch-inbox-unread-dot--read"></div>
							{/if}
							<div class="ch-inbox-avatar">
								{(email.from_name || email.from_email).charAt(0).toUpperCase()}
							</div>
							<div class="ch-inbox-sender">{email.from_name || email.from_email}</div>
							<div class="ch-inbox-content">
								<span class="ch-inbox-subject">{email.subject || '(no subject)'}</span>
								<span class="ch-inbox-preview"> — {email.snippet || ''}</span>
							</div>
							<time class="ch-inbox-time">{formatDate(email.date)}</time>
							</button>
						</li>
					{/each}
				{/if}
			</ul>
		</div>

		<!-- Email Preview Panel -->
		<div class="ch-inbox-preview-panel">
			{#if selectedEmail}
				<div class="ch-inbox-preview__header">
					<h2 class="ch-inbox-preview__subject">{selectedEmail.subject || '(no subject)'}</h2>
					<div class="ch-inbox-preview__sender">
						<div class="ch-inbox-avatar ch-inbox-avatar--lg">
							{(selectedEmail.from_name || selectedEmail.from_email).charAt(0).toUpperCase()}
						</div>
						<div class="ch-inbox-preview__sender-info">
							<p class="ch-inbox-preview__sender-name">{selectedEmail.from_name || selectedEmail.from_email}</p>
							<p class="ch-inbox-preview__sender-email">{selectedEmail.from_email}</p>
							<time class="ch-inbox-preview__date">{new Date(selectedEmail.date).toLocaleString()}</time>
						</div>
					</div>
				</div>

				<div class="ch-inbox-preview__body">
					{#if selectedEmail.body_html}
						<div class="ch-inbox-preview__html">
							{@html DOMPurify.sanitize(selectedEmail.body_html, { ALLOWED_TAGS: ['p', 'br', 'b', 'i', 'u', 'strong', 'em', 'a', 'ul', 'ol', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'pre', 'code', 'span', 'div', 'table', 'thead', 'tbody', 'tr', 'td', 'th', 'img'], ALLOWED_ATTR: ['href', 'src', 'alt', 'class', 'style', 'target', 'rel'], ALLOW_DATA_ATTR: false })}
						</div>
					{:else if selectedEmail.body_text}
						<pre class="ch-inbox-preview__text">{selectedEmail.body_text}</pre>
					{:else}
						<p class="ch-inbox-preview__empty-text">No content</p>
					{/if}
				</div>

				<div class="ch-inbox-preview__actions">
					<button
						onclick={() => selectedEmail && openReply(selectedEmail)}
						class="btn-compact btn-compact-ghost"
						aria-label="Reply to email"
					>
						<Reply size={15} />
						Reply
					</button>
					<button
						onclick={() => selectedEmail && openForward(selectedEmail)}
						class="btn-compact btn-compact-ghost"
						aria-label="Forward email"
					>
						<Forward size={15} />
						Forward
					</button>
					<button class="btn-compact btn-compact-ghost" aria-label="Archive email">
						<Archive size={15} />
						Archive
					</button>
				</div>
			{:else}
				<div class="ch-inbox-empty ch-inbox-empty--inline">
					<Mail size={40} strokeWidth={1.2} />
					<p>Select an email to read</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Compose Modal -->
{#if showComposeModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="ch-compose-overlay" onclick={() => { resetComposeForm(); showComposeModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') { resetComposeForm(); showComposeModal = false; } }} role="dialog" tabindex="-1" aria-modal="true" aria-label="Compose email">
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div class="ch-compose" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="document">
			<div class="ch-compose-header">
				<span class="ch-compose-header__title">{replyTo ? 'Reply' : 'New Message'}</span>
				<div class="ch-compose-header__actions">
					<button
						onclick={() => { resetComposeForm(); showComposeModal = false; }}
						class="btn-compact btn-compact-ghost btn-compact-icon"
						aria-label="Close compose dialog"
					>
						<X size={16} />
					</button>
				</div>
			</div>

			{#if composeError}
				<div class="ch-compose-error">
					<p>{composeError}</p>
				</div>
			{/if}

			<div class="ch-compose-field">
				<label class="ch-compose-label" for="ch-to">To</label>
				<input
					id="ch-to"
					type="text"
					bind:value={composeTo}
					placeholder="recipient@example.com"
					class="ch-compose-tags__input"
				/>
			</div>
			<div class="ch-compose-field">
				<label class="ch-compose-label" for="ch-cc">Cc</label>
				<input
					id="ch-cc"
					type="text"
					bind:value={composeCc}
					placeholder="cc@example.com (optional)"
					class="ch-compose-tags__input"
				/>
			</div>
			<div class="ch-compose-field">
				<label class="ch-compose-label" for="ch-subject">Subject</label>
				<input
					id="ch-subject"
					type="text"
					bind:value={composeSubject}
					placeholder="Email subject"
					class="ch-compose-subject"
				/>
			</div>

			<textarea
				bind:value={composeBody}
				rows="12"
				placeholder="Write your message..."
				class="ch-compose-body"
				aria-label="Email body"
			></textarea>

			<div class="ch-compose-actions">
				<button
					onclick={handleSendEmail}
					disabled={isSending || !composeTo.trim()}
					class="btn-pill btn-pill-primary btn-pill-sm"
					aria-label="Send email"
				>
					{#if isSending}
						<Loader2 size={15} class="ch-spin" />
						Sending...
					{:else}
						<Send size={15} />
						Send
					{/if}
				</button>
				<button
					class="btn-compact btn-compact-ghost btn-compact-icon"
					aria-label="Attach file"
				>
					<Paperclip size={16} />
				</button>
				<div class="ch-compose-actions__spacer"></div>
				<button
					onclick={() => { resetComposeForm(); showComposeModal = false; }}
					class="btn-compact btn-compact-ghost"
					aria-label="Discard draft"
				>
					Discard
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ===== EMAIL INBOX LAYOUT ===== */
	.ch-inbox-layout {
		display: flex;
		height: 100%;
		overflow: hidden;
		background: var(--dbg);
	}

	/* ===== SIDEBAR ===== */
	.ch-inbox-sidebar {
		width: 185px;
		background: var(--dbg2);
		border-right: 1px solid var(--dbd);
		padding: 12px 0;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
	}

	.ch-inbox-compose-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		width: calc(100% - 24px);
		margin: 0 12px 12px;
		justify-content: center;
	}

	.ch-inbox-folders {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.ch-inbox-folder {
		list-style: none;
	}

	.ch-inbox-folder__btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 14px;
		font-size: 0.81rem;
		color: var(--dt2);
		cursor: pointer;
		transition: background 0.15s;
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		font-family: inherit;
	}

	.ch-inbox-folder__btn:hover {
		background: var(--dbg3);
	}

	.ch-inbox-folder--active .ch-inbox-folder__btn {
		background: var(--dbg3);
		color: var(--dt);
		font-weight: 600;
	}

	.ch-inbox-folder__icon {
		width: 18px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.ch-inbox-folder__label {
		flex: 1;
	}

	.ch-inbox-folder__count {
		background: var(--accent-blue, #3b82f6);
		color: #fff;
		font-size: 0.68rem;
		font-weight: 700;
		padding: 1px 6px;
		border-radius: 10px;
		min-width: 18px;
		text-align: center;
	}

	.ch-inbox-sidebar__footer {
		margin-top: auto;
		padding: 12px 14px;
		border-top: 1px solid var(--dbd);
	}

	.ch-inbox-sidebar__stat {
		font-size: 0.7rem;
		color: var(--dt3);
		margin: 0;
	}

	/* ===== MAIN LIST ===== */
	.ch-inbox-main {
		width: 320px;
		border-right: 1px solid var(--dbd);
		display: flex;
		flex-direction: column;
		min-width: 0;
		flex-shrink: 0;
	}

	.ch-inbox-toolbar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 14px;
		border-bottom: 1px solid var(--dbd);
		background: var(--dbg);
	}

	.ch-inbox-toolbar__title {
		font-size: 0.88rem;
		font-weight: 600;
		color: var(--dt);
		text-transform: capitalize;
		margin: 0;
	}

	.ch-inbox-toolbar__spacer {
		flex: 1;
	}

	.ch-inbox-search {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		border-bottom: 1px solid var(--dbd);
	}

	:global(.ch-inbox-search__icon) {
		color: var(--dt3);
		flex-shrink: 0;
	}

	.ch-inbox-search__input {
		flex: 1;
		border: none;
		background: transparent;
		font-size: 0.8rem;
		color: var(--dt);
		outline: none;
		padding: 2px 0;
	}

	.ch-inbox-search__input::placeholder {
		color: var(--dt4);
	}

	.ch-inbox-list {
		list-style: none;
		margin: 0;
		padding: 0;
		flex: 1;
		overflow-y: auto;
	}

	.ch-inbox-row {
		border-bottom: 1px solid var(--dbd2);
	}

	.ch-inbox-row__btn {
		display: flex;
		align-items: center;
		gap: 9px;
		padding: 10px 12px;
		cursor: pointer;
		transition: background 0.15s;
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		font-family: inherit;
	}

	.ch-inbox-row__btn:hover {
		background: var(--dbg2);
	}

	.ch-inbox-row--selected .ch-inbox-row__btn {
		background: var(--dbg3);
	}

	.ch-inbox-row--unread .ch-inbox-sender,
	.ch-inbox-row--unread .ch-inbox-subject {
		font-weight: 700;
		color: var(--dt);
	}

	.ch-inbox-unread-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--accent-blue, #3b82f6);
		flex-shrink: 0;
	}

	.ch-inbox-unread-dot--read {
		background: transparent;
	}

	.ch-inbox-avatar {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		background: var(--dt3);
		color: var(--dbg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.66rem;
		font-weight: 700;
		flex-shrink: 0;
	}

	.ch-inbox-avatar--lg {
		width: 40px;
		height: 40px;
		font-size: 0.82rem;
	}

	.ch-inbox-sender {
		width: 100px;
		font-size: 0.8rem;
		color: var(--dt2);
		flex-shrink: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.ch-inbox-content {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: 0.78rem;
	}

	.ch-inbox-subject {
		color: var(--dt2);
	}

	.ch-inbox-preview {
		color: var(--dt3);
	}

	.ch-inbox-time {
		font-size: 0.7rem;
		color: var(--dt3);
		flex-shrink: 0;
		white-space: nowrap;
	}

	/* ===== LOADING / ERROR ===== */
	.ch-inbox-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 16px;
		color: var(--dt3);
		gap: 8px;
	}

	.ch-inbox-loading__text {
		font-size: 0.8rem;
	}

	.ch-inbox-error {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 24px 16px;
		gap: 8px;
	}

	.ch-inbox-error__text {
		font-size: 0.8rem;
		color: var(--color-error, #e05252);
		margin: 0;
	}

	/* ===== PREVIEW PANEL ===== */
	.ch-inbox-preview-panel {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		background: var(--dbg);
	}

	.ch-inbox-preview__header {
		padding: 20px 24px 16px;
		border-bottom: 1px solid var(--dbd);
	}

	.ch-inbox-preview__subject {
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0 0 12px;
	}

	.ch-inbox-preview__sender {
		display: flex;
		align-items: flex-start;
		gap: 12px;
	}

	.ch-inbox-preview__sender-info {
		display: flex;
		flex-direction: column;
		gap: 1px;
	}

	.ch-inbox-preview__sender-name {
		font-size: 0.88rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ch-inbox-preview__sender-email {
		font-size: 0.78rem;
		color: var(--dt3);
		margin: 0;
	}

	.ch-inbox-preview__date {
		font-size: 0.7rem;
		color: var(--dt4);
		margin-top: 2px;
	}

	.ch-inbox-preview__body {
		flex: 1;
		overflow-y: auto;
		padding: 20px 24px;
	}

	.ch-inbox-preview__html {
		font-size: 0.88rem;
		color: var(--dt2);
		line-height: 1.65;
	}

	.ch-inbox-preview__html :global(a) {
		color: var(--accent-blue, #3b82f6);
	}

	.ch-inbox-preview__text {
		white-space: pre-wrap;
		font-size: 0.85rem;
		color: var(--dt2);
		font-family: inherit;
		line-height: 1.6;
		margin: 0;
	}

	.ch-inbox-preview__empty-text {
		color: var(--dt4);
		font-style: italic;
		font-size: 0.85rem;
	}

	.ch-inbox-preview__actions {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 24px;
		border-top: 1px solid var(--dbd);
	}

	.ch-inbox-preview__actions :global(button) {
		display: flex;
		align-items: center;
		gap: 5px;
	}

	/* ===== EMPTY STATE ===== */
	.ch-inbox-empty {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		padding: 32px;
		text-align: center;
		color: var(--dt3);
	}

	.ch-inbox-empty--inline {
		color: var(--dt4);
	}

	.ch-inbox-empty__icon {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		background: var(--dbg2);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--dt3);
		margin-bottom: 8px;
	}

	.ch-inbox-empty__title {
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ch-inbox-empty__desc {
		font-size: 0.85rem;
		color: var(--dt3);
		max-width: 360px;
		margin: 0 0 8px;
		line-height: 1.5;
	}

	.ch-inbox-empty__upgrade {
		font-size: 0.8rem;
		color: var(--color-warning, #f59e0b);
		margin: 0 0 4px;
	}

	/* ===== COMPOSE MODAL ===== */
	.ch-compose-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
	}

	.ch-compose {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
		width: 100%;
		max-width: 640px;
		margin: 16px;
	}

	.ch-compose-header {
		display: flex;
		align-items: center;
		padding: 10px 14px;
		background: var(--dbg2);
		border-bottom: 1px solid var(--dbd);
	}

	.ch-compose-header__title {
		flex: 1;
		font-size: 0.88rem;
		font-weight: 600;
		color: var(--dt);
	}

	.ch-compose-header__actions {
		display: flex;
		gap: 4px;
	}

	.ch-compose-error {
		padding: 8px 14px;
		background: rgba(224, 82, 82, 0.1);
		border-bottom: 1px solid rgba(224, 82, 82, 0.2);
		font-size: 0.8rem;
	}

	.ch-compose-error p {
		margin: 0;
		color: var(--color-error, #e05252);
	}

	.ch-compose-field {
		display: flex;
		align-items: center;
		padding: 8px 14px;
		border-bottom: 1px solid var(--dbd);
		gap: 8px;
	}

	.ch-compose-label {
		font-size: 0.76rem;
		color: var(--dt3);
		width: 46px;
		flex-shrink: 0;
		font-weight: 500;
	}

	.ch-compose-tags__input {
		flex: 1;
		border: none;
		background: transparent;
		font-size: 0.83rem;
		color: var(--dt);
		outline: none;
		padding: 4px 0;
	}

	.ch-compose-tags__input::placeholder {
		color: var(--dt4);
	}

	.ch-compose-subject {
		flex: 1;
		border: none;
		background: transparent;
		font-size: 0.85rem;
		color: var(--dt);
		font-weight: 600;
		outline: none;
		padding: 4px 0;
	}

	.ch-compose-subject::placeholder {
		color: var(--dt4);
		font-weight: 400;
	}

	.ch-compose-body {
		width: 100%;
		border: none;
		padding: 14px;
		font-size: 0.83rem;
		color: var(--dt);
		background: var(--dbg);
		resize: vertical;
		outline: none;
		font-family: inherit;
		line-height: 1.65;
		box-sizing: border-box;
	}

	.ch-compose-body::placeholder {
		color: var(--dt4);
	}

	.ch-compose-actions {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-top: 1px solid var(--dbd);
		background: var(--dbg2);
	}

	.ch-compose-actions :global(button) {
		display: flex;
		align-items: center;
		gap: 5px;
	}

	.ch-compose-actions__spacer {
		flex: 1;
	}

	/* ===== SPIN ANIMATION ===== */
	:global(.ch-spin) {
		animation: ch-spin 1s linear infinite;
	}

	@keyframes ch-spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
