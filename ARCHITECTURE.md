# GoNext Template Architecture

This file is the compact AI-facing map for runtime topology, backend layer boundaries, dependency direction, and extension points. See `AGENTS.md` for task playbooks and guardrails, `CONVENTIONS.md` for naming/shape/error rules, `docs/README.md` for the bilingual index, and `docs/adr/` for why the stack is wired this way.

## System Topology

The canonical data path flows through the following components:

```
api/openapi.yaml
-> frontend/app/
-> frontend/lib/api-client.ts + frontend/lib/api-client.gen.ts
-> frontend/lib/query-provider.tsx / frontend/stores/auth.ts
-> backend/cmd/server/main.go
-> global and route-group middleware
-> handler -> service -> repository -> model
-> backend/internal/config/database.go
```

The OpenAPI contract powers the generated frontend client, TanStack Query relies on `QueryProvider`, and frontend auth state feeds the client before the request reaches Gin. Middleware and the handler/service/repository/model chain push the work down to the database initializer in `backend/internal/config/database.go`.

## Backend Layer Responsibilities

### handler

- Owns HTTP binding, validation, error translation, and response envelopes for each API surface.
- Depends on services, DTOs, `pkg/response`, application middleware for auth/rate limits, and the running context wired in `backend/cmd/server/main.go`.
- Must not reach into GORM, mutate domain models, or emit raw JSON that bypasses `pkg/response`.
- Lives in `backend/internal/handler/`, and routes plus middleware chains are tied together in `backend/cmd/server/main.go`.
- When adding a module, add handler constructors, wire them through `backend/cmd/server/main.go`, register routes, and expose them via `wire.go`/`providers.go`.

### service

- Owns orchestration, business rules, and conversion of infrastructure errors into `pkg/errcode.AppError`.
- Depends on repositories, DTOs, auth helpers (JWT manager, rate limiter), and the shapes defined in `backend/internal/model/`.
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
