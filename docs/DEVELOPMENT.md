# 团队统一开发规范

本规范是项目开发与协作的统一标准。若与其他文档冲突，以本文件为准。

配套文档：

- [快速开始](./QUICK_START.md)
- [架构说明](./ARCHITECTURE.md)
- [API 指南](./API_GUIDE.md)
- [数据库说明](./DATABASE.md)
- [CI/CD 说明](./CI_CD.md)
- [Changelog 指南](./CHANGELOG_GUIDE.md)

## 1. 目标与原则

- 目标：保证迭代速度的同时，降低分支漂移、环境不一致、接口不一致、数据库变更风险。
- 原则：
  - 契约优先：接口变更先更新 OpenAPI。
  - 小步提交：单功能单 PR，避免“巨型变更”。
  - 本地先过门禁：提交前必须通过 `make check`。
  - 文档即流程：命令、CI、文档三者保持一致。

## 2. 分支与发布策略

- `main`：生产稳定分支，受保护。
- `develop`：日常集成分支。
- `feature/<name>`：新功能分支，从 `develop` 切出。
- `fix/<name>`：缺陷修复分支，从 `develop` 切出。
- `hotfix/<name>`：线上紧急修复分支，从 `main` 切出，合并回 `main` 后必须回合并到 `develop`。

标准流程：

1. 从 `develop` 创建 `feature/*` 或 `fix/*`。
2. 本地开发并自测。
3. 提交 PR 到 `develop`。
4. `develop` 验证稳定后，按版本节奏合并到 `main` 并打 `v*` tag。

## 3. 环境基线

- Go 版本：以 `backend/go.mod` 为准（CI 读取该文件）。
- Node.js：18+。
- 默认本地数据库：SQLite（`.env` 默认 `DB_DRIVER=sqlite`）。
- 容器集成环境：PostgreSQL（`docker-compose.yml`）。

## 4. 日常开发流程

1. 初始化：`make init`
2. 启动联调：`make dev`
3. 开发迭代：后端遵循 `handler -> service -> repository -> model`
4. 本地质量门禁：`make check`
5. 提交：Conventional Commits
6. 创建 PR：目标分支 `develop`

建议：每个 PR 控制在单一业务目标内，避免同时混入重构和功能开发。

## 5. API 变更流程

必须按顺序执行：

1. 修改 `api/openapi.yaml`
2. 执行 `make gen-types` 同步前端类型
3. 实现后端 handler/service/repository
4. 前端联调并自测
5. `make check` 通过后提交

禁止：后端接口已变更但未同步 OpenAPI 或前端类型。

## 6. 数据库变更流程

### 6.1 本地开发

- 本地默认 SQLite，开发模式可通过 `AutoMigrate` 提升效率。

### 6.2 可发布变更

涉及 schema 变更时，必须提供 migration：

1. `make new-migration name=<change_name>`
2. 补全 `up/down.sql`
3. 在 PR 说明中写清回滚方案

### 6.3 风险控制

当前仓库存在 SQLite（本地）和 PostgreSQL（容器）两种数据库形态。迁移脚本必须在目标数据库验证，禁止“只在一种数据库跑通就合并”。

## 7. 质量门禁与 CI

### 7.1 本地门禁

提交前必须通过 `make check`：

1. `make lint`
2. `make typecheck-frontend`
3. `make test`

### 7.2 CI 门禁

- Backend CI：lint -> test -> build
- Frontend CI：lint -> typecheck -> build
- CI 对 `develop`、`main` 的 push/PR 生效（按路径过滤）。

## 8. 提交与评审规范

- Commit message：Conventional Commits（`feat(scope): subject`）
- PR 必须包含：
  1. 变更背景
  2. 影响范围
  3. 测试说明
  4. 回滚方式（若涉及 DB/API）
- 至少 1 名 reviewer 通过后合并

## 9. 发布流程

发布记录维护要求见 [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md)。

1. `develop` 达到发布条件后合并到 `main`
2. 整理并更新 `CHANGELOG.md`（从 `Unreleased` 切出版本条目）
3. 在 `main` 打 tag：`vX.Y.Z`
4. 触发 Docker Build 工作流构建并推送镜像

## 10. 文档维护要求

以下变更必须同步更新 `docs/`：

1. 新增/修改 Make 命令
2. 新增/修改环境变量
3. 新增/修改 API
4. 新增/修改 CI 流程
