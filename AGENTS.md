# GoNext Template Agent Playbook

This repository exposes a compact AI-facing documentation layer at the root:

1. Read `ARCHITECTURE.md` for runtime topology, dependency injection, middleware order, and extension points.
2. Read `CONVENTIONS.md` for coding rules, layer boundaries, error handling, response shape, and testing expectations.
3. Treat `api/openapi.yaml` as the contract source of truth for backend and generated frontend types.
4. Use `docs/adr/` for the rationale behind OpenAPI, database strategy, and Google Wire.
5. Use `docs/README.md` when you need the curated human-facing bilingual index.

## Repository Landmarks

- API contract: `api/openapi.yaml`
- Backend runtime entry point: `backend/cmd/server/main.go`
- Dependency injection graph: `backend/cmd/server/wire.go`, `backend/cmd/server/providers.go`, `backend/cmd/server/wire_gen.go`
- Database initialization: `backend/internal/config/database.go`
- Middleware implementations: `backend/internal/middleware/`
- Response envelope helpers: `backend/pkg/response/response.go`
- Application error catalog: `backend/pkg/errcode/errcode.go`
- Frontend route tree: `frontend/app/`
- Frontend request layer: `frontend/lib/api-client.ts`
- Frontend query bootstrap: `frontend/lib/query-provider.tsx`
- Frontend auth state: `frontend/stores/auth.ts`

## Task Playbooks

### Backend API or Module Change

1. If the request changes API behavior, update `api/openapi.yaml` first.
2. Update or add DTOs in `backend/internal/dto/`.
3. Implement handler, service, repository, and model changes in their respective `backend/internal/*` layers.
4. Wire new dependencies through `backend/cmd/server/wire.go` and `backend/cmd/server/providers.go`.
5. Register routes in `backend/cmd/server/main.go`.
6. If the change adds persistence, update model registration for development `AutoMigrate` and add SQL migrations under `backend/migrations/` for deployable schema changes.
7. Regenerate derived artifacts with `make swagger` and `make gen-types` when the contract changed.
8. Run practical verification before finishing.

### Frontend Page or Feature Change

1. Confirm the API contract in `api/openapi.yaml` and generated types in `frontend/types/api.ts`.
2. Add or update routes under `frontend/app/`.
3. Keep network access in `frontend/lib/api-client.ts` or feature-specific wrappers that use it.
4. Keep auth state in `frontend/stores/auth.ts`; do not duplicate token persistence logic.
5. Reuse `frontend/lib/query-provider.tsx` for server-state flows and existing UI/component patterns in `frontend/components/`.
6. Run `npm run build` in `frontend/` or `make build` when practical.

### Middleware or Runtime Wiring Change

1. Implement middleware in `backend/internal/middleware/`.
2. Mount global middleware in `backend/cmd/server/main.go` only when it must affect every request.
3. Mount route-scoped middleware on the relevant Gin groups when it applies only to public or authenticated endpoints.
4. Preserve current context keys and auth expectations used by handlers (`user_id`, `user_role`) unless the change explicitly updates all consumers.
5. Add or update package-local tests alongside the middleware.

### Documentation Change

1. Keep root docs (`AGENTS.md`, `ARCHITECTURE.md`, `CONVENTIONS.md`) English-only and compact.
2. Keep curated human docs under `docs/` bilingual Chinese/English.
3. Add an ADR in `docs/adr/` for repository-wide architecture decisions; do not hide rationale only in code comments.
4. Prefer links and summaries in `docs/` over duplicating the full operational guidance from the root docs.
5. When runtime entry points or workflows change, update both the relevant root doc and the minimal human summary that points to it.

## Guardrails

- Do not bypass the backend layer chain: handler -> service -> repository.
- Do not treat generated files as the source of truth when `api/openapi.yaml` or Wire inputs disagree.
- Do not write raw JSON envelopes in handlers; use `backend/pkg/response`.
- Do not introduce framework drift. This project uses Gin, GORM, Google Wire, Next.js App Router, Zustand, TanStack Query, and OpenAPI-driven types.
- Prefer minimal, file-local changes over broad rewrites unless the task explicitly requires a refactor.

## Verification Shortlist

- `make check` for practical repository-level validation
- `make swagger` after backend API annotation or contract changes
- `make gen-types` after `api/openapi.yaml` changes
- Review `docs/README.md` after documentation work to confirm the navigation still makes sense
