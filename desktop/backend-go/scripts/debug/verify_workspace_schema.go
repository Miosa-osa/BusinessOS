package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// verify_workspace_schema.go - Verify workspace tables and schema
// Run with: go run verify_workspace_schema.go

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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

	fmt.Println("🔍 Verifying Workspace Schema (Migration 026)")
	fmt.Println("=" + string(make([]byte, 60)))

	// Check tables
	tables := []string{
		"workspaces",
		"workspace_roles",
		"workspace_members",
		"user_workspace_profiles",
		"workspace_memories",
		"role_permissions",
	}

	fmt.Println("\n📊 Table Verification:")
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
			table).Scan(&exists)

		if err != nil {
			log.Printf("  ❌ Error checking %s: %v", table, err)
			continue
		}

		if exists {
			// Get row count
			var count int64
			err = pool.QueryRow(context.Background(),
				fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err != nil {
				log.Printf("  ✅ %s exists (couldn't get count)", table)
			} else {
				fmt.Printf("  ✅ %s exists (%d rows)\n", table, count)
			}
		} else {
			fmt.Printf("  ❌ %s NOT FOUND\n", table)
		}
	}

	// Check function
	fmt.Println("\n🔧 Function Verification:")
	var functionExists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'seed_default_workspace_roles')").Scan(&functionExists)

	if err != nil {
		log.Printf("  ❌ Error checking function: %v", err)
	} else if functionExists {
		fmt.Println("  ✅ seed_default_workspace_roles() exists")
	} else {
		fmt.Println("  ❌ seed_default_workspace_roles() NOT FOUND")
	}

	// Check workspace_id column in projects table
	fmt.Println("\n🔗 Integration Verification:")
	var workspaceIdExists bool
	err = pool.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'projects' AND column_name = 'workspace_id'
		)`).Scan(&workspaceIdExists)

	if err != nil {
		log.Printf("  ❌ Error checking projects.workspace_id: %v", err)
	} else if workspaceIdExists {
		fmt.Println("  ✅ projects.workspace_id column exists")
	} else {
		fmt.Println("  ⚠️  projects.workspace_id column NOT FOUND (migration may not be complete)")
	}

	// Check vector extension
	fmt.Println("\n📦 Extension Verification:")
	var vectorExists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&vectorExists)

	if err != nil {
		log.Printf("  ❌ Error checking vector extension: %v", err)
	} else if vectorExists {
		fmt.Println("  ✅ pgvector extension installed")
	} else {
		fmt.Println("  ⚠️  pgvector extension NOT FOUND (needed for embeddings)")
	}

	// Sample workspace query
	fmt.Println("\n🧪 Sample Data Verification:")
	var workspaceCount int64
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM workspaces").Scan(&workspaceCount)

	if err != nil {
		log.Printf("  ⚠️  Could not query workspaces: %v", err)
	} else {
		fmt.Printf("  📊 Workspaces: %d\n", workspaceCount)

		if workspaceCount > 0 {
			// Show first workspace
			var id, name, slug, owner string
			var planType string
			err = pool.QueryRow(context.Background(),
				"SELECT id, name, slug, owner_id, plan_type FROM workspaces LIMIT 1").
				Scan(&id, &name, &slug, &owner, &planType)

			if err == nil {
				fmt.Printf("  📌 Sample workspace: %s (%s) - %s plan\n", name, slug, planType)

				// Check if roles were seeded
				var roleCount int64
				err = pool.QueryRow(context.Background(),
					"SELECT COUNT(*) FROM workspace_roles WHERE workspace_id = $1", id).Scan(&roleCount)

				if err == nil {
					fmt.Printf("  👥 Roles for this workspace: %d (expected: 6)\n", roleCount)
				}
			}
		}
	}

	// Index verification
	fmt.Println("\n🗂️  Index Verification:")
	indexes := []string{
		"idx_workspaces_slug",
		"idx_workspace_roles_workspace",
		"idx_workspace_members_workspace",
		"idx_workspace_memories_embedding",
	}

	for _, idx := range indexes {
		var exists bool
		err := pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)", idx).Scan(&exists)

		if err != nil {
			log.Printf("  ❌ Error checking %s: %v", idx, err)
		} else if exists {
			fmt.Printf("  ✅ %s exists\n", idx)
		} else {
			fmt.Printf("  ❌ %s NOT FOUND\n", idx)
		}
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("✅ Verification complete!")
}
