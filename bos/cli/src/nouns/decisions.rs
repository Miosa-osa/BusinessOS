use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DecisionListed {
    pub total_decisions: usize,
    pub workspace: String,
    pub decisions: Vec<DecisionEntry>,
    pub categories: Vec<CategoryCount>,
}

#[derive(Serialize)]
pub struct DecisionEntry {
    pub number: u32,
    pub status: String,
    pub title: String,
    pub category: String,
}

#[derive(Serialize)]
pub struct CategoryCount {
    pub category: String,
    pub count: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DecisionsExported {
    pub total_decisions: usize,
    pub output_path: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DecisionsGenerated {
    pub from: String,
    pub workspace: String,
    pub status: String,
}

#[noun("decisions", "MADR decision records")]

/// List all decision records
///
/// # Arguments
/// * `workspace` - Workspace directory [default: .]
#[verb("list")]
fn list(workspace: Option<String>) -> Result<DecisionListed> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    let path = std::path::Path::new(&ws);
    let index = bos_core::DecisionGenerator::list(path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let decisions: Vec<DecisionEntry> = index.decisions.iter().map(|d| DecisionEntry {
        number: d.number,
        status: d.status.clone(),
        title: d.title.clone(),
        category: d.category.clone(),
    }).collect();
    let categories: Vec<CategoryCount> = index.categories.iter().map(|(cat, count)| CategoryCount {
        category: cat.clone(),
        count: *count,
    }).collect();
    Ok(DecisionListed {
        total_decisions: index.total_decisions,
        workspace: index.workspace,
        decisions,
        categories,
    })
}

/// Generate decision records from analysis
///
/// # Arguments
/// * `from` - Analysis file (mining.json, etc.)
/// * `workspace` - Workspace directory [default: .]
#[verb("generate")]
fn generate(from: String, workspace: Option<String>) -> Result<DecisionsGenerated> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    Ok(DecisionsGenerated {
        from,
        workspace: ws,
        status: "not_yet_implemented".to_string(),
    })
}

/// Export decisions as markdown or YAML
///
/// # Arguments
/// * `workspace` - Workspace directory [default: .]
/// * `format` - Output format [default: md]
#[verb("export")]
fn export(workspace: Option<String>, format: Option<String>) -> Result<DecisionsExported> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    let fmt = format.unwrap_or_else(|| "md".to_string());
    let path = std::path::Path::new(&ws);
    let md = bos_core::DecisionGenerator::export_markdown(path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let ext = match fmt.as_str() {
        "yaml" => "decisions.yaml",
        _ => "decisions.md",
    };
    let output = std::path::Path::new(&ws).join(ext);
    std::fs::write(&output, &md)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    let total = bos_core::DecisionGenerator::list(path)
        .map(|i| i.total_decisions).unwrap_or(0);
    Ok(DecisionsExported {
        total_decisions: total,
        output_path: output.display().to_string(),
    })
}
