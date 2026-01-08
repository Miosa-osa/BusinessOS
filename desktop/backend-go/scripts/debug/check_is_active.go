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

	fmt.Println("=== CHECKING is_active FIELD ON MEMORIES ===")

	rows, err := pool.Query(ctx, `
		SELECT id, title, is_active, visibility
		FROM workspace_memories
		WHERE workspace_id = $1
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Memories:")
	count := 0
	for rows.Next() {
		var id string
		var title string
		var isActive *bool
		var visibility string

		rows.Scan(&id, &title, &isActive, &visibility)

		activeStr := "NULL"
		if isActive != nil {
			if *isActive {
				activeStr = "true"
			} else {
				activeStr = "false"
			}
		}

		fmt.Printf("  - %s (is_active=%s, visibility=%s)\n", title, activeStr, visibility)
		count++
	}

	if count == 0 {
		fmt.Println("  NO MEMORIES FOUND")
	} else {
		fmt.Printf("\nTotal: %d memories\n", count)

		// If any have NULL is_active, that's the problem
		var nullCount int
		pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM workspace_memories
			WHERE workspace_id = $1 AND is_active IS NULL
		`, workspaceID).Scan(&nullCount)

		if nullCount > 0 {
			fmt.Printf("\n⚠️  %d memories have is_active=NULL\n", nullCount)
			fmt.Println("The function requires is_active=true, so these are excluded!")
			fmt.Println("\nFixing by setting is_active=true...")

			_, err = pool.Exec(ctx, `
				UPDATE workspace_memories
				SET is_active = true
				WHERE workspace_id = $1 AND is_active IS NULL
			`, workspaceID)

			if err != nil {
				fmt.Printf("Failed to update: %v\n", err)
			} else {
				fmt.Println("✓ Updated all memories to is_active=true")
			}
		}
	}
}
