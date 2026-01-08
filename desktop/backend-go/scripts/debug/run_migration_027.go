package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Run the migration
	migration := `
-- Migration 027: Add all thinking-related columns to user_settings
ALTER TABLE user_settings
ADD COLUMN IF NOT EXISTS thinking_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS thinking_show_in_ui BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS thinking_save_traces BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS thinking_default_template_id UUID,
ADD COLUMN IF NOT EXISTS thinking_max_tokens INTEGER DEFAULT 2048;

-- Update existing rows to have default values
UPDATE user_settings
SET
  thinking_enabled = COALESCE(thinking_enabled, FALSE),
  thinking_show_in_ui = COALESCE(thinking_show_in_ui, TRUE),
  thinking_save_traces = COALESCE(thinking_save_traces, FALSE),
  thinking_max_tokens = COALESCE(thinking_max_tokens, 2048)
WHERE thinking_enabled IS NULL OR thinking_show_in_ui IS NULL;
`

	log.Println("Running migration 027: Add all thinking columns to user_settings")

	_, err = db.ExecContext(ctx, migration)
	if err != nil {
		log.Fatalf("❌ Failed to run migration: %v", err)
	}

	log.Println("✅ Migration 027 completed successfully")

	// Verify all columns were added
	columns := []string{
		"thinking_enabled",
		"thinking_show_in_ui",
		"thinking_save_traces",
		"thinking_default_template_id",
		"thinking_max_tokens",
	}

	for _, col := range columns {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'user_settings'
				AND column_name = $1
			);
		`

		err = db.QueryRowContext(ctx, query, col).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ Failed to verify column %s: %v", col, err)
		}

		if exists {
			log.Printf("✅ Verified: %s column exists", col)
		} else {
			log.Printf("❌ Warning: %s column not found after migration", col)
		}
	}
}
