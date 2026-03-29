<script lang="ts">
	import type { CalendarEvent } from '$lib/api';
	import type { SyncStats } from './calendarUtils';

	interface Props {
		upcomingEvents: CalendarEvent[];
		syncStats: SyncStats | null;
		events: CalendarEvent[];
		onOpenEventModal: (event: CalendarEvent) => void;
		onJumpToFirstEvent: () => void;
		onJumpToLatestEvent: () => void;
	}

	let {
		upcomingEvents,
		syncStats,
		events,
		onOpenEventModal,
		onJumpToFirstEvent,
		onJumpToLatestEvent
	}: Props = $props();
</script>

<!-- Upcoming Events Quick View -->
{#if upcomingEvents.length > 0}
	<div class="mx-6 mt-3 p-4 bg-gradient-to-r from-green-50 to-emerald-50 border border-green-200 rounded-xl">
		<div class="flex items-center justify-between mb-3">
			<div class="flex items-center gap-2">
				<div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
					<svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<span class="text-sm font-semibold text-green-900">Upcoming Events</span>
				<span class="text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded-full">
					{upcomingEvents.length} next
				</span>
			</div>
		</div>
		<div class="flex gap-3 overflow-x-auto pb-1">
			{#each upcomingEvents.slice(0, 5) as event (event.id)}
				<button
					onclick={() => onOpenEventModal(event)}
					class="flex-shrink-0 w-48 p-3 bg-white border border-green-100 rounded-lg hover:shadow-md transition-shadow text-left"
				>
					<p class="text-xs text-green-600 font-medium mb-1">
						{new Date(event.start_time).toLocaleDateString('en-US', {
							weekday: 'short',
							month: 'short',
							day: 'numeric'
						})}
						{#if !event.all_day}
							&bull; {new Date(event.start_time).toLocaleTimeString('en-US', {
								hour: 'numeric',
								minute: '2-digit'
							})}
						{/if}
					</p>
					<p class="text-sm font-medium text-gray-900 truncate">{event.title || 'Untitled'}</p>
					{#if event.location}
						<p class="text-xs text-gray-500 truncate mt-0.5">{event.location}</p>
					{/if}
				</button>
			{/each}
		</div>
	</div>
{/if}

<!-- Sync Stats Banner (no events in view but have synced events) -->
{#if syncStats && syncStats.totalEvents > 0 && events.length === 0 && upcomingEvents.length === 0}
	<div class="mx-6 mt-3 p-4 bg-blue-50 border border-blue-200 rounded-xl flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
				<svg class="w-5 h-5 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
			</div>
			<div>
				<p class="text-sm font-semibold text-blue-900">{syncStats.totalEvents} events synced</p>
				<p class="text-xs text-blue-700">
					{#if syncStats.dateRange?.from && syncStats.dateRange?.to}
						Range: {new Date(syncStats.dateRange.from).toLocaleDateString()} - {new Date(
							syncStats.dateRange.to
						).toLocaleDateString()}
					{/if}
				</p>
			</div>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={onJumpToFirstEvent}
				class="px-3 py-1.5 text-sm font-medium text-blue-700 bg-blue-100 hover:bg-blue-200 rounded-lg transition-colors"
			>
				View Past Events
			</button>
			<button
				onclick={onJumpToLatestEvent}
				class="px-3 py-1.5 text-sm font-medium text-blue-700 bg-blue-100 hover:bg-blue-200 rounded-lg transition-colors"
			>
				View Recent Events
			</button>
		</div>
	</div>
{/if}

<!-- Sync Summary Bar -->
{#if syncStats && syncStats.totalEvents > 0}
	<div class="mx-6 mt-3 flex items-center justify-between text-xs text-gray-500">
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-1.5">
				<svg class="w-4 h-4" viewBox="0 0 24 24">
					<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
					<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
					<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
					<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
				</svg>
				<span class="font-medium">{syncStats.googleEvents} Google</span>
			</div>
			{#if syncStats.localEvents > 0}
				<span>&bull;</span>
				<span>{syncStats.localEvents} local</span>
			{/if}
			{#if events.length > 0}
				<span>&bull;</span>
				<span class="text-gray-700 font-medium">{events.length} in view</span>
			{/if}
		</div>
		{#if syncStats.lastSync}
			<span>Last sync: {new Date(syncStats.lastSync).toLocaleString()}</span>
		{/if}
	</div>
{/if}
