# Voice Agent Context Hierarchy Plan

## Overview

The voice agent should NOT receive all context at once. Instead, it follows a hierarchical approach where Nodes are at the top, and deeper information is fetched on-demand via tool calls.

## Context Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│ LEVEL 0: IDENTITY (Always Loaded - ~200 tokens)                 │
├─────────────────────────────────────────────────────────────────┤
│ - Agent personality ("You are OSA...")                          │
│ - Voice style guidelines (casual, 5-10 words)                   │
│ - Core capabilities overview                                    │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ LEVEL 1: USER CONTEXT (Loaded on Session Start - ~300 tokens)   │
├─────────────────────────────────────────────────────────────────┤
│ - User name, role                                               │
│ - Current workspace name                                        │
│ - Current active Node (if any)                                  │
│ - Recent node names (last 3-5 accessed)                         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ LEVEL 2: NODE CONTEXT (Fetched via Tool Call)                   │
├─────────────────────────────────────────────────────────────────┤
│ Node Identity:                                                  │
│   - Name, Type, Purpose, Status, Owner                          │
│ Node Relationships:                                             │
│   - Parent, Children (names only), Connected nodes              │
│ Node State:                                                     │
│   - Current status, Health, Progress                            │
│ Node Focus:                                                     │
│   - This week's priorities, Active projects, Blockers           │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ LEVEL 3: DEEP CONTEXT (Fetched via Specific Tool Calls)         │
├─────────────────────────────────────────────────────────────────┤
│ get_node_children(node_id)      → Child node details            │
│ get_node_decisions(node_id)     → Decision queue & history      │
│ get_node_documents(node_id)     → Key documents                 │
│ get_project_tasks(project_id)   → Task list with status         │
│ get_person_details(person_id)   → Full person profile           │
│ search_nodes(query)             → Find nodes by name/content    │
│ get_recent_activity(node_id)    → Activity log                  │
└─────────────────────────────────────────────────────────────────┘
```

## Node Type Hierarchy (from user spec)

```
NODE TYPES
├── ENTITY NODE      (Company/Organization)
│   └── Contains: Departments, Teams, Projects
├── DEPARTMENT NODE  (Functional area)
│   └── Contains: Teams, Projects, Processes
├── TEAM NODE        (Group of people)
│   └── Contains: People, Projects, Processes
├── PROJECT NODE     (Bounded initiative)
│   └── Contains: Tasks, Milestones, Deliverables
├── OPERATIONAL NODE (Ongoing process)
│   └── Contains: SOPs, Workflows, Metrics
├── LEARNING NODE    (Knowledge)
│   └── Contains: Resources, Notes, Frameworks
├── PERSON NODE      (Individual)
│   └── Contains: Profile, History, Relationships
├── PRODUCT NODE     (Product/Platform)
│   └── Contains: Features, Roadmap, Users
├── PARTNERSHIP NODE (Collaboration)
│   └── Contains: Terms, People, Projects
└── CONTEXT NODE     (Pure reference)
    └── Contains: Documents, Data, References
```

## Tool Definitions for Voice Agent

### 1. get_node_context
```json
{
  "name": "get_node_context",
  "description": "Get the full context of a specific node including identity, relationships, state, and focus",
  "parameters": {
    "node_id": "string (required) - The node ID or name"
  }
}
```

### 2. get_node_children
```json
{
  "name": "get_node_children",
  "description": "Get all child nodes of a parent node",
  "parameters": {
    "node_id": "string (required) - The parent node ID"
  }
}
```

### 3. search_nodes
```json
{
  "name": "search_nodes",
  "description": "Search for nodes by name, type, or content",
  "parameters": {
    "query": "string (required) - Search query",
    "type": "string (optional) - Filter by node type"
  }
}
```

### 4. get_project_tasks
```json
{
  "name": "get_project_tasks",
  "description": "Get tasks for a specific project node",
  "parameters": {
    "project_id": "string (required) - The project node ID"
  }
}
```

### 5. get_recent_activity
```json
{
  "name": "get_recent_activity",
  "description": "Get recent activity and updates for a node",
  "parameters": {
    "node_id": "string (required) - The node ID",
    "limit": "number (optional) - Number of items, default 5"
  }
}
```

### 6. get_node_decisions
```json
{
  "name": "get_node_decisions",
  "description": "Get pending decisions and recent decision history for a node",
  "parameters": {
    "node_id": "string (required) - The node ID"
  }
}
```

## System Prompt Structure

```
LEVEL 0 (Identity):
You are OSA, a casual voice assistant. Keep responses SHORT (5-10 words max).
You help users navigate their Business OS using nodes.

LEVEL 1 (User Context - injected):
User: {user_name}
Workspace: {workspace_name}
Current Node: {current_node_name} ({current_node_type})
Recent Nodes: {recent_node_names}

AVAILABLE TOOLS:
- get_node_context(node_id) - Get full node details
- get_node_children(node_id) - List child nodes
- search_nodes(query) - Find nodes
- get_project_tasks(project_id) - Get project tasks
- get_recent_activity(node_id) - Get recent updates
- get_node_decisions(node_id) - Get decisions

When user asks about something specific, USE A TOOL to get the info.
Don't guess - fetch the data first, then respond briefly.
```

## Implementation Steps

1. **Update voice agent prompt** - Use minimal Level 0 + Level 1 context
2. **Add tool definitions** to Groq LLM call
3. **Implement tool handlers** in Go backend
4. **Handle tool call responses** - Parse and execute
5. **Stream tool results** back to voice response

## Example Conversation Flow

```
User: "What's the status of the HBAI project?"

Agent thinks: I need to get project details
Agent calls: get_node_context("HBAI Automation")

Tool returns:
{
  "name": "HBAI Automation",
  "type": "PROJECT",
  "status": "Active",
  "health": "Yellow",
  "progress": 65,
  "blockers": ["Waiting on client approval"],
  "this_week_focus": "Complete automation flow"
}

Agent responds: "HBAI is at 65%, yellow status. Blocked on client approval."
```

## Token Budget

| Level | Max Tokens | Description |
|-------|------------|-------------|
| 0 | 200 | Identity/personality |
| 1 | 300 | User + workspace context |
| 2 | 500 | Single node context (on-demand) |
| 3 | 300 | Specific data (tasks, decisions) |
| **Total per turn** | ~1300 | Much less than loading everything |

---

*Plan created: January 17, 2026*
