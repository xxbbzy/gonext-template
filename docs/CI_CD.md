# CI/CD 说明 / CI/CD Guide

工作流目录：`.github/workflows`  
Workflow directory: `.github/workflows`

## 1. CI / Quality Gate（主质量门禁 / Primary quality gate）

文件 / File: `ci-quality-gate.yml`

触发条件 / Triggers:

1. `pull_request` 到 `main/develop`  
   `pull_request` to `main/develop`
2. `push` 到 `main/develop`  
   `push` to `main/develop`
3. 手动触发 `workflow_dispatch`  
   Manual `workflow_dispatch`

核心设计 / Core design:

1. 先执行 `detect-changes`，输出 `backend/frontend/api/shared/codegen` 影响范围。  
   Run `detect-changes` first to determine affected scopes (`backend/frontend/api/shared/codegen`).
2. 根据输出有条件运行。  
   Conditionally run jobs based on detected scope:
   - `backend-quality`：lint + unit tests + build  
     `backend-quality`: lint + unit tests + build
   - `frontend-quality`：lint + typecheck + unit tests + build  
     `frontend-quality`: lint + typecheck + unit tests + build
   - `codegen-drift`：版本策略检查 + `make check-codegen-drift`  
     `codegen-drift`: runtime-version policy checks + `make check-codegen-drift`
3. 最终由 `quality-gate` 汇总上游作业结果，作为稳定门禁结果。  
   The final `quality-gate` job summarizes upstream results and serves as the stable gate output.

## 2. CI / Merge Validation（合并后/手动重验证 / Post-merge or manual re-validation）

文件 / File: `merge-validation.yml`

触发条件 / Triggers:

1. `push` 到 `main/develop`（用于合并后验证）  
   `push` to `main/develop` (post-merge validation)
2. 手动触发 `workflow_dispatch`（trusted maintainer）  
   Manual `workflow_dispatch` (trusted maintainer)

执行命令 / Command:

1. `make e2e`

说明 / Notes:

- 初始版本仅提供合并后 + 手动触发。  
  The initial version supports post-merge and manual triggers only.
- `/run-e2e` 这类评论命令触发暂不纳入本阶段，后续可扩展。  
  Comment-command triggers such as `/run-e2e` are intentionally deferred and may be added later.

## 3. Release / Docker Images（发布流水线 / Release pipeline）

文件 / File: `docker-build.yml`

触发条件 / Triggers:

1. `push` tag：`v*`  
   Tag push: `v*`
2. `release` 发布  
   `release` published

行为 / Behavior:

1. 构建并推送 backend 镜像到 `ghcr.io/<repo>/backend:<tag>`。  
   Build and push backend image to `ghcr.io/<repo>/backend:<tag>`.
2. 构建并推送 frontend 镜像到 `ghcr.io/<repo>/frontend:<tag>`。  
   Build and push frontend image to `ghcr.io/<repo>/frontend:<tag>`.

## 4. 本地对齐 CI 的建议命令 / Recommended local commands aligned with CI

```bash
make check
make check-codegen-drift
make e2e
```

## 5. 分支保护与质量门禁建议 / Branch protection guidance

对受保护分支，建议将 `CI / Quality Gate` 工作流中的最终作业 `quality-gate` 设为主要必需检查项（required check）。  
For protected branches, set the final `quality-gate` job in `CI / Quality Gate` as the primary required check.

推荐分支策略与评审要求见 [DEVELOPMENT.md](./DEVELOPMENT.md)。  
See [DEVELOPMENT.md](./DEVELOPMENT.md) for recommended branch strategy and review requirements.

版本号与变更日志维护见 [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md)。  
See [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md) for versioning and changelog maintenance.
