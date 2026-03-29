//! SPARQL CONSTRUCT query generation from mapping configuration.
//!
//! Transforms `TableMapping` definitions into executable SPARQL CONSTRUCT
//! queries that produce RDF triples from relational data patterns.

use crate::ontology::mapping::{PropertyMapping, ResolvedPrefixes, TableMapping};

/// Generate a SPARQL CONSTRUCT query for a table mapping.
///
/// The generated query uses `BIND(IRI(CONCAT(...)))` with `ENCODE_FOR_URI(STR(...))`
/// inside, which is required by oxigraph for IRI construction from string variables.
pub fn generate_construct_query(
    mapping: &TableMapping,
    prefixes: &ResolvedPrefixes,
) -> String {
    let table = &mapping.table;
    let pk = find_primary_key(mapping);
    let pk_var = pk.as_deref().unwrap_or("id");

    let class_uri = prefixes.resolve(&format!("{}:{}", mapping.ontology, mapping.class));
    let rdf_type = prefixes.resolve("rdf:type");

    // Build CONSTRUCT template
    let mut construct_triples: Vec<String> = Vec::new();
    // Build WHERE clause
    let mut where_clauses: Vec<String> = Vec::new();

    // Subject URI
    let subject_var = format!("{}_uri", table);
    let subject_bind = format!(
        "BIND(IRI(CONCAT(\"http://businessos.dev/id/{table}\", ENCODE_FOR_URI(STR(?{pk_var})))) AS ?{subject_var})",
        table = table,
        pk_var = pk_var,
        subject_var = subject_var
    );

    // rdf:type
    construct_triples.push(format!(
        "  ?{subject_var} <{rdf_type}> <{class_uri}> .",
        subject_var = subject_var,
        rdf_type = rdf_type,
        class_uri = class_uri
    ));

    // Activity URI for PROV-O provenance
    let activity_var = format!("activity_uri");
    let activity_bind = format!(
        "BIND(IRI(CONCAT(\"http://businessos.dev/activity/{table}/\", ENCODE_FOR_URI(STR(?{pk_var})))) AS ?{activity_var})",
        table = table,
        pk_var = pk_var,
        activity_var = activity_var
    );

    // PROV-O: prov:wasGeneratedBy
    let prov_was_gen = prefixes.resolve("prov:wasGeneratedBy");
    construct_triples.push(format!(
        "  ?{subject_var} <{prov_was_gen}> ?{activity_var} .",
        subject_var = subject_var,
        prov_was_gen = prov_was_gen,
        activity_var = activity_var
    ));

    // Property triples
    for prop in &mapping.properties {
        let pred_uri = prefixes.resolve(&prop.predicate);
        let var_name = sanitize_var_name(&prop.column);
        let construct_obj = build_construct_object(prop, &var_name);

        construct_triples.push(format!(
            "  ?{subject_var} <{pred_uri}> {construct_obj} .",
            subject_var = subject_var,
            pred_uri = pred_uri,
            construct_obj = construct_obj
        ));

        where_clauses.push(build_where_clause(prop, &var_name));
    }

    // WHERE clause: primary key variable, subject bind, activity bind, property clauses
    let mut where_block = Vec::new();
    where_block.push(format!("  # Subject URI bind"));
    where_block.push(format!("  {}", subject_bind));
    where_block.push(format!("  # Activity URI bind (PROV-O)"));
    where_block.push(format!("  {}", activity_bind));
    where_block.push(format!(""));

    for wc in &where_clauses {
        where_block.push(format!("  {}", wc));
    }

    let prefix_block = prefixes.to_sparql_prefixes();

    format!(
        "# CONSTRUCT query for table: {table}\n\
         # Class: {class_uri}\n\n\
         {prefix_block}\n\n\
         CONSTRUCT {{\n\
         {construct_triples}\n\
         }}\n\
         WHERE {{\n\
         {where_block}\n\
         }}\n",
        table = table,
        class_uri = class_uri,
        prefix_block = prefix_block,
        construct_triples = construct_triples.join("\n"),
        where_block = where_block.join("\n"),
    )
}

/// Find the primary key property in a table mapping.
fn find_primary_key(mapping: &TableMapping) -> Option<String> {
    mapping
        .properties
        .iter()
        .find(|p| p.is_primary_key)
        .map(|p| p.column.clone())
}

/// Sanitize a column name for use as a SPARQL variable name.
fn sanitize_var_name(column: &str) -> String {
    column
        .chars()
        .map(|c| {
            if c.is_alphanumeric() || c == '_' {
                c
            } else {
                '_'
            }
        })
        .collect()
}

/// Build the CONSTRUCT template object for a property.
fn build_construct_object(prop: &PropertyMapping, var_name: &str) -> String {
    // Foreign key reference -> URI
    if prop.object_type.as_deref() == Some("uri") {
        let fk_var = format!("{}_uri", var_name);
        if let Some(target) = &prop.target_table {
            // URI with guard
            return format!(
                "IF(BOUND(?{var}) && ?{var} != \"\", \
                 IRI(CONCAT(\"http://businessos.dev/id/{target}/\", ENCODE_FOR_URI(STR(?{var})))), \
                 UNDEF)",
                var = var_name,
                target = target
            );
        }
        return format!("?{}", fk_var);
    }

    // Value map -> chained IF() calls
    if !prop.value_map.is_empty() {
        return build_value_map_construct(prop);
    }

    // Typed literal
    let datatype_uri = resolve_datatype(&prop.datatype);
    if datatype_uri == "http://www.w3.org/2001/XMLSchema#string" {
        format!("?{}", var_name)
    } else {
        format!("?{}_typed", var_name)
    }
}

/// Build the WHERE clause for a property.
fn build_where_clause(prop: &PropertyMapping, var_name: &str) -> String {
    // Foreign key reference
    if prop.object_type.as_deref() == Some("uri") {
        if let Some(target) = &prop.target_table {
            return format!(
                "# FK reference: {column} -> {target}\n\
                 BIND(IF(BOUND(?{var}) && ?{var} != \"\", \
                 IRI(CONCAT(\"http://businessos.dev/id/{target}/\", ENCODE_FOR_URI(STR(?{var})))), \
                 UNDEF) AS ?{var}_uri)",
                column = prop.column,
                target = target,
                var = var_name
            );
        }
        return format!(
            "# FK: {column}\n?{subject} <{pred}> ?{var}_uri",
            column = prop.column,
            subject = "subject", // placeholder
            pred = prop.predicate,
            var = var_name
        );
    }

    // Value map
    if !prop.value_map.is_empty() {
        return build_value_map_where(prop, var_name);
    }

    // Typed literal
    let datatype_uri = resolve_datatype(&prop.datatype);
    if datatype_uri == "http://www.w3.org/2001/XMLSchema#string" {
        format!(
            "# Property: {column}\n?{subject} <{pred}> ?{var}",
            column = prop.column,
            subject = "subject",
            pred = prop.predicate,
            var = var_name
        )
    } else {
        format!(
            "# Property: {column} (typed)\nBIND(?{var} AS ?{var}_typed)",
            column = prop.column,
            var = var_name
        )
    }
}

/// Build a CONSTRUCT template object using IF() chains for value maps.
fn build_value_map_construct(prop: &PropertyMapping) -> String {
    let mut chain = String::new();
    let entries: Vec<_> = prop.value_map.iter().collect();
    for (i, (value, uri)) in entries.iter().enumerate() {
        let next = if i < entries.len() - 1 {
            build_value_map_construct_inner(&entries[i + 1..])
        } else {
            "UNDEF".to_string()
        };
        if chain.is_empty() {
            chain = format!(
                "IF(?{var} = \"{val}\", <{uri}>, {next})",
                var = sanitize_var_name(&prop.column),
                val = value,
                uri = uri,
                next = next
            );
        }
    }
    chain
}

fn build_value_map_construct_inner(entries: &[(&String, &String)]) -> String {
    if entries.is_empty() {
        return "UNDEF".to_string();
    }
    let (value, uri) = &entries[0];
    format!(
        "IF(?{var} = \"{val}\", <{uri}>, {next})",
        var = "col", // placeholder, actual var from outer call
        val = value,
        uri = uri,
        next = build_value_map_construct_inner(&entries[1..])
    )
}

/// Build a WHERE clause for value-mapped properties.
fn build_value_map_where(prop: &PropertyMapping, var_name: &str) -> String {
    format!(
        "# Value map: {column}\n?{subject} <{pred}> ?{var}",
        column = prop.column,
        subject = "subject",
        pred = prop.predicate,
        var = var_name
    )
}

/// Resolve a short datatype (e.g., "xsd:string") to a full URI.
fn resolve_datatype(short: &str) -> String {
    match short {
        "xsd:string" => "http://www.w3.org/2001/XMLSchema#string".to_string(),
        "xsd:integer" => "http://www.w3.org/2001/XMLSchema#integer".to_string(),
        "xsd:decimal" => "http://www.w3.org/2001/XMLSchema#decimal".to_string(),
        "xsd:float" => "http://www.w3.org/2001/XMLSchema#float".to_string(),
        "xsd:double" => "http://www.w3.org/2001/XMLSchema#double".to_string(),
        "xsd:boolean" => "http://www.w3.org/2001/XMLSchema#boolean".to_string(),
        "xsd:date" => "http://www.w3.org/2001/XMLSchema#date".to_string(),
        "xsd:dateTime" => "http://www.w3.org/2001/XMLSchema#dateTime".to_string(),
        "xsd:anyURI" => "http://www.w3.org/2001/XMLSchema#anyURI".to_string(),
        other if other.starts_with("http://") || other.starts_with("urn:") => other.to_string(),
        other => format!("http://www.w3.org/2001/XMLSchema#{}", other),
    }
}

/// Convenience wrapper: generate CONSTRUCT queries for all mappings in a config.
pub struct ConstructGenerator<'a> {
    config: &'a crate::ontology::mapping::MappingConfig,
    prefixes: ResolvedPrefixes,
}

impl<'a> ConstructGenerator<'a> {
    pub fn new(config: &'a crate::ontology::mapping::MappingConfig) -> Self {
        let prefixes = ResolvedPrefixes::from_config(config);
        Self { config, prefixes }
    }

    /// Generate CONSTRUCT queries for all mapped tables.
    pub fn generate_all(&self) -> Result<std::collections::HashMap<String, String>, crate::ontology::mapping::MappingError> {
        let mut queries = std::collections::HashMap::new();
        for mapping in &self.config.mappings {
            let query = generate_construct_query(mapping, &self.prefixes);
            queries.insert(mapping.table.clone(), query);
        }
        Ok(queries)
    }

    /// Generate a CONSTRUCT query for a specific table.
    pub fn generate_for_table(
        &self,
        table: &str,
    ) -> Result<String, crate::ontology::mapping::MappingError> {
        let mapping = self.config.find_mapping(table)
            .ok_or_else(|| crate::ontology::mapping::MappingError::TableNotFound(table.to_string()))?;
        Ok(generate_construct_query(mapping, &self.prefixes))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ontology::mapping::{MappingConfig, PropertyMapping, TableMapping};

    fn sample_mapping() -> (TableMapping, ResolvedPrefixes) {
        let mapping = TableMapping {
            table: "organizations".to_string(),
            ontology: "schema".to_string(),
            class: "Organization".to_string(),
            uri_template: "http://businessos.dev/id/organizations".to_string(),
            properties: vec![
                PropertyMapping {
                    column: "id".to_string(),
                    predicate: "schema:identifier".to_string(),
                    datatype: "xsd:integer".to_string(),
                    is_primary_key: true,
                    object_type: None,
                    target_table: None,
                    value_map: std::collections::HashMap::new(),
                },
                PropertyMapping {
                    column: "name".to_string(),
                    predicate: "schema:name".to_string(),
                    datatype: "xsd:string".to_string(),
                    is_primary_key: false,
                    object_type: None,
                    target_table: None,
                    value_map: std::collections::HashMap::new(),
                },
                PropertyMapping {
                    column: "parent_id".to_string(),
                    predicate: "schema:parentOrganization".to_string(),
                    datatype: "xsd:string".to_string(),
                    is_primary_key: false,
                    object_type: Some("uri".to_string()),
                    target_table: Some("organizations".to_string()),
                    value_map: std::collections::HashMap::new(),
                },
            ],
        };
        let config_with_schema = MappingConfig {
            prefixes: {
                let mut p = std::collections::HashMap::new();
                p.insert("schema".to_string(), "http://schema.org/".to_string());
                p
            },
            mappings: vec![],
            relationships: vec![],
        };
        let prefixes = ResolvedPrefixes::from_config(&config_with_schema);
        (mapping, prefixes)
    }

    #[test]
    fn test_construct_query_structure() {
        let (mapping, prefixes) = sample_mapping();
        let query = generate_construct_query(&mapping, &prefixes);

        // Should contain CONSTRUCT and WHERE keywords
        assert!(query.contains("CONSTRUCT {"));
        assert!(query.contains("WHERE {"));
        assert!(query.contains("PREFIX"));

        // Should contain the class URI
        assert!(query.contains("http://schema.org/Organization"));

        // Should contain rdf:type
        assert!(query.contains("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"));

        // Should contain subject URI bind with ENCODE_FOR_URI
        assert!(query.contains("ENCODE_FOR_URI"));

        // Should contain IRI(CONCAT(...)) pattern
        assert!(query.contains("IRI(CONCAT"));

        // Should contain PROV-O provenance (full URI in CONSTRUCT template)
        assert!(query.contains("http://www.w3.org/ns/prov#wasGeneratedBy"));
        assert!(query.contains("activity"));

        // Should contain FK reference with target table
        assert!(query.contains("http://businessos.dev/id/organizations/"));

        // Should contain the property predicates
        assert!(query.contains("http://schema.org/name"));
        assert!(query.contains("http://schema.org/identifier"));
    }

    #[test]
    fn test_construct_query_prefixes() {
        let (mapping, prefixes) = sample_mapping();
        let query = generate_construct_query(&mapping, &prefixes);

        // Should have standard prefixes
        assert!(query.contains("PREFIX rdf:"));
        assert!(query.contains("PREFIX xsd:"));
        assert!(query.contains("PREFIX prov:"));
        assert!(query.contains("PREFIX bdev:"));
    }

    #[test]
    fn test_resolve_datatype() {
        assert_eq!(resolve_datatype("xsd:string"), "http://www.w3.org/2001/XMLSchema#string");
        assert_eq!(resolve_datatype("xsd:integer"), "http://www.w3.org/2001/XMLSchema#integer");
        assert_eq!(resolve_datatype("xsd:date"), "http://www.w3.org/2001/XMLSchema#date");
        assert_eq!(resolve_datatype("http://custom.org/type"), "http://custom.org/type");
    }

    #[test]
    fn test_sanitize_var_name() {
        assert_eq!(sanitize_var_name("first_name"), "first_name");
        assert_eq!(sanitize_var_name("user-id"), "user_id");
        assert_eq!(sanitize_var_name("col.name"), "col_name");
    }

    #[test]
    fn test_construct_generator_all() {
        let config_json = r#"{
            "prefixes": {"schema": "http://schema.org/"},
            "mappings": [{
                "table": "test_table",
                "ontology": "schema",
                "class": "Thing",
                "uri_template": "http://businessos.dev/id/test_table",
                "properties": [{
                    "column": "id",
                    "predicate": "schema:identifier",
                    "datatype": "xsd:integer",
                    "is_primary_key": true
                }]
            }],
            "relationships": []
        }"#;
        let config = MappingConfig::from_str(config_json).expect("parse");
        let generator = ConstructGenerator::new(&config);
        let queries = generator.generate_all().expect("generate_all");

        assert_eq!(queries.len(), 1);
        assert!(queries.contains_key("test_table"));
        let query = &queries["test_table"];
        assert!(query.contains("CONSTRUCT"));
        assert!(query.contains("test_table_uri"));
    }
}
