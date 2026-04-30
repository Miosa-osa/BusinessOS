<script lang="ts">
	import ApiSpec from '$lib/components/ApiSpec.svelte';
	import { retrievalEndpoints } from '$lib/data/endpoints.js';
</script>

<svelte:head>
	<title>Retrieval API — Optimal Engine</title>
</svelte:head>

<h1>Retrieval</h1>
<p>
	Five retrieval surfaces. Choosing the right one changes both latency and answer quality.
</p>

<h2>Decision Tree</h2>

<div class="decision-tree">
	<div class="dt-row">
		<div class="dt-q">Asking an open-ended question about organizational knowledge?</div>
		<div class="dt-a"><a href="#rag">→ use <code>/api/rag</code></a></div>
	</div>
	<div class="dt-row">
		<div class="dt-q">Want to find specific documents or signals?</div>
		<div class="dt-a"><a href="#search">→ use <code>/api/search</code></a></div>
	</div>
	<div class="dt-row">
		<div class="dt-q">Inspecting chunks at a specific scale, intent, or modality?</div>
		<div class="dt-a"><a href="#grep">→ use <code>/api/grep</code></a></div>
	</div>
	<div class="dt-row">
		<div class="dt-q">Typed memory-failure query (who/when/where/owns)?</div>
		<div class="dt-a"><a href="/api/recall">→ use <code>/api/recall/:type</code></a></div>
	</div>
	<div class="dt-row">
		<div class="dt-q">Seeding an agent system prompt with full workspace context?</div>
		<div class="dt-a"><a href="#profile">→ use <code>/api/profile</code></a></div>
	</div>
</div>

<h2>Retrieval Method Comparison</h2>

<table>
	<thead>
		<tr>
			<th></th>
			<th>ask (/rag)</th>
			<th>search</th>
			<th>grep</th>
			<th>recall</th>
			<th>profile</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Unit returned</td>
			<td>Answer + sources</td>
			<td>Signals (metadata)</td>
			<td>Chunks (with trace)</td>
			<td>Answer + sources</td>
			<td>4-tier snapshot</td>
		</tr>
		<tr>
			<td>Wiki-first</td>
			<td>Yes</td>
			<td>No</td>
			<td>No</td>
			<td>Yes</td>
			<td>Yes (curated tier)</td>
		</tr>
		<tr>
			<td>Intent-optimized</td>
			<td>Yes (general)</td>
			<td>No</td>
			<td>Filter only</td>
			<td>Yes (typed)</td>
			<td>No</td>
		</tr>
		<tr>
			<td>Audience-aware</td>
			<td>Yes</td>
			<td>No</td>
			<td>No</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
	</tbody>
</table>

<div id="rag"></div>
<div id="search"></div>
<div id="grep"></div>
<div id="profile"></div>

{#each retrievalEndpoints as endpoint}
	<ApiSpec {endpoint} />
{/each}

<h2>Format Options</h2>

<table>
	<thead>
		<tr><th>format</th><th>What you get</th></tr>
	</thead>
	<tbody>
		<tr><td><code>markdown</code></td><td>Answer in markdown with inline citations</td></tr>
		<tr><td><code>text</code></td><td>Plain text, no markup</td></tr>
		<tr><td><code>claude</code></td><td>System prompt string ready to pass as <code>system:</code></td></tr>
		<tr><td><code>openai</code></td><td><code>messages</code> array ready for OpenAI <code>/v1/chat/completions</code></td></tr>
		<tr><td><code>json</code></td><td>Structured envelope with <code>answer</code>, <code>sources</code>, <code>wiki_hit</code></td></tr>
	</tbody>
</table>

<h2>Bandwidth Options</h2>

<table>
	<thead>
		<tr><th>bandwidth</th><th>Token budget</th><th>When to use</th></tr>
	</thead>
	<tbody>
		<tr><td><code>l0</code></td><td>~100 tokens</td><td>One-liner summary; tight prompt budgets; mobile</td></tr>
		<tr><td><code>medium</code> / <code>l1</code></td><td>~2,000 tokens</td><td>Standard agent turn; most use cases</td></tr>
		<tr><td><code>full</code></td><td>Up to token limit</td><td>Deep analysis; multi-step reasoning; RAG with full doc</td></tr>
	</tbody>
</table>

<p>Default to <code>medium</code>. Only use <code>full</code> when the query requires reasoning over complete documents.</p>

<h2>See Also</h2>
<ul>
	<li><a href="/api/recall">Recall API</a> — five typed recall endpoints for specific memory-failure patterns</li>
	<li><a href="/concepts/three-tiers">Three Tiers</a> — why the wiki is tried first</li>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — typed wrapper over these endpoints</li>
</ul>

<style>
	.decision-tree {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		overflow: hidden;
		margin-bottom: 2rem;
	}

	.dt-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1.25rem;
		border-bottom: 1px solid var(--border);
		gap: 1rem;
	}

	.dt-row:last-child {
		border-bottom: none;
	}

	.dt-q {
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.dt-a {
		font-size: 0.875rem;
		white-space: nowrap;
	}
</style>
