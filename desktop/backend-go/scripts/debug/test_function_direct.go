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

	fmt.Println("=== CALLING FUNCTION DIRECTLY ===")
	fmt.Printf("Workspace: %s\n", workspaceID)
	fmt.Printf("User: %s\n", userID)
	fmt.Printf("Memory Type: NULL\n")
	fmt.Printf("Limit: 20\n\n")

	// Try 1: With NULL as interface{}
	fmt.Println("Attempt 1: Using interface{}(nil) with ::text cast")
	var memTypeNil interface{} = nil
	rows, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, $3::text, $4)",
		workspaceID, userID, memTypeNil, 20)
	if err != nil {
		fmt.Printf("  Error: %v\n\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			values, _ := rows.Values()
			if count == 0 && len(values) > 1 {
				fmt.Printf("  Got result: %v\n", values[1])
			}
			count++
		}
		if count == 0 {
			fmt.Println("  NO RESULTS\n")
		} else {
			fmt.Printf("  Total: %d results\n\n", count)
		}
	}

	// Try 2: Using DEFAULT for optional parameter
	fmt.Println("Attempt 2: Using DEFAULT for optional parameter")
	rows2, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2)",
		workspaceID, userID)
	if err != nil {
		fmt.Printf("  Error: %v\n\n", err)
	} else {
		defer rows2.Close()
		count := 0
		for rows2.Next() {
			values, _ := rows2.Values()
			if count == 0 && len(values) > 1 {
				fmt.Printf("  Got result: %v\n", values[1])
			}
			count++
		}
		if count == 0 {
			fmt.Println("  NO RESULTS\n")
		} else {
			fmt.Printf("  Total: %d results\n\n", count)
		}
	}

	// Try 3: Using explicit NULL
	fmt.Println("Attempt 3: Hardcoded NULL in SQL")
	rows3, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1::uuid, $2::text, NULL::text, 20::integer)",
		workspaceID, userID)
	if err != nil {
		fmt.Printf("  Error: %v\n\n", err)
	} else {
		defer rows3.Close()
		count := 0
		for rows3.Next() {
			values, _ := rows3.Values()
			if count == 0 && len(values) > 1 {
				fmt.Printf("  Got result: %v\n", values[1])
			}
			count++
		}
		if count == 0 {
			fmt.Println("  NO RESULTS\n")
		} else {
			fmt.Printf("  Total: %d results\n\n", count)
		}
	}
}
