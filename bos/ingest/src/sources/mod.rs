//! Data source implementations.

use crate::DataRow;
use anyhow::{Context, Result};
use serde_json::Value;
use std::collections::HashMap;
use std::path::Path;

/// File-based data source (JSON, CSV).
pub struct FileSource {
    pub path: std::path::PathBuf,
}

impl FileSource {
    pub fn new(path: impl Into<std::path::PathBuf>) -> Self {
        Self { path: path.into() }
    }
}

impl crate::DataSource for FileSource {
    fn name(&self) -> &str {
        "file"
    }

    async fn read(&self) -> Result<Vec<DataRow>> {
        let ext = self.path.extension()
            .and_then(|e| e.to_str())
            .unwrap_or("");

        match ext {
            "json" => read_json_file(&self.path).await,
            _ => read_json_file(&self.path).await, // Default to JSON
        }
    }
}

async fn read_json_file(path: &Path) -> Result<Vec<DataRow>> {
    let content = tokio::fs::read_to_string(path)
        .await
        .with_context(|| format!("Failed to read file: {}", path.display()))?;

    let val: Value = serde_json::from_str(&content)
        .with_context(|| format!("Failed to parse JSON: {}", path.display()))?;

    let rows = match val {
        Value::Array(arr) => arr
            .into_iter()
            .enumerate()
            .map(|(i, item)| {
                let values = item.as_object()
                    .map(|obj| obj.iter().map(|(k, v)| (k.clone(), v.clone())).collect())
                    .unwrap_or_default();
                let mut metadata = HashMap::new();
                metadata.insert("source".to_string(), path.display().to_string());
                metadata.insert("row".to_string(), i.to_string());
                DataRow { values, metadata }
            })
            .collect(),
        Value::Object(obj) => {
            let values = obj.into_iter().map(|(k, v)| (k, v)).collect();
            let mut metadata = HashMap::new();
            metadata.insert("source".to_string(), path.display().to_string());
            vec![DataRow { values, metadata }]
        }
        _ => vec![],
    };

    Ok(rows)
}
