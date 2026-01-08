package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// Check if the function exists
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_proc p
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE n.nspname = 'public'
			  AND p.proname = 'get_accessible_memories'
		)
	`).Scan(&exists)

	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("✓ Function get_accessible_memories EXISTS")
	} else {
		fmt.Println("✗ Function get_accessible_memories DOES NOT EXIST")
		fmt.Println("\nThe function needs to be created by running migration 030_memory_hierarchy.sql")
	}
}
