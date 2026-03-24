use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct QueriesGenerated {
    pub mapping: String,
    pub output_dir: String,
    pub queries: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct OntologyExported {
    pub tables: usize,
    pub format: String,
    pub output_path: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct TableExecutionResult {
    pub table: String,
    pub rows_loaded: usize,
    pub triples_generated: usize,
    pub construct_triples: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct OntologyExecuted {
    pub total_rows: usize,
    pub total_construct_triples: usize,
    pub tables: Vec<TableExecutionResult>,
}

fn execute_single_table(
    config: bos_core::ontology::mapping::MappingConfig,
    database: String,
    table_name: String,
) -> Result<TableExecutionResult> {
    let executor = bos_core::ontology::execute::QueryExecutor::new(config, database);
    let result = executor.execute_table(&table_name, None)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;
    Ok(TableExecutionResult {
        table: table_name,
        rows_loaded: result.rows_loaded,
        triples_generated: result.triples_generated,
        construct_triples: result.construct_triples,
    })
}

fn execute_all_tables(
    config: bos_core::ontology::mapping::MappingConfig,
    database: String,
) -> Result<Vec<TableExecutionResult>> {
    let executor = bos_core::ontology::execute::QueryExecutor::new(config, database);
    let results = executor.execute_all()
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let mut tables: Vec<_> = results.keys().cloned().collect();
    tables.sort();

    let mut out = Vec::new();
    for tbl in &tables {
        let r = &results[tbl];
        out.push(TableExecutionResult {
            table: tbl.clone(),
            rows_loaded: r.rows_loaded,
            triples_generated: r.triples_generated,
            construct_triples: r.construct_triples,
        });
    }
    Ok(out)
}

#[noun("ontology", "Ontology bridge operations")]

/// Generate SPARQL CONSTRUCT queries from mapping config
///
/// # Arguments
/// * `mapping` - Mapping config file (JSON)
/// * `output` - Output directory for .rq files [default: .obsr/queries]
#[verb("construct")]
fn construct(
    mapping: String,
    output: Option<String>,
) -> Result<QueriesGenerated> {
    let output_dir = output.unwrap_or_else(|| ".obsr/queries".to_string());
    let mapping_path = std::path::Path::new(&mapping);
    let config = bos_core::ontology::mapping::MappingConfig::from_file(mapping_path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to load mapping: {e}")))?;
    let generator = bos_core::ontology::construct::ConstructGenerator::new(&config);
    let queries = generator.generate_all()
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to generate queries: {e}")))?;

    std::fs::create_dir_all(&output_dir)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to create output dir: {e}")))?;

    let mut table_names: Vec<String> = queries.keys().cloned().collect();
    table_names.sort();

    for (table, query) in &queries {
        let filename = std::path::Path::new(&output_dir).join(format!("{}.rq", table));
        std::fs::write(&filename, query)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to write query: {e}")))?;
    }

    Ok(QueriesGenerated {
        mapping,
        output_dir,
        queries: table_names,
    })
}

/// Export ODCS workspace with ontology mappings as RDF
///
/// # Arguments
/// * `mapping` - Mapping config file (JSON)
/// * `output` - Output file path
/// * `format` - Output format (ttl, nt) [default: ttl]
#[verb("export")]
fn export(
    mapping: String,
    output: String,
    format: Option<String>,
) -> Result<OntologyExported> {
    let fmt = format.unwrap_or_else(|| "ttl".to_string());
    let mapping_path = std::path::Path::new(&mapping);
    let config = bos_core::ontology::mapping::MappingConfig::from_file(mapping_path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to load mapping: {e}")))?;
    let generator = bos_core::ontology::construct::ConstructGenerator::new(&config);
    let queries = generator.generate_all()
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to generate queries: {e}")))?;

    let mut turtle = String::new();
    turtle.push_str("# BusinessOS Ontology Export\n");
    turtle.push_str(&format!("# Generated: {}\n", chrono::Utc::now().format("%Y-%m-%dT%H:%M:%SZ")));
    turtle.push_str(&format!("# Tables: {}\n\n", queries.len()));

    for (table, query) in &queries {
        turtle.push_str(&format!("# --- {} ---\n", table));
        turtle.push_str(query);
        turtle.push_str("\n\n");
    }

    std::fs::write(&output, &turtle)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to write export: {e}")))?;

    Ok(OntologyExported {
        tables: queries.len(),
        format: fmt,
        output_path: output,
    })
}

/// Execute SPARQL CONSTRUCT query via oxigraph with PostgreSQL data
///
/// # Arguments
/// * `mapping` - Mapping config file (JSON)
/// * `database` - Database connection string [env: DATABASE_URL]
/// * `table` - Specific table to execute (omit for all) [hide]
/// * `format` - Output format (nt, ttl, json) [default: nt]
#[verb("execute")]
fn execute(
    mapping: String,
    database: String,
    table: Option<String>,
    _format: Option<String>,
) -> Result<OntologyExecuted> {
    let mapping_path = std::path::Path::new(&mapping);
    let config = bos_core::ontology::mapping::MappingConfig::from_file(mapping_path)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Failed to load mapping: {e}")))?;

    clap_noun_verb::async_verb::run_async(async move {
        let tables = if let Some(table_name) = table {
            let config_clone = config.clone();
            let db_clone = database.clone();
            let tbl = table_name.clone();
            let r = tokio::task::spawn_blocking(move || {
                execute_single_table(config_clone, db_clone, tbl)
            }).await
                .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Task join error: {e}")))??;
            vec![r]
        } else {
            let config_clone = config.clone();
            let db_clone = database.clone();
            tokio::task::spawn_blocking(move || {
                execute_all_tables(config_clone, db_clone)
            }).await
                .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Task join error: {e}")))?
                .map_err(|e| clap_noun_verb::NounVerbError::execution_error(format!("Execution failed: {e}")))?
        };

        let total_rows: usize = tables.iter().map(|t| t.rows_loaded).sum();
        let total_construct: usize = tables.iter().map(|t| t.construct_triples).sum();

        Ok(OntologyExecuted {
            total_rows,
            total_construct_triples: total_construct,
            tables,
        })
    })
}
