# ADR 002: SQLite for Local Development, PostgreSQL for Deployable Environments

- Status: Accepted
- Date: 2026-03-07

## Context

The project needs a low-friction local setup while still supporting a database
that matches common deployment requirements. The current codebase already
supports both SQLite and PostgreSQL through GORM dialectors selected in
`backend/internal/config/database.go`.

## Decision

Use SQLite by default for local development and PostgreSQL for deployable or
shared environments.

- Local development favors fast bootstrap and minimal external dependencies.
- Deployable environments favor PostgreSQL for operational compatibility.
- Development mode may use `AutoMigrate` for speed, but release-oriented schema
  changes should still be expressed as SQL migrations under `backend/migrations/`.

## Consequences

- New contributors can start the project without provisioning PostgreSQL first.
- The repository still supports a production-ready relational database path.
- Schema changes must be checked with awareness of dialect differences.
- Teams cannot assume that local `AutoMigrate` alone is sufficient for
  deployable migrations.

## Alternatives Considered

- PostgreSQL everywhere: rejected because it raises the local setup cost.
- SQLite everywhere: rejected because it weakens the deployable environment
  story and narrows operational options.

## References

- `backend/internal/config/database.go`
- `backend/cmd/server/main.go`
- `backend/migrations/`
- `docker-compose.yml`
