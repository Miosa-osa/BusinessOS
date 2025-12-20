// Usage Analytics API Types

export type UsagePeriod = 'today' | 'week' | 'month' | 'year' | 'all';

export interface UsageSummary {
  total_requests: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
  total_cost: number;
  period: string;
  start_date: string;
  end_date: string;
}

export interface ProviderUsage {
  provider: string;
  request_count: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
  total_cost: number;
}

export interface ModelUsage {
  model: string;
  provider: string;
  request_count: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
  total_cost: number;
}

export interface AgentUsage {
  agent_name: string;
  request_count: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
  avg_duration_ms: number;
}

export interface UsageTrendPoint {
  date: string;
  ai_requests: number;
  total_tokens: number;
  estimated_cost: number;
  mcp_requests: number;
  messages_sent: number;
}

export interface MCPToolUsage {
  tool_name: string;
  server_name: string | null;
  request_count: number;
  success_count: number;
  avg_duration_ms: number;
}
