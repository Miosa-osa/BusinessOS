# Sample Workspace — "Acme Corp"

A complete example of how an organization lays out its knowledge on disk
for the Optimal Engine to ingest.

This directory is **opt-in reference data** — nothing here is required for
the engine to run. Use it to:

- See the three-tier convention (raw sources → derivatives → wiki) with
  real files you can open and edit.
- Bootstrap your own workspace via `mix optimal.init <target-dir>`.
- Run the engine end-to-end against it: `mix optimal.ingest_workspace
  sample-workspace/`.

---

## Directory anatomy

```
sample-workspace/
├── nodes/                      # Organizational units (Tier 1 source)
│   ├── 01-founder/
│   │   ├── context.md          # Persistent facts (edit in place)
│   │   ├── signal.md           # Weekly / rolling status
│   │   └── signals/            # Dated signal files (append-only)
│   │       └── YYYY-MM-DD-slug.md
│   ├── 02-platform/
│   └── …
│
├── .wiki/                      # Tier 3 — LLM-maintained curation
│   ├── SCHEMA.md               # Governance rules the curator honors
│   └── <slug>.md               # Curated pages with citations back to Tier 1
│
├── architectures/              # User-defined data-point schemas
│   └── clinical_visit.yaml     # Example custom Architecture
│
└── assets/                     # Binary attachments (images, PDFs, audio)
```

### Tier 1 — raw sources (this directory)

The file is the source of truth. The engine ingests, classifies, and
indexes but **never rewrites** these files.

- `context.md` — persistent facts about the node. Updated manually
  when ground truth changes.
- `signal.md` — the rolling weekly status for that node. Overwritten
  each cycle.
- `signals/YYYY-MM-DD-slug.md` — append-only. One file per event:
  transcripts, decisions, plans, notes. Convention:
  `YYYY-MM-DD-lowercase-slug.md`.

### Tier 2 — derivatives

Lives in `.optimal/index.db` (SQLite). Rebuildable from Tier 1 at any
time with `mix optimal.reindex`. Holds the FTS5 index, per-chunk
classifications + intents, extracted entities, edges, chunks at 4
scales, embeddings, cluster assignments, audit events.

### Tier 3 — the wiki

Lives in `.wiki/`. LLM-maintained. Each file is a curated page with
YAML frontmatter and `{{cite: …}}` directives pointing back to Tier 1
chunks. The curator:

1. Reads the governance rules in `SCHEMA.md`.
2. Appends new signals under a `## New signals` section with citations.
3. Rewrites existing sections when the underlying facts change.

---

## Signal file convention

Every signal file has YAML frontmatter + body:

```markdown
---
title: Customer pricing call — Q4
genre: transcript
mode: linguistic
node: 04-academy
sn_ratio: 0.75
entities:
  - { name: "Alice",   type: person }
  - { name: "Bob",     type: person }
  - { name: "Healthtech Product", type: product }
---

## Summary

One-sentence abstract. The engine pulls this for L0 (~100 tokens).

## Key points

- Bullets here. The engine pulls this block for L1 (~2K tokens).
- Each bullet should be an atomic claim.

## Detail

Full content (L2 / `content` column). Prose is fine; the decomposer
splits into paragraphs + sentences automatically.
```

Fields the engine reads from frontmatter:

| Field       | Required | Used for                                          |
|-------------|----------|---------------------------------------------------|
| `title`     | yes      | Display + search ranking                          |
| `genre`     | yes      | Classification + retrieval filter                 |
| `mode`      | no       | `linguistic` / `visual` / `code` / `data` / `mixed` |
| `node`      | inferred | Routed from directory; override if needed         |
| `sn_ratio`  | no       | Signal/noise boost; defaults to 0.5               |
| `entities`  | no       | Extracted entities (the engine can derive them)   |
| `authored_at` | no     | ISO-8601 timestamp; defaults to file mtime        |

---

## Running the engine against this workspace

```bash
# One-shot setup — point the engine at this tree
mix optimal.ingest_workspace sample-workspace/

# Interactive usage
mix optimal.search "healthtech pricing"
mix optimal.rag "what's the decision on Q4 pricing?" --trace
mix optimal.wiki view healthtech-pricing-decision
mix optimal.graph hubs
```

Or from a web UI:

```bash
# config/dev.exs
config :optimal_engine, :api, enabled: true, port: 4200

iex -S mix     # boots the engine + HTTP API
```

Then `cd desktop && npm run dev` and open http://localhost:1420.

---

## Starting fresh

Use `mix optimal.init <my-workspace>` to scaffold a new workspace
that mirrors this layout. Six empty nodes, a SCHEMA.md, and a tiny
starter signal so you can verify the pipeline works before you
ingest real data.
