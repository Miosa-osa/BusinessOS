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

	fmt.Println("=== MEMORY DETAILS ===")
	fmt.Printf("Workspace: %s\n", workspaceID)
	fmt.Printf("User: %s\n\n", userID)

	rows, err := pool.Query(ctx, `
		SELECT id, title, visibility, owner_user_id, is_active, memory_type
		FROM workspace_memories
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Memories:")
	for rows.Next() {
		var id string
		var title, visibility, memType string
		var ownerUserID *string
		var isActive *bool

		rows.Scan(&id, &title, &visibility, &ownerUserID, &isActive, &memType)

		activeStr := "NULL"
		if isActive != nil {
			if *isActive {
				activeStr = "true"
			} else {
				activeStr = "false"
			}
		}

		ownerStr := "NULL"
		if ownerUserID != nil {
			ownerStr = *ownerUserID
			if *ownerUserID == userID {
				ownerStr += " ✓ MATCHES"
			} else {
				ownerStr += " ✗ MISMATCH"
			}
		}

		fmt.Printf("\n  Title: %s\n", title)
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  Visibility: %s\n", visibility)
		fmt.Printf("  Owner: %s\n", ownerStr)
		fmt.Printf("  Is Active: %s\n", activeStr)
		fmt.Printf("  Memory Type: %s\n", memType)

		// Check if this memory should be returned
		shouldReturn := false
		reason := ""

		if visibility == "workspace" {
			shouldReturn = true
			reason = "visibility='workspace'"
		} else if visibility == "private" && ownerUserID != nil && *ownerUserID == userID {
			shouldReturn = true
			reason = "visibility='private' AND owner matches"
		}

		if shouldReturn {
			fmt.Printf("  ✓ SHOULD BE RETURNED: %s\n", reason)
		} else {
			fmt.Printf("  ✗ SHOULD NOT BE RETURNED\n")
		}
	}
}
