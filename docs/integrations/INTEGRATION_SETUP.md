# Integration Setup Guide

This document provides step-by-step instructions for setting up OAuth integrations in BusinessOS.

## Table of Contents
- [Architecture Overview](#architecture-overview)
- [Google Workspace Setup](#google-workspace-setup)
- [Slack Setup](#slack-setup)
- [Microsoft 365 Setup](#microsoft-365-setup) (Future)
- [Other Integrations](#other-integrations)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

BusinessOS uses a **Provider Pattern** for integrations:

```
┌─────────────────────────────────────────────────────────────────┐
│                    INTEGRATION ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  /internal/integrations/                                         │
│    ├── types.go          # Provider interface                   │
│    ├── registry.go       # Provider registry                    │
│    ├── oauth.go          # OAuth utilities                      │
│    ├── handler.go        # Unified HTTP handler                 │
│    └── providers/                                                │
│        ├── google/       # Google Workspace provider            │
│        ├── slack/        # Slack provider                       │
│        ├── microsoft/    # Microsoft 365 (future)               │
│        ├── notion/       # Notion (future)                      │
│        └── register.go   # Provider initialization              │
│                                                                  │
│  Each provider implements:                                       │
│    - Name(), DisplayName(), Category(), Icon()                  │
│    - GetAuthURL(), ExchangeCode(), RefreshToken()               │
│    - GetConnectionStatus(), Disconnect()                        │
│    - SaveToken(), GetToken()                                    │
│    - SupportsSync(), Sync()                                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### How OAuth Flow Works

1. User clicks "Connect" on integration page
2. Frontend calls `GET /api/integrations/oauth/{provider}/start`
3. Backend returns auth URL with state token
4. User is redirected to provider's consent screen
5. Provider redirects back with authorization code
6. Backend exchanges code for tokens
7. Tokens are stored encrypted in `credential_vault` table
8. User integration is created in `user_integrations` table

---

## Google Workspace Setup

### Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Note your **Project ID**

### Step 2: Enable APIs

Enable the following APIs in your project:

```
APIs to Enable:
├── Google Calendar API
├── Gmail API
├── People API (Contacts)
├── Google Tasks API
├── Google Drive API (optional)
└── Google OAuth2 API
```

To enable:
1. Go to "APIs & Services" > "Library"
2. Search for each API and click "Enable"

### Step 3: Configure OAuth Consent Screen

1. Go to "APIs & Services" > "OAuth consent screen"
2. Choose "External" user type (or Internal for Workspace)
3. Fill in required fields:
   - App name: "BusinessOS"
   - User support email: your email
   - Developer contact: your email
4. Add scopes:
   ```
   Required Scopes:
   ├── .../auth/userinfo.email
   ├── .../auth/userinfo.profile
   ├── .../auth/calendar.readonly
   ├── .../auth/calendar.events
   ├── .../auth/gmail.readonly
   ├── .../auth/gmail.send
   └── .../auth/gmail.modify
   ```
5. Add test users (during development)
6. Submit for verification (for production)

### Step 4: Create OAuth Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. Application type: "Web application"
4. Name: "BusinessOS Backend"
5. Add authorized redirect URIs:
   ```
   Development:
   http://localhost:8080/api/oauth/google/callback
   http://localhost:3000/api/oauth/google/callback

   Production:
   https://api.yourdomain.com/api/oauth/google/callback
   https://app.yourdomain.com/api/oauth/google/callback
   ```
6. Click "Create"
7. Download the JSON or copy Client ID and Client Secret

### Step 5: Set Environment Variables

```bash
# Google OAuth (Login)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8080/api/oauth/google/callback

# Google Integration OAuth (separate for integrations)
GOOGLE_INTEGRATION_REDIRECT_URI=http://localhost:8080/api/integrations/oauth/google/callback
```

### Verification Requirements (Production)

For production, Google requires verification for sensitive scopes:
- Gmail read/send requires verification
- Calendar write requires verification
- Process takes 2-4 weeks
- Prepare privacy policy and documentation

---

## Slack Setup

### Step 1: Create Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App"
3. Choose "From scratch"
4. Enter app name: "BusinessOS"
5. Select your workspace

### Step 2: Configure OAuth & Permissions

1. Go to "OAuth & Permissions"
2. Add Bot Token Scopes:
   ```
   Bot Scopes (what the app can do):
   ├── channels:read        # View channel info
   ├── channels:history     # View messages in public channels
   ├── chat:write           # Send messages
   ├── users:read           # View users
   ├── users:read.email     # View user emails
   ├── groups:read          # View private channels
   ├── groups:history       # View messages in private channels
   ├── im:read              # View direct messages
   ├── im:history           # View DM history
   ├── mpim:read            # View group DMs
   └── mpim:history         # View group DM history
   ```
3. Add User Token Scopes (optional):
   ```
   User Scopes (authorized by user):
   └── search:read          # Search messages
   ```
4. Add Redirect URLs:
   ```
   Development:
   http://localhost:8080/api/integrations/oauth/slack/callback

   Production:
   https://api.yourdomain.com/api/integrations/oauth/slack/callback
   ```

### Step 3: Get Credentials

1. Go to "Basic Information"
2. Under "App Credentials", copy:
   - Client ID
   - Client Secret
   - Signing Secret (for webhooks)

### Step 4: Set Environment Variables

```bash
SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_REDIRECT_URI=http://localhost:8080/api/integrations/oauth/slack/callback
SLACK_SIGNING_SECRET=your-signing-secret  # For webhooks
```

### Step 5: Install to Workspace (Development)

1. Go to "Install App"
2. Click "Install to Workspace"
3. Authorize the app

---

## Microsoft 365 Setup (Future)

### Coming Soon

Microsoft 365 integration will support:
- Outlook Calendar
- Outlook Mail
- OneDrive
- Microsoft Teams

### Preliminary Setup

1. Go to [Azure Portal](https://portal.azure.com)
2. Register an application in Azure AD
3. Configure redirect URIs
4. Add API permissions for Graph API

---

## Other Integrations

### Integration Roadmap

| Provider | Category | Status | Priority |
|----------|----------|--------|----------|
| Google Workspace | Productivity | Available | P0 |
| Slack | Communication | Available | P0 |
| Microsoft 365 | Productivity | Planned | P1 |
| Notion | Productivity | Planned | P1 |
| Linear | Project Mgmt | Planned | P2 |
| GitHub | Development | Planned | P2 |
| Fathom | Analytics | Planned | P2 |
| Discord | Communication | Planned | P3 |
| Zoom | Communication | Planned | P3 |
| Loom | Communication | Planned | P3 |

### Adding a New Provider

To add a new integration provider:

1. Create provider package:
   ```
   /internal/integrations/providers/{name}/
   └── provider.go
   ```

2. Implement the `Provider` interface:
   ```go
   type Provider interface {
       Name() string
       DisplayName() string
       Category() string
       Icon() string
       GetAuthURL(state string) string
       ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
       RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
       GetConnectionStatus(ctx context.Context, userID string) (*ConnectionStatus, error)
       Disconnect(ctx context.Context, userID string) error
       SaveToken(ctx context.Context, userID string, token *TokenResponse) error
       GetToken(ctx context.Context, userID string) (*Token, error)
       SupportsSync() bool
       Sync(ctx context.Context, userID string, options SyncOptions) (*SyncResult, error)
   }
   ```

3. Register in `providers/register.go`:
   ```go
   func InitializeProviders(pool *pgxpool.Pool) {
       // ... existing providers
       integrations.Register(newprovider.NewFromEnv(pool))
   }
   ```

4. Add environment variables to config
5. Update frontend integration settings UI

---

## Environment Variables

Complete list of integration-related environment variables:

```bash
# ============================================================================
# GOOGLE
# ============================================================================
GOOGLE_CLIENT_ID=                           # OAuth client ID
GOOGLE_CLIENT_SECRET=                       # OAuth client secret
GOOGLE_REDIRECT_URI=                        # Login OAuth redirect
GOOGLE_INTEGRATION_REDIRECT_URI=            # Integration OAuth redirect

# ============================================================================
# SLACK
# ============================================================================
SLACK_CLIENT_ID=                            # Slack app client ID
SLACK_CLIENT_SECRET=                        # Slack app client secret
SLACK_REDIRECT_URI=                         # OAuth redirect URI
SLACK_SIGNING_SECRET=                       # For webhook verification

# ============================================================================
# MICROSOFT (Future)
# ============================================================================
MICROSOFT_CLIENT_ID=                        # Azure AD app ID
MICROSOFT_CLIENT_SECRET=                    # Azure AD secret
MICROSOFT_TENANT_ID=                        # Azure AD tenant
MICROSOFT_REDIRECT_URI=                     # OAuth redirect

# ============================================================================
# NOTION (Future)
# ============================================================================
NOTION_CLIENT_ID=                           # Notion OAuth client ID
NOTION_CLIENT_SECRET=                       # Notion OAuth secret
NOTION_REDIRECT_URI=                        # OAuth redirect

# ============================================================================
# DATABASE (for credential storage)
# ============================================================================
DATABASE_URL=                               # PostgreSQL connection string
ENCRYPTION_KEY=                             # 32-byte key for encrypting tokens
```

---

## Troubleshooting

### Common Issues

#### "redirect_uri_mismatch" Error
- Ensure redirect URI in Google/Slack console exactly matches your environment variable
- Check for trailing slashes
- Ensure using correct protocol (http vs https)

#### "invalid_grant" Error
- Authorization code may have expired (valid for ~10 minutes)
- Code may have been used already
- Clock skew between servers

#### Token Not Refreshing
- Check refresh token is being stored
- Verify `offline` access was requested
- Some providers require re-authorization after long periods

#### Scopes Not Granted
- User may have denied some scopes
- Check `scopes` field in credential_vault
- May need to re-authorize to add new scopes

### Debug Commands

```bash
# Check provider registration
curl http://localhost:8080/api/integrations/providers

# Check connection status
curl http://localhost:8080/api/integrations/google/status

# Manually trigger sync
curl -X POST http://localhost:8080/api/integrations/google/sync

# View credential vault (dev only)
psql -c "SELECT user_id, provider_id, scopes, created_at FROM credential_vault"
```

### Logs to Check

```bash
# Backend logs for OAuth flow
grep -i "oauth\|token\|integration" logs/app.log

# Database logs for credential storage
grep -i "credential_vault\|user_integrations" logs/db.log
```

---

## Security Considerations

1. **Token Storage**: All tokens are encrypted at rest using AES-256
2. **State Parameter**: Always verify state to prevent CSRF
3. **Scopes**: Request minimum necessary scopes
4. **Token Refresh**: Implement automatic token refresh
5. **Revocation**: Properly revoke tokens on disconnect
6. **Audit Logging**: Log all OAuth events for security audit
