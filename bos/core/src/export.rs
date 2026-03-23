//! Export utilities for BPMN, DMN, and ODCS formats.

use anyhow::Result;
use serde::{Deserialize, Serialize};

/// Result of a data export operation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExportResult {
    pub format: String,
    pub output_path: String,
    pub entities_count: usize,
}

/// Export manager.
pub struct ExportManager;

impl ExportManager {
    /// Export data to ODCS format.
    pub fn to_odcs(data: &serde_json::Value, output_path: &str) -> Result<ExportResult> {
        let content = serde_yaml::to_string(data)?;
        std::fs::write(output_path, &content)?;
        Ok(ExportResult {
            format: "odc".to_string(),
            output_path: output_path.to_string(),
            entities_count: 1,
        })
    }

    /// Export data to JSON format.
    pub fn to_json(data: &serde_json::Value, output_path: &str) -> Result<ExportResult> {
        let content = serde_json::to_string_pretty(data)?;
        std::fs::write(output_path, &content)?;
        Ok(ExportResult {
            format: "json".to_string(),
            output_path: output_path.to_string(),
            entities_count: 1,
        })
    }
}
