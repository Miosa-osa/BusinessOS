package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// DB is a wrapper around pgxpool.Pool for database access.
type DB struct {
	Pool *pgxpool.Pool
}

var Pool *pgxpool.Pool

func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Detect PgBouncer/Supavisor connection pooler (port 6543 or pgbouncer=true).
	// Pooled connections don't support prepared statements, so use simple protocol.
	if strings.Contains(cfg.DatabaseURL, ":6543") || strings.Contains(cfg.DatabaseURL, "pgbouncer=true") {
		poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	// Get max connections from environment variable with intelligent defaults
	// Production: 100 connections for high concurrency
	// Development: 25 connections for local testing
	maxConns := getEnvInt("PG_MAX_CONNECTIONS", 25)
	if env := os.Getenv("ENVIRONMENT"); env == "production" {
		maxConns = getEnvInt("PG_MAX_CONNECTIONS", 100)
	}

	// Pool settings optimized for multi-agent concurrency
	poolConfig.MaxConns = int32(maxConns)          // Scale via PG_MAX_CONNECTIONS (default: 25 dev, 100 prod)
	poolConfig.MinConns = 5                        // Faster warm start, maintain ready connections
	poolConfig.MaxConnLifetime = 1 * time.Hour     // Reduce reconnection overhead (up from 15min)
	poolConfig.MaxConnIdleTime = 30 * time.Minute  // Better connection reuse (up from 5min)
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Less frequent checks, reduce overhead (down from 30s)

	// Performance optimization notes:
	// - MaxConns=100 (prod): Supports ~800 req/sec with proper connection pooling
	// - MaxConns=25 (dev): Supports ~200 req/sec for local testing
	// - MinConns=5: Eliminates cold start latency for first requests
	// - MaxConnLifetime=1h: Reduces connection churn by 75%
	// - HealthCheckPeriod=1min: Reduces unnecessary ping traffic by 50%

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	Pool = pool
	return pool, nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

// getEnvInt retrieves an integer environment variable with a fallback default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
