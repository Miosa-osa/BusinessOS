# Notes: Agentic Context Engineering (ACE)

**Paper ID:** 2510.04618
**Date Added:** 2026-01-07
**Relevance Score:** 89/100 🔥

---

## Executive Summary

Stanford/NVIDIA paper introducing ACE framework for self-improving LLMs. Solves brevity bias and context collapse via evolving playbooks. Achieves +10.6% on agent benchmarks, +8.6% on finance tasks.

## Critical Problems Solved

### 1. Brevity Bias ⚠️

**Problem:** LLMs favor concise summaries over detailed domain expertise.

**Impact on BusinessOS:**
- Agents give short, incomplete answers
- Loss of nuanced context in workspace memories
- Users need to ask follow-up questions repeatedly

**ACE Solution:**
- Evolving playbooks that accumulate strategies
- Structured updates preserve detailed knowledge
- Natural execution feedback guides evolution

### 2. Context Collapse ⚠️

**Problem:** Information erosion through iterative context rewrites.

**Impact on BusinessOS:**
- Long COT chains lose critical details
- Multi-turn debugging sessions degrade
- Workspace context becomes vague over time

**ACE Solution:**
- Modular process: Generate → Reflect → Curate
- No lossy overwrites
- Incremental updates maintain full history

---

## ACE Framework Components

### 1. Generation Phase

**Purpose:** Create new strategies/knowledge from execution.

**For BusinessOS:**
```go
type GenerationService struct {
    observeExecution(task Task, result Result) Strategy
    extractPatterns(history []Execution) []Pattern
    proposeImprovements(failures []Error) []Strategy
}
```

**Example:**
- User asks about RAG implementation
- Agent executes, user provides feedback
- System generates: "When discussing RAG, always mention query expansion"

### 2. Reflection Phase

**Purpose:** Evaluate quality of generated strategies.

**For BusinessOS:**
```go
type ReflectionService struct {
    scoreStrategy(strategy Strategy) QualityScore
    identifyConflicts(new Strategy, existing []Strategy) []Conflict
    assessImpact(strategy Strategy, history []Execution) Impact
}
```

**Metrics:**
- Success rate with strategy
- User satisfaction signals
- Conflict with existing knowledge

### 3. Curation Phase

**Purpose:** Merge best strategies into playbook without degradation.

**For BusinessOS:**
```go
type CurationService struct {
    mergeStrategies(candidates []Strategy) Playbook
    resolveConflicts(conflicts []Conflict) Resolution
    pruneRedundant(playbook Playbook) Playbook
}
```

**Rules:**
- Keep high-impact strategies
- Remove redundant/outdated ones
- Resolve conflicts via voting/recency

---

## Implementation Plan for BusinessOS

### Phase 1: Foundation (Week 1-2)

**Goal:** Capture execution feedback

```go
// Add to role_context.go
type ExecutionFeedback struct {
    TaskID      uuid.UUID
    AgentID     uuid.UUID
    Success     bool
    UserRating  *int // 1-5 stars
    Context     string
    Result      string
    Timestamp   time.Time
}

func (s *RoleContextService) RecordFeedback(feedback ExecutionFeedback) error {
    // Store in role_execution_history table
    // Extract implicit feedback (success/failure)
    // Trigger strategy generation
}
```

### Phase 2: Generation (Week 3-4)

**Goal:** Generate strategies from feedback

```go
func (s *LearningService) GenerateStrategies(feedbacks []ExecutionFeedback) []Strategy {
    // Group by context/task type
    // Identify patterns in successful executions
    // Propose improvements for failures
    // Return candidate strategies
}
```

### Phase 3: Reflection (Week 5-6)

**Goal:** Evaluate strategy quality

```go
func (s *LearningService) ReflectOnStrategy(strategy Strategy) QualityScore {
    // Test against historical executions
    // Measure improvement over baseline
    // Check for conflicts with existing strategies
    // Return score + recommendation
}
```

### Phase 4: Curation (Week 7-8)

**Goal:** Merge into evolving playbook

```go
type Playbook struct {
    WorkspaceID uuid.UUID
    Strategies  []Strategy
    Version     int
    UpdatedAt   time.Time
}

func (s *RoleContextService) CuratePlaybook(new []Strategy) error {
    // Load current playbook
    // Merge high-quality strategies
    // Prune low-performing ones
    // Increment version
    // Persist to DB
}
```

---

## Metrics to Track

### Performance
- Agent success rate (before/after ACE)
- User satisfaction scores
- Task completion time
- Follow-up question frequency

### Playbook Health
- Number of active strategies
- Strategy churn rate (added/removed per week)
- Conflict frequency
- Playbook size growth

### Business Impact
- +10.6% agent benchmark improvement (target)
- -20% user complaints (target)
- +30% first-shot success rate (target)

---

## Key Insights

1. **No Supervision Required:** ACE learns from natural execution feedback - no labeled data needed.

2. **Scales with Context:** Long-context models benefit more - our Claude integration is perfect.

3. **Smaller Models Competitive:** With ACE, smaller open-source models match production-level agents.

4. **Modular Design:** Can implement phases incrementally - doesn't require big-bang rewrite.

---

## Potential Challenges

### 1. Storage Growth
- Playbooks grow over time
- Need pruning strategy
- Consider TTL for old strategies

### 2. Cold Start
- New workspaces have empty playbooks
- Need baseline strategies
- Transfer learning from other workspaces?

### 3. Conflict Resolution
- Multiple strategies for same context
- How to choose? Recency vs quality?
- User override mechanism?

---

## Related Papers

- Dynamic Cheatsheet (ACE builds on this)
- Context Engineering Survey (2507.13334)
- AppWorld benchmark papers

---

## Code Availability

- [ ] Check GitHub for ACE implementation
- [ ] Look for Stanford/NVIDIA repos
- [ ] Check if code released with paper

---

## Meeting Notes

**2026-01-07 - Initial Review**
- HIGH PRIORITY - solves real BusinessOS problems
- Start with feedback capture infrastructure
- Prototype generation phase first
- Target: 2-month implementation

---

## Pseudocode Examples

```go
// Example: Evolving playbook for code review agent

// Week 1 execution
user: "Review this Go code"
agent: *gives short, generic review*
feedback: UserRating=2, "Too brief"

// Generation phase extracts
strategy1: "For code reviews, provide detailed explanations"

// Week 2 execution
user: "Review this authentication code"
agent: *uses strategy1, gives detailed review*
feedback: UserRating=5, "Much better!"

// Reflection confirms
strategy1.score = 0.95 // high quality

// Curation adds permanently
playbook.strategies.append(strategy1)

// Week 3+: All code reviews now detailed by default
```

---

**Status:** High priority, ready for implementation planning
**Next Action:** Create detailed technical design doc, assign to backend team
**Expected Impact:** +10-15% improvement in agent quality, user satisfaction
