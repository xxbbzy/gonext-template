# 快速开始

## 1. 环境要求

- Go：以 `backend/go.mod` 为准（当前为 `1.25.3`）
- Node.js：`18+`（CI 使用 18）
- npm：随 Node.js 安装
- Make

可选但常用的本地工具：

- `golangci-lint`（`make lint-backend` 依赖）
- `migrate`（`make migrate-*`、`make new-migration` 依赖）
- `swag`（`make swagger` 依赖）

## 2. 初始化

```bash
git clone <your-repo-url>
cd gonext-template
make init
```

`make init` 会执行：

1. 若 `.env` 不存在，则从 `.env.example` 复制
2. 下载后端 Go 依赖
3. 安装前端 npm 依赖
4. 创建 `data/` 与 `uploads/` 目录

## 3. 启动开发环境

```bash
make dev
```

等价于并行启动：

1. 后端：`make dev-backend`（默认 `http://localhost:8080`）
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

1. `make lint`
2. `make typecheck-frontend`
3. `make test`

## 6. 常用命令速查

```bash
make help            # 查看所有命令
make gen-types       # 基于 OpenAPI 生成前端类型
make new-module name=product
make new-migration name=add_xxx
make docker-up       # 用 docker compose 启动
make docker-down
```
