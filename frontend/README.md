# 前端概览

本仓库前端由 Next.js App Router 驱动，负责渲染用户界面并通过 `frontend/lib/api-client.gen.ts` 等类型安全客户端直连后端 API，保持与 Gin + OpenAPI 约定一致的数据交互。

## 启动方式

- 优先在仓库根目录运行 `make dev-frontend` 以统一加载前后端联调配置。
- 如需在前端目录单独调试，可进入 `frontend` 目录执行 `npm run dev`。

## 联调配置

- 默认后端地址：`NEXT_PUBLIC_API_URL=http://localhost:8080`。
- 本地前端访问地址：`http://localhost:3000`。
- 所有页面和请求通过上述环境变量与后端保持一致。

## 常用脚本

- `npm run dev`：启动开发服务器。
- `npm run build`：执行生产构建。
- `npm run lint`：检查 lint 规则。
- `npm run typecheck`：执行 TypeScript 类型检查（`tsc --noEmit`）。
- `npm run test`：运行 Vitest 单元测试。

## 目录说明

- `app/`：App Router 页面、布局、服务器组件等入口，保持与路由对应。
- `lib/`：共享逻辑层，含自动生成的 OpenAPI 客户端、查询提供器、工具函数等。
- `stores/`：Zustand 状态管理（如 `auth.ts`）统一保存认证与会话状态。
- `types/`：前端手写类型与接口补充，扩展 `lib/api-client.gen.ts` 的契约。
