<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		ArrowLeft,
		FileCode,
		FileText,
		Database,
		Settings,
		Rocket,
		Download,
		Copy,
		Check,
		Loader2,
		AlertCircle,
		Clock,
		CheckCircle,
		XCircle,
		Package,
		ExternalLink
	} from 'lucide-svelte';
	import { marked } from 'marked';
	import { installModule } from '$lib/api/osa/files';
	import { request } from '$lib/api/base';

	// Get workflow ID from URL params
	const workflowId = $derived($page.params.id);

	// State
	let workflow = $state<any>(null);
	let files = $state<any[]>([]);
	let selectedFile = $state<any>(null);
	let fileContent = $state<string | null>(null);
	let renderedMarkdown = $state<string>('');
	let activeTab = $state<string>('all');
	let loading = $state(true);
	let loadingContent = $state(false);
	let error = $state<string | null>(null);
	let copied = $state(false);
	let installing = $state(false);
	let installSuccess = $state(false);
	let installError = $state<string | null>(null);
	let deploying = $state(false);
	let deploymentUrl = $state('');
	let deployError = $state('');

	// File type categories with icons
	const fileCategories = {
		all: { label: 'All Files', icon: FileCode },
		code: { label: 'Code', icon: FileCode },
		schema: { label: 'Schema', icon: Database },
		config: { label: 'Config', icon: Settings },
		documentation: { label: 'Docs', icon: FileText },
		deployment: { label: 'Deploy', icon: Rocket }
	};

	// Load workflow and files on mount
	onMount(() => {
		loadWorkflow();
	});

	async function loadWorkflow() {
		loading = true;
		error = null;

		try {
			// Fetch workflow details
			const workflowData = await request<any>(`/osa/workflows/${workflowId}`);
			workflow = workflowData;

			// Fetch workflow files
			const filesData = await request<{ files: any[] }>(`/osa/workflows/${workflowId}/files`);
			files = filesData.files || [];

			// Select first file by default
			if (files.length > 0) {
				selectFile(files[0]);
			}
		} catch (err: any) {
			console.error('Failed to load workflow:', err);
			error = err?.message || 'Failed to load workflow';
		} finally {
			loading = false;
		}
	}

	async function selectFile(file: any) {
		selectedFile = file;
		loadingContent = true;
		fileContent = null;
		renderedMarkdown = '';

		try {
			const response = await request<{ content: string; file: any }>(
				`/osa/files/${file.id}/content`
			);
			fileContent = response.content;

			// Render markdown if applicable
			if (file.type === 'markdown' || file.name.endsWith('.md')) {
				renderedMarkdown = await marked.parse(fileContent);
			}
		} catch (err: any) {
			console.error('Failed to load file content:', err);
			fileContent = `Error loading file: ${err?.message || 'Unknown error'}`;
		} finally {
			loadingContent = false;
		}
	}

	async function handleCopy() {
		if (!fileContent) return;

		try {
			await navigator.clipboard.writeText(fileContent);
			copied = true;
			setTimeout(() => {
				copied = false;
			}, 2000);
		} catch (err) {
			console.error('Failed to copy:', err);
		}
	}

	async function handleDownload() {
		if (!selectedFile || !fileContent) return;

		try {
			const blob = new Blob([fileContent], { type: 'text/plain' });
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = selectedFile.name;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err) {
			console.error('Failed to download file:', err);
		}
	}

	async function handleInstallModule() {
		if (!workflow) return;

		installing = true;
		installError = null;
		installSuccess = false;

		try {
			const result = await installModule(workflow.id, {
				module_name: workflow.name,
				install_path: undefined, // Let backend decide
				file_ids: files.map((f) => f.id)
			});

			installSuccess = true;
			setTimeout(() => {
				installSuccess = false;
			}, 3000);
		} catch (err: any) {
			console.error('Failed to install module:', err);
			installError = err?.message || 'Failed to install module';
		} finally {
			installing = false;
		}
	}

	async function handleDeploy() {
		deploying = true;
		deployError = '';
		deploymentUrl = '';

		try {
			const response = await fetch(`/api/osa/apps/${workflow.id}/deploy`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
			});

			if (!response.ok) {
				const error = await response.json();
				throw new Error(error.details || 'Failed to deploy app');
			}

			const result = await response.json();
			deploymentUrl = result.url;
		} catch (err) {
			deployError = err instanceof Error ? err.message : 'Failed to deploy';
		} finally {
			deploying = false;
		}
	}

	function getFilesByCategory(category: string) {
		if (category === 'all') return files;
		return files.filter((f) => f.type === category);
	}

	function getLanguageFromFile(file: any): string {
		if (file.language) return file.language;

		const ext = file.name.split('.').pop()?.toLowerCase() || '';
		const langMap: Record<string, string> = {
			js: 'javascript',
			ts: 'typescript',
			tsx: 'typescript',
			jsx: 'javascript',
			py: 'python',
			go: 'go',
			rs: 'rust',
			java: 'java',
			cpp: 'cpp',
			c: 'c',
			cs: 'csharp',
			rb: 'ruby',
			php: 'php',
			sh: 'bash',
			yaml: 'yaml',
			yml: 'yaml',
			json: 'json',
			xml: 'xml',
			html: 'html',
			css: 'css',
			scss: 'scss',
			md: 'markdown',
			sql: 'sql',
			dockerfile: 'dockerfile'
		};

		return langMap[ext] || 'plaintext';
	}

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return Math.round((bytes / Math.pow(k, i)) * 10) / 10 + ' ' + sizes[i];
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'completed':
				return CheckCircle;
			case 'failed':
				return XCircle;
			case 'processing':
				return Clock;
			default:
				return Clock;
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'completed':
				return 'text-green-400';
			case 'failed':
				return 'text-red-400';
			case 'processing':
				return 'text-yellow-400';
			default:
				return 'text-gray-400';
		}
	}

	const filteredFiles = $derived(getFilesByCategory(activeTab));
</script>

<div class="workflow-viewer">
	{#if loading}
		<div class="loading-state">
			<Loader2 size={48} class="animate-spin text-blue-400" />
			<p>Loading workflow...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<AlertCircle size={48} class="text-red-400" />
			<p class="error-text">{error}</p>
			<button class="back-btn" onclick={() => goto('/window')}>
				<ArrowLeft size={16} />
				<span>Back to Desktop</span>
			</button>
		</div>
	{:else if workflow}
		<!-- Header -->
		<div class="workflow-header">
			<button class="back-btn" onclick={() => goto('/window')}>
				<ArrowLeft size={16} />
				<span>Back</span>
			</button>

			<div class="workflow-info">
				<div class="workflow-title-row">
					<h1 class="workflow-title">{workflow.display_name || workflow.name}</h1>
					<div class="status-badge" class:status-completed={workflow.status === 'completed'}>
						<svelte:component this={getStatusIcon(workflow.status)} size={16} class={getStatusColor(workflow.status)} />
						<span>{workflow.status}</span>
					</div>
				</div>

				<p class="workflow-description">{workflow.description}</p>

				<div class="workflow-meta">
					<span>{files.length} files</span>
					<span>•</span>
					<span>Created {formatDate(workflow.created_at)}</span>
					{#if workflow.generated_at}
						<span>•</span>
						<span>Generated {formatDate(workflow.generated_at)}</span>
					{/if}
				</div>
			</div>

			<button class="install-btn" onclick={handleInstallModule} disabled={installing || installSuccess}>
				{#if installing}
					<Loader2 size={16} class="animate-spin" />
					<span>Installing...</span>
				{:else if installSuccess}
					<CheckCircle size={16} />
					<span>Installed!</span>
				{:else}
					<Package size={16} />
					<span>Install Module</span>
				{/if}
			</button>

			<button class="deploy-btn" onclick={handleDeploy} disabled={deploying}>
				{#if deploying}
					<Loader2 size={16} class="animate-spin" />
					<span>Deploying...</span>
				{:else if deploymentUrl}
					<ExternalLink size={16} />
					<a href={deploymentUrl} target="_blank">Running</a>
				{:else}
					<Rocket size={16} />
					<span>Deploy & Run</span>
				{/if}
			</button>
		</div>

		{#if installError}
			<div class="install-error">
				<AlertCircle size={16} />
				<span>{installError}</span>
			</div>
		{/if}

		{#if deployError}
			<div class="deploy-error">
				<AlertCircle size={16} />
				<span>{deployError}</span>
			</div>
		{/if}

		<!-- Main Content -->
		<div class="workflow-content">
			<!-- Sidebar: File List -->
			<div class="file-sidebar">
				<!-- File Type Tabs -->
				<div class="file-tabs">
					{#each Object.entries(fileCategories) as [key, category]}
						{@const count = getFilesByCategory(key).length}
						{#if count > 0 || key === 'all'}
							<button
								class="file-tab"
								class:active={activeTab === key}
								onclick={() => (activeTab = key)}
							>
								<svelte:component this={category.icon} size={16} />
								<span>{category.label}</span>
								<span class="file-count">{count}</span>
							</button>
						{/if}
					{/each}
				</div>

				<!-- File List -->
				<div class="file-list">
					{#each filteredFiles as file}
						<button
							class="file-item"
							class:selected={selectedFile?.id === file.id}
							onclick={() => selectFile(file)}
						>
							<svelte:component this={fileCategories[file.type as keyof typeof fileCategories]?.icon || FileCode} size={16} class="file-icon" />
							<div class="file-item-info">
								<span class="file-name">{file.name}</span>
								<span class="file-size">{formatFileSize(file.size)}</span>
							</div>
						</button>
					{/each}

					{#if filteredFiles.length === 0}
						<div class="empty-files">
							<FileText size={32} class="text-gray-500" />
							<p>No files in this category</p>
						</div>
					{/if}
				</div>
			</div>

			<!-- Main: File Preview -->
			<div class="file-preview">
				{#if !selectedFile}
					<div class="empty-preview">
						<FileCode size={48} class="text-gray-400" />
						<p>Select a file to preview</p>
					</div>
				{:else}
					<!-- File Header -->
					<div class="preview-header">
						<div class="file-info">
							<svelte:component this={fileCategories[selectedFile.type as keyof typeof fileCategories]?.icon || FileCode} size={20} class="text-blue-400" />
							<div class="file-details">
								<h3 class="file-header-name">{selectedFile.name}</h3>
								<p class="file-header-meta">
									{formatFileSize(selectedFile.size)} • {selectedFile.type} • Updated {formatDate(
										selectedFile.updated_at
									)}
								</p>
							</div>
						</div>

						<div class="preview-actions">
							<button
								class="action-btn"
								onclick={handleCopy}
								disabled={!fileContent || loadingContent}
							>
								{#if copied}
									<Check size={16} class="text-green-400" />
								{:else}
									<Copy size={16} />
								{/if}
								<span>{copied ? 'Copied!' : 'Copy'}</span>
							</button>

							<button class="action-btn" onclick={handleDownload} disabled={loadingContent}>
								<Download size={16} />
								<span>Download</span>
							</button>
						</div>
					</div>

					<!-- File Content -->
					<div class="preview-content">
						{#if loadingContent}
							<div class="loading-content">
								<Loader2 size={32} class="animate-spin text-blue-400" />
								<p>Loading file content...</p>
							</div>
						{:else if fileContent}
							{#if renderedMarkdown}
								<!-- Markdown Preview -->
								<div class="markdown-preview">
									{@html renderedMarkdown}
								</div>
							{:else}
								<!-- Code Preview with Syntax Highlighting -->
								<pre class="code-preview"><code class="language-{getLanguageFromFile(
									selectedFile
								)}">{fileContent}</code></pre>
							{/if}
						{:else}
							<div class="empty-content">
								<FileText size={32} class="text-gray-400" />
								<p>No content available</p>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.workflow-viewer {
		display: flex;
		flex-direction: column;
		height: 100vh;
		background: #0f172a;
		color: #e2e8f0;
		overflow: hidden;
	}

	/* Loading & Error States */
	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: #94a3b8;
	}

	.error-text {
		color: #ef4444;
		font-size: 16px;
	}

	/* Header */
	.workflow-header {
		display: flex;
		align-items: flex-start;
		gap: 20px;
		padding: 24px;
		background: #1e293b;
		border-bottom: 1px solid #334155;
	}

	.back-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		background: #0f172a;
		border: 1px solid #334155;
		border-radius: 6px;
		color: #e2e8f0;
		font-size: 14px;
		cursor: pointer;
		transition: all 0.15s ease;
		white-space: nowrap;
	}

	.back-btn:hover {
		background: #1e293b;
		border-color: #60a5fa;
	}

	.workflow-info {
		flex: 1;
		min-width: 0;
	}

	.workflow-title-row {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 8px;
	}

	.workflow-title {
		font-size: 24px;
		font-weight: 600;
		color: #f1f5f9;
		margin: 0;
	}

	.status-badge {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 4px 12px;
		background: #1e293b;
		border: 1px solid #334155;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		text-transform: capitalize;
	}

	.status-badge.status-completed {
		background: #064e3b;
		border-color: #059669;
	}

	.workflow-description {
		color: #94a3b8;
		font-size: 14px;
		margin: 0 0 12px 0;
		line-height: 1.5;
	}

	.workflow-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		color: #64748b;
	}

	.install-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 20px;
		background: #3b82f6;
		border: none;
		border-radius: 6px;
		color: white;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
		white-space: nowrap;
	}

	.install-btn:hover:not(:disabled) {
		background: #2563eb;
	}

	.install-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.install-error {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 24px;
		background: #7f1d1d;
		border-bottom: 1px solid #991b1b;
		color: #fca5a5;
		font-size: 14px;
	}

	.deploy-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 20px;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		white-space: nowrap;
	}

	.deploy-btn:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	.deploy-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.deploy-btn a {
		color: white;
		text-decoration: none;
	}

	.deploy-error {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 24px;
		background: #7f1d1d;
		border-bottom: 1px solid #991b1b;
		color: #fca5a5;
		font-size: 14px;
	}

	/* Main Content */
	.workflow-content {
		display: flex;
		flex: 1;
		overflow: hidden;
	}

	/* File Sidebar */
	.file-sidebar {
		width: 320px;
		display: flex;
		flex-direction: column;
		background: #1e293b;
		border-right: 1px solid #334155;
	}

	.file-tabs {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 16px;
		border-bottom: 1px solid #334155;
	}

	.file-tab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		background: transparent;
		border: 1px solid transparent;
		border-radius: 6px;
		color: #94a3b8;
		font-size: 14px;
		cursor: pointer;
		transition: all 0.15s ease;
		text-align: left;
	}

	.file-tab:hover {
		background: #0f172a;
		color: #e2e8f0;
	}

	.file-tab.active {
		background: #0f172a;
		border-color: #3b82f6;
		color: #60a5fa;
	}

	.file-count {
		margin-left: auto;
		padding: 2px 8px;
		background: #334155;
		border-radius: 10px;
		font-size: 12px;
	}

	.file-tab.active .file-count {
		background: #1e3a8a;
		color: #93c5fd;
	}

	.file-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
	}

	.file-item {
		display: flex;
		align-items: center;
		gap: 12px;
		width: 100%;
		padding: 12px;
		background: transparent;
		border: 1px solid transparent;
		border-radius: 6px;
		color: #e2e8f0;
		font-size: 14px;
		cursor: pointer;
		transition: all 0.15s ease;
		text-align: left;
		margin-bottom: 4px;
	}

	.file-item:hover {
		background: #0f172a;
		border-color: #334155;
	}

	.file-item.selected {
		background: #1e3a8a;
		border-color: #3b82f6;
	}

	.file-icon {
		flex-shrink: 0;
		color: #60a5fa;
	}

	.file-item-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.file-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-weight: 500;
	}

	.file-size {
		font-size: 12px;
		color: #64748b;
	}

	.empty-files {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px 20px;
		gap: 12px;
		color: #64748b;
	}

	/* File Preview */
	.file-preview {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: #0f172a;
		overflow: hidden;
	}

	.empty-preview {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: #64748b;
	}

	.preview-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 24px;
		background: #1e293b;
		border-bottom: 1px solid #334155;
	}

	.file-info {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1;
		min-width: 0;
	}

	.file-details {
		flex: 1;
		min-width: 0;
	}

	.file-header-name {
		font-size: 16px;
		font-weight: 600;
		color: #f1f5f9;
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.file-header-meta {
		font-size: 12px;
		color: #64748b;
		margin: 4px 0 0 0;
	}

	.preview-actions {
		display: flex;
		gap: 8px;
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		background: #0f172a;
		border: 1px solid #334155;
		border-radius: 6px;
		color: #e2e8f0;
		font-size: 14px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.action-btn:hover:not(:disabled) {
		background: #1e293b;
		border-color: #60a5fa;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.preview-content {
		flex: 1;
		overflow: auto;
		padding: 24px;
	}

	.loading-content,
	.empty-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: #64748b;
	}

	.code-preview {
		margin: 0;
		padding: 20px;
		background: #020617;
		border: 1px solid #1e293b;
		border-radius: 8px;
		overflow-x: auto;
		font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
		font-size: 13px;
		line-height: 1.7;
		color: #e2e8f0;
	}

	.code-preview code {
		display: block;
	}

	/* Markdown Styling */
	.markdown-preview {
		color: #e2e8f0;
		line-height: 1.7;
	}

	.markdown-preview :global(h1),
	.markdown-preview :global(h2),
	.markdown-preview :global(h3),
	.markdown-preview :global(h4),
	.markdown-preview :global(h5),
	.markdown-preview :global(h6) {
		color: #f1f5f9;
		margin-top: 24px;
		margin-bottom: 16px;
		font-weight: 600;
	}

	.markdown-preview :global(h1) {
		font-size: 28px;
		border-bottom: 1px solid #334155;
		padding-bottom: 8px;
	}

	.markdown-preview :global(h2) {
		font-size: 24px;
		border-bottom: 1px solid #1e293b;
		padding-bottom: 6px;
	}

	.markdown-preview :global(h3) {
		font-size: 20px;
	}

	.markdown-preview :global(code) {
		background: #020617;
		padding: 2px 8px;
		border-radius: 4px;
		font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
		font-size: 13px;
		border: 1px solid #1e293b;
	}

	.markdown-preview :global(pre) {
		background: #020617;
		padding: 20px;
		border: 1px solid #1e293b;
		border-radius: 8px;
		overflow-x: auto;
		margin: 16px 0;
	}

	.markdown-preview :global(pre code) {
		background: transparent;
		padding: 0;
		border: none;
	}

	.markdown-preview :global(a) {
		color: #60a5fa;
		text-decoration: none;
	}

	.markdown-preview :global(a:hover) {
		text-decoration: underline;
	}

	.markdown-preview :global(ul),
	.markdown-preview :global(ol) {
		padding-left: 24px;
		margin: 16px 0;
	}

	.markdown-preview :global(li) {
		margin: 8px 0;
	}

	.markdown-preview :global(blockquote) {
		border-left: 4px solid #334155;
		padding-left: 16px;
		margin: 16px 0;
		color: #94a3b8;
	}

	.markdown-preview :global(table) {
		border-collapse: collapse;
		width: 100%;
		margin: 16px 0;
	}

	.markdown-preview :global(th),
	.markdown-preview :global(td) {
		border: 1px solid #334155;
		padding: 8px 12px;
		text-align: left;
	}

	.markdown-preview :global(th) {
		background: #1e293b;
		font-weight: 600;
	}

	.markdown-preview :global(img) {
		max-width: 100%;
		border-radius: 8px;
		margin: 16px 0;
	}
</style>
