# Pedro Tasks: Document Processing

> **Priority:** P1 - High Value (Pedro's Work)
> **Backend Status:** Complete (8 endpoints)
> **Frontend Status:** Not Started
> **Owner:** Pedro
> **Estimated Effort:** 1 sprint

---

## Overview

Document processing enables uploading, parsing, chunking, and semantic search of documents. This is foundational for RAG (Retrieval-Augmented Generation) - allowing AI to reference uploaded documents.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/documents` | Upload and process document |
| GET | `/api/documents` | List documents with filters |
| POST | `/api/documents/search` | Semantic document search |
| POST | `/api/documents/chunks` | Get relevant text chunks |
| GET | `/api/documents/:id` | Get document details |
| DELETE | `/api/documents/:id` | Delete document |
| POST | `/api/documents/:id/reprocess` | Re-process document |
| GET | `/api/documents/:id/content` | Download raw content |

---

## Data Models

```typescript
interface Document {
  id: string;
  user_id: string;
  workspace_id: string;

  // File info
  display_name: string;
  original_filename: string;
  mime_type: string;
  file_size: number;

  // Metadata
  description?: string;
  document_type: DocumentType;
  category?: string;

  // Associations
  project_id?: string;
  node_id?: string;

  // Processing
  status: 'pending' | 'processing' | 'ready' | 'error';
  chunk_count: number;
  embedding_status: 'pending' | 'complete' | 'error';
  error_message?: string;

  created_at: string;
  updated_at: string;
  processed_at?: string;
}

type DocumentType =
  | 'pdf'
  | 'doc'
  | 'docx'
  | 'txt'
  | 'md'
  | 'html'
  | 'csv'
  | 'xlsx'
  | 'pptx'
  | 'image'
  | 'other';

interface DocumentChunk {
  id: string;
  document_id: string;
  content: string;
  chunk_index: number;
  start_page?: number;
  end_page?: number;
  metadata: Record<string, any>;
  embedding: number[];
}
```

---

## Frontend Implementation Tasks

### Phase 1: Document Upload

#### 1.1 Upload Interface
**File:** `src/lib/components/documents/DocumentUpload.svelte`

- [ ] Drag-and-drop zone
- [ ] File type validation
- [ ] Size limit display (50MB max)
- [ ] Multiple file upload
- [ ] Upload progress indicator

#### 1.2 Upload Form
**File:** `src/lib/components/documents/UploadForm.svelte`

- [ ] Display name input
- [ ] Description textarea
- [ ] Document type selector (auto-detect from extension)
- [ ] Category input
- [ ] Project association dropdown
- [ ] Node association dropdown

#### 1.3 Processing Status
- [ ] Real-time processing status
- [ ] Chunking progress
- [ ] Embedding progress
- [ ] Error display with retry

### Phase 2: Document Library

#### 2.1 Documents Page
**File:** `src/routes/(app)/documents/+page.svelte`

- [ ] Grid/List view toggle
- [ ] Filter by: type, category, project, node, status
- [ ] Search documents
- [ ] Sort by: name, date, size

#### 2.2 Document Card
**File:** `src/lib/components/documents/DocumentCard.svelte`

```svelte
<div class="document-card">
  <DocumentIcon type={doc.document_type} />
  <h3>{doc.display_name}</h3>
  <p class="meta">
    {formatFileSize(doc.file_size)} · {doc.chunk_count} chunks
  </p>
  <Badge>{doc.status}</Badge>
  {#if doc.project_id}
    <Link to="/projects/{doc.project_id}">View Project</Link>
  {/if}
  <DropdownMenu>
    <DropdownItem on:click={() => download(doc)}>Download</DropdownItem>
    <DropdownItem on:click={() => reprocess(doc)}>Reprocess</DropdownItem>
    <DropdownItem on:click={() => delete(doc)} class="text-red">Delete</DropdownItem>
  </DropdownMenu>
</div>
```

#### 2.3 Document Detail View
**File:** `src/routes/(app)/documents/[id]/+page.svelte`

- [ ] Document metadata display
- [ ] Preview (if applicable)
- [ ] Chunks list with pagination
- [ ] Related conversations
- [ ] Edit metadata
- [ ] Delete with confirmation

### Phase 3: Document Search

#### 3.1 Search Interface
**File:** `src/lib/components/documents/DocumentSearch.svelte`

- [ ] Search input
- [ ] Results list with relevance scores
- [ ] Highlighted matching text
- [ ] Jump to chunk in document

#### 3.2 Search Results
- [ ] Show matching chunks
- [ ] Expand to see context
- [ ] "Use in chat" button

### Phase 4: Chat Integration

#### 4.1 Document Context in Chat
- [ ] "Add document" button in chat
- [ ] Document picker modal
- [ ] Show attached documents
- [ ] AI references document in responses

#### 4.2 Document Citations
- [ ] Link to source document in AI response
- [ ] Show which chunks were used
- [ ] Page numbers if applicable

### Phase 5: API Client

#### 5.1 Documents API
**File:** `src/lib/api/documents/documents.ts`

```typescript
export async function uploadDocument(
  file: File,
  metadata: DocumentMetadata
): Promise<Document>

export async function getDocuments(
  filters?: DocumentFilters
): Promise<Document[]>

export async function getDocument(id: string): Promise<Document>

export async function deleteDocument(id: string): Promise<void>

export async function reprocessDocument(id: string): Promise<void>

export async function downloadDocument(id: string): Promise<Blob>

export async function searchDocuments(
  query: string,
  options?: SearchOptions
): Promise<SearchResult[]>

export async function getDocumentChunks(
  query: string,
  documentIds?: string[]
): Promise<DocumentChunk[]>
```

#### 5.2 Documents Store
**File:** `src/lib/stores/documents.ts`

```typescript
interface DocumentsStore {
  documents: Document[];
  isLoading: boolean;
  uploadProgress: number;

  loadDocuments(filters?: DocumentFilters): Promise<void>;
  uploadDocument(file: File, metadata: DocumentMetadata): Promise<Document>;
  deleteDocument(id: string): Promise<void>;
  reprocessDocument(id: string): Promise<void>;
  searchDocuments(query: string): Promise<SearchResult[]>;
}
```

---

## UI/UX Requirements

### Upload Experience
- Clear supported file types
- Progress bar during upload
- Processing status after upload
- Error recovery (retry)

### Document Preview
- PDF: Embedded viewer
- Images: Preview
- Text/Code: Syntax highlighted
- Office docs: Basic preview or "Open in..." prompt

### Search UX
- Instant search as you type
- Highlight matching terms
- Show relevance scores
- Easy to add to chat context

---

## Supported File Types

| Type | Extensions | Max Size |
|------|------------|----------|
| PDF | .pdf | 50MB |
| Word | .doc, .docx | 50MB |
| Text | .txt, .md | 10MB |
| HTML | .html, .htm | 10MB |
| Excel | .xlsx, .xls, .csv | 50MB |
| PowerPoint | .pptx, .ppt | 50MB |
| Images | .png, .jpg, .jpeg | 20MB |

---

## Testing Requirements

- [ ] Unit tests for documents store
- [ ] Component tests for upload, card
- [ ] E2E: Upload document flow
- [ ] E2E: Search documents
- [ ] E2E: Use document in chat

---

## Linear Issues to Create

1. **[DOC-001]** Create document upload interface
2. **[DOC-002]** Build upload form with metadata
3. **[DOC-003]** Implement processing status display
4. **[DOC-004]** Create Documents library page
5. **[DOC-005]** Build DocumentCard component
6. **[DOC-006]** Add document detail view
7. **[DOC-007]** Implement document search
8. **[DOC-008]** Integrate documents into chat
9. **[DOC-009]** Add document citations in responses
10. **[DOC-010]** API client and store
11. **[DOC-011]** E2E tests

---

## Dependencies

- Embedding service must be running
- pgvector extension in PostgreSQL

## Blockers

- None identified

---

## Notes

- Document processing is async - need good status feedback
- Consider document versioning in future
- OCR for scanned PDFs could be valuable
- Integration with Drive/Notion imports
