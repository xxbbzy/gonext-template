# CI/CD 说明

工作流目录：`.github/workflows`

## 1. Backend CI

文件：`backend-ci.yml`

触发条件：

1. `push` 到 `main/develop` 且变更 `backend/**`
2. `pull_request` 到 `main/develop` 且变更 `backend/**`

执行阶段：

1. `lint`：`golangci-lint`
2. `test`：`go test -v -race -coverprofile=coverage.out ./...`
3. `build`：`go build -o server ./cmd/server/`

## 2. Frontend CI

文件：`frontend-ci.yml`

触发条件：

1. `push` 到 `main/develop` 且变更 `frontend/**`
2. `pull_request` 到 `main/develop` 且变更 `frontend/**`

执行阶段：

1. `lint`：`npm run lint`
2. `typecheck`：`npm run typecheck`
3. `build`：`npm run build`

## 3. Docker Build

文件：`docker-build.yml`

触发条件：

1. `push` tag：`v*`
2. `release` 发布

行为：

1. 构建并推送 backend 镜像到 `ghcr.io/<repo>/backend:<tag>`
2. 构建并推送 frontend 镜像到 `ghcr.io/<repo>/frontend:<tag>`

## 4. 本地对齐 CI 的建议命令

```bash
make check
cd backend && go test -v -race ./...
cd frontend && npm run build
```

## 5. 分支与质量门禁

推荐分支策略与评审要求见 [DEVELOPMENT.md](./DEVELOPMENT.md)。
版本号与变更日志维护见 [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md)。
