# Business OS - Agent Architecture

## Overview

BusinessOS is an internal command center for businesses. It manages Projects, Tasks, Clients, Team, and Knowledge.

**NO code generation** - that's OSA Terminal's job (separate system).

---

## Agents

| Agent | Role | Tools |
|-------|------|-------|
| **Orchestrator** | Primary interface, routing, quick tasks | `search_documents`, `get_project`, `get_task`, `get_client`, `create_task`, `log_activity` |
| **Document** | Proposals, SOPs, reports, frameworks | `create_artifact`, `search_documents`, `get_project`, `get_client` |
| **Project** | Project planning, milestones, team allocation | `create_project`, `update_project`, `create_task`, `bulk_create_tasks`, `assign_task`, `get_team_capacity` |
| **Task** | Task management, prioritization, scheduling | `create_task`, `update_task`, `bulk_create_tasks`, `move_task`, `assign_task`, `get_team_capacity` |
| **Client** | CRM, pipeline, interactions | `create_client`, `update_client`, `log_client_interaction`, `update_client_pipeline` |
| **Analyst** | Metrics, analysis, insights | `query_metrics`, `get_team_capacity`, `list_projects`, `list_tasks`, `create_artifact` |

---

## Routing Logic

```
USER REQUEST
    │
    ▼
FORMAL DOCUMENT? (proposal, SOP, report) ──► Document Agent
    │ NO
    ▼
PROJECT MANAGEMENT? (planning, milestones) ──► Project Agent
    │ NO
    ▼
TASK MANAGEMENT? (bulk, prioritization) ──► Task Agent
    │ NO
    ▼
CLIENT MANAGEMENT? (pipeline, interactions) ──► Client Agent
    │ NO
    ▼
DATA ANALYSIS? (metrics, trends) ──► Analyst Agent
    │ NO
    ▼
ORCHESTRATOR handles directly
```

---

## Data Modules

- **Projects**: Project, Milestones, Project Members
- **Tasks**: Task, Subtasks, Dependencies
- **Clients**: Client, Contacts, Pipeline
- **Team**: Team Member, Roles, Capacity
- **Knowledge**: Contexts, Documents, Blocks
- **Daily Log**: Log Entry, Activities
- **Calendar**: Events, Schedule

---

## Focus Modes (Frontend)

| Mode | Agent |
|------|-------|
| Research | Analyst |
| Analyze | Analyst |
| Write | Document |
| Build | Project |
| Do more | Orchestrator |
