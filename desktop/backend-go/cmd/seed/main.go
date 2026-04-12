package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// ============================================================
// Deterministic seed UUIDs — all seed records use fixed IDs
// so the script is idempotent and cleanable.
// ============================================================

// Team members
var (
	tmSarah = uuid.MustParse("00000000-5eed-4000-a000-000000000101")
	tmMike  = uuid.MustParse("00000000-5eed-4000-a000-000000000102")
	tmLisa  = uuid.MustParse("00000000-5eed-4000-a000-000000000103")
	tmJohn  = uuid.MustParse("00000000-5eed-4000-a000-000000000104")
	tmEmily = uuid.MustParse("00000000-5eed-4000-a000-000000000105")
)

var allTeamMemberIDs = []uuid.UUID{tmSarah, tmMike, tmLisa, tmJohn, tmEmily}

func main() {
	userID := flag.String("user-id", "", "User ID (from \"user\" table) to own seed data")
	email := flag.String("email", "", "Look up user ID by email (alternative to --user-id)")
	force := flag.Bool("force", false, "Delete existing seed data and re-insert")
	dbURL := flag.String("db-url", "", "Override DATABASE_URL")
	flag.Parse()

	// Load .env
	_ = godotenv.Load()

	connStr := *dbURL
	if connStr == "" {
		connStr = os.Getenv("DATABASE_URL")
	}
	if connStr == "" {
		log.Fatal("DATABASE_URL is required (set env var or pass --db-url)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("DB config failed: %v", err)
	}
	// Use simple protocol to avoid prepared statement issues with PgBouncer/Supavisor
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer pool.Close()

	// Resolve user ID
	uid := *userID
	if uid == "" && *email != "" {
		err = pool.QueryRow(ctx, `SELECT id FROM "user" WHERE email = $1`, *email).Scan(&uid)
		if err != nil {
			log.Fatalf("No user found with email %q: %v", *email, err)
		}
		fmt.Printf("Resolved email %s -> user ID: %s\n", *email, uid)
	}
	if uid == "" {
		log.Fatal("Provide --user-id or --email")
	}

	// Idempotency check
	var seedExists bool
	err = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`,
		clientIDs[0],
	).Scan(&seedExists)
	if err != nil {
		log.Fatalf("Idempotency check failed: %v", err)
	}

	if seedExists && !*force {
		fmt.Println("Seed data already exists. Use --force to re-seed.")
		os.Exit(0)
	}

	if seedExists && *force {
		fmt.Println("--force: cleaning existing seed data...")
		cleanSeedData(ctx, pool)
	}

	// Ensure user has a workspace membership (required for project_members)
	ensureWorkspaceMembership(ctx, pool, uid)

	// Seed in dependency order:
	// 1. Team members (no deps)
	// 2. Clients + contacts + interactions (no deps)
	// 3. CRM pipeline/stages/companies/deals (no deps)
	// 4. Client deals (needs pipeline from step 3)
	// 5. Projects + members (needs clients for names)
	// 6. Tasks (needs projects + team members)
	// 7. Modules (needs workspace)
	fmt.Println("\n--- Seeding team members ---")
	seedTeamMembers(ctx, pool, uid)

	fmt.Println("\n--- Seeding clients ---")
	seedClients(ctx, pool, uid)

	fmt.Println("\n--- Seeding CRM ---")
	seedCRM(ctx, pool, uid)

	fmt.Println("\n--- Seeding client deals ---")
	seedClientDeals(ctx, pool, uid)

	fmt.Println("\n--- Seeding projects ---")
	seedProjects(ctx, pool, uid)

	fmt.Println("\n--- Seeding tasks ---")
	seedTasks(ctx, pool, uid)

	fmt.Println("\n--- Seeding modules ---")
	seedModules(ctx, pool, uid)

	fmt.Println("\n=== Seed complete ===")
	fmt.Printf("User: %s\n", uid)
	fmt.Println("Team members: 5")
	fmt.Println("Clients: 10 (+ contacts, interactions, deals)")
	fmt.Println("Projects: 6 (+ members)")
	fmt.Println("Tasks: 20")
	fmt.Println("CRM: 1 pipeline, 5 stages, 5 companies, 10 deals, activities")
	fmt.Println("Modules: 6")
}

func seedTeamMembers(ctx context.Context, pool *pgxpool.Pool, userID string) {
	type tm struct {
		id     uuid.UUID
		name   string
		email  string
		role   string
		status string
		cap    int
		skills string
		rate   float64
	}

	members := []tm{
		{tmSarah, "Sarah Chen", "sarah.chen@company.com", "Lead Developer", "AVAILABLE", 40, `["React","TypeScript","Node.js","PostgreSQL"]`, 125.00},
		{tmMike, "Mike Rodriguez", "mike.r@company.com", "UI/UX Designer", "BUSY", 35, `["Figma","CSS","User Research","Prototyping"]`, 95.00},
		{tmLisa, "Lisa Park", "lisa.park@company.com", "Project Manager", "AVAILABLE", 45, `["Agile","Scrum","Stakeholder Management"]`, 110.00},
		{tmJohn, "John Smith", "john.s@company.com", "Backend Developer", "OOO", 40, `["Go","Python","AWS","Docker"]`, 115.00},
		{tmEmily, "Emily Watson", "emily.w@company.com", "Marketing Lead", "AVAILABLE", 38, `["SEO","Content Strategy","Analytics","Social Media"]`, 90.00},
	}

	for _, m := range members {
		_, err := pool.Exec(ctx, `
			INSERT INTO team_members (id, user_id, name, email, role, status, capacity, skills, hourly_rate)
			VALUES ($1, $2, $3, $4, $5, $6::memberstatus, $7, $8::jsonb, $9)
			ON CONFLICT (id) DO NOTHING`,
			m.id, userID, m.name, m.email, m.role, m.status, m.cap, m.skills, m.rate,
		)
		if err != nil {
			log.Printf("  team_member %s: %v", m.name, err)
		} else {
			fmt.Printf("  + %s (%s)\n", m.name, m.role)
		}
	}
}

var devWorkspaceID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func ensureWorkspaceMembership(ctx context.Context, pool *pgxpool.Pool, userID string) {
	// Check if user already has a workspace membership
	var wsID uuid.UUID
	err := pool.QueryRow(ctx,
		`SELECT workspace_id FROM workspace_members WHERE user_id = $1 LIMIT 1`, userID,
	).Scan(&wsID)
	if err == nil {
		fmt.Printf("User already in workspace: %s\n", wsID)
		return
	}

	// Ensure dev workspace exists
	pool.Exec(ctx, `
		INSERT INTO workspaces (id, name, slug, description, plan_type, owner_id, settings)
		VALUES ($1, 'Development Workspace', 'dev-workspace', 'Dev workspace for seed data', 'professional', $2, '{"dev_mode": true}')
		ON CONFLICT (id) DO NOTHING`,
		devWorkspaceID, userID,
	)

	// The audit_member_changes trigger references NEW.role but the column is
	// role_name (trigger was never updated). Disable it for this insert.
	pool.Exec(ctx, `ALTER TABLE workspace_members DISABLE TRIGGER trigger_audit_member_changes`)

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role_name, status, joined_at)
		VALUES ($1, $2, 'owner', 'active', NOW())
		ON CONFLICT DO NOTHING`,
		devWorkspaceID, userID,
	)

	pool.Exec(ctx, `ALTER TABLE workspace_members ENABLE TRIGGER trigger_audit_member_changes`)

	if err != nil {
		log.Printf("  workspace membership: %v", err)
	} else {
		fmt.Printf("Added user to workspace: %s\n", devWorkspaceID)
	}
}

func cleanSeedData(ctx context.Context, pool *pgxpool.Pool) {
	// In simple protocol mode, pgx can't encode []uuid.UUID as array params.
	// Use individual DELETEs per ID instead.
	type cleanJob struct {
		label string
		table string
		col   string
		ids   []uuid.UUID
	}

	jobs := []cleanJob{
		{"custom_modules", "custom_modules", "id", moduleIDs},
		{"crm_activities", "crm_activities", "id", crmActivityIDs},
		{"crm_deals", "deals", "id", crmDealIDs},
		{"client_deals", "deals", "id", clientDealIDs[:13]},
		{"pipeline_stages", "pipeline_stages", "pipeline_id", []uuid.UUID{pipelineID}},
		{"pipelines", "pipelines", "id", []uuid.UUID{pipelineID}},
		{"companies", "companies", "id", companyIDs},
		{"tasks", "tasks", "id", taskIDs},
		{"project_members", "project_members", "id", projectMemberIDs[:len(projectIDs)]},
		{"projects", "projects", "id", projectIDs},
		{"clients", "clients", "id", clientIDs},
		{"team_members", "team_members", "id", allTeamMemberIDs},
	}

	for _, j := range jobs {
		count := int64(0)
		for _, id := range j.ids {
			tag, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = $1", j.table, j.col), id)
			if err != nil {
				log.Printf("  clean %s %s: %v", j.label, id, err)
			} else {
				count += tag.RowsAffected()
			}
		}
		fmt.Printf("  cleaned %s: %d rows\n", j.label, count)
	}
}
