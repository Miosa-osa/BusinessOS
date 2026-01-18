# PR Review: [Brief Title]

> **Template for documenting pull request reviews**
> Copy this file and rename to: `YYYY-MM-DD-pr-###-brief-description.md`

---

## PR Information

**PR Number:** #XXX
**Date:** YYYY-MM-DD
**Author:** @username
**Reviewers:** @reviewer1, @reviewer2
**Status:** [Draft | In Review | Approved | Merged]
**Branch:** `feature/branch-name` → `main`

---

## Summary

Brief description of what this PR accomplishes (2-3 sentences).

### Problem Statement
What problem does this solve?

### Solution
High-level approach to solving the problem.

---

## Changes

### Frontend Changes
- File: `path/to/file.svelte`
  - What changed and why
  - Impact on UI/UX

### Backend Changes
- File: `path/to/file.go`
  - What changed and why
  - Impact on API

### Database Changes
- [ ] Schema changes (list migrations)
- [ ] Data migrations required
- [ ] No database changes

### Configuration Changes
- [ ] Environment variables added/changed
- [ ] Config files updated
- [ ] No configuration changes

---

## Testing

### Manual Testing
- [ ] Feature tested locally
- [ ] Edge cases verified
- [ ] Cross-browser testing (if applicable)
- [ ] Mobile responsive (if applicable)

### Automated Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] E2E tests pass (if applicable)
- [ ] Test coverage maintained/improved

### Test Evidence
```bash
# Show test output or screenshots
```

---

## Code Review Checklist

### Code Quality
- [ ] Code follows project standards
- [ ] No unnecessary complexity
- [ ] Proper error handling
- [ ] Logging added where appropriate
- [ ] No security vulnerabilities
- [ ] Performance considerations addressed

### Frontend-Specific (if applicable)
- [ ] TypeScript types correct
- [ ] Svelte runes used properly
- [ ] No console.log in production code
- [ ] Accessibility considerations
- [ ] Responsive design

### Backend-Specific (if applicable)
- [ ] Uses `slog` for logging (not fmt.Printf)
- [ ] Context propagation correct
- [ ] Proper error wrapping
- [ ] No panic in handlers
- [ ] SQL injection prevention

---

## Security Review

- [ ] Input validation
- [ ] Authentication/authorization
- [ ] No sensitive data exposed
- [ ] CSRF protection (if applicable)
- [ ] SQL injection prevention
- [ ] XSS prevention

---

## Documentation

- [ ] Code comments where needed
- [ ] API documentation updated
- [ ] README updated (if needed)
- [ ] Team review document created (this file)
- [ ] RECENT_CHANGES.md updated

---

## Deployment Notes

### Pre-Deployment Checklist
- [ ] Database migrations prepared
- [ ] Environment variables documented
- [ ] Feature flags configured (if applicable)
- [ ] Monitoring/logging ready

### Deployment Steps
1. Step 1
2. Step 2
3. ...

### Rollback Plan
How to rollback if issues occur:
1. Rollback step 1
2. Rollback step 2

---

## Performance Impact

- [ ] No performance degradation
- [ ] Performance improved
- [ ] Performance impact measured and acceptable
- [ ] Not applicable

### Metrics
If performance was measured, show before/after metrics.

---

## Dependencies

### External Dependencies
- [ ] No new dependencies
- [ ] New dependencies added (list below)

List new dependencies:
- Package name version X.X.X (reason)

### Breaking Changes
- [ ] No breaking changes
- [ ] Breaking changes (documented below)

List breaking changes:
- Change 1: What broke and migration path
- Change 2: What broke and migration path

---

## Follow-up Tasks

- [ ] Task 1 - Description (assignee)
- [ ] Task 2 - Description (assignee)
- [ ] Issue #XXX - Related issue to close

---

## Screenshots/Videos

If applicable, add screenshots or videos showing:
- Before state
- After state
- New features

---

## Reviewer Notes

### Review Comments
Key feedback from reviewers:
- Reviewer 1: Comment summary
- Reviewer 2: Comment summary

### Changes Made
Changes made based on feedback:
- Change 1
- Change 2

---

## Approval

### Final Checklist Before Merge
- [ ] All tests pass
- [ ] Code reviewed and approved
- [ ] Documentation complete
- [ ] No merge conflicts
- [ ] CI/CD pipeline green
- [ ] Team lead approval (if required)

### Merge Decision
**Decision:** [Approved and Merged | Needs Changes | Rejected]
**Date Merged:** YYYY-MM-DD
**Merged By:** @username

---

**Document Created:** YYYY-MM-DD
**Last Updated:** YYYY-MM-DD
**Status:** [Active | Merged | Archived]
