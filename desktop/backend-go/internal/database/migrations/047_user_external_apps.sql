-- Migration 047: User External Apps
-- Allows users to add external web applications (Notion, Slack, ClickUp, Linear, etc.)
-- to their 3D Desktop as iframe windows

-- Create user_external_apps table
CREATE TABLE IF NOT EXISTS user_external_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- App Identity
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,

    -- Visual Metadata (matches MODULE_INFO pattern from frontend)
    icon VARCHAR(100) NOT NULL DEFAULT 'app-window',  -- Lucide icon name (e.g., 'file-text', 'message-square')
    color VARCHAR(7) NOT NULL DEFAULT '#6366F1',      -- Hex color code (e.g., '#000000' for Notion)

    -- Categorization
    category VARCHAR(100) DEFAULT 'productivity',     -- e.g., 'productivity', 'communication', 'design'
    description TEXT,                                 -- Optional description

    -- Desktop Positioning (stored for persistence across sessions)
    position_x INTEGER DEFAULT 0,
    position_y INTEGER DEFAULT 0,
    position_z INTEGER DEFAULT 0,

    -- Iframe Configuration (JSON for flexibility)
    iframe_config JSONB DEFAULT '{"sandbox": ["allow-same-origin", "allow-scripts", "allow-popups", "allow-forms"], "allowFullscreen": true}'::jsonb,

    -- State Management
    is_active BOOLEAN DEFAULT true,                   -- Can be disabled without deletion
    open_on_startup BOOLEAN DEFAULT false,            -- Auto-open when desktop loads

    -- App Type (for future native app support)
    app_type VARCHAR(50) DEFAULT 'web' NOT NULL,      -- 'web' (iframe), 'native' (future)

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_opened_at TIMESTAMPTZ                        -- Track usage
);

-- Indexes for performance
CREATE INDEX idx_user_external_apps_user ON user_external_apps(user_id);
CREATE INDEX idx_user_external_apps_workspace ON user_external_apps(workspace_id);
CREATE INDEX idx_user_external_apps_active ON user_external_apps(workspace_id, is_active) WHERE is_active = true;

-- Unique constraint: one app name per workspace
ALTER TABLE user_external_apps
ADD CONSTRAINT user_external_apps_name_workspace_unique
UNIQUE(workspace_id, name);

-- Comments for documentation
COMMENT ON TABLE user_external_apps IS 'External web applications added by users to their 3D Desktop';
COMMENT ON COLUMN user_external_apps.icon IS 'Lucide icon name (e.g., file-text, message-square, check-square)';
COMMENT ON COLUMN user_external_apps.color IS 'Hex color code for app branding (e.g., #000000 for Notion black)';
COMMENT ON COLUMN user_external_apps.iframe_config IS 'JSON config for iframe sandbox attributes and permissions';
COMMENT ON COLUMN user_external_apps.app_type IS 'web = iframe embed, native = future desktop app capture';
