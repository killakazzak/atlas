# Atlas

[![CI](https://github.com/killakazzak/atlas/actions/workflows/ci.yml/badge.svg)](https://github.com/killakazzak/atlas/actions/workflows/ci.yml)

Enterprise control plane for **1C** (1С:Предприятие) infrastructure — inventory, discovery, agent-based monitoring, and orchestrated operations across Windows and Linux environments.

---

## Overview

Atlas provides a single authoritative source of truth for 1C infrastructure: hosts, clusters, infobases, and agents. It exposes a REST API consumed by the Atlas Console (web UI) and the Atlas Agent running on managed hosts.

## Features

- Inventory management for servers, clusters, and infobases
- REST API with structured error responses and request tracing
- OpenAPI 3.1 specification with embedded Swagger UI
- PostgreSQL persistence with versioned migrations
- Middleware: panic recovery, request ID propagation, structured logging
- Unit and integration tests
- CI pipeline with formatting, vet, lint, and OpenAPI validation

## Architecture

| Layer | Technology |
|-------|------------|
| Backend | Go 1.25 |
| Console | TypeScript, React (Phase 2) |
| Agents | Go — Windows and Linux (Phase 2) |
| Database | PostgreSQL 17 |
| Message broker | RabbitMQ 3.x (Phase 2) |
| API contract | REST, OpenAPI 3.1 |
| CI/CD | GitHub Actions |
| Packaging | Docker (services), OS-native packages (agents) |

Full design: [Architecture](docs/architecture/architecture.md)

## Requirements

- [Go 1.25+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for local linting)
- [vacuum](https://quobix.com/vacuum/) (for local OpenAPI validation)

## Quick Start

```bash
git clone https://github.com/killakazzak/atlas.git
cd atlas
```

```powershell
.\scripts\db-up.ps1
.\scripts\migrate-up.ps1
.\scripts\run-local.ps1
```

The server starts at `http://localhost:8080`.

## Running Locally

| Script | Purpose |
|--------|---------|
| `.\scripts\db-up.ps1` | Start PostgreSQL and wait until ready |
| `.\scripts\db-down.ps1` | Stop PostgreSQL |
| `.\scripts\migrate-up.ps1` | Apply all pending migrations |
| `.\scripts\migrate-down.ps1` | Roll back one migration step |
| `.\scripts\run-local.ps1` | Start the Atlas server |

`DATABASE_URL` defaults to `postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable`.
Set it in your environment to override.

> Migrations are **not** applied automatically on server start. Always run `migrate-up.ps1` after pulling new migrations.

## Database

PostgreSQL 17 is managed via Docker Compose. The container is defined in `docker-compose.yml` at the repository root.

Migrations live in `backend/migrations/` and follow the [golang-migrate](https://github.com/golang-migrate/migrate) naming convention:

```
000001_create_servers.up.sql
000001_create_servers.down.sql
```

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

All error responses follow a consistent structure:

```json
{
  "error": {
    "code": "not_found",
    "message": "server not found"
  },
  "requestId": "a3f1b2c4d5e6f708"
}
```

## Swagger

Interactive API documentation is available when the server is running:

| Resource | URL |
|----------|-----|
| Swagger UI | http://localhost:8080/swagger |
| OpenAPI spec | http://localhost:8080/openapi/openapi.yaml |

The specification source is at `backend/openapi/openapi.yaml`.

## Development

Run all checks from the `backend/` directory:

```bash
go test ./...
go vet ./...
gofmt -l .
golangci-lint run
```

Or using Make (requires `make` on PATH):

```bash
cd backend
make check   # fmt + vet + lint + test + openapi
make test
make run
```

**Git workflow**

1. Branch from `main`: `feature/<ticket>-<slug>` or `fix/<ticket>-<slug>`
2. Keep changes scoped; update docs when behavior or structure changes
3. Open a pull request with a linked issue; require one reviewer approval
4. Record significant design decisions as ADRs in `docs/adr/`
5. Merge via squash after CI passes

Full process: [Development Workflow](docs/architecture/development-workflow.md)

## CI

GitHub Actions runs on every push and pull request (Go 1.24, Ubuntu):

| Step | Tool |
|------|------|
| Formatting | `gofmt -l` — fails if any file needs formatting |
| Static analysis | `go vet ./...` |
| Linting | `golangci-lint` |
| Tests | `go test -race ./...` |
| OpenAPI validation | `vacuum lint` — fails if spec is invalid |

Pipeline config: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

## Project Structure

```
atlas/
├── backend/                   # Go backend service
│   ├── cmd/server/            # Entrypoint
│   ├── internal/
│   │   ├── app/               # Application wiring
│   │   ├── config/            # Configuration
│   │   ├── database/          # PostgreSQL connection
│   │   ├── http/              # HTTP server and middleware
│   │   ├── httpx/             # Shared HTTP helpers
│   │   ├── inventory/         # Inventory domain and service
│   │   └── logger/            # Structured logger
│   ├── migrations/            # SQL migrations
│   ├── openapi/               # OpenAPI 3.1 spec + embed
│   ├── Makefile
│   └── .golangci.yml
├── frontend/                  # React console (Phase 2)
├── agent/                     # Atlas agent (Phase 2)
├── docs/                      # Architecture, ADRs, vision
├── scripts/                   # PowerShell dev scripts
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
