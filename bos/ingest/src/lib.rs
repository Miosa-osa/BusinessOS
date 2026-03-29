//! bos-ingest — Data ingestion for the BusinessOS data layer.
//!
//! Provides a `DataSource` trait and implementations for ingesting data
//! from files, APIs, and other sources into ODCS-compatible formats.

pub mod sources;

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// A single row of ingested data.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DataRow {
    /// Column name → value mapping.
    pub values: HashMap<String, serde_json::Value>,
    /// Source metadata (file path, API endpoint, timestamp, etc.).
    pub metadata: HashMap<String, String>,
}

/// Result of an ingestion operation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IngestResult {
    pub source: String,
    pub rows_ingested: usize,
    pub errors: Vec<String>,
    pub duration_ms: u64,
}

/// Trait for data sources that bos can ingest from.
pub trait DataSource: Send + Sync {
    /// Human-readable name of this source.
    fn name(&self) -> &str;

    /// Read all rows from this source.
    fn read(&self) -> impl std::future::Future<Output = Result<Vec<DataRow>>> + Send;
}
