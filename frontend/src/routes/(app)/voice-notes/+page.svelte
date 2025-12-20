<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	interface VoiceNote {
		id: string;
		transcript: string;
		duration: number;
		word_count: number;
		words_per_minute: number;
		language?: string;
		created_at: string;
		context_id?: string;
		project_id?: string;
	}

	interface VoiceNoteStats {
		total_notes: number;
		total_duration_seconds: number;
		total_words: number;
		avg_words_per_minute: number;
	}

	// State
	let voiceNotes = $state<VoiceNote[]>([]);
	let stats = $state<VoiceNoteStats | null>(null);
	let isLoading = $state(true);
	let activeTab = $state<'date' | 'project'>('date');
	let expandedNotes = $state<Set<string>>(new Set());
	let copiedId = $state<string | null>(null);

	onMount(async () => {
		await loadVoiceNotes();
	});

	async function loadVoiceNotes() {
		isLoading = true;
		try {
			const [notesData, statsData] = await Promise.all([
				fetchVoiceNotes(),
				fetchVoiceNoteStats()
			]);
			voiceNotes = notesData;
			stats = statsData;
		} catch (error) {
			console.error('Error loading voice notes:', error);
		} finally {
			isLoading = false;
		}
	}

	async function fetchVoiceNotes(): Promise<VoiceNote[]> {
		const response = await fetch('/api/voice-notes', {
			credentials: 'include'
		});
		if (!response.ok) return [];
		return response.json();
	}

	async function fetchVoiceNoteStats(): Promise<VoiceNoteStats | null> {
		const response = await fetch('/api/voice-notes/stats', {
			credentials: 'include'
		});
		if (!response.ok) return null;
		return response.json();
	}

	// Group notes by date
	function groupByDate(notes: VoiceNote[]): Map<string, VoiceNote[]> {
		const groups = new Map<string, VoiceNote[]>();
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const yesterday = new Date(today);
		yesterday.setDate(yesterday.getDate() - 1);

		for (const note of notes) {
			const noteDate = new Date(note.created_at);
			noteDate.setHours(0, 0, 0, 0);

			let key: string;
			if (noteDate.getTime() === today.getTime()) {
				key = 'TODAY';
			} else if (noteDate.getTime() === yesterday.getTime()) {
				key = 'YESTERDAY';
			} else {
				key = noteDate.toLocaleDateString('en-US', {
					weekday: 'long',
					month: 'short',
					day: 'numeric'
				});
			}

			if (!groups.has(key)) {
				groups.set(key, []);
			}
			groups.get(key)!.push(note);
		}

		return groups;
	}

	// Group notes by project
	function groupByProject(notes: VoiceNote[]): Map<string, VoiceNote[]> {
		const groups = new Map<string, VoiceNote[]>();

		for (const note of notes) {
			const key = note.project_id || 'NO_PROJECT';
			if (!groups.has(key)) {
				groups.set(key, []);
			}
			groups.get(key)!.push(note);
		}

		return groups;
	}

	function toggleExpand(noteId: string) {
		const newSet = new Set(expandedNotes);
		if (newSet.has(noteId)) {
			newSet.delete(noteId);
		} else {
			newSet.add(noteId);
		}
		expandedNotes = newSet;
	}

	async function copyTranscript(note: VoiceNote) {
		try {
			await navigator.clipboard.writeText(note.transcript);
			copiedId = note.id;
			setTimeout(() => {
				copiedId = null;
			}, 2000);
		} catch (error) {
			console.error('Failed to copy:', error);
		}
	}

	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		if (mins > 0) {
			return `${mins}m ${secs}s`;
		}
		return `${secs}s`;
	}

	function formatTotalDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const mins = Math.floor((seconds % 3600) / 60);
		if (hours > 0) {
			return `${hours}h ${mins}m`;
		}
		return `${mins}m`;
	}

	function formatTime(dateStr: string): string {
		return new Date(dateStr).toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	function truncateText(text: string, maxLength: number = 150): string {
		if (text.length <= maxLength) return text;
		return text.slice(0, maxLength).trim() + '...';
	}

	let groupedByDate = $derived(groupByDate(voiceNotes));
	let groupedByProject = $derived(groupByProject(voiceNotes));
</script>

<div class="voice-notes-page">
	<!-- Page Header -->
	<div class="page-header">
		<div class="header-content">
			<div class="header-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
					<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
					<line x1="12" y1="19" x2="12" y2="23"/>
					<line x1="8" y1="23" x2="16" y2="23"/>
				</svg>
			</div>
			<div>
				<h1 class="page-title">Voice Notes</h1>
				<p class="page-subtitle">Your voice transcription history</p>
			</div>
		</div>
		<div class="tab-selector">
			<button
				onclick={() => activeTab = 'date'}
				class="tab-btn"
				class:active={activeTab === 'date'}
			>
				By Date
			</button>
			<button
				onclick={() => activeTab = 'project'}
				class="tab-btn"
				class:active={activeTab === 'project'}
			>
				By Project
			</button>
		</div>
	</div>

	{#if isLoading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Loading voice notes...</p>
		</div>
	{:else if voiceNotes.length === 0}
		<!-- Empty State -->
		<div class="empty-state">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
					<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
					<line x1="12" y1="19" x2="12" y2="23"/>
					<line x1="8" y1="23" x2="16" y2="23"/>
				</svg>
			</div>
			<h3>No Voice Notes Yet</h3>
			<p>Start recording voice notes in the chat to see them here.</p>
		</div>
	{:else}
		<!-- Stats Grid -->
		{#if stats}
			<div class="stats-grid">
				<div class="stat-card notes">
					<div class="stat-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
						</svg>
					</div>
					<div class="stat-content">
						<span class="stat-value">{stats.total_notes}</span>
						<span class="stat-label">Notes</span>
					</div>
				</div>

				<div class="stat-card duration">
					<div class="stat-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"/>
							<polyline points="12 6 12 12 16 14"/>
						</svg>
					</div>
					<div class="stat-content">
						<span class="stat-value">{formatTotalDuration(stats.total_duration_seconds)}</span>
						<span class="stat-label">Total Duration</span>
					</div>
				</div>

				<div class="stat-card words">
					<div class="stat-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
							<polyline points="14 2 14 8 20 8"/>
							<line x1="16" y1="13" x2="8" y2="13"/>
							<line x1="16" y1="17" x2="8" y2="17"/>
						</svg>
					</div>
					<div class="stat-content">
						<span class="stat-value">{stats.total_words.toLocaleString()}</span>
						<span class="stat-label">Total Words</span>
					</div>
				</div>

				<div class="stat-card wpm">
					<div class="stat-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
						</svg>
					</div>
					<div class="stat-content">
						<span class="stat-value">{Math.round(stats.avg_words_per_minute)}</span>
						<span class="stat-label">Avg WPM</span>
					</div>
				</div>
			</div>
		{/if}

		<!-- Notes List -->
		<div class="notes-container">
			{#if activeTab === 'date'}
				{#each groupedByDate as [dateLabel, notes]}
					<div class="date-group">
						<h3 class="group-label">{dateLabel}</h3>
						<div class="notes-list">
							{#each notes as note}
								{@const isExpanded = expandedNotes.has(note.id)}
								<div class="note-card" class:expanded={isExpanded}>
									<button class="note-main" onclick={() => toggleExpand(note.id)}>
										<div class="note-time">{formatTime(note.created_at)}</div>
										<div class="note-content">
											<p class="note-transcript">
												{isExpanded ? note.transcript : truncateText(note.transcript)}
											</p>
										</div>
										<div class="note-meta">
											<span class="meta-item">{formatDuration(note.duration)}</span>
											<span class="meta-sep"></span>
											<span class="meta-item">{note.word_count} words</span>
											<span class="meta-sep"></span>
											<span class="meta-item">{Math.round(note.words_per_minute)} WPM</span>
										</div>
									</button>
									{#if isExpanded}
										<div class="note-actions">
											<button class="action-btn copy" onclick={() => copyTranscript(note)}>
												{#if copiedId === note.id}
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<polyline points="20 6 9 17 4 12"/>
													</svg>
													Copied!
												{:else}
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
														<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
													</svg>
													Copy
												{/if}
											</button>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/each}
			{:else}
				{#each groupedByProject as [projectId, notes]}
					<div class="date-group">
						<h3 class="group-label">
							{projectId === 'NO_PROJECT' ? 'No Project' : `Project ${projectId.slice(0, 8)}`}
						</h3>
						<div class="notes-list">
							{#each notes as note}
								{@const isExpanded = expandedNotes.has(note.id)}
								<div class="note-card" class:expanded={isExpanded}>
									<button class="note-main" onclick={() => toggleExpand(note.id)}>
										<div class="note-time">{formatTime(note.created_at)}</div>
										<div class="note-content">
											<p class="note-transcript">
												{isExpanded ? note.transcript : truncateText(note.transcript)}
											</p>
										</div>
										<div class="note-meta">
											<span class="meta-item">{formatDuration(note.duration)}</span>
											<span class="meta-sep"></span>
											<span class="meta-item">{note.word_count} words</span>
										</div>
									</button>
									{#if isExpanded}
										<div class="note-actions">
											<button class="action-btn copy" onclick={() => copyTranscript(note)}>
												{#if copiedId === note.id}
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<polyline points="20 6 9 17 4 12"/>
													</svg>
													Copied!
												{:else}
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
														<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
													</svg>
													Copy
												{/if}
											</button>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.voice-notes-page {
		padding: 24px;
		max-width: 900px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		gap: 20px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--color-border, #e5e7eb);
	}

	:global(.dark) .page-header {
		border-color: rgba(255, 255, 255, 0.08);
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.header-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		background: linear-gradient(135deg, #10b981, #34d399);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.header-icon svg {
		width: 24px;
		height: 24px;
		color: white;
	}

	.page-title {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		margin: 0;
	}

	:global(.dark) .page-title {
		color: #f9fafb;
	}

	.page-subtitle {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
		margin: 4px 0 0;
	}

	:global(.dark) .page-subtitle {
		color: #9ca3af;
	}

	/* Tab Selector */
	.tab-selector {
		display: flex;
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 12px;
		padding: 4px;
		gap: 4px;
	}

	:global(.dark) .tab-selector {
		background: #1f1f1f;
	}

	.tab-btn {
		padding: 10px 20px;
		border: none;
		background: transparent;
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text-secondary, #6b7280);
		cursor: pointer;
		transition: all 0.15s;
	}

	:global(.dark) .tab-btn {
		color: #9ca3af;
	}

	.tab-btn:hover {
		color: var(--color-text, #111827);
	}

	:global(.dark) .tab-btn:hover {
		color: #f9fafb;
	}

	.tab-btn.active {
		background: white;
		color: var(--color-text, #111827);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	:global(.dark) .tab-btn.active {
		background: #2c2c2e;
		color: #f9fafb;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}

	/* Loading State */
	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 120px 20px;
		gap: 16px;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .loading-state {
		color: #9ca3af;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--color-border, #e5e7eb);
		border-top-color: #10b981;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	:global(.dark) .loading-spinner {
		border-color: rgba(255, 255, 255, 0.1);
		border-top-color: #10b981;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 120px 20px;
		text-align: center;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		border-radius: 20px;
		background: var(--color-bg-secondary, #f3f4f6);
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 20px;
	}

	:global(.dark) .empty-icon {
		background: #1f1f1f;
	}

	.empty-icon svg {
		width: 40px;
		height: 40px;
		color: var(--color-text-muted, #9ca3af);
	}

	:global(.dark) .empty-icon svg {
		color: #6b7280;
	}

	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--color-text, #111827);
		margin: 0 0 8px;
	}

	:global(.dark) .empty-state h3 {
		color: #f9fafb;
	}

	.empty-state p {
		font-size: 0.875rem;
		color: var(--color-text-secondary, #6b7280);
		margin: 0;
	}

	:global(.dark) .empty-state p {
		color: #9ca3af;
	}

	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	@media (max-width: 768px) {
		.stats-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 480px) {
		.stats-grid { grid-template-columns: 1fr; }
	}

	.stat-card {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 16px;
		padding: 20px;
		display: flex;
		align-items: center;
		gap: 14px;
	}

	:global(.dark) .stat-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.stat-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stat-icon svg {
		width: 24px;
		height: 24px;
	}

	.stat-card.notes .stat-icon { background: #d1fae5; color: #059669; }
	.stat-card.duration .stat-icon { background: #dbeafe; color: #2563eb; }
	.stat-card.words .stat-icon { background: #f3e8ff; color: #9333ea; }
	.stat-card.wpm .stat-icon { background: #fef3c7; color: #d97706; }

	:global(.dark) .stat-card.notes .stat-icon { background: rgba(5, 150, 105, 0.2); }
	:global(.dark) .stat-card.duration .stat-icon { background: rgba(37, 99, 235, 0.2); }
	:global(.dark) .stat-card.words .stat-icon { background: rgba(147, 51, 234, 0.2); }
	:global(.dark) .stat-card.wpm .stat-icon { background: rgba(217, 119, 6, 0.2); }

	.stat-content {
		display: flex;
		flex-direction: column;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-text, #111827);
		line-height: 1;
	}

	:global(.dark) .stat-value {
		color: #f9fafb;
	}

	.stat-label {
		font-size: 0.7rem;
		color: var(--color-text-muted, #9ca3af);
		margin-top: 4px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	:global(.dark) .stat-label {
		color: #6b7280;
	}

	/* Notes Container */
	.notes-container {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.date-group {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.group-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-text-muted, #9ca3af);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin: 0;
		padding-left: 4px;
	}

	:global(.dark) .group-label {
		color: #6b7280;
	}

	.notes-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.note-card {
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 12px;
		overflow: hidden;
		transition: border-color 0.15s;
	}

	:global(.dark) .note-card {
		background: #0a0a0a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.note-card:hover {
		border-color: #10b981;
	}

	.note-card.expanded {
		border-color: #10b981;
	}

	.note-main {
		width: 100%;
		padding: 16px;
		background: none;
		border: none;
		cursor: pointer;
		text-align: left;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.note-time {
		font-size: 0.75rem;
		font-weight: 500;
		color: #10b981;
	}

	.note-content {
		flex: 1;
	}

	.note-transcript {
		font-size: 0.9rem;
		color: var(--color-text, #111827);
		line-height: 1.5;
		margin: 0;
	}

	:global(.dark) .note-transcript {
		color: #f9fafb;
	}

	.note-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.75rem;
		color: var(--color-text-muted, #9ca3af);
	}

	:global(.dark) .note-meta {
		color: #6b7280;
	}

	.meta-sep {
		width: 3px;
		height: 3px;
		border-radius: 50%;
		background: currentColor;
		opacity: 0.5;
	}

	.note-actions {
		padding: 12px 16px;
		background: var(--color-bg-secondary, #f9fafb);
		border-top: 1px solid var(--color-border, #e5e7eb);
		display: flex;
		gap: 8px;
	}

	:global(.dark) .note-actions {
		background: #141414;
		border-color: rgba(255, 255, 255, 0.06);
	}

	.action-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		background: white;
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 8px;
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--color-text-secondary, #6b7280);
		cursor: pointer;
		transition: all 0.15s;
	}

	:global(.dark) .action-btn {
		background: #1f1f1f;
		border-color: rgba(255, 255, 255, 0.1);
		color: #9ca3af;
	}

	.action-btn:hover {
		background: var(--color-bg-secondary, #f3f4f6);
		color: var(--color-text, #111827);
	}

	:global(.dark) .action-btn:hover {
		background: #2c2c2e;
		color: #f9fafb;
	}

	.action-btn.copy {
		color: #10b981;
	}

	.action-btn svg {
		width: 14px;
		height: 14px;
	}
</style>
