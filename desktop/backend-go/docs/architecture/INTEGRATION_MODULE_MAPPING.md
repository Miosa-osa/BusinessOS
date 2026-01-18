# Integration Data → Module Mapping

This document maps all external integration data sources to BusinessOS modules, showing where data flows and how it can be used across the platform.

## Integration Matrix

| Integration | Primary Modules | Data Types | Sync Direction |
|-------------|-----------------|------------|----------------|
| **Google Calendar** | `calendar`, `daily_log`, `projects` | Events, meetings, attendees | Bi-directional |
| **Gmail** | `chat`, `daily_log`, `clients` | Emails, threads, attachments | Read + Send |
| **Google Drive** | `contexts`, `projects` | Files, folders, documents | Bi-directional |
| **Google Contacts** | `clients`, `team` | Contacts, contact groups | Bi-directional |
| **Google Tasks** | `tasks`, `projects` | Task lists, tasks | Bi-directional |
| **Slack** | `chat`, `tasks`, `team` | Messages, channels, users | Read + Send |
| **Linear** | `tasks`, `projects` | Issues, projects, cycles | Bi-directional |
| **Notion** | `contexts`, `projects` | Pages, databases, blocks | Bi-directional |
| **HubSpot** | `clients`, `crm`, `projects` | Contacts, companies, deals, activities | Bi-directional |
| **Microsoft 365** | `calendar`, `chat`, `contexts` | Calendar, Mail, OneDrive, Teams | Bi-directional |
| **ClickUp** | `tasks`, `projects` | Tasks, lists, spaces | Bi-directional |
| **Airtable** | `tables`, `projects` | Bases, tables, records | Bi-directional |
| **Fathom** | `daily_log`, `calendar` | Meeting recordings, transcripts, summaries | Read-only |
| **Fireflies** | `daily_log`, `contexts` | Transcripts, action items, summaries | Read-only |

---

## Module-Centric View

### calendar
| Integration | Data Flow |
|-------------|-----------|
| Google Calendar | Events, meetings, attendees |
| Microsoft 365 | Outlook events |
| Fathom | Meeting recordings (linked to events) |

### tasks
| Integration | Data Flow |
|-------------|-----------|
| Linear | Issues, sub-issues, cycles |
| ClickUp | Tasks, checklists |
| Google Tasks | Task lists |
| Slack | /todo messages, reminders |

### projects
| Integration | Data Flow |
|-------------|-----------|
| Linear | Projects, roadmaps, milestones |
| Notion | Project pages, databases |
| ClickUp | Spaces, folders |
| Airtable | Project bases |
| HubSpot | Project/deal association |

### contexts (Knowledge Base)
| Integration | Data Flow |
|-------------|-----------|
| Google Drive | Documents, files |
| Notion | Knowledge pages, wikis |
| Microsoft 365 | OneDrive, SharePoint |
| Fireflies | Meeting transcripts |

### clients (Contact Management)
| Integration | Data Flow |
|-------------|-----------|
| Google Contacts | Personal/business contacts |
| Gmail | Email threads with contacts |
| HubSpot | Contact records |

### crm (Sales Pipeline)
| Integration | Data Flow |
|-------------|-----------|
| HubSpot | Companies, deals, pipelines, activities |
| Salesforce* | Full CRM sync (future) |
| Pipedrive* | Pipeline deals (future) |
| GoHighLevel* | Leads, opportunities (future) |

### daily_log
| Integration | Data Flow |
|-------------|-----------|
| Google Calendar | Today's events |
| Gmail | Important emails |
| Fathom | Today's meeting recordings |
| Fireflies | Meeting summaries |

### chat
| Integration | Data Flow |
|-------------|-----------|
| Slack | Channels, DMs, threads |
| Gmail | Email conversations |
| Microsoft Teams | Team chats |

### team
| Integration | Data Flow |
|-------------|-----------|
| Slack | Workspace members |
| Google Contacts | Team directory |
| Microsoft 365 | Organization directory |

---

## Cross-Module Data Flows

```
                    ┌─────────────────┐
                    │  Integrations   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   calendar    │   │    tasks      │   │   contexts    │
│  (meetings)   │   │   (work)      │   │  (knowledge)  │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └─────────┬─────────┴─────────┬─────────┘
                  │                   │
                  ▼                   ▼
         ┌───────────────┐   ┌───────────────┐
         │  daily_log    │   │   projects    │
         │  (activity)   │   │  (tracking)   │
         └───────────────┘   └───────────────┘
                  │                   │
        ┌─────────┴─────────┬─────────┴─────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   clients     │◄──│      crm      │──►│     team      │
│  (contacts)   │   │   (pipeline)  │   │   (members)   │
└───────────────┘   └───────────────┘   └───────────────┘
```

---

## Integration Data Best Practices

### 1. Single Source of Truth
- Each data type should have ONE primary source
- Other modules reference, not duplicate

### 2. Bi-directional Sync Rules
- Changes in BusinessOS sync back to source
- Conflict resolution: Last-write-wins with timestamp

### 3. Read-Only Integrations
- Fathom, Fireflies: Import only, no export
- Use for enrichment, not as primary storage

### 4. Cross-Module References
- Use entity IDs to link data across modules
- Example: Calendar event links to Client contact

---

## OAuth Scopes by Integration

All integrations are configured with comprehensive scopes. See:
- `/desktop/backend-go/internal/integrations/google/tools.go` - Google tools
- `/desktop/backend-go/internal/integrations/*/provider.go` - Other providers

Last Updated: 2026-01-11
