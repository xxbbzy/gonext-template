# GoNext Template

> AI-Friendly 全栈项目脚手架，基于 Go + Next.js，强调约定优于配置、端到端类型安全、一键启动。

## 快速开始

```bash
# 克隆项目
git clone <your-repo-url>
cd gonext-template

# 一键初始化
make init

# 启动开发服务器
make dev
```

## 环境要求

- Go 1.21+
- Node.js 18+
- Make

## 项目结构

```
├── api/                    # OpenAPI spec（前后端共享契约）
├── backend/                # Go 服务
│   ├── cmd/server/         # 入口
│   ├── internal/           # 内部包
│   │   ├── config/         # 配置
│   │   ├── middleware/     # 中间件
│   │   ├── handler/        # HTTP 处理器
│   │   ├── dto/            # 请求/响应 DTO
│   │   ├── service/        # 业务逻辑
│   │   ├── repository/     # 数据访问
│   │   └── model/          # 数据模型
│   ├── pkg/                # 公共工具
│   ├── migrations/         # SQL 迁移
│   └── docs/               # Swagger 文档
├── frontend/               # Next.js 应用
├── scripts/                # 辅助脚本
├── .github/workflows/      # CI/CD
├── docker-compose.yml
├── Makefile
└── .env.example
```

## 可用命令

```bash
make help          # 列出所有命令
make init          # 初始化项目
make dev           # 启动开发服务器
make lint          # 代码检查
make test          # 运行测试
make build         # 构建生产镜像
make new-module    # 生成新模块
make seed          # 生成测试数据
```

## 技术栈

### 后端
- **Gin** - Web 框架
- **GORM** - ORM
- **Wire** - 依赖注入
- **Viper** - 配置管理
- **Zap** - 结构化日志

### 前端
- **Next.js 15** - React 框架
- **TypeScript** - 类型安全
- **shadcn/ui** - UI 组件
- **Zustand** - 状态管理
- **TanStack Query** - 数据请求

## License

MIT
