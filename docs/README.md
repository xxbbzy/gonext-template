# GoNext Template 文档中心

本目录提供项目的完整开发与交付文档，覆盖本地开发、架构设计、API 约定、数据库、CI/CD、部署与排障。

## 文档地图

| 文档 | 适用场景 |
|---|---|
| [QUICK_START.md](./QUICK_START.md) | 5-10 分钟内启动项目并完成联调 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 理解后端分层、前端结构、鉴权与请求链路 |
| [CONFIGURATION.md](./CONFIGURATION.md) | 配置环境变量与运行参数 |
| [API_GUIDE.md](./API_GUIDE.md) | 使用和调试后端 API |
| [DATABASE.md](./DATABASE.md) | 数据模型、迁移、种子数据与数据库策略 |
| [DEVELOPMENT.md](./DEVELOPMENT.md) | 团队协作规范、分支策略、提交流程 |
| [CI_CD.md](./CI_CD.md) | GitHub Actions 流水线说明 |
| [DEPLOYMENT.md](./DEPLOYMENT.md) | Docker/容器化部署实践 |
| [CHANGELOG_GUIDE.md](./CHANGELOG_GUIDE.md) | 版本号、变更日志与发布记录规范 |
| [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) | 常见问题与排查路径 |

## 推荐阅读顺序

1. 新成员：`QUICK_START -> ARCHITECTURE -> DEVELOPMENT`
2. 后端开发：`ARCHITECTURE -> API_GUIDE -> DATABASE`
3. 前端开发：`QUICK_START -> API_GUIDE -> CONFIGURATION`
4. 发布/运维：`CI_CD -> CHANGELOG_GUIDE -> DEPLOYMENT -> TROUBLESHOOTING`

## 文档维护约定

1. 修改 API 行为时，必须同步更新 `api/openapi.yaml` 与 [API_GUIDE.md](./API_GUIDE.md)。
2. 修改环境变量时，必须同步 `.env.example` 与 [CONFIGURATION.md](./CONFIGURATION.md)。
3. 修改流程类命令（`Makefile`、CI）时，必须同步 [QUICK_START.md](./QUICK_START.md) 与 [CI_CD.md](./CI_CD.md)。
