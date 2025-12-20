<script lang="ts">
	import { api, apiClient, type CalendarEvent, type GoogleConnectionStatus, type MeetingType, type ActionItem } from '$lib/api';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import CalendarEventCard from '$lib/components/calendar/CalendarEventCard.svelte';

	type ViewMode = 'week' | 'month';

	let viewMode = $state<ViewMode>('week');
	let currentDate = $state(new Date());
	let events = $state<CalendarEvent[]>([]);
	let isLoading = $state(true);
	let isSyncing = $state(false);
	let connectionStatus = $state<GoogleConnectionStatus | null>(null);
	let selectedEvent = $state<CalendarEvent | null>(null);
	let showEventModal = $state(false);

	// Create/Edit event modal
	let showCreateModal = $state(false);
	let editingEvent = $state<CalendarEvent | null>(null);
	let isSaving = $state(false);
	let formError = $state('');

	// Form state
	let formData = $state({
		title: '',
		description: '',
		start_date: '',
		start_time: '09:00',
		end_date: '',
		end_time: '10:00',
		all_day: false,
		location: '',
		meeting_type: '' as MeetingType | '',
		meeting_link: ''
	});

	// Meeting type filter
	let selectedMeetingType = $state<MeetingType | ''>('');

	// AI Processing state
	let meetingNotes = $state('');
	let meetingSummary = $state('');
	let actionItems = $state<ActionItem[]>([]);
	let isGeneratingSummary = $state(false);
	let isExtractingActions = $state(false);
	let showNotesSection = $state(false);

	const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const hours = Array.from({ length: 24 }, (_, i) => i);

	// Compute date range based on view mode
	const dateRange = $derived(() => {
		const start = new Date(currentDate);
		const end = new Date(currentDate);

		if (viewMode === 'week') {
			// Start of week (Sunday)
			start.setDate(start.getDate() - start.getDay());
			start.setHours(0, 0, 0, 0);
			// End of week (Saturday)
			end.setDate(start.getDate() + 6);
			end.setHours(23, 59, 59, 999);
		} else {
			// Start of month
			start.setDate(1);
			start.setHours(0, 0, 0, 0);
			// End of month
			end.setMonth(end.getMonth() + 1, 0);
			end.setHours(23, 59, 59, 999);
		}

		return { start, end };
	});

	// Week days for current view
	const weekDates = $derived(() => {
		const range = dateRange();
		const dates: Date[] = [];
		const current = new Date(range.start);

		if (viewMode === 'week') {
			for (let i = 0; i < 7; i++) {
				dates.push(new Date(current));
				current.setDate(current.getDate() + 1);
			}
		}

		return dates;
	});

	// Month calendar data
	const monthData = $derived(() => {
		if (viewMode !== 'month') return [];

		const range = dateRange();
		const firstDayOfMonth = range.start.getDay();
		const daysInMonth = new Date(range.end).getDate();

		const weeks: Date[][] = [];
		let currentWeek: Date[] = [];

		// Add empty cells for days before the first of the month
		for (let i = 0; i < firstDayOfMonth; i++) {
			const prevDate = new Date(range.start);
			prevDate.setDate(prevDate.getDate() - (firstDayOfMonth - i));
			currentWeek.push(prevDate);
		}

		// Add days of the month
		for (let day = 1; day <= daysInMonth; day++) {
			const date = new Date(currentDate.getFullYear(), currentDate.getMonth(), day);
			currentWeek.push(date);

			if (currentWeek.length === 7) {
				weeks.push(currentWeek);
				currentWeek = [];
			}
		}

		// Fill remaining days
		if (currentWeek.length > 0) {
			const nextMonth = new Date(range.end);
			nextMonth.setDate(nextMonth.getDate() + 1);
			while (currentWeek.length < 7) {
				currentWeek.push(new Date(nextMonth));
				nextMonth.setDate(nextMonth.getDate() + 1);
			}
			weeks.push(currentWeek);
		}

		return weeks;
	});

	// Format header based on view mode
	const headerText = $derived(() => {
		if (viewMode === 'week') {
			const range = dateRange();
			const startMonth = range.start.toLocaleString('default', { month: 'short' });
			const endMonth = range.end.toLocaleString('default', { month: 'short' });
			const year = range.start.getFullYear();

			if (startMonth === endMonth) {
				return `${startMonth} ${range.start.getDate()} - ${range.end.getDate()}, ${year}`;
			}
			return `${startMonth} ${range.start.getDate()} - ${endMonth} ${range.end.getDate()}, ${year}`;
		}

		return currentDate.toLocaleString('default', { month: 'long', year: 'numeric' });
	});

	onMount(async () => {
		await loadConnectionStatus();
		await loadEvents(); // Always load events, regardless of Google connection
		isLoading = false;
	});

	async function loadConnectionStatus() {
		try {
			connectionStatus = await api.getGoogleConnectionStatus();
		} catch (error) {
			console.error('Error loading connection status:', error);
			connectionStatus = { connected: false };
		}
	}

	async function loadEvents() {
		const range = dateRange();
		try {
			events = await api.getCalendarEvents({
				start: range.start.toISOString(),
				end: range.end.toISOString(),
				meetingType: selectedMeetingType || undefined
			});
		} catch (error) {
			console.error('Error loading events:', error);
			events = [];
		}
	}

	// Open create modal
	function openCreateModal(date?: Date) {
		editingEvent = null;
		const targetDate = date || new Date();
		const dateStr = targetDate.toISOString().split('T')[0];
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

	// Open edit modal
	function openEditModal(event: CalendarEvent) {
		editingEvent = event;
		const startDate = new Date(event.start_time);
		const endDate = new Date(event.end_time);
		formData = {
			title: event.title || '',
			description: event.description || '',
			start_date: startDate.toISOString().split('T')[0],
			start_time: startDate.toTimeString().slice(0, 5),
			end_date: endDate.toISOString().split('T')[0],
			end_time: endDate.toTimeString().slice(0, 5),
			all_day: event.all_day || false,
			location: event.location || '',
			meeting_type: event.meeting_type || '',
			meeting_link: event.meeting_link || ''
		};
		formError = '';
		showCreateModal = true;
	}

	// Save event (create or update)
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
				// Update existing event
				const res = await apiClient.put(`/calendar/events/${editingEvent.id}`, payload);
				if (!res.ok) throw new Error('Failed to update event');
			} else {
				// Create new event
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

	// Delete event
	async function deleteEvent(event: CalendarEvent) {
		if (!confirm(`Delete "${event.title}"?`)) return;

		try {
			const res = await apiClient.delete(`/calendar/events/${event.id}`);
			if (res.ok) {
				showEventModal = false;
				await loadEvents();
			}
		} catch (error) {
			console.error('Error deleting event:', error);
		}
	}

	// Open event modal with notes section
	function openEventModal(event: CalendarEvent) {
		selectedEvent = event;
		meetingNotes = event.meeting_notes || '';
		meetingSummary = event.meeting_summary || '';
		actionItems = event.action_items || [];
		showNotesSection = !!(event.meeting_notes || event.meeting_summary || (event.action_items && event.action_items.length > 0));
		showEventModal = true;
	}

	// AI Processing functions
	async function generateSummary() {
		if (!meetingNotes.trim()) {
			alert('Please add meeting notes first');
			return;
		}

		isGeneratingSummary = true;
		try {
			const response = await fetch('/api/chat/message', {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					message: `Please summarize the following meeting notes in 2-3 concise bullet points. Focus on key decisions and outcomes.\n\nMeeting: ${selectedEvent?.title || 'Meeting'}\n\nNotes:\n${meetingNotes}`,
					model: 'llama3.2:latest', // Use local model
					stream: false
				})
			});

			if (response.ok) {
				const reader = response.body?.getReader();
				const decoder = new TextDecoder();
				let fullContent = '';

				if (reader) {
					while (true) {
						const { done, value } = await reader.read();
						if (done) break;
						fullContent += decoder.decode(value, { stream: true });
					}
				}

				meetingSummary = fullContent.trim();

				// Save summary to event
				if (selectedEvent) {
					await apiClient.put(`/calendar/events/${selectedEvent.id}`, {
						meeting_summary: meetingSummary
					});
				}
			}
		} catch (error) {
			console.error('Error generating summary:', error);
			alert('Failed to generate summary. Please try again.');
		} finally {
			isGeneratingSummary = false;
		}
	}

	async function extractActionItems() {
		if (!meetingNotes.trim()) {
			alert('Please add meeting notes first');
			return;
		}

		isExtractingActions = true;
		try {
			const response = await apiClient.post('/chat/ai/extract-tasks', {
				content: meetingNotes
			});

			if (response.ok) {
				const data = await response.json();
				if (data.tasks && Array.isArray(data.tasks)) {
					actionItems = data.tasks.map((t: { title: string; description?: string }, index: number) => ({
						id: `action-${index}-${Date.now()}`,
						text: t.title,
						completed: false
					}));

					// Save action items to event
					if (selectedEvent) {
						await apiClient.put(`/calendar/events/${selectedEvent.id}`, {
							action_items: actionItems
						});
					}
				}
			}
		} catch (error) {
			console.error('Error extracting action items:', error);
			alert('Failed to extract action items. Please try again.');
		} finally {
			isExtractingActions = false;
		}
	}

	async function saveNotes() {
		if (!selectedEvent) return;

		try {
			await apiClient.put(`/calendar/events/${selectedEvent.id}`, {
				notes: meetingNotes
			});
		} catch (error) {
			console.error('Error saving notes:', error);
		}
	}

	async function createTaskFromActionItem(item: ActionItem) {
		try {
			const response = await apiClient.post('/tasks', {
				title: item.text,
				description: `From meeting: ${selectedEvent?.title || 'Meeting'}`,
				status: 'todo',
				priority: 'medium',
				assignee_id: item.assignee_id,
				due_date: item.due_date
			});

			if (response.ok) {
				alert('Task created successfully!');
			}
		} catch (error) {
			console.error('Error creating task:', error);
			alert('Failed to create task');
		}
	}

	async function syncCalendar() {
		isSyncing = true;
		try {
			await api.syncCalendar();
			await loadEvents();
		} catch (error) {
			console.error('Error syncing calendar:', error);
		} finally {
			isSyncing = false;
		}
	}

	function navigatePrev() {
		const newDate = new Date(currentDate);
		if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() - 7);
		} else {
			newDate.setMonth(newDate.getMonth() - 1);
		}
		currentDate = newDate;
		loadEvents();
	}

	function navigateNext() {
		const newDate = new Date(currentDate);
		if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() + 7);
		} else {
			newDate.setMonth(newDate.getMonth() + 1);
		}
		currentDate = newDate;
		loadEvents();
	}

	function navigateToday() {
		currentDate = new Date();
		loadEvents();
	}

	function getEventsForDate(date: Date): CalendarEvent[] {
		return events.filter((event) => {
			const eventDate = new Date(event.start_time);
			return (
				eventDate.getFullYear() === date.getFullYear() &&
				eventDate.getMonth() === date.getMonth() &&
				eventDate.getDate() === date.getDate()
			);
		});
	}

	function getEventsForHour(date: Date, hour: number): CalendarEvent[] {
		return events.filter((event) => {
			const eventStart = new Date(event.start_time);
			return (
				eventStart.getFullYear() === date.getFullYear() &&
				eventStart.getMonth() === date.getMonth() &&
				eventStart.getDate() === date.getDate() &&
				eventStart.getHours() === hour
			);
		});
	}

	function isToday(date: Date): boolean {
		const today = new Date();
		return (
			date.getFullYear() === today.getFullYear() &&
			date.getMonth() === today.getMonth() &&
			date.getDate() === today.getDate()
		);
	}

	function isCurrentMonth(date: Date): boolean {
		return date.getMonth() === currentDate.getMonth();
	}

	function formatHour(hour: number): string {
		if (hour === 0) return '12 AM';
		if (hour === 12) return '12 PM';
		return hour > 12 ? `${hour - 12} PM` : `${hour} AM`;
	}

	function closeEventModal() {
		selectedEvent = null;
		showEventModal = false;
	}

	// Reload events when filter changes
	$effect(() => {
		selectedMeetingType; // track dependency
		loadEvents();
	});

	// Open create modal at specific hour
	function openCreateModalAtHour(date: Date, hour: number) {
		const targetDate = new Date(date);
		targetDate.setHours(hour, 0, 0, 0);
		const dateStr = targetDate.toISOString().split('T')[0];
		const timeStr = hour.toString().padStart(2, '0') + ':00';
		const endHour = (hour + 1) % 24;
		const endTimeStr = endHour.toString().padStart(2, '0') + ':00';

		editingEvent = null;
		formData = {
			title: '',
			description: '',
			start_date: dateStr,
			start_time: timeStr,
			end_date: dateStr,
			end_time: endTimeStr,
			all_day: false,
			location: '',
			meeting_type: '',
			meeting_link: ''
		};
		formError = '';
		showCreateModal = true;
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 flex-shrink-0">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-xl font-semibold text-gray-900">Calendar</h1>
				<p class="text-sm text-gray-500 mt-0.5">Manage your schedule and meetings</p>
			</div>
			<div class="flex items-center gap-2">
				<button
					onclick={() => openCreateModal()}
					class="btn btn-primary text-sm"
				>
					<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Add Event
				</button>
				{#if connectionStatus?.connected}
					<button
						onclick={syncCalendar}
						disabled={isSyncing}
						class="btn btn-secondary text-sm disabled:opacity-50"
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
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Calendar Controls -->
		<div class="px-6 py-3 border-b border-gray-100 flex items-center justify-between flex-shrink-0">
			<div class="flex items-center gap-4">
				<div class="flex items-center gap-1">
					<button
						onclick={navigatePrev}
						class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button
						onclick={navigateNext}
						class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				</div>
				<button
					onclick={navigateToday}
					class="px-3 py-1.5 text-sm font-medium border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Today
				</button>
				<h2 class="text-lg font-semibold text-gray-900">{headerText()}</h2>
			</div>
			<div class="flex items-center gap-3">
				<!-- Meeting Type Filter -->
				<select
					bind:value={selectedMeetingType}
					class="input text-sm py-1.5 w-40"
				>
					<option value="">All types</option>
					<option value="team">Team</option>
					<option value="sales">Sales</option>
					<option value="client">Client</option>
					<option value="onboarding">Onboarding</option>
					<option value="kickoff">Kickoff</option>
					<option value="implementation">Implementation</option>
					<option value="standup">Standup</option>
					<option value="planning">Planning</option>
					<option value="review">Review</option>
					<option value="one_on_one">1:1</option>
					<option value="retrospective">Retrospective</option>
					<option value="internal">Internal</option>
					<option value="external">External</option>
					<option value="other">Other</option>
				</select>

				<!-- View Mode Toggle -->
				<div class="flex items-center bg-gray-100 rounded-lg p-0.5">
					<button
						onclick={() => { viewMode = 'week'; loadEvents(); }}
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'week' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
					>
						Week
					</button>
					<button
						onclick={() => { viewMode = 'month'; loadEvents(); }}
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'month' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
					>
						Month
					</button>
				</div>
			</div>
		</div>

		<!-- Calendar Grid -->
		<div class="flex-1 overflow-auto">
			{#if viewMode === 'week'}
				<!-- Week View -->
				<div class="min-w-[800px]">
					<!-- Day Headers -->
					<div class="grid grid-cols-8 border-b border-gray-200 sticky top-0 bg-white z-10">
						<div class="p-2 text-xs text-gray-500"></div>
						{#each weekDates() as date}
							<div class="p-2 text-center border-l border-gray-200">
								<p class="text-xs text-gray-500">{weekDays[date.getDay()]}</p>
								<p class="text-lg font-semibold {isToday(date) ? 'bg-gray-900 text-white w-8 h-8 rounded-full mx-auto flex items-center justify-center' : 'text-gray-900'}">
									{date.getDate()}
								</p>
							</div>
						{/each}
					</div>

					<!-- Time Grid -->
					<div class="relative">
						{#each hours as hour}
							<div class="grid grid-cols-8 border-b border-gray-100" style="height: 60px;">
								<div class="p-2 text-xs text-gray-400 text-right pr-3">
									{formatHour(hour)}
								</div>
								{#each weekDates() as date}
								<button
									type="button"
									onclick={() => openCreateModalAtHour(date, hour)}
									class="border-l border-gray-100 relative p-0.5 w-full h-full text-left hover:bg-gray-50 transition-colors cursor-pointer"
								>
										{#each getEventsForHour(date, hour) as event (event.id)}
											<div onclick={(e) => { e.stopPropagation(); openEventModal(event); }}>
												<CalendarEventCard
													{event}
													compact
													onClick={() => openEventModal(event)}
												/>
											</div>
										{/each}
								</button>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			{:else}
				<!-- Month View -->
				<div class="p-4">
					<!-- Day Headers -->
					<div class="grid grid-cols-7 mb-2">
						{#each weekDays as day}
							<div class="text-center text-sm font-medium text-gray-500 py-2">
								{day}
							</div>
						{/each}
					</div>

					<!-- Month Grid -->
					<div class="grid grid-cols-7 gap-1">
						{#each monthData().flat() as date}
							<button
								type="button"
								onclick={() => openCreateModal(date)}
								class="min-h-[100px] p-2 border rounded-lg text-left hover:border-gray-400 transition-colors {isCurrentMonth(date) ? 'bg-white' : 'bg-gray-50'} {isToday(date) ? 'ring-2 ring-gray-900' : 'border-gray-200'}"
							>
								<p class="text-sm font-medium {isCurrentMonth(date) ? 'text-gray-900' : 'text-gray-400'}">
									{date.getDate()}
								</p>
								<div class="mt-1 space-y-1">
									{#each getEventsForDate(date).slice(0, 3) as event (event.id)}
										<div onclick={(e) => { e.stopPropagation(); openEventModal(event); }}>
											<CalendarEventCard
												{event}
												compact
												onClick={() => openEventModal(event)}
											/>
										</div>
									{/each}
									{#if getEventsForDate(date).length > 3}
										<p class="text-xs text-gray-500 pl-2">
											+{getEventsForDate(date).length - 3} more
										</p>
									{/if}
								</div>
							</button>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Event Detail Modal -->
{#if showEventModal && selectedEvent}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" role="dialog" aria-modal="true">
		<div class="bg-white rounded-xl shadow-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-auto">
			<div class="p-6">
				<div class="flex items-start justify-between mb-4">
					<div>
						<h2 class="text-xl font-semibold text-gray-900">
							{selectedEvent.title || 'Untitled Event'}
						</h2>
						{#if selectedEvent.meeting_type && selectedEvent.meeting_type !== 'other'}
							<span class="inline-block mt-1 px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-700 rounded-full">
								{selectedEvent.meeting_type.replace('_', ' ')}
							</span>
						{/if}
					</div>
					<button
						onclick={closeEventModal}
						class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<div class="space-y-4">
					<!-- Time -->
					<div class="flex items-start gap-3">
						<svg class="w-5 h-5 text-gray-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<div>
							<p class="text-gray-900">
								{new Date(selectedEvent.start_time).toLocaleDateString('en-US', {
									weekday: 'long',
									month: 'long',
									day: 'numeric',
									year: 'numeric'
								})}
							</p>
							<p class="text-sm text-gray-500">
								{#if selectedEvent.all_day}
									All day
								{:else}
									{new Date(selectedEvent.start_time).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}
									-
									{new Date(selectedEvent.end_time).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}
								{/if}
							</p>
						</div>
					</div>

					<!-- Location -->
					{#if selectedEvent.location}
						<div class="flex items-start gap-3">
							<svg class="w-5 h-5 text-gray-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							<p class="text-gray-900">{selectedEvent.location}</p>
						</div>
					{/if}

					<!-- Meeting Link -->
					{#if selectedEvent.meeting_link}
						<div class="flex items-start gap-3">
							<svg class="w-5 h-5 text-gray-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							<a href={selectedEvent.meeting_link} target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline">
								Join Meeting
							</a>
						</div>
					{/if}

					<!-- Description -->
					{#if selectedEvent.description}
						<div class="flex items-start gap-3">
							<svg class="w-5 h-5 text-gray-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
							</svg>
							<p class="text-gray-700 whitespace-pre-wrap">{selectedEvent.description}</p>
						</div>
					{/if}

					<!-- Attendees -->
					{#if selectedEvent.attendees && selectedEvent.attendees.length > 0}
						<div class="flex items-start gap-3">
							<svg class="w-5 h-5 text-gray-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
							<div>
								<p class="text-sm font-medium text-gray-700 mb-1">{selectedEvent.attendees.length} Attendees</p>
								<div class="space-y-1">
									{#each selectedEvent.attendees as attendee}
										<p class="text-sm text-gray-600">{attendee.name || attendee.email}</p>
									{/each}
								</div>
							</div>
						</div>
					{/if}

					<!-- Google Calendar Link -->
					{#if selectedEvent.html_link}
						<div class="pt-4 border-t border-gray-200">
							<a
								href={selectedEvent.html_link}
								target="_blank"
								rel="noopener noreferrer"
								class="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700"
							>
								<svg class="w-4 h-4" viewBox="0 0 24 24">
									<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
									<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
									<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
									<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
								</svg>
								Open in Google Calendar
							</a>
						</div>
					{/if}

					<!-- Meeting Notes & AI Section -->
					<div class="pt-4 border-t border-gray-200">
						<button
							onclick={() => showNotesSection = !showNotesSection}
							class="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-gray-900 mb-3"
						>
							<svg class="w-4 h-4 transition-transform {showNotesSection ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
							<svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
							</svg>
							Meeting Notes & AI
						</button>

						{#if showNotesSection}
							<div class="space-y-4 animate-fadeIn">
								<!-- Meeting Notes -->
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-1">Meeting Notes / Transcription</label>
									<textarea
										bind:value={meetingNotes}
										onblur={saveNotes}
										placeholder="Add meeting notes, transcription, or paste voice recording text here..."
										rows="4"
										class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm resize-none focus:outline-none focus:ring-2 focus:ring-purple-500"
									></textarea>
								</div>

								<!-- AI Actions -->
								<div class="flex items-center gap-2">
									<button
										onclick={generateSummary}
										disabled={isGeneratingSummary || !meetingNotes.trim()}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-purple-700 bg-purple-50 hover:bg-purple-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
									>
										{#if isGeneratingSummary}
											<svg class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											Summarizing...
										{:else}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
											</svg>
											Summarize
										{/if}
									</button>

									<button
										onclick={extractActionItems}
										disabled={isExtractingActions || !meetingNotes.trim()}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
									>
										{#if isExtractingActions}
											<svg class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											Extracting...
										{:else}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
											</svg>
											Extract Actions
										{/if}
									</button>
								</div>

								<!-- Summary -->
								{#if meetingSummary}
									<div class="p-3 bg-purple-50 rounded-lg">
										<div class="flex items-center gap-2 text-sm font-medium text-purple-700 mb-2">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
											</svg>
											Summary
										</div>
										<p class="text-sm text-gray-700 whitespace-pre-wrap">{meetingSummary}</p>
									</div>
								{/if}

								<!-- Action Items -->
								{#if actionItems.length > 0}
									<div class="p-3 bg-blue-50 rounded-lg">
										<div class="flex items-center gap-2 text-sm font-medium text-blue-700 mb-2">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
											</svg>
											Action Items ({actionItems.length})
										</div>
										<ul class="space-y-2">
											{#each actionItems as item, index}
												<li class="flex items-start gap-2 text-sm text-gray-700">
													<span class="flex-shrink-0 w-5 h-5 flex items-center justify-center bg-blue-100 text-blue-600 rounded-full text-xs font-medium">{index + 1}</span>
													<span class="flex-1">{item.text}</span>
													<button
														onclick={() => createTaskFromActionItem(item)}
														class="flex-shrink-0 text-xs text-blue-600 hover:text-blue-800 font-medium"
														title="Create task"
													>
														+ Task
													</button>
												</li>
											{/each}
										</ul>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Action Buttons -->
					<div class="pt-4 border-t border-gray-200 flex items-center justify-end gap-2">
						<button
							onclick={() => { if (selectedEvent) { closeEventModal(); openEditModal(selectedEvent); } }}
							class="btn btn-secondary text-sm"
						>
							<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Edit
						</button>
						<button
							onclick={() => { if (selectedEvent) deleteEvent(selectedEvent); }}
							class="btn text-sm bg-red-50 text-red-600 hover:bg-red-100"
						>
							<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Create/Edit Event Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" role="dialog" aria-modal="true">
		<div class="bg-white rounded-xl shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-auto">
			<div class="p-6">
				<div class="flex items-center justify-between mb-6">
					<h2 class="text-xl font-semibold text-gray-900">
						{editingEvent ? 'Edit Event' : 'Create Event'}
					</h2>
					<button
						onclick={() => showCreateModal = false}
						class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				{#if formError}
					<div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
						{formError}
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); saveEvent(); }} class="space-y-4">
					<!-- Title -->
					<div>
						<label for="event-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
						<input
							id="event-title"
							type="text"
							bind:value={formData.title}
							placeholder="Event title"
							class="input w-full"
							required
						/>
					</div>

					<!-- Description -->
					<div>
						<label for="event-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
						<textarea
							id="event-description"
							bind:value={formData.description}
							placeholder="Add description..."
							rows="3"
							class="input w-full resize-none"
						></textarea>
					</div>

					<!-- All Day Toggle -->
					<div class="flex items-center gap-2">
						<input
							id="event-allday"
							type="checkbox"
							bind:checked={formData.all_day}
							class="w-4 h-4 text-gray-900 border-gray-300 rounded focus:ring-gray-900"
						/>
						<label for="event-allday" class="text-sm font-medium text-gray-700">All day event</label>
					</div>

					<!-- Date/Time -->
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="event-start-date" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
							<input
								id="event-start-date"
								type="date"
								bind:value={formData.start_date}
								class="input w-full"
								required
							/>
						</div>
						{#if !formData.all_day}
							<div>
								<label for="event-start-time" class="block text-sm font-medium text-gray-700 mb-1">Start Time</label>
								<input
									id="event-start-time"
									type="time"
									bind:value={formData.start_time}
									class="input w-full"
								/>
							</div>
						{/if}
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="event-end-date" class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
							<input
								id="event-end-date"
								type="date"
								bind:value={formData.end_date}
								class="input w-full"
								required
							/>
						</div>
						{#if !formData.all_day}
							<div>
								<label for="event-end-time" class="block text-sm font-medium text-gray-700 mb-1">End Time</label>
								<input
									id="event-end-time"
									type="time"
									bind:value={formData.end_time}
									class="input w-full"
								/>
							</div>
						{/if}
					</div>

					<!-- Location -->
					<div>
						<label for="event-location" class="block text-sm font-medium text-gray-700 mb-1">Location</label>
						<input
							id="event-location"
							type="text"
							bind:value={formData.location}
							placeholder="Add location..."
							class="input w-full"
						/>
					</div>

					<!-- Meeting Type -->
					<div>
						<label for="event-type" class="block text-sm font-medium text-gray-700 mb-1">Meeting Type</label>
						<select
							id="event-type"
							bind:value={formData.meeting_type}
							class="input w-full"
						>
							<option value="">No type</option>
							<option value="team">Team</option>
							<option value="sales">Sales</option>
							<option value="client">Client</option>
							<option value="onboarding">Onboarding</option>
							<option value="kickoff">Kickoff</option>
							<option value="implementation">Implementation</option>
							<option value="standup">Standup</option>
							<option value="planning">Planning</option>
							<option value="review">Review</option>
							<option value="one_on_one">1:1</option>
							<option value="retrospective">Retrospective</option>
							<option value="internal">Internal</option>
							<option value="external">External</option>
							<option value="other">Other</option>
						</select>
					</div>

					<!-- Meeting Link -->
					<div>
						<label for="event-link" class="block text-sm font-medium text-gray-700 mb-1">Meeting Link</label>
						<input
							id="event-link"
							type="url"
							bind:value={formData.meeting_link}
							placeholder="https://..."
							class="input w-full"
						/>
					</div>

					<!-- Google Sync Info -->
					{#if connectionStatus?.connected}
						<div class="flex items-center gap-2 text-sm text-gray-500">
							<svg class="w-4 h-4" viewBox="0 0 24 24">
								<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
								<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
								<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
								<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
							</svg>
							Will sync to Google Calendar
						</div>
					{/if}

					<!-- Submit Buttons -->
					<div class="flex items-center justify-end gap-2 pt-4">
						<button
							type="button"
							onclick={() => showCreateModal = false}
							class="btn btn-secondary"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={isSaving}
							class="btn btn-primary disabled:opacity-50"
						>
							{#if isSaving}
								<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Saving...
							{:else}
								{editingEvent ? 'Update Event' : 'Create Event'}
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
