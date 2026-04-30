---
name: optimal-engine
description: |
  Optimal Engine is the second brain of a company — open-source memory infrastructure
  that ingests every signal flowing through an org (chat, email, docs, meetings,
  tickets, CRM, code, voice, video), curates it into an LLM-maintained wiki, and
  delivers permission-scoped context to any agent in under 200ms. Use this skill
  when building applications that need: persistent organizational memory, multi-
  workspace knowledge isolation, audience-aware curation, hot-cited retrieval, or
  proactive surfacing. Architectural commitments: 3-tier memory (raw / derivatives
  / wiki), 9-stage pipeline, S=(M,G,T,F,W) signal classification, 14-category
  proactive surfacing, 5-relation memory graph (updates / extends / derives /
  contradicts / cites). Self-hosted, MIT, runs on a laptop.
triggers:
  - persistent memory
  - organizational memory
  - memory across sessions
  - agent should remember
  - knowledge base
  - second brain
  - wiki for the team
  - curated context
  - retrieval augmented generation
  - workspace knowledge
  - signal classification
  - proactive surfacing
---

# Optimal Engine Skill

## What Optimal Engine Is

The Optimal Engine is the **second brain of a company** — an intent machine that ingests every signal flowing through an organization, decodes the intent behind each one, curates it into a permission-aware wiki, and delivers the right context to any agent runtime in under 200ms.

**Three promises:**

1. **Signal integrity.** Nothing is stored without classification (`S=(M,G,T,F,W)`), intent, and citation lineage. The engine knows *why* each fact exists.
2. **Scale alignment.** Every piece of content exists simultaneously as document / section / paragraph / chunk. Retrieval returns the coarsest scale that answers the query.
3. **Modality alignment.** Text, image, and audio embed into the same 768-dim vector space (nomic-embed-text + nomic-embed-vision + whisper.cpp). A text query can retrieve an image.

**Four invariants — violate any one and the system rots:**

- Tier 1 is append-only. Nothing gets edited in place.
- Tier 2 is fully rebuildable from Tier 1. `mix optimal.rebuild` reconstructs it.
- Tier 3 is LLM-owned. Humans write the schema; the curator writes the pages.
- Citations only point downward: T3 → T2/T1, T2 → T1. Never upward.

**Competitive position:** Classical RAG re-discovers the same facts on every query. Optimal Engine discovers once, curates forever, and delivers permission-gated, receiver-matched context. **Wiki-first, not chunk-and-pray. Curated memory beats infinite memory.**

---

## When to Use This Skill

Activate when the user says or implies:

- "I need persistent context across sessions"
- "My agent should remember what happened last week"
- "I want a knowledge base the whole team can query"
- "Build me a second brain / company wiki"
- "I need memory that survives conversation resets"
- "My agent needs to know who owns X / what Y decided"
- "I want to search all our documents semantically"
- "I need proactive surfacing — notify me when X changes"
- "I want RAG with citations I can trust"
- "Multi-tenant / workspace-isolated knowledge"

---

## Three Core Capabilities

### 1. Memory API

First-class memory entries: versioned, cited, relation-tracked. Not a key-value store — every memory carries provenance.

```bash
# Create
POST /api/memory  {"content": "Alice owns the pricing negotiation", "workspace": "sales"}

# Version it
POST /api/memory/:id/update  {"content": "Alice + Bob now co-own pricing"}

# Branch it
POST /api/memory/:id/extend  {"content": "Bob is the fallback if Alice is OOO"}

# Forget it (soft delete with reason)
POST /api/memory/:id/forget  {"reason": "resolved", "forget_after": "2026-12-31"}

# Navigate the graph
GET /api/memory/:id/versions   # full version chain
GET /api/memory/:id/relations  # inbound + outbound typed edges
```

Five relation types: `updates` / `extends` / `derives` / `contradicts` / `cites`.

### 2. Workspace + Profile

Workspaces are isolated brains — per-tenant, per-team, per-project. Profile delivers a 4-tier context snapshot in one call.

```bash
# Create a workspace
POST /api/workspaces  {"slug": "sales", "name": "Sales Brain"}

# 4-tier profile snapshot (static + dynamic + curated + activity)
GET /api/profile?workspace=sales&audience=executive&bandwidth=l1
```

Profile tiers:
- **static** — facts that rarely change (team composition, product definitions)
- **dynamic** — rolling context (this week's signals, open decisions)
- **curated** — the wiki front-door (LLM-maintained, audience-aware)
- **activity** — recent events (what happened in the last N days)

### 3. Retrieval

Wiki-first → hybrid fallback → typed recall. Never starts with the chunk index.

```bash
# Wiki-first open question (recommended for agents)
POST /api/rag  {"query": "what's the pricing decision?", "workspace": "sales", "format": "claude"}

# Keyword + semantic hybrid
GET /api/search?q=pricing+negotiation&workspace=sales

# Chunk-level grep with signal trace (intent, scale, sn_ratio)
GET /api/grep?q=pricing&workspace=sales&intent=record_fact&scale=paragraph

# Typed cued recall — one endpoint per memory-failure pattern
GET /api/recall/actions?actor=Alice&topic=pricing&since=2026-01-01
GET /api/recall/who?topic=pricing&role=owner
GET /api/recall/when?event=Q4+pricing+decision
GET /api/recall/where?thing=pricing+deck
GET /api/recall/owns?actor=Alice

# Contextual snapshot
GET /api/profile?workspace=sales&audience=sales&bandwidth=l1
```

---

## Quick Integration Examples

### TypeScript — Vercel AI SDK

```typescript
import Anthropic from '@anthropic-ai/sdk';

const ENGINE = process.env.OPTIMAL_ENGINE_URL ?? 'http://localhost:4200';

async function getContext(query: string, workspace: string): Promise<string> {
  const res = await fetch(`${ENGINE}/api/rag`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, workspace, format: 'claude', bandwidth: 'medium' }),
  });
  const data = await res.json() as { answer?: string; context?: string };
  return data.answer ?? data.context ?? '';
}

// Use with Vercel AI SDK
import { streamText } from 'ai';
import { anthropic } from '@ai-sdk/anthropic';

export async function POST(req: Request) {
  const { messages, workspace = 'default' } = await req.json();
  const lastMessage = messages[messages.length - 1].content as string;
  const memory = await getContext(lastMessage, workspace);

  return streamText({
    model: anthropic('claude-opus-4-5'),
    system: `You have access to organizational memory:\n\n${memory}`,
    messages,
  }).toDataStreamResponse();
}
```

### Python — OpenAI Agents SDK

```python
import os, requests
from agents import Agent, Runner, function_tool

ENGINE = os.getenv("OPTIMAL_ENGINE_URL", "http://localhost:4200")

@function_tool
def recall(query: str, workspace: str = "default") -> str:
    """Retrieve organizational memory for a query."""
    r = requests.post(f"{ENGINE}/api/rag", json={
        "query": query,
        "workspace": workspace,
        "format": "markdown",
        "bandwidth": "medium",
    })
    r.raise_for_status()
    return r.json().get("answer", "")

agent = Agent(
    name="Company Brain",
    instructions="You are a company assistant. Use recall() before answering any question about company decisions, people, or projects.",
    tools=[recall],
)

result = Runner.run_sync(agent, "What did Alice decide about Q4 pricing?")
print(result.final_output)
```

### MCP — Claude Desktop / Cursor / Windsurf

```json
{
  "mcpServers": {
    "optimal-engine": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp-server"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "default"
      }
    }
  }
}
```

### curl — raw HTTP

```bash
ENGINE=http://localhost:4200

# Ask the wiki
curl -s -X POST $ENGINE/api/rag \
  -H 'Content-Type: application/json' \
  -d '{"query":"current pricing strategy","workspace":"sales","format":"markdown"}' \
  | jq '.answer'

# Store a memory
curl -s -X POST $ENGINE/api/memory \
  -H 'Content-Type: application/json' \
  -d '{"content":"Alice confirmed $2K/seat pricing on 2026-04-28","workspace":"sales"}' \
  | jq '.id'

# Recall who owns something
curl -s "$ENGINE/api/recall/who?topic=pricing&workspace=sales" | jq '.answer'
```

---

## Key Concepts Cheat Sheet

| Concept | What it is | Why it matters |
|---|---|---|
| **Workspace** | Isolated brain (per-tenant, per-team) | Sales, Engineering, M&A never bleed into each other |
| **Memory** | First-class versioned entry with typed relations | Not a chunk — a claim with provenance |
| **Wiki (Tier 3)** | LLM-maintained curated front door | Most agent queries answered here — zero retriever hits |
| **Hot citations** | Every wiki claim → source chunk | Fail-closed integrity gate — no unverified assertions |
| **Signal `S=(M,G,T,F,W)`** | 5-dimension classification per chunk | Routes intent correctly, enables audience filtering |
| **Audience** | Tag on retrieval (`sales`, `legal`, `exec`, `engineering`) | Same wiki, different variants — no manual branching |
| **Bandwidth** | `l0` (~100 tok) / `l1` (~2K tok) / `full` | Match context density to LLM token budget |
| **9-stage pipeline** | Intake → Parse → Decompose → Classify → Embed → Route → Store → Cluster → Curate | Every signal gets full treatment, no shortcuts |
| **Proactive surfacing** | 14-category taxonomy pushed via SSE | Engine notifies agents when relevant facts change |
| **HDBSCAN clusters** | Incremental theme grouping | Context assembler expands a hit to its whole cluster |

---

## References

| Doc | Contents |
|---|---|
| [`references/api-reference.md`](references/api-reference.md) | Every endpoint — method, path, params, response, use-when |
| [`references/concepts.md`](references/concepts.md) | 3 tiers, 9 stages, 4 chunk scales, 10-intent enum, S=(M,G,T,F,W), 3 constraints |
| [`references/workspace-pattern.md`](references/workspace-pattern.md) | Multi-workspace design, config schema, isolation guarantees |
| [`references/memory-pattern.md`](references/memory-pattern.md) | When to add a memory, 5 relation types, versioning, forgetting |
| [`references/retrieval-pattern.md`](references/retrieval-pattern.md) | Decision tree: ask vs search vs grep vs recall vs profile |
| [`references/integration-examples.md`](references/integration-examples.md) | TypeScript, Python, MCP, curl — complete runnable examples |
| [`scripts/bootstrap.sh`](scripts/bootstrap.sh) | Scaffold a workspace and verify the pipeline |
