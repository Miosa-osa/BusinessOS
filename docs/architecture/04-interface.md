# Layer 4: Interface

> **Governing Constraint:** Shannon — Bandwidth Matching
> **Purpose:** Progressive disclosure. Receiver bandwidth matching. Noise reduction at delivery.
> **Position in onion:** Wraps L3 Composition. Wrapped by L5 Data.
> **Existing code:** `miosa_memory.Injector`, `miosa_knowledge.Context`, SignalGraph `context_tiers` table

---

## What Interface Governs

L3 Composition produces pre-structured artifacts at four granularity levels (L0–L3).

Interface governs **which** of those artifacts reach a given receiver, **when**, and **how much** — constrained by a token budget, receiver type, and task context.

The Shannon constraint is the ceiling. No receiver has infinite bandwidth. Exceeding the receiver's decoding capacity degrades signal quality — more content does not mean more information if the receiver cannot process it. A 500-line context dump delivered to a salesperson checking her next step is a Shannon violation. The Interface layer enforces the ceiling.

Three responsibilities:
1. **Tiered loading** — what to load at each granularity level and in what order
2. **Budget-aware assembly** — allocating a finite token budget across tiers and content types
3. **Audience routing** — matching output genre, depth, and format to receiver type

The Interface layer does not write signals. It does not classify or structure them. It selects pre-built artifacts from L3 and L5, assembles them within a budget, and delivers them formatted for the receiver. It is a read-only, selection-and-delivery layer.

---

## Tiered Loading Protocol

Context is loaded in four tiers, ordered by retrieval confidence and receiver utility. The guiding principle: load the minimum tier necessary to answer the query. Only escalate when the current tier's content is insufficient.

---

### L0 — HOT (always loaded, ~2K tokens)

L0 is non-negotiable. It is loaded at the start of every session, every context injection, and every agent handoff. It is the minimum viable context — the executive summary of the system's current state.

**Contents:**

| Item | Format | Source | Notes |
|------|--------|--------|-------|
| System identity | 3–5 sentences | `miosa_context.Governance` | Who this system is, what it is optimizing for |
| Active operations list | One-liner each | SignalGraph `operations` table, `status = active` | Max 10. Format: `[op_id] Name — owner — current_phase` |
| Top 5 recent decisions | One-liner each | SignalGraph `decisions`, ordered by `decided_at DESC` | Format: `[ADR-NNN] Decision — date` |
| Top 10 entities by recent mention | Name + role + last-seen date | `miosa_knowledge` entity index | Sorted by `last_mentioned_at DESC` |
| Session handoff | Previous session's terminal state | `miosa_memory.Episodic`, last session | Included only when resuming a prior session |

**Loading condition:** Always. No exception.

**Staleness rule:** L0 must be regenerated if any of the following are true:
- A state-change event has occurred since the last L0 generation (new ADR, operation status change, entity promotion/demotion)
- The session is new and the stored L0 was generated > 24 hours ago
- L0 `generated_at` predates the session start

**Eviction policy:** L0 is never evicted mid-session. If the token budget is critically constrained, L0 is the last thing cut.

---

### L1 — WARM (loaded when relevant, ~10K tokens)

L1 is loaded when the current task or query maps to a known operation, entity, or pattern type. It provides decision-quality context for the active task without loading full historical depth.

**Contents:**

| Item | Format | Source | Loading Trigger |
|------|--------|--------|-----------------|
| Full `signal.md` for relevant operations | Genre-structured document | SignalGraph `context_tiers` L1 artifact | Operation ID matched in query |
| Entity profiles with relationship summaries | Name + role + 3–5 key relationships | `miosa_knowledge.Context.for_agent/2` | Entity mentioned in query or task |
| Recent 20 decisions with rationale | One paragraph each | SignalGraph `decisions` + `decision_traces` | Query touches decision-relevant domain |
| Pattern library for current task type | Pattern name + when to use + example | `miosa_memory` pattern store | Task type classified at L2 |
| Tool / integration context | Auth status, rate limits, known issues | `miosa_context.Integrations` | Tool referenced in task |

**Loading condition:** Task classification matches one or more of the trigger conditions above.

**Selectivity:** Load only the subset of L1 content relevant to the current query. Do not load all L1 content by default. The Injector scores each L1 candidate against the query before loading (scoring formula documented in the Context Assembly Pipeline section below).

**Promotion trigger:** If L2 content is loaded for the same scope on three or more consecutive sessions, that content is promoted to L1 for future sessions. Promotion is logged to the `decision_log`.

---

### L2 — COLD (loaded on demand, ~50K tokens)

L2 is loaded only when explicitly requested or when L1 context is insufficient to resolve the query. It provides the full genre skeleton with historical depth.

**Contents:**

| Item | Format | Source | Loading Trigger |
|------|--------|--------|-----------------|
| Full `context.md` for relevant operations | Complete L2 genre artifact | SignalGraph `context_tiers` L2 artifact | Explicit drill-down request or L1 miss |
| Complete decision chains (temporal) | Ordered sequence of related ADRs | SignalGraph `decision_traces` temporal index | Query requires historical decision context |
| Deep entity relationship graphs | Full SPO subgraph for entity | `miosa_knowledge` SPARQL traversal (2-hop) | Entity relationship query |
| Historical session logs | Episodic memory summaries | `miosa_memory.Episodic` | Session continuity across > 1-day gap |
| Domain knowledge articles | Long-form reference content | `miosa_knowledge` article store | Domain knowledge query with no L1 hit |

**Loading condition:** Explicit drill-down, or L1 context produces insufficient signal for the task.

**Cost:** L2 loading is the most expensive operation in normal usage. Repeated L2 loads on the same content scope signal a promotion candidate. The Assembler logs all L2 loads with the triggering query and L1 relevance score at time of escalation.

**Demotion trigger:** L1 content not accessed in 30 days is demoted to L2. The L2 L1-artifact remains accessible on demand but is no longer loaded by default.

---

### L3 — ARCHIVE (not loaded, searched)

L3 content is never injected into context proactively. It is searched and retrieved in fragments. It represents the full historical record — superseded decisions, old sessions, complete temporal chains — that the system must be able to answer questions about but does not need in every interaction.

**Contents:**

| Item | Source | Access Method |
|------|--------|---------------|
| All episodic memories | `miosa_memory.Episodic` full store | FTS5 keyword + semantic search |
| Historical sessions | Session log archive | Temporal query (`session_date < N days ago`) |
| Superseded decisions | SignalGraph `decisions` where `status = superseded` | Explicit ADR ID lookup |
| Full temporal chains | All versions of a mutable signal | Signal ID + version range query |
| Archived operations | `operations` where `status = archived` | Entity ID or name search |

**Loading condition:** Never auto-loaded. Only retrieved via explicit search queries.

**Fragment injection:** Retrieved L3 content is injected as L0 or L1 granularity fragments of the archived artifact — not the full L3 document. This keeps the injected payload within budget. The fragment includes a provenance note: `[L3 archive — signal_id: X, retrieved: timestamp]`.

---

## Tier Summary

| Tier | Name | Token Budget | Load Policy | Staleness TTL | Promotion Trigger |
|------|------|-------------|-------------|--------------|------------------|
| L0 | HOT | ~2K | Always | 24h or state-change event | — (always loaded) |
| L1 | WARM | ~10K | Relevance-scored selection | 7 days no access | L2 loaded 3+ consecutive sessions for same scope |
| L2 | COLD | ~50K | Explicit request or L1 miss | 30 days no access | — |
| L3 | ARCHIVE | Fragment only | Explicit search, never injected | Never evicted (append-only) | — |

---

## Bandwidth Profiles

Different receivers have fundamentally different decoding capacity. Delivering a `spec` genre artifact to a salesperson is a Shannon violation — the genre has higher information density than the receiver's bandwidth supports. The Interface layer resolves the receiver's bandwidth profile before assembly and uses it to constrain content selection and format.

### Profile Definitions

| Profile | Decoding Capacity | Primary Genre Competence | Token Budget Tolerance | Format Preference |
|---------|------------------|--------------------------|----------------------|------------------|
| `human:executive` | Low-bandwidth. Decisions and outcomes only. | `brief`, `report`, `announcement` | < 2K tokens. Max 1 page. | Markdown prose, no code blocks |
| `human:operator` | Medium-bandwidth. Context + action items. | `spec`, `plan`, `status`, `report` | 2K–10K tokens | Structured markdown with headers |
| `human:technical` | High-bandwidth. Full implementation detail. | `spec`, `decision`, `report`, all genres | Up to 50K tokens | Full L2 artifact + code blocks |
| `human:salesperson` | Relationship-bandwidth. Outcome + ask only. | `brief`, `email`, `chat` | < 2K tokens. Single CTA. | Conversational markdown |
| `agent:working` | Structured data preference. Token-budgeted. | Any — uses structured map | Budget provided by caller | `Context.to_prompt/1` rendered markdown |
| `agent:orchestrator` | High-bandwidth. Multi-task coordination. | Any — full context assembly | Up to 150K tokens | Full context assembly within budget |

### Profile → Content Selection Rules

Each profile determines not just how much to load, but what to include and exclude within a given tier:

**`human:executive`**
- Include: L0 always, decisions summary, current blockers
- Exclude: Implementation details, code, full ADR text, entity relationship graphs
- Genre conversion: `spec` → `report`, `decision` → `brief`

**`human:operator`**
- Include: L0 + L1 full, relevant `plan` and `status` artifacts
- Exclude: Raw SPARQL results, embedding vectors, internal telemetry
- Genre conversion: None required

**`human:technical`**
- Include: L0 + L1 + L2 if requested, all genre artifacts as stored
- Exclude: Nothing by profile — limited only by token budget
- Genre conversion: None

**`human:salesperson`**
- Include: L0 with only relationship-relevant entities, brief summaries of active operations they own
- Exclude: Technical architecture, code, multi-hop relationship graphs, ADRs
- Genre conversion: `spec` → `brief`, `decision` → `brief`

**`agent:working`**
- Include: L0 + scored L1 subset, structured maps from `Context.for_agent/2`
- Exclude: Human-narrative prose sections (replaced with structured data)
- Genre conversion: None — agents handle all genres via structured map

**`agent:orchestrator`**
- Include: Full tiered assembly within provided budget
- Exclude: Nothing by profile — limited only by token budget
- Genre conversion: None

### Genre Competence Matrix

When the signal's genre falls outside the receiver's competence, the Interface layer **converts** to the nearest simpler genre before delivery. The original signal is stored unchanged in L5. The conversion is logged as a `genre_conversion` event in `decision_log`.

| Receiver Role | Can Decode Natively | Converts From → To |
|---------------|--------------------|--------------------|
| Executive | `brief`, `report`, `announcement` | `spec` → `report`, `plan` → `brief`, `decision` → `brief` |
| Operator | `spec`, `plan`, `status`, `report` | `brief` → `status` if more detail needed |
| Technical | All genres | — (no conversion needed) |
| Salesperson | `brief`, `email`, `chat` | `spec` → `brief`, `report` → `brief`, `decision` → `brief` |
| Working agent | All genres via structured map | `video_script` → skip (no platform rendering context) |
| Orchestrator agent | All genres | — |

---

## Context Assembly Pipeline

How the system assembles context for a query, from raw request to formatted output.

### Step 1: Determine Receiver Bandwidth Profile

```
INPUT: receiver metadata (node_id, role, channel, token_budget_override)

1. Look up receiver node in miosa_knowledge (entity type + role)
2. Map role → bandwidth profile (table above)
3. If channel is known, apply channel budget cap (see Token Budget Allocation)
4. If token_budget_override provided, use that as ceiling; profile otherwise

OUTPUT: {profile, effective_budget}
```

### Step 2: Allocate Token Budget Across Tiers

```
INPUT: {profile, effective_budget}

1. Compute tier allocations via adaptive scaling formula (see Token Budget Allocation)
2. Apply profile constraints (executive profile caps L1/L2 regardless of budget)
3. Check if L2 is warranted (based on query type and prior L1 miss signals)

OUTPUT: {l0_alloc, l1_alloc, l2_alloc, l3_enabled}
```

### Step 3: Score and Rank Available Context Candidates

For each tier, candidates are scored before selection. Scoring runs three dimensions, combined as a weighted sum:

```
score(candidate, query) =
    keyword_overlap(candidate, query) × 0.50
  + recency_score(candidate.modified_at)  × 0.30
  + importance_weight(candidate.category) × 0.20
```

**Keyword overlap:** Shared significant words (length > 3) between query tokens and candidate content, normalized by total query tokens. Implemented in `MiosaMemory.Search.score/3`.

**Recency score:** Exponential decay with 48-hour half-life:
```
recency = exp(-0.693 × age_in_hours / 48.0)
```
This means content from 48 hours ago scores 0.5; content from 1 week ago scores ~0.1. Implemented in `MiosaMemory.Injector` and `MiosaMemory.Search`.

**Importance weight:** Fixed per-category weights derived from information value:
```
decision     → 1.0
architecture → 0.95
insight      → 0.85
preference   → 0.9
bug          → 0.8
workflow     → 0.75
contact      → 0.7
general      → 0.5
note         → 0.4
```

Additionally, for L1 candidates, **graph proximity** is added as a fourth dimension when Phase 2 search is active:
```
score(candidate, query) +=
    graph_proximity(candidate, query_entities) × 0.15
```
Where `graph_proximity` is the inverse hop distance from the candidate's entities to the query's entities in the `miosa_knowledge` graph. Direct neighbors score 1.0; 2-hop neighbors score 0.5; beyond 2 hops score 0.

Across all search modes, **Reciprocal Rank Fusion** merges the ranked lists:
```
rrf_score(doc) = sum over all modes i of: 1 / (60 + rank_i(doc))
```

### Step 4: Fill Tiers from L0 Outward Until Budget Exhausted

```
FUNCTION assemble_context(query, task_type, profile, budget):

  allocation = compute_allocation(budget, profile)
  context    = []

  // L0: always load, never skip
  context += load_l0()                               // ~2K tokens
  if l0_is_stale(): regenerate_l0()

  // L1: score candidates, select within allocation
  l1_candidates = fetch_l1_candidates(query, task_type)
  l1_ranked     = score_and_rank(l1_candidates, query)
  context      += select_within_budget(l1_ranked, allocation.l1)
  // Documents are NEVER truncated. Drop whole document or include whole document.

  // L2: only if needed
  IF needs_l2(query, context):
    l2_candidates = fetch_l2_candidates(query)
    l2_ranked     = score_and_rank(l2_candidates, query)
    context      += select_within_budget(l2_ranked, allocation.l2)

  // Apply profile content filters (strip excluded sections per profile)
  context = apply_profile_filters(context, profile)

  RETURN context
```

**Key invariants:**
- Documents are never truncated mid-content. A document is either included whole or dropped.
- L0 is always the first element of the assembled context.
- Profile content filters are applied after selection, not before — selection runs on full content, then sections excluded by profile are stripped from the rendered output.

### Step 5: Format Output Matching Receiver's Preferred Genre

```
FUNCTION format_output(context, profile, signal):

  MATCH profile:
    human:executive →
      genre   = convert_to_receiver_genre(signal.genre, profile)
      content = render_human(context, genre, granularity=L1)
      RETURN markdown_document(content)     // max 1 page

    human:operator, human:technical →
      genre   = signal.genre                // no conversion
      content = render_human(context, genre, granularity=L2)
      RETURN markdown_document(content)

    human:salesperson →
      genre   = convert_to_receiver_genre(signal.genre, profile)
      content = render_human(context, genre, granularity=L1)
      RETURN markdown_document(content)     // brief format, single CTA

    agent:* →
      budget  = profile.token_budget
      ctx_map = Context.for_agent(signal.source_entity, agent_id: signal.target_agent)
      RETURN Context.to_prompt(ctx_map)     // structured map rendered as markdown
```

---

## Progressive Disclosure Protocol

The Interface layer implements progressive disclosure: start at L0, expand only when the receiver explicitly requests more. This is Shannon's constraint applied as a UX principle. The system never dumps L2 when L0 would suffice.

### Disclosure Trigger Chain

```
RECEIVER ACTION                  SYSTEM RESPONSE                   TIER SERVED
────────────────────────────────────────────────────────────────────────────────
Ask a question                → Return L0 answer                 L0
"Tell me more about X"        → Expand relevant section to L1    L0 + L1 subset
"What's the full history?"    → Load L2 for that scope           L0 + L1 + L2
"What did we decide in Q3?"   → Search L3, inject fragment       L0 + L3 fragment
"Show me the original brief"  → Retrieve L3 document by ID       L3 artifact
────────────────────────────────────────────────────────────────────────────────
```

### Drill-Down Parsing

The system recognizes the following patterns as L1 expansion triggers:
- "more detail", "expand", "tell me more about", "what happened with"
- Any follow-up question that references an entity from the previous L0 response
- A task classification that requires operational context not present in L0

L2 expansion triggers:
- "full history", "all decisions", "complete context", "everything about"
- Explicit reference to a historical time period ("what did we decide last October")
- A task that cannot be resolved with L1 content (L1 miss: zero relevant candidates)

L3 search triggers:
- Explicit reference to a superseded decision by ADR ID
- Reference to an archived operation by name
- Time-travel query: "what was the state of X on [date]"

### Disclosure State Machine

```
                  ┌─────────────────────────────────────┐
                  │              L0 (HOT)               │
                  │    Always the entry point            │
                  └──────────┬──────────────────────────┘
                             │ expansion trigger
                             ▼
                  ┌─────────────────────────────────────┐
                  │              L1 (WARM)              │
                  │    Selective, relevance-scored       │
                  └──────────┬──────────────────────────┘
                             │ explicit drill-down
                             │ or L1 miss
                             ▼
                  ┌─────────────────────────────────────┐
                  │              L2 (COLD)              │
                  │    On-demand, logged                 │
                  └──────────┬──────────────────────────┘
                             │ explicit archive query
                             ▼
                  ┌─────────────────────────────────────┐
                  │           L3 (ARCHIVE)              │
                  │    Searched, never auto-injected     │
                  └─────────────────────────────────────┘
```

**Critical constraint:** Transitions in this state machine are one-way per query. A single query is answered at one tier level. To move from L0 to L1, the receiver must send a new request. The system does not pre-emptively expand context. This enforces the Shannon ceiling on every interaction.

---

## Token Budget Allocation

### Channel Budgets

Different delivery channels impose hard ceilings regardless of profile:

| Channel | Hard Ceiling | Notes |
|---------|-------------|-------|
| Agent context window (standard) | 150K tokens | Most agent interactions |
| Agent context window (Haiku) | 50K tokens | Utility-tier agent operations |
| CLI output | 100K tokens | Terminal rendering constraint |
| Slack message | 4K tokens | Platform character limit |
| Email body | 2K tokens | Readability constraint |
| API response (default) | 32K tokens | HTTP response size |

The effective budget for assembly is `min(channel_ceiling, profile_tolerance, caller_override)`.

### Default Allocation (150K budget)

```
Budget:         150,000 tokens
─────────────────────────────────────────────────────────────────
L0  (HOT):        2,000   (1.3%)   — non-negotiable
L1  (WARM):      10,000   (6.7%)   — task-relevant subset
L2  (COLD):      30,000   (20.0%)  — on-demand retrieval
─────────────────────────────────────────────────────────────────
Reserved:       108,000   (72.0%)  — reasoning + tool use + output
```

### Adaptive Scaling Formula

For non-standard budgets:

```
L0_alloc = max(2_000, budget × 0.013)
L1_alloc = max(5_000, budget × 0.067)
L2_alloc = max(10_000, budget × 0.200)
reserved = budget - L0_alloc - L1_alloc - L2_alloc

IF budget < 17_000: drop L2 (insufficient headroom)
IF budget < 7_000:  reduce L1 to minimum needed for L0 task context
IF budget < 3_000:  L0 only
```

### Budget Split by Content Type

Within L1 allocation, the budget is sub-divided by content category:

```
L1 sub-allocation (example: 10K total):
  40%  query-relevant context   (4,000 tokens) — highest relevance-scored candidates
  30%  active context           (3,000 tokens) — currently active operations + decisions
  20%  relationship context     (2,000 tokens) — entity relationships relevant to query
  10%  temporal context         (1,000 tokens) — recent changes, session handoff
```

### Over-Budget Handling: Compression, Not Truncation

When a tier's content exceeds its allocation, the Assembler applies lossy compression rather than truncating mid-document:

1. **Score and drop:** Remove the lowest-scored candidate documents entirely. Never cut a document mid-content.
2. **Granularity downgrade:** If a document's L2 artifact exceeds budget, use its L1 artifact instead. If L1 still exceeds budget, use L0.
3. **Relationship pruning:** For entity relationship graphs, reduce from 2-hop to 1-hop traversal.

This preserves structural integrity. A receiver gets fewer complete documents rather than more broken fragments.

---

## Interface to Adjacent Layers

### From L3 (Composition)

Composition pre-computes L0/L1/L2 artifacts for every signal at intake and stores them in the SignalGraph `context_tiers` table. Interface consumes these artifacts directly. It never re-structures content — it selects and assembles.

If a tier artifact does not exist for a signal (signals ingested before Composition tier generation was in place), Interface falls back to on-the-fly rendering via `Context.to_prompt/1`.

Composition also provides the **importance classification** for each section within a genre skeleton — `[R]`, `[O]`, `[A]` flags. Interface uses these when applying profile content filters: `[O]` sections are the first candidates for exclusion when profile or budget constraints are tight.

### From L2 (Signal)

Signal classification determines the incoming `G` (genre) dimension of the signal being assembled. Interface uses `G` to:
- Select the appropriate genre template for rendering the output
- Apply the correct genre competence check against the receiver's profile
- Trigger genre conversion if the receiver cannot decode `G` natively

The S/N ratio measurement from L2 also informs assembly priority. Signals with high S/N ratios are preferred candidates when the budget forces a ranking decision between otherwise equally-scored items.

### To L5 (Data)

Interface queries L5 for context candidates during assembly. L5 provides:
- BM25 search results from FTS5 (`documents_fts`)
- SPARQL graph traversal results from `miosa_knowledge`
- MCTS expansion results from `miosa_context` (Phase 4)
- Temporal decay scores from `miosa_memory.Episodic`
- Pre-computed tier content from `context_tiers` table

Interface is a **read-only** consumer of L5 at query time. It does not write to L5. The sole exception is the promotion/demotion log, which is written to `decision_log` as a system actor event.

### To L1 (Network)

After assembly, the formatted context is handed to L1 for routing to the destination node. Interface specifies the target receiver type and selected channel. L1 handles the physical delivery — which socket, which channel, which transport protocol. Interface is agnostic to delivery mechanics.

---

## What Exists Today

The two primary modules that implement L4 Interface are `MiosaMemory.Injector` and `MiosaKnowledge.Context`. They cover the core mechanics of the layer but predate the formal tier architecture. The mapping between the specification and the code is as follows:

### `MiosaMemory.Injector`

**Location:** `miosa_memory/lib/miosa_memory/injector.ex`

**What it implements:**
- Budget-aware candidate selection: `inject_relevant/2` takes a pool of taxonomy entries and a context map, scores each entry, and returns the subset that fits within `max_tokens`
- Scoring pipeline: composite of base score (category/scope), contextual score (file match + task match + error match + session match), and recency score
- Budget enforcement: `maybe_trim_to_budget/2` greedily selects entries until `max_tokens × 4 chars` is exhausted — whole entries only, never truncated
- Prompt formatting: `format_for_prompt/1` renders selected entries as `[memory [category] [scope]] content` blocks

**Mapping to specification:**

| Spec Concept | Injector Implementation |
|-------------|------------------------|
| Tier scoring (Step 3) | `score_entry/2` — combines base, contextual, recency (weights: 0.3 / 0.5 / 0.2) |
| Recency score | `recency_score/1` — exponential decay, 48h half-life: `exp(-0.693 × age_h / 48)` |
| Budget enforcement | `maybe_trim_to_budget/2` — greedy whole-document selection within char budget |
| File-based context matching | `file_match_score/2` — extension keyword expansion + filename hit detection |
| Task-based context matching | `task_match_score/2` — word overlap ratio against task description |
| Error-based context matching | `error_match_score/2` — token overlap against error message |
| Session continuity | `session_match_score/2` — session-scoped `:context` entries score 1.0 when session_id matches |
| Prompt rendering | `format_for_prompt/1` |

**Gaps relative to specification:**
- No tier awareness (L0/L1/L2/L3) — operates on a flat pool of taxonomy entries
- No bandwidth profile routing — no receiver type concept
- No channel budget cap — caller must provide `max_tokens` externally
- No genre conversion — no genre concept at this layer
- No L2 escalation trigger — no concept of "L1 miss → load L2"

### `MiosaKnowledge.Context`

**Location:** `miosa_knowledge/lib/miosa_knowledge/context.ex`

**What it implements:**
- Agent-scoped graph context: `for_agent/2` queries the knowledge store for all facts scoped to a given agent (by subject prefix), up to `max_facts`
- Relationship extraction: separates entity-typed objects (containing `:`) into a `relationships` map grouped by predicate
- Property extraction: separates literal objects into a flat `properties` map
- Prompt rendering: `to_prompt/1` formats the structured map as a `# Knowledge Context` markdown block with Properties and Relationships sections

**Mapping to specification:**

| Spec Concept | Context Implementation |
|-------------|----------------------|
| Agent bandwidth routing | `for_agent/2` — scopes graph query to agent ID, returns structured map |
| Receiver-appropriate format | `to_prompt/1` — renders structured map as markdown for LLM injection |
| Relationship context (L1 sub-allocation) | `build_relationships/1` — groups entity-typed triples by predicate |
| Property context | `build_properties/1` — extracts literal-typed triples as flat map |
| Fact count / budget awareness | `max_facts` option (default 100) — crude budget control |

**Gaps relative to specification:**
- No tier awareness — returns raw facts, not L0/L1/L2 tier artifacts
- No relevance scoring — returns facts in query order, not scored by relevance to a query
- No token budget enforcement — `max_facts` is a count ceiling, not a token ceiling
- No staleness check — no TTL or regeneration logic
- No bandwidth profile — same output structure regardless of receiver type
- No progressive disclosure — `for_agent/2` always returns the same depth

### `MiosaMemory.Search`

**Location:** `miosa_memory/lib/miosa_memory/search.ex`

**What it implements:**
- `recall_relevant/2`: BM25-style keyword overlap scoring with recency and category importance weighting, within a token budget
- `search/3`: Multi-sort search (relevance / recency / importance) with category filtering
- Scoring: `overlap × 0.5 + recency × 0.3 + importance × 0.2`
- Category importance weights (hardcoded): `decision → 1.0`, `architecture → 0.95`, `insight → 0.85`, through `note → 0.4`

**Mapping to specification:** This is the implementation of Step 3 (Score and Rank) for the current flat-pool model. The weights in `MiosaMemory.Search.score/3` are the direct precursor to the scoring formula in the Context Assembly Pipeline section above.

### SignalGraph `context_tiers` Table

Pre-computed tier content stored at intake. This table is the bridge between L3 Composition (which generates the artifacts) and L4 Interface (which selects and assembles them).

```sql
CREATE TABLE context_tiers (
  scope         TEXT NOT NULL,    -- 'global', project name, entity name
  tier          INTEGER NOT NULL, -- 0, 1, or 2
  content       TEXT NOT NULL,    -- pre-assembled context string
  token_count   INTEGER NOT NULL,
  last_computed TEXT NOT NULL,    -- ISO 8601; staleness check uses this
  PRIMARY KEY (scope, tier)
);
```

**What it enables:** L0 and L1 loads become single-row reads from this table. No assembly required at query time for pre-computed scopes. Assembly overhead only applies to query-specific L1 candidates (entities, patterns) that vary by query.

**Gap:** Current schema uses `scope` as a text key (project name or entity name) rather than `signal_id`. The full specification calls for per-signal tier artifacts keyed by signal ID. This means the current table stores scope-level summaries, not per-signal granularity artifacts. Per-signal tiers are in scope for `miosa_context` (Phase 2).

---

## Failure Modes at the Interface Layer

| Failure | Description | Detection | Remediation |
|---------|-------------|-----------|-------------|
| **Bandwidth overload** | Assembled context exceeds receiver's decoding capacity | Token count > profile tolerance, or receiver reports confusion | Drop lowest-scored L1/L2 candidates; apply granularity downgrade cascade |
| **Under-encoding** | L0 only returned when L1 context was needed to act | Task completion failure rate + L1 miss rate above threshold | Lower L1 expansion trigger threshold; check L1 candidate scoring |
| **L0 stale** | L0 content not regenerated after state-change events | `last_computed` < most recent state-change event timestamp | Regenerate L0 before assembly; add state-change hook to L0 invalidation |
| **Genre mismatch for receiver** | Signal delivered in genre the receiver cannot decode | `receiver.role NOT IN genre.competent_receivers` | Convert to mapped genre; log `genre_conversion` event |
| **Tier miss** | L1 search returns zero relevant candidates | `l1_candidates.length == 0` after scoring | Escalate to L2 retrieval; if L2 also misses, surface L0 with a retrieval failure note |
| **Premature L2 load** | L2 loaded when L1 was sufficient (Shannon violation) | L2 load event where L1 relevance score > 0.8 | Tighten `needs_l2` threshold; log pattern for review |
| **Dual output divergence** | Human and agent outputs contradict each other for same signal | Content diff between human and agent renderings | Flag for review; human output takes precedence; log divergence |
| **Budget exhausted before L0** | Token budget too small for minimum viable context | `budget < 2_000` tokens | Return L0 only; strip all optional sections within L0; warn caller |
| **Compression cascade** | Repeated granularity downgrades produce unintelligible L0 fragments | Fragment below minimum viable signal threshold (~50 tokens) | Drop fragment entirely; log and surface only the L0 baseline |

---

## Relationship to Adjacent Layers

**From L3 (Composition):**
Composition pre-computes L0/L1/L2 artifacts for every signal at intake. Interface consumes these artifacts. It never re-structures content — it selects. Composition also tags sections with importance flags (`[R]`, `[O]`, `[A]`) that Interface uses when applying profile content filters under budget pressure.

**From L5 (Data):**
Interface queries L5 for candidates during assembly. L5 provides hybrid search results (BM25 + SPARQL + MCTS + temporal decay). Interface is a read-only consumer at query time. Promotion/demotion events are the only writes Interface originates, and those land in `decision_log` as system-actor entries.

**From L2 (Signal):**
Signal classification provides the genre and S/N ratio needed to select the appropriate rendering template and evaluate genre competence against the receiver's profile.

**To L1 (Network):**
After assembly and formatting, the context package is handed to L1 for physical routing. Interface specifies receiver type and target channel. L1 handles delivery mechanics. Interface has no dependency on transport.

---

## Related Specifications

- [Layer 3: Composition](03-composition.md) — granularity levels (L0–L3) that this layer selects from
- [Layer 5: Data](05-data.md) — `context_tiers` table, hybrid search, and temporal versioning
- [Layer 2: Signal](02-signal.md) — S=(M,G,T,F,W) classification that drives genre routing
- [Layer 1: Network](01-network.md) — receiver node properties (bandwidth, genre competence) that drive profile selection
- [Operations: Search and Retrieval](../operations/search-retrieval.md) — hybrid search pipeline in full detail
