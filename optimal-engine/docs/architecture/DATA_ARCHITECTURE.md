# Data Architecture — the universal data-point layer

> **Thesis.** The engine is not an LLM wrapper. It is a **model-agnostic
> data substrate** — a storage, retrieval, and reasoning layer that any
> processor (LLM, vision model, classical ML, rule engine, agent) can
> read from and write to. The unit of storage is a **data point**
> whose shape is declared by a **data architecture**.

## Motivation

Every retrieval system built in the last three years assumes one
dominant shape: a chunk of text embedded into a single vector space.
That assumption is wrong for the workloads we care about. An
organization produces:

- text (notes, transcripts, specs, commits)
- images (diagrams, screenshots, scans, photos)
- audio (calls, meetings, voice memos)
- video (product demos, walkthroughs, webinars)
- structured records (invoices, tickets, deals, clinical encounters)
- time series (telemetry, vitals, prices, metrics)
- code (diffs, modules, configs, SQL)
- graphs (ontologies, dependency trees, social nets)
- tensors (feature vectors, activations, gradients)
- geospatial (polygons, trajectories, points)

Each of these has its own embedding family, its own decomposition
strategy, its own classification taxonomy, its own retention policy.
A system that collapses them all into "text chunks" throws away
everything that makes each modality retrievable.

The **DataArchitecture** layer is how we encode that variety without
privileging any single processor.

## Primitives

### Field (`OptimalEngine.Architecture.Field`)

One typed slot in a data point. Carries:

| attribute    | purpose                                             |
|--------------|------------------------------------------------------|
| `:name`      | atom identifier                                     |
| `:modality`  | enum from `Field.modalities()`                      |
| `:dims`      | shape (`[768]`, `[3, 224, 224]`, `[:any, 768]`)     |
| `:required`  | can the field be absent?                            |
| `:processor` | atom id of the preferred processor (optional)       |

Modalities: `:text`, `:code`, `:image`, `:audio`, `:video`,
`:time_series`, `:table`, `:structured`, `:graph`, `:tensor`,
`:geo`, `:binary`.

### Architecture (`OptimalEngine.Architecture.Architecture`)

A named composition of fields plus:

- `:modality_primary` — the main modality of the data point
- `:granularity` — ordered list (`[:document, :section, :paragraph, :sentence]`)
- `:retention` — atom policy id
- `:metadata` — free-form map

### Processor (`OptimalEngine.Architecture.Processor`)

A behaviour. Any module that implements the contract can be a
processor:

```elixir
@callback id() :: atom()
@callback modality() :: Field.modality() | :any
@callback emits() :: [emit_kind()]
@callback init(config :: map()) :: {:ok, state} | {:error, term()}
@callback process(field, value, state) ::
            {:ok, output()} | {:error, term()}
```

A processor wraps *anything* that turns a field value into an
indexable output: a Whisper transcription, a CLIP image embedder,
a scikit-learn classifier, a regex rule engine, an agent loop. The
engine doesn't care which category it's in — it calls `process/3`.

## Built-in catalog

The first release ships seven architectures covering the dimension
space:

| name                 | primary modality | granularity                               |
|----------------------|------------------|-------------------------------------------|
| `text_signal`        | text             | document → section → paragraph → sentence |
| `image_asset`        | image            | image → region                            |
| `audio_transcript`   | audio            | clip → utterance → word                   |
| `video_clip` (*)     | video            | clip → scene → shot → frame               |
| `structured_record`  | structured       | record → field                            |
| `time_series_window` | time_series      | window → sample                           |
| `code_commit`        | code             | commit → file → hunk → line               |
| `multimodal_media`   | video            | clip → scene → shot → frame               |

(*) `video_clip` is covered by `multimodal_media` with the visual
embedder alone; split out if we need a leaner shape.

And six processor stubs:

| processor              | modality       | emits                       |
|------------------------|----------------|-----------------------------|
| `text_embedder`        | text           | embedding                   |
| `image_embedder`       | image          | embedding                   |
| `audio_embedder`       | audio          | embedding, transcription    |
| `video_embedder`       | video          | embedding, caption          |
| `code_embedder`        | code           | embedding                   |
| `ts_feature_extractor` | time_series    | features                    |

`text_embedder` wraps `OptimalEngine.Embed.Ollama` (nomic-embed-text).
`image_embedder` wraps the same module's `embed_image/2`
(nomic-embed-vision, 768-d aligned). `ts_feature_extractor` is pure
classical statistics — no model dependency, and a good template for
future rule-based or algorithmic processors.

## State-of-the-art inspirations

| Pattern in the wider field                          | How we apply it                                                          |
|------------------------------------------------------|--------------------------------------------------------------------------|
| **Schema-on-read** (Iceberg, Parquet, Delta Lake)   | Architecture is the schema; `Apply.validate/2` enforces shape at ingest  |
| **Multi-modal alignment** (CLIP, ImageBind, LanguageBind) | Every embedder emits into the engine's 768-d aligned space; the architecture declares which embedder owns which field |
| **Hybrid retrieval** (ColBERT, SPLADE, BM25+dense)  | Wiki-first → FTS5 BM25 → dense vectors → graph boost, composed by `Retrieval.RAG` |
| **Memory consolidation** (Hippocampal replay, Titans, MemGPT) | `Memory.Cortex` episodic → semantic consolidation; LLM-curated wiki as Tier 3 |
| **Neuro-symbolic** (RETRO, SimOn, OWL reasoners)    | Knowledge module runs OWL 2 RL reasoner next to vector retrieval         |
| **Data lineage** (DBT, Dagster, OpenLineage)         | `processor_runs` audit row per field × processor; every emission carries its origin |
| **Versioned data** (DVC, LakeFS, Git-LFS)           | Wiki pages are versioned (Tier 3); signals carry `supersedes` pointers; retention policies gated by legal holds |
| **Provenance + confidence** (Trusted Data Framework)| Every processor output carries confidence + `processor_runs` carries the method; auditable by `mix optimal.compliance dsar` |
| **Hierarchical chunking** (RAPTOR, Hierarchical Navigable Small World) | 4-scale decomposition (document → section → paragraph → sentence) persisted in `chunks` |
| **Active retrieval** (FLARE, Self-RAG)              | Retrieval.Receiver declares the receiver's bandwidth; BandwidthPlanner stops fetching once the budget is spent |

## Model-agnostic by construction

The architecture layer does not name GPT, Claude, Gemini, or any
specific model. It names **processor atoms** — `text_embedder`,
`image_embedder` — and the registry resolves them to modules at
runtime.

Consequences:

1. A user wanting to swap the text embedder from `nomic-embed-text`
   to `bge-large` changes one line in `OptimalEngine.Architecture.Processors.TextEmbedder`
   (or registers a new processor and points the architecture at it).
   No other code path changes.
2. A field whose value is a time series bypasses the LLM path
   entirely — `ts_feature_extractor` is classical statistics and
   has no model dependency.
3. Agents and LLMs are **consumers**, not privileged clients. They
   hit the same `/api/rag`, `/api/signals/:id`, `/api/architectures`
   endpoints a vision model or a Prolog proof engine would.

## How a data point flows through the engine

```
  INGEST                    DISPATCH                  PERSIST
  ──────                    ────────                  ───────
  %{body: "...",            Architecture.fetch(       INSERT contexts
    authored_at: "..."}       "text_signal")          INSERT chunks (4 scales)
       │                          │                   INSERT entities
       ▼                          ▼                   INSERT classifications
  Architecture              Architecture.apply(       INSERT intents
  .validate/2                 arch, data, opts)       INSERT processor_runs
                                 │                    INSERT events (audit)
                                 ▼
                            For each field with a
                            declared processor:
                              process/3 → output
                                 │
                                 ▼
                            {:ok, %{kind: :embedding,
                                    value: [...],
                                    metadata: %{...}}}
```

The `Apply.run/3` orchestrator is stateless and parallelizable. It
logs every processor invocation to `processor_runs` so the audit
trail can answer "what processor produced this embedding, when,
with what confidence?" for compliance + debugging.

## What's next (Phase 15+)

1. **Dedicated audio / video embedders** — wav2vec-style + VideoMAE.
   The processor stubs return `:not_implemented` today; the
   architecture layer is ready.
2. **Pluggable tenant processors** — let a tenant register a Python
   gRPC service as a processor for a custom field type.
3. **Cross-architecture retrieval** — a query matches against
   `text_signal` bodies AND `code_commit` diffs AND
   `multimodal_media` transcripts simultaneously, weighted by the
   receiver's intent.
4. **Architecture inheritance** — a tenant declares
   `clinical_visit` extending `structured_record` by adding
   vitals + imaging fields. Schema-on-read composes the parent.
5. **Native vector store for non-768-d spaces** — `chunk_embeddings`
   is aligned at 768 today; sparse + higher-dim + per-modality
   stores land when we add processors whose output doesn't fit the
   aligned space.

## Files

```
lib/optimal_engine/architecture.ex               — facade
lib/optimal_engine/architecture/
├── architecture.ex                              — struct + helpers
├── field.ex                                     — field struct + modality enum
├── processor.ex                                 — @behaviour
├── registry.ex                                  — built-in + tenant catalog
├── processor_registry.ex                        — processor lookup
├── apply.ex                                     — validate/3 + run/3
├── architectures/
│   ├── text_signal.ex
│   ├── image_asset.ex
│   ├── audio_transcript.ex
│   ├── structured_record.ex
│   ├── time_series_window.ex
│   ├── code_commit.ex
│   └── multimodal_media.ex
└── processors/
    ├── text_embedder.ex
    ├── image_embedder.ex
    ├── audio_embedder.ex
    ├── video_embedder.ex
    ├── code_embedder.ex
    └── ts_feature_extractor.ex

lib/mix/tasks/optimal.architectures.ex           — CLI
```

CLI:

```
mix optimal.architectures              list all
mix optimal.architectures show <name>  field-level detail
mix optimal.architectures processors   registered processors
```

HTTP API:

```
GET /api/architectures                            catalog + processors
GET /api/architectures/:id                        field-level detail
```

Desktop UI:

```
/architectures     browse catalog, inspect fields + processors
```
