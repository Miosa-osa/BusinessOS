<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Proactive Surfacing — Optimal Engine</title>
</svelte:head>

<h1>Proactive Surfacing</h1>
<p>
	The Surfacer watches for relevant changes in the engine and pushes notifications to subscribers
	via Server-Sent Events. Instead of agents polling for new information, the engine pushes when
	something relevant changes.
</p>

<h2>The 14 Surfacing Categories</h2>

<p>Derived from the Engramme enterprise memory taxonomy:</p>

<table>
	<thead>
		<tr><th>#</th><th>Category</th><th>Triggered by</th></tr>
	</thead>
	<tbody>
		<tr><td>1</td><td><code>recent_actions</code></td><td>New decisions or commits by tracked actors</td></tr>
		<tr><td>2</td><td><code>ownership</code></td><td>Ownership changes for entities in scope</td></tr>
		<tr><td>3</td><td><code>contradictions</code></td><td>New signals that contradict existing wiki claims</td></tr>
		<tr><td>4</td><td><code>blockers</code></td><td><code>express_concern</code> chunks about tracked topics</td></tr>
		<tr><td>5</td><td><code>deadlines</code></td><td>Temporal signals about upcoming events</td></tr>
		<tr><td>6</td><td><code>handoffs</code></td><td>Actor transitions or delegation events</td></tr>
		<tr><td>7</td><td><code>escalations</code></td><td>Signals with elevated concern scores</td></tr>
		<tr><td>8</td><td><code>metrics</code></td><td>New <code>measure</code>-intent chunks for tracked entities</td></tr>
		<tr><td>9</td><td><code>specifications</code></td><td>New <code>specify</code>-intent chunks (contract changes)</td></tr>
		<tr><td>10</td><td><code>references</code></td><td>Cross-references to tracked topics from other nodes</td></tr>
		<tr><td>11</td><td><code>clusters</code></td><td>New thematic clusters containing tracked entities</td></tr>
		<tr><td>12</td><td><code>wiki_updates</code></td><td>Wiki page updates for subscribed slugs</td></tr>
		<tr><td>13</td><td><code>entity_graph</code></td><td>New edges in the entity graph for tracked entities</td></tr>
		<tr><td>14</td><td><code>compliance</code></td><td>PII detection or retention-trigger events</td></tr>
	</tbody>
</table>

<h2>Creating a Subscription</h2>

<p>Subscribe to any workspace, topic, node, or entity:</p>

<CodeBlock code={`# Subscribe to blockers and contradictions across the sales workspace
curl -X POST http://localhost:4200/api/subscriptions \\
  -H 'Content-Type: application/json' \\
  -d '{
    "workspace": "sales",
    "scope": "workspace",
    "scope_value": "sales",
    "categories": ["blockers","contradictions","recent_actions"],
    "principal_id": "agent:sales-assistant"
  }'

# Subscribe to a specific topic
curl -X POST http://localhost:4200/api/subscriptions \\
  -H 'Content-Type: application/json' \\
  -d '{
    "workspace": "sales",
    "scope": "topic",
    "scope_value": "pricing",
    "categories": ["ownership","contradictions","specifications"],
    "principal_id": "agent:pricing-monitor"
  }'`} lang="bash" />

<h2>Receiving Events via SSE</h2>

<CodeBlock code={`# Connect to the SSE stream
curl -N 'http://localhost:4200/api/surface/stream?subscription=sub:abc123'

# Each event is a newline-delimited JSON envelope:
# data: {"category":"contradictions","workspace":"sales","page_slug":"healthtech-pricing","detected_at":"2026-04-28T15:00:00Z","score":0.82}`} lang="bash" />

<p>In TypeScript:</p>

<CodeBlock code={`const es = new EventSource(
  'http://localhost:4200/api/surface/stream?subscription=sub:abc123'
);

es.onmessage = (event) => {
  const payload = JSON.parse(event.data);
  console.log('Surfaced:', payload.category, payload.page_slug);
};`} lang="typescript" />

<h2>Scope Types</h2>

<table>
	<thead>
		<tr><th>Scope</th><th>scope_value</th><th>What it watches</th></tr>
	</thead>
	<tbody>
		<tr><td><code>workspace</code></td><td>workspace slug</td><td>All events in the workspace</td></tr>
		<tr><td><code>topic</code></td><td>keyword or entity name</td><td>Events involving this topic across nodes</td></tr>
		<tr><td><code>node</code></td><td>node slug</td><td>Events in a specific node (e.g., <code>03-sales</code>)</td></tr>
		<tr><td><code>entity</code></td><td>entity name</td><td>Events where this entity appears</td></tr>
	</tbody>
</table>

<h2>Managing Subscriptions</h2>

<CodeBlock code={`# List all subscriptions for a workspace
curl 'http://localhost:4200/api/subscriptions?workspace=sales'

# Pause without deleting
curl -X POST http://localhost:4200/api/subscriptions/sub:abc123/pause

# Resume
curl -X POST http://localhost:4200/api/subscriptions/sub:abc123/resume

# Delete
curl -X DELETE http://localhost:4200/api/subscriptions/sub:abc123`} lang="bash" />

<h2>Testing a Subscription</h2>

<CodeBlock code={`# Trigger a synthetic push to test your listener
curl -X POST http://localhost:4200/api/surface/test \\
  -H 'Content-Type: application/json' \\
  -d '{"subscription":"sub:abc123","slug":"healthtech-pricing"}'`} lang="bash" />

<h2>Workspace Config</h2>

<p>
	Configure which categories a workspace monitors in <code>.optimal/config.yaml</code>. Only
	categories listed here will fire events:
</p>

<CodeBlock code={`surfacing:
  categories:
    - recent_actions
    - blockers
    - contradictions
    - specifications
    - metrics`} lang="yaml" filename=".optimal/config.yaml" />

<h2>See Also</h2>
<ul>
	<li><a href="/api/surfacing">Surfacing API</a> — full endpoint reference</li>
	<li><a href="/concepts/workspaces">Workspaces</a> — subscription isolation guarantees</li>
	<li><a href="/concepts/signal-theory">Signal Theory</a> — contradiction detection maps to Beer failure modes</li>
</ul>
