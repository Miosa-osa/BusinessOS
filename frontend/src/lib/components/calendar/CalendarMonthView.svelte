<script lang="ts">
	import type { CalendarEvent } from '$lib/api';
	import CalendarEventCard from './CalendarEventCard.svelte';
	import { weekDays, isToday, getEventsForDate, buildMonthData } from './calendarUtils';

	interface Props {
		currentDate: Date;
		selectedDay: Date;
		events: CalendarEvent[];
		dateRange: { start: Date; end: Date };
		onSelectDay: (date: Date) => void;
		onGoToDayView: (date: Date) => void;
		onOpenCreateModal: (date: Date) => void;
		onOpenEventModal: (event: CalendarEvent) => void;
	}

	let {
		currentDate,
		selectedDay,
		events,
		dateRange,
		onSelectDay,
		onGoToDayView,
		onOpenCreateModal,
		onOpenEventModal
	}: Props = $props();

	const monthData = $derived(buildMonthData(currentDate, dateRange));

	function isCurrentMonth(date: Date): boolean {
		return date.getMonth() === currentDate.getMonth();
	}

	function isSelected(date: Date): boolean {
		return (
			selectedDay.getDate() === date.getDate() &&
			selectedDay.getMonth() === date.getMonth() &&
			selectedDay.getFullYear() === date.getFullYear()
		);
	}
</script>

<div class="p-4">
	<!-- Day Headers -->
	<div class="grid grid-cols-7 mb-2">
		{#each weekDays as day (day)}
			<div class="text-center text-sm font-medium py-2" style="color: var(--dt3)">{day}</div>
		{/each}
	</div>

	<!-- Month Grid -->
	<div class="grid grid-cols-7 gap-1">
		{#each monthData.flat() as date (date.toISOString())}
			{@const todayDate = isToday(date)}
			{@const selectedDate = isSelected(date)}
			{@const dateEvents = getEventsForDate(events, date)}
			<div
				role="button"
				tabindex="0"
				onclick={() => onSelectDay(date)}
				ondblclick={() => onGoToDayView(date)}
				onkeydown={(e) => { if (e.key === 'Enter') onSelectDay(date); }}
				class="group min-h-[100px] p-2 border rounded-lg text-left hover:border-gray-400 transition-colors cursor-pointer {todayDate ? '' : 'border-gray-200'}"
				style="{isCurrentMonth(date) ? 'background: var(--dbg)' : 'background: var(--dbg2)'}; {selectedDate ? 'box-shadow: 0 0 0 2px var(--bos-nav-active); background: var(--bos-nav-active-bg)' : todayDate ? 'box-shadow: 0 0 0 2px var(--dt)' : ''}"
			>
				<div class="flex items-center justify-between">
					<p class="text-sm font-medium" style="{isCurrentMonth(date) ? 'color: var(--dt)' : 'color: var(--dt4)'}">{date.getDate()}</p>
					<button
						onclick={(e) => { e.stopPropagation(); onOpenCreateModal(date); }}
						class="btn-pill btn-pill-ghost btn-pill-icon w-5 h-5 opacity-0 group-hover:opacity-100"
						title="Add event"
						aria-label="Add event on {date.toLocaleDateString()}"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</button>
				</div>

				<div class="mt-1 space-y-1">
					{#each dateEvents.slice(0, 3) as event (event.id)}
						<button
							onclick={(e) => { e.stopPropagation(); onOpenEventModal(event); }}
							class="w-full"
						>
							<CalendarEventCard {event} compact onClick={() => onOpenEventModal(event)} />
						</button>
					{/each}
					{#if dateEvents.length > 3}
						<button
							onclick={(e) => { e.stopPropagation(); onGoToDayView(date); }}
							class="btn-pill btn-pill-ghost btn-pill-xs pl-2"
						>
							+{dateEvents.length - 3} more
						</button>
					{/if}
				</div>
			</div>
		{/each}
	</div>
</div>
