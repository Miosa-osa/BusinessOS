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

	fmt.Println("=== FUNCTION SIGNATURE ===\n")

	// Get function return type details
	rows, err := pool.Query(ctx, `
		SELECT
			p.proname as function_name,
			pg_get_function_result(p.oid) as result_type,
			pg_get_function_arguments(p.oid) as arguments
		FROM pg_proc p
		WHERE p.proname = 'get_accessible_memories'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, result, args string
		rows.Scan(&name, &result, &args)
		fmt.Printf("Function: %s\n", name)
		fmt.Printf("Arguments: %s\n", args)
		fmt.Printf("Returns: %s\n\n", result)
	}

	// Try to get detailed column information
	fmt.Println("=== FUNCTION COLUMNS ===\n")

	rows2, err := pool.Query(ctx, `
		SELECT
			a.attname as column_name,
			format_type(a.atttypid, a.atttypmod) as data_type,
			a.attnum as position
		FROM pg_proc p
		JOIN pg_type t ON p.prorettype = t.oid
		JOIN pg_attribute a ON a.attrelid = t.typrelid
		WHERE p.proname = 'get_accessible_memories'
		AND a.attnum > 0
		ORDER BY a.attnum
	`)
	if err != nil {
		fmt.Printf("Note: Could not get column details (this is normal for functions): %v\n", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var colName, dataType string
			var pos int
			rows2.Scan(&colName, &dataType, &pos)
			fmt.Printf("%d. %s: %s\n", pos, colName, dataType)
		}
	}
}
