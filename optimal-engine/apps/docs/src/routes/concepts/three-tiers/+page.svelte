<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Three Tiers — Optimal Engine</title>
</svelte:head>

<h1>Three Tiers</h1>
<p>
	Every piece of knowledge in Optimal Engine lives at exactly one tier. Retrieval always tries Tier
	3 first. The architecture is a strict hierarchy: Tier 1 is ground truth, Tier 2 is the machine's
	derivation of it, and Tier 3 is the LLM's curated front door.
</p>

<div class="tier-diagram">
	<div class="tier tier-3">
		<div class="tier-label">Tier 3 — The Wiki</div>
		<div class="tier-body">
			<p>LLM-maintained. Read first. Audience-aware.</p>
			<p>Path: <code>.wiki/</code></p>
			<p>Hot citations + executable directives. The agent's front door. Curated by Ollama on every ingest. Humans write the schema; the curator writes the pages.</p>
		</div>
	</div>
	<div class="tier-arrow">▲ CURATE ▼</div>
	<div class="tier tier-2">
		<div class="tier-label">Tier 2 — Derivatives</div>
		<div class="tier-body">
			<p>Machine-maintained. Rebuildable.</p>
			<p>Path: <code>.optimal/index.db</code></p>
			<p>SQLite + FTS5 + vectors + graph + clusters + L0 abstracts. Produced by the 9-stage pipeline. <code>mix optimal.rebuild</code> recreates it from Tier 1 exactly.</p>
		</div>
	</div>
	<div class="tier-arrow">▲ DERIVE ▼</div>
	<div class="tier tier-1">
		<div class="tier-label">Tier 1 — Raw Sources</div>
		<div class="tier-body">
			<p>Immutable. Append-only. Hash-addressed.</p>
			<p>Path: <code>nodes/**/signals/*.md</code>, <code>assets/</code></p>
			<p>Signal files, PDFs, images, audio, video. The engine NEVER rewrites them.</p>
		</div>
	</div>
</div>

<h2>Four Invariants</h2>

<p>Violate any one of these and the engine rots.</p>

<table>
	<thead>
		<tr>
			<th>Invariant</th>
			<th>Rule</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><strong>Tier 1 append-only</strong></td>
			<td>Every write is a new file at a new path. Nothing gets edited in place.</td>
		</tr>
		<tr>
			<td><strong>Tier 2 rebuildable</strong></td>
			<td>
				<code>mix optimal.rebuild</code> reconstructs the entire derivative index from Tier 1 exactly.
			</td>
		</tr>
		<tr>
			<td><strong>Tier 3 LLM-owned</strong></td>
			<td>Humans write the schema; the curator writes the pages. Versioned.</td>
		</tr>
		<tr>
			<td><strong>Citations downward-only</strong></td>
			<td>T3 → T2/T1. T2 → T1. Never upward. Acyclic. Auditable.</td>
		</tr>
	</tbody>
</table>

<h2>Why This Matters</h2>

<p>
	Classical RAG systems have a single index of chunks. Every query re-discovers the same facts
	through cosine similarity. There is no persistent understanding — only a lookup table.
</p>

<p>
	The three-tier architecture changes the economics of retrieval. Tier 3 (the wiki) absorbs the
	cost of understanding <em>once</em> — on ingest, not on query. When an agent asks "who owns the
	pricing negotiation?", the engine checks the wiki first. If the wiki page exists and is current,
	the answer is returned immediately, zero retriever hits. Only unanswered queries fall through to
	hybrid retrieval against Tier 2.
</p>

<p>
	<strong>Wiki-first, not chunk-and-pray. Curated memory beats infinite memory.</strong>
</p>

<h2>How Tier 3 Stays Current</h2>

<p>
	Every time Stage 7 (Store) commits a new chunk, it emits a <code>store.chunk.indexed</code>
	event. Stage 9 (Curator) receives the event, computes which wiki pages cite entities or clusters
	the new chunk belongs to, and enqueues curation jobs. The Ollama-driven curator rewrites the
	affected pages with new citations, then passes a verify gate:
</p>

<ul>
	<li>Every claim in the page cites a real chunk.</li>
	<li>Every citation resolves to an existing chunk ID.</li>
	<li>Schema rules pass (page structure matches <code>.wiki/SCHEMA.md</code>).</li>
	<li>No contradictions are silently swallowed — they are flagged and surfaced.</li>
</ul>

<p>
	The gate is fail-closed: a page that fails verification is rejected and flagged, not silently
	committed.
</p>

<h2>Storage Schema</h2>

<p>Tier 2 is a single SQLite database at <code>.optimal/index.db</code> with these tables:</p>

<table>
	<thead>
		<tr>
			<th>Table</th>
			<th>Purpose</th>
		</tr>
	</thead>
	<tbody>
		<tr><td><code>signals</code></td><td>Tier-1 metadata: source path, hash, received_at, size</td></tr>
		<tr><td><code>chunks</code></td><td>Hierarchical decomposition, one row per chunk at every scale</td></tr>
		<tr><td><code>classifications</code></td><td>Per-chunk S=(M,G,T,F,W) + intent + confidence</td></tr>
		<tr><td><code>embeddings</code></td><td>Per-chunk 768-dim vectors, modality-tagged</td></tr>
		<tr><td><code>entities</code></td><td>Extracted entities per chunk</td></tr>
		<tr><td><code>edges</code></td><td>Typed relations in the knowledge graph</td></tr>
		<tr><td><code>clusters</code></td><td>HDBSCAN theme groupings</td></tr>
		<tr><td><code>cluster_members</code></td><td>chunk_id ↔ cluster_id with membership weight</td></tr>
		<tr><td><code>assets</code></td><td>Binary blobs (images, audio, PDFs) by hash</td></tr>
		<tr><td><code>wiki_pages</code></td><td>Tier-3 curated pages with frontmatter + body</td></tr>
		<tr><td><code>citations</code></td><td>wiki_page_id → chunk_id with claim_hash</td></tr>
		<tr><td><code>events</code></td><td>Append-only log of pipeline stage transitions</td></tr>
	</tbody>
</table>

<h2>Rebuilding</h2>

<CodeBlock code={`# Rebuild Tier 2 from Tier 1 (non-destructive, idempotent)
mix optimal.rebuild

# Verify integrity after rebuild
mix optimal.verify --sample 100`} lang="bash" />

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/nine-stages">Nine Stages</a> — how signals flow from Tier 1 to Tier 3</li>
	<li><a href="/api/wiki">Wiki API</a> — read, verify, and list curated pages</li>
	<li><a href="/concepts/workspaces">Workspaces</a> — how each workspace gets its own tier set</li>
</ul>

<style>
	.tier-diagram {
		display: flex;
		flex-direction: column;
		gap: 0;
		margin: 1.5rem 0 2rem;
	}

	.tier {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 1rem 1.25rem;
	}

	.tier-3 {
		border-color: var(--accent);
		border-left-width: 3px;
	}

	.tier-label {
		font-weight: 700;
		font-size: 0.875rem;
		color: var(--text);
		margin-bottom: 0.5rem;
	}

	.tier-body p {
		font-size: 0.8125rem;
		margin-bottom: 0.25rem;
	}

	.tier-body p:last-child {
		margin-bottom: 0;
	}

	.tier-arrow {
		text-align: center;
		color: var(--text-subtle);
		font-size: 0.75rem;
		font-family: var(--font-mono);
		padding: 0.375rem 0;
		letter-spacing: 0.05em;
	}
</style>
