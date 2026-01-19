# Knowledge Module - Incomplete Items & TODO

**Date:** January 3, 2026
**Status:** In Progress
**Priority:** High

---

## Overview

The Knowledge Module UI is built and functional, but several critical features need completion before it's production-ready. This document tracks what's incomplete and what needs to be done.

---

## Critical Incomplete Items

### 1. Integrations System (NOT WORKING)

**Status:** UI Built, Backend Missing

The integrations modal displays all the data sources but **none of them actually connect**.

**What's Missing:**
- [ ] Backend OAuth endpoints for each provider
- [ ] `/api/integrations/status` - Returns 404 (needs implementation)
- [ ] `/api/integrations/auth/{provider}` - OAuth flow initiation
- [ ] `/api/integrations/callback` - OAuth callback handler
- [ ] Token storage and refresh logic
- [ ] Data sync workers for each provider

**Providers That Need OAuth Implementation:**
| Provider | OAuth Type | Priority |
|----------|-----------|----------|
| Gmail | Google OAuth 2.0 | High |
| Google Calendar | Google OAuth 2.0 | High |
| Google Drive | Google OAuth 2.0 | High |
| Notion | Notion OAuth | High |
| Slack | Slack OAuth | Medium |
| Dropbox | Dropbox OAuth | Medium |
| Microsoft Teams | Microsoft OAuth | Medium |
| HubSpot | HubSpot OAuth | Low |
| Linear | Linear OAuth | Low |

**File Import Providers (No OAuth, just file upload):**
- ChatGPT (JSON export)
- Claude (JSON export)
- Perplexity (JSON export)
- Granola (meeting notes)

---

### 2. AI Knowledge Indexing (NOT CONNECTED)

**Status:** Chat UI exists, but AI doesn't index the knowledge graph

**What's Missing:**
- [ ] AI should index ALL memories/contexts in the knowledge graph
- [ ] Semantic search across the entire knowledge base
- [ ] RAG (Retrieval Augmented Generation) pipeline
- [ ] Vector embeddings for each memory
- [ ] Context injection from relevant memories when chatting

**Current State:**
- Chat panel exists but just sends messages to generic AI
- No knowledge base context is included
- No semantic search over memories

**Required:**
```
1. Generate embeddings for each memory/context
2. Store embeddings in vector database (pgvector?)
3. On chat query, find relevant memories via similarity search
4. Inject relevant context into AI prompt
5. Return AI response with citations to source memories
```

---

### 3. Memory CRUD Operations (PARTIAL)

**Status:** Read & Update work, Create & Delete incomplete

**What's Missing:**
- [ ] Create new memory/bubble from UI
- [ ] Delete memory/bubble
- [ ] Duplicate memory
- [ ] Move memory between nodes
- [ ] Bulk operations (select multiple, delete multiple)

**Current State:**
- Can view memories in graph
- Can edit existing memories via document panel
- Cannot create new memories from the graph view
- Cannot delete memories

---

### 4. Data Source Sync Workers (NOT BUILT)

**Status:** Not Started

Once OAuth is connected, we need background workers to actually sync data.

**What's Missing:**
- [ ] Gmail sync worker - fetch emails, extract key info
- [ ] Calendar sync worker - fetch events
- [ ] Notion sync worker - fetch pages/databases
- [ ] Slack sync worker - fetch messages from channels
- [ ] Google Drive sync worker - index documents

**Each worker needs:**
1. Initial full sync
2. Incremental sync (new items only)
3. Rate limiting / API quota handling
4. Error recovery
5. Progress reporting

---

### 5. Search & Filtering (BASIC ONLY)

**Status:** Basic text search exists, advanced filtering missing

**What's Missing:**
- [ ] Filter by memory type (fact, decision, learning, etc.)
- [ ] Filter by source (Gmail, Notion, manual, etc.)
- [ ] Filter by date range
- [ ] Filter by importance score
- [ ] Semantic/AI-powered search
- [ ] Search highlighting in 3D graph

**Current State:**
- Basic text search highlights matching bubbles
- No filters available

---

### 6. Graph Clustering (NOT IMPLEMENTED)

**Status:** Not Started

**What's Missing:**
- [ ] Visual clustering of related memories
- [ ] Cluster by topic/theme
- [ ] Cluster by source
- [ ] Cluster by time period
- [ ] Expand/collapse clusters

**Current State:**
- All bubbles positioned individually
- No visual grouping

---

## Medium Priority Items

### 7. Cover Image System

**Status:** UI exists, upload not working

- [ ] Cover image upload to storage (S3/Supabase)
- [ ] Image cropping/positioning
- [ ] Default gradient/color fallbacks work

### 8. Related Memories

**Status:** Shows related items, but logic is basic

- [ ] Improve relationship detection algorithm
- [ ] Show relationship type (references, similar topic, same source)
- [ ] Click to navigate between related memories

### 9. Keyboard Shortcuts

**Status:** Not implemented

- [ ] `Escape` to deselect
- [ ] Arrow keys to navigate between bubbles
- [ ] `Enter` to open selected
- [ ] `/` to open search
- [ ] `N` to create new memory

### 10. Mobile Responsiveness

**Status:** Desktop only

- [ ] Touch controls for 3D graph
- [ ] Responsive panels
- [ ] Mobile-friendly integrations modal

---

## Low Priority / Future Enhancements

### 11. Import/Export

- [ ] Export knowledge graph as JSON
- [ ] Export as markdown
- [ ] Import from JSON backup
- [ ] Share individual memories

### 12. Collaboration

- [ ] Share knowledge graph with team
- [ ] Collaborative editing
- [ ] Comments on memories

### 13. Analytics

- [ ] Knowledge growth over time
- [ ] Most accessed memories
- [ ] Connection density metrics

### 14. Visualizations

- [ ] Timeline view
- [ ] List/table view
- [ ] Mind map view
- [ ] Network graph (force-directed)

---

## Backend API Endpoints Needed

### Integrations
```
GET    /api/integrations/status              # All integration statuses
GET    /api/integrations/{provider}/status   # Single provider status
POST   /api/integrations/auth/{provider}     # Start OAuth
GET    /api/integrations/callback            # OAuth callback
DELETE /api/integrations/{provider}          # Disconnect
POST   /api/integrations/{provider}/sync     # Trigger manual sync
GET    /api/integrations/{provider}/logs     # Sync history
```

### Memories
```
POST   /api/memories                         # Create memory
DELETE /api/memories/{id}                    # Delete memory
POST   /api/memories/bulk-delete             # Delete multiple
POST   /api/memories/{id}/duplicate          # Duplicate
```

### AI/Search
```
POST   /api/knowledge/search                 # Semantic search
POST   /api/knowledge/chat                   # Chat with context
POST   /api/knowledge/index                  # Trigger re-indexing
GET    /api/knowledge/embeddings/status      # Embedding status
```

---

## Estimated Effort

| Item | Effort | Dependencies |
|------|--------|--------------|
| OAuth Infrastructure | 3-5 days | Backend |
| Individual Provider OAuth | 1-2 days each | OAuth Infra |
| Sync Workers | 2-3 days each | OAuth |
| AI Indexing/RAG | 3-5 days | Vector DB |
| Memory CRUD | 1-2 days | Backend API |
| Advanced Search | 2-3 days | AI Indexing |
| Clustering | 3-5 days | - |

---

## Who Does What

| Task | Owner | Notes |
|------|-------|-------|
| OAuth Endpoints | Pedro | Go backend |
| Sync Workers | Pedro | Background jobs |
| AI/RAG Pipeline | TBD | May need ML expertise |
| Frontend Polish | Roberto | Svelte |
| Vector DB Setup | Pedro | pgvector or Pinecone |

---

## Next Steps (Priority Order)

1. **Pedro**: Implement `/api/integrations/status` endpoint
2. **Pedro**: Set up Google OAuth (Gmail, Calendar, Drive)
3. **Pedro**: Create basic sync worker framework
4. **Roberto**: Add memory create/delete UI
5. **Team**: Decide on vector DB solution for AI indexing
6. **TBD**: Implement RAG pipeline for knowledge chat

---

*Last Updated: January 3, 2026*
