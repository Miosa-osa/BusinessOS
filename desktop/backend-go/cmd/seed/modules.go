package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fixed module seed IDs
var moduleIDs = []uuid.UUID{
	uuid.MustParse("00000000-5eed-4000-a000-000000000601"), // Slack Notifier
	uuid.MustParse("00000000-5eed-4000-a000-000000000602"), // Invoice Generator
	uuid.MustParse("00000000-5eed-4000-a000-000000000603"), // Time Tracker
	uuid.MustParse("00000000-5eed-4000-a000-000000000604"), // Email Campaign
	uuid.MustParse("00000000-5eed-4000-a000-000000000605"), // Data Exporter
	uuid.MustParse("00000000-5eed-4000-a000-000000000606"), // KPI Dashboard
}

func seedModules(ctx context.Context, pool *pgxpool.Pool, userID string) {
	type mod struct {
		id       uuid.UUID
		name     string
		slug     string
		desc     string
		category string
		version  string
		icon     string
		manifest string
		tags     string
		installs int
		stars    int
	}

	modules := []mod{
		{
			moduleIDs[0],
			"Slack Notifier",
			"slack-notifier",
			"Send automated notifications to Slack channels when events occur in your workspace. Supports custom templates and channel routing.",
			"communication",
			"1.2.0",
			"message-square",
			`{"name":"Slack Notifier","version":"1.2.0","description":"Automated Slack notifications","author":"BusinessOS","category":"communication","actions":[{"name":"send_notification","description":"Send a message to a Slack channel","type":"api","parameters":{"channel":"string","message":"string","mention_users":"boolean"},"returns":{"success":"boolean","message_id":"string"}}],"config_schema":{"webhook_url":{"type":"string","required":true},"default_channel":{"type":"string","default":"general"}},"permissions":["webhooks:write"]}`,
			`{slack,notifications,messaging}`,
			24, 12,
		},
		{
			moduleIDs[1],
			"Invoice Generator",
			"invoice-generator",
			"Generate professional PDF invoices from project and client data. Supports custom templates, tax calculations, and multi-currency.",
			"finance",
			"2.0.1",
			"file-text",
			`{"name":"Invoice Generator","version":"2.0.1","description":"PDF invoice generation","author":"BusinessOS","category":"finance","actions":[{"name":"generate_invoice","description":"Generate a PDF invoice","type":"function","parameters":{"client_id":"string","items":"array","currency":"string","tax_rate":"number"},"returns":{"pdf_url":"string","invoice_number":"string"}},{"name":"list_invoices","description":"List generated invoices","type":"function","parameters":{"status":"string","date_range":"object"},"returns":{"invoices":"array"}}],"config_schema":{"company_name":{"type":"string","required":true},"logo_url":{"type":"string"},"default_currency":{"type":"string","default":"USD"}},"permissions":["clients:read","projects:read"]}`,
			`{invoices,billing,pdf,finance}`,
			38, 21,
		},
		{
			moduleIDs[2],
			"Time Tracker",
			"time-tracker",
			"Track time spent on tasks and projects. Includes timer widget, weekly reports, and team utilization dashboards.",
			"productivity",
			"1.5.0",
			"clock",
			`{"name":"Time Tracker","version":"1.5.0","description":"Task and project time tracking","author":"BusinessOS","category":"productivity","actions":[{"name":"start_timer","description":"Start tracking time","type":"function","parameters":{"task_id":"string","project_id":"string","notes":"string"},"returns":{"timer_id":"string"}},{"name":"stop_timer","description":"Stop the active timer","type":"function","parameters":{"timer_id":"string"},"returns":{"duration_minutes":"number"}},{"name":"get_report","description":"Generate time report","type":"function","parameters":{"period":"string","group_by":"string"},"returns":{"entries":"array","total_hours":"number"}}],"config_schema":{"rounding_interval":{"type":"number","default":15},"require_notes":{"type":"boolean","default":false}},"permissions":["tasks:read","projects:read"]}`,
			`{time,tracking,productivity,reports}`,
			45, 28,
		},
		{
			moduleIDs[3],
			"Email Campaign",
			"email-campaign",
			"Create, schedule, and send email campaigns to client segments. Includes template builder and open/click analytics.",
			"communication",
			"1.0.0",
			"mail",
			`{"name":"Email Campaign","version":"1.0.0","description":"Email campaign management","author":"BusinessOS","category":"communication","actions":[{"name":"create_campaign","description":"Create a new email campaign","type":"function","parameters":{"name":"string","subject":"string","template_id":"string","recipients":"array"},"returns":{"campaign_id":"string"}},{"name":"send_campaign","description":"Send a scheduled campaign","type":"function","parameters":{"campaign_id":"string","schedule_at":"string"},"returns":{"sent_count":"number"}}],"config_schema":{"smtp_host":{"type":"string","required":true},"from_email":{"type":"string","required":true},"from_name":{"type":"string"}},"permissions":["clients:read","contacts:read"]}`,
			`{email,campaigns,marketing}`,
			15, 8,
		},
		{
			moduleIDs[4],
			"Data Exporter",
			"data-exporter",
			"Export workspace data to CSV, Excel, or JSON. Schedule automated exports and send to cloud storage or email.",
			"utilities",
			"1.1.0",
			"download",
			`{"name":"Data Exporter","version":"1.1.0","description":"Data export to CSV/Excel/JSON","author":"BusinessOS","category":"utilities","actions":[{"name":"export_data","description":"Export data from a module","type":"function","parameters":{"source":"string","format":"string","filters":"object"},"returns":{"download_url":"string","row_count":"number"}}],"config_schema":{"default_format":{"type":"string","default":"csv"},"include_headers":{"type":"boolean","default":true}},"permissions":["data:read"]}`,
			`{export,csv,excel,data}`,
			31, 15,
		},
		{
			moduleIDs[5],
			"KPI Dashboard",
			"kpi-dashboard",
			"Build custom KPI dashboards with real-time metrics from your workspace data. Drag-and-drop widgets and scheduled email reports.",
			"analytics",
			"0.9.0",
			"bar-chart-2",
			`{"name":"KPI Dashboard","version":"0.9.0","description":"Custom KPI dashboards","author":"BusinessOS","category":"analytics","actions":[{"name":"create_dashboard","description":"Create a new dashboard","type":"function","parameters":{"name":"string","widgets":"array"},"returns":{"dashboard_id":"string"}},{"name":"add_widget","description":"Add a widget to dashboard","type":"function","parameters":{"dashboard_id":"string","widget_type":"string","data_source":"string","config":"object"},"returns":{"widget_id":"string"}}],"config_schema":{"refresh_interval":{"type":"number","default":300},"default_period":{"type":"string","default":"30d"}},"permissions":["analytics:read","projects:read","tasks:read"]}`,
			`{analytics,kpi,dashboard,metrics}`,
			19, 10,
		},
	}

	for _, m := range modules {
		_, err := pool.Exec(ctx, `
			INSERT INTO custom_modules (id, created_by, workspace_id, name, slug, description, category, version, manifest, icon, tags, install_count, star_count, is_published)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11::text[], $12, $13, true)
			ON CONFLICT (id) DO NOTHING`,
			m.id, userID, devWorkspaceID, m.name, m.slug, m.desc, m.category, m.version, m.manifest, m.icon, m.tags, m.installs, m.stars,
		)
		if err != nil {
			log.Printf("  module %s: %v", m.name, err)
		} else {
			fmt.Printf("  + %s (v%s, %s)\n", m.name, m.version, m.category)
		}
	}
}
