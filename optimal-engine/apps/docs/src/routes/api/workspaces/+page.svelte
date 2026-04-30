<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Workspaces API — Optimal Engine</title>
</svelte:head>

<h1>Workspaces API</h1>
<p>Create, manage, configure, and archive workspaces. Each workspace is an isolated brain with its own node tree, index, and wiki.</p>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/workspaces</code>
	</div>
	<p>List workspaces in an organization.</p>
	<table>
		<thead><tr><th>Param</th><th>Default</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>tenant</code></td><td>"default"</td><td>Tenant ID</td></tr>
			<tr><td><code>status</code></td><td>"active"</td><td>active | archived | all</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>{'{ tenant_id, workspaces: [{ id, tenant_id, slug, name, description, status, created_at, archived_at, metadata }] }'}</code></p>
	<CodeBlock code={`curl 'http://localhost:4200/api/workspaces?status=all'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/workspaces</code>
	</div>
	<p>Create a workspace.</p>
	<table>
		<thead><tr><th>Field</th><th>Required</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>slug</code></td><td>yes</td><td>URL-safe identifier</td></tr>
			<tr><td><code>name</code></td><td>yes</td><td>Display name</td></tr>
			<tr><td><code>description</code></td><td>no</td><td>Optional description</td></tr>
			<tr><td><code>tenant</code></td><td>no</td><td>Tenant ID (default: "default")</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>201</code> + workspace struct</p>
	<CodeBlock code={`curl -X POST http://localhost:4200/api/workspaces \\
  -H 'Content-Type: application/json' \\
  -d '{"slug":"engineering","name":"Engineering Brain","tenant":"default"}'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/workspaces/:id</code>
	</div>
	<p>Fetch one workspace by ID or slug. Returns 404 if not found.</p>
	<CodeBlock code={`curl http://localhost:4200/api/workspaces/engineering`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-patch">PATCH</span>
		<code>/api/workspaces/:id</code>
	</div>
	<p>Update workspace name or description.</p>
	<CodeBlock code={`curl -X PATCH http://localhost:4200/api/workspaces/engineering \\
  -H 'Content-Type: application/json' \\
  -d '{"name":"Engineering + Platform Brain"}'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/workspaces/:id/archive</code>
	</div>
	<p>Soft-delete a workspace. Data is preserved; queries return 404. Returns <code>204</code>.</p>
	<CodeBlock code={`curl -X POST http://localhost:4200/api/workspaces/engineering/archive`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/workspaces/:id/config</code>
	</div>
	<p>Read merged workspace config (defaults + on-disk <code>.optimal/config.yaml</code>).</p>
	<p>Returns: <code>{'{ workspace_id, config: { ingestion, retrieval, wiki, surfacing } }'}</code></p>
	<CodeBlock code={`curl http://localhost:4200/api/workspaces/engineering/config | jq '.config'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-patch">PATCH</span>
		<code>/api/workspaces/:id/config</code>
	</div>
	<p>Deep-merge body into on-disk config. Returns full merged config after write.</p>
	<CodeBlock code={`curl -X PATCH http://localhost:4200/api/workspaces/engineering/config \\
  -H 'Content-Type: application/json' \\
  -d '{"ingestion":{"min_sn_ratio":0.4},"wiki":{"curation_interval":1800}}'`} lang="bash" />
</div>

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/workspaces">Workspaces concept</a> — on-disk conventions, isolation guarantees</li>
	<li><a href="/api/surfacing">Surfacing API</a> — workspace-scoped subscriptions</li>
</ul>

<style>
	.endpoint-group {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.eg-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}

	.eg-header code {
		font-size: 0.9375rem;
		color: var(--text);
		background: none;
		border: none;
		padding: 0;
	}

	.endpoint-group p {
		font-size: 0.875rem;
	}
</style>
