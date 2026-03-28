-- Migration 106: L0 Process Events for Board Intelligence
-- Creates the three tables that boardchair_l0_sync.go reads to build L0 RDF triples.
-- WvdA: L0 = ground truth event log. All SPARQL L1-L3 materialization derives from these.

-- Active process cases (90-day rolling window used by boardchair_l0_sync)
CREATE TABLE IF NOT EXISTS process_cases (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id      UUID NOT NULL,
    case_ref          VARCHAR(255) NOT NULL,
    department        VARCHAR(255) NOT NULL,
    status            VARCHAR(50)  NOT NULL DEFAULT 'active'
                          CHECK (status IN ('active', 'completed', 'suspended', 'archived')),
    cycle_time_seconds INT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_process_cases_workspace  ON process_cases (workspace_id);
CREATE INDEX IF NOT EXISTS idx_process_cases_dept       ON process_cases (department);
CREATE INDEX IF NOT EXISTS idx_process_cases_status     ON process_cases (status);
CREATE INDEX IF NOT EXISTS idx_process_cases_created_at ON process_cases (created_at);

-- Handoffs between departments (process flow edges)
CREATE TABLE IF NOT EXISTS process_handoffs (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id      UUID        NOT NULL,
    source_department VARCHAR(255) NOT NULL,
    target_department VARCHAR(255) NOT NULL,
    duration          INTERVAL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_process_handoffs_workspace ON process_handoffs (workspace_id);
CREATE INDEX IF NOT EXISTS idx_process_handoffs_source    ON process_handoffs (source_department);
CREATE INDEX IF NOT EXISTS idx_process_handoffs_created   ON process_handoffs (created_at);

-- Discovery results from pm4py-rust / BusinessOS BOS engine
-- One row per (workspace_id, model_id) — updated after each conformance run.
CREATE TABLE IF NOT EXISTS process_discovery_results (
    id                      UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id            UUID          NOT NULL,
    model_id                VARCHAR(255)  NOT NULL,
    department              VARCHAR(255),
    algorithm               VARCHAR(100),
    -- fitness -1.0 = not-yet-computed sentinel; [0.0, 1.0] = actual fitness
    fitness                 DECIMAL(5, 4) NOT NULL DEFAULT -1.0
                                CHECK (fitness >= -1.0 AND fitness <= 1.0),
    avg_cycle_time_hours    DECIMAL(10, 4),
    bottleneck_activity     VARCHAR(255),
    activities_count        INT,
    traces_count            INT,
    raw_result              JSONB         NOT NULL DEFAULT '{}',
    discovered_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (workspace_id, model_id)
);

CREATE INDEX IF NOT EXISTS idx_discovery_workspace   ON process_discovery_results (workspace_id);
CREATE INDEX IF NOT EXISTS idx_discovery_dept        ON process_discovery_results (department);
CREATE INDEX IF NOT EXISTS idx_discovery_discovered  ON process_discovery_results (discovered_at);
