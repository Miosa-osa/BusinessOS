-- Agent Experience Store for Learning from Past Outcomes
-- Enables agents to avoid repeating failures and build on successes

CREATE TABLE IF NOT EXISTS agent_experience (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    task_type TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    outcome TEXT NOT NULL,  -- success, failure, timeout
    learned_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    UNIQUE(agent_id, task_type, input_hash)
);

-- Index for fast lookup during agent decision-making
CREATE INDEX IF NOT EXISTS idx_agent_experience_lookup
    ON agent_experience(agent_id, task_type, input_hash);

-- Index for temporal analysis of agent learning
CREATE INDEX IF NOT EXISTS idx_agent_experience_learned_at
    ON agent_experience(agent_id, learned_at DESC);

-- Index for querying by outcome type
CREATE INDEX IF NOT EXISTS idx_agent_experience_outcome
    ON agent_experience(agent_id, outcome);

-- Add comments for documentation
COMMENT ON TABLE agent_experience IS 'Stores agent learning outcomes to avoid repeating mistakes';
COMMENT ON COLUMN agent_experience.input_hash IS 'Hash of input parameters for deduplication';
COMMENT ON COLUMN agent_experience.outcome IS 'Result: success, failure, or timeout';
COMMENT ON COLUMN agent_experience.metadata IS 'Additional context (error messages, retry counts, etc.)';
