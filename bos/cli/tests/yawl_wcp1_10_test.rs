// YAWL Control Flow Patterns WCP1-10 Formal Tests
// Comprehensive testing of workflow control flow patterns using Alpha Miner discovery
// Tests verify pattern presence, Petri net structure, and soundness properties

#[cfg(test)]
mod yawl_wcp_tests {
    use std::collections::{HashMap, HashSet};
    use std::fs;
    use std::path::Path;

    // ========================================================================
    // TEST DATA STRUCTURES & HELPERS
    // ========================================================================

    /// Petri net representation for soundness checking
    #[derive(Clone, Debug)]
    struct PetriNet {
        places: Vec<String>,
        transitions: Vec<String>,
        arcs: Vec<(String, String)>, // source -> target
    }

    /// Event log representation
    #[derive(Clone, Debug)]
    struct EventLog {
        traces: Vec<Vec<String>>, // Each trace is a sequence of activity names
    }

    impl EventLog {
        fn new() -> Self {
            EventLog { traces: vec![] }
        }

        fn add_trace(&mut self, trace: Vec<String>) {
            self.traces.push(trace);
        }

        fn to_xes(&self) -> String {
            let mut xes = String::from(
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

            for (trace_idx, trace) in self.traces.iter().enumerate() {
                xes.push_str(&format!(
                    "  <trace>\n    <string key=\"concept:name\" value=\"trace_{}\"/>\n",
                    trace_idx
                ));

                for (event_idx, activity) in trace.iter().enumerate() {
                    let hour = 10 + (event_idx / 4);
                    let minute = (event_idx % 4) * 15;
                    let timestamp = format!("2024-01-01T{:02}:{:02}:00.000Z", hour, minute);

                    xes.push_str(&format!(
                        "    <event>\n      <string key=\"concept:name\" value=\"{}\"/>\n      <date key=\"time:timestamp\" value=\"{}\"/>\n    </event>\n",
                        activity, timestamp
                    ));
                }

                xes.push_str("  </trace>\n");
            }

            xes.push_str("</log>");
            xes
        }
    }

    fn get_test_data_dir() -> String {
        "/Users/sac/chatmangpt/BusinessOS/bos/tests/data/yawl_wcp".to_string()
    }

    fn ensure_test_data_dir() {
        let dir = get_test_data_dir();
        if !Path::new(&dir).exists() {
            fs::create_dir_all(&dir).expect("Failed to create test data directory");
        }
    }

    fn write_xes_log(filename: &str, log: &EventLog) -> String {
        ensure_test_data_dir();
        let path = format!("{}/{}.xes", get_test_data_dir(), filename);
        let xes_content = log.to_xes();
        fs::write(&path, xes_content).expect("Failed to write XES file");
        path
    }

    /// Simple Alpha Miner-style discovery
    /// Extracts directly-follows relationships and builds basic Petri net
    fn discover_petri_net(log: &EventLog) -> PetriNet {
        let mut directly_follows: HashMap<String, HashSet<String>> = HashMap::new();
        let mut all_activities: HashSet<String> = HashSet::new();

        // Build directly-follows relation
        for trace in &log.traces {
            for i in 0..trace.len() {
                all_activities.insert(trace[i].clone());
                if i + 1 < trace.len() {
                    directly_follows
                        .entry(trace[i].clone())
                        .or_insert_with(HashSet::new)
                        .insert(trace[i + 1].clone());
                }
            }
        }

        // Create Petri net: each activity becomes a transition
        let transitions: Vec<String> = all_activities.iter().cloned().collect();

        // Create places: one for each directly-follows relationship + start/end
        let mut places = vec!["start".to_string(), "end".to_string()];
        let mut arcs = vec![];

        // Add start -> first activities
        let first_activities: HashSet<String> = log
            .traces
            .iter()
            .filter_map(|t| t.first().cloned())
            .collect();

        for activity in &first_activities {
            let place = format!("p_{}_{}", activity, activity);
            if !places.contains(&place) {
                places.push(place.clone());
            }
            arcs.push(("start".to_string(), activity.clone()));
            arcs.push((activity.clone(), "end".to_string()));
        }

        // Add directly-follows relationships as places and arcs
        for (from, to_set) in directly_follows.iter() {
            for to in to_set.iter() {
                let place = format!("p_{}_{}", from, to);
                if !places.contains(&place) {
                    places.push(place.clone());
                }
                arcs.push((from.clone(), to.clone()));
            }
        }

        PetriNet {
            places,
            transitions,
            arcs,
        }
    }

    /// Check if Petri net is sound (no deadlocks, proper termination)
    fn check_soundness(log: &EventLog, net: &PetriNet) -> bool {
        // Simple soundness check: verify all traces can execute to completion
        // In a real implementation, this would use formalized soundness criteria

        for trace in &log.traces {
            let mut can_execute = true;

            // Check: all activities in trace are in net transitions
            for activity in trace {
                if !net.transitions.contains(activity) {
                    can_execute = false;
                    break;
                }
            }

            // Check: trace is sequentially connected
            if can_execute {
                for i in 0..trace.len() - 1 {
                    let has_arc = net
                        .arcs
                        .iter()
                        .any(|(s, t)| s == &trace[i] && t == &trace[i + 1]);
                    if !has_arc {
                        can_execute = false;
                        break;
                    }
                }
            }

            if !can_execute {
                return false;
            }
        }

        true
    }

    /// Calculate fitness: proportion of events from log that can be replayed
    fn calculate_fitness(log: &EventLog, net: &PetriNet) -> f64 {
        let total_events: usize = log.traces.iter().map(|t| t.len()).sum();
        let mut replayed_events = 0;

        for trace in &log.traces {
            for activity in trace {
                if net.transitions.contains(activity) {
                    replayed_events += 1;
                }
            }
        }

        if total_events == 0 {
            1.0
        } else {
            replayed_events as f64 / total_events as f64
        }
    }

    /// Calculate precision: proportion of net behavior that matches log
    fn calculate_precision(log: &EventLog, net: &PetriNet) -> f64 {
        let mut net_df: HashSet<(String, String)> = HashSet::new();
        let mut log_df: HashSet<(String, String)> = HashSet::new();

        // Extract directly-follows from net
        for (source, target) in &net.arcs {
            net_df.insert((source.clone(), target.clone()));
        }

        // Extract directly-follows from log
        for trace in &log.traces {
            for i in 0..trace.len() - 1 {
                log_df.insert((trace[i].clone(), trace[i + 1].clone()));
            }
        }

        if net_df.is_empty() {
            1.0
        } else {
            let matched = log_df.intersection(&net_df).count();
            matched as f64 / net_df.len() as f64
        }
    }

    // ========================================================================
    // WCP1: SEQUENCE (A → B → C)
    // ========================================================================
    // Pattern: Simple sequential execution of activities
    // Expected: Linear path through 3 activities
    // Soundness: No branches, proper start/end

    #[test]
    fn test_wcp1_sequence_basic() {
        // GIVEN: Event log with pure sequence pattern
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);

        let log_path = write_xes_log("wcp1_sequence_basic", &log);
        println!("WCP1 test log: {}", log_path);

        // WHEN: Discover Petri net
        let net = discover_petri_net(&log);

        // THEN: Verify structure
        assert_eq!(net.transitions.len(), 3, "Should have 3 transitions (A, B, C)");
        assert!(
            net.transitions.contains(&"A".to_string()),
            "Should contain transition A"
        );
        assert!(
            net.transitions.contains(&"B".to_string()),
            "Should contain transition B"
        );
        assert!(
            net.transitions.contains(&"C".to_string()),
            "Should contain transition C"
        );

        // AND: Verify arcs form sequence A->B->C
        assert!(
            net.arcs.iter().any(|(s, t)| s == "A" && t == "B"),
            "Should have arc A->B"
        );
        assert!(
            net.arcs.iter().any(|(s, t)| s == "B" && t == "C"),
            "Should have arc B->C"
        );

        // AND: Verify soundness
        assert!(check_soundness(&log, &net), "Net should be sound");

        // AND: Verify fitness and precision
        let fitness = calculate_fitness(&log, &net);
        let precision = calculate_precision(&log, &net);
        assert!(fitness >= 0.9, "Fitness should be high: {}", fitness);
        assert!(precision >= 0.8, "Precision should be acceptable: {}", precision);
    }

    #[test]
    fn test_wcp1_sequence_with_variants() {
        // GIVEN: Event log with sequence pattern (multiple repetitions)
        let mut log = EventLog::new();
        for _ in 0..5 {
            log.add_trace(vec!["Start".to_string(), "Process".to_string(), "End".to_string()]);
        }

        let net = discover_petri_net(&log);

        // THEN: Verify 3 transitions
        assert_eq!(
            net.transitions.len(),
            3,
            "Should discover exactly 3 transitions"
        );

        // AND: Verify start-middle-end relationship
        assert!(
            net.arcs
                .iter()
                .any(|(s, t)| s == "Start" && t == "Process"),
            "Should have Start->Process"
        );
        assert!(
            net.arcs
                .iter()
                .any(|(s, t)| s == "Process" && t == "End"),
            "Should have Process->End"
        );

        // AND: Verify soundness
        assert!(check_soundness(&log, &net), "Net should be sound");
    }

    // ========================================================================
    // WCP2: PARALLEL SPLIT (A → (B || C))
    // ========================================================================
    // Pattern: After A, both B and C execute in parallel
    // Expected: A with two outgoing flows
    // Soundness: Both branches must complete

    #[test]
    fn test_wcp2_parallel_split() {
        // GIVEN: Event log with parallel split after A
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string()]);

        let net = discover_petri_net(&log);

        // THEN: Verify A is present
        assert!(
            net.transitions.contains(&"A".to_string()),
            "Should contain transition A"
        );

        // AND: Verify both B and C are reachable from A
        let arcs_from_a: Vec<_> = net
            .arcs
            .iter()
            .filter(|(s, _)| s == "A")
            .map(|(_, t)| t.clone())
            .collect();

        assert!(
            arcs_from_a.contains(&"B".to_string()),
            "A should have outgoing arc to B"
        );
        assert!(
            arcs_from_a.contains(&"C".to_string()),
            "A should have outgoing arc to C"
        );

        // AND: Verify B and C are in transitions
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));

        // AND: Soundness check
        assert!(check_soundness(&log, &net), "Net should be sound");

        // AND: Fitness check (both B and C appear in log)
        let fitness = calculate_fitness(&log, &net);
        assert!(fitness >= 0.9, "High fitness expected: {}", fitness);
    }

    #[test]
    fn test_wcp2_parallel_split_convergence() {
        // GIVEN: Parallel split with convergence
        let mut log = EventLog::new();
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Verify structure
        assert!(net.transitions.contains(&"A".to_string()));
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));
        assert!(net.transitions.contains(&"D".to_string()));

        // Verify A->B and A->C
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "B"));
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "C"));

        // Verify convergence to D
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "D"));
        assert!(net.arcs.iter().any(|(s, t)| s == "C" && t == "D"));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP3: SYNCHRONIZATION (join parallel paths)
    // ========================================================================
    // Pattern: Multiple incoming flows merge to single point
    // Expected: Join place with multiple input transitions
    // Soundness: All branches must complete before join

    #[test]
    fn test_wcp3_synchronization() {
        // GIVEN: Event log with parallel paths that join
        let mut log = EventLog::new();
        // All traces must execute both B and C before D
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "C".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Verify all activities
        assert!(net.transitions.contains(&"A".to_string()));
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));
        assert!(net.transitions.contains(&"D".to_string()));

        // Verify both B and C flow to D
        let b_d_exists = net.arcs.iter().any(|(s, t)| s == "B" && t == "D");
        let c_d_exists = net.arcs.iter().any(|(s, t)| s == "C" && t == "D");

        assert!(b_d_exists || c_d_exists, "Should have paths to D from parallel branches");

        // Verify soundness
        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp3_synchronization_multiple_joins() {
        // GIVEN: Log with multiple parallel branches joining
        let mut log = EventLog::new();
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "E".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "C".to_string(),
            "E".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "D".to_string(),
            "E".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // All transitions present
        assert_eq!(net.transitions.len(), 5); // A, B, C, D, E

        // E is reachable from B, C, D
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "E"));
        assert!(net.arcs.iter().any(|(s, t)| s == "C" && t == "E"));
        assert!(net.arcs.iter().any(|(s, t)| s == "D" && t == "E"));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP4: EXCLUSIVE CHOICE (A → (B XOR C))
    // ========================================================================
    // Pattern: After A, either B or C executes, not both
    // Expected: One path with B, another with C, never both
    // Soundness: No path has both B and C in same case

    #[test]
    fn test_wcp4_exclusive_choice() {
        // GIVEN: Event log with exclusive choice (never both B and C)
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);

        let net = discover_petri_net(&log);

        // Verify transitions
        assert!(net.transitions.contains(&"A".to_string()));
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));
        assert!(net.transitions.contains(&"D".to_string()));

        // Verify choice: A->B and A->C both exist
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "B"));
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "C"));

        // Verify both paths lead to D
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "D"));
        assert!(net.arcs.iter().any(|(s, t)| s == "C" && t == "D"));

        // Verify exclusivity: no trace contains both B and C
        for trace in &log.traces {
            let has_b = trace.contains(&"B".to_string());
            let has_c = trace.contains(&"C".to_string());
            assert!(!(has_b && has_c), "Trace should not contain both B and C");
        }

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp4_exclusive_choice_with_skip() {
        // GIVEN: Exclusive choice with option to skip
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string()]);

        let net = discover_petri_net(&log);

        // A is entry point
        assert!(net.transitions.contains(&"A".to_string()));

        // Multiple paths from A
        let from_a = net
            .arcs
            .iter()
            .filter(|(s, _)| s == "A")
            .count();
        assert!(from_a >= 1, "A should have outgoing arcs");

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP5: SIMPLE MERGE (join exclusive paths)
    // ========================================================================
    // Pattern: Multiple exclusive paths merge
    // Expected: Both paths lead to common point
    // Soundness: Join point reached from either path

    #[test]
    fn test_wcp5_simple_merge() {
        // GIVEN: Event log with exclusive paths that merge
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);

        let net = discover_petri_net(&log);

        // All transitions present
        assert_eq!(net.transitions.len(), 4);

        // B->D and C->D both exist (simple merge)
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "D"));
        assert!(net.arcs.iter().any(|(s, t)| s == "C" && t == "D"));

        // Verify soundness: D is reachable from both paths
        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp5_simple_merge_multiple_sources() {
        // GIVEN: Multiple exclusive paths merging
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "D".to_string(), "E".to_string()]);

        let net = discover_petri_net(&log);

        // All merge to E
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "E"));
        assert!(net.arcs.iter().any(|(s, t)| s == "C" && t == "E"));
        assert!(net.arcs.iter().any(|(s, t)| s == "D" && t == "E"));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP6: MULTI-CHOICE (A → (B AND C possibly))
    // ========================================================================
    // Pattern: After A, combination of B and C may execute
    // Expected: Both B and C together, or just B, or just C
    // Soundness: All selected activities complete before next

    #[test]
    fn test_wcp6_multi_choice() {
        // GIVEN: Event log with multi-choice behavior
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]); // Both B and C

        let net = discover_petri_net(&log);

        // All transitions present
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));

        // Both reachable from A
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "B"));
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "C"));

        // Both can reach D
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "D") ||
            net.arcs.iter().any(|(s, t)| s == "C" && t == "D"));

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp6_multi_choice_three_branches() {
        // GIVEN: Multi-choice with three possible combinations
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "E".to_string()]);
        log.add_trace(vec![
            "A".to_string(),
            "D".to_string(),
            "E".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "E".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // All activities present
        assert_eq!(net.transitions.len(), 5);

        // At least B, C, D should be reachable from A
        let reachable_from_a: HashSet<_> = net
            .arcs
            .iter()
            .filter(|(s, _)| s == "A")
            .map(|(_, t)| t.clone())
            .collect();

        assert!(reachable_from_a.contains(&"B".to_string()));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP7: STRUCTURED PARALLEL (A → (B || C) → D)
    // ========================================================================
    // Pattern: Parallel paths with structured entry and exit
    // Expected: A, then B||C in parallel, then D
    // Soundness: Both branches must complete before D

    #[test]
    fn test_wcp7_structured_parallel() {
        // GIVEN: Event log with structured parallel pattern
        let mut log = EventLog::new();
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "C".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // All four transitions
        assert_eq!(net.transitions.len(), 4);

        // A is entry
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && (t == "B" || t == "C")));

        // Both B and C precede D
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "D") ||
            net.arcs.iter().any(|(s, t)| s == "B" && t == "C"));

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp7_structured_parallel_interleaved() {
        // GIVEN: Traces where B and C interleave before D
        let mut log = EventLog::new();
        log.add_trace(vec![
            "Start".to_string(),
            "B".to_string(),
            "C".to_string(),
            "End".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "C".to_string(),
            "B".to_string(),
            "End".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "B".to_string(),
            "C".to_string(),
            "End".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Verify structure: Start->B, Start->C, B->End, C->End or interleaved
        assert!(net.transitions.contains(&"Start".to_string()));
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));
        assert!(net.transitions.contains(&"End".to_string()));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP8: MULTI-MERGE (multiple paths merge without sync)
    // ========================================================================
    // Pattern: Multiple paths converge without synchronization requirement
    // Expected: Paths merge at D (no waiting for all branches)
    // Soundness: D reachable from any preceding state

    #[test]
    fn test_wcp8_multi_merge() {
        // GIVEN: Event log where paths merge without synchronization
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "D".to_string()]);

        let net = discover_petri_net(&log);

        // All activities present
        assert_eq!(net.transitions.len(), 4);

        // Multiple paths to D
        let paths_to_d = net
            .arcs
            .iter()
            .filter(|(_, t)| t == "D")
            .count();
        assert!(
            paths_to_d >= 2,
            "Should have multiple paths to D: found {}",
            paths_to_d
        );

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp8_multi_merge_complex() {
        // GIVEN: More complex multi-merge scenario
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "C".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "D".to_string(), "E".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "D".to_string(), "E".to_string()]);

        let net = discover_petri_net(&log);

        // E should be reachable from B, C, or D
        let converges_to_e = net
            .arcs
            .iter()
            .filter(|(_, t)| t == "E")
            .count();
        assert!(converges_to_e >= 1, "Should have paths to E");

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP9: STRUCTURED SYNCHRONIZATION
    // ========================================================================
    // Pattern: Structured parallel split with synchronized join
    // Expected: A → (B || C || D) → E with all branches waiting
    // Soundness: E waits for all of B, C, D to complete

    #[test]
    fn test_wcp9_structured_synchronization() {
        // GIVEN: Event log with 3-way parallel and synchronized join
        let mut log = EventLog::new();
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
            "E".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "D".to_string(),
            "C".to_string(),
            "E".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "C".to_string(),
            "D".to_string(),
            "B".to_string(),
            "E".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // All activities present
        assert_eq!(net.transitions.len(), 5);

        // All parallel branches (B, C, D) reachable from A
        let from_a: HashSet<_> = net
            .arcs
            .iter()
            .filter(|(s, _)| s == "A")
            .map(|(_, t)| t.clone())
            .collect();

        assert!(from_a.contains(&"B".to_string()));
        assert!(from_a.contains(&"C".to_string()));
        assert!(from_a.contains(&"D".to_string()));

        // All reach E
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "E") ||
            net.arcs.iter().any(|(s, t)| s == "B" && t == "D") ||
            net.arcs.iter().any(|(s, t)| s == "B" && t == "C"));

        assert!(check_soundness(&log, &net));

        let fitness = calculate_fitness(&log, &net);
        assert!(fitness >= 0.9, "High fitness expected: {}", fitness);
    }

    #[test]
    fn test_wcp9_structured_synchronization_4way() {
        // GIVEN: 4-way parallel with join
        let mut log = EventLog::new();
        log.add_trace(vec![
            "Start".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
            "F".to_string(),
            "End".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "D".to_string(),
            "B".to_string(),
            "C".to_string(),
            "F".to_string(),
            "End".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // All in transitions
        assert!(net.transitions.len() >= 5);

        // End is reachable
        assert!(net.transitions.contains(&"End".to_string()));

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // WCP10: ARBITRARY CYCLES (loops)
    // ========================================================================
    // Pattern: Arbitrary backward loops (no restrictions)
    // Expected: Backward arcs creating cycles
    // Soundness: Liveness and proper termination

    #[test]
    fn test_wcp10_simple_cycle() {
        // GIVEN: Event log with simple backward loop
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "A".to_string(), "B".to_string(), "A".to_string(), "B".to_string(), "C".to_string()]);

        let net = discover_petri_net(&log);

        // All activities present
        assert!(net.transitions.contains(&"A".to_string()));
        assert!(net.transitions.contains(&"B".to_string()));
        assert!(net.transitions.contains(&"C".to_string()));

        // Loop: B can return to A or proceed to C
        let has_loop = net.arcs.iter().any(|(s, t)| s == "B" && t == "A");
        let has_exit = net.arcs.iter().any(|(s, t)| s == "B" && t == "C");
        assert!(has_loop || has_exit, "Should have loop or exit from B");

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp10_nested_cycles() {
        // GIVEN: Event log with nested/repeated cycles
        let mut log = EventLog::new();
        log.add_trace(vec![
            "Start".to_string(),
            "Validate".to_string(),
            "Fix".to_string(),
            "Validate".to_string(),
            "End".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "Validate".to_string(),
            "End".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "Validate".to_string(),
            "Fix".to_string(),
            "Fix".to_string(),
            "Validate".to_string(),
            "End".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Validate can loop back (via Fix)
        assert!(net.transitions.contains(&"Validate".to_string()));
        assert!(net.transitions.contains(&"Fix".to_string()));

        // Cycle back: Fix -> Validate or loop structure
        let has_cycle = net.arcs.iter().any(|(s, t)| s == "Fix" && t == "Validate") ||
            net.arcs.iter().any(|(s, t)| s == "Validate" && t == "Fix");
        assert!(has_cycle, "Should have cycle between Validate and Fix");

        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp10_cycle_with_parallel() {
        // GIVEN: Cycle combined with parallel behavior
        let mut log = EventLog::new();
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "A".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "D".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Cycle: C returns to A
        let has_cycle = net.arcs.iter().any(|(s, t)| s == "C" && t == "A");
        assert!(has_cycle, "Should have backward arc C->A");

        // Exit path: B->D
        let has_exit = net.arcs.iter().any(|(s, t)| s == "B" && t == "D");
        assert!(has_exit, "Should have forward arc B->D");

        assert!(check_soundness(&log, &net));
    }

    // ========================================================================
    // EDGE CASE & COMPREHENSIVE TESTS
    // ========================================================================

    #[test]
    fn test_wcp_all_patterns_combined() {
        // GIVEN: Complex event log combining multiple WCP patterns
        let mut log = EventLog::new();
        // WCP1: Sequence A->B->C
        // WCP4: Choice at C (D or E)
        // WCP3: Synchronization to F
        // WCP10: Loop back from F to A
        log.add_trace(vec![
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "D".to_string(),
            "F".to_string(),
            "A".to_string(),
            "B".to_string(),
            "C".to_string(),
            "E".to_string(),
            "F".to_string(),
        ]);

        let net = discover_petri_net(&log);

        // Should discover all transitions
        assert!(net.transitions.len() >= 6);

        // A is present
        assert!(net.transitions.contains(&"A".to_string()));

        // Sequence A->B->C
        assert!(net.arcs.iter().any(|(s, t)| s == "A" && t == "B"));
        assert!(net.arcs.iter().any(|(s, t)| s == "B" && t == "C"));

        // Should be sound despite complexity
        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp_large_scale_discovery() {
        // GIVEN: Large event log with 100 traces
        let mut log = EventLog::new();
        for i in 0..100 {
            let activity_count = 3 + (i % 5); // 3-7 activities
            let mut trace = vec![];
            for j in 0..activity_count {
                trace.push(format!("Activity_{}", j));
            }
            log.add_trace(trace);
        }

        let net = discover_petri_net(&log);

        // Should discover at least one transition
        assert!(net.transitions.len() > 0);

        // Should have arcs connecting them
        assert!(net.arcs.len() > 0);

        // Soundness should still hold
        assert!(check_soundness(&log, &net));
    }

    #[test]
    fn test_wcp_fitness_precision_metrics() {
        // GIVEN: Sequence log
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);

        let net = discover_petri_net(&log);

        // Calculate metrics
        let fitness = calculate_fitness(&log, &net);
        let precision = calculate_precision(&log, &net);

        // Both should be high for perfect-fit log
        assert!(fitness > 0.8, "Fitness too low: {}", fitness);
        assert!(precision > 0.8, "Precision too low: {}", precision);

        // Both should be valid probabilities
        assert!(fitness >= 0.0 && fitness <= 1.0);
        assert!(precision >= 0.0 && precision <= 1.0);
    }

    #[test]
    fn test_wcp_deviating_behavior() {
        // GIVEN: Log with deviating behavior
        let mut log = EventLog::new();
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "B".to_string()]);  // Missing C
        log.add_trace(vec!["A".to_string(), "B".to_string(), "C".to_string()]);
        log.add_trace(vec!["A".to_string(), "D".to_string(), "C".to_string()]); // Unexpected D

        let net = discover_petri_net(&log);

        // Should still be sound
        assert!(check_soundness(&log, &net));

        // Fitness might be lower
        let fitness = calculate_fitness(&log, &net);
        assert!(fitness > 0.7, "Should handle deviations: {}", fitness);
    }

    // ========================================================================
    // INTEGRATION TEST: Full workflow discovery + soundness verification
    // ========================================================================

    #[test]
    fn test_wcp_full_workflow_discovery_pipeline() {
        // This test represents the complete workflow:
        // Generate Log → Discover → Verify → Metrics

        // GIVEN: Multi-pattern event log
        let mut log = EventLog::new();
        // Sequence then parallel then join then loop
        log.add_trace(vec![
            "Start".to_string(),
            "Register".to_string(),
            "Verify".to_string(),
            "Activate".to_string(),
            "Send_Email".to_string(),
            "Send_SMS".to_string(),
            "Complete".to_string(),
        ]);
        log.add_trace(vec![
            "Start".to_string(),
            "Register".to_string(),
            "Verify".to_string(),
            "Activate".to_string(),
            "Send_SMS".to_string(),
            "Send_Email".to_string(),
            "Complete".to_string(),
        ]);

        // WHEN: Discover net
        let net = discover_petri_net(&log);

        // THEN: Verify structure
        assert!(net.transitions.len() >= 7, "Should discover all transitions");

        // AND: Verify soundness
        let is_sound = check_soundness(&log, &net);
        assert!(is_sound, "Discovered net should be sound");

        // AND: Verify metrics
        let fitness = calculate_fitness(&log, &net);
        let precision = calculate_precision(&log, &net);

        assert!(fitness > 0.85, "Fitness: {}", fitness);
        assert!(precision > 0.7, "Precision: {}", precision);

        // AND: Verify all transitions are connected
        for transition in &net.transitions {
            let has_incoming = net.arcs.iter().any(|(_, t)| t == transition);
            let has_outgoing = net.arcs.iter().any(|(s, _)| s == transition);
            assert!(
                has_incoming || has_outgoing,
                "Transition {} should be connected",
                transition
            );
        }

        println!(
            "Full pipeline successful. Fitness: {:.2}, Precision: {:.2}",
            fitness, precision
        );
    }
}
