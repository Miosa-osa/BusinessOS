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

// Helper to compute analysis statistics (local engine path)
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

/// Returns the X-Correlation-ID to inject on gateway requests.
/// Reads BOS_CORRELATION_ID env var or generates a fresh UUID v4.
fn correlation_id() -> String {
    std::env::var("BOS_CORRELATION_ID")
        .unwrap_or_else(|_| uuid::Uuid::new_v4().to_string())
}

/// Gateway routing: POST discover request to BusinessOS gateway.
///
/// Cloud mode: routes through BusinessOS (http) instead of local pm4py-rust engine.
/// WvdA: 30-second timeout enforces bounded execution.
/// Armstrong: non-2xx responses bail immediately with a clear error message.
fn discover_via_gateway(
    gw_url: &str,
    log_path: &str,
    algorithm: &str,
) -> anyhow::Result<ModelDiscovered> {
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(30))
        .build()
        .map_err(|e| anyhow::anyhow!("Failed to build HTTP client: {}", e))?;

    let payload = serde_json::json!({
        "log_path": log_path,
        "algorithm": algorithm
    });

    let corr_id = correlation_id();
    tracing::info!(correlation_id = %corr_id, gateway = %gw_url, verb = "discover", "gateway request");
    let resp = client
        .post(format!("{}/api/bos/discover", gw_url))
        .header("X-Correlation-ID", &corr_id)
        .json(&payload)
        .send()
        .map_err(|e| anyhow::anyhow!("Gateway request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().unwrap_or_default();
        anyhow::bail!("Gateway returned {}: {}", status, body);
    }

    let gw_resp: serde_json::Value = resp
        .json()
        .map_err(|e| anyhow::anyhow!("Failed to parse gateway response: {}", e))?;

    Ok(ModelDiscovered {
        algorithm: gw_resp["algorithm"].as_str().unwrap_or(algorithm).to_string(),
        places: gw_resp["places"].as_u64().unwrap_or(0) as usize,
        transitions: gw_resp["transitions"].as_u64().unwrap_or(0) as usize,
        arcs: gw_resp["arcs"].as_u64().unwrap_or(0) as usize,
    })
}

/// Gateway routing: POST conform request to BusinessOS gateway.
///
/// Cloud mode: routes through BusinessOS (http) instead of local pm4py-rust engine.
/// WvdA: 30-second timeout enforces bounded execution.
/// Armstrong: non-2xx responses bail immediately with a clear error message.
fn conform_via_gateway(
    gw_url: &str,
    log_path: &str,
    model_id: &str,
) -> anyhow::Result<ConformanceChecked> {
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(30))
        .build()
        .map_err(|e| anyhow::anyhow!("Failed to build HTTP client: {}", e))?;

    let payload = serde_json::json!({
        "log_path": log_path,
        "model_id": model_id
    });

    let corr_id = correlation_id();
    tracing::info!(correlation_id = %corr_id, gateway = %gw_url, verb = "conformance", "gateway request");
    let resp = client
        .post(format!("{}/api/bos/conformance", gw_url))
        .header("X-Correlation-ID", &corr_id)
        .json(&payload)
        .send()
        .map_err(|e| anyhow::anyhow!("Gateway request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().unwrap_or_default();
        anyhow::bail!("Gateway returned {}: {}", status, body);
    }

    let gw_resp: serde_json::Value = resp
        .json()
        .map_err(|e| anyhow::anyhow!("Failed to parse gateway response: {}", e))?;

    let total = gw_resp["traces_checked"].as_u64().unwrap_or(0) as usize;
    let fitness = gw_resp["fitness"].as_f64().unwrap_or(0.0);

    Ok(ConformanceChecked {
        traces_checked: total,
        fitting_traces: gw_resp["fitting_traces"].as_u64().unwrap_or_else(|| (total as f64 * fitness) as u64) as usize,
        fitness,
    })
}

/// Gateway routing: POST analyze request to BusinessOS gateway.
///
/// Cloud mode: routes through BusinessOS (http) instead of local pm4py-rust engine.
/// WvdA: 30-second timeout enforces bounded execution.
/// Armstrong: non-2xx responses bail immediately with a clear error message.
fn analyze_via_gateway(
    gw_url: &str,
    log_path: &str,
) -> anyhow::Result<AnalysisOutput> {
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(30))
        .build()
        .map_err(|e| anyhow::anyhow!("Failed to build HTTP client: {}", e))?;

    let payload = serde_json::json!({
        "log_path": log_path
    });

    let corr_id = correlation_id();
    tracing::info!(correlation_id = %corr_id, gateway = %gw_url, verb = "statistics", "gateway request");
    let resp = client
        .post(format!("{}/api/bos/statistics", gw_url))
        .header("X-Correlation-ID", &corr_id)
        .json(&payload)
        .send()
        .map_err(|e| anyhow::anyhow!("Gateway request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().unwrap_or_default();
        anyhow::bail!("Gateway returned {}: {}", status, body);
    }

    let gw_resp: serde_json::Value = resp
        .json()
        .map_err(|e| anyhow::anyhow!("Failed to parse gateway response: {}", e))?;

    let activity_frequency: Vec<ActivityStat> = gw_resp["activity_frequency"]
        .as_array()
        .unwrap_or(&vec![])
        .iter()
        .map(|a| ActivityStat {
            activity: a["activity"].as_str().unwrap_or("").to_string(),
            frequency: a["frequency"].as_u64().unwrap_or(0) as usize,
            percentage: a["percentage"].as_f64().unwrap_or(0.0),
        })
        .collect();

    let cd = &gw_resp["case_duration"];
    let case_duration = CaseDurationStat {
        min_seconds: cd["min_seconds"].as_i64().unwrap_or(0),
        max_seconds: cd["max_seconds"].as_i64().unwrap_or(0),
        avg_seconds: cd["avg_seconds"].as_f64().unwrap_or(0.0),
        median_seconds: cd["median_seconds"].as_f64().unwrap_or(0.0),
    };

    Ok(AnalysisOutput {
        log_name: gw_resp["log_name"].as_str().unwrap_or(log_path).to_string(),
        num_traces: gw_resp["num_traces"].as_u64().unwrap_or(0) as usize,
        num_events: gw_resp["num_events"].as_u64().unwrap_or(0) as usize,
        num_unique_activities: gw_resp["num_unique_activities"].as_u64().unwrap_or(0) as usize,
        num_variants: gw_resp["num_variants"].as_u64().unwrap_or(0) as usize,
        avg_trace_length: gw_resp["avg_trace_length"].as_f64().unwrap_or(0.0),
        min_trace_length: gw_resp["min_trace_length"].as_u64().unwrap_or(0) as usize,
        max_trace_length: gw_resp["max_trace_length"].as_u64().unwrap_or(0) as usize,
        activity_frequency,
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
/// Routes through BusinessOS gateway when --gateway / BOS_GATEWAY_URL is set (cloud mode).
/// Falls back to local pm4py-rust engine when not configured.
///
/// # Arguments
/// * `source` - Path to event log file
/// * `algorithm` - Discovery algorithm (alpha, inductive, heuristic, dfg) [default: alpha]
/// * `gateway` - BusinessOS gateway URL (cloud mode) [env: BOS_GATEWAY_URL]
#[verb("discover")]
fn discover(
    source: String,
    algorithm: Option<String>,
    gateway: Option<String>,
) -> Result<ModelDiscovered> {
    if let Some(ref gw_url) = gateway {
        let algo = algorithm.as_deref().unwrap_or("alpha");
        return discover_via_gateway(gw_url, &source, algo)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()));
    }

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
/// Routes through BusinessOS gateway when --gateway / BOS_GATEWAY_URL is set (cloud mode).
/// Falls back to local pm4py-rust engine when not configured.
///
/// # Arguments
/// * `log` - Path to event log file
/// * `model` - Path to model file (optional, will discover if not provided)
/// * `gateway` - BusinessOS gateway URL (cloud mode) [env: BOS_GATEWAY_URL]
#[verb("conform")]
fn conform(
    log: String,
    model: Option<String>,
    gateway: Option<String>,
) -> Result<ConformanceChecked> {
    if let Some(ref gw_url) = gateway {
        let model_id = model.as_deref().unwrap_or("auto");
        return conform_via_gateway(gw_url, &log, model_id)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()));
    }

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
/// Routes through BusinessOS gateway when --gateway / BOS_GATEWAY_URL is set (cloud mode).
/// Falls back to local pm4py-rust engine when not configured.
///
/// # Arguments
/// * `source` - Path to event log file
/// * `gateway` - BusinessOS gateway URL (cloud mode) [env: BOS_GATEWAY_URL]
#[verb("analyze")]
fn analyze(
    source: String,
    gateway: Option<String>,
) -> Result<AnalysisOutput> {
    if let Some(ref gw_url) = gateway {
        return analyze_via_gateway(gw_url, &source)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()));
    }
    compute_analysis_stats(&source)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))
}
