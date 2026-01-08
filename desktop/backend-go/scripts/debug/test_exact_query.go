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
	var memoryType *string = nil

	fmt.Println("=== TESTING EXACT QUERY FROM FUNCTION ===\n")

	// This is the exact query from the function
	rows, err := pool.Query(ctx, `
		SELECT
			wm.id,
			wm.title,
			wm.content,
			wm.memory_type,
			wm.visibility,
			wm.importance_score as importance,
			wm.tags,
			wm.metadata,
			(wm.owner_user_id = $2 OR wm.owner_user_id IS NULL) as is_owner,
			wm.access_count,
			wm.created_at
		FROM workspace_memories wm
		WHERE wm.workspace_id = $1
		AND wm.is_active = true
		AND (
			wm.visibility = 'workspace' OR wm.visibility IS NULL
			OR
			(wm.visibility = 'private' AND wm.owner_user_id = $2)
			OR
			(wm.visibility = 'shared' AND (wm.owner_user_id = $2 OR $2 = ANY(COALESCE(wm.shared_with, ARRAY[]::TEXT[]))))
		)
		AND ($3 IS NULL OR wm.memory_type = $3)
		ORDER BY wm.importance_score DESC NULLS LAST, wm.created_at DESC
		LIMIT 20
	`, workspaceID, userID, memoryType)

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Results:")
	count := 0
	for rows.Next() {
		var id string
		var title, content, memType, visibility string
		var importance float64
		var tags []string
		var metadata map[string]interface{}
		var isOwner bool
		var accessCount int
		var createdAt string

		err := rows.Scan(&id, &title, &content, &memType, &visibility, &importance, &tags, &metadata, &isOwner, &accessCount, &createdAt)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		fmt.Printf("  %d. %s (visibility=%s, importance=%.2f)\n", count+1, title, visibility, importance)
		count++
	}

	if count == 0 {
		fmt.Println("  NO RESULTS!")
		fmt.Println("\nDebugging: Let's check each condition separately...\n")

		// Check basic filter
		var basicCount int
		pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM workspace_memories
			WHERE workspace_id = $1 AND is_active = true
		`, workspaceID).Scan(&basicCount)
		fmt.Printf("1. Workspace + is_active filter: %d memories\n", basicCount)

		// Check visibility filter
		var visCount int
		pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM workspace_memories
			WHERE workspace_id = $1
			AND is_active = true
			AND (
				visibility = 'workspace' OR visibility IS NULL
				OR
				(visibility = 'private' AND owner_user_id = $2)
				OR
				(visibility = 'shared' AND (owner_user_id = $2 OR $2 = ANY(COALESCE(shared_with, ARRAY[]::TEXT[]))))
			)
		`, workspaceID, userID).Scan(&visCount)
		fmt.Printf("2. + Visibility filter: %d memories\n", visCount)

		// Check if tags/metadata columns exist
		var colCount int
		pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_name = 'workspace_memories'
			AND column_name IN ('tags', 'metadata', 'access_count')
		`).Scan(&colCount)
		fmt.Printf("\n3. Columns tags/metadata/access_count exist: %d/3\n", colCount)

		if colCount < 3 {
			fmt.Println("\n⚠️  PROBLEM: Missing columns!")
			fmt.Println("The function expects tags, metadata, and access_count columns")
			fmt.Println("but they don't exist in the table!")
		}
	} else {
		fmt.Printf("\nTotal: %d results\n", count)
	}
}
