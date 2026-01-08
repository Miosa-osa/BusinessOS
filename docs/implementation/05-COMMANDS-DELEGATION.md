# P1: Slash Commands & Agent Delegation

> **Priority:** P1 - High Value
> **Backend Status:** Complete (9 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 1 sprint

---

## Overview

Two related features that enhance chat productivity:
1. **Slash Commands**: `/command` shortcuts (like Discord/Slack)
2. **Agent Delegation**: `@agent` mentions to invoke specialized agents

---

## Backend API Endpoints

### Slash Commands
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/commands` | List all commands (built-in + custom) |
| POST | `/api/ai/commands` | Create custom command |
| GET | `/api/ai/commands/:id` | Get command |
| PUT | `/api/ai/commands/:id` | Update command |
| DELETE | `/api/ai/commands/:id` | Delete command |

### Agent Delegation
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/delegation/agents` | List delegation agents |
| GET | `/api/ai/delegation/resolve/:mention` | Resolve @mention to agent |
| POST | `/api/ai/delegation/mentions` | Extract mentions from text |
| POST | `/api/ai/delegation/delegate` | Delegate task to agent |

---

## Data Models

### Slash Command
```typescript
interface SlashCommand {
  id: string;
  name: string;           // e.g., "summarize"
  display_name: string;   // e.g., "Summarize"
  description: string;
  prompt_template: string;
  category: CommandCategory;
  is_builtin: boolean;
  parameters?: CommandParameter[];
  example_usage?: string;
  created_at: string;
}

interface CommandParameter {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'selection';
  required: boolean;
  default_value?: any;
  options?: string[];  // For selection type
  description: string;
}

type CommandCategory =
  | 'content'     // Content generation
  | 'analysis'    // Analysis & research
  | 'coding'      // Code-related
  | 'workflow'    // Process automation
  | 'custom';     // User-created
```

### Delegation Agent
```typescript
interface DelegationAgent {
  id: string;
  name: string;           // e.g., "coder"
  display_name: string;   // e.g., "Code Assistant"
  trigger: string;        // e.g., "@coder"
  description: string;
  capabilities: string[];
  system_prompt: string;
  icon?: string;
}
```

---

## Built-in Slash Commands (Examples)

| Command | Description | Template |
|---------|-------------|----------|
| `/summarize` | Summarize text | "Summarize the following: {{input}}" |
| `/explain` | Explain concept | "Explain {{input}} in simple terms" |
| `/translate` | Translate text | "Translate to {{language}}: {{input}}" |
| `/improve` | Improve writing | "Improve this text: {{input}}" |
| `/code` | Generate code | "Write {{language}} code to: {{input}}" |
| `/debug` | Debug code | "Debug this code: {{input}}" |
| `/review` | Code review | "Review this code: {{input}}" |
| `/tasks` | Extract tasks | "Extract action items from: {{input}}" |
| `/ideas` | Generate ideas | "Generate ideas for: {{input}}" |
| `/outline` | Create outline | "Create an outline for: {{input}}" |

---

## Built-in Delegation Agents (Examples)

| Trigger | Agent | Capabilities |
|---------|-------|--------------|
| `@coder` | Code Assistant | Write, review, debug code |
| `@writer` | Writing Assistant | Content, copywriting, editing |
| `@analyst` | Data Analyst | Analysis, insights, reports |
| `@researcher` | Research Assistant | Research, summarize, cite |
| `@planner` | Project Planner | Plans, timelines, breakdown |
| `@support` | Support Agent | Answer questions, troubleshoot |

---

## Frontend Implementation Tasks

### Phase 1: Slash Commands in Chat

#### 1.1 Command Autocomplete
**File:** `src/lib/components/chat/CommandAutocomplete.svelte`

```svelte
<script lang="ts">
  export let input: string;
  export let onSelect: (command: SlashCommand) => void;

  let commands = [];
  let filteredCommands = [];
  let selectedIndex = 0;

  $: if (input.startsWith('/')) {
    const query = input.slice(1).toLowerCase();
    filteredCommands = commands.filter(c =>
      c.name.includes(query) || c.description.toLowerCase().includes(query)
    );
  }
</script>

{#if filteredCommands.length > 0}
  <div class="command-autocomplete">
    {#each filteredCommands as command, i}
      <button
        class="command-item"
        class:selected={i === selectedIndex}
        on:click={() => onSelect(command)}
      >
        <span class="name">/{command.name}</span>
        <span class="description">{command.description}</span>
      </button>
    {/each}
  </div>
{/if}
```

#### 1.2 Integrate into Chat Input
**File:** Update `src/routes/(app)/chat/+page.svelte`

- [ ] Detect `/` at start of input
- [ ] Show CommandAutocomplete popup
- [ ] Arrow key navigation
- [ ] Enter to select command
- [ ] Escape to dismiss
- [ ] Tab completion for command name

#### 1.3 Command Parameter Modal
**File:** `src/lib/components/chat/CommandParameterModal.svelte`

- [ ] Show when command has parameters
- [ ] Input for each parameter
- [ ] Validation for required params
- [ ] Submit to execute command

### Phase 2: Agent Delegation

#### 2.1 Mention Autocomplete
**File:** `src/lib/components/chat/MentionAutocomplete.svelte`

```svelte
<script lang="ts">
  export let input: string;
  export let cursorPosition: number;
  export let onSelect: (agent: DelegationAgent) => void;

  let agents = [];
  let filteredAgents = [];

  // Detect @mention at cursor position
  $: {
    const textBeforeCursor = input.slice(0, cursorPosition);
    const mentionMatch = textBeforeCursor.match(/@(\w*)$/);
    if (mentionMatch) {
      const query = mentionMatch[1].toLowerCase();
      filteredAgents = agents.filter(a =>
        a.name.includes(query) || a.display_name.toLowerCase().includes(query)
      );
    } else {
      filteredAgents = [];
    }
  }
</script>
```

#### 2.2 Integrate into Chat Input
- [ ] Detect `@` in input
- [ ] Show MentionAutocomplete at cursor position
- [ ] Insert agent trigger on select
- [ ] Style mentions in input (highlight)

#### 2.3 Agent Badge in Message
**File:** `src/lib/components/chat/AgentBadge.svelte`

- [ ] Show which agent handled the message
- [ ] Display agent icon and name
- [ ] Tooltip with agent description

### Phase 3: Command Management

#### 3.1 Commands Settings Page
**File:** `src/routes/(app)/settings/ai/commands/+page.svelte`

- [ ] List all commands (built-in + custom)
- [ ] Filter by category
- [ ] Search commands
- [ ] Create new command button

#### 3.2 Command Editor
**File:** `src/lib/components/commands/CommandEditor.svelte`

- [ ] Name input (auto-generates from display name)
- [ ] Display name input
- [ ] Description textarea
- [ ] Category selector
- [ ] Prompt template editor with variable highlighting
- [ ] Parameters builder
- [ ] Example usage input
- [ ] Test command feature

#### 3.3 Command Card
**File:** `src/lib/components/commands/CommandCard.svelte`

```svelte
<div class="command-card">
  <div class="header">
    <code>/{command.name}</code>
    <Badge>{command.category}</Badge>
    {#if command.is_builtin}
      <Badge variant="outline">Built-in</Badge>
    {/if}
  </div>
  <p>{command.description}</p>
  {#if command.example_usage}
    <code class="example">{command.example_usage}</code>
  {/if}
  {#if !command.is_builtin}
    <div class="actions">
      <Button on:click={() => edit(command)}>Edit</Button>
      <Button variant="ghost" on:click={() => delete(command)}>Delete</Button>
    </div>
  {/if}
</div>
```

### Phase 4: API Client

#### 4.1 Commands API
**File:** `src/lib/api/commands/commands.ts`

```typescript
export async function getCommands(): Promise<SlashCommand[]>
export async function createCommand(data: CreateCommandInput): Promise<SlashCommand>
export async function updateCommand(id: string, data: UpdateCommandInput): Promise<SlashCommand>
export async function deleteCommand(id: string): Promise<void>
```

#### 4.2 Delegation API
**File:** `src/lib/api/delegation/delegation.ts`

```typescript
export async function getDelegationAgents(): Promise<DelegationAgent[]>
export async function resolveMention(mention: string): Promise<DelegationAgent | null>
export async function extractMentions(text: string): Promise<string[]>
export async function delegateToAgent(agentId: string, task: string): Promise<DelegationResult>
```

#### 4.3 Commands Store
**File:** `src/lib/stores/commands.ts`

```typescript
interface CommandsStore {
  commands: SlashCommand[];
  delegationAgents: DelegationAgent[];
  isLoading: boolean;

  loadCommands(): Promise<void>;
  loadAgents(): Promise<void>;
  createCommand(data: CreateCommandInput): Promise<void>;
  updateCommand(id: string, data: UpdateCommandInput): Promise<void>;
  deleteCommand(id: string): Promise<void>;
}
```

---

## UI/UX Requirements

### Command Autocomplete
- Appears above input field
- Keyboard navigation (up/down/enter/escape)
- Shows command name, description, category
- Fuzzy search matching
- Most used commands at top

### Mention Autocomplete
- Appears at cursor position
- Shows agent avatar, name, brief description
- Keyboard navigation

### Visual Styling
- Commands: monospace blue text (`/command`)
- Mentions: highlighted background (`@agent`)
- Both should be visually distinct

---

## Testing Requirements

- [ ] Unit tests for autocomplete logic
- [ ] Component tests for autocomplete UI
- [ ] E2E: Use slash command in chat
- [ ] E2E: Mention agent in chat
- [ ] E2E: Create custom command

---

## Linear Issues to Create

1. **[CMD-001]** Create CommandAutocomplete component
2. **[CMD-002]** Integrate slash commands into chat input
3. **[CMD-003]** Build Command Parameter Modal
4. **[CMD-004]** Create MentionAutocomplete component
5. **[CMD-005]** Integrate agent mentions into chat
6. **[CMD-006]** Build Commands Settings page
7. **[CMD-007]** Create Command Editor
8. **[CMD-008]** API client and store implementation
9. **[CMD-009]** E2E tests

---

## Dependencies

- Chat input refactor may be needed

## Blockers

- None identified

---

## Notes

- Consider command aliases (e.g., `/s` for `/summarize`)
- Command history could be useful
- Favorite commands feature
