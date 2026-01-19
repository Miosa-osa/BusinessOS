# P2: Workflows (Automation)

> **Priority:** P2 - Nice to Have
> **Backend Status:** Complete (8 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2 sprints

---

## Overview

Workflows allow users to create automated sequences of AI actions triggered by events or schedules. Think: Zapier/n8n for AI tasks within BusinessOS.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ai/workflows` | List workflows |
| POST | `/api/ai/workflows` | Create workflow |
| GET | `/api/ai/workflows/:id` | Get workflow |
| DELETE | `/api/ai/workflows/:id` | Delete workflow |
| POST | `/api/ai/workflows/:id/execute` | Execute workflow manually |
| POST | `/api/ai/workflows/trigger/:trigger` | Execute by trigger name |
| GET | `/api/ai/workflows/executions` | List executions |
| GET | `/api/ai/workflows/executions/:id` | Get execution details |

---

## Data Models

```typescript
interface Workflow {
  id: string;
  name: string;
  description: string;
  trigger: WorkflowTrigger;
  steps: WorkflowStep[];
  is_active: boolean;
  created_at: string;
}

interface WorkflowTrigger {
  type: 'manual' | 'schedule' | 'event' | 'webhook';
  config: TriggerConfig;
}

interface WorkflowStep {
  id: string;
  type: 'ai_prompt' | 'condition' | 'action' | 'delay';
  config: StepConfig;
  next_step_id?: string;
  on_error?: 'stop' | 'continue' | 'retry';
}

interface WorkflowExecution {
  id: string;
  workflow_id: string;
  status: 'running' | 'completed' | 'failed';
  started_at: string;
  completed_at?: string;
  step_results: StepResult[];
  error?: string;
}
```

---

## Frontend Implementation Tasks

### Phase 1: Workflow List & Management
- [ ] Workflows page with list view
- [ ] Create workflow button
- [ ] Enable/disable toggle
- [ ] Delete workflow
- [ ] Manual execute button

### Phase 2: Workflow Builder (Visual)
- [ ] Drag-and-drop workflow canvas
- [ ] Step library sidebar
- [ ] Connect steps with arrows
- [ ] Step configuration panel
- [ ] Trigger configuration

### Phase 3: Execution Monitoring
- [ ] Execution history list
- [ ] Real-time execution viewer
- [ ] Step-by-step progress
- [ ] Error details and retry

---

## Linear Issues to Create

1. **[WF-001]** Create Workflows list page
2. **[WF-002]** Build visual workflow builder
3. **[WF-003]** Implement step configuration
4. **[WF-004]** Add execution monitoring
5. **[WF-005]** API client and store

---

## Notes

- Start with simple linear workflows
- Visual builder is complex - consider phased approach
- May want to use existing flow library (ReactFlow equivalent for Svelte)
