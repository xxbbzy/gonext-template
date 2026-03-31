# 开发与维护 / Development and Maintenance

本页给人类贡献者一个简洁的工作流摘要；仓库级执行规则以根目录 `../AGENTS.md` 和 `../CONVENTIONS.md` 为准。
This page is the concise workflow summary for human contributors; repository-wide operational rules live in `../AGENTS.md` and `../CONVENTIONS.md`.

## 日常流程 / Daily Workflow

1. 初始化 / Bootstrap: `make init`
2. 启动开发环境 / Start dev servers: `make dev`
3. 进行改动 / Implement changes in the existing layer structure
4. 本地检查 / Run practical checks: `make check`
   - 现在同时包含架构 guardrails 与脚手架回归检查。
   - This now includes architecture guardrails and scaffold regression coverage.
5. 生成产物（如需要）/ Refresh generated artifacts when needed:
   - `make gen-types`
   - `make swagger`
   - `make gen`
6. 更新必要文档 / Update the minimal docs affected by the change

## 后端模块脚手架 / Backend Module Scaffolding

- 新增后端模块时，优先执行 `make new-module name=<module>`，它会生成与仓库约定对齐的 handler/service/repository/model/dto 模板、基础测试和后续检查清单。
- When adding a backend module, start with `make new-module name=<module>`; it generates convention-aligned handler/service/repository/model/dto templates, baseline tests, and a follow-up checklist.
- 生成后仍需回到 `api/openapi.yaml`、Wire、路由注册、AutoMigrate/migrations 与验证命令，把模板补齐成真实功能。
- After scaffolding, you still need to finish the OpenAPI contract, Wire registration, route mounting, AutoMigrate/migrations, and verification commands.

## API 变更 / API Changes

- 先改 `api/openapi.yaml`，它是契约真相源。
- Update `api/openapi.yaml` first because it is the contract source of truth.
- 契约变更后，默认先执行 `make gen-types` 刷新前端类型（输出 `frontend/types/api.ts`）。
- After contract changes, run `make gen-types` as the default step to refresh frontend types (writes `frontend/types/api.ts`).
- 准备提交/合并包含契约变更的 PR 时，确保后端派生文件也已同步：推荐执行 `make gen`；或至少执行 `make gen-server`（并在需要更新 Swagger 输出时执行 `make swagger`）。
- Before committing/merging a PR that includes contract changes, keep committed backend-derived artifacts in sync: prefer `make gen`; or at minimum `make gen-server` (and run `make swagger` when Swagger output must be refreshed).
- 在推送涉及 OpenAPI/代码生成输入的改动前，执行 `make check-codegen-drift`，它会复用 CI 的漂移检查规则（先生成，再校验仓库状态）。
- Before pushing changes that touch OpenAPI/codegen inputs, run `make check-codegen-drift`; it uses the same CI drift check semantics (regenerate first, then verify repo state).
- 若 `make check-codegen-drift` 失败，先执行 `make gen`，再检查 `git status` 并提交生成文件，最后重新运行 `make check-codegen-drift` 直到通过。
- If `make check-codegen-drift` fails, run `make gen`, review `git status`, commit generated artifacts, then rerun `make check-codegen-drift` until it passes.
- 仅需刷新 Swagger 文档输出时，执行 `make swagger`。
- If you only need to refresh Swagger docs output, run `make swagger`.
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
- 影响 backend/API/runtime 路径的 PR 现在会在 `CI / Quality Gate` 中运行 `make e2e`；`CI / Merge Validation` 继续提供合并后的 smoke 覆盖。
- PRs that touch backend/API/runtime paths now run `make e2e` in `CI / Quality Gate`; `CI / Merge Validation` still provides post-merge smoke coverage.
- 相关原因见 `docs/adr/003-wire-for-di.md`。
- See `docs/adr/003-wire-for-di.md` for the decision record.

## 文档维护 / Documentation Maintenance

- 根目录文档 `../AGENTS.md`、`../ARCHITECTURE.md`、`../CONVENTIONS.md` 面向 AI，保持英文。
- Root docs `../AGENTS.md`, `../ARCHITECTURE.md`, and `../CONVENTIONS.md` are AI-facing and stay English-only.
- `docs/` 中的精选文档保持中英双语，并尽量做摘要与导航，不复制整套操作说明。
- Curated docs in `docs/` stay bilingual and should summarize or link rather than duplicate the full operational guidance.
- 架构级决策放到 `docs/adr/`，不要只散落在 PR 或代码注释里。
- Repository-wide architecture decisions should be captured in `docs/adr/`, not only in PR text or code comments.

## 文档一致性检查清单 / Documentation Consistency Checklist

- 修改 `Makefile` 命令行为时，同步检查并更新 `README.md`、`docs/QUICK_START.md`、`docs/DEPLOYMENT.md` 中的相关说明。
- When `Makefile` command behavior changes, update the matching descriptions in `README.md`, `docs/QUICK_START.md`, and `docs/DEPLOYMENT.md`.

- 修改 `docker-compose.yml` 时涉及服务、端口或部署约定的更改，必须同步刷新部署/服务/端口相关文档（如部署指南或运行指标）。
- When `docker-compose.yml` changes involve services, ports, or deployment conventions, sync the corresponding deployment/service/port docs.

- 修改依赖注入策略或运行时装配（例如 Wire provider/wiring）时，同步更新 `README.md` 中的概要说明以及相关架构章节或 ADR 引用。
- When the DI strategy or runtime wiring (e.g., Wire providers/wiring) changes, refresh the README guidance plus the linked architecture/ADR references.

- 模块目录内的 README 不应保留默认脚手架模板内容（例如 create-next-app 默认文本），必须描述本仓库的实际使用方式。
- Do not commit default scaffold README templates in module directories; every README must describe how this repository actually works.

## 参考入口 / Recommended References

- AI 执行入口 / AI operational entrypoint: `../AGENTS.md`
- 架构总览 / Architecture overview: `../ARCHITECTURE.md` and [ARCHITECTURE.md](./ARCHITECTURE.md)
- 规范与边界 / Rules and boundaries: `../CONVENTIONS.md`
- ADR / Architecture decisions: `./adr/`
- API 契约 / API contract: `../api/openapi.yaml`
