# Sandbox Quota Service Integration Guide

## Overview

The `QuotaService` enforces per-user resource limits for sandbox containers. It validates resource requests before sandbox deployment to ensure users stay within their allocated quotas.

## Files Created

1. **internal/services/sandbox_quota.go** - Main quota management service
2. **internal/services/sandbox_quota_test.go** - Comprehensive test suite
3. **internal/config/config.go** - Updated with quota configuration fields

## Architecture

```
SandboxDeploymentService.Deploy()
    ↓
QuotaService.CheckQuota()  ← Validates before deployment
    ↓
Container deployment proceeds (if quota check passes)
```

## Integration with SandboxDeploymentService

### Current Implementation

The `SandboxDeploymentService` already has a basic quota check:

```go
// Line 146 in sandbox_deployment.go
if err := s.checkUserQuota(ctx, req.UserID); err != nil {
    return nil, err
}
```

This currently only checks sandbox count. Replace it with the comprehensive quota service.

### Recommended Integration Steps

#### 1. Add QuotaService to SandboxDeploymentService

```go
type SandboxDeploymentService struct {
    portAllocator    *SandboxPortAllocator
    containerManager *container.AppContainerManager
    quotaService     *QuotaService  // ADD THIS
    pool             *pgxpool.Pool
    queries          *sqlc.Queries
    config           *config.Config
    logger           *slog.Logger
    mu               sync.Mutex
    inProgress       map[uuid.UUID]bool
}
```

#### 2. Initialize QuotaService in Constructor

```go
func NewSandboxDeploymentService(
    pool *pgxpool.Pool,
    dockerClient *client.Client,
    cfg *config.Config,
    logger *slog.Logger,
) (*SandboxDeploymentService, error) {
    // ... existing code ...

    // Create quota service
    quotaService := NewQuotaService(pool, cfg, logger)

    service := &SandboxDeploymentService{
        portAllocator:    portAllocator,
        containerManager: containerManager,
        quotaService:     quotaService,  // ADD THIS
        pool:             pool,
        queries:          sqlc.New(pool),
        config:           cfg,
        logger:           logger.With("service", "sandbox_deployment"),
        inProgress:       make(map[uuid.UUID]bool),
    }

    return service, nil
}
```

#### 3. Replace checkUserQuota with Comprehensive Check

Replace the existing `checkUserQuota` method (lines 467-484) with:

```go
// checkUserQuota verifies the user hasn't exceeded their sandbox quota.
func (s *SandboxDeploymentService) checkUserQuota(
    ctx context.Context,
    userID uuid.UUID,
    req SandboxDeploymentRequest,
) error {
    // Build quota request from deployment request
    quotaReq := QuotaRequest{
        SandboxCount:      1,
        MemoryPerSandbox:  req.MemoryLimit,
        CPUPerSandbox:     req.CPUQuota,
        StoragePerSandbox: 0, // TODO: Calculate from workspace size
    }

    // Use defaults if not specified
    if quotaReq.MemoryPerSandbox == 0 {
        quotaReq.MemoryPerSandbox = s.config.SandboxDefaultMemory
    }
    if quotaReq.CPUPerSandbox == 0 {
        quotaReq.CPUPerSandbox = int64(s.config.SandboxDefaultCPU)
    }

    // Validate the request
    if err := ValidateQuotaRequest(quotaReq); err != nil {
        s.logger.Warn("invalid quota request", "user_id", userID, "error", err)
        return err
    }

    // Check quota
    if err := s.quotaService.CheckQuota(ctx, userID, quotaReq); err != nil {
        s.logger.Info("quota check failed",
            "user_id", userID,
            "error", err)
        return err
    }

    return nil
}
```

#### 4. Update Deploy Method Call

Change line 146 in `Deploy()` method:

```go
// OLD:
if err := s.checkUserQuota(ctx, req.UserID); err != nil {
    return nil, err
}

// NEW:
if err := s.checkUserQuota(ctx, req.UserID, req); err != nil {
    return nil, err
}
```

## Quota Limits

### Default Limits (Configurable via Environment Variables)

| Resource | Default | Config Key | Description |
|----------|---------|------------|-------------|
| Max Sandboxes | 5 | `SANDBOX_MAX_PER_USER` | Max concurrent sandboxes per user |
| Memory per Sandbox | 512MB | `SANDBOX_DEFAULT_MEMORY` | Max memory per sandbox (bytes) |
| CPU per Sandbox | 50% | `SANDBOX_DEFAULT_CPU` | Max CPU quota (50000 = 50% of 1 CPU) |
| Total Memory | 2GB | `SANDBOX_MAX_TOTAL_MEMORY` | Total memory across all sandboxes |
| Total Storage | 5GB | `SANDBOX_MAX_TOTAL_STORAGE` | Total workspace storage |

### Environment Variable Configuration

Add to `.env` file:

```bash
# Sandbox Quota Configuration
SANDBOX_MAX_PER_USER=5
SANDBOX_DEFAULT_MEMORY=536870912      # 512MB in bytes
SANDBOX_DEFAULT_CPU=50000             # 50% of 1 CPU
SANDBOX_MAX_TOTAL_MEMORY=2147483648   # 2GB in bytes
SANDBOX_MAX_TOTAL_STORAGE=5368709120  # 5GB in bytes
```

### Admin Quota Overrides

Admins can set custom quotas for specific users:

```go
// Example: Grant premium user higher limits
premiumQuota := UserQuota{
    MaxSandboxes:        20,
    MaxMemoryPerSandbox: 2 * 1024 * 1024 * 1024, // 2GB
    MaxCPUPerSandbox:    200000,                 // 200% (2 CPUs)
    MaxTotalMemory:      16 * 1024 * 1024 * 1024, // 16GB
    MaxTotalStorage:     50 * 1024 * 1024 * 1024, // 50GB
}

err := quotaService.SetUserQuotaOverride(ctx, premiumUserID, premiumQuota)
```

## Error Handling

The quota service returns specific errors that can be handled differently:

```go
err := quotaService.CheckQuota(ctx, userID, request)
if err != nil {
    switch {
    case errors.Is(err, ErrMaxSandboxesReached):
        // User has too many sandboxes running
        return fmt.Errorf("you have reached your sandbox limit (%d)", quota.MaxSandboxes)

    case errors.Is(err, ErrMemoryLimitExceeded):
        // Memory quota exceeded
        return fmt.Errorf("insufficient memory quota available")

    case errors.Is(err, ErrCPULimitExceeded):
        // CPU quota exceeded
        return fmt.Errorf("insufficient CPU quota available")

    case errors.Is(err, ErrStorageLimitExceeded):
        // Storage quota exceeded
        return fmt.Errorf("insufficient storage quota available")

    case errors.Is(err, ErrInvalidQuotaRequest):
        // Invalid request parameters
        return fmt.Errorf("invalid resource request: %w", err)

    default:
        return fmt.Errorf("quota check failed: %w", err)
    }
}
```

## API Usage Examples

### Check Quota Before Deployment

```go
quotaReq := QuotaRequest{
    SandboxCount:      1,
    MemoryPerSandbox:  512 * 1024 * 1024,  // 512MB
    CPUPerSandbox:     50000,               // 50%
    StoragePerSandbox: 1024 * 1024 * 1024,  // 1GB
}

err := quotaService.CheckQuota(ctx, userID, quotaReq)
if err != nil {
    // Handle quota error
    return nil, err
}

// Proceed with deployment
```

### Get User's Current Usage

```go
usage, err := quotaService.GetUserUsage(ctx, userID)
if err != nil {
    return err
}

fmt.Printf("User has %d/%d sandboxes running\n",
    usage.CurrentSandboxes,
    quota.MaxSandboxes)
fmt.Printf("Memory: %d/%d bytes\n",
    usage.CurrentTotalMemory,
    quota.MaxTotalMemory)
```

### Get User's Quota Limits

```go
quota, err := quotaService.GetUserQuota(ctx, userID)
if err != nil {
    return err
}

fmt.Printf("Max sandboxes: %d\n", quota.MaxSandboxes)
fmt.Printf("Max memory per sandbox: %d MB\n", quota.MaxMemoryPerSandbox/(1024*1024))
fmt.Printf("Is override: %v\n", quota.IsOverride)
```

## Future Enhancements

1. **Database Persistence**: Store quota overrides in database for durability across restarts
2. **Usage Tracking**: Store historical usage data for analytics
3. **Dynamic Quotas**: Adjust quotas based on user tier/subscription
4. **Quota API Endpoints**: REST endpoints for users to view their quota/usage
5. **Storage Calculation**: Actual workspace storage usage tracking
6. **Quota Notifications**: Alert users when approaching quota limits
7. **Billing Integration**: Usage-based billing based on quota consumption

## Testing

Run tests:

```bash
# Test quota service
go test -v ./internal/services -run TestQuota

# Test all sandbox services
go test -v ./internal/services -run TestSandbox
```

## Notes

- Quota checks are **non-blocking** if database is unavailable (fail-open for availability)
- Admin overrides are **in-memory only** (currently) - persist to database for production
- Storage quota is **not yet enforced** - requires workspace size tracking
- CPU quota uses Docker's CPU quota format (100000 = 1 CPU)

## Status

✅ **QuotaService Implementation**: Complete
✅ **Comprehensive Test Suite**: Complete
✅ **Config Integration**: Complete
⏳ **SandboxDeploymentService Integration**: Ready for implementation
⏳ **API Endpoints**: Not yet implemented
⏳ **Database Persistence**: Not yet implemented
