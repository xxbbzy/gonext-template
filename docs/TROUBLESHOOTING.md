# 常见问题排查

## 1. 前端请求全部 401

排查顺序：

1. 登录是否成功拿到 `access_token` 和 `refresh_token`
2. 浏览器本地存储是否存在 `auth-storage`
3. 请求头是否带 `Authorization: Bearer <token>`
4. `JWT_SECRET` 是否在服务重启前后变化（变化会导致旧 token 失效）

## 2. 刷新 token 失败后反复跳转登录

可能原因：

1. `refresh_token` 已过期（默认 168h）
2. 后端返回 401，前端拦截器触发 `logout`

处理建议：重新登录并检查后端时间与时区设置。

## 3. CORS 报错

检查：

1. `CORS_ALLOWED_ORIGINS` 是否包含前端地址
2. 多域名时是否使用逗号分隔
3. 前端是否请求了错误 API 域名（`NEXT_PUBLIC_API_URL`）

## 4. 文件上传失败

重点检查：

1. 文件后缀是否在 `UPLOAD_ALLOWED_TYPES`
2. 文件大小是否超过 `UPLOAD_MAX_SIZE`
3. 上传目录 `UPLOAD_DIR` 是否可写

## 5. `make lint-backend` 报命令不存在

本地未安装 `golangci-lint`。安装后重试，或使用容器化方式执行。

## 6. `make migrate-up` 失败

重点检查：

1. 是否安装 `migrate` CLI
2. `DB_DSN` 是否已导出且格式正确
3. 数据库服务是否可达

## 7. `make dev` 启动后前端无法访问后端

重点检查：

1. 后端是否监听在 `APP_PORT`（默认 8080）
2. 前端 `NEXT_PUBLIC_API_URL` 是否正确
3. 本机端口是否冲突
