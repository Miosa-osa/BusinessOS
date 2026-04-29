# Layer 6: Feedback

> **Governing Constraint:** Wiener — closed loops
> **One-Line Purpose:** Did it happen? Was it right? Are we asking the right questions?
> **Existing Code:** `miosa_memory.Learning` (SICA), `miosa_memory.Episodic`, `miosa_memory.Cortex`, `miosa_signal.FailureModes`

---

## What This Layer Does

Layer 6 is the self-correction engine. It closes every loop that Layers 1–5 open.

Without L6, the system is a broadcast machine — signals go out, work gets done, but there is no mechanism to verify that intent survived transmission, or that the actions taken were the right ones to take. A system without closed feedback loops will optimize itself into irrelevance.

L6 operates at three distinct levels — operational, strategic, and meta-strategic — each running on a different cadence, asking a fundamentally different question, and producing a different class of output.

```
LAYER 6 SCOPE

  Triple-loop  →  Are we asking the right questions?  [Monthly]
       ↑               Challenges: the framework itself
       │               Produces: structural changes, metric rewrites
       │
  Double-loop  →  Are we doing the right things?      [Weekly]
       ↑               Challenges: current strategy
       │               Produces: node reprioritization, strategy pivots
       │
  Single-loop  →  Did the action happen?              [Per-signal / Friday]
                      Challenges: execution only
                      Produces: confirmation, carry-forwards, fidelity scores
```

These three loops are not parallel alternatives. They are nested. Single-loop data
feeds the double-loop analysis. Double-loop conclusions feed the triple-loop questions.
A system running only the bottom loop is executing efficiently in the wrong direction.

---

## The Three Feedback Levels

### Level 1 — Single-Loop (Operational Feedback)

**Question:** Did the action happen?

**What it challenges:** Execution only. The single-loop does not ask whether the action
was the right one — only whether it was performed and whether the outcome matched the
stated intent.

**Trigger:** Per-signal (after any delegated task's deadline) and every Friday as the
end-of-week review.

**Data inputs:**
- Week plan checklist: which non-negotiables were marked complete
- Signal documents: per-node priority checklist status
- "Waiting on others" table in `week-plan.md`: did delegated signals return?
- GitHub, Slack, deliverable artifacts: external evidence of completion

**Questions:**
- Did each non-negotiable get done? If not, why not?
- Were blockers resolved or carried?
- Did delegated tasks return as intended?

**Output:**
- Checked boxes, marked completions
- Carry-forwards (with explicit reason — not implicit drift)
- Fidelity check trigger for any delegated signal that did not resolve

**Where it lives:** End-of-week section in `week-plan.md`

**Time budget:** 5 minutes on Friday. If it takes longer, the week plan was too complex.

**Failure mode:** The single-loop becomes the only loop — the system gets faster at
executing without ever asking whether the work matters.

---

### Level 2 — Double-Loop (Strategic Feedback)

**Question:** Are we doing the right things?

**What it challenges:** The strategy behind the week's choices. Not "did we execute"
but "did executing move the needle, and is this the right needle?"

**Trigger:** Every Friday, after the single-loop check. Approximately 15 minutes.

**Data inputs:**
- Last 4 weeks of single-loop results (what shipped, what slipped)
- All active node statuses across the 12 nodes
- Revenue and output metrics (Node 11)
- Time allocation actual (week-plan.md time blocks) vs stated priorities

**Questions:**
- We did X. Did X move Y? (execution → outcome correlation)
- Are the right nodes getting the most time, or are we defaulting to comfort?
- Where are we busy but not making progress?
- What assumption drove last week's priorities — is that assumption still valid?
- If we run the exact same week again, what meaningfully changes?

**Output:**
- Reweighted node priorities for next week
- Kill or pause decisions for nodes not returning value
- Updated stated strategy in `alignment.md`
- Explicit "one thing we are NOT doing next week" — conscious de-prioritization

**Where it lives:** `weekly-review.md` — filled every Friday after single-loop check

**Time budget:** 15 minutes. Hard cap. If the review requires more time, the system has
accumulated too many unresolved strategic questions from prior weeks.

**Failure mode:** The double-loop becomes a feelings exercise — honest-sounding questions
that produce no concrete priority changes. If Part 5 of `weekly-review.md` looks
identical week over week, the double-loop is not functioning.

---

### Level 3 — Triple-Loop (Meta-Strategic Feedback)

**Question:** Are we asking the right questions?

**What it challenges:** The framework itself. Not "what are the right actions" but
"are the categories we use to evaluate actions still valid?"

**Trigger:** Last Friday of each month. 45 minutes minimum. Requires uninterrupted time.

**Data inputs:**
- Last 4 weekly reviews (the double-loop corpus)
- All `context.md` files across the 12 nodes (skim: are stated purposes still accurate?)
- `alignment.md` — full month trend, not just last week
- Node 11: actual revenue vs 90-day projection from 30 days prior
- Which double-loop questions recurred (a question appearing 3+ weeks is a framework problem)

**Questions:**
- Are we measuring the right things? (Measuring revenue per client when the real lever is platform adoption = optimizing the wrong variable)
- Is the portfolio of nodes still the right portfolio?
- Are the categories in the operating system still the right categories?
- What would a completely fresh start look like — and what does that reveal about baked-in assumptions?
- What have we been avoiding asking?

**Output:**
- Node restructure decisions (merge, kill, add)
- Metric rewrites per node (change what success means)
- `alignment.md` framework assumptions section updated
- 90-day forward question (not a goal — a question)

**Where it lives:** `monthly-review.md` — filled on the last Friday of each month

**Time budget:** 45 minutes budgeted. No maximum. This is the highest-leverage cognitive
work in the system. Under financial pressure, this session is the first thing eliminated
and the last thing that should be.

**Failure mode:** The triple-loop gets skipped when the system is under stress — exactly
when it is most needed. A system under pressure that stops questioning its own framework
will accelerate in the wrong direction.

---

## Fidelity Measurement

Fidelity is the measure of intent preservation through the signal chain.

```
Alice forms intent
      |
      v
Alice encodes signal  (email, doc, verbal instruction, task)
      |
      v
Signal transmits        (receiver reads, hears, or receives it)
      |
      v
Receiver decodes signal (their interpretation of the intent)
      |
      v
Receiver acts
      |
      v
FIDELITY CHECK: Did the action match the original intent?
```

**Fidelity Score per Signal:**

| Score | Meaning |
|-------|---------|
| 1.0 | Intent fully preserved — action matched exactly as intended |
| 0.75 | Minor interpretation gap — direction correct, scope or detail slightly off |
| 0.5 | Partial match — right general direction, wrong execution |
| 0.25 | Significant gap — some relationship to intent, but not what was meant |
| 0.0 | No match — signal failed somewhere in the chain |

**Fidelity failure modes:**

| Failure Type | Diagnosis | Fix |
|--------------|-----------|-----|
| Encoding failure | Alice stated the intent unclearly or incompletely | Improve signal encoding: use M/G/T/F/W classification on delegated tasks |
| Transmission failure | Signal never reached the receiver | Confirm receipt explicitly before assuming transmission |
| Decoding failure | Receiver interpreted differently than intended | Add context to encoding; verify understanding before action starts |
| Action failure | Receiver understood but executed differently | Escalate — this is a trust/ownership question, not a communication question |

**Where it lives:** `signal-fidelity` table in each node's `signal.md` file.
Add one row per non-trivial delegated task. Check at end-of-week single-loop review.

**Aggregate signal fidelity** is tracked in `alignment.md` — Signal Fidelity History
section — as a weekly average. A node with persistent low fidelity (two or more
consecutive weeks below 0.5) has a communication architecture problem,
not a people problem.

**Fidelity threshold for architectural concern:**
- Single node, single week below 0.5: log it, watch it
- Single node, two consecutive weeks below 0.5: diagnose failure mode
- Any node, average below 0.5 for one month: treat as L3 structural issue,
  raise in triple-loop review

---

## Alignment Drift Detection

Alignment drift is the gap between what the system says it is for and what it
actually does with time and resources.

A system can drift without any individual decision being wrong. Drift accumulates
through thousands of small, locally-reasonable choices that collectively diverge from
stated intent. The only way to catch it is to compare stated purpose against actual
allocation on a regular cadence.

**Drift Dimensions (scored 1–5 each week):**

| Dimension | 5 (Aligned) | 3 (Drifting) | 1 (Severe Drift) |
|-----------|-------------|-------------|-----------------|
| Time → Stated Node Priorities | Top 3 priority nodes got top 3 time allocation | 2 of 3 top nodes got proportional time | Time went to comfort/default nodes regardless of priority |
| Activities → 90-Day Goals | Every major activity links directly to a 90-day goal | Most activities link, some are faith-based | Activities don't connect to any stated 90-day goal |
| Revenue Focus → Actual Closes | Pipeline is moving, conversations converting | Activity happening, slow conversion | Revenue activity without measurable pipeline progress |
| Delegated Signals → Outcomes | Avg fidelity ≥ 0.8 across all delegated signals | Avg fidelity 0.5–0.8, some gaps | Avg fidelity < 0.5, signals not surviving transmission |

**Total drift score: 4–20**
- 17–20: Aligned. Continue current approach.
- 13–16: Minor drift. Investigate the lowest-scoring dimension.
- 9–12: Structural drift. Identify root cause before next week starts.
- Below 8: Stop. Course correct before executing anything further.

**Drift detection examples:**

```
EXAMPLE 1 — Time vs. Stated Priority
Stated priority: "Platform first — MIOSA is the primary focus."
Actual time allocation: 60% to agency client work (Node 03), 15% to MIOSA (Node 02).
Drift type: Resource allocation drift.
Action: Either update stated priority to reflect reality, or restructure the week plan
to enforce the stated priority. Both are valid — but the gap cannot persist unexamined.

EXAMPLE 2 — Identity vs. Revenue Composition
Stated identity: "We are a technology platform company."
Actual revenue composition: 80% from agency services, 20% from platform.
Drift type: Identity drift.
Action: Triple-loop question — "Should we stop calling ourselves a platform company
until platform revenue exceeds 50% of total?" This is not a failure. It is honest data.

EXAMPLE 3 — S4 vs. S3 Time Balance
System 4 requirement: 10% of Alice's weekly time on environmental scanning.
Actual allocation: 0% in last 3 weeks (all time consumed by S3 operational urgency).
Drift type: Strategic attention drift.
Action: Protect one 90-minute block per week as S4 time. Non-negotiable.
```

**Where it lives:** `alignment.md` — updated at double-loop review (weekly drift scores)
and triple-loop review (framework assumptions section).

---

## Existing Code Assets

Layer 6 has the strongest existing code coverage of any layer.

### `miosa_memory.Learning` — SICA Self-Improvement Engine

SICA (Systematic Iterative Correction Algorithm) runs the machine equivalent
of double-loop learning: it observes system behavior, reflects on patterns,
proposes corrections, tests them, and integrates successful fixes.

```
OBSERVE  → Monitor system outputs and outcomes
REFLECT  → Identify patterns and gaps
PROPOSE  → Generate correction hypotheses
TEST     → Apply correction in controlled context
INTEGRATE → If improvement confirmed, persist the pattern
```

SICA is the automated Layer 6 running continuously at the code level. The three-loop
framework (single/double/triple) is the human-level equivalent, running on weekly
and monthly cadences. Both serve the same Wiener constraint: closed loops.

### `miosa_memory.Episodic` — Event Logging with Temporal Decay

Every significant system event is logged as an episodic memory with:
- Timestamp and node context
- Event type (action taken, outcome observed, signal sent)
- Temporal decay weight (recent events weighted higher in retrieval)

Episodic memory is the raw data source for the single-loop check. The Friday
end-of-week review is the human reading the episodic log and verifying completions.

### `miosa_memory.Cortex` — Cross-Session Pattern Detection

Cortex synthesizes episodic memories across sessions to detect patterns that
no single session can see. This is the automated triple-loop: Cortex identifies
recurring failure patterns, recurring questions, and persistent drift signals.

Cortex output feeds the monthly review by surfacing: "These 3 things have been
unresolved for 4+ weeks" — the kind of pattern a human reviewer will miss
without systematic cross-session tracking.

### `miosa_signal.FailureModes` — Real-Time Signal Quality Feedback

The 11 failure mode detectors in `miosa_signal` run L6 at the signal level:
before a signal leaves the system, it is assessed for the 11 failure modes
(Routing Failure, Bandwidth Overload, Genre Mismatch, etc.). This is a real-time
single-loop check at Layer 2 — immediate feedback before the signal enters
the transmission phase.

---

## Layer 6 Feedback Architecture Summary

```
FEEDBACK LAYER (L6) — FULL SYSTEM VIEW

HUMAN-LEVEL FEEDBACK:

  Per-signal  →  signal-fidelity table in each signal.md
                 Did the delegated signal produce intended outcome?

  Weekly      →  week-plan.md end-of-week (single-loop, 5 min)
                 weekly-review.md (double-loop, 15 min)
                 alignment.md drift score update

  Monthly     →  monthly-review.md (triple-loop, 45 min)
                 alignment.md framework assumptions update
                 Node portfolio audit and metric rewrites

MACHINE-LEVEL FEEDBACK (miosa_memory):

  Continuous  →  miosa_signal.FailureModes (pre-transmission signal quality)
  Session     →  miosa_memory.Episodic (event log with temporal decay)
  Cross-session  miosa_memory.Cortex (pattern detection across sessions)
  Iterative   →  miosa_memory.Learning / SICA (self-improvement engine)

ESCALATION PATH:

  Signal fidelity < 0.5 for 2 weeks
    → double-loop strategy question: is this a communication architecture issue?

  Double-loop question recurring for 3+ weeks
    → triple-loop trigger: is this a framework problem?

  Alignment drift score < 8 any week
    → stop execution, course correct immediately (S5 review)

  Cortex flags persistent pattern
    → surfaced in Monday brain dump, Alice reviews before week plan
```

---

## Cadence Reference

| Loop | Cadence | Duration | Trigger | Primary Artifact | Output |
|------|---------|----------|---------|-----------------|--------|
| Single-loop | Weekly (Friday) | 5 min | End-of-week review | `week-plan.md` | Checked boxes, carry-forwards |
| Fidelity check | Weekly (Friday) | 5 min | Any delegated task past deadline | `*/signal.md` fidelity table | Fidelity score per signal |
| Double-loop | Weekly (Friday) | 15 min | After single-loop | `weekly-review.md` | Reprioritization, `alignment.md` update |
| Triple-loop | Monthly (last Friday) | 45 min | After 4 double-loops | `monthly-review.md` | Node restructure, framework update |
| Cortex synthesis | Session-triggered | Auto | Cross-session pattern detected | `miosa_memory.Cortex` | Pattern bulletin |
| SICA improvement | Continuous | Auto | Error pattern threshold | `miosa_memory.Learning` | Correction proposal |

---

## Integration with Other Layers

| Layer | L6 Relationship |
|-------|----------------|
| L1 — Network | L6 verifies signal routing was correct: did it reach the right node? |
| L2 — Signal | `miosa_signal.FailureModes` is an L2-triggered L6 check |
| L3 — Composition | L6 evaluates whether genre selection produced intended decoding |
| L4 — Interface | L6 checks whether the tiered disclosure was appropriate for receiver bandwidth |
| L5 — Data | L6 writes fidelity scores and drift data back into L5 storage for retrieval |
| L7 — Governance | Triple-loop outputs feed L7 governance decisions; algedonic triggers are detected by L6 |

---

## References

- Wiener, N. — *Cybernetics: Or Control and Communication in the Animal and the Machine* (1948)
- Argyris, C. — *Double Loop Learning in Organizations* (Harvard Business Review, 1977)
- Senge, P. — *The Fifth Discipline: The Art and Practice of the Learning Organization* (1990)
- Beer, S. — *Brain of the Firm* (2nd ed., 1981) — algedonic channel and feedback integration
- Luna, Alice H. — *Signal Theory: The Architecture of Optimal Intent Encoding* (MIOSA Research, Feb 2026)
- ADR-001: Feedback Loop Architecture — `/Users/rhl/Desktop/OptimalOS/tasks/ADR-001-feedback-loop-architecture.md`
- Working documents: `feedback-loops.md`, `weekly-review.md`, `monthly-review.md`, `alignment.md`
