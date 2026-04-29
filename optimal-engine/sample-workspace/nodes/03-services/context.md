---
node: 03-services
kind: org
style: internal
---

# Services Division — context

Revenue + delivery side of the company. Contracts, invoices, managed
engagements.

## What lives here

- Customer invoices (`signals/YYYY-MM-DD-invoice-<id>.md`)
- Statements of work
- Delivery runbooks per engagement
- Retainer cadence + renewal dates

## Team

| Role                | Person |
|---------------------|--------|
| Services lead       | Eve    |
| Partner success     | Dan    |
| Legal               | Heidi  |

## Operating rules

- Every invoice has a corresponding signal file. The file is the
  source of truth; the accounting system is a mirror.
- Renewal dates are tracked in the `renewals` column of the retainer
  frontmatter; the retention sweep surfaces anything inside a 30-day
  window.
