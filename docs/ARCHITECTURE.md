# 架构说明

## 1. 系统拓扑

```text
Browser (Next.js)
  -> REST API (Gin)
     -> Service Layer
        -> Repository Layer
           -> Database (SQLite/PostgreSQL via GORM)
```

- 前端通过 `frontend/lib/api-client.ts` 统一请求后端 API。
- 后端统一从 `api/openapi.yaml` 维护接口契约。
- JWT 鉴权 + 全局中间件（日志、恢复、限流、CORS）。

## 2. 后端分层

目录：`backend/internal`

1. `handler/`：HTTP 入参解析、响应封装、路由注册
2. `service/`：业务逻辑（鉴权、数据校验、流程编排）
3. `repository/`：数据访问与查询
4. `model/`：GORM 数据模型
5. `dto/`：请求/响应数据结构

依赖方向：`handler -> service -> repository -> model`

## 3. 请求链路

后端启动后在 `main.go` 中注册的中间件顺序：

1. `Recovery`：panic 恢复
2. `RequestLogger`：结构化访问日志
3. `ErrorHandler`：统一错误回包
4. `RateLimiter`：按 IP 限流
5. `CORS`

鉴权路由通过 `middleware.Auth(jwtManager)` 解析 `Authorization: Bearer <token>`，把 `user_id` 和 `user_role` 注入 Gin Context。

## 4. 认证与会话

1. 登录/注册成功返回 `access_token` + `refresh_token`
2. 前端 `zustand` 持久化 token（`auth-storage`）
3. `axios` 请求拦截器自动附带 access token
4. 遇到 `401` 自动调用 `/api/v1/auth/refresh`
5. 刷新失败则登出并跳转登录页

## 5. 文件上传

- 路由：`POST /api/v1/upload`
- 存储：本地目录（`UPLOAD_DIR`）
- 静态访问：`/uploads/*`
- 校验：后缀白名单 + 大小限制

## 6. 数据库策略

- 本地默认 `sqlite`（开发模式可自动 `AutoMigrate`）
- `docker-compose.yml` 默认 `postgres`
- 可发布变更应优先通过 `backend/migrations/*.sql` 管理

## 7. 类型安全链路

```text
backend behavior
  -> api/openapi.yaml
     -> make gen-types
        -> frontend/types/api.ts
```

接口改动必须先更新 OpenAPI，再生成前端类型并联调。
