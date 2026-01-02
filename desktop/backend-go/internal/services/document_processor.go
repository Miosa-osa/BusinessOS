package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ledongthuc/pdf"
	"github.com/pgvector/pgvector-go"
)

// DocumentProcessor handles document processing, chunking, and semantic search
type DocumentProcessor struct {
	pool             *pgxpool.Pool
	embeddingService *EmbeddingService
	logger           *slog.Logger
	storagePath      string
}

// NewDocumentProcessor creates a new document processor
func NewDocumentProcessor(pool *pgxpool.Pool, embeddingService *EmbeddingService, storagePath string) *DocumentProcessor {
	if storagePath == "" {
		storagePath = "./uploads"
	}
	return &DocumentProcessor{
		pool:             pool,
		embeddingService: embeddingService,
		logger:           slog.Default().With("service", "document_processor"),
		storagePath:      storagePath,
	}
}

// ============================================================================
// Types
// ============================================================================

// DocumentUpload represents an uploaded document
type DocumentUpload struct {
	ID               uuid.UUID  `json:"id"`
	UserID           string     `json:"user_id"`
	Filename         string     `json:"filename"`
	OriginalFilename string     `json:"original_filename"`
	DisplayName      string     `json:"display_name,omitempty"`
	Description      string     `json:"description,omitempty"`
	FileType         string     `json:"file_type"`
	MimeType         string     `json:"mime_type"`
	FileSizeBytes    int64      `json:"file_size_bytes"`
	StoragePath      string     `json:"storage_path"`
	ExtractedText    string     `json:"extracted_text,omitempty"`
	PageCount        int        `json:"page_count,omitempty"`
	WordCount        int        `json:"word_count,omitempty"`
	DocumentType     string     `json:"document_type,omitempty"`
	Category         string     `json:"category,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	ProjectID        *uuid.UUID `json:"project_id,omitempty"`
	NodeID           *uuid.UUID `json:"node_id,omitempty"`
	ProcessingStatus string     `json:"processing_status"`
	ProcessingError  string     `json:"processing_error,omitempty"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// DocumentChunk represents a chunk of a document
type DocumentChunk struct {
	ID           uuid.UUID `json:"id"`
	DocumentID   uuid.UUID `json:"document_id"`
	ChunkIndex   int       `json:"chunk_index"`
	Content      string    `json:"content"`
	TokenCount   int       `json:"token_count"`
	PageNumber   *int      `json:"page_number,omitempty"`
	StartChar    int       `json:"start_char"`
	EndChar      int       `json:"end_char"`
	SectionTitle string    `json:"section_title,omitempty"`
	ChunkType    string    `json:"chunk_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProcessDocumentInput contains parameters for processing a document
type ProcessDocumentInput struct {
	UserID           string
	Filename         string
	OriginalFilename string
	DisplayName      string
	Description      string
	MimeType         string
	Content          []byte
	ProjectID        *uuid.UUID
	NodeID           *uuid.UUID
	DocumentType     string
	Category         string
	Tags             []string
}

// ChunkingOptions configures how documents are chunked
type ChunkingOptions struct {
	MaxChunkSize    int  // Maximum characters per chunk
	ChunkOverlap    int  // Character overlap between chunks
	PreserveHeaders bool // Try to keep headers with their content
	SplitOnHeaders  bool // Split at markdown headers
}

// DefaultChunkingOptions returns sensible defaults
func DefaultChunkingOptions() ChunkingOptions {
	return ChunkingOptions{
		MaxChunkSize:    1500,
		ChunkOverlap:    200,
		PreserveHeaders: true,
		SplitOnHeaders:  true,
	}
}

// DocumentSearchResult represents a search result
type DocumentSearchResult struct {
	DocumentID      uuid.UUID `json:"document_id"`
	ChunkID         uuid.UUID `json:"chunk_id"`
	DocumentTitle   string    `json:"document_title"`
	ChunkContent    string    `json:"chunk_content"`
	RelevanceScore  float64   `json:"relevance_score"`
	PageNumber      *int      `json:"page_number,omitempty"`
	SectionTitle    string    `json:"section_title,omitempty"`
	DocumentType    string    `json:"document_type,omitempty"`
}

// ============================================================================
// Document Processing
// ============================================================================

// ProcessDocument processes an uploaded document
func (p *DocumentProcessor) ProcessDocument(ctx context.Context, input ProcessDocumentInput) (*DocumentUpload, error) {
	p.logger.Info("processing document", "filename", input.OriginalFilename, "size", len(input.Content))

	// Determine file type from mime type
	fileType := p.getFileType(input.MimeType)
	if fileType == "" {
		return nil, fmt.Errorf("unsupported file type: %s", input.MimeType)
	}

	// Create document record
	doc := &DocumentUpload{
		ID:               uuid.New(),
		UserID:           input.UserID,
		Filename:         input.Filename,
		OriginalFilename: input.OriginalFilename,
		DisplayName:      input.DisplayName,
		Description:      input.Description,
		FileType:         fileType,
		MimeType:         input.MimeType,
		FileSizeBytes:    int64(len(input.Content)),
		ProjectID:        input.ProjectID,
		NodeID:           input.NodeID,
		DocumentType:     input.DocumentType,
		Category:         input.Category,
		Tags:             input.Tags,
		ProcessingStatus: "processing",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save file to storage
	storagePath, err := p.saveFile(input.UserID, doc.ID.String(), input.Filename, input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	doc.StoragePath = storagePath

	// Insert document record
	if err := p.insertDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	// Process document asynchronously
	go p.processDocumentAsync(doc.ID, input.Content)

	return doc, nil
}

// processDocumentAsync processes document content in background
func (p *DocumentProcessor) processDocumentAsync(docID uuid.UUID, content []byte) {
	ctx := context.Background()

	// Extract text
	extractedText, pageCount, err := p.extractText(content)
	if err != nil {
		p.updateProcessingStatus(ctx, docID, "failed", err.Error())
		return
	}

	// Count words
	wordCount := p.countWords(extractedText)

	// Update document with extracted text
	_, err = p.pool.Exec(ctx, `
		UPDATE uploaded_documents
		SET extracted_text = $1, page_count = $2, word_count = $3, updated_at = NOW()
		WHERE id = $4
	`, extractedText, pageCount, wordCount, docID)
	if err != nil {
		p.logger.Error("failed to update extracted text", "error", err, "doc_id", docID)
	}

	// Generate embedding for the whole document
	if p.embeddingService != nil && len(extractedText) > 0 {
		// Use summary or first 8000 chars for document-level embedding
		textForEmbedding := extractedText
		if len(textForEmbedding) > 8000 {
			textForEmbedding = textForEmbedding[:8000]
		}

		embedding, err := p.embeddingService.GenerateEmbedding(ctx, textForEmbedding)
		if err == nil && len(embedding) > 0 {
			vec := pgvector.NewVector(embedding)
			if _, execErr := p.pool.Exec(ctx, `UPDATE uploaded_documents SET embedding = $1 WHERE id = $2`, vec, docID); execErr != nil {
				p.logger.Warn("failed to store document embedding", "error", execErr, "doc_id", docID)
			}
		}
	}

	// Chunk the document
	chunks := p.chunkDocument(extractedText, DefaultChunkingOptions())

	// Insert chunks and generate embeddings
	for i, chunk := range chunks {
		chunkID := uuid.New()

		// Insert chunk
		_, err := p.pool.Exec(ctx, `
			INSERT INTO document_chunks (id, document_id, chunk_index, content, token_count, start_char, end_char, section_title, chunk_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, chunkID, docID, i, chunk.Content, chunk.TokenCount, chunk.StartChar, chunk.EndChar, chunk.SectionTitle, chunk.ChunkType)

		if err != nil {
			p.logger.Error("failed to insert chunk", "error", err, "doc_id", docID, "chunk", i)
			continue
		}

		// Generate chunk embedding
		if p.embeddingService != nil {
			embedding, err := p.embeddingService.GenerateEmbedding(ctx, chunk.Content)
			if err == nil && len(embedding) > 0 {
				vec := pgvector.NewVector(embedding)
				if _, execErr := p.pool.Exec(ctx, `UPDATE document_chunks SET embedding = $1 WHERE id = $2`, vec, chunkID); execErr != nil {
					p.logger.Warn("failed to store chunk embedding", "error", execErr, "doc_id", docID, "chunk_id", chunkID, "chunk", i)
				}
			}
		}
	}

	// Mark as completed
	now := time.Now()
	p.pool.Exec(ctx, `
		UPDATE uploaded_documents SET processing_status = 'completed', processed_at = $1, updated_at = NOW() WHERE id = $2
	`, now, docID)

	p.logger.Info("document processing completed", "doc_id", docID, "chunks", len(chunks), "words", wordCount)
}

// ============================================================================
// Text Extraction
// ============================================================================

// extractText extracts text from document content
func (p *DocumentProcessor) extractText(content []byte) (string, int, error) {
	// Check if it's plain text or markdown (UTF-8 valid and mostly printable)
	if utf8.Valid(content) && p.isProbablyText(content) {
		text := string(content)
		// Count "pages" based on ~3000 chars per page
		pageCount := (len(text) / 3000) + 1
		return text, pageCount, nil
	}

	// Try PDF extraction
	if p.isPDF(content) {
		return p.extractPDFText(content)
	}

	// Try DOCX extraction
	if p.isDOCX(content) {
		return p.extractDOCXText(content)
	}

	return "", 0, fmt.Errorf("unsupported binary format - upload as PDF, DOCX, or text")
}

// isProbablyText checks if content is likely plain text (not binary disguised as UTF-8)
func (p *DocumentProcessor) isProbablyText(content []byte) bool {
	if len(content) == 0 {
		return true
	}

	// Check first 1000 bytes for binary indicators
	checkLen := len(content)
	if checkLen > 1000 {
		checkLen = 1000
	}

	nullCount := 0
	controlCount := 0
	for i := 0; i < checkLen; i++ {
		b := content[i]
		if b == 0 {
			nullCount++
		}
		// Count control characters (except common ones like tab, newline, carriage return)
		if b < 32 && b != 9 && b != 10 && b != 13 {
			controlCount++
		}
	}

	// If more than 1% null bytes or 5% control chars, probably binary
	return nullCount < checkLen/100 && controlCount < checkLen/20
}

// isPDF checks if content is a PDF file
func (p *DocumentProcessor) isPDF(content []byte) bool {
	return len(content) > 4 && string(content[:4]) == "%PDF"
}

// isDOCX checks if content is a DOCX file (ZIP with specific structure)
func (p *DocumentProcessor) isDOCX(content []byte) bool {
	// DOCX files are ZIP files starting with PK signature
	if len(content) < 4 {
		return false
	}
	// Check ZIP signature
	if content[0] != 0x50 || content[1] != 0x4B {
		return false
	}
	// Try to open as ZIP and check for word/document.xml
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return false
	}
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			return true
		}
	}
	return false
}

// extractPDFText extracts text from PDF content
func (p *DocumentProcessor) extractPDFText(content []byte) (string, int, error) {
	p.logger.Info("extracting text from PDF", "size", len(content))

	// Create a temporary file for pdf library (it requires file path)
	tmpFile, err := os.CreateTemp("", "pdf_*.pdf")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		return "", 0, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Open PDF
	f, r, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return "", 0, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	pageCount := r.NumPage()
	if pageCount == 0 {
		return "", 0, fmt.Errorf("PDF has no pages")
	}

	var textBuilder strings.Builder

	// Extract text from each page
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			p.logger.Warn("failed to extract text from page", "page", pageNum, "error", err)
			continue
		}

		if textBuilder.Len() > 0 {
			textBuilder.WriteString("\n\n--- Page ")
			textBuilder.WriteString(fmt.Sprintf("%d", pageNum))
			textBuilder.WriteString(" ---\n\n")
		}
		textBuilder.WriteString(text)
	}

	extractedText := strings.TrimSpace(textBuilder.String())
	if len(extractedText) == 0 {
		return "", pageCount, fmt.Errorf("no text could be extracted from PDF (may be image-based)")
	}

	p.logger.Info("PDF text extracted", "pages", pageCount, "chars", len(extractedText))
	return extractedText, pageCount, nil
}

// extractDOCXText extracts text from DOCX content
func (p *DocumentProcessor) extractDOCXText(content []byte) (string, int, error) {
	p.logger.Info("extracting text from DOCX", "size", len(content))

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", 0, fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}

	var documentXML *zip.File
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			documentXML = f
			break
		}
	}

	if documentXML == nil {
		return "", 0, fmt.Errorf("word/document.xml not found in DOCX")
	}

	rc, err := documentXML.Open()
	if err != nil {
		return "", 0, fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	xmlContent, err := io.ReadAll(rc)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read document.xml: %w", err)
	}

	// Parse DOCX XML and extract text
	extractedText := p.parseDocxXML(xmlContent)

	// Estimate page count (~3000 chars per page)
	pageCount := (len(extractedText) / 3000) + 1

	p.logger.Info("DOCX text extracted", "chars", len(extractedText), "estimated_pages", pageCount)
	return extractedText, pageCount, nil
}

// docxDocument represents the simplified DOCX XML structure
type docxDocument struct {
	Body docxBody `xml:"body"`
}

type docxBody struct {
	Paragraphs []docxParagraph `xml:"p"`
}

type docxParagraph struct {
	Runs []docxRun `xml:"r"`
}

type docxRun struct {
	Text    string `xml:"t"`
	Tab     string `xml:"tab"`
	Break   string `xml:"br"`
	InnerT  []docxText `xml:",any"`
}

type docxText struct {
	XMLName xml.Name
	Content string `xml:",chardata"`
}

// parseDocxXML extracts plain text from DOCX XML content
func (p *DocumentProcessor) parseDocxXML(xmlContent []byte) string {
	var textBuilder strings.Builder

	// Use a simple approach: extract all text between <w:t> tags
	// This handles the complex namespace issues better than full XML parsing
	decoder := xml.NewDecoder(bytes.NewReader(xmlContent))

	var inTextElement bool
	var currentParagraph strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			p.logger.Warn("XML parsing error", "error", err)
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			localName := t.Name.Local
			if localName == "t" {
				inTextElement = true
			} else if localName == "tab" {
				currentParagraph.WriteString("\t")
			} else if localName == "br" {
				currentParagraph.WriteString("\n")
			}
		case xml.EndElement:
			localName := t.Name.Local
			if localName == "t" {
				inTextElement = false
			} else if localName == "p" {
				// End of paragraph
				if currentParagraph.Len() > 0 {
					if textBuilder.Len() > 0 {
						textBuilder.WriteString("\n")
					}
					textBuilder.WriteString(strings.TrimSpace(currentParagraph.String()))
					currentParagraph.Reset()
				}
			}
		case xml.CharData:
			if inTextElement {
				currentParagraph.Write(t)
			}
		}
	}

	// Add any remaining paragraph
	if currentParagraph.Len() > 0 {
		if textBuilder.Len() > 0 {
			textBuilder.WriteString("\n")
		}
		textBuilder.WriteString(strings.TrimSpace(currentParagraph.String()))
	}

	return strings.TrimSpace(textBuilder.String())
}

// ============================================================================
// Chunking
// ============================================================================

// chunkDocument splits document into chunks for RAG
func (p *DocumentProcessor) chunkDocument(text string, opts ChunkingOptions) []DocumentChunk {
	if len(text) == 0 {
		return nil
	}

	var chunks []DocumentChunk
	var currentSection string

	// Split by headers if enabled
	if opts.SplitOnHeaders {
		chunks = p.chunkByHeaders(text, opts)
	} else {
		chunks = p.chunkBySize(text, opts, currentSection)
	}

	return chunks
}

// chunkByHeaders splits text at markdown headers
func (p *DocumentProcessor) chunkByHeaders(text string, opts ChunkingOptions) []DocumentChunk {
	var chunks []DocumentChunk

	// Regex for markdown headers
	headerPattern := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

	// Find all headers
	matches := headerPattern.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		// No headers, chunk by size
		return p.chunkBySize(text, opts, "")
	}

	// Process sections between headers
	for i, match := range matches {
		headerStart := match[0]
		headerEnd := match[1]
		titleStart := match[4]
		titleEnd := match[5]

		sectionTitle := text[titleStart:titleEnd]

		// Determine section end
		var sectionEnd int
		if i+1 < len(matches) {
			sectionEnd = matches[i+1][0]
		} else {
			sectionEnd = len(text)
		}

		sectionContent := text[headerEnd:sectionEnd]
		sectionContent = strings.TrimSpace(sectionContent)

		if len(sectionContent) == 0 {
			continue
		}

		// If section is too large, chunk it further
		if len(sectionContent) > opts.MaxChunkSize {
			subChunks := p.chunkBySize(sectionContent, opts, sectionTitle)
			for _, sc := range subChunks {
				sc.StartChar += headerStart
				sc.EndChar += headerStart
				chunks = append(chunks, sc)
			}
		} else {
			chunks = append(chunks, DocumentChunk{
				Content:      sectionContent,
				TokenCount:   p.estimateTokens(sectionContent),
				StartChar:    headerStart,
				EndChar:      sectionEnd,
				SectionTitle: sectionTitle,
				ChunkType:    "text",
			})
		}
	}

	// Handle text before first header
	if len(matches) > 0 && matches[0][0] > 0 {
		preContent := strings.TrimSpace(text[:matches[0][0]])
		if len(preContent) > 0 {
			preChunks := p.chunkBySize(preContent, opts, "")
			chunks = append(preChunks, chunks...)
		}
	}

	return chunks
}

// chunkBySize splits text into fixed-size chunks with overlap
func (p *DocumentProcessor) chunkBySize(text string, opts ChunkingOptions, sectionTitle string) []DocumentChunk {
	var chunks []DocumentChunk

	if len(text) <= opts.MaxChunkSize {
		chunks = append(chunks, DocumentChunk{
			Content:      text,
			TokenCount:   p.estimateTokens(text),
			StartChar:    0,
			EndChar:      len(text),
			SectionTitle: sectionTitle,
			ChunkType:    "text",
		})
		return chunks
	}

	// Split into sentences for better boundaries
	sentences := p.splitIntoSentences(text)

	var currentChunk strings.Builder
	var chunkStart int
	currentStart := 0

	for _, sentence := range sentences {
		// If adding this sentence would exceed max size, finalize current chunk
		if currentChunk.Len()+len(sentence) > opts.MaxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, DocumentChunk{
				Content:      currentChunk.String(),
				TokenCount:   p.estimateTokens(currentChunk.String()),
				StartChar:    chunkStart,
				EndChar:      currentStart,
				SectionTitle: sectionTitle,
				ChunkType:    "text",
			})

			// Start new chunk with overlap
			overlapText := p.getOverlapText(currentChunk.String(), opts.ChunkOverlap)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			chunkStart = currentStart - len(overlapText)
		}

		currentChunk.WriteString(sentence)
		currentStart += len(sentence)
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, DocumentChunk{
			Content:      currentChunk.String(),
			TokenCount:   p.estimateTokens(currentChunk.String()),
			StartChar:    chunkStart,
			EndChar:      len(text),
			SectionTitle: sectionTitle,
			ChunkType:    "text",
		})
	}

	return chunks
}

// splitIntoSentences splits text into sentences
func (p *DocumentProcessor) splitIntoSentences(text string) []string {
	// Simple sentence splitting - could be improved with NLP
	pattern := regexp.MustCompile(`([.!?]+[\s]+)`)
	parts := pattern.Split(text, -1)
	delimiters := pattern.FindAllString(text, -1)

	var sentences []string
	for i, part := range parts {
		if i < len(delimiters) {
			sentences = append(sentences, part+delimiters[i])
		} else {
			sentences = append(sentences, part)
		}
	}

	return sentences
}

// getOverlapText returns the last N characters for overlap
func (p *DocumentProcessor) getOverlapText(text string, overlap int) string {
	if len(text) <= overlap {
		return text
	}

	// Try to find a sentence boundary within overlap range
	overlapText := text[len(text)-overlap:]

	// Find first sentence start
	sentenceStart := strings.Index(overlapText, ". ")
	if sentenceStart > 0 && sentenceStart < overlap/2 {
		return overlapText[sentenceStart+2:]
	}

	return overlapText
}

// ============================================================================
// Semantic Search
// ============================================================================

// SearchDocuments performs semantic search on documents
func (p *DocumentProcessor) SearchDocuments(ctx context.Context, userID, query string, limit int, projectID, nodeID *uuid.UUID) ([]DocumentSearchResult, error) {
	if p.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not available")
	}

	if limit <= 0 {
		limit = 10
	}

	// Generate query embedding
	queryEmbedding, err := p.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	vec := pgvector.NewVector(queryEmbedding)

	// Search chunks
	queryStr := `
		SELECT dc.id, dc.document_id, dc.content, dc.page_number, dc.section_title,
		       ud.display_name, ud.document_type,
		       1 - (dc.embedding <=> $1) as similarity
		FROM document_chunks dc
		JOIN uploaded_documents ud ON dc.document_id = ud.id
		WHERE ud.user_id = $2 AND dc.embedding IS NOT NULL
	`
	args := []interface{}{vec, userID}
	argIdx := 3

	if projectID != nil {
		queryStr += fmt.Sprintf(" AND ud.project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	}

	if nodeID != nil {
		queryStr += fmt.Sprintf(" AND ud.node_id = $%d", argIdx)
		args = append(args, *nodeID)
		argIdx++
	}

	queryStr += fmt.Sprintf(" ORDER BY dc.embedding <=> $1 LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := p.pool.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var results []DocumentSearchResult
	for rows.Next() {
		var r DocumentSearchResult
		var displayName, docType *string

		err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.ChunkContent, &r.PageNumber,
			&r.SectionTitle, &displayName, &docType, &r.RelevanceScore)
		if err != nil {
			continue
		}

		if displayName != nil {
			r.DocumentTitle = *displayName
		}
		if docType != nil {
			r.DocumentType = *docType
		}

		results = append(results, r)
	}

	return results, nil
}

// GetRelevantChunks retrieves chunks most relevant to a context
func (p *DocumentProcessor) GetRelevantChunks(ctx context.Context, userID, contextText string, limit int) ([]DocumentChunk, error) {
	if p.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not available")
	}

	if limit <= 0 {
		limit = 5
	}

	// Generate embedding
	embedding, err := p.embeddingService.GenerateEmbedding(ctx, contextText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	vec := pgvector.NewVector(embedding)

	rows, err := p.pool.Query(ctx, `
		SELECT dc.id, dc.document_id, dc.chunk_index, dc.content, dc.token_count,
		       dc.page_number, dc.start_char, dc.end_char, dc.section_title, dc.chunk_type
		FROM document_chunks dc
		JOIN uploaded_documents ud ON dc.document_id = ud.id
		WHERE ud.user_id = $1 AND dc.embedding IS NOT NULL
		ORDER BY dc.embedding <=> $2
		LIMIT $3
	`, userID, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var chunks []DocumentChunk
	for rows.Next() {
		var c DocumentChunk
		err := rows.Scan(&c.ID, &c.DocumentID, &c.ChunkIndex, &c.Content, &c.TokenCount,
			&c.PageNumber, &c.StartChar, &c.EndChar, &c.SectionTitle, &c.ChunkType)
		if err != nil {
			continue
		}
		chunks = append(chunks, c)
	}

	return chunks, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (p *DocumentProcessor) getFileType(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "pdf"):
		return "pdf"
	case strings.Contains(mimeType, "markdown") || strings.HasSuffix(mimeType, "md"):
		return "markdown"
	case strings.Contains(mimeType, "word") || strings.Contains(mimeType, "docx"):
		return "docx"
	case strings.Contains(mimeType, "text/plain"):
		return "txt"
	case strings.Contains(mimeType, "image"):
		return "image"
	default:
		return ""
	}
}

func (p *DocumentProcessor) saveFile(userID, docID, filename string, content []byte) (string, error) {
	// Create user directory
	userDir := filepath.Join(p.storagePath, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	storageName := fmt.Sprintf("%s%s", docID, ext)
	storagePath := filepath.Join(userDir, storageName)

	// Write file
	if err := os.WriteFile(storagePath, content, 0644); err != nil {
		return "", err
	}

	return storagePath, nil
}

func (p *DocumentProcessor) insertDocument(ctx context.Context, doc *DocumentUpload) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO uploaded_documents (
			id, user_id, filename, original_filename, display_name, description,
			file_type, mime_type, file_size_bytes, storage_path,
			project_id, node_id, document_type, category, tags,
			processing_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, doc.ID, doc.UserID, doc.Filename, doc.OriginalFilename, doc.DisplayName, doc.Description,
		doc.FileType, doc.MimeType, doc.FileSizeBytes, doc.StoragePath,
		doc.ProjectID, doc.NodeID, doc.DocumentType, doc.Category, doc.Tags,
		doc.ProcessingStatus, doc.CreatedAt, doc.UpdatedAt)

	return err
}

func (p *DocumentProcessor) updateProcessingStatus(ctx context.Context, docID uuid.UUID, status, errorMsg string) {
	p.pool.Exec(ctx, `
		UPDATE uploaded_documents SET processing_status = $1, processing_error = $2, updated_at = NOW() WHERE id = $3
	`, status, errorMsg, docID)
}

func (p *DocumentProcessor) countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func (p *DocumentProcessor) estimateTokens(text string) int {
	// Rough estimate: ~4 characters per token for English
	return len(text) / 4
}

// GetDocument retrieves a document by ID
func (p *DocumentProcessor) GetDocument(ctx context.Context, userID string, docID uuid.UUID) (*DocumentUpload, error) {
	var doc DocumentUpload
	var processedAt *time.Time

	err := p.pool.QueryRow(ctx, `
		SELECT id, user_id, filename, original_filename, display_name, description,
		       file_type, mime_type, file_size_bytes, storage_path, extracted_text,
		       page_count, word_count, project_id, node_id, document_type, category, tags,
		       processing_status, processing_error, processed_at, created_at, updated_at
		FROM uploaded_documents
		WHERE id = $1 AND user_id = $2
	`, docID, userID).Scan(
		&doc.ID, &doc.UserID, &doc.Filename, &doc.OriginalFilename, &doc.DisplayName, &doc.Description,
		&doc.FileType, &doc.MimeType, &doc.FileSizeBytes, &doc.StoragePath, &doc.ExtractedText,
		&doc.PageCount, &doc.WordCount, &doc.ProjectID, &doc.NodeID, &doc.DocumentType, &doc.Category, &doc.Tags,
		&doc.ProcessingStatus, &doc.ProcessingError, &processedAt, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	doc.ProcessedAt = processedAt
	return &doc, nil
}

// GetDocumentContent reads the file content
func (p *DocumentProcessor) GetDocumentContent(ctx context.Context, userID string, docID uuid.UUID) ([]byte, error) {
	doc, err := p.GetDocument(ctx, userID, docID)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(doc.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

// DeleteDocument deletes a document and its chunks
func (p *DocumentProcessor) DeleteDocument(ctx context.Context, userID string, docID uuid.UUID) error {
	// Get document for file path
	doc, err := p.GetDocument(ctx, userID, docID)
	if err != nil {
		return err
	}

	// Delete from database (chunks deleted via cascade)
	_, err = p.pool.Exec(ctx, `DELETE FROM uploaded_documents WHERE id = $1 AND user_id = $2`, docID, userID)
	if err != nil {
		return err
	}

	// Delete file
	if doc.StoragePath != "" {
		os.Remove(doc.StoragePath)
	}

	return nil
}

// ReprocessDocument reprocesses an existing document
func (p *DocumentProcessor) ReprocessDocument(ctx context.Context, userID string, docID uuid.UUID) error {
	// Get document
	doc, err := p.GetDocument(ctx, userID, docID)
	if err != nil {
		return err
	}

	// Read content
	content, err := os.ReadFile(doc.StoragePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Delete existing chunks
	p.pool.Exec(ctx, `DELETE FROM document_chunks WHERE document_id = $1`, docID)

	// Update status
	p.pool.Exec(ctx, `UPDATE uploaded_documents SET processing_status = 'processing', updated_at = NOW() WHERE id = $1`, docID)

	// Reprocess
	go p.processDocumentAsync(docID, content)

	return nil
}

// Ensure buffer is imported
var _ = bytes.Buffer{}
var _ = io.Reader(nil)
