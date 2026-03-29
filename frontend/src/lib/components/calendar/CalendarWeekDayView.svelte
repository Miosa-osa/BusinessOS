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
	<div class="cwdv">
		<!-- Day Headers -->
		<div class="cwdv-week-header">
			<div class="cwdv-gutter"></div>
			{#each weekDates as date (date.toISOString())}
				<div class="cwdv-day-col-header">
					<p class="cwdv-day-label">{weekDays[date.getDay()]}</p>
					<p class="cwdv-day-num" class:cwdv-day-num--today={isToday(date)}>
						{date.getDate()}
					</p>
				</div>
			{/each}
		</div>

		<!-- Time Grid -->
		<div class="cwdv-grid">
			<!-- Current Time Indicator (week) -->
			{#if isCurrentWeek()}
				<div
					class="cwdv-now-line"
					style="top: {currentTimePosition()}px;"
				>
					<div class="cwdv-now-line__inner">
						<div class="cwdv-now-line__gutter"></div>
						<div class="cwdv-now-line__track">
							<div
								class="cwdv-now-dot"
								style="left: calc({todayColumnIndex()} * 14.285%);"
							></div>
							<div
								class="cwdv-now-rule"
								style="left: calc({todayColumnIndex()} * 14.285% + 6px); right: calc((6 - {todayColumnIndex()}) * 14.285%);"
							></div>
						</div>
					</div>
				</div>
			{/if}

			{#each hours as hour (hour)}
				<div class="cwdv-hour-row cwdv-hour-row--week" class:cwdv-hour-row--even={hour % 2 === 0}>
					<div class="cwdv-gutter cwdv-hour-label">
						{formatHour(hour)}
					</div>
					{#each weekDates as date (date.toISOString())}
						<div
							role="button"
							tabindex="0"
							onclick={() => onOpenCreateModalAtHour(date, hour)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onOpenCreateModalAtHour(date, hour); }}
							class="cwdv-cell"
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
	<div class="cwdv cwdv--day">
		<!-- Day Header -->
		<div class="cwdv-day-header">
			<div class="cwdv-day-header__inner">
				<p class="cwdv-day-header__weekday">
					{currentDate.toLocaleDateString('en-US', { weekday: 'long' })}
				</p>
				<div class="cwdv-day-header__badge" class:cwdv-day-header__badge--today={isToday(currentDate)}>
					{currentDate.getDate()}
				</div>
			</div>
		</div>

		<!-- All Day Events -->
		{#if events.filter((e) => e.all_day).length > 0}
			<div class="cwdv-allday">
				<p class="cwdv-allday__label">All Day</p>
				<div class="cwdv-allday__list">
					{#each events.filter((e) => e.all_day) as event (event.id)}
						{@const colors = getEventColors(event)}
						<button
							onclick={() => onOpenEventModal(event)}
							class="cwdv-allday__event {colors.bg} {colors.border} {colors.text} border"
						>
							{event.title || 'Untitled'}
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Time Grid -->
		<div class="cwdv-grid">
			<!-- Current Time Indicator (day) -->
			{#if isToday(currentDate)}
				<div
					class="cwdv-now-line"
					style="top: {currentTimePosition()}px;"
				>
					<div class="cwdv-now-line__inner">
						<div class="cwdv-now-line__gutter cwdv-now-line__gutter--day"></div>
						<div class="cwdv-now-line__track">
							<div class="cwdv-now-dot cwdv-now-dot--day"></div>
							<div class="cwdv-now-rule cwdv-now-rule--day"></div>
						</div>
					</div>
				</div>
			{/if}

			{#each hours as hour (hour)}
				{@const hourEvents = getEventsForHour(events, currentDate, hour).filter((e) => !e.all_day)}
				<div class="cwdv-hour-row cwdv-hour-row--day" class:cwdv-hour-row--even={hour % 2 === 0}>
					<div class="cwdv-gutter cwdv-hour-label cwdv-gutter--day">
						{formatHour(hour)}
					</div>
					<div
						role="button"
						tabindex="0"
						onclick={() => onOpenCreateModalAtHour(currentDate, hour)}
						onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onOpenCreateModalAtHour(currentDate, hour); }}
						class="cwdv-cell cwdv-cell--day"
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
								class="cwdv-day-event {colors.bg} {colors.border} {colors.text} border"
								style="top: {topOffset}px; height: {Math.max(20, durationMinutes)}px;"
								aria-label="View event: {event.title || 'Untitled'}"
							>
								<p class="font-medium truncate">{event.title || 'Untitled'}</p>
								{#if durationMinutes >= 40}
									<p class="cwdv-day-event__time">
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

<style>
	/* ── Calendar Week/Day View ─────────────────────────────────── */
	.cwdv {
		min-width: 800px;
	}

	.cwdv--day {
		min-width: 400px;
		height: 100%;
	}

	/* ── Week header row ──────────────────────────────────────── */
	.cwdv-week-header {
		display: grid;
		grid-template-columns: 3.5rem repeat(7, 1fr);
		border-bottom: 1px solid var(--dbd);
		position: sticky;
		top: 0;
		background: var(--dbg);
		z-index: 10;
		padding: 0.5rem 0;
	}

	.cwdv-day-col-header {
		padding: 0.25rem 0.5rem 0.35rem;
		text-align: center;
		border-left: 1px solid var(--dbd2);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
	}

	.cwdv-day-label {
		font-size: 0.72rem;
		font-weight: 500;
		letter-spacing: 0.02em;
		color: var(--dt3);
		margin: 0;
		text-transform: uppercase;
	}

	.cwdv-day-num {
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
		line-height: 1;
		width: 1.85rem;
		height: 1.85rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cwdv-day-num--today {
		background: var(--dt);
		color: var(--dbg);
		width: 1.85rem;
		height: 1.85rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.88rem;
		font-weight: 700;
		box-shadow:
			0 0 0 3px color-mix(in srgb, var(--dt) 15%, transparent),
			0 2px 8px color-mix(in srgb, var(--dt) 25%, transparent);
	}

	/* ── Time grid ────────────────────────────────────────────── */
	.cwdv-grid {
		position: relative;
	}

	.cwdv-gutter {
		width: 3.5rem;
		flex-shrink: 0;
	}

	.cwdv-gutter--day {
		width: 3.5rem;
	}

	.cwdv-hour-label {
		padding: 0.3rem 0.6rem 0 0;
		font-size: 0.7rem;
		font-weight: 400;
		color: var(--dt4);
		text-align: right;
		line-height: 1;
	}

	.cwdv-hour-row {
		height: 60px;
		border-bottom: 1px solid var(--dbd2);
	}

	.cwdv-hour-row--even {
		background: color-mix(in srgb, var(--dt) 1.5%, transparent);
	}

	.cwdv-hour-row--week {
		display: grid;
		grid-template-columns: 3.5rem repeat(7, 1fr);
	}

	.cwdv-hour-row--day {
		display: flex;
	}

	.cwdv-cell {
		border-left: 1px solid var(--dbd2);
		position: relative;
		padding: 2px;
		width: 100%;
		height: 100%;
		cursor: pointer;
		transition: background 0.12s ease;
	}

	.cwdv-cell:hover {
		background: color-mix(in srgb, var(--dt) 5%, transparent);
	}

	.cwdv-cell--day {
		flex: 1;
	}

	/* ── Now indicator line ───────────────────────────────────── */
	.cwdv-now-line {
		position: absolute;
		left: 0;
		right: 0;
		z-index: 20;
		pointer-events: none;
	}

	.cwdv-now-line__inner {
		display: flex;
		align-items: center;
	}

	.cwdv-now-line__gutter {
		width: 3.5rem;
		flex-shrink: 0;
	}

	.cwdv-now-line__gutter--day {
		width: 3.5rem;
	}

	.cwdv-now-line__track {
		flex: 1;
		position: relative;
	}

	.cwdv-now-dot {
		position: absolute;
		width: 12px;
		height: 12px;
		background: color-mix(in srgb, var(--bos-status-error) 95%, #ff0000 5%);
		border-radius: 50%;
		transform: translateY(-50%);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--bos-status-error) 25%, transparent);
	}

	.cwdv-now-dot--day {
		left: -6px;
	}

	.cwdv-now-rule {
		position: absolute;
		height: 2.5px;
		background: color-mix(in srgb, var(--bos-status-error) 95%, #ff0000 5%);
		transform: translateY(-50%);
		top: 0;
	}

	.cwdv-now-rule--day {
		left: 0;
		right: 0;
	}

	/* ── Day view header ──────────────────────────────────────── */
	.cwdv-day-header {
		border-bottom: 1px solid var(--dbd);
		position: sticky;
		top: 0;
		background: var(--dbg);
		z-index: 10;
		padding: 0.75rem 1rem;
	}

	.cwdv-day-header__inner {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.65rem;
	}

	.cwdv-day-header__weekday {
		font-size: 1rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.cwdv-day-header__badge {
		width: 2.25rem;
		height: 2.25rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1rem;
		font-weight: 700;
		background: var(--dbg3);
		color: var(--dt);
	}

	.cwdv-day-header__badge--today {
		background: var(--dt);
		color: var(--dbg);
		box-shadow:
			0 0 0 3px color-mix(in srgb, var(--dt) 15%, transparent),
			0 2px 10px color-mix(in srgb, var(--dt) 20%, transparent);
	}

	/* ── All day strip ────────────────────────────────────────── */
	.cwdv-allday {
		border-bottom: 1px solid var(--dbd);
		padding: 0.6rem 1rem;
		background: var(--dbg2);
	}

	.cwdv-allday__label {
		font-size: 0.72rem;
		font-weight: 500;
		letter-spacing: 0.02em;
		color: var(--dt3);
		margin: 0 0 0.4rem;
		text-transform: uppercase;
	}

	.cwdv-allday__list {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.cwdv-allday__event {
		width: 100%;
		text-align: left;
		padding: 0.3rem 0.6rem;
		font-size: 0.78rem;
		border-radius: 5px;
		cursor: pointer;
		transition: opacity 0.1s ease;
	}

	.cwdv-allday__event:hover {
		opacity: 0.85;
	}

	/* ── Day view event cards ─────────────────────────────────── */
	.cwdv-day-event {
		position: absolute;
		left: 4px;
		right: 4px;
		border-radius: 6px;
		padding: 0.35rem 0.5rem;
		font-size: 0.72rem;
		overflow: hidden;
		cursor: pointer;
		transition: opacity 0.1s ease, filter 0.1s ease;
	}

	.cwdv-day-event:hover {
		filter: brightness(0.95);
	}

	.cwdv-day-event__time {
		font-size: 0.65rem;
		opacity: 0.75;
		margin: 0.15rem 0 0;
	}
</style>
