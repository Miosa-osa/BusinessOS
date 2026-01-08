package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Read SQL file
	sqlContent, err := os.ReadFile("fix_custom_agents.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
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

	fmt.Println("📦 Creating custom_agents and agent_presets tables...")

	// Execute SQL
	_, err = pool.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Fatal("Failed to execute SQL:", err)
	}

	// Verify
	tables := []string{"custom_agents", "agent_presets"}
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

			// Count rows
			var count int
			pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("     (%d rows)\n", count)
		} else {
			fmt.Printf("   ✗ %s (MISSING)\n", table)
		}
	}

	fmt.Println("\n🎉 Custom agents setup complete!")
}
