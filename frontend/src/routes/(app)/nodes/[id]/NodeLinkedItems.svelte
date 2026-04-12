<script lang="ts">
	import { slide } from 'svelte/transition';

	interface LinkedProject {
		id: string;
		name: string;
		status: string;
	}

	interface LinkedContext {
		id: string;
		name: string;
		type: string;
	}

	interface LinkedConversation {
		id: string;
		title?: string | null;
		updated_at: string;
	}

	interface Props {
		linkedProjects: LinkedProject[];
		linkedContexts: LinkedContext[];
		linkedConversations: LinkedConversation[];
		onManageLinks: () => void;
	}

	let { linkedProjects, linkedContexts, linkedConversations, onManageLinks }: Props = $props();

	let expandedLinkedSection: 'projects' | 'contexts' | 'conversations' | null = $state(null);
</script>

<div class="ng-section">
	<div class="ng-section__header">
		<h2 class="ng-section__title">Linked Items</h2>
		<button
			onclick={onManageLinks}
			class="btn-pill btn-pill-ghost btn-pill-xs ng-manage-btn"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
			</svg>
			Manage Links
		</button>
	</div>

	<div class="ng-accordion-list">
		<!-- Projects Section -->
		<div class="ng-accordion">
			<button
				onclick={() => expandedLinkedSection = expandedLinkedSection === 'projects' ? null : 'projects'}
				class="btn-pill btn-pill-ghost ng-accordion__trigger"
			>
				<span class="ng-accordion__left">
					<div class="ng-link-icon ng-link-icon--project">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
						</svg>
					</div>
					<span class="ng-accordion__label">Projects</span>
				</span>
				<div class="ng-accordion__right">
					<span class="ng-count-badge">{linkedProjects.length}</span>
					<svg class="w-4 h-4 ng-accordion__chevron {expandedLinkedSection === 'projects' ? 'ng-accordion__chevron--open' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</div>
			</button>
			{#if expandedLinkedSection === 'projects'}
				<div class="ng-accordion__content" transition:slide={{ duration: 200 }}>
					{#if linkedProjects.length === 0}
						<p class="ng-empty-text ng-empty-text--center">No projects linked</p>
					{:else}
						<div class="ng-link-list">
							{#each linkedProjects as project}
								<a
									href="/projects/{project.id}"
									class="btn-pill btn-pill-ghost ng-link-item"
								>
									<div class="ng-link-icon ng-link-icon--project ng-link-icon--sm">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
									</div>
									<div class="ng-link-item__body">
										<p class="ng-link-item__name">{project.name}</p>
										<p class="ng-link-item__meta">{project.status}</p>
									</div>
									<svg class="w-4 h-4 ng-link-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Contexts Section -->
		<div class="ng-accordion">
			<button
				onclick={() => expandedLinkedSection = expandedLinkedSection === 'contexts' ? null : 'contexts'}
				class="btn-pill btn-pill-ghost ng-accordion__trigger"
			>
				<span class="ng-accordion__left">
					<div class="ng-link-icon ng-link-icon--context">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
						</svg>
					</div>
					<span class="ng-accordion__label">Context Profiles</span>
				</span>
				<div class="ng-accordion__right">
					<span class="ng-count-badge">{linkedContexts.length}</span>
					<svg class="w-4 h-4 ng-accordion__chevron {expandedLinkedSection === 'contexts' ? 'ng-accordion__chevron--open' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</div>
			</button>
			{#if expandedLinkedSection === 'contexts'}
				<div class="ng-accordion__content" transition:slide={{ duration: 200 }}>
					{#if linkedContexts.length === 0}
						<p class="ng-empty-text ng-empty-text--center">No context profiles linked</p>
					{:else}
						<div class="ng-link-list">
							{#each linkedContexts as context}
								<a
									href="/knowledge/{context.id}"
									class="btn-pill btn-pill-ghost ng-link-item"
								>
									<div class="ng-link-icon ng-link-icon--context ng-link-icon--sm">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
										</svg>
									</div>
									<div class="ng-link-item__body">
										<p class="ng-link-item__name">{context.name}</p>
										<p class="ng-link-item__meta">{context.type}</p>
									</div>
									<svg class="w-4 h-4 ng-link-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Conversations Section -->
		<div class="ng-accordion">
			<button
				onclick={() => expandedLinkedSection = expandedLinkedSection === 'conversations' ? null : 'conversations'}
				class="btn-pill btn-pill-ghost ng-accordion__trigger"
			>
				<span class="ng-accordion__left">
					<div class="ng-link-icon ng-link-icon--conversation">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</div>
					<span class="ng-accordion__label">Conversations</span>
				</span>
				<div class="ng-accordion__right">
					<span class="ng-count-badge">{linkedConversations.length}</span>
					<svg class="w-4 h-4 ng-accordion__chevron {expandedLinkedSection === 'conversations' ? 'ng-accordion__chevron--open' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</div>
			</button>
			{#if expandedLinkedSection === 'conversations'}
				<div class="ng-accordion__content" transition:slide={{ duration: 200 }}>
					{#if linkedConversations.length === 0}
						<p class="ng-empty-text ng-empty-text--center">No conversations linked</p>
					{:else}
						<div class="ng-link-list">
							{#each linkedConversations as conversation}
								<a
									href="/chat/{conversation.id}"
									class="btn-pill btn-pill-ghost ng-link-item"
								>
									<div class="ng-link-icon ng-link-icon--conversation ng-link-icon--sm">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
										</svg>
									</div>
									<div class="ng-link-item__body">
										<p class="ng-link-item__name">{conversation.title || 'Untitled Conversation'}</p>
										<p class="ng-link-item__meta">{new Date(conversation.updated_at).toLocaleDateString()}</p>
									</div>
									<svg class="w-4 h-4 ng-link-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.ng-section {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 1.25rem;
	}
	.ng-section__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	.ng-section__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
	}
	.ng-manage-btn { display: flex; align-items: center; gap: 0.25rem; }
	.ng-accordion-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.ng-accordion {
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
		overflow: hidden;
	}
	.ng-accordion__trigger {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem;
	}
	.ng-accordion__left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		color: var(--dt2);
	}
	.ng-accordion__label { font-weight: 500; }
	.ng-accordion__right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.ng-accordion__chevron {
		color: var(--dt4);
		transition: transform 0.2s;
	}
	.ng-accordion__chevron--open { transform: rotate(180deg); }
	.ng-link-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.ng-link-icon--sm { width: 1.5rem; height: 1.5rem; border-radius: 0.25rem; flex-shrink: 0; }
	.ng-link-icon--project { background: rgba(34, 197, 94, 0.1); color: #16a34a; }
	.ng-link-icon--context { background: rgba(168, 85, 247, 0.1); color: #9333ea; }
	.ng-link-icon--conversation { background: rgba(59, 130, 246, 0.1); color: #2563eb; }
	.ng-count-badge {
		padding: 0.125rem 0.5rem;
		font-size: 0.75rem;
		font-weight: 500;
		background: var(--dbg2);
		color: var(--dt2);
		border-radius: 9999px;
	}
	.ng-accordion__content {
		padding: 0.75rem;
		border-top: 1px solid var(--dbd);
	}
	.ng-link-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.ng-link-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ng-link-item__body { flex: 1; min-width: 0; }
	.ng-link-item__name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.ng-link-item__meta {
		font-size: 0.75rem;
		color: var(--dt3);
		text-transform: capitalize;
	}
	.ng-link-chevron {
		color: var(--dt4);
		opacity: 0;
		transition: opacity 0.15s;
	}
	.ng-link-item:hover .ng-link-chevron { opacity: 1; }
	.ng-empty-text { color: var(--dt4); font-size: 0.875rem; }
	.ng-empty-text--center { text-align: center; padding: 0.5rem 0; }
</style>
