# 架构速览 / Architecture Summary

本页给人类贡献者一个简洁的技术栈地图；更完整的 AI 执行视图在仓库根目录的 `../ARCHITECTURE.md`。
This page is the concise contributor summary; the fuller AI operational map lives in the root `../ARCHITECTURE.md`.

## 系统拓扑 / System Topology

```text
Next.js App Router
  -> frontend/lib/api-client.ts
     -> Gin server (backend/cmd/server/main.go)
        -> handler -> service -> repository -> GORM
           -> SQLite or PostgreSQL
```

- API 契约文件是 `api/openapi.yaml`。
- The API contract file is `api/openapi.yaml`.
- 前端页面位于 `frontend/app/`，共享请求逻辑位于 `frontend/lib/api-client.ts`。
- Frontend routes live in `frontend/app/`, and shared request logic lives in `frontend/lib/api-client.ts`.

## 后端运行时 / Backend Runtime

- 启动入口：`backend/cmd/server/main.go`
- Entry point: `backend/cmd/server/main.go`
- DI 声明：`backend/cmd/server/wire.go`，辅助 provider 在 `backend/cmd/server/providers.go`
- DI graph declaration: `backend/cmd/server/wire.go`, with helper providers in `backend/cmd/server/providers.go`
- 数据库初始化：`backend/internal/config/database.go`
- Database initialization: `backend/internal/config/database.go`

当前后端全局中间件顺序：
Current global middleware order:

1. `Recovery`
2. `RequestLogger`
3. `ErrorHandler`
4. `CORS`

认证与限流挂在路由组上，而不是所有请求全局生效。
Authentication and rate limiting are attached to route groups rather than applied globally.

## 前端运行时 / Frontend Runtime

- `frontend/app/layout.tsx` 负责 `next-intl`、TanStack Query 和 toast provider。
- `frontend/app/layout.tsx` sets up `next-intl`, TanStack Query, and the toast provider.
- `frontend/stores/auth.ts` 用 Zustand 持久化 token。
- `frontend/stores/auth.ts` persists auth state with Zustand.
- `frontend/lib/api-client.ts` 自动附带 Bearer Token，并在 `401` 时尝试刷新。
- `frontend/lib/api-client.ts` attaches bearer tokens and attempts token refresh on `401`.

## 数据与迁移 / Data and Migrations

- 本地默认使用 SQLite，减少初始化成本。
- SQLite is the default local database to keep bootstrap lightweight.
- 可部署环境使用 PostgreSQL，详见 `docker-compose.yml` 与 `docs/adr/002-sqlite-dev-postgres-prod.md`。
- Deployable environments use PostgreSQL; see `docker-compose.yml` and `docs/adr/002-sqlite-dev-postgres-prod.md`.
- 开发模式下会在启动时执行有限的 `AutoMigrate`，但可发布 schema 变更仍应落到 `backend/migrations/`。
- Development mode performs limited `AutoMigrate`, but release-oriented schema changes should still use `backend/migrations/`.

## 扩展点 / Extension Points

- 新增后端模块：查看 `../AGENTS.md` 和 `../CONVENTIONS.md`
- Adding a backend module: start with `../AGENTS.md` and `../CONVENTIONS.md`
- 新增 API：先改 `api/openapi.yaml`，再同步 handler/service/repository 与生成产物
- Adding or changing an API: update `api/openapi.yaml` first, then sync handler/service/repository and generated artifacts
- 新增前端页面：在 `frontend/app/` 建 route，并复用 `frontend/lib/api-client.ts`
- Adding a frontend page: create a route in `frontend/app/` and reuse `frontend/lib/api-client.ts`

## 继续阅读 / Continue Reading

- 根目录 AI 文档 / Root AI docs: `../AGENTS.md`, `../ARCHITECTURE.md`, `../CONVENTIONS.md`
- ADR / Decision records: `./adr/`
- 开发与维护流程 / Workflow and maintenance: [DEVELOPMENT.md](./DEVELOPMENT.md)
