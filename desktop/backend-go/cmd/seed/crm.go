package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fixed CRM seed IDs
var pipelineID = uuid.MustParse("00000000-5eed-4000-a000-000000000701")

var stageIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000711"), // Lead
	uuid.MustParse("00000000-5eed-4000-a000-000000000712"), // Qualified
	uuid.MustParse("00000000-5eed-4000-a000-000000000713"), // Proposal
	uuid.MustParse("00000000-5eed-4000-a000-000000000714"), // Negotiation
	uuid.MustParse("00000000-5eed-4000-a000-000000000715"), // Closed Won
}

var companyIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000731"), // Quantum Dynamics
	uuid.MustParse("00000000-5eed-4000-a000-000000000732"), // Solaris Energy
	uuid.MustParse("00000000-5eed-4000-a000-000000000733"), // Vanguard Technologies
	uuid.MustParse("00000000-5eed-4000-a000-000000000734"), // Pinnacle Retail
	uuid.MustParse("00000000-5eed-4000-a000-000000000735"), // Oasis Hospitality
}

var crmDealIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000721"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000722"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000723"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000724"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000725"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000726"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000727"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000728"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000729"),
	uuid.MustParse("00000000-5eed-4000-a000-000000000730"),
}

// Maps deal index to company index for activity insertion
var dealToCompany = []int{
	0, // deal 0 -> Quantum Dynamics
	1, // deal 1 -> Solaris Energy
	3, // deal 2 -> Pinnacle Retail
	2, // deal 3 -> Vanguard Technologies
	4, // deal 4 -> Oasis Hospitality
	0, // deal 5 -> Quantum Dynamics
	1, // deal 6 -> Solaris Energy
	2, // deal 7 -> Vanguard Technologies
	4, // deal 8 -> Oasis Hospitality
	3, // deal 9 -> Pinnacle Retail
}

var crmActivityIDs []uuid.UUID

func init() {
	for i := 1; i <= 40; i++ {
		crmActivityIDs = append(crmActivityIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 740+i)))
	}
}

func seedCRM(ctx context.Context, pool *pgxpool.Pool, userID string) {
	// --- Pipeline ---
	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'Sales Pipeline', 'Primary sales pipeline for tracking deals from lead to close', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT (id) DO NOTHING`,
		pipelineID, userID,
	)
	if err != nil {
		log.Printf("  pipeline: %v", err)
	} else {
		fmt.Println("  + Pipeline: Sales Pipeline")
	}

	// --- Stages ---
	type stage struct {
		id        uuid.UUID
		name      string
		pos       int
		prob      int
		stageType string
		rotting   int
		color     string
	}

	stages := []stage{
		{stageIDs[0], "Lead", 0, 10, "open", 14, "#94a3b8"},
		{stageIDs[1], "Qualified", 1, 30, "open", 10, "#3b82f6"},
		{stageIDs[2], "Proposal", 2, 60, "open", 7, "#f59e0b"},
		{stageIDs[3], "Negotiation", 3, 80, "open", 5, "#8b5cf6"},
		{stageIDs[4], "Closed Won", 4, 100, "won", 0, "#22c55e"},
	}

	for _, s := range stages {
		_, err := pool.Exec(ctx, `
			INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING`,
			s.id, pipelineID, s.name, s.pos, s.prob, s.stageType, s.rotting, s.color,
		)
		if err != nil {
			log.Printf("  stage %s: %v", s.name, err)
		}
	}
	fmt.Println("  + 5 pipeline stages")

	// --- Companies ---
	type company struct {
		id         uuid.UUID
		name       string
		industry   string
		size       string
		website    string
		email      string
		phone      string
		city       string
		country    string
		revenue    float64
		lifecycle  string
		health     int
		engagement int
	}

	companies := []company{
		{companyIDs[0], "Quantum Dynamics Inc.", "Technology", "51-200", "https://quantumdynamics.tech", "info@quantumdynamics.tech", "+1-650-555-0001", "Palo Alto", "USA", 12000000, "customer", 85, 72},
		{companyIDs[1], "Solaris Energy Corp.", "Energy", "201-500", "https://solarisenergy.com", "biz@solarisenergy.com", "+1-303-555-0002", "Denver", "USA", 48000000, "prospect", 60, 45},
		{companyIDs[2], "Vanguard Technologies", "Software", "11-50", "https://vanguardtech.io", "hello@vanguardtech.io", "+1-206-555-0003", "Seattle", "USA", 5200000, "customer", 92, 88},
		{companyIDs[3], "Pinnacle Retail Group", "Retail", "501-1000", "https://pinnacleretail.com", "partnerships@pinnacleretail.com", "+1-312-555-0004", "Chicago", "USA", 85000000, "prospect", 40, 30},
		{companyIDs[4], "Oasis Hospitality", "Hospitality", "201-500", "https://oasishospitality.com", "tech@oasishospitality.com", "+1-702-555-0005", "Las Vegas", "USA", 32000000, "partner", 75, 65},
	}

	for _, c := range companies {
		_, err := pool.Exec(ctx, `
			INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO NOTHING`,
			c.id, userID, c.name, c.industry, c.size, c.website, c.email, c.phone,
			c.city, c.country, c.revenue, c.lifecycle, c.health, c.engagement,
		)
		if err != nil {
			log.Printf("  company %s: %v", c.name, err)
		}
	}
	fmt.Println("  + 5 CRM companies")

	// --- CRM Deals (10, spread across stages) ---
	type deal struct {
		id         uuid.UUID
		name       string
		stageIdx   int
		companyIdx int
		amount     float64
		prob       int
		status     string
		priority   string
		source     string
		closeDays  int // positive = future, negative = past
		desc       string
	}

	deals := []deal{
		// Lead stage (3)
		{crmDealIDs[0], "Quantum AI Platform License", 0, 0, 45000, 10, "open", "medium", "Website", 60, "Enterprise license for AI analytics platform"},
		{crmDealIDs[1], "Solaris Dashboard Prototype", 0, 1, 18000, 10, "open", "low", "Conference", 90, "Proof-of-concept energy monitoring dashboard"},
		{crmDealIDs[2], "Pinnacle Inventory System", 0, 3, 120000, 15, "open", "high", "Cold Outreach", 75, "Enterprise inventory management overhaul"},
		// Qualified stage (2)
		{crmDealIDs[3], "Vanguard API Integration Suite", 1, 2, 32000, 35, "open", "high", "Referral", 45, "Custom API integrations for their SaaS platform"},
		{crmDealIDs[4], "Oasis Booking Platform", 1, 4, 78000, 30, "open", "medium", "Website", 60, "Online booking and reservation management system"},
		// Proposal stage (2)
		{crmDealIDs[5], "Quantum Data Migration", 2, 0, 56000, 60, "open", "high", "Existing Client", 30, "Migrate legacy data warehouse to modern stack"},
		{crmDealIDs[6], "Solaris Monitoring Suite", 2, 1, 95000, 55, "open", "urgent", "Trade Show", 45, "Full monitoring and alerting platform for energy grid"},
		// Negotiation stage (2)
		{crmDealIDs[7], "Vanguard Enterprise Plan", 3, 2, 85000, 80, "open", "high", "Upsell", 14, "Upgrading from startup to enterprise tier"},
		{crmDealIDs[8], "Oasis Mobile App", 3, 4, 62000, 75, "open", "medium", "Existing Client", 21, "Guest-facing mobile app for hotel chain"},
		// Closed Won (1)
		{crmDealIDs[9], "Pinnacle POS Integration", 4, 3, 42000, 100, "won", "medium", "Referral", -10, "Point-of-sale system integration completed and delivered"},
	}

	for _, d := range deals {
		var closeExpr string
		if d.closeDays >= 0 {
			closeExpr = fmt.Sprintf("CURRENT_DATE + INTERVAL '%d days'", d.closeDays)
		} else {
			closeExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", -d.closeDays)
		}

		q := fmt.Sprintf(`
			INSERT INTO deals (id, user_id, pipeline_id, stage_id, name, description, amount, currency, probability,
				expected_close_date, owner_id, company_id, status, priority, lead_source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'USD', $8, %s, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO NOTHING`, closeExpr)

		_, err := pool.Exec(ctx, q,
			d.id, userID, pipelineID, stageIDs[d.stageIdx], d.name, d.desc, d.amount, d.prob,
			userID, companyIDs[d.companyIdx], d.status, d.priority, d.source,
		)
		if err != nil {
			log.Printf("  deal %s: %v", d.name, err)
		}
	}
	fmt.Println("  + 10 CRM deals")

	// --- CRM Activities (3-5 per deal) ---
	type activity struct {
		id       uuid.UUID
		dealIdx  int
		aType    string
		subject  string
		desc     string
		outcome  string
		daysAgo  int
		duration int
		done     bool
	}

	activities := []activity{
		// Deal 0: Quantum AI Platform License
		{crmActivityIDs[0], 0, "email", "Intro Email - AI Platform", "Sent initial product info and pricing sheet", "Positive response, requested demo", 20, 0, true},
		{crmActivityIDs[1], 0, "call", "Discovery Call with Quantum CTO", "Discussed AI platform requirements and current tech stack", "Good fit, needs board approval for budget", 14, 30, true},
		{crmActivityIDs[2], 0, "follow_up", "Budget Approval Follow-up", "Check on budget approval status", "", 3, 0, false},
		// Deal 1: Solaris Dashboard Prototype
		{crmActivityIDs[3], 1, "meeting", "Conference Booth Chat", "Met at CleanTech 2026 conference booth", "Exchanged cards, interested in dashboards", 25, 15, true},
		{crmActivityIDs[4], 1, "email", "Post-Conference Follow-up", "Sent company overview and relevant case studies", "No response yet", 20, 0, true},
		{crmActivityIDs[5], 1, "call", "Follow-up Call", "Called to follow up on email", "Reached voicemail, left message", 15, 5, true},
		// Deal 2: Pinnacle Inventory System
		{crmActivityIDs[6], 2, "call", "Cold Call - Pinnacle IT Dept", "Reached IT procurement team lead", "Interested, requested capabilities document", 18, 20, true},
		{crmActivityIDs[7], 2, "email", "Capabilities Document Sent", "Sent detailed capabilities and case studies", "Document forwarded to VP Supply Chain", 15, 0, true},
		{crmActivityIDs[8], 2, "meeting", "Initial Requirements Gathering", "Video call with IT and Supply Chain teams", "Documented 12 key requirements", 8, 60, true},
		{crmActivityIDs[9], 2, "proposal_sent", "RFP Response Submitted", "Submitted formal RFP response", "", 5, 0, true},
		// Deal 3: Vanguard API Integration
		{crmActivityIDs[10], 3, "call", "Referral Introduction", "Call with mutual connection at Vanguard", "Strong interest, scheduling technical deep-dive", 12, 25, true},
		{crmActivityIDs[11], 3, "demo", "Technical Demo - API Suite", "Demonstrated integration capabilities with live environment", "Very impressed, requesting proposal", 7, 45, true},
		{crmActivityIDs[12], 3, "email", "Proposal Draft Sent", "Sent preliminary proposal for review", "Minor revisions requested", 4, 0, true},
		// Deal 4: Oasis Booking Platform
		{crmActivityIDs[13], 4, "email", "Inbound Lead - Website Form", "Oasis submitted inquiry through website contact form", "Auto-responded, assigned to sales", 30, 0, true},
		{crmActivityIDs[14], 4, "call", "Qualification Call", "Discussed booking platform needs and timeline", "Budget available, timeline is Q2", 25, 35, true},
		{crmActivityIDs[15], 4, "meeting", "Requirements Workshop", "Half-day workshop with hotel operations team", "Detailed requirements document produced", 18, 240, true},
		{crmActivityIDs[16], 4, "follow_up", "Timeline Confirmation", "Confirming Q2 start date still holds", "", 5, 0, false},
		// Deal 5: Quantum Data Migration
		{crmActivityIDs[17], 5, "meeting", "Data Audit Session", "Reviewed current data warehouse architecture", "Identified 3 legacy systems to migrate", 10, 90, true},
		{crmActivityIDs[18], 5, "proposal_sent", "Migration Proposal", "Sent phased migration proposal with timeline", "Under review by engineering team", 6, 0, true},
		{crmActivityIDs[19], 5, "call", "Proposal Review Call", "Discussed proposal details and clarifications", "Requested minor scope adjustment", 3, 40, true},
		// Deal 6: Solaris Monitoring Suite
		{crmActivityIDs[20], 6, "demo", "Platform Demo to Engineering", "Full demo to Solaris engineering team", "Team very enthusiastic", 12, 60, true},
		{crmActivityIDs[21], 6, "meeting", "Executive Presentation", "Presented ROI case to Solaris C-suite", "CFO wants to see cost breakdown", 8, 45, true},
		{crmActivityIDs[22], 6, "proposal_sent", "Detailed Cost Proposal", "Sent itemized cost breakdown with 3 tier options", "Evaluating Tier 2 option", 5, 0, true},
		{crmActivityIDs[23], 6, "call", "Negotiation - Pricing Discussion", "Discussed pricing flexibility on Tier 2", "Agreed to 8% volume discount", 2, 30, true},
		// Deal 7: Vanguard Enterprise Plan
		{crmActivityIDs[24], 7, "meeting", "Enterprise Features Review", "Walked through enterprise tier features and SLAs", "Ready to proceed pending legal review", 10, 60, true},
		{crmActivityIDs[25], 7, "email", "Contract Draft Sent", "Sent enterprise agreement for legal review", "Legal reviewing, expect response in 5 days", 7, 0, true},
		{crmActivityIDs[26], 7, "call", "Contract Negotiation", "Discussed contract terms with legal and procurement", "Two minor clause changes requested", 3, 45, true},
		{crmActivityIDs[27], 7, "follow_up", "Awaiting Final Signature", "Contract at CEO desk for signature", "", 1, 0, false},
		// Deal 8: Oasis Mobile App
		{crmActivityIDs[28], 8, "demo", "Mobile App Prototype Demo", "Showed interactive prototype to hotel GM team", "Very positive feedback, minor UX tweaks requested", 15, 50, true},
		{crmActivityIDs[29], 8, "meeting", "UX Revision Review", "Reviewed revised prototype with changes", "Approved, moving to contract stage", 8, 30, true},
		{crmActivityIDs[30], 8, "contract_sent", "Service Agreement Sent", "Sent mobile app development agreement", "Under legal review", 5, 0, true},
		// Deal 9: Pinnacle POS Integration (Closed Won)
		{crmActivityIDs[31], 9, "call", "Initial Scoping Call", "Discussed POS integration requirements", "Clear scope defined", 60, 30, true},
		{crmActivityIDs[32], 9, "proposal_sent", "Integration Proposal", "Sent technical proposal and timeline", "Accepted within 48 hours", 55, 0, true},
		{crmActivityIDs[33], 9, "meeting", "Project Kickoff", "Kicked off integration project with dev team", "Sprint plan established", 45, 60, true},
		{crmActivityIDs[34], 9, "meeting", "Final Delivery & Handoff", "Delivered completed integration with documentation", "Client signed off, project complete", 10, 90, true},
	}

	for _, a := range activities {
		companyIdx := dealToCompany[a.dealIdx]

		completedExpr := "NULL"
		if a.done {
			completedExpr = fmt.Sprintf("NOW() - INTERVAL '%d days'", a.daysAgo)
		}

		q := fmt.Sprintf(`
			INSERT INTO crm_activities (id, user_id, activity_type, subject, description, outcome,
				deal_id, company_id, activity_date, duration_minutes, is_completed, completed_at, owner_id)
			VALUES ($1, $2, $3::crm_activity_type, $4, $5, $6, $7, $8,
				NOW() - INTERVAL '%d days', $9, $10, %s, $11)
			ON CONFLICT (id) DO NOTHING`,
			a.daysAgo, completedExpr,
		)

		_, err := pool.Exec(ctx, q,
			a.id, userID, a.aType, a.subject, a.desc, a.outcome,
			crmDealIDs[a.dealIdx], companyIDs[companyIdx],
			a.duration, a.done, userID,
		)
		if err != nil {
			log.Printf("  activity %s: %v", a.subject, err)
		}
	}
	fmt.Printf("  + %d CRM activities\n", len(activities))
}
