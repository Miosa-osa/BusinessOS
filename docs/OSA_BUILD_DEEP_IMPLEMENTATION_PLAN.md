# OSA Build - Deep Implementation Plan
## Complete Redesign with iOS Visual System & Onboarding Flow

**Date:** January 18, 2026
**Branch:** `feature/ios-desktop-flow-migration`
**Status:** Planning Phase
**Estimated Timeline:** 16 weeks

---

## 🎯 EXECUTIVE SUMMARY

This document outlines the **complete architectural transformation** from traditional BusinessOS to the new **OSA Build** experience, inspired by the iOS Wabi app's onboarding and visual design system.

### What's Changing

**Current State:**
- Traditional login/register pages
- Generic onboarding (Grok conversation)
- Standard dark/light theme
- Plain forms with minimal styling
- User goes directly to dashboard after signup

**New State:**
- **OSA Build onboarding experience** - AI-generated personalized apps on Day 1
- **iOS-inspired visual design** - Gradients, glassmorphism, pill buttons
- **"Build Your OS" philosophy** - Users create their Operating System through AI
- **Social discovery** - Explore, remix, share apps built by others

---

## 🎨 PART 1: VISUAL DESIGN SYSTEM (iOS Wabi Inspired)

### Color Palette

Based on the iOS screenshots analyzed, here's the complete design system:

#### Gradient Backgrounds (Onboarding)

```css
/* Gradient 1: Soft Purple-Pink */
--gradient-welcome: linear-gradient(180deg, #E8F4F8 0%, #F5E8F8 100%);

/* Gradient 2: Blue-Purple */
--gradient-signin: linear-gradient(180deg, #D4F1F9 0%, #F0D9F5 100%);

/* Gradient 3: Light Blue-Peach */
--gradient-personalization: linear-gradient(180deg, #E3E8F8 0%, #F5EDE8 100%);

/* Gradient 4: Warm Peach-Pink */
--gradient-apps-showcase: linear-gradient(180deg, #FFF5E6 0%, #FFECF1 100%);

/* Gradient 5: Cool Mint-Blue */
--gradient-ready: linear-gradient(180deg, #E8F9F9 0%, #EFF5FF 100%);
```

#### App Background

```css
/* Main app background - Clean white/gray */
--bg-main: #F8F9FA;
--bg-card: #FFFFFF;
--bg-elevated: #FFFFFF;

/* Dark mode */
.dark {
  --bg-main: #1C1C1E;
  --bg-card: #2C2C2E;
  --bg-elevated: #3A3A3C;
}
```

#### Glassmorphism

```css
/* Glass card effect (used in app cards, modals) */
.glass {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow:
    0 8px 32px 0 rgba(31, 38, 135, 0.15),
    inset 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.dark .glass {
  background: rgba(28, 28, 30, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow:
    0 8px 32px 0 rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 255, 255, 0.05);
}
```

#### Text Colors

```css
--text-primary: #1A1A1A;
--text-secondary: #666666;
--text-tertiary: #999999;
--text-muted: #BEBEBE;

.dark {
  --text-primary: #F5F5F7;
  --text-secondary: #A1A1A6;
  --text-tertiary: #8E8E93;
  --text-muted: #6E6E73;
}
```

#### Accent Colors

```css
/* Primary accents from iOS app */
--accent-blue: #4A90E2;
--accent-purple: #A855F7;
--accent-pink: #EC4899;
--accent-orange: #FF6B35; /* Feature Feedback Hub color */
--accent-green: #10B981;
```

### Typography

```css
/* Use SF Pro for that iOS feel */
:root {
  --font-display: 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-text: 'SF Pro Text', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'SF Mono', 'Menlo', 'Monaco', monospace;
}

/* Size scale */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
--text-4xl: 2.25rem;   /* 36px */

/* Weight scale */
--font-light: 300;
--font-normal: 400;
--font-medium: 500;
--font-semibold: 600;
--font-bold: 700;
```

### Spacing & Layout

```css
/* Spacing scale (8px base) */
--space-0: 0;
--space-1: 0.25rem;  /* 4px */
--space-2: 0.5rem;   /* 8px */
--space-3: 0.75rem;  /* 12px */
--space-4: 1rem;     /* 16px */
--space-5: 1.25rem;  /* 20px */
--space-6: 1.5rem;   /* 24px */
--space-8: 2rem;     /* 32px */
--space-10: 2.5rem;  /* 40px */
--space-12: 3rem;    /* 48px */
--space-16: 4rem;    /* 64px */
--space-20: 5rem;    /* 80px */

/* Border radius */
--radius-sm: 8px;
--radius-md: 12px;
--radius-lg: 16px;
--radius-xl: 24px;
--radius-2xl: 32px;
--radius-full: 9999px; /* Pills */

/* Icon sizes */
--icon-xs: 16px;
--icon-sm: 20px;
--icon-md: 24px;
--icon-lg: 32px;
--icon-xl: 48px;
--icon-2xl: 64px;
--icon-3xl: 96px; /* App icons */
```

### Component Styles

#### Pill Buttons (iOS Style)

```css
.btn-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 0.9375rem; /* 15px */
  font-weight: 500;
  border-radius: 9999px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  border: none;
  font-family: var(--font-text);
}

.btn-pill-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 14px 0 rgba(102, 126, 234, 0.39);
}

.btn-pill-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px 0 rgba(102, 126, 234, 0.5);
}

.btn-pill-primary:active {
  transform: translateY(0);
}

.btn-pill-secondary {
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.1);
  color: var(--text-primary);
  backdrop-filter: blur(10px);
}

.btn-pill-secondary:hover {
  background: rgba(255, 255, 255, 1);
  border-color: rgba(0, 0, 0, 0.15);
}

.dark .btn-pill-secondary {
  background: rgba(58, 58, 60, 0.8);
  border-color: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.dark .btn-pill-secondary:hover {
  background: rgba(58, 58, 60, 1);
  border-color: rgba(255, 255, 255, 0.2);
}
```

#### Rounded Inputs (iOS Style)

```css
.input-rounded {
  width: 100%;
  padding: 0.875rem 1.125rem;
  font-size: 0.9375rem; /* 15px */
  font-family: var(--font-text);
  border: 1.5px solid transparent;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  color: var(--text-primary);
  transition: all 0.2s ease;
  backdrop-filter: blur(10px);
}

.input-rounded:focus {
  outline: none;
  border-color: rgba(102, 126, 234, 0.5);
  background: rgba(255, 255, 255, 1);
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.1);
}

.input-rounded::placeholder {
  color: var(--text-tertiary);
}

.dark .input-rounded {
  background: rgba(58, 58, 60, 0.6);
  border-color: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.dark .input-rounded:focus {
  background: rgba(58, 58, 60, 0.9);
  border-color: rgba(102, 126, 234, 0.6);
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.2);
}
```

#### App Card (Circular Icon with Glass Effect)

```css
.app-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
}

.app-card:hover {
  transform: translateY(-4px);
  box-shadow:
    0 20px 40px 0 rgba(31, 38, 135, 0.2),
    inset 0 0 0 1px rgba(255, 255, 255, 0.3);
}

.app-card-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  position: relative;
}

.app-card-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-card-title {
  font-size: 0.875rem; /* 14px */
  font-weight: 500;
  color: var(--text-primary);
  text-align: center;
  line-height: 1.3;
}

.app-card-usage {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem; /* 12px */
  font-weight: 600;
  border-radius: 999px;
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
  backdrop-filter: blur(4px);
}

.dark .app-card {
  background: rgba(44, 44, 46, 0.7);
  border-color: rgba(255, 255, 255, 0.1);
}

.dark .app-card:hover {
  box-shadow:
    0 20px 40px 0 rgba(0, 0, 0, 0.4),
    inset 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.dark .app-card-icon {
  border-color: rgba(255, 255, 255, 0.2);
}
```

#### Progress Dots (iOS Onboarding Style)

```css
.progress-dots {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
  align-items: center;
}

.progress-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.2);
  transition: all 0.3s ease;
}

.progress-dot.active {
  background: rgba(102, 126, 234, 1);
  width: 24px;
  border-radius: 4px;
}

.dark .progress-dot {
  background: rgba(255, 255, 255, 0.2);
}

.dark .progress-dot.active {
  background: rgba(102, 126, 234, 1);
}
```

---

## 📱 PART 2: CURRENT STATE ANALYSIS

### Current Authentication Flow

```
1. User visits app
2. Lands on /login or /register
3. Traditional form-based signup/login
4. After signup → Goes to /onboarding (Grok conversation)
5. After onboarding → Goes to /window (main app)
```

### Current Files Structure

```
frontend/src/routes/
├── login/+page.svelte           # Traditional login form
├── register/+page.svelte        # Traditional register form
├── onboarding/+page.svelte      # Grok conversational onboarding
├── window/+page.svelte          # Main app entry (embedded windows)
├── (app)/                       # Protected app routes
│   ├── dashboard/+page.svelte
│   ├── tasks/+page.svelte
│   └── ...
└── auth/
    └── callback/+page.svelte    # OAuth callback handler
```

### Current Design Issues

**Problems:**
1. ❌ Generic login/register pages - no personality
2. ❌ No visual excitement - plain forms
3. ❌ No personalization on Day 1 - empty dashboard
4. ❌ Disconnected onboarding - separate from app building
5. ❌ No "aha moment" - users don't see value immediately

**What We Need:**
1. ✅ Exciting first launch experience
2. ✅ AI-generated starter apps on Day 1
3. ✅ Beautiful iOS-inspired gradients and glassmorphism
4. ✅ Smooth, delightful animations
5. ✅ "Build Your OS" messaging from the start

---

## 🚀 PART 3: NEW OSA BUILD FLOW

### New First Launch Experience

```
1. User installs OSA Build (desktop app)
2. First launch → Onboarding flow begins

   Screen 1: Welcome to OSA Build (gradient background)
   Screen 2: Meet OSA (AI agent introduction)
   Screen 3: Sign in with Google/Apple (OAuth, pill buttons)
   Screen 4: Connect Gmail (optional, for personalization)
   Screen 5: Claim username (inline validation)
   Screen 6-8: AI Analyzing... (3 loading messages with personality insights)
   Screen 9-12: Personalized Apps Showcase (4 AI-generated apps)
   Screen 13: Your OS is Ready! (success state)

3. User enters main OS → Home screen with 4 starter apps
4. User can:
   - Use starter apps immediately
   - Build new apps (click + → talk to OSA)
   - Discover apps from others (Explore tab)
   - Customize apps (right-click → Edit → Chat with OSA)
```

### New Files Structure

```
frontend/src/routes/
├── onboarding/                  # NEW - Complete onboarding flow
│   ├── +page.svelte            # Orchestrator component
│   ├── steps/
│   │   ├── 01-Welcome.svelte
│   │   ├── 02-MeetOSA.svelte
│   │   ├── 03-SignIn.svelte
│   │   ├── 04-ConnectData.svelte
│   │   ├── 05-Username.svelte
│   │   ├── 06-Analyzing.svelte  # Shows 3 analysis messages
│   │   ├── 07-AppsShowcase.svelte # Shows 4 personalized apps
│   │   └── 08-Ready.svelte
│   └── components/
│       ├── ProgressDots.svelte
│       ├── GradientBackground.svelte
│       └── AppCard.svelte
├── home/+page.svelte            # NEW - Main OS home screen (replaces /window)
├── explore/+page.svelte         # NEW - Discover apps from others
├── profile/+page.svelte         # NEW - User profile
├── builder/                     # NEW - App builder interface
│   └── [appId]/+page.svelte    # 4-tab builder (Chat, General, Icon, Prompts)
└── (deprecated)/
    ├── login/                   # OLD - To be removed
    ├── register/                # OLD - To be removed
    └── window/                  # OLD - To be replaced
```

---

## 🎯 PART 4: BACKEND CHANGES REQUIRED

### New Database Tables

```sql
-- Users table (modify existing)
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS handle VARCHAR(50) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS gmail_connected BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS gmail_access_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS gmail_refresh_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS interests JSONB; -- AI-extracted interests
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_step INTEGER DEFAULT 0;

-- Mini-apps table (NEW)
CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    icon_prompt TEXT,
    visibility VARCHAR(20) DEFAULT 'private', -- 'public' | 'private'
    components JSONB, -- Array of app components
    usage_percentage INTEGER DEFAULT 0,
    remixed_from UUID REFERENCES apps(id), -- Track remix source
    is_starter_app BOOLEAN DEFAULT false, -- AI-generated starter apps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- App components (for multi-model architecture)
CREATE TABLE app_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'text' | 'image' | 'code' | 'data'
    prompt TEXT,
    model_provider VARCHAR(50), -- 'openai' | 'anthropic' | 'google'
    model_name VARCHAR(100),
    model_config JSONB, -- temperature, maxTokens, etc.
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Social interactions
CREATE TABLE interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- 'like' | 'comment' | 'remix' | 'get' | 'share'
    comment_text TEXT,
    parent_comment_id UUID REFERENCES interactions(id), -- For threading
    created_at TIMESTAMP DEFAULT NOW()
);

-- Follows
CREATE TABLE follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

-- App embeddings (for semantic search)
CREATE TABLE app_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    embedding vector(1536), -- OpenAI embedding dimension
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_apps_user_id ON apps(user_id);
CREATE INDEX idx_apps_visibility ON apps(visibility);
CREATE INDEX idx_interactions_app_id ON interactions(app_id);
CREATE INDEX idx_interactions_user_id ON interactions(user_id);
CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);
CREATE INDEX idx_app_embeddings_vector ON app_embeddings
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

### New API Endpoints

```
# Onboarding
POST   /api/onboarding/start           - Initialize onboarding session
POST   /api/onboarding/gmail/connect   - Connect Gmail + analyze
POST   /api/onboarding/username/check  - Check username availability
POST   /api/onboarding/username/claim  - Claim username
POST   /api/onboarding/analyze          - Trigger AI analysis
GET    /api/onboarding/generate-apps    - Generate personalized starter apps
POST   /api/onboarding/complete         - Mark onboarding complete

# Apps
GET    /api/apps                        - List user's apps
GET    /api/apps/:id                    - Get app details
POST   /api/apps                        - Create new app
PATCH  /api/apps/:id                    - Update app
DELETE /api/apps/:id                    - Delete app
POST   /api/apps/:id/remix              - Remix app (create copy)

# Builder (AI Interface)
POST   /api/builder/chat                - Send chat message (SSE response)
POST   /api/builder/icon                - Regenerate icon
POST   /api/builder/apply               - Apply changes to app
GET    /api/builder/models              - List available models

# Social
GET    /api/users/:handle               - Get user profile
POST   /api/users/:id/follow            - Follow user
DELETE /api/users/:id/follow            - Unfollow user
GET    /api/apps/:id/likes              - Get app likes
POST   /api/apps/:id/like               - Like app
DELETE /api/apps/:id/like               - Unlike app
GET    /api/apps/:id/comments           - Get app comments
POST   /api/apps/:id/comments           - Add comment
POST   /api/apps/:id/share              - Share app

# Explore
GET    /api/explore/featured            - Get featured apps
GET    /api/explore/popular             - Get popular apps
GET    /api/explore/recent              - Get recently added apps
GET    /api/explore/search              - Search apps (query param)
```

### Backend Services to Implement

```go
// internal/services/

// onboarding.go
type OnboardingService struct {
    gmailClient   *gmail.Service
    aiClient      *ai.Client
    appGenerator  *AppGenerator
}

func (s *OnboardingService) ConnectGmail(code string) (*GmailConnection, error)
func (s *OnboardingService) AnalyzeUserData(userID uuid.UUID) (*UserAnalysis, error)
func (s *OnboardingService) GenerateStarterApps(userID uuid.UUID, interests []string) ([]*App, error)

// app.go
type AppService struct {
    db       *sql.DB
    aiClient *ai.Client
}

func (s *AppService) CreateApp(userID uuid.UUID, request CreateAppRequest) (*App, error)
func (s *AppService) RemixApp(userID uuid.UUID, appID uuid.UUID) (*App, error)
func (s *AppService) UpdateApp(appID uuid.UUID, changes AppChanges) (*App, error)

// builder.go
type BuilderService struct {
    aiClient      *ai.Client
    modelRouter   *ModelRouter
    iconGenerator *IconGenerator
}

func (s *BuilderService) ProcessChatMessage(ctx context.Context, appID uuid.UUID, message string) (chan string, error)
func (s *BuilderService) RegenerateIcon(appID uuid.UUID, prompt string) (*Icon, error)
func (s *BuilderService) ApplyChanges(appID uuid.UUID, changes []Change) error

// social.go
type SocialService struct {
    db *sql.DB
}

func (s *SocialService) FollowUser(followerID, followingID uuid.UUID) error
func (s *SocialService) LikeApp(userID, appID uuid.UUID) error
func (s *SocialService) AddComment(userID, appID uuid.UUID, text string) (*Comment, error)

// personalization.go
type PersonalizationService struct {
    gmailClient     *gmail.Service
    nlpClient       *ai.Client
    embeddingClient *ai.Client
}

func (s *PersonalizationService) ExtractInterests(userID uuid.UUID) ([]string, error)
func (s *PersonalizationService) RecommendApps(userID uuid.UUID) ([]*App, error)
```

---

## 🛠️ PART 5: IMPLEMENTATION ROADMAP

### Phase 1: Foundation & Design System (Weeks 1-2)

**Goals:**
- Set up new visual design system
- Create reusable UI components
- Prepare database schema

**Backend Tasks:**
```
□ Create database migration for new tables (apps, app_components, interactions, follows)
□ Set up pgvector extension
□ Create base API structure for onboarding endpoints
□ Set up Gmail OAuth flow
□ Configure multi-provider AI routing (OpenRouter/direct)
```

**Frontend Tasks:**
```
□ Create new app.css with iOS design system
  □ Gradient backgrounds
  □ Glassmorphism styles
  □ Pill button styles
  □ Rounded input styles
  □ App card styles
  □ Progress dots

□ Build reusable components in lib/components/osa/
  □ GradientBackground.svelte
  □ PillButton.svelte
  □ RoundedInput.svelte
  □ AppCard.svelte
  □ ProgressDots.svelte
  □ GlassCard.svelte

□ Create component library Storybook (optional but recommended)
```

**Deliverable:** Design system ready, components built, database schema in place

---

### Phase 2: Onboarding Flow - Steps 1-5 (Weeks 3-4)

**Goals:**
- Build first 5 onboarding screens
- Implement OAuth flow
- Username validation

**Backend Tasks:**
```
□ POST /api/onboarding/start - Create onboarding session
□ Gmail OAuth integration (access token, refresh token storage)
□ POST /api/onboarding/username/check - Real-time username availability
□ POST /api/onboarding/username/claim - Claim username
□ Session management for onboarding state
```

**Frontend Tasks:**
```
□ Create /onboarding route structure
□ Build step components:
  □ 01-Welcome.svelte (gradient, OSA logo animation)
  □ 02-MeetOSA.svelte (agent introduction)
  □ 03-SignIn.svelte (OAuth buttons, pill style)
  □ 04-ConnectData.svelte (Gmail connection modal)
  □ 05-Username.svelte (username input with live validation)

□ Build onboarding orchestrator (+page.svelte)
  □ Step navigation logic
  □ Progress tracking
  □ State management (Svelte stores)

□ Implement smooth transitions between steps
□ Add progress dots indicator
```

**Deliverable:** User can complete first 5 onboarding steps

---

### Phase 3: Onboarding Flow - AI Analysis & Apps (Weeks 5-6)

**Goals:**
- AI analyzes user data
- Generate 4 personalized starter apps
- Showcase apps to user

**Backend Tasks:**
```
□ POST /api/onboarding/analyze - Gmail analysis endpoint
  □ Fetch recent emails (last 100)
  □ NLP analysis for interests/tools/follows
  □ Store interests in user.interests JSONB

□ GET /api/onboarding/generate-apps - Generate starter apps
  □ Select 4 app templates based on interests
  □ Generate app icons (DALL-E/Midjourney)
  □ Create app records in database
  □ Mark as is_starter_app = true

□ App template system
  □ Define 20+ app templates (productivity, social, creative, etc.)
  □ Variable injection based on user interests
```

**Frontend Tasks:**
```
□ Build step components:
  □ 06-Analyzing.svelte (animated loading, 3 personality insights)
  □ 07-AppsShowcase.svelte (carousel of 4 apps)
  □ 08-Ready.svelte (success state, "Enter Your OS" button)

□ Implement AI analysis loading state
  □ Pulsing animation
  □ Rotating personality insights ("No-code builder energy...", etc.)

□ Build apps carousel
  □ Swipeable/keyboard navigable
  □ Show app icon, title, description, "why this app" explanation

□ Smooth transition to main app
```

**Deliverable:** AI generates personalized apps, user sees them in onboarding

---

### Phase 4: Main OS - Home Screen (Weeks 7-8)

**Goals:**
- Build main OS home screen
- Display user's apps in grid
- Basic app opening

**Backend Tasks:**
```
□ GET /api/apps - List user's apps endpoint
□ GET /api/apps/:id - Get single app details
□ DELETE /api/apps/:id - Delete app
□ PATCH /api/apps/:id - Update app metadata
```

**Frontend Tasks:**
```
□ Create /home route (replaces /window)
□ Build Home.svelte
  □ App grid layout (2-4 columns responsive)
  □ App cards with circular icons
  □ Usage percentage indicator
  □ Quick actions on hover/right-click

□ Build app opening logic
  □ Open app in new view/modal
  □ Display app content
  □ Close/minimize controls

□ Build header
  □ Notification bell
  □ Search icon
  □ Create (+) button
  □ Profile avatar

□ Build "Create App" modal
  □ Input for describing app
  □ "Build" button → calls AI
```

**Deliverable:** User lands on Home after onboarding, sees their 4 starter apps

---

### Phase 5: Builder Interface (Weeks 9-11)

**Goals:**
- Build 4-tab builder interface
- Implement Chat tab with AI streaming
- Icon regeneration
- Model selection

**Backend Tasks:**
```
□ POST /api/builder/chat - Chat endpoint with SSE streaming
  □ Stream AI responses in real-time
  □ Context awareness (previous messages, app state)
  □ Action detection (e.g., "add dark mode" → generates code change)

□ POST /api/builder/icon - Icon regeneration
  □ Call DALL-E/Midjourney with prompt
  □ Store new icon URL
  □ Return updated icon

□ GET /api/builder/models - List available models
  □ Return OpenAI models (GPT-4, GPT-3.5)
  □ Return Anthropic models (Claude 3.5 Sonnet, Claude 3 Haiku)
  □ Return Google models (Gemini Pro, Gemini Flash)

□ POST /api/builder/apply - Apply changes
  □ Validate changes
  □ Update app components
  □ Return success/failure
```

**Frontend Tasks:**
```
□ Create /builder/[appId] route
□ Build Builder.svelte with 4 tabs:

  Tab 1: Chat
    □ Chat history display
    □ Message streaming (SSE)
    □ Input field with send button
    □ Voice input (optional)

  Tab 2: General
    □ Title input (inline edit)
    □ Description input (inline edit)
    □ Visibility toggle (Public/Private)
    □ "Clear all app data" button

  Tab 3: Icon
    □ Current icon display (large)
    □ Regenerate button
    □ Icon prompt display/edit
    □ Loading state for regeneration

  Tab 4: Prompts
    □ List of app components
    □ Per-component prompt editor
    □ Per-component model selector
    □ Dropdown menus for models

□ Implement SSE streaming for chat
□ Build model selection dropdowns
□ Add change preview (optional)
```

**Deliverable:** Users can edit apps through conversational AI interface

---

### Phase 6: Social Features (Weeks 12-14)

**Goals:**
- Explore feed with public apps
- User profiles
- Like/comment/remix/share

**Backend Tasks:**
```
□ Implement visibility system (public apps only in explore)
□ GET /api/explore/featured - Curated featured apps
□ GET /api/explore/popular - Most liked/remixed apps
□ GET /api/explore/search - Full-text + vector search
□ GET /api/users/:handle - User profile with stats
□ POST /api/users/:id/follow - Follow system
□ POST /api/apps/:id/like - Like system
□ POST /api/apps/:id/comments - Comments with threading
□ POST /api/apps/:id/remix - Create editable copy
```

**Frontend Tasks:**
```
□ Create /explore route
□ Build Explore.svelte
  □ Featured section
  □ Popular section
  □ Search bar
  □ Infinite scroll

□ Create /profile/[handle] route
□ Build ProfileView.svelte
  □ User info (avatar, username, stats)
  □ Public/Private app tabs
  □ Follow button

□ Build AppDetail modal
  □ App display
  □ Stats (likes, comments, remixes)
  □ Action buttons (Remix, Get, Like, Share)
  □ Comments section

□ Build Comments.svelte
  □ Comment list with threading
  □ Add comment input
  □ Like comments

□ Build ShareModal.svelte
  □ App preview
  □ "Share" button
  □ Privacy notice
```

**Deliverable:** Users can discover, remix, and interact with others' apps

---

### Phase 7: Desktop Enhancements (Weeks 15-16)

**Goals:**
- Desktop-specific features
- Keyboard shortcuts
- Window management
- Polish & optimization

**Tasks:**
```
□ Multi-window support (open multiple apps simultaneously)
□ Keyboard shortcuts
  □ Cmd/Ctrl+N - New app
  □ Cmd/Ctrl+O - Open app
  □ Cmd/Ctrl+E - Edit current app
  □ Cmd/Ctrl+/ - Search
  □ Cmd/Ctrl+, - Settings

□ Window management
  □ Minimize/maximize
  □ Snap to edges
  □ Picture-in-picture mode

□ System tray integration
□ Auto-update system
□ File system integration (export/import apps)
□ Settings panel (model config, API keys)
□ Performance optimization
  □ Lazy loading
  □ Image optimization
  □ Caching strategy

□ Analytics (Posthog integration)
□ Error tracking
□ Onboarding tooltips
□ Help documentation
```

**Deliverable:** Production-ready desktop app with native OS integration

---

## 📊 PART 6: MIGRATION STRATEGY

### Handling Existing Users

**Option 1: Fresh Start (Recommended)**
- OSA Build is a new product
- Existing BusinessOS users stay on current system
- OSA Build launches as separate desktop app
- No migration needed

**Option 2: Gradual Migration**
- Add "Try OSA Build" banner in BusinessOS
- Allow users to opt-in to new experience
- Migrate user data (projects → apps)
- Provide rollback option

**Recommendation:** Option 1 - fresh start. OSA Build is different enough to warrant a new product launch.

---

## ✅ PART 7: SUCCESS METRICS

**Technical Metrics:**
```
□ Onboarding completion rate > 80%
□ Time to first app < 5 minutes
□ AI response time < 2 seconds (SSE start)
□ App creation success rate > 95%
□ System uptime > 99.9%
```

**User Metrics:**
```
□ Average apps created per user (first week) > 5
□ Daily active users (DAU) growth
□ Remix rate (apps remixed / total public apps) > 10%
□ Social engagement (comments, likes per app) > 3
□ Net Promoter Score (NPS) > 50
```

**Business Metrics:**
```
□ 1,000 beta users in first month
□ 10,000 total apps created
□ 500 daily remixes
□ User retention (7-day) > 40%
□ User retention (30-day) > 25%
```

---

## 🎯 PART 8: NEXT IMMEDIATE STEPS

### Week 1 Action Items

**Monday:**
1. Review this plan with team
2. Get design approval for visual system
3. Set up new branch: `feature/osa-build-v1`

**Tuesday-Wednesday:**
4. Create database migration script
5. Run migration on dev environment
6. Set up new API endpoint structure

**Thursday-Friday:**
7. Build design system in app.css
8. Create first 5 reusable components
9. Build Storybook for components (optional)

**Weekend:**
10. Start onboarding flow (screens 1-2)
11. Test on local environment

---

## 📚 PART 9: DOCUMENTATION TO CREATE

```
□ API_REFERENCE.md - Complete API documentation
□ DESIGN_SYSTEM.md - Visual design system guide
□ ONBOARDING_FLOW.md - Detailed onboarding spec
□ BUILDER_SPEC.md - Builder interface specification
□ SOCIAL_FEATURES.md - Social features specification
□ DEPLOYMENT_GUIDE.md - How to deploy OSA Build
□ USER_GUIDE.md - End-user documentation
```

---

## 🚨 RISKS & MITIGATION

**Risk 1: AI Costs**
- **Risk:** High API costs for image generation and chat
- **Mitigation:**
  - Cache generated icons
  - Use cheaper models for non-critical tasks
  - Implement usage limits per user
  - Pre-generate common app templates

**Risk 2: Gmail API Limits**
- **Risk:** Gmail API rate limits, approval delays
- **Mitigation:**
  - Make Gmail connection optional
  - Fallback to manual interest selection
  - Implement request queuing

**Risk 3: Onboarding Complexity**
- **Risk:** Users abandon onboarding (too many steps)
- **Mitigation:**
  - Allow "Skip" at any step
  - Save progress (resume later)
  - Track abandonment points, optimize

**Risk 4: Performance**
- **Risk:** Slow app with many animations/effects
- **Mitigation:**
  - Lazy load components
  - Optimize images/icons
  - Use virtual scrolling for long lists
  - Profile and optimize hot paths

---

## 🎉 CONCLUSION

This is a **complete product transformation**, not just a feature addition. We're building:

✅ A beautiful, iOS-inspired onboarding experience
✅ AI-powered personalization from Day 1
✅ A living, growing Operating System built by the user
✅ A social platform for discovering and remixing apps
✅ A desktop-class application with native integrations

**Timeline:** 16 weeks
**Effort:** ~1,600 hours (2 full-time devs for 16 weeks)
**Budget:** AI API costs (~$500-$2,000/month in beta)

**Ready to build the future of personal software. Let's go! 🚀**
