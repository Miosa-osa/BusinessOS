//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const devWorkspaceID = "00000000-0000-0000-0000-000000000001"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("🔧 Setting up development workspace...")

	// Parse the fixed workspace ID
	workspaceID, _ := uuid.Parse(devWorkspaceID)

	// Check if workspace already exists
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1)", workspaceID).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check workspace: %v", err)
	}

	if !exists {
		// Create the dev workspace
		_, err = pool.Exec(ctx, `
			INSERT INTO workspaces (
				id, name, slug, description, plan_type,
				max_members, max_projects, max_storage_gb,
				owner_id, settings, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`,
			workspaceID,
			"Development Workspace",
			"dev-workspace",
			"Auto-created workspace for local development and testing",
			"professional",
			50, 100, 50,
			"system",
			`{"dev_mode": true}`,
		)
		if err != nil {
			log.Fatalf("Failed to create workspace: %v", err)
		}
		fmt.Printf("✅ Created dev workspace: %s\n", workspaceID)
	} else {
		fmt.Printf("✅ Dev workspace already exists: %s\n", workspaceID)
	}

	// Check/create owner role
	var ownerRoleID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM workspace_roles
		WHERE workspace_id = $1 AND name = 'owner'
	`, workspaceID).Scan(&ownerRoleID)

	if err != nil {
		// Create owner role
		ownerRoleID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO workspace_roles (
				id, workspace_id, name, display_name, description,
				color, icon, hierarchy_level, is_system, is_default,
				permissions, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		`,
			ownerRoleID,
			workspaceID,
			"owner",
			"Owner",
			"Full workspace access",
			"#10b981",
			"crown",
			0,
			true,
			false,
			`{
				"workspace": {"manage": true, "delete": true, "transfer": true},
				"members": {"view": true, "invite": true, "manage": true, "remove": true},
				"roles": {"view": true, "create": true, "manage": true, "delete": true},
				"projects": {"view": true, "create": true, "manage": true, "delete": true},
				"settings": {"view": true, "manage": true},
				"billing": {"view": true, "manage": true},
				"apps": {"view": true, "create": true, "manage": true, "delete": true}
			}`,
		)
		if err != nil {
			log.Fatalf("Failed to create owner role: %v", err)
		}
		fmt.Println("✅ Created owner role")
	} else {
		fmt.Println("✅ Owner role already exists")
	}

	// Get all users who aren't members yet
	rows, err := pool.Query(ctx, `
		SELECT u.id, u.email
		FROM "user" u
		WHERE NOT EXISTS (
			SELECT 1 FROM workspace_members wm
			WHERE wm.workspace_id = $1 AND wm.user_id = u.id
		)
	`, workspaceID)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	addedCount := 0
	var firstUserID string
	for rows.Next() {
		var userID, email string
		if err := rows.Scan(&userID, &email); err != nil {
			log.Printf("Warning: Failed to scan user: %v", err)
			continue
		}

		// Add user as workspace member (use role column, matching production schema)
		_, err = pool.Exec(ctx, `
			INSERT INTO workspace_members (
				workspace_id, user_id, role, status, joined_at
			) VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (workspace_id, user_id) DO NOTHING
		`, workspaceID, userID, "owner", "active")

		if err != nil {
			log.Printf("Warning: Failed to add user %s: %v", email, err)
			continue
		}

		if firstUserID == "" {
			firstUserID = userID
		}
		addedCount++
		fmt.Printf("✅ Added user %s (%s) as owner\n", email, userID)
	}

	// Update workspace owner if needed
	if firstUserID != "" {
		pool.Exec(ctx, `
			UPDATE workspaces SET owner_id = $1
			WHERE id = $2 AND owner_id = 'system'
		`, firstUserID, workspaceID)
	}

	// Final verification
	var memberCount int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1
	`, workspaceID).Scan(&memberCount)

	fmt.Println("")
	fmt.Println("📊 Summary:")
	fmt.Printf("   Workspace ID: %s\n", workspaceID)
	fmt.Printf("   Members: %d\n", memberCount)
	fmt.Printf("   New members added: %d\n", addedCount)
	fmt.Println("")
	fmt.Println("🎉 Dev workspace is ready! Refresh the browser to use it.")
}
