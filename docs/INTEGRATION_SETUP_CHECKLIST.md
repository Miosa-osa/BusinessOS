# Integration Setup Checklist - What You Need to Get

## Overview

For users to connect their tools to BusinessOS, **you** (the developer) need to create OAuth apps / API credentials in each provider's developer portal. Users will then authorize YOUR app to access their data.

---

## PRIORITY 0: Already Partially Set Up

### Google (Login + Calendar + Drive + Gmail)

**Status:** Config exists, need to verify scopes

**Portal:** https://console.cloud.google.com/apis/credentials

**Steps:**
1. Go to Google Cloud Console
2. Select your project (or create one)
3. Go to "APIs & Services" > "Credentials"
4. Create "OAuth 2.0 Client ID" (Web Application)
5. Add authorized redirect URIs:
   - `http://localhost:8001/api/auth/google/callback/login` (dev)
   - `http://localhost:8001/api/oauth/google/callback` (dev - new unified)
   - `https://your-backend.run.app/api/auth/google/callback/login` (prod)
   - `https://your-backend.run.app/api/oauth/google/callback` (prod)
6. Enable these APIs (APIs & Services > Library):
   - Google Calendar API
   - Google Drive API
   - Gmail API
   - Google People API

**You'll get:**
```
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxx
```

**Scopes needed:**
```
openid
email
profile
https://www.googleapis.com/auth/calendar
https://www.googleapis.com/auth/calendar.events
https://www.googleapis.com/auth/drive.readonly
https://www.googleapis.com/auth/gmail.readonly
https://www.googleapis.com/auth/gmail.send
```

---

## PRIORITY 1: Need to Set Up

### Slack

**Status:** Config exists in code, needs credentials

**Portal:** https://api.slack.com/apps

**Steps:**
1. Go to https://api.slack.com/apps
2. Click "Create New App" > "From scratch"
3. Name: "BusinessOS" (or your app name)
4. Select workspace for development
5. Go to "OAuth & Permissions"
6. Add redirect URLs:
   - `http://localhost:8001/api/oauth/slack/callback` (dev)
   - `https://your-backend.run.app/api/oauth/slack/callback` (prod)
7. Add Bot Token Scopes:
   - `channels:read` - View basic channel info
   - `channels:history` - View messages in public channels
   - `chat:write` - Send messages
   - `users:read` - View users
   - `users:read.email` - View user emails
8. Add User Token Scopes (if needed):
   - `identity.basic`
   - `identity.email`
9. Go to "Basic Information" and get credentials

**You'll get:**
```
SLACK_CLIENT_ID=xxxxx.xxxxx
SLACK_CLIENT_SECRET=xxxxx
SLACK_SIGNING_SECRET=xxxxx  # For webhook verification
```

---

### Notion

**Status:** Config exists in code, needs credentials

**Portal:** https://www.notion.so/my-integrations

**Steps:**
1. Go to https://www.notion.so/my-integrations
2. Click "New integration"
3. Name: "BusinessOS"
4. Select workspace
5. Set type: "Public" (so any user can connect)
6. Add redirect URI:
   - `http://localhost:8001/api/oauth/notion/callback` (dev)
   - `https://your-backend.run.app/api/oauth/notion/callback` (prod)
7. Set capabilities:
   - Read content
   - Update content
   - Insert content
   - Read user information (with email)

**You'll get:**
```
NOTION_CLIENT_ID=xxxxx
NOTION_CLIENT_SECRET=secret_xxxxx
```

**Note:** Notion uses OAuth 2.0 with their own flow. Users must also share specific pages/databases with your integration.

---

### HubSpot

**Status:** NOT YET IMPLEMENTED - High Priority

**Portal:** https://developers.hubspot.com/

**Steps:**
1. Go to https://developers.hubspot.com/
2. Click "Create app"
3. Fill in app info
4. Go to "Auth" tab
5. Add redirect URL:
   - `http://localhost:8001/api/oauth/hubspot/callback` (dev)
   - `https://your-backend.run.app/api/oauth/hubspot/callback` (prod)
6. Select scopes:
   - `crm.objects.contacts.read`
   - `crm.objects.contacts.write`
   - `crm.objects.deals.read`
   - `crm.objects.deals.write`
   - `crm.objects.companies.read`

**You'll get:**
```
HUBSPOT_CLIENT_ID=xxxxx
HUBSPOT_CLIENT_SECRET=xxxxx
HUBSPOT_APP_ID=xxxxx  # Numeric app ID
```

---

### Linear

**Status:** NOT YET IMPLEMENTED

**Portal:** https://linear.app/settings/api

**Steps:**
1. Go to https://linear.app/YOUR-WORKSPACE/settings/api
2. Click "Create new OAuth application"
3. Name: "BusinessOS"
4. Add redirect URI:
   - `http://localhost:8001/api/oauth/linear/callback`
5. Select scopes:
   - `read` - Read issues, projects, etc.
   - `write` - Create/update issues
   - `issues:create`

**You'll get:**
```
LINEAR_CLIENT_ID=xxxxx
LINEAR_CLIENT_SECRET=lin_oauth_xxxxx
```

---

### ClickUp

**Status:** NOT YET IMPLEMENTED

**Portal:** https://app.clickup.com/settings/apps

**Steps:**
1. Go to ClickUp Settings > Integrations > ClickUp API
2. Create new app
3. Add redirect URI
4. Select permissions

**You'll get:**
```
CLICKUP_CLIENT_ID=xxxxx
CLICKUP_CLIENT_SECRET=xxxxx
```

---

## PRIORITY 2: Future Integrations

### GitHub

**Portal:** https://github.com/settings/developers

```
GITHUB_CLIENT_ID=xxxxx
GITHUB_CLIENT_SECRET=xxxxx
```

### Stripe

**Portal:** https://dashboard.stripe.com/apikeys

```
STRIPE_SECRET_KEY=sk_live_xxxxx  # Or sk_test_ for dev
STRIPE_PUBLISHABLE_KEY=pk_live_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
```

### Zoom

**Portal:** https://marketplace.zoom.us/develop/create

```
ZOOM_CLIENT_ID=xxxxx
ZOOM_CLIENT_SECRET=xxxxx
```

### Calendly

**Portal:** https://developer.calendly.com/

```
CALENDLY_CLIENT_ID=xxxxx
CALENDLY_CLIENT_SECRET=xxxxx
```

### Fathom (Meeting Notes)

**Portal:** https://fathom.video/settings/integrations

```
FATHOM_API_KEY=xxxxx
```

---

## AI Provider Keys (You Already Have Some)

These are for Sorx skills to call AI models:

### Anthropic (Claude)
**Portal:** https://console.anthropic.com/settings/keys
```
ANTHROPIC_API_KEY=sk-ant-xxxxx
```

### OpenAI
**Portal:** https://platform.openai.com/api-keys
```
OPENAI_API_KEY=sk-xxxxx
```

### Groq
**Portal:** https://console.groq.com/keys
```
GROQ_API_KEY=gsk_xxxxx
```

---

## Security Keys You Need to Generate

These are NOT from external services - you generate them yourself:

```bash
# Generate a 32-byte key for token encryption
openssl rand -base64 32
# Example output: K7gNU3sdo+OL0wNhqoVWhr3g6s1xYv72ol/pe/Unols=

# Generate a secret key for JWT signing
openssl rand -base64 64
# Example output: long-random-string...

# Generate HMAC secret for Redis
openssl rand -base64 32
```

**Add to .env:**
```
TOKEN_ENCRYPTION_KEY=<32-byte-base64-from-above>
SECRET_KEY=<64-byte-base64-from-above>
REDIS_KEY_HMAC_SECRET=<32-byte-base64-from-above>
```

---

## Complete .env Template for Integrations

Add these to your `desktop/backend-go/.env`:

```bash
# ===========================================
# SECURITY (REQUIRED - Generate these!)
# ===========================================
SECRET_KEY=<generate-with-openssl>
TOKEN_ENCRYPTION_KEY=<generate-with-openssl>
REDIS_KEY_HMAC_SECRET=<generate-with-openssl>

# ===========================================
# GOOGLE (Required for login + integrations)
# ===========================================
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxx
GOOGLE_REDIRECT_URI=http://localhost:8001/api/auth/google/callback

# ===========================================
# SLACK
# ===========================================
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_SIGNING_SECRET=

# ===========================================
# NOTION
# ===========================================
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=

# ===========================================
# HUBSPOT
# ===========================================
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=

# ===========================================
# LINEAR
# ===========================================
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=

# ===========================================
# CLICKUP
# ===========================================
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=

# ===========================================
# AI PROVIDERS (for Sorx skills)
# ===========================================
ANTHROPIC_API_KEY=
OPENAI_API_KEY=
GROQ_API_KEY=
```

---

## Quick Action Checklist

### Today (Before Security Fixes):
- [ ] Generate security keys (TOKEN_ENCRYPTION_KEY, SECRET_KEY, REDIS_KEY_HMAC_SECRET)
- [ ] Verify Google OAuth is set up with correct redirect URIs
- [ ] Add .env entries for new security keys

### This Week:
- [ ] Create Slack app at https://api.slack.com/apps
- [ ] Create Notion integration at https://www.notion.so/my-integrations
- [ ] Create HubSpot app at https://developers.hubspot.com/

### Next Week:
- [ ] Linear, ClickUp, GitHub integrations
- [ ] Stripe for payment processing

---

## Testing OAuth Flows

After setting up each provider, test the flow:

```bash
# 1. Start backend
cd desktop/backend-go && go run main.go

# 2. Test OAuth initiation
curl http://localhost:8001/api/oauth/slack/init

# Should return:
# {"auth_url": "https://slack.com/oauth/v2/authorize?client_id=..."}

# 3. Open that URL in browser
# 4. Authorize the app
# 5. Should redirect back to your callback URL
```

---

## Costs

Most of these are **FREE** for development:

| Provider | Free Tier | Paid When |
|----------|-----------|-----------|
| Google | Free | High volume |
| Slack | Free | Enterprise features |
| Notion | Free | - |
| HubSpot | Free | CRM features |
| Linear | Free | Team features |
| Anthropic | Pay per use | Always |
| OpenAI | Pay per use | Always |
| Stripe | Free + 2.9% per transaction | - |

---

**Document Version:** 1.0
**Created:** 2026-01-05
