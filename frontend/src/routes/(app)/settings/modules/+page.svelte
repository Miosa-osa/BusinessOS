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
		Circle
	} from 'lucide-svelte';

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

	function getStatusClasses(isActive: boolean): string {
		return isActive
			? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
			: 'st-mod-status-disabled';
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
				console.warn('[Modules] API not implemented yet, showing empty state');
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
						{installedModules.length} module{installedModules.length !== 1 ? 's' : ''} installed
					</span>
				{/if}
			</div>
		</div>

		{#if useMockData}
			<div class="flex items-center gap-2 px-4 py-3 mb-6 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-300 text-sm">
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
			<div class="flex flex-col items-center justify-center py-16 gap-3 text-red-500">
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

		<!-- Empty state -->
		{:else if installedModules.length === 0}
			<div class="flex flex-col items-center justify-center py-20 gap-4 text-center">
				<div class="w-16 h-16 rounded-full st-mod-empty-icon flex items-center justify-center">
					<Package class="w-8 h-8 st-icon" />
				</div>
				<div>
					<h2 class="text-lg font-semibold st-title">No modules installed yet</h2>
					<p class="text-sm st-muted mt-1 max-w-sm">
						Use BUILD mode to create your first app. Modules you install will appear here.
					</p>
				</div>
				<button
					onclick={() => goto('/chat')}
					class="mt-2 btn-pill btn-pill-primary btn-pill-sm"
					aria-label="Open chat to use BUILD mode"
				>
					Open Chat
				</button>
			</div>

		<!-- Module list -->
		{:else}
			<div class="space-y-3">
				{#each installedModules as view (view.installation.id)}
					{@const mod = view.module}
					{@const IconComponent = getCategoryIcon(mod?.category ?? 'custom')}
					<div class="flex items-center gap-4 p-4 rounded-lg st-mod-card transition-colors">
						<!-- Icon -->
						<div class="flex-shrink-0 w-10 h-10 rounded-lg st-mod-icon-bg flex items-center justify-center">
							<IconComponent class="w-5 h-5 st-mod-icon" />
						</div>

						<!-- Info -->
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<span class="font-semibold st-title text-sm">
									{mod?.name ?? view.installation.module_id}
								</span>
								<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {getStatusClasses(view.installation.is_active)}">
									{getStatusLabel(view.installation.is_active)}
								</span>
								{#if mod?.version}
									<span class="text-xs st-icon">v{mod.version}</span>
								{/if}
							</div>
							<p class="text-sm st-muted mt-0.5 truncate" title={mod?.description ?? ''}>
								{truncate(mod?.description ?? 'No description available', 80)}
							</p>
						</div>

						<!-- Install date -->
						<div class="flex-shrink-0 text-right hidden sm:block">
							<span class="text-xs st-icon">
								{relativeTime(view.installation.installed_at)}
							</span>
						</div>

						<!-- Actions -->
						<div class="flex-shrink-0 flex items-center gap-1">
							{#if !view.installation.is_active && view.module}
								<button
									onclick={() => reinstallModule(view)}
									class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-xs"
									aria-label="Re-enable {mod?.name ?? 'module'}"
								>
									<Download class="w-4 h-4" />
								</button>
							{/if}
							{#if view.installation.is_active}
								<button
									onclick={() => openUninstallDialog(view)}
									class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-xs"
									aria-label="Uninstall {mod?.name ?? 'module'}"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							{/if}
						</div>
					</div>
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
				<div class="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
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
					<div class="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
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
								<CircleCheck class="w-5 h-5 text-emerald-500 flex-shrink-0" />
							{:else if isCurrent}
								<Loader2 class="w-5 h-5 text-blue-500 animate-spin flex-shrink-0" />
							{:else}
								<Circle class="w-5 h-5 st-icon flex-shrink-0" />
							{/if}
							<span class="text-sm {isComplete ? 'text-emerald-600 dark:text-emerald-400' : isCurrent ? 'st-title font-medium' : 'st-icon'}">
								{step.label}
							</span>
						</div>
					{/each}
				</div>

				{#if installError}
					<div class="mt-4 flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
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
  :global(.st-page-bg) { background: var(--dbg, var(--bos-card, #fff)); }
  :global(.st-title) { color: var(--dt, var(--bos-text-primary, #111)); }
  :global(.st-muted) { color: var(--dt3, var(--bos-text-tertiary, #888)); }
  :global(.st-icon) { color: var(--dt4, #bbb); }
  :global(.st-back-link) { color: var(--dt3, var(--bos-text-tertiary, #888)); }
  :global(.st-back-link:hover) { color: var(--dt2, var(--bos-text-secondary, #555)); }
  :global(.st-mod-status-disabled) {
    background: var(--dbg3, #eee);
    color: var(--dt3, var(--bos-text-tertiary, #888));
  }
  :global(.st-mod-empty-icon) {
    background: var(--dbg2, var(--bos-bg-secondary, #f5f5f5));
  }
  :global(.st-mod-card) {
    border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
    background: var(--dbg, var(--bos-card, #fff));
  }
  :global(.st-mod-card:hover) {
    border-color: var(--dt4, #bbb);
  }
  :global(.st-mod-icon-bg) {
    background: var(--dbg2, var(--bos-bg-secondary, #f5f5f5));
  }
  :global(.st-mod-icon) {
    color: var(--dt2, var(--bos-text-secondary, #555));
  }
  :global(.st-mod-dialog) {
    background: var(--dbg, var(--bos-card, #fff));
    box-shadow: var(--bos-shadow-2, 0 4px 24px rgba(0,0,0,.12));
  }
</style>
