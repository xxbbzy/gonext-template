# 前端概览

本仓库前端由 Next.js App Router 驱动，负责渲染用户界面。共享请求辅助函数集中在 `frontend/lib/api-client.gen.ts`，TanStack Query 选项封装集中在 `frontend/lib/api-query.ts`，而请求/响应类型来自 OpenAPI 生成的 `frontend/types/api.ts`。契约变更后先运行 `make gen-types` 刷新这些类型。

## 启动方式

- 优先在仓库根目录运行 `make dev-frontend`，与仓库统一工作流保持一致（该命令仅启动前端）。
- 如需在前端目录单独调试，可进入 `frontend` 目录执行 `npm run dev`。

## 联调配置

- 默认后端地址：`NEXT_PUBLIC_API_URL=http://localhost:8080`。
- 本地前端访问地址：`http://localhost:3000`。
- API 请求基地址由上述环境变量控制。

## 常用脚本

- `npm run dev`：启动开发服务器。
- `npm run build`：执行生产构建。
- `npm run lint`：检查 lint 规则。
- `npm run typecheck`：执行 TypeScript 类型检查（`tsc --noEmit`）。
- `npm run test`：运行 Vitest 单元测试。
- `make gen-types`：刷新由 OpenAPI 驱动的 `frontend/types/api.ts`，在契约变更后执行；若需要同时刷新 Go 服务 stub 和 Swagger 文档，可使用 `make gen`。

## 目录说明

- `app/`：App Router 页面、布局、服务器组件等入口，保持与路由对应。
- `lib/`：共享逻辑层，包含 OpenAPI 请求封装（`api-client.gen.ts`）、TanStack Query 选项（`api-query.ts`）、查询提供器、工具函数等。
- `stores/`：Zustand 状态管理（如 `auth.ts`）统一保存认证与会话状态。
- `types/`：OpenAPI 生成的 API 类型定义（当前位于 `types/api.ts`）。
