-- ================================================
-- Migration 024: Align embedding dimensions to 768
-- Description: Several tables were created with vector(1536) but the runtime embedding model is nomic-embed-text (768).
--              This migration resets embedding columns to vector(768) so semantic search/indexing works consistently.
-- Notes: Embeddings are derived data; this migration drops existing embedding columns (and indexes) and recreates them.
-- Date: 2026-01-02
-- ================================================

-- Ensure pgvector exists
CREATE EXTENSION IF NOT EXISTS vector;

-- Helper: reset a table's embedding column to vector(768)
-- We intentionally DROP COLUMN to avoid cast failures when old embeddings (1536) exist.

-- Memories
ALTER TABLE memories DROP COLUMN IF EXISTS embedding;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_memories_embedding;
CREATE INDEX IF NOT EXISTS idx_memories_embedding ON memories USING hnsw (embedding vector_cosine_ops);

-- Uploaded documents
ALTER TABLE uploaded_documents DROP COLUMN IF EXISTS embedding;
ALTER TABLE uploaded_documents ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_uploaded_docs_embedding;
CREATE INDEX IF NOT EXISTS idx_uploaded_docs_embedding ON uploaded_documents USING hnsw (embedding vector_cosine_ops);

-- Document chunks
ALTER TABLE document_chunks DROP COLUMN IF EXISTS embedding;
ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_doc_chunks_embedding;
CREATE INDEX IF NOT EXISTS idx_doc_chunks_embedding ON document_chunks USING hnsw (embedding vector_cosine_ops);

-- Conversations (context integration)
ALTER TABLE conversations DROP COLUMN IF EXISTS embedding;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_conversations_embedding;
CREATE INDEX IF NOT EXISTS idx_conversations_embedding ON conversations USING hnsw (embedding vector_cosine_ops);

-- Conversation summaries
ALTER TABLE conversation_summaries DROP COLUMN IF EXISTS embedding;
ALTER TABLE conversation_summaries ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_conv_summaries_embedding;
CREATE INDEX IF NOT EXISTS idx_conv_summaries_embedding ON conversation_summaries USING hnsw (embedding vector_cosine_ops);

-- Voice notes
ALTER TABLE voice_notes DROP COLUMN IF EXISTS embedding;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS embedding vector(768);
DROP INDEX IF EXISTS idx_voice_notes_embedding;
CREATE INDEX IF NOT EXISTS idx_voice_notes_embedding ON voice_notes USING hnsw (embedding vector_cosine_ops);

-- Optional tables (some environments may not have these yet)
DO $$
BEGIN
	IF to_regclass('public.context_profiles') IS NOT NULL THEN
		ALTER TABLE context_profiles DROP COLUMN IF EXISTS embedding;
		ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS embedding vector(768);
		DROP INDEX IF EXISTS idx_context_profiles_embedding;
		CREATE INDEX IF NOT EXISTS idx_context_profiles_embedding ON context_profiles USING hnsw (embedding vector_cosine_ops);
	END IF;

	IF to_regclass('public.application_profiles') IS NOT NULL THEN
		ALTER TABLE application_profiles DROP COLUMN IF EXISTS embedding;
		ALTER TABLE application_profiles ADD COLUMN IF NOT EXISTS embedding vector(768);
		DROP INDEX IF EXISTS idx_app_profiles_embedding;
		CREATE INDEX IF NOT EXISTS idx_app_profiles_embedding ON application_profiles USING hnsw (embedding vector_cosine_ops);
	END IF;

	IF to_regclass('public.application_components') IS NOT NULL THEN
		ALTER TABLE application_components DROP COLUMN IF EXISTS embedding;
		ALTER TABLE application_components ADD COLUMN IF NOT EXISTS embedding vector(768);
		DROP INDEX IF EXISTS idx_app_components_embedding;
		CREATE INDEX IF NOT EXISTS idx_app_components_embedding ON application_components USING hnsw (embedding vector_cosine_ops);
	END IF;

	IF to_regclass('public.application_api_endpoints') IS NOT NULL THEN
		ALTER TABLE application_api_endpoints DROP COLUMN IF EXISTS embedding;
		ALTER TABLE application_api_endpoints ADD COLUMN IF NOT EXISTS embedding vector(768);
		DROP INDEX IF EXISTS idx_app_endpoints_embedding;
		CREATE INDEX IF NOT EXISTS idx_app_endpoints_embedding ON application_api_endpoints USING hnsw (embedding vector_cosine_ops);
	END IF;

	IF to_regclass('public.code_patterns') IS NOT NULL THEN
		ALTER TABLE code_patterns DROP COLUMN IF EXISTS embedding;
		ALTER TABLE code_patterns ADD COLUMN IF NOT EXISTS embedding vector(768);
		DROP INDEX IF EXISTS idx_code_patterns_embedding;
		CREATE INDEX IF NOT EXISTS idx_code_patterns_embedding ON code_patterns USING hnsw (embedding vector_cosine_ops);
	END IF;
END $$;
