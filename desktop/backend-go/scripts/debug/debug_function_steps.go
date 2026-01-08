package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	fmt.Println("=== DEBUGGING FUNCTION LOGIC STEP BY STEP ===\n")

	// Step 1: Check membership (exactly as function does)
	fmt.Println("Step 1: Membership check")
	var memberExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM workspace_members
			WHERE workspace_id = $1
			AND user_id = $2
			AND status = 'active'
		)
	`, workspaceID, userID).Scan(&memberExists)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Member exists: %v\n", memberExists)
	if !memberExists {
		fmt.Println("  ❌ MEMBERSHIP CHECK WOULD FAIL - Function would return empty!")
		return
	}
	fmt.Println("  ✓ Membership check passed\n")

	// Step 2: Execute the main query (exactly as function does)
	fmt.Println("Step 2: Main query")
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
		AND ($3::text IS NULL OR wm.memory_type = $3::text)
		ORDER BY wm.importance_score DESC NULLS LAST, wm.created_at DESC
		LIMIT $4
	`, workspaceID, userID, nil, 20)

	if err != nil {
		fmt.Printf("  ❌ Query error: %v\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id uuid.UUID
		var title, content, memType, visibility string
		var importance *float64
		var tags []string
		var metadata map[string]interface{}
		var isOwner bool
		var accessCount *int
		var createdAt time.Time

		err := rows.Scan(&id, &title, &content, &memType, &visibility, &importance, &tags, &metadata, &isOwner, &accessCount, &createdAt)
		if err != nil {
			fmt.Printf("  Scan error: %v\n", err)
			continue
		}

		count++
		fmt.Printf("  %d. %s (visibility=%s, type=%s)\n", count, title, visibility, memType)
	}

	if count == 0 {
		fmt.Println("  ❌ NO RESULTS from main query!")
	} else {
		fmt.Printf("\n  ✓ Total: %d memories\n", count)
	}
}
