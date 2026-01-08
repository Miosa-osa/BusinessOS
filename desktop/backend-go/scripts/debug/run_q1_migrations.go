package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║        Running Q1 Completion Migrations (029-030)            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	// Migration 029: project_members
	fmt.Println("📝 Migration 029: Project Members (project-level access control)")
	fmt.Println("-----------------------------------------------------------")
	m029, err := os.ReadFile("internal/database/migrations/029_project_members.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 029: %v", err)
	}

	_, err = pool.Exec(ctx, string(m029))
	if err != nil {
		if isAlreadyExists(err) {
			fmt.Println("⚠️  Migration 029 already applied")
		} else {
			log.Fatalf("Failed to apply migration 029: %v", err)
		}
	} else {
		fmt.Println("✅ Migration 029 applied successfully")
	}
	fmt.Println("")

	// Migration 030: memory_hierarchy
	fmt.Println("📝 Migration 030: Memory Hierarchy System")
	fmt.Println("-----------------------------------------------------------")
	m030, err := os.ReadFile("internal/database/migrations/030_memory_hierarchy.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 030: %v", err)
	}

	_, err = pool.Exec(ctx, string(m030))
	if err != nil {
		if isAlreadyExists(err) {
			fmt.Println("⚠️  Migration 030 already applied")
		} else {
			log.Fatalf("Failed to apply migration 030: %v", err)
		}
	} else {
		fmt.Println("✅ Migration 030 applied successfully")
	}
	fmt.Println("")

	// Verify tables
	fmt.Println("🔍 Verifying tables...")
	fmt.Println("-----------------------------------------------------------")
	tables := []string{"project_members", "project_role_definitions"}
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = $1
			)
		`, table).Scan(&exists)

		if err != nil {
			log.Printf("❌ Error checking table %s: %v", table, err)
			continue
		}

		if exists {
			fmt.Printf("✅ Table '%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%s' NOT found\n", table)
		}
	}

	// Check workspace_memories columns
	var hasVisibility bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.columns
			WHERE table_name = 'workspace_memories'
			AND column_name = 'visibility'
		)
	`).Scan(&hasVisibility)

	if hasVisibility {
		fmt.Println("✅ workspace_memories extended with visibility/hierarchy")
	} else {
		fmt.Println("❌ workspace_memories visibility column NOT found")
	}

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   MIGRATIONS COMPLETE                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "already exists") ||
		contains(errMsg, "duplicate") ||
		(contains(errMsg, "relation") && contains(errMsg, "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
