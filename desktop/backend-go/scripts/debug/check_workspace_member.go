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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")

	// Check workspace exists
	var name, ownerID string
	err = pool.QueryRow(context.Background(), `
		SELECT name, owner_id FROM workspaces WHERE id = $1
	`, workspaceID).Scan(&name, &ownerID)
	if err != nil {
		log.Printf("Workspace query error: %v\n", err)
	} else {
		fmt.Printf("✅ Workspace found: %s (owner: %s)\n", name, ownerID)
	}

	// Check workspace members
	rows, err := pool.Query(context.Background(), `
		SELECT user_id, role, status, joined_at
		FROM workspace_members
		WHERE workspace_id = $1
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("\n📋 Workspace Members:")
	memberCount := 0
	for rows.Next() {
		var userID, role, status string
		var joinedAt interface{}
		if err := rows.Scan(&userID, &role, &status, &joinedAt); err != nil {
			log.Fatal(err)
		}
		memberCount++
		fmt.Printf("  %d. User: %s | Role: %s | Status: %s\n", memberCount, userID, role, status)
	}

	if memberCount == 0 {
		fmt.Println("  ❌ NO MEMBERS FOUND - This is the problem!")
		fmt.Println("\n🔧 Solution: The workspace was created but the owner wasn't added as a member.")
		fmt.Println("   This happens when the transaction failed or the insert was skipped.")
	}
}
