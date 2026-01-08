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
		dbURL = "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	ctx := context.Background()

	// Get workspace details
	var name, ownerID string
	err = pool.QueryRow(ctx, `
		SELECT name, owner_id FROM workspaces WHERE id = $1
	`, workspaceID).Scan(&name, &ownerID)
	if err != nil {
		log.Fatal("Workspace not found:", err)
	}

	fmt.Printf("📋 Workspace: %s (ID: %s)\n", name, workspaceID)
	fmt.Printf("👤 Owner: %s\n\n", ownerID)

	// Get all members
	rows, err := pool.Query(ctx, `
		SELECT user_id, role, status, joined_at
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY joined_at
	`, workspaceID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("👥 Members:")
	count := 0
	for rows.Next() {
		var userID, role, status string
		var joinedAt interface{}
		if err := rows.Scan(&userID, &role, &status, &joinedAt); err != nil {
			log.Fatal(err)
		}
		count++

		ownerBadge := ""
		if userID == ownerID {
			ownerBadge = " ⭐ (OWNER)"
		}

		fmt.Printf("  %d. %s | Role: %s | Status: %s%s\n", count, userID, role, status, ownerBadge)
	}

	if count == 0 {
		fmt.Println("  ❌ NO MEMBERS FOUND!")
	} else {
		fmt.Printf("\n✅ Total: %d member(s)\n", count)

		// Check if owner is a member
		var ownerRole string
		err = pool.QueryRow(ctx, `
			SELECT role FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2 AND status = 'active'
		`, workspaceID, ownerID).Scan(&ownerRole)

		if err != nil {
			fmt.Println("\n⚠️  WARNING: Owner is NOT registered as an active member!")
			fmt.Println("   This will cause 403 errors when trying to access the workspace.")
		} else {
			fmt.Printf("\n✅ Owner is registered as: %s\n", ownerRole)
			fmt.Println("   Workspace access should work now!")
		}
	}
}
