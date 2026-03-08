<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { Hash, Lock, MessageSquare, Search, MoreVertical, Paperclip, Send, Loader2, Users, ThumbsUp, PartyPopper, Eye, RefreshCw } from 'lucide-svelte';

	// State
	let slackConnected = $state(false);
	let slackChannels = $state<any[]>([]);
	let selectedChannel = $state<any | null>(null);
	let isLoading = $state(true);
	let messageInput = $state('');
	let channelMessages = $state<any[]>([]);
	let isLoadingMessages = $state(false);
	let isSending = $state(false);
	let isSyncing = $state(false);

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

	async function loadMessages(channelId: string) {
		isLoadingMessages = true;
		channelMessages = [];
		try {
			const response = await api.getSlackMessages(channelId);
			channelMessages = response.messages ?? [];
		} catch (e) {
			console.error('Failed to load messages:', e);
		} finally {
			isLoadingMessages = false;
		}
	}

	async function handleSendMessage() {
		if (!messageInput.trim() || !selectedChannel) return;
		isSending = true;
		try {
			await api.sendSlackMessage(selectedChannel.id, messageInput.trim());
			messageInput = '';
			// Reload messages to show the new one
			await loadMessages(selectedChannel.id);
		} catch (e) {
			console.error('Failed to send message:', e);
		} finally {
			isSending = false;
		}
	}

	async function handleSyncChannels() {
		isSyncing = true;
		try {
			await api.syncSlackChannels();
			await loadChannels();
		} catch (e) {
			console.error('Failed to sync channels:', e);
		} finally {
			isSyncing = false;
		}
	}

	async function selectChannel(channel: any) {
		selectedChannel = channel;
		await loadMessages(channel.id);
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

	async function handleDisconnectSlack() {
		try {
			await api.disconnectSlack();
			slackConnected = false;
			slackChannels = [];
			selectedChannel = null;
			channelMessages = [];
		} catch (e) {
			console.error('Failed to disconnect Slack:', e);
		}
	}

	onMount(() => {
		checkSlackConnection();
	});

	// Preview channels for not-connected state
	const previewChannels = [
		{ name: 'general', member_count: 42 },
		{ name: 'engineering', member_count: 15 },
		{ name: 'design', member_count: 8 },
	];
</script>

<div class="ch-chat-layout">
	{#if isLoading}
		<div class="ch-chat-empty">
			<Loader2 size={28} class="ch-spin" />
		</div>
	{:else if !slackConnected}
		<!-- Slack Not Connected -->
		<div class="ch-chat-empty">
			<div class="ch-chat-empty__icon">
				<MessageSquare size={32} />
			</div>
			<h2 class="ch-chat-empty__title">Connect Slack</h2>
			<p class="ch-chat-empty__desc">
				Connect your Slack workspace to view channels and messages directly from BusinessOS.
			</p>
			<button
				onclick={handleConnectSlack}
				class="btn-pill btn-pill-primary"
				aria-label="Connect Slack workspace"
			>
				<MessageSquare size={18} />
				Connect Slack
			</button>

			<!-- Preview -->
			<div class="ch-chat-preview">
				<p class="ch-chat-preview__label">Preview of what you'll see:</p>
				<div class="ch-chat-preview__list">
					{#each previewChannels as channel}
						<div class="ch-chat-preview__item">
							<Hash size={14} />
							<span class="ch-chat-preview__name">{channel.name}</span>
							<span class="ch-chat-preview__members">{channel.member_count} members</span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{:else}
		<!-- Channel Sidebar -->
		<aside class="ch-chat-sidebar">
			<div class="ch-chat-sidebar__header">
				<h3 class="ch-chat-sidebar__title">Channels</h3>
				<span class="ch-chat-sidebar__count">
					{slackChannels.length}
				</span>
				<button
					class="btn-compact btn-compact-ghost btn-compact-icon"
					aria-label="Sync channels"
					onclick={handleSyncChannels}
					disabled={isSyncing}
				>
					<RefreshCw size={14} class={isSyncing ? 'ch-spin' : ''} />
				</button>
			</div>

			{#if slackChannels.length === 0}
				<div class="ch-chat-sidebar__empty">
					<p>No channels synced yet.</p>
					<button class="btn-pill btn-pill-secondary btn-pill-sm" onclick={handleSyncChannels} disabled={isSyncing}>
						{isSyncing ? 'Syncing...' : 'Sync Channels'}
					</button>
				</div>
			{:else}
				<ul class="ch-chat-channels">
					{#each slackChannels as channel}
						<li class="ch-chat-channel {selectedChannel?.id === channel.id ? 'ch-chat-channel--active' : ''}">
							<button
								class="ch-chat-channel__btn"
								aria-label="{channel.is_private ? 'Private' : ''} channel {channel.name}"
								onclick={() => { selectChannel(channel); }}
							>
								<span class="ch-chat-channel__hash">
									{#if channel.is_private}
										<Lock size={13} />
									{:else}
										<Hash size={13} />
									{/if}
								</span>
								<span class="ch-chat-channel__name">{channel.name}</span>
							</button>
						</li>
					{/each}
				</ul>
			{/if}

			<div class="ch-chat-sidebar__footer">
				<button class="btn-compact btn-compact-ghost" onclick={handleDisconnectSlack} aria-label="Disconnect Slack">
					Disconnect
				</button>
			</div>
		</aside>

		<!-- Channel Main -->
		<div class="ch-chat-main">
			{#if selectedChannel}
				<div class="ch-chat-header">
					<div class="ch-chat-header__info">
						<h2 class="ch-chat-header__title">
							<span class="ch-chat-channel__hash">
								{#if selectedChannel.is_private}
									<Lock size={14} />
								{:else}
									<Hash size={14} />
								{/if}
							</span>
							{selectedChannel.name}
						</h2>
						<span class="ch-chat-header__members">
							<Users size={13} />
							{selectedChannel.member_count} members
						</span>
					</div>
					<div class="ch-chat-header__actions">
						<button class="btn-compact btn-compact-ghost btn-compact-icon" aria-label="Search messages">
							<Search size={16} />
						</button>
						<button class="btn-compact btn-compact-ghost btn-compact-icon" aria-label="More options">
							<MoreVertical size={16} />
						</button>
					</div>
				</div>

				<!-- Messages -->
				<div class="ch-chat-messages">
					{#if isLoadingMessages}
						<div class="ch-chat-empty ch-chat-empty--inline">
							<Loader2 size={24} class="ch-spin" />
							<p>Loading messages...</p>
						</div>
					{:else if channelMessages.length === 0}
						<div class="ch-chat-empty ch-chat-empty--inline">
							<MessageSquare size={32} strokeWidth={1.2} />
							<p>No messages yet in this channel</p>
						</div>
					{:else}
						{#each channelMessages as message}
							<div class="ch-chat-msg">
								<div class="ch-chat-msg__avatar">
									{(message.sender || message.user || '?').charAt(0)}
								</div>
								<div class="ch-chat-msg__body">
									<div class="ch-chat-msg__meta">
										<span class="ch-chat-msg__sender">{message.sender || message.user || 'Unknown'}</span>
										<time class="ch-chat-msg__time">{message.time || message.timestamp || ''}</time>
									</div>
									<p class="ch-chat-msg__text">{message.content || message.text || ''}</p>
									{#if message.reactions?.length > 0}
										<div class="ch-chat-msg__reactions">
											{#each message.reactions as reaction}
												<span class="ch-chat-msg__reaction">
													{#if reaction.icon === 'thumbs-up'}
														<ThumbsUp size={12} />
													{:else if reaction.icon === 'party'}
														<PartyPopper size={12} />
													{:else if reaction.icon === 'eye'}
														<Eye size={12} />
													{/if}
													<span>{reaction.count}</span>
												</span>
											{/each}
										</div>
									{/if}
								</div>
							</div>
						{/each}
					{/if}
				</div>

				<!-- Message Input -->
				<div class="ch-chat-compose">
					<div class="ch-chat-input-row">
						<input
							type="text"
							bind:value={messageInput}
							placeholder="Message #{selectedChannel.name}"
							class="ch-chat-input"
							aria-label="Type a message"
							onkeydown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSendMessage(); } }}
						/>
						<button class="btn-compact btn-compact-ghost btn-compact-icon" aria-label="Attach file">
							<Paperclip size={16} />
						</button>
						<button
							class="btn-pill btn-pill-primary btn-pill-sm"
							disabled={!messageInput.trim() || isSending}
							onclick={handleSendMessage}
							aria-label="Send message"
						>
							<Send size={14} />
						</button>
					</div>
				</div>
			{:else}
				<div class="ch-chat-empty ch-chat-empty--inline">
					<MessageSquare size={40} strokeWidth={1.2} />
					<p>Select a channel to view messages</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	/* ===== CHAT LAYOUT ===== */
	.ch-chat-layout {
		display: flex;
		height: 100%;
		overflow: hidden;
		background: var(--dbg);
	}

	/* ===== SIDEBAR ===== */
	.ch-chat-sidebar {
		width: 220px;
		background: var(--dbg2);
		border-right: 1px solid var(--dbd);
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
	}

	.ch-chat-sidebar__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 16px;
		border-bottom: 1px solid var(--dbd);
	}

	.ch-chat-sidebar__title {
		font-size: 0.88rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ch-chat-sidebar__count {
		font-size: 0.7rem;
		color: var(--dt3);
	}

	.ch-chat-channels {
		list-style: none;
		margin: 0;
		padding: 6px 0;
		flex: 1;
		overflow-y: auto;
	}

	.ch-chat-channel {
		list-style: none;
	}

	.ch-chat-channel__btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 16px;
		font-size: 0.82rem;
		color: var(--dt2);
		cursor: pointer;
		transition: background 0.15s;
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		font-family: inherit;
	}

	.ch-chat-channel__btn:hover {
		background: var(--dbg3);
	}

	.ch-chat-channel--active .ch-chat-channel__btn {
		background: var(--dbg3);
		color: var(--dt);
		font-weight: 600;
	}

	.ch-chat-channel__hash {
		color: var(--dt3);
		display: flex;
		align-items: center;
		flex-shrink: 0;
	}

	.ch-chat-channel__name {
		flex: 1;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	/* ===== MAIN ===== */
	.ch-chat-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.ch-chat-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 20px;
		border-bottom: 1px solid var(--dbd);
	}

	.ch-chat-header__info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.ch-chat-header__title {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ch-chat-header__members {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.75rem;
		color: var(--dt3);
	}

	.ch-chat-header__actions {
		display: flex;
		gap: 4px;
	}

	/* ===== MESSAGES ===== */
	.ch-chat-messages {
		flex: 1;
		overflow-y: auto;
		padding: 16px 20px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.ch-chat-msg {
		display: flex;
		gap: 10px;
	}

	.ch-chat-msg__avatar {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: var(--accent-purple, #8b5cf6);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.78rem;
		font-weight: 700;
		flex-shrink: 0;
	}

	.ch-chat-msg__body {
		flex: 1;
		min-width: 0;
	}

	.ch-chat-msg__meta {
		display: flex;
		align-items: baseline;
		gap: 8px;
	}

	.ch-chat-msg__sender {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--dt);
	}

	.ch-chat-msg__time {
		font-size: 0.7rem;
		color: var(--dt4);
	}

	.ch-chat-msg__text {
		font-size: 0.85rem;
		color: var(--dt2);
		margin: 3px 0 0;
		line-height: 1.5;
	}

	.ch-chat-msg__reactions {
		display: flex;
		gap: 6px;
		margin-top: 6px;
	}

	.ch-chat-msg__reaction {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		border-radius: 12px;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		font-size: 0.72rem;
		color: var(--dt3);
	}

	/* ===== COMPOSE ===== */
	.ch-chat-compose {
		padding: 12px 20px;
		border-top: 1px solid var(--dbd);
	}

	.ch-chat-input-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.ch-chat-input {
		flex: 1;
		border: 1px solid var(--dbd);
		background: var(--dbg2);
		border-radius: 8px;
		padding: 8px 12px;
		font-size: 0.83rem;
		color: var(--dt);
		outline: none;
		transition: border-color 0.2s;
	}

	.ch-chat-input:focus {
		border-color: var(--accent-purple, #8b5cf6);
	}

	.ch-chat-input::placeholder {
		color: var(--dt4);
	}

	/* ===== EMPTY STATES ===== */
	.ch-chat-empty {
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

	.ch-chat-empty--inline {
		color: var(--dt4);
	}

	.ch-chat-empty__icon {
		width: 64px;
		height: 64px;
		border-radius: 16px;
		background: var(--dbg2);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--accent-purple, #8b5cf6);
		margin-bottom: 8px;
	}

	.ch-chat-empty__title {
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ch-chat-empty__desc {
		font-size: 0.85rem;
		color: var(--dt3);
		max-width: 360px;
		margin: 0 0 12px;
		line-height: 1.5;
	}

	/* ===== PREVIEW ===== */
	.ch-chat-preview {
		margin-top: 24px;
		padding-top: 20px;
		border-top: 1px solid var(--dbd);
		width: 100%;
		max-width: 360px;
	}

	.ch-chat-preview__label {
		font-size: 0.78rem;
		color: var(--dt4);
		margin: 0 0 10px;
	}

	.ch-chat-preview__list {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.ch-chat-preview__item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		background: var(--dbg2);
		border-radius: 8px;
		border: 1px solid var(--dbd2);
		color: var(--dt3);
	}

	.ch-chat-preview__name {
		font-size: 0.83rem;
		font-weight: 600;
		color: var(--dt);
		flex: 1;
		text-align: left;
	}

	.ch-chat-preview__members {
		font-size: 0.7rem;
		color: var(--dt4);
	}

	/* ===== SIDEBAR EXTRAS ===== */
	.ch-chat-sidebar__empty {
		padding: 20px 16px;
		text-align: center;
		color: var(--dt3);
		font-size: 0.82rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}

	.ch-chat-sidebar__footer {
		padding: 8px 16px;
		border-top: 1px solid var(--dbd);
		margin-top: auto;
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
