package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	userID := "ZVtQRaictVbO9lN0p-csSA"

	fmt.Println("=== CHECKING WORKSPACE MEMBERSHIP ===")
	fmt.Printf("Workspace: %s\n", workspaceID)
	fmt.Printf("User: %s\n\n", userID)

	// Check the exact query the function uses
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM workspace_members
			WHERE workspace_id = $1
			AND user_id = $2
			AND status = 'active'
		)
	`, workspaceID, userID).Scan(&exists)

	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("✓ User IS an active member of this workspace")

		// Get details
		var role, status string
		err = pool.QueryRow(ctx, `
			SELECT role, status
			FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		`, workspaceID, userID).Scan(&role, &status)

		if err == nil {
			fmt.Printf("  Role: %s\n", role)
			fmt.Printf("  Status: %s\n", status)
		}
	} else {
		fmt.Println("✗ User IS NOT an active member of this workspace")
		fmt.Println("\nThis is why get_accessible_memories returns 0 results!")

		// Check if user exists in table with different status
		rows, _ := pool.Query(ctx, `
			SELECT user_id, role, status
			FROM workspace_members
			WHERE workspace_id = $1
		`, workspaceID)
		defer rows.Close()

		fmt.Println("\nAll members in workspace:")
		for rows.Next() {
			var uid, role, status string
			rows.Scan(&uid, &role, &status)
			match := ""
			if uid == userID {
				match = " <-- THIS IS OUR USER"
			}
			fmt.Printf("  - %s (role: %s, status: %s)%s\n", uid, role, status, match)
		}
	}
}
