# BusinessOS Onboarding - Remaining Tasks

## Overview
This document tracks the remaining work to complete the conversational AI onboarding system.
UI components are complete. Backend integration and AI agent work remains.

---

## ✅ COMPLETED

### Frontend UI Components
- [x] Icon components (Send, Check, Arrow, Mic, Sun, Moon)
- [x] Animation components (PurpleOrb, TypewriterText, SequentialTypewriter, TypingIndicator)
- [x] Chat components (FloatingChatScreen, MessageBubble, ChatInput)
- [x] Feature components (ToolPicker, IntegrationCard, EmailInviteInput, FileUpload)
- [x] Fallback components (FallbackForm, CompletionScreen, MemberWelcome)
- [x] Main ConversationalOnboarding component
- [x] Component index exports
- [x] CSS variables for orb theming

---

## 🔄 IN PROGRESS / TODO

### 1. Database Schema (Priority: HIGH)
**Location:** `desktop/backend-go/internal/database/migrations/`

- [ ] Create migration file `0XX_onboarding_system.sql`:
  ```sql
  -- onboarding_sessions table
  -- workspace_onboarding_profiles table  
  -- onboarding_conversation_history table
  -- integration_pending_connections table
  ```
- [ ] Add fields to existing `workspaces` table:
  - `onboarding_completed_at`
  - `onboarding_data` (JSONB)
- [ ] Generate SQLC queries for onboarding operations

**Estimated effort:** 2-3 hours

---

### 2. Backend API Endpoints (Priority: HIGH)
**Location:** `desktop/backend-go/internal/handlers/`

Create new file `onboarding_handlers.go`:

- [ ] `POST /api/onboarding/sessions` - Start new session
- [ ] `GET /api/onboarding/sessions/:id` - Get session with conversation history
- [ ] `POST /api/onboarding/sessions/:id/messages` - Send message to AI
- [ ] `PUT /api/onboarding/sessions/:id/complete` - Mark complete
- [ ] `DELETE /api/onboarding/sessions/:id` - Abandon session
- [ ] `GET /api/onboarding/resume` - Check for resumable session
- [ ] `POST /api/onboarding/fallback` - Submit fallback form data

**Estimated effort:** 4-6 hours

---

### 3. Grok AI Integration (Priority: HIGH)
**Location:** `desktop/backend-go/internal/services/`

Create new file `onboarding_ai_service.go`:

- [ ] Grok API client setup (x.ai API)
- [ ] System prompt for onboarding agent
- [ ] Conversation context management
- [ ] Data extraction from responses
- [ ] Confidence scoring logic
- [ ] Response formatting

**Key prompt elements:**
```
- Extract: workspace_name, business_type, team_size, role, goals, challenges
- Ask follow-up questions naturally
- Return confidence scores per field
- Know when enough data is collected
```

**Estimated effort:** 6-8 hours

---

### 4. OAuth Integration Handlers (Priority: MEDIUM)
**Location:** `desktop/backend-go/internal/handlers/`

Update `integrations.go` or create `onboarding_oauth_handlers.go`:

- [ ] `GET /api/onboarding/oauth/:provider/start` - Initiate OAuth flow
- [ ] `GET /api/onboarding/oauth/:provider/callback` - Handle callback
- [ ] `GET /api/onboarding/integrations/status` - Get all integration statuses
- [ ] `DELETE /api/onboarding/integrations/:provider` - Disconnect

**Providers to support (existing infra):**
- [x] Google (already implemented)
- [x] Microsoft (already implemented)
- [x] Slack (already implemented)
- [x] Notion (already implemented)
- [x] Linear (already implemented)
- [x] HubSpot (already implemented)
- [x] Airtable (already implemented)
- [x] ClickUp (already implemented)
- [x] Fathom (already implemented)

**Work needed:** Wire existing providers to onboarding flow

**Estimated effort:** 3-4 hours

---

### 5. Frontend API Integration (Priority: HIGH)
**Location:** `frontend/src/lib/`

- [ ] Create `services/onboarding.ts`:
  - API client for onboarding endpoints
  - Session management functions
  - OAuth flow handlers

- [ ] Update `ConversationalOnboarding.svelte`:
  - Replace mock `getAgentResponse()` with real API call
  - Add session persistence (localStorage + API)
  - Implement resume flow on mount
  - Add OAuth popup handling

- [ ] Create onboarding route:
  - `routes/onboarding/+page.svelte` - Update to use ConversationalOnboarding
  - `routes/onboarding/+page.server.ts` - Server-side session check
  - `routes/onboarding/callback/+page.svelte` - OAuth callback handler

**Estimated effort:** 4-5 hours

---

### 6. Session Management (Priority: MEDIUM)
**Location:** Backend + Frontend

- [ ] Session expiration (24 hours per architecture)
- [ ] Resume logic (show last 3 exchanges)
- [ ] Low confidence counter (2 attempts → fallback)
- [ ] Workspace creation on completion
- [ ] Member vs Admin flow differentiation

**Estimated effort:** 3-4 hours

---

### 7. Team Invites Integration (Priority: LOW)
**Location:** Backend + Frontend

- [ ] Wire `EmailInviteInput` to existing invite system
- [ ] Send invite emails on onboarding completion
- [ ] `MemberWelcome` screen for invited users

**Estimated effort:** 2-3 hours

---

### 8. Data Sync After OAuth (Priority: MEDIUM)
**Location:** `desktop/backend-go/internal/workers/`

- [ ] Background job to sync initial data after OAuth connect:
  - Google: Calendar events, Contacts
  - Slack: Channels, Users
  - Notion: Workspaces, Pages
  - Linear: Projects, Issues
  - etc.

**Estimated effort:** 4-6 hours (per integration)

---

### 9. Testing (Priority: MEDIUM)

- [ ] Unit tests for onboarding service
- [ ] Integration tests for API endpoints
- [ ] E2E tests for full onboarding flow
- [ ] Test OAuth flows with mock providers

**Estimated effort:** 4-6 hours

---

### 10. Polish & Edge Cases (Priority: LOW)

- [ ] Error handling UI (network errors, API failures)
- [ ] Loading states during API calls
- [ ] Keyboard navigation (Enter to send, Escape to go back)
- [ ] Mobile responsiveness
- [ ] Analytics tracking (funnel, drop-off points)
- [ ] A/B test hooks for conversation variations

**Estimated effort:** 3-4 hours

---

## Summary

| Category | Tasks | Est. Hours |
|----------|-------|------------|
| Database | 3 | 2-3 |
| Backend API | 7 | 4-6 |
| Grok AI | 6 | 6-8 |
| OAuth Wiring | 4 | 3-4 |
| Frontend Integration | 6 | 4-5 |
| Session Management | 5 | 3-4 |
| Team Invites | 3 | 2-3 |
| Data Sync | Variable | 4-6 per integration |
| Testing | 4 | 4-6 |
| Polish | 6 | 3-4 |
| **TOTAL** | ~44+ tasks | ~35-50 hours |

---

## Recommended Order of Implementation

1. **Database schema** - Foundation for everything
2. **Backend API endpoints** - Basic CRUD without AI
3. **Frontend API integration** - Wire up to backend
4. **Grok AI integration** - Add intelligence
5. **OAuth wiring** - Connect integrations
6. **Session management** - Resume, expiration
7. **Testing** - Validate everything
8. **Polish** - Error handling, edge cases
9. **Team invites** - Nice to have
10. **Data sync** - Background enhancement

---

## Environment Variables Needed

```env
# Grok AI (x.ai)
XAI_API_KEY=
XAI_MODEL=grok-beta

# OAuth (if not already configured)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
FATHOM_CLIENT_ID=
FATHOM_CLIENT_SECRET=
```

---

## Questions to Resolve

1. **Grok API access** - Do we have x.ai API credentials?
2. **OAuth credentials** - Which providers are already configured?
3. **Workspace limit** - Any limit on workspaces per user during onboarding?
4. **Analytics** - What events should we track?
5. **Feature flags** - Should we gate this behind a flag initially?
