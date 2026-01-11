# 🎯 Claude Code - Commands Cheatsheet

Quick reference for using your optimized Claude Code setup.

---

## 🎯 Testing Skills

### Test Go Backend Expert Skill
```
Add a new endpoint GET /api/v1/health that returns server status
```
**Expected:** Skill `go-backend-expert` activates, follows Handler→Service→Repository pattern, uses `slog`

### Test Frontend Skill
```
Create a reusable Card component with props: title, description, and onClick handler
```
**Expected:** Skill `svelte-frontend-expert` activates, uses Svelte 5 syntax, TypeScript

### Test Migration Skill
```
Add a column 'last_login' to the users table
```
**Expected:** Skill `database-migration-expert` activates, creates UP and DOWN migrations

### Test Testing Skill
```
Write unit tests for the authentication service
```
**Expected:** Skill `testing-expert` activates, follows testing best practices

---

## 🤖 Invoking Custom Agents

### Backend Specialist
```
Use the backend-specialist to implement a new API endpoint for user preferences
```
```
Have the backend-specialist refactor the authentication handlers
```

### Frontend Specialist
```
Use the frontend-specialist to build a settings modal component
```
```
Ask the frontend-specialist to implement the dashboard page
```

### Migration Specialist
```
Use the migration-specialist to create a migration for the notifications table
```
```
Have the migration-specialist design the schema for the audit log
```

---

## 📋 Verifying Configuration

### Check Skills
```bash
# List all skills
ls .claude/skills/

# View skill content
cat .claude/skills/go-backend-expert/SKILL.md
```

### Check Agents
```bash
# List all agents
ls .claude/agents/

# View agent content
cat .claude/agents/backend-specialist.md
```

### Check Hooks
```bash
# View settings
cat .claude/settings.json

# Check command log (created by hooks)
tail -20 .claude/command-log.txt
```

---

## 🔍 Debugging

### Check if Skill is Active
During a Claude Code conversation, ask:
```
What skills are currently active?
```

### Check Active Agent
```
Which agent are you using right now?
```

### View Hook Logs
```bash
# See all commands executed
cat .claude/command-log.txt

# Live monitoring
tail -f .claude/command-log.txt
```

---

## ⚙️ Managing Configuration

### Temporarily Disable Skill
```bash
# Rename to disable
mv .claude/skills/go-backend-expert .claude/skills/_go-backend-expert.disabled

# Rename back to enable
mv .claude/skills/_go-backend-expert.disabled .claude/skills/go-backend-expert
```

### Edit Skill
```bash
# Edit with your preferred editor
nano .claude/skills/go-backend-expert/SKILL.md
# or
code .claude/skills/go-backend-expert/SKILL.md
```

### Edit Agent
```bash
nano .claude/agents/backend-specialist.md
```

### Edit Hooks
```bash
nano .claude/settings.json
```

---

## 🔌 MCP Servers (Optional)

### Install PostgreSQL Access
```bash
claude mcp add --scope project --transport stdio business-os-db -- \
  npx -y @modelcontextprotocol/server-postgres \
  postgresql://user:pass@localhost:5432/business_os
```

Then in Claude Code:
```
Show me all custom agents in the database
What's the schema of the agents table?
```

### Install GitHub Integration
```bash
claude mcp add --scope user --transport http github \
  https://api.githubcopilot.com/mcp/
```

Then in Claude Code:
```
Create a PR for my changes
Review PR #123
Show open issues labeled "bug"
```

### Manage MCP Servers
```bash
# List installed servers
claude mcp list

# View server details
claude mcp get business-os-db

# Remove server
claude mcp remove business-os-db

# Check status in Claude Code
/mcp
```

---

## 📝 Common Workflows

### Creating New Backend Feature
```
I need to add a new feature: user can save favorite items
Use the backend-specialist to:
1. Design the database schema
2. Create the migration
3. Implement the API endpoints
4. Write tests
```

### Creating New Frontend Component
```
Use the frontend-specialist to create a FavoriteButton component that:
- Shows a star icon
- Toggles filled/outlined on click
- Calls /api/favorites endpoint
- Shows loading state
- Has proper error handling
```

### Database Changes
```
Use the migration-specialist to:
1. Add a favorites table with user_id and item_id
2. Add indexes for performance
3. Create sqlc queries for CRUD operations
4. Update the Go code
```

---

## 🚀 Advanced Usage

### Running Multiple Agents in Parallel
```
I need to implement notifications feature.

In parallel:
1. Have backend-specialist implement the API
2. Have frontend-specialist build the UI
3. Have migration-specialist create the schema

Then integrate everything together.
```

### Combining Skills and Agents
```
Use the backend-specialist to add a health check endpoint.
Make sure it follows the go-backend-expert skill patterns.
```

---

## 🐛 Troubleshooting

### Skill Not Activating
**Problem:** Skill doesn't apply automatically

**Fix:** Check the `description` field in SKILL.md
```bash
cat .claude/skills/go-backend-expert/SKILL.md | head -5
```

Make description more specific with keywords:
```yaml
---
name: go-backend-expert
description: Expert in Go backend with Handler→Service→Repository, slog logging, sqlc. Use when working with .go files, backend code, or files in desktop/backend-go/
---
```

### Hook Not Executing
**Problem:** Hook configured but not running

**Check syntax:**
```bash
# Validate JSON
jq . .claude/settings.json
```

**Check permissions:**
```bash
chmod +x .claude/hooks/*.sh
```

**Test manually:**
```bash
echo '{"tool_input":{"file_path":"test.go"}}' | .claude/hooks/auto-test.sh
```

### Agent Not Found
**Problem:** "Agent not found" error

**Fix:** Verify file naming
```bash
ls -la .claude/agents/

# File must be .md not .txt
# Name in frontmatter must match filename
```

---

## 📚 Documentation Links

- **Complete Guide:** `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`
- **Quick Start:** `docs/CLAUDE_CODE_QUICKSTART.md`
- **Summary:** `docs/CLAUDE_CODE_SUMMARY.md`
- **Config Docs:** `.claude/README.md`
- **Official Docs:** https://docs.claude.ai/claude-code

---

## 💡 Pro Tips

1. **Skills activate automatically** - Just work normally and they'll apply when relevant
2. **Use agents for complex tasks** - Delegate entire features to specialized agents
3. **Hooks work transparently** - Code formatting, validation happens automatically
4. **MCP servers boost productivity** - Direct DB/GitHub access saves time
5. **Customize as you go** - Add new skills/agents for your specific workflows

---

**Last updated:** 2026-01-11
