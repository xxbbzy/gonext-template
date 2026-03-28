# GoNext Template

> AI-native full-stack template for solo builders who work like a team.

## Who This Is For

- Solo founders and consultants who need a fully wired Go + Next.js scaffold before their engineers get back online.
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

- **Backend:** Go, Gin, GORM, Google Wire + manual DI helpers, Zap logging, Viper configs.
- **Frontend:** Next.js 16 (App Router), TypeScript, shadcn/ui, Zustand, TanStack Query, OpenAPI fetch client.
- **OpenAPI contract:** `api/openapi.yaml` is the source of truth for every request/response shape.
- **Layering rule:** handlers → services → repositories; keep handlers thin and services orchestrating logic.
- **Response envelope:** use helpers in `backend/pkg/response` instead of raw JSON writes.
- **Generated files:** treat artifacts such as `frontend/types/api.ts` and `backend/internal/api/server.gen.go` as outputs, never the source of truth.

## Local Development Flow

1. `make init` (once) beefs up `.env`, migrations, and schema seeds.
2. `make dev` (runs backend + frontend watchers). Keep editing UI or Go sources.
3. Run `make check` whenever you change logic/code paths to verify lint, typecheck, and tests.
4. If the change touches runtime behavior or APIs, re-run `make e2e` to exercise the register → login → CRUD cycle.
5. Repeat: edit → lint/type/test → `make check` → `make e2e` (if needed) → commit.

## Docker Workflow

```bash
make docker-build
make docker-up
make docker-down
```

Services & ports:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Postgres: `localhost:5432`

## OpenAPI & Type Generation

- Update `api/openapi.yaml` first whenever you touch API behavior, then regenerate downstream artifacts.
- Run `make gen` to refresh `frontend/types/api.ts` and `backend/internal/api/server.gen.go`.
- Run `make swagger` whenever you adjust OpenAPI metadata or docs, keeping `backend/docs/swagger.*` in sync.

## How To Add A New Module

Run `make new-module name=product`, then:

1. Update `api/openapi.yaml` first if the module exposes an API surface.
2. Implement handler → service → repository (DTOs live under `backend/internal/dto/`).
3. Wire providers/constructors via `backend/cmd/server/wire.go` and `providers.go`.
4. Register routes in `backend/cmd/server/main.go`.
5. Add GORM migrations (plugins call `AutoMigrate`) and seed data if needed.
6. Verify with `make check` and `make e2e` (if behavior changed) before merging.

## Roadmap

**Near-term (1‑2 months):** sharpen developer onboarding, shore up e2e coverage, and automate doc generation for latest APIs.
**Longer-term (quarterly horizons):** invest in AI-native ops (agent-friendly scripts, observability), scale modules for plugin scenarios, and explore multi-cluster Docker compose support.

## Documentation Map

- [AGENTS.md](AGENTS.md)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [CONVENTIONS.md](CONVENTIONS.md)
- [docs/README.md](docs/README.md)
