<script lang="ts">
	import type { CalendarEvent } from '$lib/api';
	import CalendarEventCard from './CalendarEventCard.svelte';
	import {
		weekDays,
		hours,
		isToday,
		formatHour,
		getEventColors,
		getEventsForHour
	} from './calendarUtils';

	interface Props {
		/** 'week' renders 7 columns; 'day' renders 1 column for currentDate */
		mode: 'week' | 'day';
		currentDate: Date;
		weekDates: Date[];
		events: CalendarEvent[];
		currentTime: Date;
		onOpenEventModal: (event: CalendarEvent) => void;
		onOpenCreateModalAtHour: (date: Date, hour: number) => void;
	}

	let {
		mode,
		currentDate,
		weekDates,
		events,
		currentTime,
		onOpenEventModal,
		onOpenCreateModalAtHour
	}: Props = $props();

	const currentTimePosition = $derived(() => {
		const h = currentTime.getHours();
		const m = currentTime.getMinutes();
		return h * 60 + m;
	});

	// Week-view: is today in the visible week?
	const isCurrentWeek = $derived(() => {
		if (mode !== 'week' || weekDates.length === 0) return false;
		const today = new Date();
		return today >= weekDates[0] && today <= weekDates[6];
	});

	const todayColumnIndex = $derived(() => new Date().getDay());
</script>

{#if mode === 'week'}
	<div class="min-w-[800px]">
		<!-- Day Headers -->
		<div class="grid grid-cols-8 border-b border-gray-200 sticky top-0 bg-white z-10">
			<div class="p-2 text-xs text-gray-500"></div>
			{#each weekDates as date (date.toISOString())}
				<div class="p-2 text-center border-l border-gray-200">
					<p class="text-xs text-gray-500">{weekDays[date.getDay()]}</p>
					<p
						class="text-lg font-semibold {isToday(date)
							? 'bg-gray-900 text-white w-8 h-8 rounded-full mx-auto flex items-center justify-center'
							: 'text-gray-900'}"
					>
						{date.getDate()}
					</p>
				</div>
			{/each}
		</div>

		<!-- Time Grid -->
		<div class="relative">
			<!-- Current Time Indicator (week) -->
			{#if isCurrentWeek()}
				<div
					class="absolute left-0 right-0 z-20 pointer-events-none"
					style="top: {currentTimePosition()}px;"
				>
					<div class="flex items-center">
						<div class="w-[calc(12.5%)]"></div>
						<div class="flex-1 relative">
							<div
								class="absolute w-3 h-3 bg-red-500 rounded-full -translate-y-1/2"
								style="left: calc({todayColumnIndex()} * 14.285%);"
							></div>
							<div
								class="absolute h-0.5 bg-red-500"
								style="left: calc({todayColumnIndex()} * 14.285% + 6px); right: calc((6 - {todayColumnIndex()}) * 14.285%);"
							></div>
						</div>
					</div>
				</div>
			{/if}

			{#each hours as hour (hour)}
				<div class="grid grid-cols-8 border-b border-gray-100" style="height: 60px;">
					<div class="p-2 text-xs text-gray-400 text-right pr-3">
						{formatHour(hour)}
					</div>
					{#each weekDates as date (date.toISOString())}
						<!-- Using div+role instead of button to allow nested interactive event cards -->
						<div
							role="button"
							tabindex="0"
							onclick={() => onOpenCreateModalAtHour(date, hour)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onOpenCreateModalAtHour(date, hour); }}
							class="border-l border-gray-100 relative p-0.5 w-full h-full hover:bg-gray-50 transition-colors cursor-pointer"
							aria-label="Add event at {date.toLocaleDateString()} {formatHour(hour)}"
						>
							{#each getEventsForHour(events, date, hour) as event (event.id)}
								<CalendarEventCard
									{event}
									compact
									onClick={() => onOpenEventModal(event)}
								/>
							{/each}
						</div>
					{/each}
				</div>
			{/each}
		</div>
	</div>
{:else}
	<!-- Day View -->
	<div class="min-w-[400px] h-full">
		<!-- Day Header -->
		<div class="border-b border-gray-200 sticky top-0 bg-white z-10 p-4">
			<div class="flex items-center justify-center">
				<p class="text-lg font-semibold {isToday(currentDate) ? 'text-gray-900' : 'text-gray-700'}">
					{currentDate.toLocaleDateString('en-US', { weekday: 'long' })}
				</p>
				<div
					class="ml-3 {isToday(currentDate)
						? 'bg-gray-900 text-white'
						: 'bg-gray-100 text-gray-900'} w-10 h-10 rounded-full flex items-center justify-center text-lg font-bold"
				>
					{currentDate.getDate()}
				</div>
			</div>
		</div>

		<!-- All Day Events -->
		{#if events.filter((e) => e.all_day).length > 0}
			<div class="border-b border-gray-200 p-3 bg-gray-50">
				<p class="text-xs text-gray-500 font-medium mb-2">All Day</p>
				<div class="space-y-1">
					{#each events.filter((e) => e.all_day) as event (event.id)}
						{@const colors = getEventColors(event)}
						<button
							onclick={() => onOpenEventModal(event)}
							class="w-full text-left px-2 py-1.5 text-sm rounded {colors.bg} {colors.border} {colors.text} border"
						>
							{event.title || 'Untitled'}
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Time Grid -->
		<div class="relative">
			<!-- Current Time Indicator (day) -->
			{#if isToday(currentDate)}
				<div
					class="absolute left-0 right-0 z-20 pointer-events-none"
					style="top: {currentTimePosition()}px;"
				>
					<div class="flex items-center">
						<div class="w-16"></div>
						<div class="flex-1 relative">
							<div class="absolute -left-1.5 w-3 h-3 bg-red-500 rounded-full -translate-y-1/2"></div>
							<div class="h-0.5 bg-red-500"></div>
						</div>
					</div>
				</div>
			{/if}

			{#each hours as hour (hour)}
				{@const hourEvents = getEventsForHour(events, currentDate, hour).filter((e) => !e.all_day)}
				<div class="flex border-b border-gray-100" style="height: 60px;">
					<div class="w-16 flex-shrink-0 p-2 text-xs text-gray-400 text-right pr-3">
						{formatHour(hour)}
					</div>
					<!-- Using div+role to allow absolute-positioned event buttons inside -->
					<div
						role="button"
						tabindex="0"
						onclick={() => onOpenCreateModalAtHour(currentDate, hour)}
						onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onOpenCreateModalAtHour(currentDate, hour); }}
						class="flex-1 relative border-l border-gray-100 p-0.5 hover:bg-gray-50 transition-colors cursor-pointer"
						aria-label="Add event at {formatHour(hour)}"
					>
						{#each hourEvents as event (event.id)}
							{@const colors = getEventColors(event)}
							{@const startTime = new Date(event.start_time)}
							{@const endTime = new Date(event.end_time)}
							{@const durationMinutes = Math.min(
								180,
								(endTime.getTime() - startTime.getTime()) / 60000
							)}
							{@const topOffset = startTime.getMinutes()}
							<button
								onclick={(e) => { e.stopPropagation(); onOpenEventModal(event); }}
								class="absolute left-1 right-1 rounded px-2 py-1 text-xs overflow-hidden border {colors.bg} {colors.border} {colors.text}"
								style="top: {topOffset}px; height: {Math.max(20, durationMinutes)}px;"
								aria-label="View event: {event.title || 'Untitled'}"
							>
								<p class="font-medium truncate">{event.title || 'Untitled'}</p>
								{#if durationMinutes >= 40}
									<p class="text-xs opacity-75">
										{startTime.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })} - {endTime.toLocaleTimeString(
											'en-US',
											{ hour: 'numeric', minute: '2-digit' }
										)}
									</p>
								{/if}
							</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	</div>
{/if}
