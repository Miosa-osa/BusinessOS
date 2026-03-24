use bos_ingest::DataSource;
use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DataImported {
    pub rows: usize,
    pub source: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DataExported {
    pub output_path: String,
    pub format: String,
}

#[noun("data", "Data import and export")]

/// Import data from a source
///
/// # Arguments
/// * `source` - Source file or directory
/// * `target` - Target workspace [hide]
#[verb("import")]
fn import(source: String, target: Option<String>) -> Result<DataImported> {
    let _target = target;
    let source_path = source.clone();
    clap_noun_verb::async_verb::run_async(async move {
        let source = bos_ingest::sources::FileSource::new(&source_path);
        let rows: Vec<bos_ingest::DataRow> = source.read().await
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
        Ok(DataImported {
            rows: rows.len(),
            source: source.path.display().to_string(),
        })
    })
}

/// Export data from workspace
///
/// # Arguments
/// * `source` - Source workspace [default: .]
/// * `format` - Output format (odc, json) [default: json]
/// * `output` - Output file
#[verb("export")]
fn export(source: Option<String>, format: Option<String>, output: String) -> Result<DataExported> {
    let source_str = source.unwrap_or_else(|| ".".to_string());
    let fmt = format.unwrap_or_else(|| "json".to_string());
    let p = std::path::Path::new(&source_str);
    let content = std::fs::read_to_string(p)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let val: serde_json::Value = serde_json::from_str(&content)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let result = match fmt.as_str() {
        "odc" | "yaml" => bos_core::export::ExportManager::to_odcs(&val, &output)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?,
        _ => bos_core::export::ExportManager::to_json(&val, &output)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?,
    };
    Ok(DataExported {
        output_path: result.output_path,
        format: result.format,
    })
}
