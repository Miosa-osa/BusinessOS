# Optimal Engine Skill

Claude Code skill for building with [Optimal Engine](../../README.md) — the second brain of a company.

## What This Skill Does

When activated, this skill gives Claude Code full working knowledge of:
- Every HTTP API endpoint (method, params, response shape)
- The 3-tier memory architecture and 9-stage ingestion pipeline
- When to use `ask` vs `search` vs `grep` vs `recall` vs `profile`
- Memory primitive: versioning, relations, forgetting
- Workspace isolation and config schema
- Runnable integration examples in TypeScript, Python, MCP, and curl

## Activation

The skill auto-activates when you say things like:
- "I need persistent memory for my agent"
- "Build a second brain for our team"
- "Set up RAG with citations"
- "My agent should remember what happened last week"
- "Knowledge base", "organizational memory", "workspace knowledge"

Or reference it directly with `@optimal-engine`.

## Contents

```
skills/optimal-engine/
├── SKILL.md                    Manifest (Claude Code reads this)
├── README.md                   This file
├── references/
│   ├── api-reference.md        Every endpoint with params + examples
│   ├── concepts.md             3 tiers, 9 stages, intents, S=(M,G,T,F,W)
│   ├── workspace-pattern.md    Multi-workspace design + config schema
│   ├── memory-pattern.md       Versioning, relations, forgetting
│   ├── retrieval-pattern.md    Decision tree: which endpoint to use
│   └── integration-examples.md TypeScript, Python, MCP, curl, Elixir
└── scripts/
    └── bootstrap.sh            Scaffold + smoke-test a workspace
```

## Quick Start

```bash
# 1. Start the engine (needs config :api, enabled: true)
iex -S mix

# 2. Bootstrap a workspace
bash skills/optimal-engine/scripts/bootstrap.sh my-workspace

# 3. Ingest your knowledge
mix optimal.ingest_workspace ~/my-workspace/

# 4. Query
curl -X POST http://localhost:4200/api/rag \
  -H 'Content-Type: application/json' \
  -d '{"query":"what do we know?","workspace":"my-workspace","format":"markdown"}'
```

## Key References

- Engine README: [`../../README.md`](../../README.md)
- Architecture: [`../../docs/architecture/ARCHITECTURE.md`](../../docs/architecture/ARCHITECTURE.md)
- Signal Theory: [`../../docs/concepts/signal-theory.md`](../../docs/concepts/signal-theory.md)
- Mix tasks: [`../../docs/guides/mix-tasks.md`](../../docs/guides/mix-tasks.md)
