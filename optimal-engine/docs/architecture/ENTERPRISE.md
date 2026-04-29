# Enterprise Scale — What the Optimal Engine Is Really For

> The Optimal Engine is the **second brain of a company** — not a personal
> knowledge tool. Every architectural decision assumes enterprise volumes,
> multi-tenant isolation, permission-gated access at the chunk level, and
> integration with the systems a company already runs on (chat, email, docs,
> tickets, CRM, code, meetings).

This doc sits beside `ARCHITECTURE.md` and `UI.md`. It specifies what enterprise
scale changes about the architecture. When `ARCHITECTURE.md` and this doc
disagree, this doc wins.

---

## 1. The competitive position

The Optimal Engine competes at the **application layer** against Glean, Dust,
and Google NotebookLM — the category that tries to give entire companies
context-aware AI. At the **infrastructure layer** it overlaps with OpenViking
(ByteDance) and MemOS. Most vector databases (Pinecone, Weaviate, Qdrant,
Chroma) and retrieval frameworks (LlamaIndex, LangChain, GraphRAG) are Layer-5
commodities — we use patterns from them but don't compete against them.

### What everyone else has

| Platform                | Strength                                          | Missing                                                      |
|------------------------|---------------------------------------------------|--------------------------------------------------------------|
| Glean                   | 100+ enterprise integrations, permission-aware RAG| No classification; no genre composition; no tiered disclosure|
| Dust                    | "AI OS for Work", MCP, enterprise governance      | No classification; no receiver modeling                      |
| NotebookLM              | 1M context, multimedia output, Workspace-native   | Per-notebook silos; no organizational topology               |
| OpenViking (ByteDance)  | Filesystem paradigm, L0/L1/L2, self-evolution     | No signal classification; agent-centric, not org-centric     |
| Mem0 / Letta / Zep / Cognee / HydraDB | Memory primitives                   | Agent-scale; no receiver modeling; no composition             |
| MCP                     | Universal tool/resource interface                 | Eats 40–50% of context window; no classification             |

### What we have that none of them do

The entire middle of the stack. Everyone else built **storage → retrieval → dump to LLM**. We build:

```
Classification → Routing → Composition → Tiered Assembly → Delivery → Feedback
```

With any storage engine underneath. Signal Theory is the layer nobody else has.

---

## 2. What enterprise scale actually changes

| Dimension              | Personal tool                     | Enterprise second brain                                          |
|-----------------------|-----------------------------------|------------------------------------------------------------------|
| **Users**              | 1 human + 1 agent                 | 100–10,000 humans × M agents per human × service accounts        |
| **Signals**            | 1K–100K files                     | 10M–1B+ signals, 10TB+ raw, millions ingested per day            |
| **Sources**            | filesystem + pasted text          | Slack, Gmail, Drive, Docs, Notion, Jira, Linear, GitHub, Salesforce, HubSpot, Zoom, Intercom, Zendesk, 100+ connectors |
| **Authority**          | user owns everything              | ACLs at chunk level; role-based visibility; legal holds; retention policies |
| **Audit**              | none                              | Every read and write logged; who-saw-what for compliance         |
| **Tenancy**            | single-tenant, local              | Multi-tenant (per-org isolation) with optional self-hosted deploy|
| **Compliance**         | not in scope                      | SOC 2 Type II, GDPR, HIPAA, CCPA, ISO 27001                     |
| **SLA**                | best-effort                       | p99 < 200ms wiki reads, p99 < 2s RAG, 99.9% uptime               |
| **Failure model**      | losses acceptable                 | Append-only Tier 1 + backup/restore + point-in-time recovery     |

This isn't a flag-switch. It's a set of architectural non-negotiables that
must be designed in from the start.

---

## 3. Required primitives (beyond the 9-stage pipeline)

### 3.1 Tenancy

Every row in every table carries a `tenant_id`. There is no cross-tenant data
access at the database level. Multi-tenant deployments run one Elixir VM per
tenant or one VM with strict tenant isolation at the Store layer (default:
per-tenant SQLite file under `tenants/<tenant_id>/index.db`, with the tenant
wiki at `tenants/<tenant_id>/.wiki/`).

### 3.2 Identity and ACLs

**Identity providers:**
- SAML 2.0, OIDC (Okta, Azure AD, Google Workspace, JumpCloud)
- Group membership synced hourly via SCIM 2.0
- Service accounts (for connectors) with separately-scoped tokens

**Permissions at chunk level:**

| Table              | New columns                                                         |
|--------------------|---------------------------------------------------------------------|
| `chunks`           | `acl_read: [principal]`, `acl_write: [principal]`, `classification_level: :public \| :internal \| :confidential \| :restricted` |
| `wiki_pages`       | `acl_read: [principal]`, `audience: [role]`                        |
| `citations`        | propagate visibility from source chunk — a wiki page citing a restricted chunk is visible only to principals who can read the source |

**Enforcement:** every query carries the caller's principal set. The Store
filters at query time (no post-filter leaks). When a wiki page would cite a
chunk the caller can't see, the citation is rendered as an opaque handle
with a "request access" action — never the content.

### 3.3 Connectors (source integrations)

Every connector is a separately-configured, rate-limited, incremental syncer
that writes to Tier 1 (Raw Sources). Connectors do not touch Tier 2 or Tier 3
— the pipeline handles that.

**Phase 1 connector set (the minimum to be competitive with Glean):**

| # | Source              | Mode                       | Auth                 |
|---|--------------------|-----------------------------|-----------------------|
| 1 | Local filesystem    | filesystem watch            | OS permissions        |
| 2 | Slack               | events API + backfill       | OAuth bot token       |
| 3 | Google Drive        | Changes API + OAuth         | OAuth 2.0             |
| 4 | Google Docs         | via Drive + Docs API        | OAuth 2.0             |
| 5 | Gmail               | history API                 | OAuth 2.0             |
| 6 | Microsoft 365       | Graph API (Outlook + OneDrive + SharePoint) | OAuth 2.0 |
| 7 | Notion              | polling + webhook           | OAuth / integration   |
| 8 | Jira / Linear       | webhooks + polling          | OAuth / API token     |
| 9 | GitHub / GitLab     | webhooks + GraphQL          | OAuth App / PAT       |
| 10| Zoom / Meet / Teams | transcript webhook + fetch  | OAuth 2.0             |
| 11| Confluence          | REST + CQL                  | OAuth / API token     |
| 12| Salesforce / HubSpot| Bulk API + streaming        | OAuth 2.0             |
| 13| Intercom / Zendesk  | webhooks + polling          | OAuth / API token     |
| 14| Custom HTTP webhook | push                        | signed request        |

**Connector contract:** idempotent, resumable, honors rate limits, reports
incremental cursor state to the Store, emits structured errors the Queue view
can surface.

### 3.4 Permission-aware retrieval

The hybrid retriever (`SearchEngine`) accepts a `principal` argument. It
produces a SQL WHERE clause joining to the ACL tables. Empty result sets
due to permission filtering are returned as `{:ok, [], filtered: N}` so the
composer can tell the agent "N results exist but you don't have access —
here's the access-request affordance."

Permissions are also audited: every access attempt (granted or denied) writes
to `events` with principal, query, chunks returned, chunks filtered.

### 3.5 Retention and legal hold

| Rule                      | Mechanism                                                                 |
|---------------------------|---------------------------------------------------------------------------|
| Retention policy          | Per-node, per-genre, or per-tenant. TTL on `signals`. Default: infinite. |
| Automatic compaction      | Signals past retention are moved to cold storage (S3/Glacier) + replaced with an abstract + citation remains valid. |
| Legal hold                | Principal with `legal:hold` role can pin a signal. Overrides retention. Expires only on explicit release. |
| Right-to-be-forgotten     | GDPR delete propagates to Tier 2 derivatives + queues curation of affected Tier 3 pages to remove citations. |

### 3.6 Audit trail

Everything goes into `events`. Schema:

```sql
CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id TEXT NOT NULL,
  ts TEXT NOT NULL,
  principal TEXT NOT NULL,
  kind TEXT NOT NULL,          -- ingest.started, ingest.completed, retrieval.executed,
                                -- wiki.curated, permission.denied, connector.synced, etc.
  target_uri TEXT,
  latency_ms INTEGER,
  metadata TEXT                 -- JSON
);
CREATE INDEX events_tenant_ts ON events(tenant_id, ts);
CREATE INDEX events_principal ON events(principal, ts);
CREATE INDEX events_kind ON events(kind, ts);
```

Exportable to SIEM (Splunk, Datadog, Elastic) via a standard JSON schema.

### 3.7 Deployment modes

| Mode                         | When                                     |
|------------------------------|------------------------------------------|
| Single-tenant SaaS           | Default. Per-org dedicated VM + storage. |
| Self-hosted (Docker Compose) | Regulated industries (HIPAA, FedRAMP).   |
| Self-hosted (Kubernetes)     | Large enterprises with ops teams.        |
| Hybrid                       | Metadata in SaaS; raw sources on-prem.   |

All four modes use the same engine. Configuration differs.

---

## 4. Schema additions (over what ARCHITECTURE.md specifies)

| Table                | Purpose                                                      |
|---------------------|--------------------------------------------------------------|
| `tenants`            | Tenant metadata (name, plan, created_at, region)             |
| `principals`         | Humans, service accounts, agents                             |
| `groups`             | SCIM-synced groups (membership edges in `principal_groups`)  |
| `roles`              | Tenant-defined roles (e.g., `sales`, `legal:hold`, `audit`)  |
| `role_grants`        | Principal/group → role                                       |
| `acls`               | (resource_uri, principal_or_group, permission) rows          |
| `connectors`         | Per-tenant connector config + cursor state                   |
| `connector_runs`     | Per-sync run history                                         |
| `retention_policies` | Per-node/genre retention rules                                |
| `legal_holds`        | Pinned signals + pinner + reason                             |
| `audiences`          | Role sets that define who a wiki page is curated for         |

All primary tables (`signals`, `chunks`, `wiki_pages`, `citations`, etc.) gain
`tenant_id` as the first column. Every index leads with it.

---

## 5. The enterprise Wiki layer (role-aware curation)

A single source signal can produce **multiple wiki-page variants** — one per
audience. The curator prompts include the audience's role and allowed content
classification level. Examples:

| Signal source              | Audience: sales              | Audience: engineering         | Audience: exec-brief           |
|---------------------------|------------------------------|-------------------------------|--------------------------------|
| Customer escalation ticket | "ACME is unhappy about X — open opportunity to propose Y" | "API endpoint X degrading; owning team Z" | "ACME churn risk; est. ARR impact" |
| Engineering design doc     | Hidden (classification: restricted) | Full doc with cite chain      | "New platform capability landing in Q3" |
| Financial projection deck  | "Here's what you can promise customers" | Hidden                         | Full numbers with cite chain   |

The curator enforces: **no cross-audience leakage**. A `sales` wiki page never
includes a restricted engineering chunk, even transitively. Integrity check
verifies every citation on a page resolves for every principal in the page's
audience.

---

## 6. Performance targets

Enterprise scale demands actual numbers, not "fast."

| Operation                              | p50     | p99      | Notes                             |
|---------------------------------------|---------|----------|-----------------------------------|
| Wiki page read (cache hit)             | < 20ms  | < 100ms  | Served from memory                |
| Wiki page read (cache miss)            | < 80ms  | < 300ms  | SQLite + ACL check                |
| `optimal rag` query (wiki hit)         | < 50ms  | < 200ms  | No retriever path                 |
| `optimal rag` query (wiki miss → hybrid) | < 400ms | < 2s    | BM25 + vector + graph + compose   |
| Ingest one signal (text)               | < 500ms | < 3s     | Full 9-stage pipeline             |
| Ingest one signal (PDF, 10 pages)      | < 3s    | < 15s    | Includes parse + OCR              |
| Connector sync batch (1K items)        | < 30s   | < 2min   | Rate-limited by source            |
| Wiki curation (one page, incremental)  | < 10s   | < 60s    | Dominated by Ollama call          |
| Audit query ("what did Ada read Mon?") | < 100ms | < 500ms  | Indexed `events` query            |

These targets are the acceptance criteria for Phase 10 (production hardening).

---

## 7. How this changes the build order

Enterprise-scale concerns don't replace the 9-stage pipeline — they wrap it.
The revised build order over `ARCHITECTURE.md`'s 11 phases:

| Phase | Topic                                                              |
|-------|--------------------------------------------------------------------|
| 1     | Storage schema — **now includes** tenancy, principals, ACLs, events tables from the start, not later |
| 2     | Parser backends                                                    |
| 3     | Decomposer                                                         |
| 4     | Classify + IntentExtractor                                         |
| 5     | Embedder (multi-modal)                                             |
| 6     | Clusterer                                                          |
| 7     | Wiki Layer (audience-aware curation)                               |
| 8     | Scale-aware Deliver + Composer (permission-aware retriever)        |
| 9     | Connectors (Phase 1 set: Slack, Drive, Docs, Gmail, Notion, Jira, GitHub, Zoom) |
| 10    | Production hardening (perf targets, backup/restore, SIEM export, telemetry) |
| 11    | Compliance (SOC 2 audit trail, GDPR flows, HIPAA-ready deployment) |
| 12    | Push                                                                |

Phase 9 (Connectors) is **not optional**. Without it, we're a better OpenViking
— a filesystem-paradigm context engine — but not a Glean/Dust competitor. The
whole value is that an organization's existing context is already indexed.

---

## 8. What this means for the invariants

The three invariants from `ARCHITECTURE.md` get two additions:

4. **Tenant isolation is absolute.** No query crosses tenants. No wiki page
   cites a chunk from another tenant. No error message reveals another tenant's
   existence.
5. **Permissions propagate.** A wiki claim's visibility is the intersection of
   visibilities of all cited chunks. Never the union. Never inferred.

Violate either and we're not an enterprise second brain — we're a liability.
