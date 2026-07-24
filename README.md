# Atlas

Enterprise control plane for **1C** (1С:Предприятие) infrastructure — inventory, discovery, agent-based monitoring, and orchestrated operations across Windows and Linux environments.

## Goals

- Maintain a single, authoritative inventory of 1C hosts, clusters, and infobases
- Automate discovery and reconciliation with minimal manual input
- Enable audited, orchestrated bootstrap and maintenance workflows
- Operate consistently across Windows and Linux with enterprise-grade security

## Non-Goals

- 1C application configuration (metadata, extensions, business logic)
- Replacement of native 1C administration tools for deep debugging
- End-user licensing or billing portals
- Multi-tenant SaaS in MVP (on-premises first)

Details: [Vision](docs/vision/vision.md)

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend services | .NET 8 |
| Atlas Console | TypeScript, React |
| Agents | .NET 8 (Windows), .NET 8 / Go (Linux — TBD in Phase 2) |
| Database | PostgreSQL 16 |
| Message broker | RabbitMQ 3.x |
| API contract | REST, OpenAPI 3.x |
| CI/CD | GitHub Actions |
| Runtime packaging | Docker (services), OS-native packages (agents) |

Rationale and constraints: [Architecture — Technology Stack](docs/architecture/architecture.md#technology-stack)

## Documentation

| Document | Purpose |
|----------|---------|
| [Vision](docs/vision/vision.md) | Problem statement, personas, success criteria |
| [Roadmap](docs/vision/roadmap.md) | MVP delivery phases |
| [Architecture](docs/architecture/architecture.md) | Components, diagrams, messaging, security |
| [Development Workflow](docs/architecture/development-workflow.md) | Branching, PRs, ADRs, CI |
| [ADR-0001](docs/adr/ADR-0001-project-structure.md) | Monorepo structure decision |

## Development Workflow

1. Branch from `main` using `feature/<ticket>-<short-description>` or `fix/<ticket>-<short-description>`
2. Keep changes scoped; update documentation when behavior or structure changes
3. Open a pull request with linked issue; require one reviewer approval
4. Record significant design decisions as ADRs in `docs/adr/`
5. Merge via squash after CI passes

Full process: [Development Workflow](docs/architecture/development-workflow.md)

## Repository Layout

```
backend/   frontend/   agent/{windows,linux}/   docs/   scripts/   .github/workflows/
```

Structure rationale: [ADR-0001](docs/adr/ADR-0001-project-structure.md)

## Status

Phase 0 (Foundation) — repository structure and documentation baseline. Application code begins in Phase 1. See [Roadmap](docs/vision/roadmap.md).

## License

TBD
