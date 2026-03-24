// CLI Integration Breakage Test
// Tests that the bos pm4py commands actually work with real files

#[cfg(test)]
mod cli_pm4py_breakage_tests {
    use std::process::Command;

    fn get_bos_bin() -> String {
        // Run from workspace root
        "./target/debug/bos".to_string()
    }

    #[test]
    fn test_cli_pm4py_load_actually_reads_file() {
        // GIVEN: A real XES file with 5 traces, 15 events
        let xes_path = "/Users/sac/chatmangpt/test_simple.xes";

        // WHEN: We run bos pm4py load
        let output = Command::new(get_bos_bin())
            .args(["pm4py", "load", "--source", xes_path])
            .current_dir("/Users/sac/chatmangpt/BusinessOS/bos")
            .output()
            .expect("Failed to execute bos command");

        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);

        // THEN: It should NOT report 0 traces and 0 events
        // THIS MUST FAIL - proves the CLI integration is broken
        assert!(
            !stdout.contains("\"traces\":0"),
            "CLI reports 0 traces! Output: {}\nStderr: {}",
            stdout,
            stderr
        );

        assert!(
            !stdout.contains("\"events\":0"),
            "CLI reports 0 events! Output: {}\nStderr: {}",
            stdout,
            stderr
        );

        // Should report actual numbers
        assert!(
            stdout.contains("\"traces\":5"),
            "CLI should report 5 traces, but output was: {}",
            stdout
        );

        assert!(
            stdout.contains("\"events\":15"),
            "CLI should report 15 events, but output was: {}",
            stdout
        );
    }

    #[test]
    fn test_cli_pm4py_analyze_actually_works() {
        // GIVEN: A real XES file
        let xes_path = "/Users/sac/chatmangpt/test_simple.xes";

        // WHEN: We run bos pm4py analyze
        let output = Command::new(get_bos_bin())
            .args(["pm4py", "analyze", "--source", xes_path])
            .current_dir("/Users/sac/chatmangpt/BusinessOS/bos")
            .output()
            .expect("Failed to execute bos command");

        let stdout = String::from_utf8_lossy(&output.stdout);

        // THEN: Should report actual statistics, not zeros
        assert!(
            stdout.contains("\"traces\":5"),
            "analyze should report 5 traces, got: {}",
            stdout
        );

        assert!(
            stdout.contains("\"total_events\":15"),
            "analyze should report 15 events, got: {}",
            stdout
        );

        assert!(
            stdout.contains("\"unique_activities\":3"),
            "analyze should report 3 activities, got: {}",
            stdout
        );
    }

    #[test]
    fn test_cli_pm4py_discover_actually_works() {
        // GIVEN: A real XES file with A->B->C process
        let xes_path = "/Users/sac/chatmangpt/test_simple.xes";

        // WHEN: We run bos pm4py discover
        let output = Command::new(get_bos_bin())
            .args(["pm4py", "discover", "--source", xes_path])
            .current_dir("/Users/sac/chatmangpt/BusinessOS/bos")
            .output()
            .expect("Failed to execute bos command");

        let stdout = String::from_utf8_lossy(&output.stdout);

        // THEN: Should discover a model with 3 transitions (A, B, C)
        // NOT an empty model with 0 transitions
        assert!(
            stdout.contains("\"transitions\":3"),
            "discover should find 3 transitions (A, B, C), got: {}",
            stdout
        );

        assert!(
            stdout.contains("\"places\":4"),
            "discover should find 4 places for A->B->C, got: {}",
            stdout
        );
    }
}
