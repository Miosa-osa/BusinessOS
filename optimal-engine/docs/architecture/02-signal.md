---
signal:
  mode: linguistic
  genre: spec
  type: inform
  format: markdown
  structure: layer_spec
  sn_ratio: 1.0
  audience: [human, agent]
  intent: "Layer 2 Signal specification — S=(M,G,T,F,W) classification system for OptimalOS"
---

# Layer 2: Signal — S=(M,G,T,F,W)

> Every piece of data in the system MUST be classified across all 5 dimensions.
> No unclassified data exists.

---

## Purpose

Raw information is noise. A Signal is information that has been classified, structured, and made actionable. Layer 2 imposes Signal classification on every artifact that enters, traverses, or exits the system — messages, documents, code, decisions, data records, and agent outputs alike.

Classification resolves the 5-tuple **S = (M, G, T, F, W)**:

| Dimension | Name | Question answered |
|-----------|------|-------------------|
| M | Mode | How is it perceived? |
| G | Genre | What conventionalized form? |
| T | Type | What speech act does it perform? |
| F | Format | What container holds it? |
| W | Structure | What internal skeleton does it follow? |

A Signal with all 5 dimensions resolved has **S/N ratio = 1.0**. Every unresolved dimension reduces fidelity and increases decoding effort at the receiver.

---

## Governing Constraint

**Ashby's Law of Requisite Variety** — the variety in a controller must be at least as great as the variety in the system being controlled.

Applied here: the system's repertoire of Modes, Genres, Types, Formats, and Structures must be large enough to classify every signal it will ever encounter. Gaps in the catalogue produce `variety_failure` — the system cannot represent the signal correctly, and meaning is lost.

The catalogue below is the current variety budget. It must be extended whenever a new signal form appears that cannot be mapped to an existing entry.

---

## The Five Dimensions

### Mode (M) — How is it perceived?

Mode determines the perceptual channel through which a receiver decodes the signal. Mismatched mode and content (e.g., using `:linguistic` to describe a spatial relationship that belongs in a `:visual`) violates mode-message alignment.

| Value | Description | Examples |
|-------|-------------|---------|
| `:linguistic` | Text or speech — sequential, propositional | Docs, messages, transcripts, emails |
| `:visual` | Images, diagrams, video — relational, spatial | Architecture diagrams, screenshots, videos |
| `:code` | Source code, configs, executable structures | `.ex`, `.ts`, `.yaml`, scripts |
| `:data` | Structured records — machine-readable | JSON, CSV, database rows, tables |
| `:mixed` | Multiple modes combined in one artifact | Slide deck (visual + linguistic), notebook (code + linguistic + data) |

**Selection rule:** choose the mode that matches the dominant perceptual channel of the artifact. For `:mixed`, list all contributing modes in order of dominance.

---

### Genre (G) — What conventionalized form?

Genre determines the social contract between sender and receiver — what the receiver expects to find, in what order, at what level of detail. Genre mismatch (`genre_mismatch` failure) is one of the most common sources of decoding failure: a developer receiving a `:brief` when they need a `:spec` will miss required implementation detail.

**Full genre catalogue:**

| Genre | Purpose | Primary receiver |
|-------|---------|-----------------|
| `spec` | Defines requirements, behavior, or contracts precisely | Engineer, agent |
| `report` | Presents findings, metrics, or analysis | Decision-maker, operator |
| `brief` | Concise directive summary for action | Salesperson, executive |
| `decision` / `adr` | Records a choice with rationale and alternatives | Team, future self |
| `status` | Current state of an operation or project | Operator, stakeholder |
| `plan` | Ordered steps toward a goal with owners and timelines | Team, operator |
| `review` | Evaluative assessment of an artifact or system | Author, team |
| `guide` | How-to instructions for a process or tool | New user, operator |
| `script` | Step-by-step runbook or automation script | Operator, CI system |
| `template` | Reusable structure with fill-in slots | Author, agent |
| `meeting_notes` | Record of discussion, decisions, and actions | Attendees, team |
| `chat` | Conversational exchange — low formality | Human, agent |
| `email` | Asynchronous directed message | Named recipient |
| `announcement` | Broadcast notification — no reply expected | Wide audience |
| `article` | Long-form explanatory or argumentative prose | General reader |
| `video_script` | Narration and visual cues for video production | Presenter, editor |
| `social_post` | Short-form public content for social channels | Public audience |
| `course_module` | Instructional unit with learning objectives | Learner |
| `presentation` | Structured slide or talk content | Live audience |
| `proposal` | Pitch or recommendation requesting approval | Approver, sponsor |
| `invoice` | Financial billing record | Accountant, client |
| `contract` | Legally binding agreement | Parties, legal |

**Selection rule:** if no genre fits, document the gap as a `variety_failure` and define a new genre entry before proceeding.

---

### Type (T) — What speech act?

Type classifies the illocutionary force of the signal — what the sender intends it to *do* in the world. Based on Searle's speech act taxonomy, reduced to five operative categories.

| Value | Force | Commitment | Examples |
|-------|-------|------------|---------|
| `:direct` | Compels action | Receiver must act | "Deploy this build", "Fix this bug", task assignments |
| `:inform` | States facts | Receiver gains knowledge | Status updates, reports, meeting notes |
| `:commit` | Makes a promise | Sender will act | "I will deliver X by Y", sprint commitments, contracts |
| `:decide` | Declares a choice | Choice is now canonical | ADRs, architecture decisions, policy changes |
| `:express` | Conveys attitude or evaluation | No binding obligation | "I'm concerned about...", praise, risk flags |

**Selection rule:** a single signal can carry multiple speech acts (e.g., a status report that also flags risk is `:inform` + `:express`). Record the primary type first.

---

### Format (F) — What container?

Format is the file or transmission container — the encoding that a tool, renderer, or channel will process. Format affects what tooling can consume the signal and what rendering is possible.

| Value | Use |
|-------|-----|
| `:markdown` | Human-readable docs with lightweight markup |
| `:code` | Source files (any language) |
| `:json` | Machine-readable structured data |
| `:yaml` | Human-readable config and structured data |
| `:pdf` | Print-ready, non-editable documents |
| `:doc` | Word-processor documents (.docx, .gdoc) |
| `:slides` | Presentation files (.pptx, Google Slides) |
| `:video` | Video files (.mp4, .mov) |
| `:audio` | Audio files (.mp3, .wav) |
| `:image` | Raster or vector images (.png, .svg, .jpg) |
| `:csv` | Tabular plain-text data |
| `:html` | Web-rendered markup |
| `:text` | Plain unformatted text |

---

### Structure (W) — Internal skeleton

Structure maps from Genre to the specific template key that governs internal organization: sections, order, required fields, and granularity. Structure is the genre skeleton made concrete.

| Genre | Structure key | Required sections |
|-------|--------------|-------------------|
| `spec` | `spec_template` | Overview, Requirements, Constraints, Interface, Acceptance Criteria |
| `report` | `report_template` | Summary, Findings, Metrics, Recommendations |
| `brief` | `sales_brief` | Objective, Audience, Key Messages, Call to Action |
| `decision` / `adr` | `adr_template` | Status, Context, Decision, Consequences, Alternatives, References |
| `status` | `weekly_signal` | Period, Wins, Blockers, Next Actions, Metrics |
| `plan` | `plan_template` | Goal, Milestones, Tasks, Owners, Timeline, Risks |
| `review` | `review_template` | Summary, Findings (Critical/Major/Minor), Recommendations, Positive Notes |
| `guide` | `guide_template` | Prerequisites, Steps, Examples, Troubleshooting |
| `script` | `runbook_template` | Purpose, Prerequisites, Steps, Rollback, Verification |
| `template` | `template_meta` | Purpose, Variables, Usage, Example |
| `meeting_notes` | `meeting_template` | Date, Attendees, Agenda, Decisions, Actions, Next Meeting |
| `chat` | `thread` | (conversational — no enforced skeleton) |
| `email` | `email_template` | Subject, Greeting, Body, CTA, Signature |
| `announcement` | `announcement_template` | Headline, Context, What Changed, Who Is Affected, Next Steps |
| `article` | `article_template` | Headline, Lede, Body, Conclusion, References |
| `video_script` | `video_script_template` | Intro Hook, Sections, Transitions, Outro, B-roll Notes |
| `social_post` | `social_template` | Hook, Body, CTA, Hashtags |
| `course_module` | `course_module_template` | Learning Objectives, Content, Exercises, Assessment, Summary |
| `presentation` | `slide_template` | Title Slide, Agenda, Content Sections, Summary, Q&A |
| `proposal` | `proposal_template` | Executive Summary, Problem, Solution, Timeline, Budget, Ask |
| `invoice` | `invoice_template` | Parties, Line Items, Totals, Payment Terms, Due Date |
| `contract` | `contract_template` | Parties, Recitals, Terms, Obligations, Signatures |

**Selection rule:** if the genre has no registered structure key, define one before storing the signal.

---

## S/N Ratio Measurement

The Signal-to-Noise ratio for a classified signal is the fraction of the 5 dimensions that have been resolved to a non-null, valid value.

```
sn_ratio = resolved_dimensions / 5
```

| S/N Ratio | State | Action required |
|-----------|-------|-----------------|
| 1.0 | Fully classified | None |
| 0.8 | One dimension missing | Infer or flag for review |
| 0.6 | Two dimensions missing | Needs attention before routing |
| < 0.6 | Three+ dimensions missing | Blocked — classify before use |
| < 0.3 | Fidelity failure | Re-ingest or discard |

Auto-classification (see below) attempts to resolve all 5 dimensions. Signals that cannot reach ≥ 0.6 after auto-classification are quarantined pending manual review.

---

## 11 Failure Modes

Every signal is checked against all 11 failure modes on intake. Detected failures are logged as structured events and surfaced to Layer 6 (Feedback) for learning.

### Shannon Violations (channel capacity)

| Code | Name | Cause | Resolution |
|------|------|-------|------------|
| `routing_failure` | Wrong recipient | Signal delivered to a node that cannot decode it | Re-route to correct node |
| `bandwidth_overload` | Too much for the channel | Output exceeds receiver's decoding capacity | Reduce, prioritize, or batch |
| `fidelity_failure` | Meaning lost in transmission | Encoding/decoding mismatch, ambiguous structure | Re-encode with explicit structure |

### Ashby Violations (requisite variety)

| Code | Name | Cause | Resolution |
|------|------|-------|------------|
| `genre_mismatch` | Wrong form for receiver | Receiver lacks competence to decode this genre | Re-encode in genre the receiver can decode |
| `variety_failure` | No genre covers this signal | Catalogue gap — signal type not representable | Define new genre entry; extend catalogue |
| `structure_failure` | No internal skeleton | Genre has no registered structure key | Define and register structure template |

### Beer Violations (viable structure)

| Code | Name | Cause | Resolution |
|------|------|-------|------------|
| `bridge_failure` | No shared context | Sender and receiver lack common ground | Add preamble, establish conventions |
| `herniation_failure` | Incoherence across layers | Signal inconsistent between layers of abstraction | Re-encode with coherent layer traversal |
| `decay_failure` | Signal is outdated | Temporal decay — content no longer reflects reality | Audit, version, or sunset the signal |

### Wiener Violations (closed loops)

| Code | Name | Cause | Resolution |
|------|------|-------|------------|
| `feedback_failure` | No confirmation loop | Action was taken but no verification that it landed | Close the loop — verify, check, confirm |

### Cross-Cutting

| Code | Name | Cause | Resolution |
|------|------|-------|------------|
| `adversarial_noise` | Deliberate signal degradation | Intentional injection of misleading or corrupting content | Make noise visible; escalate to Layer 7 |

---

## Auto-Classification

The system infers Signal dimensions from content heuristics rather than requiring manual tagging on every intake. The auto-classifier runs on all incoming signals before storage.

### Classification Pipeline

```
RAW INPUT
    │
    ▼
1. FORMAT DETECTION
   Inspect file extension, MIME type, byte patterns
   → resolves F (Format)

    │
    ▼
2. MODE INFERENCE
   Format → Mode heuristic:
   .md/.txt/.docx → :linguistic
   .py/.ts/.ex/.yaml → :code
   .json/.csv → :data
   .png/.svg/.mp4 → :visual
   Multiple detected → :mixed
   → resolves M (Mode)

    │
    ▼
3. GENRE CLASSIFICATION
   Content heuristics + metadata:
   - Frontmatter signal.genre field (authoritative if present)
   - Filename patterns (YYYY-MM-DD → status/meeting_notes)
   - Section header patterns (## Status, ## Decision → adr)
   - Type→genre mapping table (see below)
   → resolves G (Genre)

    │
    ▼
4. TYPE INFERENCE
   Linguistic analysis of primary verb/intent:
   - Imperative verbs → :direct
   - Declarative present tense → :inform
   - "I will / we will / committed to" → :commit
   - "We chose / decision:" → :decide
   - "Concerned / pleased / worried" → :express
   → resolves T (Type)

    │
    ▼
5. STRUCTURE RESOLUTION
   Genre → structure key lookup (table above)
   → resolves W (Structure)

    │
    ▼
6. S/N RATIO COMPUTATION
   Count resolved dimensions / 5
   If < 0.6 → quarantine

    │
    ▼
7. FAILURE MODE DETECTION
   Run all 11 checks
   Log detected failures
```

### Handling by Data Shape

| Data shape | Classification approach |
|------------|------------------------|
| Structured (has schema) | Classify by schema/format; Genre = `spec` or `report` depending on purpose |
| Semi-structured (JSON, YAML with known keys) | Content heuristics + metadata fields; resolve all 5 from field names and values |
| Unstructured (free text, voice, image) | Transcribe/parse first; then run full pipeline above |
| Code files | Format = `:code`, Mode = `:code`; Genre inferred from file path and content purpose |

### Existing Implementation

| Module | Function | Purpose |
|--------|----------|---------|
| `MiosaSignal` | struct | CloudEvents v1.0.2 envelope + Signal Theory extensions (mode, genre, type, format, structure, sn_ratio) |
| `MiosaSignal.Classifier` | `classify/1` | Classify a single signal against known type→genre mapping |
| `MiosaSignal.Classifier` | `auto_classify/1` | Full auto-classification pipeline — runs all 5 dimension inference steps |
| `MiosaSignal.FailureModes` | `detect/1` | Runs all 11 failure mode checks against a classified signal |
| `MiosaSignal` | `measure_sn_ratio/1` | Computes S/N ratio from resolved dimension count |

---

## YAML Frontmatter Format

Every markdown file in OptimalOS carries Signal metadata in YAML frontmatter. This makes every document self-describing and machine-readable without requiring external indexing.

```yaml
---
signal:
  mode: linguistic
  genre: status
  type: inform
  format: markdown
  structure: weekly_signal
  sn_ratio: 1.0
  audience: [human, agent]
  intent: "Weekly status update for AI Masters operation"
---
```

### Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `mode` | Yes | One of the Mode values above (string, no colon prefix in YAML) |
| `genre` | Yes | One of the Genre catalogue entries |
| `type` | Yes | One of the Type values above |
| `format` | Yes | One of the Format values above |
| `structure` | Yes | Structure key from the Genre→Structure table |
| `sn_ratio` | Yes | Float 0.0–1.0; compute as resolved_dimensions / 5 |
| `audience` | Yes | Array — receivers this signal is optimized for |
| `intent` | Yes | One sentence: what this signal is meant to accomplish |

**Rule:** no markdown file is committed without complete frontmatter. The auto-classifier generates a frontmatter block for any file missing one; authors review and confirm before merge.

---

## Layer Boundaries

**L2 receives from L1 (Network):** a raw signal with source node and routing metadata. L2's job is classification — it does not modify content.

**L2 hands off to L3 (Composition):** a fully classified Signal with S/N ratio and failure mode flags. L3 uses the Genre and Structure dimensions to validate and apply the internal skeleton.

**L2 classifies outputs too:** every signal leaving the system — agent responses, generated documents, API payloads — is classified before routing. The output classification feeds L6 (Feedback) to measure whether the signal achieved its intent.

---

## References

- [Layer Overview: 7-Layer Architecture](00-overview.md)
- [Layer 1: Network](01-network.md)
- [Layer 3: Composition](03-composition.md)
- [Taxonomy: Genre Catalogue](../taxonomy/genres.md)
- Ashby, W.R. — *An Introduction to Cybernetics* (1956), Chapter 11: Requisite Variety
- Searle, J.R. — *Speech Acts* (1969)
- Luna, R.H. — *Signal Theory: The Architecture of Optimal Intent Encoding* (MIOSA Research, Feb 2026)
