# BusinessOS Onboarding - System Overview

## Overview
This document tracks the conversational AI onboarding system for BusinessOS.
The core onboarding flow is now **fully functional** with AI-powered conversations, 
personalized dashboard generation, and session management.

**Last Updated:** January 2026

---

## 🎉 CURRENT STATE: PRODUCTION-READY

### What Works Today
- ✅ Full conversational AI onboarding flow (5 questions)
- ✅ Hybrid UI: Chips for quick-select + Chat input for open questions
- ✅ Multi-provider AI support (Grok, OpenAI, Anthropic, Groq)
- ✅ Deterministic fallback when no AI configured
- ✅ Session persistence and resume capability
- ✅ Welcome back message for returning users
- ✅ Personalized dashboard creation from onboarding data
- ✅ Smart integration recommendations based on answers
- ✅ Raw user input preserved (not just normalized values)
- ✅ Natural conversational transitions (AI-generated messages)

---

## 🔒 USER FLOW LOGIC (DECIDED)

### Who Gets Onboarding?

**1. New User (Signs Up Fresh)**
- Creates their own account with no existing workspace
- Needs **full onboarding**: company name, business type, team size, role, challenge, integrations
- Becomes the **workspace owner/admin**
- All settings saved to their workspace profile

**2. Invited User (Accepts Invite Link)**
- Joins an **existing** workspace
- Company settings already exist (name, type, integrations, etc.)
- They **inherit** the workspace's configuration
- Role is assigned by whoever invited them
- **Skip full onboarding** → Show a "Member Welcome" screen instead:
  - "Welcome to [Company Name]!"
  - "You've been added as a [Role]."
  - Quick personal setup: display name, profile pic, notification prefs
  - Then straight to dashboard

### Routing Logic (After Login)
```
User logs in
    ↓
Check: Are they a member of any workspace?
    ├── NO  → Full Onboarding (creating new workspace)
    └── YES → Check: Is their user profile complete?
                ├── NO  → Quick Personal Setup
                └── YES → Redirect to /window (Dashboard)
```

### After Onboarding Redirect
- **Current bug:** 404 after completing onboarding
- **Expected:** Redirect to `/window` route (same as login/signin destination)

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
- [x] Progress indicator (dots)
- [x] Skip button (softened - "Skip for now")
- [x] Orb thinking animation
- [x] Auto-focus input
- [x] Voice input button (UI only)

### Conversational Flow Improvements (NEW)
- [x] Welcome back message when resuming session
- [x] Context display ("Setting up [Company] - [Type]")
- [x] Continue button before resuming questions
- [x] AI-generated messages prioritized over static text
- [x] Natural transition delays between questions
- [x] Final message shown before integrations screen
- [x] Skip button hidden during intro/resume states

### Dashboard Personalization
**Status: REMOVED** - Dashboard creation is handled by Agent Skills dashboard tool instead of onboarding.
Onboarding data (business type, challenge, integrations) is stored in `workspace_onboarding_profiles` and can be used by the agent to suggest personalized dashboards.

---

## 🔄 REMAINING WORK

### 1. OAuth Integration During Onboarding (Priority: MEDIUM)
**Location:** `desktop/backend-go/internal/handlers/`

Currently, integration cards are shown but OAuth flows may need polish:

- [ ] Test each provider's OAuth flow from onboarding screen
- [ ] Verify callback handles correctly during onboarding context
- [ ] Add loading states during OAuth popup
- [ ] Handle OAuth failures gracefully
- [ ] Persist connected integrations to workspace profile

**Estimated effort:** 3-4 hours

---

### 2. Session Management Polish (Priority: MEDIUM)

- [x] Session resume with context (Welcome back message)
- [ ] Session expiration cleanup (24 hours)
- [ ] Low confidence counter (2 attempts → fallback form)
- [ ] Member vs Admin flow differentiation

**Estimated effort:** 2-3 hours

---

### 3. Team Invites Integration (Priority: LOW)

- [ ] Wire `EmailInviteInput` to existing invite system
- [ ] Send invite emails on onboarding completion
- [ ] `MemberWelcome` screen for invited users

**Estimated effort:** 2-3 hours

---

### 4. Data Sync After OAuth (Priority: LOW)

- [ ] Background job to sync initial data after OAuth connect
- [ ] Per-integration sync workers

**Estimated effort:** 4-6 hours (per integration)

---

### 5. Testing (Priority: MEDIUM)

- [ ] Unit tests for onboarding service
- [ ] Unit tests for dashboard builder
- [ ] Integration tests for API endpoints
- [ ] E2E tests for full onboarding flow

**Estimated effort:** 4-6 hours

---

### 6. Polish & Edge Cases (Priority: LOW)

- [ ] Error handling UI improvements
- [ ] Keyboard navigation
- [ ] Mobile responsiveness
- [ ] Voice input implementation (currently UI only)

**Estimated effort:** 3-4 hours

---

## 📊 Summary

| Category | Status | Notes |
|----------|--------|-------|
| Frontend UI | ✅ Complete | All components working |
| Conversational Flow | ✅ Complete | AI responses, transitions, resume |
| Backend API | ✅ Complete | All endpoints working |
| AI Integration | ✅ Complete | Multi-provider support |
| Data Persistence | ✅ Complete | Raw + normalized values stored |
| OAuth Wiring | 🔄 Partial | UI exists, needs testing |
| Session Management | 🔄 Partial | Resume works, expiration TODO |
| Testing | ⏳ TODO | Not started |

**Total Completed:** ~75%  
**Remaining Effort:** ~12-16 hours

---

## 💡 SUGGESTIONS FOR IMPROVEMENT

### High Value / Quick Wins
1. **Add completion celebration** - Confetti or animation when onboarding finishes
2. **Progress bar instead of dots** - Show "Step 2 of 5" for clearer progress
3. **Undo/go back** - Let users correct previous answers
4. **Typewriter effect on all AI messages** - More conversational feel

### Medium Effort
5. **Onboarding analytics** - Track completion rates, drop-off points
6. **A/B test question order** - Find optimal conversion flow
7. **Smart defaults** - Pre-fill based on email domain (company name from @acme.com)
8. **Integration sync preview** - Show what data will be imported before connecting

### Future Enhancements
9. **Voice input** - Implement actual speech-to-text (UI exists)
10. **Multi-language support** - Translate onboarding questions
11. **Custom onboarding per plan** - Different flows for Enterprise vs Free
12. **Onboarding templates** - Industry-specific quick setups

---

## 🗂️ KEY FILES

### Backend
- `desktop/backend-go/internal/services/onboarding_service.go` - Main service
- `desktop/backend-go/internal/services/onboarding_ai_service.go` - AI integration
- `desktop/backend-go/internal/services/onboarding_validation.go` - Input validation
- `desktop/backend-go/internal/handlers/onboarding_handlers.go` - API endpoints

### Frontend  
- `frontend/src/lib/components/onboarding/ConversationalOnboarding.svelte` - Main component
- `frontend/src/lib/api/onboarding/index.ts` - API client
- `frontend/src/routes/onboarding/+page.svelte` - Route page

### Database
- `desktop/backend-go/internal/database/migrations/045_onboarding_system.sql` - Schema

---

## 🔧 ENVIRONMENT VARIABLES

See `.env.example` in `desktop/backend-go/` for full list. Key requirements:

| Variable | Required | Purpose |
|----------|----------|---------|
| `XAI_API_KEY` or `OPENAI_API_KEY` | ✅ | AI provider for conversations |
| `DATABASE_URL` | ✅ | PostgreSQL connection |
| `SECRET_KEY` | ✅ | JWT signing |
| `REDIS_URL` | ✅ | Session storage |

---

## ❓ OPEN QUESTIONS

1. **Completion celebration** - Add confetti/animation?
2. **Analytics tracking** - PostHog integration for funnel analysis?
3. **A/B testing** - Test different question orders?
4. **Mobile app** - Same onboarding for React Native?
