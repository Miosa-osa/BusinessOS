//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Generate fake email metadata for testing onboarding without Gmail API
// Usage: go run scripts/generate_test_emails.go [workspace_id] [user_id]

var (
	// Sample email domains (realistic)
	senderDomains = []string{
		"slack.com", "github.com", "notion.so", "figma.com", "linear.app",
		"asana.com", "trello.com", "monday.com", "jira.atlassian.com",
		"google.com", "microsoft.com", "zoom.us", "calendly.com",
	}

	// Sample subject keywords
	subjectKeywords = []string{
		"project", "deadline", "meeting", "review", "update",
		"feedback", "team", "design", "code", "deploy",
		"urgent", "reminder", "invitation", "notification", "alert",
	}

	// Sample body keywords
	bodyKeywords = []string{
		"Hi team", "Let's discuss", "Can you review", "Thanks for",
		"Please update", "Status update", "Next steps", "Action items",
		"Follow up", "Quick question", "FYI", "Heads up",
	}

	// Sample detected tools
	detectedTools = []string{
		"Slack", "GitHub", "Notion", "Figma", "Linear",
		"Asana", "Trello", "Monday", "Jira", "Google Docs",
		"Zoom", "Calendly", "VS Code", "Docker",
	}

	// Sample topics
	topics = []string{
		"project_management", "code_review", "design_feedback",
		"team_meeting", "sprint_planning", "bug_tracking",
		"documentation", "deployment", "client_communication",
	}
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║            GENERATE TEST EMAIL METADATA                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Warning: .env file not found")
	}

	ctx := context.Background()

	// Parse arguments
	var workspaceID, userID string
	if len(os.Args) > 2 {
		workspaceID = os.Args[1]
		userID = os.Args[2]
	} else {
		fmt.Println("Usage: go run scripts/generate_test_emails.go [workspace_id] [user_id]")
		fmt.Println()
		fmt.Println("Creating test user and workspace...")
		fmt.Println()

		// Create test user and workspace
		dbURL := os.Getenv("DATABASE_URL")
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			fmt.Printf("❌ Database connection failed: %v\n", err)
			os.Exit(1)
		}
		defer pool.Close()

		userID = uuid.New().String()
		workspaceUUID := uuid.New()
		email := fmt.Sprintf("test-emails-%d@example.com", time.Now().Unix())
		slug := fmt.Sprintf("test-emails-%d", time.Now().Unix())

		_, err = pool.Exec(ctx, `
			INSERT INTO "user" (id, email, name, "emailVerified")
			VALUES ($1, $2, $3, true)
		`, userID, email, "Test Email User")

		if err != nil {
			fmt.Printf("❌ Failed to create user: %v\n", err)
			os.Exit(1)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO workspaces (id, name, slug, owner_id)
			VALUES ($1, $2, $3, $4)
		`, workspaceUUID, "Test Email Workspace", slug, userID)

		if err != nil {
			fmt.Printf("❌ Failed to create workspace: %v\n", err)
			os.Exit(1)
		}

		workspaceID = workspaceUUID.String()

		fmt.Printf("✅ Created test user: %s\n", email)
		fmt.Printf("✅ Created test workspace: %s\n", slug)
		fmt.Printf("   Workspace ID: %s\n", workspaceID)
		fmt.Printf("   User ID: %s\n", userID)
		fmt.Println()
	}

	// First create an analysis record
	fmt.Println("📧 Creating analysis record...")
	fmt.Println("─────────────────────────────────────────────────────────────")

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create analysis record first
	var analysisID uuid.UUID
	err = pool.QueryRow(ctx, `
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
		RETURNING id
	`, userID, workspaceID,
		`["Email analysis user"]`,
		`["Communication", "Collaboration"]`,
		`["Slack", "Gmail", "GitHub"]`,
		"User communicates frequently with team and external partners.",
		"test-model", "test-provider", "completed", time.Now()).Scan(&analysisID)

	if err != nil {
		fmt.Printf("❌ Failed to create analysis: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created analysis: %s\n", analysisID)
	fmt.Println()

	// Generate individual email records
	fmt.Println("📧 Generating 20 fake emails...")
	fmt.Println("─────────────────────────────────────────────────────────────")

	emailCount := 20
	now := time.Now()
	domainCounts := make(map[string]int)
	toolCounts := make(map[string]int)
	topicCounts := make(map[string]int)

	for i := 0; i < emailCount; i++ {
		daysAgo := rand.Intn(30)
		emailDate := now.AddDate(0, 0, -daysAgo)

		domain := senderDomains[rand.Intn(len(senderDomains))]
		domainCounts[domain]++

		// Select random keywords/tools/topics for this email
		numTools := rand.Intn(3) + 1
		emailTools := make([]string, 0, numTools)
		for j := 0; j < numTools; j++ {
			tool := detectedTools[rand.Intn(len(detectedTools))]
			emailTools = append(emailTools, tool)
			toolCounts[tool]++
		}

		numTopics := rand.Intn(2) + 1
		emailTopics := make([]string, 0, numTopics)
		for j := 0; j < numTopics; j++ {
			topic := topics[rand.Intn(len(topics))]
			emailTopics = append(emailTopics, topic)
			topicCounts[topic]++
		}

		numSubjectKw := rand.Intn(3) + 1
		emailSubjectKw := make([]string, 0, numSubjectKw)
		for j := 0; j < numSubjectKw; j++ {
			emailSubjectKw = append(emailSubjectKw, subjectKeywords[rand.Intn(len(subjectKeywords))])
		}

		numBodyKw := rand.Intn(4) + 1
		emailBodyKw := make([]string, 0, numBodyKw)
		for j := 0; j < numBodyKw; j++ {
			emailBodyKw = append(emailBodyKw, bodyKeywords[rand.Intn(len(bodyKeywords))])
		}

		toolsJSON, _ := json.Marshal(emailTools)
		topicsJSON, _ := json.Marshal(emailTopics)
		subjectJSON, _ := json.Marshal(emailSubjectKw)
		bodyJSON, _ := json.Marshal(emailBodyKw)

		// Insert individual email
		_, err = pool.Exec(ctx, `
			INSERT INTO onboarding_email_metadata (
				user_id,
				analysis_id,
				external_id,
				sender_domain,
				sender_email,
				subject_keywords,
				body_keywords,
				detected_tools,
				detected_topics,
				category,
				sentiment,
				importance_score,
				email_date
			) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8::jsonb, $9::jsonb, $10, $11, $12, $13)
		`, userID, analysisID,
			fmt.Sprintf("fake-email-%d", i),
			domain,
			fmt.Sprintf("sender%d@%s", i, domain),
			string(subjectJSON), string(bodyJSON), string(toolsJSON), string(topicsJSON),
			"work", "neutral", float64(rand.Intn(100))/100.0, emailDate)

		if err != nil {
			fmt.Printf("❌ Failed to insert email %d: %v\n", i, err)
			continue
		}
	}

	fmt.Printf("✅ Generated %d fake emails\n", emailCount)
	fmt.Println()

	// Show what was generated
	fmt.Println("📊 Generated Email Metadata:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("Total Emails:    %d\n", emailCount)
	fmt.Println()

	fmt.Println("Top Sender Domains:")
	for domain, count := range domainCounts {
		fmt.Printf("  • %s: %d emails\n", domain, count)
	}
	fmt.Println()

	fmt.Println("Detected Tools:")
	for tool, count := range toolCounts {
		fmt.Printf("  • %s: %d mentions\n", tool, count)
	}
	fmt.Println()

	fmt.Println("Discussion Topics:")
	for topic, count := range topicCounts {
		fmt.Printf("  • %s: %d occurrences\n", topic, count)
	}
	fmt.Println()

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  GENERATION COMPLETE                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Fake email metadata inserted into database")
	fmt.Println()
	fmt.Println("🎯 Use this data to test:")
	fmt.Println("   • AI profile analysis (without Gmail API)")
	fmt.Println("   • Template matching algorithm")
	fmt.Println("   • App recommendation engine")
	fmt.Println()
	fmt.Println("📋 Next steps:")
	fmt.Println("   1. Trigger AI analysis on this workspace")
	fmt.Println("   2. Check results in onboarding_user_analysis table")
	fmt.Println("   3. Verify recommendations in app_generation_queue")
	fmt.Println()
	fmt.Printf("   Workspace ID: %s\n", workspaceID)
	fmt.Printf("   User ID: %s\n", userID)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
