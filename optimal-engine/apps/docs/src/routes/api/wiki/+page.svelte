<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Wiki API — Optimal Engine</title>
</svelte:head>

<h1>Wiki API</h1>
<p>
	The wiki is Tier 3 — the LLM-maintained, audience-aware front door every agent reads first. These
	endpoints let you list, render, and audit curated pages.
</p>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/wiki</code>
	</div>
	<p>List all curated wiki pages for a workspace.</p>
	<table>
		<thead><tr><th>Param</th><th>Default</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>workspace</code></td><td>"default"</td><td>Workspace slug or ID</td></tr>
			<tr><td><code>tenant</code></td><td>"default"</td><td>Tenant ID</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>{'{ tenant_id, workspace_id, pages: [{ slug, audience, version, last_curated, curated_by, size_bytes, workspace_id }] }'}</code></p>
	<CodeBlock code={`curl 'http://localhost:4200/api/wiki?workspace=sales'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/wiki/:slug</code>
	</div>
	<p>
		Render a single wiki page. Resolves <code>{'{{cite}}'}</code>, <code>{'{{include}}'}</code>,
		and <code>{'{{expand}}'}</code> directives inline.
	</p>
	<table>
		<thead><tr><th>Param</th><th>Default</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>workspace</code></td><td>"default"</td><td>Workspace scope</td></tr>
			<tr><td><code>tenant</code></td><td>"default"</td><td>Tenant ID</td></tr>
			<tr><td><code>audience</code></td><td>"default"</td><td>Audience variant to serve</td></tr>
			<tr><td><code>format</code></td><td>"markdown"</td><td>markdown | text</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>{'{ slug, audience, version, workspace_id, body, warnings }'}</code></p>
	<CodeBlock code={`# Serve the sales variant of a pricing page
curl 'http://localhost:4200/api/wiki/healthtech-pricing?workspace=sales&audience=sales'

# Serve the exec variant
curl 'http://localhost:4200/api/wiki/healthtech-pricing?workspace=sales&audience=exec'`} lang="bash" />
</div>

<div class="endpoint-group">
	<div class="eg-header">
		<span class="method-badge method-get">GET</span>
		<code>/api/wiki/contradictions</code>
	</div>
	<p>
		Active contradiction surfacing events from the last 90 days. Contradictions are detected by the
		curator's verify gate when a new signal conflicts with an existing wiki claim.
	</p>
	<table>
		<thead><tr><th>Param</th><th>Default</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td><code>workspace</code></td><td>"default"</td><td>Workspace scope</td></tr>
		</tbody>
	</table>
	<p>Returns: <code>{'{ workspace_id, contradictions: [{ page_slug, contradictions, entities, score, detected_at }], count }'}</code></p>
	<CodeBlock code={`curl 'http://localhost:4200/api/wiki/contradictions?workspace=sales' | jq '.'`} lang="bash" />
</div>

<h2>Wiki Page Format</h2>

<p>
	Wiki pages are markdown files with YAML frontmatter. The curator maintains them automatically.
	Directives in the body are resolved at render time:
</p>

<CodeBlock code={`---
title: Q4 Pricing Strategy
audience: sales
version: 3
last_curated: 2026-04-28T15:00:00Z
curated_by: ollama/llama3.2
---

## Current State

Alice and Bob co-lead pricing negotiations. {{cite:chunk:sha256:abc123}}

The base price is $2,000/seat/year. {{cite:chunk:sha256:def456}}

## Open Items

{{include:nodes/03-sales/signals/2026-04-28-call}}

## Related

{{expand:topic:pricing}}`} lang="yaml" />

<h2>Directive Reference</h2>

<table>
	<thead>
		<tr><th>Directive</th><th>Behavior</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>{'{{cite:chunk_id}}'}</code></td>
			<td>Inline citation link back to the source chunk. Rendered as a superscript link.</td>
		</tr>
		<tr>
			<td><code>{'{{include:node/path}}'}</code></td>
			<td>Embed the contents of a node's signal at render time. Resolved from Tier 1.</td>
		</tr>
		<tr>
			<td><code>{'{{expand:topic:name}}'}</code></td>
			<td>Pull the top-5 chunks for a topic from Tier 2 and inline them as a bulleted list.</td>
		</tr>
	</tbody>
</table>

<h2>Audience Variants</h2>

<p>
	If the requested audience variant doesn't exist for a page, the engine falls back to
	<code>"default"</code>. Configure which audiences the curator maintains in
	<code>.optimal/config.yaml</code>:
</p>

<CodeBlock code={`wiki:
  audiences: [default, engineering, exec, sales, legal]`} lang="yaml" />

<h2>CLI Equivalents</h2>

<CodeBlock code={`# List pages
mix optimal.wiki list

# View a page
mix optimal.wiki view healthtech-pricing

# Force re-curate
mix optimal.wiki rebuild healthtech-pricing

# Verify all pages
mix optimal.wiki verify all`} lang="bash" />

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/three-tiers">Three Tiers</a> — how wiki pages are curated and maintained</li>
	<li><a href="/api/retrieval">RAG endpoint</a> — reads from the wiki first</li>
	<li><a href="/api/surfacing">Surfacing API</a> — wiki_updates surfacing category</li>
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
