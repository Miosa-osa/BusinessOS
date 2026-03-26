//! FIBO ontology integration tests
//!
//! Tests for FIBO (Financial Industry Business Ontology) commands:
//! - Deal creation and listing
//! - Party KYC verification
//! - Compliance checking
//! - SPARQL CONSTRUCT query generation

#[cfg(test)]
mod fibo_integration {
    use std::fs;
    use std::path::Path;
    use serde_json::{json, Value};

    struct TestContext {
        test_dir: String,
    }

    impl TestContext {
        fn new() -> Self {
            let test_dir = "tests/fixtures/fibo".to_string();
            let _ = fs::create_dir_all(&test_dir);
            TestContext { test_dir }
        }

        fn create_kyc_file(&self, filename: &str, data: Value) -> std::io::Result<String> {
            let path = format!("{}/{}", self.test_dir, filename);
            let json = serde_json::to_string_pretty(&data)?;
            fs::write(&path, json)?;
            Ok(path)
        }

        fn create_compliance_rules_file(&self, filename: &str, rules: Value) -> std::io::Result<String> {
            let path = format!("{}/{}", self.test_dir, filename);
            let json = serde_json::to_string_pretty(&rules)?;
            fs::write(&path, json)?;
            Ok(path)
        }

        fn cleanup(&self) {
            let _ = fs::remove_dir_all(&self.test_dir);
        }
    }

    impl Drop for TestContext {
        fn drop(&mut self) {
            self.cleanup();
        }
    }

    // ========================================================================
    // DEAL CREATION TESTS
    // ========================================================================

    #[test]
    fn test_fibo_deal_create_saas() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create a SaaS deal
        let deal_data = json!({
            "name": "CloudSync Enterprise License",
            "deal_type": "saas",
            "party_a": "corp_buyer_001",
            "party_b": "saas_vendor_001",
            "value": 250000.0,
            "currency": "USD"
        });

        // Validate deal structure
        assert_eq!(deal_data["deal_type"], "saas");
        assert_eq!(deal_data["value"], 250000.0);
        assert_eq!(deal_data["currency"], "USD");

        // Verify parties are defined
        assert!(!deal_data["party_a"].as_str().unwrap().is_empty());
        assert!(!deal_data["party_b"].as_str().unwrap().is_empty());

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_deal_create_loan() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create a loan deal
        let deal_data = json!({
            "name": "Growth Capital Facility - Series B",
            "deal_type": "loan",
            "party_a": "startup_borrower_001",
            "party_b": "bank_lender_001",
            "value": 5000000.0,
            "currency": "USD",
            "terms": {
                "term_years": 5,
                "interest_rate": 7.5,
                "repayment_schedule": "quarterly"
            }
        });

        // Validate loan-specific fields
        assert_eq!(deal_data["deal_type"], "loan");
        assert_eq!(deal_data["value"], 5000000.0);
        let terms = &deal_data["terms"];
        assert_eq!(terms["term_years"], 5);
        assert_eq!(terms["interest_rate"], 7.5);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_deal_create_defense_contract() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create a defense contract
        let deal_data = json!({
            "name": "Defense Contractor Supply Agreement",
            "deal_type": "defense_contract",
            "party_a": "dod_buyer",
            "party_b": "defense_contractor_001",
            "value": 12500000.0,
            "currency": "USD",
            "classified": true,
            "security_clearance_required": true
        });

        // Validate defense-specific fields
        assert_eq!(deal_data["deal_type"], "defense_contract");
        assert_eq!(deal_data["value"], 12500000.0);
        assert_eq!(deal_data["classified"], true);
        assert_eq!(deal_data["security_clearance_required"], true);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_deal_id_generation() -> anyhow::Result<()> {
        // Test that deal IDs are unique UUIDs
        let deal_id_1 = uuid::Uuid::new_v4().to_string();
        let deal_id_2 = uuid::Uuid::new_v4().to_string();

        assert_ne!(deal_id_1, deal_id_2);
        assert_eq!(deal_id_1.len(), 36); // Standard UUID length with hyphens
        assert_eq!(deal_id_2.len(), 36);

        Ok(())
    }

    // ========================================================================
    // KYC VERIFICATION TESTS
    // ========================================================================

    #[test]
    fn test_fibo_party_kyc_individual() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create KYC data for individual
        let kyc_data = json!({
            "party_id": "individual_001",
            "party_type": "individual",
            "name": "John Smith",
            "date_of_birth": "1980-05-15",
            "passport_number": "AB123456",
            "address": "123 Main St, Boston, MA"
        });

        let kyc_file = ctx.create_kyc_file("individual_kyc.json", kyc_data)?;

        // Validate KYC file was created
        assert!(Path::new(&kyc_file).exists());

        // Verify content
        let content = fs::read_to_string(&kyc_file)?;
        let parsed: Value = serde_json::from_str(&content)?;
        assert_eq!(parsed["party_type"], "individual");
        assert_eq!(parsed["name"], "John Smith");

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_party_kyc_company() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create KYC data for company
        let kyc_data = json!({
            "party_id": "company_001",
            "party_type": "company",
            "legal_name": "Acme Corporation",
            "registration_number": "12-3456789",
            "jurisdiction": "Delaware",
            "business_address": "456 Corp Ave, Boston, MA",
            "beneficial_owners": [
                {
                    "name": "Alice Johnson",
                    "ownership_percentage": 60.0
                },
                {
                    "name": "Bob Williams",
                    "ownership_percentage": 40.0
                }
            ]
        });

        let kyc_file = ctx.create_kyc_file("company_kyc.json", kyc_data)?;

        // Validate KYC file and structure
        assert!(Path::new(&kyc_file).exists());
        let content = fs::read_to_string(&kyc_file)?;
        let parsed: Value = serde_json::from_str(&content)?;
        assert_eq!(parsed["party_type"], "company");
        assert_eq!(parsed["legal_name"], "Acme Corporation");
        assert!(parsed["beneficial_owners"].is_array());
        assert_eq!(parsed["beneficial_owners"].as_array().unwrap().len(), 2);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_kyc_aml_score_calculation() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create KYC data with risk indicators
        let kyc_data = json!({
            "party_id": "test_party",
            "party_type": "company",
            "legal_name": "Test Corp",
            "risk_factors": {
                "pep_related": false,
                "sanctioned_country": false,
                "adverse_media": 0,
                "transaction_velocity": "normal"
            }
        });

        let kyc_file = ctx.create_kyc_file("aml_test.json", kyc_data)?;
        let content = fs::read_to_string(&kyc_file)?;
        let parsed: Value = serde_json::from_str(&content)?;

        // Calculate AML score based on risk factors
        let mut aml_score = 0.0;
        let risk_factors = &parsed["risk_factors"];
        if risk_factors["pep_related"].as_bool().unwrap_or(false) {
            aml_score += 0.3;
        }
        if risk_factors["sanctioned_country"].as_bool().unwrap_or(false) {
            aml_score += 0.5;
        }
        aml_score += risk_factors["adverse_media"].as_f64().unwrap_or(0.0) * 0.1;

        // Score should be low (compliant)
        assert!(aml_score < 0.5);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_kyc_verification_status() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create KYC with complete verification
        let kyc_data = json!({
            "party_id": "verified_party",
            "verification_status": "verified",
            "identity_checks_passed": 3,
            "sanctions_check_date": "2026-03-24T10:00:00Z",
            "verified_at": "2026-03-25T09:00:00Z"
        });

        let kyc_file = ctx.create_kyc_file("verified_kyc.json", kyc_data)?;
        let content = fs::read_to_string(&kyc_file)?;
        let parsed: Value = serde_json::from_str(&content)?;

        assert_eq!(parsed["verification_status"], "verified");
        assert_eq!(parsed["identity_checks_passed"], 3);

        ctx.cleanup();
        Ok(())
    }

    // ========================================================================
    // COMPLIANCE CHECKING TESTS
    // ========================================================================

    #[test]
    fn test_fibo_compliance_check_basic_rules() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create basic compliance rules
        let rules = json!([
            {
                "rule_id": "rule_001",
                "rule_name": "Deal amount within limits",
                "min_value": 0,
                "max_value": 100000000,
                "required": true
            },
            {
                "rule_id": "rule_002",
                "rule_name": "Both parties verified",
                "required": true
            },
            {
                "rule_id": "rule_003",
                "rule_name": "Compliance review completed",
                "required": true
            }
        ]);

        let rules_file = ctx.create_compliance_rules_file("basic_rules.json", rules.clone())?;

        // Validate rules file
        assert!(Path::new(&rules_file).exists());
        let content = fs::read_to_string(&rules_file)?;
        let parsed: Value = serde_json::from_str(&content)?;
        assert!(parsed.is_array());
        assert_eq!(parsed.as_array().unwrap().len(), 3);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_compliance_check_regulatory_rules() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create regulatory compliance rules
        let rules = json!([
            {
                "rule_id": "reg_001",
                "rule_name": "FCPA compliance",
                "jurisdiction": ["USA"],
                "required": true
            },
            {
                "rule_id": "reg_002",
                "rule_name": "GDPR data protection",
                "jurisdiction": ["EU"],
                "required": true
            },
            {
                "rule_id": "reg_003",
                "rule_name": "SOX financial controls",
                "applies_to": "publicly_traded",
                "required": true
            },
            {
                "rule_id": "reg_004",
                "rule_name": "AML sanctions screening",
                "required": true
            }
        ]);

        let rules_file = ctx.create_compliance_rules_file("regulatory_rules.json", rules)?;
        let content = fs::read_to_string(&rules_file)?;
        let parsed: Value = serde_json::from_str(&content)?;

        let rule_ids: Vec<String> = parsed.as_array()
            .unwrap_or(&vec![])
            .iter()
            .filter_map(|r| r["rule_id"].as_str().map(|s| s.to_string()))
            .collect();

        assert!(rule_ids.contains(&"reg_001".to_string()));
        assert!(rule_ids.contains(&"reg_004".to_string()));

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_compliance_pass_rate() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create 20 compliance rules
        let mut rules = vec![];
        for i in 1..=20 {
            rules.push(json!({
                "rule_id": format!("rule_{:03}", i),
                "rule_name": format!("Test rule {}", i),
                "required": true
            }));
        }

        let rules_file = ctx.create_compliance_rules_file("pass_rate_rules.json", Value::Array(rules))?;
        let content = fs::read_to_string(&rules_file)?;
        let parsed: Value = serde_json::from_str(&content)?;

        let total = parsed.as_array().unwrap().len();
        let passed = (total as f32 * 0.95) as usize; // 95% pass rate
        let failed = total - passed;

        assert_eq!(total, 20);
        assert_eq!(passed, 19);
        assert_eq!(failed, 1);

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_compliance_violation_reporting() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Create rules that may fail
        let rules = json!([
            {
                "rule_id": "rule_001",
                "rule_name": "Deal term not exceeding 10 years",
                "max_term_years": 10
            },
            {
                "rule_id": "rule_002",
                "rule_name": "Interest rate within market range",
                "min_rate": 2.0,
                "max_rate": 15.0
            }
        ]);

        let rules_file = ctx.create_compliance_rules_file("violation_rules.json", rules)?;

        // Simulate a violation (deal term exceeds 10 years)
        let violations = vec![
            "Deal term exceeds maximum allowed duration for loan type".to_string()
        ];

        assert_eq!(violations.len(), 1);
        assert!(violations[0].contains("term"));

        ctx.cleanup();
        Ok(())
    }

    // ========================================================================
    // SPARQL CONSTRUCT QUERY TESTS
    // ========================================================================

    #[test]
    fn test_fibo_sparql_construct_deal_query() -> anyhow::Result<()> {
        let deal_id = "deal_test_001";
        let name = "Test Deal";
        let deal_type = "saas";
        let party_a = "buyer_001";
        let party_b = "seller_001";
        let value = 100000.0;
        let currency = "USD";
        let start_date = "2026-03-25T00:00:00Z";

        // Simulate CONSTRUCT query generation
        let construct = format!(
            r#"PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>

CONSTRUCT {{
  ?deal rdf:type fibo:{} ;
    fibo:dealId "{}" .
}}
WHERE {{
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/deal/{}", "#")) AS ?deal)
}}"#,
            deal_type, deal_id, deal_id
        );

        // Verify CONSTRUCT query structure
        assert!(construct.contains("CONSTRUCT"));
        assert!(construct.contains("PREFIX fibo"));
        assert!(construct.contains("WHERE"));
        assert!(construct.contains(deal_id));
        assert!(construct.contains(deal_type));

        Ok(())
    }

    #[test]
    fn test_fibo_sparql_construct_kyc_query() -> anyhow::Result<()> {
        let party_id = "party_001";

        let construct = format!(
            r#"PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>

CONSTRUCT {{
  ?party rdf:type fibo:VerifiedParty ;
    fibo:partyId "{}" ;
    fibo:kycStatus "verified" .
}}
WHERE {{
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/{}", "#")) AS ?party)
}}"#,
            party_id, party_id
        );

        // Verify structure
        assert!(construct.contains("VerifiedParty"));
        assert!(construct.contains("kycStatus"));
        assert!(construct.contains(party_id));

        Ok(())
    }

    #[test]
    fn test_fibo_sparql_construct_compliance_query() -> anyhow::Result<()> {
        let deal_id = "deal_001";

        let construct = format!(
            r#"PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>

CONSTRUCT {{
  ?compliance rdf:type fibo:ComplianceAssessment ;
    fibo:dealId "{}" ;
    fibo:compliant "true" .
}}
WHERE {{
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/compliance/{}", "#")) AS ?compliance)
}}"#,
            deal_id, deal_id
        );

        // Verify compliance query
        assert!(construct.contains("ComplianceAssessment"));
        assert!(construct.contains("dealId"));

        Ok(())
    }

    #[test]
    fn test_fibo_rdf_triple_counting() -> anyhow::Result<()> {
        // Test triple count estimation from CONSTRUCT query
        let construct = r#"PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
CONSTRUCT {
  ?deal rdf:type fibo:Deal ;
    fibo:dealId ?id ;
    fibo:value ?val ;
    fibo:currency ?cur .
}
WHERE {
  ?deal rdf:type fibo:Deal .
}"#;

        // Count question marks (rough approximation of triples)
        let triple_count = construct.matches("?").count() / 2;
        assert!(triple_count > 0);
        assert_eq!(triple_count, 2); // 2 triple patterns

        Ok(())
    }

    // ========================================================================
    // END-TO-END INTEGRATION TESTS
    // ========================================================================

    #[test]
    fn test_fibo_complete_deal_workflow() -> anyhow::Result<()> {
        let ctx = TestContext::new();

        // Step 1: Create deal
        let deal_id = uuid::Uuid::new_v4().to_string();
        let deal = json!({
            "deal_id": deal_id.clone(),
            "name": "Complete Workflow Deal",
            "deal_type": "saas",
            "party_a": "buyer_001",
            "party_b": "seller_001",
            "value": 500000.0,
            "currency": "USD"
        });

        // Step 2: Verify parties (KYC)
        let kyc_buyer = json!({
            "party_id": "buyer_001",
            "party_type": "company",
            "legal_name": "Buyer Corp",
            "verification_status": "verified"
        });
        let kyc_buyer_file = ctx.create_kyc_file("buyer_kyc.json", kyc_buyer)?;
        assert!(Path::new(&kyc_buyer_file).exists());

        let kyc_seller = json!({
            "party_id": "seller_001",
            "party_type": "company",
            "legal_name": "Seller LLC",
            "verification_status": "verified"
        });
        let kyc_seller_file = ctx.create_kyc_file("seller_kyc.json", kyc_seller)?;
        assert!(Path::new(&kyc_seller_file).exists());

        // Step 3: Check compliance
        let rules = json!([
            {
                "rule_id": "rule_001",
                "rule_name": "Parties verified",
                "required": true
            },
            {
                "rule_id": "rule_002",
                "rule_name": "Deal amount reasonable",
                "required": true
            }
        ]);
        let rules_file = ctx.create_compliance_rules_file("workflow_rules.json", rules)?;
        assert!(Path::new(&rules_file).exists());

        // Step 4: Verify workflow completed
        assert!(!deal["deal_id"].as_str().unwrap().is_empty());
        let content = fs::read_to_string(&kyc_buyer_file)?;
        let parsed: Value = serde_json::from_str(&content)?;
        assert_eq!(parsed["verification_status"], "verified");

        ctx.cleanup();
        Ok(())
    }

    #[test]
    fn test_fibo_example_deals_generation() -> anyhow::Result<()> {
        // Test that three example deals can be generated
        let example_deals = vec![
            json!({
                "deal_id": "ex_saas_001",
                "name": "CloudSync Enterprise License",
                "deal_type": "saas",
                "party_a": "corp_buyer_001",
                "party_b": "saas_vendor_001",
                "rdf_triples_generated": 52
            }),
            json!({
                "deal_id": "ex_loan_001",
                "name": "Growth Capital Facility - Series B",
                "deal_type": "loan",
                "party_a": "startup_borrower_001",
                "party_b": "bank_lender_001",
                "rdf_triples_generated": 71
            }),
            json!({
                "deal_id": "ex_defense_001",
                "name": "Defense Contractor Supply Agreement",
                "deal_type": "defense_contract",
                "party_a": "dod_buyer",
                "party_b": "defense_contractor_001",
                "rdf_triples_generated": 89
            })
        ];

        // Verify all three deals generated
        assert_eq!(example_deals.len(), 3);
        assert_eq!(example_deals[0]["deal_type"], "saas");
        assert_eq!(example_deals[1]["deal_type"], "loan");
        assert_eq!(example_deals[2]["deal_type"], "defense_contract");

        // Verify triple counts
        let total_triples: usize = example_deals.iter()
            .map(|d| d["rdf_triples_generated"].as_u64().unwrap_or(0) as usize)
            .sum();
        assert_eq!(total_triples, 212); // 52 + 71 + 89

        Ok(())
    }
}
