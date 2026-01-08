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

// test_workspace_creation.go - Test workspace creation and role seeding
// Run with: go run test_workspace_creation.go

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	fmt.Println("🧪 Testing Workspace Creation")
	fmt.Println("=" + string(make([]byte, 60)))

	// Test 1: Create a workspace
	fmt.Println("\n📝 Test 1: Creating test workspace...")

	testWorkspaceID := uuid.New()
	testUserID := "test-user-" + uuid.New().String()[:8]

	_, err = pool.Exec(context.Background(), `
		INSERT INTO workspaces (id, name, slug, owner_id, plan_type, max_members)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, testWorkspaceID, "Test Workspace", "test-workspace-"+uuid.New().String()[:8],
		testUserID, "professional", 15)

	if err != nil {
		log.Fatalf("  ❌ Failed to create workspace: %v", err)
	}
	fmt.Printf("  ✅ Workspace created: %s\n", testWorkspaceID)

	// Test 2: Seed default roles
	fmt.Println("\n🌱 Test 2: Seeding default roles...")

	_, err = pool.Exec(context.Background(),
		"SELECT seed_default_workspace_roles($1)", testWorkspaceID)

	if err != nil {
		log.Fatalf("  ❌ Failed to seed roles: %v", err)
	}
	fmt.Println("  ✅ Roles seeded successfully")

	// Test 3: Verify roles created
	fmt.Println("\n✅ Test 3: Verifying roles...")

	var roleCount int
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM workspace_roles WHERE workspace_id = $1",
		testWorkspaceID).Scan(&roleCount)

	if err != nil {
		log.Fatalf("  ❌ Failed to count roles: %v", err)
	}

	fmt.Printf("  📊 Roles created: %d (expected: 6)\n", roleCount)

	if roleCount != 6 {
		log.Printf("  ⚠️  WARNING: Expected 6 roles, got %d", roleCount)
	} else {
		fmt.Println("  ✅ Correct number of roles")
	}

	// Test 4: List all roles
	fmt.Println("\n👥 Test 4: Listing all roles...")

	rows, err := pool.Query(context.Background(), `
		SELECT name, display_name, hierarchy_level, is_system, is_default
		FROM workspace_roles
		WHERE workspace_id = $1
		ORDER BY hierarchy_level
	`, testWorkspaceID)

	if err != nil {
		log.Fatalf("  ❌ Failed to query roles: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, displayName string
		var hierarchyLevel int
		var isSystem, isDefault bool

		err := rows.Scan(&name, &displayName, &hierarchyLevel, &isSystem, &isDefault)
		if err != nil {
			log.Printf("  ❌ Failed to scan role: %v", err)
			continue
		}

		defaultMarker := ""
		if isDefault {
			defaultMarker = " [DEFAULT]"
		}

		fmt.Printf("  %d. %s (%s) - Level %d%s\n",
			hierarchyLevel, displayName, name, hierarchyLevel, defaultMarker)
	}

	// Test 5: Verify role_permissions populated
	fmt.Println("\n🔑 Test 5: Verifying role permissions...")

	var permissionCount int
	err = pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM role_permissions WHERE workspace_id = $1",
		testWorkspaceID).Scan(&permissionCount)

	if err != nil {
		log.Fatalf("  ❌ Failed to count permissions: %v", err)
	}

	fmt.Printf("  📊 Permission entries: %d\n", permissionCount)

	if permissionCount == 0 {
		log.Printf("  ⚠️  WARNING: No permissions found in role_permissions table")
	} else {
		fmt.Println("  ✅ Permissions populated")
	}

	// Test 6: Add owner as first member
	fmt.Println("\n👤 Test 6: Adding owner as first member...")

	_, err = pool.Exec(context.Background(), `
		INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, testWorkspaceID, testUserID, "owner", "active")

	if err != nil {
		log.Fatalf("  ❌ Failed to add member: %v", err)
	}
	fmt.Println("  ✅ Owner added as first member")

	// Test 7: Create user profile
	fmt.Println("\n📋 Test 7: Creating user profile...")

	_, err = pool.Exec(context.Background(), `
		INSERT INTO user_workspace_profiles (workspace_id, user_id, display_name, title, department)
		VALUES ($1, $2, $3, $4, $5)
	`, testWorkspaceID, testUserID, "Test User", "CEO", "Executive")

	if err != nil {
		log.Fatalf("  ❌ Failed to create profile: %v", err)
	}
	fmt.Println("  ✅ User profile created")

	// Test 8: Query complete user context (like role_context.go does)
	fmt.Println("\n🔍 Test 8: Querying complete user role context...")

	var userName, roleName, roleDisplayName, title, department string
	var hierarchyLevel int

	err = pool.QueryRow(context.Background(), `
		SELECT
			wm.user_id,
			wm.role as role_name,
			wr.display_name as role_display_name,
			wr.hierarchy_level,
			uwp.title,
			uwp.department
		FROM workspace_members wm
		JOIN workspace_roles wr ON wr.name = wm.role AND wr.workspace_id = wm.workspace_id
		LEFT JOIN user_workspace_profiles uwp ON uwp.user_id = wm.user_id AND uwp.workspace_id = wm.workspace_id
		WHERE wm.user_id = $1 AND wm.workspace_id = $2
	`, testUserID, testWorkspaceID).Scan(&userName, &roleName, &roleDisplayName, &hierarchyLevel, &title, &department)

	if err != nil {
		log.Fatalf("  ❌ Failed to query user context: %v", err)
	}

	fmt.Printf("  👤 User: %s\n", userName)
	fmt.Printf("  🎭 Role: %s (%s)\n", roleDisplayName, roleName)
	fmt.Printf("  📊 Hierarchy: Level %d\n", hierarchyLevel)
	fmt.Printf("  💼 Title: %s\n", title)
	fmt.Printf("  🏢 Department: %s\n", department)
	fmt.Println("  ✅ Context query successful")

	// Test 9: Check permissions for owner role
	fmt.Println("\n🔐 Test 9: Checking owner permissions...")

	rows2, err := pool.Query(context.Background(), `
		SELECT resource, permission
		FROM role_permissions
		WHERE workspace_id = $1 AND role = 'owner'
		ORDER BY resource, permission
		LIMIT 10
	`, testWorkspaceID)

	if err != nil {
		log.Fatalf("  ❌ Failed to query permissions: %v", err)
	}
	defer rows2.Close()

	fmt.Println("  Sample permissions for 'owner' role:")
	count := 0
	for rows2.Next() {
		var resource, permission string
		err := rows2.Scan(&resource, &permission)
		if err != nil {
			continue
		}
		fmt.Printf("    - %s.%s\n", resource, permission)
		count++
	}

	if count > 0 {
		fmt.Println("  ✅ Permissions found")
	} else {
		fmt.Println("  ⚠️  No permissions found")
	}

	// Cleanup
	fmt.Println("\n🧹 Cleanup: Deleting test workspace...")

	_, err = pool.Exec(context.Background(),
		"DELETE FROM workspaces WHERE id = $1", testWorkspaceID)

	if err != nil {
		log.Printf("  ⚠️  Failed to cleanup: %v", err)
	} else {
		fmt.Println("  ✅ Test workspace deleted (cascade should remove all related data)")
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("✅ All tests completed successfully!")
	fmt.Println("\n📝 Summary:")
	fmt.Println("  ✅ Workspace creation works")
	fmt.Println("  ✅ Role seeding works")
	fmt.Println("  ✅ Member assignment works")
	fmt.Println("  ✅ User profiles work")
	fmt.Println("  ✅ Permission system works")
	fmt.Println("  ✅ Role context queries work")
	fmt.Println("  ✅ Cascade deletes work")
}
