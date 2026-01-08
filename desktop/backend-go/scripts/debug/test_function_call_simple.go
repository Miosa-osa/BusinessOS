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

	fmt.Println("=== CALLING POSTGRESQL FUNCTION ===")
	fmt.Printf("Workspace: %s\n", workspaceID)
	fmt.Printf("User: %s\n", userID)
	fmt.Println()

	// Call the function with explicit schema
	rows, err := pool.Query(ctx, `
		SELECT id, title, content, memory_type, visibility, importance, tags, metadata, is_owner, access_count, created_at
		FROM public.get_accessible_memories($1::uuid, $2::text, $3::text, $4::integer)
	`, workspaceID, userID, nil, 20)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
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
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		count++
		fmt.Printf("%d. %s (visibility=%s, type=%s)\n", count, title, visibility, memType)
	}

	if count == 0 {
		fmt.Println("NO RESULTS!")
		fmt.Println("\nChecking for row errors...")
		if rows.Err() != nil {
			fmt.Printf("Row error: %v\n", rows.Err())
		} else {
			fmt.Println("No row errors - function genuinely returned 0 rows")
		}
	} else {
		fmt.Printf("\nTotal: %d memories\n", count)
	}
}
