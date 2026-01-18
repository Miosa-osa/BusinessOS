# End-to-End Testing Plan: AI-Powered Onboarding Flow

## Test Environment Status

✅ **Backend**: Running (Go server on port 8001)
✅ **Frontend**: Running (Vite dev server)
✅ **Groq API**: Configured (`llama-3.1-8b-instant`)

---

## Test Objectives

Verify the complete onboarding flow:
1. OAuth flow initiates and completes
2. Analysis is triggered automatically after Gmail OAuth
3. SSE streaming updates analyzing screens in real-time
4. Insights display correctly (Groq AI or fallbacks)
5. Auto-advance works between screens (2s delay)
6. No console errors or runtime issues

---

## Test Procedure

### Phase 1: Pre-Test Verification

**Backend Health Check:**
```bash
curl http://localhost:8001/health
# Expected: {"status":"ok"}
```

**Frontend Access:**
```
Navigate to: http://localhost:5173/onboarding
# Expected: Onboarding screen loads
```

**Groq API Check:**
```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
grep "GROQ_API_KEY" .env
# Expected: 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
```

---

### Phase 2: Complete Onboarding Flow

#### Step 1: Start Onboarding
1. Navigate to `http://localhost:5173/onboarding`
2. **Expected**: Welcome screen appears
3. **Verify**: No console errors

#### Step 2: Progress Through Screens
1. Click "Continue" or navigation button
2. **Expected**: Screens 1-5 navigate correctly
3. **Verify**: Animations and transitions work

#### Step 3: Gmail OAuth Initiation
1. Reach the "Connect Gmail" screen
2. Click "Connect Gmail" button
3. **Expected**:
   - localStorage sets `oauth_context` = "onboarding-signin"
   - localStorage sets `oauth_next_route` = "/onboarding/analyzing"
   - Redirects to Google OAuth consent screen
4. **Verify in Console**:
   ```javascript
   localStorage.getItem('oauth_context') === 'onboarding-signin'
   localStorage.getItem('oauth_next_route') === '/onboarding/analyzing'
   ```

#### Step 4: Complete OAuth
1. Complete Google OAuth authorization
2. **Expected**:
   - Redirects to `/auth/callback`
   - Backend creates workspace and sets session cookie
   - Analysis trigger fires automatically
   - Redirects to `/onboarding/analyzing`
3. **Verify in Console**:
   ```javascript
   // Should see in console:
   "🚀 Starting AI analysis: { userId: '...', workspaceId: '...' }"
   ```

#### Step 5: Analyzing Screen 1 (First Insight)
1. Land on `/onboarding/analyzing`
2. **Expected Behavior**:
   - Shows spinner with "Analyzing your workspace..."
   - If SSE connected: Shows "Reading your emails with AI..."
   - After 2 seconds or completion: Shows first insight
3. **Verify**:
   - Check Network tab for SSE request:
     ```
     GET /api/v1/osa-onboarding/analyze/{analysisId}/stream
     ```
   - Check Console for streaming events
   - Auto-advances to `/onboarding/analyzing-2` after 2s

**With Groq AI (Success Case):**
```
1. Spinner appears
2. "Reading your emails with AI..." appears
3. AI-generated insight appears (e.g., "Startup founder vibes")
4. "✨ AI-Generated" badge appears
5. Auto-advances after 2s
```

**With Fallback (Groq Failure Case):**
```
1. Spinner appears briefly
2. Fallback insight appears: "No-code builder energy"
3. No AI badge
4. Auto-advances after 2s
```

#### Step 6: Analyzing Screen 2 (Second Insight)
1. Land on `/onboarding/analyzing-2`
2. **Expected**:
   - Shows second insight from Groq OR fallback: "Design tools are your playground"
   - Auto-advances to `/onboarding/analyzing-3` after 2s

#### Step 7: Analyzing Screen 3 (Third Insight)
1. Land on `/onboarding/analyzing-3`
2. **Expected**:
   - Shows third insight from Groq OR fallback: "AI-curious, testing new platforms"
   - Auto-advances to next screen after 2s

---

### Phase 3: Verification Checklist

#### ✅ Frontend Verification

**Console Checks:**
```javascript
// Open Browser DevTools → Console
// Should NOT see:
- "Analysis not started - using fallback insights" (unless OAuth failed)
- Any red errors related to onboardingAnalysis
- TypeScript errors
- 404 errors for API endpoints

// SHOULD see:
- "🚀 Starting AI analysis: { userId: '...', workspaceId: '...' }"
- SSE event logs (if verbose logging enabled)
```

**Network Tab Checks:**
```
1. POST /api/v1/osa-onboarding/analyze
   Status: 200
   Response: { analysis_id: "...", status: "analyzing" }

2. GET /api/v1/osa-onboarding/analyze/{analysisId}/stream
   Status: 200
   Type: text/event-stream
   Content: SSE events streaming

3. No 500 errors
4. No CORS errors
```

**Store State Checks:**
```javascript
// In Browser DevTools → Console
import { get } from 'svelte/store';
import { onboardingAnalysis } from '$lib/stores/onboardingAnalysis';

console.log(get(onboardingAnalysis));
// Expected structure:
{
  analysisId: "uuid-here",
  status: "completed" or "analyzing",
  insights: ["insight1", "insight2", "insight3"],
  interests: [...],
  isStreaming: false,
  isLoading: false,
  error: null
}
```

#### ✅ Backend Verification

**Database Checks:**
```sql
-- Connect to PostgreSQL
-- Check analysis was created
SELECT * FROM onboarding_analyses ORDER BY created_at DESC LIMIT 1;

-- Expected columns:
-- id, user_id, workspace_id, status ('analyzing'|'completed'|'failed')
-- insights JSONB (array of 3 strings)
-- interests JSONB
-- created_at, completed_at
```

**Log Checks:**
```bash
# Backend terminal should show:
[INFO] Starting onboarding analysis for user=... workspace=...
[INFO] Fetching 50 emails from Gmail API
[INFO] Analyzing email patterns...
[INFO] Sending analysis to Groq: llama-3.1-8b-instant
[INFO] Groq response received: 3 insights generated
[INFO] SSE client connected: /analyze/{analysisId}/stream
[INFO] Streaming progress event: analyzing
[INFO] Streaming done event: completed
```

**Groq API Verification:**
```bash
# Should see in backend logs:
# Request to Groq
POST https://api.groq.com/openai/v1/chat/completions
Model: llama-3.1-8b-instant
Temperature: 0.7
Max tokens: 500

# Response from Groq
{
  "choices": [{
    "message": {
      "content": "[\"insight1\", \"insight2\", \"insight3\"]"
    }
  }]
}
```

---

### Phase 4: Edge Cases Testing

#### Test 1: Groq API Failure (Simulate)
**Setup**: Temporarily set invalid Groq API key
```bash
# Edit .env
GROQ_API_KEY=invalid-key-test
# Restart backend
```

**Expected Behavior**:
1. Analysis starts normally
2. Groq call fails (logged in backend)
3. Frontend shows fallback insights:
   - "No-code builder energy"
   - "Design tools are your playground"
   - "AI-curious, testing new platforms"
4. No "✨ AI-Generated" badge
5. Flow continues normally

**Verify**: User experience is seamless despite API failure

---

#### Test 2: SSE Stream Disconnection
**Simulate**: Disconnect network mid-stream

**Expected Behavior**:
1. SSE stream breaks
2. Frontend fallback polling kicks in (`pollForCompletion`)
3. Analysis completes via polling (every 2s)
4. Insights display correctly

**Verify**: Graceful fallback to polling

---

#### Test 3: OAuth Without Gmail Access
**Simulate**: Deny Gmail scope during OAuth

**Expected Behavior**:
1. OAuth completes without Gmail access
2. Analysis trigger logs warning: "Missing Gmail access"
3. Analyzing screens show fallback insights
4. Flow continues normally

**Verify**: No crashes, user can continue

---

#### Test 4: Fast Network (Groq Completes Instantly)
**Scenario**: Groq responds in <100ms

**Expected Behavior**:
1. Analyzing screen shows spinner briefly
2. Insight appears almost immediately
3. 2s delay still applies before auto-advance
4. Smooth transition

**Verify**: Minimum screen time enforced

---

### Phase 5: Performance Checks

**Metrics to Measure**:
```
1. OAuth → Callback → Analyzing: < 2 seconds
2. Analysis trigger → First insight: < 5 seconds (with Groq)
3. SSE connection established: < 500ms
4. Screen transitions: Smooth, no flicker
5. Memory usage: Stable (no leaks)
```

**Tools**:
- Chrome DevTools → Performance tab
- Network tab → Timing
- Memory tab → Heap snapshot

---

## Test Results Template

### ✅ Successful Test Example:
```
✅ OAuth flow completed
✅ Analysis triggered: userId=abc123, workspaceId=xyz789
✅ SSE connected: /analyze/analysis-id-123/stream
✅ Groq API responded in 2.3s
✅ Insights generated:
   1. "Startup founder vibes"
   2. "Deep in Notion and Linear workflows"
   3. "AI-first mindset"
✅ All 3 screens displayed correctly
✅ Auto-advance worked (2s delay)
✅ No console errors
✅ Database record created
```

### ❌ Failed Test Example:
```
❌ Issue: SSE stream never connected
❌ Error in console: "Failed to stream analysis progress"
❌ Network: 500 Internal Server Error on /stream endpoint
❌ Root cause: Backend logs show "Groq API key not configured"
❌ Fix needed: Check GROQ_API_KEY in .env
```

---

## Quick Test Script (Manual Browser Testing)

Open Browser Console and run:

```javascript
// 1. Check localStorage context
console.log('OAuth Context:', localStorage.getItem('oauth_context'));
console.log('Next Route:', localStorage.getItem('oauth_next_route'));

// 2. Check store state (after landing on analyzing screen)
setTimeout(async () => {
  const { onboardingAnalysis } = await import('$lib/stores/onboardingAnalysis');
  const { get } = await import('svelte/store');
  console.log('Analysis State:', get(onboardingAnalysis));
}, 3000);

// 3. Monitor SSE events
const evtSource = new EventSource('http://localhost:8001/api/v1/osa-onboarding/analyze/YOUR_ANALYSIS_ID/stream', {
  withCredentials: true
});

evtSource.onmessage = (event) => {
  console.log('SSE Event:', JSON.parse(event.data));
};

evtSource.onerror = (err) => {
  console.error('SSE Error:', err);
};
```

---

## Automated Testing (Future)

For future automation, create Playwright test:

```typescript
// e2e/onboarding-flow.spec.ts
import { test, expect } from '@playwright/test';

test('complete onboarding flow with AI analysis', async ({ page, context }) => {
  // Navigate to onboarding
  await page.goto('http://localhost:5173/onboarding');

  // Progress through screens
  await page.click('text=Continue');

  // Mock Google OAuth (requires setup)
  await context.route('https://accounts.google.com/o/oauth2/**', route => {
    route.fulfill({
      status: 302,
      headers: { Location: 'http://localhost:8001/api/auth/google/callback?code=test' }
    });
  });

  // Click Connect Gmail
  await page.click('text=Connect Gmail');

  // Wait for analyzing screen
  await page.waitForURL('**/onboarding/analyzing');

  // Verify insight appears
  const insight = await page.waitForSelector('h1.title', { timeout: 10000 });
  const text = await insight.textContent();
  expect(text).toBeTruthy();

  // Verify auto-advance
  await page.waitForURL('**/onboarding/analyzing-2', { timeout: 3000 });
});
```

---

## Known Issues & Workarounds


**Issue 1**: Groq model mismatch ✅ RESOLVED
- Previously `.env` had `llama-3.1-8b-instant`
- Code referenced `llama-3.3-70b-versatile`
- **Fixed**: Updated `.env` to `llama-3.3-70b-versatile` (2026-01-18)
**Issue 2**: CORS errors in local dev
- **Fix**: Ensure `ALLOWED_ORIGINS` includes `http://localhost:5173`

**Issue 3**: Gmail API quota limits
- **Workaround**: Use fallback insights if quota exceeded

---

## Post-Test Cleanup

```bash
# Reset test data (optional)
psql $DATABASE_URL -c "DELETE FROM onboarding_analyses WHERE created_at > NOW() - INTERVAL '1 hour';"

# Clear localStorage
localStorage.clear();

# Restart servers if needed
# Backend: Ctrl+C, then: go run cmd/server/main.go
# Frontend: Ctrl+C, then: npm run dev
```

---

## Sign-Off Criteria

Before marking this task complete, verify:

- [ ] OAuth → Analysis → Insights flow works end-to-end
- [ ] SSE streaming updates screens in real-time
- [ ] Fallback insights work when Groq fails
- [ ] Auto-advance timing is correct (2s)
- [ ] No console errors
- [ ] No network errors
- [ ] Database records created correctly
- [ ] All 3 analyzing screens display properly
- [ ] Performance is acceptable (<5s total)
- [ ] Edge cases handled gracefully

---

**Test Date**: 2026-01-18
**Tester**: Claude Code
**Environment**: Local Development
**Status**: Ready for Execution
