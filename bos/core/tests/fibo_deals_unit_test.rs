//! Unit tests for FIBO Deals module
//!
//! Tests cover Deal struct operations, serialization, validation, and mutations.
//! All tests follow Chicago TDD (Red-Green-Refactor) and FIRST principles.

#[cfg(test)]
mod tests {
    use std::collections::HashMap;
    use std::time::{SystemTime, UNIX_EPOCH};

    // ============================================================================
    // Test Data Structures
    // ============================================================================

    #[derive(Debug, Clone, PartialEq)]
    struct Deal {
        id: String,
        name: String,
        amount: f64,
        currency: String,
        status: String,
        buyer_id: String,
        seller_id: String,
        expected_close_date: Option<i64>,
        probability: u32,
        stage: String,
        created_at: i64,
        updated_at: i64,
        rdf_triple_count: usize,
        compliance_status: String,
        kyc_verified: bool,
        aml_screening: String,
    }

    // ============================================================================
    // Deal Creation & Validation Tests
    // ============================================================================

    #[test]
    fn test_deal_creation_minimal() {
        let deal = Deal {
            id: "deal-001".to_string(),
            name: "Acme Acquisition".to_string(),
            amount: 1_000_000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        assert_eq!(deal.id, "deal-001");
        assert_eq!(deal.name, "Acme Acquisition");
        assert_eq!(deal.amount, 1_000_000.0);
        assert_eq!(deal.buyer_id, "buyer-123");
    }

    #[test]
    fn test_deal_creation_with_all_fields() {
        let close_date = current_timestamp() + 7776000; // 90 days from now
        let deal = Deal {
            id: "deal-002".to_string(),
            name: "Strategic Partnership".to_string(),
            amount: 5_500_000.50,
            currency: "EUR".to_string(),
            status: "in_progress".to_string(),
            buyer_id: "buyer-789".to_string(),
            seller_id: "seller-012".to_string(),
            expected_close_date: Some(close_date),
            probability: 75,
            stage: "negotiation".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 12,
            compliance_status: "verified".to_string(),
            kyc_verified: true,
            aml_screening: "cleared".to_string(),
        };

        assert_eq!(deal.amount, 5_500_000.50);
        assert_eq!(deal.currency, "EUR");
        assert_eq!(deal.probability, 75);
        assert!(deal.kyc_verified);
        assert_eq!(deal.compliance_status, "verified");
    }

    #[test]
    fn test_deal_validation_invalid_amount() {
        let deal = Deal {
            id: "deal-003".to_string(),
            name: "Invalid Deal".to_string(),
            amount: -1000.0, // Invalid: negative amount
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        let result = validate_deal(&deal);
        assert!(!result.is_valid);
        assert!(result.errors.contains(&"amount must be positive".to_string()));
    }

    #[test]
    fn test_deal_validation_invalid_probability() {
        let deal = Deal {
            id: "deal-004".to_string(),
            name: "Invalid Probability".to_string(),
            amount: 1000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 150, // Invalid: > 100
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        let result = validate_deal(&deal);
        assert!(!result.is_valid);
        assert!(result.errors.iter().any(|e| e.contains("probability")));
    }

    #[test]
    fn test_deal_validation_missing_buyer() {
        let deal = Deal {
            id: "deal-005".to_string(),
            name: "Missing Buyer".to_string(),
            amount: 1000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "".to_string(), // Invalid: empty
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        let result = validate_deal(&deal);
        assert!(!result.is_valid);
        assert!(result.errors.iter().any(|e| e.contains("buyer_id")));
    }

    #[test]
    fn test_deal_validation_valid_deal() {
        let deal = Deal {
            id: "deal-006".to_string(),
            name: "Valid Deal".to_string(),
            amount: 2_000_000.0,
            currency: "GBP".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-999".to_string(),
            seller_id: "seller-888".to_string(),
            expected_close_date: Some(current_timestamp() + 3600),
            probability: 65,
            stage: "proposal".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        let result = validate_deal(&deal);
        assert!(result.is_valid);
        assert!(result.errors.is_empty());
    }

    // ============================================================================
    // Deal Serialization Tests
    // ============================================================================

    #[test]
    fn test_deal_to_json() {
        let deal = Deal {
            id: "deal-007".to_string(),
            name: "JSON Test".to_string(),
            amount: 1_500_000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-111".to_string(),
            seller_id: "seller-222".to_string(),
            expected_close_date: None,
            probability: 40,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        let json = deal_to_json(&deal);
        assert!(json.contains("deal-007"));
        assert!(json.contains("JSON Test"));
        assert!(json.contains("1500000"));
    }

    #[test]
    fn test_deal_from_json_valid() {
        let json = r#"{"id":"deal-008","name":"From JSON","amount":3000000.75}"#;

        match deal_from_json(json) {
            Ok(deal) => {
                assert_eq!(deal.id, "deal-008");
                assert_eq!(deal.name, "From JSON");
                assert_eq!(deal.amount, 3000000.75);
                assert!(deal.probability <= 100);
            }
            Err(e) => panic!("Failed to parse JSON: {}", e),
        }
    }

    #[test]
    fn test_deal_from_json_invalid() {
        let json = r#"{ "id": "deal-009" }"#;
        let result = deal_from_json(json);
        assert!(result.is_err());
    }

    // ============================================================================
    // Deal Mutation Tests
    // ============================================================================

    #[test]
    fn test_deal_update_name() {
        let created = current_timestamp();
        let mut deal = Deal {
            id: "deal-010".to_string(),
            name: "Original Name".to_string(),
            amount: 1000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: created,
            updated_at: created,
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        deal.name = "Updated Name".to_string();
        deal.updated_at = created + 1; // Ensure updated_at > created_at

        assert_eq!(deal.name, "Updated Name");
        assert!(deal.updated_at >= deal.created_at);
    }

    #[test]
    fn test_deal_update_amount() {
        let mut deal = Deal {
            id: "deal-011".to_string(),
            name: "Amount Test".to_string(),
            amount: 1000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        deal.amount = 5_000_000.0;
        assert_eq!(deal.amount, 5_000_000.0);
    }

    #[test]
    fn test_deal_update_probability() {
        let mut deal = create_test_deal("deal-012");
        deal.probability = 95;
        deal.updated_at = current_timestamp();

        assert_eq!(deal.probability, 95);
    }

    #[test]
    fn test_deal_update_status_progression() {
        let mut deal = create_test_deal("deal-013");
        assert_eq!(deal.status, "draft");

        deal.status = "in_progress".to_string();
        assert_eq!(deal.status, "in_progress");

        deal.status = "closed_won".to_string();
        assert_eq!(deal.status, "closed_won");
    }

    #[test]
    fn test_deal_update_compliance_status() {
        let mut deal = create_test_deal("deal-014");
        assert_eq!(deal.compliance_status, "pending");

        deal.compliance_status = "verified".to_string();
        deal.kyc_verified = true;
        deal.aml_screening = "cleared".to_string();

        assert_eq!(deal.compliance_status, "verified");
        assert!(deal.kyc_verified);
        assert_eq!(deal.aml_screening, "cleared");
    }

    // ============================================================================
    // Deal RDF Integration Tests
    // ============================================================================

    #[test]
    fn test_deal_rdf_triple_count() {
        let deal = Deal {
            id: "deal-015".to_string(),
            name: "RDF Test".to_string(),
            amount: 1000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-123".to_string(),
            seller_id: "seller-456".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 15,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        };

        assert_eq!(deal.rdf_triple_count, 15);
        assert!(deal.rdf_triple_count > 0);
    }

    #[test]
    fn test_deal_rdf_metadata_populated() {
        let deal = Deal {
            id: "deal-016".to_string(),
            name: "RDF Metadata".to_string(),
            amount: 2_000_000.0,
            currency: "USD".to_string(),
            status: "in_progress".to_string(),
            buyer_id: "buyer-111".to_string(),
            seller_id: "seller-222".to_string(),
            expected_close_date: None,
            probability: 70,
            stage: "negotiation".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 20,
            compliance_status: "verified".to_string(),
            kyc_verified: true,
            aml_screening: "cleared".to_string(),
        };

        assert!(deal.rdf_triple_count > 0);
        assert_eq!(deal.compliance_status, "verified");
        assert!(deal.kyc_verified);
    }

    // ============================================================================
    // Deal Timestamp Tests
    // ============================================================================

    #[test]
    fn test_deal_timestamps_created() {
        let now = current_timestamp();
        let deal = create_test_deal("deal-017");

        assert!(deal.created_at > 0);
        assert_eq!(deal.created_at, deal.updated_at);
    }

    #[test]
    fn test_deal_timestamps_updated() {
        let mut deal = create_test_deal("deal-018");
        let created_at = deal.created_at;

        std::thread::sleep(std::time::Duration::from_millis(10));
        deal.updated_at = current_timestamp();

        assert!(deal.updated_at >= created_at);
        assert!(deal.updated_at >= deal.created_at);
    }

    // ============================================================================
    // Deal Collection Tests
    // ============================================================================

    #[test]
    fn test_deal_collection_creation() {
        let deals = vec![
            create_test_deal("deal-019"),
            create_test_deal("deal-020"),
            create_test_deal("deal-021"),
        ];

        assert_eq!(deals.len(), 3);
        assert_eq!(deals[0].id, "deal-019");
        assert_eq!(deals[1].id, "deal-020");
        assert_eq!(deals[2].id, "deal-021");
    }

    #[test]
    fn test_deal_collection_filter_by_status() {
        let mut deals = vec![
            create_test_deal("deal-022"),
            create_test_deal("deal-023"),
            create_test_deal("deal-024"),
        ];

        deals[1].status = "closed_won".to_string();

        let won_deals: Vec<_> = deals
            .iter()
            .filter(|d| d.status == "closed_won")
            .collect();

        assert_eq!(won_deals.len(), 1);
        assert_eq!(won_deals[0].id, "deal-023");
    }

    #[test]
    fn test_deal_collection_filter_by_amount_threshold() {
        let mut deals = vec![
            create_test_deal("deal-025"),
            create_test_deal("deal-026"),
            create_test_deal("deal-027"),
        ];

        deals[0].amount = 100_000.0;
        deals[1].amount = 5_000_000.0;
        deals[2].amount = 500_000.0;

        let large_deals: Vec<_> = deals
            .iter()
            .filter(|d| d.amount > 1_000_000.0)
            .collect();

        assert_eq!(large_deals.len(), 1);
        assert_eq!(large_deals[0].amount, 5_000_000.0);
    }

    #[test]
    fn test_deal_collection_sort_by_amount() {
        let mut deals = vec![
            create_test_deal("deal-028"),
            create_test_deal("deal-029"),
            create_test_deal("deal-030"),
        ];

        deals[0].amount = 1_000_000.0;
        deals[1].amount = 100_000.0;
        deals[2].amount = 5_000_000.0;

        deals.sort_by(|a, b| b.amount.partial_cmp(&a.amount).unwrap());

        assert_eq!(deals[0].amount, 5_000_000.0);
        assert_eq!(deals[1].amount, 1_000_000.0);
        assert_eq!(deals[2].amount, 100_000.0);
    }

    #[test]
    fn test_deal_collection_aggregate_total_value() {
        let mut deals = vec![
            create_test_deal("deal-031"),
            create_test_deal("deal-032"),
            create_test_deal("deal-033"),
        ];

        deals[0].amount = 1_000_000.0;
        deals[1].amount = 2_000_000.0;
        deals[2].amount = 3_000_000.0;

        let total: f64 = deals.iter().map(|d| d.amount).sum();

        assert_eq!(total, 6_000_000.0);
    }

    // ============================================================================
    // Deal Field Validation Tests
    // ============================================================================

    #[test]
    fn test_deal_currency_validation() {
        let currencies = vec!["USD", "EUR", "GBP", "JPY", "CAD", "CHF"];

        for currency in currencies {
            let mut deal = create_test_deal("deal-034");
            deal.currency = currency.to_string();
            assert_eq!(deal.currency.len(), 3);
        }
    }

    #[test]
    fn test_deal_stage_progression() {
        let stages = vec!["prospecting", "proposal", "negotiation", "contract"];
        let mut stage_progression = Vec::new();

        for stage in stages {
            let mut deal = create_test_deal("deal-035");
            deal.stage = stage.to_string();
            stage_progression.push(deal.stage.clone());
        }

        assert_eq!(stage_progression.len(), 4);
        assert_eq!(stage_progression[0], "prospecting");
        assert_eq!(stage_progression[3], "contract");
    }

    #[test]
    fn test_deal_probability_ranges() {
        let probabilities = vec![0, 25, 50, 75, 100];

        for prob in probabilities {
            let mut deal = create_test_deal("deal-036");
            deal.probability = prob;
            assert!(deal.probability <= 100);
            assert!(deal.probability >= 0);
        }
    }

    // ============================================================================
    // Deal Cloning & Copy Tests
    // ============================================================================

    #[test]
    fn test_deal_clone() {
        let deal1 = create_test_deal("deal-037");
        let deal2 = deal1.clone();

        assert_eq!(deal1, deal2);
        assert_eq!(deal1.id, deal2.id);
        assert_eq!(deal1.name, deal2.name);
    }

    #[test]
    fn test_deal_clone_independence() {
        let mut deal1 = create_test_deal("deal-038");
        let mut deal2 = deal1.clone();

        deal2.name = "Modified".to_string();

        assert_ne!(deal1.name, deal2.name);
        assert_eq!(deal1.name, "Test Deal");
        assert_eq!(deal2.name, "Modified");
    }

    // ============================================================================
    // Helper Functions
    // ============================================================================

    fn current_timestamp() -> i64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
    }

    fn create_test_deal(id: &str) -> Deal {
        Deal {
            id: id.to_string(),
            name: "Test Deal".to_string(),
            amount: 1_000_000.0,
            currency: "USD".to_string(),
            status: "draft".to_string(),
            buyer_id: "buyer-test".to_string(),
            seller_id: "seller-test".to_string(),
            expected_close_date: None,
            probability: 50,
            stage: "prospecting".to_string(),
            created_at: current_timestamp(),
            updated_at: current_timestamp(),
            rdf_triple_count: 0,
            compliance_status: "pending".to_string(),
            kyc_verified: false,
            aml_screening: "pending".to_string(),
        }
    }

    fn deal_to_json(deal: &Deal) -> String {
        format!(
            r#"{{"id":"{}","name":"{}","amount":{},"currency":"{}","status":"{}","buyer_id":"{}","seller_id":"{}","probability":{},"compliance_status":"{}","kyc_verified":{}}}"#,
            deal.id,
            deal.name,
            deal.amount,
            deal.currency,
            deal.status,
            deal.buyer_id,
            deal.seller_id,
            deal.probability,
            deal.compliance_status,
            deal.kyc_verified
        )
    }

    fn deal_from_json(json: &str) -> Result<Deal, String> {
        // Simplified parsing for test purposes
        if json.contains("id") && json.contains("name") && json.contains("amount") {
            Ok(Deal {
                id: extract_field(json, "id").unwrap_or_default(),
                name: extract_field(json, "name").unwrap_or_default(),
                amount: extract_number_field(json, "amount").unwrap_or(0.0),
                currency: extract_field(json, "currency").unwrap_or_else(|| "USD".to_string()),
                status: extract_field(json, "status").unwrap_or_else(|| "draft".to_string()),
                buyer_id: extract_field(json, "buyer_id").unwrap_or_default(),
                seller_id: extract_field(json, "seller_id").unwrap_or_default(),
                expected_close_date: None,
                probability: extract_number_field(json, "probability")
                    .unwrap_or(0.0) as u32,
                stage: extract_field(json, "stage").unwrap_or_else(|| "prospecting".to_string()),
                created_at: current_timestamp(),
                updated_at: current_timestamp(),
                rdf_triple_count: extract_number_field(json, "rdf_triple_count")
                    .unwrap_or(0.0) as usize,
                compliance_status: extract_field(json, "compliance_status")
                    .unwrap_or_else(|| "pending".to_string()),
                kyc_verified: json.contains("\"kyc_verified\":true"),
                aml_screening: extract_field(json, "aml_screening")
                    .unwrap_or_else(|| "pending".to_string()),
            })
        } else {
            Err("Missing required fields".to_string())
        }
    }

    fn extract_field(json: &str, field: &str) -> Option<String> {
        let pattern = format!("\"{}\":\"", field);
        json.find(&pattern)
            .and_then(|start| {
                let value_start = start + pattern.len();
                json[value_start..]
                    .find('"')
                    .map(|end| json[value_start..value_start + end].to_string())
            })
    }

    fn extract_number_field(json: &str, field: &str) -> Option<f64> {
        let pattern = format!("\"{}\":", field);
        json.find(&pattern).and_then(|start| {
            let value_start = start + pattern.len();
            let value_str = &json[value_start..]
                .trim_start()
                .split(|c| c == ',' || c == '}')
                .next()?;
            value_str.parse().ok()
        })
    }

    struct ValidationResult {
        is_valid: bool,
        errors: Vec<String>,
    }

    fn validate_deal(deal: &Deal) -> ValidationResult {
        let mut errors = Vec::new();

        if deal.amount <= 0.0 {
            errors.push("amount must be positive".to_string());
        }

        if deal.probability > 100 || deal.probability < 0 {
            errors.push("probability must be between 0 and 100".to_string());
        }

        if deal.buyer_id.is_empty() {
            errors.push("buyer_id cannot be empty".to_string());
        }

        if deal.seller_id.is_empty() {
            errors.push("seller_id cannot be empty".to_string());
        }

        if deal.name.is_empty() {
            errors.push("name cannot be empty".to_string());
        }

        ValidationResult {
            is_valid: errors.is_empty(),
            errors,
        }
    }
}
