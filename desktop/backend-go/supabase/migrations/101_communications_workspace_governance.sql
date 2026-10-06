-- Communications are account-owned source records with one explicit workspace route.
-- Provider data remains immutable in its connector tables. This overlay stores the
-- operational state without duplicating sensitive messages across workspaces.

CREATE TABLE IF NOT EXISTS communication_workspace_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    owner_user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    source_kind VARCHAR(50) NOT NULL,
    source_external_id TEXT NOT NULL,
    state VARCHAR(32) NOT NULL DEFAULT 'triage',
    assigned_to VARCHAR(255),
    waiting_on VARCHAR(255),
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    snoozed_until TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT communication_workspace_items_state_check CHECK (
        state IN ('triage', 'waiting', 'delegated', 'scheduled', 'resolved', 'archived')
    ),
    -- A source is deliberately routed once. Moving it changes this workspace_id.
    CONSTRAINT communication_workspace_items_source_owner_unique UNIQUE (
        owner_user_id, provider, source_kind, source_external_id
    )
);

CREATE INDEX IF NOT EXISTS idx_communication_workspace_items_workspace_state
    ON communication_workspace_items (workspace_id, state, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_communication_workspace_items_owner_source
    ON communication_workspace_items (owner_user_id, provider, source_kind);

CREATE TABLE IF NOT EXISTS communication_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_item_id UUID REFERENCES communication_workspace_items(id) ON DELETE SET NULL,
    created_by VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    recipient JSONB NOT NULL DEFAULT '{}',
    subject TEXT,
    body TEXT NOT NULL DEFAULT '',
    state VARCHAR(32) NOT NULL DEFAULT 'draft',
    approved_by VARCHAR(255),
    approved_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT communication_drafts_state_check CHECK (
        state IN ('draft', 'pending_approval', 'approved', 'sent', 'cancelled')
    )
);

CREATE INDEX IF NOT EXISTS idx_communication_drafts_workspace_state
    ON communication_drafts (workspace_id, state, updated_at DESC);

COMMENT ON TABLE communication_workspace_items IS
    'Workspace routing and operating state over immutable communication connector records.';

COMMENT ON TABLE communication_drafts IS
    'Outbound communication drafts. Sending remains a separate explicit approval action.';
