package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Quick database status checker
func main() {
	dbURL := "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("❌ Failed to connect:", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}

	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║          BusinessOS Database Status Check            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")

	// 1. Check migrations
	var migrationCount int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
	if err != nil {
		fmt.Println("\n❌ Migration table error:", err)
	} else {
		fmt.Printf("\n📋 Migrations: %d applied\n", migrationCount)

		// Get latest migration
		var latestVersion string
		err = pool.QueryRow(context.Background(),
			"SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&latestVersion)
		if err == nil {
			fmt.Printf("   Latest: %s\n", latestVersion)
		}
	}

	// 2. Check key tables with counts
	fmt.Println("\n🗄️  Key Tables:")
	tables := []string{
		"workspaces", "workspace_members", "workspace_roles",
		"projects", "project_members", "tasks",
		"image_embeddings", "workspace_memories",
	}

	for _, table := range tables {
		var count int64
		err := pool.QueryRow(context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("   ✗ %-25s ERROR\n", table)
		} else {
			fmt.Printf("   ✓ %-25s %d rows\n", table, count)
		}
	}

	// 3. Check extensions
	fmt.Println("\n🧩 Extensions:")
	extensions := []string{"vector", "uuid-ossp"}
	for _, ext := range extensions {
		var exists bool
		err := pool.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = $1)", ext).Scan(&exists)
		if err != nil || !exists {
			fmt.Printf("   ✗ %s\n", ext)
		} else {
			fmt.Printf("   ✓ %s\n", ext)
		}
	}

	// 4. Database stats
	fmt.Println("\n📊 Database Stats:")

	var dbSize string
	err = pool.QueryRow(context.Background(),
		"SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)
	if err == nil {
		fmt.Printf("   Size: %s\n", dbSize)
	}

	var tableCount int
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err == nil {
		fmt.Printf("   Tables: %d\n", tableCount)
	}

	var functionCount int
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM pg_proc WHERE pronamespace = 'public'::regnamespace").Scan(&functionCount)
	if err == nil {
		fmt.Printf("   Functions: %d\n", functionCount)
	}

	// 5. Recent activity
	fmt.Println("\n🕐 Recent Activity:")
	var auditCount int64
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM workspace_audit_logs WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&auditCount)
	if err == nil {
		fmt.Printf("   Audit logs (24h): %d\n", auditCount)
	}

	fmt.Println("\n" + strings.Repeat("═", 55))
	fmt.Println("✅ Database is operational")
	fmt.Println(strings.Repeat("═", 55))
}
