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

<div class="ct-bar">
	<!-- Left: Navigation -->
	<div class="ct-nav">
		<div class="ct-nav__arrows">
			<button
				onclick={onNavigatePrev}
				class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm"
				aria-label="Previous period"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
			<button
				onclick={onNavigateNext}
				class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm"
				aria-label="Next period"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		</div>
		<button
			onclick={onNavigateToday}
			class="btn-pill btn-pill-secondary btn-pill-xs"
		>
			Today
		</button>
		<h2 class="ct-heading">{headerText}</h2>
	</div>

	<!-- Right: Filters + View Toggle + Sidebar -->
	<div class="ct-controls">
		<div class="ct-select-wrap">
			<select
				bind:value={selectedMeetingType}
				onchange={(e) => onMeetingTypeChange((e.currentTarget as HTMLSelectElement).value as MeetingType | '')}
				class="ct-select"
				aria-label="Filter events by meeting type"
			>
				{#each meetingTypes as type (type.value)}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
			<svg class="ct-select__chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
			</svg>
		</div>

		<div class="ct-view-toggle">
			{#each viewModes as mode (mode)}
				<button
					onclick={() => onViewModeChange(mode)}
					class="btn-pill btn-pill-xs capitalize {viewMode === mode ? 'btn-pill-secondary' : 'btn-pill-ghost'}"
				>
					{mode}
				</button>
			{/each}
		</div>

		<button
			onclick={onToggleSidebar}
			class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm"
			aria-label={showSidebar ? 'Hide sidebar' : 'Show sidebar'}
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
			</svg>
		</button>
	</div>
</div>

<style>
	.ct-bar {
		padding: 0.5rem 1.25rem;
		border-bottom: 1px solid var(--dbd2);
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;
	}

	.ct-nav {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.ct-nav__arrows {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.ct-heading {
		font-size: 0.925rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.ct-controls {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	/* Styled select dropdown */
	.ct-select-wrap {
		position: relative;
	}

	.ct-select {
		appearance: none;
		-webkit-appearance: none;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 6px;
		color: var(--dt2);
		font-size: 0.75rem;
		font-family: inherit;
		padding: 0.3rem 1.75rem 0.3rem 0.625rem;
		cursor: pointer;
		outline: none;
		transition: border-color 0.15s;
		min-width: 8rem;
	}

	.ct-select:hover {
		border-color: var(--dt4);
	}

	.ct-select:focus {
		border-color: var(--dt3);
	}

	.ct-select option {
		background: var(--dbg);
		color: var(--dt);
	}

	.ct-select__chevron {
		position: absolute;
		right: 0.5rem;
		top: 50%;
		transform: translateY(-50%);
		width: 14px;
		height: 14px;
		color: var(--dt3);
		pointer-events: none;
	}

	/* View mode toggle group */
	.ct-view-toggle {
		display: flex;
		align-items: center;
		background: var(--dbg3);
		border-radius: 6px;
		padding: 2px;
	}
</style>
