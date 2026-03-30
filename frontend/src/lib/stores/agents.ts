import { writable, derived } from "svelte/store";
import {
  getCustomAgents,
  getCustomAgent,
  createCustomAgent,
  updateCustomAgent,
  deleteCustomAgent,
  getAgentsByCategory,
  getAgentPresets,
  getAgentPreset,
  createFromPreset,
  testAgent,
  testSandbox,
} from "$lib/api/ai/ai";
import type { CustomAgent, AgentPreset, CustomAgentsResponse, AgentPresetsResponse } from "$lib/api/ai/types";

interface AgentFilters {
  category: string | null;
  search: string;
  status: "active" | "inactive" | null;
}

interface AgentsState {
  agents: CustomAgent[];
  currentAgent: CustomAgent | null;
  presets: AgentPreset[];
  loading: boolean;
  error: string | null;
  filters: AgentFilters;
}

function createAgentsStore() {
  const { subscribe, update } = writable<AgentsState>({
    agents: [],
    currentAgent: null,
    presets: [],
    loading: false,
    error: null,
    filters: {
      category: null,
      search: "",
      status: null,
    },
  });

  // Request versioning to prevent race conditions
  let loadRequestId = 0;

  return {
    subscribe,

    // ============ Agent Methods ============

    async loadAgents(filters?: Partial<AgentFilters>) {
      // Increment request ID to track latest request
      const thisRequestId = ++loadRequestId;
      update((s) => ({ ...s, loading: true, error: null }));

      // Apply filters immediately if provided
      if (filters) {
        update((s) => ({
          ...s,
          filters: { ...s.filters, ...filters },
        }));
      }

      // Snapshot current filters
      let currentFilters: AgentFilters = {
        category: null,
        search: "",
        status: null,
      };
      update((s) => {
        currentFilters = { ...s.filters };
        return s;
      });

      try {
        // Safety timeout: if fetch hangs longer than 6s, bail to mock data
        const includeInactive =
          currentFilters.status === null ||
          currentFilters.status === "inactive";

        const fetchPromise = getCustomAgents(includeInactive);
        const timeoutPromise = new Promise<never>((_, reject) =>
          setTimeout(() => reject(new Error("Request timed out")), 6000),
        );

        const response = await Promise.race([fetchPromise, timeoutPromise]);

        // Safely access agents array — backend may return different shapes
        let agents: CustomAgent[] = Array.isArray(response)
          ? (response as CustomAgent[])
          : Array.isArray((response as CustomAgentsResponse)?.agents)
            ? (response as CustomAgentsResponse).agents
            : [];

        // Apply client-side filters
        if (currentFilters.category) {
          agents = agents.filter((a) => a.category === currentFilters.category);
        }

        if (currentFilters.search) {
          const searchLower = currentFilters.search.toLowerCase();
          agents = agents.filter(
            (a) =>
              a.name?.toLowerCase().includes(searchLower) ||
              a.display_name?.toLowerCase().includes(searchLower) ||
              a.description?.toLowerCase().includes(searchLower),
          );
        }

        if (currentFilters.status === "active") {
          agents = agents.filter((a) => a.is_active);
        } else if (currentFilters.status === "inactive") {
          agents = agents.filter((a) => !a.is_active);
        }

        // Only update if this is still the latest request
        if (thisRequestId === loadRequestId) {
          update((s) => ({ ...s, agents, loading: false, error: null }));
        }
      } catch (error) {
        console.error("Failed to load agents:", error);

        // Only update if this is still the latest request
        if (thisRequestId === loadRequestId) {
          // Show mock agents so the page is usable while backend is down
          const now = new Date().toISOString();
          const ago = (days: number) =>
            new Date(Date.now() - days * 86400000).toISOString();
          const mockAgents: CustomAgent[] = [
            {
              id: "demo-1",
              user_id: "demo",
              name: "sales-closer",
              display_name: "Sales Closer",
              description:
                "Qualifies inbound leads, drafts personalized follow-up sequences, and scores pipeline opportunities based on engagement signals",
              system_prompt:
                "You are a sales assistant specializing in B2B SaaS. Qualify leads by asking about company size, budget, timeline, and decision-making process.",
              model_preference: "claude-sonnet-4",
              temperature: 0.7,
              max_tokens: 4000,
              tools_enabled: ["web_search", "email", "crm"],
              capabilities: [
                "lead_scoring",
                "email_drafting",
                "pipeline_analysis",
              ],
              context_sources: ["crm_data", "email_history"],
              category: "sales",
              is_active: true,
              is_featured: true,
              times_used: 142,
              welcome_message:
                "Ready to close deals. Paste a lead or ask me to draft a follow-up.",
              suggested_prompts: [
                "Qualify this lead",
                "Draft a follow-up email",
                "Analyze my pipeline",
              ],
              thinking_enabled: true,
              streaming_enabled: true,
              created_at: ago(30),
              updated_at: ago(1),
            },
            {
              id: "demo-2",
              user_id: "demo",
              name: "code-reviewer",
              display_name: "Code Reviewer",
              description:
                "Reviews pull requests for correctness, security vulnerabilities, and style guide compliance. Suggests concrete improvements with code examples.",
              system_prompt:
                "You are a senior code reviewer. Focus on correctness, security, performance, and readability. Always suggest fixes, not just point out problems.",
              model_preference: "claude-sonnet-4",
              temperature: 0.2,
              max_tokens: 8000,
              tools_enabled: ["github", "code_search"],
              capabilities: ["code_review", "security_audit", "refactoring"],
              context_sources: ["repository"],
              category: "coding",
              is_active: true,
              is_featured: true,
              times_used: 289,
              welcome_message:
                "Paste code or a PR link. I'll review for bugs, security, and style.",
              suggested_prompts: [
                "Review this function",
                "Find security issues",
                "Suggest refactoring",
              ],
              thinking_enabled: true,
              streaming_enabled: true,
              created_at: ago(45),
              updated_at: ago(2),
            },
            {
              id: "demo-3",
              user_id: "demo",
              name: "content-strategist",
              display_name: "Content Strategist",
              description:
                "Creates blog posts, newsletters, and social media copy aligned with your brand voice. Optimizes for SEO and engagement metrics.",
              system_prompt:
                "You are a content strategist. Write in a professional but approachable tone. Optimize for readability and SEO. Include meta descriptions and suggested headlines.",
              model_preference: "claude-sonnet-4",
              temperature: 0.8,
              max_tokens: 6000,
              tools_enabled: ["web_search", "seo_tools"],
              capabilities: [
                "blog_writing",
                "social_media",
                "seo_optimization",
              ],
              context_sources: ["brand_guidelines"],
              category: "writing",
              is_active: true,
              times_used: 67,
              welcome_message:
                "What content do you need? Blog post, newsletter, or social copy?",
              suggested_prompts: [
                "Write a blog post about...",
                "Create a LinkedIn post",
                "Optimize this for SEO",
              ],
              streaming_enabled: true,
              created_at: ago(20),
              updated_at: ago(3),
            },
            {
              id: "demo-4",
              user_id: "demo",
              name: "data-analyst",
              display_name: "Data Analyst",
              description:
                "Analyzes datasets, identifies trends, generates SQL queries, and creates executive-ready summaries with actionable insights",
              system_prompt:
                "You are a data analyst. When given data, identify patterns, outliers, and trends. Present findings clearly with recommendations.",
              model_preference: "claude-sonnet-4",
              temperature: 0.1,
              max_tokens: 4000,
              tools_enabled: ["calculator", "charts", "sql"],
              capabilities: ["data_analysis", "sql_generation", "reporting"],
              context_sources: ["database"],
              category: "analysis",
              is_active: false,
              times_used: 34,
              thinking_enabled: true,
              streaming_enabled: true,
              created_at: ago(60),
              updated_at: ago(15),
            },
            {
              id: "demo-5",
              user_id: "demo",
              name: "support-agent",
              display_name: "Support Agent",
              description:
                "Handles tier-1 customer inquiries, resolves common issues from the knowledge base, and escalates complex tickets with full context",
              system_prompt:
                "You are a customer support agent. Be empathetic, concise, and solution-oriented. Always confirm the issue before suggesting a fix.",
              model_preference: "claude-haiku-4",
              temperature: 0.4,
              max_tokens: 2000,
              tools_enabled: ["knowledge_base", "ticket_system"],
              capabilities: ["ticket_resolution", "faq_lookup", "escalation"],
              context_sources: ["help_docs", "ticket_history"],
              category: "support",
              is_active: true,
              times_used: 512,
              welcome_message: "How can I help you today?",
              suggested_prompts: [
                "I have an issue with...",
                "How do I...",
                "Check my ticket status",
              ],
              apply_personalization: true,
              streaming_enabled: true,
              created_at: ago(90),
              updated_at: ago(0),
            },
            {
              id: "demo-6",
              user_id: "demo",
              name: "deep-researcher",
              display_name: "Deep Researcher",
              description:
                "Conducts multi-source research on any topic. Synthesizes findings with citations, identifies conflicting viewpoints, and produces structured reports.",
              system_prompt:
                "You are a research analyst. Provide thorough, well-sourced analysis. Always cite sources. Present multiple perspectives when they exist.",
              model_preference: "claude-opus-4",
              temperature: 0.3,
              max_tokens: 16000,
              tools_enabled: ["web_search", "arxiv", "document_reader"],
              capabilities: ["research", "citation", "report_generation"],
              context_sources: ["web", "academic_papers"],
              category: "research",
              is_active: true,
              is_featured: true,
              times_used: 78,
              welcome_message:
                "What topic should I research? I'll provide a comprehensive analysis with sources.",
              suggested_prompts: [
                "Research the latest on...",
                "Compare X vs Y",
                "Summarize this paper",
              ],
              thinking_enabled: true,
              streaming_enabled: true,
              created_at: ago(15),
              updated_at: ago(1),
            },
            {
              id: "demo-7",
              user_id: "demo",
              name: "marketing-copilot",
              display_name: "Marketing Copilot",
              description:
                "Plans campaigns, writes ad copy, analyzes competitor positioning, and generates A/B test variants for landing pages and emails",
              system_prompt:
                "You are a marketing strategist. Focus on conversion-oriented copy. Use proven frameworks like AIDA, PAS, and BAB.",
              model_preference: "claude-sonnet-4",
              temperature: 0.7,
              max_tokens: 4000,
              tools_enabled: ["web_search", "analytics"],
              capabilities: [
                "campaign_planning",
                "ad_copy",
                "competitor_analysis",
              ],
              category: "marketing",
              is_active: true,
              times_used: 23,
              suggested_prompts: [
                "Write ad copy for...",
                "Analyze competitor...",
                "Plan a campaign for...",
              ],
              streaming_enabled: true,
              created_at: ago(7),
              updated_at: ago(1),
            },
            {
              id: "demo-8",
              user_id: "demo",
              name: "general-assistant",
              display_name: "General Assistant",
              description:
                "Versatile all-purpose agent for everyday tasks — drafting emails, summarizing documents, brainstorming ideas, and answering questions",
              system_prompt:
                "You are a helpful general-purpose assistant. Be concise, accurate, and proactive in offering next steps.",
              model_preference: "claude-sonnet-4",
              temperature: 0.5,
              max_tokens: 4000,
              tools_enabled: ["web_search", "calculator"],
              capabilities: ["general", "summarization", "brainstorming"],
              category: "general",
              is_active: true,
              times_used: 891,
              is_featured: true,
              welcome_message: "What can I help you with?",
              suggested_prompts: [
                "Summarize this",
                "Draft an email",
                "Help me brainstorm",
              ],
              apply_personalization: true,
              streaming_enabled: true,
              created_at: ago(120),
              updated_at: ago(0),
            },
          ];
          update((s) => ({
            ...s,
            loading: false,
            agents: mockAgents,
            error: "demo", // Signal to UI to show demo banner
          }));
        }
      }
    },

    async loadAgent(id: string) {
      update((s) => ({ ...s, loading: true, error: null }));
      try {
        const agent = await getCustomAgent(id);
        update((s) => ({ ...s, currentAgent: agent, loading: false }));
        return agent;
      } catch (error) {
        console.error("Failed to load agent:", error);
        update((s) => ({
          ...s,
          loading: false,
          error:
            error instanceof Error ? error.message : "Failed to load agent",
        }));
        return null;
      }
    },

    async createAgent(data: Partial<CustomAgent>) {
      try {
        const agent = await createCustomAgent(data);
        update((s) => ({ ...s, agents: [agent, ...s.agents] }));
        return agent;
      } catch (error) {
        console.error("Failed to create agent:", error);
        throw error;
      }
    },

    async updateAgent(id: string, data: Partial<CustomAgent>) {
      try {
        const agent = await updateCustomAgent(id, data);
        update((s) => ({
          ...s,
          agents: s.agents.map((a) => (a.id === id ? agent : a)),
          currentAgent: s.currentAgent?.id === id ? agent : s.currentAgent,
        }));
        return agent;
      } catch (error) {
        console.error("Failed to update agent:", error);
        throw error;
      }
    },

    async deleteAgent(id: string) {
      try {
        await deleteCustomAgent(id);
        update((s) => ({
          ...s,
          agents: s.agents.filter((a) => a.id !== id),
          currentAgent: s.currentAgent?.id === id ? null : s.currentAgent,
        }));
      } catch (error) {
        console.error("Failed to delete agent:", error);
        throw error;
      }
    },

    // ============ Current Agent Methods ============

    setCurrentAgent(agent: CustomAgent | null) {
      update((s) => ({ ...s, currentAgent: agent }));
    },

    clearCurrent() {
      update((s) => ({ ...s, currentAgent: null }));
    },

    // ============ Filter Methods ============

    setFilters(filters: Partial<AgentFilters>) {
      update((s) => ({
        ...s,
        filters: { ...s.filters, ...filters },
      }));
    },

    clearFilters() {
      update((s) => ({
        ...s,
        filters: {
          category: null,
          search: "",
          status: null,
        },
      }));
    },

    // ============ Error Methods ============

    clearError() {
      update((s) => ({ ...s, error: null }));
    },

    // ============ Test Utilities ============

    reset() {
      update((s) => ({
        agents: [],
        currentAgent: null,
        presets: [],
        loading: false,
        error: null,
        filters: {
          category: null,
          search: "",
          status: null,
        },
      }));
    },

    // ============ Preset Methods ============

    async loadPresets() {
      update((s) => ({ ...s, loading: true, error: null }));
      try {
        const fetchPromise = getAgentPresets();
        const timeoutPromise = new Promise<never>((_, reject) =>
          setTimeout(() => reject(new Error("Presets timeout")), 6000),
        );
        const response = await Promise.race([fetchPromise, timeoutPromise]);
        const presets = Array.isArray(response)
          ? (response as AgentPreset[])
          : Array.isArray((response as AgentPresetsResponse)?.presets)
            ? (response as AgentPresetsResponse).presets
            : [];
        update((s) => ({ ...s, presets, loading: false }));
      } catch (error) {
        console.error("Failed to load agent presets:", error);
        const now = new Date().toISOString();
        const mockPresets: AgentPreset[] = [
          {
            id: "preset-1",
            name: "executive-briefer",
            display_name: "Executive Briefer",
            description:
              "Summarizes long documents into concise executive briefs with key takeaways, action items, and risk flags",
            category: "general",
            system_prompt:
              "You are an executive assistant. Summarize documents into 3 sections: Key Takeaways, Action Items, and Risks. Be concise.",
            model_preference: "claude-sonnet-4",
            temperature: 0.3,
            capabilities: ["summarization", "document_analysis"],
            tools_enabled: ["document_reader"],
            is_featured: true,
            copy_count: 342,
            created_at: now,
          },
          {
            id: "preset-2",
            name: "api-architect",
            display_name: "API Architect",
            description:
              "Designs REST and GraphQL APIs with OpenAPI specs, pagination patterns, error handling, and versioning strategies",
            category: "coding",
            system_prompt:
              "You are an API design expert. Follow REST best practices, use proper HTTP status codes, and always include pagination and error schemas.",
            model_preference: "claude-sonnet-4",
            temperature: 0.2,
            capabilities: ["api_design", "schema_generation", "documentation"],
            tools_enabled: ["code_search"],
            is_featured: true,
            copy_count: 189,
            created_at: now,
          },
          {
            id: "preset-3",
            name: "seo-writer",
            display_name: "SEO Content Writer",
            description:
              "Writes search-optimized articles with keyword research, meta descriptions, internal linking suggestions, and readability scoring",
            category: "writing",
            system_prompt:
              "You are an SEO content specialist. Include target keywords naturally, write compelling meta descriptions, and suggest internal links.",
            model_preference: "claude-sonnet-4",
            temperature: 0.7,
            capabilities: ["seo_writing", "keyword_research"],
            tools_enabled: ["web_search", "seo_tools"],
            is_featured: false,
            copy_count: 156,
            created_at: now,
          },
          {
            id: "preset-4",
            name: "financial-analyst",
            display_name: "Financial Analyst",
            description:
              "Analyzes financial statements, calculates key ratios, builds projections, and generates investor-ready reports",
            category: "analysis",
            system_prompt:
              "You are a financial analyst. Calculate key financial ratios, identify trends, and present findings with tables and charts.",
            model_preference: "claude-sonnet-4",
            temperature: 0.1,
            capabilities: [
              "financial_analysis",
              "ratio_calculation",
              "projections",
            ],
            tools_enabled: ["calculator", "charts"],
            is_featured: true,
            copy_count: 98,
            created_at: now,
          },
          {
            id: "preset-5",
            name: "onboarding-guide",
            display_name: "Onboarding Guide",
            description:
              "Walks new team members through setup, policies, tools, and first-week checklists with a friendly, supportive tone",
            category: "support",
            system_prompt:
              "You are a friendly onboarding guide. Walk new team members through their first week step by step. Be encouraging and thorough.",
            model_preference: "claude-haiku-4",
            temperature: 0.6,
            capabilities: ["onboarding", "checklist_generation"],
            tools_enabled: ["knowledge_base"],
            is_featured: false,
            copy_count: 67,
            created_at: now,
          },
          {
            id: "preset-6",
            name: "competitive-intel",
            display_name: "Competitive Intelligence",
            description:
              "Monitors competitors, analyzes their positioning, pricing, and product changes, and generates actionable battlecards",
            category: "research",
            system_prompt:
              "You are a competitive intelligence analyst. Research competitors thoroughly, compare features, pricing, and market positioning.",
            model_preference: "claude-opus-4",
            temperature: 0.3,
            capabilities: ["competitor_analysis", "battlecard_generation"],
            tools_enabled: ["web_search"],
            is_featured: true,
            copy_count: 134,
            created_at: now,
          },
          {
            id: "preset-7",
            name: "email-drafter",
            display_name: "Email Drafter",
            description:
              "Drafts professional emails for any context — cold outreach, follow-ups, internal memos, and client updates",
            category: "writing",
            system_prompt:
              "You are an email writing expert. Match the tone to the context. Keep emails concise and action-oriented.",
            model_preference: "claude-haiku-4",
            temperature: 0.5,
            capabilities: ["email_writing", "tone_matching"],
            tools_enabled: ["email"],
            is_featured: false,
            copy_count: 445,
            created_at: now,
          },
          {
            id: "preset-8",
            name: "sql-expert",
            display_name: "SQL Expert",
            description:
              "Writes optimized SQL queries, designs schemas, suggests indexes, and explains query execution plans",
            category: "coding",
            system_prompt:
              "You are a database expert. Write efficient SQL, suggest proper indexes, and explain query plans. Support PostgreSQL syntax.",
            model_preference: "claude-sonnet-4",
            temperature: 0.1,
            capabilities: [
              "sql_generation",
              "schema_design",
              "query_optimization",
            ],
            tools_enabled: ["sql", "code_search"],
            is_featured: false,
            copy_count: 223,
            created_at: now,
          },
          {
            id: "preset-9",
            name: "meeting-summarizer",
            display_name: "Meeting Summarizer",
            description:
              "Converts meeting transcripts into structured notes with decisions, action items, owners, and deadlines",
            category: "general",
            system_prompt:
              "You are a meeting notes specialist. Extract: Decisions Made, Action Items (with owners and deadlines), Key Discussion Points, and Parking Lot items.",
            model_preference: "claude-sonnet-4",
            temperature: 0.2,
            capabilities: ["transcript_analysis", "action_extraction"],
            tools_enabled: ["document_reader"],
            is_featured: true,
            copy_count: 567,
            created_at: now,
          },
          {
            id: "preset-10",
            name: "brand-voice-coach",
            display_name: "Brand Voice Coach",
            description:
              "Analyzes your existing content to define your brand voice, then rewrites any text to match it consistently",
            category: "marketing",
            system_prompt:
              "You are a brand strategist. First analyze the user's existing content to understand their voice. Then apply that voice to new content.",
            model_preference: "claude-sonnet-4",
            temperature: 0.6,
            capabilities: [
              "brand_analysis",
              "voice_matching",
              "content_rewriting",
            ],
            tools_enabled: ["document_reader"],
            is_featured: false,
            copy_count: 89,
            created_at: now,
          },
        ];
        update((s) => ({
          ...s,
          loading: false,
          presets: mockPresets,
          error: s.error === "demo" ? "demo" : null,
        }));
      }
    },

    async loadPreset(id: string) {
      try {
        return await getAgentPreset(id);
      } catch (error) {
        console.error("Failed to load agent preset:", error);
        throw error;
      }
    },

    async createFromPreset(presetId: string, name?: string) {
      try {
        const agent = await createFromPreset(presetId, name);
        update((s) => ({ ...s, agents: [agent, ...s.agents] }));
        return agent;
      } catch (error) {
        console.error("Failed to create agent from preset:", error);
        throw error;
      }
    },

    // ============ Testing Methods ============

    async testAgent(
      id: string,
      message: string,
    ): Promise<ReadableStream<Uint8Array> | null> {
      try {
        return await testAgent(id, message);
      } catch (error) {
        console.error("Failed to test agent:", error);
        throw error;
      }
    },

    async testSandbox(config: {
      system_prompt: string;
      message: string;
      model?: string;
      temperature?: number;
    }): Promise<ReadableStream<Uint8Array> | null> {
      try {
        // Convert message to test_message for API
        const apiConfig = {
          system_prompt: config.system_prompt,
          test_message: config.message,
          model: config.model,
          temperature: config.temperature,
        };
        return await testSandbox(apiConfig);
      } catch (error) {
        console.error("Failed to test in sandbox:", error);
        throw error;
      }
    },
  };
}

export const agents = createAgentsStore();

// ============ Derived Stores ============

export const selectedAgent = derived(agents, ($agents) => $agents.currentAgent);

export const agentsByCategory = derived(agents, ($agents) => {
  const byCategory: Record<string, CustomAgent[]> = {};

  for (const agent of $agents.agents) {
    const category = agent.category || "uncategorized";
    if (!byCategory[category]) {
      byCategory[category] = [];
    }
    byCategory[category].push(agent);
  }

  return byCategory;
});

export const activeAgents = derived(agents, ($agents) =>
  $agents.agents.filter((a) => a.is_active),
);

// ============ UI Constants ============

export const categoryColors: Record<string, string> = {
  general: "agm-cat agm-cat--general",
  coding: "agm-cat agm-cat--coding",
  writing: "agm-cat agm-cat--writing",
  analysis: "agm-cat agm-cat--analysis",
  research: "agm-cat agm-cat--research",
  support: "agm-cat agm-cat--support",
  sales: "agm-cat agm-cat--sales",
  marketing: "agm-cat agm-cat--marketing",
  uncategorized: "agm-cat",
};

export const categoryLabels: Record<string, string> = {
  general: "General",
  coding: "Coding",
  writing: "Writing",
  analysis: "Analysis",
  research: "Research",
  support: "Support",
  sales: "Sales",
  marketing: "Marketing",
  uncategorized: "Other",
};
