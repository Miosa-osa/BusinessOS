# OSA Build Onboarding Documentation

Comprehensive API and integration documentation for the OSA Build onboarding system that powers personalized OS creation in BusinessOS.

## Documents in This Directory

### 1. API_OSA_ONBOARDING.md (Primary Reference)
**1,287 lines | 32 KB | Comprehensive Guide**

The complete API reference with every detail needed for implementation and maintenance.

**Includes:**
- Full endpoint specifications (6 endpoints documented)
- Request/response schemas with examples
- Error handling and status codes
- Data models and types
- Complete integration guide with TypeScript/Svelte examples
- Testing instructions (cURL and automated tests)
- Database migration guidance
- Environment configuration
- Troubleshooting guide
- Production deployment considerations

**Read this if you:**
- Are implementing the onboarding flow
- Need complete endpoint documentation
- Are debugging API issues
- Want architectural understanding
- Need to maintain or extend the system

**Time to read:** 20-30 minutes

---

### 2. ONBOARDING_QUICK_REFERENCE.md (Quick Start)
**382 lines | 9 KB | Quick Reference**

A condensed guide for quick lookups and rapid integration.

**Includes:**
- 5-minute integration checklist
- Endpoint cheat sheet (one-liners)
- Response models at a glance
- Frontend flow template (copy-paste ready)
- Error handling patterns
- State management essentials
- Testing quick commands
- Common issues and fixes
- Key files in codebase
- Performance tips

**Read this if you:**
- Need to integrate quickly
- Want a cheat sheet while coding
- Are looking for copy-paste templates
- Need quick troubleshooting
- Want file reference paths

**Time to read:** 5-10 minutes

---

## Document Comparison

| Aspect | Full Reference | Quick Reference |
|--------|----------------|-----------------|
| Depth | Complete | Summary |
| Use Case | Implementation, maintenance | Development, lookup |
| Detail Level | Every option documented | Essential only |
| Examples | Extensive | Focused |
| Troubleshooting | Comprehensive | Common issues |
| Learning Curve | Steeper | Gentle |
| Reference Format | Searchable | Quick scan |

---

## The Onboarding System at a Glance

### What It Does
The OSA Build onboarding system takes a new user through a 13-step journey to:
1. Collect basic info (email, username)
2. Connect integrations (Gmail, Calendar)
3. Analyze their interests and workflow patterns
4. Generate 3 personalized insights
5. Create 4 customized starter applications
6. Launch into their personalized OS

### Architecture
```
Frontend (SvelteKit)
    ↓
API Client ($lib/api/osa-onboarding)
    ↓
Backend (Go/Gin)
    ↓
Services (business logic)
    ↓
OSA Orchestrator (app generation)
    ↓
Database (analysis storage)
```

### Core Endpoints
- `POST /api/osa-onboarding/analyze` - Analyze user
- `POST /api/osa-onboarding/generate-apps` - Create starter apps
- `GET /api/osa-onboarding/apps-status` - Check generation status
- `GET /api/osa-onboarding/profile` - Retrieve saved profile
- `GET /api/users/check-username/:username` - Validate username
- `PATCH /api/users/me/username` - Update username

---

## Getting Started

### For Frontend Developers
1. Read: **ONBOARDING_QUICK_REFERENCE.md** (5 min)
2. Copy: Frontend flow template
3. Integrate: API client and store
4. Reference: API_OSA_ONBOARDING.md as needed

### For Backend Developers
1. Read: **API_OSA_ONBOARDING.md** sections:
   - Data Models
   - Endpoints
   - Database Migrations
   - Troubleshooting
2. Review: Source code
   - `/desktop/backend-go/internal/handlers/osa_onboarding_handler.go`
   - `/desktop/backend-go/internal/services/osa_onboarding_service.go`
3. Test: Using provided cURL examples

### For QA Engineers
1. Read: **ONBOARDING_QUICK_REFERENCE.md** (testing section)
2. Read: **API_OSA_ONBOARDING.md** (testing section)
3. Use: cURL examples to validate
4. Reference: Common issues for edge cases

### For Product Managers
1. Read: **API_OSA_ONBOARDING.md** (overview section)
2. Understand: 13-step flow diagram
3. Reference: Error handling and user experience implications
4. Note: Rate limiting and performance considerations

---

## Key Concepts

### Flow Stages
- **Stage 1 (Steps 1-5):** Signup & integrations
- **Stage 2 (Steps 6-9):** Analysis & personalization
- **Stage 3 (Steps 10-13):** App generation & launch

### Data Flow
1. User provides email → Analysis generated
2. Analysis + interests → 4 starter apps created
3. Apps generated asynchronously (polling required)
4. Profile saved to database for future reference

### Key Decisions
- **Async generation:** Apps don't block the flow
- **Polling pattern:** Client polls until ready (no WebSockets)
- **Fallback analysis:** AI analysis fails gracefully
- **Profile persistence:** Full history saved in database

---

## Common Workflows

### Workflow 1: Implement Onboarding Screen
1. Reference: Quick Reference (Frontend Flow Template)
2. Copy: Flow template code
3. Integrate: API calls at each step
4. Store: Save state using onboardingStore
5. Test: Using Quick Reference (Testing section)

### Workflow 2: Debug Analysis Not Working
1. Check: AI_PROVIDER env variable
2. Verify: ANTHROPIC_API_KEY exists
3. Test: cURL analyze endpoint (Quick Reference)
4. Review: Backend logs for errors (API Reference)
5. Fallback: System uses defaults on failure

### Workflow 3: Debug App Generation Timeout
1. Increase: Polling timeout (Quick Reference)
2. Check: OSA orchestrator status
3. Verify: Database connection
4. Review: App generation logs (API Reference)
5. Monitor: Exponential backoff not too aggressive

### Workflow 4: Add New Onboarding Screen
1. Read: Integration guide in API Reference
2. Add: State management in onboardingStore
3. Create: UI component
4. Test: Following testing section

---

## Important Files Reference

### Frontend
- **API Client**: `/frontend/src/lib/api/osa-onboarding/index.ts`
- **Types**: `/frontend/src/lib/api/osa-onboarding/types.ts`
- **Store**: `/frontend/src/lib/stores/onboardingStore.ts`
- **Routes**: `/frontend/src/routes/onboarding/`
- **Components**: `/frontend/src/lib/components/onboarding/`

### Backend
- **Handler**: `/desktop/backend-go/internal/handlers/osa_onboarding_handler.go`
- **Service**: `/desktop/backend-go/internal/services/osa_onboarding_service.go`
- **Routing**: Set up in main handler registration

### Database
- **Migrations**: Run `go run ./cmd/migrate`
- **Table**: `workspace_onboarding_profiles`
- **Schema**: Defined in migrations

---

## Development Checklist

### Before Starting
- [ ] Read appropriate document (Quick Reference or API Reference)
- [ ] Understand the 13-step flow
- [ ] Know your role (frontend/backend/testing)

### During Development
- [ ] Reference templates and examples
- [ ] Test as you go (use Quick Reference testing section)
- [ ] Follow error handling patterns
- [ ] Use proper logging (slog on backend)

### Before Deployment
- [ ] All tests pass
- [ ] Error cases handled gracefully
- [ ] Fallbacks in place (analysis, timeouts)
- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] Load testing (polling pattern)

---

## Troubleshooting Path

**Problem → Solution Path:**
1. Check error code (API Reference: API Response Codes Reference)
2. Match to issue (API Reference: Troubleshooting)
3. Apply solution
4. If persists, check logs and database (Troubleshooting section)

**Common Issues:**
- **401 Unauthorized** → Invalid token (Quick Reference)
- **Empty insights** → AI provider not configured (API Reference)
- **Timeout** → Increase polling delay (Quick Reference)
- **Profile not found** → Check workspace_id (API Reference)

---

## Performance Considerations

### Frontend
- Parallel requests: Check username while analyzing
- Exponential backoff: 2s → 10s max for polling
- Client caching: onboardingStore persists to localStorage
- Early validation: Check username immediately

### Backend
- Database indexes on workspace_id and user_id
- AI provider timeout: 30s default
- Profile caching not implemented (db hit each time)
- Rate limiting not enforced (future work)

### Database
- workspace_onboarding_profiles: ~1KB per profile
- Estimated growth: ~100KB per 100 users
- Indexes speed up lookups 100x+
- Cleanup old profiles regularly

---

## Security Notes

All endpoints require Bearer token authentication except:
- `GET /api/users/check-username/:username` (public)

Passwords are never stored in profiles. User analysis is non-sensitive. OAuth tokens are encrypted in database.

---

## Support & Escalation

### Quick Issues (5 min)
- Check Quick Reference troubleshooting
- Use cURL to test endpoint
- Check logs: `docker logs businessos-backend`

### Medium Issues (20 min)
- Reference API_OSA_ONBOARDING.md troubleshooting
- Check database: `psql $DATABASE_URL`
- Verify environment: `env | grep OSA`

### Complex Issues
- Full API Reference investigation
- Team sync in #backend-support
- Code review of implementation

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | Jan 2024 | Initial release |

## Last Updated
January 18, 2024

## Contributors
- Backend: Go service implementation
- Frontend: Svelte/SvelteKit integration
- Documentation: Complete API reference

---

## Next Steps

**To get started:**
1. Choose your document: API Reference or Quick Reference
2. Find your section: Implementation, testing, or troubleshooting
3. Use provided examples and templates
4. Refer back as needed during development

**Questions?**
- Check the appropriate document first
- Slack: #backend-support
- Email: api-support@businessos.io

---

**Happy building!**
