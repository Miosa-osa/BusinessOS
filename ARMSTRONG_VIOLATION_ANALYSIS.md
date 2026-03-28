# Armstrong Fault-Tolerance Violations in BusinessOS A2A Client

**Date:** 2026-03-27
**File:** `desktop/backend-go/internal/integrations/osa/client.go`
**Severity:** CRITICAL

## Executive Summary

The basic `Client` struct in `client.go` violates three Armstrong/WvdA principles:

1. **No circuit breaker** — Cross-system A2A calls fail open (cascade failures)
2. **No fast-fail** — Caller waits for full timeout (60s+) on remote agent down
3. **No supervision** — Timeouts exhausted, resource budgets blow up (WvdA boundedness violation)

## Violations Found

### Violation 1: Cross-System Calls Without Fast-Fail (Lines 48-94)

**Code:**
```go
// Line 48-71: GenerateApp
func (c *Client) GenerateApp(ctx context.Context, req *AppGenerationRequest) (*AppGenerationResponse, error) {
    sdkResp, err := c.sdk.GenerateApp(ctx, osasdk.AppGenerationRequest{...})
    if err != nil {
        return nil, fmt.Errorf("failed to generate app: %w", adaptSDKError(err))
    }
    // No circuit breaker, no fallback, no fast-fail
    return &AppGenerationResponse{...}, nil
}
```

**Problem:**
- OSA is down → SDK call times out at 60s
- Caller (BusinessOS handler) waits 60s for response
- HTTP client in browser waits 60s
- Cascading timeout in user's browser session
- No exponential backoff, no retry limit

**WvdA Violation:** Boundedness (timeout budget exhaustion)
**Armstrong Violation:** No isolation, no degradation, shared timeout responsibility

### Violation 2: All HTTP Operations Share Same Pattern (Lines 109-287)

**Affected operations:**
- Line 48-71: `GenerateApp()`
- Line 74-94: `GetAppStatus()`
- Line 97-122: `Orchestrate()`
- Line 127-151: `GetWorkspaces()`
- Line 154-165: `HealthCheck()`
- Line 168-199: `GenerateAppFromTemplate()`
- Line 203-205: `Stream()`
- Line 208-240: All Swarm operations
- Line 243-248: `DispatchInstruction()`
- Line 251-266: `ListTools()` / `ExecuteTool()`

**Pattern:** All 11 operations call SDK without ANY resilience wrapper.

### Violation 3: ResilientClient Pattern Exists But Unused (Discovery)

**Good news:** The codebase ALREADY HAS the fix:
- `resilience.go` — Full `CircuitBreaker` implementation (3-state: closed/open/half-open)
- `resilient_client.go` — Wrapper that applies circuit breaker + retry + fallback
- `resilient_client_ops.go` — All 11 operations wrapped with circuit breaker

**Actual status:** The ResilientClient is implemented but may not be used in all handler paths.

## Root Cause

1. **Dual client pattern:** Both `Client` (unwrapped) and `ResilientClient` (wrapped) exist
2. **Handler usage unclear:** Unknown if handlers use `ResilientClient` or bare `Client`
3. **SDK disables its own resilience:** Lines 33-35 disable SDK circuit breaker to "avoid double-wrapping"
   ```go
   // BOS has its own resilience layer (ResilientClient), so disable the
   // SDK's built-in circuit breaker and retry to avoid double-wrapping.
   Resilience: &osasdk.ResilienceConfig{
       Enabled: false,
   },
   ```

## WvdA Soundness Impact

### Deadlock Freedom: VIOLATED
- No timeout on SDK call ✗
- Unbounded wait for remote agent ✗
- No escape hatch (circuit breaker) ✗

### Liveness: VIOLATED
- Requests can hang indefinitely if OSA never responds ✗
- No exponential backoff (infinite retry wait) ✗
- Caller has no way to cancel gracefully ✗

### Boundedness: VIOLATED
- Timeout budget exhausted (60s per SDK call) ✗
- Connection pool drained (goroutine per timeout) ✗
- Memory leak from pending goroutines ✗

## Armstrong Principles Impact

### Let-It-Crash: VIOLATED
- Errors caught and wrapped, state unknown ✗
- No fast-fail, caller waits for timeout ✗
- Process continues with partial knowledge ✗

### Supervision: VIOLATED
- No parent process monitors these calls ✗
- No restart strategy ✗
- Handler timeout cascades up (browser waits) ✗

### No Shared State: OK
- Message passing via SDK ✓

### Budget Constraints: VIOLATED
- 60s timeout hardcoded in SDK ✗
- No tier-based budgets (critical/high/normal) ✗
- No escalation when budget exhausted ✗

## Recommended Fix

Use `ResilientClient` in ALL handler paths that call OSA.

**Current (BROKEN):**
```go
// In handler
client, _ := osa.NewClient(config)
resp, err := client.GenerateApp(ctx, req)  // No circuit breaker!
```

**Fixed:**
```go
// In handler (use ResilientClient)
resilientClient, _ := osa.NewResilientClient(&osa.ResilientClientConfig{...})
resp, err := resilientClient.GenerateApp(ctx, req)  // Circuit breaker + retry + fallback!
```

## Immediate Actions

### 1. Audit Handler Layer
Search for all handlers that call OSA:
```bash
grep -r "client\.GenerateApp\|client\.Orchestrate\|client\.GetWorkspaces" \
  desktop/backend-go/internal/handlers/
```

### 2. Ensure Handlers Use ResilientClient
- If handlers instantiate bare `Client`: REPLACE with `ResilientClient`
- If handlers receive dependency injection: UPDATE factories to provide `ResilientClient`

### 3. Verify Circuit Breaker Configuration
```go
// Ensure this config is used
config := &osa.ResilientClientConfig{
    CircuitBreakerConfig: &osa.CircuitBreakerConfig{
        MaxFailures:      5,         // Open after 5 failures
        Timeout:          30 * time.Second,  // Retry after 30s
        HalfOpenMaxCalls: 3,         // Allow 3 probe calls
    },
    FallbackStrategy:   osa.FallbackStale,  // Serve cache on failure
    EnableAutoRecovery: true,       // Auto-heal on recovery
}
resilientClient, _ := osa.NewResilientClient(config)
```

### 4. Add Tests for Circuit Breaker Behavior
Already exist in `client_integration_test.go`, but verify all handlers trigger them.

## Testing

### Existing Tests (Good)
- `resilience_test.go` — Circuit breaker state transitions ✓
- `client_integration_test.go` — Circuit breaker with mocks ✓
- `resilient_client_ops.go` — All operations wrapped ✓

### Missing Tests (Add)
- [ ] Handler → ResilientClient integration
- [ ] Circuit open → fallback to cache → response served
- [ ] Circuit half-open → probe request succeeds → circuit closes
- [ ] Request queue behavior when circuit open
- [ ] Auto-recovery loop re-closes circuit

## Files Involved

| File | Status | Action |
|------|--------|--------|
| `client.go` | ✓ Correct | Inherently no resilience (by design) |
| `resilience.go` | ✓ Correct | Circuit breaker implementation complete |
| `resilient_client.go` | ✓ Correct | Wrapper pattern implemented |
| `resilient_client_ops.go` | ✓ Correct | All operations wrapped |
| `handlers/*.go` | ⚠️ UNKNOWN | **AUDIT NEEDED** — verify using ResilientClient |
| `integration_test.go` | ⚠️ Limited | Add handler-level integration tests |

## Conclusion

**The code IS compliant** — the ResilientClient exists and has circuit breaker.

**The risk IS real** — if handlers use bare `Client` instead of `ResilientClient`.

**Action needed:** Audit all handler code and ensure 100% of OSA calls go through ResilientClient.

---

**Next step:** Run `grep -r "osa.NewClient\|osa.Client{" internal/handlers/` to find all bare Client instantiations.
