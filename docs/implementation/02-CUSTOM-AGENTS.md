# P0: Custom Agents

> **Priority:** P0 - Critical for Beta
> **Backend Status:** Complete (15 endpoints)
> **Frontend Status:** Partial (2 endpoints used)
> **Estimated Effort:** 1-2 sprints

---

## Overview

Custom Agents allow users to create their own AI personas with specific instructions, behavior, and tools. This is a **key differentiator** for BusinessOS - users can build agents tailored to their business needs.

---

## Backend API Endpoints (Ready to Use)

### Agent Presets (Templates)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/agents/presets` | List agent preset templates |
| GET | `/api/ai/agents/presets/:id` | Get specific preset |

### Built-in Agents
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/agents` | Get all built-in agent prompts |
| GET | `/api/ai/agents/:id` | Get specific agent prompt |

### Custom Agents (User-Created)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/custom-agents` | List user's custom agents |
| POST | `/api/ai/custom-agents` | Create custom agent |
| GET | `/api/ai/custom-agents/category/:category` | List by category |
| POST | `/api/ai/custom-agents/from-preset/:presetId` | Create from preset |
| GET | `/api/ai/custom-agents/:id` | Get custom agent |
| PUT | `/api/ai/custom-agents/:id` | Update custom agent |
| DELETE | `/api/ai/custom-agents/:id` | Delete custom agent |
| POST | `/api/ai/custom-agents/:id/test` | Test custom agent |
| POST | `/api/ai/custom-agents/sandbox` | Test arbitrary prompt (sandbox) |

---

## Custom Agent Data Model

```typescript
interface CustomAgent {
  id: string;
  user_id: string;
  workspace_id: string;

  // Identity
  name: string;
  display_name: string;
  description: string;
  avatar_url?: string;
  category: AgentCategory;

  // Behavior
  system_prompt: string;
  welcome_message?: string;
  suggested_prompts?: string[];

  // Configuration
  model_preference?: string;  // e.g., "gpt-4", "claude-3"
  temperature?: number;
  max_tokens?: number;

  // Tools & Capabilities
  enabled_tools?: string[];   // MCP tools this agent can use
  can_create_artifacts?: boolean;
  can_execute_code?: boolean;
  can_search_web?: boolean;

  // Access
  is_public: boolean;         // Share with workspace
  is_featured: boolean;       // Show in featured list

  // Metadata
  usage_count: number;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

type AgentCategory =
  | 'assistant'      // General purpose
  | 'analyst'        // Data & research
  | 'writer'         // Content creation
  | 'coder'          // Development
  | 'strategist'     // Planning & strategy
  | 'support'        // Customer support
  | 'custom';        // User-defined
```

---

## Frontend Implementation Tasks

### Phase 1: Agent Library

#### 1.1 Agents Page
**File:** `src/routes/(app)/agents/+page.svelte`

- [ ] Grid/List view of all agents (built-in + custom)
- [ ] Category filter tabs
- [ ] Search by name/description
- [ ] "Create Agent" button
- [ ] Agent cards showing: name, description, category, usage count

#### 1.2 Agent Card Component
**File:** `src/lib/components/agents/AgentCard.svelte`

```svelte
<!-- Props: agent, onSelect, onEdit, onDelete -->
<div class="agent-card">
  <Avatar src={agent.avatar_url} fallback={agent.name[0]} />
  <h3>{agent.display_name}</h3>
  <p>{agent.description}</p>
  <Badge>{agent.category}</Badge>
  <span>Used {agent.usage_count} times</span>
  <Button on:click={() => onSelect(agent)}>Chat</Button>
  {#if agent.is_custom}
    <DropdownMenu>
      <DropdownItem on:click={() => onEdit(agent)}>Edit</DropdownItem>
      <DropdownItem on:click={() => onDelete(agent)}>Delete</DropdownItem>
    </DropdownMenu>
  {/if}
</div>
```

### Phase 2: Agent Builder

#### 2.1 Create/Edit Agent Page
**File:** `src/routes/(app)/agents/[id]/edit/+page.svelte`

- [ ] **Identity Section**
  - Name input
  - Display name input
  - Description textarea
  - Avatar upload
  - Category dropdown

- [ ] **Behavior Section**
  - System prompt textarea (large, with syntax highlighting)
  - Welcome message input
  - Suggested prompts (array input)

- [ ] **Configuration Section**
  - Model preference dropdown
  - Temperature slider (0-2)
  - Max tokens input

- [ ] **Tools Section**
  - Checkbox list of available MCP tools
  - Toggle: Can create artifacts
  - Toggle: Can execute code
  - Toggle: Can search web

- [ ] **Access Section**
  - Toggle: Public (share with workspace)
  - Toggle: Featured (show prominently)

#### 2.2 Agent Builder from Preset
**File:** `src/routes/(app)/agents/new/+page.svelte`

- [ ] Show preset templates gallery
- [ ] Click preset → pre-fills builder form
- [ ] "Start from scratch" option

#### 2.3 System Prompt Editor
**File:** `src/lib/components/agents/SystemPromptEditor.svelte`

- [ ] Monaco editor or CodeMirror
- [ ] Variable suggestions: `{{user_name}}`, `{{workspace_name}}`, `{{date}}`
- [ ] Character count
- [ ] Template snippets sidebar

### Phase 3: Agent Testing

#### 3.1 Agent Sandbox
**File:** `src/lib/components/agents/AgentSandbox.svelte`

- [ ] Split view: Editor | Chat Preview
- [ ] Test message input
- [ ] Real-time preview of agent responses
- [ ] "Test" button → calls `/api/ai/custom-agents/:id/test`

#### 3.2 Sandbox Modal
**File:** `src/lib/components/agents/SandboxModal.svelte`

- [ ] Quick test any prompt without saving
- [ ] Uses `/api/ai/custom-agents/sandbox`

### Phase 4: Chat Integration

#### 4.1 Agent Selector in Chat
**File:** Update `src/routes/(app)/chat/+page.svelte`

- [ ] Agent dropdown/selector in chat header
- [ ] Show selected agent's avatar and name
- [ ] "Change Agent" button
- [ ] Agent-specific welcome message on new chat

#### 4.2 Agent Quick Switch
- [ ] Keyboard shortcut: `Cmd+Shift+A` to switch agents
- [ ] Command palette integration

### Phase 5: API Client

#### 5.1 Custom Agents API Module
**File:** `src/lib/api/agents/customAgents.ts`

```typescript
export async function getCustomAgents(): Promise<CustomAgent[]>
export async function getCustomAgent(id: string): Promise<CustomAgent>
export async function createCustomAgent(data: CreateAgentInput): Promise<CustomAgent>
export async function updateCustomAgent(id: string, data: UpdateAgentInput): Promise<CustomAgent>
export async function deleteCustomAgent(id: string): Promise<void>
export async function createFromPreset(presetId: string, overrides?: Partial<CreateAgentInput>): Promise<CustomAgent>
export async function testAgent(id: string, message: string): Promise<TestResult>
export async function testSandbox(prompt: string, message: string): Promise<TestResult>
export async function getAgentsByCategory(category: string): Promise<CustomAgent[]>
export async function getAgentPresets(): Promise<AgentPreset[]>
```

#### 5.2 Agents Store
**File:** `src/lib/stores/agents.ts`

```typescript
interface AgentStore {
  agents: CustomAgent[];
  presets: AgentPreset[];
  selectedAgentId: string | null;
  isLoading: boolean;

  // Actions
  loadAgents(): Promise<void>;
  loadPresets(): Promise<void>;
  selectAgent(id: string): void;
  createAgent(data: CreateAgentInput): Promise<CustomAgent>;
  updateAgent(id: string, data: UpdateAgentInput): Promise<void>;
  deleteAgent(id: string): Promise<void>;
}
```

---

## UI/UX Requirements

### Agent Creation Flow
1. Choose: Start from preset OR start from scratch
2. Fill in identity (name, description)
3. Write system prompt (with guidance)
4. Configure capabilities
5. Test in sandbox
6. Save

### System Prompt Best Practices (Show in UI)
- Start with role definition
- Include capabilities and limitations
- Define tone and style
- Provide example interactions
- Set boundaries

### Visual Design
- Each category has a distinct color
- Custom agents show "Custom" badge
- Featured agents show star icon
- Usage stats visible on cards

---

## Testing Requirements

- [ ] Unit tests for agents store
- [ ] Component tests for AgentCard, AgentBuilder
- [ ] E2E: Create agent from preset
- [ ] E2E: Create agent from scratch
- [ ] E2E: Test agent in sandbox
- [ ] E2E: Use custom agent in chat

---

## Linear Issues to Create

1. **[AGENT-001]** Create Agents library page with grid view
2. **[AGENT-002]** Build AgentCard component
3. **[AGENT-003]** Implement Agent Builder form
4. **[AGENT-004]** Add System Prompt Editor with syntax highlighting
5. **[AGENT-005]** Create Agent Sandbox for testing
6. **[AGENT-006]** Integrate agent selector in chat
7. **[AGENT-007]** Add preset templates gallery
8. **[AGENT-008]** API client and store implementation
9. **[AGENT-009]** E2E tests for agent flows

---

## Dependencies

- MCP tools list (for capabilities selection)

## Blockers

- None identified

---

## Notes

- Consider agent marketplace in future (share/sell agents)
- Agent versioning could be useful
- Analytics: which agents are most popular/effective
