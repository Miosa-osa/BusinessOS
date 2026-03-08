<script lang="ts">
	import {
		api,
		apiClient,
		type CalendarEvent,
		type GoogleConnectionStatus,
		type MeetingType
	} from '$lib/api';
	import { getCalendarConnectionStatus, getCalendarAuthUrl } from '$lib/api/calendar';
	import { onMount, tick } from 'svelte';

	import CalendarSidebar from '$lib/components/calendar/CalendarSidebar.svelte';
	import CalendarWeekDayView from '$lib/components/calendar/CalendarWeekDayView.svelte';
	import CalendarMonthView from '$lib/components/calendar/CalendarMonthView.svelte';
	import CalendarAgendaView from '$lib/components/calendar/CalendarAgendaView.svelte';
	import CalendarEventModal from '$lib/components/calendar/CalendarEventModal.svelte';
	import CalendarEventForm from '$lib/components/calendar/CalendarEventForm.svelte';
	import CalendarStatusBanner from '$lib/components/calendar/CalendarStatusBanner.svelte';
	import CalendarToolbar from '$lib/components/calendar/CalendarToolbar.svelte';

	import {
		type ViewMode,
		type SyncStats,
		type EventFormData,
		buildDateRange,
		isToday
	} from '$lib/components/calendar/calendarUtils';

	// ── View State ────────────────────────────────────────────────────────────
	let viewMode = $state<ViewMode>('week');
	let currentDate = $state(new Date());
	let selectedDay = $state(new Date());
	let showSidebar = $state(true);

	// ── Data ─────────────────────────────────────────────────────────────────
	let events = $state<CalendarEvent[]>([]);
	let upcomingEvents = $state<CalendarEvent[]>([]);
	let syncStats = $state<SyncStats | null>(null);
	let isLoading = $state(true);
	let isSyncing = $state(false);
	let connectionStatus = $state<GoogleConnectionStatus | null>(null);
	let selectedMeetingType = $state<MeetingType | ''>('');

	// ── Event Modal ───────────────────────────────────────────────────────────
	let selectedEvent = $state<CalendarEvent | null>(null);
	let showEventModal = $state(false);

	// ── Create/Edit Modal ─────────────────────────────────────────────────────
	let showCreateModal = $state(false);
	let editingEvent = $state<CalendarEvent | null>(null);
	let isSaving = $state(false);
	let formError = $state('');
	let formData = $state<EventFormData>({
		title: '',
		description: '',
		start_date: '',
		start_time: '09:00',
		end_date: '',
		end_time: '10:00',
		all_day: false,
		location: '',
		meeting_type: '',
		meeting_link: ''
	});

	// ── Time Indicator ────────────────────────────────────────────────────────
	let currentTime = $state(new Date());
	let timeGridRef: HTMLDivElement | null = $state(null);

	$effect(() => {
		const interval = setInterval(() => {
			currentTime = new Date();
		}, 60000);
		return () => clearInterval(interval);
	});

	// Scroll to current time on view change
	$effect(() => {
		if (timeGridRef && (viewMode === 'week' || viewMode === 'day') && !isLoading) {
			const scrollTop = Math.max(0, (new Date().getHours() - 2) * 60);
			setTimeout(() => {
				timeGridRef?.scrollTo({ top: scrollTop, behavior: 'smooth' });
			}, 100);
		}
	});

	// ── Derived ───────────────────────────────────────────────────────────────
	const dateRange = $derived(buildDateRange(viewMode, currentDate));

	const weekDates = $derived((): Date[] => {
		if (viewMode !== 'week') return [];
		const dates: Date[] = [];
		const current = new Date(dateRange.start);
		for (let i = 0; i < 7; i++) {
			dates.push(new Date(current));
			current.setDate(current.getDate() + 1);
		}
		return dates;
	});

	const headerText = $derived((): string => {
		if (viewMode === 'day') {
			return currentDate.toLocaleString('default', {
				weekday: 'long',
				month: 'long',
				day: 'numeric',
				year: 'numeric'
			});
		}
		if (viewMode === 'week') {
			const start = dateRange.start;
			const end = dateRange.end;
			const startMonth = start.toLocaleString('default', { month: 'short' });
			const endMonth = end.toLocaleString('default', { month: 'short' });
			const year = start.getFullYear();
			if (startMonth === endMonth) {
				return `${startMonth} ${start.getDate()} - ${end.getDate()}, ${year}`;
			}
			return `${startMonth} ${start.getDate()} - ${endMonth} ${end.getDate()}, ${year}`;
		}
		if (viewMode === 'agenda') return 'Upcoming Events';
		return currentDate.toLocaleString('default', { month: 'long', year: 'numeric' });
	});

	// ── Bootstrap ─────────────────────────────────────────────────────────────
	onMount(async () => {
		currentDate = new Date();
		await Promise.all([
			loadConnectionStatus(),
			loadEvents(),
			loadSyncStats(),
			loadUpcomingEvents()
		]);
		isLoading = false;
	});

	// Reload events when filter changes
	$effect(() => {
		selectedMeetingType; // track dependency
		loadEvents();
	});

	// ── Data Loaders ──────────────────────────────────────────────────────────
	async function loadUpcomingEvents() {
		try {
			const res = await apiClient.get('/calendar/upcoming');
			const data = res.ok ? await res.json() : [];
			upcomingEvents = Array.isArray(data) ? data : [];
		} catch {
			upcomingEvents = [];
		}
	}

	async function loadSyncStats() {
		try {
			const res = await apiClient.get('/calendar/stats');
			if (res.ok) {
				const data = await res.json();
				syncStats = {
					totalEvents: data.total_events || 0,
					googleEvents: data.google_events || 0,
					localEvents: data.local_events || 0,
					dateRange: data.date_range || null,
					lastSync: data.last_sync || null
				};
			}
		} catch {
			// stats unavailable — non-critical
		}
	}

	async function loadConnectionStatus() {
		try {
			const status = await getCalendarConnectionStatus();
			connectionStatus = { connected: status.connected, email: status.email };
		} catch {
			connectionStatus = { connected: false };
		}
	}

	async function loadEvents() {
		try {
			const result = await api.getCalendarEvents({
				start: dateRange.start.toISOString(),
				end: dateRange.end.toISOString(),
				meetingType: selectedMeetingType || undefined
			});
			events = Array.isArray(result) ? result : [];
		} catch {
			events = [];
		}
	}

	// ── Navigation ────────────────────────────────────────────────────────────
	async function navigatePrev() {
		const d = new Date(currentDate);
		if (viewMode === 'day') d.setDate(d.getDate() - 1);
		else if (viewMode === 'week') d.setDate(d.getDate() - 7);
		else d.setMonth(d.getMonth() - 1);
		currentDate = d;
		selectedDay = new Date(d);
		await tick();
		await loadEvents();
	}

	async function navigateNext() {
		const d = new Date(currentDate);
		if (viewMode === 'day') d.setDate(d.getDate() + 1);
		else if (viewMode === 'week') d.setDate(d.getDate() + 7);
		else d.setMonth(d.getMonth() + 1);
		currentDate = d;
		selectedDay = new Date(d);
		await tick();
		await loadEvents();
	}

	async function navigateToday() {
		currentDate = new Date();
		selectedDay = new Date();
		await tick();
		await loadEvents();
	}

	async function goToDayView(date: Date) {
		currentDate = new Date(date);
		selectedDay = new Date(date);
		viewMode = 'day';
		await tick();
		await loadEvents();
	}

	async function jumpToFirstEvent() {
		if (syncStats?.dateRange?.from) {
			currentDate = new Date(syncStats.dateRange.from);
			await tick();
			await loadEvents();
		}
	}

	async function jumpToLatestEvent() {
		if (syncStats?.dateRange?.to) {
			currentDate = new Date(syncStats.dateRange.to);
			await tick();
			await loadEvents();
		}
	}

	// ── Google Sync ───────────────────────────────────────────────────────────
	async function connectCalendar() {
		try {
			const result = await getCalendarAuthUrl();
			if (result.auth_url) window.location.href = result.auth_url;
		} catch (error) {
			console.error('Error initiating calendar auth:', error);
		}
	}

	async function syncCalendar() {
		isSyncing = true;
		try {
			const result = (await api.syncCalendar()) as {
				message: string;
				synced_count: number;
				details?: { total_events?: number; date_range?: string };
			};
			await loadEvents();
			await loadSyncStats();
			if (result?.details) {
				let dr: { from: string | null; to: string | null } | null = null;
				if (result.details.date_range) {
					const parts = result.details.date_range.split(' - ');
					dr = { from: parts[0] || null, to: parts[1] || null };
				}
				syncStats = {
					totalEvents: result.details.total_events || result.synced_count || 0,
					dateRange: dr,
					lastSync: new Date().toISOString(),
					googleEvents: 0,
					localEvents: 0
				};
			}
		} catch (error) {
			console.error('Error syncing calendar:', error);
		} finally {
			isSyncing = false;
		}
	}

	// ── Event Modal ───────────────────────────────────────────────────────────
	function openEventModal(event: CalendarEvent) {
		selectedEvent = event;
		showEventModal = true;
	}

	function closeEventModal() {
		selectedEvent = null;
		showEventModal = false;
	}

	// ── Create/Edit Modal ─────────────────────────────────────────────────────
	function openCreateModal(date?: Date) {
		editingEvent = null;
		const d = date || new Date();
		const dateStr = d.toISOString().split('T')[0];
		formData = {
			title: '',
			description: '',
			start_date: dateStr,
			start_time: '09:00',
			end_date: dateStr,
			end_time: '10:00',
			all_day: false,
			location: '',
			meeting_type: '',
			meeting_link: ''
		};
		formError = '';
		showCreateModal = true;
	}

	function openCreateModalAtHour(date: Date, hour: number) {
		editingEvent = null;
		const dateStr = date.toISOString().split('T')[0];
		const startTime = hour.toString().padStart(2, '0') + ':00';
		const endTime = ((hour + 1) % 24).toString().padStart(2, '0') + ':00';
		formData = {
			title: '',
			description: '',
			start_date: dateStr,
			start_time: startTime,
			end_date: dateStr,
			end_time: endTime,
			all_day: false,
			location: '',
			meeting_type: '',
			meeting_link: ''
		};
		formError = '';
		showCreateModal = true;
	}

	function openEditModal(event: CalendarEvent) {
		editingEvent = event;
		const start = new Date(event.start_time);
		const end = new Date(event.end_time);
		formData = {
			title: event.title || '',
			description: event.description || '',
			start_date: start.toISOString().split('T')[0],
			start_time: start.toTimeString().slice(0, 5),
			end_date: end.toISOString().split('T')[0],
			end_time: end.toTimeString().slice(0, 5),
			all_day: event.all_day || false,
			location: event.location || '',
			meeting_type: event.meeting_type || '',
			meeting_link: event.meeting_link || ''
		};
		formError = '';
		showCreateModal = true;
	}

	async function saveEvent() {
		if (!formData.title.trim()) {
			formError = 'Title is required';
			return;
		}
		isSaving = true;
		formError = '';
		try {
			const startTime = new Date(`${formData.start_date}T${formData.start_time}:00`).toISOString();
			const endTime = new Date(`${formData.end_date}T${formData.end_time}:00`).toISOString();
			const payload = {
				title: formData.title,
				description: formData.description || undefined,
				start_time: startTime,
				end_time: endTime,
				all_day: formData.all_day,
				location: formData.location || undefined,
				meeting_type: formData.meeting_type || undefined,
				meeting_link: formData.meeting_link || undefined,
				sync_to_google: connectionStatus?.connected || false
			};
			if (editingEvent) {
				const res = await apiClient.put(`/calendar/events/${editingEvent.id}`, payload);
				if (!res.ok) throw new Error('Failed to update event');
			} else {
				const res = await apiClient.post('/calendar/events', payload);
				if (!res.ok) throw new Error('Failed to create event');
			}
			showCreateModal = false;
			await loadEvents();
		} catch (error) {
			console.error('Error saving event:', error);
			formError = 'Failed to save event';
		} finally {
			isSaving = false;
		}
	}

	async function deleteEvent(event: CalendarEvent) {
		if (!confirm(`Delete "${event.title}"?`)) return;
		try {
			const res = await apiClient.delete(`/calendar/events/${event.id}`);
			if (res.ok) {
				closeEventModal();
				await loadEvents();
			}
		} catch (error) {
			console.error('Error deleting event:', error);
		}
	}

	function handleEditFromModal(event: CalendarEvent) {
		closeEventModal();
		openEditModal(event);
	}
</script>

<div class="min-h-full flex flex-col overflow-auto">
	<!-- Page Header -->
	<div class="cal-header">
		<div class="cal-header__row">
			<div>
				<h1 class="cal-header__title">Calendar</h1>
				<p class="cal-header__sub">Manage your schedule and meetings</p>
			</div>
			<div class="flex items-center gap-2">
				<button onclick={() => openCreateModal()} class="btn-pill btn-pill-primary btn-pill-xs">
					<svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Add Event
				</button>
				{#if connectionStatus?.connected}
					<button
						onclick={syncCalendar}
						disabled={isSyncing}
						class="btn-pill btn-pill-ghost btn-pill-xs"
					>
						{#if isSyncing}
							<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Syncing...
						{:else}
							<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							Sync
						{/if}
					</button>
				{/if}
			</div>
		</div>
	</div>

	{#if isLoading}
		<div class="cal-loading">
			<div class="cal-spinner"></div>
		</div>
	{:else}
		<CalendarStatusBanner
			{upcomingEvents}
			{syncStats}
			{events}
			onOpenEventModal={openEventModal}
			onJumpToFirstEvent={jumpToFirstEvent}
			onJumpToLatestEvent={jumpToLatestEvent}
		/>

		<CalendarToolbar
			{viewMode}
			headerText={headerText()}
			bind:selectedMeetingType
			{showSidebar}
			onNavigatePrev={navigatePrev}
			onNavigateNext={navigateNext}
			onNavigateToday={navigateToday}
			onViewModeChange={(mode) => { viewMode = mode; loadEvents(); }}
			onMeetingTypeChange={(type) => { selectedMeetingType = type; }}
			onToggleSidebar={() => (showSidebar = !showSidebar)}
		/>

		<!-- Main Content -->
		<div class="flex-1 flex min-h-0">
			{#if showSidebar}
				<CalendarSidebar
					bind:selectedDay
					{events}
					onSelectDay={(date) => { selectedDay = new Date(date); }}
					onGoToDayView={goToDayView}
					onOpenCreateModal={openCreateModal}
					onOpenEventModal={openEventModal}
				/>
			{/if}

			<!-- Calendar Grid -->
			<div class="flex-1 overflow-auto" bind:this={timeGridRef}>
				{#if viewMode === 'week'}
					<CalendarWeekDayView
						mode="week"
						{currentDate}
						weekDates={weekDates()}
						{events}
						{currentTime}
						onOpenEventModal={openEventModal}
						onOpenCreateModalAtHour={openCreateModalAtHour}
					/>
				{:else if viewMode === 'day'}
					<CalendarWeekDayView
						mode="day"
						{currentDate}
						weekDates={[]}
						{events}
						{currentTime}
						onOpenEventModal={openEventModal}
						onOpenCreateModalAtHour={openCreateModalAtHour}
					/>
				{:else if viewMode === 'month'}
					<CalendarMonthView
						{currentDate}
						{selectedDay}
						{events}
						{dateRange}
						onSelectDay={(date) => { selectedDay = new Date(date); }}
						onGoToDayView={goToDayView}
						onOpenCreateModal={openCreateModal}
						onOpenEventModal={openEventModal}
					/>
				{:else}
					<CalendarAgendaView
						{events}
						{syncStats}
						onOpenEventModal={openEventModal}
						onJumpToFirstEvent={jumpToFirstEvent}
					/>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Event Detail Modal -->
{#if showEventModal && selectedEvent}
	<CalendarEventModal
		event={selectedEvent}
		onClose={closeEventModal}
		onEdit={handleEditFromModal}
		onDelete={deleteEvent}
	/>
{/if}

<!-- Create/Edit Event Modal -->
{#if showCreateModal}
	<CalendarEventForm
		{editingEvent}
		bind:formData
		{isSaving}
		{formError}
		{connectionStatus}
		onSave={saveEvent}
		onClose={() => (showCreateModal = false)}
	/>
{/if}

<style>
	.cal-header {
		padding: 1rem 1.5rem;
		border-bottom: 1px solid var(--dbd2);
		flex-shrink: 0;
	}

	.cal-header__row {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.cal-header__title {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.cal-header__sub {
		font-size: 0.8rem;
		color: var(--dt3);
		margin: 0.15rem 0 0;
	}

	.cal-loading {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cal-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dt);
		border-top-color: transparent;
		border-radius: 50%;
		animation: cal-spin 0.6s linear infinite;
	}

	@keyframes cal-spin {
		to { transform: rotate(360deg); }
	}
</style>
