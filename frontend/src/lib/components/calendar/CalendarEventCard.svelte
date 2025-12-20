<script lang="ts">
	import type { CalendarEvent, MeetingType } from '$lib/api/calendar';

	interface Props {
		event: CalendarEvent;
		compact?: boolean;
		onClick?: () => void;
	}

	let { event, compact = false, onClick }: Props = $props();

	function formatTime(dateString: string): string {
		return new Date(dateString).toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	function getMeetingTypeColor(type: MeetingType): { bg: string; text: string; border: string } {
		const colors: Record<MeetingType, { bg: string; text: string; border: string }> = {
			team: { bg: 'bg-blue-100', text: 'text-blue-800', border: 'border-blue-300' },
			sales: { bg: 'bg-green-100', text: 'text-green-800', border: 'border-green-300' },
			onboarding: { bg: 'bg-purple-100', text: 'text-purple-800', border: 'border-purple-300' },
			kickoff: { bg: 'bg-orange-100', text: 'text-orange-800', border: 'border-orange-300' },
			implementation: { bg: 'bg-cyan-100', text: 'text-cyan-800', border: 'border-cyan-300' },
			standup: { bg: 'bg-indigo-100', text: 'text-indigo-800', border: 'border-indigo-300' },
			retrospective: { bg: 'bg-pink-100', text: 'text-pink-800', border: 'border-pink-300' },
			planning: { bg: 'bg-yellow-100', text: 'text-yellow-800', border: 'border-yellow-300' },
			review: { bg: 'bg-teal-100', text: 'text-teal-800', border: 'border-teal-300' },
			one_on_one: { bg: 'bg-rose-100', text: 'text-rose-800', border: 'border-rose-300' },
			client: { bg: 'bg-emerald-100', text: 'text-emerald-800', border: 'border-emerald-300' },
			internal: { bg: 'bg-slate-100', text: 'text-slate-800', border: 'border-slate-300' },
			external: { bg: 'bg-amber-100', text: 'text-amber-800', border: 'border-amber-300' },
			other: { bg: 'bg-gray-100', text: 'text-gray-800', border: 'border-gray-300' }
		};
		return colors[type] || colors.other;
	}

	const colors = $derived(getMeetingTypeColor(event.meeting_type));
</script>

{#if compact}
	<button
		class="w-full text-left px-2 py-1 rounded text-xs truncate {colors.bg} {colors.text} hover:opacity-80 transition-opacity"
		onclick={onClick}
	>
		{event.all_day ? '' : formatTime(event.start_time) + ' '}{event.title || 'Untitled'}
	</button>
{:else}
	<button
		class="w-full text-left p-3 rounded-lg border {colors.border} {colors.bg} hover:shadow-sm transition-shadow"
		onclick={onClick}
	>
		<div class="flex items-start justify-between gap-2">
			<div class="flex-1 min-w-0">
				<p class="font-medium {colors.text} truncate">
					{event.title || 'Untitled Event'}
				</p>
				<p class="text-sm opacity-75 {colors.text}">
					{#if event.all_day}
						All day
					{:else}
						{formatTime(event.start_time)} - {formatTime(event.end_time)}
					{/if}
				</p>
				{#if event.location}
					<p class="text-xs mt-1 opacity-60 {colors.text} truncate flex items-center gap-1">
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						{event.location}
					</p>
				{/if}
			</div>
			{#if event.source === 'google'}
				<svg class="w-4 h-4 flex-shrink-0 opacity-50" viewBox="0 0 24 24">
					<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
					<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
					<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
					<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
				</svg>
			{/if}
		</div>
		{#if event.attendees && event.attendees.length > 0}
			<div class="mt-2 flex items-center gap-1">
				<svg class="w-3 h-3 opacity-50 {colors.text}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
				</svg>
				<span class="text-xs opacity-60 {colors.text}">
					{event.attendees.length} attendee{event.attendees.length > 1 ? 's' : ''}
				</span>
			</div>
		{/if}
	</button>
{/if}
