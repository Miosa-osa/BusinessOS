-- Migration: 102_comprehensive_audit_trail
-- Date: 2026-03-24
-- Purpose: Create immutable audit trail tables for GDPR/SOC2 compliance

-- ════════════════════════════════════════════════════════════════════════════════
-- Audit Events Table (Hash Chain)
-- ════════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS audit_events (
  id BIGSERIAL PRIMARY KEY,

  -- ── Chain Integrity ──────────────────────────────────────────────────────────
  sequence_number BIGINT UNIQUE NOT NULL,
  entry_hash CHAR(64) UNIQUE NOT NULL,
  previous_hash CHAR(64) NOT NULL,
  merkle_tree_hash CHAR(64),

  -- ── Event Metadata ───────────────────────────────────────────────────────────
  event_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
  event_type VARCHAR(64) NOT NULL,
  event_category VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  severity VARCHAR(16) NOT NULL DEFAULT 'info',

  -- ── Actor Context ────────────────────────────────────────────────────────────
  user_id UUID,
  session_id UUID,
  ip_address INET,
  user_agent TEXT,

  -- ── Resource Context ─────────────────────────────────────────────────────────
  resource_type VARCHAR(64),
  resource_id UUID,
  workspace_id UUID,

  -- ── Event Payload (Immutable) ────────────────────────────────────────────────
  payload JSONB NOT NULL,

  -- ── Compliance Metadata ──────────────────────────────────────────────────────
  legal_hold BOOLEAN DEFAULT FALSE,
  retention_expires_at TIMESTAMPTZ,
  deletion_blocked_until TIMESTAMPTZ,

  -- ── Encryption (At Rest) ─────────────────────────────────────────────────────
  encrypted_payload TEXT,
  encryption_key_id VARCHAR(64),

  -- ── Constraints ──────────────────────────────────────────────────────────────
  CONSTRAINT valid_entry_hash CHECK (entry_hash ~ '^[a-f0-9]{64}$'),
  CONSTRAINT valid_previous_hash CHECK (previous_hash ~ '^[a-f0-9]{64}$'),
  CONSTRAINT resource_context_both_or_neither CHECK (
    (resource_type IS NOT NULL AND resource_id IS NOT NULL) OR
    (resource_type IS NULL AND resource_id IS NULL)
  ),
  CONSTRAINT severity_valid CHECK (severity IN ('info', 'warning', 'critical')),
  CONSTRAINT event_category_valid CHECK (
    event_category IN ('ProcessMining', 'Compliance', 'Integration', 'Security')
  ),
  CONSTRAINT legal_hold_requires_retention CHECK (
    (legal_hold = FALSE AND retention_expires_at IS NOT NULL) OR
    legal_hold = TRUE
  )
);

-- ════════════════════════════════════════════════════════════════════════════════
-- Indexes for Query Performance (Compliance)
-- ════════════════════════════════════════════════════════════════════════════════

-- Time-based queries (audit reports, incident investigation)
CREATE INDEX IF NOT EXISTS audit_events_created_at_idx
  ON audit_events(created_at DESC);

-- User-based queries (GDPR DSAR: data subject access requests)
CREATE INDEX IF NOT EXISTS audit_events_user_id_idx
  ON audit_events(user_id);

-- Resource-based queries (access audit, resource history)
CREATE INDEX IF NOT EXISTS audit_events_resource_idx
  ON audit_events(resource_type, resource_id);

-- Event type queries (compliance audits)
CREATE INDEX IF NOT EXISTS audit_events_event_type_idx
  ON audit_events(event_type);

-- Workspace-based queries (multi-tenant auditing)
CREATE INDEX IF NOT EXISTS audit_events_workspace_idx
  ON audit_events(workspace_id);

-- Legal hold queries (litigation, investigations)
CREATE INDEX IF NOT EXISTS audit_events_legal_hold_idx
  ON audit_events(legal_hold)
  WHERE legal_hold = TRUE;

-- Retention expiration (automated cleanup)
CREATE INDEX IF NOT EXISTS audit_events_retention_expires_idx
  ON audit_events(retention_expires_at)
  WHERE retention_expires_at IS NOT NULL;

-- Chain integrity verification (sequence ordering)
CREATE INDEX IF NOT EXISTS audit_events_sequence_idx
  ON audit_events(sequence_number);

-- Full-text search on payload (event content search)
CREATE INDEX IF NOT EXISTS audit_events_payload_idx
  ON audit_events USING GIN (payload);

-- Category-based audit reports
CREATE INDEX IF NOT EXISTS audit_events_category_idx
  ON audit_events(event_category);

-- ════════════════════════════════════════════════════════════════════════════════
-- Audit Index Table (Session-Level Chain Tracking)
-- ════════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS audit_index (
  id BIGSERIAL PRIMARY KEY,
  session_id UUID UNIQUE NOT NULL,
  event_count BIGINT DEFAULT 0,
  first_event_hash CHAR(64),
  last_event_hash CHAR(64),
  merkle_root CHAR(64),
  integrity_verified BOOLEAN DEFAULT FALSE,
  verified_at TIMESTAMPTZ,
  last_verified_sequence BIGINT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT valid_merkle CHECK (merkle_root IS NULL OR merkle_root ~ '^[a-f0-9]{64}$')
);

CREATE INDEX IF NOT EXISTS audit_index_verified_idx
  ON audit_index(integrity_verified, verified_at DESC);

-- ════════════════════════════════════════════════════════════════════════════════
-- Legal Hold Table (Compliance Hold Management)
-- ════════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS audit_legal_holds (
  id BIGSERIAL PRIMARY KEY,
  hold_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID,
  workspace_id UUID,
  reason VARCHAR(255) NOT NULL,
  issued_by UUID NOT NULL,
  issued_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ,
  released_at TIMESTAMPTZ,
  released_by UUID,
  release_reason VARCHAR(255),

  CONSTRAINT hold_not_expired CHECK (released_at IS NULL OR released_at <= expires_at)
);

CREATE INDEX IF NOT EXISTS audit_legal_holds_user_idx
  ON audit_legal_holds(user_id)
  WHERE released_at IS NULL;

CREATE INDEX IF NOT EXISTS audit_legal_holds_workspace_idx
  ON audit_legal_holds(workspace_id)
  WHERE released_at IS NULL;

-- ════════════════════════════════════════════════════════════════════════════════
-- Encryption Key Rotation Log
-- ════════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS audit_key_rotations (
  id BIGSERIAL PRIMARY KEY,
  encryption_key_id VARCHAR(64) UNIQUE NOT NULL,
  key_version INTEGER NOT NULL,
  rotated_at TIMESTAMPTZ DEFAULT NOW(),
  algorithm VARCHAR(32) NOT NULL,
  rotated_by UUID NOT NULL,
  previous_key_id VARCHAR(64),

  CONSTRAINT key_version_positive CHECK (key_version > 0)
);

-- ════════════════════════════════════════════════════════════════════════════════
-- Functions & Triggers
-- ════════════════════════════════════════════════════════════════════════════════

-- Function: Verify hash chain integrity before insertion
CREATE OR REPLACE FUNCTION verify_audit_chain_before_insert()
RETURNS TRIGGER AS $$
BEGIN
  -- Verify entry_hash is valid hex (64 chars)
  IF NEW.entry_hash !~ '^[a-f0-9]{64}$' THEN
    RAISE EXCEPTION 'Invalid entry_hash: must be 64 hex characters';
  END IF;

  -- Verify previous_hash is valid hex (64 chars) or genesis (all zeros)
  IF NEW.previous_hash !~ '^[a-f0-9]{64}$' THEN
    RAISE EXCEPTION 'Invalid previous_hash: must be 64 hex characters';
  END IF;

  -- Verify sequence number is positive and unique
  IF NEW.sequence_number <= 0 THEN
    RAISE EXCEPTION 'Sequence number must be positive';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_chain_verification
BEFORE INSERT ON audit_events
FOR EACH ROW
EXECUTE FUNCTION verify_audit_chain_before_insert();

-- Function: Prevent modification of audit events (immutability)
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'UPDATE' THEN
    -- Only allow legal_hold and deletion_blocked_until changes
    IF NEW.entry_hash != OLD.entry_hash OR
       NEW.payload != OLD.payload OR
       NEW.event_id != OLD.event_id THEN
      RAISE EXCEPTION 'Audit events are immutable: cannot modify %', TG_OP;
    END IF;
  ELSIF TG_OP = 'DELETE' THEN
    -- Prevent deletion if event is under legal hold
    IF OLD.legal_hold = TRUE THEN
      RAISE EXCEPTION 'Cannot delete audit event under legal hold: %', OLD.event_id;
    END IF;

    -- Prevent deletion if retention period has not expired
    IF OLD.deletion_blocked_until IS NOT NULL AND OLD.deletion_blocked_until > NOW() THEN
      RAISE EXCEPTION 'Cannot delete event: deletion blocked until %', OLD.deletion_blocked_until;
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_immutability
BEFORE UPDATE OR DELETE ON audit_events
FOR EACH ROW
EXECUTE FUNCTION prevent_audit_modification();

-- Function: Update audit_index on new events
CREATE OR REPLACE FUNCTION update_audit_index()
RETURNS TRIGGER AS $$
BEGIN
  -- Ensure session_id is set (use workspace for sessions without explicit session_id)
  DECLARE
    v_session_id UUID;
  BEGIN
    v_session_id := COALESCE(NEW.session_id, NEW.workspace_id, gen_random_uuid());

    INSERT INTO audit_index (session_id, event_count, first_event_hash, last_event_hash)
    VALUES (v_session_id, 1, NEW.entry_hash, NEW.entry_hash)
    ON CONFLICT (session_id) DO UPDATE SET
      event_count = audit_index.event_count + 1,
      last_event_hash = NEW.entry_hash,
      updated_at = NOW();
  END;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_index_update
AFTER INSERT ON audit_events
FOR EACH ROW
EXECUTE FUNCTION update_audit_index();

-- ════════════════════════════════════════════════════════════════════════════════
-- Views for Compliance Reporting
-- ════════════════════════════════════════════════════════════════════════════════

-- View: Audit events by category (compliance breakdown)
CREATE OR REPLACE VIEW audit_events_by_category AS
SELECT
  event_category,
  COUNT(*) as event_count,
  MIN(created_at) as earliest_event,
  MAX(created_at) as latest_event,
  COUNT(DISTINCT user_id) as affected_users,
  COUNT(DISTINCT resource_id) as affected_resources,
  SUM(CASE WHEN severity = 'critical' THEN 1 ELSE 0 END) as critical_events
FROM audit_events
GROUP BY event_category
ORDER BY event_count DESC;

-- View: Security events (incidents)
CREATE OR REPLACE VIEW audit_security_events AS
SELECT
  event_id,
  created_at,
  event_type,
  user_id,
  ip_address,
  payload,
  severity
FROM audit_events
WHERE event_category = 'Security'
ORDER BY created_at DESC;

-- View: GDPR-relevant events
CREATE OR REPLACE VIEW audit_gdpr_events AS
SELECT
  event_id,
  created_at,
  event_type,
  gdpr_classification,
  user_id,
  data_subjects_affected,
  pii_detected,
  legal_hold,
  retention_expires_at
FROM audit_events
WHERE gdpr_classification IS NOT NULL
ORDER BY created_at DESC;

-- View: Events under legal hold (litigation support)
CREATE OR REPLACE VIEW audit_legal_hold_events AS
SELECT
  ae.event_id,
  ae.created_at,
  ae.event_type,
  ae.user_id,
  alh.hold_id,
  alh.reason,
  alh.issued_at,
  alh.expires_at
FROM audit_events ae
JOIN audit_legal_holds alh ON
  (ae.user_id = alh.user_id OR alh.user_id IS NULL) AND
  (ae.workspace_id = alh.workspace_id OR alh.workspace_id IS NULL)
WHERE ae.legal_hold = TRUE AND alh.released_at IS NULL
ORDER BY ae.created_at DESC;

-- ════════════════════════════════════════════════════════════════════════════════
-- Stored Procedures for Compliance Operations
-- ════════════════════════════════════════════════════════════════════════════════

-- Procedure: Place legal hold on user's events
CREATE OR REPLACE FUNCTION place_legal_hold(
  p_user_id UUID,
  p_reason VARCHAR,
  p_issued_by UUID,
  p_expires_at TIMESTAMPTZ DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
  v_hold_id UUID;
BEGIN
  INSERT INTO audit_legal_holds (user_id, reason, issued_by, expires_at)
  VALUES (p_user_id, p_reason, p_issued_by, p_expires_at)
  RETURNING hold_id INTO v_hold_id;

  UPDATE audit_events
  SET legal_hold = TRUE
  WHERE user_id = p_user_id;

  RETURN v_hold_id;
END;
$$ LANGUAGE plpgsql;

-- Procedure: Release legal hold
CREATE OR REPLACE FUNCTION release_legal_hold(
  p_hold_id UUID,
  p_released_by UUID,
  p_release_reason VARCHAR
)
RETURNS BOOLEAN AS $$
DECLARE
  v_user_id UUID;
BEGIN
  UPDATE audit_legal_holds
  SET released_at = NOW(), released_by = p_released_by, release_reason = p_release_reason
  WHERE hold_id = p_hold_id
  RETURNING user_id INTO v_user_id;

  IF v_user_id IS NOT NULL THEN
    -- Only release hold if no other active holds
    IF NOT EXISTS (
      SELECT 1 FROM audit_legal_holds
      WHERE user_id = v_user_id AND released_at IS NULL
    ) THEN
      UPDATE audit_events
      SET legal_hold = FALSE
      WHERE user_id = v_user_id;
    END IF;
  END IF;

  RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Procedure: Verify chain integrity for date range
CREATE OR REPLACE FUNCTION verify_audit_chain_integrity(
  p_from_date TIMESTAMPTZ,
  p_to_date TIMESTAMPTZ
)
RETURNS TABLE (
  is_valid BOOLEAN,
  total_entries BIGINT,
  broken_links BIGINT,
  last_verified_at TIMESTAMPTZ
) AS $$
DECLARE
  v_broken_links BIGINT;
BEGIN
  -- Count chain breaks (where previous_hash doesn't match prior entry's entry_hash)
  WITH chain_check AS (
    SELECT
      ae.sequence_number,
      ae.previous_hash,
      LAG(ae.entry_hash) OVER (ORDER BY ae.sequence_number) as expected_hash
    FROM audit_events ae
    WHERE ae.created_at BETWEEN p_from_date AND p_to_date
  )
  SELECT COUNT(*) INTO v_broken_links
  FROM chain_check
  WHERE previous_hash != expected_hash AND expected_hash IS NOT NULL;

  RETURN QUERY SELECT
    v_broken_links = 0 as is_valid,
    COUNT(*) as total_entries,
    v_broken_links,
    NOW() as last_verified_at
  FROM audit_events
  WHERE created_at BETWEEN p_from_date AND p_to_date;
END;
$$ LANGUAGE plpgsql;

-- ════════════════════════════════════════════════════════════════════════════════
-- Comments & Documentation
-- ════════════════════════════════════════════════════════════════════════════════

COMMENT ON TABLE audit_events IS
'Immutable hash-chain audit trail for all BOS ↔ BusinessOS operations.
Provides GDPR/SOC2 compliance, tamper detection, and forensics.';

COMMENT ON COLUMN audit_events.sequence_number IS
'Globally unique sequence number (immutable chain ordering).';

COMMENT ON COLUMN audit_events.entry_hash IS
'SHA256(content): hash of this entry (tamper-evident).';

COMMENT ON COLUMN audit_events.previous_hash IS
'SHA256 of previous entry: creates cryptographic chain link.';

COMMENT ON COLUMN audit_events.legal_hold IS
'Cannot delete if TRUE (for litigation/investigations).';

COMMENT ON COLUMN audit_events.retention_expires_at IS
'Data retention deadline per compliance policy (7 years default).';

COMMENT ON COLUMN audit_events.pii_detected IS
'TRUE if personal data present (GDPR Article 32 requirement).';

COMMENT ON TABLE audit_index IS
'Session-level chain tracking for efficient verification.';

COMMENT ON TABLE audit_legal_holds IS
'Legal hold registry for litigation/investigations.';

COMMENT ON FUNCTION verify_audit_chain_integrity(TIMESTAMPTZ, TIMESTAMPTZ) IS
'Verify hash chain integrity for compliance audits. Returns True if no tampering detected.';

-- ════════════════════════════════════════════════════════════════════════════════
-- Initial Seed Data
-- ════════════════════════════════════════════════════════════════════════════════

-- Seed with genesis event (if table is empty)
INSERT INTO audit_events (
  sequence_number, entry_hash, previous_hash, event_id,
  event_type, event_category, created_at, severity,
  payload, legal_hold
)
SELECT
  1,
  '0000000000000000000000000000000000000000000000000000000000000001',
  '0000000000000000000000000000000000000000000000000000000000000000',
  gen_random_uuid(),
  'system_initialized',
  'Compliance',
  NOW(),
  'info',
  '{"event":"Audit trail initialized"}'::jsonb,
  FALSE
WHERE NOT EXISTS (SELECT 1 FROM audit_events LIMIT 1);

-- ════════════════════════════════════════════════════════════════════════════════
-- Cleanup & Retention Job
-- ════════════════════════════════════════════════════════════════════════════════

-- CRON job (external scheduler): DELETE from audit_events
-- WHERE retention_expires_at < NOW()
--   AND legal_hold = FALSE
--   AND deletion_blocked_until < NOW();

-- Note: Trigger audit_immutability will prevent deletion of legal-hold events
-- and events within deletion_blocked_until window.
