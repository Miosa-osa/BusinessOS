<!-- 
  FileUpload.svelte
  Drag and drop file upload component for data import
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';

  export let accept: string = ".csv,.xlsx,.json";
  export let maxSizeMB: number = 10;
  export let className: string = "";

  const dispatch = createEventDispatcher();

  let isDragging = false;
  let file: File | null = null;
  let error: string = "";

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    isDragging = false;

    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      validateAndSetFile(files[0]);
    }
  }

  function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      validateAndSetFile(input.files[0]);
    }
  }

  function validateAndSetFile(selectedFile: File) {
    error = "";

    // Check file type
    const validTypes = accept.split(',').map(t => t.trim());
    const fileExtension = '.' + selectedFile.name.split('.').pop()?.toLowerCase();
    
    if (!validTypes.includes(fileExtension)) {
      error = `Invalid file type. Accepted: ${accept}`;
      return;
    }

    // Check file size
    const maxSizeBytes = maxSizeMB * 1024 * 1024;
    if (selectedFile.size > maxSizeBytes) {
      error = `File too large. Maximum size: ${maxSizeMB}MB`;
      return;
    }

    file = selectedFile;
    dispatch('fileSelected', { file });
  }

  function handleRemove() {
    file = null;
    error = "";
    dispatch('fileRemoved');
  }

  function handleSkip() {
    dispatch('skip');
  }

  function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
</script>

<div class="file-upload {className}">
  <p class="title">📤 Upload your data</p>

  {#if !file}
    <div 
      class="drop-zone"
      class:dragging={isDragging}
      on:dragover={handleDragOver}
      on:dragleave={handleDragLeave}
      on:drop={handleDrop}
      role="button"
      tabindex="0"
    >
      <p class="drop-text">Drag & drop CSV, Excel, or JSON file here</p>
      <p class="drop-or">or</p>
      <label class="browse-btn">
        Browse files
        <input 
          type="file" 
          {accept}
          on:change={handleFileSelect}
          class="file-input"
        />
      </label>
    </div>
  {:else}
    <div class="file-preview">
      <div class="file-info">
        <span class="file-icon">📄</span>
        <div class="file-details">
          <span class="file-name">{file.name}</span>
          <span class="file-size">{formatFileSize(file.size)}</span>
        </div>
      </div>
      <button class="remove-btn" on:click={handleRemove}>×</button>
    </div>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <Button variant="ghost" on:click={handleSkip} className="skip-btn">
    Skip for now
  </Button>
</div>

<style>
  .file-upload {
    width: 100%;
    max-width: 28rem;
    padding: 1rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.75rem;
  }

  :global(.dark) .file-upload {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .title {
    margin: 0 0 1rem;
    font-size: 0.9375rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .title {
    color: var(--foreground, #f9fafb);
  }

  .drop-zone {
    padding: 2rem 1rem;
    border: 2px dashed var(--border, #e5e7eb);
    border-radius: 0.5rem;
    text-align: center;
    cursor: pointer;
    transition: all 150ms;
  }

  .drop-zone.dragging,
  .drop-zone:hover {
    border-color: var(--primary, #000000);
    background-color: var(--accent, #f3f4f6);
  }

  :global(.dark) .drop-zone {
    border-color: var(--border, #2a2a2a);
  }

  :global(.dark) .drop-zone.dragging,
  :global(.dark) .drop-zone:hover {
    border-color: var(--primary, #ffffff);
    background-color: var(--accent, #2a2a2a);
  }

  .drop-text {
    margin: 0;
    font-size: 0.875rem;
    color: var(--muted-foreground, #6b7280);
  }

  .drop-or {
    margin: 0.5rem 0;
    font-size: 0.8125rem;
    color: var(--muted-foreground, #6b7280);
  }

  .browse-btn {
    display: inline-block;
    padding: 0.5rem 1rem;
    background-color: var(--primary, #000000);
    color: var(--primary-foreground, #ffffff);
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 150ms;
  }

  .browse-btn:hover {
    opacity: 0.9;
  }

  :global(.dark) .browse-btn {
    background-color: var(--primary, #ffffff);
    color: var(--primary-foreground, #000000);
  }

  .file-input {
    display: none;
  }

  .file-preview {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem;
    background-color: var(--muted, #f9fafb);
    border-radius: 0.5rem;
  }

  :global(.dark) .file-preview {
    background-color: var(--muted, #1a1a1a);
  }

  .file-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .file-icon {
    font-size: 1.5rem;
  }

  .file-details {
    display: flex;
    flex-direction: column;
  }

  .file-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .file-name {
    color: var(--foreground, #f9fafb);
  }

  .file-size {
    font-size: 0.75rem;
    color: var(--muted-foreground, #6b7280);
  }

  .remove-btn {
    width: 1.75rem;
    height: 1.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 1.25rem;
    color: var(--muted-foreground, #6b7280);
    border-radius: 0.25rem;
    transition: all 150ms;
  }

  .remove-btn:hover {
    background-color: var(--destructive, #ef4444);
    color: white;
  }

  .error {
    margin: 0.5rem 0 0;
    font-size: 0.8125rem;
    color: var(--error, #ef4444);
  }

  :global(.skip-btn) {
    margin-top: 0.75rem !important;
    width: 100% !important;
  }
</style>
