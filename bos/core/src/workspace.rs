//! ODCS Workspace operations — init, validate, export.
//!
//! Uses data-modelling-sdk types (Table, Column, Relationship, DataModel)
//! to create and manage ODCS workspaces.

use anyhow::{Context, Result};
use data_modelling_sdk::DataModel;
use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};

/// Options for initializing a new workspace.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkspaceInitOptions {
    pub name: String,
    pub description: Option<String>,
    pub output_dir: Option<String>,
}

/// Result of validating a workspace.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkspaceValidationResult {
    pub workspace_path: String,
    pub is_valid: bool,
    pub tables: usize,
    pub relationships: usize,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

/// Result of exporting a workspace.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkspaceExportResult {
    pub workspace_name: String,
    pub format: String,
    pub output_path: String,
    pub tables_exported: usize,
    pub relationships_exported: usize,
}

/// Generates and manages ODCS workspaces.
pub struct WorkspaceGenerator;

impl WorkspaceGenerator {
    /// Initialize a new ODCS workspace directory.
    pub fn init(opts: &WorkspaceInitOptions) -> Result<PathBuf> {
        let output_dir = opts.output_dir.as_deref()
            .map(Path::new)
            .unwrap_or(Path::new("."));

        let workspace_dir = output_dir.join(&opts.name);
        std::fs::create_dir_all(&workspace_dir)
            .with_context(|| format!("Failed to create workspace: {}", workspace_dir.display()))?;

        // Write workspace manifest
        let manifest = serde_json::json!({
            "name": opts.name,
            "description": opts.description.clone().unwrap_or_default(),
            "version": "1.0.0",
            "created_at": chrono::Utc::now().to_rfc3339(),
            "schema": "odcs",
        });
        let manifest_path = workspace_dir.join("workspace.json");
        std::fs::write(&manifest_path, serde_json::to_string_pretty(&manifest)?)
            .with_context(|| format!("Failed to write manifest: {}", manifest_path.display()))?;

        // Write empty data model using DataModel::new() constructor
        let data_model = DataModel::new(
            opts.name.clone(),
            workspace_dir.to_string_lossy().to_string(),
            "control.yaml".to_string(),
        );
        let model_path = workspace_dir.join("model.json");
        std::fs::write(&model_path, serde_json::to_string_pretty(&data_model)?)
            .with_context(|| format!("Failed to write model: {}", model_path.display()))?;

        Ok(workspace_dir)
    }

    /// Validate an existing workspace directory.
    pub fn validate(workspace_path: &Path) -> Result<WorkspaceValidationResult> {
        let manifest_path = workspace_path.join("workspace.json");
        let model_path = workspace_path.join("model.json");

        let mut errors = Vec::new();
        let mut warnings = Vec::new();

        // Check workspace.json exists
        if !manifest_path.exists() {
            errors.push("workspace.json not found".to_string());
        } else {
            let content = std::fs::read_to_string(&manifest_path)?;
            let manifest: serde_json::Value = serde_json::from_str(&content)
                .unwrap_or_else(|_| serde_json::json!({}));
            if manifest.get("name").is_none() {
                errors.push("workspace.json missing 'name' field".to_string());
            }
        }

        // Check and parse model.json
        let mut table_count = 0;
        let mut rel_count = 0;

        if !model_path.exists() {
            errors.push("model.json not found".to_string());
        } else {
            let content = std::fs::read_to_string(&model_path)?;
            let model: DataModel = serde_json::from_str(&content)
                .unwrap_or_else(|e| {
                    errors.push(format!("model.json parse error: {e}"));
                    DataModel::new(String::new(), ".".to_string(), "control.yaml".to_string())
                });
            table_count = model.tables.len();
            rel_count = model.relationships.len();

            // Validate tables
            for (i, table) in model.tables.iter().enumerate() {
                if table.name.is_empty() {
                    errors.push(format!("Table {i} has empty name"));
                }
                if table.columns.is_empty() {
                    warnings.push(format!("Table '{}' has no columns", table.name));
                }
            }

            // Warn on empty model
            if table_count == 0 {
                warnings.push("Workspace has no tables defined".to_string());
            }
        }

        Ok(WorkspaceValidationResult {
            workspace_path: workspace_path.display().to_string(),
            is_valid: errors.is_empty(),
            tables: table_count,
            relationships: rel_count,
            errors,
            warnings,
        })
    }

    /// Export a workspace to the specified format.
    pub fn export(workspace_path: &Path, format: &str) -> Result<WorkspaceExportResult> {
        let model_path = workspace_path.join("model.json");
        let content = std::fs::read_to_string(&model_path)
            .with_context(|| format!("Failed to read model: {}", model_path.display()))?;

        let model: DataModel = serde_json::from_str(&content)?;

        let output_ext = match format {
            "odc" | "yaml" => "yaml",
            "json" => "json",
            _ => "json",
        };

        let output_path = workspace_path.join(format!("export.{}", output_ext));
        let output_content = match format {
            "odc" | "yaml" => serde_yaml::to_string(&model)?,
            _ => serde_json::to_string_pretty(&model)?,
        };

        std::fs::write(&output_path, &output_content)?;

        Ok(WorkspaceExportResult {
            workspace_name: model.name,
            format: format.to_string(),
            output_path: output_path.display().to_string(),
            tables_exported: model.tables.len(),
            relationships_exported: model.relationships.len(),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_workspace_init() {
        let tmp = TempDir::new().unwrap();
        let opts = WorkspaceInitOptions {
            name: "test-ws".to_string(),
            description: Some("Test workspace".to_string()),
            output_dir: Some(tmp.path().to_string_lossy().to_string()),
        };
        let dir = WorkspaceGenerator::init(&opts).unwrap();
        assert!(dir.join("workspace.json").exists());
        assert!(dir.join("model.json").exists());
    }

    #[test]
    fn test_workspace_validate_valid() {
        let tmp = TempDir::new().unwrap();
        let opts = WorkspaceInitOptions {
            name: "valid-ws".to_string(),
            description: None,
            output_dir: Some(tmp.path().to_string_lossy().to_string()),
        };
        let dir = WorkspaceGenerator::init(&opts).unwrap();
        let result = WorkspaceGenerator::validate(&dir).unwrap();
        assert!(result.is_valid);
    }

    #[test]
    fn test_workspace_validate_missing() {
        let tmp = TempDir::new().unwrap();
        let result = WorkspaceGenerator::validate(tmp.path()).unwrap();
        assert!(!result.is_valid);
        assert!(result.errors.len() >= 2); // missing workspace.json and model.json
    }

    #[test]
    fn test_workspace_export_json() {
        let tmp = TempDir::new().unwrap();
        let opts = WorkspaceInitOptions {
            name: "export-ws".to_string(),
            description: None,
            output_dir: Some(tmp.path().to_string_lossy().to_string()),
        };
        let dir = WorkspaceGenerator::init(&opts).unwrap();
        let result = WorkspaceGenerator::export(&dir, "json").unwrap();
        assert!(Path::new(&result.output_path).exists());
        assert_eq!(result.format, "json");
    }
}
