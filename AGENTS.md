# Atlas Agent Instructions

## Project overview
Atlas is an enterprise platform for managing 1C infrastructure, servers, services, agents, inventory, discovery, deployment and orchestration.

## Core principles
- Prefer simple and maintainable solutions.
- Do not add unnecessary abstractions.
- Do not delete existing files without explicit approval.
- Do not introduce new dependencies without explaining why.
- Keep documentation consistent with implementation.
- Make small, reviewable changes.
- Never commit secrets, API keys, passwords or certificates.

## Repository structure
- backend/ — backend services and API
- frontend/ — web console
- agent/windows/ — Windows agent
- agent/linux/ — Linux agent
- docs/ — architecture, ADRs, API and product documentation
- scripts/ — development and deployment scripts
- .github/workflows/ — CI/CD

## Backend rules
- Backend language: Go.
- Follow standard Go project conventions.
- Use clear package boundaries.
- Prefer the standard library where practical.
- Use context.Context for I/O and request-scoped operations.
- Add structured logging.
- Return explicit errors and wrap them with context.
- Avoid global mutable state.
- Write tests for business-critical logic.

## API rules
- Use REST for the initial MVP.
- Use JSON for request and response bodies.
- Version public endpoints under /api/v1.
- Use consistent error responses.
- Validate all external input.
- Do not expose internal implementation details.
- Document endpoints in OpenAPI.

## Database rules
- Primary database: PostgreSQL.
- Use migrations for all schema changes.
- Never modify an applied migration.
- Use UUIDs for distributed entity identifiers.
- Store timestamps in UTC.
- Add indexes only with a documented reason.

## Messaging
- RabbitMQ may be used for asynchronous tasks and events.
- Events must have explicit names and versioned payloads.
- Consumers must be idempotent where possible.

## Security
- Follow least-privilege principles.
- Do not log secrets or sensitive credentials.
- Validate authorization on every protected operation.
- Use secure defaults.
- Separate authentication, authorization and audit logging.

## Documentation
- Record important architectural decisions in docs/adr/.
- Update architecture documentation when system boundaries change.
- Use Mermaid for diagrams where practical.
- Keep README concise and current.

## Agent behavior
Before making changes:
1. Inspect relevant files.
2. Explain the intended change briefly.
3. Make the smallest necessary change.
4. Run available tests or validation.
5. Summarize changed files and any remaining risks.
