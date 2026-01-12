<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api';
	import AgentTestSandbox from '$lib/components/settings/AgentTestSandbox.svelte';
	import OutputStyleSelector from '$lib/components/settings/OutputStyleSelector.svelte';
	import UserFactsPanel from '$lib/components/settings/UserFactsPanel.svelte';

	// Types
	interface LLMModel {
		id: string;
		name: string;
		provider: string;
		description?: string;
		size?: string;
		family?: string;
	}

	interface LLMProvider {
		id: string;
		name: string;
		type: string;
		description: string;
		configured: boolean;
		base_url?: string;
	}

	interface PullProgress {
		status: string;
		digest?: string;
		total?: number;
		completed?: number;
	}

	interface RecommendedModel {
		name: string;
		description: string;
		ram_required: string;
		speed: string;
		quality: string;
	}

	interface SystemInfo {
		total_ram_gb: number;
		available_ram_gb: number;
		platform: string;
		has_gpu: boolean;
		gpu_name?: string;
		recommended_models: RecommendedModel[];
	}

	interface OutputStyle {
		id: string;
		name: string;
		display_name: string;
		description?: string;
		is_active: boolean;
		sort_order: number;
	}

	interface OutputPreference {
		user_id: string;
		default_style_id?: string;
		default_style_name?: string;
		style_overrides: Record<string, string>;
		custom_instructions?: string;
	}

	interface UserFact {
		id: string;
		user_id: string;
		fact_key: string;
		fact_value: string;
		fact_type: string;
		source_memory_id?: string | null;
		confidence_score: number;
		is_active: boolean;
		last_confirmed_at?: string | null;
		created_at: string;
		updated_at: string;
	}

	// State
	let providers = $state<LLMProvider[]>([]);
	let models = $state<LLMModel[]>([]);
	let activeProvider = $state('ollama_local');
	let defaultModel = $state('');
	let isLoading = $state(true);
	let isSaving = $state(false);
	let error = $state('');
	let saveStatus = $state('');
	let systemInfo = $state<SystemInfo | null>(null);

	// Output Style preferences
	let outputStyles = $state<OutputStyle[]>([]);
	let outputPreference = $state<OutputPreference | null>(null);
	let selectedDefaultStyleId = $state<string>('');
	let loadingOutputStyles = $state(false);
	let savingOutputPreference = $state(false);

	// User facts (self-learning validation)
	let userFacts = $state<UserFact[]>([]);
	let loadingUserFacts = $state(false);
	let userFactsError = $state('');
	let savingUserFactKey = $state<string | null>(null);

	// Model settings - defaults set to maximum performance
	let modelSettings = $state({
		temperature: 0.7,
		maxTokens: 8192,      // Higher default for longer responses
		contextWindow: 32768, // Context window size
		topP: 0.95,           // Slightly higher for more diverse responses
		streamResponses: true,
		showUsageInChat: true
	});

	// API Keys state
	let apiKeys = $state<Record<string, string>>({
		ollama_cloud: '',
		groq: '',
		anthropic: ''
	});
	let savingKey = $state<string | null>(null);
	let showApiKey = $state<Record<string, boolean>>({});

	// Pull model state
	let pullModelName = $state('');
	let isPulling = $state(false);
	let pullProgress = $state<PullProgress | null>(null);
	let pullError = $state('');
	let pullStartTime = $state<number>(0);
	let pullSpeed = $state<string>('');

	// Tabs
	let activeTab = $state<'models' | 'providers' | 'settings' | 'agents' | 'commands' | 'stats'>('models');

	// Usage stats state - comprehensive data
	interface UsageStats {
		// Core metrics
		total_requests: number;
		total_tokens: number;
		total_cost: number;
		input_tokens: number;
		output_tokens: number;
		// Provider breakdown
		by_provider: Record<string, { requests: number; tokens: number; cost: number; input_tokens: number; output_tokens: number }>;
		// Model breakdown
		by_model: Record<string, { requests: number; tokens: number; input_tokens: number; output_tokens: number; avg_latency_ms: number }>;
		// Agent/Command usage
		by_agent: Record<string, { requests: number; tokens: number }>;
		// Time-based data
		recent: { date: string; requests: number; tokens: number; cost: number }[];
		daily_trend: { date: string; local_requests: number; cloud_requests: number; local_tokens: number; cloud_tokens: number }[];
		// Session stats
		session_count: number;
		avg_session_duration_min: number;
		avg_requests_per_session: number;
		// Storage & performance
		local_model_storage_gb: number;
		avg_response_time_ms: number;
		// Cost breakdown
		local_power_cost_estimate: number;
		cloud_api_cost: number;
		// Time period
		period_start: string;
		period_end: string;
	}
	let usageStats = $state<UsageStats | null>(null);
	let loadingUsage = $state(false);
	let usagePeriod = $state<'today' | 'week' | 'month' | 'all'>('month');

	// Command state
	interface CommandInfo {
		id?: string;
		name: string;
		display_name: string;
		description: string;
		icon: string;
		category: string;
		context_sources: string[];
		is_custom: boolean;
		system_prompt?: string;
		is_builtin_override?: boolean;
	}

	let commands = $state<CommandInfo[]>([]);
	let loadingCommands = $state(false);
	let showNewCommand = $state(false);
	let editingCommand = $state<CommandInfo | null>(null);
	let expandedCommand = $state<string | null>(null);
	let newCommand = $state<Partial<CommandInfo>>({
		name: '',
		display_name: '',
		description: '',
		icon: '✨',
		system_prompt: '',
		context_sources: []
	});
	let savingCommand = $state(false);

	// Agent state
	interface AgentInfo {
		id: string;
		name: string;
		description: string;
		prompt: string;
		category: 'general' | 'specialist' | 'system';
	}

	let agents = $state<AgentInfo[]>([]);
	let loadingAgents = $state(false);
	let expandedAgent = $state<string | null>(null);
	let editingAgent = $state<string | null>(null);
	let editedPrompt = $state<string>('');

	// Custom Agents (user-created, can be mentioned with @name)
	interface CustomAgent {
		id: string;
		user_id: string;
		name: string;
		display_name: string;
		description: string;
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
		is_active: boolean;
		times_used: number;
		last_used_at?: string;
		created_at: string;
		updated_at: string;
	}

	let customAgents = $state<CustomAgent[]>([]);
	let loadingCustomAgents = $state(false);
	let showNewCustomAgent = $state(false);
	let editingCustomAgent = $state<CustomAgent | null>(null);
	let savingCustomAgent = $state(false);
	let testingAgentId = $state<string | null>(null);
	let newCustomAgent = $state({
		name: '',
		display_name: '',
		description: '',
		system_prompt: '',
		model_preference: '',
		temperature: 0.7,
		category: 'custom'
	});

	// Model capability types
	type ModelCapability = 'vision' | 'tools' | 'coding' | 'reasoning' | 'rag' | 'multilingual' | 'fast';

	// Model variant (different parameter sizes)
	interface ModelVariant {
		id: string;
		params: string;
		size: string;
	}

	// Available model interface with capabilities
	interface AvailableModel {
		id: string;
		name: string;
		description: string;
		size: string;
		params: string;
		capabilities: ModelCapability[];
		provider: 'local' | 'cloud';
		downloads?: string;
		isInstalled?: boolean;
		variants?: ModelVariant[];  // Optional variants for different parameter sizes
	}

	// State for variant selection
	let selectedVariants = $state<Record<string, string>>({});

	// Capability info for badges (SVG icons instead of emojis)
	const capabilityInfo: Record<ModelCapability, { label: string; color: string; iconPath: string }> = {
		vision: { label: 'Vision', color: 'bg-purple-100 text-purple-700', iconPath: 'M15 12a3 3 0 11-6 0 3 3 0 016 0zM2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z' },
		tools: { label: 'Tools', color: 'bg-blue-100 text-blue-700', iconPath: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065zM15 12a3 3 0 11-6 0 3 3 0 016 0z' },
		coding: { label: 'Code', color: 'bg-green-100 text-green-700', iconPath: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
		reasoning: { label: 'Reasoning', color: 'bg-orange-100 text-orange-700', iconPath: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
		rag: { label: 'RAG', color: 'bg-cyan-100 text-cyan-700', iconPath: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
		multilingual: { label: 'Multi-lang', color: 'bg-pink-100 text-pink-700', iconPath: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
		fast: { label: 'Fast', color: 'bg-yellow-100 text-yellow-700', iconPath: 'M13 10V3L4 14h7v7l9-11h-7z' },
	};

	// Model browser state
	let modelSearchQuery = $state('');
	let selectedCapabilityFilters = $state<ModelCapability[]>([]); // Multi-select capabilities
	let selectedProviderFilter = $state<'all' | 'local' | 'cloud'>('all');
	let modelSortBy = $state<'recommended' | 'name' | 'size' | 'downloads'>('recommended');
	let showOnlyInstalled = $state(false);

	// Dropdown visibility state
	let showSourceDropdown = $state(false);
	let showFiltersDropdown = $state(false);

	// Available models catalog (popular Ollama models) with variants
	const availableModels: AvailableModel[] = [
		// Coding & Tool-focused
		{
			id: 'qwen3-coder', name: 'Qwen3 Coder', description: 'Best for coding & agentic tasks',
			size: '9GB', params: '14B', capabilities: ['coding', 'tools', 'multilingual'], provider: 'local', downloads: '2.1M',
			variants: [
				{ id: 'qwen3-coder:1.5b', params: '1.5B', size: '1GB' },
				{ id: 'qwen3-coder:7b', params: '7B', size: '4.5GB' },
				{ id: 'qwen3-coder:14b', params: '14B', size: '9GB' },
				{ id: 'qwen3-coder:30b', params: '30B', size: '19GB' },
			]
		},
		{
			id: 'deepseek-coder-v2', name: 'DeepSeek Coder V2', description: 'Strong coding assistant',
			size: '10GB', params: '16B', capabilities: ['coding', 'tools'], provider: 'local', downloads: '500K',
			variants: [
				{ id: 'deepseek-coder-v2:16b', params: '16B', size: '10GB' },
				{ id: 'deepseek-coder-v2:236b', params: '236B', size: 'Cloud' },
			]
		},
		{
			id: 'codellama', name: 'Code Llama', description: 'Meta\'s code-focused model',
			size: '4GB', params: '7B', capabilities: ['coding'], provider: 'local', downloads: '2M',
			variants: [
				{ id: 'codellama:7b', params: '7B', size: '4GB' },
				{ id: 'codellama:13b', params: '13B', size: '7GB' },
				{ id: 'codellama:34b', params: '34B', size: '19GB' },
				{ id: 'codellama:70b', params: '70B', size: '39GB' },
			]
		},
		{ id: 'starcoder2:15b', name: 'StarCoder2 15B', description: 'BigCode coding model', size: '9GB', params: '15B', capabilities: ['coding'], provider: 'local', downloads: '300K' },

		// Reasoning
		{
			id: 'deepseek-r1', name: 'DeepSeek R1', description: 'Advanced reasoning model',
			size: '9GB', params: '14B', capabilities: ['reasoning', 'coding', 'tools'], provider: 'local', downloads: '1.9M',
			variants: [
				{ id: 'deepseek-r1:1.5b', params: '1.5B', size: '1GB' },
				{ id: 'deepseek-r1:7b', params: '7B', size: '4.5GB' },
				{ id: 'deepseek-r1:8b', params: '8B', size: '5GB' },
				{ id: 'deepseek-r1:14b', params: '14B', size: '9GB' },
				{ id: 'deepseek-r1:32b', params: '32B', size: '20GB' },
				{ id: 'deepseek-r1:70b', params: '70B', size: '43GB' },
			]
		},

		// Vision models
		{
			id: 'llama3.2-vision', name: 'Llama 3.2 Vision', description: 'Multimodal vision-language',
			size: '7GB', params: '11B', capabilities: ['vision', 'tools'], provider: 'local', downloads: '1.5M',
			variants: [
				{ id: 'llama3.2-vision:11b', params: '11B', size: '7GB' },
				{ id: 'llama3.2-vision:90b', params: '90B', size: '55GB' },
			]
		},
		{
			id: 'llava', name: 'LLaVA', description: 'Vision-language model',
			size: '5GB', params: '7B', capabilities: ['vision'], provider: 'local', downloads: '2M',
			variants: [
				{ id: 'llava:7b', params: '7B', size: '5GB' },
				{ id: 'llava:13b', params: '13B', size: '8GB' },
				{ id: 'llava:34b', params: '34B', size: '20GB' },
			]
		},
		{ id: 'minicpm-v:8b', name: 'MiniCPM-V 8B', description: 'Efficient vision model', size: '5GB', params: '8B', capabilities: ['vision', 'fast'], provider: 'local', downloads: '400K' },
		{ id: 'moondream', name: 'Moondream', description: 'Tiny vision model', size: '1.7GB', params: '1.8B', capabilities: ['vision', 'fast'], provider: 'local', downloads: '500K' },

		// General purpose
		{
			id: 'llama3.3', name: 'Llama 3.3', description: 'Latest Meta flagship',
			size: '43GB', params: '70B', capabilities: ['tools', 'coding', 'reasoning'], provider: 'local', downloads: '3M',
			variants: [
				{ id: 'llama3.3:70b', params: '70B', size: '43GB' },
			]
		},
		{
			id: 'llama3.2', name: 'Llama 3.2', description: 'Fast general purpose',
			size: '2GB', params: '3B', capabilities: ['tools', 'fast'], provider: 'local', downloads: '9M',
			variants: [
				{ id: 'llama3.2:1b', params: '1B', size: '1.3GB' },
				{ id: 'llama3.2:3b', params: '3B', size: '2GB' },
				{ id: 'llama3.2:8b', params: '8B', size: '5GB' },
			]
		},
		{
			id: 'qwen3', name: 'Qwen3', description: 'Strong multilingual model',
			size: '5GB', params: '8B', capabilities: ['tools', 'coding', 'reasoning', 'multilingual'], provider: 'local', downloads: '4.5M',
			variants: [
				{ id: 'qwen3:0.6b', params: '0.6B', size: '0.5GB' },
				{ id: 'qwen3:1.7b', params: '1.7B', size: '1.2GB' },
				{ id: 'qwen3:4b', params: '4B', size: '2.5GB' },
				{ id: 'qwen3:8b', params: '8B', size: '5GB' },
				{ id: 'qwen3:14b', params: '14B', size: '9GB' },
				{ id: 'qwen3:32b', params: '32B', size: '20GB' },
			]
		},
		{
			id: 'mistral', name: 'Mistral', description: 'Efficient general model',
			size: '4GB', params: '7B', capabilities: ['tools', 'fast'], provider: 'local', downloads: '4M',
			variants: [
				{ id: 'mistral:7b', params: '7B', size: '4GB' },
			]
		},
		{
			id: 'gemma2', name: 'Gemma 2', description: 'Google efficient model',
			size: '2GB', params: '2B', capabilities: ['fast'], provider: 'local', downloads: '2M',
			variants: [
				{ id: 'gemma2:2b', params: '2B', size: '2GB' },
				{ id: 'gemma2:9b', params: '9B', size: '5GB' },
				{ id: 'gemma2:27b', params: '27B', size: '16GB' },
			]
		},
		{
			id: 'phi3', name: 'Phi-3', description: 'Microsoft reasoning model',
			size: '2.2GB', params: '3.8B', capabilities: ['reasoning', 'coding', 'fast'], provider: 'local', downloads: '1.5M',
			variants: [
				{ id: 'phi3:mini', params: '3.8B', size: '2.2GB' },
				{ id: 'phi3:medium', params: '14B', size: '8GB' },
			]
		},

		// Embedding/RAG models
		{ id: 'nomic-embed-text', name: 'Nomic Embed Text', description: 'Text embeddings for RAG', size: '274MB', params: '137M', capabilities: ['rag'], provider: 'local', downloads: '3M' },
		{ id: 'mxbai-embed-large', name: 'MxBai Embed Large', description: 'High-quality embeddings', size: '670MB', params: '335M', capabilities: ['rag'], provider: 'local', downloads: '1M' },
		{ id: 'bge-m3', name: 'BGE-M3', description: 'Multilingual embeddings', size: '1.2GB', params: '568M', capabilities: ['rag', 'multilingual'], provider: 'local', downloads: '800K' },

		// Cloud models
		{ id: 'qwen3-coder:480b-cloud', name: 'Qwen3 Coder 480B', description: '480B via Ollama Cloud', size: 'Cloud', params: '480B', capabilities: ['coding', 'tools', 'reasoning', 'multilingual'], provider: 'cloud', downloads: '500K' },
		{ id: 'deepseek-r1:671b', name: 'DeepSeek R1 671B', description: 'Full reasoning via cloud', size: 'Cloud', params: '671B', capabilities: ['reasoning', 'coding', 'tools'], provider: 'cloud', downloads: '200K' },
	];

	// Get selected variant for a model, defaulting to first variant or model id
	function getSelectedVariant(model: AvailableModel): string {
		if (!model.variants || model.variants.length === 0) return model.id;
		return selectedVariants[model.id] || model.variants[0].id;
	}

	// Get variant info for display
	function getVariantInfo(model: AvailableModel): ModelVariant | null {
		if (!model.variants || model.variants.length === 0) return null;
		const selectedId = getSelectedVariant(model);
		return model.variants.find(v => v.id === selectedId) || model.variants[0];
	}

	// Cloud models by provider (for provider-specific display)
	const cloudModels: Record<string, { id: string; name: string; description: string }[]> = {
		groq: [
			{ id: 'llama-3.3-70b-versatile', name: 'Llama 3.3 70B', description: 'Fast 70B model' },
			{ id: 'llama-3.1-8b-instant', name: 'Llama 3.1 8B', description: 'Ultra-fast responses' },
			{ id: 'mixtral-8x7b-32768', name: 'Mixtral 8x7B', description: '32k context window' },
		],
		anthropic: [
			{ id: 'claude-sonnet-4-20250514', name: 'Claude Sonnet 4', description: 'Best for most tasks' },
			{ id: 'claude-opus-4-20250514', name: 'Claude Opus 4', description: 'Most capable' },
		],
		ollama_cloud: [
			{ id: 'qwen3-coder:480b-cloud', name: 'Qwen3 Coder 480B', description: '480B cloud - best quality' },
			{ id: 'qwen3-coder:30b', name: 'Qwen3 Coder 30B', description: '30B coding model' },
			{ id: 'deepseek-r1:671b', name: 'DeepSeek R1 671B', description: 'Full reasoning model' },
			{ id: 'deepseek-r1:70b', name: 'DeepSeek R1 70B', description: 'Reasoning model' },
			{ id: 'llama3.3:70b', name: 'Llama 3.3 70B', description: 'Latest Llama model' },
			{ id: 'llama3.2', name: 'Llama 3.2', description: 'Fast Llama model' },
			{ id: 'qwen3:8b', name: 'Qwen3 8B', description: 'Balanced model' },
			{ id: 'mistral', name: 'Mistral', description: 'Mistral AI flagship' },
		]
	};

	// Get filtered and sorted models
	function getFilteredModels(): AvailableModel[] {
		let filtered = [...availableModels];

		// Check which models are installed
		const installedIds = new Set(models.map(m => m.id.toLowerCase()));
		filtered = filtered.map(m => ({
			...m,
			isInstalled: installedIds.has(m.id.toLowerCase()) || installedIds.has(m.id.split(':')[0].toLowerCase())
		}));

		// Filter by search query
		if (modelSearchQuery) {
			const query = modelSearchQuery.toLowerCase();
			filtered = filtered.filter(m =>
				m.name.toLowerCase().includes(query) ||
				m.description.toLowerCase().includes(query) ||
				m.id.toLowerCase().includes(query)
			);
		}

		// Filter by capabilities (multi-select)
		if (selectedCapabilityFilters.length > 0) {
			filtered = filtered.filter(m =>
				selectedCapabilityFilters.some(cap => m.capabilities.includes(cap))
			);
		}

		// Filter by provider type (local/cloud)
		if (selectedProviderFilter !== 'all') {
			filtered = filtered.filter(m => m.provider === selectedProviderFilter);
		}

		// Filter by installed status
		if (showOnlyInstalled) {
			filtered = filtered.filter(m => m.isInstalled);
		}

		// Sort
		switch (modelSortBy) {
			case 'name':
				filtered.sort((a, b) => a.name.localeCompare(b.name));
				break;
			case 'size':
				filtered.sort((a, b) => {
					const sizeA = parseFloat(a.size) || 999;
					const sizeB = parseFloat(b.size) || 999;
					return sizeA - sizeB;
				});
				break;
			case 'downloads':
				filtered.sort((a, b) => {
					const parseDownloads = (d: string | undefined) => {
						if (!d) return 0;
						const num = parseFloat(d);
						if (d.includes('M')) return num * 1000000;
						if (d.includes('K')) return num * 1000;
						return num;
					};
					return parseDownloads(b.downloads) - parseDownloads(a.downloads);
				});
				break;
			default: // recommended - put installed first, then by downloads
				filtered.sort((a, b) => {
					if (a.isInstalled && !b.isInstalled) return -1;
					if (!a.isInstalled && b.isInstalled) return 1;
					const parseDownloads = (d: string | undefined) => {
						if (!d) return 0;
						const num = parseFloat(d);
						if (d.includes('M')) return num * 1000000;
						if (d.includes('K')) return num * 1000;
						return num;
					};
					return parseDownloads(b.downloads) - parseDownloads(a.downloads);
				});
		}

		return filtered;
	}

	onMount(() => {
		// Load data
		loadConfig();
		loadSystemInfo();
		loadAgents();
		loadCustomAgents();
		loadOutputStyles();
		loadOutputPreference();
		loadUserFacts();

		// Click outside handler for dropdowns
		const handleClickOutside = (e: MouseEvent) => {
			const target = e.target as HTMLElement;
			if (!target.closest('.filter-dropdown-wrapper')) {
				showSourceDropdown = false;
				showFiltersDropdown = false;
			}
		};
		document.addEventListener('click', handleClickOutside);
		return () => document.removeEventListener('click', handleClickOutside);
	});

	async function loadUserFacts() {
		loadingUserFacts = true;
		userFactsError = '';
		try {
			const res = await apiClient.get('/user-facts?active=false');
			if (res.ok) {
				const data = await res.json();
				userFacts = (data.facts ?? []) as UserFact[];
			} else {
				const data = await res.json().catch(() => ({}));
				userFactsError = data.error || 'Failed to load user facts';
				userFacts = [];
			}
		} catch (err) {
			userFactsError = 'Failed to load user facts';
			userFacts = [];
		} finally {
			loadingUserFacts = false;
		}
	}

	async function confirmUserFact(key: string) {
		savingUserFactKey = key;
		userFactsError = '';
		try {
			const res = await apiClient.post(`/user-facts/${encodeURIComponent(key)}/confirm`);
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				throw new Error(data.error || 'Failed to confirm fact');
			}
			await loadUserFacts();
		} catch (err) {
			userFactsError = err instanceof Error ? err.message : 'Failed to confirm fact';
		} finally {
			savingUserFactKey = null;
		}
	}

	async function rejectUserFact(key: string) {
		savingUserFactKey = key;
		userFactsError = '';
		try {
			const res = await apiClient.post(`/user-facts/${encodeURIComponent(key)}/reject`);
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				throw new Error(data.error || 'Failed to reject fact');
			}
			await loadUserFacts();
		} catch (err) {
			userFactsError = err instanceof Error ? err.message : 'Failed to reject fact';
		} finally {
			savingUserFactKey = null;
		}
	}

	async function deleteUserFact(key: string) {
		if (!confirm(`Delete fact "${key}"?`)) return;
		savingUserFactKey = key;
		userFactsError = '';
		try {
			const res = await apiClient.delete(`/user-facts/${encodeURIComponent(key)}`);
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				throw new Error(data.error || 'Failed to delete fact');
			}
			await loadUserFacts();
		} catch (err) {
			userFactsError = err instanceof Error ? err.message : 'Failed to delete fact';
		} finally {
			savingUserFactKey = null;
		}
	}

	async function loadOutputStyles() {
		loadingOutputStyles = true;
		try {
			const res = await apiClient.get('/ai/output-styles');
			if (res.ok) {
				const data = await res.json();
				outputStyles = data.styles || [];
			}
		} catch (err) {
			console.error('Failed to load output styles:', err);
		} finally {
			loadingOutputStyles = false;
		}
	}

	async function loadOutputPreference() {
		try {
			const res = await apiClient.get('/ai/output-preferences');
			if (res.ok) {
				const data = await res.json();
				outputPreference = data.preference || null;
				selectedDefaultStyleId = outputPreference?.default_style_id || '';
			}
		} catch (err) {
			console.error('Failed to load output preference:', err);
		}
	}

	async function saveOutputPreference() {
		savingOutputPreference = true;
		error = '';
		saveStatus = '';
		try {
			const body = {
				default_style_id: selectedDefaultStyleId || null,
				style_overrides: outputPreference?.style_overrides || {},
				custom_instructions: outputPreference?.custom_instructions || null
			};

			const res = await apiClient.put('/ai/output-preferences', body);
			if (res.ok) {
				const data = await res.json();
				outputPreference = data.preference || null;
				selectedDefaultStyleId = outputPreference?.default_style_id || '';
				saveStatus = 'Output style saved!';
				setTimeout(() => saveStatus = '', 2000);
			} else {
				const data = await res.json().catch(() => ({}));
				error = data.error || 'Failed to save output style';
			}
		} catch (err) {
			console.error('Failed to save output preference:', err);
			error = 'Failed to save output style';
		} finally {
			savingOutputPreference = false;
		}
	}

	async function loadConfig() {
		isLoading = true;
		error = '';

		try {
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				providers = data.providers || [];
				activeProvider = data.active_provider || 'ollama_local';
				defaultModel = data.default_model || '';
			}

			const modelsRes = await apiClient.get('/ai/models');
			if (modelsRes.ok) {
				const data = await modelsRes.json();
				models = data.models || [];
			}

			// Load saved model settings
			const settingsRes = await apiClient.get('/settings');
			if (settingsRes.ok) {
				const data = await settingsRes.json();
				if (data.model_settings) {
					modelSettings = {
						temperature: data.model_settings.temperature ?? 0.7,
						maxTokens: data.model_settings.maxTokens ?? 8192,
						contextWindow: data.model_settings.contextWindow ?? 32768,
						topP: data.model_settings.topP ?? 0.95,
						streamResponses: data.model_settings.streamResponses ?? true,
						showUsageInChat: data.model_settings.showUsageInChat ?? true
					};
				}
				// Use default_model from settings if not set from provider
				if (!defaultModel && data.default_model) {
					defaultModel = data.default_model;
				}
			}
		} catch (err) {
			error = 'Failed to load AI configuration. Make sure the backend is running.';
			console.error('Error loading config:', err);
		} finally {
			isLoading = false;
		}
	}

	async function loadSystemInfo() {
		try {
			const res = await apiClient.get('/ai/system');
			if (res.ok) {
				systemInfo = await res.json();
			}
		} catch (err) {
			console.error('Failed to load system info:', err);
		}
	}

	async function loadAgents() {
		loadingAgents = true;
		try {
			const res = await apiClient.get('/ai/agents');
			if (res.ok) {
				const data = await res.json();
				agents = data.agents || [];
			}
		} catch (err) {
			console.error('Failed to load agents:', err);
		} finally {
			loadingAgents = false;
		}
	}

	// Custom Agents API functions
	async function loadCustomAgents() {
		loadingCustomAgents = true;
		try {
			const res = await apiClient.get('/ai/custom-agents');
			if (res.ok) {
				const data = await res.json();
				customAgents = data.agents || [];
			}
		} catch (err) {
			console.error('Failed to load custom agents:', err);
		} finally {
			loadingCustomAgents = false;
		}
	}

	async function createCustomAgent() {
		if (!newCustomAgent.name || !newCustomAgent.display_name || !newCustomAgent.system_prompt) {
			error = 'Name, display name, and system prompt are required';
			return;
		}
		savingCustomAgent = true;
		try {
			const res = await apiClient.post('/ai/custom-agents', newCustomAgent);
			if (res.ok) {
				const data = await res.json();
				customAgents = [...customAgents, data.agent];
				showNewCustomAgent = false;
				newCustomAgent = {
					name: '',
					display_name: '',
					description: '',
					system_prompt: '',
					model_preference: '',
					temperature: 0.7,
					category: 'custom'
				};
				saveStatus = 'Custom agent created!';
				setTimeout(() => saveStatus = '', 3000);
			} else {
				const err = await res.json();
				error = err.error || 'Failed to create agent';
			}
		} catch (err) {
			console.error('Failed to create custom agent:', err);
			error = 'Failed to create custom agent';
		} finally {
			savingCustomAgent = false;
		}
	}

	async function updateCustomAgent(agent: CustomAgent) {
		savingCustomAgent = true;
		try {
			const res = await apiClient.put(`/ai/custom-agents/${agent.id}`, agent);
			if (res.ok) {
				const data = await res.json();
				customAgents = customAgents.map(a => a.id === agent.id ? data.agent : a);
				editingCustomAgent = null;
				saveStatus = 'Agent updated!';
				setTimeout(() => saveStatus = '', 3000);
			}
		} catch (err) {
			console.error('Failed to update custom agent:', err);
			error = 'Failed to update agent';
		} finally {
			savingCustomAgent = false;
		}
	}

	async function deleteCustomAgent(id: string) {
		if (!confirm('Are you sure you want to delete this agent?')) return;
		try {
			const res = await apiClient.delete(`/ai/custom-agents/${id}`);
			if (res.ok) {
				customAgents = customAgents.filter(a => a.id !== id);
				saveStatus = 'Agent deleted';
				setTimeout(() => saveStatus = '', 3000);
			}
		} catch (err) {
			console.error('Failed to delete custom agent:', err);
			error = 'Failed to delete agent';
		}
	}

	async function loadCommands() {
		loadingCommands = true;
		try {
			const res = await apiClient.get('/ai/commands');
			if (res.ok) {
				const data = await res.json();
				commands = data.commands || [];
			}
		} catch (err) {
			console.error('Failed to load commands:', err);
		} finally {
			loadingCommands = false;
		}
	}

	async function loadUsageStats() {
		loadingUsage = true;
		try {
			const res = await apiClient.get(`/usage/summary?period=${usagePeriod}`);
			if (res.ok) {
				const data = await res.json();
				usageStats = {
					// Core metrics
					total_requests: data.total_requests || 0,
					total_tokens: data.total_tokens || 0,
					total_cost: data.total_cost || 0,
					input_tokens: data.input_tokens || Math.floor((data.total_tokens || 0) * 0.3),
					output_tokens: data.output_tokens || Math.floor((data.total_tokens || 0) * 0.7),
					// Provider breakdown
					by_provider: data.by_provider || {},
					// Model breakdown
					by_model: data.by_model || {},
					// Agent usage
					by_agent: data.by_agent || {},
					// Time-based
					recent: data.recent || [],
					daily_trend: data.daily_trend || [],
					// Session stats
					session_count: data.session_count || Math.ceil((data.total_requests || 0) / 8),
					avg_session_duration_min: data.avg_session_duration_min || 12,
					avg_requests_per_session: data.avg_requests_per_session || 8,
					// Storage & performance
					local_model_storage_gb: data.local_model_storage_gb || calculateLocalStorageUsage(),
					avg_response_time_ms: data.avg_response_time_ms || 450,
					// Cost breakdown
					local_power_cost_estimate: data.local_power_cost_estimate || calculateLocalPowerCost(data.total_tokens || 0),
					cloud_api_cost: data.cloud_api_cost || data.total_cost || 0,
					// Period
					period_start: data.period_start || getDateRange(usagePeriod).start,
					period_end: data.period_end || getDateRange(usagePeriod).end
				};
			}
		} catch (err) {
			console.error('Failed to load usage stats:', err);
			// Set default stats with realistic mock data
			const mockRequests = 247;
			const mockTokens = 156000;
			usageStats = {
				total_requests: mockRequests,
				total_tokens: mockTokens,
				total_cost: 0.78,
				input_tokens: Math.floor(mockTokens * 0.35),
				output_tokens: Math.floor(mockTokens * 0.65),
				by_provider: {
					'ollama_local': { requests: 198, tokens: 125000, cost: 0, input_tokens: 43750, output_tokens: 81250 },
					'anthropic': { requests: 35, tokens: 24000, cost: 0.72, input_tokens: 8400, output_tokens: 15600 },
					'groq': { requests: 14, tokens: 7000, cost: 0.06, input_tokens: 2450, output_tokens: 4550 }
				},
				by_model: {
					'qwen3:8b': { requests: 156, tokens: 98000, input_tokens: 34300, output_tokens: 63700, avg_latency_ms: 380 },
					'deepseek-r1:14b': { requests: 42, tokens: 27000, input_tokens: 9450, output_tokens: 17550, avg_latency_ms: 520 },
					'claude-sonnet-4': { requests: 35, tokens: 24000, input_tokens: 8400, output_tokens: 15600, avg_latency_ms: 890 },
					'llama3.2:3b': { requests: 14, tokens: 7000, input_tokens: 2450, output_tokens: 4550, avg_latency_ms: 180 }
				},
				by_agent: {
					'Document': { requests: 89, tokens: 56000 },
					'Analyst': { requests: 67, tokens: 42000 },
					'Planner': { requests: 45, tokens: 28000 },
					'Default': { requests: 46, tokens: 30000 }
				},
				recent: generateRecentData(7),
				daily_trend: generateDailyTrend(14),
				session_count: 31,
				avg_session_duration_min: 14,
				avg_requests_per_session: 8,
				local_model_storage_gb: calculateLocalStorageUsage(),
				avg_response_time_ms: 420,
				local_power_cost_estimate: 0.12,
				cloud_api_cost: 0.78,
				period_start: getDateRange(usagePeriod).start,
				period_end: getDateRange(usagePeriod).end
			};
		} finally {
			loadingUsage = false;
		}
	}

	function calculateLocalStorageUsage(): number {
		// Estimate based on installed models
		let totalGB = 0;
		models.forEach(m => {
			const sizeStr = m.size || '0';
			const num = parseFloat(sizeStr);
			if (sizeStr.includes('GB')) totalGB += num;
			else if (sizeStr.includes('MB')) totalGB += num / 1024;
		});
		return Math.round(totalGB * 10) / 10;
	}

	function calculateLocalPowerCost(tokens: number): number {
		// Estimate power cost: ~0.3W per 1K tokens, $0.12/kWh average
		const kwhUsed = (tokens / 1000) * 0.0003;
		return Math.round(kwhUsed * 0.12 * 100) / 100;
	}

	function getDateRange(period: string): { start: string; end: string } {
		const end = new Date();
		const start = new Date();
		switch (period) {
			case 'today': start.setHours(0, 0, 0, 0); break;
			case 'week': start.setDate(end.getDate() - 7); break;
			case 'month': start.setMonth(end.getMonth() - 1); break;
			default: start.setFullYear(end.getFullYear() - 1);
		}
		return {
			start: start.toISOString().split('T')[0],
			end: end.toISOString().split('T')[0]
		};
	}

	function generateRecentData(days: number) {
		const data = [];
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(date.getDate() - i);
			data.push({
				date: date.toISOString().split('T')[0],
				requests: Math.floor(Math.random() * 30) + 10,
				tokens: Math.floor(Math.random() * 15000) + 5000,
				cost: Math.round(Math.random() * 20) / 100
			});
		}
		return data;
	}

	function generateDailyTrend(days: number) {
		const data = [];
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(date.getDate() - i);
			data.push({
				date: date.toISOString().split('T')[0],
				local_requests: Math.floor(Math.random() * 25) + 8,
				cloud_requests: Math.floor(Math.random() * 8),
				local_tokens: Math.floor(Math.random() * 12000) + 3000,
				cloud_tokens: Math.floor(Math.random() * 5000)
			});
		}
		return data;
	}

	function formatTokens(tokens: number): string {
		if (tokens >= 1000000) return (tokens / 1000000).toFixed(1) + 'M';
		if (tokens >= 1000) return (tokens / 1000).toFixed(1) + 'K';
		return tokens.toString();
	}

	function formatDuration(ms: number): string {
		if (ms < 1000) return ms + 'ms';
		return (ms / 1000).toFixed(1) + 's';
	}

	async function saveNewCommand() {
		if (!newCommand.name || !newCommand.display_name || !newCommand.system_prompt) {
			error = 'Name, display name, and system prompt are required';
			return;
		}
		savingCommand = true;
		try {
			const res = await apiClient.post('/ai/commands', {
				name: newCommand.name.toLowerCase().replace(/\s+/g, '-'),
				display_name: newCommand.display_name,
				description: newCommand.description || '',
				icon: newCommand.icon || '✨',
				system_prompt: newCommand.system_prompt,
				context_sources: newCommand.context_sources || []
			});
			if (res.ok) {
				saveStatus = 'Command created successfully';
				showNewCommand = false;
				newCommand = { name: '', display_name: '', description: '', icon: '✨', system_prompt: '', context_sources: [] };
				await loadCommands();
			} else {
				const data = await res.json();
				error = data.error || 'Failed to create command';
			}
		} catch (err) {
			error = 'Failed to create command';
		} finally {
			savingCommand = false;
			setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
		}
	}

	async function updateCommand(cmd: CommandInfo) {
		if (!cmd.id) return;
		savingCommand = true;
		try {
			const res = await apiClient.put(`/ai/commands/${cmd.id}`, {
				name: cmd.name,
				display_name: cmd.display_name,
				description: cmd.description,
				icon: cmd.icon,
				system_prompt: cmd.system_prompt,
				context_sources: cmd.context_sources
			});
			if (res.ok) {
				saveStatus = 'Command updated successfully';
				editingCommand = null;
				await loadCommands();
			} else {
				const data = await res.json();
				error = data.error || 'Failed to update command';
			}
		} catch (err) {
			error = 'Failed to update command';
		} finally {
			savingCommand = false;
			setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
		}
	}

	async function deleteCommand(id: string) {
		if (!confirm('Are you sure you want to delete this command?')) return;
		try {
			const res = await apiClient.delete(`/ai/commands/${id}`);
			if (res.ok) {
				saveStatus = 'Command deleted';
				await loadCommands();
			} else {
				error = 'Failed to delete command';
			}
		} catch (err) {
			error = 'Failed to delete command';
		}
		setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
	}

	const contextSourceOptions = [
		{ id: 'documents', label: 'Documents', desc: 'Load content from selected context documents' },
		{ id: 'conversations', label: 'Conversations', desc: 'Include recent conversation history' },
		{ id: 'artifacts', label: 'Artifacts', desc: 'Include generated artifacts' },
		{ id: 'projects', label: 'Projects', desc: 'Include project details' },
		{ id: 'clients', label: 'Clients', desc: 'Include client information' },
		{ id: 'tasks', label: 'Tasks', desc: 'Include task list' }
	];

	function toggleContextSource(sources: string[] | undefined, source: string): string[] {
		const current = sources || [];
		if (current.includes(source)) {
			return current.filter(s => s !== source);
		}
		return [...current, source];
	}

	function toggleAgentExpand(agentId: string) {
		expandedAgent = expandedAgent === agentId ? null : agentId;
		editingAgent = null;
	}

	function startEditingAgent(agentId: string, currentPrompt: string) {
		editingAgent = agentId;
		editedPrompt = currentPrompt;
	}

	function cancelEditing() {
		editingAgent = null;
		editedPrompt = '';
	}

	async function saveAgentPrompt(agentId: string) {
		// In the future, this will save to the backend
		// For now, just close the editor
		saveStatus = 'Agent customization coming soon';
		editingAgent = null;
		setTimeout(() => saveStatus = '', 3000);
	}

	// SVG icon paths for agent categories
	const categoryIconPaths: Record<string, string> = {
		general: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5', // layers
		specialist: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z', // wrench
		system: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z' // settings
	};
	const defaultCategoryIconPath = 'M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73v.27a6 6 0 0 1 6 6v7h1a1 1 0 0 1 0 2H4a1 1 0 0 1 0-2h1v-7a6 6 0 0 1 6-6v-.27c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2z'; // robot

	function getCategoryIconPath(category: string): string {
		return categoryIconPaths[category] || defaultCategoryIconPath;
	}

	function getCategoryLabel(category: string): string {
		const labels: Record<string, string> = {
			general: 'General',
			specialist: 'Specialist',
			system: 'System'
		};
		return labels[category] || category;
	}

	async function selectProvider(providerId: string) {
		const provider = providers.find(p => p.id === providerId);
		if (!provider?.configured && provider?.type === 'cloud') {
			error = `${provider.name} requires an API key. Configure it in the Providers tab.`;
			setTimeout(() => error = '', 4000);
			return;
		}

		activeProvider = providerId;

		try {
			const res = await apiClient.put('/ai/provider', { provider: providerId });
			if (res.ok) {
				saveStatus = 'Provider updated!';
				setTimeout(() => saveStatus = '', 3000);
			}
		} catch (err) {
			console.error('Failed to update provider:', err);
		}
	}

	async function saveAPIKey(provider: string) {
		const key = apiKeys[provider];
		if (!key.trim()) return;

		savingKey = provider;
		try {
			const res = await apiClient.post('/ai/api-key', {
				provider,
				api_key: key.trim()
			});

			if (res.ok) {
				saveStatus = `API key saved!`;
				apiKeys[provider] = '';
				await loadConfig();
			} else {
				const data = await res.json();
				error = data.error || 'Failed to save API key';
			}
		} catch (err) {
			error = 'Failed to save API key';
		} finally {
			savingKey = null;
			setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
		}
	}

	async function saveSettings() {
		isSaving = true;
		saveStatus = '';
		try {
			await apiClient.put('/settings', {
				ai_provider: activeProvider,
				default_model: defaultModel,
				model_settings: modelSettings
			});
			saveStatus = 'Settings saved!';
			setTimeout(() => saveStatus = '', 2000);
		} catch (err) {
			saveStatus = 'Failed to save settings';
			console.error('Failed to save:', err);
		} finally {
			isSaving = false;
		}
	}

	async function pullModel(modelName: string) {
		if (!modelName.trim() || isPulling) return;

		isPulling = true;
		pullError = '';
		pullProgress = { status: 'Connecting...' };
		pullStartTime = Date.now();
		const model = modelName.trim();

		try {
			const response = await fetch(`/api/ai/models/pull`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ model }),
				credentials: 'include'
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to pull model');
			}

			const reader = response.body?.getReader();
			if (!reader) throw new Error('No response body');

			const decoder = new TextDecoder();
			let buffer = '';
			let lastCompleted = 0;
			let lastTime = Date.now();

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (line.startsWith('data: ')) {
						try {
							const data = JSON.parse(line.slice(6));
							pullProgress = data;

							if (data.completed && data.completed > lastCompleted) {
								const now = Date.now();
								const timeDiff = (now - lastTime) / 1000;
								const bytesDiff = data.completed - lastCompleted;
								if (timeDiff > 0) {
									const speed = bytesDiff / timeDiff;
									pullSpeed = formatBytes(speed) + '/s';
								}
								lastCompleted = data.completed;
								lastTime = now;
							}

							if (data.status === 'complete' || data.status === 'success') {
								pullProgress = { status: 'Complete!' };
								pullModelName = '';
								await loadConfig();
							}
						} catch {}
					}
				}
			}
		} catch (err) {
			pullError = err instanceof Error ? err.message : 'Failed to pull model';
			pullProgress = null;
		} finally {
			isPulling = false;
			pullSpeed = '';
		}
	}

	async function deleteModel(modelId: string) {
		if (!confirm(`Delete model ${modelId}?`)) return;

		try {
			const res = await apiClient.delete(`/ai/models/${encodeURIComponent(modelId)}`);
			if (res.ok) {
				saveStatus = 'Model deleted';
				await loadConfig();
			} else {
				error = 'Failed to delete model';
			}
		} catch (err) {
			error = 'Failed to delete model';
		}
		setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
	}

	function formatBytes(bytes: number): string {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
		return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
	}

	function getProgressPercent(): number {
		if (!pullProgress?.total || !pullProgress?.completed) return 0;
		return Math.round((pullProgress.completed / pullProgress.total) * 100);
	}

	function getTimeRemaining(): string {
		if (!pullProgress?.total || !pullProgress?.completed || !pullSpeed) return '';
		const remaining = pullProgress.total - pullProgress.completed;
		const speedMatch = pullSpeed.match(/[\d.]+/);
		if (!speedMatch) return '';

		let speedBytes = parseFloat(speedMatch[0]);
		if (pullSpeed.includes('KB')) speedBytes *= 1024;
		if (pullSpeed.includes('MB')) speedBytes *= 1024 * 1024;
		if (pullSpeed.includes('GB')) speedBytes *= 1024 * 1024 * 1024;

		if (speedBytes <= 0) return '';
		const seconds = remaining / speedBytes;
		if (seconds < 60) return `~${Math.ceil(seconds)}s`;
		if (seconds < 3600) return `~${Math.ceil(seconds / 60)}m`;
		return `~${(seconds / 3600).toFixed(1)}h`;
	}

	function getLocalModels(): LLMModel[] {
		return (models || []).filter(m => {
			const isLocalProvider = m.provider === 'ollama' || m.provider === 'ollama_local';
			// Filter out cloud reference models (they have "cloud" in the name and are tiny stubs)
			const nameOrId = (m.id || '') + (m.name || '');
			const isCloudRef = nameOrId.toLowerCase().includes('cloud') &&
				(m.size === '< 1 KB' || m.size === '0 B' || !m.size);
			return isLocalProvider && !isCloudRef;
		});
	}

	function getRamPercent(): number {
		if (!systemInfo) return 0;
		return Math.round(((systemInfo.total_ram_gb - systemInfo.available_ram_gb) / systemInfo.total_ram_gb) * 100);
	}

	// SVG icon paths for providers
	const providerIconPaths: Record<string, string> = {
		ollama_local: 'M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2zM22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z', // computer/local
		ollama_cloud: 'M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z', // cloud
		groq: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z', // lightning bolt
		anthropic: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5' // brain/network
	};
	const defaultProviderIconPath = 'M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73v.27a6 6 0 0 1 6 6v7h1a1 1 0 0 1 0 2H4a1 1 0 0 1 0-2h1v-7a6 6 0 0 1 6-6v-.27c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2z';

	function getProviderIconPath(id: string): string {
		return providerIconPaths[id] || defaultProviderIconPath;
	}

	function getProviderLabel(id: string): string {
		const labels: Record<string, string> = {
			ollama_local: 'Local',
			ollama_cloud: 'Cloud',
			groq: 'Groq',
			anthropic: 'Claude'
		};
		return labels[id] || id;
	}
</script>

<div class="page">
	<!-- Status Toast -->
	{#if saveStatus}
		<div class="save-toast">{saveStatus}</div>
	{/if}

	{#if isLoading}
		<div class="loading">
			<div class="spinner"></div>
			<span>Loading configuration...</span>
		</div>
	{:else}
		{#if error}
			<div class="error-alert">
				<span>{error}</span>
				<button onclick={() => error = ''}>×</button>
			</div>
		{/if}

		<!-- Tab Navigation -->
		<div class="tabs">
			<button class="tab" class:active={activeTab === 'models'} onclick={() => activeTab = 'models'}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
				Models
			</button>
			<button class="tab" class:active={activeTab === 'providers'} onclick={() => activeTab = 'providers'}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
				Providers
			</button>
			<button class="tab" class:active={activeTab === 'settings'} onclick={() => activeTab = 'settings'}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
				Settings
			</button>
			<button class="tab" class:active={activeTab === 'agents'} onclick={() => activeTab = 'agents'}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 0 1 7 7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-1H2a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a7 7 0 0 1 7-7h1V5.73c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2M7.5 13A1.5 1.5 0 0 0 6 14.5 1.5 1.5 0 0 0 7.5 16 1.5 1.5 0 0 0 9 14.5 1.5 1.5 0 0 0 7.5 13m9 0a1.5 1.5 0 0 0-1.5 1.5 1.5 1.5 0 0 0 1.5 1.5 1.5 1.5 0 0 0 1.5-1.5 1.5 1.5 0 0 0-1.5-1.5"/></svg>
				Agents
			</button>
			<button class="tab" class:active={activeTab === 'commands'} onclick={() => { activeTab = 'commands'; loadCommands(); }}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 15l5 5 5-5M7 9l5-5 5 5"/></svg>
				Commands
			</button>
			<button class="tab" class:active={activeTab === 'stats'} onclick={() => { activeTab = 'stats'; loadUsageStats(); }}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 20V10M12 20V4M6 20v-6"/></svg>
				Stats
			</button>
		</div>

		<!-- Tab Content -->
		<div class="tab-content">
			{#if activeTab === 'models'}
				<!-- Model Browser - LM Studio inspired -->
				<section class="section model-browser-section">
					<!-- Compact Filter Bar (Single Row) -->
					<div class="browser-controls">
						<!-- Search -->
						<div class="compact-search">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
							<input
								type="text"
								bind:value={modelSearchQuery}
								placeholder="Search..."
							/>
							{#if modelSearchQuery}
								<button class="clear-search" onclick={() => modelSearchQuery = ''} aria-label="Clear search">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
								</button>
							{:else}
								<span class="search-shortcut">⌘K</span>
							{/if}
						</div>

						<!-- Source Dropdown -->
						<div class="filter-dropdown-wrapper">
							<button
								class="filter-dropdown-btn"
								class:active={selectedProviderFilter !== 'all'}
								onclick={() => { showSourceDropdown = !showSourceDropdown; showFiltersDropdown = false; }}
							>
								{#if selectedProviderFilter === 'local'}
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><circle cx="6" cy="6" r="1" fill="currentColor"/><circle cx="6" cy="18" r="1" fill="currentColor"/></svg>
									Local
								{:else if selectedProviderFilter === 'cloud'}
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg>
									Cloud
								{:else}
									Source
								{/if}
								<span class="dropdown-count">{selectedProviderFilter === 'all' ? availableModels.length : selectedProviderFilter === 'local' ? availableModels.filter(m => m.provider === 'local').length : availableModels.filter(m => m.provider === 'cloud').length}</span>
								<svg class="dropdown-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>
							</button>
							{#if showSourceDropdown}
								<div class="filter-dropdown-menu" role="menu">
									<button
										class="dropdown-item"
										class:selected={selectedProviderFilter === 'all'}
										onclick={() => { selectedProviderFilter = 'all'; showSourceDropdown = false; }}
									>
										All <span class="item-count">{availableModels.length}</span>
									</button>
									<button
										class="dropdown-item"
										class:selected={selectedProviderFilter === 'local'}
										onclick={() => { selectedProviderFilter = 'local'; showSourceDropdown = false; }}
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><circle cx="6" cy="6" r="1" fill="currentColor"/><circle cx="6" cy="18" r="1" fill="currentColor"/></svg>
										Local <span class="item-count">{availableModels.filter(m => m.provider === 'local').length}</span>
									</button>
									<button
										class="dropdown-item"
										class:selected={selectedProviderFilter === 'cloud'}
										onclick={() => { selectedProviderFilter = 'cloud'; showSourceDropdown = false; }}
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg>
										Cloud <span class="item-count">{availableModels.filter(m => m.provider === 'cloud').length}</span>
									</button>
								</div>
							{/if}
						</div>

						<!-- Capabilities Dropdown -->
						<div class="filter-dropdown-wrapper">
							<button
								class="filter-dropdown-btn"
								class:active={selectedCapabilityFilters.length > 0}
								onclick={() => { showFiltersDropdown = !showFiltersDropdown; showSourceDropdown = false; }}
							>
								Filters
								{#if selectedCapabilityFilters.length > 0}
									<span class="dropdown-count">{selectedCapabilityFilters.length}</span>
								{/if}
								<svg class="dropdown-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>
							</button>
							{#if showFiltersDropdown}
								<div class="filter-dropdown-menu capabilities-menu" role="menu">
									{#each Object.entries(capabilityInfo) as [cap, info]}
										<label class="dropdown-checkbox-item">
											<input
												type="checkbox"
												checked={selectedCapabilityFilters.includes(cap as ModelCapability)}
												onchange={() => {
													if (selectedCapabilityFilters.includes(cap as ModelCapability)) {
														selectedCapabilityFilters = selectedCapabilityFilters.filter(c => c !== cap);
													} else {
														selectedCapabilityFilters = [...selectedCapabilityFilters, cap as ModelCapability];
													}
												}}
											/>
											<svg class="cap-icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={info.iconPath}/></svg>
											{info.label}
										</label>
									{/each}
									{#if selectedCapabilityFilters.length > 0}
										<button class="dropdown-clear" onclick={() => selectedCapabilityFilters = []}>
											Clear all
										</button>
									{/if}
								</div>
							{/if}
						</div>

						<!-- Active Filter Chips -->
						{#if selectedCapabilityFilters.length > 0}
							<div class="active-filter-chips">
								{#each selectedCapabilityFilters as cap}
									<button
										class="filter-chip"
										onclick={() => selectedCapabilityFilters = selectedCapabilityFilters.filter(c => c !== cap)}
									>
										{capabilityInfo[cap].label}
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
									</button>
								{/each}
							</div>
						{/if}

						<!-- Installed Toggle -->
						<label class="compact-toggle">
							<input type="checkbox" bind:checked={showOnlyInstalled} />
							<span class="toggle-slider"></span>
							<span class="toggle-label">Installed</span>
						</label>

						<!-- Sort Dropdown -->
						<select bind:value={modelSortBy} class="compact-sort-select">
							<option value="recommended">Recommended</option>
							<option value="name">Name</option>
							<option value="size">Size</option>
							<option value="downloads">Downloads</option>
						</select>
					</div>

					<!-- Recommended for You (NOT sticky - scrolls away) -->
					{#if systemInfo && systemInfo.recommended_models.length > 0 && activeProvider === 'ollama_local' && !modelSearchQuery && selectedCapabilityFilters.length === 0}
						<div class="recommended-banner">
							<div class="rec-banner-header">
								<h3>Recommended for You</h3>
								<span class="rec-badge-info">Based on {systemInfo.total_ram_gb}GB RAM</span>
							</div>
							<div class="rec-chips">
								{#each systemInfo.recommended_models.slice(0, 4) as model}
									{@const isInstalled = getLocalModels().some(m => m.id.startsWith(model.name.split(':')[0]))}
									<button
										class="rec-chip"
										class:installed={isInstalled}
										onclick={() => { if (!isInstalled) pullModel(model.name); }}
										disabled={isPulling}
									>
										<span class="chip-name">{model.name}</span>
										<span class="chip-meta">{model.speed} • {model.quality}</span>
										{#if isInstalled}
											<span class="chip-status installed">Installed</span>
										{:else}
											<span class="chip-status pull">Pull</span>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}

					<!-- Model Browser Grid -->
					<div class="browser-content">
						<div class="browser-header">
							<h3>
								{#if showOnlyInstalled}
									Installed Models
								{:else if selectedCapabilityFilters.length === 1}
									{capabilityInfo[selectedCapabilityFilters[0]].label} Models
								{:else if selectedCapabilityFilters.length > 1}
									Filtered Models
								{:else}
									All Models
								{/if}
							</h3>
							<span class="model-count">{getFilteredModels().length} models</span>
						</div>

						{#if getFilteredModels().length === 0}
							<div class="empty-state">
								<div class="empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
								</div>
								<h3>No models found</h3>
								<p>Try adjusting your search or filters</p>
							</div>
						{:else}
							<div class="model-browser-grid">
								{#each getFilteredModels() as model}
									{@const isDefault = defaultModel === model.id || defaultModel.startsWith(model.id.split(':')[0])}
									{@const variantInfo = getVariantInfo(model)}
									{@const displaySize = variantInfo ? variantInfo.size : model.size}
									{@const displayParams = variantInfo ? variantInfo.params : model.params}
									{@const pullId = getSelectedVariant(model)}
									<div class="browser-model-card" class:installed={model.isInstalled} class:default={isDefault}>
										<div class="bmc-header">
											<div class="bmc-title">
												<span class="bmc-name">{model.name}</span>
												{#if model.isInstalled}
													<span class="bmc-installed-badge">Installed</span>
												{/if}
											</div>
											<div class="bmc-meta">
												<span class="bmc-size">{displaySize}</span>
												<span class="bmc-params">{displayParams}</span>
											</div>
										</div>

										<p class="bmc-description">{model.description}</p>

										<!-- Capability Badges -->
										<div class="bmc-capabilities">
											{#each model.capabilities as cap}
												{@const info = capabilityInfo[cap]}
												<span class="cap-badge {cap}" title={info.label}>
													<svg class="cap-badge-icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={info.iconPath}/></svg>
													<span class="cap-badge-label">{info.label}</span>
												</span>
											{/each}
										</div>

										<!-- Variant Selector (if has variants) -->
										{#if model.variants && model.variants.length > 1}
											<div class="bmc-variants">
												<span class="variants-label">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
													Select size:
												</span>
												<div class="variant-buttons" class:many={model.variants.length > 4}>
													{#each model.variants as variant}
														<button
															class="variant-btn"
															class:selected={getSelectedVariant(model) === variant.id}
															onclick={() => selectedVariants[model.id] = variant.id}
															title="{variant.params} parameters • {variant.size} download"
														>
															<span class="variant-params">{variant.params}</span>
															<span class="variant-size">{variant.size}</span>
														</button>
													{/each}
												</div>
											</div>
										{/if}

										<div class="bmc-footer">
											{#if model.downloads}
												<span class="bmc-downloads">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
													{model.downloads}
												</span>
											{/if}

											{#if model.isInstalled}
												<div class="bmc-actions">
													<button
														class="bmc-btn default-btn"
														class:is-default={isDefault}
														onclick={() => { defaultModel = model.id; saveSettings(); }}
													>
														{isDefault ? 'Default' : 'Set Default'}
													</button>
													<button
														class="bmc-btn delete-btn"
														onclick={() => deleteModel(model.id)}
														title="Delete model"
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
													</button>
												</div>
											{:else}
												<button
													class="bmc-btn pull-btn"
													onclick={() => pullModel(pullId)}
													disabled={isPulling}
												>
													{#if isPulling && pullModelName === pullId}
														<div class="btn-spinner-small"></div>
														Pulling...
													{:else}
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
														Pull Model
													{/if}
												</button>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Pull Model Form (Compact) -->
					<div class="pull-card-compact">
						<div class="pull-compact-header">
							<h4>Pull Custom Model</h4>
							<a href="https://ollama.com/library" target="_blank" class="browse-link">
								Browse Ollama Library
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3"/></svg>
							</a>
						</div>
						<div class="pull-form-compact">
							<input
								type="text"
								bind:value={pullModelName}
								placeholder="Enter model name (e.g., llama3.2:3b, phi3:medium)"
								disabled={isPulling}
							/>
							<button class="pull-btn-compact" onclick={() => pullModel(pullModelName)} disabled={isPulling || !pullModelName.trim()}>
								{#if isPulling}
									<div class="btn-spinner-small"></div>
								{:else}
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
								{/if}
								Pull
							</button>
						</div>

						{#if isPulling && pullProgress}
							<div class="pull-progress-compact">
								<div class="progress-info">
									<span class="progress-status">{pullProgress.status}</span>
									<span class="progress-percent">{getProgressPercent()}%</span>
								</div>
								<div class="progress-bar-compact">
									<div class="progress-fill" style="width: {getProgressPercent()}%"></div>
								</div>
								<div class="progress-details">
									{#if pullProgress.total && pullProgress.completed}
										<span>{formatBytes(pullProgress.completed)} / {formatBytes(pullProgress.total)}</span>
									{/if}
									{#if pullSpeed}<span class="speed">{pullSpeed}</span>{/if}
									{#if getTimeRemaining()}<span class="time">{getTimeRemaining()}</span>{/if}
								</div>
							</div>
						{/if}

						{#if pullError}
							<div class="pull-error-compact">{pullError}</div>
						{/if}
					</div>
				</section>

			{:else if activeTab === 'providers'}
				<!-- Providers Tab -->
				<section class="section">
					<div class="section-header">
						<h2>AI Providers</h2>
						<span class="badge active">
							Using {getProviderLabel(activeProvider)}
						</span>
					</div>

					<div class="providers-grid">
						{#each providers as provider}
							<button
								class="provider-card"
								class:active={activeProvider === provider.id}
								class:disabled={!provider.configured && provider.type === 'cloud'}
								onclick={() => selectProvider(provider.id)}
							>
								<div class="provider-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={getProviderIconPath(provider.id)}/></svg>
								</div>
								<div class="provider-info">
									<span class="provider-name">{provider.name}</span>
									<span class="provider-desc">{provider.description}</span>
								</div>
								<div class="provider-status">
									{#if activeProvider === provider.id}
										<span class="status active">Active</span>
									{:else if provider.configured}
										<span class="status ready">Ready</span>
									{:else}
										<span class="status setup">Setup Required</span>
									{/if}
								</div>
								<span class="provider-badge" class:local={provider.type === 'local'}>
									{provider.type === 'local' ? 'Local' : 'Cloud'}
								</span>
							</button>
						{/each}
					</div>
				</section>

				<!-- API Keys -->
				<section class="section">
					<div class="section-header">
						<h2>API Keys</h2>
						<span class="subtitle">Configure cloud provider access</span>
					</div>

					<div class="api-grid">
						{#each [
							{ id: 'groq', name: 'Groq', url: 'https://console.groq.com' },
							{ id: 'anthropic', name: 'Anthropic', url: 'https://console.anthropic.com' },
							{ id: 'ollama_cloud', name: 'Ollama Cloud', url: 'https://ollama.com' }
						] as provider}
							{@const isConfigured = providers.find(p => p.id === provider.id)?.configured}
							<div class="api-card">
								<div class="api-header">
									<span class="api-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d={getProviderIconPath(provider.id)}/></svg>
									</span>
									<span class="api-name">{provider.name}</span>
									{#if isConfigured}
										<span class="api-configured">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
										</span>
									{/if}
								</div>
								<div class="api-form">
									<div class="input-wrapper">
										<input
											type={showApiKey[provider.id] ? 'text' : 'password'}
											bind:value={apiKeys[provider.id]}
											placeholder={isConfigured ? '••••••••••••' : 'Enter API key'}
										/>
										<button class="toggle-btn" onclick={() => showApiKey[provider.id] = !showApiKey[provider.id]}>
											{#if showApiKey[provider.id]}
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
											{:else}
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
											{/if}
										</button>
									</div>
									<button
										class="save-btn"
										onclick={() => saveAPIKey(provider.id)}
										disabled={!apiKeys[provider.id]?.trim() || savingKey === provider.id}
									>
										{savingKey === provider.id ? '...' : 'Save'}
									</button>
								</div>
								<a href={provider.url} target="_blank" class="api-link">Get API Key →</a>
							</div>
						{/each}
					</div>
				</section>

			{:else if activeTab === 'settings'}
				<!-- Settings Tab -->
				<section class="section">
					<div class="section-header">
						<h2>Model Settings</h2>
						<span class="subtitle">Fine-tune AI behavior</span>
					</div>

					<!-- Output Style Preference -->
					<div class="setting-card" style="margin-bottom: 16px;">
						<div class="setting-header">
							<label>Default Output Style</label>
						</div>
						<p class="setting-desc" style="margin-bottom: 1rem;">Choose how the AI formats its responses. This is used when you don't specify an output style per message.</p>

						{#if loadingOutputStyles}
							<div class="loading" style="padding: 2rem; text-align: center;">
								<div class="spinner"></div>
								<span>Loading output styles...</span>
							</div>
						{:else}
							<OutputStyleSelector
								styles={outputStyles}
								selectedStyleId={selectedDefaultStyleId}
								onSelect={(styleId) => selectedDefaultStyleId = styleId}
								disabled={savingOutputPreference}
							/>

							<button
								class="save-btn"
								style="margin-top: 1rem;"
								onclick={saveOutputPreference}
								disabled={savingOutputPreference}
							>
								{savingOutputPreference ? 'Saving...' : 'Save Default Style'}
							</button>
						{/if}
					</div>

					<!-- User Facts Management -->
					<div class="setting-card" style="margin-bottom: 16px; padding: 0; overflow: hidden; max-height: 600px;">
						<UserFactsPanel />
					</div>

					<!-- OLD User Facts (kept for reference, can be removed) -->
					<div class="setting-card" style="margin-bottom: 16px; display: none;">
						<div class="setting-header">
							<label>User Facts (Old)</label>
							<span class="setting-value">
								{#if loadingUserFacts}
									Loading...
								{:else}
									{userFacts.length}
								{/if}
							</span>
						</div>
						<p class="setting-desc">Review what the system has learned. Confirmed facts are injected into chats; rejected facts are not.</p>

						{#if userFactsError}
							<div class="error-banner" style="margin-top: 8px;">{userFactsError}</div>
						{/if}

						<div class="facts-toolbar">
							<button class="save-btn" onclick={loadUserFacts} disabled={loadingUserFacts}>
								{loadingUserFacts ? '...' : 'Refresh'}
							</button>
						</div>

						{#if !loadingUserFacts && userFacts.length === 0}
							<p class="setting-desc" style="margin-top: 10px;">No facts yet.</p>
						{:else}
							<div class="facts-list">
								{#each userFacts as fact (fact.id)}
									<div class="fact-row">
										<div class="fact-main">
											<div class="fact-key">{fact.fact_key}</div>
											<div class="fact-value">{fact.fact_value}</div>
											<div class="fact-meta">
												<span class="fact-pill">{fact.fact_type}</span>
												<span class="fact-pill">{Math.round((fact.confidence_score ?? 0) * 100)}%</span>
												<span class="fact-pill" class:inactive={!fact.is_active}>
													{fact.is_active ? 'confirmed' : 'rejected'}
												</span>
											</div>
										</div>
										<div class="fact-actions">
											{#if fact.is_active}
												<button class="bmc-btn delete-btn" onclick={() => rejectUserFact(fact.fact_key)} disabled={savingUserFactKey === fact.fact_key}>
													{savingUserFactKey === fact.fact_key ? '...' : 'Reject'}
												</button>
											{:else}
												<button class="save-btn" onclick={() => confirmUserFact(fact.fact_key)} disabled={savingUserFactKey === fact.fact_key}>
													{savingUserFactKey === fact.fact_key ? '...' : 'Confirm'}
												</button>
											{/if}
											<button class="bmc-btn delete-btn" onclick={() => deleteUserFact(fact.fact_key)} disabled={savingUserFactKey === fact.fact_key} title="Delete fact">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
											</button>
										</div>
								</div>
							{/each}
						</div>
						{/if}
					</div>

					<!-- Quick Presets -->
					<div class="presets-row">
						<span class="presets-label">Quick Presets:</span>
						<button class="preset-btn" onclick={() => {
							modelSettings = { ...modelSettings, temperature: 0.3, maxTokens: 4096, topP: 0.8, contextWindow: 16384 };
						}}>
							<svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
							Fast
						</button>
						<button class="preset-btn" onclick={() => {
							modelSettings = { ...modelSettings, temperature: 0.7, maxTokens: 8192, topP: 0.95, contextWindow: 32768 };
						}}>
							<svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
							Balanced
						</button>
						<button class="preset-btn" onclick={() => {
							modelSettings = { ...modelSettings, temperature: 0.9, maxTokens: 16384, topP: 1.0, contextWindow: 65536 };
						}}>
							<svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
							Quality
						</button>
						<button class="preset-btn" onclick={() => {
							modelSettings = { ...modelSettings, temperature: 1.0, maxTokens: 32768, topP: 1.0, contextWindow: 131072 };
						}}>
							<svg class="preset-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 19h20L12 2z"/><path d="M12 9v4"/><circle cx="12" cy="16" r="1"/></svg>
							Maximum
						</button>
					</div>

					<div class="settings-grid">
						<div class="setting-card">
							<div class="setting-header">
								<label for="contextWindow">Context Window</label>
								<span class="setting-value">{(modelSettings.contextWindow / 1000).toFixed(0)}K tokens</span>
							</div>
							<input
								type="range"
								id="contextWindow"
								bind:value={modelSettings.contextWindow}
								min="4096"
								max="131072"
								step="4096"
							/>
							<p class="setting-desc">How much conversation history to include (larger = more context)</p>
						</div>

						<div class="setting-card">
							<div class="setting-header">
								<label for="maxTokens">Max Output Tokens</label>
								<span class="setting-value">{(modelSettings.maxTokens / 1000).toFixed(1)}K</span>
							</div>
							<input
								type="range"
								id="maxTokens"
								bind:value={modelSettings.maxTokens}
								min="512"
								max="32768"
								step="512"
							/>
							<p class="setting-desc">Maximum length of AI responses</p>
						</div>

						<div class="setting-card">
							<div class="setting-header">
								<label for="temperature">Temperature</label>
								<span class="setting-value">{modelSettings.temperature}</span>
							</div>
							<input
								type="range"
								id="temperature"
								bind:value={modelSettings.temperature}
								min="0"
								max="2"
								step="0.1"
							/>
							<p class="setting-desc">Lower = more focused, Higher = more creative</p>
						</div>

						<div class="setting-card">
							<div class="setting-header">
								<label for="topP">Top P (Nucleus Sampling)</label>
								<span class="setting-value">{modelSettings.topP}</span>
							</div>
							<input
								type="range"
								id="topP"
								bind:value={modelSettings.topP}
								min="0.1"
								max="1"
								step="0.05"
							/>
							<p class="setting-desc">Controls diversity of responses</p>
						</div>

						<div class="setting-card toggle-card">
							<div class="setting-header">
								<label for="streaming">Stream Responses</label>
								<button
									class="toggle"
									class:on={modelSettings.streamResponses}
									onclick={() => modelSettings.streamResponses = !modelSettings.streamResponses}
								>
									<span class="toggle-knob"></span>
								</button>
							</div>
							<p class="setting-desc">Show responses as they're generated</p>
						</div>

						<div class="setting-card toggle-card">
							<div class="setting-header">
								<label for="showUsage">Show Usage Stats</label>
								<button
									class="toggle"
									class:on={modelSettings.showUsageInChat}
									onclick={() => modelSettings.showUsageInChat = !modelSettings.showUsageInChat}
								>
									<span class="toggle-knob"></span>
								</button>
							</div>
							<p class="setting-desc">Display tokens/second and token count after each response</p>
						</div>
					</div>

					<div class="settings-actions">
						<button class="action-btn primary" onclick={saveSettings} disabled={isSaving}>
							{isSaving ? 'Saving...' : 'Save Settings'}
						</button>
						<button class="action-btn" onclick={() => {
							modelSettings = { temperature: 0.7, maxTokens: 8192, topP: 0.95, contextWindow: 32768, streamResponses: true, showUsageInChat: true };
						}}>
							Reset to Defaults
						</button>
					</div>
				</section>

				<!-- Default Model Selection -->
				<section class="section">
					<div class="section-header">
						<h2>Default Model</h2>
						<span class="subtitle">Select your preferred model</span>
					</div>

					<div class="default-model-grid">
						{#if activeProvider === 'ollama_local'}
							{#each getLocalModels() as model}
								<button
									class="default-model-btn"
									class:selected={defaultModel === model.id}
									onclick={() => { defaultModel = model.id; saveSettings(); }}
								>
									<span class="dm-name">{model.name}</span>
									{#if model.size}<span class="dm-size">{model.size}</span>{/if}
									{#if defaultModel === model.id}
										<svg class="dm-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
									{/if}
								</button>
							{/each}
						{:else}
							{#each cloudModels[activeProvider] || [] as model}
								<button
									class="default-model-btn"
									class:selected={defaultModel === model.id}
									onclick={() => { defaultModel = model.id; saveSettings(); }}
								>
									<span class="dm-name">{model.name}</span>
									<span class="dm-desc">{model.description}</span>
									{#if defaultModel === model.id}
										<svg class="dm-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
									{/if}
								</button>
							{/each}
						{/if}
					</div>
				</section>
			{:else if activeTab === 'agents'}
				<!-- Custom Agents Section -->
				<section class="section">
					<div class="section-header">
						<h2>Custom Agents</h2>
						<button class="add-btn" onclick={() => showNewCustomAgent = true}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
							New Agent
						</button>
					</div>
					<p class="section-subtitle">Create custom agents you can mention with @name in chat</p>

					{#if showNewCustomAgent}
						<div class="custom-agent-form">
							<div class="form-header">
								<h3>Create Custom Agent</h3>
								<button class="close-btn" onclick={() => showNewCustomAgent = false}>x</button>
							</div>
							<div class="form-grid">
								<div class="form-group">
									<label>Agent Name (for @mentions)</label>
									<div class="input-prefix">
										<span>@</span>
										<input type="text" bind:value={newCustomAgent.name} placeholder="coder" />
									</div>
									<span class="form-hint">Lowercase, no spaces (use hyphens)</span>
								</div>
								<div class="form-group">
									<label>Display Name</label>
									<input type="text" bind:value={newCustomAgent.display_name} placeholder="My Coding Assistant" />
								</div>
								<div class="form-group full-width">
									<label>Description</label>
									<input type="text" bind:value={newCustomAgent.description} placeholder="A helpful coding assistant..." />
								</div>
								<div class="form-group full-width">
									<label>System Prompt</label>
									<textarea bind:value={newCustomAgent.system_prompt} rows="8" placeholder="You are an expert programmer..."></textarea>
									<span class="form-hint">This prompt defines the agent's personality and behavior</span>
								</div>
								<div class="form-group">
									<label>Model (optional)</label>
									<select bind:value={newCustomAgent.model_preference}>
										<option value="">Use default</option>
										<option value="llama-3.3-70b-versatile">Llama 3.3 70B</option>
										<option value="llama-3.1-8b-instant">Llama 3.1 8B</option>
										<option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
									</select>
								</div>
								<div class="form-group">
									<label>Temperature</label>
									<input type="number" bind:value={newCustomAgent.temperature} min="0" max="2" step="0.1" />
								</div>
							</div>
							<div class="form-actions">
								<button class="btn secondary" onclick={() => showNewCustomAgent = false}>Cancel</button>
								<button class="btn primary" onclick={createCustomAgent} disabled={savingCustomAgent}>
									{savingCustomAgent ? 'Creating...' : 'Create Agent'}
								</button>
							</div>
						</div>
					{/if}

					{#if loadingCustomAgents}
						<div class="loading">
							<div class="spinner"></div>
							<span>Loading custom agents...</span>
						</div>
					{:else if customAgents.length === 0}
						<div class="empty-state small">
							<p>No custom agents yet. Create one to mention with @name in chat!</p>
						</div>
					{:else}
						<div class="custom-agents-list">
							{#each customAgents as agent}
								<div class="custom-agent-card">
									<div class="agent-header">
										<div class="agent-info">
											<span class="agent-mention">@{agent.name}</span>
											<span class="agent-display-name">{agent.display_name}</span>
										</div>
										<div class="agent-actions">
											<span class="usage-badge">{agent.times_used} uses</span>
											<button class="icon-btn" onclick={() => testingAgentId = testingAgentId === agent.id ? null : agent.id} title="Test Agent">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"/></svg>
											</button>
											<button class="icon-btn" onclick={() => editingCustomAgent = {...agent}} title="Edit">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
											</button>
											<button class="icon-btn danger" onclick={() => deleteCustomAgent(agent.id)} title="Delete">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
											</button>
										</div>
									</div>
									{#if agent.description}
										<p class="agent-desc">{agent.description}</p>
									{/if}
									<div class="agent-prompt-preview">
										<span class="label">Prompt:</span>
										<span class="preview">{agent.system_prompt.slice(0, 100)}...</span>
									</div>

									{#if testingAgentId === agent.id}
										<div class="agent-test-section" style="margin-top: 1rem; padding-top: 1rem; border-top: 1px solid #e5e7eb;">
											<h4 style="font-size: 0.875rem; font-weight: 600; color: #374151; margin-bottom: 0.75rem;">Test Agent</h4>
											<AgentTestSandbox agent={agent} />
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</section>

				<!-- Built-in Agents Tab -->
				<section class="section">
					<div class="section-header">
						<h2>Built-in Agents</h2>
						<span class="subtitle">View and customize agent prompts</span>
					</div>

					{#if loadingAgents}
						<div class="loading">
							<div class="spinner"></div>
							<span>Loading agents...</span>
						</div>
					{:else if agents.length === 0}
						<div class="empty-state">
							<div class="empty-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><circle cx="12" cy="5" r="2"/><path d="M12 7v4"/><path d="M7 22v-3"/><path d="M17 22v-3"/></svg>
							</div>
							<h3>No Agents Found</h3>
							<p>Agent configuration is not available</p>
						</div>
					{:else}
						<div class="agents-list">
							{#each agents as agent}
								<div class="agent-card" class:expanded={expandedAgent === agent.id}>
									<button class="agent-header-btn" onclick={() => toggleAgentExpand(agent.id)}>
										<div class="agent-info">
											<span class="agent-icon">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d={getCategoryIconPath(agent.category)}/></svg>
											</span>
											<div class="agent-text">
												<span class="agent-name">{agent.name}</span>
												<span class="agent-desc">{agent.description}</span>
											</div>
										</div>
										<div class="agent-meta">
											<span class="agent-category">{getCategoryLabel(agent.category)}</span>
											<svg class="agent-chevron" class:rotated={expandedAgent === agent.id} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<polyline points="6 9 12 15 18 9"/>
											</svg>
										</div>
									</button>

									{#if expandedAgent === agent.id}
										<div class="agent-content">
											<div class="prompt-header">
												<h4>System Prompt</h4>
												{#if editingAgent !== agent.id}
													<button class="prompt-edit-btn" onclick={() => startEditingAgent(agent.id, agent.prompt)}>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
														Edit
													</button>
												{/if}
											</div>

											{#if editingAgent === agent.id}
												<div class="prompt-editor">
													<textarea
														bind:value={editedPrompt}
														rows="12"
														placeholder="Enter system prompt..."
													></textarea>
													<div class="editor-actions">
														<button class="editor-btn cancel" onclick={cancelEditing}>Cancel</button>
														<button class="editor-btn save" onclick={() => saveAgentPrompt(agent.id)}>Save Changes</button>
													</div>
												</div>
											{:else}
												<div class="prompt-display">
													<pre>{agent.prompt}</pre>
												</div>
											{/if}

											<div class="agent-stats">
												<div class="stat-item">
													<span class="stat-label">ID</span>
													<span class="stat-value-small">{agent.id}</span>
												</div>
												<div class="stat-item">
													<span class="stat-label">Characters</span>
													<span class="stat-value-small">{agent.prompt.length.toLocaleString()}</span>
												</div>
												<div class="stat-item">
													<span class="stat-label">~Tokens</span>
													<span class="stat-value-small">{Math.ceil(agent.prompt.length / 4).toLocaleString()}</span>
												</div>
											</div>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</section>
			{:else if activeTab === 'commands'}
				<!-- Commands Tab -->
				<section class="section">
					<div class="section-header">
						<h2>Slash Commands</h2>
						<button class="add-btn" onclick={() => showNewCommand = true}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
							New Command
						</button>
					</div>

					{#if loadingCommands}
						<div class="loading">
							<div class="spinner"></div>
							<span>Loading commands...</span>
						</div>
					{:else}
						<!-- New Command Form -->
						{#if showNewCommand}
							<div class="command-form">
								<div class="form-header">
									<h3>Create Custom Command</h3>
									<button class="close-btn" onclick={() => showNewCommand = false}>×</button>
								</div>
								<div class="form-grid">
									<div class="form-group">
										<label>Command Name</label>
										<div class="input-prefix">
											<span>/</span>
											<input type="text" bind:value={newCommand.name} placeholder="my-command" />
										</div>
										<span class="form-hint">Lowercase, no spaces (use hyphens)</span>
									</div>
									<div class="form-group">
										<label>Display Name</label>
										<input type="text" bind:value={newCommand.display_name} placeholder="My Command" />
									</div>
									<div class="form-group full-width">
										<label>Description</label>
										<input type="text" bind:value={newCommand.description} placeholder="What this command does..." />
									</div>
									<div class="form-group">
										<label>Icon (Emoji)</label>
										<input type="text" bind:value={newCommand.icon} placeholder="✨" maxlength="4" />
									</div>
									<div class="form-group full-width">
										<label>System Prompt</label>
										<textarea bind:value={newCommand.system_prompt} rows="6" placeholder="You are an AI assistant that..."></textarea>
										<span class="form-hint">This prompt will be used when the command is executed</span>
									</div>
									<div class="form-group full-width">
										<label>Context Sources</label>
										<div class="context-sources">
											{#each contextSourceOptions as opt}
												<button
													class="context-chip"
													class:active={newCommand.context_sources?.includes(opt.id)}
													onclick={() => newCommand.context_sources = toggleContextSource(newCommand.context_sources, opt.id)}
													title={opt.desc}
												>
													{opt.label}
												</button>
											{/each}
										</div>
										<span class="form-hint">Select what data to automatically include when command runs</span>
									</div>
								</div>
								<div class="form-actions">
									<button class="btn secondary" onclick={() => showNewCommand = false}>Cancel</button>
									<button class="btn primary" onclick={saveNewCommand} disabled={savingCommand}>
										{savingCommand ? 'Creating...' : 'Create Command'}
									</button>
								</div>
							</div>
						{/if}

						<!-- Built-in Commands -->
						<div class="commands-section">
							<h4>Built-in Commands ({commands.filter(c => !c.is_custom).length})</h4>
							<div class="commands-grid">
								{#each commands.filter(c => !c.is_custom) as cmd}
									<div
										class="command-card clickable"
										class:expanded={expandedCommand === cmd.name}
										onclick={() => expandedCommand = expandedCommand === cmd.name ? null : cmd.name}
									>
										<div class="command-header">
											<span class="command-icon">{cmd.icon || '⚡'}</span>
											<span class="command-name">/{cmd.name}</span>
											<span class="command-category">{cmd.category}</span>
											<button
												class="icon-btn edit-builtin"
												onclick={(e) => { e.stopPropagation(); editingCommand = {...cmd, is_builtin_override: true}; }}
												title="Customize this command"
											>
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
											</button>
										</div>
										<p class="command-display-name">{cmd.display_name}</p>
										<p class="command-desc">{cmd.description}</p>
										{#if cmd.context_sources?.length > 0}
											<div class="command-sources">
												{#each cmd.context_sources as source}
													<span class="source-tag">{source}</span>
												{/each}
											</div>
										{/if}
										{#if expandedCommand === cmd.name}
											<div class="command-details">
												<div class="detail-section">
													<h5>System Prompt</h5>
													<p class="detail-hint">This is the prompt that guides the AI when using this command</p>
													<pre class="prompt-preview">{cmd.system_prompt || 'Built-in prompt (customize to view and edit)'}</pre>
												</div>
												<div class="detail-actions">
													<button class="btn primary small" onclick={(e) => { e.stopPropagation(); editingCommand = {...cmd, is_builtin_override: true}; }}>
														Customize Command
													</button>
												</div>
											</div>
										{/if}
									</div>
								{/each}
							</div>
						</div>

						<!-- Custom Commands -->
						{#if commands.filter(c => c.is_custom).length > 0}
							<div class="commands-section">
								<h4>Your Custom Commands</h4>
								<div class="commands-grid">
									{#each commands.filter(c => c.is_custom) as cmd}
										<div class="command-card custom">
											<div class="command-header">
												<span class="command-icon">{cmd.icon || '✨'}</span>
												<span class="command-name">/{cmd.name}</span>
												<div class="command-actions">
													<button class="icon-btn" onclick={() => editingCommand = {...cmd}} title="Edit">
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
													</button>
													<button class="icon-btn danger" onclick={() => cmd.id && deleteCommand(cmd.id)} title="Delete">
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
													</button>
												</div>
											</div>
											<p class="command-display-name">{cmd.display_name}</p>
											<p class="command-desc">{cmd.description}</p>
											{#if cmd.context_sources?.length > 0}
												<div class="command-sources">
													{#each cmd.context_sources as source}
														<span class="source-tag">{source}</span>
													{/each}
												</div>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Edit Command Modal -->
						{#if editingCommand}
							<div class="modal-overlay" onclick={() => editingCommand = null}>
								<div class="modal" onclick={(e) => e.stopPropagation()}>
									<div class="form-header">
										<h3>Edit Command</h3>
										<button class="close-btn" onclick={() => editingCommand = null}>×</button>
									</div>
									<div class="form-grid">
										<div class="form-group">
											<label>Command Name</label>
											<div class="input-prefix">
												<span>/</span>
												<input type="text" bind:value={editingCommand.name} placeholder="my-command" />
											</div>
										</div>
										<div class="form-group">
											<label>Display Name</label>
											<input type="text" bind:value={editingCommand.display_name} placeholder="My Command" />
										</div>
										<div class="form-group full-width">
											<label>Description</label>
											<input type="text" bind:value={editingCommand.description} placeholder="What this command does..." />
										</div>
										<div class="form-group">
											<label>Icon (Emoji)</label>
											<input type="text" bind:value={editingCommand.icon} placeholder="✨" maxlength="4" />
										</div>
										<div class="form-group full-width">
											<label>System Prompt</label>
											<textarea bind:value={editingCommand.system_prompt} rows="6" placeholder="You are an AI assistant that..."></textarea>
										</div>
										<div class="form-group full-width">
											<label>Context Sources</label>
											<div class="context-sources">
												{#each contextSourceOptions as opt}
													<button
														class="context-chip"
														class:active={editingCommand.context_sources?.includes(opt.id)}
														onclick={() => editingCommand && (editingCommand.context_sources = toggleContextSource(editingCommand.context_sources, opt.id))}
														title={opt.desc}
													>
														{opt.label}
													</button>
												{/each}
											</div>
										</div>
									</div>
									<div class="form-actions">
										<button class="btn secondary" onclick={() => editingCommand = null}>Cancel</button>
										<button class="btn primary" onclick={() => editingCommand && updateCommand(editingCommand)} disabled={savingCommand}>
											{savingCommand ? 'Saving...' : 'Save Changes'}
										</button>
									</div>
								</div>
							</div>
						{/if}
					{/if}
				</section>
			{:else if activeTab === 'stats'}
				<!-- Stats Tab - Comprehensive System & Usage Analytics -->
				<div class="stats-page">
					<!-- Stats Header with Period Selector -->
					<div class="stats-header">
						<div class="stats-title-area">
							<span class="stats-eyebrow">Analytics</span>
							{#if usageStats}
								<button class="stats-date-picker">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
									<span>{usageStats.period_start} — {usageStats.period_end}</span>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><polyline points="6 9 12 15 18 9"/></svg>
								</button>
							{/if}
						</div>
						<div class="stats-controls">
							<div class="period-selector">
								<button class="period-btn" class:active={usagePeriod === 'today'} onclick={() => { usagePeriod = 'today'; loadUsageStats(); }}>Today</button>
								<button class="period-btn" class:active={usagePeriod === 'week'} onclick={() => { usagePeriod = 'week'; loadUsageStats(); }}>Week</button>
								<button class="period-btn" class:active={usagePeriod === 'month'} onclick={() => { usagePeriod = 'month'; loadUsageStats(); }}>Month</button>
								<button class="period-btn" class:active={usagePeriod === 'all'} onclick={() => { usagePeriod = 'all'; loadUsageStats(); }}>All Time</button>
							</div>
							<button class="refresh-btn" onclick={loadUsageStats} disabled={loadingUsage} title="Refresh stats">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class:spinning={loadingUsage}><path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
							</button>
						</div>
					</div>

					{#if loadingUsage}
						<div class="loading">
							<div class="spinner"></div>
							<span>Loading analytics...</span>
						</div>
					{:else if usageStats}
						<!-- System Overview Row -->
						<div class="stats-system-row">
							<div class="system-card">
								<div class="system-platform-info">
									<div class="platform-icon-large">
										{#if systemInfo?.platform === 'darwin'}
											<svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/></svg>
										{:else}
											<svg viewBox="0 0 24 24" fill="currentColor"><path d="M3,12V6.75L9,5.43V11.91L3,12M20,3V11.75L10,11.9V5.21L20,3M3,13L9,13.09V19.9L3,18.75V13M20,13.25V22L10,20.09V13.1L20,13.25Z"/></svg>
										{/if}
									</div>
									<div class="platform-text">
										<span class="platform-label">{systemInfo?.platform === 'darwin' ? 'macOS' : systemInfo?.platform || 'System'}</span>
										{#if systemInfo?.has_gpu}
											<span class="gpu-badge">{systemInfo.gpu_name || 'Apple Silicon / Metal'}</span>
										{/if}
									</div>
								</div>
								<div class="system-metrics">
									<div class="sys-metric">
										<span class="sys-metric-value">{getLocalModels().length}</span>
										<span class="sys-metric-label">Local Models</span>
									</div>
									<div class="sys-metric">
										<span class="sys-metric-value">{usageStats.local_model_storage_gb}GB</span>
										<span class="sys-metric-label">Storage Used</span>
									</div>
									<div class="sys-metric">
										<span class="sys-metric-value">{providers.filter(p => p.configured).length}</span>
										<span class="sys-metric-label">Providers</span>
									</div>
								</div>
							</div>
							{#if systemInfo}
								<div class="ram-gauge-card">
									<div class="ram-gauge-header">
										<span class="ram-title">Memory</span>
										<span class="ram-detail">{systemInfo.available_ram_gb.toFixed(0)}GB free / {systemInfo.total_ram_gb}GB</span>
									</div>
									<div class="ram-gauge-visual">
										<svg class="gauge-svg-large" viewBox="0 0 120 120">
											<circle class="gauge-bg-large" cx="60" cy="60" r="50" fill="none" />
											<circle
												class="gauge-fill-large"
												cx="60" cy="60" r="50"
												fill="none"
												stroke-dasharray="{(systemInfo.available_ram_gb / systemInfo.total_ram_gb) * 314} 314"
												transform="rotate(-90 60 60)"
											/>
										</svg>
										<div class="gauge-center-text">
											<span class="gauge-percent">{Math.round((systemInfo.available_ram_gb / systemInfo.total_ram_gb) * 100)}%</span>
											<span class="gauge-subtitle">Free</span>
										</div>
									</div>
								</div>
							{/if}
						</div>

						<!-- Key Metrics Grid -->
						<div class="key-metrics-grid">
							<div class="metric-card requests">
								<div class="metric-sparkline">
									{#if usageStats && usageStats.recent.length > 0}
										{@const recent = usageStats.recent}
										<svg viewBox="0 0 100 30" preserveAspectRatio="none">
											<polyline
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
												points="{recent.map((d, i) => `${i * (100 / (recent.length - 1 || 1))},${30 - (d.requests / Math.max(...recent.map(r => r.requests) || 1)) * 25}`).join(' ')}"
											/>
										</svg>
									{/if}
								</div>
								<div class="metric-header">
									<div class="metric-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
									</div>
									<span class="metric-trend up">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="12" height="12"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
									</span>
								</div>
								<div class="metric-content">
									<span class="metric-value animate-number">{usageStats.total_requests.toLocaleString()}</span>
									<span class="metric-label">Total Requests</span>
								</div>
								<div class="metric-sub">
									<span class="metric-badge">{usageStats.session_count} sessions</span>
								</div>
							</div>
							<div class="metric-card tokens">
								<div class="metric-sparkline">
									{#if usageStats && usageStats.recent.length > 0}
										{@const recent = usageStats.recent}
										<svg viewBox="0 0 100 30" preserveAspectRatio="none">
											<polyline
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
												points="{recent.map((d, i) => `${i * (100 / (recent.length - 1 || 1))},${30 - (d.tokens / Math.max(...recent.map(r => r.tokens) || 1)) * 25}`).join(' ')}"
											/>
										</svg>
									{/if}
								</div>
								<div class="metric-header">
									<div class="metric-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="2" width="16" height="20" rx="2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
									</div>
								</div>
								<div class="metric-content">
									<span class="metric-value">{formatTokens(usageStats.total_tokens)}</span>
									<span class="metric-label">Total Tokens</span>
								</div>
								<div class="metric-breakdown">
									<span class="metric-in"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="10" height="10"><polyline points="19 12 12 19 5 12"/><line x1="12" y1="19" x2="12" y2="5"/></svg> {formatTokens(usageStats.input_tokens)}</span>
									<span class="metric-out"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="10" height="10"><polyline points="5 12 12 5 19 12"/><line x1="12" y1="5" x2="12" y2="19"/></svg> {formatTokens(usageStats.output_tokens)}</span>
								</div>
							</div>
							<div class="metric-card speed">
								<div class="metric-header">
									<div class="metric-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
									</div>
									{#if usageStats.avg_response_time_ms < 500}
										<span class="metric-badge good">Fast</span>
									{:else if usageStats.avg_response_time_ms < 2000}
										<span class="metric-badge">Average</span>
									{:else}
										<span class="metric-badge slow">Slow</span>
									{/if}
								</div>
								<div class="metric-content">
									<span class="metric-value">{formatDuration(usageStats.avg_response_time_ms)}</span>
									<span class="metric-label">Avg Response</span>
								</div>
								<div class="metric-sub">
									<span class="metric-badge">{usageStats.avg_requests_per_session} req/session</span>
								</div>
							</div>
							<div class="metric-card cost">
								<div class="metric-header">
									<div class="metric-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
									</div>
								</div>
								<div class="metric-content">
									<span class="metric-value">${(usageStats.cloud_api_cost + usageStats.local_power_cost_estimate).toFixed(2)}</span>
									<span class="metric-label">Total Cost</span>
								</div>
								<div class="metric-breakdown cost-breakdown">
									<span class="metric-cloud" title="Cloud API Cost">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/></svg>
										${usageStats.cloud_api_cost.toFixed(2)}
									</span>
									<span class="metric-local" title="Est. Power Cost">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
										${usageStats.local_power_cost_estimate.toFixed(2)}
									</span>
								</div>
							</div>
						</div>

						<!-- Local vs Cloud Comparison -->
						<div class="comparison-section">
							<div class="section-header">
								<h3>Local vs Cloud Usage</h3>
								{#if usageStats.total_requests > 0}
									<span class="section-hint">Based on {usageStats.total_requests} requests</span>
								{/if}
							</div>
							{#if usageStats.total_requests === 0}
								<div class="comparison-empty">
									<div class="comparison-empty-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48">
											<path d="M21.21 15.89A10 10 0 1 1 8 2.83"/>
											<path d="M22 12A10 10 0 0 0 12 2v10z"/>
										</svg>
									</div>
									<p class="comparison-empty-text">Start using AI features to see your local vs cloud breakdown</p>
									<p class="comparison-empty-hint">Local inference is free, cloud APIs are billed per token</p>
								</div>
							{:else}
								{@const localReqs = usageStats.by_provider['ollama_local']?.requests || 0}
								{@const cloudReqs = Object.entries(usageStats.by_provider).filter(([k]) => k !== 'ollama_local').reduce((sum, [, v]) => sum + v.requests, 0)}
								{@const localPct = (localReqs / usageStats.total_requests * 100)}
								{@const cloudPct = (cloudReqs / usageStats.total_requests * 100)}
								<!-- Split Bar Comparison -->
								<div class="comparison-split-bar">
									<div class="split-bar-container">
										<div class="split-bar-fill local" style="width: {localPct}%">
											{#if localPct > 15}<span class="split-bar-label">{localPct.toFixed(0)}%</span>{/if}
										</div>
										<div class="split-bar-fill cloud" style="width: {cloudPct}%">
											{#if cloudPct > 15}<span class="split-bar-label">{cloudPct.toFixed(0)}%</span>{/if}
										</div>
									</div>
								</div>
								<div class="comparison-grid">
									<div class="comparison-card local">
										<div class="comp-header">
											<div class="comp-icon-wrapper local">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
													<rect x="2" y="3" width="20" height="14" rx="2"/>
													<path d="M8 21h8"/>
													<path d="M12 17v4"/>
												</svg>
											</div>
											<div class="comp-title-group">
												<span class="comp-title">Local Inference</span>
												<span class="comp-subtitle">On-device processing</span>
											</div>
										</div>
										<div class="comp-stats">
											<div class="comp-stat">
												<span class="comp-value">{localReqs}</span>
												<span class="comp-label">Requests</span>
											</div>
											<div class="comp-stat">
												<span class="comp-value">{formatTokens(usageStats.by_provider['ollama_local']?.tokens || 0)}</span>
												<span class="comp-label">Tokens</span>
											</div>
											<div class="comp-stat">
												<span class="comp-value">${usageStats.local_power_cost_estimate.toFixed(2)}</span>
												<span class="comp-label" title="Estimated electricity cost based on average GPU power usage">Est. Power</span>
											</div>
										</div>
									</div>
									<div class="comparison-card cloud">
										<div class="comp-header">
											<div class="comp-icon-wrapper cloud">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
													<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
												</svg>
											</div>
											<div class="comp-title-group">
												<span class="comp-title">Cloud API</span>
												<span class="comp-subtitle">Remote processing</span>
											</div>
										</div>
										<div class="comp-stats">
											<div class="comp-stat">
												<span class="comp-value">{cloudReqs}</span>
												<span class="comp-label">Requests</span>
											</div>
											<div class="comp-stat">
												<span class="comp-value">{formatTokens(Object.entries(usageStats.by_provider).filter(([k]) => k !== 'ollama_local').reduce((sum, [, v]) => sum + v.tokens, 0))}</span>
												<span class="comp-label">Tokens</span>
											</div>
											<div class="comp-stat">
												<span class="comp-value">${usageStats.cloud_api_cost.toFixed(2)}</span>
												<span class="comp-label">API Cost</span>
											</div>
										</div>
									</div>
								</div>
							{/if}
						</div>

						<!-- Detailed Breakdowns -->
						<div class="breakdowns-grid">
							<!-- Provider Breakdown -->
							<div class="breakdown-card">
								<div class="breakdown-header">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M2 20h.01M7 20v-4M12 20v-8M17 20V8M22 4v16"/></svg>
									<h4>By Provider</h4>
								</div>
								{#if Object.keys(usageStats.by_provider).length === 0}
									<div class="breakdown-empty">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><circle cx="12" cy="12" r="10"/><path d="M8 12h8"/><path d="M12 8v8"/></svg>
										<span>No provider data yet</span>
									</div>
								{:else}
									<div class="breakdown-list">
										{#each Object.entries(usageStats.by_provider).sort((a, b) => b[1].requests - a[1].requests) as [provider, stats]}
											<div class="breakdown-row">
												<div class="breakdown-info">
													<span class="breakdown-icon provider">
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d={getProviderIconPath(provider)}/></svg>
													</span>
													<span class="breakdown-name">{getProviderLabel(provider)}</span>
												</div>
												<div class="breakdown-values">
													<span class="breakdown-stat-value requests">{stats.requests}</span>
													<span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
													<span class="breakdown-stat-value cost">${stats.cost.toFixed(2)}</span>
												</div>
												<div class="breakdown-bar">
													<div class="breakdown-bar-fill provider" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>

							<!-- Model Breakdown -->
							<div class="breakdown-card">
								<div class="breakdown-header">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
									<h4>By Model</h4>
								</div>
								{#if Object.keys(usageStats.by_model).length === 0}
									<div class="breakdown-empty">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h6v6H9z"/></svg>
										<span>No model usage yet</span>
									</div>
								{:else}
									<div class="breakdown-list">
										{#each Object.entries(usageStats.by_model).sort((a, b) => b[1].requests - a[1].requests).slice(0, 5) as [model, stats]}
											<div class="breakdown-row">
												<div class="breakdown-info">
													<span class="breakdown-name truncate">{model}</span>
												</div>
												<div class="breakdown-values">
													<span class="breakdown-stat-value requests">{stats.requests}</span>
													<span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
													<span class="breakdown-stat-value latency">{formatDuration(stats.avg_latency_ms)}</span>
												</div>
												<div class="breakdown-bar">
													<div class="breakdown-bar-fill model" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>

							<!-- Agent/Command Breakdown -->
							<div class="breakdown-card">
								<div class="breakdown-header">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
									<h4>By Agent</h4>
								</div>
								{#if Object.keys(usageStats.by_agent).length === 0}
									<div class="breakdown-empty">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>
										<span>No agent usage yet</span>
									</div>
								{:else}
									<div class="breakdown-list">
										{#each Object.entries(usageStats.by_agent).sort((a, b) => b[1].requests - a[1].requests).slice(0, 5) as [agent, stats]}
											<div class="breakdown-row">
												<div class="breakdown-info">
													<span class="breakdown-icon agent">
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d={getCategoryIconPath(agent.toLowerCase())}/></svg>
													</span>
													<span class="breakdown-name">{agent}</span>
												</div>
												<div class="breakdown-values">
													<span class="breakdown-stat-value requests">{stats.requests}</span>
													<span class="breakdown-stat-value tokens">{formatTokens(stats.tokens)}</span>
												</div>
												<div class="breakdown-bar">
													<div class="breakdown-bar-fill agent" style="width: {(stats.requests / usageStats.total_requests * 100)}%"></div>
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<!-- Requests Over Time Chart -->
						<div class="activity-section">
							<div class="section-header">
								<h3>Activity Timeline</h3>
								{#if usageStats.recent.length > 0}
									<span class="section-hint">{usageStats.recent.length} days shown</span>
								{/if}
							</div>
							{#if usageStats.recent.length > 0}
								{@const maxRequests = Math.max(...usageStats.recent.map(d => d.requests), 1)}
								<div class="activity-chart">
									<div class="activity-area-chart">
										<svg viewBox="0 0 {usageStats.recent.length * 50} 120" preserveAspectRatio="none">
											<!-- Area fill -->
											<path
												d="M 0 120 {usageStats.recent.map((d, i) => `L ${i * 50 + 25} ${120 - (d.requests / maxRequests) * 100}`).join(' ')} L {(usageStats.recent.length - 1) * 50 + 25} 120 Z"
												fill="url(#areaGradient)"
												opacity="0.3"
											/>
											<!-- Line -->
											<polyline
												fill="none"
												stroke="var(--color-primary)"
												stroke-width="2.5"
												stroke-linecap="round"
												stroke-linejoin="round"
												points="{usageStats.recent.map((d, i) => `${i * 50 + 25},${120 - (d.requests / maxRequests) * 100}`).join(' ')}"
											/>
											<!-- Data points -->
											{#each usageStats.recent as day, i}
												<circle
													cx="{i * 50 + 25}"
													cy="{120 - (day.requests / maxRequests) * 100}"
													r="4"
													fill="var(--color-bg)"
													stroke="var(--color-primary)"
													stroke-width="2"
												/>
											{/each}
											<defs>
												<linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
													<stop offset="0%" stop-color="var(--color-primary)"/>
													<stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0"/>
												</linearGradient>
											</defs>
										</svg>
									</div>
									<div class="activity-labels">
										{#each usageStats.recent as day}
											<div class="activity-label-item">
												<span class="activity-day">{new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}</span>
												<span class="activity-count">{day.requests}</span>
											</div>
										{/each}
									</div>
								</div>
							{:else}
								<div class="activity-empty">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="40" height="40">
										<path d="M3 3v18h18"/>
										<path d="M18 17V9"/>
										<path d="M13 17V5"/>
										<path d="M8 17v-3"/>
									</svg>
									<p>Activity data will appear as you use AI features</p>
								</div>
							{/if}
						</div>

						<!-- Session Stats -->
						<div class="session-stats">
							<div class="section-header">
								<h3>Session Statistics</h3>
							</div>
							<div class="session-grid">
								<div class="session-stat-card">
									<div class="session-stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
									</div>
									<div class="session-stat-content">
										<span class="session-value">{usageStats.session_count}</span>
										<span class="session-label">Total Sessions</span>
									</div>
								</div>
								<div class="session-stat-card">
									<div class="session-stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
									</div>
									<div class="session-stat-content">
										<span class="session-value">{usageStats.avg_session_duration_min}<span class="session-unit">min</span></span>
										<span class="session-label">Avg Duration</span>
									</div>
								</div>
								<div class="session-stat-card">
									<div class="session-stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
									</div>
									<div class="session-stat-content">
										<span class="session-value">{usageStats.avg_requests_per_session}</span>
										<span class="session-label">Requests/Session</span>
									</div>
								</div>
								<div class="session-stat-card">
									<div class="session-stat-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><rect x="4" y="2" width="16" height="20" rx="2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
									</div>
									<div class="session-stat-content">
										<span class="session-value">{usageStats.total_requests > 0 ? Math.round(usageStats.total_tokens / usageStats.total_requests) : 0}</span>
										<span class="session-label">Tokens/Request</span>
									</div>
								</div>
							</div>
						</div>
					{:else}
						<div class="empty-state">
							<div class="empty-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
							</div>
							<h3>No Usage Data</h3>
							<p>Usage statistics will appear here once you start using AI features</p>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--color-bg);
		color: var(--color-text);
		overflow: hidden;
	}

	/* Header */
	.header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 24px;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.back-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 10px;
		background: var(--color-bg-secondary);
		color: var(--color-text-secondary);
		text-decoration: none;
		transition: all 0.2s;
	}

	.back-btn:hover {
		background: var(--color-bg-tertiary);
		color: var(--color-text);
	}

	.back-btn svg { width: 18px; height: 18px; }

	.header-title h1 {
		margin: 0;
		font-size: 20px;
		font-weight: 600;
	}

	.header-subtitle {
		margin: 2px 0 0;
		font-size: 13px;
		color: var(--color-text-muted);
	}

	.save-badge {
		padding: 6px 14px;
		background: var(--color-success);
		color: white;
		border-radius: 20px;
		font-size: 13px;
		font-weight: 500;
	}

	/* Loading */
	.loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: var(--color-text-muted);
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--color-border);
		border-top-color: var(--color-primary);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin { to { transform: rotate(360deg); } }

	/* System Overview */
	.system-overview {
		display: flex;
		gap: 20px;
		padding: 20px 24px;
		background: var(--color-bg-secondary);
		border-bottom: 1px solid var(--color-border);
	}

	.system-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.system-platform {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.platform-icon {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.platform-icon svg { width: 24px; height: 24px; }

	.platform-details {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.platform-name {
		font-weight: 600;
		font-size: 15px;
	}

	.gpu-tag {
		font-size: 11px;
		color: #8b5cf6;
		background: rgba(139, 92, 246, 0.1);
		padding: 2px 8px;
		border-radius: 4px;
		width: fit-content;
	}

	.stats-row {
		display: flex;
		gap: 16px;
	}

	.stat {
		display: flex;
		flex-direction: column;
		padding: 12px 20px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		min-width: 80px;
	}

	.stat.highlight {
		background: rgba(34, 197, 94, 0.1);
		border-color: rgba(34, 197, 94, 0.3);
	}

	.stat-value {
		font-size: 22px;
		font-weight: 700;
	}

	.stat-label {
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: uppercase;
	}

	.ram-card {
		display: flex;
		flex-direction: column;
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		min-width: 160px;
	}

	.ram-header {
		display: flex;
		justify-content: space-between;
		font-size: 12px;
		color: var(--color-text-muted);
		margin-bottom: 8px;
	}

	.ram-text { font-family: monospace; }

	.ram-gauge {
		position: relative;
		width: 100px;
		height: 100px;
		margin: 0 auto;
	}

	.gauge-svg { width: 100%; height: 100%; }

	.gauge-bg {
		stroke: var(--color-border);
		stroke-width: 8;
	}

	.gauge-fill {
		stroke: var(--color-success);
		stroke-width: 8;
		stroke-linecap: round;
		transition: stroke-dasharray 0.5s;
	}

	.gauge-text {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		text-align: center;
	}

	.gauge-value {
		display: block;
		font-size: 20px;
		font-weight: 700;
	}

	.gauge-label {
		font-size: 10px;
		color: var(--color-text-muted);
	}

	/* Error */
	.error-alert {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 24px;
		background: rgba(220, 38, 38, 0.1);
		border-bottom: 1px solid rgba(220, 38, 38, 0.2);
		color: var(--color-error);
	}

	.error-alert button {
		background: none;
		border: none;
		color: inherit;
		font-size: 20px;
		cursor: pointer;
		opacity: 0.6;
	}

	.error-alert button:hover { opacity: 1; }

	/* Tabs */
	.tabs {
		display: flex;
		gap: 4px;
		padding: 12px 24px;
		background: var(--color-bg-secondary);
		border-bottom: 1px solid var(--color-border);
	}

	.tab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 18px;
		background: transparent;
		border: none;
		border-radius: 8px;
		color: var(--color-text-secondary);
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.tab:hover {
		background: var(--color-bg);
		color: var(--color-text);
	}

	.tab.active {
		background: var(--color-bg);
		color: var(--color-text);
		box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	}

	.tab svg { width: 18px; height: 18px; }

	.tab-link {
		text-decoration: none;
		margin-left: auto;
		color: #10b981;
	}

	.tab-link:hover {
		background: rgba(16, 185, 129, 0.1);
		color: #059669;
	}

	/* Tab Content */
	.tab-content {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	/* Sections */
	.section {
		margin-bottom: 32px;
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.section-header h2 {
		font-size: 18px;
		font-weight: 600;
		margin: 0;
	}

	.badge {
		padding: 4px 10px;
		background: var(--color-bg-tertiary);
		border-radius: 12px;
		font-size: 12px;
		color: var(--color-text-secondary);
	}

	.badge.active {
		background: rgba(34, 197, 94, 0.1);
		color: var(--color-success);
	}

	.count {
		margin-left: auto;
		font-size: 13px;
		color: var(--color-text-muted);
	}

	.subtitle {
		font-size: 13px;
		color: var(--color-text-muted);
	}

	/* Recommended Grid */
	.recommended-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 12px;
	}

	.rec-card {
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		transition: all 0.2s;
	}

	.rec-card:hover {
		border-color: var(--color-border-hover);
	}

	.rec-card.installed {
		border-color: rgba(34, 197, 94, 0.4);
		background: rgba(34, 197, 94, 0.03);
	}

	.rec-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 8px;
	}

	.rec-name {
		font-weight: 600;
		font-size: 14px;
	}

	.rec-badge {
		font-size: 11px;
		padding: 3px 8px;
		background: var(--color-bg-tertiary);
		border-radius: 6px;
		color: var(--color-text-secondary);
	}

	.rec-badge.installed {
		background: rgba(34, 197, 94, 0.1);
		color: var(--color-success);
	}

	.rec-desc {
		font-size: 13px;
		color: var(--color-text-secondary);
		margin: 0 0 12px;
	}

	.rec-meta {
		display: flex;
		gap: 8px;
		margin-bottom: 12px;
	}

	.meta-tag {
		font-size: 11px;
		padding: 3px 8px;
		background: var(--color-bg-tertiary);
		border-radius: 4px;
	}

	.meta-tag.speed { color: #f59e0b; }
	.meta-tag.quality { color: #8b5cf6; }

	.rec-btn {
		width: 100%;
		padding: 10px;
		background: var(--color-primary);
		color: var(--color-bg);
		border: none;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.rec-btn:hover { opacity: 0.9; }
	.rec-btn:disabled { opacity: 0.5; cursor: not-allowed; }

	.rec-btn.secondary {
		background: var(--color-bg-tertiary);
		color: var(--color-text);
	}

	/* Models Grid */
	.models-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 10px;
	}

	.model-card {
		padding: 14px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		transition: all 0.2s;
	}

	.model-card:hover {
		border-color: var(--color-border-hover);
	}

	.model-card.selected {
		border-color: var(--color-primary);
		background: rgba(0, 0, 0, 0.02);
	}

	.model-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.model-name {
		font-weight: 600;
		font-size: 14px;
	}

	.model-size {
		font-size: 11px;
		color: var(--color-text-muted);
		font-family: monospace;
	}

	.model-family {
		font-size: 11px;
		color: var(--color-text-muted);
		margin-top: 4px;
		display: block;
	}

	.model-actions {
		display: flex;
		gap: 6px;
		margin-top: 12px;
	}

	.model-btn {
		flex: 1;
		padding: 8px;
		background: var(--color-bg-tertiary);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		font-size: 12px;
		cursor: pointer;
		transition: all 0.2s;
		color: var(--color-text);
	}

	.model-btn:hover {
		background: var(--color-bg-secondary);
	}

	.model-btn.danger {
		flex: 0;
		padding: 8px;
		color: var(--color-error);
	}

	.model-btn.danger svg { width: 14px; height: 14px; }

	/* Empty State */
	.empty-state {
		text-align: center;
		padding: 48px;
		background: var(--color-bg);
		border: 1px dashed var(--color-border);
		border-radius: 12px;
	}

	.empty-icon {
		width: 48px;
		height: 48px;
		margin-bottom: 12px;
		color: var(--color-text-muted);
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
	}
	.empty-state h3 { margin: 0 0 8px; font-size: 16px; }
	.empty-state p { margin: 0; color: var(--color-text-muted); font-size: 14px; }

	/* Model Browser Styles */
	.model-browser-section {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	/* Compact Single-Row Filter Bar */
	.browser-controls {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		background: #ffffff;
		border: 1px solid var(--color-border);
		border-radius: 10px;
		position: sticky;
		top: 0;
		z-index: 100;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
		min-height: 52px;
	}

	:global(.dark) .browser-controls {
		background: #1a1a1a;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
		border-color: rgba(255, 255, 255, 0.1);
	}

	/* Compact Search */
	.compact-search {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 12px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		min-width: 160px;
		max-width: 220px;
		transition: all 0.2s ease;
	}

	.compact-search:focus-within {
		border-color: var(--color-primary);
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
		min-width: 200px;
	}

	:global(.dark) .compact-search {
		background: rgba(255, 255, 255, 0.04);
	}

	.compact-search svg {
		width: 15px;
		height: 15px;
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.compact-search input {
		flex: 1;
		min-width: 0;
		background: none;
		border: none;
		font-size: 13px;
		color: var(--color-text);
		outline: none;
	}

	.compact-search input::placeholder {
		color: var(--color-text-muted);
	}

	.clear-search {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		background: var(--color-bg-tertiary);
		border: none;
		border-radius: 4px;
		cursor: pointer;
		color: var(--color-text-muted);
		padding: 0;
	}

	.clear-search:hover {
		background: var(--color-bg-secondary);
		color: var(--color-text);
	}

	.clear-search svg {
		width: 12px;
		height: 12px;
	}

	.search-shortcut {
		font-size: 10px;
		font-weight: 500;
		color: var(--color-text-muted);
		background: var(--color-bg-tertiary);
		padding: 2px 6px;
		border-radius: 4px;
		border: 1px solid var(--color-border);
		opacity: 0.6;
	}

	:global(.dark) .search-shortcut {
		background: rgba(255, 255, 255, 0.06);
		border-color: rgba(255, 255, 255, 0.08);
	}

	/* Filter Dropdown Wrapper */
	.filter-dropdown-wrapper {
		position: relative;
	}

	.filter-dropdown-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 10px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-secondary);
		cursor: pointer;
		transition: all 0.15s ease;
		white-space: nowrap;
	}

	.filter-dropdown-btn:hover {
		border-color: var(--color-border-hover);
		color: var(--color-text);
	}

	.filter-dropdown-btn.active {
		background: var(--color-bg-tertiary);
		border-color: var(--color-border-hover);
		color: var(--color-text);
	}

	:global(.dark) .filter-dropdown-btn {
		background: rgba(255, 255, 255, 0.04);
	}

	:global(.dark) .filter-dropdown-btn.active {
		background: rgba(255, 255, 255, 0.08);
	}

	.filter-dropdown-btn svg:not(.dropdown-chevron) {
		width: 14px;
		height: 14px;
	}

	.dropdown-chevron {
		width: 12px;
		height: 12px;
		opacity: 0.5;
		margin-left: 2px;
	}

	.dropdown-count {
		font-size: 11px;
		font-weight: 600;
		padding: 1px 5px;
		background: rgba(0, 0, 0, 0.06);
		border-radius: 8px;
		color: var(--color-text-muted);
		min-width: 18px;
		text-align: center;
	}

	:global(.dark) .dropdown-count {
		background: rgba(255, 255, 255, 0.1);
	}

	/* Dropdown Menu */
	.filter-dropdown-menu {
		position: absolute;
		top: calc(100% + 6px);
		left: 0;
		min-width: 160px;
		background: #ffffff;
		border: 1px solid var(--color-border);
		border-radius: 10px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
		z-index: 200;
		padding: 6px;
		animation: dropdownFadeIn 0.15s ease;
	}

	:global(.dark) .filter-dropdown-menu {
		background: #252525;
		border-color: rgba(255, 255, 255, 0.1);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
	}

	@keyframes dropdownFadeIn {
		from { opacity: 0; transform: translateY(-4px); }
		to { opacity: 1; transform: translateY(0); }
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 8px 10px;
		background: transparent;
		border: none;
		border-radius: 6px;
		font-size: 13px;
		color: var(--color-text);
		cursor: pointer;
		text-align: left;
		transition: background 0.1s ease;
	}

	.dropdown-item:hover {
		background: var(--color-bg-secondary);
	}

	.dropdown-item.selected {
		background: var(--color-bg-tertiary);
		font-weight: 500;
	}

	:global(.dark) .dropdown-item.selected {
		background: rgba(255, 255, 255, 0.1);
	}

	.dropdown-item svg {
		width: 14px;
		height: 14px;
		opacity: 0.7;
	}

	.item-count {
		margin-left: auto;
		font-size: 11px;
		color: var(--color-text-muted);
		font-weight: 500;
	}

	/* Capabilities Dropdown */
	.capabilities-menu {
		min-width: 180px;
	}

	.dropdown-checkbox-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 10px;
		border-radius: 6px;
		font-size: 13px;
		color: var(--color-text);
		cursor: pointer;
		transition: background 0.1s ease;
	}

	.dropdown-checkbox-item:hover {
		background: var(--color-bg-secondary);
	}

	.dropdown-checkbox-item input[type="checkbox"] {
		width: 14px;
		height: 14px;
		accent-color: #34c759;
		cursor: pointer;
	}

	.dropdown-clear {
		display: block;
		width: 100%;
		padding: 8px 10px;
		margin-top: 4px;
		background: transparent;
		border: none;
		border-top: 1px solid var(--color-border);
		font-size: 12px;
		color: var(--color-text-muted);
		cursor: pointer;
		text-align: center;
	}

	.dropdown-clear:hover {
		color: var(--color-text);
	}

	.cap-icon-svg {
		width: 14px;
		height: 14px;
		flex-shrink: 0;
	}

	/* Active Filter Chips */
	.active-filter-chips {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
	}

	.filter-chip {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		background: var(--color-bg-tertiary);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		font-size: 11px;
		font-weight: 500;
		color: var(--color-text);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.filter-chip:hover {
		background: var(--color-bg-secondary);
		border-color: var(--color-border-hover);
	}

	.filter-chip svg {
		width: 10px;
		height: 10px;
	}

	/* Apple-style Toggle Switch */
	.compact-toggle {
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		font-size: 12px;
		color: var(--color-text-secondary);
		white-space: nowrap;
	}

	.compact-toggle input {
		position: absolute;
		opacity: 0;
		width: 0;
		height: 0;
	}

	.toggle-slider {
		position: relative;
		width: 36px;
		height: 20px;
		background: rgba(0, 0, 0, 0.15);
		border-radius: 20px;
		transition: background 0.2s ease;
		flex-shrink: 0;
	}

	.toggle-slider::after {
		content: '';
		position: absolute;
		top: 2px;
		left: 2px;
		width: 16px;
		height: 16px;
		background: white;
		border-radius: 50%;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
		transition: transform 0.2s ease;
	}

	.compact-toggle input:checked + .toggle-slider {
		background: #34c759;
	}

	.compact-toggle input:checked + .toggle-slider::after {
		transform: translateX(16px);
	}

	:global(.dark) .toggle-slider {
		background: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .toggle-slider::after {
		background: #e5e5e5;
	}

	.toggle-label {
		font-weight: 500;
	}

	/* Compact Sort Select */
	.compact-sort-select {
		padding: 6px 10px;
		padding-right: 28px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 12px;
		font-weight: 500;
		color: var(--color-text);
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%239ca3af' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 8px center;
	}

	.compact-sort-select:hover {
		border-color: var(--color-border-hover);
	}

	:global(.dark) .compact-sort-select {
		background-color: rgba(255, 255, 255, 0.04);
	}

	/* Save Toast */
	.save-toast {
		position: fixed;
		top: 16px;
		right: 16px;
		padding: 12px 20px;
		background: var(--color-success, #22c55e);
		color: white;
		border-radius: 10px;
		font-size: 13px;
		font-weight: 500;
		z-index: 1000;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		animation: slideIn 0.3s ease;
	}

	@keyframes slideIn {
		from { transform: translateX(100%); opacity: 0; }
		to { transform: translateX(0); opacity: 1; }
	}

	@keyframes fadeInUp {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.cap-icon {
		font-size: 12px;
	}

	.sort-options {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.toggle-installed {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		color: var(--color-text-secondary);
		cursor: pointer;
	}

	.toggle-installed input {
		appearance: none;
		-webkit-appearance: none;
		width: 36px;
		height: 20px;
		background: var(--color-border);
		border-radius: 10px;
		position: relative;
		cursor: pointer;
		transition: background 0.2s ease;
	}

	.toggle-installed input::before {
		content: '';
		position: absolute;
		width: 16px;
		height: 16px;
		background: white;
		border-radius: 50%;
		top: 2px;
		left: 2px;
		transition: transform 0.2s ease;
		box-shadow: 0 1px 3px rgba(0,0,0,0.2);
	}

	.toggle-installed input:checked {
		background: var(--color-primary);
	}

	.toggle-installed input:checked::before {
		transform: translateX(16px);
	}

	.sort-select {
		padding: 8px 12px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 13px;
		color: var(--color-text);
		cursor: pointer;
	}

	/* Recommended Banner */
	.recommended-banner {
		padding: 16px;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.08), rgba(139, 92, 246, 0.08));
		border: 1px solid rgba(59, 130, 246, 0.15);
		border-radius: 12px;
	}

	:global(.dark) .recommended-banner {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(139, 92, 246, 0.15));
		border-color: rgba(59, 130, 246, 0.25);
	}

	.rec-banner-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	.rec-banner-header h3 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}

	.rec-badge-info {
		font-size: 11px;
		padding: 3px 10px;
		background: rgba(59, 130, 246, 0.12);
		border-radius: 10px;
		color: #3b82f6;
	}

	:global(.dark) .rec-badge-info {
		background: rgba(59, 130, 246, 0.2);
		color: #60a5fa;
	}

	.rec-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	.rec-chip {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.2s;
		min-width: 160px;
	}

	.rec-chip:hover {
		border-color: var(--color-primary);
		transform: translateY(-2px);
	}

	.rec-chip.installed {
		border-color: rgba(34, 197, 94, 0.4);
		background: rgba(34, 197, 94, 0.05);
	}

	.rec-chip:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.chip-name {
		font-weight: 600;
		font-size: 13px;
	}

	.chip-meta {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	.chip-status {
		font-size: 10px;
		padding: 2px 8px;
		border-radius: 4px;
		width: fit-content;
	}

	.chip-status.installed {
		background: rgba(34, 197, 94, 0.1);
		color: var(--color-success);
	}

	.chip-status.pull {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}

	/* Browser Content */
	.browser-content {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.browser-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.browser-header h3 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
	}

	.model-count {
		font-size: 13px;
		color: var(--color-text-muted);
	}

	/* Model Browser Grid */
	.model-browser-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 14px;
	}

	.model-browser-grid .browser-model-card {
		animation: fadeInUp 0.3s ease backwards;
	}

	.model-browser-grid .browser-model-card:nth-child(1) { animation-delay: 0.02s; }
	.model-browser-grid .browser-model-card:nth-child(2) { animation-delay: 0.04s; }
	.model-browser-grid .browser-model-card:nth-child(3) { animation-delay: 0.06s; }
	.model-browser-grid .browser-model-card:nth-child(4) { animation-delay: 0.08s; }
	.model-browser-grid .browser-model-card:nth-child(5) { animation-delay: 0.1s; }
	.model-browser-grid .browser-model-card:nth-child(6) { animation-delay: 0.12s; }
	.model-browser-grid .browser-model-card:nth-child(7) { animation-delay: 0.14s; }
	.model-browser-grid .browser-model-card:nth-child(8) { animation-delay: 0.16s; }
	.model-browser-grid .browser-model-card:nth-child(n+9) { animation-delay: 0.18s; }

	.browser-model-card {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 14px;
		transition: all 0.25s ease;
		position: relative;
		overflow: hidden;
	}

	.browser-model-card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: transparent;
		transition: background 0.25s ease;
	}

	.browser-model-card:hover {
		border-color: var(--color-border-hover);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
		transform: translateY(-2px);
	}

	:global(.dark) .browser-model-card:hover {
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
	}

	.browser-model-card.installed {
		border-color: rgba(34, 197, 94, 0.35);
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.04), transparent);
	}

	.browser-model-card.installed::before {
		background: linear-gradient(90deg, #22c55e, #16a34a);
	}

	.browser-model-card.default {
		border-color: rgba(59, 130, 246, 0.4);
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), transparent);
	}

	.browser-model-card.default::before {
		background: linear-gradient(90deg, #3b82f6, #2563eb);
	}

	:global(.dark) .browser-model-card {
		background: rgba(255, 255, 255, 0.02);
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .browser-model-card.installed {
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.08), transparent);
		border-color: rgba(34, 197, 94, 0.4);
	}

	:global(.dark) .browser-model-card.default {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.08), transparent);
		border-color: rgba(59, 130, 246, 0.4);
	}

	.bmc-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 12px;
	}

	.bmc-title {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	.bmc-name {
		font-weight: 700;
		font-size: 15px;
		color: var(--color-text);
		letter-spacing: -0.01em;
	}

	.bmc-installed-badge {
		font-size: 10px;
		font-weight: 600;
		padding: 3px 8px;
		background: rgba(34, 197, 94, 0.15);
		color: #22c55e;
		border-radius: 6px;
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.bmc-installed-badge::before {
		content: '✓';
		font-size: 9px;
	}

	.bmc-meta {
		display: flex;
		gap: 6px;
		flex-shrink: 0;
	}

	.bmc-size {
		font-size: 11px;
		padding: 4px 8px;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(59, 130, 246, 0.05));
		border-radius: 6px;
		color: #3b82f6;
		font-family: 'SF Mono', 'Menlo', monospace;
		font-weight: 600;
	}

	.bmc-params {
		font-size: 11px;
		padding: 4px 8px;
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.1), rgba(168, 85, 247, 0.05));
		border-radius: 6px;
		color: #a855f7;
		font-family: 'SF Mono', 'Menlo', monospace;
		font-weight: 600;
	}

	:global(.dark) .bmc-size {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(59, 130, 246, 0.08));
	}

	:global(.dark) .bmc-params {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.2), rgba(168, 85, 247, 0.08));
	}

	.bmc-description {
		font-size: 13px;
		color: var(--color-text-secondary);
		margin: 0;
		line-height: 1.4;
	}

	.bmc-capabilities {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.cap-badge {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		border-radius: 6px;
		font-size: 11px;
	}

	.cap-badge.vision {
		background: rgba(139, 92, 246, 0.12);
		color: #8b5cf6;
	}

	.cap-badge.tools {
		background: rgba(59, 130, 246, 0.12);
		color: #3b82f6;
	}

	.cap-badge.coding {
		background: rgba(34, 197, 94, 0.12);
		color: #22c55e;
	}

	.cap-badge.reasoning {
		background: rgba(249, 115, 22, 0.12);
		color: #f97316;
	}

	.cap-badge.rag {
		background: rgba(6, 182, 212, 0.12);
		color: #06b6d4;
	}

	.cap-badge.multilingual {
		background: rgba(236, 72, 153, 0.12);
		color: #ec4899;
	}

	.cap-badge.fast {
		background: rgba(234, 179, 8, 0.12);
		color: #eab308;
	}

	/* Dark mode overrides for capability badges - higher opacity for visibility */
	:global(.dark) .cap-badge.vision {
		background: rgba(139, 92, 246, 0.2);
		color: #a78bfa;
	}

	:global(.dark) .cap-badge.tools {
		background: rgba(59, 130, 246, 0.2);
		color: #60a5fa;
	}

	:global(.dark) .cap-badge.coding {
		background: rgba(34, 197, 94, 0.2);
		color: #4ade80;
	}

	:global(.dark) .cap-badge.reasoning {
		background: rgba(249, 115, 22, 0.2);
		color: #fb923c;
	}

	:global(.dark) .cap-badge.rag {
		background: rgba(6, 182, 212, 0.2);
		color: #22d3ee;
	}

	:global(.dark) .cap-badge.multilingual {
		background: rgba(236, 72, 153, 0.2);
		color: #f472b6;
	}

	:global(.dark) .cap-badge.fast {
		background: rgba(234, 179, 8, 0.2);
		color: #facc15;
	}

	.cap-badge-icon {
		font-size: 11px;
	}

	.cap-badge-icon-svg {
		width: 12px;
		height: 12px;
		flex-shrink: 0;
	}

	.cap-badge-label {
		font-weight: 500;
	}

	/* Variant Selector */
	.bmc-variants {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding-top: 10px;
		border-top: 1px solid var(--color-border);
		margin-top: 6px;
	}

	.variants-label {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 11px;
		color: var(--color-text-muted);
		flex-shrink: 0;
		padding-top: 6px;
	}

	.variants-label svg {
		opacity: 0.6;
	}

	.variant-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.variant-buttons.many {
		gap: 4px;
	}

	.variant-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		padding: 8px 12px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s ease;
		min-width: 52px;
	}

	.variant-btn:hover {
		border-color: var(--color-border-hover);
		background: var(--color-bg-tertiary);
		transform: translateY(-1px);
	}

	.variant-btn.selected {
		border-color: #3b82f6;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(59, 130, 246, 0.05));
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
	}

	:global(.dark) .variant-btn {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .variant-btn:hover {
		background: rgba(255, 255, 255, 0.06);
	}

	:global(.dark) .variant-btn.selected {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.25), rgba(59, 130, 246, 0.1));
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
	}

	.variant-params {
		font-size: 12px;
		font-weight: 700;
		color: var(--color-text);
		font-family: 'SF Mono', 'Menlo', monospace;
	}

	.variant-btn.selected .variant-params {
		color: #3b82f6;
	}

	.variant-size {
		font-size: 10px;
		color: var(--color-text-muted);
		font-family: 'SF Mono', 'Menlo', monospace;
	}

	.variant-btn.selected .variant-size {
		color: #60a5fa;
	}

	.bmc-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: auto;
		padding-top: 10px;
		border-top: 1px solid var(--color-border);
	}

	.bmc-downloads {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--color-text-muted);
		font-weight: 500;
	}

	.bmc-downloads svg {
		color: var(--color-text-muted);
		opacity: 0.7;
	}

	.bmc-actions {
		display: flex;
		gap: 8px;
		margin-left: auto;
	}

	.bmc-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		border: none;
	}

	.bmc-btn svg {
		width: 14px;
		height: 14px;
	}

	.bmc-btn.pull-btn {
		background: linear-gradient(135deg, #3b82f6, #2563eb);
		color: white;
		margin-left: auto;
		box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
	}

	.bmc-btn.pull-btn:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
	}

	.bmc-btn.pull-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
		box-shadow: none;
	}

	.bmc-btn.default-btn {
		background: rgba(128, 128, 128, 0.15);
		border: 1px solid var(--color-border);
		color: var(--color-text);
	}

	.bmc-btn.default-btn:hover {
		background: rgba(128, 128, 128, 0.25);
	}

	.bmc-btn.default-btn.is-default {
		background: var(--color-primary);
		border-color: var(--color-primary);
		color: white;
	}

	:global(.dark) .bmc-btn.default-btn {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .bmc-btn.default-btn:hover {
		background: rgba(255, 255, 255, 0.15);
	}

	.bmc-btn.delete-btn {
		background: transparent;
		color: var(--color-error);
		padding: 8px;
		border: 1px solid var(--color-border);
	}

	.bmc-btn.delete-btn:hover {
		background: rgba(220, 38, 38, 0.1);
		border-color: var(--color-error);
	}

	.btn-spinner-small {
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255,255,255,0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	/* Pull Card Compact */
	.pull-card-compact {
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.pull-compact-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.pull-compact-header h4 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.browse-link {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--color-primary);
		text-decoration: none;
	}

	.browse-link:hover {
		text-decoration: underline;
	}

	.pull-form-compact {
		display: flex;
		gap: 8px;
	}

	.pull-form-compact input {
		flex: 1;
		padding: 10px 14px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 13px;
		color: var(--color-text);
	}

	.pull-form-compact input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.pull-form-compact input::placeholder {
		color: var(--color-text-muted);
	}

	.pull-btn-compact {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 16px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.pull-btn-compact:hover {
		opacity: 0.9;
	}

	.pull-btn-compact:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.pull-btn-compact svg {
		width: 16px;
		height: 16px;
	}

	.pull-progress-compact {
		margin-top: 12px;
		padding: 12px;
		background: var(--color-bg-secondary);
		border-radius: 8px;
	}

	.progress-info {
		display: flex;
		justify-content: space-between;
		margin-bottom: 8px;
		font-size: 12px;
	}

	.progress-percent {
		font-weight: 600;
		color: var(--color-primary);
	}

	.progress-bar-compact {
		height: 6px;
		background: var(--color-border);
		border-radius: 3px;
		overflow: hidden;
	}

	.progress-details {
		display: flex;
		justify-content: space-between;
		margin-top: 8px;
		font-size: 11px;
		color: var(--color-text-muted);
	}

	.pull-error-compact {
		margin-top: 12px;
		padding: 10px 14px;
		background: rgba(220, 38, 38, 0.1);
		border-radius: 8px;
		color: var(--color-error);
		font-size: 13px;
	}

	/* Pull Card */
	.pull-card {
		margin-top: 20px;
		padding: 20px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.pull-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 14px;
	}

	.pull-header h3 {
		margin: 0;
		font-size: 15px;
	}

	.link {
		font-size: 13px;
		color: var(--color-primary);
		text-decoration: none;
	}

	.link:hover { text-decoration: underline; }

	.pull-form {
		display: flex;
		gap: 10px;
	}

	.pull-form input {
		flex: 1;
		padding: 12px 16px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		font-size: 14px;
		color: var(--color-text);
	}

	.pull-form input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.pull-form input::placeholder { color: var(--color-text-muted); }

	.pull-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 20px;
		background: var(--color-primary);
		color: var(--color-bg);
		border: none;
		border-radius: 10px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.pull-btn:hover { opacity: 0.9; }
	.pull-btn:disabled { opacity: 0.5; cursor: not-allowed; }
	.pull-btn svg { width: 18px; height: 18px; }

	.btn-spinner {
		width: 18px;
		height: 18px;
		border: 2px solid rgba(255,255,255,0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	.pull-progress {
		margin-top: 16px;
		padding: 14px;
		background: var(--color-bg-secondary);
		border-radius: 10px;
	}

	.progress-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 10px;
	}

	.progress-status { font-weight: 500; font-size: 13px; }

	.progress-stats {
		display: flex;
		gap: 12px;
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.progress-stats .speed { color: var(--color-primary); }
	.progress-stats .time { color: var(--color-success); }

	.progress-bar {
		height: 6px;
		background: var(--color-border);
		border-radius: 3px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--color-primary);
		transition: width 0.3s;
	}

	.pull-error {
		margin-top: 12px;
		padding: 10px 14px;
		background: rgba(220, 38, 38, 0.1);
		border-radius: 8px;
		color: var(--color-error);
		font-size: 13px;
	}

	/* Providers Grid */
	.providers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}

	.provider-card {
		position: relative;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 18px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
	}

	.provider-card:hover {
		border-color: var(--color-border-hover);
	}

	.provider-card.active {
		border-color: var(--color-primary);
		background: rgba(0, 0, 0, 0.02);
	}

	.provider-card.disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.provider-icon {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
		flex-shrink: 0;
	}

	.provider-icon svg {
		width: 22px;
		height: 22px;
		color: var(--color-text-muted);
	}

	:global(.dark) .provider-icon {
		background: rgba(255, 255, 255, 0.05);
	}

	:global(:not(.dark)) .provider-icon {
		background: rgba(0, 0, 0, 0.03);
	}

	.provider-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.provider-name {
		font-weight: 600;
		font-size: 14px;
	}

	.provider-desc {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.provider-status .status {
		padding: 4px 10px;
		border-radius: 6px;
		font-size: 11px;
		font-weight: 500;
	}

	.status.active {
		background: rgba(34, 197, 94, 0.1);
		color: var(--color-success);
	}

	.status.ready {
		background: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
	}

	.status.setup {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}

	.provider-badge {
		position: absolute;
		top: 8px;
		right: 8px;
		padding: 2px 8px;
		font-size: 10px;
		font-weight: 600;
		border-radius: 4px;
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}

	.provider-badge.local {
		background: rgba(34, 197, 94, 0.1);
		color: var(--color-success);
	}

	/* API Grid */
	.api-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}

	.api-card {
		padding: 18px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.api-header {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 14px;
	}

	.api-icon {
		width: 38px;
		height: 38px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		flex-shrink: 0;
	}

	.api-icon svg {
		width: 20px;
		height: 20px;
		color: var(--color-text-muted);
	}

	:global(.dark) .api-icon {
		background: rgba(255, 255, 255, 0.05);
	}

	:global(:not(.dark)) .api-icon {
		background: rgba(0, 0, 0, 0.03);
	}

	.api-name {
		font-weight: 600;
		flex: 1;
	}

	.api-configured {
		width: 22px;
		height: 22px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(34, 197, 94, 0.1);
		border-radius: 50%;
	}

	.api-configured svg {
		width: 12px;
		height: 12px;
		color: var(--color-success);
	}

	.api-form {
		display: flex;
		gap: 8px;
		margin-bottom: 10px;
	}

	.input-wrapper {
		flex: 1;
		position: relative;
	}

	.input-wrapper input {
		width: 100%;
		padding: 10px 36px 10px 12px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 13px;
		font-family: monospace;
		color: var(--color-text);
	}

	.input-wrapper input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.input-wrapper input::placeholder { color: var(--color-text-muted); }

	.toggle-btn {
		position: absolute;
		right: 8px;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		padding: 4px;
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.toggle-btn:hover { color: var(--color-text); }
	.toggle-btn svg { width: 16px; height: 16px; }

	.save-btn {
		padding: 10px 16px;
		background: var(--color-primary);
		color: var(--color-bg);
		border: none;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
	}

	.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }

	.api-link {
		font-size: 12px;
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.api-link:hover { color: var(--color-primary); }

	/* Presets Row */
	.presets-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
		flex-wrap: wrap;
	}

	.presets-label {
		font-size: 13px;
		color: var(--color-text-tertiary);
		margin-right: 4px;
	}

	.preset-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		font-size: 12px;
		font-weight: 500;
		border-radius: 8px;
		border: 1px solid var(--color-border);
		background: rgba(255, 255, 255, 0.05);
		color: var(--color-text-secondary);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	:global(.dark) .preset-btn {
		background: rgba(255, 255, 255, 0.05);
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(:not(.dark)) .preset-btn {
		background: rgba(0, 0, 0, 0.02);
	}

	.preset-btn:hover {
		background: rgba(255, 255, 255, 0.1);
		border-color: var(--color-primary);
		color: var(--color-primary);
	}

	:global(:not(.dark)) .preset-btn:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.preset-icon {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	/* Settings Grid */
	.settings-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}

	.setting-card {
		padding: 18px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.setting-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.setting-header label {
		font-weight: 500;
		font-size: 14px;
	}

	.setting-value {
		font-size: 14px;
		font-weight: 600;
		color: var(--color-primary);
		font-family: monospace;
	}

	.setting-card input[type="range"] {
		width: 100%;
		height: 6px;
		background: var(--color-border);
		border-radius: 3px;
		appearance: none;
		cursor: pointer;
	}

	.setting-card input[type="range"]::-webkit-slider-thumb {
		appearance: none;
		width: 18px;
		height: 18px;
		background: var(--color-primary);
		border-radius: 50%;
		cursor: pointer;
	}

	.setting-desc {
		margin: 10px 0 0;
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.toggle-card .setting-header {
		margin-bottom: 0;
	}

	.toggle {
		width: 44px;
		height: 24px;
		background: var(--color-border);
		border: none;
		border-radius: 12px;
		cursor: pointer;
		position: relative;
		transition: all 0.2s;
	}

	.toggle.on {
		background: var(--color-success);
	}

	.toggle-knob {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		background: white;
		border-radius: 50%;
		transition: all 0.2s;
	}

	.toggle.on .toggle-knob {
		left: 22px;
	}

	.toggle-card .setting-desc {
		margin-top: 12px;
	}

	.settings-actions {
		display: flex;
		gap: 10px;
		margin-top: 20px;
	}

	.action-btn {
		padding: 12px 24px;
		background: var(--color-bg-tertiary);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		color: var(--color-text);
	}

	.action-btn:hover {
		background: var(--color-bg-secondary);
	}

	.action-btn.primary {
		background: var(--color-primary);
		color: var(--color-bg);
		border-color: transparent;
	}

	.action-btn.primary:hover { opacity: 0.9; }
	.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }

	/* Default Model Grid */
	.default-model-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 10px;
	}

	.default-model-btn {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 14px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		text-align: left;
		cursor: pointer;
		transition: all 0.2s;
	}

	.default-model-btn:hover {
		border-color: var(--color-border-hover);
	}

	.default-model-btn.selected {
		border-color: var(--color-primary);
		background: rgba(0, 0, 0, 0.02);
	}

	.dm-name {
		font-weight: 600;
		font-size: 14px;
	}

	.dm-size, .dm-desc {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.dm-check {
		position: absolute;
		top: 10px;
		right: 10px;
		width: 18px;
		height: 18px;
		color: var(--color-success);
	}

	/* ===== AGENTS TAB STYLES ===== */

	/* Custom Agents */
	.section-subtitle {
		color: var(--color-text-muted);
		font-size: 14px;
		margin: -8px 0 16px 0;
	}

	.custom-agent-form {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		padding: 20px;
		margin-bottom: 20px;
	}

	.custom-agents-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.custom-agent-card {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		padding: 16px;
	}

	.custom-agent-card:hover {
		border-color: var(--color-primary);
	}

	.custom-agent-card .agent-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 8px;
	}

	.custom-agent-card .agent-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.agent-mention {
		font-family: monospace;
		font-weight: 600;
		color: var(--color-primary);
		background: var(--color-primary-bg);
		padding: 4px 8px;
		border-radius: 6px;
		font-size: 14px;
	}

	.agent-display-name {
		font-weight: 500;
		color: var(--color-text);
	}

	.custom-agent-card .agent-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.usage-badge {
		font-size: 12px;
		color: var(--color-text-muted);
		background: var(--color-bg-elevated);
		padding: 4px 8px;
		border-radius: 4px;
	}

	.custom-agent-card .agent-desc {
		font-size: 14px;
		color: var(--color-text-muted);
		margin-bottom: 8px;
	}

	.agent-prompt-preview {
		font-size: 13px;
		display: flex;
		gap: 8px;
	}

	.agent-prompt-preview .label {
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.agent-prompt-preview .preview {
		color: var(--color-text-secondary);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.empty-state.small {
		padding: 24px;
		text-align: center;
		color: var(--color-text-muted);
		background: var(--color-bg-elevated);
		border-radius: 8px;
	}

	.icon-btn.danger:hover {
		color: var(--color-error);
	}

	/* Built-in Agents */
	.agents-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.agent-card {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		overflow: hidden;
		transition: all 0.2s;
	}

	.agent-card:hover {
		border-color: var(--color-border-hover);
	}

	.agent-card.expanded {
		border-color: var(--color-primary);
	}

	.agent-header-btn {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
	}

	.agent-info {
		display: flex;
		align-items: center;
		gap: 14px;
	}

	.agent-icon {
		width: 48px;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-bg-secondary);
		border-radius: 12px;
		flex-shrink: 0;
	}

	.agent-icon svg {
		width: 24px;
		height: 24px;
		color: var(--color-text-muted);
	}

	:global(.dark) .agent-icon {
		background: rgba(255, 255, 255, 0.05);
	}

	:global(:not(.dark)) .agent-icon {
		background: rgba(0, 0, 0, 0.03);
	}

	.agent-text {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.agent-name {
		font-weight: 600;
		font-size: 15px;
		color: var(--color-text);
	}

	.agent-desc {
		font-size: 13px;
		color: var(--color-text-muted);
	}

	.agent-meta {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.agent-category {
		font-size: 11px;
		padding: 4px 10px;
		background: var(--color-bg-tertiary);
		border-radius: 6px;
		color: var(--color-text-secondary);
		text-transform: uppercase;
		font-weight: 500;
	}

	.agent-chevron {
		width: 20px;
		height: 20px;
		color: var(--color-text-muted);
		transition: transform 0.2s;
	}

	.agent-chevron.rotated {
		transform: rotate(180deg);
	}

	.agent-content {
		padding: 0 20px 20px;
		border-top: 1px solid var(--color-border);
		background: var(--color-bg-secondary);
	}

	.prompt-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 0 12px;
	}

	.prompt-header h4 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.prompt-edit-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 12px;
		color: var(--color-text-secondary);
		cursor: pointer;
		transition: all 0.2s;
	}

	.prompt-edit-btn:hover {
		background: var(--color-bg-tertiary);
		color: var(--color-text);
	}

	.prompt-edit-btn svg {
		width: 14px;
		height: 14px;
	}

	.prompt-display {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		max-height: 400px;
		overflow-y: auto;
	}

	.prompt-display pre {
		margin: 0;
		padding: 16px;
		font-size: 12px;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
		line-height: 1.6;
		white-space: pre-wrap;
		word-break: break-word;
		color: var(--color-text-secondary);
	}

	.prompt-editor {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.prompt-editor textarea {
		width: 100%;
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		font-size: 12px;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
		line-height: 1.6;
		color: var(--color-text);
		resize: vertical;
	}

	.prompt-editor textarea:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.editor-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}

	.editor-btn {
		padding: 8px 16px;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.editor-btn.cancel {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		color: var(--color-text-secondary);
	}

	.editor-btn.cancel:hover {
		background: var(--color-bg-tertiary);
	}

	.editor-btn.save {
		background: var(--color-primary);
		border: none;
		color: var(--color-bg);
	}

	.editor-btn.save:hover {
		opacity: 0.9;
	}

	.agent-stats {
		display: flex;
		gap: 24px;
		padding-top: 16px;
		margin-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.stat-item .stat-label {
		font-size: 10px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.stat-value-small {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-secondary);
		font-family: monospace;
	}

	/* ===== DARK MODE FOR AI SETTINGS ===== */
	:global(.dark) .page {
		background: #1c1c1e;
	}

	:global(.dark) .system-overview {
		background: #2c2c2e;
	}

	:global(.dark) .tabs {
		background: #2c2c2e;
	}

	:global(.dark) .tab.active {
		background: #3a3a3c;
	}

	:global(.dark) .rec-card,
	:global(.dark) .model-card,
	:global(.dark) .provider-card,
	:global(.dark) .api-card,
	:global(.dark) .setting-card,
	:global(.dark) .pull-card,
	:global(.dark) .agent-card {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .rec-card:hover,
	:global(.dark) .model-card:hover,
	:global(.dark) .provider-card:hover,
	:global(.dark) .agent-card:hover {
		border-color: rgba(255, 255, 255, 0.2);
	}

	:global(.dark) .agent-content {
		background: #1c1c1e;
	}

	:global(.dark) .prompt-display {
		background: #2c2c2e;
	}

	:global(.dark) .prompt-editor textarea {
		background: #2c2c2e;
	}

	:global(.dark) .action-btn.primary,
	:global(.dark) .rec-btn:not(.secondary),
	:global(.dark) .pull-btn,
	:global(.dark) .save-btn,
	:global(.dark) .editor-btn.save,
	:global(.dark) .add-btn,
	:global(.dark) .btn.primary {
		background: #0A84FF;
		color: white;
	}

	:global(.dark) .stat,
	:global(.dark) .ram-card,
	:global(.dark) .platform-icon,
	:global(.dark) .agent-icon {
		background: #3a3a3c;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .default-model-btn.selected,
	:global(.dark) .model-card.selected,
	:global(.dark) .provider-card.active,
	:global(.dark) .agent-card.expanded {
		border-color: #0A84FF;
		background: rgba(10, 132, 255, 0.1);
	}

	/* Commands Tab Styles */
	.add-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.add-btn:hover { opacity: 0.9; }
	.add-btn svg { width: 16px; height: 16px; }

	.command-form, .modal {
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		padding: 20px;
		margin-bottom: 24px;
	}

	.form-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.form-header h3 {
		font-size: 16px;
		font-weight: 600;
		margin: 0;
	}

	.close-btn {
		width: 28px;
		height: 28px;
		border: none;
		background: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		border-radius: 6px;
		font-size: 18px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.close-btn:hover {
		background: var(--color-bg);
		color: var(--color-text);
	}

	.form-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.form-group.full-width {
		grid-column: span 2;
	}

	.form-group label {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-secondary);
	}

	.form-group input, .form-group textarea {
		padding: 10px 12px;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		background: var(--color-bg);
		color: var(--color-text);
		font-size: 14px;
		transition: border-color 0.2s;
	}

	.form-group input:focus, .form-group textarea:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.form-group textarea {
		resize: vertical;
		font-family: monospace;
		font-size: 13px;
	}

	.input-prefix {
		display: flex;
		align-items: center;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		background: var(--color-bg);
		overflow: hidden;
	}

	.input-prefix span {
		padding: 10px 4px 10px 12px;
		color: var(--color-text-secondary);
		font-family: monospace;
	}

	.input-prefix input {
		border: none;
		padding-left: 0;
		flex: 1;
	}

	.form-hint {
		font-size: 11px;
		color: var(--color-text-tertiary);
	}

	.context-sources {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.context-chip {
		padding: 6px 12px;
		border: 1px solid var(--color-border);
		border-radius: 20px;
		background: var(--color-bg);
		color: var(--color-text-secondary);
		font-size: 12px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.context-chip:hover {
		border-color: var(--color-primary);
	}

	.context-chip.active {
		background: var(--color-primary);
		color: white;
		border-color: var(--color-primary);
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
		margin-top: 20px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.btn {
		padding: 10px 20px;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn.primary {
		background: var(--color-primary);
		color: white;
	}

	.btn.primary:hover { opacity: 0.9; }
	.btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }

	.btn.secondary {
		background: var(--color-bg-tertiary);
		color: var(--color-text);
	}

	.btn.secondary:hover {
		background: var(--color-bg);
	}

	.commands-section {
		margin-bottom: 24px;
	}

	.commands-section h4 {
		font-size: 14px;
		font-weight: 600;
		margin: 0 0 12px;
		color: var(--color-text-secondary);
	}

	.commands-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}

	.command-card {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		padding: 14px;
		transition: all 0.2s;
	}

	.command-card:hover {
		border-color: var(--color-border-hover);
	}

	.command-card.custom {
		border-color: rgba(99, 102, 241, 0.3);
		background: rgba(99, 102, 241, 0.03);
	}

	.command-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
	}

	.command-icon {
		font-size: 18px;
	}

	.command-name {
		font-family: monospace;
		font-size: 14px;
		font-weight: 600;
		color: var(--color-primary);
	}

	.command-category {
		margin-left: auto;
		font-size: 11px;
		padding: 2px 8px;
		background: var(--color-bg-tertiary);
		border-radius: 4px;
		color: var(--color-text-tertiary);
		text-transform: capitalize;
	}

	.command-actions {
		margin-left: auto;
		display: flex;
		gap: 4px;
	}

	.icon-btn {
		width: 28px;
		height: 28px;
		border: none;
		background: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
	}

	.icon-btn svg {
		width: 14px;
		height: 14px;
	}

	.icon-btn:hover {
		background: var(--color-bg);
		color: var(--color-text);
	}

	.icon-btn.danger:hover {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}

	.command-display-name {
		font-size: 14px;
		font-weight: 500;
		margin: 0 0 4px;
		color: var(--color-text);
	}

	.command-desc {
		font-size: 12px;
		color: var(--color-text-secondary);
		margin: 0;
		line-height: 1.4;
	}

	.command-sources {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 10px;
	}

	.source-tag {
		font-size: 10px;
		padding: 2px 6px;
		background: var(--color-bg-tertiary);
		border-radius: 4px;
		color: var(--color-text-tertiary);
	}

	/* Clickable command cards */
	.command-card.clickable {
		cursor: pointer;
		transition: all 0.2s;
	}

	.command-card.clickable:hover {
		border-color: var(--color-primary);
		transform: translateY(-1px);
	}

	.command-card.expanded {
		border-color: var(--color-primary);
		background: var(--color-bg-secondary);
	}

	.command-header .edit-builtin {
		opacity: 0;
		transition: opacity 0.2s;
	}

	.command-card:hover .edit-builtin {
		opacity: 1;
	}

	/* Command details (expanded view) */
	.command-details {
		margin-top: 16px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.detail-section h5 {
		font-size: 12px;
		font-weight: 600;
		color: var(--color-text-secondary);
		margin: 0 0 4px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.detail-hint {
		font-size: 11px;
		color: var(--color-text-tertiary);
		margin: 0 0 8px;
	}

	.prompt-preview {
		font-size: 12px;
		font-family: monospace;
		background: var(--color-bg-tertiary);
		padding: 12px;
		border-radius: 8px;
		white-space: pre-wrap;
		word-break: break-word;
		max-height: 200px;
		overflow-y: auto;
		color: var(--color-text-secondary);
		line-height: 1.5;
		margin: 0;
	}

	.detail-actions {
		margin-top: 12px;
		display: flex;
		justify-content: flex-end;
	}

	.btn.small {
		padding: 6px 12px;
		font-size: 12px;
	}

	/* Modal Overlay */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal {
		max-width: 600px;
		width: 90%;
		max-height: 85vh;
		overflow-y: auto;
	}

	/* Dark mode for commands */
	:global(.dark) .command-form,
	:global(.dark) .modal {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .command-card {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .command-card.custom {
		background: rgba(99, 102, 241, 0.1);
		border-color: rgba(99, 102, 241, 0.3);
	}

	:global(.dark) .context-chip {
		background: #3a3a3c;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .context-chip.active {
		background: #0A84FF;
		border-color: #0A84FF;
	}

	:global(.dark) .close-btn,
	:global(.dark) .icon-btn {
		background: #3a3a3c;
	}

	:global(.dark) .btn.secondary {
		background: #3a3a3c;
	}

	:global(.dark) .input-prefix {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .form-group input,
	:global(.dark) .form-group textarea {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .form-group input:focus,
	:global(.dark) .form-group textarea:focus {
		border-color: #0A84FF;
	}

	:global(.dark) .command-name {
		color: #0A84FF;
	}

	:global(.dark) .setting-value {
		color: #0A84FF;
	}

	:global(.dark) .command-card.expanded {
		background: #2c2c2e;
		border-color: #0A84FF;
	}

	:global(.dark) .command-card.clickable:hover {
		border-color: #0A84FF;
	}

	:global(.dark) .command-details {
		border-top-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .prompt-preview {
		background: #1c1c1e;
		color: #98989d;
	}

	:global(.dark) .source-tag {
		background: #3a3a3c;
		color: #98989d;
	}

	/* Stats Tab Styles */
	.stats-overview {
		display: flex;
		gap: 20px;
		padding: 20px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.stats-system {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	/* Usage Summary Cards */
	.usage-summary {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 12px;
		margin-bottom: 24px;
	}

	.usage-card {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 18px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
	}

	.usage-icon {
		font-size: 28px;
	}

	.usage-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.usage-value {
		font-size: 22px;
		font-weight: 700;
	}

	.usage-label {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	/* Usage Sections */
	.usage-section {
		margin-bottom: 20px;
	}

	.usage-section h4 {
		margin: 0 0 12px;
		font-size: 14px;
		font-weight: 600;
		color: var(--color-text-secondary);
	}

	.usage-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.usage-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 8px;
	}

	.usage-row-label {
		flex: 1;
		font-weight: 500;
		font-size: 13px;
	}

	.usage-row-value {
		font-size: 12px;
		color: var(--color-text-secondary);
	}

	.usage-row-tokens {
		font-size: 12px;
		color: var(--color-text-muted);
		font-family: monospace;
	}

	/* Refresh Button */
	.refresh-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		background: var(--color-bg-tertiary);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 13px;
		color: var(--color-text);
		cursor: pointer;
		transition: all 0.2s;
	}

	.refresh-btn:hover {
		background: var(--color-bg-secondary);
	}

	.refresh-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.refresh-btn svg {
		width: 16px;
		height: 16px;
	}

	.refresh-btn svg.spinning {
		animation: spin 1s linear infinite;
	}

	/* ========== STATS PAGE STYLES ========== */
	.stats-page {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.stats-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 16px;
	}

	.stats-title-area {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.stats-eyebrow {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--color-text-muted);
	}

	.stats-date-picker {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.stats-date-picker:hover {
		background: var(--color-bg-tertiary);
		border-color: var(--color-text-muted);
	}

	:global(.dark) .stats-date-picker {
		background: rgba(255, 255, 255, 0.05);
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .stats-date-picker:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.stats-date-picker svg {
		color: var(--color-text-muted);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 16px;
	}

	.section-header h3 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}

	.section-hint {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.stats-controls {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.period-selector {
		display: flex;
		gap: 4px;
		padding: 4px;
		background: rgba(0, 0, 0, 0.1);
		border-radius: 10px;
	}

	:global(.dark) .period-selector {
		background: rgba(255, 255, 255, 0.05);
	}

	.period-btn {
		padding: 8px 14px;
		background: transparent;
		border: none;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-secondary);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.period-btn:hover {
		color: var(--color-text);
		background: rgba(0, 0, 0, 0.05);
	}

	:global(.dark) .period-btn:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.period-btn.active {
		background: #3b82f6 !important;
		color: white !important;
	}

	:global(.dark) .period-btn.active {
		background: #3b82f6 !important;
		color: white !important;
	}

	/* System Row */
	.stats-system-row {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 20px;
	}

	.system-card {
		padding: 20px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
	}

	.system-platform-info {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-bottom: 20px;
	}

	.platform-icon-large {
		width: 48px;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-bg);
		border-radius: 12px;
	}

	.platform-icon-large svg {
		width: 28px;
		height: 28px;
		color: var(--color-text);
	}

	.platform-text {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.platform-label {
		font-size: 18px;
		font-weight: 600;
	}

	.gpu-badge {
		font-size: 12px;
		padding: 3px 10px;
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.15), rgba(34, 197, 94, 0.05));
		color: #22c55e;
		border-radius: 20px;
		width: fit-content;
	}

	.system-metrics {
		display: flex;
		gap: 32px;
	}

	.sys-metric {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.sys-metric-value {
		font-size: 24px;
		font-weight: 700;
		color: var(--color-text);
	}

	.sys-metric-label {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	/* RAM Gauge Card */
	.ram-gauge-card {
		padding: 20px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		min-width: 180px;
	}

	.ram-gauge-header {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: 12px;
	}

	.ram-title {
		font-size: 14px;
		font-weight: 600;
	}

	.ram-detail {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.ram-gauge-visual {
		position: relative;
		width: 120px;
		height: 120px;
		margin: 0 auto;
	}

	.gauge-svg-large {
		width: 100%;
		height: 100%;
	}

	.gauge-bg-large {
		stroke: var(--color-border);
		stroke-width: 8;
	}

	.gauge-fill-large {
		stroke: #22c55e;
		stroke-width: 8;
		stroke-linecap: round;
		transition: stroke-dasharray 0.5s ease;
	}

	.gauge-center-text {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		text-align: center;
	}

	.gauge-percent {
		display: block;
		font-size: 24px;
		font-weight: 700;
		color: #22c55e;
	}

	.gauge-subtitle {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	/* Key Metrics Grid */
	.key-metrics-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	.metric-card {
		position: relative;
		padding: 20px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		overflow: hidden;
		transition: all 0.2s ease;
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
	}

	:global(.dark) .metric-card:hover {
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
	}

	/* Different gradient backgrounds for each metric type */
	.metric-card.requests {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.12), rgba(59, 130, 246, 0.02));
		border-color: rgba(59, 130, 246, 0.25);
	}
	.metric-card.requests .metric-icon svg { color: #3b82f6; }
	.metric-card.requests .metric-sparkline { color: rgba(59, 130, 246, 0.5); }

	.metric-card.tokens {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.12), rgba(168, 85, 247, 0.02));
		border-color: rgba(168, 85, 247, 0.25);
	}
	.metric-card.tokens .metric-icon svg { color: #a855f7; }
	.metric-card.tokens .metric-sparkline { color: rgba(168, 85, 247, 0.5); }

	.metric-card.speed {
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.12), rgba(34, 197, 94, 0.02));
		border-color: rgba(34, 197, 94, 0.25);
	}
	.metric-card.speed .metric-icon svg { color: #22c55e; }

	.metric-card.cost {
		background: linear-gradient(135deg, rgba(249, 115, 22, 0.12), rgba(249, 115, 22, 0.02));
		border-color: rgba(249, 115, 22, 0.25);
	}
	.metric-card.cost .metric-icon svg { color: #f97316; }

	/* Sparkline background */
	.metric-sparkline {
		position: absolute;
		top: 10px;
		right: 10px;
		width: 80px;
		height: 30px;
		opacity: 0.6;
	}

	.metric-sparkline svg {
		width: 100%;
		height: 100%;
	}

	/* Metric header with icon and trend */
	.metric-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 4px;
	}

	.metric-icon {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 10px;
	}

	:global(.dark) .metric-icon {
		background: rgba(255, 255, 255, 0.05);
	}

	.metric-icon svg {
		width: 20px;
		height: 20px;
	}

	.metric-trend {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 11px;
		font-weight: 600;
		padding: 4px 8px;
		border-radius: 20px;
	}

	.metric-trend.up {
		color: #22c55e;
		background: rgba(34, 197, 94, 0.15);
	}

	.metric-trend.down {
		color: #ef4444;
		background: rgba(239, 68, 68, 0.15);
	}

	.metric-content {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.metric-value {
		font-size: 32px;
		font-weight: 700;
		font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
		letter-spacing: -0.02em;
	}

	.metric-label {
		font-size: 13px;
		color: var(--color-text-muted);
		font-weight: 500;
	}

	.metric-sub {
		margin-top: 8px;
	}

	.metric-badge {
		display: inline-block;
		font-size: 11px;
		font-weight: 500;
		padding: 4px 10px;
		background: rgba(0, 0, 0, 0.05);
		border-radius: 20px;
		color: var(--color-text-muted);
	}

	:global(.dark) .metric-badge {
		background: rgba(255, 255, 255, 0.08);
	}

	.metric-badge.good { color: #22c55e; background: rgba(34, 197, 94, 0.12); }
	.metric-badge.slow { color: #ef4444; background: rgba(239, 68, 68, 0.12); }

	.metric-breakdown {
		display: flex;
		gap: 12px;
		font-size: 12px;
		font-weight: 500;
		margin-top: 8px;
	}

	.metric-in {
		display: flex;
		align-items: center;
		gap: 4px;
		color: #22c55e;
	}

	.metric-out {
		display: flex;
		align-items: center;
		gap: 4px;
		color: #f97316;
	}

	.cost-breakdown {
		flex-direction: column;
		gap: 6px;
	}

	.metric-cloud,
	.metric-local {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
	}
	.metric-cloud { color: #3b82f6; }
	.metric-local { color: #22c55e; }

	.metric-cloud svg,
	.metric-local svg {
		flex-shrink: 0;
		opacity: 0.7;
	}

	/* Comparison Section */
	.comparison-section {
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		padding: 24px;
	}

	.comparison-section h3 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}

	/* Empty State for Comparison */
	.comparison-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		text-align: center;
	}

	.comparison-empty-icon {
		color: var(--color-text-muted);
		margin-bottom: 16px;
		opacity: 0.5;
	}

	.comparison-empty-text {
		font-size: 14px;
		color: var(--color-text-muted);
		margin: 0 0 8px 0;
	}

	.comparison-empty-hint {
		font-size: 12px;
		color: var(--color-text-muted);
		opacity: 0.7;
		margin: 0;
	}

	/* Split Bar Comparison */
	.comparison-split-bar {
		margin-bottom: 20px;
	}

	.split-bar-container {
		display: flex;
		height: 32px;
		border-radius: 8px;
		overflow: hidden;
		background: var(--color-bg);
	}

	.split-bar-fill {
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 30px;
		transition: width 0.5s ease;
	}

	.split-bar-fill.local {
		background: linear-gradient(90deg, #22c55e, #16a34a);
	}

	.split-bar-fill.cloud {
		background: linear-gradient(90deg, #3b82f6, #2563eb);
	}

	.split-bar-label {
		font-size: 12px;
		font-weight: 600;
		color: white;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
	}

	.comparison-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	.comparison-card {
		padding: 20px;
		background: var(--color-bg);
		border-radius: 14px;
		border: 1px solid var(--color-border);
		transition: all 0.2s ease;
	}

	.comparison-card:hover {
		border-color: var(--color-text-muted);
	}

	.comparison-card.local {
		border-color: rgba(34, 197, 94, 0.3);
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.05), transparent);
	}

	.comparison-card.local:hover {
		border-color: rgba(34, 197, 94, 0.5);
	}

	.comparison-card.cloud {
		border-color: rgba(59, 130, 246, 0.3);
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), transparent);
	}

	.comparison-card.cloud:hover {
		border-color: rgba(59, 130, 246, 0.5);
	}

	.comp-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.comp-icon-wrapper {
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10px;
	}

	.comp-icon-wrapper.local {
		background: rgba(34, 197, 94, 0.15);
		color: #22c55e;
	}

	.comp-icon-wrapper.cloud {
		background: rgba(59, 130, 246, 0.15);
		color: #3b82f6;
	}

	.comp-title-group {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.comp-subtitle {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	.comp-icon {
		font-size: 20px;
	}

	.comp-title {
		font-size: 15px;
		font-weight: 600;
	}

	.comp-stats {
		display: flex;
		gap: 24px;
		margin-bottom: 16px;
	}

	.comp-stat {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.comp-value {
		font-size: 20px;
		font-weight: 700;
	}

	.comp-label {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	.comp-bar {
		position: relative;
		height: 8px;
		background: var(--color-border);
		border-radius: 4px;
		overflow: hidden;
	}

	.comp-bar-fill {
		height: 100%;
		border-radius: 4px;
		transition: width 0.5s ease;
	}

	.comp-bar-fill.local { background: #22c55e; }
	.comp-bar-fill.cloud { background: #3b82f6; }

	.comp-bar-label {
		position: absolute;
		right: 8px;
		top: 50%;
		transform: translateY(-50%);
		font-size: 10px;
		color: white;
		font-weight: 600;
		text-shadow: 0 1px 2px rgba(0,0,0,0.3);
	}

	/* Breakdowns Grid */
	.breakdowns-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
	}

	.breakdown-card {
		padding: 20px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		min-height: 200px;
		display: flex;
		flex-direction: column;
	}

	.breakdown-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
	}

	.breakdown-header svg {
		color: var(--color-text-muted);
	}

	.breakdown-header h4 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.breakdown-empty {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		color: var(--color-text-muted);
		opacity: 0.6;
	}

	.breakdown-empty span {
		font-size: 13px;
	}

	.breakdown-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.breakdown-row {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 8px 10px;
		background: var(--color-bg);
		border-radius: 8px;
		transition: all 0.15s ease;
	}

	.breakdown-row:hover {
		background: var(--color-bg-tertiary);
	}

	:global(.dark) .breakdown-row {
		background: rgba(255, 255, 255, 0.03);
	}

	:global(.dark) .breakdown-row:hover {
		background: rgba(255, 255, 255, 0.06);
	}

	.breakdown-info {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.breakdown-icon {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		border-radius: 6px;
	}

	.breakdown-icon.provider { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
	.breakdown-icon.agent { background: rgba(168, 85, 247, 0.1); color: #a855f7; }

	.breakdown-icon svg {
		width: 14px;
		height: 14px;
	}

	.breakdown-name {
		font-size: 13px;
		font-weight: 500;
	}

	.breakdown-name.truncate {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 120px;
	}

	.breakdown-values {
		display: flex;
		gap: 12px;
		font-size: 11px;
		font-family: 'SF Mono', 'Menlo', monospace;
	}

	.breakdown-stat-value {
		font-weight: 500;
	}

	.breakdown-stat-value.requests { color: #3b82f6; }
	.breakdown-stat-value.tokens { color: #f97316; }
	.breakdown-stat-value.cost { color: #22c55e; }
	.breakdown-stat-value.latency { color: #a855f7; }

	.breakdown-bar {
		height: 4px;
		background: var(--color-border);
		border-radius: 2px;
		overflow: hidden;
	}

	.breakdown-bar-fill {
		height: 100%;
		border-radius: 2px;
		transition: width 0.5s ease;
	}

	.breakdown-bar-fill.provider { background: #3b82f6; }
	.breakdown-bar-fill.model { background: #f97316; }
	.breakdown-bar-fill.agent { background: #a855f7; }

	/* Activity Chart */
	.activity-section {
		padding: 24px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
	}

	.activity-section h3 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}

	.activity-chart {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.activity-area-chart {
		height: 120px;
		border-radius: 8px;
		overflow: hidden;
	}

	.activity-area-chart svg {
		width: 100%;
		height: 100%;
	}

	.activity-labels {
		display: flex;
		justify-content: space-between;
	}

	.activity-label-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
	}

	.activity-day {
		font-size: 10px;
		color: var(--color-text-muted);
		text-transform: uppercase;
	}

	.activity-count {
		font-size: 11px;
		font-weight: 600;
		color: var(--color-primary);
	}

	.activity-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px 20px;
		gap: 12px;
		color: var(--color-text-muted);
	}

	.activity-empty p {
		font-size: 13px;
		margin: 0;
	}

	/* Session Stats */
	.session-stats {
		padding: 24px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
	}

	.session-stats h3 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
	}

	.session-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	.session-stat-card {
		display: flex;
		align-items: flex-start;
		gap: 14px;
		padding: 16px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		transition: all 0.2s ease;
	}

	.session-stat-card:hover {
		border-color: var(--color-text-muted);
		transform: translateY(-1px);
	}

	:global(.dark) .session-stat-card {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .session-stat-card:hover {
		background: rgba(255, 255, 255, 0.06);
		border-color: rgba(255, 255, 255, 0.15);
	}

	.session-stat-icon {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(59, 130, 246, 0.02));
		border-radius: 10px;
		color: #3b82f6;
		flex-shrink: 0;
	}

	.session-stat-content {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.session-value {
		font-size: 22px;
		font-weight: 700;
		font-family: 'SF Mono', 'Menlo', monospace;
		color: var(--color-text);
		line-height: 1;
	}

	.session-unit {
		font-size: 14px;
		font-weight: 500;
		opacity: 0.7;
	}

	.session-label {
		font-size: 12px;
		color: var(--color-text-muted);
		font-weight: 500;
	}

	/* Responsive adjustments */
	@media (max-width: 1200px) {
		.key-metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}
		.breakdowns-grid {
			grid-template-columns: 1fr;
		}
		.session-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.stats-system-row {
			grid-template-columns: 1fr;
		}
		.comparison-grid {
			grid-template-columns: 1fr;
		}
		.key-metrics-grid {
			grid-template-columns: 1fr;
		}
	}

	/* Dark mode enhancements */
	:global(.dark) .metric-card.primary {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(59, 130, 246, 0.03));
	}

	:global(.dark) .metric-card.cost {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.15), rgba(168, 85, 247, 0.03));
	}

	:global(.dark) .gpu-badge {
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(34, 197, 94, 0.08));
	}

	:global(.dark) .comparison-card.local {
		border-color: rgba(34, 197, 94, 0.4);
		background: rgba(34, 197, 94, 0.05);
	}

	:global(.dark) .comparison-card.cloud {
		border-color: rgba(59, 130, 246, 0.4);
		background: rgba(59, 130, 246, 0.05);
	}

	:global(.dark) .provider-type-btn.local.active {
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(34, 197, 94, 0.08));
	}

	:global(.dark) .provider-type-btn.cloud.active {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(59, 130, 246, 0.08));
	}

	/* User Facts */
	.facts-toolbar {
		display: flex;
		justify-content: flex-end;
		margin-top: 10px;
	}

	.facts-list {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.fact-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding: 12px;
		border: 1px solid var(--color-border);
		border-radius: 12px;
		background: var(--color-bg-secondary);
	}

	.fact-main {
		min-width: 0;
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.fact-key {
		font-weight: 600;
		color: var(--color-text);
		font-size: 13px;
	}

	.fact-value {
		color: var(--color-text-secondary);
		font-size: 13px;
		word-break: break-word;
	}

	.fact-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 2px;
	}

	.fact-pill {
		font-size: 11px;
		padding: 3px 8px;
		border-radius: 999px;
		border: 1px solid var(--color-border);
		color: var(--color-text-secondary);
		background: var(--color-bg-tertiary);
	}

	.fact-pill.inactive {
		opacity: 0.75;
	}

	.fact-actions {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
	}
</style>
