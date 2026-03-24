//! bos-core process mining — pm4py-rust integration for BusinessOS.
//!
//! Provides process discovery, conformance checking, and analysis
//! capabilities using the pm4py-rust library.

use anyhow::Result;
use pm4py::{EventLog, Trace, Event};
use chrono::{DateTime, Utc};
use std::path::Path;

// Re-export key pm4py types
pub use pm4py::{
    discovery::{AlphaMiner, InductiveMiner, HeuristicMiner, TreeMiner},
    models::{ProcessTree, PetriNet, CausalNet, Footprints, ProcessTreeNode},
    log::{Event as PMEvent, Trace as PMTrace, EventLog as PMEventLog},
};

/// Process mining result from discovery
#[derive(Debug, Clone)]
pub struct ProcessDiscoveryResult {
    pub algorithm: String,
    pub places: usize,
    pub transitions: usize,
    pub arcs: usize,
    pub fitness: Option<f64>,
}

/// Conformance checking result
#[derive(Debug, Clone)]
pub struct ConformanceResult {
    pub traces_checked: usize,
    pub fitting_traces: usize,
    pub fitness: f64,
    pub details: Vec<String>,
}

/// Process Mining Engine
pub struct ProcessMiningEngine {
    config: pm4py::Config,
}

impl ProcessMiningEngine {
    /// Create a new process mining engine
    pub fn new() -> Self {
        Self {
            config: pm4py::Config::default(),
        }
    }

    /// Load an event log from a file
    pub fn load_log<P: AsRef<Path>>(&self, path: P) -> Result<EventLog> {
        let path_str = path.as_ref().to_str().unwrap();
        let extension = path.as_ref().extension()
            .and_then(|e| e.to_str())
            .unwrap_or("");

        let log = match extension {
            "xes" => {
                let reader = pm4py::io::XESReader::new();
                reader.read(path.as_ref())?
            },
            "csv" => {
                let reader = pm4py::io::CSVReader::new();
                reader.read(path.as_ref())?
            },
            "json" => {
                let reader = pm4py::io::JsonEventLogReader::new();
                reader.read(path.as_ref())?
            },
            _ => return Err(anyhow::anyhow!("Unsupported log format: {}", extension)),
        };

        Ok(log)
    }

    /// Discover a process model using Alpha Miner
    pub fn discover_alpha(&self, log: &EventLog) -> Result<ProcessDiscoveryResult> {
        let miner = AlphaMiner::new();
        let net = miner.discover(log);

        Ok(ProcessDiscoveryResult {
            algorithm: "Alpha Miner".to_string(),
            places: net.places.len(),
            transitions: net.transitions.len(),
            arcs: net.arcs.len(),
            fitness: None,
        })
    }

    /// Discover a process tree using Tree Miner
    pub fn discover_tree(&self, log: &EventLog) -> Result<ProcessDiscoveryResult> {
        let miner = TreeMiner::new();
        let tree = miner.discover(log);

        let (nodes, operators) = Self::count_tree_nodes(&tree.root);

        Ok(ProcessDiscoveryResult {
            algorithm: "Tree Miner".to_string(),
            places: nodes,
            transitions: operators,
            arcs: nodes - 1,
            fitness: None,
        })
    }

    /// Discover a process model using Heuristic Miner
    pub fn discover_heuristic(&self, log: &EventLog) -> Result<ProcessDiscoveryResult> {
        let miner = HeuristicMiner::new();
        let net = miner.discover(log);

        Ok(ProcessDiscoveryResult {
            algorithm: "Heuristic Miner".to_string(),
            places: net.places.len(),
            transitions: net.transitions.len(),
            arcs: net.arcs.len(),
            fitness: None,
        })
    }

    /// Create an event log from workspace data
    pub fn create_log_from_events(&self, events: Vec<ProcessEvent>) -> EventLog {
        let mut log = EventLog::new();

        // Group events by case_id
        let mut cases: std::collections::HashMap<String, Vec<ProcessEvent>> = std::collections::HashMap::new();
        for event in events {
            cases.entry(event.case_id.clone()).or_default().push(event);
        }

        // Sort events by timestamp within each case
        for (_, mut case_events) in cases {
            case_events.sort_by(|a, b| a.timestamp.cmp(&b.timestamp));

            let mut trace = Trace::new(case_events[0].case_id.clone());
            for event in case_events {
                let pm_event = Event::new(event.activity, event.timestamp);
                // Note: pm4py Event doesn't have set_attribute, attributes are set during construction
                trace.add_event(pm_event);
            }
            log.add_trace(trace);
        }

        log
    }

    fn count_tree_nodes(node: &ProcessTreeNode) -> (usize, usize) {
        match node {
            ProcessTreeNode::Activity(_) => (1, 0),
            ProcessTreeNode::Operator { children, .. } => {
                let mut nodes = 1;
                let mut operators = 1;
                for child in children {
                    let (n, o) = Self::count_tree_nodes(child);
                    nodes += n;
                    operators += o;
                }
                (nodes, operators)
            }
        }
    }
}

impl Default for ProcessMiningEngine {
    fn default() -> Self {
        Self::new()
    }
}

/// Process event for creating event logs
#[derive(Debug, Clone)]
pub struct ProcessEvent {
    pub case_id: String,
    pub activity: String,
    pub timestamp: DateTime<Utc>,
    pub attributes: Option<std::collections::HashMap<String, String>>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Duration;

    #[test]
    fn test_create_log_from_events() {
        let engine = ProcessMiningEngine::new();

        let events = vec![
            ProcessEvent {
                case_id: "case1".to_string(),
                activity: "A".to_string(),
                timestamp: Utc::now(),
                attributes: None,
            },
            ProcessEvent {
                case_id: "case1".to_string(),
                activity: "B".to_string(),
                timestamp: Utc::now() + Duration::hours(1),
                attributes: None,
            },
            ProcessEvent {
                case_id: "case2".to_string(),
                activity: "A".to_string(),
                timestamp: Utc::now(),
                attributes: None,
            },
        ];

        let log = engine.create_log_from_events(events);
        assert_eq!(log.traces.len(), 2);
        assert_eq!(log.traces[0].events.len(), 2);
        assert_eq!(log.traces[1].events.len(), 1);
    }

    #[test]
    fn test_discover_alpha() {
        let engine = ProcessMiningEngine::new();

        let mut log = EventLog::new();
        let mut trace1 = Trace::new("case1".to_string());
        trace1.add_event(Event::new("A", Utc::now()));
        trace1.add_event(Event::new("B", Utc::now()));
        trace1.add_event(Event::new("C", Utc::now()));
        log.add_trace(trace1);

        let result = engine.discover_alpha(&log).unwrap();
        assert_eq!(result.algorithm, "Alpha Miner");
    }
}
