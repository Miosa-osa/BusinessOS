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

<div class="csb">
	<!-- Mini Calendar Navigator -->
	<div class="csb-cal">
		<div class="csb-cal__header">
			<h3 class="csb-cal__month">
				{selectedDay.toLocaleString('default', { month: 'long', year: 'numeric' })}
			</h3>
			<div class="csb-cal__nav">
				<button onclick={prevMonth} class="btn-pill btn-pill-ghost btn-pill-icon csb-cal__nav-btn" aria-label="Previous month">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<button onclick={nextMonth} class="btn-pill btn-pill-ghost btn-pill-icon csb-cal__nav-btn" aria-label="Next month">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Mini Month Grid -->
		<div class="csb-grid">
			{#each ['S', 'M', 'T', 'W', 'T', 'F', 'S'] as dayLabel}
				<div class="csb-grid__label">{dayLabel}</div>
			{/each}
			{#each Array(miniCalFirstDayOffset) as _, i (i)}
				<div class="csb-grid__empty"></div>
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
					class="csb-grid__day"
					class:csb-grid__day--today={todayDate}
					class:csb-grid__day--selected={isSelected && !todayDate}
				>
					{day}
					{#if dotVisible}
						<span class="csb-grid__dot"></span>
					{/if}
				</button>
			{/each}
		</div>
	</div>

	<!-- Daily Agenda -->
	<div class="csb-agenda">
		<div class="csb-agenda__header">
			<h3 class="csb-agenda__title">
				{isSelectedDayToday
					? "Today's Agenda"
					: selectedDay.toLocaleDateString('en-US', {
							weekday: 'short',
							month: 'short',
							day: 'numeric'
					  })}
			</h3>
			<span class="csb-agenda__count">
				{selectedDayEvents.length} events
			</span>
		</div>

		{#if selectedDayEvents.length === 0}
			<div class="csb-agenda__empty">
				<svg class="csb-agenda__empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				<p class="csb-agenda__empty-text">No events</p>
				<button
					onclick={() => onOpenCreateModal(selectedDay)}
					class="btn-pill btn-pill-ghost btn-pill-xs mt-1"
				>
					+ Add event
				</button>
			</div>
		{:else}
			<div class="csb-agenda__list">
				{#each selectedDayEvents as event (event.id)}
					{@const colors = getEventColors(event)}
					<button
						onclick={() => onOpenEventModal(event)}
						class="csb-agenda__event {colors.bg} {colors.border} border"
					>
						<p class="csb-agenda__event-time {colors.text}">
							{#if event.all_day}
								All day
							{:else}
								{new Date(event.start_time).toLocaleTimeString('en-US', {
									hour: 'numeric',
									minute: '2-digit'
								})}
							{/if}
						</p>
						<p class="csb-agenda__event-title">
							{event.title || 'Untitled'}
						</p>
						{#if event.location}
							<p class="csb-agenda__event-loc">{event.location}</p>
						{/if}
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	/* ── Calendar Sidebar ────────────────────────────────────────── */
	.csb {
		width: 15rem;
		border-right: 1px solid var(--dbd2);
		display: flex;
		flex-direction: column;
		background: var(--dbg2);
		flex-shrink: 0;
	}

	/* Mini Calendar */
	.csb-cal {
		padding: 0.65rem;
		border-bottom: 1px solid var(--dbd2);
	}

	.csb-cal__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.4rem;
	}

	.csb-cal__month {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.csb-cal__nav {
		display: flex;
		align-items: center;
		gap: 0.15rem;
	}

	.csb-cal__nav-btn {
		width: 22px !important;
		height: 22px !important;
		min-width: 22px !important;
		padding: 0 !important;
	}

	/* Mini calendar grid */
	.csb-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 1px;
		text-align: center;
	}

	.csb-grid__label {
		font-size: 0.62rem;
		font-weight: 600;
		color: var(--dt4);
		padding: 0.2rem 0;
	}

	.csb-grid__empty {
		width: 100%;
		aspect-ratio: 1;
	}

	.csb-grid__day {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.68rem;
		color: var(--dt2);
		border-radius: 6px;
		cursor: pointer;
		border: none;
		background: transparent;
		padding: 0;
		aspect-ratio: 1;
		transition: background 0.12s;
	}

	.csb-grid__day:hover {
		background: var(--dbg3);
	}

	.csb-grid__day--today {
		background: var(--dt);
		color: var(--dbg);
		font-weight: 700;
	}

	.csb-grid__day--today:hover {
		background: var(--dt);
	}

	.csb-grid__day--selected {
		background: var(--dbg3);
		color: var(--dt);
		font-weight: 600;
	}

	.csb-grid__dot {
		position: absolute;
		bottom: 2px;
		width: 3px;
		height: 3px;
		background: #3b82f6;
		border-radius: 50%;
	}

	/* Daily Agenda */
	.csb-agenda {
		flex: 1;
		overflow: auto;
		padding: 0.65rem;
	}

	.csb-agenda__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}

	.csb-agenda__title {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.csb-agenda__count {
		font-size: 0.68rem;
		color: var(--dt3);
		background: var(--dbg3);
		padding: 0.1rem 0.45rem;
		border-radius: 999px;
	}

	.csb-agenda__empty {
		text-align: center;
		padding: 1.5rem 0;
	}

	.csb-agenda__empty-icon {
		width: 2rem;
		height: 2rem;
		margin: 0 auto;
		color: var(--dt4);
	}

	.csb-agenda__empty-text {
		font-size: 0.8rem;
		color: var(--dt3);
		margin: 0.4rem 0 0;
	}

	.csb-agenda__list {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.csb-agenda__event {
		width: 100%;
		text-align: left;
		padding: 0.45rem 0.6rem;
		border-radius: 7px;
		cursor: pointer;
	}

	.csb-agenda__event-time {
		font-size: 0.68rem;
		font-weight: 600;
		margin: 0;
	}

	.csb-agenda__event-title {
		font-size: 0.78rem;
		font-weight: 500;
		color: var(--dt);
		margin: 0.15rem 0 0;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.csb-agenda__event-loc {
		font-size: 0.68rem;
		color: var(--dt3);
		margin: 0.2rem 0 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
