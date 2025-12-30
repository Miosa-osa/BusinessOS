-- name: GetFocusModeTemplates :many
SELECT * FROM focus_mode_templates
WHERE is_active = true
ORDER BY sort_order ASC;

-- name: GetFocusModeTemplateByName :one
SELECT * FROM focus_mode_templates
WHERE name = $1 AND is_active = true;

-- name: GetFocusModeTemplateByID :one
SELECT * FROM focus_mode_templates
WHERE id = $1;

-- name: GetUserFocusConfigurations :many
SELECT fc.*, fmt.name as template_name, fmt.display_name as template_display_name, fmt.icon as template_icon
FROM focus_configurations fc
JOIN focus_mode_templates fmt ON fc.template_id = fmt.id
WHERE fc.user_id = $1
ORDER BY fc.is_favorite DESC, fc.use_count DESC;

-- name: GetUserFocusConfiguration :one
SELECT fc.*, fmt.name as template_name, fmt.display_name as template_display_name
FROM focus_configurations fc
JOIN focus_mode_templates fmt ON fc.template_id = fmt.id
WHERE fc.user_id = $1 AND fmt.name = $2;

-- name: GetUserFocusConfigurationByID :one
SELECT fc.*, fmt.name as template_name, fmt.display_name as template_display_name
FROM focus_configurations fc
JOIN focus_mode_templates fmt ON fc.template_id = fmt.id
WHERE fc.id = $1 AND fc.user_id = $2;

-- name: CreateFocusConfiguration :one
INSERT INTO focus_configurations (
    user_id, template_id, custom_name, temperature, max_tokens,
    output_style, response_format, max_response_length, require_sources,
    auto_search, search_depth, kb_context_limit, include_history_count,
    thinking_enabled, thinking_style, custom_system_prompt, preferred_model,
    auto_load_kb_categories, keyboard_shortcut, is_favorite
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
)
RETURNING *;

-- name: UpdateFocusConfiguration :one
UPDATE focus_configurations SET
    custom_name = COALESCE($3, custom_name),
    temperature = COALESCE($4, temperature),
    max_tokens = COALESCE($5, max_tokens),
    output_style = COALESCE($6, output_style),
    response_format = COALESCE($7, response_format),
    max_response_length = COALESCE($8, max_response_length),
    require_sources = COALESCE($9, require_sources),
    auto_search = COALESCE($10, auto_search),
    search_depth = COALESCE($11, search_depth),
    kb_context_limit = COALESCE($12, kb_context_limit),
    include_history_count = COALESCE($13, include_history_count),
    thinking_enabled = COALESCE($14, thinking_enabled),
    thinking_style = COALESCE($15, thinking_style),
    custom_system_prompt = COALESCE($16, custom_system_prompt),
    preferred_model = COALESCE($17, preferred_model),
    auto_load_kb_categories = COALESCE($18, auto_load_kb_categories),
    keyboard_shortcut = COALESCE($19, keyboard_shortcut),
    is_favorite = COALESCE($20, is_favorite),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteFocusConfiguration :exec
DELETE FROM focus_configurations
WHERE id = $1 AND user_id = $2;

-- name: IncrementFocusConfigurationUseCount :exec
UPDATE focus_configurations SET
    use_count = use_count + 1,
    last_used_at = NOW()
WHERE user_id = $1 AND template_id = (
    SELECT id FROM focus_mode_templates WHERE name = $2
);

-- name: ToggleFocusConfigurationFavorite :one
UPDATE focus_configurations SET
    is_favorite = NOT is_favorite,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetEffectiveFocusSettings :one
-- Returns merged settings (template defaults + user overrides)
SELECT
    fmt.name,
    fmt.display_name,
    COALESCE(fc.custom_name, fmt.display_name) as effective_name,
    COALESCE(fc.preferred_model, fmt.default_model) as effective_model,
    COALESCE(fc.temperature, fmt.temperature) as effective_temperature,
    COALESCE(fc.max_tokens, fmt.max_tokens) as effective_max_tokens,
    COALESCE(fc.output_style, fmt.output_style) as effective_output_style,
    COALESCE(fc.response_format, fmt.response_format) as effective_response_format,
    COALESCE(fc.max_response_length, fmt.max_response_length) as effective_max_response_length,
    COALESCE(fc.require_sources, fmt.require_sources) as effective_require_sources,
    COALESCE(fc.auto_search, fmt.auto_search) as effective_auto_search,
    COALESCE(fc.search_depth, fmt.search_depth) as effective_search_depth,
    COALESCE(fc.kb_context_limit, fmt.kb_context_limit) as effective_kb_context_limit,
    COALESCE(fc.include_history_count, fmt.include_history_count) as effective_include_history_count,
    COALESCE(fc.thinking_enabled, fmt.thinking_enabled) as effective_thinking_enabled,
    COALESCE(fc.thinking_style, fmt.thinking_style) as effective_thinking_style,
    COALESCE(fc.custom_system_prompt, '') as custom_system_prompt,
    fmt.system_prompt_prefix,
    fmt.system_prompt_suffix,
    fc.auto_load_kb_categories
FROM focus_mode_templates fmt
LEFT JOIN focus_configurations fc ON fc.template_id = fmt.id AND fc.user_id = $1
WHERE fmt.name = $2 AND fmt.is_active = true;

-- name: GetFocusContextPresets :many
SELECT * FROM focus_context_presets
WHERE user_id = $1
ORDER BY name ASC;

-- name: CreateFocusContextPreset :one
INSERT INTO focus_context_presets (
    user_id, name, description, kb_artifact_ids, kb_categories, project_ids, default_search_queries, search_domains
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateFocusContextPreset :one
UPDATE focus_context_presets SET
    name = COALESCE($3, name),
    description = COALESCE($4, description),
    kb_artifact_ids = COALESCE($5, kb_artifact_ids),
    kb_categories = COALESCE($6, kb_categories),
    project_ids = COALESCE($7, project_ids),
    default_search_queries = COALESCE($8, default_search_queries),
    search_domains = COALESCE($9, search_domains),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteFocusContextPreset :exec
DELETE FROM focus_context_presets
WHERE id = $1 AND user_id = $2;

-- name: LinkPresetToFocusConfig :exec
INSERT INTO focus_configuration_presets (focus_config_id, preset_id, sort_order)
VALUES ($1, $2, $3)
ON CONFLICT (focus_config_id, preset_id) DO UPDATE SET sort_order = $3;

-- name: UnlinkPresetFromFocusConfig :exec
DELETE FROM focus_configuration_presets
WHERE focus_config_id = $1 AND preset_id = $2;

-- name: GetPresetsForFocusConfig :many
SELECT fcp.* FROM focus_context_presets fcp
JOIN focus_configuration_presets link ON fcp.id = link.preset_id
WHERE link.focus_config_id = $1
ORDER BY link.sort_order ASC;
