# OSA Build - Comprehensive Integration Testing Plan
## Complete Onboarding Flow Testing Guide

**Document Version:** 2.0
**Created:** 2026-01-18
**Branch:** `feature/ios-desktop-flow-migration`
**Status:** Active - Ready for QA

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Test Strategy Overview](#test-strategy-overview)
3. [Critical User Flows](#critical-user-flows)
4. [Manual Test Scenarios](#manual-test-scenarios)
5. [API Integration Tests](#api-integration-tests)
6. [Automated Test Recommendations](#automated-test-recommendations)
7. [Performance & Load Testing](#performance--load-testing)
8. [Commands & Quick Reference](#commands--quick-reference)
9. [Bug Report Template](#bug-report-template)
10. [Appendix](#appendix)

---

## Executive Summary

### What We're Testing

The **OSA Build Onboarding Flow** guides users from first install through personalized OS creation in 13 screens:

```
Welcome → Meet OSA → Sign In → Connect Gmail → Username →
Analyzing (1/3) → Analyzing (2/3) → Analyzing (3/3) →
Starter Apps → Ready to Launch → Main App
```

### Critical Success Criteria

- ✅ **Happy path completes** in under 3 minutes
- ✅ **All 4 starter apps generate** successfully
- ✅ **State persists** across browser refresh/crash
- ✅ **API errors handled** gracefully with retry options
- ✅ **Username validation** works correctly
- ✅ **Navigation (back/forward)** maintains state

### Testing Priorities

| Priority | Area | Why Critical |
|----------|------|--------------|
| **P0** | Happy path completion | Core user experience |
| **P0** | Username validation | Prevents duplicate accounts |
| **P0** | API error handling | Prevents user frustration |
| **P1** | State persistence | Recovery from crashes |
| **P1** | Performance | 2-minute target |
| **P2** | Edge cases | Robustness |

---

## Test Strategy Overview

### 1. Testing Pyramid

```
              /\
             /  \        E2E Tests (5-10 tests)
            /____\       - Happy path
           /      \      - Critical error scenarios
          /        \
         /__________\    Integration Tests (20-30 tests)
        /            \   - API endpoints
       /              \  - State management
      /________________\
     /                  \ Unit Tests (50+ tests)
    /____________________\ - Validation functions
                           - API client methods
```

### 2. Test Environment Setup

**Required:**
- Backend running: `http://localhost:5000`
- Frontend running: `http://localhost:5173`
- PostgreSQL database seeded
- Redis (optional, for session caching)

**Quick Start:**
```bash
cd /Users/rhl/Desktop/BusinessOS2

# 1. Start backend
cd desktop/backend-go
go run cmd/main.go

# 2. Start frontend (new terminal)
cd ../frontend
npm run dev

# 3. Verify services (new terminal)
curl http://localhost:5000/health
curl http://localhost:5173
```

### 3. Test Data Setup

Create test user before running tests:

```bash
# Test username API
./test-username-api.sh

# Expected output:
# ✓ admin (reserved) - should be unavailable
# ✓ validuser - should be available
# ✓ abc (too short) - should fail validation
```

---

## Critical User Flows

### Flow 1: Happy Path - Complete Onboarding

**Objective:** User completes entire onboarding flow successfully

**Prerequisites:**
- No existing account with test email
- Backend and frontend running
- Clean browser state (no localStorage)

**Test Steps:**

| Step | Screen | Action | Expected Result | Verify |
|------|--------|--------|-----------------|--------|
| 1 | Welcome | Visit `/onboarding` | Page loads, displays 3 feature cards | Title visible, "Get Started" button |
| 2 | Welcome | Click "Get Started" | Navigate to `/onboarding/meet-osa` | Smooth animation |
| 3 | Meet OSA | Click "Next" | Navigate to `/onboarding/signin` | OSA message displayed |
| 4 | Sign In | Enter valid email: `testuser@example.com` | Email validates | No error message |
| 5 | Sign In | Click "Continue" | Navigate to `/onboarding/gmail` | API call succeeds (check Network) |
| 6 | Gmail | Click "Skip" | Navigate to `/onboarding/username` | `gmail_connected = false` in store |
| 7 | Username | Enter: `testuser123` | Input accepts | No validation error |
| 8 | Username | Click "Check Availability" | Shows loading spinner | Button shows loading state |
| 9 | Username | Wait for response | "✓ Available!" displayed | Continue button enabled |
| 10 | Username | Click "Continue" | Navigate to `/onboarding/analyzing/1` | Progress shows "1/3" |
| 11 | Analyzing 1 | Auto-analyze starts | Loading animation | API: POST `/api/osa-onboarding/analyze` |
| 12 | Analyzing 1 | Wait 5-10s | Transition to step 2 | Insight message displayed |
| 13 | Analyzing 2 | Wait | Transition to step 3 | Progress shows "2/3" |
| 14 | Analyzing 3 | Wait | Transition to starter apps | Progress shows "3/3" |
| 15 | Starter Apps | Auto-generation starts | 4 app cards appear | Each shows "generating" status |
| 16 | Starter Apps | Wait up to 30s | All apps show "ready" | API polling: GET `/api/osa-onboarding/apps-status` |
| 17 | Starter Apps | Verify apps | All 4 apps display correctly | Title, description, icon, reasoning visible |
| 18 | Starter Apps | Click "Ready to Launch" | Navigate to `/onboarding/ready` | Success message shown |
| 19 | Ready | Verify profile summary | Username, 4 apps listed | All data correct |
| 20 | Ready | Click "Enter OSA Build" | Redirect to main app | Onboarding complete |

**Verification Checklist:**

- [ ] Total time: < 3 minutes
- [ ] No console errors in DevTools
- [ ] No API errors in Network tab
- [ ] localStorage cleared after completion
- [ ] Database has onboarding profile saved
- [ ] All 4 starter apps in database with status 'ready'

**Database Verification:**

```sql
-- Check onboarding profile was saved
SELECT
  workspace_id,
  analysis_data->>'profile_summary' as summary,
  jsonb_array_length(starter_apps_data) as app_count,
  created_at
FROM workspace_onboarding_profiles
WHERE user_id = (SELECT id FROM users WHERE email = 'testuser@example.com')
ORDER BY created_at DESC
LIMIT 1;

-- Expected: 1 row with app_count = 4
```

---

### Flow 2: Skip Gmail Connection

**Objective:** Onboarding completes without Gmail data

**Key Differences:**
- Analysis uses email-based heuristics only
- Apps may be less personalized
- Faster analysis (no external API calls)

**Test Steps:**

1. Through Step 5 (Sign In)
2. **At Gmail screen:** Click "Skip" instead of connecting
3. Verify: `gmailConnected = false` in onboardingStore
4. Continue through Username → Analyzing → Apps
5. Verify: Apps still generated (4 apps)
6. Complete onboarding

**Verification:**
- [ ] Analysis completes faster (< 5s)
- [ ] 4 apps generated with reasonable defaults
- [ ] No Gmail-specific insights in analysis

---

### Flow 3: Navigate Back Through Flow

**Objective:** User can go back without losing data

**Test Steps:**

| From Screen | Click "Back" | Expected Destination | State Preserved |
|-------------|--------------|---------------------|-----------------|
| Username | Back | Gmail | email still set |
| Gmail | Back | Sign In | email still set |
| Sign In | Back | Meet OSA | email still set |
| Meet OSA | Back | Welcome | email still set |
| Welcome | Back | Stay on Welcome | No crash |

**After going back to Welcome:**
- Click "Get Started" again
- Verify: Email still in input field
- Verify: Can resume from where left off

**Verification:**
- [ ] No data lost when navigating back
- [ ] Can resume forward from any point
- [ ] Store state updates correctly
- [ ] No console errors during navigation

---

### Flow 4: Browser Close & Resume

**Objective:** Onboarding state recovers after browser crash

**Test Steps:**

1. Complete through Step 9 (Username claimed, Continue clicked)
2. **Before Analyzing completes:**
   - Open DevTools → Application → Local Storage
   - Verify `onboarding` key exists with full state
3. **Close browser completely** (not just tab)
4. Reopen browser, visit `/onboarding`
5. **Expected:** Redirect to current step (Analyzing 1)
6. **Verify:** Analysis resumes automatically
7. Complete flow normally

**Verification:**
- [ ] State restored from localStorage
- [ ] Username preserved
- [ ] Email preserved
- [ ] Analysis resumes without re-entering data
- [ ] Apps generate successfully

**localStorage Inspection:**

```javascript
// In DevTools Console
const stored = JSON.parse(localStorage.getItem('onboarding'));
console.log('Current step:', stored.currentStep);
console.log('Username:', stored.userData.username);
console.log('Email:', stored.userData.email);
console.log('Gmail connected:', stored.userData.gmailConnected);
```

---

### Flow 5: Username Validation Edge Cases

**Objective:** All username rules enforced correctly

**Test Cases:**

| Username | Expected Result | Error Message |
|----------|----------------|---------------|
| `ab` | ❌ Rejected | "Must be at least 3 characters" |
| `validuser` | ✅ Available | "✓ Available!" |
| `admin` | ❌ Taken (reserved) | "Username already taken" |
| `root` | ❌ Taken (reserved) | "Username already taken" |
| `test123` | ✅ Available | "✓ Available!" |
| `user-name` | ❌ Invalid format | "Invalid characters" |
| `user!name` | ❌ Invalid format | "Invalid characters" |
| `user name` | ❌ Invalid format | "Invalid characters" |
| `user.name` | ❌ Invalid format | "Invalid characters" |
| `_underscore` | ✅ Available (if not taken) | "✓ Available!" |
| `user_123` | ✅ Available | "✓ Available!" |
| `123user` | ❌ Invalid format | "Must start with letter" |
| `a` | ❌ Too short | "Must be at least 3 characters" |
| (255+ chars) | ❌ Too long | "Must be less than 30 characters" |

**Reserved Username List:**
```
admin, root, osa, test, support, system, api, www, mail, ftp
```

**Validation Rules (from API):**
- 3-30 characters
- Alphanumeric and underscores only
- Must start with letter
- Case-insensitive matching

**Test Script:**

```bash
# Run automated username validation tests
./test-username-api.sh

# Expected output shows pass/fail for each test case
```

---

### Flow 6: Network Error Recovery

**Objective:** App handles network errors gracefully

**Test Setup:**
1. Open DevTools → Network → Throttling → Offline
2. Attempt actions that require API calls

**Test Scenarios:**

| Action | Error Type | Expected Behavior |
|--------|-----------|-------------------|
| Check username availability | Network offline | "Network error. Check connection." + Retry button |
| Submit email | Network offline | "Cannot connect. Try again." + Retry button |
| Analyzing step | Network timeout | "Taking longer than expected..." + Continue waiting or Skip option |
| Apps status poll | Network error | Retry automatically with exponential backoff |

**Recovery Steps:**

1. **At Username screen:**
   - Go offline (DevTools → Network → Offline)
   - Enter username, click "Check Availability"
   - Verify: Error message shows
   - Go online
   - Click "Retry"
   - Verify: Availability check succeeds

2. **At Analyzing screen:**
   - Throttle to slow 3G
   - Verify: "Analyzing..." message shows
   - After 30s, verify: "Taking longer than expected" shows
   - Restore network
   - Verify: Analysis completes

**Verification:**
- [ ] Error messages are user-friendly
- [ ] Retry button present and functional
- [ ] No data lost during error
- [ ] App doesn't crash on network errors
- [ ] Auto-retry works for polling endpoints

---

### Flow 7: API Error Handling

**Objective:** Backend errors handled gracefully

**Test Scenarios:**

| API Endpoint | Error Code | Trigger Method | Expected Behavior |
|--------------|-----------|----------------|-------------------|
| `/api/osa-onboarding/analyze` | 500 | Backend crash | "Analysis failed. Please try again." |
| `/api/osa-onboarding/generate-apps` | 500 | Backend error | "App generation failed. Retry?" |
| `/api/users/check-username/:username` | 409 | Duplicate username | "Username already taken" |
| `/api/osa-onboarding/analyze` | 400 | Invalid email | "Invalid email format" |
| `/api/osa-onboarding/apps-status` | 404 | No profile | "Profile not found. Restart onboarding?" |

**How to Trigger Errors:**

```bash
# Simulate backend crash (stop backend temporarily)
# Then attempt API calls

# Or mock 500 errors in backend code:
# internal/handler/onboarding.go
func (h *Handler) AnalyzeUser(c *gin.Context) {
    c.JSON(500, gin.H{"error": "Internal server error"})
    return
}
```

**Verification:**
- [ ] Error message displayed clearly
- [ ] User can retry action
- [ ] No unhandled exceptions
- [ ] App doesn't freeze
- [ ] State remains intact

---

## Manual Test Scenarios

### Scenario 1: Mobile Responsive Design

**Device Matrix:**

| Device | Viewport | Test Priority |
|--------|----------|---------------|
| iPhone SE | 375x667 | P0 |
| iPhone 12/13 | 390x844 | P0 |
| iPhone 14 Pro Max | 430x932 | P1 |
| iPad Mini | 768x1024 | P1 |
| iPad Pro | 1024x1366 | P2 |

**Test Checklist per Device:**

- [ ] All text readable without zoom
- [ ] Buttons large enough to tap (min 44x44px)
- [ ] No horizontal scroll
- [ ] Feature cards stack vertically
- [ ] App cards in single column
- [ ] Forms fill screen width appropriately
- [ ] Back button accessible
- [ ] Progress indicator visible

**Quick Mobile Test:**

```javascript
// In DevTools Console
// Resize viewport programmatically
window.resizeTo(375, 667); // iPhone SE

// Or use Device Toolbar (Cmd+Shift+M)
```

---

### Scenario 2: Keyboard Navigation

**Objective:** Entire flow accessible via keyboard only

**Test Steps:**

1. Start at Welcome screen
2. Press Tab repeatedly
3. Verify: Focus moves to "Get Started" button
4. Press Enter → Navigate to Meet OSA
5. Press Tab → Focus on "Next" button
6. Press Enter → Navigate to Sign In
7. Press Tab → Focus on email input
8. Type email, press Tab → Focus on Continue button
9. Press Enter → Submit form
10. Continue through entire flow using only keyboard

**Verification:**
- [ ] All interactive elements reachable via Tab
- [ ] Focus indicators clearly visible
- [ ] Enter key activates buttons
- [ ] Escape key cancels modals (if any)
- [ ] No keyboard traps
- [ ] Logical tab order

**Focus Indicator Test:**

```css
/* Verify this CSS exists and is visible */
button:focus,
input:focus {
  outline: 2px solid blue;
  outline-offset: 2px;
}
```

---

### Scenario 3: Screen Reader Accessibility

**Tools:**
- macOS: VoiceOver (Cmd+F5)
- Windows: NVDA or JAWS
- Linux: Orca

**Test Checklist:**

- [ ] Page title announced on each screen
- [ ] Form labels associated with inputs
- [ ] Buttons have descriptive labels
- [ ] Error messages announced
- [ ] Loading states announced (aria-live)
- [ ] Success messages announced
- [ ] Progress updates announced
- [ ] No empty links or buttons

**VoiceOver Test Script:**

1. Enable VoiceOver (Cmd+F5)
2. Navigate to `/onboarding`
3. Use VO+Right Arrow to move through content
4. Verify each element announced correctly
5. At email input, verify: "Email input, required"
6. At Continue button, verify: "Continue button, disabled" or "enabled"
7. At error, verify: Error message read aloud

**ARIA Attributes to Verify:**

```html
<!-- Username input -->
<input
  type="text"
  id="username"
  aria-label="Username"
  aria-describedby="username-error username-help"
  aria-required="true"
  aria-invalid="false"
/>
<span id="username-help">3-30 characters, letters and underscores</span>
<span id="username-error" role="alert">Username already taken</span>

<!-- Loading state -->
<div aria-live="polite" aria-busy="true">
  Analyzing your profile...
</div>

<!-- App cards -->
<div role="article" aria-label="Starter App 1">
  <h3>Productivity Tracker</h3>
  <p>Track and organize your productivity projects</p>
</div>
```

---

### Scenario 4: Browser Compatibility

**Test Matrix:**

| Browser | Version | OS | Priority | Status |
|---------|---------|----|---------:|--------|
| Chrome | Latest | macOS | P0 | ⬜ |
| Chrome | Latest | Windows | P0 | ⬜ |
| Firefox | Latest | macOS | P0 | ⬜ |
| Firefox | Latest | Windows | P1 | ⬜ |
| Safari | 16+ | macOS | P0 | ⬜ |
| Safari | Latest | iOS | P0 | ⬜ |
| Edge | Latest | Windows | P1 | ⬜ |

**Critical Features per Browser:**

- [ ] OAuth flow works
- [ ] Form inputs functional
- [ ] Animations smooth
- [ ] API calls succeed
- [ ] localStorage works
- [ ] CSS Grid layout correct
- [ ] Fetch API works

**Known Issues to Check:**

```markdown
## Safari iOS Specific
- OAuth redirect may require user gesture
- localStorage may clear on private browsing

## Firefox Specific
- Autofill may conflict with controlled inputs
- CSS Grid gap property syntax

## Edge Specific
- None currently known
```

---

### Scenario 5: Performance Under Load

**Test Cases:**

| Scenario | Expected Time | Actual | Pass/Fail |
|----------|--------------|--------|-----------|
| Welcome page load | < 500ms | | |
| Sign In page load | < 500ms | | |
| Username availability check | < 1s | | |
| Analysis API call | < 15s | | |
| App generation | < 30s | | |
| Total onboarding | < 3 min | | |

**How to Measure:**

```javascript
// In DevTools Console
performance.mark('start');

// ... perform action ...

performance.mark('end');
performance.measure('action-time', 'start', 'end');
console.log(performance.getEntriesByName('action-time')[0].duration);
```

**Lighthouse Audit:**

```bash
# Run Lighthouse on onboarding pages
# DevTools → Lighthouse → Analyze page load

# Target scores:
# Performance: > 90
# Accessibility: 100
# Best Practices: > 90
```

---

## API Integration Tests

### Test Suite 1: Analyze Endpoint

**File:** `desktop/backend-go/internal/handler/onboarding_test.go`

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeUser_Success(t *testing.T) {
	// Setup test handler
	h := setupTestHandler(t)

	// Request payload
	payload := map[string]interface{}{
		"email":             "test@example.com",
		"gmail_connected":   true,
		"calendar_connected": false,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))

	// Execute
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	analysis := resp["analysis"].(map[string]interface{})
	assert.NotNil(t, analysis["insights"])
	assert.NotNil(t, analysis["interests"])
	assert.NotNil(t, analysis["tools_used"])
	assert.NotNil(t, analysis["profile_summary"])

	// Verify insights count
	insights := analysis["insights"].([]interface{})
	assert.Equal(t, 3, len(insights))
}

func TestAnalyzeUser_InvalidEmail(t *testing.T) {
	h := setupTestHandler(t)

	payload := map[string]interface{}{
		"email":           "invalid-email",
		"gmail_connected": true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeUser_Unauthorized(t *testing.T) {
	h := setupTestHandler(t)

	payload := map[string]interface{}{
		"email":           "test@example.com",
		"gmail_connected": true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// NO Authorization header

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnalyzeUser_MissingGmailFlag(t *testing.T) {
	h := setupTestHandler(t)

	payload := map[string]interface{}{
		"email": "test@example.com",
		// Missing gmail_connected field
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	// Should succeed with gmail_connected defaulting to false
	assert.Equal(t, http.StatusOK, w.Code)
}
```

**Run Tests:**

```bash
cd desktop/backend-go
go test ./internal/handler/... -v -run TestAnalyzeUser
```

**Expected Output:**

```
=== RUN   TestAnalyzeUser_Success
--- PASS: TestAnalyzeUser_Success (0.12s)
=== RUN   TestAnalyzeUser_InvalidEmail
--- PASS: TestAnalyzeUser_InvalidEmail (0.03s)
=== RUN   TestAnalyzeUser_Unauthorized
--- PASS: TestAnalyzeUser_Unauthorized (0.02s)
=== RUN   TestAnalyzeUser_MissingGmailFlag
--- PASS: TestAnalyzeUser_MissingGmailFlag (0.08s)
PASS
```

---

### Test Suite 2: Generate Apps Endpoint

```go
func TestGenerateApps_Success(t *testing.T) {
	h := setupTestHandler(t)

	// Create analysis data
	analysis := map[string]interface{}{
		"insights":        []string{"insight1", "insight2", "insight3"},
		"interests":       []string{"productivity", "automation"},
		"tools_used":      []string{"Figma", "Notion"},
		"profile_summary": "Test user profile",
		"raw_data":        map[string]interface{}{},
	}

	payload := map[string]interface{}{
		"workspace_id": "550e8400-e29b-41d4-a716-446655440000",
		"analysis":     analysis,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/generate-apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	apps := resp["starter_apps"].([]interface{})
	assert.Equal(t, 4, len(apps)) // Always 4 starter apps

	// Verify first app structure
	app := apps[0].(map[string]interface{})
	assert.NotEmpty(t, app["id"])
	assert.NotEmpty(t, app["title"])
	assert.NotEmpty(t, app["description"])
	assert.NotEmpty(t, app["icon_emoji"])
	assert.NotEmpty(t, app["reasoning"])
	assert.Contains(t, []string{"generating", "ready", "failed"}, app["status"])
}

func TestGenerateApps_MissingWorkspaceID(t *testing.T) {
	h := setupTestHandler(t)

	payload := map[string]interface{}{
		// Missing workspace_id
		"analysis": map[string]interface{}{
			"insights": []string{"test"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/osa-onboarding/generate-apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

---

### Test Suite 3: Username Check Endpoint

```bash
#!/bin/bash
# test-username-api.sh

BASE_URL="http://localhost:5000"

echo "Testing Username Availability API"
echo "=================================="

# Test 1: Valid available username
echo -n "Test 1: Valid username (should be available)... "
RESP=$(curl -s "$BASE_URL/api/users/check-username/validuser123")
AVAILABLE=$(echo $RESP | jq -r '.available')
if [ "$AVAILABLE" == "true" ]; then
  echo "✓ PASS"
else
  echo "✗ FAIL: $RESP"
fi

# Test 2: Reserved username
echo -n "Test 2: Reserved username 'admin' (should be taken)... "
RESP=$(curl -s "$BASE_URL/api/users/check-username/admin")
AVAILABLE=$(echo $RESP | jq -r '.available')
if [ "$AVAILABLE" == "false" ]; then
  echo "✓ PASS"
else
  echo "✗ FAIL: $RESP"
fi

# Test 3: Too short
echo -n "Test 3: Too short 'ab' (should fail)... "
RESP=$(curl -s "$BASE_URL/api/users/check-username/ab")
STATUS=$(echo $RESP | jq -r '.error')
if [ "$STATUS" != "null" ]; then
  echo "✓ PASS"
else
  echo "✗ FAIL: Should have returned error"
fi

# Test 4: Invalid characters
echo -n "Test 4: Invalid chars 'user!name' (should fail)... "
RESP=$(curl -s "$BASE_URL/api/users/check-username/user!name")
STATUS=$(echo $RESP | jq -r '.error')
if [ "$STATUS" != "null" ]; then
  echo "✓ PASS"
else
  echo "✗ FAIL: Should have returned error"
fi

# Test 5: Valid with underscores
echo -n "Test 5: Valid 'user_name_123' (should be available)... "
RESP=$(curl -s "$BASE_URL/api/users/check-username/user_name_123")
AVAILABLE=$(echo $RESP | jq -r '.available')
if [ "$AVAILABLE" == "true" ]; then
  echo "✓ PASS"
else
  echo "✗ FAIL: $RESP"
fi

echo "=================================="
echo "Username API tests complete"
```

**Run:**

```bash
chmod +x test-username-api.sh
./test-username-api.sh
```

---

## Automated Test Recommendations

### E2E Test Suite with Playwright

**File:** `frontend/tests/e2e/onboarding.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('OSA Onboarding Flow', () => {
	test.beforeEach(async ({ page }) => {
		await page.context().clearCookies();
		await page.goto('/onboarding');
	});

	test('complete happy path onboarding', async ({ page }) => {
		// Screen 1: Welcome
		await expect(page.locator('h1')).toContainText('Welcome to OSA Build');
		await page.click('button:has-text("Get Started")');

		// Screen 2: Meet OSA
		await expect(page).toHaveURL(/meet-osa/);
		await page.click('button:has-text("Next")');

		// Screen 3: Sign In
		await expect(page).toHaveURL(/signin/);
		await page.fill('input[type="email"]', 'playwright@test.com');
		await page.click('button:has-text("Continue")');

		// Screen 4: Gmail (Skip)
		await expect(page).toHaveURL(/gmail/);
		await page.click('text=Skip');

		// Screen 5: Username
		await expect(page).toHaveURL(/username/);
		const username = `test_${Date.now()}`;
		await page.fill('input[placeholder*="bekorains"]', username);
		await page.click('text=Check Availability');
		await page.waitForSelector('text=/Available/i');
		await page.click('button:has-text("Continue")');

		// Screen 6-8: Analyzing
		await expect(page).toHaveURL(/analyzing/);

		// Wait for analysis to complete (max 20s)
		await page.waitForURL(/starter-apps/, { timeout: 20000 });

		// Screen 9: Starter Apps
		await expect(page).toHaveURL(/starter-apps/);

		// Wait for all apps to be ready (max 30s)
		await page.waitForSelector('button:has-text("Ready to Launch")', {
			timeout: 30000
		});

		// Verify 4 apps displayed
		const appCards = page.locator('[data-testid="app-card"]');
		await expect(appCards).toHaveCount(4);

		// Click Ready to Launch
		await page.click('button:has-text("Ready to Launch")');

		// Screen 10: Ready
		await expect(page).toHaveURL(/ready/);
		await expect(page.locator('h1')).toContainText(/ready|complete/i);
	});

	test('skip gmail connection', async ({ page }) => {
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'nogmail@test.com');
		await page.click('button:has-text("Continue")');

		// Skip Gmail
		await page.click('text=Skip');

		// Should continue to username
		await expect(page).toHaveURL(/username/);

		// Verify gmail_connected is false in localStorage
		const stored = await page.evaluate(() => {
			const data = localStorage.getItem('onboarding');
			return data ? JSON.parse(data) : null;
		});

		expect(stored?.userData?.gmailConnected).toBe(false);
	});

	test('back navigation preserves state', async ({ page }) => {
		// Navigate forward to Username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'back@test.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Now on Username screen
		await expect(page).toHaveURL(/username/);

		// Navigate back
		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/gmail/);

		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/signin/);

		// Verify email still filled
		const emailValue = await page.inputValue('input[type="email"]');
		expect(emailValue).toBe('back@test.com');
	});

	test('network error recovery', async ({ page }) => {
		// Navigate to username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'offline@test.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Go offline
		await page.context().setOffline(true);

		// Try to check username
		await page.fill('input[placeholder*="bekorains"]', 'testuser');
		await page.click('text=Check Availability');

		// Should show error
		await expect(page.locator('text=/network|offline|error/i')).toBeVisible();

		// Go back online
		await page.context().setOffline(false);

		// Retry
		await page.click('button:has-text("Retry")');

		// Should succeed
		await expect(page.locator('text=/available/i')).toBeVisible();
	});

	test('username validation', async ({ page }) => {
		// Navigate to username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'validation@test.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Test too short
		await page.fill('input[placeholder*="bekorains"]', 'ab');
		await page.click('text=Check Availability');
		await expect(page.locator('text=/at least 3 characters/i')).toBeVisible();

		// Test invalid characters
		await page.fill('input[placeholder*="bekorains"]', 'user!name');
		await page.click('text=Check Availability');
		await expect(page.locator('text=/invalid/i')).toBeVisible();

		// Test reserved username
		await page.fill('input[placeholder*="bekorains"]', 'admin');
		await page.click('text=Check Availability');
		await expect(page.locator('text=/taken|unavailable/i')).toBeVisible();
	});

	test('browser refresh resumes onboarding', async ({ page, context }) => {
		// Start onboarding
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'resume@test.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');
		await page.fill('input[placeholder*="bekorains"]', `resume_${Date.now()}`);
		await page.click('text=Check Availability');
		await page.waitForSelector('text=/available/i');

		// Get current localStorage state
		const beforeRefresh = await page.evaluate(() => localStorage.getItem('onboarding'));

		// Refresh page
		await page.reload();

		// Should resume at same step
		await expect(page).toHaveURL(/username/);

		// Verify state preserved
		const afterRefresh = await page.evaluate(() => localStorage.getItem('onboarding'));
		expect(afterRefresh).toBe(beforeRefresh);
	});
});
```

**Run E2E Tests:**

```bash
cd frontend
npx playwright install # First time only
npx playwright test

# Or run in UI mode
npx playwright test --ui

# Run specific test
npx playwright test -g "happy path"

# Debug mode
npx playwright test --debug
```

**Expected Output:**

```
Running 7 tests using 1 worker

  ✓ complete happy path onboarding (35s)
  ✓ skip gmail connection (8s)
  ✓ back navigation preserves state (12s)
  ✓ network error recovery (10s)
  ✓ username validation (15s)
  ✓ browser refresh resumes onboarding (10s)

  7 passed (1.5m)
```

---

## Performance & Load Testing

### Performance Benchmarks

| Metric | Target | Measurement Method | Priority |
|--------|--------|-------------------|----------|
| Welcome page load | < 500ms | Lighthouse, Network tab | P0 |
| Username availability | < 1s | Network tab (API call time) | P0 |
| Analysis API | < 15s | Network tab (total time) | P0 |
| App generation | < 30s | Apps status polling | P0 |
| Total onboarding | < 3 min | Stopwatch (manual) | P0 |
| Animation FPS | > 60fps | DevTools Performance tab | P1 |
| Memory usage | < 50MB | DevTools Memory tab | P2 |

### How to Measure Performance

**1. Lighthouse Audit:**

```bash
# CLI method
npm install -g lighthouse

lighthouse http://localhost:5173/onboarding \
  --output html \
  --output-path ./lighthouse-report.html \
  --chrome-flags="--headless"

# Or in DevTools:
# 1. Open DevTools (F12)
# 2. Go to Lighthouse tab
# 3. Select "Desktop" device
# 4. Click "Analyze page load"
```

**Target Scores:**
- Performance: ≥ 90
- Accessibility: 100
- Best Practices: ≥ 90
- SEO: ≥ 80

**2. Network Performance:**

```javascript
// In DevTools Console
// Measure API call time
performance.mark('api-start');

fetch('/api/osa-onboarding/analyze', { /* ... */ })
  .then(() => {
    performance.mark('api-end');
    performance.measure('api-time', 'api-start', 'api-end');
    const measure = performance.getEntriesByName('api-time')[0];
    console.log(`API took ${measure.duration}ms`);
  });
```

**3. Animation Performance:**

```javascript
// DevTools → Performance
// 1. Click Record
// 2. Navigate through onboarding screens
// 3. Stop recording
// 4. Analyze:
//    - FPS graph (should be green, above 60fps)
//    - Main thread activity (should be minimal)
//    - Memory usage (should be stable)
```

**4. Memory Leak Detection:**

```javascript
// DevTools → Memory → Take heap snapshot
// 1. Take snapshot before onboarding
// 2. Complete onboarding
// 3. Take snapshot after
// 4. Compare:
//    - Memory should decrease after onboarding complete
//    - No detached DOM nodes
//    - Event listeners cleaned up
```

### Load Testing with k6

**File:** `tests/load/onboarding.js`

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '10s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests < 2s
    http_req_failed: ['rate<0.01'],    // <1% failure rate
  },
};

const BASE_URL = 'http://localhost:5000';

export default function () {
  // Test 1: Check username availability
  const usernameResp = http.get(`${BASE_URL}/api/users/check-username/testuser${__VU}`);
  check(usernameResp, {
    'username check status 200': (r) => r.status === 200,
    'username check < 1s': (r) => r.timings.duration < 1000,
  });

  sleep(1);

  // Test 2: Analyze user
  const analyzePayload = JSON.stringify({
    email: `user${__VU}@test.com`,
    gmail_connected: false,
  });

  const analyzeResp = http.post(`${BASE_URL}/api/osa-onboarding/analyze`, analyzePayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.TEST_TOKEN}`,
    },
  });

  check(analyzeResp, {
    'analyze status 200': (r) => r.status === 200,
    'analyze < 15s': (r) => r.timings.duration < 15000,
    'analyze returns insights': (r) => {
      const body = JSON.parse(r.body);
      return body.analysis && body.analysis.insights.length === 3;
    },
  });

  sleep(2);
}
```

**Run Load Test:**

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io/docs/get-started/installation/

# Set test token
export TEST_TOKEN="your-jwt-token-here"

# Run test
k6 run tests/load/onboarding.js

# Expected output:
# ✓ username check status 200
# ✓ username check < 1s
# ✓ analyze status 200
# ✓ analyze < 15s
# ✓ analyze returns insights
#
# checks.........................: 100.00% ✓ 500  ✗ 0
# http_req_duration..............: avg=1.2s  p(95)=1.8s
# http_req_failed................: 0.00%   ✓ 0    ✗ 500
```

---

## Commands & Quick Reference

### Setup Commands

```bash
# Start backend
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/main.go

# Start frontend
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev

# Run migrations
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/migrate

# Seed test data
psql $DATABASE_URL -f seed_test_data.sql
```

### Testing Commands

```bash
# Backend unit tests
cd desktop/backend-go
go test ./... -v

# Frontend unit tests
cd frontend
npm test

# E2E tests
cd frontend
npx playwright test

# Load tests
k6 run tests/load/onboarding.js

# Username API test
./test-username-api.sh

# Lighthouse audit
lighthouse http://localhost:5173/onboarding --view
```

### Database Verification

```sql
-- Check onboarding profiles
SELECT
  wp.workspace_id,
  u.email,
  wp.analysis_data->>'profile_summary' as summary,
  jsonb_array_length(wp.starter_apps_data) as app_count,
  wp.created_at
FROM workspace_onboarding_profiles wp
JOIN users u ON wp.user_id = u.id
ORDER BY wp.created_at DESC
LIMIT 10;

-- Check starter apps status
SELECT
  workspace_id,
  app->>'title' as app_title,
  app->>'status' as status,
  app->>'category' as category
FROM workspace_onboarding_profiles,
jsonb_array_elements(starter_apps_data) as app
WHERE workspace_id = '550e8400-e29b-41d4-a716-446655440000';

-- Clean test data
DELETE FROM workspace_onboarding_profiles
WHERE user_id IN (
  SELECT id FROM users WHERE email LIKE '%@test.com'
);
```

### DevTools Inspection

```javascript
// Check localStorage state
const state = JSON.parse(localStorage.getItem('onboarding'));
console.log('Current step:', state.currentStep);
console.log('User data:', state.userData);
console.log('Starter apps:', state.userData.starterApps);

// Clear localStorage (reset onboarding)
localStorage.removeItem('onboarding');
location.reload();

// Monitor API calls
performance.getEntriesByType('resource')
  .filter(r => r.name.includes('/api/'))
  .forEach(r => console.log(`${r.name}: ${r.duration}ms`));
```

---

## Bug Report Template

### Issue Report Format

**Title:** [Screen] - [Brief description]

**Example:** `Username Screen - Availability check fails with special characters`

---

**Severity:**
- [ ] Critical (blocks onboarding completion)
- [ ] High (major feature broken)
- [ ] Medium (workaround available)
- [ ] Low (cosmetic issue)

---

**Environment:**
- **OS:** macOS 13.5 / Windows 11 / Ubuntu 22.04
- **Browser:** Chrome 120.0.6099.109
- **Device:** MacBook Pro M1 / iPhone 14 / Desktop PC
- **Build:** `feature/ios-desktop-flow-migration` @ commit `abc123`
- **Date:** 2026-01-18

---

**Steps to Reproduce:**

1. Start at `/onboarding`
2. Complete through Username screen
3. Enter username: `test_user!123`
4. Click "Check Availability"
5. Observe: Error message or unexpected behavior

---

**Expected Behavior:**

Should show error: "Invalid characters. Use only letters, numbers, and underscores."

---

**Actual Behavior:**

Shows generic error: "Something went wrong. Try again."

---

**Screenshots/Videos:**

[Attach screenshot or screen recording]

---

**Console Errors:**

```
TypeError: Cannot read property 'available' of undefined
    at checkUsername (username.svelte:45)
    at onClick (username.svelte:78)
```

---

**Network Errors:**

```
GET /api/users/check-username/test_user!123
Status: 400 Bad Request
Response: { "error": "invalid username format" }
```

---

**Additional Context:**

- Does this happen consistently? **Yes**
- Workaround discovered? **No**
- Related to which feature? **Username validation**
- First seen after which change? **After adding special char validation**

---

**Priority:**

- Blocking release? **No**
- Affects core functionality? **Yes**
- Impacts user experience? **Yes**

---

### Example Bug Reports

**Bug #1: Username check returns 500 on reserved names**

```markdown
**Title:** Username Screen - 500 error when checking reserved usernames

**Severity:** High

**Environment:**
- OS: macOS 13.5
- Browser: Chrome 120
- Build: feature/ios-desktop-flow-migration @ 894f0f7

**Steps:**
1. Navigate to username screen
2. Enter "admin"
3. Click "Check Availability"
4. Error: 500 Internal Server Error

**Expected:**
Should return 200 with `{ "available": false }`

**Actual:**
Returns 500 error

**Console:**
```
POST /api/users/check-username/admin
Status: 500
Response: { "error": "Internal server error" }
```

**Fix:**
Backend should handle reserved usernames before database query.
```

---

## Appendix

### A. Reserved Usernames List

```
admin, root, osa, test, support, system, api, www, mail, ftp,
administrator, webmaster, postmaster, hostmaster, info, help,
user, guest, demo, example, null, undefined, localhost
```

### B. Validation Rules Reference

**Email:**
- RFC 5322 compliant
- Max 254 characters
- Must contain @ and valid domain

**Username:**
- 3-30 characters
- Alphanumeric and underscores only
- Must start with letter
- Case-insensitive matching
- Cannot be reserved word

**Workspace ID:**
- Valid UUID v4 format
- Example: `550e8400-e29b-41d4-a716-446655440000`

### C. API Response Codes

| Code | Meaning | Action |
|------|---------|--------|
| 200 | OK | Request successful |
| 400 | Bad Request | Fix request payload |
| 401 | Unauthorized | Refresh auth token |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate username/email |
| 422 | Validation Error | Fix input format |
| 500 | Server Error | Check backend logs |
| 503 | Service Unavailable | Retry later |

### D. State Machine Diagram

```
┌─────────────────────────────────────────────────────────┐
│ Onboarding State Machine                                 │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  START                                                    │
│    ↓                                                      │
│  [Welcome] ──────────────────────────────────────────┐   │
│    ↓                                                  │   │
│  [Meet OSA]                                           │   │
│    ↓                                                  │   │
│  [Sign In] ← Can go back ────────────────────────────┘   │
│    ↓                                                      │
│  [Gmail Connect] ← Can skip                               │
│    ↓                                                      │
│  [Username] ← Must succeed                                │
│    ↓                                                      │
│  [Analyzing 1/3] ← Auto-advance                           │
│    ↓                                                      │
│  [Analyzing 2/3] ← Auto-advance                           │
│    ↓                                                      │
│  [Analyzing 3/3] ← Auto-advance                           │
│    ↓                                                      │
│  [Starter Apps] ← Poll until ready                        │
│    ↓                                                      │
│  [Ready] ← Final confirmation                             │
│    ↓                                                      │
│  COMPLETE → Main App                                      │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### E. Test Coverage Goals

| Layer | Target | Actual | Status |
|-------|--------|--------|--------|
| E2E Tests | 100% critical paths | TBD | 🟡 |
| Integration Tests | 80% API endpoints | TBD | 🟡 |
| Unit Tests | 70% functions | TBD | 🟡 |
| Manual Tests | 100% user flows | TBD | 🟡 |

### F. Useful Resources

**Documentation:**
- API Docs: `/docs/API_OSA_ONBOARDING.md`
- Flow Design: `/docs/OSA_BUILD_ONBOARDING_FLOW.md`
- Testing Docs: `/docs/TESTING_OSA_ONBOARDING.md`

**Tools:**
- Playwright: https://playwright.dev/
- k6: https://k6.io/docs/
- Lighthouse: https://developers.google.com/web/tools/lighthouse

**Contact:**
- QA Lead: [Your name]
- Backend: Pedro
- Frontend: Nejd/Javaris
- Product: Roberto

---

## Final Checklist

Before marking testing complete:

- [ ] All P0 scenarios pass
- [ ] E2E tests written and passing
- [ ] Backend unit tests passing
- [ ] Performance benchmarks met
- [ ] Accessibility audit passed
- [ ] Cross-browser testing complete
- [ ] Load testing complete
- [ ] Bug reports filed for any issues
- [ ] Test results documented
- [ ] Sign-off from QA Lead
- [ ] Sign-off from Engineering Lead

---

**Document Version:** 2.0
**Last Updated:** 2026-01-18
**Status:** Ready for QA Execution
**Next Review:** After test run completion
