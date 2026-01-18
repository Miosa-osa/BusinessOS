# Team Review Documents

> **Purpose:** Central location for PR reviews, change summaries, and team updates

---

## 📋 Quick Links

- **[Main Entry Point](../TEAM_START_HERE.md)** - Start here
- **[Recent Changes](../RECENT_CHANGES.md)** - Latest updates across project
- **[Frontend Team Review](../../frontend/docs/team-review/)** - Frontend-specific updates
- **[Backend Team Review](../../desktop/backend-go/docs/team-review/)** - Backend-specific updates

---

## 📂 What Goes Here?

This folder contains:

### Pull Request Reviews
- PR review documents
- Code review summaries
- Merge checklists

### Change Summaries
- Sprint summaries
- Feature completion reports
- Release notes

### Team Updates
- Architecture decisions affecting multiple teams
- Cross-functional feature updates
- System-wide changes

---

## 🗂️ Organization

```
team-review/
├── README.md                    # This file
├── pr-reviews/                  # PR-specific reviews
│   └── YYYY-MM-DD-pr-###.md
├── sprint-summaries/            # Sprint completion reports
│   └── YYYY-QX-sprint-X.md
└── release-notes/               # Release documentation
    └── vX.X.X-release-notes.md
```

---

## 📝 Document Templates

### PR Review Template

```markdown
# PR Review: [Title]

**PR Number:** #XXX
**Date:** YYYY-MM-DD
**Author:** @username
**Reviewers:** @user1, @user2

## Summary
Brief description of changes

## Changes
- Frontend: ...
- Backend: ...
- Database: ...

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Deployment Notes
Any special deployment considerations

## Follow-up Tasks
- [ ] Task 1
- [ ] Task 2
```

### Sprint Summary Template

```markdown
# Sprint Summary: Q1 Sprint X

**Dates:** YYYY-MM-DD to YYYY-MM-DD
**Team Size:** X developers

## Completed Features
- Feature 1
- Feature 2

## In Progress
- Feature X (80% complete)

## Blockers Resolved
- Blocker 1
- Blocker 2

## Next Sprint
- Priority 1
- Priority 2
```

---

## 🔗 Related Documentation

- **Architecture**: [../architecture/](../architecture/)
- **Features**: [../features/](../features/)
- **Implementation**: [../implementation/](../implementation/)
- **Planning**: [../planning/](../planning/)

---

## 📊 Current Status

### Active Reviews
- Check individual team folders for active PR reviews

### Recent Summaries
- See [../RECENT_CHANGES.md](../RECENT_CHANGES.md) for latest project-wide updates

---

**Maintained by:** Development Team
**Last Updated:** January 19, 2026
