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
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("🔍 Checking project_members table schema...\n")

	// Get columns
	rows, err := pool.Query(context.Background(), `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'project_members'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns:")
	for rows.Next() {
		var colName, dataType, isNullable string
		var colDefault *string
		rows.Scan(&colName, &dataType, &isNullable, &colDefault)

		defaultVal := "NULL"
		if colDefault != nil {
			defaultVal = *colDefault
		}
		fmt.Printf("  - %s: %s (nullable: %s, default: %s)\n", colName, dataType, isNullable, defaultVal)
	}
}
