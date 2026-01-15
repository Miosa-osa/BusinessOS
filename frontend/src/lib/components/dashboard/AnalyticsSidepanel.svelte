<script lang="ts">
	import { fly, fade, scale } from 'svelte/transition';
	import { tweened } from 'svelte/motion';
	import { cubicOut } from 'svelte/easing';
	
	interface Props {
		isOpen: boolean;
		onClose: () => void;
		analytics: DashboardAnalytics | null;
		isLoading?: boolean;
		onTimeRangeChange?: (range: TimeRange) => void;
	}
	
	type TimeRange = 'today' | 'week' | 'month' | '30days';
	
	interface DashboardAnalytics {
		focus: {
			completionRate: number;
			completedToday: number;
			totalToday: number;
			streak: number;
			avgCompletionTime: string;
			weeklyData: number[];
		};
		tasks: {
			completedThisWeek: number;
			dueToday: number;
			overdue: number;
			completionRate: number;
			byPriority: { critical: number; high: number; medium: number; low: number };
			weeklyData: number[];
		};
		projects: {
			active: number;
			completed: number;
			atRisk: number;
			onTimeRate: number;
			avgProgress: number;
		};
		activity: {
			totalActions: number;
			mostActiveDay: string;
			topActivityType: string;
			weeklyData: number[];
		};
	}
	
	let { isOpen, onClose, analytics, isLoading = false, onTimeRangeChange }: Props = $props();
	
	// Time range selector
	let timeRange = $state<TimeRange>('week');
	
	const timeRanges = [
		{ value: 'today', label: 'Today' },
		{ value: 'week', label: 'Week' },
		{ value: 'month', label: 'Month' },
		{ value: '30days', label: '30d' },
	];
	
	// Animated number stores
	const animatedFocusRate = tweened(0, { duration: 600, easing: cubicOut });
	const animatedTasksCompleted = tweened(0, { duration: 600, easing: cubicOut });
	const animatedActiveProjects = tweened(0, { duration: 600, easing: cubicOut });
	const animatedStreak = tweened(0, { duration: 600, easing: cubicOut });
	const animatedAvgProgress = tweened(0, { duration: 800, easing: cubicOut });
	const animatedTotalActions = tweened(0, { duration: 600, easing: cubicOut });
	
	// Update animated values when analytics change
	$effect(() => {
		if (analytics && isOpen) {
			animatedFocusRate.set(analytics.focus.completionRate);
			animatedTasksCompleted.set(analytics.tasks.completedThisWeek);
			animatedActiveProjects.set(analytics.projects.active);
			animatedStreak.set(analytics.focus.streak);
			animatedAvgProgress.set(analytics.projects.avgProgress);
			animatedTotalActions.set(analytics.activity.totalActions);
		}
	});
	
	// Reset animations when panel closes
	$effect(() => {
		if (!isOpen) {
			animatedFocusRate.set(0, { duration: 0 });
			animatedTasksCompleted.set(0, { duration: 0 });
			animatedActiveProjects.set(0, { duration: 0 });
			animatedStreak.set(0, { duration: 0 });
			animatedAvgProgress.set(0, { duration: 0 });
			animatedTotalActions.set(0, { duration: 0 });
		}
	});
	
	function handleTimeRangeChange(range: TimeRange) {
		timeRange = range;
		onTimeRangeChange?.(range);
	}
	
	// Check if data is empty
	function hasNoData(data: DashboardAnalytics | null): boolean {
		if (!data) return true;
		return (
			data.focus.completionRate === 0 &&
			data.tasks.completedThisWeek === 0 &&
			data.projects.active === 0 &&
			data.activity.totalActions === 0
		);
	}
	
	// Get time range label
	function getTimeRangeLabel(range: TimeRange): string {
		switch (range) {
			case 'today': return 'Today';
			case 'week': return 'This Week';
			case 'month': return 'This Month';
			case '30days': return 'Last 30 Days';
		}
	}
</script>

{#if isOpen}
	<!-- Backdrop -->
	<button
		class="fixed inset-0 bg-black/20 z-40"
		onclick={onClose}
		transition:fade={{ duration: 150 }}
		aria-label="Close analytics"
	></button>
	
	<!-- Sidepanel - Full screen on mobile, sidebar on desktop -->
	<div 
		class="fixed inset-0 sm:inset-auto sm:top-0 sm:right-0 sm:bottom-0 sm:w-[420px] bg-white sm:border-l border-gray-200 z-50 flex flex-col"
		transition:fly={{ x: 420, duration: 300 }}
	>
		<!-- Header -->
		<div class="flex items-center justify-between px-5 sm:px-6 py-4 sm:py-5 border-b border-gray-100">
			<div>
				<h2 class="text-lg font-semibold text-gray-900">Analytics</h2>
				<p class="text-sm text-gray-500 mt-0.5">Dashboard performance overview</p>
			</div>
			<button
				onclick={onClose}
				class="w-9 h-9 sm:w-8 sm:h-8 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
				aria-label="Close"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
		
		<!-- Time Range Selector -->
		<div class="px-5 sm:px-6 py-4 border-b border-gray-100">
			<div class="flex items-center gap-1 p-1 bg-gray-100 rounded-lg">
				{#each timeRanges as range}
					<button
						onclick={() => handleTimeRangeChange(range.value as TimeRange)}
						class="flex-1 px-3 py-2 sm:py-1.5 text-xs font-medium rounded-md transition-all
							{timeRange === range.value 
								? 'bg-white shadow-sm text-gray-900' 
								: 'text-gray-500 hover:text-gray-700'}"
					>
						{range.label}
					</button>
				{/each}
			</div>
		</div>
		
		<!-- Analytics Content -->
		<div class="flex-1 overflow-y-auto">
			{#if isLoading}
				<!-- Loading Skeleton -->
				<div class="px-5 sm:px-6 py-5 border-b border-gray-100">
					<div class="grid grid-cols-2 gap-3">
						{#each [1, 2, 3, 4] as _}
							<div class="p-4 bg-gray-50 rounded-xl border border-gray-100 animate-pulse">
								<div class="h-7 w-16 bg-gray-200 rounded mb-2"></div>
								<div class="h-4 w-24 bg-gray-200 rounded"></div>
							</div>
						{/each}
					</div>
				</div>
				
				<div class="px-5 sm:px-6 py-5 border-b border-gray-100">
					<div class="h-4 w-12 bg-gray-200 rounded mb-4 animate-pulse"></div>
					<div class="flex items-end gap-1.5 h-20 mb-4">
						{#each [1, 2, 3, 4, 5, 6, 7] as _, i}
							<div class="flex-1 bg-gray-200 rounded-t animate-pulse" style="height: {20 + Math.random() * 60}%"></div>
						{/each}
					</div>
					<div class="space-y-3">
						{#each [1, 2, 3] as _}
							<div class="flex justify-between animate-pulse">
								<div class="h-4 w-20 bg-gray-200 rounded"></div>
								<div class="h-4 w-8 bg-gray-200 rounded"></div>
							</div>
						{/each}
					</div>
				</div>
			{:else if hasNoData(analytics)}
				<!-- Empty State -->
				<div class="flex-1 flex flex-col items-center justify-center px-8 py-16 text-center">
					<div class="w-16 h-16 bg-gray-100 rounded-2xl flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
						</svg>
					</div>
					<h3 class="text-base font-medium text-gray-900 mb-1">No activity yet</h3>
					<p class="text-sm text-gray-500 max-w-[240px]">
						Complete your first focus item or task to start tracking your productivity.
					</p>
					<button
						onclick={onClose}
						class="mt-6 px-4 py-2 bg-gray-900 hover:bg-gray-800 text-white text-sm font-medium rounded-lg transition-colors"
					>
						Get Started
					</button>
				</div>
			{:else if analytics}
				<!-- Summary Cards -->
				<div class="px-5 sm:px-6 py-5 border-b border-gray-100">
					<div class="grid grid-cols-2 gap-3">
						<div class="p-4 bg-gray-50 rounded-xl border border-gray-100" transition:scale={{ delay: 50, duration: 200 }}>
							<div class="text-2xl font-semibold text-gray-900 tabular-nums">{Math.round($animatedFocusRate)}%</div>
							<div class="text-sm text-gray-500 mt-1">Focus Completion</div>
						</div>
						<div class="p-4 bg-gray-50 rounded-xl border border-gray-100" transition:scale={{ delay: 100, duration: 200 }}>
							<div class="text-2xl font-semibold text-gray-900 tabular-nums">{Math.round($animatedTasksCompleted)}</div>
							<div class="text-sm text-gray-500 mt-1">Tasks Completed</div>
						</div>
						<div class="p-4 bg-gray-50 rounded-xl border border-gray-100" transition:scale={{ delay: 150, duration: 200 }}>
							<div class="text-2xl font-semibold text-gray-900 tabular-nums">{Math.round($animatedActiveProjects)}</div>
							<div class="text-sm text-gray-500 mt-1">Active Projects</div>
						</div>
						<div class="p-4 bg-gray-50 rounded-xl border border-gray-100" transition:scale={{ delay: 200, duration: 200 }}>
							<div class="flex items-center gap-1.5">
								<span class="text-2xl font-semibold text-gray-900 tabular-nums">{Math.round($animatedStreak)}</span>
								<span class="text-lg">🔥</span>
							</div>
							<div class="text-sm text-gray-500 mt-1">Day Streak</div>
						</div>
					</div>
				</div>
				
				<!-- Tasks Section -->
				<div class="px-5 sm:px-6 py-5 border-b border-gray-100">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Tasks</h3>
						<span class="text-xs text-gray-400">{getTimeRangeLabel(timeRange)}</span>
					</div>
					
					<!-- Bar Chart - Polished -->
					<div class="mb-5 p-4 bg-gray-50 rounded-xl border border-gray-100">
						<div class="flex items-end justify-between gap-2" style="height: 80px;">
							{#each analytics.tasks.weeklyData as value, i}
								{@const max = Math.max(...analytics.tasks.weeklyData, 1)}
								{@const heightPx = Math.max(6, Math.round((value / max) * 60))}
								{@const isToday = i === new Date().getDay() - 1 || (i === 6 && new Date().getDay() === 0)}
								<div class="flex-1 flex flex-col items-center justify-end h-full">
									<span class="text-[10px] text-gray-500 tabular-nums mb-1">{value > 0 ? value : ''}</span>
									<div 
										class="w-full max-w-[32px] rounded-t-md transition-all duration-500 ease-out {isToday ? 'bg-gray-900' : 'bg-gray-400 hover:bg-gray-500'}"
										style="height: {heightPx}px; transition-delay: {i * 50}ms"
									></div>
								</div>
							{/each}
						</div>
						<div class="flex justify-between text-[10px] text-gray-400 mt-3 font-medium">
							<span>Mon</span>
							<span>Tue</span>
							<span>Wed</span>
							<span>Thu</span>
							<span>Fri</span>
							<span>Sat</span>
							<span>Sun</span>
						</div>
					</div>
					
					<!-- Task Stats -->
					<div class="space-y-0.5">
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">Due Today</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.tasks.dueToday}</span>
						</div>
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">Overdue</span>
							<div class="flex items-center gap-2">
								{#if analytics.tasks.overdue > 0}
									<span class="w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse"></span>
								{/if}
								<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.tasks.overdue}</span>
							</div>
						</div>
						<div class="flex items-center justify-between py-2.5">
							<span class="text-sm text-gray-600">Completion Rate</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.tasks.completionRate}%</span>
						</div>
					</div>
				</div>
				
				<!-- Projects Section -->
				<div class="px-5 sm:px-6 py-5 border-b border-gray-100">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Projects</h3>
					</div>
					
					<!-- Progress Ring - Animated -->
					<div class="flex items-center gap-4 mb-5 p-4 bg-gray-50 rounded-xl border border-gray-100">
						<div class="relative w-16 h-16 flex-shrink-0">
							<svg class="w-16 h-16 transform -rotate-90">
								<circle cx="32" cy="32" r="28" stroke="#e5e7eb" stroke-width="5" fill="none" />
								<circle 
									cx="32" cy="32" r="28" 
									stroke="#111827" 
									stroke-width="5" 
									fill="none"
									stroke-dasharray="{$animatedAvgProgress * 1.76} 176"
									stroke-linecap="round"
									class="transition-all duration-700"
								/>
							</svg>
							<div class="absolute inset-0 flex items-center justify-center">
								<span class="text-sm font-semibold text-gray-900 tabular-nums">{Math.round($animatedAvgProgress)}%</span>
							</div>
						</div>
						<div>
							<div class="text-sm font-medium text-gray-900">Average Progress</div>
							<div class="text-xs text-gray-500 mt-0.5">Across {analytics.projects.active} active projects</div>
						</div>
					</div>
					
					<!-- Project Stats -->
					<div class="space-y-0.5">
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">Active</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.projects.active}</span>
						</div>
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">Completed</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.projects.completed}</span>
						</div>
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">At Risk</span>
							<div class="flex items-center gap-2">
								{#if analytics.projects.atRisk > 0}
									<span class="w-1.5 h-1.5 bg-amber-500 rounded-full"></span>
								{/if}
								<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.projects.atRisk}</span>
							</div>
						</div>
						<div class="flex items-center justify-between py-2.5">
							<span class="text-sm text-gray-600">On-time Delivery</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{analytics.projects.onTimeRate}%</span>
						</div>
					</div>
				</div>
				
				<!-- Activity Section -->
				<div class="px-5 sm:px-6 py-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Activity</h3>
					</div>
					
					<!-- Activity Chart - Gradient bars -->
					<div class="mb-5 p-4 bg-gray-50 rounded-xl border border-gray-100">
						<div class="flex items-end justify-between gap-2" style="height: 56px;">
							{#each analytics.activity.weeklyData as value, i}
								{@const max = Math.max(...analytics.activity.weeklyData, 1)}
								{@const heightPx = Math.max(4, Math.round((value / max) * 48))}
								{@const intensity = max > 0 ? value / max : 0}
								<div class="flex-1 flex flex-col items-center justify-end h-full">
									<div 
										class="w-full max-w-[28px] rounded-t-md transition-all duration-500 ease-out"
										style="height: {heightPx}px; background-color: rgba(17, 24, 39, {0.25 + intensity * 0.75}); transition-delay: {i * 50}ms"
									></div>
								</div>
							{/each}
						</div>
						<div class="flex justify-between text-[10px] text-gray-400 mt-3 font-medium">
							<span>M</span>
							<span>T</span>
							<span>W</span>
							<span>T</span>
							<span>F</span>
							<span>S</span>
							<span>S</span>
						</div>
					</div>
					
					<!-- Activity Stats -->
					<div class="space-y-0.5">
						<div class="flex items-center justify-between py-2.5 border-b border-gray-100">
							<span class="text-sm text-gray-600">Total Actions</span>
							<span class="text-sm font-medium text-gray-900 tabular-nums">{Math.round($animatedTotalActions)}</span>
						</div>
						<div class="flex items-center justify-between py-2.5">
							<span class="text-sm text-gray-600">Most Active Day</span>
							<span class="text-sm font-medium text-gray-900">{analytics.activity.mostActiveDay}</span>
						</div>
					</div>
				</div>
			{/if}
		</div>
		
		<!-- Footer -->
		{#if !isLoading && !hasNoData(analytics)}
			<div class="px-5 sm:px-6 py-4 border-t border-gray-100">
				<button
					class="w-full flex items-center justify-center gap-2 px-4 py-3 sm:py-2.5 bg-gray-900 hover:bg-gray-800 active:bg-gray-950 text-white rounded-lg text-sm font-medium transition-colors"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
					Ask AI About This Data
				</button>
			</div>
		{/if}
	</div>
{/if}
