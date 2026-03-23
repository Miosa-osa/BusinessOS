//! Schema validation and conversion.
//!
//! Bridges data-modelling-sdk's `convert_to_odcs` for universal schema conversion.

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::Path;

/// Detected or specified format hint.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum FormatHint {
    Sql,
    JsonSchema,
    Avro,
    Protobuf,
    Odc,
}

impl FormatHint {
    pub fn from_str_lossy(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "sql" | "ddl" | "mysql" | "postgresql" | "sqlite" => Some(Self::Sql),
            "json" | "jsonschema" | "json-schema" => Some(Self::JsonSchema),
            "avro" => Some(Self::Avro),
            "proto" | "protobuf" => Some(Self::Protobuf),
            "odc" | "odcs" => Some(Self::Odc),
            _ => None,
        }
    }

    pub fn to_sdk_hint(&self) -> Option<&'static str> {
        match self {
            Self::Sql => Some("sql"),
            Self::JsonSchema => Some("json_schema"),
            Self::Avro => Some("avro"),
            Self::Protobuf => Some("protobuf"),
            Self::Odc => None,
        }
    }
}

/// Result of validating a schema.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SchemaValidationResult {
    pub path: String,
    pub is_valid: bool,
    pub format: Option<String>,
    pub tables: usize,
    pub columns: usize,
    pub errors: Vec<String>,
}

/// Result of converting a schema to ODCS.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SchemaConversionResult {
    pub source_path: String,
    pub detected_format: String,
    pub odcs_output: String,
    pub table_count: usize,
    pub column_count: usize,
    pub relationship_count: usize,
    pub conversion_time_ms: u64,
}

/// Schema converter using data-modelling-sdk.
pub struct SchemaConverter;

impl SchemaConverter {
    /// Validate a schema file against ODCS format.
    pub fn validate(path: &Path, against: Option<FormatHint>) -> Result<SchemaValidationResult> {
        let content = std::fs::read_to_string(path)
            .with_context(|| format!("Failed to read schema: {}", path.display()))?;

        let mut errors = Vec::new();
        let mut tables = 0;
        let mut columns = 0;

        let hint = against.or_else(|| Self::detect_format(path, &content));
        let format_name = hint.map(|h| format!("{:?}", h)).unwrap_or_else(|| "unknown".to_string());

        match hint {
            Some(FormatHint::Odc) => {
                // Parse as ODCS data model
                match serde_json::from_str::<data_modelling_sdk::DataModel>(&content) {
                    Ok(model) => {
                        tables = model.tables.len();
                        columns = model.tables.iter().map(|t| t.columns.len()).sum();
                    }
                    Err(e) => errors.push(format!("Invalid ODCS: {e}")),
                }
            }
            Some(FormatHint::JsonSchema) => {
                // Validate as JSON Schema
                match serde_json::from_str::<serde_json::Value>(&content) {
                    Ok(val) => {
                        if let Some(obj) = val.as_object() {
                            if obj.contains_key("type") || obj.contains_key("$schema") {
                                // Looks like valid JSON Schema
                            } else {
                                errors.push("Missing 'type' or '$schema' field".to_string());
                            }
                        } else {
                            errors.push("JSON Schema must be an object".to_string());
                        }
                    }
                    Err(e) => errors.push(format!("Invalid JSON: {e}")),
                }
            }
            Some(FormatHint::Sql) => {
                // Basic SQL validation — check for common patterns
                if content.to_uppercase().contains("CREATE TABLE") {
                    let table_count = content.to_uppercase().matches("CREATE TABLE").count();
                    tables = table_count;
                } else {
                    errors.push("No CREATE TABLE statements found".to_string());
                }
            }
            _ => {
                // Unknown format — just check it's valid JSON/YAML
                if serde_json::from_str::<serde_json::Value>(&content).is_err() {
                    errors.push("Could not parse as JSON".to_string());
                }
            }
        }

        Ok(SchemaValidationResult {
            path: path.display().to_string(),
            is_valid: errors.is_empty(),
            format: Some(format_name),
            tables,
            columns,
            errors,
        })
    }

    /// Convert a schema file to ODCS format.
    pub fn convert(path: &Path, format_hint: Option<FormatHint>) -> Result<SchemaConversionResult> {
        let content = std::fs::read_to_string(path)
            .with_context(|| format!("Failed to read schema: {}", path.display()))?;

        let hint = format_hint.or_else(|| Self::detect_format(path, &content));
        let format_name = hint
            .map(|h| format!("{:?}", h))
            .unwrap_or_else(|| "unknown".to_string());

        let start = std::time::Instant::now();

        let odcs_output = if let Some(sdk_hint) = hint.and_then(|h| h.to_sdk_hint()) {
            data_modelling_sdk::convert::convert_to_odcs(&content, Some(sdk_hint))
                .map_err(|e| anyhow::anyhow!("Conversion failed: {e}"))?
        } else {
            // Already ODCS or unknown — return as-is
            content.clone()
        };

        let elapsed = start.elapsed().as_millis() as u64;

        // Parse the output to count tables/columns
        let (table_count, column_count, rel_count) = Self::count_odcs_elements(&odcs_output);

        Ok(SchemaConversionResult {
            source_path: path.display().to_string(),
            detected_format: format_name,
            odcs_output,
            table_count,
            column_count,
            relationship_count: rel_count,
            conversion_time_ms: elapsed,
        })
    }

    /// Detect format from file extension and content.
    fn detect_format(path: &Path, content: &str) -> Option<FormatHint> {
        // Try extension first
        let ext = path.extension()?.to_str()?.to_lowercase();
        match ext.as_str() {
            "sql" => return Some(FormatHint::Sql),
            "avsc" | "avro" => return Some(FormatHint::Avro),
            "proto" => return Some(FormatHint::Protobuf),
            "json" => {
                // Could be JSON Schema or ODCS
                if content.contains("$schema") && content.contains("\"type\"") {
                    return Some(FormatHint::JsonSchema);
                }
                return None;
            }
            "yaml" | "yml" => return Some(FormatHint::Odc),
            _ => return None,
        }
    }

    fn count_odcs_elements(content: &str) -> (usize, usize, usize) {
        let val: serde_json::Value = serde_json::from_str(content).unwrap_or_default();
        let tables = val.get("tables")
            .and_then(|t| t.as_array())
            .map(|a| a.len())
            .unwrap_or(0);
        let columns = val.get("tables")
            .and_then(|t| t.as_array())
            .map(|arr| arr.iter()
                .filter_map(|t| t.get("columns"))
                .filter_map(|c| c.as_array())
                .map(|c| c.len())
                .sum())
            .unwrap_or(0);
        let rels = val.get("relationships")
            .and_then(|r| r.as_array())
            .map(|a| a.len())
            .unwrap_or(0);
        (tables, columns, rels)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_format_hint_from_str() {
        assert_eq!(FormatHint::from_str_lossy("sql"), Some(FormatHint::Sql));
        assert_eq!(FormatHint::from_str_lossy("jsonschema"), Some(FormatHint::JsonSchema));
        assert_eq!(FormatHint::from_str_lossy("odc"), Some(FormatHint::Odc));
        assert_eq!(FormatHint::from_str_lossy("avro"), Some(FormatHint::Avro));
        assert_eq!(FormatHint::from_str_lossy("unknown"), None);
    }

    #[test]
    fn test_validate_json_schema() {
        let dir = tempfile::TempDir::new().unwrap();
        let path = dir.path().join("schema.json");
        std::fs::write(&path, r#"{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object"}"#).unwrap();
        let result = SchemaConverter::validate(&path, Some(FormatHint::JsonSchema)).unwrap();
        assert!(result.is_valid);
    }
}
