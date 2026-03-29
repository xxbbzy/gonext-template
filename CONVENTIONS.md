# GoNext Template Conventions

This file defines the compact, canonical rules for implementation work. Human
summaries live under `docs/`, but repository-wide operational rules belong here.

## Naming and File Placement

- Go exports use `PascalCase`; internal helpers use `camelCase`.
- Keep one resource-oriented file per layer when possible: `handler/item.go`, `service/item.go`, `repository/item.go`, `model/item.go`, `dto/item.go`.
- DTO names should communicate intent, for example `CreateItemRequest`, `UpdateItemRequest`, `ItemResponse`.
- JSON tags should stay `snake_case`.
- Keep frontend routes in `frontend/app/`; keep shared client utilities in `frontend/lib/`; keep persisted client state in `frontend/stores/`.

## Layer Boundaries

- Handlers parse HTTP input, call services, and return responses. They should not query GORM directly.
- Services own business rules, orchestration, and translation of infrastructure failures into application errors.
- Repositories own database access only.
- Models describe persistence shape; DTOs describe transport shape.
- Middleware owns cross-cutting HTTP concerns such as auth, request logging, panic recovery, and error translation.

## Error Handling

- Prefer predefined `pkg/errcode` values or `errcode.New(...)` for application-level errors.
- Services should return typed application errors when the handler must preserve a specific HTTP status or business code.
- Handlers should translate `*errcode.AppError` through `pkg/response.Error`.
- Use `response.BadRequest`, `response.Unauthorized`, `response.Forbidden`, `response.NotFound`, and `response.InternalServerError` rather than ad hoc JSON bodies.
- Let middleware handle panic recovery and `c.Errors` propagation instead of re-implementing global recovery logic in handlers.

## Response Rules

- The API envelope is defined in `backend/pkg/response/response.go`.
- Successful responses should use `Code: 0` and the appropriate helper (`Success`, `Created`, `PagedSuccess`).
- Error responses should keep the shared envelope shape and use the matching HTTP status and application code.
- Do not return raw `c.JSON(...)` payloads from handlers unless you are extending `pkg/response` itself.

## Logging Rules

- Use Zap for application logging.
- Keep request-level access logging in `backend/internal/middleware/logger.go`.
- Prefer structured fields over interpolated strings.
- Avoid introducing parallel logging frameworks or per-handler logging patterns unless the change has a clear operational need.
- Startup and shutdown logging belongs in the runtime entry points, not scattered across feature handlers.

## Testing

- Put backend tests in the same package area as the code they exercise, using `_test.go` files.
- Use `internal/testutil.NewTestDB(t, models...)` to get a fresh SQLite in-memory database for each test — zero external dependencies.
- Existing backend test examples live in `backend/internal/handler/`, `backend/internal/service/`, `backend/internal/repository/`, `backend/internal/middleware/`, and `backend/internal/config/`.
- When adding middleware, handlers, or modules, add corresponding tests. The `new-module.sh` scaffold auto-generates `_test.go` files.
- Frontend tests use Vitest + React Testing Library (`npm test` or `make test-frontend`).
- `make check` is the canonical validation command — it runs lint, typecheck, test, and build in one pipeline.
- `make e2e` runs a full register → login → CRUD smoke test against a real backend (SQLite, ephemeral port).

## API Contract Synchronization

1. Update `api/openapi.yaml` before or alongside backend API behavior changes.
2. Run `make gen-types` to refresh the generated TypeScript request/response models in `frontend/types/api.ts`.
3. Run `make gen` when the generated Go server code, TypeScript types, and Swagger/docs artifacts must be refreshed together.
4. Run `make swagger` whenever the Swagger output must stay aligned with the OpenAPI contract.
5. Keep handler comments, generated docs, and frontend request/response typing aligned with the OpenAPI contract.

## Documentation Maintenance

- Root docs (`AGENTS.md`, `ARCHITECTURE.md`, `CONVENTIONS.md`) are the canonical AI-facing reference and stay English-only.
- Curated human docs under `docs/` stay bilingual Chinese/English and should summarize or link, not duplicate this file.
- Record repository-wide architectural choices in `docs/adr/`.
- When changing runtime entry points, update both the relevant root doc and the minimal human summary that references it.
