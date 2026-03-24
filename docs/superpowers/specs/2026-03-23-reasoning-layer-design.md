# BusinessOS Vision 2030: The Reasoning Layer

**Date:** 2026-03-23
**Status:** Approved
**Approach:** A — The Reasoning Layer
**Scope:** Cross-platform, 7 innovations

---

## Core Insight

**Incumbents store data. BusinessOS stores meaning.**

Salesforce, Monday, and Notion store business data in relational tables. When an AI agent needs to answer relational questions, it must guess schemas, construct joins, and hope for consistency. With The Reasoning Layer, every ODCS workspace becomes a SPARQL-queryable knowledge graph where relationships are first-class citizens. The ontology *is* the schema.

**80/20 Bet:** Innovation #1 (Auto-Ontology) — automatically generating RDF mappings from any ODCS workspace without manual configuration. Once this exists, every workspace becomes a knowledge graph by default, and all 6 other innovations become trivial.

---

## Innovation 1: Auto-Ontology

**Problem:** Requires hand-crafted `ontology-mappings.json` with explicit table-to-class mappings, property mappings, value maps, and FK references. Every new table needs a human to write the mapping.

**Solution:** `bos ontology infer` automatically generates ontology mappings by analyzing the ODCS workspace's `model.json` (produced by `data-modelling-sdk::DataModel`):

| Signal | Mapping Rule |
|--------|-------------|
| Table name `projects` | → `schema:Project` (singularize, capitalize) |
| Table name `tasks` | → `bpmn:Task` |
| Table name `team_members` | → `org:Member` |
| Column `name` | → `schema:name` (direct) |
| Column `created_at` | → `schema:dateCreated` (temporal pattern) |
| Column `parent_id` | → `schema:parentOrganization` (FK + convention) |
| Column type `text` | → `xsd:string` |
| Column type `integer` + `is_primary_key` | → `xsd:integer` |
| Column type `timestamp` | → `xsd:dateTime` |
| FK `client_id → clients.id` | → `object_type: "uri"`, `target_table: "clients"` |
| Enum-like column (≤5 unique values) | → `value_map` from observed values |

**Schema source:** Reads from the ODCS workspace's `model.json` file (the `data-modelling-sdk::DataModel` type). This contains table definitions with column names, types, and relationship metadata. Does NOT require a live PostgreSQL connection. For FK detection, uses the `DataModel.relationships` field. For enum detection, uses value frequency analysis on string columns.

**Convention mapping table:**

```
projects    → schema:Project
tasks       → bpmn:Task
clients     → org:Organization
team_members → org:Member
contexts    → skos:Concept
conversations → schema:Discussion
artifacts   → prov:Entity
orders      → schema:Order
invoices    → schema:Invoice
employees   → org:FormalOrganization
```

**User overrides:** A `--conventions <file>` flag accepts a JSON file with custom table-to-class and column-to-predicate overrides. Built-in conventions are applied first, then overrides override on a per-table/per-column basis.

**Confidence scoring:** Each inferred mapping gets a confidence score 0.0-1.0:
- **high (≥0.8):** Direct convention match (e.g., `name` → `schema:name`, FK detected)
- **medium (0.5-0.8):** Partial match or common pattern (e.g., `status_val` → `schema:status` with abbreviation)
- **low (<0.5):** No pattern match (e.g., `cnt`, `sts`, `dt`) — requires human review

The `--confidence` flag filters output: `--confidence high` only includes high-confidence mappings.

**Dry-run mode:** `--dry-run` prints proposed mappings to stdout without writing a file, so humans can review before applying.

**CLI:**
```bash
bos ontology infer --workspace ./my-project --output ontology-mappings.json
bos ontology infer --workspace . --confidence high --output auto-mappings.json
bos ontology infer --workspace . --dry-run                    # Review before applying
bos ontology infer --workspace . --conventions custom.json   # Override conventions
```

**Output:** Valid `MappingConfig` JSON compatible with existing `ontology construct`, `ontology export`, and `ontology execute` commands.

**Files to create:**
- `bos/core/src/ontology/infer.rs` (~700 lines)
- `bos/cli/src/nouns/search.rs` (~50 lines) for Innovation 2 CLI wiring
- Tests in `bos/core/src/ontology/infer.rs` (8-10 tests)

**Dependencies:** Existing `MappingConfig`, `TableMapping`, `PropertyMapping` types. `data-modelling-sdk::DataModel` for schema introspection.

---

## Innovation 2: Semantic Search

**Problem:** Current search uses pgvector + keyword matching. Can't answer relational questions like "what tasks block the Q4 release?"

**Solution:** Two search modes:

1. **SPARQL mode** — Direct SPARQL SELECT execution against the oxigraph triple store loaded with workspace data.
2. **NL mode** — Natural language query converted to SPARQL via LLM, then executed.

**NL-to-SPARQL flow:**
1. Extract entity types and predicates from the workspace ontology (from `ontology-mappings.json`)
2. Inject workspace schema into LLM prompt as PREFIX declarations with rdfs:label/rdfs:comment
3. Send NL query to LLM (configurable: Claude, Ollama local, or any OpenAI-compatible endpoint)
4. Parse LLM output as SPARQL, validate with oxigraph's parser
5. Execute valid SPARQL, return structured JSON results
6. On parse failure, return error with suggested fix

**CLI:**
```bash
bos search --query "overdue tasks for enterprise clients"
bos search --sparql "SELECT ?s ?p ?o WHERE { ?s a bpmn:Task . ?s bpmn:status ?st FILTER(?st = \"overdue\") }"
```

**API:** `POST /api/v1/search/semantic` with body `{"query": "..."}` or `{"sparql": "..."}`.

**Files to create:**
- `bos/core/src/ontology/select.rs` (~500 lines) — SPARQL SELECT execution + NL conversion
- `bos/cli/src/nouns/search.rs` (~80 lines) — search noun with query/sparql verbs
- Go handler in `internal/handlers/search_semantic.go` (~300 lines) — HTTP handler + LLM call

**Dependencies:** oxigraph `SparqlEvaluator`, existing `QueryExecutor` pattern, `reqwest` HTTP client (already in workspace).

---

## Innovation 3: Decision Replay

**Problem:** Business decisions (ADR-001, ADR-002) are stored as static markdown. No way to trace WHY a decision was made, what data informed it, or what happened after.

**Solution:** Extend MADR decision records with PROV-O provenance triples:

```turtle
<decision/ADR-003> a prov:Activity ;
    prov:wasAssociatedWith <agent/claude> ;
    prov:used <data/sales-report-Q4> ;
    prov:generated <artifact/pricing-model-v2> ;
    prov:startedAtTime "2026-03-15T10:00:00Z"^^xsd:dateTime ;
    prov:endedAtTime "2026-03-15T10:30:00Z"^^xsd:dateTime ;
    prov:wasDerivedFrom <decision/ADR-001> .
```

**CLI:**
```bash
bos decisions trace ADR-003    # Full provenance chain
bos decisions impact --task 42  # What decisions affect this task?
```

**Files to create:**
- `bos/core/src/ontology/provenance.rs` (~200 lines) — PROV-O triple generation
- Extend `decisions.rs` with `trace()` and `impact()` methods (~150 lines)

---

## Innovation 4: Context Grounding

**Problem:** RAG uses vector similarity — fuzzy matching on embeddings. Sometimes retrieves irrelevant context.

**Solution:** Ground RAG on the knowledge graph. When the AI needs context for "project X":
1. Traverse the knowledge graph from the project node
2. Collect all related tasks, clients, artifacts, decisions (1-hop, 2-hop)
3. Feed structured triples to the LLM instead of fuzzy text chunks

**Result:** Bounded-scope context retrieval — every piece of context is graph-reachable from the query entity. Precision depends on graph density: sparse graphs yield high precision; dense hub nodes (e.g., a client connected to 100 projects) may require depth filtering or relevance scoring to avoid noise.

**CLI:**
```bash
bos context --entity project/42 --depth 2
bos context --sparql "CONSTRUCT { ?s ?p ?o } WHERE { <project/42> ?p1 ?s . ?s ?p ?o }"
```

**API:** `POST /api/v1/rag/grounded` with body `{"entity": "project/42", "depth": 2}`.

**Files to create:**
- `bos/core/src/ontology/grounding.rs` (~400 lines) — graph traversal and context collection
- Go handler in `internal/handlers/rag.go` (~300 lines)

---

## Innovation 5: Agent Intent Protocol

**Problem:** AI agents executing business operations have no standardized way to declare what they're doing, why, and what they need.

**Solution:** Four-phase protocol:

1. **DECLARE** — `POST /api/v1/agent/intent` — agent declares planned action with reasoning
2. **REQUEST** — `POST /api/v1/agent/permission` — request approval for sensitive actions
3. **EXECUTE** — `POST /api/v1/agent/execute` — perform action with provenance tracking
4. **REPORT** — `POST /api/v1/agent/outcome` — report result back to knowledge graph

Every action becomes an RDF triple: `<agent/claude> <bdev:executed> <task/42/complete> .`

**Security:**
- Agents must authenticate with a valid JWT (same auth as users)
- Agent capabilities are scoped by role-based permissions (same RBAC as human users)
- Sensitive actions (deletes, financial, PII access) always require explicit human approval
- Intent provenance data is persisted to PostgreSQL alongside the existing audit trail
- Rate limiting: max 100 intent declarations per agent per hour

**Persistence:** Agent intent triples are written to PostgreSQL via the existing `postgres` crate (not in-memory oxigraph). A `bos persist intents` command flushes in-memory provenance to the database.

**Files to create:**
- `bos/core/src/ontology/agent.rs` (~300 lines) — intent/provenance RDF generation + PostgreSQL persistence
- Go package `internal/agent/` (~600 lines) — protocol handlers, middleware, RBAC

---

## Innovation 6: Predictive Operations

**Problem:** Business operations are reactive — you know a deal is lost AFTER it's lost.

**Solution:** Implement k-NN prediction algorithms natively in bos (not depending on obsr):

- **Deal prediction:** k-NN on historical deal attributes (value, stage duration, client type, interactions) → P(close)
- **Task estimation:** Statistical analysis on past task durations per type/priority/team → realistic estimates
- **Capacity planning:** Bottleneck detection from team workload graph (high in-degree nodes)

**Note on obsr:** obsr is a standalone binary at `/Users/sac/obsr/`, not a library dependency. The prediction algorithms (k-NN, Jaccard similarity, heuristic miner) will be reimplemented in `bos/core/src/predict/` using the same mathematical approach but independent code. This avoids cross-repo dependencies.

**CLI:**
```bash
bos predict deals --probability-threshold 0.7
bos predict sprint --team backend --weeks 4
bos predict capacity --team all --horizon 30d
```

**Files to create:**
- `bos/core/src/predict/mod.rs` (~100 lines)
- `bos/core/src/predict/deals.rs` (~250 lines) — deal outcome prediction (k-NN + feature extraction)
- `bos/core/src/predict/tasks.rs` (~250 lines) — task duration estimation (statistical analysis)
- `bos/core/src/predict/capacity.rs` (~250 lines) — bottleneck prediction (graph analysis)

**Dependencies:** oxigraph for data access, standard math (no external ML dependencies).

---

## Innovation 7: Self-Describing Workspace

**Problem:** An AI agent encountering a new workspace has no way to understand what data exists or how to interact with it.

**Solution:** Every workspace serves its own schema as SPARQL-queryable metadata:

```bash
# Start the semantic server
bos ontology serve --workspace ./my-project --port 7878

# An agent can then discover the workspace:
curl http://localhost:7878/sparql -d "SELECT DISTINCT ?type WHERE { ?s a ?type }"
curl http://localhost:7878/sparql -d "SELECT ?prop ?range WHERE { <project/1> ?prop ?o }"
```

**Files to create:**
- `bos/core/src/ontology/serve.rs` (~400 lines) — oxigraph HTTP server with CORS for browser agents
- CLI verb `ontology serve` in `nouns/ontology.rs` (~50 lines)

---

## Implementation Order

| Phase | Innovation | Rust LOC | Go LOC | Unlocks |
|-------|-----------|---------|-------|---------|
| **1** | Auto-Ontology | 700 | 0 | All others |
| **2** | Semantic Search | 580 | 300 | SPARQL access |
| **3** | Self-Describing Workspace | 450 | 0 | Agent discovery |
| **4** | Context Grounding | 400 | 300 | Grounded RAG |
| **5** | Decision Replay | 350 | 200 | Audit trail |
| **6** | Agent Intent Protocol | 300 | 600 | Agent governance |
| **7** | Predictive Operations | 850 | 0 | Proactive operations |

**Total:** ~3,630 lines of new Rust, ~1,400 lines of new Go, 8 new `bos` noun-verb commands.

**Testing target:** Add ~50 new tests across all innovations, bringing bos from 41 to ~91 tests total.

---

## Non-Goals

- **Not replacing** the existing Go API layer — all innovations add new endpoints, not modify existing ones
- **Not changing** the ODCS workspace format — it works as-is
- **Not building** a new frontend for these features — they're consumed by agents and the existing API
- **Not requiring** PostgreSQL for read-only operations — oxigraph works with in-memory data
- **Not building** a new ontology from scratch — reusing schema.org, BPMN, ORG, PROV-O, SKOS
- **Not depending on obsr as a library** — prediction algorithms reimplemented natively in bos
- **Not supporting** non-English table/column names in v1 — English-only convention mapping

---

## Success Metrics

1. `bos ontology infer` generates correct mappings for all 7 existing tables with ≥90% accuracy (human review via `--dry-run`)
2. Semantic search returns results for relational queries that keyword search cannot answer
3. Context grounding retrieves graph-reachable context with precision ≥90% vs ~70% for vector-only RAG
4. Agent intent protocol makes every agent action auditable in <1 SPARQL query against the provenance graph
5. Self-describing workspace allows a fresh AI agent to understand any workspace in <5 SPARQL queries
6. Total test count: ≥91 tests in bos (41 existing + ~50 new)
