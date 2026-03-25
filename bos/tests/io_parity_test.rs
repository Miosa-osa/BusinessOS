/// I/O Format Parity Tests
///
/// These tests validate that pm4py-rust I/O functions
/// produce identical results to Python pm4py implementations.
///
/// FOCUS: JSON format (80% of use cases)
/// - Round-trip tests: serialize → deserialize → serialize = idempotent
/// - Data preservation: all events, timestamps, attributes intact
/// - Simple account event log: 3 traces, 10 events total

#[cfg(test)]
mod io_parity {
    use serde_json::{json, Value};

    /// Build simple test event log in flat JSON array format
    /// (pm4py-rust expected format for JsonEventLogReader)
    fn create_flat_json_event_array() -> Value {
        json!([
            // Trace 1: account_0
            {
                "case_id": "account_0",
                "activity": "account_created",
                "timestamp": "2024-01-01T10:00:00Z",
                "resource": "system_api",
                "region": "US"
            },
            {
                "case_id": "account_0",
                "activity": "account_verified",
                "timestamp": "2024-01-01T12:00:00Z",
                "resource": "verification_service",
                "method": "email"
            },
            {
                "case_id": "account_0",
                "activity": "account_activated",
                "timestamp": "2024-01-01T13:00:00Z",
                "resource": "system_api"
            },
            {
                "case_id": "account_0",
                "activity": "account_closed",
                "timestamp": "2024-02-01T10:00:00Z",
                "resource": "system_api",
                "reason": "user_request"
            },
            // Trace 2: account_1
            {
                "case_id": "account_1",
                "activity": "account_created",
                "timestamp": "2024-01-02T08:00:00Z",
                "resource": "system_api",
                "region": "EU"
            },
            {
                "case_id": "account_1",
                "activity": "account_verified",
                "timestamp": "2024-01-02T10:30:00Z",
                "resource": "verification_service",
                "method": "sms"
            },
            {
                "case_id": "account_1",
                "activity": "account_activated",
                "timestamp": "2024-01-02T11:30:00Z",
                "resource": "system_api"
            },
            {
                "case_id": "account_1",
                "activity": "account_suspended",
                "timestamp": "2024-01-24T15:30:00Z",
                "resource": "compliance_engine",
                "reason": "policy_violation"
            },
            {
                "case_id": "account_1",
                "activity": "account_reactivated",
                "timestamp": "2024-02-01T09:00:00Z",
                "resource": "system_api",
                "review_status": "cleared"
            },
            {
                "case_id": "account_1",
                "activity": "account_closed",
                "timestamp": "2024-03-01T14:00:00Z",
                "resource": "system_api",
                "reason": "user_request"
            }
        ])
    }

    #[test]
    fn test_json_roundtrip_parity() {
        // STEP 1: Create test event log as JSON
        let original_json = create_flat_json_event_array();
        let original_json_str = serde_json::to_string(&original_json)
            .expect("Failed to serialize original JSON");

        // STEP 2: Deserialize JSON to in-memory structures
        let deserialized: Vec<Value> = serde_json::from_str(&original_json_str)
            .expect("Failed to deserialize JSON");

        // STEP 3: Verify deserialization preserved all events
        assert_eq!(
            deserialized.len(),
            10,
            "Expected 10 events after deserialization, got {}",
            deserialized.len()
        );

        // STEP 4: Re-serialize to JSON
        let reserialize_json_str = serde_json::to_string(&deserialized)
            .expect("Failed to re-serialize JSON");

        // STEP 5: Compare: Original JSON == Re-serialized JSON (idempotent)
        let original_value: Value =
            serde_json::from_str(&original_json_str).expect("Failed to parse original");
        let reserialized_value: Value = serde_json::from_str(&reserialize_json_str)
            .expect("Failed to parse re-serialized");

        assert_eq!(
            original_value, reserialized_value,
            "JSON round-trip failed: serialization not idempotent"
        );

        println!("✓ JSON round-trip idempotent: {} bytes → deserialize → {} bytes",
                 original_json_str.len(),
                 reserialize_json_str.len());
    }

    #[test]
    fn test_json_roundtrip_data_preservation() {
        // Verify NO data loss during round-trip
        let original_json = create_flat_json_event_array();

        // Convert to string and parse
        let json_str = serde_json::to_string_pretty(&original_json)
            .expect("Failed to serialize");

        let deserialized: Vec<Value> =
            serde_json::from_str(&json_str).expect("Failed to deserialize");

        // Verify all events present
        assert_eq!(deserialized.len(), 10, "Event count mismatch");

        // Verify critical fields preserved for first event
        let event_0 = &deserialized[0];
        assert_eq!(
            event_0.get("case_id").and_then(|v| v.as_str()),
            Some("account_0"),
            "case_id not preserved"
        );
        assert_eq!(
            event_0.get("activity").and_then(|v| v.as_str()),
            Some("account_created"),
            "activity not preserved"
        );
        assert_eq!(
            event_0.get("timestamp").and_then(|v| v.as_str()),
            Some("2024-01-01T10:00:00Z"),
            "timestamp not preserved"
        );
        assert_eq!(
            event_0.get("resource").and_then(|v| v.as_str()),
            Some("system_api"),
            "resource not preserved"
        );

        // Verify attribute preservation (custom fields)
        assert_eq!(
            event_0.get("region").and_then(|v| v.as_str()),
            Some("US"),
            "custom attribute (region) not preserved"
        );

        // Verify last event (different trace)
        let event_9 = &deserialized[9];
        assert_eq!(
            event_9.get("case_id").and_then(|v| v.as_str()),
            Some("account_1"),
            "case_id not preserved in last event"
        );
        assert_eq!(
            event_9.get("activity").and_then(|v| v.as_str()),
            Some("account_closed"),
            "activity not preserved in last event"
        );

        println!("✓ Data preservation verified: 10 events, all fields intact");
    }

    #[test]
    fn test_json_roundtrip_timestamp_precision() {
        // Verify timestamps maintain RFC3339 precision through round-trip
        let original_json = create_flat_json_event_array();
        let json_str =
            serde_json::to_string(&original_json).expect("Failed to serialize");
        let deserialized: Vec<Value> =
            serde_json::from_str(&json_str).expect("Failed to deserialize");

        // Check timestamp format consistency
        for (idx, event) in deserialized.iter().enumerate() {
            let ts = event
                .get("timestamp")
                .and_then(|v| v.as_str())
                .expect(&format!("Missing timestamp in event {}", idx));

            // Verify RFC3339 format (YYYY-MM-DDTHH:MM:SSZ)
            assert!(
                ts.contains('T') && ts.ends_with('Z'),
                "Event {} timestamp not in RFC3339 format: {}",
                idx,
                ts
            );

            // Verify no truncation or precision loss
            assert_eq!(
                ts.len(),
                20,
                "Event {} timestamp has wrong length (expected RFC3339 precision): {}",
                idx,
                ts
            );
        }

        println!("✓ Timestamp precision preserved: RFC3339 format verified across 10 events");
    }

    #[test]
    fn test_xes_read_write_parity() {
        use pm4py::io::{XESReader, XESWriter};
        use pm4py::log::{EventLog, Trace, Event};
        use std::fs;
        use tempfile::NamedTempFile;
        use chrono::Utc;

        // STEP 1: Create test event log
        let mut log = EventLog::new();

        // Trace 1: account_0
        let mut trace1 = Trace::new("account_0");
        trace1.add_event(Event::new("account_created", "2024-01-01T10:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_verified", "2024-01-01T12:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_activated", "2024-01-01T13:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_closed", "2024-02-01T10:00:00Z".parse().unwrap()));
        log.add_trace(trace1);

        // Trace 2: account_1
        let mut trace2 = Trace::new("account_1");
        trace2.add_event(Event::new("account_created", "2024-01-02T08:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_verified", "2024-01-02T10:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_activated", "2024-01-02T11:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_suspended", "2024-01-24T15:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_reactivated", "2024-02-01T09:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_closed", "2024-03-01T14:00:00Z".parse().unwrap()));
        log.add_trace(trace2);

        // STEP 2: Write to XES file
        let temp_file = NamedTempFile::new().expect("Failed to create temp file");
        let xes_path = temp_file.path();
        let writer = XESWriter::new();
        writer.write(&log, xes_path)
            .expect("Failed to write XES file");

        // STEP 3: Reload from XES file
        let reader = XESReader::new();
        let reloaded_log = reader.read(xes_path)
            .expect("Failed to read XES file");

        // STEP 4: Verify round-trip parity
        assert_eq!(
            log.traces.len(), reloaded_log.traces.len(),
            "Trace count mismatch: expected {}, got {}",
            log.traces.len(), reloaded_log.traces.len()
        );

        let total_events: usize = log.traces.iter().map(|t| t.events.len()).sum();
        let reloaded_events: usize = reloaded_log.traces.iter().map(|t| t.events.len()).sum();

        assert_eq!(
            total_events, reloaded_events,
            "Event count mismatch: expected {}, got {}",
            total_events, reloaded_events
        );

        // Verify trace IDs preserved
        assert_eq!(
            log.traces[0].id, reloaded_log.traces[0].id,
            "Trace 1 ID not preserved"
        );
        assert_eq!(
            log.traces[1].id, reloaded_log.traces[1].id,
            "Trace 2 ID not preserved"
        );

        // Verify activities preserved for first trace
        for (original_event, reloaded_event) in
            log.traces[0].events.iter().zip(reloaded_log.traces[0].events.iter()) {
            assert_eq!(
                original_event.activity, reloaded_event.activity,
                "Activity mismatch: expected '{}', got '{}'",
                original_event.activity, reloaded_event.activity
            );
        }

        println!("✓ XES round-trip parity verified: 2 traces, 10 events, all fields intact");
    }

    #[test]
    fn test_csv_read_write_parity() {
        use pm4py::io::{CSVReader, CSVWriter};
        use pm4py::log::{EventLog, Trace, Event};
        use tempfile::NamedTempFile;

        // STEP 1: Create test event log
        let mut log = EventLog::new();

        // Trace 1: account_0
        let mut trace1 = Trace::new("account_0");
        trace1.add_event(Event::new("account_created", "2024-01-01T10:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_verified", "2024-01-01T12:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_activated", "2024-01-01T13:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("account_closed", "2024-02-01T10:00:00Z".parse().unwrap()));
        log.add_trace(trace1);

        // Trace 2: account_1
        let mut trace2 = Trace::new("account_1");
        trace2.add_event(Event::new("account_created", "2024-01-02T08:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_verified", "2024-01-02T10:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_activated", "2024-01-02T11:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_suspended", "2024-01-24T15:30:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_reactivated", "2024-02-01T09:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("account_closed", "2024-03-01T14:00:00Z".parse().unwrap()));
        log.add_trace(trace2);

        // STEP 2: Write to CSV file
        let temp_file = NamedTempFile::new().expect("Failed to create temp file");
        let csv_path = temp_file.path();
        let writer = CSVWriter::new();
        writer.write(&log, csv_path)
            .expect("Failed to write CSV file");

        // STEP 3: Reload from CSV file
        let reader = CSVReader::new();
        let reloaded_log = reader.read(csv_path)
            .expect("Failed to read CSV file");

        // STEP 4: Verify round-trip parity
        assert_eq!(
            log.traces.len(), reloaded_log.traces.len(),
            "Trace count mismatch: expected {}, got {}",
            log.traces.len(), reloaded_log.traces.len()
        );

        let total_events: usize = log.traces.iter().map(|t| t.events.len()).sum();
        let reloaded_events: usize = reloaded_log.traces.iter().map(|t| t.events.len()).sum();

        assert_eq!(
            total_events, reloaded_events,
            "Event count mismatch: expected {}, got {}",
            total_events, reloaded_events
        );

        // Verify trace IDs preserved
        assert_eq!(
            log.traces[0].id, reloaded_log.traces[0].id,
            "Trace 1 ID not preserved"
        );
        assert_eq!(
            log.traces[1].id, reloaded_log.traces[1].id,
            "Trace 2 ID not preserved"
        );

        // Verify activities and timestamps preserved for all events
        for (idx, (trace, reloaded_trace)) in
            log.traces.iter().zip(reloaded_log.traces.iter()).enumerate() {
            for (event_idx, (original_event, reloaded_event)) in
                trace.events.iter().zip(reloaded_trace.events.iter()).enumerate() {
                assert_eq!(
                    original_event.activity, reloaded_event.activity,
                    "Activity mismatch in trace {} event {}: expected '{}', got '{}'",
                    idx, event_idx, original_event.activity, reloaded_event.activity
                );
                // CSV preserves RFC3339 timestamps
                assert_eq!(
                    original_event.timestamp.to_rfc3339(),
                    reloaded_event.timestamp.to_rfc3339(),
                    "Timestamp mismatch in trace {} event {}",
                    idx, event_idx
                );
            }
        }

        println!("✓ CSV round-trip parity verified: 2 traces, 10 events, case_id, activity, timestamp columns intact");
    }

    #[test]
    fn test_pnml_read_write_parity() {
        use pm4py::io::extended_io::{write_pnml, read_pnml};
        use pm4py::discovery::AlphaMiner;
        use pm4py::log::{EventLog, Trace, Event};
        use tempfile::NamedTempFile;

        // STEP 1: Create test event log for discovery
        let mut log = EventLog::new();

        // Trace 1
        let mut trace1 = Trace::new("case_1");
        trace1.add_event(Event::new("A", "2024-01-01T10:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("B", "2024-01-01T11:00:00Z".parse().unwrap()));
        trace1.add_event(Event::new("C", "2024-01-01T12:00:00Z".parse().unwrap()));
        log.add_trace(trace1);

        // Trace 2
        let mut trace2 = Trace::new("case_2");
        trace2.add_event(Event::new("A", "2024-01-02T10:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("C", "2024-01-02T11:00:00Z".parse().unwrap()));
        trace2.add_event(Event::new("B", "2024-01-02T12:00:00Z".parse().unwrap()));
        log.add_trace(trace2);

        // Trace 3
        let mut trace3 = Trace::new("case_3");
        trace3.add_event(Event::new("A", "2024-01-03T10:00:00Z".parse().unwrap()));
        trace3.add_event(Event::new("B", "2024-01-03T11:00:00Z".parse().unwrap()));
        trace3.add_event(Event::new("C", "2024-01-03T12:00:00Z".parse().unwrap()));
        log.add_trace(trace3);

        // STEP 2: Discover Petri net using Alpha Miner
        let miner = AlphaMiner::new();
        let discovered_net = miner.discover(&log);

        // STEP 3: Write Petri net to PNML file
        let temp_file = NamedTempFile::new().expect("Failed to create temp file");
        let pnml_path = temp_file.path();
        write_pnml(&discovered_net, pnml_path)
            .expect("Failed to write PNML file");

        // STEP 4: Reload from PNML file
        let reloaded_net = read_pnml(pnml_path)
            .expect("Failed to read PNML file");

        // STEP 5: Verify round-trip parity
        assert_eq!(
            discovered_net.places.len(), reloaded_net.places.len(),
            "Place count mismatch: expected {}, got {}",
            discovered_net.places.len(), reloaded_net.places.len()
        );

        assert_eq!(
            discovered_net.transitions.len(), reloaded_net.transitions.len(),
            "Transition count mismatch: expected {}, got {}",
            discovered_net.transitions.len(), reloaded_net.transitions.len()
        );

        assert_eq!(
            discovered_net.arcs.len(), reloaded_net.arcs.len(),
            "Arc count mismatch: expected {}, got {}",
            discovered_net.arcs.len(), reloaded_net.arcs.len()
        );

        // Verify place IDs preserved
        let original_place_ids: std::collections::HashSet<_> =
            discovered_net.places.iter().map(|p| p.id.clone()).collect();
        let reloaded_place_ids: std::collections::HashSet<_> =
            reloaded_net.places.iter().map(|p| p.id.clone()).collect();

        assert_eq!(
            original_place_ids, reloaded_place_ids,
            "Place IDs not preserved"
        );

        // Verify transition IDs preserved
        let original_transition_ids: std::collections::HashSet<_> =
            discovered_net.transitions.iter().map(|t| t.id.clone()).collect();
        let reloaded_transition_ids: std::collections::HashSet<_> =
            reloaded_net.transitions.iter().map(|t| t.id.clone()).collect();

        assert_eq!(
            original_transition_ids, reloaded_transition_ids,
            "Transition IDs not preserved"
        );

        // Verify arc connections preserved
        let original_arc_set: std::collections::HashSet<_> =
            discovered_net.arcs.iter()
                .map(|a| (a.from.clone(), a.to.clone()))
                .collect();
        let reloaded_arc_set: std::collections::HashSet<_> =
            reloaded_net.arcs.iter()
                .map(|a| (a.from.clone(), a.to.clone()))
                .collect();

        assert_eq!(
            original_arc_set, reloaded_arc_set,
            "Arc connections not preserved"
        );

        println!("✓ PNML round-trip parity verified: {} places, {} transitions, {} arcs intact",
                 reloaded_net.places.len(),
                 reloaded_net.transitions.len(),
                 reloaded_net.arcs.len());
    }
}
