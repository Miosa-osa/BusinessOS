# Agent Skills Overview

**Created:** January 5, 2026
**Updated:** January 6, 2026
**Status:** Planning
**Audience:** BusinessOS Team

---

## Critical Concept: Skills vs Tools

> **Skills are instructions, NOT tools.**

This is the most important distinction to understand:

| Layer | What It Is | Format | Purpose |
|-------|-----------|--------|---------|
| **Tool** | Executable code | Go functions, MCP handlers | Does the actual work (DB queries, API calls) |
| **Skill** | Markdown instructions | SKILL.md files | Teaches agent WHEN and HOW to use tools |

### The Two-Layer Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SKILLS VS TOOLS                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   USER: "Show me my tasks by project"               
                         │
│                     │                                                        │
│                     ▼                                                        │
│   ┌─────────────────────────────────────────────────────────────────┐       │
│   │  AGENT (LLM)                                                     │       │
│   │                                                                  │       │
│   │  Step 1: Check <available_skills> in system prompt               │       │
│   │          └─ dashboard-management: "Configure dashboards..."      │       │
│   │          └─ task-management: "Create and manage tasks..."        │       │
│   │                                                                  │       │
│   │  Step 2: Match request → "dashboard-management" skill            │       │
│   │                                                                  │       │
│   │  Step 3: Load SKILL.md (instructions)                            │       │
│   │          └─ "For task views grouped by project,                  │       │
│   │              call configure_dashboard with:                      │       │
│   │              action=add_widget, widget_type=task_summary,        │       │
│   │              config={group_by: 'project'}"                       │       │
│   │                                                                  │       │
│   │  Step 4: Call the MCP TOOL with those parameters                 │       │
│   │                                                                  │       │
│   └─────────────────────────────────────────────────────────────────┘       │
│                     │                                                        │
│                     ▼                                                        │
│   ┌─────────────────────────────────────────────────────────────────┐       │
│   │  MCP TOOL: configure_dashboard (Go code)                         │       │
│   │  - Validates input                                               │       │
│   │  - Queries database                                              │       │
│   │  - Creates widget                                                │       │
│   │  - Returns result                                                │       │
│   └─────────────────────────────────────────────────────────────────┘       │
│                     │                                                        │
│                     ▼                                                        │
│   AGENT: "Done! I added a task summary grouped by project."                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Why Separate Skills from Tools?

| Problem | Solution with Skills |
|---------|---------------------|
| Agent doesn't know WHEN to use a tool | Skill says "use this tool when user asks about X" |
| Agent uses wrong parameters | Skill provides example mappings: "user says X → params Y" |
| Too much context loaded | Skills use progressive disclosure (metadata first, full content on-demand) |
| Hard to update agent behavior | Edit SKILL.md files without redeploying code |
| Can't test decision-making | Test skill matching independently from tool execution |

---

## What is a Skill?

A skill is a **folder containing a SKILL.md file** with instructions for the agent.

### Directory Structure

```
skills/
├── skills.yaml                      # Config: which skills are enabled
│
├── dashboard-management/            # One skill
│   ├── SKILL.md                     # Required: instructions for agent
│   └── references/                  # Optional: detailed docs
│       ├── WIDGETS.md
│       └── EXAMPLES.md
│
├── task-management/                 # Another skill
│   └── SKILL.md
│
└── analytics-insights/              # Another skill
    ├── SKILL.md
    └── references/
        └── METRICS.md
```

### The SKILL.md File

Every skill has YAML frontmatter + Markdown instructions:

```markdown
---
name: dashboard-management
description: Create and configure custom dashboards with widgets. Add task summaries, burndown charts, project progress, and metric cards.
metadata:
  version: "1.0.0"
  author: businessos
  tools_used:
    - configure_dashboard
---

# Dashboard Management

## When to Use This Skill

Use when the user wants to:
- Create a new dashboard
- Add widgets to see their data
- Customize what they see on login
- Remove or rearrange widgets

## Available Tool

This skill uses the `configure_dashboard` MCP tool.

## Request → Tool Mapping

| User Says | Tool Call |
|-----------|-----------|
| "Show me my tasks" | `configure_dashboard(action=add_widget, widget_type=task_summary)` |
| "I want to see deadlines" | `configure_dashboard(action=add_widget, widget_type=upcoming_deadlines)` |
| "Create a project dashboard" | `configure_dashboard(action=create_dashboard, name="Project View")` |
| "Group tasks by project" | `configure_dashboard(action=add_widget, widget_type=task_summary, config={group_by:"project"})` |

## Examples

### Simple: Add a widget
User: "Show me what's due this week"
→ Call: `configure_dashboard(action=add_widget, widget_type=upcoming_deadlines, config={days_ahead:7})`
→ Say: "Added a deadlines widget showing the next 7 days."

### Complex: Create dashboard with widgets
User: "Create a dashboard for the Alpha project"
→ Call: `configure_dashboard(action=create_dashboard, name="Alpha Project", widgets=[{type:"project_progress"}, {type:"task_burndown", config:{days:30}}])`
→ Say: "Created 'Alpha Project' dashboard with progress and burndown charts."

## Error Handling

If the tool returns an error with `available_alternatives`, suggest those to the user.

Example:
- Tool returns: `{error: "team_workload not available", alternatives: ["workload_heatmap"]}`
- Say: "Team workload widget isn't available yet. I can add a workload heatmap instead - would you like that?"
```

---

## Progressive Disclosure: Token Efficiency

Skills use **three levels** of loading to minimize context usage:

```
LEVEL 1: DISCOVERY (Always in system prompt)
══════════════════════════════════════════

<available_skills>
  <skill name="dashboard-management">
    Create and configure custom dashboards with widgets.
  </skill>
  <skill name="task-management">
    Create, update, and organize tasks.
  </skill>
</available_skills>

Cost: ~50 tokens per skill
Agent knows: "These capabilities exist"


LEVEL 2: ACTIVATION (When request matches)
══════════════════════════════════════════

User: "Add a task chart to my dashboard"
       ↓
Agent: Matches "dashboard-management"
       ↓
Load full SKILL.md into context (~1500 tokens)

Agent knows: "Here's exactly how to do this"


LEVEL 3: REFERENCES (On demand)
═══════════════════════════════

Agent: "I need widget details"
       ↓
Load references/WIDGETS.md (~500 tokens)

Agent knows: "Here are all 12 widget types"
```

### Token Comparison

| Approach | Tokens Used | When |
|----------|-------------|------|
| Skills discovery only | ~300-500 | Always (idle) |
| One skill activated | ~1500-2500 | When needed |
| Hardcoded everything | ~8000-15000 | Always (wasteful) |

---

## Skill Discovery: How Agent Matches

The agent receives `<available_skills>` in its system prompt and matches based on description keywords:

```
User: "Show me a chart of tasks completed this week"

Agent scans available_skills:

┌─────────────────────────────────────────────────────────────┐
│ dashboard-management                                         │
│ "Create and configure custom dashboards with widgets.        │
│  Add task summaries, burndown charts..."                     │
│                                        ↑                     │
│                                  MATCH: "chart", "tasks"     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ analytics-insights                                           │
│ "Fetch analytics data including KPIs, burndown charts,       │
│  workload heatmaps..."                                       │
│                        ↑                                     │
│                  MATCH: "chart", "completed"                 │
└─────────────────────────────────────────────────────────────┘

Agent activates best match (or both if needed)
```

### Writing Good Descriptions

**Good (keywords for matching):**
```
Create and configure custom dashboards with widgets. Add task summaries,
burndown charts, project progress bars, upcoming deadlines, and metric cards.
```

**Bad (too vague):**
```
Handles dashboards.
```

---

## Skill vs Tool: Quick Reference

| Question | Skill (SKILL.md) | Tool (Go code) |
|----------|------------------|----------------|
| What is it? | Markdown instructions | Executable function |
| Where does it live? | `skills/` directory | `internal/tools/` |
| Who reads it? | The agent (LLM) | The MCP server |
| What does it do? | Teaches WHEN/HOW to act | Actually DOES the action |
| How to update? | Edit markdown file | Redeploy code |
| Can it be tested? | Test matching logic | Test function output |

---

## Implementation: What You Need to Build

### 1. MCP Tools (Go code) - The Execution Layer

These are the actual functions that do work:

| Tool | File | Purpose |
|------|------|---------|
| `configure_dashboard` | `internal/tools/dashboard_tool.go` | CRUD for dashboards/widgets |
| `query_analytics` | `internal/tools/analytics_tool.go` | Fetch metrics data |
| `manage_tasks` | `internal/tools/task_tool.go` | CRUD for tasks |

See `DASHBOARD_AGENT_TOOL.md` for the dashboard tool spec.

### 2. Skills (Markdown) - The Instruction Layer

These teach the agent when/how to use tools:

| Skill | Path | Tools Used |
|-------|------|------------|
| dashboard-management | `skills/dashboard-management/SKILL.md` | `configure_dashboard` |
| analytics-insights | `skills/analytics-insights/SKILL.md` | `query_analytics` |
| task-management | `skills/task-management/SKILL.md` | `manage_tasks` |

### 3. Skills Loader (Go service) - Discovery System

Reads skills.yaml, parses SKILL.md frontmatter, exposes via API:

```go
// internal/services/skills_loader.go

type SkillsLoader struct {
    configPath string
    skills     []SkillMetadata
}

func (l *SkillsLoader) GetEnabledSkills() []SkillMetadata
func (l *SkillsLoader) GetSkillContent(name string) (string, error)
func (l *SkillsLoader) GetSkillsPromptXML() string  // For system prompt
```

### 4. Skills API (Endpoints)

```
GET /api/agent/skills          → List enabled skills (metadata only)
GET /api/agent/skills/:name    → Get full SKILL.md content
GET /api/agent/skills/prompt   → Get <available_skills> XML block
```

---

## BusinessOS Skills Plan

### Phase 1: Foundation
- [ ] Create `skills/` directory structure
- [ ] Create `skills.yaml` config
- [ ] Build `SkillsLoader` service
- [ ] Add `/api/agent/skills` endpoints

### Phase 2: Dashboard (Most Complex)
- [ ] Build `configure_dashboard` MCP tool
- [ ] Create `skills/dashboard-management/SKILL.md`
- [ ] Create reference files (WIDGETS.md, EXAMPLES.md)

### Phase 3: Other Skills
- [ ] `analytics-insights` skill
- [ ] `task-management` skill
- [ ] `project-management` skill
- [ ] `notification-management` skill

### Phase 4: Integration
- [ ] Inject `<available_skills>` into agent system prompt
- [ ] Implement skill activation flow
- [ ] Test end-to-end

---

## Security Notes

Since BusinessOS uses MCP tools (not shell execution):

- **Tools have auth middleware** - User must be logged in
- **Queries scoped to user** - Can only access own data
- **No arbitrary code execution** - Tools are predefined functions
- **Input validation** - JSON schema enforced by MCP

---

## Key Takeaways

1. **Skills = Instructions** (Markdown) - Tell agent WHEN/HOW to use tools
2. **Tools = Execution** (Go code) - Actually DO the work via MCP
3. **Progressive Disclosure** - Load metadata first, full skill on activation
4. **Good Descriptions Matter** - Agent matches based on keywords
5. **Edit Skills Without Redeploy** - Just update SKILL.md files
6. **Test Separately** - Skill matching vs tool execution

---

## Related Documents

| Document | Purpose |
|----------|---------|
| `DASHBOARD_AGENT_TOOL.md` | MCP tool specification for dashboards |
| `AGENT_SKILLS_TASK_LIST.md` | Implementation checklist |
| `FUTURE_FEATURES.md` | Feature 4: Analytics/Custom Dashboards |