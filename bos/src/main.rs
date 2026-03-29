//! bosctl — BusinessOS CLI binary.
//!
//! Flat clap-derive CLI mapping 22 subcommands to BosCommand variants.

use std::path::PathBuf;
use std::time::Duration;

use clap::{Parser, Subcommand, ValueEnum};

use bos_commands::{
    BosCommand, BosCommandHandler, ResultFormatter,
    commands::businessos_commands::{
        BatchArgs, BatchConformArgs, BatchDiscoverArgs, ConformArgs, DiscoverArgs, ExportArgs,
        FingerprintArgs, ImportArgs, ListModelsArgs, OntologyArgs, OrgEvolutionArgs,
        QualityCheckArgs, StatisticsArgs, ValidateModelArgs, VariabilityArgs,
        VariantAnalysisArgs,
    },
};
use bos_core::GatewayConfig;

/// bosctl — BusinessOS command-line interface.
#[derive(Parser)]
#[command(name = "bosctl", version, about = "BusinessOS CLI", long_about = None)]
struct Cli {
    /// Output format
    #[arg(short = 'f', long, value_enum, default_value = "pretty", global = true)]
    format: CliFormat,

    /// BusinessOS gateway URL
    #[arg(long, env = "BOS_GATEWAY_URL", default_value = "http://localhost:8001", global = true)]
    gateway: String,

    /// Gateway request timeout in milliseconds
    #[arg(long, env = "BOS_GATEWAY_TIMEOUT_MS", default_value = "10000", global = true)]
    timeout_ms: u64,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Clone, Copy, ValueEnum)]
enum CliFormat {
    Json,
    Pretty,
    Table,
    Csv,
}

impl From<CliFormat> for bos_commands::OutputFormat {
    fn from(f: CliFormat) -> Self {
        match f {
            CliFormat::Json => bos_commands::OutputFormat::Json,
            CliFormat::Pretty => bos_commands::OutputFormat::PrettyJson,
            CliFormat::Table => bos_commands::OutputFormat::Table,
            CliFormat::Csv => bos_commands::OutputFormat::Csv,
        }
    }
}

#[derive(Subcommand)]
enum Commands {
    /// Discover process model from event log (via BusinessOS gateway)
    Discover {
        /// Path to event log
        log_path: PathBuf,
        /// Discovery algorithm: inductive, alpha, heuristic
        #[arg(long)]
        algorithm: Option<String>,
        /// Max traces to process
        #[arg(long)]
        max_traces: Option<usize>,
        /// Filter by activity pattern
        #[arg(long)]
        activity_filter: Option<String>,
        /// Output model ID
        #[arg(long)]
        model_id: Option<String>,
    },
    /// Check conformance of event log against model (via BusinessOS gateway)
    Conform {
        /// Path to event log
        log_path: PathBuf,
        /// Model ID
        model_id: String,
        /// Alignment strategy
        #[arg(long)]
        alignment: Option<String>,
    },
    /// Extract statistics from event log (via BusinessOS gateway)
    Statistics {
        /// Path to event log
        log_path: PathBuf,
        #[arg(long)]
        with_variants: Option<bool>,
        #[arg(long)]
        with_activities: Option<bool>,
        #[arg(long)]
        with_durations: Option<bool>,
    },
    /// Check conformance (alias for conform)
    CheckConformance {
        log_path: PathBuf,
        model_id: String,
        #[arg(long)]
        alignment: Option<String>,
    },
    /// Discover batch (requires batch runner — not yet implemented)
    DiscoverBatch {
        config_path: PathBuf,
        #[arg(long)]
        parallel: Option<bool>,
        #[arg(long)]
        workers: Option<usize>,
    },
    /// List process models (requires model registry — not yet implemented)
    ListModels {
        #[arg(long)]
        algorithm_filter: Option<String>,
        #[arg(long)]
        date_from: Option<String>,
        #[arg(long)]
        date_to: Option<String>,
        #[arg(long)]
        sort_by: Option<String>,
    },
    /// Validate process model (requires model registry — not yet implemented)
    ValidateModel {
        model_id: String,
        #[arg(long)]
        check_soundness: Option<bool>,
        #[arg(long)]
        check_liveness: Option<bool>,
    },
    /// Quality check on data or log file
    QualityCheck {
        data_path: PathBuf,
        #[arg(long)]
        metrics: Option<Vec<String>>,
        #[arg(long)]
        report: Option<bool>,
    },
    /// Calculate trace fingerprint from event log
    Fingerprint {
        log_path: PathBuf,
        #[arg(long)]
        baseline_model: Option<String>,
        #[arg(long)]
        algorithm: Option<String>,
    },
    /// Analyze process variability in event log
    Variability {
        log_path: PathBuf,
        #[arg(long)]
        baseline_variant: Option<String>,
        #[arg(long)]
        variance_threshold: Option<f64>,
    },
    /// Analyze organizational evolution over time
    OrgEvolution {
        log_path: PathBuf,
        #[arg(long)]
        start_date: Option<String>,
        #[arg(long)]
        end_date: Option<String>,
        #[arg(long)]
        granularity: Option<String>,
    },
    /// Analyze process variants in event log
    VariantAnalysis {
        log_path: PathBuf,
        #[arg(long)]
        top_n: Option<usize>,
        #[arg(long)]
        similarity_threshold: Option<f64>,
    },
    /// Export Petri net to file
    ExportPetriNet {
        source_id: String,
        output_path: PathBuf,
        #[arg(long)]
        format: Option<String>,
        #[arg(long)]
        with_metadata: Option<bool>,
    },
    /// Export event log to file
    ExportLog {
        source_id: String,
        output_path: PathBuf,
        #[arg(long)]
        format: Option<String>,
        #[arg(long)]
        with_metadata: Option<bool>,
    },
    /// Import event log from file
    ImportLog {
        input_path: PathBuf,
        #[arg(long)]
        target_format: Option<String>,
        #[arg(long)]
        merge_with: Option<String>,
    },
    /// Export process model to file
    ExportModel {
        source_id: String,
        output_path: PathBuf,
        #[arg(long)]
        format: Option<String>,
        #[arg(long)]
        with_metadata: Option<bool>,
    },
    /// Construct ontology from mapping config
    Construct {
        path: PathBuf,
        #[arg(long)]
        database: Option<String>,
        #[arg(long)]
        mapping: Option<String>,
    },
    /// Execute ontology SPARQL queries
    Execute {
        path: PathBuf,
        #[arg(long)]
        database: Option<String>,
        #[arg(long)]
        mapping: Option<String>,
    },
    /// Validate ontology mapping config
    Validate {
        path: PathBuf,
        #[arg(long)]
        database: Option<String>,
        #[arg(long)]
        mapping: Option<String>,
    },
    /// Compile ontology to N-Triples
    Compile {
        path: PathBuf,
        #[arg(long)]
        database: Option<String>,
        #[arg(long)]
        mapping: Option<String>,
    },
    /// Batch-discover process models from a directory of logs
    BatchDiscover {
        log_directory: PathBuf,
        #[arg(long)]
        pattern: Option<String>,
        #[arg(long)]
        algorithm: Option<String>,
        #[arg(long)]
        workers: Option<usize>,
    },
    /// Batch conformance check across a directory of logs
    BatchConform {
        log_directory: PathBuf,
        model_id: String,
        #[arg(long)]
        pattern: Option<String>,
        #[arg(long)]
        workers: Option<usize>,
    },
}

fn map_to_bos_command(cmd: Commands) -> BosCommand {
    match cmd {
        Commands::Discover { log_path, algorithm, max_traces, activity_filter, model_id } => {
            BosCommand::Discover(DiscoverArgs { log_path, algorithm, max_traces, activity_filter, model_id })
        }
        Commands::Conform { log_path, model_id, alignment } => {
            BosCommand::Conform(ConformArgs { log_path, model_id, alignment })
        }
        Commands::Statistics { log_path, with_variants, with_activities, with_durations } => {
            BosCommand::Statistics(StatisticsArgs { log_path, with_variants, with_activities, with_durations })
        }
        Commands::CheckConformance { log_path, model_id, alignment } => {
            BosCommand::CheckConformance(ConformArgs { log_path, model_id, alignment })
        }
        Commands::DiscoverBatch { config_path, parallel, workers } => {
            BosCommand::DiscoverBatch(BatchArgs { config_path, parallel, workers })
        }
        Commands::ListModels { algorithm_filter, date_from, date_to, sort_by } => {
            BosCommand::ListModels(ListModelsArgs { algorithm_filter, date_from, date_to, sort_by })
        }
        Commands::ValidateModel { model_id, check_soundness, check_liveness } => {
            BosCommand::ValidateModel(ValidateModelArgs { model_id, check_soundness, check_liveness })
        }
        Commands::QualityCheck { data_path, metrics, report } => {
            BosCommand::QualityCheck(QualityCheckArgs { data_path, metrics, report })
        }
        Commands::Fingerprint { log_path, baseline_model, algorithm } => {
            BosCommand::Fingerprint(FingerprintArgs { log_path, baseline_model, algorithm })
        }
        Commands::Variability { log_path, baseline_variant, variance_threshold } => {
            BosCommand::Variability(VariabilityArgs { log_path, baseline_variant, variance_threshold })
        }
        Commands::OrgEvolution { log_path, start_date, end_date, granularity } => {
            BosCommand::OrgEvolution(OrgEvolutionArgs { log_path, start_date, end_date, granularity })
        }
        Commands::VariantAnalysis { log_path, top_n, similarity_threshold } => {
            BosCommand::VariantAnalysis(VariantAnalysisArgs { log_path, top_n, similarity_threshold })
        }
        Commands::ExportPetriNet { source_id, output_path, format, with_metadata } => {
            BosCommand::ExportPetriNet(ExportArgs { source_id, output_path, format, with_metadata })
        }
        Commands::ExportLog { source_id, output_path, format, with_metadata } => {
            BosCommand::ExportLog(ExportArgs { source_id, output_path, format, with_metadata })
        }
        Commands::ImportLog { input_path, target_format, merge_with } => {
            BosCommand::ImportLog(ImportArgs { input_path, target_format, merge_with })
        }
        Commands::ExportModel { source_id, output_path, format, with_metadata } => {
            BosCommand::ExportModel(ExportArgs { source_id, output_path, format, with_metadata })
        }
        Commands::Construct { path, database, mapping } => {
            BosCommand::Construct(OntologyArgs { path, database, mapping })
        }
        Commands::Execute { path, database, mapping } => {
            BosCommand::Execute(OntologyArgs { path, database, mapping })
        }
        Commands::Validate { path, database, mapping } => {
            BosCommand::Validate(OntologyArgs { path, database, mapping })
        }
        Commands::Compile { path, database, mapping } => {
            BosCommand::Compile(OntologyArgs { path, database, mapping })
        }
        Commands::BatchDiscover { log_directory, pattern, algorithm, workers } => {
            BosCommand::BatchDiscover(BatchDiscoverArgs { log_directory, pattern, algorithm, workers })
        }
        Commands::BatchConform { log_directory, model_id, pattern, workers } => {
            BosCommand::BatchConform(BatchConformArgs { log_directory, model_id, pattern, workers })
        }
    }
}

fn main() {
    let cli = Cli::parse();

    let config = GatewayConfig {
        base_url: cli.gateway,
        timeout_ms: cli.timeout_ms,
        ..GatewayConfig::from_env()
    };

    let handler = BosCommandHandler::new(config, Duration::from_millis(cli.timeout_ms));
    let cmd = map_to_bos_command(cli.command);
    let output_format = cli.format.into();

    match handler.execute(cmd) {
        Ok(result) => {
            println!("{}", ResultFormatter::format(&result.data, output_format));
        }
        Err(e) => {
            eprintln!("error: {}", e);
            std::process::exit(1);
        }
    }
}
