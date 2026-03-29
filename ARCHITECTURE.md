# GoNext Template Architecture

This file is the compact AI-facing map for runtime topology, backend layer boundaries, dependency direction, and extension points. See `AGENTS.md` for task playbooks and guardrails, `CONVENTIONS.md` for naming/shape/error rules, `docs/README.md` for the bilingual index, and `docs/adr/` for why the stack is wired this way.

## System Topology

The canonical runtime path flows through:

```
api/openapi.yaml
-> frontend/app/
-> frontend/lib/api-client.gen.ts
-> frontend/stores/auth.ts
-> frontend/lib/query-provider.tsx
-> backend/cmd/server/main.go
-> global and route-group middleware
-> backend/internal/api/server_impl.go -> services
   backend/internal/handler/ (manual Gin routes)
-> repositories
-> models (with `*gorm.DB` bootstrapped at startup via `backend/internal/config/database.go`)
```

The OpenAPI contract feeds the generated client in `frontend/lib/api-client.gen.ts`, which injects tokens from `frontend/stores/auth.ts`, handles 401 refresh, and is consumed by routes under `frontend/app/` that bootstrap TanStack Query through `frontend/lib/query-provider.tsx`. Gin starts in `backend/cmd/server/main.go`, applies middleware, and either dispatches through the generated strict handler implemented in `backend/internal/api/server_impl.go` (auth/item routes) or hits manual Gin endpoints in `backend/internal/handler/`. Services orchestrate use cases after middleware unwinds, repositories talk to models, and the shared `*gorm.DB` connection comes from `backend/internal/config/database.go` during startup.

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

## Extension Points

- Add new modules through `backend/cmd/server/wire.go`, supplement helper constructors in `backend/cmd/server/providers.go`, and register their routes/middleware in `backend/cmd/server/main.go`.
- Middleware belongs in `backend/internal/middleware/`, with Gin global middleware on the router and route-group middleware scoped in `backend/cmd/server/main.go`, keeping context keys and auth expectations stable.
- Database wiring stays in `backend/internal/config/database.go`, so any persistence change must feed through that initializer before services and repositories use the connection.
