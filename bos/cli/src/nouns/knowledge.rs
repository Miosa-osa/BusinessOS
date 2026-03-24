use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct KnowledgeIndexed {
    pub total_articles: usize,
    pub workspace: String,
    pub types: Vec<TypeCount>,
}

#[derive(Serialize)]
pub struct TypeCount {
    pub type_name: String,
    pub count: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct KnowledgeExported {
    pub output_path: String,
}

#[noun("knowledge", "Knowledge base management")]

/// Index knowledge articles from a directory
///
/// # Arguments
/// * `directory` - Directory to index
#[verb("index")]
fn index(directory: String) -> Result<KnowledgeIndexed> {
    let path = std::path::Path::new(&directory);
    let idx = bos_core::KnowledgeBase::index(path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let types: Vec<TypeCount> = idx.types.iter().map(|(t, count)| TypeCount {
        type_name: t.clone(),
        count: *count,
    }).collect();
    Ok(KnowledgeIndexed {
        total_articles: idx.total_articles,
        workspace: idx.workspace,
        types,
    })
}

/// Export knowledge base
///
/// # Arguments
/// * `workspace` - Workspace directory [default: .]
#[verb("export")]
fn export(workspace: Option<String>) -> Result<KnowledgeExported> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    let path = std::path::Path::new(&ws);
    let md = bos_core::KnowledgeBase::export(path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let output = std::path::Path::new(&ws).join("knowledge.md");
    std::fs::write(&output, &md)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(KnowledgeExported {
        output_path: output.display().to_string(),
    })
}
