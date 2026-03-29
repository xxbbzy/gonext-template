# 文档入口 / Documentation Index

本目录面向人类贡献者，提供简洁的中英双语导航。
The `docs/` directory is the concise bilingual entrypoint for human contributors.

## AI 根文档 / Root AI Docs

这些文件位于仓库根目录，面向 AI 代理，保持英文且更偏执行规则。
These files live at the repository root, are English-only, and act as the operational reference for AI agents.

| 文件 / File                                | 作用 / Purpose                                                                                                     |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| [`../AGENTS.md`](../AGENTS.md)             | 任务入口、常见改动 playbook、验证清单 / Task flow, common change playbooks, verification checklist                 |
| [`../ARCHITECTURE.md`](../ARCHITECTURE.md) | 运行拓扑、请求链路、DI、数据库初始化、扩展点 / Runtime topology, request flow, DI, database init, extension points |
| [`../CONVENTIONS.md`](../CONVENTIONS.md)   | 命名、分层、错误处理、响应与测试规则 / Naming, layer boundaries, error handling, response and testing rules        |

## 人类文档 / Human-Facing Docs

| 文档 / Doc                                 | 适用场景 / When to read                                                                   |
| ------------------------------------------ | ----------------------------------------------------------------------------------------- |
| [ARCHITECTURE.md](./ARCHITECTURE.md)       | 快速理解实际技术栈与关键入口文件 / Fast summary of the actual stack and key entry points  |
| [DEVELOPMENT.md](./DEVELOPMENT.md)         | 日常开发、生成流程、文档维护 / Daily workflow, generation flow, documentation maintenance |
| [QUICK_START.md](./QUICK_START.md)         | 本地初始化与快速启动 / Local bootstrap and quick start                                    |
| [CONFIGURATION.md](./CONFIGURATION.md)     | 环境变量与运行配置 / Environment variables and runtime configuration                      |
| [API_GUIDE.md](./API_GUIDE.md)             | API 使用与调试 / API usage and debugging                                                  |
| [DATABASE.md](./DATABASE.md)               | 数据模型、迁移与数据库策略 / Data model, migrations, database strategy                    |
| [CI_CD.md](./CI_CD.md)                     | 持续集成与发布流水线 / CI and delivery pipeline                                           |
| [DEPLOYMENT.md](./DEPLOYMENT.md)           | 部署与容器化说明 / Deployment and container guidance                                      |
| [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md) | 版本与变更记录流程 / Versioning and changelog workflow                                    |
| [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) | 常见问题排查 / Troubleshooting                                                            |

## ADR / 架构决策记录

当你要理解“为什么这样做”，先看 ADR，而不是猜测或直接替换框架。
When you need the rationale behind the stack, check ADRs before proposing framework-level changes.

| ADR                                                                          | 主题 / Topic                                                                               |
| ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| [adr/001-openapi-as-contract.md](./adr/001-openapi-as-contract.md)           | `api/openapi.yaml` 是契约真相源 / `api/openapi.yaml` is the contract source of truth       |
| [adr/002-sqlite-dev-postgres-prod.md](./adr/002-sqlite-dev-postgres-prod.md) | 本地 SQLite、可部署环境 PostgreSQL / SQLite locally, PostgreSQL in deployable environments |
| [adr/003-wire-for-di.md](./adr/003-wire-for-di.md)                           | 使用 Google Wire 做依赖注入 / Using Google Wire for dependency injection                   |

## 关键入口 / Key Entry Points

- 后端启动入口：`backend/cmd/server/main.go`
- 依赖注入图：`backend/cmd/server/wire.go`
- 数据库初始化：`backend/internal/config/database.go`
- API 契约：`api/openapi.yaml`
- Frontend route tree: `frontend/app/`
- Frontend API client (preferred): `frontend/lib/api-client.gen.ts` (legacy/compat: `frontend/lib/api-client.ts`)

## 推荐阅读顺序 / Suggested Reading Order

1. 新贡献者 / New contributors: `README -> QUICK_START -> ARCHITECTURE -> DEVELOPMENT`
2. 后端改动 / Backend changes: `../AGENTS.md -> ../ARCHITECTURE.md -> API_GUIDE -> DATABASE`
3. 前端改动 / Frontend changes: `../AGENTS.md -> ARCHITECTURE -> DEVELOPMENT -> API_GUIDE`
4. 架构调整 / Architecture changes: `../ARCHITECTURE.md -> adr/ -> DEVELOPMENT`

## 维护规则 / Maintenance Rules

- 修改 API 行为/契约时，先更新 `api/openapi.yaml`；开发联调阶段可先运行 `make gen-types` 刷新 `frontend/types/api.ts`。提交/合并前需确保后端生成文件与 Swagger 同步：推荐执行 `make gen`（或至少 `make gen-server` + `make swagger`）。
- When API behavior/contract changes, update `api/openapi.yaml` first; during iteration you can run `make gen-types` to refresh `frontend/types/api.ts`. Before committing/merging, keep backend generated artifacts and Swagger in sync: prefer `make gen` (or at minimum `make gen-server` + `make swagger`).
- 根目录 AI 文档是结构和规则的主参考；`docs/` 负责双语摘要与导航。
- The root AI docs are canonical for structure and rules; `docs/` should summarize and link rather than duplicate them.
