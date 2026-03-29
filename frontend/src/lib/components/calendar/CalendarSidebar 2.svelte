<script lang="ts">
	import type { CalendarEvent } from '$lib/api';
	import { isToday, getEventColors } from './calendarUtils';

	interface Props {
		selectedDay: Date;
		events: CalendarEvent[];
		onSelectDay: (date: Date) => void;
		onGoToDayView: (date: Date) => void;
		onOpenCreateModal: (date: Date) => void;
		onOpenEventModal: (event: CalendarEvent) => void;
	}

	let {
		selectedDay = $bindable(),
		events,
		onSelectDay,
		onGoToDayView,
		onOpenCreateModal,
		onOpenEventModal
	}: Props = $props();

	const selectedDayEvents = $derived(
		events
			.filter((event) => {
				const eventDate = new Date(event.start_time);
				return (
					eventDate.getFullYear() === selectedDay.getFullYear() &&
					eventDate.getMonth() === selectedDay.getMonth() &&
					eventDate.getDate() === selectedDay.getDate()
				);
			})
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
	);

	const isSelectedDayToday = $derived(isToday(selectedDay));

	const miniCalDaysInMonth = $derived(
		new Date(selectedDay.getFullYear(), selectedDay.getMonth() + 1, 0).getDate()
	);
	const miniCalFirstDayOffset = $derived(
		new Date(selectedDay.getFullYear(), selectedDay.getMonth(), 1).getDay()
	);

	function hasEventsOnDate(date: Date): boolean {
		return events.some((e) => {
			const ed = new Date(e.start_time);
			return (
				ed.getFullYear() === date.getFullYear() &&
				ed.getMonth() === date.getMonth() &&
				ed.getDate() === date.getDate()
			);
		});
	}

	function prevMonth() {
		const d = new Date(selectedDay);
		d.setMonth(d.getMonth() - 1);
		selectedDay = d;
	}

	function nextMonth() {
		const d = new Date(selectedDay);
		d.setMonth(d.getMonth() + 1);
		selectedDay = d;
	}
</script>

<div class="w-72 border-r border-gray-200 flex flex-col bg-gray-50 flex-shrink-0">
	<!-- Mini Calendar Navigator -->
	<div class="p-4 border-b border-gray-200">
		<div class="flex items-center justify-between mb-3">
			<h3 class="text-sm font-semibold text-gray-900">
				{selectedDay.toLocaleString('default', { month: 'long', year: 'numeric' })}
			</h3>
			<div class="flex items-center gap-1">
				<button onclick={prevMonth} class="p-1 hover:bg-gray-200 rounded" aria-label="Previous month">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<button onclick={nextMonth} class="p-1 hover:bg-gray-200 rounded" aria-label="Next month">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Mini Month Grid -->
		<div class="grid grid-cols-7 gap-1 text-center">
			{#each ['S', 'M', 'T', 'W', 'T', 'F', 'S'] as dayLabel}
				<div class="text-xs font-medium text-gray-500 py-1">{dayLabel}</div>
			{/each}
			{#each Array(miniCalFirstDayOffset) as _, i (i)}
				<div class="w-7 h-7"></div>
			{/each}
			{#each Array(miniCalDaysInMonth) as _, i (i)}
				{@const day = i + 1}
				{@const date = new Date(selectedDay.getFullYear(), selectedDay.getMonth(), day)}
				{@const todayDate = isToday(date)}
				{@const isSelected =
					selectedDay.getDate() === day && selectedDay.getMonth() === date.getMonth()}
				{@const dotVisible = hasEventsOnDate(date) && !todayDate}
				<button
					onclick={() => { onSelectDay(date); onGoToDayView(date); }}
					class="w-7 h-7 text-xs rounded-full flex items-center justify-center transition-colors relative
						{todayDate ? 'bg-gray-900 text-white font-bold' : ''}
						{isSelected && !todayDate ? 'bg-blue-100 text-blue-700 font-medium' : ''}
						{!todayDate && !isSelected ? 'hover:bg-gray-200 text-gray-700' : ''}"
				>
					{day}
					{#if dotVisible}
						<span class="absolute bottom-0.5 w-1 h-1 bg-blue-500 rounded-full"></span>
					{/if}
				</button>
			{/each}
		</div>
	</div>

	<!-- Daily Agenda -->
	<div class="flex-1 overflow-auto p-4">
		<div class="flex items-center justify-between mb-3">
			<h3 class="text-sm font-semibold text-gray-900">
				{isSelectedDayToday
					? "Today's Agenda"
					: selectedDay.toLocaleDateString('en-US', {
							weekday: 'short',
							month: 'short',
							day: 'numeric'
					  })}
			</h3>
			<span class="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
				{selectedDayEvents.length} events
			</span>
		</div>

		{#if selectedDayEvents.length === 0}
			<div class="text-center py-8">
				<svg class="mx-auto w-10 h-10 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				<p class="text-sm text-gray-500 mt-2">No events</p>
				<button
					onclick={() => onOpenCreateModal(selectedDay)}
					class="mt-2 text-xs text-blue-600 hover:text-blue-800 font-medium"
				>
					+ Add event
				</button>
			</div>
		{:else}
			<div class="space-y-2">
				{#each selectedDayEvents as event (event.id)}
					{@const colors = getEventColors(event)}
					<button
						onclick={() => onOpenEventModal(event)}
						class="w-full text-left p-2.5 rounded-lg border transition-all hover:shadow-sm {colors.bg} {colors.border}"
					>
						<p class="text-xs font-medium {colors.text}">
							{#if event.all_day}
								All day
							{:else}
								{new Date(event.start_time).toLocaleTimeString('en-US', {
									hour: 'numeric',
									minute: '2-digit'
								})}
							{/if}
						</p>
						<p class="text-sm font-medium text-gray-900 mt-0.5 line-clamp-2">
							{event.title || 'Untitled'}
						</p>
						{#if event.location}
							<p class="text-xs text-gray-500 mt-1 truncate">{event.location}</p>
						{/if}
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>
