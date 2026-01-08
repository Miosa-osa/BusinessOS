package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("Running migration 030 v2...")
	migration, err := os.ReadFile("internal/database/migrations/030_memory_hierarchy_v2.sql")
	if err != nil {
		log.Fatal(err)
	}

	_, err = pool.Exec(context.Background(), string(migration))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("✅ Migration 030 v2 applied successfully!")
}
