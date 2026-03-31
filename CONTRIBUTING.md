# 贡献指南

感谢你对 GoNext Template 的贡献！

> 团队统一开发规范请先阅读：[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)

## 分支策略

- `main`：稳定分支（受保护）
- `develop`：日常集成分支
- `feature/<name>`：功能分支（从 `develop` 切）
- `fix/<name>`：修复分支（从 `develop` 切）
- `hotfix/<name>`：紧急修复分支（从 `main` 切，修复后回合并到 `develop`）

## 提交流程

```bash
# 1. Fork & Clone
git clone <your-fork-url>

# 2. 安装依赖
make init

# 3. 创建分支
git checkout -b feature/your-feature

# 4. 开发联调
make dev

# 5. 本地质量门禁（必须通过）
make check

# 6. 提交
git add .
git commit -m "feat(scope): description"

# 7. 推送并创建 PR 到 develop
git push origin feature/your-feature
```

## Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/)：

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

常用 `type`：`feat`、`fix`、`docs`、`refactor`、`test`、`chore`、`ci`。

## 代码规范

### Go

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 通过 `golangci-lint`
- 导出函数和类型需有注释

### TypeScript

- 通过 ESLint + TypeScript typecheck
- 优先使用明确类型，避免 `any`
- 使用函数式组件

## PR 必填清单

- [ ] 目标分支是 `develop`
- [ ] 本地已通过 `make check`
- [ ] PR CI 已通过主质量门禁：`CI / Quality Gate` 的 `quality-gate` 作业
- [ ] 若改动 API 契约，已同步 `api/openapi.yaml`：开发联调可先执行 `make gen-types` 刷新前端类型；提交/合并前需确保后端生成文件与 Swagger 同步（推荐 `make gen`，或至少 `make gen-server` + `make swagger`）
- [ ] 若改动数据库 schema，已提供 migration（含回滚）
- [ ] 文档已同步更新（如适用）
