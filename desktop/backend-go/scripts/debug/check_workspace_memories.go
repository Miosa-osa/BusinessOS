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

	rows, err := pool.Query(context.Background(), `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'workspace_memories'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("workspace_memories table columns:")
	for rows.Next() {
		var col, typ, nullable string
		rows.Scan(&col, &typ, &nullable)
		fmt.Printf("  %s: %s (nullable: %s)\n", col, typ, nullable)
	}
}
