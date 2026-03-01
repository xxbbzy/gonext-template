# 贡献指南

感谢你对 GoNext Template 的贡献！

## 分支策略

- `main` — 稳定版本，受保护
- `develop` — 开发分支
- `feature/<name>` — 功能分支
- `fix/<name>` — 修复分支

### 工作流

1. 从 `develop` 创建功能分支
2. 开发并提交
3. 创建 PR 到 `develop`
4. Review 通过后合并

## Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

| Type       | 说明       |
| ---------- | ---------- |
| `feat`     | 新功能     |
| `fix`      | Bug 修复   |
| `docs`     | 文档更新   |
| `style`    | 格式调整   |
| `refactor` | 重构       |
| `test`     | 测试相关   |
| `chore`    | 构建/工具  |
| `ci`       | CI/CD 配置 |

### 示例

```
feat(auth): add JWT token refresh endpoint

Add POST /api/v1/auth/refresh endpoint that accepts
a refresh token and returns a new access token.

Closes #42
```

## 开发流程

```bash
# 1. Fork & Clone
git clone <your-fork-url>

# 2. 安装依赖
make init

# 3. 创建分支
git checkout -b feature/your-feature

# 4. 开发
make dev

# 5. 代码检查
make lint

# 6. 测试
make test

# 7. 提交
git add .
git commit -m "feat(scope): description"

# 8. 推送并创建 PR
git push origin feature/your-feature
```

## 代码规范

### Go
- 遵循 [Effective Go](https://go.dev/doc/effective_go) 规范
- 使用 `golangci-lint` 检查
- 函数和类型必须有注释

### TypeScript
- 遵循 ESLint + Prettier 配置
- 优先使用 TypeScript 类型而非 `any`
- 组件使用函数式组件

## PR 模板

```markdown
## 描述
<!-- 简要描述你的改动 -->

## 改动类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 文档更新
- [ ] 重构
- [ ] 其他

## 检查清单
- [ ] 代码通过 lint 检查
- [ ] 添加了必要的测试
- [ ] 文档已更新
- [ ] Commit message 遵循 Conventional Commits
```
