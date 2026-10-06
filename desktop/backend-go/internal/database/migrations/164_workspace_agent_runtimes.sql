ALTER TABLE workspace_agents
    ADD COLUMN IF NOT EXISTS runtime TEXT NOT NULL DEFAULT 'osa';

ALTER TABLE workspace_agent_runs
    ADD COLUMN IF NOT EXISTS runtime TEXT NOT NULL DEFAULT 'osa';

ALTER TABLE workspace_agents
    DROP CONSTRAINT IF EXISTS workspace_agents_runtime_check;

ALTER TABLE workspace_agents
    ADD CONSTRAINT workspace_agents_runtime_check
    CHECK (runtime IN ('osa', 'claude-code', 'codex', 'hermes'));
