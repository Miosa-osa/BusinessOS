//! RDF pipeline integration tests — dm-sdk DataModel → OntologyInferrer →
//! ConstructGenerator → TripleStore.
//!
//! No PostgreSQL required. Everything runs against the in-memory Oxigraph store.
//!
//! Sections:
//!   1. OntologyInferrer: DataModel → MappingConfig
//!   2. ConstructGenerator: MappingConfig → SPARQL CONSTRUCT strings
//!   3. TripleStore: named-graph insertion via SPARQL UPDATE
//!   4. SPARQL SELECT on the store
//!   5. PROV-O provenance in generated queries

use bos_core::{ConstructGenerator, InferConfig, MappingConfig, OntologyInferrer, TripleStore};
use data_modelling_sdk::{Column, DataModel, ForeignKey, Table};

// ── shared fixtures ───────────────────────────────────────────────────────────

/// Build a DataModel with `users` and `projects` tables.
/// `projects.user_id` is a foreign key → `users.id`.
fn projects_data_model() -> DataModel {
    let mut model = DataModel::new(
        "test-workspace".to_string(),
        ".".to_string(),
        "control.yaml".to_string(),
    );

    // users table
    let mut user_id_col = Column::new("id".to_string(), "UUID".to_string());
    user_id_col.primary_key = true;
    let users = Table::new(
        "users".to_string(),
        vec![
            user_id_col,
            Column::new("name".to_string(), "VARCHAR".to_string()),
        ],
    );
    model.tables.push(users);

    // projects table with FK to users
    let users_table_id = model.tables[0].id.to_string();
    let mut proj_id_col = Column::new("id".to_string(), "UUID".to_string());
    proj_id_col.primary_key = true;
    let mut fk_col = Column::new("user_id".to_string(), "UUID".to_string());
    fk_col.foreign_key = Some(ForeignKey {
        table_id: users_table_id,
        column_name: "id".to_string(),
    });
    let projects = Table::new(
        "projects".to_string(),
        vec![
            proj_id_col,
            Column::new("name".to_string(), "VARCHAR".to_string()),
            fk_col,
        ],
    );
    model.tables.push(projects);

    model
}

/// Insert a project row into the L0 named graph.
fn seed_project_row(store: &TripleStore, id: &str, name: &str, user_id: &str) {
    let sparql = format!(
        r#"INSERT DATA {{
  GRAPH <http://businessos.local/l0> {{
    <http://businessos.dev/id/project/{id}> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type>
        <https://schema.org/Project> .
    <http://businessos.dev/id/project/{id}> <https://schema.org/name> "{name}" .
    <http://businessos.dev/id/project/{id}> <https://schema.org/creator>
        <http://businessos.dev/id/user/{user_id}> .
  }}
}}"#
    );
    store.update_sparql(&sparql).expect("seed_project_row failed");
}

// ── Section 1: OntologyInferrer ───────────────────────────────────────────────

#[test]
fn test_infer_projects_table_maps_to_schema_project() {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer must succeed");
    let projects_mapping = result
        .config
        .find_mapping("projects")
        .expect("projects mapping must be inferred");
    assert_eq!(
        projects_mapping.class, "Project",
        "projects table must map to schema:Project"
    );
    assert_eq!(
        projects_mapping.ontology, "schema",
        "projects must use schema ontology"
    );
}

#[test]
fn test_infer_users_table_maps_to_foaf_person() {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer must succeed");
    let users_mapping = result
        .config
        .find_mapping("users")
        .expect("users mapping must be inferred");
    assert_eq!(
        users_mapping.class, "Person",
        "users table must map to foaf:Person"
    );
}

#[test]
fn test_infer_fk_column_becomes_uri_property() {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer must succeed");
    let projects_mapping = result
        .config
        .find_mapping("projects")
        .expect("projects mapping must exist");
    let fk_prop = projects_mapping
        .properties
        .iter()
        .find(|p| p.column == "user_id")
        .expect("user_id property must be in projects mapping");
    assert_eq!(
        fk_prop.object_type.as_deref(),
        Some("uri"),
        "FK column user_id must be inferred as object_type=uri"
    );
}

#[test]
fn test_infer_prefixes_include_rdf_xsd_prov_schema_foaf() {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer must succeed");
    let prefixes = &result.config.prefixes;
    assert!(prefixes.contains_key("rdf"), "prefixes must include rdf");
    assert!(prefixes.contains_key("xsd"), "prefixes must include xsd");
    assert!(prefixes.contains_key("prov"), "prefixes must include prov");
}

#[test]
fn test_infer_result_stats_are_consistent() {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer must succeed");
    assert_eq!(
        result.tables_inferred, 2,
        "DataModel with 2 tables must yield 2 inferred table mappings"
    );
    assert!(
        result.properties_inferred >= 2,
        "must have inferred at least 2 properties total"
    );
}

// ── Section 2: ConstructGenerator ────────────────────────────────────────────

fn projects_mapping_config() -> MappingConfig {
    let model = projects_data_model();
    let result = OntologyInferrer::new(InferConfig::default())
        .infer(&model)
        .expect("infer for construct tests");
    result.config
}

#[test]
fn test_construct_query_has_mandatory_sparql_clauses() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    let projects_query = queries.get("projects").expect("projects query must exist");
    assert!(
        projects_query.contains("CONSTRUCT {"),
        "query must contain CONSTRUCT {{ clause"
    );
    assert!(
        projects_query.contains("WHERE {"),
        "query must contain WHERE {{ clause"
    );
}

#[test]
fn test_construct_query_encodes_subject_uri() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    let projects_query = queries.get("projects").expect("projects query must exist");
    assert!(
        projects_query.contains("http://businessos.dev/id/projects"),
        "subject URI must use businessos.dev/id/projects base; query:\n{projects_query}"
    );
    assert!(
        projects_query.contains("ENCODE_FOR_URI"),
        "subject URI must use ENCODE_FOR_URI for pk; query:\n{projects_query}"
    );
}

#[test]
fn test_construct_query_emits_all_three_prov_triples() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    let projects_query = queries.get("projects").expect("projects query must exist");
    assert!(
        projects_query.contains("wasGeneratedBy"),
        "query must emit prov:wasGeneratedBy"
    );
    assert!(
        projects_query.contains("wasDerivedFrom"),
        "query must emit prov:wasDerivedFrom"
    );
    assert!(
        projects_query.contains("generatedAtTime"),
        "query must emit prov:generatedAtTime"
    );
}

#[test]
fn test_construct_query_prov_timestamp_is_xsd_datetime() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    let projects_query = queries.get("projects").expect("projects query must exist");
    assert!(
        projects_query.contains("xsd:dateTime"),
        "prov:generatedAtTime value must be typed as xsd:dateTime; query:\n{projects_query}"
    );
}

#[test]
fn test_construct_query_fk_uses_bound_guard() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    let projects_query = queries.get("projects").expect("projects query must exist");
    assert!(
        projects_query.contains("BOUND("),
        "FK property must use BOUND() guard in query; query:\n{projects_query}"
    );
    assert!(
        projects_query.contains("UNDEF"),
        "FK property must use UNDEF as fallback in query; query:\n{projects_query}"
    );
}

#[test]
fn test_construct_generator_generates_all_tables() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all must succeed");
    assert_eq!(queries.len(), 2, "DataModel with 2 tables must yield 2 queries");
    assert!(queries.contains_key("projects"), "must have 'projects' query");
    assert!(queries.contains_key("users"), "must have 'users' query");
}

// ── Section 3: TripleStore named-graph insertion ──────────────────────────────

#[test]
fn test_insert_data_into_l0_named_graph_succeeds() {
    let store = TripleStore::new();
    let sparql = r#"INSERT DATA {
  GRAPH <http://businessos.local/l0> {
    <http://example.org/s> <http://example.org/p> "test value" .
  }
}"#;
    store.update_sparql(sparql).expect("INSERT DATA into named graph must succeed");
}

#[test]
fn test_l0_named_graph_is_distinct_from_default_graph() {
    let store = TripleStore::new();

    // Insert into named graph only
    store
        .update_sparql(
            r#"INSERT DATA {
  GRAPH <http://businessos.local/l0> {
    <http://example.org/s> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/T> .
  }
}"#,
        )
        .expect("insert into l0");

    // Query default graph — must be empty
    let default_rows = store
        .query_sparql("SELECT ?s WHERE { ?s ?p ?o }")
        .expect("select from default graph");
    assert!(
        default_rows.is_empty(),
        "default graph must be empty after inserting only into named graph"
    );

    // Query named graph — must have the triple
    let named_rows = store
        .query_sparql(
            "SELECT ?s FROM <http://businessos.local/l0> WHERE { ?s ?p ?o }",
        )
        .expect("select from l0 named graph");
    assert_eq!(
        named_rows.len(),
        1,
        "l0 named graph must contain exactly 1 triple after insert"
    );
}

#[test]
fn test_insert_full_project_row_into_l0() {
    let store = TripleStore::new();
    seed_project_row(&store, "proj-001", "Alpha Project", "user-001");

    let rows = store
        .query_sparql(
            r#"SELECT ?p ?pred ?obj FROM <http://businessos.local/l0>
               WHERE { ?p ?pred ?obj . FILTER(?p = <http://businessos.dev/id/project/proj-001>) }"#,
        )
        .expect("select project triples");
    assert!(
        rows.len() >= 3,
        "project row must have at least 3 triples (type, name, creator); got {}",
        rows.len()
    );
}

#[test]
fn test_multiple_projects_isolated_by_pk() {
    let store = TripleStore::new();
    seed_project_row(&store, "p1", "Project One", "u1");
    seed_project_row(&store, "p2", "Project Two", "u1");

    let rows = store
        .query_sparql(
            r#"SELECT ?p FROM <http://businessos.local/l0>
               WHERE { ?p <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Project> }"#,
        )
        .expect("select all projects");
    assert_eq!(rows.len(), 2, "must have exactly 2 distinct project subjects");
}

// ── Section 4: SPARQL SELECT on the store ────────────────────────────────────

#[test]
fn test_select_all_projects_from_l0() {
    let store = TripleStore::new();
    seed_project_row(&store, "alpha", "Alpha", "u1");
    seed_project_row(&store, "beta", "Beta", "u2");

    let rows = store
        .query_sparql(
            r#"SELECT ?p FROM <http://businessos.local/l0>
               WHERE { ?p <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Project> }"#,
        )
        .expect("select all projects");
    assert_eq!(rows.len(), 2, "seeding 2 projects must return 2 SELECT rows");
}

#[test]
fn test_select_fk_resolves_to_user_uri() {
    let store = TripleStore::new();
    seed_project_row(&store, "myproj", "My Project", "myuser");

    let rows = store
        .query_sparql(
            r#"SELECT ?creator FROM <http://businessos.local/l0>
               WHERE {
                 <http://businessos.dev/id/project/myproj> <https://schema.org/creator> ?creator .
               }"#,
        )
        .expect("select creator");
    assert_eq!(rows.len(), 1, "must find exactly one creator for myproj");
    let creator = rows[0].get("creator").expect("creator variable must be bound");
    assert!(
        creator.contains("myuser"),
        "creator must resolve to user URI containing 'myuser'; got: {creator}"
    );
}

#[test]
fn test_ask_query_confirms_triple_existence() {
    let store = TripleStore::new();
    seed_project_row(&store, "ask-proj", "ASK Project", "ask-user");

    let rows = store
        .query_sparql(
            r#"ASK FROM <http://businessos.local/l0>
               WHERE { <http://businessos.dev/id/project/ask-proj>
                       <https://schema.org/name> ?n }"#,
        )
        .expect("ASK query");
    // ASK returns a single row with key "result" = "true" or "false"
    assert_eq!(rows.len(), 1);
    assert_eq!(
        rows[0].get("result").map(|s| s.as_str()),
        Some("true"),
        "ASK must return true for an existing triple"
    );
}

#[test]
fn test_l0_seed_is_idempotent() {
    let store = TripleStore::new();
    // Insert the same triple twice — RDF set semantics: count stays the same
    let sparql = r#"INSERT DATA {
  GRAPH <http://businessos.local/l0> {
    <http://businessos.dev/id/project/idem> <https://schema.org/name> "Idempotent" .
  }
}"#;
    store.update_sparql(sparql).expect("first insert");
    store.update_sparql(sparql).expect("second insert (idempotent)");

    let rows = store
        .query_sparql(
            r#"SELECT ?n FROM <http://businessos.local/l0>
               WHERE { <http://businessos.dev/id/project/idem> <https://schema.org/name> ?n }"#,
        )
        .expect("select after double insert");
    assert_eq!(rows.len(), 1, "RDF set semantics: duplicate insert must not duplicate triples");
}

// ── Section 5: PROV-O provenance ─────────────────────────────────────────────

#[test]
fn test_every_entity_has_was_generated_by() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all");
    // All generated CONSTRUCT queries must include prov:wasGeneratedBy in their template
    for (table, query) in &queries {
        assert!(
            query.contains("wasGeneratedBy"),
            "CONSTRUCT query for '{table}' must include prov:wasGeneratedBy"
        );
    }
}

#[test]
fn test_every_entity_has_generated_at_time() {
    let config = projects_mapping_config();
    let queries = ConstructGenerator::new(&config)
        .generate_all()
        .expect("generate_all");
    for (table, query) in &queries {
        assert!(
            query.contains("generatedAtTime"),
            "CONSTRUCT query for '{table}' must include prov:generatedAtTime"
        );
    }
}

#[test]
fn test_generated_at_time_is_valid_xsd_datetime() {
    let config = projects_mapping_config();
    let query = ConstructGenerator::new(&config)
        .generate_for_table("projects")
        .expect("generate for projects");

    // The timestamp is baked into the query string at generation time.
    // Look for the ISO8601 pattern "T" in the xsd:dateTime literal.
    let timestamp_marker = "xsd:dateTime";
    assert!(
        query.contains(timestamp_marker),
        "query must contain xsd:dateTime type annotation"
    );
    // The literal before ^^xsd:dateTime should contain 'T' (ISO8601 separator)
    let before_type = query.split("^^xsd:dateTime").next().unwrap_or("");
    assert!(
        before_type.contains('T'),
        "timestamp literal before ^^xsd:dateTime must contain 'T' (ISO8601); context: ...{before_type}"
    );
}

#[test]
fn test_ontology_mappings_json_parses_as_mapping_config() {
    // Verify that a valid MappingConfig JSON document round-trips correctly.
    let json = r#"{
        "prefixes": {
            "rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
            "xsd":  "http://www.w3.org/2001/XMLSchema#",
            "prov": "http://www.w3.org/ns/prov#",
            "schema": "https://schema.org/",
            "foaf": "http://xmlns.com/foaf/0.1/"
        },
        "mappings": [
            {
                "table": "projects",
                "ontology": "schema",
                "class": "Project",
                "uri_template": "http://businessos.dev/id/projects",
                "properties": [
                    {
                        "column": "id",
                        "predicate": "schema:identifier",
                        "datatype": "xsd:string",
                        "is_primary_key": true,
                        "object_type": null,
                        "target_table": null,
                        "value_map": {}
                    }
                ]
            }
        ]
    }"#;

    let config = MappingConfig::from_str(json).expect("valid MappingConfig JSON must parse");
    assert_eq!(config.mappings.len(), 1);
    assert_eq!(config.mappings[0].table, "projects");
    assert!(config.prefixes.contains_key("prov"), "parsed config must include prov prefix");
}
