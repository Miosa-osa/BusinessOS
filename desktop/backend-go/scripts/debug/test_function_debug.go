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

	fmt.Println("=== DEBUGGING FUNCTION ===\n")

	// Test 1: Membership check (exactly as function does it)
	var memberExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM workspace_members
			WHERE workspace_id = $1
			AND user_id = $2
			AND status = 'active'
		)
	`, workspaceID, userID).Scan(&memberExists)
	fmt.Printf("1. Membership check: %v\n", memberExists)
	if !memberExists {
		fmt.Println("   ❌ User is NOT a workspace member!")
		fmt.Println("   The function returns early here, that's why we get 0 results!\n")

		// Check what's actually in the table
		rows, _ := pool.Query(ctx, `
			SELECT user_id, status, role
			FROM workspace_members
			WHERE workspace_id = $1
		`, workspaceID)
		defer rows.Close()

		fmt.Println("All members in workspace:")
		for rows.Next() {
			var uid, status, role string
			rows.Scan(&uid, &status, &role)
			match := ""
			if uid == userID {
				match = " ✓ THIS IS OUR USER"
			}
			fmt.Printf("  - user_id=%s, status=%s, role=%s%s\n", uid, status, role, match)
		}
		return
	}
	fmt.Println("   ✓ Membership check passed\n")

	// Test 2: Count matching rows (without function)
	var matchCount int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
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
	`, workspaceID, userID).Scan(&matchCount)
	fmt.Printf("2. Matching rows (raw query): %d\n", matchCount)
	if matchCount == 0 {
		fmt.Println("   ❌ No matching rows!")
	} else {
		fmt.Printf("   ✓ Found %d matching rows\n\n", matchCount)
	}

	// Test 3: Call the function
	rows, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, NULL, 20)", workspaceID, userID)
	if err != nil {
		fmt.Printf("3. Function call error: %v\n", err)
		return
	}
	defer rows.Close()

	funcCount := 0
	for rows.Next() {
		funcCount++
	}
	fmt.Printf("3. Function returned: %d rows\n", funcCount)
	if funcCount == 0 {
		fmt.Println("   ❌ Function returns 0 rows even though raw query works!")
	} else {
		fmt.Printf("   ✓ Function works!\n")
	}
}
