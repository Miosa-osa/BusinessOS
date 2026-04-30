<script lang="ts">
</script>

<svelte:head>
	<title>Signal Theory — Optimal Engine</title>
</svelte:head>

<h1>Signal Theory</h1>
<p>
	The theoretical foundation the Optimal Engine is built to instantiate. Adapted from
	<em>Signal Theory: The Architecture of Optimal Intent Encoding</em> (MIOSA Research, February 2026).
</p>

<h2>Root Objective</h2>

<p>
	<strong>Maximize S/N</strong> — the ratio of actionable intent to noise in every output. S/N is not
	one metric among many; it is the measure of communication quality.
</p>

<h2>The Signal: S = (M, G, T, F, W)</h2>

<p>Every output is a Signal, modeled across five dimensions:</p>

<table>
	<thead>
		<tr>
			<th>Dim</th>
			<th>Name</th>
			<th>Question it answers</th>
			<th>Examples</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><strong>M</strong></td>
			<td>Mode</td>
			<td>How is it perceived?</td>
			<td>linguistic, visual, code, data, mixed</td>
		</tr>
		<tr>
			<td><strong>G</strong></td>
			<td>Genre</td>
			<td>What conventionalized form?</td>
			<td>spec, brief, plan, transcript, report, ADR, note</td>
		</tr>
		<tr>
			<td><strong>T</strong></td>
			<td>Type</td>
			<td>What does it DO?</td>
			<td>direct, inform, commit, decide, express</td>
		</tr>
		<tr>
			<td><strong>F</strong></td>
			<td>Format</td>
			<td>What container?</td>
			<td>markdown, code, JSON, CLI output, diff</td>
		</tr>
		<tr>
			<td><strong>W</strong></td>
			<td>Structure</td>
			<td>Internal skeleton?</td>
			<td>genre-specific template</td>
		</tr>
	</tbody>
</table>

<p>
	Before storing any non-trivial content, the engine resolves all five dimensions. Unresolved
	dimensions are flagged as noise by the <code>OptimalEngine.Classifier</code>.
</p>

<h2>Four Governing Constraints</h2>

<p>Any Signal that violates one of these fails regardless of its content.</p>

<div class="constraint-grid">
	<div class="constraint-card">
		<div class="constraint-label">Shannon — the ceiling</div>
		<p>
			Every channel has finite capacity. Don't exceed the receiver's bandwidth. A 500-line
			explanation when 20 lines suffice is a Shannon violation.
		</p>
		<div class="engine-artifact">
			Engine artifact: <code>ContextAssembler</code> — selects chunks to fit token budget, prefers
			coarsest scale that answers. Bandwidth tiers: <code>l0</code> (~100 tok) /
			<code>l1</code> (~2K tok) / <code>full</code>.
		</div>
	</div>
	<div class="constraint-card">
		<div class="constraint-label">Ashby — the repertoire</div>
		<p>
			Have enough Signal variety to handle every situation. Prose when a table is needed is an
			Ashby violation.
		</p>
		<div class="engine-artifact">
			Engine artifact: <code>OptimalEngine.Classifier</code> + <code>Composer</code> format
			variants (<code>text</code> / <code>json</code> / <code>claude</code> /
			<code>openai</code>).
		</div>
	</div>
	<div class="constraint-card">
		<div class="constraint-label">Beer — the architecture</div>
		<p>
			Maintain viable structure at every scale. A response, a file, a system — each must be
			coherently structured. Orphaned logic is a Beer violation.
		</p>
		<div class="engine-artifact">
			Engine artifact: <code>OptimalEngine.Wiki.Integrity</code> — citation + contradiction
			checker. Every claim either lives in a chunk (Tier 2) or in a wiki page that cites a chunk
			(Tier 3). No orphaned facts.
		</div>
	</div>
	<div class="constraint-card">
		<div class="constraint-label">Wiener — the feedback loop</div>
		<p>
			Never broadcast without confirmation. Close the loop — verify the receiver decoded correctly.
			Ask when ambiguous. Check that the action happened.
		</p>
		<div class="engine-artifact">
			Engine artifact: <code>OptimalEngine.Memory.Learning</code> (SICA) — corrections feed back
			into the retrieval weights.
		</div>
	</div>
</div>

<h2>Six Encoding Principles</h2>

<ol>
	<li>
		<strong>Mode-message alignment</strong> — sequential logic goes to text/code; relational logic
		goes to diagrams/tables.
	</li>
	<li>
		<strong>Genre-receiver alignment</strong> — match genre to receiver. Developers decode specs.
		Executives decode briefs. Wrong genre = failed Signal.
	</li>
	<li>
		<strong>Structure imposition</strong> — raw information is noise. Always impose structure.
		Headers, sections, genre-specific skeletons.
	</li>
	<li>
		<strong>Redundancy proportional to noise</strong> — high-stakes contexts get more structure and
		explicit intent. Simple contexts get minimal framing.
	</li>
	<li>
		<strong>Entropy preservation</strong> — maximum meaning per unit of output. No filler, no
		hedging, no padding.
	</li>
	<li>
		<strong>Bandwidth matching</strong> — match output density to receiver capacity. Three bullet
		points when that's what's needed. Full spec when that's what's needed.
	</li>
</ol>

<h2>Eleven Failure Modes</h2>

<table>
	<thead>
		<tr>
			<th>Category</th>
			<th>Failure mode</th>
			<th>Cause</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td rowspan="3">Shannon</td>
			<td>Routing failure</td>
			<td>Wrong recipient. Re-route.</td>
		</tr>
		<tr>
			<td>Bandwidth overload</td>
			<td>Too much output. Reduce, prioritize, batch.</td>
		</tr>
		<tr>
			<td>Fidelity failure</td>
			<td>Meaning lost. Re-encode with clearer structure.</td>
		</tr>
		<tr>
			<td rowspan="3">Ashby</td>
			<td>Genre mismatch</td>
			<td>Wrong form. Re-encode in correct genre.</td>
		</tr>
		<tr>
			<td>Variety failure</td>
			<td>No genre exists for this situation. Create one.</td>
		</tr>
		<tr>
			<td>Structure failure</td>
			<td>No internal skeleton. Impose genre structure.</td>
		</tr>
		<tr>
			<td rowspan="3">Beer</td>
			<td>Bridge failure</td>
			<td>No shared context. Add preamble/conventions.</td>
		</tr>
		<tr>
			<td>Herniation failure</td>
			<td>Incoherence across layers. Re-encode with proper traversal.</td>
		</tr>
		<tr>
			<td>Decay failure</td>
			<td>Outdated Signal. Audit, version, or sunset.</td>
		</tr>
		<tr>
			<td>Wiener</td>
			<td>Feedback failure</td>
			<td>No confirmation loop. Close it.</td>
		</tr>
		<tr>
			<td>Cross-cutting</td>
			<td>Adversarial noise</td>
			<td>Deliberate degradation. Make visible, escalate.</td>
		</tr>
	</tbody>
</table>

<p>
	Detected at the data layer by <code>mix optimal.health</code> and
	<code>OptimalEngine.Signal.FailureModes</code>.
</p>

<h2>How the Engine Instantiates This</h2>

<table>
	<thead>
		<tr>
			<th>Theory element</th>
			<th>Engine artifact</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Signal S=(M,G,T,F,W)</td>
			<td><code>OptimalEngine.Signal.Envelope</code> — CloudEvents + classification</td>
		</tr>
		<tr>
			<td>Classification</td>
			<td><code>OptimalEngine.Classifier</code> — auto-infers all 5 dimensions</td>
		</tr>
		<tr>
			<td>Genre routing</td>
			<td><code>OptimalEngine.Router</code> — trie-based pattern matching</td>
		</tr>
		<tr>
			<td>Failure-mode detection</td>
			<td><code>OptimalEngine.Signal.Classifier.FailureModes</code></td>
		</tr>
		<tr>
			<td>Bandwidth matching</td>
			<td><code>ContextAssembler</code> — l0 / l1 / full tiers</td>
		</tr>
		<tr>
			<td>Entropy preservation</td>
			<td>Redundancy budget enforced in <code>Indexer</code></td>
		</tr>
		<tr>
			<td>Feedback loop</td>
			<td><code>OptimalEngine.Memory.Learning</code> (SICA)</td>
		</tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/concepts/nine-stages">Nine Stages</a> — Stage 4 (Classify) instantiates S=(M,G,T,F,W)</li>
	<li><a href="/api/retrieval">Retrieval API</a> — bandwidth parameter maps to Shannon tier</li>
	<li><a href="/api/surfacing">Surfacing API</a> — contradiction category maps to Beer failure mode detection</li>
</ul>

<style>
	.constraint-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.constraint-card {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 1rem 1.25rem;
	}

	.constraint-label {
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--accent);
		margin-bottom: 0.5rem;
	}

	.constraint-card p {
		font-size: 0.875rem;
		margin-bottom: 0.75rem;
	}

	.engine-artifact {
		font-size: 0.8rem;
		color: var(--text-subtle);
		border-top: 1px solid var(--border);
		padding-top: 0.625rem;
		line-height: 1.5;
	}

	@media (max-width: 640px) {
		.constraint-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
