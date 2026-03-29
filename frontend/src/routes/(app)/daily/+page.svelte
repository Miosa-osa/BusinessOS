<script lang="ts">
	import { api, type DailyLog } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { analyticsNotes, type AnalyticsNote } from '$lib/stores/analyticsNotes';

	// ── State ─────────────────────────────────────────────────────────────────
	let todayEntry     = $state('');
	let energyLevel    = $state(7);
	let currentLog     = $state<DailyLog | null>(null);
	let pastLogs       = $state<DailyLog[]>([]);
	let isLoading      = $state(true);
	let isSaving       = $state(false);
	let saveSuccess    = $state(false);
	let saveError      = $state(false);
	let placeholderIdx = $state(0);
	let showHistory    = $state(false);
	let promptTimer: ReturnType<typeof setInterval>;

	// ── Rotating prompts ──────────────────────────────────────────────────────
	const PROMPTS = [
		'What moved the needle today?',
		'What are you grateful for?',
		'What challenged you most?',
		"What's still on your mind?",
		'What did you learn today?',
		'Any wins worth celebrating?',
		'What would make tomorrow better?',
		'What are you proud of today?',
	];

	// ── Derived ───────────────────────────────────────────────────────────────
	let sliderPercent = $derived(((energyLevel - 1) / 9) * 100);
	let wordCount     = $derived(todayEntry.trim() ? todayEntry.trim().split(/\s+/).length : 0);

	let streak = $derived(
		(() => {
			let count = currentLog ? 1 : 0;
			const today = new Date();
			for (let i = 0; i < pastLogs.length; i++) {
				const expected = new Date(today);
				expected.setDate(today.getDate() - (currentLog ? i + 1 : i));
				const d = (pastLogs[i]?.date ?? '').split('T')[0];
				if (!d) break;
				if (d === expected.toISOString().split('T')[0]) count++;
				else break;
			}
			return count;
		})()
	);

	// ── Analytics Notes ───────────────────────────────────────────────────────
	const todayDate = new Date().toISOString().split('T')[0];
	const todayAnalyticsNotes = $derived($analyticsNotes.filter(n => n.date === todayDate));
	let editingNoteId      = $state<string | null>(null);
	let editingNoteContent = $state('');

	// ── Lifecycle ─────────────────────────────────────────────────────────────
	onMount(async () => {
		await Promise.all([loadTodayLog(), loadPastLogs()]);
		isLoading = false;
		promptTimer = setInterval(() => {
			placeholderIdx = (placeholderIdx + 1) % PROMPTS.length;
		}, 5000);
		window.addEventListener('keydown', handleKeydown);
	});

	onDestroy(() => {
		clearInterval(promptTimer);
		window.removeEventListener('keydown', handleKeydown);
	});

	// ── API ───────────────────────────────────────────────────────────────────
	async function loadTodayLog() {
		try {
			const log = await api.getTodayLog();
			if (log) {
				currentLog = log;
				todayEntry = log.content;
				energyLevel = log.energy_level || 7;
			}
		} catch (e) { console.error('loadTodayLog:', e); }
	}

	async function loadPastLogs() {
		try {
			const logs = await api.getDailyLogs(0, 14);
			const today = new Date().toISOString().split('T')[0];
			pastLogs = logs.filter(l => l.date !== today);
		} catch (e) { console.error('loadPastLogs:', e); }
	}

	async function handleSave() {
		if (!todayEntry.trim() || isSaving) return;
		isSaving = true; saveSuccess = false; saveError = false;
		try {
			const log = await api.saveDailyLog({ content: todayEntry, energy_level: energyLevel });
			currentLog = log;
			saveSuccess = true;
			setTimeout(() => saveSuccess = false, 2200);
		} catch (e) {
			console.error('handleSave:', e);
			saveError = true;
			setTimeout(() => saveError = false, 2200);
		} finally {
			isSaving = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
			e.preventDefault();
			handleSave();
		}
	}

	// ── Auto-grow textarea action ──────────────────────────────────────────────
	function autoGrow(node: HTMLTextAreaElement) {
		function resize() {
			node.style.height = 'auto';
			node.style.height = Math.max(200, node.scrollHeight) + 'px';
		}
		node.addEventListener('input', resize);
		resize();
		return { destroy: () => node.removeEventListener('input', resize) };
	}

	// ── Analytics Notes helpers ───────────────────────────────────────────────
	function startEditNote(note: AnalyticsNote) {
		editingNoteId = note.id;
		editingNoteContent = note.content;
	}
	function saveEditNote() {
		if (editingNoteId && editingNoteContent.trim()) analyticsNotes.updateNote(editingNoteId, editingNoteContent);
		editingNoteId = null; editingNoteContent = '';
	}
	function cancelEditNote() { editingNoteId = null; editingNoteContent = ''; }
	function deleteAnalyticsNote(id: string) {
		analyticsNotes.deleteNote(id);
		if (editingNoteId === id) { editingNoteId = null; editingNoteContent = ''; }
	}

	// ── Date helpers ──────────────────────────────────────────────────────────
	const _now      = new Date();
	const _weekday  = _now.toLocaleDateString(undefined, { weekday: 'long' });
	const _dayMonth = _now.toLocaleDateString(undefined, { month: 'long', day: 'numeric' });
	const _year     = _now.getFullYear();

	function formatDate(s: string) {
		return new Date(s).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
	}
</script>

<div class="dlp">
	<!-- ── Header ── -->
	<div class="dlp-header">
		<div class="dlp-header__date-block">
			<span class="dlp-header__weekday">{_weekday}</span>
			<h1 class="dlp-header__day">{_dayMonth}<span class="dlp-header__year">{_year}</span></h1>
		</div>
		<div class="dlp-header__actions">
			{#if streak > 1}
				<div class="dlp-streak">
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
					</svg>
					<span class="dlp-streak__num">{streak}</span>
					<span class="dlp-streak__label">day streak</span>
				</div>
			{/if}
			{#if currentLog}
				<div class="dlp-saved-pill">
					<svg width="10" height="10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					{new Date(currentLog.updated_at).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })}
				</div>
			{/if}
			<button
				onclick={() => showHistory = !showHistory}
				class="btn-pill btn-pill-ghost btn-pill-sm"
				class:active={showHistory}
			>
				<svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
				</svg>
				History
			</button>
		</div>
	</div>

	{#if isLoading}
		<div class="dlp-loading"><div class="dlp-spinner"></div></div>
	{:else}
		<div class="dlp-body" class:dlp-body--with-history={showHistory}>

			<!-- ── Main: Entry form ── -->
			<div class="dlp-form">

				<!-- Energy card -->
				<div class="dlp-card dlp-energy-card">
					<div class="dlp-energy-header">
						<label for="energy-slider" class="dlp-label">Energy level</label>
						<div class="dlp-energy-readout">
							<span class="dlp-energy-num">{energyLevel}</span>
							<span class="dlp-energy-slash">/10</span>
						</div>
					</div>
					<input
						id="energy-slider"
						type="range" min="1" max="10"
						bind:value={energyLevel}
						class="dlp-energy__slider"
						style="--slider-pct: {sliderPercent}%"
					/>
					<div class="dlp-energy__row">
						<span class="dlp-energy__end-label">Low</span>
						<div class="dlp-energy__dots">
							{#each Array(10) as _, i}
								<span class="dlp-energy__dot" class:active={i < energyLevel}></span>
							{/each}
						</div>
						<span class="dlp-energy__end-label">High</span>
					</div>
				</div>

				<!-- Entry card -->
				<div class="dlp-card">
					<div class="dlp-entry-header">
						<label for="entry" class="dlp-label">What's on your mind?</label>
						{#if wordCount > 0}
							<span class="dlp-word-count">{wordCount} {wordCount === 1 ? 'word' : 'words'}</span>
						{/if}
					</div>
					<textarea
						id="entry"
						use:autoGrow
						bind:value={todayEntry}
						class="dlp-textarea"
						placeholder={PROMPTS[placeholderIdx]}
					></textarea>
				</div>

				<!-- Save button -->
				<button
					onclick={handleSave}
					disabled={isSaving || !todayEntry.trim()}
					class="btn-cta dlp-save"
					class:dlp-save--success={saveSuccess}
					class:dlp-save--error={saveError}
				>
					{#if isSaving}
						<svg class="dlp-save__spinner" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3"></circle>
							<path class="opacity-80" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Saving...
					{:else if saveSuccess}
						<svg width="15" height="15" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
						</svg>
						Saved
					{:else if saveError}
						Error — try again
					{:else}
						Save Entry
						<kbd class="dlp-save__kbd">⌘↵</kbd>
					{/if}
				</button>

				<!-- Analytics Notes -->
				<div class="dlp-analytics-notes">
					<div class="dlp-analytics-notes__header">
						<h3 class="dlp-analytics-notes__title">
							<svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
							Analytics Notes
						</h3>
						{#if todayAnalyticsNotes.length > 0}
						<span class="dlp-analytics-notes__count">{todayAnalyticsNotes.length}</span>
					{/if}
					</div>
					{#if todayAnalyticsNotes.length === 0}
						<p class="dlp-analytics-notes__subtitle">Click on chart data points in Analytics to add notes here</p>
					{:else}
					<div class="dlp-analytics-notes__list">
						{#each todayAnalyticsNotes as note (note.id)}
							<div class="dlp-analytics-note-card">
								<div class="dlp-analytics-note-card__metrics">
									<span class="dlp-analytics-note-card__metric">
										<span class="dlp-analytics-note-card__metric-label">Requests</span>
										<span class="dlp-analytics-note-card__metric-value">{note.metricContext.requests.toLocaleString()}</span>
									</span>
									<span class="dlp-analytics-note-card__metric">
										<span class="dlp-analytics-note-card__metric-label">Tokens</span>
										<span class="dlp-analytics-note-card__metric-value">{note.metricContext.tokens >= 1000 ? (note.metricContext.tokens / 1000).toFixed(1) + 'K' : note.metricContext.tokens}</span>
									</span>
									<span class="dlp-analytics-note-card__metric">
										<span class="dlp-analytics-note-card__metric-label">Cost</span>
										<span class="dlp-analytics-note-card__metric-value">${note.metricContext.cost.toFixed(2)}</span>
									</span>
									<span class="dlp-analytics-note-card__date">{note.metricContext.label}</span>
								</div>
								{#if editingNoteId === note.id}
									<textarea bind:value={editingNoteContent} class="dlp-analytics-note-card__edit-textarea" rows="3"></textarea>
									<div class="dlp-analytics-note-card__edit-actions">
										<button onclick={saveEditNote} class="dlp-analytics-note-card__edit-save">Save</button>
										<button onclick={cancelEditNote} class="dlp-analytics-note-card__edit-cancel">Cancel</button>
									</div>
								{:else}
									<p class="dlp-analytics-note-card__content">{note.content}</p>
									<div class="dlp-analytics-note-card__actions">
										<button onclick={() => startEditNote(note)} class="dlp-analytics-note-card__action-btn" aria-label="Edit note">
											<svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
											</svg>
										</button>
										<button onclick={() => deleteAnalyticsNote(note.id)} class="dlp-analytics-note-card__action-btn dlp-analytics-note-card__action-btn--delete" aria-label="Delete note">
											<svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									</div>
								{/if}
							</div>
						{/each}
					</div>
					{/if}
				</div>
			</div>

			<!-- ── Right: History panel (toggled) ── -->
			{#if showHistory}
			<aside class="dlp-history-panel">
				<div class="dlp-history-panel__head">
					<h2 class="dlp-history-panel__title">Recent Entries</h2>
					{#if pastLogs.length > 0}
						<span class="dlp-history-panel__count">{pastLogs.length}</span>
					{/if}
				</div>
				{#if pastLogs.length === 0}
					<div class="dlp-history-panel__empty">
						<svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
						</svg>
						<p>No past entries yet.<br/>Start writing today!</p>
					</div>
				{:else}
					<div class="dlp-history-panel__list">
						{#each pastLogs as log}
							<button
								onclick={() => { todayEntry = log.content; energyLevel = log.energy_level || 7; }}
								class="dlp-history-card"
							>
								<div class="dlp-history-card__top">
									<span class="dlp-history-card__date">{formatDate(log.date)}</span>
									{#if log.energy_level}
										<span class="dlp-history-card__energy">{log.energy_level}/10</span>
									{/if}
								</div>
								<p class="dlp-history-card__preview">{log.content}</p>
							</button>
						{/each}
					</div>
				{/if}
			</aside>
			{/if}

		</div>
	{/if}
</div>

<style>
	/* ── Layout ──────────────────────────────────────────────────────────────── */
	.dlp {
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* ── Header ──────────────────────────────────────────────────────────────── */
	.dlp-header {
		padding: 1rem 1.5rem 0.85rem;
		border-bottom: 1px solid var(--dbd2);
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		flex-shrink: 0;
		gap: 1rem;
	}

	.dlp-header__date-block {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.dlp-header__weekday {
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--dt4);
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.dlp-header__day {
		font-size: 1.45rem;
		font-weight: 700;
		color: var(--dt);
		margin: 0;
		line-height: 1.1;
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
	}

	.dlp-header__year {
		font-size: 1rem;
		font-weight: 400;
		color: var(--dt3);
	}

	.dlp-header__actions {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		flex-shrink: 0;
	}

	/* Streak badge — neutral gray */
	.dlp-streak {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.25rem 0.65rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 999px;
		color: var(--dt3);
	}

	.dlp-streak__num {
		font-size: 0.82rem;
		font-weight: 700;
		font-variant-numeric: tabular-nums;
		color: var(--dt);
	}

	.dlp-streak__label {
		font-size: 0.72rem;
		font-weight: 500;
	}

	/* Last saved pill */
	.dlp-saved-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.22rem 0.65rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 999px;
		font-size: 0.7rem;
		color: var(--dt4);
	}

	/* View History button active state */
	.btn-pill.active {
		background: var(--dbg3);
		border-color: var(--dt4);
		color: var(--dt);
	}

	/* ── Loading ─────────────────────────────────────────────────────────────── */
	.dlp-loading {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.dlp-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dbd2);
		border-top-color: var(--dt2);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin { to { transform: rotate(360deg); } }

	/* ── Body ────────────────────────────────────────────────────────────────── */
	.dlp-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	/* ── Form ────────────────────────────────────────────────────────────────── */
	.dlp-form {
		flex: 1;
		min-width: 0;
		overflow-y: auto;
		scrollbar-width: none;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.dlp-body--with-history .dlp-form {
		border-right: 1px solid var(--dbd2);
	}

	.dlp-form::-webkit-scrollbar { display: none; }

	/* Card base */
	.dlp-card {
		padding: 1.1rem 1.25rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 12px;
	}

	.dlp-label {
		display: block;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--dt2);
		margin: 0;
	}

	/* ── Energy card ─────────────────────────────────────────────────────────── */
	.dlp-energy-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.9rem;
	}

	.dlp-energy-readout {
		display: flex;
		align-items: baseline;
		gap: 0.2rem;
	}

	.dlp-energy-num {
		font-size: 1.5rem;
		font-weight: 800;
		font-variant-numeric: tabular-nums;
		line-height: 1;
		color: var(--dt);
	}

	.dlp-energy-slash {
		font-size: 0.8rem;
		color: var(--dt4);
	}

	/* Slider — blue fill only, no dynamic colors */
	.dlp-energy__slider {
		width: 100%;
		height: 5px;
		-webkit-appearance: none;
		appearance: none;
		border-radius: 999px;
		outline: none;
		cursor: pointer;
		background: linear-gradient(
			to right,
			#3b82f6 0%,
			#3b82f6 var(--slider-pct),
			var(--dbd2) var(--slider-pct)
		);
		margin-bottom: 0.1rem;
	}

	.dlp-energy__slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: var(--dbg);
		cursor: pointer;
		border: 2px solid var(--dt2);
		box-shadow: 0 1px 6px rgba(0, 0, 0, 0.2);
	}

	.dlp-energy__slider::-webkit-slider-thumb:hover {
		border-color: var(--dt);
		box-shadow: 0 0 0 5px rgba(59, 130, 246, 0.15), 0 1px 6px rgba(0, 0, 0, 0.2);
	}

	.dlp-energy__slider::-moz-range-thumb {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: var(--dbg);
		cursor: pointer;
		border: 2px solid var(--dt2);
		box-shadow: 0 1px 6px rgba(0, 0, 0, 0.2);
	}

	.dlp-energy__row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		margin-top: 0.55rem;
	}

	.dlp-energy__end-label {
		font-size: 0.68rem;
		color: var(--dt4);
		white-space: nowrap;
		flex-shrink: 0;
	}

	.dlp-energy__dots {
		display: flex;
		align-items: center;
		gap: 4px;
		flex: 1;
		justify-content: center;
	}

	.dlp-energy__dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--dbd2);
		transition: background 0.15s;
		flex-shrink: 0;
	}

	.dlp-energy__dot.active {
		background: #3b82f6;
	}

	/* ── Entry card ──────────────────────────────────────────────────────────── */
	.dlp-entry-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.7rem;
	}

	.dlp-word-count {
		font-size: 0.7rem;
		color: var(--dt4);
		font-variant-numeric: tabular-nums;
	}

	.dlp-textarea {
		width: 100%;
		min-height: 200px;
		resize: none;
		padding: 0.85rem;
		background: var(--dbg);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		color: var(--dt);
		font-family: inherit;
		font-size: 0.88rem;
		line-height: 1.7;
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
		box-sizing: border-box;
		display: block;
	}

	.dlp-textarea::placeholder {
		color: var(--dt4);
		font-style: italic;
	}

	.dlp-textarea:focus {
		border-color: var(--dt3);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08);
	}

	/* ── Save button ─────────────────────────────────────────────────────────── */
	.dlp-save {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.45rem;
		font-size: 0.88rem;
		align-self: flex-end;
	}

	.dlp-save--success {
		background: #16a34a !important;
		color: #fff !important;
		box-shadow: none !important;
		animation: save-pulse 0.65s ease-out;
	}

	.dlp-save--error {
		background: #ef4444 !important;
		color: #fff !important;
		box-shadow: none !important;
	}

	@keyframes save-pulse {
		0%   { box-shadow: 0 0 0 0 rgba(22, 163, 74, 0.5); }
		50%  { box-shadow: 0 0 0 12px rgba(22, 163, 74, 0); }
		100% { box-shadow: 0 0 0 0 rgba(22, 163, 74, 0); }
	}

	.dlp-save__spinner {
		width: 1rem;
		height: 1rem;
		animation: spin 0.7s linear infinite;
	}

	.dlp-save__kbd {
		font-family: inherit;
		font-size: 0.72rem;
		opacity: 0.55;
		padding: 0.1rem 0.35rem;
		border: 1px solid currentColor;
		border-radius: 4px;
		letter-spacing: 0;
	}

	/* ── Analytics Notes ─────────────────────────────────────────────────────── */
	.dlp-analytics-notes {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.dlp-analytics-notes__header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.dlp-analytics-notes__title {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--dt3);
		margin: 0;
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.dlp-analytics-notes__subtitle {
		font-size: 0.75rem;
		color: var(--dt4);
		margin: 0;
		line-height: 1.5;
	}

	.dlp-analytics-notes__count {
		font-size: 0.68rem;
		font-weight: 600;
		padding: 0.1rem 0.45rem;
		background: var(--dbg3);
		border: 1px solid var(--dbd2);
		border-radius: 999px;
		color: var(--dt3);
		font-variant-numeric: tabular-nums;
	}

	.dlp-analytics-notes__list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.dlp-analytics-note-card {
		padding: 0.85rem 1rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.dlp-analytics-note-card__metrics {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.dlp-analytics-note-card__metric {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.dlp-analytics-note-card__metric-label {
		font-size: 0.62rem;
		color: var(--dt4);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.dlp-analytics-note-card__metric-value {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--dt);
		font-variant-numeric: tabular-nums;
	}

	.dlp-analytics-note-card__date {
		font-size: 0.68rem;
		color: var(--dt4);
		margin-left: auto;
	}

	.dlp-analytics-note-card__content {
		font-size: 0.82rem;
		color: var(--dt2);
		margin: 0;
		line-height: 1.55;
	}

	.dlp-analytics-note-card__actions {
		display: flex;
		gap: 0.35rem;
	}

	.dlp-analytics-note-card__action-btn {
		padding: 0.25rem;
		border: none;
		background: none;
		color: var(--dt4);
		cursor: pointer;
		border-radius: 4px;
		display: flex;
		align-items: center;
		transition: color 0.15s, background 0.15s;
	}

	.dlp-analytics-note-card__action-btn:hover {
		color: var(--dt);
		background: var(--dbg3);
	}

	.dlp-analytics-note-card__action-btn--delete:hover {
		color: #ef4444;
	}

	.dlp-analytics-note-card__edit-textarea {
		width: 100%;
		padding: 0.55rem 0.7rem;
		background: var(--dbg);
		border: 1px solid var(--dbd2);
		border-radius: 6px;
		color: var(--dt);
		font-family: inherit;
		font-size: 0.82rem;
		line-height: 1.55;
		outline: none;
		resize: vertical;
		box-sizing: border-box;
	}

	.dlp-analytics-note-card__edit-actions {
		display: flex;
		gap: 0.4rem;
	}

	.dlp-analytics-note-card__edit-save,
	.dlp-analytics-note-card__edit-cancel {
		padding: 0.25rem 0.75rem;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		border: 1px solid var(--dbd2);
	}

	.dlp-analytics-note-card__edit-save {
		background: var(--dt);
		color: var(--dbg);
		border-color: transparent;
	}

	.dlp-analytics-note-card__edit-cancel {
		background: none;
		color: var(--dt3);
	}

	/* ── History panel ───────────────────────────────────────────────────────── */
	.dlp-history-panel {
		width: 280px;
		flex-shrink: 0;
		overflow-y: auto;
		scrollbar-width: none;
		padding: 1.25rem 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		border-left: 1px solid var(--dbd2);
	}

	.dlp-history-panel::-webkit-scrollbar { display: none; }

	.dlp-history-panel__head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--dbd2);
	}

	.dlp-history-panel__title {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--dt3);
		margin: 0;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.dlp-history-panel__count {
		font-size: 0.68rem;
		font-weight: 600;
		padding: 0.1rem 0.45rem;
		background: var(--dbg3);
		border: 1px solid var(--dbd2);
		border-radius: 999px;
		color: var(--dt3);
		font-variant-numeric: tabular-nums;
	}

	.dlp-history-panel__empty {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.65rem;
		color: var(--dt4);
		text-align: center;
		padding: 2rem 0;
	}

	.dlp-history-panel__empty p {
		font-size: 0.78rem;
		line-height: 1.55;
		margin: 0;
	}

	.dlp-history-panel__list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	/* History cards */
	.dlp-history-card {
		width: 100%;
		text-align: left;
		padding: 0.75rem 0.85rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.dlp-history-card:hover {
		background: var(--dbg3);
		border-color: var(--dt4);
	}

	.dlp-history-card__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.dlp-history-card__date {
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--dt2);
	}

	.dlp-history-card__energy {
		font-size: 0.65rem;
		color: var(--dt4);
		font-variant-numeric: tabular-nums;
	}

	.dlp-history-card__preview {
		font-size: 0.75rem;
		color: var(--dt3);
		margin: 0;
		line-height: 1.5;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
