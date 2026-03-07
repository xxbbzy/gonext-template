# ADR 003: Google Wire for Dependency Injection

- Status: Accepted
- Date: 2026-03-07

## Context

The backend assembles configuration, logging, database access, handlers,
services, JWT support, and rate limiting in a single application graph.
Manually wiring every dependency in `main.go` would increase boilerplate and
make constructor changes harder to track.

The current repository already uses Google Wire to define the object graph in
`backend/cmd/server/wire.go`.

## Decision

Use Google Wire for compile-time dependency injection in the backend.

- Constructor inputs remain explicit in the lower layers.
- `wire.go` and `providers.go` define the assembly graph.
- `wire_gen.go` remains generated output and should not be edited directly.

## Consequences

- The application graph stays explicit and reviewable.
- Runtime startup code in `main.go` stays focused on server composition.
- Contributors must regenerate the Wire output when provider inputs change.
- Developers unfamiliar with Wire need a small amount of onboarding.

## Alternatives Considered

- Manual assembly only in `main.go`: rejected because it does not scale cleanly
  as the graph grows.
- Reflection-based DI container: rejected because it adds runtime indirection
  and hides wiring errors until execution.

## References

- `backend/cmd/server/wire.go`
- `backend/cmd/server/providers.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/main.go`
