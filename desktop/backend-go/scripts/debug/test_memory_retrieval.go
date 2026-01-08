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

	fmt.Println("=== TESTING MEMORY RETRIEVAL ===")
	fmt.Printf("Workspace: %s\nUser: %s\n\n", workspaceID, userID)

	// Test 1: Get all workspace memories
	fmt.Println("1. All workspace memories:")
	rows, err := pool.Query(ctx, `
		SELECT id, title, content, visibility, importance_score, is_pinned, owner_user_id
		FROM workspace_memories
		WHERE workspace_id = $1
		ORDER BY importance_score DESC
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id string
		var title, content, visibility string
		var score float64
		var pinned bool
		var ownerID *string
		rows.Scan(&id, &title, &content, &visibility, &score, &pinned, &ownerID)

		ownerStr := "NULL"
		if ownerID != nil {
			ownerStr = *ownerID
		}

		fmt.Printf("  - [%s] %s (score: %.2f, pinned: %v, owner: %s)\n", visibility, title, score, pinned, ownerStr)
		count++
	}

	if count == 0 {
		fmt.Println("  NO MEMORIES FOUND!")
	} else {
		fmt.Printf("\nTotal: %d memories\n", count)
	}

	// Test 2: Simulate GetAccessibleMemories query
	fmt.Println("\n2. Accessible memories for user (simulating GetAccessibleMemories):")
	rows2, err := pool.Query(ctx, `
		SELECT id, title, content, visibility, importance_score, is_pinned
		FROM workspace_memories
		WHERE workspace_id = $1
		  AND (
		    visibility = 'workspace'
		    OR (visibility = 'private' AND owner_user_id = $2)
		    OR (visibility = 'shared' AND owner_user_id = $2)
		  )
		ORDER BY is_pinned DESC, importance_score DESC
		LIMIT 20
	`, workspaceID, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	count2 := 0
	for rows2.Next() {
		var id, title, content, visibility string
		var score float64
		var pinned bool
		rows2.Scan(&id, &title, &content, &visibility, &score, &pinned)
		fmt.Printf("  - [%s] %s (score: %.2f, pinned: %v)\n", visibility, title, score, pinned)
		if content != "" {
			fmt.Printf("    Content: %s\n", content)
		}
		count2++
	}

	if count2 == 0 {
		fmt.Println("  NO ACCESSIBLE MEMORIES!")
		fmt.Println("\nThis explains why memory injection didn't happen.")
	} else {
		fmt.Printf("\nTotal accessible: %d memories\n", count2)
		fmt.Println("\nThese memories SHOULD be injected into chat.")
	}
}
