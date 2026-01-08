package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database URL from environment or hardcoded
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres"
	}

	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Read migration file
	migrationSQL, err := os.ReadFile("internal/database/migrations/032_fix_thinking_traces_user_id.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	fmt.Println("Running migration 032_fix_thinking_traces_user_id.sql...")
	fmt.Println("This will change user_id columns from UUID to TEXT in thinking system tables")
	fmt.Println()

	// Execute migration
	_, err = pool.Exec(ctx, string(migrationSQL))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println()
	fmt.Println("Changed columns:")
	fmt.Println("  - thinking_traces.user_id: UUID -> TEXT")
	fmt.Println("  - reasoning_templates.user_id: UUID -> TEXT")
}
