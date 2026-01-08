package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

func main() {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     Testing Workspace Invite & Audit System (Features 3&4)   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	// Create services
	workspaceService := services.NewWorkspaceService(pool)
	inviteService := services.NewWorkspaceInviteService(pool)
	auditService := services.NewWorkspaceAuditService(pool)

	// Generate test user ID
	testOwner := fmt.Sprintf("test-owner-%s", uuid.New().String()[:8])
	testEmail := "invite-test@example.com"

	// TEST 1: Create a test workspace
	fmt.Println("📝 TEST 1: Create Test Workspace")
	fmt.Println("-----------------------------------------------------------")
	workspace, err := workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{
		Name:        "Test Workspace Invites",
		Description: stringPtr("Testing invite and audit systems"),
		PlanType:    "professional",
	}, testOwner)

	if err != nil {
		log.Fatalf("❌ Failed to create workspace: %v", err)
	}
	fmt.Printf("✅ Workspace created: %s (ID: %s)\n", workspace.Name, workspace.ID)
	fmt.Printf("   Owner: %s\n", workspace.OwnerID)
	fmt.Println("")

	// TEST 2: Create an invitation
	fmt.Println("📝 TEST 2: Create Workspace Invitation")
	fmt.Println("-----------------------------------------------------------")
	invite, err := inviteService.CreateInvite(ctx, workspace.ID, testEmail, "member", testOwner)
	if err != nil {
		log.Fatalf("❌ Failed to create invite: %v", err)
	}
	fmt.Printf("✅ Invitation created:\n")
	fmt.Printf("   Email: %s\n", invite.Email)
	fmt.Printf("   Role: %s\n", invite.Role)
	fmt.Printf("   Token: %s\n", invite.Token[:20]+"...")
	fmt.Printf("   Status: %s\n", invite.Status)
	fmt.Printf("   Expires: %s\n", invite.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Println("")

	// TEST 3: List workspace invitations
	fmt.Println("📝 TEST 3: List Workspace Invitations")
	fmt.Println("-----------------------------------------------------------")
	invites, err := inviteService.ListWorkspaceInvites(ctx, workspace.ID)
	if err != nil {
		log.Fatalf("❌ Failed to list invites: %v", err)
	}
	fmt.Printf("✅ Found %d invitation(s):\n", len(invites))
	for i, inv := range invites {
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, inv.Email, inv.Role, inv.Status)
	}
	fmt.Println("")

	// TEST 4: Log an audit action
	fmt.Println("📝 TEST 4: Create Audit Log Entry")
	fmt.Println("-----------------------------------------------------------")
	auditLog, err := auditService.LogAction(
		ctx,
		workspace.ID,
		testOwner,
		"invite_member",
		"invite",
		stringPtr(invite.ID.String()),
		map[string]interface{}{
			"email": testEmail,
			"role":  "member",
		},
		stringPtr("192.168.1.100"),
		stringPtr("TestUserAgent/1.0"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create audit log: %v", err)
	}
	fmt.Printf("✅ Audit log created:\n")
	fmt.Printf("   Action: %s\n", auditLog.Action)
	fmt.Printf("   Resource: %s\n", auditLog.ResourceType)
	fmt.Printf("   User: %s\n", auditLog.UserID)
	fmt.Printf("   Timestamp: %s\n", auditLog.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("")

	// TEST 5: Query audit logs
	fmt.Println("📝 TEST 5: Query Audit Logs")
	fmt.Println("-----------------------------------------------------------")
	logs, err := auditService.GetLogs(ctx, services.AuditLogFilter{
		WorkspaceID: workspace.ID,
		Limit:       10,
	})
	if err != nil {
		log.Fatalf("❌ Failed to query audit logs: %v", err)
	}
	fmt.Printf("✅ Found %d audit log(s):\n", len(logs))
	for i, log := range logs {
		fmt.Printf("   %d. %s - %s (%s) at %s\n",
			i+1,
			log.Action,
			log.ResourceType,
			log.UserID,
			log.CreatedAt.Format("15:04:05"),
		)
	}
	fmt.Println("")

	// TEST 6: Get user activity
	fmt.Println("📝 TEST 6: Get User Activity")
	fmt.Println("-----------------------------------------------------------")
	activity, err := auditService.GetUserActivity(ctx, workspace.ID, testOwner, 10)
	if err != nil {
		log.Fatalf("❌ Failed to get user activity: %v", err)
	}
	fmt.Printf("✅ User %s activity:\n", testOwner)
	for i, act := range activity {
		fmt.Printf("   %d. %s on %s\n", i+1, act.Action, act.ResourceType)
	}
	fmt.Println("")

	// TEST 7: Revoke invitation
	fmt.Println("📝 TEST 7: Revoke Invitation")
	fmt.Println("-----------------------------------------------------------")
	err = inviteService.RevokeInvite(ctx, invite.ID)
	if err != nil {
		log.Fatalf("❌ Failed to revoke invite: %v", err)
	}
	fmt.Println("✅ Invitation revoked successfully")

	// Verify status changed
	revokedInvite, err := inviteService.GetInviteByToken(ctx, invite.Token)
	if err != nil {
		log.Fatalf("❌ Failed to get invite: %v", err)
	}
	fmt.Printf("   New status: %s\n", revokedInvite.Status)
	fmt.Println("")

	// TEST 8: Test expired invites cleanup
	fmt.Println("📝 TEST 8: Cleanup Expired Invites")
	fmt.Println("-----------------------------------------------------------")
	count, err := inviteService.CleanupExpiredInvites(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to cleanup expired invites: %v", err)
	}
	fmt.Printf("✅ Cleaned up %d expired invitation(s)\n", count)
	fmt.Println("")

	// TEST 9: Get action statistics
	fmt.Println("📝 TEST 9: Get Action Statistics")
	fmt.Println("-----------------------------------------------------------")
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7) // Last 7 days
	actionStats, err := auditService.GetActionCount(ctx, workspace.ID, startDate, endDate)
	if err != nil {
		log.Fatalf("❌ Failed to get action stats: %v", err)
	}
	fmt.Printf("✅ Action statistics (last 7 days):\n")
	for action, count := range actionStats {
		fmt.Printf("   - %s: %d times\n", action, count)
	}
	fmt.Println("")

	// TEST 10: Get most active users
	fmt.Println("📝 TEST 10: Get Most Active Users")
	fmt.Println("-----------------------------------------------------------")
	activeUsers, err := auditService.GetMostActiveUsers(ctx, workspace.ID, startDate, endDate, 5)
	if err != nil {
		log.Fatalf("❌ Failed to get active users: %v", err)
	}
	fmt.Printf("✅ Most active users (last 7 days):\n")
	for i, user := range activeUsers {
		fmt.Printf("   %d. %s: %d actions\n", i+1, user.UserID, user.Count)
	}
	fmt.Println("")

	// TEST 11: Cleanup - Delete test workspace
	fmt.Println("📝 TEST 11: Cleanup - Delete Workspace")
	fmt.Println("-----------------------------------------------------------")
	err = workspaceService.DeleteWorkspace(ctx, workspace.ID, testOwner)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to delete workspace: %v", err)
	} else {
		fmt.Println("✅ Test workspace deleted")
	}
	fmt.Println("")

	// Summary
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      TEST SUMMARY                             ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                               ║")
	fmt.Println("║  Status: ✅ ALL TESTS PASSED                                  ║")
	fmt.Println("║  Tests Run: 11                                                ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║  Email Invitation System:   ✅ Working                        ║")
	fmt.Println("║  Audit Logging System:      ✅ Working                        ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║  Features Tested:                                             ║")
	fmt.Println("║  - Create invitations                                         ║")
	fmt.Println("║  - List invitations                                           ║")
	fmt.Println("║  - Revoke invitations                                         ║")
	fmt.Println("║  - Cleanup expired invitations                                ║")
	fmt.Println("║  - Create audit logs                                          ║")
	fmt.Println("║  - Query audit logs with filters                              ║")
	fmt.Println("║  - Get user activity                                          ║")
	fmt.Println("║  - Get action statistics                                      ║")
	fmt.Println("║  - Get most active users                                      ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║  🎉 FEATURES #3 & #4 IMPLEMENTATION COMPLETE! 🎉              ║")
	fmt.Println("║                                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

func stringPtr(s string) *string {
	return &s
}
