# OSA Build Onboarding Flow - Comprehensive Integration Testing Plan

**Document Version:** 1.0
**Created:** 2026-01-18
**Last Updated:** 2026-01-18
**Status:** Active

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Testing Strategy](#testing-strategy)
3. [Test Cases by Screen](#test-cases-by-screen)
4. [API Test Cases](#api-test-cases)
5. [State Management Tests](#state-management-tests)
6. [User Flows to Test](#user-flows-to-test)
7. [Performance Tests](#performance-tests)
8. [Accessibility Tests](#accessibility-tests)
9. [Browser Compatibility](#browser-compatibility)
10. [Manual Testing Checklist](#manual-testing-checklist)
11. [Automated Test Scripts](#automated-test-scripts)
12. [Test Data](#test-data)
13. [Bug Report Template](#bug-report-template)

---

## Executive Summary

The OSA Build onboarding flow guides new users through a personalized setup experience across 11 screens:

1. Welcome to OSA Build
2. Meet OSA
3. Sign In
4. Connect Gmail
5. Claim Username
6. Analyzing (Screens 1-3)
7. Starter Apps
8. Ready to Launch
9. Additional screens (ready, starter-apps variations)

This testing plan ensures:
- All user flows complete successfully
- API integrations handle errors gracefully
- State persists correctly across navigation
- Performance meets expectations
- Accessibility standards are met
- No regressions between deployments

---

## Testing Strategy

### 1. Unit Tests (Backend Services)

**Scope:** Individual Go handler functions and service methods

**Tools:** Go testing standard library + testify/assert

**Coverage Target:** >80% for critical paths

**Components to test:**
- User analysis service
- App generation service
- Username availability check
- Profile retrieval service
- Error handling and validation

### 2. Integration Tests (API Endpoints)

**Scope:** End-to-end API flows with database

**Tools:** Go testing + test fixtures + testable database

**Coverage Target:** All happy paths + error scenarios

**Endpoints to test:**
- POST `/api/osa-onboarding/analyze`
- POST `/api/osa-onboarding/generate-apps`
- GET `/api/osa-onboarding/apps-status`
- GET `/api/osa-onboarding/profile`
- POST `/api/auth/username-check` (if separate)

### 3. E2E Tests (Frontend Flow)

**Scope:** Complete user journey through UI

**Tools:** Playwright + SvelteKit integration

**Coverage Target:** All critical user paths

**Flows to test:**
- Happy path: Welcome → Meet OSA → SignIn → Gmail → Username → Analyzing → Apps → Ready
- Alternative paths: Skip Gmail, go back, resume session
- Error handling: API failures, invalid input, network errors

### 4. Manual Testing Checklist

**Scope:** Visual QA, accessibility, user experience

**Frequency:** Before every production release

**Includes:** Visual inspection, keyboard navigation, screen reader testing

---

## Test Cases by Screen

### Screen 1: Welcome to OSA Build

#### Happy Path
- [ ] Page loads successfully
- [ ] Title displays correctly: "Welcome to OSA Build"
- [ ] Hero image/gradient displays
- [ ] Three feature cards display (AI-Powered, Day 1 Apps, Fully Customizable)
- [ ] "Get Started" button is clickable and prominent
- [ ] "Takes less than 2 minutes" text displays

#### Navigation
- [ ] Clicking "Get Started" navigates to `/onboarding/meet-osa`
- [ ] onboardingStore.currentStep updates to 2
- [ ] Animation slide-up plays smoothly

#### Edge Cases
- [ ] Long page title doesn't break layout
- [ ] Very small screens (<320px) still display content
- [ ] Very large screens (>1920px) content remains centered
- [ ] Dark mode styling applies correctly

#### Responsive Design
- [ ] Mobile (375px): Single column layout
- [ ] Tablet (768px): 2-3 column grid
- [ ] Desktop (1440px+): Full 3-column grid

#### Error Scenarios
- [ ] Navigation works with JS disabled (graceful degradation)
- [ ] Store initialization handles errors

#### Performance
- [ ] Page load: <1.5s on 3G
- [ ] Animation frame rate: >60fps
- [ ] No layout shift (CLS)

---

### Screen 2: Meet OSA

#### Happy Path
- [ ] Page loads with introduction copy
- [ ] OSA persona message displays correctly
- [ ] "Get Started" / "Next" button displays
- [ ] Back navigation available

#### Navigation
- [ ] "Next" button → `/onboarding/signin`
- [ ] "Back" button → `/onboarding` (Screen 1)
- [ ] Store tracking updates currentStep

#### Content Validation
- [ ] All copy renders without truncation
- [ ] Links (if any) are clickable and properly styled

#### Error Scenarios
- [ ] Handles missing copy from backend gracefully
- [ ] Navigation works if store fails to update

#### Accessibility
- [ ] Proper heading hierarchy (h1, p)
- [ ] Button text is descriptive
- [ ] Color contrast meets WCAG AA

---

### Screen 3: Sign In

#### Happy Path
- [ ] Email input field displays
- [ ] Email validation works (no submit until valid email)
- [ ] "Continue" button enabled only for valid email
- [ ] Form submission works
- [ ] User authenticated successfully
- [ ] Redirects to Gmail connection screen

#### Input Validation
- [ ] Accepts valid emails: user@example.com, test+tag@domain.co.uk
- [ ] Rejects invalid emails: test@, @domain.com, no-at-sign
- [ ] Rejects empty input
- [ ] Shows error message for invalid format
- [ ] Shows error message for duplicate account
- [ ] Shows error message for API errors

#### Edge Cases
- [ ] Very long email addresses (254 chars max)
- [ ] International domain names
- [ ] Gmail spam check (if applicable)
- [ ] Sign in while already authenticated

#### Error Scenarios
- [ ] API timeout: Shows error, allows retry
- [ ] 400 Bad Request: Shows validation error
- [ ] 409 Conflict: "Account already exists" message
- [ ] 500 Server Error: "Try again" message
- [ ] Network offline: Appropriate error message

#### Performance
- [ ] Email validation is instant (<100ms)
- [ ] Form submission: <2s response time
- [ ] Smooth focus states and interactions

#### Accessibility
- [ ] Email input has proper label
- [ ] Error messages associated with input (aria-describedby)
- [ ] Tab navigation works correctly
- [ ] Focus visible on buttons

---

### Screen 4: Connect Gmail

#### Happy Path
- [ ] Gmail OAuth button displays
- [ ] Clicking OAuth button opens Google login
- [ ] OAuth callback redirects back to app
- [ ] Gmail connection confirmation displays
- [ ] "Skip" option available
- [ ] "Continue" button shows after connection/skip

#### OAuth Flow
- [ ] Correct OAuth scopes requested (email, calendar, contacts)
- [ ] Successfully stores Gmail token
- [ ] Handles OAuth cancellation (user clicks "Cancel")
- [ ] Handles OAuth denial (user denies permissions)

#### Skip Path
- [ ] Skipping Gmail works without errors
- [ ] onboardingStore.gmail_connected = false
- [ ] Flow continues to next screen normally

#### Error Scenarios
- [ ] OAuth endpoint unreachable: Shows error with retry
- [ ] Token storage fails: Shows error
- [ ] Network error during redirect: Shows error
- [ ] Google account not found: Shows appropriate error

#### Accessibility
- [ ] OAuth button has clear label
- [ ] Skip link/button properly labeled
- [ ] Focus management after OAuth callback

#### Performance
- [ ] OAuth loads in <2s
- [ ] No unnecessary re-renders after callback

---

### Screen 5: Claim Username

#### Happy Path
- [ ] Username input field displays
- [ ] Placeholder text "bekorains" shows
- [ ] "Check Availability" button works
- [ ] Username availability API call succeeds
- [ ] Success message shows: "✓ Available!"
- [ ] "Continue" button enabled after success
- [ ] Continue button navigates to analyzing screen
- [ ] Username saved to store

#### Username Validation
- [ ] Minimum 3 characters required
- [ ] Letters, numbers, underscores allowed
- [ ] Hyphens allowed (if specified)
- [ ] Special characters rejected with error
- [ ] Spaces rejected with error
- [ ] Case-insensitive check

#### Availability Check
- [ ] API called with sanitized input
- [ ] Reserved usernames rejected: admin, root, osa, test
- [ ] Case-insensitive: "Admin" and "admin" both taken
- [ ] Proper error message: "Username is already taken"
- [ ] Loading state shows during check
- [ ] Results cached (no redundant API calls)

#### Edge Cases
- [ ] Very long usernames (255+ chars)
- [ ] Unicode characters attempted
- [ ] Rapid successive availability checks (debounce)
- [ ] Network flakiness during check
- [ ] Going back after successful check

#### Error Scenarios
- [ ] Availability API timeout: Shows error + retry option
- [ ] 400 Bad Request: "Invalid username format"
- [ ] 500 Server Error: "Try again later"
- [ ] Network offline: Appropriate error

#### User Experience
- [ ] Real-time feedback on input change
- [ ] Clear validation messages
- [ ] Continue button disabled until available
- [ ] Back button works correctly

#### Accessibility
- [ ] Username input has proper label
- [ ] Helper text for valid username
- [ ] Error messages associated with input
- [ ] Loading indicator announced
- [ ] Tab navigation works

---

### Screen 6: Analyzing (Step 1/3)

#### Happy Path
- [ ] Page loads with analysis starting message
- [ ] Progress indicator shows "1/3"
- [ ] Loading animation plays smoothly
- [ ] Call to `/api/osa-onboarding/analyze` is made automatically
- [ ] Wait time: 5-15 seconds (show appropriate messaging)
- [ ] Transitions to Screen 7 after analysis completes

#### API Call
- [ ] Email included in request
- [ ] gmail_connected status included
- [ ] Handles streaming response correctly
- [ ] Updates UI with real-time insights

#### Error Scenarios
- [ ] API timeout (>30s): Shows "Taking longer than expected" + retry
- [ ] 400 Bad Request: Shows error message
- [ ] 500 Server Error: Shows error with retry option
- [ ] Network error: Shows offline message
- [ ] User navigates away: Cancels request gracefully

#### Performance
- [ ] Loading animation: >60fps
- [ ] No jank during analysis
- [ ] Memory doesn't leak during long wait

#### Accessibility
- [ ] Progress indicator announced (e.g., "Analyzing your profile, step 1 of 3")
- [ ] Loading animation has aria-busy=true
- [ ] Focus visible if interactive elements appear

#### UI States
- [ ] Visual feedback that something is happening
- [ ] Clear messaging about what's being analyzed
- [ ] Estimated time remaining (if available)

---

### Screen 7: Analyzing (Step 2/3)

#### Same as Screen 6
- [ ] Progress indicator shows "2/3"
- [ ] Analysis continues
- [ ] Smooth transition to Screen 8

---

### Screen 8: Analyzing (Step 3/3)

#### Same as Screen 6
- [ ] Progress indicator shows "3/3"
- [ ] Final analysis step
- [ ] Transitions to Screen 9 (Starter Apps)
- [ ] All analysis data ready

---

### Screen 9: Starter Apps

#### Happy Path
- [ ] Page loads with 4 starter app cards
- [ ] Each app displays: title, description, icon, reasoning
- [ ] App generation status shows (generating/ready/failed)
- [ ] Loading indicators for apps still generating
- [ ] "Ready to Launch" button shown when all apps ready
- [ ] Clicking app shows more details (if implemented)
- [ ] Visual hierarchy clear

#### App Cards Display
- [ ] App title visible and readable
- [ ] App icon displays correctly (emoji or URL)
- [ ] Category badge shows (e.g., "Productivity")
- [ ] Description truncates gracefully (max 3 lines)
- [ ] Reasoning shows for context (if space available)

#### App Generation Status
- [ ] 'generating': Shows loading spinner
- [ ] 'ready': Shows checkmark or success indicator
- [ ] 'failed': Shows error state + retry option
- [ ] Status updates in real-time

#### API Integration
- [ ] GET `/api/osa-onboarding/apps-status` polls periodically
- [ ] Poll interval: 2-3 seconds while generating
- [ ] Stops polling when all apps ready
- [ ] Handles partial failures gracefully

#### Error Scenarios
- [ ] Some apps fail to generate: Shows error on affected cards
- [ ] All apps fail: Shows error + retry all button
- [ ] API timeout: Shows error + manual retry
- [ ] Network error: Shows offline indicator
- [ ] User leaves page during generation: Resumes on return

#### Accessibility
- [ ] App cards have semantic structure
- [ ] Status updates announced via aria-live
- [ ] Buttons properly labeled
- [ ] Keyboard navigation works through cards
- [ ] Focus management clear

#### Responsive Design
- [ ] Mobile (375px): 1-column layout
- [ ] Tablet (768px): 2-column grid
- [ ] Desktop (1440px): 4-column grid or 2x2

#### Performance
- [ ] Initial load: <2s
- [ ] Status polling doesn't cause UI jank
- [ ] Animations smooth (<60fps)
- [ ] Memory stable during long polling

---

### Screen 10: Ready to Launch

#### Happy Path
- [ ] Page displays success message
- [ ] Summary of user profile shows
- [ ] Summary of 4 starter apps shows
- [ ] "Enter OSA Build" / "Launch" button prominent
- [ ] Clicking launches main app

#### Content Display
- [ ] User's username displayed
- [ ] Profile summary accurate (based on analysis)
- [ ] All 4 apps listed with proper metadata
- [ ] Visual celebration/success theme

#### Navigation
- [ ] "Launch" button → Main app dashboard
- [ ] Back button → Review previous steps (if needed)
- [ ] Direct URL access works if onboarding complete

#### Error Scenarios
- [ ] Profile data fails to load: Shows error + retry
- [ ] Apps data missing: Shows partial success
- [ ] Launch button fails: Shows error + retry

#### User Experience
- [ ] Celebration/success tone
- [ ] Smooth transition to main app
- [ ] Onboarding mark complete in database

#### Accessibility
- [ ] Success confirmed in text (not just visual)
- [ ] Summary content properly structured
- [ ] Focus moves to launch button

#### Performance
- [ ] Page loads quickly: <1s
- [ ] No layout shift when content loads

---

## API Test Cases

### Endpoint: POST `/api/osa-onboarding/analyze`

**Request Body:**
```json
{
  "email": "user@example.com",
  "gmail_connected": true,
  "calendar_connected": false
}
```

**Response (200 OK):**
```json
{
  "analysis": {
    "insights": ["string"],
    "interests": ["string"],
    "tools_used": ["string"],
    "profile_summary": "string",
    "raw_data": {}
  }
}
```

#### Test Cases

| # | Scenario | Input | Expected Status | Expected Behavior |
|---|----------|-------|-----------------|-------------------|
| 1 | Happy path | Valid email, gmail=true | 200 | Returns analysis object |
| 2 | Gmail not connected | Valid email, gmail=false | 200 | Returns analysis (less detailed) |
| 3 | Missing email | {} | 400 | Error: "email required" |
| 4 | Invalid email | {email: "invalid"} | 400 | Error: "Invalid email format" |
| 5 | Unauthorized | No auth token | 401 | Error: "Unauthorized" |
| 6 | Timeout | Long processing | 504 | Error: "Request timeout" |
| 7 | Server error | Any | 500 | Error: "Internal server error" |
| 8 | Gmail token invalid | gmail=true, invalid token | 400 | Error: "Gmail connection failed" |
| 9 | Concurrent requests | 2+ simultaneous | 200 | Each returns separate analysis |
| 10 | Large email lists | email with +aliases | 200 | Handles correctly |

#### Code Examples

```go
// Unit test
func TestAnalyzeUser_HappyPath(t *testing.T) {
    service := NewOSAOnboardingService(mockDB, mockGmail)

    req := AnalyzeUserRequest{
        Email: "user@example.com",
        GmailConnected: true,
    }

    result, err := service.AnalyzeUser(context.Background(), req)

    assert.NoError(t, err)
    assert.NotEmpty(t, result.Insights)
    assert.NotEmpty(t, result.ProfileSummary)
}

// Integration test
func TestAnalyzeEndpoint(t *testing.T) {
    router := setupTestRouter()

    body := `{"email":"user@example.com","gmail_connected":true}`
    req, _ := http.NewRequest("POST", "/api/osa-onboarding/analyze",
        strings.NewReader(body))

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp AnalyzeUserResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotNil(t, resp.Analysis)
}
```

---

### Endpoint: POST `/api/osa-onboarding/generate-apps`

**Request Body:**
```json
{
  "workspace_id": "workspace-123",
  "analysis": {
    "insights": ["..."],
    "interests": ["..."],
    "tools_used": ["..."],
    "profile_summary": "..."
  }
}
```

**Response (200 OK):**
```json
{
  "starter_apps": [
    {
      "id": "app-1",
      "title": "App Title",
      "description": "Description",
      "icon_emoji": "🎯",
      "icon_url": "https://...",
      "reasoning": "Why this app",
      "category": "Productivity",
      "status": "ready",
      "workflow_id": "workflow-123"
    }
  ],
  "ready_to_launch": true
}
```

#### Test Cases

| # | Scenario | Expected | Notes |
|---|----------|----------|-------|
| 1 | Valid analysis | 200, 4 apps returned | All apps should be generated |
| 2 | Missing workspace_id | 400 | Validation error |
| 3 | Invalid analysis object | 400 | Validation error |
| 4 | Workspace not found | 404 | "Workspace not found" |
| 5 | App generation fails | 500 | Error message |
| 6 | Partial failure | 200 | Some apps marked as failed |
| 7 | Concurrent requests | 200 | Each gets unique apps (or same cached) |
| 8 | Very large analysis | 200 | Handles without timeout |
| 9 | Empty analysis | 400 | Validation error |
| 10 | Rerun generation | 200 | Returns new apps (overwrites old) |

#### Code Example

```go
func TestGenerateApps_HappyPath(t *testing.T) {
    service := NewOSAOnboardingService(mockDB, mockAI)

    analysis := UserAnalysisResult{
        Insights: []string{"insight1"},
        Interests: []string{"productivity"},
        ToolsUsed: []string{"email"},
        ProfileSummary: "Test profile",
    }

    req := GenerateAppsRequest{
        WorkspaceID: "ws-123",
        Analysis: analysis,
    }

    result, err := service.GenerateApps(context.Background(), req)

    assert.NoError(t, err)
    assert.Len(t, result.StarterApps, 4)
    assert.True(t, result.ReadyToLaunch)

    for _, app := range result.StarterApps {
        assert.NotEmpty(t, app.Title)
        assert.NotEmpty(t, app.Description)
    }
}
```

---

### Endpoint: GET `/api/osa-onboarding/apps-status?workspace_id=xxx`

**Response (200 OK):**
```json
{
  "analysis": { ... },
  "starter_apps": [ ... ],
  "ready_to_launch": false
}
```

#### Test Cases

| # | Scenario | Expected |
|---|----------|----------|
| 1 | All apps ready | Status: ready for all, ready_to_launch: true |
| 2 | Some apps generating | Status mix, ready_to_launch: false |
| 3 | All generating | Status: generating for all, ready_to_launch: false |
| 4 | Some failed | Status: failed for failed apps, others ready |
| 5 | Missing workspace_id | 400 Bad Request |
| 6 | Invalid workspace_id | 404 Not Found |
| 7 | No apps generated yet | 404 or empty array |
| 8 | Concurrent polls | 200, consistent data |
| 9 | After completion | ready_to_launch: true |

---

### Endpoint: GET `/api/osa-onboarding/profile?workspace_id=xxx`

**Response (200 OK):**
```json
{
  "analysis": { ... },
  "starter_apps": [ ... ]
}
```

#### Test Cases

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Profile exists | 200, returns analysis + apps |
| 2 | Profile not complete | 200, partial data |
| 3 | No profile | 404 Not Found |
| 4 | Missing workspace_id | 400 Bad Request |
| 5 | Invalid workspace_id | 404 Not Found |
| 6 | Multiple calls | Consistent data |

---

## State Management Tests

### onboardingStore Tests

#### Test: Store Initialization

```svelte
<script lang="ts">
import { onboardingStore } from '$lib/stores/onboardingStore';

// Test that store initializes with correct defaults
export const testStoreInit = () => {
    const store = onboardingStore;

    assert.equal(store.currentStep, 1); // Welcome screen
    assert.isFalse(store.gmailConnected);
    assert.isEmpty(store.userdata.username);
    assert.isEmpty(store.userdata.email);
};
</script>
```

#### Test: Navigation State

```svelte
<script>
// Test that currentStep updates correctly
export const testNavigation = () => {
    onboardingStore.nextStep();
    assert.equal(onboardingStore.currentStep, 2);

    onboardingStore.prevStep();
    assert.equal(onboardingStore.currentStep, 1);

    onboardingStore.goToStep(5);
    assert.equal(onboardingStore.currentStep, 5);
};
</script>
```

#### Test: Data Persistence

```svelte
<script>
// Test that user data saves and loads from localStorage
export const testPersistence = () => {
    onboardingStore.setUserData({
        email: 'user@example.com',
        username: 'johndoe',
    });

    // Simulate page reload
    const savedData = localStorage.getItem('onboarding');
    const restored = JSON.parse(savedData);

    assert.equal(restored.email, 'user@example.com');
    assert.equal(restored.username, 'johndoe');
};
</script>
```

#### Test: State Recovery

```svelte
<script>
// Test that state recovers after browser crash
export const testRecovery = () => {
    // 1. Save state
    onboardingStore.setUserData({
        email: 'user@example.com',
        username: 'johndoe',
    });
    onboardingStore.goToStep(5);

    // 2. Simulate crash (clear in-memory store)
    // 3. Reload page
    const recovered = loadOnboardingState();

    assert.equal(recovered.currentStep, 5);
    assert.equal(recovered.username, 'johndoe');
};
</script>
```

#### Test: localStorage Cleanup

```svelte
<script>
// Test that completed onboarding clears from localStorage
export const testCleanup = () => {
    onboardingStore.completeOnboarding();

    const saved = localStorage.getItem('onboarding');
    assert.isNull(saved);
};
</script>
```

#### Test: Gmail Connection State

```svelte
<script>
export const testGmailState = () => {
    onboardingStore.setGmailConnected(true);
    assert.isTrue(onboardingStore.gmailConnected);

    // State should persist
    const saved = localStorage.getItem('onboarding');
    assert.include(saved, 'gmailConnected: true');
};
</script>
```

---

## User Flows to Test

### Flow 1: Happy Path (Complete Onboarding)

**Steps:**
1. Visit `/onboarding`
2. Welcome screen → Click "Get Started"
3. Meet OSA → Click "Next"
4. Sign In → Enter valid email → Click "Continue"
5. Gmail → Click "Connect to Gmail" → Approve OAuth → Return to app
6. Username → Enter "johndoe" → Click "Check Availability" → Click "Continue"
7. Analyzing (Step 1/3) → Wait for analysis
8. Analyzing (Step 2/3) → Wait for analysis
9. Analyzing (Step 3/3) → Wait for analysis
10. Starter Apps → Wait for 4 apps to generate → Verify all ready
11. Ready to Launch → Click "Enter OSA Build"
12. Redirect to dashboard → Verify onboarding complete

**Verification:**
- All 11 steps complete successfully
- No errors shown
- Data saved correctly in database
- User can access main app
- Onboarding marked complete

**Expected Time:** 2-3 minutes

### Flow 2: Skip Gmail Connection

**Steps:**
1. Through Step 4 (Sign In)
5. Gmail → Click "Skip" (skip Gmail connection)
6. Username → Enter "johndoe" → Check availability → Continue
7. Analyzing → Complete
8. Starter Apps → Wait for apps (note: fewer apps or less detailed due to missing Gmail data)
9. Ready to Launch → Enter app

**Verification:**
- Skip button works
- gmailConnected = false
- Analysis still completes (with less data)
- Apps still generated (possibly different than Gmail-connected path)
- No errors

### Flow 3: Go Back During Flow

**Steps:**
1. Through Step 6 (Username)
2. Click "Back" → Should go to Gmail screen
3. Click "Back" → Should go to SignIn screen
4. Click "Back" → Should go to Meet OSA screen
5. Click "Back" → Should go to Welcome screen
6. Click "Back" → Should stay on Welcome (no earlier screen)

**Verification:**
- Back navigation works at each step
- No data lost
- Can resume forward navigation

### Flow 4: Browser Close and Resume

**Steps:**
1. Complete through Step 6 (Username claimed)
2. Close browser completely
3. Reopen and visit `/onboarding`
4. App should detect incomplete onboarding
5. Redirect to current step (Step 7: Analyzing)
6. Resume analysis and complete flow

**Verification:**
- State recovered from localStorage
- Analysis resumes from correct step
- No data lost
- User can complete onboarding

### Flow 5: Existing Account Login

**Steps:**
1. Visit `/onboarding`
2. Sign In → Enter email of existing account
3. Should show error: "Account already exists"
4. Offer login redirect

**Verification:**
- Error message clear
- Existing users directed to login, not onboarding
- No duplicate accounts created

### Flow 6: Invalid Username Attempts

**Steps:**
1. Through Step 5 (Username)
2. Enter "ad" (too short) → Check Availability → Error: "Must be 3+ characters"
3. Enter "admin" → Check Availability → Error: "Already taken"
4. Enter "user!name" → Check Availability → Error: "Invalid characters"
5. Enter "valid_name" → Check Availability → Success
6. Continue

**Verification:**
- All validation rules enforced
- Error messages clear
- Can recover and try again
- Valid username accepted

### Flow 7: Network Error Recovery

**Steps:**
1. Start onboarding
2. At some step, disable network (DevTools → Offline)
3. Continue (e.g., click Check Availability)
4. Error message should show
5. Re-enable network
6. Retry action
7. Should succeed

**Verification:**
- Error message clear: "Network error" or "No internet"
- Retry button present
- Can recover without restarting
- Data not corrupted

### Flow 8: API Timeout Handling

**Steps:**
1. Start onboarding
2. Throttle network to very slow (>30s response)
3. Trigger API call (analyze, check username availability)
4. Should show: "Taking longer than expected"
5. Option to retry
6. Click retry
7. Should succeed or show different error

**Verification:**
- Timeout error shown
- UX handles long waits gracefully
- Retry mechanism works
- No data loss

### Flow 9: Gmail OAuth Denial

**Steps:**
1. Through Step 4 (Gmail connection)
2. Click "Connect to Gmail"
3. In Google login: Click "Cancel" or Deny permissions
4. Should return to onboarding
5. Should show error or allow skip
6. Should be able to skip and continue

**Verification:**
- OAuth denial handled gracefully
- No app crash
- Can skip or retry
- Flow continues

### Flow 10: Multiple Device Resume

**Steps:**
1. Start onboarding on Device A (mobile)
2. Complete through Step 6
3. Open Device B (desktop)
4. Visit `/onboarding` on Device B
5. Should NOT see Device A's state (different browser/device)
6. Should start fresh

**Verification:**
- Each device has independent state
- No cross-device leakage
- User can have different onboarding states on different devices (expected)

---

## Performance Tests

### Load Times

| Page | Target | Method |
|------|--------|--------|
| Welcome | <500ms | Measure from navigation start |
| Meet OSA | <500ms | Should load quickly |
| Sign In | <500ms | Form interactive quickly |
| Gmail Connect | <1.5s | OAuth loads Google widget |
| Username | <500ms | Input ready immediately |
| Analyzing (each) | <2s | Show loading state |
| Starter Apps | <2s | Show loading state |
| Ready | <1s | Final confirmation |

### API Response Times

| Endpoint | Target | Note |
|----------|--------|------|
| /analyze | <15s | Takes time for AI analysis |
| /generate-apps | <10s | App generation |
| /apps-status | <1s | Status polling |
| /profile | <1s | Retrieval |

### Animation Performance

| Animation | Target | Note |
|-----------|--------|------|
| Slide-up entrance | 60fps | Smooth entry |
| Status updates | 60fps | Real-time updates |
| Loading spinner | 60fps | Constant animation |
| Button interactions | >120fps | Responsive feedback |

### Memory Usage

| Metric | Target | Measurement |
|--------|--------|-------------|
| Initial page load | <5MB | Browser DevTools |
| After analyzing | <10MB | After long analysis |
| Memory leak | None | Check after 10 min polling |
| Garbage collection | <100ms pause | Watch for freezes |

### Stress Tests

**Test: Rapid navigation**
- Click through all screens rapidly (2-3 seconds per screen)
- Result: No crashes, store handles rapid updates

**Test: Long analysis wait**
- Leave Analyzing screen for 5 minutes
- Result: No timeouts, polling continues, completes successfully

**Test: Rapid status polling**
- 100 status poll requests over 10 seconds
- Result: No API errors, server handles load

**Test: Large app count**
- Generate 50+ starter apps (if possible)
- Result: UI renders without lag

---

## Accessibility Tests

### WCAG 2.1 Level AA Compliance

#### Test: Keyboard Navigation

```
Test: Can complete entire onboarding using only keyboard

Steps:
1. Tab through all interactive elements
2. Verify focus order is logical
3. Use Enter to click buttons
4. Use Escape to cancel (if applicable)
5. Check if any content is keyboard-inaccessible

Expected:
- All buttons reachable via Tab
- All inputs focusable
- Focus indicators visible (understandable)
- No keyboard traps
```

#### Test: Screen Reader (NVDA, JAWS, VoiceOver)

```
Test: Onboarding usable with screen reader

Steps:
1. Enable screen reader
2. Navigate through each screen
3. Verify all text is readable
4. Verify buttons are properly labeled
5. Verify form inputs have labels
6. Verify error messages are announced
7. Verify status updates are announced (aria-live)

Expected:
- All content readable
- Buttons labeled properly
- Inputs have associated labels
- Errors announced clearly
- Loading state announced
```

#### Test: Color Contrast

```
Test: Text/background contrast meets WCAG AA (4.5:1 for normal, 3:1 for large)

Checking:
- Button text vs background
- Form labels vs background
- Error messages (usually red)
- Helper text (usually gray)
- All text in gradient backgrounds

Tool: WebAIM Contrast Checker, Lighthouse
```

#### Test: Focus Management

```
Test: Focus is managed correctly after actions

Scenarios:
1. After form submission → Focus moves to next form or confirmation
2. After error → Focus moves to error message
3. After modal/dialog → Focus trapped in modal
4. On page navigation → Focus moves to main content

Tool: DevTools → Inspect → Find focus outline
```

#### Test: Form Accessibility

```
Test: All form inputs accessible and usable

Checks:
- Labels associated with inputs (for/id)
- Required fields marked (required attribute or text)
- Error messages associated with inputs (aria-describedby)
- Password fields identified (type="password")
- Error messages clear and actionable
- Success messages clear

Example:
<label for="email">Email Address</label>
<input id="email" type="email" aria-describedby="email-error" />
<div id="email-error" class="error">Invalid email</div>
```

#### Test: Status/Loading Updates

```
Test: Loading states and status updates announced

Implementation:
<div aria-busy="true" aria-live="polite">
  Loading your profile...
</div>

When status changes:
<div aria-live="assertive">
  Profile loaded successfully!
</div>

Verification: Screen reader announces updates without page reload
```

#### Test: Skip Links (if applicable)

```
Test: Users can skip repetitive content

Pattern:
<a href="#main-content" class="skip-link">Skip to main content</a>

Verification: Accessible to keyboard users, skips header/nav
```

#### Test: Zoom and Text Scaling

```
Test: Content remains usable at 200% zoom

Steps:
1. Zoom page to 200% (Ctrl/Cmd + +)
2. Verify no content is cut off
3. Verify horizontal scroll is minimal
4. Verify buttons still clickable

Acceptance: All content remains usable
```

#### Test: Motion/Animation

```
Test: No problematic animations for motion-sensitive users

Check:
1. No autoplaying animations
2. Respect prefers-reduced-motion CSS media query
3. Animations can be paused or skipped
4. No flashing content (>3 per second)

CSS:
@media (prefers-reduced-motion: reduce) {
  * {
    animation: none !important;
    transition: none !important;
  }
}
```

---

## Browser Compatibility

### Supported Browsers

| Browser | Min Version | Status | Test Priority |
|---------|-------------|--------|----------------|
| Chrome | 100+ | Required | Critical |
| Firefox | 100+ | Required | Critical |
| Safari | 14+ | Required | Critical |
| Edge | 100+ | Required | High |
| Chrome Mobile | 100+ | Required | Critical |
| Safari iOS | 14+ | Required | Critical |
| Firefox Mobile | 100+ | Required | High |

### Test Matrix

**Desktop Combinations:**
- Chrome (latest) on Windows 11
- Chrome (latest) on macOS
- Chrome (latest) on Ubuntu
- Firefox (latest) on Windows 11
- Firefox (latest) on macOS
- Safari (latest) on macOS
- Edge (latest) on Windows 11

**Mobile Combinations:**
- Chrome Mobile on Android 12+
- Safari on iOS 14+
- Firefox Mobile on Android 12+
- Samsung Internet on Android

### Critical Tests Per Browser

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| OAuth flow | ✓ | ✓ | ✓ | ✓ |
| Form inputs | ✓ | ✓ | ✓ | ✓ |
| Animations | ✓ | ✓ | ✓ | ✓ |
| Fetch/API | ✓ | ✓ | ✓ | ✓ |
| localStorage | ✓ | ✓ | ✓ | ✓ |
| CSS Grid/Flex | ✓ | ✓ | ✓ | ✓ |
| SVG rendering | ✓ | ✓ | ✓ | ✓ |

### Known Issues Template

```markdown
## Known Issues

### Issue: Safari iOS - OAuth callback
- Affected: Safari on iOS 14-15
- Impact: OAuth login fails
- Workaround: None available
- Fix: Awaiting WebKit update
- Priority: High
- Status: Investigating

### Issue: Firefox - Form autofill
- Affected: Firefox 100-102
- Impact: Username field may autofill unexpectedly
- Workaround: Clear browser autofill
- Fix: Monitor Firefox updates
- Priority: Low
- Status: Monitoring
```

---

## Manual Testing Checklist

### Pre-Release Testing (Before Deploy to Production)

**Tester:** ____________________
**Date:** ____________________
**Build Version:** ____________________
**Environment:** staging / production

#### Screen 1: Welcome
- [ ] Page loads within 1 second
- [ ] All three feature cards display correctly
- [ ] "Get Started" button is clickable
- [ ] Animation plays smoothly
- [ ] Mobile (375px) layout correct
- [ ] Tablet (768px) layout correct
- [ ] Desktop (1440px) layout correct
- [ ] Dark mode styling correct
- [ ] All copy is readable and correct

#### Screen 2: Meet OSA
- [ ] Page loads correctly
- [ ] OSA message displays
- [ ] Next button works
- [ ] Back button works
- [ ] Mobile layout correct

#### Screen 3: Sign In
- [ ] Email input accepts valid emails
- [ ] Invalid emails rejected with error message
- [ ] Empty email rejected
- [ ] Continue button disabled until valid email
- [ ] Submitting valid email works
- [ ] Duplicate email shows error
- [ ] Form submission under 2 seconds
- [ ] Error messages clear and actionable
- [ ] Mobile layout correct

#### Screen 4: Gmail
- [ ] Gmail OAuth button displays
- [ ] Clicking OAuth opens Google login (new window/tab)
- [ ] OAuth callback returns to app
- [ ] Gmail connection confirmed
- [ ] Skip button works
- [ ] Continue button enabled after connection or skip
- [ ] Mobile layout correct

#### Screen 5: Username
- [ ] Username input focuses on load
- [ ] 2-character username rejected with error
- [ ] 3-character username accepted
- [ ] 255+ character username handled
- [ ] Special characters rejected
- [ ] Check Availability button works
- [ ] Taken usernames show error
- [ ] Available username shows success checkmark
- [ ] Continue button disabled until available
- [ ] Back button works
- [ ] Mobile layout correct

#### Screen 6: Analyzing (Step 1/3)
- [ ] Page loads
- [ ] Loading animation plays smoothly
- [ ] Progress shows "1/3"
- [ ] Analysis completes within 15 seconds
- [ ] Transitions to Step 2 after completion
- [ ] Mobile layout correct

#### Screen 7: Analyzing (Step 2/3)
- [ ] Progress shows "2/3"
- [ ] Analysis completes
- [ ] Transitions to Step 3
- [ ] Mobile layout correct

#### Screen 8: Analyzing (Step 3/3)
- [ ] Progress shows "3/3"
- [ ] Analysis completes
- [ ] Transitions to Starter Apps
- [ ] Mobile layout correct

#### Screen 9: Starter Apps
- [ ] 4 app cards display
- [ ] Each app shows title, description, icon, category
- [ ] Loading state shows for apps still generating
- [ ] "Ready to Launch" button disabled until all ready
- [ ] All apps ready within 30 seconds
- [ ] "Ready to Launch" button enabled when done
- [ ] Button click works
- [ ] Mobile 1-column layout
- [ ] Tablet 2-column layout
- [ ] Desktop 4-column or 2x2 layout

#### Screen 10: Ready to Launch
- [ ] Page loads successfully
- [ ] Success message displays
- [ ] User profile summary shows
- [ ] 4 starter apps listed
- [ ] "Enter OSA Build" button is prominent
- [ ] Clicking button navigates to main app
- [ ] Mobile layout correct

#### State & Navigation
- [ ] Onboarding state persists after page reload
- [ ] Can go back multiple screens
- [ ] Back button works at each screen
- [ ] Forward navigation works after going back
- [ ] Completing onboarding clears localStorage
- [ ] Returning to /onboarding after completion goes to dashboard

#### Error Handling
- [ ] Disable network → Error messages show
- [ ] Re-enable network → Retry works
- [ ] API errors handled with clear messages
- [ ] No console errors during flow
- [ ] No unhandled promise rejections

#### Performance
- [ ] Navigation between screens is smooth
- [ ] Animations are fluid (no stuttering)
- [ ] Loading states are responsive
- [ ] No lag when entering text

#### Accessibility (Keyboard)
- [ ] Can tab through all screens
- [ ] Focus indicators visible
- [ ] Can submit forms with Enter key
- [ ] No keyboard traps
- [ ] Focus order is logical

#### Accessibility (Screen Reader)
- [ ] Screen reader announces all content
- [ ] Form labels are associated
- [ ] Error messages are announced
- [ ] Loading state announced
- [ ] Status updates announced

#### Cross-Browser (if applicable)
- [ ] Chrome: All tests pass
- [ ] Firefox: All tests pass
- [ ] Safari: All tests pass
- [ ] Edge: All tests pass
- [ ] Chrome Mobile: All tests pass
- [ ] Safari iOS: All tests pass

#### Final Sign-Off
- [ ] All checks passed
- [ ] No critical bugs found
- [ ] No major UX issues
- [ ] Ready for release

**Issues Found:**
1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

**Tester Signature:** ____________________

---

## Automated Test Scripts

### Example: Go Backend Handler Test

**File:** `internal/handlers/onboarding_test.go`

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAnalyzeUser_Success tests the happy path
func TestAnalyzeUser_Success(t *testing.T) {
	// Setup
	h := setupTestHandlers(t)

	payload := map[string]interface{}{
		"email":             "user@example.com",
		"gmail_connected":   true,
		"calendar_connected": false,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(
		"POST",
		"/api/osa-onboarding/analyze",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)

	// Execute
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	analysis := resp["analysis"].(map[string]interface{})
	assert.NotNil(t, analysis["insights"])
	assert.NotNil(t, analysis["profile_summary"])
}

// TestAnalyzeUser_InvalidEmail tests validation
func TestAnalyzeUser_InvalidEmail(t *testing.T) {
	h := setupTestHandlers(t)

	payload := map[string]interface{}{
		"email":           "invalid",
		"gmail_connected": true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(
		"POST",
		"/api/osa-onboarding/analyze",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAnalyzeUser_Unauthorized tests auth
func TestAnalyzeUser_Unauthorized(t *testing.T) {
	h := setupTestHandlers(t)

	payload := map[string]interface{}{
		"email":           "user@example.com",
		"gmail_connected": true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(
		"POST",
		"/api/osa-onboarding/analyze",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	// NO Authorization header

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

---

### Example: Svelte Component Test

**File:** `src/routes/onboarding/username/+page.test.ts`

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import Page from './+page.svelte';

// Mock the API
vi.mock('$lib/api/osa-onboarding', () => ({
	checkUsernameAvailability: vi.fn(),
}));

describe('Username Screen', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('should render the username input', () => {
		render(Page);

		const input = screen.getByPlaceholderText('bekorains');
		expect(input).toBeInTheDocument();
	});

	it('should validate minimum length', async () => {
		render(Page);

		const input = screen.getByPlaceholderText('bekorains');
		const checkBtn = screen.getByText('Check Availability');

		// Enter short username
		await userEvent.type(input, 'ab');
		fireEvent.click(checkBtn);

		const error = screen.getByText(/at least 3 characters/i);
		expect(error).toBeInTheDocument();
	});

	it('should check availability', async () => {
		const { checkUsernameAvailability } = await import('$lib/api/osa-onboarding');
		checkUsernameAvailability.mockResolvedValue({ available: true });

		render(Page);

		const input = screen.getByPlaceholderText('bekorains');
		const checkBtn = screen.getByText('Check Availability');

		await userEvent.type(input, 'validuser');
		fireEvent.click(checkBtn);

		await waitFor(() => {
			expect(screen.getByText(/Available!/i)).toBeInTheDocument();
		});
	});

	it('should show taken username error', async () => {
		const { checkUsernameAvailability } = await import('$lib/api/osa-onboarding');
		checkUsernameAvailability.mockResolvedValue({ available: false });

		render(Page);

		const input = screen.getByPlaceholderText('bekorains');
		const checkBtn = screen.getByText('Check Availability');

		await userEvent.type(input, 'admin');
		fireEvent.click(checkBtn);

		await waitFor(() => {
			expect(screen.getByText(/already taken/i)).toBeInTheDocument();
		});
	});

	it('should enable continue button only when available', async () => {
		const { checkUsernameAvailability } = await import('$lib/api/osa-onboarding');
		checkUsernameAvailability.mockResolvedValue({ available: true });

		render(Page);

		const continueBtn = screen.getByText('Continue');
		expect(continueBtn).toBeDisabled();

		const input = screen.getByPlaceholderText('bekorains');
		const checkBtn = screen.getByText('Check Availability');

		await userEvent.type(input, 'validuser');
		fireEvent.click(checkBtn);

		await waitFor(() => {
			expect(continueBtn).not.toBeDisabled();
		});
	});

	it('should navigate back', async () => {
		const { goto } = await import('$app/navigation');
		vi.mock('$app/navigation');

		render(Page);

		const backBtn = screen.getByText('Back');
		fireEvent.click(backBtn);

		expect(goto).toHaveBeenCalledWith('/onboarding/gmail');
	});
});
```

---

### Example: E2E Test with Playwright

**File:** `tests/e2e/onboarding.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
	test.beforeEach(async ({ page }) => {
		// Clear localStorage before each test
		await page.context().clearCookies();
		await page.goto('/onboarding');
	});

	test('complete onboarding happy path', async ({ page }) => {
		// Screen 1: Welcome
		await expect(page).toHaveTitle(/Welcome to OSA Build/);
		await page.click('text=Get Started');

		// Screen 2: Meet OSA
		await expect(page).toHaveURL(/meet-osa/);
		await page.click('text=Next');

		// Screen 3: Sign In
		await expect(page).toHaveURL(/signin/);
		await page.fill('input[type="email"]', 'newuser@example.com');
		await page.click('button:has-text("Continue")');

		// Screen 4: Gmail (Skip)
		await expect(page).toHaveURL(/gmail/);
		await page.click('text=Skip');

		// Screen 5: Username
		await expect(page).toHaveURL(/username/);
		await page.fill('input[placeholder="bekorains"]', 'validuser');
		await page.click('text=Check Availability');

		// Wait for availability check
		await page.waitForSelector('text=Available');
		await page.click('button:has-text("Continue")');

		// Screen 6-8: Analyzing
		await expect(page).toHaveURL(/analyzing/);

		// Wait for analysis to complete (max 20 seconds)
		await page.waitForURL(/starter-apps/, { timeout: 20000 });

		// Screen 9: Starter Apps
		await expect(page).toHaveURL(/starter-apps/);

		// Wait for apps to generate
		const appsReady = page.locator('text=Ready to Launch');
		await appsReady.waitFor({ timeout: 30000 });

		// Click Ready to Launch
		await page.click('text=Ready to Launch');

		// Should redirect to dashboard
		await expect(page).toHaveURL(/dashboard|app/);

		// Verify localStorage is cleared
		const stored = await page.evaluate(() => localStorage.getItem('onboarding'));
		expect(stored).toBeNull();
	});

	test('skip gmail connection', async ({ page }) => {
		// Navigate to Gmail screen
		await page.click('text=Get Started'); // Screen 1
		await page.click('text=Next'); // Screen 2
		await page.fill('input[type="email"]', 'user@example.com'); // Screen 3
		await page.click('button:has-text("Continue")');

		// Skip Gmail
		await expect(page).toHaveURL(/gmail/);
		await page.click('text=Skip');

		// Should continue to Username
		await expect(page).toHaveURL(/username/);
	});

	test('navigate back through screens', async ({ page }) => {
		// Go forward to Username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'user@example.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Now go back
		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/gmail/);

		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/signin/);

		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/meet-osa/);

		await page.click('button:has-text("Back")');
		await expect(page).toHaveURL(/\/onboarding$/);
	});

	test('resume after browser close', async ({ page, context }) => {
		// Go through some screens
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'user@example.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');
		await page.fill('input[placeholder="bekorains"]', 'testuser');
		await page.click('text=Check Availability');
		await page.waitForSelector('text=Available');
		await page.click('button:has-text("Continue")');

		// Get current step from localStorage
		const step = await page.evaluate(() => {
			const stored = localStorage.getItem('onboarding');
			return stored ? JSON.parse(stored).currentStep : null;
		});

		// Close page and reopen
		await page.close();
		const newPage = await context.newPage();
		await newPage.goto('/onboarding');

		// Should be on analyzing screen
		await expect(newPage).toHaveURL(/analyzing/);

		// Should still have stored data
		const restored = await newPage.evaluate(() => {
			const stored = localStorage.getItem('onboarding');
			return stored ? JSON.parse(stored) : null;
		});

		expect(restored).not.toBeNull();
		expect(restored.username).toBe('testuser');
	});

	test('error recovery - network error', async ({ page }) => {
		// Go to username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'user@example.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Simulate offline
		await page.context().setOffline(true);

		await page.fill('input[placeholder="bekorains"]', 'testuser');
		await page.click('text=Check Availability');

		// Should show error
		await expect(page.locator('text=/Network|offline/i')).toBeVisible({ timeout: 5000 });

		// Go back online
		await page.context().setOffline(false);

		// Retry
		await page.click('text=Retry');
		await page.waitForSelector('text=Available');

		expect(await page.locator('text=Available').count()).toBeGreaterThan(0);
	});

	test('form validation - duplicate username', async ({ page }) => {
		// Navigate to username screen
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'user@example.com');
		await page.click('button:has-text("Continue")');
		await page.click('text=Skip');

		// Try reserved username
		await page.fill('input[placeholder="bekorains"]', 'admin');
		await page.click('text=Check Availability');

		// Should show "already taken"
		await expect(page.locator('text=/already taken/i')).toBeVisible();

		// Continue button should be disabled
		const continueBtn = page.locator('button:has-text("Continue")');
		await expect(continueBtn).toBeDisabled();
	});

	test('mobile responsiveness', async ({ page }) => {
		// Set mobile viewport
		await page.setViewportSize({ width: 375, height: 667 });

		// Welcome screen
		await expect(page.locator('text=Welcome to OSA Build')).toBeVisible();

		// Feature cards should stack vertically
		const cards = page.locator('[class*="feature-card"]');
		const count = await cards.count();
		expect(count).toBe(3);

		// Continue through flow
		await page.click('text=Get Started');
		await page.click('text=Next');
		await page.fill('input[type="email"]', 'mobile@example.com');
		await page.click('button:has-text("Continue")');

		// Content should be readable without horizontal scroll
		const bodyWidth = await page.evaluate(() => document.body.offsetWidth);
		const maxElementWidth = await page.evaluate(() => {
			return Math.max(
				...Array.from(document.querySelectorAll('*')).map(el =>
					el.getBoundingClientRect().width
				)
			);
		});

		// Allow small overflow for scrollbar
		expect(maxElementWidth).toBeLessThanOrEqual(bodyWidth + 20);
	});
});
```

---

## Test Data

### Sample User Data

```json
{
  "valid_emails": [
    "user@example.com",
    "test.user@domain.co.uk",
    "user+tag@gmail.com",
    "john.doe@organization.io",
    "a@b.co"
  ],
  "invalid_emails": [
    "plainaddress",
    "@nodomain.com",
    "username@.com",
    "user@domain",
    "user @domain.com",
    "user@domain..com",
    ""
  ],
  "valid_usernames": [
    "john_doe",
    "johndoe",
    "user123",
    "a1b2c3",
    "the_user_name",
    "user_99"
  ],
  "invalid_usernames": [
    "ab",
    "user-name",
    "user!name",
    "user@domain",
    "user name",
    "user.name",
    ""
  ],
  "reserved_usernames": [
    "admin",
    "root",
    "osa",
    "test",
    "support",
    "system",
    "api",
    "www",
    "mail",
    "ftp"
  ]
}
```

### Mock Analysis Response

```json
{
  "analysis": {
    "insights": [
      "Focuses on team productivity",
      "Values automation and efficiency",
      "Interested in project management",
      "Needs cross-team collaboration"
    ],
    "interests": [
      "Productivity",
      "Project Management",
      "Team Collaboration",
      "Analytics",
      "Integration"
    ],
    "tools_used": [
      "Gmail",
      "Google Calendar",
      "Google Drive",
      "Slack",
      "Asana"
    ],
    "profile_summary": "John is a project manager who uses email, calendar, and task management tools extensively. They work in a team environment and would benefit from centralized project tracking and team communication tools.",
    "raw_data": {
      "emails_count": 2847,
      "calendar_events": 156,
      "collaboration_score": 8.5,
      "email_response_time_avg": "2.3 hours"
    }
  }
}
```

### Mock Starter Apps Response

```json
{
  "starter_apps": [
    {
      "id": "app-1",
      "title": "Project Dashboard",
      "description": "Central hub for tracking all your projects and their status",
      "icon_emoji": "📊",
      "icon_url": "https://api.example.com/icons/dashboard.svg",
      "reasoning": "Based on your project management focus and frequent calendar usage, a centralized dashboard would help you track all projects in one place.",
      "category": "Productivity",
      "status": "ready",
      "workflow_id": "wf-proj-dashboard-1"
    },
    {
      "id": "app-2",
      "title": "Team Sync",
      "description": "Manage team meetings, agendas, and meeting notes",
      "icon_emoji": "👥",
      "icon_url": "https://api.example.com/icons/team.svg",
      "reasoning": "Your high collaboration score and frequent team interactions suggest a dedicated team sync tool would be valuable for organizing meetings and notes.",
      "category": "Collaboration",
      "status": "ready",
      "workflow_id": "wf-team-sync-1"
    },
    {
      "id": "app-3",
      "title": "Task Inbox",
      "description": "Smart inbox for managing tasks from email and other sources",
      "icon_emoji": "✅",
      "icon_url": "https://api.example.com/icons/tasks.svg",
      "reasoning": "With over 2800 emails analyzed, a task inbox would help you convert emails into actionable tasks and reduce inbox overload.",
      "category": "Productivity",
      "status": "ready",
      "workflow_id": "wf-task-inbox-1"
    },
    {
      "id": "app-4",
      "title": "Analytics Hub",
      "description": "Track metrics, KPIs, and team performance",
      "icon_emoji": "📈",
      "icon_url": "https://api.example.com/icons/analytics.svg",
      "reasoning": "Your analytical approach and project management focus suggest you'd benefit from a centralized analytics dashboard for performance tracking.",
      "category": "Analytics",
      "status": "ready",
      "workflow_id": "wf-analytics-hub-1"
    }
  ],
  "ready_to_launch": true
}
```

### Test Fixtures

**File:** `tests/fixtures/onboarding.ts`

```typescript
export const testUser = {
	id: 'user-123',
	email: 'testuser@example.com',
	username: 'testuser',
	created_at: new Date().toISOString(),
	gmail_connected: true,
};

export const testAnalysis = {
	insights: ['insight-1', 'insight-2'],
	interests: ['productivity', 'collaboration'],
	tools_used: ['email', 'calendar'],
	profile_summary: 'Test profile',
	raw_data: {},
};

export const testApps = [
	{
		id: 'app-1',
		title: 'App 1',
		description: 'Test app 1',
		icon_emoji: '📊',
		icon_url: '',
		reasoning: 'Test reasoning',
		category: 'Productivity',
		status: 'ready' as const,
		workflow_id: 'wf-1',
	},
	// ... 3 more apps
];

export const mockAnalyzeResponse = {
	analysis: testAnalysis,
};

export const mockGenerateAppsResponse = {
	starter_apps: testApps,
	ready_to_launch: true,
};
```

---

## Bug Report Template

### Issue Report Format

**Title:** [Brief description of issue]

**Severity:** Critical / High / Medium / Low

**Environment:**
- OS: Windows 11 / macOS 12 / Ubuntu 22.04 / iOS 16 / Android 12
- Browser: Chrome 120 / Firefox 121 / Safari 17 / Edge 120
- Build Version: v1.2.3
- Environment: staging / production

**Reproduction Steps:**
1. Start at `/onboarding`
2. Click "Get Started"
3. ... (detailed steps)
4. Observe: [what happens]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Screenshots/Videos:**
[Attach if applicable]

**Console Errors:**
```
[Paste any JavaScript errors from DevTools]
```

**Network Errors:**
```
[Paste any API errors from DevTools Network tab]
```

**Additional Context:**
- Does this happen consistently? Yes / No / Sometimes
- When did this start? After which deployment?
- Does this affect all users or just some?
- Any workarounds discovered?

**Priority for Fix:**
- Blocking release? Yes / No
- Affects core functionality? Yes / No
- Impacts user experience? Yes / No

---

## Appendix: Quick Reference

### Common Test Commands

```bash
# Run all tests
npm test

# Run Go tests
go test ./...

# Run E2E tests
npx playwright test

# Run specific test file
npm test -- src/routes/onboarding/username

# Generate coverage report
npm test -- --coverage

# Watch mode
npm test -- --watch
```

### Key URLs

- **Local:** http://localhost:5173/onboarding
- **Staging:** https://staging.example.com/onboarding
- **Production:** https://app.example.com/onboarding

### Contact

- QA Lead: [name]
- Backend Lead: [name]
- Frontend Lead: [name]
- Product Manager: [name]

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-18 | QA Team | Initial comprehensive testing plan |

---

**Approval Sign-Off:**

- QA Lead: __________________ Date: __________
- Engineering Lead: __________________ Date: __________
- Product Manager: __________________ Date: __________
