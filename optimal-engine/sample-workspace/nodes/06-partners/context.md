---
node: 06-partners
kind: org
style: external
---

# Partner Network — context

Channel + integration partners who resell, deliver, or embed the
platform. External-style node (not a headcount org), but every
interaction is a first-class signal here.

## Active partner tiers

- **Healthtech** (primary) — 3 signed, 2 at renewal risk.
- **Adjacent verticals** — legal-tech + education, exploratory.

## Team

| Role                  | Person |
|-----------------------|--------|
| Partner success lead  | Dan    |
| Technical handoff     | Eve    |
| Legal / contracts     | Heidi  |

## Key partners

| Partner      | Status     | Primary contact     |
|--------------|------------|---------------------|
| Acme Clinics | Active     | internal only       |
| BetaHealth   | Active     | internal only       |
| CareLink     | Active     | internal only       |
| DentAI       | Active     | internal only       |
| EpicMed      | At-risk    | Dan                 |
| FastScan     | At-risk    | Dan                 |

## Operating rules

- Every partner event (kickoff, renewal, escalation) gets a dated signal.
- Renewals: tracked via `frontmatter.renewal_date`; 30-day window
  surfaces via retention sweep.
- No partner contact info in this node's files — keep in the encrypted
  credentials store (connector kind = `crm` when wired).
