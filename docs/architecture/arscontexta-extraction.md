# Ars Contexta → OptimalOS Extraction Map

> Complete analysis from 8 parallel extraction agents. What to steal, what we already beat.

## TL;DR

**Ars Contexta governs the INTERNAL side** (how you organize what you know).
**Signal Theory governs the OUTPUT side** (what you send to people).
They are complementary, not competing. We steal their internal organization patterns and graft them onto our superior engine.

---

## 1. FEATURE COMPARISON MAP

| Feature | Ars Contexta | OptimalOS | Winner | Notes |
|---------|-------------|-----------|--------|-------|
| **Classification** | None (manual tagging) | S=(M,G,T,F,W) auto-classify | **OptimalOS** | Signal Theory is far superior |
| **Routing** | Manual folder placement | 12-node auto-routing + cross-ref | **OptimalOS** | AC has 3 spaces, we have 12 nodes |
| **Search** | FTS5 + optional `qmd` semantic | FTS5 + vector + RRF hybrid | **OptimalOS** | Our hybrid search is leagues ahead |
| **Knowledge Graph** | Wiki-links between notes | OWL-materialized graph + edges | **OptimalOS** | We have entities, edges, probabilities |
| **Simulation** | None | MCTS + Monte Carlo + BFS | **OptimalOS** | AC can't simulate anything |
| **Quality Gates** | Write-validation hook | S/N ratio gates + FailureModes audit | **OptimalOS** | We reject/quarantine bad signals |
| **Genre System** | None | 27 genres with skeletons per receiver | **OptimalOS** | AC has no concept of receiver adaptation |
| **Processing Pipeline** | 6R (Record→Reduce→Reflect→Reweave→Verify→Rethink) | Intake (classify→route→write→index) | **Draw** | Different purposes — steal their Reflect/Reweave/Rethink |
| **Self-Space** | `self/identity.md`, `self/methodology.md`, `self/goals.md` | None (CLAUDE.md only) | **Ars Contexta** | We need an agent self-model |
| **Note Templates** | Domain-derived via 8 dimensions | Genre skeletons (static) | **Draw** | Their derivation is interesting |
| **Hooks** | 4 data-integrity hooks (session-orient, write-validate, auto-commit, session-capture) | 13 agent-instrumentation hooks | **Complementary** | We have MORE hooks but ZERO vault-integrity hooks |
| **Memory Extraction** | Via `/reduce` phase | `MemoryExtractor` (6-category LLM) | **OptimalOS** | Our extractor is more structured |
| **Reweaving** | `/reweave` backward pass | None | **Ars Contexta** | We need this badly |
| **Health Diagnostics** | Condition-based maintenance triggers | `mix optimal.stats` only | **Ars Contexta** | We need `mix optimal.health` |
| **Research Backing** | 249 interconnected claims | Signal Theory paper | **Draw** | Both are rigorous |
| **Tiered Loading** | None | L0/L1/L2 with token budgets | **OptimalOS** | AC loads everything or nothing |
| **Agent Orchestration** | `/ralph` queue-based, fresh context per phase | 52+ specialized subagents | **OptimalOS** | Our agent system is industrial |
| **Derivation Engine** | 8-dimension config → generates entire system | None | **Ars Contexta** | Unique — we don't need it (we already HAVE a system) |
| **Graph Triangle Detection** | Finds synthesis opportunities | None | **Ars Contexta** | Steal this for knowledge graph |

---

## 2. WHAT TO STEAL (Prioritized)

### Priority 1: Reweaving (`/reweave` backward pass)
**What it does**: When new information arrives, revisits older notes that reference the same topics and updates them with new context. A "backward pass" through the knowledge graph.

**Why we need it**: Our intake is forward-only — signals go in, get indexed, done. But when Alice learns "Ed actually wants $3K not $2K", the old signal about $2K pricing still exists unchanged. Reweaving would propagate the update.

**Implementation**: `mix optimal.reweave "Alice pricing"` — finds all contexts mentioning the entity/topic, presents them for review, suggests updates. Uses knowledge graph edges to find related contexts.

**Where it goes**: New Mix task + new `OptimalEngine.Reweaver` module. Uses `Store.raw_query` + graph traversal.

### Priority 2: Write-Validation Hook
**What it does**: Before ANY write to the vault, validates the file meets structural requirements — correct frontmatter, required fields present, no orphan links, genre skeleton followed.

**Why we need it**: We have 13 hooks for agent instrumentation but ZERO for data integrity. A bad signal file can corrupt the index silently.

**Implementation**: PostToolUse hook on Write/Edit that checks:
- Frontmatter is valid YAML
- Required fields present (title, date, genre for signals)
- Node reference is valid (one of 12 known nodes)
- No broken internal references

**Where it goes**: `~/.claude/hooks/write-validate.py` + hook config in `settings.json`

### Priority 3: Health Diagnostics (`mix optimal.health`)
**What it does**: Condition-based health check across the entire system. Not just stats — actionable diagnostics.

**Why we need it**: `mix optimal.stats` shows counts. We need: orphaned contexts (no edges), stale signals (>30 days untouched), missing cross-references, FTS/index drift, entity merge candidates, node imbalance warnings.

**Implementation**: New Mix task that runs ~10 diagnostic queries and outputs a health report with severity levels and fix suggestions.

**Where it goes**: New `engine/lib/mix/tasks/optimal.health.ex`

### Priority 4: Graph Triangle Detection
**What it does**: Finds three nodes A→B, B→C, A→C where a synthesis opportunity exists. If Alice has signals about "Ed + pricing" and "pricing + AI Masters" and "AI Masters + Ed", there's a triangle — a natural synthesis point.

**Why we need it**: Our knowledge graph has 1045 edges but we never analyze topology for patterns. Triangles reveal where contexts should be combined into higher-level insights.

**Implementation**: SQL query on edges table: find entity pairs that share neighbors. Surface as `mix optimal.graph triangles`.

**Where it goes**: Add to existing `OptimalEngine.Bridge.Knowledge` or new `graph.triangles` Mix task.

### Priority 5: Agent Self-Space
**What it does**: AC has `self/identity.md` (who am I), `self/methodology.md` (how I work), `self/goals.md` (what I'm optimizing for). The agent has a persistent self-model it can reference and update.

**Why we need it**: Our CLAUDE.md is static — it describes the system but not the agent's evolving understanding. When the agent learns "Alice prefers single bundled PRs", that goes in Claude memory but not in a structured self-model.

**Implementation**: Add `self/` directory with:
- `self/identity.md` — "I am OptimalOS, Alice's cognitive engine..."
- `self/methodology.md` — "I use Signal Theory S=(M,G,T,F,W)..."
- `self/patterns.md` — Learned patterns from SICA

**Where it goes**: New `self/` directory at OptimalOS root. Referenced in CLAUDE.md.

### Priority 6: Session Capture Hook
**What it does**: On session stop, automatically commits a summary of what was discussed/decided/changed to the ops space.

**Why we need it**: When a conversation ends, the context evaporates. AC captures it automatically. Our `Session.commit/1` exists but isn't triggered automatically.

**Implementation**: Add a Stop hook that calls `mix optimal.session commit` with auto-generated summary.

**Where it goes**: Hook config + existing Session module integration.

---

## 3. WHAT WE ALREADY DO BETTER

### Signal Theory Classification
AC has no classification at all. Users manually decide where notes go. We auto-classify across 5 dimensions and 27 genres. This is our biggest structural advantage.

### Simulation Engine
AC can't answer "What if Ed leaves?" — it's a note-taking system. We have MCTS (32-iteration action planning), Monte Carlo (1000-simulation probability distributions), and BFS graph traversal (3-depth impact analysis). No competitor has this.

### Hybrid Search with RRF
AC uses basic FTS5 or optional `qmd` semantic search. We fuse FTS5 + vector similarity + temporal decay + S/N boost + graph boost via Reciprocal Rank Fusion. Our search is dramatically better.

### 12-Node Routing with Cross-References
AC has 3 spaces (self/notes/ops). We have 12 semantic nodes with auto-routing, cross-referencing, and a routing table that maps keywords to destinations. A signal about "Ed pricing" automatically goes to ai-masters AND money-revenue.

### Quality Gates
AC validates file structure. We validate signal quality — S/N ratio < 0.3 gets rejected, routing failures get quarantined, bandwidth overload triggers L1 truncation. This is Signal Theory in action.

### Knowledge Graph with Probabilities
AC uses wiki-links between notes. We have a full entity-edge graph with probability weights, OWL inference, and the ability to run Monte Carlo simulations over edge probabilities. Our graph is a computational structure, not just navigation.

### Industrial Agent System
AC has one subagent (`knowledge-guide.md`) and queue-based orchestration via `/ralph`. We have 52+ specialized agents with parallel dispatch, batch processing, and file-based communication. Different scale entirely.

---

## 4. ARS CONTEXTA ARCHITECTURE (Reference)

### Three-Space Model
| Space | Purpose | Growth | OptimalOS Equivalent |
|-------|---------|--------|---------------------|
| `self/` | Agent identity, methodology, goals | Slow (tens of files) | **MISSING** — need to add |
| `notes/` | Knowledge graph (main content) | Steady (10-50/week) | Our 12 numbered folders |
| `ops/` | Queue state, sessions, operations | Fluctuating | `.system/` directory |

### 15 Kernel Primitives
1. Three-space separation (self/notes/ops)
2. CLAUDE.md as system instruction
3. Navigation map (index of all content)
4. Note templates (domain-derived)
5. Processing pipeline (6R)
6. Automation hooks (4 types)
7. Session lifecycle (orient → work → capture)
8. Queue-based task orchestration
9. Subagent spawning per phase
10. Wiki-link connections between notes
11. Reduce extraction (key insights)
12. Reflect connections (find links)
13. Reweave backward pass (update old notes)
14. Verify quality checks
15. Rethink assumption challenges
+1. Derivation engine (generates the whole system)

### 6R Pipeline
| Phase | What It Does | OptimalOS Equivalent |
|-------|-------------|---------------------|
| **Record** | Zero-friction capture | `mix optimal.ingest` |
| **Reduce** | Extract insights | `MemoryExtractor.extract/1` |
| **Reflect** | Find connections | `Bridge.Knowledge.graph_boost` |
| **Reweave** | Update older notes with new context | **MISSING** — Priority 1 steal |
| **Verify** | Quality checks | `BridgeSignal.audit/1` |
| **Rethink** | Challenge assumptions | **MISSING** — future |

### 4 Hooks
| Hook | Trigger | What It Does | OptimalOS Has? |
|------|---------|-------------|---------------|
| Session Orient | SessionStart | Load context, show nav map | Partial (memory injection) |
| Write Validate | PostToolUse(Write) | Check file integrity | **NO** — Priority 2 steal |
| Auto Commit | PostToolUse(Write) | Git commit on vault changes | No (manual commits) |
| Session Capture | Stop | Save session summary | Partial (Session.commit) |

### 8 Configuration Dimensions (Derivation Engine)
The derivation engine uses conversation to score the user across 8 dimensions, then generates a system configuration. We don't need this (we already have our system), but the dimensions are interesting as a validation framework:

1. **Domain complexity** — How interconnected is the knowledge?
2. **Processing depth** — How much reduction/reflection needed?
3. **Temporal sensitivity** — How much does recency matter?
4. **Collaboration level** — Solo vs team knowledge?
5. **Formality** — Academic vs conversational?
6. **Volume** — How many notes per week?
7. **Connection density** — How linked should notes be?
8. **Automation tolerance** — How much auto-processing?

---

## 5. IMPLEMENTATION ROADMAP (Updated Post-Deep-Extraction)

### Phase 1: Data Integrity + Health (1 session)
- [ ] Write-validation hook (`write-validate.py`) — YAML frontmatter + required fields
- [ ] `mix optimal.health` — 8 diagnostic categories (orphans, stale, cross-ref integrity, FTS drift, entity merge candidates, node imbalance, routing coverage, financial sync)
- [ ] Auto-commit hook for signal files

### Phase 2: Knowledge Enhancement (1 session)
- [ ] `mix optimal.reweave "topic"` — backward pass through graph (their Priority 1)
- [ ] `mix optimal.reflect` — find unlinked relationships across nodes
- [ ] Graph triangle detection in `mix optimal.graph triangles`
- [ ] Agent self-space (`self/` directory with identity, methodology, patterns)

### Phase 3: Learning Loop (1 session)
- [ ] `mix optimal.remember` — 3-mode friction capture (explicit, contextual, session-mining)
- [ ] Escalation: 3+ same-category observations triggers review prompt
- [ ] `mix optimal.rethink` — formalized Friday double-loop with evidence-based proposals

### Phase 4: Orchestration (1 session)
- [ ] Ralph-style subagent isolation for brain dump pipeline
- [ ] Structured handoff blocks for agent communication
- [ ] Queue state with `current_phase` / `completed_phases` for resumability
- [ ] Arithmetic verification (`subagents_spawned == tasks_processed`)

### Phase 5: Quality Gates (1 session)
- [ ] `mix optimal.verify` — cold-read test (predict content from title/L0, score match)
- [ ] Session capture hook (auto-commit session summary on Stop)
- [ ] Session orient hook (load recent context + maintenance thresholds on SessionStart)

---

## 6. DEEP EXTRACTION: RALPH ORCHESTRATOR

The deepest finding. Ralph is AC's orchestration engine and the most interesting pattern.

### Mandatory Subagent Spawning
Every pipeline task MUST spawn a fresh subagent via the Task tool. The lead session's ONLY job: read queue → spawn subagent → evaluate return → update queue → repeat.

**Why**: LLM attention degrades as context fills. Fresh context per phase = each phase gets max reasoning quality.

### RALPH HANDOFF Protocol
Every subagent returns a structured block:
```
=== RALPH HANDOFF ===
Target: {description}
Work Done: [list]
Learnings: [Friction|Surprise|Methodology|Process gap]: {desc} | NONE
Queue Updates: [state changes]
=== END HANDOFF ===
```

### Parallel Mode
Up to 5 concurrent workers with "sibling awareness". Two-phase gate: Phase A (parallel work), Phase B (cross-connect validation, must wait for all A to complete).

### Verification as Arithmetic
`subagents_spawned == tasks_processed` — hard gate. No vibes. Countable proof.

### What to Steal for OptimalOS
- The weekly brain dump pipeline (classify → route → write → cross-reference) should use subagent isolation per node
- Structured handoff blocks instead of free-form file communication
- Queue state with `current_phase` / `completed_phases` for resumability
- Arithmetic verification for orchestration tasks

## 7. DEEP EXTRACTION: THE REMEMBER→RETHINK LOOP

AC's learning system is more sophisticated than `tasks/lessons.md`.

### /remember (Three Modes)
1. **Explicit**: `/remember "check duplicates before creating"` — parses friction, creates methodology note
2. **Contextual**: `/remember` (no args) — scans conversation for correction signals (redirections, frustration, preference statements)
3. **Session mining**: `/remember --mine-sessions` — scans stored session transcripts for uncaptured friction patterns

### Escalation Pattern
3+ observations in same category → triggers `/rethink` suggestion. This is the missing link in our `tasks/lessons.md` — we capture corrections but never escalate patterns.

### /rethink (Scientific Method for Systems)
Six phases:
0. Drift check (methodology docs vs actual config)
1. Triage (PROMOTE / IMPLEMENT / METHODOLOGY / ARCHIVE / KEEP PENDING)
2. Methodology updates
3. Pattern detection (minimum 3 observations)
4. Proposal generation (2+ evidence sources, risk assessment, reversibility)
5. Approval & implementation (NEVER auto-implements)

### What to Steal for OptimalOS
- Three-mode `/remember` → `mix optimal.remember` with explicit, contextual, and session-mining modes
- Escalation at 3+ same-category corrections
- `/rethink` as a formalized version of our Friday double-loop review

## 8. DEEP EXTRACTION: DERIVATION ENGINE MECHANICS

### Signal Extraction Over Forms
The engine NEVER asks "pick from these options." It asks "Tell me about what you want to track, remember, or think about" and listens for signals.

### Confidence Scoring
| Level | Weight | Definition |
|-------|--------|-----------|
| HIGH | 1.0 | Explicit statement with concrete examples |
| MEDIUM | 0.6 | Implicit preference |
| LOW | 0.3 | Ambiguous, single mention |
| INFERRED | 0.2 | Cascaded from resolved dimensions |

Resolution at cumulative confidence >= 1.5. One HIGH + one MEDIUM = 1.6 = resolved.

### Anti-Signals
"I want Zettelkasten" ≠ atomic granularity (they may just want the label). Follow-up: "Walk me through your actual last week of note-taking."

### Context Resilience
`derivation.md` written FIRST — if context window compacts during generation, all subsequent steps re-read this file as source of truth. Smart engineering.

### What to Steal for OptimalOS
- Not the derivation engine itself (we already HAVE our system), but the confidence-scoring pattern for classification
- Anti-signal awareness for the Classifier module
- Context resilience pattern (write intermediate state to disk during long operations)

## 9. DEEP EXTRACTION: 26 SKILLS MAPPED

| AC Skill | What It Does | OptimalOS Equivalent | Gap? |
|----------|-------------|---------------------|------|
| `/setup` | Derive system from conversation | CLAUDE.md (manual) | N/A |
| `/help` | State-aware contextual help | Static help | Minor |
| `/health` | 8-category diagnostics | `mix optimal.stats` | **YES** |
| `/recommend` | Architecture advisory | None | Minor |
| `/architect` | Evidence-based evolution | Weekly/monthly review (manual) | Medium |
| `/ask` | Query 249 research claims | None | Minor |
| `/tutorial` | Interactive onboarding | None | Minor |
| `/reseed` | Re-derive when drift accumulates | Monthly review (manual) | Medium |
| `/reduce` | Extract atomic insights | `mix optimal.ingest` | Partial |
| `/reflect` | Find connections + update MOCs | None | **YES** |
| `/reweave` | Backward pass — update old notes | None | **YES** |
| `/verify` | Cold-read test + schema + health | None | **YES** |
| `/ralph` | Queue orchestrator (subagent spawning) | None | **YES** |
| `/pipeline` | End-to-end processing chain | `mix optimal.ingest` (single-step) | Medium |
| `/next` | 14-signal priority cascade | rhythm/boot.md (manual) | Medium |
| `/learn` | Research with provenance | None | Minor |
| `/remember` | 3-mode friction capture | `tasks/lessons.md` | **YES** |
| `/rethink` | Scientific method for system evolution | Friday review (manual) | **YES** |
| `/seed` | Queue intake with duplicate detection | `mix optimal.ingest --file` | Partial |
| `/stats` | Vault metrics | `mix optimal.stats` | Exists |
| `/graph` | Graph topology analysis | Knowledge graph (partial) | Medium |
| `/tasks` | Task stack + pipeline queue | `tasks/todo.md` | Partial |
| `/validate` | Schema compliance | None | Medium |
| `/refactor` | Cascading config changes | Manual | Minor |
| `/add-domain` | Multi-domain extension | Manual folder creation | Minor |
| `/upgrade` | Version management | N/A | N/A |

## 10. ARCHITECTURAL PHILOSOPHY

| Dimension | Ars Contexta | OptimalOS |
|-----------|-------------|-----------|
| Core metaphor | Knowledge graph (notes=nodes, links=edges) | Decision tree library (folders=categories, files=patterns) |
| Design approach | Generated from conversation via derivation | Hand-crafted by architect (Alice) |
| Target user | Anyone (researchers, therapists, PMs) | One person (Alice) across all life domains |
| Theory | 249 cognitive science claims | Signal Theory (4 constraints, 6 principles, 11 failure modes) |
| File philosophy | Atomic: one claim per file | Comprehensive: one context.md per domain node |
| Navigation | Wiki-links + MOC hierarchy | Folder numbers + engine search + routing table |
| Processing | 6-phase pipeline with subagent isolation | Single-step ingest with auto-classification |
| Self-evolution | observations → tensions → /rethink → proposals | lessons.md + weekly/monthly reviews |
| Agent model | 1 knowledge guide + subagents per phase | 52+ specialized agents dispatched by context |

**The fundamental difference**: AC builds a **knowledge graph** where value compounds through connections between atomic notes. OptimalOS builds a **decision tree library** where value compounds through comprehensive context per domain node. They optimize for discovery and synthesis. We optimize for routing and execution.

## 11. COMPETITIVE POSITIONING

After this extraction, OptimalOS's position is clear:

**Ars Contexta** = Best-in-class personal knowledge management via Claude Code.
**OptimalOS** = Signal-theoretic cognitive operating system with simulation capabilities.

They're playing chess. We're playing 4D chess with Monte Carlo.

What AC does well (and we should steal): data integrity hooks, reweaving, health diagnostics, agent self-model.

What they can never match: Signal Theory classification, MCTS simulation, Monte Carlo probability analysis, 12-node semantic routing, hybrid search with RRF, knowledge graph with edge probabilities, industrial agent orchestration.

The agent-as-markdown pattern validates our architecture completely. AC proves the market exists. We prove the ceiling is much higher.
