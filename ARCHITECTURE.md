# GoNext Template Architecture

This file is the compact AI-facing architecture map for the current repository.
Use `docs/ARCHITECTURE.md` for the bilingual contributor summary and `docs/adr/`
for the rationale behind major stack choices.

## System Shape

- Contract source: `api/openapi.yaml`
- Backend runtime: Gin HTTP server in `backend/cmd/server/main.go`
- Backend DI: Google Wire inputs in `backend/cmd/server/wire.go` and provider helpers in `backend/cmd/server/providers.go`
- Persistence: GORM models and repositories, initialized by `backend/internal/config/database.go`
- Frontend runtime: Next.js App Router under `frontend/app/`
- Frontend data access: Axios client in `frontend/lib/api-client.ts`, the generated OpenAPI-backed wrapper in `frontend/lib/api-client.gen.ts`, and the generated request/response DTOs in `frontend/types/api.ts` (refresh these types with `make gen-types` when the contract changes).

## Request Flow

1. A browser request enters a Next.js App Router page under `frontend/app/`.
2. Client-side data access goes through `frontend/lib/api-client.ts`, which attaches bearer tokens from `frontend/stores/auth.ts`.
3. The request hits Gin in `backend/cmd/server/main.go`.
4. Global middleware runs in this order:
   - `middleware.Recovery`
   - `middleware.RequestLogger`
   - `middleware.ErrorHandler`
   - `cors.New(...)`
5. Health routes and static routes are registered before API groups.
6. API requests enter `/api/v1`, then route-group middleware applies:
   - public auth endpoints use the public rate limiter
   - protected routes use JWT auth and user rate limiting
7. Handlers in `backend/internal/handler/` bind DTOs, call services, and return `pkg/response` envelopes.
8. Services in `backend/internal/service/` implement business logic and convert storage errors into `pkg/errcode.AppError`.
9. Repositories in `backend/internal/repository/` perform GORM operations against models in `backend/internal/model/`.

## Dependency Injection and Runtime Wiring

- `backend/cmd/server/wire.go` declares the compile-time object graph.
- `backend/cmd/server/providers.go` contains adapter constructors that do not live naturally in the lower layers, such as JWT manager creation, rate limiter creation, and application assembly.
- `backend/cmd/server/wire_gen.go` is generated output; update Wire inputs, not the generated file.
- `backend/cmd/server/main.go` is responsible for:
  - creating the application via `InitializeApplication()`
  - running development-only `AutoMigrate`
  - composing middleware
  - registering routes and static assets
  - starting and gracefully shutting down the HTTP server

When adding a new backend module, the usual extension points are:

1. Add repository/service/handler constructors in their layer packages.
2. Add provider helpers in `backend/cmd/server/providers.go` if extra assembly is required.
3. Register constructors in `backend/cmd/server/wire.go`.
4. Extend the `Application` struct when the new handler or dependency must be stored.
5. Register routes in `backend/cmd/server/main.go`.

## Database Initialization

- Configuration is loaded before database setup and passed into `config.NewDatabase`.
- `backend/internal/config/database.go` selects the GORM dialector from `cfg.Database.Driver`.
- SQLite mode creates the parent directory for the DSN path before opening the database.
- PostgreSQL mode opens the configured DSN directly.
- Production-like modes silence the default GORM logger.
- Development mode currently runs `AutoMigrate` for `model.User` and `model.Item` in `backend/cmd/server/main.go`.
- Deployable schema changes should still use SQL migrations under `backend/migrations/`.

## Middleware Topology

- Global middleware should be added only when every route needs it.
- Route-specific middleware belongs on the relevant Gin group in `main.go`.
- Auth middleware depends on the JWT manager and injects `user_id` and `user_role` into Gin context.
- Rate limiting is currently split between public auth traffic and authenticated user traffic.
- Panic recovery and application error translation are centralized in middleware; handlers should not duplicate that logic.

## Frontend Runtime Notes

- `frontend/app/layout.tsx` sets up `NextIntlClientProvider`, `QueryProvider`, and the toast provider.
- `frontend/lib/query-provider.tsx` owns the shared TanStack Query client.
- `frontend/lib/api-client.ts` computes the API base URL, attaches tokens, and handles refresh-on-401 behavior.
- `frontend/stores/auth.ts` persists auth state in local storage via Zustand middleware.
- `frontend/types/api.ts` contains the generated OpenAPI request/response models; refresh them with `make gen-types` when the contract changes, and pair them with the generated wrapper `frontend/lib/api-client.gen.ts` for type-safe fetchers.

## Documentation and Decision Links

- Operational rules: `CONVENTIONS.md`
- Agent task flow: `AGENTS.md`
- Human-facing index: `docs/README.md`
- ADR baseline:
  - `docs/adr/001-openapi-as-contract.md`
  - `docs/adr/002-sqlite-dev-postgres-prod.md`
  - `docs/adr/003-wire-for-di.md`
