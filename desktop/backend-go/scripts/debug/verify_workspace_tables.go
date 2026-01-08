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

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("✅ Connected to database")
	fmt.Println("\n🔍 Checking workspace-related tables...")

	tables := []string{
		"workspaces",
		"workspace_roles",
		"workspace_members",
		"user_workspace_profiles",
		"workspace_memories",
		"project_members",
		"role_permissions",
	}

	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			table).Scan(&exists)

		if err != nil {
			fmt.Printf("❌ Error checking %s: %v\n", table, err)
			continue
		}

		if exists {
			// Get row count
			var count int
			err = pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err != nil {
				fmt.Printf("✅ Table '%s' exists (couldn't get count: %v)\n", table, err)
			} else {
				fmt.Printf("✅ Table '%s' exists (%d rows)\n", table, count)
			}
		} else {
			fmt.Printf("❌ Table '%s' does NOT exist\n", table)
		}
	}

	// Check if projects table has workspace_id column
	fmt.Println("\n🔍 Checking projects table for workspace_id column...")
	var hasColumn bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'projects' AND column_name = 'workspace_id')").Scan(&hasColumn)

	if err != nil {
		fmt.Printf("❌ Error checking projects.workspace_id: %v\n", err)
	} else if hasColumn {
		fmt.Println("✅ Column 'workspace_id' exists in projects table")
	} else {
		fmt.Println("❌ Column 'workspace_id' does NOT exist in projects table")
	}

	// Check seed function exists
	fmt.Println("\n🔍 Checking seed function...")
	var funcExists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'seed_default_workspace_roles')").Scan(&funcExists)

	if err != nil {
		fmt.Printf("❌ Error checking function: %v\n", err)
	} else if funcExists {
		fmt.Println("✅ Function 'seed_default_workspace_roles' exists")
	} else {
		fmt.Println("❌ Function 'seed_default_workspace_roles' does NOT exist")
	}

	fmt.Println("\n✅ Verification complete!")
}
