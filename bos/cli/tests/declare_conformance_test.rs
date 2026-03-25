/// DECLARE Miner Conformance Tests
///
/// DECLARE (Declarative Specification and Verification) is a constraint-based
/// process discovery paradigm. Unlike procedural models (e.g., Petri nets) that
/// explicitly specify allowed sequences, DECLARE discovers the implicit constraints
/// that must be satisfied for a trace to be compliant with the process.
///
/// This test suite validates four core DECLARE constraint types:
///
/// 1. **Existence Constraints** - Activity occurrence requirements
///    Formula: Existence(A, n) = "activity A must occur at least n times in a trace"
///    Use Case: Account creation must happen once per account lifecycle
///    Test: `test_existence_constraint_single_occurrence`
///
/// 2. **Relation Constraints** - Successor/sequence relationships
///    Formula: Succession(A, B) = "if A occurs, then B must eventually occur"
///    Use Case: Account activation follows account creation eventually
///    Test: `test_succession_constraint_order`
///
/// 3. **Forbidden Constraints** - Negative relations
///    Formula: NotSuccession(A, B) = "A and B must never be consecutive"
///    Use Case: Account creation never directly precedes account closure
///    Test: `test_forbidden_direct_succession`
///
/// 4. **Cardinality Constraints** - Maximum occurrence bounds
///    Formula: AtMost(A, m) = "activity A can occur at most m times in a trace"
///    Use Case: Account suspension can occur at most once per lifecycle
///    Test: `test_cardinality_max_occurrence`
///
/// Conformance Metrics Returned:
/// - **Trace Fitness** (0.0-1.0): % of trace satisfying all constraints
/// - **Constraint Violations**: Count of broken constraints per trace
/// - **Overall Compliance**: % of traces satisfying all constraints
///
/// Academic References:
/// - Pesic, M., Schonenberg, H., & van der Aalst, W. M. (2007).
///   "DECLARE: Full Support for Declaratively Specified Processes"
/// - Maggi, F. M., et al. (2013). "Decoupling Execution from Modeling:
///   The Power of Declarative Process Mining"

#[cfg(test)]
mod declare_conformance {
    use std::collections::HashSet;
    use chrono::Utc;

    // ═══════════════════════════════════════════════════════════════════════════════
    // DOMAIN MODELS
    // ═══════════════════════════════════════════════════════════════════════════════

    /// Represents a single event in an event log trace
    #[derive(Clone, Debug, PartialEq, Eq, Hash)]
    struct Event {
        activity: String,
        timestamp: chrono::DateTime<Utc>,
        case_id: String,
    }

    impl Event {
        fn new(activity: &str, timestamp: chrono::DateTime<Utc>, case_id: &str) -> Self {
            Event {
                activity: activity.to_string(),
                timestamp,
                case_id: case_id.to_string(),
            }
        }
    }

    /// A trace is a sequence of events for a single case
    #[derive(Clone, Debug)]
    struct Trace {
        case_id: String,
        events: Vec<Event>,
    }

    impl Trace {
        #[allow(dead_code)]
        fn new(case_id: &str) -> Self {
            Trace {
                case_id: case_id.to_string(),
                events: Vec::new(),
            }
        }

        fn add_event(&mut self, activity: &str, timestamp: chrono::DateTime<Utc>) {
            self.events.push(Event::new(activity, timestamp, &self.case_id));
        }

        fn activity_sequence(&self) -> Vec<String> {
            self.events.iter().map(|e| e.activity.clone()).collect()
        }

        #[allow(dead_code)]
        fn get_activities(&self) -> Vec<&str> {
            self.events.iter().map(|e| e.activity.as_str()).collect()
        }

        fn activity_count(&self, activity: &str) -> usize {
            self.events.iter().filter(|e| e.activity == activity).count()
        }

        #[allow(dead_code)]
        fn is_empty(&self) -> bool {
            self.events.is_empty()
        }

        #[allow(dead_code)]
        fn len(&self) -> usize {
            self.events.len()
        }
    }

    /// An event log is a collection of traces
    #[derive(Clone, Debug)]
    struct EventLog {
        traces: Vec<Trace>,
    }

    impl EventLog {
        #[allow(dead_code)]
        fn new() -> Self {
            EventLog {
                traces: Vec::new(),
            }
        }

        fn add_trace(&mut self, trace: Trace) {
            self.traces.push(trace);
        }

        #[allow(dead_code)]
        fn num_traces(&self) -> usize {
            self.traces.len()
        }

        #[allow(dead_code)]
        fn get_activities(&self) -> HashSet<String> {
            let mut activities = HashSet::new();
            for trace in &self.traces {
                for event in &trace.events {
                    activities.insert(event.activity.clone());
                }
            }
            activities
        }
    }

    /// Result of DECLARE constraint checking on a single trace
    #[allow(dead_code)]
    #[derive(Clone, Debug)]
    struct TraceConformanceResult {
        case_id: String,
        is_conformant: bool,
        constraint_violations: Vec<String>,
        violated_count: usize,
        fitness: f64,
    }

    /// Aggregated results for all traces and constraints in a log
    #[allow(dead_code)]
    #[derive(Clone, Debug)]
    struct LogConformanceResult {
        total_traces: usize,
        conformant_traces: usize,
        non_conformant_traces: usize,
        overall_compliance: f64,
        total_constraint_violations: usize,
        average_fitness: f64,
        trace_results: Vec<TraceConformanceResult>,
    }

    impl LogConformanceResult {
        #[allow(dead_code)]
        fn new() -> Self {
            LogConformanceResult {
                total_traces: 0,
                conformant_traces: 0,
                non_conformant_traces: 0,
                overall_compliance: 0.0,
                total_constraint_violations: 0,
                average_fitness: 0.0,
                trace_results: Vec::new(),
            }
        }

        fn from_trace_results(results: Vec<TraceConformanceResult>) -> Self {
            let total_traces = results.len();
            let conformant_traces = results.iter().filter(|r| r.is_conformant).count();
            let non_conformant_traces = total_traces - conformant_traces;
            let total_constraint_violations: usize = results.iter().map(|r| r.violated_count).sum();
            let average_fitness: f64 = results.iter().map(|r| r.fitness).sum::<f64>() / total_traces as f64;

            LogConformanceResult {
                total_traces,
                conformant_traces,
                non_conformant_traces,
                overall_compliance: (conformant_traces as f64) / (total_traces as f64),
                total_constraint_violations,
                average_fitness,
                trace_results: results,
            }
        }
    }

    // ═══════════════════════════════════════════════════════════════════════════════
    // DECLARE CONSTRAINT VALIDATORS
    // ═══════════════════════════════════════════════════════════════════════════════

    /// Validator for Existence constraints
    /// Constraint: Existence(A, n) - Activity A must occur at least n times
    struct ExistenceConstraint {
        activity: String,
        min_count: usize,
    }

    impl ExistenceConstraint {
        fn new(activity: &str, min_count: usize) -> Self {
            ExistenceConstraint {
                activity: activity.to_string(),
                min_count,
            }
        }

        /// Check if a trace satisfies this constraint
        fn check(&self, trace: &Trace) -> (bool, String) {
            let actual_count = trace.activity_count(&self.activity);
            let satisfied = actual_count >= self.min_count;
            let message = if satisfied {
                format!("✓ Existence({}, {}) satisfied: {} ≥ {}",
                    self.activity, self.min_count, actual_count, self.min_count)
            } else {
                format!("✗ Existence({}, {}) violated: {} < {}",
                    self.activity, self.min_count, actual_count, self.min_count)
            };
            (satisfied, message)
        }
    }

    /// Validator for Succession constraints
    /// Constraint: Succession(A, B) - If A occurs, then B must eventually occur after A
    struct SuccessionConstraint {
        predecessor: String,
        successor: String,
    }

    impl SuccessionConstraint {
        fn new(predecessor: &str, successor: &str) -> Self {
            SuccessionConstraint {
                predecessor: predecessor.to_string(),
                successor: successor.to_string(),
            }
        }

        /// Check if a trace satisfies this constraint
        fn check(&self, trace: &Trace) -> (bool, String) {
            let activities = trace.activity_sequence();

            // Find all positions where predecessor occurs
            let predecessor_positions: Vec<usize> = activities
                .iter()
                .enumerate()
                .filter(|(_, a)| *a == &self.predecessor)
                .map(|(i, _)| i)
                .collect();

            // If predecessor never occurs, constraint is vacuously satisfied
            if predecessor_positions.is_empty() {
                let message = format!("✓ Succession({}, {}) vacuously satisfied: {} never occurs",
                    self.predecessor, self.successor, self.predecessor);
                return (true, message);
            }

            // For each occurrence of predecessor, check if successor eventually occurs after it
            let mut satisfied = true;
            for &pred_pos in &predecessor_positions {
                let successor_after = activities[pred_pos + 1..].iter().any(|a| a == &self.successor);
                if !successor_after {
                    satisfied = false;
                    break;
                }
            }

            let message = if satisfied {
                format!("✓ Succession({}, {}) satisfied: {} always followed by {}",
                    self.predecessor, self.successor, self.predecessor, self.successor)
            } else {
                format!("✗ Succession({}, {}) violated: {} sometimes not followed by {}",
                    self.predecessor, self.successor, self.predecessor, self.successor)
            };
            (satisfied, message)
        }
    }

    /// Validator for NotSuccession (Forbidden Direct Succession) constraints
    /// Constraint: NotSuccession(A, B) - A and B must never be directly consecutive
    struct NotSuccessionConstraint {
        activity_a: String,
        activity_b: String,
    }

    impl NotSuccessionConstraint {
        fn new(activity_a: &str, activity_b: &str) -> Self {
            NotSuccessionConstraint {
                activity_a: activity_a.to_string(),
                activity_b: activity_b.to_string(),
            }
        }

        /// Check if a trace satisfies this constraint
        fn check(&self, trace: &Trace) -> (bool, String) {
            let activities = trace.activity_sequence();

            // Check for direct succession pattern: activity_a followed immediately by activity_b
            let mut has_direct_succession = false;
            for i in 0..activities.len() - 1 {
                if activities[i] == self.activity_a && activities[i + 1] == self.activity_b {
                    has_direct_succession = true;
                    break;
                }
            }

            let satisfied = !has_direct_succession;
            let message = if satisfied {
                format!("✓ NotSuccession({}, {}) satisfied: {} never directly precedes {}",
                    self.activity_a, self.activity_b, self.activity_a, self.activity_b)
            } else {
                format!("✗ NotSuccession({}, {}) violated: {} directly precedes {} at least once",
                    self.activity_a, self.activity_b, self.activity_a, self.activity_b)
            };
            (satisfied, message)
        }
    }

    /// Validator for AtMost (Cardinality) constraints
    /// Constraint: AtMost(A, m) - Activity A can occur at most m times in a trace
    struct AtMostConstraint {
        activity: String,
        max_count: usize,
    }

    impl AtMostConstraint {
        fn new(activity: &str, max_count: usize) -> Self {
            AtMostConstraint {
                activity: activity.to_string(),
                max_count,
            }
        }

        /// Check if a trace satisfies this constraint
        fn check(&self, trace: &Trace) -> (bool, String) {
            let actual_count = trace.activity_count(&self.activity);
            let satisfied = actual_count <= self.max_count;
            let message = if satisfied {
                format!("✓ AtMost({}, {}) satisfied: {} ≤ {}",
                    self.activity, self.max_count, actual_count, self.max_count)
            } else {
                format!("✗ AtMost({}, {}) violated: {} > {}",
                    self.activity, self.max_count, actual_count, self.max_count)
            };
            (satisfied, message)
        }
    }

    // ═══════════════════════════════════════════════════════════════════════════════
    // DECLARE CHECKER
    // ═══════════════════════════════════════════════════════════════════════════════

    /// Main DECLARE conformance checker
    struct DeclareChecker {
        existence_constraints: Vec<ExistenceConstraint>,
        succession_constraints: Vec<SuccessionConstraint>,
        not_succession_constraints: Vec<NotSuccessionConstraint>,
        at_most_constraints: Vec<AtMostConstraint>,
    }

    impl DeclareChecker {
        fn new() -> Self {
            DeclareChecker {
                existence_constraints: Vec::new(),
                succession_constraints: Vec::new(),
                not_succession_constraints: Vec::new(),
                at_most_constraints: Vec::new(),
            }
        }

        fn add_existence(&mut self, activity: &str, min_count: usize) {
            self.existence_constraints.push(ExistenceConstraint::new(activity, min_count));
        }

        fn add_succession(&mut self, predecessor: &str, successor: &str) {
            self.succession_constraints.push(SuccessionConstraint::new(predecessor, successor));
        }

        fn add_not_succession(&mut self, activity_a: &str, activity_b: &str) {
            self.not_succession_constraints.push(NotSuccessionConstraint::new(activity_a, activity_b));
        }

        fn add_at_most(&mut self, activity: &str, max_count: usize) {
            self.at_most_constraints.push(AtMostConstraint::new(activity, max_count));
        }

        /// Check a single trace against all constraints
        fn check_trace(&self, trace: &Trace) -> TraceConformanceResult {
            let mut violations = Vec::new();
            let mut violated_count = 0;

            // Check all existence constraints
            for constraint in &self.existence_constraints {
                let (satisfied, message) = constraint.check(trace);
                if !satisfied {
                    violations.push(message);
                    violated_count += 1;
                }
            }

            // Check all succession constraints
            for constraint in &self.succession_constraints {
                let (satisfied, message) = constraint.check(trace);
                if !satisfied {
                    violations.push(message);
                    violated_count += 1;
                }
            }

            // Check all not-succession constraints
            for constraint in &self.not_succession_constraints {
                let (satisfied, message) = constraint.check(trace);
                if !satisfied {
                    violations.push(message);
                    violated_count += 1;
                }
            }

            // Check all at-most constraints
            for constraint in &self.at_most_constraints {
                let (satisfied, message) = constraint.check(trace);
                if !satisfied {
                    violations.push(message);
                    violated_count += 1;
                }
            }

            let total_constraints = self.existence_constraints.len()
                + self.succession_constraints.len()
                + self.not_succession_constraints.len()
                + self.at_most_constraints.len();

            let fitness = if total_constraints > 0 {
                ((total_constraints - violated_count) as f64) / (total_constraints as f64)
            } else {
                1.0
            };

            let is_conformant = violated_count == 0;

            TraceConformanceResult {
                case_id: trace.case_id.clone(),
                is_conformant,
                constraint_violations: violations,
                violated_count,
                fitness,
            }
        }

        /// Check an entire event log
        fn check_log(&self, log: &EventLog) -> LogConformanceResult {
            let results: Vec<TraceConformanceResult> = log.traces
                .iter()
                .map(|trace| self.check_trace(trace))
                .collect();

            LogConformanceResult::from_trace_results(results)
        }
    }

    // ═══════════════════════════════════════════════════════════════════════════════
    // TEST DATA FIXTURES
    // ═══════════════════════════════════════════════════════════════════════════════

    /// Create an event log representing perfect account lifecycle conforming to all constraints
    fn create_conformant_account_log() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Trace 1: Perfect account lifecycle
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event("account_created", base_time);
        trace1.add_event("account_activated", base_time);
        trace1.add_event("account_closed", base_time);
        log.add_trace(trace1);

        // Trace 2: Perfect account lifecycle
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event("account_created", base_time);
        trace2.add_event("account_activated", base_time);
        trace2.add_event("account_closed", base_time);
        log.add_trace(trace2);

        // Trace 3: Perfect account lifecycle
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event("account_created", base_time);
        trace3.add_event("account_activated", base_time);
        trace3.add_event("account_closed", base_time);
        log.add_trace(trace3);

        log
    }

    /// Create an event log violating existence constraint (missing required activity)
    fn create_log_violating_existence() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Trace 1: Conformant - has account_created
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event("account_created", base_time);
        trace1.add_event("account_activated", base_time);
        trace1.add_event("account_closed", base_time);
        log.add_trace(trace1);

        // Trace 2: Non-conformant - MISSING account_created
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event("account_activated", base_time);
        trace2.add_event("account_closed", base_time);
        log.add_trace(trace2);

        // Trace 3: Conformant - has account_created
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event("account_created", base_time);
        trace3.add_event("account_activated", base_time);
        trace3.add_event("account_closed", base_time);
        log.add_trace(trace3);

        log
    }

    /// Create an event log violating succession constraint (successor doesn't occur)
    fn create_log_violating_succession() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Trace 1: Conformant - account_activated follows account_created
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event("account_created", base_time);
        trace1.add_event("account_activated", base_time);
        trace1.add_event("account_closed", base_time);
        log.add_trace(trace1);

        // Trace 2: Non-conformant - account_created without account_activated
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event("account_created", base_time);
        trace2.add_event("account_closed", base_time);
        log.add_trace(trace2);

        // Trace 3: Conformant - account_activated follows account_created
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event("account_created", base_time);
        trace3.add_event("account_activated", base_time);
        trace3.add_event("account_closed", base_time);
        log.add_trace(trace3);

        log
    }

    /// Create an event log violating forbidden succession constraint
    fn create_log_violating_forbidden_succession() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Trace 1: Conformant - account_created never directly precedes account_closed
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event("account_created", base_time);
        trace1.add_event("account_activated", base_time);
        trace1.add_event("account_closed", base_time);
        log.add_trace(trace1);

        // Trace 2: Non-conformant - account_created DIRECTLY precedes account_closed
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event("account_created", base_time);
        trace2.add_event("account_closed", base_time);
        log.add_trace(trace2);

        // Trace 3: Conformant - account_created never directly precedes account_closed
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event("account_created", base_time);
        trace3.add_event("account_activated", base_time);
        trace3.add_event("account_closed", base_time);
        log.add_trace(trace3);

        log
    }

    /// Create an event log violating cardinality constraint (too many occurrences)
    fn create_log_violating_cardinality() -> EventLog {
        let mut log = EventLog::new();
        let base_time = Utc::now();

        // Trace 1: Conformant - account_suspended occurs only once
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event("account_created", base_time);
        trace1.add_event("account_suspended", base_time);
        trace1.add_event("account_activated", base_time);
        trace1.add_event("account_closed", base_time);
        log.add_trace(trace1);

        // Trace 2: Non-conformant - account_suspended occurs TWICE (exceeds max of 1)
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event("account_created", base_time);
        trace2.add_event("account_suspended", base_time);
        trace2.add_event("account_activated", base_time);
        trace2.add_event("account_suspended", base_time);  // Second suspension!
        trace2.add_event("account_closed", base_time);
        log.add_trace(trace2);

        // Trace 3: Conformant - account_suspended occurs only once
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event("account_created", base_time);
        trace3.add_event("account_suspended", base_time);
        trace3.add_event("account_activated", base_time);
        trace3.add_event("account_closed", base_time);
        log.add_trace(trace3);

        log
    }

    // ═══════════════════════════════════════════════════════════════════════════════
    // TESTS
    // ═══════════════════════════════════════════════════════════════════════════════

    /// Test 1: Existence Constraint
    /// Constraint: "account_created must occur at least 1 time per case"
    /// Expected: 1 trace violates (ACC002), 2 traces conform
    #[test]
    fn test_existence_constraint_single_occurrence() {
        let log = create_log_violating_existence();

        let mut checker = DeclareChecker::new();
        checker.add_existence("account_created", 1);

        let result = checker.check_log(&log);

        // Verify metrics
        assert_eq!(result.total_traces, 3, "Log should have 3 traces");
        assert_eq!(result.conformant_traces, 2, "2 traces should conform (have account_created)");
        assert_eq!(result.non_conformant_traces, 1, "1 trace should violate (missing account_created)");
        assert!(
            result.overall_compliance > 0.6 && result.overall_compliance < 0.7,
            "Overall compliance should be ~0.667 (2/3), got {}",
            result.overall_compliance
        );
        assert_eq!(
            result.total_constraint_violations, 1,
            "Exactly 1 constraint violation expected"
        );

        // Verify specific trace results
        let trace2_result = &result.trace_results[1];  // ACC002
        assert!(!trace2_result.is_conformant, "ACC002 should be non-conformant");
        assert_eq!(trace2_result.violated_count, 1, "ACC002 should have 1 violation");
        assert!(
            trace2_result.constraint_violations.len() > 0,
            "ACC002 should have violation messages"
        );
        assert!(
            trace2_result.constraint_violations[0].contains("Existence"),
            "Violation message should mention Existence constraint"
        );

        println!("✓ Existence constraint test passed!");
        println!("  Overall compliance: {:.1}%", result.overall_compliance * 100.0);
        println!("  Conformant traces: {}/{}", result.conformant_traces, result.total_traces);
    }

    /// Test 2: Succession Constraint
    /// Constraint: "if account_created, then account_activated must eventually occur"
    /// Expected: 1 trace violates (ACC002), 2 traces conform
    #[test]
    fn test_succession_constraint_order() {
        let log = create_log_violating_succession();

        let mut checker = DeclareChecker::new();
        checker.add_succession("account_created", "account_activated");

        let result = checker.check_log(&log);

        // Verify metrics
        assert_eq!(result.total_traces, 3, "Log should have 3 traces");
        assert_eq!(result.conformant_traces, 2, "2 traces should conform");
        assert_eq!(result.non_conformant_traces, 1, "1 trace should violate");
        assert!(
            result.overall_compliance > 0.6 && result.overall_compliance < 0.7,
            "Overall compliance should be ~0.667 (2/3), got {}",
            result.overall_compliance
        );

        // Verify specific trace results
        let trace2_result = &result.trace_results[1];  // ACC002
        assert!(!trace2_result.is_conformant, "ACC002 should be non-conformant");
        assert_eq!(trace2_result.violated_count, 1, "ACC002 should have 1 violation");
        assert!(
            trace2_result.constraint_violations[0].contains("Succession"),
            "Violation should mention Succession constraint"
        );

        println!("✓ Succession constraint test passed!");
        println!("  Overall compliance: {:.1}%", result.overall_compliance * 100.0);
    }

    /// Test 3: Forbidden Direct Succession Constraint
    /// Constraint: "account_created must never directly precede account_closed"
    /// Expected: 1 trace violates (ACC002), 2 traces conform
    #[test]
    fn test_forbidden_direct_succession() {
        let log = create_log_violating_forbidden_succession();

        let mut checker = DeclareChecker::new();
        checker.add_not_succession("account_created", "account_closed");

        let result = checker.check_log(&log);

        // Verify metrics
        assert_eq!(result.total_traces, 3, "Log should have 3 traces");
        assert_eq!(result.conformant_traces, 2, "2 traces should conform");
        assert_eq!(result.non_conformant_traces, 1, "1 trace should violate");
        assert!(
            result.overall_compliance > 0.6 && result.overall_compliance < 0.7,
            "Overall compliance should be ~0.667 (2/3), got {}",
            result.overall_compliance
        );

        // Verify specific trace results
        let trace2_result = &result.trace_results[1];  // ACC002
        assert!(!trace2_result.is_conformant, "ACC002 should be non-conformant");
        assert_eq!(trace2_result.violated_count, 1, "ACC002 should have 1 violation");
        assert!(
            trace2_result.constraint_violations[0].contains("NotSuccession"),
            "Violation should mention NotSuccession constraint"
        );

        println!("✓ Forbidden direct succession constraint test passed!");
        println!("  Overall compliance: {:.1}%", result.overall_compliance * 100.0);
    }

    /// Test 4: Cardinality (AtMost) Constraint
    /// Constraint: "account_suspended can occur at most 1 time per case"
    /// Expected: 1 trace violates (ACC002), 2 traces conform
    #[test]
    fn test_cardinality_max_occurrence() {
        let log = create_log_violating_cardinality();

        let mut checker = DeclareChecker::new();
        checker.add_at_most("account_suspended", 1);

        let result = checker.check_log(&log);

        // Verify metrics
        assert_eq!(result.total_traces, 3, "Log should have 3 traces");
        assert_eq!(result.conformant_traces, 2, "2 traces should conform");
        assert_eq!(result.non_conformant_traces, 1, "1 trace should violate");
        assert!(
            result.overall_compliance > 0.6 && result.overall_compliance < 0.7,
            "Overall compliance should be ~0.667 (2/3), got {}",
            result.overall_compliance
        );

        // Verify specific trace results
        let trace2_result = &result.trace_results[1];  // ACC002
        assert!(!trace2_result.is_conformant, "ACC002 should be non-conformant");
        assert_eq!(trace2_result.violated_count, 1, "ACC002 should have 1 violation");
        assert!(
            trace2_result.constraint_violations[0].contains("AtMost"),
            "Violation should mention AtMost constraint"
        );

        println!("✓ Cardinality constraint test passed!");
        println!("  Overall compliance: {:.1}%", result.overall_compliance * 100.0);
    }

    /// Bonus Test: All constraints together (perfect conformance)
    #[test]
    fn test_all_constraints_conformant_log() {
        let log = create_conformant_account_log();

        let mut checker = DeclareChecker::new();
        checker.add_existence("account_created", 1);
        checker.add_succession("account_created", "account_activated");
        checker.add_not_succession("account_created", "account_closed");
        checker.add_at_most("account_suspended", 1);

        let result = checker.check_log(&log);

        // Perfect conformance: all traces should satisfy all constraints
        assert_eq!(result.total_traces, 3);
        assert_eq!(result.conformant_traces, 3, "All 3 traces should be fully conformant");
        assert_eq!(result.non_conformant_traces, 0);
        assert_eq!(result.overall_compliance, 1.0, "Compliance should be 100%");
        assert_eq!(result.total_constraint_violations, 0);

        println!("✓ All constraints conformant test passed!");
        println!("  Perfect compliance: {:.1}%", result.overall_compliance * 100.0);
    }
}
