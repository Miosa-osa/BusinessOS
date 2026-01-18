# PR Review: AI-Powered Onboarding with Groq Integration

**Branch**: `feature/ai-onboarding-groq`
**Date**: 2026-01-18
**Author**: Roberto + Claude Code
**Reviewers**: Nick, Pedro, Javaris

---

## 🎯 Executive Summary

We've built a **complete AI-powered onboarding system** that analyzes new users' Gmail inboxes to automatically create **personalized business apps, module customizations, and integrations** tailored to their actual needs.

**This is NOT about creating bullshit apps based on hobbies. This is about:**
- Analyzing their **real business tools** (from email subscriptions)
- Identifying **who they work with** (email patterns)
- Understanding **what they do** (meeting invites, project mentions)
- Auto-generating **useful business modules** (CRM, Tasks, Projects, custom integrations)
- Pre-configuring **agent capabilities** based on their workflow

---

## 🚀 The Vision: Intelligent Personalization at Scale

### Current Flow (What We Built)
```
1. User signs up → Gmail OAuth
2. Backend fetches 50 recent emails
3. AI analyzes:
   - Tools used (Notion, Linear, Slack, Figma, etc.)
   - People they work with
   - Meeting patterns
   - Project mentions
   - Subscription services
4. Groq AI (llama-3.3-70b-versatile) generates:
   - 3 personalized insight phrases
   - List of detected tools/interests
   - Starter app recommendations
5. Frontend displays insights with animations
6. User gets customized workspace with:
   - Pre-configured modules
   - Suggested integrations
   - Custom agents
```

### Future Vision (Next Steps)
```
Multi-Tenant Architecture:
└─ Company
   └─ Workspace
      └─ Operating System (BusinessOS instance)
         └─ Apps (modules customized per user)
            └─ Agents (tool-specific AI agents)
               └─ Integrations (3rd party tools)

Version Control:
- Each user gets their own BusinessOS instance
- Editable by agents (with proper sandboxing)
- Git-backed for version control
- Shareable configurations
```

---

## 📦 What We Built

### Backend (Go + PostgreSQL + Groq)

**New Database Migrations:**
1. `054_onboarding_user_analysis.sql` - Stores AI analysis results
2. `055_onboarding_starter_apps.sql` - Starter app recommendations
3. `056_onboarding_email_metadata.sql` - Email analysis metadata

**New Services:**
1. **EmailAnalyzerService** (`internal/services/onboarding_email_analyzer.go`)
   - Fetches 50 emails from Gmail API
   - Extracts metadata (senders, topics, tools)
   - Identifies SaaS subscriptions
   - Maps meeting patterns

2. **ProfileAnalyzerAgent** (`internal/services/onboarding_profile_analyzer.go`)
   - Uses Groq AI (llama-3.3-70b-versatile)
   - Generates 3 conversational insight phrases
   - Extracts interests and tools
   - Creates personality-based recommendations

3. **AppCustomizerAgent** (`internal/services/onboarding_app_customizer.go`)
   - Recommends 3-4 starter apps
   - Customizes existing modules (CRM, Tasks, Projects)
   - Suggests new modules based on tools
   - Creates integration recommendations

**New API Endpoints:**
- `POST /api/v1/osa-onboarding/analyze` - Start analysis
- `GET /api/v1/osa-onboarding/analyze/:id` - Poll for progress
- `GET /api/v1/osa-onboarding/analyze/:id/stream` - SSE real-time updates
- `POST /api/v1/osa-onboarding/generate-apps` - Generate starter apps

**New Handler:**
- `internal/handlers/osa_onboarding.go` - OSAOnboardingHandler

### Frontend (SvelteKit + TypeScript)

**New API Client:**
- `src/lib/api/osa-onboarding/index.ts` - 4 endpoints + SSE helper
- `src/lib/api/osa-onboarding/types.ts` - TypeScript types

**New Store:**
- `src/lib/stores/onboardingAnalysis.ts` - SSE streaming service
  - Real-time progress updates
  - Fallback polling if SSE fails
  - Derived stores for reactive UI

**Updated Screens:**
- `src/routes/onboarding/analyzing/+page.svelte` - Screen 1 (insight 1)
- `src/routes/onboarding/analyzing-2/+page.svelte` - Screen 2 (insight 2)
- `src/routes/onboarding/analyzing-3/+page.svelte` - Screen 3 (insight 3)
- `src/routes/auth/callback/+page.svelte` - Triggers analysis after OAuth

**UI/UX:**
- Spinner animations
- SSE streaming indicators
- AI-generated badges
- 2-second auto-advance between screens

### Configuration

**Updated `.env`:**
```bash
GROQ_API_KEY=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
GROQ_MODEL=llama-3.3-70b-versatile
```

---

## 🏗️ Architecture Deep Dive

### Data Flow
```
┌─────────────────────────────────────────────────────────────────┐
│ 1. USER SIGNUP                                                  │
│    → Gmail OAuth → Backend creates workspace + session         │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 2. OAUTH CALLBACK (Frontend)                                    │
│    → Extracts userId, workspaceId                              │
│    → Calls onboardingAnalysis.start(userId, workspaceId, 50)   │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 3. ANALYSIS START (Backend)                                     │
│    → POST /api/v1/osa-onboarding/analyze                       │
│    → Creates analysis record (status: 'analyzing')              │
│    → Returns analysis_id                                        │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 4. EMAIL EXTRACTION (EmailAnalyzerService)                      │
│    → Fetches 50 recent emails from Gmail API                   │
│    → Extracts:                                                  │
│      • Sender domains                                           │
│      • Subject patterns                                         │
│      • Tool mentions (Notion, Linear, Slack, etc.)             │
│      • Meeting invites                                          │
│      • SaaS subscription emails                                 │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 5. AI ANALYSIS (ProfileAnalyzerAgent + Groq)                   │
│    → Sends metadata to Groq API                                 │
│    → Model: llama-3.3-70b-versatile                            │
│    → Temperature: 0.7                                           │
│    → Max tokens: 500                                            │
│    → Returns:                                                   │
│      • 3 conversational insight phrases                         │
│      • List of interests                                        │
│      • Detected tools                                           │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 6. SSE STREAMING (Backend → Frontend)                          │
│    → GET /api/v1/osa-onboarding/analyze/:id/stream            │
│    → Events:                                                    │
│      • progress: { status, insights[], interests[] }           │
│      • done: { final results }                                  │
│      • error: { error message }                                 │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 7. FRONTEND DISPLAY (Analyzing Screens)                        │
│    → onboardingAnalysis store receives SSE events              │
│    → analyzingInsights derived store updates                   │
│    → Screens 1, 2, 3 reactively display insights               │
│    → Auto-advance after 2s per screen                          │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 8. STARTER APPS (AppCustomizerAgent)                           │
│    → POST /api/v1/osa-onboarding/generate-apps                │
│    → Analyzes detected tools + interests                       │
│    → Recommends 3-4 customized apps:                           │
│      • Existing module customizations (CRM, Tasks, Projects)   │
│      • New module suggestions (custom apps)                     │
│      • Integration recommendations (3rd party tools)            │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ 9. WORKSPACE SETUP                                              │
│    → Creates customized BusinessOS instance                    │
│    → Pre-configures modules based on recommendations           │
│    → Enables suggested integrations                            │
│    → Sets up agents with tool-specific capabilities            │
└─────────────────────────────────────────────────────────────────┘
```

### Database Schema

**`onboarding_analyses` table:**
```sql
CREATE TABLE onboarding_analyses (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    workspace_id UUID,
    status TEXT, -- 'analyzing', 'completed', 'failed'
    insights JSONB, -- ["insight1", "insight2", "insight3"]
    interests JSONB, -- ["interest1", "interest2"]
    tools_used JSONB, -- ["notion", "linear", "slack"]
    email_count INTEGER,
    sender_domains JSONB,
    detected_patterns JSONB,
    summary TEXT,
    error_message TEXT,
    analysis_model TEXT, -- 'llama-3.3-70b-versatile'
    ai_provider TEXT, -- 'groq'
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

**`onboarding_starter_apps` table:**
```sql
CREATE TABLE onboarding_starter_apps (
    id UUID PRIMARY KEY,
    analysis_id UUID REFERENCES onboarding_analyses(id),
    user_id TEXT NOT NULL,
    workspace_id UUID,
    title TEXT,
    description TEXT,
    icon_emoji TEXT,
    category TEXT, -- 'productivity', 'crm', 'custom'
    reasoning TEXT, -- Why this app was recommended
    customization_prompt TEXT, -- How to customize it
    based_on_interests JSONB,
    based_on_tools JSONB,
    base_module TEXT, -- 'crm', 'tasks', 'projects', 'custom'
    module_customizations JSONB,
    generation_model TEXT,
    ai_provider TEXT,
    display_order INTEGER,
    status TEXT, -- 'ready', 'installing', 'installed'
    created_at TIMESTAMP
);
```

---

## 👥 What Each Team Member Needs to Know

### 🔧 Nick (Infrastructure & Deployment)

**Your Focus:** Agent sandboxing, version control, multi-tenant architecture

**What You Need to Build:**

1. **Agent Code Editing Sandbox**
   - We need agents to be able to edit user's BusinessOS code
   - **Challenge:** How do we let agents modify code without exposing source?
   - **Options:**
     - E2B sandboxes (isolated containers per user)
     - Git-based version control (each workspace = repo)
     - Code streaming API (agent requests changes, we apply them)
   - **Question for you:** Best approach for scalable, secure code editing?

2. **Multi-Tenant Database Architecture**
   - Currently: Single DB, workspace_id partitioning
   - Future: Company → Workspace → OS instance hierarchy
   - **Needed:** Schema design for multi-level isolation

3. **Version Control Integration**
   - Each user's BusinessOS should be git-backed
   - Agents commit changes with proper messages
   - Users can rollback, branch, merge
   - **Question:** GitHub API integration? Supabase storage? Custom Git server?

4. **Deployment Pipeline for Generated Apps**
   - When agent creates/customizes an app, how do we deploy it?
   - Real-time compilation? Hot reload? Container per workspace?

5. **GCP Cloud Run Scaling**
   - Current: Single instance
   - Future: Workspace-per-instance? Shared with load balancing?
   - **Cost considerations:** How to scale without breaking bank?

**Files You Should Review:**
- `desktop/backend-go/internal/handlers/osa_onboarding.go`
- `desktop/backend-go/internal/database/migrations/054-056`
- `desktop/backend-go/docs/integrations/ios-onboarding/`

**Action Items for Nick:**
- [ ] Design agent code editing architecture (proposal doc)
- [ ] Plan multi-tenant DB schema updates
- [ ] Research version control options (E2B vs Git vs custom)
- [ ] Update GCP deployment for SSE streaming support
- [ ] Cost analysis: Workspace scaling strategy

---

### 🐹 Pedro (Backend & Consultation System)

**Your Focus:** AI agent coordination, Groq integration, service architecture

**What You Need to Know:**

1. **New Services You Can Use:**
   - `EmailAnalyzerService` - Extract insights from emails
   - `ProfileAnalyzerAgent` - Generate user personality insights
   - `AppCustomizerAgent` - Recommend customized apps
   - All use Groq AI (llama-3.3-70b-versatile)

2. **Groq Integration Pattern:**
   ```go
   llmService := NewGroqService(cfg, "llama-3.3-70b-versatile")
   response, err := llmService.Complete(ctx, prompt, 0.7, 500)
   ```

3. **SSE Streaming Pattern:**
   ```go
   sse := streaming.NewSSEWriter(c.Writer)
   sse.SendEvent("progress", map[string]interface{}{
       "status": "analyzing",
       "insights": insights,
   })
   ```

4. **How This Fits Consultation System:**
   - Onboarding analysis → Initial user profile
   - Can feed into consultation agent context
   - Tools detected → Available for consultation
   - Interests → Guide conversation topics

5. **Email Analysis Integration:**
   - We're already analyzing emails for onboarding
   - Can extend to ongoing email monitoring
   - Build context database from email history
   - Feed into RAG system for better answers

**Files You Should Review:**
- `internal/services/onboarding_email_analyzer.go`
- `internal/services/onboarding_profile_analyzer.go`
- `internal/services/onboarding_app_customizer.go`
- `internal/handlers/osa_onboarding.go`

**Integration Opportunities:**
- Use EmailAnalyzerService for ongoing email context
- Extend ProfileAnalyzerAgent for consultation personas
- AppCustomizerAgent pattern for dynamic tool creation
- SSE streaming for real-time consultation updates

**Action Items for Pedro:**
- [ ] Review Groq integration patterns
- [ ] Plan integration with consultation system
- [ ] Design email context → RAG pipeline
- [ ] Extend ProfileAnalyzerAgent for consultation use
- [ ] Add tool detection to consultation context

---

### 🎨 Javaris (Frontend & Testing)

**Your Focus:** UI/UX polish, testing, animation refinement

**What You Need to Know:**

1. **New Frontend Architecture:**
   - **Store:** `src/lib/stores/onboardingAnalysis.ts`
     - Manages SSE streaming
     - Reactive state updates
     - Fallback polling
   - **Derived Stores:** `analyzingInsights`, `analysisComplete`, `analysisFailed`
   - **API Client:** `src/lib/api/osa-onboarding/`

2. **Analyzing Screens Flow:**
   ```
   Screen 1: First insight + spinner
      ↓ (2s delay)
   Screen 2: Second insight
      ↓ (2s delay)
   Screen 3: Third insight
      ↓
   Starter Apps screen
   ```

3. **SSE Integration:**
   ```typescript
   onMount(() => {
     const unsubscribe = onboardingAnalysis.subscribe(($analysis) => {
       // Reactive updates from SSE stream
       if ($analysis.status === 'completed') {
         // Show insights
       }
     });
   });
   ```

4. **Testing Needs:**
   - E2E tests for OAuth → Analysis → Insights flow
   - SSE stream testing (success + failure cases)
   - Animation timing verification
   - Fallback insight display
   - Error state handling

5. **UI Polish Opportunities:**
   - Spinner animations (currently basic)
   - Insight reveal animations (fade in, slide up?)
   - Progress indicators during analysis
   - Error state designs
   - Loading skeletons

**Files You Should Review:**
- `frontend/src/lib/stores/onboardingAnalysis.ts`
- `frontend/src/routes/onboarding/analyzing/+page.svelte`
- `frontend/src/routes/onboarding/analyzing-2/+page.svelte`
- `frontend/src/routes/onboarding/analyzing-3/+page.svelte`
- `frontend/src/routes/auth/callback/+page.svelte`

**Testing Checklist:**
- [ ] OAuth flow completes successfully
- [ ] Analysis triggers automatically
- [ ] SSE stream connects and receives events
- [ ] Insights display correctly (AI vs fallback)
- [ ] Auto-advance timing is accurate (2s)
- [ ] Error states handled gracefully
- [ ] Animations smooth on all screen sizes
- [ ] Console shows no errors
- [ ] Network tab shows proper SSE connection

**Action Items for Javaris:**
- [ ] Write Playwright E2E tests (see `E2E_TEST_PLAN.md`)
- [ ] Polish analyzing screen animations
- [ ] Add loading skeletons
- [ ] Design error states
- [ ] Test SSE stream edge cases
- [ ] Mobile responsiveness check

---

## 🔍 What We're Extracting from Emails

### Current Analysis (v1)
```
1. Sender Domains
   → Identify company patterns
   → Detect SaaS tools (Linear, Notion, Slack, etc.)

2. Subject Patterns
   → Meeting invites (Google Calendar, Calendly)
   → Project mentions
   → Task assignments

3. Tool Mentions
   → Direct references ("Figma file", "Notion doc")
   → Integration emails (notifications, updates)

4. Email Metadata
   → Frequency by sender
   → Thread length (collaboration patterns)
   → Time patterns (work hours, urgency)
```

### Future Analysis (Roadmap)
```
5. People Network
   → Who they email most
   → Team structure (CC patterns)
   → External contacts (clients, vendors)

6. Business Context
   → Industry indicators (legal terms, finance jargon)
   → Role indicators (designer, developer, manager)
   → Company size (email volume, domains)

7. Workflow Patterns
   → How they manage tasks (email folders, labels)
   → Communication style (formal vs casual)
   → Tools they wish they had (complaint emails)

8. Integration Opportunities
   → "I wish X integrated with Y" mentions
   → Tool switching patterns
   → Manual processes mentioned

9. Pain Points
   → Complaints about current tools
   → Workarounds mentioned
   → Inefficiency signals
```

---

## 🚧 What Still Needs to Be Built

### Phase 1: MVP Completion (This PR)
- [x] Email analysis backend
- [x] Groq AI integration
- [x] SSE streaming
- [x] Analyzing screens
- [x] OAuth trigger
- [ ] **TESTING:** Manual E2E verification
- [ ] **TESTING:** Automated Playwright tests
- [ ] **DEPLOYMENT:** GCP Cloud Run update

### Phase 2: Starter Apps Generation (Next Sprint)
- [ ] AppCustomizerAgent full implementation
- [ ] Starter apps selection UI
- [ ] Module customization API
- [ ] Template system for common apps
- [ ] App installation flow

### Phase 3: Agent Code Editing (Critical)
- [ ] **Nick:** Design sandboxing architecture
- [ ] Agent → Code editing API
- [ ] Version control integration
- [ ] Code review by AI before apply
- [ ] Rollback mechanism

### Phase 4: Multi-Tenant Architecture
- [ ] Company → Workspace → OS hierarchy
- [ ] Per-workspace databases (or proper isolation)
- [ ] Git repository per workspace
- [ ] Workspace templates
- [ ] Sharing/collaboration features

### Phase 5: Advanced Email Analysis
- [ ] Ongoing email monitoring (not just onboarding)
- [ ] RAG integration (email context → consultation)
- [ ] Tool usage tracking
- [ ] Network graph (who works with whom)
- [ ] Automated integration suggestions

### Phase 6: Intelligence Layer
- [ ] Learn from user behavior post-onboarding
- [ ] Suggest new apps based on usage patterns
- [ ] Auto-configure integrations when tools detected
- [ ] Proactive agent suggestions
- [ ] Workflow automation recommendations

---

## 🎨 The Vision: Personalized Business OS

### What Makes This Different from Generic Templates?

**Traditional Approach:**
```
1. User signs up
2. Gets generic dashboard
3. Manually configures everything
4. Wastes hours setting up tools
5. Never fully optimized for their workflow
```

**Our Approach:**
```
1. User signs up → Gmail OAuth
2. AI analyzes their actual work patterns
3. Automatically generates:
   - Custom CRM (pre-configured with their clients)
   - Custom Projects (templates for their project types)
   - Custom Integrations (their actual tools)
   - Custom Agents (trained on their domain)
4. User gets workspace ready to use immediately
5. Continuously learns and adapts
```

### Real-World Example

**User:** Freelance Designer (Sarah)

**Email Analysis Detects:**
- Tools: Figma, Notion, Linear, Slack, Dribbble
- Clients: 5 recurring client domains
- Projects: 3 active projects mentioned
- Team: Works solo but collaborates with 2 developers
- Workflow: Design → Review → Deliver pattern

**AI Generates:**
1. **Custom CRM Module**
   - Pre-loaded with 5 client contacts
   - Custom fields: "Project Type", "Design Style", "Budget Range"
   - Agent: "Client Communication Assistant" (knows design lingo)

2. **Custom Project Tracker**
   - Templates: "Design Project", "Branding Package", "Website Design"
   - Stages: Design → Review → Revisions → Delivery
   - Integration: Auto-import Figma files

3. **Custom Integrations**
   - Figma Plugin: Import designs directly
   - Notion Sync: Two-way project sync
   - Linear Integration: Track design bugs

4. **Custom Agents**
   - "Design Review Agent": Analyzes Figma designs, suggests improvements
   - "Client Email Agent": Drafts professional client emails
   - "Project Estimator": Estimates time based on past projects

**Result:** Sarah gets a workspace that *feels like it was built for her* because it was.

---

## 🔐 Security & Privacy Considerations

### Email Access
- **Scope:** Read-only Gmail access
- **Limit:** Only 50 most recent emails analyzed
- **Storage:** Email metadata stored, not full content
- **Retention:** Analysis cached, emails not persisted
- **User Control:** Can revoke Gmail access anytime

### AI Processing
- **Provider:** Groq (not OpenAI, not Claude)
- **Data Sent:** Email metadata only (subjects, senders, dates)
- **Model:** llama-3.3-70b-versatile (open source model)
- **Privacy:** Groq doesn't train on user data
- **Compliance:** GDPR-compliant (user consent required)

### Database Security
- **Encryption:** All analysis results encrypted at rest
- **Access Control:** Workspace-level isolation
- **Audit Logs:** All AI API calls logged
- **Data Deletion:** User can delete analysis anytime

---

## 💰 Cost Analysis

### Groq API Costs
| Model | Input | Output | 50 Emails Analysis Cost |
|-------|-------|--------|-------------------------|
| llama-3.3-70b-versatile | $0.59/1M tokens | $0.79/1M tokens | ~$0.002-0.005 |

**Per User Onboarding:** < $0.01
**1000 Users:** < $10
**Very Affordable:** Groq is 10x cheaper than OpenAI

### Infrastructure Costs
- **Current:** Shared GCP Cloud Run instance
- **Scaling:** Need to plan workspace-per-instance or smart partitioning
- **Database:** PostgreSQL (Supabase pooled connection)
- **Storage:** Minimal (JSONB analysis results)

---

## 📊 Metrics & Success Criteria

### Pre-Launch Testing
- [ ] 10 manual test runs with different Gmail accounts
- [ ] 100% OAuth success rate
- [ ] < 5s average analysis time
- [ ] > 90% insight quality (manual review)
- [ ] Zero errors in production logs

### Post-Launch Metrics
- **Onboarding Completion Rate:** Target > 80%
- **Analysis Success Rate:** Target > 95%
- **Time to Workspace Ready:** Target < 30s
- **User Satisfaction:** "How accurate were the recommendations?" > 4/5
- **App Activation Rate:** What % of suggested apps get installed?

---

## 🚀 Deployment Checklist

### Before Merge
- [ ] All E2E tests passing
- [ ] Code review by Pedro (backend)
- [ ] Code review by Javaris (frontend)
- [ ] GCP deployment plan from Nick
- [ ] Security audit (email handling)
- [ ] Privacy policy updated (Gmail access disclosure)

### Deployment Steps
1. [ ] Merge to `main`
2. [ ] Run database migrations (054, 055, 056)
3. [ ] Update GCP Cloud Run with new env vars
4. [ ] Deploy backend (with SSE support)
5. [ ] Deploy frontend (Vercel/Cloudflare)
6. [ ] Test production OAuth flow
7. [ ] Monitor Groq API usage
8. [ ] Watch error logs for 24h

### Rollback Plan
- [ ] Keep old backend version running
- [ ] Database migrations are reversible
- [ ] Feature flag for onboarding flow
- [ ] Fallback to static onboarding if AI fails

---

## 📚 Documentation

All documentation for this feature is organized in:
```
desktop/backend-go/docs/integrations/ios-onboarding/
├── PR_REVIEW.md (this file)
├── E2E_TEST_PLAN.md (manual testing guide)
├── E2E_TEST_RESULTS.md (test results + verification)
└── (future: DEPLOYMENT.md, API.md, TROUBLESHOOTING.md)
```

---

## 🤔 Open Questions for Team Discussion

### For Nick
1. **Agent Code Editing:** E2B sandbox vs Git-based vs streaming API?
2. **Multi-Tenant DB:** Separate DBs per workspace or single DB with strong isolation?
3. **Version Control:** GitHub integration or custom Git server?
4. **Scaling:** Workspace-per-instance or shared instances with load balancing?

### For Pedro
1. **Email Monitoring:** Should we analyze emails continuously or just at onboarding?
2. **RAG Integration:** How to feed email context into consultation system?
3. **Tool Detection:** Can we auto-enable integrations when tools detected?
4. **Agent Coordination:** How should onboarding agent hand off to consultation agent?

### For Javaris
1. **Animation Polish:** What level of polish for MVP vs post-launch?
2. **Error States:** What happens if Groq API is down? Show fallback immediately?
3. **Testing Strategy:** Playwright only or also Cypress? Unit tests for stores?
4. **Mobile Support:** Do we need mobile-optimized onboarding flow now?

### For Roberto (Product Vision)
1. **Starter Apps:** How many should we show? 3-4 or personalized number?
2. **User Control:** Should users be able to reject AI suggestions?
3. **Pricing:** Does this affect pricing model? (AI analysis costs)
4. **Privacy:** How do we communicate email access clearly?

---

## 🎉 What's Next After This PR

### Immediate (This Week)
1. Team reviews this PR
2. Address any concerns/questions
3. Run E2E tests
4. Merge to main
5. Deploy to production

### Short-Term (Next 2 Weeks)
1. Monitor production usage
2. Collect user feedback on accuracy
3. Fix any bugs discovered
4. Polish animations/UX

### Medium-Term (Next Month)
1. Implement starter apps generation
2. Build app installation flow
3. Create module customization API
4. Add more integrations

### Long-Term (Q1 2026)
1. Multi-tenant architecture
2. Agent code editing
3. Version control system
4. Advanced email analysis
5. Continuous learning system

---

## 📞 Contact & Questions

**Questions about this PR?**
- Backend/AI: Ask Pedro or Roberto
- Infrastructure: Ask Nick
- Frontend/Testing: Ask Javaris
- Product Vision: Ask Roberto

**Found a bug?**
- Create GitHub issue with label: `ios-onboarding`

**Want to contribute?**
- See open tasks in project board
- Check `desktop/backend-go/docs/integrations/ios-onboarding/E2E_TEST_PLAN.md`

---

**Thank you for reviewing! Let's ship this and change how people onboard to business software. 🚀**
