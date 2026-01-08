package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Test data
	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	userID := "ZVtQRaictVbO9lN0p-csSA"

	fmt.Println("🧪 Testing Role Context Integration\n")
	fmt.Printf("📋 Workspace: %s\n", workspaceID)
	fmt.Printf("👤 User: %s\n\n", userID)

	// Initialize role context service
	roleCtxService := services.NewRoleContextService(pool)

	// Get role context
	fmt.Println("1️⃣ Fetching role context...")
	roleCtx, err := roleCtxService.GetUserRoleContext(ctx, userID, workspaceID)
	if err != nil {
		log.Fatal("❌ Failed to get role context:", err)
	}

	fmt.Printf("✅ Role context fetched successfully\n\n")

	// Display role context
	fmt.Println("📊 Role Context Details:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Role Name: %s\n", roleCtx.RoleName)
	fmt.Printf("Display Name: %s\n", roleCtx.RoleDisplayName)
	fmt.Printf("Hierarchy Level: %d (1=highest, 6=lowest)\n", roleCtx.HierarchyLevel)

	if roleCtx.Title != "" {
		fmt.Printf("Title: %s\n", roleCtx.Title)
	}
	if roleCtx.Department != "" {
		fmt.Printf("Department: %s\n", roleCtx.Department)
	}

	fmt.Println("\n📋 Permissions:")
	if len(roleCtx.Permissions) == 0 {
		fmt.Println("  ⚠️  No permissions loaded")
	} else {
		for resource, perms := range roleCtx.Permissions {
			fmt.Printf("  • %s: %v\n", resource, perms)
		}
	}

	if len(roleCtx.ProjectRoles) > 0 {
		fmt.Println("\n📁 Project Roles:")
		for projectID, role := range roleCtx.ProjectRoles {
			fmt.Printf("  • %s: %s\n", projectID, role)
		}
	}

	if len(roleCtx.ExpertiseAreas) > 0 {
		fmt.Println("\n🎯 Expertise Areas:")
		for _, area := range roleCtx.ExpertiseAreas {
			fmt.Printf("  • %s\n", area)
		}
	}

	// Generate AI prompt
	fmt.Println("\n2️⃣ Generating AI Prompt...")
	prompt := roleCtx.GetRoleContextPrompt()

	fmt.Println("✅ Prompt generated\n")
	fmt.Println("📝 AI Prompt (first 500 chars):")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if len(prompt) > 500 {
		fmt.Println(prompt[:500] + "...")
	} else {
		fmt.Println(prompt)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Test permission checks
	fmt.Println("\n3️⃣ Testing Permission Checks...")

	testPermissions := []struct {
		resource   string
		permission string
	}{
		{"workspaces", "delete"},
		{"members", "manage"},
		{"projects", "create"},
		{"tasks", "create"},
	}

	for _, test := range testPermissions {
		hasPermission := roleCtx.HasPermission(test.resource, test.permission)
		status := "❌"
		if hasPermission {
			status = "✅"
		}
		fmt.Printf("  %s %s.%s\n", status, test.resource, test.permission)
	}

	// Summary
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ VERIFICATION COMPLETE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if roleCtx.RoleName == "owner" {
		fmt.Println("\n🎯 As an OWNER, the agent should:")
		fmt.Println("  ✅ Confirm you have full permissions")
		fmt.Println("  ✅ Allow all workspace operations")
		fmt.Println("  ✅ Explain you can delete the workspace")
		fmt.Println("  ✅ Suggest advanced management features")
	}

	fmt.Println("\n📝 Next Step:")
	fmt.Println("  Open http://localhost:5173/chat")
	fmt.Println("  Send: 'What can I do in this workspace?'")
	fmt.Println("  Agent should mention your role and permissions!")
}
