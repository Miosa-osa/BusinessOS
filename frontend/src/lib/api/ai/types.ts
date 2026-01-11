
export interface LLMProvider {
  id: string;
  name: string;
  type: 'local' | 'cloud';
  description: string;
  configured: boolean;
  base_url?: string;
}

export interface LLMModel {
  id: string;
  name: string;
  provider: string;
  description?: string;
  size?: string;
  family?: string;
}

export interface AIProvidersResponse {
  providers: LLMProvider[];
  active_provider: string;
  default_model: string;
}

export interface AllModelsResponse {
  models: LLMModel[];
  active_provider: string;
  default_model: string;
}

export interface LocalModelsResponse {
  models: LLMModel[];
  provider: string;
  base_url: string;
}

export interface RecommendedModel {
  name: string;
  description: string;
  ram_required: string;
  speed: string;
  quality: string;
}

export interface AISystemInfo {
  total_ram_gb: number;
  available_ram_gb: number;
  platform: string;
  has_gpu: boolean;
  gpu_name?: string;
  recommended_models: RecommendedModel[];
}

export interface AgentInfo {
  id: string;
  name: string;
  description: string;
  prompt: string;
  category: 'general' | 'specialist' | 'system';
}

export interface WarmupResponse {
  status: string;
  model: string;
  provider: string;
  message: string;
}

// MCP Tools Types
export interface Tool {
  name: string;
  description: string;
  input_schema: Record<string, unknown>;
  source: 'builtin' | 'custom';
}

// Discriminated union for ToolResponse - prevents illegal state where both result and error are set
export type ToolResponse =
  | { success: true; result: string; error?: never }
  | { success: false; result?: never; error: string };

// Custom Agents Types
export interface CustomAgent {
  id: string;
  user_id: string;
  name: string;
  display_name: string;
  description?: string;
  avatar?: string;
  system_prompt: string;
  model_preference?: string;
  temperature?: number;
  max_tokens?: number;
  capabilities?: string[];
  tools_enabled?: string[];
  context_sources?: string[];
  thinking_enabled?: boolean;
  streaming_enabled?: boolean;
  category?: string;
  is_active?: boolean;
  usage_count?: number;
  created_at: string;
  updated_at: string;
}

export interface CustomAgentsResponse {
  agents: CustomAgent[];
}
