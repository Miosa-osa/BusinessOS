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
		await loadTodayLog();
		await loadPastLogs();
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

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-gray-900">Daily Log</h1>
			<p class="text-sm text-gray-500 mt-0.5">{formatToday()}</p>
		</div>
		<button
			onclick={() => showHistory = !showHistory}
			class="btn btn-secondary text-sm"
		>
			<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			{showHistory ? 'Hide History' : 'View History'}
		</button>
	</div>

	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Content -->
		<div class="flex-1 overflow-y-auto p-6">
			{#if showHistory}
				<!-- Past Logs View -->
				<div class="max-w-2xl mx-auto space-y-4">
					<h2 class="text-lg font-medium text-gray-900">Past Entries</h2>
					{#if pastLogs.length === 0}
						<p class="text-gray-500 text-center py-8">No past entries yet</p>
					{:else}
						{#each pastLogs as log}
							<button
								onclick={() => loadPastLog(log)}
								class="w-full text-left card hover:bg-gray-50 transition-colors"
							>
								<div class="flex items-center justify-between mb-2">
									<span class="text-sm font-medium text-gray-900">{formatDate(log.date)}</span>
									{#if log.energy_level}
										<span class="text-sm text-gray-500">Energy: {log.energy_level}/10</span>
									{/if}
								</div>
								<p class="text-sm text-gray-600 line-clamp-2">{log.content}</p>
							</button>
						{/each}
					{/if}
				</div>
			{:else}
				<!-- Today's Entry View -->
				<div class="max-w-2xl mx-auto space-y-6">
					<!-- Energy Level -->
					<div class="card">
						<label class="block text-sm font-medium text-gray-700 mb-3">How's your energy today?</label>
						<div class="flex items-center gap-4">
							<input
								type="range"
								min="1"
								max="10"
								bind:value={energyLevel}
								class="flex-1 h-2 bg-gray-200 rounded-full appearance-none cursor-pointer accent-gray-900"
							/>
							<span class="text-2xl font-medium text-gray-900 w-8 text-center">{energyLevel}</span>
						</div>
						<div class="flex justify-between text-xs text-gray-400 mt-2">
							<span>Low energy</span>
							<span>High energy</span>
						</div>
					</div>

					<!-- Daily Entry -->
					<div class="card">
						<label for="entry" class="block text-sm font-medium text-gray-700 mb-3">What's on your mind?</label>
						<textarea
							id="entry"
							bind:value={todayEntry}
							class="input input-square resize-none min-h-[200px]"
							placeholder="Write about your day, thoughts, tasks, wins, challenges..."
						></textarea>
					</div>

					<!-- Quick Actions -->
					<div class="flex gap-3">
						<button
							class="btn btn-secondary flex-1 opacity-50 cursor-not-allowed"
							disabled
							title="Voice input coming soon"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
							Voice Input
						</button>
						<button
							class="btn btn-secondary flex-1 opacity-50 cursor-not-allowed"
							disabled
							title="AI action extraction coming soon"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
							</svg>
							Extract Actions
						</button>
					</div>

					<!-- Save Button -->
					<button
						onclick={handleSave}
						disabled={isSaving || !todayEntry.trim()}
						class="btn btn-primary w-full py-3 disabled:opacity-50 disabled:cursor-not-allowed"
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
						<p class="text-xs text-gray-400 text-center">
							Last saved: {new Date(currentLog.updated_at).toLocaleTimeString()}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>
