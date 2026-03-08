# Changelog 与发布记录指南

本指南定义项目的版本号规则、`CHANGELOG.md` 维护方式和发版步骤，目标是让每次发布都可追溯、可回滚、可沟通。

## 1. 版本号规范（SemVer）

采用 `MAJOR.MINOR.PATCH`：

1. `MAJOR`：不兼容变更（Breaking Change）
2. `MINOR`：向后兼容的新功能
3. `PATCH`：向后兼容的问题修复

示例：

- `v1.3.0`：新增功能，不破坏兼容
- `v1.3.1`：修复 bug
- `v2.0.0`：接口/行为有不兼容调整

## 2. CHANGELOG 结构建议

建议在仓库根目录维护 `CHANGELOG.md`，采用以下分组：

1. `Added`：新增
2. `Changed`：修改
3. `Fixed`：修复
4. `Removed`：移除
5. `Security`：安全相关

推荐格式：

```md
# Changelog

## [Unreleased]

### Added

- ...

## [1.2.0] - 2026-03-02

### Added

- ...

### Changed

- ...

### Fixed

- ...
```

## 3. 变更收集规则

每个 PR 在描述中至少包含：

1. 变更类型：`Added/Changed/Fixed/Removed/Security`
2. 用户影响：哪些角色或页面/接口会受影响
3. 回滚方式：涉及 API/DB 时必须给出

建议在合并 PR 前，将条目先写入 `Unreleased`。

## 4. 发版流程（与当前 CI 对齐）

1. 确认 `develop` 已通过联调与回归。
2. 整理 `Unreleased` 变更，生成目标版本条目。
3. 本地执行质量门禁：`make check`。
4. 将 `develop` 合并到 `main`。
5. 在 `main` 打 tag：`vX.Y.Z` 并 push。
6. 等待 `docker-build.yml` 自动构建并推送镜像。
7. 在 Release 页面补充版本说明和回滚提示。

## 5. 发版前检查清单

- [ ] `api/openapi.yaml` 与实现一致
- [ ] 若改 API，已执行 `make gen`
- [ ] 若改 DB schema，已提供 migration 与回滚 SQL
- [ ] `docs/` 已同步更新（至少含 API/配置/流程）
- [ ] `CHANGELOG.md` 已从 `Unreleased` 切出对应版本
- [ ] 已明确 `vX.Y.Z` 的升级/回滚影响

## 6. 不同变更类型的版本建议

1. 仅修文档、注释、非行为变更：可不发版或 `PATCH`
2. 功能增强且兼容：`MINOR`
3. 修复线上缺陷：`PATCH`
4. 删除字段、修改接口语义、破坏兼容：`MAJOR`

## 7. 推荐提交模板（可复制）

```text
版本: vX.Y.Z
发布日期: YYYY-MM-DD

Added:
- ...

Changed:
- ...

Fixed:
- ...

升级提示:
- ...

回滚方式:
- ...
```
