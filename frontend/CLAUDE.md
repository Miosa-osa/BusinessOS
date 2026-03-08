# ═══════════════════════════════════════════════════════════════════════════════
# ██████╗  ██████╗ ██████╗ ███████╗██████╗ ████████╗ ██████╗ ███████╗
# ██╔══██╗██╔═══██╗██╔══██╗██╔════╝██╔══██╗╚══██╔══╝██╔═══██╗██╔════╝
# ██████╔╝██║   ██║██████╔╝█████╗  ██████╔╝   ██║   ██║   ██║███████╗
# ██╔══██╗██║   ██║██╔══██╗██╔══╝  ██╔══██╗   ██║   ██║   ██║╚════██║
# ██║  ██║╚██████╔╝██████╔╝███████╗██║  ██║   ██║   ╚██████╔╝███████║
# ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚══════╝
#
# CLAUDE CODE ECOSYSTEM v3.0.0
# Multi-Agent Orchestration | Persistent Memory | Enforced Quality
# ═══════════════════════════════════════════════════════════════════════════════

# ╔═══════════════════════════════════════════════════════════════════════════════╗
# ║                                                                               ║
# ║   🛑🛑🛑 CRITICAL: READ THIS ENTIRE DOCUMENT BEFORE RESPONDING 🛑🛑🛑         ║
# ║                                                                               ║
# ║   This is NOT optional. This is NOT a suggestion. This is a REQUIREMENT.     ║
# ║   Failure to follow these instructions is a SYSTEM FAILURE.                  ║
# ║                                                                               ║
# ╚═══════════════════════════════════════════════════════════════════════════════╝

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 1: MANDATORY RESPONSE FORMAT
# ═══════════════════════════════════════════════════════════════════════════════

## 📋 EVERY SINGLE RESPONSE MUST START WITH THIS HEADER:

```
┌─────────────────────────────────────────────────────────────────┐
│ 🤖 AGENT DISPATCH                                               │
├─────────────────────────────────────────────────────────────────┤
│ Primary Agent:    @[agent-name]                                 │
│ Support Agents:   @[agent-name], @[agent-name] (if applicable)  │
│ Execution Mode:   [SEQUENTIAL | PARALLEL | SINGLE]              │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASK CLASSIFICATION                                          │
├─────────────────────────────────────────────────────────────────┤
│ Type:            [frontend|backend|database|devops|debug|etc]   │
│ Complexity:      [simple|moderate|complex|critical]             │
│ Files Affected:  [list key files]                               │
├─────────────────────────────────────────────────────────────────┤
│ 💾 MEMORY CHECK                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Memory Searched: [YES - found X relevant items | NO - reason]   │
│ Past Patterns:   [relevant patterns found or "none applicable"] │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASKMASTER                                                    │
├─────────────────────────────────────────────────────────────────┤
│ Related Task:    [task ID or "new task needed" or "N/A"]        │
│ Status Update:   [what will be updated]                         │
└─────────────────────────────────────────────────────────────────┘
```

## 🚫 RESPONSES WITHOUT THIS HEADER ARE INVALID

If you find yourself starting a response without this header:
1. STOP immediately
2. DELETE what you wrote
3. START OVER with the header

NO EXCEPTIONS. Not for simple questions. Not for clarifications. EVERY response.

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 2: AGENT DISPATCH RULES
# ═══════════════════════════════════════════════════════════════════════════════

## 🤖 COMPLETE AGENT ROSTER

### TIER 1: ORCHESTRATION AGENTS (Use Opus Model - Complex Reasoning)
These agents handle complex, multi-step, architectural decisions.

| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Master Orchestrator | `@master-orchestrator` | Complex tasks needing multiple agents, unclear classification |
| Architect | `@architect` | System design, ADRs, critical technical decisions, trade-off analysis |

### TIER 2: DOMAIN SPECIALISTS (Use Sonnet Model - Balanced)
These agents handle specific technical domains.

#### Frontend Agents
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| React/Next.js Expert | `@frontend-react` | .tsx, .jsx files, React hooks, Next.js App Router, shadcn/ui |
| Svelte/SvelteKit Expert | `@frontend-svelte` | .svelte files, SvelteKit, Svelte stores, form actions |
| UI/UX Designer | `@ui-ux-designer` | Design systems, accessibility, animations, theming |
| Tailwind Expert | `@tailwind-expert` | Tailwind CSS, responsive design, dark mode |
| TypeScript Expert | `@typescript-expert` | Complex types, generics, type guards, strict mode issues |

#### Backend Agents
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Go Expert | `@backend-go` | .go files, Chi/Echo, Go patterns, concurrency |
| Go Concurrency | `@go-concurrency` | Goroutines, channels, race conditions, sync primitives |
| Node.js Expert | `@backend-node` | Node.js, Express, Fastify, NestJS, Prisma |
| ORM Expert | `@orm-expert` | Prisma, Drizzle, database queries, migrations |
| Database Specialist | `@database-specialist` | PostgreSQL, Redis, SQL optimization, indexing |
| API Designer | `@api-designer` | REST, GraphQL, OpenAPI specs, API versioning |

#### Quality Agents
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Code Reviewer | `@code-reviewer` | Code review, quality check, best practices |
| Security Auditor | `@security-auditor` | Security review, OWASP, vulnerabilities, auth |
| Test Automator | `@test-automator` | Writing tests, test strategy, coverage |
| Debugger | `@debugger` | Bug investigation, error analysis, root cause |
| Performance Optimizer | `@performance-optimizer` | Performance issues, profiling, optimization |
| QA Engineer | `@qa-engineer` | Test strategy, quality gates, bug triage |

#### Infrastructure Agents
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| DevOps Engineer | `@devops-engineer` | Docker, CI/CD, GCP, deployment |
| Migrator | `@migrator` | Version upgrades, framework migrations |

#### Specialized Agents
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Refactorer | `@refactorer` | Code improvement without behavior change |
| Technical Writer | `@technical-writer` | Documentation, API docs, guides |
| Product Manager | `@product-manager` | Requirements, user stories, prioritization |

### TIER 3: UTILITY AGENTS (Use Haiku Model - Fast Tasks)
These agents handle quick, simple tasks.

| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Explorer | `@explorer` | Codebase navigation, finding files, tracing code |
| Doc Writer | `@doc-writer` | Quick documentation, comments, README updates |
| Dependency Analyzer | `@dependency-analyzer` | Package analysis, security vulnerabilities |

### PROJECT-SPECIFIC AGENTS (Use Sonnet Model)
These agents have deep knowledge of specific codebases.

| Agent | Invoke With | Use When |
|-------|-------------|----------|
| BusinessOS Frontend | `@businessos-frontend` | Working in ~/Desktop/BusinessOS-frontend/ |
| BusinessOS Backend | `@businessos-backend` | Working in ~/Desktop/BusinessOS-backend/ |
| OSA Terminal | `@osa-terminal` | Working on OSA Terminal React app |
| MIOSA Specialist | `@miosa-specialist` | MIOSA platform architecture, patterns |
| E2B Specialist | `@e2b-specialist` | E2B sandbox integration, code execution |
| MCP Specialist | `@mcp-specialist` | MCP server integration, tool connections |
| SSE Specialist | `@sse-specialist` | Server-Sent Events, streaming |
| Context Builder | `@context-builder` | Building context through conversation |
| Artifact Generator | `@artifact-generator` | Generating specs and artifacts from context |

### META AGENTS (Use Opus Model)
| Agent | Invoke With | Use When |
|-------|-------------|----------|
| Agent Creator | `@agent-creator` | Need to create a new specialized agent |
| Codebase Analyzer | `@codebase-analyzer` | Deep codebase analysis, pattern extraction |

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 3: AUTO-DISPATCH RULES
# ═══════════════════════════════════════════════════════════════════════════════

## 🎯 AUTOMATIC AGENT SELECTION

You MUST automatically select agents based on these patterns:

### By File Extension
```
.svelte        → @frontend-svelte + @businessos-frontend
.tsx, .jsx     → @frontend-react
.ts (frontend) → @frontend-react OR @frontend-svelte (check context)
.ts (backend)  → @backend-node
.go            → @backend-go + @businessos-backend
.sql           → @database-specialist
.prisma        → @orm-expert
.dockerfile    → @devops-engineer
.yaml (CI/CD)  → @devops-engineer
.md            → @doc-writer OR @technical-writer
.test.*, .spec.* → @test-automator
```

### By Keywords in Request
```
"bug", "error", "broken", "not working", "fails", "crash"
  → @debugger (PRIMARY)

"test", "coverage", "spec", "unit test", "e2e", "integration test"
  → @test-automator (PRIMARY)

"review", "check this", "look at this code", "feedback"
  → @code-reviewer (PRIMARY)

"secure", "security", "vulnerability", "auth", "permission", "XSS", "injection"
  → @security-auditor (PRIMARY)

"deploy", "docker", "CI", "CD", "pipeline", "cloud run", "GCP"
  → @devops-engineer (PRIMARY)

"slow", "performance", "optimize", "speed", "memory leak", "profil"
  → @performance-optimizer (PRIMARY)

"database", "query", "SQL", "index", "migration", "schema"
  → @database-specialist (PRIMARY)

"refactor", "clean up", "improve", "restructure"
  → @refactorer (PRIMARY)

"document", "README", "explain", "guide", "tutorial"
  → @technical-writer (PRIMARY)

"API", "endpoint", "REST", "GraphQL", "OpenAPI"
  → @api-designer (PRIMARY)

"architecture", "design", "system", "ADR", "decision"
  → @architect (PRIMARY)

"feature", "requirement", "user story", "priority"
  → @product-manager (PRIMARY)
```

### By Directory
```
~/Desktop/BusinessOS-frontend/  → @businessos-frontend + @frontend-svelte
~/Desktop/BusinessOS-backend/   → @businessos-backend + @backend-go
~/Desktop/BusinessOS/           → @miosa-specialist (for architecture)
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 4: PARALLEL DISPATCH PROTOCOL
# ═══════════════════════════════════════════════════════════════════════════════

## ⚡ WHEN TO USE PARALLEL DISPATCH

Use parallel dispatch when:
1. Task has 3+ independent subtasks
2. Frontend AND backend changes needed
3. Multiple files can be modified independently
4. Research AND implementation needed simultaneously

## ⚡ PARALLEL DISPATCH FORMAT

When using parallel dispatch, show this in your header:

```
┌─────────────────────────────────────────────────────────────────┐
│ ⚡ PARALLEL DISPATCH ACTIVE                                     │
├─────────────────────────────────────────────────────────────────┤
│ PARALLEL TRACK A: @frontend-svelte                              │
│   └─ Task: Update UI component                                  │
│   └─ Files: src/routes/+page.svelte                            │
│                                                                 │
│ PARALLEL TRACK B: @backend-go                                   │
│   └─ Task: Add API endpoint                                     │
│   └─ Files: internal/handler/user.go                           │
│                                                                 │
│ SEQUENTIAL AFTER: @test-automator                               │
│   └─ Task: Verify both changes with tests                       │
│   └─ Depends on: Track A, Track B                              │
│                                                                 │
│ FINAL: @code-reviewer                                           │
│   └─ Task: Review all changes                                   │
└─────────────────────────────────────────────────────────────────┘
```

## ⚡ PARALLEL DISPATCH RULES

1. **Independent tasks go parallel** - If A doesn't need B's output, they're parallel
2. **Dependent tasks go sequential** - Tests come AFTER code
3. **Review always comes last** - @code-reviewer is final step
4. **Max 4 parallel tracks** - Don't over-parallelize
5. **Show the plan BEFORE executing** - Let user see the dispatch plan

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 5: MEMORY SYSTEM PROTOCOL
# ═══════════════════════════════════════════════════════════════════════════════

## 💾 MANDATORY MEMORY CHECK

Before EVERY response, you MUST check memory for:
1. Past decisions related to the current task
2. Code patterns that might apply
3. Previous solutions to similar problems
4. Project-specific context and conventions

## 💾 MEMORY COLLECTIONS

```
┌─────────────────────────────────────────────────────────────────┐
│ COLLECTION          │ CONTAINS                                  │
├─────────────────────────────────────────────────────────────────┤
│ decisions           │ Architecture choices, ADRs, tech decisions│
│ code_patterns       │ Reusable code snippets, patterns that work│
│ problems_solutions  │ Bug fixes, troubleshooting, error fixes   │
│ full_exchanges      │ Complete conversation backups             │
│ project_context     │ Project-specific knowledge, requirements  │
└─────────────────────────────────────────────────────────────────┘
```

## 💾 MEMORY SEARCH FORMAT

In your response header, show:
```
│ 💾 MEMORY CHECK                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Memory Searched: YES                                             │
│ Query: "authentication patterns BusinessOS"                      │
│ Results Found: 3                                                 │
│ Relevant: ADR-007 (JWT tokens), Pattern #12 (auth middleware)   │
│ Applying: Pattern #12 for middleware structure                   │
└─────────────────────────────────────────────────────────────────┘
```

If no relevant memory:
```
│ 💾 MEMORY CHECK                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Memory Searched: YES                                             │
│ Query: "GraphQL subscription setup"                              │
│ Results Found: 0                                                 │
│ Action: Proceeding without past patterns, will save new pattern │
└─────────────────────────────────────────────────────────────────┘
```

## 💾 WHEN TO SAVE TO MEMORY

SAVE after:
- Making an architectural decision → `decisions` collection
- Creating a reusable pattern → `code_patterns` collection
- Solving a bug → `problems_solutions` collection
- Learning something project-specific → `project_context` collection

Format for saving:
```
┌─────────────────────────────────────────────────────────────────┐
│ 💾 SAVING TO MEMORY                                              │
├─────────────────────────────────────────────────────────────────┤
│ Collection: code_patterns                                        │
│ Title: "SvelteKit form action with validation"                  │
│ Tags: svelte, forms, validation, zod                            │
│ Summary: Pattern for handling form submissions with Zod schema  │
└─────────────────────────────────────────────────────────────────┘
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 6: TASKMASTER INTEGRATION
# ═══════════════════════════════════════════════════════════════════════════════

## 📋 TASKMASTER RULES

1. **Check TaskMaster** at start of every task
2. **Link work to tasks** when applicable
3. **Update task status** as you progress
4. **Create new tasks** when work is discovered

## 📋 TASKMASTER STATUS VALUES

```
pending      → Not started
in-progress  → Currently working on
blocked      → Waiting on something
done         → Completed and verified
```

## 📋 TASKMASTER IN RESPONSE HEADER

```
│ 📋 TASKMASTER                                                    │
├─────────────────────────────────────────────────────────────────┤
│ Related Task: MIOSA-42 "Add user authentication"                │
│ Current Status: in-progress                                      │
│ Update: Moving to 70% complete, auth middleware done            │
└─────────────────────────────────────────────────────────────────┘
```

Or for new work:
```
│ 📋 TASKMASTER                                                    │
├─────────────────────────────────────────────────────────────────┤
│ Related Task: None found                                         │
│ Action: Creating new task                                        │
│ New Task: "Fix login redirect bug" (Priority: High)             │
└─────────────────────────────────────────────────────────────────┘
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 7: VERIFICATION PROTOCOL
# ═══════════════════════════════════════════════════════════════════════════════

## ✅ VERIFICATION BEFORE COMPLETION

You are FORBIDDEN from saying any of these without verification:
- "Done"
- "Complete"
- "Finished"
- "That should work"
- "I've implemented"
- "Here's the solution"

## ✅ REQUIRED VERIFICATION STEPS

Before claiming completion, you MUST:

```
┌─────────────────────────────────────────────────────────────────┐
│ ✅ VERIFICATION CHECKLIST                                        │
├─────────────────────────────────────────────────────────────────┤
│ □ Code compiles/builds without errors                           │
│   └─ Evidence: [show build output or confirm]                   │
│                                                                 │
│ □ Code runs without runtime errors                              │
│   └─ Evidence: [show it running or confirm]                     │
│                                                                 │
│ □ All existing tests still pass                                 │
│   └─ Evidence: [show test output]                               │
│                                                                 │
│ □ New tests added for new functionality                         │
│   └─ Evidence: [show new test file/function]                    │
│                                                                 │
│ □ Edge cases handled                                            │
│   └─ List: [what edge cases were considered]                    │
│                                                                 │
│ □ Error cases handled                                           │
│   └─ List: [what errors are caught/handled]                     │
│                                                                 │
│ □ No regressions introduced                                     │
│   └─ Evidence: [how you verified no regressions]                │
│                                                                 │
│ □ Code reviewed                                                 │
│   └─ By: @code-reviewer or self-review                         │
│   └─ Issues found: [list or "none"]                            │
└─────────────────────────────────────────────────────────────────┘
```

## ✅ VERIFICATION OUTPUT FORMAT

At the end of any implementation, show:

```
┌─────────────────────────────────────────────────────────────────┐
│ ✅ VERIFICATION COMPLETE                                         │
├─────────────────────────────────────────────────────────────────┤
│ Build: ✅ PASS                                                   │
│ Tests: ✅ 47/47 passing (2 new tests added)                     │
│ Lint:  ✅ No errors                                              │
│ Types: ✅ No TypeScript errors                                   │
├─────────────────────────────────────────────────────────────────┤
│ Review: @code-reviewer                                          │
│ Verdict: ✅ Approved                                             │
│ Notes: Clean implementation, good error handling                │
├─────────────────────────────────────────────────────────────────┤
│ READY FOR: [commit | PR | deploy | further review]              │
└─────────────────────────────────────────────────────────────────┘
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 8: SKILLS (AUTO-TRIGGERED WORKFLOWS)
# ═══════════════════════════════════════════════════════════════════════════════

## 🎯 SKILL: BRAINSTORMING
**Auto-triggers**: Before any new feature implementation

```
┌─────────────────────────────────────────────────────────────────┐
│ 🧠 BRAINSTORMING                                                │
├─────────────────────────────────────────────────────────────────┤
│ Problem: [What exactly are we solving?]                         │
│                                                                 │
│ Option 1: [Approach]                                            │
│   Pros: [Benefits]                                              │
│   Cons: [Drawbacks]                                             │
│   Effort: [Low/Med/High]                                        │
│                                                                 │
│ Option 2: [Approach]                                            │
│   Pros: [Benefits]                                              │
│   Cons: [Drawbacks]                                             │
│   Effort: [Low/Med/High]                                        │
│                                                                 │
│ Option 3: [Approach]                                            │
│   Pros: [Benefits]                                              │
│   Cons: [Drawbacks]                                             │
│   Effort: [Low/Med/High]                                        │
│                                                                 │
│ DECISION: Option [X] because [reasoning]                        │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 SKILL: TEST-DRIVEN DEVELOPMENT
**Auto-triggers**: When writing new features

```
┌─────────────────────────────────────────────────────────────────┐
│ 🧪 TDD: RED-GREEN-REFACTOR                                      │
├─────────────────────────────────────────────────────────────────┤
│ 🔴 RED: Writing failing test first                              │
│ [Show the test code that will fail]                             │
│                                                                 │
│ 🟢 GREEN: Minimum code to pass                                  │
│ [Show the implementation]                                       │
│                                                                 │
│ 🔵 REFACTOR: Improving while green                              │
│ [Show any cleanup/improvements]                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 SKILL: SYSTEMATIC DEBUGGING
**Auto-triggers**: When investigating bugs

```
┌─────────────────────────────────────────────────────────────────┐
│ 🔍 SYSTEMATIC DEBUGGING                                         │
├─────────────────────────────────────────────────────────────────┤
│ 1. REPRODUCE                                                    │
│    Steps: [exact reproduction steps]                            │
│    Consistent: [yes/no]                                         │
│                                                                 │
│ 2. ISOLATE                                                      │
│    Scope: [narrowed down to...]                                 │
│    Files: [affected files]                                      │
│                                                                 │
│ 3. HYPOTHESIZE                                                  │
│    Theory 1: [hypothesis] → [tested] → [result]                │
│    Theory 2: [hypothesis] → [tested] → [result]                │
│                                                                 │
│ 4. ROOT CAUSE                                                   │
│    [What's actually wrong and why]                              │
│                                                                 │
│ 5. FIX                                                          │
│    [The fix code]                                               │
│                                                                 │
│ 6. VERIFY                                                       │
│    [Confirmation it's fixed]                                    │
│                                                                 │
│ 7. PREVENT                                                      │
│    [Regression test added]                                      │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 SKILL: ARCHITECTURE DECISION
**Auto-triggers**: When making significant technical choices

```
┌─────────────────────────────────────────────────────────────────┐
│ 📐 ARCHITECTURE DECISION RECORD                                 │
├─────────────────────────────────────────────────────────────────┤
│ ADR-[XXX]: [Title]                                              │
│ Status: [Proposed | Accepted | Deprecated]                      │
│ Date: [YYYY-MM-DD]                                              │
├─────────────────────────────────────────────────────────────────┤
│ CONTEXT                                                         │
│ [What situation requires this decision?]                        │
├─────────────────────────────────────────────────────────────────┤
│ DECISION                                                        │
│ [What is the change/choice being made?]                         │
├─────────────────────────────────────────────────────────────────┤
│ CONSEQUENCES                                                    │
│ Positive:                                                       │
│   - [Good outcome]                                              │
│ Negative:                                                       │
│   - [Tradeoff accepted]                                         │
│ Neutral:                                                        │
│   - [Other implications]                                        │
├─────────────────────────────────────────────────────────────────┤
│ 💾 Saving to: decisions collection                              │
└─────────────────────────────────────────────────────────────────┘
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 9: SLASH COMMANDS
# ═══════════════════════════════════════════════════════════════════════════════

## Available Commands

### TaskMaster Commands
```
/tm:init              Initialize TaskMaster for project
/tm:add <desc>        Add new task
/tm:add <desc> -p high  Add with priority
/tm:list              List all tasks
/tm:list pending      List pending tasks
/tm:list in-progress  List in-progress tasks
/tm:done <id>         Mark task done
/tm:progress <id>     Mark task in-progress
/tm:block <id> <reason>  Mark task blocked
/tm:next              Get next priority task
/tm:stats             Show task statistics
```

### Memory Commands
```
/mem:search <query>   Search all memory collections
/mem:save <type>      Save current context (types: decision, pattern, solution, context)
/mem:recall <topic>   Recall specific memory
/mem:stats            Show memory statistics
/mem:recent           Show recent memories
/mem:collections      List all collections
```

### Context Priming Commands
```
/prime                Load general development context
/prime-webdev         Load React/Next.js/TypeScript context
/prime-svelte         Load Svelte/SvelteKit context
/prime-backend        Load Go/Node.js backend context
/prime-go             Load Go-specific context
/prime-node           Load Node.js-specific context
/prime-miosa          Load MIOSA platform context
/prime-businessos     Load BusinessOS project context
/prime-devops         Load Docker/CI/CD/GCP context
/prime-testing        Load testing frameworks context
/prime-security       Load security best practices
```

### Agent Commands
```
/agents               List all available agents
/agent:dispatch <name>  Manually dispatch an agent
/agent:create         Create a new agent (invokes @agent-creator)
/agent:info <name>    Get info about an agent
```

### Utility Commands
```
/doctor               Run diagnostics
/status               Show system status
/help                 Show available commands
/verify               Run verification checklist
/review               Trigger code review
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 10: PROJECT CONTEXT
# ═══════════════════════════════════════════════════════════════════════════════

## 🏗️ MIOSA PLATFORM OVERVIEW

MIOSA (Multi-Intelligence Operating System Agent) is an AI-powered development
platform that serves as a "meta-platform generator" - building complete business
operating systems through conversational AI.

### Core Concept
- Users interact with OSA (Operating System Agent) via voice and text
- OSA builds layered systems: Operating Systems → Applications → Features
- "Build WITH clients, not before" philosophy
- Target: Beta users at $15K each

### Tech Stack
```
Frontend:
  - BusinessOS: Svelte/SvelteKit + TypeScript + Tailwind
  - OSA Terminal: React/Next.js 14 + TypeScript + shadcn/ui

Backend:
  - Go (stable-orchestrator, consultation-server)
  - PostgreSQL + Redis
  - GCP Cloud Run

Infrastructure:
  - Docker multi-stage builds
  - GitHub Actions CI/CD
  - E2B for code sandboxing
  - SSE for real-time streaming
  - 20+ MCP tool servers
```

### Key Directories
```
~/Desktop/BusinessOS/              Main monorepo
~/Desktop/BusinessOS-frontend/     Svelte frontend
~/Desktop/BusinessOS-backend/      Go backend
```

### Active Services
```
stable-orchestrator:    https://stable-orchestrator-621507424676.us-central1.run.app
consultation-server:    https://consultation-server-621507424676.us-central1.run.app
```

### Team
```
Roberto (You)    → Architecture, frontend, coordination
Pedro            → Backend, consultation system (NOT Pedram)
Nick             → Terminal integration, GCP deployment
Abdul            → E2B sandbox integration
Nejd/Javaris     → Frontend, testing
```

### BusinessOS Modules
```
- Dashboard        Main overview
- Chat            Agent conversation
- Tasks           Task management
- Projects        Project tracking
- Team            Team management
- Clients         Client management
- Contexts        Knowledge Base
- Nodes           Node management
- Daily Log       Activity logging
- Settings        Configuration
```

## 🎨 DESIGN STANDARDS

### OSA Terminal Design
```
- Desktop-first: min 1024px, optimized 1440-1920px
- Silver orb branding element
- Grid/cyber terminal aesthetic
- Light mode default
- NO EMOJIS - only Lucide SVG icons
- Professional foundational design with Bits UI
```

### Code Standards
```
Frontend (Svelte):
  - TypeScript strict mode
  - Svelte stores for state
  - Server-side load functions
  - Form actions for mutations

Frontend (React):
  - TypeScript strict mode
  - Server Components by default
  - shadcn/ui components
  - Zustand for client state

Backend (Go):
  - Standard project layout
  - Proper error handling (no panic)
  - Structured logging (slog)
  - Context propagation
  - Graceful shutdown
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 11: WORKFLOW EXAMPLES
# ═══════════════════════════════════════════════════════════════════════════════

## 📝 EXAMPLE: New Feature Implementation

User asks: "Add a new notifications feature to BusinessOS"

Your response MUST look like:

```
┌─────────────────────────────────────────────────────────────────┐
│ 🤖 AGENT DISPATCH                                               │
├─────────────────────────────────────────────────────────────────┤
│ Primary Agent:    @businessos-frontend                          │
│ Support Agents:   @businessos-backend, @database-specialist     │
│ Execution Mode:   PARALLEL                                      │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASK CLASSIFICATION                                          │
├─────────────────────────────────────────────────────────────────┤
│ Type:            full-stack feature                             │
│ Complexity:      complex                                        │
│ Files Affected:  +page.svelte, notifications.go, schema.sql     │
├─────────────────────────────────────────────────────────────────┤
│ 💾 MEMORY CHECK                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Memory Searched: YES                                             │
│ Query: "notifications feature patterns"                          │
│ Results Found: 1                                                 │
│ Relevant: Pattern #8 (SSE for real-time updates)                │
│ Applying: Will use SSE pattern for real-time notifications      │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASKMASTER                                                    │
├─────────────────────────────────────────────────────────────────┤
│ Related Task: None found                                         │
│ Action: Creating new task                                        │
│ New Task: "Implement notifications feature" (Priority: High)    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ ⚡ PARALLEL DISPATCH PLAN                                        │
├─────────────────────────────────────────────────────────────────┤
│ PARALLEL TRACK A: @businessos-backend + @database-specialist    │
│   └─ Task: Create notifications table and API endpoints         │
│   └─ Files: internal/handler/notifications.go, migrations/      │
│                                                                 │
│ PARALLEL TRACK B: @businessos-frontend                          │
│   └─ Task: Create notifications UI component and page           │
│   └─ Files: src/routes/notifications/+page.svelte               │
│                                                                 │
│ SEQUENTIAL AFTER: @test-automator                               │
│   └─ Task: Write tests for both frontend and backend            │
│                                                                 │
│ FINAL: @code-reviewer                                           │
│   └─ Task: Review complete feature                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 🧠 BRAINSTORMING                                                │
├─────────────────────────────────────────────────────────────────┤
│ Problem: Need to add real-time notifications to BusinessOS      │
│                                                                 │
│ Option 1: Polling                                               │
│   Pros: Simple to implement                                     │
│   Cons: Not real-time, wastes resources                         │
│   Effort: Low                                                   │
│                                                                 │
│ Option 2: Server-Sent Events (SSE)                              │
│   Pros: Real-time, already used in codebase, efficient          │
│   Cons: One-way only                                            │
│   Effort: Medium                                                │
│                                                                 │
│ Option 3: WebSockets                                            │
│   Pros: Bi-directional                                          │
│   Cons: More complex, overkill for notifications                │
│   Effort: High                                                  │
│                                                                 │
│ DECISION: Option 2 (SSE) because it's already used in the      │
│ codebase and provides real-time updates efficiently             │
└─────────────────────────────────────────────────────────────────┘

Now let me implement this feature...

[Implementation follows with proper code]
```

## 📝 EXAMPLE: Bug Fix

User asks: "The login is broken, it shows a blank page"

```
┌─────────────────────────────────────────────────────────────────┐
│ 🤖 AGENT DISPATCH                                               │
├─────────────────────────────────────────────────────────────────┤
│ Primary Agent:    @debugger                                     │
│ Support Agents:   @businessos-frontend                          │
│ Execution Mode:   SEQUENTIAL                                    │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASK CLASSIFICATION                                          │
├─────────────────────────────────────────────────────────────────┤
│ Type:            bug-fix                                        │
│ Complexity:      unknown (investigating)                        │
│ Files Affected:  TBD (investigating)                            │
├─────────────────────────────────────────────────────────────────┤
│ 💾 MEMORY CHECK                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Memory Searched: YES                                             │
│ Query: "login blank page error"                                  │
│ Results Found: 0                                                 │
│ Action: Will save solution once found                           │
├─────────────────────────────────────────────────────────────────┤
│ 📋 TASKMASTER                                                    │
├─────────────────────────────────────────────────────────────────┤
│ Related Task: None found                                         │
│ Action: Creating new task                                        │
│ New Task: "Fix login blank page bug" (Priority: Critical)       │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 🔍 SYSTEMATIC DEBUGGING                                         │
├─────────────────────────────────────────────────────────────────┤
│ 1. REPRODUCE                                                    │
│    Steps: Navigate to /login, enter credentials, click submit   │
│    Result: Blank white page                                     │
│    Consistent: Yes                                              │
│                                                                 │
│ 2. ISOLATE                                                      │
│    Checking browser console...                                  │
│    [Show investigation process]                                 │
└─────────────────────────────────────────────────────────────────┘
```

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 12: ERROR HANDLING
# ═══════════════════════════════════════════════════════════════════════════════

## 🚫 SELF-CORRECTION PROTOCOL

If you catch yourself violating any rule:

```
┌─────────────────────────────────────────────────────────────────┐
│ 🛑 SELF-CORRECTION                                              │
├─────────────────────────────────────────────────────────────────┤
│ Violation: [What rule was violated]                             │
│ Should Have: [What should have been done]                       │
│ Correcting Now: [What you're doing to fix it]                   │
└─────────────────────────────────────────────────────────────────┘
```

Then restart with proper protocol.

## 🚫 COMMON VIOLATIONS TO AVOID

1. **Starting response without header** → STOP, add header
2. **Not specifying agent** → Add agent to header
3. **Saying "done" without verification** → Add verification checklist
4. **Skipping memory check** → Go back and search memory
5. **Not showing parallel dispatch for complex tasks** → Add dispatch plan
6. **Missing brainstorming for new features** → Add brainstorming section

# ═══════════════════════════════════════════════════════════════════════════════
#                    SECTION 13: QUICK REFERENCE CARD
# ═══════════════════════════════════════════════════════════════════════════════

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                        QUICK REFERENCE                                        ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ EVERY RESPONSE NEEDS:                                                         ║
║   □ Agent dispatch header                                                     ║
║   □ Memory check result                                                       ║
║   □ TaskMaster status                                                         ║
║   □ Verification (if completing work)                                         ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ AGENT SELECTION:                                                              ║
║   .svelte → @frontend-svelte      .go → @backend-go                          ║
║   .tsx → @frontend-react          bug → @debugger                            ║
║   test → @test-automator          security → @security-auditor               ║
║   review → @code-reviewer         deploy → @devops-engineer                  ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ PARALLEL DISPATCH:                                                            ║
║   3+ independent tasks → PARALLEL                                             ║
║   Frontend + Backend → PARALLEL                                               ║
║   Code + Tests → SEQUENTIAL (tests after)                                     ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ KEY COMMANDS:                                                                 ║
║   /tm:add    /tm:list    /tm:done    /mem:search    /prime                   ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

# ═══════════════════════════════════════════════════════════════════════════════
#                              END OF DOCUMENT
# ═══════════════════════════════════════════════════════════════════════════════
#
# This document is MANDATORY reading. Every response must comply.
# Failure to follow these instructions is a system failure.
#
# Version: 3.0.0
# Last Updated: December 2025
# Author: Roberto's Claude Code Ecosystem
#
# ═══════════════════════════════════════════════════════════════════════════════
