<script lang="ts">
	import ApiSpec from '$lib/components/ApiSpec.svelte';
	import { recallEndpoints } from '$lib/data/endpoints.js';
</script>

<svelte:head>
	<title>Recall API — Optimal Engine</title>
</svelte:head>

<h1>Cued Recall</h1>
<p>
	Five typed endpoints, one per memory-failure pattern. Each builds an intent-optimized query
	internally and routes through the same <code>/api/rag</code> pipeline. The benefit:
	<code>IntentAnalyzer</code> decodes intent with maximum confidence, and the retrieval boost picks
	chunks that match the intent type directly.
</p>

<div class="recall-map">
	{#each recallEndpoints as ep, i}
		<div class="recall-card">
			<div class="recall-type">{['What happened?', 'Who owns it?', 'When is it?', 'Where is it?', "What's Alice responsible for?"][i]}</div>
			<code class="recall-path">{ep.path}</code>
		</div>
	{/each}
</div>

<p>All five return the same envelope as <code>/api/rag</code> plus a <code>recall_query</code> field showing the synthesized query string.</p>

{#each recallEndpoints as endpoint}
	<ApiSpec {endpoint} />
{/each}

<h2>See Also</h2>
<ul>
	<li><a href="/api/retrieval">Retrieval API</a> — /api/rag for open-ended questions</li>
	<li><a href="/concepts/memory-primitive">Memory Primitive</a> — the data these endpoints query</li>
</ul>

<style>
	.recall-map {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 0.625rem;
		margin-bottom: 2rem;
	}

	.recall-card {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 0.875rem;
		text-align: center;
	}

	.recall-type {
		font-size: 0.75rem;
		color: var(--text-muted);
		margin-bottom: 0.5rem;
	}

	.recall-path {
		font-size: 0.7rem;
		display: block;
		word-break: break-all;
	}

	@media (max-width: 640px) {
		.recall-map {
			grid-template-columns: 1fr;
		}
	}
</style>
