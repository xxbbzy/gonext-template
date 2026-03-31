# 配置说明

## 1. 配置加载规则

后端配置加载逻辑在 `backend/internal/config/config.go`：

1. 读取 `.env`（若存在）
2. 读取系统环境变量
3. 对未提供项使用默认值

前端通过环境变量 `NEXT_PUBLIC_API_URL` 指向后端地址。

## 2. 后端环境变量

| 变量 Variable            | 默认值 Default                                                        | 说明 Description                                                                                                                            |
| ------------------------ | --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `APP_NAME`               | `gonext-template`                                                     | 服务名 / Service name                                                                                                                       |
| `APP_ENV`                | `development`                                                         | 环境（影响日志、自动迁移等） / Environment (affects logging, auto-migration, etc.)                                                          |
| `APP_PORT`               | `8080`                                                                | HTTP 监听端口 / HTTP listen port                                                                                                            |
| `DB_DRIVER`              | `sqlite`                                                              | 数据库驱动：`sqlite`/`postgres` / Database driver: `sqlite`/`postgres`                                                                      |
| `DB_DSN`                 | `./data/app.db`                                                       | 数据库连接串 / Database connection string                                                                                                   |
| `JWT_SECRET`             | `change-me-in-production`                                             | JWT 签名密钥 / JWT signing secret                                                                                                           |
| `JWT_ACCESS_EXPIRY`      | `15m`                                                                 | Access Token 有效期 / Access token expiry                                                                                                   |
| `JWT_REFRESH_EXPIRY`     | `168h`                                                                | Refresh Token 有效期 / Refresh token expiry                                                                                                 |
| `CORS_ALLOWED_ORIGINS`   | `http://localhost:3000`                                               | 允许来源，逗号分隔 / Allowed origins, comma-separated                                                                                       |
| `RATE_LIMIT_REQUESTS`    | `100`                                                                 | 单窗口内允许请求数 / Allowed requests per window                                                                                            |
| `RATE_LIMIT_DURATION`    | `1m`                                                                  | 限流窗口 / Rate limit window                                                                                                                |
| `UPLOAD_MAX_SIZE`        | `10485760`                                                            | 上传大小上限（字节） / Max upload size (bytes)                                                                                              |
| `UPLOAD_DIR`             | `./uploads`                                                           | 上传目录 / Upload directory                                                                                                                 |
| `UPLOAD_ALLOWED_TYPES`   | `.jpg,.jpeg,.png,.gif,.pdf,.doc,.docx`                                | 上传后缀白名单（每个后缀必须有内置 MIME 兼容规则） / Allowed upload extensions (each extension must have built-in MIME compatibility rules) |
| `UPLOAD_PUBLIC_BASE_URL` | 空（可选） / empty (optional)                                          | 上传文件对外访问地址前缀；未设置时 local 模式回退到 `APP_BASE_URL`，S3 模式按存储配置推导 URL / Public base URL override for uploads; when unset local falls back to `APP_BASE_URL`, while S3 mode derives URL from storage settings |
| `STORAGE_DRIVER`         | `local`                                                               | 上传存储驱动：`local` 或 `s3` / Upload storage driver: `local` or `s3`                                                                       |
| `S3_BUCKET`              | 空 / empty                                                            | `STORAGE_DRIVER=s3` 时必填：S3 bucket 名称 / Required when `STORAGE_DRIVER=s3`: S3 bucket name                                              |
| `S3_REGION`              | 空 / empty                                                            | `STORAGE_DRIVER=s3` 时必填：S3 region / Required when `STORAGE_DRIVER=s3`: S3 region                                                        |
| `S3_ENDPOINT`            | 空 / empty                                                            | 可选，自定义 S3 endpoint（MinIO 等） / Optional custom S3 endpoint (MinIO, etc.)                                                            |
| `S3_ACCESS_KEY_ID`       | 空 / empty                                                            | `STORAGE_DRIVER=s3` 时必填 / Required when `STORAGE_DRIVER=s3`                                                                              |
| `S3_SECRET_ACCESS_KEY`   | 空 / empty                                                            | `STORAGE_DRIVER=s3` 时必填 / Required when `STORAGE_DRIVER=s3`                                                                              |
| `S3_PREFIX`              | 空 / empty                                                            | 可选，对象 key 前缀 / Optional object key prefix                                                                                             |
| `S3_USE_SSL`             | `true`                                                                | S3 URL 生成默认是否使用 HTTPS / Whether S3 URL generation defaults to HTTPS                                                                  |
| `S3_FORCE_PATH_STYLE`    | `false`                                                               | 是否强制 path-style（MinIO 常见） / Force path-style addressing (common for MinIO)                                                          |
| `LOG_LEVEL`              | `debug`                                                               | 日志级别 / Log level                                                                                                                        |
| `LOG_FORMAT`             | `json`                                                                | 日志格式（当前由 zap config 控制） / Log format (currently controlled by zap config)                                                        |

## 3. 前端环境变量

| 变量                  | 默认值                  | 说明                        |
| --------------------- | ----------------------- | --------------------------- |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | 前端请求后端 API 的基础地址 |

## 4. 推荐配置示例

本地开发（SQLite）：

```env
APP_ENV=development
APP_PORT=8080
DB_DRIVER=sqlite
DB_DSN=./data/app.db
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

容器化（PostgreSQL）：

```env
APP_ENV=production
APP_PORT=8080
DB_DRIVER=postgres
DB_DSN=postgres://gonext:gonext@db:5432/gonext?sslmode=disable
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

## 5. 生产环境建议

1. 必须替换 `JWT_SECRET`，禁止使用默认值。
2. `CORS_ALLOWED_ORIGINS` 仅保留实际站点域名。
3. 生产环境推荐 `APP_ENV=production`，避免开发行为（如自动迁移）。
4. 若使用对象存储，设置 `STORAGE_DRIVER=s3` 并提供完整的 S3/MinIO 参数。

4. For object storage, set `STORAGE_DRIVER=s3` and provide the required S3/MinIO settings.

## 6. 启动时校验（Fail-Fast） / Startup Validation (Fail-Fast)

后端在 `backend/internal/config` 中执行集中式配置校验，`Load()` 会在应用依赖初始化前进行验证。  
任一关键配置无效时，服务会在启动阶段直接退出，并输出包含字段名的错误信息。

The backend performs centralized configuration validation in `backend/internal/config`.  
`Load()` validates config before dependency initialization. If any critical setting is invalid, startup exits early with field-level errors.

当前强制约束如下（Enforced constraints）：

1. `APP_ENV` 仅允许：`development`、`test`、`staging`、`production`。 / `APP_ENV` must be one of `development`, `test`, `staging`, `production`.
2. `DB_DRIVER` 仅允许：`sqlite`、`postgres`。 / `DB_DRIVER` must be one of `sqlite`, `postgres`.
3. `JWT_SECRET` 必须非空；当 `APP_ENV=production` 时，禁止使用默认/占位值（例如 `change-me-in-production`）。 / `JWT_SECRET` must be non-empty; when `APP_ENV=production`, default or placeholder values (for example `change-me-in-production`) are rejected.
4. `RATE_LIMIT_REQUESTS` 必须大于 0。 / `RATE_LIMIT_REQUESTS` must be greater than 0.
5. `RATE_LIMIT_DURATION` 必须是可解析且大于 0 的 Go duration（例如 `30s`、`1m`）。 / `RATE_LIMIT_DURATION` must be a parseable positive Go duration (for example `30s`, `1m`).
6. `UPLOAD_MAX_SIZE` 必须大于 0。 / `UPLOAD_MAX_SIZE` must be greater than 0.
7. `UPLOAD_DIR` 必须非空。 / `UPLOAD_DIR` must be non-empty.
8. `UPLOAD_ALLOWED_TYPES` 必须是非空的逗号分隔后缀列表，且每项满足 `.ext` 形式（如 `.jpg,.png`）；列表中的每个后缀都必须有内置 MIME 兼容规则。 / `UPLOAD_ALLOWED_TYPES` must be a non-empty comma-separated extension list, and each item must match the `.ext` format (for example `.jpg,.png`); every configured extension must have built-in MIME compatibility rules.
9. `UPLOAD_PUBLIC_BASE_URL` 若设置，必须是可解析的 `http`/`https` URL。 / If set, `UPLOAD_PUBLIC_BASE_URL` must be a parseable `http`/`https` URL.
10. `STORAGE_DRIVER` 仅允许 `local` 或 `s3`。 / `STORAGE_DRIVER` must be `local` or `s3`.
11. 当 `STORAGE_DRIVER=local` 时，`UPLOAD_DIR` 必须非空。 / When `STORAGE_DRIVER=local`, `UPLOAD_DIR` must be non-empty.
12. 当 `STORAGE_DRIVER=s3` 时，`S3_BUCKET`、`S3_REGION`、`S3_ACCESS_KEY_ID`、`S3_SECRET_ACCESS_KEY` 必须提供；`S3_ENDPOINT` 若设置必须是可解析的 `http`/`https` URL。 / When `STORAGE_DRIVER=s3`, `S3_BUCKET`, `S3_REGION`, `S3_ACCESS_KEY_ID`, and `S3_SECRET_ACCESS_KEY` are required; if set, `S3_ENDPOINT` must be a parseable `http`/`https` URL.
