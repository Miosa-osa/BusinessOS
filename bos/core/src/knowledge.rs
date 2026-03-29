//! Knowledge base management.
//!
//! Provides indexing, search, and export of knowledge articles.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

/// A knowledge base article.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KnowledgeArticle {
    pub number: u32,
    pub title: String,
    pub article_type: ArticleType,
    pub status: ArticleStatus,
    pub content: String,
    pub tags: Vec<String>,
    pub created_at: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ArticleType {
    HowTo,
    Decision,
    Pattern,
    Lesson,
    Other(String),
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ArticleStatus {
    Draft,
    Published,
    Archived,
}

/// Index of all knowledge articles.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KnowledgeIndex {
    pub workspace: String,
    pub total_articles: usize,
    pub types: HashMap<String, usize>,
}

/// Knowledge base manager.
pub struct KnowledgeBase;

impl KnowledgeBase {
    /// Index a directory of knowledge articles.
    pub fn index(workspace_path: &Path) -> Result<KnowledgeIndex> {
        let kb_dir = workspace_path.join("knowledge");
        if !kb_dir.exists() {
            return Ok(KnowledgeIndex {
                workspace: workspace_path.display().to_string(),
                total_articles: 0,
                types: HashMap::new(),
            });
        }

        let mut types: HashMap<String, usize> = HashMap::new();
        let mut total = 0;

        for entry in std::fs::read_dir(&kb_dir)? {
            let entry = entry?;
            let path = entry.path();
            if path.extension().map(|e| e == "json").unwrap_or(false) {
                if let Ok(content) = std::fs::read_to_string(&path) {
                    if let Ok(article) = serde_json::from_str::<KnowledgeArticle>(&content) {
                        let t = match &article.article_type {
                            ArticleType::Other(s) => s.clone(),
                            a => format!("{:?}", a),
                        };
                        *types.entry(t).or_insert(0) += 1;
                        total += 1;
                    }
                }
            }
        }

        Ok(KnowledgeIndex {
            workspace: workspace_path.display().to_string(),
            total_articles: total,
            types,
        })
    }

    /// Export knowledge base as markdown.
    pub fn export(workspace_path: &Path) -> Result<String> {
        let index = Self::index(workspace_path)?;
        let mut md = String::new();
        md.push_str(&format!("# Knowledge Base — {}\n\n", index.workspace));
        md.push_str(&format!("**Total articles:** {}\n\n", index.total_articles));

        if !index.types.is_empty() {
            md.push_str("## Types\n\n");
            for (t, count) in &index.types {
                md.push_str(&format!("- {}: {}\n", t, count));
            }
        }

        Ok(md)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_index_empty_workspace() {
        let tmp = TempDir::new().unwrap();
        let index = KnowledgeBase::index(tmp.path()).unwrap();
        assert_eq!(index.total_articles, 0);
    }
}
