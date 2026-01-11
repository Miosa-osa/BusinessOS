# TRACK 3A: Import Map Visualization

## Complete Import Structure - Agents System

```
┌─────────────────────────────────────────────────────────────────────┐
│                          EXTERNAL PACKAGES                          │
├─────────────────────────────────────────────────────────────────────┤
│  • svelte (onMount, onDestroy, etc.)                                │
│  • $app/navigation (goto)                                           │
│  • $app/stores (page)                                               │
│  • lucide-svelte (Search, X, Sparkles, Bot, ChevronLeft)            │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                          PAGES LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. /agents/+page.svelte (List View)                                │
│     ├─→ stores/agents (agents, categoryLabels)                      │
│     ├─→ api/ai/types (CustomAgent)                                  │
│     └─→ components/agents/AgentCard.svelte                          │
│                                                                     │
│  2. /agents/new/+page.svelte (Create New)                           │
│     ├─→ stores/agents (agents)                                      │
│     ├─→ api/ai/types (CustomAgent)                                  │
│     └─→ components/agents/AgentBuilder.svelte                       │
│                                                                     │
│  3. /agents/[id]/+page.svelte (Detail View)                         │
│     ├─→ stores/agents (agents, categoryColors, categoryLabels)      │
│     ├─→ api/ai/types (CustomAgent)                                  │
│     └─→ components/agents/AgentSandbox.svelte                       │
│                                                                     │
│  4. /agents/[id]/edit/+page.svelte (Edit)                           │
│     ├─→ stores/agents (agents)                                      │
│     ├─→ api/ai/types (CustomAgent)                                  │
│     └─→ components/agents/AgentBuilder.svelte                       │
│                                                                     │
│  5. /agents/presets/+page.svelte (Preset Gallery)                   │
│     ├─→ stores/agents (agents)                                      │
│     ├─→ api/ai/types (AgentPreset)                                  │
│     ├─→ components/agents/PresetCard.svelte                         │
│     └─→ lucide-svelte (icons)                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                       COMPONENTS LAYER                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. AgentCard.svelte (Display single agent)                         │
│     ├─→ api/ai/types (CustomAgent)                                  │
│     └─→ stores/agents (categoryColors)                              │
│                                                                     │
│  2. AgentBuilder.svelte (Create/Edit form)                          │
│     └─→ api/ai/types (CustomAgent)                                  │
│                                                                     │
│  3. AgentSandbox.svelte (Test agent)                                │
│     ├─→ api/ai (testAgent, testSandbox)                             │
│     └─→ api/ai/types (SandboxTestRequest)                           │
│                                                                     │
│  4. PresetCard.svelte (Display preset)                              │
│     └─→ lucide-svelte (Bot)                                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                         STORES LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  agents.ts (State Management)                                       │
│  ├─→ api/ai/ai (getCustomAgents, createCustomAgent, etc.)           │
│  └─→ api/ai/types (CustomAgent, AgentPreset)                        │
│                                                                     │
│  Exports:                                                           │
│  • agents (store)                                                   │
│  • categoryLabels (constant)                                        │
│  • categoryColors (constant)                                        │
│  • activeAgents (derived)                                           │
│  • inactiveAgents (derived)                                         │
│  • agentsByCategory (derived)                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                         API LAYER                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  api/ai/ai.ts (API Functions)                                       │
│  • getCustomAgents()                                                │
│  • getCustomAgent(id)                                               │
│  • createCustomAgent(data)                                          │
│  • updateCustomAgent(id, data)                                      │
│  • deleteCustomAgent(id)                                            │
│  • getAgentsByCategory(category)                                    │
│  • getAgentPresets()                                                │
│  • getAgentPreset(id)                                               │
│  • createFromPreset(presetId, name)                                 │
│  • testAgent(agentId, request)                                      │
│  • testSandbox(request)                                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                         TYPES LAYER                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  api/ai/types.ts (Type Definitions)                                 │
│  • CustomAgent                                                      │
│  • CustomAgentsResponse                                             │
│  • AgentPreset                                                      │
│  • AgentPresetsResponse                                             │
│  • AgentTestRequest                                                 │
│  • AgentTestResponse                                                │
│  • SandboxTestRequest                                               │
│  • SandboxTestResponse                                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Import Flow by Feature

### Feature 1: List All Agents
```
+page.svelte
  ↓ uses
agents store
  ↓ calls
getCustomAgents() API
  ↓ returns
CustomAgent[] type
  ↓ displayed by
AgentCard component
```

### Feature 2: Create New Agent
```
new/+page.svelte
  ↓ uses
AgentBuilder component
  ↓ calls
agents.createAgent()
  ↓ calls
createCustomAgent() API
  ↓ returns
CustomAgent type
```

### Feature 3: View Agent Details
```
[id]/+page.svelte
  ↓ uses
agents.loadAgent(id)
  ↓ calls
getCustomAgent(id) API
  ↓ returns
CustomAgent type
  ↓ displayed with
AgentSandbox component
```

### Feature 4: Edit Agent
```
[id]/edit/+page.svelte
  ↓ uses
AgentBuilder component
  ↓ calls
agents.updateAgent(id, data)
  ↓ calls
updateCustomAgent(id, data) API
  ↓ returns
CustomAgent type
```

### Feature 5: Browse Presets
```
presets/+page.svelte
  ↓ uses
agents.loadPresets()
  ↓ calls
getAgentPresets() API
  ↓ returns
AgentPreset[] type
  ↓ displayed by
PresetCard component
```

---

## Shared Dependencies Matrix

| File | Svelte | SvelteKit | Store | Types | Components | External |
|------|--------|-----------|-------|-------|------------|----------|
| **Pages** |
| /agents/+page.svelte | ✓ | ✓ | ✓ | ✓ | AgentCard | - |
| /agents/new/+page.svelte | - | ✓ | ✓ | ✓ | AgentBuilder | - |
| /agents/[id]/+page.svelte | ✓ | ✓ | ✓ | ✓ | AgentSandbox | - |
| /agents/[id]/edit/+page.svelte | ✓ | ✓ | ✓ | ✓ | AgentBuilder | - |
| /agents/presets/+page.svelte | ✓ | ✓ | ✓ | ✓ | PresetCard | lucide |
| **Components** |
| AgentCard.svelte | - | - | ✓ | ✓ | - | - |
| AgentBuilder.svelte | - | - | - | ✓ | - | - |
| AgentSandbox.svelte | ✓ | - | - | ✓ | - | - |
| PresetCard.svelte | - | - | - | - | - | lucide |
| **Store** |
| agents.ts | ✓ | - | - | ✓ | - | - |

---

## Import Resolution Paths

```
$lib                 → frontend/src/lib
$lib/stores          → frontend/src/lib/stores
$lib/api/ai          → frontend/src/lib/api/ai
$lib/api/ai/types    → frontend/src/lib/api/ai/types.ts
$lib/components      → frontend/src/lib/components
$app/navigation      → SvelteKit (built-in)
$app/stores          → SvelteKit (built-in)
```

---

## Component Reusability

### AgentBuilder
- Used by: `/agents/new/+page.svelte`
- Used by: `/agents/[id]/edit/+page.svelte`
- Props: `agent?`, `onSave`, `onCancel`
- **Reusability Score:** ⭐⭐⭐⭐⭐

### AgentCard
- Used by: `/agents/+page.svelte`
- Props: `agent`, `onSelect?`, `onEdit?`, `onDelete?`, `variant?`
- **Reusability Score:** ⭐⭐⭐⭐⭐

### AgentSandbox
- Used by: `/agents/[id]/+page.svelte`
- Props: `agentId?`, `systemPrompt?`, `onTest?`
- **Reusability Score:** ⭐⭐⭐⭐

### PresetCard
- Used by: `/agents/presets/+page.svelte`
- Props: `preset`, `onUse`
- **Reusability Score:** ⭐⭐⭐⭐

---

## Store Method Usage Map

```
agents.loadAgents()
  ← /agents/+page.svelte

agents.loadAgent(id)
  ← /agents/[id]/+page.svelte
  ← /agents/[id]/edit/+page.svelte

agents.createAgent(data)
  ← /agents/new/+page.svelte

agents.updateAgent(id, data)
  ← /agents/[id]/edit/+page.svelte
  ← /agents/[id]/+page.svelte (toggle active)

agents.deleteAgent(id)
  ← /agents/+page.svelte
  ← /agents/[id]/+page.svelte

agents.loadPresets()
  ← /agents/presets/+page.svelte

agents.createFromPreset(presetId, name)
  ← /agents/presets/+page.svelte
```

---

## Type Usage Map

```
CustomAgent
  ← Used in: All pages except presets
  ← Used in: AgentCard, AgentBuilder, AgentSandbox
  ← Used in: agents store

AgentPreset
  ← Used in: /agents/presets/+page.svelte
  ← Used in: PresetCard (as PresetTemplate)
  ← Used in: agents store

SandboxTestRequest
  ← Used in: AgentSandbox.svelte
  ← Used in: api/ai functions
```

---

## Verification Summary

| Category | Count | Status |
|----------|-------|--------|
| Pages | 5 | ✅ All imports valid |
| Components | 4 | ✅ All imports valid |
| Store files | 1 | ✅ All imports valid |
| Type files | 1 | ✅ All types defined |
| Total imports | ~50 | ✅ All resolve correctly |
| Circular deps | 0 | ✅ None found |
| Type errors | 0 | ✅ Compilation succeeds |

---

**Status:** ✅ **ALL VERIFIED**

All import paths are correct, all types exist, and no circular dependencies detected.
