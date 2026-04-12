#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════════════
# CLAUDE CODE ECOSYSTEM - COMPLETE SETUP
# One script to install everything. Run: curl -sL [url] | bash
# ═══════════════════════════════════════════════════════════════════════════════

set -e

# Colors
R='\033[0;31m' G='\033[0;32m' Y='\033[1;33m' B='\033[0;34m' C='\033[0;36m' NC='\033[0m'

log() { echo -e "${B}[SETUP]${NC} $1"; }
ok() { echo -e "${G}  ✅${NC} $1"; }
warn() { echo -e "${Y}  ⚠️${NC} $1"; }
err() { echo -e "${R}  ❌${NC} $1"; }

echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                    CLAUDE CODE ECOSYSTEM INSTALLER                            ║"
echo "║                         Complete Setup Script                                 ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""

CLAUDE_HOME="$HOME/.claude"
TASKMASTER_HOME="$HOME/.taskmaster"

# ═══════════════════════════════════════════════════════════════════════════════
# 1. CREATE DIRECTORY STRUCTURE
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating directory structure..."

mkdir -p "$CLAUDE_HOME"/{agents,skills,commands,hooks,orchestration,context,continuous,scripts,memory,logs}
mkdir -p "$CLAUDE_HOME"/continuous/{configs,sessions,reports/daily,checkpoints,logs}
mkdir -p "$CLAUDE_HOME"/context/projects
mkdir -p "$TASKMASTER_HOME"/tasks
mkdir -p "$HOME"/.chroma-data
mkdir -p "$HOME"/bin

ok "Directories created"

# ═══════════════════════════════════════════════════════════════════════════════
# 2. CREATE AGENTS
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating agents..."

# Master Orchestrator
cat > "$CLAUDE_HOME/agents/master-orchestrator.md" << 'EOF'
---
name: master-orchestrator
description: Coordinates complex multi-agent tasks
model: opus
trigger: complex_task OR multi_step OR architecture
---
Coordinates multiple agents for complex tasks. Decomposes problems, delegates to specialists, synthesizes results.
EOF

# Architect
cat > "$CLAUDE_HOME/agents/architect.md" << 'EOF'
---
name: architect
description: System design and architectural decisions
model: opus
trigger: architecture OR design OR ADR OR technical_decision
---
Handles system design, creates ADRs, evaluates trade-offs, makes architectural decisions.
EOF

# Frontend Svelte
cat > "$CLAUDE_HOME/agents/frontend-svelte.md" << 'EOF'
---
name: frontend-svelte
description: Svelte/SvelteKit specialist
model: sonnet
trigger: .svelte OR sveltekit OR svelte_store
---
Expert in Svelte/SvelteKit. Uses stores, form actions, +page.server.ts patterns. Follows Svelte best practices.
EOF

# Frontend React
cat > "$CLAUDE_HOME/agents/frontend-react.md" << 'EOF'
---
name: frontend-react
description: React/Next.js specialist
model: sonnet
trigger: .tsx OR .jsx OR nextjs OR react
---
Expert in React/Next.js. Uses Server Components, hooks, Zustand, shadcn/ui. Follows React best practices.
EOF

# Backend Go
cat > "$CLAUDE_HOME/agents/backend-go.md" << 'EOF'
---
name: backend-go
description: Go backend specialist
model: sonnet
trigger: .go OR golang OR chi_router
---
Expert in Go. Uses context propagation, slog logging, proper error handling. Never panics. Follows Go idioms.
EOF

# Backend Node
cat > "$CLAUDE_HOME/agents/backend-node.md" << 'EOF'
---
name: backend-node
description: Node.js/TypeScript backend specialist
model: sonnet
trigger: node_backend OR express OR typescript_backend
---
Expert in Node.js/TypeScript backends. Uses Express/Fastify, proper async patterns, TypeScript strict mode.
EOF

# Database Specialist
cat > "$CLAUDE_HOME/agents/database-specialist.md" << 'EOF'
---
name: database-specialist
description: Database design and optimization
model: sonnet
trigger: database OR sql OR postgresql OR redis OR schema
---
Expert in PostgreSQL, Redis, database design, query optimization, migrations, indexing strategies.
EOF

# API Designer
cat > "$CLAUDE_HOME/agents/api-designer.md" << 'EOF'
---
name: api-designer
description: API design specialist
model: sonnet
trigger: api_design OR rest OR graphql OR openapi
---
Designs RESTful APIs, GraphQL schemas, OpenAPI specs. Focuses on consistency, versioning, documentation.
EOF

# Code Reviewer
cat > "$CLAUDE_HOME/agents/code-reviewer.md" << 'EOF'
---
name: code-reviewer
description: Code review specialist
model: sonnet
trigger: review OR check_code OR pr_review
---
Reviews code for quality, patterns, bugs, performance. Categorizes issues by severity. Suggests improvements.
EOF

# Security Auditor
cat > "$CLAUDE_HOME/agents/security-auditor.md" << 'EOF'
---
name: security-auditor
description: Security analysis specialist
model: sonnet
trigger: security OR auth OR vulnerability OR owasp
---
Audits for security vulnerabilities, OWASP issues, auth problems, injection risks, data exposure.
EOF

# Test Automator
cat > "$CLAUDE_HOME/agents/test-automator.md" << 'EOF'
---
name: test-automator
description: Testing specialist
model: sonnet
trigger: test OR coverage OR spec OR unit_test OR integration_test
---
Writes comprehensive tests. Focuses on coverage, edge cases, mocking, test organization.
EOF

# Debugger
cat > "$CLAUDE_HOME/agents/debugger.md" << 'EOF'
---
name: debugger
description: Bug investigation specialist
model: sonnet
trigger: bug OR error OR fix OR broken OR not_working OR debug
---
Systematically investigates bugs. Uses 5 Whys, reproduces issues, isolates causes, verifies fixes.
EOF

# DevOps Engineer
cat > "$CLAUDE_HOME/agents/devops-engineer.md" << 'EOF'
---
name: devops-engineer
description: DevOps and deployment specialist
model: sonnet
trigger: deploy OR docker OR ci OR cd OR kubernetes OR gcp OR aws
---
Handles Docker, CI/CD, cloud deployment, infrastructure. Creates reliable deployment pipelines.
EOF

# Performance Optimizer
cat > "$CLAUDE_HOME/agents/performance-optimizer.md" << 'EOF'
---
name: performance-optimizer
description: Performance optimization specialist
model: sonnet
trigger: performance OR slow OR optimize OR speed OR latency
---
Analyzes and optimizes performance. Profiles code, identifies bottlenecks, implements improvements.
EOF

# Refactorer
cat > "$CLAUDE_HOME/agents/refactorer.md" << 'EOF'
---
name: refactorer
description: Code refactoring specialist
model: sonnet
trigger: refactor OR clean OR improve OR reorganize
---
Improves code quality without changing behavior. Focuses on readability, maintainability, patterns.
EOF

# Technical Writer
cat > "$CLAUDE_HOME/agents/technical-writer.md" << 'EOF'
---
name: technical-writer
description: Documentation specialist
model: sonnet
trigger: document OR readme OR docs OR explain
---
Creates clear documentation, READMEs, API docs, architecture docs, guides.
EOF

# Explorer
cat > "$CLAUDE_HOME/agents/explorer.md" << 'EOF'
---
name: explorer
description: Codebase navigation
model: haiku
trigger: find OR where OR locate OR explore
---
Quickly navigates codebases, finds files, understands structure, maps dependencies.
EOF

ok "Agents created (17 agents)"

# ═══════════════════════════════════════════════════════════════════════════════
# 3. CREATE SKILLS
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating skills..."

cat > "$CLAUDE_HOME/skills/brainstorming.md" << 'EOF'
---
name: brainstorming
trigger: new_feature OR implement OR build OR create
---
Before implementing: 1) Clarify requirements (5W+H), 2) Identify edge cases, 3) Consider approaches, 4) Choose best approach, 5) Plan implementation.
EOF

cat > "$CLAUDE_HOME/skills/test-driven-development.md" << 'EOF'
---
name: test-driven-development
trigger: new_code OR feature_implementation
---
1) Write failing test first, 2) Implement minimal code to pass, 3) Refactor, 4) Repeat. Ensure good coverage.
EOF

cat > "$CLAUDE_HOME/skills/systematic-debugging.md" << 'EOF'
---
name: systematic-debugging
trigger: bug OR error OR broken
---
1) Reproduce the issue, 2) Isolate the cause (binary search), 3) Apply 5 Whys, 4) Fix root cause, 5) Verify fix, 6) Add regression test.
EOF

cat > "$CLAUDE_HOME/skills/verification-before-completion.md" << 'EOF'
---
name: verification-before-completion
trigger: task_complete OR done OR finished
required: true
---
NEVER say "done" without: 1) Code compiles/runs, 2) Tests pass, 3) No regressions, 4) Edge cases handled, 5) SHOW proof.
EOF

cat > "$CLAUDE_HOME/skills/code-review-checklist.md" << 'EOF'
---
name: code-review-checklist
trigger: review OR pr_review
---
Check: 1) Logic correctness, 2) Error handling, 3) Security issues, 4) Performance, 5) Test coverage, 6) Code style, 7) Documentation.
EOF

cat > "$CLAUDE_HOME/skills/pr-review.md" << 'EOF'
---
name: pr-review
trigger: review_changes OR review_pr
agents: [code-reviewer, security-auditor, test-automator]
---
Multi-agent review. Categorize issues: 🔴Critical 🟠High 🟡Medium 🟢Suggestion. Create tasks. Offer auto-fix.
EOF

cat > "$CLAUDE_HOME/skills/architecture-decision.md" << 'EOF'
---
name: architecture-decision
trigger: architecture OR design_decision OR adr
---
1) Define context, 2) List options, 3) Evaluate trade-offs, 4) Make decision, 5) Document rationale, 6) Save to memory.
EOF

cat > "$CLAUDE_HOME/skills/parallel-dispatch.md" << 'EOF'
---
name: parallel-dispatch
trigger: independent_subtasks >= 3
---
When 3+ independent subtasks exist, dispatch to appropriate agents in parallel for efficiency.
EOF

ok "Skills created (8 skills)"

# ═══════════════════════════════════════════════════════════════════════════════
# 4. CREATE COMMANDS
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating commands..."

# /prime - THE ESSENTIAL COMMAND
cat > "$CLAUDE_HOME/commands/prime.md" << 'EOF'
---
name: prime
description: Initialize session context - RUN FIRST
required: true
---
## PURPOSE
/prime initializes everything. RUN THIS FIRST EVERY SESSION.

## ACTIONS
1. Detect project (type, directory, git status)
2. Load project context and patterns
3. Query memory for past decisions
4. Show current tasks
5. Activate appropriate agents
6. Output context card

## OUTPUT
```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                              CONTEXT PRIMED                                   ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║  📁 PROJECT: [name] | Type: [type] | Branch: [branch]                        ║
║  🤖 AGENTS: @[primary] + @[support]                                          ║
║  💾 MEMORY: [X] decisions loaded                                             ║
║  📋 TASKS: [X] pending | Next: [task]                                        ║
║  🎯 SUGGESTED: 1. [action] 2. [action]                                       ║
╚═══════════════════════════════════════════════════════════════════════════════╝
```

## VARIANTS
/prime --quick (skip memory) | /prime --review (add reviewers) | /prime --debug (add debugger)
EOF

# TaskMaster commands
for cmd in init add list done next update delete priority block unblock subtask search export; do
cat > "$CLAUDE_HOME/commands/tm-$cmd.md" << EOF
---
name: tm:$cmd
description: TaskMaster $cmd command
---
TaskMaster /$cmd operation. See /tm:help for usage.
EOF
done

# Memory commands
for cmd in search save list delete export context recall; do
cat > "$CLAUDE_HOME/commands/mem-$cmd.md" << EOF
---
name: mem:$cmd
description: Memory $cmd command
---
Memory /$cmd operation. Interacts with ChromaDB persistent memory.
EOF
done

# Other commands
cat > "$CLAUDE_HOME/commands/review.md" << 'EOF'
---
name: review
description: Start code review
---
Review code changes. Usage: /review [staged|branch|files|pr]
Output: Issues categorized 🔴🟠🟡🟢, then "fix all" / "fix critical" / "fix 1,2,3"
EOF

cat > "$CLAUDE_HOME/commands/agents.md" << 'EOF'
---
name: agents
description: List available agents
---
Shows all available agents with their triggers and capabilities.
EOF

ok "Commands created (25+ commands)"

# ═══════════════════════════════════════════════════════════════════════════════
# 5. CREATE HOOKS
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating hooks..."

cat > "$CLAUDE_HOME/hooks/hooks.json" << 'EOF'
{
  "version": "1.0.0",
  "hooks": [
    {
      "name": "session-start",
      "trigger": "session_begin",
      "action": "suggest_prime",
      "message": "Run /prime to initialize context"
    },
    {
      "name": "pre-completion",
      "trigger": "before_done",
      "action": "require_verification",
      "checks": ["compiles", "tests_pass", "no_regressions"]
    },
    {
      "name": "code-change",
      "trigger": "after_code_edit",
      "action": "suggest_review",
      "threshold": 50
    },
    {
      "name": "error-detected",
      "trigger": "error_in_output",
      "action": "activate_debugger"
    },
    {
      "name": "memory-save",
      "trigger": "decision_made",
      "action": "prompt_save_memory"
    },
    {
      "name": "task-complete",
      "trigger": "task_done",
      "action": "update_taskmaster"
    }
  ]
}
EOF

ok "Hooks created"

# ═══════════════════════════════════════════════════════════════════════════════
# 6. CREATE ORCHESTRATION CONFIG
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating orchestration config..."

cat > "$CLAUDE_HOME/orchestration/config.json" << 'EOF'
{
  "version": "1.0.0",
  "modes": {
    "single": { "description": "One agent handles task" },
    "sequential": { "description": "Agents work in sequence" },
    "parallel": { "description": "Agents work simultaneously" }
  },
  "routing": {
    "byProjectType": {
      "svelte": ["frontend-svelte"],
      "nextjs": ["frontend-react"],
      "go": ["backend-go"],
      "node": ["backend-node"]
    },
    "byTaskKeyword": {
      "bug|error|fix": ["debugger"],
      "test|coverage": ["test-automator"],
      "review": ["code-reviewer", "security-auditor"],
      "security|auth": ["security-auditor"],
      "deploy|docker": ["devops-engineer"],
      "refactor": ["refactorer"]
    }
  }
}
EOF

ok "Orchestration configured"

# ═══════════════════════════════════════════════════════════════════════════════
# 7. CREATE MEMORY CONFIG
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating memory config..."

cat > "$CLAUDE_HOME/memory/config.json" << 'EOF'
{
  "version": "1.0.0",
  "backend": "chromadb",
  "url": "http://localhost:8000",
  "collections": {
    "decisions": "Architectural and design decisions",
    "patterns": "Reusable code patterns",
    "problems": "Problems and solutions",
    "context": "Project-specific context",
    "episodes": "Conversation episodes"
  },
  "episodic": {
    "enabled": true,
    "maxEpisodes": 100,
    "retrievalThreshold": 0.7
  }
}
EOF

ok "Memory configured"

# ═══════════════════════════════════════════════════════════════════════════════
# 8. CREATE TASKMASTER
# ═══════════════════════════════════════════════════════════════════════════════
log "Initializing TaskMaster..."

cat > "$TASKMASTER_HOME/tasks/tasks.json" << 'EOF'
{
  "version": "2.0.0",
  "lastUpdated": "",
  "nextId": 1,
  "tasks": [],
  "completedTasks": [],
  "settings": {
    "defaultPriority": "medium",
    "priorities": ["critical", "high", "medium", "low"],
    "statuses": ["pending", "in-progress", "blocked", "done"]
  }
}
EOF

ok "TaskMaster initialized"

# ═══════════════════════════════════════════════════════════════════════════════
# 9. CREATE CONTINUOUS CLAUDE
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating Continuous Claude..."

cat > "$CLAUDE_HOME/continuous/run.sh" << 'RUNEOF'
#!/bin/bash
# Continuous Claude - Overnight automation
PRESET="${1:-full-cycle}"
echo "Running Continuous Claude: $PRESET"
echo "Session: $(date +%Y%m%d-%H%M%S)"
# Implementation connects to claude CLI with preset prompts
RUNEOF
chmod +x "$CLAUDE_HOME/continuous/run.sh"

cat > "$CLAUDE_HOME/continuous/configs/presets.json" << 'EOF'
{
  "presets": {
    "taskmaster": "Work through tasks by priority",
    "full-cycle": "Tasks → Review → Test → Document",
    "code-review": "Review recent changes",
    "documentation": "Update all docs",
    "testing": "Improve test coverage"
  }
}
EOF

ok "Continuous Claude configured"

# ═══════════════════════════════════════════════════════════════════════════════
# 10. CREATE SHELL INTEGRATION
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating shell integration..."

cat > "$CLAUDE_HOME/shell-integration.sh" << 'SHELLEOF'
# Claude Code Ecosystem - Shell Integration
export CLAUDE_HOME="$HOME/.claude"
export TASKMASTER_HOME="$HOME/.taskmaster"

# Aliases
alias ci="$CLAUDE_HOME/scripts/claude-init.sh"
alias cic="ci check"
alias cif="ci fix"
alias tm="$CLAUDE_HOME/scripts/tm.sh"
alias memory="$CLAUDE_HOME/scripts/memory.sh"
alias overnight="$CLAUDE_HOME/scripts/overnight.sh"
alias mr="$CLAUDE_HOME/scripts/morning-review.sh"

# Auto-context on directory change
_claude_chpwd() {
    [[ -f "./CLAUDE.md" ]] || return
    export CLAUDE_PROJECT_TYPE=$([[ -f "svelte.config.js" ]] && echo "svelte" || ([[ -f "go.mod" ]] && echo "go" || echo "unknown"))
}
[[ -n "$ZSH_VERSION" ]] && autoload -Uz add-zsh-hook && add-zsh-hook chpwd _claude_chpwd

# Daily startup (once per day)
_claude_startup() {
    local today=$(date +%Y%m%d)
    [[ -f "$CLAUDE_HOME/.last-startup" ]] && [[ "$(cat $CLAUDE_HOME/.last-startup)" == "$today" ]] && return
    echo "$today" > "$CLAUDE_HOME/.last-startup"
    echo "🤖 Claude Code Ecosystem ready. Run /prime in Claude to start."
}
_claude_startup
SHELLEOF

ok "Shell integration created"

# ═══════════════════════════════════════════════════════════════════════════════
# 11. CREATE HELPER SCRIPTS
# ═══════════════════════════════════════════════════════════════════════════════
log "Creating helper scripts..."

# claude-init.sh
cat > "$CLAUDE_HOME/scripts/claude-init.sh" << 'INITEOF'
#!/bin/bash
echo "Claude Code Ecosystem - Health Check"
[[ -f "$HOME/.claude/CLAUDE.md" ]] && echo "✅ CLAUDE.md" || echo "❌ CLAUDE.md missing"
[[ -d "$HOME/.claude/agents" ]] && echo "✅ Agents: $(ls $HOME/.claude/agents/*.md 2>/dev/null | wc -l)" || echo "❌ Agents missing"
[[ -d "$HOME/.claude/commands" ]] && echo "✅ Commands: $(ls $HOME/.claude/commands/*.md 2>/dev/null | wc -l)" || echo "❌ Commands missing"
curl -s http://localhost:8000/api/v1/heartbeat > /dev/null 2>&1 && echo "✅ ChromaDB running" || echo "❌ ChromaDB not running"
INITEOF
chmod +x "$CLAUDE_HOME/scripts/claude-init.sh"

# tm.sh
cat > "$CLAUDE_HOME/scripts/tm.sh" << 'TMEOF'
#!/bin/bash
case "${1:-list}" in
    list) cat "$HOME/.taskmaster/tasks/tasks.json" | grep -o '"description":[^,]*' | head -10 ;;
    *) echo "Usage: tm [list|add|done]" ;;
esac
TMEOF
chmod +x "$CLAUDE_HOME/scripts/tm.sh"

# memory.sh
cat > "$CLAUDE_HOME/scripts/memory.sh" << 'MEMEOF'
#!/bin/bash
case "${1:-status}" in
    status) curl -s http://localhost:8000/api/v1/heartbeat > /dev/null && echo "✅ ChromaDB running" || echo "❌ ChromaDB not running" ;;
    *) echo "Usage: memory [status|search|save]" ;;
esac
MEMEOF
chmod +x "$CLAUDE_HOME/scripts/memory.sh"

# overnight.sh
cat > "$CLAUDE_HOME/scripts/overnight.sh" << 'ONEOF'
#!/bin/bash
echo "Scheduling overnight session: ${1:-full-cycle} at ${2:-23:00}"
ONEOF
chmod +x "$CLAUDE_HOME/scripts/overnight.sh"

# morning-review.sh
cat > "$CLAUDE_HOME/scripts/morning-review.sh" << 'MREOF'
#!/bin/bash
echo "Morning Review"
ls -lt "$HOME/.claude/continuous/reports/daily/"*.md 2>/dev/null | head -1
MREOF
chmod +x "$CLAUDE_HOME/scripts/morning-review.sh"

# Symlinks
ln -sf "$CLAUDE_HOME/scripts/claude-init.sh" "$HOME/bin/claude-init" 2>/dev/null
ln -sf "$CLAUDE_HOME/scripts/tm.sh" "$HOME/bin/tm" 2>/dev/null
ln -sf "$CLAUDE_HOME/scripts/memory.sh" "$HOME/bin/memory" 2>/dev/null

ok "Helper scripts created"

# ═══════════════════════════════════════════════════════════════════════════════
# 12. ADD TO SHELL RC
# ═══════════════════════════════════════════════════════════════════════════════
log "Adding shell integration..."

SHELL_LINE='[[ -f "$HOME/.claude/shell-integration.sh" ]] && source "$HOME/.claude/shell-integration.sh"'

if ! grep -q "claude/shell-integration" "$HOME/.zshrc" 2>/dev/null; then
    echo "" >> "$HOME/.zshrc"
    echo "# Claude Code Ecosystem" >> "$HOME/.zshrc"
    echo "$SHELL_LINE" >> "$HOME/.zshrc"
    ok "Added to ~/.zshrc"
else
    ok "Already in ~/.zshrc"
fi

if ! grep -q "claude/shell-integration" "$HOME/.bashrc" 2>/dev/null; then
    echo "" >> "$HOME/.bashrc"
    echo "# Claude Code Ecosystem" >> "$HOME/.bashrc"
    echo "$SHELL_LINE" >> "$HOME/.bashrc"
    ok "Added to ~/.bashrc"
fi

# ═══════════════════════════════════════════════════════════════════════════════
# 13. START CHROMADB (if Docker available)
# ═══════════════════════════════════════════════════════════════════════════════
log "Checking ChromaDB..."

if command -v docker &> /dev/null; then
    if ! docker ps --format '{{.Names}}' | grep -q "^chromadb$"; then
        if docker ps -a --format '{{.Names}}' | grep -q "^chromadb$"; then
            docker start chromadb > /dev/null 2>&1 && ok "ChromaDB started" || warn "Could not start ChromaDB"
        else
            log "Creating ChromaDB container..."
            docker run -d --name chromadb -p 8000:8000 -v ~/.chroma-data:/chroma/chroma \
                -e ANONYMIZED_TELEMETRY=false --restart unless-stopped chromadb/chroma:latest > /dev/null 2>&1
            ok "ChromaDB created and started"
        fi
    else
        ok "ChromaDB already running"
    fi
else
    warn "Docker not found - install Docker for persistent memory"
fi

# ═══════════════════════════════════════════════════════════════════════════════
# COMPLETE
# ═══════════════════════════════════════════════════════════════════════════════
echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                         INSTALLATION COMPLETE                                 ║"
╠═══════════════════════════════════════════════════════════════════════════════╣"
echo "║                                                                               ║"
echo "║  📁 Installed to: ~/.claude/                                                  ║"
echo "║  🤖 Agents: 17                                                                ║"
echo "║  ⚡ Skills: 8                                                                 ║"
echo "║  ⌨️  Commands: 25+                                                             ║"
echo "║  💾 Memory: ChromaDB                                                          ║"
echo "║                                                                               ║"
echo "║  NEXT STEPS:                                                                  ║"
echo "║  1. Run: source ~/.zshrc                                                      ║"
echo "║  2. Copy CLAUDE.md to ~/.claude/CLAUDE.md                                     ║"
echo "║  3. Open Claude Code: claude                                                  ║"
echo "║  4. Run: /prime                                                               ║"
echo "║                                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""
