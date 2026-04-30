<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';

	const quickExample = `curl -X POST http://localhost:4200/api/rag \\
  -H 'Content-Type: application/json' \\
  -d '{"query":"who owns the pricing negotiation?","workspace":"sales","format":"claude"}'`;
</script>

<svelte:head>
	<title>Optimal Engine — Docs</title>
</svelte:head>

<div class="landing">
	<div class="hero">
		<div class="hero-eyebrow">Open-source · MIT · Self-hosted</div>
		<h1>The second brain of a company.</h1>
		<p class="hero-lead">
			Optimal Engine ingests every signal flowing through your organization — chat, email, docs,
			meetings, tickets, code, voice — curates it into an LLM-maintained wiki, and delivers
			permission-scoped context to any agent runtime in under 200ms.
		</p>
		<div class="hero-actions">
			<a href="/quickstart" class="pill pill-primary">Get started</a>
			<a href="/concepts/three-tiers" class="pill pill-ghost">How it works</a>
			<a href="/api" class="pill pill-ghost">API reference</a>
		</div>
	</div>

	<div class="pitch-grid">
		<div class="pitch-card">
			<div class="pitch-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
				</svg>
			</div>
			<h3>Wiki-first, not chunk-and-pray</h3>
			<p>
				Classical RAG re-discovers the same facts on every query. Optimal Engine discovers once,
				curates forever. Most agent queries are answered from the LLM-maintained wiki — zero
				retriever hits.
			</p>
		</div>
		<div class="pitch-card">
			<div class="pitch-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<circle cx="12" cy="12" r="3"/>
					<path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/>
				</svg>
			</div>
			<h3>Signal integrity — S=(M,G,T,F,W)</h3>
			<p>
				Every chunk is classified across five dimensions: Mode, Genre, Type, Format, Structure.
				Intent is extracted at every scale. Nothing is stored without provenance.
			</p>
		</div>
		<div class="pitch-card">
			<div class="pitch-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<rect x="2" y="3" width="7" height="9" rx="1"/>
					<rect x="15" y="3" width="7" height="5" rx="1"/>
					<rect x="15" y="12" width="7" height="9" rx="1"/>
					<rect x="2" y="16" width="7" height="5" rx="1"/>
				</svg>
			</div>
			<h3>Workspace isolation</h3>
			<p>
				Sales never bleeds into Engineering. M&A stays private. Each workspace has its own node
				tree, vector index, and wiki. Audience tags let one workspace serve multiple receiver types.
			</p>
		</div>
		<div class="pitch-card">
			<div class="pitch-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
				</svg>
			</div>
			<h3>Proactive surfacing</h3>
			<p>
				Subscribe to any workspace, topic, node, or entity. The engine pushes relevant changes via
				SSE across 14 categories — contradictions, blockers, ownership changes, new specifications,
				and more.
			</p>
		</div>
	</div>

	<div class="quick-section">
		<h2>Five minutes to running</h2>
		<CodeBlock code={`git clone https://github.com/Miosa-osa/OptimalEngine.git
cd OptimalEngine
make install       # deps + compile
make bootstrap     # migrate + ingest sample-workspace/`} lang="bash" />

		<h2>First query</h2>
		<CodeBlock code={quickExample} lang="bash" />
		<p>
			<a href="/quickstart">Full quickstart guide &rarr;</a>
		</p>
	</div>

	<div class="nav-grid">
		<a href="/concepts/three-tiers" class="nav-card">
			<div class="nav-card-title">Three Tiers</div>
			<div class="nav-card-desc">Raw sources → derivatives → LLM-maintained wiki. Four invariants that keep the engine coherent as it grows.</div>
		</a>
		<a href="/concepts/nine-stages" class="nav-card">
			<div class="nav-card-title">Nine Stages</div>
			<div class="nav-card-desc">Every signal through Intake → Parse → Decompose → Classify → Embed → Route → Store → Cluster → Curate.</div>
		</a>
		<a href="/api/retrieval" class="nav-card">
			<div class="nav-card-title">Retrieval API</div>
			<div class="nav-card-desc">/api/rag, /api/search, /api/grep, /api/profile — the full retrieval surface with decision guide.</div>
		</a>
		<a href="/api/memory" class="nav-card">
			<div class="nav-card-title">Memory Primitive</div>
			<div class="nav-card-desc">Versioned, relation-tracked, audience-scoped facts. Five relation types, soft delete, citation URIs.</div>
		</a>
		<a href="/sdks/typescript" class="nav-card">
			<div class="nav-card-title">TypeScript SDK</div>
			<div class="nav-card-desc">npm install + Vercel AI SDK adapter. Drop-in organizational memory for any Next.js or Node app.</div>
		</a>
		<a href="/self-host" class="nav-card">
			<div class="nav-card-title">Self-host</div>
			<div class="nav-card-desc">Docker Compose, environment variables, production config. Runs on a laptop. MIT licensed.</div>
		</a>
	</div>
</div>

<style>
	.landing {
		padding-bottom: 4rem;
	}

	.hero {
		padding: 3rem 0 2.5rem;
		border-bottom: 1px solid var(--border);
		margin-bottom: 2.5rem;
	}

	.hero-eyebrow {
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--accent);
		margin-bottom: 1rem;
	}

	.hero h1 {
		font-size: 2.5rem;
		font-weight: 700;
		margin-bottom: 1rem;
		letter-spacing: -0.03em;
	}

	.hero-lead {
		font-size: 1.0625rem;
		color: var(--text-muted);
		max-width: 600px;
		line-height: 1.7;
		margin-bottom: 1.75rem;
	}

	.hero-actions {
		display: flex;
		gap: 0.625rem;
		flex-wrap: wrap;
	}

	/* Pitch cards */
	.pitch-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.pitch-card {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 1.25rem;
	}

	.pitch-icon {
		color: var(--accent);
		margin-bottom: 0.75rem;
	}

	.pitch-card h3 {
		font-size: 0.9375rem;
		margin-top: 0;
		margin-bottom: 0.5rem;
	}

	.pitch-card p {
		font-size: 0.875rem;
		margin-bottom: 0;
	}

	/* Quick section */
	.quick-section {
		margin-bottom: 3rem;
	}

	.quick-section h2 {
		margin-top: 0;
	}

	/* Nav grid */
	.nav-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 0.75rem;
	}

	.nav-card {
		display: block;
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 1rem 1.25rem;
		text-decoration: none;
		transition: border-color 0.15s, background 0.15s;
	}

	.nav-card:hover {
		border-color: var(--border-accent);
		background: var(--bg-hover);
		text-decoration: none;
	}

	.nav-card-title {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--text);
		margin-bottom: 0.375rem;
	}

	.nav-card-desc {
		font-size: 0.8125rem;
		color: var(--text-muted);
		line-height: 1.5;
	}

	@media (max-width: 640px) {
		.pitch-grid,
		.nav-grid {
			grid-template-columns: 1fr;
		}

		.hero h1 {
			font-size: 1.75rem;
		}
	}
</style>
