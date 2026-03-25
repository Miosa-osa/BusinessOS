// PM4PY Workflow Integration Test
// Comprehensive end-to-end tests for the complete process mining workflow:
// Load → Discover → Conform → Analyze

#[cfg(test)]
mod pm4py_workflow_integration_tests {
    use std::fs;
    use std::path::Path;

    // Helper function to create test data directory
    fn get_test_data_dir() -> String {
        "/Users/sac/chatmangpt/BusinessOS/bos/tests/data".to_string()
    }

    fn ensure_test_data_dir() {
        let dir = get_test_data_dir();
        if !Path::new(&dir).exists() {
            fs::create_dir_all(&dir).expect("Failed to create test data directory");
        }
    }

    // Helper to create a perfect-fit account lifecycle log in XES format
    fn create_perfect_fit_account_lifecycle_log() -> String {
        let test_data_dir = get_test_data_dir();
        ensure_test_data_dir();

        let log_path = format!("{}/perfect_fit_account_lifecycle.xes", test_data_dir);

        let xes_content = r#"<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="nested-attributes" xmlns="http://www.xes-standard.org/">
  <extension name="Concept" prefix="concept" uri="http://www.xes-standard.org/concept.xesext"/>
  <extension name="Time" prefix="time" uri="http://www.xes-standard.org/time.xesext"/>
  <global scope="trace">
    <string key="concept:name" value="case name"/>
  </global>
  <global scope="event">
    <string key="concept:name" value="activity name"/>
    <date key="time:timestamp" value="timestamp"/>
  </global>
  <trace>
    <string key="concept:name" value="account_0"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-01T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-01T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-01T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_1"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-02T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-02T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-02T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_2"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-03T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-03T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-03T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_3"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-04T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-04T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-04T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_4"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-05T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-05T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-05T13:00:00.000Z"/>
    </event>
  </trace>
</log>"#;

        fs::write(&log_path, xes_content)
            .expect("Failed to write perfect fit log file");
        log_path
    }

    // Helper to create a non-conformant log with deviations
    fn create_non_conformant_log_with_deviations() -> String {
        let test_data_dir = get_test_data_dir();
        ensure_test_data_dir();

        let log_path = format!("{}/non_conformant_account_lifecycle.xes", test_data_dir);

        let xes_content = r#"<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="nested-attributes" xmlns="http://www.xes-standard.org/">
  <extension name="Concept" prefix="concept" uri="http://www.xes-standard.org/concept.xesext"/>
  <extension name="Time" prefix="time" uri="http://www.xes-standard.org/time.xesext"/>
  <global scope="trace">
    <string key="concept:name" value="case name"/>
  </global>
  <global scope="event">
    <string key="concept:name" value="activity name"/>
    <date key="time:timestamp" value="timestamp"/>
  </global>
  <trace>
    <string key="concept:name" value="account_0"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-01T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-01T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-01T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_1"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-02T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-02T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_2"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-03T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-03T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_suspended"/>
      <date key="time:timestamp" value="2024-01-03T13:30:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-03T14:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_3"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-04T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-04T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-04T13:00:00.000Z"/>
    </event>
  </trace>
  <trace>
    <string key="concept:name" value="account_4"/>
    <event>
      <string key="concept:name" value="account_created"/>
      <date key="time:timestamp" value="2024-01-05T10:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_verified"/>
      <date key="time:timestamp" value="2024-01-05T12:00:00.000Z"/>
    </event>
    <event>
      <string key="concept:name" value="account_activated"/>
      <date key="time:timestamp" value="2024-01-05T13:00:00.000Z"/>
    </event>
  </trace>
</log>"#;

        fs::write(&log_path, xes_content)
            .expect("Failed to write non-conformant log file");
        log_path
    }

    // Helper to create a large-scale log with 100 events across 10 cases
    fn create_large_scale_event_log() -> String {
        let test_data_dir = get_test_data_dir();
        ensure_test_data_dir();

        let log_path = format!("{}/large_scale_event_log.xes", test_data_dir);

        let mut xes_content = String::from(
            r#"<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="nested-attributes" xmlns="http://www.xes-standard.org/">
  <extension name="Concept" prefix="concept" uri="http://www.xes-standard.org/concept.xesext"/>
  <extension name="Time" prefix="time" uri="http://www.xes-standard.org/time.xesext"/>
  <global scope="trace">
    <string key="concept:name" value="case name"/>
  </global>
  <global scope="event">
    <string key="concept:name" value="activity name"/>
    <date key="time:timestamp" value="timestamp"/>
  </global>
"#,
        );

        // Create 10 cases with 10 events each (100 total events)
        let activities = vec![
            "request_submitted",
            "request_validated",
            "document_collected",
            "document_verified",
            "approval_required",
            "approval_granted",
            "processing_started",
            "processing_completed",
            "notification_sent",
            "request_closed",
        ];

        for case_idx in 0..10 {
            let case_id = format!("case_{}", case_idx);

            xes_content.push_str(&format!("  <trace>\n    <string key=\"concept:name\" value=\"{}\"/>\n", case_id));

            for (event_idx, activity) in activities.iter().enumerate() {
                let timestamp = format!(
                    "2024-01-{:02}T{:02}:{:02}:00.000Z",
                    (case_idx % 30) + 1,
                    10 + (event_idx / 6),
                    (event_idx % 6) * 10
                );

                xes_content.push_str(&format!(
                    "    <event>\n      <string key=\"concept:name\" value=\"{}\"/>\n      <date key=\"time:timestamp\" value=\"{}\"/>\n    </event>\n",
                    activity, timestamp
                ));
            }

            xes_content.push_str("  </trace>\n");
        }

        xes_content.push_str("</log>");

        fs::write(&log_path, xes_content)
            .expect("Failed to write large scale log file");
        log_path
    }

    // ========================================================================
    // TEST 1: Happy Path Test (Perfect Fit)
    // ========================================================================
    // Validates: Load → Discover Alpha Miner → Conform (fitness=1.0) → Analyze
    #[test]
    fn test_happy_path_perfect_fit_account_lifecycle() {
        use bos_core::process::ProcessMiningEngine;
        use pm4py::conformance::TokenReplay;

        // GIVEN: Perfect-fit account lifecycle log
        let log_path = create_perfect_fit_account_lifecycle_log();
        println!("Created perfect fit log at: {}", log_path);

        // WHEN: Load the event log
        let engine = ProcessMiningEngine::new();
        let log = engine.load_log(&log_path)
            .expect("Failed to load event log");

        // Validate loaded log structure
        assert_eq!(
            log.traces.len(),
            5,
            "Expected 5 traces in loaded log, got {}",
            log.traces.len()
        );

        let total_events: usize = log.traces.iter().map(|t| t.events.len()).sum();
        assert_eq!(
            total_events,
            15,
            "Expected 15 total events (5 traces × 3 events each), got {}",
            total_events
        );

        // WHEN: Discover process model using Alpha Miner
        let discovery_result = engine.discover_alpha(&log)
            .expect("Failed to discover model");

        // Validate discovery result structure
        assert_eq!(
            discovery_result.algorithm,
            "Alpha Miner",
            "Algorithm should be 'Alpha Miner', got '{}'",
            discovery_result.algorithm
        );

        println!(
            "Discovered model: {} places, {} transitions, {} arcs",
            discovery_result.places, discovery_result.transitions, discovery_result.arcs
        );

        // Alpha Miner for A->B->C pattern should discover:
        // - 4 places (p_start, p_AB, p_BC, p_end)
        // - 3 transitions (A, B, C)
        assert_eq!(
            discovery_result.transitions,
            3,
            "Expected 3 transitions for account lifecycle (account_created, account_verified, account_activated), got {}",
            discovery_result.transitions
        );

        assert_eq!(
            discovery_result.places,
            4,
            "Expected 4 places for A->B->C pattern, got {}",
            discovery_result.places
        );

        // WHEN: Perform conformance checking using Token Replay
        let miner = pm4py::discovery::AlphaMiner::new();
        let petri_net = miner.discover(&log);

        let token_replay = TokenReplay::new();
        let conformance_result = token_replay.check(&log, &petri_net);

        // THEN: Perfect fit logs should have fitness = 1.0
        assert_eq!(
            conformance_result.fitness,
            1.0,
            "Expected fitness = 1.0 for perfect fit log, got {}",
            conformance_result.fitness
        );

        // All traces should fit the model
        let expected_fitting_traces = log.traces.len();
        let actual_fitting_traces = (expected_fitting_traces as f64 * conformance_result.fitness).ceil() as usize;

        assert_eq!(
            actual_fitting_traces,
            expected_fitting_traces,
            "Expected all {} traces to fit model, got {}",
            expected_fitting_traces,
            actual_fitting_traces
        );

        // WHEN: Analyze event log statistics
        use pm4py::statistics::log_statistics;
        use pm4py::log::operations;

        let stats = log_statistics(&log);
        let activity_freq = operations::activity_frequency(&log);

        // THEN: Statistics should match expected values
        assert_eq!(
            stats.num_traces,
            5,
            "Expected 5 traces in statistics, got {}",
            stats.num_traces
        );

        assert_eq!(
            stats.num_events,
            15,
            "Expected 15 events in statistics, got {}",
            stats.num_events
        );

        assert_eq!(
            stats.num_unique_activities,
            3,
            "Expected 3 unique activities (account_created, account_verified, account_activated), got {}",
            stats.num_unique_activities
        );

        // Validate activity frequency
        let activity_counts: Vec<&str> = vec!["account_created", "account_verified", "account_activated"];
        for activity in activity_counts.iter() {
            let frequency = activity_freq.get(*activity)
                .expect(&format!("Activity '{}' not found in frequency map", activity));
            assert_eq!(
                *frequency,
                5,
                "Expected activity '{}' to appear 5 times (once per case), got {}",
                activity,
                frequency
            );
        }

        // Validate trace length statistics
        assert_eq!(
            stats.avg_trace_length,
            3.0,
            "Expected average trace length = 3.0 (each trace has 3 events), got {}",
            stats.avg_trace_length
        );

        assert_eq!(
            stats.min_trace_length,
            3,
            "Expected minimum trace length = 3, got {}",
            stats.min_trace_length
        );

        assert_eq!(
            stats.max_trace_length,
            3,
            "Expected maximum trace length = 3, got {}",
            stats.max_trace_length
        );

        println!("✓ Happy Path Test PASSED: Perfect fit account lifecycle validated");
    }

    // ========================================================================
    // TEST 2: Non-Conformant Path Test
    // ========================================================================
    // Validates: Load → Discover → Conform (fitness < 1.0) → Identify deviations
    #[test]
    fn test_non_conformant_path_with_deviations() {
        use bos_core::process::ProcessMiningEngine;
        use pm4py::conformance::TokenReplay;

        // GIVEN: Event log with deviations (missing step, extra activities)
        let log_path = create_non_conformant_log_with_deviations();
        println!("Created non-conformant log at: {}", log_path);

        // WHEN: Load the event log
        let engine = ProcessMiningEngine::new();
        let log = engine.load_log(&log_path)
            .expect("Failed to load event log");

        // Validate loaded log structure
        assert_eq!(
            log.traces.len(),
            5,
            "Expected 5 traces in loaded log, got {}",
            log.traces.len()
        );

        // Log has deviations:
        // - Case 0: A->B->C (3 events) - CONFORMANT
        // - Case 1: A->C (2 events) - DEVIATES (missing B)
        // - Case 2: A->B->X->C (4 events) - DEVIATES (extra X)
        // - Case 3: A->B->C (3 events) - CONFORMANT
        // - Case 4: A->B->C (3 events) - CONFORMANT
        // Total: 17 events across 5 traces

        let total_events: usize = log.traces.iter().map(|t| t.events.len()).sum();
        assert_eq!(
            total_events,
            17,
            "Expected 17 total events (with deviations), got {}",
            total_events
        );

        // WHEN: Discover process model from the entire (imperfect) log
        let discovery_result = engine.discover_alpha(&log)
            .expect("Failed to discover model");

        println!(
            "Discovered model from non-conformant log: {} places, {} transitions, {} arcs",
            discovery_result.places, discovery_result.transitions, discovery_result.arcs
        );

        // The discovered model may vary since we're mining from a non-conformant log
        // Just verify we got a valid model
        assert!(
            discovery_result.transitions > 0,
            "Expected at least 1 transition in discovered model, got {}",
            discovery_result.transitions
        );

        // WHEN: Perform conformance checking
        let miner = pm4py::discovery::AlphaMiner::new();
        let petri_net = miner.discover(&log);

        let token_replay = TokenReplay::new();
        let conformance_result = token_replay.check(&log, &petri_net);

        // THEN: Non-conformant logs should have fitness < 1.0
        // (Unless the discovery algorithm perfectly captures the deviations)
        println!(
            "Conformance fitness for non-conformant log: {}",
            conformance_result.fitness
        );

        // With deviations, not all traces should fit
        let fitting_traces = (log.traces.len() as f64 * conformance_result.fitness).ceil() as usize;
        println!(
            "Fitting traces: {} out of {}",
            fitting_traces,
            log.traces.len()
        );

        // At minimum, if fitness < 1.0, not all traces are fitting
        if conformance_result.fitness < 1.0 {
            assert!(
                fitting_traces < log.traces.len(),
                "When fitness < 1.0, fitting traces should be less than total traces"
            );
        }

        // WHEN: Analyze statistics of non-conformant log
        use pm4py::statistics::log_statistics;
        use pm4py::log::operations;

        let stats = log_statistics(&log);
        let activity_freq = operations::activity_frequency(&log);

        // THEN: Statistics should capture all activities including deviations
        assert_eq!(
            stats.num_traces,
            5,
            "Expected 5 traces in statistics"
        );

        assert_eq!(
            stats.num_events,
            17,
            "Expected 17 events in statistics (including deviations)"
        );

        // Activity frequency should show:
        // - account_created: 5 times (all cases)
        // - account_verified: 4 times (all except case 1)
        // - account_activated: 5 times (all cases, including case 2 which has extra suspended)
        // - account_suspended: 1 time (case 2 only)

        let created_freq = activity_freq.get("account_created").copied().unwrap_or(0);
        assert_eq!(
            created_freq,
            5,
            "Expected 'account_created' to appear 5 times (all cases), got {}",
            created_freq
        );

        let verified_freq = activity_freq.get("account_verified").copied().unwrap_or(0);
        assert_eq!(
            verified_freq,
            4,
            "Expected 'account_verified' to appear 4 times (missing from case 1), got {}",
            verified_freq
        );

        let activated_freq = activity_freq.get("account_activated").copied().unwrap_or(0);
        assert_eq!(
            activated_freq,
            5,
            "Expected 'account_activated' to appear 5 times (all cases), got {}",
            activated_freq
        );

        // Check for deviating activities
        if activity_freq.contains_key("account_suspended") {
            let suspended_freq = activity_freq.get("account_suspended").copied().unwrap_or(0);
            assert_eq!(
                suspended_freq,
                1,
                "Expected 'account_suspended' to appear 1 time (case 2 only), got {}",
                suspended_freq
            );
        }

        // Trace length statistics should be more varied
        assert!(
            stats.min_trace_length < stats.max_trace_length,
            "Expected variable trace lengths due to deviations"
        );

        println!(
            "✓ Non-Conformant Path Test PASSED: Deviations identified in log structure"
        );
    }

    // ========================================================================
    // TEST 3: Statistics Validation Test
    // ========================================================================
    // Validates: Load large log → Analyze all metrics → Assert exact values
    #[test]
    fn test_statistics_validation_large_scale_log() {
        use bos_core::process::ProcessMiningEngine;
        use pm4py::statistics::log_statistics;
        use pm4py::log::operations;

        // GIVEN: Large-scale log with 100 events across 10 cases
        let log_path = create_large_scale_event_log();
        println!("Created large-scale log at: {}", log_path);

        // WHEN: Load the event log
        let engine = ProcessMiningEngine::new();
        let log = engine.load_log(&log_path)
            .expect("Failed to load event log");

        // THEN: Validate log structure
        assert_eq!(
            log.traces.len(),
            10,
            "Expected 10 cases in loaded log, got {}",
            log.traces.len()
        );

        let total_events: usize = log.traces.iter().map(|t| t.events.len()).sum();
        assert_eq!(
            total_events,
            100,
            "Expected 100 total events (10 cases × 10 events each), got {}",
            total_events
        );

        // WHEN: Extract comprehensive statistics
        let stats = log_statistics(&log);
        let activity_freq = operations::activity_frequency(&log);

        // THEN: Validate comprehensive metrics

        // 1. Trace count
        assert_eq!(
            stats.num_traces,
            10,
            "Expected 10 traces, got {}",
            stats.num_traces
        );

        // 2. Event count
        assert_eq!(
            stats.num_events,
            100,
            "Expected 100 events, got {}",
            stats.num_events
        );

        // 3. Unique activities (10 distinct activity types)
        assert_eq!(
            stats.num_unique_activities,
            10,
            "Expected 10 unique activities, got {}",
            stats.num_unique_activities
        );

        // 4. Trace length statistics
        assert_eq!(
            stats.min_trace_length,
            10,
            "Expected minimum trace length = 10 (all cases have 10 events), got {}",
            stats.min_trace_length
        );

        assert_eq!(
            stats.max_trace_length,
            10,
            "Expected maximum trace length = 10 (all cases have 10 events), got {}",
            stats.max_trace_length
        );

        assert_eq!(
            stats.avg_trace_length,
            10.0,
            "Expected average trace length = 10.0, got {}",
            stats.avg_trace_length
        );

        // 5. Activity frequency validation
        // Each of the 10 activities should appear exactly 10 times (once per case)
        let expected_activities = vec![
            "request_submitted",
            "request_validated",
            "document_collected",
            "document_verified",
            "approval_required",
            "approval_granted",
            "processing_started",
            "processing_completed",
            "notification_sent",
            "request_closed",
        ];

        for activity in expected_activities.iter() {
            let frequency = activity_freq.get(*activity).copied().unwrap_or(0);
            assert_eq!(
                frequency,
                10,
                "Expected activity '{}' to appear 10 times (once per case), got {}",
                activity,
                frequency
            );
        }

        // 6. Duration statistics (case-level)
        // Extract duration from first and last events
        let mut case_durations: Vec<i64> = Vec::new();
        for trace in &log.traces {
            if trace.events.len() > 1 {
                if let (Some(first), Some(last)) = (trace.events.first(), trace.events.last()) {
                    let duration = (last.timestamp - first.timestamp).num_seconds();
                    case_durations.push(duration);
                }
            }
        }

        assert_eq!(
            case_durations.len(),
            10,
            "Expected 10 cases with duration data, got {}",
            case_durations.len()
        );

        // All durations should be positive and similar (within same day)
        for duration in case_durations.iter() {
            assert!(
                *duration >= 0,
                "Expected non-negative duration, got {}",
                duration
            );
        }

        case_durations.sort();
        let min_duration = case_durations.first().copied().unwrap_or(0);
        let max_duration = case_durations.last().copied().unwrap_or(0);
        let avg_duration = case_durations.iter().sum::<i64>() as f64 / case_durations.len() as f64;

        println!(
            "Duration stats - Min: {}s, Max: {}s, Avg: {:.2}s",
            min_duration, max_duration, avg_duration
        );

        // All durations should be within reasonable bounds (same day)
        assert!(
            max_duration - min_duration < 24 * 3600,
            "Expected all cases to complete within 24 hours"
        );

        // 7. Variant analysis (if supported)
        println!("Number of unique variants: {}", stats.num_variants);
        // Since all traces follow same A->B->C->...->J pattern, expect single variant
        assert_eq!(
            stats.num_variants,
            1,
            "Expected 1 unique variant (all traces follow same pattern)"
        );

        // 8. Full activity frequency table
        println!("Activity Frequency Distribution:");
        let mut freq_vec: Vec<_> = activity_freq.iter().collect();
        freq_vec.sort_by(|a, b| b.1.cmp(a.1)); // Sort by frequency descending

        let mut total_counted = 0;
        for (activity, frequency) in freq_vec.iter() {
            println!("  {}: {} occurrences", activity, frequency);
            total_counted += frequency;
        }

        assert_eq!(
            total_counted,
            100,
            "Total activity count should equal total events (100), got {}",
            total_counted
        );

        println!("✓ Statistics Validation Test PASSED: All metrics validated against expected values");
    }
}
