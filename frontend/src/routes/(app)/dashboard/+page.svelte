<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { fade, fly, scale } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { DropdownMenu } from 'bits-ui';
	import {
		DashboardHeader,
		TodaysFocusWidget,
		QuickActionsWidget,
		ActiveProjectsWidget,
		MyTasksWidget,
		RecentActivityWidget,
		AnalyticsSidepanel
	} from '$lib/components/dashboard';
	import { NotificationDropdown } from '$lib/components/notifications';
	import {
		api,
		type DashboardProject,
		type DashboardTask,
		type DashboardActivity
	} from '$lib/api';

	const session = useSession();

	// ============================================================================
	// DASHBOARD EDITOR STATE
	// ============================================================================
	
	type WidgetType = 'focus' | 'quick-actions' | 'projects' | 'tasks' | 'activity' | 'metric';
	type WidgetSize = 'small' | 'medium' | 'large';
	
	interface Widget {
		id: string;
		type: WidgetType;
		title: string;
		size: WidgetSize;
		config?: Record<string, unknown>;
		collapsed?: boolean;
		accentColor?: string;
		showAnalytics?: boolean; // For per-widget analytics flip
	}
	
	// Accent color options
	const accentColors = [
		{ name: 'Default', value: '' },
		{ name: 'Blue', value: 'blue' },
		{ name: 'Green', value: 'green' },
		{ name: 'Purple', value: 'purple' },
		{ name: 'Orange', value: 'orange' },
		{ name: 'Pink', value: 'pink' },
	];
	
	// Edit mode state
	let isEditMode = $state(false);
	let draggedWidget = $state<string | null>(null);
	let showWidgetPicker = $state(false);
	let selectedWidgetIndex = $state<number>(-1); // For keyboard nav
	let pickerSelectedSize = $state<WidgetSize>('medium'); // Size picker in widget drawer
	
	// Analytics state
	let showAnalyticsSidepanel = $state(false);
	let analyticsLoading = $state(false);
	let analyticsTimeRange = $state<'today' | 'week' | 'month' | '30days'>('week');
	
	// Seeded analytics data (mock data for demonstration)
	let seededAnalytics = $state<{
		focus: { completionRate: number; completedToday: number; totalToday: number; streak: number; avgCompletionTime: string; weeklyData: number[] };
		tasks: { completedThisWeek: number; dueToday: number; overdue: number; completionRate: number; byPriority: { critical: number; high: number; medium: number; low: number }; weeklyData: number[] };
		projects: { active: number; completed: number; atRisk: number; onTimeRate: number; avgProgress: number };
		activity: { totalActions: number; mostActiveDay: string; topActivityType: string; weeklyData: number[] };
	} | null>({
		focus: {
			completionRate: 78,
			completedToday: 3,
			totalToday: 4,
			streak: 7,
			avgCompletionTime: '2.3 hrs',
			weeklyData: [4, 5, 6, 4, 3, 2, 0] // Mon-Sun
		},
		tasks: {
			completedThisWeek: 23,
			dueToday: 5,
			overdue: 2,
			completionRate: 82,
			byPriority: { critical: 2, high: 5, medium: 8, low: 3 },
			weeklyData: [5, 6, 8, 4, 0, 0, 0]
		},
		projects: {
			active: 3,
			completed: 2,
			atRisk: 1,
			onTimeRate: 85,
			avgProgress: 67
		},
		activity: {
			totalActions: 47,
			mostActiveDay: 'Wednesday',
			topActivityType: 'task_completed',
			weeklyData: [12, 15, 18, 8, 0, 0, 0]
		}
	});
	
	// Handle time range change with loading simulation
	async function handleAnalyticsTimeRangeChange(range: 'today' | 'week' | 'month' | '30days') {
		analyticsTimeRange = range;
		analyticsLoading = true;
		
		// TODO: Replace with actual API call
		// const response = await fetch(`/api/analytics?range=${range}`);
		// seededAnalytics = await response.json();
		
		// Simulate API call delay
		await new Promise(resolve => setTimeout(resolve, 600));
		
		// Mock different data based on range
		const multiplier = range === 'today' ? 0.3 : range === 'week' ? 1 : range === 'month' ? 2.5 : 4;
		seededAnalytics = {
			focus: {
				completionRate: Math.min(100, Math.round(78 * (range === 'today' ? 0.8 : 1))),
				completedToday: range === 'today' ? 2 : 3,
				totalToday: 4,
				streak: 7,
				avgCompletionTime: '2.3 hrs',
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 2] : [4, 5, 6, 4, 3, 2, 0]
			},
			tasks: {
				completedThisWeek: Math.round(23 * multiplier),
				dueToday: 5,
				overdue: range === 'today' ? 1 : 2,
				completionRate: Math.min(100, Math.round(82 * (0.9 + Math.random() * 0.2))),
				byPriority: { critical: 2, high: 5, medium: 8, low: 3 },
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 3] : [5, 6, 8, 4, 0, 0, 0].map(v => Math.round(v * multiplier))
			},
			projects: {
				active: 3,
				completed: Math.round(2 * multiplier),
				atRisk: 1,
				onTimeRate: Math.min(100, Math.round(85 * (0.9 + Math.random() * 0.2))),
				avgProgress: Math.min(100, Math.round(67 + (range === 'month' ? 10 : range === '30days' ? 15 : 0)))
			},
			activity: {
				totalActions: Math.round(47 * multiplier),
				mostActiveDay: range === 'today' ? 'Today' : 'Wednesday',
				topActivityType: 'task_completed',
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 8] : [12, 15, 18, 8, 0, 0, 0].map(v => Math.round(v * multiplier))
			}
		};
		
		analyticsLoading = false;
	}
	
	// Per-widget analytics data
	const widgetAnalytics: Record<WidgetType, { title: string; stats: { label: string; value: string | number; trend?: string }[] }> = {
		focus: {
			title: 'Focus Analytics',
			stats: [
				{ label: 'Completion Rate', value: '78%', trend: '+12%' },
				{ label: 'Avg Time per Item', value: '2.3 hrs' },
				{ label: 'Current Streak', value: '🔥 7 days' },
				{ label: 'Best Day', value: 'Tuesday' }
			]
		},
		'quick-actions': {
			title: 'Quick Actions Analytics',
			stats: [
				{ label: 'Most Used', value: 'New Task' },
				{ label: 'Actions Today', value: 12 },
				{ label: 'Time Saved', value: '~45 min' }
			]
		},
		projects: {
			title: 'Projects Analytics',
			stats: [
				{ label: 'Active Projects', value: 3 },
				{ label: 'Avg Progress', value: '67%' },
				{ label: 'On-time Rate', value: '85%', trend: '+5%' },
				{ label: 'At Risk', value: 1 }
			]
		},
		tasks: {
			title: 'Tasks Analytics',
			stats: [
				{ label: 'Completed This Week', value: 23, trend: '+18%' },
				{ label: 'Due Today', value: 5 },
				{ label: 'Overdue', value: 2 },
				{ label: 'Avg/Day', value: '4.6' }
			]
		},
		activity: {
			title: 'Activity Analytics',
			stats: [
				{ label: 'Total Actions', value: 47 },
				{ label: 'Most Active', value: 'Wednesday' },
				{ label: 'Top Activity', value: 'Completing Tasks' }
			]
		},
		metric: {
			title: 'Metric Analytics',
			stats: [
				{ label: 'Current Value', value: 8 },
				{ label: 'vs Yesterday', value: '+12%' }
			]
		}
	};
	
	// Toggle per-widget analytics view
	function toggleWidgetAnalytics(id: string) {
		widgets = widgets.map(w => 
			w.id === id ? { ...w, showAnalytics: !w.showAnalytics } : w
		);
	}
	
	// Undo stack for removed widgets
	let undoStack = $state<{ widget: Widget; index: number; timestamp: number }[]>([]);
	let showUndoToast = $state(false);
	let undoTimeoutId: ReturnType<typeof setTimeout> | null = null;
	
	// Default widget layout
	let widgets = $state<Widget[]>([
		{ id: 'w1', type: 'focus', title: "Today's Focus", size: 'medium' },
		{ id: 'w2', type: 'quick-actions', title: 'Quick Actions', size: 'small' },
		{ id: 'w3', type: 'projects', title: 'Active Projects', size: 'medium' },
		{ id: 'w4', type: 'tasks', title: 'My Tasks', size: 'medium' },
		{ id: 'w5', type: 'activity', title: 'Recent Activity', size: 'large' },
	]);
	
	// Available widget types for picker
	const availableWidgets: { type: WidgetType; title: string; description: string; icon: string }[] = [
		{ type: 'focus', title: "Today's Focus", description: 'Track your daily priorities', icon: '🎯' },
		{ type: 'quick-actions', title: 'Quick Actions', description: 'Common shortcuts', icon: '⚡' },
		{ type: 'projects', title: 'Active Projects', description: 'Project progress overview', icon: '📁' },
		{ type: 'tasks', title: 'My Tasks', description: 'Tasks due soon', icon: '✓' },
		{ type: 'activity', title: 'Recent Activity', description: 'Latest workspace activity', icon: '📊' },
		{ type: 'metric', title: 'Metric Card', description: 'Single KPI display', icon: '📈' },
	];
	
	// Widget type categories for hybrid reusability policy
	const uniqueWidgetTypes: WidgetType[] = ['focus', 'quick-actions', 'activity'];
	const configurableWidgetTypes: WidgetType[] = ['metric', 'projects', 'tasks'];
	
	// Track which unique widget types are already on dashboard
	const addedUniqueTypes = $derived(new Set(
		widgets.filter(w => uniqueWidgetTypes.includes(w.type)).map(w => w.type)
	));
	
	// Check if a widget type can be added
	function canAddWidget(type: WidgetType): boolean {
		if (uniqueWidgetTypes.includes(type)) {
			return !addedUniqueTypes.has(type);
		}
		return true; // Configurable widgets can always be added
	}

	// Dashboard state
	let energyLevel = $state<number | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Focus items from API
	let focusItems = $state<{ id: string; text: string; completed: boolean }[]>([]);

	// Projects, tasks, activities from API
	let projects = $state<
		{
			id: string;
			name: string;
			clientName?: string;
			projectType: string;
			dueDate?: string;
			progress: number;
			health: 'healthy' | 'at_risk' | 'critical';
			teamCount: number;
		}[]
	>([]);

	let tasks = $state<
		{
			id: string;
			title: string;
			projectName?: string;
			dueDate?: string;
			priority: 'critical' | 'high' | 'medium' | 'low';
			completed: boolean;
		}[]
	>([]);

	let activities = $state<
		{
			id: string;
			type:
				| 'task_completed'
				| 'task_started'
				| 'project_created'
				| 'project_updated'
				| 'conversation'
				| 'team'
				| 'artifact';
			description: string;
			actorName?: string;
			actorAvatar?: string;
			targetId?: string;
			targetType?: string;
			createdAt: string;
		}[]
	>([]);

	// Load dashboard data
	async function loadDashboard() {
		try {
			isLoading = true;
			error = null;

			const summary = await api.getDashboardSummary();

			// Transform focus items
			focusItems = summary.focus_items.map((item) => ({
				id: item.id,
				text: item.text,
				completed: item.completed
			}));

			// Transform projects
			projects = summary.projects.map((p) => ({
				id: p.id,
				name: p.name,
				clientName: p.client_name ?? undefined,
				projectType: p.project_type,
				dueDate: p.due_date ?? undefined,
				progress: p.progress,
				health: p.health,
				teamCount: p.team_count
			}));

			// Transform tasks
			tasks = summary.tasks.map((t) => ({
				id: t.id,
				title: t.title,
				projectName: t.project_name ?? undefined,
				dueDate: t.due_date ?? undefined,
				priority: t.priority,
				completed: t.completed
			}));

			// Transform activities
			activities = summary.activities.map((a) => ({
				id: a.id,
				type: a.type,
				description: a.description,
				actorName: a.actor_name ?? undefined,
				actorAvatar: a.actor_avatar ?? undefined,
				targetId: a.target_id ?? undefined,
				targetType: a.target_type ?? undefined,
				createdAt: a.created_at
			}));

			energyLevel = summary.energy_level;
		} catch (err) {
			console.error('Failed to load dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to load dashboard';
		} finally {
			isLoading = false;
		}
	}

	// Load on mount
	$effect(() => {
		if ($session.data) {
			loadDashboard();
		}
	});

	// ============================================================================
	// EDIT MODE FUNCTIONS
	// ============================================================================
	
	function toggleEditMode() {
		isEditMode = !isEditMode;
		if (!isEditMode) {
			showWidgetPicker = false;
			// TODO: Save layout to backend
		}
	}
	
	function addWidget(type: WidgetType) {
		// Prevent duplicate unique widgets
		if (!canAddWidget(type)) {
			return;
		}
		
		const template = availableWidgets.find(w => w.type === type);
		if (!template) return;
		
		// Generate a distinguishing suffix for configurable widgets
		const existingOfType = widgets.filter(w => w.type === type).length;
		const title = configurableWidgetTypes.includes(type) && existingOfType > 0
			? `${template.title} ${existingOfType + 1}`
			: template.title;
		
		const newWidget: Widget = {
			id: `w${Date.now()}`,
			type,
			title,
			size: pickerSelectedSize,
			collapsed: false,
			accentColor: '',
		};
		widgets = [...widgets, newWidget];
		showWidgetPicker = false;
		pickerSelectedSize = 'medium'; // Reset for next time
	}
	
	function removeWidget(id: string) {
		const index = widgets.findIndex(w => w.id === id);
		if (index === -1) return;
		
		const removedWidget = widgets[index];
		
		// Add to undo stack
		undoStack = [...undoStack, { widget: removedWidget, index, timestamp: Date.now() }];
		
		// Remove from widgets
		widgets = widgets.filter(w => w.id !== id);
		
		// Show undo toast
		showUndoToast = true;
		
		// Clear previous timeout if exists
		if (undoTimeoutId) clearTimeout(undoTimeoutId);
		
		// Auto-dismiss after 5 seconds
		undoTimeoutId = setTimeout(() => {
			showUndoToast = false;
			// Remove old items from undo stack
			undoStack = undoStack.filter(item => Date.now() - item.timestamp < 5000);
		}, 5000);
	}
	
	function undoRemove() {
		if (undoStack.length === 0) return;
		
		const lastRemoved = undoStack[undoStack.length - 1];
		undoStack = undoStack.slice(0, -1);
		
		// Restore widget at original position
		const newWidgets = [...widgets];
		newWidgets.splice(lastRemoved.index, 0, lastRemoved.widget);
		widgets = newWidgets;
		
		if (undoStack.length === 0) {
			showUndoToast = false;
			if (undoTimeoutId) clearTimeout(undoTimeoutId);
		}
	}
	
	function toggleWidgetCollapse(id: string) {
		widgets = widgets.map(w => 
			w.id === id ? { ...w, collapsed: !w.collapsed } : w
		);
	}
	
	function setWidgetSize(id: string, size: WidgetSize) {
		widgets = widgets.map(w => 
			w.id === id ? { ...w, size } : w
		);
	}
	
	function setWidgetAccentColor(id: string, color: string) {
		widgets = widgets.map(w => 
			w.id === id ? { ...w, accentColor: color } : w
		);
	}
	
	function moveWidget(fromIndex: number, toIndex: number) {
		const newWidgets = [...widgets];
		const [moved] = newWidgets.splice(fromIndex, 1);
		newWidgets.splice(toIndex, 0, moved);
		widgets = newWidgets;
	}
	
	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		// Don't trigger if typing in an input
		if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
		
		if (e.key === 'e' || e.key === 'E') {
			e.preventDefault();
			toggleEditMode();
		}
		if (e.key === 'Escape' && isEditMode) {
			e.preventDefault();
			isEditMode = false;
			showWidgetPicker = false;
			selectedWidgetIndex = -1;
		}
		
		// Keyboard navigation in edit mode
		if (isEditMode && widgets.length > 0) {
			if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
				e.preventDefault();
				selectedWidgetIndex = (selectedWidgetIndex + 1) % widgets.length;
			}
			if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
				e.preventDefault();
				selectedWidgetIndex = selectedWidgetIndex <= 0 ? widgets.length - 1 : selectedWidgetIndex - 1;
			}
			if (e.key === 'Enter' && selectedWidgetIndex >= 0) {
				e.preventDefault();
				// Toggle collapse on Enter
				toggleWidgetCollapse(widgets[selectedWidgetIndex].id);
			}
			if (e.key === 'Delete' || e.key === 'Backspace') {
				if (selectedWidgetIndex >= 0) {
					e.preventDefault();
					removeWidget(widgets[selectedWidgetIndex].id);
					selectedWidgetIndex = Math.min(selectedWidgetIndex, widgets.length - 1);
				}
			}
		}
		
		// Undo shortcut (Ctrl+Z)
		if ((e.ctrlKey || e.metaKey) && e.key === 'z' && undoStack.length > 0) {
			e.preventDefault();
			undoRemove();
		}
	}
	
	// Simple drag handlers
	function handleDragStart(e: DragEvent, widgetId: string) {
		draggedWidget = widgetId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}
	
	function handleDragOver(e: DragEvent) {
		e.preventDefault();
	}
	
	function handleDrop(e: DragEvent, targetId: string) {
		e.preventDefault();
		if (!draggedWidget || draggedWidget === targetId) return;
		
		const fromIndex = widgets.findIndex(w => w.id === draggedWidget);
		const toIndex = widgets.findIndex(w => w.id === targetId);
		if (fromIndex !== -1 && toIndex !== -1) {
			moveWidget(fromIndex, toIndex);
		}
		draggedWidget = null;
	}
	
	function handleDragEnd() {
		draggedWidget = null;
	}

	// Quick action handlers
	function handleQuickAction(action: string) {
		switch (action) {
			case 'new-task':
				goto('/tasks?new=true');
				break;
			case 'new-project':
				goto('/projects?new=true');
				break;
			case 'new-chat':
				goto('/chat?new=true');
				break;
			case 'daily-log':
				goto('/daily');
				break;
		}
	}

	// Focus item handlers
	async function handleFocusToggle(id: string) {
		const item = focusItems.find((i) => i.id === id);
		if (!item) return;

		try {
			await api.updateFocusItem(id, { completed: !item.completed });
			focusItems = focusItems.map((i) =>
				i.id === id ? { ...i, completed: !i.completed } : i
			);
		} catch (err) {
			console.error('Failed to toggle focus item:', err);
		}
	}

	async function handleFocusAdd(text: string) {
		try {
			const newItem = await api.createFocusItem(text);
			focusItems = [
				...focusItems,
				{ id: newItem.id, text: newItem.text, completed: newItem.completed }
			];
		} catch (err) {
			console.error('Failed to add focus item:', err);
		}
	}

	async function handleFocusRemove(id: string) {
		try {
			await api.deleteFocusItem(id);
			focusItems = focusItems.filter((item) => item.id !== id);
		} catch (err) {
			console.error('Failed to remove focus item:', err);
		}
	}

	function handleFocusEdit() {
		// TODO: Implement focus edit mode
		console.log('Edit focus items');
	}

	// Task handlers
	async function handleTaskToggle(id: string) {
		try {
			await api.toggleTask(id);
			tasks = tasks.map((task) =>
				task.id === id ? { ...task, completed: !task.completed } : task
			);
		} catch (err) {
			console.error('Failed to toggle task:', err);
		}
	}

	// Energy check handler
	function handleEnergySet(level: number) {
		energyLevel = level;
		// TODO: Save to backend
	}
	
	// Get grid class based on widget size
	function getWidgetGridClass(size: WidgetSize): string {
		switch (size) {
			case 'small': return 'col-span-1';
			case 'medium': return 'col-span-1 md:col-span-1 lg:col-span-2';
			case 'large': return 'col-span-1 md:col-span-2 lg:col-span-3';
		}
	}
	
	// Get accent color border class
	function getAccentBorderClass(color?: string): string {
		switch (color) {
			case 'blue': return 'border-l-4 border-l-blue-500';
			case 'green': return 'border-l-4 border-l-green-500';
			case 'purple': return 'border-l-4 border-l-purple-500';
			case 'orange': return 'border-l-4 border-l-orange-500';
			case 'pink': return 'border-l-4 border-l-pink-500';
			default: return '';
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="h-full flex flex-col bg-gradient-to-br from-gray-50 via-gray-50 to-gray-100/50 relative">
	<!-- Subtle decorative background elements -->
	<div class="absolute top-0 left-1/4 w-96 h-96 bg-blue-100/20 rounded-full blur-3xl pointer-events-none"></div>
	<div class="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-100/20 rounded-full blur-3xl pointer-events-none"></div>
	
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="flex flex-col items-center gap-3">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
				<p class="text-sm text-gray-500">Loading dashboard...</p>
			</div>
		</div>
	{:else if error}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="text-center bg-white p-8 rounded-2xl shadow-sm border border-gray-200 max-w-md">
				<div class="w-16 h-16 bg-red-50 rounded-xl flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<p class="text-gray-600 mb-4">{error}</p>
				<button onclick={() => loadDashboard()} class="btn-pill btn-pill-primary">
					Try Again
				</button>
			</div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto relative" in:fade={{ duration: 300 }}>
			<!-- Top Toolbar -->
			<div class="px-6 pt-4 pb-2">
				<div class="flex items-center justify-end gap-3">
					<!-- Segmented Control: View / Edit -->
					<div class="flex items-center p-1 bg-gray-100 rounded-lg" role="tablist">
						<button
							onclick={() => isEditMode && toggleEditMode()}
							role="tab"
							aria-selected={!isEditMode}
							class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200
								{!isEditMode 
									? 'bg-white text-gray-900 shadow-sm' 
									: 'text-gray-500 hover:text-gray-700'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
							</svg>
							View
						</button>
						<button
							onclick={() => !isEditMode && toggleEditMode()}
							role="tab"
							aria-selected={isEditMode}
							class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200
								{isEditMode 
									? 'bg-white text-blue-600 shadow-sm' 
									: 'text-gray-500 hover:text-gray-700'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Edit
						</button>
					</div>
					
					<!-- Separator -->
					<div class="h-6 w-px bg-gray-200"></div>
					
					<!-- Notification Bell -->
					<NotificationDropdown />
				</div>
			</div>
			
			<!-- Header with Greeting -->
			<div class="px-6">
				<DashboardHeader
					userName={$session.data?.user?.name || 'there'}
					{energyLevel}
					onEnergySet={handleEnergySet}
				/>
				
				<!-- Edit Mode Banner -->
				{#if isEditMode}
					<div 
						class="mb-4 px-4 py-3 bg-blue-50 border border-blue-200 rounded-xl flex items-center justify-between"
						transition:fly={{ y: -10, duration: 200 }}
					>
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
								<svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
								</svg>
							</div>
							<div>
								<p class="text-sm font-medium text-blue-900">Edit Mode</p>
								<p class="text-xs text-blue-600">Drag widgets to reorder • Press <kbd class="px-1.5 py-0.5 bg-blue-100 rounded text-xs">Esc</kbd> to exit</p>
							</div>
						</div>
						<button
							onclick={() => showWidgetPicker = true}
							class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							Add Widget
						</button>
					</div>
				{/if}
			</div>

			<!-- Widget Grid -->
			<div class="px-6 py-4">
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" role={isEditMode ? 'list' : undefined}>
					{#each widgets as widget, index (widget.id)}
						{@const isSelected = isEditMode && selectedWidgetIndex === index}
						<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
						<div
							class="{getWidgetGridClass(widget.size)} transition-all duration-200
								{isEditMode ? 'cursor-move' : ''}
								{draggedWidget === widget.id ? 'opacity-50 scale-95' : ''}
								{isSelected ? 'scale-[1.02]' : ''}"
							role={isEditMode ? 'listitem' : undefined}
							tabindex={isEditMode ? 0 : -1}
							draggable={isEditMode}
							ondragstart={(e) => handleDragStart(e, widget.id)}
							ondragover={handleDragOver}
							ondrop={(e) => handleDrop(e, widget.id)}
							ondragend={handleDragEnd}
							onclick={() => isEditMode && (selectedWidgetIndex = index)}
							animate:flip={{ duration: 300 }}
						>
							<!-- Widget Container - pt-2 adds space for edit controls above -->
							<div class="relative {isEditMode ? 'pt-2' : ''}">
								<!-- Edit Mode Overlay Controls - positioned inside padding area -->
								{#if isEditMode}
									<div class="absolute top-0 right-1 z-10 flex gap-1">
										<!-- Collapse Toggle -->
										<button
											onclick={(e) => { e.stopPropagation(); toggleWidgetCollapse(widget.id); }}
											class="w-6 h-6 bg-white border border-gray-200 rounded-full flex items-center justify-center shadow-sm hover:bg-gray-50 transition-colors"
											title={widget.collapsed ? 'Expand' : 'Collapse'}
										>
											<svg class="w-3 h-3 text-gray-500 transition-transform {widget.collapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
											</svg>
										</button>
										
										<!-- Widget Menu -->
										<DropdownMenu.Root>
											<DropdownMenu.Trigger
												class="w-6 h-6 bg-white border border-gray-200 rounded-full flex items-center justify-center shadow-sm hover:bg-gray-50 transition-colors"
											>
												<svg class="w-3 h-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
												</svg>
											</DropdownMenu.Trigger>
											<DropdownMenu.Content 
												class="z-50 min-w-[160px] bg-white rounded-lg border border-gray-200 shadow-lg py-1"
												sideOffset={4}
											>
												<!-- Size Options -->
												<DropdownMenu.Sub>
													<DropdownMenu.SubTrigger class="flex items-center justify-between gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 cursor-pointer w-full">
														<div class="flex items-center gap-2">
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
															</svg>
															Size
														</div>
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
														</svg>
													</DropdownMenu.SubTrigger>
													<DropdownMenu.SubContent class="z-50 min-w-[120px] bg-white rounded-lg border border-gray-200 shadow-lg py-1">
														<DropdownMenu.Item 
															class="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 cursor-pointer {widget.size === 'small' ? 'text-blue-600 bg-blue-50' : 'text-gray-700'}"
															onclick={() => setWidgetSize(widget.id, 'small')}
														>
															Small
														</DropdownMenu.Item>
														<DropdownMenu.Item 
															class="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 cursor-pointer {widget.size === 'medium' ? 'text-blue-600 bg-blue-50' : 'text-gray-700'}"
															onclick={() => setWidgetSize(widget.id, 'medium')}
														>
															Medium
														</DropdownMenu.Item>
														<DropdownMenu.Item 
															class="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 cursor-pointer {widget.size === 'large' ? 'text-blue-600 bg-blue-50' : 'text-gray-700'}"
															onclick={() => setWidgetSize(widget.id, 'large')}
														>
															Large
														</DropdownMenu.Item>
													</DropdownMenu.SubContent>
												</DropdownMenu.Sub>
												
												<!-- Color Options -->
												<DropdownMenu.Sub>
													<DropdownMenu.SubTrigger class="flex items-center justify-between gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 cursor-pointer w-full">
														<div class="flex items-center gap-2">
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
															</svg>
															Color
														</div>
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
														</svg>
													</DropdownMenu.SubTrigger>
													<DropdownMenu.SubContent class="z-50 min-w-[120px] bg-white rounded-lg border border-gray-200 shadow-lg py-1">
														{#each accentColors as color}
															<DropdownMenu.Item 
																class="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 cursor-pointer {widget.accentColor === color.value ? 'text-blue-600 bg-blue-50' : 'text-gray-700'}"
																onclick={() => setWidgetAccentColor(widget.id, color.value)}
															>
																{#if color.value}
																	<span class="w-3 h-3 rounded-full bg-{color.value}-500"></span>
																{:else}
																	<span class="w-3 h-3 rounded-full bg-gray-300"></span>
																{/if}
																{color.name}
															</DropdownMenu.Item>
														{/each}
													</DropdownMenu.SubContent>
												</DropdownMenu.Sub>
												
												<DropdownMenu.Separator class="my-1 h-px bg-gray-100" />
												<DropdownMenu.Item 
													class="flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50 cursor-pointer"
													onclick={() => removeWidget(widget.id)}
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
													</svg>
													Remove
												</DropdownMenu.Item>
											</DropdownMenu.Content>
										</DropdownMenu.Root>
									</div>
								{/if}
								
								<!-- Widget Card -->
								<div class="bg-white rounded-xl border transition-all duration-200 overflow-hidden relative group
									{getAccentBorderClass(widget.accentColor)}
									{isSelected
										? 'border-blue-500 border-solid shadow-sm'
										: isEditMode 
											? 'border-blue-200 border-dashed shadow-sm hover:shadow-md hover:border-blue-300' 
											: 'border-gray-200 shadow-sm hover:shadow-md'}">
									
<!-- Analytics Toggle Icon (appears on hover, not in edit mode, hidden when analytics is showing) -->
										{#if !isEditMode && !widget.collapsed && !widget.showAnalytics}
											<button
												onclick={(e) => { e.stopPropagation(); toggleWidgetAnalytics(widget.id); }}
												class="absolute top-3 right-3 z-20 w-7 h-7 flex items-center justify-center rounded-lg transition-all duration-200
													bg-white/80 text-gray-400 opacity-0 group-hover:opacity-100 hover:bg-gray-100 hover:text-gray-600 border border-transparent hover:border-gray-200"
												title="View Analytics"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
											</svg>
										</button>
									{/if}
									
									<!-- Drag Handle Indicator (inside card) -->
									{#if isEditMode}
										<div class="absolute top-4 left-2 text-gray-400 z-10">
											<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
												<path d="M8 6a2 2 0 11-4 0 2 2 0 014 0zM8 12a2 2 0 11-4 0 2 2 0 014 0zM8 18a2 2 0 11-4 0 2 2 0 014 0zM14 6a2 2 0 11-4 0 2 2 0 014 0zM14 12a2 2 0 11-4 0 2 2 0 014 0zM14 18a2 2 0 11-4 0 2 2 0 014 0z" />
											</svg>
										</div>
									{/if}
								
									<!-- Collapsed Title Bar -->
									{#if widget.collapsed}
										<div class="px-4 py-3 flex items-center justify-between bg-gray-50">
											<span class="text-sm font-medium text-gray-700">{widget.title}</span>
											<button 
												onclick={() => toggleWidgetCollapse(widget.id)}
												class="text-gray-400 hover:text-gray-600"
												aria-label="Expand widget"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
												</svg>
											</button>
										</div>
									{:else if widget.showAnalytics}
										<!-- Analytics Flip View -->
										<div class="p-5" transition:fade={{ duration: 150 }}>
											<div class="flex items-center justify-between mb-4 pb-3 border-b border-gray-100">
												<div class="flex items-center gap-2">
													<div class="w-8 h-8 bg-gradient-to-br from-gray-700 to-gray-800 rounded-lg flex items-center justify-center shadow-sm">
														<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
														</svg>
													</div>
													<span class="text-sm font-semibold text-gray-900">{widgetAnalytics[widget.type].title}</span>
												</div>
												<button
													onclick={(e) => { e.stopPropagation(); toggleWidgetAnalytics(widget.id); }}
													class="flex items-center gap-1 px-2.5 py-1.5 text-xs text-gray-500 hover:text-gray-700 bg-gray-50 hover:bg-gray-100 font-medium rounded-lg transition-colors border border-gray-200"
												>
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
													</svg>
													Back
												</button>
											</div>
											
											<div class="space-y-1">
												{#each widgetAnalytics[widget.type].stats as stat}
													<div class="flex items-center justify-between py-2.5 px-3 rounded-lg hover:bg-gray-50 transition-colors">
														<span class="text-sm text-gray-600">{stat.label}</span>
														<div class="flex items-center gap-2">
															<span class="text-sm font-semibold text-gray-900">{stat.value}</span>
															{#if stat.trend}
																<span class="text-xs font-medium px-1.5 py-0.5 rounded {stat.trend.startsWith('+') ? 'text-green-700 bg-green-50' : 'text-red-700 bg-red-50'}">{stat.trend}</span>
															{/if}
														</div>
													</div>
												{/each}
											</div>
											
											<button
												onclick={() => showAnalyticsSidepanel = true}
												class="btn-pill btn-pill-primary btn-pill-block btn-pill-sm mt-4 flex items-center justify-center gap-2"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
												</svg>
												View Full Analytics
											</button>
										</div>
									{:else}
										<!-- Widget Content -->
										<div class="{isEditMode ? 'pointer-events-none' : ''}">
											{#if widget.type === 'focus'}
												<TodaysFocusWidget
													items={focusItems}
													onToggle={handleFocusToggle}
													onAdd={handleFocusAdd}
													onRemove={handleFocusRemove}
													onEdit={handleFocusEdit}
												/>
											{:else if widget.type === 'quick-actions'}
												<QuickActionsWidget onAction={handleQuickAction} />
											{:else if widget.type === 'projects'}
												<ActiveProjectsWidget
													{projects}
													onViewAll={() => goto('/projects')}
												/>
											{:else if widget.type === 'tasks'}
												<MyTasksWidget
													{tasks}
													onToggle={handleTaskToggle}
													onViewAll={() => goto('/tasks')}
												/>
											{:else if widget.type === 'activity'}
												<RecentActivityWidget {activities} onViewAll={() => goto('/chat')} />
											{:else if widget.type === 'metric'}
												<!-- Placeholder Metric Card -->
												<div class="p-5">
													<div class="flex items-center justify-between mb-3">
														<span class="text-sm text-gray-500">Tasks Due Today</span>
														<span class="text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded-full">+12%</span>
													</div>
													<div class="text-3xl font-bold text-gray-900">8</div>
													<div class="text-xs text-gray-400 mt-1">vs 7 yesterday</div>
												</div>
											{/if}
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/each}
					
					<!-- Empty State Add Widget Card (shown in edit mode when few widgets) -->
					{#if isEditMode && widgets.length < 6}
						<button
							onclick={() => showWidgetPicker = true}
							class="col-span-1 min-h-[200px] border-2 border-dashed border-gray-300 rounded-xl flex flex-col items-center justify-center gap-2 text-gray-400 hover:border-blue-400 hover:text-blue-500 hover:bg-blue-50/50 transition-all duration-200"
						>
							<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4v16m8-8H4" />
							</svg>
							<span class="text-sm font-medium">Add Widget</span>
						</button>
					{/if}
				</div>
			</div>
		</div>
	{/if}
	
	<!-- Widget Picker Drawer -->
	{#if showWidgetPicker}
		<!-- Backdrop -->
		<button
			class="fixed inset-0 bg-black/20 z-40"
			onclick={() => showWidgetPicker = false}
			transition:fade={{ duration: 150 }}
			aria-label="Close widget picker"
		></button>
		
		<!-- Drawer -->
		<div 
			class="fixed bottom-0 left-0 right-0 bg-white rounded-t-2xl shadow-2xl z-50 max-h-[70vh] overflow-hidden"
			transition:fly={{ y: 300, duration: 300 }}
		>
			<!-- Handle -->
			<div class="flex justify-center pt-3 pb-2">
				<div class="w-10 h-1 bg-gray-300 rounded-full"></div>
			</div>
			
			<!-- Header -->
			<div class="px-6 pb-4 border-b border-gray-100">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-lg font-semibold text-gray-900">Add Widget</h2>
						<p class="text-sm text-gray-500">Choose a widget to add to your dashboard</p>
					</div>
					<button
						onclick={() => showWidgetPicker = false}
						class="w-8 h-8 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						aria-label="Close widget picker"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				
				<!-- Size Picker -->
				<div class="mt-4 flex items-center gap-2">
					<span class="text-sm text-gray-500">Size:</span>
					<div class="flex items-center p-0.5 bg-gray-100 rounded-lg">
						<button
							onclick={() => pickerSelectedSize = 'small'}
							class="px-3 py-1 text-xs font-medium rounded-md transition-all
								{pickerSelectedSize === 'small' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
						>
							Small
						</button>
						<button
							onclick={() => pickerSelectedSize = 'medium'}
							class="px-3 py-1 text-xs font-medium rounded-md transition-all
								{pickerSelectedSize === 'medium' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
						>
							Medium
						</button>
						<button
							onclick={() => pickerSelectedSize = 'large'}
							class="px-3 py-1 text-xs font-medium rounded-md transition-all
								{pickerSelectedSize === 'large' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
						>
							Large
						</button>
					</div>
				</div>
			</div>
			
			<!-- Widget List -->
			<div class="p-6 overflow-y-auto max-h-[calc(70vh-160px)]">
				<div class="grid grid-cols-2 md:grid-cols-3 gap-3">
					{#each availableWidgets as widgetOption}
						{@const isUnique = uniqueWidgetTypes.includes(widgetOption.type)}
						{@const isAlreadyAdded = isUnique && addedUniqueTypes.has(widgetOption.type)}
						{@const existingCount = widgets.filter(w => w.type === widgetOption.type).length}
						<button
							onclick={() => addWidget(widgetOption.type)}
							disabled={isAlreadyAdded}
							class="flex flex-col items-start p-4 border rounded-xl transition-all duration-200 text-left group relative
								{isAlreadyAdded 
									? 'bg-gray-100 border-gray-200 cursor-not-allowed opacity-60' 
									: 'bg-gray-50 hover:bg-blue-50 border-gray-200 hover:border-blue-300'}"
						>
							{#if isAlreadyAdded}
								<span class="absolute top-2 right-2 flex items-center gap-1 text-xs font-medium text-green-600 bg-green-100 px-1.5 py-0.5 rounded">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Already added
								</span>
							{:else if !isUnique && existingCount > 0}
								<span class="absolute top-2 right-2 text-xs font-medium text-blue-600 bg-blue-100 px-1.5 py-0.5 rounded">
									{existingCount} added
								</span>
							{/if}
							<span class="text-2xl mb-2">{widgetOption.icon}</span>
							<span class="font-medium {isAlreadyAdded ? 'text-gray-500' : 'text-gray-900 group-hover:text-blue-900'}">{widgetOption.title}</span>
							<span class="text-xs {isAlreadyAdded ? 'text-gray-400' : 'text-gray-500 group-hover:text-blue-600'} mt-0.5">{widgetOption.description}</span>
						</button>
					{/each}
				</div>
			</div>
		</div>
	{/if}
	
	<!-- Undo Toast -->
	{#if showUndoToast && undoStack.length > 0}
		<div 
			class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 bg-gray-900 text-white rounded-xl shadow-xl"
			transition:fly={{ y: 50, duration: 200 }}
		>
			<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
			</svg>
			<span class="text-sm">Widget removed</span>
			<button
				onclick={undoRemove}
				class="btn-pill btn-pill-sm bg-white/20 hover:bg-white/30 text-white border-white/30"
			>
				Undo
			</button>
			<button
				onclick={() => { showUndoToast = false; undoStack = []; }}
				class="ml-1 p-1 hover:bg-white/10 rounded transition-colors"
				aria-label="Dismiss"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	{/if}
	
	<!-- Floating Analytics Button (FAB) -->
	{#if !isEditMode && !showWidgetPicker}
		<button
			onclick={() => showAnalyticsSidepanel = true}
			class="fixed bottom-6 right-6 z-30 w-12 h-12 bg-gray-900 hover:bg-gray-800 text-white rounded-xl shadow-md hover:shadow-lg flex items-center justify-center transition-all duration-200"
			transition:scale={{ duration: 200 }}
			title="View Dashboard Analytics"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
			</svg>
		</button>
	{/if}
	
	<!-- Analytics Sidepanel -->
	<AnalyticsSidepanel 
		isOpen={showAnalyticsSidepanel} 
		onClose={() => showAnalyticsSidepanel = false}
		analytics={seededAnalytics}
		isLoading={analyticsLoading}
		onTimeRangeChange={handleAnalyticsTimeRangeChange}
	/>
</div>
