use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct SearchResult {
    pub rows: usize,
    pub columns: Vec<String>,
    pub bindings: Vec<std::collections::HashMap<String, String>>,
}

#[noun("search", "Semantic SPARQL search")]

/// Execute a SPARQL SELECT query against RDF data
///
/// # Arguments
/// * `query` - SPARQL SELECT query string
/// * `rdf` - Path to RDF data file (N-Triples or Turtle)
#[verb("sparql")]
fn sparql(query: String, rdf: String) -> Result<SearchResult> {
    let search = bos_core::SemanticSearch::new();
    let data = std::fs::read_to_string(&rdf)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to read RDF: {e}")))?;
    search.load_rdf(&data)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let result = search.query(&query)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    Ok(SearchResult {
        rows: result.rows,
        columns: result.columns,
        bindings: result.bindings,
    })
}

/// List all entity types in the RDF data
///
/// # Arguments
/// * `rdf` - Path to RDF data file (N-Triples or Turtle)
#[verb("types")]
fn types(rdf: String) -> Result<SearchResult> {
    let search = bos_core::SemanticSearch::new();
    let data = std::fs::read_to_string(&rdf)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to read RDF: {e}")))?;
    search.load_rdf(&data)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let entity_types = search.list_types()
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let bindings: Vec<std::collections::HashMap<String, String>> = entity_types
        .into_iter()
        .map(|t| {
            let mut m = std::collections::HashMap::new();
            m.insert("type".to_string(), t);
            m
        })
        .collect();

    Ok(SearchResult {
        rows: bindings.len(),
        columns: vec!["type".to_string()],
        bindings,
    })
}
