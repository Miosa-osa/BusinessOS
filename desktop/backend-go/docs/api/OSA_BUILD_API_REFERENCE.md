# OSA Build API Reference

Comprehensive API documentation for the OSA Build Phase 3 endpoints, including onboarding flow, user analysis, personalized app generation, and username management.

**Base URL**: `https://api.businessos.io/api`

**Current Version**: `v1.0.0`

---

## Table of Contents

1. [Authentication](#authentication)
2. [Endpoints](#endpoints)
   - [OSA Onboarding](#osa-onboarding-endpoints)
   - [Username Management](#username-management-endpoints)
3. [Data Models](#data-models)
4. [Error Handling](#error-handling)
5. [Rate Limiting](#rate-limiting)
6. [Examples](#examples)

---

## Authentication

All endpoints (unless marked as public) require authentication via Bearer token in the `Authorization` header.

**Authentication Header Format:**
```
Authorization: Bearer <jwt_token>
```

**Token Source**: Obtained from `/api/auth/sign-in/email` or `/api/auth/google/callback/login`

**Token Validation**:
- Tokens are validated against the `sessions` table
- For horizontal scaling, sessions are cached in Redis
- Token expiration: 24 hours (configurable)

**Public Endpoints** (no auth required):
- `GET /users/check-username/:username` - Check username availability
- `GET /osa-onboarding/profile?workspace_id=...` - Get saved profile (any workspace)

---

## Endpoints

### OSA Onboarding Endpoints

#### 1. Analyze User Data

Analyzes user's email, Gmail data, and Calendar data to generate personalized insights about their interests, tools, and workflow patterns.

**Endpoint**: `POST /osa-onboarding/analyze`

**Authentication**: Required (Bearer token)

**Content-Type**: `application/json`

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address (used for analysis) |
| `gmail_connected` | boolean | Yes | Whether Gmail is connected and analyzed |
| `calendar_connected` | boolean | No | Whether Calendar data is available (default: false) |

**Request Example:**

```bash
curl -X POST https://api.businessos.io/api/osa-onboarding/analyze \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alex@example.com",
    "gmail_connected": true,
    "calendar_connected": false
  }'
```

**Request Body (JSON):**

```json
{
  "email": "alex@example.com",
  "gmail_connected": true,
  "calendar_connected": false
}
```

**Response (200 OK):**

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
      "gmail_summary": "Analyzed recent emails",
      "email_domain": "example.com"
    }
  }
}
```

**Response Schema:**

```typescript
interface AnalyzeUserResponse {
  analysis: {
    insights: string[];              // 3 conversational insights about user
    interests: string[];             // 3-5 key interests detected
    tools_used: string[];            // Tools the user actively uses
    profile_summary: string;         // Full text summary of user profile
    raw_data: Record<string, any>;   // Debug data from analysis
  }
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Invalid request | Missing required fields (email, gmail_connected) |
| `401` | Unauthorized | Missing or invalid authentication token |
| `500` | Internal error | Failed to analyze user data (check logs for AI provider errors) |

**Error Example:**

```json
{
  "error": "email is required"
}
```

**Notes:**

- Analysis is performed based on email domain and (optionally) Gmail data
- If Gmail connection fails, analysis uses email-based heuristics as fallback
- Results are deterministic based on email for testing purposes
- Actual AI analysis powered by configured AI provider (Anthropic, Groq, or Ollama)
- Response time: 2-5 seconds (depends on AI provider)

---

#### 2. Generate Starter Apps

Creates 4 personalized starter applications based on user analysis. Each app is tailored to the user's interests and workflow patterns.

**Endpoint**: `POST /osa-onboarding/generate-apps`

**Authentication**: Required (Bearer token)

**Content-Type**: `application/json`

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Target workspace for app creation |
| `analysis` | UserAnalysisResult | Yes | Output from `/analyze` endpoint |

**Request Example:**

```bash
curl -X POST https://api.businessos.io/api/osa-onboarding/generate-apps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
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
      "profile_summary": "A no-code builder...",
      "raw_data": {}
    }
  }'
```

**Request Body (JSON):**

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

**Response (200 OK):**

```json
{
  "starter_apps": [
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
    },
    {
      "id": "app-002",
      "title": "Figma Companion",
      "description": "Quick access to your Figma workflows",
      "icon_emoji": "🎨",
      "icon_url": "https://cdn.businessos.io/icons/figma-companion.png",
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
      "icon_url": "https://cdn.businessos.io/icons/idea-inbox.png",
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
      "icon_url": "https://cdn.businessos.io/icons/daily-focus.png",
      "reasoning": "For staying focused on priorities",
      "category": "daily",
      "status": "ready",
      "workflow_id": "wf-12348"
    }
  ],
  "ready_to_launch": false
}
```

**Response Schema:**

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

**Error Example:**

```json
{
  "error": "workspace_id is required"
}
```

**Notes:**

- This endpoint both generates apps and saves the profile to database
- Generation is asynchronous; apps start in `generating` status
- Use `/apps-status` endpoint to poll for completion
- 4 apps are always generated with these categories:
  1. Interest-based tracker
  2. Tool-based companion
  3. Feedback/idea inbox
  4. Daily focus utility
- Response time: 1-2 seconds (app generation happens asynchronously)

---

#### 3. Check App Generation Status

Polls the current status of starter app generation. Use this to track when apps transition from `generating` to `ready`.

**Endpoint**: `GET /osa-onboarding/apps-status?workspace_id=<uuid>`

**Authentication**: Required (Bearer token)

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Workspace ID to check status for |

**Request Example:**

```bash
curl -X GET "https://api.businessos.io/api/osa-onboarding/apps-status?workspace_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer <token>"
```

**Response (200 OK):**

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
      "icon_url": "https://cdn.businessos.io/icons/productivity-tracker.png",
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
      "icon_url": "https://cdn.businessos.io/icons/figma-companion.png",
      "reasoning": "Because you use Figma frequently",
      "category": "companion",
      "status": "ready",
      "workflow_id": "wf-12346"
    },
    {
      "id": "app-003",
      "title": "Idea Inbox",
      "description": "Capture ideas and get feedback",
      "icon_emoji": "💡",
      "icon_url": "https://cdn.businessos.io/icons/idea-inbox.png",
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
      "icon_url": "https://cdn.businessos.io/icons/daily-focus.png",
      "reasoning": "For staying focused on priorities",
      "category": "daily",
      "status": "ready",
      "workflow_id": "wf-12348"
    }
  ],
  "ready_to_launch": true
}
```

**Response Schema:**

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

**Error Example:**

```json
{
  "error": "workspace_id required"
}
```

**Notes:**

- Safe to call repeatedly; no rate limiting
- Returns cached profile plus current status of each app
- `ready_to_launch` becomes true when all 4 apps have status `ready`
- Recommended polling strategy: exponential backoff (start at 2s, max 10s)
- Response time: < 200ms

---

#### 4. Get Onboarding Profile

Retrieves the complete saved onboarding profile for a workspace, including analysis and starter apps. Does not require authentication.

**Endpoint**: `GET /osa-onboarding/profile?workspace_id=<uuid>`

**Authentication**: Optional

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspace_id` | string (UUID) | Yes | Workspace ID to retrieve profile for |

**Request Example:**

```bash
curl -X GET "https://api.businessos.io/api/osa-onboarding/profile?workspace_id=550e8400-e29b-41d4-a716-446655440000"
```

**Response (200 OK):**

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
      "icon_url": "https://cdn.businessos.io/icons/productivity-tracker.png",
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
      "icon_url": "https://cdn.businessos.io/icons/figma-companion.png",
      "reasoning": "Because you use Figma frequently",
      "category": "companion",
      "status": "ready",
      "workflow_id": "wf-12346"
    },
    {
      "id": "app-003",
      "title": "Idea Inbox",
      "description": "Capture ideas and get feedback",
      "icon_emoji": "💡",
      "icon_url": "https://cdn.businessos.io/icons/idea-inbox.png",
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
      "icon_url": "https://cdn.businessos.io/icons/daily-focus.png",
      "reasoning": "For staying focused on priorities",
      "category": "daily",
      "status": "ready",
      "workflow_id": "wf-12348"
    }
  ]
}
```

**Response Schema:**

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
| `404` | Onboarding profile not found | No profile exists for workspace |

**Error Example:**

```json
{
  "error": "Onboarding profile not found"
}
```

**Notes:**

- Returns the complete saved profile from database
- Useful for retrieving profile after onboarding completes
- No status polling needed; this is a read-only endpoint
- Public endpoint (no authentication required)
- Response time: < 100ms

---

### Username Management Endpoints

#### 5. Check Username Availability

Validates if a username is available for registration. Also performs format validation.

**Endpoint**: `GET /users/check-username/:username`

**Authentication**: Not required

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | string | Yes | Username to validate (3-50 characters) |

**Request Example:**

```bash
curl -X GET "https://api.businessos.io/api/users/check-username/alex_builds"
```

**Response (200 OK - Available):**

```json
{
  "available": true
}
```

**Response (200 OK - Not Available):**

```json
{
  "available": false,
  "reason": "This username is already taken"
}
```

**Response Schema:**

```typescript
interface CheckUsernameResponse {
  available: boolean;
  reason?: string;  // Reason why unavailable (if available == false)
}
```

**Validation Failure Response (200 OK):**

```json
{
  "available": false,
  "reason": "Username must be at least 3 characters long"
}
```

**Possible Rejection Reasons:**

| Reason | Description |
|--------|-------------|
| `Username must be at least 3 characters long` | Length validation failed |
| `Username must be 50 characters or less` | Length validation failed |
| `Username can only contain letters, numbers, and underscores` | Format validation failed |
| `This username is reserved and cannot be used` | Reserved username list check failed |
| `This username is already taken` | Existing user check failed |

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `500` | Internal error | Database error during check |

**Validation Rules:**

- **Length**: 3-50 characters
- **Format**: Alphanumeric (a-z, A-Z, 0-9) and underscores (_) only
- **Case-insensitive**: Check against all existing usernames (case-insensitive)
- **Reserved list**: Cannot use reserved system usernames (admin, osa, api, etc.)

**Reserved Usernames:**

```
admin, osa, test, root, system, support, help, api, www, mail,
ftp, smtp, info, login, register, signup, signin, profile,
settings, search, discover, marketplace, workspace, team, project,
task, dashboard, about, contact, terms, privacy, blog, docs,
status, businessos, miosa
```

**Notes:**

- No authentication required
- Public endpoint for UX (show availability in real-time as user types)
- Returns HTTP 200 for all cases (invalid format, taken, or available)
- Response time: < 50ms (simple database query)

---

#### 6. Set User Username

Sets or updates the username for the authenticated user.

**Endpoint**: `PATCH /users/me/username`

**Authentication**: Required (Bearer token)

**Content-Type**: `application/json`

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | New username (3-50 chars, alphanumeric + underscore) |

**Request Example:**

```bash
curl -X PATCH "https://api.businessos.io/api/users/me/username" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alex_builds"
  }'
```

**Request Body (JSON):**

```json
{
  "username": "alex_builds"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "username": "alex_builds"
}
```

**Response Schema:**

```typescript
interface SetUsernameResponse {
  success: boolean;
  username: string;
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `400` | Invalid request | Missing or malformed request body |
| `401` | Not authenticated | Missing or invalid authentication token |
| `409` | Username is already taken | Another user has claimed this username |
| `422` | Invalid username format | Username fails validation rules |
| `500` | Internal error | Database error during update |

**Error Examples:**

```json
{
  "error": "Invalid request: Field validation for 'username' failed on the 'required' tag"
}
```

```json
{
  "error": "Invalid username format",
  "reason": "Username must be at least 3 characters long"
}
```

```json
{
  "error": "Username is already taken",
  "reason": "This username is already in use by another user"
}
```

**Notes:**

- Requires valid JWT authentication
- User ID extracted from authentication token
- Username must pass all validation rules (see Check Username endpoint)
- Uses database transaction for atomicity
- Sets `username_claimed_at` timestamp on first claim
- Allows username changes after initial claim (configurable policy)
- Response time: < 200ms

---

#### 7. Get Current User

Retrieves information about the currently authenticated user.

**Endpoint**: `GET /users/me`

**Authentication**: Required (Bearer token)

**Request Example:**

```bash
curl -X GET "https://api.businessos.io/api/users/me" \
  -H "Authorization: Bearer <token>"
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "email": "alex@example.com",
  "username": "alex_builds",
  "username_claimed_at": "2024-01-18T10:30:00Z",
  "display_name": "Alex",
  "avatar_url": "https://cdn.businessos.io/avatars/alex.png",
  "created_at": "2024-01-15T08:00:00Z",
  "updated_at": "2024-01-18T10:30:00Z"
}
```

**Response Schema:**

```typescript
interface UserProfile {
  id: string;                        // User UUID
  email: string;                     // Email address
  username: string;                  // Username (if claimed)
  username_claimed_at?: string;      // ISO-8601 timestamp of username claim
  display_name?: string;             // Display name
  avatar_url?: string;               // URL to avatar image
  created_at: string;                // ISO-8601 timestamp
  updated_at: string;                // ISO-8601 timestamp
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| `401` | Unauthorized | Missing or invalid authentication token |
| `500` | Internal error | Database error |

**Notes:**

- Requires valid JWT authentication
- Returns complete user profile
- Username field only present if user has claimed a username
- Response time: < 100ms

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
  raw_data: Record<string, any>;   // Debug/analytics data (internal use)
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
    "gmail_summary": "Analyzed recent emails",
    "email_domain": "example.com"
  }
}
```

---

### StarterApp

Represents a single personalized starter application.

```typescript
interface StarterApp {
  id: string;                          // UUID for this app
  title: string;                       // App name (e.g., "Productivity Tracker")
  description: string;                 // Brief description (1-2 sentences)
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

**App Categories:**

| Category | Purpose | Example |
|----------|---------|---------|
| `tracker` | Interest-based project tracker | "Productivity Tracker" |
| `companion` | Tool-specific companion app | "Figma Companion" |
| `feedback` | Idea/feedback collection | "Idea Inbox" |
| `daily` | Daily planning/focus tool | "Daily Focus" |

---

### OnboardingProfile

Represents complete onboarding state saved in database.

```typescript
interface OnboardingProfile {
  workspace_id: string;              // Workspace UUID
  user_id: string;                   // User UUID
  analysis_data: UserAnalysisResult; // Analysis result
  starter_apps_data: StarterApp[];   // Generated apps (array of 4)
  onboarding_method: string;         // Always "osa_build"
  created_at: string;                // ISO-8601 timestamp
  updated_at: string;                // ISO-8601 timestamp
}
```

---

## Error Handling

### Standard Error Response Format

All errors follow a consistent format:

```json
{
  "error": "Human-readable error message",
  "reason": "Optional: More detailed reason or suggestion"
}
```

### HTTP Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| `200` | OK | Request successful |
| `400` | Bad Request | Invalid input, missing required fields, malformed JSON |
| `401` | Unauthorized | Missing, invalid, or expired authentication token |
| `404` | Not Found | Resource does not exist (workspace, profile, user) |
| `409` | Conflict | Username already taken, duplicate entry |
| `422` | Unprocessable Entity | Validation error (invalid username format) |
| `500` | Internal Server Error | Server error, check logs for details |
| `503` | Service Unavailable | External service unavailable (AI provider, OSA) |

### Common Error Scenarios

**Missing Authentication:**
```json
{
  "error": "Unauthorized"
}
```

**Invalid JSON Body:**
```json
{
  "error": "Invalid request: json: cannot unmarshal string into Go value of type struct"
}
```

**Invalid UUID:**
```json
{
  "error": "Invalid workspace ID"
}
```

**Resource Not Found:**
```json
{
  "error": "Onboarding profile not found"
}
```

---

## Rate Limiting

### Current Policy

Currently no rate limiting is enforced on these endpoints. Future implementation will include:

| Endpoint | Limit | Window | Description |
|----------|-------|--------|-------------|
| `/osa-onboarding/analyze` | 1 | per user per hour | Prevent repeated analysis calls |
| `/osa-onboarding/generate-apps` | 1 | per workspace per day | Prevent excessive app generation |
| `/osa-onboarding/apps-status` | 30 | per minute | Allow polling without abuse |
| `/users/check-username/:username` | 10 | per minute per IP | Prevent username enumeration |
| `/users/me/username` | 5 | per hour per user | Prevent username spam |

### Rate Limit Headers (Future)

When rate limiting is implemented, responses will include:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1642521600
```

---

## Examples

### Complete Onboarding Flow

This example shows the typical sequence of API calls during the onboarding flow.

**Step 1: Analyze User**

```bash
curl -X POST https://api.businessos.io/api/osa-onboarding/analyze \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alex@example.com",
    "gmail_connected": true,
    "calendar_connected": false
  }'
```

Response:
```json
{
  "analysis": {
    "insights": [
      "No-code builder energy, big time",
      "Design tools are your playground",
      "AI-curious, testing new platforms"
    ],
    "interests": ["productivity", "automation", "design"],
    "tools_used": ["Figma", "Notion", "Gmail"],
    "profile_summary": "A no-code builder...",
    "raw_data": {}
  }
}
```

**Step 2: Generate Apps**

```bash
curl -X POST https://api.businessos.io/api/osa-onboarding/generate-apps \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "analysis": {
      "insights": ["No-code builder energy...", "Design tools...", "AI-curious..."],
      "interests": ["productivity", "automation", "design"],
      "tools_used": ["Figma", "Notion", "Gmail"],
      "profile_summary": "A no-code builder...",
      "raw_data": {}
    }
  }'
```

Response:
```json
{
  "starter_apps": [
    { "id": "app-001", "title": "Productivity Tracker", "status": "ready", ... },
    { "id": "app-002", "title": "Figma Companion", "status": "generating", ... },
    { "id": "app-003", "title": "Idea Inbox", "status": "ready", ... },
    { "id": "app-004", "title": "Daily Focus", "status": "ready", ... }
  ],
  "ready_to_launch": false
}
```

**Step 3: Poll for Completion** (if needed)

```bash
# Initial poll (app-002 still generating)
curl -X GET "https://api.businessos.io/api/osa-onboarding/apps-status?workspace_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer eyJhbGc..."
```

Response:
```json
{
  "starter_apps": [
    { "id": "app-001", "status": "ready", ... },
    { "id": "app-002", "status": "ready", ... },  // Now ready!
    { "id": "app-003", "status": "ready", ... },
    { "id": "app-004", "status": "ready", ... }
  ],
  "ready_to_launch": true
}
```

### Username Claim Flow

**Step 1: Check Availability**

```bash
curl -X GET "https://api.businessos.io/api/users/check-username/alex_builds"
```

Response:
```json
{
  "available": true
}
```

**Step 2: Claim Username**

```bash
curl -X PATCH https://api.businessos.io/api/users/me/username \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alex_builds"
  }'
```

Response:
```json
{
  "success": true,
  "username": "alex_builds"
}
```

### Error Handling Example

**Username Already Taken**

```bash
curl -X PATCH https://api.businessos.io/api/users/me/username \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin"
  }'
```

Response (422):
```json
{
  "error": "Invalid username format",
  "reason": "This username is reserved and cannot be used"
}
```

### Polling Strategy Example (TypeScript)

```typescript
async function pollAppStatus(workspaceId: string): Promise<AppsStatusResponse> {
  const maxAttempts = 30;    // 5 minutes max
  const maxDelay = 10000;     // 10 seconds max
  let delay = 2000;           // Start at 2 seconds
  let attempts = 0;

  while (attempts < maxAttempts) {
    await new Promise(r => setTimeout(r, delay));

    try {
      const response = await fetch(
        `https://api.businessos.io/api/osa-onboarding/apps-status?workspace_id=${workspaceId}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      ).then(r => r.json());

      if (response.ready_to_launch) {
        return response;
      }

      attempts++;
      delay = Math.min(delay * 1.5, maxDelay); // Exponential backoff
    } catch (error) {
      console.error('Status check failed:', error);
      attempts++;
    }
  }

  throw new Error('App generation timeout after 5 minutes');
}
```

---

## Support & Contact

For API issues or questions:

- **Documentation**: See this file
- **Issues**: Report on GitHub or Linear
- **Slack**: #backend-support channel
- **Email**: api-support@businessos.io

---

**Last Updated**: January 2026

**API Version**: v1.0.0

**Status**: Production Ready
