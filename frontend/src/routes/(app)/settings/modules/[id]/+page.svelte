<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Loader2, AlertCircle } from 'lucide-svelte';

	import { isBuiltinModule, getBuiltinConfig } from '$lib/config/builtinModuleConfigs';
	import type { BuiltinModuleConfig, SchemaField } from '$lib/config/builtinModuleConfigs';
	import { getModule, updateInstallation, getInstallations } from '$lib/api/modules';
	import type { CustomModule, ModuleInstallation } from '$lib/types/modules';
	import { categoryLabels } from '$lib/types/modules';
	import { getCategoryColor } from '$lib/constants/colors';

	import SchemaFormRenderer from '$lib/components/modules/SchemaFormRenderer.svelte';
	import OptimalDataSourceSelector from '$lib/components/modules/OptimalDataSourceSelector.svelte';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import type { ConfigField } from '$lib/api/templates/types';

	// ─── Tab state ─────────────────────────────────────────────────────────────

	type TabId = 'general' | 'configuration' | 'datasources' | 'access' | 'about';
	let activeTab = $state<TabId>('general');

	const TABS: Array<{ id: TabId; label: string }> = [
		{ id: 'general', label: 'General' },
		{ id: 'configuration', label: 'Configuration' },
		{ id: 'datasources', label: 'Data Sources' },
		{ id: 'access', label: 'Access' },
		{ id: 'about', label: 'About' },
	];

	// ─── Module identity ────────────────────────────────────────────────────────

	let moduleId = $derived($page.params.id ?? '');
	let isBuiltin = $derived(isBuiltinModule(moduleId));

	// ─── Loading / error state ──────────────────────────────────────────────────

	let isLoading = $state(true);
	let loadError = $state<string | null>(null);

	// ─── Data ───────────────────────────────────────────────────────────────────

	let builtinConfig = $state<BuiltinModuleConfig | null>(null);
	let customModule = $state<CustomModule | null>(null);
	let installation = $state<ModuleInstallation | null>(null);

	// ─── Derived display fields ─────────────────────────────────────────────────

	let moduleName = $derived(
		builtinConfig?.name ?? customModule?.name ?? moduleId
	);
	let moduleDescription = $derived(
		builtinConfig?.description ?? customModule?.description ?? ''
	);
	let moduleCategory = $derived(
		builtinConfig?.category ?? customModule?.category ?? 'custom'
	);
	let moduleVersion = $derived(
		builtinConfig?.version ?? customModule?.version ?? '1.0.0'
	);
	let categoryColor = $derived(getCategoryColor(moduleCategory));
	let categoryLabel = $derived(
		categoryLabels[moduleCategory as keyof typeof categoryLabels] ?? moduleCategory
	);

	// ─── Config state ───────────────────────────────────────────────────────────

	let configValues = $state<Record<string, string>>({});
	let configDirty = $state(false);
	let configSaving = $state(false);

	// ─── Data source state ──────────────────────────────────────────────────────

	let selectedSources = $state<string[]>([]);

	// ─── Schema bridging ────────────────────────────────────────────────────────
	// SchemaFormRenderer expects Record<string, ConfigField>.
	// builtinModuleConfigs uses SchemaField[]. Convert on the fly.

	function schemaFieldsToRecord(fields: SchemaField[]): Record<string, ConfigField> {
		const out: Record<string, ConfigField> = {};
		for (const f of fields) {
			out[f.key] = {
				type: f.type,
				label: f.label,
				description: f.description,
				default: String(f.default ?? ''),
				required: f.required ?? false,
				options: f.options?.map((o) => o.value),
			};
		}
		return out;
	}

	let schemaRecord = $derived((): Record<string, ConfigField> => {
		if (builtinConfig) {
			return schemaFieldsToRecord(builtinConfig.configSchema);
		}
		if (customModule?.config_schema) {
			// Custom modules already store schema in backend format; cast it.
			return customModule.config_schema as unknown as Record<string, ConfigField>;
		}
		return {};
	});

	// ─── localStorage helpers ───────────────────────────────────────────────────

	function configStorageKey(id: string) {
		return `bos_module_config_${id}`;
	}

	function datasourceStorageKey(id: string) {
		return `bos_module_datasources_${id}`;
	}

	function loadConfigFromStorage(id: string): Record<string, string> {
		try {
			const raw = localStorage.getItem(configStorageKey(id));
			return raw ? (JSON.parse(raw) as Record<string, string>) : {};
		} catch {
			return {};
		}
	}

	function saveConfigToStorage(id: string, values: Record<string, string>) {
		try {
			localStorage.setItem(configStorageKey(id), JSON.stringify(values));
		} catch {
			// localStorage unavailable — silent fail
		}
	}

	function loadSourcesFromStorage(id: string): string[] {
		try {
			const raw = localStorage.getItem(datasourceStorageKey(id));
			return raw ? (JSON.parse(raw) as string[]) : [];
		} catch {
			return [];
		}
	}

	function saveSourcestoStorage(id: string, sources: string[]) {
		try {
			localStorage.setItem(datasourceStorageKey(id), JSON.stringify(sources));
		} catch {
			// silent fail
		}
	}

	// ─── Config change handler ──────────────────────────────────────────────────

	function handleConfigChange(key: string, value: string) {
		configValues = { ...configValues, [key]: value };
		configDirty = true;
		saveConfigToStorage(moduleId, configValues);
	}

	// ─── Reset to defaults ──────────────────────────────────────────────────────

	function resetToDefaults() {
		if (isBuiltin && builtinConfig) {
			const defaults: Record<string, string> = {};
			for (const f of builtinConfig.configSchema) {
				defaults[f.key] = String(f.default ?? '');
			}
			configValues = defaults;
			saveConfigToStorage(moduleId, defaults);
			configDirty = false;
		} else {
			configValues = {};
			saveConfigToStorage(moduleId, {});
			configDirty = false;
		}
	}

	// ─── Save config for custom module via API ──────────────────────────────────

	async function saveCustomConfig() {
		if (!installation) return;
		configSaving = true;
		try {
			await updateInstallation(installation.id, { config: configValues });
			configDirty = false;
		} catch (err) {
			// Non-fatal — config is already in localStorage as fallback
			console.error('[ModuleSettings] Failed to save config via API:', err);
		} finally {
			configSaving = false;
		}
	}

	// ─── Toggle enable/disable ──────────────────────────────────────────────────

	let isEnabled = $state(true);
	let isTogglingEnabled = $state(false);

	async function toggleEnabled() {
		if (isBuiltin) {
			// Built-in modules cannot be disabled from here; noop
			return;
		}
		if (!installation) return;
		isTogglingEnabled = true;
		const next = !isEnabled;
		try {
			await updateInstallation(installation.id, { is_active: next });
			isEnabled = next;
		} catch {
			// Revert on error
		} finally {
			isTogglingEnabled = false;
		}
	}

	// ─── Sidebar toggle (localStorage) ─────────────────────────────────────────

	let showInSidebar = $state(true);

	function toggleSidebar() {
		showInSidebar = !showInSidebar;
		try {
			localStorage.setItem(`bos_module_sidebar_${moduleId}`, String(showInSidebar));
		} catch {
			// silent
		}
	}

	// ─── Data sources change ────────────────────────────────────────────────────

	function handleSourcesChange(sources: string[]) {
		selectedSources = sources;
		saveSourcestoStorage(moduleId, sources);
	}

	// ─── Load ───────────────────────────────────────────────────────────────────

	onMount(async () => {
		isLoading = true;
		loadError = null;

		try {
			if (isBuiltin) {
				builtinConfig = getBuiltinConfig(moduleId);
			} else {
				// Fetch custom module + its installation
				const [mod, instResult] = await Promise.all([
					getModule(moduleId),
					getInstallations().catch(() => ({ installations: [] as ModuleInstallation[] })),
				]);
				customModule = mod;
				installation =
					instResult.installations.find((i) => i.module_id === moduleId) ?? null;
				isEnabled = installation?.is_active ?? mod.is_active;
			}

			// Load persisted config
			configValues = loadConfigFromStorage(moduleId);

			// Load persisted sources
			selectedSources = loadSourcesFromStorage(moduleId);

			// Load sidebar preference
			const sidebarPref = localStorage.getItem(`bos_module_sidebar_${moduleId}`);
			if (sidebarPref !== null) {
				showInSidebar = sidebarPref === 'true';
			}
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Failed to load module';
		} finally {
			isLoading = false;
		}
	});

	function fmtDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric',
		});
	}
</script>

<div class="sms-page">
	<!-- Top bar -->
	<div class="sms-topbar">
		<button
			onclick={() => goto('/settings/modules')}
			class="sms-back"
			aria-label="Back to Modules settings"
		>
			<ArrowLeft class="w-4 h-4" />
			<span>Modules</span>
		</button>

		{#if !isLoading && !loadError}
			<span class="sms-topbar__title">{moduleName}</span>
		{/if}
	</div>

	<!-- Main content -->
	{#if isLoading}
		<div class="sms-center" aria-busy="true" aria-label="Loading module settings">
			<Loader2 class="sms-spinner" />
			<p class="sms-muted">Loading module settings...</p>
		</div>

	{:else if loadError}
		<div class="sms-center" role="alert">
			<div class="sms-error-orb">
				<AlertCircle class="w-6 h-6" />
			</div>
			<p class="sms-text-primary">Failed to load module</p>
			<p class="sms-muted">{loadError}</p>
			<button
				onclick={() => goto('/settings/modules')}
				class="sms-btn-back"
				aria-label="Return to modules list"
			>
				Back to Modules
			</button>
		</div>

	{:else}
		<div class="sms-body">
			<!-- Tab bar -->
			<div class="sms-tab-row" role="tablist">
				{#each TABS as tab (tab.id)}
					<button
						class="sms-tab"
						class:sms-tab--active={activeTab === tab.id}
						role="tab"
						aria-selected={activeTab === tab.id}
						onclick={() => (activeTab = tab.id)}
					>
						{tab.label}
					</button>
				{/each}
			</div>

			<!-- Tab panels -->
			<div class="sms-panel">

				<!-- ── General ─────────────────────────────────────────────────── -->
				{#if activeTab === 'general'}
					<div class="sms-section">
						<h2 class="sms-section__title">General</h2>

						<!-- Module name -->
						<div class="sms-card">
							<div class="sms-row">
								<div class="sms-row__meta">
									<p class="sms-row__label">Module Name</p>
									<p class="sms-row__desc">The display name for this module.</p>
								</div>
								<div class="sms-row__control">
									{#if isBuiltin}
										<span class="sms-value-text">{moduleName}</span>
									{:else}
										<span class="sms-value-text">{moduleName}</span>
									{/if}
								</div>
							</div>

							<!-- Description -->
							<div class="sms-row sms-row--col">
								<div class="sms-row__meta">
									<p class="sms-row__label">Description</p>
								</div>
								<p class="sms-desc-text">{moduleDescription || 'No description.'}</p>
							</div>

							<!-- Category -->
							<div class="sms-row">
								<div class="sms-row__meta">
									<p class="sms-row__label">Category</p>
								</div>
								<div class="sms-row__control">
									<span
										class="sms-badge"
										style="background: color-mix(in srgb, {categoryColor} 12%, transparent); color: {categoryColor}; border-color: color-mix(in srgb, {categoryColor} 25%, transparent);"
									>
										{categoryLabel}
									</span>
								</div>
							</div>

							<!-- Icon -->
							<div class="sms-row">
								<div class="sms-row__meta">
									<p class="sms-row__label">Icon</p>
								</div>
								<div class="sms-row__control">
									<div
										class="sms-icon-preview"
										style="background: color-mix(in srgb, {categoryColor} 15%, transparent);"
										aria-label="Module icon"
									>
										<svg
											viewBox="0 0 24 24"
											fill="none"
											stroke={categoryColor}
											stroke-width="2"
											stroke-linecap="round"
											stroke-linejoin="round"
											class="sms-icon-preview__svg"
											aria-hidden="true"
										>
											<rect x="3" y="3" width="18" height="18" rx="3" />
											<path d="M9 12h6M12 9v6" />
										</svg>
									</div>
								</div>
							</div>
						</div>

						<!-- Toggles card -->
						<div class="sms-card">
							<!-- Enable toggle -->
							<div class="sms-row">
								<div class="sms-row__meta">
									<p class="sms-row__label">Enable Module</p>
									<p class="sms-row__desc">
										{isBuiltin
											? 'Built-in modules are always available.'
											: 'Disable to hide this module without uninstalling it.'}
									</p>
								</div>
								<div class="sms-row__control">
									<button
										role="switch"
										aria-checked={isBuiltin ? true : isEnabled}
										class="sms-toggle {isBuiltin || isEnabled ? 'sms-toggle--on' : ''}"
										disabled={isBuiltin || isTogglingEnabled}
										onclick={toggleEnabled}
										aria-label="Toggle module enabled"
									>
										<span class="sms-toggle__thumb"></span>
									</button>
								</div>
							</div>

							<!-- Sidebar toggle -->
							<div class="sms-row">
								<div class="sms-row__meta">
									<p class="sms-row__label">Show in Sidebar</p>
									<p class="sms-row__desc">Pin this module to the sidebar navigation.</p>
								</div>
								<div class="sms-row__control">
									<button
										role="switch"
										aria-checked={showInSidebar}
										class="sms-toggle {showInSidebar ? 'sms-toggle--on' : ''}"
										onclick={toggleSidebar}
										aria-label="Toggle sidebar visibility"
									>
										<span class="sms-toggle__thumb"></span>
									</button>
								</div>
							</div>
						</div>
					</div>

				<!-- ── Configuration ───────────────────────────────────────────── -->
				{:else if activeTab === 'configuration'}
					<div class="sms-section">
						<div class="sms-section__header">
							<h2 class="sms-section__title">Configuration</h2>
							<div class="sms-section__actions">
								{#if !isBuiltin && configDirty}
									<button
										onclick={saveCustomConfig}
										disabled={configSaving}
										class="sms-btn-primary"
										aria-label="Save configuration"
									>
										{#if configSaving}
											<Loader2 class="w-3.5 h-3.5 animate-spin" />
											Saving...
										{:else}
											Save
										{/if}
									</button>
								{/if}
								<button
									onclick={resetToDefaults}
									class="sms-btn-ghost"
									aria-label="Reset configuration to defaults"
								>
									Reset to defaults
								</button>
							</div>
						</div>

						<div class="sms-card">
							<SchemaFormRenderer
								schema={schemaRecord()}
								values={configValues}
								onchange={handleConfigChange}
								disabled={false}
							/>
						</div>
					</div>

				<!-- ── Data Sources ────────────────────────────────────────────── -->
				{:else if activeTab === 'datasources'}
					<div class="sms-section">
						<h2 class="sms-section__title">Data Sources</h2>
						<p class="sms-section__desc">
							Select which OptimalOS knowledge nodes this module should access for data enrichment.
						</p>
						<div class="sms-card">
							<OptimalDataSourceSelector
								selectedSources={selectedSources}
								onchange={handleSourcesChange}
								disabled={false}
							/>
						</div>
					</div>

				<!-- ── Access ─────────────────────────────────────────────────── -->
				{:else if activeTab === 'access'}
					<div class="sms-section">
						<h2 class="sms-section__title">Access</h2>

						<div class="sms-card">
							{#if isBuiltin}
								<div class="sms-row">
									<div class="sms-row__meta">
										<p class="sms-row__label">Availability</p>
										<p class="sms-row__desc">Who can use this module.</p>
									</div>
									<div class="sms-row__control">
										<span
											class="sms-badge"
											style="background: var(--bos-status-info-bg); color: var(--bos-status-info); border-color: color-mix(in srgb, var(--bos-status-info) 25%, transparent);"
										>
											All workspace members
										</span>
									</div>
								</div>
							{:else if customModule}
								<div class="sms-row">
									<div class="sms-row__meta">
										<p class="sms-row__label">Visibility</p>
										<p class="sms-row__desc">Who can see and install this module.</p>
									</div>
									<div class="sms-row__control">
										<span
											class="sms-badge sms-badge--vis-{customModule.visibility}"
											aria-label="Module visibility: {customModule.visibility}"
										>
											{customModule.visibility}
										</span>
									</div>
								</div>
							{/if}

							<div class="sms-coming-soon">
								<p class="sms-coming-soon__text">
									Workspace-level access control coming soon. Role-based permissions, team
									restrictions, and per-user overrides will be configurable here.
								</p>
							</div>
						</div>
					</div>

				<!-- ── About ──────────────────────────────────────────────────── -->
				{:else if activeTab === 'about'}
					<div class="sms-section">
						<h2 class="sms-section__title">About</h2>

						{#if isBuiltin && builtinConfig}
							<div class="sms-card">
								<div class="sms-detail-list">
									<div class="sms-detail-item">
										<dt>Name</dt>
										<dd>{builtinConfig.name}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Version</dt>
										<dd>{builtinConfig.version}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Type</dt>
										<dd>Built-in module — always available</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Category</dt>
										<dd>{categoryLabel}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Description</dt>
										<dd>{builtinConfig.description}</dd>
									</div>
								</div>
							</div>

						{:else if customModule}
							<div class="sms-card">
								<div class="sms-detail-list">
									<div class="sms-detail-item">
										<dt>Name</dt>
										<dd>{customModule.name}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Version</dt>
										<dd>{customModule.version}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Category</dt>
										<dd>{categoryLabel}</dd>
									</div>
									{#if customModule.creator_name}
										<div class="sms-detail-item">
											<dt>Author</dt>
											<dd>{customModule.creator_name}</dd>
										</div>
									{/if}
									<div class="sms-detail-item">
										<dt>Installs</dt>
										<dd>{customModule.install_count}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Created</dt>
										<dd>{fmtDate(customModule.created_at)}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Updated</dt>
										<dd>{fmtDate(customModule.updated_at)}</dd>
									</div>
									<div class="sms-detail-item">
										<dt>Visibility</dt>
										<dd class="capitalize">{customModule.visibility}</dd>
									</div>
								</div>
							</div>

							<!-- Manifest viewer for custom modules -->
							<div class="sms-section__sub">
								<h3 class="sms-section__subtitle">Manifest</h3>
								<ManifestViewer manifest={customModule.manifest} />
							</div>
						{/if}
					</div>
				{/if}

			</div>
		</div>
	{/if}
</div>

<style>
	/* ════════════════════════════════════════════════════════════════════════════
	   MODULE SETTINGS PAGE — sms- prefix, Foundation tokens
	   Follows md- (module detail) + st- (settings) conventions
	════════════════════════════════════════════════════════════════════════════ */

	.sms-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg);
	}

	/* ── Top bar ─────────────────────────────────────────────────────────────── */

	.sms-topbar {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 12px 24px;
		border-bottom: 1px solid var(--dbd);
		flex-shrink: 0;
	}

	.sms-back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: var(--dt3);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		transition: color 0.15s;
		flex-shrink: 0;
	}

	.sms-back:hover {
		color: var(--dt);
	}

	.sms-topbar__title {
		font-size: 14px;
		font-weight: 600;
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	/* ── Center states ───────────────────────────────────────────────────────── */

	.sms-center {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 10px;
		text-align: center;
		padding: 32px;
	}

	.sms-spinner {
		width: 26px;
		height: 26px;
		color: var(--dt3);
		animation: sms-spin 0.9s linear infinite;
	}

	@keyframes sms-spin {
		to {
			transform: rotate(360deg);
		}
	}

	.sms-error-orb {
		width: 48px;
		height: 48px;
		border-radius: 50%;
		background: var(--bos-status-error-bg);
		color: var(--bos-status-error);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.sms-text-primary {
		font-size: 14px;
		font-weight: 600;
		color: var(--dt);
	}

	.sms-muted {
		font-size: 13px;
		color: var(--dt3);
	}

	.sms-btn-back {
		margin-top: 8px;
		padding: 8px 18px;
		border-radius: 8px;
		border: 1px solid var(--dbd);
		background: var(--dbg2);
		color: var(--dt2);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s;
	}

	.sms-btn-back:hover {
		background: var(--dbg3, var(--dbg2));
		border-color: var(--dt3);
	}

	/* ── Body ────────────────────────────────────────────────────────────────── */

	.sms-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* ── Tab row ─────────────────────────────────────────────────────────────── */

	.sms-tab-row {
		display: flex;
		gap: 0;
		padding: 0 24px;
		border-bottom: 1px solid var(--dbd);
		overflow-x: auto;
		scrollbar-width: none;
		flex-shrink: 0;
	}

	.sms-tab-row::-webkit-scrollbar {
		display: none;
	}

	.sms-tab {
		padding: 10px 16px;
		border: none;
		border-bottom: 2px solid transparent;
		margin-bottom: -1px;
		background: transparent;
		color: var(--dt3);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		white-space: nowrap;
		transition: color 0.15s, border-color 0.15s;
	}

	.sms-tab:hover {
		color: var(--dt2);
	}

	.sms-tab--active {
		color: var(--dt);
		font-weight: 600;
		border-bottom-color: var(--dt);
	}

	/* ── Panel ───────────────────────────────────────────────────────────────── */

	.sms-panel {
		flex: 1;
		overflow-y: auto;
		padding: 24px 24px 48px;
	}

	/* ── Section ─────────────────────────────────────────────────────────────── */

	.sms-section {
		max-width: 640px;
	}

	.sms-section__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 16px;
	}

	.sms-section__title {
		font-size: 13px;
		font-weight: 700;
		color: var(--dt);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		margin-bottom: 16px;
	}

	.sms-section__header .sms-section__title {
		margin-bottom: 0;
	}

	.sms-section__desc {
		font-size: 13px;
		color: var(--dt3);
		margin-bottom: 16px;
		line-height: 1.6;
	}

	.sms-section__actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.sms-section__sub {
		margin-top: 20px;
	}

	.sms-section__subtitle {
		font-size: 12px;
		font-weight: 600;
		color: var(--dt3);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		margin-bottom: 10px;
	}

	/* ── Card ────────────────────────────────────────────────────────────────── */

	.sms-card {
		padding: 0 16px;
		border-radius: 12px;
		border: 1px solid var(--dbd);
		background: var(--dbg);
		margin-bottom: 16px;
	}

	/* ── Row ─────────────────────────────────────────────────────────────────── */

	.sms-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding: 14px 0;
		border-bottom: 1px solid var(--dbd2);
	}

	.sms-row:last-child {
		border-bottom: none;
	}

	.sms-row--col {
		flex-direction: column;
		align-items: flex-start;
		gap: 6px;
	}

	.sms-row__meta {
		flex: 1;
		min-width: 0;
	}

	.sms-row__label {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt);
	}

	.sms-row__desc {
		font-size: 12px;
		color: var(--dt3);
		margin-top: 2px;
		line-height: 1.5;
	}

	.sms-row__control {
		flex-shrink: 0;
	}

	.sms-value-text {
		font-size: 13px;
		color: var(--dt2);
	}

	.sms-desc-text {
		font-size: 13px;
		color: var(--dt2);
		line-height: 1.6;
	}

	/* ── Badge ───────────────────────────────────────────────────────────────── */

	.sms-badge {
		display: inline-flex;
		align-items: center;
		padding: 3px 10px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 600;
		border: 1px solid transparent;
		text-transform: capitalize;
	}

	.sms-badge--vis-public {
		background: var(--bos-status-success-bg);
		color: var(--bos-status-success);
		border-color: color-mix(in srgb, var(--bos-status-success) 25%, transparent);
	}

	.sms-badge--vis-workspace {
		background: var(--bos-status-info-bg);
		color: var(--bos-status-info);
		border-color: color-mix(in srgb, var(--bos-status-info) 25%, transparent);
	}

	.sms-badge--vis-private {
		background: var(--dbg2);
		color: var(--dt3);
		border-color: var(--dbd);
	}

	/* ── Icon preview ────────────────────────────────────────────────────────── */

	.sms-icon-preview {
		width: 38px;
		height: 38px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.sms-icon-preview__svg {
		width: 20px;
		height: 20px;
	}

	/* ── Toggle ──────────────────────────────────────────────────────────────── */

	.sms-toggle {
		position: relative;
		display: inline-flex;
		align-items: center;
		width: 40px;
		height: 22px;
		border-radius: 999px;
		border: none;
		background: var(--dbd);
		cursor: pointer;
		padding: 0;
		flex-shrink: 0;
		transition: background 0.2s;
	}

	.sms-toggle--on {
		background: var(--bos-status-success, #22c55e);
	}

	.sms-toggle:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.sms-toggle__thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--dbg, #fff);
		transition: transform 0.2s;
		pointer-events: none;
	}

	.sms-toggle--on .sms-toggle__thumb {
		transform: translateX(18px);
	}

	/* ── Buttons ─────────────────────────────────────────────────────────────── */

	.sms-btn-primary {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		border-radius: 8px;
		border: none;
		background: var(--dt);
		color: var(--dbg);
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.sms-btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.sms-btn-ghost {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		border-radius: 8px;
		border: 1px solid var(--dbd);
		background: transparent;
		color: var(--dt3);
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		transition: color 0.15s, border-color 0.15s;
	}

	.sms-btn-ghost:hover {
		color: var(--dt2);
		border-color: var(--dt3);
	}

	/* ── Detail list ─────────────────────────────────────────────────────────── */

	.sms-detail-list {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.sms-detail-item {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 16px;
		padding: 11px 0;
		border-bottom: 1px solid var(--dbd2);
	}

	.sms-detail-item:last-child {
		border-bottom: none;
	}

	.sms-detail-item dt {
		font-size: 12px;
		color: var(--dt3);
		flex-shrink: 0;
	}

	.sms-detail-item dd {
		font-size: 12px;
		font-weight: 500;
		color: var(--dt);
		text-align: right;
	}

	.capitalize {
		text-transform: capitalize;
	}

	/* ── Coming soon ─────────────────────────────────────────────────────────── */

	.sms-coming-soon {
		padding: 14px 0;
	}

	.sms-coming-soon__text {
		font-size: 12px;
		color: var(--dt3);
		line-height: 1.6;
	}

	/* ── Dark mode ───────────────────────────────────────────────────────────── */

	:global(.dark) .sms-toggle__thumb {
		background: var(--dbg);
	}
</style>
