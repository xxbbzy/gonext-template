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
- Contract changes that are merged must keep committed backend-derived outputs
  in sync (Go server stubs and Swagger), not only frontend types.
- 合并的契约变更必须保持已提交的后端派生产物（Go server stubs 和 Swagger）同步，而不仅仅是前端类型。

## Alternatives Considered

- Handler code comments as the primary source: rejected because they do not
  provide a complete frontend-facing contract.
- Frontend types as the primary source: rejected because they are generated and
  omit backend operational intent.

## References

- `api/openapi.yaml`
- `backend/docs/swagger.yaml`
- `frontend/types/api.ts`
- `Makefile` targets: `gen-types` (frontend TypeScript types) and `gen` (full regen: `gen-server`, `gen-types`, `swagger`). Older docs may refer to `gen-client`; treat it as `gen-types`.
- `Makefile` 目标：`gen-types`（前端 TypeScript 类型）和 `gen`（完整重新生成：`gen-server`、`gen-types`、`swagger`）。旧文档可能引用 `gen-client`；将其视为 `gen-types`。