# ADR 001: OpenAPI as the Contract Source of Truth

- Status: Accepted
- Date: 2026-03-07

## Context

The repository serves a Gin backend and a Next.js frontend that share request
and response expectations. Without a single contract source, handler comments,
generated Swagger assets, frontend request typing, and human documentation drift
quickly.

The current project already stores the contract at `api/openapi.yaml` and uses
that file to generate frontend types.

## Decision

`api/openapi.yaml` is the canonical API contract for the repository.

- Backend API changes must update `api/openapi.yaml`.
- Generated outputs such as Swagger artifacts and frontend types are derived
  artifacts, not primary sources.
- Curated docs may summarize API behavior, but they should link back to the
  contract rather than restate it in full.

## Consequences

- Backend and frontend changes have a clear synchronization point.
- AI agents and contributors can inspect one file to understand public API
  behavior.
- Contract updates add a small amount of extra workflow overhead.
- Generated artifacts must be refreshed after contract changes to avoid stale
  type information.

## Alternatives Considered

- Handler code comments as the primary source: rejected because they do not
  provide a complete frontend-facing contract.
- Frontend types as the primary source: rejected because they are generated and
  omit backend operational intent.

## References

- `api/openapi.yaml`
- `backend/docs/swagger.yaml`
- `frontend/types/api.ts`
- `Makefile` targets: `swagger`, `gen-types`
