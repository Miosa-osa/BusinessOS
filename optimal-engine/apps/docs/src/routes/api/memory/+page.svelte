<script lang="ts">
	import ApiSpec from '$lib/components/ApiSpec.svelte';
	import { memoryEndpoints } from '$lib/data/endpoints.js';
</script>

<svelte:head>
	<title>Memory API — Optimal Engine</title>
</svelte:head>

<h1>Memory API</h1>
<p>
	Every memory carries version chain, typed relations, soft-delete, audience tag, citation URI, and
	metadata. Not a key-value store — a first-class versioned fact with provenance.
</p>

{#each memoryEndpoints as endpoint}
	<ApiSpec {endpoint} />
{/each}

<h2>Additional Endpoints</h2>

<div class="endpoint-list">
	<div class="endpoint-row">
		<span class="method-badge method-get">GET</span>
		<code>/api/memory/:id</code>
		<span class="ep-desc">Fetch one memory by ID. Returns 404 if not found.</span>
	</div>
	<div class="endpoint-row">
		<span class="method-badge method-get">GET</span>
		<code>/api/memory/:id/versions</code>
		<span class="ep-desc">Full version chain in chronological order. Returns <code>{`{ memory_id, root_id, versions: [...] }`}</code></span>
	</div>
	<div class="endpoint-row">
		<span class="method-badge method-get">GET</span>
		<code>/api/memory/:id/relations</code>
		<span class="ep-desc">Inbound and outbound typed relations. Returns <code>{`{ memory_id, inbound: [...], outbound: [...] }`}</code></span>
	</div>
	<div class="endpoint-row">
		<span class="method-badge method-post">POST</span>
		<code>/api/memory/:id/derive</code>
		<span class="ep-desc">Create a derived memory (relation: <code>derives</code>). Use for summaries or inferences drawn from the parent.</span>
	</div>
	<div class="endpoint-row">
		<span class="method-badge method-delete">DELETE</span>
		<code>/api/memory/:id</code>
		<span class="ep-desc">Hard delete. Irreversible. Returns 204.</span>
	</div>
</div>

<h2>Memory Struct Fields</h2>

<table>
	<thead>
		<tr><th>Field</th><th>Type</th><th>Description</th></tr>
	</thead>
	<tbody>
		<tr><td><code>id</code></td><td>string</td><td>Unique memory ID (<code>mem:...</code>)</td></tr>
		<tr><td><code>content</code></td><td>string</td><td>The memory text</td></tr>
		<tr><td><code>workspace_id</code></td><td>string</td><td>Owning workspace</td></tr>
		<tr><td><code>is_static</code></td><td>boolean</td><td>Static (profile static tier) vs dynamic</td></tr>
		<tr><td><code>audience</code></td><td>string | null</td><td>Audience tag for scoped retrieval</td></tr>
		<tr><td><code>version</code></td><td>integer</td><td>Version number (1, 2, 3, …)</td></tr>
		<tr><td><code>parent_memory_id</code></td><td>string | null</td><td>Preceding version ID</td></tr>
		<tr><td><code>root_memory_id</code></td><td>string | null</td><td>First version in chain</td></tr>
		<tr><td><code>is_latest</code></td><td>boolean</td><td>True on the newest version only</td></tr>
		<tr><td><code>is_forgotten</code></td><td>boolean</td><td>Soft-deleted</td></tr>
		<tr><td><code>forget_after</code></td><td>datetime | null</td><td>Scheduled forgetting timestamp</td></tr>
		<tr><td><code>forget_reason</code></td><td>string | null</td><td>Human-readable reason for forgetting</td></tr>
		<tr><td><code>citation_uri</code></td><td>string | null</td><td><code>optimal://</code> URI to source</td></tr>
		<tr><td><code>source_chunk_id</code></td><td>string | null</td><td>Chunk ID this memory derives from</td></tr>
		<tr><td><code>metadata</code></td><td>object | null</td><td>Arbitrary key-value metadata</td></tr>
		<tr><td><code>inserted_at</code></td><td>datetime</td><td>Creation timestamp</td></tr>
		<tr><td><code>updated_at</code></td><td>datetime</td><td>Last update timestamp</td></tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/memory-primitive">Memory Primitive concept</a> — when to use each relation type</li>
	<li><a href="/api/retrieval">Profile endpoint</a> — how static/dynamic memories surface</li>
	<li><a href="/api/recall">Recall API</a> — typed queries that target memory-failure patterns</li>
</ul>

<style>
	.endpoint-list {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		overflow: hidden;
		margin-bottom: 2rem;
	}

	.endpoint-row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.75rem 1.25rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.875rem;
	}

	.endpoint-row:last-child {
		border-bottom: none;
	}

	.ep-desc {
		color: var(--text-muted);
		flex: 1;
	}
</style>
