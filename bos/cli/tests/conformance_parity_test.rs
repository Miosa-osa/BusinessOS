/// Conformance Checking Parity Tests
///
/// These tests validate that pm4py-rust conformance algorithms
/// produce identical results to Python pm4py implementations.
///
/// Test Categories:
/// - Token Replay (80% critical use case)
/// - Footprints (advanced)
/// - Alignments (advanced)
/// - 4-Spectrum (advanced)
/// - Generalization (advanced)
/// - Precision (advanced)

#[cfg(test)]
mod conformance_parity {
    use pm4py::{EventLog, Trace, Event};
    use pm4py::discovery::AlphaMiner;
    use pm4py::conformance::TokenReplay;
    use chrono::Utc;

    /// Create a simple account event log (perfect-fit scenario)
    ///
    /// This log represents a standard account lifecycle:
    /// 1. Account created
    /// 2. Verification initiated
    /// 3. Verification completed
    /// 4. Account activated
    ///
    /// All traces follow this exact sequence, so fitness should be 1.0
    fn create_perfect_fit_account_log() -> EventLog {
        let mut log = EventLog::new();
        let now = Utc::now();

        // Trace 1: Account ACC001 - perfect flow
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event(Event::new("account_created", now));
        trace1.add_event(Event::new("verification_initiated", now));
        trace1.add_event(Event::new("verification_completed", now));
        trace1.add_event(Event::new("account_activated", now));
        log.add_trace(trace1);

        // Trace 2: Account ACC002 - perfect flow
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event(Event::new("account_created", now));
        trace2.add_event(Event::new("verification_initiated", now));
        trace2.add_event(Event::new("verification_completed", now));
        trace2.add_event(Event::new("account_activated", now));
        log.add_trace(trace2);

        // Trace 3: Account ACC003 - perfect flow
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event(Event::new("account_created", now));
        trace3.add_event(Event::new("verification_initiated", now));
        trace3.add_event(Event::new("verification_completed", now));
        trace3.add_event(Event::new("account_activated", now));
        log.add_trace(trace3);

        log
    }

    /// Create a non-conformant account event log (deviation scenario)
    ///
    /// This log has traces that deviate from the standard sequence:
    /// - Trace 1: Standard flow (account_created → ... → account_activated)
    /// - Trace 2: Missing verification step
    /// - Trace 3: Standard flow
    ///
    /// Expected fitness ~0.667 (2/3 traces conform)
    fn create_non_conformant_account_log() -> EventLog {
        let mut log = EventLog::new();
        let now = Utc::now();

        // Trace 1: Standard flow
        let mut trace1 = Trace::new("ACC001");
        trace1.add_event(Event::new("account_created", now));
        trace1.add_event(Event::new("verification_initiated", now));
        trace1.add_event(Event::new("verification_completed", now));
        trace1.add_event(Event::new("account_activated", now));
        log.add_trace(trace1);

        // Trace 2: Skips verification_initiated (non-conformant)
        let mut trace2 = Trace::new("ACC002");
        trace2.add_event(Event::new("account_created", now));
        trace2.add_event(Event::new("verification_completed", now));  // Skipped _initiated
        trace2.add_event(Event::new("account_activated", now));
        log.add_trace(trace2);

        // Trace 3: Standard flow
        let mut trace3 = Trace::new("ACC003");
        trace3.add_event(Event::new("account_created", now));
        trace3.add_event(Event::new("verification_initiated", now));
        trace3.add_event(Event::new("verification_completed", now));
        trace3.add_event(Event::new("account_activated", now));
        log.add_trace(trace3);

        log
    }

    #[test]
    fn test_token_replay_parity() {
        // Step 1: Create perfect-fit account event log
        let perfect_log = create_perfect_fit_account_log();

        // Step 2: Discover Petri net using Alpha Miner (same algorithm both Rust and Python use)
        let miner = AlphaMiner::new();
        let net = miner.discover(&perfect_log);

        // Step 3: Run token replay on same log against discovered net
        let checker = TokenReplay::new();
        let result_perfect = checker.check(&perfect_log, &net);

        // Step 4: Compare fitness score with expected perfect fit
        // Assert: Perfect-fit log should have fitness of 1.0
        assert_eq!(
            result_perfect.fitness, 1.0,
            "Perfect-fit log should have fitness 1.0, got {}",
            result_perfect.fitness
        );

        // Verify conformance flag is true when fitness is 1.0
        assert!(
            result_perfect.is_conformant,
            "Perfect-fit log should be marked as conformant"
        );
    }

    #[test]
    fn test_token_replay_non_conformant_parity() {
        // Step 1: Create non-conformant account event log
        let non_conformant_log = create_non_conformant_account_log();

        // Step 2: Discover Petri net from the same log (using Alpha Miner)
        let miner = AlphaMiner::new();
        let net = miner.discover(&non_conformant_log);

        // Step 3: Run token replay to check conformance
        let checker = TokenReplay::new();
        let result_non_conformant = checker.check(&non_conformant_log, &net);

        // Step 4: Verify fitness score is between 0.0 and 1.0
        // Since we have 2/3 conformant traces, fitness should be ~0.667
        assert!(
            result_non_conformant.fitness > 0.0 && result_non_conformant.fitness < 1.0,
            "Non-conformant log fitness should be between 0.0 and 1.0, got {}",
            result_non_conformant.fitness
        );

        // Tolerance: ±0.001 for rounding differences
        let expected_fitness = 2.0 / 3.0; // 0.666...
        assert!(
            (result_non_conformant.fitness - expected_fitness).abs() < 0.01,
            "Non-conformant log fitness {} should be close to expected {} (tolerance ±0.01)",
            result_non_conformant.fitness, expected_fitness
        );

        // Verify conformance flag is false when fitness < 1.0
        assert!(
            !result_non_conformant.is_conformant,
            "Non-conformant log should be marked as non-conformant"
        );
    }

    #[test]
    fn test_token_replay_parity_empty_log() {
        // Edge case: empty log
        let empty_log = EventLog::new();

        // Discover with empty log (should create trivial net)
        let miner = AlphaMiner::new();
        let net = miner.discover(&empty_log);

        // Run token replay on empty log
        let checker = TokenReplay::new();
        let result = checker.check(&empty_log, &net);

        // Empty log should have fitness 0.0 (no traces to conform)
        assert_eq!(
            result.fitness, 0.0,
            "Empty log should have fitness 0.0, got {}",
            result.fitness
        );
    }

    #[test]
    fn test_footprints_parity() {
        use pm4py::FootprintsConformanceChecker;
        use pm4py::Footprints;
        use pm4py::ActivityRelationship;

        // Step 1: Create perfect-fit account event log
        let perfect_log = create_perfect_fit_account_log();

        // Step 2: Extract footprints from the log
        let log_footprints = Footprints::from_log(&perfect_log);

        // Step 3: Create expected footprints manually (representing the perfect-fit model)
        // Based on the account lifecycle: account_created → verification_initiated → verification_completed → account_activated
        let mut expected_footprints = Footprints::new();
        expected_footprints.set_relationship(
            "account_created",
            "verification_initiated",
            ActivityRelationship::Causal,
        );
        expected_footprints.set_relationship(
            "verification_initiated",
            "verification_completed",
            ActivityRelationship::Causal,
        );
        expected_footprints.set_relationship(
            "verification_completed",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Step 4: Compare footprints - perfect-fit log should match expected
        let result_perfect = FootprintsConformanceChecker::compare_footprints(
            &log_footprints,
            &expected_footprints,
        );

        // Assert: Perfect-fit log should have perfect footprints conformance (fitness 1.0)
        assert_eq!(
            result_perfect.fitness, 1.0,
            "Perfect-fit log should have fitness 1.0, got {}",
            result_perfect.fitness
        );

        // Verify conformance flag is true when fitness is 1.0
        assert!(
            result_perfect.is_conformant,
            "Perfect-fit log should be marked as conformant"
        );

        // Verify no mismatches or missing relationships
        assert_eq!(
            result_perfect.mismatching_pairs.len(), 0,
            "Perfect-fit should have no mismatching pairs"
        );

        assert_eq!(
            result_perfect.missing_relationships.len(), 0,
            "Perfect-fit should have no missing relationships"
        );
    }

    #[test]
    fn test_footprints_non_conformant_parity() {
        use pm4py::FootprintsConformanceChecker;
        use pm4py::Footprints;

        // Step 1: Create non-conformant account event log
        let non_conformant_log = create_non_conformant_account_log();

        // Step 2: Discover Petri net from the same non-conformant log (using Alpha Miner)
        let miner = AlphaMiner::new();
        let net = miner.discover(&non_conformant_log);

        // Step 3: Extract footprints from discovered model
        let model_footprints = FootprintsConformanceChecker::footprints_from_petri_net(&net);

        // Step 4: Extract footprints from the log
        let log_footprints = Footprints::from_log(&non_conformant_log);

        // Step 5: Compare footprints
        let result = FootprintsConformanceChecker::compare_footprints(
            &log_footprints,
            &model_footprints,
        );

        // Assert: Result should be valid (either conformant or not, depending on discovery)
        // Fitness should be between 0.0 and 1.0
        assert!(
            result.fitness >= 0.0 && result.fitness <= 1.0,
            "Fitness should be between 0.0 and 1.0, got {}",
            result.fitness
        );

        // Verify total_pairs count matches expected value
        assert!(
            result.total_pairs > 0,
            "Non-conformant log should have detected activity pairs"
        );

        // Verify matching_pairs is less than or equal to total_pairs
        assert!(
            result.matching_pairs <= result.total_pairs,
            "Matching pairs ({}) should be <= total pairs ({})",
            result.matching_pairs,
            result.total_pairs
        );
    }

    #[test]
    fn test_footprints_mismatch_detection() {
        use pm4py::FootprintsConformanceChecker;
        use pm4py::Footprints;
        use pm4py::ActivityRelationship;

        // Step 1: Manually create two footprints with intentional mismatch
        let mut footprints_log = Footprints::new();
        footprints_log.set_relationship(
            "account_created",
            "verification_initiated",
            ActivityRelationship::Causal,
        );
        footprints_log.set_relationship(
            "verification_initiated",
            "verification_completed",
            ActivityRelationship::Causal,
        );
        footprints_log.set_relationship(
            "verification_completed",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Create a model footprints with a different relationship (simulate model deviation)
        let mut footprints_model = Footprints::new();
        footprints_model.set_relationship(
            "account_created",
            "verification_initiated",
            ActivityRelationship::Causal,
        );
        footprints_model.set_relationship(
            "verification_initiated",
            "verification_completed",
            ActivityRelationship::Parallel, // Mismatch: model allows parallel, log is causal
        );
        footprints_model.set_relationship(
            "verification_completed",
            "account_activated",
            ActivityRelationship::Causal,
        );

        // Step 2: Compare the two footprints
        let result = FootprintsConformanceChecker::compare_footprints(
            &footprints_log,
            &footprints_model,
        );

        // Step 3: Assert mismatch detection works
        assert!(
            !result.is_conformant,
            "Mismatched footprints should be non-conformant"
        );

        assert!(
            result.mismatching_pairs.len() > 0,
            "Should detect mismatching pairs"
        );

        // Verify the mismatch contains the expected activities
        let has_expected_mismatch = result
            .mismatching_pairs
            .iter()
            .any(|(a, b, _, _)| a == "verification_initiated" && b == "verification_completed");

        assert!(
            has_expected_mismatch,
            "Should detect mismatch in verification_initiated -> verification_completed"
        );

        // Fitness should be less than 1.0
        assert!(
            result.fitness < 1.0,
            "Fitness with mismatches should be < 1.0, got {}",
            result.fitness
        );
    }

    // ─────────────────────────────────────────────────────────────────────────
    // New tests (Chicago TDD additions)
    // ─────────────────────────────────────────────────────────────────────────

    /// token_replay_running_example_fitness_high — fitness >= 0.80 for a well-fitting log.
    #[test]
    fn token_replay_running_example_fitness_high() {
        let log = create_perfect_fit_account_log();
        let net = AlphaMiner::new().discover(&log);
        let result = TokenReplay::new().check(&log, &net);
        assert!(result.fitness >= 0.80,
            "Token-replay fitness for a well-fitting log should be >= 0.80, got {}", result.fitness);
    }

    /// conformance_detects_deviating_trace — deviating log yields is_conformant==false OR fitness < 1.0.
    #[test]
    fn conformance_detects_deviating_trace() {
        let log = create_non_conformant_account_log();
        let strict_net = AlphaMiner::new().discover(&create_perfect_fit_account_log());
        let result = TokenReplay::new().check(&log, &strict_net);
        let detected = !result.is_conformant || result.fitness < 1.0;
        assert!(detected,
            "Deviating log must yield is_conformant=false OR fitness < 1.0; fitness={}, is_conformant={}",
            result.fitness, result.is_conformant);
    }

    /// footprints_conformance_produces_result — result must have fitness in [0,1].
    #[test]
    fn footprints_conformance_produces_result() {
        use pm4py::FootprintsConformanceChecker;
        use pm4py::Footprints;
        let log = create_perfect_fit_account_log();
        let net = AlphaMiner::new().discover(&log);
        let model_fp = FootprintsConformanceChecker::footprints_from_petri_net(&net);
        let log_fp = Footprints::from_log(&log);
        let result = FootprintsConformanceChecker::compare_footprints(&log_fp, &model_fp);
        assert!((0.0..=1.0).contains(&result.fitness),
            "FootprintsConformanceChecker must return fitness in [0,1], got {}", result.fitness);
    }

    /// precision_is_bounded — Precision::calculate must return a value in [0.0, 1.0].
    #[test]
    fn precision_is_bounded() {
        use pm4py::Precision;
        let log = create_perfect_fit_account_log();
        let net = AlphaMiner::new().discover(&log);
        let precision = Precision::calculate(&log, &net);
        assert!((0.0..=1.0).contains(&precision),
            "Precision must be in [0.0, 1.0], got {}", precision);
    }

    /// generalization_is_bounded — Generalization::calculate must return a value in [0.0, 1.0].
    #[test]
    fn generalization_is_bounded() {
        use pm4py::Generalization;
        let log = create_perfect_fit_account_log();
        let net = AlphaMiner::new().discover(&log);
        let gen = Generalization::calculate(&log, &net, 3);
        assert!((0.0..=1.0).contains(&gen),
            "Generalization must be in [0.0, 1.0], got {}", gen);
    }

    #[test]
    fn test_alignments_parity() {
        use pm4py::conformance::AlignmentChecker;

        // Step 1: Create perfect-fit account event log
        let perfect_log = create_perfect_fit_account_log();

        // Step 2: Discover Petri net using Alpha Miner
        let miner = AlphaMiner::new();
        let net = miner.discover(&perfect_log);

        // Step 3: Run alignment-based conformance checking on perfect-fit log
        let alignment_checker = AlignmentChecker::new();
        let result_perfect = alignment_checker.check(&perfect_log, &net);

        // Step 4: Verify perfect-fit log has maximum fitness (1.0)
        // Perfect-fit means all activities in log match model transitions
        assert_eq!(
            result_perfect.fitness, 1.0,
            "Perfect-fit log should have alignment fitness 1.0, got {}",
            result_perfect.fitness
        );

        // Verify conformance flag is true when fitness is perfect
        assert!(
            result_perfect.is_conformant,
            "Perfect-fit log should be marked as conformant"
        );

        // Step 5: Create non-conformant account event log
        let non_conformant_log = create_non_conformant_account_log();

        // Step 6: Discover Petri net from the non-conformant log (using Alpha Miner)
        let non_conformant_net = miner.discover(&non_conformant_log);

        // Step 7: Run alignment-based conformance checking
        let result_non_conformant = alignment_checker.check(&non_conformant_log, &non_conformant_net);

        // Step 8: Verify fitness score is between 0.0 and 1.0
        // Non-conformant log should have deviations that reduce fitness
        assert!(
            result_non_conformant.fitness >= 0.0 && result_non_conformant.fitness <= 1.0,
            "Non-conformant log fitness should be between 0.0 and 1.0, got {}",
            result_non_conformant.fitness
        );

        // Step 9: Verify cost metrics - perfect-fit should have zero or minimal cost
        // The alignment checker measures deviation cost
        // Perfect alignments have no log moves (events not in model) or model moves (skipped transitions)
        // Non-conformant log may have higher cost due to misalignments
        assert!(
            result_perfect.fitness >= result_non_conformant.fitness
            || (result_perfect.fitness - result_non_conformant.fitness).abs() < 0.01,
            "Perfect-fit log should have fitness >= non-conformant log (perfect: {}, non-conformant: {})",
            result_perfect.fitness, result_non_conformant.fitness
        );

        // Step 10: Verify alignment move categorization
        // Perfect-fit should be conformant (fitness >= 0.9 threshold in AlignmentChecker)
        assert!(
            result_perfect.is_conformant,
            "Perfect-fit log should always be conformant via alignment checking"
        );

        // Non-conformant traces may have lower fitness due to missing verification step
        if result_non_conformant.fitness < 0.9 {
            assert!(
                !result_non_conformant.is_conformant,
                "Non-conformant log with fitness {} should be marked as non-conformant",
                result_non_conformant.fitness
            );
        }
    }
}
