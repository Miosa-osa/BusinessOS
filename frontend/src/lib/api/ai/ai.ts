import { request, getApiBaseUrl } from '../base';
import type {
  AIProvidersResponse,
  AllModelsResponse,
  LocalModelsResponse,
  AISystemInfo,
  AgentInfo,
  WarmupResponse,
  Tool,
  ToolResponse,
  CustomAgentsResponse
} from './types';

// AI Providers
export async function getAIProviders() {
  return request<AIProvidersResponse>('/ai/providers');
}

export async function updateAIProvider(provider: string) {
  return request<{ message: string }>('/ai/provider', {
    method: 'PUT',
    body: { provider }
  });
}

export async function getAllModels() {
  return request<AllModelsResponse>('/ai/models');
}

export async function getLocalModels() {
  return request<LocalModelsResponse>('/ai/models/local');
}

export async function pullModel(model: string): Promise<ReadableStream<Uint8Array> | null> {
  const response = await fetch(`${getApiBaseUrl()}/ai/models/pull`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ model })
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Pull model failed' }));
    throw new Error(error.detail || `Failed to pull model (HTTP ${response.status})`);
  }

  return response.body;
}

export async function warmupModel(model: string): Promise<WarmupResponse> {
  return request('/ai/models/warmup', {
    method: 'POST',
    body: { model }
  });
}

export async function getAISystemInfo() {
  return request<AISystemInfo>('/ai/system');
}

export async function saveAPIKey(provider: string, apiKey: string) {
  return request<{ message: string }>('/ai/api-key', {
    method: 'POST',
    body: { provider, api_key: apiKey }
  });
}

export async function getAgentPrompts() {
  return request<{ agents: AgentInfo[] }>('/ai/agents');
}

export async function getAgentPrompt(id: string) {
  return request<{ id: string; prompt: string }>(`/ai/agents/${id}`);
}

export async function getTools() {
  return request<{ tools: Tool[] }>('/mcp/tools');
}

export async function executeTool(toolName: string, args: Record<string, unknown>) {
  return request<ToolResponse>('/mcp/execute', {
    method: 'POST',
    body: { tool_name: toolName, arguments: args }
  });
}

// Custom Agents
export async function getCustomAgents(includeInactive = false) {
  const params = includeInactive ? '?include_inactive=true' : '';
  return request<CustomAgentsResponse>(`/ai/custom-agents${params}`);
}
