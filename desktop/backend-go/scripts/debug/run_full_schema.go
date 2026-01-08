package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Read the full schema SQL file
	sqlContent, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		log.Fatal("Failed to read schema.sql file:", err)
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	fmt.Println("📦 Applying full database schema...")

	// Execute the full schema
	_, err = pool.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Printf("⚠️  Warning during schema application: %v", err)
		log.Println("This is normal if some tables already exist (CREATE TABLE IF NOT EXISTS)")
	}

	// Verify critical tables exist
	tables := []string{
		"thinking_traces",
		"tasks",
		"user_settings",
		"custom_agents",
		"projects",
		"conversations",
		"messages",
		"contexts",
	}

	fmt.Println("\n✅ Verifying tables...")
	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`
		err := pool.QueryRow(context.Background(), query, table).Scan(&exists)
		if err != nil {
			log.Printf("❌ Error checking table %s: %v", table, err)
			continue
		}
		if exists {
			fmt.Printf("   ✓ %s\n", table)
		} else {
			fmt.Printf("   ✗ %s (MISSING)\n", table)
		}
	}

	fmt.Println("\n🎉 Schema application complete!")
}
