---
node: 04-academy
kind: operation
style: internal
---

# Customer Academy — context

Customer onboarding + ongoing education. Runs the beginner + advanced
tracks, the premium tier, and weekly office hours.

## Tracks

- `04-academy-beginner` — 12-week curriculum, weekly live session,
  asynchronous labs.
- `04-academy-advanced` — 8 hands-on labs covering the full pipeline
  (ingest → classify → embed → retrieve → curate).
- Premium tier ($10K) — personal onboarding for enterprise tenants.

## Team

| Role            | Person |
|-----------------|--------|
| Academy lead    | Judy   |
| Co-teacher      | Alice  |
| Advanced author | Bob    |
| Labs author     | Carol  |
| Pipeline labs   | Ivan   |

## Metrics (rolling 30-day)

- NPS: 54 (target 50+)
- Retention: 91%
- Expansion revenue: 38% of customer base

## Operating rules

- Every live session produces a `signals/YYYY-MM-DD-session-<topic>.md`
  transcript within 24 h.
- Customer feedback from office hours flows into
  `signals/YYYY-MM-DD-office-hours-*.md`; concept gaps feed back into
  the curriculum.
- No marketing copy lives in this node; that goes to `08-media`.
