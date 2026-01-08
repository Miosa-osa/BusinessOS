# Dashboard MCP Tool - Implementation Spec

**Created:** January 4, 2026
**Updated:** January 6, 2026
**Status:** Ready for Implementation
**Protocol:** MCP (Model Context Protocol)
**File to Create:** `desktop/backend-go/internal/tools/dashboard_tool.go`

---

## Overview

This document defines the **MCP Tool** for dashboard configuration. This is the **executable layer** - actual Go code that performs database operations.

> **Important:** This tool is called BY the agent. The agent learns WHEN to call this tool from the **dashboard-management skill** (see `skills/dashboard-management/SKILL.md`).

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TWO-LAYER ARCHITECTURE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  SKILL LAYER (Instructions)              TOOL LAYER (Execution)             │
│  ─────────────────────────               ──────────────────────             │
│                                                                              │
│  skills/dashboard-management/            internal/tools/dashboard_tool.go   │
│  └── SKILL.md                            └── ConfigureDashboardTool         │
│      │                                       │                              │
│      │ "When user says 'show tasks',         │ func Execute(input) {        │
│      │  call configure_dashboard with        │   // Validate                │
│      │  action=add_widget..."                │   // Query DB                │
│      │                                       │   // Return result           │
│                                              │ }                            │
│      ▼                                       ▼                              │
│  Agent reads this to DECIDE            Agent CALLS this to ACT              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## MCP Tool Definition

### Tool Name
```
configure_dashboard
```

### Tool Description (for MCP registration)
```
Configure user dashboards by creating dashboards, adding/removing/updating widgets,
and setting defaults. Returns the updated dashboard state and available options.
```

---

## Input Schema (JSON Schema for MCP)

```json
{
  "type": "object",
  "properties": {
    "action": {
      "type": "string",
      "enum": [
        "create_dashboard",
        "add_widget",
        "add_widgets",
        "remove_widget",
        "update_widget",
        "set_default",
        "list_dashboards",
        "list_widgets",
        "get_dashboard"
      ],
      "description": "The action to perform"
    },
    "dashboard_id": {
      "type": "string",
      "description": "Target dashboard UUID. Uses default if not provided."
    },
    "dashboard_name": {
      "type": "string",
      "description": "Target dashboard by name (alternative to dashboard_id)"
    },
    "name": {
      "type": "string",
      "description": "For create_dashboard: the name of the new dashboard"
    },
    "description": {
      "type": "string",
      "description": "For create_dashboard: optional description"
    },
    "widget_type": {
      "type": "string",
      "enum": [
        "task_summary",
        "task_list",
        "project_progress",
        "upcoming_deadlines",
        "recent_activity",
        "metric_card",
        "task_burndown",
        "workload_heatmap",
        "client_overview",
        "notes_pinned",
        "quick_actions",
        "agent_shortcuts"
      ],
      "description": "For add_widget: the type of widget"
    },
    "widget_id": {
      "type": "string",
      "description": "For remove_widget/update_widget: the widget instance ID"
    },
    "config": {
      "type": "object",
      "description": "Widget configuration options",
      "properties": {
        "group_by": {
          "type": "string",
          "enum": ["status", "project", "priority"]
        },
        "days_ahead": {
          "type": "integer",
          "minimum": 1,
          "maximum": 90
        },
        "show_overdue": {
          "type": "boolean"
        },
        "limit": {
          "type": "integer",
          "minimum": 5,
          "maximum": 50
        },
        "project_id": {
          "type": "string"
        },
        "days": {
          "type": "integer",
          "minimum": 7,
          "maximum": 365
        },
        "metric": {
          "type": "string",
          "enum": ["tasks_due_today", "tasks_overdue", "tasks_completed_week", "active_projects"]
        },
        "date_range": {
          "type": "string",
          "enum": ["month", "quarter", "year"]
        }
      }
    },
    "widgets": {
      "type": "array",
      "description": "For add_widgets: batch add multiple widgets",
      "items": {
        "type": "object",
        "properties": {
          "type": { "type": "string" },
          "config": { "type": "object" }
        },
        "required": ["type"]
      }
    }
  },
  "required": ["action"]
}
```

---

## Actions Reference

### Read Actions (No Side Effects)

| Action | Purpose | Required Params |
|--------|---------|-----------------|
| `list_dashboards` | Get user's dashboards | none |
| `list_widgets` | Get available widget types | none |
| `get_dashboard` | Get dashboard layout | dashboard_id or name (optional) |

### Write Actions (Modify State)

| Action | Purpose | Required Params |
|--------|---------|-----------------|
| `create_dashboard` | Create new dashboard | `name` |
| `add_widget` | Add single widget | `widget_type` |
| `add_widgets` | Add multiple widgets | `widgets[]` |
| `remove_widget` | Remove a widget | `widget_id` |
| `update_widget` | Update widget config | `widget_id`, `config` |
| `set_default` | Set default dashboard | `dashboard_id` or `dashboard_name` |

---

## Response Format

All responses include context for agent decision-making:

### Success Response
```json
{
  "success": true,
  "message": "Added task summary widget to 'My Day' dashboard",
  "dashboard": {
    "id": "uuid",
    "name": "My Day",
    "widget_count": 4,
    "widgets": [
      { "id": "w1", "type": "task_summary", "position": {"x":0,"y":0,"w":4,"h":3} },
      { "id": "w2", "type": "upcoming_deadlines", "position": {"x":4,"y":0,"w":4,"h":3} }
    ]
  },
  "context": {
    "user_dashboards": [
      { "id": "uuid", "name": "My Day", "is_default": true, "widget_count": 4 }
    ],
    "available_widgets": [
      { "type": "task_summary", "name": "Task Summary", "category": "tasks" }
    ]
  }
}
```

### Error Response (with suggestions)
```json
{
  "success": false,
  "error": "Widget type 'team_workload' is not available",
  "suggestion": "Try 'workload_heatmap' to see task density over time.",
  "available_alternatives": ["workload_heatmap", "task_summary"]
}
```

---

## Design Decisions

| Decision | Answer | Rationale |
|----------|--------|-----------|
| Delete dashboards? | No, widgets only | Too risky - user might lose work |
| Config complexity | Common options only | Keep simple - filters, time ranges, limits |
| Auto-place algorithm | Left-to-right, row filling | Predictable, covers 90% of cases |
| Error handling | Suggest alternatives | Be helpful, not just "widget not found" |
| Multi-action | Yes via `add_widgets` | "Add tasks AND deadlines" = one call |

---

## Auto-Place Algorithm

When adding widgets without explicit position:

```go
func findNextPosition(layout []Widget, newWidget WidgetType) Position {
    const gridCols = 12

    // Get default size for widget type
    size := getDefaultSize(newWidget) // e.g., {w: 4, h: 3}

    // Build occupancy grid from existing widgets
    occupied := buildOccupancyGrid(layout)

    // Scan left-to-right, top-to-bottom for first fit
    for y := 0; y < 100; y++ {
        for x := 0; x <= gridCols - size.W; x++ {
            if canPlace(occupied, x, y, size.W, size.H) {
                return Position{X: x, Y: y, W: size.W, H: size.H}
            }
        }
    }

    // Fallback: place at bottom
    maxY := getMaxY(layout)
    return Position{X: 0, Y: maxY + 1, W: size.W, H: size.H}
}
```

---

## Go Implementation Structure

```go
// internal/tools/dashboard_tool.go

package tools

import (
    "context"
    "encoding/json"
)

// ConfigureDashboardTool implements the MCP tool interface
type ConfigureDashboardTool struct {
    queries *sqlc.Queries
    pool    *pgxpool.Pool
}

// Name returns the MCP tool name
func (t *ConfigureDashboardTool) Name() string {
    return "configure_dashboard"
}

// Description returns the MCP tool description
func (t *ConfigureDashboardTool) Description() string {
    return "Configure user dashboards by creating dashboards, adding/removing/updating widgets, and setting defaults."
}

// InputSchema returns the JSON Schema for input validation
func (t *ConfigureDashboardTool) InputSchema() json.RawMessage {
    // Return the schema defined above
}

// Execute runs the tool with the given input
func (t *ConfigureDashboardTool) Execute(ctx context.Context, userID string, input json.RawMessage) (json.RawMessage, error) {
    var params DashboardToolInput
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, err
    }

    switch params.Action {
    case "list_dashboards":
        return t.listDashboards(ctx, userID)
    case "create_dashboard":
        return t.createDashboard(ctx, userID, params)
    case "add_widget":
        return t.addWidget(ctx, userID, params)
    // ... other actions
    }
}
```

---

## MCP Registration

Register the tool in your MCP server:

```go
// internal/mcp/server.go

func (s *MCPServer) RegisterTools() {
    s.RegisterTool(&ConfigureDashboardTool{
        queries: s.queries,
        pool:    s.pool,
    })
}
```

---

## Implementation Checklist

- [ ] Create `internal/tools/dashboard_tool.go`
- [ ] Implement `ConfigureDashboardTool` struct
- [ ] Implement each action handler:
  - [ ] `list_dashboards`
  - [ ] `list_widgets`
  - [ ] `get_dashboard`
  - [ ] `create_dashboard`
  - [ ] `add_widget`
  - [ ] `add_widgets`
  - [ ] `remove_widget`
  - [ ] `update_widget`
  - [ ] `set_default`
- [ ] Implement `findNextPosition()` auto-place algorithm
- [ ] Add widget type validation with suggestions
- [ ] Add response context builder
- [ ] Register tool in MCP server
- [ ] Write unit tests

---

## Related Documents

| Document | Purpose |
|----------|---------|
| `AGENT_SKILLS_OVERVIEW.md` | Understanding skills vs tools architecture |
| `skills/dashboard-management/SKILL.md` | Agent instructions for WHEN to call this tool |
| `DASHBOARD_OVERVIEW.md` | Dashboard system architecture |

---

## Notes

- This tool is the **execution layer** - it does the actual work
- The **skill layer** (`SKILL.md`) teaches the agent when/how to call this tool
- Tool does NOT broadcast SSE events - handlers do that when layout is saved
- Use existing SQLC queries, not direct SQL
- Keep config options simple - power users can use the API directly
