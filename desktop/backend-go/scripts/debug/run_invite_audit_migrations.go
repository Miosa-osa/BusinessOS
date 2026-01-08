package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Running Workspace Invite & Audit Migrations (027-028)      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	// Run migration 027: workspace_invites
	fmt.Println("📝 Running Migration 027: workspace_invites")
	fmt.Println("-----------------------------------------------------------")

	migration027, err := os.ReadFile("internal/database/migrations/027_workspace_invites.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 027: %v", err)
	}

	_, err = pool.Exec(ctx, string(migration027))
	if err != nil {
		// Check if it's already applied
		if isAlreadyExists(err) {
			fmt.Println("⚠️  Migration 027 already applied (table exists)")
		} else {
			log.Fatalf("Failed to apply migration 027: %v", err)
		}
	} else {
		fmt.Println("✅ Migration 027 applied successfully")
	}
	fmt.Println("")

	// Run migration 028: workspace_audit_logs
	fmt.Println("📝 Running Migration 028: workspace_audit_logs")
	fmt.Println("-----------------------------------------------------------")

	migration028, err := os.ReadFile("internal/database/migrations/028_workspace_audit_logs.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 028: %v", err)
	}

	_, err = pool.Exec(ctx, string(migration028))
	if err != nil {
		// Check if it's already applied
		if isAlreadyExists(err) {
			fmt.Println("⚠️  Migration 028 already applied (table exists)")
		} else {
			log.Fatalf("Failed to apply migration 028: %v", err)
		}
	} else {
		fmt.Println("✅ Migration 028 applied successfully")
	}
	fmt.Println("")

	// Verify tables exist
	fmt.Println("🔍 Verifying tables...")
	fmt.Println("-----------------------------------------------------------")

	tables := []string{"workspace_invites", "workspace_audit_logs"}
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
		   contains(errMsg, "relation") && contains(errMsg, "already exists")
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
