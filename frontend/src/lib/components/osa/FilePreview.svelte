<script lang="ts">
	import { onMount } from 'svelte';
	import { Download, Copy, Check, Loader2, FileCode, FileText } from 'lucide-svelte';
	import { getFileContent, downloadFile } from '$lib/api/osa/files';
	import type { OSAFile } from './types';
	import { marked } from 'marked';

	interface Props {
		file: OSAFile | null;
	}

	let { file = null }: Props = $props();

	let content = $state<string | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let copied = $state(false);
	let renderedMarkdown = $state<string>('');

	// Load file content when file changes
	$effect(() => {
		if (file) {
			loadContent();
		} else {
			content = null;
			error = null;
			renderedMarkdown = '';
		}
	});

	async function loadContent() {
		if (!file) return;

		loading = true;
		error = null;
		content = null;
		renderedMarkdown = '';

		try {
			const response = await getFileContent(file.id);
			content = response.content;

			// Render markdown if applicable
			if (file.type === 'markdown' || file.name.endsWith('.md')) {
				renderedMarkdown = await marked.parse(content);
			}
		} catch (err: any) {
			console.error('Failed to load file content:', err);
			error = err?.message || 'Failed to load file content';
		} finally {
			loading = false;
		}
	}

	async function handleCopy() {
		if (!content) return;

		try {
			await navigator.clipboard.writeText(content);
			copied = true;
			setTimeout(() => {
				copied = false;
			}, 2000);
		} catch (err) {
			console.error('Failed to copy:', err);
		}
	}

	async function handleDownload() {
		if (!file) return;

		try {
			const blob = await downloadFile(file.id);
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = file.name;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err) {
			console.error('Failed to download file:', err);
		}
	}

	function getLanguageFromFile(file: OSAFile): string {
		if (file.language) return file.language;

		// Infer from extension
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
		return Math.round(bytes / Math.pow(k, i) * 10) / 10 + ' ' + sizes[i];
	}
</script>

<div class="file-preview">
	{#if !file}
		<div class="empty-state">
			<FileText size={48} class="text-gray-400" />
			<p class="empty-text">Select a file to preview</p>
		</div>
	{:else}
		<!-- Header -->
		<div class="preview-header">
			<div class="file-info">
				<FileCode size={20} class="text-blue-400" />
				<div class="file-details">
					<h3 class="file-name">{file.name}</h3>
					<p class="file-meta">
						{formatFileSize(file.size)} • {file.type} • Updated {formatDate(file.updated_at)}
					</p>
				</div>
			</div>

			<div class="actions">
				<button class="action-btn" onclick={handleCopy} disabled={!content || loading}>
					{#if copied}
						<Check size={16} class="text-green-400" />
					{:else}
						<Copy size={16} />
					{/if}
					<span>{copied ? 'Copied!' : 'Copy'}</span>
				</button>

				<button class="action-btn" onclick={handleDownload}>
					<Download size={16} />
					<span>Download</span>
				</button>
			</div>
		</div>

		<!-- Content -->
		<div class="preview-content">
			{#if loading}
				<div class="loading-state">
					<Loader2 size={32} class="animate-spin text-blue-400" />
					<p>Loading file content...</p>
				</div>
			{:else if error}
				<div class="error-state">
					<p class="error-text">{error}</p>
				</div>
			{:else if content}
				{#if renderedMarkdown}
					<!-- Markdown preview -->
					<div class="markdown-preview">
						{@html renderedMarkdown}
					</div>
				{:else}
					<!-- Code preview -->
					<pre class="code-preview"><code class="language-{getLanguageFromFile(file)}">{content}</code></pre>
				{/if}
			{/if}
		</div>
	{/if}
</div>

<style>
	.file-preview {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #1f2937;
		border-radius: 8px;
		overflow: hidden;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: #6b7280;
	}

	.empty-text {
		font-size: 16px;
		color: #9ca3af;
	}

	.preview-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		background: #374151;
		border-bottom: 1px solid #4b5563;
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

	.file-name {
		font-size: 16px;
		font-weight: 600;
		color: #f9fafb;
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.file-meta {
		font-size: 12px;
		color: #9ca3af;
		margin: 4px 0 0 0;
	}

	.actions {
		display: flex;
		gap: 8px;
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		background: #1f2937;
		border: 1px solid #4b5563;
		border-radius: 6px;
		color: #e5e7eb;
		font-size: 14px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.action-btn:hover:not(:disabled) {
		background: #374151;
		border-color: #60a5fa;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.preview-content {
		flex: 1;
		overflow: auto;
		padding: 20px;
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
	}

	.error-text {
		color: #ef4444;
		font-size: 14px;
	}

	.code-preview {
		margin: 0;
		padding: 16px;
		background: #111827;
		border-radius: 6px;
		overflow-x: auto;
		font-family: 'Monaco', 'Courier New', monospace;
		font-size: 13px;
		line-height: 1.6;
		color: #e5e7eb;
	}

	.code-preview code {
		display: block;
	}

	.markdown-preview {
		color: #e5e7eb;
		line-height: 1.7;
	}

	.markdown-preview :global(h1),
	.markdown-preview :global(h2),
	.markdown-preview :global(h3),
	.markdown-preview :global(h4),
	.markdown-preview :global(h5),
	.markdown-preview :global(h6) {
		color: #f9fafb;
		margin-top: 24px;
		margin-bottom: 16px;
		font-weight: 600;
	}

	.markdown-preview :global(h1) {
		font-size: 28px;
		border-bottom: 1px solid #4b5563;
		padding-bottom: 8px;
	}

	.markdown-preview :global(h2) {
		font-size: 24px;
		border-bottom: 1px solid #374151;
		padding-bottom: 6px;
	}

	.markdown-preview :global(h3) {
		font-size: 20px;
	}

	.markdown-preview :global(code) {
		background: #111827;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'Monaco', 'Courier New', monospace;
		font-size: 13px;
	}

	.markdown-preview :global(pre) {
		background: #111827;
		padding: 16px;
		border-radius: 6px;
		overflow-x: auto;
		margin: 16px 0;
	}

	.markdown-preview :global(pre code) {
		background: transparent;
		padding: 0;
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
		border-left: 4px solid #4b5563;
		padding-left: 16px;
		margin: 16px 0;
		color: #9ca3af;
	}

	.markdown-preview :global(table) {
		border-collapse: collapse;
		width: 100%;
		margin: 16px 0;
	}

	.markdown-preview :global(th),
	.markdown-preview :global(td) {
		border: 1px solid #4b5563;
		padding: 8px 12px;
		text-align: left;
	}

	.markdown-preview :global(th) {
		background: #374151;
		font-weight: 600;
	}
</style>
