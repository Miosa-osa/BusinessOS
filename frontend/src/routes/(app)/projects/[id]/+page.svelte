<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api, type Project, type Task, type CreateTaskData, type ContextListItem, type ClientListResponse, type TeamMemberListResponse, type Context } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { Dialog, Popover } from 'bits-ui';
	import { editor, wordCount, type EditorBlock } from '$lib/stores/editor';
	import { contexts } from '$lib/stores/contexts';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';

	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let linkedDocuments = $state<ContextListItem[]>([]);
	let clients = $state<ClientListResponse[]>([]);
	let teamMembers = $state<TeamMemberListResponse[]>([]);
	let isLoading = $state(true);
	let error = $state('');
	let showEditDialog = $state(false);
	let showDeleteConfirm = $state(false);
	let showAddTask = $state(false);
	let showEditTask = $state(false);
	let showLinkDocument = $state(false);
	let showLinkClient = $state(false);
	let showAssignTeam = $state(false);
	let isSaving = $state(false);
	let newNote = $state('');
	let isAddingNote = $state(false);
	let activeTab = $state<'overview' | 'tasks' | 'documents' | 'notes'>('overview');
	let editingTask = $state<Task | null>(null);

	// Drag & Drop state
	let draggedTask = $state<Task | null>(null);
	let dragOverTask = $state<Task | null>(null);

	// Available items for linking
	let availableDocuments = $state<ContextListItem[]>([]);
	let loadingAvailable = $state(false);

	// Document Editor Panel State
	type DocumentPanelMode = 'hidden' | 'side' | 'center' | 'full';
	let documentPanelMode = $state<DocumentPanelMode>('hidden');
	let selectedDocument = $state<Context | null>(null);
	let selectedDocumentId = $state<string | null>(null);
	let loadingDocument = $state(false);
	let documentPanelWidth = $state(550);
	let isResizingPanel = $state(false);
	let documentTitle = $state('');
	let autoSaveTimer: ReturnType<typeof setTimeout>;

	// New task form
	let newTask = $state<CreateTaskData>({
		title: '',
		description: '',
		priority: 'medium',
		due_date: '',
		estimated_hours: undefined,
		start_date: undefined,
		parent_task_id: undefined,
		assignee_id: undefined
	});

	// Edit form state
	let editForm = $state({
		name: '',
		description: '',
		status: 'active' as 'active' | 'paused' | 'completed' | 'archived',
		priority: 'medium' as 'critical' | 'high' | 'medium' | 'low',
		client_name: '',
		project_type: 'internal'
	});

	const projectId = $derived($page.params.id);

	onMount(async () => {
		await Promise.all([
			loadProject(),
			loadTasks(),
			loadClients(),
			loadTeamMembers()
		]);
	});

	onDestroy(() => {
		if (autoSaveTimer) clearTimeout(autoSaveTimer);
		editor.reset();
	});

	// Auto-save document with debounce
	$effect(() => {
		if ($editor.isDirty && selectedDocument && documentPanelMode !== 'hidden') {
			if (autoSaveTimer) clearTimeout(autoSaveTimer);
			autoSaveTimer = setTimeout(async () => {
				await saveDocument();
			}, 1500);
		}
	});

	async function saveDocument() {
		if (!selectedDocument || $editor.isSaving) return;
		editor.setSaving(true);
		try {
			await contexts.updateBlocks(selectedDocument.id, $editor.blocks, $wordCount);
			editor.markSaved();
		} catch (e) {
			console.error('Failed to save:', e);
			editor.setSaving(false);
		}
	}

	async function updateDocumentTitle() {
		const doc = selectedDocument;
		if (!doc || documentTitle === doc.name) return;
		try {
			await contexts.updateContext(doc.id, { name: documentTitle });
			// Update in available documents list
			availableDocuments = availableDocuments.map(d =>
				d.id === doc.id ? { ...d, name: documentTitle } : d
			);
		} catch (e) {
			console.error('Failed to update title:', e);
		}
	}

	async function openDocument(docId: string, mode: DocumentPanelMode = 'side') {
		if (selectedDocumentId === docId && documentPanelMode !== 'hidden') {
			// Already open, maybe switch mode
			documentPanelMode = mode;
			return;
		}

		loadingDocument = true;
		selectedDocumentId = docId;
		documentPanelMode = mode;

		try {
			const doc = await contexts.loadContext(docId);
			selectedDocument = doc;
			documentTitle = doc.name;
			editor.initialize(doc.blocks);
		} catch (e) {
			console.error('Failed to load document:', e);
			closeDocumentPanel();
		} finally {
			loadingDocument = false;
		}
	}

	function closeDocumentPanel() {
		documentPanelMode = 'hidden';
		selectedDocument = null;
		selectedDocumentId = null;
		editor.reset();
	}

	function handlePanelResize(e: MouseEvent) {
		if (!isResizingPanel) return;
		const newWidth = window.innerWidth - e.clientX;
		documentPanelWidth = Math.min(Math.max(newWidth, 400), 900);
	}

	function startPanelResize(e: MouseEvent) {
		e.preventDefault();
		isResizingPanel = true;
		document.addEventListener('mousemove', handlePanelResize);
		document.addEventListener('mouseup', stopPanelResize);
	}

	function stopPanelResize() {
		isResizingPanel = false;
		document.removeEventListener('mousemove', handlePanelResize);
		document.removeEventListener('mouseup', stopPanelResize);
	}

	function addNewBlockAtEnd() {
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			const newBlockId = editor.addBlockAfter(lastBlock.id);
			setTimeout(() => {
				const blockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				blockEl?.focus();
			}, 10);
		}
	}

	async function loadProject() {
		if (!projectId) {
			error = 'No project ID provided';
			return;
		}
		isLoading = true;
		error = '';
		try {
			project = await api.getProject(projectId);
			if (project) {
				// Ensure notes is always an array
				if (!project.notes) {
					project.notes = [];
				}
				editForm = {
					name: project.name,
					description: project.description || '',
					status: project.status,
					priority: project.priority,
					client_name: project.client_name || '',
					project_type: project.project_type
				};
			}
		} catch (err) {
			error = 'Failed to load project';
			console.error('Error loading project:', err);
		} finally {
			isLoading = false;
		}
	}

	async function loadTasks() {
		try {
			tasks = await api.getTasks({ projectId });
		} catch (err) {
			console.error('Error loading tasks:', err);
		}
	}

	async function loadLinkedDocuments() {
		try {
			// Load contexts that might be linked to this project
			const allContexts = await api.getContexts();
			linkedDocuments = allContexts.filter(c => c.type === 'document');
		} catch (err) {
			console.error('Error loading documents:', err);
		}
	}

	async function loadClients() {
		try {
			clients = await api.getClients();
		} catch (err) {
			console.error('Error loading clients:', err);
		}
	}

	async function loadTeamMembers() {
		try {
			teamMembers = await api.getTeamMembers();
		} catch (err) {
			console.error('Error loading team:', err);
		}
	}

	async function loadAvailableDocuments() {
		loadingAvailable = true;
		try {
			const contexts = await api.getContexts();
			availableDocuments = contexts.filter(c => c.type === 'document');
		} catch (err) {
			console.error('Error loading documents:', err);
		} finally {
			loadingAvailable = false;
		}
	}

	async function handleSave() {
		if (!project) return;
		isSaving = true;
		try {
			await api.updateProject(project.id, editForm);
			await loadProject();
			showEditDialog = false;
		} catch (err) {
			console.error('Error saving project:', err);
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete() {
		if (!project) return;
		try {
			await api.deleteProject(project.id);
			goto('/projects' + embedSuffix);
		} catch (err) {
			console.error('Error deleting project:', err);
		}
	}

	async function handleAddNote() {
		if (!project || !newNote.trim()) return;
		isAddingNote = true;
		try {
			await api.addProjectNote(project.id, newNote);
			await loadProject();
			newNote = '';
		} catch (err) {
			console.error('Error adding note:', err);
		} finally {
			isAddingNote = false;
		}
	}

	async function handleCreateTask(e: Event) {
		e.preventDefault();
		if (!project || !newTask.title.trim()) return;
		try {
			await api.createTask({
				...newTask,
				project_id: project.id
			});
			await loadTasks();
			showAddTask = false;
			// Reset form
			newTask = {
				title: '',
				description: '',
				priority: 'medium',
				due_date: '',
				estimated_hours: undefined,
				start_date: undefined,
				parent_task_id: undefined,
				assignee_id: undefined
			};
		} catch (err) {
			console.error('Error creating task:', err);
		}
	}

	async function handleToggleTask(taskId: string) {
		try {
			await api.toggleTask(taskId);
			await loadTasks();
		} catch (err) {
			console.error('Error toggling task:', err);
		}
	}

	async function handleDeleteTask(taskId: string) {
		try {
			await api.deleteTask(taskId);
			await loadTasks();
		} catch (err) {
			console.error('Error deleting task:', err);
		}
	}

	function handleEditTask(task: Task) {
		editingTask = task;
		showEditTask = true;
	}

	async function handleUpdateTask(e: Event) {
		e.preventDefault();
		if (!editingTask) return;

		try {
			await api.updateTask(editingTask.id, {
				title: editingTask.title,
				description: editingTask.description || '',
				priority: editingTask.priority,
				status: editingTask.status,
				due_date: editingTask.due_date || ''
			});
			await loadTasks();
			showEditTask = false;
			editingTask = null;
		} catch (err) {
			console.error('Error updating task:', err);
		}
	}

	// Drag & Drop handlers
	function handleDragStart(e: DragEvent, task: Task) {
		draggedTask = task;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function handleDragOver(e: DragEvent, task: Task) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		if (draggedTask && draggedTask.id !== task.id && draggedTask.status === task.status) {
			dragOverTask = task;
		}
	}

	function handleDragLeave() {
		dragOverTask = null;
	}

	async function handleDrop(e: DragEvent, targetTask: Task) {
		e.preventDefault();
		dragOverTask = null;

		const currentDraggedTask = draggedTask;
		if (!currentDraggedTask || currentDraggedTask.id === targetTask.id || currentDraggedTask.status !== targetTask.status) {
			return;
		}

		try {
			// Get tasks in the same status group
			const statusTasks = tasks.filter(t => t.status === targetTask.status);
			const draggedIndex = statusTasks.findIndex(t => t.id === currentDraggedTask.id);
			const targetIndex = statusTasks.findIndex(t => t.id === targetTask.id);

			if (draggedIndex === -1 || targetIndex === -1) return;

			// Reorder the tasks array optimistically
			const newStatusTasks = [...statusTasks];
			const [removed] = newStatusTasks.splice(draggedIndex, 1);
			newStatusTasks.splice(targetIndex, 0, removed);

			// Update positions in backend
			const updatePromises = newStatusTasks.map((task, index) =>
				api.updateTask(task.id, { position: index })
			);

			await Promise.all(updatePromises);
			await loadTasks();
		} catch (err) {
			console.error('Error reordering tasks:', err);
			await loadTasks(); // Reload to revert optimistic update
		}
	}

	function handleDragEnd() {
		draggedTask = null;
		dragOverTask = null;
	}

	async function updateClientLink(clientId: string | null) {
		if (!project) return;
		try {
			const selectedClient = clientId ? clients.find(c => c.id === clientId) : null;
			await api.updateProject(project.id, {
				client_name: selectedClient?.name || ''
			});
			await loadProject();
			showLinkClient = false;
		} catch (err) {
			console.error('Error updating client:', err);
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
			case 'paused': return 'bg-amber-100 text-amber-700 border-amber-200';
			case 'completed': return 'bg-blue-100 text-blue-700 border-blue-200';
			case 'archived': return 'bg-gray-100 text-gray-600 border-gray-200';
			default: return 'bg-gray-100 text-gray-600 border-gray-200';
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'active': return 'M13 10V3L4 14h7v7l9-11h-7z';
			case 'paused': return 'M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z';
			case 'completed': return 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z';
			default: return 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4';
		}
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'critical': return 'text-red-600 bg-red-50';
			case 'high': return 'text-orange-600 bg-orange-50';
			case 'medium': return 'text-yellow-600 bg-yellow-50';
			case 'low': return 'text-green-600 bg-green-50';
			default: return 'text-gray-600 bg-gray-50';
		}
	}

	function getTypeLabel(type: string) {
		switch (type) {
			case 'internal': return 'Internal';
			case 'client_work': return 'Client Work';
			case 'learning': return 'Learning';
			default: return type;
		}
	}

	function getTypeIcon(type: string) {
		switch (type) {
			case 'internal': return '🏢';
			case 'client_work': return '👥';
			case 'learning': return '📚';
			default: return '📁';
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatTime(dateStr: string) {
		return new Date(dateStr).toLocaleTimeString(undefined, {
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	// Task counts
	let completedTasks = $derived(tasks.filter(t => t.status === 'done').length);
	let totalTasks = $derived(tasks.length);
	let taskProgress = $derived(totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0);

	// Get linked client object
	let linkedClient = $derived(project?.client_name ? clients.find(c => c.name === project?.client_name) : null);
</script>

<div class="h-full flex flex-col bg-gray-50">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else if error || !project}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center">
				<p class="text-gray-500 mb-4">{error || 'Project not found'}</p>
				<a href="/projects{embedSuffix}" class="btn-pill btn-pill-secondary">Back to Projects</a>
			</div>
		</div>
	{:else}
		<!-- Header -->
		<div class="bg-white border-b border-gray-200">
			<div class="px-6 py-4">
				<div class="flex items-center gap-2 text-sm text-gray-500 mb-3">
					<a href="/projects{embedSuffix}" class="hover:text-gray-700 flex items-center gap-1">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
						Projects
					</a>
				</div>
				<div class="flex items-start justify-between">
					<div class="flex items-start gap-4">
						<!-- Project Icon -->
						<div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-indigo-600 flex items-center justify-center text-white text-xl">
							{getTypeIcon(project.project_type)}
						</div>
						<div>
							<div class="flex items-center gap-3 mb-1">
								<h1 class="text-xl font-semibold text-gray-900">{project.name}</h1>
								<span class="text-xs font-medium px-2.5 py-1 rounded-full border {getStatusColor(project.status)} flex items-center gap-1">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getStatusIcon(project.status)} />
									</svg>
									{project.status}
								</span>
								<span class="text-xs font-medium px-2 py-0.5 rounded {getPriorityColor(project.priority)}">
									{project.priority}
								</span>
							</div>
							<div class="flex items-center gap-3 text-sm text-gray-500">
								<span class="flex items-center gap-1">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
									</svg>
									{getTypeLabel(project.project_type)}
								</span>

								<!-- Client Link -->
								<Popover.Root bind:open={showLinkClient}>
									<Popover.Trigger class="flex items-center gap-1 hover:text-gray-700 cursor-pointer">
										{#if project.client_name}
											<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
											</svg>
											<span class="text-blue-600">{project.client_name}</span>
										{:else}
											<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
											</svg>
											<span class="text-gray-400">Link client</span>
										{/if}
									</Popover.Trigger>
									<Popover.Content class="z-50 bg-white rounded-xl shadow-lg border border-gray-200 p-2 w-64 max-h-64 overflow-y-auto">
										<div class="text-xs font-medium text-gray-400 uppercase px-2 py-1 mb-1">Link to Client</div>
										{#if project.client_name}
											<button
												onclick={() => updateClientLink(null)}
												class="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-gray-50 text-red-600"
											>
												Remove link
											</button>
											<div class="border-t border-gray-100 my-1"></div>
										{/if}
										{#each clients as client}
											<button
												onclick={() => updateClientLink(client.id)}
												class="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-gray-50 flex items-center gap-2 {project.client_name === client.name ? 'bg-blue-50 text-blue-700' : ''}"
											>
												<span class="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs">
													{client.name.charAt(0)}
												</span>
												{client.name}
											</button>
										{/each}
										{#if clients.length === 0}
											<p class="text-xs text-gray-400 text-center py-4">No clients yet</p>
										{/if}
									</Popover.Content>
								</Popover.Root>

								<span class="text-gray-300">·</span>
								<span>Created {formatDate(project.created_at)}</span>
							</div>
						</div>
					</div>
					<div class="flex gap-2">
						<button onclick={() => showEditDialog = true} class="btn-pill btn-pill-secondary btn-pill-sm">
							<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							</svg>
							Edit
						</button>
						<button onclick={() => showDeleteConfirm = true} class="btn-pill btn-pill-danger btn-pill-sm">
							<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</button>
					</div>
				</div>

				<!-- Progress Bar -->
				{#if totalTasks > 0}
					<div class="mt-4">
						<div class="flex items-center justify-between text-sm mb-1">
							<span class="text-gray-600">Progress</span>
							<span class="font-medium text-gray-900">{completedTasks}/{totalTasks} tasks ({taskProgress}%)</span>
						</div>
						<div class="h-2 bg-gray-200 rounded-full overflow-hidden">
							<div
								class="h-full bg-gradient-to-r from-purple-500 to-indigo-600 rounded-full transition-all duration-300"
								style="width: {taskProgress}%"
							></div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Tabs -->
			<div class="px-6 flex gap-6 border-t border-gray-100">
				<button
					onclick={() => activeTab = 'overview'}
					class="py-3 text-sm font-medium border-b-2 transition-colors {activeTab === 'overview' ? 'border-gray-900 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700'}"
				>
					Overview
				</button>
				<button
					onclick={() => activeTab = 'tasks'}
					class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 {activeTab === 'tasks' ? 'border-gray-900 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700'}"
				>
					Tasks
					{#if totalTasks > 0}
						<span class="px-2 py-0.5 text-xs rounded-full {activeTab === 'tasks' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600'}">{totalTasks}</span>
					{/if}
				</button>
				<button
					onclick={() => { activeTab = 'documents'; loadAvailableDocuments(); }}
					class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 {activeTab === 'documents' ? 'border-gray-900 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700'}"
				>
					Documents
					{#if availableDocuments.length > 0}
						<span class="px-2 py-0.5 text-xs rounded-full {activeTab === 'documents' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600'}">{availableDocuments.length}</span>
					{/if}
				</button>
				<button
					onclick={() => activeTab = 'notes'}
					class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 {activeTab === 'notes' ? 'border-gray-900 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700'}"
				>
					Notes
					{#if project.notes && project.notes.length > 0}
						<span class="px-2 py-0.5 text-xs rounded-full {activeTab === 'notes' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600'}">{project.notes.length}</span>
					{/if}
				</button>
			</div>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-y-auto p-6">
			<div class="max-w-5xl mx-auto">
				{#if activeTab === 'overview'}
					<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
						<!-- Main Content -->
						<div class="lg:col-span-2 space-y-6">
							<!-- Description -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h2 class="text-lg font-medium text-gray-900 mb-3 flex items-center gap-2">
									<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
									</svg>
									Description
								</h2>
								{#if project.description}
									<p class="text-gray-600 whitespace-pre-wrap">{project.description}</p>
								{:else}
									<p class="text-gray-400 italic">No description added yet. Click Edit to add one.</p>
								{/if}
							</div>

							<!-- Recent Tasks -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<div class="flex items-center justify-between mb-4">
									<h2 class="text-lg font-medium text-gray-900 flex items-center gap-2">
										<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
										</svg>
										Tasks
									</h2>
									<button onclick={() => { activeTab = 'tasks'; showAddTask = true; }} class="text-sm text-purple-600 hover:text-purple-700 font-medium">
										+ Add Task
									</button>
								</div>
								{#if tasks.length === 0}
									<div class="text-center py-8">
										<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
											<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
											</svg>
										</div>
										<p class="text-gray-500 mb-2">No tasks yet</p>
										<button onclick={() => { activeTab = 'tasks'; showAddTask = true; }} class="btn-pill btn-pill-primary btn-pill-sm">
											Add First Task
										</button>
									</div>
								{:else}
									<div class="space-y-2">
										{#each tasks.slice(0, 5) as task}
											<div class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 group">
												<button
													onclick={() => handleToggleTask(task.id)}
													class="w-5 h-5 rounded border-2 flex items-center justify-center flex-shrink-0 transition-colors {task.status === 'done' ? 'bg-purple-600 border-purple-600 text-white' : 'border-gray-300 hover:border-purple-600'}"
												>
													{#if task.status === 'done'}
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
														</svg>
													{/if}
												</button>
												<div class="flex-1 min-w-0">
													<p class="text-sm {task.status === 'done' ? 'text-gray-400 line-through' : 'text-gray-900'}">{task.title}</p>
													{#if task.due_date}
														<p class="text-xs text-gray-400">Due {formatDate(task.due_date)}</p>
													{/if}
												</div>
												<span class="text-xs px-2 py-0.5 rounded {getPriorityColor(task.priority)}">{task.priority}</span>
											</div>
										{/each}
										{#if tasks.length > 5}
											<button onclick={() => activeTab = 'tasks'} class="text-sm text-purple-600 hover:text-purple-700 font-medium w-full text-center py-2">
												View all {tasks.length} tasks
											</button>
										{/if}
									</div>
								{/if}
							</div>
						</div>

						<!-- Sidebar -->
						<div class="space-y-6">
							<!-- Quick Actions -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h2 class="text-lg font-medium text-gray-900 mb-3">Quick Actions</h2>
								<div class="space-y-2">
									{#if project.status !== 'completed'}
										<button
											onclick={async () => {
												await api.updateProject(project!.id, { status: 'completed' });
												await loadProject();
											}}
											class="btn-pill btn-pill-secondary btn-pill-sm w-full justify-start"
										>
											<svg class="w-4 h-4 mr-2 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
											Mark Complete
										</button>
									{/if}
									{#if project.status === 'active'}
										<button
											onclick={async () => {
												await api.updateProject(project!.id, { status: 'paused' });
												await loadProject();
											}}
											class="btn-pill btn-pill-secondary btn-pill-sm w-full justify-start"
										>
											<svg class="w-4 h-4 mr-2 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
											Pause Project
										</button>
									{:else if project.status === 'paused'}
										<button
											onclick={async () => {
												await api.updateProject(project!.id, { status: 'active' });
												await loadProject();
											}}
											class="btn-pill btn-pill-secondary btn-pill-sm w-full justify-start"
										>
											<svg class="w-4 h-4 mr-2 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
											Resume Project
										</button>
									{/if}
									<button
										onclick={() => { activeTab = 'tasks'; showAddTask = true; }}
										class="btn-pill btn-pill-secondary btn-pill-sm w-full justify-start"
									>
										<svg class="w-4 h-4 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
										Add Task
									</button>
									<a
										href="/knowledge-v2{embedSuffix}"
										class="btn-pill btn-pill-secondary btn-pill-sm w-full justify-start"
									>
										<svg class="w-4 h-4 mr-2 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
										View Documents
									</a>
									{#if project.status !== 'archived'}
										<button
											onclick={async () => {
												await api.updateProject(project!.id, { status: 'archived' });
												await loadProject();
											}}
											class="btn-pill btn-pill-soft btn-pill-sm w-full justify-start"
										>
											<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
											</svg>
											Archive
										</button>
									{/if}
								</div>
							</div>

							<!-- Details -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h2 class="text-lg font-medium text-gray-900 mb-3">Details</h2>
								<dl class="space-y-3">
									<div>
										<dt class="text-xs text-gray-500 uppercase">Status</dt>
										<dd class="text-sm font-medium capitalize">{project.status}</dd>
									</div>
									<div>
										<dt class="text-xs text-gray-500 uppercase">Priority</dt>
										<dd class="text-sm font-medium capitalize {getPriorityColor(project.priority)} inline-block px-2 py-0.5 rounded">{project.priority}</dd>
									</div>
									<div>
										<dt class="text-xs text-gray-500 uppercase">Type</dt>
										<dd class="text-sm text-gray-900">{getTypeLabel(project.project_type)}</dd>
									</div>
									{#if project.client_name}
										<div>
											<dt class="text-xs text-gray-500 uppercase">Client</dt>
											<dd class="text-sm text-gray-900">{project.client_name}</dd>
										</div>
									{/if}
									<div>
										<dt class="text-xs text-gray-500 uppercase">Created</dt>
										<dd class="text-sm text-gray-900">{formatDate(project.created_at)}</dd>
									</div>
									<div>
										<dt class="text-xs text-gray-500 uppercase">Last Updated</dt>
										<dd class="text-sm text-gray-900">{formatDate(project.updated_at)}</dd>
									</div>
								</dl>
							</div>

							<!-- Team Members -->
							{#if teamMembers.length > 0}
								<div class="bg-white rounded-xl border border-gray-200 p-6">
									<div class="flex items-center justify-between mb-3">
										<h2 class="text-lg font-medium text-gray-900">Team</h2>
										<button onclick={() => showAssignTeam = true} class="text-sm text-purple-600 hover:text-purple-700">
											+ Assign
										</button>
									</div>
									<div class="space-y-2">
										{#each teamMembers.slice(0, 3) as member}
											<div class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50">
												<div class="w-8 h-8 rounded-full bg-gradient-to-br from-purple-400 to-indigo-500 flex items-center justify-center text-white text-xs font-medium">
													{member.name.split(' ').map(n => n[0]).join('').slice(0, 2)}
												</div>
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium text-gray-900 truncate">{member.name}</p>
													<p class="text-xs text-gray-400 truncate">{member.role}</p>
												</div>
											</div>
										{/each}
										{#if teamMembers.length > 3}
											<p class="text-xs text-gray-400 text-center">+{teamMembers.length - 3} more</p>
										{/if}
									</div>
								</div>
							{/if}
						</div>
					</div>

				{:else if activeTab === 'tasks'}
					<div class="space-y-4">
						<!-- Task Stats Cards -->
						<div class="grid grid-cols-4 gap-3">
							<div class="bg-white rounded-xl border border-gray-200 p-4">
								<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">To Do</p>
								<p class="text-2xl font-bold text-gray-900 mt-1">{tasks.filter(t => t.status === 'todo').length}</p>
							</div>
							<div class="bg-white rounded-xl border border-gray-200 p-4">
								<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">In Progress</p>
								<p class="text-2xl font-bold text-blue-600 mt-1">{tasks.filter(t => t.status === 'in_progress').length}</p>
							</div>
							<div class="bg-white rounded-xl border border-gray-200 p-4">
								<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">Done</p>
								<p class="text-2xl font-bold text-green-600 mt-1">{tasks.filter(t => t.status === 'done').length}</p>
							</div>
							<div class="bg-white rounded-xl border border-gray-200 p-4">
								<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">Completion</p>
								<p class="text-2xl font-bold text-purple-600 mt-1">{totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0}%</p>
							</div>
						</div>

						<!-- Tasks List -->
						<div class="bg-white rounded-xl border border-gray-200">
							<div class="p-4 border-b border-gray-100 flex items-center justify-between">
								<div class="flex items-center gap-2">
									<h2 class="text-lg font-medium text-gray-900">All Tasks</h2>
									<span class="text-sm text-gray-400">({totalTasks})</span>
								</div>
								<div class="flex items-center gap-2">
									<a href="/tasks{embedSuffix}" class="btn-pill btn-pill-secondary btn-pill-sm">
										<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
										</svg>
										Open Tasks
									</a>
									<button onclick={() => showAddTask = true} class="btn-pill btn-pill-primary btn-pill-sm">
										<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
										Add Task
									</button>
								</div>
							</div>

							{#if tasks.length === 0}
								<div class="text-center py-16">
									<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
										<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
										</svg>
									</div>
									<h3 class="text-lg font-medium text-gray-900 mb-1">No tasks yet</h3>
									<p class="text-gray-500 mb-4">Break down your project into manageable tasks</p>
									<button onclick={() => showAddTask = true} class="btn-pill btn-pill-primary">
										Add First Task
									</button>
								</div>
							{:else}
								<!-- Group by status -->
								{#each [
									{ status: 'todo', label: 'To Do', color: 'gray', tasks: tasks.filter(t => t.status === 'todo') },
									{ status: 'in_progress', label: 'In Progress', color: 'blue', tasks: tasks.filter(t => t.status === 'in_progress') },
									{ status: 'done', label: 'Done', color: 'green', tasks: tasks.filter(t => t.status === 'done') }
								].filter(g => g.tasks.length > 0) as group}
									<div class="border-b border-gray-100 last:border-b-0">
										<div class="px-4 py-2 bg-gray-50/50 flex items-center gap-2">
											<span class="w-2 h-2 rounded-full bg-{group.color}-500"></span>
											<span class="text-xs font-medium text-gray-600">{group.label}</span>
											<span class="text-xs text-gray-400">({group.tasks.length})</span>
										</div>
										<div class="divide-y divide-gray-100">
											{#each group.tasks as task}
												{@const isSubtask = !!task.parent_task_id}
												{@const assignee = task.assignee_id ? teamMembers.find(m => m.id === task.assignee_id) : null}
												<div
													class="flex items-center gap-4 p-4 hover:bg-gray-50 group/task transition-all {dragOverTask?.id === task.id ? 'border-t-2 border-purple-600' : ''} {draggedTask?.id === task.id ? 'opacity-50' : 'opacity-100'} {isSubtask ? 'pl-12 bg-gray-50/50' : ''}"
													draggable="true"
													ondragstart={(e) => handleDragStart(e, task)}
													ondragover={(e) => handleDragOver(e, task)}
													ondragleave={handleDragLeave}
													ondrop={(e) => handleDrop(e, task)}
													ondragend={handleDragEnd}
												>
													<!-- Subtask Indicator -->
													{#if isSubtask}
														<div class="absolute left-6 w-4 h-4 flex items-center justify-center">
															<svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
															</svg>
														</div>
													{/if}

													<!-- Drag Handle -->
													<div class="flex-shrink-0 cursor-move text-gray-400 hover:text-gray-600 opacity-0 group-hover/task:opacity-100 transition-opacity">
														<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8h16M4 16h16" />
														</svg>
													</div>
													<button
														onclick={() => handleToggleTask(task.id)}
														class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors {task.status === 'done' ? 'bg-green-600 border-green-600 text-white' : 'border-gray-300 hover:border-purple-600'}"
													>
														{#if task.status === 'done'}
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
															</svg>
														{/if}
													</button>
													<div class="flex-1 min-w-0">
														<p class="font-medium {task.status === 'done' ? 'text-gray-400 line-through' : 'text-gray-900'}">{task.title}</p>
														{#if task.description}
															<p class="text-sm text-gray-500 mt-0.5 line-clamp-1">{task.description}</p>
														{/if}
														<div class="flex items-center gap-3 mt-1 text-xs text-gray-400">
															{#if task.start_date}
																<span class="flex items-center gap-1">
																	<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
																	</svg>
																	Start: {formatDate(task.start_date)}
																</span>
															{/if}
															{#if task.due_date}
																<span class="flex items-center gap-1 {new Date(task.due_date) < new Date() && task.status !== 'done' ? 'text-red-500' : ''}">
																	<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
																	</svg>
																	Due: {formatDate(task.due_date)}
																</span>
															{/if}
															{#if task.estimated_hours}
																<span class="flex items-center gap-1">
																	<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
																	</svg>
																	{task.estimated_hours}h
																</span>
															{/if}
														</div>
													</div>
													<div class="flex items-center gap-2">
														<span class="text-xs px-2 py-1 rounded font-medium {getPriorityColor(task.priority)}">{task.priority}</span>
														{#if assignee}
															<span class="text-xs px-2 py-1 rounded bg-blue-100 text-blue-700 font-medium flex items-center gap-1">
																<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
																</svg>
																{assignee.name}
															</span>
														{/if}
													</div>
													<div class="flex items-center gap-1 opacity-0 group-hover/task:opacity-100 transition-opacity">
														<button
															onclick={() => handleEditTask(task)}
															class="p-2 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50 transition-colors"
															title="Edit task"
														>
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
															</svg>
														</button>
														<button
															onclick={() => handleDeleteTask(task.id)}
															class="p-2 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50 transition-colors"
															title="Delete task"
														>
															<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
															</svg>
														</button>
													</div>
												</div>
											{/each}
										</div>
									</div>
								{/each}
							{/if}
						</div>
					</div>

				{:else if activeTab === 'documents'}
					<div class="bg-white rounded-xl border border-gray-200">
						<div class="p-6 border-b border-gray-100 flex items-center justify-between">
							<div>
								<h2 class="text-lg font-medium text-gray-900">Documents</h2>
								<p class="text-sm text-gray-500 mt-0.5">Knowledge base documents for this project</p>
							</div>
							<div class="flex gap-2">
								<!-- View mode selector -->
								{#if documentPanelMode !== 'hidden'}
									<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden mr-2">
										<button
											onclick={() => documentPanelMode = 'side'}
											class="p-1.5 text-xs transition-colors {documentPanelMode === 'side' ? 'bg-gray-900 text-white' : 'text-gray-500 hover:bg-gray-100'}"
											title="Side panel"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
											</svg>
										</button>
										<button
											onclick={() => documentPanelMode = 'center'}
											class="p-1.5 text-xs transition-colors {documentPanelMode === 'center' ? 'bg-gray-900 text-white' : 'text-gray-500 hover:bg-gray-100'}"
											title="Center panel"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
											</svg>
										</button>
										<button
											onclick={() => documentPanelMode = 'full'}
											class="p-1.5 text-xs transition-colors {documentPanelMode === 'full' ? 'bg-gray-900 text-white' : 'text-gray-500 hover:bg-gray-100'}"
											title="Full screen"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
											</svg>
										</button>
									</div>
								{/if}
								<a href="/knowledge-v2{embedSuffix}" class="btn-pill btn-pill-primary btn-pill-sm">
									<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
									New Document
								</a>
							</div>
						</div>

						{#if loadingAvailable}
							<div class="flex items-center justify-center py-16">
								<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
							</div>
						{:else if availableDocuments.length === 0}
							<div class="text-center py-16">
								<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
									<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</div>
								<h3 class="text-lg font-medium text-gray-900 mb-1">No documents yet</h3>
								<p class="text-gray-500 mb-4">Create documents in the Knowledge Base to link them here</p>
								<a href="/knowledge-v2{embedSuffix}" class="btn-pill btn-pill-primary">
									Go to Knowledge Base
								</a>
							</div>
						{:else}
							<div class="p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
								{#each availableDocuments as doc}
									<button
										onclick={() => openDocument(doc.id, 'side')}
										class="text-left p-4 rounded-xl border border-gray-200 hover:shadow-md hover:border-gray-300 transition-all group {selectedDocumentId === doc.id ? 'ring-2 ring-purple-500 border-purple-500' : ''}"
									>
										<div class="flex items-start gap-3">
											<span class="text-2xl">{doc.icon || '📄'}</span>
											<div class="flex-1 min-w-0">
												<h4 class="text-sm font-medium text-gray-900 group-hover:text-gray-700 truncate">{doc.name}</h4>
												<p class="text-xs text-gray-400 mt-0.5">
													{Number(doc.word_count) > 0 ? `${Number(doc.word_count).toLocaleString()} words` : 'Empty'}
												</p>
												<p class="text-xs text-gray-400">Updated {formatDate(doc.updated_at)}</p>
											</div>
											{#if selectedDocumentId === doc.id}
												<span class="text-purple-500">
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
													</svg>
												</span>
											{/if}
										</div>
									</button>
								{/each}
							</div>
						{/if}
					</div>

				{:else if activeTab === 'notes'}
					<div class="bg-white rounded-xl border border-gray-200 p-6">
						<h2 class="text-lg font-medium text-gray-900 mb-4">Notes</h2>

						<!-- Add Note -->
						<div class="mb-6">
							<textarea
								bind:value={newNote}
								placeholder="Add a note..."
								class="input input-square resize-none w-full"
								rows="3"
							></textarea>
							<div class="flex justify-end mt-2">
								<button
									onclick={handleAddNote}
									disabled={!newNote.trim() || isAddingNote}
									class="btn-pill btn-pill-primary btn-pill-sm disabled:opacity-50"
								>
									{isAddingNote ? 'Adding...' : 'Add Note'}
								</button>
							</div>
						</div>

						<!-- Notes List -->
						{#if project.notes.length === 0}
							<div class="text-center py-8 text-gray-400">
								<svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
								</svg>
								<p>No notes yet. Add your first note above.</p>
							</div>
						{:else}
							<div class="space-y-4">
								{#each project.notes as note}
									<div class="p-4 bg-gray-50 rounded-xl">
										<p class="text-gray-700 whitespace-pre-wrap">{note.content}</p>
										<p class="text-xs text-gray-400 mt-3">
											{formatDate(note.created_at)} at {formatTime(note.created_at)}
										</p>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Add Task Dialog -->
<Dialog.Root bind:open={showAddTask}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Add Task</Dialog.Title>

			<form onsubmit={handleCreateTask} class="space-y-4">
				<div>
					<label for="task-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						id="task-title"
						type="text"
						bind:value={newTask.title}
						class="input input-square"
						placeholder="What needs to be done?"
						required
					/>
				</div>

				<div>
					<label for="task-description" class="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label>
					<textarea
						id="task-description"
						bind:value={newTask.description}
						class="input input-square resize-none"
						rows="2"
						placeholder="Add more details..."
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="task-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
						<select id="task-priority" bind:value={newTask.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
					<div>
						<label for="task-due" class="block text-sm font-medium text-gray-700 mb-1">Due Date</label>
						<input
							id="task-due"
							type="date"
							bind:value={newTask.due_date}
							class="input input-square"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="task-estimated" class="block text-sm font-medium text-gray-700 mb-1">Estimated Hours</label>
						<input
							id="task-estimated"
							type="number"
							min="0"
							step="0.5"
							bind:value={newTask.estimated_hours}
							class="input input-square"
							placeholder="0.0"
						/>
					</div>
					<div>
						<label for="task-start" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
						<input
							id="task-start"
							type="date"
							bind:value={newTask.start_date}
							class="input input-square"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="task-parent" class="block text-sm font-medium text-gray-700 mb-1">Parent Task (optional)</label>
						<select id="task-parent" bind:value={newTask.parent_task_id} class="input input-square">
							<option value="">None (Top-level)</option>
							{#each tasks.filter(t => t.status !== 'done') as task}
								<option value={task.id}>{task.title}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="task-assignee" class="block text-sm font-medium text-gray-700 mb-1">Assignee (optional)</label>
						<select id="task-assignee" bind:value={newTask.assignee_id} class="input input-square">
							<option value="">Unassigned</option>
							{#each teamMembers as member}
								<option value={member.id}>{member.name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => showAddTask = false} class="btn-pill btn-pill-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn-pill btn-pill-primary flex-1">
						Add Task
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Edit Task Dialog -->
{#if editingTask}
<Dialog.Root bind:open={showEditTask}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Edit Task</Dialog.Title>

			<form onsubmit={handleUpdateTask} class="space-y-4">
				<div>
					<label for="edit-task-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						id="edit-task-title"
						type="text"
						bind:value={editingTask.title}
						class="input input-square"
						placeholder="Task title..."
						required
					/>
				</div>

				<div>
					<label for="edit-task-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						id="edit-task-description"
						bind:value={editingTask.description}
						class="input input-square resize-none"
						rows="2"
						placeholder="Add more details..."
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-task-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
						<select id="edit-task-priority" bind:value={editingTask.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
					<div>
						<label for="edit-task-due" class="block text-sm font-medium text-gray-700 mb-1">Due Date</label>
						<input
							id="edit-task-due"
							type="date"
							bind:value={editingTask.due_date}
							class="input input-square"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-task-estimated" class="block text-sm font-medium text-gray-700 mb-1">Estimated Hours</label>
						<input
							id="edit-task-estimated"
							type="number"
							min="0"
							step="0.5"
							bind:value={editingTask.estimated_hours}
							class="input input-square"
							placeholder="0.0"
						/>
					</div>
					<div>
						<label for="edit-task-start" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
						<input
							id="edit-task-start"
							type="date"
							bind:value={editingTask.start_date}
							class="input input-square"
						/>
					</div>
				</div>

				<div>
					<label for="edit-task-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
					<select id="edit-task-status" bind:value={editingTask.status} class="input input-square">
						<option value="todo">To Do</option>
						<option value="in_progress">In Progress</option>
						<option value="done">Done</option>
						<option value="cancelled">Cancelled</option>
					</select>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => { showEditTask = false; editingTask = null; }} class="btn-pill btn-pill-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn-pill btn-pill-primary flex-1">
						Save Changes
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
{/if}

<!-- Edit Dialog -->
<Dialog.Root bind:open={showEditDialog}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Edit Project</Dialog.Title>

			<form onsubmit={(e) => { e.preventDefault(); handleSave(); }} class="space-y-4">
				<div>
					<label for="edit-name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
					<input
						id="edit-name"
						type="text"
						bind:value={editForm.name}
						class="input input-square"
						required
					/>
				</div>

				<div>
					<label for="edit-client" class="block text-sm font-medium text-gray-700 mb-1">Client</label>
					<select id="edit-client" bind:value={editForm.client_name} class="input input-square">
						<option value="">No client</option>
						{#each clients as client}
							<option value={client.name}>{client.name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label for="edit-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						id="edit-description"
						bind:value={editForm.description}
						class="input input-square resize-none"
						rows="3"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select id="edit-status" bind:value={editForm.status} class="input input-square">
							<option value="active">Active</option>
							<option value="paused">Paused</option>
							<option value="completed">Completed</option>
							<option value="archived">Archived</option>
						</select>
					</div>
					<div>
						<label for="edit-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
						<select id="edit-priority" bind:value={editForm.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
				</div>

				<div>
					<label for="edit-type" class="block text-sm font-medium text-gray-700 mb-1">Type</label>
					<select id="edit-type" bind:value={editForm.project_type} class="input input-square">
						<option value="internal">Internal</option>
						<option value="client_work">Client Work</option>
						<option value="learning">Learning</option>
					</select>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => showEditDialog = false} class="btn-pill btn-pill-secondary flex-1">
						Cancel
					</button>
					<button type="submit" disabled={isSaving} class="btn-pill btn-pill-primary flex-1">
						{isSaving ? 'Saving...' : 'Save Changes'}
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Delete Confirmation -->
<Dialog.Root bind:open={showDeleteConfirm}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-2">Delete Project</Dialog.Title>
			<p class="text-sm text-gray-500 mb-6">
				Are you sure you want to delete "{project?.name}"? This action cannot be undone.
			</p>
			<div class="flex gap-3">
				<button onclick={() => showDeleteConfirm = false} class="btn-pill btn-pill-secondary flex-1">
					Cancel
				</button>
				<button onclick={handleDelete} class="btn-pill btn-pill-danger flex-1">
					Delete
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Document Editor Panel - Side Mode -->
{#if documentPanelMode === 'side' && selectedDocument}
	<div
		class="fixed inset-y-0 right-0 bg-white border-l border-gray-200 shadow-xl z-40 flex flex-col"
		style="width: {documentPanelWidth}px"
	>
		<!-- Resize Handle -->
		<div
			onmousedown={startPanelResize}
			class="absolute left-0 top-0 bottom-0 w-1 cursor-ew-resize hover:bg-purple-500 transition-colors group"
		>
			<div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-12 bg-gray-300 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"></div>
		</div>

		<!-- Panel Header -->
		<div class="px-4 py-3 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
			<div class="flex items-center gap-3 min-w-0 flex-1">
				<span class="text-xl flex-shrink-0">{selectedDocument.icon || '📄'}</span>
				<input
					type="text"
					bind:value={documentTitle}
					onblur={updateDocumentTitle}
					onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
					class="flex-1 min-w-0 font-medium text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
				/>
			</div>
			<div class="flex items-center gap-1">
				<!-- Save status -->
				<div class="text-xs text-gray-400 mr-2">
					{#if $editor.isDirty}
						<span class="text-amber-500">Unsaved</span>
					{:else if $editor.isSaving}
						<span>Saving...</span>
					{:else if $editor.lastSavedAt}
						<span class="text-green-600">Saved</span>
					{/if}
				</div>

				<!-- Mode switcher -->
				<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
					<button
						onclick={() => documentPanelMode = 'side'}
						class="p-1.5 transition-colors bg-gray-900 text-white"
						title="Side panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'center'}
						class="p-1.5 transition-colors text-gray-500 hover:bg-gray-100"
						title="Center panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'full'}
						class="p-1.5 transition-colors text-gray-500 hover:bg-gray-100"
						title="Full screen"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
						</svg>
					</button>
				</div>

				<!-- Open in full page -->
				<a
					href="/knowledge-v2/{selectedDocument.id}{embedSuffix}"
					class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors ml-1"
					title="Open in full page"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
					</svg>
				</a>

				<!-- Close button -->
				<button
					onclick={closeDocumentPanel}
					class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					title="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Editor Content -->
		{#if loadingDocument}
			<div class="flex-1 flex items-center justify-center">
				<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else}
			<div class="flex-1 overflow-y-auto">
				<div class="max-w-none mx-auto px-6 py-8">
					<!-- Blocks -->
					<div class="blocks-container" role="textbox" tabindex="-1">
						{#each $editor.blocks as block, index (block.id)}
							<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
						{/each}
					</div>

					<!-- Click area to add new blocks -->
					<button
						onclick={addNewBlockAtEnd}
						class="w-full min-h-24 mt-4 text-left cursor-text group"
					>
						<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
							Click to add a block, or press / for commands
						</span>
					</button>
				</div>
			</div>

			<!-- Status Bar -->
			<div class="px-4 py-2 border-t border-gray-100 flex items-center justify-between text-xs text-gray-400 bg-gray-50/50">
				<div class="flex items-center gap-4">
					<span>{$wordCount} words</span>
					<span>{$editor.blocks.length} blocks</span>
				</div>
				<button onclick={saveDocument} class="hover:text-gray-600" disabled={!$editor.isDirty}>
					Save now
				</button>
			</div>
		{/if}

		<!-- Slash Command Menu -->
		{#if $editor.showSlashMenu && $editor.slashMenuPosition}
			<BlockMenu />
		{/if}
	</div>
{/if}

<!-- Document Editor Panel - Center Mode -->
{#if documentPanelMode === 'center' && selectedDocument}
	<div class="fixed inset-0 bg-black/30 z-40 flex items-center justify-center p-8" onclick={(e) => { if (e.target === e.currentTarget) closeDocumentPanel(); }}>
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-4xl h-full max-h-[90vh] flex flex-col overflow-hidden">
			<!-- Panel Header -->
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<div class="flex items-center gap-3 min-w-0 flex-1">
					<span class="text-2xl flex-shrink-0">{selectedDocument.icon || '📄'}</span>
					<input
						type="text"
						bind:value={documentTitle}
						onblur={updateDocumentTitle}
						onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
						class="flex-1 min-w-0 text-lg font-semibold text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
					/>
				</div>
				<div class="flex items-center gap-2">
					<!-- Save status -->
					<div class="text-sm text-gray-400 mr-2">
						{#if $editor.isDirty}
							<span class="text-amber-500">Unsaved</span>
						{:else if $editor.isSaving}
							<span>Saving...</span>
						{:else if $editor.lastSavedAt}
							<span class="text-green-600">Saved</span>
						{/if}
					</div>

					<!-- Mode switcher -->
					<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
						<button
							onclick={() => documentPanelMode = 'side'}
							class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
							title="Side panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'center'}
							class="p-2 transition-colors bg-gray-900 text-white"
							title="Center panel"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
							</svg>
						</button>
						<button
							onclick={() => documentPanelMode = 'full'}
							class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
							title="Full screen"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
							</svg>
						</button>
					</div>

					<a
						href="/knowledge-v2/{selectedDocument.id}{embedSuffix}"
						class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						title="Open in full page"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>

					<button
						onclick={closeDocumentPanel}
						class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						title="Close"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Editor Content -->
			{#if loadingDocument}
				<div class="flex-1 flex items-center justify-center">
					<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
				</div>
			{:else}
				<div class="flex-1 overflow-y-auto">
					<div class="max-w-3xl mx-auto px-8 py-12">
						<!-- Blocks -->
						<div class="blocks-container" role="textbox" tabindex="-1">
							{#each $editor.blocks as block, index (block.id)}
								<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
							{/each}
						</div>

						<!-- Click area to add new blocks -->
						<button
							onclick={addNewBlockAtEnd}
							class="w-full min-h-24 mt-4 text-left cursor-text group"
						>
							<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
								Click to add a block, or press / for commands
							</span>
						</button>
					</div>
				</div>

				<!-- Status Bar -->
				<div class="px-6 py-3 border-t border-gray-100 flex items-center justify-between text-sm text-gray-400">
					<div class="flex items-center gap-4">
						<span>{$wordCount} words</span>
						<span>{$editor.blocks.length} blocks</span>
					</div>
					<button onclick={saveDocument} class="hover:text-gray-600" disabled={!$editor.isDirty}>
						Save now
					</button>
				</div>
			{/if}

			<!-- Slash Command Menu -->
			{#if $editor.showSlashMenu && $editor.slashMenuPosition}
				<BlockMenu />
			{/if}
		</div>
	</div>
{/if}

<!-- Document Editor Panel - Full Screen Mode -->
{#if documentPanelMode === 'full' && selectedDocument}
	<div class="fixed inset-0 bg-white z-50 flex flex-col">
		<!-- Panel Header -->
		<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between bg-white">
			<div class="flex items-center gap-3">
				<button
					onclick={closeDocumentPanel}
					class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					title="Back to project"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<span class="text-gray-300">|</span>
				<span class="text-2xl">{selectedDocument.icon || '📄'}</span>
				<input
					type="text"
					bind:value={documentTitle}
					onblur={updateDocumentTitle}
					onkeydown={(e) => e.key === 'Enter' && updateDocumentTitle()}
					class="text-xl font-semibold text-gray-900 bg-transparent border-none outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 rounded px-1"
				/>
			</div>
			<div class="flex items-center gap-2">
				<!-- Save status -->
				<div class="text-sm text-gray-400 mr-4">
					{#if $editor.isDirty}
						<span class="text-amber-500">Unsaved changes</span>
					{:else if $editor.isSaving}
						<span>Saving...</span>
					{:else if $editor.lastSavedAt}
						<span class="text-green-600">All changes saved</span>
					{/if}
				</div>

				<!-- Mode switcher -->
				<div class="flex items-center border border-gray-200 rounded-lg overflow-hidden">
					<button
						onclick={() => documentPanelMode = 'side'}
						class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
						title="Side panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'center'}
						class="p-2 transition-colors text-gray-500 hover:bg-gray-100"
						title="Center panel"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
						</svg>
					</button>
					<button
						onclick={() => documentPanelMode = 'full'}
						class="p-2 transition-colors bg-gray-900 text-white"
						title="Full screen"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
						</svg>
					</button>
				</div>

				<a
					href="/knowledge-v2/{selectedDocument.id}{embedSuffix}"
					class="btn-pill btn-pill-secondary btn-pill-sm ml-2"
					title="Open in Knowledge Base"
				>
					Open in Knowledge Base
				</a>

				<button
					onclick={closeDocumentPanel}
					class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors ml-2"
					title="Exit full screen"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Editor Content -->
		{#if loadingDocument}
			<div class="flex-1 flex items-center justify-center">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else}
			<div class="flex-1 overflow-y-auto bg-gray-50/50">
				<div class="max-w-3xl mx-auto px-8 py-12 bg-white min-h-full shadow-sm">
					<!-- Blocks -->
					<div class="blocks-container" role="textbox" tabindex="-1">
						{#each $editor.blocks as block, index (block.id)}
							<BlockComponent {block} {index} readonly={false} parentContextId={selectedDocument.id} />
						{/each}
					</div>

					<!-- Click area to add new blocks -->
					<button
						onclick={addNewBlockAtEnd}
						class="w-full min-h-32 mt-4 text-left cursor-text group"
					>
						<span class="text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
							Click to add a block, or press / for commands
						</span>
					</button>
				</div>
			</div>

			<!-- Status Bar -->
			<div class="px-6 py-3 border-t border-gray-200 flex items-center justify-between text-sm text-gray-500 bg-white">
				<div class="flex items-center gap-6">
					<span>{$wordCount} words</span>
					<span>{$editor.blocks.length} blocks</span>
				</div>
				<div class="flex items-center gap-4">
					<button onclick={saveDocument} class="text-purple-600 hover:text-purple-700 font-medium" disabled={!$editor.isDirty}>
						Save now
					</button>
				</div>
			</div>
		{/if}

		<!-- Slash Command Menu -->
		{#if $editor.showSlashMenu && $editor.slashMenuPosition}
			<BlockMenu />
		{/if}
	</div>
{/if}
