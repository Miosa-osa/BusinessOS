# Knowledge OS Integration Plan

**Date:** January 3, 2026
**Status:** Planning
**Goal:** Transform "Contexts" module into "Knowledge OS" with 3D visualization from Node Viewer

---

## Executive Summary

Integrate Node Viewer's 3D bubble graph visualization and semantic search UX into BusinessOS, rebranding the "Contexts" module as "Knowledge OS" - the intelligent memory and knowledge management layer.

---

## Architecture Comparison

### Node Viewer (Source)
```
Technology Stack:
- Frontend: React + React Three Fiber (3D)
- Backend: FastAPI (Python)
- Embeddings: LEANN + sentence-transformers/all-MiniLM-L6-v2 (384 dims)
- Storage: In-memory with JSON persistence

Core Features:
- 3D glass bubble visualization
- Particle effects and animations
- Semantic search with LEANN
- Memory connections (relatedIds)
```

### BusinessOS (Target)
```
Technology Stack:
- Frontend: SvelteKit + Threlte (for 3D)
- Backend: Go + Gin
- Embeddings: pgvector + nomic-embed-text (768 dims)
- Storage: PostgreSQL

Existing Features (from Pedro Tasks V2):
- Memories system (15 endpoints)
- Context profiles
- Tree search (semantic, title, content, browse)
- Document processing
- Conversation intelligence
- Learning system
```

---

## Feature Mapping

| Node Viewer Feature | BusinessOS Equivalent | Status | Action |
|-------------------|----------------------|--------|--------|
| Memory type | `memories` table | EXISTS | Use directly |
| Memory.relatedIds | `memory_associations` table | EXISTS | Wire up |
| Memory.tags | `memories.tags[]` column | EXISTS | Use directly |
| Memory.content | `memories.content` column | EXISTS | Use directly |
| Memory.embedding | `memories.embedding` (768) | EXISTS | Use directly |
| BubbleGraph.tsx | New: `KnowledgeGraph.svelte` | MISSING | Create |
| LEANN search | Tree search semantic mode | EXISTS | Enhance |
| Glass bubble nodes | Threlte spheres + shader | MISSING | Create |
| Particle effects | Threlte particles | MISSING | Create |
| Zoom controls | Threlte camera controls | MISSING | Create |
| Connection lines | Threlte Line component | MISSING | Create |

---

## Implementation Phases

### Phase 1: Rename and Restructure (Cosmetic)

**Goal:** Rebrand Contexts as Knowledge OS

**Files to Update:**
```
frontend/src/routes/(app)/contexts/+page.svelte
  → Update title "Contexts" → "Knowledge OS"
  → Update subtitle/description

frontend/src/routes/(app)/+layout.svelte
  → Navigation: "Contexts" → "Knowledge"
  → Icon update if needed

frontend/src/lib/components/kb/
  → Rename KBSidebar sections

frontend/src/lib/stores/kb-preferences.ts
  → Update terminology
```

**Estimated Changes:** ~20 line changes across 5 files

---

### Phase 2: Add 3D Graph View Mode

**Goal:** Create BubbleGraph equivalent in Svelte/Threlte

**New Components:**
```
frontend/src/lib/components/knowledge/
├── KnowledgeGraph.svelte          # Main 3D canvas
├── MemoryBubble.svelte            # Single bubble node
├── ConnectionLine.svelte          # Lines between related nodes
├── GraphControls.svelte           # Zoom, pan, filters
└── GlassMaterial.svelte           # Custom glass shader
```

**Dependencies to Add:**
```json
{
  "@threlte/core": "^7.0.0",
  "@threlte/extras": "^8.0.0",
  "three": "^0.160.0"
}
```

**View Toggle:**
- Add view mode toggle: Table | Grid | Graph
- Graph mode shows 3D bubble visualization
- Clicking bubble opens memory/context in side panel

**Technical Approach:**
```svelte
<!-- KnowledgeGraph.svelte -->
<script lang="ts">
  import { Canvas, T } from '@threlte/core';
  import { OrbitControls, Grid } from '@threlte/extras';
  import MemoryBubble from './MemoryBubble.svelte';

  let { memories, onSelect } = $props();

  // Convert memories to 3D positions using force-directed layout
  let nodes = $derived(layoutNodes(memories));
</script>

<Canvas>
  <T.PerspectiveCamera position={[0, 50, 100]} />
  <OrbitControls enableDamping />

  {#each nodes as node}
    <MemoryBubble
      memory={node.memory}
      position={node.position}
      onclick={() => onSelect(node.memory)}
    />
  {/each}
</Canvas>
```

---

### Phase 3: Memory Connections

**Goal:** Visualize related memories as connected nodes

**Backend Support (Already Exists):**
- `memory_associations` table with `memory_id` and `related_memory_id`
- Need to expose via API endpoint

**New API Endpoint:**
```go
// GET /api/v1/memories/:id/connections
func (h *MemoryHandler) GetConnections(c *gin.Context) {
    id := c.Param("id")
    connections, err := h.service.GetConnections(c, id)
    // Returns related memories with relationship metadata
}
```

**Frontend Integration:**
```typescript
// When memory selected, fetch and display connections
async function loadConnections(memoryId: string) {
  const connections = await api.getMemoryConnections(memoryId);
  // Update graph to highlight connected nodes
}
```

---

### Phase 4: Enhanced Search UX

**Goal:** Improve semantic search experience

**Current:** TreeSearchPanel with dropdown filters

**Improvements:**
1. **Inline search bar** at top of graph view
2. **Search-as-you-type** with debounce
3. **Visual highlighting** of matching nodes in graph
4. **Relevance score badges** on bubbles

**Code Example:**
```svelte
<input
  type="text"
  bind:value={searchQuery}
  oninput={debounce(performSearch, 300)}
  placeholder="Search your knowledge..."
/>

{#each memories as memory}
  <MemoryBubble
    highlighted={matchingIds.includes(memory.id)}
    relevanceScore={scores[memory.id]}
  />
{/each}
```

---

### Phase 5: Glass Bubble Aesthetics

**Goal:** Achieve Node Viewer visual style

**Node Viewer Style Elements:**
- Semi-transparent glass spheres
- Subtle glow effects
- Particle field background
- Color-coded by type

**Threlte Implementation:**
```svelte
<!-- MemoryBubble.svelte -->
<T.Mesh position={position}>
  <T.SphereGeometry args={[size, 32, 32]} />
  <T.MeshPhysicalMaterial
    transmission={0.9}
    thickness={0.5}
    roughness={0.1}
    metalness={0}
    color={getTypeColor(memory.type)}
    envMapIntensity={1}
  />
</T.Mesh>

<!-- Glow ring -->
<T.Mesh position={position} rotation.x={Math.PI / 2}>
  <T.RingGeometry args={[size * 1.1, size * 1.2, 32]} />
  <T.MeshBasicMaterial
    color={getTypeColor(memory.type)}
    opacity={0.3}
    transparent
  />
</T.Mesh>
```

---

## Data Layer Integration

### Memory Structure Alignment

```typescript
// Node Viewer Memory
interface Memory {
  id: string;
  title: string;
  source: string;
  date: string;
  content: string;
  tags: string[];
  relatedIds: string[];
  x?: number; y?: number; scale?: number;
}

// BusinessOS Memory (from Pedro's implementation)
interface Memory {
  id: string;
  title: string;
  summary: string;
  content: string;
  memory_type: string;  // maps to Node Viewer 'source'
  source_type: string;
  importance_score: number;
  tags: string[];
  embedding: number[];  // 768-dim vector
  created_at: string;
  // Connections fetched separately from memory_associations
}
```

### API Integration

**Existing Endpoints (Pedro's work):**
```
GET    /api/v1/memories           - List memories
POST   /api/v1/memories           - Create memory
GET    /api/v1/memories/:id       - Get memory
PUT    /api/v1/memories/:id       - Update memory
DELETE /api/v1/memories/:id       - Delete memory
POST   /api/v1/memories/search    - Search memories
GET    /api/v1/memories/facts     - Get user facts
```

**New Endpoints Needed:**
```
GET    /api/v1/memories/:id/connections  - Get related memories
POST   /api/v1/memories/:id/connect      - Create connection
DELETE /api/v1/memories/:id/connect/:rid - Remove connection
GET    /api/v1/memories/graph            - Get all with positions
```

---

## Navigation & UI Changes

### Sidebar Navigation

**Before:**
```
Dashboard
Chat
Tasks
Projects
Team
Clients
Contexts      ← Current
Nodes
Daily Log
Settings
```

**After:**
```
Dashboard
Chat
Tasks
Projects
Team
Clients
Knowledge     ← Renamed, updated icon
Nodes
Daily Log
Settings
```

### Knowledge OS Sub-Navigation

```
Knowledge
├── Home          # Dashboard with stats, recent, favorites
├── Memories      # All memories with graph view option
├── Documents     # Uploaded documents (PDF, DOCX)
├── Artifacts     # Code, diagrams, specs
├── Profiles      # Business, Person, Project profiles
└── Search        # Advanced tree search
```

---

## File Structure

### New Components
```
frontend/src/lib/components/knowledge/
├── KnowledgeGraph.svelte
├── MemoryBubble.svelte
├── ConnectionLine.svelte
├── GraphControls.svelte
├── GraphBackground.svelte
├── BubbleTooltip.svelte
└── index.ts
```

### Updated Routes
```
frontend/src/routes/(app)/knowledge/        # Renamed from contexts
├── +page.svelte
├── +page.ts
├── [id]/
│   ├── +page.svelte
│   └── +page.ts
├── memories/
│   └── +page.svelte
├── documents/
│   └── +page.svelte
└── graph/
    └── +page.svelte
```

---

## Migration Steps

1. **Install Threlte dependencies**
   ```bash
   cd frontend
   pnpm add @threlte/core @threlte/extras three @types/three
   ```

2. **Create knowledge components directory**
   ```bash
   mkdir -p src/lib/components/knowledge
   ```

3. **Update navigation (layout.svelte)**
   - Change "Contexts" → "Knowledge"
   - Update icon

4. **Create KnowledgeGraph component**
   - Port BubbleGraph logic to Threlte
   - Implement glass material
   - Add zoom/pan controls

5. **Add graph view mode to Knowledge page**
   - Toggle between Table/Grid/Graph
   - Wire up memory data

6. **Implement connections API**
   - Add backend endpoints
   - Wire up frontend

7. **Style polish**
   - Match glass bubble aesthetics
   - Add particle background
   - Smooth animations

---

## Success Criteria

1. [ ] "Contexts" renamed to "Knowledge" in navigation
2. [ ] 3D graph view option available
3. [ ] Memories display as glass bubbles
4. [ ] Connections shown as lines between nodes
5. [ ] Semantic search highlights matching nodes
6. [ ] Zoom/pan controls work smoothly
7. [ ] Clicking bubble opens memory details
8. [ ] Performance acceptable (60fps with 100+ nodes)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Threlte learning curve | Medium | Medium | Use examples from Node Viewer |
| Performance with many nodes | Medium | High | Implement level-of-detail, culling |
| Mobile/tablet support | High | Medium | Graceful fallback to 2D view |
| WebGL compatibility | Low | Medium | Check for WebGL support, fallback |

---

## Timeline

- **Phase 1 (Rename):** 1 hour
- **Phase 2 (3D Graph):** 4-6 hours
- **Phase 3 (Connections):** 2-3 hours
- **Phase 4 (Search UX):** 2 hours
- **Phase 5 (Polish):** 2-3 hours

**Total Estimate:** 11-15 hours

---

*This document should be updated as implementation progresses.*
