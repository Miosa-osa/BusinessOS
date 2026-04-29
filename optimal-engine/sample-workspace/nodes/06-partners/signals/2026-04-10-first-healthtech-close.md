---
title: First healthtech deal closed
genre: note
mode: linguistic
node: 06-partners
authored_at: 2026-04-10T16:30:00Z
sn_ratio: 0.8
entities:
  - { name: "Dan", type: person }
  - { name: "Eve", type: person }
  - { name: "Alice", type: person }
  - { name: "Healthtech Product", type: product }
  - { name: "Partner Network", type: org }
---

## Summary

Partner Dan closed the first healthtech deal with a Sacramento clinic
network. Onboarding handoff to Eve; Alice CC'd for context.

## Details

- Contract value: $180K ARR.
- 3-year term with annual price protection.
- First 25 seats active at go-live; ramp to 250 over 6 months.
- Integration surface: inbound connector (their EHR) + outbound
  reporting webhook.

## Handoff

Eve owns technical onboarding. Runbook target: go-live in ≤ 10 days.

| Day | Milestone                        |
|-----|----------------------------------|
| 1   | Kickoff + credential exchange    |
| 2-4 | Data mapping + staging ingest    |
| 5   | Staging deploy + smoke tests     |
| 7   | Pilot users (≤ 5 seats)          |
| 10  | Go-live — full 25-seat activation |

## What worked

- The partner traction slide in the deck closed the skepticism on
  "will this scale to clinical data?"
- Dan pre-aligned Heidi on the BAA language before the contract call —
  zero legal back-and-forth.

## What to improve

- Onboarding runbook needs to be a standalone document, not pieced
  together from Slack. See `02-platform-core` for that runbook.
