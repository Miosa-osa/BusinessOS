package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fixed task seed IDs
var taskIDs []uuid.UUID

func init() {
	for i := 1; i <= 20; i++ {
		taskIDs = append(taskIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 600+i)))
	}
}

func seedTasks(ctx context.Context, pool *pgxpool.Pool, userID string) {
	type task struct {
		id         uuid.UUID
		title      string
		desc       string
		status     string
		priority   string
		projectIdx *int       // index into projectIDs, nil = standalone
		assigneeID *uuid.UUID // nil = unassigned
		dueDays    int        // positive = future, negative = past (overdue)
	}

	pi := func(i int) *int { return &i }
	ai := func(id uuid.UUID) *uuid.UUID { return &id }

	tasks := []task{
		// ===== PROJECT-LINKED TASKS (12) =====

		// Website Redesign (project 0) — 3 tasks
		{taskIDs[0], "Finalize homepage wireframes", "Complete wireframe designs for homepage hero section, navigation, and footer. Get stakeholder sign-off before development.", "done", "high", pi(0), ai(tmMike), -5},
		{taskIDs[1], "Implement responsive navigation", "Build mobile-first navigation component with hamburger menu, dropdown submenus, and smooth scroll behavior.", "in_progress", "high", pi(0), ai(tmSarah), 2},
		{taskIDs[2], "SEO audit and meta tag implementation", "Run comprehensive SEO audit. Implement meta tags, Open Graph tags, structured data, and optimize Core Web Vitals scores.", "todo", "medium", pi(0), nil, 10},

		// Mobile App MVP (project 1) — 3 tasks
		{taskIDs[3], "Design app onboarding flow", "Create 4-screen onboarding sequence with illustrations explaining key features. Include skip option and progress indicators.", "done", "high", pi(1), ai(tmMike), -7},
		{taskIDs[4], "Build sensor data API endpoints", "Implement REST API for real-time sensor data ingestion, historical queries, and alert threshold configuration.", "in_progress", "critical", pi(1), ai(tmJohn), 7},
		{taskIDs[5], "Push notification system", "Integrate Firebase Cloud Messaging. Implement alert rules engine for sensor threshold breaches and shift change notifications.", "todo", "medium", pi(1), nil, 14},

		// Internal Knowledge Base (project 2) — 2 tasks
		{taskIDs[6], "Set up search indexing", "Implement full-text search with PostgreSQL tsvector. Support filtering by category, author, and date range.", "in_progress", "medium", pi(2), ai(tmSarah), 5},
		{taskIDs[7], "Create onboarding guide template", "Design reusable template for department onboarding guides. Include checklist component, timeline view, and resource links.", "todo", "low", pi(2), nil, 21},

		// CRM Integration (project 3) — 2 tasks
		{taskIDs[8], "Map FHIR data model to our schema", "Document field-by-field mapping between HL7 FHIR resources and our internal data model. Flag any lossy transformations.", "done", "high", pi(3), ai(tmJohn), -14},
		{taskIDs[9], "Build patient record sync service", "Implement bidirectional sync service with conflict resolution. Must handle partial failures and maintain audit log.", "cancelled", "high", pi(3), ai(tmSarah), 30},

		// Q1 Marketing Campaign (project 4) — 2 tasks
		{taskIDs[10], "Draft launch email sequence", "Write 5-email nurture sequence: teaser, announcement, features deep-dive, social proof, limited-time offer.", "done", "high", pi(4), ai(tmEmily), -20},
		{taskIDs[11], "Create social media content calendar", "Plan 30 days of content across LinkedIn, Twitter, and Instagram. Include copy, hashtags, and optimal posting times.", "done", "medium", pi(4), nil, -15},

		// ===== STANDALONE TASKS (8) =====

		// Overdue tasks (4)
		{taskIDs[12], "Follow up with BrightPath Education", "Send follow-up email after their webinar attendance. Include case study and offer 15-min intro call.", "todo", "medium", nil, nil, -3},
		{taskIDs[13], "Fix authentication timeout bug", "Users report being logged out after 15 minutes despite 'Remember me' being checked. Investigate session/JWT expiry logic.", "in_progress", "critical", nil, ai(tmSarah), -1},
		{taskIDs[14], "Submit quarterly tax documents", "Compile and submit Q4 2025 tax documents to accountant. Include updated P&L and balance sheet.", "todo", "high", nil, nil, -2},
		{taskIDs[15], "Renew SSL certificates", "SSL certs for api.example.com and dashboard.example.com expire this week. Renew via Let's Encrypt and verify HSTS headers.", "todo", "high", nil, ai(tmJohn), -4},

		// Due today (3)
		{taskIDs[16], "Prepare Quantum Analytics proposal", "Draft SOW and pricing for analytics platform engagement. Include timeline, deliverables, and payment milestones.", "in_progress", "high", nil, ai(tmLisa), 0},
		{taskIDs[17], "Review Q4 financial reports", "Analyze Q4 revenue, expenses, and margins for board meeting. Flag any significant variances from forecast.", "todo", "high", nil, nil, 0},
		{taskIDs[18], "Update team handbook with PTO policy", "Add new unlimited PTO policy details, approval workflow, and minimum notice requirements to the team handbook.", "todo", "low", nil, nil, 0},

		// Cancelled
		{taskIDs[19], "Evaluate Kubernetes migration", "Research feasibility of migrating from Cloud Run to GKE. Compare costs, complexity, and operational overhead.", "cancelled", "low", nil, nil, 30},
	}

	for _, t := range tasks {
		var projectID *uuid.UUID
		if t.projectIdx != nil {
			projectID = &projectIDs[*t.projectIdx]
		}

		var dueExpr string
		if t.dueDays == 0 {
			dueExpr = "CURRENT_DATE"
		} else if t.dueDays > 0 {
			dueExpr = fmt.Sprintf("CURRENT_DATE + INTERVAL '%d days'", t.dueDays)
		} else {
			dueExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", -t.dueDays)
		}

		completedExpr := "NULL"
		if t.status == "done" {
			completedExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", func() int {
				if -t.dueDays > 0 {
					return -t.dueDays - 1
				}
				return 1
			}())
		}

		q := fmt.Sprintf(`
			INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
			VALUES ($1, $2, $3, $4, $5::taskstatus, $6::taskpriority, %s, $7, $8, %s)
			ON CONFLICT (id) DO NOTHING`, dueExpr, completedExpr)

		_, err := pool.Exec(ctx, q,
			t.id, userID, t.title, t.desc, t.status, t.priority,
			projectID, t.assigneeID,
		)
		if err != nil {
			log.Printf("  task %s: %v", t.title, err)
		}
	}

	// Print summary
	counts := map[string]int{}
	for _, t := range tasks {
		counts[t.status]++
	}
	fmt.Printf("  + 20 tasks (todo:%d, in_progress:%d, done:%d, cancelled:%d)\n",
		counts["todo"], counts["in_progress"], counts["done"], counts["cancelled"])

	linked := 0
	for _, t := range tasks {
		if t.projectIdx != nil {
			linked++
		}
	}
	fmt.Printf("  + %d linked to projects, %d standalone\n", linked, 20-linked)

	overdue := 0
	for _, t := range tasks {
		if t.dueDays < 0 && t.status != "done" && t.status != "cancelled" {
			overdue++
		}
	}
	fmt.Printf("  + %d overdue, 3 due today\n", overdue)
}
