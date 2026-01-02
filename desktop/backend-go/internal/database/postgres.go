package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

var Pool *pgxpool.Pool

func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure pool settings for Supabase (cross-cloud optimized)
	// IMPORTANT: These settings are optimized for Supabase connection pooling (Supavisor)
	// Use the pooled connection string (port 6543), NOT direct connection (port 5432)
	poolConfig.MaxConns = 10              // Conservative for cross-cloud latency
	poolConfig.MinConns = 2               // Keep some connections warm
	poolConfig.MaxConnLifetime = 15 * time.Minute  // Supabase closes stale connections
	poolConfig.MaxConnIdleTime = 5 * time.Minute   // Release idle connections faster
	poolConfig.HealthCheckPeriod = 30 * time.Second  // More frequent health checks for cross-cloud

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
