<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { getInstallations, getModule, uninstallModule, installModule, updateInstallation } from '$lib/api/modules';
	import type { ModuleInstallation, CustomModule, ModuleCategory } from '$lib/types/modules';
	import { desktop3dStore, DYNAMIC_MODULE_COLORS } from '$lib/stores/desktop3dStore';
	import { windowStore } from '$lib/stores/windowStore';
	import {
		Package,
		Trash2,
		Loader2,
		AlertCircle,
		ArrowLeft,
		Check,
		X,
		Box,
		DollarSign,
		Zap,
		MessageSquare,
		BarChart3,
		Plug,
		Wrench,
		Sparkles,
		Download,
		CircleCheck,
		CircleDot,
		Circle,
		GripVertical,
		ChevronRight,
		LayoutDashboard,
		CheckSquare,
		FolderKanban,
		Users,
		Users2,
		Building2,
		Table,
		FileText,
		Inbox,
		Network,
		Bot,
		Terminal,
		Code2,
		CalendarDays
	} from 'lucide-svelte';
	import { getAllBuiltinModules, type BuiltinModuleConfig } from '$lib/config/builtinModuleConfigs';

	// ── Types ──────────────────────────────────────────────────────────────

	interface InstalledModuleView {
		installation: ModuleInstallation;
		module: CustomModule | null;
	}

	type InstallStep = 'validating' | 'migrating' | 'registering' | 'done';

	const INSTALL_STEPS: { key: InstallStep; label: string }[] = [
		{ key: 'validating', label: 'Validating manifest...' },
		{ key: 'migrating', label: 'Running migrations...' },
		{ key: 'registering', label: 'Registering routes...' },
		{ key: 'done', label: 'Done' },
	];

	// ── State ──────────────────────────────────────────────────────────────

	let installedModules = $state<InstalledModuleView[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let useMockData = $state(false);

	// Uninstall dialog state
	let showUninstallDialog = $state(false);
	let uninstallTarget = $state<InstalledModuleView | null>(null);
	let isUninstalling = $state(false);
	let uninstallError = $state<string | null>(null);

	// Install dialog state
	let showInstallDialog = $state(false);
	let installTarget = $state<CustomModule | null>(null);
	let isInstalling = $state(false);
	let installError = $state<string | null>(null);
	let installStep = $state<InstallStep | null>(null);

	// ── Built-in modules ──────────────────────────────────────────────────────

	const builtinModules: BuiltinModuleConfig[] = getAllBuiltinModules();

	// ── Module ordering (drag-and-drop) ────────────────────────────────────────

	const ORDER_KEY = 'bos_module_order';

	// Combined list of all module slugs for ordering purposes.
	// Built-ins always precede custom installations in the default order.
	let moduleOrder = $state<string[]>((() => {
		try {
			const stored = localStorage.getItem(ORDER_KEY);
			return stored ? (JSON.parse(stored) as string[]) : [];
		} catch {
			return [];
		}
	})());

	let dragIndex = $state<number | null>(null);
	let dropIndex = $state<number | null>(null);

	function saveOrder(slugs: string[]) {
		try {
			localStorage.setItem(ORDER_KEY, JSON.stringify(slugs));
		} catch {
			// localStorage unavailable — silently skip
		}
	}

	// Returns a stable ordered slug list for the combined view.
	// Slugs not yet in storage are appended in their natural order.
	function getOrderedSlugs(
		builtins: BuiltinModuleConfig[],
		customs: InstalledModuleView[]
	): string[] {
		const allSlugs = [
			...builtins.map((b) => b.id),
			...customs.map((c) => c.installation.module_id),
		];
		if (moduleOrder.length === 0) return allSlugs;
		const ordered = moduleOrder.filter((s) => allSlugs.includes(s));
		const missing = allSlugs.filter((s) => !ordered.includes(s));
		return [...ordered, ...missing];
	}

	// ── Drag-and-drop handlers ─────────────────────────────────────────────────

	function handleDragStart(index: number) {
		dragIndex = index;
	}

	function handleDragOver(e: DragEvent, index: number) {
		e.preventDefault();
		dropIndex = index;
	}

	function handleDrop(combinedSlugs: string[], index: number) {
		if (dragIndex === null || dragIndex === index) {
			dragIndex = null;
			dropIndex = null;
			return;
		}
		const reordered = [...combinedSlugs];
		const [moved] = reordered.splice(dragIndex, 1);
		reordered.splice(index > dragIndex ? index - 1 : index, 0, moved);
		moduleOrder = reordered;
		saveOrder(reordered);
		dragIndex = null;
		dropIndex = null;
	}

	function handleDragEnd() {
		dragIndex = null;
		dropIndex = null;
	}

	// ── OptimalOS indicator ────────────────────────────────────────────────────

	function hasOptimalOsConfig(slug: string): boolean {
		try {
			return localStorage.getItem(`bos_module_config_${slug}`) !== null;
		} catch {
			return false;
		}
	}

	// ── Lucide icon resolver for built-in modules ──────────────────────────────

	// Builtin configs use kebab-case icon slugs (e.g. "layout-dashboard", "check-square").
	const builtinIconMap: Record<string, typeof Package> = {
		'layout-dashboard': LayoutDashboard,
		'message-square': MessageSquare,
		'check-square': CheckSquare,
		'folder-kanban': FolderKanban,
		'folder': FolderKanban,
		'users': Users,
		'users-2': Users2,
		'building-2': Building2,
		'briefcase': Wrench,
		'table': Table,
		'file-text': FileText,
		'inbox': Inbox,
		'network': Network,
		'git-branch': Network,
		'bot': Bot,
		'terminal': Terminal,
		'code-2': Code2,
		'calendar-days': CalendarDays,
		'calendar': CalendarDays,
		'trending-up': BarChart3,
		'zap': Zap,
		'plug': Plug,
		'sparkles': Sparkles,
		'bar-chart-3': BarChart3,
	};

	function getBuiltinIcon(iconName: string): typeof Package {
		return builtinIconMap[iconName] ?? Box;
	}

	// ── Mock Data ──────────────────────────────────────────────────────────

	const MOCK_MODULES: InstalledModuleView[] = [
		{
			installation: {
				id: 'inst-001',
				module_id: 'mod-invoicing',
				workspace_id: 'ws-1',
				user_id: 'user-1',
				config: null,
				is_active: true,
				installed_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
			},
			module: {
				id: 'mod-invoicing',
				workspace_id: 'ws-1',
				creator_id: 'user-1',
				name: 'Invoicing',
				description: 'Generate and send professional invoices to clients. Track payment status and send reminders automatically.',
				category: 'finance',
				icon: 'dollar-sign',
				manifest: { name: 'Invoicing', version: '1.0.0', description: '', author: 'System', category: 'finance', actions: [] },
				config_schema: null,
				visibility: 'workspace',
				is_active: true,
				version: '1.0.0',
				install_count: 12,
				star_count: 5,
				created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
			},
		},
		{
			installation: {
				id: 'inst-002',
				module_id: 'mod-time-tracker',
				workspace_id: 'ws-1',
				user_id: 'user-1',
				config: null,
				is_active: true,
				installed_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
			},
			module: {
				id: 'mod-time-tracker',
				workspace_id: 'ws-1',
				creator_id: 'user-1',
				name: 'Time Tracker',
				description: 'Track time spent on tasks and projects. Generate timesheets and productivity reports.',
				category: 'productivity',
				icon: 'zap',
				manifest: { name: 'Time Tracker', version: '1.0.0', description: '', author: 'System', category: 'productivity', actions: [] },
				config_schema: null,
				visibility: 'workspace',
				is_active: true,
				version: '1.0.0',
				install_count: 8,
				star_count: 3,
				created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
			},
		},
		{
			installation: {
				id: 'inst-003',
				module_id: 'mod-slack-bridge',
				workspace_id: 'ws-1',
				user_id: 'user-1',
				config: null,
				is_active: false,
				installed_at: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
			},
			module: {
				id: 'mod-slack-bridge',
				workspace_id: 'ws-1',
				creator_id: 'user-1',
				name: 'Slack Bridge',
				description: 'Sync messages and notifications between your workspace and Slack channels.',
				category: 'communication',
				icon: 'message-square',
				manifest: { name: 'Slack Bridge', version: '0.9.0', description: '', author: 'System', category: 'communication', actions: [] },
				config_schema: null,
				visibility: 'workspace',
				is_active: false,
				version: '0.9.0',
				install_count: 4,
				star_count: 1,
				created_at: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
			},
		},
	];

	// ── Category icon mapping ──────────────────────────────────────────────

	const categoryIcons: Record<string, typeof Package> = {
		finance: DollarSign,
		productivity: Zap,
		communication: MessageSquare,
		analytics: BarChart3,
		integration: Plug,
		utilities: Wrench,
		automation: Sparkles,
		custom: Box,
	};

	function getCategoryIcon(category: string) {
		return categoryIcons[category] ?? Box;
	}

	// ── Status helpers ─────────────────────────────────────────────────────

	function getStatusLabel(isActive: boolean): string {
		return isActive ? 'Installed' : 'Disabled';
	}

	function getStatusStyle(isActive: boolean): string {
		return isActive
			? 'background: var(--bos-status-success-bg); color: var(--bos-status-success);'
			: '';
	}

	function getStatusClasses(isActive: boolean): string {
		return isActive ? '' : 'st-mod-status-disabled';
	}

	// ── Relative time ──────────────────────────────────────────────────────

	function relativeTime(iso: string): string {
		const diff = Date.now() - new Date(iso).getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(minutes / 60);
		const days = Math.floor(hours / 24);

		if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
		if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
		if (minutes > 0) return `${minutes} min ago`;
		return 'Just now';
	}

	// ── Truncate ───────────────────────────────────────────────────────────

	function truncate(text: string, max: number): string {
		return text.length > max ? text.slice(0, max) + '...' : text;
	}

	// ── Data loading ───────────────────────────────────────────────────────

	async function loadInstalledModules() {
		isLoading = true;
		error = null;

		try {
			const result = await getInstallations();
			const installations = result.installations ?? [];

			// Fetch module details for each installation
			const views: InstalledModuleView[] = await Promise.all(
				installations.map(async (inst) => {
					let mod: CustomModule | null = null;
					try {
						mod = await getModule(inst.module_id);
					} catch {
						// Module details unavailable — show installation anyway
					}
					return { installation: inst, module: mod };
				})
			);

			installedModules = views;
			useMockData = false;
		} catch (err) {
			// Only fall back to empty state if the API endpoint doesn't exist yet (404).
			// For real errors (network failure, 500, auth) show an error state instead.
			const message = err instanceof Error ? err.message : String(err);
			const is404 = message.includes('404') || message.includes('Not Found');
			if (is404) {
				installedModules = [];
			} else {
				console.error('[Modules] Failed to load installations:', err);
				error = message || 'Failed to load modules';
			}
		} finally {
			isLoading = false;
		}
	}

	// ── Uninstall flow ─────────────────────────────────────────────────────

	function openUninstallDialog(view: InstalledModuleView) {
		uninstallTarget = view;
		uninstallError = null;
		showUninstallDialog = true;
	}

	function closeUninstallDialog() {
		showUninstallDialog = false;
		uninstallTarget = null;
		uninstallError = null;
	}

	async function confirmUninstall() {
		if (!uninstallTarget) return;

		isUninstalling = true;
		uninstallError = null;

		try {
			if (!useMockData) {
				await uninstallModule(uninstallTarget.installation.module_id);
			}

			// Remove from 3D desktop + dock
			desktop3dStore.removeModule(uninstallTarget.installation.module_id);
			windowStore.removeFromDock(uninstallTarget.installation.module_id);

			// Remove from list
			const targetId = uninstallTarget.installation.id;
			installedModules = installedModules.filter((m) => m.installation.id !== targetId);

			closeUninstallDialog();
		} catch (err) {
			uninstallError = mapApiError(err, 'Uninstall failed');
		} finally {
			isUninstalling = false;
		}
	}

	// ── Install flow ──────────────────────────────────────────────────────

	function openInstallDialog(mod: CustomModule) {
		installTarget = mod;
		installError = null;
		installStep = null;
		showInstallDialog = true;
	}

	function closeInstallDialog() {
		showInstallDialog = false;
		installTarget = null;
		installError = null;
		installStep = null;
	}

	async function confirmInstall() {
		if (!installTarget) return;

		isInstalling = true;
		installError = null;

		try {
			// Step progress — show each step with a brief pause for UX
			installStep = 'validating';
			await delay(600);

			installStep = 'migrating';
			let result: ModuleInstallation;
			if (!useMockData) {
				result = await installModule(installTarget.id);
			} else {
				await delay(800);
				result = {
					id: `inst-mock-${Date.now()}`,
					module_id: installTarget.id,
					workspace_id: installTarget.workspace_id,
					user_id: 'user-1',
					config: null,
					is_active: true,
					installed_at: new Date().toISOString(),
					updated_at: new Date().toISOString(),
				};
			}

			installStep = 'registering';
			await delay(400);

			installStep = 'done';

			// Add to 3D desktop + dock
			desktop3dStore.addModule({
				id: installTarget.id,
				title: installTarget.name,
				icon: installTarget.icon ?? 'box',
				color: DYNAMIC_MODULE_COLORS[installTarget.category] ?? DYNAMIC_MODULE_COLORS.general,
				category: installTarget.category,
			});
			windowStore.addToDock(installTarget.id);

			// Add to list
			installedModules = [
				...installedModules,
				{ installation: result, module: installTarget },
			];

			// Brief pause to show "Done" step, then close
			await delay(500);
			closeInstallDialog();
		} catch (err) {
			installError = mapApiError(err, 'Installation failed');
			installStep = null;
		} finally {
			isInstalling = false;
		}
	}

	// ── Re-enable a disabled module ───────────────────────────────────────
	// A disabled module is already installed — it only needs is_active: true.
	// We try PATCH first (targeted re-enable). If the endpoint isn't supported
	// yet we fall back to the full install flow as a stopgap.

	async function reinstallModule(view: InstalledModuleView) {
		if (!view.module) return;

		if (useMockData) {
			// Mock mode: optimistically flip is_active in the local list
			installedModules = installedModules.map((m) =>
				m.installation.id === view.installation.id
					? { ...m, installation: { ...m.installation, is_active: true } }
					: m
			);
			return;
		}

		try {
			// Re-enable via the typed API layer
			await updateInstallation(view.installation.id, { is_active: true });
			// Refresh the list to reflect the updated state
			await loadInstalledModules();
		} catch {
			// Endpoint not supported yet — fall back to full install dialog
			openInstallDialog(view.module);
		}
	}

	// ── Error mapping ─────────────────────────────────────────────────────

	function mapApiError(err: unknown, fallback: string): string {
		const message = err instanceof Error ? err.message : fallback;
		if (message.includes('MANIFEST_INVALID')) {
			return 'The module manifest is invalid.';
		}
		if (message.includes('PROTECTION_VIOLATION')) {
			return 'This module conflicts with a protected module.';
		}
		if (message.includes('SQL_UNSAFE')) {
			return 'The module contains unsafe database operations.';
		}
		return message;
	}

	// ── Utility ───────────────────────────────────────────────────────────

	function delay(ms: number): Promise<void> {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	// ── Lifecycle ──────────────────────────────────────────────────────────

	onMount(() => {
		loadInstalledModules();
	});
</script>

<div class="h-full overflow-y-auto st-page-bg">
	<div class="max-w-4xl mx-auto px-6 py-8">
		<!-- Back link + header -->
		<div class="mb-8">
			<a
				href="/settings"
				class="inline-flex items-center gap-1.5 text-sm st-back-link mb-4 transition-colors"
				aria-label="Back to Settings"
			>
				<ArrowLeft class="w-4 h-4" />
				Back to Settings
			</a>
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-2xl font-bold st-title flex items-center gap-2">
						<Package class="w-6 h-6" />
						Modules
					</h1>
					<p class="text-sm st-muted mt-1">
						Manage installed modules in your workspace
					</p>
				</div>
				{#if !isLoading}
					<span class="text-sm st-icon">
						{builtinModules.length + installedModules.length} module{builtinModules.length + installedModules.length !== 1 ? 's' : ''}
					</span>
				{/if}
			</div>
		</div>

		{#if useMockData}
			<div class="flex items-center gap-2 px-4 py-3 mb-6 rounded-lg text-sm" style="background: var(--bos-status-warning-bg); border: 1px solid var(--bos-status-warning); color: var(--bos-status-warning)">
				<AlertCircle class="w-4 h-4 flex-shrink-0" />
				<span>Showing demo data. Module API is not connected yet.</span>
			</div>
		{/if}

		<!-- Loading state -->
		{#if isLoading}
			<div class="flex flex-col items-center justify-center py-16 gap-3 st-icon">
				<Loader2 class="w-8 h-8 animate-spin" />
				<p>Loading modules...</p>
			</div>

		<!-- Error state -->
		{:else if error}
			<div class="flex flex-col items-center justify-center py-16 gap-3" style="color: var(--bos-status-error)">
				<AlertCircle class="w-8 h-8" />
				<p>{error}</p>
				<button
					onclick={() => loadInstalledModules()}
					class="mt-2 btn-pill btn-pill-soft btn-pill-sm"
					aria-label="Retry loading modules"
				>
					Retry
				</button>
			</div>

		<!-- Module list (built-in + custom, draggable) -->
		{:else}
			{@const orderedSlugs = getOrderedSlugs(builtinModules, installedModules)}
			<div class="space-y-1">
				{#each orderedSlugs as slug, i (slug)}
					{@const builtin = builtinModules.find((b) => b.id === slug)}
					{@const customView = installedModules.find((v) => v.installation.module_id === slug)}

					<!-- Drop indicator above -->
					{#if dropIndex === i && dragIndex !== null && dragIndex !== i}
						<div class="st-mod-drop-indicator" aria-hidden="true"></div>
					{/if}

					{#if builtin}
						{@const IconComponent = getBuiltinIcon(builtin.icon)}
						{@const osConfigured = hasOptimalOsConfig(builtin.id)}
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div
							class="flex items-center gap-3 p-4 rounded-lg st-mod-card st-mod-card-clickable transition-colors {dragIndex === i ? 'st-mod-dragging' : ''}"
							draggable="true"
							ondragstart={() => handleDragStart(i)}
							ondragover={(e) => handleDragOver(e, i)}
							ondrop={() => handleDrop(orderedSlugs, i)}
							ondragend={handleDragEnd}
							onclick={() => goto(`/settings/modules/${builtin.id}`)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') goto(`/settings/modules/${builtin.id}`); }}
							role="button"
							tabindex="0"
							aria-label="Configure {builtin.name}"
						>
							<!-- Drag handle -->
							<div class="flex-shrink-0 st-mod-drag-handle" aria-hidden="true">
								<GripVertical class="w-4 h-4" />
							</div>

							<!-- Icon -->
							<div class="flex-shrink-0 w-9 h-9 rounded-lg st-mod-icon-bg flex items-center justify-center">
								<IconComponent class="w-4 h-4 st-mod-icon" />
							</div>

							<!-- Info -->
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2 flex-wrap">
									<span class="font-semibold st-title text-sm">{builtin.name}</span>
									<span class="st-mod-badge-builtin">Built-in</span>
									{#if osConfigured}
										<span class="st-mod-badge-os">OS</span>
									{/if}
								</div>
								<p class="text-xs st-muted mt-0.5 truncate" title={builtin.description}>
									{truncate(builtin.description, 80)}
								</p>
							</div>

							<!-- Status + chevron -->
							<div class="flex-shrink-0 flex items-center gap-2">
								<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium" style="background: var(--bos-status-success-bg); color: var(--bos-status-success);">
									Active
								</span>
								<ChevronRight class="w-4 h-4 st-icon" />
							</div>
						</div>

					{:else if customView}
						{@const mod = customView.module}
						{@const IconComponent = getCategoryIcon(mod?.category ?? 'custom')}
						{@const osConfigured = hasOptimalOsConfig(customView.installation.module_id)}
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div
							class="flex items-center gap-3 p-4 rounded-lg st-mod-card st-mod-card-clickable transition-colors {dragIndex === i ? 'st-mod-dragging' : ''}"
							draggable="true"
							ondragstart={() => handleDragStart(i)}
							ondragover={(e) => handleDragOver(e, i)}
							ondrop={() => handleDrop(orderedSlugs, i)}
							ondragend={handleDragEnd}
							onclick={() => goto(`/settings/modules/${customView.installation.module_id}`)}
							onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') goto(`/settings/modules/${customView.installation.module_id}`); }}
							role="button"
							tabindex="0"
							aria-label="Configure {mod?.name ?? customView.installation.module_id}"
						>
							<!-- Drag handle -->
							<div class="flex-shrink-0 st-mod-drag-handle" aria-hidden="true">
								<GripVertical class="w-4 h-4" />
							</div>

							<!-- Icon -->
							<div class="flex-shrink-0 w-9 h-9 rounded-lg st-mod-icon-bg flex items-center justify-center">
								<IconComponent class="w-4 h-4 st-mod-icon" />
							</div>

							<!-- Info -->
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2 flex-wrap">
									<span class="font-semibold st-title text-sm">
										{mod?.name ?? customView.installation.module_id}
									</span>
									<span class="st-mod-badge-custom">Custom</span>
									{#if osConfigured}
										<span class="st-mod-badge-os">OS</span>
									{/if}
									<span
										class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {getStatusClasses(customView.installation.is_active)}"
										style="{getStatusStyle(customView.installation.is_active)}"
									>
										{getStatusLabel(customView.installation.is_active)}
									</span>
									{#if mod?.version}
										<span class="text-xs st-icon">v{mod.version}</span>
									{/if}
								</div>
								<p class="text-xs st-muted mt-0.5 truncate" title={mod?.description ?? ''}>
									{truncate(mod?.description ?? 'No description available', 80)}
								</p>
							</div>

							<!-- Install date + actions + chevron -->
							<div class="flex-shrink-0 flex items-center gap-2">
								<span class="text-xs st-icon hidden sm:block">
									{relativeTime(customView.installation.installed_at)}
								</span>
								<!-- Stop propagation on action buttons so row click doesn't fire -->
								{#if !customView.installation.is_active && customView.module}
									<button
										onclick={(e) => { e.stopPropagation(); reinstallModule(customView); }}
										class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-xs"
										aria-label="Re-enable {mod?.name ?? 'module'}"
									>
										<Download class="w-4 h-4" />
									</button>
								{/if}
								{#if customView.installation.is_active}
									<button
										onclick={(e) => { e.stopPropagation(); openUninstallDialog(customView); }}
										class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-xs"
										aria-label="Uninstall {mod?.name ?? 'module'}"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								{/if}
								<ChevronRight class="w-4 h-4 st-icon" />
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Uninstall confirmation dialog -->
{#if showUninstallDialog && uninstallTarget}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onkeydown={(e) => { if (e.key === 'Escape' && !isUninstalling) closeUninstallDialog(); }}
	>
		<div
			class="st-mod-dialog rounded-xl shadow-xl max-w-md w-full mx-4 p-6"
			role="dialog"
			aria-modal="true"
			aria-label="Confirm uninstall"
		>
			<h2 class="text-lg font-semibold st-title">
				Uninstall {uninstallTarget.module?.name ?? 'module'}?
			</h2>
			<p class="text-sm st-muted mt-2">
				This will disable the module and remove its routes. Your data tables will be preserved.
			</p>

			{#if uninstallError}
				<div class="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg text-sm" style="background: var(--bos-status-error-bg); color: var(--bos-status-error)">
					<AlertCircle class="w-4 h-4 flex-shrink-0" />
					<span>{uninstallError}</span>
				</div>
			{/if}

			<div class="flex items-center justify-end gap-3 mt-6">
				<button
					onclick={closeUninstallDialog}
					disabled={isUninstalling}
					class="btn-pill btn-pill-soft btn-pill-sm"
					aria-label="Cancel uninstall"
				>
					Cancel
				</button>
				<button
					onclick={confirmUninstall}
					disabled={isUninstalling}
					class="btn-pill btn-pill-danger btn-pill-sm"
					aria-label="Confirm uninstall"
				>
					{#if isUninstalling}
						<Loader2 class="w-4 h-4 animate-spin" />
						Uninstalling...
					{:else}
						<Trash2 class="w-4 h-4" />
						Uninstall
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Install confirmation dialog -->
{#if showInstallDialog && installTarget}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onkeydown={(e) => { if (e.key === 'Escape' && !isInstalling) closeInstallDialog(); }}
	>
		<div
			class="st-mod-dialog rounded-xl shadow-xl max-w-md w-full mx-4 p-6"
			role="dialog"
			aria-modal="true"
			aria-label="Confirm install"
		>
			{#if !installStep}
				<!-- Pre-install: confirmation -->
				<h2 class="text-lg font-semibold st-title">
					Install {installTarget.name}?
				</h2>
				<p class="text-sm st-muted mt-2">
					This will add tables and routes to your workspace.
				</p>
				{#if installTarget.manifest?.actions?.length}
					<p class="text-xs st-icon mt-1">
						{installTarget.manifest.actions.length} action{installTarget.manifest.actions.length !== 1 ? 's' : ''} will be registered.
					</p>
				{/if}

				{#if installError}
					<div class="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg text-sm" style="background: var(--bos-status-error-bg); color: var(--bos-status-error)">
						<AlertCircle class="w-4 h-4 flex-shrink-0" />
						<span>{installError}</span>
					</div>
				{/if}

				<div class="flex items-center justify-end gap-3 mt-6">
					<button
						onclick={closeInstallDialog}
						class="btn-pill btn-pill-soft btn-pill-sm"
						aria-label="Cancel install"
					>
						Cancel
					</button>
					<button
						onclick={confirmInstall}
						disabled={isInstalling}
						class="btn-pill btn-pill-primary btn-pill-sm"
						aria-label="Confirm install"
					>
						<Download class="w-4 h-4" />
						Install
					</button>
				</div>
			{:else}
				<!-- During / after install: step progress -->
				<h2 class="text-lg font-semibold st-title mb-4">
					Installing {installTarget.name}...
				</h2>
				<div class="space-y-3">
					{#each INSTALL_STEPS as step}
						{@const stepIndex = INSTALL_STEPS.findIndex((s) => s.key === step.key)}
						{@const currentIndex = INSTALL_STEPS.findIndex((s) => s.key === installStep)}
						{@const isComplete = stepIndex < currentIndex || installStep === 'done'}
						{@const isCurrent = stepIndex === currentIndex && installStep !== 'done'}
						<div class="flex items-center gap-3">
							{#if isComplete}
								<CircleCheck class="w-5 h-5 flex-shrink-0" style="color: var(--bos-status-success)" />
							{:else if isCurrent}
								<Loader2 class="w-5 h-5 animate-spin flex-shrink-0" style="color: var(--bos-status-info)" />
							{:else}
								<Circle class="w-5 h-5 st-icon flex-shrink-0" />
							{/if}
							<span
								class="text-sm {isCurrent ? 'st-title font-medium' : isComplete ? '' : 'st-icon'}"
								style="{isComplete ? 'color: var(--bos-status-success)' : ''}"
							>
								{step.label}
							</span>
						</div>
					{/each}
				</div>

				{#if installError}
					<div class="mt-4 flex items-center gap-2 px-3 py-2 rounded-lg text-sm" style="background: var(--bos-status-error-bg); color: var(--bos-status-error)">
						<AlertCircle class="w-4 h-4 flex-shrink-0" />
						<span>{installError}</span>
					</div>
					<div class="flex justify-end mt-4">
						<button
							onclick={closeInstallDialog}
							class="btn-pill btn-pill-soft btn-pill-sm"
							aria-label="Close install dialog"
						>
							Close
						</button>
					</div>
				{/if}
			{/if}
		</div>
	</div>
{/if}

<style>
  :global(.st-page-bg) { background: var(--dbg); }
  :global(.st-title) { color: var(--dt); }
  :global(.st-muted) { color: var(--dt3); }
  :global(.st-icon) { color: var(--dt4); }
  :global(.st-back-link) { color: var(--dt3); }
  :global(.st-back-link:hover) { color: var(--dt2); }
  :global(.st-mod-status-disabled) {
    background: var(--dbg3);
    color: var(--dt3);
  }
  :global(.st-mod-empty-icon) {
    background: var(--dbg2);
  }
  :global(.st-mod-card) {
    border: 1px solid var(--dbd);
    background: var(--dbg);
  }
  :global(.st-mod-card:hover) {
    border-color: var(--dt4);
  }
  :global(.st-mod-icon-bg) {
    background: var(--dbg2);
  }
  :global(.st-mod-icon) {
    color: var(--dt2);
  }
  :global(.st-mod-dialog) {
    background: var(--dbg);
    box-shadow: var(--bos-shadow-2, 0 4px 24px rgba(0,0,0,.12));
  }

  /* Clickable row */
  :global(.st-mod-card-clickable) {
    cursor: pointer;
  }
  :global(.st-mod-card-clickable:hover) {
    background: var(--dbg2);
    border-color: var(--dbd2);
  }
  :global(.st-mod-card-clickable:focus-visible) {
    outline: 2px solid var(--dbd2);
    outline-offset: 2px;
  }

  /* Drag handle */
  :global(.st-mod-drag-handle) {
    color: var(--dt4);
    opacity: 0.4;
    cursor: grab;
    transition: opacity 0.15s;
  }
  :global(.st-mod-card-clickable:hover .st-mod-drag-handle) {
    opacity: 0.8;
  }
  :global(.st-mod-drag-handle:active) {
    cursor: grabbing;
  }

  /* Dragging state */
  :global(.st-mod-dragging) {
    opacity: 0.5;
  }

  /* Drop indicator */
  :global(.st-mod-drop-indicator) {
    height: 2px;
    border-radius: 1px;
    background: var(--bos-status-info, #3b82f6);
    margin: 2px 0;
  }

  /* Built-in badge */
  :global(.st-mod-badge-builtin) {
    display: inline-flex;
    align-items: center;
    padding: 0 6px;
    border-radius: 9999px;
    font-size: 0.65rem;
    font-weight: 500;
    background: var(--dbg3);
    color: var(--dt3);
    border: 1px solid var(--dbd);
  }

  /* Custom badge */
  :global(.st-mod-badge-custom) {
    display: inline-flex;
    align-items: center;
    padding: 0 6px;
    border-radius: 9999px;
    font-size: 0.65rem;
    font-weight: 500;
    background: color-mix(in srgb, var(--bos-status-info, #3b82f6) 12%, transparent);
    color: var(--bos-status-info, #3b82f6);
    border: 1px solid color-mix(in srgb, var(--bos-status-info, #3b82f6) 25%, transparent);
  }

  /* OptimalOS "OS" badge */
  :global(.st-mod-badge-os) {
    display: inline-flex;
    align-items: center;
    padding: 0 5px;
    border-radius: 9999px;
    font-size: 0.6rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    background: color-mix(in srgb, #8b5cf6 15%, transparent);
    color: #8b5cf6;
    border: 1px solid color-mix(in srgb, #8b5cf6 30%, transparent);
  }
</style>
