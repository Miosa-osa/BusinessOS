package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fixed project seed IDs
var projectIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000501"), // Website Redesign 2026
	uuid.MustParse("00000000-5eed-4000-a000-000000000502"), // Mobile App MVP
	uuid.MustParse("00000000-5eed-4000-a000-000000000503"), // Internal Knowledge Base
	uuid.MustParse("00000000-5eed-4000-a000-000000000504"), // CRM Integration
	uuid.MustParse("00000000-5eed-4000-a000-000000000505"), // Q1 Marketing Campaign
	uuid.MustParse("00000000-5eed-4000-a000-000000000506"), // ML Research
}

var projectMemberIDs []uuid.UUID

func init() {
	for i := 1; i <= 30; i++ {
		projectMemberIDs = append(projectMemberIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 550+i)))
	}
}

func seedProjects(ctx context.Context, pool *pgxpool.Pool, userID string) {
	type project struct {
		id          uuid.UUID
		name        string
		desc        string
		status      string
		priority    string
		clientIdx   *int   // index into clientIDs, nil = no client
		projectType string
		startDays   int // days ago
		dueDays     int // positive = future, negative = past
		completed   bool
	}

	ci := func(i int) *int { return &i }

	projects := []project{
		{projectIDs[0], "Website Redesign 2026", "Complete overhaul of corporate website with modern design, improved UX, and mobile-first responsive layout. Includes SEO optimization and performance improvements.", "ACTIVE", "HIGH", ci(0), "client", 30, 14, false},
		{projectIDs[1], "Mobile App MVP", "Build cross-platform mobile application for manufacturing floor monitoring. Real-time sensor data, alerts, and reporting dashboards.", "ACTIVE", "CRITICAL", ci(1), "client", 14, 45, false},
		{projectIDs[2], "Internal Knowledge Base", "Build a searchable internal wiki for team documentation, onboarding guides, and process documentation. Integrates with existing tools.", "ACTIVE", "MEDIUM", nil, "internal", 60, 30, false},
		{projectIDs[3], "CRM Integration Project", "Integrate our CRM with Meridian Healthcare's existing patient management system. HIPAA compliance required.", "PAUSED", "MEDIUM", ci(3), "client", 45, 60, false},
		{projectIDs[4], "Q1 Marketing Campaign", "Multi-channel marketing push for new product launch. Includes email sequences, social media, content marketing, and paid advertising.", "COMPLETED", "HIGH", nil, "internal", 90, -10, true},
		{projectIDs[5], "Machine Learning Research", "Exploratory research into ML models for predictive analytics. Proof-of-concept for customer churn prediction.", "ARCHIVED", "LOW", nil, "learning", 120, -30, false},
	}

	for _, p := range projects {
		var clientID *uuid.UUID
		if p.clientIdx != nil {
			clientID = &clientIDs[*p.clientIdx]
		}

		var startExpr string
		startExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", p.startDays)

		var dueExpr string
		if p.dueDays >= 0 {
			dueExpr = fmt.Sprintf("CURRENT_DATE + INTERVAL '%d days'", p.dueDays)
		} else {
			dueExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", -p.dueDays)
		}

		completedExpr := "NULL"
		if p.completed {
			completedExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", -p.dueDays+5)
		}

		q := fmt.Sprintf(`
			INSERT INTO projects (id, user_id, name, description, status, priority, client_id, project_type,
				start_date, due_date, completed_at, visibility, owner_id)
			VALUES ($1, $2, $3, $4, $5::projectstatus, $6::projectpriority, $7, $8,
				%s, %s, %s, 'private', $9)
			ON CONFLICT (id) DO NOTHING`, startExpr, dueExpr, completedExpr)

		_, err := pool.Exec(ctx, q,
			p.id, userID, p.name, p.desc, p.status, p.priority, clientID, p.projectType, userID,
		)
		if err != nil {
			log.Printf("  project %s: %v", p.name, err)
		} else {
			fmt.Printf("  + Project: %s [%s/%s]\n", p.name, p.status, p.priority)
		}
	}

	// --- Project Members ---
	// project_members requires workspace_id (NOT NULL). Look it up.
	var workspaceID *uuid.UUID
	var wsID uuid.UUID
	wsErr := pool.QueryRow(ctx, `SELECT workspace_id FROM workspace_members WHERE user_id = $1 LIMIT 1`, userID).Scan(&wsID)
	if wsErr == nil {
		workspaceID = &wsID
	} else {
		log.Printf("  warning: no workspace found for user, skipping project members: %v", wsErr)
	}

	if workspaceID != nil {
		for i, pid := range projectIDs {
			_, err := pool.Exec(ctx, `
				INSERT INTO project_members (id, project_id, user_id, workspace_id, role, assigned_by)
				VALUES ($1, $2, $3, $4, 'lead', $5)
				ON CONFLICT (id) DO NOTHING`,
				projectMemberIDs[i], pid, userID, *workspaceID, userID,
			)
			if err != nil {
				log.Printf("  project_member for %s: %v", pid, err)
			}
		}
		fmt.Printf("  + %d project members (owner)\n", len(projectIDs))
	}
}
