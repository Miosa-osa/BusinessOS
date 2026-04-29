# Layer 7: Governance

> **Governing Constraint:** Beer — viable structure
> **One-Line Purpose:** System 1-5, agent autonomy levels, algedonic bypass, autopoiesis
> **Existing Code:** New (`miosa_context.Governance`) — primary gap in the architecture

---

## What This Layer Does

Layer 7 is the outermost shell. It wraps every other layer — not as a policy list or
a rulebook, but as the structural condition for the entire system to remain viable.

Viable means something specific here, borrowed from Beer's Viable System Model: a system
is viable when it can maintain its identity in a changing environment. Not survive. Not
persist. *Maintain identity.* The distinction matters because a system can persist while
losing what it was for — it becomes a different system wearing the same name. Governance
exists to prevent that.

L7 does not manage operations. Operations are S1's domain. L7 does not run strategy.
Strategy is S4. L7 asks the question that no other layer asks: *Is this still the right
system to be?*

Without L7, the system has no self-model. It can be excellent at executing while drifting
away from the purpose that justified its existence. The feedback loops in L6 detect
execution drift. L7 detects existential drift — the gap between what the system *is*
and what it was built to *become*.

```
L7 GOVERNANCE — THE OUTERMOST LAYER

  ┌─────────────────────────────────────────────────────────────────────┐
  │  S5 — POLICY (Identity + Purpose)                                    │
  │  Alice's governing philosophy. What OptimalOS will and won't do.   │
  │                                                                       │
  │  ┌───────────────────────────────────────────────────────────────┐   │
  │  │  S4 — INTELLIGENCE (Outside/Future)                           │   │
  │  │  Environmental scanning. Threat detection. Opportunity radar. │   │
  │  │                                                               │   │
  │  │  ┌───────────────────────────────────────────────────────┐   │   │
  │  │  │  S3 — CONTROL (Inside/Now)                            │   │   │
  │  │  │  Resource allocation. Performance monitoring.         │   │   │
  │  │  │                                                       │   │   │
  │  │  │  S3* — AUDIT                                         │   │   │
  │  │  │  Independent verification. S1 sampling.              │   │   │
  │  │  │                                                       │   │   │
  │  │  │  ┌─────────────────────────────────────────────────┐ │   │   │
  │  │  │  │  S2 — COORDINATION                               │ │   │   │
  │  │  │  │  Prevents oscillation. Communication standards. │ │   │   │
  │  │  │  │                                                  │ │   │   │
  │  │  │  │  ┌───────────────────────────────────────────┐  │ │   │   │
  │  │  │  │  │  S1 — OPERATIONS (Production Units)        │  │ │   │   │
  │  │  │  │  │  Actual work. Market interaction. Output. │  │ │   │   │
  │  │  │  │  └───────────────────────────────────────────┘  │ │   │   │
  │  │  │  └─────────────────────────────────────────────────┘ │   │   │
  │  │  └───────────────────────────────────────────────────────┘   │   │
  │  └───────────────────────────────────────────────────────────────┘   │
  │                                                                       │
  │  ALGEDONIC CHANNEL (emergency bypass — fires directly to S5)         │
  └─────────────────────────────────────────────────────────────────────┘
```

---

## Beer's Viable System Model Applied to OptimalOS

The Viable System Model (Beer, 1972) identifies five necessary and sufficient functional
systems for any organization to remain viable. If any of the five is missing or severely
degraded, the organization loses viability — it either collapses, merges with another,
or becomes something different from what it was.

### The Five Systems

| System | VSM Role | What It Does | OptimalOS Instance |
|--------|----------|-------------|-------------------|
| **S1** | Operations | Executes the primary activities that produce value | Each Operation node — ClinicIQ, AI Masters, MIOSA Platform, Lunivate Agency, OS Accelerator, Agency Accelerants, OS Architect, Content Creators |
| **S2** | Coordination | Prevents oscillation between S1 units; resolves resource conflicts; maintains schedule coherence | `node:unit-team` (Team, Node 10) — shared developer allocation, calendar coordination, cross-operation handoffs |
| **S3** | Control | Resource allocation authority; performance monitoring; standards enforcement; internal audit | Alice (as operator) acting on Node 11 (Finance) data — audits actuals, reallocates resources, enforces standards |
| **S3\*** | Audit | Independent direct sampling of S1 behavior; bypasses S2 reporting chain | End-of-week review: shipped vs. planned; Node 11 actuals vs. projections |
| **S4** | Intelligence | Environmental scanning; opportunity detection; threat identification; future modeling | `node:op-new-stuff` (Node 09) — research, competitive intelligence, strategic experiments |
| **S5** | Policy | Identity, values, purpose — what the system *is* and what it will *not* become | Alice (as identity-holder) — final arbiter of S3/S4 balance, source of the system's self-description |

---

### System 1 — Operations (The Production Units)

**Beer's definition:** The actual work units that produce output and interact directly with
their environments. Each S1 unit has operational autonomy within policy bounds set by
System 5. The outputs of S1 units are what the system actually delivers to the world.

**OptimalOS implementation:**

| S1 Unit | Node | Accountable Person | Output | Revenue Logic |
|---------|------|--------------------|--------|---------------|
| MIOSA Platform | 02 | Erin / Nina / Frank | Tech infrastructure, VM compute, agent wallets | Recurring compute fees + $50K app/module builds |
| Lunivate Agency | 03 | Frank / Judy | Client deliverables (ClinicIQ, Mosaic Effect, Atlas Cards) | Project fees ($50K+), database reactivation |
| AI Masters | 04 | Ed / Robert P. / Sam | Courses, coaching, cohorts | Ed group $20K/mo, course sales |
| Agency Accelerants | 06 | Dan / Sam | Acceleration program, community | $99/mo community (~$10.7K/mo), coaching |
| OS Architect | 05 | Ivan | YouTube content, lead generation | Lead gen → funnel → paid |
| OS Accelerator | 12 | TBD | Education platform | Subscription + cohort |
| Content Creators | 08 | Ivan / Judy | Distributed content | Feed all other S1s |
| New Stuff | 09 | Alice | Research, scanning | This IS S4 — not an S1 revenue unit |

**S1 structural requirement:** Each unit must have:
1. One named accountable person — not Alice as day-to-day operator
2. A `signal.md` file that is the unit's weekly state broadcast
3. A `context.md` file that is the unit's persistent memory
4. Operational autonomy: the accountable person makes day-to-day decisions without escalating to Alice for routine decisions

**Critical S1 pathology:** Alice's recognized god-complex routes S1 operational decisions back
through him, collapsing the S1/S3 boundary. The result is a single point of failure at the
system's most productive layer. The fix is not working harder — it is structural: each S1 unit
lead must have explicit authority to operate without Alice in the execution loop.

---

### System 2 — Coordination (Preventing Oscillation)

**Beer's definition:** The mechanism that prevents S1 units from interfering with each other.
Without S2, S1 units compete for shared resources (Alice's time, developer capacity, budget)
and create oscillation — cyclical instability as units repeatedly disrupt each other.

**The oscillation problem in OptimalOS:**
Every S1 unit wants Alice's attention simultaneously. Without S2, they interrupt him in
parallel, causing context-switching chaos, dropped commitments, and the chronic overload
pattern that produces 48–72 hour coding sessions (S3 breakdown disguised as heroism).

**S2 mechanisms:**

| S2 Mechanism | Artifact | Function |
|---|---|---|
| Weekly Dump | `weekly-dump.md` | Batch capture of all S1 unit state — one channel, one time per week. Converts parallel interrupt storms into a sequential batch |
| Signal Documents | `*/signal.md` per node | Standardized state format — every unit speaks the same genre so Alice reads 12 signals in 20 minutes, not 12 ad-hoc formats |
| Signal Theory Encoding | M/G/T/F/W on all outputs | The communication standard. Prevents S1 units from sending noise-heavy requests that consume S3 bandwidth |
| Weekly Template | `weekly-template.md` | Structural protocol for S2 synchronization — same schema every cycle, no format drift |
| Node Ownership | Accountable person per node | Prevents resource conflicts — one person owns each S1, they coordinate laterally before escalating |

**S2 non-negotiable:** The Monday brain dump is not optional. It IS the S2 mechanism.
Skipping the dump breaks coordination and forces Alice into reactive mode.

**S2 gaps (March 2026):**
- No lateral coordination protocol between S1 units (Ivan's content feeds Dan's funnel —
  this handoff currently routes through Alice)
- No shared developer capacity booking (Erin/Nina/Frank pulled by multiple S1 units without
  a weekly allocation table)

---

### System 3 — Control (Resource Allocation + Performance)

**Beer's definition:** The inside-now function. S3 allocates resources between S1 units,
monitors performance against policy, and resolves resource conflicts. S3 is not operational
management of S1 units — it is governance of the space *between* S1 units.

**OptimalOS implementation:**

| S3 Function | Artifact | Operation |
|---|---|---|
| Resource allocation | `week-plan.md` — Time Blocks section | Explicit time budget assigned per node per day. If it is not in the time blocks, it is not allocated |
| Revenue monitoring | `nodes/11-money-revenue/signal.md` | Weekly revenue state: what closed, what is close, what must close to hit $25K floor |
| Priority triage | Top 3 Non-Negotiables in `week-plan.md` | Explicit S3 decision: these three items override all competing demands for the week |
| Conflict resolution | Alice's week plan choices | When two S1 units need Alice simultaneously, the week plan resolves it pre-emptively, not reactively |
| Developer allocation | Node 10 team signal | Erin/Nina/Frank assigned per S1 unit for the week — weekly capacity table |

**S3 financial thresholds:**

| Status | Monthly Recurring Revenue | System Response |
|--------|--------------------------|-----------------|
| Crisis | Below $15K | Algedonic fires → S5 |
| Warning | $15K–$25K | S3 concentrates all top 3 on revenue |
| Survival | $25K | Minimum viable. Continue current allocation |
| Target | $100K | Strategic reinvestment unlocked |
| Current (Mar 2026) | ~$30.7K | Ed $20K + Dan community $10.7K |

---

### System 3* — Audit (Independent Verification)

**Beer's definition:** The audit channel. S3* bypasses the management hierarchy and directly
samples S1 behavior to verify that what S3 is told matches what is actually happening.
S3* catches drift, misreporting, and execution gaps that S3 visibility cannot detect.

This is not punitive. It is a structural viability requirement.

**OptimalOS implementation:**

| S3* Function | Artifact | Operation |
|---|---|---|
| Outcome verification | `week-plan.md` — End of Week Review | "What shipped" vs "what was planned" — the delta IS the audit signal |
| Signal accuracy check | `signal.md` (stated) vs actual deliverables | If a node's signal says "in progress" for 3 consecutive weeks without an artifact, S3* fires |
| Revenue reconciliation | Node 11 actuals vs projections | Weekly close-rate audit: did predicted closes happen? |
| Team output audit | Deliverable review against task assignments | Node 10: did team members ship what they committed to? |

**S3* specific triggers:**
- Any node with "building" or "in progress" status for 3+ consecutive weeks without a shippable artifact
- Any sales commitment that does not convert within stated timeline
- Any team member whose output cannot be verified from concrete deliverables
- Node 11 gap-to-target widening for two consecutive weeks

---

### System 4 — Intelligence (Environmental Scanning + Future Planning)

**Beer's definition:** The outside-future function. S4 scans the environment, identifies
threats and opportunities, and models future states. S4 is the only subsystem that looks
outside the current operations. Without S4, the system is blind to environmental change
until that change is already causing damage.

**The S3/S4 asymmetry — the primary viability risk:**
Under financial pressure, S4 time is systematically stolen by S3 urgency. A system that
allocates 100% of time to S3 (current operations) cannot detect the environment that will
kill it.

**OptimalOS implementation:**

| S4 Function | Artifact | Operation |
|---|---|---|
| Competitive scanning | `nodes/09-new-stuff/context.md` | Active watch: Fiova (Saeed), NVIDIA NeMo Claw, Pulsio, Perplexity computer use |
| Opportunity identification | Node 09 weekly signal | Ivan clipper startup, Tina Motley 300K, Kong university play, Dom VC connector |
| Strategic partnerships | Node 09 + Node 01 context | Oscar/Peggy Consortium AI, Carol platform partnership, Capital Factory |
| Threat modeling | Node 09 — requires externalization | All active competitive threats should live in Node 09, not in Alice's head |
| Future scenario planning | `monthly-review.md` | 90-day forward question (not a goal — a question) |

**S4 time protection rule:** Alice must protect a minimum of 10% of his weekly time for
S4 activity. Current state (March 2026): 0–2% of Alice's week. Target: 10–15%.
S5 must enforce this structurally — S3 will fill all available time without constraint.

**S4 cadence:**
- Weekly: 15-minute Node 09 update during Monday dump
- Monthly: 1-hour competitive landscape review (currently a gap — not happening)
- Quarterly: Strategic planning session with Oscar (Consortium-level S4)

---

### System 5 — Policy (Identity, Values, Purpose)

**Beer's definition:** The identity of the system. S5 defines what the system will and
will not become. It resolves the tension between S3 (optimize now) and S4 (adapt for
future) when they conflict. S5 is the final arbiter.

**S5 governing statement:**

> "Every output is a Signal. Maximize Signal-to-Noise Ratio. Zero exceptions."
>
> — *Signal Theory: The Architecture of Optimal Intent Encoding*, Alice H. Luna (MIOSA Research, Feb 2026)

This is not a productivity principle. It is S5 policy. It defines what MIOSA builds
(high-signal infrastructure), how Alice communicates (Signal Theory M/G/T/F/W),
what Alice refuses (VC money, noise-addicted partners, bullshit-first startups),
and what the Consortium stands for.

**S5 policy artifacts:**

| S5 Element | Artifact | Content |
|---|---|---|
| Governing philosophy | Signal Theory paper (63 pages, Feb 2026) | Complete theoretical framework governing all output and communication |
| Values + non-negotiables | `nodes/01-roberto/context.md` | CPU-at-edge, no VC, self-funding, open source distribution, developer royalty framework |
| Identity boundaries | Consortium model constraints | Who Alice partners with, what he builds, what he will not sell |
| Legal structure | MIOSA LLC (TX) + Lunivate LLC (MI) | Entity separation encodes business logic at legal level |

**Active S5 tensions requiring active management:**

| S3 Pressure | S5 Boundary | Resolution |
|---|---|---|
| "Close deals NOW to survive" | "No bad-fit clients even under financial pressure" | S5 wins on quality; S3 must find volume from right-fit clients faster |
| "Move before NVIDIA launches" | "Don't ship garbage — every output is a Signal" | S5 sets quality floor; S4 sets deadline; S3 plans the sprint |
| "Alice should do everything" | "Alice must grow 4 years by tomorrow" (Oscar's mandate) | S5 wins — god-complex violates S5 identity. Napoleonic Complex = noise injection into S1 |

---

### Recursion: The VSM Is Fractal

Every viable system contains its own S1–S5 internally. This is the VSM's most important
structural insight, and the one most commonly ignored. OptimalOS as a whole has its five
systems. Each S1 operation within it also has its own five systems.

**AI Masters — internal VSM:**

```
AI MASTERS (node:op-ai-masters) — Internal VSM

  S1: Alice — delivers course content and coaching
  S1: Bob / Sam — executes sales conversations and pipeline

  S2: Coordination between Ed (delivery) and Robert/Sam (sales) —
      capacity matching: don't sell cohorts faster than Ed can deliver them

  S3: Alice (acting as AI Masters controller) — reviews enrollment metrics,
      revenue-per-cohort, delivery quality; reallocates time and resources

  S4: Alice + Ed — what is the next cohort design? What is the competitor
      landscape for AI training? What is NVIDIA doing with NeMo Claw?

  S5: Alice — what is AI Masters *for* in the portfolio? When does it
      get a dedicated lead vs. remain under Alice?
```

**Full recursive mapping across all operations:**

| Operation | S1 (Who executes) | S2 (What coordinates) | S3 (Who controls) | S4 (Who scans) | S5 (What is it for) |
|-----------|------------------|----------------------|------------------|---------------|-------------------|
| MIOSA Platform | Erin, Nina, Frank | Team capacity table | Alice + Carol | Alice + Carol | Build the platform to exit service dependency |
| Lunivate Agency | Frank, Judy, contractors | Project scheduling | Alice | Alice | Fund platform build; maintain until platform revenue > 50% |
| AI Masters | Ed, Robert P., Sam | Ed ↔ sales capacity | Alice | Alice + Ed | Education revenue; platform adoption pathway |
| Agency Accelerants | Dan, Sam | Community calendar | Alice + Dan | Alice | Build audience for OS Accelerator conversion |
| OS Architect | Ivan | Content calendar | Alice + Ivan | Ivan | Build Alice's public platform and audience |
| OS Accelerator | Alice (lead) | None yet | Alice | Alice | Premium cohort: highest-leverage revenue per hour |
| Content Creators | Ivan, Judy | Content calendar | Alice | Ivan | Expand distribution; feed all other S1s |

**Recursion applies to agents as well.** Any OSA agent operating at L3 or above must have its
own internal VSM architecture: S1 (task execution), S2 (tool coordination), S3 (self-monitoring),
S4 (context scanning), S5 (governing philosophy — Signal Theory, scope boundaries, quality floor).
An agent that lacks any of these internal systems will fail in ways that better instructions
cannot fix.

---

## Agent Autonomy Levels (L1–L5)

Agent autonomy is not a global property. It is a per-domain grant, bounded by VSM scope,
and logged every time it changes.

### The Five Levels

| Level | Name | Scope | Decision Rights | VSM Scope | Example Default Agents |
|-------|------|-------|----------------|-----------|----------------------|
| **L1** | Operator | Execute instructions exactly | None. Human decides, agent executes. | S1 execution only | `@database-specialist` (schema changes) |
| **L2** | Collaborator | Suggest alternatives, surface options, shared planning | Can propose; human approves all actions | S1 + S2 coordination input | `@database-specialist` (query optimization), `node:endpoint-nejd`, `node:endpoint-tejas` |
| **L3** | Consultant | Recommend, plan, prototype with rationale | Can plan, recommend, execute within established patterns; human ratifies consequential decisions | S1 + S2 + S3 input | `@architect`, `@backend-elixir`, `@backend-go`, `@security-auditor`, `@oracle`, OSA Master |
| **L4** | Approver | Approve within pre-set constraints | Can approve, execute, verify within explicit boundaries; escalates outside bounds | S1–S3 full; advisory to S4 | `@debugger`, `@code-reviewer`, `@performance-optimizer`, `@test-automator` |
| **L5** | Autonomous | Full authority within domain | Full cycle: plan, decide, execute, review, iterate. Bounded only by domain definition and S5 policy. | S1–S4 full; policy from S5 | None currently. Reserved for future high-trust, high-maturity agents. |

### Autonomy Level Rules

**Rule 1: Domain-scoped, not global.**
An agent can hold L4 autonomy in code review and L2 autonomy in financial decisions
simultaneously. The domain definition bounds the grant — not the agent identity.

```
EXAMPLE:
node:agent-debugger
  autonomy:
    code_fix_isolated_bug: L4    # Can fix, test, verify, commit
    architecture_change: L2      # Can propose; Alice approves
    financial_decision: L1       # Not in scope at any level
    security_critical_path: L3   # Can recommend; @security-auditor reviews
```

**Rule 2: Temporary elevation is permitted with explicit reason and expiry.**
When a time-sensitive situation requires an agent to act above its default level,
the elevation must state: the elevated level, the reason, and the expiry condition.

```
ELEVATION RECORD FORMAT:
agent: node:agent-devops
elevated_from: L3
elevated_to: L4
domain: production_deployment
reason: "Critical hotfix; Alice unavailable; pre-approved deployment window"
expiry: "2026-03-17T18:00:00Z OR when Alice reviews"
authorized_by: node:endpoint-roberto
logged: true
```

**Rule 3: Temporary restriction is also permitted.**
If an agent has taken an action that produced unexpected consequences, its autonomy level
can be temporarily reduced pending investigation, without removing the agent from service.

**Rule 4: Every autonomy level change must be logged.**
The log is a governance record, not a developer note. It lives in `miosa_context.Governance`
and is reviewable in the monthly governance review.

**Rule 5: Autonomy levels are reviewed quarterly.**
Agents with 10+ consecutive clean completions without escalation → candidate for promotion.
Agents with 3+ escalations in a single quarter → review for protocol improvement.

### Current Agent Registry

| Agent | Domain | Default Level | Escalation Condition |
|-------|--------|--------------|---------------------|
| OSA Master Orchestrator | All domains | L3 | Any S5-policy question → Alice |
| `@architect` | System design, ADRs | L3 | Database schema, cross-team impact → escalate |
| `@debugger` | Bug isolation and fix | L4 | Fix touches architecture → drop to L3 |
| `@code-reviewer` | Code quality | L4 | Security-critical findings → co-review with `@security-auditor` |
| `@security-auditor` | Security | L3 | Critical vulnerability (A01/A02) → algedonic bypass to Alice |
| `@backend-elixir` | Elixir/Phoenix/OTP | L3 | Database schema → co-review with `@database-specialist` |
| `@backend-go` | Go services | L3 | New service creation → `@architect` approval |
| `@frontend-svelte` | Svelte frontend | L3 | Design system changes → `@osa-frontend-design` |
| `@frontend-react` | React frontend | L3 | Design system changes → `@osa-frontend-design` |
| `@devops-engineer` | Infrastructure | L3 | Production deployment → Alice approval |
| `@database-specialist` | Database queries | L2 | Schema change → L1 (Alice approves all) |
| `@test-automator` | Test suites | L4 | Test suite regression → L3 |
| `@oracle` | AI/ML architecture | L3 | New model deployment → Alice + S4 context |
| `@performance-optimizer` | Performance analysis | L4 | >20% behavior change → L3 |

### Autonomy and VSM: The Structural Connection

Agent autonomy levels map to VSM scope because autonomy is a governance concept,
not a capability concept. The question is not "what can the agent do technically?"
but "what is the agent *sanctioned* to decide on behalf of the system?"

```
L1–L2 agents operate inside S1 only.
  They execute and suggest. They do not commit. They do not coordinate.
  Their decisions affect one operation in one moment.

L3 agents can inform S2 and S3.
  Their recommendations can affect resource allocation, standards, and
  cross-operation coordination — but a human ratifies before action is taken.

L4 agents operate inside S3.
  They can approve within pre-set bounds. Their decisions may affect
  multiple S1 operations. Escalation conditions are explicit.

L5 agents (if ever deployed) would operate up to S4.
  They could run environmental scanning, identify structural opportunities,
  and propose S5-level policy changes. S5 policy itself remains human-only.
  S5 is identity. Identity is not delegated.
```

---

## Algedonic Bypass

### Definition

An algedonic signal is a pain or pleasure signal that bypasses the normal communication
hierarchy when the system's viability is threatened. The word is Beer's: *algedonic* from
Greek *algos* (pain) + *hedone* (pleasure). It is the system's nervous system, not its
management chain.

Normal signals travel through proper channels: S1 → S2 → S3 → S4 → S5.
Algedonic signals skip all intermediate systems and go directly from source to S5.

This is not an emergency notification. It is a structural bypass. The hierarchy exists to
manage complexity under normal conditions. When viability itself is under threat, managing
complexity is not the priority — responding to the threat is.

**Why bypass is necessary:** Without the algedonic channel, a viability-threatening signal can
be normalized away ("we've been here before"), delayed in the S2 coordination queue, deprioritized
by S3 in favor of ongoing operations, or missed entirely by S4 because the threat comes from
inside. By the time it reaches S5 through normal channels, it is too late to act.

### Trigger Conditions

| Category | Trigger | Threshold | S5 Action Required |
|----------|---------|-----------|------------------|
| **Revenue** | Recurring monthly revenue drops below floor | MRR < $15,000 (confirmed, not projected) | Emergency S3/S4 session within 24h |
| **Revenue** | Single client above concentration threshold signals termination | Client > 30% of MRR signals churn | Immediate S3 review; diversification decision |
| **Team** | Key technical team member departs | Erin, Nina, or Frank exits | Platform delivery plan reassessment within 48h |
| **Team** | Senior contractor relationship fractures | Carol or Dan relationship breaks down | Operational continuity plan within 48h |
| **Competitive** | Strategic competitor launches in primary space | NVIDIA NeMo Claw goes live; direct education competitor emerges | S4 emergency brief within 72h |
| **Legal** | Legal or compliance threat materializes | Any legal demand, IP dispute, employment claim | Legal consultation within 24h — do not respond without counsel |
| **Security** | Security breach or data exposure | Any unauthorized access to MIOSA platform or client data | Immediate isolation, incident response within 1h |
| **Strategic Partnership** | Critical partnership fractures | Oscar/consortium relationship collapses | S5 strategic review within 24h |
| **Identity** | Mission-critical decision committed without S5 review | Agent or team commits system to a path that changes its identity | Immediate recall and S5 review |
| **System** | Governance coverage falls below floor | Governance coverage metric < 50% for two consecutive weeks | System health review within 1 week |

### Mechanism

```
ALGEDONIC SIGNAL FLOW

Standard path (normal conditions):
  S1 operation signal
    → S2 coordination queue
      → S3 control review
        → S4 environmental context
          → S5 decision (if required)

Algedonic path (viability threat detected):
  Trigger detected (any node, any layer)
    → BYPASS: skip S2, S3, S4 queues
      → S5 directly: node:endpoint-roberto [ALGEDONIC: <category>/<trigger>]
        → All other queues deprioritized pending S5 response
          → S5 decides: respond, escalate, or dismiss
            → If dismissed: reason logged, normal processing resumes
            → If acted on: decision trace created, S3/S4 notified of outcome
```

**Implementation in `weekly-dump.md`:**
The `## CRITICAL FLAGS` section appears at the top of `weekly-dump.md`, before any node-by-node
review. Any active trigger must be filled in FIRST, before S3 processing begins. This forces
S5-level attention before the system resumes S3 optimization.

### Specific Trigger Protocols

**Trigger: Revenue Below Survival Threshold**
```
Condition:   Confirmed recurring revenue below $15K/month
Signal:      ALGEDONIC: REVENUE CRITICAL
S5 Response: Alice suspends all S4 activity.
             S3 concentrates entirely on survival revenue.
             Top 3 non-negotiables = all revenue-related until $25K floor restored.
Current:     Near-threshold active (March 2026)
```

**Trigger: Key Team Member Departure**
```
Condition:   Frank, Nina, Erin, or Dan becomes unavailable
Signal:      ALGEDONIC: TEAM CAPACITY CRITICAL
S5 Response: Immediate team audit. Coverage gap identified within 48 hours.
             Pause lowest-priority S1 unit.
             Do NOT expand scope while under-staffed.
Precedent:   Past team departures created emergency revenue situations
```

**Trigger: Strategic Competitor Launches**
```
Condition:   NVIDIA NeMo Claw goes live, Pulsio hits $10M revenue,
             or any competitor captures Alice's primary market segment
Signal:      ALGEDONIC: COMPETITIVE THREAT ACTIVE
S5 Response: S4 emergency session. Evaluate whether CPU-at-edge differentiation holds.
             If yes: accelerate launch immediately.
             If no: S5 pivot — this is the rare case where identity is renegotiated.
Current:     NVIDIA NeMo Claw = active pre-launch threat (March 2026)
```

**Trigger: Strategic Partnership Fracture**
```
Condition:   Carol relationship breaks down, Oscar exits Consortium,
             or Tina Motley (300K members) withdraws distribution
Signal:      ALGEDONIC: STRATEGIC DEPENDENCY CRITICAL
S5 Response: Immediate intervention. Oscar is the appointed mediator
             (proven: brokered Alice-Carol peace).
             Invoke mediation protocol. Alice does NOT attempt to resolve alone.
```

**Trigger: Client Churn Risk**
```
Condition:   ClinicIQ/HBAI or Mosaic Effect client signals termination risk
Signal:      ALGEDONIC: CLIENT CHURN RISK
S5 Response: Alice takes direct ownership of relationship recovery.
             Deliverable review + honest conversation within 24 hours.
```

Every algedonic signal is logged permanently in `miosa_context.Governance` regardless of
whether S5 acts on it or dismisses it. The historical record of algedonic triggers is itself
a governance artifact — it shows where the system has been under threat and what responses
were chosen.

---

## Autopoiesis Protocol

Autopoiesis (Maturana and Varela, 1972) is the capacity of a system to produce and maintain
itself. The word means "self-creation." An autopoietic system is not maintained by external
forces — it maintains itself through its own internal processes.

For OptimalOS, this means the system does not depend solely on Alice's manual attention to
remain coherent. It monitors itself, identifies its own degradation signals, and initiates
corrective action.

### Property 1 — Self-Monitoring

The system continuously tracks its own health against defined baselines. Five degradation
signals are observable at the structural level (independent of operational performance):

| Structural Signal | Condition | Degradation Type |
|------------------|-----------|-----------------|
| S2 skip | `weekly-dump.md` blank after Monday noon | S2 failure — coordination broken |
| S3* drift accumulation | Same "slipped" item in End of Week Review for 3+ consecutive weeks | Chronic execution failure at that node |
| Revenue gap widening | Node 11 "what's coming in" decreasing while "what's close" does not convert | Pipeline degradation |
| Node silence | Any node's `signal.md` not updated for 2+ weeks without explicit PAUSED status | Active ghost node — draining S3 attention |
| Alice bottleneck index | Alice appears in "who's involved" for more than 7 of 12 nodes in a single dump | S3 bottleneck pathology threshold reached |

**Bottleneck index:**
```
If (Roberto_node_involvement > 7 out of 12)
  → S5 intervention required
  → Prompt: "Which of these 7+ can be delegated this week?"
```

### Property 2 — Self-Repair

Each degradation signal maps to a defined repair protocol that initiates without waiting
for manual intervention:

| Degradation Signal | System Repair Response |
|---|---|
| S2 skip (dump not done) | OSA agent prompts Monday dump at session start. No other work begins until current state is captured |
| Node chronic slip (3+ consecutive weeks) | Generate Node Diagnosis prompt: "What is the actual constraint? Resourcing? Ownership? Market?" |
| Revenue gap widening | Trigger S3 concentration protocol: all Top 3 non-negotiables become revenue-related until gap closes |
| Node silence | Alice prompted to explicitly mark node ACTIVE / PAUSED / KILLED. No implicit status. No ghost nodes. |
| Alice bottleneck above 7/12 | Generate delegation audit: for each node Alice is in, who SHOULD own this, and what is blocking the handoff? |

### Property 3 — Self-Regeneration

`miosa_memory.Cortex` auto-generates synthesis bulletins by detecting cross-session patterns.
These bulletins surface:
- Which nodes have been unresolved for 4+ weeks
- Which questions have recurred in weekly reviews
- Which signal fidelity patterns are persistent

Cortex output feeds the Monday brain dump, ensuring the system's own pattern detection is
surfaced at the moment of weekly planning — before the week plan commits resources.

Once per month, as part of the triple-loop governance review, the system audits its own
structural assumptions:

```
MONTHLY STRUCTURAL AUDIT:

Genre audit:
  Are all genres in the taxonomy still valid? (Any genre unused for 90+ days?)
  Are there new communication patterns that need a new genre?

Node audit:
  Are all nodes in topology.yaml accurately reflecting reality?
  Are any nodes marked active but producing no signals? (Ghost node detection)
  Are any new operations running without a topology entry?

Autonomy audit:
  Are all agent autonomy levels still appropriate?
  Have any agents consistently stayed within or exceeded their bounds?
  Are any bounds too restrictive (creating unnecessary escalation noise)?

Metric audit:
  Are the health metrics still the right metrics?
  Have any metrics been permanently at target (possibly measuring the wrong thing)?
  Are any thresholds miscalibrated?

Algedonic audit:
  Did any triggers fire this month? Were they handled appropriately?
  Are trigger thresholds still calibrated correctly?
  Are there new risk categories that need triggers?
```

### Property 4 — Self-Adaptation

Triple-loop feedback (L6) produces structural change proposals. L7 approves or rejects them.
This is the formal handoff from feedback to governance.

```
ADAPTATION FLOW:

L6 triple-loop produces:
  "The S4 function is consistently under-resourced. The current metric
   (node:op-new-stuff signal throughput) is not the right measure of strategic
   attention. Recommend: track Alice's S4 time directly."

L7 reviews:
  Is this a metric change (governance audit) or a structural change (VSM)?
  → Metric change: approved at L7; logged; miosa_context.Governance updated
  → Structural change (new S4 role, dedicated resource): requires S5 decision

L7 outputs:
  Approved structural changes → topology.yaml and governance.yaml updated.
  Rejected proposals → reason logged; L6 continues monitoring.
```

### Property 5 — Self-Description

This document IS autopoiesis. The system maintains its own formal specification. When the
architecture changes, the specification is updated as part of the change. A system that
cannot describe itself cannot govern itself.

The self-description protocol:
- Every ADR produced by `@architect` updates the relevant layer spec within the same work session
- Every structural change to `topology.yaml` triggers a review of the affected layer spec
- The layer specs are governance artifacts, not documentation — they are versioned, dated,
  and reviewed as part of the monthly structural audit
- An architecture doc that doesn't match operating reality is a degradation signal

---

## Viability Metrics

L7 monitors ten metrics as its primary health indicators. These measure the *structural health*
of the governance architecture itself — distinct from operational metrics (revenue, throughput,
delivery), which belong to S3.

| Metric | Definition | Target | Warning | Critical | Cadence |
|--------|-----------|--------|---------|----------|---------|
| Signal throughput | Signals processed per day, 7-day rolling average | > 80% of 4-week baseline | < 60% baseline | < 40% baseline | Daily |
| Feedback closure rate | % of signals with closed L6 single-loop confirmation | > 80% | < 60% | < 40% | Weekly |
| Knowledge freshness | % of active nodes updated within staleness threshold | 100% | < 85% | < 70% | Weekly |
| Governance coverage | % of active operations with explicit S1–S5 mapping | 100% | < 80% | < 60% | Monthly |
| Agent compliance | % of agent actions within their stated autonomy level | > 95% | < 90% | < 80% | Per-session |
| Decision trace completeness | % of significant decisions with logged rationale | > 90% | < 75% | < 60% | Weekly |
| S4 time allocation | Alice's actual time on environmental scanning vs. target | > 10% of work week | 5–10% | < 5% | Weekly |
| Algedonic response time | Time from trigger detection to S5 acknowledgment | < 24h for critical | 24–48h | > 48h | Per-event |
| Layer spec freshness | % of layer specs current with actual architecture | 100% | < 90% | < 80% | Monthly |
| Autonomy log completeness | % of autonomy level changes with logged reason | 100% | < 95% | < 85% | Per-change |

**Overall governance health score:**
- 8–10 metrics at target: System viable — continue current approach
- 5–7 metrics at target: Monitor — identify which systems are degraded and why
- 3–4 metrics at target: Alert — system viability at risk; S5 review required
- 0–2 metrics at target: Emergency — trigger algedonic review immediately

A governance review that only checks operational metrics (revenue, throughput, delivery)
and ignores these structural metrics is not a governance review — it is an operations
review wearing the wrong label.

---

## Monthly Governance Review

The monthly governance review runs on the last Friday of each month, after the triple-loop
L6 review. It is a separate session — 30–45 minutes, uninterrupted. It is not the same as
the triple-loop review. The triple-loop asks "are we asking the right questions?" The
governance review asks "is the system still structurally viable?"

### Template: `monthly-governance-review.md`

```markdown
# Monthly Governance Review — [Month YYYY]
Date: [date]
Conducted by: Alice (S5)

---

## S5 Check — Identity and Purpose

Current stated identity: [copy from alignment.md]

Questions:
1. Is the system still pursuing what it was built to pursue?
2. Have we made any commitments this month that implicitly change the identity?
3. Is there any operation running whose purpose contradicts the stated identity?
4. What would we NOT do, even if offered revenue? Is that boundary still clear?

S5 decision: [ ] No change needed  [ ] Identity update required: [specifics]

---

## S4 Check — Environmental Scanning

S4 time allocation this month (actual hours): ____
S4 time target (10% of work hours): ____
S4/S3 balance assessment: ____

External changes detected this month:
- [List significant market, competitive, or regulatory changes]

Opportunities identified (node:op-new-stuff outputs):
- [List]

Threats identified (including pre-launch: NVIDIA NeMo Claw, Pulsio, etc.):
- [List]

S4 decision: [ ] No structural action  [ ] S4 finding escalated to S5: [specifics]

---

## S3 Check — Control and Standards

Operations reviewed this month:

| Operation | Met Standards? | Resource Allocation Optimal? | Notes |
|-----------|---------------|------------------------------|-------|
| MIOSA Platform | Y/N | Y/N | |
| Lunivate Agency | Y/N | Y/N | |
| AI Masters | Y/N | Y/N | |
| Agency Accelerants | Y/N | Y/N | |
| OS Architect | Y/N | Y/N | |
| OS Accelerator | Y/N | Y/N | |
| Content Creators | Y/N | Y/N | |

Resource reallocation decisions made this month: [list]
Standards violations requiring structural response: [list]

S3* audit findings (gap between reported and actual): [list]

---

## S2 Check — Coordination Health

Cross-operation conflicts this month: [list]
Resource contention resolved: [list]
Coordination failures (operations that ran into each other without S2 handling it): [list]
Lateral coordination gaps (handoffs that still route through Alice): [list]

S2 decision: [ ] Mechanisms adequate  [ ] Coordination improvement needed: [specifics]

---

## S1 Check — Operations Throughput

Signal throughput (4-week average vs. prior 4-week period): ____% change
Operations below minimum viable throughput this month: [list]
Operations exceeding capacity this month: [list]
Ghost nodes detected (active label but no signals): [list]

---

## Autopoiesis Check — Structural Health

| Metric | This Month | Target | Status |
|--------|-----------|--------|--------|
| Governance coverage | ___% | 100% | Green/Yellow/Red |
| Feedback closure rate | ___% | >80% | Green/Yellow/Red |
| Knowledge freshness | ___% active nodes current | 100% | Green/Yellow/Red |
| Agent compliance | ___% | >95% | Green/Yellow/Red |
| Decision trace completeness | ___% | >90% | Green/Yellow/Red |
| S4 time allocation | ___% | >10% | Green/Yellow/Red |
| Layer spec freshness | ___% | 100% | Green/Yellow/Red |
| Autonomy log completeness | ___% | 100% | Green/Yellow/Red |

Overall governance health: [ ] Green  [ ] Yellow  [ ] Red

Algedonic events this month: [list — including dismissed events]

---

## Structural Decisions

Structural changes approved this month:
- [change, rationale, layer(s) affected, ADR if produced]

Structural changes rejected this month:
- [proposal, reason for rejection]

Layer specs updated this month: [list]
Layer specs requiring update next month: [list]
ADRs produced: [list]

---

## Next Month Forward Question (S5)

[One question about the system's identity or direction that will guide S4 scanning
and triple-loop framing for the next 30 days. Not a goal — a question.]
```

---

## Interface to Adjacent Layers

### FROM Layer 6 (Feedback) — Three Input Classes

| Input Class | What L6 Produces | How L7 Responds |
|-------------|-----------------|----------------|
| Triple-loop structural proposals | "This framework assumption may no longer be valid. Proposed structural change: X." | L7 reviews and approves or rejects. Approved changes → `topology.yaml` and layer specs updated. |
| Escalation triggers | "Double-loop question has recurred for 3+ consecutive weeks without resolution." | L7 classifies as governance issue, not strategy issue. Structural intervention initiated. |
| Alignment drift alerts | "Drift score below 8 this week." | L7 runs algedonic check against drift data. Persistent severe drift → algedonic signal. |

### TO All Inner Layers — Governance Wraps Everything

| Inner Layer | How L7 Constrains It |
|-------------|---------------------|
| **L1 — Network** | Algedonic bypass overrides all routing rules. Topology changes require governance coverage check before merge. New nodes require S1–S5 mapping. |
| **L2 — Signal** | Agent autonomy level determines which agents can initiate which signal act types. An L2 agent cannot issue a COMMIT signal on its own. |
| **L3 — Composition** | Genre additions or removals require monthly structural audit approval before becoming canonical. |
| **L4 — Interface** | Context retrieval tiers respect agent autonomy levels. An L2 agent cannot retrieve S4 or S5 context without explicit authorization. |
| **L5 — Data** | Decision traces and governance records are permanent. They cannot be deleted by inner layers. Only S5 can authorize archival. |
| **L6 — Feedback** | Triple-loop output requires L7 review before structural changes are implemented. L6 proposes; L7 decides. |

### Algedonic Channel — Unconditional Direct Access

Any node, at any layer, can inject an algedonic signal directly to L7. This channel is
not governed by routing rules, autonomy levels, or queue depth.

```
Source: any node in the system
Destination: miosa_context.Governance → node:endpoint-roberto
Condition: viability threat detected (any trigger category)
Bypass: all intermediate systems (S2, S3, S4)
Log: permanent record created regardless of outcome
```

---

## What Needs to Be Built

### Immediate (already defined in working documents)

1. **`## CRITICAL FLAGS` section in `weekly-dump.md`** — Five trigger types. Any active
   trigger must be filled before node-by-node review begins.

2. **S3* enforcement in `week-plan.md`** — Rename End of Week Review to "S3* AUDIT — required."
   Three-question format is mandatory: What was planned? What shipped? What is the gap?

3. **Node status protocol in weekly dump** — Every node explicitly marked ACTIVE / PAUSED / KILLED
   each week. No ghost nodes tolerated.

### Near-Term (Month 1)

4. **Lateral S2 coordination protocol** — Ivan (OS Architect) feeds Agency Accelerants funnel.
   MIOSA feeds all S1 units. These horizontal flows need an explicit handoff protocol that
   does not route through Alice.

5. **Developer capacity table in Node 10 `signal.md`** — Erin/Nina/Frank: weekly allocation
   showing who is committed where, how many hours, what is unbooked.

6. **Monthly S4 session** — 1-hour competitive landscape review on the last Friday of each month.
   Currently happening ad-hoc in Alice's head. Requires externalization into Node 09.

### Strategic (Quarter 1)

7. **`miosa_context.Governance` — Phase 1: Foundation**
   - `MiosaContext.Governance.AutonomyRegistry` — agent domain→level map, change log
   - `MiosaContext.Governance.DecisionTrace` — structured decision logging
   - `MiosaContext.Governance.HealthMetrics` — 10-metric dashboard with thresholds

8. **`miosa_context.Governance` — Phase 2: Algedonic**
   - `MiosaContext.Governance.Algedonic` — trigger definitions, threshold checks, bypass dispatch
   - Integration with `miosa_signal.FailureModes` — extend to system-level failure detection

9. **`miosa_context.Governance` — Phase 3: Autopoiesis**
   - `MiosaContext.Governance.StructuralAudit` — monthly audit runner with per-dimension checks
   - `MiosaContext.Governance.HealthDashboard` — real-time metric surfaces in brain dump context
   - Integration with `miosa_memory.Cortex` — surface governance anomalies in pattern detection

10. **`miosa_context.Governance` — Phase 4: VSM Enforcement**
    - `MiosaContext.Governance.VSMValidator` — validates `topology.yaml` against VSM completeness
    - `MiosaContext.Governance.S4Monitor` — tracks Alice's S4 time allocation
    - Monthly governance review template auto-population with live metrics

---

## Full System Map

```
┌─────────────────────────────────────────────────────────────────────────┐
│  S5 — POLICY (Alice's Identity)                                        │
│  "Every output is a Signal. Maximize S/N. Zero exceptions."             │
│  Signal Theory paper. No VC. CPU-at-edge. Self-funding.                │
│  Consortium model. Developer royalty framework. Oscar's mandate.       │
│  Artifact: 01-roberto/context.md + Signal Theory paper (Feb 2026)      │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  S4 — INTELLIGENCE (Outside/Future)                              │   │
│  │  Node 09: Fiova/NVIDIA/Pulsio watch. Tina 300K. Oscar/Peggy.     │   │
│  │  Capital Factory. Ivan clipper startup. GovCon pipeline.       │   │
│  │  Monthly: 1-hour competitive landscape session (gap to fill)    │   │
│  │  Artifact: 09-new-stuff/context.md + signal.md                  │   │
│  │                                                                  │   │
│  │  ┌────────────────────────────────────────────────────────────┐ │   │
│  │  │  S3 — CONTROL (Inside/Now)                                 │ │   │
│  │  │  week-plan.md: time blocks, top 3, team allocation         │ │   │
│  │  │  Node 11: revenue tracking, close pipeline, cash position  │ │   │
│  │  │  Node 10: developer allocation + capacity management        │ │   │
│  │  │                                                             │ │   │
│  │  │  S3* — AUDIT                                               │ │   │
│  │  │  week-plan.md End of Week Review (shipped vs planned)      │ │   │
│  │  │  Node signals vs actual deliverables                       │ │   │
│  │  │                                                             │ │   │
│  │  │  ┌──────────────────────────────────────────────────────┐  │ │   │
│  │  │  │  S2 — COORDINATION                                    │  │ │   │
│  │  │  │  weekly-dump.md (Monday batch capture)                │  │ │   │
│  │  │  │  */signal.md (standardized state per node)            │  │ │   │
│  │  │  │  Signal Theory M/G/T/F/W (communication standard)     │  │ │   │
│  │  │  │  weekly-template.md (structural protocol)             │  │ │   │
│  │  │  │                                                        │  │ │   │
│  │  │  │  ┌────────────────────────────────────────────────┐   │  │ │   │
│  │  │  │  │  S1 — OPERATIONS (Production Units)             │   │  │ │   │
│  │  │  │  │  02-MIOSA ←→ Erin, Nina, Frank             │   │  │ │   │
│  │  │  │  │  03-LUNIVATE ←→ Frank, Judy                   │   │  │ │   │
│  │  │  │  │  04-AI MASTERS ←→ Ed, Robert P., Sam           │   │  │ │   │
│  │  │  │  │  06-AGENCY ACCELERANTS ←→ Dan, Sam         │   │  │ │   │
│  │  │  │  │  05-OS ARCHITECT ←→ Ivan                      │   │  │ │   │
│  │  │  │  │  12-OS ACCELERATOR ←→ TBD                      │   │  │ │   │
│  │  │  │  └────────────────────────────────────────────────┘   │  │ │   │
│  │  │  └──────────────────────────────────────────────────────┘  │ │   │
│  │  └────────────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ALGEDONIC CHANNEL (bypasses all layers — fires directly to S5)         │
│  Revenue < $15K/mo | Team critical departure | NVIDIA NeMo Claw live   │
│  Strategic partner fracture | Client churn risk | Security breach       │
│  Implementation: ## CRITICAL FLAGS section at top of weekly-dump.md     │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## References

- Beer, S. — *Brain of the Firm* (2nd ed., 1981) — Viable System Model, algedonic channel, recursion
- Beer, S. — *The Heart of Enterprise* (1979) — S3/S4 balance, variety engineering
- Beer, S. — *Diagnosing the System for Organizations* (1985) — practical VSM application
- Maturana, H. & Varela, F. — *Autopoiesis and Cognition* (1980) — self-producing systems
- Ashby, W.R. — *An Introduction to Cybernetics* (1956) — requisite variety (governs L2, informed by L7)
- Wiener, N. — *Cybernetics* (1948) — feedback and control (governs L6, informs L7)
- Luna, Alice H. — *Signal Theory: The Architecture of Optimal Intent Encoding* (MIOSA Research, Feb 2026)
- [00-overview.md](00-overview.md) — Full 7-layer architecture
- [01-network.md](01-network.md) — Node types, VSM role assignments, topology
- [06-feedback.md](06-feedback.md) — Three-loop feedback, escalation path to L7
- Working documents: `weekly-dump.md`, `week-plan.md`, `weekly-review.md`, `monthly-review.md`, `alignment.md`
- Node contexts: `*/context.md` across all 12 OptimalOS nodes

---

*Layer 7 version 2.0 — 2026-03-17*
*Author: Architect Agent (OSA)*
*Status: Accepted*
