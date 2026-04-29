# Node Template — Standard Anatomy

> Every node in OptimalOS follows this structure. The context.md and signal.md
> together form the complete node. context.md = persistent facts that change slowly.
> signal.md = weekly state that changes fast.

---

## context.md Template

```markdown
---
signal:
  mode: linguistic
  genre: spec
  type: inform
  format: markdown
  structure: persistent-context
  sn_ratio: 0.9
  audience: [human, agent]
  intent: "[Node name] — persistent context"
node: [node-id]
cross_ref: [related-nodes]
valid_from: "[date]"
---

# [Node Name]

> [One-line purpose statement — why this node exists]

## Identity

| Field | Value |
|-------|-------|
| Type | [entity/operation/person/product/project/operational/learning/context] |
| Owner | [who is responsible] |
| Status | [draft/active/paused/blocked/completed/archived] |
| Health | [🟢 green / 🟡 yellow / 🔴 red] |
| Parent | [parent node] |
| Review | [weekly/biweekly/monthly] |

## Relationships

| Node | Relationship | How |
|------|-------------|-----|
| [node] | parent | [description] |
| [node] | depends-on | [what we need from them] |
| [node] | feeds | [what we send to them] |
| [node] | collaborates | [how we work together] |

### People Involved

| Person | Role Here | Genre | Channel |
|--------|----------|-------|---------|
| [name] | [their role in this node] | [what to send them] | [how to reach them] |

## Context

[Persistent facts about this node. What is true. What has been decided.
This is the body — update when ground truth changes.]

### Key Decisions

| Date | Decision | Rationale | Impact |
|------|----------|-----------|--------|
| [date] | [what was decided] | [why] | [what changed] |

### Open Questions

- [ ] [Question that needs answering]
- [ ] [Question that needs answering]

## Assets

[Links to key files, documents, repos, tools that belong to this node]

| Asset | Location | What |
|-------|----------|------|
| [name] | [path or URL] | [description] |
```

---

## signal.md Template

```markdown
---
signal:
  mode: linguistic
  genre: note
  type: inform
  format: markdown
  structure: weekly-status
  sn_ratio: 0.9
  audience: [human, agent]
  intent: "[Node name] weekly status"
node: [node-id]
valid_from: "[date]"
---

# [Node Name] — Signal (Week of [date])

## State

| Metric | Value | Target | Trend |
|--------|-------|--------|-------|
| Health | [🟢/🟡/🔴] | 🟢 | [↑/→/↓] |
| Progress | [X/10] | [target] | [↑/→/↓] |
| [domain-specific metric] | [value] | [target] | [trend] |

## Focus This Week

1. **[Non-negotiable 1]** — [what and why]
2. **[Non-negotiable 2]** — [what and why]
3. **[Non-negotiable 3]** — [what and why]

## Blockers

| What | Who Owns It | Status | Since |
|------|-------------|--------|-------|
| [blocker] | [person] | [status] | [date] |

## Progress Log

- [x] [What got done]
- [x] [What got done]
- [ ] [What didn't get done — why]

## Signals Out (Need to Send)

| To | What | Genre | Status |
|----|------|-------|--------|
| [person] | [what to send] | [brief/spec/etc] | [sent/pending/draft] |

## Fidelity Tracking

| Signal Sent | To | Expected Back | Status |
|-------------|-----|--------------|--------|
| [what] | [who] | [what we expect] | [status] |

## Next Week Preview

- [ ] [What's coming]
```
