# Dashboard & Analytics Schema Review

**Created:** January 7, 2026  
**Author:** Javaris Tavel  
**Status:** Ready for Manager Review  
**Last Updated:** January 7, 2026

---

## Executive Summary

This document provides a comprehensive review of our current dashboard and analytics database schema, including what's implemented, what's working well, and recommended improvements.

### Quick Assessment

| Area | Current State | Status |
|------|---------------|--------|
| Core Dashboard Schema | Implemented | ✅ Good |
| Widget Type Registry | Implemented | ✅ Good |
| Dashboard Templates | Implemented | ✅ Good |
| Analytics Queries | Implemented | ✅ Good |
| Sharing & Visibility | Implemented | ⚠️ Needs Enhancement |
| Workspace Integration | Partial | 🔶 Pending |
| Historical Analytics | Not Implemented | ❌ Missing |

---

## Current Schema Overview

### 1. User Dashboards Table (`user_dashboards`)

**Location:** [schema.sql#L1214-L1234](desktop/backend-go/internal/database/schema.sql#L1214-L1234)

```sql
CREATE TABLE user_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    
    layout JSONB NOT NULL DEFAULT '[]',  -- Widget configurations
    
    visibility VARCHAR(50) DEFAULT 'private',
    share_token VARCHAR(100) UNIQUE,
    is_enforced BOOLEAN DEFAULT FALSE,
    enforced_for_roles TEXT[],
    
    created_via VARCHAR(50) DEFAULT 'agent',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Assessment:** ✅ Solid foundation

**What's Working:**
- JSONB layout storage allows flexible widget configurations
- Visibility controls for private/workspace/public sharing
- `is_enforced` and `enforced_for_roles` supports role-based default dashboards
- `created_via` tracks whether agent, manual, or template created it

**Indexes Created:**
- `idx_dashboards_user` - User lookup ✅
- `idx_dashboards_workspace` - Workspace filtering ✅
- `idx_dashboards_share_token` - Public link access ✅
- `idx_dashboards_default` - Quick default dashboard lookup ✅

---

### 2. Widget Type Registry (`dashboard_widgets`)

**Location:** [schema.sql#L1242-L1260](desktop/backend-go/internal/database/schema.sql#L1242-L1260)

```sql
CREATE TABLE dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_type VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    config_schema JSONB NOT NULL,       -- JSON Schema for validation
    default_config JSONB DEFAULT '{}',
    default_size JSONB DEFAULT '{"w": 4, "h": 3}',
    min_size JSONB DEFAULT '{"w": 2, "h": 2}',
    
    sse_events TEXT[],                  -- Real-time event subscriptions
    
    is_enabled BOOLEAN DEFAULT TRUE,
    requires_feature VARCHAR(100),      -- Feature flag dependency
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Assessment:** ✅ Well designed

**Current Widget Types (17 total):**

| Category | Enabled Widgets | Disabled (Future) |
|----------|-----------------|-------------------|
| tasks | task_summary, task_list, upcoming_deadlines | task_calendar |
| projects | project_progress | project_timeline |
| analytics | metric_card, task_burndown, workload_heatmap | |
| activity | recent_activity | |
| clients | client_overview | client_pipeline, client_activity |
| notes | notes_pinned | |
| utility | quick_actions, agent_shortcuts | |
| team | | team_capacity, team_workload |
| advanced | | custom_query |

---

### 🏆 Top-Tier Widget Recommendations

These are the **must-have widgets** that provide the highest value for users:

#### Tier 1: Essential (Every User Should Have)

| Widget | Why It's Critical | Best Config |
|--------|-------------------|-------------|
| **task_summary** | At-a-glance view of all work status. Answers "Where am I at?" instantly. | `group_by: "status"` |
| **upcoming_deadlines** | Prevents missed deadlines. #1 productivity killer when ignored. | `days_ahead: 7, show_overdue: true` |
| **metric_card** (x4) | Executive-level KPIs. Quick health check in seconds. | Use all 4 metrics together |

**Recommended "Core 4" Layout:**
```
┌─────────────────────────────────────────────────────────────────┐
│ [Due Today: 5] [Overdue: 2] [Done This Week: 12] [Projects: 3] │  ← 4 metric_cards
├─────────────────────────────────┬───────────────────────────────┤
│                                 │                               │
│      TASK SUMMARY               │     UPCOMING DEADLINES        │
│      (by status)                │     (next 7 days)             │
│                                 │                               │
└─────────────────────────────────┴───────────────────────────────┘
```

#### Tier 2: Power User (High Productivity Impact)

| Widget | Why It Matters | Best For |
|--------|----------------|----------|
| **task_burndown** | Visualizes velocity. Shows if you're on track or falling behind. | Project managers, sprint planning |
| **project_progress** | Multi-project visibility. Critical when juggling 3+ projects. | Managers, executives |
| **workload_heatmap** | Spot overloaded days before they happen. Capacity planning. | Anyone with variable workloads |
| **recent_activity** | Audit trail. "What changed?" at a glance. | Team leads, auditors |

**Recommended "Manager View" Layout:**
```
┌─────────────────────────────────┬───────────────────────────────┐
│      PROJECT PROGRESS           │       TASK BURNDOWN           │
│      (all active)               │       (30 days)               │
├─────────────────────────────────┼───────────────────────────────┤
│      UPCOMING DEADLINES         │     WORKLOAD HEATMAP          │
│      (14 days)                  │     (this month)              │
└─────────────────────────────────┴───────────────────────────────┘
```

#### Tier 3: Workflow Accelerators

| Widget | Use Case | Value |
|--------|----------|-------|
| **quick_actions** | One-click task/note creation | Reduces friction |
| **agent_shortcuts** | Pre-built prompts for common questions | Agent adoption boost |
| **notes_pinned** | Keep reference docs visible | Context switching reduction |
| **task_list** | Deep-dive filterable view | Power users who need details |
| **client_overview** | Client-facing roles only | Account managers |

#### Tier 4: Future High-Value (Currently Disabled)

| Widget | When to Enable | Dependency |
|--------|----------------|------------|
| **team_workload** | Multi-user workspaces launch | Team/Collaboration feature |
| **team_capacity** | Resource planning needed | Team/Collaboration feature |
| **task_calendar** | Users request calendar view | Frontend calendar component |
| **project_timeline** | Gantt-style planning needed | Complex visualization work |

---

### 📊 Recommended Default Dashboards by Role

| Role | Recommended Widgets | Template |
|------|---------------------|----------|
| **Individual Contributor** | task_summary, upcoming_deadlines, quick_actions, notes_pinned | "My Day" |
| **Project Manager** | project_progress, task_burndown, upcoming_deadlines, workload_heatmap | "Project Manager" |
| **Executive/Leadership** | 4x metric_card, project_progress, task_burndown | "Executive" |
| **Developer** | task_list, notes_pinned, recent_activity, agent_shortcuts | "Developer" |
| **Account Manager** | client_overview, project_progress, upcoming_deadlines | "Client Focus" |

---

### 🎯 Widget Priority Matrix

```
                    HIGH IMPACT
                        │
    ┌───────────────────┼───────────────────┐
    │                   │                   │
    │   task_summary    │   metric_card     │
    │   upcoming_       │   task_burndown   │
    │   deadlines       │   project_        │
    │                   │   progress        │
    │                   │                   │
LOW ├───────────────────┼───────────────────┤ HIGH
EFFORT                  │                   EFFORT
    │                   │                   │
    │   quick_actions   │   workload_       │
    │   agent_shortcuts │   heatmap         │
    │   notes_pinned    │   task_calendar   │
    │                   │   (future)        │
    │                   │                   │
    └───────────────────┼───────────────────┘
                        │
                    LOW IMPACT
```

**Recommendation:** Focus frontend polish on Tier 1 widgets first. They cover 80% of user needs.

---

### 3. Dashboard Templates (`dashboard_templates`)

**Location:** [schema.sql#L1263-L1278](desktop/backend-go/internal/database/schema.sql#L1263-L1278)

```sql
CREATE TABLE dashboard_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    layout JSONB NOT NULL,
    
    thumbnail_url TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Seeded Templates (5):**
1. **My Day** - Personal daily planning (default)
2. **Project Manager** - Progress tracking with burndown charts
3. **Executive** - KPI metrics overview
4. **Developer** - Task-focused with notes
5. **Client Focus** - Client relationship management

---

### 4. Analytics Queries

**Location:** [dashboards.sql#L207-L272](desktop/backend-go/internal/database/queries/dashboards.sql#L207-L272)

**Implemented Analytics:**

| Query | Purpose | Widget |
|-------|---------|--------|
| `CountTasksDueToday` | KPI metric | metric_card |
| `CountTasksOverdue` | KPI metric | metric_card |
| `CountTasksCompletedThisWeek` | KPI metric | metric_card |
| `CountActiveProjects` | KPI metric | metric_card |
| `GetTaskBurndownData` | Time-series chart | task_burndown |
| `GetWorkloadHeatmapData` | Calendar density | workload_heatmap |
| `GetUpcomingTasksDueByDate` | Deadline grouping | upcoming_deadlines |

**API Endpoints Implemented:**
- `GET /api/analytics/summary` - All KPI metrics in one call
- `GET /api/analytics/burndown?days=30&project_id=xxx` - Burndown data
- `GET /api/analytics/workload?start=YYYY-MM-DD&end=YYYY-MM-DD` - Heatmap data
- `GET /api/analytics/deadlines?days=7` - Upcoming deadlines

---

## Recommended Improvements

### 🔴 High Priority

#### 1. Add Historical Snapshots Table

**Problem:** No way to track dashboard metrics over time for trend analysis.

**Proposed Schema:**
```sql
CREATE TABLE analytics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Snapshot date (one per day per user)
    snapshot_date DATE NOT NULL,
    
    -- Metrics captured at snapshot time
    metrics JSONB NOT NULL DEFAULT '{}',
    -- Example: {
    --   "tasks_total": 50,
    --   "tasks_completed": 30,
    --   "tasks_overdue": 5,
    --   "projects_active": 3,
    --   "avg_task_completion_days": 2.5
    -- }
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, snapshot_date)
);

CREATE INDEX idx_analytics_snapshots_user_date ON analytics_snapshots(user_id, snapshot_date DESC);
CREATE INDEX idx_analytics_snapshots_workspace ON analytics_snapshots(workspace_id);
```

**Benefits:**
- Week-over-week / month-over-month comparisons
- Trend charts (e.g., "Tasks completed trending up")
- Executive reports with historical context

**Estimated Effort:** 4-6 hours

---

#### 2. Add Widget Instance Tracking Table

**Problem:** Currently widgets are stored inline in JSONB. No way to track widget-level analytics (e.g., "Which widgets are most used?").

**Proposed Schema:**
```sql
CREATE TABLE dashboard_widget_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES user_dashboards(id) ON DELETE CASCADE,
    widget_type VARCHAR(100) NOT NULL REFERENCES dashboard_widgets(widget_type),
    
    -- Instance-specific config (overrides widget defaults)
    config JSONB DEFAULT '{}',
    
    -- Grid position
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 4,
    height INTEGER NOT NULL DEFAULT 3,
    
    -- Usage tracking
    last_interacted_at TIMESTAMPTZ,
    interaction_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_widget_instances_dashboard ON dashboard_widget_instances(dashboard_id);
CREATE INDEX idx_widget_instances_type ON dashboard_widget_instances(widget_type);
```

**Benefits:**
- Normalized widget storage (cleaner than JSONB array)
- Widget usage analytics ("Most popular widgets")
- Easier widget-level queries and updates
- Better foreign key relationships

**Trade-off:** Migration complexity from current JSONB layout

**Estimated Effort:** 8-12 hours (including migration)

---

#### 3. Add Dashboard View Tracking

**Problem:** No way to know which dashboards are actually being used.

**Proposed Schema:**
```sql
CREATE TABLE dashboard_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES user_dashboards(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    
    viewed_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized for faster analytics
    duration_seconds INTEGER,
    widget_interactions JSONB DEFAULT '[]'
);

CREATE INDEX idx_dashboard_views_dashboard ON dashboard_views(dashboard_id, viewed_at DESC);
CREATE INDEX idx_dashboard_views_user ON dashboard_views(user_id, viewed_at DESC);

-- Cleanup old entries (keep 90 days)
-- Add cron job: DELETE FROM dashboard_views WHERE viewed_at < NOW() - INTERVAL '90 days';
```

**Benefits:**
- Know which dashboards are valuable
- Identify unused dashboards for cleanup prompts
- Data for "Popular dashboards in your workspace" feature

**Estimated Effort:** 4-6 hours

---

### 🟡 Medium Priority

#### 4. Enhance Sharing Capabilities

**Current State:**
- `visibility` column supports: private, workspace, public_link
- `share_token` for public links

**Proposed Enhancements:**
```sql
-- Add to user_dashboards table
ALTER TABLE user_dashboards ADD COLUMN IF NOT EXISTS
    shared_with UUID[] DEFAULT '{}';  -- Specific user IDs

-- Add granular share permissions table
CREATE TABLE dashboard_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES user_dashboards(id) ON DELETE CASCADE,
    shared_with_user_id VARCHAR(255),
    shared_with_role VARCHAR(100),
    
    permission VARCHAR(20) DEFAULT 'view' CHECK (permission IN ('view', 'edit', 'admin')),
    
    expires_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    
    UNIQUE(dashboard_id, shared_with_user_id),
    UNIQUE(dashboard_id, shared_with_role)
);

CREATE INDEX idx_dashboard_shares_dashboard ON dashboard_shares(dashboard_id);
CREATE INDEX idx_dashboard_shares_user ON dashboard_shares(shared_with_user_id);
```

**Benefits:**
- Share with specific users (not just workspace-wide)
- Role-based sharing (e.g., "All Managers can view")
- Edit vs view permissions
- Expiring share links

**Estimated Effort:** 6-8 hours

---

#### 5. Add Widget Data Caching Table

**Problem:** Some widget queries can be expensive (burndown with 365 days of data).

**Proposed Schema:**
```sql
CREATE TABLE widget_data_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    widget_type VARCHAR(100) NOT NULL,
    
    -- Cache key (hashed config + date range)
    cache_key VARCHAR(255) NOT NULL,
    
    -- Cached response
    data JSONB NOT NULL,
    
    -- TTL
    expires_at TIMESTAMPTZ NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, widget_type, cache_key)
);

CREATE INDEX idx_widget_cache_lookup ON widget_data_cache(user_id, widget_type, cache_key);
CREATE INDEX idx_widget_cache_expiry ON widget_data_cache(expires_at);
```

**Benefits:**
- Faster dashboard loads
- Reduced database load
- Background refresh possible

**Estimated Effort:** 4-6 hours

---

### 🟢 Low Priority (Future)

#### 6. Custom Metrics Table

For power users who want to define their own KPIs:

```sql
CREATE TABLE custom_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Aggregation definition
    entity_type VARCHAR(50) NOT NULL, -- 'task', 'project', 'client'
    aggregation VARCHAR(20) NOT NULL, -- 'count', 'sum', 'avg'
    filter_conditions JSONB DEFAULT '{}',
    
    -- Display
    format VARCHAR(20) DEFAULT 'number', -- 'number', 'percentage', 'currency', 'duration'
    icon VARCHAR(50),
    color VARCHAR(20),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Estimated Effort:** 12-16 hours

---

## Layout JSON Structure

Current format stored in `user_dashboards.layout`:

```json
[
  {
    "widget_id": "w1-uuid",
    "widget_type": "task_summary",
    "position": {
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 3
    },
    "config": {
      "group_by": "status",
      "show_completed": false
    }
  },
  {
    "widget_id": "w2-uuid",
    "widget_type": "upcoming_deadlines",
    "position": {
      "x": 4,
      "y": 0,
      "w": 4,
      "h": 3
    },
    "config": {
      "days_ahead": 7,
      "show_overdue": true
    }
  }
]
```

**Assessment:** ✅ This is a valid approach (similar to Grafana, Notion)

**Pros:**
- Single query to load full dashboard
- Easy to serialize/deserialize
- Frontend-friendly format

**Cons:**
- Harder to query individual widgets
- No referential integrity on widget types
- Large dashboards = large JSON blobs

---

## API Coverage

### Currently Implemented

| Endpoint | Handler | Status |
|----------|---------|--------|
| `GET /api/dashboards` | ListUserDashboards | ✅ |
| `GET /api/dashboards/:id` | GetUserDashboard | ✅ |
| `POST /api/dashboards` | CreateUserDashboard | ✅ |
| `PUT /api/dashboards/:id` | UpdateUserDashboard | ✅ |
| `DELETE /api/dashboards/:id` | DeleteUserDashboard | ✅ |
| `POST /api/dashboards/:id/duplicate` | DuplicateDashboard | ✅ |
| `PUT /api/dashboards/:id/default` | SetDefaultDashboard | ✅ |
| `PUT /api/dashboards/:id/share` | UpdateShareToken | ✅ |
| `GET /api/dashboards/shared/:token` | GetSharedDashboard | ✅ |
| `GET /api/dashboards/templates` | ListDashboardTemplates | ✅ |
| `POST /api/dashboards/from-template/:id` | CreateFromTemplate | ✅ |
| `GET /api/widgets` | ListWidgetTypes | ✅ |
| `GET /api/analytics/summary` | GetAnalyticsSummary | ✅ |
| `GET /api/analytics/burndown` | GetTaskBurndown | ✅ |
| `GET /api/analytics/workload` | GetWorkloadHeatmap | ✅ |
| `GET /api/analytics/deadlines` | GetUpcomingDeadlines | ✅ |

### Missing/Recommended

| Endpoint | Purpose | Priority |
|----------|---------|----------|
| `GET /api/dashboards/workspace/:id` | List workspace dashboards | Medium |
| `POST /api/analytics/snapshot` | Trigger manual snapshot | Low |
| `GET /api/analytics/trends` | Historical trend data | High |
| `GET /api/widgets/:type/preview` | Widget preview data | Low |

---

## Migration Path

If improvements are approved, here's the recommended order:

### Phase 1 (Week 1)
1. Add `analytics_snapshots` table
2. Create background job to snapshot daily
3. Add `GET /api/analytics/trends` endpoint

### Phase 2 (Week 2)
1. Add `dashboard_views` tracking table
2. Implement view tracking middleware
3. Add dashboard analytics API

### Phase 3 (Week 3-4)
1. Add `dashboard_shares` table
2. Migrate sharing logic
3. Add granular permission checks

### Phase 4 (Optional)
1. Consider `dashboard_widget_instances` migration
2. Add widget caching layer
3. Custom metrics (if demand exists)

---

## Appendix: Current File Locations

| File | Purpose |
|------|---------|
| [schema.sql](desktop/backend-go/internal/database/schema.sql) | Master schema definition |
| [017_dashboards.sql](desktop/backend-go/internal/database/migrations/017_dashboards.sql) | Dashboard migration with seed data |
| [dashboards.sql](desktop/backend-go/internal/database/queries/dashboards.sql) | SQLC queries |
| [dashboard_handlers.go](desktop/backend-go/internal/handlers/dashboard_handlers.go) | CRUD handlers |
| [analytics_handlers.go](desktop/backend-go/internal/handlers/analytics_handlers.go) | Analytics endpoints |

---

## Summary

**Overall Assessment:** The current dashboard and analytics schema is well-designed for our current needs. The JSONB layout approach is flexible and performant for typical dashboard sizes.

**Key Gaps:**
1. No historical analytics snapshots (can't show trends)
2. No dashboard usage tracking (don't know what's valuable)
3. Limited sharing granularity (no user-specific shares)

**Recommended Next Steps:**
1. ✅ Approve `analytics_snapshots` table for Phase 1
2. Review sharing requirements with product
3. Decide on widget instance normalization trade-offs

---

*This document is ready for manager review. Please provide feedback or approval for the proposed improvements.*
