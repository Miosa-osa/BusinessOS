# OSA Build Onboarding API Documentation

## Overview

The OSA Build Onboarding API powers the personalized operating system creation flow in BusinessOS. It analyzes user data, generates insights, and orchestrates the creation of 4 personalized starter applications tailored to each user's interests and workflows.

### Purpose

The onboarding flow guides users through a conversational 13-step journey to:
1. Collect basic information (email, username)
2. Connect integrations (Gmail, Calendar)
3. Analyze user behavior and interests
4. Generate personalized insights
5. Create 4 customized starter apps
6. Launch into the personalized OS

### Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│ Onboarding Flow Stages                                   │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Stage 1: Signup & Integration (Steps 1-5)              │
│  ├─ Email collection                                     │
│  ├─ Username selection                                   │
│  ├─ Gmail connect                                        │
│  ├─ Calendar connect                                     │
│  └─ Email verification                                   │
│                                                           │
│  Stage 2: Analysis & Personalization (Steps 6-9)        │
│  ├─ POST /api/osa-onboarding/analyze                     │
│  │   └─ Analyzes email, Gmail, Calendar data             │
│  │   └─ Returns: insights, interests, tools_used         │
│  ├─ Display: "Analyzing your workspace..."              │
│  └─ Display: Personalized insights (3 messages)          │
│                                                           │
│  Stage 3: App Generation & Launch (Steps 10-13)         │
│  ├─ POST /api/osa-onboarding/generate-apps              │
│  │   └─ Creates 4 personalized starter apps              │
│  │   └─ Returns: apps with status 'generating' or 'ready'│
│  ├─ GET /api/osa-onboarding/apps-status                 │
│  │   └─ Polls for app generation completion              │
│  ├─ Display: "Building your starter apps..."            │
│  └─ Display: Ready-to-launch OS with 4 apps             │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### Authentication

All endpoints require:
- **Bearer token** in `Authorization` header
- Valid JWT from BusinessOS authentication
- User ID extracted from token and validated

**Protected**: All endpoints are protected by middleware that validates the session token.

---

## Endpoints

### 1. Analyze User Data

Analyzes user's email, Gmail data, and Calendar data to generate personalized insights about their interests, tools, and workflow patterns.

#### Request

```http
POST /api/osa-onboarding/analyze
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Schema:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address (used for analysis) |
| `gmail_connected` | boolean | Yes | Whether Gmail is connected and analyzed |
| `calendar_connected` | boolean | No | Whether Calendar data is available (default: false) |

**Example Request:**

```json
{
  "email": "alex@example.com",
  "gmail_connected": true,
  "calendar_connected": false
}
```

#### Response

**Success Response (200 OK):**

```json
{
  "analysis": {
    "insights": [
      "No-code builder energy, big time",
      "Design tools are your playground",
      "AI-curious, testing new platforms"
    ],
    "interests": [
      "productivity",
      "automation",
      "design"
    ],
    "tools_used": [
      "Figma",
      "Notion",
      "Gmail"
    ],
    "profile_summary": "A no-code builder who values productivity, automation, and design and uses Figma, Notion, and Gmail regularly",
    "raw_data": {
      "gmail_summary": "Analyzed recent emails"
    }
  }
}
```

**Response Body Schema:**

```typescript
interface AnalyzeUserResponse {
  analysis: {
    insights: string[];        // 3 conversational insights about user
    interests: string[];        // 3-5 key interests detected
    tools_used: string[];       // Tools the user actively uses
    profile_summary: string;    // Full text summary of user profile
    raw_data: Record<string, any>; // Debug data from analysis
  }
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Invalid request | Missing required fields or invalid format |
| `401` | Unauthorized | Missing or invalid authentication token |
| `500` | Internal error | Failed to analyze user data |

**Example Error Response:**

```json
{
  "error": "email is required"
}
```

#### Notes

- Analysis is performed based on email domain and (optionally) Gmail data
- If Gmail connection fails, analysis uses email-based heuristics as fallback
- Results are deterministic based on email for testing
- Actual AI analysis powered by configured AI provider (Anthropic, Groq, or Ollama)

---

### 2. Generate Starter Apps

Creates 4 personalized starter applications based on user analysis. Each app is tailored to the user's interests and workflow patterns.

#### Request

```http
POST /api/osa-onboarding/generate-apps
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Schema:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Target workspace for app creation |
| `analysis` | UserAnalysisResult | Yes | Output from `/analyze` endpoint |

**Example Request:**

```json
{
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "analysis": {
    "insights": [
      "No-code builder energy, big time",
      "Design tools are your playground",
      "AI-curious, testing new platforms"
    ],
    "interests": [
      "productivity",
      "automation",
      "design"
    ],
    "tools_used": [
      "Figma",
      "Notion",
      "Gmail"
    ],
    "profile_summary": "A no-code builder who values productivity, automation, and design and uses Figma, Notion, and Gmail regularly",
    "raw_data": {}
  }
}
```

#### Response

**Success Response (200 OK):**

```json
{
  "starter_apps": [
    {
      "id": "app-001",
      "title": "Productivity Tracker",
      "description": "Track and organize your productivity projects",
      "icon_emoji": "📚",
      "icon_url": "https://...",
      "reasoning": "Because you're interested in productivity",
      "category": "tracker",
      "status": "ready",
      "workflow_id": "wf-12345"
    },
    {
      "id": "app-002",
      "title": "Figma Companion",
      "description": "Quick access to your Figma workflows",
      "icon_emoji": "🎨",
      "icon_url": "https://...",
      "reasoning": "Because you use Figma frequently",
      "category": "companion",
      "status": "generating",
      "workflow_id": "wf-12346"
    },
    {
      "id": "app-003",
      "title": "Idea Inbox",
      "description": "Capture ideas and get feedback",
      "icon_emoji": "💡",
      "icon_url": "https://...",
      "reasoning": "For collecting thoughts and feedback",
      "category": "feedback",
      "status": "ready",
      "workflow_id": "wf-12347"
    },
    {
      "id": "app-004",
      "title": "Daily Focus",
      "description": "Plan your day, track what matters",
      "icon_emoji": "🎯",
      "icon_url": "https://...",
      "reasoning": "For staying focused on priorities",
      "category": "daily",
      "status": "ready",
      "workflow_id": "wf-12348"
    }
  ],
  "ready_to_launch": false
}
```

**Response Body Schema:**

```typescript
interface GenerateAppsResponse {
  starter_apps: StarterApp[];     // Array of 4 generated apps
  ready_to_launch: boolean;       // True when all apps status == 'ready'
}

interface StarterApp {
  id: string;                     // Unique app ID
  title: string;                  // App title
  description: string;            // Brief description
  icon_emoji: string;             // Emoji icon identifier
  icon_url: string;               // CDN URL to generated icon
  reasoning: string;              // Why this app was generated for the user
  category: string;               // App category (tracker, companion, feedback, daily)
  status: 'generating' | 'ready' | 'failed'; // Generation status
  workflow_id: string;            // OSA orchestrator workflow ID
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Invalid request | Missing workspace_id or analysis |
| `401` | Unauthorized | Invalid authentication |
| `500` | Internal error | Failed to generate apps |

#### Notes

- This endpoint both generates apps and saves the profile to database
- Generation is asynchronous; apps start in `generating` status
- Use `/apps-status` endpoint to poll for completion
- 4 apps are always generated with these categories:
  1. Interest-based tracker
  2. Tool-based companion
  3. Feedback/idea inbox
  4. Daily focus utility

---

### 3. Check App Generation Status

Polls the current status of starter app generation. Use this to track when apps transition from `generating` to `ready`.

#### Request

```http
GET /api/osa-onboarding/apps-status?workspace_id=<uuid>
Authorization: Bearer <token>
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Workspace ID to check status for |

**Example Request:**

```http
GET /api/osa-onboarding/apps-status?workspace_id=550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

#### Response

**Success Response (200 OK):**

```json
{
  "analysis": {
    "insights": [...],
    "interests": [...],
    "tools_used": [...],
    "profile_summary": "...",
    "raw_data": {}
  },
  "starter_apps": [
    {
      "id": "app-001",
      "title": "Productivity Tracker",
      "description": "Track and organize your productivity projects",
      "icon_emoji": "📚",
      "icon_url": "https://...",
      "reasoning": "Because you're interested in productivity",
      "category": "tracker",
      "status": "ready",
      "workflow_id": "wf-12345"
    },
    // ... 3 more apps
  ],
  "ready_to_launch": true
}
```

**Response Body Schema:**

```typescript
interface AppsStatusResponse {
  analysis: UserAnalysisResult;
  starter_apps: StarterApp[];
  ready_to_launch: boolean;  // True when all apps are 'ready'
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | workspace_id required | Missing workspace_id query parameter |
| `401` | Unauthorized | Invalid authentication |
| `404` | Onboarding profile not found | No profile exists for workspace |

#### Notes

- Safe to call repeatedly; no rate limiting
- Returns cached profile plus current status of each app
- `ready_to_launch` becomes true when all 4 apps have status `ready`
- Use exponential backoff for polling (start at 2s, max 10s)

---

### 4. Get Onboarding Profile

Retrieves the complete saved onboarding profile for a workspace, including analysis and starter apps.

#### Request

```http
GET /api/osa-onboarding/profile?workspace_id=<uuid>
Authorization: Bearer <token>
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Workspace ID to retrieve profile for |

**Example Request:**

```http
GET /api/osa-onboarding/profile?workspace_id=550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

#### Response

**Success Response (200 OK):**

```json
{
  "analysis": {
    "insights": [
      "No-code builder energy, big time",
      "Design tools are your playground",
      "AI-curious, testing new platforms"
    ],
    "interests": [
      "productivity",
      "automation",
      "design"
    ],
    "tools_used": [
      "Figma",
      "Notion",
      "Gmail"
    ],
    "profile_summary": "A no-code builder who values productivity, automation, and design and uses Figma, Notion, and Gmail regularly",
    "raw_data": {}
  },
  "starter_apps": [
    {
      "id": "app-001",
      "title": "Productivity Tracker",
      "description": "Track and organize your productivity projects",
      "icon_emoji": "📚",
      "icon_url": "https://...",
      "reasoning": "Because you're interested in productivity",
      "category": "tracker",
      "status": "ready",
      "workflow_id": "wf-12345"
    },
    // ... 3 more apps
  ]
}
```

**Response Body Schema:**

```typescript
interface GetProfileResponse {
  analysis: UserAnalysisResult;
  starter_apps: StarterApp[];
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | workspace_id required | Missing workspace_id query parameter |
| `401` | Unauthorized | Invalid authentication |
| `404` | Onboarding profile not found | No profile exists for workspace |

#### Notes

- Returns the complete saved profile from database
- Useful for retrieving profile after onboarding completes
- No status polling needed; this is a read-only endpoint

---

### 5. Check Username Availability

Validates if a username is available for registration.

#### Request

```http
GET /api/users/check-username/:username
Content-Type: application/json
```

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | string | Yes | Username to validate |

**Example Request:**

```http
GET /api/users/check-username/alex_builds
```

#### Response

**Success Response (200 OK):**

```json
{
  "available": true,
  "username": "alex_builds"
}
```

**Username Taken (200 OK):**

```json
{
  "available": false,
  "username": "alex_builds",
  "suggestion": "alex_builds_2024"
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Invalid username | Username fails validation (too short, invalid chars) |
| `500` | Internal error | Database error |

#### Notes

- No authentication required
- Username validation rules:
  - 3-30 characters
  - Alphanumeric and underscores only
  - Must start with letter
- Returns suggestions if username is taken

---

### 6. Update User Username

Updates the authenticated user's username.

#### Request

```http
PATCH /api/users/me/username
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Schema:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | New username for the user |

**Example Request:**

```json
{
  "username": "alex_builds"
}
```

#### Response

**Success Response (200 OK):**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "alex_builds",
  "email": "alex@example.com",
  "updated_at": "2024-01-18T10:30:00Z"
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Username already taken | Username is unavailable |
| `401` | Unauthorized | Invalid authentication |
| `422` | Invalid username | Username fails validation |
| `500` | Internal error | Database error |

**Example Error Response:**

```json
{
  "error": "username already taken",
  "suggestion": "alex_builds_2024"
}
```

#### Notes

- Requires valid JWT authentication
- User ID extracted from token
- Username must pass validation rules (see check-username endpoint)
- Returns updated user object on success

---

## Data Models

### UserAnalysisResult

Represents the AI analysis of a user's profile, interests, and workflow patterns.

```typescript
interface UserAnalysisResult {
  insights: string[];              // 3 conversational insights (e.g., "No-code builder energy")
  interests: string[];             // 3-5 key interests (e.g., ["productivity", "automation"])
  tools_used: string[];            // Tools actively used (e.g., ["Figma", "Notion"])
  profile_summary: string;         // Full text summary for context
  raw_data: Record<string, any>;   // Debug/analytics data
}
```

**Example:**

```json
{
  "insights": [
    "No-code builder energy, big time",
    "Design tools are your playground",
    "AI-curious, testing new platforms"
  ],
  "interests": [
    "productivity",
    "automation",
    "design"
  ],
  "tools_used": [
    "Figma",
    "Notion",
    "Gmail"
  ],
  "profile_summary": "A no-code builder who values productivity, automation, and design and uses Figma, Notion, and Gmail regularly",
  "raw_data": {
    "gmail_summary": "Analyzed recent emails"
  }
}
```

### StarterApp

Represents a single personalized starter application.

```typescript
interface StarterApp {
  id: string;                          // UUID for this app
  title: string;                       // App name (e.g., "Productivity Tracker")
  description: string;                 // Brief description
  icon_emoji: string;                  // Emoji identifier (e.g., "📚")
  icon_url: string;                    // CDN URL to generated icon
  reasoning: string;                   // Why this app was created for user
  category: string;                    // Type: "tracker" | "companion" | "feedback" | "daily"
  status: "generating" | "ready" | "failed";  // Current generation status
  workflow_id: string;                 // OSA orchestrator workflow ID
}
```

**Example:**

```json
{
  "id": "app-001",
  "title": "Productivity Tracker",
  "description": "Track and organize your productivity projects",
  "icon_emoji": "📚",
  "icon_url": "https://cdn.businessos.io/icons/productivity-tracker.png",
  "reasoning": "Because you're interested in productivity",
  "category": "tracker",
  "status": "ready",
  "workflow_id": "wf-12345"
}
```

### OnboardingAnalysisRequest

Request payload for analyzing user data.

```typescript
interface OnboardingAnalysisRequest {
  email: string;              // User's email address
  gmail_connected: boolean;   // Whether Gmail is connected
  calendar_connected?: boolean; // Whether Calendar is available (optional)
}
```

### OnboardingAnalysisResponse

Response from analysis endpoint.

```typescript
interface OnboardingAnalysisResponse {
  analysis: UserAnalysisResult;
}
```

### OnboardingProfile

Represents complete onboarding state saved in database.

```typescript
interface OnboardingProfile {
  workspace_id: string;              // Workspace UUID
  user_id: string;                   // User UUID
  analysis_data: UserAnalysisResult; // Analysis result
  starter_apps_data: StarterApp[];   // Generated apps
  onboarding_method: "osa_build";    // Always "osa_build"
  created_at: string;                // ISO-8601 timestamp
  updated_at: string;                // ISO-8601 timestamp
}
```

---

## Integration Guide

### Frontend Integration

#### 1. Setup API Client

```typescript
import {
  analyzeUser,
  generateStarterApps,
  checkAppsStatus,
  getProfile,
  type UserAnalysisResult,
  type StarterApp
} from '$lib/api/osa-onboarding';
```

#### 2. Usage Example: Full Onboarding Flow

```typescript
import { onboardingStore } from '$lib/stores/onboardingStore';

// Step 1: Collect user data on Screen 5
async function submitEmail(email: string, gmailConnected: boolean) {
  onboardingStore.setUserData({ email, gmailConnected });
  onboardingStore.nextStep(); // Move to analysis screen
}

// Step 2: Analyze user (Screen 6-7)
async function startAnalysis() {
  const data = get(onboardingStore);

  try {
    const response = await analyzeUser(
      data.userData.email!,
      data.userData.gmailConnected
    );

    // Store analysis results
    onboardingStore.setAnalysis({
      message1: response.analysis.insights[0],
      message2: response.analysis.insights[1],
      message3: response.analysis.insights[2]
    });

    onboardingStore.nextStep(); // Move to app generation
    return response.analysis;
  } catch (error) {
    console.error('Analysis failed:', error);
    throw error;
  }
}

// Step 3: Generate apps (Screen 8-10)
async function generateApps(analysis: UserAnalysisResult) {
  const data = get(onboardingStore);

  try {
    const response = await generateStarterApps(
      data.workspaceId,
      analysis
    );

    onboardingStore.setStarterApps(
      response.starter_apps.map(app => ({
        id: app.id,
        title: app.title,
        description: app.description,
        iconUrl: app.icon_url,
        reason: app.reasoning
      }))
    );

    // If not all ready, start polling
    if (!response.ready_to_launch) {
      await pollAppStatus(data.workspaceId);
    }

    onboardingStore.nextStep();
  } catch (error) {
    console.error('App generation failed:', error);
    throw error;
  }
}

// Step 4: Poll for app completion
async function pollAppStatus(workspaceId: string) {
  let attempts = 0;
  const maxAttempts = 30; // 5 minutes with 10s polling
  const maxDelay = 10000; // 10 seconds
  let delay = 2000; // Start at 2 seconds

  while (attempts < maxAttempts) {
    await new Promise(resolve => setTimeout(resolve, delay));

    try {
      const response = await checkAppsStatus(workspaceId);

      if (response.ready_to_launch) {
        onboardingStore.setStarterApps(
          response.starter_apps.map(app => ({
            id: app.id,
            title: app.title,
            description: app.description,
            iconUrl: app.icon_url,
            reason: app.reasoning
          }))
        );
        return true;
      }

      attempts++;
      delay = Math.min(delay * 1.5, maxDelay); // Exponential backoff
    } catch (error) {
      console.error('Status check failed:', error);
      attempts++;
    }
  }

  throw new Error('App generation timeout');
}

// Step 5: Complete onboarding
function completeOnboarding() {
  onboardingStore.complete();
  // Redirect to main application
}
```

#### 3. State Management with Store

```svelte
<script>
  import { onboardingStore, onboardingProgress } from '$lib/stores/onboardingStore';

  $: step = $onboardingStore.currentStep;
  $: progress = $onboardingProgress;
  $: analysis = $onboardingStore.analysis;
  $: starterApps = $onboardingStore.userData.starterApps;
</script>

<!-- Display progress -->
<div class="progress-bar">
  <div style="width: {progress}%"></div>
</div>

<!-- Display insights -->
{#if analysis.message1}
  <div class="insights">
    <p>{analysis.message1}</p>
    <p>{analysis.message2}</p>
    <p>{analysis.message3}</p>
  </div>
{/if}

<!-- Display starter apps -->
{#if starterApps}
  <div class="apps-grid">
    {#each starterApps as app (app.id)}
      <div class="app-card">
        <img src={app.iconUrl} alt={app.title} />
        <h3>{app.title}</h3>
        <p>{app.description}</p>
        <p class="reason">{app.reason}</p>
      </div>
    {/each}
  </div>
{/if}
```

### Error Handling Best Practices

#### 1. Handle Analysis Failures

```typescript
async function analyzeWithFallback(email: string) {
  try {
    return await analyzeUser(email, true);
  } catch (error) {
    console.error('Analysis failed, using defaults:', error);

    // Return fallback analysis
    return {
      analysis: {
        insights: [
          "Ready to build something amazing",
          "Organized and intentional",
          "Focused on getting things done"
        ],
        interests: ["productivity", "organization", "creativity"],
        tools_used: ["Email", "Calendar"],
        profile_summary: "A builder ready to create their personalized OS",
        raw_data: {}
      }
    };
  }
}
```

#### 2. Handle Generation Timeouts

```typescript
async function generateWithTimeout(
  workspaceId: string,
  analysis: UserAnalysisResult,
  timeoutMs = 300000 // 5 minutes
) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await generateStarterApps(workspaceId, analysis);
    return response;
  } finally {
    clearTimeout(timeout);
  }
}
```

#### 3. Retry Logic for Polling

```typescript
async function checkStatusWithRetry(
  workspaceId: string,
  maxRetries = 3
) {
  let lastError: Error | null = null;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await checkAppsStatus(workspaceId);
    } catch (error) {
      lastError = error as Error;
      console.error(`Status check failed (attempt ${attempt + 1}):`, error);

      if (attempt < maxRetries - 1) {
        await new Promise(r => setTimeout(r, 1000 * (attempt + 1)));
      }
    }
  }

  throw lastError;
}
```

---

## Testing the Endpoints

### Using cURL

#### Test Analysis Endpoint

```bash
curl -X POST http://localhost:5000/api/osa-onboarding/analyze \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alex@example.com",
    "gmail_connected": true,
    "calendar_connected": false
  }'
```

#### Test App Generation Endpoint

```bash
curl -X POST http://localhost:5000/api/osa-onboarding/generate-apps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "analysis": {
      "insights": ["insight1", "insight2", "insight3"],
      "interests": ["productivity", "automation"],
      "tools_used": ["Figma", "Notion"],
      "profile_summary": "test",
      "raw_data": {}
    }
  }'
```

#### Test Status Endpoint

```bash
curl -X GET "http://localhost:5000/api/osa-onboarding/apps-status?workspace_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer <token>"
```

### Using TypeScript Tests

```typescript
import { describe, it, expect, beforeAll } from 'vitest';
import {
  analyzeUser,
  generateStarterApps,
  checkAppsStatus
} from '$lib/api/osa-onboarding';

describe('OSA Onboarding API', () => {
  let authToken: string;
  let workspaceId: string;

  beforeAll(async () => {
    // Setup: Get auth token and workspace ID
    authToken = await getTestToken();
    workspaceId = await createTestWorkspace();
  });

  it('should analyze user', async () => {
    const response = await analyzeUser(
      'test@example.com',
      true
    );

    expect(response.analysis).toBeDefined();
    expect(response.analysis.insights).toHaveLength(3);
    expect(response.analysis.interests.length).toBeGreaterThan(0);
  });

  it('should generate starter apps', async () => {
    const analysis = await analyzeUser('test@example.com', true);
    const response = await generateStarterApps(
      workspaceId,
      analysis.analysis
    );

    expect(response.starter_apps).toHaveLength(4);
    expect(response.starter_apps[0].status).toMatch(/generating|ready/);
  });

  it('should check app status', async () => {
    const response = await checkAppsStatus(workspaceId);

    expect(response.starter_apps).toBeDefined();
    expect(response.ready_to_launch).toBe(typeof true);
  });
});
```

---

## Database Migrations

### Schema

The OSA Onboarding system uses the `workspace_onboarding_profiles` table:

```sql
CREATE TABLE workspace_onboarding_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL UNIQUE REFERENCES workspaces(id),
  user_id UUID NOT NULL REFERENCES users(id),
  analysis_data JSONB NOT NULL,
  starter_apps_data JSONB NOT NULL,
  onboarding_method VARCHAR(50) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_onboarding_workspace ON workspace_onboarding_profiles(workspace_id);
CREATE INDEX idx_onboarding_user ON workspace_onboarding_profiles(user_id);
```

### Running Migrations

```bash
# Run all pending migrations
go run ./cmd/migrate

# Verify schema
psql $DATABASE_URL -c "\d workspace_onboarding_profiles"
```

### Sample Seeding Data

```sql
INSERT INTO workspace_onboarding_profiles (
  workspace_id,
  user_id,
  analysis_data,
  starter_apps_data,
  onboarding_method
) VALUES (
  '550e8400-e29b-41d4-a716-446655440000',
  '550e8400-e29b-41d4-a716-446655440001',
  '{
    "insights": ["No-code builder energy", "Design focused", "AI curious"],
    "interests": ["productivity", "automation", "design"],
    "tools_used": ["Figma", "Notion"],
    "profile_summary": "A designer interested in no-code solutions",
    "raw_data": {}
  }'::jsonb,
  '[
    {
      "id": "app-001",
      "title": "Design Tracker",
      "description": "Track design projects",
      "icon_emoji": "🎨",
      "status": "ready",
      "workflow_id": "wf-001"
    }
  ]'::jsonb,
  'osa_build'
);
```

---

## Environment Variables

### Required for Backend

```bash
# AI Provider configuration (for user analysis)
AI_PROVIDER=anthropic              # Options: anthropic, groq, ollama_local, ollama_cloud
ANTHROPIC_API_KEY=sk-...          # If using Anthropic
GROQ_API_KEY=gsk_...              # If using Groq

# OSA Integration (for app generation)
OSA_CLIENT_ID=...
OSA_CLIENT_SECRET=...
OSA_API_URL=https://...

# Database
DATABASE_URL=postgresql://user:pass@localhost/businessos

# Authentication
SECRET_KEY=your-secret-key        # For JWT signing
TOKEN_ENCRYPTION_KEY=...          # For OAuth token encryption
```

### Optional Configuration

```bash
# Caching
REDIS_URL=redis://localhost:6379  # Optional, for session caching

# Logging
LOG_LEVEL=info                    # info, debug, warn, error

# Feature flags
ONBOARDING_ENABLED=true
STARTER_APPS_GENERATION=true
```

---

## Troubleshooting

### Common Issues

#### 1. "Analysis returns empty insights"

**Cause**: AI provider not configured or failing silently

**Solution**:
```bash
# Check AI provider is set
echo $AI_PROVIDER

# Verify API keys
echo $ANTHROPIC_API_KEY | head -c 10

# Check logs for AI errors
docker logs businessos-backend | grep -i "analysis"
```

#### 2. "App generation times out"

**Cause**: OSA orchestrator not responding

**Solution**:
```bash
# Verify OSA is reachable
curl $OSA_API_URL/health

# Check app generation logs
docker logs businessos-backend | grep -i "GenerateApp"

# Increase polling timeout on frontend
const MAX_WAIT = 600000; // 10 minutes
```

#### 3. "Workspace not found after generation"

**Cause**: Profile not saved to database

**Solution**:
```bash
# Check database connection
psql $DATABASE_URL -c "SELECT 1"

# Verify table exists
psql $DATABASE_URL -c "\d workspace_onboarding_profiles"

# Check for errors in logs
docker logs businessos-backend | grep "SaveOnboardingProfile"
```

#### 4. "401 Unauthorized on all requests"

**Cause**: Invalid or expired JWT token

**Solution**:
```bash
# Refresh authentication
# Call login endpoint to get new token

# Verify token format in request
# Should be: Authorization: Bearer <token>

# Check token expiration
# Tokens default to 24 hour expiration
```

---

## API Response Codes Reference

| Code | Meaning | Common Cause |
|------|---------|-------------|
| `200` | OK | Request successful |
| `400` | Bad Request | Invalid input, missing required fields |
| `401` | Unauthorized | Missing or invalid authentication token |
| `404` | Not Found | Resource does not exist (workspace, profile) |
| `422` | Unprocessable Entity | Validation error (invalid username) |
| `500` | Internal Server Error | Server error, check logs |
| `503` | Service Unavailable | External service down (OSA, AI provider) |

---

## Rate Limiting

Currently no rate limiting is enforced. Future implementation will include:

- **Analyze endpoint**: 1 request per user per hour
- **Generate apps**: 1 request per workspace per day
- **Status polling**: 30 requests per minute
- **Check username**: 10 requests per minute per IP

---

## Deprecation Policy

API changes follow semantic versioning:

- **Major (v2.0)**: Breaking changes, previous version deprecated for 6 months
- **Minor (v1.1)**: New features, backwards compatible
- **Patch (v1.0.1)**: Bug fixes, security patches

Current version: **v1.0.0**

---

## Support & Contact

For API issues or questions:

- **Documentation**: See this file
- **Issues**: Report on GitHub
- **Slack**: #backend-support channel
- **Email**: api-support@businessos.io

Last updated: January 2024
