package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection
	dbURL := "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Test connection
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Database connection successful")

	// Create migration tracking table if it doesn't exist
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create schema_migrations table:", err)
	}

	// Define all migrations in order
	migrations := []struct {
		version string
		name    string
		path    string
	}{
		{"025", "Image Embeddings", "internal/database/migrations/025_image_embeddings.sql"},
		{"026", "Workspaces and Roles", "internal/database/migrations/026_workspaces_and_roles.sql"},
		{"027", "Workspace Invites", "internal/database/migrations/027_workspace_invites.sql"},
		{"028", "Workspace Audit Logs", "internal/database/migrations/028_workspace_audit_logs.sql"},
		{"029", "Project Members", "internal/database/migrations/029_project_members.sql"},
		{"030", "Memory Hierarchy v2", "internal/database/migrations/030_memory_hierarchy_v2.sql"},
	}

	// Get already applied migrations
	rows, err := pool.Query(context.Background(), "SELECT version FROM schema_migrations")
	if err != nil {
		log.Fatal("Failed to query schema_migrations:", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			log.Printf("Error scanning version: %v", err)
			continue
		}
		appliedMigrations[version] = true
	}

	fmt.Printf("\n📊 Migration Status:\n")
	fmt.Printf("   Already applied: %d migrations\n", len(appliedMigrations))

	// Sort applied migrations to show them
	var applied []string
	for version := range appliedMigrations {
		applied = append(applied, version)
	}
	sort.Strings(applied)
	if len(applied) > 0 {
		fmt.Printf("   ✓ Applied versions: %s\n", strings.Join(applied, ", "))
	}

	// Apply pending migrations
	pendingCount := 0
	for _, mig := range migrations {
		if appliedMigrations[mig.version] {
			fmt.Printf("\n⏭️  Skipping %s (%s) - already applied\n", mig.version, mig.name)
			continue
		}

		pendingCount++
		fmt.Printf("\n📦 Applying migration %s: %s\n", mig.version, mig.name)
		fmt.Printf("   Path: %s\n", mig.path)

		// Read migration file
		sqlContent, err := os.ReadFile(mig.path)
		if err != nil {
			log.Printf("❌ Failed to read migration file: %v\n", err)
			continue
		}

		// Begin transaction
		tx, err := pool.Begin(context.Background())
		if err != nil {
			log.Printf("❌ Failed to begin transaction: %v\n", err)
			continue
		}

		// Execute migration
		_, err = tx.Exec(context.Background(), string(sqlContent))
		if err != nil {
			tx.Rollback(context.Background())
			log.Printf("❌ Failed to execute migration: %v\n", err)
			log.Printf("   Error details: %v\n", err)
			continue
		}

		// Record migration as applied
		_, err = tx.Exec(context.Background(),
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			mig.version)
		if err != nil {
			tx.Rollback(context.Background())
			log.Printf("❌ Failed to record migration: %v\n", err)
			continue
		}

		// Commit transaction
		err = tx.Commit(context.Background())
		if err != nil {
			log.Printf("❌ Failed to commit transaction: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Migration %s applied successfully\n", mig.version)
	}

	// Summary
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	if pendingCount == 0 {
		fmt.Printf("✅ All migrations are up to date! No pending migrations.\n")
	} else {
		fmt.Printf("🎉 Migration run complete!\n")
		fmt.Printf("   Pending migrations processed: %d\n", pendingCount)
	}

	// Verify key tables exist
	fmt.Printf("\n🔍 Verifying key tables...\n")
	tables := []string{
		"image_embeddings",
		"image_tags",
		"image_collections",
		"workspaces",
		"workspace_roles",
		"workspace_members",
		"user_workspace_profiles",
		"workspace_memories",
		"workspace_invites",
		"workspace_audit_logs",
		"project_members",
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

	// Show final migration status
	rows2, err := pool.Query(context.Background(), "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Printf("Failed to query final status: %v", err)
	} else {
		defer rows2.Close()
		var allVersions []string
		for rows2.Next() {
			var version string
			if err := rows2.Scan(&version); err == nil {
				allVersions = append(allVersions, version)
			}
		}
		fmt.Printf("\n📋 All applied migrations: %s\n", strings.Join(allVersions, ", "))
	}

	fmt.Println("\n✅ Migration verification complete!")
}
