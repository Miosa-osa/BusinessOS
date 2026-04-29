# Layer 1: Network — Topology, Routing, and Channel Capacity

> L1 is the innermost layer. It defines who the nodes are, what connects them,
> and how signals travel between them. Every layer above this depends on the
> network being well-defined.

**Governing Constraint:** Shannon's channel capacity theorem.
Every channel has finite bandwidth. A signal routed to a receiver who cannot
decode it, or who is already at capacity, produces noise — not communication.

---

## 1. Purpose

Layer 1 answers three questions before any signal is sent:

1. **Who** — What nodes exist and what are their properties?
2. **How connected** — What channels link them?
3. **Where** — Given a signal, which node(s) should receive it?

Routing is not delivery. L1 determines the path. L4 (Interface) determines
the bandwidth-matched payload. L2 (Signal) classifies what is being sent.
L1 only concerns itself with topology.

---

## 2. Node Types

The Universal Taxonomy defines eight node types. Every actor in the Optimal
System is one of these.

| Type | Definition | Examples |
|------|-----------|---------|
| **Entity** | A legal or economic unit with independent liability and identity | MIOSA LLC, Lunivate LLC |
| **Domain** | A bounded operational area with coherent purpose and ownership | Platform, Agency, Education, Content, Research |
| **Operation** | A time-bounded project or program with a goal, people, and outputs | ClinicIQ, AI Masters, OS Accelerator |
| **Unit** | A team or group within an operation, with shared function | Engineering team, Sales team, Content team |
| **Endpoint** | An individual human with a name, role, and bandwidth profile | Alice, Ed, Dan, Ivan |
| **Agent** | An autonomous AI actor with a defined autonomy level (L1–L5) | OSA Agent, @architect, @debugger |
| **Relay** | A node that re-encodes and forwards signals (manager, coordinator) | Oscar (political relay), Alice (when acting as S3 coordinator) |
| **Bridge** | A node that connects two domains without belonging to either | Carol (infrastructure ↔ platform), Tina Motley (audience ↔ product) |

### Node Type Rules

- A node can be classified as multiple types only if each type applies at a
  different scope level. Alice is an Endpoint (individual human) at the
  network layer AND a Relay when acting as coordinator between operations.
- Agents carry an autonomy level (L1–L5) as a required property.
  See [Layer 7: Governance](07-governance.md) for autonomy definitions.
- Relay nodes introduce distortion risk. Every relay in a signal path is
  a potential fidelity loss point (see Section 5: Relay Distortion).
- Bridge nodes span domain boundaries. They require genre competence in
  both domains they connect.

---

## 3. Node Properties

Every node in the network has the following required properties.

```yaml
# Node Schema
id: string                    # Unique identifier: node:<slug>
name: string                  # Human-readable name
type: entity | domain | operation | unit | endpoint | agent | relay | bridge
status: active | paused | archived

# Communication properties
genre_competence:             # Which signal genres this node can decode
  - genre_name: string        # e.g. spec, brief, adr, report, note
    level: fluent | proficient | basic
bandwidth:
  tokens_per_session: integer # Approximate token budget per interaction
  signals_per_week: integer   # How many incoming signals this node can process
  preferred_mode: linguistic | visual | code | data

# Topology
channels:                     # Communication pathways available
  - type: async | sync | broadcast
    medium: string            # slack, email, github, voice, in-person
member_of:                    # Parent nodes (one or more)
  - node_id: string
connections:                  # Peer relationships (bidirectional)
  - node_id: string
    type: reports_to | collaborates_with | bridges_to | contracts_with | partners_with
    weight: float             # Signal flow frequency (0.0–1.0)
```

### Required vs Optional Properties

| Property | Required | Notes |
|----------|----------|-------|
| id, name, type, status | Yes | Core identity |
| genre_competence | Yes | Routing depends on this |
| bandwidth | Yes | Capacity matching depends on this |
| channels | Yes | At least one channel required |
| member_of | Yes for non-root nodes | Root nodes: Entity type |
| connections | No | Populated as relationships form |

---

## 4. Alice's Node Topology

The 12 OptimalOS folders map to the following node hierarchy.

### 4.1 Entity Layer (Root Nodes)

```
node:miosa-llc        Entity — MIOSA LLC (Texas, formed Jan 31 2026)
node:lunivate-llc     Entity — Lunivate LLC (Michigan, agency services)
```

These are the two root nodes. All other nodes are transitively contained
in one or both entities.

### 4.2 Domain Layer

```
MIOSA LLC
├── node:domain-platform        Domain — Platform (Node 02)
├── node:domain-education       Domain — Education (Nodes 04, 06, 12)
└── node:domain-research        Domain — Research (Node 09)

Lunivate LLC
├── node:domain-agency          Domain — Agency (Node 03)
└── node:domain-content         Domain — Content (Nodes 05, 07, 08)

Cross-cutting (belongs to both entities functionally)
├── node:endpoint-roberto       Endpoint — Alice (Node 01)
├── node:unit-team              Unit — Team (Node 10)
└── node:domain-finance         Domain — Finance (Node 11)
```

### 4.3 Full Node Table

| Node | ID | Type | Entity | VSM Role | Status |
|------|----|------|--------|----------|--------|
| 01 – Alice | node:endpoint-roberto | Endpoint + Relay | Both | S3/S4/S5 arbiter | active |
| 02 – MIOSA Platform | node:op-miosa | Operation | MIOSA LLC | S1 production | active |
| 03 – Lunivate Agency | node:op-lunivate | Operation | Lunivate LLC | S1 production | active |
| 04 – AI Masters | node:op-ai-masters | Operation | MIOSA LLC | S1 production | active |
| 05 – OS Architect | node:op-os-architect | Operation | Lunivate LLC | S1 production | active |
| 06 – Agency Accelerants | node:op-agency-acc | Operation | Lunivate LLC | S1 production | active |
| 07 – Accelerants Community | node:op-acc-community | Operation | Lunivate LLC | S1 sub-op | active |
| 08 – Content Creators | node:op-content-creators | Operation | Lunivate LLC | S1 sub-op | active |
| 09 – New Stuff | node:op-new-stuff | Operation | Both | S4 intelligence | active |
| 10 – Team | node:unit-team | Unit | Both | S2 coordination | active |
| 11 – Money / Revenue | node:domain-finance | Domain | Both | S3 control | active |
| 12 – OS Accelerator | node:op-os-accelerator | Operation | MIOSA LLC | S1 production | active |

### 4.4 Key Endpoint Nodes (Individual Humans)

| Person | ID | Role | Member Of | Bandwidth (signals/wk) |
|--------|----|------|-----------|----------------------|
| Alice | node:endpoint-roberto | CEO / Architect / S5 | All operations | 60–80 (overloaded) |
| Alice | node:endpoint-ed | Educator / Developer | node:op-ai-masters | 20 |
| Bob | node:endpoint-robert-p | Sales / Agency | node:op-ai-masters | 15 |
| Dan | node:endpoint-bennett | Operations Lead | node:op-agency-acc | 20 |
| Ivan | node:endpoint-ahmed | Content Lead | node:op-os-architect | 20 |
| Carol | node:endpoint-pedram | Infrastructure Partner | node:op-miosa | 15 |
| Oscar | node:endpoint-jordan | Consortium Relay | node:op-miosa, node:op-lunivate | 10 |
| Sam | node:endpoint-len | Sales | node:op-ai-masters, node:op-agency-acc | 20 |
| Erin | node:endpoint-pedro | Developer | node:unit-team | 30 |
| Nina | node:endpoint-javaris | Developer | node:unit-team | 30 |
| Frank | node:endpoint-nejd | Integrations | node:unit-team | 25 |
| Judy | node:endpoint-tejas | Content Ops | node:unit-team | 20 |

### 4.5 Agent Nodes

| Agent | ID | Autonomy | Domain | Reports To |
|-------|----|----------|--------|------------|
| OSA Master Orchestrator | node:agent-osa | L3 | All | node:endpoint-roberto |
| @architect | node:agent-architect | L3 | System design | node:agent-osa |
| @debugger | node:agent-debugger | L4 | Bug fixing | node:agent-osa |
| @code-reviewer | node:agent-reviewer | L4 | Code quality | node:agent-osa |
| @security-auditor | node:agent-security | L3 | Security | node:agent-osa |
| @backend-elixir | node:agent-elixir | L3 | Elixir/Phoenix | node:agent-osa |
| @backend-go | node:agent-go | L3 | Go services | node:agent-osa |
| @frontend-svelte | node:agent-svelte | L3 | Frontend | node:agent-osa |
| @frontend-react | node:agent-react | L3 | Frontend | node:agent-osa |
| @devops-engineer | node:agent-devops | L3 | Infrastructure | node:agent-osa |
| @database-specialist | node:agent-db | L2 | Database | node:agent-osa |
| @oracle | node:agent-oracle | L3 | AI/ML | node:agent-osa |

---

## 5. Routing Rules

All signal routing follows these rules in order. A later rule only applies
when an earlier rule does not resolve the routing decision.

### Rule 1 — Genre Match (Required)

A signal must only be routed to a node whose `genre_competence` includes
the signal's genre at the required level.

```
signal.genre = "spec"
→ destination.genre_competence must include "spec" at level fluent or proficient
→ if no match: re-encode to a genre the destination can decode (L3 Composition)
→ if re-encoding is impossible: route to a Bridge node for translation
```

Genre routing violations produce **Genre Mismatch** failure (Ashby violation).
A spec sent to a salesperson is noise. A brief sent to an engineer is noise.

### Rule 2 — Bandwidth Match (Required)

A signal must not exceed the destination's `bandwidth.tokens_per_session`
or saturate its `signals_per_week` capacity.

```
if signal.estimated_tokens > destination.bandwidth.tokens_per_session:
  → apply L4 tiered disclosure (L0 first, drill on request)
  → never send L2 payload to a node that needs L0

if destination.signals_per_week is at capacity:
  → batch or defer the signal
  → never add to an overloaded queue without priority triage
```

Bandwidth violations produce **Bandwidth Overload** failure (Shannon violation).

### Rule 3 — Prefer Direct Paths (Minimize Relay Distortion)

Every relay in a signal path adds encoding/decoding cost and introduces
fidelity risk. Prefer the shortest path that satisfies Rules 1 and 2.

```
Score each path = number_of_relays × 0.15 distortion_coefficient
Lowest score wins.
Path with zero relays = score 0.0 (direct channel preferred).
```

Exception: when a relay's genre_competence or domain translation ability
is required to satisfy Rule 1, include the relay despite the cost.

### Rule 4 — Relay Fidelity Check (Distortion Prevention)

When a relay is in the path, it must re-encode faithfully. The relay's
output must preserve:

- The original signal's intent
- The original signal's act type (direct/inform/commit/decide/express)
- All commitment-critical content (dates, amounts, decisions)

A relay that transforms a `commit` act into an `inform` act has introduced
distortion. This is the most common relay failure in Alice's current system:
decisions made by Alice are re-encoded by team relays as "suggestions," losing
their binding commitment status.

**Distortion detection:** Compare the output signal of the relay against the
input signal using the S=(M,G,T,F,W) dimensions. Any change in T (Type/act)
that was not explicitly authorized is distortion.

### Rule 5 — Algedonic Bypass (Emergency Override)

When an algedonic condition is active (see [Layer 7: Governance](07-governance.md)),
normal routing rules are suspended for the affected signal. The signal routes
directly to `node:endpoint-roberto` regardless of bandwidth state or queue depth.

```
Algedonic triggers (from VSM mapping):
- Recurring revenue confirmed < $15K/month
- Key team member (Erin, Nina, Frank, Dan) departs
- NVIDIA Nemo Claw goes live
- Carol or Oscar relationship fractures
- ClinicIQ or Mosaic Effect signals termination risk

Bypass route: any_node → node:endpoint-roberto [ALGEDONIC]
Priority: overrides all queues
```

---

## 6. Relay Distortion Prevention

Relay distortion is the primary failure mode at L1. It occurs when a relay
node does not re-encode the signal faithfully — it either loses content,
changes the act type, filters without authorization, or adds noise.

### 6.1 Distortion Types

| Distortion Type | What Happens | Example |
|----------------|-------------|---------|
| **Act corruption** | `commit` re-encoded as `inform` | Decision announced as "FYI" instead of binding |
| **Content loss** | Signal shortened beyond L4 threshold | 5-item commitment summarized as 1-item, 4 lost |
| **Noise injection** | Relay adds own interpretation as fact | Coordinator adds "Alice thinks..." (fabricated) |
| **Genre shift** | Spec re-encoded as casual note | Technical requirement turned into a Slack message |
| **Bandwidth mismatch** | Full payload sent when summary was appropriate | Full ADR forwarded to a sales endpoint |

### 6.2 Fidelity Protocol for Relay Nodes

Any node operating as a relay (Alice coordinating between S1 units, Oscar
as consortium relay, etc.) must follow this protocol:

```
1. RECEIVE: Accept the signal in its original form.
2. CLASSIFY: Identify M, G, T, F, W of the incoming signal.
3. RE-ENCODE: Translate genre/mode to match destination's competence.
   - Preserve: intent, act type, commitment content, deadlines, amounts.
   - Translate: genre, mode, format as needed.
4. VERIFY: Does the re-encoded signal preserve act type T?
   If T changed, flag explicitly: "Original was a COMMIT; this re-encoding
   is for clarity, original commitment still binding."
5. FORWARD: Route to destination.
```

### 6.3 Alice as Relay (S3 Coordinator)

Alice is the most critical relay in the network and the highest-distortion
risk because he is:
- Overloaded (bandwidth at or above capacity most weeks)
- The single path between most S1 units and their strategic context
- Subject to the god-complex bottleneck pattern (Oscar's mandate)

Distortion risk mitigation:
- Alice's relay function should decrease over time as S1 unit leads
  gain direct communication channels with each other.
- The target state is: Alice relays S5 policy (identity signals),
  NOT S1 operational signals.
- Any operational signal that has been in Alice's relay queue for
  more than 48 hours should be flagged for direct routing.

---

## 7. Data Model — topology.yaml

The network topology is stored as a YAML file and read by `miosa_knowledge`
for SPARQL query and OWL reasoning.

```yaml
# /OptimalOS/config/topology.yaml
# Optimal System — Network Topology Definition
# Version: 1.0 | Date: 2026-03-16

version: "1.0"
schema: "optimal-network-v1"

entities:
  - id: node:miosa-llc
    name: "MIOSA LLC"
    type: entity
    status: active
    legal:
      state: TX
      formed: "2026-01-31"
    genre_competence:
      - genre: spec
        level: fluent
      - genre: adr
        level: fluent
    bandwidth:
      tokens_per_session: 200000
      signals_per_week: 200

  - id: node:lunivate-llc
    name: "Lunivate LLC"
    type: entity
    status: active
    legal:
      state: MI
    genre_competence:
      - genre: brief
        level: fluent
      - genre: report
        level: fluent
    bandwidth:
      tokens_per_session: 100000
      signals_per_week: 100

domains:
  - id: node:domain-platform
    name: "Platform Domain"
    type: domain
    status: active
    member_of: [node:miosa-llc]
    genre_competence:
      - genre: spec
        level: fluent
      - genre: adr
        level: fluent
    bandwidth:
      tokens_per_session: 100000
      signals_per_week: 50

  - id: node:domain-education
    name: "Education Domain"
    type: domain
    status: active
    member_of: [node:miosa-llc]
    genre_competence:
      - genre: brief
        level: fluent
      - genre: guide
        level: fluent
    bandwidth:
      tokens_per_session: 50000
      signals_per_week: 40

  - id: node:domain-research
    name: "Research Domain"
    type: domain
    status: active
    member_of: [node:miosa-llc]
    genre_competence:
      - genre: report
        level: fluent
      - genre: spec
        level: proficient
    bandwidth:
      tokens_per_session: 80000
      signals_per_week: 20

  - id: node:domain-agency
    name: "Agency Domain"
    type: domain
    status: active
    member_of: [node:lunivate-llc]
    genre_competence:
      - genre: brief
        level: fluent
      - genre: report
        level: proficient
    bandwidth:
      tokens_per_session: 60000
      signals_per_week: 60

  - id: node:domain-content
    name: "Content Domain"
    type: domain
    status: active
    member_of: [node:lunivate-llc]
    genre_competence:
      - genre: brief
        level: fluent
      - genre: script
        level: fluent
    bandwidth:
      tokens_per_session: 30000
      signals_per_week: 40

  - id: node:domain-finance
    name: "Finance Domain"
    type: domain
    status: active
    member_of: [node:miosa-llc, node:lunivate-llc]
    genre_competence:
      - genre: report
        level: fluent
    bandwidth:
      tokens_per_session: 20000
      signals_per_week: 10

operations:
  - id: node:op-miosa
    name: "MIOSA Platform"
    folder: "nodes/02-miosa"
    type: operation
    status: active
    member_of: [node:domain-platform]
    vsm_role: S1
    genre_competence:
      - genre: spec
        level: fluent
      - genre: adr
        level: fluent
      - genre: report
        level: proficient
    bandwidth:
      tokens_per_session: 100000
      signals_per_week: 50
    channels:
      - type: async
        medium: github
      - type: async
        medium: slack
      - type: sync
        medium: voice
    connections:
      - node_id: node:unit-team
        type: employs
        weight: 0.9
      - node_id: node:endpoint-pedram
        type: partners_with
        weight: 0.8

  - id: node:op-lunivate
    name: "Lunivate Agency"
    folder: "nodes/03-lunivate"
    type: operation
    status: active
    member_of: [node:domain-agency]
    vsm_role: S1
    genre_competence:
      - genre: brief
        level: fluent
      - genre: report
        level: fluent
    bandwidth:
      tokens_per_session: 60000
      signals_per_week: 40
    channels:
      - type: async
        medium: slack
      - type: sync
        medium: voice
    connections:
      - node_id: node:endpoint-nejd
        type: employs
        weight: 0.7
      - node_id: node:endpoint-tejas
        type: employs
        weight: 0.6

  - id: node:op-ai-masters
    name: "AI Masters"
    folder: "nodes/04-ai-masters"
    type: operation
    status: active
    member_of: [node:domain-education]
    vsm_role: S1
    genre_competence:
      - genre: brief
        level: fluent
      - genre: guide
        level: proficient
    bandwidth:
      tokens_per_session: 50000
      signals_per_week: 30
    channels:
      - type: async
        medium: slack
      - type: broadcast
        medium: email
      - type: sync
        medium: voice

  - id: node:op-os-architect
    name: "OS Architect"
    folder: "nodes/05-os-architect"
    type: operation
    status: active
    member_of: [node:domain-content]
    vsm_role: S1
    genre_competence:
      - genre: script
        level: fluent
      - genre: brief
        level: proficient
    bandwidth:
      tokens_per_session: 30000
      signals_per_week: 20

  - id: node:op-agency-acc
    name: "Agency Accelerants"
    folder: "nodes/06-agency-accelerants"
    type: operation
    status: active
    member_of: [node:domain-education]
    vsm_role: S1
    genre_competence:
      - genre: brief
        level: fluent
    bandwidth:
      tokens_per_session: 40000
      signals_per_week: 30

  - id: node:op-acc-community
    name: "Accelerants Community"
    folder: "nodes/07-accelerants-community"
    type: operation
    status: active
    member_of: [node:op-agency-acc]
    vsm_role: S1-sub
    genre_competence:
      - genre: brief
        level: proficient
    bandwidth:
      tokens_per_session: 20000
      signals_per_week: 20

  - id: node:op-content-creators
    name: "Content Creators"
    folder: "nodes/08-content-creators"
    type: operation
    status: active
    member_of: [node:domain-content]
    vsm_role: S1-sub
    genre_competence:
      - genre: brief
        level: fluent
      - genre: script
        level: proficient
    bandwidth:
      tokens_per_session: 20000
      signals_per_week: 20

  - id: node:op-new-stuff
    name: "New Stuff / Research"
    folder: "nodes/09-new-stuff"
    type: operation
    status: active
    member_of: [node:domain-research]
    vsm_role: S4
    genre_competence:
      - genre: report
        level: fluent
      - genre: brief
        level: proficient
    bandwidth:
      tokens_per_session: 80000
      signals_per_week: 15

  - id: node:op-os-accelerator
    name: "OS Accelerator"
    folder: "nodes/12-os-accelerator"
    type: operation
    status: active
    member_of: [node:domain-education]
    vsm_role: S1
    genre_competence:
      - genre: guide
        level: proficient
    bandwidth:
      tokens_per_session: 40000
      signals_per_week: 20

units:
  - id: node:unit-team
    name: "Team"
    folder: "nodes/10-team"
    type: unit
    status: active
    member_of: [node:op-miosa, node:op-lunivate]
    vsm_role: S2
    genre_competence:
      - genre: spec
        level: proficient
      - genre: report
        level: proficient
    bandwidth:
      tokens_per_session: 60000
      signals_per_week: 80

endpoints:
  - id: node:endpoint-roberto
    name: "Alice"
    type: endpoint
    status: active
    roles: [relay, bridge]
    member_of: [node:miosa-llc, node:lunivate-llc]
    genre_competence:
      - genre: spec
        level: fluent
      - genre: adr
        level: fluent
      - genre: brief
        level: fluent
      - genre: report
        level: fluent
      - genre: script
        level: fluent
    bandwidth:
      tokens_per_session: 200000
      signals_per_week: 70
      preferred_mode: linguistic
    channels:
      - type: async
        medium: slack
      - type: async
        medium: github
      - type: sync
        medium: voice
      - type: sync
        medium: in-person

  - id: node:endpoint-pedram
    name: "Carol"
    type: bridge
    status: active
    member_of: [node:op-miosa]
    bridges: [node:domain-platform, infrastructure]
    genre_competence:
      - genre: spec
        level: fluent
    bandwidth:
      tokens_per_session: 40000
      signals_per_week: 15
    channels:
      - type: sync
        medium: voice
      - type: async
        medium: slack

  - id: node:endpoint-jordan
    name: "Oscar"
    type: relay
    status: active
    bridges: [node:op-miosa, consortium]
    genre_competence:
      - genre: brief
        level: fluent
      - genre: report
        level: fluent
    bandwidth:
      tokens_per_session: 30000
      signals_per_week: 10
    channels:
      - type: sync
        medium: voice
      - type: sync
        medium: in-person
```

---

## 8. SPARQL Topology Queries (miosa_knowledge)

The topology.yaml is loaded into `miosa_knowledge` as RDF triples. These are
the canonical queries for routing and traversal.

### Find all endpoints in an operation

```sparql
SELECT ?endpoint ?name
WHERE {
  ?endpoint rdf:type optimal:Endpoint .
  ?endpoint optimal:member_of+ node:op-ai-masters .
  ?endpoint optimal:name ?name .
}
```

### Resolve a routing path from source to destination

```sparql
SELECT ?path ?hops ?distortion_score
WHERE {
  BIND(node:endpoint-roberto AS ?source)
  BIND(node:endpoint-bennett AS ?dest)

  ?path optimal:connects ?source TO ?dest .
  ?path optimal:hop_count ?hops .
  BIND(?hops * 0.15 AS ?distortion_score)
}
ORDER BY ASC(?distortion_score)
LIMIT 1
```

### Find all nodes at or above bandwidth capacity

```sparql
SELECT ?node ?current_load ?max_capacity
WHERE {
  ?node optimal:current_signal_load ?current_load .
  ?node optimal:bandwidth_signals_per_week ?max_capacity .
  FILTER (?current_load >= (?max_capacity * 0.85))
}
```

### Transitive membership (OWL-inferred)

```sparql
# If X member_of Y and Y member_of Z, then X transitively_in Z
SELECT ?person ?top_entity
WHERE {
  ?person rdf:type optimal:Endpoint .
  ?person optimal:transitively_in ?top_entity .
  ?top_entity rdf:type optimal:Entity .
}
```

### Find all relay nodes in the network

```sparql
SELECT ?relay ?relay_name
WHERE {
  { ?relay rdf:type optimal:Relay }
  UNION
  { ?relay optimal:roles optimal:relay }
  ?relay optimal:name ?relay_name .
  ?relay optimal:status "active" .
}
```

---

## 9. Interface to Layer 2

L1 provides routing. L2 provides signal classification. The interface
between them is a single function:

```
route(signal, source_node) → destination_nodes[]
```

### Contract

**Input:**
- `signal` — a partially-formed signal with at minimum: `genre`, `act_type`,
  `estimated_tokens`, `source_node_id`
- `source_node` — the originating node in the topology

**Output:**
- `destination_nodes[]` — ordered list of destination nodes, from most-preferred
  to least-preferred, based on the five routing rules in Section 5
- Each destination includes: `node_id`, `channel`, `distortion_score`,
  `bandwidth_available`, `requires_re_encoding: bool`

**Guarantees from L1 to L2:**
1. Every destination in the returned list has verified genre_competence for
   the signal's genre.
2. Every destination has bandwidth headroom to receive the signal.
3. The path with minimum relay distortion is ranked first.
4. If an algedonic condition is active, `node:endpoint-roberto` is the only
   destination returned, regardless of queue state.

**What L1 does NOT do:**
- L1 does not classify the signal (that is L2's job).
- L1 does not format the payload (that is L3/L4's job).
- L1 does not store the signal (that is L5's job).
- L1 does not verify delivery or close the feedback loop (that is L6's job).

### Routing Decision Record

Every routing decision is logged in L5 (Data) as a `routing_event`:

```json
{
  "id": "route_<ulid>",
  "timestamp": "2026-03-16T10:00:00Z",
  "source": "node:endpoint-roberto",
  "signal_genre": "brief",
  "signal_act_type": "direct",
  "candidates_evaluated": 3,
  "selected_destination": "node:endpoint-bennett",
  "channel": "slack/async",
  "distortion_score": 0.0,
  "routing_rule_applied": "rule_1_genre_match",
  "algedonic_bypass": false
}
```

---

## 10. Known Network Pathologies (March 2026)

These are current distortion and capacity issues in Alice's network,
recorded for L6 (Feedback) to track resolution.

| Pathology | Type | Affected Nodes | Root Cause | Resolution |
|-----------|------|----------------|------------|------------|
| Alice bottleneck | Bandwidth overload | node:endpoint-roberto | 70+ signals/week through a single relay node | Increase S1 unit lead autonomy; reduce relay dependency |
| S1 lateral silence | Missing channel | node:op-miosa ↔ node:op-os-architect | No direct channel between Ivan and Erin; both route through Alice | Create direct Slack channel; formalize handoff protocol |
| Developer over-commitment | Capacity saturation | node:unit-team | Erin, Nina, Frank allocated to multiple operations simultaneously without capacity booking | Weekly capacity table in Node 10 signal.md |
| Act corruption on relay | Relay distortion | Multiple operations | Operational decisions re-encoded as suggestions when Alice is not direct sender | All Alice decisions marked COMMIT explicitly; not re-encodable as inform |
| Ghost node status | Missing routing data | node:op-new-stuff, node:op-os-accelerator | Nodes exist in topology but no active weekly signals; routing cannot verify capacity | Enforce explicit ACTIVE / PAUSED / KILLED on all nodes in weekly dump |

---

## 11. Topology Diagram

```
ENTITY LAYER
┌─────────────────────────────────────────────────────────────────────┐
│  MIOSA LLC                          LUNIVATE LLC                    │
│  ┌─────────────────────────┐        ┌──────────────────────────┐   │
│  │ Domain: Platform        │        │ Domain: Agency           │   │
│  │  └── Op: MIOSA [02]     │        │  └── Op: Lunivate [03]   │   │
│  │                         │        │                          │   │
│  │ Domain: Education       │        │ Domain: Content          │   │
│  │  ├── Op: AI Masters[04] │        │  ├── Op: OS Arch  [05]   │   │
│  │  ├── Op: Agency Acc[06] │        │  ├── Op: Acc Comm [07]   │   │
│  │  │    └── Op: AccComm   │        │  └── Op: Content  [08]   │   │
│  │  └── Op: OS Acc   [12]  │        │                          │   │
│  │                         │        └──────────────────────────┘   │
│  │ Domain: Research        │                                        │
│  │  └── Op: NewStuff [09]  │                                        │
│  └─────────────────────────┘                                        │
│                                                                     │
│  CROSS-CUTTING (Both Entities)                                      │
│  ├── Endpoint: Alice [01]  ← S3/S4/S5 arbiter, Relay            │
│  ├── Unit: Team       [10]  ← S2 coordination                     │
│  └── Domain: Finance  [11]  ← S3 revenue control                  │
└─────────────────────────────────────────────────────────────────────┘

RELAY + BRIDGE LAYER (external connections)
┌──────────────────────────────────────────────────────────────────┐
│  Carol    → Bridge: infrastructure ↔ node:op-miosa             │
│  Oscar    → Relay:  consortium ↔ node:endpoint-roberto         │
│  Tina M.   → Bridge: 300K audience ↔ Education domain           │
│  Consortium AI → node:op-miosa (strategic + compute layer)      │
└──────────────────────────────────────────────────────────────────┘

AGENT LAYER (autonomous actors)
┌──────────────────────────────────────────────────────────────────┐
│  OSA [L3] → @architect[L3] → @debugger[L4]                      │
│          → @reviewer[L4]   → @security[L3]                      │
│          → @elixir[L3]     → @go[L3]                            │
│          → @svelte[L3]     → @react[L3]                         │
│          → @devops[L3]     → @db[L2]                            │
│          → @oracle[L3]                                           │
└──────────────────────────────────────────────────────────────────┘
```

---

## 12. Related Documents

- [00-overview.md](00-overview.md) — Full 7-layer architecture
- [02-signal.md](02-signal.md) — Signal classification (consumer of L1 routing)
- [07-governance.md](07-governance.md) — VSM mapping, agent autonomy, algedonic channel
- [../taxonomy/hierarchy.md](../taxonomy/hierarchy.md) — Node type taxonomy definitions
- [../operations/auto-routing.md](../operations/auto-routing.md) — Runtime routing implementation
- [ADR-001](../../tasks/ADR-001-feedback-loop-architecture.md) — VSM → OptimalOS mapping

---

*Layer 1 version 1.0 — 2026-03-16*
*Author: Architect Agent (OSA)*
*Status: Accepted*
