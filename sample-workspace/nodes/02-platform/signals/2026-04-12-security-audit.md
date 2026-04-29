---
title: Security audit — remediation plan
genre: decision_log
mode: linguistic
node: 02-platform
authored_at: 2026-04-12T14:00:00Z
sn_ratio: 0.8
entities:
  - { name: "Carol", type: person }
  - { name: "Bob", type: person }
  - { name: "Nina", type: person }
  - { name: "Core Platform", type: product }
---

## Summary

Security audit closed with 3 mediums + 1 high. Remediation plan and
ownership locked.

## Findings

| # | Severity | Area                  | Owner | Due        |
|---|----------|-----------------------|-------|------------|
| 1 | HIGH     | ACL propagation bug   | Carol | 2026-04-17 |
| 2 | medium   | Session handling      | Bob   | 2026-04-22 |
| 3 | medium   | Input validation      | Bob   | 2026-04-22 |
| 4 | medium   | Rate limiting         | Bob   | 2026-04-22 |

## High — ACL propagation

Chunk-level ACL intersection wasn't propagating through the wiki
citation layer — a cite from a less-restrictive audience could leak
the restricted chunk's body. Fix: propagate the most-restrictive ACL
up the citation chain at render time.

## Mediums

- **Session handling**: idle sessions didn't expire the token; fixed
  by binding token TTL to session-last-seen.
- **Input validation**: parse layer accepted pathological inputs
  (3 MB blobs tagged as markdown); add a size ceiling per genre.
- **Rate limiting**: ratelimit bucket wasn't keyed on tenant, only on
  connector kind; cross-tenant noisy-neighbor risk on shared adapters.

## Verification

Nina verifies each fix before close-out. Reality check + focused
property tests for the ACL propagation case.

## Revisit

Re-audit at Q3 — same scope + the new connector credentials path.
