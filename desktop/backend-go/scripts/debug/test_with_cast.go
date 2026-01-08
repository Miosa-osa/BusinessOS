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
	var memoryType interface{} = nil
	limit := 20

	fmt.Println("=== TESTING WITH TYPE CAST ===\n")

	rows, err := pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, $3::text, $4)",
		workspaceID, userID, memoryType, limit)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Results:")
	count := 0
	for rows.Next() {
		values, _ := rows.Values()
		if count == 0 {
			fmt.Printf("  Columns: %v\n", len(values))
		}
		fmt.Printf("  Row %d: %v\n", count+1, values[1]) // Print title (second column)
		count++
	}

	if count == 0 {
		fmt.Println("  NO RESULTS!")
		fmt.Println("\nStill not working. Let me check the table columns...")

		// Check table schema
		rows2, _ := pool.Query(ctx, `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = 'workspace_memories'
			ORDER BY ordinal_position
		`)
		defer rows2.Close()

		fmt.Println("\nworkspace_memories table schema:")
		for rows2.Next() {
			var colName, dataType, nullable string
			rows2.Scan(&colName, &dataType, &nullable)
			fmt.Printf("  %s: %s (nullable: %s)\n", colName, dataType, nullable)
		}
	} else {
		fmt.Printf("\n✓ SUCCESS! Retrieved %d memories\n", count)
	}
}
