package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Use the user_id from backend logs
	userID := "ZVtQRaictVbO9lN0p-csSA"
	log.Printf("Creating workspace for user: %s", userID)

	// Create a test workspace
	workspaceID := uuid.New()
	query := `
		INSERT INTO workspaces (id, name, slug, plan_type, max_members, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (slug) DO NOTHING
		RETURNING id
	`

	var createdID uuid.UUID
	err = db.QueryRowContext(ctx, query,
		workspaceID,
		"Test Workspace",
		"test-workspace",
		"free",
		10,
		userID,
	).Scan(&createdID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("⚠️  Workspace with slug 'test-workspace' already exists")

			// Get existing workspace
			err = db.QueryRowContext(ctx, "SELECT id FROM workspaces WHERE slug = $1", "test-workspace").Scan(&createdID)
			if err != nil {
				log.Fatalf("❌ Failed to get existing workspace: %v", err)
			}
			log.Printf("✅ Using existing workspace: %s", createdID)
		} else {
			log.Fatalf("❌ Failed to create workspace: %v", err)
		}
	} else {
		log.Printf("✅ Created workspace: %s", createdID)
	}

	// Add user as owner member
	memberQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`

	_, err = db.ExecContext(ctx, memberQuery, createdID, userID)
	if err != nil {
		log.Fatalf("❌ Failed to add member: %v", err)
	}

	log.Println("✅ Added user as workspace owner")

	// Verify workspace was created
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workspaces WHERE owner_id = $1", userID).Scan(&count)
	if err != nil {
		log.Fatalf("❌ Failed to verify workspaces: %v", err)
	}

	log.Printf("✅ User has %d workspace(s)", count)
	log.Println("✅ Test workspace ready!")
}
