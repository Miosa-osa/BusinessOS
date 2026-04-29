# Optimal Engine — Architecture

> **One canonical architecture. One source of truth. One flow of data.**
> If this doc and the code disagree, the code is wrong.
>
> **Scale:** enterprise second brain — not a personal tool. See
> [`ENTERPRISE.md`](ENTERPRISE.md) for tenancy, ACLs, connectors, retention,
> audit, and performance targets. When `ENTERPRISE.md` and this doc disagree,
> `ENTERPRISE.md` wins.

---

## 1. Thesis

The Optimal Engine is the **second brain of a company** — an intent machine
that ingests every signal flowing through an organization (chat, email, docs,
meetings, tickets, CRM, code), breaks each down from large to small, decodes
its intent at every scale, clusters it wide by theme, and delivers the right
chunks at the right grain to any agent or human who queries it — scoped to
what that caller is allowed to see.

On top of raw storage sits an **LLM-maintained Wiki** — the front door every
agent reads first. Wiki pages carry **hot citations** back to source material
and **executable directives** that pull more context on demand. Agents don't
re-discover facts on every query; the wiki already knows them.

Three promises:

1. **Signal integrity.** Nothing is stored without classification, intent, and
   citation lineage. The engine knows *why* each fact exists.
2. **Scale alignment.** Every piece of content exists simultaneously as
   document / section / paragraph / chunk. Retrieval returns the coarsest
   scale that answers the query.
3. **Modality alignment.** Text, image, and audio embed into the same vector
   space (via nomic-embed-text / nomic-embed-vision + whisper.cpp). A text
   query can retrieve an image. An image can retrieve an audio clip. One
   retrieval layer, not three.

---

## 2. The Three Tiers

```
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 3 — THE WIKI              LLM-maintained. Read first.          │
│  Path: .wiki/                    Hot citations + executable          │
│                                  directives. The agent's front door. │
│                                  Curated by Ollama on every ingest.  │
└──────────────────────────────────────────────────────────────────────┘
                          ▲ CURATE ▼
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 2 — DERIVATIVES           Machine-maintained. Rebuildable.     │
│  Path: .optimal/index.db         SQLite + FTS5 + vectors + graph +   │
│                                  clusters + L0 abstracts. Produced   │
│                                  by the 8-stage pipeline.            │
└──────────────────────────────────────────────────────────────────────┘
                          ▲ DERIVE ▼
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 1 — RAW SOURCES           Immutable. Append-only.              │
│  Path: nodes/**/signals/*.md     The signal files, PDFs, images,     │
│        assets/                   audio, video. Hash-addressed. The   │
│                                  engine NEVER rewrites them.         │
└──────────────────────────────────────────────────────────────────────┘
```

| Invariant                                                                                     |
|-----------------------------------------------------------------------------------------------|
| Tier 1 is append-only. Every write is a new file at a new path. Nothing gets edited in place. |
| Tier 2 is fully derivable from Tier 1. `mix optimal.rebuild` reconstructs it from scratch.    |
| Tier 3 is LLM-owned. Humans write the schema; the curator writes the pages. Versioned.        |
| Citations only point downward: T3 → (T2 or T1), T2 → T1. Never upward.                        |

---

## 3. The 9-Stage Pipeline

Every signal flows through the same nine stages in strict order. Each stage has
one responsibility and a typed contract with the next.

```
  1. INTAKE    2. PARSE      3. DECOMPOSE   4. CLASSIFY    5. EMBED
     │            │              │              │              │
     ▼            ▼              ▼              ▼              ▼
  receive     format→text   large → small   S=(M,G,T,F,W)  multi-modal
  any input   any format    hierarchical    + intent       aligned vectors
  + hash      + metadata    chunks          per scale      per scale
  + provenance

     │            │              │              │              │
     ▼            ▼              ▼              ▼              ▼
  6. ROUTE    7. STORE      8. CLUSTER     9. CURATE
     │            │              │              │
     ▼            ▼              ▼              ▼
  node +      SQLite +       HDBSCAN over   affected wiki
  topology    FTS5 +         feature graph  pages → Ollama
  assignment  vectors +      incremental    curator → verify
              graph edges    (never full)   → commit
```

### Stage 1 — Intake

| Responsibility | Receive signals from any surface (CLI, HTTP, filesystem watcher, programmatic). Hash the payload for deduplication. Record provenance. |
|----|----|
| Module | `OptimalEngine.Intake` |
| Input  | `{source, format_hint, payload}` |
| Output | `%RawSignal{id, content_hash, origin, received_at, raw: binary \| text}` |
| Rules  | Reject if `content_hash` already ingested (idempotent). Every signal gets a content-addressed ID (`sha256:...`). |

### Stage 2 — Parse

| Responsibility | Convert any input format into plain text + structural metadata + asset references. |
|----|----|
| Module | `OptimalEngine.Parser` → dispatches to `Parser.<Format>` backends |
| Input  | `%RawSignal{}` |
| Output | `%ParsedDoc{text, structure: [%Heading{} \| %Section{} \| %Page{} \| %Slide{} \| %Timestamp{}], assets: [%Asset{}], modality}` |

**Format backends (built in strict order, each independently testable):**

| # | Format                       | Backend                                | Modality    |
|---|------------------------------|----------------------------------------|-------------|
| 1 | `.md .txt .rst .adoc`        | native                                 | text        |
| 2 | `.yaml .yml .toml .json`     | `yaml_elixir` / `Jason`                | data        |
| 3 | `.csv .tsv`                  | `NimbleCSV`                            | data        |
| 4 | `.html`                      | `Floki`                                | text        |
| 5 | source code (30+ ext.)       | tree-sitter or native                  | code        |
| 6 | `.pdf`                       | `pdftotext` shell or `pdf_extract`     | text        |
| 7 | `.docx .pptx .xlsx`          | `zip` + OOXML parser                   | text        |
| 8 | `.png .jpg .jpeg .gif .webp` | `tesseract` OCR + keep original        | image+text  |
| 9 | `.mp3 .wav .m4a .ogg .flac`  | `whisper.cpp` local server             | audio+text  |
| 10 | `.mp4 .mov .webm`           | `ffmpeg` → frames (#8) + audio (#9)    | video+text  |

**Rule:** the Parser preserves every structural boundary the source exposes —
headings, pages, slides, timestamps, code blocks. Stage 3 depends on them.

### Stage 3 — Decompose

| Responsibility | Break a parsed document down large→small into a hierarchical chunk tree. |
|----|----|
| Module | `OptimalEngine.Decomposer` |
| Input  | `%ParsedDoc{}` |
| Output | `%ChunkTree{root, nodes: [%Chunk{}]}` where each `%Chunk{}` carries `{id, parent_id, scale, offset, length, text, modality, asset_ref}` |

**Scales (fixed, exhaustive):**

| Scale        | Target size     | Created for                                        |
|--------------|-----------------|----------------------------------------------------|
| `:document`  | full source     | every signal                                       |
| `:section`   | structural unit | every heading/page/slide boundary                  |
| `:paragraph` | semantic block  | every paragraph, code block, table row group       |
| `:chunk`     | ~512 tokens     | sliding window with 64-token overlap, respects paragraph boundaries |

Tight chunking rules:
- Never split across structural boundaries reported by the parser.
- Never split a sentence mid-word.
- Code chunks respect top-level function/class boundaries.
- Assets (image, audio) have one `:document`-scale chunk whose `text` is the OCR/transcript and whose `asset_ref` points at the binary.

### Stage 4 — Classify

| Responsibility | For every chunk at every scale, determine the full signal dimensions AND its intent. |
|----|----|
| Modules | `OptimalEngine.Classifier`, `OptimalEngine.IntentExtractor` |
| Input   | `%Chunk{}` |
| Output  | `%Classification{mode, genre, type, format, structure, intent, sn_ratio, confidence}` |

**Signal dimensions `S = (M, G, T, F, W)` — already built, extended to per-chunk.**

**Intent enum (new, fixed):**

| Intent             | Meaning                                  |
|--------------------|------------------------------------------|
| `:request_info`    | asking for something                     |
| `:propose_decision`| putting a decision on the table          |
| `:record_fact`     | stating something as ground truth        |
| `:express_concern` | flagging risk/blocker                    |
| `:commit_action`   | taking on a task                         |
| `:reference`       | pointing at other context                |
| `:narrate`         | describing a sequence of events          |
| `:reflect`         | analyzing past signals                   |
| `:specify`         | defining a contract or requirement       |
| `:measure`         | reporting a metric or quantity           |

Intent is inferred from (chunk content + signal type + local LLM vote). Ollama
optional — without it, heuristics run and confidence is reported as lower.

### Stage 5 — Embed

| Responsibility | Project every chunk into an **aligned 768-dim vector space**. |
|----|----|
| Module | `OptimalEngine.Embedder` (wraps `OptimalEngine.Ollama` + `OptimalEngine.Whisper`) |
| Input  | `%Chunk{}` |
| Output | `%Embedding{chunk_id, model, dim: 768, vector: [float], modality}` |

**Models (open-source, local, zero cloud):**

| Modality | Model                      | Dimensions | How                                 |
|----------|----------------------------|------------|-------------------------------------|
| text     | `nomic-embed-text-v1.5`    | 768        | Ollama `/api/embeddings`            |
| image    | `nomic-embed-vision-v1.5`  | 768        | Ollama (aligned with text)          |
| audio    | `whisper.cpp` → text embed | 768        | transcript routed back to text embed|
| code     | `nomic-embed-text-v1.5`    | 768        | code chunks embed as text           |

**Alignment invariant:** a text query embedding can retrieve an image chunk
embedding because both live in the nomic 768-dim aligned space. This is the
whole product.

### Stage 6 — Route

| Responsibility | Assign each chunk to one primary node + N cross-reference nodes based on entities, keywords, and topology rules. |
|----|----|
| Module | `OptimalEngine.Router` (already built; extended for per-chunk) |
| Input  | `%Chunk{} + %Classification{}` |
| Output | `%Routing{primary_node, cross_refs: [node], confidence}` |

### Stage 7 — Store

| Responsibility | Persist everything produced by stages 1–6 atomically. |
|----|----|
| Module | `OptimalEngine.Store` (schema extended) |
| Input  | `(%RawSignal, %ParsedDoc, %ChunkTree, [%Classification], [%Embedding], %Routing)` |
| Output | `:ok` + emitted signals on `store.chunk.indexed` topic |

**Storage schema (SQLite):**

| Table             | Purpose                                              | Status    |
|-------------------|------------------------------------------------------|-----------|
| `signals`         | Tier-1 metadata: source path, hash, received_at, size| extends `contexts` |
| `chunks`          | Tier-2 hierarchical decomposition, one row per chunk | **NEW**   |
| `classifications` | Per-chunk `S=(M,G,T,F,W)` + intent + confidence      | **NEW**   |
| `embeddings`      | Per-chunk 768-dim vectors, modality-tagged           | extends `vectors` |
| `entities`        | Extracted entities per chunk                         | extends   |
| `edges`           | Typed relations in the knowledge graph               | exists    |
| `clusters`        | Theme groupings (HDBSCAN output)                     | **NEW**   |
| `cluster_members` | `chunk_id ↔ cluster_id` with membership weight       | **NEW**   |
| `assets`          | Binary blobs (images, audio, PDFs) by hash           | **NEW**   |
| `wiki_pages`      | Tier-3 curated pages with frontmatter + body         | **NEW**   |
| `citations`       | `wiki_page_id → chunk_id` with `claim_hash`          | **NEW**   |
| `events`          | Append-only log of pipeline stage transitions        | **NEW**   |

**Atomicity:** all writes for a single signal happen in one SQLite transaction.
If stage 7 fails, stages 1–6 produced nothing the user sees.

### Stage 8 — Cluster

| Responsibility | Group chunks wide by theme. Runs incrementally per new chunk, never full rebuild except via `mix optimal.cluster.rebuild`. |
|----|----|
| Module | `OptimalEngine.Clusterer` |
| Algorithm | HDBSCAN over `feature = 0.6·embedding + 0.2·entity_overlap + 0.15·intent_match + 0.05·node_affinity`. Theme names auto-generated by Ollama over top-N chunks in each cluster. |
| Output | `%Cluster{id, theme, intent_dominant, member_chunk_ids, centroid_vector}` |

### Stage 9 — Curate (Wiki Maintenance)

| Responsibility | Fold new signals into the Tier-3 wiki pages they affect. |
|----|----|
| Module | `OptimalEngine.Wiki.Curator` |
| Trigger | On each `store.chunk.indexed` event, compute affected wiki pages (= pages that cite a cluster/entity the new chunk belongs to). Enqueue curation jobs. |
| Curator loop | `(existing_page, new_chunks_with_citations, schema) --Ollama--> updated_page` |
| Verify gate | Every claim cites; every citation resolves; schema rules pass; no contradictions silently swallowed. Fail closed — reject + flag. |
| Commit | Write new page version; store diff; emit `wiki.page.updated` signal. |

Full Wiki details (directives, schema, page template) in [`WIKI-LAYER.md`](WIKI-LAYER.md).

---

## 4. Retrieval Flow (how an agent reads)

```
  AGENT QUERY ──────────────────────────────────────────────────────────┐
    "what's the current state of pricing with Ed?"                      │
                                                                        │
    ┌───────────────────┐                                               │
    │ IntentAnalyzer    │  → intent_type = :lookup, scope = ai-masters  │
    │ (decode query)    │    entities = ["Alice"], temporal = :now  │
    └─────────┬─────────┘                                               │
              ▼                                                         │
    ┌─────────────────────────────────────────────────────────┐         │
    │ WIKI-FIRST LOOKUP (Tier 3)                              │         │
    │ mix optimal.wiki.find                                   │         │
    │ → returns .wiki/ed-honour-pricing.md                    │         │
    │ → directives in page resolved on demand                 │         │
    │                                                         │         │
    │ SUFFICIENT?  ─ yes ─┐                                   │         │
    └────────────┬────────┘                                   │         │
                 no                                           │         │
                 ▼                                            ▼         │
    ┌─────────────────────────────────────────┐    ┌──────────────────┐ │
    │ HYBRID RETRIEVAL (Tier 2)               │    │   COMPOSER       │ │
    │ SearchEngine.search:                    │──▶ │  Format for      │◀┘
    │   BM25 + vector + graph_boost +         │    │  target model:   │
    │   intent_match + cluster_expand +       │    │  text / json /   │
    │   temporal_decay                        │    │  claude / openai │
    │ Returns: chunks @ appropriate scale     │    └────────┬─────────┘
    └────────────┬────────────────────────────┘             │
                 ▼                                          ▼
    ┌─────────────────────────────────────────┐       PACKAGED
    │ CONTEXT ASSEMBLER (Tier 1 materialize)  │       CONTEXT
    │ Selects chunks to fit token budget.     │       (agent prompt)
    │ Prefers coarsest scale that answers.    │
    │ Preserves citation metadata.            │
    └─────────────────────────────────────────┘
```

**The wiki is always tried first.** Only unanswered queries escalate to hybrid
retrieval. This is what makes it "better than RAG" — most agent queries never
touch the retriever because the curated wiki already answered them.

---

## 5. Module Tree

Every module lives at exactly one of these paths. If a responsibility doesn't
fit cleanly into one module, that's a design smell.

```
lib/optimal_engine/
├── intake.ex              [Stage 1]
├── parser.ex              [Stage 2 dispatch]
├── parser/
│   ├── markdown.ex
│   ├── yaml.ex
│   ├── json.ex
│   ├── csv.ex
│   ├── html.ex
│   ├── code.ex
│   ├── pdf.ex
│   ├── office.ex          .docx .pptx .xlsx
│   ├── image.ex           tesseract OCR + asset
│   ├── audio.ex           whisper.cpp + asset
│   └── video.ex           ffmpeg → image.ex + audio.ex
├── decomposer.ex          [Stage 3]
├── classifier.ex          [Stage 4] — S=(M,G,T,F,W)
├── intent_extractor.ex    [Stage 4] — intent enum
├── embedder.ex            [Stage 5]
├── ollama.ex              Ollama HTTP client
├── whisper.ex             whisper.cpp client (NEW)
├── router.ex              [Stage 6]
├── store.ex               [Stage 7] — all SQLite I/O
├── clusterer.ex           [Stage 8] (NEW)
├── wiki/
│   ├── curator.ex         [Stage 9] — Ollama-driven
│   ├── directives.ex      {{cite}} {{include}} {{expand}} …
│   ├── schema.ex          reads and validates .wiki/SCHEMA.md
│   ├── integrity.ex       citation + contradiction checker
│   └── page.ex            page struct + serialization
│
├── search_engine.ex       retrieval (Tier 2, used by composer)
├── context_assembler.ex   tiered assembly + Tier 1 materialize
├── intent_analyzer.ex     query intent decoder (already built)
├── composer.ex            output formatting (Claude/OpenAI/text/json)
│
├── knowledge/             graph + OWL reasoning (already built)
├── memory/                episodic + cortex + SICA learning (already built)
├── signal/                CloudEvents envelope + failure modes (already built)
│
├── api/
│   ├── router.ex          HTTP endpoints (Plug)
│   └── graph_router.ex    (existing graph UI)
│
├── cli.ex                 escript entry point
├── application.ex         supervision tree
└── signal.ex              core signal struct (already built)

lib/mix/tasks/             one file per subcommand (already mostly built)
```

Total target module count: ~55. Today: 38. Delta = 17 new modules.

---

## 6. Protocol Surfaces

| Surface         | Audience                            | Entry point                          |
|-----------------|-------------------------------------|--------------------------------------|
| `optimal` CLI   | any agent runtime, any shell        | `lib/optimal_engine/cli.ex` (escript → `./optimal`) |
| HTTP JSON API   | cross-language agents, web UIs      | `OptimalEngine.API.Router` on `:4200`|
| Elixir API      | in-VM callers (MIOSA, other apps)   | `OptimalEngine` public module funcs  |
| Mix tasks       | developers                          | `lib/mix/tasks/optimal.*.ex`         |

All four surfaces route through the same internal modules. No surface has its
own code path. What `optimal search` does is exactly what `GET /api/search`
does is exactly what `OptimalEngine.search/2` does.

**Stable CLI subcommand surface (locked):**

```
optimal ingest    <text | --file path | --url url>  [--genre G]
optimal search    <query>                            [--limit N --node X]
optimal read      <optimal://uri>                    [--tier l0|l1|full]
optimal rag       <query>                            [--format text|json|claude|openai]
optimal assemble  <topic>                            [--budget tokens]
optimal l0
optimal ls        <optimal://uri-prefix>
optimal graph     [triangles|clusters|hubs]
optimal reflect   [--min N]
optimal reweave   <topic>                            [--days N]
optimal wiki      <view|edit|rebuild|verify> <slug-or-all>
optimal cluster   [show|rebuild]
optimal health    [--quick]
optimal verify    [--sample N]
optimal stats
optimal remember  <observation>
optimal rethink   <topic>
optimal simulate  <scenario>
optimal impact    <entity-or-node>
optimal api                                          (start HTTP server)
optimal graph-ui                                     (launch visualizer)
```

---

## 7. Build Plan — Ordered Phases

Each phase ends with an acceptance criterion that MUST pass before the next
phase starts. No parallel phases. No half-finished phases promoted forward.

### Phase 1 — Schema (foundation)

- Add the 7 new tables (`chunks`, `classifications`, `assets`, `clusters`, `cluster_members`, `wiki_pages`, `citations`, `events`).
- Write migration code runnable with `mix optimal.migrate`.
- Backfill: every existing `contexts` row gets a `:document`-scale chunk so the old path keeps working.
- **Accept:** schema migrates cleanly from current state; all 689 existing tests still pass; new `mix optimal.stats` shows row counts for new tables.

### Phase 2 — Parser backends (format coverage)

- Implement Parser dispatch + the 10 format backends in the order listed above.
- Each backend is a separate module + separate test file.
- Parser preserves structural metadata.
- **Accept:** `mix optimal.parse <file>` returns `%ParsedDoc{}` with text + structure for every format listed. Tests cover each format with a canonical fixture.

### Phase 3 — Decomposer

- Hierarchical chunking with parent_id links at four scales.
- Respects parser-reported boundaries.
- **Accept:** round-trip: a 5-page PDF produces 1 document chunk + N section chunks + M paragraph chunks + K chunk-scale chunks, all with correct parent_id links. Reassembling chunks produces byte-identical text.

### Phase 4 — Per-chunk Classify + IntentExtractor

- Run Classifier per chunk.
- Build `IntentExtractor` with heuristics-first, Ollama-augmented-if-available.
- **Accept:** on a corpus of 50 known signals, intent accuracy ≥ 80% vs hand-labeled gold set.

### Phase 5 — Embedder (multi-modal)

- Extend Ollama wrapper with `embed_text/1`, `embed_image/1`.
- Add `Whisper` module for audio transcription via local whisper.cpp server.
- `Embedder.embed/1` dispatches on modality.
- **Accept:** text-query-retrieves-image test passes. An image of "a pricing chart with $2K" embedded alongside a text chunk about "Ed's $2K pricing" land in the same retrieval top-5.

### Phase 6 — Clusterer

- HDBSCAN over feature vectors. Incremental add. Theme auto-naming via Ollama.
- **Accept:** on the existing corpus, ≥ 70% of chunks cluster coherently by human eval of 20 random clusters.

### Phase 7 — Wiki Layer

- `.wiki/SCHEMA.md` written.
- Page template defined.
- `OptimalEngine.Wiki.Directives` parser + renderer.
- `OptimalEngine.Wiki.Integrity` checker.
- `OptimalEngine.Wiki.Curator` Ollama-driven.
- Maintenance trigger: `store.chunk.indexed` → affected-pages queue → curator.
- **Accept:** ingesting a new signal about Alice updates `.wiki/ed-honour-pricing.md` autonomously with new citations verified.

### Phase 8 — Scale-aware Deliver + Composer

- `ContextAssembler` tries wiki first.
- Falls through to hybrid retrieval only on wiki miss.
- `Composer` formats output for target model (text/json/claude/openai).
- **Accept:** `optimal rag "pricing with Ed"` returns wiki-sourced summary with inline citations on first call; zero retriever hits.

### Phase 9 — CLI + HTTP surface (complete)

- Fill in any missing CLI subcommands.
- HTTP API covers the full subcommand surface with matching payload shapes.
- **Accept:** every `optimal <cmd>` has an equivalent `POST /api/<cmd>` that returns the same data.

### Phase 10 — Production hardening

- Structured JSON logging option.
- Telemetry spans for every pipeline stage.
- `mix release optimal` produces a standalone OTP release.
- Burrito single-binary build for cross-platform distribution.
- **Accept:** the release tarball runs on a clean macOS or Linux VM with no Erlang installed and passes smoke tests.

### Phase 11 — Push

- Final docs pass.
- Version 0.1.0 tagged.
- `git push -u origin main`.
- **Accept:** Alice's explicit go.

---

## 8. Data Contracts (strict types at every boundary)

```elixir
# Stage 1 output
%OptimalEngine.RawSignal{
  id: content_hash_id(),
  origin: String.t(),
  format_hint: atom(),
  received_at: DateTime.t(),
  raw: binary()
}

# Stage 2 output
%OptimalEngine.ParsedDoc{
  signal_id: id(),
  text: String.t(),
  structure: [structural_boundary()],
  assets: [%Asset{}],
  modality: :text | :image | :audio | :video | :mixed
}

# Stage 3 output
%OptimalEngine.ChunkTree{
  root_chunk_id: chunk_id(),
  chunks: [%Chunk{}]
}
%OptimalEngine.Chunk{
  id: chunk_id(),
  parent_id: chunk_id() | nil,
  scale: :document | :section | :paragraph | :chunk,
  signal_id: id(),
  offset: non_neg_integer(),
  length: non_neg_integer(),
  text: String.t(),
  modality: atom(),
  asset_ref: asset_hash() | nil
}

# Stage 4 output
%OptimalEngine.Classification{
  chunk_id: chunk_id(),
  mode: atom(),
  genre: atom(),
  type: atom(),
  format: atom(),
  structure: atom(),
  intent: atom(),
  sn_ratio: float(),
  confidence: float()
}

# Stage 5 output
%OptimalEngine.Embedding{
  chunk_id: chunk_id(),
  model: String.t(),
  modality: atom(),
  dim: 768,
  vector: [float()]
}

# Stage 6 output
%OptimalEngine.Routing{
  chunk_id: chunk_id(),
  primary_node: String.t(),
  cross_refs: [String.t()],
  confidence: float()
}

# Stage 8 output
%OptimalEngine.Cluster{
  id: cluster_id(),
  theme: String.t(),
  intent_dominant: atom(),
  member_chunks: [{chunk_id(), weight :: float()}],
  centroid: [float()]
}

# Stage 9 output
%OptimalEngine.Wiki.Page{
  slug: String.t(),
  frontmatter: map(),
  body: String.t(),
  citations: [%Citation{}],
  last_curated: DateTime.t(),
  version: non_neg_integer()
}
```

All contracts are enforced at stage boundaries via pattern match or
`@spec`-checked boundaries. A violation halts the pipeline and emits a failure
event — no silent degradation.

---

## 9. The Three Invariants

Everything above enforces three invariants that make the engine *optimal*:

1. **Single source of truth per scale.** A fact is either a chunk or a wiki
   claim citing chunks. Never both copied.
2. **Aligned vector space.** Text, image, audio all embed to the same 768-dim
   space. One retriever, not three.
3. **Citations only point down.** Tier 3 cites Tier 1/2. Tier 2 cites Tier 1.
   Tier 1 cites nothing. Acyclic. Auditable.

Keep these three invariants and the engine stays coherent as it grows.
Violate any one and the rot starts.

---

## 10. Starting now

Phase 1 (Schema) begins immediately. Everything downstream unlocks on its
acceptance criterion. I'll report progress per phase, not per file.
