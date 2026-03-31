# CI/CD 说明

工作流目录：`.github/workflows`

## 1. CI / Quality Gate（主质量门禁）

文件：`ci-quality-gate.yml`

触发条件：

1. `pull_request` 到 `main/develop`
2. `push` 到 `main/develop`
3. 手动触发 `workflow_dispatch`

核心设计：

1. 先执行 `detect-changes`，输出 `backend/frontend/api/shared/codegen` 影响范围
2. 根据输出有条件运行：
   - `backend-quality`：lint + unit tests + build
   - `frontend-quality`：lint + typecheck + unit tests + build
   - `codegen-drift`：版本策略检查 + `make check-codegen-drift`
3. 最终由 `quality-gate` 汇总上游作业结果，作为稳定门禁结果

## 2. CI / Merge Validation（合并后/手动重验证）

文件：`merge-validation.yml`

触发条件：

1. `push` 到 `main/develop`（用于合并后验证）
2. 手动触发 `workflow_dispatch`（trusted maintainer）

执行命令：

1. `make e2e`

说明：

- 初始版本仅提供合并后 + 手动触发。
- `/run-e2e` 这类评论命令触发暂不纳入本阶段，后续可扩展。

## 3. Release / Docker Images（发布流水线）

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
make check-codegen-drift
make e2e
```

## 5. 分支保护与质量门禁建议

对受保护分支，建议将 `CI / Quality Gate` 工作流中的最终作业 `quality-gate` 设为主要必需检查（required check）。

推荐分支策略与评审要求见 [DEVELOPMENT.md](./DEVELOPMENT.md)。
版本号与变更日志维护见 [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md)。
