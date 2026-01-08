# BusinessOS Database Gap Analysis

**Generated:** January 2026
**Total Tables:** 122
**Status:** Comprehensive Review

---

## Executive Summary

BusinessOS has a **robust foundation** with 122 database tables covering most core modules. The database is well-architected with proper relationships, indexes, and constraints.

### Coverage Assessment

| Category | Status | Coverage |
|----------|--------|----------|
| **Core Platform** | Excellent | 95% |
| **Project Management** | Excellent | 95% |
| **CRM/Clients** | Excellent | 95% |
| **AI/Chat** | Excellent | 100% |
| **Integrations** | Excellent | 90% |
| **Documents/Contexts** | Good | 85% |
| **Time Tracking** | Missing | 0% |
| **Invoicing/Billing** | Missing | 0% |
| **Forms/Surveys** | Missing | 0% |
| **Helpdesk/Support** | Missing | 0% |
| **Analytics/Reporting** | Partial | 40% |

---

## Current Database Inventory (122 Tables)

### 1. CORE PLATFORM (15 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `user` | User accounts (managed by Better Auth) | External |
| `session` | User sessions (managed by Better Auth) | External |
| `user_settings` | User preferences, theme, notifications | ✅ |
| `user_commands` | Custom AI commands | ✅ |
| `user_model_preferences` | AI model tier preferences | ✅ |
| `user_integrations` | Connected integrations | ✅ |
| `module_integration_settings` | Per-module integration config | ✅ |
| `credential_vault` | Encrypted credentials | ✅ |
| `system_event_logs` | System activity tracking | ✅ |

### 2. AI/CHAT SYSTEM (12 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `conversations` | Chat conversations | ✅ |
| `messages` | Chat messages | ✅ |
| `conversation_tags` | Conversation categorization | ✅ |
| `artifacts` | Generated code/content | ✅ |
| `artifact_versions` | Artifact version history | ✅ |
| `ai_usage_logs` | AI request tracking | ✅ |
| `mcp_usage_logs` | MCP tool usage | ✅ |
| `usage_daily_summary` | Aggregated usage stats | ✅ |
| `custom_agents` | User-defined AI agents | ✅ |
| `agent_presets` | Built-in agent templates | ✅ |
| `thinking_traces` | Chain-of-thought logs | ✅ |
| `reasoning_templates` | COT templates | ✅ |

### 3. PROJECT MANAGEMENT (12 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `projects` | Projects with full metadata | ✅ |
| `project_notes` | Project notes | ✅ |
| `project_statuses` | Custom status workflows | ✅ |
| `project_members` | Team assignment | ✅ |
| `project_tags` | Project labels | ✅ |
| `project_tag_assignments` | Tag junction | ✅ |
| `project_conversations` | Linked conversations | ✅ |
| `project_documents` | Linked documents | ✅ |
| `project_templates` | Project templates | ✅ |
| `tasks` | Tasks with subtasks | ✅ |
| `task_assignees` | Multi-assignee support | ✅ |
| `task_dependencies` | Task dependencies | ✅ |

### 4. CRM/CLIENTS (14 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `clients` | Contact records | ✅ |
| `client_contacts` | Client contacts | ✅ |
| `client_interactions` | Call/email/meeting logs | ✅ |
| `client_deals` | Simple deal tracking | ✅ |
| `companies` | Organization records | ✅ NEW |
| `contact_company_relations` | Contact-company links | ✅ NEW |
| `pipelines` | Sales/custom pipelines | ✅ NEW |
| `pipeline_stages` | Pipeline stages | ✅ NEW |
| `deals` | Full deal management | ✅ NEW |
| `crm_activities` | CRM activity log | ✅ NEW |
| `deal_stage_history` | Deal funnel analytics | ✅ NEW |

### 5. TEAM/PEOPLE (4 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `team_members` | Team directory | ✅ |
| `team_member_activities` | Team activity log | ✅ |
| `calendar_events` | Team calendar | ✅ |
| `focus_items` | Daily focus tracking | ✅ |

### 6. DOCUMENTS/CONTEXTS (4 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `contexts` | Documents with blocks | ✅ |
| `daily_logs` | Daily journal entries | ✅ |
| `voice_notes` | Voice transcriptions | ✅ |

### 7. NODES SYSTEM (5 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `nodes` | Business nodes/areas | ✅ |
| `node_metrics` | Node health metrics | ✅ |
| `node_projects` | Node-project links | ✅ |
| `node_contexts` | Node-context links | ✅ |
| `node_conversations` | Node-conversation links | ✅ |

### 8. FLEXIBLE TABLES (7 tables) - COMPLETE (NEW)

| Table | Purpose | Status |
|-------|---------|--------|
| `custom_tables` | User-created databases | ✅ NEW |
| `custom_fields` | Table columns | ✅ NEW |
| `custom_field_options` | Select options | ✅ NEW |
| `custom_records` | Table rows | ✅ NEW |
| `custom_views` | Saved views | ✅ NEW |
| `custom_record_history` | Change history | ✅ NEW |
| `custom_workspaces` | Table organization | ✅ NEW |

### 9. UNIVERSAL PATTERNS (6 tables) - COMPLETE (NEW)

| Table | Purpose | Status |
|-------|---------|--------|
| `activity_log` | Universal audit trail | ✅ NEW |
| `attachments` | Universal file attachments | ✅ NEW |
| `attachment_versions` | File versions | ✅ NEW |
| `attachment_folders` | File organization | ✅ NEW |
| `tags` | Universal tagging | ✅ NEW |
| `tag_assignments` | Tag junction | ✅ NEW |
| `entity_links` | Entity relationships | ✅ NEW |

### 10. INTEGRATIONS (50+ tables) - COMPREHENSIVE

#### Integration Framework
| Table | Purpose | Status |
|-------|---------|--------|
| `integration_providers` | Available integrations | ✅ |
| `user_integrations` | User connections | ✅ |
| `integration_sync_log` | Sync history | ✅ |
| `integration_webhooks` | Webhook configs | ✅ |
| `skill_executions` | Sorx skill runs | ✅ |
| `pending_decisions` | Human-in-the-loop | ✅ |

#### Google Workspace (12 tables)
| Table | Status |
|-------|--------|
| `google_oauth_tokens` | ✅ |
| `google_contacts` | ✅ |
| `google_drive_files` | ✅ |
| `google_docs` | ✅ |
| `google_sheets` | ✅ |
| `google_slides` | ✅ |
| `google_task_lists` | ✅ |
| `google_tasks` | ✅ |

#### Microsoft 365 (7 tables)
| Table | Status |
|-------|--------|
| `microsoft_oauth_tokens` | ✅ |
| `microsoft_calendar_events` | ✅ |
| `microsoft_contacts` | ✅ |
| `microsoft_mail_messages` | ✅ |
| `microsoft_onedrive_files` | ✅ |
| `microsoft_todo_lists` | ✅ |
| `microsoft_todo_tasks` | ✅ |

#### Slack
| Table | Status |
|-------|--------|
| `slack_oauth_tokens` | ✅ |
| `slack_channels` | ✅ |
| `slack_messages` | ✅ |

#### Notion
| Table | Status |
|-------|--------|
| `notion_oauth_tokens` | ✅ |
| `notion_databases` | ✅ |
| `notion_pages` | ✅ |

#### Linear
| Table | Status |
|-------|--------|
| `linear_issues` | ✅ |
| `linear_projects` | ✅ |
| `linear_teams` | ✅ |

#### ClickUp
| Table | Status |
|-------|--------|
| `clickup_workspaces` | ✅ |
| `clickup_spaces` | ✅ |
| `clickup_folders` | ✅ |
| `clickup_lists` | ✅ |
| `clickup_tasks` | ✅ |

#### Airtable
| Table | Status |
|-------|--------|
| `airtable_bases` | ✅ |
| `airtable_tables` | ✅ |
| `airtable_records` | ✅ |

#### HubSpot
| Table | Status |
|-------|--------|
| `hubspot_contacts` | ✅ |
| `hubspot_companies` | ✅ |
| `hubspot_deals` | ✅ |

#### Fathom Analytics
| Table | Status |
|-------|--------|
| `fathom_sites` | ✅ |
| `fathom_pages` | ✅ |
| `fathom_events` | ✅ |
| `fathom_aggregations` | ✅ |
| `fathom_referrers` | ✅ |

#### Other
| Table | Status |
|-------|--------|
| `emails` | ✅ |
| `channels` | ✅ |
| `channel_messages` | ✅ |

### 11. DATA IMPORT (4 tables) - COMPLETE

| Table | Purpose | Status |
|-------|---------|--------|
| `import_jobs` | Import job tracking | ✅ |
| `import_mapping_templates` | Field mapping templates | ✅ |
| `imported_records` | Imported data | ✅ |
| `imported_conversations` | Imported conversations | ✅ |
| `data_sync_mappings` | Sync field mappings | ✅ |

---

## GAPS & MISSING MODULES

### GAP 1: TIME TRACKING - HIGH PRIORITY

**Why Needed:** Essential for service businesses, billable hours, productivity tracking.

**Missing Tables:**
```sql
-- Time entries (individual time logs)
time_entries (
    id, user_id, project_id, task_id, client_id,
    description, start_time, end_time, duration_minutes,
    billable, billing_rate, hourly_rate,
    tags, created_at, updated_at
)

-- Time entry categories
time_categories (
    id, user_id, name, color, billable_default,
    created_at
)

-- Weekly timesheets
timesheets (
    id, user_id, week_start, week_end,
    status, -- draft, submitted, approved, rejected
    total_hours, billable_hours,
    submitted_at, approved_by, approved_at,
    notes
)

-- Timer state (for active timers)
active_timers (
    id, user_id, time_entry_id,
    started_at, paused_at, accumulated_seconds
)
```

**Agentic Features:**
- Auto-suggest time entries from calendar
- Smart categorization of work
- "What did you work on today?" summaries

---

### GAP 2: INVOICING & BILLING - HIGH PRIORITY

**Why Needed:** Revenue generation, client billing, financial tracking.

**Missing Tables:**
```sql
-- Invoices
invoices (
    id, user_id, client_id, company_id,
    invoice_number, status, -- draft, sent, paid, overdue, cancelled
    issue_date, due_date, paid_date,
    subtotal, tax_rate, tax_amount, discount, total,
    currency, notes, terms,
    payment_method, payment_reference,
    sent_at, viewed_at, reminder_count,
    created_at, updated_at
)

-- Invoice line items
invoice_items (
    id, invoice_id,
    description, quantity, unit_price, amount,
    tax_rate, taxable,
    project_id, task_id, time_entry_id,
    position
)

-- Recurring invoices
recurring_invoices (
    id, user_id, client_id,
    frequency, -- weekly, monthly, quarterly, yearly
    next_invoice_date, last_invoice_date,
    template_data, auto_send,
    status -- active, paused, cancelled
)

-- Payments
payments (
    id, invoice_id, user_id,
    amount, currency, payment_date,
    payment_method, reference, notes
)

-- Expenses
expenses (
    id, user_id, project_id, client_id,
    description, amount, currency,
    category, receipt_url, billable,
    expense_date, vendor,
    reimbursement_status
)

-- Tax rates
tax_rates (
    id, user_id, name, rate, description,
    is_default, region
)
```

**Agentic Features:**
- Auto-generate invoices from time entries
- Payment reminders
- Expense categorization from receipts

---

### GAP 3: PROPOSALS & QUOTES - MEDIUM PRIORITY

**Why Needed:** Sales process, client agreements, project scoping.

**Missing Tables:**
```sql
-- Proposals
proposals (
    id, user_id, client_id, deal_id,
    title, status, -- draft, sent, viewed, accepted, rejected, expired
    content_blocks, -- Notion-like blocks
    total_value, currency,
    valid_until, sent_at, viewed_at,
    accepted_at, signature_url,
    template_id, version
)

-- Proposal sections
proposal_sections (
    id, proposal_id,
    title, content, position,
    section_type -- intro, scope, pricing, terms, timeline
)

-- Proposal templates
proposal_templates (
    id, user_id, name, description,
    sections_template, default_terms,
    is_default
)

-- Quote line items
quote_items (
    id, proposal_id,
    description, quantity, unit_price, amount,
    optional, discount_percent
)
```

---

### GAP 4: FORMS & SURVEYS - MEDIUM PRIORITY

**Why Needed:** Client intake, feedback collection, data gathering.

**Missing Tables:**
```sql
-- Forms
forms (
    id, user_id,
    title, description, type, -- survey, intake, feedback, quiz
    status, -- draft, active, archived
    settings, -- anonymous, one_response, etc.
    thank_you_message,
    created_at, updated_at
)

-- Form fields
form_fields (
    id, form_id,
    label, field_type, -- text, textarea, select, checkbox, rating, file
    options, required, position,
    validation_rules, placeholder,
    conditional_logic
)

-- Form submissions
form_submissions (
    id, form_id, respondent_id,
    responses, -- JSONB of field_id: value
    submitted_at, ip_address,
    client_id, project_id
)
```

---

### GAP 5: HELPDESK / SUPPORT - LOWER PRIORITY

**Why Needed:** Client support, ticket management, knowledge base.

**Missing Tables:**
```sql
-- Support tickets
tickets (
    id, user_id, client_id,
    subject, description, status, priority,
    category, assignee_id,
    first_response_at, resolved_at,
    satisfaction_rating, satisfaction_feedback,
    created_at, updated_at
)

-- Ticket messages
ticket_messages (
    id, ticket_id, sender_id,
    content, is_internal, -- internal note vs client-visible
    attachments,
    created_at
)

-- Knowledge base articles
kb_articles (
    id, user_id,
    title, content, slug,
    category_id, status, -- draft, published
    view_count, helpful_count,
    created_at, updated_at
)

-- KB categories
kb_categories (
    id, user_id, name, slug, parent_id,
    position, icon
)
```

---

### GAP 6: REPORTING / DASHBOARDS - PARTIAL

**What We Have:**
- `usage_daily_summary` - AI usage stats
- `ai_usage_logs` - Request-level tracking
- `deal_stage_history` - CRM funnel analytics

**Missing:**
```sql
-- Saved reports
saved_reports (
    id, user_id,
    name, description, report_type,
    query_config, -- filters, grouping, date range
    visualization_type, -- table, chart, pie, funnel
    schedule, -- null for manual, or cron
    recipients,
    is_public
)

-- Dashboard widgets
dashboard_widgets (
    id, user_id,
    widget_type, title,
    data_source, config,
    position_x, position_y, width, height
)

-- Report snapshots (historical data)
report_snapshots (
    id, report_id,
    snapshot_date, data,
    created_at
)
```

---

### GAP 7: GOALS / OKRs - LOWER PRIORITY

**Missing Tables:**
```sql
-- Goals
goals (
    id, user_id,
    title, description, type, -- goal, objective, key_result
    parent_id, -- for OKR hierarchy
    target_value, current_value, unit,
    start_date, end_date,
    status, -- on_track, at_risk, behind, achieved
    owner_id, project_id,
    created_at, updated_at
)

-- Goal check-ins
goal_checkins (
    id, goal_id, user_id,
    value, notes, confidence_level,
    checked_in_at
)
```

---

### GAP 8: NOTIFICATIONS - PARTIAL

**What We Have:**
- System event logs
- No dedicated notification system

**Missing:**
```sql
-- Notifications
notifications (
    id, user_id,
    type, title, message,
    entity_type, entity_id,
    read_at, clicked_at,
    action_url, action_label,
    priority, -- normal, important, urgent
    channels, -- in_app, email, push
    created_at
)

-- Notification preferences
notification_preferences (
    id, user_id,
    notification_type,
    in_app_enabled, email_enabled, push_enabled,
    digest_frequency -- immediate, hourly, daily
)
```

---

## PRIORITY RECOMMENDATIONS

### Phase 1: Revenue Critical (Implement Next)
1. **Time Tracking** - 4 tables
2. **Invoicing** - 6 tables
3. **Notifications** - 2 tables

### Phase 2: Sales Enhancement
4. **Proposals/Quotes** - 4 tables
5. **Goals/OKRs** - 2 tables

### Phase 3: Client Experience
6. **Forms/Surveys** - 3 tables
7. **Helpdesk** - 4 tables

### Phase 4: Analytics
8. **Reporting** - 3 tables

---

## SUMMARY

### What We Have (Strong)
- Complete AI/Chat system
- Full project management
- Comprehensive CRM
- 50+ integration tables
- Flexible tables (Airtable-like)
- Universal patterns (tags, attachments, activity log, entity links)

### What's Missing (Gaps)
1. **Time Tracking** - HIGH priority for service businesses
2. **Invoicing/Billing** - HIGH priority for revenue
3. **Proposals** - MEDIUM priority for sales
4. **Forms** - MEDIUM priority for data collection
5. **Helpdesk** - LOWER priority
6. **Reporting** - PARTIAL, needs enhancement
7. **Notifications** - PARTIAL, needs dedicated system

### Total Gap: ~28 tables needed for complete coverage
