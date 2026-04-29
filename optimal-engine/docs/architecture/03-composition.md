# Layer 3: Composition

> **Governing Constraint:** Ashby — Requisite Variety (genre repertoire)
> **Purpose:** The micro-structure within signals. Genre-specific skeletons. Granularity levels.
> **Position in onion:** Wraps L2 Signal. Wrapped by L4 Interface.
> **Existing code:** `miosa_signal.Classifier` (genre→structure mapping), `miosa_context` (SQLite skeleton store)

---

## What Composition Governs

L2 classifies a signal's five dimensions: `S = (M, G, T, F, W)`.

Composition governs **W** — the internal skeleton — for every genre in G.

A signal without a composition skeleton is unstructured information. Unstructured information is noise.
Composition is the layer that converts classified signals into decodable artifacts.

Three responsibilities:
1. **Genre templates** — required sections and field metadata for every genre
2. **Granularity levels** — L0 through L3, enabling progressive disclosure at retrieval
3. **Fact extraction** — decomposing signals into atomic SPO triples for the graph

---

## Genre Templates

Each genre defines a **W skeleton**: an ordered list of required and optional sections.

**Notation:**
- `[R]` = Required. Signal is invalid without it.
- `[O]` = Optional. Include when applicable.
- `[A]` = Auto-generated. System fills from context or metadata.

---

### Business Genres

#### `brief`
Short directional document. Receiver = non-technical decision-maker. Goal = compel a specific action.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Objective | [R] | One sentence. What outcome are we driving? |
| 2 | Audience | [R] | Who is this for? Their role and decoding bandwidth. |
| 3 | Key Messages | [R] | 3–5 bullet points. Each is an atomic claim. |
| 4 | Call to Action | [R] | Single, unambiguous ask. Verb + deadline. |
| 5 | Timeline | [O] | Milestones if timing is decision-relevant. |
| 6 | Supporting Materials | [O] | Links, attachments, companion assets. |

**Failure mode:** Brief that lacks a CTA is an `inform` signal misclassified as `direct`.

---

#### `proposal`
Formal offer requiring approval. Receiver = executive or procurement. Goal = signed commitment.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Executive Summary | [R] | ≤5 sentences. Problem + solution + ask. |
| 2 | Problem | [R] | Quantified pain. Use data where available. |
| 3 | Solution | [R] | What you are offering and how it works. |
| 4 | Pricing | [R] | Itemized. Include options if applicable. |
| 5 | Timeline | [R] | Delivery milestones with dates. |
| 6 | Terms | [R] | Payment, IP, cancellation, warranties. |
| 7 | Team / Credentials | [O] | Who delivers. Relevant track record. |
| 8 | Appendices | [O] | Case studies, technical specs, references. |

---

#### `spec`
Technical contract between teams or systems. Receiver = engineer / implementer. Goal = unambiguous build target.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Requirements | [R] | Functional + non-functional. Numbered. |
| 2 | Acceptance Criteria | [R] | Testable pass/fail conditions per requirement. |
| 3 | Architecture | [R] | Component boundaries, data flow, key decisions. |
| 4 | API Contracts | [O] | Endpoint definitions, schemas, error codes. |
| 5 | Dependencies | [R] | External systems, libraries, services required. |
| 6 | Risks | [R] | Technical risks + mitigations. Probability × impact. |
| 7 | Open Questions | [O] | Unresolved decisions. Owner + deadline per item. |
| 8 | Out of Scope | [O] | Explicit exclusions to prevent scope creep. |

---

#### `decision` (ADR — Architecture Decision Record)
Immutable record of a significant choice. Receiver = current + future team. Goal = prevent re-litigation.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Status | [R] | `Proposed` / `Accepted` / `Deprecated` / `Superseded` |
| 2 | Date | [A] | ISO 8601. Auto-stamped on creation. |
| 3 | Context | [R] | What forces made this decision necessary? |
| 4 | Decision | [R] | What was chosen. One unambiguous statement. |
| 5 | Rationale | [R] | Why this option over alternatives. |
| 6 | Consequences | [R] | Positive / Negative / Neutral sub-sections. |
| 7 | Alternatives Considered | [R] | Each rejected option + reason for rejection. |
| 8 | Decision Trace | [A] | Who, when, which signals informed this. |
| 9 | References | [O] | Links to related ADRs, docs, evidence. |

**Decision Trace fields:** `decided_by`, `decided_at`, `informed_by[]` (signal IDs), `supersedes` (ADR ID if applicable).

---

#### `plan`
Forward-looking work breakdown. Receiver = team / executor. Goal = coordinated execution.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Objectives | [R] | What done looks like. Measurable outcomes. |
| 2 | Tasks | [R] | Itemized. Owner + estimate + priority per task. |
| 3 | Timeline | [R] | Gantt or milestone table. |
| 4 | Resources | [O] | People, tools, budget required. |
| 5 | Dependencies | [R] | What must be true before tasks can start. |
| 6 | Risks | [R] | What could prevent execution. Mitigations. |
| 7 | Success Metrics | [O] | How we know the plan succeeded. |

---

#### `report`
Retrospective analysis. Receiver = decision-maker. Goal = inform a future decision.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Summary | [R] | ≤3 sentences. What happened, what it means. |
| 2 | Findings | [R] | Ordered by significance. Data-backed. |
| 3 | Analysis | [R] | Interpretation. Causal not correlational where possible. |
| 4 | Recommendations | [R] | Specific, actionable, prioritized. |
| 5 | Next Steps | [R] | Owner + deadline per recommendation. |
| 6 | Methodology | [O] | How data was gathered. Relevance to confidence. |
| 7 | Appendices | [O] | Raw data, charts, source logs. |

---

### Communication Genres

#### `status` (Weekly Signal)
Recurring pulse. Receiver = team / stakeholders. Goal = shared situational awareness without meetings.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Week | [A] | ISO week number + date range. |
| 2 | Priorities | [R] | What this week is optimizing for. ≤3 items. |
| 3 | Progress | [R] | What moved. Linked to tasks/projects. |
| 4 | Blockers | [R] | What is stuck. Owner + ask per blocker. |
| 5 | Dependencies | [O] | What this team is waiting on from others. |
| 6 | Decisions Made | [O] | Significant choices. Link to ADRs if created. |
| 7 | Notes | [O] | Anything that doesn't fit above. |

---

#### `meeting_notes`
Synchronized decision record. Receiver = attendees + absent stakeholders. Goal = single source of truth for what was decided.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Date / Time | [A] | ISO 8601 + timezone. |
| 2 | Attendees | [R] | Name + role. Mark absent-but-notified. |
| 3 | Agenda | [R] | Items discussed. Time-boxed if available. |
| 4 | Decisions | [R] | Verbatim decisions. No paraphrase. |
| 5 | Action Items | [R] | Owner + task + deadline per item. |
| 6 | Follow-ups | [O] | Items deferred. Owner + target date. |
| 7 | Recording Link | [O] | URL if session was recorded. |

---

#### `chat`
Ephemeral message. Minimal structure by design.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Sender | [A] | Node ID of sender. |
| 2 | Timestamp | [A] | ISO 8601. |
| 3 | Content | [R] | The message body. |
| 4 | Thread ID | [A] | Parent message ID if reply. |
| 5 | Channel | [A] | Routing metadata (Slack channel, DM, etc.) |

**Note:** Chat signals are stored in L5 but rarely promoted beyond L3 granularity unless they contain a decision or blocker.

---

#### `email`
Asynchronous directed communication. Receiver = named individual(s). Goal = inform or request action.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Subject | [R] | ≤60 characters. Specific, not generic. |
| 2 | Context | [O] | 1–2 sentences of background if needed. |
| 3 | Ask / Inform | [R] | The core payload. One primary ask or datum. |
| 4 | Next Steps | [O] | What happens after they read this. |
| 5 | Deadline | [O] | If time-sensitive. Explicit date, not "ASAP". |

---

#### `announcement`
Broadcast to a defined population. Receiver = team or public. Goal = synchronized awareness of a change.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | What | [R] | The change. Factual, ≤2 sentences. |
| 2 | Why | [R] | Rationale. Enough to prevent confusion and rumor. |
| 3 | Impact | [R] | Who is affected and how. |
| 4 | Timeline | [R] | When it takes effect. |
| 5 | Contact | [R] | Who to ask questions to. |
| 6 | FAQ | [O] | Pre-empted questions for high-impact announcements. |

---

### Content Genres

#### `article`
Long-form published content. Receiver = target audience segment. Goal = educate, persuade, or position.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Hook | [R] | First paragraph. Creates tension or curiosity. |
| 2 | Thesis | [R] | The central claim. One sentence. |
| 3 | Body | [R] | Supporting arguments. 3–7 sections with sub-headers. |
| 4 | Conclusion | [R] | Restate thesis in light of evidence. |
| 5 | CTA | [O] | What to do next. Link or subscribe or share. |
| 6 | Meta | [A] | Author, date, tags, reading time. |

---

#### `video_script`
Structured visual-linguistic content. Receiver = viewer. Goal = retention and action within time budget.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Hook (0–30s) | [R] | Interrupt the scroll. Pain or curiosity or bold claim. |
| 2 | Problem | [R] | Expand the tension. Why does this matter to the viewer? |
| 3 | Solution | [R] | The core content. Show, don't tell where possible. |
| 4 | Proof | [O] | Social proof, data, demonstration. |
| 5 | CTA | [R] | Single ask. Subscribe / book / buy. One action only. |
| 6 | B-Roll Notes | [O] | Visual direction per section. |
| 7 | Runtime Estimate | [A] | Derived from word count at 140 wpm. |

---

#### `social_post`
Platform-constrained broadcast. Receiver = platform algorithm + human follower. Goal = engagement and reach.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Hook | [R] | Line 1. Must survive truncation at 120 chars. |
| 2 | Value | [R] | The payload. Platform-specific length constraints apply. |
| 3 | CTA | [O] | Comment / share / link. One action. |
| 4 | Platform Metadata | [A] | Channel (LinkedIn/X/etc.), char count, hashtags. |

**Platform constraints:**
- LinkedIn: ≤3000 chars, first 210 visible before "see more"
- X / Twitter: ≤280 chars per post (threads allowed)
- Instagram caption: ≤2200 chars, first 125 visible

---

#### `course_module`
Structured learning unit. Receiver = learner. Goal = measurable knowledge or skill acquisition.

| # | Section | Status | Notes |
|---|---------|--------|-------|
| 1 | Learning Objectives | [R] | Bloom's taxonomy verbs. 2–4 per module. |
| 2 | Prerequisites | [O] | What the learner must already know. |
| 3 | Content | [R] | Core instruction. Text, diagrams, code, video. |
| 4 | Exercises | [R] | Practice tasks. At least one per objective. |
| 5 | Assessment | [R] | Pass/fail check. Quiz, submission, or checklist. |
| 6 | Summary | [O] | 3–5 key takeaways. |
| 7 | Module Metadata | [A] | Duration, difficulty, version, author. |

---

## Granularity Levels (L0–L3)

Every signal, regardless of genre, has four granularity levels. These enable progressive disclosure at L4 Interface. They are pre-computed at intake and stored in L5 Data.

| Level | Name | Target Length | Contents | When Used |
|-------|------|---------------|----------|-----------|
| **L0** | Headline | ~10 words | Subject + verb + object. No qualifiers. | Always-loaded context, dashboards, entity indexes |
| **L1** | Summary | ~50 words | Key facts. The 5 Ws where available. Links to L2. | L1 warm context tier, search result snippets |
| **L2** | Detail | ~500 words | Full genre skeleton with populated sections. Actionable. | Task-relevant context, agent working memory |
| **L3** | Complete | Unlimited | Everything. Verbatim source + all metadata. | Audit, deep research, archival retrieval |

**Examples for the same signal (sales brief, AI Masters):**

```
L0: Alice directed Robert to activate AI Masters enterprise sales for Q1.

L1: Alice sent Bob a sales brief on 2026-03-17 targeting AI
    Masters enterprise accounts. Objective: close 2 deals by end of Q1. CTA:
    schedule discovery calls this week. See L2 for full brief.

L2: [Full brief — Objective / Audience / Key Messages / CTA / Timeline]

L3: [L2 + original draft, revision history, related signals, decision trace,
    SPO triples, embedding vector, metadata]
```

**Generation rules:**
- L0 is always generated at intake. System-generated via template: `{subject} {verb} {object}`.
- L1 is generated at intake by the classifier. Max 3 sentences.
- L2 maps directly to the populated genre skeleton.
- L3 is the stored artifact itself — no generation, only retrieval.

---

## Fact Extraction

Every signal is decomposed into atomic Subject-Predicate-Object triples at intake. These populate the L5 knowledge graph and enable SPARQL traversal at retrieval.

**Format:** `(Subject, Predicate, Object)` where each element is a node or literal.

**Example input:**
> "Alice decided to use OpenRouter for ClinicIQ on March 6"

**Extracted triples:**
```
(Alice,      decided,          use_OpenRouter)
(ClinicIQ,     uses,             OpenRouter)
(decision,     date,             2026-03-06)
(decision,     decided_by,       Alice)
(decision,     applies_to,       ClinicIQ)
```

**Extraction rules:**
1. One triple per atomic claim. Do not compress two claims into one triple.
2. Entities are normalized to canonical IDs (e.g., `Alice` → `node:roberto_luna`).
3. Dates are ISO 8601 literals.
4. Predicates use snake_case. Prefer existing predicates before creating new ones.
5. If a claim is uncertain, add `(triple_id, confidence, 0.0–1.0)` meta-triple.
6. Decisions produce a `decision` node with `decided_by`, `decided_at`, `applies_to`, and `informed_by` predicates.

**Triple storage:** Written to `miosa_knowledge` SPARQL store. Indexed for FTS5. Linked to originating signal ID in SignalGraph.

---

## Composition by Mode

Signals have a mode dimension (M). Each mode has its own atomic units, molecular structures, and compound structures. Composition rules apply at each level of the hierarchy.

| Mode | Atoms | Molecules | Structures | Composition Rule |
|------|-------|-----------|------------|-----------------|
| **Linguistic** | Words, punctuation | Sentences, claims | Paragraphs, sections, documents | Genre skeleton governs ordering and completeness |
| **Visual** | Elements (icon, color, shape) | Components, frames | Layouts, dashboards, diagrams | Spatial hierarchy: Z-order, grouping, alignment |
| **Code** | Tokens, literals | Functions, modules | Packages, systems | Dependency direction: inward (no infrastructure in domain) |
| **Data** | Values, cells | Records, rows | Tables, schemas | Schema governs type, nullability, and relationships |

**Multi-mode signals** (e.g., a spec with both prose and code) must satisfy the composition rules for each mode independently, then integrate via a container structure (e.g., a document with embedded code blocks).

---

## Failure Modes at the Composition Layer

| Failure | Description | Detection | Remediation |
|---------|-------------|-----------|-------------|
| **Missing required section** | Genre skeleton has [R] field with no content | Classifier validation at intake | Request missing section or escalate to sender |
| **Wrong skeleton applied** | Signal classified as `brief` but structured as `report` | Section count mismatch vs expected skeleton | Re-classify or re-structure |
| **Granularity collapse** | L2 contains only L1-level content | Token count below threshold for L2 | Expand or demote to L1 |
| **Triple extraction failure** | Claims present but no SPO triples generated | Zero triples for non-chat signal | Re-run extractor; flag for manual review |
| **Skeleton drift** | Populated sections out of order | Section position vs template | Re-order; warn but do not reject |

---

## Relationship to Adjacent Layers

**From L2 (Signal):**
L2 provides the classified `G` (genre) dimension. Composition uses `G` to select the skeleton.
L2 also provides `M` (mode), which selects the appropriate mode composition rules.

**To L4 (Interface):**
Composition produces the four granularity levels (L0–L3) that L4 uses for tiered loading.
L4 selects which granularity to serve based on token budget and receiver type.
L4 never re-structures content — it selects pre-computed granularity artifacts produced here.

**To L5 (Data):**
Extracted SPO triples are written to the knowledge graph.
L0/L1/L2/L3 artifacts are stored in SignalGraph `context_tiers` table with the originating signal ID.
