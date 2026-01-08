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

	fmt.Println("=== TESTING DIFFERENT FUNCTION CALL METHODS ===\n")

	// Method 1: With all 4 parameters, NULL as go nil
	fmt.Println("Method 1: SELECT * FROM get_accessible_memories($1, $2, $3, $4) with nil")
	rows, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, $3, $4)", workspaceID, userID, nil, 20)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()
		fmt.Printf("  Result: %d rows\n\n", count)
	}

	// Method 2: With only 2 parameters (using defaults)
	fmt.Println("Method 2: SELECT * FROM get_accessible_memories($1, $2) using defaults")
	rows, err = pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2)", workspaceID, userID)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()
		fmt.Printf("  Result: %d rows\n\n", count)
	}

	// Method 3: With explicit NULL::text cast in SQL
	fmt.Println("Method 3: SELECT * FROM get_accessible_memories($1::uuid, $2::text, NULL::text, $3::integer)")
	rows, err = pool.Query(ctx, "SELECT * FROM get_accessible_memories($1::uuid, $2::text, NULL::text, $3::integer)", workspaceID, userID, 20)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()
		fmt.Printf("  Result: %d rows\n\n", count)
	}

	// Method 4: With explicit column selection
	fmt.Println("Method 4: SELECT id, title, ... FROM get_accessible_memories(...)")
	rows, err = pool.Query(ctx, `
		SELECT id, title, content, memory_type, visibility, importance, tags, metadata, is_owner, access_count, created_at
		FROM get_accessible_memories($1, $2)
	`, workspaceID, userID)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()
		fmt.Printf("  Result: %d rows\n\n", count)
	}
}
