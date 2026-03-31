# 部署指南 / Deployment Guide

## 1. Docker Compose 一键部署 / One-command Compose startup

```bash
make docker-up
```

`make docker-up` 是 `docker compose up -d` 的封装，会启动 `docker-compose.yml` 中的三个服务：

- `backend`：对外暴露 `http://localhost:8080`（映射 `8080:8080`）
- `frontend`：对外暴露 `http://localhost:3000`（映射 `3000:3000`）
- `db`：Postgres 服务，`backend` 在容器网络内通过 `db:5432` 访问；`5432:5432` 映射主要用于宿主机直连与调试

`make docker-up` wraps `docker compose up -d` and starts the same three services above with published ports for local access and debugging.

停止服务：

```bash
make docker-down
```

该命令会执行 `docker compose down`，并停止上述三个服务。默认会停止并移除服务容器与网络，但不会删除卷；例如 `pgdata` 会保留。若需同时删除卷，请显式使用 `docker compose down -v`（或 `--volumes`）。

`make docker-down` wraps `docker compose down`. It removes containers and the Compose network by default, but leaves named volumes such as `pgdata` intact unless you add `-v`.

## 2. 健康检查与依赖关系 / Health checks and dependency flow

- `db` 使用 `pg_isready` 作为容器健康检查；`backend` 会等待 `db` 进入 `service_healthy`。
- `backend` 使用 `/readyz` 作为容器健康检查；`/healthz` 仅表示进程存活，`/readyz` 只有在数据库可用时才返回 `200`。
- `frontend` 使用自身的 `/healthz` 路由作为容器健康检查，并通过 `depends_on.condition: service_healthy` 等待 `backend` 就绪，而不是只按启动顺序串联。
- 启动后可用 `docker compose ps` 查看 `healthy` 状态，用 `docker compose logs -f backend frontend db` 观察探针与启动日志。

- `db` uses `pg_isready` for its container health signal, and `backend` waits until that signal is `service_healthy`.
- `backend` uses `/readyz` for container health. `/healthz` is a lightweight liveness endpoint, while `/readyz` returns `200` only when the database-backed runtime is ready.
- `frontend` uses its own `/healthz` route for self-health and waits on backend `service_healthy`, so frontend health stays frontend-scoped while backend readiness still gates startup.
- After startup, use `docker compose ps` to inspect health states and `docker compose logs -f backend frontend db` to watch probe behavior.

## 3. 生产镜像运行时约定 / Production image expectations

- 部署构建环境建议使用 Go 1.25+ 工具链，并与 `backend/go.mod` 保持一致。
- The deployment build environment should use Go 1.25+ and stay aligned with `backend/go.mod`.

- `backend/Dockerfile` 使用与 `backend/go.mod` 对齐的 Go 1.25 builder，多阶段构建，并在 Alpine 运行时镜像中以非 root 用户运行。
- 后端运行时会预创建 `/app/data` 与 `/app/uploads`，用于 SQLite/上传等可写路径；Compose 仍将 `uploads` 挂载为命名卷。
- `frontend/Dockerfile` 使用 Node 20，多阶段构建，并依赖 Next.js `standalone` 产物来缩小运行时镜像内容。
- 前端运行时镜像以非 root `nextjs` 用户启动，只包含 `public/`、`.next/static` 与 standalone server 所需文件。

- `backend/Dockerfile` now uses a Go 1.25 builder aligned with `backend/go.mod`, keeps a small Alpine runtime stage, and runs the server as a non-root user.
- The backend runtime image prepares `/app/data` and `/app/uploads` as writable paths, while Compose still mounts `uploads` as a named volume.
- `frontend/Dockerfile` uses Node 20, builds a minimized Next.js standalone bundle, and runs the production server as a non-root `nextjs` user.
- The frontend runtime image only carries `public/`, `.next/static`, and the standalone server output needed to serve production traffic.

## 4. 部署前检查 / Pre-deployment checklist

1. 设置生产级 `JWT_SECRET`。
2. 校验 `CORS_ALLOWED_ORIGINS`。
3. 确认 `DB_DRIVER/DB_DSN` 与目标环境一致。
4. 验证 `docker compose up --build` 后 `db`、`backend`、`frontend` 都进入预期健康状态。
5. 验证 `/healthz` 与 `/readyz` 的返回码语义符合预期。
6. 若使用对象存储，确认 `STORAGE_DRIVER=s3` 且 `S3_BUCKET/S3_REGION/S3_ACCESS_KEY_ID/S3_SECRET_ACCESS_KEY` 已设置。
7. 若使用 MinIO，确认 `S3_ENDPOINT` 指向 MinIO 地址，并启用 `S3_FORCE_PATH_STYLE=true`（通常还需要 `S3_USE_SSL=false`）。

6. Set a production-grade `JWT_SECRET`.
7. Verify `CORS_ALLOWED_ORIGINS`.
8. Confirm `DB_DRIVER/DB_DSN` matches the target environment.
9. Validate that `docker compose up --build` reaches the expected healthy states for `db`, `backend`, and `frontend`.
10. Confirm the `/healthz` and `/readyz` status-code semantics match the deployment expectation.
11. For object storage, ensure `STORAGE_DRIVER=s3` and required S3 credentials/settings are configured.
12. For MinIO, set `S3_ENDPOINT`, `S3_FORCE_PATH_STYLE=true`, and usually `S3_USE_SSL=false`.

## 5. 上传存储部署模式 / Upload storage deployment modes

- `STORAGE_DRIVER=local`：上传文件落地到本地目录（容器中通常是 `/app/uploads`），通过后端 `/uploads/...` 暴露。
- `STORAGE_DRIVER=s3`：上传文件写入对象存储；后端不再依赖本地 `/uploads` 静态路由，返回 URL 由 S3 配置或 `UPLOAD_PUBLIC_BASE_URL` 决定。

- `STORAGE_DRIVER=local`: uploads are stored on local disk (typically `/app/uploads` in container) and served via backend `/uploads/...`.
- `STORAGE_DRIVER=s3`: uploads are stored in object storage; backend no longer relies on local `/uploads` static serving, and response URLs come from S3 config or `UPLOAD_PUBLIC_BASE_URL`.

### MinIO-compatible example / MinIO 兼容示例

```env
STORAGE_DRIVER=s3
S3_BUCKET=gonext-uploads
S3_REGION=us-east-1
S3_ENDPOINT=http://minio:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_PREFIX=uploads
S3_USE_SSL=false
S3_FORCE_PATH_STYLE=true
```

## 6. 镜像发布与手动构建 / Image publishing and manual builds

- CI 在 tag（`v*`）或 Release 发布时自动构建并推送 GHCR 镜像：
  - `ghcr.io/<repo>/backend:<tag>`
  - `ghcr.io/<repo>/frontend:<tag>`
- 本地手动构建可运行：

```bash
make build
```

- 构建产物包括后端二进制 `bin/server` 和前端 Next.js 生产构建目录。

- CI builds and publishes GHCR images on tags (`v*`) or Releases.
- For local production builds, run `make build`.
- The build artifacts are the backend binary at `bin/server` and the frontend production build output.
