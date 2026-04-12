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
		clientName  string // stored as client_name (no FK)
		projectType string
	}

	projects := []project{
		{projectIDs[0], "Website Redesign 2026", "Complete overhaul of corporate website with modern design, improved UX, and mobile-first responsive layout. Includes SEO optimization and performance improvements.", "ACTIVE", "HIGH", "Nexus Digital Agency", "client"},
		{projectIDs[1], "Mobile App MVP", "Build cross-platform mobile application for manufacturing floor monitoring. Real-time sensor data, alerts, and reporting dashboards.", "ACTIVE", "CRITICAL", "Apex Manufacturing Co.", "client"},
		{projectIDs[2], "Internal Knowledge Base", "Build a searchable internal wiki for team documentation, onboarding guides, and process documentation. Integrates with existing tools.", "ACTIVE", "MEDIUM", "", "internal"},
		{projectIDs[3], "CRM Integration Project", "Integrate our CRM with Meridian Healthcare's existing patient management system. HIPAA compliance required.", "PAUSED", "MEDIUM", "Meridian Healthcare Group", "client"},
		{projectIDs[4], "Q1 Marketing Campaign", "Multi-channel marketing push for new product launch. Includes email sequences, social media, content marketing, and paid advertising.", "COMPLETED", "HIGH", "", "internal"},
		{projectIDs[5], "Machine Learning Research", "Exploratory research into ML models for predictive analytics. Proof-of-concept for customer churn prediction.", "ARCHIVED", "LOW", "", "learning"},
	}

	for _, p := range projects {
		var clientNameParam *string
		if p.clientName != "" {
			clientNameParam = &p.clientName
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
			VALUES ($1, $2, $3, $4, $5::projectstatus, $6::projectpriority, $7, $8)
			ON CONFLICT (id) DO NOTHING`,
			p.id, userID, p.name, p.desc, p.status, p.priority, clientNameParam, p.projectType,
		)
		if err != nil {
			log.Printf("  project %s: %v", p.name, err)
		} else {
			fmt.Printf("  + Project: %s [%s/%s]\n", p.name, p.status, p.priority)
		}
	}

	// --- Project Members ---
	for i, pid := range projectIDs {
		// Assign the user and a team member to each project
		_, err := pool.Exec(ctx, `
			INSERT INTO project_members (id, project_id, user_id, role, assigned_by)
			VALUES ($1, $2, $3, 'owner', $4)
			ON CONFLICT (id) DO NOTHING`,
			projectMemberIDs[i], pid, userID, userID,
		)
		if err != nil {
			log.Printf("  project_member for %s: %v", pid, err)
		}
	}
	fmt.Printf("  + %d project members (owner)\n", len(projectIDs))
}
