# 数据库说明

## 1. 当前模型

### users

- 主键：`id`
- 关键字段：`username`（唯一）、`email`（唯一）、`password_hash`、`role`
- 软删除：`deleted_at`

### items

- 主键：`id`
- 关键字段：`title`、`description`、`status`、`user_id`
- 关联：`user_id -> users.id`
- 软删除：`deleted_at`

## 2. 双数据库策略

1. 本地默认：SQLite（`DB_DRIVER=sqlite`）
2. 容器环境：PostgreSQL（`docker-compose.yml`）

注意：代码中开发环境会触发 `AutoMigrate`，但可发布变更仍建议使用 SQL migration 管理。

## 3. Migration 管理

创建迁移：

```bash
make new-migration name=add_xxx
```

执行迁移：

```bash
make migrate-up
make migrate-down
```

要求：

1. 必须提供 `up/down`。
2. PR 中必须写明回滚方式。
3. 变更需在目标数据库方言上验证。

## 4. 种子数据

执行：

```bash
make seed
```

会创建：

1. 用户 `admin@example.com`（`admin`）
2. 用户 `user@example.com`（`user`）
3. 若干示例 `items`

实现位置：`backend/cmd/seed/main.go`

## 5. 建议实践

1. 开发联调用 SQLite 提升效率。
2. 合并前至少在 PostgreSQL 环境验证一次关键路径。
3. 禁止只依赖 `AutoMigrate` 作为发布数据库变更手段。
