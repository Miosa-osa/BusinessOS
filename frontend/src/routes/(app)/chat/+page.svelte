<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { api, apiClient, type ArtifactListItem, type Artifact, type Node, type ContextListItem, type CustomAgent, getCustomAgents } from '$lib/api';
	import FocusModeSelector from '$lib/components/chat/FocusModeSelector.svelte';
	import ProgressPanel, { type DelegatedTask } from '$lib/components/chat/ProgressPanel.svelte';
	import ContextPanel, { type ActiveResource } from '$lib/components/chat/ContextPanel.svelte';
	import DocumentUploadModal from '$lib/components/chat/DocumentUploadModal.svelte';
	import HybridSearchPanel from '$lib/components/chat/HybridSearchPanel.svelte';
	import { FOCUS_MODES, getDefaultOptions, getAgentForFocusMode } from '$lib/components/chat/focusModes';
	import ArtifactEditor from '$lib/components/artifacts/ArtifactEditor.svelte';
	import { markdownToBlocks } from '$lib/utils/markdown-blocks';
	import { MessageActions } from '$lib/components/ai-elements';
	import { ConversationListPanel } from '$lib/components/chat';
	import RoleContextBadge from '$lib/components/chat/RoleContextBadge.svelte';
	import { currentWorkspaceId } from '$lib/stores/workspaces';
	import { thinking, thinkingEnabled } from '$lib/stores/thinking';
	import AgentSelector from '$lib/components/agents/AgentSelector.svelte';

	// Extract thinking content from message
	// Uses flexible regex to match variations like <thinking>, <thinkingng>, <thinkingg>, etc.
	function extractThinking(content: string): { thinking: string; mainContent: string } {
		if (!content) return { thinking: '', mainContent: '' };

		// Flexible regex that matches <think...> variations
		const thinkingMatch = content.match(/<think[a-z]*\s*>([\s\S]*?)<\/think[a-z]*\s*>/);
		if (thinkingMatch) {
			const thinking = thinkingMatch[1].trim();
			const mainContent = content.replace(/<think[a-z]*\s*>[\s\S]*?<\/think[a-z]*\s*>/, '').trim();
			return { thinking, mainContent };
		}
		return { thinking: '', mainContent: content };
	}

	// Markdown rendering helper - ChatGPT-style formatting
	// Strips thinking tags if they remain
	function renderMarkdown(text: string): string {
		if (!text) return '';

		// Remove thinking tags (any variation) and their content if extractThinking missed them? 
		// Actually extractThinking separates them. But if `renderMarkdown` receives text with tags (e.g. streaming chunks or leftover), we should strip the tags.
		// If text contains <thinking>...</thinking>, we should probably hide it or render it? 
		// The design is to show it in the separate box. So we strip it from the main markdown.
		
		// Remove robustly:
		text = text.replace(/<think[a-z]*\s*>[\s\S]*?<\/think[a-z]*\s*>/gi, '').trim(); 
		// Also strip orphan tags if they exist
		text = text.replace(/<\/?[a-z]*think[a-z]*\s*>/gi, '');

		let html = text
			// Escape HTML first to prevent XSS
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			// Code blocks (must be before other formatting)
			.replace(/```(\w+)?\n([\s\S]*?)```/g, '<pre class="chat-code-block"><code>$2</code></pre>')
			// Inline code
			.replace(/`([^`]+)`/g, '<code class="chat-inline-code">$1</code>')
			// Bold
			.replace(/\*\*([^*]+)\*\*/g, '<strong class="chat-bold">$1</strong>')
			// Italic
			.replace(/\*([^*]+)\*/g, '<em class="chat-italic">$1</em>')
			// Major section headers with numbers (1. Title, 2. Title) - ChatGPT style
			.replace(/^(\d+)\.\s+([A-Z][^\n]+)$/gm, '<div class="chat-section-divider"></div><h2 class="chat-section-header"><span class="chat-section-number">$1.</span> $2</h2>')
			// Sub-section headers with letters (A. Title, B. Title)
			.replace(/^([A-Z])\.\s+(.+)$/gm, '<h3 class="chat-subsection-header"><span class="chat-subsection-letter">$1.</span> $2</h3>')
			// Headers (###, ##, #)
			.replace(/^### (.+)$/gm, '<h4 class="chat-h4">$1</h4>')
			.replace(/^## (.+)$/gm, '<h3 class="chat-h3">$1</h3>')
			.replace(/^# (.+)$/gm, '<h2 class="chat-h2">$1</h2>')
			// Roman numeral sections (I. II. III. etc)
			.replace(/^([IVX]+)\.\s+(.+)$/gm, '<div class="chat-section-divider"></div><h2 class="chat-section-header">$1. $2</h2>')
			// Special labels like "Outcome:", "Example:", "Includes:" - make them bold
			.replace(/^(Outcome|Example|Includes|Note|Important|Key|Summary):\s*/gm, '<p class="chat-label"><strong>$1:</strong> ')
			// Numbered list items with bold text
			.replace(/^(\d+)\.\s+\*\*(.+?)\*\*:?\s*(.*)$/gm, '<div class="chat-list-item"><span class="chat-list-number">$1.</span><div class="chat-list-content"><strong class="chat-bold">$2</strong>$3</div></div>')
			// Regular numbered lists
			.replace(/^(\d+)\.\s+(.+)$/gm, '<div class="chat-list-item"><span class="chat-list-number">$1.</span><span class="chat-list-content">$2</span></div>')
			// Nested bullet points (indented with spaces/tabs)
			.replace(/^[\t\s]{2,}[\u2022\-\*]\s+(.+)$/gm, '<div class="chat-nested-bullet"><span class="chat-bullet">&bull;</span><span>$1</span></div>')
			// Bullet points with asterisk or dash
			.replace(/^[\u2022\-\*]\s+(.+)$/gm, '<div class="chat-bullet-item"><span class="chat-bullet">&bull;</span><span>$1</span></div>')
			// Horizontal rules
			.replace(/^---+$/gm, '<div class="chat-section-divider"></div>')
			// Links
			.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="chat-link" target="_blank" rel="noopener">$1</a>')
			// Paragraphs (double newlines)
			.replace(/\n\n/g, '</p><p class="chat-paragraph">')
			// Single newlines within content
			.replace(/\n/g, '<br />');

		// Wrap in paragraph if not starting with block element
		if (!html.startsWith('<h') && !html.startsWith('<pre') && !html.startsWith('<div')) {
			html = '<p class="chat-paragraph">' + html + '</p>';
		}

		return html;
	}

	// Usage data interface
	interface UsageData {
		input_tokens: number;
		output_tokens: number;
		thinking_tokens: number;
		total_tokens: number;
		duration_ms: number;
		tps: number;
		provider: string;
		model: string;
		estimated_cost: number;
	}

	// Message interface
	interface ChatMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		artifacts?: { title: string; type: string; content: string }[];
		usage?: UsageData;
	}

	// UI State
	let messagesContainer: HTMLDivElement | undefined = $state(undefined);
	let inputRef: HTMLTextAreaElement | undefined = $state(undefined);
	let inputValue = $state('');
	let selectedModel = $state('');
	let selectedContextIds = $state<string[]>([]);
	let chatSidebarOpen = $state(false);
	let artifactsPanelOpen = $state(false);
	let searchQuery = $state('');
	let showContextDropdown = $state(false);
	let showHeaderContextDropdown = $state(false);
	let showModelDropdown = $state(false);
	let showNodeDropdown = $state(false);
	let copiedMessageId: string | null = $state(null);
	let filterTab: 'all' | 'pinned' | 'recent' = $state('all');
	let showUsageInChat = $state(true); // User preference for showing usage stats
	// AI model settings from user preferences
	let aiTemperature = $state(0.7);
	let aiMaxTokens = $state(8192);
	let aiTopP = $state(0.9);
	let useCOT = $state(false); // Chain of Thought mode - synced with thinking store

	// Focus Mode state
	let focusModeEnabled = $state(true); // Focus vs Classic mode
	let selectedFocusId = $state<string | null>(null); // Current focus card
	let focusOptions = $state<Record<string, string>>({}); // Selected options
	let showFocusDropdown = $state(false); // Focus mode dropdown in existing conversations
	let rightPanelOpen = $state(false); // Progress/Context panel - closed by default
	let rightPanelTab = $state<'progress' | 'context' | 'artifacts'>('progress');
	let delegatedTasks = $state<DelegatedTask[]>([]); // Tasks from AI
	let activeResources = $state<ActiveResource[]>([]); // Loaded documents/contexts
	let focusModeInitialInput = $state(''); // Voice transcript for Focus mode

	// Chain of Thought (thinking) state
	let currentThinking = $state(''); // Current thinking content being streamed
	let thinkingExpanded = $state(true); // Whether thinking panel is expanded
	let hasThinking = $state(false); // Whether current message has thinking

	// File attachment state
	interface AttachedFile {
		id: string;
		name: string;
		type: string;
		size: number;
		content?: string; // base64 for images
		documentId?: string; // ID returned from document upload API
		uploading?: boolean; // Upload in progress
		uploadError?: string; // Upload error message
	}
	let attachedFiles = $state<AttachedFile[]>([]);
	let fileInputRef: HTMLInputElement | undefined = $state(undefined);
	let showPlusMenu = $state(false);
	let showDocumentUploadModal = $state(false);
	let showHybridSearchPanel = $state(false);

	// Voice recording state
	let isRecording = $state(false);
	let mediaRecorder: MediaRecorder | null = null;
	let audioChunks: Blob[] = [];
	let whisperAvailable = $state(false);
	let isTranscribing = $state(false);
	let recordingDuration = $state(0);
	let recordingInterval: ReturnType<typeof setInterval> | null = null;

	// Live transcript state (using Web Speech API)
	let liveTranscript = $state('');
	let speechRecognition: SpeechRecognition | null = null;

	// Audio visualization state
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let audioDataArray: Uint8Array<ArrayBuffer> | null = null;
	let waveformBars = $state<number[]>(Array(40).fill(2)); // 40 bars for waveform
	let animationFrameId: number | null = null;

	// Format recording duration as MM:SS
	let recordingTimeDisplay = $derived(() => {
		const mins = Math.floor(recordingDuration / 60);
		const secs = recordingDuration % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	});

	// Contexts state
	let availableContexts: ContextListItem[] = $state([]);
	let loadingContexts = $state(false);

	// Slash command autocomplete state
	interface SlashCommand {
		name: string;
		display_name: string;
		description: string;
		icon: string;
		category: string;
	}
	let availableCommands: SlashCommand[] = $state([]);
	let showCommandSuggestions = $state(false);
	let filteredCommands = $state<SlashCommand[]>([]);
	let commandDropdownIndex = $state(0);
	let activeCommand = $state<SlashCommand | null>(null);

	// Agent mention state (@agent)
	interface AgentPreset {
		id: string;
		name: string;
		display_name: string;
		description: string | null;
		avatar: string | null;
		category: string | null;
	}
	let availableAgents: AgentPreset[] = $state([]);
	let detectedAgent = $state<AgentPreset | null>(null);
	let showAgentSuggestions = $state(false);
	let filteredAgents = $state<AgentPreset[]>([]);
	let agentDropdownIndex = $state(0);

	// Custom Agents state
	let customAgents: CustomAgent[] = $state([]);
	let selectedAgent: CustomAgent | null = $state(null);
	let loadingAgents = $state(false);

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
	let artifactPanelWidth = $state(400);
	let isResizing = $state(false);
	let resizeStartX = $state(0);
	let resizeStartWidth = $state(0);

	// Right panel resize state
	let rightPanelWidth = $state(320);
	let isResizingRightPanel = $state(false);
	let rightPanelResizeStartX = $state(0);
	let rightPanelResizeStartWidth = $state(0);

	// Currently viewing artifact in panel
	let viewingArtifactFromMessage: { title: string; type: string; content: string } | null = $state(null);

	// Derived: Whether we're actively viewing/generating an artifact (use split view)
	let isArtifactFocused = $derived(viewingArtifactFromMessage !== null || generatingArtifact || selectedArtifact !== null);

	// Derived: All artifacts combined (API + message artifacts)
	// Type for combined artifacts
	type CombinedArtifact = {
		id: string;
		title: string;
		type: string;
		content?: string;
		summary?: string | null;
		version?: number;
		fromMessage?: boolean;
		messageId?: string;
		context_name?: string | null;
		project_id?: string | null;
		context_id?: string | null;
		conversation_id?: string | null;
		created_at?: string;
		updated_at?: string;
	};

	let allArtifacts = $derived((): CombinedArtifact[] => {
		// Get artifacts from messages
		const messageArtifacts: CombinedArtifact[] = messages.flatMap((m) =>
			(m.artifacts || []).map((a, artIndex) => ({
				id: `msg-${m.id}-${artIndex}`,
				title: a.title,
				type: a.type,
				content: a.content,
				summary: a.content?.substring(0, 100) || '',
				version: 1,
				fromMessage: true,
				messageId: m.id
			}))
		);

		// Combine with API artifacts (avoid duplicates by title)
		const apiArtifactTitles = new Set(artifacts.map(a => a.title));
		const uniqueMessageArtifacts = messageArtifacts.filter(a => !apiArtifactTitles.has(a.title));

		return [...uniqueMessageArtifacts, ...artifacts.map(a => ({ ...a } as CombinedArtifact))];
	});

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
	let projectDropdownIndex = $state(0); // For keyboard navigation
	let showNewProjectModal = $state(false);
	let newProjectName = $state('');
	let creatingProject = $state(false);

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

	// Update artifact content from editor
	function updateArtifactContent(content: string) {
		if (viewingArtifactFromMessage) {
			viewingArtifactFromMessage = {
				...viewingArtifactFromMessage,
				content
			};
		}
	}

	// Update selected artifact content from editor
	function updateSelectedArtifactContent(content: string) {
		if (selectedArtifact) {
			selectedArtifact = {
				...selectedArtifact,
				content
			};
			// TODO: Could also save to API here
		}
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
			// Convert artifact content to blocks for consistent storage
			const blocks = markdownToBlocks(viewingArtifactFromMessage.content);

			// Determine icon based on artifact type
			const icon = viewingArtifactFromMessage.type === 'plan' ? 'ðŸ“‹' :
					  viewingArtifactFromMessage.type === 'proposal' ? 'ðŸ“„' :
					  viewingArtifactFromMessage.type === 'framework' ? 'ðŸ—ï¸' :
					  viewingArtifactFromMessage.type === 'sop' ? 'ðŸ“–' :
					  viewingArtifactFromMessage.type === 'report' ? 'ðŸ“Š' : 'ðŸ“';

			// Create a new context document with the artifact content as blocks
			// If 'loose' is selected, don't set parent_id (will be a loose document)
			await api.createContext({
				name: viewingArtifactFromMessage.title,
				type: 'document',
				blocks: blocks,
				parent_id: selectedProfileForSave === 'loose' ? undefined : selectedProfileForSave,
				icon
			});

			showSaveToProfileModal = false;
			selectedProfileForSave = '';
			viewingArtifactFromMessage = null;

			// Close artifact panel after successful save
			artifactsPanelOpen = false;
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

	// Auto-save artifact to database when generated
	async function autoSaveArtifact(artifactData: { title: string; type: string; content: string }) {
		try {
			console.log('[Artifact] Auto-saving artifact:', artifactData.title);
			const allowedArtifactTypes = new Set([
				'proposal',
				'sop',
				'framework',
				'agenda',
				'report',
				'plan',
				'code',
				'document',
				'markdown',
				'other'
			]);
			const safeType = allowedArtifactTypes.has(artifactData.type) ? (artifactData.type as any) : ('other' as any);
			const savedArtifact = await api.createArtifact({
				title: artifactData.title,
				type: safeType,
				content: artifactData.content,
				conversation_id: conversationId || undefined,
				project_id: selectedProjectId || undefined
			});
			console.log('[Artifact] Auto-saved successfully:', savedArtifact.id);
			// Refresh artifacts list
			loadArtifacts();

			// Automatically select and view the saved artifact
			selectedArtifact = savedArtifact;
			viewingArtifactFromMessage = null; // Clear ephemeral view
			artifactsPanelOpen = true; // Ensure panel is open
			rightPanelOpen = true; // Ensure side panel is open regarding of tab
			
			return savedArtifact;
		} catch (e) {
			console.error('[Artifact] Failed to auto-save:', e);
			return null;
		}
	}

	// Delete an artifact
	async function deleteArtifactById(id: string) {
		if (!confirm('Are you sure you want to delete this artifact?')) return;

		try {
			await api.deleteArtifact(id);
			// Remove from local list
			artifacts = artifacts.filter(a => a.id !== id);
			// Close detail view if this artifact was selected
			if (selectedArtifact?.id === id) {
				selectedArtifact = null;
			}
		} catch (e) {
			console.error('[Artifact] Failed to delete:', e);
		}
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
			const taskCount = generatedTasks.length;
			const artifactTitle = taskGenerationArtifact?.title;

			// Create tasks via API
			for (const task of generatedTasks) {
				await api.createTask({
					title: task.title,
					description: task.description,
					project_id: selectedProjectForTasks,
					priority: task.priority,
					assignee_id: task.assignee_id
				});
			}

			// Add confirmation message to chat
			const confirmMsgId = crypto.randomUUID();
			messages = [...messages, {
				id: confirmMsgId,
				role: 'assistant',
				content: `âœ… Created ${taskCount} tasks from "${artifactTitle ?? 'the artifact'}". You can view them in the Tasks tab.`
			}];

			// Close modal and reset state
			showTaskGenerationModal = false;
			generatedTasks = [];
			taskGenerationArtifact = null;
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
		if (inlineTasksForArtifact.length === 0) return;

		creatingInlineTasks = true;
		try {
			// Create tasks via API
			for (const task of inlineTasksForArtifact) {
				const taskData: any = {
					title: task.title,
					description: task.description,
					priority: task.priority,
					assignee_id: task.assignee_id
				};
				// Only include project_id if one is selected
				if (selectedProjectId) {
					taskData.project_id = selectedProjectId;
				}
				await api.createTask(taskData);
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
				description: p.description ?? undefined
			}));
			// Also update availableProjects for task generation
			availableProjects = projectsList;
		} catch (e) {
			console.error('Failed to load projects:', e);
		} finally {
			loadingProjects = false;
		}
	}

	// Load available slash commands
	async function loadCommands() {
		try {
			const response = await fetch('/api/ai/commands');
			if (response.ok) {
				const data = await response.json();
				availableCommands = data.commands || [];
				console.log('[Chat] Loaded', availableCommands.length, 'slash commands');
			}
		} catch (e) {
			console.error('Failed to load commands:', e);
		}
	}

	// Load agent presets for @mention autocomplete
	async function loadAgentPresets() {
		try {
			const response = await fetch('/api/ai/agents/presets');
			if (response.ok) {
				const data = await response.json();
				availableAgents = data.presets || [];
				console.log('[Chat] Loaded', availableAgents.length, 'agent presets');
			}
		} catch (e) {
			console.error('Failed to load agent presets:', e);
		}
	}

	// Load custom agents for agent selector
	async function loadCustomAgents() {
		loadingAgents = true;
		try {
			const response = await getCustomAgents(false); // Only active agents
			customAgents = response.agents || [];
			console.log('[Chat] Loaded', customAgents.length, 'custom agents');

			// Add custom agents to availableAgents for @mention autocomplete
			// Convert CustomAgent to AgentPreset format
			const customAgentPresets: AgentPreset[] = customAgents.map(agent => ({
				id: agent.id,
				name: agent.name,
				display_name: agent.display_name,
				description: agent.description || null,
				avatar: agent.avatar || null,
				category: agent.category || 'custom'
			}));

			// Remove any existing custom agents to avoid duplicates, then append new ones
			// Keep only built-in agents (category !== 'custom'), then add fresh custom agent list
			const builtInAgents = availableAgents.filter(a => a.category !== 'custom');
			availableAgents = [...builtInAgents, ...customAgentPresets];
			console.log('[Chat] Total available agents for @mention:', availableAgents.length,
				'(', customAgentPresets.length, 'custom,', builtInAgents.length, 'built-in)');
		} catch (e) {
			console.error('Failed to load custom agents:', e);
			customAgents = [];
		} finally {
			loadingAgents = false;
		}
	}

	// Handle agent selection
	function handleAgentSelect(agent: CustomAgent | null) {
		selectedAgent = agent;
		console.log('[Chat] Selected agent:', agent?.display_name || 'Default');
	}

	// Detect @agent mentions in input and set detectedAgent
	function detectAgentMention(text: string) {
		// Match @agent-name at the start of the message
		const agentMatch = text.match(/^@(\w[\w-]*)/);
		if (agentMatch && availableAgents.length > 0) {
			const agentName = agentMatch[1].toLowerCase();
			const agent = availableAgents.find(a => a.name.toLowerCase() === agentName);
			if (agent) {
				detectedAgent = agent;
				return;
			}
		}
		detectedAgent = null;
	}

	// Create a new project quickly from chat
	async function createProjectQuick() {
		if (!newProjectName.trim()) return;
		creatingProject = true;
		try {
			const project = await api.createProject({
				name: newProjectName.trim(),
				status: 'active'
			});
			// Add to list and select it
			projectsList = [...projectsList, { id: project.id, name: project.name, description: project.description ?? undefined }];
			selectedProjectId = project.id;
			newProjectName = '';
			showNewProjectModal = false;
			// Send message if there's input
			if (inputValue.trim()) {
				setTimeout(() => handleSendMessage(), 50);
			}
		} catch (e) {
			console.error('Failed to create project:', e);
		} finally {
			creatingProject = false;
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

	// Load user settings for chat preferences
	async function loadUserSettings() {
		try {
			const response = await apiClient.get('/settings');
			if (response.ok) {
				const settings = await response.json();
				// Check custom settings or model_settings for showUsageInChat
				if (settings.model_settings?.showUsageInChat !== undefined) {
					showUsageInChat = settings.model_settings.showUsageInChat;
				} else if (settings.custom_settings?.showUsageInChat !== undefined) {
					showUsageInChat = settings.custom_settings.showUsageInChat;
				}
				// Load preferred model from settings
				if (settings.default_model) {
					let savedModel = settings.default_model;
					console.log('[Chat] Loaded preferred model from settings:', savedModel);

					// Migration: fix old model names that are missing the -cloud suffix
					// These cloud models require the -cloud suffix to route correctly
					const cloudModelMigrations: Record<string, string> = {
						'qwen3-coder:480b': 'qwen3-coder:480b-cloud',
						'qwen3-coder:30b': 'qwen3-coder:30b-cloud',
					};
					if (cloudModelMigrations[savedModel]) {
						console.log('[Chat] Migrating old model name:', savedModel, '->', cloudModelMigrations[savedModel]);
						savedModel = cloudModelMigrations[savedModel];
						// Save the corrected model preference
						saveModelPreference(savedModel);
					}

					selectedModel = savedModel;
				}
				// Load AI model settings (temperature, max_tokens, top_p)
				if (settings.model_settings) {
					if (typeof settings.model_settings.temperature === 'number') {
						aiTemperature = settings.model_settings.temperature;
					}
					if (typeof settings.model_settings.max_tokens === 'number') {
						aiMaxTokens = settings.model_settings.max_tokens;
					}
					if (typeof settings.model_settings.top_p === 'number') {
						aiTopP = settings.model_settings.top_p;
					}
					console.log('[Chat] Loaded AI settings:', { temperature: aiTemperature, max_tokens: aiMaxTokens, top_p: aiTopP });
				}
			}
		} catch (error) {
			console.error('Failed to load user settings:', error);
		}
	}

	// Save model preference to user settings
	async function saveModelPreference(modelId: string) {
		try {
			const response = await apiClient.put('/settings', {
				default_model: modelId
			});
			if (response.ok) {
				console.log('[Chat] Saved model preference:', modelId);
			} else {
				console.error('[Chat] Failed to save model preference');
			}
		} catch (error) {
			console.error('[Chat] Error saving model preference:', error);
		}
	}

	// Select a model and save preference
	function selectModel(modelId: string) {
		selectedModel = modelId;
		showModelDropdown = false;
		saveModelPreference(modelId);

		// Pre-warm the model for Ollama local (non-blocking)
		if (activeProvider === 'ollama_local' && !warmedUpModels.has(modelId)) {
			warmupModel(modelId);
		}
	}

	// State for tracking model warmup
	let warmingUpModel = $state<string | null>(null);

	// Pre-warm a model by sending a minimal request to load it into memory
	// Uses the dedicated warmup endpoint which is faster than the chat endpoint
	// This is a best-effort operation - failures are silently ignored
	async function warmupModel(modelId: string) {
		// Don't warm up if already warming or already warmed
		if (warmingUpModel === modelId || warmedUpModels.has(modelId)) return;

		warmingUpModel = modelId;
		console.log(`[Warmup] Starting warmup for model: ${modelId}`);

		try {
			// Use the dedicated warmup endpoint (much faster than chat endpoint)
			const result = await api.warmupModel(modelId);

			if (result.status === 'ready' || result.status === 'skipped') {
				// Mark model as warmed up
				const newSet = new Set(warmedUpModels);
				newSet.add(modelId);
				warmedUpModels = newSet;
				if (typeof window !== 'undefined') {
					localStorage.setItem('warmedUpModels', JSON.stringify([...newSet]));
				}
				console.log(`[Warmup] Model ${modelId} ready (${result.status})`);
			}
		} catch (error: unknown) {
			// Silently ignore warmup failures - the model will load on first real message
			if (error instanceof Error) {
				console.warn(`[Warmup] Model warmup for ${modelId} skipped:`, error.message);
			}
		} finally {
			warmingUpModel = null;
		}
	}

	onMount(async () => {
		// Load panel state from localStorage
		try {
			const savedPanelState = localStorage.getItem('chat_panel_state');
			if (savedPanelState) {
				const { rightPanelOpen: rp, rightPanelTab: rt, rightPanelWidth: rw } = JSON.parse(savedPanelState);
				if (typeof rp === 'boolean') rightPanelOpen = rp;
				if (rt === 'progress' || rt === 'context' || rt === 'artifacts') rightPanelTab = rt;
				if (typeof rw === 'number' && rw >= 280 && rw <= 500) rightPanelWidth = rw;
			}
		} catch (e) {
			console.warn('Failed to load panel state from localStorage:', e);
		}

		// Load user settings FIRST to get preferred model
		await loadUserSettings();

		// Then load models (will respect the saved model preference)
		await loadModels();

		// Pre-warm the selected model immediately (non-blocking)
		// This significantly reduces first-message latency for Ollama
		if (selectedModel && activeProvider === 'ollama_local') {
			warmupModel(selectedModel);
		}

		// Load other data in parallel
		loadActiveNode();
		loadContexts();
		loadConversations();
		loadProjects();
		loadTeamMembers();
		loadCommands(); // Load slash commands for autocomplete

		// Load agents in sequence: presets first, then custom agents
		// This ensures custom agents can be added to availableAgents array
		await loadAgentPresets(); // Load agent presets for @mention
		await loadCustomAgents(); // Load custom agents and add to @mention list

		checkWhisperStatus(); // Check if voice transcription is available

		// Load thinking settings to sync COT mode
		await thinking.loadSettings();

		// Set artifact panel width to ~50% of available space (window width minus sidebars)
		// Left sidebar is ~256px, chat sidebar is ~256px when open
		const availableWidth = window.innerWidth - 256; // Subtract left sidebar
		artifactPanelWidth = Math.floor(availableWidth / 2);

		// Check for quick chat message from dock
		checkForQuickChatMessage();

		// Check for voice transcript from dock recording
		checkForVoiceTranscript();

		// Check for chat transfer from Spotlight
		checkForSpotlightTransfer();
	});

	// Save panel state to localStorage when it changes
	$effect(() => {
		// Access the reactive values to track them
		const state = {
			rightPanelOpen,
			rightPanelTab,
			rightPanelWidth
		};
		// Only save after initial mount
		if (typeof window !== 'undefined') {
			localStorage.setItem('chat_panel_state', JSON.stringify(state));
		}
	});

	// Sync useCOT with thinking settings
	$effect(() => {
		useCOT = $thinkingEnabled;
	});

	// Handle chat transfer from Spotlight popup
	function checkForSpotlightTransfer() {
		const transferData = sessionStorage.getItem('spotlightChatTransfer');
		if (transferData) {
			try {
				const data = JSON.parse(transferData);
				sessionStorage.removeItem('spotlightChatTransfer');

				console.log('[Chat] Received transfer from Spotlight:', data);

				// If there's a conversation ID, load it
				if (data.conversationId) {
					conversationId = data.conversationId;
					activeConversationId = data.conversationId;
					selectConversation(data.conversationId);
				} else if (data.messages && data.messages.length > 0) {
					// If no conversation ID but has messages, display them
					messages = data.messages.map((m: { role: string; content: string }, i: number) => ({
						id: `spotlight-${i}`,
						role: m.role as 'user' | 'assistant',
						content: m.content
					}));
				}

				// Set the project if provided
				if (data.projectId) {
					selectedProjectId = data.projectId;
				}
			} catch (e) {
				console.error('Failed to parse spotlight transfer data:', e);
			}
		}
	}

	// Handle quick chat message from dock
	async function checkForQuickChatMessage() {
		const quickChatData = sessionStorage.getItem('quickChatMessage');
		if (quickChatData) {
			try {
				const data = JSON.parse(quickChatData);
				// Clear the sessionStorage immediately to prevent re-processing
				sessionStorage.removeItem('quickChatMessage');

				// Check if message is recent (within last 5 seconds)
				if (Date.now() - data.timestamp < 5000) {
					// Start a new conversation with this message
					if (data.isNewConversation) {
						startNewConversation();
					}

					// Set project if provided
					if (data.projectId) {
						selectedProjectId = data.projectId;
					}

					// Set model if provided
					if (data.model) {
						selectedModel = data.model;
					}

					// Set the input value and send
					inputValue = data.message;

					// Wait for the input to be set, then send
					await tick();
					handleSendMessage();
				}
			} catch (e) {
				console.error('Failed to parse quick chat message:', e);
			}
		}
	}

	// Handle voice transcript from dock recording
	function checkForVoiceTranscript() {
		const transcriptData = sessionStorage.getItem('voiceTranscript');
		if (transcriptData) {
			try {
				const data = JSON.parse(transcriptData);
				// Clear the sessionStorage immediately to prevent re-processing
				sessionStorage.removeItem('voiceTranscript');

				// Check if transcript is recent (within last 10 seconds - transcription can take a bit)
				if (Date.now() - data.timestamp < 10000) {
					// Route to appropriate input based on mode
					if (focusModeEnabled && !conversationId) {
						// Focus mode - pass to FocusModeSelector
						focusModeInitialInput = data.message;
						// Clear after a short delay so it can be re-used
						setTimeout(() => { focusModeInitialInput = ''; }, 500);
					} else {
						// Classic mode or in conversation - use regular input
						inputValue = data.message;
						// Focus the input after a short delay
						setTimeout(() => {
							inputRef?.focus();
						}, 100);
					}
				}
			} catch (e) {
				console.error('Failed to parse voice transcript:', e);
			}
		}
	}

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
	// (handled by artifactsLoadedOnce effect above)

	// Available models - loaded from API
	// Capabilities: vision (image analysis), tools (function calling), coding, reasoning, rag (embedding/indexing), multilingual
	type ModelCapability = 'vision' | 'tools' | 'coding' | 'reasoning' | 'rag' | 'multilingual' | 'fast';

	interface ModelOption {
		id: string;
		name: string;
		description: string;
		type: 'cloud' | 'local';
		size?: string;
		capabilities?: ModelCapability[];
	}

	// Capability badge colors and labels
	const capabilityInfo: Record<ModelCapability, { label: string; color: string; icon: string }> = {
		vision: { label: 'Vision', color: 'bg-purple-100 text-purple-700', icon: 'ðŸ‘ï¸' },
		tools: { label: 'Tools', color: 'bg-blue-100 text-blue-700', icon: 'ðŸ”§' },
		coding: { label: 'Code', color: 'bg-green-100 text-green-700', icon: 'ðŸ’»' },
		reasoning: { label: 'Reasoning', color: 'bg-orange-100 text-orange-700', icon: 'ðŸ§ ' },
		rag: { label: 'RAG', color: 'bg-cyan-100 text-cyan-700', icon: 'ðŸ“š' },
		multilingual: { label: 'Multi-lang', color: 'bg-pink-100 text-pink-700', icon: 'ðŸŒ' },
		fast: { label: 'Fast', color: 'bg-yellow-100 text-yellow-700', icon: 'âš¡' },
	};

	// Dynamic model state
	let installedModels = $state<ModelOption[]>([]);
	let ollamaCloudModels = $state<ModelOption[]>([]); // Models with -cloud suffix (routed via local Ollama)
	let loadingModels = $state(false);
	let activeProvider = $state('ollama_local');
	// Track which cloud providers are configured (have API keys set)
	let configuredProviders = $state<Set<string>>(new Set());
	// Track models that have been warmed up (completed at least one inference)
	// Persist in localStorage so it survives page reloads
	let warmedUpModels = $state<Set<string>>(
		typeof window !== 'undefined' && localStorage.getItem('warmedUpModels')
			? new Set(JSON.parse(localStorage.getItem('warmedUpModels') || '[]'))
			: new Set()
	);

	// Detect capabilities for local models based on model ID
	function getModelCapabilities(modelId: string): ModelCapability[] {
		const id = modelId.toLowerCase();
		const caps: ModelCapability[] = [];

		// Vision models
		if (id.includes('vision') || id.includes('llava') || id.includes('minicpm-v') || id.includes('bakllava') || id.includes('moondream')) {
			caps.push('vision');
		}

		// Coding models
		if (id.includes('code') || id.includes('coder') || id.includes('deepseek') || id.includes('qwen3-coder') || id.includes('starcoder') || id.includes('codellama')) {
			caps.push('coding');
		}

		// Reasoning models
		if (id.includes('deepseek-r1') || id.includes('o1') || id.includes('reasoning') || id.includes(':70b') || id.includes(':72b') || id.includes(':480b')) {
			caps.push('reasoning');
		}

		// Tool-capable models (most modern models support tools)
		if (id.includes('qwen') || id.includes('llama3') || id.includes('mistral') || id.includes('claude') || id.includes('gpt')) {
			caps.push('tools');
		}

		// Multilingual models
		if (id.includes('qwen') || id.includes('aya') || id.includes('multilingual')) {
			caps.push('multilingual');
		}

		// Embedding/RAG models
		if (id.includes('embed') || id.includes('nomic') || id.includes('mxbai') || id.includes('bge') || id.includes('e5')) {
			caps.push('rag');
		}

		// Fast models (small sizes)
		if (id.includes(':1b') || id.includes(':3b') || id.includes(':4b') || id.includes('instant') || id.includes('mini') || id.includes('tiny')) {
			caps.push('fast');
		}

		return caps;
	}

	// Get a nice description for installed models
	function getModelDescription(modelId: string, family?: string): string {
		const id = modelId.toLowerCase();
		// Match by model family/name
		if (id.includes('llama3.2-vision')) return 'Multimodal vision model';
		if (id.includes('llama3.2')) return 'Fast and capable, great all-rounder';
		if (id.includes('llama3.1')) return 'Powerful general purpose model';
		if (id.includes('llama3.3')) return 'Latest Llama, best quality';
		if (id.includes('llama')) return 'Meta AI general purpose';
		if (id.includes('mistral')) return 'Excellent for general tasks';
		if (id.includes('mixtral')) return 'MoE model, very capable';
		if (id.includes('codellama')) return 'Optimized for code tasks';
		if (id.includes('deepseek-coder') || id.includes('deepseek:coder')) return 'Strong coding assistant';
		if (id.includes('deepseek-r1')) return 'Advanced reasoning model';
		if (id.includes('deepseek')) return 'Capable reasoning model';
		if (id.includes('qwen3-coder')) return 'Qwen coding specialist';
		if (id.includes('qwen2.5') || id.includes('qwen2')) return 'Strong reasoning and math';
		if (id.includes('qwen3')) return 'Latest Qwen, very capable';
		if (id.includes('qwen')) return 'Alibaba multilingual model';
		if (id.includes('llava')) return 'Vision-language model';
		if (id.includes('minicpm-v')) return 'Efficient vision model';
		if (id.includes('nomic-embed') || id.includes('mxbai-embed')) return 'Embedding model for RAG';
		if (id.includes('phi3') || id.includes('phi-3')) return 'Microsoft efficient model';
		if (id.includes('phi')) return 'Microsoft small but capable';
		if (id.includes('gemma2') || id.includes('gemma:2')) return 'Google efficient model';
		if (id.includes('gemma')) return 'Google lightweight model';
		if (id.includes('llava')) return 'Vision + Language multimodal';
		if (id.includes('bakllava')) return 'Vision model for images';
		if (id.includes('vicuna')) return 'Fine-tuned for chat';
		if (id.includes('wizard')) return 'Instruction-following model';
		if (id.includes('openchat')) return 'Optimized for conversation';
		if (id.includes('neural-chat')) return 'Intel optimized chat';
		if (id.includes('starling')) return 'Strong reasoning ability';
		if (id.includes('yi')) return '01.AI bilingual model';
		if (id.includes('command')) return 'Cohere instruction model';
		if (id.includes('dolphin')) return 'Uncensored assistant';
		if (id.includes('orca')) return 'Reasoning focused model';
		if (id.includes('nous')) return 'Research-focused model';
		if (id.includes('solar')) return 'Upstage efficient model';
		if (id.includes('zephyr')) return 'HuggingFace chat model';
		if (family) return family;
		return 'Local AI model';
	}

	// Cloud models for configured providers with capabilities
	const cloudModelsMap: Record<string, ModelOption[]> = {
		groq: [
			{ id: 'llama-3.3-70b-versatile', name: 'Llama 3.3 70B', description: 'Fast 70B model', type: 'cloud', capabilities: ['tools', 'coding', 'fast'] },
			{ id: 'llama-3.1-8b-instant', name: 'Llama 3.1 8B', description: 'Ultra-fast', type: 'cloud', capabilities: ['fast'] },
			{ id: 'mixtral-8x7b-32768', name: 'Mixtral 8x7B', description: '32k context', type: 'cloud', capabilities: ['coding', 'fast'] },
		],
		anthropic: [
			{ id: 'claude-sonnet-4-20250514', name: 'Claude Sonnet 4', description: 'Best for most tasks', type: 'cloud', capabilities: ['vision', 'tools', 'coding', 'reasoning'] },
			{ id: 'claude-opus-4-20250514', name: 'Claude Opus 4', description: 'Most capable', type: 'cloud', capabilities: ['vision', 'tools', 'coding', 'reasoning'] },
		],
		ollama_cloud: [
			// Qwen 3 Coder models (cloud variants for agentic/coding tasks)
			{ id: 'qwen3-coder:480b-cloud', name: 'Qwen3 Coder 480B', description: '480B cloud - best quality', type: 'cloud', capabilities: ['tools', 'coding', 'reasoning', 'multilingual'] },
			{ id: 'qwen3-coder:30b', name: 'Qwen3 Coder 30B', description: '30B coding model', type: 'cloud', capabilities: ['tools', 'coding', 'multilingual'] },
			// Qwen 3 standard models
			{ id: 'qwen3:4b', name: 'Qwen3 4B', description: 'Fast, efficient', type: 'cloud', capabilities: ['fast', 'multilingual'] },
			{ id: 'qwen3:8b', name: 'Qwen3 8B', description: 'Balanced performance', type: 'cloud', capabilities: ['tools', 'multilingual'] },
			{ id: 'qwen3:14b', name: 'Qwen3 14B', description: 'Capable mid-size', type: 'cloud', capabilities: ['tools', 'coding', 'multilingual'] },
			{ id: 'qwen3:30b', name: 'Qwen3 30B', description: 'Large model', type: 'cloud', capabilities: ['tools', 'coding', 'reasoning', 'multilingual'] },
			{ id: 'qwen3:32b', name: 'Qwen3 32B', description: 'Large model', type: 'cloud', capabilities: ['tools', 'coding', 'reasoning', 'multilingual'] },
			// Llama models
			{ id: 'llama3.3:70b', name: 'Llama 3.3 70B', description: 'Latest Llama model', type: 'cloud', capabilities: ['tools', 'coding', 'reasoning'] },
			{ id: 'llama3.2', name: 'Llama 3.2', description: 'Fast Llama model', type: 'cloud', capabilities: ['fast', 'tools'] },
			{ id: 'llama3.2-vision', name: 'Llama 3.2 Vision', description: 'Multimodal Llama', type: 'cloud', capabilities: ['vision', 'tools'] },
			// DeepSeek models
			{ id: 'deepseek-r1:671b', name: 'DeepSeek R1 671B', description: 'Full reasoning - cloud', type: 'cloud', capabilities: ['reasoning', 'coding', 'tools'] },
			{ id: 'deepseek-r1:70b', name: 'DeepSeek R1 70B', description: 'Reasoning model', type: 'cloud', capabilities: ['reasoning', 'coding'] },
			{ id: 'deepseek-r1:32b', name: 'DeepSeek R1 32B', description: 'Compact reasoning', type: 'cloud', capabilities: ['reasoning', 'coding', 'fast'] },
			// Vision models
			{ id: 'llava:34b', name: 'LLaVA 34B', description: 'Vision-language model', type: 'cloud', capabilities: ['vision'] },
			{ id: 'minicpm-v', name: 'MiniCPM-V', description: 'Efficient vision model', type: 'cloud', capabilities: ['vision', 'fast'] },
			// Embedding/RAG models
			{ id: 'nomic-embed-text', name: 'Nomic Embed', description: 'Text embeddings for RAG', type: 'cloud', capabilities: ['rag'] },
			{ id: 'mxbai-embed-large', name: 'MxBai Embed Large', description: 'High-quality embeddings', type: 'cloud', capabilities: ['rag'] },
			// Mistral
			{ id: 'mistral', name: 'Mistral', description: 'Mistral AI model', type: 'cloud', capabilities: ['tools', 'coding', 'fast'] },
		]
	};

	// Load actual models from API
	async function loadModels() {
		loadingModels = true;
		try {
			// Get provider info
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				activeProvider = data.active_provider || 'ollama_local';

				// Track which cloud providers are configured (have API keys)
				const configured = new Set<string>();
				for (const provider of data.providers || []) {
					if (provider.configured && provider.id !== 'ollama_local') {
						configured.add(provider.id);
					}
				}
				configuredProviders = configured;
			}

			// Get local models
			const localRes = await apiClient.get('/ai/models/local');
			if (localRes.ok) {
				const data = await localRes.json();
				const allOllamaModels = (data.models || [])
					.filter((m: any) => {
						// Filter out tiny stub models (cloud reference stubs that are empty)
						const size = m.size || '';
						const isTinyStub = size === '< 1 KB' || size === '0 B';
						return !isTinyStub;
					})
					.map((m: any) => ({
						id: m.id,
						name: m.name,
						description: getModelDescription(m.id, m.family),
						type: m.id.toLowerCase().includes('-cloud') ? 'cloud' as const : 'local' as const,
						size: m.size
					}));

				// Separate local and cloud models
				installedModels = allOllamaModels.filter((m: ModelOption) => m.type === 'local');
				ollamaCloudModels = allOllamaModels.filter((m: ModelOption) => m.type === 'cloud');
			}

			// Build the combined list of all available models (local + ollama cloud + configured cloud providers)
			// Use Array.from to ensure $state arrays are properly converted to plain arrays
			const localModelsArr = Array.isArray(installedModels) ? installedModels : [];
			const cloudModelsArr = Array.isArray(ollamaCloudModels) ? ollamaCloudModels : [];
			const allAvailableModels: ModelOption[] = [...localModelsArr, ...cloudModelsArr];
			for (const provider of configuredProviders) {
				const providerModels = cloudModelsMap[provider] || [];
				allAvailableModels.push(...providerModels);
			}

			// Set default model based on provider ONLY if no model is selected
			// (selectedModel might be set from user settings)
			if (!selectedModel || selectedModel === '') {
				// Prefer local models if available, otherwise use first cloud model
				if (localModelsArr.length > 0) {
					selectedModel = localModelsArr[0].id;
					saveModelPreference(selectedModel); // Save the default
				} else if (allAvailableModels.length > 0) {
					selectedModel = allAvailableModels[0].id;
					saveModelPreference(selectedModel); // Save the default
				}
			} else {
				// Validate that the saved model still exists in combined list
				if (allAvailableModels.length === 0 || !allAvailableModels.some(m => m.id === selectedModel)) {
					console.warn('[Chat] Saved model "' + selectedModel + '" not available, resetting to default');
					const oldModel = selectedModel;
					if (localModelsArr.length > 0) {
						selectedModel = localModelsArr[0].id;
					} else if (allAvailableModels.length > 0) {
						selectedModel = allAvailableModels[0].id;
					}
					// Save the corrected model to prevent this from happening again
					if (selectedModel !== oldModel) {
						saveModelPreference(selectedModel);
					}
				}
			}
		} catch (error) {
			console.error('Failed to load models:', error);
		} finally {
			loadingModels = false;
		}
	}

	// Derived: combine local models with cloud models from all configured providers
	let models = $derived.by(() => {
		// Include installed local models and Ollama cloud models
		const allModels: ModelOption[] = [...installedModels, ...ollamaCloudModels];

		// Add cloud models from all configured providers
		for (const provider of configuredProviders) {
			const providerModels = cloudModelsMap[provider] || [];
			// Add provider tag to each model for identification
			for (const model of providerModels) {
				allModels.push({
					...model
				});
			}
		}

		return allModels;
	});

	// Sidebar conversations
	interface SidebarConversation {
		id: string;
		title: string;
		timestamp: string;
		preview?: string;
		pinned?: boolean;
		projectId?: string;
		projectName?: string;
		messageCount?: number;
		conversationType?: 'chat' | 'focus';
		isArchived?: boolean;
	}

	let conversations: SidebarConversation[] = $state([]);
	let archivedConversations: SidebarConversation[] = $state([]);
	let activeConversationId = $state<string | null>(null);
	let showArchivedView = $state(false);

	// Sidebar handlers
	function handleArchiveConversation(id: string) {
		// Move conversation to archived
		const conv = conversations.find(c => c.id === id);
		if (conv) {
			conversations = conversations.filter(c => c.id !== id);
			archivedConversations = [...archivedConversations, { ...conv, isArchived: true }];
			// If archiving current conversation, clear it
			if (activeConversationId === id) {
				handleNewChat();
			}
		}
		console.log('[Chat] Archive conversation:', id, '(client-side only - backend not implemented)');
	}

	function handleUnarchiveConversation(id: string) {
		// Move conversation back to active
		const conv = archivedConversations.find(c => c.id === id);
		if (conv) {
			archivedConversations = archivedConversations.filter(c => c.id !== id);
			conversations = [...conversations, { ...conv, isArchived: false }];
		}
		console.log('[Chat] Unarchive conversation:', id, '(client-side only - backend not implemented)');
	}

	function handlePinConversation(id: string) {
		conversations = conversations.map(c =>
			c.id === id ? { ...c, pinned: !c.pinned } : c
		);
		console.log('[Chat] Toggle pin for conversation:', id, '(client-side only - backend not implemented)');
	}

	function handleRenameConversation(id: string) {
		// TODO: Show rename dialog
		const newTitle = prompt('Enter new title:');
		if (newTitle) {
			conversations = conversations.map(c =>
				c.id === id ? { ...c, title: newTitle } : c
			);
			console.log('[Chat] Rename conversation:', id, 'to', newTitle, '(client-side only - backend not implemented)');
		}
	}

	function handleExportConversation(id: string) {
		// TODO: Export conversation to file
		console.log('[Chat] Export conversation:', id, '(not implemented yet)');
	}

	function handleLinkProject(id: string) {
		// TODO: Show project linking dialog
		console.log('[Chat] Link to project:', id, '(not implemented yet)');
	}

	async function handleDeleteConversation(id: string) {
		if (!confirm('Are you sure you want to delete this conversation?')) return;

		try {
			await apiClient.delete(`/conversations/${id}`);
			conversations = conversations.filter(c => c.id !== id);
			archivedConversations = archivedConversations.filter(c => c.id !== id);
			// If deleting current conversation, clear it
			if (activeConversationId === id) {
				handleNewChat();
			}
			console.log('[Chat] Deleted conversation:', id);
		} catch (error) {
			console.error('[Chat] Failed to delete conversation:', error);
		}
	}

	function handleViewArchived() {
		showArchivedView = true;
	}

	function handleBackToChats() {
		showArchivedView = false;
	}

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

	// Context window limits for different models (in tokens)
	const modelContextLimits: Record<string, number> = {
		// Ollama local models
		'llama3.2': 128000,
		'llama3.2:1b': 128000,
		'llama3.2:3b': 128000,
		'llama3.1': 128000,
		'llama3.1:8b': 128000,
		'llama3.1:70b': 128000,
		'llama3': 8192,
		'mistral': 32768,
		'mistral:7b': 32768,
		'codellama': 16384,
		'qwen2.5': 128000,
		'qwen2.5:7b': 128000,
		'qwen2.5:14b': 128000,
		'qwen2.5:32b': 128000,
		'qwen3': 40960,
		'qwen3:0.6b': 32768,
		'qwen3:1.7b': 32768,
		'qwen3:4b': 32768,
		'qwen3:8b': 40960,
		'qwen3:14b': 40960,
		'qwen3:30b': 40960,
		'qwen3:32b': 40960,
		'qwen3:235b': 40960,
		'qwen3:480b': 40960,
		'phi3': 128000,
		'gemma2': 8192,
		'deepseek-coder': 16384,
		'deepseek-r1': 128000,
		// Cloud models
		'gpt-4': 8192,
		'gpt-4-turbo': 128000,
		'gpt-4o': 128000,
		'gpt-3.5-turbo': 16384,
		'claude-3-opus': 200000,
		'claude-3-sonnet': 200000,
		'claude-3-haiku': 200000,
		'llama-3.2-3b-preview': 128000,
		'llama-3.2-11b-vision-preview': 128000,
		'llama-3.2-90b-vision-preview': 128000,
	};

	// Get context limit for current model (default to 8192 if unknown)
	let currentContextLimit = $derived(() => {
		// Try exact match first
		if (modelContextLimits[selectedModel]) {
			return modelContextLimits[selectedModel];
		}
		// Try to match base model name (e.g., "llama3.2:latest" -> "llama3.2")
		const baseModel = selectedModel.split(':')[0];
		if (modelContextLimits[baseModel]) {
			return modelContextLimits[baseModel];
		}
		// Check if model name contains a known model
		for (const [key, limit] of Object.entries(modelContextLimits)) {
			if (selectedModel.toLowerCase().includes(key.toLowerCase())) {
				return limit;
			}
		}
		return 8192; // Conservative default
	});

	// Calculate message tokens
	let messageTokens = $derived(() => {
		return messages.reduce((total, msg) => {
			if (msg.usage?.total_tokens) {
				return total + msg.usage.total_tokens;
			}
			// Estimate tokens for messages without usage data (rough estimate: ~4 chars per token)
			return total + Math.ceil(msg.content.length / 4);
		}, 0);
	});

	// Calculate node context tokens (from active node)
	let nodeContextTokens = $derived(() => {
		if (!nodeContextPrompt) return 0;
		// Estimate: ~4 chars per token
		return Math.ceil(nodeContextPrompt.length / 4);
	});

	// Calculate selected context tokens (from word count, ~1.3 tokens per word)
	let contextDocTokens = $derived(() => {
		if (selectedContexts.length === 0) return 0;
		const totalWords = selectedContexts.reduce((sum, ctx) => sum + (Number(ctx.word_count) || 0), 0);
		return Math.ceil(totalWords * 1.3);
	});

	// Calculate total tokens used in conversation (including context)
	let totalConversationTokens = $derived(() => {
		return messageTokens() + nodeContextTokens() + contextDocTokens();
	});

	// Calculate context usage percentage
	let contextUsagePercent = $derived(() => {
		const limit = currentContextLimit();
		const used = totalConversationTokens();
		return Math.min(100, Math.round((used / limit) * 100));
	});

	// Format token count for display (e.g., 1.5K, 128K)
	function formatTokenCount(tokens: number): string {
		if (tokens >= 1000000) {
			return (tokens / 1000000).toFixed(1) + 'M';
		} else if (tokens >= 1000) {
			return (tokens / 1000).toFixed(tokens >= 10000 ? 0 : 1) + 'K';
		}
		return tokens.toString();
	}

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

	// Handle context toggle from Context panel
	function handleContextToggle(contextId: string, selected: boolean) {
		if (selected) {
			if (!selectedContextIds.includes(contextId)) {
				selectedContextIds = [...selectedContextIds, contextId];
			}
		} else {
			selectedContextIds = selectedContextIds.filter(id => id !== contextId);
		}
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

		// Remove [DELEGATE:AgentName] tags and orchestrator context lines
		cleanContent = cleanContent.replace(/\[DELEGATE:\w+\]\s*/gi, '');
		cleanContent = cleanContent.replace(/^(Task|Context|Orchestrator context):\s*.*$/gim, '');

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
		cleanContent = cleanContent.replace(/```artifact\s*\n[\s\S]*?\n```/g, '').trim();

		// Auto-detect document-like content and wrap as artifact
		// Detect if content looks like a structured document (has headings and sections)
		const hasHeadings = /^#{1,3}\s+.+$|^[IVX]+\.\s+.+$|^\d+\.\s+[A-Z]/m.test(cleanContent);
		const hasMultipleSections = (cleanContent.match(/^(?:#{1,3}\s+|[IVX]+\.\s+|\d+\.\s+[A-Z])/gm) || []).length >= 2;
		const isLongContent = cleanContent.length > 500;

		if (artifacts.length === 0 && hasHeadings && hasMultipleSections && isLongContent) {
			// Extract title from first heading or first line
			const titleMatch = cleanContent.match(/^#{1,3}\s+(.+)$/m) ||
			                   cleanContent.match(/^(?:Task:\s*)?(.+?)(?:\n|$)/);
			const title = titleMatch ? titleMatch[1].trim().replace(/^(Create|Write|Draft|Make)\s+(a\s+)?/i, '') : 'Generated Document';

			// Detect type from content
			let type = 'document';
			const lowerContent = cleanContent.toLowerCase();
			if (lowerContent.includes('standard operating procedure') || lowerContent.includes('sop')) type = 'sop';
			else if (lowerContent.includes('proposal')) type = 'proposal';
			else if (lowerContent.includes('framework')) type = 'framework';
			else if (lowerContent.includes('plan') || lowerContent.includes('roadmap')) type = 'plan';
			else if (lowerContent.includes('report') || lowerContent.includes('analysis')) type = 'report';

			artifacts.push({
				title: title.substring(0, 100),
				type,
				content: cleanContent
			});
			cleanContent = ''; // Hide the raw content since we wrapped it as artifact
		}

		return { cleanContent, artifacts };
	}

	async function selectConversation(id: string) {
		console.log('[selectConversation] Starting with id:', id);
		activeConversationId = id;
		conversationId = id;
		loadingConversation = true;
		artifactsLoadedOnce = false; // Reset so artifacts reload when panel opens

		// Load conversation messages from backend
		try {
			const conv = await api.getConversation(id);
			console.log('[selectConversation] Loaded conversation:', conv);
			console.log('[selectConversation] Messages count:', conv.messages?.length ?? 0);

			if (!conv.messages || !Array.isArray(conv.messages)) {
				console.error('[selectConversation] No messages array in response');
				messages = [];
				return;
			}

			messages = conv.messages.map(m => {
				console.log('[selectConversation] Processing message:', m.id, m.role);
				if (m.role === 'assistant') {
					// Parse artifacts from assistant messages
					const hasArtifactBlock = m.content?.includes('```artifact') ?? false;
					if (hasArtifactBlock) {
						console.log('[selectConversation] Content preview:', m.content.substring(0, 500));
					}

					const { cleanContent, artifacts } = parseArtifactsFromContent(m.content || '');

					return {
						id: String(m.id),
						role: m.role as 'user' | 'assistant',
						content: cleanContent,
						artifacts: artifacts.length > 0 ? artifacts : undefined
					};
				}
				return {
					id: String(m.id),
					role: m.role as 'user' | 'assistant',
					content: m.content || ''
				};
			});

			console.log('[selectConversation] Final messages state:', $state.snapshot(messages).length, 'messages');

			// Load artifacts for this conversation
			loadArtifacts();
		} catch (e) {
			console.error('[selectConversation] Failed to load conversation:', e);
			messages = [];
		} finally {
			loadingConversation = false;
		}
	}

	// File attachment handlers
	async function handleFileSelect(event: Event) {
		console.log('[Chat] handleFileSelect triggered');
		const input = event.target as HTMLInputElement;
		const files = input.files;
		if (!files) {
			console.log('[Chat] No files selected');
			return;
		}
		console.log('[Chat] Files selected:', files.length);

		for (const file of Array.from(files)) {
			// Check file size (max 50MB for documents)
			if (file.size > 50 * 1024 * 1024) {
				alert(`File "${file.name}" is too large. Maximum size is 50MB.`);
				continue;
			}

			const fileId = crypto.randomUUID();
			const newFile: AttachedFile = {
				id: fileId,
				name: file.name,
				type: file.type,
				size: file.size,
				uploading: true
			};

			// For images, convert to base64 for preview
			if (file.type.startsWith('image/')) {
				const reader = new FileReader();
				reader.onload = () => {
					newFile.content = reader.result as string;
				};
				reader.readAsDataURL(file);
			}

			// Add to UI immediately (with uploading state)
			attachedFiles = [...attachedFiles, newFile];
			console.log('[Chat] File added to attachedFiles:', fileId, 'total:', $state.snapshot(attachedFiles).length);

			// Upload document to backend for RAG processing
			try {
				console.log('[Chat] Starting upload for:', file.name, 'size:', file.size);
				const formData = new FormData();
				formData.append('file', file);
				formData.append('display_name', file.name);
				if (selectedProjectId) {
					formData.append('project_id', selectedProjectId);
				}

				console.log('[Chat] Sending POST to /api/documents');
				const response = await fetch('/api/documents', {
					method: 'POST',
					credentials: 'include',
					body: formData
				});

				if (!response.ok) {
					throw new Error(`Upload failed: ${response.status}`);
				}

				const doc = await response.json();
				console.log('[Chat] Document uploaded:', doc);

				// Update file with document ID
				attachedFiles = attachedFiles.map(f =>
					f.id === fileId
						? { ...f, documentId: doc.id, uploading: false }
						: f
				);
				console.log('[Chat] Document ID set:', doc.id, 'attachedFiles count:', $state.snapshot(attachedFiles).length, 'files:', $state.snapshot(attachedFiles).map(f => ({ id: f.id, documentId: f.documentId })));
			} catch (err) {
				console.error('[Chat] Document upload error:', err);
				// Update file with error state
				attachedFiles = attachedFiles.map(f =>
					f.id === fileId
						? { ...f, uploading: false, uploadError: err instanceof Error ? err.message : 'Upload failed' }
						: f
				);
			}
		}

		// Reset input
		if (fileInputRef) fileInputRef.value = '';
	}

	function removeAttachedFile(fileId: string) {
		attachedFiles = attachedFiles.filter(f => f.id !== fileId);
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	function startNewConversation() {
		console.log('[Chat] startNewConversation called - clearing attachedFiles');
		conversationId = null;
		activeConversationId = null;
		messages = [];
		attachedFiles = [];
		inputValue = '';
		showPlusMenu = false;
	}

	// Voice recording functions
	async function checkWhisperStatus() {
		try {
			const response = await apiClient.get('/transcribe/status');
			if (response.ok) {
				const data = await response.json();
				whisperAvailable = data.available;
			}
		} catch (error) {
			console.error('Failed to check whisper status:', error);
			whisperAvailable = false;
		}
	}

	async function toggleRecording() {
		if (isRecording) {
			stopRecording();
		} else {
			await startRecording();
		}
	}

	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];
			liveTranscript = '';

			// Set up audio analyzer for waveform visualization
			audioContext = new AudioContext();
			analyser = audioContext.createAnalyser();
			analyser.fftSize = 256;
			analyser.smoothingTimeConstant = 0.3;
			const source = audioContext.createMediaStreamSource(stream);
			source.connect(analyser);
			audioDataArray = new Uint8Array(analyser.fftSize);

			// Start waveform animation using time domain data (actual waveform)
			function updateWaveform() {
				if (!analyser || !audioDataArray) {
					animationFrameId = requestAnimationFrame(updateWaveform);
					return;
				}

				// Get time domain data (actual audio waveform, not frequency)
				analyser.getByteTimeDomainData(audioDataArray);

				// Sample 40 points from the waveform
				const bars: number[] = [];
				const step = Math.floor(audioDataArray.length / 40);
				for (let i = 0; i < 40; i++) {
					// Time domain data is centered at 128 (silence)
					// Values go from 0-255, with 128 being center
					const value = audioDataArray[i * step] || 128;
					// Calculate deviation from center (0-128 range)
					const deviation = Math.abs(value - 128);
					// Scale to 2-24 pixels height (amplify for visibility)
					const height = Math.max(2, Math.min(24, 2 + (deviation / 128) * 44));
					bars.push(height);
				}
				waveformBars = bars;
				animationFrameId = requestAnimationFrame(updateWaveform);
			}
			// Start animation immediately
			animationFrameId = requestAnimationFrame(updateWaveform);

			// Set up Web Speech API for live transcription
			const SpeechRecognitionAPI = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
			if (SpeechRecognitionAPI) {
				const recognition = new SpeechRecognitionAPI();
				recognition.continuous = true;
				recognition.interimResults = true;
				recognition.lang = 'en-US';

				recognition.onresult = (event: SpeechRecognitionEvent) => {
					let interimTranscript = '';
					let finalTranscript = '';

					for (let i = event.resultIndex; i < event.results.length; i++) {
						const transcript = event.results[i][0].transcript;
						if (event.results[i].isFinal) {
							finalTranscript += transcript;
						} else {
							interimTranscript += transcript;
						}
					}

					// Show interim results immediately
					liveTranscript = finalTranscript || interimTranscript;
				};

				recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
					console.log('Speech recognition error:', event.error);
				};

				recognition.start();
				speechRecognition = recognition;
			}

			mediaRecorder.ondataavailable = (event) => {
				audioChunks.push(event.data);
			};

			mediaRecorder.onstop = async () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				// Use whisper for final accurate transcription
				await transcribeAudio(audioBlob);
				// Stop all tracks
				stream.getTracks().forEach(track => track.stop());
			};

			mediaRecorder.start();
			isRecording = true;
			recordingDuration = 0;
			recordingInterval = setInterval(() => {
				recordingDuration++;
			}, 1000);
		} catch (error) {
			console.error('Failed to start recording:', error);
			alert('Could not access microphone. Please grant microphone permissions.');
		}
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
			isRecording = false;
		}
		if (recordingInterval) {
			clearInterval(recordingInterval);
			recordingInterval = null;
		}

		// Stop speech recognition
		if (speechRecognition) {
			speechRecognition.stop();
			speechRecognition = null;
		}

		// Stop waveform animation
		if (animationFrameId) {
			cancelAnimationFrame(animationFrameId);
			animationFrameId = null;
		}

		// Close audio context
		if (audioContext) {
			audioContext.close();
			audioContext = null;
			analyser = null;
			audioDataArray = null;
		}

		// Reset waveform
		waveformBars = Array(40).fill(2);
		recordingDuration = 0;
	}

	async function transcribeAudio(audioBlob: Blob) {
		isTranscribing = true;
		try {
			const formData = new FormData();
			formData.append('audio', audioBlob, 'recording.webm');

			const response = await fetch('/api/transcribe', {
				method: 'POST',
				credentials: 'include',
				body: formData
			});

			if (response.ok) {
				const data = await response.json();
				if (data.text) {
					// Append transcribed text to input
					inputValue = inputValue ? inputValue + ' ' + data.text : data.text;
				}
			} else {
				const error = await response.json();
				console.error('Transcription failed:', error.message);
				// Show error as system message
				messages = [...messages, {
					id: crypto.randomUUID(),
					role: 'assistant',
					content: `Voice transcription failed: ${error.message || 'Unknown error'}. Make sure whisper.cpp is installed.`
				}];
			}
		} catch (error) {
			console.error('Transcription error:', error);
			messages = [...messages, {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: 'Voice transcription requires whisper.cpp to be installed locally.'
			}];
		} finally {
			isTranscribing = false;
		}
	}

	// Interface for files coming from FocusModeSelector (without documentId)
	interface FocusModeFile {
		id: string;
		name: string;
		type: string;
		size: number;
		content?: string;
	}

	// Handle Focus Mode submission - accepts files from FocusModeSelector
	async function handleFocusModeSubmit(
		message: string,
		focusMode: string | null,
		options: Record<string, string>,
		files?: FocusModeFile[]
	) {
		console.log('[Chat] handleFocusModeSubmit called with files:', files?.length || 0);

		selectedFocusId = focusMode;
		focusOptions = options;
		inputValue = message;

		// If files were attached in FocusModeSelector, upload them first
		if (files && files.length > 0) {
			console.log('[Chat] Processing', files.length, 'files from FocusModeSelector');

			// Clear any existing attached files and process new ones
			attachedFiles = [];

			for (const file of files) {
				const newFile: AttachedFile = {
					id: file.id,
					name: file.name,
					type: file.type,
					size: file.size,
					content: file.content,
					uploading: true
				};

				// Add to UI immediately
				attachedFiles = [...attachedFiles, newFile];

				// Upload to document API for RAG processing
				try {
					console.log('[Chat] Uploading file from FocusMode:', file.name);

					// Convert content to blob for upload
					let blob: Blob;
					if (file.content && file.content.startsWith('data:')) {
						// Base64 content (images)
						const base64Data = file.content.split(',')[1];
						const binaryData = atob(base64Data);
						const bytes = new Uint8Array(binaryData.length);
						for (let i = 0; i < binaryData.length; i++) {
							bytes[i] = binaryData.charCodeAt(i);
						}
						blob = new Blob([bytes], { type: file.type });
					} else if (file.content) {
						// Text content
						blob = new Blob([file.content], { type: file.type || 'text/plain' });
					} else {
						// No content available, skip upload but keep in list
						console.log('[Chat] No content for file, skipping upload:', file.name);
						attachedFiles = attachedFiles.map(f =>
							f.id === file.id ? { ...f, uploading: false } : f
						);
						continue;
					}

					const formData = new FormData();
					formData.append('file', blob, file.name);
					formData.append('display_name', file.name);
					if (selectedProjectId) {
						formData.append('project_id', selectedProjectId);
					}

					const response = await fetch('/api/documents', {
						method: 'POST',
						credentials: 'include',
						body: formData
					});

					if (!response.ok) {
						throw new Error(`Upload failed: ${response.status}`);
					}

					const doc = await response.json();
					console.log('[Chat] FocusMode document uploaded:', doc.id);

					// Update with document ID
					attachedFiles = attachedFiles.map(f =>
						f.id === file.id ? { ...f, documentId: doc.id, uploading: false } : f
					);
				} catch (err) {
					console.error('[Chat] FocusMode document upload error:', err);
					attachedFiles = attachedFiles.map(f =>
						f.id === file.id
							? { ...f, uploading: false, uploadError: err instanceof Error ? err.message : 'Upload failed' }
							: f
					);
				}
			}

			console.log('[Chat] All FocusMode files processed. attachedFiles:', $state.snapshot(attachedFiles).length,
				'with documentIds:', $state.snapshot(attachedFiles).filter(f => f.documentId).length);
		}

		handleSendMessage();
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

		// Reset thinking state for new message
		currentThinking = '';
		hasThinking = false;

		// Parse slash commands (e.g., "/analyze what are the key trends?")
		let command: string | undefined;
		let messageContent = userMessage;
		const commandMatch = userMessage.match(/^\/(\w+)(?:\s+(.*))?$/s);
		if (commandMatch) {
			command = commandMatch[1]; // e.g., "analyze", "summarize"
			messageContent = commandMatch[2]?.trim() || userMessage; // The rest of the message
			console.log('[Chat] Slash command detected:', command, 'Message:', messageContent);
		}

		// Add user message to UI
		const userMsgId = crypto.randomUUID();
		messages = [...messages, { id: userMsgId, role: 'user', content: userMessage }];

		// Create assistant message placeholder
		const assistantMsgId = crypto.randomUUID();
		messages = [...messages, { id: assistantMsgId, role: 'assistant', content: '' }];

		isStreaming = true;
		abortController = new AbortController();

		try {
			// Debug: Log focus mode state before building request
			console.log('[Chat] Focus Mode Debug:', { selectedFocusId: $state.snapshot(selectedFocusId), focusOptions: $state.snapshot(focusOptions), focusModeEnabled: $state.snapshot(focusModeEnabled) });

			// Build request body with context and node context
			// Note: The backend will load full context details (content, system_prompt_template)
			// using the context_id, so we just pass the ID here
			const requestBody: Record<string, unknown> = {
				message: messageContent,
				model: selectedModel,
				conversation_id: conversationId,
				project_id: selectedProjectId,
				workspace_id: $currentWorkspaceId, // Include workspace for role-based agent context
				context_id: selectedContextIds.length > 0 ? selectedContextIds[0] : null,
				context_ids: selectedContextIds.length > 0 ? selectedContextIds : undefined,
				command: command, // Send slash command to backend
				// AI model settings from user preferences
				temperature: aiTemperature,
				max_tokens: aiMaxTokens,
				top_p: aiTopP,
				// Focus Mode parameters
				focus_mode: selectedFocusId,
				focus_options: Object.keys(focusOptions).length > 0 ? focusOptions : undefined,
				structured_output: true, // Enable V2 structured output
				// Custom Agent
				agent_id: selectedAgent?.id || undefined,
			};

			// Include attached document IDs for context injection
			console.log('[Chat] Attached files at send time:', $state.snapshot(attachedFiles).length, $state.snapshot(attachedFiles));
			const uploadedDocIds = attachedFiles
				.filter(f => f.documentId && !f.uploadError)
				.map(f => f.documentId);
			console.log('[Chat] Filtered document IDs:', uploadedDocIds);
			if (uploadedDocIds.length > 0) {
				requestBody.document_ids = uploadedDocIds;
				console.log('[Chat] Including document IDs in request:', uploadedDocIds);
			} else {
				console.log('[Chat] No document IDs to include');
			}

			// Clear attached files after including in request
			attachedFiles = [];

			// Include node context if there's an active node
			if (nodeContextPrompt) {
				requestBody.node_context = nodeContextPrompt;
			}

			// Add COT flag if enabled
			if (useCOT) {
				requestBody.use_cot = true;
			}

			// Use v2 endpoint for COT mode (SSE with thinking events)
			const endpoint = useCOT ? '/api/chat/v2/message' : '/api/chat/message';
			const response = await fetch(endpoint, {
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
				let sseBuffer = ''; // Buffer for SSE parsing

				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					const chunk = decoder.decode(value, { stream: true });

					// Always parse SSE events (V2 endpoint always sends SSE format)
					sseBuffer += chunk;
					const lines = sseBuffer.split('\n');
					sseBuffer = lines.pop() || ''; // Keep incomplete line in buffer

					for (const line of lines) {
						if (line.startsWith('data: ')) {
							try {
								const data = JSON.parse(line.slice(6));
								// Handle thinking events (generic COT from backend)
								// These include search events with step: "searching" or "search_complete"
								if (data.type === 'thinking' && data.data) {
									const thinkingData = data.data;
									hasThinking = true;
									thinkingExpanded = true;
									// ThinkingStep struct: { step, content, agent, completed }
									if (thinkingData.content) {
										// Mark search-related thinking with a special prefix for preservation
										if (thinkingData.step === 'searching' || thinkingData.step === 'search_complete') {
											currentThinking += thinkingData.content + '\n';
											console.log('[COT] Search event:', thinkingData);
										} else {
											currentThinking += thinkingData.content + '\n';
										}
									}
									if (typeof thinkingData === 'string') {
										currentThinking += thinkingData + '\n';
									}
									console.log('[COT] Thinking event:', thinkingData);
								}
								// Handle thinking_start events (Chain of Thought)
								if (data.type === 'thinking_start') {
									hasThinking = true;
									// Preserve search-related content
									const searchContent = currentThinking.split('\n')
										.filter(line => line.includes('Searching the web') || line.includes('Found') && line.includes('sources'))
										.join('\n');
									currentThinking = searchContent ? searchContent + '\n' : '';
									thinkingExpanded = true;
									console.log('[COT] Thinking started (preserved search):', searchContent ? 'yes' : 'no');
								}
								// Handle thinking_chunk events (Chain of Thought)
								if (data.type === 'thinking_chunk' && data.data) {
									const thinkingData = data.data;
									if (thinkingData.content) {
										currentThinking += thinkingData.content;
										hasThinking = true;
									}
								}
								// Handle thinking_end events (Chain of Thought)
								if (data.type === 'thinking_end' && data.data) {
									console.log('[COT] Thinking completed:', data.data);
									if (data.data.content && !currentThinking) {
										currentThinking = data.data.content;
									}
								}
								// Only process token events for display
								if (data.type === 'token') {
									const tokenContent = data.content || data.data || '';
									if (tokenContent && !tokenContent.includes('Chain of Thought Summary')) {
										fullContent += tokenContent;
									}
								}
								// Handle artifact_complete events
								if (data.type === 'artifact_complete' && data.data) {
									const artifact = data.data;
									if (artifact.title && artifact.type && artifact.content) {
										viewingArtifactFromMessage = {
											title: artifact.title,
											type: artifact.type,
											content: artifact.content
										};
										artifactsPanelOpen = true;
										rightPanelOpen = true;
										rightPanelTab = 'artifacts';
										generatingArtifact = false;
										artifactCompletedInStream = true;
										autoSaveArtifact(artifact);
									}
								}
								// Handle blocks events (structured output)
								if (data.type === 'blocks' && data.data) {
									const blockDoc = data.data;
									if (blockDoc.blocks) {
										messages = messages.map(msg =>
											msg.id === assistantMsgId
												? { ...msg, blocks: blockDoc.blocks }
												: msg
										);
									}
								}
							} catch {
								// Ignore parse errors
							}
						}
					}

					// Track if we're inside artifact block to filter it from display
					if (fullContent.includes('```artifact') && !artifactStarted) {
						artifactStarted = true;
						inArtifactBlock = true;
						generatingArtifact = true;
						artifactsPanelOpen = true;
						rightPanelOpen = true;
						rightPanelTab = 'artifacts';

						// Set panel to 50% width when artifact is generated
						const availableWidth = window.innerWidth - 256; // Subtract left sidebar
						artifactPanelWidth = Math.floor(availableWidth / 2);

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

							// Parse artifact for viewing and auto-save
							try {
								const artifactMatch = fullContent.match(/```artifact\s*\n([\s\S]*?)\n```/);
								if (artifactMatch) {
									const artifactData = JSON.parse(artifactMatch[1].trim());
									if (artifactData.title && artifactData.type && artifactData.content) {
										const processedContent = artifactData.content
											.replace(/\\n/g, '\n')
											.replace(/\\"/g, '"')
											.replace(/\\\\/g, '\\');

										viewingArtifactFromMessage = {
											title: artifactData.title,
											type: artifactData.type,
											content: processedContent
										};
										artifactsPanelOpen = true;
										rightPanelOpen = true;
										rightPanelTab = 'artifacts';
										generatingArtifactTitle = artifactData.title;
										generatingArtifactType = artifactData.type;

										// Auto-save the artifact to database
										autoSaveArtifact({
											title: artifactData.title,
											type: artifactData.type,
											content: processedContent
										});
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

			// Persist thinking content by appending it to the message content
			if (currentThinking && hasThinking) {
				const thinkingBlock = `<thinking>${currentThinking}</thinking>\n\n`;
				// If artifactStarted, we update displayContent
				if (artifactStarted) {
					displayContent = thinkingBlock + displayContent;
				}
				// Always update fullContent
				fullContent = thinkingBlock + fullContent;

				// Update message immediately with persistent thinking
				messages = messages.map(msg =>
					msg.id === assistantMsgId
						? {
							...msg,
							content: artifactStarted ? displayContent : fullContent
						}
						: msg
				);
			}

			// Parse and extract usage data from the response
			let usageData: UsageData | undefined;
			// Flexible regex that matches any valid JSON object between usage markers
			const usageRegex = /<!--USAGE:(\{[^}]+\})-->/;
			const usageMatch = fullContent.match(usageRegex);
			console.log('[Chat] Full content length:', fullContent.length);
			console.log('[Chat] Last 300 chars:', fullContent.slice(-300));
			console.log('[Chat] Usage match result:', usageMatch ? 'Found' : 'Not found');
			if (usageMatch) {
				try {
					console.log('[Chat] Raw usage JSON:', usageMatch[1]);
					usageData = JSON.parse(usageMatch[1]);
					console.log('[Chat] Parsed usage data:', usageData);
					// Remove usage comment from displayed content
					fullContent = fullContent.replace(usageRegex, '').trim();
					displayContent = displayContent.replace(usageRegex, '').trim();

					// Update message with usage data - ensure this persists
					messages = messages.map(msg =>
						msg.id === assistantMsgId
							? { ...msg, content: artifactStarted ? displayContent : fullContent, usage: usageData }
							: msg
					);
					console.log('[Chat] Updated message with usage:', $state.snapshot(messages).find(m => m.id === assistantMsgId)?.usage);
				} catch (e) {
					console.error('Failed to parse usage data:', e, usageMatch[1]);
				}
			} else {
				console.log('[Chat] No usage match found. Looking for USAGE marker...');
				console.log('[Chat] Contains USAGE marker:', fullContent.includes('<!--USAGE:'));
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

			// Mark this model as warmed up (for Ollama local models)
			if (activeProvider === 'ollama_local' && selectedModel && !warmedUpModels.has(selectedModel)) {
				const newSet = new Set(warmedUpModels);
				newSet.add(selectedModel);
				warmedUpModels = newSet;
				// Persist to localStorage
				if (typeof window !== 'undefined') {
					localStorage.setItem('warmedUpModels', JSON.stringify([...newSet]));
				}
			}
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
				const hasArtifactStart = remainingText.includes('```artifact');
				const hasCompleteArtifactBlock = /```artifact\s*\n[\s\S]*?\n```/.test(remainingText);
				if (hasArtifactStart && !hasCompleteArtifactBlock) {
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

	// Right panel resize handlers
	function startRightPanelResize(e: MouseEvent) {
		isResizingRightPanel = true;
		rightPanelResizeStartX = e.clientX;
		rightPanelResizeStartWidth = rightPanelWidth;
		document.addEventListener('mousemove', handleRightPanelResize);
		document.addEventListener('mouseup', stopRightPanelResize);
		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';
	}

	function handleRightPanelResize(e: MouseEvent) {
		if (!isResizingRightPanel) return;
		const delta = rightPanelResizeStartX - e.clientX;
		const newWidth = Math.min(Math.max(rightPanelResizeStartWidth + delta, 280), 500);
		rightPanelWidth = newWidth;
	}

	function stopRightPanelResize() {
		isResizingRightPanel = false;
		document.removeEventListener('mousemove', handleRightPanelResize);
		document.removeEventListener('mouseup', stopRightPanelResize);
		document.body.style.cursor = '';
		document.body.style.userSelect = '';
	}

	function viewArtifactInPanel(artifact: { title: string; type: string; content: string }) {
		viewingArtifactFromMessage = artifact;
		selectedArtifact = null;
		artifactsPanelOpen = true;
	}

	// Inline project selector dropdown (appears in chat input when Enter pressed without project)
	let showInlineProjectPicker = $state(false);

	// Auto-scroll command item into view
	function scrollCommandIntoView(index: number) {
		setTimeout(() => {
			const item = document.querySelector(`[data-command-index="${index}"]`);
			item?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
		}, 0);
	}

	// Toggle COT mode and persist to thinking settings
	async function toggleCOT() {
		const newValue = !useCOT;
		useCOT = newValue;

		// Persist to backend
		try {
			await thinking.updateSettings({
				enabled: newValue,
				show_in_ui: $thinking.settings?.show_in_ui ?? false,
				save_traces: $thinking.settings?.save_traces ?? true,
				max_tokens: $thinking.settings?.max_tokens ?? 4096,
				default_template_id: $thinking.settings?.default_template_id ?? null
			});
		} catch (error) {
			console.error('[Chat] Failed to update thinking settings:', error);
			// Revert on error
			useCOT = !newValue;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		// Handle command suggestions navigation
		if (showCommandSuggestions && filteredCommands.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				commandDropdownIndex = (commandDropdownIndex + 1) % filteredCommands.length;
				scrollCommandIntoView(commandDropdownIndex);
			} else if (e.key === 'ArrowUp') {
				e.preventDefault();
				commandDropdownIndex = commandDropdownIndex <= 0 ? filteredCommands.length - 1 : commandDropdownIndex - 1;
				scrollCommandIntoView(commandDropdownIndex);
			} else if (e.key === 'Enter' || e.key === 'Tab') {
				e.preventDefault();
				const cmd = filteredCommands[commandDropdownIndex];
				if (cmd) {
					selectCommand(cmd);
				}
			} else if (e.key === 'Escape') {
				e.preventDefault();
				showCommandSuggestions = false;
			}
			return;
		}

		// Handle inline project picker navigation
		if (showInlineProjectPicker) {
			const totalItems = projectsList.length + 1; // +1 for create new
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				projectDropdownIndex = (projectDropdownIndex + 1) % totalItems;
			} else if (e.key === 'ArrowUp') {
				e.preventDefault();
				projectDropdownIndex = projectDropdownIndex <= 0 ? totalItems - 1 : projectDropdownIndex - 1;
			} else if (e.key === 'Enter') {
				e.preventDefault();
				if (projectDropdownIndex === projectsList.length) {
					showInlineProjectPicker = false;
					showNewProjectModal = true;
				} else {
					const project = projectsList[projectDropdownIndex];
					if (project) {
						selectedProjectId = project.id;
					}
				}
				showInlineProjectPicker = false;
				// User will press Enter again to send
			} else if (e.key === 'Escape') {
				e.preventDefault();
				showInlineProjectPicker = false;
			}
			return;
		}

		// Don't send if any dropdown is open
		if (showProjectDropdown || showContextDropdown || showModelDropdown) {
			return;
		}

		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			// If no project selected and there's input, show inline project picker
			if (!selectedProjectId && inputValue.trim()) {
				projectDropdownIndex = 0;
				showInlineProjectPicker = true;
				return;
			}
			handleSendMessage();
		}
	}

	// Select a command from the dropdown
	function selectCommand(cmd: SlashCommand) {
		activeCommand = cmd;
		inputValue = '/' + cmd.name + ' ';
		showCommandSuggestions = false;
		inputRef?.focus();
	}

	// Clear the active command
	function clearActiveCommand() {
		activeCommand = null;
		inputValue = '';
		inputRef?.focus();
	}

	// Select an agent from the dropdown
	function selectAgent(agent: AgentPreset) {
		detectedAgent = agent;
		inputValue = '@' + agent.name + ' ';
		showAgentSuggestions = false;
		inputRef?.focus();
	}

	// Clear the detected agent
	function clearDetectedAgent() {
		detectedAgent = null;
		inputValue = '';
		inputRef?.focus();
	}

	function handleInput() {
		if (inputRef) {
			inputRef.style.height = 'auto';
			inputRef.style.height = Math.min(inputRef.scrollHeight, 200) + 'px';
		}

		// Check for @agent mention input
		if (inputValue.startsWith('@')) {
			const query = inputValue.slice(1).toLowerCase().split(' ')[0]; // Get text after @ before space

			// Check if we have a complete agent mention (has space after agent name)
			const spaceIndex = inputValue.indexOf(' ');
			if (spaceIndex > 0) {
				const agentName = inputValue.slice(1, spaceIndex);
				const matchedAgent = availableAgents.find(a => a.name.toLowerCase() === agentName.toLowerCase());
				if (matchedAgent) {
					detectedAgent = matchedAgent;
					showAgentSuggestions = false;
					showCommandSuggestions = false;
					return;
				}
			}

			// Still typing agent name - show suggestions
			if (query.length === 0) {
				// Just "@" typed, show all agents
				filteredAgents = availableAgents.slice(0, 8);
			} else {
				// Filter agents by query
				filteredAgents = availableAgents
					.filter(agent => agent.name.toLowerCase().includes(query) || agent.display_name.toLowerCase().includes(query))
					.slice(0, 8);
			}
			showAgentSuggestions = filteredAgents.length > 0;
			agentDropdownIndex = 0;
			showCommandSuggestions = false;
		}
		// Check for slash command input
		else if (inputValue.startsWith('/')) {
			const query = inputValue.slice(1).toLowerCase().split(' ')[0]; // Get text after / before space

			// Check if we have a complete command (has space after command name)
			const spaceIndex = inputValue.indexOf(' ');
			if (spaceIndex > 0) {
				const cmdName = inputValue.slice(1, spaceIndex);
				const matchedCmd = availableCommands.find(c => c.name === cmdName);
				if (matchedCmd) {
					activeCommand = matchedCmd;
					showCommandSuggestions = false;
					showAgentSuggestions = false;
					return;
				}
			}

			// Still typing command - show suggestions
			if (query.length === 0) {
				// Just "/" typed, show all commands
				filteredCommands = availableCommands.slice(0, 8);
			} else {
				// Filter commands by query
				filteredCommands = availableCommands
					.filter(cmd => cmd.name.includes(query) || cmd.display_name.toLowerCase().includes(query))
					.slice(0, 8);
			}
			showCommandSuggestions = filteredCommands.length > 0;
			commandDropdownIndex = 0;
			showAgentSuggestions = false;
		} else {
			showCommandSuggestions = false;
			showAgentSuggestions = false;
			activeCommand = null;
			detectedAgent = null;
		}
	}

	// Track if we should send after project selection
	let pendingMessageAfterProject = $state(false);

	function handleProjectDropdownKeydown(e: KeyboardEvent) {
		if (!showProjectDropdown) return;

		// Total items = projects + 1 (for "create new" option)
		const totalItems = projectsList.length + 1;
		const createNewIndex = projectsList.length; // Last item is "create new"

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			projectDropdownIndex = (projectDropdownIndex + 1) % totalItems;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			projectDropdownIndex = projectDropdownIndex <= 0 ? totalItems - 1 : projectDropdownIndex - 1;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (projectDropdownIndex === createNewIndex) {
				// Create new project
				showProjectDropdown = false;
				showNewProjectModal = true;
			} else {
				const project = projectsList[projectDropdownIndex];
				if (project) {
					selectedProjectId = project.id;
					showProjectDropdown = false;
					// If there's pending input, send the message
					if (inputValue.trim()) {
						setTimeout(() => handleSendMessage(), 50);
					}
				}
			}
		} else if (e.key === 'Escape') {
			e.preventDefault();
			showProjectDropdown = false;
		}
	}

	function handleStop() {
		if (abortController) {
			abortController.abort();
		}
	}

	async function copyToClipboard(text: string) {
		try {
			if (navigator.clipboard?.writeText) {
				await navigator.clipboard.writeText(text);
				return;
			}
		} catch {
			// Fall back below
		}

		// Fallback for environments where Clipboard API is unavailable or blocked
		const textarea = document.createElement('textarea');
		textarea.value = text;
		textarea.style.position = 'fixed';
		textarea.style.top = '0';
		textarea.style.left = '0';
		textarea.style.opacity = '0';
		document.body.appendChild(textarea);
		textarea.focus();
		textarea.select();
		try {
			document.execCommand('copy');
		} finally {
			document.body.removeChild(textarea);
		}
	}

	function copyMessage(content: string, id: string) {
		copyToClipboard(content);
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

<svelte:window onkeydown={handleProjectDropdownKeydown} />

<!-- Fixed height container that fills parent -->
<div class="h-full flex overflow-hidden">
	<!-- Chat Conversations Sidebar -->
	{#if chatSidebarOpen}
		<div class="w-72 h-full flex-shrink-0 border-r border-gray-200" transition:fly={{ x: -288, duration: 200 }}>
			<ConversationListPanel
				conversations={conversations}
				archivedConversations={archivedConversations}
				activeId={activeConversationId ?? ''}
				showArchived={showArchivedView}
				bind:searchQuery
				onSelect={selectConversation}
				onNewChat={handleNewChat}
				onRename={handleRenameConversation}
				onPin={handlePinConversation}
				onDelete={handleDeleteConversation}
				onArchive={handleArchiveConversation}
				onUnarchive={handleUnarchiveConversation}
				onExport={handleExportConversation}
				onLinkProject={handleLinkProject}
				onViewArchived={handleViewArchived}
				onBackToChats={handleBackToChats}
			/>
		</div>
	{/if}

	<!-- Main Chat Area - fills remaining space (or 50% when artifact is focused) -->
	<div class="{artifactsPanelOpen && isArtifactFocused ? 'w-1/2' : 'flex-1'} flex flex-col min-w-0 h-full bg-gray-50">
		<!-- Toggle button - fixed header -->
		<div class="h-12 flex items-center justify-between px-4 flex-shrink-0 border-b border-gray-100 min-w-0">
			<!-- Left group: Hamburger + Model Selector -->
			<div class="flex items-center gap-1 flex-shrink-0">
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

				<!-- Model Selector (directly next to hamburger) -->
				<div class="relative">
					<button
						onclick={() => showModelDropdown = !showModelDropdown}
						class="flex items-center gap-1.5 px-2 py-1.5 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
						title="Select AI Model"
					>
						{#if warmingUpModel === selectedModel}
							<svg class="w-3.5 h-3.5 animate-spin text-orange-500 flex-shrink-0" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
						{/if}
						<span class="truncate max-w-[140px]">{currentModelName || 'Select model'}</span>
						{#if warmingUpModel === selectedModel}
							<span class="text-xs text-orange-500">warming...</span>
						{:else}
							<svg class="w-3 h-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						{/if}
					</button>

					{#if showModelDropdown}
						<div
							class="absolute left-0 top-full mt-2 w-72 bg-white border border-gray-200 rounded-xl shadow-lg py-2 z-30 max-h-96 overflow-y-auto"
							transition:fly={{ y: -10, duration: 200 }}
						>
							{#if loadingModels}
								<div class="px-4 py-3 text-sm text-gray-500 text-center">Loading models...</div>
							{:else if installedModels.length === 0 && ollamaCloudModels.length === 0 && configuredProviders.size === 0}
								<div class="px-4 py-3 text-center">
									<p class="text-sm text-gray-500 mb-2">No models available</p>
									<a href="/settings/ai" class="text-xs text-blue-600 hover:underline">Configure in AI Settings</a>
								</div>
							{:else}
								<!-- Selected Model at Top -->
								{@const selectedModelObj = models.find(m => m.id === selectedModel)}
								{#if selectedModelObj}
									<div class="px-3 py-1.5">
										<span class="text-xs font-semibold text-blue-500 uppercase tracking-wider flex items-center gap-1">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
											Selected
										</span>
									</div>
									{@const caps = selectedModelObj.capabilities || getModelCapabilities(selectedModelObj.id)}
									{@const isCloud = selectedModelObj.type === 'cloud' || selectedModelObj.id.toLowerCase().includes('-cloud')}
									<button
										onclick={() => { showModelDropdown = false; }}
										class="w-full px-4 py-2.5 text-left bg-blue-50 transition-colors"
									>
										<div class="flex items-start gap-2">
											<svg class="w-4 h-4 {isCloud ? 'text-blue-500' : 'text-green-500'} mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												{#if isCloud}
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
												{:else}
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
												{/if}
											</svg>
											<div class="flex-1 min-w-0">
												<div class="flex items-center gap-1.5 flex-wrap">
													<span class="text-sm font-medium text-blue-600">{selectedModelObj.name}</span>
													<span class="text-[10px] px-1.5 py-0.5 rounded {isCloud ? 'bg-blue-100 text-blue-600' : 'bg-green-100 text-green-600'}">{isCloud ? 'Cloud' : 'Local'}</span>
												</div>
												{#if selectedModelObj.size}
													<div class="text-xs text-gray-400 mt-0.5">{selectedModelObj.size}</div>
												{/if}
												{#if caps.length > 0}
													<div class="flex flex-wrap gap-1 mt-1">
														{#each caps.slice(0, 4) as cap}
															<span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[9px] font-medium rounded {capabilityInfo[cap].color}">
																<span>{capabilityInfo[cap].icon}</span>
																<span>{capabilityInfo[cap].label}</span>
															</span>
														{/each}
													</div>
												{/if}
											</div>
											<svg class="w-4 h-4 text-blue-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
										</div>
									</button>
								{/if}

								<!-- Cloud (Ollama Remote) Section -->
								{#if ollamaCloudModels.length > 0}
									<div class="px-3 py-1.5 border-t border-gray-100 mt-1 pt-1">
										<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider flex items-center gap-1">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
											</svg>
											Cloud (Ollama Remote)
										</span>
									</div>
									{#each ollamaCloudModels.filter(m => m.id !== selectedModel) as model}
									{@const caps = model.capabilities || getModelCapabilities(model.id)}
										<button
											onclick={() => { selectModel(model.id); showModelDropdown = false; }}
											class="w-full px-4 py-2.5 text-left hover:bg-gray-50 transition-colors"
										>
											<div class="flex items-start gap-2">
												<svg class="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
												</svg>
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-1.5 flex-wrap">
														<span class="text-sm font-medium text-gray-700">{model.name}</span>
													</div>
													{#if model.size}
														<div class="text-xs text-gray-400 mt-0.5">{model.size}</div>
													{/if}
													{#if caps.length > 0}
														<div class="flex flex-wrap gap-1 mt-1">
															{#each caps.slice(0, 4) as cap}
																<span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[9px] font-medium rounded {capabilityInfo[cap].color}">
																	<span>{capabilityInfo[cap].icon}</span>
																	<span>{capabilityInfo[cap].label}</span>
																</span>
															{/each}
															{#if caps.length > 4}
																<span class="text-[9px] text-gray-400">+{caps.length - 4}</span>
															{/if}
														</div>
													{/if}
												</div>
											</div>
										</button>
									{/each}
								{/if}

								<!-- Local Models Section -->
								{#if installedModels.length > 0}
									<div class="px-3 py-1.5 {ollamaCloudModels.length > 0 ? 'border-t border-gray-100 mt-1 pt-1' : ''}">
										<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider flex items-center gap-1">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
											</svg>
											Local (Ollama)
										</span>
									</div>
									{#each installedModels.filter(m => m.id !== selectedModel) as model}
									{@const caps = model.capabilities || getModelCapabilities(model.id)}
										<button
											onclick={() => { selectModel(model.id); showModelDropdown = false; }}
											class="w-full px-4 py-2.5 text-left hover:bg-gray-50 transition-colors"
										>
											<div class="flex items-start gap-2">
												<svg class="w-4 h-4 text-green-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
												</svg>
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-1.5 flex-wrap">
														<span class="text-sm font-medium text-gray-700">{model.name}</span>
													</div>
													{#if model.size}
														<div class="text-xs text-gray-400 mt-0.5">{model.size}</div>
													{/if}
													{#if caps.length > 0}
														<div class="flex flex-wrap gap-1 mt-1">
															{#each caps.slice(0, 4) as cap}
																<span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[9px] font-medium rounded {capabilityInfo[cap].color}">
																	<span>{capabilityInfo[cap].icon}</span>
																	<span>{capabilityInfo[cap].label}</span>
																</span>
															{/each}
															{#if caps.length > 4}
																<span class="text-[9px] text-gray-400">+{caps.length - 4}</span>
															{/if}
														</div>
													{/if}
												</div>
											</div>
										</button>
									{/each}
								{/if}

								<!-- Cloud Models by Provider -->
								{#each Array.from(configuredProviders) as provider}
									{@const providerModels = cloudModelsMap[provider] || []}
									{#if providerModels.length > 0}
										<div class="px-3 py-1.5 {installedModels.length > 0 || Array.from(configuredProviders).indexOf(provider) > 0 ? 'border-t border-gray-100 mt-1 pt-1' : ''}">
											<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider flex items-center gap-1">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
												</svg>
												{provider.charAt(0).toUpperCase() + provider.slice(1)}
											</span>
										</div>
										{#each providerModels as model}
											{@const caps = model.capabilities || []}
											<button
												onclick={() => { selectModel(model.id); showModelDropdown = false; }}
												class="w-full px-4 py-2.5 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
											>
												<div class="flex items-start gap-2">
													<svg class="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
													</svg>
													<div class="flex-1 min-w-0">
														<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
														{#if model.description}
															<div class="text-xs text-gray-400 mt-0.5">{model.description}</div>
														{/if}
														<!-- Capability badges -->
														{#if caps.length > 0}
															<div class="flex flex-wrap gap-1 mt-1">
																{#each caps.slice(0, 4) as cap}
																	<span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[9px] font-medium rounded {capabilityInfo[cap].color}">
																		<span>{capabilityInfo[cap].icon}</span>
																		<span>{capabilityInfo[cap].label}</span>
																	</span>
																{/each}
																{#if caps.length > 4}
																	<span class="text-[9px] text-gray-400">+{caps.length - 4}</span>
																{/if}
															</div>
														{/if}
													</div>
													{#if selectedModel === model.id}
														<svg class="w-4 h-4 text-blue-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
														</svg>
													{/if}
												</div>
											</button>
										{/each}
									{/if}
								{/each}

								<!-- Settings Link -->
								<div class="border-t border-gray-100 mt-1 pt-1">
									<a
										href="/settings/ai"
										onclick={() => showModelDropdown = false}
										class="w-full px-4 py-2 text-left text-sm text-gray-500 hover:bg-gray-50 transition-colors flex items-center gap-2"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										</svg>
										AI Settings
									</a>
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<!-- COT is always enabled - no toggle needed -->
			</div>

			<!-- Center: Role Context Badge -->
			<div class="flex items-center justify-center flex-1 min-w-0">
				<RoleContextBadge size="sm" showLabel={true} showTooltip={true} />
			</div>

			<!-- Right group: Agent, Project, Node, Panel -->
			<div class="flex items-center gap-2 min-w-0">
				<!-- Agent Selector -->
				<div class="relative flex-shrink-0" style="width: 220px;">
					<AgentSelector
						agents={customAgents}
						selectedId={selectedAgent?.id || null}
						onSelect={handleAgentSelect}
						placeholder="Default Agent"
						includeDefault={true}
						onManage={() => window.location.href = '/agents'}
					/>
				</div>

				<!-- Project Selector (required for chat) -->
				<div class="relative flex-shrink-0">
					<button
						onclick={() => {
							if (!showProjectDropdown) projectDropdownIndex = 0;
							showProjectDropdown = !showProjectDropdown;
							showHeaderContextDropdown = false;
							showNodeDropdown = false;
						}}
						onkeydown={handleProjectDropdownKeydown}
						class="header-toggle-btn {selectedProject ? 'active' : 'warning'}"
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
							onkeydown={handleProjectDropdownKeydown}
							tabindex="-1"
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
								{#each projectsList as project, i (project.id)}
									{@const isSelected = selectedProjectId === project.id}
									{@const isFocused = projectDropdownIndex === i}
									<button
										onclick={() => {
											selectedProjectId = project.id;
											showProjectDropdown = false;
											if (inputValue.trim()) setTimeout(() => handleSendMessage(), 50);
										}}
										class="w-full px-4 py-2 text-left transition-colors flex items-center gap-3 {isSelected ? 'bg-purple-50' : ''} {isFocused ? 'bg-blue-50 ring-2 ring-blue-400 ring-inset' : 'hover:bg-gray-50'}"
									>
										<div class="w-8 h-8 rounded-lg {isSelected ? 'bg-purple-500 text-white' : isFocused ? 'bg-blue-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
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
											<div class="text-sm font-medium {isSelected ? 'text-purple-600' : isFocused ? 'text-blue-600' : 'text-gray-700'} truncate">{project.name}</div>
											{#if project.description}
												<div class="text-xs text-gray-500 truncate">{project.description}</div>
											{/if}
										</div>
									</button>
								{/each}
							{/if}
							<!-- Create New Project Option -->
							<div class="border-t border-gray-100 mt-1 pt-1">
								<button
									onclick={() => { showProjectDropdown = false; showNewProjectModal = true; }}
									class="w-full px-4 py-2 text-left transition-colors flex items-center gap-3 {projectDropdownIndex === projectsList.length ? 'bg-gray-100 ring-2 ring-gray-400 ring-inset' : 'hover:bg-gray-50'}"
								>
									<div class="w-8 h-8 rounded-lg {projectDropdownIndex === projectsList.length ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600'} flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium {projectDropdownIndex === projectsList.length ? 'text-gray-900' : 'text-gray-700'}">Create new project</div>
										<div class="text-xs text-gray-500">Start a new project for this chat</div>
									</div>
								</button>
							</div>
						</div>
					{/if}
				</div>

				<!-- Active Node Indicator -->
				{#if activeNode}
					<div class="relative flex-shrink-0">
						<button
							onclick={() => showNodeDropdown = !showNodeDropdown}
							class="header-toggle-btn active"
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
						class="header-toggle-btn"
					>
						<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
						<span>MIOSA Platform</span>
					</a>
				{/if}

				<!-- Panel Toggle (combines Progress, Context, Artifacts) -->
				<button
					onclick={() => rightPanelOpen = !rightPanelOpen}
					class="header-toggle-btn {rightPanelOpen ? 'active' : ''}"
					title="Toggle Side Panel"
				>
					<svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
					</svg>
					<span>Panel</span>
					{#if artifacts.length > 0}
						<span class="header-toggle-badge">{artifacts.length}</span>
					{/if}
				</button>
			</div>
		</div>

		{#if hasConversation}
			<!-- Messages container - scrollable, takes remaining height -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto min-h-0">
				<div class="max-w-5xl mx-auto px-2 sm:px-4 py-4 sm:py-6 space-y-4 sm:space-y-6">
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
								<!-- Live Thinking Panel - shown during streaming regardless of content -->
								{#if isStreaming && isLastMessage && hasThinking && currentThinking}
										<div class="mb-3 border border-amber-200 rounded-xl overflow-hidden bg-amber-50/50 shadow-sm">
											<button
												onclick={() => thinkingExpanded = !thinkingExpanded}
												class="w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-amber-100/50 transition-colors"
											>
												<svg class="w-4 h-4 text-amber-600 transition-transform duration-200 {thinkingExpanded ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
												</svg>
												<svg class="w-4 h-4 text-amber-600 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
												</svg>
												<span class="text-sm font-medium text-amber-700">Thinking</span>
												<span class="ml-2 w-2 h-2 bg-amber-500 rounded-full animate-pulse"></span>
												<span class="ml-auto text-xs text-amber-500">{Math.ceil(currentThinking.length / 4)} tokens</span>
											</button>
											{#if thinkingExpanded}
												<div class="px-4 pb-3 text-sm text-amber-800/90 whitespace-pre-wrap border-t border-amber-200 bg-amber-50/30 max-h-72 overflow-y-auto font-mono text-xs leading-relaxed">
													{currentThinking}<span class="inline-block w-1.5 h-4 bg-amber-500 animate-pulse ml-0.5"></span>
												</div>
											{/if}
										</div>
								{/if}

								<!-- Loading indicator when waiting for first content -->
								{#if !message.content && !message.artifacts?.length && isStreaming && isLastMessage && (!hasThinking || !currentThinking)}
									{@const modelId = selectedModel.toLowerCase()}
									{@const isLargeModel = modelId.includes(':30b') || modelId.includes(':32b') || modelId.includes(':70b') || modelId.includes(':72b') || modelId.includes(':235b')}
									{@const isColdStart = (activeProvider === 'ollama_local' || activeProvider === 'ollama_cloud') && !warmedUpModels.has(selectedModel)}
									<div class="flex flex-col gap-1.5 p-3 bg-gray-50 rounded-xl border border-gray-100">
										<div class="flex items-center gap-2 text-sm text-gray-600">
											<svg class="w-4 h-4 animate-spin text-blue-500" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
											</svg>
											<span class="font-medium">
												{#if isColdStart}
													Loading {currentModelName}...
												{:else}
													Generating response...
												{/if}
											</span>
										</div>
										{#if isColdStart}
											<div class="text-xs text-gray-500 ml-6 space-y-0.5">
												<p>Loading model into memory for first use</p>
												{#if isLargeModel}
													<p class="text-orange-600">âš ï¸ Large model ({currentModelName}) - this may take 30-60 seconds</p>
												{:else}
													<p>This usually takes 5-15 seconds</p>
												{/if}
											</div>
										{:else if isLargeModel}
											<div class="text-xs text-gray-500 ml-6">
												<p class="text-orange-600">Using large model - response may be slower</p>
											</div>
										{/if}
									</div>
								{/if}

								<!-- Message content when available -->
								{#if message.content || message.artifacts?.length}
									<!-- Extract thinking from content -->
									{@const extracted = extractThinking(message.content)}

									<!-- Show thinking panel if present -->
									{#if extracted.thinking}
										<div class="mb-3 border border-amber-200 rounded-xl overflow-hidden bg-amber-50/50">
											<button
												onclick={() => thinkingExpanded = !thinkingExpanded}
												class="w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-amber-100/50 transition-colors"
											>
												<svg class="w-4 h-4 text-amber-600 transition-transform duration-200 {thinkingExpanded ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
												</svg>
												<svg class="w-4 h-4 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
												</svg>
												<span class="text-sm font-medium text-amber-700">Thinking</span>
												<span class="ml-auto text-xs text-amber-500">{Math.ceil(extracted.thinking.length / 4)} tokens</span>
											</button>
											{#if thinkingExpanded}
												<div class="px-4 pb-3 text-sm text-amber-800/90 whitespace-pre-wrap border-t border-amber-200 bg-amber-50/30 max-h-72 overflow-y-auto font-mono text-xs leading-relaxed">
													{extracted.thinking}
												</div>
											{/if}
										</div>
									{/if}

									<!-- Show text content if any -->
									{#if extracted.mainContent}
										<div class="text-sm sm:text-[15px] leading-relaxed text-gray-800 dark:text-gray-100 prose prose-sm dark:prose-invert max-w-none streaming-content">
											{@html renderMarkdown(extracted.mainContent)}{#if isLastMessage && isStreaming && !artifactCompletedInStream}<span class="streaming-cursor"></span>{/if}
										</div>
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
														<p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{artifact.title}</p>
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
															<p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{artifact.title}</p>
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
								<!-- Cursor is now inline with text content in streaming-content div -->

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
										<MessageActions 
											messageId={message.id}
											conversationId={conversationId || undefined}
											agentType={selectedModel}
											originalContent={message.content}
											onCopy={() => copyMessage(message.content, message.id)}
											copied={copiedMessageId === message.id}
										/>

										<!-- Usage stats display -->
										{#if message.usage && showUsageInChat}
											<div class="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500 ml-2 pl-2 border-l border-gray-200 dark:border-gray-700">
												<span class="flex items-center gap-1" title="Tokens per second">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
													</svg>
													{message.usage.tps.toFixed(1)} t/s
												</span>
												<span title="Total tokens">{message.usage.total_tokens} tokens</span>
												{#if message.usage.thinking_tokens > 0}
													<span class="text-purple-400" title="Thinking tokens (COT)">{message.usage.thinking_tokens} thinking</span>
												{/if}
												<span class="text-gray-300 dark:text-gray-600">.</span>
												<span title="Provider">{message.usage.model?.toLowerCase().includes('-cloud') ? 'Cloud' : (message.usage.provider === 'ollama_local' ? 'Local' : message.usage.provider)}</span>
											</div>
										{/if}
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
			<div class="chat-input-area flex-shrink-0 p-4">
				<div class="max-w-3xl mx-auto">
					<div class="chat-input-box bg-white rounded-2xl shadow-sm border border-gray-200 p-3 cursor-text" onclick={() => inputRef?.focus()}>
						<!-- Hidden file input -->
						<input
							bind:this={fileInputRef}
							type="file"
							multiple
							accept="image/*,.pdf,.txt,.md,.json,.csv,.doc,.docx"
							class="hidden"
							onchange={handleFileSelect}
						/>

						<!-- Attached files display -->
						{#if attachedFiles.length > 0}
							<div class="flex flex-wrap gap-2 mb-3">
								{#each attachedFiles as file (file.id)}
									<div class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg text-sm">
										{#if file.type.startsWith('image/') && file.content}
											<img src={file.content} alt={file.name} class="w-6 h-6 rounded object-cover" />
										{:else}
											<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
											</svg>
										{/if}
										<span class="text-gray-700 truncate max-w-[150px]">{file.name}</span>
										<span class="text-gray-400 text-xs">{formatFileSize(file.size)}</span>
										<button
											onclick={(e) => { e.stopPropagation(); removeAttachedFile(file.id); }}
											class="p-0.5 text-gray-400 hover:text-gray-600"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</div>
								{/each}
							</div>
						{/if}

						<!-- Recording overlay - Windsurf style with live transcript -->
						{#if isRecording}
							<div class="mb-3">
								<!-- Live transcript preview -->
								{#if liveTranscript}
									<div class="text-gray-600 text-sm mb-3 min-h-[24px] animate-pulse">
										{liveTranscript}
									</div>
								{:else}
									<div class="text-gray-400 text-sm mb-3 min-h-[24px]">
										Listening...
									</div>
								{/if}

								<!-- Waveform bar -->
								<div class="flex items-center gap-3 bg-gray-800 rounded-full px-4 py-2">
									<!-- Cancel button -->
									<button
										onclick={(e) => {
											e.stopPropagation();
											// Cancel without transcribing
											if (mediaRecorder && mediaRecorder.state !== 'inactive') {
												mediaRecorder.ondataavailable = null;
												mediaRecorder.onstop = null;
												mediaRecorder.stop();
											}
											isRecording = false;
											if (recordingInterval) {
												clearInterval(recordingInterval);
												recordingInterval = null;
											}
											if (speechRecognition) {
												speechRecognition.stop();
												speechRecognition = null;
											}
											if (animationFrameId) {
												cancelAnimationFrame(animationFrameId);
												animationFrameId = null;
											}
											if (audioContext) {
												audioContext.close();
												audioContext = null;
											}
											waveformBars = Array(40).fill(2);
											liveTranscript = '';
											recordingDuration = 0;
										}}
										class="p-1.5 text-gray-400 hover:text-white transition-colors"
										aria-label="Cancel recording"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>

									<!-- Waveform visualization -->
									<div class="flex-1 flex items-center justify-center gap-[1px] h-8">
										{#each waveformBars as height, i}
											<div
												class="w-[2px] bg-white rounded-full"
												style="height: {height}px"
											></div>
										{/each}
									</div>

									<!-- Duration -->
									<span class="text-white font-mono text-sm min-w-[40px] text-right">{recordingTimeDisplay()}</span>

									<!-- Confirm button -->
									<button
										onclick={(e) => { e.stopPropagation(); stopRecording(); }}
										class="p-1.5 bg-white text-gray-800 rounded-full hover:bg-gray-200 transition-colors"
										aria-label="Stop and transcribe"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
									</button>
								</div>
							</div>
						{:else if isTranscribing}
							<div class="flex items-center gap-3 mb-3 py-4 px-2">
								<svg class="w-5 h-5 animate-spin text-blue-500" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								<span class="text-blue-500 font-medium">Transcribing audio...</span>
							</div>
						{:else}
							<!-- Inline Project Picker (appears when Enter pressed without project) -->
							{#if showInlineProjectPicker}
								<div class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden" transition:fly={{ y: 10, duration: 150 }}>
									<div class="px-3 py-2 border-b border-gray-200 bg-white">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Select a project to continue</span>
									</div>
									<div class="max-h-48 overflow-y-auto">
										{#each projectsList as project, i (project.id)}
											<button
												onclick={() => { selectedProjectId = project.id; showInlineProjectPicker = false; }}
												class="w-full px-3 py-2 text-left transition-colors flex items-center gap-3 {projectDropdownIndex === i ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
											>
												<div class="w-7 h-7 rounded-lg {projectDropdownIndex === i ? 'bg-blue-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
													</svg>
												</div>
												<span class="text-sm font-medium truncate">{project.name}</span>
											</button>
										{/each}
										<!-- Create new project option -->
										<button
											onclick={() => { showInlineProjectPicker = false; showNewProjectModal = true; }}
											class="w-full px-3 py-2 text-left transition-colors flex items-center gap-3 border-t border-gray-200 {projectDropdownIndex === projectsList.length ? 'bg-gray-100 text-gray-900' : 'hover:bg-gray-50 text-gray-600'}"
										>
											<div class="w-7 h-7 rounded-lg {projectDropdownIndex === projectsList.length ? 'bg-gray-900 text-white' : 'bg-gray-200 text-gray-500'} flex items-center justify-center flex-shrink-0">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
												</svg>
											</div>
											<span class="text-sm font-medium">Create new project</span>
										</button>
									</div>
									<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
										â†‘â†“ Navigate Â· Enter Select Â· Esc Cancel
									</div>
								</div>
							{/if}

							<!-- Slash Command Suggestions -->
							{#if showCommandSuggestions}
								<div class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden" transition:fly={{ y: 10, duration: 150 }}>
									<div class="px-3 py-2 border-b border-gray-200 bg-white">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Commands</span>
									</div>
									<div class="max-h-64 overflow-y-auto" id="command-list-messages">
										{#each filteredCommands as cmd, i (cmd.name)}
											<button
												data-command-index={i}
												onclick={() => selectCommand(cmd)}
												class="w-full px-3 py-2.5 text-left transition-colors flex items-center gap-3 {commandDropdownIndex === i ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
											>
												<div class="w-8 h-8 rounded-lg {commandDropdownIndex === i ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-600'} flex items-center justify-center flex-shrink-0">
													<span class="text-sm font-bold">/</span>
												</div>
												<div class="flex-1 min-w-0">
													<div class="text-sm font-medium">{cmd.display_name}</div>
													<div class="text-xs text-gray-500 truncate">{cmd.description}</div>
												</div>
												<span class="text-xs text-gray-400 font-mono">/{cmd.name}</span>
											</button>
										{/each}
									</div>
									<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
										â†‘â†“ Navigate Â· Enter/Tab Select Â· Esc Cancel
									</div>
								</div>
							{/if}

							<!-- Agent Mention Suggestions -->
							{#if showAgentSuggestions}
								<div class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden" transition:fly={{ y: 10, duration: 150 }}>
									<div class="px-3 py-2 border-b border-gray-200 bg-white">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Agents</span>
									</div>
									<div class="max-h-64 overflow-y-auto" id="agent-list-messages">
										{#each filteredAgents as agent, i (agent.id)}
											<button
												data-agent-index={i}
												onclick={() => selectAgent(agent)}
												class="w-full px-3 py-2.5 text-left transition-colors flex items-center gap-3 {agentDropdownIndex === i ? 'bg-purple-50 text-purple-700' : 'hover:bg-gray-100 text-gray-700'}"
											>
												<div class="w-8 h-8 rounded-lg {agentDropdownIndex === i ? 'bg-purple-500 text-white' : 'bg-gray-200 text-gray-600'} flex items-center justify-center flex-shrink-0">
													<span class="text-sm font-bold">@</span>
												</div>
												<div class="flex-1 min-w-0">
													<div class="text-sm font-medium">{agent.display_name}</div>
													<div class="text-xs text-gray-500 truncate">{agent.description || agent.category || 'Agent'}</div>
												</div>
												<span class="text-xs text-gray-400 font-mono">@{agent.name}</span>
											</button>
										{/each}
									</div>
									<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
										Use arrow keys + Enter/Tab to select
									</div>
								</div>
							{/if}

							<!-- Active Command Chip (when command is selected) -->
							{#if activeCommand}
								<div class="mb-2 flex items-center gap-2" transition:fly={{ y: -5, duration: 150 }}>
									<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-full">
										<div class="w-5 h-5 rounded bg-blue-500 text-white flex items-center justify-center">
											<span class="text-xs font-bold">/</span>
										</div>
										<span class="text-sm font-medium text-blue-700">{activeCommand.display_name}</span>
										<button
											onclick={clearActiveCommand}
											class="w-4 h-4 rounded-full bg-blue-200 hover:bg-blue-300 text-blue-600 flex items-center justify-center transition-colors"
											aria-label="Clear command"
										>
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
											</svg>
										</button>
									</div>
									<span class="text-xs text-gray-500">{activeCommand.description}</span>
								</div>
							{/if}

							<!-- Detected Agent Chip (when @agent is detected) -->
							{#if detectedAgent}
								<div class="mb-2 flex items-center gap-2" transition:fly={{ y: -5, duration: 150 }}>
									<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-purple-50 border border-purple-200 rounded-full">
										<div class="w-5 h-5 rounded bg-purple-500 text-white flex items-center justify-center">
											<span class="text-xs font-bold">@</span>
										</div>
										<span class="text-sm font-medium text-purple-700">{detectedAgent.display_name}</span>
										<button
											onclick={clearDetectedAgent}
											class="w-4 h-4 rounded-full bg-purple-200 hover:bg-purple-300 text-purple-600 flex items-center justify-center transition-colors"
											aria-label="Clear agent"
										>
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
											</svg>
										</button>
									</div>
									<span class="text-xs text-gray-500">{detectedAgent.description || detectedAgent.category || 'Specialist'}</span>
								</div>
							{/if}

							<!-- Textarea -->
							<textarea
								bind:this={inputRef}
								bind:value={inputValue}
								placeholder="Ask OSA anything... (type / for commands)"
								rows={1}
								disabled={isStreaming}
								class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
								style="min-height: 24px; max-height: 200px;"
								onkeydown={handleKeydown}
								oninput={handleInput}
							></textarea>
						{/if}

						<!-- Bottom row -->
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<!-- Plus button with menu -->
								<div class="relative">
									<button
										onclick={(e) => { e.stopPropagation(); showPlusMenu = !showPlusMenu; showContextDropdown = false; showModelDropdown = false; }}
										class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
										aria-label="Add"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</button>
									{#if showPlusMenu}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[180px] z-10"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<button
												onclick={() => startNewConversation()}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
												</svg>
												New conversation
											</button>
											<button
												onclick={() => { showPlusMenu = false; showContextDropdown = true; }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
												</svg>
												Add context
											</button>
											<button
												onclick={() => { showPlusMenu = false; showHybridSearchPanel = true; }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
												</svg>
												Search knowledge
											</button>
											<button
												onclick={() => { showPlusMenu = false; fileInputRef?.click(); }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
												</svg>
												Attach image
											</button>
											<button
												onclick={() => { showPlusMenu = false; showDocumentUploadModal = true; }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
												</svg>
												Upload document
											</button>
										</div>
									{/if}
								</div>

								<!-- Attachment -->
								<button
									onclick={(e) => { e.stopPropagation(); fileInputRef?.click(); }}
									class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									aria-label="Attach image"
								>
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
														<span class="text-base">{ctx.icon || 'ðŸ“„'}</span>
													{/if}
													<span class="truncate">{ctx.name}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<!-- Focus Mode Selector (inline for existing conversations) -->
								<div class="relative">
									<button
										onclick={() => { showFocusDropdown = !showFocusDropdown; showModelDropdown = false; showContextDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors {selectedFocusId ? 'text-purple-600 bg-purple-50 hover:bg-purple-100' : 'text-gray-600 hover:bg-gray-100'}"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
										</svg>
										{#if selectedFocusId}
											{@const mode = FOCUS_MODES.find(m => m.id === selectedFocusId)}
											<span class="font-medium">{mode?.name || 'Focus'}</span>
										{:else}
											<span>Focus</span>
										{/if}
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
										</svg>
									</button>

									{#if showFocusDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[200px] z-10"
											transition:fly={{ y: 5, duration: 150 }}
										>
											{#if selectedFocusId}
												<button
													onclick={() => { selectedFocusId = null; focusOptions = {}; showFocusDropdown = false; }}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-600 border-b border-gray-100"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
													</svg>
													Clear Focus Mode
												</button>
											{/if}
											{#each FOCUS_MODES as mode (mode.id)}
												{@const isSelected = selectedFocusId === mode.id}
												<button
													onclick={() => {
														selectedFocusId = mode.id;
														focusOptions = getDefaultOptions(mode);
														showFocusDropdown = false;
													}}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 {isSelected ? 'text-purple-600 font-medium bg-purple-50' : 'text-gray-600'}"
												>
													{#if isSelected}
														<svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
														</svg>
													{:else}
														<span class="w-4 h-4 flex items-center justify-center text-gray-400">
															{#if mode.icon === 'magnifying-glass-chart'}
																<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
															{:else if mode.icon === 'chart-bar'}
																<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
															{:else if mode.icon === 'document-text'}
																<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
															{:else if mode.icon === 'cube'}
																<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>
															{:else}
																<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
															{/if}
														</span>
													{/if}
													<span class="truncate">{mode.name}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<!-- Context Window Indicator (only show when conversation has messages) -->
								{#if messages.length > 0}
									<div class="group relative flex items-center">
										<div class="flex items-center gap-1.5 px-2 py-1 text-xs text-gray-400 hover:text-gray-600 cursor-default transition-colors">
											<span class="tabular-nums font-medium">{formatTokenCount(totalConversationTokens())}</span>
											<span class="text-gray-300">/</span>
											<span class="tabular-nums">{formatTokenCount(currentContextLimit())}</span>
											{#if contextUsagePercent() >= 50}
												<div class="w-12 h-1 bg-gray-200 rounded-full overflow-hidden ml-1">
													<div
														class="h-full rounded-full transition-all duration-300 {contextUsagePercent() > 80 ? 'bg-red-500' : 'bg-yellow-500'}"
														style="width: {contextUsagePercent()}%"
													></div>
												</div>
											{/if}
										</div>
										<!-- Tooltip on hover -->
										<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-2 bg-gray-900 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
											<div class="font-medium mb-1">Context Window</div>
											<div class="text-gray-300">{totalConversationTokens().toLocaleString()} / {currentContextLimit().toLocaleString()} tokens</div>
											<div class="text-gray-400 mt-1">{contextUsagePercent()}% used</div>
											{#if nodeContextTokens() > 0 || contextDocTokens() > 0}
												<div class="border-t border-gray-700 mt-2 pt-2 text-gray-400">
													<div class="flex justify-between gap-4">
														<span>Messages:</span>
														<span>{messageTokens().toLocaleString()}</span>
													</div>
													{#if nodeContextTokens() > 0}
														<div class="flex justify-between gap-4">
															<span>Node context:</span>
															<span>~{nodeContextTokens().toLocaleString()}</span>
														</div>
													{/if}
													{#if contextDocTokens() > 0}
														<div class="flex justify-between gap-4">
															<span>Documents ({selectedContexts.length}):</span>
															<span>~{contextDocTokens().toLocaleString()}</span>
														</div>
													{/if}
												</div>
											{/if}
											<div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
										</div>
									</div>
								{/if}
							</div>

							<!-- Right controls: COT Toggle + Mic + Send/Stop -->
							<div class="flex items-center gap-2">
								<!-- Chain of Thought (COT) Toggle -->
								<button
									type="button"
									onclick={toggleCOT}
									class="flex-shrink-0 group relative flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg transition-all duration-200 {useCOT ? 'bg-amber-100 text-amber-700 hover:bg-amber-200' : 'bg-gray-100 text-gray-500 hover:bg-gray-200 hover:text-gray-700'}"
									aria-label="Toggle Chain of Thought reasoning"
									title="{useCOT ? 'Thinking enabled' : 'Thinking disabled'}"
								>
									<svg class="w-4 h-4 {useCOT ? 'animate-pulse' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
									</svg>
									<span class="text-xs font-medium">COT</span>
									<!-- Tooltip -->
									<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
										{useCOT ? 'Chain of Thought: ON' : 'Chain of Thought: OFF'}
										<div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
									</div>
								</button>

								<!-- Voice Recording Button -->
								{#if !isRecording && !isTranscribing}
									<button
										type="button"
										onclick={(e) => { e.stopPropagation(); toggleRecording(); }}
										class="flex-shrink-0 w-10 h-10 flex items-center justify-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xl transition-colors"
										aria-label="Voice input"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
										</svg>
									</button>
								{/if}

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
			</div>
		{:else}
			<!-- Empty State - centered in available space -->
			<div class="flex-1 flex items-center justify-center overflow-auto">
				<div class="w-full max-w-3xl px-6">
					{#if focusModeEnabled}
						<!-- Focus Mode UI -->
						<FocusModeSelector
							onSubmit={handleFocusModeSubmit}
							commands={availableCommands}
							onModeChange={(isFocus: boolean) => focusModeEnabled = isFocus}
							{selectedProjectId}
							onRequestProjectSelect={() => showProjectDropdown = true}
							availableContexts={availableContexts.map(c => ({ id: c.id, name: c.name, icon: c.icon ?? undefined }))}
							{selectedContextIds}
							onContextToggle={(id: string) => {
								if (selectedContextIds.includes(id)) {
									selectedContextIds = selectedContextIds.filter(cid => cid !== id);
								} else {
									selectedContextIds = [...selectedContextIds, id];
								}
							}}
							initialInput={focusModeInitialInput}
						/>
					{:else}
						<!-- Classic Mode - Personalized Title -->
						<div class="text-center mb-8">
							<h1 class="text-3xl font-semibold text-gray-900 mb-3">
								{personalizedGreeting}
							</h1>
							<p class="text-gray-500 h-6">
								Let me help you <span class="text-blue-600 font-medium">{displayedSuggestion}</span><span class="cursor-blink text-blue-600 font-light">|</span>
							</p>
						</div>

						<!-- Input Box (Classic Mode) -->
					<div class="bg-white rounded-3xl shadow-lg border border-gray-200 p-4 cursor-text" onclick={() => inputRef?.focus()}>
						<!-- Attached files display (empty state) -->
						{#if attachedFiles.length > 0}
							<div class="flex flex-wrap gap-2 mb-3">
								{#each attachedFiles as file (file.id)}
									<div class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg text-sm">
										{#if file.type.startsWith('image/') && file.content}
											<img src={file.content} alt={file.name} class="w-6 h-6 rounded object-cover" />
										{:else}
											<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
											</svg>
										{/if}
										<span class="text-gray-700 truncate max-w-[150px]">{file.name}</span>
										<span class="text-gray-400 text-xs">{formatFileSize(file.size)}</span>
										<button
											onclick={(e) => { e.stopPropagation(); removeAttachedFile(file.id); }}
											class="p-0.5 text-gray-400 hover:text-gray-600"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</div>
								{/each}
							</div>
						{/if}

						<!-- Recording overlay - Windsurf style (empty state) -->
						{#if isRecording}
							<div class="mb-3">
								<!-- Live transcript preview -->
								{#if liveTranscript}
									<div class="text-gray-600 text-sm mb-3 min-h-[24px] animate-pulse">
										{liveTranscript}
									</div>
								{:else}
									<div class="text-gray-400 text-sm mb-3 min-h-[24px]">
										Listening...
									</div>
								{/if}

								<!-- Waveform bar -->
								<div class="flex items-center gap-3 bg-gray-800 rounded-full px-4 py-2">
									<!-- Cancel button -->
									<button
										onclick={(e) => {
											e.stopPropagation();
											if (mediaRecorder && mediaRecorder.state !== 'inactive') {
												mediaRecorder.ondataavailable = null;
												mediaRecorder.onstop = null;
												mediaRecorder.stop();
											}
											isRecording = false;
											if (recordingInterval) {
												clearInterval(recordingInterval);
												recordingInterval = null;
											}
											if (speechRecognition) {
												speechRecognition.stop();
												speechRecognition = null;
											}
											if (animationFrameId) {
												cancelAnimationFrame(animationFrameId);
												animationFrameId = null;
											}
											if (audioContext) {
												audioContext.close();
												audioContext = null;
											}
											waveformBars = Array(40).fill(2);
											liveTranscript = '';
											recordingDuration = 0;
										}}
										class="p-1.5 text-gray-400 hover:text-white transition-colors"
										aria-label="Cancel recording"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>

									<!-- Waveform visualization -->
									<div class="flex-1 flex items-center justify-center gap-[1px] h-8">
										{#each waveformBars as height, i}
											<div
												class="w-[2px] bg-white rounded-full"
												style="height: {height}px"
											></div>
										{/each}
									</div>

									<!-- Duration -->
									<span class="text-white font-mono text-sm min-w-[40px] text-right">{recordingTimeDisplay()}</span>

									<!-- Confirm button -->
									<button
										onclick={(e) => { e.stopPropagation(); stopRecording(); }}
										class="p-1.5 bg-white text-gray-800 rounded-full hover:bg-gray-200 transition-colors"
										aria-label="Stop and transcribe"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
									</button>
								</div>
							</div>
						{:else if isTranscribing}
							<div class="flex items-center gap-3 mb-3 py-4 px-2">
								<svg class="w-5 h-5 animate-spin text-blue-500" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								<span class="text-blue-500 font-medium">Transcribing audio...</span>
							</div>
						{:else}
							<!-- Inline Project Picker (appears when Enter pressed without project - empty state) -->
							{#if showInlineProjectPicker}
								<div class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden" transition:fly={{ y: 10, duration: 150 }}>
									<div class="px-3 py-2 border-b border-gray-200 bg-white">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Select a project to continue</span>
									</div>
									<div class="max-h-48 overflow-y-auto">
										{#each projectsList as project, idx (project.id)}
											<button
												onclick={() => { selectedProjectId = project.id; showInlineProjectPicker = false; }}
												class="w-full px-3 py-2 text-left transition-colors flex items-center gap-3 {projectDropdownIndex === idx ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
											>
												<div class="w-7 h-7 rounded-lg {projectDropdownIndex === idx ? 'bg-blue-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
													</svg>
												</div>
												<span class="text-sm font-medium truncate">{project.name}</span>
											</button>
										{/each}
										<!-- Create new project option -->
										<button
											onclick={() => { showInlineProjectPicker = false; showNewProjectModal = true; }}
											class="w-full px-3 py-2 text-left transition-colors flex items-center gap-3 border-t border-gray-200 {projectDropdownIndex === projectsList.length ? 'bg-gray-100 text-gray-900' : 'hover:bg-gray-50 text-gray-600'}"
										>
											<div class="w-7 h-7 rounded-lg {projectDropdownIndex === projectsList.length ? 'bg-gray-900 text-white' : 'bg-gray-200 text-gray-500'} flex items-center justify-center flex-shrink-0">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
												</svg>
											</div>
											<span class="text-sm font-medium">Create new project</span>
										</button>
									</div>
									<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
										â†‘â†“ Navigate Â· Enter Select Â· Esc Cancel
									</div>
								</div>
							{/if}

							<!-- Slash Command Suggestions (empty state) -->
							{#if showCommandSuggestions}
								<div class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden" transition:fly={{ y: 10, duration: 150 }}>
									<div class="px-3 py-2 border-b border-gray-200 bg-white">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Commands</span>
									</div>
									<div class="max-h-64 overflow-y-auto" id="command-list-empty">
										{#each filteredCommands as cmd, i (cmd.name)}
											<button
												data-command-index={i}
												onclick={() => selectCommand(cmd)}
												class="w-full px-3 py-2.5 text-left transition-colors flex items-center gap-3 {commandDropdownIndex === i ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
											>
												<div class="w-8 h-8 rounded-lg {commandDropdownIndex === i ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-600'} flex items-center justify-center flex-shrink-0">
													<span class="text-sm font-bold">/</span>
												</div>
												<div class="flex-1 min-w-0">
													<div class="text-sm font-medium">{cmd.display_name}</div>
													<div class="text-xs text-gray-500 truncate">{cmd.description}</div>
												</div>
												<span class="text-xs text-gray-400 font-mono">/{cmd.name}</span>
											</button>
										{/each}
									</div>
									<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
										â†‘â†“ Navigate Â· Enter/Tab Select Â· Esc Cancel
									</div>
								</div>
							{/if}

							<!-- Active Command Chip (empty state) -->
							{#if activeCommand}
								<div class="mb-2 flex items-center gap-2" transition:fly={{ y: -5, duration: 150 }}>
									<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-full">
										<div class="w-5 h-5 rounded bg-blue-500 text-white flex items-center justify-center">
											<span class="text-xs font-bold">/</span>
										</div>
										<span class="text-sm font-medium text-blue-700">{activeCommand.display_name}</span>
										<button
											onclick={clearActiveCommand}
											class="w-4 h-4 rounded-full bg-blue-200 hover:bg-blue-300 text-blue-600 flex items-center justify-center transition-colors"
											aria-label="Clear command"
										>
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
											</svg>
										</button>
									</div>
									<span class="text-xs text-gray-500">{activeCommand.description}</span>
								</div>
							{/if}

							<textarea
								bind:this={inputRef}
								bind:value={inputValue}
								placeholder="Ask OSA anything... (type / for commands)"
								rows={1}
								disabled={isStreaming}
								class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
								style="min-height: 24px; max-height: 200px;"
								onkeydown={handleKeydown}
								oninput={handleInput}
							></textarea>
						{/if}

						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<!-- Plus button with menu (empty state) -->
								<div class="relative">
									<button
										onclick={(e) => { e.stopPropagation(); showPlusMenu = !showPlusMenu; showContextDropdown = false; showModelDropdown = false; }}
										class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
										aria-label="Add"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</button>
									{#if showPlusMenu}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[180px] z-10"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<button
												onclick={() => { showPlusMenu = false; showContextDropdown = true; }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
												</svg>
												Add context
											</button>
											<button
												onclick={() => { showPlusMenu = false; fileInputRef?.click(); }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
												</svg>
												Attach image
											</button>
											<button
												onclick={() => { showPlusMenu = false; showDocumentUploadModal = true; }}
												class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
												</svg>
												Upload document
											</button>
										</div>
									{/if}
								</div>

								<!-- Attachment -->
								<button
									onclick={(e) => { e.stopPropagation(); fileInputRef?.click(); }}
									class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									aria-label="Attach image"
								>
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
														<span class="text-base">{ctx.icon || 'ðŸ“„'}</span>
													{/if}
													<span class="truncate">{ctx.name}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>

							<!-- Right controls: Mic + Send (empty state) -->
							<div class="flex items-center gap-2">
								<!-- Voice Recording Button -->
								{#if !isRecording && !isTranscribing}
									<button
										type="button"
										onclick={(e) => { e.stopPropagation(); toggleRecording(); }}
										class="flex-shrink-0 w-10 h-10 flex items-center justify-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xl transition-colors"
										aria-label="Voice input"
									>
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
										</svg>
									</button>
								{/if}

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
					</div>

					<!-- Quick Actions (Classic Mode only) -->
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

					<!-- Switch to Focus Mode -->
					<div class="flex justify-center mt-6">
						<button
							onclick={() => focusModeEnabled = true}
							class="flex items-center gap-2 px-4 py-2 text-sm text-gray-500 border border-gray-200 rounded-full hover:text-gray-700 hover:bg-gray-50 hover:border-gray-300 transition-all"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
							</svg>
							Switch to Focus mode
						</button>
					</div>
					{/if}
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

		<div class="h-full flex flex-col bg-white dark:bg-gray-900 {isArtifactFocused ? 'w-1/2' : 'flex-shrink-0'}" style="{isArtifactFocused ? '' : `width: ${artifactPanelWidth}px`}" transition:fly={{ x: 320, duration: 200 }}>
			<!-- Panel Header -->
			<div class="p-4 border-b border-gray-100 dark:border-gray-700 flex-shrink-0">
				<div class="flex items-center justify-between mb-3">
					<h3 class="font-semibold text-gray-900 dark:text-gray-100">Artifacts</h3>
					<button
						onclick={() => { artifactsPanelOpen = false; viewingArtifactFromMessage = null; }}
						class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
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
								class="px-2.5 py-1 text-xs font-medium rounded-lg whitespace-nowrap transition-colors {artifactFilter === filter ? 'bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
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
				<div class="flex-1 flex flex-col overflow-hidden bg-[#1c1c1e]">
					<!-- Generation Header -->
					<div class="p-4 border-b border-white/10 flex-shrink-0">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg {generatingArtifactType ? getArtifactColor(generatingArtifactType) : 'bg-white/10 text-gray-300'} flex items-center justify-center flex-shrink-0 relative">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={generatingArtifactType ? getArtifactIcon(generatingArtifactType) : 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'} />
								</svg>
								<!-- Generating indicator -->
								<div class="absolute -top-1 -right-1 w-3 h-3">
									<span class="absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75 animate-ping"></span>
									<span class="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
								</div>
							</div>
							<div class="min-w-0 flex-1">
								<h4 class="font-medium text-gray-100 truncate">
									{generatingArtifactTitle || 'Generating artifact...'}
								</h4>
								<p class="text-xs text-gray-400 flex items-center gap-1.5">
									{#if generatingArtifactType}
										<span class="capitalize">{generatingArtifactType}</span>
										<span>&bull;</span>
									{/if}
									<span class="flex items-center gap-1 text-green-400">
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
					<div class="flex-1 overflow-y-auto p-4">
						<div class="prose prose-sm dark:prose-invert max-w-none" style="--tw-prose-body: #f5f5f7; --tw-prose-headings: #fff; --tw-prose-bold: #fff;">
							{@html renderMarkdown(generatingArtifactContent || 'Waiting for content...')}
							<span class="inline-block w-2 h-4 bg-green-500 animate-pulse ml-0.5"></span>
						</div>
					</div>
				</div>
			{:else if viewingArtifactFromMessage}
				<!-- Viewing artifact from message -->
				<div class="flex-1 flex flex-col overflow-hidden bg-[#1c1c1e]">
					<!-- Header - artifact info with actions -->
					<div class="p-4 border-b border-white/10 flex-shrink-0">
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(viewingArtifactFromMessage.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(viewingArtifactFromMessage.type)} />
								</svg>
							</div>
							<div class="min-w-0 flex-1">
								<h4 class="font-medium text-gray-100">{viewingArtifactFromMessage.title}</h4>
								<p class="text-xs text-gray-400 capitalize">{viewingArtifactFromMessage.type}</p>
							</div>
							<!-- Action icons -->
							<div class="flex items-center gap-1">
								<button
									onclick={() => { copyToClipboard(viewingArtifactFromMessage?.content || ''); }}
									class="p-2 text-gray-400 hover:text-gray-200 hover:bg-white/10 rounded-lg transition-colors"
									title="Copy"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								</button>
								<button
									class="p-2 text-gray-400 hover:text-gray-200 hover:bg-white/10 rounded-lg transition-colors"
									title="Export"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
									</svg>
								</button>
								<button
									onclick={openSaveToProfileModal}
									class="p-2 text-gray-400 hover:text-gray-200 hover:bg-white/10 rounded-lg transition-colors"
									title="Save to Knowledge Base"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
									</svg>
								</button>
							</div>
						</div>
					</div>

					<!-- Block Editor for artifact editing -->
					<ArtifactEditor
						artifact={viewingArtifactFromMessage}
						onSave={updateArtifactContent}
						darkMode={true}
					/>
				</div>
			{:else if selectedArtifact}
				<!-- Artifact Detail View (from API) -->
				<div class="flex-1 flex flex-col overflow-hidden bg-[#1c1c1e]">
					<!-- Header - artifact info with actions -->
					<div class="p-4 border-b border-white/10 flex-shrink-0">
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(selectedArtifact.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(selectedArtifact.type)} />
								</svg>
							</div>
							<div class="min-w-0 flex-1">
								<h4 class="font-medium text-gray-100 truncate">{selectedArtifact.title}</h4>
								<p class="text-xs text-gray-400 capitalize">{selectedArtifact.type} &bull; v{selectedArtifact.version}</p>
							</div>
							<!-- Action icons -->
							<div class="flex items-center gap-1">
								<button
									onclick={() => { copyToClipboard(selectedArtifact?.content || ''); }}
									class="p-2 text-gray-400 hover:text-gray-200 hover:bg-white/10 rounded-lg transition-colors"
									title="Copy"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								</button>
								<button
									class="p-2 text-gray-400 hover:text-gray-200 hover:bg-white/10 rounded-lg transition-colors"
									title="Export"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
									</svg>
								</button>
							</div>
						</div>
					</div>

					<!-- Block Editor for artifact editing -->
					<ArtifactEditor
						artifact={{ title: selectedArtifact.title, type: selectedArtifact.type, content: selectedArtifact.content }}
						onSave={updateSelectedArtifactContent}
						darkMode={true}
					/>
				</div>
			{:else}
				<!-- Artifacts List -->
				<div class="flex-1 overflow-y-auto bg-[#1c1c1e]">
					{#if loadingArtifacts}
						<div class="flex items-center justify-center h-32">
							<div class="animate-spin h-6 w-6 border-2 border-white/30 border-t-transparent rounded-full"></div>
						</div>
					{:else if allArtifacts().length === 0}
						<div class="flex flex-col items-center justify-center h-48 text-center px-4">
							<div class="w-12 h-12 rounded-full bg-white/10 flex items-center justify-center mb-3">
								<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
							</div>
							<p class="text-sm text-gray-400">No artifacts yet</p>
							<p class="text-xs text-gray-500 mt-1">Ask OSA to create proposals, SOPs, or frameworks</p>
						</div>
					{:else}
						<div class="p-2 space-y-1">
							{#each allArtifacts() as artifact (artifact.id)}
								<div class="group relative">
									<button
										onclick={() => {
											// If it's from a message, view it directly
											if (artifact.fromMessage) {
												viewingArtifactFromMessage = {
													title: artifact.title,
													type: artifact.type,
													content: artifact.content || ''
												};
											} else {
												selectArtifact(artifact.id);
											}
										}}
										class="w-full flex items-start gap-3 p-3 rounded-lg hover:bg-white/5 transition-colors text-left"
									>
										<div class="w-9 h-9 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(artifact.type)} />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-100 truncate">{artifact.title}</p>
											{#if artifact.summary}
												<p class="text-xs text-gray-400 line-clamp-2 mt-0.5">{artifact.summary}</p>
											{/if}
											<div class="flex items-center gap-1.5 mt-1">
												<span class="text-xs text-gray-500 capitalize">{artifact.type}</span>
												{#if artifact.context_name}
													<span class="text-xs text-gray-600">&bull;</span>
													<span class="text-xs text-blue-400 truncate">{artifact.context_name}</span>
												{:else if artifact.project_id}
													<span class="text-xs text-gray-600">&bull;</span>
													<span class="text-xs text-purple-400 truncate">Linked to project</span>
												{:else}
													<span class="text-xs text-gray-600">&bull;</span>
													<span class="text-xs text-gray-500 italic">Unlinked</span>
												{/if}
											</div>
										</div>
									</button>
									<!-- Delete button - shows on hover -->
									<button
										onclick={(e) => { e.stopPropagation(); deleteArtifactById(artifact.id); }}
										class="absolute right-2 top-2 p-1.5 rounded-md text-gray-400 hover:text-red-500 hover:bg-red-50 opacity-0 group-hover:opacity-100 transition-all"
										title="Delete artifact"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Right Panel: Progress/Context/Artifacts -->
	{#if rightPanelOpen}
		<!-- Resize Handle -->
		<div
			class="w-1 h-full bg-gray-200 hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors relative group"
			onmousedown={startRightPanelResize}
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

		<div class="h-full flex flex-col bg-white flex-shrink-0" style="width: {rightPanelWidth}px" transition:fly={{ x: 320, duration: 200 }}>
			<!-- Panel Tabs -->
			<div class="flex border-b border-gray-200">
				<button
					onclick={() => rightPanelTab = 'progress'}
					class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'progress' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
				>
					Progress
				</button>
				<button
					onclick={() => rightPanelTab = 'context'}
					class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'context' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
				>
					Context
				</button>
				<button
					onclick={() => rightPanelTab = 'artifacts'}
					class="flex-1 px-3 py-3 text-xs font-medium transition-colors {rightPanelTab === 'artifacts' ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-500 hover:text-gray-700'}"
				>
					Artifacts
					{#if artifacts.length > 0}
						<span class="ml-1 px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-gray-200">{artifacts.length}</span>
					{/if}
				</button>
				<button
					onclick={() => rightPanelOpen = false}
					class="p-3 text-gray-400 hover:text-gray-600"
					aria-label="Close panel"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Panel Content -->
			<div class="flex-1 overflow-hidden">
				{#if rightPanelTab === 'progress'}
					<ProgressPanel tasks={delegatedTasks} />
				{:else if rightPanelTab === 'context'}
					<ContextPanel
						resources={activeResources}
						availableContexts={availableContexts.map(c => ({
							id: c.id,
							name: c.name
						}))}
						{selectedContextIds}
						onContextToggle={handleContextToggle}
					/>
				{:else if rightPanelTab === 'artifacts'}
					<!-- Artifacts List in Panel -->
					<div class="flex flex-col h-full bg-[#1c1c1e]">
						<div class="p-4 border-b border-white/10">
							<div class="flex items-center justify-between">
								<h3 class="text-sm font-semibold text-gray-100">Artifacts</h3>
								{#if allArtifacts().length > 0}
									<span class="text-xs text-gray-500">{allArtifacts().length} items</span>
								{/if}
							</div>
						</div>
						<div class="flex-1 overflow-y-auto p-2">
							{#if allArtifacts().length === 0}
								<div class="flex flex-col items-center justify-center py-12 px-4 text-center">
									<svg class="w-10 h-10 text-gray-500 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
									<p class="text-sm text-gray-400">No artifacts yet</p>
									<p class="text-xs text-gray-500 mt-1">Artifacts created by AI will appear here</p>
								</div>
							{:else}
								<div class="space-y-1">
									{#each allArtifacts() as artifact (artifact.id)}
										<button
											onclick={() => {
												if (artifact.fromMessage) {
													viewingArtifactFromMessage = {
														title: artifact.title,
														type: artifact.type,
														content: artifact.content || ''
													};
												} else {
													selectArtifact(artifact.id);
												}
											}}
											class="w-full p-3 rounded-lg hover:bg-white/5 transition-colors text-left group"
										>
											<div class="flex items-start gap-3">
												<div class="w-8 h-8 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getArtifactIcon(artifact.type)} />
													</svg>
												</div>
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium text-gray-100 truncate">{artifact.title}</p>
													<p class="text-xs text-gray-500 capitalize">{artifact.type}</p>
												</div>
											</div>
										</button>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
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
				<div class="space-y-2">
					<!-- Loose Documents option -->
					<button
						onclick={() => selectedProfileForSave = 'loose'}
						class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedProfileForSave === 'loose' ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
					>
						<div class="w-10 h-10 rounded-lg bg-gray-100 text-gray-600 flex items-center justify-center flex-shrink-0 text-lg">
							ðŸ“„
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900">Loose Documents</p>
							<p class="text-xs text-gray-500">Save without a parent profile</p>
						</div>
						{#if selectedProfileForSave === 'loose'}
							<svg class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
								<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						{/if}
					</button>

					{#if availableProfiles.length > 0}
						<div class="text-xs text-gray-400 uppercase tracking-wider font-medium pt-2 pb-1">Profiles</div>
						{#each availableProfiles as profile (profile.id)}
							<button
								onclick={() => selectedProfileForSave = profile.id}
								class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedProfileForSave === profile.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center flex-shrink-0 text-lg">
									{profile.icon || 'ðŸ“'}
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
					{/if}
				</div>
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

<!-- Quick Create Project Modal -->
{#if showNewProjectModal}
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="p-5 border-b border-gray-100">
				<h3 class="text-lg font-semibold text-gray-900">Create New Project</h3>
				<p class="text-sm text-gray-500 mt-1">Give your project a name to get started</p>
			</div>
			<div class="p-5">
				<input
					type="text"
					bind:value={newProjectName}
					placeholder="Project name..."
					class="w-full px-4 py-3 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
					onkeydown={(e) => { if (e.key === 'Enter') createProjectQuick(); if (e.key === 'Escape') showNewProjectModal = false; }}
					autofocus
				/>
			</div>
			<div class="p-4 border-t border-gray-100 flex justify-end gap-3">
				<button
					onclick={() => { showNewProjectModal = false; newProjectName = ''; }}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={createProjectQuick}
					disabled={!newProjectName.trim() || creatingProject}
					class="px-4 py-2 text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
				>
					{#if creatingProject}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
						</svg>
					{/if}
					Create & Start Chat
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Document Upload Modal -->
<DocumentUploadModal
	bind:open={showDocumentUploadModal}
	onClose={() => showDocumentUploadModal = false}
	onUploadComplete={(doc) => {
		console.log('Document uploaded:', doc);
		// Optionally add to active resources or show notification
		if (doc.id) {
			activeResources = [...activeResources, {
				id: doc.id,
				type: 'document',
				title: doc.display_name || doc.original_filename,
				contextId: doc.id,
				tokenCount: doc.word_count ? doc.word_count * 2 : undefined
			}];
		}
	}}
/>

<!-- Hybrid Search Panel -->
<HybridSearchPanel
	bind:show={showHybridSearchPanel}
	workspaceId={$currentWorkspaceId ?? undefined}
	on:addToContext={(event) => {
		const { result, query } = event.detail;
		// Add search result to active resources
		activeResources = [...activeResources, {
			id: result.context_id,
			type: 'search_result',
			title: result.context_name,
			contextId: result.context_id,
			tokenCount: Math.ceil(result.content.length / 4) // Rough estimate
		}];

		// Add context to selected contexts if not already there
		if (!selectedContextIds.includes(result.context_id)) {
			selectedContextIds = [...selectedContextIds, result.context_id];
		}

		// Optionally add a message to input showing what was added
		if (inputValue.trim() === '') {
			inputValue = `Using context from "${result.context_name}": ${query}`;
		}

		console.log('Added search result to context:', result);
	}}
/>

<style>
	@keyframes blink {
		0%, 50% { opacity: 1; }
		51%, 100% { opacity: 0; }
	}

	:global(.cursor-blink) {
		animation: blink 1s step-end infinite;
	}

	/* Streaming cursor - appears inline with text */
	:global(.streaming-cursor) {
		display: inline-block;
		width: 2px;
		height: 1em;
		background-color: #3b82f6;
		margin-left: 2px;
		vertical-align: text-bottom;
		animation: blink 1s step-end infinite;
		border-radius: 1px;
	}

	:global(.dark) .streaming-cursor {
		background-color: #60a5fa;
	}

	/* Ensure cursor stays inline with last text element */
	:global(.streaming-content) :global(p:last-child),
	:global(.streaming-content) :global(div:last-child),
	:global(.streaming-content) :global(span:last-child) {
		display: inline;
	}

	/* Force streaming content text to white in dark mode */
	:global(.dark) .streaming-content,
	:global(.dark) .streaming-content * {
		color: #f5f5f7 !important;
	}

	:global(.dark) .streaming-content strong,
	:global(.dark) .streaming-content b,
	:global(.dark) .streaming-content .chat-bold {
		color: #ffffff !important;
	}

	/* ===== DARK MODE FOR CHAT PAGE ===== */
	/* Window container */
	:global(.dark) .window-container,
	:global(.dark) .chat-container {
		background: #1c1c1e !important;
	}

	/* Header bar */
	:global(.dark) .chat-header {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	/* Main input area */
	:global(.dark) .chat-input-container {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	:global(.dark) .chat-input,
	:global(.dark) .message-input {
		background: #3a3a3c !important;
		border-color: rgba(255, 255, 255, 0.15) !important;
		color: #f5f5f7 !important;
	}

	:global(.dark) .chat-input:focus,
	:global(.dark) .message-input:focus {
		border-color: #0A84FF !important;
	}

	:global(.dark) .chat-input::placeholder,
	:global(.dark) .message-input::placeholder {
		color: #6e6e73 !important;
	}

	/* Messages area */
	:global(.dark) .messages-area {
		background: #1c1c1e !important;
	}

	/* User message bubbles */
	:global(.dark) .user-message,
	:global(.dark) .message.user {
		background: #0A84FF !important;
		color: white !important;
	}

	/* Assistant message bubbles */
	:global(.dark) .assistant-message,
	:global(.dark) .message.assistant {
		background: #2c2c2e !important;
		color: #f5f5f7 !important;
	}

	/* Buttons and icons */
	:global(.dark) .chat-btn,
	:global(.dark) .action-btn {
		color: #a1a1a6 !important;
	}

	:global(.dark) .chat-btn:hover,
	:global(.dark) .action-btn:hover {
		background: #3a3a3c !important;
		color: #f5f5f7 !important;
	}

	/* Header toggle buttons */
	.header-toggle-btn {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.625rem;
		font-size: 0.8125rem;
		font-weight: 500;
		border-radius: 0.5rem;
		transition: all 0.15s ease;
		flex-shrink: 0;
		color: #6b7280;
		background: transparent;
		border: none;
		cursor: pointer;
	}

	.header-toggle-btn:hover {
		color: #374151;
		background: rgba(0, 0, 0, 0.05);
	}

	.header-toggle-btn.active {
		color: #1f2937;
		background: rgba(0, 0, 0, 0.08);
	}

	.header-toggle-btn.warning {
		color: #b45309;
		background: #fef3c7;
		border: 1px solid #fcd34d;
	}

	.header-toggle-btn.warning:hover {
		background: #fde68a;
	}

	.header-toggle-badge {
		padding: 0.125rem 0.375rem;
		font-size: 0.6875rem;
		font-weight: 600;
		border-radius: 9999px;
		background: rgba(0, 0, 0, 0.08);
		color: #4b5563;
	}

	:global(.dark) .header-toggle-btn {
		color: #9ca3af;
	}

	:global(.dark) .header-toggle-btn:hover {
		color: #e5e7eb;
		background: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .header-toggle-btn.active {
		color: #f3f4f6;
		background: rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .header-toggle-btn.warning {
		color: #fbbf24;
		background: rgba(251, 191, 36, 0.15);
		border: 1px solid rgba(251, 191, 36, 0.3);
	}

	:global(.dark) .header-toggle-btn.warning:hover {
		background: rgba(251, 191, 36, 0.25);
	}

	:global(.dark) .header-toggle-badge {
		background: rgba(255, 255, 255, 0.12);
		color: #d1d5db;
	}

	/* Custom sidebar scrollbar */
	.sidebar-scroll {
		scrollbar-width: thin;
		scrollbar-color: transparent transparent;
	}

	.sidebar-scroll:hover {
		scrollbar-color: rgba(0, 0, 0, 0.15) transparent;
	}

	.sidebar-scroll::-webkit-scrollbar {
		width: 6px;
	}

	.sidebar-scroll::-webkit-scrollbar-track {
		background: transparent;
	}

	.sidebar-scroll::-webkit-scrollbar-thumb {
		background: transparent;
		border-radius: 3px;
	}

	.sidebar-scroll:hover::-webkit-scrollbar-thumb {
		background: rgba(0, 0, 0, 0.15);
	}

	:global(.dark) .sidebar-scroll:hover {
		scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
	}

	:global(.dark) .sidebar-scroll:hover::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.2);
	}

	/* Suggestion chips */
	:global(.dark) .suggestion-chip {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.15) !important;
		color: #f5f5f7 !important;
	}

	:global(.dark) .suggestion-chip:hover {
		background: #3a3a3c !important;
		border-color: rgba(255, 255, 255, 0.25) !important;
	}

	/* Dropdowns */
	:global(.dark) .dropdown,
	:global(.dark) .context-dropdown,
	:global(.dark) .model-dropdown {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4) !important;
	}

	:global(.dark) .dropdown-item:hover {
		background: #3a3a3c !important;
	}

	/* Sidebar (conversations list) */
	:global(.dark) .conversations-sidebar {
		background: #1c1c1e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	:global(.dark) .conversation-item {
		color: #f5f5f7 !important;
	}

	:global(.dark) .conversation-item:hover {
		background: #2c2c2e !important;
	}

	:global(.dark) .conversation-item.active {
		background: #3a3a3c !important;
	}

	/* Artifacts panel */
	:global(.dark) .artifacts-panel {
		background: #1c1c1e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	:global(.dark) .artifact-card {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	/* Empty state */
	:global(.dark) .empty-state {
		color: #a1a1a6 !important;
	}

	:global(.dark) .empty-state h2 {
		color: #f5f5f7 !important;
	}

	/* Modals */
	:global(.dark) .modal-content {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	/* ChatGPT-style input area */
	.chat-input-area {
		background: transparent;
	}

	.chat-input-box {
		transition: all 0.2s ease;
	}

	:global(.dark) .chat-input-area {
		background: transparent !important;
	}

	:global(.dark) .chat-input-box {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.15) !important;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3) !important;
	}

	:global(.dark) .chat-input-box:focus-within {
		border-color: rgba(255, 255, 255, 0.25) !important;
	}

	/* Input area container in dark mode */
	:global(.dark) .bg-gray-50 {
		background: #1c1c1e !important;
	}

	:global(.dark) .border-gray-100 {
		border-color: transparent !important;
	}

	:global(.dark) .rounded-3xl.bg-white,
	:global(.dark) .rounded-2xl.bg-white {
		background: #2c2c2e !important;
		border-color: rgba(255, 255, 255, 0.15) !important;
	}

	:global(.dark) .text-gray-900 {
		color: #f5f5f7 !important;
	}

	:global(.dark) .placeholder-gray-400::placeholder {
		color: #6e6e73 !important;
	}

	/* Usage stats dark mode */
	:global(.dark) .border-l.border-gray-200 {
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	:global(.dark) .text-gray-400 {
		color: #6e6e73 !important;
	}

	:global(.dark) .text-gray-300 {
		color: #48484a !important;
	}

	/* Assistant message text dark mode */
	:global(.dark) .text-gray-800 {
		color: #e5e5e7 !important;
	}

	:global(.dark) .prose {
		color: #e5e5e7;
	}

	:global(.dark) .prose p,
	:global(.dark) .prose li,
	:global(.dark) .prose span {
		color: #e5e5e7;
	}

	/* Prose dark mode - override Tailwind Typography invert colors */
	:global(.dark) .prose,
	:global(.prose.dark\:prose-invert) {
		--tw-prose-body: #f5f5f7 !important;
		--tw-prose-headings: #ffffff !important;
		--tw-prose-bold: #ffffff !important;
		--tw-prose-links: #60a5fa !important;
		--tw-prose-code: #e879f9 !important;
		--tw-prose-quotes: #d1d5db !important;
		--tw-prose-quote-borders: #4b5563 !important;
		--tw-prose-counters: #9ca3af !important;
		--tw-prose-bullets: #9ca3af !important;
	}

	/* Explicit overrides removed for cleanup - handled by generic rule below */
	
	/* Generic dark mode override */
	:global(.dark) .prose * {
		color: #f5f5f7 !important;
	}

	/* Override Tailwind v4 prose colors directly */
	:global(.dark) [class*="prose"] {
		--tw-prose-body: #f5f5f7 !important;
		--tw-prose-headings: #ffffff !important;
		--tw-prose-lead: #e5e5e7 !important;
		--tw-prose-links: #60a5fa !important;
		--tw-prose-bold: #ffffff !important;
		--tw-prose-counters: #9ca3af !important;
		--tw-prose-bullets: #9ca3af !important;
		--tw-prose-hr: rgba(255, 255, 255, 0.12) !important;
		--tw-prose-quotes: #d1d5db !important;
		--tw-prose-quote-borders: #4b5563 !important;
		--tw-prose-captions: #9ca3af !important;
		--tw-prose-code: #e879f9 !important;
		--tw-prose-pre-code: #f5f5f7 !important;
		--tw-prose-pre-bg: #1c1c1e !important;
		--tw-prose-th-borders: rgba(255, 255, 255, 0.12) !important;
		--tw-prose-td-borders: rgba(255, 255, 255, 0.08) !important;
		--tw-prose-invert-body: #f5f5f7 !important;
		--tw-prose-invert-headings: #ffffff !important;
		--tw-prose-invert-lead: #e5e5e7 !important;
		--tw-prose-invert-links: #60a5fa !important;
		--tw-prose-invert-bold: #ffffff !important;
		--tw-prose-invert-counters: #9ca3af !important;
		--tw-prose-invert-bullets: #9ca3af !important;
		--tw-prose-invert-hr: rgba(255, 255, 255, 0.12) !important;
		--tw-prose-invert-quotes: #d1d5db !important;
		--tw-prose-invert-quote-borders: #4b5563 !important;
		--tw-prose-invert-captions: #9ca3af !important;
		--tw-prose-invert-code: #e879f9 !important;
		--tw-prose-invert-pre-code: #f5f5f7 !important;
		--tw-prose-invert-pre-bg: #0d0d0f !important;
		--tw-prose-invert-th-borders: rgba(255, 255, 255, 0.12) !important;
		--tw-prose-invert-td-borders: rgba(255, 255, 255, 0.08) !important;
	}

	/* Artifact cards dark mode */
	:global(.dark) .from-blue-50,
	:global(.dark) .to-purple-50 {
		--tw-gradient-from: rgba(30, 41, 59, 0.5) !important;
		--tw-gradient-to: rgba(30, 30, 46, 0.5) !important;
	}

	:global(.dark) .border-blue-200 {
		border-color: rgba(59, 130, 246, 0.3) !important;
	}

	:global(.dark) .bg-gray-50 {
		background-color: #1c1c1e !important;
	}

	:global(.dark) .border-gray-200 {
		border-color: rgba(255, 255, 255, 0.12) !important;
	}

	/* Recording indicator - pulsing red dot */
	.recording-indicator {
		width: 12px;
		height: 12px;
		background-color: #ef4444;
		border-radius: 50%;
		animation: pulse-recording 1.5s ease-in-out infinite;
	}

	@keyframes pulse-recording {
		0%, 100% {
			transform: scale(1);
			opacity: 1;
		}
		50% {
			transform: scale(1.2);
			opacity: 0.7;
		}
	}

	/* Voice waveform animation */
	.voice-waveform {
		display: flex;
		align-items: center;
		gap: 3px;
		height: 32px;
	}

	.voice-waveform span {
		width: 4px;
		background-color: #ef4444;
		border-radius: 2px;
		animation: waveform 0.8s ease-in-out infinite;
	}

	.voice-waveform span:nth-child(1) { animation-delay: 0s; height: 8px; }
	.voice-waveform span:nth-child(2) { animation-delay: 0.1s; height: 16px; }
	.voice-waveform span:nth-child(3) { animation-delay: 0.2s; height: 24px; }
	.voice-waveform span:nth-child(4) { animation-delay: 0.3s; height: 16px; }
	.voice-waveform span:nth-child(5) { animation-delay: 0.4s; height: 8px; }

	@keyframes waveform {
		0%, 100% {
			transform: scaleY(1);
		}
		50% {
			transform: scaleY(1.8);
		}
	}

	/* ===== CHATGPT-STYLE MARKDOWN FORMATTING ===== */

	/* Section dividers - subtle horizontal line */
	:global(.chat-section-divider) {
		height: 1px;
		background: linear-gradient(to right, transparent, #e5e5e5 20%, #e5e5e5 80%, transparent);
		margin: 1.5rem 0 1rem 0;
	}

	:global(.dark) .chat-section-divider {
		background: linear-gradient(to right, transparent, rgba(255, 255, 255, 0.12) 20%, rgba(255, 255, 255, 0.12) 80%, transparent);
	}

	/* Major section headers (1. Title, 2. Title) */
	:global(.chat-section-header) {
		font-size: 1.125rem;
		font-weight: 600;
		color: #1a1a1a;
		margin: 0.5rem 0 1rem 0;
		line-height: 1.4;
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
	}

	:global(.dark) .chat-section-header {
		color: #f5f5f7;
	}

	:global(.chat-section-number) {
		font-weight: 700;
		color: #1a1a1a;
	}

	:global(.dark) .chat-section-number {
		color: #e5e5e7;
	}

	/* Sub-section headers (A. Title, B. Title) */
	:global(.chat-subsection-header) {
		font-size: 1rem;
		font-weight: 600;
		color: #2d2d2d;
		margin: 1rem 0 0.75rem 0;
		line-height: 1.4;
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
	}

	:global(.dark) .chat-subsection-header {
		color: #e5e5e7;
	}

	:global(.chat-subsection-letter) {
		font-weight: 700;
		color: #1a1a1a;
	}

	:global(.dark) .chat-subsection-letter {
		color: #e5e5e7;
	}

	/* Regular markdown headers */
	:global(.chat-h2) {
		font-size: 1.25rem;
		font-weight: 600;
		color: #1a1a1a;
		margin: 1.25rem 0 0.75rem 0;
		line-height: 1.3;
	}

	:global(.chat-h3) {
		font-size: 1.1rem;
		font-weight: 600;
		color: #2d2d2d;
		margin: 1rem 0 0.6rem 0;
		line-height: 1.3;
	}

	:global(.chat-h4) {
		font-size: 1rem;
		font-weight: 600;
		color: #3d3d3d;
		margin: 0.875rem 0 0.5rem 0;
		line-height: 1.3;
	}

	:global(.dark) .chat-h2,
	:global(.dark) .chat-h3,
	:global(.dark) .chat-h4 {
		color: #f5f5f7;
	}

	/* Code blocks */
	:global(.chat-code-block) {
		background: #1e1e1e;
		border-radius: 0.75rem;
		padding: 1rem;
		margin: 1rem 0;
		overflow-x: auto;
		font-family: 'SF Mono', 'Fira Code', 'Monaco', 'Consolas', monospace;
		font-size: 0.875rem;
		line-height: 1.6;
		border: 1px solid #2d2d2d;
	}

	:global(.chat-code-block code) {
		color: #d4d4d4;
		background: transparent;
		padding: 0;
		font-size: inherit;
	}

	:global(.dark) .chat-code-block {
		background: #0d0d0d;
		border-color: rgba(255, 255, 255, 0.08);
	}

	/* Inline code */
	:global(.chat-inline-code) {
		background: #f3f4f6;
		color: #e11d48;
		padding: 0.125rem 0.375rem;
		border-radius: 0.375rem;
		font-family: 'SF Mono', 'Fira Code', 'Monaco', 'Consolas', monospace;
		font-size: 0.875em;
	}

	:global(.dark) .chat-inline-code {
		background: rgba(255, 255, 255, 0.1);
		color: #fb7185;
	}

	/* Bold text */
	:global(.chat-bold) {
		font-weight: 600;
		color: inherit;
	}

	:global(.dark) .chat-bold {
		color: #ffffff !important;
	}

	/* Italic text */
	:global(.chat-italic) {
		font-style: italic;
		color: inherit;
	}

	:global(.dark) .chat-italic {
		color: #f5f5f7 !important;
	}

	/* Numbered list items */
	:global(.chat-list-item) {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin: 0.625rem 0;
		padding-left: 0.25rem;
	}

	:global(.chat-list-number) {
		font-weight: 600;
		color: #1a1a1a;
		min-width: 1.5rem;
		flex-shrink: 0;
	}

	:global(.dark) .chat-list-number {
		color: #e5e5e7;
	}

	:global(.chat-list-content) {
		flex: 1;
		line-height: 1.6;
	}

	/* Bullet items */
	:global(.chat-bullet-item) {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin: 0.5rem 0;
		padding-left: 0.25rem;
	}

	:global(.chat-bullet) {
		color: #9ca3af;
		font-size: 0.75rem;
		margin-top: 0.375rem;
	}

	:global(.dark) .chat-bullet {
		color: #6b7280;
	}

	/* Nested bullets */
	:global(.chat-nested-bullet) {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		margin: 0.375rem 0 0.375rem 1.5rem;
		font-size: 0.9375rem;
	}

	/* Labels like "Outcome:", "Example:" */
	:global(.chat-label) {
		margin: 1rem 0 0.5rem 0;
		line-height: 1.6;
	}

	:global(.chat-label strong) {
		color: #4b5563;
		font-weight: 600;
	}

	/* CSS Cleanup: Removed specific unused .dark selector for chat-label strong */

	/* Links */
	:global(.chat-link) {
		color: #3b82f6;
		text-decoration: underline;
		text-underline-offset: 2px;
		transition: color 0.15s;
	}

	:global(.chat-link:hover) {
		color: #2563eb;
	}

	:global(.dark) .chat-link {
		color: #60a5fa;
	}

	:global(.dark) .chat-link:hover {
		color: #93c5fd;
	}

	/* Paragraphs */
	:global(.chat-paragraph) {
		margin: 0.75rem 0;
		line-height: 1.7;
	}

	:global(.dark) .chat-paragraph {
		color: #f5f5f7 !important;
	}

	:global(.chat-paragraph:first-child) {
		margin-top: 0;
	}

	:global(.chat-paragraph:last-child) {
		margin-bottom: 0;
	}

	/* Fix overall message content */
	:global(.prose) .chat-section-header,
	:global(.prose) .chat-subsection-header,
	:global(.prose) .chat-h2,
	:global(.prose) .chat-h3,
	:global(.prose) .chat-h4 {
		margin-top: 1rem;
	}

	:global(.prose) .chat-section-divider + .chat-section-header {
		margin-top: 0;
	}
</style>




