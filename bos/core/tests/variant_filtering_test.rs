//! Process Variant Discovery and Filtering Tests
//!
//! Comprehensive test suite for process variant analysis, discovery, and filtering.
//!
//! This module tests critical capabilities for understanding process behavior:
//! - Variant Discovery: Identify all distinct activity sequences in event logs
//! - Variant Filtering: Filter variants by frequency, activity requirements, and anomalies
//! - Coverage Analysis: Calculate cumulative coverage percentages (80/20 rule)
//! - Anomaly Detection: Flag rare variants as potential process anomalies
//!
//! Test Strategy:
//! 1. Create realistic event logs with multiple distinct process paths
//! 2. Extract variants (unique activity sequences)
//! 3. Calculate variant metrics (frequency, coverage percentage)
//! 4. Apply filters (minimum support, required activities, anomaly threshold)
//! 5. Verify filtering results match expected behavior

use chrono::Utc;
use pm4py::log::{Event, EventLog, Trace};
use std::collections::BTreeMap;

/// Represents a unique variant (activity sequence)
#[derive(Debug, Clone, Eq, PartialEq, Ord, PartialOrd)]
pub struct Variant {
    /// Sequence of activities in order
    pub sequence: Vec<String>,
    /// Number of traces following this variant
    pub frequency: usize,
}

impl Variant {
    /// Create a new variant
    pub fn new(sequence: Vec<String>, frequency: usize) -> Self {
        Self { sequence, frequency }
    }

    /// Get the coverage percentage for this variant relative to total traces
    pub fn coverage_percentage(&self, total_traces: usize) -> f64 {
        if total_traces == 0 {
            0.0
        } else {
            (self.frequency as f64 / total_traces as f64) * 100.0
        }
    }

    /// Check if this variant contains a required activity
    pub fn contains_activity(&self, activity: &str) -> bool {
        self.sequence.iter().any(|a| a == activity)
    }

    /// Get the variant key (comma-separated sequence for display)
    pub fn key(&self) -> String {
        self.sequence.join(",")
    }
}

/// Variant analysis results
#[derive(Debug, Clone)]
pub struct VariantAnalysis {
    /// All discovered variants sorted by frequency (descending)
    pub variants: Vec<Variant>,
    /// Total number of traces analyzed
    pub total_traces: usize,
    /// Total number of distinct variants
    pub variant_count: usize,
}

impl VariantAnalysis {
    /// Create new variant analysis
    pub fn new(variants: Vec<Variant>, total_traces: usize) -> Self {
        let variant_count = variants.len();
        Self {
            variants,
            total_traces,
            variant_count,
        }
    }

    /// Filter variants by minimum support percentage
    pub fn filter_by_frequency(&self, min_support_percent: f64) -> Vec<Variant> {
        let min_frequency = (self.total_traces as f64 * min_support_percent / 100.0).ceil() as usize;
        self.variants
            .iter()
            .filter(|v| v.frequency >= min_frequency)
            .cloned()
            .collect()
    }

    /// Filter variants that contain a required activity
    pub fn filter_by_activity(&self, required_activity: &str) -> Vec<Variant> {
        self.variants
            .iter()
            .filter(|v| v.contains_activity(required_activity))
            .cloned()
            .collect()
    }

    /// Calculate cumulative coverage percentage
    pub fn calculate_coverage_percentage(&self, variants: &[Variant]) -> f64 {
        let total_covered: usize = variants.iter().map(|v| v.frequency).sum();
        if self.total_traces == 0 {
            0.0
        } else {
            (total_covered as f64 / self.total_traces as f64) * 100.0
        }
    }

    /// Identify anomalous variants (rare/unusual patterns)
    /// A variant is anomalous if its frequency is below the anomaly_threshold_percent
    pub fn detect_anomalies(&self, anomaly_threshold_percent: f64) -> Vec<Variant> {
        self.variants
            .iter()
            .filter(|v| v.coverage_percentage(self.total_traces) < anomaly_threshold_percent)
            .cloned()
            .collect()
    }
}

/// Extract all variants from an event log
///
/// A variant is a unique sequence of activities. This function identifies all distinct
/// activity sequences and counts how many traces follow each sequence.
pub fn discover_variants(log: &EventLog) -> VariantAnalysis {
    let mut variant_map: BTreeMap<Vec<String>, usize> = BTreeMap::new();

    // Count occurrences of each activity sequence
    for trace in &log.traces {
        let sequence: Vec<String> = trace
            .events
            .iter()
            .map(|e| e.activity.clone())
            .collect();

        *variant_map.entry(sequence).or_insert(0) += 1;
    }

    // Convert to Variant structs and sort by frequency (descending)
    let mut variants: Vec<Variant> = variant_map
        .into_iter()
        .map(|(seq, freq)| Variant::new(seq, freq))
        .collect();

    variants.sort_by(|a, b| b.frequency.cmp(&a.frequency));

    VariantAnalysis::new(variants, log.traces.len())
}
// ============================================================================
// TEST 1: Variant Discovery
// ============================================================================

#[test]
fn test_variant_discovery() {
    // === SETUP: Create event log with 10 traces, 3 distinct paths ===

    // Distribute 10 traces: 4 follow path1, 4 follow path2, 2 follow path3
    let mut log = EventLog::new();
    let base_time = Utc::now();

    // Path 1: 4 traces (40%)
    for i in 0..4 {
        let mut trace = Trace::new(format!("trace_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("verification_initiated", base_time + chrono::Duration::minutes(10)));
        trace.add_event(Event::new("account_activated", base_time + chrono::Duration::minutes(20)));
        log.add_trace(trace);
    }

    // Path 2: 4 traces (40%)
    for i in 4..8 {
        let mut trace = Trace::new(format!("trace_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("verification_initiated", base_time + chrono::Duration::minutes(10)));
        trace.add_event(Event::new("verification_completed", base_time + chrono::Duration::minutes(20)));
        trace.add_event(Event::new("account_activated", base_time + chrono::Duration::minutes(30)));
        log.add_trace(trace);
    }

    // Path 3: 2 traces (20%)
    for i in 8..10 {
        let mut trace = Trace::new(format!("trace_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("account_activated", base_time + chrono::Duration::minutes(10)));
        log.add_trace(trace);
    }

    // === EXECUTE: Discover variants ===
    let analysis = discover_variants(&log);

    // === VERIFY: Check results ===
    assert_eq!(log.traces.len(), 10, "Log should contain 10 traces");
    assert_eq!(analysis.variant_count, 3, "Should discover exactly 3 variants");

    // Check variant frequencies
    assert_eq!(analysis.variants[0].frequency, 4, "Most frequent variant should appear 4 times");
    assert_eq!(analysis.variants[1].frequency, 4, "Second variant should appear 4 times");
    assert_eq!(analysis.variants[2].frequency, 2, "Third variant should appear 2 times");

    // Check variant sequences
    assert_eq!(
        analysis.variants[0].sequence,
        vec!["account_created", "verification_initiated", "account_activated"]
    );
    assert_eq!(
        analysis.variants[1].sequence,
        vec!["account_created", "verification_initiated", "verification_completed", "account_activated"]
    );
    assert_eq!(
        analysis.variants[2].sequence,
        vec!["account_created", "account_activated"]
    );

    // Verify total traces sum to 10
    let total_covered: usize = analysis.variants.iter().map(|v| v.frequency).sum();
    assert_eq!(total_covered, 10, "All traces should be covered by variants");
}

// ============================================================================
// TEST 2: Variant Filtering by Frequency
// ============================================================================

#[test]
fn test_variant_filtering_by_frequency() {
    // === SETUP: Create log with 5 variants, different frequencies ===
    let mut log = EventLog::new();
    let base_time = Utc::now();

    // Variant 1: A→B→C (30 traces, 30%)
    for i in 0..30 {
        let mut trace = Trace::new(format!("trace_v1_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 2: A→C→B (25 traces, 25%)
    for i in 0..25 {
        let mut trace = Trace::new(format!("trace_v2_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 3: B→A→C (20 traces, 20%)
    for i in 0..20 {
        let mut trace = Trace::new(format!("trace_v3_{}", i));
        trace.add_event(Event::new("B", base_time));
        trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 4: B→C→A (15 traces, 15%)
    for i in 0..15 {
        let mut trace = Trace::new(format!("trace_v4_{}", i));
        trace.add_event(Event::new("B", base_time));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 5: C→A→B (10 traces, 10%)
    for i in 0..10 {
        let mut trace = Trace::new(format!("trace_v5_{}", i));
        trace.add_event(Event::new("C", base_time));
        trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // === EXECUTE: Discover and filter variants ===
    let analysis = discover_variants(&log);
    assert_eq!(analysis.variant_count, 5, "Should discover 5 variants");

    // Filter by minimum support = 30%
    // Only variants with frequency >= 30 traces should remain
    let filtered_30pct = analysis.filter_by_frequency(30.0);
    assert_eq!(filtered_30pct.len(), 1, "Only 1 variant covers ≥30% of cases");
    assert_eq!(filtered_30pct[0].frequency, 30);

    // Filter by minimum support = 20%
    // Variants with frequency >= 20 traces should remain
    let filtered_20pct = analysis.filter_by_frequency(20.0);
    assert_eq!(filtered_20pct.len(), 3, "3 variants cover ≥20% of cases");
    let covered_20pct: usize = filtered_20pct.iter().map(|v| v.frequency).sum();
    assert_eq!(covered_20pct, 75, "Filtered variants cover 75 traces (30+25+20)");

    // Filter by minimum support = 15%
    // Variants with frequency >= 15 traces should remain
    let filtered_15pct = analysis.filter_by_frequency(15.0);
    assert_eq!(filtered_15pct.len(), 4, "4 variants cover ≥15% of cases");
    let covered_15pct: usize = filtered_15pct.iter().map(|v| v.frequency).sum();
    assert_eq!(covered_15pct, 90, "Filtered variants cover 90 traces (30+25+20+15)");

    // Filter by minimum support = 10%
    // All variants should be included
    let filtered_10pct = analysis.filter_by_frequency(10.0);
    assert_eq!(filtered_10pct.len(), 5, "All 5 variants cover ≥10% of cases");
    let covered_10pct: usize = filtered_10pct.iter().map(|v| v.frequency).sum();
    assert_eq!(covered_10pct, 100, "All variants cover all 100 traces");
}

// ============================================================================
// TEST 3: Variant Filtering by Activity
// ============================================================================

#[test]
fn test_variant_filtering_by_activity() {
    // === SETUP: Create log with variants that do/don't contain "account_activated" ===
    let mut log = EventLog::new();
    let base_time = Utc::now();

    // Variant 1: Contains account_activated (6 traces)
    for i in 0..6 {
        let mut trace = Trace::new(format!("trace_with_activated_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("verification_initiated", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("account_activated", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 2: Contains account_activated (4 traces)
    for i in 0..4 {
        let mut trace = Trace::new(format!("trace_alt_path_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("account_activated", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // Variant 3: Does NOT contain account_activated (3 traces)
    for i in 0..3 {
        let mut trace = Trace::new(format!("trace_no_activation_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("verification_failed", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("account_abandoned", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Variant 4: Does NOT contain account_activated (2 traces)
    for i in 0..2 {
        let mut trace = Trace::new(format!("trace_cancelled_{}", i));
        trace.add_event(Event::new("account_created", base_time));
        trace.add_event(Event::new("account_cancelled", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // === EXECUTE: Discover and filter variants ===
    let analysis = discover_variants(&log);
    assert_eq!(analysis.variant_count, 4, "Should discover 4 variants");
    assert_eq!(analysis.total_traces, 15, "Should have 15 total traces");

    // Filter by required activity: must contain "account_activated"
    let with_activation = analysis.filter_by_activity("account_activated");

    // === VERIFY: Results ===
    assert_eq!(with_activation.len(), 2, "Only 2 variants contain account_activated");

    // Total traces with account_activated should be 10 (6 + 4)
    let activated_traces: usize = with_activation.iter().map(|v| v.frequency).sum();
    assert_eq!(activated_traces, 10, "10 traces follow variants with account_activated");

    // Verify coverage percentage
    let coverage = analysis.calculate_coverage_percentage(&with_activation);
    assert!((coverage - 66.67).abs() < 0.1, "Coverage should be ~66.67%");

    // Verify all returned variants contain the activity
    for variant in &with_activation {
        assert!(
            variant.contains_activity("account_activated"),
            "All returned variants must contain account_activated"
        );
    }

    // Test filtering by different activity
    let with_created = analysis.filter_by_activity("account_created");
    assert_eq!(with_created.len(), 4, "All 4 variants contain account_created");

    let with_failed = analysis.filter_by_activity("verification_failed");
    assert_eq!(with_failed.len(), 1, "Only 1 variant contains verification_failed");
    assert_eq!(with_failed[0].frequency, 3);
}

// ============================================================================
// TEST 4: Variant Coverage Percentage (80/20 Rule)
// ============================================================================

#[test]
fn test_variant_coverage_percentage() {
    // === SETUP: Create log where top 20% of variants cover 80% of cases ===
    // This is a classic 80/20 (Pareto) distribution
    let mut log = EventLog::new();
    let base_time = Utc::now();

    // Variant 1: 40 traces (40% - part of the 80%)
    for i in 0..40 {
        let mut trace = Trace::new(format!("trace_v1_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // Variant 2: 30 traces (30% - part of the 80%)
    for i in 0..30 {
        let mut trace = Trace::new(format!("trace_v2_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // Variant 3: 10 traces (10% - part of remaining 20%)
    for i in 0..10 {
        let mut trace = Trace::new(format!("trace_v3_{}", i));
        trace.add_event(Event::new("B", base_time));
        trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // Variant 4: 10 traces (10% - part of remaining 20%)
    for i in 0..10 {
        let mut trace = Trace::new(format!("trace_v4_{}", i));
        trace.add_event(Event::new("C", base_time));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // Variant 5: 10 traces (10% - part of remaining 20%)
    for i in 0..10 {
        let mut trace = Trace::new(format!("trace_v5_{}", i));
        trace.add_event(Event::new("B", base_time));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(1)));
        log.add_trace(trace);
    }

    // === EXECUTE: Discover variants and calculate coverage ===
    let analysis = discover_variants(&log);
    assert_eq!(analysis.variant_count, 5, "Should discover 5 variants");
    assert_eq!(analysis.total_traces, 100, "Should have 100 total traces");

    // Calculate coverage for top variant (40 traces)
    let top_variant_coverage = analysis.variants[0].coverage_percentage(100);
    assert_eq!(top_variant_coverage, 40.0, "Top variant covers 40%");

    // Calculate coverage for top 2 variants (40 + 30 = 70 traces)
    let top_2 = &analysis.variants[0..2];
    let top_2_coverage = analysis.calculate_coverage_percentage(top_2);
    assert_eq!(top_2_coverage, 70.0, "Top 2 variants cover 70%");

    // Calculate coverage for top 3 variants (40 + 30 + 10 = 80 traces)
    let top_3 = &analysis.variants[0..3];
    let top_3_coverage = analysis.calculate_coverage_percentage(top_3);
    assert_eq!(top_3_coverage, 80.0, "Top 3 variants cover 80%");

    // Calculate coverage for all variants (100 traces)
    let all_coverage = analysis.calculate_coverage_percentage(&analysis.variants);
    assert_eq!(all_coverage, 100.0, "All variants cover 100%");

    // Verify individual coverage percentages
    assert_eq!(analysis.variants[0].coverage_percentage(100), 40.0);
    assert_eq!(analysis.variants[1].coverage_percentage(100), 30.0);
    assert_eq!(analysis.variants[2].coverage_percentage(100), 10.0);
    assert_eq!(analysis.variants[3].coverage_percentage(100), 10.0);
    assert_eq!(analysis.variants[4].coverage_percentage(100), 10.0);

    // Verify cumulative coverage
    let cumulative_coverage: f64 = analysis
        .variants
        .iter()
        .scan(0.0, |acc, v| {
            *acc += v.coverage_percentage(100);
            Some(*acc)
        })
        .collect::<Vec<_>>()
        .last()
        .copied()
        .unwrap_or(0.0);
    assert_eq!(cumulative_coverage, 100.0, "Cumulative coverage should be 100%");
}

// ============================================================================
// TEST 5: Variant Anomaly Detection
// ============================================================================

#[test]
fn test_variant_anomaly_detection() {
    // === SETUP: Create log with 100 traces: 95 follow normal variant, 5 follow rare variants ===
    let mut log = EventLog::new();
    let base_time = Utc::now();

    // Normal variant: A→B→C (95 traces, 95%)
    for i in 0..95 {
        let mut trace = Trace::new(format!("trace_normal_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Anomaly 1: A→C→B (2 traces, 2%)
    for i in 0..2 {
        let mut trace = Trace::new(format!("trace_anomaly_1_{}", i));
        trace.add_event(Event::new("A", base_time));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Anomaly 2: B→A→C (2 traces, 2%)
    for i in 0..2 {
        let mut trace = Trace::new(format!("trace_anomaly_2_{}", i));
        trace.add_event(Event::new("B", base_time));
        trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(1)));
        trace.add_event(Event::new("C", base_time + chrono::Duration::minutes(2)));
        log.add_trace(trace);
    }

    // Anomaly 3: C→A→B (1 trace, 1%)
    let mut trace = Trace::new("trace_anomaly_3_0".to_string());
    trace.add_event(Event::new("C", base_time));
    trace.add_event(Event::new("A", base_time + chrono::Duration::minutes(1)));
    trace.add_event(Event::new("B", base_time + chrono::Duration::minutes(2)));
    log.add_trace(trace);

    // === EXECUTE: Discover variants and detect anomalies ===
    let analysis = discover_variants(&log);
    assert_eq!(analysis.variant_count, 4, "Should discover 4 variants");
    assert_eq!(analysis.total_traces, 100, "Should have 100 total traces");

    // Detect anomalies: variants with less than 5% frequency
    let anomalies_5pct = analysis.detect_anomalies(5.0);
    assert_eq!(anomalies_5pct.len(), 3, "Should detect 3 anomalies (< 5%)");
    let anomaly_traces: usize = anomalies_5pct.iter().map(|v| v.frequency).sum();
    assert_eq!(anomaly_traces, 5, "5 traces are anomalies");

    // Detect anomalies: variants with less than 3% frequency
    let anomalies_3pct = analysis.detect_anomalies(3.0);
    assert_eq!(anomalies_3pct.len(), 3, "Should detect 3 anomalies (< 3%)");
    // The three rare variants are 2%, 2%, and 1%
    let rare_variant_traces: usize = anomalies_3pct.iter().map(|v| v.frequency).sum();
    assert_eq!(rare_variant_traces, 5, "The rare variants total 5 traces");

    // Detect anomalies: variants with less than 1% frequency
    let anomalies_1pct = analysis.detect_anomalies(1.0);
    assert_eq!(anomalies_1pct.len(), 0, "No variants below 1%");

    // Detect anomalies: variants with less than 10% frequency
    let anomalies_10pct = analysis.detect_anomalies(10.0);
    assert_eq!(anomalies_10pct.len(), 3, "3 anomalous variants are below 10% (the rare ones)");

    // Verify anomaly coverage
    let anomaly_coverage_5pct = analysis.calculate_coverage_percentage(&anomalies_5pct);
    assert_eq!(anomaly_coverage_5pct, 5.0, "Anomalies cover 5% at 5% threshold");

    // Verify normal variant is NOT flagged as anomaly
    let normal_variant = &analysis.variants[0];
    assert_eq!(normal_variant.frequency, 95, "Normal variant has 95 occurrences");
    assert_eq!(
        normal_variant.coverage_percentage(100),
        95.0,
        "Normal variant covers 95%"
    );

    // Verify normal variant not in anomalies
    let is_normal_anomaly = anomalies_5pct.iter().any(|a| a.frequency == 95);
    assert!(!is_normal_anomaly, "Normal variant should not be in anomalies list");
}
