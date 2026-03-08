<script lang="ts">
	import { api, type DailyLog } from '$lib/api';
	import { onMount } from 'svelte';

	let todayEntry = $state('');
	let energyLevel = $state(7);
	let currentLog = $state<DailyLog | null>(null);
	let pastLogs = $state<DailyLog[]>([]);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let showHistory = $state(false);
	let saveMessage = $state('');

	function formatToday() {
		return new Date().toLocaleDateString(undefined, {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, {
			weekday: 'short',
			month: 'short',
			day: 'numeric'
		});
	}

	onMount(async () => {
		await Promise.all([loadTodayLog(), loadPastLogs()]);
		isLoading = false;
	});

	async function loadTodayLog() {
		try {
			const log = await api.getTodayLog();
			if (log) {
				currentLog = log;
				todayEntry = log.content;
				energyLevel = log.energy_level || 7;
			}
		} catch (error) {
			console.error('Error loading today log:', error);
		}
	}

	async function loadPastLogs() {
		try {
			const logs = await api.getDailyLogs(0, 14);
			// Filter out today's log
			const today = new Date().toISOString().split('T')[0];
			pastLogs = logs.filter(log => log.date !== today);
		} catch (error) {
			console.error('Error loading past logs:', error);
		}
	}

	async function handleSave() {
		if (!todayEntry.trim()) return;

		isSaving = true;
		saveMessage = '';

		try {
			const log = await api.saveDailyLog({
				content: todayEntry,
				energy_level: energyLevel
			});
			currentLog = log;
			saveMessage = 'Saved!';
			setTimeout(() => saveMessage = '', 2000);
		} catch (error) {
			console.error('Error saving daily log:', error);
			saveMessage = 'Error saving';
		} finally {
			isSaving = false;
		}
	}

	function loadPastLog(log: DailyLog) {
		// Navigate to that date's log in view mode
		todayEntry = log.content;
		energyLevel = log.energy_level || 7;
		showHistory = false;
	}
</script>

<div class="dlp">
	<!-- Header -->
	<div class="dlp-header">
		<div>
			<h1 class="dlp-title">Daily Log</h1>
			<p class="dlp-subtitle">{formatToday()}</p>
		</div>
		<button
			onclick={() => showHistory = !showHistory}
			class="btn-pill btn-pill-secondary btn-pill-sm"
		>
			<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			{showHistory ? 'Hide History' : 'View History'}
		</button>
	</div>

	{#if isLoading}
		<div class="dlp-loading">
			<div class="dlp-spinner"></div>
		</div>
	{:else}
		<div class="dlp-content">
			{#if showHistory}
				<!-- Past Logs View -->
				<div class="dlp-history">
					<h2 class="dlp-history__heading">Past Entries</h2>
					{#if pastLogs.length === 0}
						<p class="dlp-history__empty">No past entries yet</p>
					{:else}
						{#each pastLogs as log}
							<button
								onclick={() => loadPastLog(log)}
								class="dlp-history__card"
							>
								<div class="dlp-history__card-row">
									<span class="dlp-history__card-date">{formatDate(log.date)}</span>
									{#if log.energy_level}
										<span class="dlp-history__card-energy">Energy: {log.energy_level}/10</span>
									{/if}
								</div>
								<p class="dlp-history__card-preview">{log.content}</p>
							</button>
						{/each}
					{/if}
				</div>
			{:else}
				<!-- Today's Entry View -->
				<div class="dlp-today">
					<!-- Energy Level -->
					<div class="dlp-card">
						<label class="dlp-label">How's your energy today?</label>
						<div class="dlp-energy">
							<input
								type="range"
								min="1"
								max="10"
								bind:value={energyLevel}
								class="dlp-energy__slider"
							/>
							<span class="dlp-energy__value">{energyLevel}</span>
						</div>
						<div class="dlp-energy__labels">
							<span>Low energy</span>
							<span>High energy</span>
						</div>
					</div>

					<!-- Daily Entry -->
					<div class="dlp-card">
						<label for="entry" class="dlp-label">What's on your mind?</label>
						<textarea
							id="entry"
							bind:value={todayEntry}
							class="dlp-textarea"
							placeholder="Write about your day, thoughts, tasks, wins, challenges..."
						></textarea>
					</div>

					<!-- Quick Actions -->
					<div class="dlp-actions">
						<button
							class="btn-pill btn-pill-ghost btn-pill-sm dlp-actions__btn"
							disabled
							title="Voice input coming soon"
						>
							<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
							Voice Input
						</button>
						<button
							class="btn-pill btn-pill-ghost btn-pill-sm dlp-actions__btn"
							disabled
							title="AI action extraction coming soon"
						>
							<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
							</svg>
							Extract Actions
						</button>
					</div>

					<!-- Save Button -->
					<button
						onclick={handleSave}
						disabled={isSaving || !todayEntry.trim()}
						class="btn-pill btn-pill-primary dlp-save"
					>
						{#if isSaving}
							<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Saving...
						{:else if saveMessage}
							{saveMessage}
						{:else}
							Save Entry
						{/if}
					</button>

					{#if currentLog}
						<p class="dlp-saved-at">
							Last saved: {new Date(currentLog.updated_at).toLocaleTimeString()}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	/* ── Daily Log Page — scoped styles ─────────────────────────── */
	.dlp {
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	/* Header */
	.dlp-header {
		padding: 1rem 1.5rem;
		border-bottom: 1px solid var(--dbd2);
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;
	}

	.dlp-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.dlp-subtitle {
		font-size: 0.8rem;
		color: var(--dt3);
		margin: 0.15rem 0 0;
	}

	/* Loading */
	.dlp-loading {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.dlp-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dt);
		border-top-color: transparent;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	/* Content */
	.dlp-content {
		flex: 1;
		overflow-y: auto;
		padding: 1.5rem;
	}

	/* ── History ─────────────────────────────────────────────────── */
	.dlp-history {
		max-width: 40rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.dlp-history__heading {
		font-size: 1.05rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.dlp-history__empty {
		color: var(--dt3);
		text-align: center;
		padding: 2rem 0;
		font-size: 0.88rem;
	}

	.dlp-history__card {
		width: 100%;
		text-align: left;
		padding: 0.85rem 1rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 10px;
		cursor: pointer;
		transition: border-color 0.15s;
	}

	.dlp-history__card:hover {
		border-color: var(--dbd);
	}

	.dlp-history__card-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.35rem;
	}

	.dlp-history__card-date {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--dt);
	}

	.dlp-history__card-energy {
		font-size: 0.8rem;
		color: var(--dt3);
	}

	.dlp-history__card-preview {
		font-size: 0.82rem;
		color: var(--dt2);
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		margin: 0;
		line-height: 1.5;
	}

	/* ── Today's Entry ───────────────────────────────────────────── */
	.dlp-today {
		max-width: 40rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.dlp-card {
		padding: 1rem 1.15rem;
		background: var(--dbg2);
		border: 1px solid var(--dbd2);
		border-radius: 10px;
	}

	.dlp-label {
		display: block;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--dt2);
		margin-bottom: 0.75rem;
	}

	/* Energy slider */
	.dlp-energy {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.dlp-energy__slider {
		flex: 1;
		height: 6px;
		-webkit-appearance: none;
		appearance: none;
		background: var(--dbg3);
		border-radius: 999px;
		outline: none;
		cursor: pointer;
	}

	.dlp-energy__slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--dt);
		cursor: pointer;
		border: 2px solid var(--dbg);
		box-shadow: 0 1px 3px rgba(0,0,0,0.25);
	}

	.dlp-energy__slider::-moz-range-thumb {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--dt);
		cursor: pointer;
		border: 2px solid var(--dbg);
		box-shadow: 0 1px 3px rgba(0,0,0,0.25);
	}

	.dlp-energy__value {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--dt);
		width: 2rem;
		text-align: center;
		font-variant-numeric: tabular-nums;
	}

	.dlp-energy__labels {
		display: flex;
		justify-content: space-between;
		font-size: 0.72rem;
		color: var(--dt4);
		margin-top: 0.5rem;
	}

	/* Textarea */
	.dlp-textarea {
		width: 100%;
		min-height: 200px;
		resize: none;
		padding: 0.75rem;
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 8px;
		color: var(--dt);
		font-family: inherit;
		font-size: 0.88rem;
		line-height: 1.6;
		outline: none;
		transition: border-color 0.15s;
	}

	.dlp-textarea::placeholder {
		color: var(--dt4);
	}

	.dlp-textarea:focus {
		border-color: var(--dt3);
	}

	/* Quick Actions */
	.dlp-actions {
		display: flex;
		gap: 0.75rem;
	}

	.dlp-actions__btn {
		flex: 1;
	}

	.dlp-actions__btn:disabled {
		opacity: 0.4;
	}

	/* Save */
	.dlp-save {
		width: 100%;
	}

	.dlp-saved-at {
		font-size: 0.72rem;
		color: var(--dt4);
		text-align: center;
		margin: 0;
	}
</style>
