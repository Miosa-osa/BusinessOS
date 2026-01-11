# CLAUDE CODE ECOSYSTEM v3.0

## 🚨 START EVERY SESSION WITH: @prime

If no @prime has been run, say: "Let me initialize context first..." then execute @prime.

---

# @prime - CONTEXT INITIALIZATION

## What @prime Does
1. **Detects** project type, directory, git status
2. **Loads** project patterns and conventions
3. **Queries** memory for past decisions
4. **Shows** current tasks from TaskMaster
5. **Activates** appropriate agents
6. **Outputs** context card

## @prime Output Format
```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                              CONTEXT PRIMED                                   ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║  📁 PROJECT: [name] | Type: [type] | Branch: [branch]                        ║
║  🤖 AGENTS: @[primary] + @[support agents]                                   ║
║  💾 MEMORY: [X] decisions, [Y] patterns loaded                               ║
║  📋 TASKS: [X] pending | Next: [highest priority task]                       ║
║  📂 MODIFIED: [recently changed files]                                       ║
║  🎯 SUGGESTED: 1. [action] 2. [action] 3. [action]                           ║
╚═══════════════════════════════════════════════════════════════════════════════╝
Ready. What do you need?
```

## @prime Variants
| Command | Use Case |
|---------|----------|
| `@prime` | Full initialization (default) |
| `@prime --quick` | Skip memory query, faster |
| `@prime --review` | Add @code-reviewer, @security-auditor |
| `@prime --debug` | Add @debugger agent |
| `@prime --task TASK-XXX` | Focus on specific task |

---

# RESPONSE FORMAT

## Every Response Header
```
┌─────────────────────────────────────────────────────────────────┐
│ 🤖 Agent: @[name]  │ 📋 Task: [type]  │ 💾 Memory: [Y/N]       │
└─────────────────────────────────────────────────────────────────┘
```

## Before Saying "Done"
```
┌─────────────────────────────────────────────────────────────────┐
│ ✅ VERIFICATION                                                 │
├─────────────────────────────────────────────────────────────────┤
│ □ Code compiles/runs    □ Tests pass    □ No regressions       │
│ [Show actual proof - command output, test results, etc.]       │
└─────────────────────────────────────────────────────────────────┘
```

**NEVER claim "done" without showing verification proof.**

---

# AGENTS

## Auto-Selection by @prime

| Project Type | Primary Agent | Support Agents |
|--------------|---------------|----------------|
| Svelte/SvelteKit | @frontend-svelte | @tailwind-expert |
| React/Next.js | @frontend-react | @tailwind-expert |
| Go | @backend-go | @database-specialist |
| Node.js | @backend-node | @typescript-expert |

## Task-Based Activation

| Keywords in Request | Agent Added |
|--------------------|-------------|
| bug, error, fix, broken | @debugger |
| test, coverage, spec | @test-automator |
| review, check, pr | @code-reviewer |
| security, auth, vuln | @security-auditor |
| deploy, docker, ci/cd | @devops-engineer |
| refactor, clean, improve | @refactorer |
| docs, readme, document | @technical-writer |
| architecture, design, adr | @architect |

## All Agents

**Orchestration (Opus)**
- @master-orchestrator - Complex multi-agent coordination
- @architect - System design, ADRs, technical decisions

**Specialists (Sonnet)**
- @frontend-svelte - Svelte/SvelteKit expert
- @frontend-react - React/Next.js expert
- @backend-go - Go expert
- @backend-node - Node.js/TypeScript expert
- @database-specialist - PostgreSQL, Redis, SQL
- @api-designer - REST, GraphQL, OpenAPI
- @code-reviewer - Code quality review
- @security-auditor - Security analysis
- @test-automator - Testing, coverage
- @debugger - Bug investigation
- @devops-engineer - Docker, CI/CD, deployment
- @performance-optimizer - Performance tuning
- @refactorer - Code improvement
- @technical-writer - Documentation

**Utilities (Haiku)**
- @explorer - Codebase navigation

---

# SKILLS (Auto-Trigger)

| Skill | Triggers When | What It Does |
|-------|---------------|--------------|
| brainstorming | New feature request | 5W+H, edge cases, approach options |
| test-driven-development | Writing new code | Tests first, red-green-refactor |
| systematic-debugging | Bug investigation | Reproduce, isolate, 5 Whys, fix, verify |
| verification-before-completion | Before "done" | Require proof of working code |
| code-review-checklist | Code review | Security, performance, quality checks |
| pr-review | PR review request | Multi-agent review, categorized issues |
| architecture-decision | Design decisions | Options, trade-offs, ADR, save to memory |
| parallel-dispatch | 3+ independent subtasks | Dispatch to agents in parallel |

---

# PR REVIEW WORKFLOW

## Start Review
Say: `"Review my changes"` or `"Review staged files"` or `"Check src/auth/"`

## Review Output
```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 PR REVIEW: [description]                                     │
├─────────────────────────────────────────────────────────────────┤
│ Files: X changed | Lines: +XXX / -XXX                          │
├─────────────────────────────────────────────────────────────────┤
│ 🔴 CRITICAL [X]: Must fix before merge                         │
│    1. [Issue] - file:line                                      │
├─────────────────────────────────────────────────────────────────┤
│ 🟠 HIGH [X]: Should fix                                        │
│    2. [Issue] - file:line                                      │
├─────────────────────────────────────────────────────────────────┤
│ 🟡 MEDIUM [X]: Consider fixing                                 │
│    3. [Issue] - file:line                                      │
├─────────────────────────────────────────────────────────────────┤
│ 🟢 SUGGESTIONS [X]: Optional improvements                      │
│    4. [Suggestion] - file:line                                 │
├─────────────────────────────────────────────────────────────────┤
│ ✅ GOOD: [What's done well]                                     │
└─────────────────────────────────────────────────────────────────┘

Reply: "fix all" | "fix critical" | "fix 1,2,3" | "explain 2"
```

## Fix Commands
| Command | Action |
|---------|--------|
| `fix all` | Auto-fix all issues |
| `fix critical` | Fix only 🔴 issues |
| `fix 1,2,3` | Fix specific issues |
| `explain 2` | Explain issue #2 in detail |

---

# COMMANDS

## Essential
| Command | Description |
|---------|-------------|
| `@prime` | Initialize context (RUN FIRST) |
| `/review` | Start code review |
| `@agent-name` | Direct agent invocation |

## TaskMaster
| Command | Description |
|---------|-------------|
| `/tm:list` | List all tasks |
| `/tm:add [desc]` | Add new task |
| `/tm:done [id]` | Complete task |
| `/tm:next` | Show next priority task |
| `/tm:update [id] [field] [value]` | Update task |
| `/tm:priority [id] [level]` | Change priority |
| `/tm:block [id] [reason]` | Mark blocked |
| `/tm:unblock [id]` | Remove blocked status |
| `/tm:subtask [parent] [desc]` | Add subtask |
| `/tm:search [query]` | Search tasks |

## Memory
| Command | Description |
|---------|-------------|
| `/mem:search [query]` | Search past decisions |
| `/mem:save` | Save current decision/pattern |
| `/mem:list [collection]` | List memory entries |
| `/mem:context` | Save conversation context |
| `/mem:recall [context]` | Recall by context |

---

# MEMORY SYSTEM

## ChromaDB Collections
| Collection | Content |
|------------|---------|
| decisions | Architectural and design decisions |
| patterns | Reusable code patterns |
| problems | Problems and their solutions |
| context | Project-specific knowledge |
| episodes | Conversation episodes |

## Auto-Behavior
- @prime queries memory for project context
- Important decisions → prompt to save
- Patterns discovered → save for reuse

---

# WORKFLOW

```
┌─────────────────────────────────────────────────────────────────┐
│                        SESSION FLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. START SESSION                                               │
│     └─▶ @prime                                                  │
│                                                                 │
│  2. CONTEXT CARD DISPLAYED                                      │
│     └─▶ Project, agents, memory, tasks loaded                   │
│                                                                 │
│  3. WORK                                                        │
│     └─▶ "Fix the login bug"                                     │
│     └─▶ "Add user notifications"                                │
│     └─▶ "Review my changes"                                     │
│                                                                 │
│  4. VERIFY                                                      │
│     └─▶ Show proof before "done"                                │
│                                                                 │
│  5. SAVE                                                        │
│     └─▶ /mem:save (decisions, patterns)                         │
│                                                                 │
│  6. END or CONTINUE                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

# QUICK REFERENCE

```
START:     @prime
REVIEW:    "Review my changes" → "fix all"
TASKS:     /tm:list → /tm:next → /tm:done [id]
MEMORY:    /mem:search [query] → /mem:save
AGENTS:    @debugger @code-reviewer @test-automator

AGENT AUTO-SELECT:
  .svelte file    → @frontend-svelte
  .go file        → @backend-go
  "bug" keyword   → @debugger
  "test" keyword  → @test-automator
  "review" keyword→ @code-reviewer

NEVER SAY "DONE" WITHOUT VERIFICATION PROOF
```

---

# PROJECT CONVENTIONS

## Svelte/SvelteKit
- Use Svelte stores for shared state
- Use form actions for mutations
- Use +page.server.ts for data loading
- Use +page.ts for client-side data
- Follow existing component patterns

## Go
- Context propagation everywhere
- Structured logging with slog
- Proper error handling (no panic)
- Handler → Service → Repository layers
- Graceful shutdown handling

## React/Next.js
- Server Components by default
- 'use client' only when needed
- Zustand for client state
- shadcn/ui for components
- App Router patterns

---

# VERIFICATION CHECKLIST

Before marking ANY task complete:

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] No regressions introduced
- [ ] Edge cases handled
- [ ] Error handling in place
- [ ] **SHOW OUTPUT/PROOF**

```
✅ VERIFIED:
$ npm run build
✓ Compiled successfully

$ npm test
✓ 47 tests passed

$ git diff --stat
 3 files changed, 45 insertions(+), 12 deletions(-)
```

---

**@prime FIRST. VERIFY ALWAYS. SAVE TO MEMORY.**
