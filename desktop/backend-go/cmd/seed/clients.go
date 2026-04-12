package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fixed client seed IDs
var clientIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000201"), // Nexus Digital Agency
	uuid.MustParse("00000000-5eed-4000-a000-000000000202"), // Apex Manufacturing
	uuid.MustParse("00000000-5eed-4000-a000-000000000203"), // Sarah Chen Consulting
	uuid.MustParse("00000000-5eed-4000-a000-000000000204"), // Meridian Healthcare
	uuid.MustParse("00000000-5eed-4000-a000-000000000205"), // Brightwave Studios
	uuid.MustParse("00000000-5eed-4000-a000-000000000206"), // Terraform Real Estate
	uuid.MustParse("00000000-5eed-4000-a000-000000000207"), // CloudNine SaaS
	uuid.MustParse("00000000-5eed-4000-a000-000000000208"), // GreenRoot Organics
	uuid.MustParse("00000000-5eed-4000-a000-000000000209"), // Pacific Ventures Capital
	uuid.MustParse("00000000-5eed-4000-a000-000000000210"), // Atlas Logistics Group
}

// Contact seed IDs (3 per client = 30 max, we use ~25)
var contactIDs []uuid.UUID

func init() {
	for i := 1; i <= 30; i++ {
		contactIDs = append(contactIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 300+i)))
	}
}

// Interaction seed IDs
var interactionIDs []uuid.UUID

func init() {
	for i := 1; i <= 30; i++ {
		interactionIDs = append(interactionIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 400+i)))
	}
}

// Client deal seed IDs
var clientDealIDs []uuid.UUID

func init() {
	for i := 1; i <= 20; i++ {
		clientDealIDs = append(clientDealIDs, uuid.MustParse(fmt.Sprintf("00000000-5eed-4000-a000-000000000%03d", 450+i)))
	}
}

func seedClients(ctx context.Context, pool *pgxpool.Pool, userID string) {
	type client struct {
		id       uuid.UUID
		name     string
		cType    string
		email    string
		phone    string
		website  *string
		industry string
		size     string
		status   string
		source   string
		ltv      *float64
		notes    string
		city     string
		country  string
		lastDays *int // days ago for last_contacted_at, nil = never
	}

	ws := func(s string) *string { return &s }
	wf := func(f float64) *float64 { return &f }
	wi := func(i int) *int { return &i }

	clients := []client{
		// 4 active
		{clientIDs[0], "Nexus Digital Agency", "company", "hello@nexusdigital.io", "+1-415-555-0101", ws("https://nexusdigital.io"), "Technology", "11-50", "active", "Referral", wf(92000), "Full-service digital agency. Expanding retainer to include mobile.", "San Francisco", "USA", wi(1)},
		{clientIDs[1], "Apex Manufacturing Co.", "company", "contact@apexmfg.com", "+1-313-555-0202", ws("https://apexmfg.com"), "Manufacturing", "201-500", "active", "Trade Show", wf(145000), "ERP modernization project underway. Key enterprise account.", "Detroit", "USA", wi(3)},
		{clientIDs[2], "Sarah Chen Consulting", "individual", "sarah.c@consultingpro.com", "+1-512-555-0303", nil, "Consulting", "1-10", "active", "LinkedIn", wf(28000), "Independent strategy consultant. Repeat client, low-touch.", "Austin", "USA", wi(0)},
		{clientIDs[3], "Meridian Healthcare Group", "company", "partnerships@meridianhc.org", "+1-617-555-0404", ws("https://meridianhc.org"), "Healthcare", "501-1000", "active", "Conference", wf(210000), "HIPAA-compliant patient portal. Multi-year contract.", "Boston", "USA", wi(2)},
		// 2 prospect
		{clientIDs[4], "Brightwave Studios", "company", "biz@brightwavemedia.com", "+1-323-555-0505", ws("https://brightwavemedia.com"), "Entertainment", "11-50", "prospect", "Website", nil, "Animation studio interested in project management tools. Demo scheduled.", "Los Angeles", "USA", wi(5)},
		{clientIDs[5], "Terraform Real Estate", "company", "dev@terraformre.com", "+1-212-555-0606", ws("https://terraformre.com"), "Real Estate", "51-200", "prospect", "Cold Outreach", nil, "Commercial real estate firm. Needs tenant management portal.", "New York", "USA", wi(8)},
		// 2 lead
		{clientIDs[6], "CloudNine SaaS", "company", "info@cloudninesaas.com", "+1-206-555-0707", ws("https://cloudninesaas.com"), "Technology", "11-50", "lead", "Webinar", nil, "B2B SaaS startup. Attended our API integration webinar.", "Seattle", "USA", nil},
		{clientIDs[7], "GreenRoot Organics", "company", "team@greenrootorganics.com", "+1-503-555-0808", nil, "Food & Beverage", "11-50", "lead", "Referral", nil, "Organic food distributor looking for inventory tracking.", "Portland", "USA", wi(12)},
		// 1 inactive
		{clientIDs[8], "Pacific Ventures Capital", "company", "ops@pacificvc.com", "+1-415-555-0909", ws("https://pacificvc.com"), "Finance", "11-50", "inactive", "LinkedIn", wf(55000), "Portfolio dashboard delivered. Re-engage for phase 2 in Q3.", "San Francisco", "USA", wi(45)},
		// 1 churned
		{clientIDs[9], "Atlas Logistics Group", "company", "tech@atlaslogistics.net", "+1-972-555-1010", nil, "Logistics", "201-500", "churned", "Trade Show", wf(32000), "Went with in-house solution. Price-sensitive. May revisit.", "Dallas", "USA", wi(90)},
	}

	for _, c := range clients {
		var lastContact *string
		if c.lastDays != nil {
			lc := fmt.Sprintf("NOW() - INTERVAL '%d days'", *c.lastDays)
			lastContact = &lc
		}

		q := `INSERT INTO clients (id, user_id, name, type, email, phone, website, industry, company_size, status, source, lifetime_value, notes, city, country, last_contacted_at)
			VALUES ($1, $2, $3, $4::clienttype, $5, $6, $7, $8, $9, $10::clientstatus, $11, $12, $13, $14, $15, `
		if lastContact != nil {
			q += fmt.Sprintf("NOW() - INTERVAL '%d days'", *c.lastDays)
		} else {
			q += "NULL"
		}
		q += `) ON CONFLICT (id) DO NOTHING`

		_, err := pool.Exec(ctx, q,
			c.id, userID, c.name, c.cType, c.email, c.phone, c.website,
			c.industry, c.size, c.status, c.source, c.ltv, c.notes, c.city, c.country,
		)
		if err != nil {
			log.Printf("  client %s: %v", c.name, err)
		} else {
			fmt.Printf("  + Client: %s [%s]\n", c.name, c.status)
		}
	}

	// --- Contacts (2-3 per client) ---
	type contact struct {
		id        uuid.UUID
		clientIdx int
		name      string
		email     string
		phone     string
		role      string
		isPrimary bool
	}

	contacts := []contact{
		// Nexus Digital Agency (3)
		{contactIDs[0], 0, "Jordan Rivera", "jordan@nexusdigital.io", "+1-415-555-0111", "CEO", true},
		{contactIDs[1], 0, "Priya Patel", "priya@nexusdigital.io", "+1-415-555-0112", "CTO", false},
		{contactIDs[2], 0, "Marcus Webb", "marcus@nexusdigital.io", "+1-415-555-0113", "Account Manager", false},
		// Apex Manufacturing (3)
		{contactIDs[3], 1, "Robert Hayes", "r.hayes@apexmfg.com", "+1-313-555-0211", "VP Operations", true},
		{contactIDs[4], 1, "Diana Cheng", "d.cheng@apexmfg.com", "+1-313-555-0212", "IT Director", false},
		{contactIDs[5], 1, "Tom Kowalski", "t.kowalski@apexmfg.com", "+1-313-555-0213", "Plant Manager", false},
		// Sarah Chen Consulting (2)
		{contactIDs[6], 2, "Sarah Chen", "sarah.c@consultingpro.com", "+1-512-555-0303", "Principal", true},
		{contactIDs[7], 2, "James Ortiz", "j.ortiz@consultingpro.com", "+1-512-555-0311", "Research Associate", false},
		// Meridian Healthcare (3)
		{contactIDs[8], 3, "Dr. Amara Okafor", "a.okafor@meridianhc.org", "+1-617-555-0411", "Chief Medical Officer", true},
		{contactIDs[9], 3, "Linda Torres", "l.torres@meridianhc.org", "+1-617-555-0412", "Head of IT", false},
		{contactIDs[10], 3, "Kevin Park", "k.park@meridianhc.org", "+1-617-555-0413", "Project Lead", false},
		// Brightwave Studios (2)
		{contactIDs[11], 4, "Zoe Nakamura", "zoe@brightwavemedia.com", "+1-323-555-0511", "Creative Director", true},
		{contactIDs[12], 4, "Ethan Brooks", "ethan@brightwavemedia.com", "+1-323-555-0512", "Production Manager", false},
		// Terraform Real Estate (2)
		{contactIDs[13], 5, "Victoria Grant", "v.grant@terraformre.com", "+1-212-555-0611", "Managing Director", true},
		{contactIDs[14], 5, "David Kim", "d.kim@terraformre.com", "+1-212-555-0612", "Head of Technology", false},
		// CloudNine SaaS (2)
		{contactIDs[15], 6, "Aiden Foster", "aiden@cloudninesaas.com", "+1-206-555-0711", "Founder & CEO", true},
		{contactIDs[16], 6, "Mia Chen", "mia@cloudninesaas.com", "+1-206-555-0712", "Head of Product", false},
		// GreenRoot Organics (2)
		{contactIDs[17], 7, "Olivia Sanchez", "olivia@greenrootorganics.com", "+1-503-555-0811", "Operations Director", true},
		{contactIDs[18], 7, "Raj Gupta", "raj@greenrootorganics.com", "+1-503-555-0812", "Warehouse Manager", false},
		// Pacific Ventures (2)
		{contactIDs[19], 8, "Catherine Liu", "c.liu@pacificvc.com", "+1-415-555-0911", "Managing Partner", true},
		{contactIDs[20], 8, "Steven Adler", "s.adler@pacificvc.com", "+1-415-555-0912", "Analyst", false},
		// Atlas Logistics (2)
		{contactIDs[21], 9, "Frank Morrison", "f.morrison@atlaslogistics.net", "+1-972-555-1011", "COO", true},
		{contactIDs[22], 9, "Nina Vasquez", "n.vasquez@atlaslogistics.net", "+1-972-555-1012", "Logistics Coordinator", false},
	}

	for _, c := range contacts {
		_, err := pool.Exec(ctx, `
			INSERT INTO client_contacts (id, client_id, name, email, phone, role, is_primary)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING`,
			c.id, clientIDs[c.clientIdx], c.name, c.email, c.phone, c.role, c.isPrimary,
		)
		if err != nil {
			log.Printf("  contact %s: %v", c.name, err)
		}
	}
	fmt.Printf("  + %d contacts\n", len(contacts))

	// --- Interactions (2-3 per client) ---
	type interaction struct {
		id        uuid.UUID
		clientIdx int
		contactID uuid.UUID
		iType     string
		subject   string
		desc      string
		outcome   string
		daysAgo   int
	}

	interactions := []interaction{
		// Nexus Digital
		{interactionIDs[0], 0, contactIDs[0], "meeting", "Q2 Retainer Expansion Discussion", "Met with Jordan to discuss expanding our retainer to include mobile app development.", "Agreed to draft proposal for mobile addition", 3},
		{interactionIDs[1], 0, contactIDs[1], "email", "Technical Architecture Review", "Sent updated architecture docs for the new microservices migration.", "Priya approved the approach, moving to implementation", 7},
		{interactionIDs[2], 0, contactIDs[2], "call", "Monthly Check-in", "Regular monthly check-in call with account manager.", "All deliverables on track. Invoice paid.", 14},
		// Apex Manufacturing
		{interactionIDs[3], 1, contactIDs[3], "meeting", "ERP Modernization Kickoff", "On-site kickoff meeting for the ERP modernization project.", "Signed SOW. Development starting next Monday.", 5},
		{interactionIDs[4], 1, contactIDs[4], "call", "Data Migration Planning", "Discussed legacy data migration strategy with IT director.", "Need to schedule follow-up for schema mapping", 10},
		// Sarah Chen
		{interactionIDs[5], 2, contactIDs[6], "email", "Quarterly Strategy Report Delivery", "Delivered Q1 strategy report and recommendations.", "Report accepted. Follow-up retainer confirmed.", 2},
		{interactionIDs[6], 2, contactIDs[6], "call", "New Project Scoping", "Scoped a new market analysis project for Q2.", "Proposal to follow within 48 hours", 8},
		// Meridian Healthcare
		{interactionIDs[7], 3, contactIDs[8], "meeting", "HIPAA Compliance Review", "Reviewed portal security with CMO and compliance team.", "All security controls approved. Go-live date set.", 4},
		{interactionIDs[8], 3, contactIDs[9], "call", "API Integration Troubleshooting", "Debugged HL7 FHIR integration issues with IT head.", "Identified root cause, fix deployed same day", 6},
		{interactionIDs[9], 3, contactIDs[10], "email", "Sprint Demo Invitation", "Sent invite for next sprint demo to project stakeholders.", "All attendees confirmed", 1},
		// Brightwave Studios
		{interactionIDs[10], 4, contactIDs[11], "meeting", "Product Demo", "Demonstrated project management platform capabilities.", "Very interested. Requesting pricing proposal.", 7},
		{interactionIDs[11], 4, contactIDs[12], "email", "Follow-up: Feature Comparison", "Sent feature comparison document after demo.", "Waiting for internal review", 5},
		// Terraform Real Estate
		{interactionIDs[12], 5, contactIDs[13], "call", "Discovery Call", "Initial discovery call about tenant management needs.", "Strong fit. Scheduling on-site visit.", 10},
		{interactionIDs[13], 5, contactIDs[14], "email", "Technical Requirements Doc", "Sent technical requirements questionnaire.", "Response received, reviewing", 8},
		// CloudNine SaaS
		{interactionIDs[14], 6, contactIDs[15], "email", "Post-Webinar Follow-up", "Followed up after API integration webinar attendance.", "Interested in custom integration. Needs budget approval.", 15},
		// GreenRoot Organics
		{interactionIDs[15], 7, contactIDs[17], "call", "Referral Introduction Call", "Intro call from mutual connection.", "Good conversation. Sending capabilities deck.", 12},
		// Pacific Ventures
		{interactionIDs[16], 8, contactIDs[19], "meeting", "Phase 1 Retrospective", "Reviewed delivered portfolio dashboard results.", "Client satisfied. Phase 2 discussion deferred to Q3.", 45},
		{interactionIDs[17], 8, contactIDs[20], "email", "Feature Request Documentation", "Documented requested enhancements for phase 2.", "Added to backlog for Q3 planning", 40},
		// Atlas Logistics
		{interactionIDs[18], 9, contactIDs[21], "call", "Exit Interview", "Final call to understand why they chose in-house.", "Price and control were main factors. Left door open.", 90},
	}

	for _, i := range interactions {
		_, err := pool.Exec(ctx, fmt.Sprintf(`
			INSERT INTO client_interactions (id, client_id, contact_id, type, subject, description, outcome, occurred_at)
			VALUES ($1, $2, $3, $4::interactiontype, $5, $6, $7, NOW() - INTERVAL '%d days')
			ON CONFLICT (id) DO NOTHING`, i.daysAgo),
			i.id, clientIDs[i.clientIdx], i.contactID, i.iType, i.subject, i.desc, i.outcome,
		)
		if err != nil {
			log.Printf("  interaction %s: %v", i.subject, err)
		}
	}
	fmt.Printf("  + %d interactions\n", len(interactions))
}

// seedClientDeals inserts client-linked deals into the `deals` table.
// Must be called AFTER seedCRM() so pipelineID and stageIDs exist.
func seedClientDeals(ctx context.Context, pool *pgxpool.Pool, userID string) {
	// stageIDs: 0=Lead, 1=Qualified, 2=Proposal, 3=Negotiation, 4=Closed Won
	type clientDeal struct {
		id        uuid.UUID
		clientIdx int
		name      string
		value     float64
		prob      int
		stageIdx  int  // index into stageIDs
		closeDays *int // positive = future, negative = past
		notes     string
		status    string
	}

	cd := func(d int) *int { return &d }

	clientDeals := []clientDeal{
		// Nexus Digital
		{clientDealIDs[0], 0, "Mobile App Development Retainer", 48000, 60, 2, cd(30), "Monthly retainer for ongoing mobile development work", "open"},
		{clientDealIDs[1], 0, "Website Redesign Phase 2", 35000, 75, 3, cd(14), "Continuation of initial website project", "open"},
		// Apex Manufacturing
		{clientDealIDs[2], 1, "ERP Modernization", 145000, 100, 4, cd(-5), "Signed. 18-month project timeline.", "won"},
		{clientDealIDs[3], 1, "IoT Dashboard Add-on", 32000, 25, 1, cd(60), "Exploring real-time factory monitoring", "open"},
		// Sarah Chen
		{clientDealIDs[4], 2, "Q2 Strategy Engagement", 18000, 80, 2, cd(21), "Quarterly consulting retainer renewal", "open"},
		// Meridian Healthcare
		{clientDealIDs[5], 3, "Patient Portal Phase 2", 95000, 70, 3, cd(45), "Adding telehealth integration", "open"},
		{clientDealIDs[6], 3, "Staff Scheduling Module", 42000, 30, 1, cd(90), "Early discussions about workforce management", "open"},
		// Brightwave Studios
		{clientDealIDs[7], 4, "Project Management Platform", 28000, 50, 2, cd(30), "Custom PM tool for creative workflows", "open"},
		// Terraform Real Estate
		{clientDealIDs[8], 5, "Tenant Management Portal", 67000, 20, 1, cd(60), "Comprehensive property management solution", "open"},
		// CloudNine SaaS
		{clientDealIDs[9], 6, "API Integration Package", 22000, 15, 1, cd(45), "Custom API integration development", "open"},
		// GreenRoot Organics
		{clientDealIDs[10], 7, "Inventory Tracking System", 38000, 10, 0, cd(90), "Warehouse-to-store tracking", "open"},
		// Pacific Ventures
		{clientDealIDs[11], 8, "Dashboard Phase 2", 55000, 100, 4, cd(-30), "Delivered and paid", "won"},
		// Atlas Logistics
		{clientDealIDs[12], 9, "Fleet Management App", 75000, 0, 0, cd(-60), "Client chose in-house development", "lost"},
	}

	for _, d := range clientDeals {
		var closeExpr string
		if d.closeDays != nil {
			if *d.closeDays >= 0 {
				closeExpr = fmt.Sprintf("CURRENT_DATE + INTERVAL '%d days'", *d.closeDays)
			} else {
				closeExpr = fmt.Sprintf("CURRENT_DATE - INTERVAL '%d days'", -*d.closeDays)
			}
		} else {
			closeExpr = "NULL"
		}

		q := fmt.Sprintf(`
			INSERT INTO deals (id, user_id, pipeline_id, stage_id, client_id, name, description, amount, probability, expected_close_date, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, %s, $10)
			ON CONFLICT (id) DO NOTHING`, closeExpr)

		_, err := pool.Exec(ctx, q,
			d.id, userID, pipelineID, stageIDs[d.stageIdx], clientIDs[d.clientIdx],
			d.name, d.notes, d.value, d.prob, d.status,
		)
		if err != nil {
			log.Printf("  client_deal %s: %v", d.name, err)
		}
	}
	fmt.Printf("  + %d client deals\n", len(clientDeals))
}
