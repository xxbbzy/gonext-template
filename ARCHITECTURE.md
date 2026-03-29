# GoNext Template Architecture

This file is the compact AI-facing map for runtime topology, backend layer boundaries, dependency direction, and extension points. See `AGENTS.md` for task playbooks and guardrails, `CONVENTIONS.md` for naming/shape/error rules, `docs/README.md` for the bilingual index, and `docs/adr/` for why the stack is wired this way.

## System Topology

The canonical runtime path flows through:

```
api/openapi.yaml
-> frontend/app/
-> frontend/lib/api-client.gen.ts
-> frontend/lib/api-query.ts
-> frontend/stores/auth.ts
-> frontend/lib/query-provider.tsx
-> backend/cmd/server/main.go
-> global middleware + generated per-operation middleware
-> backend/internal/api/server_impl.go -> services
   backend/internal/handler/ (manual Gin routes)
-> repositories
-> models (with `*gorm.DB` bootstrapped at startup via `backend/internal/config/database.go`)
```

The OpenAPI contract is the source of truth: it feeds the shared request helpers in `frontend/lib/api-client.gen.ts`, the TanStack Query option builders in `frontend/lib/api-query.ts`, and the generated request/response types in `frontend/types/api.ts`. Routes under `frontend/app/` consume that path while `frontend/lib/query-provider.tsx` bootstraps TanStack Query defaults. The auth store in `frontend/stores/auth.ts` provides token persistence and refresh state, while `frontend/app/layout.tsx` assembles the top-level frontend providers.

Gin starts in `backend/cmd/server/main.go`, applies global middleware in the order `middleware.Recovery`, `middleware.RequestLogger`, `middleware.ErrorHandler`, then `cors.New(...)`, and registers health/static routes before the `/api/v1` groups. Generated strict handlers are mounted via `genapi.RegisterHandlersWithOptions(...)` with per-operation middleware for auth and item routes; manual Gin endpoints (like upload) stay in `backend/internal/handler/` with their own routing. Public auth endpoints use the public rate limiter, protected routes use JWT auth plus user rate limiting, services orchestrate use cases after middleware unwinds, repositories talk to models, and the shared `*gorm.DB` connection comes from `backend/internal/config/database.go` during startup.

## Frontend And Runtime Integration Points

Start cross-stack integration work at these anchors; they are the stable entry points for wiring, not full implementation guides:

- `frontend/app/` is the route tree that calls the API layer and hosts page-level composition.
- `frontend/app/layout.tsx` assembles shared frontend providers such as i18n, TanStack Query, and toasts.
- `frontend/lib/api-client.gen.ts` is the canonical OpenAPI-backed client entry point with auth + refresh middleware plus shared operation wrappers.
- `frontend/lib/api-query.ts` provides the typed TanStack Query option builders that sit on top of the shared client path.
- `frontend/lib/query-provider.tsx` configures TanStack Query defaults for frontend data access.
- `frontend/stores/auth.ts` owns token persistence and user state for client auth.
- `frontend/types/api.ts` contains the generated OpenAPI request/response models; refresh it alongside the client artifacts when the contract changes.
- `backend/cmd/server/main.go` mounts global middleware, registers generated handlers via `genapi.RegisterHandlersWithOptions(...)`, applies per-operation middleware for generated routes, and wires manual Gin routes like upload separately.
- `backend/cmd/server/wire.go` + `backend/cmd/server/providers.go` declare constructors; `backend/cmd/server/wire_gen.go` is generated output.
- `backend/internal/config/database.go` initializes the shared `*gorm.DB` used by repositories.

## Backend Layer Responsibilities

### handler

- Owns HTTP binding, validation, error translation, and response envelopes for manual Gin endpoints (uploads, health, etc.).
- Depends on services, DTOs, `pkg/response`, and the middleware + router wiring in `backend/cmd/server/main.go`.
- Must not reach into GORM, mutate domain models, or emit raw JSON that bypasses `pkg/response`.
- Lives in `backend/internal/handler/`; generated OpenAPI routes instead run through `backend/internal/api/server_impl.go`, but this layer covers Gin-specific hooks and constructor wiring.
- When adding a module with manual routes, add handler constructors, register them and middleware in `backend/cmd/server/main.go`, and expose them via `wire.go`/`providers.go`.

### service

- Owns orchestration, business rules, and conversion of infrastructure errors into `pkg/errcode.AppError`.
- Depends on repositories, DTOs, JWT helpers, and the shapes defined in `backend/internal/model/`.
- Must not know about Gin, middleware, or HTTP response envelopes; it should not import handler packages.
- Lives in `backend/internal/service/`.
- When adding a module, extend service constructors, add providers/wire entries, and update unit tests focused on business logic.

### repository

- Owns GORM database access, queries, and persistence helpers for the domain.
- Depends on `backend/internal/model/`, the shared `*gorm.DB`, and transaction helpers if needed.
- Must not include orchestration, pagination state, or API-specific error handling.
- Lives in `backend/internal/repository/`.
- When adding a module, add the repository implementation and expose it through Wire so services can consume it; pair the work with migrations or AutoMigrate registration.

### model

- Owns the persistence schema, struct tags, table names, and helper constructors that describe how the database stores each entity.
- Depends on GORM field types, timestamps, and any base structs shared across models.
- Must not reference services, handlers, or DTOs, and should not embed business logic beyond trivial helpers.
- Lives in `backend/internal/model/`.
- When adding a module, add the model definition, include it in migrations/AutoMigrate, and document the schema change if it touches production databases.

### dto

- Owns the transport shapes for requests and responses consumed by handlers and services.
- Depends on JSON tagging rules, validation tags, and the OpenAPI schema defined in `api/openapi.yaml`.
- Must not reference database tags, persistence helpers, or embed infrastructure logic.
- Lives in `backend/internal/dto/`.
- When adding a module, add the request/response DTOs, update the OpenAPI contract if needed, and regenerate frontend types so the client and handlers stay in sync.

## Dependency Direction And Boundaries

Intended dependency flow: `handler -> service -> repository -> model`.

- DTOs stay on the transport side (handlers + services) and map to OpenAPI-defined shapes.
- Middleware owns HTTP cross-cutting concerns (auth, rate limits, recovery, logging).
- `backend/pkg/response` and `backend/pkg/errcode` are shared support packages for envelopes and error cataloging.
- Wire files assemble dependencies (`backend/cmd/server/wire.go`, `backend/cmd/server/providers.go`, generated `backend/cmd/server/wire_gen.go`) but do not own business behavior.

Anti-patterns to avoid:

- Handlers querying GORM directly.
- Services returning HTTP payload semantics.
- Repositories owning business policy.
- Models acting as default API response structs.

## Extension Points

- Add new modules through `backend/cmd/server/wire.go`, supplement helper constructors in `backend/cmd/server/providers.go`, and register their routes/middleware in `backend/cmd/server/main.go`.
- Middleware belongs in `backend/internal/middleware/`, with Gin global middleware on the router and per-operation middleware switched inside the generated handler registration in `backend/cmd/server/main.go`, keeping context keys and auth expectations stable.
- Database wiring stays in `backend/internal/config/database.go`, so any persistence change must feed through that initializer before services and repositories use the connection.

## Recommended Flow For Adding A New Module

- [ ] Update `api/openapi.yaml` when API behavior changes.
- [ ] Add or update DTOs in `backend/internal/dto/`.
- [ ] Implement handler/service/repository/model changes in their respective layers.
- [ ] Register providers/constructors in `backend/cmd/server/providers.go` and `backend/cmd/server/wire.go`.
- [ ] Regenerate `backend/cmd/server/wire_gen.go` through the existing Wire workflow.
- [ ] Register routes in `backend/cmd/server/main.go`.
- [ ] Update development AutoMigrate coverage and add SQL migrations under `backend/migrations/` when persistence changes.
- [ ] Run `make gen` when the contract changed.
- [ ] Run `make check`.
- [ ] Run `make e2e` for API/runtime behavior changes.

## Links

- `AGENTS.md`
- `CONVENTIONS.md`
- `docs/README.md`
- `docs/adr/001-openapi-as-contract.md`
- `docs/adr/002-sqlite-dev-postgres-prod.md`
- `docs/adr/003-wire-for-di.md`
