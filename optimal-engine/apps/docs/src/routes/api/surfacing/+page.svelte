<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Surfacing API — Optimal Engine</title>
</svelte:head>

<h1>Surfacing API</h1>
<p>
	Subscriptions + Server-Sent Events. Subscribe to any workspace, topic, node, or entity. The
	engine pushes relevant changes across 14 categories without polling.
</p>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/subscriptions</code>
	</div>
	<p>List subscriptions for a workspace.</p>
	<table>
		<thead><tr><th>Param</th><th>Default</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>workspace</code></td><td>"default"</td><td>Workspace scope</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>{'{ workspace_id, subscriptions: [{ id, scope, scope_value, categories, principal_id, status, created_at }] }'}</code></p>
	<CodeBlock code={`curl 'http://localhost:4200/api/subscriptions?workspace=sales'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/subscriptions</code>
	</div>
	<p>Create a subscription.</p>
	<table>
		<thead><tr><th>Field</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>workspace</code></td><td>Workspace scope</td></tr>
			<tr><td><code>scope</code></td><td>workspace | topic | node | entity</td></tr>
			<tr><td><code>scope_value</code></td><td>Slug / name / ID matching scope</td></tr>
			<tr><td><code>categories</code></td><td>Array of 14-category surfacing labels</td></tr>
			<tr><td><code>principal_id</code></td><td>Who receives the pushes (agent ID, user ID)</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>201</code> + subscription struct</p>
	<CodeBlock code={`curl -X POST http://localhost:4200/api/subscriptions \\
  -H 'Content-Type: application/json' \\
  -d '{
    "workspace": "sales",
    "scope": "workspace",
    "scope_value": "sales",
    "categories": ["blockers","contradictions","recent_actions","ownership"],
    "principal_id": "agent:sales-assistant"
  }'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/surface/stream?subscription=:id</code>
	</div>
	<p>
		Server-Sent Events stream. Connect with <code>EventSource</code>. Pushes newline-delimited JSON
		envelopes when the Surfacer fires. One connection per subscription.
	</p>
	<CodeBlock code={`# curl — keep connection open
curl -N 'http://localhost:4200/api/surface/stream?subscription=sub:abc123'

# Browser / Node.js
const es = new EventSource(
  'http://localhost:4200/api/surface/stream?subscription=sub:abc123'
);

es.onmessage = (event) => {
  const payload = JSON.parse(event.data);
  // {
  //   "category": "contradictions",
  //   "workspace": "sales",
  //   "page_slug": "healthtech-pricing",
  //   "detected_at": "2026-04-28T15:00:00Z",
  //   "score": 0.82,
  //   "entities": ["Alice", "Q4 Pricing"]
  // }
};`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/subscriptions/:id/pause</code>
	</div>
	<p>Pause delivery without deleting the subscription. Returns <code>204</code>.</p>
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/subscriptions/:id/resume</code>
	</div>
	<p>Resume a paused subscription. Returns <code>204</code>.</p>
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-delete">DELETE</span>
		<code>/api/subscriptions/:id</code>
	</div>
	<p>Delete a subscription. Returns <code>204</code>.</p>
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-post">POST</span>
		<code>/api/surface/test</code>
	</div>
	<p>Trigger a synthetic push to all listeners of a subscription. Useful for testing your SSE handler.</p>
	<CodeBlock code={`curl -X POST http://localhost:4200/api/surface/test \\
  -H 'Content-Type: application/json' \\
  -d '{"subscription":"sub:abc123","slug":"healthtech-pricing"}'`} lang="bash" />
</div>

<h2>The 14 Categories</h2>

<table>
	<thead>
		<tr><th>Category</th><th>Triggered by</th></tr>
	</thead>
	<tbody>
		<tr><td><code>recent_actions</code></td><td>New decisions or commits by tracked actors</td></tr>
		<tr><td><code>ownership</code></td><td>Ownership changes for entities in scope</td></tr>
		<tr><td><code>contradictions</code></td><td>New signals contradicting existing wiki claims</td></tr>
		<tr><td><code>blockers</code></td><td><code>express_concern</code> chunks about tracked topics</td></tr>
		<tr><td><code>deadlines</code></td><td>Temporal signals about upcoming events</td></tr>
		<tr><td><code>handoffs</code></td><td>Actor transitions or delegation events</td></tr>
		<tr><td><code>escalations</code></td><td>Signals with elevated concern scores</td></tr>
		<tr><td><code>metrics</code></td><td>New <code>measure</code>-intent chunks for tracked entities</td></tr>
		<tr><td><code>specifications</code></td><td>New <code>specify</code>-intent chunks</td></tr>
		<tr><td><code>references</code></td><td>Cross-references to tracked topics from other nodes</td></tr>
		<tr><td><code>clusters</code></td><td>New thematic clusters containing tracked entities</td></tr>
		<tr><td><code>wiki_updates</code></td><td>Wiki page updates for subscribed slugs</td></tr>
		<tr><td><code>entity_graph</code></td><td>New edges in the entity graph for tracked entities</td></tr>
		<tr><td><code>compliance</code></td><td>PII detection or retention-trigger events</td></tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/proactive-surfacing">Proactive Surfacing concept</a> — full category taxonomy</li>
	<li><a href="/concepts/workspaces">Workspaces</a> — subscription isolation guarantees</li>
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
