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

	// Check workspaces
	fmt.Println("=== WORKSPACES ===")
	rows, err := pool.Query(ctx, "SELECT id, name, slug, created_at FROM workspaces ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, name, slug string
		var createdAt string
		rows.Scan(&id, &name, &slug, &createdAt)
		fmt.Printf("%s | %s | %s | %s\n", id, name, slug, createdAt)
		count++
	}
	if count == 0 {
		fmt.Println("NO WORKSPACES FOUND")
	} else {
		fmt.Printf("\nTotal: %d workspaces\n", count)
	}

	// Check workspace members schema first
	fmt.Println("\n=== WORKSPACE_MEMBERS TABLE SCHEMA ===")
	schemaRows, _ := pool.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'workspace_members' ORDER BY ordinal_position")
	for schemaRows.Next() {
		var colName, dataType string
		schemaRows.Scan(&colName, &dataType)
		fmt.Printf("  %s: %s\n", colName, dataType)
	}
	schemaRows.Close()

	// Check workspace members
	fmt.Println("\n=== WORKSPACE MEMBERS ===")
	rows2, err := pool.Query(ctx, "SELECT * FROM workspace_members LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	count2 := 0
	for rows2.Next() {
		values, _ := rows2.Values()
		fmt.Printf("%v\n", values)
		count2++
	}
	if count2 == 0 {
		fmt.Println("NO WORKSPACE MEMBERS FOUND - USER IS NOT A MEMBER OF ANY WORKSPACE")
	} else {
		fmt.Printf("\nTotal: %d members\n", count2)
	}

	// Check current user
	fmt.Println("\n=== CURRENT USER ===")
	var currentUserID string
	err = pool.QueryRow(ctx, "SELECT id FROM users LIMIT 1").Scan(&currentUserID)
	if err != nil {
		fmt.Println("No users found")
	} else {
		fmt.Printf("First user ID: %s\n", currentUserID)

		// Check if this user is a member of any workspace
		var memberCount int
		pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspace_members WHERE user_id = $1", currentUserID).Scan(&memberCount)
		fmt.Printf("User is member of %d workspaces\n", memberCount)
	}
}
