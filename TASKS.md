# Active Tasks

This file tracks active development tasks. Used by TaskMaster for automatic task loading.

## Current Sprint - Q1 Implementation

### In Progress
- [ ] Frontend integration for workspace memories
- [ ] Production deployment preparation
- [x] **Deep Research Agent Integration** ← NEW

### Pending
- [ ] Performance testing in staging
- [ ] Security audit review
- [ ] API documentation update
- [ ] User acceptance testing

### Infrastructure
- [x] **Background Jobs System** ✅ COMPLETE (2026-01-08)
  - ✅ Migration 036 applied to Supabase
  - ✅ 3 workers + scheduler auto-start
  - ✅ 12 REST API endpoints
  - ✅ Retry logic with exponential backoff
  - ✅ Cron scheduling with timezone support
  - ✅ Complete documentation + tests
  - 📄 See: BACKGROUND_JOBS_VERIFICATION.md

- [x] **Thinking/COT System Verification** ✅ COMPLETE (2026-01-11)
  - ✅ All 13 API endpoints verified functional
  - ✅ ThinkingPanel component active in production
  - ✅ Settings pages (/settings/ai/thinking, /settings/ai/templates)
  - ✅ 4 built-in reasoning templates in database
  - ✅ Real-time SSE streaming confirmed via console logs
  - ✅ Backend + Frontend builds passing
  - ✅ Complete documentation created
  - 📄 See: THINKING_COT_COMPLETE_GUIDE.md
  - 📄 See: COT_SYSTEM_VERIFICATION_SESSION_2026-01-11.md

### Completed
- [x] CUS-25: Memory Hierarchy System (v2.1.0)
- [x] CUS-26: Role-Based Agent Behavior
- [x] CUS-27: Database Schema Implementation
- [x] CUS-28: Role-Based Agent Context Service
- [x] CUS-41: RAG/Embeddings Enhancement
- [x] Backend compilation and testing
- [x] Merge pedro-dev with main-dev

---

## 🔬 Deep Research Agent - Implementation Plan

**Goal:** Add autonomous deep research capabilities to Chat module
**Based on:** GPT Researcher architecture (adapted to Go)
**Timeline:** 2-3 weeks
**Priority:** HIGH

### Phase 1: Research Agent Core (Week 1)

#### Backend - New Agent Type
- [ ] Create `AgentTypeV2Research` in agent registry
- [ ] Implement `ResearchAgent` struct with interface
- [ ] Add research-specific system prompt
- [ ] Integrate with existing AgentV2 interface

#### Research Planner Service
- [ ] Create `internal/services/research_planner.go`
- [ ] Implement question generation from user query
- [ ] Sub-query expansion logic
- [ ] Research scope determination

#### Research Executor Service
- [ ] Create `internal/services/research_executor.go`
- [ ] Parallel search execution per sub-question
- [ ] Source ranking and filtering
- [ ] Citation extraction and tracking

#### Research Aggregator Service
- [ ] Create `internal/services/research_aggregator.go`
- [ ] Information synthesis
- [ ] Duplicate removal
- [ ] Relevance scoring

#### Database Schema
- [ ] Create migration `035_research_system.sql`
- [ ] Tables:
  - `research_tasks` (id, user_id, workspace_id, query, status, created_at)
  - `research_queries` (id, task_id, question, search_results, completed)
  - `research_sources` (id, task_id, url, title, content, relevance_score, cited)
  - `research_reports` (id, task_id, content, format, citations, word_count)

### Phase 2: Integration with Existing Systems (Week 2)

#### COT Orchestration Integration
- [ ] Add ResearchAgent to OrchestratorCOT
- [ ] Implement research workflow in COT
- [ ] Multi-step thinking for research process
- [ ] Progress streaming via thinking events

#### RAG Integration
- [ ] Connect ResearchExecutor with HybridSearchService
- [ ] Use existing embedding service for semantic search
- [ ] Leverage document RAG for local research
- [ ] Implement query expansion with existing service

#### Memory Integration
- [ ] Inject workspace memories into research context
- [ ] Save research results as workspace memories
- [ ] Link research to projects/contexts
- [ ] Enable research memory retrieval

#### Focus Mode Enhancement
- [ ] Add `research` focus mode
- [ ] Configure LLM settings for research
- [ ] Enable web search by default
- [ ] Set research-specific output style

### Phase 3: Frontend & UX (Week 2-3)

#### Chat UI Updates
- [ ] Add `/research` slash command
- [ ] Research mode indicator in UI
- [ ] Progress visualization (sub-questions)
- [ ] Source list display with citations

#### Artifact System Enhancement
- [ ] Detect research reports as artifacts
- [ ] Add `research_report` artifact type
- [ ] Citation panel in artifact viewer
- [ ] Export to PDF/Markdown

#### Research Dashboard (Optional)
- [ ] New route: `/research`
- [ ] List all research tasks
- [ ] View research history
- [ ] Re-run/refine research

### Phase 4: Advanced Features (Week 3)

#### Multi-Source Search
- [ ] Web search integration (Tavily/SerpAPI)
- [ ] Local document search (existing RAG)
- [ ] Workspace memory search
- [ ] arXiv integration (via MCP)

#### Report Generation
- [ ] Structured report templates
- [ ] Multiple formats (Markdown, PDF, Docx)
- [ ] Citation formatting (APA, MLA, Chicago)
- [ ] Executive summary generation

#### Cost & Performance Tracking
- [ ] Track LLM tokens per research task
- [ ] Measure search API costs
- [ ] Research duration metrics
- [ ] Quality scoring (user feedback)

#### Collaborative Research (Future)
- [ ] Team research tasks
- [ ] Shared research workspace
- [ ] Comment on sources
- [ ] Research versioning

---

## 🎯 Success Criteria

- [ ] User can type `/research [topic]` and get comprehensive report
- [ ] Report includes 5+ diverse sources with citations
- [ ] Research completes in < 3 minutes
- [ ] Cost < $0.01 per research task
- [ ] All sources are ranked by relevance
- [ ] Reports saved as artifacts
- [ ] Integration with existing Chat UX (seamless)

---

## 📚 Reference Materials

### GPT Researcher Architecture
- **Repo:** https://github.com/assafelovic/gpt-researcher
- **Benchmark:** #1 on CMU DeepResearchGym
- **Key Concepts:**
  - Planner → Execution → Aggregation → Writing
  - Multi-agent coordination
  - Parallel search execution
  - Citation tracking

### BusinessOS Existing Features to Leverage
- ✅ AgentV2 multi-agent system
- ✅ OrchestratorCOT for workflow
- ✅ HybridSearchService for RAG
- ✅ Embedding service
- ✅ Artifact detection system
- ✅ SSE streaming
- ✅ Memory hierarchy
- ✅ Focus mode system

---

## Technical Debt
- [ ] Remove fmt.Printf, use slog everywhere
- [ ] Consolidate duplicate migrations
- [ ] Add integration tests for COT orchestrator
- [ ] Document all API endpoints in OpenAPI format

## Documentation
- [ ] Update README with new features
- [ ] Create video demo of memory chat injection
- [ ] Write deployment runbook
- [ ] Document deep research agent

## Notes
- All Linear issues (CUS-25,26,27,28,41) completed
- 253 files changed, ~5,500 lines of code added
- 14 new endpoints, 8 new database tables
- Ready for staging deployment
- **NEW:** Deep research agent based on GPT Researcher architecture
