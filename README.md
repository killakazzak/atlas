# Atlas

[![CI](https://github.com/killakazzak/atlas/actions/workflows/ci.yml/badge.svg)](https://github.com/killakazzak/atlas/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

Atlas is a backend platform for managing 1C (1С:Предприятие) infrastructure.
It provides inventory management, JWT authentication, role-based access control,
and a REST API documented with OpenAPI 3.1. The project is written in Go using
a modular monolith architecture and is designed as a foundation for automation,
monitoring, and orchestration of enterprise 1C environments.

---

## Features

- **Inventory management** — CRUD for managed servers
- **JWT authentication** — login, token issue, `/auth/me`
- **Role-based access control** — Administrator / Operator / Viewer
- **PostgreSQL persistence** — pgx v5 driver, versioned migrations
- **In-memory fallback** — runs without a database for development and tests
- **OpenAPI 3.1** — single source of truth, served embedded from the binary
- **Swagger UI** — interactive API explorer at `/swagger`
- **Structured logging** — `log/slog` with request-ID propagation
- **Middleware pipeline** — panic recovery, request ID, logging, JWT auth, RBAC
- **GitHub Actions CI** — fmt, vet, lint (golangci-lint), tests on every push

---

## Architecture

> Source: [`docs/architecture/architecture.drawio`](docs/architecture/architecture.drawio)
> — open with [diagrams.net](https://app.diagrams.net) or the
> [VS Code draw.io extension](https://marketplace.visualstudio.com/items?itemName=hediet.vscode-drawio).

```
┌─────────────────────────────────────────────────┐
│                   HTTP Client                   │
└───────────────────────┬─────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────┐
│              Middleware Pipeline                 │
│   Recovery → RequestID → Logging → [JWTAuth]    │
│                                    [RequireRole] │
└───────────────────────┬─────────────────────────┘
                        │
          ┌─────────────┴──────────────┐
          │                            │
┌─────────▼──────────┐   ┌────────────▼───────────┐
│   Auth Handlers    │   │  Inventory Handlers     │
│  login / me /      │   │  servers CRUD           │
│  logout            │   │                         │
└─────────┬──────────┘   └────────────┬────────────┘
          │                            │
┌─────────▼──────────┐   ┌────────────▼────────────┐
│    Auth Service    │   │   Inventory Service      │
│  + PasswordHasher  │   │                          │
│  + TokenService    │   │                          │
└─────────┬──────────┘   └────────────┬─────────────┘
          │                            │
┌─────────▼──────────────────────────▼─────────────┐
│             Repository Layer                      │
│   PostgreSQL (pgx)  │  In-Memory (dev/test)       │
└───────────────────────────────────────────────────┘
```

```mermaid
flowchart TD
    Client["HTTP Client"]
    MW["Middleware\nRecovery · RequestID · Logging"]
    Auth["Auth Handlers\nPOST /auth/login\nGET  /auth/me"]
    Inv["Inventory Handlers\nGET|POST|DELETE /servers"]
    AuthSvc["Auth Service\nbcrypt · JWT"]
    InvSvc["Inventory Service"]
    Repo["Repository Layer"]
    PG[("PostgreSQL")]
    Mem["In-Memory\n(dev/test)"]

    Client --> MW
    MW --> Auth
    MW --> Inv
    Auth --> AuthSvc
    Inv --> InvSvc
    AuthSvc --> Repo
    InvSvc --> Repo
    Repo --> PG
    Repo --> Mem
```

---

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| HTTP | `net/http` (stdlib) |
| Database | PostgreSQL 17, pgx v5 |
| Auth | JWT (golang-jwt/jwt v5), bcrypt |
| Migrations | golang-migrate |
| Spec | OpenAPI 3.1 |
| CI | GitHub Actions |

---

## Project Structure

```
atlas/
├── backend/
│   ├── cmd/server/              # Binary entrypoint
│   ├── internal/
│   │   ├── app/                 # Dependency wiring
│   │   ├── auth/                # Auth domain
│   │   │   ├── http/            # Auth HTTP handlers
│   │   │   └── postgres/        # Auth PostgreSQL repository
│   │   ├── config/              # Environment-based configuration
│   │   ├── database/            # pgxpool factory
│   │   ├── http/                # HTTP server + middleware chain
│   │   │   └── middleware/      # Recovery, RequestID, Logging
│   │   ├── httpx/               # WriteJSON / DecodeJSON / WriteError
│   │   ├── inventory/           # Inventory domain
│   │   │   ├── http/            # Inventory HTTP handlers
│   │   │   └── postgres/        # Inventory PostgreSQL repository
│   │   ├── logger/              # slog factory
│   │   └── version/             # Build version info
│   ├── migrations/              # SQL migration files
│   ├── openapi/                 # openapi.yaml + embed + Swagger UI
│   ├── Makefile
│   └── .golangci.yml
├── scripts/                     # PowerShell dev scripts
├── docker-compose.yml
└── .github/workflows/ci.yml
```

---

## Getting Started

### Requirements

- [Go 1.25+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)
- [golangci-lint](https://golangci-lint.run/usage/install/) (optional, for local linting)

### Clone

```bash
git clone https://github.com/killakazzak/atlas.git
cd atlas
```

### Start PostgreSQL

```powershell
.\scripts\db-up.ps1
```

### Apply migrations

```powershell
.\scripts\migrate-up.ps1
```

### Run Atlas

```powershell
.\scripts\run-local.ps1
```

The server starts at `http://localhost:8080`.

> Migrations are **not** applied automatically on server start.
> Always run `migrate-up.ps1` after pulling changes that include new migration files.

---

## Configuration

All settings are read from environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | _(empty)_ | PostgreSQL connection string. When empty the server uses in-memory storage |
| `JWT_SECRET` | `change-me-in-production` | HMAC-SHA256 signing key. **Always override in production** |
| `JWT_ISSUER` | `atlas` | Value placed in the `iss` JWT claim |
| `JWT_ACCESS_TOKEN_TTL` | `1h` | Access token lifetime (Go duration string, e.g. `30m`, `2h`) |

Example `.env` for local development:

```env
DATABASE_URL=postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable
JWT_SECRET=dev-secret-not-for-production
JWT_ACCESS_TOKEN_TTL=8h
```

---

## API

Base URL: `http://localhost:8080`

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/auth/login` | — | Obtain a JWT access token |
| `POST` | `/api/v1/auth/logout` | — | Client-side logout stub (stateless) |
| `GET` | `/api/v1/auth/me` | Any role | Current user profile |

### Inventory

| Method | Path | Required role | Description |
|--------|------|---------------|-------------|
| `GET` | `/api/v1/servers` | Viewer+ | List servers |
| `GET` | `/api/v1/servers/{id}` | Viewer+ | Get a server |
| `POST` | `/api/v1/servers` | Operator+ | Register a server |
| `PUT` | `/api/v1/servers/{id}` | Operator+ | Update a server |
| `DELETE` | `/api/v1/servers/{id}` | Administrator | Delete a server |

### System

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | — | Health check |
| `GET` | `/version` | — | Service version |

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

---

## Authentication

### 1. Obtain a token

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"password"}' | jq .
```

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresIn": 3600,
  "tokenType": "Bearer"
}
```

### 2. Use the token

```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

```bash
curl http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Role-based access

| Role | Permissions |
|------|------------|
| `viewer` | Read-only access to inventory |
| `operator` | Read + write inventory (no delete) |
| `administrator` | Full access including delete and user management |

Insufficient role → **403 Forbidden**. Missing or invalid token → **401 Unauthorized**.

---

## Swagger UI

Interactive API documentation is available when the server is running:

| Resource | URL |
|----------|-----|
| Swagger UI | http://localhost:8080/swagger |
| OpenAPI spec | http://localhost:8080/openapi/openapi.yaml |

The specification source lives at `backend/openapi/openapi.yaml` and is embedded into the binary at build time.

---

## Development

### Scripts

| Script | Purpose |
|--------|---------|
| `.\scripts\db-up.ps1` | Start PostgreSQL container |
| `.\scripts\db-down.ps1` | Stop PostgreSQL container |
| `.\scripts\migrate-up.ps1` | Apply all pending migrations |
| `.\scripts\migrate-down.ps1` | Roll back one migration step |
| `.\scripts\run-local.ps1` | Start the Atlas server |

### Make targets

Run from the `backend/` directory:

```bash
make check   # fmt + vet + lint + test
make test    # go test ./...
make fmt     # go fmt ./...
make vet     # go vet ./...
make lint    # golangci-lint run
make run     # go run ./cmd/server
```

### Run tests directly

```bash
cd backend
go test -timeout 60s ./...
```

### Linting

```bash
cd backend
golangci-lint run
```

---

## CI

GitHub Actions runs on every push to `main` and on every pull request.

| Step | Tool |
|------|------|
| Formatting | `go fmt ./...` |
| Static analysis | `go vet ./...` |
| Linting | `golangci-lint` (built from source with Go 1.25) |
| Tests | `go test -timeout 60s ./...` |

Pipeline config: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

> `golangci-lint` is built from source in CI (`go install`) to ensure binary
> compatibility with Go 1.25. Pre-built binaries up to v2.1.6 were compiled with
> Go 1.24 and cannot analyse Go 1.25 modules.

---

## Roadmap

| Phase | Description | Status |
|-------|-------------|--------|
| 0 | Foundation — repository, docs, CI | ✅ Done |
| 1 | REST API, PostgreSQL, OpenAPI, Swagger | ✅ Done |
| 2 | JWT authentication, RBAC, auth middleware | ✅ Done |
| 3 | Discovery, agents, RabbitMQ | Planned |
| 4 | Console (React UI) | Planned |
| 5 | Infobase and cluster inventory | Planned |

---

## License

[Apache-2.0](LICENSE)
