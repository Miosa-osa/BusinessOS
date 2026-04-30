<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';

	const stages = [
		{ name: 'Intake', desc: 'Receive any input, hash for deduplication, record provenance' },
		{ name: 'Parse', desc: 'Convert any format to text + structural metadata. 10 backends.' },
		{ name: 'Decompose', desc: 'Large → small hierarchical chunks at 4 fixed scales' },
		{ name: 'Classify', desc: 'S=(M,G,T,F,W) + intent enum per chunk at every scale' },
		{ name: 'Embed', desc: 'Multi-modal aligned 768-dim vectors (text, image, audio, code)' },
		{ name: 'Route', desc: 'Assign each chunk to primary node + cross-reference nodes' },
		{ name: 'Store', desc: 'Atomic SQLite commit. Emit store.chunk.indexed event.' },
		{ name: 'Cluster', desc: 'Incremental HDBSCAN over weighted feature vectors. Theme naming.' },
		{ name: 'Curate', desc: 'Ollama rewrites affected wiki pages. Verify gate. Commit or reject.' }
	];
</script>

<svelte:head>
	<title>Nine Stages — Optimal Engine</title>
</svelte:head>

<h1>Nine Stages</h1>
<p>
	Every signal flows through the same nine stages in strict order. Each stage has one
	responsibility and a typed contract with the next. No stage is skipped. No shortcuts.
</p>

<div class="pipeline-diagram">
	{#each stages as stage, i}
		<div class="stage-card">
			<div class="stage-num">{i + 1}</div>
			<div class="stage-body">
				<div class="stage-name">{stage.name}</div>
				<div class="stage-desc">{stage.desc}</div>
			</div>
		</div>
		{#if i < stages.length - 1}
			<div class="stage-arrow">↓</div>
		{/if}
	{/each}
</div>

<h2>Stage 1 — Intake</h2>

<p>
	Receives signals from any surface: CLI, HTTP, filesystem watcher, or programmatic call. Hashes
	the payload (SHA-256 content address) for deduplication. Records provenance: origin, format hint,
	received_at.
</p>

<p>
	<strong>Idempotent:</strong> a signal with the same <code>content_hash</code> that has already
	been ingested is rejected without error. Re-ingesting a workspace is safe.
</p>

<CodeBlock code={`# Ingest a single file
mix optimal.ingest --file ./notes/2026-04-28-pricing-call.md

# Ingest a workspace (walk all nodes)
mix optimal.ingest_workspace ~/company-brain/sales`} lang="bash" />

<h2>Stage 2 — Parse</h2>

<p>
	Converts any input format into plain text + structural metadata. Preserves every boundary the
	source exposes: headings, pages, slides, timestamps, code blocks. Stage 3 depends on these
	boundaries.
</p>

<table>
	<thead>
		<tr><th>Format</th><th>Backend</th><th>Modality</th></tr>
	</thead>
	<tbody>
		<tr><td><code>.md .txt .rst .adoc</code></td><td>Native</td><td>text</td></tr>
		<tr><td><code>.yaml .yml .toml .json</code></td><td>yaml_elixir / Jason</td><td>data</td></tr>
		<tr><td><code>.csv .tsv</code></td><td>NimbleCSV</td><td>data</td></tr>
		<tr><td><code>.html</code></td><td>Floki</td><td>text</td></tr>
		<tr><td>Source code (30+ extensions)</td><td>tree-sitter or native</td><td>code</td></tr>
		<tr><td><code>.pdf</code></td><td>pdftotext shell or pdf_extract</td><td>text</td></tr>
		<tr><td><code>.docx .pptx .xlsx</code></td><td>zip + OOXML parser</td><td>text</td></tr>
		<tr><td><code>.png .jpg .jpeg .gif .webp</code></td><td>tesseract OCR</td><td>image+text</td></tr>
		<tr><td><code>.mp3 .wav .m4a .ogg .flac</code></td><td>whisper.cpp</td><td>audio+text</td></tr>
		<tr><td><code>.mp4 .mov .webm</code></td><td>ffmpeg → image + audio backends</td><td>video+text</td></tr>
	</tbody>
</table>

<h2>Stage 3 — Decompose</h2>

<p>
	Breaks parsed documents into a hierarchical chunk tree at four fixed scales. Parent-child links
	are maintained. The decomposer never splits across structural boundaries, never splits a sentence
	mid-word, and respects top-level function/class boundaries in code.
</p>

<table>
	<thead>
		<tr><th>Scale</th><th>Target size</th><th>Created for</th></tr>
	</thead>
	<tbody>
		<tr><td><code>:document</code></td><td>Full source</td><td>Every signal</td></tr>
		<tr><td><code>:section</code></td><td>Structural unit</td><td>Every heading / page / slide boundary</td></tr>
		<tr><td><code>:paragraph</code></td><td>Semantic block</td><td>Every paragraph, code block, table row group</td></tr>
		<tr><td><code>:chunk</code></td><td>~512 tokens</td><td>Sliding window, 64-token overlap, respects paragraph boundaries</td></tr>
	</tbody>
</table>

<h2>Stage 4 — Classify</h2>

<p>
	For every chunk at every scale: determines S=(M,G,T,F,W) signal dimensions and extracts intent.
	Heuristics-first; Ollama augments confidence when available.
</p>

<p>The ten-value intent enum:</p>

<table>
	<thead>
		<tr><th>Intent</th><th>Meaning</th></tr>
	</thead>
	<tbody>
		<tr><td><code>request_info</code></td><td>Asking for something</td></tr>
		<tr><td><code>propose_decision</code></td><td>Putting a decision on the table</td></tr>
		<tr><td><code>record_fact</code></td><td>Stating something as ground truth</td></tr>
		<tr><td><code>express_concern</code></td><td>Flagging risk or blocker</td></tr>
		<tr><td><code>commit_action</code></td><td>Taking on a task</td></tr>
		<tr><td><code>reference</code></td><td>Pointing at other context</td></tr>
		<tr><td><code>narrate</code></td><td>Describing a sequence of events</td></tr>
		<tr><td><code>reflect</code></td><td>Analyzing past signals</td></tr>
		<tr><td><code>specify</code></td><td>Defining a contract or requirement</td></tr>
		<tr><td><code>measure</code></td><td>Reporting a metric or quantity</td></tr>
	</tbody>
</table>

<h2>Stage 5 — Embed</h2>

<p>
	Projects every chunk into a shared 768-dim aligned vector space. The alignment invariant: a text
	query embedding can retrieve an image chunk embedding because both live in the nomic 768-dim
	space. One retriever, not three.
</p>

<table>
	<thead>
		<tr><th>Modality</th><th>Model</th><th>Dimensions</th></tr>
	</thead>
	<tbody>
		<tr><td>text</td><td>nomic-embed-text-v1.5</td><td>768</td></tr>
		<tr><td>image</td><td>nomic-embed-vision-v1.5</td><td>768 (aligned with text)</td></tr>
		<tr><td>audio</td><td>whisper.cpp → text embed</td><td>768</td></tr>
		<tr><td>code</td><td>nomic-embed-text-v1.5</td><td>768</td></tr>
	</tbody>
</table>

<h2>Stage 6 — Route</h2>

<p>
	Assigns each chunk to one primary node + N cross-reference nodes based on entities, keywords, and
	topology rules. The router uses trie-based pattern matching against the workspace's node tree.
</p>

<h2>Stage 7 — Store</h2>

<p>
	Persists everything produced by stages 1–6 in a single atomic SQLite transaction. If Stage 7
	fails, stages 1–6 produced nothing visible. After commit, emits a
	<code>store.chunk.indexed</code> event that triggers Stages 8 and 9 asynchronously.
</p>

<h2>Stage 8 — Cluster</h2>

<p>
	HDBSCAN over a weighted feature vector. Runs incrementally per new chunk — never a full rebuild
	except via <code>mix optimal.cluster.rebuild</code>. Theme names are auto-generated by Ollama
	over the top-N chunks in each cluster.
</p>

<CodeBlock code={`feature = 0.6 × embedding
        + 0.2 × entity_overlap
        + 0.15 × intent_match
        + 0.05 × node_affinity`} lang="text" />

<h2>Stage 9 — Curate</h2>

<p>
	On each <code>store.chunk.indexed</code> event: computes which wiki pages cite entities or
	clusters the new chunk belongs to, enqueues curation jobs, runs the Ollama curator loop, then
	passes a verify gate before committing.
</p>

<p>The verify gate is fail-closed: reject and flag on any citation that does not resolve.</p>

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/three-tiers">Three Tiers</a> — what the pipeline writes to</li>
	<li><a href="/concepts/signal-theory">Signal Theory</a> — S=(M,G,T,F,W) classification</li>
	<li><a href="/api/retrieval">Retrieval API</a> — how agents read from the result</li>
</ul>

<style>
	.pipeline-diagram {
		display: flex;
		flex-direction: column;
		margin: 1.5rem 0 2rem;
	}

	.stage-card {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 0.875rem 1rem;
	}

	.stage-num {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--accent);
		background: var(--accent-dim);
		border-radius: var(--r-sm);
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stage-name {
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--text);
		margin-bottom: 0.125rem;
	}

	.stage-desc {
		font-size: 0.8125rem;
		color: var(--text-muted);
	}

	.stage-arrow {
		text-align: left;
		color: var(--border);
		font-size: 1rem;
		line-height: 1;
		padding: 3px 0 3px 1.75rem;
	}
</style>
