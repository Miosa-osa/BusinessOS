# The Optimal Data Model
## Complete Type System, Storage Patterns, and Naming Conventions
### Every Data Type. Every Storage Pattern. Nothing Missing.

---

## Governing Principle

> **Every piece of data in the system is a Signal.**
> A Signal has 5 resolved dimensions: S = (Mode, Genre, Type, Format, Structure).
> Unresolved dimensions = noise. Resolved dimensions = context.

This model defines every data type, how it's classified, where it's stored, how it's named, and how it relates to everything else. The goal: any piece of information can be classified, routed, stored, retrieved, composed, and delivered through one universal system.

---

## 1. THE SIGNAL — The Atomic Unit

Every artifact in the system is a Signal. No exceptions.

### Signal Schema

```yaml
signal:
  # Identity
  id: sig_<ulid>                    # Unique, sortable, time-encoded
  created_at: <ISO 8601>            # When it was created
  version: <integer>                # Append-only version number
  supersedes: <signal_id | null>    # Previous version (temporal chain)

  # Classification — S=(M,G,T,F,W)
  mode: <mode_enum>                 # How is it perceived?
  genre: <genre_enum>               # What conventionalized form?
  type: <type_enum>                 # What speech act?
  format: <format_enum>             # What container?
  structure: <structure_enum>       # What internal skeleton?

  # Routing
  source_node: <node_id>            # Where it came from
  target_nodes: [<node_id>, ...]    # Where it's routed to
  target_endpoints: [<endpoint_id>, ...]  # Who receives it

  # Metadata
  sn_ratio: <float 0.0-1.0>        # Signal-to-noise quality score
  tier: <L0 | L1 | L2 | L3>        # Disclosure depth
  sensitivity: <public | internal | confidential | restricted>
  ttl: <duration | null>            # Time-to-live (null = permanent)
  tags: [<string>, ...]             # Freeform tags for additional filtering

  # Content
  body: <string>                    # The actual content
  summary_l0: <string>              # ~50 words — headline
  summary_l1: <string>              # ~200 words — key facts
  # L2 = full body, L3 = body + all versions + decision traces

  # Relations
  parent: <signal_id | null>        # Parent signal (thread/conversation)
  references: [<signal_id>, ...]    # Signals this references
  cross_refs: [<node_id>, ...]      # Nodes this touches beyond primary route

  # Extraction
  decisions: [{decision, by, date, rationale}]
  actions: [{action, owner, deadline, status}]
  people: [<endpoint_id>, ...]      # People mentioned
  financial: [{amount, currency, context}]  # Financial data extracted
```

---

## 2. DIMENSION ENUMS — The Complete Classification Vocabulary

### Mode (M) — How is it perceived?

| Value | Description | Examples |
|-------|-------------|---------|
| `linguistic` | Text or speech — sequential, propositional | Docs, messages, transcripts, emails |
| `visual` | Images, diagrams, video — relational, spatial | Architecture diagrams, screenshots, videos |
| `code` | Source code, configs, executable structures | .ex, .ts, .yaml, scripts |
| `data` | Structured records — machine-readable | JSON, CSV, database rows, API responses |
| `audio` | Sound — speech, music, ambient | Voice notes, recordings, podcasts |
| `mixed` | Multiple modes combined | Slide decks, notebooks, annotated screenshots |

### Genre (G) — What conventionalized form?

**Operational Genres (day-to-day business):**

| Genre | Purpose | Primary Receiver | Skeleton Sections |
|-------|---------|-----------------|-------------------|
| `spec` | Define requirements precisely | Engineer, agent | Goal, Requirements, Constraints, Architecture, Acceptance |
| `brief` | Concise directive for action | Salesperson, executive | Objective, Key Messages, CTA, Supporting Materials |
| `plan` | Structured execution steps | Operator, self | Objective, Non-Negotiables, Time Blocks, Dependencies, Success |
| `status` | Current state report | Stakeholder | Summary, Progress, Blockers, Next Steps |
| `report` | Findings, metrics, analysis | Decision-maker | Executive Summary, Methodology, Findings, Recommendations |
| `decision` / `adr` | Record a choice with rationale | Team, future self | Context, Decision, Alternatives, Consequences |
| `review` | Evaluative assessment | Author, team | Summary, Findings, Rating, Recommendations |
| `guide` | How-to instructions | New user, operator | Prerequisites, Steps, Troubleshooting, References |
| `script` / `runbook` | Step-by-step procedure | Operator, CI system | Trigger, Steps, Rollback, Verification |
| `template` | Reusable structure with slots | Author, agent | [varies by template purpose] |

**Communication Genres:**

| Genre | Purpose | Primary Receiver | Skeleton Sections |
|-------|---------|-----------------|-------------------|
| `email` | Async directed message | Named recipient | Subject, Body, CTA |
| `chat` | Conversational exchange | Human, agent | [freeform with context] |
| `meeting_notes` / `transcript` | Record of discussion | Attendees | Participants, Key Points, Decisions, Actions, Open Questions |
| `announcement` | Broadcast notification | Wide audience | What, Why, Impact, Action Required |
| `proposal` | Pitch requesting approval | Approver | Problem, Solution, Cost, Timeline, Ask |
| `pitch` | Persuasive presentation | Prospect, investor | Hook, Problem, Solution, Proof, CTA |

**Content Genres:**

| Genre | Purpose | Primary Receiver | Skeleton Sections |
|-------|---------|-----------------|-------------------|
| `article` | Long-form explanatory | General reader | Hook, Thesis, Body, Conclusion |
| `video_script` | Narration + visual cues | Presenter, editor | Sections with talking points + B-roll notes |
| `social_post` | Short-form public content | Public audience | Hook, Body, CTA, Hashtags |
| `course_module` | Instructional unit | Learner | Objectives, Lessons, Deliverable |
| `presentation` | Talk/slide content | Live audience | Title, Sections, Key Takeaways |
| `ad_script` | Advertising content | Prospect | Hook, Problem, Solution, CTA |
| `vsl` | Video sales letter | Prospect | Problem, Bridge, Solution, Proof, Offer, CTA |

**Operational/Financial Genres:**

| Genre | Purpose | Primary Receiver | Skeleton Sections |
|-------|---------|-----------------|-------------------|
| `invoice` | Financial billing | Accountant, client | Items, Amounts, Terms, Due Date |
| `contract` | Binding agreement | Parties, legal | Parties, Terms, Scope, Payment, Signatures |
| `case_study` | Proof of results | Prospect | Client, Problem, Solution, Result |
| `role_description` | Job/role definition | Candidate, team | Title, Responsibilities, Requirements, Compensation |

**System Genres:**

| Genre | Purpose | Primary Receiver | Skeleton Sections |
|-------|---------|-----------------|-------------------|
| `note` | Quick capture, route later | Self | Context, Content, Route |
| `signal_log` | System event record | Operator, agent | Timestamp, Event, Source, Impact |
| `context_update` | Persistent fact change | System | Node, Field, Old Value, New Value, Reason |
| `weekly_dump` | Raw brain dump | Self | Per-node stream of consciousness |
| `week_plan` | Structured weekly execution | Self | Non-Negotiables, Time Blocks, Dependencies |
| `weekly_review` | Friday double-loop | Self | Did it happen?, Was it right?, Fidelity check |
| `monthly_review` | Monthly triple-loop | Self | Right questions?, Assumptions to challenge? |

**Total: 37 genres.** Extend when a new form appears that can't be mapped.

### Type (T) — What speech act does it perform?

| Value | Description | Examples |
|-------|-------------|---------|
| `direct` | Compels action — imperative | Briefs, specs, runbooks, assignments |
| `inform` | Provides information — declarative | Reports, status updates, articles |
| `commit` | Makes a commitment — promissory | Plans, contracts, deadlines, promises |
| `decide` | Records a decision — declarative + binding | ADRs, pricing decisions, approvals |
| `express` | Expresses perspective — evaluative | Reviews, opinions, feedback, venting |
| `request` | Asks for something — interrogative | Questions, proposals, permission requests |

### Format (F) — What container holds it?

| Value | Description | File Extension |
|-------|-------------|---------------|
| `markdown` | Structured text | .md |
| `yaml` | Configuration/data | .yaml, .yml |
| `json` | Machine-readable data | .json |
| `code` | Source code | .ex, .ts, .go, .py, etc. |
| `html` | Web content | .html |
| `pdf` | Portable document | .pdf |
| `audio` | Sound recording | .mp3, .wav, .m4a |
| `video` | Video recording | .mp4, .mov, .webm |
| `image` | Visual artifact | .png, .jpg, .svg |
| `csv` | Tabular data | .csv |
| `sql` | Database query/schema | .sql |
| `plaintext` | Unformatted text | .txt |
| `slide` | Presentation | .pptx, .key |
| `spreadsheet` | Tabular with formulas | .xlsx, .csv |

### Structure (W) — What internal skeleton?

Maps 1:1 to genre. Every genre has a defined skeleton (see Genre table above). The structure dimension records WHICH skeleton was applied.

Format: `<genre>_skeleton` (e.g., `brief_skeleton`, `spec_skeleton`, `plan_skeleton`)

---

## 3. NODE TYPES — The Organizational Topology

Nodes are the organizational units that produce, consume, and route signals.

### Node Schema

```yaml
node:
  id: <node_id>                     # e.g., "04-ai-masters"
  name: <string>                    # Human-readable name
  type: <node_type_enum>            # See below
  status: <active | paused | archived | building | critical>
  owner: <endpoint_id>              # Primary responsible person
  endpoints: [<endpoint_id>, ...]   # People connected to this node

  # Files
  context_file: <path>/context.md   # Persistent facts
  signal_file: <path>/signal.md     # Weekly status
```

### Node Types

| Type | Description | Examples |
|------|-------------|---------|
| `person` | Individual's context and goals | 01-roberto |
| `entity` | Legal entity or organization | 03-lunivate |
| `domain:product` | Product/platform domain | 02-miosa |
| `domain:media` | Content/media domain | 08-content-creators |
| `operation:program` | Ongoing operational program | 04-ai-masters, 12-os-accelerator |
| `operation:network` | Network/partnership operation | 06-agency-accelerants |
| `operation:research` | Research/development track | 05-os-architect |
| `unit:community` | Community unit | 07-accelerants-community |
| `registry` | People/resource registry | 10-team |
| `cross-cutting` | Spans all nodes | 11-money-revenue |
| `inbox` | Unrouted signals | 09-new-stuff |

---

## 4. ENDPOINT TYPES — The People

Endpoints are people (or agents) that receive and produce signals.

### Endpoint Schema

```yaml
endpoint:
  id: <endpoint_id>                 # e.g., "robert-potter"
  name: <string>                    # Full name
  role: <string>                    # Current role
  node_memberships: [<node_id>, ...] # Nodes they belong to

  # Genre Competence — what forms they can decode
  receives: [<genre>, ...]          # Genres they understand
  does_not_receive: [<genre>, ...]  # Genres that confuse/overload them
  preferred_channel: <channel_enum> # How to reach them

  # Bandwidth
  bandwidth: <low | medium | high>  # Current capacity
  max_output_depth: <L0 | L1 | L2 | L3>  # How much detail they want
```

### Channel Types

| Value | Description |
|-------|-------------|
| `terminal` | Claude Code / CLI |
| `slack` | Slack DM or channel |
| `email` | Email |
| `voice` | Phone/Zoom call |
| `github` | GitHub issues/PRs |
| `sms` | Text message |

---

## 5. STORAGE PATTERNS — Where Everything Lives

### Tier 0: Filesystem (Markdown + AI)

```
OptimalOS/
├── CLAUDE.md                      # System prompt — the engine
├── topology.yaml                  # Node + endpoint definitions
├── week-plan.md                   # Weekly execution plan
├── weekly-dump.md                 # Monday brain dump
├── weekly-review.md               # Friday review
├── monthly-review.md              # Monthly review
├── alignment.md                   # Drift scoring
├── feedback-loops.md              # Reference
│
├── 01-roberto/
│   ├── context.md                 # Persistent facts (person type)
│   └── signal.md                  # Weekly status
├── 02-miosa/
│   ├── context.md                 # Persistent facts (domain type)
│   └── signal.md                  # Weekly status
├── ...                            # (12 numbered node folders)
│
└── docs/
    ├── architecture/              # 7-layer specs (L1-L7)
    ├── taxonomy/                  # Hierarchy, glossary, genres
    ├── operations/                # Intake, search, routing
    ├── competitors/               # Competitive landscape
    ├── data-model.md              # THIS FILE
    └── infinite-context-framework.md  # Universal framework
```

### Naming Conventions

**Files:**
| Pattern | Usage | Example |
|---------|-------|---------|
| `context.md` | Persistent node facts | `04-ai-masters/context.md` |
| `signal.md` | Weekly node status | `04-ai-masters/signal.md` |
| `week-plan.md` | Weekly execution | Root level |
| `weekly-dump.md` | Monday brain dump | Root level |
| `weekly-review.md` | Friday review | Root level |
| `<NN>-<name>/` | Numbered node folder | `04-ai-masters/` |
| `<descriptive-name>.md` | Any other artifact | `ad-scripts.md` |

**IDs:**
| Entity | Pattern | Example |
|--------|---------|---------|
| Signal | `sig_<ulid>` | `sig_01HZQX8K3M2VNPW9RG` |
| Node | `<NN>-<kebab-case>` | `04-ai-masters` |
| Endpoint | `<first>-<last>` | `robert-potter` |
| Decision | `dec_<date>_<short>` | `dec_2026-03-17_pricing-99` |
| Version | `v<N>` | `v3` |

**Signal file frontmatter (when stored as files):**
```yaml
---
signal:
  mode: linguistic
  genre: brief
  type: direct
  format: markdown
  structure: brief_skeleton
  sn_ratio: 0.95
  source_node: 04-ai-masters
  target_endpoints: [robert-potter]
---
```

### Tier 1: SQLite + FTS5

```sql
-- Signals table (append-only)
CREATE TABLE signals (
  id TEXT PRIMARY KEY,           -- sig_<ulid>
  version INTEGER NOT NULL,
  supersedes TEXT,               -- previous version ID
  created_at TEXT NOT NULL,      -- ISO 8601

  -- Classification
  mode TEXT NOT NULL,
  genre TEXT NOT NULL,
  type TEXT NOT NULL,
  format TEXT NOT NULL,
  structure TEXT NOT NULL,
  sn_ratio REAL DEFAULT 1.0,

  -- Routing
  source_node TEXT NOT NULL,
  target_nodes TEXT NOT NULL,    -- JSON array
  target_endpoints TEXT,         -- JSON array

  -- Content
  body TEXT NOT NULL,
  summary_l0 TEXT,
  summary_l1 TEXT,

  -- Metadata
  tier TEXT DEFAULT 'L2',
  sensitivity TEXT DEFAULT 'internal',
  ttl TEXT,
  tags TEXT,                     -- JSON array

  -- Extraction
  decisions TEXT,                -- JSON array
  actions TEXT,                  -- JSON array
  people TEXT,                   -- JSON array
  financial TEXT                 -- JSON array
);

-- Full-text search
CREATE VIRTUAL TABLE signals_fts USING fts5(
  body, summary_l0, summary_l1,
  content='signals', content_rowid='rowid'
);

-- Nodes table
CREATE TABLE nodes (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT DEFAULT 'active',
  owner TEXT,
  context TEXT,                  -- Full context.md content
  signal TEXT                    -- Full signal.md content
);

-- Endpoints table
CREATE TABLE endpoints (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  role TEXT,
  receives TEXT,                 -- JSON array of genres
  does_not_receive TEXT,         -- JSON array of genres
  preferred_channel TEXT,
  bandwidth TEXT DEFAULT 'medium',
  max_output_depth TEXT DEFAULT 'L2'
);

-- Relations (graph edges)
CREATE TABLE relations (
  id TEXT PRIMARY KEY,
  source_id TEXT NOT NULL,       -- signal, node, or endpoint ID
  target_id TEXT NOT NULL,
  relation_type TEXT NOT NULL,   -- parent, reference, cross_ref, member, routes_to
  created_at TEXT NOT NULL,
  valid_from TEXT,               -- Bi-temporal: when this became true
  valid_to TEXT                  -- Bi-temporal: when this stopped being true (null = current)
);

-- Decision traces
CREATE TABLE decisions (
  id TEXT PRIMARY KEY,           -- dec_<date>_<short>
  signal_id TEXT NOT NULL,       -- Which signal contained this decision
  node_id TEXT NOT NULL,         -- Which node it affects
  decision TEXT NOT NULL,
  decided_by TEXT NOT NULL,      -- endpoint_id
  rationale TEXT,
  supersedes TEXT,               -- Previous decision ID
  created_at TEXT NOT NULL
);

-- Temporal versions (append-only log)
CREATE TABLE versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  signal_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  field TEXT NOT NULL,
  old_value TEXT,
  new_value TEXT NOT NULL,
  changed_by TEXT,
  changed_at TEXT NOT NULL,
  reason TEXT
);
```

### Tier 2: SignalGraph (SQLite + SPARQL + Vectors)

Adds to Tier 1:
- **SPARQL graph layer** (via miosa_knowledge) for relationship queries
- **Vector embeddings** (via pgvector or Qdrant) for semantic search
- **Reciprocal Rank Fusion** across all 4 search modes
- **DIKW hierarchy** tagging (data→information→knowledge→wisdom)

### Tier 3: MIOSA Platform

Full production stack:
- **miosa_knowledge** — SPARQL + OWL reasoning
- **miosa_memory** — Episodic + SICA learning + Cortex synthesis
- **miosa_context** — SignalGraph + temporal versioning + tiered loading
- **miosa_signal** — Auto-classifier + S/N measurement + 11 failure modes
- **Compute engine** — Firecracker VMs for isolated agent execution

---

## 6. SEARCH MODES — How Context Is Retrieved

| Mode | Engine | Good At | Bad At | Tier Available |
|------|--------|---------|--------|---------------|
| **BM25 (lexical)** | FTS5 / Postgres FTS | Exact matches, names, technical terms | Paraphrases, synonyms | Tier 1+ |
| **Graph (relational)** | SPARQL / SQL JOINs | Connected entities, traversal | Free-text queries | Tier 2+ |
| **Vector (semantic)** | Qdrant / pgvector | Meaning similarity, concepts | Exact matches, proper nouns | Tier 2+ |
| **Temporal (recency)** | Decay function on timestamps | Recent events, current state | Timeless facts | Tier 1+ |

**Fusion:** Reciprocal Rank Fusion (RRF) merges results from all available modes:
```
RRF_score(d) = Σ 1/(k + rank_i(d))    where k = 60
```

---

## 7. TIERED DISCLOSURE — How Context Is Delivered

| Tier | Token Budget | What's Included | When to Use |
|------|-------------|-----------------|-------------|
| **L0** | ~2K tokens | One-paragraph summary per node. Headline. | Orientation. "What nodes exist?" |
| **L1** | ~10K tokens | Key facts + current priorities + blockers + people. | Working context. "What's happening this week?" |
| **L2** | ~50K tokens | Full context.md + signal.md + recent decisions + key docs. | Deep context. "I need to make a decision about X." |
| **L3** | Unlimited | Everything. All versions. All decision traces. All documents. | Full audit. "What happened with X over time?" |

**Assembly algorithm:**
1. Identify query intent and relevant nodes
2. Load L0 for ALL relevant nodes (cheap)
3. Score nodes by relevance to query
4. Expand top-N nodes to L1
5. If still insufficient → expand top node to L2
6. Never load L3 unless explicitly requested
7. Total token budget = receiver's max_output_depth × available context window

---

## 8. LIFECYCLE — How Data Moves Through States

### Signal Lifecycle
```
CREATED → CLASSIFIED → ROUTED → STORED → [RETRIEVED]* → [SUPERSEDED | ARCHIVED | EXPIRED]
```

### Node Lifecycle
```
DRAFT → ACTIVE → [PAUSED] → [ACTIVE] → COMPLETED → ARCHIVED
                   ↓
                CRITICAL (algedonic bypass)
```

### Decision Lifecycle
```
PROPOSED → DECIDED → IMPLEMENTED → [VERIFIED | REVERSED]
```

### Action Lifecycle
```
ASSIGNED → IN_PROGRESS → [BLOCKED] → COMPLETED → VERIFIED
                           ↓
                        ESCALATED
```

---

## 9. CROSS-REFERENCING RULES

| Trigger | Action |
|---------|--------|
| Financial data mentioned | ALWAYS also store in `nodes/11-money-revenue` |
| Decision made | ALWAYS log in relevant `context.md` under "Key Decisions" with date + rationale |
| Person mentioned | ALWAYS update `nodes/10-team/context.md` with latest known info |
| Multiple nodes touched | Update ALL affected `signal.md` files |
| Crisis signal | Route to `nodes/01-roberto` with URGENT flag + algedonic bypass |

---

## 10. FAILURE MODES — When Classification Breaks

| # | Mode | Constraint Violated | Symptom | Fix |
|---|------|-------------------|---------|-----|
| 1 | Routing Failure | Shannon | Signal went to wrong node | Re-route. Update routing table. |
| 2 | Bandwidth Overload | Shannon | Too much context delivered | Reduce tier. Summarize. Batch. |
| 3 | Fidelity Failure | Shannon | Meaning lost in transmission | Re-encode with clearer structure. |
| 4 | Genre Mismatch | Ashby | Wrong genre for receiver | Re-encode in correct genre. |
| 5 | Variety Failure | Ashby | No genre exists for this signal | Define new genre. Extend catalogue. |
| 6 | Structure Failure | Ashby | No skeleton applied | Apply genre-appropriate skeleton. |
| 7 | Bridge Failure | Beer | No shared context between nodes | Add preamble/context bridge. |
| 8 | Herniation Failure | Beer | Incoherent across system layers | Re-encode with proper layer traversal. |
| 9 | Decay Failure | Beer | Outdated signal still active | Audit. Version. Archive or sunset. |
| 10 | Feedback Failure | Wiener | No confirmation loop | Close the loop — verify, check, confirm. |
| 11 | Adversarial Noise | Cross-cutting | Deliberate degradation | Make noise visible. Escalate. |

---

## 11. NAMING CHEAT SHEET

| Thing | Convention | Example |
|-------|-----------|---------|
| Folder | `<NN>-<kebab-case>` | `04-ai-masters` |
| Persistent facts file | `context.md` | Always this name |
| Weekly status file | `signal.md` | Always this name |
| Signal ID | `sig_<ulid>` | `sig_01HZQX8K3M2VNPW9RG` |
| Decision ID | `dec_<date>_<slug>` | `dec_2026-03-17_pricing` |
| Node ID | Same as folder name | `04-ai-masters` |
| Endpoint ID | `<first>-<last>` | `robert-potter` |
| Version number | `v<N>` sequential | `v1`, `v2`, `v3` |
| Genre enum | `snake_case` | `video_script`, `ad_script` |
| Mode enum | `snake_case` | `linguistic`, `mixed` |
| Type enum | `snake_case` | `direct`, `inform`, `decide` |
| Format enum | `snake_case` | `markdown`, `json`, `yaml` |
| Structure enum | `<genre>_skeleton` | `brief_skeleton` |
| Tier enum | `L0`, `L1`, `L2`, `L3` | Uppercase L + number |

---

## 12. VALIDATION RULES

A signal is VALID when:
1. All 5 dimensions resolved (mode, genre, type, format, structure)
2. At least one target_node specified
3. Body is non-empty
4. summary_l0 exists (auto-generate if missing)
5. Genre matches a defined genre in the catalogue
6. If financial data → cross_refs includes `nodes/11-money-revenue`
7. If decision → decisions array is non-empty
8. If person mentioned → people array is non-empty

A node is VALID when:
1. Has context.md file
2. Has signal.md file
3. Has a defined type from the node type enum
4. Has at least one endpoint (owner)

An endpoint is VALID when:
1. Has name and role
2. Has at least one genre in receives
3. Has preferred_channel defined
