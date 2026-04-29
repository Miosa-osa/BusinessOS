---
signal:
  mode: linguistic
  genre: spec
  type: inform
  format: markdown
  structure: spec_template
  sn_ratio: 1.0
  audience: [roberto, javaris]
  intent: "Complete package architecture specification for tools/optimal/ — the Optimal Context Engine"
---

# Optimal Context Engine — Package Architecture Specification

**ADR:** ADR-002
**Phase:** 0 (filesystem + SQLite). Designed to extend to Phase 1 (vectors) and Phase 2 (graph).
**Location:** `/Users/rhl/Desktop/OptimalOS/tools/optimal/`

---

## 1. Directory Structure

```
tools/
├── optimal.py              ← KEEP as thin shim for backward compat (3 lines, imports main())
└── optimal/
    ├── __init__.py         ← Public API: exports Engine, Signal, SearchResult
    ├── __main__.py         ← Entry point: python3 -m optimal <cmd>
    │
    ├── models/             ← Data model: Signal dimensions, validated types
    │   ├── __init__.py
    │   ├── signal.py       ← Signal dataclass + dimension enums (M, G, T, F, W)
    │   └── result.py       ← SearchResult, IndexStats, TierBundle dataclasses
    │
    ├── parse/              ← Intake: reads files, extracts structure
    │   ├── __init__.py
    │   ├── frontmatter.py  ← YAML frontmatter parser (zero deps, already exists)
    │   └── tiers.py        ← L0/L1/L2 tier extraction and auto-generation
    │
    ├── classify/           ← Signal Theory: auto-classifies S=(M,G,T,F,W)
    │   ├── __init__.py
    │   ├── classifier.py   ← Auto-classification pipeline (7-step, per 02-signal.md)
    │   ├── entities.py     ← Entity extraction (people, orgs, financial patterns)
    │   └── router.py       ← Routing rule engine (reads from topology.yaml)
    │
    ├── store/              ← Persistence: SQLite + FTS5 index
    │   ├── __init__.py
    │   ├── schema.py       ← DDL: CREATE TABLE / INDEX / TRIGGER statements
    │   ├── writer.py       ← INSERT / UPSERT / DELETE operations
    │   └── reader.py       ← SELECT, aggregate queries, stats
    │
    ├── search/             ← Retrieval: hybrid scoring, ranking
    │   ├── __init__.py
    │   ├── hybrid.py       ← BM25 + temporal decay + S/N hybrid ranker
    │   └── decay.py        ← Genre half-life table + decay function
    │
    ├── compose/            ← Output: tier cache generation, receiver-shaped output
    │   ├── __init__.py
    │   ├── l0_cache.py     ← L0 cache generator (always-loaded context)
    │   └── renderer.py     ← Format signals for a specific receiver (genre-aware)
    │
    ├── config/             ← Configuration: loaded from files, not hardcoded
    │   ├── __init__.py
    │   └── topology.py     ← Loads topology.yaml → node map, routing rules, endpoints
    │
    └── cli/                ← CLI: thin dispatchers over Engine
        ├── __init__.py
        └── commands.py     ← cmd_index, cmd_search, cmd_l0, cmd_stats, cmd_ingest
```

---

## 2. Module Specifications

### `models/signal.py` — The Core Abstraction

This is the most important file. Every other module either produces or consumes a `Signal`.

```python
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional

class Mode(str, Enum):
    LINGUISTIC = "linguistic"
    VISUAL = "visual"
    CODE = "code"
    DATA = "data"
    MIXED = "mixed"

class SignalType(str, Enum):
    DIRECT = "direct"
    INFORM = "inform"
    COMMIT = "commit"
    DECIDE = "decide"
    EXPRESS = "express"

class Format(str, Enum):
    MARKDOWN = "markdown"
    CODE = "code"
    JSON = "json"
    YAML = "yaml"
    TEXT = "text"
    # ... others

# Genre is NOT an enum — 143 values, user-extensible via genre catalogue.
# Validated as a non-empty string at runtime, checked against known catalogue.
# This is the Ashby-compliant choice: variety must be extensible without code changes.

@dataclass
class Signal:
    # Identity
    id: str                          # "sig_" + md5(rel_path)[:12]
    path: str                        # relative to OptimalOS root

    # Signal Theory dimensions S=(M,G,T,F,W)
    mode: Mode = Mode.LINGUISTIC
    genre: str = "note"              # validated against genre catalogue
    signal_type: SignalType = SignalType.INFORM
    format: Format = Format.MARKDOWN
    structure: str = ""              # genre skeleton key

    # Computed quality
    sn_ratio: float = 0.5            # resolved_dimensions / 5

    # Content
    title: str = ""
    l0_summary: str = ""             # headline (~1 sentence, ~20 tokens)
    l1_description: str = ""         # summary (~50 words, ~70 tokens)
    content: str = ""                # full body (capped at 10K chars for index)

    # Temporal
    created_at: str = ""
    modified_at: str = ""
    valid_from: str = ""
    valid_until: Optional[str] = None
    supersedes: Optional[str] = None

    # Organizational
    node: str = "unknown"            # from topology.yaml node map
    entities: list[str] = field(default_factory=list)  # entity IDs
    audience: list[str] = field(default_factory=list)  # receiver IDs

    def compute_sn_ratio(self) -> float:
        """Recompute S/N from resolved dimension count."""
        resolved = sum([
            bool(self.mode),
            bool(self.genre and self.genre != "note"),
            bool(self.signal_type),
            bool(self.format),
            bool(self.structure),
        ])
        return resolved / 5.0
```

**Why Genre is a string, not an enum:**
The genre catalogue has 143 entries across 18 categories and must grow without code changes
(Ashby's Law: the controller's variety must match the system's variety). Enums require code
changes to extend. The runtime validation (`classify/classifier.py`) checks against the
loaded catalogue, not a frozen enum.

---

### `models/result.py` — Output Shapes

```python
@dataclass
class SearchResult:
    signal: Signal
    score: float           # hybrid score
    bm25: float
    temporal: float
    age_hours: float

@dataclass
class TierBundle:
    l0: str                # ~20 tokens, always loaded
    l1: str                # ~70 tokens, loaded on demand
    l2_path: str           # path to full content (lazy-loaded)

@dataclass
class IndexStats:
    total_signals: int
    by_node: dict[str, int]
    by_genre: dict[str, int]
    entity_count: int
    audit_entries: int
```

---

### `parse/frontmatter.py` — Already Proven

Direct extraction of the existing `parse_frontmatter()` and `extract_signal_dims()` functions.
No changes to logic. Module boundary makes it testable in isolation.

**Key functions:**
- `parse_frontmatter(content: str) -> tuple[dict, str]` — returns (meta, body)
- `extract_signal_dims(meta: dict) -> dict` — returns dimension dict

---

### `parse/tiers.py` — Tier Extraction

Extracts or auto-generates L0/L1/L2 tiers from frontmatter + body.

**Key functions:**
- `extract_tiers(meta: dict, body: str, title: str) -> tuple[str, str]` — returns (l0, l1)
- `estimate_tokens(text: str) -> int` — rough token count for cache budget reporting

**Phase 1 extension point:** `tiers.py` gets a `generate_tiers_with_llm()` path that
calls an LLM to produce better summaries. The interface stays the same.

---

### `classify/classifier.py` — The 7-Step Pipeline

Implements the auto-classification pipeline from `docs/architecture/02-signal.md` exactly.
This is where Signal Theory becomes executable.

```
Step 1: FORMAT DETECTION    — file extension → Format enum
Step 2: MODE INFERENCE      — Format → Mode heuristic
Step 3: GENRE CLASSIFICATION — frontmatter → filename → section headers → heuristic
Step 4: TYPE INFERENCE      — linguistic analysis of verbs/intent
Step 5: STRUCTURE RESOLUTION — genre → structure key lookup
Step 6: S/N RATIO COMPUTATION — resolved_dims / 5
Step 7: FAILURE MODE DETECTION — run all 11 checks, log results
```

**Key class:**
```python
class SignalClassifier:
    def __init__(self, genre_catalogue: set[str]):
        self.genre_catalogue = genre_catalogue

    def classify(self, meta: dict, body: str, filepath: Path) -> Signal:
        """Full 7-step pipeline. Returns classified Signal."""

    def _detect_format(self, filepath: Path) -> Format: ...
    def _infer_mode(self, fmt: Format) -> Mode: ...
    def _classify_genre(self, meta: dict, body: str, filepath: Path) -> str: ...
    def _infer_type(self, body: str) -> SignalType: ...
    def _resolve_structure(self, genre: str) -> str: ...
    def _detect_failures(self, signal: Signal) -> list[str]: ...
```

**Why this matters vs competitors:**
Mem0 and OpenViking store embeddings of raw text. They have no concept of *what a document is*.
Our classifier knows that a file named `2026-03-15-standup.md` with `## Status` headers is a
`standup` with `inform` type and `weekly_signal` structure — before reading a single word of content.
That structural knowledge changes what gets indexed, how it decays, and how results are ranked.

---

### `classify/entities.py` — Entity Extraction

Extracts entity mentions from content. Phase 0: keyword matching against topology.yaml endpoints.
Phase 1 extension: NER model.

**Key class:**
```python
class EntityExtractor:
    def __init__(self, endpoints: dict):
        """endpoints loaded from topology.yaml"""
        self.name_to_id = self._build_lookup(endpoints)

    def extract(self, content: str) -> list[str]:
        """Returns list of entity IDs found in content."""
```

**Why topology.yaml is the source:**
Entity names live in `topology.yaml` under `endpoints:`. Loading them from there means
Alice can add a new person to the YAML and the entity extractor picks them up automatically.
No code change required.

---

### `classify/router.py` — Routing Rule Engine

Evaluates `routing_rules:` from `topology.yaml` against a classified Signal to determine
destination nodes.

```python
class Router:
    def __init__(self, routing_rules: list[dict]):
        self.rules = routing_rules

    def route(self, signal: Signal) -> list[str]:
        """
        Returns list of destination node IDs.
        Multiple rules can fire (financial data → money-revenue AND roberto).
        Critical priority rules fire immediately (algedonic bypass).
        """

    def _matches(self, rule: dict, signal: Signal) -> bool: ...
```

**This is the key organizational differentiator:**
Mem0 has no routing concept. OpenViking routes by session. We route by *signal properties*
against an organizational topology. "Customer called about pricing" → `ai-masters` AND `money-revenue`.
The routing logic lives in `topology.yaml`, not in code.

---

### `store/schema.py` — DDL Only

```python
SCHEMA_SQL = """
    CREATE TABLE IF NOT EXISTS signals ( ... );
    CREATE VIRTUAL TABLE IF NOT EXISTS signals_fts USING fts5( ... );
    CREATE TRIGGER IF NOT EXISTS signals_ai ...;
    ...
"""

def init_db(db_path: Path) -> sqlite3.Connection:
    """Create database and apply schema. Idempotent."""
```

Extracted verbatim from `optimal.py`. Isolated so schema migrations are a one-file change.

---

### `store/writer.py` — Mutations

```python
class SignalWriter:
    def __init__(self, conn: sqlite3.Connection):
        self.conn = conn

    def upsert(self, signal: Signal) -> None:
        """INSERT OR REPLACE into signals table."""

    def delete(self, signal_id: str) -> None: ...

    def upsert_entity(self, entity_id: str, name: str, now: str) -> None: ...

    def log_decision(self, actor: str, action: str, what: str, why: str,
                     signal_id: str = None) -> None:
        """Append-only audit log. Never update, never delete."""

    def clear_all(self) -> None:
        """Used by full re-index only."""
```

---

### `store/reader.py` — Queries

```python
class SignalReader:
    def __init__(self, conn: sqlite3.Connection):
        self.conn = conn

    def get(self, signal_id: str) -> Optional[Signal]: ...

    def get_by_path(self, path: str) -> Optional[Signal]: ...

    def stats(self) -> IndexStats: ...

    def recent_by_genre(self, genres: list[str], limit: int) -> list[Signal]: ...

    def top_entities(self, limit: int) -> list[tuple[str, int]]: ...

    def active_nodes(self) -> list[tuple[str, str]]: ...
```

---

### `search/hybrid.py` — The Ranker

```python
class HybridSearchEngine:
    WEIGHTS = {"bm25": 0.6, "temporal": 0.3, "sn": 0.1}

    def __init__(self, conn: sqlite3.Connection, decay: "DecayFunction"):
        self.conn = conn
        self.decay = decay

    def search(self, query: str, limit: int = 20,
               node_filter: str = None,
               genre_filter: str = None) -> list[SearchResult]:
        """
        1. FTS5 BM25 retrieval (top limit*3 candidates)
        2. Re-rank with hybrid score
        3. Return top limit results
        """
```

**Phase 1 extension point:** `HybridSearchEngine.__init__` accepts an optional
`embedder` parameter. When present, adds a `semantic` score component to the hybrid.
Weights are re-normalized. Zero changes to callers.

---

### `search/decay.py` — Temporal Decay

```python
HALF_LIVES: dict[str, float] = {
    "message": 168, "transcript": 168, "standup": 336,
    "email": 336, "note": 1440, "plan": 720,
    "spec": 4320, "decision-log": 8760, "adr": 8760,
    "pattern": 17520,
}

class DecayFunction:
    def __init__(self, half_lives: dict[str, float] = HALF_LIVES):
        self.half_lives = half_lives

    def score(self, modified_at: str, genre: str) -> float:
        """Returns 0.0–1.0. Fresh = 1.0. Old = approaches 0."""
```

Isolated because the half-life table will need tuning as the system learns
(Phase 2: feedback loop adjusts half-lives based on actual retrieval utility).

---

### `compose/l0_cache.py` — L0 Generator

```python
class L0CacheGenerator:
    def __init__(self, reader: SignalReader, cache_path: Path):
        self.reader = reader
        self.cache_path = cache_path

    def generate(self) -> str:
        """
        Produces the always-loaded L0 context cache.
        Sections: System Identity, Active Operations,
                  Recent Decisions, Key People, Recent Signals.
        Returns markdown string.
        """

    def write(self, content: str) -> Path:
        """Writes to cache_path/l0.md. Returns path."""
```

---

### `compose/renderer.py` — Receiver-Shaped Output

This module has **no equivalent in Mem0, OpenViking, or LangChain.** It is the
receiver modeling capability.

```python
class SignalRenderer:
    def __init__(self, topology: "TopologyConfig"):
        self.topology = topology

    def render_for(self, signal: Signal, receiver_id: str) -> str:
        """
        Shape signal output for a specific receiver.
        - Looks up receiver's genre_competence from topology.yaml
        - If signal.genre not in receiver's competence → re-encode to their preferred genre
        - If receiver has low technical bandwidth → strip technical sections
        Returns rendered markdown string.

        Example: render_for(spec_signal, "robert-potter") → brief (not spec)
        """

    def render_tier(self, signal: Signal, tier: str) -> str:
        """Render L0 / L1 / L2 for a signal."""
```

**Phase 0:** Basic genre competence check from topology.yaml.
**Phase 1:** LLM-assisted re-encoding to a receiver's preferred genre.

---

### `config/topology.py` — Single Source of Truth

```python
@dataclass
class TopologyConfig:
    node_map: dict[str, str]          # folder_name → node_id
    endpoints: dict[str, dict]        # endpoint_id → full endpoint data
    routing_rules: list[dict]         # ordered routing rules
    index_dirs: list[str]             # directories to index
    skip_patterns: list[str]          # paths to skip

class TopologyLoader:
    def __init__(self, topology_path: Path):
        self.path = topology_path

    def load(self) -> TopologyConfig:
        """
        Parses topology.yaml (simple key-value parser, no PyYAML).
        Derives node_map from directory_mapping section.
        Extracts endpoints, routing_rules, index_dirs.
        """
```

**This eliminates the hardcoded `NODE_MAP`, `INDEX_DIRS`, and entity name dicts from `optimal.py`.**
Alice edits `topology.yaml` → engine picks it up on next run. Zero code changes required.

---

### `cli/commands.py` — Thin Dispatchers

```python
class CLI:
    def __init__(self, engine: "Engine"):
        self.engine = engine

    def cmd_index(self) -> None: ...
    def cmd_search(self, query: str, node: str = None, genre: str = None,
                   limit: int = 20) -> None: ...
    def cmd_l0(self) -> None: ...
    def cmd_stats(self) -> None: ...
    def cmd_ingest(self, text: str, source: str = "manual") -> None: ...
```

Each command is 5–15 lines max. All logic lives in the engine modules.
This boundary makes it trivial to add a FastAPI server in Phase 1 that calls
the same `Engine` methods without touching CLI code.

---

### `__init__.py` — The Public API

```python
"""
Optimal Context Engine
======================
Signal-classified context for AI agents and humans.

Usage as library:
    from optimal import Engine
    engine = Engine()
    results = engine.search("Ed pricing call")
    engine.index()

Usage as CLI:
    python3 -m optimal index
    python3 -m optimal search "Ed pricing call"
    python3 -m optimal l0
"""

from .engine import Engine
from .models.signal import Signal, Mode, SignalType, Format
from .models.result import SearchResult, IndexStats

__all__ = ["Engine", "Signal", "Mode", "SignalType", "Format",
           "SearchResult", "IndexStats"]
```

---

### `engine.py` — The Orchestrator (implicit in structure above)

Wait — there is one file not shown: `engine.py` at the package root. This is the
**orchestration layer** that wires all modules together. No business logic lives here.

```python
class Engine:
    """
    Optimal Context Engine.
    Orchestrates: TopologyLoader → SignalClassifier → SignalWriter → HybridSearchEngine
    Single entry point for all consumers (CLI, library, future API server).
    """

    def __init__(self, root: Path = None):
        root = root or Path(__file__).parent.parent.parent
        self.config = TopologyLoader(root / "topology.yaml").load()
        self.db_path = root / ".system" / "index.db"
        self._conn = None  # lazy-opened

    # ── Core operations ──────────────────────────────────────────
    def index(self, full_rebuild: bool = True) -> int:
        """Index all markdown files. Returns count indexed."""

    def search(self, query: str, limit: int = 20,
               node: str = None, genre: str = None) -> list[SearchResult]:
        """Hybrid search. Returns ranked results."""

    def ingest(self, text: str, source: str = "manual") -> Signal:
        """Classify + route + index a new signal from raw text."""

    def l0(self) -> str:
        """Generate and return L0 cache markdown."""

    def stats(self) -> IndexStats:
        """Return index statistics."""
```

---

## 3. Dependency Flow

```
topology.yaml
      │
      ▼
config/topology.py (TopologyLoader)
      │
      ├──► classify/entities.py   (endpoint names → entity extractor)
      ├──► classify/router.py     (routing_rules → rule engine)
      └──► cli/commands.py        (index_dirs → what to crawl)

Filesystem (.md files)
      │
      ▼
parse/frontmatter.py              (meta, body)
      │
      ▼
parse/tiers.py                    (l0, l1)
      │
      ▼
classify/classifier.py            (Signal with all dims resolved)
      │
      ├──► classify/entities.py   (entity list)
      └──► classify/router.py     (destination nodes)
                  │
                  ▼
            store/writer.py       (upserts to SQLite)
                  │
                  ▼
            store/schema.py       (FTS5 triggers fire automatically)

Query
      │
      ▼
search/hybrid.py
      ├──► SQLite FTS5 (BM25 scores via store/reader.py)
      └──► search/decay.py        (temporal decay per genre)
                  │
                  ▼
            list[SearchResult]    (ranked, returned to CLI or caller)
```

**Rule:** Dependencies only flow DOWN this diagram. No module in `search/` imports from `classify/`.
No module in `cli/` imports from `store/`. All cross-cutting state flows through `engine.py`.

---

## 4. How We Compare to Competitors

| Capability | OpenViking | Mem0 | LangChain | **Optimal Engine** |
|------------|-----------|------|-----------|-------------------|
| Context model | Virtual filesystem | Memory records | Documents/chains | **Signals: S=(M,G,T,F,W)** |
| Classification | File type only | LLM-based extraction | None | **7-step pipeline, Ashby-compliant** |
| Genre awareness | None | None | None | **143 genres, skeletal validation** |
| Receiver modeling | None | None | None | **genre_competence per receiver** |
| Routing | Session-based | None | None | **topology.yaml rule engine** |
| Temporal decay | None | None | None | **Genre-specific half-lives** |
| Tier loading | None | None | None | **L0/L1/L2/L3 pre-computed** |
| Config source | Code | Code | Code | **topology.yaml (Alice edits YAML)** |
| Stdlib only (core) | No | No | No | **Yes** |
| Phase 0 → Phase 1 | N/A | N/A | N/A | **Pluggable ports: add embedder to HybridSearchEngine** |

**The fundamental difference:** competitors treat context as a retrieval problem (how do I find
the right chunk of text?). We treat context as a communication problem (how do I send the right
Signal to the right receiver in the right genre at the right time?). The architecture reflects
that: `receiver.py`, `router.py`, and `classifier.py` have no equivalent in any competitor.

---

## 5. Extension Points (Phase 1 and Phase 2)

### Phase 1: Vector Embeddings
- Add `search/vector.py` with `VectorSearchEngine` (same interface as FTS5 path)
- `HybridSearchEngine.__init__` accepts optional `embedder` — weights auto-rebalance to include semantic score
- `store/schema.py` gets a `vectors` table (or sqlite-vec extension)
- Zero changes to `classify/`, `parse/`, `config/`, `compose/`

### Phase 1: Ingest Raw Text (current `ingest` command)
- `classify/classifier.py` already handles unstructured text (Step 4 linguistic analysis)
- `parse/frontmatter.py` returns empty meta for non-frontmatter input
- `ingest()` on `Engine` works today; enhancement = better genre heuristics

### Phase 2: Graph/OWL
- Add `graph/` package
- `store/reader.py` grows a `get_edges()` method
- `engine.py` gets `graph_search()` method
- Everything else unchanged

### Future: FastAPI server
- Add `server/` package with a single `app.py`
- All route handlers call `Engine` methods
- `cli/commands.py` unchanged — same methods, different caller

---

## 6. File Count and Size Expectations

| Module | Est. Lines | Responsibility |
|--------|-----------|----------------|
| `engine.py` | 80–100 | Orchestration only |
| `models/signal.py` | 60–80 | Data model |
| `models/result.py` | 30–40 | Output shapes |
| `parse/frontmatter.py` | 50–60 | Already written |
| `parse/tiers.py` | 40–50 | Already written |
| `classify/classifier.py` | 100–120 | Core pipeline |
| `classify/entities.py` | 40–50 | Already written |
| `classify/router.py` | 60–80 | New logic |
| `store/schema.py` | 60–70 | Already written |
| `store/writer.py` | 60–80 | Already written |
| `store/reader.py` | 60–80 | Already written |
| `search/hybrid.py` | 60–80 | Already written |
| `search/decay.py` | 30–40 | Already written |
| `compose/l0_cache.py` | 60–80 | Already written |
| `compose/renderer.py` | 50–70 | New logic |
| `config/topology.py` | 80–100 | New logic |
| `cli/commands.py` | 60–80 | Thin wrappers |
| `__init__.py` + `__main__.py` | 20–30 | Plumbing |
| **Total** | **~870–1100** | vs 671 lines in one file |

The total line count is modestly higher but distributed across 17 focused files with clear
single responsibilities. The monolith's 671 lines had no testable units. These 17 files
each have one.

---

## 7. Migration Path

**Step 1 (30 min):** Create package skeleton — all `__init__.py` files, empty modules.

**Step 2 (60 min):** Extract already-written code into correct modules:
- `parse_frontmatter`, `extract_signal_dims` → `parse/frontmatter.py`
- `extract_tiers` → `parse/tiers.py`
- `extract_entities` → `classify/entities.py`
- `temporal_decay`, `HALF_LIVES` → `search/decay.py`
- `init_db` DDL → `store/schema.py`; `index_file` mutations → `store/writer.py`; queries → `store/reader.py`
- `cmd_search` + hybrid scoring → `search/hybrid.py`
- `cmd_l0` → `compose/l0_cache.py`

**Step 3 (60 min):** Write new code:
- `config/topology.py` — YAML loader (replaces hardcoded `NODE_MAP`)
- `classify/classifier.py` — 7-step pipeline (formalizes what was implicit)
- `classify/router.py` — rule engine over `routing_rules:` from topology.yaml
- `compose/renderer.py` — receiver-shaped output (new capability)
- `engine.py` — orchestrator that wires everything

**Step 4 (30 min):** Wire `cli/commands.py` → `Engine`, write `__main__.py`,
keep `tools/optimal.py` as a 3-line backward-compat shim.

**Step 5 (30 min):** Verify all existing CLI commands work identically.

Total: ~3.5 hours. No behavioral changes in Phase 0. Infrastructure for Phase 1+.
