# Optimal Engine — Core Concepts

---

## The Three Tiers

Every piece of knowledge lives at exactly one tier. Retrieval always tries Tier 3 first.

```
TIER 3 — THE WIKI              LLM-maintained. Read first. Audience-aware.
Path: .wiki/                   Hot citations + executable directives.
                               The agent's front door. Curated by Ollama on
                               every ingest. Human schema; machine pages.
           ▲ CURATE ▼
TIER 2 — DERIVATIVES           Machine-maintained. Rebuildable.
Path: .optimal/index.db        SQLite + FTS5 + vectors + graph + clusters
                               + L0 abstracts. Produced by the 8-stage
                               pipeline. `mix optimal.rebuild` recreates it.
           ▲ DERIVE ▼
TIER 1 — RAW SOURCES           Immutable. Append-only. Hash-addressed.
Path: nodes/**/signals/*.md    Signal files, PDFs, images, audio, video.
      assets/                  The engine NEVER rewrites them.
```

**Four invariants:**

| Invariant | Rule |
|---|---|
| Tier 1 append-only | Every write is a new file at a new path |
| Tier 2 rebuildable | `mix optimal.rebuild` reconstructs from Tier 1 exactly |
| Tier 3 LLM-owned | Humans write the schema; the curator writes the pages |
| Citations downward-only | T3 → T2/T1, T2 → T1. Never upward. Acyclic. Auditable. |

---

## The Nine-Stage Ingestion Pipeline

Every signal flows through all nine stages in strict order. Each stage has one responsibility and a typed contract with the next. No stage is skipped. No shortcuts.

```
1. INTAKE    2. PARSE      3. DECOMPOSE   4. CLASSIFY    5. EMBED
   │            │              │              │              │
   ▼            ▼              ▼              ▼              ▼
receive      format→text   large → small   S=(M,G,T,F,W)  multi-modal
any input    any format    hierarchical    + intent       aligned vectors
+ hash       + metadata    chunks          per scale      per scale
+ provenance

6. ROUTE    7. STORE      8. CLUSTER     9. CURATE
   │            │              │              │
   ▼            ▼              ▼              ▼
node +       SQLite +       HDBSCAN over   affected wiki
topology     FTS5 +         feature graph  pages → Ollama
assignment   vectors +      incremental    curator →
             graph edges    (never full)   verify → commit
```

### Stage 1 — Intake

Receives signals from any surface (CLI, HTTP, filesystem watcher, programmatic). Hashes the payload for deduplication. Records provenance. Rejects if `content_hash` already ingested (idempotent).

### Stage 2 — Parse

Converts any input format into plain text + structural metadata. Preserves every boundary the source exposes (headings, pages, slides, timestamps, code blocks).

**10 format backends:**

| Format | Backend |
|---|---|
| `.md .txt .rst .adoc` | Native |
| `.yaml .yml .toml .json` | `yaml_elixir` / `Jason` |
| `.csv .tsv` | `NimbleCSV` |
| `.html` | `Floki` |
| Source code (30+ extensions) | Tree-sitter or native |
| `.pdf` | `pdftotext` shell or `pdf_extract` |
| `.docx .pptx .xlsx` | `zip` + OOXML parser |
| `.png .jpg .jpeg .gif .webp` | `tesseract` OCR |
| `.mp3 .wav .m4a .ogg .flac` | `whisper.cpp` |
| `.mp4 .mov .webm` | `ffmpeg` → image + audio backends |

### Stage 3 — Decompose

Breaks parsed documents into a hierarchical chunk tree at four fixed scales. Parent-child links are maintained.

**Four chunk scales:**

| Scale | Target size | Created for |
|---|---|---|
| `:document` | Full source | Every signal |
| `:section` | Structural unit | Every heading / page / slide boundary |
| `:paragraph` | Semantic block | Every paragraph, code block, table row group |
| `:chunk` | ~512 tokens | Sliding window with 64-token overlap, respects paragraph boundaries |

Rules: never split across structural boundaries; never split a sentence mid-word; code chunks respect top-level function/class boundaries.

### Stage 4 — Classify

For every chunk at every scale: determines `S=(M,G,T,F,W)` dimensions AND intent. Heuristics-first, Ollama-augmented when available.

### Stage 5 — Embed

Projects every chunk into a shared 768-dim aligned vector space.

| Modality | Model | Dimensions |
|---|---|---|
| text | `nomic-embed-text-v1.5` | 768 |
| image | `nomic-embed-vision-v1.5` | 768 (aligned with text) |
| audio | `whisper.cpp` → text embed | 768 |
| code | `nomic-embed-text-v1.5` | 768 |

**Alignment invariant:** a text query can retrieve an image because both live in the same 768-dim nomic space.

### Stage 6 — Route

Assigns each chunk to one primary node + N cross-reference nodes based on entities, keywords, and topology rules.

### Stage 7 — Store

Persists everything from stages 1–6 atomically in one SQLite transaction. If stage 7 fails, nothing the user sees was written.

**Schema tables:** `signals`, `chunks`, `classifications`, `embeddings`, `entities`, `edges`, `clusters`, `cluster_members`, `assets`, `wiki_pages`, `citations`, `events`

### Stage 8 — Cluster

HDBSCAN over a feature vector composed of: `0.6 × embedding + 0.2 × entity_overlap + 0.15 × intent_match + 0.05 × node_affinity`. Runs incrementally per new chunk. Theme names auto-generated by Ollama over top-N chunks per cluster.

### Stage 9 — Curate (Wiki Maintenance)

On each `store.chunk.indexed` event: computes affected wiki pages → enqueues curation jobs → Ollama rewrites the page with new citations → verify gate → commit. Fail closed — reject + flag on any citation that doesn't resolve.

---

## Signal Classification: S = (M, G, T, F, W)

Every chunk at every scale is classified across five dimensions. This is the theoretical foundation from "Signal Theory: The Architecture of Optimal Intent Encoding" (MIOSA Research, 2026).

| Dim | Name | Question | Examples |
|---|---|---|---|
| M | **Mode** | How is it perceived? | linguistic, visual, code, data, mixed |
| G | **Genre** | What conventionalized form? | spec, brief, plan, transcript, report, ADR, note |
| T | **Type** | What does it DO? | direct, inform, commit, decide, express |
| F | **Format** | What container? | markdown, code, JSON, CLI output, diff |
| W | **Structure** | Internal skeleton | genre-specific template |

Before storing any non-trivial content, all five dimensions are resolved. Unresolved dimensions are flagged as noise.

---

## The Ten-Value Intent Enum

Intent is inferred per chunk from (chunk content + signal type + local LLM vote). Drives retrieval boosting when the query has a clear intent type.

| Intent | Meaning |
|---|---|
| `request_info` | Asking for something |
| `propose_decision` | Putting a decision on the table |
| `record_fact` | Stating something as ground truth |
| `express_concern` | Flagging risk or blocker |
| `commit_action` | Taking on a task |
| `reference` | Pointing at other context |
| `narrate` | Describing a sequence of events |
| `reflect` | Analyzing past signals |
| `specify` | Defining a contract or requirement |
| `measure` | Reporting a metric or quantity |

---

## Three Governing Constraints (Shannon / Ashby / Beer)

Every Signal is subject to three hard constraints. Violate any one and the Signal fails regardless of its content.

### 1. Shannon — The Ceiling

Every channel has finite capacity. Don't exceed the receiver's bandwidth. A 500-line explanation when 20 lines suffice is a Shannon violation. The engine addresses this with bandwidth tiers: `l0` (~100 tok) / `l1` / `full`.

**Engine artifact:** `ContextAssembler` — selects chunks to fit token budget, prefers coarsest scale that answers.

### 2. Ashby — The Repertoire

Have enough Signal variety (genres, modes, structures) to handle every situation. Prose when a table is needed is an Ashby violation. The engine addresses this with the 5-dimension `S=(M,G,T,F,W)` classification and the Composer's format variants (`text` / `json` / `claude` / `openai`).

**Engine artifact:** `OptimalEngine.Classifier`, `OptimalEngine.Composer`

### 3. Beer — The Architecture (Recursive Viability)

Maintain viable structure at every scale. A response, a file, a system — each must be coherently structured. Orphaned logic is a Beer violation. The engine addresses this with the 3-tier architecture: every claim either lives in a chunk (Tier 2) or in a wiki page that cites a chunk (Tier 3). No orphaned facts.

**Engine artifact:** `OptimalEngine.Wiki.Integrity` — citation + contradiction checker

---

## Eleven Failure Modes

```
SHANNON VIOLATIONS
  Routing failure       wrong recipient
  Bandwidth overload    too much output
  Fidelity failure      meaning lost in encoding

ASHBY VIOLATIONS
  Genre mismatch        wrong form for the situation
  Variety failure       no genre exists for this situation
  Structure failure     no internal skeleton imposed

BEER VIOLATIONS
  Bridge failure        no shared context between sender and receiver
  Herniation failure    incoherence across layers
  Decay failure         outdated Signal — not sunsetted

WIENER VIOLATION
  Feedback failure      no confirmation loop — action not verified

CROSS-CUTTING
  Adversarial noise     deliberate degradation
```

Detected at the data layer by `mix optimal.health` and `OptimalEngine.Signal.FailureModes`.

---

## Retrieval Flow

```
AGENT QUERY
  │
  ▼
IntentAnalyzer — decode query intent + entities + temporal scope
  │
  ▼
WIKI-FIRST LOOKUP (Tier 3) ──── sufficient? ──── yes ──▶ Composer ──▶ packaged context
  │                                                                      (agent prompt)
  no
  │
  ▼
HYBRID RETRIEVAL (Tier 2)
  BM25 + vector + graph_boost + intent_match + cluster_expand + temporal_decay
  │
  ▼
Context Assembler — fits token budget, prefers coarsest scale
  │
  ▼
Composer — formats for target model
```

**The wiki is always tried first.** Most agent queries never touch the retriever because the curated wiki already answered them.

---

## Proactive Surfacing — 14 Categories

The Surfacer watches for relevant changes and pushes notifications via SSE. Categories (Engramme-derived enterprise memory taxonomy):

1. `recent_actions` — new decisions or commits by tracked actors
2. `ownership` — ownership changes for entities in scope
3. `contradictions` — new signals that contradict existing wiki claims
4. `blockers` — express_concern chunks about tracked topics
5. `deadlines` — temporal signals about upcoming events
6. `handoffs` — actor transitions or delegation events
7. `escalations` — signals with elevated concern scores
8. `metrics` — new measure-intent chunks for tracked entities
9. `specifications` — new specify-intent chunks (contract changes)
10. `references` — cross-references to tracked topics from other nodes
11. `clusters` — new thematic clusters containing tracked entities
12. `wiki_updates` — wiki page updates for subscribed slugs
13. `entity_graph` — new edges in the entity graph for tracked entities
14. `compliance` — PII detection or retention-trigger events

Subscribe via `POST /api/subscriptions` and receive pushes via `GET /api/surface/stream`.
