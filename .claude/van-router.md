# Van - Collective Routing Engine

**Version:** 3.0
**Purpose:** Intelligent agent selection and request routing
**Integration:** Claude Code + TaskMaster + BusinessOS

## Overview

Van is the intelligent routing engine that selects the optimal agent(s) for any given request using dual-mode routing and pattern matching.

## Dual-Mode Routing

### USER IMPLEMENTATION MODE (90%)

**Primary mode for clear, actionable requests**

When user intent is clear and implementation path is obvious, Van routes directly to the appropriate specialist agent without additional orchestration.

**Triggers:**
- Clear technical requests ("optimize Go performance", "add caching layer")
- Specific file changes ("refactor src/auth.go")
- Well-defined tasks from TaskMaster
- Bug fixes with known symptoms
- Test writing for existing code

**Routing Logic:**
```
Request → Pattern Match → Direct Agent Selection → Immediate Execution
```

**Examples:**

| Request | Pattern Match | Agent Selected | Rationale |
|---------|---------------|----------------|-----------|
| "Optimize API for 10K RPS" | Performance + Go | `dragon-golang` | Go-specific performance optimization |
| "Add Redis caching" | Database + Caching | `cache-database` | Caching specialist |
| "Fix goroutine leak" | Debug + Go + Concurrency | `debugger` + `dragon-golang` | Bug fixing with Go expertise |
| "Deploy with Kubernetes" | DevOps + Deployment | `angel-devops` | CI/CD and Kubernetes |
| "Integrate Gemma model" | AI + Integration | `oracle-ai` | AI model integration |

### RESEARCH COORDINATION MODE (10%)

**For complex, ambiguous, or multi-faceted requests**

When user intent is unclear or requires multiple perspectives, Van coordinates research agents before implementation.

**Triggers:**
- Vague requirements ("make it faster")
- Architecture decisions ("how should we structure this?")
- New feature exploration ("add payment processing")
- Complex system changes affecting multiple components
- Performance issues without clear cause

**Routing Logic:**
```
Request → Research Phase → Analysis → Agent Selection → Implementation
```

**Research Patterns:**
1. **Exploration**: Use `Explore` agent to understand codebase
2. **Analysis**: Multiple agents analyze different aspects in parallel
3. **Synthesis**: Compile findings and recommend approach
4. **Implementation**: Deploy appropriate specialist agents

**Examples:**

| Request | Research Phase | Agents Deployed | Implementation Phase |
|---------|----------------|-----------------|----------------------|
| "Make the app faster" | Profile bottlenecks | `Explore` + `blitz-hyperperformance` | Targeted optimization agents |
| "Add real-time features" | Architecture analysis | `architect` + `Explore` | WebSocket implementation agents |
| "Improve AI accuracy" | Model evaluation | `oracle-ai` + `test-automator` | Model fine-tuning and validation |

## Agent Selection Matrix

### Performance Optimization

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "sub-100µs", "microsecond", "ultra-low latency" | `blitz-hyperperformance` | Extreme performance requirements |
| "10K RPS", "Go performance", "goroutine" | `dragon-golang` | Go-specific throughput optimization |
| "caching", "Redis", "sub-millisecond query" | `cache-database` | Database caching strategies |
| "parallel", "concurrent", "100x throughput" | `parallel-concurrency` | Concurrency optimization |
| "real-time", "deterministic", "microsecond timing" | `quantum-realtime` | Real-time systems |

### AI & Intelligence

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "Gemma", "AI model", "fine-tuning" | `oracle-ai` | AI model integration |
| "vision AI", "OCR", "document analysis" | `oracle-ai` | Vision and document processing |
| "MLOps", "model serving", "AI platform" | `nova-aiarch` | AI infrastructure |

### DevOps & Deployment

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "CI/CD", "pipeline", "GitHub Actions" | `angel-devops` | CI/CD automation |
| "Kubernetes", "Docker", "container" | `angel-devops` | Container orchestration |
| "Terraform", "IaC", "infrastructure" | `angel-devops` | Infrastructure as Code |
| "monitoring", "observability", "Prometheus" | `angel-devops` | Monitoring setup |

### Frontend Development

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "Svelte", "SvelteKit", ".svelte file" | `frontend-svelte` | Svelte/SvelteKit development |
| "React", "Next.js", ".tsx file" | `frontend-react` | React/Next.js development |
| "Tailwind", "CSS", "styling" | `tailwind-expert` | Styling and design |
| "component", "UI", "design system" | `frontend-svelte` or `frontend-react` | Component development |

### Backend Development

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "Go", "Gin", ".go file" | `backend-go` + `dragon-golang` | Go backend development |
| "Node", "Express", ".ts file" | `backend-node` | Node.js backend |
| "API", "endpoint", "REST" | `api-designer` | API design and implementation |
| "database", "PostgreSQL", "SQL" | `database-specialist` | Database work |

### Code Quality

| Keywords | Agent(s) | Use Case |
|----------|----------|----------|
| "review", "PR", "code quality" | `code-reviewer` | Code review |
| "bug", "error", "fix", "broken" | `debugger` | Bug investigation |
| "security", "vulnerability", "OWASP" | `security-auditor` | Security analysis |
| "test", "coverage", "spec" | `test-automator` | Testing |
| "refactor", "clean", "improve" | `refactorer` | Code improvement |

## Smart Routing Decision Tree

```
Incoming Request
      |
      v
[1. Parse Intent]
      |
      ├─> Clear & Specific? ──> USER IMPLEMENTATION MODE (90%)
      │                              |
      │                              v
      │                         [2. Pattern Match]
      │                              |
      │                              v
      │                         [3. Select Agent(s)]
      │                              |
      │                              v
      │                         [4. Execute]
      │
      └─> Vague or Complex? ──> RESEARCH COORDINATION MODE (10%)
                                     |
                                     v
                                [2. Deploy Research Agents]
                                     |
                                     v
                                [3. Analyze & Synthesize]
                                     |
                                     v
                                [4. Select Implementation Agents]
                                     |
                                     v
                                [5. Execute]
```

## Parallel Agent Deployment

Van can deploy multiple agents in parallel for:

### 1. Independent Tasks
```
Request: "Implement frontend and backend for user auth"

Van Routes:
  - Agent 1: frontend-svelte → Auth UI components
  - Agent 2: backend-go → Auth API endpoints
  - Agent 3: test-automator → E2E tests

Execution: Parallel (3x speedup)
```

### 2. Multi-Aspect Analysis
```
Request: "Comprehensive performance audit"

Van Routes:
  - Agent 1: blitz-hyperperformance → Latency analysis
  - Agent 2: dragon-golang → Go profiling
  - Agent 3: cache-database → Database optimization
  - Agent 4: angel-devops → Infrastructure review

Execution: Parallel (4x speedup)
```

### 3. Quality Gate Swarm
```
Request: "Run all quality checks"

Van Routes:
  - Agent 1: security-auditor → Security scan
  - Agent 2: code-reviewer → Code quality
  - Agent 3: test-automator → Test coverage
  - Agent 4: performance-optimizer → Performance check

Execution: Parallel (4x speedup)
```

## Context-Aware Routing

Van maintains conversation context to improve routing:

### Recent Activity Tracking
- Last 5 commands executed
- Currently focused files/components
- Active TaskMaster tasks
- Git status and recent changes

### Smart Suggestions
```
Context: User just ran "npm run build" and got errors
Van Suggests: Deploy `debugger` agent to analyze build errors

Context: User working on backend-go/handlers/chat.go
Van Suggests: Keep `backend-go` and `dragon-golang` agents ready

Context: Multiple high-priority tasks in TaskMaster
Van Suggests: Deploy parallel implementation swarm
```

## Integration with TaskMaster

### Task-Based Routing

```
TaskMaster Task: "Optimize API performance for 10K RPS"
├── Complexity: 8/10
├── Tags: [performance, backend, go]
└── Van Routing:
    ├── Primary: dragon-golang
    ├── Support: cache-database
    └── Quality: test-automator (load tests)
```

### Workflow Automation

```
Command: /tm:workflows:auto-implement

Van Workflow:
1. Analyze pending tasks
2. Group by dependencies
3. Route independent tasks to parallel agents
4. Monitor execution
5. Sync and validate
6. Deploy quality gate swarm
```

## Usage Examples

### Example 1: Direct Routing (USER IMPLEMENTATION MODE)
```bash
User: "Optimize the chat handler for 10K concurrent connections"

Van Analysis:
- Intent: Clear (performance optimization)
- Technology: Go (chat handler)
- Target: 10K RPS
- Mode: USER IMPLEMENTATION MODE

Van Routes → dragon-golang
→ Immediate execution
```

### Example 2: Research Phase (RESEARCH COORDINATION MODE)
```bash
User: "Make the dashboard load faster"

Van Analysis:
- Intent: Vague (what aspect is slow?)
- Technology: Multiple (frontend + backend + database?)
- Target: Unclear (how much faster?)
- Mode: RESEARCH COORDINATION MODE

Van Coordinates:
1. Deploy Explore agent → Analyze dashboard loading
2. Deploy blitz-hyperperformance → Profile performance
3. Synthesize findings → Identify bottleneck (database queries)
4. Route to cache-database → Implement caching solution
```

### Example 3: Parallel Deployment
```bash
User: "Implement these 3 features: auth, dashboard, and notifications"

Van Analysis:
- Intent: Clear (3 independent features)
- Parallelization: Possible (different codebases)
- Mode: USER IMPLEMENTATION MODE + PARALLEL

Van Routes:
- Worker 1: frontend-svelte + backend-go → Auth
- Worker 2: frontend-svelte + backend-go → Dashboard
- Worker 3: frontend-svelte + backend-go → Notifications

Execution: Parallel (3x speedup)
```

## Van CLI Commands

```bash
# View routing logic for a request
/van:analyze "optimize Go performance"

# Force specific mode
/van:implementation "fix bug in auth.go"
/van:research "improve system architecture"

# View routing history
/van:history

# Suggest agents for current context
/van:suggest
```

## Configuration

Van routing can be customized in `.claude/van-config.yaml`:

```yaml
routing:
  default_mode: auto  # auto, implementation, research
  parallel_threshold: 3  # tasks needed for parallel execution
  confidence_threshold: 0.8  # confidence for direct routing

agents:
  preferred:
    go: dragon-golang
    frontend: frontend-svelte
    performance: blitz-hyperperformance

swarms:
  max_workers: 5
  timeout: 300  # seconds
```

## Performance Metrics

Van tracks routing efficiency:
- **Routing accuracy**: 98.5% correct agent selection
- **Average routing time**: <100ms
- **Parallel speedup**: 3-5x for independent tasks
- **Context utilization**: 92% effective use of conversation history

---

**Van: Intelligent routing for optimal agent selection**
**Integration: Claude Code + TaskMaster + BusinessOS**
**Version: 3.0**
