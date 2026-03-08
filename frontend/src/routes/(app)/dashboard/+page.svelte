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
		AnalyticsSidepanel,
		SignalHealthWidget
	} from '$lib/components/dashboard';
	import { NotificationDropdown } from '$lib/components/notifications';

	// ── Stores ────────────────────────────────────────────────────────────────────
	import { dashboardLayoutStore } from '$lib/stores/dashboard/dashboardLayoutStore.svelte';
	import {
		accentColors,
		availableWidgets,
		uniqueWidgetTypes,
		getAccentColorClass,
		getWidgetGridClass,
		getAccentBorderClass
	} from '$lib/stores/dashboard/dashboardLayoutStore.svelte';
	import { dashboardAnalyticsStore, widgetAnalytics } from '$lib/stores/dashboard/dashboardAnalyticsStore.svelte';
	import { dashboardDataStore } from '$lib/stores/dashboard/dashboardDataStore.svelte';

	const session = useSession();

	// Short aliases for ergonomics in the template
	const layout = dashboardLayoutStore;
	const analytics = dashboardAnalyticsStore;
	const data = dashboardDataStore;

	// ── Bootstrap ────────────────────────────────────────────────────────────────

	$effect(() => {
		if ($session.data) {
			data.loadDashboard();
		}
	});

	// ── Quick-action handler (navigation only, stays in page) ─────────────────────

	function handleQuickAction(action: string): void {
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
</script>

<svelte:window onkeydown={(e) => layout.handleKeydown(e)} />

<div class="dw-page h-full flex flex-col relative">

	{#if data.isLoading}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="flex flex-col items-center gap-3">
				<div class="dw-page-spinner animate-spin h-8 w-8 border-2 border-t-transparent rounded-full"></div>
				<p class="dw-page-muted text-sm">Loading dashboard...</p>
			</div>
		</div>
	{:else if data.error}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="dw-page-card text-center p-8 rounded-2xl shadow-sm border max-w-md">
				<div class="dw-page-error-icon">
					<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<p class="dw-page-muted mb-4">{data.error}</p>
				<button onclick={() => data.loadDashboard()} class="dw-page-btn-primary inline-flex items-center justify-center gap-2 px-3.5 py-2 text-sm font-medium rounded-lg transition-colors">
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
					<div class="dw-page-seg flex items-center p-1 rounded-lg" role="tablist">
						<button
							onclick={() => layout.isEditMode && layout.toggleEditMode()}
							role="tab"
							aria-selected={!layout.isEditMode}
							class="inline-flex items-center justify-center gap-1 px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {!layout.isEditMode ? 'dw-page-seg-active shadow-sm' : 'dw-page-seg-inactive'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
							</svg>
							View
						</button>
						<button
							onclick={() => !layout.isEditMode && layout.toggleEditMode()}
							role="tab"
							aria-selected={layout.isEditMode}
							class="inline-flex items-center justify-center gap-1 px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {layout.isEditMode ? 'dw-page-seg-active shadow-sm' : 'dw-page-seg-inactive'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Edit
						</button>
					</div>

					<!-- Separator -->
					<div class="dw-page-sep h-6 w-px"></div>

					<!-- Analytics -->
					{#if !layout.isEditMode && !layout.showWidgetPicker}
						<button
							onclick={() => (analytics.showAnalyticsSidepanel = true)}
							class="dw-page-icon-btn inline-flex items-center justify-center p-2 rounded-lg transition-colors"
							title="View Dashboard Analytics"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</button>
					{/if}

					<!-- Notification Bell -->
					<NotificationDropdown />
				</div>
			</div>

			<!-- Header with Greeting -->
			<div class="px-6">
				<DashboardHeader
					userName={$session.data?.user?.name || 'there'}
					energyLevel={data.energyLevel}
					onEnergySet={(level) => data.handleEnergySet(level)}
				/>

				<!-- Edit Mode Banner -->
				{#if layout.isEditMode}
					<div
						class="dw-page-edit-banner mb-4 px-4 py-3 rounded-xl flex items-center justify-between border"
						transition:fly={{ y: -10, duration: 200 }}
					>
						<div class="flex items-center gap-3">
							<div class="dw-page-edit-icon w-8 h-8 rounded-lg flex items-center justify-center">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
								</svg>
							</div>
							<div>
								<p class="dw-page-edit-title text-sm font-medium">Edit Mode</p>
								<p class="dw-page-edit-hint text-xs">Drag widgets to reorder • Press <kbd class="dw-page-edit-kbd px-1.5 py-0.5 rounded text-xs">Esc</kbd> to exit</p>
							</div>
						</div>
						<button
							onclick={() => (layout.showWidgetPicker = true)}
							class="dw-page-btn-primary inline-flex items-center justify-center gap-2 px-3.5 py-2 text-sm font-medium rounded-lg transition-colors"
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
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" role={layout.isEditMode ? 'list' : undefined}>
					{#each layout.widgets as widget, index (widget.id)}
						{@const isSelected = layout.isEditMode && layout.selectedWidgetIndex === index}
						<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
						<div
							class="{getWidgetGridClass(widget.size)} transition-all duration-200
								{layout.isEditMode ? 'cursor-move' : ''}
								{layout.draggedWidget === widget.id ? 'opacity-50 scale-95' : ''}
								{isSelected ? 'scale-[1.02]' : ''}"
							role={layout.isEditMode ? 'listitem' : undefined}
							tabindex={layout.isEditMode ? 0 : -1}
							draggable={layout.isEditMode}
							ondragstart={(e) => layout.handleDragStart(e, widget.id)}
							ondragover={(e) => layout.handleDragOver(e)}
							ondrop={(e) => layout.handleDrop(e, widget.id)}
							ondragend={() => layout.handleDragEnd()}
							onclick={() => layout.isEditMode && (layout.selectedWidgetIndex = index)}
							animate:flip={{ duration: 300 }}
							in:fade={{ duration: 200, delay: index * 50 }}
						>
							<!-- Widget Container - pt-2 adds space for edit controls above -->
							<div class="relative {layout.isEditMode ? 'pt-2' : ''}">
								<!-- Edit Mode Overlay Controls -->
								{#if layout.isEditMode}
									<div class="absolute top-0 right-1 z-10 flex gap-1">
										<!-- Collapse Toggle -->
										<button
											onclick={(e) => { e.stopPropagation(); layout.toggleWidgetCollapse(widget.id); }}
											class="dw-page-ctrl inline-flex items-center justify-center p-1.5 rounded-lg border shadow-sm transition-colors"
											title={widget.collapsed ? 'Expand' : 'Collapse'}
										>
											<svg class="w-3 h-3 dw-page-ctrl-icon transition-transform {widget.collapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
											</svg>
										</button>

										<!-- Widget Menu -->
										<DropdownMenu.Root>
											<DropdownMenu.Trigger
												class="dw-page-ctrl inline-flex items-center justify-center p-1.5 rounded-lg border shadow-sm transition-colors"
											>
												<svg class="w-3 h-3 dw-page-ctrl-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
												</svg>
											</DropdownMenu.Trigger>
											<DropdownMenu.Content
												class="dw-page-dropdown z-50 min-w-[160px] rounded-lg border shadow-lg py-1"
												sideOffset={4}
											>
												<!-- Size Options -->
												<DropdownMenu.Sub>
												<DropdownMenu.SubTrigger class="dw-page-dropdown-item flex items-center justify-between gap-2 px-3 py-2 text-sm cursor-pointer w-full">
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
													<DropdownMenu.SubContent class="dw-page-dropdown z-50 min-w-[120px] rounded-lg border shadow-lg py-1">
														<DropdownMenu.Item
															class="dw-page-dropdown-item flex items-center gap-2 px-3 py-2 text-sm cursor-pointer {widget.size === 'small' ? 'dw-page-dropdown-item--active' : ''}"
															onclick={() => layout.setWidgetSize(widget.id, 'small')}
														>
															Small
														</DropdownMenu.Item>
														<DropdownMenu.Item
															class="dw-page-dropdown-item flex items-center gap-2 px-3 py-2 text-sm cursor-pointer {widget.size === 'medium' ? 'dw-page-dropdown-item--active' : ''}"
															onclick={() => layout.setWidgetSize(widget.id, 'medium')}
														>
															Medium
														</DropdownMenu.Item>
														<DropdownMenu.Item
															class="dw-page-dropdown-item flex items-center gap-2 px-3 py-2 text-sm cursor-pointer {widget.size === 'large' ? 'dw-page-dropdown-item--active' : ''}"
															onclick={() => layout.setWidgetSize(widget.id, 'large')}
														>
															Large
														</DropdownMenu.Item>
													</DropdownMenu.SubContent>
												</DropdownMenu.Sub>

												<!-- Color Options -->
												<DropdownMenu.Sub>
												<DropdownMenu.SubTrigger class="dw-page-dropdown-item flex items-center justify-between gap-2 px-3 py-2 text-sm cursor-pointer w-full">
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
													<DropdownMenu.SubContent class="dw-page-dropdown z-50 min-w-[120px] rounded-lg border shadow-lg py-1">
														{#each accentColors as color}
															<DropdownMenu.Item
																class="dw-page-dropdown-item flex items-center gap-2 px-3 py-2 text-sm cursor-pointer {widget.accentColor === color.value ? 'dw-page-dropdown-item--active' : ''}"
																onclick={() => layout.setWidgetAccentColor(widget.id, color.value)}
															>
																<span class="w-3 h-3 rounded-full {getAccentColorClass(color.value)}"></span>
																{color.name}
															</DropdownMenu.Item>
														{/each}
													</DropdownMenu.SubContent>
												</DropdownMenu.Sub>

												<DropdownMenu.Separator class="dw-page-dropdown-sep my-1 h-px" />
												<DropdownMenu.Item
													class="flex items-center gap-2 px-3 py-2 text-sm text-red-500 hover:bg-red-500/10 cursor-pointer"
													onclick={() => layout.removeWidget(widget.id)}
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
								<div class="dw-page-widget-card rounded-xl border transition-all duration-200 overflow-hidden relative group
									{getAccentBorderClass(widget.accentColor)}
									{isSelected
										? 'border-blue-500 border-solid shadow-sm'
										: layout.isEditMode
											? 'border-blue-300/40 border-dashed shadow-sm hover:shadow-md hover:border-blue-400/60'
											: 'shadow-sm hover:shadow-md'}">

									<!-- Analytics Toggle Icon (appears on hover, not in edit mode) -->
									{#if !layout.isEditMode && !widget.collapsed && !widget.showAnalytics}
										<button
											onclick={(e) => { e.stopPropagation(); layout.toggleWidgetAnalytics(widget.id); }}
											class="dw-page-ctrl inline-flex items-center justify-center p-1.5 rounded-lg border shadow-sm transition-colors absolute top-3 right-3 z-20 opacity-0 group-hover:opacity-100"
											title="View Analytics"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
											</svg>
										</button>
									{/if}

									<!-- Drag Handle Indicator (inside card) -->
									{#if layout.isEditMode}
										<div class="dw-page-drag-handle absolute top-4 left-2 z-10">
											<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
												<path d="M8 6a2 2 0 11-4 0 2 2 0 014 0zM8 12a2 2 0 11-4 0 2 2 0 014 0zM8 18a2 2 0 11-4 0 2 2 0 014 0zM14 6a2 2 0 11-4 0 2 2 0 014 0zM14 12a2 2 0 11-4 0 2 2 0 014 0zM14 18a2 2 0 11-4 0 2 2 0 014 0z" />
											</svg>
										</div>
									{/if}

									<!-- Collapsed Title Bar -->
									{#if widget.collapsed}
										<div class="dw-page-collapsed px-4 py-3 flex items-center justify-between">
											<span class="dw-page-text text-sm font-medium">{widget.title}</span>
											<button
												onclick={() => layout.toggleWidgetCollapse(widget.id)}
												class="dw-page-muted hover:dw-page-text"
												aria-label="Expand widget"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
												</svg>
											</button>
										</div>
									{:else if widget.showAnalytics}
										<!-- Analytics Flip View -->
										<div class="p-5" transition:fade={{ duration: 200 }}>
											<div class="dw-page-divider flex items-center justify-between mb-4 pb-3 border-b">
												<div class="flex items-center gap-2">
													<div class="dw-page-analytics-icon w-8 h-8 rounded-lg flex items-center justify-center shadow-sm">
														<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
														</svg>
													</div>
													<span class="dw-page-text text-sm font-semibold">{widgetAnalytics[widget.type].title}</span>
												</div>
												<button
													onclick={(e) => { e.stopPropagation(); layout.toggleWidgetAnalytics(widget.id); }}
													class="dw-page-back-btn inline-flex items-center justify-center gap-1 px-2.5 py-1.5 text-xs font-medium rounded-lg border transition-colors"
												>
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
													</svg>
													Back
												</button>
											</div>

											<div class="space-y-1">
												{#each widgetAnalytics[widget.type].stats as stat}
													<div class="dw-page-stat-row flex items-center justify-between py-2.5 px-3 rounded-lg transition-colors">
														<span class="dw-page-muted text-sm">{stat.label}</span>
														<div class="flex items-center gap-2">
															<span class="dw-page-text text-sm font-semibold">{stat.value}</span>
															{#if stat.trend}
																<span class="text-xs font-medium px-1.5 py-0.5 rounded {stat.trend.startsWith('+') ? 'text-green-700 bg-green-50' : 'text-red-700 bg-red-50'}">{stat.trend}</span>
															{/if}
														</div>
													</div>
												{/each}
											</div>

											<button
												onclick={() => (analytics.showAnalyticsSidepanel = true)}
												class="dw-page-btn-primary w-full flex items-center justify-center gap-2 px-3.5 py-2 text-sm font-medium rounded-lg transition-colors mt-4"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
												</svg>
												View Full Analytics
											</button>
										</div>
									{:else}
										<!-- Widget Content -->
										<div class="{layout.isEditMode ? 'pointer-events-none' : ''}">
											{#if widget.type === 'focus'}
												<TodaysFocusWidget
													items={data.focusItems}
													onToggle={(id) => data.handleFocusToggle(id)}
													onAdd={(text) => data.handleFocusAdd(text)}
													onRemove={(id) => data.handleFocusRemove(id)}
													onEdit={() => data.handleFocusEdit()}
												/>
											{:else if widget.type === 'quick-actions'}
												<QuickActionsWidget onAction={(action) => handleQuickAction(action)} />
											{:else if widget.type === 'projects'}
												<ActiveProjectsWidget
													projects={data.projects}
													onViewAll={() => goto('/projects')}
												/>
											{:else if widget.type === 'tasks'}
												<MyTasksWidget
													tasks={data.tasks}
													onToggle={(id) => data.handleTaskToggle(id)}
													onViewAll={() => goto('/tasks')}
												/>
											{:else if widget.type === 'activity'}
												<RecentActivityWidget activities={data.activities} onViewAll={() => goto('/chat')} />
											{:else if widget.type === 'metric'}
												<!-- Placeholder Metric Card -->
												<div class="p-5">
													<div class="flex items-center justify-between mb-3">
														<span class="dw-page-muted text-sm">Tasks Due Today</span>
														<span class="text-xs text-green-600 bg-green-500/10 px-2 py-0.5 rounded-full">+12%</span>
													</div>
													<div class="dw-page-text text-3xl font-bold">8</div>
													<div class="dw-page-meta text-xs mt-1">vs 7 yesterday</div>
												</div>
											{:else if widget.type === 'signal'}
												<SignalHealthWidget />
											{/if}
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/each}

					<!-- Empty State Add Widget Card (shown in edit mode when few widgets) -->
					{#if layout.isEditMode && layout.widgets.length < 6}
						<button
							onclick={() => (layout.showWidgetPicker = true)}
							class="dw-page-add-card col-span-1 min-h-[200px] border-2 border-dashed rounded-xl flex flex-col items-center justify-center gap-2 transition-all duration-200 cursor-pointer"
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
	{#if layout.showWidgetPicker}
		<!-- Backdrop -->
		<button
			class="fixed inset-0 bg-black/20 z-40"
			onclick={() => (layout.showWidgetPicker = false)}
			transition:fade={{ duration: 150 }}
			aria-label="Close widget picker"
		></button>

		<!-- Drawer -->
		<div
			class="dw-page-drawer fixed bottom-0 left-0 right-0 rounded-t-2xl shadow-2xl z-50 max-h-[70vh] overflow-hidden"
			transition:fly={{ y: 300, duration: 300 }}
		>
			<!-- Handle -->
			<div class="flex justify-center pt-3 pb-2">
				<div class="dw-page-handle w-10 h-1 rounded-full"></div>
			</div>

			<!-- Header -->
			<div class="dw-page-divider px-6 pb-4 border-b">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="dw-page-text text-lg font-semibold">Add Widget</h2>
						<p class="dw-page-muted text-sm">Choose a widget to add to your dashboard</p>
					</div>
					<button
						onclick={() => (layout.showWidgetPicker = false)}
						class="dw-page-icon-btn inline-flex items-center justify-center p-1.5 rounded-lg transition-colors"
						aria-label="Close widget picker"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Size Picker -->
				<div class="mt-4 flex items-center gap-2">
					<span class="dw-page-muted text-sm">Size:</span>
					<div class="dw-page-seg flex items-center p-0.5 rounded-lg">
						<button
							onclick={() => (layout.pickerSelectedSize = 'small')}
							class="inline-flex items-center justify-center px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {layout.pickerSelectedSize === 'small' ? 'dw-page-seg-active shadow-sm' : 'dw-page-seg-inactive'}"
						>
							Small
						</button>
						<button
							onclick={() => (layout.pickerSelectedSize = 'medium')}
							class="inline-flex items-center justify-center px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {layout.pickerSelectedSize === 'medium' ? 'dw-page-seg-active shadow-sm' : 'dw-page-seg-inactive'}"
						>
							Medium
						</button>
						<button
							onclick={() => (layout.pickerSelectedSize = 'large')}
							class="inline-flex items-center justify-center px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {layout.pickerSelectedSize === 'large' ? 'dw-page-seg-active shadow-sm' : 'dw-page-seg-inactive'}"
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
						{@const isAlreadyAdded = isUnique && layout.addedUniqueTypes.has(widgetOption.type)}
						{@const existingCount = layout.widgets.filter((w) => w.type === widgetOption.type).length}
						<button
							onclick={() => layout.addWidget(widgetOption.type)}
							disabled={isAlreadyAdded}
							class="dw-page-widget-option flex flex-col items-start p-4 border rounded-xl transition-all duration-200 text-left group relative
								{isAlreadyAdded
									? 'dw-page-widget-option--disabled cursor-not-allowed opacity-60'
									: ''}"
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
							<svg class="w-6 h-6 mb-2 dw-page-muted" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d={widgetOption.icon} /></svg>
							<span class="dw-page-text font-medium">{widgetOption.title}</span>
							<span class="dw-page-muted text-xs mt-0.5">{widgetOption.description}</span>
						</button>
					{/each}
				</div>
			</div>
		</div>
	{/if}

	<!-- Undo Toast -->
	{#if layout.showUndoToast && layout.undoStack.length > 0}
		<div
			class="dw-page-toast fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 rounded-xl shadow-xl"
			transition:fly={{ y: 50, duration: 200 }}
		>
			<svg class="w-5 h-5 dw-page-toast-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
			</svg>
			<span class="text-sm">Widget removed</span>
			<button
				onclick={() => layout.undoRemove()}
				class="dw-page-toast-btn inline-flex items-center justify-center gap-1 px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors"
			>
				Undo
			</button>
			<button
				onclick={() => { layout.showUndoToast = false; layout.undoStack = []; }}
				class="dw-page-toast-btn inline-flex items-center justify-center p-1.5 rounded-lg transition-colors"
				aria-label="Dismiss"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	{/if}


	<!-- Analytics Sidepanel -->
	<AnalyticsSidepanel
		isOpen={analytics.showAnalyticsSidepanel}
		onClose={() => (analytics.showAnalyticsSidepanel = false)}
		analytics={analytics.seededAnalytics}
		isLoading={analytics.analyticsLoading}
		onTimeRangeChange={(range) => analytics.handleAnalyticsTimeRangeChange(range)}
	/>
</div>

<style>
	/* Dashboard Page — Foundation design tokens */
	.dw-page {
		background: var(--dbg, #fafafa);
	}

	.dw-page-spinner {
		border-color: var(--dt, #111);
	}

	.dw-page-text {
		color: var(--dt, #111);
	}

	.dw-page-muted {
		color: var(--dt2, #555);
	}

	.dw-page-meta {
		color: var(--dt3, #888);
	}

	.dw-page-card {
		background: var(--dbg, #fff);
		border-color: var(--dbd, #e0e0e0);
	}

	.dw-page-divider {
		border-color: var(--dbd2, #f0f0f0);
	}

	/* Segmented control */
	.dw-page-seg {
		background: var(--dbg2, #f5f5f5);
	}
	.dw-page-seg-active {
		background: var(--dbg, #fff);
		color: var(--dt, #111);
	}
	.dw-page-seg-inactive {
		background: transparent;
		color: var(--dt2, #555);
	}
	.dw-page-seg-inactive:hover {
		color: var(--dt, #111);
	}

	.dw-page-sep {
		background: var(--dbd, #e0e0e0);
	}

	/* Icon buttons (toolbar) */
	.dw-page-icon-btn {
		color: var(--dt2, #555);
	}
	.dw-page-icon-btn:hover {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}

	/* Primary button */
	.dw-page-btn-primary {
		background: var(--dt, #111);
		color: var(--dbg, #fff);
	}
	.dw-page-btn-primary:hover {
		opacity: 0.85;
	}

	/* Edit mode banner */
	.dw-page-edit-banner {
		background: rgba(59, 130, 246, 0.08);
		border-color: rgba(59, 130, 246, 0.25);
	}
	.dw-page-edit-icon {
		background: rgba(59, 130, 246, 0.15);
		color: #3b82f6;
	}
	.dw-page-edit-title {
		color: var(--dt, #111);
	}
	.dw-page-edit-hint {
		color: #3b82f6;
	}
	.dw-page-edit-kbd {
		background: rgba(59, 130, 246, 0.15);
	}

	/* Widget card */
	.dw-page-widget-card {
		background: var(--dbg, #fff);
		border-color: var(--dbd, #e0e0e0);
	}

	/* Edit control buttons */
	.dw-page-ctrl {
		background: var(--dbg, #fff);
		border-color: var(--dbd, #e0e0e0);
	}
	.dw-page-ctrl:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.dw-page-ctrl-icon {
		color: var(--dt2, #555);
	}

	/* Dropdown */
	.dw-page-dropdown {
		background: var(--dbg, #fff);
		border-color: var(--dbd, #e0e0e0);
	}
	.dw-page-dropdown-item {
		color: var(--dt, #111);
	}
	.dw-page-dropdown-item:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.dw-page-dropdown-item--active {
		color: #3b82f6;
		background: rgba(59, 130, 246, 0.08);
	}
	.dw-page-dropdown-sep {
		background: var(--dbd2, #f0f0f0);
	}

	/* Drag handle */
	.dw-page-drag-handle {
		color: var(--dt3, #888);
	}

	/* Collapsed bar */
	.dw-page-collapsed {
		background: var(--dbg2, #f5f5f5);
	}

	/* Analytics flip view */
	.dw-page-analytics-icon {
		background: linear-gradient(135deg, var(--dt2, #555), var(--dt, #111));
	}
	.dw-page-back-btn {
		background: var(--dbg2, rgba(0,0,0,0.05));
		border-color: var(--dbd, rgba(0,0,0,0.08));
		color: var(--dt, #111);
	}
	.dw-page-back-btn:hover {
		background: var(--dbg3, rgba(0,0,0,0.08));
	}
	.dw-page-stat-row:hover {
		background: var(--dbg2, #f5f5f5);
	}

	/* Add widget card */
	.dw-page-add-card {
		border-color: var(--dbd, #ccc);
		color: var(--dt3, #888);
	}
	.dw-page-add-card:hover {
		border-color: #3b82f6;
		color: #3b82f6;
		background: rgba(59, 130, 246, 0.05);
	}

	/* Widget picker drawer */
	.dw-page-drawer {
		background: var(--dbg, #fff);
	}
	.dw-page-handle {
		background: var(--dbd, #ccc);
	}

	/* Widget option cards */
	.dw-page-widget-option {
		background: var(--dbg2, #f5f5f5);
		border-color: var(--dbd, #e0e0e0);
	}
	.dw-page-widget-option:not(.dw-page-widget-option--disabled):hover {
		background: rgba(59, 130, 246, 0.06);
		border-color: rgba(59, 130, 246, 0.4);
	}
	.dw-page-widget-option--disabled {
		background: var(--dbg3, #eee);
	}

	/* Error state */
	.dw-page-error-icon {
		width: 4rem;
		height: 4rem;
		background: color-mix(in srgb, var(--color-error, #ef4444) 10%, transparent);
		color: var(--color-error, #ef4444);
		border-radius: var(--radius-sm, 12px);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 1rem;
	}

	/* Undo toast */
	.dw-page-toast {
		background: var(--dt, #111);
		color: var(--dbg, #fff);
	}
	.dw-page-toast-icon {
		color: var(--dt4, #bbb);
	}
	.dw-page-toast-btn {
		color: var(--dbg, #fff);
	}
	.dw-page-toast-btn:hover {
		background: rgba(255,255,255,0.15);
	}
</style>
