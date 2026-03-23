//! bos — BusinessOS data layer CLI.
//!
//! Noun-verb command structure for ODCS workspace operations,
//! schema management, data pipelines, decision records, and knowledge base.
//!
//! ## Examples
//!
//! ```bash
//! bos workspace init my-project
//! bos workspace validate ./my-project
//! bos schema convert schema.sql --to odc
//! bos decisions list
//! bos knowledge index ./docs
//! ```

use anyhow::Result;
use bos_core::decisions::DecisionGenerator;
use bos_ingest::DataSource;
use bos_core::knowledge::KnowledgeBase;
use bos_core::schema::{FormatHint, SchemaConverter};
use bos_core::workspace::{WorkspaceGenerator, WorkspaceInitOptions};
use bos_core::ontology::mapping::MappingConfig;
use bos_core::ontology::construct::ConstructGenerator;
use bos_core::ontology::execute::QueryExecutor;
use clap::Parser;
use std::path::Path;

/// bos — BusinessOS data layer CLI
#[derive(Parser, Debug)]
#[command(name = "bos")]
#[command(author = "Chatman Systems <sean@chatman.systems>")]
#[command(version = "0.1.0")]
#[command(about = "BusinessOS data layer — ODCS workspaces, schemas, and data pipelines")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Debug, clap::Subcommand)]
enum Commands {
    /// ODCS workspace operations
    Workspace {
        #[command(subcommand)]
        action: WorkspaceAction,
    },
    /// Schema validation and conversion
    Schema {
        #[command(subcommand)]
        action: SchemaAction,
    },
    /// Data import and export
    Data {
        #[command(subcommand)]
        action: DataAction,
    },
    /// MADR decision records
    Decisions {
        #[command(subcommand)]
        action: DecisionsAction,
    },
    /// Knowledge base management
    Knowledge {
        #[command(subcommand)]
        action: KnowledgeAction,
    },
    /// Validation and compliance
    Validate {
        /// Workspace directory to validate
        #[arg(short, long, default_value = ".")]
        workspace: String,
        /// Ruleset (soc2, hipaa)
        #[arg(short, long)]
        ruleset: Option<String>,
        /// Output as JSON
        #[arg(long)]
        json: bool,
    },
    /// Ontology bridge operations
    Ontology {
        #[command(subcommand)]
        action: OntologyAction,
    },
}

// --- Workspace ---

#[derive(Debug, clap::Subcommand)]
enum WorkspaceAction {
    /// Initialize a new ODCS workspace
    Init {
        /// Workspace name
        name: String,
        /// Description
        #[arg(short, long)]
        description: Option<String>,
        /// Output directory
        #[arg(short, long)]
        output: Option<String>,
    },
    /// Validate a workspace
    Validate {
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        path: String,
    },
    /// Export workspace to format
    Export {
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        path: String,
        /// Output format (odc, json)
        #[arg(short, long, default_value = "json")]
        format: String,
    },
}

// --- Schema ---

#[derive(Debug, clap::Subcommand)]
enum SchemaAction {
    /// Validate a schema file
    Validate {
        /// Schema file path
        path: String,
        /// Expected format (sql, json, avro, proto, odc)
        #[arg(short, long)]
        format: Option<String>,
    },
    /// Convert schema to ODCS format
    Convert {
        /// Input schema file
        input: String,
        /// Output file
        #[arg(short, long)]
        output: String,
        /// Source format hint
        #[arg(short, long)]
        from: Option<String>,
    },
}

// --- Data ---

#[derive(Debug, clap::Subcommand)]
enum DataAction {
    /// Import data from a source
    Import {
        /// Source file or directory
        #[arg(short, long)]
        source: String,
        /// Target workspace
        #[arg(short, long)]
        target: Option<String>,
    },
    /// Export data from workspace
    Export {
        /// Source workspace
        #[arg(short, long, default_value = ".")]
        source: String,
        /// Output format (odc, json)
        #[arg(short, long, default_value = "json")]
        format: String,
        /// Output file
        #[arg(short, long)]
        output: String,
    },
}

// --- Decisions ---

#[derive(Debug, clap::Subcommand)]
enum DecisionsAction {
    /// List all decision records
    List {
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        workspace: String,
    },
    /// Generate decision records from analysis
    Generate {
        /// Analysis file (mining.json, etc.)
        #[arg(short, long)]
        from: String,
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        workspace: String,
    },
    /// Export decisions as markdown or YAML
    Export {
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        workspace: String,
        /// Output format
        #[arg(short, long, default_value = "md")]
        format: String,
    },
}

// --- Knowledge ---

#[derive(Debug, clap::Subcommand)]
enum KnowledgeAction {
    /// Index knowledge articles from a directory
    Index {
        /// Directory to index
        directory: String,
    },
    /// Export knowledge base
    Export {
        /// Workspace directory
        #[arg(short, long, default_value = ".")]
        workspace: String,
    },
}

// --- Ontology ---

#[derive(Debug, clap::Subcommand)]
enum OntologyAction {
    /// Generate SPARQL CONSTRUCT queries from mapping config
    Construct {
        /// Mapping config file (JSON)
        #[arg(short, long)]
        mapping: String,
        /// Output directory for .rq files
        #[arg(short, long, default_value = ".obsr/queries")]
        output: String,
    },
    /// Export ODCS workspace with ontology mappings as RDF
    Export {
        /// Mapping config file (JSON)
        #[arg(short, long)]
        mapping: String,
        /// Output file path
        #[arg(short, long)]
        output: String,
        /// Output format (ttl, nt)
        #[arg(short, long, default_value = "ttl")]
        format: String,
    },
    /// Execute SPARQL CONSTRUCT query via oxigraph with PostgreSQL data
    Execute {
        /// Mapping config file (JSON)
        #[arg(short, long)]
        mapping: String,
        /// Database connection string (postgresql://...)
        #[arg(short, long, env = "DATABASE_URL")]
        database: String,
        /// Specific table to execute (omit for all)
        #[arg(short, long)]
        table: Option<String>,
        /// Output format (nt, ttl, json)
        #[arg(short, long, default_value = "nt")]
        format: String,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive("bos=info".parse().unwrap()),
        )
        .without_time()
        .init();

    let cli = Cli::parse();

    match cli.command {
        Commands::Workspace { action } => match action {
            WorkspaceAction::Init { name, description, output } => {
                let opts = WorkspaceInitOptions {
                    name,
                    description,
                    output_dir: output,
                };
                let dir = WorkspaceGenerator::init(&opts)?;
                eprintln!("==> Workspace created: {}", dir.display());
                Ok(())
            }
            WorkspaceAction::Validate { path } => {
                let path = Path::new(&path);
                let result = WorkspaceGenerator::validate(path)?;
                if result.is_valid {
                    eprintln!("==> VALID: {} ({} tables, {} relationships)",
                        result.workspace_path, result.tables, result.relationships);
                } else {
                    eprintln!("==> INVALID: {}", result.workspace_path);
                    for err in &result.errors {
                        eprintln!("    ERROR: {}", err);
                    }
                }
                for warn in &result.warnings {
                    eprintln!("    WARN: {}", warn);
                }
                Ok(())
            }
            WorkspaceAction::Export { path, format } => {
                let path = Path::new(&path);
                let result = WorkspaceGenerator::export(path, &format)?;
                eprintln!("==> Exported to {} ({} tables, format={})",
                    result.output_path, result.tables_exported, result.format);
                Ok(())
            }
        },

        Commands::Schema { action } => match action {
            SchemaAction::Validate { path, format } => {
                let path = Path::new(&path);
                let hint = format.as_deref().and_then(FormatHint::from_str_lossy);
                let result = SchemaConverter::validate(path, hint)?;
                if result.is_valid {
                    eprintln!("==> VALID: {} (format={}, {} tables, {} columns)",
                        result.path, result.format.as_deref().unwrap_or("?"),
                        result.tables, result.columns);
                } else {
                    eprintln!("==> INVALID: {}", result.path);
                    for err in &result.errors {
                        eprintln!("    ERROR: {}", err);
                    }
                }
                Ok(())
            }
            SchemaAction::Convert { input, output, from } => {
                let input_path = Path::new(&input);
                let hint = from.as_deref().and_then(FormatHint::from_str_lossy);
                let result = SchemaConverter::convert(input_path, hint)?;
                std::fs::write(&output, &result.odcs_output)?;
                eprintln!("==> Converted {} → {} (format: {}, {} tables, {} columns, {}ms)",
                    result.source_path, output, result.detected_format,
                    result.table_count, result.column_count, result.conversion_time_ms);
                Ok(())
            }
        },

        Commands::Data { action } => match action {
            DataAction::Import { source, target: _ } => {
                let source = bos_ingest::sources::FileSource::new(&source);
                let rows: Vec<bos_ingest::DataRow> = source.read().await?;
                eprintln!("==> Imported {} rows from {}", rows.len(), source.path.display());
                Ok(())
            }
            DataAction::Export { source, format, output } => {
                let source_path = Path::new(&source);
                let content = std::fs::read_to_string(source_path)?;
                let val: serde_json::Value = serde_json::from_str(&content)?;
                let result = match format.as_str() {
                    "odc" | "yaml" => bos_core::export::ExportManager::to_odcs(&val, &output)?,
                    _ => bos_core::export::ExportManager::to_json(&val, &output)?,
                };
                eprintln!("==> Exported to {} (format={})", result.output_path, result.format);
                Ok(())
            }
        },

        Commands::Decisions { action } => match action {
            DecisionsAction::List { workspace } => {
                let path = Path::new(&workspace);
                let index = DecisionGenerator::list(path)?;
                eprintln!("==> {} decisions in {}", index.total_decisions, index.workspace);
                for d in &index.decisions {
                    eprintln!("    ADR-{:03} [{}] {} ({})",
                        d.number, d.status, d.title, d.category);
                }
                if !index.categories.is_empty() {
                    eprintln!("\n    Categories:");
                    for (cat, count) in &index.categories {
                        eprintln!("      {}: {}", cat, count);
                    }
                }
                Ok(())
            }
            DecisionsAction::Generate { from, workspace } => {
                eprintln!("==> Generating decisions from {} for {}...", from, workspace);
                eprintln!("    (analysis file parsing not yet implemented)");
                Ok(())
            }
            DecisionsAction::Export { workspace, format } => {
                let path = Path::new(&workspace);
                let md = DecisionGenerator::export_markdown(path)?;
                let ext = match format.as_str() {
                    "yaml" => "decisions.yaml",
                    _ => "decisions.md",
                };
                let output = Path::new(&workspace).join(ext);
                std::fs::write(&output, &md)?;
                eprintln!("==> Exported {} decisions to {}",
                    DecisionGenerator::list(path).map(|i| i.total_decisions).unwrap_or(0),
                    output.display());
                Ok(())
            }
        },

        Commands::Knowledge { action } => match action {
            KnowledgeAction::Index { directory } => {
                let path = Path::new(&directory);
                let index = KnowledgeBase::index(path)?;
                eprintln!("==> Indexed {} articles in {}", index.total_articles, index.workspace);
                for (t, count) in &index.types {
                    eprintln!("    {}: {}", t, count);
                }
                Ok(())
            }
            KnowledgeAction::Export { workspace } => {
                let path = Path::new(&workspace);
                let md = KnowledgeBase::export(path)?;
                let output = Path::new(&workspace).join("knowledge.md");
                std::fs::write(&output, &md)?;
                eprintln!("==> Exported knowledge base to {}", output.display());
                Ok(())
            }
        },

        Commands::Validate { workspace, ruleset, json } => {
            let path = Path::new(&workspace);
            let result = WorkspaceGenerator::validate(path)?;

            if json {
                let output = serde_json::to_string_pretty(&result)?;
                println!("{}", output);
            } else {
                eprintln!("==> bos validate {}", workspace);
                if let Some(rs) = &ruleset {
                    eprintln!("    Ruleset: {}", rs);
                }
                if result.is_valid {
                    eprintln!("    Result: PASS ({} tables, {} relationships)",
                        result.tables, result.relationships);
                } else {
                    eprintln!("    Result: FAIL");
                    for err in &result.errors {
                        eprintln!("      ERROR: {}", err);
                    }
                }
                for warn in &result.warnings {
                    eprintln!("      WARN: {}", warn);
                }
            }
            Ok(())
        }

        Commands::Ontology { action } => match action {
            OntologyAction::Construct { mapping, output } => {
                let mapping_path = Path::new(&mapping);
                let config = MappingConfig::from_file(mapping_path)
                    .map_err(|e| anyhow::anyhow!("Failed to load mapping: {e}"))?;
                let generator = ConstructGenerator::new(&config);
                let queries = generator.generate_all()
                    .map_err(|e| anyhow::anyhow!("Failed to generate queries: {e}"))?;

                std::fs::create_dir_all(&output)
                    .map_err(|e| anyhow::anyhow!("Failed to create output dir: {e}"))?;

                for (table, query) in &queries {
                    let filename = Path::new(&output).join(format!("{}.rq", table));
                    std::fs::write(&filename, query)
                        .map_err(|e| anyhow::anyhow!("Failed to write query: {e}"))?;
                }

                eprintln!("==> Generated {} CONSTRUCT queries", queries.len());
                eprintln!("    Mapping: {}", mapping);
                eprintln!("    Output: {}/", output);
                for table in queries.keys() {
                    eprintln!("      - {}.rq", table);
                }
                Ok(())
            }

            OntologyAction::Export { mapping, output, format } => {
                let mapping_path = Path::new(&mapping);
                let config = MappingConfig::from_file(mapping_path)
                    .map_err(|e| anyhow::anyhow!("Failed to load mapping: {e}"))?;
                let generator = ConstructGenerator::new(&config);
                let queries = generator.generate_all()
                    .map_err(|e| anyhow::anyhow!("Failed to generate queries: {e}"))?;

                // Export as Turtle by default — concatenate all CONSTRUCT queries
                // into a single file with prefix block
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
                    .map_err(|e| anyhow::anyhow!("Failed to write export: {e}"))?;

                eprintln!("==> Exported {} tables as {} to {}", queries.len(), format, output);
                Ok(())
            }

            OntologyAction::Execute { mapping, database, table, format } => {
                let mapping_path = Path::new(&mapping);
                let config = MappingConfig::from_file(mapping_path)
                    .map_err(|e| anyhow::anyhow!("Failed to load mapping: {e}"))?;

                // Use spawn_blocking because postgres crate is synchronous
                // and cannot run inside a tokio runtime
                if let Some(table_name) = table {
                    let config_clone = config.clone();
                    let db_clone = database.clone();
                    let tbl_name = table_name.clone();
                    let result = tokio::task::spawn_blocking(move || {
                        let executor = QueryExecutor::new(config_clone, db_clone);
                        executor.execute_table(&tbl_name, None)
                    }).await.map_err(|e| anyhow::anyhow!("Task join error: {e}"))?;

                    match result {
                        Ok(result) => {
                            eprintln!("==> Executed table: {}", table_name);
                            eprintln!("    Rows loaded: {}", result.rows_loaded);
                            eprintln!("    Triples generated: {}", result.triples_generated);
                            eprintln!("    CONSTRUCT triples: {}", result.construct_triples);

                            match format.as_str() {
                                "ttl" => {
                                    println!("{}", result.ntriples);
                                }
                                "json" => {
                                    let json_output = serde_json::json!({
                                        "table": table_name,
                                        "rows_loaded": result.rows_loaded,
                                        "triples_generated": result.triples_generated,
                                        "construct_triples": result.construct_triples,
                                        "ntriples": result.ntriples,
                                    });
                                    println!("{}", serde_json::to_string_pretty(&json_output)?);
                                }
                                _ => {
                                    println!("{}", result.ntriples);
                                }
                            }
                        }
                        Err(e) => {
                            eprintln!("==> ERROR: Failed to execute table '{}': {}", table_name, e);
                            std::process::exit(1);
                        }
                    }
                } else {
                    // Execute all tables
                    eprintln!("==> Executing all mapped tables against PostgreSQL...");
                    let config_clone = config.clone();
                    let db_clone = database.clone();
                    let results = tokio::task::spawn_blocking(move || {
                        let executor = QueryExecutor::new(config_clone, db_clone);
                        executor.execute_all()
                    }).await.map_err(|e| anyhow::anyhow!("Task join error: {e}"))?
                        .map_err(|e| anyhow::anyhow!("Execution failed: {e}"))?;

                    let mut total_rows = 0usize;
                    let mut total_construct = 0usize;

                    for (tbl, result) in &results {
                        eprintln!("    {}: {} rows, {} triples", tbl, result.rows_loaded, result.construct_triples);
                        total_rows += result.rows_loaded;
                        total_construct += result.construct_triples;
                    }

                    eprintln!("==> Total: {} rows, {} CONSTRUCT triples across {} tables",
                        total_rows, total_construct, results.len());

                    // Output combined N-Triples
                    match format.as_str() {
                        "ttl" => {
                            for (_tbl, result) in &results {
                                println!("{}", result.ntriples);
                            }
                        }
                        "json" => {
                            let tables_json: serde_json::Value = results.iter()
                                .map(|(tbl, r)| {
                                    serde_json::json!({
                                        "table": tbl,
                                        "rows_loaded": r.rows_loaded,
                                        "triples_generated": r.triples_generated,
                                        "construct_triples": r.construct_triples,
                                    })
                                })
                                .collect();
                            let output = serde_json::json!({
                                "total_rows": total_rows,
                                "total_construct_triples": total_construct,
                                "tables": tables_json,
                            });
                            println!("{}", serde_json::to_string_pretty(&output)?);
                        }
                        _ => {
                            for (tbl, result) in &results {
                                if !result.ntriples.is_empty() {
                                    eprintln!("# --- {} ---", tbl);
                                    println!("{}", result.ntriples);
                                }
                            }
                        }
                    }
                }
                Ok(())
            }
        },
    }
}
