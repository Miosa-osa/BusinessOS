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
	limit := 20

	fmt.Println("=== TESTING get_accessible_memories FUNCTION ===")
	fmt.Printf("Parameters:\n")
	fmt.Printf("  workspace_id: %s\n", workspaceID)
	fmt.Printf("  user_id: %s\n", userID)
	fmt.Printf("  memory_type: %v\n", memoryType)
	fmt.Printf("  limit: %d\n\n", limit)

	rows, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, $3, $4)",
		workspaceID, userID, memoryType, limit)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Results:")
	count := 0
	for rows.Next() {
		values, _ := rows.Values()
		fmt.Printf("  Row %d: %v\n", count+1, values)
		count++
	}

	if count == 0 {
		fmt.Println("  NO RESULTS RETURNED")
		fmt.Println("\nThis is why memory injection is not working!")
		fmt.Println("\nLet me check the function definition...")

		// Get function definition
		var funcDef string
		err = pool.QueryRow(ctx, `
			SELECT pg_get_functiondef(oid)
			FROM pg_proc
			WHERE proname = 'get_accessible_memories'
		`).Scan(&funcDef)

		if err != nil {
			log.Printf("Failed to get function definition: %v", err)
		} else {
			fmt.Println("\nFunction definition:")
			fmt.Println(funcDef)
		}
	} else {
		fmt.Printf("\nTotal: %d results\n", count)
	}
}
