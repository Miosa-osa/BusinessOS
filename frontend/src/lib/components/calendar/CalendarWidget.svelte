<script lang="ts">
	import type { CalendarEvent } from '$lib/api/calendar';
	import type { GoogleConnectionStatus } from '$lib/api/integrations';
	import { getTodayEvents } from '$lib/api/calendar';
	import { getGoogleConnectionStatus } from '$lib/api/integrations';
	import { onMount } from 'svelte';

	interface Props {
		maxEvents?: number;
		showHeader?: boolean;
		onViewAll?: () => void;
	}

	let { maxEvents = 5, showHeader = true, onViewAll }: Props = $props();

	let events = $state<CalendarEvent[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let isConnected = $state(false);

	onMount(async () => {
		await checkConnectionAndLoadEvents();
	});

	async function checkConnectionAndLoadEvents() {
		try {
			const status = await getGoogleConnectionStatus();
			isConnected = status.connected;

			if (status.connected) {
				const todayEvents = await getTodayEvents();
				events = todayEvents.slice(0, maxEvents);
			}
		} catch (err) {
			console.error('Error loading calendar events:', err);
			error = 'Failed to load events';
		} finally {
			isLoading = false;
		}
	}

	function formatTime(dateString: string): string {
		return new Date(dateString).toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	function formatTimeRange(start: string, end: string, allDay: boolean): string {
		if (allDay) return 'All day';
		return `${formatTime(start)} - ${formatTime(end)}`;
	}

	function getMeetingTypeColor(type: string): string {
		const colors: Record<string, string> = {
			team: 'bg-blue-100 text-blue-700',
			sales: 'bg-green-100 text-green-700',
			onboarding: 'bg-purple-100 text-purple-700',
			kickoff: 'bg-orange-100 text-orange-700',
			implementation: 'bg-cyan-100 text-cyan-700',
			standup: 'bg-indigo-100 text-indigo-700',
			retrospective: 'bg-pink-100 text-pink-700',
			planning: 'bg-yellow-100 text-yellow-700',
			review: 'bg-teal-100 text-teal-700',
			one_on_one: 'bg-rose-100 text-rose-700',
			client: 'bg-emerald-100 text-emerald-700',
			internal: 'bg-slate-100 text-slate-700',
			external: 'bg-amber-100 text-amber-700',
			other: 'bg-gray-100 text-gray-700'
		};
		return colors[type] || colors.other;
	}

	function isEventNow(event: CalendarEvent): boolean {
		const now = new Date();
		const start = new Date(event.start_time);
		const end = new Date(event.end_time);
		return now >= start && now <= end;
	}
</script>

<div class="card">
	{#if showHeader}
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-medium text-gray-900">Today's Schedule</h2>
			{#if onViewAll}
				<button
					onclick={onViewAll}
					class="text-sm text-gray-500 hover:text-gray-700"
				>
					View all
				</button>
			{/if}
		</div>
	{/if}

	{#if isLoading}
		<div class="flex items-center justify-center py-8">
			<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else if !isConnected}
		<div class="text-center py-6">
			<svg class="w-12 h-12 mx-auto text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
			</svg>
			<p class="text-sm text-gray-500 mb-3">Connect your Google Calendar to see your schedule</p>
			<a href="/settings?tab=integrations" class="btn-pill-sm btn-pill-primary">
				Connect Calendar
			</a>
		</div>
	{:else if error}
		<div class="text-center py-6 text-red-500">
			<p class="text-sm">{error}</p>
		</div>
	{:else if events.length === 0}
		<div class="text-center py-6">
			<svg class="w-12 h-12 mx-auto text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
			</svg>
			<p class="text-sm text-gray-500">No events scheduled for today</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each events as event (event.id)}
				<div
					class="flex items-start gap-3 p-3 rounded-lg transition-colors {isEventNow(event) ? 'bg-blue-50 border border-blue-200' : 'bg-gray-50 hover:bg-gray-100'}"
				>
					<div class="flex-shrink-0 w-1 h-full min-h-[40px] rounded-full {isEventNow(event) ? 'bg-blue-500' : 'bg-gray-300'}"></div>
					<div class="flex-1 min-w-0">
						<div class="flex items-start justify-between gap-2">
							<p class="font-medium text-gray-900 truncate">
								{event.title || 'Untitled Event'}
							</p>
							{#if isEventNow(event)}
								<span class="flex-shrink-0 px-1.5 py-0.5 text-xs font-medium bg-blue-500 text-white rounded">
									Now
								</span>
							{/if}
						</div>
						<p class="text-sm text-gray-500">
							{formatTimeRange(event.start_time, event.end_time, event.all_day)}
						</p>
						{#if event.location}
							<p class="text-xs text-gray-400 mt-1 truncate">
								{event.location}
							</p>
						{/if}
						{#if event.meeting_type && event.meeting_type !== 'other'}
							<span class="inline-block mt-1.5 px-2 py-0.5 text-xs font-medium rounded-full {getMeetingTypeColor(event.meeting_type)}">
								{event.meeting_type.replace('_', ' ')}
							</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
