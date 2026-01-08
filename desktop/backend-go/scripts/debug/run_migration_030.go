package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Running Migration 030                            ║")
	fmt.Println("║           Memory Hierarchy System                             ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	migration, err := os.ReadFile("internal/database/migrations/030_memory_hierarchy.sql")
	if err != nil {
		log.Fatalf("Failed to read migration: %v", err)
	}

	_, err = pool.Exec(context.Background(), string(migration))
	if err != nil {
		log.Fatalf("Failed to apply migration: %v", err)
	}

	fmt.Println("✅ Migration 030 applied successfully!")
	fmt.Println("")

	// Verify
	var hasVisibility bool
	pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.columns
			WHERE table_name = 'workspace_memories'
			AND column_name = 'visibility'
		)
	`).Scan(&hasVisibility)

	if hasVisibility {
		fmt.Println("✅ workspace_memories extended with visibility/hierarchy columns")
	}

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  MIGRATION COMPLETE                           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}
