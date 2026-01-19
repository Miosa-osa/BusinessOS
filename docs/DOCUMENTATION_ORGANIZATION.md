# Documentation Organization

**Last Updated:** January 19, 2026

## Overview

All loose documentation files have been organized into a structured folder hierarchy under `/docs/`.

## Folder Structure

```
docs/
├── 0-START-HERE/           # Quick start guides and orientation
├── 1-active/               # Currently active work and ongoing features
├── 2-reference/            # Reference documentation
├── 3-completed/            # Completed work documentation
├── archive/                # Historical/deprecated documentation
│   ├── voice/              # Voice agent system (removed)
│   ├── integrations/       # Old integration docs
│   └── features/           # Archived feature docs
├── testing/                # Testing documentation
├── reports/                # Quality and analysis reports
├── database/               # Database setup and migrations
├── integrations/           # Integration guides
│   └── osa/                # OSA integration docs
├── economics/              # Business model and economics
├── features/               # Feature-specific documentation
├── api/                    # API documentation
├── architecture/           # System architecture docs
├── deployment/             # Deployment guides
├── development/            # Development guides
├── planning/               # Planning and roadmap docs
└── security/               # Security documentation
```

## File Moves

### Root Directory → Organized

**Archive Documentation:**
- `BACKEND_DOCS_MOVE_COMPLETE.md` → `docs/archive/`
- `DOCUMENTATION_REORGANIZATION_SUMMARY.md` → `docs/archive/`
- `ELECTRON_WEB_PARITY_REPORT.md` → `docs/archive/`
- `PR_DESCRIPTION.md` → `docs/archive/`
- `RESTART_INSTRUCTIONS.md` → `docs/archive/`
- `SUPABASE_MIGRATION_COMPLETE.md` → `docs/archive/`
- `TEAM_DOCUMENTATION_COMPLETE.md` → `docs/archive/`
- `TRANSCRIPT_FIX_TEST.md` → `docs/archive/`

**Voice-Related (Archived):**
- `FIX_AUDIO_OUTPUT_NOW.md` → `docs/archive/voice/`
- `FIX_LIVEKIT_CONNECTION.md` → `docs/archive/voice/`

**Integration Documentation:**
- `GETTING_STARTED_OSA.md` → `docs/integrations/osa/`
- `OSA_INTEGRATION_GUIDE.md` → `docs/integrations/osa/`
- `test-google-oauth-flow.md` → `docs/archive/integrations/`

**Kept in Root:**
- `README.md` - Main project README
- `CLAUDE.md` - Claude Code instructions (project-specific)

### Backend-Go Directory → Organized

**Testing Documentation:**
- `INTEGRATION_TEST_SUMMARY.md` → `docs/testing/`
- `MASTER_TEST_REPORT.md` → `docs/testing/`
- `QUICKSTART_INTEGRATION_TESTS.md` → `docs/testing/`
- `TEST_RESULTS.md` → `docs/testing/`
- `TESTING.md` → `docs/testing/`

**Reports:**
- `DUPLICATE_CODE_ANALYSIS.md` → `docs/reports/`
- `QUALITY_REPORT.md` → `docs/reports/`

**Database:**
- `SUPABASE_SETUP.md` → `docs/database/`

**Archive:**
- `REFACTORING_LOG.md` → `docs/archive/`
- `REFACTORING_PRIORITY.md` → `docs/archive/`
- `VALIDATION_IMPLEMENTATION_COMPLETE.md` → `docs/archive/`

**Voice System (Archived):**
- `LIVEKIT_ROOM_MONITOR_SUMMARY.md` → `docs/archive/voice/`
- `VOICE_AGENT_QUICK_START.md` → `docs/archive/voice/`
- `VOICE_AGENT_SWITCHOVER.md` → `docs/archive/voice/`
- `VOICE_E2E_TESTING_COMPLETE.md` → `docs/archive/voice/`
- `VOICE_SECURITY_AUDIT_REPORT.md` → `docs/archive/voice/`
- `VOICE_SECURITY_AUDIT.md` → `docs/archive/voice/`
- `VOICE_TESTING_INDEX.md` → `docs/archive/voice/`
- `VOICE_TESTING_FILES.txt` → `docs/archive/voice/`

**Kept in Backend-Go:**
- `CLAUDE.md` - Backend-specific Claude instructions

## Navigation Guide

### For New Developers
Start here:
1. `/README.md` - Project overview
2. `/docs/0-START-HERE/` - Quick start guides
3. `/docs/development/` - Development setup

### For Active Work
- `/docs/1-active/` - Current sprint work
- `/docs/planning/` - Roadmap and task lists

### For Reference
- `/docs/2-reference/` - Technical reference
- `/docs/api/` - API documentation
- `/docs/architecture/` - System design

### For Testing
- `/docs/testing/` - All testing documentation
- `/docs/reports/` - Quality reports

### For Integrations
- `/docs/integrations/` - Integration guides by provider

### For Historical Context
- `/docs/archive/` - Deprecated features and old documentation
- `/docs/3-completed/` - Completed feature documentation

## Maintenance

When adding new documentation:
1. **Active work** → `/docs/1-active/`
2. **Feature docs** → `/docs/features/[feature-name]/`
3. **API changes** → `/docs/api/`
4. **Tests** → `/docs/testing/`
5. **Completed work** → `/docs/3-completed/`
6. **Deprecated** → `/docs/archive/`

## Related Files

- `/docs/DOCUMENTATION_INVENTORY.md` - Complete file inventory
- `/docs/INDEX.md` - Documentation index
- `/docs/MASTER_DOCUMENTATION_GUIDE.md` - Comprehensive guide
