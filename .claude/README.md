# Claude Code Configuration for BusinessOS

This directory contains optimizations for Claude Code:

## Structure

```
.claude/
├── skills/               # Auto-applied knowledge
│   ├── go-backend-expert/
│   ├── svelte-frontend-expert/
│   ├── database-migration-expert/
│   └── testing-expert/
├── agents/               # Specialized sub-agents
│   ├── backend-specialist.md
│   ├── frontend-specialist.md
│   └── migration-specialist.md
├── hooks/                # Automation scripts
│   └── auto-test.sh
├── settings.json         # Configuration
└── command-log.txt       # Command history (auto-generated)
```

## Skills

Skills are automatically applied when Claude detects relevant context:

- **go-backend-expert**: Go backend patterns, slog, sqlc
- **svelte-frontend-expert**: Svelte 5, TypeScript, Tailwind
- **database-migration-expert**: PostgreSQL migrations
- **testing-expert**: Go and frontend testing

## Agents

Specialized agents for complex tasks:

- **backend-specialist**: API development, DB operations
- **frontend-specialist**: UI components, state management
- **migration-specialist**: Database schema changes

## Hooks

Automated workflows:

- **PostToolUse**: Auto-format Go code with gofmt
- **PreToolUse**: Block editing sensitive files (.env, secrets)
- **PreToolUse (Bash)**: Log all commands to command-log.txt

## Usage

Skills and agents are used automatically. To verify:

```bash
# List available skills
ls .claude/skills/

# List available agents
ls .claude/agents/

# View configuration
cat .claude/settings.json
```

## See Also

- Full documentation: `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`
- Project conventions: `CLAUDE.md`
