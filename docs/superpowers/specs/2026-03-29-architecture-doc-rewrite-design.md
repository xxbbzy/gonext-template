# Architecture document rewrite design

## Context

- The root `ARCHITECTURE.md` needs to shift from a topology-only summary toward an explicit guide that covers runtime topology, backend layer boundaries, dependency direction, and extension points for AI agents.
- The new text must steer readers to `AGENTS.md`, `CONVENTIONS.md`, `docs/README.md`, and `docs/adr/` without duplicating their full contents.
- Layer rules must reference the actual directories (`backend/internal/handler/`, `service/`, `repository/`, `model/`, `dto/`) and acknowledge how `backend/cmd/server/main.go` wires routes and middleware.

## Goals

1. Reframe the intro so it states the file’s remit and points to the adjacent guiding documents.
2. Add a System Topology section that traces the request/data path from `api/openapi.yaml` through the frontend, middleware, handlers, and database initialization.
3. Add a Backend Layer Responsibilities section with subsections for handler/service/repository/model/dto; each subsection will describe ownership, dependencies, prohibitions, file location, and typical changes when a module is added.
4. Keep the rewrite compact, using short paragraphs and rule-style bullets.

## Structure

- **Intro**: single paragraph listing the four coverage areas and linking to the required references.
- **System Topology**: inline path that includes `frontend/app/`, `frontend/lib/api-client.ts` + `.gen.ts`, `frontend/lib/query-provider.tsx`, `frontend/stores/auth.ts`, `backend/cmd/server/main.go`, middleware, handler→service→repository→model, and `backend/internal/config/database.go`.
- **Backend Layer Responsibilities**: five subsections, each with 5 concise bullets (owns, depends on, must not do, location, addition impact). Mention `backend/cmd/server/main.go` within the handler discussion to keep wiring context front-and-center.
- **Extension notes**: emphasize middleware wiring, DI via `backend/cmd/server/wire.go`/`providers.go`, and database setup, keeping the entire doc tightly focused on actionable layers rather than a handbook.

## Execution steps

1. Replace the existing intro, System Shape, Request Flow, DI, Database, Middleware, and Frontend sections with the planned structure to avoid duplication.
2. Keep the new text formatted with short sentences and bullet lists that stay under four lines per paragraph.
3. Double-check that every required anchor appears (handler/service/repository/model/dto directories plus `backend/cmd/server/main.go`).

## Verification

- Run `make check` after the edit per `AGENTS.md`.
- Note that `make check` already covers lint, typecheck, test, and build for both backend and frontend.

## Notes

- This doc is the AI-facing map; human readers rely on the bilingual `docs/` folder. No ADR or coding convention details should be moved here, only the links.
