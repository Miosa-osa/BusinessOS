# OSA Build Onboarding - Quick Reference Guide

## At a Glance

**What**: OSA Build onboarding system that analyzes users and generates 4 personalized starter apps

**Where**: `/api/osa-onboarding/` endpoints

**When**: Part of new user registration flow (13-step process)

**Who**: Any authenticated user can access analysis endpoints

---

## 5-Minute Integration Checklist

- [ ] Import API client: `import { analyzeUser, generateStarterApps, checkAppsStatus } from '$lib/api/osa-onboarding'`
- [ ] Import store: `import { onboardingStore } from '$lib/stores/onboardingStore'`
- [ ] Call analyze after email collection
- [ ] Display 3 insights on analysis screen
- [ ] Call generate-apps with workspace ID
- [ ] Poll apps-status with exponential backoff
- [ ] Display 4 apps when ready_to_launch = true
- [ ] Call complete() when done

---

## API Endpoints Cheat Sheet

### Analyze User
```
POST /api/osa-onboarding/analyze
{
  "email": "user@example.com",
  "gmail_connected": true
}
→ Returns: insights, interests, tools_used, profile_summary
```

### Generate Apps
```
POST /api/osa-onboarding/generate-apps
{
  "workspace_id": "uuid",
  "analysis": { ...from analyze endpoint... }
}
→ Returns: 4 starter_apps (status: generating|ready), ready_to_launch: bool
```

### Check Status
```
GET /api/osa-onboarding/apps-status?workspace_id=uuid
→ Returns: Same as generate-apps + ready_to_launch flag
```

### Get Profile
```
GET /api/osa-onboarding/profile?workspace_id=uuid
→ Returns: Saved analysis + starter_apps
```

### Check Username
```
GET /api/users/check-username/:username
→ Returns: { available: bool, suggestion?: string }
```

### Update Username
```
PATCH /api/users/me/username
{ "username": "new_name" }
→ Returns: Updated user object
```

---

## Response Models Quick View

### UserAnalysisResult
```typescript
{
  insights: string[];        // ["Insight 1", "Insight 2", "Insight 3"]
  interests: string[];       // ["productivity", "design", "automation"]
  tools_used: string[];      // ["Figma", "Notion", "Gmail"]
  profile_summary: string;   // "Full text description"
  raw_data: {};              // Debug data
}
```

### StarterApp
```typescript
{
  id: string;                // UUID
  title: string;             // "Productivity Tracker"
  description: string;       // "Track your projects"
  icon_emoji: string;        // "📚"
  icon_url: string;          // CDN URL
  reasoning: string;         // "Why it was created"
  category: string;          // "tracker" | "companion" | "feedback" | "daily"
  status: string;            // "generating" | "ready" | "failed"
  workflow_id: string;       // OSA workflow ID
}
```

---

## Frontend Flow Template

```typescript
// Step 1: Collect email (Screen 5)
async function submitEmail(email: string) {
  onboardingStore.setUserData({ email });
  onboardingStore.nextStep();
}

// Step 2: Analyze (Screen 6-7)
async function startAnalysis() {
  const { email } = get(onboardingStore).userData;
  const response = await analyzeUser(email, true);

  onboardingStore.setAnalysis({
    message1: response.analysis.insights[0],
    message2: response.analysis.insights[1],
    message3: response.analysis.insights[2]
  });

  onboardingStore.nextStep();
  return response.analysis;
}

// Step 3: Generate & Poll (Screen 8-10)
async function generateAndWait(analysis) {
  const { workspaceId } = get(onboardingStore);

  const response = await generateStarterApps(workspaceId, analysis);
  onboardingStore.setStarterApps(response.starter_apps);

  if (!response.ready_to_launch) {
    // Poll with backoff
    let delay = 2000;
    while (!allReady) {
      await sleep(delay);
      const status = await checkAppsStatus(workspaceId);
      if (status.ready_to_launch) {
        onboardingStore.setStarterApps(status.starter_apps);
        break;
      }
      delay = Math.min(delay * 1.5, 10000);
    }
  }

  onboardingStore.nextStep();
}

// Step 4: Complete
onboardingStore.complete();
// Redirect to main app
```

---

## Error Handling Patterns

### Fallback Analysis
```typescript
try {
  return await analyzeUser(email, true);
} catch (error) {
  // Return defaults
  return {
    analysis: {
      insights: ["Ready to build", "Organized", "Focused"],
      interests: ["productivity", "organization"],
      tools_used: ["Email"],
      profile_summary: "A builder ready to create",
      raw_data: {}
    }
  };
}
```

### Retry Status Check
```typescript
async function checkWithRetry(workspaceId, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await checkAppsStatus(workspaceId);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await sleep(1000 * (i + 1));
    }
  }
}
```

### Timeout Wrapper
```typescript
async function withTimeout(promise, ms = 300000) {
  const timeout = new Promise((_, reject) =>
    setTimeout(() => reject(new Error('Timeout')), ms)
  );
  return Promise.race([promise, timeout]);
}
```

---

## State Management

### Store Essentials
```typescript
import { onboardingStore } from '$lib/stores/onboardingStore';

// Navigate
$onboardingStore.nextStep();
$onboardingStore.prevStep();
$onboardingStore.goToStep(5);

// Update data
onboardingStore.setUserData({ email, username });
onboardingStore.setAnalysis({ message1, message2, message3 });
onboardingStore.setStarterApps(apps);

// Complete
onboardingStore.complete();

// Reset
onboardingStore.reset();
```

### In Svelte Components
```svelte
<script>
  import { onboardingStore, onboardingProgress } from '$lib/stores/onboardingStore';

  $: currentStep = $onboardingStore.currentStep;
  $: progress = $onboardingProgress; // 0-100%
  $: apps = $onboardingStore.userData.starterApps;
</script>

<div>Progress: {progress}%</div>
{#each apps as app (app.id)}
  <AppCard {app} />
{/each}
```

---

## Testing Quick Commands

### Test Analyze
```bash
curl -X POST http://localhost:5000/api/osa-onboarding/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","gmail_connected":true}'
```

### Test App Generation
```bash
curl -X POST http://localhost:5000/api/osa-onboarding/generate-apps \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id":"550e8400-e29b-41d4-a716-446655440000",
    "analysis":{"insights":["a","b","c"],"interests":["x"],"tools_used":["y"],"profile_summary":"z","raw_data":{}}
  }'
```

### Test Status
```bash
curl -X GET "http://localhost:5000/api/osa-onboarding/apps-status?workspace_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"
```

### Test Username
```bash
curl -X GET http://localhost:5000/api/users/check-username/newuser
```

---

## Database Queries

### Check Profile Exists
```sql
SELECT * FROM workspace_onboarding_profiles
WHERE workspace_id = '550e8400-e29b-41d4-a716-446655440000';
```

### View Analysis Data
```sql
SELECT workspace_id, analysis_data->>'insights' as insights
FROM workspace_onboarding_profiles;
```

### Update Profile
```sql
UPDATE workspace_onboarding_profiles
SET updated_at = NOW()
WHERE workspace_id = $1;
```

### Clean Test Data
```sql
DELETE FROM workspace_onboarding_profiles
WHERE created_at < NOW() - INTERVAL '7 days';
```

---

## Common Issues & Fixes

| Issue | Fix |
|-------|-----|
| 401 Unauthorized | Check Bearer token, verify expiration |
| Empty insights | Verify AI_PROVIDER env var is set |
| Apps timeout | Increase polling max delay to 10s-20s |
| Profile not found | Verify workspace_id exists and generation completed |
| Username taken | API returns suggestion, let user choose |

---

## Environment Variables

```bash
# REQUIRED
AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-...
OSA_API_URL=https://...

# OPTIONAL
REDIS_URL=redis://localhost:6379
LOG_LEVEL=info
```

---

## Key Files in Codebase

| File | Purpose |
|------|---------|
| `/desktop/backend-go/internal/handlers/osa_onboarding_handler.go` | Backend handlers |
| `/desktop/backend-go/internal/services/osa_onboarding_service.go` | Business logic |
| `/frontend/src/lib/api/osa-onboarding/index.ts` | API client |
| `/frontend/src/lib/api/osa-onboarding/types.ts` | TypeScript types |
| `/frontend/src/lib/stores/onboardingStore.ts` | Svelte state |
| `/frontend/src/routes/onboarding/` | UI screens |

---

## Performance Tips

1. **Parallel requests**: Call analyze and username check in parallel
2. **Exponential backoff**: Start polling at 2s, increase by 1.5x to max 10s
3. **Client caching**: Store analysis in localStorage via onboardingStore
4. **Early validation**: Check username availability immediately after input
5. **Timeout gracefully**: Set 5min max wait for app generation

---

## Security Notes

- All endpoints require authentication (Bearer token)
- Tokens expire after 24 hours
- Use HTTPS in production only
- No sensitive data in URLs (use POST for analysis)
- Database saves user IDs, never store passwords

---

## See Also

- Full documentation: `/docs/API_OSA_ONBOARDING.md`
- Frontend code: `/frontend/src/lib/api/osa-onboarding/`
- Backend code: `/desktop/backend-go/internal/handlers/osa_onboarding_handler.go`
- Store: `/frontend/src/lib/stores/onboardingStore.ts`

---

Last updated: January 2024
Version: 1.0.0
