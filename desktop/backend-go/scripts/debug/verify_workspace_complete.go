package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("🔍 Checking Workspace Implementation...")
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

	fmt.Println("\n📊 TABLES:")
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = $1
			)
		`, table).Scan(&exists)

		if err != nil {
			fmt.Printf("  ❌ Error checking %s: %v\n", table, err)
		} else if exists {
			// Count rows
			var count int
			pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("  ✅ %-30s (rows: %d)\n", table, count)
		} else {
			fmt.Printf("  ❌ %-30s NOT FOUND\n", table)
		}
	}

	// Check function
	fmt.Println("\n🔧 FUNCTIONS:")
	var funcExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM pg_proc
			WHERE proname = 'seed_default_workspace_roles'
		)
	`).Scan(&funcExists)

	if err != nil {
		fmt.Printf("  ❌ Error checking function: %v\n", err)
	} else if funcExists {
		fmt.Println("  ✅ seed_default_workspace_roles()")
	} else {
		fmt.Println("  ❌ seed_default_workspace_roles() NOT FOUND")
	}

	// Check if any workspaces exist
	var wsCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspaces").Scan(&wsCount)

	fmt.Println("\n📈 CURRENT DATA:")
	fmt.Printf("  Workspaces: %d\n", wsCount)

	if wsCount > 0 {
		var roleCount int
		pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspace_roles").Scan(&roleCount)
		fmt.Printf("  Total roles: %d\n", roleCount)

		// List workspaces
		rows, _ := pool.Query(ctx, "SELECT id, name, slug, owner_id FROM workspaces LIMIT 5")
		defer rows.Close()

		fmt.Println("\n  Recent workspaces:")
		for rows.Next() {
			var id, name, slug, ownerID string
			rows.Scan(&id, &name, &slug, &ownerID)
			fmt.Printf("    - %s (slug: %s, owner: %s)\n", name, slug, ownerID)
		}
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("✅ Verification complete!")
}
