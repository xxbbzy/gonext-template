# 开发与维护 / Development and Maintenance

本页给人类贡献者一个简洁的工作流摘要；仓库级执行规则以根目录 `../AGENTS.md` 和 `../CONVENTIONS.md` 为准。
This page is the concise workflow summary for human contributors; repository-wide operational rules live in `../AGENTS.md` and `../CONVENTIONS.md`.

## 日常流程 / Daily Workflow

1. 初始化 / Bootstrap: `make init`
2. 启动开发环境 / Start dev servers: `make dev`
3. 进行改动 / Implement changes in the existing layer structure
4. 本地检查 / Run practical checks: `make check`
5. 生成产物（如需要）/ Refresh generated artifacts when needed:
   - `make swagger`
   - `make gen-types`
6. 更新必要文档 / Update the minimal docs affected by the change

## API 变更 / API Changes

- 先改 `api/openapi.yaml`，它是契约真相源。
- Update `api/openapi.yaml` first because it is the contract source of truth.
- 后端实现后，如果契约或 Swagger 输出受影响，执行 `make swagger`。
- After backend API changes, run `make swagger` when contract-derived backend docs must be refreshed.
- 前端类型依赖 OpenAPI，契约变更后执行 `make gen-types`。
- Frontend types depend on OpenAPI, so run `make gen-types` after contract changes.
- 相关背景见 `docs/adr/001-openapi-as-contract.md`。
- For rationale, see `docs/adr/001-openapi-as-contract.md`.

## 数据库变更 / Database Changes

- 本地默认 SQLite，可部署环境使用 PostgreSQL。
- SQLite is the local default, while deployable environments use PostgreSQL.
- 开发时的 `AutoMigrate` 只解决局部便利性，不替代发布用 migration。
- Development `AutoMigrate` is a convenience and does not replace release-oriented migrations.
- 涉及 schema 变更时，在 `backend/migrations/` 添加 migration，并考虑两种数据库方言差异。
- When schema changes are introduced, add migrations under `backend/migrations/` and check dialect differences.
- 设计理由见 `docs/adr/002-sqlite-dev-postgres-prod.md`。
- See `docs/adr/002-sqlite-dev-postgres-prod.md` for the rationale.

## 依赖注入与运行时 / Dependency Injection and Runtime

- 新增后端依赖时，优先查看 `backend/cmd/server/wire.go`、`backend/cmd/server/providers.go`、`backend/cmd/server/main.go`。
- For new backend dependencies, start with `backend/cmd/server/wire.go`, `backend/cmd/server/providers.go`, and `backend/cmd/server/main.go`.
- 不要直接编辑 `backend/cmd/server/wire_gen.go`。
- Do not edit `backend/cmd/server/wire_gen.go` directly.
- 相关原因见 `docs/adr/003-wire-for-di.md`。
- See `docs/adr/003-wire-for-di.md` for the decision record.

## 文档维护 / Documentation Maintenance

- 根目录文档 `../AGENTS.md`、`../ARCHITECTURE.md`、`../CONVENTIONS.md` 面向 AI，保持英文。
- Root docs `../AGENTS.md`, `../ARCHITECTURE.md`, and `../CONVENTIONS.md` are AI-facing and stay English-only.
- `docs/` 中的精选文档保持中英双语，并尽量做摘要与导航，不复制整套操作说明。
- Curated docs in `docs/` stay bilingual and should summarize or link rather than duplicate the full operational guidance.
- 架构级决策放到 `docs/adr/`，不要只散落在 PR 或代码注释里。
- Repository-wide architecture decisions should be captured in `docs/adr/`, not only in PR text or code comments.

## 参考入口 / Recommended References

- AI 执行入口 / AI operational entrypoint: `../AGENTS.md`
- 架构总览 / Architecture overview: `../ARCHITECTURE.md` and [ARCHITECTURE.md](./ARCHITECTURE.md)
- 规范与边界 / Rules and boundaries: `../CONVENTIONS.md`
- ADR / Architecture decisions: `./adr/`
- API 契约 / API contract: `../api/openapi.yaml`
