# Workspace — the Organizational Topology Layer

> A tenant's **workspace** is the first-class model of the company it represents.
> Nodes are the map. People and agents are the first-class principals inside
> that map. Skills are what principals can do. Integrations are the tools
> (and the translation layer between the human and agent layers).
>
> Before this layer: `contexts.node` was a freeform string. Each ingest picked
> whatever node slug it felt like. Routing rules pointed at strings. There was
> no canonical registry of "what nodes exist in this company, who owns them,
> who has access, what are they about."
>
> After this layer: every node is a row with a versioned identity, a kind
> (team, project, entity, domain, person), a membership roster (internal +
> external), and a skill-labeled workforce (humans + agents).

---

## 1. The two parallel layers

A tenant has one workspace. The workspace has two parallel surfaces:

```
         HUMAN LAYER                        AGENT LAYER
         ───────────                        ───────────
    Person (principal.kind=:user)      Agent (principal.kind=:agent)
    ├─ classifications (groups, roles) ├─ skills (what it can do)
    ├─ skills (what they know)         ├─ autonomy level
    └─ memberships (which nodes)       └─ tool grants (which integrations)
                  │                                  │
                  └──────────────┬───────────────────┘
                                 ▼
                      TOOLS / INTEGRATIONS
                 (Slack / Gmail / Drive / Jira / MCP / ...)
                   the translation layer between them
```

A principal's **kind** (`:user | :agent | :service`) already distinguishes the
two layers. This doc formalizes everything else — the nodes they work inside,
the skills they bring, the tools they wield.

---

## 2. The five primitives

### 2.1 Node

An organizational unit within the tenant. Nodes form a tree:

```
Tenant "acme-corp"
├── "engineering"          (kind: unit)
│   ├── "platform-team"   (kind: team)
│   └── "data-team"       (kind: team)
├── "sales"                (kind: unit)
│   ├── "enterprise-sales" (kind: team)
│   └── "amer-pod"        (kind: team)
├── "accounts"             (kind: unit — houses external entities)
│   ├── "acme-supplier-a"  (kind: entity, style: external)
│   └── "acme-customer-b"  (kind: entity, style: external)
└── "projects"             (kind: unit)
    ├── "q2-launch"        (kind: project)
    └── "compliance-soc2"  (kind: project)
```

**Kinds** (`node.kind`): `:unit | :team | :project | :entity | :domain | :person`
**Styles** (`node.style`): `:internal | :external | :mixed`
**Statuses** (`node.status`): `:active | :archived | :draft`

### 2.2 NodeMember

Ties a principal to a node with an explicit membership type.

| membership   | meaning                                                                           |
|--------------|-----------------------------------------------------------------------------------|
| `:owner`     | operates the node; can add members, edit node metadata                            |
| `:internal`  | employee / contributor inside the company working on the node                     |
| `:external`  | client, partner, vendor — outside the company, participates in the node           |
| `:observer`  | read-only visibility; often an auditor, cross-team onlooker, or AI agent          |

A principal can be a member of many nodes; a node has many members.

### 2.3 Skill

A named capability, tenant-scoped. Examples: `"elixir"`, `"go-to-market"`,
`"phoenix-arbiter-model"`, `"enterprise-sales"`, `"sql-optimization"`.

`skill.kind` classifies the capability: `:technical | :communication | :strategic | :domain | :tool`.

### 2.4 PrincipalSkill

A principal (human OR agent) "has" a skill at some level.

| level           | meaning                                                   |
|-----------------|-----------------------------------------------------------|
| `:novice`       | familiar; can ask good questions                          |
| `:intermediate` | gets work done; needs review on edge cases                |
| `:expert`       | self-directed; can train juniors                          |
| `:lead`         | sets direction for the skill across a team or tenant      |

Same shape for humans and agents — the engine doesn't care which kind a
skill-holder is. "Slack integration expert" could be a Principal of kind
`:user` (a human who writes the integration) or kind `:agent` (an AI that
operates it). Both route identically for capability lookups.

### 2.5 Integration (Phase 9 territory, referenced here)

Tool available to the workspace. `connectors` table from Phase 1 is the
implementation surface for ingestion-side integrations (Slack / Gmail /
Drive / …). Phase 9 will extend this with **`integration_grants`** —
principal↔integration membership so the engine can answer
"what tools does this agent have access to?"

Not built in Phase 3.5. Called out here so the model is coherent.

---

## 3. Schema (Phase 3.5 migrations 017–020)

```sql
-- 017: nodes
CREATE TABLE nodes (
  id                TEXT PRIMARY KEY,                 -- "tenant-id:node-slug" or UUID
  tenant_id         TEXT NOT NULL,
  slug              TEXT NOT NULL,                    -- "engineering", "04-ai-masters", etc.
  name              TEXT NOT NULL,                    -- display name
  kind              TEXT NOT NULL,                    -- unit | team | project | entity | domain | person
  parent_id         TEXT REFERENCES nodes(id) ON DELETE CASCADE,
  description       TEXT,
  style             TEXT NOT NULL DEFAULT 'internal', -- internal | external | mixed
  status            TEXT NOT NULL DEFAULT 'active',   -- active | archived | draft
  path              TEXT NOT NULL DEFAULT '',         -- filesystem path (e.g. "nodes/04-ai-masters")
  metadata          TEXT NOT NULL DEFAULT '{}',
  created_at        TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(tenant_id, slug)
);

-- 018: node_members
CREATE TABLE node_members (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id         TEXT NOT NULL,
  node_id           TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  principal_id      TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
  membership        TEXT NOT NULL DEFAULT 'internal', -- owner | internal | external | observer
  role              TEXT,                             -- optional role within the node
  started_at        TEXT NOT NULL DEFAULT (datetime('now')),
  ended_at          TEXT,
  UNIQUE(node_id, principal_id, membership)
);

-- 019: skills
CREATE TABLE skills (
  id                TEXT PRIMARY KEY,
  tenant_id         TEXT NOT NULL,
  name              TEXT NOT NULL,
  kind              TEXT,                             -- technical | communication | strategic | domain | tool
  description       TEXT,
  UNIQUE(tenant_id, name)
);

-- 020: principal_skills
CREATE TABLE principal_skills (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id         TEXT NOT NULL,
  principal_id      TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
  skill_id          TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  level             TEXT NOT NULL DEFAULT 'intermediate', -- novice | intermediate | expert | lead
  evidence          TEXT,
  acquired_at       TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(principal_id, skill_id)
);
```

Indexes (migration 021): tenant-first on every new table.

---

## 4. The Elixir surface

```
lib/optimal_engine/workspace/
├── node.ex              # Node struct + CRUD + tree traversal
├── node_member.ex       # add_member / remove_member / members_of / nodes_of
├── skill.ex             # Skill struct + upsert + list_by_kind
└── principal_skill.ex   # grant / revoke / skills_of_principal / principals_with_skill
```

Public API (facade):

```elixir
OptimalEngine.Workspace.create_node(%{slug: "ai-masters", name: "AI Masters", kind: :project})
OptimalEngine.Workspace.add_member(node_id, principal_id, membership: :internal, role: "lead")
OptimalEngine.Workspace.grant_skill(principal_id, skill_id, level: :expert, evidence: "…")
OptimalEngine.Workspace.skills_of(principal_id)
OptimalEngine.Workspace.members_of(node_id)
OptimalEngine.Workspace.nodes_of(principal_id)
OptimalEngine.Workspace.children(node_id)
OptimalEngine.Workspace.ancestors(node_id)
```

All functions are tenant-scoped (default tenant when omitted).

---

## 5. How downstream phases use this

**Phase 4 — Classifier** can enrich chunk classification with node context:
knowing the signal landed in `nodes/sales/enterprise-pod`, classification picks
`genre: :decision-log` vs `:brief` with higher confidence.

**Phase 7 — Wiki Curator** uses **audiences**, which are role sets. An audience
like `sales` resolves to "every principal in the `sales` node-tree who has
the `sales` role or is tagged with skill `enterprise-sales`." The wiki page
is curated with that audience's skill+role vocabulary.

**Phase 8 — Retrieval** filters chunks by node membership: when an
"enterprise-sales" principal asks a question, chunks rooted in the `sales`
node tree get a boost; chunks from `engineering/platform-team` that mention
the same entity get a lower boost but don't get filtered out.

**Phase 9 — Connectors** grant tool access via `integration_grants` keyed on
principals + groups. A new `slack-connector` service-account principal is
granted a Slack token — the engine knows which node(s) that token can read
from via the same membership model.

---

## 6. Invariants

1. **Every node belongs to a tenant.** No cross-tenant node references.
2. **A node has at most one `parent_id`.** Trees, not DAGs.
3. **A principal is a member of a node via explicit `node_members` row.**
   No implicit membership by path or group alone. Audit-friendly.
4. **Skills are tenant-scoped identifiers.** Two tenants can each have a
   skill called "enterprise-sales" — they're different rows.
5. **Membership has time.** `started_at` + `ended_at` on every row so
   "who was on the sales team on 2026-03-15" is answerable.
6. **Skills have time.** `acquired_at` ditto. A principal's capability
   graph is queryable at any historical timestamp.

---

## 7. Seeding

Phase 3.5 migration 022 is a **backfill**:

- For every distinct `contexts.node` value seen in the tenant, create a
  `nodes` row with `kind: :domain, style: :internal, status: :active` and
  `path = "nodes/<value>"` so existing behavior continues working.
- Seed the default tenant with a baseline set of well-known nodes
  (`"inbox"`, `"team"`, `"money-revenue"`, `"new-stuff"`) that match the
  routing rules in `config.yaml`.

Keeps every existing caller working (routing rules, L0Cache, search) while
letting new code treat nodes as first-class.
