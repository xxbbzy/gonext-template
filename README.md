# GoNext Template

A compact full-stack starter with:

- **Backend:** Go + Gin + GORM + Wire + OpenAPI
- **Frontend:** Next.js App Router + TypeScript + Tailwind + Zustand + TanStack Query
- **Contract-first API:** `api/openapi.yaml` drives generated frontend types and backend server stubs

## Quick Start

Requirements:

- Go 1.25+
- Node.js 20+

```bash
cp .env.example .env
make init
make dev
```

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Repo Guides

- AI-facing playbook: `AGENTS.md`
- Runtime topology: `ARCHITECTURE.md`
- Coding/testing rules: `CONVENTIONS.md`
- Human bilingual docs index: `docs/README.md`
- API contract: `api/openapi.yaml`

## Core Workflow Expectations

- **Contract-first:** update `api/openapi.yaml` before or alongside API behavior changes.
- **Type refresh:** run `make gen-types` after contract changes; run `make gen` when generated Go server/docs artifacts must also change.
- **Layering rule:** handlers → services → repositories; keep handlers thin and services orchestrating logic.
- **Response envelope:** use helpers in `backend/pkg/response` instead of raw JSON writes.
- **Generated files:** treat artifacts such as `frontend/types/api.ts` and `backend/internal/api/server.gen.go` as outputs, never the source of truth.

## Local Development Flow

1. `make init` (once) bootstraps `.env`, installs dependencies and local folders, runs the backend bootstrap (AutoMigrate convenience), and refreshes generated artifacts.
2. `make dev` (runs backend + frontend watchers). Keep editing UI or Go sources.
3. Run `make check` (lint + typecheck + test + build) whenever you change logic/code paths to verify the full gate.
4. If the change touches runtime behavior or APIs, re-run `make e2e` to exercise the register → login → CRUD cycle.
5. Repeat: edit → lint/type/test → `make check` → `make e2e` (if needed) → commit.

## Upload Storage Modes

- **Default (local):** keep `STORAGE_DRIVER=local` and uploaded files are saved under `UPLOAD_DIR`, served by backend `/uploads/...`.
- **Object storage (S3-compatible):** set `STORAGE_DRIVER=s3` and configure:
  - `S3_BUCKET`, `S3_REGION`
  - `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`
  - Optional: `S3_ENDPOINT` (for MinIO/custom endpoints), `S3_PREFIX`, `S3_USE_SSL`, `S3_FORCE_PATH_STYLE`
- `UPLOAD_PUBLIC_BASE_URL` is optional in both modes. When set, upload responses use that public base URL (useful for CDN/custom domains).

Example MinIO settings:

```env
STORAGE_DRIVER=s3
S3_BUCKET=gonext-uploads
S3_REGION=us-east-1
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_PREFIX=uploads
S3_USE_SSL=false
S3_FORCE_PATH_STYLE=true
```

## Observability (Prometheus)

- Prometheus scraping is opt-in. Set `METRICS_ENABLED=true` to expose `http://localhost:8080/metrics`.
- `/metrics` is an operational endpoint (not part of `api/openapi.yaml` or frontend codegen).
- Baseline backend metric families:
  - `http_requests_total{method,route,status}`
  - `http_request_duration_seconds{method,route}`
- Scrape output also includes default Go/runtime + process metrics from the Prometheus Go client.
- Local check:

  ```bash
  METRICS_ENABLED=true make dev
  curl -s http://localhost:8080/metrics | head -n 40
  ```

- Production note: protect `/metrics` with network/proxy controls where needed.

## Docker Workflow

```bash
make docker-build
make docker-up
make docker-down
```

Services & ports (per `docker-compose.yml` services):

- `frontend`: `http://localhost:3000`
- `backend`: `http://localhost:8080`
- `db` (PostgreSQL): `localhost:5432`

## OpenAPI & Type Generation

- Update `api/openapi.yaml` first whenever you touch API behavior, then refresh downstream artifacts.
- Run `make gen-types` as the standard TypeScript refresh; it always writes the API shapes to `frontend/types/api.ts` and should follow any contract change that affects the frontend.
- Run `make gen` only when you need to regenerate the Go server stubs and Swagger docs in addition to the TypeScript types (it runs `gen-server`, `gen-types`, and `swagger`).
- Before pushing a PR that changes OpenAPI/codegen inputs, run `make check-codegen-drift` to execute the same regeneration drift rule as CI.
- If `make check-codegen-drift` fails, run `make gen`, inspect `git status`, commit all generated artifact changes, then rerun `make check-codegen-drift`.
- Run `make swagger` whenever you adjust OpenAPI metadata or docs, keeping `backend/docs/swagger.*` in sync.

## How To Add A New Module

Run `make new-module name=product`, then:

1. Update `api/openapi.yaml` if the module is API-backed.
2. Implement the backend chain in `backend/internal/{handler,service,repository,model,dto}`.
3. Wire dependencies through `backend/cmd/server/{providers.go,wire.go}`.
4. Register generated/manual routes in `backend/cmd/server/main.go`.
5. Refresh generated artifacts as needed (`make gen-types` / `make gen`).
6. Run `make check`, and `make e2e` if runtime/API behavior changed.
