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

	fmt.Println("=== workspace_members TABLE SCHEMA ===\n")

	rows, err := pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'workspace_members'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var colName, dataType, nullable string
		rows.Scan(&colName, &dataType, &nullable)
		fmt.Printf("  %s: %s (nullable: %s)\n", colName, dataType, nullable)
	}

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	userID := "ZVtQRaictVbO9lN0p-csSA"

	fmt.Println("\n=== ACTUAL DATA ===\n")

	rows2, err := pool.Query(ctx, `
		SELECT workspace_id, user_id, role, status, pg_typeof(user_id) as user_id_type
		FROM workspace_members
		WHERE workspace_id = $1
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var wsID, uid, role, status, uidType string
		rows2.Scan(&wsID, &uid, &role, &status, &uidType)
		match := ""
		if uid == userID {
			match = " ✓ MATCHES"
		}
		fmt.Printf("  User: %s%s\n", uid, match)
		fmt.Printf("  Type: %s\n", uidType)
		fmt.Printf("  Role: %s\n", role)
		fmt.Printf("  Status: %s\n\n", status)
	}
}
