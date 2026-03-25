use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;
use std::path::Path;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct LogLoaded {
    pub traces: usize,
    pub events: usize,
    pub source: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct ModelDiscovered {
    pub algorithm: String,
    pub places: usize,
    pub transitions: usize,
    pub arcs: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct ConformanceChecked {
    pub traces_checked: usize,
    pub fitting_traces: usize,
    pub fitness: f64,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct ActivityStat {
    pub activity: String,
    pub frequency: usize,
    pub percentage: f64,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct CaseDurationStat {
    pub min_seconds: i64,
    pub max_seconds: i64,
    pub avg_seconds: f64,
    pub median_seconds: f64,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct AnalysisOutput {
    pub log_name: String,
    pub num_traces: usize,
    pub num_events: usize,
    pub num_unique_activities: usize,
    pub num_variants: usize,
    pub avg_trace_length: f64,
    pub min_trace_length: usize,
    pub max_trace_length: usize,
    pub activity_frequency: Vec<ActivityStat>,
    pub case_duration: CaseDurationStat,
}

// Helper to compute analysis statistics
fn compute_analysis_stats(
    source: &str,
) -> anyhow::Result<AnalysisOutput> {
    use pm4py::statistics::log_statistics;
    use pm4py::log::operations;

    let engine = bos_core::process::ProcessMiningEngine::new();
    let log = engine.load_log(source)?;

    let stats = log_statistics(&log);
    let activity_freq = operations::activity_frequency(&log);
    let total_events = stats.num_events;

    let mut activity_stats: Vec<ActivityStat> = activity_freq
        .iter()
        .map(|(activity, &frequency)| ActivityStat {
            activity: activity.clone(),
            frequency,
            percentage: if total_events > 0 {
                (frequency as f64 / total_events as f64) * 100.0
            } else {
                0.0
            },
        })
        .collect();

    activity_stats.sort_by(|a, b| b.frequency.cmp(&a.frequency));

    let mut case_durations: Vec<i64> = Vec::new();
    for trace in &log.traces {
        if trace.events.len() > 1 {
            if let (Some(first), Some(last)) = (trace.events.first(), trace.events.last()) {
                let duration = (last.timestamp - first.timestamp).num_seconds();
                case_durations.push(duration);
            }
        }
    }

    let case_duration = if !case_durations.is_empty() {
        case_durations.sort();
        let min_seconds = *case_durations.iter().min().unwrap_or(&0);
        let max_seconds = *case_durations.iter().max().unwrap_or(&0);
        let avg_seconds = case_durations.iter().sum::<i64>() as f64 / case_durations.len() as f64;
        let median_seconds = if case_durations.len() % 2 == 0 {
            let mid = case_durations.len() / 2;
            (case_durations[mid - 1] as f64 + case_durations[mid] as f64) / 2.0
        } else {
            case_durations[case_durations.len() / 2] as f64
        };

        CaseDurationStat {
            min_seconds,
            max_seconds,
            avg_seconds,
            median_seconds,
        }
    } else {
        CaseDurationStat {
            min_seconds: 0,
            max_seconds: 0,
            avg_seconds: 0.0,
            median_seconds: 0.0,
        }
    };

    let log_name = Path::new(source)
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or(source)
        .to_string();

    Ok(AnalysisOutput {
        log_name,
        num_traces: stats.num_traces,
        num_events: stats.num_events,
        num_unique_activities: stats.num_unique_activities,
        num_variants: stats.num_variants,
        avg_trace_length: stats.avg_trace_length,
        min_trace_length: stats.min_trace_length,
        max_trace_length: stats.max_trace_length,
        activity_frequency: activity_stats,
        case_duration,
    })
}

#[noun("pm4py", "Process mining with pm4py-rust — discover, analyze, and check conformance of business processes")]

/// Load an event log from file
///
/// # Arguments
/// * `source` - Path to event log file (XES, CSV, JSON)
#[verb("load")]
fn load(source: String) -> Result<LogLoaded> {
    use bos_core::process::ProcessMiningEngine;

    let engine = ProcessMiningEngine::new();
    let log = engine.load_log(&source)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let event_count = log.traces.iter()
        .map(|t| t.events.len())
        .sum::<usize>();

    Ok(LogLoaded {
        traces: log.traces.len(),
        events: event_count,
        source: Path::new(&source)
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or(&source)
            .to_string(),
    })
}

/// Discover a process model from an event log
///
/// # Arguments
/// * `source` - Path to event log file
/// * `algorithm` - Discovery algorithm (alpha, inductive, heuristic, dfg) [default: alpha]
#[verb("discover")]
fn discover(source: String, algorithm: Option<String>) -> Result<ModelDiscovered> {
    use bos_core::process::ProcessMiningEngine;

    let engine = ProcessMiningEngine::new();
    let log = engine.load_log(&source)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    let algo = algorithm.unwrap_or_else(|| "alpha".to_string());

    let result = match algo.as_str() {
        "alpha" => engine.discover_alpha(&log),
        "inductive" | "tree" => engine.discover_tree(&log),
        "heuristic" => engine.discover_heuristic(&log),
        "dfg" => {
            // DFG discovery uses AlphaMiner as fallback since we have it readily available
            engine.discover_alpha(&log)
        }
        _ => return Err(clap_noun_verb::NounVerbError::execution_error(
            format!("Unknown algorithm: {}. Use: alpha, inductive, heuristic, dfg", algo)
        )),
    };

    let discovered = result
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    Ok(ModelDiscovered {
        algorithm: discovered.algorithm,
        places: discovered.places,
        transitions: discovered.transitions,
        arcs: discovered.arcs,
    })
}

/// Check conformance of log against model
///
/// # Arguments
/// * `log` - Path to event log file
/// * `model` - Path to model file (optional, will discover if not provided)
#[verb("conform")]
fn conform(log: String, model: Option<String>) -> Result<ConformanceChecked> {
    use bos_core::process::ProcessMiningEngine;
    use pm4py::conformance::TokenReplay;

    let engine = ProcessMiningEngine::new();
    let event_log = engine.load_log(&log)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    // Use discovered Petri net for conformance checking if no model provided
    // (model file support can be extended in future versions)
    let _model = model; // Will be used for external model loading in future
    let miner = pm4py::discovery::AlphaMiner::new();
    let petri_net = miner.discover(&event_log);

    // Perform token replay conformance check
    let token_replay = TokenReplay::new();
    let conformance_result = token_replay.check(&event_log, &petri_net);

    let total_traces = event_log.traces.len();
    let fitting_traces = (total_traces as f64 * conformance_result.fitness) as usize;

    Ok(ConformanceChecked {
        traces_checked: total_traces,
        fitting_traces,
        fitness: conformance_result.fitness,
    })
}

/// Analyze event log statistics
///
/// # Arguments
/// * `source` - Path to event log file
#[verb("analyze")]
fn analyze(source: String) -> Result<AnalysisOutput> {
    compute_analysis_stats(&source)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))
}
