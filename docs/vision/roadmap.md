# Atlas Roadmap

Phased delivery from foundation to production-ready MVP. Each phase produces a demonstrable increment. Phases are **sequentially dependent**; work within a phase (e.g., Windows vs. Linux agent) may run in parallel.

Current status: **Phase 0**.

## Phase 0 — Foundation

**Objective:** Repository, documentation baseline, and engineering conventions.

| Deliverable | Owner area |
|-------------|------------|
| Monorepo structure | `backend/`, `frontend/`, `agent/`, `docs/`, `scripts/` |
| Vision, architecture, ADR-0001 | `docs/` |
| Development workflow | `docs/architecture/development-workflow.md` |
| CI skeleton | `.github/workflows/` |

**Exit criteria:** Contributors can clone the repo, navigate documentation, and understand the delivery plan.

---

## Phase 1 — Core Platform (MVP-1)

**Objective:** Manual inventory management via API and read-only Console.

| Deliverable | Component |
|-------------|-----------|
| Inventory Service REST API | `backend/` |
| PostgreSQL schema v1 | `docs/database/` |
| Atlas Console — host and cluster views | `frontend/` |
| OpenAPI specification | `docs/api/` |
| OIDC authentication stub | `backend/`, `frontend/` |

**Exit criteria:** Operator registers a host and 1C cluster via API or UI; data persists and displays in Console.

---

## Phase 2 — Agents & Discovery (MVP-2)

**Objective:** Automated visibility through endpoint agents.

| Deliverable | Component |
|-------------|-----------|
| Linux agent — heartbeat, 1C facts, registration | `agent/linux/` |
| Windows agent — feature parity | `agent/windows/` |
| Agent Service | `backend/` |
| Discovery Service — agent-driven reconciliation | `backend/` |
| RabbitMQ event pipeline | Discovery → Inventory |
| ADR-0003 — agent communication protocol | `docs/adr/` |

**Exit criteria:** Agent on a test host appears in Console within 5 minutes; discovered components merge into inventory without manual entry.

---

## Phase 3 — Bootstrap & Orchestration (MVP-3)

**Objective:** Guided onboarding and multi-step operational workflows.

| Deliverable | Component |
|-------------|-----------|
| Bootstrap Service — tokens, packages, validation | `backend/` |
| Orchestrator — linear workflows | `backend/` |
| Console bootstrap wizard and progress UI | `frontend/` |
| Audit log for mutations and workflow actions | `backend/`, PostgreSQL |

**Exit criteria:** New host onboarded end-to-end through Console without direct API calls.

---

## Phase 4 — Production Hardening (MVP-4)

**Objective:** Pilot-ready deployment in an enterprise environment.

| Deliverable | Component |
|-------------|-----------|
| RBAC (`viewer`, `operator`, `admin`) | `backend/`, `frontend/` |
| Observability baseline — logs, metrics, health probes | All services |
| HA deployment guide | `docs/architecture/` |
| Backup and recovery runbook | `scripts/`, `docs/` |
| Scale validation — 500 hosts, 2000 infobases | Test environment |

**Exit criteria:** Security review complete; pilot deployed; operations runbook delivered.

---

## Post-MVP

Tracked for prioritization after Phase 4; not committed to timeline:

- CMDB and ITSM integrations (ServiceNow, Jira Service Management)
- Parallel orchestration steps, approval gates, rollback
- Infobase lifecycle (scheduled backup, environment copy)
- License usage analytics
- Multi-tenant SaaS deployment model

Sprint-level planning artifacts: `docs/sprint/`.

## Related Documentation

- [Vision](vision.md) — goals and non-goals
- [Architecture](../architecture/architecture.md) — components and technology stack
- [Development Workflow](../architecture/development-workflow.md) — contribution process
