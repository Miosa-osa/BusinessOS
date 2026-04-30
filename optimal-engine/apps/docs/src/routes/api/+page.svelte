<script lang="ts">
</script>

<svelte:head>
	<title>API Overview — Optimal Engine</title>
</svelte:head>

<h1>API Reference</h1>
<p>
	Base URL: <code>http://localhost:4200</code> (configurable via
	<code>config :optimal_engine, :api, port: 4200</code>). All responses are JSON. CORS is open
	(<code>*</code>) by default. All <code>POST</code>/<code>PATCH</code> bodies are JSON.
</p>

<h2>API Surface Map</h2>

<div class="api-map">
	<a href="/api/retrieval" class="api-group">
		<div class="api-group-title">Retrieval</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/rag</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/search</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/grep</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/profile</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/l0</div>
		</div>
	</a>
	<a href="/api/recall" class="api-group">
		<div class="api-group-title">Recall</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/recall/actions</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/recall/who</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/recall/when</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/recall/where</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/recall/owns</div>
		</div>
	</a>
	<a href="/api/memory" class="api-group">
		<div class="api-group-title">Memory</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/memory</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/memory</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/memory/:id</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/memory/:id/versions</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/memory/:id/relations</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/memory/:id/update</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/memory/:id/extend</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/memory/:id/derive</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/memory/:id/forget</div>
			<div class="api-item"><span class="method-badge method-delete">DELETE</span> /api/memory/:id</div>
		</div>
	</a>
	<a href="/api/workspaces" class="api-group">
		<div class="api-group-title">Workspaces</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/workspaces</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/workspaces</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/workspaces/:id</div>
			<div class="api-item"><span class="method-badge method-patch">PATCH</span> /api/workspaces/:id</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/workspaces/:id/archive</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/workspaces/:id/config</div>
			<div class="api-item"><span class="method-badge method-patch">PATCH</span> /api/workspaces/:id/config</div>
		</div>
	</a>
	<a href="/api/wiki" class="api-group">
		<div class="api-group-title">Wiki</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/wiki</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/wiki/:slug</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/wiki/contradictions</div>
		</div>
	</a>
	<a href="/api/surfacing" class="api-group">
		<div class="api-group-title">Surfacing</div>
		<div class="api-group-items">
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/subscriptions</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/subscriptions</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/subscriptions/:id/pause</div>
			<div class="api-item"><span class="method-badge method-post">POST</span> /api/subscriptions/:id/resume</div>
			<div class="api-item"><span class="method-badge method-delete">DELETE</span> /api/subscriptions/:id</div>
			<div class="api-item"><span class="method-badge method-get">GET</span> /api/surface/stream</div>
		</div>
	</a>
</div>

<h2>All Four Surfaces Route the Same Code</h2>

<p>
	The HTTP API, CLI, in-VM Elixir API, and Mix tasks all route through the same internal modules.
	No surface has its own code path. What <code>optimal search</code> does is exactly what
	<code>GET /api/search</code> does is exactly what <code>OptimalEngine.search/2</code> does.
</p>

<table>
	<thead>
		<tr><th>Surface</th><th>Audience</th><th>Entry point</th></tr>
	</thead>
	<tbody>
		<tr><td><code>optimal</code> CLI</td><td>Any agent runtime, any shell</td><td><code>lib/optimal_engine/cli.ex</code></td></tr>
		<tr><td>HTTP JSON API</td><td>Cross-language agents, web UIs</td><td><code>OptimalEngine.API.Router</code> on <code>:4200</code></td></tr>
		<tr><td>Elixir API</td><td>In-VM callers</td><td><code>OptimalEngine</code> public module functions</td></tr>
		<tr><td>Mix tasks</td><td>Developers</td><td><code>lib/mix/tasks/optimal.*.ex</code></td></tr>
	</tbody>
</table>

<style>
	.api-map {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin: 1.5rem 0 2rem;
	}

	.api-group {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 1rem 1.25rem;
		text-decoration: none;
		display: block;
		transition: border-color 0.15s;
	}

	.api-group:hover {
		border-color: var(--border-accent);
		text-decoration: none;
	}

	.api-group-title {
		font-weight: 600;
		color: var(--text);
		margin-bottom: 0.625rem;
		font-size: 0.9375rem;
	}

	.api-item {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--text-muted);
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.2rem 0;
	}

	@media (max-width: 640px) {
		.api-map {
			grid-template-columns: 1fr;
		}
	}
</style>
