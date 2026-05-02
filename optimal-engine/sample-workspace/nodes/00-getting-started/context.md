---
node: 00-getting-started
kind: guide
style: internal
pinned: true
---

# Getting Started with BusinessOS

Welcome to BusinessOS. This workspace comes pre-loaded with a sample company
to show you how everything fits together. Explore, edit, or delete anything
here — it's yours.

## What you're looking at

BusinessOS is built on OptimalEngine, a knowledge substrate that stores
everything your business knows as **nodes** (departments, people, projects)
and **signals** (weekly updates, decisions, invoices, meeting notes).

### Sample company structure

| Node | What it represents |
|------|-------------------|
| **Founder** | The CEO's desk — priorities, delegations, decision log |
| **Platform** | Engineering division — architecture, security, hiring |
| **Services** | Revenue & delivery — contracts, invoices, retainers |
| **Academy** | Customer education — courses, office hours, NPS |
| **Partners** | Channel partners — renewals, onboarding, at-risk accounts |
| **Media** | Content & brand — YouTube, customer stories, industry takes |

## How to use this

### 1. Dashboard
Your command center. Shows nodes, signals, today's rhythm, and quick access
to all modules. The intelligence feed surfaces proactive suggestions from
the engine.

### 2. Pages / Knowledge Graph
Click **Pages** in the sidebar, then **Knowledge Graph**. You'll see the
force-directed graph of all nodes and their connections. Click any node to
drill into its context and signals.

### 3. Nodes
The raw knowledge structure. Each node is a folder of markdown files:
- `context.md` — persistent facts (team, rules, priorities)
- `signal.md` — latest weekly status
- `signals/` — dated entries (decisions, meetings, invoices)

### 4. Chat + Recall
Open **Chat** and use the **Recall** tab in the right panel to pull context
from the engine into your conversations. The AI sees your full knowledge
graph.

### 5. Search (Cmd+K)
Press **Cmd+K** anywhere to search across all nodes, signals, and wiki
pages. Results link directly to the relevant module.

## Making it yours

1. **Edit any node** — update `context.md` with your real company info
2. **Add signals** — drop markdown files into `signals/` folders
3. **Create new nodes** — add a new folder under `nodes/` with a `context.md`
4. **Connect integrations** — Settings > Integrations to wire Google, Slack, etc.
5. **Invite your team** — Settings > Workspace to add team members

## Architecture

```
optimal-engine/
└── sample-workspace/
    └── nodes/
        ├── 00-getting-started/   ← You are here
        ├── 01-founder/           ← CEO context + signals
        ├── 02-platform/          ← Engineering division
        ├── 03-services/          ← Revenue & delivery
        ├── 04-academy/           ← Customer education
        ├── 06-partners/          ← Channel partners
        └── 08-media/             ← Content & brand
```

Each node folder is self-contained. The engine indexes everything
automatically — no config needed.

## Need help?

- **Docs**: github.com/Miosa-osa/BusinessOS
- **OptimalEngine**: github.com/Miosa-osa/OptimalEngine
- **Community**: github.com/Miosa-osa/BusinessOS/discussions
