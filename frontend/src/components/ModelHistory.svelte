<script lang="ts">
	import { onMount } from 'svelte';
	import type { ProcessModelVersion, VersionDiffResult } from '$lib/types';

	interface Props {
		modelId: string;
	}

	let { modelId }: Props = $props();

	let versions: ProcessModelVersion[] = $state([]);
	let selectedVersion: ProcessModelVersion | null = $state(null);
	let compareMode = $state(false);
	let compareWith: ProcessModelVersion | null = $state(null);
	let diffResult: VersionDiffResult | null = $state(null);
	let loading = $state(false);
	let error = $state<string | null>(null);

	onMount(async () => {
		await loadVersionHistory();
	});

	async function loadVersionHistory() {
		loading = true;
		error = null;

		try {
			const response = await fetch(`/api/process-models/${modelId}/versions`);
			if (!response.ok) throw new Error('Failed to load version history');

			const data = await response.json();
			versions = data.versions || [];

			if (versions.length > 0) {
				selectedVersion = versions[0];
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	async function selectVersion(version: ProcessModelVersion) {
		selectedVersion = version;
		compareMode = false;
		diffResult = null;
	}

	async function startComparison(version: ProcessModelVersion) {
		compareMode = true;
		compareWith = version;
		await loadDiff();
	}

	async function loadDiff() {
		if (!selectedVersion || !compareWith) return;

		loading = true;
		error = null;

		try {
			const response = await fetch(
				`/api/process-models/${modelId}/versions/compare?from=${selectedVersion.version}&to=${compareWith.version}`
			);
			if (!response.ok) throw new Error('Failed to load comparison');

			diffResult = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	async function releaseVersion(version: ProcessModelVersion) {
		loading = true;
		error = null;

		try {
			const response = await fetch(`/api/process-models/${modelId}/versions/${version.id}/release`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ release_notes: '' })
			});

			if (!response.ok) throw new Error('Failed to release version');

			await loadVersionHistory();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	async function rollbackToVersion(version: ProcessModelVersion) {
		if (!confirm(`Rollback to version ${version.version}? This will affect new instances.`)) {
			return;
		}

		loading = true;
		error = null;

		try {
			const response = await fetch(`/api/process-models/${modelId}/rollback`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					target_version: version.version,
					reason: 'User-initiated rollback',
					approved_by: 'current_user'
				})
			});

			if (!response.ok) throw new Error('Rollback failed');

			await loadVersionHistory();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	function formatDate(date: string | Date): string {
		const d = typeof date === 'string' ? new Date(date) : date;
		return d.toLocaleString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getVersionBadgeColor(changeType: string): string {
		switch (changeType) {
			case 'major':
				return 'bg-red-100 text-red-800';
			case 'minor':
				return 'bg-yellow-100 text-yellow-800';
			case 'patch':
				return 'bg-green-100 text-green-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}
</script>

<div class="model-history-container">
	<div class="header">
		<h2>Model Version History</h2>
		<p class="subtitle">Track and manage process model versions over time</p>
	</div>

	{#if error}
		<div class="error-banner">
			<p>{error}</p>
		</div>
	{/if}

	<div class="timeline-container">
		<!-- Timeline View -->
		<div class="timeline-column">
			<h3>Versions</h3>
			<div class="timeline-list">
				{#if loading && versions.length === 0}
					<div class="loading">Loading versions...</div>
				{:else if versions.length === 0}
					<div class="empty">No versions found</div>
				{:else}
					{#each versions as version (version.id)}
						<div
							class="timeline-item {selectedVersion?.id === version.id ? 'selected' : ''}"
							onclick={() => selectVersion(version)}
						>
							<div class="version-dot"></div>
							<div class="version-info">
								<div class="version-header">
									<span class="version-number">{version.version}</span>
									<span class="change-type {getVersionBadgeColor(version.change_type)}">
										{version.change_type.toUpperCase()}
									</span>
									{#if version.is_released}
										<span class="released-badge">Released</span>
									{/if}
								</div>
								<p class="version-date">{formatDate(version.created_at)}</p>
								<p class="version-creator">By {version.created_by}</p>
								{#if version.description}
									<p class="version-desc">{version.description}</p>
								{/if}
							</div>
						</div>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Details Panel -->
		<div class="details-column">
			{#if selectedVersion}
				<div class="details-panel">
					<h3>Version {selectedVersion.version} Details</h3>

					<!-- Metadata -->
					<section class="metadata-section">
						<h4>Metadata</h4>
						<div class="metadata-grid">
							<div class="metadata-item">
								<label>Created At</label>
								<span>{formatDate(selectedVersion.created_at)}</span>
							</div>
							<div class="metadata-item">
								<label>Created By</label>
								<span>{selectedVersion.created_by}</span>
							</div>
							<div class="metadata-item">
								<label>Content Hash</label>
								<span class="hash">{selectedVersion.content_hash}</span>
							</div>
							{#if selectedVersion.discovery_source}
								<div class="metadata-item">
									<label>Discovery Source</label>
									<span>{selectedVersion.discovery_source}</span>
								</div>
							{/if}
						</div>
					</section>

					<!-- Metrics -->
					{#if selectedVersion.nodes_count !== undefined}
						<section class="metrics-section">
							<h4>Metrics</h4>
							<div class="metrics-grid">
								<div class="metric-card">
									<span class="metric-label">Nodes</span>
									<span class="metric-value">{selectedVersion.nodes_count}</span>
								</div>
								<div class="metric-card">
									<span class="metric-label">Edges</span>
									<span class="metric-value">{selectedVersion.edges_count}</span>
								</div>
								<div class="metric-card">
									<span class="metric-label">Variants</span>
									<span class="metric-value">{selectedVersion.variants}</span>
								</div>
								<div class="metric-card">
									<span class="metric-label">Fitness</span>
									<span class="metric-value">{(selectedVersion.fitness * 100).toFixed(1)}%</span>
								</div>
								<div class="metric-card">
									<span class="metric-label">Avg Duration</span>
									<span class="metric-value">{selectedVersion.average_duration?.toFixed(1)} min</span>
								</div>
								<div class="metric-card">
									<span class="metric-label">Covered Traces</span>
									<span class="metric-value">{selectedVersion.covered_traces}</span>
								</div>
							</div>
						</section>
					{/if}

					<!-- Changes -->
					{#if selectedVersion.nodes_added || selectedVersion.edges_added}
						<section class="changes-section">
							<h4>Changes</h4>
							<div class="changes-grid">
								{#if selectedVersion.nodes_added > 0}
									<div class="change-item added">
										<span class="change-icon">+</span>
										<span>{selectedVersion.nodes_added} nodes added</span>
									</div>
								{/if}
								{#if selectedVersion.nodes_removed > 0}
									<div class="change-item removed">
										<span class="change-icon">−</span>
										<span>{selectedVersion.nodes_removed} nodes removed</span>
									</div>
								{/if}
								{#if selectedVersion.edges_added > 0}
									<div class="change-item added">
										<span class="change-icon">+</span>
										<span>{selectedVersion.edges_added} edges added</span>
									</div>
								{/if}
								{#if selectedVersion.edges_removed > 0}
									<div class="change-item removed">
										<span class="change-icon">−</span>
										<span>{selectedVersion.edges_removed} edges removed</span>
									</div>
								{/if}
							</div>
						</section>
					{/if}

					<!-- Actions -->
					<section class="actions-section">
						<h4>Actions</h4>
						<div class="action-buttons">
							{#if !selectedVersion.is_released}
								<button
									class="btn btn-primary"
									onclick={() => releaseVersion(selectedVersion)}
									disabled={loading}
								>
									Release Version
								</button>
							{/if}
							<button
								class="btn btn-secondary"
								onclick={() => startComparison(selectedVersion)}
								disabled={versions.length < 2}
							>
								Compare with Another
							</button>
							{#if selectedVersion.is_released}
								<button
									class="btn btn-warning"
									onclick={() => rollbackToVersion(selectedVersion)}
								>
									Rollback to This Version
								</button>
							{/if}
						</div>
					</section>
				</div>
			{:else}
				<div class="empty-state">
					<p>Select a version to view details</p>
				</div>
			{/if}
		</div>
	</div>

	<!-- Comparison View -->
	{#if compareMode && compareWith && selectedVersion}
		<div class="comparison-panel">
			<div class="comparison-header">
				<h3>Comparing versions</h3>
				<button class="close-btn" onclick={() => { compareMode = false; diffResult = null; }}>×</button>
			</div>

			<div class="comparison-versions">
				<div class="comparison-side">
					<h4>From: {selectedVersion.version}</h4>
					<p class="version-date">{formatDate(selectedVersion.created_at)}</p>
				</div>
				<div class="comparison-arrow">→</div>
				<div class="comparison-side">
					<h4>To: {compareWith.version}</h4>
					<p class="version-date">{formatDate(compareWith.created_at)}</p>
				</div>
			</div>

			{#if diffResult}
				<div class="diff-results">
					<!-- Structural Changes -->
					<section class="diff-section">
						<h4>Structural Changes</h4>
						<div class="structural-diff">
							{#if diffResult.structural_diff.nodes_added?.length > 0}
								<div class="diff-category added">
									<h5>Nodes Added ({diffResult.structural_diff.nodes_added.length})</h5>
									<ul>
										{#each diffResult.structural_diff.nodes_added as node}
											<li>{node.label} <code>{node.type}</code></li>
										{/each}
									</ul>
								</div>
							{/if}
							{#if diffResult.structural_diff.nodes_removed?.length > 0}
								<div class="diff-category removed">
									<h5>Nodes Removed ({diffResult.structural_diff.nodes_removed.length})</h5>
									<ul>
										{#each diffResult.structural_diff.nodes_removed as node}
											<li>{node.label} <code>{node.type}</code></li>
										{/each}
									</ul>
								</div>
							{/if}
							{#if diffResult.structural_diff.edges_added?.length > 0}
								<div class="diff-category added">
									<h5>Edges Added ({diffResult.structural_diff.edges_added.length})</h5>
									<ul>
										{#each diffResult.structural_diff.edges_added as edge}
											<li>{edge.source} → {edge.target}</li>
										{/each}
									</ul>
								</div>
							{/if}
						</div>
					</section>

					<!-- Metrics Changes -->
					<section class="diff-section">
						<h4>Metrics Changes</h4>
						<div class="metrics-diff">
							<div class="metric-diff-item">
								<span>Nodes: </span>
								<span class="from">{diffResult.metrics_diff.nodes_count.before}</span>
								<span class="arrow">→</span>
								<span class="to">{diffResult.metrics_diff.nodes_count.after}</span>
								<span class="delta" class:positive={diffResult.metrics_diff.nodes_count.delta > 0}>
									({diffResult.metrics_diff.nodes_count.delta > 0 ? '+' : ''}{diffResult.metrics_diff.nodes_count.delta})
								</span>
							</div>
							<div class="metric-diff-item">
								<span>Fitness: </span>
								<span class="from">{(diffResult.metrics_diff.fitness.before * 100).toFixed(1)}%</span>
								<span class="arrow">→</span>
								<span class="to">{(diffResult.metrics_diff.fitness.after * 100).toFixed(1)}%</span>
								<span class="delta" class:positive={diffResult.metrics_diff.fitness.delta > 0}>
									({diffResult.metrics_diff.fitness.delta > 0 ? '+' : ''}{(diffResult.metrics_diff.fitness.delta * 100).toFixed(1)}%)
								</span>
							</div>
						</div>
					</section>

					<!-- Breaking Changes -->
					{#if diffResult.breaking_changes?.length > 0}
						<section class="diff-section warning">
							<h4>⚠️ Breaking Changes</h4>
							<ul class="breaking-changes">
								{#each diffResult.breaking_changes as change}
									<li>{change}</li>
								{/each}
							</ul>
						</section>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.model-history-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
		padding: 2rem;
		background: var(--bg-secondary);
		border-radius: 0.5rem;
	}

	.header {
		border-bottom: 2px solid var(--border-color);
		padding-bottom: 1rem;
	}

	.header h2 {
		margin: 0 0 0.5rem 0;
		font-size: 1.75rem;
		font-weight: 600;
	}

	.subtitle {
		margin: 0;
		color: var(--text-secondary);
		font-size: 0.95rem;
	}

	.error-banner {
		padding: 1rem;
		background: #fee;
		border: 1px solid #fcc;
		border-radius: 0.375rem;
		color: #c00;
	}

	.timeline-container {
		display: grid;
		grid-template-columns: 350px 1fr;
		gap: 2rem;
	}

	.timeline-column h3,
	.details-column h3 {
		margin: 0 0 1rem 0;
		font-size: 1.1rem;
		font-weight: 600;
	}

	.timeline-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		max-height: 600px;
		overflow-y: auto;
	}

	.timeline-item {
		display: flex;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-primary);
		border: 2px solid transparent;
		border-radius: 0.375rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.timeline-item:hover {
		background: var(--bg-hover);
	}

	.timeline-item.selected {
		border-color: var(--primary-color);
		background: var(--bg-hover);
	}

	.version-dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: var(--primary-color);
		margin-top: 0.25rem;
		flex-shrink: 0;
	}

	.version-info {
		flex: 1;
		min-width: 0;
	}

	.version-header {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		margin-bottom: 0.25rem;
		flex-wrap: wrap;
	}

	.version-number {
		font-weight: 600;
		font-family: monospace;
		font-size: 0.9rem;
	}

	.change-type {
		font-size: 0.7rem;
		padding: 0.2rem 0.4rem;
		border-radius: 0.2rem;
		font-weight: 600;
	}

	.released-badge {
		font-size: 0.7rem;
		padding: 0.2rem 0.4rem;
		border-radius: 0.2rem;
		background: #d0f0c0;
		color: #0a5f0a;
		font-weight: 600;
	}

	.version-date {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.version-creator {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-tertiary);
	}

	.version-desc {
		margin: 0.25rem 0 0 0;
		font-size: 0.85rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.details-panel {
		padding: 1.5rem;
		background: var(--bg-primary);
		border-radius: 0.375rem;
		border: 1px solid var(--border-color);
	}

	.details-panel h3 {
		margin: 0 0 1.5rem 0;
	}

	section {
		margin-bottom: 2rem;
	}

	section h4 {
		margin: 0 0 1rem 0;
		font-size: 1rem;
		font-weight: 600;
	}

	.metadata-section,
	.metrics-section,
	.changes-section {
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 0.375rem;
	}

	.metadata-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
	}

	.metadata-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.metadata-item label {
		font-size: 0.85rem;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.metadata-item span {
		font-size: 0.95rem;
		color: var(--text-primary);
		word-break: break-word;
	}

	.hash {
		font-family: monospace;
		font-size: 0.85rem;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
	}

	.metric-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1rem;
		background: var(--bg-primary);
		border-radius: 0.375rem;
		text-align: center;
	}

	.metric-label {
		font-size: 0.85rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.metric-value {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--primary-color);
	}

	.changes-grid {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.change-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		border-radius: 0.3rem;
		font-size: 0.9rem;
	}

	.change-item.added {
		background: #efe;
		color: #0a5f0a;
	}

	.change-item.removed {
		background: #fee;
		color: #c00;
	}

	.change-icon {
		font-weight: 600;
		font-size: 1.1rem;
	}

	.action-buttons {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.btn {
		padding: 0.5rem 1rem;
		border: 1px solid transparent;
		border-radius: 0.375rem;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s ease;
		font-size: 0.9rem;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-hover);
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text-primary);
		border-color: var(--border-color);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--bg-hover);
	}

	.btn-warning {
		background: #ff9800;
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: #fb8c00;
	}

	.empty-state {
		text-align: center;
		padding: 3rem 1rem;
		color: var(--text-secondary);
	}

	.loading,
	.empty {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
	}

	.comparison-panel {
		margin-top: 2rem;
		padding: 1.5rem;
		background: var(--bg-secondary);
		border: 2px solid var(--primary-color);
		border-radius: 0.5rem;
	}

	.comparison-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-color);
		padding-bottom: 1rem;
	}

	.comparison-header h3 {
		margin: 0;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: var(--text-secondary);
	}

	.comparison-versions {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding: 1rem;
		background: var(--bg-primary);
		border-radius: 0.375rem;
	}

	.comparison-side {
		flex: 1;
	}

	.comparison-side h4 {
		margin: 0 0 0.5rem 0;
	}

	.comparison-arrow {
		padding: 0 1rem;
		font-size: 1.5rem;
		font-weight: 600;
	}

	.diff-results {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.diff-section {
		padding: 1rem;
		background: var(--bg-primary);
		border-radius: 0.375rem;
	}

	.diff-section.warning {
		border-left: 4px solid #ff9800;
	}

	.diff-section h4 {
		margin: 0 0 1rem 0;
	}

	.structural-diff {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.diff-category {
		padding: 0.75rem;
		border-radius: 0.3rem;
	}

	.diff-category.added {
		background: #f0fdf4;
		border-left: 3px solid #22c55e;
	}

	.diff-category.removed {
		background: #fef2f2;
		border-left: 3px solid #ef4444;
	}

	.diff-category h5 {
		margin: 0 0 0.5rem 0;
		font-size: 0.9rem;
	}

	.diff-category ul {
		margin: 0;
		padding-left: 1.5rem;
	}

	.diff-category li {
		margin: 0.25rem 0;
		font-size: 0.9rem;
	}

	.diff-category code {
		background: var(--bg-secondary);
		padding: 0.1rem 0.3rem;
		border-radius: 0.2rem;
		font-size: 0.85rem;
		font-family: monospace;
	}

	.metrics-diff {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.metric-diff-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background: var(--bg-secondary);
		border-radius: 0.3rem;
		font-size: 0.9rem;
	}

	.from,
	.to {
		font-weight: 500;
	}

	.arrow {
		color: var(--text-secondary);
	}

	.delta {
		font-weight: 600;
		color: #ef4444;
	}

	.delta.positive {
		color: #22c55e;
	}

	.breaking-changes {
		margin: 0;
		padding-left: 1.5rem;
	}

	.breaking-changes li {
		margin: 0.5rem 0;
		color: #d97706;
		font-weight: 500;
	}

	@media (max-width: 1024px) {
		.timeline-container {
			grid-template-columns: 1fr;
		}

		.metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 640px) {
		.timeline-item {
			padding: 0.75rem;
		}

		.metadata-grid {
			grid-template-columns: 1fr;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.comparison-versions {
			flex-direction: column;
			gap: 1rem;
		}

		.comparison-arrow {
			transform: rotate(90deg);
		}

		.action-buttons {
			flex-direction: column;
		}

		.btn {
			width: 100%;
		}
	}
</style>
