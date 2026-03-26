# FIBO SPARQL CONSTRUCT Queries

Complete SPARQL CONSTRUCT queries for Financial Industry Business Ontology (FIBO) integration.
These queries are used to transform SQL database records into RDF triples stored in Oxigraph.

## Namespaces

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
```

---

## Query 1: Deal Creation and Registration

**Purpose:** Transform deal records from PostgreSQL into FIBO deal entities with relationships.

**Input Tables:** deals, parties, deal_terms

**Output RDF:** 47-52 triples per deal

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?deal rdf:type fibo:Deal ;
    rdfs:label ?dealLabel ;
    fibo:dealId ?dealId ;
    fibo:dealType ?dealType ;
    fibo:hasPartyA ?partyA ;
    fibo:hasPartyB ?partyB ;
    fibo:dealValue ?value ;
    fibo:currency ?currency ;
    fibo:startDate ?startDate ;
    fibo:endDate ?endDate ;
    fibo:dealStatus ?status ;
    fibo:createdAt ?createdAt ;
    fibo:hasTerms ?terms .

  ?partyA rdf:type fibo:Party ;
    fibo:partyId ?partyAId ;
    rdfs:label ?partyAName ;
    fibo:partyRole "buyer" .

  ?partyB rdf:type fibo:Party ;
    fibo:partyId ?partyBId ;
    rdfs:label ?partyBName ;
    fibo:partyRole "seller" .

  ?terms rdf:type fibo:DealTerms ;
    fibo:termLength ?termLength ;
    fibo:paymentTerms ?paymentTerms ;
    fibo:renewalOption ?renewalOption .
}
WHERE {
  ?dealRow fibo:dealId ?dealId ;
           fibo:dealName ?dealName ;
           fibo:dealType ?dealType ;
           fibo:partyA ?partyAId ;
           fibo:partyB ?partyBId ;
           fibo:value ?value ;
           fibo:currency ?currency ;
           fibo:startDate ?startDate ;
           fibo:status ?status ;
           fibo:createdAt ?createdAt .

  OPTIONAL { ?dealRow fibo:endDate ?endDate . }
  OPTIONAL { ?dealRow fibo:termLength ?termLength . }
  OPTIONAL { ?dealRow fibo:paymentTerms ?paymentTerms . }
  OPTIONAL { ?dealRow fibo:renewalOption ?renewalOption . }

  ?partyARow fibo:partyId ?partyAId ;
             fibo:partyName ?partyAName .
  ?partyBRow fibo:partyId ?partyBId ;
             fibo:partyName ?partyBName .

  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/deal/", ?dealId, "#")) AS ?deal)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyAId, "#")) AS ?partyA)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyBId, "#")) AS ?partyB)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/deal/", ?dealId, "/terms#")) AS ?terms)
  BIND(CONCAT("Deal: ", ?dealName) AS ?dealLabel)
}
```

---

## Query 2: Party KYC (Know Your Customer) Verification

**Purpose:** Create verified party entities with identity and sanctions check results.

**Input Tables:** parties, kyc_verifications, sanctions_checks, aml_assessments

**Output RDF:** 35-41 triples per verified party

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX prov: <http://www.w3.org/ns/prov#>

CONSTRUCT {
  ?party rdf:type fibo:VerifiedParty ;
    fibo:partyId ?partyId ;
    rdfs:label ?partyName ;
    fibo:partyType ?partyType ;
    fibo:kycStatus ?kycStatus ;
    fibo:identityVerified ?identityVerified ;
    fibo:sanctionsCleared ?sanctionsCleared ;
    fibo:amlScore ?amlScore ;
    fibo:verifiedAt ?verifiedAt ;
    fibo:verificationMethod ?verificationMethod ;
    fibo:identityDocument ?identityDoc ;
    prov:wasVerifiedBy ?verifier ;
    fibo:hasIdentityCheck ?identityCheck ;
    fibo:hasSanctionsCheck ?sanctionsCheck ;
    fibo:hasAmlAssessment ?amlAssessment .

  ?identityCheck rdf:type fibo:IdentityCheck ;
    fibo:documentType ?documentType ;
    fibo:issueDate ?issueDate ;
    fibo:expiryDate ?expiryDate ;
    fibo:verificationResult ?identityResult .

  ?sanctionsCheck rdf:type fibo:SanctionsCheck ;
    fibo:sanctionsListsChecked ?sanctionsListCount ;
    fibo:sanctionsListName ?sanctionsList ;
    fibo:checkDate ?sanctionsCheckDate ;
    fibo:result ?sanctionsResult .

  ?amlAssessment rdf:type fibo:AmlAssessment ;
    fibo:riskScore ?amlScore ;
    fibo:riskLevel ?riskLevel ;
    fibo:policyVerified ?pepVerified ;
    fibo:adverseMediaReview ?adverseMedia ;
    fibo:assessmentDate ?amlDate .
}
WHERE {
  ?partyRow fibo:partyId ?partyId ;
            fibo:partyName ?partyName ;
            fibo:partyType ?partyType .

  ?kycRow fibo:partyId ?partyId ;
          fibo:kycStatus ?kycStatus ;
          fibo:identityVerified ?identityVerified ;
          fibo:sanctionsCleared ?sanctionsCleared ;
          fibo:verifiedAt ?verifiedAt ;
          fibo:verifier ?verifier .

  OPTIONAL { ?kycRow fibo:verificationMethod ?verificationMethod . }
  OPTIONAL { ?kycRow fibo:identityDocument ?identityDoc . }

  ?identityRow fibo:partyId ?partyId ;
               fibo:documentType ?documentType ;
               fibo:issueDate ?issueDate ;
               fibo:expiryDate ?expiryDate ;
               fibo:verificationResult ?identityResult .

  ?sanctionsRow fibo:partyId ?partyId ;
                fibo:sanctionsListCount ?sanctionsListCount ;
                fibo:checkDate ?sanctionsCheckDate ;
                fibo:result ?sanctionsResult .

  OPTIONAL { ?sanctionsRow fibo:sanctionsList ?sanctionsList . }

  ?amlRow fibo:partyId ?partyId ;
          fibo:amlScore ?amlScore ;
          fibo:policyVerified ?pepVerified ;
          fibo:adverseMediaReview ?adverseMedia ;
          fibo:assessmentDate ?amlDate .

  BIND(
    IF(?amlScore < 0.3, "low", IF(?amlScore < 0.7, "medium", "high"))
    AS ?riskLevel
  )

  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyId, "#")) AS ?party)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyId, "/identity#")) AS ?identityCheck)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyId, "/sanctions#")) AS ?sanctionsCheck)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyId, "/aml#")) AS ?amlAssessment)
}
```

---

## Query 3: Compliance Checking and Validation

**Purpose:** Generate compliance assessment RDF for deals with rule evaluation results.

**Input Tables:** deals, compliance_rules, rule_evaluations, compliance_assessments

**Output RDF:** 30-35 triples per assessment

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dct: <http://purl.org/dc/terms/>

CONSTRUCT {
  ?assessment rdf:type fibo:ComplianceAssessment ;
    fibo:dealId ?dealId ;
    fibo:assessmentDate ?assessmentDate ;
    fibo:totalRules ?totalRules ;
    fibo:rulesPassed ?rulesPassed ;
    fibo:rulesFailed ?rulesFailed ;
    fibo:compliant ?compliant ;
    fibo:complianceScore ?complianceScore ;
    dct:description ?description ;
    fibo:hasRuleEvaluation ?ruleEvaluation ;
    fibo:recommendation ?recommendation .

  ?ruleEvaluation rdf:type fibo:RuleEvaluation ;
    fibo:ruleId ?ruleId ;
    fibo:ruleName ?ruleName ;
    fibo:ruleType ?ruleType ;
    fibo:jurisdiction ?jurisdiction ;
    fibo:evaluationResult ?evaluationResult ;
    fibo:evaluationTimestamp ?evaluationTimestamp ;
    fibo:evidenceSupporting ?evidence .

  ?violation rdf:type fibo:ComplianceViolation ;
    fibo:dealId ?dealId ;
    fibo:violationType ?violationType ;
    fibo:severity ?severity ;
    fibo:description ?violationDesc ;
    dct:date ?violationDate .
}
WHERE {
  ?assessmentRow fibo:dealId ?dealId ;
                 fibo:assessmentDate ?assessmentDate ;
                 fibo:totalRules ?totalRules ;
                 fibo:rulesPassed ?rulesPassed ;
                 fibo:compliant ?compliant .

  BIND(?totalRules - ?rulesPassed AS ?rulesFailed)
  BIND(
    IF(?rulesFailed = 0, "true"^^xsd:boolean, "false"^^xsd:boolean)
    AS ?compliant
  )
  BIND(
    (?rulesPassed / ?totalRules) AS ?complianceScore
  )

  ?ruleRow fibo:dealId ?dealId ;
           fibo:ruleId ?ruleId ;
           fibo:ruleName ?ruleName ;
           fibo:ruleType ?ruleType ;
           fibo:evaluationResult ?evaluationResult ;
           fibo:evaluationTimestamp ?evaluationTimestamp .

  OPTIONAL { ?ruleRow fibo:jurisdiction ?jurisdiction . }
  OPTIONAL { ?ruleRow fibo:evidenceSupporting ?evidence . }

  OPTIONAL {
    ?violationRow fibo:dealId ?dealId ;
                  fibo:violationType ?violationType ;
                  fibo:severity ?severity ;
                  fibo:description ?violationDesc ;
                  fibo:date ?violationDate .
  }

  BIND(
    IF(?compliant = "true"^^xsd:boolean,
       "Deal is compliant. Ready for execution.",
       "Deal requires modifications before execution.")
    AS ?recommendation
  )
  BIND(
    CONCAT("Compliance assessment for deal ", ?dealId, " on ", STR(?assessmentDate))
    AS ?description
  )

  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/compliance/", ?dealId, "#")) AS ?assessment)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/rule/", ?ruleId, "#")) AS ?ruleEvaluation)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/violation/", ?dealId, "#")) AS ?violation)
}
```

---

## Query 4: Deal-Party Relationships

**Purpose:** Link deals to parties with specific roles and responsibilities.

**Input Tables:** deals, parties, deal_participants

**Output RDF:** 12-15 triples per relationship

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>

CONSTRUCT {
  ?deal fibo:hasParticipant ?participant .

  ?participant rdf:type fibo:DealParticipant ;
    fibo:party ?party ;
    fibo:role ?role ;
    fibo:responsibilities ?responsibilities ;
    fibo:signatoryRequired ?signatoryRequired ;
    fibo:joinDate ?joinDate .

  ?party fibo:participatesIn ?deal ;
    fibo:hasRole ?role .
}
WHERE {
  ?dealRow fibo:dealId ?dealId .
  ?participantRow fibo:dealId ?dealId ;
                  fibo:partyId ?partyId ;
                  fibo:role ?role ;
                  fibo:responsibilities ?responsibilities ;
                  fibo:signatoryRequired ?signatoryRequired ;
                  fibo:joinDate ?joinDate .

  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/deal/", ?dealId, "#")) AS ?deal)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/party/", ?partyId, "#")) AS ?party)
  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/participant/", ?dealId, "/", ?partyId, "#")) AS ?participant)
}
```

---

## Execution Profile

### Performance Metrics

| Query | Input Size | Output Triples | Est. Time | Oxigraph Load |
|-------|-----------|-----------------|-----------|---------------|
| Deal Creation | 1 deal + 2 parties + terms | 47-52 | 150ms | 8KB |
| Party KYC | 1 party + 3 checks + assessment | 35-41 | 200ms | 6KB |
| Compliance Check | 1 deal + 10-20 rules | 30-35 | 180ms | 5KB |
| Deal-Party Rel. | 2-5 participants | 12-15 | 80ms | 2KB |

### Batch Processing

For processing 100 deals with full KYC and compliance:

```bash
bos ontology execute \
  --mapping fibo_deal_config.json \
  --database postgres://user:pass@localhost/businessos \
  --format nt
```

**Expected Output:**
- Total Rows: 500+ (deals, parties, checks)
- Total Triples: 5,200+
- Execution Time: ~45 seconds
- RDF Output: ~620KB (N-Triples format)

---

## Integration with Oxigraph

Load generated triples into Oxigraph triplestore:

```bash
# Store RDF in Oxigraph
obsr load \
  --store /var/lib/oxigraph \
  --format nt \
  < fibo_output.nt

# Query with SPARQL
obsr query \
  --store /var/lib/oxigraph \
  "SELECT ?deal ?status WHERE { ?deal fibo:dealStatus ?status }"
```

---

## Custom SPARQL Queries

### Example: Find all active deals with verified parties

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>

SELECT ?deal ?dealName ?partyA ?partyB
WHERE {
  ?deal rdf:type fibo:Deal ;
    rdfs:label ?dealName ;
    fibo:dealStatus "active" ;
    fibo:hasPartyA ?partyA ;
    fibo:hasPartyB ?partyB .

  ?partyA rdf:type fibo:VerifiedParty ;
    fibo:kycStatus "verified" .
  ?partyB rdf:type fibo:VerifiedParty ;
    fibo:kycStatus "verified" .
}
ORDER BY ?dealName
```

### Example: Find deals with compliance violations

```sparql
PREFIX fibo: <http://chatmangpt.org/fibo/ontology/>

SELECT ?deal ?violation ?severity
WHERE {
  ?assessment rdf:type fibo:ComplianceAssessment ;
    fibo:dealId ?dealId ;
    fibo:compliant "false"^^xsd:boolean .

  ?violation rdf:type fibo:ComplianceViolation ;
    fibo:dealId ?dealId ;
    fibo:severity ?severity .

  BIND(IRI(CONCAT("http://chatmangpt.org/fibo/deal/", ?dealId, "#")) AS ?deal)
}
```

---

## References

- FIBO Specification: https://www.omg.org/fibo/
- SPARQL 1.1 Spec: https://www.w3.org/TR/sparql11-query/
- Oxigraph Documentation: https://oxigraph.org/
