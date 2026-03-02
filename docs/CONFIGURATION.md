# 配置说明

## 1. 配置加载规则

后端配置加载逻辑在 `backend/internal/config/config.go`：

1. 读取 `.env`（若存在）
2. 读取系统环境变量
3. 对未提供项使用默认值

前端通过环境变量 `NEXT_PUBLIC_API_URL` 指向后端地址。

## 2. 后端环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_NAME` | `gonext-template` | 服务名 |
| `APP_ENV` | `development` | 环境（影响日志、自动迁移等） |
| `APP_PORT` | `8080` | HTTP 监听端口 |
| `DB_DRIVER` | `sqlite` | 数据库驱动：`sqlite`/`postgres` |
| `DB_DSN` | `./data/app.db` | 数据库连接串 |
| `JWT_SECRET` | `change-me-in-production` | JWT 签名密钥 |
| `JWT_ACCESS_EXPIRY` | `15m` | Access Token 有效期 |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh Token 有效期 |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | 允许来源，逗号分隔 |
| `RATE_LIMIT_REQUESTS` | `100` | 单窗口内允许请求数 |
| `RATE_LIMIT_DURATION` | `1m` | 限流窗口 |
| `UPLOAD_MAX_SIZE` | `10485760` | 上传大小上限（字节） |
| `UPLOAD_DIR` | `./uploads` | 上传目录 |
| `UPLOAD_ALLOWED_TYPES` | `.jpg,.jpeg,.png,.gif,.pdf,.doc,.docx` | 上传后缀白名单 |
| `STORAGE_DRIVER` | `local` | 当前仅本地存储实现 |
| `LOG_LEVEL` | `debug` | 日志级别 |
| `LOG_FORMAT` | `json` | 日志格式（当前由 zap config 控制） |

## 3. 前端环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
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
4. 使用独立存储（对象存储）时，建议扩展 `Storage` 接口并替换本地实现。
