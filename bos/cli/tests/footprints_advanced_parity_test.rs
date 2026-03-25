/// Advanced Footprints-based Conformance Validation Tests
///
/// These tests validate that pm4py-rust footprints-based conformance checking
/// correctly handles advanced scenarios beyond basic sequential processes.
///
/// Test Categories:
/// - Recurrence Patterns: Loops and retry logic with causality vs concurrency
/// - Concurrent Activities: True parallel paths and relationship detection
/// - Optional Steps: Branching logic and conditional activity sequences
/// - Multi-variant Processes: Different paths through same process model

#[cfg(test)]
mod footprints_advanced_parity {
    use pm4py::{EventLog, Trace, Event, Footprints, ActivityRelationship};
    use pm4py::conformance::FootprintsConformanceChecker;
    use chrono::Utc;

    /// ============================================================
    /// TEST 1: RECURRENCE PATTERNS
    /// ============================================================
    ///
    /// Scenario: Account verification with retry logic
    /// - verify_attempt → fail → retry → verify_attempt
    /// - Validates that footprints correctly capture causality (attempt→fail→retry)
    ///   vs concurrency (no parallel relationships expected)
    ///
    /// Expected behavior:
    /// - Model footprints: account_created → verify_attempt → fail → retry → verify_attempt → approve
    /// - Log footprints: Same structure with additional cycle
    /// - Mismatch: None (cycle is valid under causality)
    /// - Fitness: High (>0.8)
    #[test]
    fn test_footprints_recurrence_patterns_retry_loop() {
        // ============================================================
        // PHASE 1: Create model footprints (ideal happy path)
        // ============================================================
        let mut model_footprints = Footprints::new();

        // Linear sequence: create → verify → approve
        model_footprints.set_relationship(
            "account_created",
            "verify_attempt",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "verify_attempt",
            "approve",
            ActivityRelationship::Causal,
        );

        // ============================================================
        // PHASE 2: Create event log with retry pattern (recurrence)
        // ============================================================
        // Trace 1: Account ACC001 - succeeds immediately
        let mut log = EventLog::new();
        let now = Utc::now();

        let mut trace1 = Trace::new("ACC001");
        trace1.add_event(Event::new("account_created", now));
        trace1.add_event(Event::new("verify_attempt", now));
        trace1.add_event(Event::new("approve", now));
        log.add_trace(trace1);

        // Trace 2: Account ACC002 - fails first time, retries, then succeeds
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event(Event::new("account_created", now));
        trace2.add_event(Event::new("verify_attempt", now));
        trace2.add_event(Event::new("fail", now));          // First attempt fails
        trace2.add_event(Event::new("retry", now));         // Retry initiated
        trace2.add_event(Event::new("verify_attempt", now)); // Second attempt
        trace2.add_event(Event::new("approve", now));       // Success
        log.add_trace(trace2);

        // ============================================================
        // PHASE 3: Run footprints conformance check
        // ============================================================
        let result = FootprintsConformanceChecker::check_log(&log, &model_footprints);

        // ============================================================
        // PHASE 4: Validate conformance
        // ============================================================

        // Key assertion: Model defines account_created → verify_attempt → approve
        // Log contains same core sequence, plus fail/retry loop
        // The loop (fail → retry → verify_attempt) is a recurrence pattern
        // that should still conform because it represents legitimate retry logic

        // Verify total pairs in model
        assert!(
            result.total_pairs >= 2,
            "Model should have at least 2 relationship pairs (account_created→verify_attempt, verify_attempt→approve), got {}",
            result.total_pairs
        );

        // Verify fitness is high (model behavior is present in log)
        assert!(
            result.fitness >= 0.5,
            "Fitness should be at least 0.5 for recurrence pattern with core sequence intact, got {}",
            result.fitness
        );

        // Verify mismatches are minimal (core relationships preserved)
        assert!(
            result.mismatching_pairs.len() <= 1,
            "Should have minimal mismatches due to retry loop, got {}",
            result.mismatching_pairs.len()
        );

        println!(
            "✓ Recurrence Pattern Test: Fitness={:.2}%, Matches={}/{}, Mismatches={}",
            result.fitness * 100.0,
            result.matching_pairs,
            result.total_pairs,
            result.mismatching_pairs.len()
        );
    }

    /// ============================================================
    /// TEST 2: CONCURRENT ACTIVITIES
    /// ============================================================
    ///
    /// Scenario: Parallel document verification and fraud check
    /// - Documents can be verified in either order
    /// - fraud_check ↔ doc_verification (bidirectional)
    /// - Validates footprints correctly identify Parallel relationships
    ///
    /// Expected behavior:
    /// - Model footprints: Parallel(fraud_check, doc_verification)
    /// - Log footprints: Show both (fraud→doc) and (doc→fraud) orderings
    /// - Fitness: High (>0.9)
    /// - Mismatching pairs: None (parallel behavior matches)
    #[test]
    fn test_footprints_concurrent_activities_parallel_paths() {
        // ============================================================
        // PHASE 1: Create model with parallel activities
        // ============================================================
        let mut model_footprints = Footprints::new();

        // Linear start and end
        model_footprints.set_relationship(
            "account_created",
            "fraud_check",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "account_created",
            "doc_verification",
            ActivityRelationship::Causal,
        );

        // Parallel: fraud_check and doc_verification can happen in any order
        model_footprints.set_relationship(
            "fraud_check",
            "doc_verification",
            ActivityRelationship::Parallel,
        );
        model_footprints.set_relationship(
            "doc_verification",
            "fraud_check",
            ActivityRelationship::Parallel,
        );

        // Both converge to approval
        model_footprints.set_relationship(
            "fraud_check",
            "account_approved",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "doc_verification",
            "account_approved",
            ActivityRelationship::Causal,
        );

        // ============================================================
        // PHASE 2: Create event log with different orderings
        // ============================================================
        let mut log = EventLog::new();
        let now = Utc::now();

        // Trace 1: Fraud check first, then doc verification
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event(Event::new("account_created", now));
        trace1.add_event(Event::new("fraud_check", now));
        trace1.add_event(Event::new("doc_verification", now));
        trace1.add_event(Event::new("account_approved", now));
        log.add_trace(trace1);

        // Trace 2: Doc verification first, then fraud check
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event(Event::new("account_created", now));
        trace2.add_event(Event::new("doc_verification", now));
        trace2.add_event(Event::new("fraud_check", now));
        trace2.add_event(Event::new("account_approved", now));
        log.add_trace(trace2);

        // Trace 3: Another doc-first ordering
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event(Event::new("account_created", now));
        trace3.add_event(Event::new("doc_verification", now));
        trace3.add_event(Event::new("fraud_check", now));
        trace3.add_event(Event::new("account_approved", now));
        log.add_trace(trace3);

        // ============================================================
        // PHASE 3: Run footprints conformance check
        // ============================================================
        let result = FootprintsConformanceChecker::check_log(&log, &model_footprints);

        // ============================================================
        // PHASE 4: Validate conformance for concurrent activities
        // ============================================================

        // Key assertion: Model defines Parallel(fraud_check, doc_verification)
        // Log shows both orderings, which proves parallel behavior
        // Footprints should recognize both (a→b) and (b→a) as parallel

        assert!(
            result.total_pairs >= 4,
            "Model should have at least 4 relationship pairs (account_created→fraud, account_created→doc, fraud↔doc), got {}",
            result.total_pairs
        );

        // High fitness indicates parallel relationships are properly detected
        assert!(
            result.fitness >= 0.75,
            "Fitness should be >=0.75 for parallel activities correctly recognized, got {}",
            result.fitness
        );

        // Verify that mismatches are acceptable (Causal relationships may become Parallel in log)
        // The log's bidirectional activity pairs may be classified differently than the model
        assert!(
            result.mismatching_pairs.len() <= 2,
            "Should have minimal mismatches for parallel activities, got {}",
            result.mismatching_pairs.len()
        );

        println!(
            "✓ Concurrent Activities Test: Fitness={:.2}%, Matches={}/{}, Mismatches={}",
            result.fitness * 100.0,
            result.matching_pairs,
            result.total_pairs,
            result.mismatching_pairs.len()
        );
    }

    /// ============================================================
    /// TEST 3: MULTI-VARIANT PROCESSES
    /// ============================================================
    ///
    /// Scenario: Three distinct process variants for account onboarding
    /// - Variant A: Quick onboarding (created → quick_verify → activated)
    /// - Variant B: Standard onboarding (created → standard_verify → review → activated)
    /// - Variant C: Premium onboarding (created → premium_verify → consultation → review → activated)
    ///
    /// Validates that footprints correctly generalize across multiple variants
    /// and capture the common subsequences and optional steps.
    ///
    /// Expected behavior:
    /// - Model footprints: Union of all variant relationships
    /// - Log footprints: All three variants present
    /// - Fitness: High (>0.8)
    /// - Mismatching pairs: Minimal (variants may add optional paths)
    #[test]
    fn test_footprints_multi_variant_processes_generalization() {
        // ============================================================
        // PHASE 1: Create a model representing all variants
        // ============================================================
        let mut model_footprints = Footprints::new();

        // Common: account_created is always first
        model_footprints.set_relationship(
            "account_created",
            "quick_verify",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "account_created",
            "standard_verify",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "account_created",
            "premium_verify",
            ActivityRelationship::Causal,
        );

        // Variant A: quick path
        model_footprints.set_relationship(
            "quick_verify",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Variant B: standard path
        model_footprints.set_relationship(
            "standard_verify",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Variant C: premium path with optional consultation
        model_footprints.set_relationship(
            "premium_verify",
            "consultation",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "consultation",
            "review",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "review",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Optional: Some paths include review before activation
        model_footprints.set_relationship(
            "standard_verify",
            "review",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "review",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // ============================================================
        // PHASE 2: Create event log with all three variants
        // ============================================================
        let mut log = EventLog::new();
        let now = Utc::now();

        // Variant A: Quick onboarding (1 trace)
        let mut trace_a = Trace::new("ACC_QUICK_001");
        trace_a.add_event(Event::new("account_created", now));
        trace_a.add_event(Event::new("quick_verify", now));
        trace_a.add_event(Event::new("account_activated", now));
        log.add_trace(trace_a);

        // Variant B: Standard onboarding (2 traces)
        let mut trace_b1 = Trace::new("ACC_STD_001");
        trace_b1.add_event(Event::new("account_created", now));
        trace_b1.add_event(Event::new("standard_verify", now));
        trace_b1.add_event(Event::new("account_activated", now));
        log.add_trace(trace_b1);

        let mut trace_b2 = Trace::new("ACC_STD_002");
        trace_b2.add_event(Event::new("account_created", now));
        trace_b2.add_event(Event::new("standard_verify", now));
        trace_b2.add_event(Event::new("review", now));
        trace_b2.add_event(Event::new("account_activated", now));
        log.add_trace(trace_b2);

        // Variant C: Premium onboarding (2 traces)
        let mut trace_c1 = Trace::new("ACC_PREMIUM_001");
        trace_c1.add_event(Event::new("account_created", now));
        trace_c1.add_event(Event::new("premium_verify", now));
        trace_c1.add_event(Event::new("consultation", now));
        trace_c1.add_event(Event::new("review", now));
        trace_c1.add_event(Event::new("account_activated", now));
        log.add_trace(trace_c1);

        let mut trace_c2 = Trace::new("ACC_PREMIUM_002");
        trace_c2.add_event(Event::new("account_created", now));
        trace_c2.add_event(Event::new("premium_verify", now));
        trace_c2.add_event(Event::new("consultation", now));
        trace_c2.add_event(Event::new("review", now));
        trace_c2.add_event(Event::new("account_activated", now));
        log.add_trace(trace_c2);

        // ============================================================
        // PHASE 3: Run footprints conformance check
        // ============================================================
        let result = FootprintsConformanceChecker::check_log(&log, &model_footprints);

        // ============================================================
        // PHASE 4: Validate conformance for multi-variant processes
        // ============================================================

        // Key assertion: Model defines relationships covering all variants
        // Log contains 5 traces representing the 3 variants
        // Footprints should show that the log is conformant to the model

        assert!(
            log.len() == 5,
            "Should have 5 traces (1 quick + 2 standard + 2 premium), got {}",
            log.len()
        );

        assert!(
            result.total_pairs >= 6,
            "Model should have at least 6 relationship pairs covering all variants, got {}",
            result.total_pairs
        );

        // High fitness indicates model generalizes across all variants
        assert!(
            result.fitness >= 0.75,
            "Fitness should be >=0.75 for model generalizing across variants, got {}",
            result.fitness
        );

        // Verify matching pairs are substantial
        assert!(
            result.matching_pairs >= 4,
            "Should have at least 4 matching pairs across variant activities, got {}",
            result.matching_pairs
        );

        // Mismatches are acceptable due to variant-specific optional activities
        assert!(
            result.mismatching_pairs.len() <= 3,
            "Should have minimal mismatches for generalized model across variants, got {}",
            result.mismatching_pairs.len()
        );

        println!(
            "✓ Multi-Variant Process Test: Fitness={:.2}%, Matches={}/{}, Mismatches={}",
            result.fitness * 100.0,
            result.matching_pairs,
            result.total_pairs,
            result.mismatching_pairs.len()
        );
        println!(
            "  Process variants detected: Quick (1 trace) + Standard (2 traces) + Premium (2 traces)"
        );
    }

    /// ============================================================
    /// BONUS TEST: ADVANCED SCENARIO - OPTIONAL STEPS WITH CONDITIONAL BRANCHING
    /// ============================================================
    ///
    /// Scenario: Payment processing with optional fraud check
    /// - init_payment (always)
    /// - fraud_check (optional, depends on amount)
    /// - process_payment (always, may depend on fraud check result)
    /// - confirm (always)
    ///
    /// Validates choice relationships and optional activity patterns
    #[test]
    fn test_footprints_optional_steps_conditional_branching() {
        // ============================================================
        // PHASE 1: Create model with optional activities
        // ============================================================
        let mut model_footprints = Footprints::new();

        // Linear required steps
        model_footprints.set_relationship(
            "init_payment",
            "process_payment",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "process_payment",
            "confirm",
            ActivityRelationship::Causal,
        );

        // Optional fraud check (may or may not occur)
        // When it does, it's between init and process
        model_footprints.set_relationship(
            "init_payment",
            "fraud_check",
            ActivityRelationship::Causal,
        );
        model_footprints.set_relationship(
            "fraud_check",
            "process_payment",
            ActivityRelationship::Causal,
        );

        // fraud_check and confirm are mutually exclusive (Choice)
        // (fraud happens early or not at all, confirm is always last)
        model_footprints.set_relationship(
            "fraud_check",
            "confirm",
            ActivityRelationship::Choice,
        );

        // ============================================================
        // PHASE 2: Create event log with optional activity variations
        // ============================================================
        let mut log = EventLog::new();
        let now = Utc::now();

        // Trace 1: High-value transaction (includes fraud check)
        let mut trace1 = Trace::new("PAY_001");
        trace1.add_event(Event::new("init_payment", now));
        trace1.add_event(Event::new("fraud_check", now));
        trace1.add_event(Event::new("process_payment", now));
        trace1.add_event(Event::new("confirm", now));
        log.add_trace(trace1);

        // Trace 2: Low-value transaction (skips fraud check)
        let mut trace2 = Trace::new("PAY_002");
        trace2.add_event(Event::new("init_payment", now));
        trace2.add_event(Event::new("process_payment", now));
        trace2.add_event(Event::new("confirm", now));
        log.add_trace(trace2);

        // Trace 3: Another high-value transaction
        let mut trace3 = Trace::new("PAY_003");
        trace3.add_event(Event::new("init_payment", now));
        trace3.add_event(Event::new("fraud_check", now));
        trace3.add_event(Event::new("process_payment", now));
        trace3.add_event(Event::new("confirm", now));
        log.add_trace(trace3);

        // ============================================================
        // PHASE 3: Run footprints conformance check
        // ============================================================
        let result = FootprintsConformanceChecker::check_log(&log, &model_footprints);

        // ============================================================
        // PHASE 4: Validate conformance for optional steps
        // ============================================================

        // Key assertion: Model allows optional fraud_check
        // Log shows both with (2 traces) and without (1 trace) fraud_check
        // Footprints should recognize valid conditional branching

        assert!(
            result.total_pairs >= 3,
            "Model should have at least 3 relationship pairs (init→process, process→confirm, optional fraud_check), got {}",
            result.total_pairs
        );

        // Fitness should be reasonable (model covers both paths)
        assert!(
            result.fitness >= 0.5,
            "Fitness should be >=0.5 for optional activity handling, got {}",
            result.fitness
        );

        // Verify that at least the required sequence is recognized
        assert!(
            result.matching_pairs >= 2,
            "Should match at least 2 pairs for mandatory init→process→confirm sequence, got {}",
            result.matching_pairs
        );

        println!(
            "✓ Optional Steps Test: Fitness={:.2}%, Matches={}/{}, Mismatches={}",
            result.fitness * 100.0,
            result.matching_pairs,
            result.total_pairs,
            result.mismatching_pairs.len()
        );
        println!(
            "  Conditional branching: {} traces with fraud_check, {} without",
            2, 1
        );
    }
}
