# iOS to Desktop Architecture Migration
## Wabi (iOS) → OSA Build with OSA (Desktop)

**Date:** January 18, 2026
**Version:** 1.0
**Project:** BusinessOS Desktop Application
**Branch:** `feature/ios-desktop-flow-migration`

---

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Complete User Flow Mapping](#complete-user-flow-mapping)
3. [Feature Analysis & Breakdown](#feature-analysis--breakdown)
4. [Desktop Architecture Design](#desktop-architecture-design)
5. [UI/UX Pattern Translation](#uiux-pattern-translation)
6. [Technical Implementation Roadmap](#technical-implementation-roadmap)
7. [Data Models & API Contracts](#data-models--api-contracts)

---

## 1. Executive Summary

### Overview
This document provides a comprehensive mapping of the iOS "Wabi" application flow to the desktop "OSA Build with OSA" application. The goal is to preserve the core user experience while adapting it for desktop form factors and enhancing it with desktop-specific capabilities.

### Key Insights from iOS App Analysis

**Core Value Proposition:**
- **Personalized software platform** - AI-generated mini-apps based on user interests
- **Social discovery** - Explore, remix, and share apps created by others
- **No-code builder** - Conversational AI interface for app creation/modification
- **Community-driven** - Comments, likes, follows, and app remixing

**User Journey Highlights:**
1. **Invite-only onboarding** with AI-powered personalization
2. **Gmail integration** for interest analysis and app generation
3. **Swipeable example apps** showcase during onboarding
4. **Grid-based home** with circular app icons
5. **Explore feed** with featured apps and social features
6. **Profile system** with public/private app separation
7. **Builder interface** with 4 tabs: Chat, General, Icon, Prompts
8. **Multi-model support** (GPT, Gemini, Claude) for different app components

### Desktop Adaptation Goals

**Preserve:**
- Personalization engine and AI-driven app generation
- Social features (explore, follow, remix, comments)
- Builder conversational interface
- Visual identity (gradient backgrounds, circular icons, clean typography)

**Enhance:**
- Multi-window/multi-app workflow
- Keyboard shortcuts and power-user features
- Larger canvas for app display and editing
- File system integration
- Desktop notifications

**New Capabilities:**
- Window management (minimize, maximize, snap)
- Drag-and-drop between apps
- System tray integration
- Local app storage and sync

---

## 2. Complete User Flow Mapping

### 📱 iOS App Flow (50+ Screens Mapped)

#### **PHASE 1: ONBOARDING (Screens 1-16)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 1 | **Waitlistinvite** | Display invite code for app access | Welcome modal with invite code validation |
| 2 | **AppLoadfirstlaunch** | Splash: "A new era of software is here" | Splash screen on first launch |
| 3 | **AppLaunchSwipeEffect** | Animated Wabi logo with swipe-up gesture | Animated logo with "Click to continue" |
| 4 | **AppAIIntro** | "Meet Wabi" with + icon introduction | "Meet OSA" intro screen |
| 5 | **AppSignupscreenwitheffect** | OAuth signup with bubble animation | OAuth modal with desktop-style animation |
| 6 | **AppSigninGooglePath** | Gmail connection explanation | Gmail/email connection modal |
| 7 | **AppSignUpUsername** | Username claim with availability check | Username input dialog |
| 8 | **AppusernameSelected** | Username confirmed state | Confirmation animation |
| 9 | **AppSigninprocessAIAnalysis** | AI analyzing user: "No-code builder energy" | Loading screen with AI analysis messages |
| 10 | **AppSignInAnalysis2** | AI analyzing: "Design tools are your playground" | Second analysis message |
| 11 | **AppSigninPersonalizaiton** | AI analyzing: "AI-curious, testing new platforms" | Third analysis message |
| 12 | **AppFirstExmapleView** | Personalized app 1: "No-Code Book Finds" | Carousel slide 1 |
| 13 | **App2ndExampleView** | Personalized app 2: "Motion Design Muse" | Carousel slide 2 |
| 14 | **App3rdExampleView** | Personalized app 3: "Feature feedback hub" | Carousel slide 3 |
| 15 | **AppViewExampleFinalView** | Final app + notification permission request | Carousel slide 4 + notification prompt |
| 16 | **AppsAfterAuthContinue** | Transition to main app (device frame) | Fade to main interface |

**Key Onboarding Patterns:**
- ✅ **Progressive profiling** - Username → Gmail → AI analysis → Personalized apps
- ✅ **Visual consistency** - Gradient backgrounds throughout
- ✅ **Dot navigation** - 4-dot progress indicator
- ✅ **Swipe navigation** - Card-based progression
- ✅ **Permission requests** - Notifications asked at end

**Desktop Adaptations:**
- Replace swipe gestures with next/previous buttons or arrow keys
- Expand carousel to show multiple apps side-by-side
- Add "Skip" option for power users
- Store onboarding state for multi-session completion

---

#### **PHASE 2: MAIN APP - HOME & NAVIGATION (Screens 17-19)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 17 | **Home** | Grid of user's mini-apps with circular icons | Main dashboard with app grid |
| 18 | **Explore** | Featured apps, invites, popular this week | Explore sidebar/panel |
| 19 | **Profile** | User profile with public/private tabs | Profile window/modal |

**Home Screen Layout:**
- **Header:** "Home" title + notification bell + profile avatar
- **App Grid:** 2-column grid of circular app icons with titles
- **App States:** Usage percentage indicator (e.g., "98%")
- **Bottom Nav:** Home, Explore, Search, Create (+)

**Desktop Home Layout:**
```
┌─────────────────────────────────────────────────────────┐
│  OSA Build                    🔔  🔍  ➕  👤            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  MY APPS                                    [Grid View] │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐              │
│  │  🏙️   │  │  🔥   │  │  📅   │  │  📚   │              │
│  │ Wabi  │  │Feed- │  │ SF    │  │Book  │              │
│  │ block │  │back  │  │Found. │  │Finds │              │
│  │ 98%   │  │ hub  │  │Wkend  │  │      │              │
│  └──────┘  └──────┘  └──────┘  └──────┘              │
│  ┌──────┐  ┌──────┐                                   │
│  │  🎨   │  │  ⏱️   │                                   │
│  │Motion│  │Softw.│                                   │
│  │Design│  │count.│                                   │
│  └──────┘  └──────┘                                   │
│                                                         │
│  EXPLORE                                   [See All >] │
│  Featured Today                                        │
│  ┌──────────────┐  ┌──────────────┐                  │
│  │   MenuIQ     │  │   PANTUNE    │                  │
│  │ by @jivrao   │  │ by @chikapul │                  │
│  └──────────────┘  └──────────────┘                  │
└─────────────────────────────────────────────────────────┘
```

**Desktop Enhancements:**
- Resizable grid (2, 3, 4 columns)
- List view alternative
- App search/filter
- Quick actions on hover (Open, Edit, Share, Delete)
- Keyboard navigation (Tab, Enter, Arrow keys)

---

#### **PHASE 3: PROFILE & SOCIAL (Screens 19-20)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 19 | **Profile** | Own profile (@bekorains) - Public/Private tabs | Profile settings panel |
| 20 | **PublicProfileView** | Other user's profile (@blas) with apps | User profile modal |

**Profile Features:**
- Avatar + edit button
- Username + handle
- Follower/Following counts
- Public/Private app tabs
- Settings/Share buttons
- App grid display

**Desktop Profile:**
- Full-width header with cover image (optional)
- Sidebar with user info
- Tabbed interface (Public Apps, Private Apps, Activity, Settings)
- Follow button with hover state
- Share profile link

---

#### **PHASE 4: APP INTERACTION - PUBLIC VIEW (Screens 21-26)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 21 | **PublicAppView** | App detail header (creator, title, stats, actions) | App detail modal/window |
| 22 | **PublicAppFullView** | App in fullscreen mode | Fullscreen app view |
| 23 | **PublicCommentsView** | Comments section (12 Likes, 9 Comments, 160 Remixes) | Comments sidebar/panel |
| 24 | **PublicLikesView** | List of users who liked the app | Likes modal |
| 25 | **PublicShareApp** | Share modal with app icon and message | Share dialog |
| 26 | **RemixingLoad** | "Remixing..." loading state | Loading overlay |

**App Detail Layout (iOS):**
```
┌─────────────────────────────────┐
│  ← 🔔   @tykra              ⋯  │
│                                 │
│        [App Display Area]       │
│                                 │
│  Wabi block                     │
│  Describe your city...          │
│                                 │
│  ♡ 12   💬 9   🔀 160   ⬇ 496  │
│                                 │
│  [🔀 Remix]     [⬇ Get]         │
└─────────────────────────────────┘
```

**Desktop App Detail:**
```
┌─────────────────────────────────────────────────────────┐
│  ←  Wabi block by @tykra                      ⋯  ✕     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────────┐  ┌─────────────────────┐ │
│  │                         │  │  💬 Comments (9)     │ │
│  │   [App Display Area]    │  │  ──────────────────  │ │
│  │                         │  │  @alisonyyh          │ │
│  │                         │  │  Remixing to fix...  │ │
│  │                         │  │                      │ │
│  │                         │  │  @tempic             │ │
│  │                         │  │  Play my game        │ │
│  │                         │  │                      │ │
│  └─────────────────────────┘  │  [Say something...]  │ │
│                                └─────────────────────┘ │
│  Wabi block                                            │
│  Describe your city and step into the Wabi Block...   │
│                                                         │
│  ♡ 12 Likes   💬 9 Comments   🔀 160 Remixes   ⬇ 496  │
│                                                         │
│  [🔀 Remix]  [⬇ Get]  [🔗 Share]                       │
└─────────────────────────────────────────────────────────┘
```

**Comments Feature:**
- Real-time comment feed
- Like individual comments
- Reply threading
- Comment timestamps (8d, 10d, 12d, 29d, 30d)
- Voice input support
- "Say something nice" placeholder

**Social Actions:**
- **Remix** - Create editable copy of app
- **Get** - Add to personal collection
- **Share** - Share with friends (no data shared)
- **Like** - Show appreciation
- **Comment** - Leave feedback

---

#### **PHASE 5: APP BUILDER INTERFACE (Screens 27-45)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 27 | **BuildChatView** | Chat tab - Conversational AI interface | Chat panel in builder |
| 28 | **BuildGeneralView** | General tab - Title + Description + Clear data | Settings panel |
| 29 | **Edit TitleView** | Edit title modal with character limit | Inline title edit |
| 30 | **ClearAppData** | Confirm clear app data action | Confirmation dialog |
| 31 | **BuildIconView** | Icon tab - Current icon + Regenerate + Prompt | Icon editor panel |
| 32 | **EditIconPrompt** | Edit icon generation prompt | Icon prompt editor |
| 33 | **BuildPromptsView** | Prompts tab - System prompts + Model selection | Prompts configuration |
| 34 | **ImageModelsOptions** | Image model selection (Nano-Banana, etc.) | Model dropdown |
| 35 | **OthersImageModelsOptions** | Other image model options | Extended model list |
| 36 | **ChatModelsOptions** | Chat model selection (GPT-5.2, Gemini 3, Claude 4.5) | Chat model dropdown |
| 37 | **ChatModelsOtherOptions** | Other chat model options | Extended model list |
| 38 | **BuildMainPromptEdit** | Edit main system prompt | Prompt editor |
| 39 | **PromptExample1** | Example prompt text (.md file) | Markdown viewer |
| 40 | **PromptExample1Image** | Example image prompt (.md file) | Markdown viewer |
| 41 | **IconRegenLoad** | Icon regeneration loading state | Loading spinner |
| 42 | **IconRegenAfterLoad** | New icon after regeneration | Updated icon display |
| 43 | **ChatPlusButton** | Add new chat message (+) | Message input |
| 44 | **Chat1Message** | First chat interaction in builder | Chat message |
| 45 | **BuildChatRequest1** | User request in chat | Chat request bubble |

**Builder Tab Structure:**

```
┌─────────────────────────────────────────────────────────┐
│  ○ Chat    ≡ General    ◎ Icon    🎨 Prompts           │
├─────────────────────────────────────────────────────────┤
│  [Tab Content]                                          │
└─────────────────────────────────────────────────────────┘
```

**1. Chat Tab - Conversational Builder**
```
Welcome to Wabi Block! 🎉 This app is now yours—your
cozy corner where your city comes alive in a playful
neighborhood shaped by your imagination.

Tell me what to tweak, and I'll do my best to make it
perfect! Or, dive into settings to get creative with
prompts, the cover, or integrations. Let's build something
amazing!

─────────────────────────────────────────────────────────

What can we work on here today

We can enhance or fix your existing app, "Wabi block," by
adding features, improving design, or resolving issues.
Let me know what you'd like to adjust!

─────────────────────────────────────────────────────────

[+ Ask for any changes]                              [🎤]
```

**2. General Tab - App Settings**
```
Title
Wabi block                                            >

Description
Describe your city and step into the Wabi Block,      >
a cozy neighborhood shaped by your world.

                  Clear all app data
```

**3. Icon Tab - Icon Management**
```
Current icon
    ┌─────────┐
    │   🏙️    │
    │         │
    └─────────┘
      [Regenerate]

Prompt                                            Edit >
ULTRA-PHOTOREALISTIC COZY-WARM
ISOMETRIC CITY BLOCK

(FILM MACRO / HUMAN WARMTH REALISM)

A cinematic studio macro photograph, indistingu...
```

**4. Prompts Tab - Model Configuration**
```
🎯 City Title

Prompt                                            Edit >
Render a hyper-realistic isometric 3D city tile of
[CITY], designed as a physically accurate miniature
urban environment, comparable to a cinema-grade
architectural visualization.

Models

≡  City Text                                      Auto ▼

📷 Hyper-Realistic                   Nano-Banana P... ▼
   City Image
```

**Model Selection Options:**
- **Chat Models:**
  - ✓ Auto (We'll pick the best model for your task)
  - OpenAI GPT-5.2 Chat
  - Gemini 3 Pro
  - Claude Sonnet 4.5
  - Other models >

- **Image Models:**
  - Nano-Banana Pro
  - Other image models >

---

#### **PHASE 6: APP USAGE & ITERATION (Screens 46-50+)**

| # | Screen Name | Purpose | Desktop Equivalent |
|---|-------------|---------|-------------------|
| 46 | **BuildChatEditRequest** | Edit previous chat request | Edit message |
| 47 | **BuildChatRequestLoad** | Loading AI response | Loading indicator |
| 48 | **HomeAfterAppUsage** | Home with usage stats (98% on Wabi block) | Updated home view |
| 49 | **EdittingFirstLoad** | First edit loading state | Loading overlay |
| 50 | **EdittingComplete** | Edit completed successfully | Success state |
| 51 | **AppViewTools** | App tools/settings menu | Tools menu |

**Usage Tracking:**
- Apps show usage percentage on home screen
- Most-used apps highlighted
- Activity tracking for engagement

**Iteration Loop:**
1. Use app
2. Notice improvement needed
3. Open builder chat
4. Request change
5. AI processes request
6. App updates
7. Return to usage

---

## 3. Feature Analysis & Breakdown

### 🎯 Core Features (Must-Have for Desktop)

#### 3.1 Personalized Onboarding
**iOS Implementation:**
- Gmail OAuth → Email analysis → Interest extraction
- AI generates 4 personalized mini-apps
- Progressive reveal with swipe navigation

**Desktop Implementation:**
- Email OAuth (Gmail, Outlook, Apple)
- Background analysis with progress indicator
- Grid display of generated apps with "Add to Dashboard" buttons
- Skip option for manual app creation

**Technical Requirements:**
- Email API integration (Gmail API, Microsoft Graph)
- Email parsing and NLP for interest extraction
- App template system with variable injection
- Personalization engine (likely OpenAI/Anthropic)

---

#### 3.2 App Builder - Conversational Interface
**iOS Implementation:**
- 4-tab interface: Chat, General, Icon, Prompts
- Natural language requests in Chat tab
- Real-time AI responses
- Model selection per component (text, image)

**Desktop Implementation:**
```
┌─────────────────────────────────────────────────────────┐
│  [App Name] - Builder                              ✕    │
├──────────┬──────────────────────────────────────────────┤
│  Chat    │  Welcome to [App Name]! 🎉                   │
│  General │                                              │
│  Icon    │  Tell me what to tweak...                    │
│  Prompts │                                              │
│          │  ─────────────────────────────────────────── │
│          │                                              │
│          │  > Add a dark mode toggle                    │
│          │                                              │
│          │  I'll add a dark mode toggle for you! This   │
│          │  will allow users to switch between light    │
│          │  and dark themes. Give me a moment...        │
│          │                                              │
│          │  ✓ Dark mode toggle added to settings       │
│          │                                              │
│          │  ─────────────────────────────────────────── │
│          │                                              │
│          │  [Type your request...]              [Send]  │
└──────────┴──────────────────────────────────────────────┘
```

**Key Interactions:**
- Conversational requests with context awareness
- Streaming responses (SSE/WebSocket)
- Action confirmation before applying changes
- Undo/Redo support for iterations
- Change preview before commit

---

#### 3.3 Multi-Model Architecture
**iOS Implementation:**
- Separate model selection for different app components:
  - **Chat/Text generation:** GPT-5.2, Gemini 3 Pro, Claude Sonnet 4.5
  - **Image generation:** Nano-Banana Pro, other models
- "Auto" mode for intelligent model selection

**Desktop Implementation:**
- Model configuration panel in Prompts tab
- Global model preferences + per-app overrides
- Cost estimation for model usage
- Fallback model chains
- Local model support (Ollama integration)

**Architecture Pattern:**
```typescript
interface AppComponent {
  id: string;
  name: string;
  type: 'text' | 'image' | 'code' | 'data';
  model: ModelConfig;
  prompt: string;
}

interface ModelConfig {
  provider: 'openai' | 'anthropic' | 'google' | 'local' | 'auto';
  model: string;
  temperature?: number;
  maxTokens?: number;
}

interface MiniApp {
  id: string;
  title: string;
  description: string;
  icon: {
    url: string;
    prompt: string;
    model: ModelConfig;
  };
  components: AppComponent[];
  visibility: 'public' | 'private';
  stats: {
    likes: number;
    comments: number;
    remixes: number;
    gets: number;
  };
}
```

---

#### 3.4 Social Features
**iOS Implementation:**
- Explore feed with featured apps
- User profiles (followers, following)
- App interactions: Like, Comment, Remix, Get, Share
- Comments with timestamps and threading
- Remix counter showing adoption

**Desktop Implementation:**
- Explore sidebar/panel with infinite scroll
- Profile modals with full app history
- Inline commenting with rich text
- Share to external platforms (Twitter, Discord, Email)
- Remix with attribution tracking
- Activity feed showing followers' actions

**Social Graph:**
```typescript
interface User {
  id: string;
  username: string;
  handle: string;
  avatar: string;
  bio?: string;
  stats: {
    followers: number;
    following: number;
    appsCreated: number;
    appsRemixed: number;
  };
}

interface AppInteraction {
  userId: string;
  appId: string;
  type: 'like' | 'comment' | 'remix' | 'get' | 'share';
  timestamp: Date;
  metadata?: {
    commentText?: string;
    remixId?: string;
    shareTarget?: string;
  };
}
```

---

#### 3.5 App Discovery & Organization
**iOS Implementation:**
- Grid view with circular icons
- Usage percentage indicator
- Categories: My Apps, Featured, Popular, Recent
- Search and filter

**Desktop Implementation:**
- Multiple view modes:
  - **Grid** (2x2, 3x3, 4x4)
  - **List** (detailed info)
  - **Timeline** (chronological)
- Smart folders:
  - Recently Used
  - Most Popular
  - Remixed from Others
  - Shared with Me
- Tagging system
- Advanced search (by creator, category, date, usage)

---

### 🚀 Desktop-Specific Enhancements (Nice-to-Have)

#### 3.6 Window Management
- Open multiple apps simultaneously
- Minimize to tray
- Snap to edges
- Picture-in-picture mode for select apps

#### 3.7 Keyboard Shortcuts
```
Cmd/Ctrl + N       - Create new app
Cmd/Ctrl + O       - Open app
Cmd/Ctrl + E       - Open builder for current app
Cmd/Ctrl + /       - Open search
Cmd/Ctrl + ,       - Open settings
Cmd/Ctrl + 1-9     - Quick access to first 9 apps
Cmd/Ctrl + Shift+N - New app from template
Cmd/Ctrl + D       - Duplicate current app
```

#### 3.8 File System Integration
- Export app as standalone HTML/PWA
- Import app from file
- Auto-save with version history
- Backup to cloud storage

#### 3.9 Developer Mode
- View generated code
- Edit code directly
- Custom CSS/JS injection
- API endpoint configuration
- Webhooks for integrations

---

## 4. Desktop Architecture Design

### 🏗️ Technology Stack

**Frontend:**
- **Framework:** SvelteKit (existing)
- **UI Components:** Tailwind CSS + shadcn-svelte
- **State Management:** Svelte stores + Context API
- **3D/Animations:** Three.js (for Desktop3D integration)
- **Real-time:** SSE for AI streaming responses

**Backend:**
- **Framework:** Go 1.24+ (existing)
- **API:** REST + SSE
- **Database:** PostgreSQL (user data, apps, social graph)
- **Vector DB:** pgvector (for app search, recommendations)
- **Cache:** Redis (session, real-time data)
- **Queue:** Redis + Go routines (for async AI tasks)

**AI/ML:**
- **LLM Routing:** OpenRouter or custom multi-provider layer
- **Models:** OpenAI GPT-5.2, Anthropic Claude 4.5, Google Gemini 3 Pro
- **Image Generation:** DALL-E 3, Midjourney API, Stable Diffusion
- **Embeddings:** OpenAI text-embedding-3 (for search)

**Desktop Integration:**
- **Electron:** Window management, system tray, file system
- **IPC:** Main ↔ Renderer communication
- **Auto-update:** electron-updater
- **Analytics:** Posthog (existing)

---

### 📐 System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         DESKTOP LAYER                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Electron Main Process                                    │  │
│  │  - Window management                                      │  │
│  │  - System tray                                            │  │
│  │  - File system access                                     │  │
│  │  - Auto-update                                            │  │
│  └──────────────────┬───────────────────────────────────────┘  │
│                     │ IPC                                       │
│  ┌──────────────────▼───────────────────────────────────────┐  │
│  │  Electron Renderer Process (SvelteKit App)               │  │
│  │                                                           │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │  │
│  │  │   Home      │  │   Explore   │  │   Builder   │     │  │
│  │  │   View      │  │   Feed      │  │   Interface │     │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘     │  │
│  │                                                           │  │
│  │  ┌───────────────────────────────────────────────────┐  │  │
│  │  │  Svelte Stores (State Management)                 │  │  │
│  │  │  - userStore                                      │  │  │
│  │  │  - appsStore                                      │  │  │
│  │  │  - builderStore                                   │  │  │
│  │  │  - socialStore                                    │  │  │
│  │  └───────────────────────────────────────────────────┘  │  │
│  └──────────────────┬───────────────────────────────────────┘  │
└─────────────────────┼───────────────────────────────────────────┘
                      │ HTTP/SSE
┌─────────────────────▼───────────────────────────────────────────┐
│                      BACKEND LAYER (Go)                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  API Gateway (Gin Router)                                │  │
│  │  - Authentication middleware (JWT)                       │  │
│  │  - Rate limiting                                         │  │
│  │  - Request logging (slog)                                │  │
│  └──────────────────┬───────────────────────────────────────┘  │
│                     │                                           │
│  ┌──────────────────▼───────────────────────────────────────┐  │
│  │  Handler Layer                                           │  │
│  │  - /api/auth/*        (OAuth, session)                   │  │
│  │  - /api/apps/*        (CRUD, search)                     │  │
│  │  - /api/builder/*     (AI requests, SSE)                 │  │
│  │  - /api/social/*      (follow, like, comment)            │  │
│  │  - /api/explore/*     (feed, featured)                   │  │
│  └──────────────────┬───────────────────────────────────────┘  │
│                     │                                           │
│  ┌──────────────────▼───────────────────────────────────────┐  │
│  │  Service Layer                                           │  │
│  │  - AuthService                                           │  │
│  │  - AppService                                            │  │
│  │  - BuilderService (AI orchestration)                     │  │
│  │  - SocialService                                         │  │
│  │  - PersonalizationService                                │  │
│  └──────────────────┬───────────────────────────────────────┘  │
│                     │                                           │
│  ┌──────────────────▼───────────────────────────────────────┐  │
│  │  Repository Layer                                        │  │
│  │  - UserRepository                                        │  │
│  │  - AppRepository                                         │  │
│  │  - InteractionRepository                                 │  │
│  └──────────────────┬───────────────────────────────────────┘  │
└─────────────────────┼───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                      DATA LAYER                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ PostgreSQL   │  │   pgvector   │  │    Redis     │          │
│  │              │  │              │  │              │          │
│  │ - users      │  │ - embeddings │  │ - sessions   │          │
│  │ - apps       │  │ - search     │  │ - cache      │          │
│  │ - comments   │  │              │  │ - queues     │          │
│  │ - follows    │  │              │  │              │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      EXTERNAL SERVICES                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   OpenAI     │  │  Anthropic   │  │   Google     │          │
│  │   GPT-5.2    │  │  Claude 4.5  │  │  Gemini 3    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Gmail API   │  │  Posthog     │  │   LiveKit    │          │
│  │  (OAuth)     │  │  (Analytics) │  │   (Voice)    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────────┘
```

---

### 🗄️ Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    handle VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    avatar_url TEXT,
    bio TEXT,
    gmail_connected BOOLEAN DEFAULT false,
    gmail_access_token TEXT,
    gmail_refresh_token TEXT,
    interests JSONB, -- AI-extracted interests
    onboarding_completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Mini-apps table
CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    icon_prompt TEXT,
    visibility VARCHAR(20) DEFAULT 'private', -- 'public' | 'private'
    components JSONB, -- Array of app components with prompts/models
    usage_percentage INTEGER DEFAULT 0,
    remixed_from UUID REFERENCES apps(id), -- Track remix source
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

-- Indexes
CREATE INDEX idx_apps_user_id ON apps(user_id);
CREATE INDEX idx_apps_visibility ON apps(visibility);
CREATE INDEX idx_interactions_app_id ON interactions(app_id);
CREATE INDEX idx_interactions_user_id ON interactions(user_id);
CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);

-- pgvector index for similarity search
CREATE INDEX idx_app_embeddings_vector ON app_embeddings
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

---

### 🔄 API Endpoints

#### Authentication
```
POST   /api/auth/google          - OAuth with Google
POST   /api/auth/logout          - Logout user
GET    /api/auth/me              - Get current user
```

#### Apps
```
GET    /api/apps                 - List user's apps
GET    /api/apps/:id             - Get app details
POST   /api/apps                 - Create new app
PATCH  /api/apps/:id             - Update app
DELETE /api/apps/:id             - Delete app
GET    /api/apps/:id/components  - Get app components
POST   /api/apps/:id/remix       - Remix app (create copy)
```

#### Builder (AI Interface)
```
POST   /api/builder/chat         - Send chat message (SSE response)
POST   /api/builder/icon         - Regenerate icon
POST   /api/builder/apply        - Apply changes to app
GET    /api/builder/models       - List available models
```

#### Social
```
GET    /api/users/:handle        - Get user profile
POST   /api/users/:id/follow     - Follow user
DELETE /api/users/:id/follow     - Unfollow user
GET    /api/apps/:id/likes       - Get app likes
POST   /api/apps/:id/like        - Like app
DELETE /api/apps/:id/like        - Unlike app
GET    /api/apps/:id/comments    - Get app comments
POST   /api/apps/:id/comments    - Add comment
POST   /api/apps/:id/share       - Share app
```

#### Explore
```
GET    /api/explore/featured     - Get featured apps
GET    /api/explore/popular      - Get popular apps
GET    /api/explore/recent       - Get recently added apps
GET    /api/explore/search       - Search apps (query param)
```

#### Personalization
```
POST   /api/onboarding/analyze   - Analyze Gmail for interests
GET    /api/onboarding/generate  - Generate personalized apps
```

---

## 5. UI/UX Pattern Translation

### 🎨 Visual Design System

#### Color Palette (from iOS screenshots)
```css
/* Gradient backgrounds */
--gradient-onboarding-1: linear-gradient(180deg, #E8F4F8 0%, #F5E8F8 100%);
--gradient-onboarding-2: linear-gradient(180deg, #D4F1F9 0%, #F0D9F5 100%);
--gradient-onboarding-3: linear-gradient(180deg, #E3E8F8 0%, #F5EDE8 100%);

/* App background */
--bg-primary: #F8F9FA;
--bg-secondary: #FFFFFF;

/* Text */
--text-primary: #1A1A1A;
--text-secondary: #666666;
--text-tertiary: #999999;

/* Accent */
--accent-primary: #FF6B35; /* Feature feedback hub */
--accent-blue: #4A90E2;
--accent-green: #7ED321;
```

#### Typography
```css
/* Headers */
--font-display: 'SF Pro Display', -apple-system, sans-serif;
--font-size-hero: 32px;
--font-size-h1: 24px;
--font-size-h2: 20px;

/* Body */
--font-body: 'SF Pro Text', -apple-system, sans-serif;
--font-size-body: 16px;
--font-size-small: 14px;
--font-size-tiny: 12px;
```

#### Spacing & Layout
```css
--spacing-xs: 8px;
--spacing-sm: 12px;
--spacing-md: 16px;
--spacing-lg: 24px;
--spacing-xl: 32px;
--spacing-2xl: 48px;

--border-radius-sm: 8px;
--border-radius-md: 12px;
--border-radius-lg: 16px;
--border-radius-xl: 24px;
--border-radius-full: 9999px; /* For circular icons */
```

---

### 🔄 Interaction Patterns

#### 1. **Swipe → Click/Keyboard**
**iOS:** Swipe between onboarding cards
**Desktop:** Next/Previous buttons + Arrow keys

#### 2. **Long Press → Right-Click**
**iOS:** Long-press on app for context menu
**Desktop:** Right-click for context menu

#### 3. **Pull to Refresh → Refresh Button**
**iOS:** Pull down on feed
**Desktop:** Refresh icon in header

#### 4. **Bottom Sheet → Modal/Sidebar**
**iOS:** Bottom sheet for app details
**Desktop:** Modal window or side panel

#### 5. **Haptic Feedback → Visual Feedback**
**iOS:** Haptic on button press
**Desktop:** Button press animation, ripple effect

---

### 📱 Responsive Breakpoints

```css
/* Desktop-first approach (since it's a desktop app) */
--breakpoint-desktop-large: 1920px;
--breakpoint-desktop: 1440px;
--breakpoint-laptop: 1024px;
--breakpoint-tablet: 768px; /* Minimum supported */
```

**Grid Adaptation:**
- **1920px+:** 4-column app grid
- **1440px:** 3-column app grid
- **1024px:** 2-column app grid
- **768px:** 2-column app grid (minimum)

---

## 6. Technical Implementation Roadmap

### 🗓️ Phase 1: Foundation (Weeks 1-2)

**Goal:** Set up core infrastructure and data models

#### Backend Tasks
- [ ] Create database schema (users, apps, interactions, follows)
- [ ] Set up pgvector extension for embeddings
- [ ] Implement JWT authentication
- [ ] Create base API structure (handlers, services, repositories)
- [ ] Set up Gmail OAuth flow
- [ ] Configure multi-provider AI routing (OpenRouter)

#### Frontend Tasks
- [ ] Create base layout with navigation
- [ ] Set up Svelte stores (userStore, appsStore, builderStore)
- [ ] Implement authentication flow
- [ ] Design component library (buttons, cards, modals)
- [ ] Set up Electron wrapper

**Deliverable:** Login works, empty dashboard loads

---

### 🗓️ Phase 2: Onboarding (Weeks 3-4)

**Goal:** Implement personalized onboarding flow

#### Backend Tasks
- [ ] Gmail API integration for email fetching
- [ ] Email parsing and interest extraction (NLP)
- [ ] App generation from templates
- [ ] Personalization algorithm

#### Frontend Tasks
- [ ] Invite code validation screen
- [ ] OAuth connection screen
- [ ] Username claim with availability check
- [ ] AI analysis loading states (3 messages)
- [ ] Personalized app carousel
- [ ] Notification permission prompt
- [ ] Transition to main app

**Deliverable:** New user can complete onboarding and see 4 personalized apps

---

### 🗓️ Phase 3: Home & App Management (Weeks 5-6)

**Goal:** Build home screen and basic app operations

#### Backend Tasks
- [ ] App CRUD endpoints
- [ ] App search with pgvector
- [ ] Usage tracking
- [ ] App visibility (public/private)

#### Frontend Tasks
- [ ] Home grid view with app icons
- [ ] App usage percentage display
- [ ] App card hover effects
- [ ] Quick actions (Open, Edit, Delete)
- [ ] App search/filter
- [ ] Empty state for new users

**Deliverable:** Users can view, search, and delete apps

---

### 🗓️ Phase 4: Builder Interface (Weeks 7-9)

**Goal:** Implement conversational AI builder

#### Backend Tasks
- [ ] Builder chat endpoint with SSE streaming
- [ ] Multi-model routing (GPT, Claude, Gemini)
- [ ] Icon generation (DALL-E integration)
- [ ] Prompt management
- [ ] Change preview system
- [ ] Apply changes to app

#### Frontend Tasks
- [ ] 4-tab builder interface (Chat, General, Icon, Prompts)
- [ ] Chat UI with message streaming
- [ ] General settings (title, description, clear data)
- [ ] Icon viewer with regenerate
- [ ] Prompts editor with model selection
- [ ] Model dropdown components
- [ ] Loading states for AI operations

**Deliverable:** Users can create and modify apps through chat interface

---

### 🗓️ Phase 5: Social Features (Weeks 10-12)

**Goal:** Add explore, profiles, and social interactions

#### Backend Tasks
- [ ] User profile endpoints
- [ ] Follow/unfollow system
- [ ] Explore feed algorithm (featured, popular, recent)
- [ ] Like/comment/remix/get/share endpoints
- [ ] Comment threading
- [ ] Notification system

#### Frontend Tasks
- [ ] Explore sidebar with infinite scroll
- [ ] User profile modal
- [ ] Public app detail view
- [ ] Comments panel with real-time updates
- [ ] Like/remix/share buttons
- [ ] Share modal
- [ ] Remix loading state
- [ ] Activity feed

**Deliverable:** Users can discover, interact with, and remix apps from others

---

### 🗓️ Phase 6: Desktop Enhancements (Weeks 13-14)

**Goal:** Add desktop-specific features

#### Tasks
- [ ] Window management (minimize, maximize, snap)
- [ ] System tray integration
- [ ] Keyboard shortcuts
- [ ] File system integration (export, import)
- [ ] Multi-window support (open multiple apps)
- [ ] Auto-update system
- [ ] Settings panel (preferences, model config, API keys)

**Deliverable:** Full-featured desktop app with native OS integration

---

### 🗓️ Phase 7: Polish & Launch (Weeks 15-16)

**Goal:** Finalize and prepare for beta launch

#### Tasks
- [ ] Performance optimization (lazy loading, caching)
- [ ] Error handling and logging
- [ ] Onboarding tooltips
- [ ] Help documentation
- [ ] Analytics integration (Posthog)
- [ ] Beta tester feedback loop
- [ ] Bug fixes
- [ ] App store preparation (if applicable)

**Deliverable:** Production-ready desktop app for beta launch

---

## 7. Data Models & API Contracts

### 📦 TypeScript Interfaces (Frontend)

```typescript
// User
interface User {
  id: string;
  username: string;
  handle: string; // @username
  email: string;
  avatar: string;
  bio?: string;
  gmailConnected: boolean;
  interests?: string[];
  onboardingCompleted: boolean;
  stats: {
    followers: number;
    following: number;
    appsCreated: number;
    appsRemixed: number;
  };
  createdAt: Date;
  updatedAt: Date;
}

// Mini-App
interface MiniApp {
  id: string;
  userId: string;
  title: string;
  description: string;
  icon: {
    url: string;
    prompt: string;
  };
  visibility: 'public' | 'private';
  components: AppComponent[];
  usagePercentage: number;
  remixedFrom?: string; // App ID if remixed
  stats: {
    likes: number;
    comments: number;
    remixes: number;
    gets: number;
  };
  createdAt: Date;
  updatedAt: Date;
}

// App Component
interface AppComponent {
  id: string;
  name: string;
  type: 'text' | 'image' | 'code' | 'data';
  prompt: string;
  model: {
    provider: 'openai' | 'anthropic' | 'google' | 'auto';
    name: string;
    config?: {
      temperature?: number;
      maxTokens?: number;
      topP?: number;
    };
  };
}

// Interaction
interface Interaction {
  id: string;
  userId: string;
  appId: string;
  type: 'like' | 'comment' | 'remix' | 'get' | 'share';
  commentText?: string;
  parentCommentId?: string; // For threading
  createdAt: Date;
}

// Comment (enriched interaction)
interface Comment extends Interaction {
  type: 'comment';
  user: {
    handle: string;
    avatar: string;
  };
  replies?: Comment[];
  likeCount: number;
}

// Builder Chat Message
interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  metadata?: {
    actionTaken?: string;
    changePreview?: any;
  };
}

// Model Option
interface ModelOption {
  id: string;
  provider: string;
  name: string;
  displayName: string;
  description?: string;
  capabilities: ('text' | 'image' | 'code')[];
  costPer1kTokens?: number;
}

// Explore Feed Item
interface FeedItem {
  type: 'featured' | 'popular' | 'recent' | 'following';
  app: MiniApp;
  creator: User;
  context?: string; // e.g., "3 people you follow remixed this"
}
```

---

### 🔌 API Request/Response Examples

#### POST /api/auth/google
**Request:**
```json
{
  "code": "4/0AeanEI...",
  "redirectUri": "http://localhost:3000/auth/callback"
}
```

**Response:**
```json
{
  "user": {
    "id": "user_123",
    "username": "bekorains",
    "handle": "@bekorains",
    "email": "beko@example.com",
    "avatar": "https://...",
    "gmailConnected": true,
    "onboardingCompleted": false
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

#### POST /api/builder/chat (SSE)
**Request:**
```json
{
  "appId": "app_456",
  "message": "Add a dark mode toggle to settings"
}
```

**Response (Server-Sent Events):**
```
event: message
data: {"content": "I'll add a dark mode toggle for you! ", "done": false}

event: message
data: {"content": "This will allow users to switch between light and dark themes. ", "done": false}

event: message
data: {"content": "Give me a moment...", "done": false}

event: action
data: {"type": "preview", "changes": {"settings": {"darkMode": true}}}

event: message
data: {"content": "\n\n✓ Dark mode toggle added to settings", "done": true}

event: done
data: {"success": true}
```

---

#### GET /api/apps/:id
**Response:**
```json
{
  "id": "app_789",
  "userId": "user_123",
  "title": "Wabi block",
  "description": "Describe your city and step into the Wabi Block...",
  "icon": {
    "url": "https://storage.../icon.png",
    "prompt": "ULTRA-PHOTOREALISTIC COZY-WARM ISOMETRIC CITY BLOCK"
  },
  "visibility": "public",
  "components": [
    {
      "id": "comp_1",
      "name": "City Title",
      "type": "text",
      "prompt": "Render a hyper-realistic isometric 3D city tile...",
      "model": {
        "provider": "auto",
        "name": "auto"
      }
    },
    {
      "id": "comp_2",
      "name": "Hyper-Realistic City Image",
      "type": "image",
      "prompt": "Generate an isometric city block...",
      "model": {
        "provider": "openai",
        "name": "dall-e-3"
      }
    }
  ],
  "usagePercentage": 98,
  "stats": {
    "likes": 12,
    "comments": 9,
    "remixes": 160,
    "gets": 496
  },
  "createdAt": "2026-01-15T10:30:00Z",
  "updatedAt": "2026-01-18T14:22:00Z"
}
```

---

#### GET /api/explore/featured
**Response:**
```json
{
  "items": [
    {
      "type": "featured",
      "app": {
        "id": "app_menu",
        "title": "MenuIQ",
        "description": "Scan menus and get meal recommendations...",
        "icon": { "url": "https://..." },
        "stats": { "likes": 245, "remixes": 34 }
      },
      "creator": {
        "handle": "@jivrao",
        "avatar": "https://..."
      },
      "context": "Featured today"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

---

## 8. Next Steps

### ✅ Immediate Actions

1. **Review & Approve Architecture**
   - Stakeholder review of this document
   - Confirm technical stack choices
   - Validate UI/UX translations

2. **Set Up Development Environment**
   - Create feature branch: `feature/ios-desktop-flow-migration`
   - Set up database migrations
   - Configure AI provider API keys

3. **Design System**
   - Create Figma mockups for desktop layouts
   - Build component library in Storybook
   - Define animation/transition specs

4. **Sprint Planning**
   - Break down roadmap into 2-week sprints
   - Assign tasks to team members
   - Set up tracking in Linear/Jira

### 🎯 Success Metrics

**Technical:**
- [ ] All iOS features ported to desktop
- [ ] < 2s app load time
- [ ] < 500ms AI response start (SSE)
- [ ] 99.9% uptime for backend

**User Experience:**
- [ ] < 5 min onboarding time
- [ ] > 80% onboarding completion rate
- [ ] > 5 apps created per user (first week)
- [ ] > 50% daily active usage

**Business:**
- [ ] 1,000 beta users in first month
- [ ] 10,000 apps created
- [ ] 500 daily remixes
- [ ] NPS > 50

---

## 📝 Appendix

### A. iOS Screen Reference Index

All 50+ screens have been documented in this architecture. Reference the "Complete User Flow Mapping" section for detailed screen-by-screen analysis.

### B. Glossary

- **Mini-App:** User-created AI application within OSA Build
- **Remix:** Create an editable copy of someone else's public app
- **Get:** Add someone else's app to your collection (read-only)
- **Component:** Individual AI-powered element of an app (text, image, etc.)
- **Builder:** Conversational interface for creating/editing apps
- **SSE:** Server-Sent Events (for real-time AI streaming)

### C. References

- [iOS App Screenshots](/Users/rhl/Desktop/Fallen/TheOSApp/)
- [BusinessOS Backend Repo](~/Desktop/BusinessOS2/desktop/backend-go/)
- [BusinessOS Frontend Repo](~/Desktop/BusinessOS2/frontend/)
- [OpenRouter Docs](https://openrouter.ai/docs)
- [pgvector Docs](https://github.com/pgvector/pgvector)

---

**Document Version:** 1.0
**Last Updated:** January 18, 2026
**Author:** Claude (Architecture Analysis)
**Next Review:** February 1, 2026
