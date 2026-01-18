# BusinessOS Documentation Map

> **Visual Guide:** Where to find everything in the documentation

---

## 🗺️ Documentation Structure Overview

```
BusinessOS2/
│
├── 📄 README.md                           ← Project overview + Team review links
│
├── 📁 docs/                               ← Main documentation hub
│   ├── 🎯 TEAM_START_HERE.md             ← START HERE (main entry point)
│   ├── 📋 RECENT_CHANGES.md              ← Latest project-wide updates
│   ├── 📚 START_HERE.md                  ← Project introduction
│   ├── 📖 TECHNICAL_REFERENCE.md         ← System architecture
│   ├── 💼 EXECUTIVE_SUMMARY.md           ← Business overview
│   │
│   ├── 📂 team-review/                   ← Team updates & PR reviews
│   │   ├── README.md                     ← Team review index
│   │   ├── DOCUMENTATION_MAP.md          ← This file
│   │   ├── pr-reviews/                   ← PR review documents
│   │   ├── sprint-summaries/             ← Sprint reports
│   │   └── release-notes/                ← Release documentation
│   │
│   ├── 📂 features/                      ← Feature specifications
│   │   ├── onboarding/                   ← Onboarding feature docs
│   │   ├── integrations/                 ← Integration features
│   │   └── ...
│   │
│   ├── 📂 architecture/                  ← Architecture documents
│   │   ├── system-design/
│   │   ├── data-models/
│   │   └── ...
│   │
│   ├── 📂 api/                          ← API specifications
│   │   ├── API_DOCUMENTATION_INDEX.md
│   │   ├── API_CHEATSHEET.md
│   │   └── ...
│   │
│   ├── 📂 database/                     ← Database schemas & migrations
│   │   ├── DATABASE_SETUP.md
│   │   ├── migrations/
│   │   └── ...
│   │
│   ├── 📂 development/                  ← Development guides
│   │   ├── ENVIRONMENT_SETUP.md
│   │   ├── ENVIRONMENT_VALIDATION_GUIDE.md
│   │   └── ...
│   │
│   ├── 📂 integrations/                 ← External integrations
│   │   ├── TEAM_INTEGRATION_SETUP_GUIDE.md
│   │   └── ...
│   │
│   ├── 📂 planning/                     ← Sprint plans & roadmaps
│   ├── 📂 reports/                      ← Test reports & reviews
│   ├── 📂 research/                     ← Research documents
│   ├── 📂 security/                     ← Security assessments
│   └── 📂 deployment/                   ← Deployment guides
│
├── 📁 frontend/                          ← Frontend code
│   └── 📁 docs/                         ← Frontend documentation
│       ├── 📂 team-review/              ← Frontend team updates
│       │   └── README.md                ← Frontend review index
│       ├── PAGES_ARCHITECTURE.md
│       ├── 3D_DESKTOP_APP_INTEGRATION.md
│       └── gesture-system/
│
└── 📁 desktop/backend-go/               ← Backend code
    └── 📁 docs/                         ← Backend documentation
        ├── 📂 team-review/              ← Backend team updates
        │   └── README.md                ← Backend review index
        ├── ENVIRONMENT_SETUP.md
        ├── integrations/
        └── ...
```

---

## 🎯 How to Navigate

### I want to...

#### Get started with the project
1. **[docs/TEAM_START_HERE.md](../TEAM_START_HERE.md)** - Main entry point
2. **[docs/START_HERE.md](../START_HERE.md)** - Project introduction
3. **[docs/development/ENVIRONMENT_SETUP.md](../development/ENVIRONMENT_SETUP.md)** - Setup guide

#### See what changed recently
1. **[docs/RECENT_CHANGES.md](../RECENT_CHANGES.md)** - Project-wide changes
2. **[frontend/docs/team-review/](../../../frontend/docs/team-review/)** - Frontend changes
3. **[desktop/backend-go/docs/team-review/](../../../desktop/backend-go/docs/team-review/)** - Backend changes

#### Review a PR
1. **[docs/team-review/pr-reviews/](pr-reviews/)** - PR review documents
2. Team-specific review folders

#### Understand the architecture
1. **[docs/TECHNICAL_REFERENCE.md](../TECHNICAL_REFERENCE.md)** - System overview
2. **[docs/architecture/](../architecture/)** - Architecture details
3. **[docs/IOS_TO_DESKTOP_ARCHITECTURE.md](../IOS_TO_DESKTOP_ARCHITECTURE.md)** - iOS to Desktop flow

#### Work with APIs
1. **[docs/API_DOCUMENTATION_INDEX.md](../API_DOCUMENTATION_INDEX.md)** - API index
2. **[docs/API_CHEATSHEET.md](../API_CHEATSHEET.md)** - Quick reference
3. **[docs/api/](../api/)** - Detailed API specs

#### Set up integrations
1. **[docs/integrations/TEAM_INTEGRATION_SETUP_GUIDE.md](../integrations/TEAM_INTEGRATION_SETUP_GUIDE.md)** - Setup guide
2. **[docs/integrations/](../integrations/)** - Integration docs

#### Work on a feature
1. **[docs/features/](../features/)** - Feature specifications
2. Team-specific docs (frontend/backend)

#### Deploy the application
1. **[docs/deployment/](../deployment/)** - Deployment guides
2. **[docs/development/ENVIRONMENT_VALIDATION_GUIDE.md](../development/ENVIRONMENT_VALIDATION_GUIDE.md)**

---

## 📚 Documentation Categories

### By Audience

#### New Team Members
```
Start → docs/TEAM_START_HERE.md
Then → docs/START_HERE.md
Then → docs/development/ENVIRONMENT_SETUP.md
Then → Team-specific docs (frontend or backend)
```

#### Frontend Developers
```
Main Hub → frontend/docs/team-review/README.md
Architecture → frontend/docs/PAGES_ARCHITECTURE.md
Components → Frontend component library
Updates → frontend/docs/team-review/
```

#### Backend Developers
```
Main Hub → desktop/backend-go/docs/team-review/README.md
API Docs → docs/API_DOCUMENTATION_INDEX.md
Setup → desktop/backend-go/docs/ENVIRONMENT_SETUP.md
Updates → desktop/backend-go/docs/team-review/
```

#### Product/Project Managers
```
Overview → docs/EXECUTIVE_SUMMARY.md
Recent → docs/RECENT_CHANGES.md
Planning → docs/planning/
Features → docs/features/
```

#### QA/Testing
```
Testing Plan → docs/OSA_BUILD_TESTING_PLAN.md
Test Guide → docs/TESTING_OSA_ONBOARDING.md
Validation → docs/development/ENVIRONMENT_VALIDATION_GUIDE.md
```

---

## 🔍 Finding Specific Information

### Frontend Topics

| Topic | Location |
|-------|----------|
| Page structure | `frontend/docs/PAGES_ARCHITECTURE.md` |
| Recent changes | `frontend/docs/team-review/` |
| Components | Frontend component library |
| Desktop mode | `frontend/docs/3D_DESKTOP_APP_INTEGRATION.md` |
| Gesture system | `frontend/docs/gesture-system/` |

### Backend Topics

| Topic | Location |
|-------|----------|
| API reference | `docs/API_DOCUMENTATION_INDEX.md` |
| Recent changes | `desktop/backend-go/docs/team-review/` |
| Environment setup | `desktop/backend-go/docs/ENVIRONMENT_SETUP.md` |
| Integrations | `docs/integrations/TEAM_INTEGRATION_SETUP_GUIDE.md` |
| Database | `docs/database/` |

### Feature Topics

| Topic | Location |
|-------|----------|
| Onboarding | `docs/features/onboarding/` |
| Google OAuth | `docs/features/onboarding/GOOGLE_OAUTH_ONBOARDING.md` |
| OSA Build | `docs/OSA_BUILD_*.md` files |
| Social system | `docs/OSA_BUILD_SOCIAL_SYSTEM.md` |

### Process Topics

| Topic | Location |
|-------|----------|
| PR reviews | `docs/team-review/pr-reviews/` |
| Sprint planning | `docs/planning/` |
| Testing | `docs/OSA_BUILD_TESTING_PLAN.md` |
| Deployment | `docs/deployment/` |

---

## 🎨 Visual Flow: Where to Start

```
┌─────────────────────────────────────────────────────────────────┐
│                         NEW TO PROJECT?                         │
│                                                                 │
│                  Start: docs/TEAM_START_HERE.md                 │
│                              ↓                                  │
│                  Read: docs/START_HERE.md                       │
│                              ↓                                  │
│                  Setup: docs/development/                       │
│                         ENVIRONMENT_SETUP.md                    │
│                              ↓                                  │
│              Choose your role:                                  │
│         ┌────────────────┴────────────────┐                    │
│         ↓                                 ↓                     │
│    Frontend Dev                     Backend Dev                 │
│    frontend/docs/team-review/       desktop/backend-go/        │
│                                     docs/team-review/           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    LOOKING FOR UPDATES?                         │
│                                                                 │
│              Start: docs/RECENT_CHANGES.md                      │
│                              ↓                                  │
│          Check team-specific updates:                           │
│    ┌────────────────────┬────────────────────┐                 │
│    ↓                    ↓                    ↓                  │
│ Frontend            Backend            Cross-team               │
│ frontend/docs/      backend/docs/      docs/team-review/       │
│ team-review/        team-review/                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   WORKING ON A FEATURE?                         │
│                                                                 │
│              Start: docs/features/{feature-name}/               │
│                              ↓                                  │
│          Check technical reference:                             │
│              docs/TECHNICAL_REFERENCE.md                        │
│                              ↓                                  │
│          Check API if needed:                                   │
│              docs/API_DOCUMENTATION_INDEX.md                    │
│                              ↓                                  │
│          Document your work in:                                 │
│              docs/team-review/                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📂 Folder Purposes Quick Reference

| Folder | Purpose | Who Uses It |
|--------|---------|-------------|
| `docs/` | Main documentation hub | Everyone |
| `docs/team-review/` | Team updates & PR reviews | All teams |
| `docs/features/` | Feature specifications | Dev teams + PM |
| `docs/architecture/` | System design | Architects + Senior devs |
| `docs/api/` | API documentation | Backend + Frontend devs |
| `docs/database/` | Database schemas | Backend devs |
| `docs/development/` | Development setup | All devs |
| `docs/integrations/` | External integrations | Backend devs |
| `docs/planning/` | Sprints & roadmaps | PM + Leads |
| `docs/reports/` | Test reports | QA + Devs |
| `frontend/docs/` | Frontend-specific | Frontend team |
| `desktop/backend-go/docs/` | Backend-specific | Backend team |

---

## 🔄 Document Lifecycle

### Where to Add New Documentation

1. **PR Review**: Add to `docs/team-review/pr-reviews/`
2. **Feature Doc**: Add to `docs/features/{feature-name}/`
3. **Architecture**: Add to `docs/architecture/`
4. **Frontend Work**: Add to `frontend/docs/team-review/`
5. **Backend Work**: Add to `desktop/backend-go/docs/team-review/`
6. **Integration**: Add to `docs/integrations/`
7. **Sprint Plan**: Add to `docs/planning/`

### When to Update

- **RECENT_CHANGES.md**: After major features or milestones
- **Team review folders**: After each PR or sprint
- **API docs**: When API changes
- **Feature docs**: When feature is completed

---

## 🆘 Still Can't Find It?

### Search Strategy

1. **Check main hub**: [docs/TEAM_START_HERE.md](../TEAM_START_HERE.md)
2. **Search by keyword**: Use IDE search across `docs/`
3. **Check recent changes**: [docs/RECENT_CHANGES.md](../RECENT_CHANGES.md)
4. **Ask the team**: Use team communication channels

### Common Questions

**Q: Where's the API documentation?**
A: [docs/API_DOCUMENTATION_INDEX.md](../API_DOCUMENTATION_INDEX.md)

**Q: How do I set up integrations?**
A: [docs/integrations/TEAM_INTEGRATION_SETUP_GUIDE.md](../integrations/TEAM_INTEGRATION_SETUP_GUIDE.md)

**Q: What changed recently?**
A: [docs/RECENT_CHANGES.md](../RECENT_CHANGES.md)

**Q: Where are frontend updates?**
A: [frontend/docs/team-review/](../../../frontend/docs/team-review/)

**Q: Where are backend updates?**
A: [desktop/backend-go/docs/team-review/](../../../desktop/backend-go/docs/team-review/)

---

**Maintained by:** Development Team
**Last Updated:** January 19, 2026
**Version:** 1.0.0
