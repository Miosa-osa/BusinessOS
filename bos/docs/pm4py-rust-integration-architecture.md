# bos CLI + pm4py-rust Integration Architecture

**Document Date:** 2026-03-24
**Status:** Design (Not Implemented)
**Architect:** Claude Code Agent

---

## Executive Summary

This document designs the integration of pm4py-rust (process mining library) into the bos CLI, enabling enterprise-grade process mining capabilities for BusinessOS workspaces. The architecture leverages the existing noun-verb command structure and integrates pm4py-rust as a workspace-managed dependency.

**Key Outcomes:**
- Full process discovery, conformance checking, and analytics via CLI
- 5 MVP commands (Phase 1) for immediate value
- Extensible architecture for 40+ advanced commands
- Integration with ODCS workspace data layer
- Signal Theory encoding for all outputs

---

## Current bos CLI Architecture

### Structure Overview

```
BusinessOS/bos/
├── Cargo.toml (workspace manifest)
├── cli/              (bos CLI binary)
│   ├── Cargo.toml
│   ├── src/
│   │   ├── main.rs   (entry point, clap-noun-verb dispatcher)
│   │   └── nouns/    (noun subcommands)
│   │       ├── mod.rs
│   │       ├── workspace.rs
│   │       ├── schema.rs
│   │       ├── data.rs
│   │       ├── decisions.rs
│   │       ├── knowledge.rs
│   │       ├── ontology.rs
│   │       ├── search.rs
│   │       ├── validate.rs
│   │       └── pm4py.rs (EXISTING — stub)
├── core/             (shared business logic)
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs
│       ├── workspace.rs
│       ├── schema.rs
│       ├── decisions.rs
│       ├── knowledge.rs
│       ├── ontology/
│       ├── rdf/
│       └── process/   (PROCESS MINING ENGINE)
│           └── mod.rs (ProcessMiningEngine using pm4py)
├── ingest/           (data import)
├── config/           (configuration)
└── tests/            (integration tests)
```

### Key Architectural Principles

1. **Noun-Verb Structure**: Commands follow `bos <noun> <verb> [args]`
   - Example: `bos workspace init`, `bos schema convert`, `bos pm4py discover`

2. **Layered Separation**:
   - **CLI Layer** (`cli/src/nouns/*.rs`): Argument parsing, formatting, output serialization
   - **Core Layer** (`core/src/*.rs`): Business logic, pm4py-rust integration
   - **Dependency Layer** (`Cargo.toml`): External crates (pm4py-rust is path dependency)

3. **Async-First**: Uses `clap-noun-verb` with `async` feature for concurrent operations

4. **Result Serialization**: All outputs serialized as JSON via `serde`

### Existing PM4Py Stubs

**File:** `cli/src/nouns/pm4py.rs`

Already contains 4 verbs (stub implementations):
- `load` — Load event log, report traces/events
- `discover` — Process discovery with 3 algorithms (alpha, inductive, heuristic)
- `conform` — Conformance checking (currently simulated at 85% fitness)
- `analyze` — Event log statistics

**File:** `core/src/process/mod.rs`

Implements `ProcessMiningEngine` struct:
- `load_log(path)` — Read XES, CSV, JSON
- `discover_alpha(log)` — Alpha Miner discovery
- `discover_tree(log)` — Tree (Inductive) Miner discovery
- `discover_heuristic(log)` — Heuristic Miner discovery
- `create_log_from_events(events)` — Programmatic log creation

**Current Dependency Status:**
- pm4py-rust is already in `core/Cargo.toml` as path dependency
- Already depends on pm4py 0.3.0 (via path at `/Users/sac/chatmangpt/pm4py-rust`)

---

## Integration Points

### 1. Dependency Declaration

**Location:** `/Users/sac/chatmangpt/BusinessOS/bos/core/Cargo.toml`

**Current Status:** ✓ Already present
```toml
[dependencies]
pm4py = { path = "/Users/sac/chatmangpt/pm4py-rust" }
```

**Action Required:** None. Dependency already configured.

### 2. Core Module: ProcessMiningEngine

**Location:** `core/src/process/mod.rs`

**Current Capabilities:**
- Event log loading (XES, CSV, JSON)
- Discovery: Alpha, Inductive, Heuristic miners
- Trace creation from structured events
- Tree node counting for statistics

**Integration Points:**
- Uses `pm4py::EventLog`, `pm4py::Trace`, `pm4py::Event`
- Uses discovery traits: `AlphaMiner`, `TreeMiner`, `HeuristicMiner`
- Wraps results in `ProcessDiscoveryResult` struct

**Future Expansion:** Add conformance, performance, statistics, and predictive analytics

### 3. CLI Layer: Noun Definition

**Location:** `cli/src/nouns/pm4py.rs`

**Pattern:** Uses `#[noun(...)]` and `#[verb(...)]` macros from `clap-noun-verb`
```rust
#[noun("pm4py", "Process mining with pm4py-rust...")]

#[verb("discover")]
fn discover(source: String, algorithm: Option<String>) -> Result<ModelDiscovered> {
    // Call core engine, format output
}
```

**Integration Pattern:**
1. Accept CLI args (file paths, algorithm names, options)
2. Instantiate `ProcessMiningEngine`
3. Call core methods
4. Wrap results in serializable structs
5. Return via `Result<T: Serialize>`

---

## Proposed CLI Commands

### Command Taxonomy

All commands follow `bos pm4py <verb> [args]`.

#### Discovery Commands
```bash
# Alpha Miner (simplest, fast)
bos pm4py discover <log.xes>
bos pm4py discover <log.xes> --algorithm alpha

# Inductive Miner (handles noise better)
bos pm4py discover <log.xes> --algorithm inductive

# Heuristic Miner (handles loops, flexible thresholds)
bos pm4py discover <log.xes> --algorithm heuristic

# Directly-Follows Graph (mining without model)
bos pm4py dfg <log.xes>

# Extended discovery (DFG, performance, variants)
bos pm4py variants <log.xes>
bos pm4py performance <log.xes>
```

#### Conformance Commands
```bash
# Check fitness of log against discovered model
bos pm4py conform <log.xes>

# Check fitness against existing model
bos pm4py conform <log.xes> --model model.pnml

# Alignment-based conformance
bos pm4py align <log.xes> --model model.pnml

# Footprints-based checking
bos pm4py footprints <log.xes> --model model.pnml
```

#### Analytics Commands
```bash
# Log statistics
bos pm4py stats <log.xes>

# Trace analysis
bos pm4py traces <log.xes>

# Activity analysis
bos pm4py activities <log.xes>

# Performance metrics
bos pm4py perf <log.xes>

# Predictive (remaining time, next activity)
bos pm4py predict <log.xes> --mode remaining-time
bos pm4py predict <log.xes> --mode next-activity
```

#### Model Export Commands
```bash
# Export discovered model as PNML
bos pm4py export <log.xes> --format pnml --output model.pnml

# Export as BPMN
bos pm4py export <log.xes> --format bpmn --output model.bpmn

# Export as DFG JSON
bos pm4py export <log.xes> --format dfg-json --output dfg.json
```

#### Log Manipulation Commands
```bash
# Filter by activity
bos pm4py filter <log.xes> --activity A --activity B

# Filter by variant
bos pm4py variants <log.xes> --top-k 10

# Statistics filtering
bos pm4py filter-stats <log.xes>
```

---

## Data Flow Architecture

### Overall Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ Command Line Input                                              │
│ bos pm4py discover ./log.xes --algorithm alpha                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ CLI Layer (nouns/pm4py.rs)                                      │
│ - Parse arguments (source, algorithm)                           │
│ - Validate file existence                                       │
│ - Create ProcessMiningEngine                                    │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ Core Layer (core/src/process/mod.rs)                            │
│ - ProcessMiningEngine::load_log()                               │
│   └─> pm4py::io::{XESReader, CSVReader, JsonReader}           │
│       └─> Returns pm4py::EventLog                              │
│                                                                 │
│ - ProcessMiningEngine::discover_alpha()                        │
│   └─> pm4py::discovery::AlphaMiner::discover()                │
│       └─> Returns pm4py::models::PetriNet                      │
│           (places, transitions, arcs)                           │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ Result Struct (CLI Layer)                                       │
│ ModelDiscovered {                                               │
│   algorithm: "Alpha Miner",                                     │
│   places: 5,                                                    │
│   transitions: 4,                                               │
│   arcs: 12,                                                     │
│ }                                                               │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ JSON Serialization & Output                                     │
│ serde::Serialize via clap-noun-verb                             │
│ {                                                               │
│   "algorithm": "Alpha Miner",                                   │
│   "places": 5,                                                  │
│   "transitions": 4,                                             │
│   "arcs": 12                                                    │
│ }                                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Per-Feature Data Flows

#### Discovery Flow

```
File (XES/CSV/JSON)
    ↓
pm4py::io::{XESReader, CSVReader, JsonReader}
    ↓
EventLog (traces, events, attributes)
    ↓
AlphaMiner::discover(log) → PetriNet
    or
TreeMiner::discover(log) → ProcessTree
    or
HeuristicMiner::discover(log) → PetriNet
    ↓
ProcessDiscoveryResult struct
    ↓
JSON output
```

#### Conformance Flow

```
EventLog (from file or created)
    ↓
AutomaticDiscovery or ProvidedModel
    ↓
TokenReplay::check(log, net)
    or
FootprintsChecker::check(log, net)
    or
AlignmentChecker::check(log, net)
    ↓
ConformanceChecked struct {
  traces_checked,
  fitting_traces,
  fitness: f64
}
    ↓
JSON output
```

#### Statistics Flow

```
EventLog
    ↓
Multiple Analysis Functions:
  - Trace count
  - Event count
  - Unique activities
  - Activity frequency
  - Performance metrics
  - Variant analysis
    ↓
serde_json::json!{...}
    ↓
JSON output
```

---

## Module Organization

### CLI Module Structure (`cli/src/nouns/pm4py.rs`)

**Suggested organization by complexity:**

```rust
// ============= Phase 1: MVP (Implement First) =============
#[verb("load")]
fn load(source: String) -> Result<LogLoaded>

#[verb("discover")]
fn discover(source: String, algorithm: Option<String>) -> Result<ModelDiscovered>

#[verb("conform")]
fn conform(log: String, model: Option<String>) -> Result<ConformanceChecked>

#[verb("stats")]
fn stats(source: String) -> Result<serde_json::Value>

#[verb("dfg")]
fn dfg(source: String) -> Result<serde_json::Value>

// ============= Phase 2: Extended Discovery =============
#[verb("variants")]
fn variants(source: String, top_k: Option<usize>) -> Result<serde_json::Value>

#[verb("performance")]
fn performance(source: String) -> Result<serde_json::Value>

// ============= Phase 3: Advanced Conformance =============
#[verb("align")]
fn align(log: String, model: String) -> Result<serde_json::Value>

#[verb("footprints")]
fn footprints(log: String, model: String) -> Result<serde_json::Value>

// ============= Phase 4: Predictive & Advanced =============
#[verb("predict")]
fn predict(source: String, mode: String) -> Result<serde_json::Value>

#[verb("filter")]
fn filter(source: String, activities: Vec<String>) -> Result<serde_json::Value>
```

### Core Module Structure (`core/src/process/mod.rs`)

**Suggested organization:**

```rust
// ========== Event Log Types & Loading ==========
pub struct ProcessEvent { ... }
pub struct ProcessDiscoveryResult { ... }
pub struct ConformanceResult { ... }

pub struct ProcessMiningEngine { ... }

impl ProcessMiningEngine {
    pub fn new() -> Self
    pub fn load_log<P: AsRef<Path>>(&self, path: P) -> Result<EventLog>
    pub fn create_log_from_events(&self, events: Vec<ProcessEvent>) -> EventLog
}

// ========== Discovery Methods ==========
impl ProcessMiningEngine {
    pub fn discover_alpha(&self, log: &EventLog) -> Result<ProcessDiscoveryResult>
    pub fn discover_tree(&self, log: &EventLog) -> Result<ProcessDiscoveryResult>
    pub fn discover_heuristic(&self, log: &EventLog) -> Result<ProcessDiscoveryResult>
    pub fn discover_dfg(&self, log: &EventLog) -> Result<DFGResult>
    pub fn discover_causal_net(&self, log: &EventLog) -> Result<CausalNetResult>
    // Phase 2+
    pub fn discover_split_miner(&self, log: &EventLog, param: f64) -> Result<ProcessDiscoveryResult>
    pub fn discover_ilp(&self, log: &EventLog) -> Result<ProcessDiscoveryResult>
}

// ========== Conformance Methods ==========
impl ProcessMiningEngine {
    pub fn check_conformance_token_replay(
        &self,
        log: &EventLog,
        net: &PetriNet,
    ) -> Result<ConformanceResult>

    pub fn check_conformance_footprints(
        &self,
        log: &EventLog,
        net: &PetriNet,
    ) -> Result<ConformanceResult>

    pub fn check_conformance_alignment(
        &self,
        log: &EventLog,
        net: &PetriNet,
    ) -> Result<ConformanceResult>
}

// ========== Analytics Methods ==========
impl ProcessMiningEngine {
    pub fn analyze_log_statistics(&self, log: &EventLog) -> Result<serde_json::Value>
    pub fn analyze_performance(&self, log: &EventLog) -> Result<serde_json::Value>
    pub fn analyze_variants(&self, log: &EventLog) -> Result<serde_json::Value>
    pub fn predict_next_activity(&self, log: &EventLog) -> Result<serde_json::Value>
    pub fn predict_remaining_time(&self, log: &EventLog) -> Result<serde_json::Value>
}

// ========== Model Export Methods ==========
impl ProcessMiningEngine {
    pub fn export_pnml(&self, net: &PetriNet, path: &str) -> Result<()>
    pub fn export_bpmn(&self, net: &PetriNet, path: &str) -> Result<()>
    pub fn export_dfg_json(&self, dfg: &DFG, path: &str) -> Result<()>
}

// ========== Filtering Methods ==========
impl ProcessMiningEngine {
    pub fn filter_by_activity(&self, log: &EventLog, activities: Vec<&str>) -> EventLog
    pub fn filter_by_variant(&self, log: &EventLog, top_k: usize) -> EventLog
    pub fn filter_by_frequency(&self, log: &EventLog, min_freq: usize) -> EventLog
}
```

---

## PM4Py-Rust API Mapping

### Discovery Algorithms

| PM4Py-Rust API | Purpose | Complexity | CLI Command | Phase |
|----------------|---------|-----------|-------------|-------|
| `AlphaMiner::discover()` | Classical α-algorithm | Low | `discover --algorithm alpha` | 1 |
| `TreeMiner::discover()` | Process tree mining | Low-Med | `discover --algorithm inductive` | 1 |
| `HeuristicMiner::discover()` | Flexible heuristic | Med | `discover --algorithm heuristic` | 1 |
| `DFGMiner::discover()` | Directly-Follows Graph | Low | `dfg` | 1 |
| `SplitMiner::discover()` | Split-based discovery | Med | `discover --algorithm split` | 2 |
| `ILPMiner::discover()` | Integer Linear Programming | High | `discover --algorithm ilp` | 2 |
| `CausalNetMiner::discover()` | Causal net discovery | Med-High | (Phase 3) | 3 |
| `OCPMDiscoveryMiner` | Object-centric process mining | High | (Phase 4) | 4 |

### Conformance Checking

| PM4Py-Rust API | Purpose | Speed | Accuracy | Phase |
|---|---|---|---|---|
| `TokenReplay::check()` | Token-based replay | ★★★★★ | ★★★☆☆ | 1 |
| `FootprintsConformanceChecker` | Footprints-based | ★★★★★ | ★★★★☆ | 2 |
| `AlignmentChecker` | Optimal alignment | ★★☆☆☆ | ★★★★★ | 2 |
| `BehavioralProfile` | Activity relationships | ★★★★☆ | ★★★☆☆ | 2 |
| `Precision`, `Generalization`, `Simplicity` | Quality metrics | Varies | ★★★★★ | 3 |
| `FourSpectrum` | Unified quality metric | ★★★☆☆ | ★★★★★ | 3 |
| `AlignmentVariants::AStarAligner` | A* alignment | ★★☆☆☆ | ★★★★★ | 3 |

### Statistics & Analytics

| PM4Py-Rust API | Purpose | Phase |
|---|---|---|
| `statistics::log_stats` | Basic event log metrics | 1 |
| `statistics::trace_stats` | Trace-level metrics | 2 |
| `statistics::performance` | Sojourn time, cycle time | 2 |
| `statistics::tree_stats` | Process tree analysis | 2 |
| `statistics::ml_features` | ML feature extraction | 3 |
| `predictive::NextActivityPredictor` | Next activity prediction | 2 |
| `predictive::RemainingTimePredictor` | Remaining time prediction | 2 |
| `predictive::OutcomePredictor` | Case outcome prediction | 3 |

---

## Command Naming Convention

### Verb Naming Rules

**Tense:** Always imperative (discover, not discovering)
**Specificity:** More specific is better (align vs check-conformance)
**Brevity:** Short names for frequent commands

### Pattern Examples

```bash
# Object-centric naming
bos pm4py <object> <verb>
bos pm4py log load <file>           # Load a log
bos pm4py model discover <file>     # Discover a model
bos pm4py trace analyze <file>      # Analyze traces

# Action-centric naming (simpler, preferred)
bos pm4py <verb> [args]
bos pm4py load <file>               # Load event log
bos pm4py discover <file>           # Discover model
bos pm4py conform <log> --model <m> # Check conformance
bos pm4py align <log> --model <m>   # Alignment-based
bos pm4py stats <file>              # Statistics
bos pm4py perf <file>               # Performance metrics
bos pm4py predict <file> --mode X   # Prediction
bos pm4py export <file> --format X  # Export model
bos pm4py filter <file> --activity A,B  # Filter activities
bos pm4py variants <file> --top-k 10    # Top variants
```

**Decision:** Use action-centric naming (simpler, more discoverable).

---

## Implementation Roadmap

### Phase 1: MVP (Week 1 — 16-20 hours)

**Goal:** Full-featured process discovery and basic conformance

**Commands:**
1. `bos pm4py load` — Load and report on event logs
2. `bos pm4py discover` — Discover with 3 algorithms (alpha, inductive, heuristic)
3. `bos pm4py conform` — Token replay conformance checking
4. `bos pm4py stats` — Event log statistics
5. `bos pm4py dfg` — Directly-Follows Graph mining

**Implementation Checklist:**
- [ ] Enhance `load()` verb: better file validation, attribute extraction
- [ ] Implement `discover()` with all 3 algorithms from stubs
- [ ] Implement token-replay `conform()` (replace 85% simulation)
- [ ] Enhance `stats()` to use pm4py statistics module
- [ ] Add `dfg()` verb for directly-follows graph
- [ ] Add Signal Theory encoding to all outputs
- [ ] Write 10-15 integration tests
- [ ] Document CLI examples in README

**Output Samples:**

```bash
$ bos pm4py load ./logs/invoice.xes
{
  "traces": 245,
  "events": 1823,
  "source": "invoice.xes",
  "unique_activities": 12,
  "trace_variants": 8
}

$ bos pm4py discover ./logs/invoice.xes --algorithm alpha
{
  "algorithm": "Alpha Miner",
  "places": 14,
  "transitions": 12,
  "arcs": 43,
  "fitness": 0.96
}

$ bos pm4py stats ./logs/invoice.xes
{
  "traces": 245,
  "total_events": 1823,
  "unique_activities": 12,
  "most_common_activity": { "invoice_create": 243 },
  "avg_events_per_trace": 7.44,
  "event_attributes": ["timestamp", "resource", "cost"],
  "trace_variants": 8
}
```

### Phase 2: Extended Discovery (Week 2 — 14-18 hours)

**Goal:** Advanced mining algorithms, performance analysis, variant analysis

**New Commands:**
1. `bos pm4py variants` — Variant filtering and analysis
2. `bos pm4py performance` — Performance metrics (sojourn time, cycle time)
3. `bos pm4py discover --algorithm split` — Split Miner
4. `bos pm4py footprints` — Footprints-based conformance
5. `bos pm4py perf-dfg` — Performance-enhanced DFG

**Implementation Checklist:**
- [ ] Integrate `pm4py::statistics::TreeStatistics` for analysis
- [ ] Implement variant extraction and top-k filtering
- [ ] Add performance metrics from `pm4py::performance`
- [ ] Implement Split Miner integration
- [ ] Implement Footprints-based conformance
- [ ] Add filtering by activity, variant, time range
- [ ] Write 15-20 tests
- [ ] Add performance benchmarks

### Phase 3: Advanced Conformance (Week 3 — 18-24 hours)

**Goal:** Alignment-based conformance, quality metrics, detailed diagnostics

**New Commands:**
1. `bos pm4py align` — Alignment-based conformance
2. `bos pm4py quality` — 4-Spectrum quality metrics
3. `bos pm4py precision` — Precision metric
4. `bos pm4py simplicity` — Simplicity metric
5. `bos pm4py export` — Export models (PNML, BPMN, JSON)

**Implementation Checklist:**
- [ ] Implement alignment-based conformance checking
- [ ] Add 4-Spectrum quality metric
- [ ] Add precision and generalization metrics
- [ ] Implement model export (PNML, BPMN, DFG JSON)
- [ ] Add diagnostic output with trace-level details
- [ ] Implement ILP Miner discovery
- [ ] Write 20-25 tests

### Phase 4: Predictive & Advanced (Week 4+ — 20+ hours)

**Goal:** Predictive analytics, object-centric mining, advanced filtering

**New Commands:**
1. `bos pm4py predict --mode next-activity` — Next activity prediction
2. `bos pm4py predict --mode remaining-time` — Remaining time prediction
3. `bos pm4py discover --algorithm ocpm` — Object-centric mining
4. `bos pm4py declare` — Declare constraint mining
5. `bos pm4py filter` — Advanced filtering

**Later Enhancements:**
- Streaming log analysis
- Temporal process mining
- Resource and organizational mining
- Anomaly detection
- Process comparison

---

## Signal Theory Integration

### Output Encoding S=(M,G,T,F,W)

All pm4py-rust CLI outputs encode using Signal Theory:

| Mode (M) | Genre (G) | Type (T) | Format (F) | Structure (W) |
|----------|-----------|----------|-----------|---------------|
| `data` | `analysis` | `inform` | `json` | `list` or `object` |

**Example: `bos pm4py discover` output**

```json
{
  "_signal": {
    "mode": "data",
    "genre": "analysis",
    "type": "inform",
    "format": "json",
    "structure": "result-object"
  },
  "algorithm": "Alpha Miner",
  "places": 14,
  "transitions": 12,
  "arcs": 43,
  "fitness": 0.96,
  "quality_metrics": {
    "precision": 0.89,
    "generalization": 0.92,
    "simplicity": 0.85,
    "four_spectrum": 0.88
  }
}
```

### S/N (Signal-to-Noise) Scoring

All outputs scored for signal quality (target: S/N ≥ 0.7):

**High Signal Outputs:**
- Structured metrics (places, transitions, fitness)
- Statistical summaries
- Performance measurements

**Low Noise Features:**
- Exclude debug info from standard output
- No gratuitous intermediate steps
- Focus on actionable metrics

---

## Error Handling & Validation

### Input Validation

```rust
// File existence
if !Path::new(&source).exists() {
    return Err(NounVerbError::execution_error(
        format!("File not found: {}", source)
    ));
}

// Format detection
match extension {
    "xes" | "csv" | "json" => Ok(()),
    ext => Err(NounVerbError::execution_error(
        format!("Unsupported format: {}. Use: xes, csv, json", ext)
    )),
}

// Algorithm validation
match algo.as_str() {
    "alpha" | "inductive" | "heuristic" => Ok(()),
    unknown => Err(NounVerbError::execution_error(
        format!("Unknown algorithm: {}. Use: alpha, inductive, heuristic", unknown)
    )),
}
```

### Error Categories

| Category | Example | Recovery |
|----------|---------|----------|
| **File I/O** | File not found, unreadable | Exit with error code 2 |
| **Format Error** | Invalid XES structure | Show parse error with line number |
| **Algorithm Error** | Unsupported algorithm | Show available algorithms |
| **Resource Error** | Out of memory on large log | Suggest filtering |
| **Validation Error** | Log has no events | Show log structure |

---

## Testing Strategy

### Unit Tests (Core Module)

**Location:** `core/src/process/mod.rs` (existing tests + new)

```rust
#[test]
fn test_load_xes_file()
fn test_load_csv_file()
fn test_load_json_file()
fn test_discover_alpha()
fn test_discover_tree()
fn test_discover_heuristic()
fn test_conformance_token_replay()
fn test_conformance_footprints()
fn test_statistics_extraction()
fn test_dfg_discovery()
fn test_filter_by_activity()
fn test_create_log_from_events()
```

### Integration Tests (CLI Layer)

**Location:** `tests/` directory

```bash
# Phase 1 integration tests
test_load_command.rs
test_discover_command.rs
test_conform_command.rs
test_stats_command.rs
test_dfg_command.rs

# Phase 2+ integration tests
test_variants_command.rs
test_performance_command.rs
test_align_command.rs
test_export_command.rs
test_filter_command.rs
```

**Test Pattern:**
```rust
#[test]
fn test_discover_alpha_integration() {
    let output = Command::new("bos")
        .args(&["pm4py", "discover", "test-logs/simple.xes", "--algorithm", "alpha"])
        .output()
        .expect("Failed to run bos");

    assert!(output.status.success());
    let json: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert_eq!(json["algorithm"], "Alpha Miner");
}
```

### Test Data Sets

**Location:** `tests/test-logs/`

- `simple.xes` — 10 traces, 3 activities (α-algorithm test)
- `loop.xes` — 20 traces with loops (tree miner test)
- `noise.xes` — 50 traces with 10% noise (heuristic test)
- `large.csv` — 5000 events, 200 traces
- `complex.json` — Multi-variant with resources

---

## Workspace Integration

### ODCS Workspace Connection

Future enhancement: Use bos CLI to query workspace event logs directly.

```rust
// Future: Read from workspace RDF store
pub fn discover_from_workspace(
    &self,
    workspace_path: &str,
    process_name: &str,
) -> Result<ProcessDiscoveryResult>
```

### Data Modelling SDK Bridge

Future: Export discovered models as ODC (Open Data Common) schemas.

```bash
bos pm4py discover ./log.xes --export-odc schema.odc
```

---

## Performance Considerations

### Scalability Targets

| Log Size | Traces | Events | Alpha Time | Inductive Time | Heuristic Time |
|----------|--------|--------|-----------|----------------|----------------|
| Small | 100 | 1K | <1s | <2s | <1s |
| Medium | 1K | 10K | 2-5s | 5-10s | 3-5s |
| Large | 10K | 100K | 20-40s | 1-2m | 30-60s |
| XL | 100K | 1M | 5-10m | 10-20m | 3-5m |

**Optimization Strategy:**
- Phase 1: No optimization (baseline)
- Phase 2: Lazy event attribute parsing
- Phase 3: Parallel trace processing
- Phase 4: Memory-mapped file I/O for >1M events

### Memory Management

```rust
// Streaming approach for large logs (Phase 3+)
pub fn discover_streaming(
    &self,
    file_path: &str,
    chunk_size: usize,
) -> Result<ProcessDiscoveryResult>
```

---

## Dependencies Summary

### Current Dependencies

**Already in Workspace `Cargo.toml`:**
- `clap` 4.5 — CLI argument parsing
- `clap-noun-verb` 5.5 — Noun-verb command structure
- `serde` / `serde_json` — Serialization
- `tokio` — Async runtime
- `anyhow` / `thiserror` — Error handling
- `tracing` — Logging
- `chrono` — Date/time

**Already in Core `Cargo.toml`:**
- `pm4py` (path) — pm4py-rust library (v0.3.0)
- All dependencies above

### No New Dependencies Required

The pm4py-rust crate provides all needed capabilities internally:
- XES/CSV/JSON parsing
- Process mining algorithms
- Conformance checking
- Statistical analysis
- Model representations (PetriNet, ProcessTree, etc.)

---

## Documentation Plan

### README Section

Add to `BusinessOS/bos/README.md`:

```markdown
## Process Mining with pm4py-rust

Discover, analyze, and validate business processes using pm4py-rust integration.

### Quick Start

Load an event log:
\`\`\`bash
bos pm4py load invoice-logs.xes
\`\`\`

Discover a process model:
\`\`\`bash
bos pm4py discover invoice-logs.xes --algorithm inductive
\`\`\`

Check conformance:
\`\`\`bash
bos pm4py conform invoice-logs.xes
\`\`\`

See full documentation in [Process Mining Guide](./docs/process-mining.md).
```

### Detailed Guide: `docs/process-mining.md`

- Feature overview
- Full command reference
- Examples per algorithm
- Best practices
- Troubleshooting

### API Documentation

Update `core/src/process/mod.rs` with complete rustdoc comments.

---

## Summary Table

| Aspect | Current State | Target State | Effort |
|--------|---------------|--------------|--------|
| **Dependency** | ✓ Present in Cargo.toml | ✓ No change needed | 0h |
| **Core Engine** | ✓ ProcessMiningEngine exists | Enhance with 15+ new methods | 8-10h |
| **CLI Stubs** | ✓ 4 verbs exist | Implement + 11 new verbs | 10-15h |
| **Phase 1 MVP** | Partial stubs | Full working commands | 16-20h |
| **Phase 2 Extended** | None | 5 new commands + tests | 14-18h |
| **Phase 3 Advanced** | None | 5 new commands + export | 18-24h |
| **Phase 4 Predictive** | None | 5+ advanced commands | 20+ h |
| **Signal Theory** | None | All outputs encoded | 3-5h |
| **Tests** | Minimal | 60-80 tests | 12-18h |
| **Documentation** | None | Complete guides + examples | 4-6h |

**Total Effort to Phase 1 MVP:** 16-20 hours
**Total Effort to Phase 3 (production-ready):** 60-80 hours

---

## Decision Record

### Why Path Dependency for pm4py-rust?

**Decision:** Keep pm4py-rust as path dependency at `/Users/sac/chatmangpt/pm4py-rust`

**Rationale:**
1. **Active development**: pm4py-rust is being actively built (v0.3.0)
2. **Local control**: Easy to integrate latest features without waiting for crate release
3. **Testing**: Can test integration immediately as pm4py-rust evolves
4. **Future**: Can publish to crates.io later when stable

**Alternative Rejected:**
- Crates.io dependency: Would require waiting for releases, less flexibility

### Why Noun-Verb Structure?

**Decision:** Extend existing `pm4py` noun rather than new structure

**Rationale:**
1. **Consistency**: Matches all other bos CLI nouns
2. **Discoverability**: `bos pm4py --help` shows all commands
3. **Simplicity**: No new DSL, just function verbs
4. **Extensibility**: Easy to add 40+ commands

---

## Appendix: Full Command Reference (Future)

```
USAGE:
    bos pm4py <COMMAND>

COMMANDS:
    load          Load and analyze event log
    discover      Discover process model
    conform       Check conformance
    align         Alignment-based conformance
    footprints    Footprints-based conformance
    stats         Log statistics
    perf          Performance metrics
    dfg           Directly-Follows Graph
    variants      Trace variants
    predict       Predict next activity or remaining time
    export        Export discovered model
    filter        Filter event log by criteria
    quality       Quality metrics (4-Spectrum)
    declare       Discover Declare constraints
    help          Show help
```

---

**Document Version:** 1.0
**Last Updated:** 2026-03-24
**Author:** Claude Code Agent
**Status:** Ready for Implementation Planning
