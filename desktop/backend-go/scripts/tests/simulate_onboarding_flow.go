//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Simulate complete onboarding flow (API level testing)
// Usage: go run scripts/simulate_onboarding_flow.go
// This tests the backend API endpoints without requiring browser/OAuth

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║         SIMULATE ONBOARDING FLOW (API LEVEL TEST)               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("This script simulates the onboarding flow by calling backend APIs")
	fmt.Println("directly, bypassing the need for browser/OAuth during testing.")
	fmt.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Warning: .env file not found")
	}

	ctx := context.Background()

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Step 1: Create test user directly in DB
	fmt.Println("🔍 STEP 1: Create Test User")
	fmt.Println("─────────────────────────────────────────────────────────────")

	userID := uuid.New().String()
	workspaceID := uuid.New()
	email := fmt.Sprintf("test-sim-%d@example.com", time.Now().Unix())

	_, err = pool.Exec(ctx, `
		INSERT INTO "user" (id, email, name, "emailVerified")
		VALUES ($1, $2, $3, true)
	`, userID, email, "Test Sim User")

	if err != nil {
		fmt.Printf("❌ Failed to create user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created user: %s\n", email)
	fmt.Printf("   User ID: %s\n", userID)
	fmt.Println()

	// Step 2: Create workspace
	fmt.Println("🔍 STEP 2: Create Workspace")
	fmt.Println("─────────────────────────────────────────────────────────────")

	workspaceSlug := fmt.Sprintf("test-sim-%d", time.Now().Unix())
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (id, name, slug, owner_id)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, "Test Sim Workspace", workspaceSlug, userID)

	if err != nil {
		fmt.Printf("❌ Failed to create workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created workspace: %s\n", workspaceSlug)
	fmt.Printf("   Workspace ID: %s\n", workspaceID)
	fmt.Println()

	// Step 3: Insert fake AI analysis
	fmt.Println("🔍 STEP 3: Insert AI Analysis (Simulated)")
	fmt.Println("─────────────────────────────────────────────────────────────")

	insights := []string{
		"Project management enthusiast",
		"Collaborative team player",
		"Productivity-focused workflow",
	}
	interests := []string{
		"Project Management",
		"Team Collaboration",
		"Task Automation",
	}
	toolsUsed := []string{"Slack", "Trello", "GitHub"}

	insightsJSON, _ := json.Marshal(insights)
	interestsJSON, _ := json.Marshal(interests)
	toolsJSON, _ := json.Marshal(toolsUsed)

	_, err = pool.Exec(ctx, `
		INSERT INTO onboarding_user_analysis (
			user_id,
			workspace_id,
			insights,
			interests,
			tools_used,
			profile_summary,
			analysis_model,
			ai_provider,
			status,
			completed_at
		) VALUES ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6, $7, $8, $9, $10)
	`, userID, workspaceID, string(insightsJSON), string(interestsJSON), string(toolsJSON),
		"User works with project management tools and values team collaboration.",
		"groq-llama3", "groq", "completed", time.Now())

	if err != nil {
		fmt.Printf("❌ Failed to insert AI analysis: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ AI analysis inserted")
	fmt.Printf("   Insights: %v\n", insights)
	fmt.Printf("   Interests: %v\n", interests)
	fmt.Printf("   Tools: %v\n", toolsUsed)
	fmt.Println()

	// Step 4: Create onboarding profile
	fmt.Println("🔍 STEP 4: Create Onboarding Profile")
	fmt.Println("─────────────────────────────────────────────────────────────")

	recommendedIntegrations := []string{"Slack", "GitHub", "Google Calendar"}
	integrationsJSON, _ := json.Marshal(recommendedIntegrations)

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_onboarding_profiles (
			workspace_id,
			business_type,
			team_size,
			owner_role,
			main_challenge,
			recommended_integrations
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb)
	`, workspaceID, "startup", "small", "founder", "project_management",
		string(integrationsJSON))

	if err != nil {
		fmt.Printf("❌ Failed to create profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Onboarding profile created")
	fmt.Println("   Business Type: startup")
	fmt.Println("   Team Size: small")
	fmt.Println("   Main Challenge: project_management")
	fmt.Println()

	// Step 5: Trigger app generation (via API if available)
	fmt.Println("🔍 STEP 5: Queue App Generation")
	fmt.Println("─────────────────────────────────────────────────────────────")

	// Get template IDs
	rows, err := pool.Query(ctx, `
		SELECT id, template_name FROM app_templates
		ORDER BY priority_score DESC
		LIMIT 3
	`)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch templates: %v\n", err)
	} else {
		defer rows.Close()

		queuedCount := 0
		for rows.Next() {
			var templateID uuid.UUID
			var templateName string
			rows.Scan(&templateID, &templateName)

			// Insert into queue
			_, err := pool.Exec(ctx, `
				INSERT INTO app_generation_queue (
					workspace_id,
					template_id,
					status,
					priority,
					generation_context
				) VALUES ($1, $2, 'pending', 80, '{"simulated": true}'::jsonb)
			`, workspaceID, templateID)

			if err == nil {
				fmt.Printf("✅ Queued: %s\n", templateName)
				queuedCount++
			}
		}

		if queuedCount > 0 {
			fmt.Printf("\n✅ %d apps queued for generation\n", queuedCount)
		}
	}
	fmt.Println()

	// Step 6: Create a mock generated app
	fmt.Println("🔍 STEP 6: Create Mock Generated App")
	fmt.Println("─────────────────────────────────────────────────────────────")

	var templateID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM app_templates
		ORDER BY priority_score DESC
		LIMIT 1
	`).Scan(&templateID)

	if err == nil {
		osaAppID := uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO user_generated_apps (
				workspace_id,
				template_id,
				app_name,
				osa_app_id,
				is_visible,
				is_pinned
			) VALUES ($1, $2, $3, $4, true, false)
		`, workspaceID, templateID, "Test Generated App", osaAppID)

		if err != nil {
			fmt.Printf("⚠️  Could not create mock app: %v\n", err)
		} else {
			fmt.Println("✅ Mock generated app created")
			fmt.Printf("   App Name: Test Generated App\n")
			fmt.Printf("   OSA App ID: %s\n", osaAppID)
		}
	}
	fmt.Println()

	// Step 7: Verify data in dashboard
	fmt.Println("🔍 STEP 7: Verify Data")
	fmt.Println("─────────────────────────────────────────────────────────────")

	var analysisCount, profileCount, queueCount, appCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM onboarding_user_analysis WHERE workspace_id = $1`, workspaceID).Scan(&analysisCount)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM workspace_onboarding_profiles WHERE workspace_id = $1`, workspaceID).Scan(&profileCount)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM app_generation_queue WHERE workspace_id = $1`, workspaceID).Scan(&queueCount)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_generated_apps WHERE workspace_id = $1`, workspaceID).Scan(&appCount)

	fmt.Printf("✅ AI Analyses: %d\n", analysisCount)
	fmt.Printf("✅ Onboarding Profiles: %d\n", profileCount)
	fmt.Printf("✅ Queued Apps: %d\n", queueCount)
	fmt.Printf("✅ Generated Apps: %d\n", appCount)
	fmt.Println()

	// Final summary
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    SIMULATION COMPLETE                           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Simulated onboarding flow completed successfully!")
	fmt.Println()
	fmt.Printf("📊 Test Data Created:\n")
	fmt.Printf("   User:             %s\n", email)
	fmt.Printf("   Workspace:        %s\n", workspaceSlug)
	fmt.Printf("   AI Analysis:      %d record(s)\n", analysisCount)
	fmt.Printf("   Profile:          %d record(s)\n", profileCount)
	fmt.Printf("   Queued Apps:      %d record(s)\n", queueCount)
	fmt.Printf("   Generated Apps:   %d record(s)\n", appCount)
	fmt.Println()
	fmt.Println("🎯 Use this data to test:")
	fmt.Println("   • Dashboard display")
	fmt.Println("   • App generation worker")
	fmt.Println("   • Profile recommendations")
	fmt.Println()
	fmt.Println("🧹 Cleanup: go run scripts/cleanup_test_data.go --dry-run")
}
