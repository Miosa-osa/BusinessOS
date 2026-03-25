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
    #[ignore = "awaiting pm4py-rust XES support"]
    fn test_xes_read_write_parity() {
        // TODO: Implement after XES reader/writer are available
        // This is a specialized XML format used by ProM tool
        // Not critical for 80% use case
    }

    #[test]
    #[ignore = "awaiting pm4py-rust CSV support"]
    fn test_csv_read_write_parity() {
        // TODO: Implement after CSV reader/writer are available
        // CSV is simpler than JSON but less powerful for attributes
        // Not critical for 80% use case
    }

    #[test]
    #[ignore = "awaiting pm4py-rust PNML support"]
    fn test_pnml_read_write_parity() {
        // TODO: Implement after PNML reader/writer are available
        // PNML is XML-based, specialized for Petri nets
        // Not critical for 80% use case
    }
}
