# BOS Process Mining Test Suite Guide

## Overview

This document describes the `bos_process_mining_test.py` test suite, which adapts pm4py test cases to use "bos" (BusinessOS) as the case study with **account lifecycle workflows** instead of generic process mining examples.

**File Location:** `/Users/sac/chatmangpt/BusinessOS/bos/tests/bos_process_mining_test.py`

## Philosophy: Account-Based Case Study

Instead of generic activities like "A → B → C", this test suite uses realistic **account lifecycle workflows**:

```
Standard Account Path:
created → verified → activated → used (3x) → closed
Duration: ~30 days

Suspension Variant:
created → verified → activated → used → suspended → reactivated → used → closed
Duration: ~60 days

Abnormal Pattern (Fraud Risk):
created → activated → used → closed
(Missing verification step)
Duration: ~5 days
```

## Test Structure

### 1. AccountEventLogGenerator

**Purpose:** Creates realistic account event logs for testing.

**Key Methods:**

| Method | Purpose | Pattern |
|--------|---------|---------|
| `create_standard_account_trace()` | Standard 30-day account lifecycle | created → verified → activated → used* → closed |
| `create_suspension_account_trace()` | Account with suspension/reactivation | created → verified → activated → used → suspended → reactivated → used → closed |
| `create_abnormal_account_trace()` | Fraud-risk pattern (skipped verification) | created → activated → used → closed |
| `create_account_log_mixed()` | Mixed log with all patterns | Combines all three patterns |

**Usage Example:**
```python
# Create log with 40 standard, 30 suspended, 10 abnormal accounts
log = AccountEventLogGenerator.create_account_log_mixed(
    num_standard=40,
    num_suspension=30,
    num_abnormal=10
)
```

### 2. Test Classes

#### TestAccountProcessDiscovery
Discovers process models from account event logs.

**Test Cases:**
- `test_discover_standard_account_lifecycle()` - Learn standard 30-day account flow
- `test_discover_with_suspension_variant()` - Learn model with suspension path
- `test_alpha_miner_account_discovery()` - Use Alpha Miner algorithm
- `test_heuristic_miner_account_discovery()` - Use Heuristic Miner algorithm
- `test_inductive_miner_account_discovery()` - Use Inductive Miner algorithm
- `test_discover_anomalous_patterns()` - Learn to identify fraud patterns

**Expected Behavior:**
- Discover model captures standard account path
- Model includes both standard and suspension variants
- Anomalous patterns are identifiable as deviations

---

#### TestAccountLifecycleConformance
Validates that accounts follow expected process models.

**Test Cases:**
- `test_conform_standard_accounts()` - 100% fitness on conformant accounts
- `test_conform_with_suspension_variant()` - >90% fitness with multiple paths
- `test_detect_skipped_verification()` - Low fitness identifies fraud risk
- `test_conformance_fitness_metrics()` - Proper fitness scoring
- `test_footprints_conformance_account_workflow()` - Footprints checker validation

**Key Metric: Fitness Score**
```
Fitness = (Conformant Traces) / (Total Traces)
Range: 0.0 (no conformance) to 1.0 (perfect conformance)
```

---

#### TestAccountProcessStatistics
Analyzes account workflow performance metrics.

**Test Cases:**
- `test_basic_account_statistics()` - Count traces, events, variants
- `test_account_cycle_time_analysis()` - Duration from creation → closure
- `test_activity_frequency_analysis()` - How often each activity occurs
- `test_account_lifecycle_variants()` - Identify distinct process paths
- `test_bottleneck_analysis()` - Where accounts get stuck
- `test_account_throughput_metrics()` - Accounts per day / processing time

**Expected Statistics:**

| Metric | Standard (30 accounts) | Mixed (80 accounts) |
|--------|----------------------|-------------------|
| Total Traces | 30 | 80 |
| Total Events | ~210 (7 events/trace) | ~550 |
| Created Activity | 30 | 80 |
| Used Activity | ~90 | ~220 |
| Variants | 1 | 3+ |
| Avg Cycle Time | ~30 days | ~35 days |

---

#### TestAccountFileRepresentation
Tests account event log serialization and I/O.

**Test Cases:**
- `test_account_log_json_export()` - Export to JSON format
- `test_account_log_json_import()` - Import from JSON format
- `test_account_log_csv_export()` - Export to CSV format
- `test_account_log_csv_import()` - Import from CSV format
- `test_account_event_record_structure()` - Validate event fields

**Account Event Record Schema:**
```json
{
  "account_id": "account_0",
  "activity": "account_created",
  "timestamp": "2024-01-01T10:30:00Z",
  "resource": "system_api",
  "metadata": {
    "region": "US",
    "account_type": "premium"
  }
}
```

**CSV Format:**
```csv
account_id,activity,timestamp,resource
account_0,account_created,2024-01-01T10:30:00Z,system_api
account_0,account_verified,2024-01-01T12:30:00Z,verification_service
account_0,account_activated,2024-01-01T13:30:00Z,system_api
```

---

#### TestBOSCLIIntegration
Tests BOS command-line interface integration.

**Test Cases:**
- `test_bos_discover_command()` - `bos discover --input accounts.json --algorithm alpha`
- `test_bos_conform_command()` - `bos conform --input accounts.json --model model.json`
- `test_bos_stats_command()` - `bos stats --input accounts.json`
- `test_bos_export_command()` - `bos export --input accounts.json --format csv`
- `test_bos_workflow_end_to_end()` - Full pipeline test

**Example CLI Usage:**
```bash
# Discover process model
bos discover \
  --input account_logs.json \
  --algorithm alpha \
  --output model.json

# Check conformance
bos conform \
  --input account_logs.json \
  --model model.json \
  --output conformance_report.json

# Get statistics
bos stats \
  --input account_logs.json \
  --output stats.json

# Export for visualization
bos export \
  --input account_logs.json \
  --format csv \
  --output account_events.csv
```

---

#### TestAccountProcessMiningIntegration
End-to-end integration tests.

**Test Cases:**
- `test_full_account_mining_pipeline()` - Complete workflow
- `test_account_anomaly_detection()` - Identify abnormal accounts
- `test_process_model_evolution()` - Model stability over time

---

## Data Flow Diagram

```
┌─────────────────────────────────────────┐
│ AccountEventLogGenerator                │
│ - create_standard_account_trace()       │
│ - create_suspension_account_trace()     │
│ - create_abnormal_account_trace()       │
│ - create_account_log_mixed()            │
└────────────────┬────────────────────────┘
                 │
                 ├──→ EventLog (pm4py_rust)
                 │
         ┌───────┴────────┬─────────────┬──────────────┐
         │                │             │              │
         v                v             v              v
    Discovery         Conformance    Statistics     File I/O
    (AlphaMiner)     (Footprints)    (LogStats)    (JSON/CSV)
         │                │             │              │
         └───────┬────────┴─────────────┴──────────────┘
                 │
                 v
         Test Assertion & Metrics
```

## Test Data Characteristics

### Standard Account Lifecycle (40% of mixed logs)
```
Trace: created → verified → activated → used → used → used → closed
Duration: ~31 days
Events: 7 per trace
Pattern: Ideal account onboarding and lifecycle
```

### Suspension Variant (30% of mixed logs)
```
Trace: created → verified → activated → used → suspended → reactivated → used → closed
Duration: ~62 days
Events: 8 per trace
Pattern: Account with compliance issue or suspension recovery
```

### Abnormal/Anomalous (10% of mixed logs)
```
Trace: created → activated → used → closed
Duration: ~5 days
Events: 4 per trace
Pattern: Fraud risk (skipped verification), suspicious rapid closure
```

## Execution Strategy

### Phase 1: Structure Validation (Current)
✅ File created with:
- Test class structure defined
- Test methods stubbed (pass statements)
- Comments indicating implementation locations
- AccountEventLogGenerator fully implemented
- Helper classes defined

### Phase 2: Implementation (Next)
Will implement test method bodies:
1. Import pm4py_rust bindings
2. Create sample account logs
3. Run discovery/conformance/statistics
4. Assert expected outcomes

### Phase 3: CLI Integration (Future)
Will add BOS CLI command testing:
1. Generate account files
2. Run bos CLI commands
3. Parse results
4. Validate output

### Phase 4: Smoke Tests (Future)
Add lightweight smoke tests for CI/CD pipeline.

## Running Tests

```bash
# Run all tests
pytest bos/tests/bos_process_mining_test.py -v

# Run specific test class
pytest bos/tests/bos_process_mining_test.py::TestAccountProcessDiscovery -v

# Run specific test
pytest bos/tests/bos_process_mining_test.py::TestAccountProcessDiscovery::test_discover_standard_account_lifecycle -v

# Run with output capture disabled (see print statements)
pytest bos/tests/bos_process_mining_test.py -v -s

# Run only skipped tests (when pm4py_rust is available)
pytest bos/tests/bos_process_mining_test.py -v --runxfail
```

## Files Referenced

| File | Purpose |
|------|---------|
| `/Users/sac/chatmangpt/pm4py-rust/tests/test_python_bindings.py` | Original pm4py test template |
| `/Users/sac/chatmangpt/pm4py-rust/tests/businessos_http_integration_tests.py` | BusinessOS API integration pattern |
| `/Users/sac/chatmangpt/BusinessOS/bos/tests/bos_process_mining_test.py` | **This test suite** |

## Key Differences from pm4py Tests

| Aspect | pm4py Tests | BOS Tests |
|--------|-------------|----------|
| **Case Study** | Generic "A → B → C" | Real account lifecycle workflows |
| **Event Names** | activity_A, activity_B | account_created, account_verified, etc. |
| **Trace IDs** | case_0, case_1 | account_0, account_1 |
| **Business Context** | Process mining theory | Account onboarding, compliance, fraud detection |
| **Statistics Focus** | Basic log metrics | Cycle time, throughput, bottlenecks |
| **Conformance Use Case** | Model validation | Fraud risk detection, policy compliance |
| **File Formats** | Generic event logs | Account event records (account_id, activity, timestamp) |

## Future Extensions

1. **Account Attributes:** Add account_type, region, risk_score to traces
2. **Resource Tracking:** Include processor/system information per event
3. **Performance Benchmarks:** Measure algorithm speed on account logs
4. **Visualization:** Generate process model diagrams for account workflows
5. **Compliance Reporting:** Create reports for account audit trails
6. **Anomaly Scoring:** Rank accounts by conformance/anomaly score

## References

- **pm4py Documentation:** https://pm4py.fit.fraunhofer.de/
- **pm4py-rust GitHub:** https://github.com/pm4py/pm4py-rust
- **Process Mining Theory:** van der Aalst, W. M. P. (2016). Process Mining
- **Signal Theory (BOS Context):** `docs/diataxis/explanation/signal-theory-complete.md`
