# API 使用指南

接口契约源文件：`api/openapi.yaml`

默认本地地址：`http://localhost:8080`

## 1. 通用响应格式

```json
{
  "code": 0,
  "data": {},
  "message": "success"
}
```

错误示例：

```json
{
  "code": 401,
  "data": null,
  "message": "unauthorized"
}
```

## 2. 认证接口

1. `POST /api/v1/auth/register`：注册
2. `POST /api/v1/auth/login`：登录
3. `POST /api/v1/auth/refresh`：刷新 token
4. `GET /api/v1/auth/profile`：获取当前用户（需 Bearer Token）

登录示例：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

## 3. Items 接口（需鉴权）

1. `GET /api/v1/items`：分页查询
2. `POST /api/v1/items`：创建
3. `GET /api/v1/items/{id}`：详情
4. `PUT /api/v1/items/{id}`：更新
5. `DELETE /api/v1/items/{id}`：软删除

分页参数：

- `page`：默认 `1`
- `page_size`：默认 `10`，最大 `100`
- `keyword`：按 `title`/`description` 模糊搜索
- `status`：`active`/`inactive`

## 4. 上传接口（需鉴权）

- `POST /api/v1/upload`
- `Content-Type: multipart/form-data`
- 字段名：`file`
- 限制：
  - 最大文件：`UPLOAD_MAX_SIZE`
  - 后缀白名单：`UPLOAD_ALLOWED_TYPES`

上传成功后返回可访问 URL（`/uploads/...`）。

## 5. 健康检查

1. `GET /healthz`：进程存活
2. `GET /readyz`：依赖就绪（包含 DB Ping）

## 6. 错误码

| 业务码 | 含义                     |
| ------ | ------------------------ |
| `0`    | success                  |
| `400`  | bad request              |
| `401`  | unauthorized             |
| `403`  | forbidden                |
| `404`  | resource not found       |
| `409`  | conflict                 |
| `413`  | file too large           |
| `429`  | too many requests        |
| `1002` | email already registered |
| `1003` | invalid credentials      |
| `1004` | token expired            |
| `1005` | invalid token            |
| `1006` | file type not allowed    |

## 7. OpenAPI 与前端类型同步

当接口契约变更时：

```bash
make gen-types
# 准备提交/合并含契约变更的 PR 时，推荐全量刷新派生产物：
# make gen
# 或至少：
# make gen-server
# make swagger
```

When the API contract changes:

```bash
make gen-types
# To prepare for PRs that change the contract, refresh generated artifacts:
# make gen
# or at least:
# make gen-server
# make swagger
```

默认会生成：`frontend/types/api.ts`

## 8. 当前实现与契约差异

`GET /api/v1/auth/profile` 的 OpenAPI 定义包含 `username/email`，但当前实现主要依赖 JWT context，返回体可能仅稳定包含 `id` 与 `role`。建议后续对齐实现与契约。