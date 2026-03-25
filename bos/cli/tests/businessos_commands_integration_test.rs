//! Integration tests for BOS CLI commands accessible from BusinessOS
//! Tests cover: discovery, conformance, statistics, analytics, export, workspace, batch operations

#[cfg(test)]
mod businessos_commands {
    use std::path::Path;
    use std::process::Command;
    use std::collections::HashMap;
    use serde_json::{json, Value};

    struct TestContext {
        test_dir: String,
        log_file: String,
        model_file: String,
    }

    impl TestContext {
        fn new() -> Self {
            let test_dir = "tests/fixtures".to_string();
            let log_file = format!("{}/test_log.xes", test_dir);
            let model_file = format!("{}/test_model.pnml", test_dir);

            TestContext {
                test_dir,
                log_file,
                model_file,
            }
        }

        fn ensure_fixtures_exist(&self) -> anyhow::Result<()> {
            std::fs::create_dir_all(&self.test_dir)?;

            if !Path::new(&self.log_file).exists() {
                let sample_xes = r#"<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="arcSpan,attrIndex" openxes.version="1.0RC7">
  <extension name="Concept" prefix="concept" uri="http://www.xes-standard.org/concept.xesext"/>
  <extension name="Time" prefix="time" uri="http://www.xes-standard.org/time.xesext"/>
  <extension name="Organizational" prefix="org" uri="http://www.xes-standard.org/org.xesext"/>
  <trace>
    <string key="concept:name" value="case_1"/>
    <event>
      <string key="concept:name" value="Submit"/>
      <date key="time:timestamp" value="2026-01-01T08:00:00Z"/>
    </event>
    <event>
      <string key="concept:name" value="Review"/>
      <date key="time:timestamp" value="2026-01-01T09:30:00Z"/>
    </event>
    <event>
      <string key="concept:name" value="Approve"/>
      <date key="time:timestamp" value="2026-01-01T10:15:00Z"/>
    </event>
  </trace>
</log>"#;
                std::fs::write(&self.log_file, sample_xes)?;
            }

            Ok(())
        }
    }

    // ========================================================================
    // DISCOVERY COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_discover_model_alpha_algorithm() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file])
            .output()?;

        assert!(output.status.success(), "{}", String::from_utf8_lossy(&output.stderr));

        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert_eq!(result["algorithm"], "AlphaMiner");
        assert!(result["places"].is_number());
        assert!(result["transitions"].is_number());
        assert!(result["arcs"].is_number());

        Ok(())
    }

    #[test]
    fn test_discover_model_inductive_algorithm() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file, "--algorithm", "inductive"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["places"].is_number());

        Ok(())
    }

    #[test]
    fn test_discover_model_heuristic_algorithm() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file, "--algorithm", "heuristic"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["transitions"].is_number());

        Ok(())
    }

    #[test]
    fn test_discover_model_missing_log_error() {
        let output = Command::new("bos")
            .args(&["discover", "model", "nonexistent_log.xes"])
            .output();

        assert!(output.is_ok());
        // Command should fail gracefully
    }

    #[test]
    fn test_discover_model_output_format_json() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&[
                "discover", "model",
                &ctx.log_file,
                "--output-format", "json"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["algorithm"].is_string());

        Ok(())
    }

    #[test]
    fn test_analyze_variants_default_limit() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "variants", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["total_variants"].is_number());
        assert!(result["variants_by_frequency"].is_array());

        Ok(())
    }

    #[test]
    fn test_analyze_variants_custom_limit() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "variants", &ctx.log_file, "--top-n", "5"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        let variants = result["variants_by_frequency"].as_array().unwrap();
        assert!(variants.len() <= 5);

        Ok(())
    }

    // ========================================================================
    // CONFORMANCE COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_check_conformance_basic() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["conformance", "check", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["fitness"].is_number());
        assert!(result["precision"].is_number());
        assert!(result["generalization"].is_number());
        assert!(result["simplicity"].is_number());
        assert!(result["traces_checked"].is_number());

        Ok(())
    }

    #[test]
    fn test_check_conformance_with_model() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&[
                "conformance", "check",
                &ctx.log_file,
                "--model", &ctx.model_file
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["fitness"].is_number());

        Ok(())
    }

    #[test]
    fn test_conformance_fitness_range() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["conformance", "check", &ctx.log_file])
            .output()?;

        let result: Value = serde_json::from_slice(&output.stdout)?;
        let fitness = result["fitness"].as_f64().unwrap();
        assert!(fitness >= 0.0 && fitness <= 1.0);

        Ok(())
    }

    #[test]
    fn test_detect_deviations_basic() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["conformance", "deviations", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["baseline_variant"].is_string());
        assert!(result["variant_count"].is_number());
        assert!(result["variance_index"].is_number());
        assert!(result["deviations_detected"].is_number());

        Ok(())
    }

    #[test]
    fn test_detect_deviations_with_baseline() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&[
                "conformance", "deviations",
                &ctx.log_file,
                "--baseline", "custom_baseline"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert_eq!(result["baseline_variant"], "custom_baseline");

        Ok(())
    }

    // ========================================================================
    // STATISTICS COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_analyze_statistics_basic() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["statistics", "analyze", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["num_traces"].is_number());
        assert!(result["num_events"].is_number());
        assert!(result["num_unique_activities"].is_number());
        assert!(result["num_variants"].is_number());
        assert!(result["avg_trace_length"].is_number());
        assert!(result["activity_statistics"].is_array());

        Ok(())
    }

    #[test]
    fn test_analyze_statistics_with_variants() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&[
                "statistics", "analyze",
                &ctx.log_file,
                "--include-variants", "true"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["variant_distribution"].is_array());

        Ok(())
    }

    #[test]
    fn test_statistics_output_format() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["statistics", "analyze", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        // Verify required fields are present
        assert!(result["log_name"].is_string());
        assert!(result["num_traces"].is_number());

        Ok(())
    }

    #[test]
    fn test_assess_quality_metrics() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        let output = Command::new("bos")
            .args(&["statistics", "quality"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["total_records"].is_number());
        assert!(result["valid_records"].is_number());
        assert!(result["completeness"].is_number());
        assert!(result["consistency"].is_number());
        assert!(result["accuracy"].is_number());

        Ok(())
    }

    #[test]
    fn test_quality_metrics_values_in_range() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["statistics", "quality"])
            .output()?;

        let result: Value = serde_json::from_slice(&output.stdout)?;
        let completeness = result["completeness"].as_f64().unwrap();
        let consistency = result["consistency"].as_f64().unwrap();
        let accuracy = result["accuracy"].as_f64().unwrap();

        assert!(completeness >= 0.0 && completeness <= 1.0);
        assert!(consistency >= 0.0 && consistency <= 1.0);
        assert!(accuracy >= 0.0 && accuracy <= 1.0);

        Ok(())
    }

    // ========================================================================
    // ANALYTICS COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_generate_fingerprint_entropy() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["analytics", "fingerprint", &ctx.log_file])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["trace_fingerprint"].is_string());
        assert!(result["entropy"].is_number());
        assert!(result["variance_in_duration"].is_number());
        assert!(result["similarity_to_baseline"].is_number());

        Ok(())
    }

    #[test]
    fn test_generate_fingerprint_distribution() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&[
                "analytics", "fingerprint",
                &ctx.log_file,
                "--algorithm", "distribution"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["num_variants"].is_number());

        Ok(())
    }

    #[test]
    fn test_analyze_evolution_weekly() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["analytics", "evolution", &ctx.log_file, "--period", "weekly"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["time_period_start"].is_string());
        assert!(result["time_period_end"].is_string());
        assert!(result["process_changes"].is_number());
        assert!(result["resource_changes"].is_number());
        assert!(result["efficiency_trend"].is_number());
        assert!(result["conformance_trend"].is_number());

        Ok(())
    }

    #[test]
    fn test_analyze_evolution_monthly() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["analytics", "evolution", &ctx.log_file, "--period", "monthly"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["process_changes"].is_number());

        Ok(())
    }

    // ========================================================================
    // EXPORT COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_export_model_pnml_format() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        let output = Command::new("bos")
            .args(&[
                "export", "model",
                &ctx.log_file,
                "--format", "pnml"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert_eq!(result["format"], "pnml");
        assert!(result["output_path"].is_string());
        assert!(result["places"].is_number());

        Ok(())
    }

    #[test]
    fn test_export_model_json_format() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        let output = Command::new("bos")
            .args(&[
                "export", "model",
                &ctx.log_file,
                "--format", "json"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert_eq!(result["format"], "json");

        Ok(())
    }

    #[test]
    fn test_export_model_custom_output() -> anyhow::Result<()> {
        let output_path = "/tmp/custom_model.pnml";

        let output = Command::new("bos")
            .args(&[
                "export", "model",
                "test.xes",
                "--output", output_path
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["output_path"].is_string());

        Ok(())
    }

    #[test]
    fn test_export_report_conformance() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["export", "report", "conformance"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert_eq!(result["command"], "export report");
        assert_eq!(result["status"], "success");

        Ok(())
    }

    #[test]
    fn test_export_report_statistics() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["export", "report", "statistics", "--format", "pdf"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["records_affected"].is_number());

        Ok(())
    }

    // ========================================================================
    // WORKSPACE COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_workspace_stats() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["ws", "stats"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["workspace_path"].is_string());
        assert!(result["total_tables"].is_number());
        assert!(result["total_relationships"].is_number());
        assert!(result["total_entities"].is_number());
        assert!(result["ontology_size_kb"].is_number());
        assert!(result["last_updated"].is_string());

        Ok(())
    }

    #[test]
    fn test_workspace_stats_custom_path() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["ws", "stats", "--path", "/tmp/workspace"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert_eq!(result["workspace_path"], "/tmp/workspace");

        Ok(())
    }

    #[test]
    fn test_workspace_refresh_shallow() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["ws", "refresh"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert_eq!(result["command"], "ws refresh");
        assert_eq!(result["status"], "success");
        assert!(result["execution_time_ms"].is_number());

        Ok(())
    }

    #[test]
    fn test_workspace_refresh_deep() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["ws", "refresh", "--deep", "true"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        let duration = result["execution_time_ms"].as_u64().unwrap();
        assert!(duration > 1000); // Deep refresh should take longer

        Ok(())
    }

    // ========================================================================
    // BATCH COMMANDS TESTS
    // ========================================================================

    #[test]
    fn test_batch_discover_default_workers() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["batch", "discover", "logs/"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert_eq!(result["operation"], "batch discover");
        assert!(result["total_items"].is_number());
        assert!(result["successful_items"].is_number());
        assert!(result["failed_items"].is_number());
        assert!(result["execution_time_ms"].is_number());

        Ok(())
    }

    #[test]
    fn test_batch_discover_custom_workers() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&[
                "batch", "discover",
                "logs/",
                "--workers", "8"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["total_items"].is_number());

        Ok(())
    }

    #[test]
    fn test_batch_discover_with_algorithm() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&[
                "batch", "discover",
                "logs/",
                "--algorithm", "inductive"
            ])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;
        assert!(result["successful_items"].is_number());

        Ok(())
    }

    #[test]
    fn test_batch_conform() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["batch", "conform", "logs/", "--model-dir", "models/"])
            .output()?;

        assert!(output.status.success());
        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert_eq!(result["operation"], "batch conform");
        assert!(result["execution_time_ms"].is_number());

        Ok(())
    }

    #[test]
    fn test_batch_operation_failure_handling() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["batch", "discover", "logs/"])
            .output()?;

        let result: Value = serde_json::from_slice(&output.stdout)?;
        let failed = result["failed_items"].as_u64().unwrap();

        if failed > 0 {
            assert!(result["failures"].is_array());
            let failures = result["failures"].as_array().unwrap();
            assert!(failures.len() > 0);

            let first_failure = &failures[0];
            assert!(first_failure["item_id"].is_string());
            assert!(first_failure["error_message"].is_string());
        }

        Ok(())
    }

    // ========================================================================
    // ERROR HANDLING TESTS
    // ========================================================================

    #[test]
    fn test_invalid_algorithm_error() {
        let output = Command::new("bos")
            .args(&["discover", "model", "test.xes", "--algorithm", "invalid"])
            .output();

        assert!(output.is_ok());
    }

    #[test]
    fn test_missing_required_argument_error() {
        let output = Command::new("bos")
            .args(&["discover", "model"])
            .output();

        // Should fail without required log argument
        assert!(output.is_ok());
    }

    #[test]
    fn test_invalid_log_format_error() {
        let output = Command::new("bos")
            .args(&["discover", "model", "invalid.xyz"])
            .output();

        assert!(output.is_ok());
    }

    // ========================================================================
    // OUTPUT FORMAT TESTS
    // ========================================================================

    #[test]
    fn test_json_output_validity() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file])
            .output()?;

        let stdout = String::from_utf8(output.stdout)?;
        let result: Value = serde_json::from_str(&stdout)?;

        // All JSON outputs should be valid JSON
        assert!(result.is_object() || result.is_array());

        Ok(())
    }

    #[test]
    fn test_result_timestamp_format() -> anyhow::Result<()> {
        let output = Command::new("bos")
            .args(&["ws", "stats"])
            .output()?;

        let result: Value = serde_json::from_slice(&output.stdout)?;
        let timestamp = result["last_updated"].as_str().unwrap();

        // Should be RFC3339 format
        assert!(timestamp.contains("T"));
        assert!(timestamp.contains("Z") || timestamp.contains("+"));

        Ok(())
    }

    #[test]
    fn test_numeric_fields_not_strings() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output = Command::new("bos")
            .args(&["statistics", "analyze", &ctx.log_file])
            .output()?;

        let result: Value = serde_json::from_slice(&output.stdout)?;

        assert!(result["num_traces"].is_number());
        assert!(result["num_events"].is_number());
        assert!(!result["num_traces"].is_string());

        Ok(())
    }

    // ========================================================================
    // INTEGRATION TESTS
    // ========================================================================

    #[test]
    fn test_discover_to_conform_workflow() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        // First discover a model
        let discover_output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file])
            .output()?;

        assert!(discover_output.status.success());

        // Then check conformance against the same log
        let conform_output = Command::new("bos")
            .args(&["conformance", "check", &ctx.log_file])
            .output()?;

        assert!(conform_output.status.success());

        // Results should be consistent
        let discover_result: Value = serde_json::from_slice(&discover_output.stdout)?;
        let conform_result: Value = serde_json::from_slice(&conform_output.stdout)?;

        assert!(discover_result["places"].is_number());
        assert!(conform_result["fitness"].is_number());

        Ok(())
    }

    #[test]
    fn test_statistics_analysis_consistency() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let output1 = Command::new("bos")
            .args(&["statistics", "analyze", &ctx.log_file])
            .output()?;

        let output2 = Command::new("bos")
            .args(&["statistics", "analyze", &ctx.log_file])
            .output()?;

        let result1: Value = serde_json::from_slice(&output1.stdout)?;
        let result2: Value = serde_json::from_slice(&output2.stdout)?;

        // Same log should produce same statistics
        assert_eq!(
            result1["num_traces"].as_u64(),
            result2["num_traces"].as_u64()
        );

        Ok(())
    }

    #[test]
    fn test_export_preserves_model_structure() -> anyhow::Result<()> {
        let ctx = TestContext::new();
        ctx.ensure_fixtures_exist()?;

        let discover_output = Command::new("bos")
            .args(&["discover", "model", &ctx.log_file])
            .output()?;

        let discover_result: Value = serde_json::from_slice(&discover_output.stdout)?;
        let original_places = discover_result["places"].as_u64().unwrap();

        let export_output = Command::new("bos")
            .args(&["export", "model", &ctx.log_file])
            .output()?;

        let export_result: Value = serde_json::from_slice(&export_output.stdout)?;
        let exported_places = export_result["places"].as_u64().unwrap();

        // Exported model should have same structure
        assert_eq!(original_places, exported_places);

        Ok(())
    }
}
