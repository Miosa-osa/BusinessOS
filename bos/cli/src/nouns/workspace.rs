use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct WorkspaceCreated {
    pub path: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct WorkspaceValidated {
    pub is_valid: bool,
    pub workspace_path: String,
    pub tables: usize,
    pub relationships: usize,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct WorkspaceExported {
    pub output_path: String,
    pub tables_exported: usize,
    pub format: String,
}

#[noun("workspace", "ODCS workspace operations")]

/// Initialize a new ODCS workspace
///
/// # Arguments
/// * `name` - Workspace name
/// * `description` - Description [hide]
/// * `output` - Output directory [hide]
#[verb("init")]
fn init(
    name: String,
    description: Option<String>,
    output: Option<String>,
) -> Result<WorkspaceCreated> {
    let opts = bos_core::workspace::WorkspaceInitOptions {
        name,
        description,
        output_dir: output,
    };
    let dir = bos_core::WorkspaceGenerator::init(&opts)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(WorkspaceCreated {
        path: dir.display().to_string(),
    })
}

/// Validate a workspace
///
/// # Arguments
/// * `path` - Workspace directory [default: .]
#[verb("validate")]
fn validate(path: Option<String>) -> Result<WorkspaceValidated> {
    let path_str = path.unwrap_or_else(|| ".".to_string());
    let p = std::path::Path::new(&path_str);
    let result = bos_core::WorkspaceGenerator::validate(p)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(WorkspaceValidated {
        is_valid: result.is_valid,
        workspace_path: result.workspace_path,
        tables: result.tables,
        relationships: result.relationships,
        errors: result.errors,
        warnings: result.warnings,
    })
}

/// Export workspace to format
///
/// # Arguments
/// * `path` - Workspace directory [default: .]
/// * `format` - Output format (odc, json) [default: json]
#[verb("export")]
fn export(path: Option<String>, format: Option<String>) -> Result<WorkspaceExported> {
    let path_str = path.unwrap_or_else(|| ".".to_string());
    let fmt = format.unwrap_or_else(|| "json".to_string());
    let p = std::path::Path::new(&path_str);
    let result = bos_core::WorkspaceGenerator::export(p, &fmt)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(WorkspaceExported {
        output_path: result.output_path,
        tables_exported: result.tables_exported,
        format: result.format,
    })
}
