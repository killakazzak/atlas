# Atlas Vision

## Problem Statement

Enterprises operating 1C at scale manage hundreds of servers, multiple RAS clusters, and dozens of infobases across isolated environments. Operational data is scattered across spreadsheets, ad-hoc scripts, and generic monitoring tools that do not understand 1C topology.

This fragmentation causes stale inventory, slow incident response, unaudited changes, and inconsistent agent deployment across Windows and Linux.

## Vision

Atlas is the **operational control plane** for 1C infrastructure: one place to see what exists, how components relate, and how to change the environment safely.

## Goals

| # | Goal | Description |
|---|------|-------------|
| G1 | Single source of truth | Authoritative inventory of hosts, clusters, working processes, infobases, databases, and dependencies |
| G2 | Automated discovery | Detect new and changed infrastructure through agents and scheduled reconciliation |
| G3 | Safe automation | Bootstrap, rediscovery, and maintenance via audited orchestrated workflows |
| G4 | Platform parity | Consistent operational model on Windows and Linux |
| G5 | Enterprise readiness | RBAC, audit logging, HA deployment path, ITSM/monitoring integration hooks |

## Non-Goals

The following are explicitly **out of scope** for Atlas, including initial MVP releases:

| # | Non-Goal | Rationale |
|---|----------|-----------|
| NG1 | 1C metadata and extension management | Application lifecycle belongs in 1C tooling; Atlas manages infrastructure |
| NG2 | Deep 1C debugging and session administration | Native 1C consoles remain the tool of record for runtime diagnostics |
| NG3 | End-user licensing portal | License *inventory* may be tracked; provisioning and billing are separate products |
| NG4 | Generic CMDB replacement | Atlas integrates with CMDB sources; it is specialized for 1C infrastructure |
| NG5 | Multi-tenant SaaS (MVP) | Initial delivery targets on-premises or dedicated cloud; SaaS model deferred |

Post-MVP items (CMDB sync, advanced orchestration, infobase lifecycle) are tracked in the [Roadmap](roadmap.md#post-mvp).

## Target Users

| Persona | Primary Need |
|---------|--------------|
| 1C Administrator | Health visibility, configuration context, controlled changes |
| Infrastructure Engineer | Host enrollment, agent deployment, discovery configuration |
| Operations Manager | Environment overview, compliance, change audit trail |
| Release Manager | Infobase placement, environment promotion visibility |

## Success Criteria

| Metric | Target |
|--------|--------|
| Host onboarding time | Hours → minutes (via bootstrap workflow) |
| Inventory accuracy | ≥ 95% without manual reconciliation |
| Change traceability | 100% of production inventory mutations in audit log |
| Agent coverage | 100% of enrolled hosts reporting within one sprint |

## Related Documentation

- [Roadmap](roadmap.md) — phased delivery plan
- [Architecture](../architecture/architecture.md) — system design
- [Development Workflow](../architecture/development-workflow.md) — contribution process
