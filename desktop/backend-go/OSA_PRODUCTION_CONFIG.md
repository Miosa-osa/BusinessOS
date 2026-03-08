# OSA-5 Production Configuration

**Add these environment variables to your `.env.production` file:**

```bash
# ===========================================
# OSA-5 INTEGRATION (App Generation Service)
# ===========================================
# Enable/disable OSA-5 integration
OSA_ENABLED=true

# OSA-5 API base URL (production endpoint)
# Development: http://localhost:8089
# Production: https://osa-prod.your-domain.com
OSA_BASE_URL=https://osa-prod.your-domain.com

# Shared secret for OSA API authentication
# Generate with: openssl rand -base64 32
OSA_SHARED_SECRET=your-strong-osa-shared-secret-change-this

# Request timeout (in seconds)
OSA_TIMEOUT=30

# Circuit breaker configuration (failures before circuit opens)
OSA_MAX_FAILURES=5

# Circuit breaker timeout (seconds before attempting half-open)
OSA_CIRCUIT_TIMEOUT=30

# Response cache TTL (seconds) for fallback strategy
OSA_CACHE_TTL=300

# Request queue size for auto-retry during outages
OSA_QUEUE_SIZE=1000

# Enable auto-recovery loop (attempts to recover from circuit break)
OSA_AUTO_RECOVERY=true
```

## Configuration Details

### OSA_BASE_URL
- **Development:** `http://localhost:8089` (default if not set)
- **Staging:** `https://osa-staging.your-domain.com`
- **Production:** `https://osa-prod.your-domain.com`

### OSA_SHARED_SECRET
Generate a strong secret:
```bash
openssl rand -base64 32
```
This is used for API authentication between BusinessOS and OSA-5.

### Circuit Breaker Settings
- `OSA_MAX_FAILURES`: Number of consecutive failures before circuit opens (default: 5)
- `OSA_CIRCUIT_TIMEOUT`: Seconds before attempting half-open state (default: 30)

### Fallback Strategy
- `OSA_CACHE_TTL`: How long to cache OSA responses for fallback (default: 300s = 5 minutes)
- When OSA is unavailable, cached responses are returned (FallbackStale strategy)

### Auto-Recovery
- `OSA_AUTO_RECOVERY`: Enable background loop that retries failed requests (default: true)
- `OSA_QUEUE_SIZE`: Maximum queued requests during outages (default: 1000)

## Resilience Features

The OSA integration includes production-grade resilience:

1. **Circuit Breaker**: Prevents cascading failures
2. **Fallback Strategy**: Returns cached data when service unavailable
3. **Auto-Recovery**: Background loop retries failed requests
4. **Request Queue**: Queues requests during outages
5. **Health Check Caching**: Prevents hammering health endpoints

See `/Users/oso/BusinessOS/desktop/backend-go/internal/integrations/osa/` for implementation details.

## Deployment Checklist

- [ ] Set `OSA_BASE_URL` to production endpoint
- [ ] Generate and set `OSA_SHARED_SECRET`
- [ ] Verify OSA service is accessible from backend
- [ ] Test health check: `curl $OSA_BASE_URL/health`
- [ ] Monitor circuit breaker metrics
- [ ] Set up alerts for circuit open events

## Environment-Specific Configurations

### Development
```bash
OSA_ENABLED=true
OSA_BASE_URL=http://localhost:8089
OSA_TIMEOUT=30
```

### Staging
```bash
OSA_ENABLED=true
OSA_BASE_URL=https://osa-staging.your-domain.com
OSA_SHARED_SECRET=<staging-secret>
OSA_TIMEOUT=30
OSA_CACHE_TTL=300
```

### Production
```bash
OSA_ENABLED=true
OSA_BASE_URL=https://osa-prod.your-domain.com
OSA_SHARED_SECRET=<production-secret>
OSA_TIMEOUT=30
OSA_MAX_FAILURES=5
OSA_CIRCUIT_TIMEOUT=30
OSA_CACHE_TTL=300
OSA_QUEUE_SIZE=1000
OSA_AUTO_RECOVERY=true
```
