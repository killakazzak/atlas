# Development Workflow

Engineering conventions for the Atlas monorepo. Applies from Phase 0 onward; CI enforcement strengthens as application code lands.

## Principles

- **Documentation is part of the deliverable** — structural or behavioral changes include doc updates in the same pull request
- **Decisions are recorded** — non-trivial architecture choices become ADRs before or alongside implementation
- **Small, reviewable changes** — prefer incremental PRs aligned to roadmap phases
- **No secrets in source** — credentials via environment variables or vault; `.env` files remain local and gitignored

## Branching Model

Trunk-based development with short-lived feature branches.

| Branch pattern | Use |
|----------------|-----|
| `main` | Always deployable; protected |
| `feature/<ticket>-<description>` | New capability |
| `fix/<ticket>-<description>` | Bug fix |
| `docs/<description>` | Documentation-only changes |
| `chore/<description>` | Tooling, CI, dependencies |

Branches merge to `main` via pull request. Long-lived release branches may be introduced post-MVP if deployment cadence requires it.

## Pull Request Process

1. **Create branch** from latest `main`
2. **Implement change** with tests when code exists; update affected docs
3. **Open PR** with:
   - Linked issue or ticket reference
   - Summary of change and motivation
   - Test evidence (or N/A for docs-only)
   - Screenshot for UI changes (when applicable)
4. **Review** — minimum one approval from a code owner
5. **CI** — all required checks pass
6. **Merge** — squash merge to `main`; delete branch

### PR size guidance

| Size | Lines changed (approx.) | Expectation |
|------|-------------------------|-------------|
| Small | < 200 | Preferred |
| Medium | 200–500 | Acceptable with clear description |
| Large | > 500 | Split unless mechanically generated |

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

[optional body]

[optional footer: Refs #123]
```

| Type | Usage |
|------|-------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Tooling, CI, non-functional changes |
| `refactor` | Code change without behavior change |
| `test` | Tests only |

**Scopes** (examples): `inventory`, `console`, `agent-win`, `agent-linux`, `orchestrator`, `ci`, `docs`

## Architecture Decision Records

Create an ADR in `docs/adr/` when a decision:

- Affects multiple services or repositories areas
- Is difficult or costly to reverse
- Introduces a new dependency, protocol, or deployment pattern

Format: `ADR-NNNN-kebab-title.md`. Follow the template established in [ADR-0001](../adr/ADR-0001-project-structure.md).

Status lifecycle: `Proposed` → `Accepted` → `Deprecated` | `Superseded`

## Documentation Standards

| Change type | Update |
|-------------|--------|
| New service or component | `docs/architecture/architecture.md` |
| Goal or scope shift | `docs/vision/vision.md` |
| Delivery timeline change | `docs/vision/roadmap.md` |
| API contract | `docs/api/` (OpenAPI) |
| Schema change | `docs/database/` + migration |
| Repo layout change | New ADR |

Avoid duplicating content across files. README links to canonical sources; vision owns goals/non-goals; architecture owns components and stack.

## CI Pipeline (Phased)

| Phase | CI expectations |
|-------|-----------------|
| 0 | Structure validation, markdown lint (when configured) |
| 1 | Backend build, unit tests, OpenAPI diff |
| 2 | Agent build matrix (Windows, Linux), integration tests |
| 3+ | End-to-end workflow tests, security scanning |

Workflow definitions live in `.github/workflows/`.

## Local Development (Upcoming)

Phase 1 introduces Docker Compose for PostgreSQL and RabbitMQ. Until then:

1. Clone repository
2. Read [Architecture](architecture.md) and [Roadmap](../vision/roadmap.md)
3. Use `docs/sprint/` for active sprint planning artifacts

Detailed setup instructions will be added to this document when `backend/` and `frontend/` scaffolding lands.

## Code Review Checklist

Reviewers verify:

- [ ] Change aligns with stated goal and roadmap phase
- [ ] Documentation updated or correctly omitted
- [ ] No secrets, credentials, or environment-specific hardcoding
- [ ] ADR created if decision warrants it
- [ ] Tests included for behavioral changes (when code exists)

## Related Documentation

- [README](../../README.md) — project entry point
- [ADR-0001](../adr/ADR-0001-project-structure.md) — repository layout
- [Roadmap](../vision/roadmap.md) — delivery phases
