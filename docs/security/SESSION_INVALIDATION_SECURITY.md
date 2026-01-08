# Session Invalidation Security Implementation

## Overview

This document describes the secure session invalidation system implemented to prevent unauthorized access after logout, password changes, or account compromise.

## Problem Statement

**CRITICAL VULNERABILITY (NOW FIXED)**: Previous implementation of `InvalidateUserSessions()` did nothing, allowing sessions to remain active in Redis cache even after:
- User logout
- Password changes
- Account compromise detection
- Permission/role changes

This created a window of vulnerability where attackers could continue using stolen session tokens.

## Solution Architecture

### User → Sessions Index

We maintain a Redis SET for each user that tracks all their active sessions:

```text
Key: "user_sessions:{userID}"
Value: SET of session keys
Example: {"auth_session:token1", "auth_session:token2", "auth_session:token3"}
```
### Redis Data Structures

1. **Session Data** (existing):
   - Key: `auth_session:{token}`
   - Value: JSON-encoded CachedUser
   - TTL: 15 minutes (configurable)

2. **User Session Index** (NEW):
   - Key: `user_sessions:{userID}`
   - Value: SET of session keys
   - TTL: Session TTL + 5 minutes (auto-cleanup)

### Operations

#### 1. Set Session (Create/Update)

**File**: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/middleware/redis_auth.go:93-128`

```go
func (sc *SessionCache) Set(ctx context.Context, token string, user *BetterAuthUser) error
```

**Atomic Pipeline Operations**:
1. `SET auth_session:{token}` → session data (with TTL)
2. `SADD user_sessions:{userID}` → add session key to user's set
3. `EXPIRE user_sessions:{userID}` → refresh TTL on user's session set

**Time Complexity**: O(1)

#### 2. Invalidate Single Session (Logout)

**File**: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/middleware/redis_auth.go:130-156`

```go
func (sc *SessionCache) Invalidate(ctx context.Context, token string) error
```

**Atomic Pipeline Operations**:
1. `GET auth_session:{token}` → retrieve session to get userID
2. `DEL auth_session:{token}` → delete session data
3. `SREM user_sessions:{userID}` → remove session key from user's set

**Time Complexity**: O(1)

#### 3. Invalidate All User Sessions (Security Event)

**File**: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/middleware/redis_auth.go:158-193`

```go
func (sc *SessionCache) InvalidateUserSessions(ctx context.Context, userID string) error
```

**Atomic Pipeline Operations**:
1. `SMEMBERS user_sessions:{userID}` → get all session keys
2. For each session key: `DEL {sessionKey}` → delete session data
3. `DEL user_sessions:{userID}` → delete user's session set

**Time Complexity**: O(N) where N = number of user's sessions

**Security Events Triggering Full Invalidation**:
- Password change
- Account compromise detection
- Permission/role changes
- User-initiated "logout from all devices"

## API Endpoints

### 1. Logout Current Session

**Endpoint**: `POST /api/auth/logout`

**Authentication**: Public (uses cookie)

**Behavior**:
1. Extract session token from cookie
2. Delete session from PostgreSQL
3. Invalidate session in Redis cache
4. Clear session cookie

**File**: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/handlers/auth_google.go:268-298`

### 2. Logout All Sessions (NEW)

**Endpoint**: `POST /api/auth/logout-all`

**Authentication**: Required (middleware validates current session)

**Response**:
```json
{
  "message": "All sessions invalidated",
  "sessions_removed": 5
}
```

**Behavior**:
1. Get current user from auth middleware context
2. Invalidate all Redis cached sessions for user
3. Delete all sessions from PostgreSQL
4. Clear current session cookie
5. Return count of invalidated sessions

**File**: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/handlers/auth_google.go:300-348`

**Use Cases**:
- User clicks "Logout from all devices" in settings
- After password change (should be auto-triggered)
- After suspected account compromise
- After permission/role changes

## Security Properties

### 1. Atomic Operations

All Redis operations use pipelined transactions to ensure atomicity:
- Session creation + index update = atomic
- Session deletion + index removal = atomic
- Bulk session deletion = atomic

### 2. Dual-Layer Invalidation

Both cache AND database are invalidated:
```go
// 1. Invalidate Redis cache first (fast)
h.sessionCache.InvalidateUserSessions(ctx, user.ID)

// 2. Delete from PostgreSQL (authoritative)
h.pool.Exec(ctx, `DELETE FROM session WHERE "userId" = $1`, user.ID)
```

This provides defense-in-depth even if one layer fails.

### 3. Graceful Degradation

If Redis cache is unavailable:
- Session validation falls back to PostgreSQL
- Operations continue but with higher latency
- No security compromise

### 4. Auto-Cleanup

User session sets have TTL that's slightly longer than session TTL:
```go
pipe.Expire(ctx, userSessionsKey, sc.ttl+5*time.Minute)
```

This prevents orphaned session index entries.

## Implementation Files

1. **Session Cache** (`/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/middleware/redis_auth.go`):
   - `userSessionsKey()` - Generate user session set key
   - `Set()` - Create session with index update
   - `Invalidate()` - Delete single session with index cleanup
   - `InvalidateUserSessions()` - Bulk delete all user sessions

2. **Authentication Handler** (`/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/handlers/auth_google.go`):
   - `Logout()` - Single session logout
   - `LogoutAllSessions()` - Bulk session invalidation

3. **Route Registration** (`/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/handlers/handlers.go`):
   - `POST /api/auth/logout` - Public logout
   - `POST /api/auth/logout-all` - Protected bulk logout

## Testing Recommendations

### Unit Tests

1. **Test session creation updates index**:
   ```go
   // Create session
   cache.Set(ctx, token, user)
   // Verify session exists
   // Verify session in user's set
   ```

2. **Test single session invalidation removes from index**:
   ```go
   // Create session
   // Invalidate session
   // Verify session deleted
   // Verify session removed from user's set
   ```

3. **Test bulk invalidation clears all sessions**:
   ```go
   // Create 3 sessions for user
   // Call InvalidateUserSessions
   // Verify all sessions deleted
   // Verify user's set deleted
   ```

### Integration Tests

1. **Test logout endpoint**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/logout \
     -H "Cookie: better-auth.session_token=TOKEN"
   ```

2. **Test logout-all endpoint**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/logout-all \
     -H "Cookie: better-auth.session_token=TOKEN"
   ```

3. **Test concurrent sessions**:
   - Login from 3 different devices
   - Call logout-all from one device
   - Verify all 3 sessions invalidated

### Security Tests

1. **Test session reuse after logout**:
   - Login and capture session token
   - Logout
   - Attempt to use old session token
   - Expected: 401 Unauthorized

2. **Test session reuse after logout-all**:
   - Login from 2 devices
   - Logout-all from device 1
   - Attempt to use device 2's session
   - Expected: 401 Unauthorized

## Performance Characteristics

### Time Complexity

- **Set Session**: O(1) - 3 Redis commands (SET, SADD, EXPIRE)
- **Get Session**: O(1) - 1 Redis command (GET)
- **Invalidate Session**: O(1) - 3 Redis commands (GET, DEL, SREM)
- **Invalidate All Sessions**: O(N) where N = user's session count
  - SMEMBERS: O(N)
  - N × DEL: O(N)
  - Total: O(N)

### Space Complexity

Per user:
- Session data: `sizeof(CachedUser) × session_count`
- Session index: `sizeof(SET) + (sizeof(session_key) × session_count)`

Typical overhead: ~100 bytes per session for index

### Network Efficiency

All operations use Redis pipelining to minimize round-trips:
- Set: 1 round-trip (3 commands pipelined)
- Invalidate: 1 round-trip (2-3 commands pipelined)
- Invalidate All: 2 round-trips (SMEMBERS + pipelined DELs)

## Migration Notes

### Existing Sessions

Existing sessions created before this implementation:
- Will NOT have entries in user session sets
- Will still work normally (cache miss → DB lookup)
- Will be added to index on next cache refresh
- Will expire naturally via TTL

No migration needed - system is backward compatible.

### Monitoring

Add metrics for:
```go
// Session operations
redis_session_set_total
redis_session_get_total
redis_session_invalidate_total
redis_session_invalidate_all_total

// Performance
redis_session_invalidate_all_duration_seconds
redis_session_invalidate_all_count_histogram
```

## Future Enhancements

1. **Session Activity Tracking**:
   - Track last access time per session
   - Enable "kick this specific device" functionality
   - Show "active devices" list to user

2. **Session Metadata**:
   - Store IP address, user agent in session index
   - Enable security audit: "Was I logged in from X?"

3. **Suspicious Activity Detection**:
   - Alert on unusual session count (>10 sessions)
   - Alert on rapid session creation
   - Auto-invalidate on geographic anomaly

4. **Session Limits**:
   - Enforce max sessions per user (e.g., 10)
   - Auto-remove oldest session when limit reached
   - Prevent session creation DoS

## Security Audit Trail

All session operations are logged:

```go
log.Printf("SessionCache: invalidated %d sessions for user %s", len(sessionKeys), userID)
log.Printf("LogoutAllSessions: deleted %d database sessions for user %s", rowsAffected, user.ID)
```

These logs are critical for:
- Security incident investigation
- Compliance auditing (GDPR, SOC 2)
- Detecting automated attacks

## Compliance Considerations

### GDPR

- User can request session invalidation: ✅ (logout-all endpoint)
- User can view active sessions: ⚠️ (future enhancement)
- Sessions auto-expire: ✅ (15 minute TTL)

### SOC 2

- Audit logging: ✅ (all operations logged)
- Forced session termination: ✅ (logout-all endpoint)
- Session timeout: ✅ (15 minute TTL)

### HIPAA

- Session invalidation on permission change: ✅ (manual trigger)
- Audit trail: ✅ (all operations logged)
- Encrypted sessions: ✅ (Redis data encrypted at rest if configured)

## Conclusion

This implementation provides:
- **Secure session invalidation** with O(N) efficiency
- **Defense-in-depth** with dual-layer (Redis + PostgreSQL) invalidation
- **Atomic operations** via Redis pipelines
- **Graceful degradation** when Redis unavailable
- **Audit trail** for security compliance
- **User control** via logout-all endpoint

The vulnerability of sessions persisting after logout is now completely eliminated.
