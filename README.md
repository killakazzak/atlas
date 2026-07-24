# Atlas

[![CI](https://github.com/killakazzak/atlas/actions/workflows/ci.yml/badge.svg)](https://github.com/killakazzak/atlas/actions/workflows/ci.yml)

Enterprise control plane for **1C** (1С:Предприятие) infrastructure — inventory, discovery, agent-based monitoring, and orchestrated operations across Windows and Linux environments.

---

## Overview

Atlas provides a single authoritative source of truth for 1C infrastructure: hosts, clusters, infobases, and agents. It exposes a REST API consumed by the Atlas Console (web UI) and the Atlas Agent running on managed hosts.

## Features

- Inventory management for servers, clusters, and infobases
- Agent-based monitoring with heartbeat and status tracking
- REST API with structured error responses and request tracing
- OpenAPI 3.1 specification with embedded Swagger UI
- PostgreSQL persistence with versioned migrations
- Middleware: panic recovery, request ID propagation, structured logging
- Unit and integration tests
- CI pipeline with formatting, vet, lint, and OpenAPI validation

## Architecture

| Layer | Technology |
|-------|------------|
| Backend | Go |
| Console | TypeScript, React |
| Agents | Go (Windows and Linux) |
| Database | PostgreSQL 17 |
| Message broker | RabbitMQ 3.x (Phase 2) |
| API contract | REST, OpenAPI 3.1 |
| CI/CD | GitHub Actions |
| Packaging | Docker (services), OS-native packages (agents) |

Full design: [Architecture](docs/architecture/architecture.md)

## Getting Started

**Prerequisites**

- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [Go 1.25+](https://go.dev/dl/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)

**Clone and enter the repository**

```bash
git clone https://github.com/killakazzak/atlas.git
cd atlas
```

## Running Locally

Start PostgreSQL, apply migrations, and run the server:

```powershell
.\scripts\db-up.ps1
.\scripts\migrate-up.ps1
.\scripts\run-local.ps1
```

The server starts on `http://localhost:8080`.

Stop PostgreSQL when done:

```powershell
.\scripts\db-down.ps1
```

> Migrations are **not** applied automatically on server start. Always run `migrate-up.ps1` after pulling new migrations.

## Database

Atlas uses PostgreSQL 17 managed via Docker Compose.

| Script | Purpose |
|--------|---------|
| `.\scripts\db-up.ps1` | Start PostgreSQL and wait until ready |
| `.\scripts\db-down.ps1` | Stop PostgreSQL |
| `.\scripts\migrate-up.ps1` | Apply all pending migrations |
| `.\scripts\migrate-down.ps1` | Roll back one migration step |

`DATABASE_URL` defaults to `postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable`.
Set it in your environment to override.

Migrations live in `backend/migrations/` and follow the `golang-migrate` naming convention.

## API

Base URL: `http://localhost:8080`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/version` | Service version |
| `GET` | `/api/v1/servers` | List servers |
| `POST` | `/api/v1/servers` | Register a server |
| `GET` | `/api/v1/servers/{id}` | Get a server |
| `DELETE` | `/api/v1/servers/{id}` | Delete a server |

All error responses use a consistent format:

```json
{
  "error": {
    "code": "not_found",
    "message": "server not found"
  },
  "requestId": "a3f1b2c4d5e6f708"
}
```

## Swagger UI

Interactive API documentation is available when the server is running:

| Resource | URL |
|----------|-----|
| Swagger UI | http://localhost:8080/swagger |
| OpenAPI spec (YAML) | http://localhost:8080/openapi/openapi.yaml |

The spec source lives at `backend/openapi/openapi.yaml`.

## Development

**Run all checks** (format, vet, lint, test, OpenAPI validation):

```bash
cd backend
make check
```

Individual targets:

```bash
make fmt      # format source files
make vet      # go vet
make lint     # golangci-lint
make test     # go test -race ./...
make openapi  # validate openapi.yaml with vacuum
make run      # start the server
```

**Workflow**

1. Branch from `main`: `feature/<ticket>-<short-description>` or `fix/<ticket>-<short-description>`
2. Keep changes scoped; update docs when behavior or structure changes
3. Open a pull request; require one reviewer approval
4. Record significant design decisions as ADRs in `docs/adr/`
5. Merge via squash after CI passes

Full process: [Development Workflow](docs/architecture/development-workflow.md)

## CI

GitHub Actions runs on every push and pull request:

| Step | Tool |
|------|------|
| Formatting | `gofmt` |
| Static analysis | `go vet` |
| Linting | `golangci-lint` |
| Tests | `go test -race` |
| OpenAPI validation | `vacuum` |

Pipeline config: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

## Project Structure

```
atlas/
├── backend/                  # Go backend service
│   ├── cmd/server/           # Entrypoint
│   ├── internal/
│   │   ├── app/              # Application wiring
│   │   ├── config/           # Configuration
│   │   ├── database/         # PostgreSQL connection
│   │   ├── http/             # HTTP server and middleware
│   │   ├── httpx/            # Shared HTTP helpers
│   │   ├── inventory/        # Inventory domain and service
│   │   └── logger/           # Structured logger
│   ├── migrations/           # SQL migrations
│   ├── openapi/              # OpenAPI 3.1 spec + embed
│   ├── Makefile
│   └── .golangci.yml
├── frontend/                 # React console (Phase 2)
├── agent/                    # Atlas agent (Phase 2)
├── docs/                     # Architecture, ADRs, vision
├── scripts/                  # PowerShell dev scripts
├── docker-compose.yml
└── .github/workflows/ci.yml
```

Structure rationale: [ADR-0001](docs/adr/ADR-0001-project-structure.md)

## Roadmap

| Phase | Description | Status |
|-------|-------------|--------|
| 0 | Foundation — repository, docs, CI | ✅ Done |
| 1 | REST API, PostgreSQL, OpenAPI, Swagger | ✅ Done |
| 2 | Discovery, agents, RabbitMQ | Planned |
| 3 | Console (React UI) | Planned |
| 4 | Authentication, RBAC | Planned |

Full roadmap: [Roadmap](docs/vision/roadmap.md)

## License

Apache-2.0
