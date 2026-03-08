-- ============================================================================
-- COMPANIES QUERIES
-- ============================================================================

-- name: ListCompanies :many
SELECT * FROM companies
WHERE user_id = $1
  AND (sqlc.narg(industry)::varchar IS NULL OR industry = sqlc.narg(industry))
  AND (sqlc.narg(lifecycle_stage)::varchar IS NULL OR lifecycle_stage = sqlc.narg(lifecycle_stage))
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_val)::int OFFSET sqlc.arg(offset_val)::int;

-- name: GetCompany :one
SELECT * FROM companies
WHERE id = $1 AND user_id = $2;

-- name: GetCompanyByName :one
SELECT * FROM companies
WHERE user_id = $1 AND name ILIKE $2
LIMIT 1;

-- name: CreateCompany :one
INSERT INTO companies (
    user_id, name, legal_name, industry, company_size,
    website, email, phone,
    address_line1, address_line2, city, state, postal_code, country,
    annual_revenue, currency, tax_id,
    linkedin_url, twitter_handle,
    owner_id, lifecycle_stage, lead_source,
    logo_url, custom_fields, metadata
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17,
    $18, $19,
    $20, $21, $22,
    $23, $24, $25
)
RETURNING *;

-- name: UpdateCompany :one
UPDATE companies
SET name = $2, legal_name = $3, industry = $4, company_size = $5,
    website = $6, email = $7, phone = $8,
    address_line1 = $9, address_line2 = $10, city = $11, state = $12, postal_code = $13, country = $14,
    annual_revenue = $15, lifecycle_stage = $16,
    linkedin_url = $17, twitter_handle = $18,
    logo_url = $19, custom_fields = $20, metadata = $21,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCompany :exec
DELETE FROM companies
WHERE id = $1 AND user_id = $2;

-- name: SearchCompanies :many
SELECT * FROM companies
WHERE user_id = $1
  AND (name ILIKE '%' || $2 || '%' OR website ILIKE '%' || $2 || '%')
ORDER BY name ASC
LIMIT sqlc.arg(limit_val)::int;

-- name: UpdateCompanyScores :exec
UPDATE companies
SET health_score = $2, engagement_score = $3, updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- CONTACT-COMPANY RELATIONS QUERIES
-- ============================================================================

-- name: ListCompanyContacts :many
SELECT ccr.*, c.name as contact_name, c.email as contact_email
FROM contact_company_relations ccr
JOIN clients c ON ccr.contact_id = c.id
WHERE ccr.company_id = $1
ORDER BY ccr.is_primary DESC, c.name ASC;

-- name: ListContactCompanies :many
SELECT ccr.*, co.name as company_name
FROM contact_company_relations ccr
JOIN companies co ON ccr.company_id = co.id
WHERE ccr.contact_id = $1
ORDER BY ccr.is_primary DESC, co.name ASC;

-- name: CreateContactCompanyRelation :one
INSERT INTO contact_company_relations (contact_id, company_id, job_title, department, role_type, is_primary)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateContactCompanyRelation :one
UPDATE contact_company_relations
SET job_title = $2, department = $3, role_type = $4, is_primary = $5, is_active = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteContactCompanyRelation :exec
DELETE FROM contact_company_relations
WHERE id = $1;

-- name: SetPrimaryContact :exec
UPDATE contact_company_relations
SET is_primary = (contact_id = $2)
WHERE company_id = $1;

-- ============================================================================
-- PIPELINES QUERIES
-- ============================================================================

-- name: ListPipelines :many
SELECT * FROM pipelines
WHERE user_id = $1 AND is_active = TRUE
ORDER BY is_default DESC, name ASC;

-- name: GetPipeline :one
SELECT * FROM pipelines
WHERE id = $1;

-- name: GetDefaultPipeline :one
SELECT * FROM pipelines
WHERE user_id = $1 AND is_default = TRUE
LIMIT 1;

-- name: CreatePipeline :one
INSERT INTO pipelines (user_id, name, description, pipeline_type, currency, is_default, color, icon)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdatePipeline :one
UPDATE pipelines
SET name = $2, description = $3, currency = $4, color = $5, icon = $6, is_active = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePipeline :exec
DELETE FROM pipelines
WHERE id = $1 AND user_id = $2;

-- name: SetDefaultPipeline :exec
UPDATE pipelines
SET is_default = (id = $2)
WHERE user_id = $1;

-- ============================================================================
-- PIPELINE STAGES QUERIES
-- ============================================================================

-- name: ListPipelineStages :many
SELECT * FROM pipeline_stages
WHERE pipeline_id = $1
ORDER BY position ASC;

-- name: GetPipelineStage :one
SELECT * FROM pipeline_stages
WHERE id = $1;

-- name: CreatePipelineStage :one
INSERT INTO pipeline_stages (pipeline_id, name, description, position, probability, stage_type, rotting_days, color)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdatePipelineStage :one
UPDATE pipeline_stages
SET name = $2, description = $3, probability = $4, rotting_days = $5, color = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateStagePosition :exec
UPDATE pipeline_stages
SET position = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeletePipelineStage :exec
DELETE FROM pipeline_stages
WHERE id = $1;

-- ============================================================================
-- DEALS QUERIES (CRM Pipeline Deals)
-- ============================================================================

-- name: ListCRMDeals :many
SELECT d.*, ps.name as stage_name, p.name as pipeline_name, c.name as company_name
FROM deals d
JOIN pipeline_stages ps ON d.stage_id = ps.id
JOIN pipelines p ON d.pipeline_id = p.id
LEFT JOIN companies c ON d.company_id = c.id
WHERE d.user_id = $1
  AND (sqlc.narg(pipeline_id)::uuid IS NULL OR d.pipeline_id = sqlc.narg(pipeline_id))
  AND (sqlc.narg(stage_id)::uuid IS NULL OR d.stage_id = sqlc.narg(stage_id))
  AND (sqlc.narg(status)::varchar IS NULL OR d.status = sqlc.narg(status))
  AND (sqlc.narg(owner_id)::varchar IS NULL OR d.owner_id = sqlc.narg(owner_id))
ORDER BY d.updated_at DESC
LIMIT sqlc.arg(limit_val)::int OFFSET sqlc.arg(offset_val)::int;

-- name: ListCRMDealsByStage :many
SELECT d.*, c.name as company_name
FROM deals d
LEFT JOIN companies c ON d.company_id = c.id
WHERE d.stage_id = $1
ORDER BY d.created_at DESC;

-- name: GetCRMDeal :one
SELECT d.*, ps.name as stage_name, p.name as pipeline_name, c.name as company_name
FROM deals d
JOIN pipeline_stages ps ON d.stage_id = ps.id
JOIN pipelines p ON d.pipeline_id = p.id
LEFT JOIN companies c ON d.company_id = c.id
WHERE d.id = $1;

-- name: CreateCRMDeal :one
INSERT INTO deals (
    user_id, pipeline_id, stage_id,
    name, description, amount, currency, probability,
    expected_close_date, owner_id, company_id, primary_contact_id,
    status, priority, lead_source, custom_fields
) VALUES (
    $1, $2, $3,
    $4, $5, $6, $7, $8,
    $9, $10, $11, $12,
    $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateCRMDeal :one
UPDATE deals
SET name = $2, description = $3, amount = $4, probability = $5,
    expected_close_date = $6, owner_id = $7, company_id = $8, primary_contact_id = $9,
    priority = $10, custom_fields = $11, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCRMDealStage :one
UPDATE deals
SET stage_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCRMDealStatus :one
UPDATE deals
SET status = $2,
    lost_reason = $3,
    actual_close_date = CASE WHEN $2 IN ('won', 'lost') THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCRMDeal :exec
DELETE FROM deals
WHERE id = $1 AND user_id = $2;

-- name: GetCRMDealStats :one
SELECT
    COUNT(*) as total_deals,
    COUNT(*) FILTER (WHERE status = 'open') as open_deals,
    COUNT(*) FILTER (WHERE status = 'won') as won_deals,
    COUNT(*) FILTER (WHERE status = 'lost') as lost_deals,
    COALESCE(SUM(amount) FILTER (WHERE status = 'open'), 0) as open_value,
    COALESCE(SUM(amount) FILTER (WHERE status = 'won'), 0) as won_value,
    COALESCE(SUM(amount) FILTER (WHERE status = 'lost'), 0) as lost_value
FROM deals
WHERE user_id = $1
  AND (sqlc.narg(pipeline_id)::uuid IS NULL OR pipeline_id = sqlc.narg(pipeline_id));

-- name: GetCRMDealsByCompany :many
SELECT * FROM deals
WHERE company_id = $1
ORDER BY updated_at DESC;

-- name: SearchCRMDeals :many
SELECT d.*, ps.name as stage_name, c.name as company_name
FROM deals d
JOIN pipeline_stages ps ON d.stage_id = ps.id
LEFT JOIN companies c ON d.company_id = c.id
WHERE d.user_id = $1
  AND d.name ILIKE '%' || $2 || '%'
ORDER BY d.updated_at DESC
LIMIT sqlc.arg(limit_val)::int;

-- ============================================================================
-- CRM ACTIVITIES QUERIES
-- ============================================================================

-- name: ListCRMActivities :many
SELECT * FROM crm_activities
WHERE user_id = $1
  AND (sqlc.narg(activity_type)::crm_activity_type IS NULL OR activity_type = sqlc.narg(activity_type))
  AND (sqlc.narg(is_completed)::boolean IS NULL OR is_completed = sqlc.narg(is_completed))
ORDER BY activity_date DESC
LIMIT sqlc.arg(limit_val)::int OFFSET sqlc.arg(offset_val)::int;

-- name: ListDealActivities :many
SELECT * FROM crm_activities
WHERE deal_id = $1
ORDER BY activity_date DESC;

-- name: ListCompanyActivities :many
SELECT * FROM crm_activities
WHERE company_id = $1
ORDER BY activity_date DESC;

-- name: ListContactActivities :many
SELECT * FROM crm_activities
WHERE contact_id = $1
ORDER BY activity_date DESC;

-- name: GetCRMActivity :one
SELECT * FROM crm_activities
WHERE id = $1;

-- name: CreateCRMActivity :one
INSERT INTO crm_activities (
    user_id, activity_type, subject, description, outcome,
    deal_id, company_id, contact_id, participants,
    activity_date, duration_minutes,
    call_direction, call_disposition, call_recording_url,
    email_direction, email_message_id,
    meeting_location, meeting_url,
    owner_id, is_completed, completed_by, completed_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11,
    $12, $13, $14,
    $15, $16,
    $17, $18,
    $19, $20, $21, $22
)
RETURNING *;

-- name: UpdateCRMActivity :one
UPDATE crm_activities
SET subject = $2, description = $3, outcome = $4,
    activity_date = $5, duration_minutes = $6,
    is_completed = $7, completed_by = $8, completed_at = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteCRMActivity :one
UPDATE crm_activities
SET is_completed = TRUE, completed_by = $2, completed_at = NOW(), outcome = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCRMActivity :exec
DELETE FROM crm_activities
WHERE id = $1;

-- name: ListPendingActivities :many
SELECT * FROM crm_activities
WHERE user_id = $1
  AND is_completed = FALSE
  AND activity_date <= NOW() + INTERVAL '7 days'
ORDER BY activity_date ASC
LIMIT sqlc.arg(limit_val)::int;

-- name: ListOverdueActivities :many
SELECT * FROM crm_activities
WHERE user_id = $1
  AND is_completed = FALSE
  AND activity_date < NOW()
ORDER BY activity_date ASC;

-- ============================================================================
-- DEAL STAGE HISTORY QUERIES
-- ============================================================================

-- name: ListDealStageHistory :many
SELECT dsh.*, ps_from.name as from_stage_name, ps_to.name as to_stage_name
FROM deal_stage_history dsh
LEFT JOIN pipeline_stages ps_from ON dsh.from_stage_id = ps_from.id
JOIN pipeline_stages ps_to ON dsh.to_stage_id = ps_to.id
WHERE dsh.deal_id = $1
ORDER BY dsh.changed_at DESC;

-- name: GetStageConversionRate :one
SELECT
    from_stage_id,
    to_stage_id,
    COUNT(*) as total_transitions,
    AVG(duration_seconds) as avg_duration_seconds
FROM deal_stage_history
WHERE from_stage_id = $1 AND to_stage_id = $2
GROUP BY from_stage_id, to_stage_id;

-- name: GetAverageTimeInStage :one
SELECT
    to_stage_id,
    AVG(duration_seconds) as avg_seconds,
    COUNT(*) as sample_size
FROM deal_stage_history
WHERE to_stage_id = $1
GROUP BY to_stage_id;
