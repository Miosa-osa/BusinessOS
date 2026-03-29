# BusinessOS — Diátaxis Documentation

> **Fortune 500-grade AI business operating system.**
>
> Diátaxis documentation for BusinessOS — tutorials, how-to guides, explanations, and reference.

---

## Recent Updates (2026-03-27)

- **BOS Gateway Pattern** — New how-to for routing requests to pm4py-rust, Canopy, and OSA
- **Circuit Breaker Reference** — New configuration reference for compliance circuit breakers
- **MCP Server Integration** — New how-to for connecting stdio/HTTP/SSE MCP servers
- **OpenTelemetry Spans** — New how-to for adding custom OTEL instrumentation
- **Compliance Rules** — Expanded to cover GDPR + SOX (was HIPAA-only)
- **API Endpoints** — Updated with BOS gateway and A2A routes

---

## About BusinessOS

BusinessOS is a self-hosted AI business OS built on the MIOSA four-layer stack. It gives your AI agents a home — with projects, clients, documents, team structure, and the full context of your operation.

**Tech Stack**: Go 1.24 backend + SvelteKit 2 frontend + Electron desktop

**Integration**: Uses Signal Theory, YAWL patterns, and Data Operating Standard from the root theory.

---

## Diátaxis Documentation

### [Tutorials](../../docs/diataxis/tutorials/) — Learn by Doing

| Tutorial | What You'll Learn | Time |
|----------|-------------------|------|
| [BusinessOS Desktop Deployment](../../docs/diataxis/tutorials/businessos-deploy.md) | Deploy self-hosted BusinessOS with Docker | 40 min |
| [Your First AI Operation](../../docs/diataxis/tutorials/first-operation.md) | Build Canopy workspace (foundation for BusinessOS ops) | 30 min |
| [First API Call](./tutorial-first-api-call.md) | Call a BusinessOS REST endpoint end-to-end | 15 min |
| [First Frontend Component](./tutorial-first-frontend-component.md) | Build and render a SvelteKit component | 15 min |
| [First Database Record](./tutorial-first-database-record.md) | Write and query a PostgreSQL record via sqlc | 20 min |

### [How-to Guides](../../docs/diataxis/how-to/) — Solve Problems

**Root how-to guides:**

| Guide | Solves | Complexity |
|-------|--------|------------|
| [Add S/N Quality Gates](../../docs/diataxis/how-to/add-quality-gates.md) | Reject low-quality agent output | Intermediate |
| [Create Client Communications](../../docs/diataxis/how-to/client-communications.md) | Signal-encoded client emails, proposals | Intermediate |
| [Deploy Multi-OS Fleet](../../docs/diataxis/how-to/multi-os-fleet.md) | Multiple BusinessOS instances | Expert |
| [Integrate Data Operating Standard](../../docs/diataxis/how-to/data-operating-standard.md) | SDK-backed data operations | Advanced |

**BusinessOS-specific how-to guides:**

| Guide | Solves | Complexity |
|-------|--------|------------|
| [Add a New API Endpoint](./how-to/add-api-endpoint.md) | Create REST endpoint with handler, route, and database query | Beginner |
| [Debug a Frontend Issue](./how-to/debug-frontend-issue.md) | Find and fix rendering, state, or network problems in SvelteKit | Intermediate |
| [Add a Feature Behind a Feature Flag](./how-to/feature-flag-rollout.md) | Roll out new feature gradually with zero downtime | Intermediate |
| [Artifact Construct Handler](./how-to/artifact-construct-handler.md) | Build an artifact CONSTRUCT query handler in Go | Intermediate |
| [FIBO Deal Integration Setup](./how-to/fibo-deal-integration-setup.md) | Wire FIBO ontology deal objects into BusinessOS | Advanced |
| [FIBO Deals Go Implementation](./how-to/fibo-deals-go-implementation.md) | Implement FIBO deal types in Go with full type safety | Advanced |
| [Query Ontology](./how-to/query-ontology.md) | Run SPARQL queries against the BusinessOS ontology layer | Intermediate |
| [SOX Audit Trail Integration](./how-to/sox-audit-trail-integration.md) | Persist SOX-compliant audit trails via PROV-O | Advanced |
| [BOS Gateway Pattern](./how-to/bos-gateway-pattern.md) | Route requests to pm4py-rust, Canopy, and OSA via the BOS gateway | Intermediate |
| [MCP Server Integration](./how-to/mcp-server-integration.md) | Connect stdio/HTTP/SSE MCP servers to BusinessOS | Intermediate |
| [OpenTelemetry Spans](./how-to/opentelemetry-spans.md) | Add custom OTEL spans to handlers and services | Intermediate |

### [Explanation](../../docs/diataxis/explanation/) — Understand the System

| Explanation | Topic | Why It Matters |
|-------------|-------|----------------|
| [The Chatman Equation](../../docs/diataxis/explanation/chatman-equation.md) | A=μ(O) foundation | BusinessOS applies this to business operations |
| [Signal Theory Complete](../../docs/diataxis/explanation/signal-theory-complete.md) | S=(M,G,T,F,W) encoding | Client comms use this |
| [The 7-Layer Architecture](../../docs/diataxis/explanation/seven-layer-architecture.md) | Optimal Systems design | BusinessOS implements all 7 layers |
| [FIBO Deal Integration](./explanation/fibo-deal-integration-complete.md) | FIBO ontology in Go | How deal objects map to the FinancialInstrument ontology |

### [Reference](../../docs/diataxis/reference/) — Look Up Details

| Reference | Covers | Format |
|-----------|--------|--------|
| [Signal Format](../../docs/diataxis/reference/signal-format.md) | S=(M,G,T,F,W) spec | BNF grammar |
| [API Endpoints](./reference/api-endpoints.md) | All REST/WebSocket/SSE APIs including BOS gateway and A2A routes | OpenAPI specs |
| [Configuration Options](./reference/configuration-options.md) | All config keys and environment variables | YAML reference |
| [Error Codes](./reference/error-codes.md) | All HTTP and business-logic error codes | Table |
| [Database Schema](./reference/database-schema.md) | PostgreSQL tables, indexes, and relations | ERD + SQL |
| [Artifact Construct Queries](./reference/artifact-construct-queries.md) | SPARQL CONSTRUCT queries for artifacts | SPARQL reference |
| [FIBO Deal Quick Reference](./reference/fibo-deal-quick-reference.md) | FIBO deal types, attributes, and mappings | Cheat sheet |
| [HIPAA Compliance Validator](./reference/hipaa-compliance-validator.md) | HIPAA rule identifiers and validation logic | Rule table |
| [Circuit Breaker Configuration](./reference/circuit-breaker-configuration.md) | Circuit breaker states, thresholds, and compliance integration | Configuration reference |
| [YAWL v6 Integration](./reference/yawl-integration.md) | YAWL engine endpoints, auth, configuration, and health.jsp workaround | API reference |
| [Genre Catalog](../../docs/diataxis/reference/genre-catalog.md) | All Signal Theory genres | Usage guide |

---

## BusinessOS-Specific Documentation

### Core Modules

| Module | Diátaxis Docs | BusinessOS Docs |
|--------|---------------|-----------------|
| **Dashboard** | [Signal Theory](../../docs/diataxis/explanation/signal-theory-complete.md) | [Dashboard Guide](../modules/dashboard.md) |
| **Projects** | [YAWL Patterns](../../docs/diataxis/reference/yawl-43-patterns.md) | [Projects Module](../modules/projects.md) |
| **Tasks** | [How-to: Agent Handoffs](../../docs/diataxis/how-to/agent-handoffs.md) | [Tasks Module](../modules/tasks.md) |
| **AI Chat** | [Tutorial: Signal Theory](../../docs/diataxis/tutorials/signal-theory-practice.md) | [AI Chat Module](../modules/ai-chat.md) |
| **Clients** | [How-to: Client Communications](../../docs/diataxis/how-to/client-communications.md) | [CRM Module](../modules/clients.md) |
| **Documents** | [Explanation: Ontology Closure](../../docs/diataxis/explanation/ontology-closure.md) | [Documents Module](../modules/documents.md) |
| **OSA Integration** | [OSA Diátaxis Docs](../../OSA/docs/diataxis/README.md) | [OSA Module](../osa/) |

### Development

| Topic | Diátaxis Docs | Local Docs |
|-------|---------------|------------|
| **Backend** | [The Chatman Equation](../../docs/diataxis/explanation/chatman-equation.md) | [Go Backend Guide](../development/BACKEND.md) |
| **Frontend** | [Signal Theory](../../docs/diataxis/explanation/signal-theory-complete.md) | [SvelteKit Guide](../development/FRONTEND.md) |
| **Desktop** | [Tutorial: BusinessOS Deploy](../../docs/diataxis/tutorials/businessos-deploy.md) | [Electron Guide](../desktop/README.md) |
| **Testing** | [How-to: Add Quality Gates](../../docs/diataxis/how-to/add-quality-gates.md) | [Testing Guide](../development/TESTING.md) |

---

## AGI-Level Connections

### Signal Theory in BusinessOS

BusinessOS uses Signal Theory S=(M,G,T,F,W) for:

| Use Case | Signal Encoding |
|----------|-----------------|
| **Client Emails** | `S=(linguistic, email, direct, markdown, cold-email-anatomy)` |
| **Proposals** | `S=(linguistic, proposal, direct, markdown, persuasion)` |
| **Reports** | `S=(mixed, report, inform, markdown, dashboard-metrics)` |
| **Code Gen** | `S=(code, implementation, direct, typescript, module-pattern)` |

### YAWL Patterns in BusinessOS

BusinessOS workflows use YAWL patterns:

| Pattern | BusinessOS Usage |
|---------|------------------|
| **Parallel Split** | Multi-module updates |
| **Synchronization** | Multi-approval workflows |
| **Multi-Choice** | Dynamic agent activation |
| **Cancel Region** | Deployment rollback |

### Data Operating Standard

BusinessOS enforces SDK-backed data operations:

| Operation | SDK Method |
|-----------|------------|
| **Schema work** | `data-modelling-cli import/export` |
| **Decision records** | `bos decisions new` |
| **Knowledge articles** | `bos knowledge index` |
| **Data contracts** | `bos workspace init` |
| **Audit trail** | PROV-O via CONSTRUCT into Oxigraph |

---

## Quick Start

### For New Users

1. **Deploy BusinessOS**: [Tutorial](../../docs/diataxis/tutorials/businessos-deploy.md)
2. **Learn Signal Theory**: [Tutorial](../../docs/diataxis/tutorials/signal-theory-practice.md)
3. **Understand the System**: [7-Layer Architecture](../../docs/diataxis/explanation/seven-layer-architecture.md)

### For Developers

1. **Add API Endpoint**: [How-To Guide](./how-to/add-api-endpoint.md)
2. **BOS Gateway**: [How-To Guide](./how-to/bos-gateway-pattern.md)
3. **MCP Integration**: [How-To Guide](./how-to/mcp-server-integration.md)
4. **OTEL Instrumentation**: [How-To Guide](./how-to/opentelemetry-spans.md)
5. **Debug Frontend**: [How-To Guide](./how-to/debug-frontend-issue.md)
6. **Feature Flags**: [How-To Guide](./how-to/feature-flag-rollout.md)
7. **Backend**: [Go Backend Guide](../development/BACKEND.md)
8. **Frontend**: [SvelteKit Guide](../development/FRONTEND.md)
9. **Data Operations**: [Data Operating Standard Guide](../../docs/diataxis/how-to/data-operating-standard.md)

### For Operators

1. **Deploy**: [Tutorial](../../docs/diataxis/tutorials/businessos-deploy.md)
2. **Configure**: [Configuration Options](./reference/configuration-options.md)
3. **Scale**: [Multi-OS Fleet](../../docs/diataxis/how-to/multi-os-fleet.md)
4. **Circuit Breakers**: [Circuit Breaker Configuration](./reference/circuit-breaker-configuration.md)

---

## Cross-Project Documentation

BusinessOS-specific docs are here. For cross-project architecture and integration docs, see the parent:

- [Signal Theory S=(M,G,T,F,W)](../../docs/diataxis/explanation/signal-theory-complete.md)
- [A2A Protocol Reference](../../docs/diataxis/reference/a2a-protocol-api-reference.md)
- [YAWL 43 Patterns](../../docs/diataxis/reference/yawl-43-patterns.md)
- [Cross-Project Integration Chain](../../docs/diataxis/explanation/five-project-integration-chain.md)

---

## Cross-Project Links

- **Root Diátaxis**: [Main Documentation](../../docs/diataxis/README.md)
- **Canopy**: [Canopy Diátaxis](../../canopy/docs/diataxis/README.md)
- **OSA**: [OSA Diátaxis](../../OSA/docs/diataxis/README.md)

---

*BusinessOS Diátaxis Documentation — Part of the ChatmanGPT Knowledge System*
*Version 1.1.0 — Updated: 2026-03-27*
