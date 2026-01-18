# BusinessOS Integration Setup Guide for Developers

> **Last Updated:** January 2026
> **Status:** Production Ready (10 integrations implemented)
> **Backend Location:** `desktop/backend-go/internal/integrations/`

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture Overview](#architecture-overview)
3. [Integration Priority Matrix](#integration-priority-matrix)
4. [Setup Instructions by Provider](#setup-instructions-by-provider)
   - [Google Workspace](#1-google-workspace-oauth-20)
   - [Slack](#2-slack-oauth-20)
   - [Microsoft 365](#3-microsoft-365-oauth-20)
   - [Notion](#4-notion-oauth-20)
   - [Linear](#5-linear-oauth-20--graphql)
   - [HubSpot](#6-hubspot-crm-oauth-20)
   - [ClickUp](#7-clickup-oauth-20)
   - [Airtable](#8-airtable-oauth-20)
   - [Fathom Analytics](#9-fathom-analytics-api-key)
   - [OSA](#10-osa-agent-jwt)
5. [Environment Variables Reference](#environment-variables-reference)
6. [Security Requirements](#security-requirements)
7. [Testing Your Integration](#testing-your-integration)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- Go 1.24+ installed
- PostgreSQL running with migrations applied
- Redis running (for session management)
- Access to create OAuth apps on target platforms

### Minimum Required Environment Variables

```bash
# Copy to your .env file
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/businessos

# Security (REQUIRED - generate unique values)
SECRET_KEY=your-32-char-minimum-secret-key
TOKEN_ENCRYPTION_KEY=$(openssl rand -base64 32)
REDIS_KEY_HMAC_SECRET=$(openssl rand -base64 32)

# At minimum, configure ONE integration to start:
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback
GOOGLE_INTEGRATION_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback
```

---

## Architecture Overview

### Provider Interface Pattern

All integrations implement a unified `Provider` interface:

```go
type Provider interface {
    Name() string                           // "google", "slack", "notion"
    DisplayName() string                    // "Google Workspace"
    Category() string                       // "productivity", "communication"
    GetAuthURL(state string) string         // OAuth authorization URL
    ExchangeCode(ctx, code) (*TokenResponse, error)
    RefreshToken(ctx, refreshToken) (*TokenResponse, error)
    SaveToken(ctx, userID, token) error
    GetToken(ctx, userID) (*Token, error)
    SupportsSync() bool
    Sync(ctx, userID, options) (*SyncResult, error)
}
```

### Directory Structure

```
internal/integrations/
├── handler.go          # HTTP endpoints for OAuth flows
├── types.go            # Provider interface definition
├── registry.go         # Provider registration (thread-safe)
├── oauth.go            # OAuth helpers
├── errors.go           # Error types
├── providers/
│   └── register.go     # Initializes all providers
├── google/             # 10+ files: calendar, gmail, drive, etc.
├── slack/              # channels, messages
├── microsoft/          # outlook, onedrive, todo
├── notion/             # databases, pages
├── linear/             # graphql, issues, projects
├── hubspot/            # contacts, companies, deals
├── clickup/            # tasks, lists, workspaces
├── airtable/           # bases, records, tables
├── fathom/             # analytics (API key auth)
└── osa/                # agent coordination (JWT)
```

### API Endpoints

```
GET    /api/integrations/providers              # List all available providers
GET    /api/integrations/providers/:name        # Get provider details
GET    /api/integrations/oauth/:provider/start  # Start OAuth flow
GET    /api/integrations/oauth/:provider/callback  # OAuth callback
GET    /api/integrations/                       # List user's connected integrations
GET    /api/integrations/:provider              # Get integration status
DELETE /api/integrations/:provider              # Disconnect integration
POST   /api/integrations/:provider/sync         # Trigger sync
```

---

## Integration Priority Matrix

| Priority | Integration | Auth Type | Difficulty | Status | Use Case |
|----------|-------------|-----------|------------|--------|----------|
| **P0** | Google | OAuth 2.0 | Medium | ✅ Full | Calendar, Gmail, Drive, Tasks |
| **P0** | Slack | OAuth 2.0 | Easy | ✅ Full | Team communication |
| **P0** | Notion | OAuth 2.0 | Easy | ✅ Full | Knowledge base, docs |
| **P1** | Microsoft 365 | OAuth 2.0 | Hard | ✅ Structure | Outlook, OneDrive, Teams |
| **P1** | Linear | OAuth 2.0 | Medium | ✅ Full | Issue tracking |
| **P1** | HubSpot | OAuth 2.0 | Medium | ✅ Full | CRM |
| **P2** | ClickUp | OAuth 2.0 | Easy | ✅ Structure | Task management |
| **P2** | Airtable | OAuth 2.0 | Easy | ✅ Structure | Database/spreadsheet |
| **P2** | Fathom | API Key | **Easiest** | ✅ Full | Analytics |
| **P3** | OSA | JWT | Special | ✅ Full | AI orchestration |

### Quick Wins (Easiest to Set Up)

1. **Fathom Analytics** - Just needs API key (no OAuth)
2. **Slack** - Simple OAuth, tokens don't expire
3. **Notion** - Simple OAuth, straightforward scopes
4. **ClickUp** - No granular scopes, full workspace access
5. **Airtable** - Clear scope model

---

## Setup Instructions by Provider

---

### 1. Google Workspace (OAuth 2.0)

**Difficulty:** Medium
**Features:** Calendar, Gmail, Drive, Tasks, Contacts, Docs, Sheets

#### Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable required APIs:
   - Google Calendar API
   - Gmail API
   - Google Drive API
   - Google Tasks API

#### Step 2: Configure OAuth Consent Screen

1. Navigate to **APIs & Services → OAuth consent screen**
2. Choose **External** (or Internal for workspace-only)
3. Fill required fields:
   - App name
   - User support email
   - Developer contact email
4. Add scopes (see below)
5. Add test users during development

#### Step 3: Create OAuth Credentials

1. Go to **APIs & Services → Credentials**
2. Click **Create Credentials → OAuth client ID**
3. Select **Web application**
4. Add Authorized redirect URIs:
   - Development: `http://localhost:8001/api/integrations/oauth/google/callback`
   - Production: `https://yourdomain.com/api/integrations/oauth/google/callback`
5. Save Client ID and Client Secret

#### Step 4: Configure Environment Variables

```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback
GOOGLE_INTEGRATION_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback
```

#### Required Scopes (Requested Automatically)

```
# User info (always)
openid
profile
email

# Calendar
https://www.googleapis.com/auth/calendar.readonly
https://www.googleapis.com/auth/calendar.events

# Gmail
https://www.googleapis.com/auth/gmail.readonly
https://www.googleapis.com/auth/gmail.send
https://www.googleapis.com/auth/gmail.modify

# Drive
https://www.googleapis.com/auth/drive.file
https://www.googleapis.com/auth/drive.readonly

# Tasks
https://www.googleapis.com/auth/tasks
```

#### Rate Limits

- Per-minute quotas vary by API
- Implement exponential backoff for 429 errors
- Request quota increases in Cloud Console if needed

#### Common Issues

| Issue | Solution |
|-------|----------|
| `redirect_uri_mismatch` | URIs must match exactly (case-sensitive, trailing slash) |
| `invalid_scope` | Scope not added to consent screen |
| `access_denied` | User not added as test user (during development) |
| Token expires | Tokens last 1 hour, refresh automatically |

---

### 2. Slack (OAuth 2.0)

**Difficulty:** Easy
**Features:** Channels, Messages, Users, Search

#### Step 1: Create Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click **Create New App → From scratch**
3. Enter app name and select workspace

#### Step 2: Configure OAuth

1. Navigate to **OAuth & Permissions**
2. Add Redirect URLs:
   - `http://localhost:8001/api/integrations/oauth/slack/callback`
3. Add Bot Token Scopes:
   ```
   channels:read
   channels:history
   chat:write
   users:read
   users:read.email
   ```

#### Step 3: Configure Environment Variables

```bash
SLACK_CLIENT_ID=your-client-id
SLACK_CLIENT_SECRET=your-client-secret
SLACK_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/slack/callback
```

#### Important Notes

- **Tokens don't expire** - Slack tokens are valid until revoked
- **No refresh tokens** - Not needed due to above
- Rate limits: 100 req/min (non-Marketplace apps face stricter limits starting March 2026)

---

### 3. Microsoft 365 (OAuth 2.0)

**Difficulty:** Hard (Azure AD setup)
**Features:** Outlook Mail, Calendar, OneDrive, Teams, To Do

#### Step 1: Register App in Azure AD

1. Go to [Azure Portal](https://portal.azure.com/) → **Azure Active Directory**
2. Navigate to **App registrations → New registration**
3. Configure:
   - Name: "BusinessOS Integration"
   - Account types: Multi-tenant (or single-tenant for internal)
   - Redirect URI: Web → `http://localhost:8001/api/integrations/oauth/microsoft/callback`

#### Step 2: Configure API Permissions

1. Go to **API permissions → Add a permission → Microsoft Graph**
2. Add Delegated permissions:
   ```
   User.Read
   Mail.Read
   Mail.ReadWrite
   Mail.Send
   Calendars.Read
   Calendars.ReadWrite
   Files.Read
   Files.ReadWrite
   Tasks.Read
   Tasks.ReadWrite
   ```
3. Click **Grant admin consent** (for your tenant)

#### Step 3: Create Client Secret

1. Go to **Certificates & secrets**
2. Click **New client secret**
3. Copy the value immediately (shown only once)

#### Step 4: Configure Environment Variables

```bash
MICROSOFT_CLIENT_ID=your-application-id
MICROSOFT_CLIENT_SECRET=your-client-secret
MICROSOFT_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/microsoft/callback
```

#### Rate Limits

- Graph API: Per-user and per-app limits
- Use `$select` to reduce response sizes
- Implement exponential backoff

---

### 4. Notion (OAuth 2.0)

**Difficulty:** Easy
**Features:** Databases, Pages, Search

#### Step 1: Create Notion Integration

1. Go to [notion.so/my-integrations](https://www.notion.so/my-integrations)
2. Click **New integration**
3. Choose:
   - Internal (single workspace) or
   - Public (multi-workspace with OAuth)
4. For Public OAuth, enter Redirect URI

#### Step 2: Configure Capabilities

Select what your integration can do:
- Read content
- Update content
- Insert content
- Read user information

#### Step 3: Configure Environment Variables

```bash
NOTION_CLIENT_ID=your-client-id
NOTION_CLIENT_SECRET=your-client-secret
NOTION_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/notion/callback
```

#### Important Notes

- **No granular scopes** - Access is binary (full access to authorized pages)
- **Rate limit:** 3 requests per second average
- Users must manually share pages/databases with your integration

---

### 5. Linear (OAuth 2.0 + GraphQL)

**Difficulty:** Medium
**Features:** Issues, Projects, Teams, Cycles

#### Step 1: Create Linear OAuth App

1. Go to Linear workspace settings
2. Navigate to **Settings → API → OAuth applications**
3. Create new application
4. Add redirect URI

#### Step 2: Configure Environment Variables

```bash
LINEAR_CLIENT_ID=your-client-id
LINEAR_CLIENT_SECRET=your-client-secret
LINEAR_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/linear/callback
```

#### Available Scopes

```
read   - Read issues, projects, teams
write  - Create/update issues
admin  - Workspace administration
```

#### Important Notes

- **GraphQL API** - Not REST
- Token lifetime: 24 hours (new apps Oct 2025+)
- Refresh tokens required (mandatory by April 2026)
- Rate limit: 5,000 requests/hour

---

### 6. HubSpot CRM (OAuth 2.0)

**Difficulty:** Medium
**Features:** Contacts, Companies, Deals, Tasks

#### Step 1: Create HubSpot Developer App

1. Go to HubSpot account → **Settings → Integrations → Private apps** (or Developer portal)
2. Create a **Public app** (for OAuth) or **Private app** (single account)
3. Configure OAuth redirect URL

#### Step 2: Configure Scopes

```
crm.objects.contacts.read
crm.objects.contacts.write
crm.objects.companies.read
crm.objects.companies.write
crm.objects.deals.read
crm.objects.deals.write
crm.objects.owners.read
oauth
```

#### Step 3: Configure Environment Variables

```bash
HUBSPOT_CLIENT_ID=your-client-id
HUBSPOT_CLIENT_SECRET=your-client-secret
HUBSPOT_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/hubspot/callback
```

#### Rate Limits

| Plan | Requests/Min |
|------|-------------|
| Free/Starter | 100 |
| Professional | 190 |
| Enterprise | 190 |

---

### 7. ClickUp (OAuth 2.0)

**Difficulty:** Easy
**Features:** Tasks, Lists, Spaces, Workspaces

#### Step 1: Create ClickUp App

1. Log into ClickUp
2. Go to **Settings → Apps**
3. Click **Create new app**
4. Add redirect URI

#### Step 2: Configure Environment Variables

```bash
CLICKUP_CLIENT_ID=your-client-id
CLICKUP_CLIENT_SECRET=your-client-secret
CLICKUP_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/clickup/callback
```

#### Important Notes

- **No granular scopes** - Access based on user's workspace permissions
- Personal tokens start with `pk_`
- Rate limits by plan: 100-10,000 req/min

---

### 8. Airtable (OAuth 2.0)

**Difficulty:** Easy
**Features:** Bases, Tables, Records

#### Step 1: Create Airtable OAuth App

1. Go to [airtable.com/create/oauth](https://airtable.com/create/oauth)
2. Create new OAuth integration
3. Configure redirect URI and scopes

#### Step 2: Configure Scopes

```
data.records:read
data.records:write
schema.bases:read
```

#### Step 3: Configure Environment Variables

```bash
AIRTABLE_CLIENT_ID=your-client-id
AIRTABLE_CLIENT_SECRET=your-client-secret
AIRTABLE_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/airtable/callback
```

#### Important Notes

- **Rate limit:** 5 requests per second per base (all plans)
- Users choose which bases to grant access to during authorization

---

### 9. Fathom Analytics (API Key)

**Difficulty:** Easiest
**Features:** Sites, Visitors, Events, Aggregations

#### Step 1: Get API Key

1. Log into Fathom Analytics
2. Go to **Settings → API**
3. Generate new API key

#### Step 2: User Configuration

Unlike OAuth providers, Fathom uses API key authentication:
- Users enter their API key in BusinessOS settings
- Key is validated and stored in credential vault
- No redirect flow needed

#### Environment Variables

No provider-level env vars needed. Users provide their own API keys.

---

### 10. OSA Agent (JWT)

**Difficulty:** Special (internal service)
**Features:** AI agent orchestration

This is a service-to-service integration, not user-facing OAuth.

```bash
OSA_ENABLED=true
OSA_BASE_URL=http://localhost:8089
OSA_SHARED_SECRET=your-shared-secret-minimum-32-chars
OSA_TIMEOUT=30
OSA_MAX_RETRIES=3
OSA_RETRY_DELAY=1
```

---

## Environment Variables Reference

### Complete .env Template

```bash
# ═══════════════════════════════════════════════════════════════════
# BUSINESSOS INTEGRATION ENVIRONMENT VARIABLES
# ═══════════════════════════════════════════════════════════════════

# ─── SECURITY (REQUIRED) ───────────────────────────────────────────
SECRET_KEY=your-32-char-minimum-secret-key-here
TOKEN_ENCRYPTION_KEY=base64-encoded-32-byte-key-here
REDIS_KEY_HMAC_SECRET=base64-encoded-32-byte-key-here

# ─── GOOGLE WORKSPACE ──────────────────────────────────────────────
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback
GOOGLE_INTEGRATION_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/google/callback

# ─── SLACK ─────────────────────────────────────────────────────────
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/slack/callback

# ─── MICROSOFT 365 ─────────────────────────────────────────────────
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/microsoft/callback

# ─── NOTION ────────────────────────────────────────────────────────
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/notion/callback

# ─── LINEAR ────────────────────────────────────────────────────────
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
LINEAR_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/linear/callback

# ─── HUBSPOT ───────────────────────────────────────────────────────
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
HUBSPOT_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/hubspot/callback

# ─── CLICKUP ───────────────────────────────────────────────────────
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
CLICKUP_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/clickup/callback

# ─── AIRTABLE ──────────────────────────────────────────────────────
AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
AIRTABLE_REDIRECT_URI=http://localhost:8001/api/integrations/oauth/airtable/callback

# ─── OSA (INTERNAL SERVICE) ────────────────────────────────────────
OSA_ENABLED=true
OSA_BASE_URL=http://localhost:8089
OSA_SHARED_SECRET=your-shared-secret-here
OSA_TIMEOUT=30
OSA_MAX_RETRIES=3
OSA_RETRY_DELAY=1
```

---

## Security Requirements

### Production Checklist

- [ ] **TOKEN_ENCRYPTION_KEY set** - Generate with `openssl rand -base64 32`
- [ ] **REDIS_KEY_HMAC_SECRET set** - Generate with `openssl rand -base64 32`
- [ ] **SECRET_KEY is strong** - Minimum 32 characters, random
- [ ] **ENVIRONMENT=production**
- [ ] **All redirect URIs use HTTPS**
- [ ] **Client secrets not in frontend code**
- [ ] **Tokens stored encrypted in database**

### Token Storage

All OAuth tokens are encrypted using AES-256-GCM before storage:

```go
// Tokens stored in credential_vault table
// Encrypted with TOKEN_ENCRYPTION_KEY
// Decrypted only when needed for API calls
```

### CSRF Protection

OAuth state parameter verified using constant-time comparison:

```go
subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
```

---

## Testing Your Integration

### 1. Start OAuth Flow

```bash
# Open in browser
http://localhost:8001/api/integrations/oauth/google/start
```

### 2. Check Connection Status

```bash
curl -H "Authorization: Bearer YOUR_SESSION_TOKEN" \
  http://localhost:8001/api/integrations/google
```

### 3. List All Connected Integrations

```bash
curl -H "Authorization: Bearer YOUR_SESSION_TOKEN" \
  http://localhost:8001/api/integrations/
```

### 4. Test MCP Tools

```bash
# List available tools
curl -H "Authorization: Bearer YOUR_SESSION_TOKEN" \
  http://localhost:8001/api/mcp/tools

# Execute a tool
curl -X POST http://localhost:8001/api/mcp/execute \
  -H "Authorization: Bearer YOUR_SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "calendar_list_events",
    "arguments": {
      "start_date": "2026-01-01",
      "end_date": "2026-01-31"
    }
  }'
```

---

## Troubleshooting

### Common OAuth Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `redirect_uri_mismatch` | URI doesn't match registered | Check exact match (case, trailing slash) |
| `invalid_client` | Wrong client ID/secret | Verify credentials |
| `invalid_scope` | Scope not configured | Add scope in provider's console |
| `access_denied` | User denied or not test user | Add as test user (Google) |
| `token expired` | Access token expired | Refresh token implementation |

### Debug Commands

```bash
# Check if integration is loaded
curl http://localhost:8001/api/integrations/providers

# Check specific provider config
curl http://localhost:8001/api/integrations/providers/google

# View server logs for OAuth errors
tail -f logs/server.log | grep -i oauth
```

### Rate Limit Handling

All providers implement exponential backoff:

```go
// Automatic retry on 429 responses
// Wait time: 2^attempt seconds
// Max retries: 5
```

---

## Next Steps

1. **Start with quick wins**: Fathom (API key), Slack (simple OAuth)
2. **Test OAuth flow locally** before production
3. **Generate security keys** for production
4. **Add test users** where required (Google)
5. **Monitor rate limits** in provider dashboards

---

## Support

- **Backend code**: `desktop/backend-go/internal/integrations/`
- **Existing docs**: `docs/integrations/`
- **API reference**: `docs/api/API_REFERENCE.md`

---

*Document generated: January 2026*
