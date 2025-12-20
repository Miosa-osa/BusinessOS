<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { fade } from 'svelte/transition';
	import {
		DashboardHeader,
		TodaysFocusWidget,
		QuickActionsWidget,
		ActiveProjectsWidget,
		MyTasksWidget,
		RecentActivityWidget
	} from '$lib/components/dashboard';
	import {
		api,
		type DashboardProject,
		type DashboardTask,
		type DashboardActivity
	} from '$lib/api';

	const session = useSession();

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

	async function handleFocusEdit(id: string, newText: string) {
		try {
			await api.updateFocusItem(id, { text: newText });
			focusItems = focusItems.map((item) => (item.id === id ? { ...item, text: newText } : item));
		} catch (err) {
			console.error('Failed to edit focus item:', err);
		}
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
</script>

<div class="h-full flex flex-col bg-gray-50">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div
				class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"
			></div>
		</div>
	{:else if error}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="text-center">
				<div class="text-red-500 mb-4">
					<svg class="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<p class="text-gray-600 mb-4">{error}</p>
				<button onclick={() => loadDashboard()} class="btn btn-secondary">
					Try Again
				</button>
			</div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto" in:fade={{ duration: 300 }}>
			<!-- Header -->
			<div class="px-6 pt-6">
				<DashboardHeader
					userName={$session.data?.user?.name || 'there'}
					{energyLevel}
					onEnergySet={handleEnergySet}
				/>
			</div>

			<!-- Main Grid -->
			<div class="px-6 py-6">
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Left Column: Focus + Quick Actions -->
					<div class="lg:col-span-1 space-y-6">
						<TodaysFocusWidget
							items={focusItems}
							onToggle={handleFocusToggle}
							onAdd={handleFocusAdd}
							onRemove={handleFocusRemove}
							onEdit={handleFocusEdit}
						/>
						<QuickActionsWidget onAction={handleQuickAction} />
					</div>

					<!-- Middle Column: Projects + Tasks -->
					<div class="lg:col-span-1 space-y-6">
						<ActiveProjectsWidget
							{projects}
							onViewAll={() => goto('/projects')}
						/>
						<MyTasksWidget
							{tasks}
							onToggle={handleTaskToggle}
							onViewAll={() => goto('/tasks')}
						/>
					</div>

					<!-- Right Column: Activity -->
					<div class="lg:col-span-1">
						<RecentActivityWidget {activities} onViewAll={() => goto('/chat')} />
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
