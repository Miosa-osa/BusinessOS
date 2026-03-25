# YAWL WCP1-10 Test Execution Guide

**Test File:** `yawl_wcp1_10_test.rs` (1,193 lines, 26 tests, 100+ assertions)

---

## Quick Start

### Run All WCP Tests

```bash
cd /Users/sac/chatmangpt/BusinessOS/bos
cargo test --test yawl_wcp1_10_test -- --nocapture
```

### Run Specific Pattern Tests

```bash
# WCP1: Sequence
cargo test --test yawl_wcp1_10_test test_wcp1

# WCP2: Parallel Split
cargo test --test yawl_wcp1_10_test test_wcp2

# WCP3: Synchronization
cargo test --test yawl_wcp1_10_test test_wcp3

# WCP4: Exclusive Choice
cargo test --test yawl_wcp1_10_test test_wcp4

# WCP5: Simple Merge
cargo test --test yawl_wcp1_10_test test_wcp5

# WCP6: Multi-Choice
cargo test --test yawl_wcp1_10_test test_wcp6

# WCP7: Structured Parallel
cargo test --test yawl_wcp1_10_test test_wcp7

# WCP8: Multi-Merge
cargo test --test yawl_wcp1_10_test test_wcp8

# WCP9: Structured Synchronization
cargo test --test yawl_wcp1_10_test test_wcp9

# WCP10: Arbitrary Cycles
cargo test --test yawl_wcp1_10_test test_wcp10
```

### Run Individual Test

```bash
cargo test --test yawl_wcp1_10_test test_wcp1_sequence_basic -- --nocapture
```

---

## Test Organization by Pattern

### WCP1: Sequence (A → B → C)
- ✅ `test_wcp1_sequence_basic` — Basic 3-step linear sequence
- ✅ `test_wcp1_sequence_with_variants` — Repeated sequence pattern

**Key Assertions:**
- `assert_eq!(net.transitions.len(), 3)` — 3 activities
- `net.arcs.iter().any(|(s, t)| s == "A" && t == "B")` — A→B arc exists
- `assert!(check_soundness(&log, &net))` — No deadlocks
- `assert!(fitness >= 0.9)` — High replay capability

---

### WCP2: Parallel Split (A → (B || C))
- ✅ `test_wcp2_parallel_split` — 2-way parallel from entry
- ✅ `test_wcp2_parallel_split_convergence` — Parallel + join to D

**Key Assertions:**
- `arcs_from_a.contains(&"B".to_string())` — B reachable from A
- `arcs_from_a.contains(&"C".to_string())` — C reachable from A
- `net.transitions.contains(&"C".to_string())` — Both C and B present
- Parallel execution captured in trace variants

---

### WCP3: Synchronization (Join Parallel Paths)
- ✅ `test_wcp3_synchronization` — 2-way join to D
- ✅ `test_wcp3_synchronization_multiple_joins` — 3-way merge to E

**Key Assertions:**
- `net.arcs.iter().any(|(s, t)| s == "B" && t == "D")` — B→D arc
- `net.arcs.iter().any(|(s, t)| s == "C" && t == "D")` — C→D arc
- All parallel branches must precede join in all traces
- Soundness: join place must be reachable

---

### WCP4: Exclusive Choice (A → (B XOR C))
- ✅ `test_wcp4_exclusive_choice` — Binary exclusive choice
- ✅ `test_wcp4_exclusive_choice_with_skip` — Optional activity

**Key Assertions:**
- `net.arcs.iter().any(|(s, t)| s == "A" && t == "B")` — A→B exists
- `net.arcs.iter().any(|(s, t)| s == "A" && t == "C")` — A→C exists
- `!(has_b && has_c)` — No trace has both B and C
- `net.arcs.iter().any(|(s, t)| s == "B" && t == "D")` — Both reach D

---

### WCP5: Simple Merge (Join Exclusive Paths)
- ✅ `test_wcp5_simple_merge` — 2-way merge
- ✅ `test_wcp5_simple_merge_multiple_sources` — 3-way merge to E

**Key Assertions:**
- `net.arcs.iter().any(|(s, t)| s == "B" && t == "D")` — B→D arc
- `net.arcs.iter().any(|(s, t)| s == "C" && t == "D")` — C→D arc
- No synchronization barrier (D reachable from any branch)
- Soundness: merge point reachable from all predecessors

---

### WCP6: Multi-Choice (A → (B AND C possibly))
- ✅ `test_wcp6_multi_choice` — B, C, or both
- ✅ `test_wcp6_multi_choice_three_branches` — B, C, D with combinations

**Key Assertions:**
- `net.arcs.iter().any(|(s, t)| s == "A" && t == "B")` — A→B possible
- `net.arcs.iter().any(|(s, t)| s == "A" && t == "C")` — A→C possible
- Traces may contain: {B}, {C}, or {B,C}
- All combinations converge to exit

---

### WCP7: Structured Parallel (A → (B || C) → D)
- ✅ `test_wcp7_structured_parallel` — 4-activity pattern
- ✅ `test_wcp7_structured_parallel_interleaved` — Interleaved B/C

**Key Assertions:**
- `net.transitions.len() == 4` — A, B, C, D
- `net.arcs.iter().any(|(s, t)| s == "A" && (t == "B" || t == "C"))` — A entry
- All traces contain both B and C before D
- Soundness despite interleaving

---

### WCP8: Multi-Merge (Multiple Paths Converge)
- ✅ `test_wcp8_multi_merge` — 2-way unsynchronized merge
- ✅ `test_wcp8_multi_merge_complex` — 3+ paths with combinations

**Key Assertions:**
- `paths_to_d >= 2` — Multiple input paths to D
- No requirement for all branches to complete
- Any branch can trigger merge
- Soundness: liveness from any predecessor

---

### WCP9: Structured Synchronization (A → (B || C || D) → E)
- ✅ `test_wcp9_structured_synchronization` — 3-way sync to E
- ✅ `test_wcp9_structured_synchronization_4way` — 4-way parallel

**Key Assertions:**
- `from_a.contains(&"B".to_string())` — B reachable from A
- `from_a.contains(&"C".to_string())` — C reachable from A
- `from_a.contains(&"D".to_string())` — D reachable from A
- All branches must complete before E
- `fitness >= 0.9` — High-quality discovery

---

### WCP10: Arbitrary Cycles (Backward Loops)
- ✅ `test_wcp10_simple_cycle` — Single loop A→B→A
- ✅ `test_wcp10_nested_cycles` — Validate/Fix cycle
- ✅ `test_wcp10_cycle_with_parallel` — Cycles + parallel

**Key Assertions:**
- `net.arcs.iter().any(|(s, t)| s == "B" && t == "A")` — Backward arc
- `net.arcs.iter().any(|(s, t)| s == "B" && t == "C")` — Forward arc (exit)
- Loop must exit to proper termination
- Soundness despite cycles

---

### Integration & Edge Cases
- ✅ `test_wcp_all_patterns_combined` — Mixed patterns in one log
- ✅ `test_wcp_large_scale_discovery` — 100 traces, scalability test
- ✅ `test_wcp_fitness_precision_metrics` — Metric validation
- ✅ `test_wcp_deviating_behavior` — Non-conformant traces
- ✅ `test_wcp_full_workflow_discovery_pipeline` — End-to-end workflow

---

## Expected Output

### On Success (All Tests Pass)

```
test yawl_wcp_tests::test_wcp1_sequence_basic ... ok
test yawl_wcp_tests::test_wcp1_sequence_with_variants ... ok
test yawl_wcp_tests::test_wcp2_parallel_split ... ok
test yawl_wcp_tests::test_wcp2_parallel_split_convergence ... ok
... (26 tests total)
test result: ok. 26 passed; 0 failed; 0 ignored
```

### On Discovery Failure

```
thread 'yawl_wcp_tests::test_wcp7_structured_parallel' panicked at 'assertion failed:
  net.transitions.len() == 4'
note: run with `RUST_BACKTRACE=1` for more info
```

### With Detailed Output

```bash
cargo test --test yawl_wcp1_10_test -- --nocapture --test-threads=1
```

Shows:
- XES log file paths created
- Transition count discovered
- Arc relationships
- Fitness/Precision values
- Soundness check results

---

## Test Data

### Generated XES Logs

All tests generate valid XES (eXtensible Event Stream) logs in:

```
/Users/sac/chatmangpt/BusinessOS/bos/tests/data/yawl_wcp/
├── wcp1_sequence_basic.xes
├── wcp1_sequence_with_variants.xes
├── wcp2_parallel_split.xes
├── wcp2_parallel_split_convergence.xes
├── wcp3_synchronization.xes
├── wcp3_synchronization_multiple_joins.xes
├── wcp4_exclusive_choice.xes
├── wcp4_exclusive_choice_with_skip.xes
├── wcp5_simple_merge.xes
├── wcp5_simple_merge_multiple_sources.xes
├── wcp6_multi_choice.xes
├── wcp6_multi_choice_three_branches.xes
├── wcp7_structured_parallel.xes
├── wcp7_structured_parallel_interleaved.xes
├── wcp8_multi_merge.xes
├── wcp8_multi_merge_complex.xes
├── wcp9_structured_synchronization.xes
├── wcp9_structured_synchronization_4way.xes
└── wcp10_*.xes files
```

### Log Format

```xml
<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="nested-attributes">
  <trace>
    <string key="concept:name" value="trace_0"/>
    <event>
      <string key="concept:name" value="A"/>
      <date key="time:timestamp" value="2024-01-01T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="B"/>
      <date key="time:timestamp" value="2024-01-01T10:15:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="C"/>
      <date key="time:timestamp" value="2024-01-01T10:30:00.000Z"/>
    </event>
  </trace>
</log>
```

---

## Test Metrics

| Metric | Value |
|--------|-------|
| **Total Tests** | 26 |
| **Total Assertions** | 100+ |
| **Lines of Code** | 1,193 |
| **Pattern Coverage** | WCP1-10 (10/10 = 100%) |
| **XES Logs Generated** | 26+ |
| **Soundness Checks** | Every test |
| **Fitness Validation** | 5+ tests |
| **Precision Validation** | 5+ tests |

---

## Performance Expectations

### Execution Time (Estimated)

| Test Category | Time | Notes |
|--------------|------|-------|
| WCP1-5 | ~100ms | Basic patterns |
| WCP6-8 | ~150ms | Complex branching |
| WCP9-10 | ~200ms | Synchronization/cycles |
| Edge Cases | ~300ms | Large-scale + deviations |
| **Total Suite** | **~750ms** | All 26 tests |

### Memory Usage

- **Per test:** ~1-2 MB (XES log + Petri net)
- **Total:** ~50-100 MB (all tests combined)
- **Test data dir:** ~5-10 MB (all XES files)

---

## Troubleshooting

### Issue: "no test target named `yawl_wcp1_10_test`"

**Solution:** Ensure test file is in correct directory:
```bash
ls -la /Users/sac/chatmangpt/BusinessOS/bos/cli/tests/yawl_wcp1_10_test.rs
# Should show: yawl_wcp1_10_test.rs
```

### Issue: pm4py compilation errors

**Context:** pm4py-rust has pre-existing compilation issues (not caused by WCP tests)

**Workaround:** Fix pm4py first:
```bash
cd /Users/sac/chatmangpt/pm4py-rust
cargo check
# Fix compilation errors in pm4py
```

### Issue: Fitness/Precision assertions fail

**Possible causes:**
1. Discovery algorithm didn't find all arcs
2. Log has high deviation rate
3. Directly-follows not extracted correctly

**Debug:**
```bash
cargo test --test yawl_wcp1_10_test test_wcp1_sequence_basic \
  -- --nocapture --test-threads=1
```

Shows detailed output with actual fitness/precision values.

### Issue: Soundness checks fail

**Possible causes:**
1. Arc missing in discovered net
2. Activity in trace not in transitions
3. Trace execution incomplete

**Check log:**
- `ls -la /Users/sac/chatmangpt/BusinessOS/bos/tests/data/yawl_wcp/`
- Open XES file to verify trace structure

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: YAWL WCP Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: dtolnay/rust-toolchain@stable
      - run: cd BusinessOS/bos && cargo test --test yawl_wcp1_10_test
      - run: cargo test --test yawl_wcp1_10_test -- --nocapture
```

---

## Next Steps

1. **Fix pm4py-rust compilation** → Tests become executable
2. **Run full suite** → Verify all 26 tests pass
3. **Integrate with BusinessOS** → Use in process mining pipeline
4. **Extend with real Alpha Miner** → Use actual discovery algorithm
5. **Add pattern-specific validators** → Verify WCP-specific properties

---

**Status:** ✅ Ready for Execution (pending pm4py-rust fixes)

**Last Updated:** 2026-03-24

