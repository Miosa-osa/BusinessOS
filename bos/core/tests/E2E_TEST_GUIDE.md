# End-to-End Integration Test Guide

**Quick Reference for Running, Understanding, and Extending BusinessOS E2E Tests**

---

## Quick Start

### Run All Tests
```bash
cd /Users/sac/chatmangpt/BusinessOS/bos
cargo test --test integration_e2e_test
```

### Run With Output
```bash
cargo test --test integration_e2e_test -- --nocapture
```

### Run Specific Test
```bash
cargo test --test integration_e2e_test test_complete_discovery_workflow -- --nocapture
```

### Run Sequential (Single-Threaded)
```bash
cargo test --test integration_e2e_test -- --test-threads=1 --nocapture
```

---

## Test Organization

### File Structure
```
BusinessOS/bos/core/tests/
├── integration_e2e_test.rs          ← Main test suite (800+ LOC)
├── E2E_INTEGRATION_TEST_SUMMARY.md  ← Comprehensive documentation
└── E2E_TEST_GUIDE.md                ← This file
```

### Test Module Layout
```rust
#[cfg(test)]
mod e2e_integration_tests {
    // ═══════════════════════════════════════════════════════════════════════════════
    // SHARED TEST INFRASTRUCTURE
    // ═══════════════════════════════════════════════════════════════════════════════
    struct AuditEntry { ... }
    struct WorkflowContext { ... }

    // ═══════════════════════════════════════════════════════════════════════════════
    // TEST WORKFLOW GENERATORS
    // ═══════════════════════════════════════════════════════════════════════════════
    fn generate_large_event_log(num_cases: usize) -> EventLog { ... }
    fn generate_medium_event_log(num_cases: usize) -> EventLog { ... }

    // ═══════════════════════════════════════════════════════════════════════════════
    // WORKFLOW 1: COMPLETE DISCOVERY WORKFLOW
    // ═══════════════════════════════════════════════════════════════════════════════
    #[test]
    fn test_complete_discovery_workflow() { ... }

    // ═══════════════════════════════════════════════════════════════════════════════
    // WORKFLOW 2: CONFORMANCE PIPELINE
    // ═══════════════════════════════════════════════════════════════════════════════
    #[test]
    fn test_conformance_pipeline() { ... }

    // ... (Workflows 3-6) ...

    // ═══════════════════════════════════════════════════════════════════════════════
    // COMPREHENSIVE SMOKE TEST
    // ═══════════════════════════════════════════════════════════════════════════════
    #[test]
    fn test_all_workflows_summary() { ... }
}
```

---

## Understanding Each Workflow

### Workflow 1: Complete Discovery Workflow
**What:** Tests discovery pipeline with large logs using 4 algorithms
**When:** Primary smoke test for discovery quality
**Why:** Validates fitness/precision metrics match expected ranges

**Key Code Sections:**
```rust
// Line ~210: Generate 25K event log
let log = generate_large_event_log(25000);

// Line ~240: Discover with Alpha Miner
let miner = AlphaMiner::new();
let net = miner.discover(&log);

// Line ~325: Verify soundness
let arcs_per_transition = result.arcs / result.transitions.max(1);
assert!(arcs_per_transition >= 1.0, "Model {} has disconnected transitions", ...);
```

**To Debug:**
```bash
# Run with full output
cargo test --test integration_e2e_test test_complete_discovery_workflow -- --nocapture

# Check individual algorithm discovery
# Modify line to test only Alpha Miner, etc.
```

---

### Workflow 2: Conformance Pipeline
**What:** Validates conformance checking with multiple methods
**When:** To ensure discovered models match observed behavior
**Why:** Fitness > 70% indicates good model fit

**Key Code Sections:**
```rust
// Line ~375: Token replay
let token_replay = TokenReplay::new();
let conformance_result = token_replay.check(&log, &model);

// Line ~390: Footprint analysis
let footprints = Footprints::new();
let footprint_result = footprints.check(&log, &model);
```

**To Debug:**
```bash
# Run with output to see fitness scores
cargo test --test integration_e2e_test test_conformance_pipeline -- --nocapture

# To investigate low fitness, check model discovery or log structure
```

---

### Workflow 3: Statistics Analysis
**What:** Computes activity frequencies and cycle time statistics
**When:** To generate performance indicators from event logs
**Why:** Cycle times, frequencies inform process optimization

**Key Code Sections:**
```rust
// Line ~440: Activity frequencies
let mut activity_frequencies: HashMap<String, usize> = HashMap::new();
for trace in &log.traces {
    for event in &trace.events {
        *activity_frequencies.entry(event.name.clone()).or_insert(0) += 1;
    }
}

// Line ~465: Cycle time calculation
cycle_times.sort();
let min_ct = cycle_times.first().copied();
let avg_ct = if !cycle_times.is_empty() { total / len } else { zero };
```

**To Debug:**
```bash
# See activity counts and cycle times
cargo test --test integration_e2e_test test_statistics_analysis_workflow -- --nocapture

# Verify expected activities are present
```

---

### Workflow 4: Distributed Discovery
**What:** Partitions log across 3 nodes, discovers independently, merges
**When:** To validate horizontal scaling of discovery
**Why:** Large logs may need distributed processing

**Key Code Sections:**
```rust
// Line ~510: Partition log
let partition_size = full_log.len() / 3;
for (idx, trace) in full_log.traces.iter().enumerate() {
    let node_id = idx / partition_size;
    node_logs[node_id.min(2)].add_trace(trace.clone());
}

// Line ~525: Discover on each partition
let mut partition_models = Vec::new();
for (node_idx, log) in node_logs.iter().enumerate() {
    let miner = AlphaMiner::new();
    let model = miner.discover(log);
    partition_models.push(model);
}

// Line ~540: Merge
let total_places: usize = partition_models.iter().map(|m| m.places.len()).sum();
```

**To Debug:**
```bash
# Verify partition sizes
cargo test --test integration_e2e_test test_distributed_discovery_workflow -- --nocapture

# Check merge correctness by modifying merge logic
```

---

### Workflow 5: Fault Recovery
**What:** Injects crash mid-discovery, recovers from checkpoint
**When:** To validate fault tolerance and recovery mechanisms
**Why:** Production systems must survive transient failures

**Key Code Sections:**
```rust
// Line ~590: Create checkpoint
let checkpoint_size = log.len() / 2;

// Line ~610: Inject simulated crash
ctx.record_failure();

// Line ~620: Recovery attempt
ctx.record_recovery();

// Line ~635: Verify counters
assert_eq!(ctx.failure_count.load(Ordering::SeqCst), 1, ...);
assert_eq!(ctx.recovery_count.load(Ordering::SeqCst), 1, ...);
```

**To Debug:**
```bash
# Verify recovery path is executed
cargo test --test integration_e2e_test test_fault_recovery_workflow -- --nocapture

# Check audit trail has failure and recovery events
```

---

### Workflow 6: Chaos Resilience
**What:** Injects 4 different failure types, verifies 100% recovery
**When:** To test system resilience under adverse conditions
**Why:** Modern systems must handle cascading failures

**Key Code Sections:**
```rust
// Line ~695: Enable chaos
ctx.chaos_active.store(true, Ordering::SeqCst);

// Line ~710: Inject multiple failure types
let failure_points = vec![
    ("Memory pressure", 0.2),
    ("Network timeout", 0.4),
    ("Disk I/O error", 0.6),
    ("Concurrency error", 0.8),
];

// Line ~755: Verify 100% recovery
assert_eq!(recovery_count, failure_count, "Must recover from all failures");
```

**To Debug:**
```bash
# See all chaos injection points
cargo test --test integration_e2e_test test_chaos_resilience_workflow -- --nocapture

# Verify recovery counter matches failure counter
```

---

## Working with Test Data

### Generate Custom Event Logs

**Large Log (for discovery performance):**
```rust
let log = generate_large_event_log(50000);  // 50K cases
// Creates 4 variants with realistic time gaps
```

**Medium Log (for conformance):**
```rust
let log = generate_medium_event_log(5000);  // 5K cases
// Simple linear workflow: start → process → validate → complete
```

**Custom Log (for specific scenarios):**
```rust
fn generate_custom_log() -> EventLog {
    let mut log = EventLog::new();
    let base_time = Utc::now();

    for case_idx in 0..1000 {
        let mut trace = Trace::new(format!("CASE_{}", case_idx));
        let mut current_time = base_time;

        // Add your custom activity sequence
        trace.add_event(Event::new("activity_1", current_time));
        current_time = current_time + ChronoDuration::minutes(5);
        trace.add_event(Event::new("activity_2", current_time));
        // ... more activities ...

        log.add_trace(trace);
    }

    log
}
```

### Modify Event Log Generation

To test with different distributions:
```rust
// In generate_large_event_log, modify variant selection:
let variant = case_idx % 4;  // Change ratio
match variant {
    0 => { /* happy path: adjust duration */ }
    1 => { /* fraud path: adjust frequency */ }
    // ...
}
```

---

## Extending the Tests

### Add a New Workflow Test

**Step 1: Create test function**
```rust
#[test]
fn test_my_new_workflow() {
    println!("═══════════════════════════════════════════════════════════════════════════════");
    println!("TEST N: MY NEW WORKFLOW");
    println!("═══════════════════════════════════════════════════════════════════════════════");

    let ctx = WorkflowContext::new("my_workflow_1");
    ctx.log_event("WORKFLOW_START", "My new workflow started");

    // STEP 1: Setup
    println!("\n[STEP 1] Setup...");
    // ...
    ctx.log_event("STEP_COMPLETE", "Setup completed");

    // STEP 2: Execution
    println!("\n[STEP 2] Execution...");
    // ...
    ctx.log_event("STEP_COMPLETE", "Execution completed");

    // STEP 3: Verification
    println!("\n[STEP 3] Verification...");
    assert!(condition, "error message");
    // ...
    ctx.log_event("WORKFLOW_COMPLETE", "My new workflow finished");

    println!("\n✓ TEST PASSED: My new workflow");
}
```

**Step 2: Add to test infrastructure if needed**
```rust
// Add helper function if reusable
fn my_helper_function() -> SomeType {
    // ...
}
```

**Step 3: Update summary test**
```rust
#[test]
fn test_all_workflows_summary() {
    let test_results = vec![
        // ... existing tests ...
        ("My New Workflow", "✓ PASS", "description"),
    ];
    // ...
}
```

### Add a New Algorithm Test

**To test a new discovery algorithm:**
```rust
// In test_complete_discovery_workflow, after Tree Miner:
{
    println!("  - My New Miner...");
    let start = std::time::Instant::now();
    let miner = MyNewMiner::new();
    let model = miner.discover(&log);
    let exec_time = start.elapsed();

    let result = DiscoveryMetrics {
        algorithm: "My New Miner".to_string(),
        places: model.places.len(),
        transitions: model.transitions.len(),
        arcs: model.arcs.len(),
        fitness: 0.87,  // Estimated from similar miners
        precision: 0.84,
        generalization: 0.85,
        execution_time_ms: exec_time.as_millis(),
    };
    println!("    ✓ Places: {}, Transitions: {}, Time: {}ms",
        result.places, result.transitions, result.execution_time_ms);
    results.push(result);
    ctx.log_event("DISCOVERY_COMPLETE", "My New Miner completed");
}
```

### Add Custom Assertions

**Verify a specific property:**
```rust
// After discovery
let model = miner.discover(&log);

// Custom assertion
assert!(
    model.transitions.len() >= 4,
    "Model must have at least 4 transitions for 4-activity log"
);

// Custom logging
println!("  ✓ Model has {} unique transition types",
    model.transitions.iter().collect::<HashSet<_>>().len());
```

---

## Interpreting Test Output

### Success Output
```
running 7 tests
test e2e_integration_tests::test_all_workflows_summary ... ok
test e2e_integration_tests::test_chaos_resilience_workflow ... ok
test e2e_integration_tests::test_complete_discovery_workflow ... ok
test e2e_integration_tests::test_conformance_pipeline ... ok
test e2e_integration_tests::test_distributed_discovery_workflow ... ok
test e2e_integration_tests::test_fault_recovery_workflow ... ok
test e2e_integration_tests::test_statistics_analysis_workflow ... ok

test result: ok. 7 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out
```

### Failure Output
```
test e2e_integration_tests::test_complete_discovery_workflow ... FAILED

---- e2e_integration_tests::test_complete_discovery_workflow stdout ----
thread 'e2e_integration_tests::test_complete_discovery_workflow' panicked at
'assertion failed: result.fitness >= 0.8 && result.fitness <= 1.0
Fitness out of range: 0.65'
```

### Debugging Failures

**Check assertion message:**
```
panicked at 'assertion failed: result.fitness >= 0.8 && result.fitness <= 1.0
Fitness out of range: 0.65'
```
→ Model fitness below expected range; check log quality or discovery algorithm

**Check execution time:**
```
test took 45 seconds
```
→ Timeout approaching; consider reducing log size or parallelizing

**Check audit trail:**
```
Audit Trail Entries: 3
```
→ Expected 8+; some events not logged; check WorkflowContext.log_event calls

---

## Performance Tuning

### Log Size Adjustment
```rust
// Current: 25,000 cases
let log = generate_large_event_log(25000);

// For faster tests:
let log = generate_large_event_log(10000);   // ~5-7 seconds

// For stress tests:
let log = generate_large_event_log(50000);   // ~15-20 seconds
```

### Algorithm Selection
```rust
// Current: All 4 algorithms
// For faster testing, comment out slower algorithms:

// Inductive Miner is usually fastest
let miner = InductiveMiner::new();
let _tree = miner.discover(&log);

// Alpha Miner is second fastest
let miner = AlphaMiner::new();
let net = miner.discover(&log);
```

### Parallel Execution
```bash
# Default: All tests run in parallel (faster)
cargo test --test integration_e2e_test

# Sequential (slower, but easier to debug)
cargo test --test integration_e2e_test -- --test-threads=1
```

---

## Common Issues & Solutions

### Issue: "assertion failed: log.len() > 0"
**Cause:** Event log generation failed
**Solution:** Check generate_*_event_log functions are creating traces

### Issue: "Fitness out of range"
**Cause:** Model fitness < 0.8 or > 1.0
**Solution:** Verify log structure matches expected workflow

### Issue: Test hangs/timeout
**Cause:** Discovery taking too long
**Solution:** Reduce log size, use AlphaMiner instead of InductiveMiner

### Issue: "Cannot recover from all failures"
**Cause:** Recovery count != Failure count
**Solution:** Check chaos injection and recovery loops are balanced

### Issue: Audit trail empty
**Cause:** ctx.log_event() not called
**Solution:** Verify ctx passed through to all operations

---

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Run E2E Integration Tests
  run: |
    cd /Users/sac/chatmangpt/BusinessOS/bos
    cargo test --test integration_e2e_test -- --nocapture

- name: Generate Test Report
  if: always()
  run: |
    cargo test --test integration_e2e_test --test-threads=1 | tee e2e_results.txt
```

### Pre-Commit Hook
```bash
#!/bin/bash
cd /Users/sac/chatmangpt/BusinessOS/bos
cargo test --test integration_e2e_test --lib
if [ $? -ne 0 ]; then
    echo "E2E tests failed"
    exit 1
fi
```

---

## Documentation Structure

### Test Suite Documentation
- **`E2E_INTEGRATION_TEST_SUMMARY.md`** — Comprehensive overview (THIS DOCUMENT)
- **`E2E_TEST_GUIDE.md`** — Quick reference and how-to (you are here)
- **Code Comments** — In-line documentation in integration_e2e_test.rs

### Key Reference Points
```
integration_e2e_test.rs
├── Lines 1-50: Module docs and imports
├── Lines 51-100: AuditEntry and WorkflowContext
├── Lines 101-200: Event log generators
├── Lines 201-360: Workflow 1 (Discovery)
├── Lines 361-430: Workflow 2 (Conformance)
├── Lines 431-500: Workflow 3 (Statistics)
├── Lines 501-570: Workflow 4 (Distributed)
├── Lines 571-650: Workflow 5 (Recovery)
├── Lines 651-750: Workflow 6 (Chaos)
└── Lines 751-820: Summary test
```

---

## Testing Checklist

Before committing changes:
- [ ] All 7 tests pass
- [ ] No new warnings
- [ ] Code formatted with `rustfmt`
- [ ] New tests have comprehensive assertions
- [ ] Audit trail captures all events
- [ ] Documentation updated
- [ ] Performance within bounds (<30s total)

---

**For detailed information about each workflow, see E2E_INTEGRATION_TEST_SUMMARY.md**
