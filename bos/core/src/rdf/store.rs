//! TripleStore backed by oxigraph 0.5.6.
//!
//! Provides an in-memory RDF triple store with SPARQL 1.1 query/update support
//! via the `SparqlEvaluator` API. Serialization to N-Triples and Turtle formats
//! is supported through `oxigraph::io::RdfSerializer`.

use std::collections::HashMap;
use std::io::Cursor;
use std::sync::Arc;

use oxigraph::io::RdfFormat;
use oxigraph::io::RdfParser;
use oxigraph::io::RdfSerializer;
use oxigraph::model::GraphNameRef;
use oxigraph::model::NamedNode;
use oxigraph::model::NamedOrBlankNode;
use oxigraph::model::QuadRef;
use oxigraph::model::Term;
use oxigraph::sparql::QueryResults;
use oxigraph::sparql::SparqlEvaluator;
use oxigraph::store::Store;
use thiserror::Error;

/// Errors produced by TripleStore operations.
#[derive(Debug, Error)]
pub enum TripleStoreError {
    #[error("SPARQL query failed: {0}")]
    QueryError(String),

    #[error("SPARQL update failed: {0}")]
    UpdateError(String),

    #[error("Serialization failed: {0}")]
    SerializationError(String),

    #[error("Parsing failed: {0}")]
    ParseError(String),

    #[error("Store error: {0}")]
    StoreError(String),
}

/// A simple RDF triple with string subject, predicate, and object.
#[derive(Debug, Clone)]
pub struct Triple {
    pub subject: String,
    pub predicate: String,
    pub object: String,
}

/// An in-memory RDF triple store backed by oxigraph.
///
/// Provides insert, query (SELECT/CONSTRUCT), update (INSERT/DELETE), and
/// serialization capabilities. Thread-safe via `Arc<Store>`.
pub struct TripleStore {
    store: Arc<Store>,
}

impl TripleStore {
    /// Create a new empty in-memory triple store.
    pub fn new() -> Self {
        Self {
            store: Arc::new(
                Store::new().expect("Failed to create oxigraph Store"),
            ),
        }
    }

    /// Insert a triple into the default graph.
    ///
    /// Takes a `NamedOrBlankNode` subject, `NamedNode` predicate, and any `Term` object.
    pub fn insert(
        &self,
        subject_ref: NamedOrBlankNode,
        predicate: NamedNode,
        object_term: Term,
    ) -> Result<(), TripleStoreError> {
        let quad = QuadRef::new(
            subject_ref.as_ref(),
            predicate.as_ref(),
            object_term.as_ref(),
            GraphNameRef::DefaultGraph,
        );
        self.store
            .insert(quad)
            .map_err(|e| TripleStoreError::StoreError(e.to_string()))?;
        Ok(())
    }

    /// Get all objects for the given subject and predicate.
    pub fn get(&self, subject: &str, predicate: &str) -> Option<Vec<String>> {
        let query = format!(
            "SELECT ?o WHERE {{ <{subject}> <{predicate}> ?o }}",
            subject = subject,
            predicate = predicate
        );
        let results = SparqlEvaluator::new()
            .parse_query(&query)
            .ok()?
            .on_store(&self.store)
            .execute()
            .ok()?;

        match results {
            QueryResults::Solutions(solutions) => {
                let mut objects = Vec::new();
                for solution in solutions {
                    let sol = solution.ok()?;
                    if let Some(term) = sol.get("o") {
                        objects.push(clean_term(term));
                    }
                }
                if objects.is_empty() { None } else { Some(objects) }
            }
            _ => None,
        }
    }

    /// Return the total number of triples in the store.
    pub fn count(&self) -> usize {
        self.store
            .len()
            .map_err(|e| TripleStoreError::StoreError(e.to_string()))
            .unwrap_or(0)
    }

    /// Serialize all triples to N-Triples format.
    pub fn to_ntriples(&self) -> String {
        let mut buf = Vec::new();
        let mut serializer =
            RdfSerializer::from_format(RdfFormat::NTriples).for_writer(&mut buf);
        for quad_result in self.store.iter() {
            match quad_result {
                Ok(quad) => {
                    if serializer.serialize_quad(&quad).is_err() {
                        tracing::warn!("Failed to serialize quad: {:?}", quad);
                    }
                }
                Err(e) => {
                    tracing::warn!("Failed to iterate quad: {}", e);
                }
            }
        }
        drop(serializer); // flush
        String::from_utf8(buf).unwrap_or_default()
    }

    /// Serialize all triples to Turtle format.
    pub fn to_turtle(&self) -> Result<String, TripleStoreError> {
        let mut buf = Vec::new();
        let mut serializer = RdfSerializer::from_format(RdfFormat::Turtle)
            .with_prefix("bdev", "http://businessos.dev/id/")
            .map_err(|e| TripleStoreError::SerializationError(e.to_string()))?
            .for_writer(&mut buf);
        for quad_result in self.store.iter() {
            match quad_result {
                Ok(quad) => {
                    serializer
                        .serialize_quad(&quad)
                        .map_err(|e: std::io::Error| TripleStoreError::SerializationError(e.to_string()))?;
                }
                Err(e) => {
                    return Err(TripleStoreError::ParseError(e.to_string()));
                }
            }
        }
        drop(serializer); // flush
        String::from_utf8(buf)
            .map_err(|e| TripleStoreError::SerializationError(e.to_string()))
    }

    /// Return all distinct predicates in the store.
    pub fn predicates(&self) -> Vec<String> {
        let query = "SELECT DISTINCT ?p WHERE { ?s ?p ?o }";
        let results = SparqlEvaluator::new()
            .parse_query(query)
            .ok()
            .and_then(|pq| pq.on_store(&self.store).execute().ok());

        match results {
            Some(QueryResults::Solutions(solutions)) => {
                let mut out = Vec::new();
                for sol in solutions {
                    if let Ok(s) = sol {
                        if let Some(term) = s.get("p") {
                            out.push(clean_term(term));
                        }
                    }
                }
                out
            }
            _ => Vec::new(),
        }
    }

    /// Return all distinct subjects in the store.
    pub fn subjects(&self) -> Vec<String> {
        let query = "SELECT DISTINCT ?s WHERE { ?s ?p ?o }";
        let results = SparqlEvaluator::new()
            .parse_query(query)
            .ok()
            .and_then(|pq| pq.on_store(&self.store).execute().ok());

        match results {
            Some(QueryResults::Solutions(solutions)) => {
                let mut out = Vec::new();
                for sol in solutions {
                    if let Ok(s) = sol {
                        if let Some(term) = s.get("s") {
                            out.push(clean_term(term));
                        }
                    }
                }
                out
            }
            _ => Vec::new(),
        }
    }

    /// Return subject -> triple count mapping.
    pub fn subject_counts(&self) -> Vec<(String, usize)> {
        let query = "SELECT ?s (COUNT(?o) AS ?c) WHERE { ?s ?p ?o } GROUP BY ?s";
        let results = SparqlEvaluator::new()
            .parse_query(query)
            .ok()
            .and_then(|pq| pq.on_store(&self.store).execute().ok());

        match results {
            Some(QueryResults::Solutions(solutions)) => {
                let mut out = Vec::new();
                for sol in solutions {
                    if let Ok(s) = sol {
                        let subject = s.get("s").map(clean_term).unwrap_or_default();
                        let count = s
                            .get("c")
                            .and_then(|t| match t {
                                Term::Literal(lit) => lit.value().parse::<usize>().ok(),
                                _ => None,
                            })
                            .unwrap_or(0);
                        out.push((subject, count));
                    }
                }
                out
            }
            _ => Vec::new(),
        }
    }

    /// Return all triples for a given subject.
    pub fn triples_for_subject(&self, subject: &str) -> Vec<Triple> {
        let query = format!(
            "SELECT ?p ?o WHERE {{ <{subject}> ?p ?o }}",
            subject = subject
        );
        let results = SparqlEvaluator::new()
            .parse_query(&query)
            .ok()
            .and_then(|pq| pq.on_store(&self.store).execute().ok());

        match results {
            Some(QueryResults::Solutions(solutions)) => {
                let mut out = Vec::new();
                for sol in solutions {
                    if let Ok(s) = sol {
                        let predicate = s.get("p").map(clean_term).unwrap_or_default();
                        let object = s.get("o").map(clean_term).unwrap_or_default();
                        out.push(Triple {
                            subject: subject.to_string(),
                            predicate,
                            object,
                        });
                    }
                }
                out
            }
            _ => Vec::new(),
        }
    }

    /// Execute a SPARQL SELECT query and return results as rows of key-value maps.
    pub fn query_sparql(
        &self,
        query: &str,
    ) -> Result<Vec<HashMap<String, String>>, TripleStoreError> {
        let results = SparqlEvaluator::new()
            .parse_query(query)
            .map_err(|e| TripleStoreError::QueryError(e.to_string()))?
            .on_store(&self.store)
            .execute()
            .map_err(|e| TripleStoreError::QueryError(e.to_string()))?;

        match results {
            QueryResults::Solutions(solutions) => {
                let mut rows = Vec::new();
                for solution in solutions {
                    let sol = solution
                        .map_err(|e| TripleStoreError::QueryError(e.to_string()))?;
                    let mut row = HashMap::new();
                    for (var, term) in sol.iter() {
                        row.insert(var.as_str().to_string(), clean_term(term));
                    }
                    rows.push(row);
                }
                Ok(rows)
            }
            QueryResults::Graph(_) => {
                // CONSTRUCT query on a SELECT entry point -- return empty
                Ok(Vec::new())
            }
            QueryResults::Boolean(val) => {
                let mut row = HashMap::new();
                row.insert("result".to_string(), val.to_string());
                Ok(vec![row])
            }
        }
    }

    /// Execute a SPARQL CONSTRUCT query and return resulting triples.
    pub fn query_construct(&self, query: &str) -> Result<Vec<Triple>, TripleStoreError> {
        let results = SparqlEvaluator::new()
            .parse_query(query)
            .map_err(|e| TripleStoreError::QueryError(e.to_string()))?
            .on_store(&self.store)
            .execute()
            .map_err(|e| TripleStoreError::QueryError(e.to_string()))?;

        match results {
            QueryResults::Graph(triples) => {
                let mut result = Vec::new();
                for triple_result in triples {
                    let t = triple_result
                        .map_err(|e| TripleStoreError::QueryError(e.to_string()))?;
                    result.push(Triple {
                        subject: clean_node(&t.subject),
                        predicate: clean_uri(&t.predicate.to_string()),
                        object: clean_term(&t.object),
                    });
                }
                Ok(result)
            }
            _ => Ok(Vec::new()),
        }
    }

    /// Execute a SPARQL UPDATE query (INSERT DATA, DELETE DATA, etc.).
    pub fn update_sparql(&self, query: &str) -> Result<(), TripleStoreError> {
        SparqlEvaluator::new()
            .parse_update(query)
            .map_err(|e| TripleStoreError::UpdateError(e.to_string()))?
            .on_store(&self.store)
            .execute()
            .map_err(|e| TripleStoreError::UpdateError(e.to_string()))
    }

    /// Load Turtle-formatted RDF data into the store.
    pub fn load_turtle(&self, turtle_str: &str) -> Result<(), TripleStoreError> {
        let cursor = Cursor::new(turtle_str);
        let parser = RdfParser::from_format(RdfFormat::Turtle).for_reader(cursor);
        for quad_result in parser {
            match quad_result {
                Ok(quad) => {
                    self.store
                        .insert(quad.as_ref())
                        .map_err(|e| TripleStoreError::StoreError(e.to_string()))?;
                }
                Err(e) => {
                    return Err(TripleStoreError::ParseError(e.to_string()));
                }
            }
        }
        Ok(())
    }

    /// Clear all triples from the store. Test-only.
    #[cfg(test)]
    pub fn clear(&self) {
        let _ = self.update_sparql("DELETE WHERE { ?s ?p ?o }");
    }
}

impl Default for TripleStore {
    fn default() -> Self {
        Self::new()
    }
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

/// Strip angle brackets from a `NamedNode::to_string()` output.
///
/// oxrdf 0.3 returns `<http://example.org/...>` format from
/// `NamedNode::to_string()`, but we want bare URIs.
fn clean_uri(raw: &str) -> String {
    if raw.starts_with('<') && raw.ends_with('>') {
        raw[1..raw.len() - 1].to_string()
    } else {
        raw.to_string()
    }
}

/// Clean a `Term` to a human-readable string.
fn clean_term(term: &Term) -> String {
    match term {
        Term::NamedNode(node) => clean_uri(&node.to_string()),
        Term::BlankNode(node) => format!("_:{}", node.to_string()),
        Term::Literal(lit) => {
            if let Some(lang) = lit.language() {
                format!("\"{}\"@{}", lit.value(), lang)
            } else {
                let dt = lit.datatype();
                let dt_str = clean_uri(&dt.to_string());
                if dt_str == "http://www.w3.org/2001/XMLSchema#string" {
                    format!("\"{}\"", lit.value())
                } else {
                    format!("\"{}\"^^<{}>", lit.value(), dt_str)
                }
            }
        }
    }
}

/// Clean a `NamedOrBlankNode` to a string.
fn clean_node(node: &NamedOrBlankNode) -> String {
    match node {
        NamedOrBlankNode::NamedNode(nn) => clean_uri(&nn.to_string()),
        NamedOrBlankNode::BlankNode(bn) => format!("_:{}", bn.to_string()),
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use oxigraph::model::Literal;

    fn make_store() -> TripleStore {
        TripleStore::new()
    }

    fn insert_example_triples(store: &TripleStore) {
        // Insert: Alice knows Bob
        let alice = NamedNode::new("http://example.org/alice").expect("valid IRI");
        let knows = NamedNode::new("http://example.org/knows").expect("valid IRI");
        let bob = NamedNode::new("http://example.org/bob").expect("valid IRI");

        store
            .insert(alice.clone().into(), knows.clone(), bob.clone().into())
            .expect("insert alice->knows->bob");

        // Alice name "Alice"
        let name = NamedNode::new("http://xmlns.com/foaf/0.1/name").expect("valid IRI");
        let alice_lit: Term = Literal::new_simple_literal("Alice").into();
        store
            .insert(alice.clone().into(), name.clone(), alice_lit)
            .expect("insert alice->name->Alice");

        // Alice age 30
        let age = NamedNode::new("http://xmlns.com/foaf/0.1/age").expect("valid IRI");
        let xsd_integer =
            NamedNode::new("http://www.w3.org/2001/XMLSchema#integer").expect("valid IRI");
        let age_lit: Term = Literal::new_typed_literal("30", xsd_integer.clone()).into();
        store
            .insert(alice.clone().into(), age, age_lit)
            .expect("insert alice->age->30");

        // Bob name "Bob"
        let bob_lit: Term = Literal::new_simple_literal("Bob").into();
        store
            .insert(bob.clone().into(), name, bob_lit)
            .expect("insert bob->name->Bob");
    }

    #[test]
    fn test_insert_and_count() {
        let store = make_store();
        assert_eq!(store.count(), 0);

        insert_example_triples(&store);
        assert_eq!(store.count(), 4);
    }

    #[test]
    fn test_get_objects() {
        let store = make_store();
        insert_example_triples(&store);

        let objects = store.get("http://example.org/alice", "http://xmlns.com/foaf/0.1/name");
        assert!(objects.is_some());
        let objects = objects.unwrap();
        assert_eq!(objects.len(), 1);
        assert!(objects[0].contains("Alice"));

        let missing = store.get("http://example.org/alice", "http://example.org/nonexistent");
        assert!(missing.is_none());
    }

    #[test]
    fn test_to_ntriples() {
        let store = make_store();
        insert_example_triples(&store);

        let nt = store.to_ntriples();
        assert!(nt.contains("http://example.org/alice"));
        assert!(nt.contains("http://example.org/knows"));
        assert!(nt.contains("http://example.org/bob"));
    }

    #[test]
    fn test_to_turtle() {
        let store = make_store();
        insert_example_triples(&store);

        let turtle = store.to_turtle().expect("to_turtle should succeed");
        // Turtle format should contain URIs
        assert!(turtle.contains("alice") || turtle.contains("Alice"));
    }

    #[test]
    fn test_predicates() {
        let store = make_store();
        insert_example_triples(&store);

        let preds = store.predicates();
        assert!(preds.len() >= 2); // at least knows and name
        assert!(preds.iter().any(|p| p.contains("knows")));
        assert!(preds.iter().any(|p| p.contains("name")));
    }

    #[test]
    fn test_subjects() {
        let store = make_store();
        insert_example_triples(&store);

        let subs = store.subjects();
        assert!(subs.len() >= 2);
        assert!(subs.iter().any(|s| s.contains("alice")));
        assert!(subs.iter().any(|s| s.contains("bob")));
    }

    #[test]
    fn test_subject_counts() {
        let store = make_store();
        insert_example_triples(&store);

        let counts = store.subject_counts();
        // Alice has 3 triples (knows bob, name, age), Bob has 1 (name)
        let alice_count = counts
            .iter()
            .find(|(s, _)| s.contains("alice"))
            .map(|(_, c)| *c);
        assert_eq!(alice_count, Some(3));

        let bob_count = counts
            .iter()
            .find(|(s, _)| s.contains("bob"))
            .map(|(_, c)| *c);
        assert_eq!(bob_count, Some(1));
    }

    #[test]
    fn test_triples_for_subject() {
        let store = make_store();
        insert_example_triples(&store);

        let triples = store.triples_for_subject("http://example.org/alice");
        assert_eq!(triples.len(), 3);
        assert!(triples.iter().any(|t| t.predicate.contains("knows")));
        assert!(triples.iter().any(|t| t.predicate.contains("name")));
        assert!(triples.iter().any(|t| t.predicate.contains("age")));
    }

    #[test]
    fn test_query_sparql_select() {
        let store = make_store();
        insert_example_triples(&store);

        let query = "SELECT ?name WHERE { ?s <http://xmlns.com/foaf/0.1/name> ?name }";
        let results = store.query_sparql(query).expect("query should succeed");
        assert_eq!(results.len(), 2);
    }

    #[test]
    fn test_query_construct() {
        let store = make_store();
        insert_example_triples(&store);

        let query = "CONSTRUCT { ?s <http://example.org/friendOf> ?o } WHERE { ?s <http://example.org/knows> ?o }";
        let triples = store.query_construct(query).expect("construct should succeed");
        assert_eq!(triples.len(), 1);
        assert!(triples[0].predicate.contains("friendOf"));
        assert!(triples[0].subject.contains("alice"));
        assert!(triples[0].object.contains("bob"));
    }

    #[test]
    fn test_update_sparql() {
        let store = make_store();
        insert_example_triples(&store);
        assert_eq!(store.count(), 4);

        store
            .update_sparql(
                "INSERT DATA { <http://example.org/charlie> <http://xmlns.com/foaf/0.1/name> \"Charlie\" }",
            )
            .expect("update should succeed");
        assert_eq!(store.count(), 5);

        let objects = store.get("http://example.org/charlie", "http://xmlns.com/foaf/0.1/name");
        assert!(objects.is_some());
    }

    #[test]
    fn test_load_turtle() {
        let store = make_store();
        let turtle = r#"
            @prefix ex: <http://example.org/> .
            ex:carol ex:name "Carol" ;
                    ex:age 25 .
        "#;
        store.load_turtle(turtle).expect("load_turtle should succeed");
        assert_eq!(store.count(), 2);

        let objects = store.get("http://example.org/carol", "http://example.org/name");
        assert!(objects.is_some());
        assert!(objects.unwrap()[0].contains("Carol"));
    }
}
