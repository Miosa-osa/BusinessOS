# The Wiki Layer

> A self-maintaining, citation-rich, incrementally-compacted top layer that
> sits over immutable raw sources and makes context a first-class artifact of
> the Optimal Engine — not a byproduct of retrieval.

## The thesis

Classical RAG re-discovers the same facts on every query. It's lossy (the
retriever only sees what matches a similarity score) and amnesic (what the LLM
figured out last time has to be rediscovered this time).

The Wiki Layer flips this. Every time new information arrives, a curation step
**integrates it into a persistent, structured, LLM-maintained wiki**. Agents
read the wiki FIRST. The wiki is a map; it carries "hot citations" — explicit
URIs — back to the immutable sources whenever the agent needs to zoom in.

Retrieval still exists, but it's scoped to what the wiki couldn't already
answer.

## The three tiers

```
┌──────────────────────────────────────────────────────────────┐
│  TIER 3 — THE WIKI (LLM-maintained, read-first)              │
│  Top-level, curated, cross-referenced, always-loaded.        │
│  Every fact carries a citation URI back to Tier 1.           │
│  Can contain executable directives (see below).              │
└──────────────────────────────────────────────────────────────┘
                          ▲  ▼
┌──────────────────────────────────────────────────────────────┐
│  TIER 2 — DERIVATIVES (machine-maintained, cheap to rebuild) │
│  Embeddings, FTS index, graph edges, clusters, L0 abstracts. │
│  Rebuildable from Tier 1 without loss.                       │
└──────────────────────────────────────────────────────────────┘
                          ▲  ▼
┌──────────────────────────────────────────────────────────────┐
│  TIER 1 — RAW SOURCES (immutable, append-only)               │
│  The signal files, PDFs, images, audio, transcripts.         │
│  The engine NEVER rewrites these. Only appends new ones.     │
└──────────────────────────────────────────────────────────────┘
```

Hard invariant: **the LLM owns Tier 3 and does not touch Tier 1**.

## Why three tiers (not two)

Tier 2 exists because a wiki that tried to embed all the raw text would
collapse under its own weight. Tier 2 (derivatives) is the fast-path for
retrieval; Tier 3 (wiki) is the curated read-through layer that sits on top of
BOTH raw sources and derivatives. When the wiki cites a fact, the citation can
point at any tier.

## Anatomy of a Wiki page

A wiki page is a markdown file at `.wiki/<slug>.md` in the engine's root. It
has three sections:

```markdown
---
title: Alice — pricing conversations
last_curated: 2026-04-17T14:30:00Z
curated_by: ollama:qwen3:8b
source_count: 7
---

# Alice — pricing conversations

## Summary

Ed has been negotiating on $2K/seat pricing since 2026-03-18. He has asked
three times for a discount {{cite: optimal://nodes/ai-masters/signals/2026-03-18-ed-pricing-call.md}}
{{cite: optimal://nodes/ai-masters/signals/2026-03-22-ed-followup.md}}
{{cite: optimal://nodes/ai-masters/signals/2026-04-03-ed-pricing-pushback.md}}.

Alice's position: $2K is the floor for AI Masters
{{cite: optimal://nodes/04-ai-masters/context.md}}.

## Open threads

- Discount request pending Alice's response {{expand: ed-counter-offer-options}}
- Waiting on partnership terms {{include: optimal://nodes/04-ai-masters/deliverables/external/ed-partnership/offer-stack.md}}

## Related

- [AI Masters offer stack](ai-masters-offer-stack.md)
- [Partnership structure](partnership-structure.md)
```

Three things matter here:

1. **Every factual claim has a `{{cite: uri}}` directive.** Agents see exactly
   where each sentence came from. Claims without citations are flagged by the
   maintenance loop and either grounded or removed.

2. **Executable directives pull more context on demand.** The agent (or the
   engine) resolves `{{include: uri}}` inline and `{{expand: topic}}` as a
   sub-query. Progressive disclosure — don't load what the agent doesn't ask
   for.

3. **Cross-links point to other wiki pages.** The wiki is a graph, not a flat
   bag. Following links is cheaper than re-querying.

## Executable directives

All directives are `{{verb: argument}}` and are resolved by the engine OR by
an agent with access to the engine.

| Directive                          | Resolves to                                                    |
|-----------------------------------|----------------------------------------------------------------|
| `{{cite: optimal://...}}`         | URI pointer. Rendered as an inline footnote with clickable link. Cheap — no content load. |
| `{{include: optimal://...}}`      | Inline the referenced content at the cite location. Loads at `:l1` tier by default; `{{include: ... tier=full}}` for full. |
| `{{expand: topic-slug}}`          | Run a sub-query against the wiki for `topic-slug`. Returns the summary section of the matching page. |
| `{{search: "query"}}`             | Run hybrid search, inject top-k chunks. Useful for truly dynamic lookups. |
| `{{table: uri#column=value}}`     | Fetch a structured row from a CSV/table asset. |
| `{{trace: entity}}`               | Walk the knowledge graph from `entity` and inject neighbors. |
| `{{recent: node limit=5}}`        | Inject the 5 most recent signals from a node. |

The engine rejects any directive verb not in this whitelist, so agents can't
be tricked into executing arbitrary code from a wiki page.

## Citation integrity

The maintenance loop enforces:

1. **Every factual sentence cites at least one source.** If it can't find a
   source, the claim is either removed or moved to a `## Unverified` section.
2. **Every `{{cite: uri}}` resolves to a real context.** Broken citations are
   flagged and repaired (redirect to current location if renamed) or removed.
3. **Citation recency.** Each citation carries a `last_verified` timestamp.
   When a source is re-ingested, the citation is re-verified.
4. **Conflict detection.** If two cited sources contradict each other, the
   page gets a `## Contradictions` section flagging the divergence. The agent
   (or a human) decides how to resolve.

## The maintenance loop

```
┌──────────────────────────────────────────────────────────────┐
│ NEW SIGNAL INGESTED → pipeline runs (stages 1–8)             │
└──────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ ROUTE TO WIKI PAGES                                          │
│ For each affected topic, queue a curation job.               │
│ Affected topics = clusters the new chunks landed in +        │
│ entities mentioned + pages explicitly cited.                 │
└──────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ CURATE (Ollama local LLM)                                    │
│   Input:  existing wiki page + new signals + their citations │
│   Prompt: integrate new facts; keep citations; flag contras- │
│           dictions; preserve structure; compact redundancy.  │
│   Output: updated wiki page with full citation coverage.     │
└──────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ VERIFY → COMMIT                                              │
│ - Integrity checks (every claim cited, no broken URIs).      │
│ - Diff against previous version, store in page history.      │
│ - Emit a Signal on the "wiki.page.updated" topic.            │
└──────────────────────────────────────────────────────────────┘
```

Maintenance is **triggered, not scheduled**. A new signal into the `04-ai-masters`
node with the entity "Alice" queues a curation job for exactly the wiki
pages that mention Alice. No blanket rebuild.

Hard rebuild (`mix optimal.wiki.rebuild`) exists for disaster recovery: it
re-curates every page from Tier 1 + Tier 2. Idempotent, takes minutes, not
seconds.

## The schema

Wiki behavior is governed by `.wiki/SCHEMA.md`, a markdown-formatted
declaration the curator LLM is prompted with. It specifies:

- Allowed top-level sections (`## Summary`, `## Open threads`, `## Decisions`, `## Related`, etc.)
- Citation density minimums (e.g., "every claim in `## Decisions` must cite a `decision-log` signal")
- Naming conventions for page slugs (kebab-case, entity-first when applicable)
- When to create a new page vs extend an existing one
- Page size ceilings (e.g., "sections beyond 500 tokens should spawn a child page")
- Allowed directive verbs (from the table above)

The schema is a human-owned, LLM-read contract that makes the curator's
behavior predictable and auditable.

## How the Wiki Layer maps to the 8-stage pipeline

| Stage            | Wiki Layer's role                                                             |
|------------------|-------------------------------------------------------------------------------|
| 1 Intake         | wiki receives no input directly — always through the pipeline                 |
| 2 Parse          | same                                                                          |
| 3 Decompose      | wiki pages are themselves decomposed (their chunks feed retrieval like any source) |
| 4 Classify       | wiki pages are classified as `genre: wiki_page, intent: reference`            |
| 5 Embed          | wiki pages are embedded so they can be retrieved by semantic search           |
| 6 Cluster        | wiki pages tend to BE cluster summaries (or get co-clustered with their sources) |
| 7 Store          | wiki pages are stored in `.wiki/` (Tier 1 wrt themselves, Tier 3 wrt sources) |
| 8 Deliver        | **wiki is always-loaded as the "front door."** Assembly reads wiki first, retrieves raw sources only for sub-queries the wiki doesn't answer |
| 9 Curate (new)   | the maintenance loop                                                          |

## Build status

| Component                   | Status                                                  |
|----------------------------|---------------------------------------------------------|
| `.wiki/` directory          | ❌ does not exist                                        |
| Wiki schema                 | ❌ not written                                           |
| Wiki page template          | ❌ not defined                                           |
| Directive parser            | ❌ not built (no `{{cite: ...}}` handler)                |
| Citation integrity checker  | ❌ not built                                             |
| Curation trigger (signal → affected pages)  | ❌ not built                             |
| Curator (Ollama prompting)  | ❌ not built                                             |
| Page history / diff storage | ❌ not built                                             |
| `mix optimal.wiki.*` tasks  | ❌ not built                                             |
| Deliver reads wiki first    | ❌ not wired (today ContextAssembler goes straight to chunks) |

The L0 cache (existing) is a **distant cousin** of the Wiki — it's
machine-generated, not LLM-curated; it's flat, not hierarchical; it has no
citations. The Wiki replaces it as the top layer once built.

## Build order (delta from INTENT-MACHINE.md)

Wiki layer is a SEPARATE track from the chunking/multimodal build, but it
depends on a few primitives:

1. **Schema writing** (prose, no code) — `.wiki/SCHEMA.md` + page template.
2. **Directive parser** — `OptimalEngine.Wiki.Directives` — `{{cite}}`, `{{include}}`, `{{expand}}`, `{{search}}`. Renders to plain text, HTML, or Claude-format.
3. **Citation integrity checker** — walks every wiki page, verifies every `{{cite}}` resolves, flags orphans.
4. **Curator** — `OptimalEngine.Wiki.Curator` with an Ollama-driven prompt that takes `(existing_page, new_signals, schema) -> updated_page`.
5. **Maintenance trigger** — on signal ingest, compute affected wiki pages, enqueue curation jobs.
6. **Deliver integration** — `ContextAssembler` prepends relevant wiki pages to its output, then falls through to raw chunks only if the query isn't answered.
7. **CLI + HTTP surface** — `mix optimal.wiki.view <slug>`, `mix optimal.wiki.rebuild`, `GET /api/wiki/:slug`, etc.

This track CAN run in parallel with chunking/multimodal because it operates at
the top of the stack — whatever the lower pipeline produces, the wiki curates
a view of it.
