// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL not set")
		os.Exit(1)
	}

	fmt.Printf("🔗 Testing connection to Supabase...\n")
	fmt.Printf("   URL: %s...%s\n", databaseURL[:30], databaseURL[len(databaseURL)-20:])

	// Parse connection config
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		fmt.Printf("❌ Failed to parse database URL: %v\n", err)
		os.Exit(1)
	}

	// Configure pool
	poolConfig.MaxConns = 2
	poolConfig.MinConns = 1

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Println("⏳ Connecting...")
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		fmt.Printf("❌ Failed to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Ping
	fmt.Println("⏳ Pinging database...")
	if err := pool.Ping(ctx); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Successfully connected to Supabase!")

	// Run a simple query
	var version string
	err = pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		fmt.Printf("❌ Failed to query version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📊 PostgreSQL Version: %s\n", version)

	// Check tables
	fmt.Println("\n📋 Checking existing tables...")
	rows, err := pool.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
		LIMIT 20
	`)
	if err != nil {
		fmt.Printf("❌ Failed to list tables: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	tableCount := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		fmt.Printf("   • %s\n", tableName)
		tableCount++
	}
	
	if tableCount == 0 {
		fmt.Println("   (no tables found)")
	} else {
		fmt.Printf("\n✅ Found %d tables\n", tableCount)
	}
}
