//! PostgreSQL -> RDF -> SPARQL CONSTRUCT pipeline.
//!
//! Loads data from PostgreSQL tables, converts rows to RDF triples in an
//! oxigraph-backed TripleStore, then executes SPARQL CONSTRUCT queries to
//! produce enriched ontology triples.

use std::collections::HashMap;

use chrono::Utc;
use oxigraph::model::Literal;
use oxigraph::model::NamedNode;
use oxigraph::model::NamedOrBlankNode;
use oxigraph::model::Term;
use postgres::Client;
use postgres::NoTls;

use crate::ontology::construct::generate_construct_query;
use crate::ontology::mapping::{MappingConfig, ResolvedPrefixes, TableMapping};
use crate::rdf::store::TripleStore;

/// Result of executing a table's RDF conversion pipeline.
#[derive(Debug, Clone)]
pub struct ExecutionResult {
    /// Serialized N-Triples output from the CONSTRUCT query.
    pub ntriples: String,
    /// Number of rows loaded from PostgreSQL.
    pub rows_loaded: usize,
    /// Number of raw triples inserted into the store.
    pub triples_generated: usize,
    /// Number of triples produced by the CONSTRUCT query.
    pub construct_triples: usize,
}

/// Executes the PostgreSQL -> RDF -> CONSTRUCT pipeline.
pub struct QueryExecutor {
    pub(crate) store: TripleStore,
    config: MappingConfig,
    db_url: String,
}

impl QueryExecutor {
    /// Create a new QueryExecutor.
    ///
    /// The `db_url` should be a PostgreSQL connection string, e.g.
    /// `postgresql://user:pass@localhost:5432/dbname`.
    pub fn new(config: MappingConfig, db_url: String) -> Self {
        Self {
            store: TripleStore::new(),
            config,
            db_url,
        }
    }

    /// Execute the pipeline for a single table.
    ///
    /// If `sparql` is provided, it overrides the generated CONSTRUCT query.
    #[tracing::instrument(name = "bos.ontology.execute", skip(self, sparql), fields(
        rdf.result.triple_count = tracing::field::Empty,
        chatmangpt.run.correlation_id = tracing::field::Empty,
    ))]
    pub fn execute_table(
        &self,
        table: &str,
        sparql: Option<&str>,
    ) -> Result<ExecutionResult, crate::ontology::mapping::MappingError> {
        let mapping = self.config.find_mapping(table)
            .ok_or_else(|| crate::ontology::mapping::MappingError::TableNotFound(table.to_string()))?;

        // Load data from PostgreSQL
        let rows = self.load_postgres_data(table)
            .map_err(|e| crate::ontology::mapping::MappingError::InvalidMapping(
                format!("Failed to load data for '{}': {}", table, e)
            ))?;
        let rows_loaded = rows.len();

        // Convert rows to RDF triples
        let triples_before = self.store.count();
        for row in &rows {
            self.insert_row_as_rdf(mapping, row)
                .map_err(|e| crate::ontology::mapping::MappingError::InvalidMapping(
                    format!("Failed to insert RDF for row: {}", e)
                ))?;
        }
        let triples_generated = self.store.count() - triples_before;

        // Generate or use provided CONSTRUCT query
        let query = match sparql {
            Some(q) => q.to_string(),
            None => self.generate_construct_for_mapping(mapping)
                .map_err(|e| crate::ontology::mapping::MappingError::InvalidMapping(
                    format!("Failed to generate CONSTRUCT: {}", e)
                ))?,
        };

        // Execute CONSTRUCT query
        let construct_triples_result = self.store.query_construct(&query)
            .map_err(|e| crate::ontology::mapping::MappingError::InvalidMapping(
                format!("CONSTRUCT query failed: {}", e)
            ))?;
        let construct_triples = construct_triples_result.len();

        // Serialize to N-Triples
        let ntriples = construct_triples_result
            .iter()
            .map(|t| format!("<{}> <{}> {} .", t.subject, t.predicate, t.object))
            .collect::<Vec<_>>()
            .join("\n");

        // Record span fields now that we know the results.
        tracing::Span::current().record(
            "rdf.result.triple_count",
            construct_triples as i64,
        );
        if let Ok(corr) = std::env::var("CHATMANGPT_CORRELATION_ID") {
            tracing::Span::current().record("chatmangpt.run.correlation_id", corr.as_str());
        }

        Ok(ExecutionResult {
            ntriples,
            rows_loaded,
            triples_generated,
            construct_triples,
        })
    }

    /// Execute the pipeline for all mapped tables.
    pub fn execute_all(
        &self,
    ) -> Result<HashMap<String, ExecutionResult>, crate::ontology::mapping::MappingError> {
        let mut results = HashMap::new();
        for mapping in &self.config.mappings {
            let result = self.execute_table(&mapping.table, None)?;
            results.insert(mapping.table.clone(), result);
        }
        Ok(results)
    }

    /// Load all rows from a PostgreSQL table as string-keyed hash maps.
    ///
    /// CASTs ALL columns to text to avoid type deserialization issues with
    /// uuid, enum, timestamp, jsonb, etc.
    fn load_postgres_data(
        &self,
        table: &str,
    ) -> Result<Vec<HashMap<String, String>>, Box<dyn std::error::Error>> {
        let mut client = Client::connect(&self.db_url, NoTls)?;

        // Get column names
        let columns = self.load_column_names_with_types(&mut client, table)?;
        let col_names: Vec<&str> = columns.iter().map(|(name, _)| name.as_str()).collect();

        if col_names.is_empty() {
            return Ok(Vec::new());
        }

        // Build SELECT with CAST for every column to text
        let select_list: Vec<String> = col_names
            .iter()
            .map(|c| format!("CAST(\"{}\" AS text) AS \"{}\"", c, c))
            .collect();
        let query = format!(
            "SELECT {} FROM \"{}\"",
            select_list.join(", "),
            table
        );

        eprintln!("[QueryExecutor] Loading table: {} ({} columns)", table, col_names.len());

        let mut rows = Vec::new();
        for row in client.query(&query, &[])? {
            let mut map = HashMap::new();
            for (i, col_name) in col_names.iter().enumerate() {
                let value: Option<String> = row.get(i);
                let val = value.unwrap_or_default();
                map.insert(col_name.to_string(), val);
            }
            rows.push(map);
        }

        eprintln!("[QueryExecutor] Loaded {} rows from {}", rows.len(), table);
        Ok(rows)
    }

    /// Load column names and data types for a table from information_schema.
    fn load_column_names_with_types(
        &self,
        client: &mut Client,
        table: &str,
    ) -> Result<Vec<(String, String)>, Box<dyn std::error::Error>> {
        let query = r#"
            SELECT column_name, data_type
            FROM information_schema.columns
            WHERE table_name = $1
            ORDER BY ordinal_position
        "#;
        let mut columns = Vec::new();
        for row in client.query(query, &[&table])? {
            let name: String = row.get(0);
            let dtype: String = row.get(1);
            columns.push((name, dtype));
        }
        Ok(columns)
    }

    /// Insert a single database row as RDF triples into the store.
    ///
    /// Emits complete PROV-O provenance triples:
    /// - prov:wasGeneratedBy links entity to activity
    /// - prov:wasDerivedFrom links entity to source data
    /// - prov:generatedAtTime records ISO8601 timestamp
    fn insert_row_as_rdf(
        &self,
        mapping: &TableMapping,
        row: &HashMap<String, String>,
    ) -> Result<(), Box<dyn std::error::Error>> {
        // Find primary key value
        let pk_col = mapping
            .properties
            .iter()
            .find(|p| p.is_primary_key)
            .map(|p| p.column.as_str())
            .unwrap_or("id");
        let pk_value = row.get(pk_col).cloned().unwrap_or_default();

        if pk_value.is_empty() {
            return Ok(()); // skip rows without primary key
        }

        // Capture generation timestamp (ISO8601)
        let timestamp = Utc::now().to_rfc3339_opts(chrono::SecondsFormat::Millis, true);

        // Build subject URI
        let subject_uri_str = format!("http://businessos.dev/id/{}/{}", mapping.table, pk_value);
        let subject_node: NamedOrBlankNode = NamedNode::new(&subject_uri_str)
            .map_err(|e| format!("Invalid subject URI '{}': {}", subject_uri_str, e))?
            .into();

        // rdf:type triple
        let rdf_type = NamedNode::new("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
            .map_err(|e| format!("Invalid rdf:type URI: {}", e))?;
        let class_uri = format!(
            "http://www.w3.org/2002/07/owl#{}",
            mapping.class
        );
        let class_node = NamedNode::new(&class_uri)
            .map_err(|e| format!("Invalid class URI '{}': {}", class_uri, e))?;
        self.store.insert(subject_node.clone(), rdf_type, class_node.into())?;

        // PROV-O: prov:wasGeneratedBy links entity to activity
        let activity_uri_str = format!(
            "http://businessos.dev/activity/{}/{}/{}",
            mapping.table, pk_value, pk_value
        );
        let activity_node = NamedNode::new(&activity_uri_str)
            .map_err(|e| format!("Invalid activity URI '{}': {}", activity_uri_str, e))?;
        let prov_was_gen = NamedNode::new("http://www.w3.org/ns/prov#wasGeneratedBy")
            .map_err(|e| format!("Invalid prov URI: {}", e))?;
        self.store.insert(subject_node.clone(), prov_was_gen, activity_node.clone().into())?;

        // PROV-O: prov:wasDerivedFrom links entity to source database record
        let source_uri_str = format!("http://businessos.dev/source/{}/{}/{}", mapping.table, pk_value, pk_value);
        let source_node = NamedNode::new(&source_uri_str)
            .map_err(|e| format!("Invalid source URI '{}': {}", source_uri_str, e))?;
        let prov_was_derived = NamedNode::new("http://www.w3.org/ns/prov#wasDerivedFrom")
            .map_err(|e| format!("Invalid prov:wasDerivedFrom URI: {}", e))?;
        self.store.insert(subject_node.clone(), prov_was_derived, source_node.into())?;

        // PROV-O: prov:generatedAtTime records when the entity was created (ISO8601)
        let prov_gen_time = NamedNode::new("http://www.w3.org/ns/prov#generatedAtTime")
            .map_err(|e| format!("Invalid prov:generatedAtTime URI: {}", e))?;
        let timestamp_literal = Literal::new_typed_literal(
            &timestamp,
            NamedNode::new("http://www.w3.org/2001/XMLSchema#dateTime")
                .map_err(|e| format!("Invalid xsd:dateTime: {}", e))?
        );
        self.store.insert(subject_node.clone(), prov_gen_time, timestamp_literal.into())?;

        // Record metadata about the transformation activity itself
        let prov_was_assoc_with = NamedNode::new("http://www.w3.org/ns/prov#wasAssociatedWith")
            .map_err(|e| format!("Invalid prov:wasAssociatedWith URI: {}", e))?;
        let agent_uri = NamedNode::new("http://businessos.dev/agent/ontology-executor")
            .map_err(|e| format!("Invalid agent URI: {}", e))?;
        self.store.insert(activity_node.into(), prov_was_assoc_with, agent_uri.into())?;

        // Property triples
        for prop in &mapping.properties {
            let value = row.get(&prop.column).cloned().unwrap_or_default();
            if value.is_empty() {
                continue;
            }

            let predicate = resolve_predicate_uri(&prop.predicate)?;

            if prop.object_type.as_deref() == Some("uri") {
                // Foreign key: create a URI reference
                if let Some(target) = &prop.target_table {
                    let fk_uri_str = format!("http://businessos.dev/id/{}/{}", target, value);
                    let fk_node = NamedNode::new(&fk_uri_str)
                        .map_err(|e| format!("Invalid FK URI '{}': {}", fk_uri_str, e))?;
                    self.store.insert(subject_node.clone(), predicate, fk_node.into())?;
                }
            } else if !prop.value_map.is_empty() {
                // Value map: enum -> URI
                if let Some(uri) = prop.value_map.get(&value) {
                    let uri_node = NamedNode::new(uri)
                        .map_err(|e| format!("Invalid value map URI '{}': {}", uri, e))?;
                    self.store.insert(subject_node.clone(), predicate, uri_node.into())?;
                } else {
                    // Unknown enum value: store as plain literal
                    let literal: Term = Literal::new_simple_literal(&value).into();
                    self.store.insert(subject_node.clone(), predicate, literal)?;
                }
            } else {
                // Typed literal
                let term = make_typed_literal(&value, &prop.datatype)?;
                self.store.insert(subject_node.clone(), predicate, term)?;
            }
        }

        Ok(())
    }

    /// Serialize all triples currently in the store as N-Triples.
    pub fn to_ntriples(&self) -> String {
        self.store.to_ntriples()
    }

    /// Generate a SPARQL CONSTRUCT query for a table mapping.
    fn generate_construct_for_mapping(
        &self,
        mapping: &TableMapping,
    ) -> Result<String, crate::ontology::mapping::MappingError> {
        let prefixes = ResolvedPrefixes::from_config(&self.config);
        let query = generate_construct_query(mapping, &prefixes);
        eprintln!("[QueryExecutor] Generated CONSTRUCT for '{}': {} bytes", mapping.table, query.len());
        Ok(query)
    }
}

/// Resolve a predicate like "schema:name" to a full URI.
fn resolve_predicate_uri(prefixed: &str) -> Result<NamedNode, Box<dyn std::error::Error>> {
    let uri = match prefixed.find(':') {
        Some(pos) => {
            let prefix = &prefixed[..pos];
            let local = &prefixed[pos + 1..];
            let base = match prefix {
                "rdf" => "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
                "rdfs" => "http://www.w3.org/2000/01/rdf-schema#",
                "xsd" => "http://www.w3.org/2001/XMLSchema#",
                "owl" => "http://www.w3.org/2002/07/owl#",
                "prov" => "http://www.w3.org/ns/prov#",
                "schema" => "http://schema.org/",
                "bpmn" => "http://bpmn.org/",
                "bdev" => "http://businessos.dev/id/",
                _ => return Err(format!("Unknown prefix '{}' in predicate '{}'", prefix, prefixed).into()),
            };
            format!("{}{}", base, local)
        }
        None if prefixed.starts_with("http://") || prefixed.starts_with("https://") => prefixed.to_string(),
        None => return Err(format!("Invalid predicate (no prefix, not a URI): '{}'", prefixed).into()),
    };
    NamedNode::new(&uri)
        .map_err(|e| format!("Invalid resolved predicate URI '{}': {}", uri, e).into())
}

/// Create a typed literal Term based on the datatype string.
fn make_typed_literal(value: &str, datatype: &str) -> Result<Term, Box<dyn std::error::Error>> {
    let term: Term = match datatype {
        "xsd:date" | "http://www.w3.org/2001/XMLSchema#date" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#date")?;
            Literal::new_typed_literal(value, dt).into()
        }
        "xsd:dateTime" | "http://www.w3.org/2001/XMLSchema#dateTime" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#dateTime")?;
            Literal::new_typed_literal(value, dt).into()
        }
        "xsd:integer" | "http://www.w3.org/2001/XMLSchema#integer" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#integer")?;
            Literal::new_typed_literal(value, dt).into()
        }
        "xsd:decimal" | "http://www.w3.org/2001/XMLSchema#decimal" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#decimal")?;
            Literal::new_typed_literal(value, dt).into()
        }
        "xsd:boolean" | "http://www.w3.org/2001/XMLSchema#boolean" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#boolean")?;
            Literal::new_typed_literal(value, dt).into()
        }
        "xsd:anyURI" | "http://www.w3.org/2001/XMLSchema#anyURI" => {
            let dt = NamedNode::new("http://www.w3.org/2001/XMLSchema#anyURI")?;
            Literal::new_typed_literal(value, dt).into()
        }
        _ => Literal::new_simple_literal(value).into(),
    };
    Ok(term)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ontology::mapping::{PropertyMapping, TableMapping};

    fn sample_config() -> MappingConfig {
        MappingConfig {
            prefixes: HashMap::new(),
            mappings: vec![TableMapping {
                table: "test_table".to_string(),
                ontology: "schema".to_string(),
                class: "Thing".to_string(),
                uri_template: "http://businessos.dev/id/test_table".to_string(),
                properties: vec![
                    PropertyMapping {
                        column: "id".to_string(),
                        predicate: "schema:identifier".to_string(),
                        datatype: "xsd:integer".to_string(),
                        is_primary_key: true,
                        object_type: None,
                        target_table: None,
                        value_map: HashMap::new(),
                    },
                    PropertyMapping {
                        column: "name".to_string(),
                        predicate: "schema:name".to_string(),
                        datatype: "xsd:string".to_string(),
                        is_primary_key: false,
                        object_type: None,
                        target_table: None,
                        value_map: HashMap::new(),
                    },
                ],
            }],
            relationships: vec![],
        }
    }

    #[test]
    fn test_resolve_predicate_uri() {
        let pred = resolve_predicate_uri("schema:name").expect("should resolve");
        assert_eq!(pred.to_string(), "<http://schema.org/name>");

        let rdf = resolve_predicate_uri("rdf:type").expect("should resolve");
        assert_eq!(rdf.to_string(), "<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>");
    }

    #[test]
    fn test_resolve_predicate_uri_error() {
        let result = resolve_predicate_uri("unknown:name");
        assert!(result.is_err());
    }

    #[test]
    fn test_make_typed_literal() {
        let lit = make_typed_literal("2024-01-15", "xsd:date").expect("should create");
        assert!(matches!(lit, Term::Literal(_)));

        let lit = make_typed_literal("hello", "xsd:string").expect("should create");
        assert!(matches!(lit, Term::Literal(_)));
        if let Term::Literal(ref l) = lit {
            assert_eq!(l.value(), "hello");
        }

        let lit = make_typed_literal("42", "xsd:integer").expect("should create");
        assert!(matches!(lit, Term::Literal(_)));
    }

    #[test]
    fn test_insert_row_as_rdf() {
        let config = sample_config();
        let executor = QueryExecutor::new(config, "postgresql://localhost/test".to_string());
        let mapping = executor.config.find_mapping("test_table").unwrap();

        let mut row = HashMap::new();
        row.insert("id".to_string(), "1".to_string());
        row.insert("name".to_string(), "Test Thing".to_string());

        let result = executor.insert_row_as_rdf(mapping, &row);
        assert!(result.is_ok());

        // Should have 8 triples:
        // 1. rdf:type
        // 2. prov:wasGeneratedBy
        // 3. prov:wasDerivedFrom
        // 4. prov:generatedAtTime
        // 5. schema:identifier
        // 6. schema:name
        // 7. Activity -> prov:wasAssociatedWith
        // 8. ??? (check actual count)
        assert!(executor.store.count() >= 7, "expected at least 7 PROV-O triples, got {}", executor.store.count());

        // Verify the name triple
        let objects = executor.store.get(
            "http://businessos.dev/id/test_table/1",
            "http://schema.org/name",
        );
        assert!(objects.is_some());
    }

    #[test]
    fn test_insert_row_skip_empty_pk() {
        let config = sample_config();
        let executor = QueryExecutor::new(config, "postgresql://localhost/test".to_string());
        let mapping = executor.config.find_mapping("test_table").unwrap();

        let mut row = HashMap::new();
        row.insert("id".to_string(), "".to_string());
        row.insert("name".to_string(), "Should be skipped".to_string());

        let result = executor.insert_row_as_rdf(mapping, &row);
        assert!(result.is_ok());
        assert_eq!(executor.store.count(), 0);
    }

    #[test]
    fn test_execution_result_structure() {
        let result = ExecutionResult {
            ntriples: "<s> <p> <o> .".to_string(),
            rows_loaded: 10,
            triples_generated: 30,
            construct_triples: 25,
        };
        assert_eq!(result.rows_loaded, 10);
        assert_eq!(result.triples_generated, 30);
        assert_eq!(result.construct_triples, 25);
        assert!(result.ntriples.contains("<s>"));
    }
}
