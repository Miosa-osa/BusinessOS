use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct SchemaValidated {
    pub is_valid: bool,
    pub path: String,
    pub format: Option<String>,
    pub tables: usize,
    pub columns: usize,
    pub errors: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct SchemaConverted {
    pub source_path: String,
    pub output_path: String,
    pub detected_format: String,
    pub table_count: usize,
    pub column_count: usize,
    pub relationship_count: usize,
    pub conversion_time_ms: u64,
}

#[noun("schema", "Schema validation and conversion")]

/// Validate a schema file
///
/// # Arguments
/// * `path` - Schema file path
/// * `format` - Expected format (sql, json, avro, proto, odc) [hide]
#[verb("validate")]
fn validate(path: String, format: Option<String>) -> Result<SchemaValidated> {
    let p = std::path::Path::new(&path);
    let hint = format.as_deref().and_then(bos_core::FormatHint::from_str_lossy);
    let result = bos_core::SchemaConverter::validate(p, hint)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(SchemaValidated {
        is_valid: result.is_valid,
        path: result.path,
        format: result.format,
        tables: result.tables,
        columns: result.columns,
        errors: result.errors,
    })
}

/// Convert schema to ODCS format
///
/// # Arguments
/// * `input` - Input schema file
/// * `output` - Output file
/// * `from` - Source format hint [hide]
#[verb("convert")]
fn convert(input: String, output: String, from: Option<String>) -> Result<SchemaConverted> {
    let input_path = std::path::Path::new(&input);
    let hint = from.as_deref().and_then(bos_core::FormatHint::from_str_lossy);
    let result = bos_core::SchemaConverter::convert(input_path, hint)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    std::fs::write(&output, &result.odcs_output)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(SchemaConverted {
        source_path: result.source_path,
        output_path: output,
        detected_format: result.detected_format,
        table_count: result.table_count,
        column_count: result.column_count,
        relationship_count: result.relationship_count,
        conversion_time_ms: result.conversion_time_ms,
    })
}
