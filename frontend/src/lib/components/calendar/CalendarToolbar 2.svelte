<script lang="ts">
	import type { MeetingType } from '$lib/api';
	import type { ViewMode } from './calendarUtils';

	interface Props {
		viewMode: ViewMode;
		headerText: string;
		selectedMeetingType: MeetingType | '';
		showSidebar: boolean;
		onNavigatePrev: () => void;
		onNavigateNext: () => void;
		onNavigateToday: () => void;
		onViewModeChange: (mode: ViewMode) => void;
		onMeetingTypeChange: (type: MeetingType | '') => void;
		onToggleSidebar: () => void;
	}

	let {
		viewMode,
		headerText,
		selectedMeetingType = $bindable(),
		showSidebar,
		onNavigatePrev,
		onNavigateNext,
		onNavigateToday,
		onViewModeChange,
		onMeetingTypeChange,
		onToggleSidebar
	}: Props = $props();

	const meetingTypes: Array<{ value: MeetingType | ''; label: string }> = [
		{ value: '', label: 'All types' },
		{ value: 'team', label: 'Team' },
		{ value: 'sales', label: 'Sales' },
		{ value: 'client', label: 'Client' },
		{ value: 'onboarding', label: 'Onboarding' },
		{ value: 'kickoff', label: 'Kickoff' },
		{ value: 'implementation', label: 'Implementation' },
		{ value: 'standup', label: 'Standup' },
		{ value: 'planning', label: 'Planning' },
		{ value: 'review', label: 'Review' },
		{ value: 'one_on_one', label: '1:1' },
		{ value: 'retrospective', label: 'Retrospective' },
		{ value: 'internal', label: 'Internal' },
		{ value: 'external', label: 'External' },
		{ value: 'other', label: 'Other' }
	];

	const viewModes: ViewMode[] = ['day', 'week', 'month', 'agenda'];
</script>

<div class="px-6 py-3 border-b border-gray-100 flex items-center justify-between flex-shrink-0">
	<!-- Left: Navigation -->
	<div class="flex items-center gap-4">
		<div class="flex items-center gap-1">
			<button
				onclick={onNavigatePrev}
				class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
				aria-label="Previous period"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
			<button
				onclick={onNavigateNext}
				class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
				aria-label="Next period"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		</div>
		<button
			onclick={onNavigateToday}
			class="px-3 py-1.5 text-sm font-medium border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
		>
			Today
		</button>
		<h2 class="text-lg font-semibold text-gray-900">{headerText}</h2>
	</div>

	<!-- Right: Filters + View Toggle + Sidebar -->
	<div class="flex items-center gap-3">
		<select
			bind:value={selectedMeetingType}
			onchange={(e) => onMeetingTypeChange((e.currentTarget as HTMLSelectElement).value as MeetingType | '')}
			class="input text-sm py-1.5 w-40"
			aria-label="Filter events by meeting type"
		>
			{#each meetingTypes as type (type.value)}
				<option value={type.value}>{type.label}</option>
			{/each}
		</select>

		<div class="flex items-center bg-gray-100 rounded-lg p-0.5">
			{#each viewModes as mode (mode)}
				<button
					onclick={() => onViewModeChange(mode)}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors capitalize
						{viewMode === mode ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
				>
					{mode}
				</button>
			{/each}
		</div>

		<button
			onclick={onToggleSidebar}
			class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
			aria-label={showSidebar ? 'Hide sidebar' : 'Show sidebar'}
		>
			<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
			</svg>
		</button>
	</div>
</div>
