# Production Monitoring & Observability

## Overview

Comprehensive monitoring setup for BusinessOS using Railway service metrics and logs,
Cloudflare analytics, Sentry, structured logging, and health checks.

## Table of Contents

1. [Structured Logging](#structured-logging)
2. [Railway Metrics & Logs](#railway-metrics--logs)
3. [Cloudflare Analytics](#cloudflare-analytics)
4. [Health Endpoints](#health-endpoints)
5. [Uptime Monitoring](#uptime-monitoring)
6. [Alert Policies](#alert-policies)
7. [Performance Metrics](#performance-metrics)
8. [Log Searches](#log-searches)
9. [Notification Channels](#notification-channels)

---

## Structured Logging

### Backend: slog Configuration

The backend already uses `slog` for structured logging. Here's the production configuration:

#### Update `cmd/server/main.go`:

```go
package main

import (
	"context"
	"log/slog"
	"os"
)

func setupLogger(environment string) {
	var handler slog.Handler

	if environment == "production" {
		// Production: JSON to stdout. Railway captures stdout, so JSON keeps
		// logs parseable by `railway logs` and by any log drain.
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Development: Human-readable format
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger initialized",
		"environment", environment,
		"level", handler.Enabled(context.Background(), slog.LevelDebug),
	)
}

func main() {
	environment := os.Getenv("ENVIRONMENT")
	setupLogger(environment)

	// Rest of initialization...
}
```

### Request Logging Middleware

Add request ID and user ID to all logs:

```go
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger logs all HTTP requests with structured data
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user ID if authenticated
		userID := ""
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				if id, ok := u["id"].(string); ok {
					userID = id
				}
			}
		}

		// Log request
		slog.Info("HTTP request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"user_id", userID,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Log errors separately
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				slog.Error("Request error",
					"request_id", requestID,
					"error", err.Error(),
				)
			}
		}
	}
}
```

### Log Critical Events

```go
// Database connection issues
slog.Error("Database connection failed",
	"error", err.Error(),
	"connection_string", redactedURL,
)

// Slow queries
slog.Warn("Slow database query",
	"query", "GetUserWorkspaces",
	"duration_ms", duration.Milliseconds(),
	"user_id", userID,
)

// Failed authentication attempts
slog.Warn("Authentication failed",
	"ip", ip,
	"email", sanitizedEmail,
	"reason", "invalid_password",
)

// Background job failures
slog.Error("Background job failed",
	"job_id", jobID,
	"job_type", jobType,
	"error", err.Error(),
	"retry_count", retryCount,
)
```

---

## Railway Metrics & Logs

Railway provides built-in logs and metrics per service.

```bash
# Stream runtime logs
railway logs --service businessos-api

# Deployment history
railway deployment list --service businessos-api

# Resource metrics (CPU, memory, network, HTTP)
railway metrics --service businessos-api
```

In the dashboard (project `BusinessOS` > service `businessos-api`): the Metrics tab shows
CPU, memory, network, and HTTP request metrics; the Logs tab supports filtering by text.

| Signal | Where |
|--------|-------|
| Build logs | Railway > service > Deployments > build |
| Runtime logs | `railway logs` or Railway > service > Logs |
| CPU / memory | Railway > service > Metrics |
| HTTP request volume / latency | Railway > service > Metrics (HTTP) |
| Postgres metrics | Railway > Postgres service > Metrics |

---

## Cloudflare Analytics

| Signal | Where |
|--------|-------|
| Pages requests, bandwidth, errors | Cloudflare dashboard > Workers & Pages > `businessos-5` > Metrics |
| Pages build/deploy status | Cloudflare dashboard > Workers & Pages > `businessos-5` > Deployments |
| R2 object count / storage | Cloudflare dashboard > R2 > `businessos-downloads` |
| Proxy Function errors | Cloudflare dashboard > Workers & Pages > `businessos-5` > Functions logs |

---

## Health Endpoints

The backend exposes health probes (`cmd/server/routes.go`):

| Endpoint | Meaning |
|----------|---------|
| `GET /health` | Liveness (process is up). Used by CI smoke tests. |
| `GET /healthz` | Liveness alias |
| `GET /ready` / `GET /readyz` | Readiness |
| `GET /health/detailed` | Dependency status: `database`, `redis`, `containers` |

```bash
curl -fsS https://businessos-api-production.up.railway.app/health
curl -fsS https://businessos-api-production.up.railway.app/health/detailed
# {"components":{"database":{"status":"connected"},...},"status":"healthy"}
```

---

## Uptime Monitoring

Use an external uptime provider (Better Stack, UptimeRobot, or Cloudflare Health Checks).
Railway does not provide built-in uptime alerts.

Checks to configure:

| Target | Interval | Alert |
|--------|----------|-------|
| `https://businessos-api-production.up.railway.app/health` | 5 min | down for 2+ checks |
| `https://businessos.dev/api/v1/health` (proxy path) | 5 min | down for 2+ checks |
| `https://businessos.dev` | 5 min | down for 2+ checks |

---

## Alert Policies

Railway exposes usage alerts; application-level alerts come from Sentry.

### 1. High Error Rate (5xx)

- Source: Sentry issue alerts on the backend project.
- Rule: alert when the 5xx rate exceeds 5 errors per minute sustained for 5 minutes.
- Action: email + Slack.

### 2. High Latency (p95)

- Source: Railway HTTP metrics.
- Rule: alert when p95 request latency exceeds 2 seconds for 5 minutes.

### 3. Database Connection Usage

- Source: Railway Postgres metrics.
- Rule: alert when active connections exceed 80% of the instance limit.

### 4. Memory Pressure

- Source: Railway service memory metrics.
- Rule: alert when memory stays above 80% of the service limit.

### 5. Usage / Spend

- Source: Railway workspace usage limits.
- Rule: set a soft limit so a runaway deploy cannot bill without notice.

---

## Performance Metrics

### Exporting Custom Metrics

Prefer OpenTelemetry (OTLP) or structured stdout logs, which Railway and any log drain can
ingest. Do not add a vendor-specific metrics client to the backend.

```go
// Emit a structured metric log line; scrape it from the log drain if needed.
slog.Info("metric",
	"name", "redis.hit_rate",
	"value", hitRate,
	"unit", "ratio",
)
```

### Tracked Metrics

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| HTTP request rate | Counter | Requests per second | N/A |
| HTTP error rate | Counter | 5xx errors per minute | > 5 |
| Request latency | Histogram | p50/p95/p99 latency | p95 > 2s |
| Database connections | Gauge | Active DB connections | > 80% limit |
| Redis hit rate | Gauge | Cache hit percentage | < 90% |
| Redis memory usage | Gauge | Memory utilization | > 80% |
| Active sessions | Gauge | Concurrent user sessions | N/A |
| Background jobs queued | Gauge | Pending jobs | > 100 |
| Background jobs failed | Counter | Failed job executions | > 5/min |

---

## Log Searches

Filter Railway logs for the events that matter. Log lines are JSON, so filter on the message:

```bash
# Failed authentication attempts
railway logs --service businessos-api | grep "Authentication failed"

# Slow database queries
railway logs --service businessos-api | grep "Slow database query"

# Background job failures
railway logs --service businessos-api | grep "Background job failed"
```

---

## Notification Channels

### Email

Configure email notification in Sentry (project alerts) and in the external uptime provider.

### Slack

Point Sentry alert rules and the uptime provider at a Slack webhook (Slack incoming
webhook, or Sentry's native Slack integration).

---

## Monitoring Checklist

### Initial Setup

- [ ] Sentry project created for backend and frontend, DSNs configured
- [ ] Sentry alert rules configured (error rate, new issue)
- [ ] Uptime checks configured (Railway health + proxy + frontend)
- [ ] Notification channels set up (email/Slack)
- [ ] Railway workspace usage limit set

### Post-Deployment

- [ ] No new Sentry issues in the first hour
- [ ] `/health/detailed` reports `database: connected`
- [ ] Railway HTTP metrics show expected traffic
- [ ] p95 latency within budget (< 2s)
- [ ] No 5xx spikes

---

## Resources

- Railway metrics and logs: https://docs.railway.com/guides/metrics
- Cloudflare Pages analytics: https://developers.cloudflare.com/pages/
- Sentry Go: https://docs.sentry.io/platforms/go/
- Sentry SvelteKit: https://docs.sentry.io/platforms/javascript/guides/sveltekit/
- Health checks: see `cmd/server/routes.go` in the backend
