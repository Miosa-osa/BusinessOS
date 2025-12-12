// Use relative URL - Vite proxy forwards /api to FastAPI backend
const API_BASE = '/api';

interface RequestOptions {
	method?: string;
	body?: unknown;
	headers?: Record<string, string>;
}

class ApiClient {
	private async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
		const { method = 'GET', body, headers = {} } = options;

		if (body && !headers['Content-Type']) {
			headers['Content-Type'] = 'application/json';
		}

		const response = await fetch(`${API_BASE}${endpoint}`, {
			method,
			headers,
			credentials: 'include', // Send Better Auth cookies
			body: body ? JSON.stringify(body) : undefined
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ detail: 'Request failed' }));
			throw new Error(error.detail || 'Request failed');
		}

		return response.json();
	}

	// Conversations
	async getConversations() {
		return this.request<Conversation[]>('/chat/conversations');
	}

	async getConversation(id: string) {
		return this.request<Conversation>(`/chat/conversations/${id}`);
	}

	async createConversation(title?: string, contextId?: string) {
		return this.request<Conversation>('/chat/conversations', {
			method: 'POST',
			body: { title, context_id: contextId }
		});
	}

	async deleteConversation(id: string) {
		return this.request(`/chat/conversations/${id}`, { method: 'DELETE' });
	}

	// Chat - returns a ReadableStream for streaming
	async sendMessage(message: string, conversationId?: string, contextId?: string, model?: string) {
		const response = await fetch(`${API_BASE}/chat/message`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			credentials: 'include', // Send Better Auth cookies
			body: JSON.stringify({
				message,
				conversation_id: conversationId,
				context_id: contextId,
				model
			})
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ detail: 'Chat failed' }));
			throw new Error(error.detail || 'Chat failed');
		}

		return {
			stream: response.body,
			conversationId: response.headers.get('X-Conversation-Id')
		};
	}

	async searchConversations(query: string) {
		return this.request<SearchResult[]>(`/chat/search?q=${encodeURIComponent(query)}`);
	}

	// Projects
	async getProjects(status?: string) {
		const params = status ? `?status_filter=${status}` : '';
		return this.request<Project[]>(`/projects${params}`);
	}

	async getProject(id: string) {
		return this.request<Project>(`/projects/${id}`);
	}

	async createProject(data: CreateProjectData) {
		return this.request<Project>('/projects', { method: 'POST', body: data });
	}

	async updateProject(id: string, data: Partial<CreateProjectData>) {
		return this.request<Project>(`/projects/${id}`, { method: 'PUT', body: data });
	}

	async deleteProject(id: string) {
		return this.request(`/projects/${id}`, { method: 'DELETE' });
	}

	async addProjectNote(projectId: string, content: string) {
		return this.request(`/projects/${projectId}/notes`, {
			method: 'POST',
			body: { content }
		});
	}

	// Contexts
	async getContexts(filters?: { type?: string; includeArchived?: boolean; templatesOnly?: boolean; parentId?: string; search?: string }) {
		const params = new URLSearchParams();
		if (filters?.type) params.set('type_filter', filters.type);
		if (filters?.includeArchived) params.set('include_archived', 'true');
		if (filters?.templatesOnly) params.set('templates_only', 'true');
		if (filters?.parentId) params.set('parent_id', filters.parentId);
		if (filters?.search) params.set('search', filters.search);
		const query = params.toString();
		return this.request<ContextListItem[]>(`/contexts${query ? `?${query}` : ''}`);
	}

	async getContext(id: string) {
		return this.request<Context>(`/contexts/${id}`);
	}

	async createContext(data: CreateContextData) {
		return this.request<Context>('/contexts', { method: 'POST', body: data });
	}

	async updateContext(id: string, data: UpdateContextData) {
		return this.request<Context>(`/contexts/${id}`, { method: 'PUT', body: data });
	}

	async updateContextBlocks(id: string, data: BlocksUpdateData) {
		return this.request<Context>(`/contexts/${id}/blocks`, { method: 'PATCH', body: data });
	}

	async enableContextSharing(id: string) {
		return this.request<ShareResponse>(`/contexts/${id}/share`, { method: 'POST' });
	}

	async disableContextSharing(id: string) {
		return this.request(`/contexts/${id}/share`, { method: 'DELETE' });
	}

	async getPublicContext(shareId: string) {
		return this.request<Context>(`/contexts/public/${shareId}`);
	}

	async duplicateContext(id: string) {
		return this.request<Context>(`/contexts/${id}/duplicate`, { method: 'POST' });
	}

	async archiveContext(id: string) {
		return this.request(`/contexts/${id}/archive`, { method: 'PATCH' });
	}

	async unarchiveContext(id: string) {
		return this.request(`/contexts/${id}/unarchive`, { method: 'PATCH' });
	}

	async deleteContext(id: string) {
		return this.request(`/contexts/${id}`, { method: 'DELETE' });
	}

	async aggregateContext(data: AggregateContextRequest) {
		return this.request<AggregateContextResponse>('/contexts/aggregate', { method: 'POST', body: data });
	}

	// MCP Tools
	async getTools() {
		return this.request<{ tools: Tool[] }>('/mcp/tools');
	}

	async executeTool(toolName: string, args: Record<string, unknown>) {
		return this.request<ToolResponse>('/mcp/execute', {
			method: 'POST',
			body: { tool_name: toolName, arguments: args }
		});
	}

	// Team Members
	async getTeamMembers(status?: string) {
		const params = status ? `?status_filter=${status}` : '';
		return this.request<TeamMemberListResponse[]>(`/team${params}`);
	}

	async getTeamMember(id: string) {
		return this.request<TeamMemberDetailResponse>(`/team/${id}`);
	}

	async createTeamMember(data: CreateTeamMemberData) {
		return this.request<TeamMemberResponse>('/team', { method: 'POST', body: data });
	}

	async updateTeamMember(id: string, data: UpdateTeamMemberData) {
		return this.request<TeamMemberResponse>(`/team/${id}`, { method: 'PUT', body: data });
	}

	async deleteTeamMember(id: string) {
		return this.request(`/team/${id}`, { method: 'DELETE' });
	}

	async updateTeamMemberStatus(id: string, status: string) {
		return this.request<TeamMemberResponse>(`/team/${id}/status?new_status=${encodeURIComponent(status)}`, {
			method: 'PATCH'
		});
	}

	async updateTeamMemberCapacity(id: string, capacity: number) {
		return this.request<TeamMemberResponse>(`/team/${id}/capacity?capacity=${capacity}`, {
			method: 'PATCH'
		});
	}

	// Dashboard
	async getDashboardSummary() {
		return this.request<DashboardSummary>('/dashboard/summary');
	}

	async getFocusItems() {
		return this.request<FocusItem[]>('/dashboard/focus');
	}

	async createFocusItem(text: string) {
		return this.request<FocusItem>('/dashboard/focus', { method: 'POST', body: { text } });
	}

	async updateFocusItem(id: string, data: { text?: string; completed?: boolean }) {
		return this.request<FocusItem>(`/dashboard/focus/${id}`, { method: 'PUT', body: data });
	}

	async deleteFocusItem(id: string) {
		return this.request(`/dashboard/focus/${id}`, { method: 'DELETE' });
	}

	async getTasks(filters?: { status?: string; priority?: string; projectId?: string }) {
		const params = new URLSearchParams();
		if (filters?.status) params.set('status_filter', filters.status);
		if (filters?.priority) params.set('priority_filter', filters.priority);
		if (filters?.projectId) params.set('project_id', filters.projectId);
		const query = params.toString();
		return this.request<Task[]>(`/dashboard/tasks${query ? `?${query}` : ''}`);
	}

	async createTask(data: CreateTaskData) {
		return this.request<Task>('/dashboard/tasks', { method: 'POST', body: data });
	}

	async updateTask(id: string, data: UpdateTaskData) {
		return this.request<Task>(`/dashboard/tasks/${id}`, { method: 'PUT', body: data });
	}

	async toggleTask(id: string) {
		return this.request<Task>(`/dashboard/tasks/${id}/toggle`, { method: 'POST' });
	}

	async deleteTask(id: string) {
		return this.request(`/dashboard/tasks/${id}`, { method: 'DELETE' });
	}

	// Daily Logs
	async getDailyLogs(skip: number = 0, limit: number = 30) {
		return this.request<DailyLog[]>(`/daily/logs?skip=${skip}&limit=${limit}`);
	}

	async getTodayLog() {
		return this.request<DailyLog | null>('/daily/logs/today');
	}

	async getDailyLogByDate(date: string) {
		return this.request<DailyLog | null>(`/daily/logs/${date}`);
	}

	async saveDailyLog(data: { content: string; energy_level?: number; date?: string }) {
		return this.request<DailyLog>('/daily/logs', { method: 'POST', body: data });
	}

	async updateDailyLog(id: string, data: { content?: string; energy_level?: number }) {
		return this.request<DailyLog>(`/daily/logs/${id}`, { method: 'PUT', body: data });
	}

	async deleteDailyLog(id: string) {
		return this.request(`/daily/logs/${id}`, { method: 'DELETE' });
	}

	// Settings
	async getSettings() {
		return this.request<UserSettings>('/settings');
	}

	async updateSettings(data: UserSettingsUpdate) {
		return this.request<UserSettings>('/settings', { method: 'PUT', body: data });
	}

	async getSystemInfo() {
		return this.request<SystemInfo>('/settings/system');
	}

	// Artifacts
	async getArtifacts(filters?: { type?: string; conversationId?: string; projectId?: string; contextId?: string; unassignedOnly?: boolean }) {
		const params = new URLSearchParams();
		if (filters?.type) params.set('type', filters.type);
		if (filters?.conversationId) params.set('conversation_id', filters.conversationId);
		if (filters?.projectId) params.set('project_id', filters.projectId);
		if (filters?.contextId) params.set('context_id', filters.contextId);
		if (filters?.unassignedOnly) params.set('unassigned_only', 'true');
		const query = params.toString();
		return this.request<ArtifactListItem[]>(`/artifacts${query ? `?${query}` : ''}`);
	}

	async linkArtifact(id: string, data: { project_id?: string; context_id?: string }) {
		return this.request<Artifact>(`/artifacts/${id}/link`, { method: 'PATCH', body: data });
	}

	async getArtifact(id: string) {
		return this.request<Artifact>(`/artifacts/${id}`);
	}

	async createArtifact(data: CreateArtifactData) {
		return this.request<Artifact>('/artifacts', { method: 'POST', body: data });
	}

	async updateArtifact(id: string, data: UpdateArtifactData) {
		return this.request<Artifact>(`/artifacts/${id}`, { method: 'PATCH', body: data });
	}

	async deleteArtifact(id: string) {
		return this.request(`/artifacts/${id}`, { method: 'DELETE' });
	}

	// Nodes
	async getNodes(includeArchived = false) {
		const params = includeArchived ? '?include_archived=true' : '';
		return this.request<Node[]>(`/nodes${params}`);
	}

	async getNodeTree(includeArchived = false) {
		const params = includeArchived ? '?include_archived=true' : '';
		return this.request<NodeTree[]>(`/nodes/tree${params}`);
	}

	async getActiveNode() {
		return this.request<Node | null>('/nodes/active');
	}

	async getNode(id: string) {
		return this.request<NodeDetail>(`/nodes/${id}`);
	}

	async createNode(data: CreateNodeData) {
		return this.request<Node>('/nodes', { method: 'POST', body: data });
	}

	async updateNode(id: string, data: UpdateNodeData) {
		return this.request<Node>(`/nodes/${id}`, { method: 'PATCH', body: data });
	}

	async activateNode(id: string) {
		return this.request<NodeActivateResponse>(`/nodes/${id}/activate`, { method: 'POST' });
	}

	async deactivateNode(id: string) {
		return this.request<Node>(`/nodes/${id}/deactivate`, { method: 'POST' });
	}

	async deleteNode(id: string) {
		return this.request(`/nodes/${id}`, { method: 'DELETE' });
	}

	async getNodeChildren(id: string, includeArchived = false) {
		const params = includeArchived ? '?include_archived=true' : '';
		return this.request<Node[]>(`/nodes/${id}/children${params}`);
	}

	async reorderNode(id: string, newOrder: number) {
		return this.request(`/nodes/${id}/reorder?new_order=${newOrder}`, { method: 'POST' });
	}

	// Clients
	async getClients(filters?: { status?: ClientStatus; type?: ClientType; search?: string; tags?: string[] }) {
		const params = new URLSearchParams();
		if (filters?.status) params.set('status_filter', filters.status);
		if (filters?.type) params.set('type_filter', filters.type);
		if (filters?.search) params.set('search', filters.search);
		if (filters?.tags) {
			filters.tags.forEach(tag => params.append('tags', tag));
		}
		const query = params.toString();
		return this.request<ClientListResponse[]>(`/clients${query ? `?${query}` : ''}`);
	}

	async getClient(id: string) {
		return this.request<ClientDetailResponse>(`/clients/${id}`);
	}

	async createClient(data: CreateClientData) {
		return this.request<ClientResponse>('/clients', { method: 'POST', body: data });
	}

	async updateClient(id: string, data: UpdateClientData) {
		return this.request<ClientResponse>(`/clients/${id}`, { method: 'PUT', body: data });
	}

	async updateClientStatus(id: string, status: ClientStatus) {
		return this.request<ClientResponse>(`/clients/${id}/status`, {
			method: 'PATCH',
			body: { status }
		});
	}

	async deleteClient(id: string) {
		return this.request(`/clients/${id}`, { method: 'DELETE' });
	}

	// Client Contacts
	async getClientContacts(clientId: string) {
		return this.request<ContactResponse[]>(`/clients/${clientId}/contacts`);
	}

	async createContact(clientId: string, data: CreateContactData) {
		return this.request<ContactResponse>(`/clients/${clientId}/contacts`, { method: 'POST', body: data });
	}

	async updateContact(clientId: string, contactId: string, data: UpdateContactData) {
		return this.request<ContactResponse>(`/clients/${clientId}/contacts/${contactId}`, { method: 'PUT', body: data });
	}

	async deleteContact(clientId: string, contactId: string) {
		return this.request(`/clients/${clientId}/contacts/${contactId}`, { method: 'DELETE' });
	}

	// Client Interactions
	async getClientInteractions(clientId: string, skip = 0, limit = 50) {
		return this.request<InteractionResponse[]>(`/clients/${clientId}/interactions?skip=${skip}&limit=${limit}`);
	}

	async createInteraction(clientId: string, data: CreateInteractionData) {
		return this.request<InteractionResponse>(`/clients/${clientId}/interactions`, { method: 'POST', body: data });
	}

	// Client Deals
	async getClientDeals(clientId: string) {
		return this.request<DealResponse[]>(`/clients/${clientId}/deals`);
	}

	async createDeal(clientId: string, data: CreateDealData) {
		return this.request<DealResponse>(`/clients/${clientId}/deals`, { method: 'POST', body: data });
	}

	async updateDeal(clientId: string, dealId: string, data: UpdateDealData) {
		return this.request<DealResponse>(`/clients/${clientId}/deals/${dealId}`, { method: 'PUT', body: data });
	}

	// Deals (standalone for pipeline view)
	async getAllDeals(stage?: DealStage) {
		const params = stage ? `?stage_filter=${stage}` : '';
		return this.request<DealResponse[]>(`/deals${params}`);
	}

	async updateDealStage(dealId: string, stage: DealStage) {
		return this.request<DealResponse>(`/deals/${dealId}/stage`, {
			method: 'PATCH',
			body: { stage }
		});
	}
}

// Types
export interface Message {
	id: string;
	role: 'user' | 'assistant' | 'system';
	content: string;
	created_at: string;
	message_metadata?: Record<string, unknown>;
}

export interface Conversation {
	id: string;
	title: string;
	context_id: string | null;
	created_at: string;
	updated_at: string;
	messages: Message[];
	message_count?: number;
}

export interface SearchResult {
	message_id: string;
	conversation_id: string;
	content: string;
	role: string;
	created_at: string;
}

export interface Project {
	id: string;
	name: string;
	description: string | null;
	status: 'active' | 'paused' | 'completed' | 'archived';
	priority: 'critical' | 'high' | 'medium' | 'low';
	client_name: string | null;
	project_type: string;
	project_metadata: Record<string, unknown> | null;
	created_at: string;
	updated_at: string;
	notes: ProjectNote[];
}

export interface ProjectNote {
	id: string;
	content: string;
	created_at: string;
}

export interface CreateProjectData {
	name: string;
	description?: string;
	status?: 'active' | 'paused' | 'completed' | 'archived';
	priority?: 'critical' | 'high' | 'medium' | 'low';
	client_name?: string;
	project_type?: string;
	project_metadata?: Record<string, unknown>;
}

export type ContextType = 'person' | 'business' | 'project' | 'custom' | 'document';

export interface Block {
	id: string;
	type: string;
	content: string | null;
	properties?: Record<string, unknown>;
	children?: Block[];
}

export interface PropertySchema {
	name: string;
	type: 'text' | 'select' | 'multi_select' | 'date' | 'person' | 'relation' | 'number' | 'checkbox' | 'url' | 'email';
	options?: string[];
	relation_type?: 'context' | 'project' | 'client';
}

export interface Context {
	id: string;
	name: string;
	type: ContextType;
	content: string | null;
	structured_data: Record<string, unknown> | null;
	system_prompt_template: string | null;
	// Document editor fields
	blocks: Block[] | null;
	cover_image: string | null;
	icon: string | null;
	parent_id: string | null;
	is_template: boolean;
	is_archived: boolean;
	last_edited_at: string | null;
	word_count: number;
	is_public: boolean;
	share_id: string | null;
	// Document properties (Notion-like)
	property_schema: PropertySchema[] | null;
	properties: Record<string, unknown> | null;
	// Entity linking
	client_id: string | null;
	created_at: string;
	updated_at: string;
}

export interface ContextListItem {
	id: string;
	name: string;
	type: ContextType;
	icon: string | null;
	cover_image: string | null;
	parent_id: string | null;
	is_template: boolean;
	is_archived: boolean;
	word_count: number;
	// Document properties (Notion-like)
	property_schema: PropertySchema[] | null;
	properties: Record<string, unknown> | null;
	// Entity linking
	client_id: string | null;
	updated_at: string;
}

export interface CreateContextData {
	name: string;
	type?: ContextType;
	content?: string;
	structured_data?: Record<string, unknown>;
	system_prompt_template?: string;
	blocks?: Block[];
	cover_image?: string;
	icon?: string;
	parent_id?: string;
	is_template?: boolean;
	property_schema?: PropertySchema[];
	properties?: Record<string, unknown>;
	client_id?: string;
}

export interface UpdateContextData {
	name?: string;
	type?: ContextType;
	content?: string;
	structured_data?: Record<string, unknown>;
	system_prompt_template?: string;
	blocks?: Block[];
	cover_image?: string;
	icon?: string;
	parent_id?: string | null;
	is_template?: boolean;
	is_archived?: boolean;
	is_public?: boolean;
	property_schema?: PropertySchema[];
	properties?: Record<string, unknown>;
	client_id?: string | null;
}

export interface BlocksUpdateData {
	blocks: Block[];
	word_count?: number;
}

export interface ShareResponse {
	share_id: string;
	is_public: boolean;
	share_url: string;
}

export interface AggregateContextRequest {
	context_ids?: string[];
	project_ids?: string[];
	node_ids?: string[];
	include_children?: boolean;
	include_artifacts?: boolean;
	include_tasks?: boolean;
	max_depth?: number;
}

export interface AggregatedContextItem {
	source_type: string;
	source_id: string;
	source_name: string;
	content: string;
	metadata?: Record<string, unknown>;
}

export interface AggregateContextResponse {
	items: AggregatedContextItem[];
	total_items: number;
	total_characters: number;
	formatted_context: string;
}

export interface Tool {
	name: string;
	description: string;
	input_schema: Record<string, unknown>;
	source: 'builtin' | 'custom';
}

export interface ToolResponse {
	success: boolean;
	result: string | null;
	error: string | null;
}

// Team Member Types
export type TeamMemberStatus = 'available' | 'busy' | 'overloaded' | 'ooo';

export interface TeamMemberActivityResponse {
	id: string;
	activity_type: string;
	description: string;
	created_at: string;
}

export interface TeamMemberResponse {
	id: string;
	name: string;
	email: string;
	role: string;
	avatar_url: string | null;
	status: TeamMemberStatus;
	capacity: number;
	manager_id: string | null;
	skills: string[] | null;
	hourly_rate: number | null;
	joined_at: string;
	created_at: string;
	updated_at: string;
}

export interface TeamMemberListResponse {
	id: string;
	name: string;
	email: string;
	role: string;
	avatar_url: string | null;
	status: TeamMemberStatus;
	capacity: number;
	manager_id: string | null;
	active_projects: number;
	open_tasks: number;
	joined_at: string;
}

export interface TeamMemberDetailResponse extends TeamMemberResponse {
	active_projects: number;
	open_tasks: number;
	activities: TeamMemberActivityResponse[];
}

export interface CreateTeamMemberData {
	name: string;
	email: string;
	role: string;
	avatar_url?: string;
	manager_id?: string;
	skills?: string[];
	hourly_rate?: number;
}

export interface UpdateTeamMemberData {
	name?: string;
	email?: string;
	role?: string;
	avatar_url?: string;
	status?: TeamMemberStatus;
	capacity?: number;
	manager_id?: string | null;
	skills?: string[];
	hourly_rate?: number;
}

// Dashboard Types
export type TaskPriority = 'critical' | 'high' | 'medium' | 'low';
export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'cancelled';

export interface FocusItem {
	id: string;
	text: string;
	completed: boolean;
	focus_date: string;
	created_at: string;
}

export interface Task {
	id: string;
	title: string;
	description: string | null;
	status: TaskStatus;
	priority: TaskPriority;
	due_date: string | null;
	completed_at: string | null;
	project_id: string | null;
	assignee_id: string | null;
	created_at: string;
	updated_at: string;
}

export interface CreateTaskData {
	title: string;
	description?: string;
	priority?: TaskPriority;
	due_date?: string;
	project_id?: string;
	assignee_id?: string;
}

export interface UpdateTaskData {
	title?: string;
	description?: string;
	status?: TaskStatus;
	priority?: TaskPriority;
	due_date?: string;
	project_id?: string;
	assignee_id?: string;
}

export interface DashboardTask {
	id: string;
	title: string;
	project_name: string | null;
	due_date: string | null;
	priority: TaskPriority;
	completed: boolean;
}

export interface DashboardProject {
	id: string;
	name: string;
	client_name: string | null;
	project_type: string;
	due_date: string | null;
	progress: number;
	health: 'healthy' | 'at_risk' | 'critical';
	team_count: number;
}

export type ActivityType =
	| 'task_completed'
	| 'task_started'
	| 'project_created'
	| 'project_updated'
	| 'conversation'
	| 'team'
	| 'artifact';

export interface DashboardActivity {
	id: string;
	type: ActivityType;
	description: string;
	actor_name: string | null;
	actor_avatar: string | null;
	target_id: string | null;
	target_type: string | null;
	created_at: string;
}

export interface DashboardSummary {
	focus_items: FocusItem[];
	tasks: DashboardTask[];
	projects: DashboardProject[];
	activities: DashboardActivity[];
	energy_level: number | null;
}

// Daily Log Types
export interface DailyLog {
	id: string;
	date: string;
	content: string;
	energy_level: number | null;
	extracted_actions: Record<string, unknown> | null;
	extracted_patterns: Record<string, unknown> | null;
	created_at: string;
	updated_at: string;
}

// Settings Types
export interface UserSettings {
	id: string;
	user_id: string;
	default_model: string | null;
	email_notifications: boolean;
	daily_summary: boolean;
	theme: string;
	sidebar_collapsed: boolean;
	share_analytics: boolean;
	custom_settings: Record<string, unknown> | null;
	created_at: string;
	updated_at: string;
}

export interface UserSettingsUpdate {
	default_model?: string | null;
	email_notifications?: boolean;
	daily_summary?: boolean;
	theme?: string;
	sidebar_collapsed?: boolean;
	share_analytics?: boolean;
	custom_settings?: Record<string, unknown>;
}

export interface AvailableModel {
	name: string;
	display_name: string;
	provider: string;
	description: string | null;
}

export interface SystemInfo {
	ollama_mode: string;
	available_models: AvailableModel[];
	default_model: string;
}

// Artifact Types
export type ArtifactType = 'proposal' | 'sop' | 'framework' | 'agenda' | 'report' | 'plan' | 'code' | 'document' | 'markdown' | 'other';

export interface ArtifactListItem {
	id: string;
	title: string;
	type: ArtifactType;
	summary: string | null;
	conversation_id: string | null;
	project_id: string | null;
	context_id: string | null;
	context_name: string | null;
	created_at: string;
	updated_at: string;
}

export interface Artifact extends ArtifactListItem {
	content: string;
	version: number;
}

export interface CreateArtifactData {
	title: string;
	content: string;
	type?: ArtifactType;
	summary?: string;
	conversation_id?: string;
	project_id?: string;
}

export interface UpdateArtifactData {
	title?: string;
	content?: string;
	summary?: string;
}

// Node Types
export type NodeType = 'business' | 'project' | 'learning' | 'operational';
export type NodeHealth = 'healthy' | 'needs_attention' | 'critical' | 'not_started';

export interface DecisionItem {
	id: string;
	question: string;
	added_at: string;
	decided: boolean;
	decision: string | null;
}

export interface DelegationItem {
	id: string;
	task: string;
	assignee_id: string | null;
	assignee_name: string | null;
	status: string;
}

export interface Node {
	id: string;
	user_id: string;
	parent_id: string | null;
	context_id: string | null;
	name: string;
	type: NodeType;
	health: NodeHealth;
	purpose: string | null;
	current_status: string | null;
	this_week_focus: string[] | null;
	decision_queue: DecisionItem[] | null;
	delegation_ready: DelegationItem[] | null;
	is_active: boolean;
	is_archived: boolean;
	sort_order: number;
	created_at: string;
	updated_at: string;
}

export interface NodeTree extends Node {
	children: NodeTree[];
	children_count: number;
}

export interface NodeDetail extends Node {
	parent_name: string | null;
	children_count: number;
	linked_projects_count: number;
	linked_conversations_count: number;
	linked_artifacts_count: number;
}

export interface NodeActivateResponse {
	node: Node;
	previous_active_id: string | null;
	context_prompt: string | null;
}

export interface CreateNodeData {
	name: string;
	type: NodeType;
	parent_id?: string;
	purpose?: string;
	context_id?: string;
}

export interface UpdateNodeData {
	name?: string;
	type?: NodeType;
	parent_id?: string | null;
	health?: NodeHealth;
	purpose?: string;
	current_status?: string;
	this_week_focus?: string[];
	decision_queue?: DecisionItem[];
	delegation_ready?: DelegationItem[];
	is_active?: boolean;
	is_archived?: boolean;
	sort_order?: number;
	context_id?: string;
}

// Client Types
export type ClientType = 'company' | 'individual';
export type ClientStatus = 'lead' | 'prospect' | 'active' | 'inactive' | 'churned';
export type InteractionType = 'call' | 'email' | 'meeting' | 'note';
export type DealStage = 'qualification' | 'proposal' | 'negotiation' | 'closed_won' | 'closed_lost';

export interface ContactResponse {
	id: string;
	client_id: string;
	name: string;
	email: string | null;
	phone: string | null;
	role: string | null;
	is_primary: boolean;
	notes: string | null;
	created_at: string;
	updated_at: string;
}

export interface InteractionResponse {
	id: string;
	client_id: string;
	contact_id: string | null;
	type: InteractionType;
	subject: string;
	description: string | null;
	outcome: string | null;
	occurred_at: string;
	created_at: string;
}

export interface DealResponse {
	id: string;
	client_id: string;
	name: string;
	value: number;
	stage: DealStage;
	probability: number;
	expected_close_date: string | null;
	notes: string | null;
	created_at: string;
	updated_at: string;
	closed_at: string | null;
}

export interface ClientResponse {
	id: string;
	user_id: string;
	name: string;
	type: ClientType;
	email: string | null;
	phone: string | null;
	website: string | null;
	industry: string | null;
	company_size: string | null;
	address: string | null;
	city: string | null;
	state: string | null;
	zip_code: string | null;
	country: string | null;
	status: ClientStatus;
	source: string | null;
	assigned_to: string | null;
	lifetime_value: number | null;
	tags: string[] | null;
	custom_fields: Record<string, unknown> | null;
	notes: string | null;
	created_at: string;
	updated_at: string;
	last_contacted_at: string | null;
}

export interface ClientDetailResponse extends ClientResponse {
	contacts: ContactResponse[];
	interactions: InteractionResponse[];
	deals: DealResponse[];
}

export interface ClientListResponse {
	id: string;
	name: string;
	type: ClientType;
	email: string | null;
	phone: string | null;
	status: ClientStatus;
	source: string | null;
	assigned_to: string | null;
	lifetime_value: number | null;
	tags: string[] | null;
	created_at: string;
	last_contacted_at: string | null;
	contacts_count: number;
	interactions_count: number;
	deals_count: number;
	active_deals_value: number;
}

export interface CreateClientData {
	name: string;
	type?: ClientType;
	email?: string;
	phone?: string;
	website?: string;
	industry?: string;
	company_size?: string;
	address?: string;
	city?: string;
	state?: string;
	zip_code?: string;
	country?: string;
	status?: ClientStatus;
	source?: string;
	assigned_to?: string;
	tags?: string[];
	custom_fields?: Record<string, unknown>;
	notes?: string;
}

export interface UpdateClientData {
	name?: string;
	type?: ClientType;
	email?: string;
	phone?: string;
	website?: string;
	industry?: string;
	company_size?: string;
	address?: string;
	city?: string;
	state?: string;
	zip_code?: string;
	country?: string;
	status?: ClientStatus;
	source?: string;
	assigned_to?: string;
	lifetime_value?: number;
	tags?: string[];
	custom_fields?: Record<string, unknown>;
	notes?: string;
}

export interface CreateContactData {
	name: string;
	email?: string;
	phone?: string;
	role?: string;
	is_primary?: boolean;
	notes?: string;
}

export interface UpdateContactData {
	name?: string;
	email?: string;
	phone?: string;
	role?: string;
	is_primary?: boolean;
	notes?: string;
}

export interface CreateInteractionData {
	type: InteractionType;
	subject: string;
	description?: string;
	outcome?: string;
	contact_id?: string;
	occurred_at?: string;
}

export interface CreateDealData {
	name: string;
	value?: number;
	stage?: DealStage;
	probability?: number;
	expected_close_date?: string;
	notes?: string;
}

export interface UpdateDealData {
	name?: string;
	value?: number;
	stage?: DealStage;
	probability?: number;
	expected_close_date?: string;
	notes?: string;
}

export const api = new ApiClient();
