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

	workspaceID := "064e8e2a-5d3e-4d00-8492-df3628b1ec96"

	// Check workspace_roles for this workspace
	fmt.Println("=== WORKSPACE_ROLES ===")
	rows, err := pool.Query(ctx, "SELECT id, workspace_id, name, display_name, hierarchy_level FROM workspace_roles WHERE workspace_id = $1", workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, workspaceIDCol, name, displayName string
		var hierarchyLevel int
		rows.Scan(&id, &workspaceIDCol, &name, &displayName, &hierarchyLevel)
		fmt.Printf("%s | %s | %s | %d\n", name, displayName, id, hierarchyLevel)
		count++
	}
	if count == 0 {
		fmt.Println("NO ROLES FOUND FOR THIS WORKSPACE!")
		fmt.Println("\nThis is the problem! The workspace needs default roles created.")
	} else {
		fmt.Printf("\nTotal: %d roles\n", count)
	}

	// Check what the workspace members expect
	fmt.Println("\n=== WORKSPACE_MEMBERS ROLES ===")
	rows2, err := pool.Query(ctx, "SELECT user_id, role FROM workspace_members WHERE workspace_id = $1", workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var userID, role string
		rows2.Scan(&userID, &role)
		fmt.Printf("User %s expects role: %s\n", userID, role)
	}
}
