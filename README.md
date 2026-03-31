# GoNext Template

> AI-native full-stack template for solo builders who work like a team.

## Prerequisites

- Go 1.25+ toolchain (see `backend/go.mod` toolchain declaration).
- Node.js 20+ with matching `npm` (`>=20.9.0` recommended for Next.js 16).
- GNU `make` (for the included targets and workflows).
- Optional tooling: Docker/Compose for the Docker workflow and the migrate CLI when adding SQL migrations.

## Who This Is For

- Solo founders and consultants who need a Google Wire-based Go + Next.js scaffold before their engineers get back online.
- AI engineers prototyping agent-assisted workflows that must stay in sync between backend, frontend, and instrumentation.
- Bootcampers or learners who want a production-like full-stack example with OpenAPI-driven types and strong conventions.
- Agencies building bespoke products that must move fast without sacrificing backend structure or deployment hygiene.

**Not ideal for:** Regulated enterprises with large, siloed teams needing multi-tenant SSO or legacy transformation initiatives.

## Quick Start

```bash
git clone <your-repo-url>
cd gonext-template
make init
make dev
```

**Success checks**

- Backend health: `http://localhost:8080/healthz`
- Frontend dev: `http://localhost:3000`

## Tech Stack & Architecture Conventions

- **Backend:** Go, Gin, GORM, Google Wire compile-time dependency injection, Zap logging, Viper configs.
- **Frontend:** Next.js 16 (App Router), TypeScript, shadcn/ui, Zustand, TanStack Query, OpenAPI fetch client.
- **OpenAPI contract:** `api/openapi.yaml` is the source of truth for every request/response shape.
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

1. Update `api/openapi.yaml` first if the module exposes an API surface.
2. Implement handler → service → repository (DTOs live under `backend/internal/dto/`).
3. Declare providers/constructors in `backend/cmd/server/wire.go` and `providers.go`, then run `wire` (or `go run github.com/google/wire/cmd/wire`; if using `go generate`, run `go generate -tags wireinject`) in `backend/cmd/server` to refresh `wire_gen.go`.
4. Register routes in `backend/cmd/server/main.go`.
5. Add deployable SQL migrations under `backend/migrations` (AutoMigrate remains a dev convenience) and any seed data if needed.
6. Verify with `make check` and `make e2e` (if behavior changed) before merging.

## Roadmap

**Near-term focus:** sharpen developer onboarding, shore up e2e coverage, and automate doc generation for latest APIs.
**Longer-term direction:** invest in AI-native ops (agent-friendly scripts, observability), scale modules for plugin scenarios, and explore multi-cluster Docker compose support.

## Documentation Map

- [AGENTS.md](AGENTS.md)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [CONVENTIONS.md](CONVENTIONS.md)
- [docs/README.md](docs/README.md)
