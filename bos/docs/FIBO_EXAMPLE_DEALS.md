# FIBO Example Deals

Three complete example deals demonstrating FIBO ontology integration for different deal types.

---

## Example 1: SaaS Deal

**Deal ID:** `ex_saas_001`
**Type:** Software-as-a-Service License Agreement
**Generated Triples:** 52

### Deal Data

```json
{
  "deal_id": "ex_saas_001",
  "name": "CloudSync Enterprise License",
  "deal_type": "saas",
  "party_a": "corp_buyer_001",
  "party_b": "saas_vendor_001",
  "value": 250000.0,
  "currency": "USD",
  "start_date": "2026-04-01T00:00:00Z",
  "end_date": "2027-03-31T23:59:59Z",
  "status": "active",
  "created_at": "2026-03-25T10:30:00Z"
}
```

### Parties

**Buyer (Party A) - corp_buyer_001**
```json
{
  "party_id": "corp_buyer_001",
  "name": "Acme Corporation",
  "party_type": "company",
  "registration_number": "12-3456789",
  "jurisdiction": "Delaware",
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.12,
  "verified_at": "2026-03-20T14:00:00Z"
}
```

**Seller (Party B) - saas_vendor_001**
```json
{
  "party_id": "saas_vendor_001",
  "name": "CloudSync Inc.",
  "party_type": "company",
  "registration_number": "98-7654321",
  "jurisdiction": "California",
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.08,
  "verified_at": "2026-03-18T09:30:00Z"
}
```

### Deal Terms

```json
{
  "deal_id": "ex_saas_001",
  "term_length_months": 12,
  "payment_terms": "monthly",
  "payment_amount": 20833.33,
  "users_licensed": 500,
  "data_retention_days": 2555,
  "sla_uptime_percent": 99.9,
  "support_level": "premium",
  "renewal_option": "auto_renew",
  "cancellation_notice_days": 30,
  "ip_ownership": "vendor"
}
```

### Compliance Rules Applied

1. **Party Verification** - Both parties KYC verified ✓
2. **Deal Amount** - Within standard SaaS range ($100K-$1M) ✓
3. **Data Protection** - GDPR compliance for EU data ✓
4. **Service Levels** - SLA defined (99.9% uptime) ✓
5. **Termination Clause** - 30-day cancellation notice ✓

### Generated RDF Triples (Example)

```ttl
@prefix fibo: <http://chatmangpt.org/fibo/ontology/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<http://chatmangpt.org/fibo/deal/ex_saas_001#>
  rdf:type fibo:Deal ;
  rdfs:label "CloudSync Enterprise License" ;
  fibo:dealId "ex_saas_001" ;
  fibo:dealType "saas" ;
  fibo:hasPartyA <http://chatmangpt.org/fibo/party/corp_buyer_001#> ;
  fibo:hasPartyB <http://chatmangpt.org/fibo/party/saas_vendor_001#> ;
  fibo:dealValue "250000.0"^^xsd:decimal ;
  fibo:currency "USD" ;
  fibo:startDate "2026-04-01T00:00:00Z"^^xsd:dateTime ;
  fibo:endDate "2027-03-31T23:59:59Z"^^xsd:dateTime ;
  fibo:dealStatus "active" ;
  fibo:hasTerms <http://chatmangpt.org/fibo/deal/ex_saas_001/terms#> .

<http://chatmangpt.org/fibo/party/corp_buyer_001#>
  rdf:type fibo:VerifiedParty ;
  fibo:partyId "corp_buyer_001" ;
  rdfs:label "Acme Corporation" ;
  fibo:partyType "company" ;
  fibo:kycStatus "verified" ;
  fibo:amlScore "0.12"^^xsd:float .
```

### Compliance Assessment

| Rule | Status | Notes |
|------|--------|-------|
| Party A Verified | PASS | Acme Corp verified 2026-03-20 |
| Party B Verified | PASS | CloudSync verified 2026-03-18 |
| Deal Value Allowed | PASS | $250K within SaaS limits |
| SLA Defined | PASS | 99.9% uptime SLA |
| Data Residency | PASS | US data storage (no GDPR conflict) |
| Termination Rights | PASS | 30-day notice included |

**Overall Compliance:** ✓ COMPLIANT - Ready for execution

---

## Example 2: Business Loan Deal

**Deal ID:** `ex_loan_001`
**Type:** Growth Capital Facility (Series B)
**Generated Triples:** 71

### Deal Data

```json
{
  "deal_id": "ex_loan_001",
  "name": "Growth Capital Facility - Series B",
  "deal_type": "loan",
  "party_a": "startup_borrower_001",
  "party_b": "bank_lender_001",
  "value": 5000000.0,
  "currency": "USD",
  "start_date": "2026-02-01T00:00:00Z",
  "end_date": "2031-02-01T23:59:59Z",
  "status": "active",
  "created_at": "2026-01-20T14:30:00Z"
}
```

### Parties

**Borrower (Party A) - startup_borrower_001**
```json
{
  "party_id": "startup_borrower_001",
  "name": "TechVenture Labs",
  "party_type": "company",
  "registration_number": "45-6789012",
  "jurisdiction": "Delaware",
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.18,
  "verified_at": "2026-01-15T10:00:00Z"
}
```

**Lender (Party B) - bank_lender_001**
```json
{
  "party_id": "bank_lender_001",
  "name": "Enterprise Bank Corp",
  "party_type": "company",
  "registration_number": "99-1234567",
  "jurisdiction": "New York",
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.05,
  "verified_at": "2026-01-10T11:00:00Z"
}
```

### Loan Terms

```json
{
  "deal_id": "ex_loan_001",
  "loan_amount": 5000000.0,
  "interest_rate": 7.5,
  "term_years": 5,
  "repayment_schedule": "quarterly",
  "quarterly_payment": 312500.0,
  "origination_fee_percent": 1.5,
  "prepayment_penalty": false,
  "subordination": "none",
  "covenants": [
    "minimum_debt_service_coverage_ratio_1.25x",
    "maximum_leverage_ratio_3.0x",
    "minimum_cash_reserve_percent_10%"
  ],
  "collateral": "equipment_and_receivables",
  "guarantor": "founder_personal_guarantee"
}
```

### Compliance Rules Applied

1. **Lender Verification** - Bank registered and compliance-verified ✓
2. **Borrower Financial** - Startup KYC verified, financials reviewed ✓
3. **Interest Rate** - 7.5% within market range for Series B loans ✓
4. **Debt Covenants** - Standard financial covenants in place ✓
5. **Consumer Protection** - Truth in Lending Act compliance ✓
6. **AML/CFT** - Both parties sanctions-cleared ✓

### Generated RDF Triples (Example)

```ttl
@prefix fibo: <http://chatmangpt.org/fibo/ontology/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<http://chatmangpt.org/fibo/deal/ex_loan_001#>
  rdf:type fibo:Deal ;
  fibo:dealType "loan" ;
  fibo:dealValue "5000000.0"^^xsd:decimal ;
  fibo:hasTerms <http://chatmangpt.org/fibo/deal/ex_loan_001/terms#> .

<http://chatmangpt.org/fibo/deal/ex_loan_001/terms#>
  rdf:type fibo:LoanTerms ;
  fibo:interestRate "7.5"^^xsd:float ;
  fibo:termYears "5"^^xsd:int ;
  fibo:repaymentSchedule "quarterly" ;
  fibo:originationFeePercent "1.5"^^xsd:float ;
  fibo:prepaymentPenalty "false"^^xsd:boolean ;
  fibo:collateral "equipment_and_receivables" ;
  fibo:hasCovenant fibo:DSCR_1_25x ;
  fibo:hasCovenant fibo:LeverageRatio_3_0x .
```

### Compliance Assessment

| Rule | Status | Notes |
|------|--------|-------|
| Lender Verification | PASS | Enterprise Bank verified 2026-01-10 |
| Borrower Verification | PASS | TechVenture Labs verified 2026-01-15 |
| Interest Rate Allowed | PASS | 7.5% within range for Series B |
| Debt Covenants | PASS | Standard covenants: DSCR 1.25x, Leverage 3.0x |
| Truth in Lending | PASS | Origination fee and rate disclosed |
| AML Compliance | PASS | Both parties sanctions-cleared |
| Collateral Perfection | PENDING | UCC-1 filing verification needed |

**Overall Compliance:** ⚠️ CONDITIONAL - Awaiting UCC-1 filing confirmation

---

## Example 3: Defense Contractor Supply Agreement

**Deal ID:** `ex_defense_001`
**Type:** Defense Contract (Classified)
**Generated Triples:** 89

### Deal Data

```json
{
  "deal_id": "ex_defense_001",
  "name": "Defense Contractor Supply Agreement",
  "deal_type": "defense_contract",
  "party_a": "dod_buyer",
  "party_b": "defense_contractor_001",
  "value": 12500000.0,
  "currency": "USD",
  "start_date": "2026-05-01T00:00:00Z",
  "end_date": "2028-04-30T23:59:59Z",
  "status": "draft",
  "classified": true,
  "classification_level": "secret",
  "created_at": "2026-03-24T08:00:00Z"
}
```

### Parties

**Buyer (Party A) - dod_buyer**
```json
{
  "party_id": "dod_buyer",
  "name": "Department of Defense - Procurement",
  "party_type": "government_agency",
  "jurisdiction": "Federal",
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.0,
  "verified_at": "2026-03-20T12:00:00Z"
}
```

**Seller (Party B) - defense_contractor_001**
```json
{
  "party_id": "defense_contractor_001",
  "name": "Precision Defense Systems Inc.",
  "party_type": "company",
  "registration_number": "56-7890123",
  "jurisdiction": "Connecticut",
  "security_clearance_level": "secret",
  "security_clearance_expiry": "2028-06-30",
  "dfars_compliant": true,
  "itar_registered": true,
  "kyc_status": "verified",
  "identity_verified": true,
  "sanctions_cleared": true,
  "aml_score": 0.06,
  "verified_at": "2026-03-22T09:15:00Z"
}
```

### Contract Terms

```json
{
  "deal_id": "ex_defense_001",
  "contract_type": "firm_fixed_price",
  "firm_fixed_price": 12500000.0,
  "delivery_timeline": "24_months",
  "payment_schedule": [
    {
      "milestone": "contract_signature",
      "percent": 10,
      "amount": 1250000.0
    },
    {
      "milestone": "design_review_complete",
      "percent": 25,
      "amount": 3125000.0
    },
    {
      "milestone": "prototype_delivery",
      "percent": 35,
      "amount": 4375000.0
    },
    {
      "milestone": "final_acceptance",
      "percent": 30,
      "amount": 3750000.0
    }
  ],
  "quality_requirements": "MIL-STD-9858A",
  "inspection_and_acceptance": "government_witnessed",
  "warranty_period_months": 24,
  "compliance_requirements": [
    "DFARS_clauses",
    "ITAR_compliance",
    "NIST_cybersecurity_standards",
    "security_classification_guide",
    "export_control"
  ],
  "security_requirements": {
    "facility_clearance": "secret",
    "personnel_clearance": "secret",
    "information_security_plan": "required",
    "facility_inspections": "quarterly"
  },
  "intellectual_property": "government_purpose_rights",
  "indemnification": "contractor_liable"
}
```

### Compliance Rules Applied

1. **DOD Verification** - Federal procurement authority verified ✓
2. **Contractor Security Clearance** - Secret clearance valid through 2028-06-30 ✓
3. **DFARS Compliance** - Defense Federal Acquisition Regulation Supplement ✓
4. **ITAR Compliance** - International Traffic in Arms Regulations ✓
5. **Cybersecurity** - NIST SP 800-171 controls required ✓
6. **Classification Handling** - Secret material handling procedures ✓
7. **Export Control** - EAR/ITAR export authorization ✓
8. **Quality Standards** - MIL-STD-9858A quality management ✓
9. **Facility Inspections** - Quarterly security inspections ✓
10. **Flow-Down Requirements** - Subcontractor compliance clauses ✓

### Generated RDF Triples (Example)

```ttl
@prefix fibo: <http://chatmangpt.org/fibo/ontology/> .
@prefix govprocure: <http://chatmangpt.org/govprocure/ontology/> .

<http://chatmangpt.org/fibo/deal/ex_defense_001#>
  rdf:type fibo:DefenseContract ;
  govprocure:contractType "firm_fixed_price" ;
  govprocure:firmFixedPrice "12500000.0"^^xsd:decimal ;
  govprocure:classified "true"^^xsd:boolean ;
  govprocure:classificationLevel "secret" ;
  govprocure:hasSecurityRequirements <http://chatmangpt.org/fibo/deal/ex_defense_001/security#> ;
  govprocure:hasComplianceRequirement govprocure:DFARS ;
  govprocure:hasComplianceRequirement govprocure:ITAR ;
  govprocure:qualityStandard "MIL-STD-9858A" .

<http://chatmangpt.org/fibo/deal/ex_defense_001/security#>
  rdf:type govprocure:SecurityRequirements ;
  govprocure:facilitySecurityClearance "secret" ;
  govprocure:personnelSecurityClearance "secret" ;
  govprocure:facilityInspectionFrequency "quarterly" ;
  govprocure:informationSecurityPlanRequired "true"^^xsd:boolean .

<http://chatmangpt.org/fibo/party/defense_contractor_001#>
  rdf:type fibo:DefenseContractor ;
  govprocure:securityClearanceLevel "secret" ;
  govprocure:securityClearanceExpiry "2028-06-30"^^xsd:date ;
  govprocure:dfarsCompliant "true"^^xsd:boolean ;
  govprocure:itarRegistered "true"^^xsd:boolean .
```

### Compliance Assessment

| Rule | Status | Notes |
|------|--------|-------|
| DOD Verification | PASS | Federal procurement authority verified |
| Contractor Clearance | PASS | Secret clearance valid through 2028-06-30 |
| DFARS Compliance | PASS | All required clauses included in contract |
| ITAR Compliance | PASS | Contractor ITAR-registered, tech data controlled |
| Cybersecurity (NIST) | PASS | NIST SP 800-171 requirements imposed |
| Classification | PASS | Secret facility and personnel clearance verified |
| Export Control | PASS | Items subject to EAR/ITAR, authorization required |
| Quality Standards | PASS | MIL-STD-9858A quality system in place |
| Facility Inspection | PASS | Quarterly inspection schedule established |
| Flow-Down Clauses | PASS | Subcontractor compliance requirements included |

**Overall Compliance:** ✓ COMPLIANT (SECRET) - Ready for signature (classified handling)

---

## Usage Examples

### Generate Example Deals

```bash
bos fibo deal examples
```

Output:
```json
[
  {
    "deal_id": "ex_saas_001",
    "name": "CloudSync Enterprise License",
    "deal_type": "saas",
    "rdf_triples_generated": 52
  },
  {
    "deal_id": "ex_loan_001",
    "name": "Growth Capital Facility - Series B",
    "deal_type": "loan",
    "rdf_triples_generated": 71
  },
  {
    "deal_id": "ex_defense_001",
    "name": "Defense Contractor Supply Agreement",
    "deal_type": "defense_contract",
    "rdf_triples_generated": 89
  }
]
```

### Load Example Deal into Oxigraph

```bash
# Create deal
bos fibo deal create \
  --name "CloudSync Enterprise License" \
  --deal-type saas \
  --party-a corp_buyer_001 \
  --party-b saas_vendor_001 \
  --value 250000 \
  --currency USD
```

### Export to RDF

```bash
# Execute CONSTRUCT queries and store in Oxigraph
bos ontology execute \
  --mapping fibo_deal_config.json \
  --database postgres://localhost/businessos

# Query results
obsr query --store /var/lib/oxigraph \
  "SELECT ?deal ?status WHERE {
     ?deal fibo:dealStatus ?status
   }"
```

---

## Summary

| Deal Type | Deal ID | Value | Parties | Triples | Status |
|-----------|---------|-------|---------|---------|--------|
| SaaS | ex_saas_001 | $250K | 2 | 52 | ✓ COMPLIANT |
| Loan | ex_loan_001 | $5M | 2 | 71 | ⚠️ CONDITIONAL |
| Defense | ex_defense_001 | $12.5M | 2 | 89 | ✓ COMPLIANT (SECRET) |

**Total RDF Triples Generated:** 212
**Total Deal Value:** $17.75M
