package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Step 1: Create project_members table
	fmt.Println("Step 1: Creating project_members table...")
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS project_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			can_edit BOOLEAN NOT NULL DEFAULT true,
			can_delete BOOLEAN NOT NULL DEFAULT false,
			can_invite BOOLEAN NOT NULL DEFAULT false,
			assigned_by TEXT NOT NULL,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			removed_at TIMESTAMPTZ,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT unique_project_member UNIQUE (project_id, user_id),
			CONSTRAINT check_project_role CHECK (role IN ('lead', 'contributor', 'reviewer', 'viewer')),
			CONSTRAINT check_member_status CHECK (status IN ('active', 'inactive', 'removed'))
		)
	`)
	if err != nil {
		log.Printf("Error creating table: %v", err)
	} else {
		fmt.Println("✅ Table created")
	}

	// Step 2: Create indexes
	fmt.Println("\nStep 2: Creating indexes...")
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_workspace_id ON project_members(workspace_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_status ON project_members(status)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_role ON project_members(role)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_project_user ON project_members(project_id, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_workspace_user ON project_members(workspace_id, user_id)",
	}

	for _, idx := range indexes {
		_, err = pool.Exec(ctx, idx)
		if err != nil {
			log.Printf("Error creating index: %v", err)
		}
	}
	fmt.Println("✅ Indexes created")

	// Step 3: Create trigger function
	fmt.Println("\nStep 3: Creating trigger function...")
	_, err = pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION update_project_members_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	if err != nil {
		log.Printf("Error creating function: %v", err)
	} else {
		fmt.Println("✅ Function created")
	}

	// Step 4: Create trigger
	fmt.Println("\nStep 4: Creating trigger...")
	_, err = pool.Exec(ctx, `
		DROP TRIGGER IF EXISTS trigger_update_project_members_updated_at ON project_members;
		CREATE TRIGGER trigger_update_project_members_updated_at
			BEFORE UPDATE ON project_members
			FOR EACH ROW
			EXECUTE FUNCTION update_project_members_updated_at()
	`)
	if err != nil {
		log.Printf("Error creating trigger: %v", err)
	} else {
		fmt.Println("✅ Trigger created")
	}

	// Step 5: Create project_role_definitions
	fmt.Println("\nStep 5: Creating project_role_definitions table...")
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS project_role_definitions (
			role TEXT PRIMARY KEY,
			display_name TEXT NOT NULL,
			description TEXT,
			hierarchy_level INT NOT NULL,
			default_can_edit BOOLEAN NOT NULL DEFAULT false,
			default_can_delete BOOLEAN NOT NULL DEFAULT false,
			default_can_invite BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Error creating role definitions table: %v", err)
	} else {
		fmt.Println("✅ Role definitions table created")
	}

	// Step 6: Seed default roles
	fmt.Println("\nStep 6: Seeding default roles...")
	_, err = pool.Exec(ctx, `
		INSERT INTO project_role_definitions (role, display_name, description, hierarchy_level, default_can_edit, default_can_delete, default_can_invite)
		VALUES
			('lead', 'Project Lead', 'Full project control, can manage members and settings', 1, true, true, true),
			('contributor', 'Contributor', 'Can edit and contribute to project', 2, true, false, false),
			('reviewer', 'Reviewer', 'Can review and comment, limited editing', 3, false, false, false),
			('viewer', 'Viewer', 'Read-only access to project', 4, false, false, false)
		ON CONFLICT (role) DO NOTHING
	`)
	if err != nil {
		log.Printf("Error seeding roles: %v", err)
	} else {
		fmt.Println("✅ Default roles seeded")
	}

	fmt.Println("\n✅ Migration 029 completed successfully!")
}
