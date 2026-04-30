<script lang="ts">
	import type { Endpoint } from '$lib/data/endpoints.js';
	import CodeBlock from './CodeBlock.svelte';

	interface Props {
		endpoint: Endpoint;
	}

	let { endpoint }: Props = $props();

	const methodClass: Record<string, string> = {
		GET: 'method-get',
		POST: 'method-post',
		PATCH: 'method-patch',
		DELETE: 'method-delete'
	};
</script>

<div class="api-spec">
	<div class="spec-header">
		<span class="method-badge {methodClass[endpoint.method]}">{endpoint.method}</span>
		<code class="path">{endpoint.path}</code>
	</div>

	<p class="summary">{endpoint.summary}</p>
	<p class="description">{endpoint.description}</p>

	{#if endpoint.params && endpoint.params.length > 0}
		<h4>Query Parameters</h4>
		<table>
			<thead>
				<tr>
					<th>Param</th>
					<th>Type</th>
					<th>Required</th>
					<th>Default</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				{#each endpoint.params as p}
					<tr>
						<td><code>{p.name}</code></td>
						<td><span class="type-tag">{p.type}</span></td>
						<td>{p.required ? 'yes' : 'no'}</td>
						<td>{p.default ?? '—'}</td>
						<td>{p.description}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}

	{#if endpoint.body && endpoint.body.length > 0}
		<h4>Request Body</h4>
		<table>
			<thead>
				<tr>
					<th>Field</th>
					<th>Type</th>
					<th>Required</th>
					<th>Default</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				{#each endpoint.body as p}
					<tr>
						<td><code>{p.name}</code></td>
						<td><span class="type-tag">{p.type}</span></td>
						<td>{p.required ? 'yes' : 'no'}</td>
						<td>{p.default ?? '—'}</td>
						<td>{p.description}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}

	<h4>Returns</h4>
	<p class="returns-val"><code>{endpoint.returns}</code></p>

	<h4>Use When</h4>
	<p class="use-when">{endpoint.useWhen}</p>

	<CodeBlock code={endpoint.example} lang="bash" />
</div>

<style>
	.api-spec {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.spec-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}

	.path {
		font-size: 0.9375rem;
		color: var(--text);
		background: none;
		border: none;
		padding: 0;
	}

	.summary {
		font-weight: 600;
		color: var(--text);
		margin-bottom: 0.375rem;
	}

	.description {
		color: var(--text-muted);
		font-size: 0.875rem;
		margin-bottom: 1rem;
	}

	.api-spec :global(h4) {
		margin-top: 1.25rem;
		margin-bottom: 0.5rem;
	}

	.type-tag {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.returns-val {
		margin-bottom: 0.75rem;
	}

	.use-when {
		margin-bottom: 1rem;
	}
</style>
