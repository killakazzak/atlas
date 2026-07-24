# ADR-0001: Monorepo Project Structure

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-07-24 |
| **Deciders** | Architecture team |
| **Related** | [Architecture](../architecture/architecture.md), [Roadmap](../vision/roadmap.md) |

## Context

Atlas spans backend services, a web console, OS-specific agents, and operational tooling. The team requires a repository layout that:

- Separates runtime targets (server, browser, Windows, Linux)
- Co-locates documentation and ADRs with implementation
- Supports phased MVP delivery without premature multi-repo overhead

## Decision

Adopt a **monorepo** with platform boundaries at the top level:

```
atlas/
├── backend/              # Platform services (Inventory, Discovery, Bootstrap, Agent, Orchestrator)
├── frontend/             # Atlas Console
├── agent/
│   ├── windows/
│   └── linux/
├── docs/                 # vision/, architecture/, api/, database/, diagrams/, sprint/, adr/
├── scripts/
├── .github/workflows/
└── README.md
```

### Rationale

| Path | Why separate |
|------|--------------|
| `backend/` | Shared deployment unit; services subdivide internally as code grows |
| `frontend/` | Independent build and release cadence from server artifacts |
| `agent/windows/`, `agent/linux/` | Different packaging (MSI vs DEB/RPM), OS APIs, CI matrix |
| `docs/` | Canonical documentation hub; survives backend refactors |
| `scripts/` | Operational automation outside application boundaries |

Full mapping to architectural components: [Architecture — Repository Mapping](../architecture/architecture.md#repository-mapping).

## Alternatives Considered

| Option | Verdict |
|--------|---------|
| Multi-repo (one repo per service) | Rejected for MVP — coordination overhead, shared types friction |
| Flat `src/` without agent split | Rejected — obscures OS-specific build pipelines |
| External wiki for docs | Rejected — docs drift from code; ADRs lose version correlation |

## Consequences

**Positive:** clear onboarding, path-scoped CI, ADRs versioned with code, roadmap phases map to directories.

**Negative:** repository growth requires CI path filtering; `backend/` will need internal module layout (follow-up ADR in Phase 1).

**Neutral:** empty directories use `.gitkeep` until Phase 1 scaffolding.

## Compliance

- Existing files preserved during initial structure creation
- Application code deferred per [Roadmap — Phase 0](../vision/roadmap.md#phase-0--foundation)

## Follow-up

| ADR | Trigger |
|-----|---------|
| ADR-0002 | Backend service module layout (Phase 1 start) |
| ADR-0003 | Agent communication protocol (Phase 2 start) |
