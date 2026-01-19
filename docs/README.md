# BusinessOS Documentation

## Background Jobs System (NEW)

The Background Jobs System enables asynchronous task processing with PostgreSQL-backed queues, retry logic, and scheduling.

**Quick Links:**
- [MASTER DOCUMENTATION](BACKGROUND_JOBS_COMPLETE_DOCUMENTATION.md) - Complete guide (58KB, 15,000 words)
- [Quick Start](BACKGROUND_JOBS_QUICKSTART.md) - Get started in 5 minutes
- [Final Summary](FINAL_SUMMARY.md) - Portuguese executive summary
- [Delivery Checklist](DELIVERY_CHECKLIST.md) - Complete verification

## Folder Structure

| Folder | Description |
|--------|-------------|
| `adrs/` | Architecture Decision Records |
| `api/` | API documentation and endpoint references |
| `architecture/` | System architecture, diagrams, and design docs |
| `archive/` | Historical docs, status updates, and completed tasks |
| `context/` | Context templates and examples |
| `database/` | Database setup, schema, and troubleshooting |
| `decisions/` | Technical decision logs |
| `deployment/` | Deployment guides and infrastructure docs |
| `development/` | Developer guides, setup, and contribution docs |
| `features/` | Feature specs, roadmaps, and module documentation |
| `integrations/` | Third-party integration docs (Google, Slack, etc.) |
| `patterns/` | Code patterns and templates |
| `planning/` | Project planning, taxonomy, and Linear issues |
| `reports/` | Generated reports |
| `research/` | Research documents and analysis |
| `security/` | Security documentation and audits |
| `sorxdocs/` | SORX engine documentation |

## Quick Links

- **Getting Started**: `development/DEVELOPER_QUICKSTART.md`
- **Architecture Overview**: `architecture/ARCHITECTURE.md`
- **API Reference**: `api/API_REFERENCE.md`
- **Database Setup**: `database/DATABASE_SETUP.md`
- **Deployment Guide**: `deployment/DEPLOYMENT_GUIDE.md`

## For Developers

### Claude Code Workflow
- [**CLAUDE.md (root)**](../CLAUDE.md) - Complete workflow guide with subagents
- [**ADVANCED_TASKMANAGER.md**](ADVANCED_TASKMANAGER.md) - Complete system: Microtasks, Milestones, Feedback Loop
- [**TASKMANAGER_EXAMPLES.md**](TASKMANAGER_EXAMPLES.md) - TaskManager in action
- [**WORKFLOW_EXAMPLE.md**](WORKFLOW_EXAMPLE.md) - Practical decomposition example

### Development Commands

**Starting development:**
```bash
# From project root
./dev.sh
```

**Running tests:**
```bash
cd desktop/backend-go
./scripts/tests/test_all_endpoints.sh
```

**Database migrations:**
```bash
go run scripts/debug/run_q1_migrations.go
```

## Documentation Standards

1. **File naming:**
   - Use snake_case: `feature_name_guide.md`
   - Include date for status reports: `2026-01-06_status.md`

2. **Location:**
   - Root: Only README.md and CLAUDE.md
   - Everything else: organized in docs/

3. **Keep updated:**
   - Update docs when features change
   - Archive old status reports in archive/
