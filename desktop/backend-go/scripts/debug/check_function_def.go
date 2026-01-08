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

	fmt.Println("=== FUNCTION DEFINITION ===\n")

	rows, err := pool.Query(ctx, `
		SELECT pg_get_functiondef(oid)
		FROM pg_proc
		WHERE proname = 'get_accessible_memories'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var def string
		rows.Scan(&def)
		fmt.Println(def)
	}
}
