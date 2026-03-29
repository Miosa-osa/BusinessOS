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
		SignalHealthWidget,
		AnalyticsOverviewWidget,
		DashboardRightPanel,
		DashboardEditToolbar
	} from '$lib/components/dashboard';

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

	function toggleRightPanel(): void {
		layout.showRightPanel = !layout.showRightPanel;
	}

	function openWidgetPickerInPanel(): void {
		layout.showRightPanel = true;
		// The right panel's widget tab auto-activates in edit mode via $effect
	}
</script>

<svelte:window onkeydown={(e) => layout.handleKeydown(e)} />

<div class="dw-page">

	{#if data.isLoading}
		<div class="dw-page-center" in:fade>
			<div class="dw-page-loader">
				<div class="dw-page-spinner animate-spin"></div>
				<p class="dw-page-muted">Loading dashboard...</p>
			</div>
		</div>
	{:else if data.error}
		<div class="dw-page-center" in:fade>
			<div class="dw-page-error-card">
				<div class="dw-page-error-icon" aria-hidden="true">
					<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<p class="dw-page-muted">{data.error}</p>
				<button onclick={() => data.loadDashboard()} class="dw-page-btn-primary">
					Try Again
				</button>
			</div>
		</div>
	{:else}
		<!-- Main content area: scrollable left + fixed right panel -->
		<div class="dw-page-body" in:fade={{ duration: 300 }}>

			<!-- Left: scrollable content -->
			<div class="dw-page-main">

				<!-- Top Toolbar -->
				<div class="dw-page-toolbar">
					<div class="dw-page-toolbar-actions">
						<!-- Segmented Control: View / Edit -->
						<div class="dw-page-seg" role="tablist">
							<button
								onclick={() => layout.isEditMode && layout.toggleEditMode()}
								role="tab"
								aria-selected={!layout.isEditMode}
								class="dw-page-seg-btn {!layout.isEditMode ? 'dw-page-seg-btn--active' : ''}"
							>
								<svg class="dw-page-seg-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
								</svg>
								View
							</button>
							<button
								onclick={() => !layout.isEditMode && layout.toggleEditMode()}
								role="tab"
								aria-selected={layout.isEditMode}
								class="dw-page-seg-btn {layout.isEditMode ? 'dw-page-seg-btn--active' : ''}"
							>
								<svg class="dw-page-seg-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
								</svg>
								Edit
							</button>
						</div>

						<!-- Separator -->
						<div class="dw-page-sep"></div>

						<!-- Panel toggle -->
						<button
							onclick={toggleRightPanel}
							class="dw-page-icon-btn"
							title={layout.showRightPanel ? 'Hide panel' : 'Show panel'}
							aria-label={layout.showRightPanel ? 'Hide right panel' : 'Show right panel'}
						>
							<svg class="dw-page-icon-btn-svg" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</button>

					</div>
				</div>

				<!-- Header with Greeting -->
				<div class="dw-page-header-area">
					<DashboardHeader
						userName={$session.data?.user?.name || 'there'}
						energyLevel={data.energyLevel}
						onEnergySet={(level) => data.handleEnergySet(level)}
					/>
				</div>

				<!-- Floating Edit Toolbar -->
				{#if layout.isEditMode}
					<div class="dw-page-toolbar-float">
						<DashboardEditToolbar onOpenWidgetPicker={openWidgetPickerInPanel} />
					</div>
				{/if}

				<!-- Widget Grid -->
				<div class="dw-page-grid-area">
					<div class="dw-page-grid" role={layout.isEditMode ? 'list' : undefined}>
						{#each layout.widgets as widget, index (widget.id)}
							{@const isSelected = layout.isEditMode && layout.selectedWidgetIndex === index}
							<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
							<div
								class="{getWidgetGridClass(widget.size)} dw-page-widget-slot
									{layout.isEditMode ? 'dw-page-widget-slot--editing' : ''}
									{layout.draggedWidget === widget.id ? 'dw-page-widget-slot--dragging' : ''}
									{layout.dragOverWidget === widget.id ? 'dw-page-widget-slot--dragover' : ''}
									{isSelected ? 'dw-page-widget-slot--selected' : ''}"
								role={layout.isEditMode ? 'listitem' : undefined}
								tabindex={layout.isEditMode ? 0 : -1}
								draggable={layout.isEditMode}
								ondragstart={(e) => layout.handleDragStart(e, widget.id)}
								ondragover={(e) => layout.handleDragOver(e)}
								ondragenter={(e) => layout.handleDragEnter(e, widget.id)}
								ondragleave={(e) => layout.handleDragLeave(e)}
								ondrop={(e) => layout.handleDrop(e, widget.id)}
								ondragend={() => layout.handleDragEnd()}
								onclick={() => layout.isEditMode && (layout.selectedWidgetIndex = index)}
								animate:flip={{ duration: 300 }}
								in:fade={{ duration: 200, delay: index * 50 }}
							>
								<!-- Widget Container -->
								<div class="dw-page-widget-inner {layout.isEditMode ? 'dw-page-widget-inner--editing' : ''}">
									<!-- Edit Mode Overlay Controls -->
									{#if layout.isEditMode}
										<div class="dw-page-edit-controls">
											<!-- Collapse Toggle -->
											<button
												onclick={(e) => { e.stopPropagation(); layout.toggleWidgetCollapse(widget.id); }}
												class="dw-page-ctrl"
												title={widget.collapsed ? 'Expand' : 'Collapse'}
											>
												<svg class="dw-page-ctrl-icon {widget.collapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
												</svg>
											</button>

											<!-- Widget Menu -->
											<DropdownMenu.Root>
												<DropdownMenu.Trigger class="dw-page-ctrl">
													<svg class="dw-page-ctrl-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
													</svg>
												</DropdownMenu.Trigger>
												<DropdownMenu.Content
													class="dw-page-dropdown"
													sideOffset={4}
												>
													<!-- Size Options -->
													<DropdownMenu.Sub>
													<DropdownMenu.SubTrigger class="dw-page-dropdown-item dw-page-dropdown-item--between">
															<div class="dw-page-dropdown-item-left">
																<svg class="dw-page-dropdown-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
																</svg>
																Size
															</div>
															<svg class="dw-page-dropdown-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
															</svg>
														</DropdownMenu.SubTrigger>
														<DropdownMenu.SubContent class="dw-page-dropdown dw-page-dropdown--sub">
															<DropdownMenu.Item
																class="dw-page-dropdown-item {widget.size === 'small' ? 'dw-page-dropdown-item--active' : ''}"
																onclick={() => layout.setWidgetSize(widget.id, 'small')}
															>
																Small
															</DropdownMenu.Item>
															<DropdownMenu.Item
																class="dw-page-dropdown-item {widget.size === 'medium' ? 'dw-page-dropdown-item--active' : ''}"
																onclick={() => layout.setWidgetSize(widget.id, 'medium')}
															>
																Medium
															</DropdownMenu.Item>
															<DropdownMenu.Item
																class="dw-page-dropdown-item {widget.size === 'large' ? 'dw-page-dropdown-item--active' : ''}"
																onclick={() => layout.setWidgetSize(widget.id, 'large')}
															>
																Large
															</DropdownMenu.Item>
														</DropdownMenu.SubContent>
													</DropdownMenu.Sub>

													<!-- Color Options -->
													<DropdownMenu.Sub>
													<DropdownMenu.SubTrigger class="dw-page-dropdown-item dw-page-dropdown-item--between">
															<div class="dw-page-dropdown-item-left">
																<svg class="dw-page-dropdown-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
																</svg>
																Color
															</div>
															<svg class="dw-page-dropdown-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
															</svg>
														</DropdownMenu.SubTrigger>
														<DropdownMenu.SubContent class="dw-page-dropdown dw-page-dropdown--sub">
															{#each accentColors as color}
																<DropdownMenu.Item
																	class="dw-page-dropdown-item {widget.accentColor === color.value ? 'dw-page-dropdown-item--active' : ''}"
																	onclick={() => layout.setWidgetAccentColor(widget.id, color.value)}
																>
																	<span class="dw-page-color-swatch {getAccentColorClass(color.value)}"></span>
																	{color.name}
																</DropdownMenu.Item>
															{/each}
														</DropdownMenu.SubContent>
													</DropdownMenu.Sub>

													<DropdownMenu.Separator class="dw-page-dropdown-sep" />
													<DropdownMenu.Item
														class="dw-page-dropdown-item dw-page-dropdown-item--danger"
														onclick={() => layout.removeWidget(widget.id)}
													>
														<svg class="dw-page-dropdown-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
														</svg>
														Remove
													</DropdownMenu.Item>
												</DropdownMenu.Content>
											</DropdownMenu.Root>
										</div>
									{/if}

									<!-- Widget Card -->
									<div class="dw-page-widget-card {getAccentBorderClass(widget.accentColor)}
										{isSelected ? 'dw-page-widget-card--selected' : layout.isEditMode ? 'dw-page-widget-card--editing' : ''}"
									>
										<!-- Analytics Toggle Icon (appears on hover, not in edit mode) -->
										{#if !layout.isEditMode && !widget.collapsed && !widget.showAnalytics}
											<button
												onclick={(e) => { e.stopPropagation(); layout.toggleWidgetAnalytics(widget.id); }}
												class="dw-page-ctrl dw-page-analytics-toggle"
												title="View Analytics"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
												</svg>
											</button>
										{/if}

										<!-- Drag Handle -->
										{#if layout.isEditMode}
											<div class="dw-page-drag-handle" aria-hidden="true">
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M8 6a2 2 0 11-4 0 2 2 0 014 0zM8 12a2 2 0 11-4 0 2 2 0 014 0zM8 18a2 2 0 11-4 0 2 2 0 014 0zM14 6a2 2 0 11-4 0 2 2 0 014 0zM14 12a2 2 0 11-4 0 2 2 0 014 0zM14 18a2 2 0 11-4 0 2 2 0 014 0z" />
												</svg>
											</div>
										{/if}

										<!-- Collapsed Title Bar -->
										{#if widget.collapsed}
											<div class="dw-page-collapsed">
												<span class="dw-page-collapsed-title">{widget.title}</span>
												<button
													onclick={() => layout.toggleWidgetCollapse(widget.id)}
													class="dw-page-collapsed-expand"
													aria-label="Expand widget"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
													</svg>
												</button>
											</div>
										{:else if widget.showAnalytics}
											<!-- Analytics Flip View -->
											<div class="dw-page-analytics-view" transition:fade={{ duration: 200 }}>
												<div class="dw-page-analytics-header">
													<div class="dw-page-analytics-title-group">
														<div class="dw-page-analytics-icon-wrap">
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" style="color: var(--dbg)">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
															</svg>
														</div>
														<span class="dw-page-analytics-title">{widgetAnalytics[widget.type].title}</span>
													</div>
													<button
														onclick={(e) => { e.stopPropagation(); layout.toggleWidgetAnalytics(widget.id); }}
														class="dw-page-back-btn"
													>
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
														</svg>
														Back
													</button>
												</div>

												<div class="dw-page-stat-list">
													{#each widgetAnalytics[widget.type].stats as stat}
														<div class="dw-page-stat-row">
															<span class="dw-page-stat-label">{stat.label}</span>
															<div class="dw-page-stat-value-group">
																<span class="dw-page-stat-value">{stat.value}</span>
																{#if stat.trend}
																	<span class="dw-page-stat-trend {stat.trend.startsWith('+') ? 'dw-page-stat-trend--up' : 'dw-page-stat-trend--down'}">{stat.trend}</span>
																{/if}
															</div>
														</div>
													{/each}
												</div>

												<button
													onclick={() => { layout.showRightPanel = true; }}
													class="dw-page-btn-primary dw-page-full-analytics-btn"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
													</svg>
													View Full Analytics
												</button>
											</div>
										{:else}
											<!-- Widget Content -->
											<div class="{layout.isEditMode ? 'dw-page-widget-content--locked' : ''}">
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
													<div class="dw-page-metric-placeholder">
														<div class="dw-page-metric-top">
															<span class="dw-page-metric-label">Tasks Due Today</span>
															<span class="dw-page-metric-badge">+12%</span>
														</div>
														<div class="dw-page-metric-value">8</div>
														<div class="dw-page-metric-sub">vs 7 yesterday</div>
													</div>
												{:else if widget.type === 'signal'}
													<SignalHealthWidget />
												{:else if widget.type === 'analytics-overview'}
													<AnalyticsOverviewWidget />
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
								onclick={openWidgetPickerInPanel}
								class="dw-page-add-card"
							>
								<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4v16m8-8H4" />
								</svg>
								<span class="dw-page-add-card-label">Add Widget</span>
							</button>
						{/if}
					</div>
				</div>
			</div>

			<!-- Right Panel -->
			<div class="dw-page-panel-area" class:dw-page-panel-area--open={layout.showRightPanel}>
				<DashboardRightPanel
					isOpen={layout.showRightPanel}
					onToggle={toggleRightPanel}
				/>
			</div>
		</div>
	{/if}

	<!-- Undo Toast -->
	{#if layout.showUndoToast && layout.undoStack.length > 0}
		<div
			class="dw-page-toast"
			transition:fly={{ y: 50, duration: 200 }}
		>
			<svg class="dw-page-toast-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
			</svg>
			<span class="dw-page-toast-text">Widget removed</span>
			<button
				onclick={() => layout.undoRemove()}
				class="dw-page-toast-btn"
			>
				Undo
			</button>
			<button
				onclick={() => { layout.showUndoToast = false; layout.undoStack = []; }}
				class="dw-page-toast-btn dw-page-toast-btn--dismiss"
				aria-label="Dismiss"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	{/if}
</div>

<style>
	/* ── Page shell ─────────────────────────────────────────────────────────────── */
	.dw-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg);
		position: relative;
	}

	/* ── Loading / error centered states ────────────────────────────────────────── */
	.dw-page-center {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.dw-page-loader {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-3);
	}

	.dw-page-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dt);
		border-top-color: transparent;
		border-radius: 50%;
	}

	.dw-page-muted {
		font-size: var(--text-sm);
		color: var(--dt2);
	}

	.dw-page-error-card {
		text-align: center;
		padding: var(--space-8);
		border-radius: var(--radius-lg);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		box-shadow: var(--shadow-sm);
		max-width: 28rem;
	}

	.dw-page-error-icon {
		width: 4rem;
		height: 4rem;
		background: color-mix(in srgb, var(--color-error, #ef4444) 10%, transparent);
		color: var(--color-error, #ef4444);
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-4);
	}

	/* ── Body: main + right panel ──────────────────────────────────────────────── */
	.dw-page-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	.dw-page-main {
		flex: 1;
		min-width: 0;
		overflow-y: auto;
		overflow-x: hidden;
	}

	.dw-page-panel-area {
		position: relative;
		flex-shrink: 0;
		width: 0;
		transition: width 240ms ease;
	}

	.dw-page-panel-area--open {
		width: 320px;
	}

	/* ── Toolbar ───────────────────────────────────────────────────────────────── */
	.dw-page-toolbar {
		padding: var(--space-4) var(--space-6) var(--space-2);
	}

	.dw-page-toolbar-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
	}

	/* Segmented control */
	.dw-page-seg {
		display: flex;
		align-items: center;
		padding: 0.25rem;
		border-radius: var(--radius-md);
		background: var(--dbg2);
	}

	.dw-page-seg-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-1);
		padding: 0.375rem 0.625rem;
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		border-radius: var(--radius-md);
		border: none;
		cursor: pointer;
		transition: background 150ms, color 150ms, box-shadow 150ms;
		background: transparent;
		color: var(--dt2);
	}

	.dw-page-seg-btn:hover:not(.dw-page-seg-btn--active) {
		color: var(--dt);
	}

	.dw-page-seg-btn--active {
		background: var(--dbg);
		color: var(--dt);
		box-shadow: var(--shadow-sm);
	}

	.dw-page-seg-icon {
		width: 1rem;
		height: 1rem;
	}

	.dw-page-sep {
		height: 1.5rem;
		width: 1px;
		background: var(--dbd);
	}

	/* Icon buttons */
	.dw-page-icon-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem;
		border-radius: var(--radius-md);
		border: none;
		background: transparent;
		color: var(--dt2);
		cursor: pointer;
		transition: background 150ms, color 150ms;
	}

	.dw-page-icon-btn:hover {
		background: var(--dbg2);
		color: var(--dt);
	}

	.dw-page-icon-btn-svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	/* Primary button */
	.dw-page-btn-primary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		padding: 0.5rem 0.875rem;
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		border-radius: var(--radius-md);
		border: none;
		background: var(--dt);
		color: var(--dbg);
		cursor: pointer;
		transition: opacity 150ms;
	}

	.dw-page-btn-primary:hover {
		opacity: 0.85;
	}

	/* ── Header area ───────────────────────────────────────────────────────────── */
	.dw-page-header-area {
		padding: 0 var(--space-6);
	}

	/* ── Floating edit toolbar wrapper ─────────────────────────────────────────── */
	.dw-page-toolbar-float {
		padding: 0 var(--space-6);
		margin-bottom: var(--space-3);
	}

	/* ── Widget grid ───────────────────────────────────────────────────────────── */
	.dw-page-grid-area {
		padding: var(--space-4) var(--space-6) var(--space-8);
	}

	.dw-page-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-5);
	}

	@media (min-width: 768px) {
		.dw-page-grid {
			grid-template-columns: repeat(2, 1fr);
			gap: var(--space-5);
		}
	}

	@media (min-width: 1280px) {
		.dw-page-grid {
			grid-template-columns: repeat(3, 1fr);
			gap: var(--space-6);
		}
	}

	/* ── Widget size classes ───────────────────────────────────────────────────── */
	/* Small: always 1 column */
	:global(.dw-widget-sm) {
		grid-column: span 1;
	}

	/* Medium: 1 col on mobile, 2 cols on large screens */
	:global(.dw-widget-md) {
		grid-column: span 1;
	}

	@media (min-width: 1280px) {
		:global(.dw-widget-md) {
			grid-column: span 2;
		}
	}

	/* Large: full width row */
	:global(.dw-widget-lg) {
		grid-column: span 1;
	}

	@media (min-width: 768px) {
		:global(.dw-widget-lg) {
			grid-column: span 2;
		}
	}

	@media (min-width: 1280px) {
		:global(.dw-widget-lg) {
			grid-column: span 3;
		}
	}

	/* ── Widget slot states ────────────────────────────────────────────────────── */
	.dw-page-widget-slot {
		transition: transform 200ms ease, opacity 200ms ease;
		border-radius: var(--radius-lg);
	}

	.dw-page-widget-slot--editing {
		cursor: grab;
	}

	.dw-page-widget-slot--editing:active {
		cursor: grabbing;
	}

	.dw-page-widget-slot--dragging {
		opacity: 0.4;
		transform: scale(0.96);
	}

	.dw-page-widget-slot--dragover {
		transform: scale(1.02);
	}

	.dw-page-widget-slot--dragover .dw-page-widget-card {
		border-color: var(--accent-blue, #3b82f6);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-blue, #3b82f6) 25%, transparent),
			var(--shadow-md);
	}

	.dw-page-widget-slot--selected {
		transform: scale(1.01);
	}

	/* ── Widget inner container ────────────────────────────────────────────────── */
	.dw-page-widget-inner {
		position: relative;
		height: 100%;
	}

	.dw-page-widget-inner--editing {
		padding-top: var(--space-2);
	}

	/* ── Edit controls (above widget) ──────────────────────────────────────────── */
	.dw-page-edit-controls {
		position: absolute;
		top: 0;
		right: 0.25rem;
		z-index: 10;
		display: flex;
		gap: 0.25rem;
	}

	.dw-page-ctrl {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.375rem;
		border-radius: var(--radius-md);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		box-shadow: var(--shadow-sm);
		cursor: pointer;
		transition: background 150ms, color 150ms;
		color: var(--dt2);
	}

	.dw-page-ctrl:hover {
		background: var(--dbg2);
		color: var(--dt);
	}

	.dw-page-ctrl-icon {
		width: 0.75rem;
		height: 0.75rem;
		transition: transform 200ms;
	}

	/* ── Widget card ───────────────────────────────────────────────────────────── */
	.dw-page-widget-card {
		border-radius: var(--radius-lg);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		transition: border-color 200ms ease, box-shadow 200ms ease;
		overflow: hidden;
		position: relative;
		box-shadow: var(--shadow-sm);
		height: 100%;
	}

	.dw-page-widget-card:hover {
		box-shadow: var(--shadow-md);
		border-color: var(--dbd2);
	}

	.dw-page-widget-card--editing {
		border: 2px dashed color-mix(in srgb, var(--dt) 20%, transparent);
	}

	.dw-page-widget-card--editing:hover {
		border-color: color-mix(in srgb, var(--dt) 35%, transparent);
	}

	.dw-page-widget-card--selected {
		border: 2px solid var(--accent-blue, #3b82f6);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-blue, #3b82f6) 15%, transparent);
	}

	/* Lock widget interactions in edit mode */
	.dw-page-widget-content--locked {
		pointer-events: none;
	}

	/* Analytics toggle (appears on hover) */
	.dw-page-analytics-toggle {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		z-index: 20;
		opacity: 0;
		transition: opacity 150ms;
	}

	.dw-page-widget-card:hover .dw-page-analytics-toggle {
		opacity: 1;
	}

	/* Drag handle */
	.dw-page-drag-handle {
		position: absolute;
		top: 0.75rem;
		left: 0.625rem;
		z-index: 10;
		color: var(--dt4);
		opacity: 0.5;
		transition: opacity 150ms, color 150ms;
		cursor: grab;
	}

	.dw-page-drag-handle:hover {
		opacity: 1;
		color: var(--dt2);
	}

	.dw-page-widget-slot--editing:active .dw-page-drag-handle {
		cursor: grabbing;
		opacity: 1;
	}

	/* ── Collapsed bar ─────────────────────────────────────────────────────────── */
	.dw-page-collapsed {
		padding: var(--space-3) var(--space-4);
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--dbg2);
	}

	.dw-page-collapsed-title {
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		color: var(--dt);
	}

	.dw-page-collapsed-expand {
		color: var(--dt3);
		background: transparent;
		border: none;
		cursor: pointer;
		padding: 0;
		display: flex;
	}

	.dw-page-collapsed-expand:hover {
		color: var(--dt);
	}

	/* ── Analytics flip view ───────────────────────────────────────────────────── */
	.dw-page-analytics-view {
		padding: var(--space-5);
	}

	.dw-page-analytics-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-4);
		padding-bottom: var(--space-3);
		border-bottom: 1px solid var(--dbd2);
	}

	.dw-page-analytics-title-group {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.dw-page-analytics-icon-wrap {
		width: 2rem;
		height: 2rem;
		border-radius: var(--radius-md);
		background: linear-gradient(135deg, var(--dt2), var(--dt));
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-sm);
	}

	.dw-page-analytics-title {
		font-size: var(--text-sm);
		font-weight: var(--font-semibold);
		color: var(--dt);
	}

	.dw-page-back-btn {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		padding: 0.375rem 0.625rem;
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		border-radius: var(--radius-md);
		border: 1px solid var(--dbd);
		background: var(--dbg2);
		color: var(--dt);
		cursor: pointer;
		transition: background 150ms;
	}

	.dw-page-back-btn:hover {
		background: var(--dbg3);
	}

	.dw-page-stat-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.dw-page-stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.625rem 0.75rem;
		border-radius: var(--radius-md);
		transition: background 150ms;
	}

	.dw-page-stat-row:hover {
		background: var(--dbg2);
	}

	.dw-page-stat-label {
		font-size: var(--text-sm);
		color: var(--dt2);
	}

	.dw-page-stat-value-group {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.dw-page-stat-value {
		font-size: var(--text-sm);
		font-weight: var(--font-semibold);
		color: var(--dt);
	}

	.dw-page-stat-trend {
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		padding: 0.125rem 0.375rem;
		border-radius: var(--radius-sm);
	}

	.dw-page-stat-trend--up {
		color: var(--color-success, #16a34a);
		background: color-mix(in srgb, var(--color-success, #22c55e) 10%, transparent);
	}

	.dw-page-stat-trend--down {
		color: var(--color-error, #dc2626);
		background: color-mix(in srgb, var(--color-error, #ef4444) 10%, transparent);
	}

	.dw-page-full-analytics-btn {
		width: 100%;
		margin-top: var(--space-4);
	}

	/* ── Metric placeholder ────────────────────────────────────────────────────── */
	.dw-page-metric-placeholder {
		padding: var(--space-5);
	}

	.dw-page-metric-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-3);
	}

	.dw-page-metric-label {
		font-size: var(--text-sm);
		color: var(--dt2);
	}

	.dw-page-metric-badge {
		font-size: var(--text-xs);
		color: var(--color-success, #16a34a);
		background: color-mix(in srgb, var(--color-success, #22c55e) 10%, transparent);
		padding: 0.125rem 0.5rem;
		border-radius: var(--radius-full);
	}

	.dw-page-metric-value {
		font-size: 1.875rem;
		font-weight: var(--font-bold);
		color: var(--dt);
	}

	.dw-page-metric-sub {
		font-size: var(--text-xs);
		color: var(--dt3);
		margin-top: var(--space-1);
	}

	/* ── Add widget card ───────────────────────────────────────────────────────── */
	.dw-page-add-card {
		min-height: 200px;
		border: 2px dashed var(--dbd);
		border-radius: var(--radius-lg);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		color: var(--dt3);
		background: transparent;
		cursor: pointer;
		transition: border-color 200ms, color 200ms, background 200ms;
	}

	.dw-page-add-card:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
		background: color-mix(in srgb, var(--accent-blue, #3b82f6) 5%, transparent);
	}

	.dw-page-add-card-label {
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
	}

	/* ── Dropdown ──────────────────────────────────────────────────────────────── */
	.dw-page-dropdown {
		z-index: 50;
		min-width: 160px;
		border-radius: var(--radius-md);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		box-shadow: var(--shadow-lg);
		padding: var(--space-1) 0;
	}

	.dw-page-dropdown--sub {
		min-width: 120px;
	}

	.dw-page-dropdown-item {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-3);
		font-size: var(--text-sm);
		color: var(--dt);
		cursor: pointer;
		transition: background 120ms;
	}

	.dw-page-dropdown-item:hover {
		background: var(--dbg2);
	}

	.dw-page-dropdown-item--between {
		justify-content: space-between;
		width: 100%;
	}

	.dw-page-dropdown-item-left {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.dw-page-dropdown-item--active {
		color: var(--accent-blue, #3b82f6);
		background: color-mix(in srgb, var(--accent-blue, #3b82f6) 8%, transparent);
	}

	.dw-page-dropdown-item--danger {
		color: var(--color-error, #ef4444);
	}

	.dw-page-dropdown-item--danger:hover {
		background: color-mix(in srgb, var(--color-error, #ef4444) 10%, transparent);
	}

	.dw-page-dropdown-icon {
		width: 1rem;
		height: 1rem;
	}

	.dw-page-dropdown-chevron {
		width: 0.75rem;
		height: 0.75rem;
	}

	.dw-page-dropdown-sep {
		margin: var(--space-1) 0;
		height: 1px;
		background: var(--dbd2);
	}

	.dw-page-color-swatch {
		width: 0.75rem;
		height: 0.75rem;
		border-radius: 50%;
	}

	/* ── Undo toast ────────────────────────────────────────────────────────────── */
	.dw-page-toast {
		position: fixed;
		bottom: 1.5rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 50;
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-lg);
		background: var(--dt);
		color: var(--dbg);
		box-shadow: var(--shadow-xl);
	}

	.dw-page-toast-icon {
		width: 1.25rem;
		height: 1.25rem;
		color: var(--dt4);
	}

	.dw-page-toast-text {
		font-size: var(--text-sm);
	}

	.dw-page-toast-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-1);
		padding: 0.375rem 0.625rem;
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		border-radius: var(--radius-md);
		border: none;
		background: transparent;
		color: var(--dbg);
		cursor: pointer;
		transition: background 150ms;
	}

	.dw-page-toast-btn:hover {
		background: rgba(255, 255, 255, 0.15);
	}

	.dw-page-toast-btn--dismiss {
		padding: 0.375rem;
	}
</style>
