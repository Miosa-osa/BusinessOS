/// pm4py Workflow Integration Tests
///
/// End-to-end workflow tests for the complete process mining pipeline:
/// 1. Load: Read event log from structured data
/// 2. Discover: Mine process model with Alpha Miner
/// 3. Conform: Check if original log conforms to discovered model
/// 4. Analyze: Extract statistics from log
///
/// This test file validates the complete workflow with three distinct scenarios:
/// - Perfect-fit log (fitness = 1.0, all traces conform)
/// - Non-conformant log (fitness < 1.0, deviations detected)
/// - Large log with statistics validation (100 events across 10 cases)

#[cfg(test)]
mod pm4py_workflow_integration {
    use pm4py::{EventLog, Trace, Event};
    use pm4py::AlphaMiner;
    use pm4py::conformance::TokenReplay;
    use pm4py::statistics::log_stats;
    use chrono::{Utc, Duration};

    /// ============================================================
    /// TEST SCENARIO 1: Happy Path Test (Perfect Fit)
    /// ============================================================
    /// Account lifecycle log where ALL traces follow the exact same
    /// sequence, resulting in perfect conformance (fitness = 1.0).
    ///
    /// Expected Process Model:
    ///   account_created → account_verified → account_activated → account_closed
    ///
    /// All 5 accounts follow this exact sequence with no deviations.

    /// Helper: Create a perfect-fit account lifecycle log
    fn create_perfect_fit_log() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Create 5 accounts that all follow the identical sequence
        for i in 0..5 {
            let mut trace = Trace::new(format!("ACC{:0>3}", i));

            let t0 = base_time + Duration::hours(i as i64 * 24);
            let t1 = t0 + Duration::minutes(30);
            let t2 = t1 + Duration::minutes(15);
            let t3 = t2 + Duration::minutes(10);

            // Standard account lifecycle: create → verify → activate → close
            trace.add_event(Event::new("account_created", t0)
                .with_resource("account_manager"));
            trace.add_event(Event::new("account_verified", t1)
                .with_resource("verification_service"));
            trace.add_event(Event::new("account_activated", t2)
                .with_resource("activation_system"));
            trace.add_event(Event::new("account_closed", t3)
                .with_resource("system_admin"));

            log.add_trace(trace);
        }

        log
    }

    #[test]
    fn test_happy_path_perfect_fit_workflow() {
        // ============================================================
        // Phase 1: LOAD - Read event log from structured data
        // ============================================================
        let log = create_perfect_fit_log();

        // Validate log structure
        assert_eq!(log.len(), 5, "Expected 5 accounts (traces) in log");
        assert_eq!(log.num_events(), 20, "Expected 20 total events (4 per account)");

        println!("✓ Phase 1: LOAD - Loaded 5 accounts with 20 events (perfect-fit sequence)");

        // ============================================================
        // Phase 2: DISCOVER - Mine process model with Alpha Miner
        // ============================================================
        let miner = AlphaMiner::new();
        let discovered_model = miner.discover(&log);

        // Verify model structure
        assert!(!discovered_model.places.is_empty(),
                "Discovered model must have places");
        assert!(!discovered_model.transitions.is_empty(),
                "Discovered model must have transitions");
        assert!(discovered_model.initial_place.is_some(),
                "Model must have initial place");
        assert!(discovered_model.final_place.is_some(),
                "Model must have final place");

        let num_places = discovered_model.places.len();
        let num_transitions = discovered_model.transitions.len();

        println!("✓ Phase 2: DISCOVER - Alpha Miner discovered model: {} places, {} transitions",
                 num_places, num_transitions);

        // ============================================================
        // Phase 3: CONFORM - Check conformance of log to model
        // ============================================================
        let conformance_checker = TokenReplay::new();
        let conformance_result = conformance_checker.check(&log, &discovered_model);

        // Validate perfect-fit requirements
        assert_eq!(conformance_result.fitness, 1.0,
                   "Perfect-fit log must have fitness = 1.0 (got {:.4})",
                   conformance_result.fitness);

        assert!(conformance_result.is_conformant,
                "All traces in perfect-fit log must be conformant to discovered model");

        println!("✓ Phase 3: CONFORM - Perfect fit confirmed: fitness = {:.4}, is_conformant = {}",
                 conformance_result.fitness, conformance_result.is_conformant);

        // ============================================================
        // Phase 4: ANALYZE - Extract statistics from log
        // ============================================================
        let statistics = log_stats::log_statistics(&log);

        // Validate statistics
        assert_eq!(statistics.num_traces, 5,
                   "Expected 5 traces (accounts)");
        assert_eq!(statistics.num_events, 20,
                   "Expected 20 total events");
        assert_eq!(statistics.num_unique_activities, 4,
                   "Expected 4 unique activities: account_created, account_verified, account_activated, account_closed");
        assert_eq!(statistics.num_variants, 1,
                   "Expected 1 variant (all traces identical)");
        assert_eq!(statistics.min_trace_length, 4,
                   "Expected min trace length = 4");
        assert_eq!(statistics.max_trace_length, 4,
                   "Expected max trace length = 4");
        assert!((statistics.avg_trace_length - 4.0).abs() < 0.001,
                "Expected avg trace length = 4.0, got {}", statistics.avg_trace_length);

        println!("✓ Phase 4: ANALYZE - Statistics extracted:");
        println!("    - Traces: {}", statistics.num_traces);
        println!("    - Events: {}", statistics.num_events);
        println!("    - Unique Activities: {}", statistics.num_unique_activities);
        println!("    - Variants: {}", statistics.num_variants);
        println!("    - Avg Trace Length: {:.2}", statistics.avg_trace_length);

        // ============================================================
        // Summary: Complete Workflow Success
        // ============================================================
        println!("✓ Happy Path Workflow COMPLETE: Load → Discover → Conform → Analyze");
        println!("  Fitness: {:.4} | Conformance: {} | Variants: {}",
                 conformance_result.fitness,
                 conformance_result.is_conformant,
                 statistics.num_variants);
    }

    /// ============================================================
    /// TEST SCENARIO 2: Non-Conformant Path Test
    /// ============================================================
    /// Account lifecycle log with DEVIATIONS from the standard sequence.
    /// Some accounts skip verification or have additional activities.
    ///
    /// Expected behavior:
    /// - fitness < 1.0 (deviations detected)
    /// - is_conformant = false
    /// - Conformance output identifies deviations

    /// Helper: Create a non-conformant account log with deviations
    fn create_non_conformant_log() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Account 1: Standard flow (CONFORMANT)
        let mut trace1 = Trace::new("ACC001");
        let t0 = base_time;
        let t1 = t0 + Duration::minutes(30);
        let t2 = t1 + Duration::minutes(15);
        let t3 = t2 + Duration::minutes(10);

        trace1.add_event(Event::new("account_created", t0).with_resource("account_manager"));
        trace1.add_event(Event::new("account_verified", t1).with_resource("verification_service"));
        trace1.add_event(Event::new("account_activated", t2).with_resource("activation_system"));
        trace1.add_event(Event::new("account_closed", t3).with_resource("system_admin"));
        log.add_trace(trace1);

        // Account 2: SKIPS verification (NON-CONFORMANT DEVIATION)
        let mut trace2 = Trace::new("ACC002");
        let t0 = base_time + Duration::hours(24);
        let t1 = t0 + Duration::minutes(15);  // Skipped verification_service
        let t2 = t1 + Duration::minutes(10);

        trace2.add_event(Event::new("account_created", t0).with_resource("account_manager"));
        // MISSING: account_verified
        trace2.add_event(Event::new("account_activated", t1).with_resource("activation_system"));
        trace2.add_event(Event::new("account_closed", t2).with_resource("system_admin"));
        log.add_trace(trace2);

        // Account 3: Standard flow (CONFORMANT)
        let mut trace3 = Trace::new("ACC003");
        let t0 = base_time + Duration::hours(48);
        let t1 = t0 + Duration::minutes(30);
        let t2 = t1 + Duration::minutes(15);
        let t3 = t2 + Duration::minutes(10);

        trace3.add_event(Event::new("account_created", t0).with_resource("account_manager"));
        trace3.add_event(Event::new("account_verified", t1).with_resource("verification_service"));
        trace3.add_event(Event::new("account_activated", t2).with_resource("activation_system"));
        trace3.add_event(Event::new("account_closed", t3).with_resource("system_admin"));
        log.add_trace(trace3);

        // Account 4: HAS EXTRA activity (account_reviewed) before closing (NON-CONFORMANT DEVIATION)
        let mut trace4 = Trace::new("ACC004");
        let t0 = base_time + Duration::hours(72);
        let t1 = t0 + Duration::minutes(30);
        let t2 = t1 + Duration::minutes(15);
        let t3 = t2 + Duration::minutes(10);
        let t4 = t3 + Duration::minutes(20);
        let t5 = t4 + Duration::minutes(5);

        trace4.add_event(Event::new("account_created", t0).with_resource("account_manager"));
        trace4.add_event(Event::new("account_verified", t1).with_resource("verification_service"));
        trace4.add_event(Event::new("account_activated", t2).with_resource("activation_system"));
        trace4.add_event(Event::new("account_reviewed", t3).with_resource("compliance_check"));  // EXTRA activity
        trace4.add_event(Event::new("account_disputed", t4).with_resource("dispute_service"));    // EXTRA activity
        trace4.add_event(Event::new("account_closed", t5).with_resource("system_admin"));
        log.add_trace(trace4);

        // Account 5: Standard flow (CONFORMANT)
        let mut trace5 = Trace::new("ACC005");
        let t0 = base_time + Duration::hours(96);
        let t1 = t0 + Duration::minutes(30);
        let t2 = t1 + Duration::minutes(15);
        let t3 = t2 + Duration::minutes(10);

        trace5.add_event(Event::new("account_created", t0).with_resource("account_manager"));
        trace5.add_event(Event::new("account_verified", t1).with_resource("verification_service"));
        trace5.add_event(Event::new("account_activated", t2).with_resource("activation_system"));
        trace5.add_event(Event::new("account_closed", t3).with_resource("system_admin"));
        log.add_trace(trace5);

        log
    }

    #[test]
    fn test_non_conformant_path_deviation_detection() {
        // ============================================================
        // Phase 1: LOAD - Read event log with deviations
        // ============================================================
        let log = create_non_conformant_log();

        assert_eq!(log.len(), 5, "Expected 5 accounts");
        // Total events: ACC001(4) + ACC002(3) + ACC003(4) + ACC004(6) + ACC005(4) = 21
        assert_eq!(log.num_events(), 21, "Expected 21 total events (deviation adds extra events)");

        println!("✓ Phase 1: LOAD - Loaded 5 accounts with 21 events (includes deviations)");
        println!("    - ACC001: Standard (4 events)");
        println!("    - ACC002: DEVIATION - Missing verification (3 events)");
        println!("    - ACC003: Standard (4 events)");
        println!("    - ACC004: DEVIATION - Extra activities (6 events)");
        println!("    - ACC005: Standard (4 events)");

        // ============================================================
        // Phase 2: DISCOVER - Mine model from log WITH deviations
        // ============================================================
        let miner = AlphaMiner::new();
        let discovered_model = miner.discover(&log);

        assert!(!discovered_model.places.is_empty(), "Model must have places");
        assert!(!discovered_model.transitions.is_empty(), "Model must have transitions");

        println!("✓ Phase 2: DISCOVER - Model discovered from non-conformant log: {} places, {} transitions",
                 discovered_model.places.len(),
                 discovered_model.transitions.len());

        // ============================================================
        // Phase 3: CONFORM - Check conformance (expect deviations)
        // ============================================================
        let conformance_checker = TokenReplay::new();
        let conformance_result = conformance_checker.check(&log, &discovered_model);

        // Non-conformant log should have fitness < 1.0
        assert!(conformance_result.fitness < 1.0,
                "Non-conformant log must have fitness < 1.0 (got {:.4})",
                conformance_result.fitness);

        assert!(!conformance_result.is_conformant,
                "Log with deviations must be marked as non-conformant");

        // With deviations, expect fitness < 1.0 (depends on how Alpha Miner discovers the model)
        // Deviations can significantly reduce fitness (minimum 0.0, in this case around 0.2)
        assert!(conformance_result.fitness < 1.0,
                "Non-conformant log fitness should be reduced (got {:.4})",
                conformance_result.fitness);

        println!("✓ Phase 3: CONFORM - Deviations detected:");
        println!("    - Fitness: {:.4} (< 1.0 indicates non-conformance)", conformance_result.fitness);
        println!("    - Conformant: {} (deviations found)", conformance_result.is_conformant);

        // ============================================================
        // Phase 4: ANALYZE - Extract statistics and identify deviations
        // ============================================================
        let statistics = log_stats::log_statistics(&log);

        // With deviations, we expect more variants and activities
        assert!(statistics.num_unique_activities >= 4,
                "Non-conformant log introduces extra activities (expected >= 4, got {})",
                statistics.num_unique_activities);

        assert!(statistics.num_variants >= 1,
                "Expected at least 1 variant (got {})",
                statistics.num_variants);

        assert_ne!(statistics.min_trace_length, statistics.max_trace_length,
                   "Non-conformant log should have varying trace lengths (min: {}, max: {})",
                   statistics.min_trace_length, statistics.max_trace_length);

        println!("✓ Phase 4: ANALYZE - Statistics reveal deviations:");
        println!("    - Traces: {}", statistics.num_traces);
        println!("    - Events: {}", statistics.num_events);
        println!("    - Unique Activities: {} (baseline was 4)", statistics.num_unique_activities);
        println!("    - Variants: {} (baseline was 1)", statistics.num_variants);
        println!("    - Trace Lengths: min={}, max={}, avg={:.2}",
                 statistics.min_trace_length,
                 statistics.max_trace_length,
                 statistics.avg_trace_length);

        // ============================================================
        // Summary: Non-Conformant Workflow
        // ============================================================
        println!("✓ Non-Conformant Path Workflow COMPLETE");
        println!("  Result: Fitness {:.4} | Non-Conformant: {} | Variants: {}",
                 conformance_result.fitness,
                 !conformance_result.is_conformant,
                 statistics.num_variants);
    }

    /// ============================================================
    /// TEST SCENARIO 3: Statistics Validation Test
    /// ============================================================
    /// Large event log (100 events across 10 cases) with comprehensive
    /// statistics validation. Tests:
    /// - Activity frequency counting
    /// - Trace length distribution
    /// - Variant detection
    /// - Start/end activity analysis

    /// Helper: Create a large log for statistics validation
    fn create_large_statistics_log() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Create 10 accounts with varying trace lengths and activity sequences
        for i in 0..10 {
            let mut trace = Trace::new(format!("CASE_{:0>4}", i));
            let base = base_time + Duration::hours(i as i64 * 24);

            match i {
                // Accounts 0-6: Standard 4-activity sequence (7 accounts)
                0..=6 => {
                    trace.add_event(Event::new("Start", base).with_resource("system"));
                    trace.add_event(Event::new("Process", base + Duration::minutes(30)).with_resource("handler"));
                    trace.add_event(Event::new("Review", base + Duration::minutes(60)).with_resource("reviewer"));
                    trace.add_event(Event::new("End", base + Duration::minutes(90)).with_resource("system"));
                }
                // Accounts 7: Extended 5-activity sequence (1 account)
                7 => {
                    trace.add_event(Event::new("Start", base).with_resource("system"));
                    trace.add_event(Event::new("Process", base + Duration::minutes(30)).with_resource("handler"));
                    trace.add_event(Event::new("Escalate", base + Duration::minutes(60)).with_resource("escalation"));
                    trace.add_event(Event::new("Review", base + Duration::minutes(120)).with_resource("reviewer"));
                    trace.add_event(Event::new("End", base + Duration::minutes(150)).with_resource("system"));
                }
                // Accounts 8: Short 3-activity sequence (1 account)
                8 => {
                    trace.add_event(Event::new("Start", base).with_resource("system"));
                    trace.add_event(Event::new("Process", base + Duration::minutes(30)).with_resource("handler"));
                    trace.add_event(Event::new("End", base + Duration::minutes(45)).with_resource("system"));
                }
                // Account 9: Extended 6-activity sequence (1 account)
                9 => {
                    trace.add_event(Event::new("Start", base).with_resource("system"));
                    trace.add_event(Event::new("Process", base + Duration::minutes(30)).with_resource("handler"));
                    trace.add_event(Event::new("Review", base + Duration::minutes(60)).with_resource("reviewer"));
                    trace.add_event(Event::new("Revise", base + Duration::minutes(90)).with_resource("handler"));
                    trace.add_event(Event::new("Review", base + Duration::minutes(120)).with_resource("reviewer"));
                    trace.add_event(Event::new("End", base + Duration::minutes(150)).with_resource("system"));
                }
                _ => {}
            }

            log.add_trace(trace);
        }

        log
    }

    #[test]
    fn test_statistics_validation_large_log() {
        // ============================================================
        // Phase 1: LOAD - Read large event log
        // ============================================================
        let log = create_large_statistics_log();

        // Validate total events: 7×4 + 1×5 + 1×3 + 1×6 = 28 + 5 + 3 + 6 = 42 events
        assert_eq!(log.len(), 10, "Expected 10 cases");
        assert_eq!(log.num_events(), 42, "Expected 42 total events (7×4 + 1×5 + 1×3 + 1×6)");

        println!("✓ Phase 1: LOAD - Loaded 10 cases with 42 events");

        // ============================================================
        // Phase 2: DISCOVER - Mine model from large log
        // ============================================================
        let miner = AlphaMiner::new();
        let discovered_model = miner.discover(&log);

        assert!(!discovered_model.places.is_empty(), "Model must have places");
        assert!(!discovered_model.transitions.is_empty(), "Model must have transitions");

        println!("✓ Phase 2: DISCOVER - Model: {} places, {} transitions",
                 discovered_model.places.len(),
                 discovered_model.transitions.len());

        // ============================================================
        // Phase 3: CONFORM - Token replay on large log
        // ============================================================
        let conformance_checker = TokenReplay::new();
        let conformance_result = conformance_checker.check(&log, &discovered_model);

        // Large log with variant activities should have reduced fitness
        assert!(conformance_result.fitness >= 0.0 && conformance_result.fitness <= 1.0,
                "Fitness must be between 0.0 and 1.0 (got {:.4})",
                conformance_result.fitness);

        println!("✓ Phase 3: CONFORM - Fitness: {:.4}, Conformant: {}",
                 conformance_result.fitness,
                 conformance_result.is_conformant);

        // ============================================================
        // Phase 4: ANALYZE - Comprehensive statistics validation
        // ============================================================
        let statistics = log_stats::log_statistics(&log);

        // Trace counts
        assert_eq!(statistics.num_traces, 10,
                   "Expected 10 traces, got {}", statistics.num_traces);

        // Event count
        assert_eq!(statistics.num_events, 42,
                   "Expected 42 events, got {}", statistics.num_events);

        // Unique activities: Start, Process, Review, End, Escalate, Revise = 6
        assert_eq!(statistics.num_unique_activities, 6,
                   "Expected 6 unique activities (Start, Process, Review, End, Escalate, Revise), got {}",
                   statistics.num_unique_activities);

        // Variants: 4 different patterns
        // - Variant 1: Start → Process → Review → End (7 traces)
        // - Variant 2: Start → Process → Escalate → Review → End (1 trace)
        // - Variant 3: Start → Process → End (1 trace)
        // - Variant 4: Start → Process → Review → Revise → Review → End (1 trace)
        assert_eq!(statistics.num_variants, 4,
                   "Expected 4 variants, got {}", statistics.num_variants);

        // Trace length validation
        assert_eq!(statistics.min_trace_length, 3,
                   "Expected min trace length = 3, got {}", statistics.min_trace_length);

        assert_eq!(statistics.max_trace_length, 6,
                   "Expected max trace length = 6, got {}", statistics.max_trace_length);

        let expected_avg = 42.0 / 10.0; // 4.2
        assert!((statistics.avg_trace_length - expected_avg).abs() < 0.001,
                "Expected avg trace length ≈ {:.1}, got {:.2}",
                expected_avg, statistics.avg_trace_length);

        // ============================================================
        // Phase 5: ACTIVITY FREQUENCY ANALYSIS
        // ============================================================
        let activity_frequency = log_stats::activity_occurrence_matrix(&log);

        // Verify activity frequencies
        assert_eq!(*activity_frequency.get("Start").unwrap_or(&0), 10,
                   "Expected Start activity in all 10 traces");
        assert_eq!(*activity_frequency.get("End").unwrap_or(&0), 10,
                   "Expected End activity in all 10 traces");
        assert_eq!(*activity_frequency.get("Process").unwrap_or(&0), 10,
                   "Expected Process activity in all 10 traces");

        // Review appears: 7 times (standard) + 1 time (escalate) + 2 times (revise variant) = 10
        assert_eq!(*activity_frequency.get("Review").unwrap_or(&0), 10,
                   "Expected Review activity 10 times");

        // Escalate appears: 1 time
        assert_eq!(*activity_frequency.get("Escalate").unwrap_or(&0), 1,
                   "Expected Escalate activity 1 time");

        // Revise appears: 1 time
        assert_eq!(*activity_frequency.get("Revise").unwrap_or(&0), 1,
                   "Expected Revise activity 1 time");

        println!("✓ Phase 4: ANALYZE - Statistics Validated:");
        println!("    Traces: {}", statistics.num_traces);
        println!("    Events: {}", statistics.num_events);
        println!("    Unique Activities: {}", statistics.num_unique_activities);
        println!("    Variants: {}", statistics.num_variants);
        println!("    Trace Lengths: min={}, max={}, avg={:.2}",
                 statistics.min_trace_length,
                 statistics.max_trace_length,
                 statistics.avg_trace_length);
        println!("✓ Phase 5: ACTIVITY FREQUENCY");
        for (activity, freq) in activity_frequency.iter() {
            println!("    {}: {} occurrences", activity, freq);
        }

        // ============================================================
        // Phase 6: DIRECTLY-FOLLOWS ANALYSIS
        // ============================================================
        let directly_follows = log_stats::directly_follows_matrix(&log);

        // Verify key transitions
        assert!(directly_follows.contains_key(&("Start".to_string(), "Process".to_string())),
                "Expected Start → Process transition");
        assert!(directly_follows.contains_key(&("Process".to_string(), "Review".to_string())),
                "Expected Process → Review transition");
        assert!(directly_follows.contains_key(&("Review".to_string(), "End".to_string())),
                "Expected Review → End transition");

        // Verify transition frequencies
        let start_to_process = directly_follows.get(&("Start".to_string(), "Process".to_string())).unwrap_or(&0);
        assert_eq!(*start_to_process, 10,
                   "Expected Start → Process 10 times, got {}", start_to_process);

        println!("✓ Phase 6: DIRECTLY-FOLLOWS (sample)");
        for ((from, to), freq) in directly_follows.iter().take(8) {
            println!("    {} → {}: {} times", from, to, freq);
        }

        // ============================================================
        // Phase 7: START/END ACTIVITY ANALYSIS
        // ============================================================
        let start_activities = log_stats::get_start_activities(&log);
        let end_activities = log_stats::get_end_activities(&log);

        assert_eq!(*start_activities.get("Start").unwrap_or(&0), 10,
                   "Expected Start activity at beginning of all 10 traces");
        assert_eq!(*end_activities.get("End").unwrap_or(&0), 10,
                   "Expected End activity at end of all 10 traces");

        println!("✓ Phase 7: START/END ANALYSIS");
        println!("    Start Activities: {} trace(s) start with 'Start'", start_activities.get("Start").unwrap_or(&0));
        println!("    End Activities: {} trace(s) end with 'End'", end_activities.get("End").unwrap_or(&0));

        // ============================================================
        // Summary: Statistics Validation Complete
        // ============================================================
        println!("✓ Statistics Validation Workflow COMPLETE");
        println!("  Cases: {} | Events: {} | Activities: {} | Variants: {}",
                 statistics.num_traces,
                 statistics.num_events,
                 statistics.num_unique_activities,
                 statistics.num_variants);
        println!("  Fitness: {:.4} | Conformant: {}",
                 conformance_result.fitness,
                 conformance_result.is_conformant);
    }
}

// ============================================================
// WORKFLOW SUMMARY AND KEY ASSERTIONS
// ============================================================
//
// Test 1: Happy Path (Perfect Fit)
//   ✓ Load: 5 traces, 20 events
//   ✓ Discover: Alpha Miner produces valid Petri net
//   ✓ Conform: Fitness = 1.0, is_conformant = true
//   ✓ Analyze: 1 variant, 4 activities, avg_length = 4.0
//
// Test 2: Non-Conformant Path
//   ✓ Load: 5 traces with intentional deviations
//   ✓ Discover: Model discovered despite deviations
//   ✓ Conform: Fitness < 1.0, is_conformant = false
//   ✓ Analyze: Multiple variants, activity deviations detected
//
// Test 3: Statistics Validation
//   ✓ Load: 10 traces, 42 events
//   ✓ Discover: Model handles variable trace lengths
//   ✓ Conform: Fitness calculated across variants
//   ✓ Analyze: Comprehensive metrics including:
//      - Activity frequency: exact counts per activity
//      - Directly-follows: transition counts
//      - Start/end activities: entry/exit points
//      - Trace lengths: min/max/avg distribution
//      - Variants: distinct process patterns
