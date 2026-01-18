# BusinessOS Onboarding System

**Complete Technical Documentation for Team**

Version: 1.0
Last Updated: January 2026
Status: Production-Ready

---

## Table of Contents

1. [System Overview](#system-overview)
2. [User Journey Flow](#user-journey-flow)
3. [OAuth Google Integration](#oauth-google-integration)
4. [Real Data Analysis Pipeline](#real-data-analysis-pipeline)
5. [Frontend Implementation](#frontend-implementation)
6. [Backend Architecture](#backend-architecture)
7. [Database Schema](#database-schema)
8. [Key Implementation Files](#key-implementation-files)
9. [Configuration & Deployment](#configuration--deployment)
10. [Testing & Debugging](#testing--debugging)

---

## System Overview

The BusinessOS onboarding system is an AI-powered, personalized onboarding flow that:
- Authenticates users via Google OAuth
- Analyzes Gmail data to understand user workflow
- Generates AI-powered insights about the user's work style
- Recommends personalized "starter apps" tailored to their needs
- Guides users through a smooth 11-screen journey from signin to ready state

### Key Technologies

**Frontend:**
- SvelteKit (Svelte 5 with runes)
- TypeScript
- Svelte Stores for state management
- Server-Sent Events (SSE) for real-time updates
- Custom OSA UI components (PillButton, GlassCard)

**Backend:**
- Go 1.24.1
- Gin HTTP framework
- PostgreSQL with pgx/v5
- Gmail API integration
- Groq AI (Llama 3.3 70B) for profile analysis
- SSE streaming for real-time progress

**Architecture:**
- Handler → Service → Repository pattern
- Background goroutines for long-running analysis
- Real-time polling for analysis status
- Encrypted OAuth token storage

---

## User Journey Flow

### Complete 11-Screen Journey

```
1. /onboarding             → Welcome screen ("A new era of business software")
2. /onboarding/meet-osa    → Meet OSA introduction
3. /onboarding/signin      → Google OAuth sign-in
4. /onboarding/gmail       → Gmail connection (full access scope)
5. /onboarding/username    → Claim unique username
6. /onboarding/analyzing   → AI analysis screen 1 (shows insight #1)
7. /onboarding/analyzing-2 → AI analysis screen 2 (shows insight #2)
8. /onboarding/analyzing-3 → AI analysis screen 3 (shows insight #3)
9. /onboarding/starter-apps → View personalized apps (4 AI-generated apps)
10. /onboarding/ready      → "You're all set!" celebration
11. → Redirect to /dashboard (main app)
```

### Flow Diagram

```
┌──────────────┐
│   Welcome    │
│  (Screen 1)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Meet OSA    │
│  (Screen 2)  │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────┐
│        Google OAuth Signin           │
│  ┌────────────────────────────────┐  │
│  │  1. Initiate OAuth flow        │  │
│  │  2. User grants Gmail access   │  │
│  │  3. Callback with auth code    │  │
│  │  4. Exchange for tokens        │  │
│  │  5. Store tokens (encrypted)   │  │
│  │  6. Trigger BACKGROUND analysis│ ◄── CRITICAL: Non-blocking
│  └────────────────────────────────┘  │
└───────────────┬──────────────────────┘
                │
                ▼
┌───────────────────────────────────────┐
│       Gmail Connection Check          │
│  (Frontend confirms Gmail connected)  │
└───────────────┬───────────────────────┘
                │
                ▼
┌───────────────────────────────────────┐
│         Username Selection            │
│  - Validate: alphanumeric + underscore│
│  - Check availability via API         │
│  - Save to DB                         │
└───────────────┬───────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│                  AI Analysis (3 screens)                     │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  BACKGROUND PROCESS (started at OAuth callback):        │ │
│  │  1. Fetch 100 recent emails from Gmail                 │ │
│  │  2. Extract metadata (senders, keywords, tools)        │ │
│  │  3. AI analysis via Groq (Llama 3.3 70B)              │ │
│  │  4. Generate 3 insights + interests + tools            │ │
│  │  5. Save to onboarding_user_analysis table             │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  FRONTEND (polling for status):                              │
│  - Screen 1: Shows insight #1 when ready                     │
│  - Screen 2: Shows insight #2 when ready                     │
│  - Screen 3: Shows insight #3 when ready                     │
│  - Auto-advance every 2 seconds                              │
└────────────────┬──────────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────┐
│       Starter Apps Generation          │
│  - Use AI to recommend 3-4 apps        │
│  - Based on insights/interests/tools   │
│  - Save to onboarding_starter_apps     │
└────────────────┬───────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────┐
│         Ready Screen                   │
│  "You're all set!" celebration         │
└────────────────┬───────────────────────┘
                 │
                 ▼
          /dashboard
```

### Screen-by-Screen Details

#### Screen 1: Welcome (`/onboarding`)
**Purpose:** First impression, brand introduction
**Key Elements:**
- Cloud logo with float animation
- Headline: "A new era of business software is here."
- CTA: "Get Started" button
- Helper text: "Takes less than 2 minutes"

**Implementation:**
- Static Svelte page
- No API calls
- Simple navigation to next screen

---

#### Screen 2: Meet OSA (`/onboarding/meet-osa`)
**Purpose:** Introduce OSA (Operating System Agent)
**Key Elements:**
- OSA introduction content
- Explanation of personalized business OS
- CTA: Continue to sign-in

**Implementation:**
- Static content
- Sets user expectations

---

#### Screen 3: Sign In (`/onboarding/signin`)
**Purpose:** Google OAuth authentication
**Key Elements:**
- "Sign in with Google" button
- Privacy/permissions notice

**Implementation:**
- Initiates OAuth flow via `/api/v1/auth/google/login`
- Redirects to Google consent screen
- Returns to callback handler

---

#### Screen 4: Gmail Connection (`/onboarding/gmail`)
**Purpose:** Confirm Gmail access and explain why
**Key Elements:**
- Checkmark confirmation
- Explanation: "We'll analyze your emails to personalize your experience"
- Privacy note: "Your data is encrypted and never shared"

**Implementation:**
- Checks if Gmail tokens exist in DB
- Shows connected state
- Explains next steps

---

#### Screen 5: Username (`/onboarding/username`)
**Purpose:** Claim unique username
**Key Elements:**
- Text input with real-time validation
- Availability check (debounced 500ms)
- Visual feedback (checkmark/X icon)
- Validation rules: min 3 chars, alphanumeric + underscore

**API Calls:**
- `GET /api/v1/users/username/check/:username` - Check availability
- `PATCH /api/v1/users/username` - Set username

**Validation:**
- Regex: `/^[a-zA-Z0-9_]+$/`
- Min length: 3 characters
- Unique across all users

---

#### Screens 6-8: Analyzing (`/onboarding/analyzing`, `/analyzing-2`, `/analyzing-3`)
**Purpose:** Show AI-generated insights while analysis runs
**Key Elements:**
- Loading spinner (only on screen 1)
- Insight message (one per screen)
- "AI-Generated" badge
- Auto-advance every 2 seconds

**Data Flow:**
1. **OAuth callback triggers background analysis** (non-blocking)
2. Frontend **polls** for status via `GET /api/osa-onboarding/user-analysis/:user_id`
3. When status becomes `completed`, insights are available:
   - `insights[0]` → Screen 1
   - `insights[1]` → Screen 2
   - `insights[2]` → Screen 3
4. Auto-advance to next screen after 2 seconds

**Example Insights:**
- "No-code builder energy, big time"
- "Design tools are your playground"
- "AI-curious, testing new platforms"

**Fallback Behavior:**
- If analysis fails or times out, show generic insights
- Still auto-advance (graceful degradation)

---

#### Screen 9: Starter Apps (`/onboarding/starter-apps`)
**Purpose:** Display 3-4 personalized app recommendations
**Key Elements:**
- Grid of app cards (4 apps typically)
- Each card shows:
  - Icon emoji
  - Title
  - Description
  - "Why this helps you" reasoning

**Data Flow:**
1. Analysis completes
2. Backend calls AppCustomizerAgent to generate apps
3. Apps saved to `onboarding_starter_apps` table
4. Frontend fetches and displays

**Example Apps:**
- "Design Tracker" (because user uses Figma)
- "Client CRM" (because user has client emails)
- "Task Board" (because user mentions deadlines)

---

#### Screen 10: Ready (`/onboarding/ready`)
**Purpose:** Celebration, completion
**Key Elements:**
- Celebration message: "You're all set!"
- Final CTA: "Enter BusinessOS"
- Marks `onboarding_completed = true` in DB

**Implementation:**
- Updates user record
- Clears onboarding localStorage
- Redirects to `/dashboard`

---

## OAuth Google Integration

### OAuth Flow Architecture

```
┌─────────────┐
│   Browser   │
│  (Frontend) │
└──────┬──────┘
       │
       │ 1. Click "Sign in with Google"
       ▼
┌──────────────────────────────────────┐
│  Backend: /api/v1/auth/google/login  │
│  ┌────────────────────────────────┐  │
│  │ 1. Generate random state       │  │
│  │ 2. Store in cookie             │  │
│  │ 3. Build Google OAuth URL      │  │
│  │ 4. Redirect to Google          │  │
│  └────────────────────────────────┘  │
└──────────────┬───────────────────────┘
               │
               ▼
┌───────────────────────────────────────┐
│       Google OAuth Consent Screen     │
│  - User grants permissions            │
│  - Scopes requested:                  │
│    * userinfo.email                   │
│    * userinfo.profile                 │
│    * https://mail.google.com/         │  ◄── FULL GMAIL ACCESS
└───────────────┬───────────────────────┘
                │
                │ 2. Redirect with auth code
                ▼
┌──────────────────────────────────────────────┐
│  Backend: /api/v1/auth/google/callback       │
│  ┌────────────────────────────────────────┐  │
│  │ 1. Verify state parameter (CSRF)      │  │
│  │ 2. Exchange code for tokens           │  │
│  │ 3. Fetch user info from Google        │  │
│  │ 4. Upsert user in DB                  │  │
│  │ 5. Store OAuth tokens (encrypted)     │  │
│  │ 6. Trigger background Gmail analysis  │ ◄── CRITICAL
│  │ 7. Create session                     │  │
│  │ 8. Set session cookie                 │  │
│  │ 9. Set new_user cookie (if new)       │  │
│  │ 10. Redirect to /onboarding/gmail     │  │
│  └────────────────────────────────────────┘  │
└───────────────────────────────────────────────┘
```

### Gmail Scopes Required

**Scope:** `https://mail.google.com/`

**Permissions Granted:**
- Read all emails (inbox, sent, drafts, etc.)
- Send emails on behalf of user
- Modify and delete emails
- Access labels and filters

**Why Full Access?**
- Needed to analyze user's email patterns
- Extract metadata (senders, keywords, tools mentioned)
- Build personalized profile
- Future feature: AI email assistant

**Security Considerations:**
- Tokens stored encrypted in `user_integrations` table
- `access_token_encrypted` and `refresh_token_encrypted` columns
- TODO: Implement AES encryption with `TOKEN_ENCRYPTION_KEY` env var
- Currently stored as plain bytea (needs encryption before production)

### Backend OAuth Handler

**File:** `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/handlers/auth_google.go`

**Key Functions:**

#### `InitiateGoogleLogin`
```go
func (h *GoogleAuthHandler) InitiateGoogleLogin(c *gin.Context) {
    // 1. Generate CSRF state token
    state := generateRandomState()

    // 2. Store state in cookie (10 min expiry)
    c.SetCookie("oauth_state", state, 600, "/", "", false, true)

    // 3. Get redirect destination (default: /dashboard)
    redirectAfter := c.Query("redirect")
    if redirectAfter == "" {
        redirectAfter = "/dashboard"
    }
    c.SetCookie("oauth_redirect", redirectAfter, 600, "/", "", false, true)

    // 4. Build Google OAuth URL
    authURL := h.oauthConfig.AuthCodeURL(state,
        oauth2.AccessTypeOffline,
        oauth2.SetAuthURLParam("prompt", "select_account"))

    // 5. Redirect to Google
    c.Redirect(http.StatusTemporaryRedirect, authURL)
}
```

#### `HandleGoogleLoginCallback`
```go
func (h *GoogleAuthHandler) HandleGoogleLoginCallback(c *gin.Context) {
    // 1. Verify state (CSRF protection)
    state := c.Query("state")
    storedState, _ := c.Cookie("oauth_state")
    if state != storedState {
        return error
    }

    // 2. Exchange authorization code for tokens
    code := c.Query("code")
    token, err := h.oauthConfig.Exchange(ctx, code)

    // 3. Fetch user info from Google
    userInfo, err := h.getGoogleUserInfo(token.AccessToken)

    // 4. Create or update user in DB
    userID, isNewUser, err := h.upsertUser(ctx, userInfo)

    // 5. Store Gmail tokens + trigger background analysis
    err := h.storeGmailTokensAndStartAnalysis(ctx, userID, token)

    // 6. Create session
    sessionToken, err := h.createSession(ctx, userID)

    // 7. Set session cookie
    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "better-auth.session_token",
        Value:    sessionToken,
        MaxAge:   60 * 60 * 24 * 30, // 30 days
        HttpOnly: true,
        Secure:   isProduction,
        SameSite: sameSite,
    })

    // 8. Set new_user cookie for frontend routing
    if isNewUser {
        c.SetCookie("new_user", "true", 60, "/", "", false, true)
    }

    // 9. Redirect to app
    c.Redirect(http.StatusTemporaryRedirect, redirectAfter)
}
```

#### `storeGmailTokensAndStartAnalysis`
```go
func (h *GoogleAuthHandler) storeGmailTokensAndStartAnalysis(
    ctx context.Context,
    userID string,
    token *oauth2.Token,
) error {
    // 1. Store tokens in user_integrations table
    _, err := h.pool.Exec(ctx, `
        INSERT INTO user_integrations (
            user_id, provider_id, status,
            access_token_encrypted, refresh_token_encrypted,
            token_expires_at, scopes, ...
        ) VALUES ($1, 'google_gmail', 'connected', $2, $3, $4, $5, ...)
        ON CONFLICT (user_id, provider_id) DO UPDATE SET ...
    `, userID, token.AccessToken, token.RefreshToken, token.Expiry, scopes)

    // 2. Trigger background analysis (non-blocking goroutine)
    go func() {
        // Use background context (not request context)
        analysisCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
        defer cancel()

        // Run Gmail analysis
        if err := h.runGmailAnalysis(analysisCtx, userID); err != nil {
            log.Printf("Gmail analysis failed: %v", err)
            // Store error in analysis table
        }
    }()

    return nil
}
```

**CRITICAL DESIGN DECISION:**
- Analysis runs in **background goroutine**
- Does NOT block OAuth callback response
- Frontend redirects immediately to `/onboarding/gmail`
- Analysis continues asynchronously
- Frontend **polls** for completion

---

## Real Data Analysis Pipeline

### Analysis Flow

```
┌─────────────────────────────────────────┐
│  OAuth Callback Completes               │
│  (user redirected to /onboarding/gmail) │
└───────────────┬─────────────────────────┘
                │
                │ BACKGROUND GOROUTINE STARTS
                ▼
┌──────────────────────────────────────────────────────────┐
│  Step 1: Email Metadata Extraction                       │
│  Service: EmailAnalyzerService                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 1. Call Gmail API: list 100 recent messages       │  │
│  │ 2. For each message, extract:                     │  │
│  │    - Sender email → domain                        │  │
│  │    - Subject → keywords                           │  │
│  │    - Body snippet → keywords                      │  │
│  │    - Date                                         │  │
│  │ 3. Pattern detection:                             │  │
│  │    - Tools mentioned (Figma, Notion, GitHub, etc) │  │
│  │    - Topics (design, development, marketing)      │  │
│  │    - Sender domains (frequency analysis)          │  │
│  │ 4. Output: EmailAnalysisMetadata                  │  │
│  └────────────────────────────────────────────────────┘  │
└───────────────┬──────────────────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────────────────┐
│  Step 2: AI Profile Analysis                             │
│  Service: ProfileAnalyzerAgent                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 1. Build prompt with metadata                      │  │
│  │ 2. Call Groq API (Llama 3.3 70B)                  │  │
│  │ 3. Request JSON response with:                     │  │
│  │    - insights: [3 conversational phrases]         │  │
│  │    - interests: [detected interests]              │  │
│  │    - tools_used: [top tools]                      │  │
│  │    - profile_summary: "narrative summary"         │  │
│  │ 4. Parse and validate response                     │  │
│  │ 5. Output: ProfileAnalysisResult                   │  │
│  └────────────────────────────────────────────────────┘  │
└───────────────┬──────────────────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────────────────┐
│  Step 3: Save to Database                                │
│  Table: onboarding_user_analysis                         │
│  ┌────────────────────────────────────────────────────┐  │
│  │ INSERT INTO onboarding_user_analysis (              │  │
│  │   user_id, workspace_id,                           │  │
│  │   insights,           -- JSONB array               │  │
│  │   interests,          -- JSONB array               │  │
│  │   tools_used,         -- JSONB array               │  │
│  │   profile_summary,    -- TEXT                      │  │
│  │   total_emails_analyzed,                           │  │
│  │   sender_domains,                                  │  │
│  │   analysis_model,     -- "llama-3.3-70b-versatile" │  │
│  │   ai_provider,        -- "groq"                    │  │
│  │   status,             -- "completed"               │  │
│  │   completed_at                                     │  │
│  │ )                                                  │  │
│  └────────────────────────────────────────────────────┘  │
└───────────────┬──────────────────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────┐
│  Frontend Polling Detects Completion     │
│  (status: "analyzing" → "completed")     │
│  → Auto-advance to next screen           │
└──────────────────────────────────────────┘
```

### Service 1: Email Analyzer

**File:** `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/services/onboarding_email_analyzer.go`

**Purpose:** Extract structured metadata from Gmail messages

**Key Data Structures:**
```go
type EmailAnalysisMetadata struct {
    TotalEmails     int                    // Number of emails analyzed
    SenderDomains   map[string]int         // domain → count
    SubjectKeywords []string               // Top 20 keywords from subjects
    BodyKeywords    []string               // Top 20 keywords from bodies
    DetectedTools   map[string]int         // tool → count
    TopicFrequency  map[string]int         // topic → count
    EmailDates      []time.Time            // For activity patterns
}
```

**Tool Detection:**
Searches for mentions of 50+ known tools:
- **Design:** Figma, Sketch, Adobe XD, Canva, Photoshop
- **Development:** GitHub, GitLab, VS Code, Docker, Kubernetes
- **Project Management:** Notion, Asana, Trello, Jira, Linear
- **Communication:** Slack, Discord, Teams, Zoom
- **No-code:** Webflow, Bubble, Zapier, Retool
- **CRM:** Salesforce, HubSpot, Pipedrive
- **Other:** Stripe, Shopify, WordPress

**Topic Patterns (Regex):**
```go
var topicPatterns = map[string]*regexp.Regexp{
    "design":        regexp.MustCompile(`(?i)\b(design|ui|ux|prototype|mockup)\b`),
    "development":   regexp.MustCompile(`(?i)\b(code|develop|build|api|backend)\b`),
    "marketing":     regexp.MustCompile(`(?i)\b(market|campaign|seo|content)\b`),
    "sales":         regexp.MustCompile(`(?i)\b(sale|deal|prospect|lead|client)\b`),
    "product":       regexp.MustCompile(`(?i)\b(product|feature|roadmap|launch)\b`),
    "analytics":     regexp.MustCompile(`(?i)\b(analytic|metric|data|report)\b`),
    "automation":    regexp.MustCompile(`(?i)\b(automat|workflow|integration)\b`),
    "collaboration": regexp.MustCompile(`(?i)\b(collaborat|team|meeting)\b`),
}
```

**Process:**
1. Call Gmail API to sync 100 recent emails
2. For each email:
   - Extract domain from sender email
   - Tokenize subject and body into keywords
   - Search for tool mentions (case-insensitive)
   - Match topic patterns via regex
3. Aggregate into frequency maps
4. Return structured metadata

---

### Service 2: Profile Analyzer (AI Agent)

**File:** `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/services/onboarding_profile_analyzer.go`

**Purpose:** Use AI to analyze email metadata and generate conversational insights

**LLM Configuration:**
- **Provider:** Groq
- **Model:** `llama-3.3-70b-versatile`
- **Temperature:** Default (balanced creativity)
- **Output:** Structured JSON

**System Prompt:**
```
You are a user profiling expert analyzing email patterns to understand
someone's work style, interests, and tool usage.

Your task is to:
1. Generate 3 short, conversational insight phrases
   (like "No-code builder energy ✨", "Design tools are your playground")
2. Identify core interests (max 5)
3. List top tools/platforms used (max 5)
4. Write a concise profile summary (2-3 sentences)

Be specific, personal, and encouraging. Use natural language that
feels like a friend describing them.

IMPORTANT: Respond ONLY with valid JSON matching this schema:
{
  "insights": ["phrase 1", "phrase 2", "phrase 3"],
  "interests": ["interest1", "interest2", ...],
  "tools_used": ["tool1", "tool2", ...],
  "profile_summary": "A concise summary...",
  "work_patterns": {},
  "confidence": 0.85
}
```

**User Prompt Construction:**
```go
func buildAnalysisPrompt(metadata *EmailMetadataInput) string {
    prompt := fmt.Sprintf("Analyze this user's email activity from %d recent emails:\n\n",
        metadata.TotalEmails)

    // Top sender domains
    prompt += "**Top Email Senders:**\n"
    for domain, count := range metadata.SenderDomains {
        prompt += fmt.Sprintf("- %s: %d emails\n", domain, count)
    }

    // Detected tools
    prompt += "**Tools/Platforms Mentioned:**\n"
    for tool, count := range metadata.DetectedTools {
        prompt += fmt.Sprintf("- %s: mentioned %d times\n", tool, count)
    }

    // Topics
    prompt += "**Discussion Topics:**\n"
    for topic, count := range metadata.TopicFrequency {
        prompt += fmt.Sprintf("- %s: %d occurrences\n", topic, count)
    }

    // Sample keywords
    prompt += "**Subject Keywords (sample):**\n"
    prompt += strings.Join(metadata.SubjectKeywords[:10], ", ")

    prompt += "\n\nBased on this data, create a personalized profile analysis."
    return prompt
}
```

**Output Validation:**
- Ensures exactly 3 insights (truncate or pad)
- Validates required fields (interests, summary)
- Cleans JSON (removes markdown code blocks)

**Example Output:**
```json
{
  "insights": [
    "No-code builder energy, big time",
    "Design tools are your playground",
    "AI-curious, testing new platforms"
  ],
  "interests": [
    "no-code development",
    "UI/UX design",
    "AI automation",
    "product development"
  ],
  "tools_used": [
    "Figma",
    "Notion",
    "Zapier",
    "GitHub"
  ],
  "profile_summary": "You're a no-code enthusiast with a strong design background. You love experimenting with AI tools and building automations to streamline workflows.",
  "confidence": 0.92
}
```

---

### Service 3: App Customizer (AI Agent)

**File:** `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/services/onboarding_app_customizer.go`

**Purpose:** Generate 3-4 personalized "starter app" recommendations

**Core Modules Available:**
- CRM (Client relationship management)
- Tasks (Task and todo management)
- Projects (Project tracking)
- Calendar (Scheduling)
- Notes (Documentation)
- Dashboard (Analytics)
- Team (Collaboration)
- Knowledge (Knowledge base)

**System Prompt:**
```
You are an app recommendation specialist for BusinessOS,
a personalized business operating system.

Based on the user's profile, recommend 3-4 starter apps that will
be most valuable to them.

For each app, you can either:
1. Customize an existing core module (CRM, Tasks, Projects, etc.)
2. Create a completely new app concept tailored to their needs

Guidelines:
- Be specific and personal - use their actual interests and tools
- Keep app titles short and memorable (2-4 words)
- Use relevant emojis for icons
- Explain WHY this app helps them specifically
- Category: tracker, companion, feedback, daily, workflow, automation
- Priority 1 = most important

Respond ONLY with valid JSON matching this schema:
{
  "apps": [
    {
      "title": "Design Tracker",
      "description": "Track Figma projects and design feedback",
      "icon_emoji": "🎨",
      "category": "tracker",
      "reasoning": "You use Figma daily and need a place to organize design work",
      "customization_prompt": "Full prompt for AI to build this app...",
      "based_on_interests": ["design", "collaboration"],
      "based_on_tools": ["Figma"],
      "base_module": "Projects",
      "module_customizations": {"theme": "design-focused"},
      "priority": 1
    }
  ],
  "confidence": 0.9
}
```

**Example Recommendations:**
```json
{
  "apps": [
    {
      "title": "Design Tracker",
      "description": "Organize Figma projects, track feedback, and manage design sprints",
      "icon_emoji": "🎨",
      "category": "tracker",
      "reasoning": "You use Figma extensively and collaborate on design projects",
      "based_on_interests": ["UI/UX design", "collaboration"],
      "based_on_tools": ["Figma", "Notion"],
      "base_module": "Projects",
      "priority": 1
    },
    {
      "title": "Client CRM",
      "description": "Manage client relationships, track conversations, and follow-ups",
      "icon_emoji": "👥",
      "category": "companion",
      "reasoning": "Your emails show frequent client communication",
      "based_on_interests": ["client management"],
      "based_on_tools": [],
      "base_module": "CRM",
      "priority": 2
    },
    {
      "title": "Automation Hub",
      "description": "Connect Zapier workflows and track automation performance",
      "icon_emoji": "⚡",
      "category": "automation",
      "reasoning": "You mention Zapier and automation frequently",
      "based_on_interests": ["AI automation", "workflow optimization"],
      "based_on_tools": ["Zapier"],
      "base_module": null,
      "priority": 3
    }
  ]
}
```

**Database Storage:**
Apps saved to `onboarding_starter_apps` table with:
- User ID + Workspace ID
- Analysis ID (foreign key)
- Display order (1-4)
- Status (pending/generating/ready/failed)
- Generation metadata (model, tokens, duration)

---

## Frontend Implementation

### State Management Architecture

```
┌─────────────────────────────────────────┐
│  Svelte Stores (Reactive State)         │
│  ┌───────────────────────────────────┐  │
│  │  onboardingStore.ts               │  │
│  │  - currentStep                    │  │
│  │  - completed                      │  │
│  │  - userData (username, gmail)     │  │
│  │  - analysis (3 messages)          │  │
│  │  - starterApps                    │  │
│  │  Persisted: localStorage          │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  onboardingAnalysis.ts            │  │
│  │  - analysisId                     │  │
│  │  - status (analyzing/completed)   │  │
│  │  - insights [3]                   │  │
│  │  - interests []                   │  │
│  │  - toolsUsed []                   │  │
│  │  - isStreaming                    │  │
│  │  - error                          │  │
│  │  Methods:                         │  │
│  │  - start(userId, workspaceId)     │  │
│  │  - pollByUserId(userId)           │  │
│  │  - cancel()                       │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Store 1: onboardingStore

**File:** `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/stores/onboardingStore.ts`

**State Shape:**
```typescript
interface OnboardingState {
    currentStep: number;           // 0-12
    totalSteps: number;            // 13
    completed: boolean;
    userData: {
        email?: string;
        username?: string;
        gmailConnected: boolean;
        interests?: string[];
        starterApps?: StarterApp[];
    };
    analysis: {
        message1?: string;         // Insight for screen 1
        message2?: string;         // Insight for screen 2
        message3?: string;         // Insight for screen 3
    };
}
```

**Methods:**
- `nextStep()` - Advance to next screen
- `prevStep()` - Go back one screen
- `goToStep(n)` - Jump to specific step
- `setUserData(data)` - Update user data
- `setAnalysis(analysis)` - Set AI insights
- `setStarterApps(apps)` - Set recommended apps
- `complete()` - Mark onboarding as done
- `reset()` - Clear all data

**Persistence:**
- Auto-saves to `localStorage` on every state change
- Key: `osa_onboarding_state`
- Loaded on mount

---

### Store 2: onboardingAnalysis

**File:** `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/stores/onboardingAnalysis.ts`

**State Shape:**
```typescript
interface OnboardingAnalysisState {
    analysisId: string | null;
    status: 'analyzing' | 'completed' | 'failed' | null;
    insights: string[];            // [message1, message2, message3]
    interests: string[];
    toolsUsed: string[];
    summary: string;
    isStreaming: boolean;
    isLoading: boolean;
    error: string | null;
    startedAt: number | null;
    completedAt: number | null;
}
```

**Methods:**

#### `start(userId, workspaceId, maxEmails)`
Initiates analysis via API (not currently used - OAuth triggers analysis)

#### `pollByUserId(userId)`
**CRITICAL METHOD:** Polls backend for analysis status

```typescript
async function pollByUserId(userId: string) {
    update((s) => ({ ...s, isLoading: true }));

    const maxAttempts = 60; // 2 minutes max
    let attempts = 0;

    const pollInterval = setInterval(async () => {
        attempts++;

        if (attempts > maxAttempts) {
            // Timeout - use fallback
            clearInterval(pollInterval);
            update((s) => ({
                ...s,
                error: 'Analysis timeout',
                status: 'failed'
            }));
            return;
        }

        try {
            const response = await fetch(
                `${backendUrl}/api/osa-onboarding/user-analysis/${userId}`
            );
            const data = await response.json();

            update((s) => ({
                ...s,
                status: data.status,
                insights: data.insights || [],
                toolsUsed: data.tools || [],
                isLoading: data.status === 'analyzing',
                completedAt: data.status === 'completed' ? Date.now() : null
            }));

            // Stop polling when complete/failed
            if (data.status === 'completed' || data.status === 'failed') {
                clearInterval(pollInterval);
            }
        } catch (err) {
            console.error('Polling error:', err);
        }
    }, 2000); // Poll every 2 seconds
}
```

**Derived Stores:**
```typescript
// Extract insights for each screen
export const analyzingInsights = derived(onboardingAnalysis, ($analysis) => ({
    message1: $analysis.insights[0] || 'No-code builder energy',
    message2: $analysis.insights[1] || 'Design tools are your playground',
    message3: $analysis.insights[2] || 'AI-curious, testing new platforms',
    hasRealData: $analysis.insights.length >= 3
}));

// Check if complete
export const analysisComplete = derived(
    onboardingAnalysis,
    ($analysis) => $analysis.status === 'completed'
);

// Calculate duration
export const analysisDuration = derived(onboardingAnalysis, ($analysis) => {
    if (!$analysis.startedAt) return 0;
    const endTime = $analysis.completedAt || Date.now();
    return endTime - $analysis.startedAt;
});
```

---

### Component System: OSA UI Library

**Location:** `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/components/osa/`

**Components:**
- `PillButton.svelte` - Primary CTA button with pill shape
- `GlassCard.svelte` - Frosted glass effect card
- `AppCard.svelte` - Starter app display card
- `LoadingSpinner.svelte` - Analysis loading animation

**PillButton Usage:**
```svelte
<PillButton
    variant="primary"   {/* primary | secondary | outline */}
    size="lg"           {/* sm | md | lg */}
    disabled={false}
    onclick={handleClick}
>
    Get Started
</PillButton>
```

**Styling:**
- Matches Wabi iOS design language
- Pill-shaped buttons (fully rounded corners)
- Smooth animations (fadeIn, float, spin)
- Consistent spacing and typography
- Professional, minimal aesthetic

---

### Screen Implementation Pattern

Each screen follows this pattern:

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { onboardingStore } from '$lib/stores/onboardingStore';

    // Local state
    let loading = $state(false);
    let error = $state('');

    // Handle next step
    function handleNext() {
        onboardingStore.nextStep();
        goto('/onboarding/next-screen');
    }

    // Handle back
    function handleBack() {
        onboardingStore.prevStep();
        goto('/onboarding/previous-screen');
    }
</script>

<svelte:head>
    <title>Screen Title - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
    <div class="screen">
        <div class="content">
            <!-- Screen content -->
        </div>
    </div>
</div>

<style>
    .onboarding-background {
        min-height: 100vh;
        background-image: url('/logos/integrations/MIOSABRANDBackround.png');
        background-size: cover;
    }
    /* ... */
</style>
```

**Key Features:**
- Consistent background image
- Centered content layout
- FadeIn animations
- Standard navigation pattern
- Svelte 5 runes (`$state`, `$effect`)

---

### Real-Time Polling Implementation

**Analyzing Screens Use This Pattern:**

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { onboardingAnalysis, analyzingInsights } from '$lib/stores/onboardingAnalysis';
    import { getSession } from '$lib/auth-client';

    let analyzing = true;
    let insightMessage = '';

    onMount(async () => {
        // Subscribe to insights
        const unsubscribe = analyzingInsights.subscribe(($insights) => {
            insightMessage = $insights.message1; // or message2, message3
        });

        // Subscribe to analysis state
        const unsubscribeAnalysis = onboardingAnalysis.subscribe(($analysis) => {
            analyzing = $analysis.isStreaming || $analysis.isLoading;

            // Auto-advance when complete
            if ($analysis.status === 'completed' || $analysis.status === 'failed') {
                analyzing = false;
                setTimeout(() => {
                    goto('/onboarding/next-screen');
                }, 2000);
            }
        });

        // Get user session and start polling
        const session = await getSession();
        if (session.data?.user?.id) {
            onboardingAnalysis.pollByUserId(session.data.user.id);
        } else {
            // Fallback to generic insights
            insightMessage = 'No-code builder energy';
            analyzing = false;
        }

        return () => {
            unsubscribe();
            unsubscribeAnalysis();
        };
    });
</script>

<div>
    {#if analyzing}
        <h1>Analyzing your workspace...</h1>
        <div class="spinner"></div>
    {:else}
        <h1>{insightMessage}</h1>
        {#if $analyzingInsights.hasRealData}
            <p class="ai-badge">✨ AI-Generated</p>
        {/if}
    {/if}
</div>
```

**Why Polling Instead of SSE?**
- Simpler implementation for MVP
- Works reliably across all network conditions
- No WebSocket/SSE connection management
- Easy fallback on timeout
- 2-second polling is fast enough for UX

**Future Enhancement:**
- Could replace with SSE streaming
- Endpoint exists: `GET /api/v1/osa-onboarding/analyze/:analysis_id/stream`
- Would provide real-time progress updates

---

## Backend Architecture

### Layer Architecture

```
┌─────────────────────────────────────────┐
│           HTTP Layer (Gin)              │
│  /api/v1/auth/google/login              │
│  /api/v1/auth/google/callback           │
│  /api/osa-onboarding/user-analysis/:id  │
│  /api/osa-onboarding/analyze            │
│  /api/osa-onboarding/generate-apps      │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         Handler Layer                   │
│  ┌───────────────────────────────────┐  │
│  │  GoogleAuthHandler                │  │
│  │  - InitiateGoogleLogin            │  │
│  │  - HandleGoogleLoginCallback      │  │
│  │  - storeGmailTokensAndStartAnalysis│ │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │  OSAOnboardingHandler             │  │
│  │  - GetUserAnalysisStatus          │  │
│  │  - StartAnalysis                  │  │
│  │  - GenerateStarterApps            │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         Service Layer                   │
│  ┌───────────────────────────────────┐  │
│  │  EmailAnalyzerService             │  │
│  │  - AnalyzeRecentEmails()          │  │
│  │  - extractMetadata()              │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │  ProfileAnalyzerAgent             │  │
│  │  - AnalyzeProfile()               │  │
│  │  - buildAnalysisPrompt()          │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │  AppCustomizerAgent               │  │
│  │  - RecommendApps()                │  │
│  │  - buildRecommendationPrompt()    │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │  GmailService (Integration)       │  │
│  │  - SyncEmails()                   │  │
│  │  - GetEmails()                    │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│       Database Layer (pgx/SQLC)         │
│  - user table                           │
│  - user_integrations                    │
│  - onboarding_user_analysis             │
│  - onboarding_starter_apps              │
│  - onboarding_email_metadata            │
│  - emails table (Gmail sync)            │
└─────────────────────────────────────────┘
```

### Handler Pattern

**Standard Handler Structure:**
```go
type OSAOnboardingHandler struct {
    pool            *pgxpool.Pool          // Database connection pool
    queries         *sqlc.Queries          // Type-safe queries (SQLC)
    cfg             *config.Config         // App configuration
    emailAnalyzer   *services.EmailAnalyzerService
    profileAnalyzer *services.ProfileAnalyzerAgent
    appCustomizer   *services.AppCustomizerAgent
    gmailService    *integrationGoogle.GmailService
}

func NewOSAOnboardingHandler(pool *pgxpool.Pool, cfg *config.Config, googleProvider *integrations.Provider) *OSAOnboardingHandler {
    gmailService := integrations.NewGmailService(googleProvider)

    return &OSAOnboardingHandler{
        pool:            pool,
        queries:         sqlc.New(pool),
        cfg:             cfg,
        emailAnalyzer:   services.NewEmailAnalyzerService(pool, gmailService),
        profileAnalyzer: services.NewProfileAnalyzerAgent(cfg),
        appCustomizer:   services.NewAppCustomizerAgent(cfg),
        gmailService:    gmailService,
    }
}
```

**Handler Method Pattern:**
```go
func (h *OSAOnboardingHandler) GetUserAnalysisStatus(c *gin.Context) {
    userID := c.Param("user_id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
        return
    }

    // Query database
    var status, errorMessage *string
    var insights, toolsUsed []byte
    var totalEmails *int

    err := h.pool.QueryRow(c.Request.Context(), `
        SELECT status, insights, tools_used, total_emails_analyzed
        FROM onboarding_user_analysis
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `, userID).Scan(&status, &insights, &toolsUsed, &totalEmails)

    if err != nil {
        // Not found - return not_started
        c.JSON(http.StatusOK, gin.H{
            "status":       "not_started",
            "insights":     []string{},
            "tools":        []string{},
            "total_emails": 0,
        })
        return
    }

    // Parse JSON arrays
    insightsArray := parseJSONArray(insights)
    toolsArray := parseJSONArray(toolsUsed)

    c.JSON(http.StatusOK, gin.H{
        "status":       derefString(status),
        "insights":     insightsArray,
        "tools":        toolsArray,
        "total_emails": *totalEmails,
    })
}
```

### Error Handling Strategy

**Non-Blocking Background Work:**
- OAuth callback NEVER fails due to analysis errors
- Analysis runs in goroutine with its own error handling
- Errors stored in `onboarding_user_analysis.error_message`
- Frontend gracefully degrades with fallback insights

**Example:**
```go
go func() {
    if err := h.runGmailAnalysis(ctx, userID); err != nil {
        log.Printf("❌ Gmail analysis failed: %v", err)
        // Store error but DON'T fail the request
        h.pool.Exec(ctx, `
            INSERT INTO onboarding_user_analysis (
                user_id, status, error_message
            ) VALUES ($1, 'failed', $2)
            ON CONFLICT (user_id, workspace_id) DO UPDATE
            SET status = 'failed', error_message = EXCLUDED.error_message
        `, userID, err.Error())
    }
}()
```

### Logging Standards

**Use slog (structured logging):**
```go
slog.Info("Starting OSA onboarding analysis",
    "user_id", userID,
    "workspace_id", workspaceID,
    "max_emails", maxEmails,
)

slog.Error("Email analysis failed",
    "error", err,
    "user_id", userID,
)
```

**NEVER use fmt.Printf in production code!**

---

## Database Schema

### Table 1: `user`

**Purpose:** Core user account data

**Key Columns:**
```sql
CREATE TABLE "user" (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    "emailVerified" BOOLEAN DEFAULT false,
    image VARCHAR(500),
    username VARCHAR(50) UNIQUE,
    onboarding_completed BOOLEAN DEFAULT false,
    "createdAt" TIMESTAMPTZ DEFAULT NOW(),
    "updatedAt" TIMESTAMPTZ DEFAULT NOW()
);
```

**Onboarding Flow:**
1. User created with `onboarding_completed = FALSE` on OAuth
2. `username` set when user claims username
3. `onboarding_completed = TRUE` when user finishes onboarding

---

### Table 2: `user_integrations`

**Purpose:** Store encrypted OAuth tokens for Gmail access

**Schema:**
```sql
CREATE TABLE user_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider_id VARCHAR(100) NOT NULL,  -- 'google_gmail'
    status VARCHAR(50) DEFAULT 'connected',

    -- OAuth Tokens (ENCRYPTED)
    access_token_encrypted BYTEA,
    refresh_token_encrypted BYTEA,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT[],

    -- External Account Info
    external_account_id VARCHAR(255),
    external_account_name VARCHAR(255),

    metadata JSONB DEFAULT '{}'::jsonb,

    connected_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider_id)
);
```

**Security Note:**
- `access_token_encrypted` and `refresh_token_encrypted` should be AES-encrypted
- Currently stored as plain bytea (TODO: encrypt before production)
- Use `TOKEN_ENCRYPTION_KEY` environment variable for AES key

---

### Table 3: `onboarding_user_analysis`

**Migration:** `054_onboarding_user_analysis.sql`

**Purpose:** Store AI-generated profile analysis results

**Schema:**
```sql
CREATE TABLE onboarding_user_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,

    -- AI Analysis Results
    insights JSONB DEFAULT '[]'::jsonb,           -- ["phrase1", "phrase2", "phrase3"]
    interests JSONB DEFAULT '[]'::jsonb,          -- ["interest1", "interest2", ...]
    tools_used JSONB DEFAULT '[]'::jsonb,         -- ["Figma", "Notion", ...]
    profile_summary TEXT,                         -- Full narrative summary

    -- Email Metadata
    email_metadata JSONB DEFAULT '{}'::jsonb,
    total_emails_analyzed INTEGER DEFAULT 0,
    sender_domains JSONB DEFAULT '[]'::jsonb,
    detected_patterns JSONB DEFAULT '{}'::jsonb,

    -- AI Provider Tracking
    analysis_model VARCHAR(100) NOT NULL,         -- "llama-3.3-70b-versatile"
    ai_provider VARCHAR(50) NOT NULL,             -- "groq"
    analysis_tokens_used INTEGER DEFAULT 0,
    analysis_duration_ms INTEGER,

    -- Status
    status VARCHAR(50) DEFAULT 'analyzing',       -- analyzing | completed | failed
    error_message TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    UNIQUE(user_id, workspace_id)
);

CREATE INDEX idx_onboarding_analysis_user ON onboarding_user_analysis(user_id);
CREATE INDEX idx_onboarding_analysis_status ON onboarding_user_analysis(status);
```

**Example Row:**
```json
{
  "id": "a1b2c3d4-...",
  "user_id": "usr_xyz123",
  "workspace_id": "00000000-0000-0000-0000-000000000000",
  "insights": [
    "No-code builder energy, big time",
    "Design tools are your playground",
    "AI-curious, testing new platforms"
  ],
  "interests": ["no-code development", "UI/UX design", "AI automation"],
  "tools_used": ["Figma", "Notion", "Zapier", "GitHub"],
  "profile_summary": "You're a no-code enthusiast with a strong design background...",
  "total_emails_analyzed": 100,
  "analysis_model": "llama-3.3-70b-versatile",
  "ai_provider": "groq",
  "status": "completed",
  "completed_at": "2026-01-19T10:45:32Z"
}
```

---

### Table 4: `onboarding_starter_apps`

**Migration:** `055_onboarding_starter_apps.sql`

**Purpose:** Store personalized starter app recommendations

**Schema:**
```sql
CREATE TABLE onboarding_starter_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    analysis_id UUID NOT NULL REFERENCES onboarding_user_analysis(id) ON DELETE CASCADE,

    -- App Details
    title VARCHAR(255) NOT NULL,
    description TEXT,
    icon_emoji VARCHAR(10),
    icon_url VARCHAR(500),
    category VARCHAR(100),                        -- tracker | companion | feedback | daily

    -- AI Customization
    reasoning TEXT,                               -- Why recommended
    customization_prompt TEXT NOT NULL,
    based_on_interests JSONB DEFAULT '[]'::jsonb,
    based_on_tools JSONB DEFAULT '[]'::jsonb,

    -- Module Customization
    base_module VARCHAR(100),                     -- CRM | Tasks | Projects | null
    module_customizations JSONB DEFAULT '{}'::jsonb,

    -- Generation Tracking
    status VARCHAR(50) DEFAULT 'pending',         -- pending | generating | ready | failed
    osa_workflow_id VARCHAR(255),
    error_message TEXT,

    -- AI Provider
    generation_model VARCHAR(100),
    ai_provider VARCHAR(50),
    generation_tokens_used INTEGER DEFAULT 0,
    generation_duration_ms INTEGER,

    display_order INTEGER DEFAULT 0,              -- 1-4

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    UNIQUE(user_id, workspace_id, display_order)
);

CREATE INDEX idx_starter_apps_user ON onboarding_starter_apps(user_id);
CREATE INDEX idx_starter_apps_analysis ON onboarding_starter_apps(analysis_id);
CREATE INDEX idx_starter_apps_display_order ON onboarding_starter_apps(workspace_id, display_order);
```

**Example Row:**
```json
{
  "id": "e5f6g7h8-...",
  "user_id": "usr_xyz123",
  "workspace_id": "00000000-0000-0000-0000-000000000000",
  "analysis_id": "a1b2c3d4-...",
  "title": "Design Tracker",
  "description": "Organize Figma projects, track feedback, and manage design sprints",
  "icon_emoji": "🎨",
  "category": "tracker",
  "reasoning": "You use Figma extensively and collaborate on design projects",
  "based_on_interests": ["UI/UX design", "collaboration"],
  "based_on_tools": ["Figma", "Notion"],
  "base_module": "Projects",
  "module_customizations": {"theme": "design-focused"},
  "status": "ready",
  "display_order": 1,
  "generation_model": "llama-3.3-70b-versatile",
  "ai_provider": "groq"
}
```

---

### Table 5: `onboarding_email_metadata`

**Migration:** `056_onboarding_email_metadata.sql`

**Purpose:** Store extracted metadata from individual emails (detailed breakdown)

**Schema:**
```sql
CREATE TABLE onboarding_email_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    analysis_id UUID NOT NULL REFERENCES onboarding_user_analysis(id) ON DELETE CASCADE,

    email_id UUID,                                -- Reference to emails table
    external_id VARCHAR(255),                     -- Gmail message ID

    -- Extracted Data
    sender_domain VARCHAR(255),
    sender_email VARCHAR(255),
    subject_keywords JSONB DEFAULT '[]'::jsonb,
    body_keywords JSONB DEFAULT '[]'::jsonb,
    detected_tools JSONB DEFAULT '[]'::jsonb,
    detected_topics JSONB DEFAULT '[]'::jsonb,

    -- Classification
    category VARCHAR(100),                        -- work | personal | marketing | newsletter
    sentiment VARCHAR(50),                        -- positive | neutral | negative
    importance_score DECIMAL(3, 2),               -- 0.00 to 1.00

    email_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_email_metadata_user ON onboarding_email_metadata(user_id);
CREATE INDEX idx_email_metadata_analysis ON onboarding_email_metadata(analysis_id);
CREATE INDEX idx_email_metadata_sender_domain ON onboarding_email_metadata(sender_domain);
```

**Note:** Currently NOT populated by email analyzer (future enhancement for detailed per-email tracking)

---

## Key Implementation Files

### Backend Files

```
desktop/backend-go/
├── internal/
│   ├── handlers/
│   │   ├── auth_google.go                   # OAuth flow + Gmail token storage
│   │   └── osa_onboarding.go                # Analysis status endpoints
│   ├── services/
│   │   ├── onboarding_email_analyzer.go     # Gmail metadata extraction
│   │   ├── onboarding_profile_analyzer.go   # AI profile analysis (Groq)
│   │   ├── onboarding_app_customizer.go     # Starter app generation
│   │   ├── onboarding_ai_service.go         # Groq LLM wrapper
│   │   ├── onboarding_service.go            # Orchestration
│   │   └── onboarding_validation.go         # Input validation
│   ├── integrations/
│   │   └── google/
│   │       ├── provider.go                  # Google OAuth provider
│   │       └── gmail.go                     # Gmail API client
│   └── database/
│       └── migrations/
│           ├── 054_onboarding_user_analysis.sql
│           ├── 055_onboarding_starter_apps.sql
│           └── 056_onboarding_email_metadata.sql
```

### Frontend Files

```
frontend/src/
├── routes/
│   └── onboarding/
│       ├── +layout.svelte                   # Shared onboarding layout
│       ├── +page.svelte                     # Screen 1: Welcome
│       ├── meet-osa/+page.svelte            # Screen 2: Meet OSA
│       ├── signin/+page.svelte              # Screen 3: Sign In
│       ├── gmail/+page.svelte               # Screen 4: Gmail Connection
│       ├── username/+page.svelte            # Screen 5: Username
│       ├── analyzing/+page.svelte           # Screen 6: Analysis 1
│       ├── analyzing-2/+page.svelte         # Screen 7: Analysis 2
│       ├── analyzing-3/+page.svelte         # Screen 8: Analysis 3
│       ├── starter-apps/+page.svelte        # Screen 9: Starter Apps
│       └── ready/+page.svelte               # Screen 10: Ready
├── lib/
│   ├── stores/
│   │   ├── onboardingStore.ts               # Main onboarding state
│   │   └── onboardingAnalysis.ts            # Analysis polling store
│   ├── components/
│   │   └── osa/
│   │       ├── PillButton.svelte            # Primary button
│   │       ├── GlassCard.svelte             # Card component
│   │       └── AppCard.svelte               # Starter app card
│   └── api/
│       ├── osa-onboarding.ts                # Analysis API client
│       └── users.ts                         # User API client
```

---

## Configuration & Deployment

### Environment Variables

**Backend (.env):**
```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/businessos

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/v1/auth/google/callback

# AI Provider (Groq)
GROQ_API_KEY=your-groq-api-key
AI_PROVIDER=groq

# Session/Auth
SECRET_KEY=your-secret-key-change-in-production
TOKEN_ENCRYPTION_KEY=your-aes-256-key  # TODO: Implement token encryption

# Environment
ENVIRONMENT=development  # development | production
COOKIE_DOMAIN=          # Empty for localhost
ALLOW_CROSS_ORIGIN=false
```

**Frontend (.env):**
```bash
# Backend API URL
VITE_API_URL=http://localhost:8001

# Public app URL
PUBLIC_APP_URL=http://localhost:5173
```

### Google Cloud Console Setup

1. **Create OAuth Client:**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Enable Gmail API
   - Create OAuth 2.0 credentials
   - Add authorized redirect URI: `http://localhost:8001/api/v1/auth/google/callback`

2. **OAuth Consent Screen:**
   - App name: BusinessOS
   - Scopes:
     - `userinfo.email`
     - `userinfo.profile`
     - `https://mail.google.com/` (requires verification for production)

3. **Test Users:**
   - Add test users while app is in testing mode
   - For production: Submit for OAuth verification

### Deployment Checklist

**Pre-Production:**
- [ ] Implement token encryption (use `TOKEN_ENCRYPTION_KEY`)
- [ ] Set up production Google OAuth client
- [ ] Configure production redirect URI
- [ ] Enable HTTPS/secure cookies
- [ ] Set `ENVIRONMENT=production`
- [ ] Rotate `SECRET_KEY`
- [ ] Set proper `COOKIE_DOMAIN`

**Security:**
- [ ] Encrypt OAuth tokens in DB
- [ ] Enable CSRF protection
- [ ] Use secure session cookies
- [ ] Implement rate limiting
- [ ] Add input validation
- [ ] Sanitize user inputs

**Monitoring:**
- [ ] Track analysis success/failure rates
- [ ] Monitor Groq API usage/costs
- [ ] Log error rates
- [ ] Track onboarding completion rates

---

## Testing & Debugging

### Manual Testing Flow

**Complete End-to-End Test:**

1. **Start Fresh:**
   - Clear browser cookies
   - Clear localStorage
   - Delete test user from DB

2. **OAuth Flow:**
   - Visit `http://localhost:5173/onboarding`
   - Click "Get Started"
   - Click "Sign in with Google"
   - Grant permissions (Gmail access)
   - Verify redirect to `/onboarding/gmail`

3. **Gmail Connection:**
   - Confirm "Gmail Connected" shows
   - Check DB: `user_integrations` row exists for user
   - Verify tokens are stored

4. **Username:**
   - Enter username (test validation: <3 chars, special chars)
   - Verify availability check works
   - Submit and verify redirect to `/onboarding/analyzing`

5. **Analysis Screens:**
   - Wait for analysis to complete (check logs)
   - Verify insight messages show (not generic fallback)
   - Confirm auto-advance every 2 seconds
   - Check DB: `onboarding_user_analysis` status = `completed`

6. **Starter Apps:**
   - Verify 3-4 apps display
   - Check that apps are personalized (not generic)
   - Verify DB: `onboarding_starter_apps` rows exist

7. **Ready Screen:**
   - Confirm completion message
   - Click "Enter BusinessOS"
   - Verify redirect to `/dashboard`
   - Check DB: `user.onboarding_completed = true`

### Backend Debugging

**Check Analysis Status:**
```sql
SELECT
    user_id,
    status,
    insights,
    tools_used,
    total_emails_analyzed,
    created_at,
    completed_at,
    error_message
FROM onboarding_user_analysis
WHERE user_id = 'usr_xyz123';
```

**Check Gmail Tokens:**
```sql
SELECT
    user_id,
    provider_id,
    status,
    scopes,
    token_expires_at,
    connected_at
FROM user_integrations
WHERE user_id = 'usr_xyz123' AND provider_id = 'google_gmail';
```

**Check Starter Apps:**
```sql
SELECT
    title,
    description,
    reasoning,
    based_on_interests,
    based_on_tools,
    status,
    display_order
FROM onboarding_starter_apps
WHERE user_id = 'usr_xyz123'
ORDER BY display_order;
```

### Common Issues & Solutions

**Issue: Analysis stuck on "analyzing"**
- Check backend logs for errors
- Verify Gmail API credentials are valid
- Check if analysis goroutine crashed
- Manually set status to `failed` and retry

**Issue: Generic insights show (not AI-generated)**
- Check Groq API key is valid
- Verify analysis completed successfully
- Check `onboarding_user_analysis.insights` is populated
- Look for error_message in DB

**Issue: OAuth redirect fails**
- Verify `GOOGLE_REDIRECT_URI` matches Google Console
- Check cookie domain settings
- Ensure CORS is configured correctly
- Verify state parameter matches

**Issue: Username already taken error**
- Check if username exists: `SELECT * FROM "user" WHERE username = 'test'`
- Verify uniqueness constraint
- Test with different username

### Logs to Watch

**Backend (Go):**
```
📧 [Gmail] Storing tokens and starting analysis for user: usr_xyz123
✅ [Gmail] Tokens stored successfully for user: usr_xyz123
🔍 [Gmail Analysis] Starting background analysis for user: usr_xyz123
📊 [Analysis] Analyzing recent emails for user: usr_xyz123
✅ [Analysis] Email analysis complete: 100 emails analyzed, 5 tools detected
💾 [Analysis] Storing results for user: usr_xyz123
✅ [Analysis] Results stored successfully for user: usr_xyz123
```

**Frontend (Console):**
```
[Analyzing] Starting analysis polling for user: usr_xyz123
[Polling] Status: analyzing, insights: []
[Polling] Status: completed, insights: ["No-code builder...", "Design tools...", ...]
[Analyzing] Analysis complete, auto-advancing...
```

---

## Future Enhancements

### Phase 2 Features

1. **Token Encryption:**
   - Implement AES-256 encryption for OAuth tokens
   - Use `TOKEN_ENCRYPTION_KEY` env variable
   - Encrypt before storing in `user_integrations`

2. **SSE Streaming:**
   - Replace polling with real-time SSE
   - Stream analysis progress updates
   - Show live email processing count

3. **Advanced Analysis:**
   - Populate `onboarding_email_metadata` table
   - Per-email classification
   - Sentiment analysis
   - Importance scoring

4. **App Generation:**
   - Actually generate starter apps (not just recommendations)
   - Use OSA Build agent to create apps
   - Track generation progress via `osa_workflow_id`

5. **Customization:**
   - Allow users to edit starter apps
   - Choose different base modules
   - Regenerate specific apps

### Analytics to Track

- Onboarding completion rate (% who finish)
- Drop-off points (which screen loses users)
- Analysis success rate
- Average analysis duration
- Tool detection accuracy
- User satisfaction with starter apps

---

## Conclusion

The BusinessOS onboarding system is a sophisticated, AI-powered flow that:
- Seamlessly integrates Google OAuth with Gmail analysis
- Analyzes real user data to generate personalized insights
- Uses Groq AI (Llama 3.3 70B) for intelligent profiling
- Recommends tailored "starter apps" based on user workflow
- Provides a smooth, delightful UX with real-time feedback

**Key Technical Achievements:**
- Non-blocking background analysis (doesn't slow OAuth)
- Real-time polling for status updates
- Graceful degradation with fallback insights
- Type-safe backend with SQLC
- Reactive frontend with Svelte 5 stores
- Proper separation of concerns (Handler → Service → Repository)

**For New Team Members:**
- Start with the user journey flow
- Read the OAuth integration section
- Understand the analysis pipeline
- Explore the database schema
- Test the flow end-to-end

**For Debugging:**
- Check backend logs first
- Query DB tables for status
- Verify OAuth tokens are stored
- Test each screen individually

---

**Document Maintained By:** BusinessOS Engineering Team
**Questions?** Check code comments or ask in #engineering-onboarding
**Last Reviewed:** January 19, 2026
