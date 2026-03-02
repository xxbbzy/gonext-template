# 部署指南

## 1. Docker Compose 一键部署

```bash
make docker-up
```

默认服务：

1. backend：`http://localhost:8080`
2. frontend：`http://localhost:3000`
3. postgres：`localhost:5432`

停止服务：

```bash
make docker-down
```

## 2. 部署前检查

1. 设置生产级 `JWT_SECRET`。
2. 校验 `CORS_ALLOWED_ORIGINS`。
3. 确认 `DB_DRIVER/DB_DSN` 与目标环境一致。
4. 验证健康检查：`/healthz`、`/readyz`。

## 3. 镜像发布

- CI 在 tag（`v*`）或 Release 发布时自动构建并推送 GHCR 镜像。
- 镜像命名：
  - `ghcr.io/<repo>/backend:<tag>`
  - `ghcr.io/<repo>/frontend:<tag>`

## 4. 手动构建

```bash
make build
```

构建产物：

1. 后端二进制：`bin/server`
2. 前端产物：Next.js 生产构建目录

## 5. 版本注意事项

当前仓库存在 Go 版本配置漂移风险：

1. `backend/go.mod` 指定 Go `1.25.3`
2. `backend/Dockerfile` 的构建镜像为 `golang:1.21-alpine`

建议在正式部署前统一版本，避免容器构建与本地/CI 行为不一致。
