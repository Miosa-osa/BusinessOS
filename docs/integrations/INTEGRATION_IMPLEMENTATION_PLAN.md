# Integration Infrastructure - Complete Implementation Plan

## Executive Summary

This is the **COMPLETE** integration architecture for BusinessOS. Every file, every endpoint, every component is mapped here. The team should use this as the definitive reference.

**Total Scope:**
- 30+ Integration Providers
- 50+ MCP Tools
- 100+ API Endpoints
- 50+ Frontend Components
- 20+ Database Tables

---

## Complete File Tree Structure

```
backend-go/
├── internal/
│   ├── integrations/
│   │   ├── types.go                          # Core interfaces (Provider, Token, Status)
│   │   ├── registry.go                       # Provider registry (Register, Get, List)
│   │   ├── oauth.go                          # Generic OAuth 2.0 helpers
│   │   ├── errors.go                         # Integration-specific errors
│   │   ├── config.go                         # Environment config loader
│   │   │
│   │   ├── google/                           # Google Ecosystem
│   │   │   ├── provider.go                   # Google provider implementation
│   │   │   ├── oauth.go                      # Google OAuth (shared across services)
│   │   │   ├── calendar.go                   # Google Calendar service [EXISTS]
│   │   │   ├── calendar_sync.go              # Calendar sync worker
│   │   │   ├── gmail.go                      # Gmail service [NEW]
│   │   │   ├── gmail_sync.go                 # Gmail sync worker
│   │   │   ├── drive.go                      # Google Drive service [NEW]
│   │   │   ├── drive_sync.go                 # Drive sync worker
│   │   │   ├── contacts.go                   # Google Contacts service [NEW]
│   │   │   ├── meet.go                       # Google Meet service [NEW]
│   │   │   └── workspace.go                  # Google Workspace directory [NEW]
│   │   │
│   │   ├── microsoft/                        # Microsoft Ecosystem
│   │   │   ├── provider.go                   # Microsoft provider
│   │   │   ├── oauth.go                      # Microsoft Graph OAuth
│   │   │   ├── outlook.go                    # Outlook Calendar [NEW]
│   │   │   ├── outlook_sync.go               # Outlook sync worker
│   │   │   ├── teams.go                      # Microsoft Teams [NEW]
│   │   │   ├── teams_sync.go                 # Teams sync worker
│   │   │   ├── onedrive.go                   # OneDrive [NEW]
│   │   │   ├── todo.go                       # Microsoft To Do [NEW]
│   │   │   └── azure_ad.go                   # Azure AD directory [NEW]
│   │   │
│   │   ├── slack/                            # Slack
│   │   │   ├── provider.go                   # Slack provider [EXISTS]
│   │   │   ├── oauth.go                      # Slack OAuth [EXISTS]
│   │   │   ├── channels.go                   # Channel operations [EXISTS]
│   │   │   ├── messages.go                   # Message operations [EXISTS]
│   │   │   ├── users.go                      # User operations [EXISTS]
│   │   │   ├── sync.go                       # Slack sync worker [NEW]
│   │   │   └── webhooks.go                   # Incoming webhooks [NEW]
│   │   │
│   │   ├── notion/                           # Notion
│   │   │   ├── provider.go                   # Notion provider [EXISTS]
│   │   │   ├── oauth.go                      # Notion OAuth [EXISTS]
│   │   │   ├── databases.go                  # Database operations [EXISTS]
│   │   │   ├── pages.go                      # Page operations [EXISTS]
│   │   │   ├── blocks.go                     # Block operations [NEW]
│   │   │   ├── sync.go                       # Notion sync worker [NEW]
│   │   │   └── search.go                     # Notion search [EXISTS]
│   │   │
│   │   ├── discord/                          # Discord
│   │   │   ├── provider.go                   # Discord provider [NEW]
│   │   │   ├── oauth.go                      # Discord OAuth [NEW]
│   │   │   ├── guilds.go                     # Server operations [NEW]
│   │   │   ├── channels.go                   # Channel operations [NEW]
│   │   │   └── messages.go                   # Message operations [NEW]
│   │   │
│   │   ├── clickup/                          # ClickUp
│   │   │   ├── provider.go                   # ClickUp provider [NEW]
│   │   │   ├── oauth.go                      # ClickUp OAuth [NEW]
│   │   │   ├── workspaces.go                 # Workspace operations [NEW]
│   │   │   ├── spaces.go                     # Space operations [NEW]
│   │   │   ├── folders.go                    # Folder operations [NEW]
│   │   │   ├── lists.go                      # List operations [NEW]
│   │   │   ├── tasks.go                      # Task operations [NEW]
│   │   │   ├── sync.go                       # ClickUp sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping config [NEW]
│   │   │
│   │   ├── asana/                            # Asana
│   │   │   ├── provider.go                   # Asana provider [NEW]
│   │   │   ├── oauth.go                      # Asana OAuth [NEW]
│   │   │   ├── workspaces.go                 # Workspace operations [NEW]
│   │   │   ├── projects.go                   # Project operations [NEW]
│   │   │   ├── tasks.go                      # Task operations [NEW]
│   │   │   ├── portfolios.go                 # Portfolio operations [NEW]
│   │   │   ├── sync.go                       # Asana sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── linear/                           # Linear
│   │   │   ├── provider.go                   # Linear provider [STUB EXISTS]
│   │   │   ├── oauth.go                      # Linear OAuth [NEW]
│   │   │   ├── teams.go                      # Team operations [NEW]
│   │   │   ├── projects.go                   # Project operations [NEW]
│   │   │   ├── issues.go                     # Issue operations [NEW]
│   │   │   ├── cycles.go                     # Cycle operations [NEW]
│   │   │   ├── sync.go                       # Linear sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── monday/                           # Monday.com
│   │   │   ├── provider.go                   # Monday provider [NEW]
│   │   │   ├── oauth.go                      # Monday OAuth [NEW]
│   │   │   ├── boards.go                     # Board operations [NEW]
│   │   │   ├── items.go                      # Item operations [NEW]
│   │   │   ├── columns.go                    # Column operations [NEW]
│   │   │   ├── sync.go                       # Monday sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── jira/                             # Jira
│   │   │   ├── provider.go                   # Jira provider [NEW]
│   │   │   ├── oauth.go                      # Atlassian OAuth [NEW]
│   │   │   ├── projects.go                   # Project operations [NEW]
│   │   │   ├── issues.go                     # Issue operations [NEW]
│   │   │   ├── sprints.go                    # Sprint operations [NEW]
│   │   │   ├── boards.go                     # Board operations [NEW]
│   │   │   ├── sync.go                       # Jira sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── trello/                           # Trello
│   │   │   ├── provider.go                   # Trello provider [NEW]
│   │   │   ├── oauth.go                      # Atlassian OAuth (shared) [NEW]
│   │   │   ├── boards.go                     # Board operations [NEW]
│   │   │   ├── lists.go                      # List operations [NEW]
│   │   │   ├── cards.go                      # Card operations [NEW]
│   │   │   ├── sync.go                       # Trello sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── todoist/                          # Todoist
│   │   │   ├── provider.go                   # Todoist provider [NEW]
│   │   │   ├── oauth.go                      # Todoist OAuth [NEW]
│   │   │   ├── projects.go                   # Project operations [NEW]
│   │   │   ├── tasks.go                      # Task operations [NEW]
│   │   │   ├── sync.go                       # Todoist sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── hubspot/                          # HubSpot
│   │   │   ├── provider.go                   # HubSpot provider [STUB EXISTS]
│   │   │   ├── oauth.go                      # HubSpot OAuth [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── companies.go                  # Company operations [NEW]
│   │   │   ├── deals.go                      # Deal operations [NEW]
│   │   │   ├── tickets.go                    # Ticket operations [NEW]
│   │   │   ├── activities.go                 # Activity operations [NEW]
│   │   │   ├── sync.go                       # HubSpot sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── salesforce/                       # Salesforce
│   │   │   ├── provider.go                   # Salesforce provider [NEW]
│   │   │   ├── oauth.go                      # Salesforce OAuth [NEW]
│   │   │   ├── accounts.go                   # Account operations [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── opportunities.go              # Opportunity operations [NEW]
│   │   │   ├── leads.go                      # Lead operations [NEW]
│   │   │   ├── sync.go                       # Salesforce sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── pipedrive/                        # Pipedrive
│   │   │   ├── provider.go                   # Pipedrive provider [NEW]
│   │   │   ├── oauth.go                      # Pipedrive OAuth [NEW]
│   │   │   ├── persons.go                    # Person operations [NEW]
│   │   │   ├── organizations.go              # Organization operations [NEW]
│   │   │   ├── deals.go                      # Deal operations [NEW]
│   │   │   ├── activities.go                 # Activity operations [NEW]
│   │   │   ├── sync.go                       # Pipedrive sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── gohighlevel/                      # GoHighLevel
│   │   │   ├── provider.go                   # GHL provider [STUB EXISTS]
│   │   │   ├── oauth.go                      # GHL OAuth [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── opportunities.go              # Opportunity operations [NEW]
│   │   │   ├── calendars.go                  # Calendar operations [NEW]
│   │   │   ├── conversations.go              # Conversation operations [NEW]
│   │   │   ├── sync.go                       # GHL sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── freshsales/                       # Freshsales
│   │   │   ├── provider.go                   # Freshsales provider [NEW]
│   │   │   ├── oauth.go                      # Freshsales OAuth [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── accounts.go                   # Account operations [NEW]
│   │   │   ├── deals.go                      # Deal operations [NEW]
│   │   │   ├── sync.go                       # Freshsales sync worker [NEW]
│   │   │   └── mapping.go                    # Field mapping [NEW]
│   │   │
│   │   ├── zoom/                             # Zoom
│   │   │   ├── provider.go                   # Zoom provider [NEW]
│   │   │   ├── oauth.go                      # Zoom OAuth [NEW]
│   │   │   ├── meetings.go                   # Meeting operations [NEW]
│   │   │   ├── recordings.go                 # Recording operations [NEW]
│   │   │   ├── transcripts.go                # Transcript operations [NEW]
│   │   │   ├── users.go                      # User operations [NEW]
│   │   │   └── sync.go                       # Zoom sync worker [NEW]
│   │   │
│   │   ├── loom/                             # Loom
│   │   │   ├── provider.go                   # Loom provider [NEW]
│   │   │   ├── oauth.go                      # Loom OAuth [NEW]
│   │   │   ├── videos.go                     # Video operations [NEW]
│   │   │   ├── transcripts.go                # Transcript operations [NEW]
│   │   │   └── sync.go                       # Loom sync worker [NEW]
│   │   │
│   │   ├── fireflies/                        # Fireflies.ai
│   │   │   ├── provider.go                   # Fireflies provider [NEW]
│   │   │   ├── api.go                        # Fireflies API (API key, not OAuth) [NEW]
│   │   │   ├── meetings.go                   # Meeting operations [NEW]
│   │   │   ├── transcripts.go                # Transcript operations [NEW]
│   │   │   └── sync.go                       # Fireflies sync worker [NEW]
│   │   │
│   │   ├── fathom/                           # Fathom
│   │   │   ├── provider.go                   # Fathom provider [NEW]
│   │   │   ├── api.go                        # Fathom API [NEW]
│   │   │   ├── calls.go                      # Call operations [NEW]
│   │   │   ├── summaries.go                  # Summary operations [NEW]
│   │   │   └── sync.go                       # Fathom sync worker [NEW]
│   │   │
│   │   ├── tldv/                             # tl;dv
│   │   │   ├── provider.go                   # TLDV provider [NEW]
│   │   │   ├── api.go                        # TLDV API [NEW]
│   │   │   ├── meetings.go                   # Meeting operations [NEW]
│   │   │   ├── highlights.go                 # Highlight operations [NEW]
│   │   │   └── sync.go                       # TLDV sync worker [NEW]
│   │   │
│   │   ├── calendly/                         # Calendly
│   │   │   ├── provider.go                   # Calendly provider [NEW]
│   │   │   ├── oauth.go                      # Calendly OAuth [NEW]
│   │   │   ├── events.go                     # Event operations [NEW]
│   │   │   ├── event_types.go                # Event type operations [NEW]
│   │   │   ├── invitees.go                   # Invitee operations [NEW]
│   │   │   └── sync.go                       # Calendly sync worker [NEW]
│   │   │
│   │   ├── dropbox/                          # Dropbox
│   │   │   ├── provider.go                   # Dropbox provider [NEW]
│   │   │   ├── oauth.go                      # Dropbox OAuth [NEW]
│   │   │   ├── files.go                      # File operations [NEW]
│   │   │   ├── folders.go                    # Folder operations [NEW]
│   │   │   ├── sharing.go                    # Sharing operations [NEW]
│   │   │   └── sync.go                       # Dropbox sync worker [NEW]
│   │   │
│   │   ├── box/                              # Box
│   │   │   ├── provider.go                   # Box provider [NEW]
│   │   │   ├── oauth.go                      # Box OAuth [NEW]
│   │   │   ├── files.go                      # File operations [NEW]
│   │   │   ├── folders.go                    # Folder operations [NEW]
│   │   │   └── sync.go                       # Box sync worker [NEW]
│   │   │
│   │   ├── github/                           # GitHub
│   │   │   ├── provider.go                   # GitHub provider [NEW]
│   │   │   ├── oauth.go                      # GitHub OAuth [NEW]
│   │   │   ├── repos.go                      # Repository operations [NEW]
│   │   │   ├── issues.go                     # Issue operations [NEW]
│   │   │   ├── prs.go                        # Pull request operations [NEW]
│   │   │   ├── commits.go                    # Commit operations [NEW]
│   │   │   └── sync.go                       # GitHub sync worker [NEW]
│   │   │
│   │   ├── gitlab/                           # GitLab
│   │   │   ├── provider.go                   # GitLab provider [NEW]
│   │   │   ├── oauth.go                      # GitLab OAuth [NEW]
│   │   │   ├── projects.go                   # Project operations [NEW]
│   │   │   ├── issues.go                     # Issue operations [NEW]
│   │   │   ├── merge_requests.go             # MR operations [NEW]
│   │   │   └── sync.go                       # GitLab sync worker [NEW]
│   │   │
│   │   ├── zendesk/                          # Zendesk
│   │   │   ├── provider.go                   # Zendesk provider [NEW]
│   │   │   ├── oauth.go                      # Zendesk OAuth [NEW]
│   │   │   ├── tickets.go                    # Ticket operations [NEW]
│   │   │   ├── users.go                      # User operations [NEW]
│   │   │   └── sync.go                       # Zendesk sync worker [NEW]
│   │   │
│   │   ├── intercom/                         # Intercom
│   │   │   ├── provider.go                   # Intercom provider [NEW]
│   │   │   ├── oauth.go                      # Intercom OAuth [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── conversations.go              # Conversation operations [NEW]
│   │   │   └── sync.go                       # Intercom sync worker [NEW]
│   │   │
│   │   ├── stripe/                           # Stripe
│   │   │   ├── provider.go                   # Stripe provider [NEW]
│   │   │   ├── oauth.go                      # Stripe Connect OAuth [NEW]
│   │   │   ├── customers.go                  # Customer operations [NEW]
│   │   │   ├── invoices.go                   # Invoice operations [NEW]
│   │   │   ├── subscriptions.go              # Subscription operations [NEW]
│   │   │   └── sync.go                       # Stripe sync worker [NEW]
│   │   │
│   │   ├── quickbooks/                       # QuickBooks
│   │   │   ├── provider.go                   # QuickBooks provider [NEW]
│   │   │   ├── oauth.go                      # Intuit OAuth [NEW]
│   │   │   ├── customers.go                  # Customer operations [NEW]
│   │   │   ├── invoices.go                   # Invoice operations [NEW]
│   │   │   ├── payments.go                   # Payment operations [NEW]
│   │   │   └── sync.go                       # QuickBooks sync worker [NEW]
│   │   │
│   │   ├── xero/                             # Xero
│   │   │   ├── provider.go                   # Xero provider [NEW]
│   │   │   ├── oauth.go                      # Xero OAuth [NEW]
│   │   │   ├── contacts.go                   # Contact operations [NEW]
│   │   │   ├── invoices.go                   # Invoice operations [NEW]
│   │   │   └── sync.go                       # Xero sync worker [NEW]
│   │   │
│   │   ├── bamboohr/                         # BambooHR
│   │   │   ├── provider.go                   # BambooHR provider [NEW]
│   │   │   ├── api.go                        # BambooHR API [NEW]
│   │   │   ├── employees.go                  # Employee operations [NEW]
│   │   │   ├── timeoff.go                    # Time-off operations [NEW]
│   │   │   └── sync.go                       # BambooHR sync worker [NEW]
│   │   │
│   │   ├── airtable/                         # Airtable
│   │   │   ├── provider.go                   # Airtable provider [NEW]
│   │   │   ├── oauth.go                      # Airtable OAuth [NEW]
│   │   │   ├── bases.go                      # Base operations [NEW]
│   │   │   ├── tables.go                     # Table operations [NEW]
│   │   │   ├── records.go                    # Record operations [NEW]
│   │   │   └── sync.go                       # Airtable sync worker [NEW]
│   │   │
│   │   ├── figma/                            # Figma
│   │   │   ├── provider.go                   # Figma provider [NEW]
│   │   │   ├── oauth.go                      # Figma OAuth [NEW]
│   │   │   ├── files.go                      # File operations [NEW]
│   │   │   ├── comments.go                   # Comment operations [NEW]
│   │   │   └── sync.go                       # Figma sync worker [NEW]
│   │   │
│   │   ├── miro/                             # Miro
│   │   │   ├── provider.go                   # Miro provider [NEW]
│   │   │   ├── oauth.go                      # Miro OAuth [NEW]
│   │   │   ├── boards.go                     # Board operations [NEW]
│   │   │   └── sync.go                       # Miro sync worker [NEW]
│   │   │
│   │   └── confluence/                       # Confluence
│   │       ├── provider.go                   # Confluence provider [NEW]
│   │       ├── oauth.go                      # Atlassian OAuth (shared) [NEW]
│   │       ├── spaces.go                     # Space operations [NEW]
│   │       ├── pages.go                      # Page operations [NEW]
│   │       └── sync.go                       # Confluence sync worker [NEW]
│   │
│   ├── imports/                              # File Import System
│   │   ├── types.go                          # Import types [NEW]
│   │   ├── service.go                        # Import orchestration service [NEW]
│   │   ├── storage.go                        # File storage (GCS/local) [NEW]
│   │   ├── processor.go                      # Processing pipeline [NEW]
│   │   │
│   │   ├── parsers/                          # Provider-specific parsers
│   │   │   ├── parser.go                     # Parser interface [NEW]
│   │   │   ├── chatgpt.go                    # ChatGPT JSON parser [NEW]
│   │   │   ├── chatgpt_test.go               # ChatGPT parser tests [NEW]
│   │   │   ├── claude.go                     # Claude JSON parser [NEW]
│   │   │   ├── claude_test.go                # Claude parser tests [NEW]
│   │   │   ├── perplexity.go                 # Perplexity JSON parser [NEW]
│   │   │   ├── perplexity_test.go            # Perplexity parser tests [NEW]
│   │   │   ├── gemini.go                     # Gemini JSON parser [NEW]
│   │   │   ├── gemini_test.go                # Gemini parser tests [NEW]
│   │   │   ├── granola.go                    # Granola meeting notes parser [NEW]
│   │   │   ├── granola_test.go               # Granola parser tests [NEW]
│   │   │   ├── obsidian.go                   # Obsidian vault parser [NEW]
│   │   │   ├── roam.go                       # Roam Research parser [NEW]
│   │   │   ├── evernote.go                   # Evernote ENEX parser [NEW]
│   │   │   ├── notion_export.go              # Notion export parser [NEW]
│   │   │   ├── apple_notes.go                # Apple Notes parser [NEW]
│   │   │   ├── csv.go                        # Generic CSV parser [NEW]
│   │   │   └── markdown.go                   # Markdown folder parser [NEW]
│   │   │
│   │   └── enrichment/                       # AI Enrichment
│   │       ├── service.go                    # Enrichment orchestration [NEW]
│   │       ├── summarizer.go                 # Summary generation [NEW]
│   │       ├── topics.go                     # Topic extraction [NEW]
│   │       ├── entities.go                   # Entity extraction [NEW]
│   │       ├── knowledge.go                  # Knowledge extraction [NEW]
│   │       ├── sentiment.go                  # Sentiment analysis [NEW]
│   │       └── embeddings.go                 # Vector embedding generation [NEW]
│   │
│   ├── workers/                              # Background Workers
│   │   ├── types.go                          # Worker interfaces [NEW]
│   │   ├── scheduler.go                      # Job scheduler [NEW]
│   │   ├── sync_runner.go                    # Sync job runner [NEW]
│   │   ├── import_runner.go                  # Import job runner [NEW]
│   │   ├── enrichment_runner.go              # Enrichment job runner [NEW]
│   │   └── webhook_runner.go                 # Webhook delivery runner [NEW]
│   │
│   ├── webhooks/                             # Webhook Infrastructure
│   │   ├── types.go                          # Webhook types [NEW]
│   │   ├── receiver.go                       # Incoming webhook receiver [NEW]
│   │   ├── dispatcher.go                     # Outgoing webhook dispatcher [NEW]
│   │   ├── verifier.go                       # Signature verification [NEW]
│   │   ├── retry.go                          # Retry logic [NEW]
│   │   └── handlers/                         # Provider-specific handlers
│   │       ├── slack.go                      # Slack webhook handler [NEW]
│   │       ├── github.go                     # GitHub webhook handler [NEW]
│   │       ├── stripe.go                     # Stripe webhook handler [NEW]
│   │       ├── hubspot.go                    # HubSpot webhook handler [NEW]
│   │       └── clickup.go                    # ClickUp webhook handler [NEW]
│   │
│   ├── handlers/                             # HTTP Handlers
│   │   ├── integrations.go                   # Unified integration endpoints [NEW]
│   │   ├── imports.go                        # File import endpoints [NEW]
│   │   ├── webhooks.go                       # Webhook endpoints [NEW]
│   │   ├── sync.go                           # Sync control endpoints [NEW]
│   │   └── search.go                         # Semantic search endpoints [NEW]
│   │
│   ├── services/                             # MCP Tool Services
│   │   ├── mcp.go                            # MCP aggregation [EXISTS]
│   │   ├── mcp_calendar.go                   # Calendar MCP tools [EXISTS]
│   │   ├── mcp_slack.go                      # Slack MCP tools [EXISTS]
│   │   ├── mcp_notion.go                     # Notion MCP tools [EXISTS]
│   │   ├── mcp_gmail.go                      # Gmail MCP tools [NEW]
│   │   ├── mcp_drive.go                      # Drive MCP tools [NEW]
│   │   ├── mcp_clickup.go                    # ClickUp MCP tools [NEW]
│   │   ├── mcp_asana.go                      # Asana MCP tools [NEW]
│   │   ├── mcp_linear.go                     # Linear MCP tools [NEW]
│   │   ├── mcp_hubspot.go                    # HubSpot MCP tools [NEW]
│   │   ├── mcp_zoom.go                       # Zoom MCP tools [NEW]
│   │   ├── mcp_teams.go                      # Teams MCP tools [NEW]
│   │   ├── mcp_github.go                     # GitHub MCP tools [NEW]
│   │   ├── mcp_jira.go                       # Jira MCP tools [NEW]
│   │   └── mcp_search.go                     # Semantic search MCP tools [NEW]
│   │
│   └── database/
│       ├── migrations/
│       │   ├── 025_integration_connections.sql    # Core integration tables [NEW]
│       │   ├── 026_sync_jobs.sql                  # Sync job tracking [NEW]
│       │   ├── 027_integration_audit.sql          # Audit logging [NEW]
│       │   ├── 028_webhooks.sql                   # Webhook tables [NEW]
│       │   ├── 029_file_imports.sql               # File import tables [NEW]
│       │   ├── 030_imported_conversations.sql     # Imported conversations [NEW]
│       │   ├── 031_imported_messages.sql          # Imported messages [NEW]
│       │   ├── 032_imported_knowledge.sql         # Extracted knowledge [NEW]
│       │   ├── 033_synced_tasks.sql               # Task sync mapping [NEW]
│       │   ├── 034_synced_contacts.sql            # Contact sync mapping [NEW]
│       │   ├── 035_synced_deals.sql               # Deal sync mapping [NEW]
│       │   ├── 036_synced_files.sql               # File sync mapping [NEW]
│       │   └── 037_synced_meetings.sql            # Meeting sync mapping [NEW]
│       │
│       └── queries/
│           ├── integration_connections.sql        # Connection queries [NEW]
│           ├── sync_jobs.sql                      # Sync job queries [NEW]
│           ├── webhooks.sql                       # Webhook queries [NEW]
│           ├── file_imports.sql                   # Import queries [NEW]
│           ├── imported_conversations.sql         # Conversation queries [NEW]
│           ├── imported_messages.sql              # Message queries [NEW]
│           └── imported_knowledge.sql             # Knowledge queries [NEW]

frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   ├── integrations/
│   │   │   │   ├── index.ts                  # Integration API exports [NEW]
│   │   │   │   ├── types.ts                  # Integration types [EXISTS - update]
│   │   │   │   ├── client.ts                 # Integration API client [NEW]
│   │   │   │   ├── google.ts                 # Google-specific API [NEW]
│   │   │   │   ├── microsoft.ts              # Microsoft-specific API [NEW]
│   │   │   │   ├── slack.ts                  # Slack-specific API [NEW]
│   │   │   │   ├── notion.ts                 # Notion-specific API [NEW]
│   │   │   │   ├── clickup.ts                # ClickUp-specific API [NEW]
│   │   │   │   ├── asana.ts                  # Asana-specific API [NEW]
│   │   │   │   ├── hubspot.ts                # HubSpot-specific API [NEW]
│   │   │   │   └── zoom.ts                   # Zoom-specific API [NEW]
│   │   │   │
│   │   │   └── imports/
│   │   │       ├── index.ts                  # Import API exports [NEW]
│   │   │       ├── types.ts                  # Import types [NEW]
│   │   │       └── client.ts                 # Import API client [NEW]
│   │   │
│   │   ├── stores/
│   │   │   ├── integrationsStore.ts          # Integration state [NEW]
│   │   │   ├── importsStore.ts               # Import state [NEW]
│   │   │   └── syncStore.ts                  # Sync state [NEW]
│   │   │
│   │   └── components/
│   │       ├── settings/
│   │       │   └── integrations/
│   │       │       ├── IntegrationsPage.svelte        # Main integrations page [NEW]
│   │       │       ├── IntegrationGrid.svelte         # Grid of all integrations [NEW]
│   │       │       ├── IntegrationCategory.svelte     # Category section [NEW]
│   │       │       │
│   │       │       ├── cards/
│   │       │       │   ├── IntegrationCard.svelte     # Base integration card [NEW]
│   │       │       │   ├── GoogleCard.svelte          # Google services card [NEW]
│   │       │       │   ├── MicrosoftCard.svelte       # Microsoft services card [NEW]
│   │       │       │   ├── SlackCard.svelte           # Slack card [NEW]
│   │       │       │   ├── NotionCard.svelte          # Notion card [NEW]
│   │       │       │   ├── ClickUpCard.svelte         # ClickUp card [NEW]
│   │       │       │   ├── AsanaCard.svelte           # Asana card [NEW]
│   │       │       │   ├── LinearCard.svelte          # Linear card [NEW]
│   │       │       │   ├── HubSpotCard.svelte         # HubSpot card [NEW]
│   │       │       │   ├── ZoomCard.svelte            # Zoom card [NEW]
│   │       │       │   └── GenericCard.svelte         # Generic card template [NEW]
│   │       │       │
│   │       │       ├── modals/
│   │       │       │   ├── ConnectModal.svelte        # OAuth connection modal [NEW]
│   │       │       │   ├── DisconnectModal.svelte     # Disconnect confirmation [NEW]
│   │       │       │   ├── SyncSettingsModal.svelte   # Sync configuration [NEW]
│   │       │       │   ├── MappingModal.svelte        # Field mapping config [NEW]
│   │       │       │   └── LogsModal.svelte           # View sync logs [NEW]
│   │       │       │
│   │       │       ├── status/
│   │       │       │   ├── ConnectionStatus.svelte    # Connection indicator [NEW]
│   │       │       │   ├── SyncStatus.svelte          # Sync progress/status [NEW]
│   │       │       │   ├── SyncHistory.svelte         # Sync history list [NEW]
│   │       │       │   └── ErrorDisplay.svelte        # Error handling [NEW]
│   │       │       │
│   │       │       └── sync/
│   │       │           ├── SyncControls.svelte        # Sync trigger controls [NEW]
│   │       │           ├── SyncSchedule.svelte        # Schedule configuration [NEW]
│   │       │           └── SyncProgress.svelte        # Real-time progress [NEW]
│   │       │
│   │       └── imports/
│   │           ├── ImportsPage.svelte             # Main imports page [NEW]
│   │           │
│   │           ├── upload/
│   │           │   ├── FileDropzone.svelte        # Drag-drop upload [NEW]
│   │           │   ├── ProviderSelector.svelte    # Select provider type [NEW]
│   │           │   ├── FilePreview.svelte         # Preview before import [NEW]
│   │           │   └── ImportOptions.svelte       # Import configuration [NEW]
│   │           │
│   │           ├── progress/
│   │           │   ├── ImportProgress.svelte      # Progress bar [NEW]
│   │           │   ├── ProcessingStatus.svelte    # Processing steps [NEW]
│   │           │   └── EnrichmentStatus.svelte    # AI enrichment status [NEW]
│   │           │
│   │           ├── history/
│   │           │   ├── ImportHistory.svelte       # List of imports [NEW]
│   │           │   ├── ImportCard.svelte          # Single import card [NEW]
│   │           │   └── ImportDetails.svelte       # Import details modal [NEW]
│   │           │
│   │           ├── browser/
│   │           │   ├── ConversationList.svelte    # Browse imported convos [NEW]
│   │           │   ├── ConversationView.svelte    # View single convo [NEW]
│   │           │   ├── MessageList.svelte         # Messages in convo [NEW]
│   │           │   └── KnowledgePanel.svelte      # Extracted knowledge [NEW]
│   │           │
│   │           └── search/
│   │               ├── ImportSearch.svelte        # Search imported data [NEW]
│   │               ├── SearchResults.svelte       # Search results [NEW]
│   │               └── SemanticSearch.svelte      # AI-powered search [NEW]
│   │
│   └── routes/
│       └── (app)/
│           └── settings/
│               ├── integrations/
│               │   ├── +page.svelte               # Integrations list page [NEW]
│               │   ├── +page.server.ts            # Server data loading [NEW]
│               │   └── [provider]/
│               │       ├── +page.svelte           # Provider detail page [NEW]
│               │       ├── +page.server.ts        # Provider data loading [NEW]
│               │       └── callback/
│               │           └── +page.server.ts    # OAuth callback [NEW]
│               │
│               └── imports/
│                   ├── +page.svelte               # Imports list page [NEW]
│                   ├── +page.server.ts            # Server data loading [NEW]
│                   ├── upload/
│                   │   └── +page.svelte           # Upload page [NEW]
│                   └── [id]/
│                       ├── +page.svelte           # Import detail page [NEW]
│                       └── +page.server.ts        # Import data loading [NEW]
```

---

## Complete Database Schema

### Core Integration Tables

```sql
-- ═══════════════════════════════════════════════════════════════════
-- INTEGRATION CONNECTION MANAGEMENT
-- ═══════════════════════════════════════════════════════════════════

-- Master integration connections table
CREATE TABLE integration_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_category VARCHAR(30) NOT NULL, -- 'calendar', 'communication', 'tasks', 'crm', 'storage', 'video', 'code', 'finance', 'hr', 'design'

    -- Connection status
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected', -- 'connected', 'disconnected', 'error', 'expired'
    connected_at TIMESTAMPTZ,
    disconnected_at TIMESTAMPTZ,

    -- Account info
    account_id VARCHAR(255),
    account_name VARCHAR(255),
    account_email VARCHAR(255),
    account_avatar_url TEXT,

    -- Tokens (encrypted)
    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT[],

    -- Sync configuration
    sync_enabled BOOLEAN DEFAULT true,
    sync_frequency VARCHAR(20) DEFAULT 'hourly', -- 'realtime', 'hourly', 'daily', 'manual'
    last_sync_at TIMESTAMPTZ,
    last_sync_status VARCHAR(20),
    last_sync_error TEXT,
    next_sync_at TIMESTAMPTZ,

    -- Settings
    settings JSONB DEFAULT '{}',
    field_mappings JSONB DEFAULT '{}',

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider),
    CONSTRAINT valid_status CHECK (status IN ('connected', 'disconnected', 'error', 'expired')),
    CONSTRAINT valid_sync_frequency CHECK (sync_frequency IN ('realtime', 'hourly', 'daily', 'weekly', 'manual'))
);

CREATE INDEX idx_integration_connections_user ON integration_connections(user_id);
CREATE INDEX idx_integration_connections_provider ON integration_connections(provider);
CREATE INDEX idx_integration_connections_status ON integration_connections(status);
CREATE INDEX idx_integration_connections_category ON integration_connections(provider_category);
CREATE INDEX idx_integration_connections_next_sync ON integration_connections(next_sync_at) WHERE sync_enabled = true;

-- ═══════════════════════════════════════════════════════════════════
-- SYNC JOB MANAGEMENT
-- ═══════════════════════════════════════════════════════════════════

CREATE TABLE sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    -- Job type and status
    job_type VARCHAR(30) NOT NULL, -- 'full', 'incremental', 'manual', 'webhook_triggered'
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    priority INT DEFAULT 5, -- 1 = highest

    -- Resources being synced
    resources TEXT[], -- e.g., ['tasks', 'projects'] or ['contacts', 'deals']

    -- Progress tracking
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    items_total INT DEFAULT 0,
    items_processed INT DEFAULT 0,
    items_created INT DEFAULT 0,
    items_updated INT DEFAULT 0,
    items_deleted INT DEFAULT 0,
    items_failed INT DEFAULT 0,

    -- Error handling
    error_message TEXT,
    error_details JSONB,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,

    -- Sync state
    sync_token TEXT, -- For incremental sync
    last_cursor TEXT, -- For pagination

    -- Metadata
    triggered_by VARCHAR(50), -- 'schedule', 'user', 'webhook', 'system'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sync_jobs_connection ON sync_jobs(connection_id);
CREATE INDEX idx_sync_jobs_user ON sync_jobs(user_id);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);
CREATE INDEX idx_sync_jobs_created ON sync_jobs(created_at DESC);
CREATE INDEX idx_sync_jobs_pending ON sync_jobs(priority, created_at) WHERE status = 'pending';

-- ═══════════════════════════════════════════════════════════════════
-- WEBHOOK MANAGEMENT
-- ═══════════════════════════════════════════════════════════════════

-- Incoming webhook registrations
CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    -- Subscription details
    webhook_id VARCHAR(255), -- Provider's webhook ID
    webhook_url TEXT NOT NULL,
    events TEXT[] NOT NULL, -- Events subscribed to
    secret_encrypted TEXT, -- For signature verification

    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'failed'
    last_triggered_at TIMESTAMPTZ,
    failure_count INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Webhook delivery log
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES webhook_subscriptions(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_id VARCHAR(255),
    payload JSONB NOT NULL,

    -- Delivery status
    status VARCHAR(20) NOT NULL, -- 'received', 'processing', 'processed', 'failed'
    processed_at TIMESTAMPTZ,
    error_message TEXT,

    -- Response actions taken
    actions_taken JSONB DEFAULT '[]',

    received_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_user ON webhook_deliveries(user_id);
CREATE INDEX idx_webhook_deliveries_provider ON webhook_deliveries(provider);
CREATE INDEX idx_webhook_deliveries_received ON webhook_deliveries(received_at DESC);

-- ═══════════════════════════════════════════════════════════════════
-- AUDIT LOGGING
-- ═══════════════════════════════════════════════════════════════════

CREATE TABLE integration_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID REFERENCES integration_connections(id) ON DELETE SET NULL,
    provider VARCHAR(50) NOT NULL,

    -- Action details
    action VARCHAR(50) NOT NULL, -- 'connect', 'disconnect', 'sync_start', 'sync_complete', 'sync_fail', 'settings_change', 'error'
    action_details JSONB DEFAULT '{}',

    -- Context
    ip_address INET,
    user_agent TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_integration_audit_user ON integration_audit_log(user_id);
CREATE INDEX idx_integration_audit_provider ON integration_audit_log(provider);
CREATE INDEX idx_integration_audit_created ON integration_audit_log(created_at DESC);
CREATE INDEX idx_integration_audit_action ON integration_audit_log(action);
```

### File Import Tables

```sql
-- ═══════════════════════════════════════════════════════════════════
-- FILE IMPORT SYSTEM
-- ═══════════════════════════════════════════════════════════════════

CREATE TABLE file_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- File info
    provider VARCHAR(50) NOT NULL, -- 'chatgpt', 'claude', 'perplexity', 'gemini', 'granola', 'obsidian', 'roam', 'evernote', 'notion_export', 'csv', 'markdown'
    filename VARCHAR(500) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    file_hash VARCHAR(64), -- SHA-256 for deduplication
    storage_path TEXT, -- GCS or local path

    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'uploading', 'parsing', 'enriching', 'completed', 'failed', 'cancelled'

    -- Statistics
    total_conversations INT DEFAULT 0,
    total_messages INT DEFAULT 0,
    total_words INT DEFAULT 0,
    processed_conversations INT DEFAULT 0,
    processed_messages INT DEFAULT 0,

    -- Processing options
    options JSONB DEFAULT '{}', -- {generate_summaries, extract_knowledge, create_memories, date_range, etc.}

    -- Progress tracking
    current_step VARCHAR(50),
    step_progress FLOAT DEFAULT 0,

    -- Timing
    uploaded_at TIMESTAMPTZ,
    parsing_started_at TIMESTAMPTZ,
    parsing_completed_at TIMESTAMPTZ,
    enrichment_started_at TIMESTAMPTZ,
    enrichment_completed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Error handling
    error_message TEXT,
    error_details JSONB,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT valid_import_status CHECK (status IN ('pending', 'uploading', 'parsing', 'enriching', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX idx_file_imports_user ON file_imports(user_id);
CREATE INDEX idx_file_imports_status ON file_imports(status);
CREATE INDEX idx_file_imports_provider ON file_imports(provider);
CREATE INDEX idx_file_imports_created ON file_imports(created_at DESC);

-- Imported conversations
CREATE TABLE imported_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    import_id UUID NOT NULL REFERENCES file_imports(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Source identification
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(500), -- Original conversation ID

    -- Content
    title VARCHAR(1000),
    summary TEXT,
    summary_short VARCHAR(500),

    -- Timestamps from source
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,

    -- Classification
    topics TEXT[],
    entities JSONB DEFAULT '[]', -- [{type, value, count}]
    sentiment VARCHAR(20), -- 'positive', 'negative', 'neutral', 'mixed'
    category VARCHAR(100),
    language VARCHAR(10) DEFAULT 'en',

    -- Statistics
    message_count INT DEFAULT 0,
    word_count INT DEFAULT 0,

    -- Integration links
    linked_project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    linked_client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    linked_context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,

    -- Flags
    is_starred BOOLEAN DEFAULT false,
    is_archived BOOLEAN DEFAULT false,
    is_processed BOOLEAN DEFAULT false,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(import_id, external_id)
);

CREATE INDEX idx_imported_conversations_user ON imported_conversations(user_id);
CREATE INDEX idx_imported_conversations_import ON imported_conversations(import_id);
CREATE INDEX idx_imported_conversations_provider ON imported_conversations(provider);
CREATE INDEX idx_imported_conversations_topics ON imported_conversations USING GIN(topics);
CREATE INDEX idx_imported_conversations_started ON imported_conversations(started_at DESC);
CREATE INDEX idx_imported_conversations_starred ON imported_conversations(user_id) WHERE is_starred = true;

-- Imported messages
CREATE TABLE imported_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES imported_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Message content
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant', 'system', 'tool'
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text', -- 'text', 'code', 'image', 'file'

    -- Ordering
    sequence_number INT NOT NULL,
    parent_id UUID REFERENCES imported_messages(id) ON DELETE SET NULL, -- For branching conversations

    -- Source timestamp
    created_at_source TIMESTAMPTZ,

    -- Extracted data
    code_blocks JSONB DEFAULT '[]', -- [{language, code}]
    urls TEXT[],
    mentions TEXT[], -- @mentions

    -- AI analysis
    entities JSONB DEFAULT '[]',
    sentiment VARCHAR(20),

    -- Vector embedding
    embedding vector(1536),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_imported_messages_conversation ON imported_messages(conversation_id);
CREATE INDEX idx_imported_messages_user ON imported_messages(user_id);
CREATE INDEX idx_imported_messages_role ON imported_messages(role);
CREATE INDEX idx_imported_messages_sequence ON imported_messages(conversation_id, sequence_number);
CREATE INDEX idx_imported_messages_embedding ON imported_messages USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Extracted knowledge
CREATE TABLE imported_knowledge (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES imported_conversations(id) ON DELETE SET NULL,
    message_id UUID REFERENCES imported_messages(id) ON DELETE SET NULL,
    import_id UUID REFERENCES file_imports(id) ON DELETE SET NULL,

    -- Classification
    knowledge_type VARCHAR(30) NOT NULL, -- 'fact', 'decision', 'insight', 'preference', 'goal', 'task', 'question', 'answer', 'code_snippet', 'workflow'

    -- Content
    title VARCHAR(500),
    content TEXT NOT NULL,
    content_short VARCHAR(500),

    -- Confidence and validation
    confidence FLOAT DEFAULT 1.0,
    ai_generated BOOLEAN DEFAULT true,
    user_verified BOOLEAN DEFAULT false,

    -- Categorization
    category VARCHAR(100),
    tags TEXT[],
    related_entities JSONB DEFAULT '[]',

    -- Integration
    linked_memory_id UUID REFERENCES memories(id) ON DELETE SET NULL,
    linked_context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    linked_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,

    -- Source tracking
    source_provider VARCHAR(50),
    source_quote TEXT,

    -- Vector embedding
    embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_imported_knowledge_user ON imported_knowledge(user_id);
CREATE INDEX idx_imported_knowledge_type ON imported_knowledge(knowledge_type);
CREATE INDEX idx_imported_knowledge_conversation ON imported_knowledge(conversation_id);
CREATE INDEX idx_imported_knowledge_tags ON imported_knowledge USING GIN(tags);
CREATE INDEX idx_imported_knowledge_verified ON imported_knowledge(user_id) WHERE user_verified = true;
CREATE INDEX idx_imported_knowledge_embedding ON imported_knowledge USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

### Sync Mapping Tables

```sql
-- ═══════════════════════════════════════════════════════════════════
-- SYNC MAPPING TABLES (External <-> Internal)
-- ═══════════════════════════════════════════════════════════════════

-- Task sync mapping
CREATE TABLE synced_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    -- IDs
    internal_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    external_task_id VARCHAR(500) NOT NULL,
    external_project_id VARCHAR(500),
    external_list_id VARCHAR(500),

    -- Sync state
    sync_status VARCHAR(20) DEFAULT 'synced', -- 'synced', 'pending_push', 'pending_pull', 'conflict'
    last_synced_at TIMESTAMPTZ,
    internal_updated_at TIMESTAMPTZ,
    external_updated_at TIMESTAMPTZ,

    -- Conflict handling
    conflict_data JSONB,

    -- Field mapping overrides
    field_mappings JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_task_id),
    UNIQUE(connection_id, internal_task_id)
);

CREATE INDEX idx_synced_tasks_user ON synced_tasks(user_id);
CREATE INDEX idx_synced_tasks_internal ON synced_tasks(internal_task_id);
CREATE INDEX idx_synced_tasks_external ON synced_tasks(connection_id, external_task_id);
CREATE INDEX idx_synced_tasks_status ON synced_tasks(sync_status) WHERE sync_status != 'synced';

-- Project sync mapping
CREATE TABLE synced_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    internal_project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    external_project_id VARCHAR(500) NOT NULL,
    external_workspace_id VARCHAR(500),

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,

    field_mappings JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_project_id),
    UNIQUE(connection_id, internal_project_id)
);

-- Contact/Client sync mapping
CREATE TABLE synced_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    internal_client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    external_contact_id VARCHAR(500) NOT NULL,
    external_company_id VARCHAR(500),

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,

    field_mappings JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_contact_id),
    UNIQUE(connection_id, internal_client_id)
);

-- Calendar event sync mapping
CREATE TABLE synced_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    internal_event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    external_event_id VARCHAR(500) NOT NULL,
    external_calendar_id VARCHAR(500),

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_event_id)
);

-- File/Document sync mapping
CREATE TABLE synced_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    internal_document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    internal_context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    external_file_id VARCHAR(500) NOT NULL,
    external_folder_id VARCHAR(500),

    file_name VARCHAR(500),
    file_type VARCHAR(100),
    file_size_bytes BIGINT,

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,
    external_modified_at TIMESTAMPTZ,

    -- Content indexing
    content_indexed BOOLEAN DEFAULT false,
    content_embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_file_id)
);

-- Team member sync mapping
CREATE TABLE synced_team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,

    internal_member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    external_user_id VARCHAR(500) NOT NULL,

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_user_id)
);

-- Meeting/Recording sync mapping
CREATE TABLE synced_meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'zoom', 'teams', 'meet', 'fireflies', 'fathom', 'tldv', 'granola'

    internal_event_id UUID REFERENCES calendar_events(id) ON DELETE SET NULL,
    internal_context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    external_meeting_id VARCHAR(500) NOT NULL,

    -- Meeting details
    title VARCHAR(500),
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_minutes INT,

    -- Recording/Transcript
    recording_url TEXT,
    transcript_text TEXT,
    transcript_summary TEXT,

    -- Participants
    participants JSONB DEFAULT '[]',

    -- AI analysis
    topics TEXT[],
    action_items JSONB DEFAULT '[]',
    decisions JSONB DEFAULT '[]',

    sync_status VARCHAR(20) DEFAULT 'synced',
    last_synced_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(connection_id, external_meeting_id)
);

CREATE INDEX idx_synced_meetings_user ON synced_meetings(user_id);
CREATE INDEX idx_synced_meetings_provider ON synced_meetings(provider);
CREATE INDEX idx_synced_meetings_start ON synced_meetings(start_time DESC);
```

---

## Complete API Endpoints

### Integration Management

```
# Core Integration Endpoints
GET    /api/integrations                              # List all available integrations
GET    /api/integrations/status                       # Get status of all connected integrations
GET    /api/integrations/categories                   # List integration categories

# Per-Provider Endpoints (pattern for ALL providers)
GET    /api/integrations/:provider                    # Get provider details
GET    /api/integrations/:provider/status             # Get connection status
GET    /api/integrations/:provider/auth               # Get OAuth URL
GET    /api/integrations/:provider/callback           # OAuth callback (handled server-side)
DELETE /api/integrations/:provider                    # Disconnect integration
PUT    /api/integrations/:provider/settings           # Update settings
GET    /api/integrations/:provider/settings           # Get settings

# Sync Endpoints
POST   /api/integrations/:provider/sync               # Trigger manual sync
GET    /api/integrations/:provider/sync               # Get current sync status
GET    /api/integrations/:provider/sync/history       # Get sync history
DELETE /api/integrations/:provider/sync/:job_id       # Cancel sync job
PUT    /api/integrations/:provider/sync/schedule      # Update sync schedule

# Audit & Logs
GET    /api/integrations/:provider/logs               # Get audit logs
GET    /api/integrations/:provider/errors             # Get recent errors
```

### Provider-Specific Endpoints

```
# Google Calendar
GET    /api/integrations/google/calendars             # List calendars
GET    /api/integrations/google/calendars/:id/events  # List events
POST   /api/integrations/google/calendars/:id/events  # Create event
PUT    /api/integrations/google/calendars/:id/events/:eventId
DELETE /api/integrations/google/calendars/:id/events/:eventId

# Gmail
GET    /api/integrations/gmail/threads                # List threads
GET    /api/integrations/gmail/threads/:id            # Get thread
GET    /api/integrations/gmail/messages               # Search messages
GET    /api/integrations/gmail/labels                 # List labels

# Google Drive
GET    /api/integrations/drive/files                  # List/search files
GET    /api/integrations/drive/files/:id              # Get file details
GET    /api/integrations/drive/files/:id/content      # Get file content
GET    /api/integrations/drive/folders                # List folders
GET    /api/integrations/drive/folders/:id            # List folder contents

# Slack
GET    /api/integrations/slack/channels               # List channels
GET    /api/integrations/slack/channels/:id/messages  # Get channel messages
POST   /api/integrations/slack/channels/:id/messages  # Send message
GET    /api/integrations/slack/users                  # List users
GET    /api/integrations/slack/search                 # Search messages

# Notion
GET    /api/integrations/notion/databases             # List databases
GET    /api/integrations/notion/databases/:id         # Get database
POST   /api/integrations/notion/databases/:id/query   # Query database
GET    /api/integrations/notion/pages/:id             # Get page
POST   /api/integrations/notion/pages                 # Create page
PATCH  /api/integrations/notion/pages/:id             # Update page
GET    /api/integrations/notion/search                # Search

# ClickUp
GET    /api/integrations/clickup/workspaces           # List workspaces
GET    /api/integrations/clickup/spaces               # List spaces
GET    /api/integrations/clickup/folders              # List folders
GET    /api/integrations/clickup/lists                # List lists
GET    /api/integrations/clickup/tasks                # List tasks
POST   /api/integrations/clickup/tasks                # Create task
PUT    /api/integrations/clickup/tasks/:id            # Update task
DELETE /api/integrations/clickup/tasks/:id            # Delete task

# Asana
GET    /api/integrations/asana/workspaces             # List workspaces
GET    /api/integrations/asana/projects               # List projects
GET    /api/integrations/asana/tasks                  # List tasks
POST   /api/integrations/asana/tasks                  # Create task
PUT    /api/integrations/asana/tasks/:id              # Update task

# Linear
GET    /api/integrations/linear/teams                 # List teams
GET    /api/integrations/linear/projects              # List projects
GET    /api/integrations/linear/issues                # List issues
POST   /api/integrations/linear/issues                # Create issue
PUT    /api/integrations/linear/issues/:id            # Update issue

# HubSpot
GET    /api/integrations/hubspot/contacts             # List contacts
GET    /api/integrations/hubspot/contacts/:id         # Get contact
POST   /api/integrations/hubspot/contacts             # Create contact
PUT    /api/integrations/hubspot/contacts/:id         # Update contact
GET    /api/integrations/hubspot/companies            # List companies
GET    /api/integrations/hubspot/deals                # List deals
POST   /api/integrations/hubspot/deals                # Create deal

# Zoom
GET    /api/integrations/zoom/meetings                # List meetings
POST   /api/integrations/zoom/meetings                # Create meeting
GET    /api/integrations/zoom/meetings/:id            # Get meeting
GET    /api/integrations/zoom/recordings              # List recordings
GET    /api/integrations/zoom/recordings/:id          # Get recording/transcript

# Microsoft Teams
GET    /api/integrations/teams/teams                  # List teams
GET    /api/integrations/teams/channels               # List channels
POST   /api/integrations/teams/messages               # Send message
GET    /api/integrations/teams/presence               # Get presence

# GitHub
GET    /api/integrations/github/repos                 # List repos
GET    /api/integrations/github/repos/:owner/:repo/issues
POST   /api/integrations/github/repos/:owner/:repo/issues
GET    /api/integrations/github/repos/:owner/:repo/pulls
```

### File Import Endpoints

```
# Import Management
POST   /api/imports/upload                            # Upload file for import
GET    /api/imports                                   # List all imports
GET    /api/imports/:id                               # Get import details
DELETE /api/imports/:id                               # Delete import
POST   /api/imports/:id/retry                         # Retry failed import
POST   /api/imports/:id/cancel                        # Cancel in-progress import

# Import Progress
GET    /api/imports/:id/progress                      # Get real-time progress (SSE)
GET    /api/imports/:id/stats                         # Get import statistics

# Imported Data Access
GET    /api/imports/conversations                     # List all imported conversations
GET    /api/imports/conversations/:id                 # Get conversation with messages
GET    /api/imports/conversations/:id/messages        # Get messages only
GET    /api/imports/conversations/:id/knowledge       # Get extracted knowledge
POST   /api/imports/conversations/:id/link            # Link to project/client

# Knowledge Access
GET    /api/imports/knowledge                         # List extracted knowledge
GET    /api/imports/knowledge/:id                     # Get knowledge item
PUT    /api/imports/knowledge/:id                     # Update/verify knowledge
POST   /api/imports/knowledge/:id/create-memory       # Create memory from knowledge

# Search
GET    /api/imports/search                            # Semantic search across imports
GET    /api/imports/search/similar                    # Find similar content
```

### Webhook Endpoints

```
# Incoming Webhooks (from providers)
POST   /api/webhooks/slack                            # Slack events
POST   /api/webhooks/github                           # GitHub events
POST   /api/webhooks/stripe                           # Stripe events
POST   /api/webhooks/hubspot                          # HubSpot events
POST   /api/webhooks/clickup                          # ClickUp events
POST   /api/webhooks/linear                           # Linear events
POST   /api/webhooks/zoom                             # Zoom events
POST   /api/webhooks/calendly                         # Calendly events

# Webhook Management
GET    /api/webhooks/subscriptions                    # List webhook subscriptions
POST   /api/webhooks/subscriptions                    # Create subscription
DELETE /api/webhooks/subscriptions/:id                # Delete subscription
GET    /api/webhooks/deliveries                       # List webhook deliveries
GET    /api/webhooks/deliveries/:id                   # Get delivery details
POST   /api/webhooks/deliveries/:id/retry             # Retry delivery
```

---

## Complete MCP Tools List

### Currently Implemented (18 tools)

```
# Google Calendar (5)
calendar_list_events          # List events in date range
calendar_create_event         # Create event with attendees
calendar_update_event         # Update existing event
calendar_delete_event         # Delete event
calendar_sync_events          # Sync events to database

# Slack (6)
slack_list_channels           # List public/private channels
slack_send_message            # Send message with thread support
slack_get_channel_history     # Get channel messages
slack_search_messages         # Search across workspace
slack_list_users              # List workspace members
slack_get_user_info           # Get user details

# Notion (7)
notion_list_databases         # List accessible databases
notion_get_database           # Get database schema
notion_query_database         # Query with filters/sorts
notion_get_page               # Get page content
notion_create_page            # Create new page
notion_update_page            # Update page properties
notion_search                 # Search workspace
```

### Planned MCP Tools (50+ more)

```
# Gmail (5)
gmail_search                  # Search emails
gmail_get_thread              # Get email thread
gmail_list_labels             # List labels
gmail_get_message             # Get single message
gmail_list_recent             # List recent emails

# Google Drive (5)
drive_search                  # Search files
drive_get_file                # Get file metadata
drive_get_content             # Get file content
drive_list_folder             # List folder contents
drive_get_permissions         # Get sharing permissions

# Google Meet (3)
meet_create_meeting           # Create meeting link
meet_get_meeting              # Get meeting details
meet_list_recordings          # List recordings

# Microsoft Outlook (5)
outlook_list_events           # List calendar events
outlook_create_event          # Create event
outlook_list_emails           # List emails
outlook_get_email             # Get email
outlook_search                # Search emails

# Microsoft Teams (4)
teams_send_message            # Send message
teams_list_channels           # List channels
teams_get_presence            # Get user presence
teams_list_members            # List team members

# OneDrive (4)
onedrive_search               # Search files
onedrive_get_file             # Get file
onedrive_list_folder          # List folder
onedrive_get_content          # Get content

# ClickUp (6)
clickup_list_tasks            # List tasks
clickup_get_task              # Get task details
clickup_create_task           # Create task
clickup_update_task           # Update task
clickup_list_spaces           # List spaces
clickup_list_lists            # List lists

# Asana (5)
asana_list_tasks              # List tasks
asana_get_task                # Get task
asana_create_task             # Create task
asana_update_task             # Update task
asana_list_projects           # List projects

# Linear (5)
linear_list_issues            # List issues
linear_get_issue              # Get issue
linear_create_issue           # Create issue
linear_update_issue           # Update issue
linear_list_projects          # List projects

# HubSpot (6)
hubspot_get_contact           # Get contact
hubspot_list_contacts         # List contacts
hubspot_create_contact        # Create contact
hubspot_list_deals            # List deals
hubspot_create_deal           # Create deal
hubspot_search                # Search CRM

# Zoom (4)
zoom_create_meeting           # Schedule meeting
zoom_list_meetings            # List meetings
zoom_get_recording            # Get recording
zoom_get_transcript           # Get transcript

# GitHub (5)
github_list_repos             # List repositories
github_list_issues            # List issues
github_create_issue           # Create issue
github_list_prs               # List pull requests
github_get_pr                 # Get PR details

# Jira (5)
jira_list_issues              # List issues
jira_get_issue                # Get issue
jira_create_issue             # Create issue
jira_update_issue             # Update issue
jira_list_projects            # List projects

# Search Tools (4)
search_imports                # Search imported conversations
search_knowledge              # Search extracted knowledge
search_files                  # Search synced files
search_all                    # Unified search across all data
```

---

## Implementation Priority

### PHASE 1: Foundation (CRITICAL) - Week 1-2

**Backend:**
1. `internal/integrations/types.go` - Core interfaces
2. `internal/integrations/registry.go` - Provider registry
3. `internal/integrations/oauth.go` - Generic OAuth helpers
4. `internal/handlers/integrations.go` - Unified endpoints
5. Database migrations for core tables

**Frontend:**
1. `integrationsStore.ts` - State management
2. `IntegrationsPage.svelte` - Main page
3. `IntegrationCard.svelte` - Base card component
4. `ConnectModal.svelte` - OAuth flow

### PHASE 2: File Imports (HIGH) - Week 2-3

**Backend:**
1. `internal/imports/service.go` - Import orchestration
2. `internal/imports/parsers/*.go` - ChatGPT, Claude, Perplexity parsers
3. `internal/handlers/imports.go` - Import endpoints
4. Database migrations for import tables

**Frontend:**
1. `ImportsPage.svelte` - Upload page
2. `FileDropzone.svelte` - Drag-drop upload
3. `ImportProgress.svelte` - Real-time progress
4. `ConversationList.svelte` - Browse imports

### PHASE 3: Refactor Existing (MEDIUM) - Week 3-4

**Backend:**
1. Migrate Google Calendar to new framework
2. Migrate Slack to new framework
3. Migrate Notion to new framework
4. Add to registry, update handlers

**Frontend:**
1. Display existing integrations in new UI
2. Add sync status indicators
3. Add settings modals

### PHASE 4: Google Ecosystem (MEDIUM) - Week 4-5

**Backend:**
1. `internal/integrations/google/gmail.go`
2. `internal/integrations/google/drive.go`
3. `internal/services/mcp_gmail.go`
4. `internal/services/mcp_drive.go`

### PHASE 5: Task Integrations (MEDIUM) - Week 5-6

**Backend:**
1. `internal/integrations/clickup/*.go`
2. `internal/integrations/asana/*.go`
3. `internal/services/mcp_clickup.go`
4. `internal/services/mcp_asana.go`
5. Task sync mapping tables

### PHASE 6: CRM Integrations (MEDIUM) - Week 6-7

**Backend:**
1. `internal/integrations/hubspot/*.go`
2. `internal/services/mcp_hubspot.go`
3. Contact sync mapping

### PHASE 7: Video/Meeting (LOW) - Week 7-8

**Backend:**
1. `internal/integrations/zoom/*.go`
2. `internal/integrations/fireflies/*.go`
3. Meeting sync and transcript processing

### PHASE 8: Enrichment & Search (LOW) - Week 8-9

**Backend:**
1. `internal/imports/enrichment/*.go`
2. Vector embeddings
3. Semantic search
4. Knowledge extraction

### PHASE 9: Webhooks (LOW) - Week 9-10

**Backend:**
1. `internal/webhooks/*.go`
2. Provider-specific handlers
3. Retry and delivery logging

### PHASE 10: Remaining Integrations (ONGOING)

Based on user demand:
- Microsoft ecosystem
- Linear, Monday, Jira, Trello
- Salesforce, Pipedrive
- GitHub, GitLab
- Stripe, QuickBooks
- And more...

---

## Environment Variables

```bash
# ═══════════════════════════════════════════════════════════════════
# EXISTING (Production Ready)
# ═══════════════════════════════════════════════════════════════════

# Google (Calendar, Gmail, Drive, Meet)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=

# Slack
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=

# Notion
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=

# ═══════════════════════════════════════════════════════════════════
# PHASE 4-6: Task & CRM Integrations
# ═══════════════════════════════════════════════════════════════════

# ClickUp
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
CLICKUP_REDIRECT_URI=

# Asana
ASANA_CLIENT_ID=
ASANA_CLIENT_SECRET=
ASANA_REDIRECT_URI=

# Linear
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
LINEAR_REDIRECT_URI=

# HubSpot
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
HUBSPOT_REDIRECT_URI=

# ═══════════════════════════════════════════════════════════════════
# PHASE 7: Video/Meeting Integrations
# ═══════════════════════════════════════════════════════════════════

# Zoom
ZOOM_CLIENT_ID=
ZOOM_CLIENT_SECRET=
ZOOM_REDIRECT_URI=

# Fireflies (API Key based)
FIREFLIES_API_KEY=

# Fathom (API Key based)
FATHOM_API_KEY=

# ═══════════════════════════════════════════════════════════════════
# PHASE 8: AI/Enrichment Services
# ═══════════════════════════════════════════════════════════════════

# OpenAI (for embeddings)
OPENAI_API_KEY=

# ═══════════════════════════════════════════════════════════════════
# FUTURE INTEGRATIONS
# ═══════════════════════════════════════════════════════════════════

# Microsoft (Outlook, Teams, OneDrive)
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=

# Jira/Trello/Confluence (Atlassian)
ATLASSIAN_CLIENT_ID=
ATLASSIAN_CLIENT_SECRET=
ATLASSIAN_REDIRECT_URI=

# Monday.com
MONDAY_CLIENT_ID=
MONDAY_CLIENT_SECRET=
MONDAY_REDIRECT_URI=

# Salesforce
SALESFORCE_CLIENT_ID=
SALESFORCE_CLIENT_SECRET=
SALESFORCE_REDIRECT_URI=

# Pipedrive
PIPEDRIVE_CLIENT_ID=
PIPEDRIVE_CLIENT_SECRET=
PIPEDRIVE_REDIRECT_URI=

# GoHighLevel
GHL_CLIENT_ID=
GHL_CLIENT_SECRET=
GHL_REDIRECT_URI=

# GitHub
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URI=

# GitLab
GITLAB_CLIENT_ID=
GITLAB_CLIENT_SECRET=
GITLAB_REDIRECT_URI=

# Discord
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
DISCORD_REDIRECT_URI=

# Dropbox
DROPBOX_CLIENT_ID=
DROPBOX_CLIENT_SECRET=
DROPBOX_REDIRECT_URI=

# Stripe
STRIPE_CLIENT_ID=
STRIPE_CLIENT_SECRET=
STRIPE_REDIRECT_URI=

# Calendly
CALENDLY_CLIENT_ID=
CALENDLY_CLIENT_SECRET=
CALENDLY_REDIRECT_URI=

# Airtable
AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
AIRTABLE_REDIRECT_URI=

# Figma
FIGMA_CLIENT_ID=
FIGMA_CLIENT_SECRET=
FIGMA_REDIRECT_URI=
```

---

## Success Criteria by Phase

### Phase 1 Complete When:
- [ ] Registry pattern working with mock provider
- [ ] Settings UI rendering all available integrations
- [ ] Connect flow working (OAuth redirect)
- [ ] Disconnect flow working
- [ ] Database tables created

### Phase 2 Complete When:
- [ ] Can upload ChatGPT JSON export
- [ ] Can upload Claude JSON export
- [ ] Can upload Perplexity JSON export
- [ ] Real-time progress display working
- [ ] Can browse imported conversations
- [ ] Can delete imports

### Phase 3 Complete When:
- [ ] Google Calendar using new registry
- [ ] Slack using new registry
- [ ] Notion using new registry
- [ ] All displayed in settings UI
- [ ] Sync status showing correctly

### Phase 4 Complete When:
- [ ] Gmail connected (read-only)
- [ ] Gmail search working
- [ ] Drive connected (read-only)
- [ ] Drive file listing working
- [ ] MCP tools working

### Phase 5 Complete When:
- [ ] ClickUp OAuth working
- [ ] ClickUp tasks syncing to BusinessOS tasks
- [ ] Asana OAuth working
- [ ] Asana tasks syncing
- [ ] Bi-directional sync for both

### Phase 6 Complete When:
- [ ] HubSpot OAuth working
- [ ] Contacts syncing to Clients module
- [ ] Deals visible
- [ ] MCP tools working

### Phase 7 Complete When:
- [ ] Zoom connected
- [ ] Meetings syncing to calendar
- [ ] Recordings accessible
- [ ] Transcripts imported

### Phase 8 Complete When:
- [ ] Imported conversations have summaries
- [ ] Topics extracted
- [ ] Entities extracted
- [ ] Knowledge items created
- [ ] Semantic search working

### Phase 9 Complete When:
- [ ] Incoming webhooks receiving events
- [ ] Events triggering sync updates
- [ ] Delivery logging working
- [ ] Retry mechanism working

---

This is the COMPLETE architecture. Every file, every endpoint, every component is mapped. The team can now see the full scope and work on any part independently.
