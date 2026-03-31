# 快速开始

## 1. 环境要求

- Go：以 `backend/go.mod` 为准（要求 `1.25+`）
- Node.js：`20+`（CI 使用 20）
- npm：随 Node.js 安装
- Make

可选但常用的本地工具：

- `golangci-lint`（`make lint-backend` 依赖）
- `migrate`（`make migrate-*`、`make new-migration` 依赖）
- `make swagger` 无需额外 CLI（由 `go run ../scripts/swagger/main.go` 执行）

## 2. 初始化

```bash
git clone <your-repo-url>
cd gonext-template
make init
```

`make init` 会执行：

1. 若 `.env` 不存在，则从 `.env.example` 复制
2. 若仓库根目录存在 `package.json`，执行根目录 `npm install`
3. 下载后端 Go 依赖
4. 安装前端 npm 依赖
5. 创建 `data/` 与 `uploads/` 目录
6. 初始化数据库（`go run ./cmd/bootstrap`）
7. 生成代码与文档（`make gen`，包含 `gen-server`、`gen-types`、`swagger`）
8. Generate code and documentation (`make gen`, includes `gen-server`, `gen-types`, `swagger`)

## 3. 启动开发环境

```bash
make dev
```

等价于并行启动：

1. 后端：`make dev-backend`（默认 `http://localhost:8080`，`make dev` 会通过 `make -j2 dev-backend dev-frontend` 触发）
2. 前端：`make dev-frontend`（默认 `http://localhost:3000`）

## 4. 初始化测试数据（可选）

```bash
make seed
```

默认会写入示例用户：

- 管理员：`admin@example.com / admin123`
- 普通用户：`user@example.com / user123`

## 5. 运行质量门禁

```bash
make check
```

执行内容：

1. `make lint`（包含 `lint-backend`、`lint-frontend` 与 `check-architecture`）
2. `make typecheck-frontend`
3. `make test`（包含 `test-backend`、`test-frontend` 与 `test-tooling`）
4. `make build`（包含 `build-backend` 与 `build-frontend`）

## 6. 常用命令速查

```bash
make help            # 查看所有命令 / View all commands
make gen-types        # 基于 OpenAPI 生成前端 TypeScript 类型（frontend/types/api.ts，适合联调迭代）
                     # Generate frontend TypeScript types from OpenAPI (frontend/types/api.ts, suitable for iteration)
make gen              # 全量生成：Go server stub + 前端类型 + Swagger（gen-server + gen-types + swagger，提交/合并前推荐）
                     # Full regeneration: Go server stub + frontend types + Swagger (gen-server + gen-types + swagger, recommended before commit/merge)
make new-module name=product
                     # 生成约定对齐的后端模块模板、基础测试与后续检查清单
                     # Generate a convention-aligned backend module scaffold, baseline tests, and a follow-up checklist
make new-migration name=add_xxx
make docker-up       # 用 docker compose 启动 / Start with docker compose
make docker-down
```
