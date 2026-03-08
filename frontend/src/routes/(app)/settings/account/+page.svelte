<script lang="ts">
	import { goto } from '$app/navigation';
	import { getApiBaseUrl, getCSRFToken } from '$lib/api/base';

	let showDeleteConfirm = $state(false);
	let deleteConfirmText = $state('');
	let isExporting = $state(false);
	let isDeleting = $state(false);
	let error = $state('');
	let exportSuccess = $state(false);

	async function exportData() {
		isExporting = true;
		error = '';
		exportSuccess = false;
		try {
			const response = await fetch(`${getApiBaseUrl()}/account/export`, {
				method: 'GET',
				credentials: 'include'
			});
			if (!response.ok) throw new Error(`Export failed (HTTP ${response.status})`);
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'businessos-data-export.json';
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			exportSuccess = true;
			setTimeout(() => (exportSuccess = false), 4000);
		} catch (e) {
			error = 'Failed to export data. Please try again.';
		} finally {
			isExporting = false;
		}
	}

	async function deleteAccount() {
		if (deleteConfirmText !== 'DELETE') return;
		isDeleting = true;
		error = '';
		try {
			const csrfToken = getCSRFToken();
			const headers: Record<string, string> = {
				'Content-Type': 'application/json'
			};
			if (csrfToken) {
				headers['X-CSRF-Token'] = csrfToken;
			}
			const response = await fetch(`${getApiBaseUrl()}/account`, {
				method: 'DELETE',
				credentials: 'include',
				headers,
				body: JSON.stringify({ confirm: true })
			});
			if (!response.ok) throw new Error(`Deletion failed (HTTP ${response.status})`);
			goto('/login');
		} catch (e) {
			error = 'Failed to delete account. Please try again.';
			isDeleting = false;
		}
	}
</script>

<div class="h-full overflow-y-auto st-page-bg">
	<div class="max-w-2xl mx-auto px-6 py-8">

		<!-- Page header -->
		<div class="mb-8">
			<a
				href="/settings"
				class="inline-flex items-center gap-1.5 text-sm st-back-link mb-4 transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
				Back to Settings
			</a>
			<h1 class="text-2xl font-bold st-title">Account</h1>
			<p class="text-sm st-muted mt-1">Manage your data and account lifecycle</p>
		</div>

		{#if error}
			<div class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-400">
				{error}
			</div>
		{/if}

		<!-- Export My Data -->
		<section class="mb-6 p-6 st-acct-card rounded-xl">
			<div class="flex items-start justify-between gap-4">
				<div>
					<h2 class="text-base font-semibold st-title">Export My Data</h2>
					<p class="text-sm st-muted mt-1">
						Download a copy of all personal data we hold about you, including your profile, settings, conversation history, and activity. The export is provided as a JSON file.
					</p>
					{#if exportSuccess}
						<p class="text-sm text-green-600 dark:text-green-400 mt-2 font-medium">Export downloaded successfully.</p>
					{/if}
				</div>
				<button
					onclick={exportData}
					disabled={isExporting}
					class="shrink-0 btn-pill btn-pill-secondary btn-pill-sm"
					aria-label="Export my data as a JSON file"
				>
					{#if isExporting}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Exporting...
					{:else}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
						</svg>
						Export Data
					{/if}
				</button>
			</div>
		</section>

		<!-- Danger Zone -->
		<section class="p-6 border border-red-200 dark:border-red-800 rounded-xl bg-red-50 dark:bg-red-900/10">
			<h2 class="text-base font-semibold text-red-700 dark:text-red-400 mb-1">Danger Zone</h2>
			<p class="text-sm text-red-600 dark:text-red-500 mb-4">
				Deleting your account is permanent and cannot be undone. All your data — including conversation history, settings, and projects — will be erased within 30 days.
			</p>

			{#if !showDeleteConfirm}
				<button
					onclick={() => { showDeleteConfirm = true; deleteConfirmText = ''; error = ''; }}
					class="btn-pill btn-pill-danger btn-pill-sm"
					aria-label="Begin account deletion process"
				>
					Delete My Account
				</button>
			{:else}
				<div class="space-y-4">
					<p class="text-sm text-red-700 dark:text-red-400 font-medium">
						To confirm, type <strong>DELETE</strong> in the box below:
					</p>
					<input
						type="text"
						bind:value={deleteConfirmText}
						placeholder="Type DELETE to confirm"
						class="w-full max-w-xs px-3 py-2 text-sm st-acct-danger-input rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
						aria-label="Confirmation text field — type DELETE"
					/>
					<div class="flex gap-3">
						<button
							onclick={deleteAccount}
							disabled={deleteConfirmText !== 'DELETE' || isDeleting}
							class="btn-pill btn-pill-danger btn-pill-sm"
							aria-label="Confirm account deletion"
						>
							{#if isDeleting}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Deleting...
							{:else}
								Permanently Delete Account
							{/if}
						</button>
						<button
							onclick={() => { showDeleteConfirm = false; deleteConfirmText = ''; }}
							class="btn-pill btn-pill-soft btn-pill-sm"
							aria-label="Cancel account deletion"
						>
							Cancel
						</button>
					</div>
				</div>
			{/if}
		</section>

	</div>
</div>

<style>
  :global(.st-page-bg) { background: var(--dbg, var(--bos-card, #fff)); }
  :global(.st-title) { color: var(--dt, var(--bos-text-primary, #111)); }
  :global(.st-muted) { color: var(--dt3, var(--bos-text-tertiary, #888)); }
  :global(.st-back-link) {
    color: var(--dt3, var(--bos-text-tertiary, #888));
  }
  :global(.st-back-link:hover) {
    color: var(--dt2, var(--bos-text-secondary, #555));
  }
  :global(.st-acct-card) {
    background: var(--dbg2, var(--bos-bg-secondary, #f5f5f5));
    border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
  }
  :global(.st-acct-danger-input) {
    background: var(--dbg, var(--bos-card, #fff));
    border: 1px solid #ef4444;
    color: var(--dt, var(--bos-text-primary, #111));
  }
  :global(.st-acct-danger-input::placeholder) {
    color: var(--dt4, #bbb);
  }
</style>
