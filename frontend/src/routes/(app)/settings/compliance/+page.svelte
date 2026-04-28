<script lang="ts">
  import { onMount } from 'svelte';
  import { getApiBaseUrl } from '$lib/api/base';
  import {
    exportDSAR,
    previewErasure,
    eraseData,
    listLegalHolds,
    placeLegalHold,
    releaseLegalHold,
    listRetentionPolicies,
    createRetentionPolicy,
    deleteRetentionPolicy,
    runRetentionSweep,
    queryAuditEvents,
    AUDIT_EVENT_KINDS,
  } from '$lib/api/compliance';
  import type {
    DSARExportResponse,
    ErasurePreviewResponse,
    ErasureResponse,
    LegalHold,
    RetentionPolicy,
    RetentionScopeType,
    RetentionAction,
    AuditEvent,
    AuditEventKind,
  } from '$lib/api/compliance';

  // ---------------------------------------------------------------------------
  // Tab state
  // ---------------------------------------------------------------------------

  type TabId = 'dsar' | 'erasure' | 'holds' | 'retention' | 'audit';

  let activeTab = $state<TabId>('dsar');

  const tabs: Array<{ id: TabId; label: string }> = [
    { id: 'dsar', label: 'DSAR Export' },
    { id: 'erasure', label: 'Erasure' },
    { id: 'holds', label: 'Legal Holds' },
    { id: 'retention', label: 'Retention' },
    { id: 'audit', label: 'Audit Log' },
  ];

  // ---------------------------------------------------------------------------
  // DSAR state
  // ---------------------------------------------------------------------------

  let dsarPrincipalId = $state('');
  let dsarTenantId = $state('');
  let dsarLoading = $state(false);
  let dsarError = $state('');
  let dsarResult = $state<DSARExportResponse | null>(null);

  async function handleDsarExport() {
    if (!dsarPrincipalId.trim() || !dsarTenantId.trim()) {
      dsarError = 'Both Principal ID and Tenant ID are required.';
      return;
    }
    dsarLoading = true;
    dsarError = '';
    dsarResult = null;
    try {
      dsarResult = await exportDSAR({
        principal_id: dsarPrincipalId.trim(),
        tenant_id: dsarTenantId.trim(),
      });
    } catch (e) {
      dsarError = e instanceof Error ? e.message : 'DSAR export failed.';
    } finally {
      dsarLoading = false;
    }
  }

  // ---------------------------------------------------------------------------
  // Erasure state
  // ---------------------------------------------------------------------------

  type ErasureStep = 'form' | 'preview' | 'confirm' | 'done';

  let erasurePrincipalId = $state('');
  let erasureTenantId = $state('');
  let erasureForce = $state(false);
  let erasureConfirmText = $state('');
  let erasureStep = $state<ErasureStep>('form');
  let erasureLoading = $state(false);
  let erasureError = $state('');
  let erasurePreview = $state<ErasurePreviewResponse | null>(null);
  let erasureResult = $state<ErasureResponse | null>(null);

  let erasureConfirmReady = $derived(erasureConfirmText === 'ERASE');

  async function handleErasurePreview() {
    if (!erasurePrincipalId.trim() || !erasureTenantId.trim()) {
      erasureError = 'Both Principal ID and Tenant ID are required.';
      return;
    }
    erasureLoading = true;
    erasureError = '';
    erasurePreview = null;
    try {
      erasurePreview = await previewErasure({
        principal_id: erasurePrincipalId.trim(),
        tenant_id: erasureTenantId.trim(),
        force: erasureForce,
      });
      erasureStep = 'preview';
    } catch (e) {
      erasureError = e instanceof Error ? e.message : 'Preview failed.';
    } finally {
      erasureLoading = false;
    }
  }

  async function handleErasureExecute() {
    if (!erasureConfirmReady) return;
    erasureLoading = true;
    erasureError = '';
    try {
      erasureResult = await eraseData({
        principal_id: erasurePrincipalId.trim(),
        tenant_id: erasureTenantId.trim(),
        force: erasureForce,
        confirm: true,
      });
      erasureStep = 'done';
    } catch (e) {
      erasureError = e instanceof Error ? e.message : 'Erasure failed.';
    } finally {
      erasureLoading = false;
    }
  }

  function resetErasure() {
    erasurePrincipalId = '';
    erasureTenantId = '';
    erasureForce = false;
    erasureConfirmText = '';
    erasureStep = 'form';
    erasureError = '';
    erasurePreview = null;
    erasureResult = null;
  }

  // ---------------------------------------------------------------------------
  // Legal Holds state
  // ---------------------------------------------------------------------------

  let holdsTenantId = $state('');
  let holdsLoading = $state(false);
  let holdsError = $state('');
  let holds = $state<LegalHold[]>([]);
  let holdsLoaded = $state(false);

  let newHoldSignalId = $state('');
  let newHoldReason = $state('');
  let newHoldLoading = $state(false);
  let newHoldError = $state('');

  let releasingHoldId = $state<string | null>(null);

  async function loadHolds() {
    if (!holdsTenantId.trim()) {
      holdsError = 'Tenant ID is required.';
      return;
    }
    holdsLoading = true;
    holdsError = '';
    try {
      const res = await listLegalHolds(holdsTenantId.trim());
      holds = res.holds ?? [];
      holdsLoaded = true;
    } catch (e) {
      holdsError = e instanceof Error ? e.message : 'Failed to load holds.';
    } finally {
      holdsLoading = false;
    }
  }

  async function handlePlaceHold() {
    if (!holdsTenantId.trim() || !newHoldSignalId.trim() || !newHoldReason.trim()) {
      newHoldError = 'Tenant ID, Signal ID, and Reason are all required.';
      return;
    }
    newHoldLoading = true;
    newHoldError = '';
    try {
      const hold = await placeLegalHold({
        tenant_id: holdsTenantId.trim(),
        signal_id: newHoldSignalId.trim(),
        reason: newHoldReason.trim(),
      });
      holds = [hold, ...holds];
      newHoldSignalId = '';
      newHoldReason = '';
    } catch (e) {
      newHoldError = e instanceof Error ? e.message : 'Failed to place hold.';
    } finally {
      newHoldLoading = false;
    }
  }

  async function handleReleaseHold(holdId: string) {
    releasingHoldId = holdId;
    try {
      await releaseLegalHold(holdId);
      holds = holds.filter((h) => h.id !== holdId);
    } catch (e) {
      holdsError = e instanceof Error ? e.message : 'Failed to release hold.';
    } finally {
      releasingHoldId = null;
    }
  }

  // ---------------------------------------------------------------------------
  // Retention Policies state
  // ---------------------------------------------------------------------------

  let retentionTenantId = $state('');
  let retentionLoading = $state(false);
  let retentionError = $state('');
  let policies = $state<RetentionPolicy[]>([]);
  let policiesLoaded = $state(false);

  let newPolicyScopeType = $state<RetentionScopeType>('tenant');
  let newPolicyScopeValue = $state('');
  let newPolicyTtlDays = $state(90);
  let newPolicyAction = $state<RetentionAction>('archive');
  let newPolicyLoading = $state(false);
  let newPolicyError = $state('');

  let deletingPolicyId = $state<string | null>(null);

  let sweepDryRun = $state(true);
  let sweepLoading = $state(false);
  let sweepError = $state('');
  let sweepResult = $state<import('$lib/api/compliance').RetentionSweepResponse | null>(null);

  async function loadPolicies() {
    if (!retentionTenantId.trim()) {
      retentionError = 'Tenant ID is required.';
      return;
    }
    retentionLoading = true;
    retentionError = '';
    try {
      const res = await listRetentionPolicies(retentionTenantId.trim());
      policies = res.policies ?? [];
      policiesLoaded = true;
    } catch (e) {
      retentionError = e instanceof Error ? e.message : 'Failed to load policies.';
    } finally {
      retentionLoading = false;
    }
  }

  async function handleCreatePolicy() {
    if (!retentionTenantId.trim() || !newPolicyScopeValue.trim() || newPolicyTtlDays < 1) {
      newPolicyError = 'Tenant ID, Scope Value, and TTL (≥ 1 day) are required.';
      return;
    }
    newPolicyLoading = true;
    newPolicyError = '';
    try {
      const policy = await createRetentionPolicy({
        tenant_id: retentionTenantId.trim(),
        scope_type: newPolicyScopeType,
        scope_value: newPolicyScopeValue.trim(),
        ttl_days: newPolicyTtlDays,
        action: newPolicyAction,
      });
      policies = [policy, ...policies];
      newPolicyScopeValue = '';
      newPolicyTtlDays = 90;
    } catch (e) {
      newPolicyError = e instanceof Error ? e.message : 'Failed to create policy.';
    } finally {
      newPolicyLoading = false;
    }
  }

  async function handleDeletePolicy(policyId: string) {
    deletingPolicyId = policyId;
    try {
      await deleteRetentionPolicy(policyId);
      policies = policies.filter((p) => p.id !== policyId);
    } catch (e) {
      retentionError = e instanceof Error ? e.message : 'Failed to delete policy.';
    } finally {
      deletingPolicyId = null;
    }
  }

  async function handleSweep() {
    if (!retentionTenantId.trim()) {
      sweepError = 'Tenant ID is required to run a sweep.';
      return;
    }
    sweepLoading = true;
    sweepError = '';
    sweepResult = null;
    try {
      sweepResult = await runRetentionSweep({
        tenant_id: retentionTenantId.trim(),
        dry_run: sweepDryRun,
      });
    } catch (e) {
      sweepError = e instanceof Error ? e.message : 'Sweep failed.';
    } finally {
      sweepLoading = false;
    }
  }

  // ---------------------------------------------------------------------------
  // Audit Log state
  // ---------------------------------------------------------------------------

  let auditTenantId = $state('');
  let auditPrincipal = $state('');
  let auditKind = $state<AuditEventKind | ''>('');
  let auditSince = $state('');
  let auditUntil = $state('');
  let auditLimit = $state(100);
  let auditLoading = $state(false);
  let auditError = $state('');
  let auditEvents = $state<AuditEvent[]>([]);
  let auditLoaded = $state(false);
  let auditExpandedId = $state<string | null>(null);

  async function handleAuditQuery() {
    if (!auditTenantId.trim()) {
      auditError = 'Tenant ID is required.';
      return;
    }
    auditLoading = true;
    auditError = '';
    auditEvents = [];
    try {
      const res = await queryAuditEvents({
        tenant_id: auditTenantId.trim(),
        principal: auditPrincipal.trim() || undefined,
        kind: auditKind || undefined,
        since: auditSince || undefined,
        until: auditUntil || undefined,
        limit: auditLimit,
      });
      auditEvents = res.events ?? [];
      auditLoaded = true;
    } catch (e) {
      auditError = e instanceof Error ? e.message : 'Query failed.';
    } finally {
      auditLoading = false;
    }
  }

  function toggleAuditExpand(id: string) {
    auditExpandedId = auditExpandedId === id ? null : id;
  }

  function formatDate(iso: string): string {
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  }
</script>

<div class="h-full overflow-y-auto cp-page-bg">
  <div class="max-w-3xl mx-auto px-6 py-8">

    <!-- Page header -->
    <div class="mb-8">
      <a
        href="/settings"
        class="inline-flex items-center gap-1.5 text-sm cp-back-link mb-4 transition-colors"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back to Settings
      </a>
      <h1 class="text-2xl font-bold cp-title">Compliance</h1>
      <p class="text-sm cp-muted mt-1">
        DSAR exports, erasure, legal holds, retention policies, and audit log
      </p>
    </div>

    <!-- Tab Navigation -->
    <div class="cp-tab-row mb-6">
      {#each tabs as tab}
        <button
          onclick={() => (activeTab = tab.id)}
          class="cp-tab"
          class:cp-tab--active={activeTab === tab.id}
          aria-selected={activeTab === tab.id}
        >
          {tab.label}
        </button>
      {/each}
    </div>

    <!-- ======================================================================
         DSAR EXPORT TAB
         ====================================================================== -->
    {#if activeTab === 'dsar'}
      <section>
        <div class="cp-section-header mb-4">
          <h2 class="cp-section-title">Data Subject Access Request</h2>
          <p class="cp-section-desc">
            Export all data held for a principal. Returns a structured counts summary and the full
            data payload as JSON.
          </p>
        </div>

        <!-- Form card -->
        <div class="cp-card mb-4">
          <div class="cp-field-row">
            <div class="cp-field">
              <label for="dsar-principal" class="cp-label">Principal ID</label>
              <input
                id="dsar-principal"
                type="text"
                bind:value={dsarPrincipalId}
                placeholder="user_abc123"
                class="cp-input"
                aria-label="Principal ID for DSAR export"
              />
            </div>
            <div class="cp-field">
              <label for="dsar-tenant" class="cp-label">Tenant ID</label>
              <input
                id="dsar-tenant"
                type="text"
                bind:value={dsarTenantId}
                placeholder="tenant_xyz"
                class="cp-input"
                aria-label="Tenant ID for DSAR export"
              />
            </div>
          </div>

          {#if dsarError}
            <div class="cp-error mt-3" role="alert">{dsarError}</div>
          {/if}

          <div class="mt-4">
            <button
              onclick={handleDsarExport}
              disabled={dsarLoading}
              class="btn-pill btn-pill-primary btn-pill-sm"
              aria-label="Export DSAR data"
            >
              {#if dsarLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Exporting...
              {:else}
                Export Data
              {/if}
            </button>
          </div>
        </div>

        <!-- Results -->
        {#if dsarResult}
          <div class="cp-card">
            <div class="cp-result-header mb-3">
              <h3 class="cp-result-title">Export Result</h3>
              <span class="cp-badge cp-badge-success">Success</span>
            </div>
            <div class="cp-meta-row mb-4">
              <span class="cp-meta-label">Principal</span>
              <span class="cp-meta-value">{dsarResult.principal_id}</span>
              <span class="cp-meta-label">Tenant</span>
              <span class="cp-meta-value">{dsarResult.tenant_id}</span>
              <span class="cp-meta-label">Exported at</span>
              <span class="cp-meta-value">{formatDate(dsarResult.exported_at)}</span>
            </div>

            <!-- Counts table -->
            <h4 class="cp-subsection-label mb-2">Record Counts</h4>
            <div class="cp-table-wrap mb-4">
              <table class="cp-table" aria-label="DSAR export record counts">
                <thead>
                  <tr>
                    <th class="cp-th">Resource</th>
                    <th class="cp-th cp-th-right">Count</th>
                  </tr>
                </thead>
                <tbody>
                  {#each Object.entries(dsarResult.counts) as [resource, count]}
                    <tr class="cp-tr">
                      <td class="cp-td cp-td-mono">{resource}</td>
                      <td class="cp-td cp-td-right cp-td-num">{count}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            <!-- Raw JSON -->
            <h4 class="cp-subsection-label mb-2">Raw Payload</h4>
            <pre class="cp-json" aria-label="Raw DSAR export JSON">{JSON.stringify(dsarResult.data, null, 2)}</pre>
          </div>
        {/if}
      </section>
    {/if}

    <!-- ======================================================================
         ERASURE TAB
         ====================================================================== -->
    {#if activeTab === 'erasure'}
      <section>
        <div class="cp-section-header mb-4">
          <h2 class="cp-section-title">Data Erasure</h2>
          <p class="cp-section-desc">
            Permanently erase all data for a principal. This action is irreversible. Preview first
            to see what will be affected.
          </p>
        </div>

        <!-- Warning banner -->
        <div class="cp-danger-banner mb-5" role="alert">
          <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          </svg>
          <div>
            <p class="font-semibold">Irreversible operation</p>
            <p class="text-sm mt-0.5">
              Data erasure permanently removes all records for the specified principal. Legal holds
              will block erasure unless <strong>Force</strong> is enabled.
              Always preview before executing.
            </p>
          </div>
        </div>

        {#if erasureStep === 'done' && erasureResult}
          <!-- Done state -->
          <div class="cp-card">
            <div class="cp-result-header mb-3">
              <h3 class="cp-result-title">Erasure Complete</h3>
              <span class="cp-badge cp-badge-success">Erased</span>
            </div>
            <div class="cp-meta-row mb-4">
              <span class="cp-meta-label">Principal</span>
              <span class="cp-meta-value">{erasureResult.principal_id}</span>
              <span class="cp-meta-label">Tenant</span>
              <span class="cp-meta-value">{erasureResult.tenant_id}</span>
              <span class="cp-meta-label">Erased at</span>
              <span class="cp-meta-value">{formatDate(erasureResult.erased_at)}</span>
            </div>
            <h4 class="cp-subsection-label mb-2">Erased Record Counts</h4>
            <div class="cp-table-wrap mb-4">
              <table class="cp-table" aria-label="Erased record counts">
                <thead>
                  <tr>
                    <th class="cp-th">Resource</th>
                    <th class="cp-th cp-th-right">Count</th>
                  </tr>
                </thead>
                <tbody>
                  {#each Object.entries(erasureResult.counts) as [resource, count]}
                    <tr class="cp-tr">
                      <td class="cp-td cp-td-mono">{resource}</td>
                      <td class="cp-td cp-td-right cp-td-num">{count}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            <button onclick={resetErasure} class="btn-pill btn-pill-secondary btn-pill-sm">
              Start New Erasure
            </button>
          </div>

        {:else if erasureStep === 'preview' && erasurePreview}
          <!-- Preview state -->
          <div class="cp-card">
            <div class="cp-result-header mb-3">
              <h3 class="cp-result-title">Erasure Preview</h3>
              {#if erasurePreview.can_proceed}
                <span class="cp-badge cp-badge-warning">Ready to erase</span>
              {:else}
                <span class="cp-badge cp-badge-error">Blocked</span>
              {/if}
            </div>

            {#if erasurePreview.blocked_by_holds && erasurePreview.blocked_by_holds.length > 0}
              <div class="cp-error mb-4" role="alert">
                Blocked by {erasurePreview.blocked_by_holds.length} legal hold(s):
                {erasurePreview.blocked_by_holds.join(', ')}
              </div>
            {/if}

            <h4 class="cp-subsection-label mb-2">Records that will be erased</h4>
            <div class="cp-table-wrap mb-5">
              <table class="cp-table" aria-label="Records to be erased">
                <thead>
                  <tr>
                    <th class="cp-th">Resource</th>
                    <th class="cp-th cp-th-right">Count</th>
                  </tr>
                </thead>
                <tbody>
                  {#each Object.entries(erasurePreview.counts) as [resource, count]}
                    <tr class="cp-tr">
                      <td class="cp-td cp-td-mono">{resource}</td>
                      <td class="cp-td cp-td-right cp-td-num">{count}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            {#if erasurePreview.can_proceed}
              <div class="cp-confirm-zone">
                <p class="text-sm font-medium" style="color: var(--bos-status-error)">
                  To confirm permanent erasure, type <strong>ERASE</strong> in the box below:
                </p>
                <input
                  type="text"
                  bind:value={erasureConfirmText}
                  placeholder="Type ERASE to confirm"
                  class="cp-input cp-input-danger mt-3 max-w-xs"
                  aria-label="Erasure confirmation field — type ERASE"
                />
                {#if erasureError}
                  <div class="cp-error mt-2" role="alert">{erasureError}</div>
                {/if}
                <div class="flex gap-3 mt-4">
                  <button
                    onclick={handleErasureExecute}
                    disabled={!erasureConfirmReady || erasureLoading}
                    class="btn-pill btn-pill-danger btn-pill-sm"
                    aria-label="Confirm and execute erasure"
                  >
                    {#if erasureLoading}
                      <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Erasing...
                    {:else}
                      Permanently Erase Data
                    {/if}
                  </button>
                  <button
                    onclick={resetErasure}
                    class="btn-pill btn-pill-soft btn-pill-sm"
                    aria-label="Cancel erasure"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            {:else}
              <div class="flex gap-3">
                <button
                  onclick={resetErasure}
                  class="btn-pill btn-pill-secondary btn-pill-sm"
                >
                  Back
                </button>
              </div>
            {/if}
          </div>

        {:else}
          <!-- Form state -->
          <div class="cp-card">
            <div class="cp-field-row mb-4">
              <div class="cp-field">
                <label for="erasure-principal" class="cp-label">Principal ID</label>
                <input
                  id="erasure-principal"
                  type="text"
                  bind:value={erasurePrincipalId}
                  placeholder="user_abc123"
                  class="cp-input"
                  aria-label="Principal ID for erasure"
                />
              </div>
              <div class="cp-field">
                <label for="erasure-tenant" class="cp-label">Tenant ID</label>
                <input
                  id="erasure-tenant"
                  type="text"
                  bind:value={erasureTenantId}
                  placeholder="tenant_xyz"
                  class="cp-input"
                  aria-label="Tenant ID for erasure"
                />
              </div>
            </div>

            <label class="cp-checkbox-row" aria-label="Force erasure ignoring holds">
              <input type="checkbox" bind:checked={erasureForce} class="cp-checkbox" />
              <span class="text-sm cp-muted">
                Force erasure (bypasses legal holds — use with caution)
              </span>
            </label>

            {#if erasureError}
              <div class="cp-error mt-3" role="alert">{erasureError}</div>
            {/if}

            <div class="mt-4">
              <button
                onclick={handleErasurePreview}
                disabled={erasureLoading}
                class="btn-pill btn-pill-warning btn-pill-sm"
                aria-label="Preview erasure impact"
              >
                {#if erasureLoading}
                  <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Loading preview...
                {:else}
                  Preview Erasure Impact
                {/if}
              </button>
            </div>
          </div>
        {/if}
      </section>
    {/if}

    <!-- ======================================================================
         LEGAL HOLDS TAB
         ====================================================================== -->
    {#if activeTab === 'holds'}
      <section>
        <div class="cp-section-header mb-4">
          <h2 class="cp-section-title">Legal Holds</h2>
          <p class="cp-section-desc">
            Place and manage legal holds on signals. Active holds block erasure unless forced.
          </p>
        </div>

        <!-- Tenant ID + load -->
        <div class="cp-card mb-4">
          <div class="flex gap-3 items-end">
            <div class="cp-field flex-1">
              <label for="holds-tenant" class="cp-label">Tenant ID</label>
              <input
                id="holds-tenant"
                type="text"
                bind:value={holdsTenantId}
                placeholder="tenant_xyz"
                class="cp-input"
                aria-label="Tenant ID for legal holds"
              />
            </div>
            <button
              onclick={loadHolds}
              disabled={holdsLoading}
              class="btn-pill btn-pill-secondary btn-pill-sm"
              aria-label="Load legal holds for this tenant"
            >
              {#if holdsLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Loading...
              {:else}
                Load Holds
              {/if}
            </button>
          </div>
          {#if holdsError}
            <div class="cp-error mt-3" role="alert">{holdsError}</div>
          {/if}
        </div>

        <!-- Active holds table -->
        {#if holdsLoaded}
          <div class="cp-card mb-4">
            <h3 class="cp-subsection-label mb-3">Active Holds ({holds.length})</h3>
            {#if holds.length === 0}
              <p class="text-sm cp-muted">No active legal holds for this tenant.</p>
            {:else}
              <div class="cp-table-wrap">
                <table class="cp-table" aria-label="Active legal holds">
                  <thead>
                    <tr>
                      <th class="cp-th">Signal ID</th>
                      <th class="cp-th">Reason</th>
                      <th class="cp-th">Placed At</th>
                      <th class="cp-th cp-th-right">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each holds as hold (hold.id)}
                      <tr class="cp-tr">
                        <td class="cp-td cp-td-mono">{hold.signal_id}</td>
                        <td class="cp-td">{hold.reason}</td>
                        <td class="cp-td cp-td-dim">{formatDate(hold.placed_at)}</td>
                        <td class="cp-td cp-td-right">
                          <button
                            onclick={() => handleReleaseHold(hold.id)}
                            disabled={releasingHoldId === hold.id}
                            class="btn-pill btn-pill-ghost btn-pill-xs"
                            aria-label="Release legal hold on signal {hold.signal_id}"
                          >
                            {releasingHoldId === hold.id ? 'Releasing...' : 'Release'}
                          </button>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        {/if}

        <!-- Place new hold form -->
        <div class="cp-card">
          <h3 class="cp-subsection-label mb-3">Place New Hold</h3>
          <div class="space-y-3">
            <div class="cp-field">
              <label for="hold-signal" class="cp-label">Signal ID</label>
              <input
                id="hold-signal"
                type="text"
                bind:value={newHoldSignalId}
                placeholder="signal_abc123"
                class="cp-input"
                aria-label="Signal ID for new legal hold"
              />
            </div>
            <div class="cp-field">
              <label for="hold-reason" class="cp-label">Reason</label>
              <textarea
                id="hold-reason"
                bind:value={newHoldReason}
                placeholder="Describe the legal or compliance reason for this hold"
                rows="3"
                class="cp-input cp-textarea"
                aria-label="Reason for new legal hold"
              ></textarea>
            </div>
          </div>
          {#if newHoldError}
            <div class="cp-error mt-3" role="alert">{newHoldError}</div>
          {/if}
          <div class="mt-4">
            <button
              onclick={handlePlaceHold}
              disabled={newHoldLoading}
              class="btn-pill btn-pill-primary btn-pill-sm"
              aria-label="Place new legal hold"
            >
              {#if newHoldLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Placing...
              {:else}
                Place Hold
              {/if}
            </button>
          </div>
        </div>
      </section>
    {/if}

    <!-- ======================================================================
         RETENTION POLICIES TAB
         ====================================================================== -->
    {#if activeTab === 'retention'}
      <section>
        <div class="cp-section-header mb-4">
          <h2 class="cp-section-title">Retention Policies</h2>
          <p class="cp-section-desc">
            Define how long data is kept before it is archived, deleted, or redacted. Run sweeps
            to apply policies — dry-run by default.
          </p>
        </div>

        <!-- Tenant ID + load -->
        <div class="cp-card mb-4">
          <div class="flex gap-3 items-end">
            <div class="cp-field flex-1">
              <label for="retention-tenant" class="cp-label">Tenant ID</label>
              <input
                id="retention-tenant"
                type="text"
                bind:value={retentionTenantId}
                placeholder="tenant_xyz"
                class="cp-input"
                aria-label="Tenant ID for retention policies"
              />
            </div>
            <button
              onclick={loadPolicies}
              disabled={retentionLoading}
              class="btn-pill btn-pill-secondary btn-pill-sm"
              aria-label="Load retention policies for this tenant"
            >
              {#if retentionLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Loading...
              {:else}
                Load Policies
              {/if}
            </button>
          </div>
          {#if retentionError}
            <div class="cp-error mt-3" role="alert">{retentionError}</div>
          {/if}
        </div>

        <!-- Policies table -->
        {#if policiesLoaded}
          <div class="cp-card mb-4">
            <h3 class="cp-subsection-label mb-3">Policies ({policies.length})</h3>
            {#if policies.length === 0}
              <p class="text-sm cp-muted">No retention policies defined for this tenant.</p>
            {:else}
              <div class="cp-table-wrap">
                <table class="cp-table" aria-label="Retention policies">
                  <thead>
                    <tr>
                      <th class="cp-th">Scope Type</th>
                      <th class="cp-th">Scope Value</th>
                      <th class="cp-th cp-th-right">TTL (days)</th>
                      <th class="cp-th">Action</th>
                      <th class="cp-th cp-th-right">Delete</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each policies as policy (policy.id)}
                      <tr class="cp-tr">
                        <td class="cp-td"><span class="cp-scope-badge">{policy.scope_type}</span></td>
                        <td class="cp-td cp-td-mono">{policy.scope_value}</td>
                        <td class="cp-td cp-td-right cp-td-num">{policy.ttl_days}</td>
                        <td class="cp-td"><span class="cp-action-badge cp-action-badge--{policy.action}">{policy.action}</span></td>
                        <td class="cp-td cp-td-right">
                          <button
                            onclick={() => handleDeletePolicy(policy.id)}
                            disabled={deletingPolicyId === policy.id}
                            class="btn-pill btn-pill-ghost btn-pill-xs"
                            aria-label="Delete retention policy for {policy.scope_value}"
                          >
                            {deletingPolicyId === policy.id ? 'Deleting...' : 'Delete'}
                          </button>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>

          <!-- Sweep -->
          <div class="cp-card mb-4">
            <h3 class="cp-subsection-label mb-3">Run Retention Sweep</h3>
            <label class="cp-checkbox-row mb-4" aria-label="Dry run toggle">
              <input type="checkbox" bind:checked={sweepDryRun} class="cp-checkbox" />
              <span class="text-sm cp-muted">
                Dry run — preview affected records without making changes
              </span>
            </label>
            {#if !sweepDryRun}
              <div class="cp-danger-banner mb-4" role="alert">
                <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                </svg>
                <span class="text-sm">Live sweep — this will modify or delete real data.</span>
              </div>
            {/if}
            {#if sweepError}
              <div class="cp-error mb-3" role="alert">{sweepError}</div>
            {/if}
            <button
              onclick={handleSweep}
              disabled={sweepLoading}
              class="btn-pill btn-pill-sm {sweepDryRun ? 'btn-pill-secondary' : 'btn-pill-warning'}"
              aria-label="Run retention sweep"
            >
              {#if sweepLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Running sweep...
              {:else}
                {sweepDryRun ? 'Run Dry Sweep' : 'Run Live Sweep'}
              {/if}
            </button>

            {#if sweepResult}
              <div class="mt-4">
                <div class="cp-result-header mb-2">
                  <span class="cp-subsection-label">Sweep Result</span>
                  <span class="cp-badge {sweepResult.dry_run ? 'cp-badge-info' : 'cp-badge-success'}">
                    {sweepResult.dry_run ? 'Dry run' : 'Applied'}
                  </span>
                </div>
                <p class="text-sm cp-muted mb-3">
                  {sweepResult.affected} record(s) affected · swept at {formatDate(sweepResult.swept_at)}
                </p>
                {#if Object.keys(sweepResult.counts).length > 0}
                  <div class="cp-table-wrap">
                    <table class="cp-table" aria-label="Sweep result counts">
                      <thead>
                        <tr>
                          <th class="cp-th">Resource</th>
                          <th class="cp-th cp-th-right">Count</th>
                        </tr>
                      </thead>
                      <tbody>
                        {#each Object.entries(sweepResult.counts) as [resource, count]}
                          <tr class="cp-tr">
                            <td class="cp-td cp-td-mono">{resource}</td>
                            <td class="cp-td cp-td-right cp-td-num">{count}</td>
                          </tr>
                        {/each}
                      </tbody>
                    </table>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/if}

        <!-- Add policy form -->
        <div class="cp-card">
          <h3 class="cp-subsection-label mb-3">Add Retention Policy</h3>
          <div class="space-y-3">
            <div class="cp-field-row">
              <div class="cp-field">
                <label for="policy-scope-type" class="cp-label">Scope Type</label>
                <select
                  id="policy-scope-type"
                  bind:value={newPolicyScopeType}
                  class="cp-select"
                  aria-label="Retention policy scope type"
                >
                  <option value="node">node</option>
                  <option value="genre">genre</option>
                  <option value="tenant">tenant</option>
                  <option value="global">global</option>
                </select>
              </div>
              <div class="cp-field">
                <label for="policy-scope-value" class="cp-label">Scope Value</label>
                <input
                  id="policy-scope-value"
                  type="text"
                  bind:value={newPolicyScopeValue}
                  placeholder="e.g. task, email, tenant_xyz"
                  class="cp-input"
                  aria-label="Retention policy scope value"
                />
              </div>
            </div>
            <div class="cp-field-row">
              <div class="cp-field">
                <label for="policy-ttl" class="cp-label">TTL (days)</label>
                <input
                  id="policy-ttl"
                  type="number"
                  bind:value={newPolicyTtlDays}
                  min="1"
                  class="cp-input"
                  aria-label="Retention policy TTL in days"
                />
              </div>
              <div class="cp-field">
                <label for="policy-action" class="cp-label">Action</label>
                <select
                  id="policy-action"
                  bind:value={newPolicyAction}
                  class="cp-select"
                  aria-label="Retention policy action"
                >
                  <option value="archive">archive</option>
                  <option value="delete">delete</option>
                  <option value="redact">redact</option>
                </select>
              </div>
            </div>
          </div>
          {#if newPolicyError}
            <div class="cp-error mt-3" role="alert">{newPolicyError}</div>
          {/if}
          <div class="mt-4">
            <button
              onclick={handleCreatePolicy}
              disabled={newPolicyLoading}
              class="btn-pill btn-pill-primary btn-pill-sm"
              aria-label="Create retention policy"
            >
              {#if newPolicyLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Creating...
              {:else}
                Add Policy
              {/if}
            </button>
          </div>
        </div>
      </section>
    {/if}

    <!-- ======================================================================
         AUDIT LOG TAB
         ====================================================================== -->
    {#if activeTab === 'audit'}
      <section>
        <div class="cp-section-header mb-4">
          <h2 class="cp-section-title">Audit Log</h2>
          <p class="cp-section-desc">
            Query compliance audit events. Maximum 500 results per query.
          </p>
        </div>

        <!-- Filter form -->
        <div class="cp-card mb-4">
          <div class="space-y-3">
            <div class="cp-field-row">
              <div class="cp-field">
                <label for="audit-tenant" class="cp-label">Tenant ID <span class="cp-required">*</span></label>
                <input
                  id="audit-tenant"
                  type="text"
                  bind:value={auditTenantId}
                  placeholder="tenant_xyz"
                  class="cp-input"
                  aria-label="Tenant ID for audit query"
                />
              </div>
              <div class="cp-field">
                <label for="audit-principal" class="cp-label">Principal</label>
                <input
                  id="audit-principal"
                  type="text"
                  bind:value={auditPrincipal}
                  placeholder="user_abc123 (optional)"
                  class="cp-input"
                  aria-label="Principal filter for audit query"
                />
              </div>
            </div>
            <div class="cp-field-row">
              <div class="cp-field">
                <label for="audit-kind" class="cp-label">Event Kind</label>
                <select
                  id="audit-kind"
                  bind:value={auditKind}
                  class="cp-select"
                  aria-label="Event kind filter for audit query"
                >
                  <option value="">All kinds</option>
                  {#each AUDIT_EVENT_KINDS as kind}
                    <option value={kind}>{kind}</option>
                  {/each}
                </select>
              </div>
              <div class="cp-field">
                <label for="audit-limit" class="cp-label">Limit</label>
                <input
                  id="audit-limit"
                  type="number"
                  bind:value={auditLimit}
                  min="1"
                  max="500"
                  class="cp-input"
                  aria-label="Result limit for audit query"
                />
              </div>
            </div>
            <div class="cp-field-row">
              <div class="cp-field">
                <label for="audit-since" class="cp-label">Since</label>
                <input
                  id="audit-since"
                  type="datetime-local"
                  bind:value={auditSince}
                  class="cp-input"
                  aria-label="Start datetime for audit query"
                />
              </div>
              <div class="cp-field">
                <label for="audit-until" class="cp-label">Until</label>
                <input
                  id="audit-until"
                  type="datetime-local"
                  bind:value={auditUntil}
                  class="cp-input"
                  aria-label="End datetime for audit query"
                />
              </div>
            </div>
          </div>

          {#if auditError}
            <div class="cp-error mt-3" role="alert">{auditError}</div>
          {/if}

          <div class="mt-4">
            <button
              onclick={handleAuditQuery}
              disabled={auditLoading}
              class="btn-pill btn-pill-primary btn-pill-sm"
              aria-label="Run audit log query"
            >
              {#if auditLoading}
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Querying...
              {:else}
                Query Audit Log
              {/if}
            </button>
          </div>
        </div>

        <!-- Results table -->
        {#if auditLoaded}
          <div class="cp-card">
            <div class="cp-result-header mb-3">
              <span class="cp-subsection-label">Results ({auditEvents.length})</span>
            </div>
            {#if auditEvents.length === 0}
              <p class="text-sm cp-muted">No events found matching your filters.</p>
            {:else}
              <div class="cp-table-wrap">
                <table class="cp-table" aria-label="Audit log events">
                  <thead>
                    <tr>
                      <th class="cp-th">Occurred At</th>
                      <th class="cp-th">Kind</th>
                      <th class="cp-th">Principal</th>
                      <th class="cp-th cp-th-right">Metadata</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each auditEvents as event (event.id)}
                      <tr class="cp-tr cp-tr-clickable" onclick={() => toggleAuditExpand(event.id)}>
                        <td class="cp-td cp-td-dim">{formatDate(event.occurred_at)}</td>
                        <td class="cp-td"><span class="cp-kind-badge">{event.kind}</span></td>
                        <td class="cp-td cp-td-mono">{event.principal}</td>
                        <td class="cp-td cp-td-right">
                          {#if event.metadata && Object.keys(event.metadata).length > 0}
                            <span class="cp-expand-hint" aria-label="Toggle metadata for event {event.id}">
                              {auditExpandedId === event.id ? 'Collapse' : 'Expand'}
                            </span>
                          {/if}
                        </td>
                      </tr>
                      {#if auditExpandedId === event.id && event.metadata}
                        <tr class="cp-tr cp-tr-meta">
                          <td colspan="4" class="cp-td-meta">
                            <pre class="cp-json cp-json-sm" aria-label="Event metadata for {event.id}">{JSON.stringify(event.metadata, null, 2)}</pre>
                          </td>
                        </tr>
                      {/if}
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        {/if}
      </section>
    {/if}

  </div>
</div>

<style>
  /* ---- Page chrome ---- */
  .cp-page-bg { background: var(--dbg); }
  .cp-title { color: var(--dt); }
  .cp-muted { color: var(--dt3); }
  .cp-back-link { color: var(--dt3); }
  .cp-back-link:hover { color: var(--dt2); }

  /* ---- Tab row — mirrors .st-tab-row / .st-tab from settings/+page.svelte ---- */
  .cp-tab-row {
    display: flex;
    gap: 4px;
    border-bottom: 1px solid var(--dbd);
    overflow-x: auto;
    scrollbar-width: none;
  }
  .cp-tab-row::-webkit-scrollbar { display: none; }

  .cp-tab {
    padding: 8px 14px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--dt3);
    border-bottom: 2px solid transparent;
    transition: all 0.15s;
    cursor: pointer;
    white-space: nowrap;
    background: none;
    border-top: none;
    border-left: none;
    border-right: none;
    flex-shrink: 0;
    margin-bottom: -1px;
  }
  .cp-tab:hover {
    color: var(--dt2);
    background: var(--bos-hover-color);
    border-radius: 6px 6px 0 0;
  }
  .cp-tab--active {
    color: var(--dt);
    font-weight: 600;
    border-bottom-color: var(--dt);
  }

  /* ---- Section headers ---- */
  .cp-section-header { }
  .cp-section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--dt);
  }
  .cp-section-desc {
    margin-top: 4px;
    font-size: 0.8125rem;
    color: var(--dt3);
  }

  /* ---- Cards — mirrors .st-acct-card ---- */
  .cp-card {
    background: var(--dbg2);
    border: 1px solid var(--dbd);
    border-radius: 0.75rem;
    padding: 1.25rem 1.5rem;
  }

  /* ---- Form elements ---- */
  .cp-field-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  @media (max-width: 540px) {
    .cp-field-row { grid-template-columns: 1fr; }
  }

  .cp-field {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .cp-label {
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--dt2);
  }
  .cp-required { color: var(--bos-status-error); }

  .cp-input {
    width: 100%;
    padding: 7px 10px;
    font-size: 0.8125rem;
    background: var(--dbg);
    border: 1px solid var(--dbd);
    border-radius: 7px;
    color: var(--dt);
    transition: border-color 0.15s;
    outline: none;
  }
  .cp-input:focus { border-color: var(--dt2); }
  .cp-input::placeholder { color: var(--dt4); }

  .cp-input-danger {
    border-color: var(--bos-status-error);
  }
  .cp-input-danger:focus { border-color: var(--bos-status-error); }

  .cp-textarea { resize: vertical; min-height: 72px; }

  .cp-select {
    width: 100%;
    padding: 7px 10px;
    font-size: 0.8125rem;
    background: var(--dbg);
    border: 1px solid var(--dbd);
    border-radius: 7px;
    color: var(--dt);
    outline: none;
    cursor: pointer;
  }
  .cp-select:focus { border-color: var(--dt2); }

  .cp-checkbox-row {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }
  .cp-checkbox {
    width: 15px;
    height: 15px;
    accent-color: var(--dt);
    cursor: pointer;
    flex-shrink: 0;
  }

  /* ---- Error / info banners ---- */
  .cp-error {
    padding: 8px 12px;
    background: var(--bos-status-error-bg);
    border: 1px solid var(--bos-status-error);
    border-radius: 7px;
    font-size: 0.8125rem;
    color: var(--bos-status-error);
  }

  .cp-danger-banner {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 12px 14px;
    background: var(--bos-status-error-bg);
    border: 1px solid var(--bos-status-error);
    border-radius: 8px;
    color: var(--bos-status-error);
    font-size: 0.8125rem;
  }

  /* ---- Result chrome ---- */
  .cp-result-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .cp-result-title {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--dt);
  }

  .cp-meta-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 16px;
    font-size: 0.8125rem;
  }
  .cp-meta-label { color: var(--dt3); }
  .cp-meta-value { color: var(--dt2); font-weight: 500; }

  .cp-subsection-label {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--dt2);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  /* ---- Badges ---- */
  .cp-badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .cp-badge-success {
    background: var(--bos-status-success-bg);
    color: var(--bos-status-success);
  }
  .cp-badge-warning {
    background: var(--bos-status-warning-bg);
    color: var(--bos-status-warning);
  }
  .cp-badge-error {
    background: var(--bos-status-error-bg);
    color: var(--bos-status-error);
  }
  .cp-badge-info {
    background: var(--bos-status-info-bg);
    color: var(--bos-status-info);
  }

  .cp-kind-badge {
    display: inline-block;
    padding: 2px 6px;
    background: var(--dbg3);
    border-radius: 4px;
    font-size: 0.75rem;
    font-family: ui-monospace, monospace;
    color: var(--dt2);
  }

  .cp-scope-badge {
    display: inline-block;
    padding: 2px 6px;
    background: var(--bos-status-info-bg);
    border-radius: 4px;
    font-size: 0.75rem;
    color: var(--bos-status-info);
    font-weight: 500;
  }

  .cp-action-badge {
    display: inline-block;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
  }
  .cp-action-badge--archive {
    background: var(--bos-status-info-bg);
    color: var(--bos-status-info);
  }
  .cp-action-badge--delete {
    background: var(--bos-status-error-bg);
    color: var(--bos-status-error);
  }
  .cp-action-badge--redact {
    background: var(--bos-status-warning-bg);
    color: var(--bos-status-warning);
  }

  /* ---- Tables ---- */
  .cp-table-wrap {
    overflow-x: auto;
    border-radius: 7px;
    border: 1px solid var(--dbd);
  }
  .cp-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
  }
  .cp-th {
    padding: 8px 12px;
    text-align: left;
    font-weight: 600;
    font-size: 0.75rem;
    color: var(--dt3);
    background: var(--dbg3);
    border-bottom: 1px solid var(--dbd);
    white-space: nowrap;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }
  .cp-th-right { text-align: right; }

  .cp-tr { border-bottom: 1px solid var(--dbd); }
  .cp-tr:last-child { border-bottom: none; }
  .cp-tr-clickable { cursor: pointer; }
  .cp-tr-clickable:hover td { background: var(--dbg3); }

  .cp-tr-meta td { background: var(--dbg2); }

  .cp-td {
    padding: 9px 12px;
    color: var(--dt);
    vertical-align: top;
  }
  .cp-td-right { text-align: right; }
  .cp-td-mono { font-family: ui-monospace, monospace; font-size: 0.75rem; color: var(--dt2); }
  .cp-td-num { font-variant-numeric: tabular-nums; font-weight: 500; }
  .cp-td-dim { color: var(--dt3); font-size: 0.75rem; }

  .cp-td-meta {
    padding: 0;
  }

  .cp-expand-hint {
    font-size: 0.75rem;
    color: var(--dt3);
    text-decoration: underline;
    text-decoration-style: dotted;
    text-underline-offset: 2px;
  }

  /* ---- JSON display ---- */
  .cp-json {
    background: var(--dbg3);
    border: 1px solid var(--dbd);
    border-radius: 7px;
    padding: 12px 14px;
    font-size: 0.75rem;
    font-family: ui-monospace, monospace;
    color: var(--dt2);
    overflow-x: auto;
    white-space: pre;
    max-height: 400px;
    overflow-y: auto;
  }
  .cp-json-sm {
    margin: 0;
    border: none;
    border-radius: 0;
    max-height: 200px;
  }

  /* ---- Confirm zone ---- */
  .cp-confirm-zone {
    padding: 16px;
    background: var(--bos-status-error-bg);
    border: 1px solid var(--bos-status-error);
    border-radius: 8px;
  }
</style>
