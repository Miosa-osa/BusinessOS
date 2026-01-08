package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
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

	migrations := []struct {
		name string
		path string
	}{
		{"Thinking System", "internal/database/migrations/008_thinking_system.sql"},
		{"Custom Agents", "internal/database/migrations/009_custom_agents.sql"},
	}

	for _, migration := range migrations {
		fmt.Printf("📦 Applying %s...\n", migration.name)

		sqlContent, err := os.ReadFile(migration.path)
		if err != nil {
			log.Printf("❌ Failed to read %s: %v\n", migration.path, err)
			continue
		}

		_, err = pool.Exec(context.Background(), string(sqlContent))
		if err != nil {
			log.Printf("⚠️  Warning applying %s: %v\n", migration.name, err)
		} else {
			fmt.Printf("✅ %s applied successfully\n", migration.name)
		}
	}

	// Verify the tables exist
	fmt.Println("\n🔍 Verifying missing tables...")
	tables := []string{
		"thinking_traces",
		"reasoning_templates",
		"custom_agents",
		"agent_presets",
	}

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

	fmt.Println("\n🎉 Migration complete!")
}
