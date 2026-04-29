---
slug: core-platform-architecture
title: Core platform architecture
audience: engineering
version: 1
last_curated: 2026-04-15T12:00:00Z
curated_by: deterministic:sample
---

## Summary

Per-tenant microVM isolation on the compute plane; proxy-fronted
routing; in-VM daemon (`envd`) as the only ingress; warm pools per
tier. Goal: p95 cold-start under 100 ms
{{cite: optimal://nodes/02-platform/signals/2026-04-10-microvm-spec.md}}.

## Architecture

```
Client  →  Proxy (TLS, tenant routing)  →  microVM[:envd]
                          │
                          └→  Warm pool (per-tier quota)
```

- One microVM per active tenant session.
- Proxy terminates TLS; intra-VM traffic is plain HTTP over the
  local socket.
- Warm pool tiers: free=0, pro=5, business=25.

## Acceptance Criteria

- Cold-start p95 ≤ 100 ms (rolling 10-minute window).
- Zero cross-tenant leakage under load
  (`mix optimal.reality_check --hard` gates every merge).
- Every tenant VM accepts only the `envd` socket — no SSH, no debug
  endpoints, no alternate ingress.

## Related

- Recent security audit surfaced one high (ACL propagation) + three
  medium findings — remediation in progress
  {{cite: optimal://nodes/02-platform/signals/2026-04-12-security-audit.md}}.
- New signal modalities (time-series, geo) land in two sprints
  {{cite: optimal://nodes/02-platform/signals/2026-04-15-data-arch-review.md}}.
- Pricing model depends on this architecture's unit economics — see
  [[healthtech-pricing-decision]].
