//! MADR Decision Record generation and management.
//!
//! Uses data-modelling-sdk's Decision models for architecture decision records.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

/// A MADR (Machine-Readable Architecture Decision Record).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecisionRecord {
    pub number: u32,
    pub title: String,
    pub status: DecisionStatus,
    pub context: String,
    pub decision: String,
    pub consequences: String,
    pub category: DecisionCategory,
    pub drivers: Vec<String>,
    pub options: Vec<DecisionOption>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DecisionStatus {
    Proposed,
    Accepted,
    Deprecated,
    Superseded,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum DecisionCategory {
    Architecture,
    Data,
    Security,
    Performance,
    Infrastructure,
    Process,
    Other(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecisionOption {
    pub title: String,
    pub pros: Vec<String>,
    pub cons: Vec<String>,
}

/// Index of all decision records in a workspace.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecisionIndex {
    pub workspace: String,
    pub total_decisions: usize,
    pub categories: HashMap<String, usize>,
    pub decisions: Vec<DecisionSummary>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecisionSummary {
    pub number: u32,
    pub title: String,
    pub status: String,
    pub category: String,
}

/// Generates MADR decision records.
pub struct DecisionGenerator;

impl DecisionGenerator {
    /// List all decisions in a workspace directory.
    pub fn list(workspace_path: &Path) -> Result<DecisionIndex> {
        let decisions_dir = workspace_path.join("decisions");
        if !decisions_dir.exists() {
            return Ok(DecisionIndex {
                workspace: workspace_path.display().to_string(),
                total_decisions: 0,
                categories: HashMap::new(),
                decisions: vec![],
            });
        }

        let mut categories: HashMap<String, usize> = HashMap::new();
        let mut summaries = Vec::new();

        for entry in std::fs::read_dir(&decisions_dir)? {
            let entry = entry?;
            let path = entry.path();
            if path.extension().map(|e| e == "json").unwrap_or(false) {
                if let Ok(content) = std::fs::read_to_string(&path) {
                    if let Ok(record) = serde_json::from_str::<DecisionRecord>(&content) {
                        let cat = match &record.category {
                            DecisionCategory::Other(s) => s.clone(),
                            c => format!("{:?}", c),
                        };
                        *categories.entry(cat.clone()).or_insert(0) += 1;
                        summaries.push(DecisionSummary {
                            number: record.number,
                            title: record.title,
                            status: format!("{:?}", record.status),
                            category: cat,
                        });
                    }
                }
            }
        }

        let total = summaries.len();
        Ok(DecisionIndex {
            workspace: workspace_path.display().to_string(),
            total_decisions: total,
            categories,
            decisions: summaries,
        })
    }

    /// Export all decisions as markdown.
    pub fn export_markdown(workspace_path: &Path) -> Result<String> {
        let index = Self::list(workspace_path)?;
        let mut md = String::new();
        md.push_str(&format!("# Decision Records — {}\n\n", index.workspace));
        md.push_str(&format!("**Total:** {} decisions\n\n", index.total_decisions));

        if !index.categories.is_empty() {
            md.push_str("## Categories\n\n");
            for (cat, count) in &index.categories {
                md.push_str(&format!("- {}: {}\n", cat, count));
            }
            md.push('\n');
        }

        md.push_str("## Decisions\n\n");
        for d in &index.decisions {
            md.push_str(&format!("### ADR-{}: {}\n\n", d.number, d.title));
            md.push_str(&format!("- **Status:** {}\n", d.status));
            md.push_str(&format!("- **Category:** {}\n\n", d.category));
        }

        Ok(md)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_list_empty_workspace() {
        let tmp = TempDir::new().unwrap();
        let index = DecisionGenerator::list(tmp.path()).unwrap();
        assert_eq!(index.total_decisions, 0);
    }
}
