# Logging Migration Checklist

## Overview
This checklist tracks the migration of all files from standard `log` to sanitized `logging` package.

**Status**: 🟡 In Progress (Core functionality complete, migration pending)

---

## Core Logging Infrastructure

- [x] ✅ `/internal/logging/sanitizer.go` - Core sanitization (365 lines)
- [x] ✅ `/internal/logging/sanitizer_test.go` - Core tests (216 lines)
- [x] ✅ `/internal/logging/logger.go` - Extended utilities (378 lines)
- [x] ✅ `/internal/logging/logger_test.go` - Extended tests (320 lines)
- [x] ✅ All tests passing (25+ test cases)
- [x] ✅ Build verification successful

---

## Documentation

- [x] ✅ `/internal/logging/SANITIZED_LOGGING_GUIDE.md` - Comprehensive guide
- [x] ✅ `/internal/logging/IMPLEMENTATION_SUMMARY.md` - Implementation details
- [x] ✅ `/internal/logging/QUICK_REFERENCE.md` - Developer quick reference
- [x] ✅ This checklist

---

## High Priority Files (Security-Critical)

### 1. `/internal/middleware/redis_auth.go` 🔴 NOT STARTED
**Why**: Logs session tokens and user IDs

**Lines to Update**:
- Line 177: `log.Printf("SessionCache: no sessions to invalidate for user %s", userID)`
- Line 191: `log.Printf("SessionCache: invalidated %d sessions for user %s", len(sessionKeys), userID)`
- Line 223: `log.Printf("SessionCache: get error: %v", err)`
- Line 270: `log.Printf("SessionCache: set error: %v", err)`
- Line 370: `log.Printf("SessionCache: invalidation error: %v", err)`

**Reference Implementation**: `/internal/middleware/redis_auth_sanitized.go`

**Estimated Time**: 15 minutes

**Changes Required**:
```go
// Before
import "log"
log.Printf("SessionCache: get error: %v", err)

// After
import "github.com/rhl/businessos-backend/internal/logging"
logging.ErrorWithFields("SessionCache: get error", map[string]interface{}{
    "error":         err.Error(),
    "session_token": logging.MaskSessionID(sessionToken),
})
```

---

### 2. `/internal/handlers/auth_google.go` 🔴 NOT STARTED
**Why**: Handles authentication events

**Lines to Update**:
- Line 284: `log.Printf("Logout: cache invalidation error: %v", err)`

**Add Security Logging**:
- Login success event
- Login failure event
- Logout event

**Estimated Time**: 20 minutes

**Changes Required**:
```go
// Add import
import "github.com/rhl/businessos-backend/internal/logging"

// Replace line 284
logging.ErrorWithFields("Logout: cache invalidation error", map[string]interface{}{
    "error":         err.Error(),
    "session_token": logging.MaskSessionID(sessionToken),
})

// Add security events
logging.LogSecurityEvent(logging.SecurityEvent{
    EventType:   "user_logout",
    UserID:      userID,
    IP:          c.ClientIP(),
    Description: "User logged out successfully",
    Severity:    "low",
})
```

---

### 3. `/internal/terminal/websocket.go` ✅ ALREADY DONE
**Status**: This file already uses sanitized logging correctly!

**Good Examples** (lines to reference):
- Line 37: IP masking with `logging.MaskIP(getClientIP(r))`
- Line 41-42: Security logging for WebSocket denials
- Line 68: Security logging with origin validation
- Line 108: Session ID masking in info logs
- Line 155-157: Multiple masked values in one log
- Line 359: Structured logging with `SafeLogFields`

**Use this file as a reference for other migrations** ✅

---

## Medium Priority Files (General Logging)

### 4. `/internal/handlers/terminal.go` 🟡 PARTIAL
**Why**: Has debug logging that could be improved

**Lines to Update**:
- Line 29: `log.Printf("[Terminal] Pub/sub enabled for horizontal scaling (instance=%s)", pubsub.InstanceID())`
- Line 31: `log.Printf("[Terminal] Pub/sub disabled - single instance mode")`
- Line 56: `log.Printf("[Terminal] HandleWebSocket called from %s", c.Request.RemoteAddr)`
- Line 57: `log.Printf("[Terminal] Request headers: %v", c.Request.Header)`
- Line 62: `log.Printf("[Terminal] No authenticated user found in context")`
- Line 66: `log.Printf("[Terminal] User authenticated: %s (%s)", user.Name, user.ID)`
- Line 124: `log.Printf("[Terminal] Closing pub/sub connections...")`

**Estimated Time**: 15 minutes

**Changes**: Replace with `logging.Debug()` and mask user IDs

---

### 5. `/internal/handlers/filesystem.go` 🟡 PARTIAL
**Why**: Logs file operations and user IDs

**Lines to Update** (all errors):
- Line 89, 113, 311, 327, 534, 551, 634, 655, 763, 780

**Estimated Time**: 20 minutes

**Pattern**:
```go
// Before
log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)

// After
logging.ErrorWithFields("Failed to get container", map[string]interface{}{
    "error":   err.Error(),
    "user_id": logging.MaskUserID(userIDStr),
})
```

---

## Low Priority Files (Non-Sensitive)

### 6. `/internal/handlers/chat.go` 🟢 LOW RISK
**Lines**: 481, 509, 523, 547, 555
**Why**: Mostly debug output for streaming
**Estimated Time**: 10 minutes

### 7. `/internal/handlers/profile.go` 🟢 LOW RISK
**Lines**: 88, 143
**Why**: Image upload errors (no sensitive data)
**Estimated Time**: 5 minutes

### 8. `/internal/handlers/voice_notes.go` 🟢 LOW RISK
**Lines**: 118, 168
**Why**: Transcription errors
**Estimated Time**: 5 minutes

### 9. Other Handler Files 🟢 LOW RISK
**Files**: contexts.go, projects.go, nodes.go, daily_logs.go, team.go
**Why**: General errors, no sensitive data
**Estimated Time**: 15 minutes total

---

## Configuration Update

### 10. `/cmd/server/main.go` 🔴 NOT STARTED
**Why**: Need to initialize logging configuration

**Add to main() function**:
```go
import "github.com/rhl/businessos-backend/internal/logging"

func main() {
    // Initialize logging (early in main function)
    initLogging()

    // ... rest of main
}

func initLogging() {
    config := &logging.LogConfig{
        Format:              getEnv("LOG_FORMAT", "json"),
        MinLevel:            getLogLevel(),
        MaskSensitiveData:   true, // Always true
        SessionIDMaskLength: 8,
        FilterTerminalIO:    true,
    }

    logging.SetGlobalConfig(config)
    logging.Info("Logging initialized with format: %s", config.Format)
}

func getLogLevel() logging.LogLevel {
    level := getEnv("LOG_LEVEL", "info")
    switch level {
    case "debug":
        return logging.LevelDebug
    case "warn":
        return logging.LevelWarn
    case "error":
        return logging.LevelError
    case "security":
        return logging.LevelSecurity
    default:
        return logging.LevelInfo
    }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Estimated Time**: 10 minutes

---

## Testing Checklist

### Unit Tests
- [x] ✅ Email masking tests
- [x] ✅ Token masking tests
- [x] ✅ Secret detection tests
- [x] ✅ SQL sanitization tests
- [x] ✅ URL sanitization tests
- [x] ✅ Cookie sanitization tests
- [x] ✅ Structured logging tests
- [x] ✅ Concurrent logging tests

### Integration Tests
- [ ] 🔴 Test logging in HTTP middleware
- [ ] 🔴 Test security event aggregation
- [ ] 🔴 Test log format switching (text/JSON)
- [ ] 🔴 Test log level filtering

### Manual Testing
- [ ] 🔴 Verify no sensitive data in production logs
- [ ] 🔴 Verify log aggregation works (if using)
- [ ] 🔴 Verify performance impact is minimal
- [ ] 🔴 Verify JSON format is valid

---

## Environment Variables

Add to `.env` or deployment config:

```bash
# Logging Configuration
LOG_FORMAT=json              # json or text
LOG_LEVEL=info               # debug, info, warn, error, security
```

---

## Progress Summary

### Completed ✅
- Core logging infrastructure
- Comprehensive test suite (25+ tests)
- Documentation (3 guides)
- Reference implementation (redis_auth_sanitized.go)
- Build verification

### In Progress 🟡
- None currently

### Not Started 🔴
1. redis_auth.go migration (15 min)
2. auth_google.go migration (20 min)
3. terminal.go improvements (15 min)
4. filesystem.go migration (20 min)
5. main.go configuration (10 min)
6. Other handlers (35 min)

**Total Remaining Work**: ~2 hours

---

## Migration Priority Order

**Recommended order** (security-first):

1. **Update main.go** - Initialize logging (10 min)
2. **Update redis_auth.go** - Session security (15 min)
3. **Update auth_google.go** - Authentication audit (20 min)
4. **Update terminal.go** - Debug logging (15 min)
5. **Update filesystem.go** - File operations (20 min)
6. **Update remaining handlers** - General cleanup (35 min)
7. **Integration testing** - Verify everything (30 min)

**Total**: ~2.5 hours for complete migration

---

## Verification Steps

After each file update:

1. ✅ Run tests: `go test ./internal/logging/... -v`
2. ✅ Build project: `go build ./cmd/server`
3. ✅ Check for sensitive data in sample logs
4. ✅ Verify log format (JSON in prod, text in dev)

---

## Rollback Plan

If issues occur:

1. Revert to previous commit
2. The `logging` package is backward compatible
3. Standard `log` package still works alongside
4. Can migrate file-by-file without breaking changes

---

## Success Criteria

- [ ] All `log.Printf` calls replaced with `logging.*`
- [ ] No sensitive data visible in logs (verified manually)
- [ ] All tests passing
- [ ] Build successful
- [ ] Production logs in JSON format
- [ ] Security events properly logged
- [ ] Performance impact < 1ms per log call

---

## Notes

- **websocket.go is the gold standard** - use it as a reference
- **redis_auth_sanitized.go** shows middleware pattern
- Keep MaskSensitiveData=true even in development
- Test manually with real user data (in safe environment)
- Review logs before production deployment

---

## Questions or Issues?

- See `/internal/logging/QUICK_REFERENCE.md` for syntax
- See `/internal/logging/SANITIZED_LOGGING_GUIDE.md` for details
- See `/internal/terminal/websocket.go` for examples
- Check test files for edge cases

---

**Last Updated**: 2025-12-23
**Status**: Core implementation complete, migration ready to begin
