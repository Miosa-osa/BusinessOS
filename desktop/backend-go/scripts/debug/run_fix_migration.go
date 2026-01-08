package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Try to get DATABASE_URL from environment, or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try common default locations
		dbURL = "postgresql://postgres:postgres@localhost:5432/businessos?sslmode=disable"
		fmt.Println("⚠️  DATABASE_URL not set, using default:", dbURL)
		fmt.Println("    If this fails, set DATABASE_URL environment variable")
		fmt.Println()
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v\n", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Test connection
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v\n", err)
	}
	fmt.Println("✅ Connected to database")

	// Run the fix SQL
	fmt.Println("🔧 Running workspace members fix...")

	sql := `
		INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at)
		SELECT w.id, w.owner_id, 'owner', 'active', NOW()
		FROM workspaces w
		WHERE NOT EXISTS (
			SELECT 1 FROM workspace_members wm
			WHERE wm.workspace_id = w.id AND wm.user_id = w.owner_id
		)
		ON CONFLICT (workspace_id, user_id) DO UPDATE
		SET role = 'owner', status = 'active';
	`

	result, err := pool.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("❌ Fix failed: %v\n", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected > 0 {
		fmt.Printf("✅ Fixed %d workspace(s) by adding owners as members\n", rowsAffected)
	} else {
		fmt.Println("✅ All workspaces already have their owners as members")
	}

	// Verify the specific workspace
	var memberCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM workspace_members
		WHERE workspace_id = '064e8e2a-5d3e-4d00-8492-df3628b1ec96'
	`).Scan(&memberCount)

	if err != nil {
		fmt.Printf("⚠️  Could not verify workspace: %v\n", err)
	} else {
		fmt.Printf("\n📊 Workspace '064e8e2a-5d3e-4d00-8492-df3628b1ec96' now has %d member(s)\n", memberCount)
	}

	fmt.Println("\n🎉 Fix complete! Refresh your browser to test.")
}
