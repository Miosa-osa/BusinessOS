-- Migration: 102_process_model_versioning.sql
-- Description: Process Model Versioning System
-- Created: 2026-03-24
-- Phase: 4 (Analytics & Discovery)

-- =============================================================================
-- PROCESS MODELS
-- Stores discovered process model definitions
-- =============================================================================
CREATE TABLE IF NOT EXISTS process_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership
    workspace_id UUID NOT NULL REFERENCES osa_workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Model identification
    name VARCHAR(255) NOT NULL,
    description TEXT,
    model_type VARCHAR(100), -- 'bpmn', 'petri_net', 'workflow', 'process_tree'

    -- Current version reference
    current_version_id UUID REFERENCES process_model_versions(id),

    -- Workflow context
    workflow_id UUID REFERENCES osa_workflows(id) ON DELETE SET NULL,
    discovery_method VARCHAR(100), -- 'inductive', 'heuristic', 'alpha', 'user_edit'

    -- Status
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'validated', 'released', 'archived'
    is_active BOOLEAN DEFAULT true,

    -- Metadata
    tags TEXT[],
    metadata JSONB DEFAULT '{}',

    -- Statistics
    total_versions INT DEFAULT 0,
    released_versions INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    archived_at TIMESTAMPTZ
);

CREATE INDEX idx_process_models_workspace ON process_models(workspace_id);
CREATE INDEX idx_process_models_user ON process_models(user_id);
CREATE INDEX idx_process_models_status ON process_models(status);
CREATE INDEX idx_process_models_active ON process_models(is_active) WHERE is_active;
CREATE INDEX idx_process_models_created ON process_models(created_at DESC);
CREATE INDEX idx_process_models_tags ON process_models USING GIN(tags);

-- =============================================================================
-- PROCESS MODEL VERSIONS
-- Immutable version history for each process model
-- =============================================================================
CREATE TABLE IF NOT EXISTS process_model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Model identification
    model_id UUID NOT NULL REFERENCES process_models(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL, -- e.g., "2.1.3+a7c3e9f1"

    -- Semantic versioning components
    major INT NOT NULL,
    minor INT NOT NULL,
    patch INT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,

    -- Creation info
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    discovery_source VARCHAR(100), -- 'inductive', 'heuristic', 'alpha', 'user_edit'

    -- Relationships
    previous_version_id UUID REFERENCES process_model_versions(id),

    -- Content storage (compressed)
    model_json BYTEA NOT NULL,
    delta_json BYTEA,

    -- Metrics snapshot
    nodes_count INT,
    edges_count INT,
    variants INT,
    fitness DECIMAL(5,4),
    average_duration DECIMAL(10,2),
    covered_traces INT,

    -- Change tracking
    change_type VARCHAR(50) NOT NULL, -- 'major', 'minor', 'patch'
    nodes_added INT DEFAULT 0,
    nodes_removed INT DEFAULT 0,
    edges_added INT DEFAULT 0,
    edges_removed INT DEFAULT 0,

    -- Metadata
    description TEXT,
    tags TEXT[],

    -- Release status
    is_released BOOLEAN DEFAULT FALSE,
    release_notes TEXT,
    released_at TIMESTAMPTZ,

    -- Archival
    archived_at TIMESTAMPTZ,

    -- Constraints
    UNIQUE(model_id, version),
    UNIQUE(model_id, major, minor, patch),
    CONSTRAINT fitness_range CHECK (fitness >= 0 AND fitness <= 1)
);

CREATE INDEX idx_model_versions_model_id ON process_model_versions(model_id);
CREATE INDEX idx_model_versions_created_at ON process_model_versions(created_at DESC);
CREATE INDEX idx_model_versions_hash ON process_model_versions(content_hash);
CREATE INDEX idx_model_versions_prev ON process_model_versions(previous_version_id);
CREATE INDEX idx_model_versions_released ON process_model_versions(is_released, released_at DESC) WHERE is_released;
CREATE INDEX idx_model_versions_change_type ON process_model_versions(change_type);
CREATE INDEX idx_model_versions_version_components ON process_model_versions(model_id, major, minor, patch);

-- =============================================================================
-- VERSION COMPARISON CACHE
-- Caches computed diffs to avoid repeated computation
-- =============================================================================
CREATE TABLE IF NOT EXISTS model_version_diffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    model_id UUID NOT NULL REFERENCES process_models(id) ON DELETE CASCADE,
    from_version_id UUID NOT NULL REFERENCES process_model_versions(id) ON DELETE CASCADE,
    to_version_id UUID NOT NULL REFERENCES process_model_versions(id) ON DELETE CASCADE,

    -- Cached diff
    structural_diff JSONB NOT NULL,
    metrics_diff JSONB NOT NULL,
    breaking_changes TEXT[],

    -- Metadata
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    computed_by VARCHAR(255),

    UNIQUE(from_version_id, to_version_id)
);

CREATE INDEX idx_model_diffs_model ON model_version_diffs(model_id);
CREATE INDEX idx_model_diffs_from_to ON model_version_diffs(from_version_id, to_version_id);

-- =============================================================================
-- ROLLBACK AUDIT TRAIL
-- Immutable record of all rollback operations
-- =============================================================================
CREATE TABLE IF NOT EXISTS model_version_rollback_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    model_id UUID NOT NULL REFERENCES process_models(id) ON DELETE CASCADE,
    from_version VARCHAR(20) NOT NULL,
    to_version VARCHAR(20) NOT NULL,

    -- Rollback details
    reason TEXT NOT NULL,
    approved_by VARCHAR(255) NOT NULL,
    performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Impact tracking
    instances_affected INT,
    instances_paused INT,
    instances_replayed INT,

    -- Status
    status VARCHAR(50) DEFAULT 'completed', -- 'pending', 'completed', 'failed'
    error_message TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_rollback_model ON model_version_rollback_audits(model_id);
CREATE INDEX idx_rollback_performed ON model_version_rollback_audits(performed_at DESC);
CREATE INDEX idx_rollback_approved ON model_version_rollback_audits(approved_by);

-- =============================================================================
-- VERSION TAGS & LABELS
-- User-friendly tagging for versions
-- =============================================================================
CREATE TABLE IF NOT EXISTS model_version_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    version_id UUID NOT NULL REFERENCES process_model_versions(id) ON DELETE CASCADE,
    tag_name VARCHAR(100) NOT NULL,
    tag_value VARCHAR(255),

    -- Metadata
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(version_id, tag_name)
);

CREATE INDEX idx_version_tags_version ON model_version_tags(version_id);
CREATE INDEX idx_version_tags_tag ON model_version_tags(tag_name, tag_value);

-- =============================================================================
-- VERSION RETENTION POLICY TRACKING
-- Tracks which versions should be retained/archived
-- =============================================================================
CREATE TABLE IF NOT EXISTS model_version_retention (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    version_id UUID NOT NULL REFERENCES process_model_versions(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES process_models(id) ON DELETE CASCADE,

    -- Retention decision
    should_retain BOOLEAN NOT NULL,
    retention_reason VARCHAR(255),
    expires_at TIMESTAMPTZ,

    -- Archival tracking
    is_archived BOOLEAN DEFAULT FALSE,
    archived_at TIMESTAMPTZ,
    archived_location VARCHAR(1024), -- S3 path, GCS, etc.

    -- Backup status
    is_backed_up BOOLEAN DEFAULT FALSE,
    backup_location VARCHAR(1024),
    backup_created_at TIMESTAMPTZ,

    -- Metadata
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    policy_version VARCHAR(20)
);

CREATE INDEX idx_retention_version ON model_version_retention(version_id);
CREATE INDEX idx_retention_model ON model_version_retention(model_id);
CREATE INDEX idx_retention_expires ON model_version_retention(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_retention_archived ON model_version_retention(is_archived, archived_at DESC);

-- =============================================================================
-- CONSTRAINTS & TRIGGERS
-- =============================================================================

-- Trigger: Update process_models.updated_at when new version created
CREATE OR REPLACE FUNCTION update_model_on_version_create()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE process_models
    SET updated_at = NOW(),
        total_versions = total_versions + 1,
        current_version_id = NEW.id,
        released_versions = CASE
            WHEN NEW.is_released THEN released_versions + 1
            ELSE released_versions
        END
    WHERE id = NEW.model_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_model_version_create ON process_model_versions;
CREATE TRIGGER trg_model_version_create
AFTER INSERT ON process_model_versions
FOR EACH ROW
EXECUTE FUNCTION update_model_on_version_create();

-- Trigger: Prevent deletion of released versions
CREATE OR REPLACE FUNCTION prevent_released_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.is_released THEN
        RAISE EXCEPTION 'Cannot delete released version %', OLD.version;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_released_delete ON process_model_versions;
CREATE TRIGGER trg_prevent_released_delete
BEFORE DELETE ON process_model_versions
FOR EACH ROW
EXECUTE FUNCTION prevent_released_delete();

-- =============================================================================
-- VIEWS FOR COMMON QUERIES
-- =============================================================================

-- Latest version of each model
CREATE OR REPLACE VIEW v_latest_model_versions AS
SELECT DISTINCT ON (model_id)
    id, model_id, version, major, minor, patch,
    created_at, created_by, change_type,
    nodes_count, edges_count, variants, fitness,
    is_released, released_at, archived_at
FROM process_model_versions
WHERE archived_at IS NULL
ORDER BY model_id, created_at DESC;

-- Released versions only
CREATE OR REPLACE VIEW v_released_model_versions AS
SELECT
    id, model_id, version, major, minor, patch,
    created_at, created_by, released_at,
    nodes_count, edges_count, variants, fitness,
    release_notes
FROM process_model_versions
WHERE is_released
  AND archived_at IS NULL
ORDER BY model_id, major DESC, minor DESC, patch DESC;

-- Version counts by model
CREATE OR REPLACE VIEW v_model_version_counts AS
SELECT
    model_id,
    COUNT(*) as total_versions,
    COUNT(*) FILTER (WHERE is_released) as released_count,
    COUNT(*) FILTER (WHERE archived_at IS NULL) as active_count,
    MAX(created_at) as last_version_date
FROM process_model_versions
GROUP BY model_id;
