//! Ontology mapping configuration.
//!
//! Loads and resolves mapping configuration from JSON files that define
//! how relational database tables map to RDF/OWL ontology classes and properties.

use std::collections::HashMap;
use std::path::Path;

use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Errors from mapping configuration operations.
#[derive(Debug, Error)]
pub enum MappingError {
    #[error("Failed to read mapping file: {0}")]
    FileReadError(String),

    #[error("Failed to parse mapping JSON: {0}")]
    ParseError(String),

    #[error("No mapping found for table: {0}")]
    TableNotFound(String),

    #[error("Invalid mapping: {0}")]
    InvalidMapping(String),
}

/// Configuration for ontology mappings from relational tables to RDF.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MappingConfig {
    /// URI prefix definitions (e.g., {"schema": "http://schema.org/", "bpmn": "http://bpmn.org/"}).
    pub prefixes: HashMap<String, String>,

    /// Table-to-ontology mappings.
    #[serde(default)]
    pub mappings: Vec<TableMapping>,

    /// Inter-table relationships.
    #[serde(default)]
    pub relationships: Vec<Relationship>,
}

/// Maps a single relational table to an ontology class.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TableMapping {
    /// SQL table name.
    pub table: String,

    /// Ontology namespace (e.g., "schema", "bpmn").
    pub ontology: String,

    /// OWL class name (e.g., "Organization", "Person").
    pub class: String,

    /// URI template for instances (e.g., "http://businessos.dev/id/{table}").
    pub uri_template: String,

    /// Column-to-property mappings.
    #[serde(default)]
    pub properties: Vec<PropertyMapping>,
}

/// Maps a single table column to an RDF property.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PropertyMapping {
    /// SQL column name.
    pub column: String,

    /// RDF predicate (e.g., "schema:name", "bpmn:name").
    pub predicate: String,

    /// RDF datatype (default: "xsd:string").
    #[serde(default = "default_datatype")]
    pub datatype: String,

    /// Whether this column is the primary key (default: false).
    #[serde(default)]
    pub is_primary_key: bool,

    /// If set, the object is treated as a URI reference to `target_table`.
    pub object_type: Option<String>,

    /// Target table for foreign key references.
    pub target_table: Option<String>,

    /// Value map for enum columns: maps SQL values to URIs.
    #[serde(default)]
    pub value_map: HashMap<String, String>,
}

fn default_datatype() -> String {
    "xsd:string".to_string()
}

/// Defines a relationship between two tables.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Relationship {
    /// Source table.
    pub from_table: String,

    /// Target table.
    pub to_table: String,

    /// Property linking them (e.g., "schema:subOrganization").
    pub property: String,

    /// Inverse property (e.g., "schema:parentOrganization").
    pub inverse: String,
}

/// Resolved prefix map: short prefix -> full URI.
#[derive(Debug, Clone, Default)]
pub struct ResolvedPrefixes {
    map: HashMap<String, String>,
}

impl MappingConfig {
    /// Load a mapping configuration from a JSON file.
    pub fn from_file(path: &Path) -> Result<Self, MappingError> {
        let content = std::fs::read_to_string(path).map_err(|e| {
            MappingError::FileReadError(format!("{}: {}", path.display(), e))
        })?;
        Self::from_str(&content)
    }

    /// Parse a mapping configuration from a JSON string.
    pub fn from_str(json: &str) -> Result<Self, MappingError> {
        serde_json::from_str(json).map_err(|e| MappingError::ParseError(e.to_string()))
    }

    /// Find the mapping for a specific table name.
    pub fn find_mapping(&self, table: &str) -> Option<&TableMapping> {
        self.mappings.iter().find(|m| m.table == table)
    }

    /// Return the primary key column for a table mapping, if defined.
    pub fn primary_key(&self, table: &str) -> Option<&str> {
        self.find_mapping(table).and_then(|m| {
            m.properties.iter().find(|p| p.is_primary_key).map(|p| p.column.as_str())
        })
    }
}

impl ResolvedPrefixes {
    /// Build a resolved prefix map from a MappingConfig.
    ///
    /// Merges the config's `prefixes` map with well-known defaults.
    pub fn from_config(config: &MappingConfig) -> Self {
        let mut map = HashMap::new();

        // Well-known defaults
        map.insert("rdf".to_string(), "http://www.w3.org/1999/02/22-rdf-syntax-ns#".to_string());
        map.insert("rdfs".to_string(), "http://www.w3.org/2000/01/rdf-schema#".to_string());
        map.insert("xsd".to_string(), "http://www.w3.org/2001/XMLSchema#".to_string());
        map.insert("owl".to_string(), "http://www.w3.org/2002/07/owl#".to_string());
        map.insert("prov".to_string(), "http://www.w3.org/ns/prov#".to_string());
        map.insert("bdev".to_string(), "http://businessos.dev/id/".to_string());
        map.insert("bactivity".to_string(), "http://businessos.dev/activity/".to_string());

        // User-defined prefixes override defaults
        for (prefix, uri) in &config.prefixes {
            map.insert(prefix.clone(), uri.clone());
        }

        Self { map }
    }

    /// Resolve a prefixed URI (e.g., "schema:name") to a full URI.
    ///
    /// If the input does not contain a colon, it is returned as-is.
    pub fn resolve(&self, prefixed: &str) -> String {
        if let Some(colon_pos) = prefixed.find(':') {
            let prefix = &prefixed[..colon_pos];
            let local = &prefixed[colon_pos + 1..];
            if let Some(base) = self.map.get(prefix) {
                return format!("{}{}", base, local);
            }
        }
        prefixed.to_string()
    }

    /// Return a SPARQL PREFIX declaration string.
    pub fn to_sparql_prefixes(&self) -> String {
        let mut lines = Vec::new();
        let mut entries: Vec<_> = self.map.iter().collect();
        entries.sort_by_key(|(k, _)| *k);
        for (prefix, uri) in entries {
            lines.push(format!("PREFIX {}: <{}>", prefix, uri));
        }
        lines.join("\n")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    fn sample_config_json() -> String {
        r#"{
            "prefixes": {
                "schema": "http://schema.org/",
                "bpmn": "http://bpmn.org/"
            },
            "mappings": [
                {
                    "table": "organizations",
                    "ontology": "schema",
                    "class": "Organization",
                    "uri_template": "http://businessos.dev/id/organizations",
                    "properties": [
                        {
                            "column": "id",
                            "predicate": "schema:identifier",
                            "datatype": "xsd:integer",
                            "is_primary_key": true
                        },
                        {
                            "column": "name",
                            "predicate": "schema:name"
                        },
                        {
                            "column": "type",
                            "predicate": "schema:organizationType",
                            "value_map": {
                                "LLC": "http://businessos.dev/enum/org-type/llc",
                                "CORP": "http://businessos.dev/enum/org-type/corp"
                            }
                        }
                    ]
                }
            ],
            "relationships": [
                {
                    "from_table": "organizations",
                    "to_table": "departments",
                    "property": "schema:subOrganization",
                    "inverse": "schema:parentOrganization"
                }
            ]
        }"#
        .to_string()
    }

    #[test]
    fn test_parse_mapping_config() {
        let json = sample_config_json();
        let config = MappingConfig::from_str(&json).expect("parse should succeed");
        assert_eq!(config.prefixes.len(), 2);
        assert_eq!(config.mappings.len(), 1);
        assert_eq!(config.relationships.len(), 1);
    }

    #[test]
    fn test_find_mapping() {
        let json = sample_config_json();
        let config = MappingConfig::from_str(&json).expect("parse should succeed");

        let mapping = config.find_mapping("organizations");
        assert!(mapping.is_some());
        let m = mapping.unwrap();
        assert_eq!(m.table, "organizations");
        assert_eq!(m.class, "Organization");
        assert_eq!(m.properties.len(), 3);

        assert!(config.find_mapping("nonexistent").is_none());
    }

    #[test]
    fn test_primary_key() {
        let json = sample_config_json();
        let config = MappingConfig::from_str(&json).expect("parse should succeed");
        assert_eq!(config.primary_key("organizations"), Some("id"));
        assert_eq!(config.primary_key("nonexistent"), None);
    }

    #[test]
    fn test_from_file() {
        let mut tmp = NamedTempFile::new().expect("temp file");
        write!(tmp, "{}", sample_config_json()).expect("write");
        let config = MappingConfig::from_file(tmp.path()).expect("from_file should succeed");
        assert_eq!(config.mappings.len(), 1);
    }

    #[test]
    fn test_resolve_prefixes() {
        let json = sample_config_json();
        let config = MappingConfig::from_str(&json).expect("parse should succeed");
        let resolved = ResolvedPrefixes::from_config(&config);

        assert_eq!(
            resolved.resolve("schema:name"),
            "http://schema.org/name"
        );
        assert_eq!(
            resolved.resolve("bpmn:process"),
            "http://bpmn.org/process"
        );
        // Well-known default
        assert_eq!(
            resolved.resolve("rdf:type"),
            "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
        );
        // No prefix
        assert_eq!(resolved.resolve("http://example.org/foo"), "http://example.org/foo");
    }

    #[test]
    fn test_sparql_prefixes() {
        let json = sample_config_json();
        let config = MappingConfig::from_str(&json).expect("parse should succeed");
        let resolved = ResolvedPrefixes::from_config(&config);
        let prefixes = resolved.to_sparql_prefixes();

        assert!(prefixes.contains("PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>"));
        assert!(prefixes.contains("PREFIX schema: <http://schema.org/>"));
        assert!(prefixes.contains("PREFIX bdev: <http://businessos.dev/id/>"));
    }

    #[test]
    fn test_property_mapping_defaults() {
        // When deserializing from JSON without a datatype field, serde applies
        // the default function. Verify via JSON round-trip.
        let json = r#"{
            "column": "name",
            "predicate": "schema:name"
        }"#;
        let prop: PropertyMapping = serde_json::from_str(json).expect("parse");
        assert_eq!(prop.datatype, "xsd:string");
        assert_eq!(prop.is_primary_key, false);
        assert_eq!(prop.object_type, None);
    }
}
