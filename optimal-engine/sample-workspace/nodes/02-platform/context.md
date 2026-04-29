---
node: 02-platform
kind: org
style: internal
---

# Platform Division — context

The engineering organization responsible for the core platform, managed
services, and investor-facing technical materials.

## Sub-nodes

- `02-platform-core` — core compute + data plane
- `02-platform-services` — managed-services offering (sales + delivery)
- `02-platform-investors` — deck + diligence materials

## Team

| Role                  | Person |
|-----------------------|--------|
| Technical partner     | Bob    |
| Platform engineer     | Carol  |
| Infra + proxy         | Dan    |
| Pipeline + embeddings | Ivan   |
| Media / DevRel        | Frank  |
| Legal + compliance    | Heidi  |
| SRE / audit           | Nina   |

## Architecture north stars

1. **Per-tenant microVM isolation** on the compute plane — each tenant
   gets an isolated VM; an in-VM daemon handles the local agent API.
   Proxy layer sits in front of all VMs.
2. **Multi-modal aligned embeddings** — text + vision in one shared
   768-d space; adds audio + video in the next release.
3. **Wiki-first retrieval** — Tier 3 curation consulted before hybrid
   BM25+vector+graph.

## Operating rules

- Every schema change ships with a versioned migration + a reality-check
  probe.
- Nothing hits main without a green `mix optimal.reality_check --hard`.
- New data-point architectures start with a YAML in
  `architectures/` + a processor binding; no implicit modalities.
