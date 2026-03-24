use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct ValidationOutput {
    pub is_valid: bool,
    pub workspace: String,
    pub ruleset: Option<String>,
    pub tables: usize,
    pub relationships: usize,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

#[noun("validate", "Validation and compliance")]

/// Validate a workspace with optional ruleset
///
/// # Arguments
/// * `workspace` - Workspace directory [default: .]
/// * `ruleset` - Ruleset (soc2, hipaa) [hide]
#[verb("")]
fn run(workspace: Option<String>, ruleset: Option<String>) -> Result<ValidationOutput> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    let path = std::path::Path::new(&ws);
    let result = bos_core::WorkspaceGenerator::validate(path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(ValidationOutput {
        is_valid: result.is_valid,
        workspace: ws,
        ruleset,
        tables: result.tables,
        relationships: result.relationships,
        errors: result.errors,
        warnings: result.warnings,
    })
}
