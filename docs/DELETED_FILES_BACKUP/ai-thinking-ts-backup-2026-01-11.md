# DELETED FILE BACKUP

## Metadata
- **Original Path**: `frontend/src/lib/api/ai/thinking.ts`
- **Deletion Date**: 2026-01-11
- **Backup Date**: 2026-01-11
- **Git Branch**: pedro-dev
- **Reason for Deletion**: File no longer needed - thinking system consolidated into main AI module

## Original File Content

```typescript
import { request } from '../base';
import type {
  ReasoningTemplate,
  ReasoningTemplatesResponse,
  CreateTemplateData,
  UpdateTemplateData,
  ThinkingSettings,
  ThinkingTrace
} from './types';

// Reasoning Templates
export async function getReasoningTemplates() {
  return request<ReasoningTemplatesResponse>('/thinking/templates');
}

export async function getReasoningTemplate(id: string) {
  return request<ReasoningTemplate>(`/thinking/templates/${id}`);
}

export async function createReasoningTemplate(data: CreateTemplateData) {
  return request<ReasoningTemplate>('/thinking/templates', {
    method: 'POST',
    body: data
  });
}

export async function updateReasoningTemplate(id: string, data: UpdateTemplateData) {
  return request<ReasoningTemplate>(`/thinking/templates/${id}`, {
    method: 'PUT',
    body: data
  });
}

export async function deleteReasoningTemplate(id: string) {
  return request<{ message: string }>(`/thinking/templates/${id}`, {
    method: 'DELETE'
  });
}

export async function setDefaultTemplate(id: string) {
  return request<{ message: string }>(`/thinking/templates/${id}/default`, {
    method: 'POST'
  });
}

// Thinking Settings
export async function getThinkingSettings() {
  return request<ThinkingSettings>('/thinking/settings');
}

export async function updateThinkingSettings(data: Partial<ThinkingSettings>) {
  return request<ThinkingSettings>('/thinking/settings', {
    method: 'PUT',
    body: data
  });
}

// Thinking Traces
export async function getThinkingTraces(conversationId: string) {
  return request<{ traces: ThinkingTrace[] }>(`/thinking/traces/${conversationId}`);
}

export async function getThinkingTrace(conversationId: string, messageId: string) {
  return request<ThinkingTrace | null>(`/thinking/trace/${messageId}`);
}

export async function deleteThinkingTraces(conversationId: string) {
  return request<{ message: string }>(`/thinking/traces/${conversationId}`, {
    method: 'DELETE'
  });
}
```

## Summary of Functions

### Reasoning Templates API
- `getReasoningTemplates()` - Fetch all reasoning templates
- `getReasoningTemplate(id)` - Fetch a single template by ID
- `createReasoningTemplate(data)` - Create a new reasoning template
- `updateReasoningTemplate(id, data)` - Update an existing template
- `deleteReasoningTemplate(id)` - Delete a reasoning template
- `setDefaultTemplate(id)` - Set a template as default

### Thinking Settings API
- `getThinkingSettings()` - Fetch thinking/COT settings
- `updateThinkingSettings(data)` - Update thinking settings

### Thinking Traces API
- `getThinkingTraces(conversationId)` - Fetch all traces for a conversation
- `getThinkingTrace(conversationId, messageId)` - Fetch a specific trace
- `deleteThinkingTraces(conversationId)` - Delete traces for a conversation

## Context
This file was part of the AI module API layer and provided typed wrappers around HTTP requests for the thinking/reasoning system. The functions use TypeScript generics with the `request` function from the base API module.

**Dependencies**:
- Imports from `../base` - Base request function
- Imports from `./types` - TypeScript type definitions

**Endpoints Accessed**:
- `/thinking/templates` - Reasoning template CRUD
- `/thinking/settings` - Thinking configuration
- `/thinking/traces` - Thinking trace retrieval and deletion

## Notes
- Uses `request<T>` pattern for type-safe API calls
- All functions are async and return Promises
- Follows RESTful conventions (GET, POST, PUT, DELETE)
- Part of the COT (Chain of Thought) system implementation
