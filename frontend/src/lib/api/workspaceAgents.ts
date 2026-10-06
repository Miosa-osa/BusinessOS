// Workspace Agents API client. Defines agents and their execution harnesses per
// workspace. Hosted OSA runs execute through BusinessOS, while connected-machine
// harnesses remain explicit until that runtime is available.
import { request } from "./base";

export interface WorkspaceAgent {
  id: string;
  name: string;
  role: string;
  description: string;
  runtime: AgentRuntime;
  model: string;
  system_prompt: string;
  status: string;
  created_by: string | null;
  created_at: string;
  updated_at: string;
}

export interface WorkspaceAgentRun {
  id: string;
  agent_id: string;
  input: string;
  output: string;
  runtime: AgentRuntime;
  model: string;
  status: string; // "done" | "error"
  created_at: string;
}

export interface AgentModel {
  id: string;
  label: string;
  hint: string;
}

export type AgentRuntime = "osa" | "claude-code" | "codex" | "hermes";

export interface AgentRuntimeOption {
  id: AgentRuntime;
  label: string;
  hint: string;
  hosted: boolean;
}

export const AGENT_RUNTIMES: AgentRuntimeOption[] = [
  { id: "osa", label: "OSA", hint: "Hosted BusinessOS harness", hosted: true },
  { id: "claude-code", label: "Claude Code", hint: "Connected machine runtime", hosted: false },
  { id: "codex", label: "Codex", hint: "Connected machine runtime", hosted: false },
  { id: "hermes", label: "Hermes", hint: "Connected machine runtime", hosted: false },
];

export const DEFAULT_AGENT_RUNTIME: AgentRuntime = "osa";

export function runtimeOption(id: AgentRuntime): AgentRuntimeOption {
  return AGENT_RUNTIMES.find((runtime) => runtime.id === id) ?? AGENT_RUNTIMES[0];
}

export function runtimeLabel(id: AgentRuntime): string {
  return runtimeOption(id).label;
}

// The Claude models an agent can be configured with (mirrors the backend allow-list).
export const AGENT_MODELS: AgentModel[] = [
  { id: "claude-sonnet-4-5-20250929", label: "Sonnet 4.5", hint: "Balanced (default)" },
  { id: "claude-opus-4-1-20250805", label: "Opus 4", hint: "Most capable" },
  { id: "claude-haiku-4-5-20251001", label: "Haiku 4.5", hint: "Fast and economical" },
];

export const DEFAULT_AGENT_MODEL = "claude-sonnet-4-5-20250929";

export function modelLabel(id: string): string {
  return AGENT_MODELS.find((m) => m.id === id)?.label ?? id;
}

export interface AgentInput {
  name: string;
  role?: string;
  description?: string;
  runtime?: AgentRuntime;
  model?: string;
  system_prompt?: string;
  status?: string;
}

export interface RunResult {
  output: string;
  runtime: AgentRuntime;
  model: string;
  status: string;
  run_id: string;
  created_at: string;
}

export async function listAgents(): Promise<{
  agents: WorkspaceAgent[];
  count: number;
  ai_available: boolean;
}> {
  return request(`/workspace-agents`, { skipCache: true });
}

export async function createAgent(input: AgentInput): Promise<WorkspaceAgent> {
  return request(`/workspace-agents`, { method: "POST", body: input });
}

export async function updateAgent(
  id: string,
  input: AgentInput,
): Promise<WorkspaceAgent> {
  return request(`/workspace-agents/${id}`, { method: "PUT", body: input });
}

export async function deleteAgent(id: string): Promise<{ message: string }> {
  return request(`/workspace-agents/${id}`, { method: "DELETE" });
}

export async function runAgent(
  id: string,
  input: string,
): Promise<RunResult> {
  return request(`/workspace-agents/${id}/run`, {
    method: "POST",
    body: { input },
    timeout: 120_000,
  });
}

export async function listRuns(
  id: string,
): Promise<{ runs: WorkspaceAgentRun[]; count: number }> {
  return request(`/workspace-agents/${id}/runs`, { skipCache: true });
}
