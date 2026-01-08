-- Migration: Comments and Mentions System
-- Supports comments on tasks, projects, and other entities with @user mentions

-- ===== COMMENTS TABLE =====
-- Generic comments that can be attached to any entity (tasks, projects, etc.)
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Who wrote the comment
    user_id VARCHAR(255) NOT NULL,
    
    -- What entity this comment is on (polymorphic)
    entity_type VARCHAR(50) NOT NULL,  -- 'task', 'project', 'client', 'note'
    entity_id UUID NOT NULL,
    
    -- Comment content
    content TEXT NOT NULL,
    
    -- Optional reply threading
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    
    -- Edit tracking
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMPTZ,
    
    -- Soft delete
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_comments_entity ON comments(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_comments_created ON comments(entity_type, entity_id, created_at DESC);


-- ===== ENTITY MENTIONS TABLE =====
-- Tracks @user mentions within comments (separate from agent mentions)
CREATE TABLE IF NOT EXISTS entity_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Source: where the mention occurred
    source_type VARCHAR(50) NOT NULL,  -- 'comment', 'task_description', 'project_description'
    source_id UUID NOT NULL,           -- ID of comment, task, or project
    
    -- Target: who was mentioned
    mentioned_user_id VARCHAR(255) NOT NULL,  -- The user being @mentioned
    
    -- Context
    mention_text VARCHAR(100) NOT NULL,  -- The actual @mention text (e.g., "@john")
    position_in_text INT,                -- Character position in the source text
    
    -- The parent entity (for quick lookups)
    entity_type VARCHAR(50),  -- 'task', 'project', etc.
    entity_id UUID,           -- The task/project ID
    
    -- Who made the mention
    mentioned_by VARCHAR(255) NOT NULL,
    
    -- Was notification sent?
    notified BOOLEAN DEFAULT FALSE,
    notified_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_entity_mentions_source ON entity_mentions(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_entity_mentions_user ON entity_mentions(mentioned_user_id);
CREATE INDEX IF NOT EXISTS idx_entity_mentions_entity ON entity_mentions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_mentions_by ON entity_mentions(mentioned_by);


-- ===== COMMENT REACTIONS TABLE (FUTURE) =====
-- For emoji reactions on comments
CREATE TABLE IF NOT EXISTS comment_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    emoji VARCHAR(20) NOT NULL,  -- Unicode emoji or shortcode
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(comment_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment ON comment_reactions(comment_id);


-- Comments
COMMENT ON TABLE comments IS 'Generic comments attachable to any entity (task, project, etc.)';
COMMENT ON TABLE entity_mentions IS 'Tracks @user mentions within comments and descriptions';
COMMENT ON TABLE comment_reactions IS 'Emoji reactions on comments';
COMMENT ON COLUMN comments.entity_type IS 'Type of entity: task, project, client, note';
COMMENT ON COLUMN entity_mentions.source_type IS 'Where mention occurred: comment, task_description, etc.';
