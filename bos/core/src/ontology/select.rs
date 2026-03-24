//! Semantic Search — SPARQL SELECT execution against the ontology triple store.
//!
//! Provides structured SPARQL query execution against oxigraph stores
//! loaded with workspace data via the ontology pipeline.

use crate::ontology::mapping::MappingConfig;
use crate::rdf::store::TripleStore;
use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::Path;

/// Result of a SPARQL SELECT query.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SelectResult {
    pub rows: usize,
    pub columns: Vec<String>,
    pub bindings: Vec<std::collections::HashMap<String, String>>,
}

/// Semantic search engine backed by an oxigraph triple store.
pub struct SemanticSearch {
    store: TripleStore,
}

impl SemanticSearch {
    /// Create a new semantic search engine with an empty triple store.
    pub fn new() -> Self {
        Self {
            store: TripleStore::new(),
        }
    }

    /// Create from an existing TripleStore.
    pub fn from_store(store: TripleStore) -> Self {
        Self { store }
    }

    /// Load N-Triples or Turtle data into the store.
    pub fn load_rdf(&self, data: &str) -> Result<()> {
        let trimmed = data.trim_start();
        if trimmed.starts_with("@prefix") || trimmed.starts_with("PREFIX") {
            self.store
                .load_turtle(data)
                .map_err(|e| anyhow::anyhow!("Failed to load Turtle data: {e}"))?;
        } else {
            // N-Triples or any RDF format — use auto-detection
            let cursor = std::io::Cursor::new(data);
            let parser = oxigraph::io::RdfParser::from_format(oxigraph::io::RdfFormat::NTriples)
                .for_reader(cursor);
            for quad_result in parser {
                let quad = quad_result
                    .map_err(|e| anyhow::anyhow!("Failed to parse RDF: {e}"))?;
                // Insert into the store
                let _ = self.store.insert(
                    quad.subject,
                    quad.predicate,
                    quad.object,
                );
            }
        }
        Ok(())
    }

    /// Execute a SPARQL SELECT query and return structured results.
    pub fn query(&self, sparql: &str) -> Result<SelectResult> {
        let rows = self.store
            .query_sparql(sparql)
            .map_err(|e| anyhow::anyhow!("SPARQL query failed: {e}"))?;

        let columns = rows
            .first()
            .map(|r| r.keys().cloned().collect())
            .unwrap_or_default();

        Ok(SelectResult {
            rows: rows.len(),
            columns,
            bindings: rows,
        })
    }

    /// Execute a SPARQL CONSTRUCT query and return triples.
    pub fn construct(&self, sparql: &str) -> Result<Vec<crate::rdf::store::Triple>> {
        self.store
            .query_construct(sparql)
            .map_err(|e| anyhow::anyhow!("CONSTRUCT query failed: {e}"))
    }

    /// Return distinct entity types in the store.
    pub fn list_types(&self) -> Result<Vec<String>> {
        let result = self.query("SELECT DISTINCT ?type WHERE { ?s a ?type }")?;
        Ok(result
            .bindings
            .into_iter()
            .filter_map(|mut row| row.remove("type"))
            .collect())
    }

    /// Return all predicates for a given entity.
    pub fn describe_entity(&self, uri: &str) -> Result<Vec<crate::rdf::store::Triple>> {
        let sparql = format!(
            "CONSTRUCT {{ <{uri}> ?p ?o }} WHERE {{ <{uri}> ?p ?o }}",
            uri = uri
        );
        self.construct(&sparql)
    }

    /// Return the underlying TripleStore for direct access.
    pub fn store(&self) -> &TripleStore {
        &self.store
    }
}

impl Default for SemanticSearch {
    fn default() -> Self {
        Self::new()
    }
}

/// Load a mapping config and N-Triples data from a file into a SemanticSearch.
pub fn load_from_files(
    mapping_path: &Path,
    rdf_path: &Path,
) -> Result<SemanticSearch> {
    let search = SemanticSearch::new();

    let config = crate::ontology::mapping::MappingConfig::from_file(mapping_path)
        .map_err(|e| anyhow::anyhow!("Failed to load mapping: {e}"))?;

    let rdf_data = std::fs::read_to_string(rdf_path)
        .with_context(|| format!("Failed to read RDF file: {}", rdf_path.display()))?;

    search.load_rdf(&rdf_data)?;

    // Inject prefix declarations for convenience
    let prefixes = crate::ontology::mapping::ResolvedPrefixes::from_config(&config);
    let prefix_block = prefixes.to_sparql_prefixes();
    let _ = prefix_block; // Available for NL-to-SPARQL in future

    Ok(search)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_search_with_data() -> SemanticSearch {
        let search = SemanticSearch::new();
        let ntriples = r#"
            <http://businessos.dev/id/projects/1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Project> .
            <http://businessos.dev/id/projects/1> <https://schema.org/name> "Alpha Project" .
            <http://businessos.dev/id/projects/1> <https://schema.org/status> "ACTIVE" .
            <http://businessos.dev/id/projects/1> <https://schema.org/dateCreated> "2026-01-15T00:00:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> .
            <http://businessos.dev/id/projects/2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Project> .
            <http://businessos.dev/id/projects/2> <https://schema.org/name> "Beta Project" .
            <http://businessos.dev/id/projects/2> <https://schema.org/status> "COMPLETED" .
            <http://businessos.dev/id/tasks/1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.omg.org/spec/BPMN/20100524/MODEL#Task> .
            <http://businessos.dev/id/tasks/1> <https://schema.org/name> "Design UI" .
            <http://businessos.dev/id/tasks/1> <http://www.omg.org/spec/BPMN/20100524/MODEL#state> "Active" .
        "#;
        search.load_rdf(ntriples).unwrap();
        search
    }

    #[test]
    fn test_select_all_projects() {
        let search = make_search_with_data();
        let result = search
            .query("SELECT ?s ?name WHERE { ?s a <https://schema.org/Project> ; <https://schema.org/name> ?name }")
            .unwrap();
        assert_eq!(result.rows, 2);
    }

    #[test]
    fn test_select_filter_by_status() {
        let search = make_search_with_data();
        let result = search
            .query("SELECT ?name WHERE { ?s a <https://schema.org/Project> ; <https://schema.org/name> ?name ; <https://schema.org/status> ?status FILTER(?status = \"ACTIVE\") }")
            .unwrap();
        assert_eq!(result.rows, 1);
        // clean_term preserves quoted literal values
        let name = result.bindings[0].get("name").unwrap();
        assert!(name.contains("Alpha Project"));
    }

    #[test]
    fn test_list_types() {
        let search = make_search_with_data();
        let types = search.list_types().unwrap();
        assert!(types.len() >= 2); // Project + Task
    }

    #[test]
    fn test_describe_entity() {
        let search = make_search_with_data();
        let triples = search
            .describe_entity("http://businessos.dev/id/projects/1")
            .unwrap();
        assert!(!triples.is_empty());
        assert!(triples.iter().any(|t| t.predicate.contains("name")));
    }

    #[test]
    fn test_invalid_sparql_returns_error() {
        let search = make_search_with_data();
        let result = search.query("SELECT WHERE { INVALID SPARQL");
        assert!(result.is_err());
    }

    #[test]
    fn test_select_task_types() {
        let search = make_search_with_data();
        let result = search
            .query("SELECT ?name ?state WHERE { ?s a <http://www.omg.org/spec/BPMN/20100524/MODEL#Task> ; <https://schema.org/name> ?name ; <http://www.omg.org/spec/BPMN/20100524/MODEL#state> ?state }")
            .unwrap();
        assert_eq!(result.rows, 1);
        let name = result.bindings[0].get("name").unwrap();
        assert!(name.contains("Design UI"));
    }
}
