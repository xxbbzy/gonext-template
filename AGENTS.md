# GoNext Template Agent Playbook

This repository exposes a compact AI-facing documentation layer at the root:

1. Read `ARCHITECTURE.md` for runtime topology, dependency injection, middleware order, and extension points.
2. Read `CONVENTIONS.md` for coding rules, layer boundaries, error handling, response shape, and testing expectations.
3. Treat `api/openapi.yaml` as the contract source of truth for backend behavior and for the generated frontend types that live at `frontend/types/api.ts`.
4. Use `docs/adr/` for the rationale behind OpenAPI, database strategy, and Google Wire.
5. Use `docs/README.md` when you need the curated human-facing bilingual index.

## Repository Landmarks

- API contract: `api/openapi.yaml`
- Backend runtime entry point: `backend/cmd/server/main.go`
- Dependency injection graph: `backend/cmd/server/wire.go`, `backend/cmd/server/providers.go`, `backend/cmd/server/wire_gen.go`
- Database initialization: `backend/internal/config/database.go`
- Module scaffold generator: `scripts/new-module.sh`
- Architecture guardrail check: `scripts/check-architecture.sh`
- Middleware implementations: `backend/internal/middleware/`
- Prometheus registry wiring: `backend/internal/observability/` + `backend/cmd/server/providers.go`
- Response envelope helpers: `backend/pkg/response/response.go`
- Application error catalog: `backend/pkg/errcode/errcode.go`
- PR quality gate workflow: `.github/workflows/ci-quality-gate.yml`
- Post-merge runtime smoke workflow: `.github/workflows/merge-validation.yml`
- Frontend route tree: `frontend/app/`
- Generated frontend types: `frontend/types/api.ts` (OpenAPI request/response models refreshed via `make gen-types`, the standard command after contract changes; run `make gen` when committed server code or Swagger artifacts must also be refreshed).
- Frontend request layer: `frontend/lib/api-client.gen.ts` (typed OpenAPI-backed client wrapper for base URL configuration, bearer token injection, refresh-on-401, and shared operation helpers).
- Frontend TanStack Query helpers: `frontend/lib/api-query.ts`
- Frontend query bootstrap: `frontend/lib/query-provider.tsx`
- Frontend auth state: `frontend/stores/auth.ts`

## Task Playbooks

### Backend API or Module Change

1. If you need a new backend module skeleton, start with `make new-module name=<module>`; it generates convention-aligned layer files, handler/service/repository tests, and a follow-up checklist.
2. If the request changes API behavior, update `api/openapi.yaml` first.
3. Update or add DTOs in `backend/internal/dto/`.
4. Implement handler, service, repository, and model changes in their respective `backend/internal/*` layers.
5. Wire new dependencies through `backend/cmd/server/wire.go` and `backend/cmd/server/providers.go`.
6. Register routes in `backend/cmd/server/main.go`.
7. If the change adds persistence, update model registration for development `AutoMigrate` and add SQL migrations under `backend/migrations/` for deployable schema changes.
8. After contract changes, run `make gen-types` to refresh frontend types; run `make gen` when committed server code or Swagger artifacts must also be regenerated.
9. If backend behavior is business-critical (key workflows, background jobs, rate-sensitive operations, critical failures), evaluate whether Prometheus instrumentation should ship in the same change.
10. Run practical verification before finishing.

### Frontend Page or Feature Change

1. Confirm the API contract in `api/openapi.yaml`, run `make gen-types`, and import the generated request/response types from `frontend/types/api.ts` (especially for auth and item work) so the UI relies on the OpenAPI DTOs.
2. Add or update routes under `frontend/app/`.
3. Keep network access and shared operation helpers in `frontend/lib/api-client.gen.ts`, and prefer `frontend/lib/api-query.ts` for TanStack Query option builders instead of page-local request wrappers.
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
- Do not write raw JSON envelopes in handlers; use `backend/pkg/response`, except intentionally compact operational endpoints that are explicitly marked for the guardrail allowlist.
- Avoid high-cardinality Prometheus labels (raw paths, IDs, free-text, emails, tokens, etc.); prefer bounded labels and route templates.
- Do not introduce framework drift. This project uses Gin, GORM, Google Wire, Next.js App Router, Zustand, TanStack Query, and OpenAPI-driven types.
- Prefer minimal, file-local changes over broad rewrites unless the task explicitly requires a refactor.

## Mandatory Verification

> **After completing ANY code change, you MUST run `make check` and confirm all checks pass (exit code 0).**
> Do **not** consider a task complete until `make check` is green.

`make check` runs the full validation pipeline:

1. **Lint & Guardrails** — `golangci-lint` (backend) + `eslint` (frontend) + `make check-architecture`
2. **Typecheck** — `tsc --noEmit` (frontend)
3. **Test** — `go test ./...` (backend) + `vitest run` (frontend) + `make test-tooling`
4. **Build** — `go build` (backend) + `next build` (frontend)

For API behavior changes that involve runtime, also run `make e2e` to exercise the register → login → CRUD cycle.

## Verification Shortlist

- `make check` — **mandatory** after every code change
- `make e2e` — after API or runtime behavior changes
- `make gen-types` after `api/openapi.yaml` changes (frontend DTO refresh)
- `make gen` when committed server code or Swagger artifacts must also be refreshed
- Review `docs/README.md` after documentation work to confirm the navigation still makes sense
