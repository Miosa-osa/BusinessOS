---
title: Core platform — microVM isolation spec
genre: spec
mode: linguistic
node: 02-platform
authored_at: 2026-04-10T10:00:00Z
sn_ratio: 0.85
entities:
  - { name: "Bob", type: person }
  - { name: "Carol", type: person }
  - { name: "Dan", type: person }
  - { name: "Core Platform", type: product }
  - { name: "microVM", type: concept }
  - { name: "in-VM daemon", type: concept }
---

## Summary

Per-tenant microVM isolation on the core compute plane. Each tenant
gets an isolated VM; an in-VM daemon runs inside. Dan owns the proxy
layer.

## Goal

Give every tenant strong process + network isolation without paying
full OS-level virtualization overhead. Target: VM cold-start under
100 ms, warm pool for SLA tiers.

## Requirements

1. One microVM per active tenant session. No shared kernel across
   tenants.
2. An in-VM daemon (`envd`) exposing a local agent API over a unix
   socket inside the VM.
3. Proxy layer in front routes `(tenant_id, session_id)` →
   appropriate VM. Zero trust across the proxy boundary.
4. Cold-start target: p95 under 100 ms from proxy request.
5. Warm-pool policy per tier: free = no pool, pro = 5, business = 25.

## Constraints

- No customer workload sees a kernel shared with another customer.
- The in-VM daemon's API is the ONLY ingress into user VMs — no SSH,
  no debug ports, no shell backdoors.
- Proxy terminates TLS; VMs receive plain HTTP over the tenant-local
  socket.

## Architecture

```
  Client  →  Proxy (TLS, tenant routing)  →  microVM[:envd]
                              │
                              └→  Warm pool (per-tier quota)
```

## Acceptance criteria

- Reality-check probe: 100 concurrent tenants × 50 requests/sec =
  zero cross-tenant leakage under ACL intersection.
- `mix optimal.reality_check --hard` passes.
- Cold-start p95 ≤ 100 ms measured over 10-minute rolling window.

## Owners

- Bob: spec approval + review.
- Carol: microVM orchestration.
- Dan: proxy + routing.
