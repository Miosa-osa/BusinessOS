<script lang="ts">
	import DocumentUploadModal from '$lib/components/chat/modals/DocumentUploadModal.svelte';
	import HybridSearchPanel from '$lib/components/chat/panels/HybridSearchPanel.svelte';
	import SaveToProfileModal from './SaveToProfileModal.svelte';
	import TaskGenerationModal from './TaskGenerationModal.svelte';
	import CreateProjectModal from './CreateProjectModal.svelte';

	import { chatUIStore } from '$lib/stores/chat/chatUIStore.svelte';
	import { chatContextStore } from '$lib/stores/chat/chatContextStore.svelte';
	import { chatArtifactStore } from '$lib/stores/chat/chatArtifactStore.svelte';
	import { chatConversationStore } from '$lib/stores/chat/chatConversationStore.svelte';
	import { currentWorkspaceId } from '$lib/stores/workspaces';

	const ui = chatUIStore;
	const cx = chatContextStore;
	const ar = chatArtifactStore;
	const cs = chatConversationStore;
</script>

<SaveToProfileModal
	show={ar.showSaveToProfileModal}
	selectedProfileForSave={ar.selectedProfileForSave}
	availableProfiles={ar.availableProfiles}
	savingArtifactToProfile={ar.savingArtifactToProfile}
	onClose={() => { ar.showSaveToProfileModal = false; }}
	onSelectProfile={(id) => { ar.selectedProfileForSave = id; }}
	onSave={ar.saveArtifactToProfile}
/>

<TaskGenerationModal
	show={ar.showTaskGenerationModal}
	generatingTasks={ar.generatingTasks}
	generatedTasks={ar.generatedTasks}
	selectedProjectForTasks={ar.selectedProjectForTasks}
	availableProjects={ar.availableProjects}
	availableTeamMembers={cx.availableTeamMembers}
	taskGenerationArtifact={ar.taskGenerationArtifact}
	onClose={() => { ar.showTaskGenerationModal = false; ar.generatedTasks = []; }}
	onSelectProject={(id) => { ar.selectedProjectForTasks = id; }}
	onRemoveTask={ar.removeGeneratedTask}
	onUpdateTaskAssignee={ar.updateTaskAssignee}
	onConfirm={ar.confirmTaskCreation}
/>

<CreateProjectModal
	show={cx.showNewProjectModal}
	bind:newProjectName={cx.newProjectName}
	creatingProject={cx.creatingProject}
	onClose={() => { cx.showNewProjectModal = false; cx.newProjectName = ''; }}
	onNameChange={(name) => { cx.newProjectName = name; }}
	onCreate={cx.createProjectQuick}
/>

<DocumentUploadModal
	bind:open={cx.showDocumentUploadModal}
	onClose={() => { cx.showDocumentUploadModal = false; }}
	onUploadComplete={(doc) => {
		if (doc.id) {
			ui.activeResources = [...ui.activeResources, {
				id: doc.id,
				type: 'document',
				title: doc.display_name || doc.original_filename,
				contextId: doc.id,
				tokenCount: doc.word_count ? doc.word_count * 2 : undefined
			}];
		}
	}}
/>

<HybridSearchPanel
	bind:show={cx.showHybridSearchPanel}
	workspaceId={$currentWorkspaceId ?? undefined}
	onaddToContext={({ result, query }) => {
		ui.activeResources = [...ui.activeResources, {
			id: result.context_id,
			type: 'search_result',
			title: result.context_name,
			contextId: result.context_id,
			tokenCount: Math.ceil(result.content.length / 4)
		}];

		if (!cx.selectedContextIds.includes(result.context_id)) {
			cx.selectedContextIds = [...cx.selectedContextIds, result.context_id];
		}

		if (cs.inputValue.trim() === '') {
			cs.inputValue = `Using context from "${result.context_name}": ${query}`;
		}
	}}
/>
