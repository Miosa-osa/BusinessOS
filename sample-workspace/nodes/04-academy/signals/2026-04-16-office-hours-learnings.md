---
title: Office hours — week 4 learnings
genre: note
mode: linguistic
node: 04-academy
authored_at: 2026-04-16T17:00:00Z
sn_ratio: 0.7
entities:
  - { name: "Judy", type: person }
  - { name: "Alice", type: person }
  - { name: "Customer Academy", type: operation }
  - { name: "S/N ratio", type: concept }
---

## Summary

Week 4 office hours surfaced two consistent concept gaps — chunk-level
ACLs and the S/N ratio. Curriculum patch coming next week.

## Key points

- Students struggle with **chunk-level ACLs** — they assume signals
  are ACL'd, not chunks. Need an explicit 10-minute lecture.
- **S/N ratio** is confused with "quality score" — it's really a noise
  filter boost. Add a worked example.
- Two repeat attendees asked about custom architectures — they're
  ready for advanced track content earlier than expected.

## Proposed curriculum changes

- Insert a focused module between weeks 8 (wiki) and 9 (ACLs):
  "ACLs at chunk vs. signal level — what the intersection propagation
  actually does."
- Add an S/N worked example in week 3: two near-identical signals,
  one with sn_ratio=0.3, one with 0.85 — show retrieval ranking
  difference.

## Follow-ups

| Who   | Task                                     | Due        |
|-------|------------------------------------------|------------|
| Judy  | Draft ACL explainer slides               | 2026-04-22 |
| Alice | Record the S/N worked example            | 2026-04-22 |
| Judy  | Sort "advanced-ready" beginners          | 2026-04-20 |
