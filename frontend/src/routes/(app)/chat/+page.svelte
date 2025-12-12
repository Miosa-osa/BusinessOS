<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { api, type ArtifactListItem, type Artifact, type Node, type ContextListItem } from '$lib/api/client';

	// Message interface
	interface ChatMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		artifacts?: { title: string; type: string; content: string }[];
	}

	// UI State
	let messagesContainer: HTMLDivElement | undefined = $state(undefined);
	let inputRef: HTMLTextAreaElement | undefined = $state(undefined);
	let inputValue = $state('');
	let selectedModel = $state('qwen3-coder:480b-cloud');
	let selectedContextIds = $state<string[]>([]);
	let chatSidebarOpen = $state(true);
	let artifactsPanelOpen = $state(false);
	let searchQuery = $state('');
	let showContextDropdown = $state(false);
	let showHeaderContextDropdown = $state(false);
	let showModelDropdown = $state(false);
	let showNodeDropdown = $state(false);
	let copiedMessageId: string | null = $state(null);
	let filterTab: 'all' | 'pinned' | 'recent' = $state('all');

	// Contexts state
	let availableContexts: ContextListItem[] = $state([]);
	let loadingContexts = $state(false);

	// Chat state
	let messages: ChatMessage[] = $state([]);
	let isStreaming = $state(false);
	let conversationId: string | null = $state(null);
	let abortController: AbortController | null = $state(null);
	let loadingConversation = $state(false);

	// Active node state
	let activeNode: Node | null = $state(null);
	let nodeContextPrompt: string | null = $state(null);

	// Artifacts state
	let artifacts: ArtifactListItem[] = $state([]);
	let selectedArtifact: Artifact | null = $state(null);
	let loadingArtifacts = $state(false);
	let artifactFilter: string = $state('all');

	// Artifact generation state (for live preview)
	let generatingArtifact = $state(false);
	let generatingArtifactTitle = $state('');
	let generatingArtifactType = $state('');
	let generatingArtifactContent = $state('');
	let artifactCompletedInStream = $state(false); // Track if artifact completed during current stream

	// Resizable panel state - default to 50% of available space (will be set in onMount)
	let artifactPanelWidth = $state(600);
	let isResizing = $state(false);
	let resizeStartX = $state(0);
	let resizeStartWidth = $state(0);

	// Currently viewing artifact in panel
	let viewingArtifactFromMessage: { title: string; type: string; content: string } | null = $state(null);

	// Editable artifact state
	let isEditingArtifact = $state(false);
	let editedArtifactContent = $state('');

	// Save to profile modal (artifacts become documents in profiles)
	let showSaveToProfileModal = $state(false);
	let availableProfiles: ContextListItem[] = $state([]);
	let selectedProfileForSave: string = $state('');
	let savingArtifactToProfile = $state(false);

	// Legacy - keeping for compatibility
	let showSaveToNodeModal = $state(false);
	let availableNodes: Node[] = $state([]);
	let selectedNodeForSave: string = $state('');

	// Project-first chat state
	interface ProjectItem {
		id: string;
		name: string;
		description?: string;
	}
	let selectedProjectId = $state<string | null>(null);
	let showProjectDropdown = $state(false);
	let projectsList = $state<ProjectItem[]>([]);
	let loadingProjects = $state(false);

	// Derived project info
	let selectedProject = $derived(
		selectedProjectId ? projectsList.find(p => p.id === selectedProjectId) : null
	);

	// Task generation from artifact
	interface GeneratedTask {
		title: string;
		description: string;
		priority: 'low' | 'medium' | 'high';
		assignee_id?: string;
		estimated_hours?: number;
	}
	let showTaskGenerationModal = $state(false);
	let generatingTasks = $state(false);
	let generatedTasks = $state<GeneratedTask[]>([]);
	let selectedProjectForTasks = $state<string>('');
	let taskGenerationArtifact = $state<{ title: string; type: string; content: string } | null>(null);
	let availableProjects = $state<{ id: string; name: string }[]>([]);
	let availableTeamMembers = $state<{ id: string; name: string; role: string }[]>([]);

	// Inline task creation state (after artifact)
	let showInlineTaskCreation = $state(false);
	let inlineTasksForArtifact = $state<GeneratedTask[]>([]);
	let creatingInlineTasks = $state(false);

	// Load available nodes for saving
	async function loadAvailableNodes() {
		try {
			availableNodes = await api.getNodes();
		} catch (e) {
			console.error('Failed to load nodes:', e);
		}
	}

	function startEditingArtifact() {
		if (viewingArtifactFromMessage) {
			editedArtifactContent = viewingArtifactFromMessage.content;
			isEditingArtifact = true;
		}
	}

	function saveArtifactEdit() {
		if (viewingArtifactFromMessage) {
			viewingArtifactFromMessage = {
				...viewingArtifactFromMessage,
				content: editedArtifactContent
			};
			isEditingArtifact = false;
		}
	}

	function cancelArtifactEdit() {
		isEditingArtifact = false;
		editedArtifactContent = '';
	}

	async function openSaveToNodeModal() {
		await loadAvailableNodes();
		showSaveToNodeModal = true;
	}

	// Load available profiles (non-document contexts) for saving artifacts
	async function loadAvailableProfiles() {
		try {
			const contexts = await api.getContexts();
			// Filter to only profiles (non-document contexts)
			availableProfiles = contexts.filter(c => c.type !== 'document');
		} catch (e) {
			console.error('Failed to load profiles:', e);
		}
	}

	// Open save to profile modal
	function openSaveToProfileModal() {
		loadAvailableProfiles();
		showSaveToProfileModal = true;
		selectedProfileForSave = '';
	}

	// Save artifact as a document in a profile
	async function saveArtifactToProfile() {
		if (!selectedProfileForSave || !viewingArtifactFromMessage) return;

		savingArtifactToProfile = true;
		try {
			// Create a new context document with the artifact content
			await api.createContext({
				name: viewingArtifactFromMessage.title,
				type: 'document',
				content: viewingArtifactFromMessage.content,
				parent_id: selectedProfileForSave,
				icon: viewingArtifactFromMessage.type === 'plan' ? '📋' :
					  viewingArtifactFromMessage.type === 'proposal' ? '📄' :
					  viewingArtifactFromMessage.type === 'framework' ? '🏗️' :
					  viewingArtifactFromMessage.type === 'sop' ? '📖' :
					  viewingArtifactFromMessage.type === 'report' ? '📊' : '📝'
			});

			showSaveToProfileModal = false;
			selectedProfileForSave = '';
			viewingArtifactFromMessage = null;
		} catch (e) {
			console.error('Failed to save artifact to profile:', e);
		} finally {
			savingArtifactToProfile = false;
		}
	}

	// Legacy function - redirect to new profile modal
	async function saveArtifactToNode() {
		openSaveToProfileModal();
	}

	// Generate tasks from artifact
	async function generateTasksFromArtifact(artifact: { title: string; type: string; content: string }) {
		taskGenerationArtifact = artifact;
		showTaskGenerationModal = true;
		generatingTasks = true;
		generatedTasks = [];

		// Load projects for assignment
		try {
			const projects = await api.getProjects();
			availableProjects = projects.map(p => ({ id: p.id, name: p.name }));
		} catch (e) {
			console.error('Failed to load projects:', e);
		}

		// Load team members
		try {
			const team = await api.getTeamMembers();
			availableTeamMembers = team.map(m => ({ id: m.id, name: m.name, role: m.role || 'Member' }));
		} catch (e) {
			console.error('Failed to load team members:', e);
		}

		// Call AI to extract tasks from artifact
		try {
			const response = await fetch('/api/chat/ai/extract-tasks', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					artifact_title: artifact.title,
					artifact_content: artifact.content,
					artifact_type: artifact.type,
					team_members: availableTeamMembers
				})
			});

			if (!response.ok) throw new Error('Failed to extract tasks');

			const data = await response.json();
			generatedTasks = data.tasks || [];
		} catch (e) {
			console.error('Failed to generate tasks:', e);
			// Fallback: show empty state with manual entry option
			generatedTasks = [];
		} finally {
			generatingTasks = false;
		}
	}

	async function confirmTaskCreation() {
		if (!selectedProjectForTasks || generatedTasks.length === 0) return;

		try {
			// Create tasks via API
			for (const task of generatedTasks) {
				await api.createTask({
					title: task.title,
					description: task.description,
					project_id: selectedProjectForTasks,
					priority: task.priority,
					assignee_id: task.assignee_id,
					status: 'todo'
				});
			}

			// Close modal and show success
			showTaskGenerationModal = false;
			generatedTasks = [];
			taskGenerationArtifact = null;

			// Add confirmation message to chat
			const confirmMsgId = crypto.randomUUID();
			messages = [...messages, {
				id: confirmMsgId,
				role: 'assistant',
				content: `✅ Created ${generatedTasks.length} tasks from "${taskGenerationArtifact?.title}". You can view them in the Tasks tab.`
			}];
		} catch (e) {
			console.error('Failed to create tasks:', e);
		}
	}

	function removeGeneratedTask(index: number) {
		generatedTasks = generatedTasks.filter((_, i) => i !== index);
	}

	function updateTaskAssignee(index: number, assigneeId: string) {
		generatedTasks = generatedTasks.map((task, i) =>
			i === index ? { ...task, assignee_id: assigneeId } : task
		);
	}

	// Inline task creation - triggered automatically after actionable artifacts
	async function triggerInlineTaskCreation(artifact: { title: string; type: string; content: string }) {
		showInlineTaskCreation = true;
		creatingInlineTasks = true;
		inlineTasksForArtifact = [];

		try {
			// Call AI to extract tasks
			const response = await fetch('/api/chat/ai/extract-tasks', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					artifact_title: artifact.title,
					artifact_content: artifact.content,
					artifact_type: artifact.type,
					team_members: availableTeamMembers
				})
			});

			if (!response.ok) throw new Error('Failed to extract tasks');

			const data = await response.json();
			const tasks = data.tasks || [];

			// Auto-assign tasks based on team member roles
			inlineTasksForArtifact = tasks.map((task: GeneratedTask) => {
				// Try to auto-assign based on task keywords and team member roles
				const assignee = findBestAssignee(task);
				return { ...task, assignee_id: assignee?.id };
			});
		} catch (e) {
			console.error('Failed to generate tasks:', e);
			inlineTasksForArtifact = [];
		} finally {
			creatingInlineTasks = false;
		}
	}

	// Auto-assign tasks based on team member roles/skills
	function findBestAssignee(task: GeneratedTask): { id: string; name: string; role: string } | undefined {
		const title = task.title.toLowerCase();
		const desc = (task.description || '').toLowerCase();
		const combined = title + ' ' + desc;

		// Role-based matching keywords
		const roleKeywords: Record<string, string[]> = {
			'developer': ['code', 'implement', 'build', 'develop', 'api', 'frontend', 'backend', 'database', 'bug', 'fix', 'feature', 'technical', 'integration'],
			'designer': ['design', 'ui', 'ux', 'mockup', 'wireframe', 'visual', 'layout', 'style', 'brand'],
			'project manager': ['coordinate', 'schedule', 'timeline', 'milestone', 'meeting', 'stakeholder', 'plan', 'track', 'report'],
			'ceo': ['strategy', 'vision', 'decision', 'executive', 'leadership', 'partnership', 'investor'],
			'cto': ['architecture', 'infrastructure', 'security', 'scalability', 'technical strategy', 'technology'],
			'marketing': ['marketing', 'campaign', 'content', 'social', 'seo', 'advertising', 'promotion', 'brand'],
			'sales': ['sales', 'client', 'customer', 'deal', 'proposal', 'pitch', 'revenue', 'lead'],
			'operations': ['operations', 'process', 'workflow', 'efficiency', 'sop', 'documentation'],
			'qa': ['test', 'quality', 'qa', 'bug', 'verification', 'validation'],
			'devops': ['deploy', 'ci/cd', 'infrastructure', 'monitoring', 'server', 'cloud', 'kubernetes', 'docker']
		};

		// Score each team member
		let bestMatch: { member: typeof availableTeamMembers[0]; score: number } | null = null;

		for (const member of availableTeamMembers) {
			const memberRole = member.role.toLowerCase();
			let score = 0;

			// Check if member's role matches any keywords
			for (const [role, keywords] of Object.entries(roleKeywords)) {
				if (memberRole.includes(role)) {
					for (const keyword of keywords) {
						if (combined.includes(keyword)) {
							score += 10;
						}
					}
				}
			}

			// Also check direct role match in task
			if (combined.includes(memberRole)) {
				score += 20;
			}

			if (score > 0 && (!bestMatch || score > bestMatch.score)) {
				bestMatch = { member, score };
			}
		}

		return bestMatch?.member;
	}

	// Confirm and create tasks inline
	async function confirmInlineTasks() {
		if (!selectedProjectId || inlineTasksForArtifact.length === 0) return;

		creatingInlineTasks = true;
		try {
			// Create tasks via API
			for (const task of inlineTasksForArtifact) {
				await api.createTask({
					title: task.title,
					description: task.description,
					project_id: selectedProjectId,
					priority: task.priority,
					assignee_id: task.assignee_id,
					status: 'todo'
				});
			}

			// Add confirmation message to chat
			const count = inlineTasksForArtifact.length;
			const confirmMsgId = crypto.randomUUID();
			messages = [...messages, {
				id: confirmMsgId,
				role: 'assistant',
				content: `Created ${count} task${count > 1 ? 's' : ''} from the artifact. You can view them in the Tasks tab or project dashboard.`
			}];

			// Reset state
			showInlineTaskCreation = false;
			inlineTasksForArtifact = [];
		} catch (e) {
			console.error('Failed to create tasks:', e);
		} finally {
			creatingInlineTasks = false;
		}
	}

	function dismissInlineTasks() {
		showInlineTaskCreation = false;
		inlineTasksForArtifact = [];
	}

	function updateInlineTaskAssignee(index: number, assigneeId: string) {
		inlineTasksForArtifact = inlineTasksForArtifact.map((task, i) =>
			i === index ? { ...task, assignee_id: assigneeId } : task
		);
	}

	function removeInlineTask(index: number) {
		inlineTasksForArtifact = inlineTasksForArtifact.filter((_, i) => i !== index);
	}

	// Load available contexts
	async function loadContexts() {
		loadingContexts = true;
		try {
			availableContexts = await api.getContexts();
		} catch (e) {
			console.error('Failed to load contexts:', e);
		} finally {
			loadingContexts = false;
		}
	}

	// Load available projects for project-first chat
	async function loadProjects() {
		loadingProjects = true;
		try {
			const projects = await api.getProjects();
			projectsList = projects.map(p => ({
				id: p.id,
				name: p.name,
				description: p.description
			}));
			// Also update availableProjects for task generation
			availableProjects = projectsList;
		} catch (e) {
			console.error('Failed to load projects:', e);
		} finally {
			loadingProjects = false;
		}
	}

	// Load team members for task assignment
	async function loadTeamMembers() {
		try {
			const team = await api.getTeamMembers();
			availableTeamMembers = team.map(m => ({
				id: m.id,
				name: m.name,
				role: m.role || 'Member'
			}));
		} catch (e) {
			console.error('Failed to load team members:', e);
		}
	}

	// Load active node on mount
	async function loadActiveNode() {
		try {
			activeNode = await api.getActiveNode();
			if (activeNode) {
				// Build context prompt from node data
				const focusItems = activeNode.this_week_focus?.map((f, i) => `${i + 1}. ${f}`).join('\n') || 'Not defined';
				nodeContextPrompt = `Current Active Node: ${activeNode.name}

Purpose: ${activeNode.purpose || 'Not defined'}

Current Status: ${activeNode.current_status || 'Not defined'}

This Week's Focus:
${focusItems}

Use this context to inform your responses.`;
			}
		} catch (e) {
			console.error('Failed to load active node:', e);
		}
	}

	async function handleDeactivateNode() {
		if (!activeNode) return;
		try {
			await api.deactivateNode(activeNode.id);
			activeNode = null;
			nodeContextPrompt = null;
			showNodeDropdown = false;
		} catch (e) {
			console.error('Failed to deactivate node:', e);
		}
	}

	onMount(() => {
		loadActiveNode();
		loadContexts();
		loadConversations();
		loadProjects();
		loadTeamMembers();
		// Set artifact panel width to ~50% of available space (window width minus sidebars)
		// Left sidebar is ~256px, chat sidebar is ~256px when open
		const availableWidth = window.innerWidth - 256; // Subtract left sidebar
		artifactPanelWidth = Math.floor(availableWidth / 2);
	});

	// Load artifacts
	async function loadArtifacts() {
		loadingArtifacts = true;
		try {
			// Load all artifacts for user, optionally filter by type
			// Don't filter by conversationId or projectId so we can see all artifacts
			const filters: { type?: string } = {};
			if (artifactFilter !== 'all') filters.type = artifactFilter;
			console.log('[loadArtifacts] Loading artifacts with filters:', filters);
			const result = await api.getArtifacts(filters);
			console.log('[loadArtifacts] Loaded', result.length, 'artifacts');
			artifacts = result;
		} catch (error) {
			console.error('Failed to load artifacts:', error);
			artifacts = [];
		} finally {
			loadingArtifacts = false;
		}
	}

	// Track if artifacts have been loaded this session
	let artifactsLoadedOnce = $state(false);

	// Load artifacts when panel opens (always reload on first open)
	$effect(() => {
		if (artifactsPanelOpen && !artifactsLoadedOnce) {
			artifactsLoadedOnce = true;
			loadArtifacts();
		}
	});

	async function selectArtifact(id: string) {
		try {
			selectedArtifact = await api.getArtifact(id);
		} catch (error) {
			console.error('Failed to load artifact:', error);
		}
	}

	function closeArtifactDetail() {
		selectedArtifact = null;
	}

	function getArtifactIcon(type: string) {
		switch (type) {
			case 'proposal': return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'sop': return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4';
			case 'framework': return 'M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z';
			case 'agenda': return 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z';
			case 'report': return 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'plan': return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2';
			default: return 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z';
		}
	}

	function getArtifactColor(type: string) {
		switch (type) {
			case 'proposal': return 'text-blue-500 bg-blue-50';
			case 'sop': return 'text-green-500 bg-green-50';
			case 'framework': return 'text-purple-500 bg-purple-50';
			case 'agenda': return 'text-orange-500 bg-orange-50';
			case 'report': return 'text-red-500 bg-red-50';
			case 'plan': return 'text-teal-500 bg-teal-50';
			default: return 'text-gray-500 bg-gray-50';
		}
	}

	// Load artifacts when panel opens
	$effect(() => {
		if (artifactsPanelOpen) {
			loadArtifacts();
		}
	});

	// Available models
	interface ModelOption {
		id: string;
		name: string;
		description: string;
		type: 'cloud' | 'local';
	}

	const models: ModelOption[] = [
		// Cloud models (via Ollama Cloud API)
		{ id: 'qwen3-coder:480b-cloud', name: 'Qwen3 Coder 480B', description: 'Best for coding tasks (Cloud)', type: 'cloud' },
		{ id: 'gpt-4o', name: 'GPT-4o', description: 'OpenAI flagship model', type: 'cloud' },
		{ id: 'gpt-4o-mini', name: 'GPT-4o Mini', description: 'Fast & affordable', type: 'cloud' },
		{ id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', description: 'Anthropic balanced model', type: 'cloud' },
		{ id: 'claude-3-5-haiku', name: 'Claude 3.5 Haiku', description: 'Fast & efficient', type: 'cloud' },
		// Local models (requires local Ollama)
		{ id: 'qwen3:latest', name: 'Qwen 3', description: 'Latest Qwen 3 model', type: 'local' },
		{ id: 'qwen3:32b', name: 'Qwen 3 32B', description: 'Large Qwen 3 model', type: 'local' },
		{ id: 'qwen3:8b', name: 'Qwen 3 8B', description: 'Medium Qwen 3 model', type: 'local' },
		{ id: 'qwen3-coder:latest', name: 'Qwen3 Coder', description: 'Best for coding', type: 'local' },
		{ id: 'qwen3-coder:30b', name: 'Qwen3 Coder 30B', description: 'Large coding model', type: 'local' },
		{ id: 'qwen2.5:latest', name: 'Qwen 2.5', description: 'Previous gen Qwen', type: 'local' },
		{ id: 'qwen2.5:7b', name: 'Qwen 2.5 7B', description: 'Fast general purpose', type: 'local' },
		{ id: 'qwen2.5:32b', name: 'Qwen 2.5 32B', description: 'Large Qwen 2.5', type: 'local' },
		{ id: 'llama3.3:latest', name: 'Llama 3.3', description: 'Meta latest model', type: 'local' },
		{ id: 'llama3.3:70b', name: 'Llama 3.3 70B', description: 'Meta large model', type: 'local' },
		{ id: 'llama3.2:latest', name: 'Llama 3.2', description: 'Meta efficient model', type: 'local' },
		{ id: 'llama3.2:3b', name: 'Llama 3.2 3B', description: 'Ultra fast & light', type: 'local' },
		{ id: 'llama3.1:latest', name: 'Llama 3.1', description: 'Meta previous gen', type: 'local' },
		{ id: 'llama3.1:70b', name: 'Llama 3.1 70B', description: 'Large Llama 3.1', type: 'local' },
		{ id: 'deepseek-r1:latest', name: 'DeepSeek R1', description: 'Reasoning model', type: 'local' },
		{ id: 'deepseek-r1:32b', name: 'DeepSeek R1 32B', description: 'Large reasoning', type: 'local' },
		{ id: 'deepseek-coder:latest', name: 'DeepSeek Coder', description: 'Coding specialist', type: 'local' },
		{ id: 'deepseek-coder-v2:latest', name: 'DeepSeek Coder V2', description: 'Latest coding model', type: 'local' },
		{ id: 'codellama:latest', name: 'Code Llama', description: 'Meta coding model', type: 'local' },
		{ id: 'codellama:34b', name: 'Code Llama 34B', description: 'Large Code Llama', type: 'local' },
		{ id: 'mistral:latest', name: 'Mistral', description: 'Mistral AI model', type: 'local' },
		{ id: 'mixtral:latest', name: 'Mixtral 8x7B', description: 'MoE model', type: 'local' },
		{ id: 'gemma2:latest', name: 'Gemma 2', description: 'Google model', type: 'local' },
		{ id: 'gemma2:27b', name: 'Gemma 2 27B', description: 'Large Gemma', type: 'local' },
		{ id: 'phi3:latest', name: 'Phi 3', description: 'Microsoft model', type: 'local' },
		{ id: 'phi3:14b', name: 'Phi 3 14B', description: 'Medium Phi 3', type: 'local' },
		{ id: 'starcoder2:latest', name: 'StarCoder 2', description: 'Code generation', type: 'local' },
		{ id: 'yi:latest', name: 'Yi', description: '01.AI model', type: 'local' },
		{ id: 'yi:34b', name: 'Yi 34B', description: 'Large Yi model', type: 'local' },
		{ id: 'command-r:latest', name: 'Command R', description: 'Cohere model', type: 'local' },
		{ id: 'neural-chat:latest', name: 'Neural Chat', description: 'Intel optimized', type: 'local' },
		{ id: 'dolphin-mixtral:latest', name: 'Dolphin Mixtral', description: 'Uncensored MoE', type: 'local' },
	];

	const cloudModels = models.filter(m => m.type === 'cloud');
	const localModels = models.filter(m => m.type === 'local');

	// Sidebar conversations
	interface SidebarConversation {
		id: string;
		title: string;
		timestamp: string;
		pinned?: boolean;
	}

	let conversations: SidebarConversation[] = $state([]);
	let activeConversationId = $state<string | null>(null);

	// Derived context
	let selectedContexts = $derived<ContextListItem[]>(
		selectedContextIds.length > 0
			? availableContexts.filter(c => selectedContextIds.includes(c.id))
			: []
	);

	// Helper for displaying selected contexts
	let selectedContextsLabel = $derived(
		selectedContexts.length === 0
			? 'Select Context'
			: selectedContexts.length === 1
				? selectedContexts[0].name
				: `${selectedContexts.length} contexts`
	);

	// Quick action prompts
	const quickActions = [
		'Write a business proposal',
		'Analyze my data',
		'Plan my week'
	];

	// Personalized greeting state
	let userName = $state('Roberto'); // TODO: Fetch from user profile
	let currentSuggestionIndex = $state(0);
	let displayedSuggestion = $state('');
	let isTyping = $state(true);
	let typewriterPaused = $state(false);

	// Time-aware greeting suggestions that rotate
	const greetingSuggestions = [
		'streamline your workflow',
		'automate repetitive tasks',
		'create a business proposal',
		'analyze your metrics',
		'draft a client email',
		'plan your week ahead',
		'optimize your processes'
	];

	// Get personalized greeting based on time of day
	function getTimeBasedGreeting(): string {
		const hour = new Date().getHours();
		if (hour >= 0 && hour < 5) {
			return `Up late, ${userName}?`;
		} else if (hour >= 5 && hour < 12) {
			return `Good morning, ${userName}`;
		} else if (hour >= 12 && hour < 17) {
			return `Good afternoon, ${userName}`;
		} else if (hour >= 17 && hour < 21) {
			return `Good evening, ${userName}`;
		} else {
			return `Working late, ${userName}?`;
		}
	}

	// Derived greeting
	let personalizedGreeting = $derived(getTimeBasedGreeting());

	// Derived state (moved before effect that uses it)
	let hasConversation = $derived(messages.length > 0 || loadingConversation);
	let currentModelName = $derived(models.find(m => m.id === selectedModel)?.name ?? selectedModel);

	// Typewriter effect for suggestions
	$effect(() => {
		if (hasConversation) return; // Don't run when there's a conversation

		const currentSuggestion = greetingSuggestions[currentSuggestionIndex];
		let charIndex = 0;
		let direction: 'typing' | 'deleting' | 'pausing' = 'typing';
		let timeoutId: ReturnType<typeof setTimeout>;

		function tick() {
			if (direction === 'typing') {
				if (charIndex <= currentSuggestion.length) {
					displayedSuggestion = currentSuggestion.slice(0, charIndex);
					charIndex++;
					timeoutId = setTimeout(tick, 50 + Math.random() * 30); // Variable typing speed
				} else {
					direction = 'pausing';
					timeoutId = setTimeout(tick, 2500); // Pause at full text
				}
			} else if (direction === 'pausing') {
				direction = 'deleting';
				timeoutId = setTimeout(tick, 50);
			} else if (direction === 'deleting') {
				if (charIndex > 0) {
					charIndex--;
					displayedSuggestion = currentSuggestion.slice(0, charIndex);
					timeoutId = setTimeout(tick, 25); // Faster deletion
				} else {
					// Move to next suggestion
					currentSuggestionIndex = (currentSuggestionIndex + 1) % greetingSuggestions.length;
				}
			}
		}

		tick();

		return () => {
			clearTimeout(timeoutId);
		};
	});

	// Auto-scroll on new messages
	$effect(() => {
		if (messagesContainer && messages.length) {
			tick().then(() => {
				if (messagesContainer) {
					messagesContainer.scrollTop = messagesContainer.scrollHeight;
				}
			});
		}
	});

	function handleQuickAction(prompt: string) {
		inputValue = prompt;
		inputRef?.focus();
	}

	function handleNewChat() {
		messages = [];
		conversationId = null;
		activeConversationId = null;
	}

	// Load conversations from API
	async function loadConversations() {
		try {
			const convs = await api.getConversations();
			conversations = convs.map(c => ({
				id: c.id,
				title: c.title,
				timestamp: c.updated_at,
				pinned: false
			}));
		} catch (e) {
			console.error('Failed to load conversations:', e);
		}
	}

	// Helper function to parse artifacts from message content
	function parseArtifactsFromContent(content: string): { cleanContent: string; artifacts: { title: string; type: string; content: string }[] } {
		const artifacts: { title: string; type: string; content: string }[] = [];
		let cleanContent = content;

		// Find all artifact blocks
		const artifactRegex = /```artifact\s*\n([\s\S]*?)\n```/g;
		let match;

		while ((match = artifactRegex.exec(content)) !== null) {
			try {
				const artifactData = JSON.parse(match[1].trim());
				if (artifactData.title && artifactData.type && artifactData.content) {
					artifacts.push({
						title: artifactData.title,
						type: artifactData.type,
						content: artifactData.content
							.replace(/\\n/g, '\n')
							.replace(/\\"/g, '"')
							.replace(/\\\\/g, '\\')
					});
				}
			} catch {
				console.error('Failed to parse artifact JSON');
			}
		}

		// Remove artifact blocks from displayed content
		cleanContent = content.replace(/```artifact\s*\n[\s\S]*?\n```/g, '').trim();

		return { cleanContent, artifacts };
	}

	async function selectConversation(id: string) {
		activeConversationId = id;
		conversationId = id;
		loadingConversation = true;
		artifactsLoadedOnce = false; // Reset so artifacts reload when panel opens

		// Load conversation messages from backend
		try {
			const conv = await api.getConversation(id);
			console.log('[selectConversation] Loaded conversation:', conv.id, 'with', conv.messages.length, 'messages');

			messages = conv.messages.map(m => {
				if (m.role === 'assistant') {
					// Parse artifacts from assistant messages
					const hasArtifactBlock = m.content.includes('```artifact');
					console.log('[selectConversation] Assistant message', m.id, 'has artifact block:', hasArtifactBlock);
					if (hasArtifactBlock) {
						console.log('[selectConversation] Content preview:', m.content.substring(0, 500));
					}

					const { cleanContent, artifacts } = parseArtifactsFromContent(m.content);
					console.log('[selectConversation] Parsed artifacts:', artifacts.length, 'artifacts found');

					return {
						id: m.id,
						role: m.role as 'user' | 'assistant',
						content: cleanContent,
						artifacts: artifacts.length > 0 ? artifacts : undefined
					};
				}
				return {
					id: m.id,
					role: m.role as 'user' | 'assistant',
					content: m.content
				};
			});

			// Load artifacts for this conversation
			loadArtifacts();
		} catch (e) {
			console.error('Failed to load conversation:', e);
		} finally {
			loadingConversation = false;
		}
	}

	async function handleSendMessage() {
		if (!inputValue.trim() || isStreaming) return;

		// Require project selection before chatting
		if (!selectedProjectId) {
			showProjectDropdown = true;
			return;
		}

		const userMessage = inputValue.trim();
		inputValue = '';
		if (inputRef) inputRef.style.height = 'auto';

		// Reset artifact state for new message
		artifactCompletedInStream = false;
		showInlineTaskCreation = false;

		// Add user message to UI
		const userMsgId = crypto.randomUUID();
		messages = [...messages, { id: userMsgId, role: 'user', content: userMessage }];

		// Create assistant message placeholder
		const assistantMsgId = crypto.randomUUID();
		messages = [...messages, { id: assistantMsgId, role: 'assistant', content: '' }];

		isStreaming = true;
		abortController = new AbortController();

		try {
			// Build request body with context and node context
			// Note: The backend will load full context details (content, system_prompt_template)
			// using the context_id, so we just pass the ID here
			const requestBody: Record<string, unknown> = {
				message: userMessage,
				model: selectedModel,
				conversation_id: conversationId,
				project_id: selectedProjectId,
				context_id: selectedContextIds.length > 0 ? selectedContextIds[0] : null,
				context_ids: selectedContextIds.length > 0 ? selectedContextIds : undefined,
			};

			// Include node context if there's an active node
			if (nodeContextPrompt) {
				requestBody.node_context = nodeContextPrompt;
			}

			const response = await fetch('/api/chat/message', {
				credentials: 'include',
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestBody),
				signal: abortController.signal,
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			// Get conversation ID from response header
			const newConvId = response.headers.get('X-Conversation-Id');
			if (newConvId && newConvId !== conversationId) {
				// New conversation was created, add to sidebar
				const isNewConversation = !conversationId;
				conversationId = newConvId;
				activeConversationId = newConvId;

				if (isNewConversation) {
					// Add new conversation to top of list
					conversations = [{
						id: newConvId,
						title: userMessage.slice(0, 50) + (userMessage.length > 50 ? '...' : ''),
						timestamp: new Date().toISOString(),
						pinned: false
					}, ...conversations];
				}
			}

			// Stream the response
			const reader = response.body?.getReader();
			const decoder = new TextDecoder();
			let fullContent = '';
			let artifactStarted = false;
			let artifactCompleted = false;
			let displayContent = ''; // Content to show in chat (without artifact JSON)
			let inArtifactBlock = false;

			if (reader) {
				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					const chunk = decoder.decode(value, { stream: true });
					fullContent += chunk;

					// Track if we're inside artifact block to filter it from display
					if (fullContent.includes('```artifact') && !artifactStarted) {
						artifactStarted = true;
						inArtifactBlock = true;
						generatingArtifact = true;
						artifactsPanelOpen = true;

						// Get content before artifact block
						const beforeArtifact = fullContent.split('```artifact')[0];
						displayContent = beforeArtifact;
					}

					// Check if artifact block has been closed
					if (inArtifactBlock) {
						const afterArtifactStart = fullContent.slice(fullContent.indexOf('```artifact'));
						const backtickMatches = afterArtifactStart.match(/```/g);
						if (backtickMatches && backtickMatches.length >= 2) {
							// Artifact block is complete
							inArtifactBlock = false;
							artifactCompleted = true;
							generatingArtifact = false;
							artifactCompletedInStream = true;

							// Get content after artifact block
							const artifactEndIndex = fullContent.indexOf('```artifact');
							const afterArtifact = fullContent.slice(artifactEndIndex);
							const closingIndex = afterArtifact.indexOf('```', afterArtifact.indexOf('\n'));
							const afterClosing = afterArtifact.slice(closingIndex + 3);
							displayContent = fullContent.split('```artifact')[0].trim();
							if (afterClosing.trim()) {
								displayContent += '\n\n' + afterClosing.trim();
							}

							// Parse artifact for viewing
							try {
								const artifactMatch = fullContent.match(/```artifact\s*\n([\s\S]*?)\n```/);
								if (artifactMatch) {
									const artifactData = JSON.parse(artifactMatch[1].trim());
									if (artifactData.title && artifactData.type && artifactData.content) {
										viewingArtifactFromMessage = {
											title: artifactData.title,
											type: artifactData.type,
											content: artifactData.content
												.replace(/\\n/g, '\n')
												.replace(/\\"/g, '"')
												.replace(/\\\\/g, '\\')
										};
										generatingArtifactTitle = artifactData.title;
										generatingArtifactType = artifactData.type;
									}
								}
							} catch {
								// Failed to parse
							}
						}
					}

					// Extract title/type for loading card
					if (artifactStarted && !artifactCompleted) {
						const titleMatch = fullContent.match(/"title":\s*"([^"]+)"/);
						if (titleMatch) generatingArtifactTitle = titleMatch[1];
						const typeMatch = fullContent.match(/"type":\s*"([^"]+)"/);
						if (typeMatch) generatingArtifactType = typeMatch[1];

						// Extract content for preview panel
						const contentMatch = fullContent.match(/"content":\s*"([\s\S]*?)(?:"\s*}|$)/);
						if (contentMatch) {
							generatingArtifactContent = contentMatch[1]
								.replace(/\\n/g, '\n')
								.replace(/\\"/g, '"')
								.replace(/\\\\/g, '\\');
						}
					}

					// Update message with filtered display content (without artifact JSON)
					if (artifactStarted) {
						// When artifact is being generated, show clean content with artifact reference
						const currentDisplayContent = inArtifactBlock ? displayContent : displayContent;
						messages = messages.map(msg =>
							msg.id === assistantMsgId
								? {
									...msg,
									content: currentDisplayContent,
									artifacts: artifactCompleted && viewingArtifactFromMessage ? [{
										title: viewingArtifactFromMessage.title,
										type: viewingArtifactFromMessage.type,
										content: viewingArtifactFromMessage.content
									}] : (inArtifactBlock ? [{
										title: generatingArtifactTitle || 'Creating artifact...',
										type: generatingArtifactType || 'document',
										content: '__generating__'
									}] : undefined)
								}
								: msg
						);
					} else {
						// No artifact - just update content normally
						messages = messages.map(msg =>
							msg.id === assistantMsgId
								? { ...msg, content: fullContent }
								: msg
						);
					}
				}
			}

			// Check if the response contains artifact blocks - final cleanup
			if (fullContent.includes('```artifact')) {
				// Artifact was created - refresh artifacts list
				await loadArtifacts();

				// If artifact is an actionable type, offer to create tasks inline
				if (viewingArtifactFromMessage) {
					const actionableTypes = ['plan', 'framework', 'proposal', 'sop'];
					if (actionableTypes.includes(viewingArtifactFromMessage.type.toLowerCase())) {
						// Trigger inline task creation prompt
						await triggerInlineTaskCreation(viewingArtifactFromMessage);
					}
				}
			}

			// Reset generation state after streaming completes
			generatingArtifact = false;
			generatingArtifactTitle = '';
			generatingArtifactType = '';
			generatingArtifactContent = '';
		} catch (error: any) {
			if (error.name === 'AbortError') {
				console.log('Request aborted');
			} else {
				console.error('Chat error:', error);
				// Update assistant message with error
				messages = messages.map(msg =>
					msg.id === assistantMsgId
						? { ...msg, content: 'Sorry, there was an error processing your request. Please try again.' }
						: msg
				);
			}
		} finally {
			isStreaming = false;
			abortController = null;
		}
	}

	// Parse artifact blocks from message content for rendering
	interface ParsedPart {
		type: 'text' | 'artifact';
		text?: string;
		artifact?: { title: string; type: string; content: string };
	}

	function parseMessageContent(content: string): ParsedPart[] {
		const parts: ParsedPart[] = [];
		// More flexible regex that matches artifact blocks with any field order
		// Match ```artifact followed by JSON block and closing ```
		const pattern = /```artifact\s*\n([\s\S]*?)\n```/g;
		let lastIndex = 0;
		let match;

		while ((match = pattern.exec(content)) !== null) {
			// Add text before the artifact block
			if (match.index > lastIndex) {
				const textBefore = content.slice(lastIndex, match.index).trim();
				if (textBefore) {
					parts.push({ type: 'text', text: textBefore });
				}
			}

			// Try to parse the JSON inside the artifact block
			try {
				const jsonStr = match[1].trim();
				const artifactData = JSON.parse(jsonStr);

				if (artifactData.title && artifactData.type && artifactData.content) {
					// Unescape content if needed
					const artifactContent = artifactData.content
						.replace(/\\n/g, '\n')
						.replace(/\\"/g, '"')
						.replace(/\\\\/g, '\\');

					parts.push({
						type: 'artifact',
						artifact: {
							title: artifactData.title,
							type: artifactData.type,
							content: artifactContent
						}
					});
				}
			} catch {
				// JSON parsing failed - this might be incomplete, skip it
				console.log('Failed to parse artifact JSON, possibly incomplete');
			}

			lastIndex = match.index + match[0].length;
		}

		// Add remaining text
		if (lastIndex < content.length) {
			const remainingText = content.slice(lastIndex).trim();
			if (remainingText) {
				// Check if remaining text contains an incomplete artifact block
				if (remainingText.includes('```artifact') && !remainingText.includes('```artifact') ||
					(remainingText.includes('```artifact') && remainingText.lastIndexOf('```') === remainingText.indexOf('```artifact') + 3)) {
					// Incomplete artifact block - don't show it
					const beforeArtifact = remainingText.split('```artifact')[0].trim();
					if (beforeArtifact) {
						parts.push({ type: 'text', text: beforeArtifact });
					}
				} else {
					parts.push({ type: 'text', text: remainingText });
				}
			}
		}

		// If no parts found, check if we're in the middle of generating an artifact
		if (parts.length === 0) {
			// Check if content contains an incomplete artifact block (started but not finished)
			if (content.includes('```artifact')) {
				// Extract text before the artifact block
				const beforeArtifact = content.split('```artifact')[0].trim();
				if (beforeArtifact) {
					return [{ type: 'text', text: beforeArtifact }];
				}
				// Nothing to show yet - artifact is being generated
				return [];
			}
			return [{ type: 'text', text: content }];
		}

		return parts;
	}

	// Simple markdown renderer
	function renderMarkdown(text: string): string {
		return text
			// Headers
			.replace(/^### (.+)$/gm, '<h3 class="text-lg font-semibold text-gray-900 mt-4 mb-2">$1</h3>')
			.replace(/^## (.+)$/gm, '<h2 class="text-xl font-semibold text-gray-900 mt-5 mb-3">$1</h2>')
			.replace(/^# (.+)$/gm, '<h1 class="text-2xl font-bold text-gray-900 mt-6 mb-4">$1</h1>')
			// Bold
			.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold">$1</strong>')
			// Italic
			.replace(/\*(.+?)\*/g, '<em class="italic">$1</em>')
			// Lists
			.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-gray-700">$1</li>')
			.replace(/^(\d+)\. (.+)$/gm, '<li class="ml-4 list-decimal text-gray-700">$2</li>')
			// Line breaks
			.replace(/\n\n/g, '</p><p class="mb-3">')
			.replace(/\n/g, '<br/>');
	}

	// Resize handlers
	function startResize(e: MouseEvent) {
		isResizing = true;
		resizeStartX = e.clientX;
		resizeStartWidth = artifactPanelWidth;
		document.addEventListener('mousemove', handleResize);
		document.addEventListener('mouseup', stopResize);
		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';
	}

	function handleResize(e: MouseEvent) {
		if (!isResizing) return;
		const delta = resizeStartX - e.clientX;
		const newWidth = Math.min(Math.max(resizeStartWidth + delta, 300), 800);
		artifactPanelWidth = newWidth;
	}

	function stopResize() {
		isResizing = false;
		document.removeEventListener('mousemove', handleResize);
		document.removeEventListener('mouseup', stopResize);
		document.body.style.cursor = '';
		document.body.style.userSelect = '';
	}

	function viewArtifactInPanel(artifact: { title: string; type: string; content: string }) {
		viewingArtifactFromMessage = artifact;
		selectedArtifact = null;
		artifactsPanelOpen = true;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSendMessage();
		}
	}

	function handleInput() {
		if (inputRef) {
			inputRef.style.height = 'auto';
			inputRef.style.height = Math.min(inputRef.scrollHeight, 200) + 'px';
		}
	}

	function handleStop() {
		if (abortController) {
			abortController.abort();
		}
	}

	function copyMessage(content: string, id: string) {
		navigator.clipboard.writeText(content);
		copiedMessageId = id;
		setTimeout(() => copiedMessageId = null, 2000);
	}

	function formatTime(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diffHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));
		if (diffHours < 1) return 'Just now';
		if (diffHours < 24) return `${diffHours}h ago`;
		return date.toLocaleDateString();
	}
</script>

<!-- Fixed height container that fills parent -->
<div class="h-full flex overflow-hidden">
	<!-- Chat Conversations Sidebar -->
	{#if chatSidebarOpen}
		<div class="w-64 h-full flex flex-col bg-white border-r border-gray-200 flex-shrink-0" transition:fly={{ x: -256, duration: 200 }}>
			<!-- Header -->
			<div class="p-4 flex-shrink-0">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Chats</h2>
					<button
						onclick={handleNewChat}
						class="w-8 h-8 flex items-center justify-center bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</button>
				</div>

				<!-- Search -->
				<div class="relative">
					<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search conversations..."
						bind:value={searchQuery}
						class="w-full pl-10 pr-4 py-2 text-sm bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
					/>
				</div>

				<!-- Filter Tabs -->
				<div class="flex items-center gap-1 mt-3">
					<button
						onclick={() => filterTab = 'all'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'all' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						All
					</button>
					<button
						onclick={() => filterTab = 'pinned'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'pinned' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						Pinned
					</button>
					<button
						onclick={() => filterTab = 'recent'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'recent' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						Recent
					</button>
				</div>
			</div>

			<!-- Conversation List - scrollable -->
			<div class="flex-1 overflow-y-auto px-2">
				{#each conversations as conv (conv.id)}
					<button
						onclick={() => selectConversation(conv.id)}
						class="w-full text-left p-3 rounded-lg mb-1 transition-colors {activeConversationId === conv.id ? 'bg-gray-100' : 'hover:bg-gray-50'}"
					>
						<p class="text-sm font-medium text-gray-900 truncate">{conv.title}</p>
						<p class="text-xs text-gray-500 mt-1">{formatTime(conv.timestamp)}</p>
					</button>
				{/each}
			</div>

			<!-- Footer -->
			<div class="p-3 flex-shrink-0 border-t border-gray-100">
				<button class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
					</svg>
					View archived
				</button>
			</div>
		</div>
	{/if}

	<!-- Main Chat Area - fills remaining space -->
	<div class="flex-1 flex flex-col min-w-0 h-full bg-gray-50">
		<!-- Toggle button - fixed header -->
		<div class="h-12 flex items-center justify-between px-4 flex-shrink-0 border-b border-gray-100 min-w-0">
			<button
				onclick={() => chatSidebarOpen = !chatSidebarOpen}
				class="p-2 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition-colors flex-shrink-0"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					{#if chatSidebarOpen}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					{:else}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					{/if}
				</svg>
			</button>

			<div class="flex items-center gap-2 min-w-0">
				<!-- Project Selector (required for chat) -->
				<div class="relative flex-shrink-0">
					<button
						onclick={() => { showProjectDropdown = !showProjectDropdown; showHeaderContextDropdown = false; showNodeDropdown = false; }}
						class="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors {selectedProject ? 'bg-purple-50 text-purple-700 hover:bg-purple-100' : 'bg-amber-50 text-amber-700 hover:bg-amber-100 border border-amber-200'}"
						title={selectedProject ? selectedProject.name : 'Select Project'}
					>
						<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
						</svg>
						<span>{selectedProject ? selectedProject.name : 'Select Project'}</span>
						{#if !selectedProject}
							<span class="text-[10px] flex-shrink-0">!</span>
						{/if}
						<svg class="w-3 h-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					{#if showProjectDropdown}
						<div
							class="absolute left-0 top-full mt-2 w-72 bg-white border border-gray-200 rounded-xl shadow-lg py-2 z-20 max-h-80 overflow-y-auto"
							transition:fly={{ y: -10, duration: 200 }}
						>
							<div class="px-3 py-1.5">
								<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Select Project</span>
							</div>
							{#if loadingProjects}
								<div class="px-4 py-3 text-sm text-gray-500">Loading projects...</div>
							{:else if projectsList.length === 0}
								<div class="px-4 py-6 text-center">
									<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
									</svg>
									<p class="text-sm text-gray-500">No projects yet</p>
									<a href="/projects" class="text-sm text-blue-600 hover:underline">Create a project</a>
								</div>
							{:else}
								{#each projectsList as project (project.id)}
									{@const isSelected = selectedProjectId === project.id}
									<button
										onclick={() => { selectedProjectId = project.id; showProjectDropdown = false; }}
										class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors flex items-center gap-3 {isSelected ? 'bg-purple-50' : ''}"
									>
										<div class="w-8 h-8 rounded-lg {isSelected ? 'bg-purple-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
											{#if isSelected}
												<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else}
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
												</svg>
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<div class="text-sm font-medium {isSelected ? 'text-purple-600' : 'text-gray-700'} truncate">{project.name}</div>
											{#if project.description}
												<div class="text-xs text-gray-500 truncate">{project.description}</div>
											{/if}
										</div>
									</button>
								{/each}
							{/if}
						</div>
					{/if}
				</div>

				<!-- Context Profile Selector (header) -->
				<div class="relative flex-shrink-0">
					<button
						onclick={() => { showHeaderContextDropdown = !showHeaderContextDropdown; showProjectDropdown = false; showNodeDropdown = false; }}
						class="flex items-center gap-1.5 px-2.5 py-1.5 text-sm bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 transition-colors"
					>
						<svg class="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
							<path d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
						<span>{selectedContextsLabel}</span>
						<svg class="w-3 h-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					{#if showHeaderContextDropdown}
						<div
							class="absolute left-0 top-full mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg py-2 z-20 max-h-80 overflow-y-auto"
							transition:fly={{ y: -10, duration: 200 }}
						>
							<div class="px-3 py-1.5">
								<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Available Contexts</span>
							</div>
							{#if loadingContexts}
								<div class="px-4 py-3 text-sm text-gray-500">Loading contexts...</div>
							{:else if availableContexts.length === 0}
								<div class="px-4 py-3 text-sm text-gray-500">No contexts available</div>
							{:else}
								<!-- Clear all option -->
								{#if selectedContextIds.length > 0}
									<button
										onclick={() => { selectedContextIds = []; }}
										class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors flex items-center gap-3 border-b border-gray-100"
									>
										<div class="w-8 h-8 rounded-lg bg-gray-100 text-gray-500 flex items-center justify-center flex-shrink-0">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<div class="text-sm font-medium text-gray-700">Clear selection</div>
											<div class="text-xs text-gray-500">{selectedContextIds.length} selected</div>
										</div>
									</button>
								{/if}
								{#each availableContexts as ctx (ctx.id)}
									{@const isSelected = selectedContextIds.includes(ctx.id)}
									<button
										onclick={() => {
											if (isSelected) {
												selectedContextIds = selectedContextIds.filter(id => id !== ctx.id);
											} else {
												selectedContextIds = [...selectedContextIds, ctx.id];
											}
										}}
										class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors flex items-center gap-3 {isSelected ? 'bg-blue-50' : ''}"
									>
										<div class="w-8 h-8 rounded-lg {isSelected ? 'bg-blue-500 text-white' : 'bg-blue-100 text-blue-600'} flex items-center justify-center flex-shrink-0 text-lg">
											{#if isSelected}
												<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else}
												{ctx.icon || '📄'}
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<div class="text-sm font-medium {isSelected ? 'text-blue-600' : 'text-gray-700'} truncate">{ctx.name}</div>
											{#if ctx.type}
												<div class="text-xs text-gray-500 capitalize">{ctx.type}</div>
											{/if}
										</div>
									</button>
								{/each}
								{#if selectedContextIds.length > 0}
									<div class="px-4 py-2 border-t border-gray-100">
										<button
											onclick={() => showHeaderContextDropdown = false}
											class="w-full py-1.5 text-sm font-medium text-blue-600 hover:text-blue-700"
										>
											Done
										</button>
									</div>
								{/if}
							{/if}
						</div>
					{/if}
				</div>

				<!-- Active Node Indicator -->
				{#if activeNode}
					<div class="relative flex-shrink-0">
						<button
							onclick={() => showNodeDropdown = !showNodeDropdown}
							class="flex items-center gap-1.5 px-2.5 py-1.5 text-sm bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 transition-colors"
						>
							<svg class="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
								<path d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							<span>{activeNode.name}</span>
							<svg class="w-3 h-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</button>

						{#if showNodeDropdown}
							<div
								class="absolute right-0 top-full mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg p-3 z-20"
								transition:fly={{ y: -10, duration: 200 }}
							>
								<div class="text-xs font-semibold text-gray-500 uppercase mb-2">Active Node</div>
								<div class="mb-3">
									<p class="text-sm font-medium text-gray-900">{activeNode.name}</p>
									{#if activeNode.purpose}
										<p class="text-xs text-gray-500 mt-1 line-clamp-2">{activeNode.purpose}</p>
									{/if}
								</div>
								<div class="flex gap-2">
									<a
										href="/nodes/{activeNode.id}"
										class="flex-1 text-center px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										View
									</a>
									<button
										onclick={handleDeactivateNode}
										class="flex-1 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
									>
										Deactivate
									</button>
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<a
						href="/nodes"
						class="flex items-center gap-1.5 px-2.5 py-1.5 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors flex-shrink-0"
					>
						<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
						<span>No Node</span>
					</a>
				{/if}

				<!-- Artifacts Toggle -->
				<button
					onclick={() => artifactsPanelOpen = !artifactsPanelOpen}
					class="flex items-center gap-1.5 px-2.5 py-1.5 text-sm rounded-lg transition-colors flex-shrink-0 {artifactsPanelOpen ? 'bg-blue-100 text-blue-700' : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'}"
				>
					<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					<span>Artifacts</span>
					{#if artifacts.length > 0}
						<span class="px-1.5 py-0.5 text-xs font-medium rounded-full {artifactsPanelOpen ? 'bg-blue-200' : 'bg-gray-200'}">{artifacts.length}</span>
					{/if}
				</button>
			</div>
		</div>

		{#if hasConversation}
			<!-- Messages container - scrollable, takes remaining height -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto min-h-0">
				<div class="max-w-3xl mx-auto px-2 sm:px-4 py-4 sm:py-6 space-y-4 sm:space-y-6">
					{#if loadingConversation}
						<div class="flex items-center justify-center py-12">
							<div class="flex items-center gap-3 text-gray-500">
								<svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								<span class="text-sm">Loading conversation...</span>
							</div>
						</div>
					{/if}
					{#each messages as message, i (message.id)}
						{@const isLastMessage = i === messages.length - 1}
						{@const parsedParts = parseMessageContent(message.content)}

						{#if message.role === 'user'}
							<!-- User message - dark bubble on right -->
							<div class="flex justify-end">
								<div class="max-w-[90%] sm:max-w-[80%] bg-gray-900 text-white px-3 sm:px-4 py-2.5 sm:py-3 rounded-2xl rounded-br-md">
									<p class="text-sm sm:text-[15px] leading-relaxed whitespace-pre-wrap break-words">{message.content}</p>
								</div>
							</div>
						{:else if message.role === 'assistant'}
							<!-- Assistant message - left aligned -->
							<div class="max-w-[95%] sm:max-w-[85%]">
								{#if !message.content && !message.artifacts?.length && isStreaming && isLastMessage}
									<!-- Still loading, show initial indicator -->
									<div class="flex items-center gap-2 text-sm text-gray-500">
										<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										<span>Thinking...</span>
									</div>
								{:else}
									<!-- Show text content if any -->
									{#if message.content}
										<p class="text-sm sm:text-[15px] leading-relaxed text-gray-800 whitespace-pre-wrap break-words">{message.content}</p>
									{/if}

									<!-- Show artifacts from message.artifacts (new approach) -->
									{#if message.artifacts?.length}
										{#each message.artifacts as artifact}
											{#if artifact.content === '__generating__'}
												<!-- Artifact is being generated - show loading card -->
												<div class="my-3 flex items-center gap-3 px-4 py-3 bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-xl animate-pulse">
													<div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center flex-shrink-0">
														<svg class="w-5 h-5 text-blue-600 animate-spin" fill="none" viewBox="0 0 24 24">
															<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
															<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
														</svg>
													</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 truncate">{artifact.title}</p>
														<p class="text-xs text-gray-500 capitalize">{artifact.type} &bull; Creating...</p>
													</div>
													<div class="h-2 w-16 bg-blue-200 rounded-full overflow-hidden">
														<div class="h-full bg-blue-500 rounded-full animate-pulse" style="width: 60%"></div>
													</div>
												</div>
											{:else}
												<!-- Completed artifact card -->
												<div class="my-3">
													<button
														onclick={() => viewArtifactInPanel(artifact)}
														class="flex items-center gap-3 px-4 py-3 bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-t-xl hover:shadow-md hover:border-blue-300 transition-all cursor-pointer w-full text-left group"
													>
														<div class="w-10 h-10 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
															<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getArtifactIcon(artifact.type)} />
															</svg>
														</div>
														<div class="flex-1 min-w-0">
															<p class="text-sm font-medium text-gray-900 truncate">{artifact.title}</p>
															<p class="text-xs text-gray-500 capitalize">{artifact.type} &bull; Click to view</p>
														</div>
														<svg class="w-5 h-5 text-gray-400 group-hover:text-blue-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
														</svg>
													</button>
													<!-- Action buttons for artifact -->
													<div class="flex items-center gap-2 px-3 py-2 bg-gray-50 border border-t-0 border-gray-200 rounded-b-xl">
														<button
															onclick={() => generateTasksFromArtifact(artifact)}
															class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-green-700 bg-green-50 hover:bg-green-100 rounded-lg transition-colors"
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
															</svg>
															Generate Tasks
														</button>
														<button
															onclick={() => viewArtifactInPanel(artifact)}
															class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
															</svg>
															View
														</button>
														<button
															onclick={() => { viewingArtifactFromMessage = artifact; openSaveToProfileModal(); }}
															class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
														>
															<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
															</svg>
															Save to Profile
														</button>
													</div>
												</div>
											{/if}
										{/each}
									{/if}

									<!-- Fallback: Show parsed artifacts from content (legacy behavior) -->
									{#if !message.artifacts?.length}
										{#each parsedParts as part}
											{#if part.type === 'artifact' && part.artifact}
												<button
													onclick={() => viewArtifactInPanel(part.artifact!)}
													class="my-3 flex items-center gap-3 px-4 py-3 bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-xl hover:shadow-md hover:border-blue-300 transition-all cursor-pointer w-full text-left group"
												>
													<div class="w-10 h-10 rounded-lg {getArtifactColor(part.artifact.type)} flex items-center justify-center flex-shrink-0">
														<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getArtifactIcon(part.artifact.type)} />
														</svg>
													</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 truncate">{part.artifact.title}</p>
														<p class="text-xs text-gray-500 capitalize">{part.artifact.type} &bull; Click to view</p>
													</div>
													<svg class="w-5 h-5 text-gray-400 group-hover:text-blue-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
													</svg>
												</button>
											{:else if part.type === 'text' && part.text && !message.content}
												<p class="text-[15px] leading-relaxed text-gray-800 whitespace-pre-wrap">{part.text}</p>
											{/if}
										{/each}
									{/if}
								{/if}
								{#if isLastMessage && isStreaming && (message.content || message.artifacts?.length) && !artifactCompletedInStream}<span class="inline-block w-2 h-5 bg-blue-500 animate-pulse ml-1 rounded-sm"></span>{/if}

								<!-- Inline Task Creation (after artifact) -->
								{#if isLastMessage && showInlineTaskCreation}
									<div class="my-4 p-4 bg-gradient-to-br from-green-50 to-emerald-50 border border-green-200 rounded-xl">
										<div class="flex items-center justify-between mb-3">
											<div class="flex items-center gap-2">
												<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
												</svg>
												<h4 class="font-medium text-gray-900">Create Tasks from Artifact?</h4>
											</div>
											<button
												onclick={dismissInlineTasks}
												class="p-1 text-gray-400 hover:text-gray-600 rounded"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											</button>
										</div>

										{#if creatingInlineTasks}
											<div class="flex items-center gap-2 py-4 justify-center">
												<svg class="w-5 h-5 animate-spin text-green-600" fill="none" viewBox="0 0 24 24">
													<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
													<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
												</svg>
												<span class="text-sm text-gray-600">Analyzing artifact and generating tasks...</span>
											</div>
										{:else if inlineTasksForArtifact.length === 0}
											<p class="text-sm text-gray-500 text-center py-3">No actionable tasks found in this artifact.</p>
											<button
												onclick={dismissInlineTasks}
												class="w-full mt-2 px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
											>
												Dismiss
											</button>
										{:else}
											<div class="space-y-2 mb-4 max-h-64 overflow-y-auto">
												{#each inlineTasksForArtifact as task, i}
													<div class="flex items-start gap-3 p-3 bg-white rounded-lg border border-gray-200">
														<div class="flex-1 min-w-0">
															<p class="text-sm font-medium text-gray-900">{task.title}</p>
															{#if task.description}
																<p class="text-xs text-gray-500 mt-0.5 line-clamp-2">{task.description}</p>
															{/if}
															<div class="flex items-center gap-2 mt-2">
																<span class="px-2 py-0.5 text-xs rounded-full {task.priority === 'high' ? 'bg-red-100 text-red-700' : task.priority === 'medium' ? 'bg-yellow-100 text-yellow-700' : 'bg-gray-100 text-gray-700'}">
																	{task.priority}
																</span>
																<select
																	value={task.assignee_id || ''}
																	onchange={(e) => updateInlineTaskAssignee(i, (e.target as HTMLSelectElement).value)}
																	class="text-xs border border-gray-200 rounded px-2 py-1 bg-white"
																>
																	<option value="">Unassigned</option>
																	{#each availableTeamMembers as member}
																		<option value={member.id}>{member.name} ({member.role})</option>
																	{/each}
																</select>
															</div>
														</div>
														<button
															onclick={() => removeInlineTask(i)}
															class="p-1 text-gray-400 hover:text-red-500 rounded"
														>
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
															</svg>
														</button>
													</div>
												{/each}
											</div>

											<div class="flex gap-2">
												<button
													onclick={dismissInlineTasks}
													class="flex-1 px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
												>
													Skip
												</button>
												<button
													onclick={confirmInlineTasks}
													disabled={creatingInlineTasks}
													class="flex-1 px-4 py-2 text-sm text-white bg-green-600 hover:bg-green-700 rounded-lg transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
												>
													{#if creatingInlineTasks}
														<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
															<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
															<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
														</svg>
														Creating...
													{:else}
														Create {inlineTasksForArtifact.length} Task{inlineTasksForArtifact.length > 1 ? 's' : ''}
													{/if}
												</button>
											</div>
										{/if}
									</div>
								{/if}

								{#if (message.content || message.artifacts?.length || parsedParts.length > 0) && (!isStreaming || !isLastMessage || artifactCompletedInStream)}
									<div class="flex items-center gap-2 mt-3">
										<button
											onclick={() => copyMessage(message.content, message.id)}
											class="flex items-center gap-1.5 px-2.5 py-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
										>
											{#if copiedMessageId === message.id}
												<svg class="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
												<span class="text-green-600">Copied</span>
											{:else}
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
												</svg>
												<span>Copy</span>
											{/if}
										</button>
									</div>
								{/if}
							</div>
						{/if}
					{/each}

					{#if isStreaming && messages[messages.length - 1]?.role === 'user'}
						<!-- Typing indicator -->
						<div class="flex items-center gap-1.5">
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms"></div>
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms"></div>
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms"></div>
						</div>
					{/if}
				</div>
			</div>

			<!-- Input Area - fixed at bottom -->
			<div class="flex-shrink-0 p-4 bg-gray-50 border-t border-gray-100">
				<div class="max-w-3xl mx-auto">
					<div class="bg-white rounded-3xl shadow-sm border border-gray-200 p-4 cursor-text" onclick={() => inputRef?.focus()}>
						<!-- Textarea -->
						<textarea
							bind:this={inputRef}
							bind:value={inputValue}
							placeholder="Ask OSA anything..."
							rows={1}
							disabled={isStreaming}
							class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
							style="min-height: 24px; max-height: 200px;"
							onkeydown={handleKeydown}
							oninput={handleInput}
						></textarea>

						<!-- Bottom row -->
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<!-- Plus button -->
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Add">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
								</button>

								<!-- Attachment -->
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Attach file">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
								</button>

								<!-- Context selector -->
								<div class="relative">
									<button
										onclick={() => { showContextDropdown = !showContextDropdown; showModelDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										{selectedContexts.length > 0 ? selectedContextsLabel : 'Default'}
									</button>

									{#if showContextDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[220px] z-10 max-h-64 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											{#if selectedContextIds.length > 0}
												<button
													onclick={() => { selectedContextIds = []; }}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-600 border-b border-gray-100"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
													</svg>
													Clear ({selectedContextIds.length})
												</button>
											{/if}
											{#each availableContexts as ctx (ctx.id)}
												{@const isSelected = selectedContextIds.includes(ctx.id)}
												<button
													onclick={() => {
														if (isSelected) {
															selectedContextIds = selectedContextIds.filter(id => id !== ctx.id);
														} else {
															selectedContextIds = [...selectedContextIds, ctx.id];
														}
													}}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 {isSelected ? 'text-blue-600 font-medium bg-blue-50' : 'text-gray-600'}"
												>
													{#if isSelected}
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
														</svg>
													{:else}
														<span class="text-base">{ctx.icon || '📄'}</span>
													{/if}
													<span class="truncate">{ctx.name}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<!-- Model selector -->
								<div class="relative">
									<button
										onclick={() => { showModelDropdown = !showModelDropdown; showContextDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
										</svg>
										{currentModelName}
									</button>

									{#if showModelDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-2 min-w-[250px] z-10 max-h-80 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<!-- Cloud Models -->
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Cloud</span>
											</div>
											{#each cloudModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}

											<div class="border-t border-gray-100 my-2"></div>

											<!-- Local Models -->
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Local</span>
											</div>
											{#each localModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>

							<!-- Send/Stop button -->
							{#if isStreaming}
								<button
									type="button"
									onclick={handleStop}
									class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors"
								>
									<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
										<rect x="6" y="6" width="12" height="12" rx="2" />
									</svg>
								</button>
							{:else}
								<button
									type="button"
									onclick={handleSendMessage}
									disabled={!inputValue.trim()}
									class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
									</svg>
								</button>
							{/if}
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Empty State - centered in available space -->
			<div class="flex-1 flex items-center justify-center overflow-auto">
				<div class="w-full max-w-3xl px-6">
					<!-- Personalized Title -->
					<div class="text-center mb-8">
						<h1 class="text-3xl font-semibold text-gray-900 mb-3">
							{personalizedGreeting}
						</h1>
						<p class="text-gray-500 h-6">
							Let me help you <span class="text-blue-600 font-medium">{displayedSuggestion}</span><span class="cursor-blink text-blue-600 font-light">|</span>
						</p>
					</div>

					<!-- Input Box -->
					<div class="bg-white rounded-3xl shadow-lg border border-gray-200 p-4 cursor-text" onclick={() => inputRef?.focus()}>
						<textarea
							bind:this={inputRef}
							bind:value={inputValue}
							placeholder="Ask OSA anything..."
							rows={1}
							disabled={isStreaming}
							class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
							style="min-height: 24px; max-height: 200px;"
							onkeydown={handleKeydown}
							oninput={handleInput}
						></textarea>

						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Add">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
								</button>

								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Attach file">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
								</button>

								<div class="relative">
									<button
										onclick={() => { showContextDropdown = !showContextDropdown; showModelDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										{selectedContexts.length > 0 ? selectedContextsLabel : 'Default'}
									</button>

									{#if showContextDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[220px] z-10 max-h-64 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											{#if selectedContextIds.length > 0}
												<button
													onclick={() => { selectedContextIds = []; }}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-600 border-b border-gray-100"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
													</svg>
													Clear ({selectedContextIds.length})
												</button>
											{/if}
											{#each availableContexts as ctx (ctx.id)}
												{@const isSelected = selectedContextIds.includes(ctx.id)}
												<button
													onclick={() => {
														if (isSelected) {
															selectedContextIds = selectedContextIds.filter(id => id !== ctx.id);
														} else {
															selectedContextIds = [...selectedContextIds, ctx.id];
														}
													}}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 {isSelected ? 'text-blue-600 font-medium bg-blue-50' : 'text-gray-600'}"
												>
													{#if isSelected}
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
														</svg>
													{:else}
														<span class="text-base">{ctx.icon || '📄'}</span>
													{/if}
													<span class="truncate">{ctx.name}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<div class="relative">
									<button
										onclick={() => { showModelDropdown = !showModelDropdown; showContextDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
										</svg>
										{currentModelName}
									</button>

									{#if showModelDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-2 min-w-[250px] z-10 max-h-80 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Cloud</span>
											</div>
											{#each cloudModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}

											<div class="border-t border-gray-100 my-2"></div>

											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Local</span>
											</div>
											{#each localModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>

							<button
								type="button"
								onclick={handleSendMessage}
								disabled={!inputValue.trim() || isStreaming}
								class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
								</svg>
							</button>
						</div>
					</div>

					<!-- Quick Actions -->
					<div class="flex flex-wrap justify-center gap-2 mt-5">
						{#each quickActions as action}
							<button
								onclick={() => handleQuickAction(action)}
								class="px-4 py-2 bg-white border border-gray-200 rounded-full text-sm text-gray-600 hover:bg-gray-50 hover:border-gray-300 transition-all"
							>
								{action}
							</button>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Resizable Divider + Artifacts Panel -->
	{#if artifactsPanelOpen}
		<!-- Resize Handle -->
		<div
			class="w-1 h-full bg-gray-200 hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors relative group"
			onmousedown={startResize}
			role="separator"
			aria-orientation="vertical"
		>
			<div class="absolute inset-y-0 -left-1 -right-1 group-hover:bg-blue-400/20"></div>
			<!-- Visible grip indicator -->
			<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex flex-col gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
			</div>
		</div>

		<div class="h-full flex flex-col bg-white flex-shrink-0" style="width: {artifactPanelWidth}px" transition:fly={{ x: 320, duration: 200 }}>
			<!-- Panel Header -->
			<div class="p-4 border-b border-gray-100 flex-shrink-0">
				<div class="flex items-center justify-between mb-3">
					<h3 class="font-semibold text-gray-900">Artifacts</h3>
					<button
						onclick={() => { artifactsPanelOpen = false; viewingArtifactFromMessage = null; }}
						class="p-1 text-gray-400 hover:text-gray-600 rounded"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Filter Tabs (only show when not viewing message artifact) -->
				{#if !viewingArtifactFromMessage}
					<div class="flex gap-1 overflow-x-auto">
						{#each ['all', 'proposal', 'sop', 'framework', 'plan', 'report'] as filter}
							<button
								onclick={() => { artifactFilter = filter; loadArtifacts(); }}
								class="px-2.5 py-1 text-xs font-medium rounded-lg whitespace-nowrap transition-colors {artifactFilter === filter ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
							>
								{filter === 'all' ? 'All' : filter.charAt(0).toUpperCase() + filter.slice(1)}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Content Area: Generating | Message Artifact | Selected Artifact | List -->
			{#if generatingArtifact}
				<!-- Live Generation View -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Generation Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg {generatingArtifactType ? getArtifactColor(generatingArtifactType) : 'bg-blue-50 text-blue-500'} flex items-center justify-center flex-shrink-0 relative">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={generatingArtifactType ? getArtifactIcon(generatingArtifactType) : 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'} />
								</svg>
								<!-- Generating indicator -->
								<div class="absolute -top-1 -right-1 w-3 h-3">
									<span class="absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75 animate-ping"></span>
									<span class="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
								</div>
							</div>
							<div class="min-w-0 flex-1">
								<h4 class="font-medium text-gray-900 truncate">
									{generatingArtifactTitle || 'Generating artifact...'}
								</h4>
								<p class="text-xs text-gray-500 flex items-center gap-1.5">
									{#if generatingArtifactType}
										<span class="capitalize">{generatingArtifactType}</span>
										<span>&bull;</span>
									{/if}
									<span class="flex items-center gap-1">
										<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										Writing...
									</span>
								</p>
							</div>
						</div>
					</div>

					<!-- Live Content Preview with Markdown -->
					<div class="flex-1 overflow-y-auto p-4 bg-gray-50">
						<div class="prose prose-sm max-w-none">
							{@html renderMarkdown(generatingArtifactContent || 'Waiting for content...')}
							<span class="inline-block w-2 h-4 bg-blue-500 animate-pulse ml-0.5"></span>
						</div>
					</div>
				</div>
			{:else if viewingArtifactFromMessage}
				<!-- Viewing artifact from message -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<div class="flex items-center justify-between mb-2">
							<button
								onclick={() => { viewingArtifactFromMessage = null; isEditingArtifact = false; }}
								class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
								</svg>
								Back
							</button>
							{#if !isEditingArtifact}
								<button
									onclick={startEditingArtifact}
									class="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
									</svg>
									Edit
								</button>
							{/if}
						</div>
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(viewingArtifactFromMessage.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(viewingArtifactFromMessage.type)} />
								</svg>
							</div>
							<div class="min-w-0">
								<h4 class="font-medium text-gray-900">{viewingArtifactFromMessage.title}</h4>
								<p class="text-xs text-gray-500 capitalize">{viewingArtifactFromMessage.type}</p>
							</div>
						</div>
					</div>

					<!-- Content - Editable or Rendered -->
					<div class="flex-1 overflow-y-auto p-4">
						{#if isEditingArtifact}
							<textarea
								bind:value={editedArtifactContent}
								class="w-full h-full min-h-[300px] p-3 text-sm font-mono text-gray-700 bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
								placeholder="Edit artifact content..."
							></textarea>
						{:else}
							<div
								class="prose prose-sm max-w-none cursor-text hover:bg-gray-50 rounded-lg p-2 -m-2 transition-colors"
								onclick={startEditingArtifact}
								role="button"
								tabindex="0"
								onkeydown={(e) => e.key === 'Enter' && startEditingArtifact()}
							>
								{@html renderMarkdown(viewingArtifactFromMessage.content)}
							</div>
						{/if}
					</div>

					<!-- Actions -->
					<div class="p-3 border-t border-gray-100 flex-shrink-0">
						{#if isEditingArtifact}
							<div class="flex gap-2">
								<button
									onclick={cancelArtifactEdit}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
								>
									Cancel
								</button>
								<button
									onclick={saveArtifactEdit}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-blue-500 hover:bg-blue-600 rounded-lg transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Save Changes
								</button>
							</div>
						{:else}
							<div class="flex gap-2 mb-2">
								<button
									onclick={() => { navigator.clipboard.writeText(viewingArtifactFromMessage?.content || ''); }}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
									Copy
								</button>
								<button class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
									</svg>
									Export
								</button>
							</div>
							<button
								onclick={openSaveToNodeModal}
								class="w-full flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
								</svg>
								Save to Profile
							</button>
						{/if}
					</div>
				</div>
			{:else if selectedArtifact}
				<!-- Artifact Detail View (from API) -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Detail Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<button
							onclick={closeArtifactDetail}
							class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 mb-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
							Back
						</button>
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(selectedArtifact.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(selectedArtifact.type)} />
								</svg>
							</div>
							<div class="min-w-0">
								<h4 class="font-medium text-gray-900 truncate">{selectedArtifact.title}</h4>
								<p class="text-xs text-gray-500 capitalize">{selectedArtifact.type} &bull; v{selectedArtifact.version}</p>
							</div>
						</div>
					</div>

					<!-- Content with Markdown -->
					<div class="flex-1 overflow-y-auto p-4">
						<div class="prose prose-sm max-w-none">
							{@html renderMarkdown(selectedArtifact.content)}
						</div>
					</div>

					<!-- Actions -->
					<div class="p-3 border-t border-gray-100 flex gap-2 flex-shrink-0">
						<button
							onclick={() => { navigator.clipboard.writeText(selectedArtifact?.content || ''); }}
							class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							Copy
						</button>
						<button class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
							</svg>
							Export
						</button>
					</div>
				</div>
			{:else}
				<!-- Artifacts List -->
				<div class="flex-1 overflow-y-auto">
					{#if loadingArtifacts}
						<div class="flex items-center justify-center h-32">
							<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
						</div>
					{:else if artifacts.length === 0}
						<div class="flex flex-col items-center justify-center h-48 text-center px-4">
							<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
								<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
							</div>
							<p class="text-sm text-gray-500">No artifacts yet</p>
							<p class="text-xs text-gray-400 mt-1">Ask OSA to create proposals, SOPs, or frameworks</p>
						</div>
					{:else}
						<div class="p-2 space-y-1">
							{#each artifacts as artifact (artifact.id)}
								<button
									onclick={() => selectArtifact(artifact.id)}
									class="w-full flex items-start gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors text-left"
								>
									<div class="w-9 h-9 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(artifact.type)} />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 truncate">{artifact.title}</p>
										{#if artifact.summary}
											<p class="text-xs text-gray-500 line-clamp-2 mt-0.5">{artifact.summary}</p>
										{/if}
										<div class="flex items-center gap-1.5 mt-1">
											<span class="text-xs text-gray-400 capitalize">{artifact.type}</span>
											{#if artifact.context_name}
												<span class="text-xs text-gray-300">&bull;</span>
												<span class="text-xs text-blue-500 truncate">{artifact.context_name}</span>
											{:else}
												<span class="text-xs text-gray-300">&bull;</span>
												<span class="text-xs text-gray-400 italic">Unlinked</span>
											{/if}
										</div>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Click outside to close dropdowns -->
{#if showContextDropdown || showModelDropdown || showNodeDropdown || showHeaderContextDropdown}
	<button
		class="fixed inset-0 z-[5] cursor-default"
		onclick={() => { showContextDropdown = false; showModelDropdown = false; showNodeDropdown = false; showHeaderContextDropdown = false; }}
		aria-label="Close dropdown"
	></button>
{/if}

<!-- Save to Profile Modal -->
{#if showSaveToProfileModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<button
			class="absolute inset-0 bg-black/50"
			onclick={() => showSaveToProfileModal = false}
			aria-label="Close modal"
		></button>

		<!-- Modal -->
		<div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="p-4 border-b border-gray-100">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Save Artifact to Profile</h3>
					<button
						onclick={() => showSaveToProfileModal = false}
						class="p-1 text-gray-400 hover:text-gray-600 rounded"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<p class="text-sm text-gray-500 mt-1">Select a context profile to save this artifact as a document</p>
			</div>

			<!-- Content -->
			<div class="p-4 max-h-80 overflow-y-auto">
				{#if availableProfiles.length === 0}
					<div class="text-center py-8">
						<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
						</div>
						<p class="text-sm text-gray-500">No profiles available</p>
						<a href="/contexts" class="text-sm text-blue-600 hover:underline mt-1 inline-block">Create a profile first</a>
					</div>
				{:else}
					<div class="space-y-2">
						{#each availableProfiles as profile (profile.id)}
							<button
								onclick={() => selectedProfileForSave = profile.id}
								class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedProfileForSave === profile.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center flex-shrink-0 text-lg">
									{profile.icon || '📁'}
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900">{profile.name}</p>
									{#if profile.type}
										<p class="text-xs text-gray-500 capitalize">{profile.type}</p>
									{/if}
								</div>
								{#if selectedProfileForSave === profile.id}
									<svg class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
										<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="p-4 border-t border-gray-100 flex gap-3">
				<button
					onclick={() => showSaveToProfileModal = false}
					class="flex-1 px-4 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={saveArtifactToProfile}
					disabled={!selectedProfileForSave || savingArtifactToProfile}
					class="flex-1 px-4 py-2.5 text-sm font-medium text-white bg-gray-900 hover:bg-gray-800 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
				>
					{#if savingArtifactToProfile}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
						</svg>
						Saving...
					{:else}
						Save to Profile
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Task Generation Modal -->
{#if showTaskGenerationModal}
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[85vh] flex flex-col">
			<!-- Header -->
			<div class="p-5 border-b border-gray-100 flex items-center justify-between">
				<div>
					<h3 class="text-lg font-semibold text-gray-900">Generate Tasks from Plan</h3>
					<p class="text-sm text-gray-500 mt-0.5">Review and assign tasks extracted from "{taskGenerationArtifact?.title}"</p>
				</div>
				<button
					onclick={() => { showTaskGenerationModal = false; generatedTasks = []; }}
					class="p-2 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-colors"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Project Selection -->
			<div class="px-5 py-3 border-b border-gray-100 bg-gray-50">
				<label class="block text-sm font-medium text-gray-700 mb-1.5">Assign to Project</label>
				<select
					bind:value={selectedProjectForTasks}
					class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				>
					<option value="">Select a project...</option>
					{#each availableProjects as project}
						<option value={project.id}>{project.name}</option>
					{/each}
				</select>
			</div>

			<!-- Tasks List -->
			<div class="flex-1 overflow-y-auto p-5">
				{#if generatingTasks}
					<div class="flex flex-col items-center justify-center py-12">
						<div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center mb-4">
							<svg class="w-6 h-6 text-blue-600 animate-spin" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
						</div>
						<p class="text-sm font-medium text-gray-900">Analyzing plan...</p>
						<p class="text-xs text-gray-500 mt-1">Extracting actionable tasks from your artifact</p>
					</div>
				{:else if generatedTasks.length === 0}
					<div class="flex flex-col items-center justify-center py-12">
						<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-4">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
							</svg>
						</div>
						<p class="text-sm font-medium text-gray-900">No tasks extracted</p>
						<p class="text-xs text-gray-500 mt-1">Try with a different artifact or add tasks manually</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each generatedTasks as task, index}
							<div class="border border-gray-200 rounded-xl p-4 hover:border-gray-300 transition-colors">
								<div class="flex items-start justify-between gap-3 mb-2">
									<div class="flex-1 min-w-0">
										<h4 class="font-medium text-gray-900 text-sm">{task.title}</h4>
										{#if task.description}
											<p class="text-xs text-gray-500 mt-1 line-clamp-2">{task.description}</p>
										{/if}
									</div>
									<button
										onclick={() => removeGeneratedTask(index)}
										class="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500 transition-colors flex-shrink-0"
										aria-label="Remove task"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</div>
								<div class="flex items-center gap-3 mt-3">
									<div class="flex items-center gap-2">
										<span class="text-xs text-gray-500">Priority:</span>
										<span class="px-2 py-0.5 text-xs font-medium rounded-full {task.priority === 'high' ? 'bg-red-100 text-red-700' : task.priority === 'medium' ? 'bg-yellow-100 text-yellow-700' : 'bg-gray-100 text-gray-700'}">{task.priority}</span>
									</div>
									<div class="flex items-center gap-2 flex-1 min-w-0">
										<span class="text-xs text-gray-500 flex-shrink-0">Assign to:</span>
										<select
											value={task.assignee_id || ''}
											onchange={(e) => updateTaskAssignee(index, (e.target as HTMLSelectElement).value)}
											class="flex-1 min-w-0 px-2 py-1 text-xs border border-gray-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-blue-500"
										>
											<option value="">Unassigned</option>
											{#each availableTeamMembers as member}
												<option value={member.id}>{member.name} ({member.role})</option>
											{/each}
										</select>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="p-4 border-t border-gray-100 flex items-center justify-between">
				<div class="text-sm text-gray-500">
					{generatedTasks.length} task{generatedTasks.length !== 1 ? 's' : ''} ready
				</div>
				<div class="flex gap-3">
					<button
						onclick={() => { showTaskGenerationModal = false; generatedTasks = []; }}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
					>
						Cancel
					</button>
					<button
						onclick={confirmTaskCreation}
						disabled={!selectedProjectForTasks || generatedTasks.length === 0}
						class="px-4 py-2 text-sm font-medium text-white bg-green-600 hover:bg-green-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Create {generatedTasks.length} Task{generatedTasks.length !== 1 ? 's' : ''}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes blink {
		0%, 50% { opacity: 1; }
		51%, 100% { opacity: 0; }
	}

	:global(.cursor-blink) {
		animation: blink 1s step-end infinite;
	}
</style>
